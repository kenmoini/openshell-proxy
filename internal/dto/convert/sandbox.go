package convert

import (
	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/NVIDIA/OpenShell/sdk/go/openshell/v1/types"
	"github.com/kenmoini/openshell-proxy/internal/dto"
)

// SandboxFromSDK maps an SDK sandbox to its DTO.
func SandboxFromSDK(s *v1.Sandbox) dto.Sandbox {
	return dto.Sandbox{
		ObjectMeta: dto.ObjectMeta{
			ID:                s.ID,
			Name:              s.Name,
			Workspace:         s.Workspace,
			CreatedAt:         TimePtr(s.CreatedAt),
			DeletionTimestamp: s.DeletionTimestamp,
			Labels:            s.Labels,
			Annotations:       s.Annotations,
			ResourceVersion:   dto.Uint64String(s.ResourceVersion),
		},
		HostKeyFingerprint: s.HostKeyFingerprint,
		CreatedFromWorkloadTemplate: mapPtr(s.CreatedFromWorkloadTemplate, func(p v1.SandboxWorkloadTemplateProvenance) dto.TemplateProvenance {
			return dto.TemplateProvenance{Name: p.Name, ResourceVersion: p.ResourceVersion}
		}),
		Spec:        SpecFromSDK(s.Spec),
		Status:      statusFromSDK(s.Status),
		ServiceURLs: s.ServiceURLs,
	}
}

// SpecFromSDK maps an SDK sandbox spec to its DTO.
func SpecFromSDK(s v1.SandboxSpec) dto.SandboxSpec {
	return dto.SandboxSpec{
		LogLevel:      s.LogLevel,
		Environment:   s.Environment,
		Template:      mapPtr(s.Template, templateFromSDK),
		Providers:     s.Providers,
		GPU:           s.GPU,
		GPUCount:      s.GPUCount,
		Policy:        PolicyFromSDK(s.Policy),
		Command:       s.Command,
		TTY:           s.TTY,
		RestartPolicy: string(s.RestartPolicy),
	}
}

// SandboxSpecToSDK maps a DTO sandbox spec to the SDK type.
func SandboxSpecToSDK(s dto.SandboxSpec) *v1.SandboxSpec {
	return &v1.SandboxSpec{
		LogLevel:      s.LogLevel,
		Environment:   s.Environment,
		Template:      mapPtr(s.Template, templateToSDK),
		Providers:     s.Providers,
		GPU:           s.GPU,
		GPUCount:      s.GPUCount,
		Policy:        PolicyToSDK(s.Policy),
		Command:       s.Command,
		TTY:           s.TTY,
		RestartPolicy: v1.SandboxRestartPolicy(s.RestartPolicy),
	}
}

func templateFromSDK(t v1.SandboxTemplate) dto.SandboxTemplate {
	return dto.SandboxTemplate{
		Image: t.Image, RuntimeClassName: t.RuntimeClassName, AgentSocket: t.AgentSocket,
		Labels: t.Labels, Annotations: t.Annotations, Environment: t.Environment,
		UserNamespaces: t.UserNamespaces, Resources: t.Resources, DriverConfig: t.DriverConfig,
	}
}

func templateToSDK(t dto.SandboxTemplate) v1.SandboxTemplate {
	return v1.SandboxTemplate{
		Image: t.Image, RuntimeClassName: t.RuntimeClassName, AgentSocket: t.AgentSocket,
		Labels: t.Labels, Annotations: t.Annotations, Environment: t.Environment,
		UserNamespaces: t.UserNamespaces, Resources: t.Resources, DriverConfig: t.DriverConfig,
	}
}

func statusFromSDK(s v1.SandboxStatus) dto.SandboxStatus {
	return dto.SandboxStatus{
		AgentPod: s.AgentPod, AgentFd: s.AgentFd, SandboxFd: s.SandboxFd,
		Phase: string(s.Phase),
		Conditions: mapSlice(s.Conditions, func(c v1.SandboxCondition) dto.SandboxCondition {
			return dto.SandboxCondition{
				Type: c.Type, Status: c.Status, Reason: c.Reason, Message: c.Message,
				LastTransitionTime: c.LastTransitionTime,
			}
		}),
		CurrentPolicyVersion: s.CurrentPolicyVersion,
		ExitCode:             s.ExitCode,
		EndpointStatuses: mapSlice(s.EndpointStatuses, func(e v1.EndpointStatus) dto.EndpointStatus {
			return dto.EndpointStatus{
				EndpointID: e.EndpointID, Host: e.Host, Ports: e.Ports, Path: e.Path,
				LastResult: string(e.LastResult), LastReportedAt: e.LastReportedAt,
			}
		}),
		ConfigurationAdmission: mapPtr(s.ConfigurationAdmission, func(a types.SandboxConfigurationAdmission) dto.ConfigurationAdmission {
			return dto.ConfigurationAdmission{
				State: string(a.State), PolicyVersion: a.PolicyVersion, PolicyHash: a.PolicyHash,
				ConfigRevision:      dto.Uint64String(a.ConfigRevision),
				ProviderEnvRevision: dto.Uint64String(a.ProviderEnvRevision),
				Error:               a.Error,
			}
		}),
		RestartCount:           s.RestartCount,
		NextRestartAtMs:        s.NextRestartAtMs,
		MainProcessStartedAtMs: s.MainProcessStartedAtMs,
	}
}
