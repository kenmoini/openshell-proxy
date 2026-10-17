# SDK pinning

The OpenShell Go SDK (`github.com/NVIDIA/OpenShell/sdk/go`) is published from a
subdirectory of the OpenShell monorepo and has no module tags. Go therefore
resolves it to a pseudo-version derived from a commit.

## Current pin

Commit `5a7728c304b5528e954e9ae900cf21f84e1a12b1` (the v0.1.4 gateway release),
resolved as `v0.0.0-20261013191600-5a7728c304b5`.

Pin to the commit matching the gateway release you deploy against, so the
proxy's client and the gateway agree on the gRPC API.

## Re-pinning

```bash
make update-sdk SDK_REF=<full-commit-sha>
```

This runs `go get`, `go mod tidy`, and `go test ./...`. Review the diff in
`go.mod`/`go.sum` and fix any compile errors before committing.

## Known signature change

The SDK API changed after v0.1.4. At v0.1.4:

```go
Sandboxes().Create(ctx, ws, name, spec, labels map[string]string, opts ...v1.CreateOptions)
// v1.CreateOptions{Annotations, ServiceExposures}
```

After v0.1.4 it became:

```go
Sandboxes().Create(ctx, ws, name, spec, opts ...v1.CreateOption)
// functional options: v1.WithLabels, v1.WithAnnotations, v1.WithServiceExposures
```

`CreateSandboxFromTemplate` changed the same way. Expect to adapt call sites
when moving past the v0.1.4 commit.
