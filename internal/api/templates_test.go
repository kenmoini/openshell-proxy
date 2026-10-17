package api

import (
	"testing"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/gateway/gatewaytest"
	"github.com/stretchr/testify/require"
)

const tplBase = "/v1/workspaces/ws1/sandbox-templates"

func TestTemplates_CRUD(t *testing.T) {
	c := gatewaytest.New()
	c.AddWorkspace(&v1.Workspace{Name: "ws1"})
	h, _ := newTestAPI(t, c, nil)

	body := `{"name":"tpl","spec":{"workload":{"image":"python:3.12","resources":{"cpu":"2","gpu":{"count":1}}},"desiredServiceLevel":{"startup":{"readyWithin":"90s","maxBurst":2}}}}`
	rec := do(t, h, "POST", tplBase, body, authHeader())
	require.Equal(t, 201, rec.Code, rec.Body.String())

	rec = do(t, h, "GET", tplBase+"/tpl", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	m := decode(t, rec.Body.Bytes())
	require.Equal(t, "tpl", m["name"])
	spec := m["spec"].(map[string]any)
	require.Equal(t, "python:3.12", spec["workload"].(map[string]any)["image"])
	require.Equal(t, "90s", spec["desiredServiceLevel"].(map[string]any)["startup"].(map[string]any)["readyWithin"])

	rec = do(t, h, "GET", tplBase, "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Len(t, decode(t, rec.Body.Bytes())["items"], 1)

	rec = do(t, h, "DELETE", tplBase+"/tpl", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())

	rec = do(t, h, "GET", tplBase+"/tpl", "", authHeader())
	require.Equal(t, 404, rec.Code, rec.Body.String())
}

func TestTemplates_BadDuration400(t *testing.T) {
	c := gatewaytest.New()
	c.AddWorkspace(&v1.Workspace{Name: "ws1"})
	h, _ := newTestAPI(t, c, nil)
	rec := do(t, h, "POST", tplBase, `{"name":"tpl","spec":{"desiredServiceLevel":{"startup":{"readyWithin":"soon"}}}}`, authHeader())
	require.Equal(t, 400, rec.Code, rec.Body.String())
}
