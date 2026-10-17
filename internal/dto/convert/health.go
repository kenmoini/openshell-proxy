// Package convert maps OpenShell SDK types to API DTOs.
package convert

import (
	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/dto"
)

// Health converts an SDK health result.
func Health(r *v1.HealthResult) dto.Health {
	if r == nil {
		return dto.Health{}
	}
	return dto.Health{Healthy: r.Healthy, Version: r.Version}
}

// GatewayInfo converts SDK gateway metadata. Slices are never nil.
func GatewayInfo(g *v1.GatewayInfo) dto.GatewayInfo {
	out := dto.GatewayInfo{ComputeDrivers: []dto.ComputeDriverInfo{}, Extensions: []dto.ExtensionInfo{}}
	if g == nil {
		return out
	}
	out.Status = string(g.Status)
	out.Version = g.Version
	for _, d := range g.ComputeDrivers {
		out.ComputeDrivers = append(out.ComputeDrivers, dto.ComputeDriverInfo{
			Name: d.Name, DriverName: d.DriverName, DriverVersion: d.DriverVersion,
		})
	}
	for _, e := range g.Extensions {
		out.Extensions = append(out.Extensions, dto.ExtensionInfo{
			Kind:                  string(e.Kind),
			ConfiguredName:        e.ConfiguredName,
			ImplementationName:    e.ImplementationName,
			ImplementationVersion: e.ImplementationVersion,
			ProtocolMajor:         e.ProtocolMajor,
			ProtocolMinor:         e.ProtocolMinor,
			SupportedCapabilities: e.SupportedCapabilities,
			RequiredCapabilities:  e.RequiredCapabilities,
		})
	}
	return out
}

// CurrentUser converts the SDK caller identity. Slices are never nil.
func CurrentUser(u *v1.CurrentUser) dto.CurrentUser {
	out := dto.CurrentUser{Roles: []string{}, Scopes: []string{}}
	if u == nil {
		return out
	}
	out.Subject = u.Subject
	out.DisplayName = u.DisplayName
	out.IdentityProvider = u.IdentityProvider
	out.Roles = append(out.Roles, u.Roles...)
	out.Scopes = append(out.Scopes, u.Scopes...)
	return out
}
