# pack.yaml

[![Go Report](https://goreportcard.com/badge/github.com/EricHripko/pack.yaml)](https://goreportcard.com/report/github.com/EricHripko/pack.yaml)

`pack.yaml` (or `packer`) aims to provide a zero-effort and performant way
to package your software with a container. It's implemented as a
[BuildKit](https://github.com/moby/buildkit) frontend and, therefore,
easily slots into all your existing `docker` and `docker-compose` workflows.
Images produced by `pack.yaml` are based on
[distroless](https://github.com/GoogleContainerTools/distroless) for compact
footprint.

## Integrations

`pack.yaml` takes advantage of the plugin system to provide deep integrations
with various language and build ecosystems. As of today, the following
functionality is available.

### Go

```text
 => [internal] load build definition from pack.yaml                      0.0s
 => => transferring dockerfile: 98B                                      0.0s
 => [internal] load .dockerignore                                        0.0s
 => => transferring context: 50B                                         0.0s
 => resolve image config for docker.io/erichripko/pack.yaml:latest       0.0s
 => CACHED docker-image://docker.io/erichripko/pack.yaml:latest          0.0s
 => [internal] load build definition from pack.yaml                      0.0s
 => => transferring dockerfile: 98B                                      0.0s
 => [internal] load context                                              0.0s
 => => transferring context: 109.58kB                                    0.0s
 => load metadata for docker.io/library/golang:1.14                      0.5s
 => load metadata for gcr.io/distroless/base:debug                       0.5s
 => Base build image is golang:1.14                                      7.9s
 => => resolve docker.io/library/golang:1.14                             0.0s
 => => sha256:0ecb575e629cd60aa802266a3bc6847dcf4073a 50.40MB / 50.40MB  0.0s
 => => sha256:feab2c490a3cea21cc051ff29c33cc9857418ed 10.00MB / 10.00MB  0.0s
 => => sha256:1517911a35d7939f446084c1d4c31afc552678e 68.72MB / 68.72MB  0.0s
 => => sha256:48bbd1746d63c372e12f884178053851d87f3 124.14MB / 124.14MB  0.0s
 => => sha256:1a7173b5b9a3af3e29a5837e0b2027e1c438fd1b8 2.36kB / 2.36kB  0.0s
 => => sha256:6a39a02f74ffee82a169f2d836134236dc6f69e59 1.79kB / 1.79kB  0.0s
 => => sha256:21a5635903d69da3c3d928ed429e3610eecdf878c 7.03kB / 7.03kB  0.0s
 => => sha256:7467d1831b6947c294d92ee957902c3cd448b17c5 7.83MB / 7.83MB  0.0s
 => => sha256:f15a0f46f8c38f4ca7daecf160ba9cdb3ddeafd 51.83MB / 51.83MB  0.0s
 => => sha256:944903612fdd2364b4647cf3c231c41103d1fd378add4 126B / 126B  0.0s
 => => extracting sha256:48bbd1746d63c372e12f884178053851d87f3ea4b415f3  7.1s
 => => extracting sha256:944903612fdd2364b4647cf3c231c41103d1fd378add43  0.0s
 => Base runtime image is gcr.io/distroless/base:debug                   0.1s
 => Create build output directory                                        0.6s
 => Build github.com/mpppk/cli-template                                  8.8s
 => CACHED Create output directory                                       0.0s
 => Install application(s)                                               0.1s
 => exporting to image                                                   0.1s
 => => exporting layers                                                  0.1s
 => => writing image sha256:86d32b9ed657253d9713fa1d6e0859219fc607fbbca  0.0s
```

[Go](https://golang.org/) - runs `go build` on your project and takes full
advantage of the build cache for best performance. Integration supports the
following dependency management methods:

- [go mod](https://golang.org/ref/mod) - automatically picks up the version
  of Go from `go.mod` and relies on module cache for best performance.

The following additional configuration is supported by the integration:

```yaml
# syntax = erichripko/pack.yaml
go:
  # Version of Go to use for the project.
  version: "1.14"
  # List of Go build tags to set.
  tags: ["wireinject"]
  # How the dependencies are specified.
  # Supported values are: modules
  dependencyMode: modules
```
