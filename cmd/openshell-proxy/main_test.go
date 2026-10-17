package main

import (
	"testing"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
)

func TestSDKImports(t *testing.T) {
	c, err := v1.NewClient(v1.Config{Address: "127.0.0.1:1", Auth: v1.NoAuth()})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := c.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}
