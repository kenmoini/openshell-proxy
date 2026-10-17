package dto

// SettingValue is a tagged union: Type selects which member is meaningful.
type SettingValue struct {
	Type   string  `json:"type" enum:"string,bool,int,bytes" required:"true"`
	String *string `json:"string,omitempty"`
	Bool   *bool   `json:"bool,omitempty"`
	Int    *int64  `json:"int,omitempty"`
	Bytes  []byte  `json:"bytes,omitempty" doc:"Base64-encoded bytes."`
}

// EffectiveSetting is a setting value with the scope it resolved from.
type EffectiveSetting struct {
	Value SettingValue `json:"value"`
	Scope string       `json:"scope,omitempty"`
}

// SandboxConfig is the effective configuration of a sandbox.
type SandboxConfig struct {
	Policy                      *SandboxPolicy              `json:"policy,omitempty"`
	PolicyVersion               uint32                      `json:"policyVersion"`
	PolicyHash                  string                      `json:"policyHash,omitempty"`
	Settings                    map[string]EffectiveSetting `json:"settings"`
	ConfigRevision              Uint64String                `json:"configRevision"`
	PolicySource                string                      `json:"policySource,omitempty"`
	GlobalPolicyVersion         uint32                      `json:"globalPolicyVersion"`
	ProviderEnvRevision         Uint64String                `json:"providerEnvRevision"`
	PolicyValidationFailureMode string                      `json:"policyValidationFailureMode,omitempty"`
}

// GatewayConfig is the gateway-global settings.
type GatewayConfig struct {
	Settings         map[string]SettingValue `json:"settings"`
	SettingsRevision Uint64String            `json:"settingsRevision"`
}

// ConfigUpdateRequest is the body of a config PATCH.
type ConfigUpdateRequest struct {
	Policy                  *SandboxPolicy         `json:"policy,omitempty"`
	SettingKey              string                 `json:"settingKey,omitempty"`
	SettingValue            *SettingValue          `json:"settingValue,omitempty"`
	DeleteSetting           bool                   `json:"deleteSetting,omitempty"`
	MergeOperations         []PolicyMergeOperation `json:"mergeOperations,omitempty"`
	Annotations             map[string]string      `json:"annotations,omitempty"`
	ExpectedResourceVersion *Uint64String          `json:"expectedResourceVersion,omitempty"`
}

// ConfigUpdateResult is the result of a config update.
type ConfigUpdateResult struct {
	Version          uint32            `json:"version"`
	PolicyHash       string            `json:"policyHash,omitempty"`
	SettingsRevision Uint64String      `json:"settingsRevision"`
	Deleted          bool              `json:"deleted"`
	Annotations      map[string]string `json:"annotations,omitempty"`
}
