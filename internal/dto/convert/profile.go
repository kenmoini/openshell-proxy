package convert

import (
	"github.com/NVIDIA/OpenShell/sdk/go/openshell/v1/types"
	"github.com/kenmoini/openshell-proxy/internal/dto"
)

func durationFromSDK(d *types.ProfileDuration) *dto.ProfileDuration {
	if d == nil {
		return nil
	}
	return &dto.ProfileDuration{Seconds: d.Seconds, Nanos: d.Nanos}
}

func durationToSDK(d *dto.ProfileDuration) *types.ProfileDuration {
	if d == nil {
		return nil
	}
	return &types.ProfileDuration{Seconds: d.Seconds, Nanos: d.Nanos}
}

func refreshFromSDK(r *types.ProfileCredentialRefresh) *dto.ProfileCredentialRefresh {
	if r == nil {
		return nil
	}
	out := &dto.ProfileCredentialRefresh{
		Strategy:      string(r.Strategy),
		TokenURL:      r.TokenURL,
		Scopes:        r.Scopes,
		RefreshBefore: durationFromSDK(r.RefreshBefore),
		MaxLifetime:   durationFromSDK(r.MaxLifetime),
	}
	for _, m := range r.Material {
		out.Material = append(out.Material, dto.ProfileCredentialRefreshMaterial{
			Name: m.Name, Description: m.Description, Required: m.Required, Secret: m.Secret,
		})
	}
	for _, o := range r.AdditionalOutputs {
		out.AdditionalOutputs = append(out.AdditionalOutputs, dto.ProfileCredentialRefreshOutput{Output: o.Output, Credential: o.Credential})
	}
	return out
}

func refreshToSDK(r *dto.ProfileCredentialRefresh) *types.ProfileCredentialRefresh {
	if r == nil {
		return nil
	}
	out := &types.ProfileCredentialRefresh{
		Strategy:      types.RefreshStrategy(r.Strategy),
		TokenURL:      r.TokenURL,
		Scopes:        r.Scopes,
		RefreshBefore: durationToSDK(r.RefreshBefore),
		MaxLifetime:   durationToSDK(r.MaxLifetime),
	}
	for _, m := range r.Material {
		out.Material = append(out.Material, types.ProfileCredentialRefreshMaterial{
			Name: m.Name, Description: m.Description, Required: m.Required, Secret: m.Secret,
		})
	}
	for _, o := range r.AdditionalOutputs {
		out.AdditionalOutputs = append(out.AdditionalOutputs, types.ProfileCredentialRefreshOutput{Output: o.Output, Credential: o.Credential})
	}
	return out
}

func grantFromSDK(g *types.CredentialTokenGrant) *dto.CredentialTokenGrant {
	if g == nil {
		return nil
	}
	out := &dto.CredentialTokenGrant{
		TokenEndpoint:       g.TokenEndpoint,
		Audience:            g.Audience,
		JWTSVIDAudience:     g.JWTSVIDAudience,
		Scopes:              g.Scopes,
		CacheTTL:            durationFromSDK(g.CacheTTL),
		ClientAssertionType: g.ClientAssertionType,
		GrantType:           string(g.GrantType),
		RequestedTokenType:  g.RequestedTokenType,
	}
	for _, o := range g.AudienceOverrides {
		out.AudienceOverrides = append(out.AudienceOverrides, dto.TokenGrantAudienceOverride{
			Host: o.Host, Port: o.Port, Path: o.Path, Audience: o.Audience, Scopes: o.Scopes,
		})
	}
	if s := g.SubjectToken; s != nil {
		out.SubjectToken = &dto.TokenGrantSubjectToken{Source: s.Source, Credential: s.Credential, SubjectTokenType: s.SubjectTokenType}
	}
	return out
}

func grantToSDK(g *dto.CredentialTokenGrant) *types.CredentialTokenGrant {
	if g == nil {
		return nil
	}
	out := &types.CredentialTokenGrant{
		TokenEndpoint:       g.TokenEndpoint,
		Audience:            g.Audience,
		JWTSVIDAudience:     g.JWTSVIDAudience,
		Scopes:              g.Scopes,
		CacheTTL:            durationToSDK(g.CacheTTL),
		ClientAssertionType: g.ClientAssertionType,
		GrantType:           types.CredentialTokenGrantType(g.GrantType),
		RequestedTokenType:  g.RequestedTokenType,
	}
	for _, o := range g.AudienceOverrides {
		out.AudienceOverrides = append(out.AudienceOverrides, types.TokenGrantAudienceOverride{
			Host: o.Host, Port: o.Port, Path: o.Path, Audience: o.Audience, Scopes: o.Scopes,
		})
	}
	if s := g.SubjectToken; s != nil {
		out.SubjectToken = &types.TokenGrantSubjectToken{Source: s.Source, Credential: s.Credential, SubjectTokenType: s.SubjectTokenType}
	}
	return out
}

// ProfileFromSDK maps an SDK provider profile to its DTO.
func ProfileFromSDK(p *types.ProviderProfile) dto.ProviderProfile {
	out := dto.ProviderProfile{
		ID:               p.ID,
		DisplayName:      p.DisplayName,
		Description:      p.Description,
		Category:         string(p.Category),
		InferenceCapable: p.InferenceCapable,
		Discovery:        dto.ProfileDiscovery{Credentials: p.Discovery.Credentials},
		ResourceVersion:  dto.Uint64String(p.ResourceVersion),
		Annotations:      p.Annotations,
		Source:           p.Source,
		Scope:            p.Scope,
	}
	for _, c := range p.Credentials {
		out.Credentials = append(out.Credentials, dto.ProfileCredential{
			Name: c.Name, Description: c.Description, EnvVars: c.EnvVars, Required: c.Required,
			StablePlaceholder: c.StablePlaceholder, Secret: c.Secret, Refresh: refreshFromSDK(c.Refresh),
			AuthStyle: c.AuthStyle, HeaderName: c.HeaderName, QueryParam: c.QueryParam,
			PathTemplate: c.PathTemplate, TokenGrant: grantFromSDK(c.TokenGrant),
		})
	}
	for _, f := range p.Files {
		out.Files = append(out.Files, dto.ProfileFile{Path: f.Path, Content: f.Content, EnvVar: f.EnvVar})
	}
	for _, e := range p.Endpoints {
		out.Endpoints = append(out.Endpoints, dto.NetworkEndpoint{Host: e.Host, Port: e.Port, Protocol: e.Protocol})
	}
	for _, b := range p.Binaries {
		out.Binaries = append(out.Binaries, dto.NetworkBinary{Path: b.Path})
	}
	return out
}

// ProfileToSDK maps a profile DTO to the SDK type.
func ProfileToSDK(p dto.ProviderProfile) types.ProviderProfile {
	out := types.ProviderProfile{
		ID:               p.ID,
		DisplayName:      p.DisplayName,
		Description:      p.Description,
		Category:         types.ProfileCategory(p.Category),
		InferenceCapable: p.InferenceCapable,
		Discovery:        types.ProfileDiscovery{Credentials: p.Discovery.Credentials},
		ResourceVersion:  uint64(p.ResourceVersion),
		Annotations:      p.Annotations,
		Source:           p.Source,
		Scope:            p.Scope,
	}
	for _, c := range p.Credentials {
		out.Credentials = append(out.Credentials, types.ProfileCredential{
			Name: c.Name, Description: c.Description, EnvVars: c.EnvVars, Required: c.Required,
			StablePlaceholder: c.StablePlaceholder, Secret: c.Secret, Refresh: refreshToSDK(c.Refresh),
			AuthStyle: c.AuthStyle, HeaderName: c.HeaderName, QueryParam: c.QueryParam,
			PathTemplate: c.PathTemplate, TokenGrant: grantToSDK(c.TokenGrant),
		})
	}
	for _, f := range p.Files {
		out.Files = append(out.Files, types.ProfileFile{Path: f.Path, Content: f.Content, EnvVar: f.EnvVar})
	}
	for _, e := range p.Endpoints {
		out.Endpoints = append(out.Endpoints, types.NetworkEndpoint{Host: e.Host, Port: e.Port, Protocol: e.Protocol})
	}
	for _, b := range p.Binaries {
		out.Binaries = append(out.Binaries, types.NetworkBinary{Path: b.Path})
	}
	return out
}

// ImportItemToSDK maps an import item DTO to the SDK type.
func ImportItemToSDK(i dto.ProfileImportItem) types.ProfileImportItem {
	return types.ProfileImportItem{Profile: ProfileToSDK(i.Profile), Source: i.Source}
}

// ImportItemsToSDK maps a list of import items.
func ImportItemsToSDK(in []dto.ProfileImportItem) []types.ProfileImportItem {
	out := make([]types.ProfileImportItem, 0, len(in))
	for _, i := range in {
		out = append(out, ImportItemToSDK(i))
	}
	return out
}

func diagnosticsFromSDK(in []types.ProfileDiagnostic) []dto.ProfileDiagnostic {
	out := make([]dto.ProfileDiagnostic, 0, len(in))
	for _, d := range in {
		out = append(out, dto.ProfileDiagnostic{Source: d.Source, ProfileID: d.ProfileID, Field: d.Field, Message: d.Message, Severity: d.Severity})
	}
	return out
}

// ImportResultFromSDK maps an import result; slices are never nil.
func ImportResultFromSDK(r *types.ImportResult) dto.ProfileImportResult {
	out := dto.ProfileImportResult{Diagnostics: diagnosticsFromSDK(r.Diagnostics), Profiles: make([]dto.ProviderProfile, 0, len(r.Profiles)), Imported: r.Imported}
	for i := range r.Profiles {
		out.Profiles = append(out.Profiles, ProfileFromSDK(&r.Profiles[i]))
	}
	return out
}

// UpdateResultFromSDK maps an update result.
func UpdateResultFromSDK(r *types.UpdateResult) dto.ProfileUpdateResult {
	out := dto.ProfileUpdateResult{Diagnostics: diagnosticsFromSDK(r.Diagnostics), Updated: r.Updated}
	if r.Profile != nil {
		p := ProfileFromSDK(r.Profile)
		out.Profile = &p
	}
	return out
}

// LintResultFromSDK maps a lint result.
func LintResultFromSDK(r *types.LintResult) dto.ProfileLintResult {
	return dto.ProfileLintResult{Diagnostics: diagnosticsFromSDK(r.Diagnostics), Valid: r.Valid}
}
