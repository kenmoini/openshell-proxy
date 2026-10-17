package api

import (
	"context"
	"testing"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/gateway/gatewaytest"
	"github.com/stretchr/testify/require"
)

const profBase = "/v1/workspaces/ws1/provider-profiles"

func TestProfiles_UpdateRequiresIfMatch428(t *testing.T) {
	var gotV uint64
	stub := &gatewaytest.ProfileStub{UpdateFn: func(_ context.Context, ws, id string, v uint64, item v1.ProfileImportItem) (*v1.UpdateResult, error) {
		gotV = v
		require.Equal(t, "ws1", ws)
		require.Equal(t, "github", id)
		require.Equal(t, "my-src", item.Source)
		return &v1.UpdateResult{Updated: true, Profile: &v1.ProviderProfile{ID: id, ResourceVersion: v + 1}}, nil
	}}
	h, _ := newTestAPI(t, gatewaytest.New(gatewaytest.WithProfiles(stub)), nil)
	body := `{"profile":{"id":"github","displayName":"GH"},"source":"my-src"}`
	rec := do(t, h, "PUT", profBase+"/github", body, authHeader())
	require.Equal(t, 428, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), "If-Match header is required")

	hdr := authHeader()
	hdr["If-Match"] = `"7"`
	rec = do(t, h, "PUT", profBase+"/github", body, hdr)
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.EqualValues(t, 7, gotV)
	m := decode(t, rec.Body.Bytes())
	require.Equal(t, true, m["updated"])

	hdr["If-Match"] = "abc"
	rec = do(t, h, "PUT", profBase+"/github", body, hdr)
	require.Equal(t, 400, rec.Code, rec.Body.String())
}

func TestProfiles_ImportLint(t *testing.T) {
	var imported, linted []v1.ProfileImportItem
	stub := &gatewaytest.ProfileStub{
		ImportFn: func(_ context.Context, _ string, items []v1.ProfileImportItem) (*v1.ImportResult, error) {
			imported = items
			return &v1.ImportResult{Imported: true, Profiles: []v1.ProviderProfile{items[0].Profile}}, nil
		},
		LintFn: func(_ context.Context, _ string, items []v1.ProfileImportItem) (*v1.LintResult, error) {
			linted = items
			return &v1.LintResult{Valid: false, Diagnostics: []v1.ProfileDiagnostic{{ProfileID: "x", Message: "bad", Severity: "error"}}}, nil
		},
	}
	h, _ := newTestAPI(t, gatewaytest.New(gatewaytest.WithProfiles(stub)), nil)
	body := `{"items":[{"profile":{"id":"x","category":"Agent","endpoints":[{"host":"h","port":443}]},"source":"s"}]}`
	rec := do(t, h, "POST", profBase+"/import", body, authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Len(t, imported, 1)
	require.Equal(t, "x", imported[0].Profile.ID)
	require.EqualValues(t, "Agent", imported[0].Profile.Category)
	require.EqualValues(t, 443, imported[0].Profile.Endpoints[0].Port)
	require.Equal(t, "s", imported[0].Source)
	m := decode(t, rec.Body.Bytes())
	require.Equal(t, true, m["imported"])
	require.Len(t, m["profiles"], 1)

	rec = do(t, h, "POST", profBase+"/lint", body, authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Len(t, linted, 1)
	m = decode(t, rec.Body.Bytes())
	require.Equal(t, false, m["valid"])
	require.Len(t, m["diagnostics"], 1)
}

func TestProfiles_ListGetDelete(t *testing.T) {
	stub := &gatewaytest.ProfileStub{
		ListFn: func(string, ...v1.ListOptions) (*v1.Pager[*v1.ProviderProfile], error) {
			return gatewaytest.StaticPager([]*v1.ProviderProfile{{ID: "a", ResourceVersion: 5}}, ""), nil
		},
		GetFn: func(_ context.Context, _, id string) (*v1.ProviderProfile, error) {
			return &v1.ProviderProfile{ID: id, ResourceVersion: 5}, nil
		},
		DeleteFn: func(context.Context, string, string, ...v1.DeleteOptions) (*v1.DeletionResult, error) {
			return &v1.DeletionResult{Outcome: v1.DeletionCompleted}, nil
		},
	}
	h, _ := newTestAPI(t, gatewaytest.New(gatewaytest.WithProfiles(stub)), nil)
	rec := do(t, h, "GET", profBase, "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	items := decode(t, rec.Body.Bytes())["items"].([]any)
	require.Len(t, items, 1)
	require.Equal(t, "5", items[0].(map[string]any)["resourceVersion"])

	rec = do(t, h, "GET", profBase+"/a", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, "a", decode(t, rec.Body.Bytes())["id"])

	rec = do(t, h, "DELETE", profBase+"/a", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
}
