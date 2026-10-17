package dto

// Sandbox is a running or provisioned sandbox instance.
type Sandbox struct {
	ObjectMeta
	HostKeyFingerprint          string              `json:"hostKeyFingerprint,omitempty"`
	CreatedFromWorkloadTemplate *TemplateProvenance `json:"createdFromWorkloadTemplate,omitempty"`
	Spec                        SandboxSpec         `json:"spec"`
	Status                      SandboxStatus       `json:"status"`
	ServiceURLs                 map[string]string   `json:"serviceUrls,omitempty"`
}

// TemplateProvenance identifies the reusable template revision a sandbox was created from.
type TemplateProvenance struct {
	Name            string `json:"name,omitempty"`
	ResourceVersion string `json:"resourceVersion,omitempty"`
}

// SandboxSpec is the desired state of a sandbox.
type SandboxSpec struct {
	LogLevel      string            `json:"logLevel,omitempty"`
	Environment   map[string]string `json:"environment,omitempty"`
	Template      *SandboxTemplate  `json:"template,omitempty"`
	Providers     []string          `json:"providers,omitempty"`
	GPU           bool              `json:"gpu,omitempty"`
	GPUCount      *uint32           `json:"gpuCount,omitempty"`
	Policy        *SandboxPolicy    `json:"policy,omitempty"`
	Command       []string          `json:"command,omitempty"`
	TTY           bool              `json:"tty,omitempty"`
	RestartPolicy string            `json:"restartPolicy,omitempty" enum:",Never,OnFailure,Always"`
}

// SandboxTemplate is the inline container template of a sandbox.
type SandboxTemplate struct {
	Image            string            `json:"image,omitempty"`
	RuntimeClassName string            `json:"runtimeClassName,omitempty"`
	AgentSocket      string            `json:"agentSocket,omitempty"`
	Labels           map[string]string `json:"labels,omitempty"`
	Annotations      map[string]string `json:"annotations,omitempty"`
	Environment      map[string]string `json:"environment,omitempty"`
	UserNamespaces   *bool             `json:"userNamespaces,omitempty"`
	Resources        map[string]any    `json:"resources,omitempty"`
	DriverConfig     map[string]any    `json:"driverConfig,omitempty"`
}

// SandboxStatus is the observed state of a sandbox.
type SandboxStatus struct {
	AgentPod               string                  `json:"agentPod,omitempty"`
	AgentFd                string                  `json:"agentFd,omitempty"`
	SandboxFd              string                  `json:"sandboxFd,omitempty"`
	Phase                  string                  `json:"phase,omitempty"`
	Conditions             []SandboxCondition      `json:"conditions,omitempty"`
	CurrentPolicyVersion   uint32                  `json:"currentPolicyVersion,omitempty"`
	ExitCode               *int32                  `json:"exitCode,omitempty"`
	EndpointStatuses       []EndpointStatus        `json:"endpointStatuses,omitempty"`
	ConfigurationAdmission *ConfigurationAdmission `json:"configurationAdmission,omitempty"`
	RestartCount           uint32                  `json:"restartCount,omitempty"`
	NextRestartAtMs        int64                   `json:"nextRestartAtMs,omitempty"`
	MainProcessStartedAtMs int64                   `json:"mainProcessStartedAtMs,omitempty"`
}

// SandboxCondition describes an observed condition of a sandbox.
type SandboxCondition struct {
	Type               string `json:"type,omitempty"`
	Status             string `json:"status,omitempty"`
	Reason             string `json:"reason,omitempty"`
	Message            string `json:"message,omitempty"`
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
}

// EndpointStatus is a configured tool endpoint and its last accepted network result.
type EndpointStatus struct {
	EndpointID     string   `json:"endpointId,omitempty"`
	Host           string   `json:"host,omitempty"`
	Ports          []uint32 `json:"ports,omitempty"`
	Path           string   `json:"path,omitempty"`
	LastResult     string   `json:"lastResult,omitempty"`
	LastReportedAt string   `json:"lastReportedAt,omitempty"`
}

// ConfigurationAdmission identifies a validated or rejected configuration.
type ConfigurationAdmission struct {
	State               string       `json:"state,omitempty"`
	PolicyVersion       uint32       `json:"policyVersion,omitempty"`
	PolicyHash          string       `json:"policyHash,omitempty"`
	ConfigRevision      Uint64String `json:"configRevision,omitempty"`
	ProviderEnvRevision Uint64String `json:"providerEnvRevision,omitempty"`
	Error               string       `json:"error,omitempty"`
}
