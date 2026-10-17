package api

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/httperr"
)

// WorkspacePath is the path parameter for workspace-scoped routes.
type WorkspacePath struct {
	Workspace string `path:"workspace" minLength:"1"`
}

// SandboxPath is the path parameters for sandbox-scoped routes.
type SandboxPath struct {
	WorkspacePath
	Sandbox string `path:"sandbox" minLength:"1"`
}

// ListQuery carries the shared pagination and filter query parameters.
type ListQuery struct {
	PageSize      int    `query:"pageSize" minimum:"0" maximum:"1000"`
	PageToken     string `query:"pageToken"`
	LabelSelector string `query:"labelSelector"`
	AllWorkspaces bool   `query:"allWorkspaces"`
}

// SDK converts the query to SDK list options.
func (q ListQuery) SDK() v1.ListOptions {
	return v1.ListOptions{
		PageSize:      q.PageSize,
		PageToken:     q.PageToken,
		LabelSelector: q.LabelSelector,
		AllWorkspaces: q.AllWorkspaces,
	}
}

// AllowMissingQuery is the shared idempotent-delete query parameter.
type AllowMissingQuery struct {
	AllowMissing bool `query:"allowMissing"`
}

// SDK converts the query to SDK delete options.
func (q AllowMissingQuery) SDK() v1.DeleteOptions {
	return v1.DeleteOptions{AllowMissing: q.AllowMissing}
}

// IfMatchHeader carries the optional optimistic-concurrency header.
type IfMatchHeader struct {
	IfMatch string `header:"If-Match"`
}

// Version parses the resource version from If-Match. Accepted forms are a
// bare integer (7) or a quoted one ("7"); weak validators, "*" and
// non-numeric values are errors.
func (h IfMatchHeader) Version() (v uint64, set bool, err error) {
	raw := strings.TrimSpace(h.IfMatch)
	if raw == "" {
		return 0, false, nil
	}
	s := raw
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	n, perr := strconv.ParseUint(s, 10, 64)
	if perr != nil {
		return 0, false, fmt.Errorf("If-Match must be a numeric resource version such as \"7\", got %q", raw)
	}
	return n, true, nil
}

// ifMatchProblem turns an IfMatchHeader.Version error into a 400 problem.
func ifMatchProblem(err error) *httperr.Problem {
	return httperr.New(400, err.Error())
}

// withTimeout bounds ctx by d.
func withTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, d)
}

// sdkErr maps an SDK error to a problem for huma; nil stays nil.
func sdkErr(err error, ifMatch bool) error {
	if err == nil {
		return nil
	}
	return httperr.FromSDK(err, httperr.Options{IfMatch: ifMatch})
}
