package dto

// SandboxWorkloadTemplate is a reusable sandbox workload definition.
type SandboxWorkloadTemplate struct {
	ObjectMeta
	Spec SandboxWorkloadTemplateSpec `json:"spec"`
}

// SandboxWorkloadTemplateSpec is the template body.
type SandboxWorkloadTemplateSpec struct {
	Workload            *SandboxWorkloadConfig `json:"workload,omitempty"`
	DriverConfig        map[string]any         `json:"driverConfig,omitempty"`
	DesiredServiceLevel *SandboxServiceLevel   `json:"desiredServiceLevel,omitempty"`
}

// SandboxWorkloadConfig describes the workload container.
type SandboxWorkloadConfig struct {
	Image       string            `json:"image,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Resources   *SandboxResources `json:"resources,omitempty"`
}

// SandboxResources are compute resource requests.
type SandboxResources struct {
	CPU    string                  `json:"cpu,omitempty"`
	Memory string                  `json:"memory,omitempty"`
	GPU    *SandboxGPURequirements `json:"gpu,omitempty"`
}

// SandboxGPURequirements requests GPUs; a nil Count means driver default.
type SandboxGPURequirements struct {
	Count *uint32 `json:"count,omitempty"`
}

// SandboxServiceLevel is the desired service level.
type SandboxServiceLevel struct {
	Startup *SandboxStartup `json:"startup,omitempty"`
}

// SandboxStartup holds startup expectations. ReadyWithin is a Go duration
// string such as "90s".
type SandboxStartup struct {
	ReadyWithin string `json:"readyWithin,omitempty"`
	MaxBurst    uint32 `json:"maxBurst,omitempty"`
}
