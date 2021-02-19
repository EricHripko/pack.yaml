package frontend

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/EricHripko/pack.yaml/pkg/cib"
	"github.com/EricHripko/pack.yaml/pkg/packer2llb"

	"github.com/containerd/containerd/platforms"
	"github.com/moby/buildkit/exporter/containerimage/exptypes"
	"github.com/moby/buildkit/frontend/gateway/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const (
	keyMultiPlatform = "multi-platform"
	keyContextSubDir = "contextsubdir"
)

// Build the image with this frontend.
func Build(ctx context.Context, c client.Client) (*client.Result, error) {
	return BuildWithService(ctx, c, cib.NewService(ctx, c))
}

// BuildWithService uses the provided container image build service to
// perform the build.
func BuildWithService(ctx context.Context, c client.Client, svc cib.Service) (*client.Result, error) {
	opts := svc.GetOpts()

	// Identify target platforms
	targetPlatforms, err := svc.GetTargetPlatforms()
	if err != nil {
		return nil, err
	}
	exportMap := len(targetPlatforms) > 1
	if v := opts[keyMultiPlatform]; v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, errors.Errorf("invalid boolean value %s", v)
		}
		if !b && exportMap {
			return nil, errors.Errorf("returning multiple target plaforms is not allowed")
		}
		exportMap = b
	}
	expPlatforms := &exptypes.Platforms{
		Platforms: make([]exptypes.Platform, len(targetPlatforms)),
	}

	// Build an image for each platform
	res := client.NewResult()
	eg, ctx := errgroup.WithContext(ctx)
	for i, tp := range targetPlatforms {
		func(i int, tp *specs.Platform) {
			eg.Go(func() error {
				// Fetch config
				dtMetadata, err := svc.GetMetadata()
				if err != nil {
					return err
				}
				metadata, err := cib.ReadConfig(dtMetadata)
				if err != nil {
					return err
				}

				// Detect project type
				plugin, err := packer2llb.Detect(ctx, svc, metadata)
				if err != nil {
					return err
				}

				// LLB
				st, img, err := plugin.Build(ctx, tp, svc)
				if err != nil {
					return errors.Wrapf(err, "failed to create LLB definition")
				}
				// Marshal
				def, err := st.Marshal(ctx)
				if err != nil {
					return errors.Wrapf(err, "failed to marshal LLB definition")
				}

				// Solve
				r, err := c.Solve(ctx, client.SolveRequest{
					Definition: def.ToPB(),
				})
				if err != nil {
					return err
				}
				ref, err := r.SingleRef()
				if err != nil {
					return err
				}

				// Image config
				img.Config.User = metadata.User
				if len(metadata.Entrypoint) > 0 || len(metadata.Command) > 0 {
					// Pre-defined command
					img.Config.Entrypoint = metadata.Entrypoint
					img.Config.Cmd = metadata.Command
				} else {
					// Find command
					cmd, err := cib.FindCommand(ctx, ref)
					if err != nil {
						return err
					}
					img.Config.Entrypoint = []string{}
					img.Config.Cmd = []string{cmd}
				}

				// Export
				config, err := json.Marshal(img)
				if err != nil {
					return errors.Wrapf(err, "failed to marshal image config")
				}
				if !exportMap {
					res.AddMeta(exptypes.ExporterImageConfigKey, config)
					res.SetRef(ref)
				} else {
					p := platforms.DefaultSpec()
					if tp != nil {
						p = *tp
					}

					k := platforms.Format(p)
					res.AddMeta(fmt.Sprintf("%s/%s", exptypes.ExporterImageConfigKey, k), config)
					res.AddRef(k, ref)
					expPlatforms.Platforms[i] = exptypes.Platform{
						ID:       k,
						Platform: p,
					}
				}
				return nil
			})
		}(i, tp)
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// Export image(s)
	if exportMap {
		dt, err := json.Marshal(expPlatforms)
		if err != nil {
			return nil, err
		}
		res.AddMeta(exptypes.ExporterPlatformsKey, dt)
	}
	return res, nil
}
