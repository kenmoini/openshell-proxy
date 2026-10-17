package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/danielgtaylor/huma/v2"
	"github.com/kenmoini/openshell-proxy/internal/dto"
	"github.com/kenmoini/openshell-proxy/internal/dto/convert"
	"github.com/kenmoini/openshell-proxy/internal/httperr"
)

type sandboxOut struct{ Body dto.Sandbox }
type sandboxPageOut struct{ Body dto.Page[dto.Sandbox] }

type sandboxCreateIn struct {
	WorkspacePath
	Body dto.SandboxCreate
}

type sandboxListIn struct {
	WorkspacePath
	ListQuery
}

type sandboxDeleteIn struct {
	SandboxPath
	AllowMissingQuery
}

// sandboxActionIn is shared by start and stop.
type sandboxActionIn struct {
	SandboxPath
	Wait    bool   `query:"wait" doc:"Block until the sandbox is Ready (start) or Stopped (stop)."`
	Timeout string `query:"timeout" doc:"Wait deadline as a Go duration; capped at the server maximum."`
}

// waitDeadline resolves a requested wait duration against the server maximum.
// Empty means max; anything above max is capped; non-positive or unparsable
// values are errors.
func waitDeadline(requested string, max time.Duration) (time.Duration, error) {
	if requested == "" {
		return max, nil
	}
	d, err := time.ParseDuration(requested)
	if err != nil {
		return 0, fmt.Errorf("invalid timeout %q: %w", requested, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("timeout must be positive, got %q", requested)
	}
	if d > max {
		return max, nil
	}
	return d, nil
}

func deadlineProblem(err error) error {
	return httperr.New(http.StatusBadRequest, err.Error())
}

// registerSandboxes registers sandbox lifecycle, wait, logs and provider routes.
func registerSandboxes(api huma.API, d Deps) {
	op := func(id, method, path, summary string) huma.Operation {
		o := d.op(id, method, path, summary)
		o.Tags = []string{"Sandboxes"}
		return o
	}
	longRunning := func(o huma.Operation) huma.Operation {
		o.Metadata = map[string]any{metaLongRunning: true}
		return o
	}
	const base = "/v1/workspaces/{workspace}/sandboxes"

	create := op("createSandbox", http.MethodPost, base, "Create a sandbox")
	create.DefaultStatus = http.StatusCreated
	huma.Register(api, create, func(ctx context.Context, in *sandboxCreateIn) (*sandboxOut, error) {
		spec := convert.SandboxSpecToSDK(in.Body.Spec)
		opts := v1.CreateOptions{
			Annotations:      in.Body.Annotations,
			ServiceExposures: convert.ServiceExposuresToSDK(in.Body.ServiceExposures),
		}
		var (
			sb  *v1.Sandbox
			err error
		)
		if in.Body.Template != "" {
			sb, err = d.Client.CreateSandboxFromTemplate(ctx, in.Workspace, in.Body.Name, in.Body.Template, spec, in.Body.Labels, opts)
		} else {
			sb, err = d.Client.Sandboxes().Create(ctx, in.Workspace, in.Body.Name, spec, in.Body.Labels, opts)
		}
		if err != nil {
			return nil, sdkErr(err, false)
		}
		return &sandboxOut{Body: convert.SandboxFromSDK(sb)}, nil
	})

	huma.Register(api, op("listSandboxes", http.MethodGet, base, "List sandboxes"),
		func(ctx context.Context, in *sandboxListIn) (*sandboxPageOut, error) {
			pager, err := d.Client.Sandboxes().List(in.Workspace, in.SDK())
			if err != nil {
				return nil, sdkErr(err, false)
			}
			page, err := listPage(ctx, pager, convert.SandboxFromSDK)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &sandboxPageOut{Body: *page}, nil
		})

	huma.Register(api, op("getSandbox", http.MethodGet, base+"/{sandbox}", "Get a sandbox"),
		func(ctx context.Context, in *SandboxPath) (*sandboxOut, error) {
			sb, err := d.Client.Sandboxes().Get(ctx, in.Workspace, in.Sandbox)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &sandboxOut{Body: convert.SandboxFromSDK(sb)}, nil
		})

	huma.Register(api, op("deleteSandbox", http.MethodDelete, base+"/{sandbox}", "Delete a sandbox"),
		func(ctx context.Context, in *sandboxDeleteIn) (*deletionOut, error) {
			res, err := d.Client.Sandboxes().Delete(ctx, in.Workspace, in.Sandbox, in.SDK())
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &deletionOut{Status: convert.DeletionStatus(res), Body: convert.Deletion(res)}, nil
		})

	action := func(id, path, summary string, run func(context.Context, v1.SandboxInterface, string, string) (*v1.Sandbox, error),
		wait func(context.Context, v1.SandboxInterface, string, string, ...v1.WaitOptions) (*v1.Sandbox, error)) {
		huma.Register(api, longRunning(op(id, http.MethodPost, path, summary)),
			func(ctx context.Context, in *sandboxActionIn) (*sandboxOut, error) {
				timeout := d.Cfg.Timeouts.Request
				if in.Wait {
					var err error
					if timeout, err = waitDeadline(in.Timeout, d.Cfg.Timeouts.WaitMax); err != nil {
						return nil, deadlineProblem(err)
					}
				}
				ctx, cancel := withTimeout(ctx, timeout)
				defer cancel()
				sbs := d.Client.Sandboxes()
				sb, err := run(ctx, sbs, in.Workspace, in.Sandbox)
				if err != nil {
					return nil, sdkErr(err, false)
				}
				if in.Wait {
					if sb, err = wait(ctx, sbs, in.Workspace, in.Sandbox); err != nil {
						return nil, sdkErr(err, false)
					}
				}
				return &sandboxOut{Body: convert.SandboxFromSDK(sb)}, nil
			})
	}
	action("startSandbox", base+"/{sandbox}/start", "Start a stopped sandbox",
		func(ctx context.Context, s v1.SandboxInterface, ws, n string) (*v1.Sandbox, error) {
			return s.Start(ctx, ws, n)
		},
		func(ctx context.Context, s v1.SandboxInterface, ws, n string, o ...v1.WaitOptions) (*v1.Sandbox, error) {
			return s.WaitReady(ctx, ws, n, o...)
		})
	action("stopSandbox", base+"/{sandbox}/stop", "Stop a running sandbox",
		func(ctx context.Context, s v1.SandboxInterface, ws, n string) (*v1.Sandbox, error) {
			return s.Stop(ctx, ws, n)
		},
		func(ctx context.Context, s v1.SandboxInterface, ws, n string, o ...v1.WaitOptions) (*v1.Sandbox, error) {
			return s.WaitStopped(ctx, ws, n, o...)
		})

	registerSandboxWait(api, d, op, longRunning)
	registerSandboxProviders(api, d, op)
}
