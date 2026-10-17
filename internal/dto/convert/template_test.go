package convert

import (
	"testing"
	"time"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/dto"
	"github.com/stretchr/testify/require"
)

func u32(v uint32) *uint32 { return &v }

func TestTemplate_RoundTrip(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	in := &v1.SandboxWorkloadTemplate{
		ID: "id1", Name: "tpl", Workspace: "ws", CreatedAt: now,
		Labels: map[string]string{"a": "b"}, ResourceVersion: 7,
		Spec: v1.SandboxWorkloadTemplateSpec{
			Workload: &v1.SandboxWorkloadConfig{
				Image:       "python:3.12",
				Environment: map[string]string{"K": "V"},
				Resources: &v1.SandboxResources{CPU: "2", Memory: "4Gi",
					GPU: &v1.SandboxGPURequirements{Count: u32(1)}},
			},
			DriverConfig: map[string]any{"x": "y"},
			DesiredServiceLevel: &v1.SandboxServiceLevel{
				Startup: &v1.SandboxStartup{ReadyWithin: 90 * time.Second, MaxBurst: 3},
			},
		},
	}
	d := TemplateFromSDK(in)
	require.Equal(t, "90s", d.Spec.DesiredServiceLevel.Startup.ReadyWithin)
	require.Equal(t, "tpl", d.Name)

	out, err := TemplateToSDK(d)
	require.NoError(t, err)
	require.Equal(t, in.Spec, out.Spec)
	require.Equal(t, "tpl", out.Name)
	require.Equal(t, map[string]string{"a": "b"}, out.Labels)
}

func TestTemplate_NilPointers(t *testing.T) {
	d := TemplateFromSDK(&v1.SandboxWorkloadTemplate{Name: "n"})
	require.Nil(t, d.Spec.Workload)
	require.Nil(t, d.Spec.DesiredServiceLevel)
	out, err := TemplateToSDK(d)
	require.NoError(t, err)
	require.Nil(t, out.Spec.Workload)
	require.Nil(t, out.Spec.DesiredServiceLevel)

	d = TemplateFromSDK(&v1.SandboxWorkloadTemplate{Name: "n", Spec: v1.SandboxWorkloadTemplateSpec{
		Workload: &v1.SandboxWorkloadConfig{Resources: &v1.SandboxResources{GPU: &v1.SandboxGPURequirements{}}},
	}})
	require.NotNil(t, d.Spec.Workload.Resources.GPU)
	require.Nil(t, d.Spec.Workload.Resources.GPU.Count)
}

func TestTemplate_BadDuration(t *testing.T) {
	d := TemplateFromSDK(&v1.SandboxWorkloadTemplate{Name: "n"})
	d.Spec.DesiredServiceLevel = &dto.SandboxServiceLevel{Startup: &dto.SandboxStartup{ReadyWithin: "soon"}}
	_, err := TemplateToSDK(d)
	require.Error(t, err)
}
