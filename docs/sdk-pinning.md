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

The SDK API changed after v0.1.4 (client/method signatures differ on newer
commits). Expect to adapt call sites when moving past the v0.1.4 commit.

## Working inside the Claude Code sandbox

- Set `GOCACHE=$TMPDIR/gocache` on every go command; the default build cache is read-only.
- Do not run `go get` or anything needing the network; the module cache is pre-populated.
- The warning "go: writing stat cache ... operation not permitted" is harmless.
- The `go` directive is `1.25.13` (the SDK's floor); no `toolchain` line is set.
