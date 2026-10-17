SDK_REF ?= 5a7728c304b5528e954e9ae900cf21f84e1a12b1
SDK_MOD := github.com/NVIDIA/OpenShell/sdk/go

.PHONY: build test lint run openapi update-sdk

build:
	go build -o openshell-proxy ./cmd/openshell-proxy

test:
	go test -race ./...

lint:
	golangci-lint run

run:
	go run ./cmd/openshell-proxy --config configs/example.yaml

openapi:
	go run ./cmd/openshell-proxy --print-openapi > openapi.json

update-sdk:
	go get $(SDK_MOD)@$(SDK_REF) && go mod tidy && go test ./...
