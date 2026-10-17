package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/NVIDIA/OpenShell/sdk/go/openshell/v1/fake"
	"github.com/danielgtaylor/huma/v2"
	"github.com/kenmoini/openshell-proxy/internal/gateway/gatewaytest"
)

func TestHealth_NoTokenNeeded(t *testing.T) {
	h, _ := newTestAPI(t, gatewaytest.New(), nil)
	rec := do(t, h, "GET", "/v1/health", "", nil)
	if rec.Code != 200 {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body)
	}
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["healthy"] != true || got["version"] != "fake" {
		t.Fatalf("body = %s", rec.Body)
	}
}

func TestHealthz_And_ReadyzHealthy(t *testing.T) {
	h, _ := newTestAPI(t, gatewaytest.New(), nil)
	rec := do(t, h, "GET", "/healthz", "", nil)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("healthz = %d %s", rec.Code, rec.Body)
	}
	rec = do(t, h, "GET", "/readyz", "", nil)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"status":"ready"`) {
		t.Fatalf("readyz = %d %s", rec.Code, rec.Body)
	}
}

func TestReadyz_UnhealthyIs503(t *testing.T) {
	c := gatewaytest.New(gatewaytest.WithFakeOptions(fake.WithHealthResult(&v1.HealthResult{Healthy: false})))
	h, _ := newTestAPI(t, c, nil)
	rec := do(t, h, "GET", "/readyz", "", nil)
	if rec.Code != 503 {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/problem+json") {
		t.Fatalf("content-type = %q", ct)
	}
}

func TestWhoami_RequiresToken(t *testing.T) {
	c := gatewaytest.New(gatewaytest.WithFakeOptions(fake.WithCurrentUser(&v1.CurrentUser{
		Subject: "alice", DisplayName: "Alice", Roles: []string{"admin"}, IdentityProvider: "idp",
	})))
	h, _ := newTestAPI(t, c, nil)
	rec := do(t, h, "GET", "/v1/whoami", "", nil)
	if rec.Code != 401 || !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/problem+json") {
		t.Fatalf("no token: %d %s %s", rec.Code, rec.Header().Get("Content-Type"), rec.Body)
	}
	rec = do(t, h, "GET", "/v1/whoami", "", authHeader())
	if rec.Code != 200 {
		t.Fatalf("with token: %d %s", rec.Code, rec.Body)
	}
	var got struct {
		Subject          string   `json:"subject"`
		DisplayName      string   `json:"displayName"`
		IdentityProvider string   `json:"identityProvider"`
		Roles            []string `json:"roles"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Subject != "alice" || got.DisplayName != "Alice" || got.IdentityProvider != "idp" || len(got.Roles) != 1 {
		t.Fatalf("body = %s", rec.Body)
	}
}

func TestGatewayInfo_MapsFields(t *testing.T) {
	c := gatewaytest.New(gatewaytest.WithFakeOptions(fake.WithGatewayInfo(&v1.GatewayInfo{
		Status:         v1.ServiceStatusHealthy,
		Version:        "1.2.3",
		ComputeDrivers: []v1.ComputeDriverInfo{{Name: "docker", DriverName: "d", DriverVersion: "9"}},
		Extensions:     []v1.ExtensionInfo{{Kind: "ComputeDriver", ConfiguredName: "x", ProtocolMajor: 1}},
	})))
	h, _ := newTestAPI(t, c, nil)
	rec := do(t, h, "GET", "/v1/gateway/info", "", authHeader())
	if rec.Code != 200 {
		t.Fatalf("status = %d %s", rec.Code, rec.Body)
	}
	var got struct {
		Status         string `json:"status"`
		Version        string `json:"version"`
		ComputeDrivers []struct {
			Name          string `json:"name"`
			DriverName    string `json:"driverName"`
			DriverVersion string `json:"driverVersion"`
		} `json:"computeDrivers"`
		Extensions []struct {
			Kind           string `json:"kind"`
			ConfiguredName string `json:"configuredName"`
			ProtocolMajor  uint32 `json:"protocolMajor"`
		} `json:"extensions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Status != "Healthy" || got.Version != "1.2.3" || len(got.ComputeDrivers) != 1 ||
		got.ComputeDrivers[0].Name != "docker" || got.ComputeDrivers[0].DriverVersion != "9" ||
		len(got.Extensions) != 1 || got.Extensions[0].Kind != "ComputeDriver" || got.Extensions[0].ProtocolMajor != 1 {
		t.Fatalf("body = %s", rec.Body)
	}
}

func TestOpenAPI_HasBearerScheme(t *testing.T) {
	raw, err := OpenAPI(testDeps(gatewaytest.New(), nil))
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Components struct {
			SecuritySchemes map[string]map[string]any `json:"securitySchemes"`
		} `json:"components"`
		Paths map[string]map[string]struct {
			Security []map[string][]string `json:"security"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Components.SecuritySchemes["bearer"]["scheme"] != "bearer" {
		t.Fatalf("schemes = %v", doc.Components.SecuritySchemes)
	}
	sec := doc.Paths["/v1/whoami"]["get"].Security
	if len(sec) == 0 {
		t.Fatal("whoami has no security")
	}
	if _, ok := sec[0]["bearer"]; !ok {
		t.Fatalf("whoami security = %v", sec)
	}
	if len(doc.Paths["/v1/health"]["get"].Security) != 0 {
		t.Fatal("health must not require security")
	}
}

func TestValidationErrorIsProblem(t *testing.T) {
	d := testDeps(gatewaytest.New(), nil)
	h, api := New(d)
	type probeIn struct {
		N int `query:"n" maximum:"10"`
	}
	type probeOut struct {
		Body struct {
			N int `json:"n"`
		}
	}
	huma.Register(api, d.op("probe", http.MethodGet, "/v1/_probe", "probe"), func(ctx context.Context, in *probeIn) (*probeOut, error) {
		return &probeOut{}, nil
	})
	rec := do(t, h, "GET", "/v1/_probe?n=50", "", authHeader())
	if rec.Code != 422 {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/problem+json") {
		t.Fatalf("content-type = %q", ct)
	}
	if !strings.Contains(rec.Body.String(), `"status":422`) {
		t.Fatalf("body = %s", rec.Body)
	}
}
