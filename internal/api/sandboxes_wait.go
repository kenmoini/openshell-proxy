package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/danielgtaylor/huma/v2"
	"github.com/kenmoini/openshell-proxy/internal/dto"
	"github.com/kenmoini/openshell-proxy/internal/dto/convert"
)

type sandboxWaitIn struct {
	SandboxPath
	Body dto.WaitRequest
}

type sandboxLogsIn struct {
	SandboxPath
	Lines    uint32 `query:"lines" doc:"Maximum number of lines to return."`
	Since    string `query:"since" doc:"Only entries at or after this RFC 3339 time."`
	Sources  string `query:"sources" doc:"Comma-separated sources, e.g. gateway,sandbox."`
	MinLevel string `query:"minLevel" doc:"Minimum log level to include."`
}

type logsOut struct{ Body dto.LogResult }

func registerSandboxWait(api huma.API, d Deps, op func(id, method, path, summary string) huma.Operation, longRunning func(huma.Operation) huma.Operation) {
	const base = "/v1/workspaces/{workspace}/sandboxes/{sandbox}"

	huma.Register(api, longRunning(op("waitSandbox", http.MethodPost, base+"/wait", "Wait for a sandbox to become ready or stopped")),
		func(ctx context.Context, in *sandboxWaitIn) (*sandboxOut, error) {
			timeout, err := waitDeadline(in.Body.Timeout, d.Cfg.Timeouts.WaitMax)
			if err != nil {
				return nil, deadlineProblem(err)
			}
			var wo v1.WaitOptions
			if in.Body.PollInterval != "" {
				if wo.PollInterval, err = time.ParseDuration(in.Body.PollInterval); err != nil || wo.PollInterval <= 0 {
					return nil, deadlineProblem(errBadPoll(in.Body.PollInterval))
				}
			}
			ctx, cancel := withTimeout(ctx, timeout)
			defer cancel()
			var sb *v1.Sandbox
			if in.Body.For == "stopped" {
				sb, err = d.Client.Sandboxes().WaitStopped(ctx, in.Workspace, in.Sandbox, wo)
			} else {
				sb, err = d.Client.Sandboxes().WaitReady(ctx, in.Workspace, in.Sandbox, wo)
			}
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &sandboxOut{Body: convert.SandboxFromSDK(sb)}, nil
		})

	huma.Register(api, op("getSandboxLogs", http.MethodGet, base+"/logs", "Fetch sandbox logs"),
		func(ctx context.Context, in *sandboxLogsIn) (*logsOut, error) {
			var opts []v1.LogOption
			if in.Lines > 0 {
				opts = append(opts, v1.WithLogLines(in.Lines))
			}
			if in.Since != "" {
				t, err := time.Parse(time.RFC3339, in.Since)
				if err != nil {
					return nil, deadlineProblem(errBadSince(in.Since))
				}
				opts = append(opts, v1.WithLogSince(t))
			}
			if in.Sources != "" {
				opts = append(opts, v1.WithLogSources(strings.Split(in.Sources, ",")...))
			}
			if in.MinLevel != "" {
				opts = append(opts, v1.WithLogMinLevel(in.MinLevel))
			}
			res, err := d.Client.Sandboxes().GetLogs(ctx, in.Workspace, in.Sandbox, opts...)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &logsOut{Body: convert.LogResultFromSDK(res)}, nil
		})
}
