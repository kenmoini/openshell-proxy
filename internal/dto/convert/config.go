package convert

import (
	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/dto"
)

// SettingValueFromSDK maps an SDK setting value, setting only the member
// matching its type.
func SettingValueFromSDK(v v1.SettingValue) dto.SettingValue {
	out := dto.SettingValue{Type: string(v.Type)}
	switch v.Type {
	case v1.SettingValueString:
		s := v.StringVal
		out.String = &s
	case v1.SettingValueBool:
		b := v.BoolVal
		out.Bool = &b
	case v1.SettingValueInt:
		i := v.IntVal
		out.Int = &i
	case v1.SettingValueBytes:
		out.Bytes = append([]byte{}, v.BytesVal...)
	}
	return out
}

// SettingValueToSDK maps a DTO setting value, reading only the member
// matching its type.
func SettingValueToSDK(v dto.SettingValue) v1.SettingValue {
	out := v1.SettingValue{Type: v1.SettingValueType(v.Type)}
	switch out.Type {
	case v1.SettingValueString:
		if v.String != nil {
			out.StringVal = *v.String
		}
	case v1.SettingValueBool:
		if v.Bool != nil {
			out.BoolVal = *v.Bool
		}
	case v1.SettingValueInt:
		if v.Int != nil {
			out.IntVal = *v.Int
		}
	case v1.SettingValueBytes:
		out.BytesVal = v.Bytes
	}
	return out
}

// SandboxConfigFromSDK maps an SDK sandbox config.
func SandboxConfigFromSDK(c *v1.SandboxConfig) dto.SandboxConfig {
	out := dto.SandboxConfig{
		Policy:                      PolicyFromSDK(c.Policy),
		PolicyVersion:               c.PolicyVersion,
		PolicyHash:                  c.PolicyHash,
		Settings:                    make(map[string]dto.EffectiveSetting, len(c.Settings)),
		ConfigRevision:              dto.Uint64String(c.ConfigRevision),
		PolicySource:                string(c.PolicySource),
		GlobalPolicyVersion:         c.GlobalPolicyVersion,
		ProviderEnvRevision:         dto.Uint64String(c.ProviderEnvRevision),
		PolicyValidationFailureMode: c.PolicyValidationFailureMode,
	}
	for k, s := range c.Settings {
		out.Settings[k] = dto.EffectiveSetting{Value: SettingValueFromSDK(s.Value), Scope: string(s.Scope)}
	}
	return out
}

// GatewayConfigFromSDK maps an SDK gateway config.
func GatewayConfigFromSDK(c *v1.GatewayConfig) dto.GatewayConfig {
	out := dto.GatewayConfig{
		Settings:         make(map[string]dto.SettingValue, len(c.Settings)),
		SettingsRevision: dto.Uint64String(c.SettingsRevision),
	}
	for k, s := range c.Settings {
		out.Settings[k] = SettingValueFromSDK(s)
	}
	return out
}

// ConfigUpdateToSDK maps an update request. name and global are set by the
// route; expected is the already-resolved expected resource version.
func ConfigUpdateToSDK(r dto.ConfigUpdateRequest, name string, global bool, expected uint64) *v1.ConfigUpdate {
	u := &v1.ConfigUpdate{
		Name:                    name,
		Policy:                  PolicyToSDK(r.Policy),
		SettingKey:              r.SettingKey,
		DeleteSetting:           r.DeleteSetting,
		Global:                  global,
		MergeOperations:         MergeOpsToSDK(r.MergeOperations),
		ExpectedResourceVersion: expected,
		Annotations:             r.Annotations,
	}
	if r.SettingValue != nil {
		sv := SettingValueToSDK(*r.SettingValue)
		u.SettingValue = &sv
	}
	return u
}

// ConfigUpdateResultFromSDK maps an SDK update result.
func ConfigUpdateResultFromSDK(r *v1.ConfigUpdateResult) dto.ConfigUpdateResult {
	return dto.ConfigUpdateResult{
		Version:          r.Version,
		PolicyHash:       r.PolicyHash,
		SettingsRevision: dto.Uint64String(r.SettingsRevision),
		Deleted:          r.Deleted,
		Annotations:      r.Annotations,
	}
}
