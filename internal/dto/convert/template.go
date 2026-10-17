package convert

import (
	"fmt"
	"time"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/dto"
)

// TemplateFromSDK maps an SDK workload template to its DTO.
func TemplateFromSDK(t *v1.SandboxWorkloadTemplate) dto.SandboxWorkloadTemplate {
	out := dto.SandboxWorkloadTemplate{
		ObjectMeta: dto.ObjectMeta{
			ID:                t.ID,
			Name:              t.Name,
			Workspace:         t.Workspace,
			CreatedAt:         TimePtr(t.CreatedAt),
			DeletionTimestamp: t.DeletionTimestamp,
			Labels:            t.Labels,
			Annotations:       t.Annotations,
			ResourceVersion:   dto.Uint64String(t.ResourceVersion),
		},
		Spec: dto.SandboxWorkloadTemplateSpec{DriverConfig: t.Spec.DriverConfig},
	}
	if w := t.Spec.Workload; w != nil {
		cfg := &dto.SandboxWorkloadConfig{Image: w.Image, Environment: w.Environment}
		if r := w.Resources; r != nil {
			res := &dto.SandboxResources{CPU: r.CPU, Memory: r.Memory}
			if r.GPU != nil {
				res.GPU = &dto.SandboxGPURequirements{Count: r.GPU.Count}
			}
			cfg.Resources = res
		}
		out.Spec.Workload = cfg
	}
	if sl := t.Spec.DesiredServiceLevel; sl != nil {
		lvl := &dto.SandboxServiceLevel{}
		if s := sl.Startup; s != nil {
			st := &dto.SandboxStartup{MaxBurst: s.MaxBurst}
			if s.ReadyWithin != 0 {
				st.ReadyWithin = formatDuration(s.ReadyWithin)
			}
			lvl.Startup = st
		}
		out.Spec.DesiredServiceLevel = lvl
	}
	return out
}

// TemplateToSDK maps a template DTO to the SDK type. An unparsable
// readyWithin duration is an error.
func TemplateToSDK(t dto.SandboxWorkloadTemplate) (*v1.SandboxWorkloadTemplate, error) {
	out := &v1.SandboxWorkloadTemplate{
		Name:        t.Name,
		Labels:      t.Labels,
		Annotations: t.Annotations,
		Spec:        v1.SandboxWorkloadTemplateSpec{DriverConfig: t.Spec.DriverConfig},
	}
	if w := t.Spec.Workload; w != nil {
		cfg := &v1.SandboxWorkloadConfig{Image: w.Image, Environment: w.Environment}
		if r := w.Resources; r != nil {
			res := &v1.SandboxResources{CPU: r.CPU, Memory: r.Memory}
			if r.GPU != nil {
				res.GPU = &v1.SandboxGPURequirements{Count: r.GPU.Count}
			}
			cfg.Resources = res
		}
		out.Spec.Workload = cfg
	}
	if sl := t.Spec.DesiredServiceLevel; sl != nil {
		lvl := &v1.SandboxServiceLevel{}
		if s := sl.Startup; s != nil {
			st := &v1.SandboxStartup{MaxBurst: s.MaxBurst}
			if s.ReadyWithin != "" {
				d, err := time.ParseDuration(s.ReadyWithin)
				if err != nil {
					return nil, fmt.Errorf("invalid readyWithin %q: %w", s.ReadyWithin, err)
				}
				st.ReadyWithin = d
			}
			lvl.Startup = st
		}
		out.Spec.DesiredServiceLevel = lvl
	}
	return out, nil
}

// formatDuration renders whole-second durations as "<n>s" (e.g. "90s") and
// anything else in Go's canonical form.
func formatDuration(d time.Duration) string {
	if d%time.Second == 0 {
		return fmt.Sprintf("%ds", int64(d/time.Second))
	}
	return d.String()
}
