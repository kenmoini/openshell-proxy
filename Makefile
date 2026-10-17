SDK_REF ?= 5a7728c304b5528e954e9ae900cf21f84e1a12b1
SDK_MOD := github.com/NVIDIA/OpenShell/sdk/go

.PHONY: build test lint run openapi check-openapi update-sdk

build:
	go build -o openshell-proxy ./cmd/openshell-proxy

test:
	go test -race ./...

lint:
	golangci-lint run

run:
	go run ./cmd/openshell-proxy --config configs/example.yaml

# Offline, honor your environment: GOCACHE=$$TMPDIR/gocache GOPROXY=off GOFLAGS=-mod=mod make openapi
openapi:
	go run ./cmd/openshell-proxy --print-openapi > openapi.json

# Fails when the committed openapi.json differs from the generated document.
check-openapi:
	@tmp=$$(mktemp "$${TMPDIR:-/tmp}/openapi.XXXXXX") && go run ./cmd/openshell-proxy --print-openapi > $$tmp && \
	  if cmp -s openapi.json $$tmp; then rm -f $$tmp; echo "openapi.json is up to date"; \
	  else rm -f $$tmp; echo "openapi.json is stale: run 'make openapi' and commit the result" >&2; exit 1; fi

update-sdk:
	go get $(SDK_MOD)@$(SDK_REF) && go mod tidy && go test ./...
