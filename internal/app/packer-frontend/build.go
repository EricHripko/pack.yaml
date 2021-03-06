package frontend

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EricHripko/pack.yaml/pkg/packer2llb"
	"github.com/EricHripko/pack.yaml/pkg/packer2llb/config"

	"github.com/EricHripko/buildkit-fdk/pkg/cib"
	"github.com/containerd/containerd/platforms"
	"github.com/moby/buildkit/exporter/containerimage/exptypes"
	"github.com/moby/buildkit/frontend/gateway/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// Build the image with this frontend.
func Build(ctx context.Context, c client.Client) (*client.Result, error) {
	return BuildWithService(ctx, c, cib.NewService(ctx, c))
}

// BuildWithService uses the provided container image build service to
// perform the build.
//nolint:gocyclo // Frontends are complex
func BuildWithService(ctx context.Context, c client.Client, svc cib.Service) (*client.Result, error) {
	// Identify target platforms
	targetPlatforms, err := svc.GetTargetPlatforms()
	if err != nil {
		return nil, err
	}
	exportMap, err := svc.GetIsMultiPlatform()
	if err != nil {
		return nil, err
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
				metadata, err := config.Read(dtMetadata)
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
					var cmd string
					cmd, err = findCommand(ctx, ref)
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
