package api

import (
	"context"
	"encoding/base64"
	"net/http"
	"unicode/utf8"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/danielgtaylor/huma/v2"
	"github.com/kenmoini/openshell-proxy/internal/dto"
)

type execIn struct {
	SandboxPath
	Body dto.ExecRequest
}

type execOut struct{ Body dto.ExecResult }

// registerExec registers the collected (non-streaming) exec operation.
func registerExec(api huma.API, d Deps) {
	op := d.op("execSandbox", http.MethodPost, "/v1/workspaces/{workspace}/sandboxes/{sandbox}/exec", "Run a command in a sandbox and collect its output")
	op.Tags = []string{"Exec"}
	op.MaxBodyBytes = d.Cfg.Limits.MaxExecBody
	op.Metadata = map[string]any{metaLongRunning: true}

	huma.Register(api, op, func(ctx context.Context, in *execIn) (*execOut, error) {
		ctx, cancel := withTimeout(ctx, d.Cfg.Timeouts.Exec)
		defer cancel()
		res, err := d.Client.Exec().Run(ctx, in.Workspace, in.Sandbox, in.Body.Command, v1.ExecOptions{
			Env:          in.Body.Env,
			WorkDir:      in.Body.WorkDir,
			NoLoginShell: in.Body.NoLoginShell,
		})
		if err != nil {
			return nil, sdkErr(err, false)
		}
		out := dto.ExecResult{ExitCode: res.ExitCode, Encoding: "utf8"}
		if in.Body.Encoding == "base64" || !utf8.Valid(res.Stdout) || !utf8.Valid(res.Stderr) {
			out.Encoding = "base64"
			out.Stdout = base64.StdEncoding.EncodeToString(res.Stdout)
			out.Stderr = base64.StdEncoding.EncodeToString(res.Stderr)
		} else {
			out.Stdout = string(res.Stdout)
			out.Stderr = string(res.Stderr)
		}
		return &execOut{Body: out}, nil
	})
}
