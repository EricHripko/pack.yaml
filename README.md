# pack.yaml

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
functionality is available:

- [Go](https://golang.org/) - runs `go build` on your project and takes full
  advantage of the build cache for best performance. Following dependency
  management methods are supported:
  - [go mod](https://golang.org/ref/mod) - automatically picks up the version
    of Go from `go.mod` and relies on module cache for best performance.
