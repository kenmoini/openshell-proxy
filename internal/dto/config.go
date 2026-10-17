package dto

import "github.com/danielgtaylor/huma/v2"

// SettingValue is a tagged union: Type selects which member is meaningful.
type SettingValue struct {
	Type   string  `json:"type" enum:"string,bool,int,bytes" required:"true"`
	String *string `json:"string,omitempty"`
	Bool   *bool   `json:"bool,omitempty"`
	Int    *int64  `json:"int,omitempty"`
	Bytes  *[]byte `json:"bytes,omitempty" doc:"Base64-encoded bytes."`
}

// Resolve requires exactly the member named by Type to be present.
func (v *SettingValue) Resolve(_ huma.Context) []error {
	present := map[string]bool{
		"string": v.String != nil,
		"bool":   v.Bool != nil,
		"int":    v.Int != nil,
		"bytes":  v.Bytes != nil,
	}
	if _, known := present[v.Type]; !known {
		return nil // enum validation reports unknown types
	}
	if !present[v.Type] {
		return []error{&huma.ErrorDetail{Location: "body.settingValue." + v.Type, Message: "member named by type is required", Value: v.Type}}
	}
	var errs []error
	for k, p := range present {
		if p && k != v.Type {
			errs = append(errs, &huma.ErrorDetail{Location: "body.settingValue." + k, Message: "member does not match type " + v.Type, Value: v.Type})
		}
	}
	return errs
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
