package golang

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/EricHripko/pack.yaml/pkg/cib"
	"github.com/EricHripko/pack.yaml/pkg/packer2llb"
	"github.com/mitchellh/mapstructure"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/frontend/dockerfile/dockerfile2llb"
	"github.com/moby/buildkit/frontend/gateway/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	fsutil "github.com/tonistiigi/fsutil/types"
	"golang.org/x/mod/modfile"
	"golang.org/x/sync/errgroup"
)

// Regular expression for detecting a Go project.
var fileRegex = regexp.MustCompile(`(?:.*\.go)|(?:go\.mod)|(?:go\.sum)`)

// Errors returned by the plugin.
var (
	ErrUnknownDep    = errors.New("golang: unknown dependency method")
	ErrModIncomplete = errors.New("golang: incomplete go.mod file")
)

// DependencyMode describes all the supported methods for dependency resolution.
type DependencyMode string

const (
	// DMUnknown represents an unrecognised dependency method.
	DMUnknown = "unknown"
	// DMGoMod represents a go.mod/go.sum project.
	DMGoMod = "modules"
)

// Config for the Go plugin.
type Config struct {
	// Version of Go used.
	Version string
	// Method for declaring dependencies.
	DependencyMode DependencyMode
	// Build tags.
	Tags []string
}

// Plugin for Go ecosystem.
type Plugin struct {
	// General configuration supplied by the user.
	config *cib.Config
	// Configuration for the plugin.
	pluginConfig *Config
	// Name of the project.
	name string
}

// NewPlugin creates a new Go plugin with correct defaults.
func NewPlugin() *Plugin {
	return &Plugin{
		config: cib.NewConfig(),
		pluginConfig: &Config{
			DependencyMode: DMUnknown,
		},
	}
}

// Detect if this is a Go project and identify the context.
func (p *Plugin) Detect(ctx context.Context, src client.Reference, config *cib.Config) error {
	// Save config
	p.config = config
	if other, ok := p.config.Other["go"]; ok {
		err := mapstructure.Decode(other, p.pluginConfig)
		if err != nil {
			return err
		}
	}

	// Look for go files
	err := cib.WalkRecursive(ctx, src, func(file *fsutil.Stat) error {
		if fileRegex.MatchString(file.Path) {
			return packer2llb.ErrActivate
		}
		return nil
	})
	if err != packer2llb.ErrActivate {
		return err
	}

	// Identify dependency method
	if p.pluginConfig.DependencyMode == DMUnknown {
		goModGroup := new(errgroup.Group)
		goModGroup.Go(func() error {
			_, err := src.ReadFile(ctx, client.ReadRequest{Filename: "go.mod"})
			return err
		})
		goModGroup.Go(func() error {
			_, err := src.ReadFile(ctx, client.ReadRequest{Filename: "go.sum"})
			return err
		})
		if err := goModGroup.Wait(); err == nil {
			p.pluginConfig.DependencyMode = DMGoMod
		}
	}

	// Pick up the project context from the dependency metadata
	switch p.pluginConfig.DependencyMode {
	case DMUnknown:
		return ErrUnknownDep
	case DMGoMod:
		data, err := src.ReadFile(ctx, client.ReadRequest{Filename: "go.mod"})
		if err != nil {
			return errors.Wrap(err, "fail to read go.mod")
		}
		goMod, err := modfile.ParseLax("go.mod", data, nil)
		if err != nil {
			return errors.Wrap(err, "fail to parse go.mod")
		}
		if goMod.Go == nil || goMod.Module == nil {
			return ErrModIncomplete
		}
		if p.pluginConfig.Version == "" {
			p.pluginConfig.Version = goMod.Go.Version
		}
		p.name = goMod.Module.Mod.Path
	}

	return packer2llb.ErrActivate
}

const (
	// Source directory.
	dirSrc = "/src"
	// Directory where installed binaries go.
	dirInstall = "/install"
	// Directory for caching dependencies of go.mod/go.sum project.
	dirGoModCache = "/go/pkg/mod"
	// Directory for caching build outputs.
	dirGoBuildCache = "/go/build"
)

// Build the image for this Go project.
func (p *Plugin) Build(ctx context.Context, platform *specs.Platform, build cib.Service) (*llb.State, *dockerfile2llb.Image, error) {
	// Choose base image
	base := "golang:" + p.pluginConfig.Version
	state, _, err := build.From(
		base,
		platform,
		fmt.Sprintf("Base build image is %s", base),
	)
	if err != nil {
		return nil, nil, err
	}

	// Fetch sources
	src, err := build.SrcState()
	if err != nil {
		return nil, nil, err
	}
	// Create output directory
	state = state.File(
		llb.Mkdir(dirInstall, 0755),
		llb.WithCustomName("Create build output directory"),
	)
	// Build
	args := []string{"go", "install", "-v"}
	if len(p.pluginConfig.Tags) > 0 {
		args = append(args, "-tags")
		args = append(args, strings.Join(p.pluginConfig.Tags, ","))
	}
	args = append(args, "./...")

	run := []llb.RunOption{
		// Mount source code
		llb.AddMount(dirSrc, src, llb.Readonly),
		// Install executables
		llb.AddEnv("GOBIN", dirInstall),
		llb.Args(args),
		// Cache build outputs
		llb.AddMount(
			dirGoBuildCache,
			llb.Scratch(),
			llb.AsPersistentCacheDir("go-build", llb.CacheMountPrivate),
		),
		llb.AddEnv("GOCACHE", dirGoBuildCache),
		llb.WithCustomNamef("Build %s", p.name),
	}
	if p.pluginConfig.DependencyMode == DMGoMod {
		// Cache modules
		run = append(run, llb.AddMount(
			dirGoModCache,
			llb.Scratch(),
			llb.AsPersistentCacheDir("go-mod", llb.CacheMountPrivate),
		))
	}
	buildState := state.Dir(dirSrc).Run(run...).Root()

	// Runtime image
	base = "gcr.io/distroless/base"
	if p.config.Debug {
		base += ":debug"
	}
	state, img, err := build.From(
		base,
		platform,
		fmt.Sprintf("Base runtime image is %s", base),
	)
	if err != nil {
		return nil, nil, err
	}
	// Install the application
	state = state.File(
		llb.Mkdir(cib.DirInstall, 0755, llb.WithParents(true)),
		llb.WithCustomName("Create output directory"),
	)
	state = state.File(
		llb.Copy(
			buildState,
			dirInstall,
			cib.DirInstall,
			&llb.CopyInfo{CopyDirContentsOnly: true},
		),
		llb.WithCustomName("Install application(s)"),
	)

	return &state, img, err
}

func init() {
	// Register the plugin with the frontend.
	packer2llb.Register(NewPlugin())
}
