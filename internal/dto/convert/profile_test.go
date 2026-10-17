package convert

import (
	"testing"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/NVIDIA/OpenShell/sdk/go/openshell/v1/types"
	"github.com/kenmoini/openshell-proxy/internal/dto"
	"github.com/stretchr/testify/require"
)

func maximalProfile() types.ProviderProfile {
	dur := &types.ProfileDuration{Seconds: 90, Nanos: 5}
	return types.ProviderProfile{
		ID: "github", DisplayName: "GitHub", Description: "d", Category: types.ProfileCategorySourceControl,
		Credentials: []types.ProfileCredential{{
			Name: "token", Description: "t", EnvVars: []string{"A", "B"}, Required: true, StablePlaceholder: true, Secret: true,
			Refresh: &types.ProfileCredentialRefresh{
				Strategy: types.RefreshStrategyOAuth2RefreshToken, TokenURL: "https://t", Scopes: []string{"s1"},
				RefreshBefore: dur, MaxLifetime: &types.ProfileDuration{Seconds: 3600},
				Material:          []types.ProfileCredentialRefreshMaterial{{Name: "m", Description: "md", Required: true, Secret: true}},
				AdditionalOutputs: []types.ProfileCredentialRefreshOutput{{Output: "o", Credential: "c"}},
			},
			AuthStyle: "header", HeaderName: "X-Tok", QueryParam: "q", PathTemplate: "/{t}",
			TokenGrant: &types.CredentialTokenGrant{
				TokenEndpoint: "https://te", Audience: "aud", JWTSVIDAudience: "jwt", Scopes: []string{"x"}, CacheTTL: dur,
				AudienceOverrides:   []types.TokenGrantAudienceOverride{{Host: "h", Port: 443, Path: "/p", Audience: "a", Scopes: []string{"y"}}},
				ClientAssertionType: "cat", GrantType: types.CredentialTokenGrantTypeTokenExchange,
				SubjectToken:       &types.TokenGrantSubjectToken{Source: "s", Credential: "c", SubjectTokenType: "st"},
				RequestedTokenType: "rt",
			},
		}},
		Files:            []types.ProfileFile{{Path: "/p", Content: "c", EnvVar: "E"}},
		Endpoints:        []types.NetworkEndpoint{{Host: "api.github.com", Port: 443, Protocol: "https"}},
		Binaries:         []types.NetworkBinary{{Path: "/usr/bin/git"}},
		InferenceCapable: true,
		Discovery:        types.ProfileDiscovery{Credentials: []string{"GH_TOKEN"}},
		ResourceVersion:  1<<60 + 3,
		Annotations:      map[string]string{"k": "v"},
		Source:           "builtin", Scope: "global",
	}
}

func TestProfile_RoundTrip(t *testing.T) {
	in := maximalProfile()
	d := ProfileFromSDK(&in)
	require.Equal(t, in, ProfileToSDK(d))
	item := ImportItemToSDK(dto.ProfileImportItem{Profile: d, Source: "src"})
	require.Equal(t, types.ProfileImportItem{Profile: in, Source: "src"}, item)
}

func TestResults_Mapping(t *testing.T) {
	p := maximalProfile()
	diag := []types.ProfileDiagnostic{{Source: "s", ProfileID: "github", Field: "f", Message: "m", Severity: "error"}}
	imp := ImportResultFromSDK(&v1.ImportResult{Diagnostics: diag, Profiles: []types.ProviderProfile{p}, Imported: true})
	require.True(t, imp.Imported)
	require.Len(t, imp.Profiles, 1)
	require.Equal(t, "error", imp.Diagnostics[0].Severity)
	empty := ImportResultFromSDK(&v1.ImportResult{})
	require.NotNil(t, empty.Diagnostics)
	require.NotNil(t, empty.Profiles)
	up := UpdateResultFromSDK(&v1.UpdateResult{Profile: &p, Updated: true})
	require.True(t, up.Updated)
	require.Equal(t, "github", up.Profile.ID)
	require.Nil(t, UpdateResultFromSDK(&v1.UpdateResult{}).Profile)
	l := LintResultFromSDK(&v1.LintResult{Diagnostics: diag, Valid: true})
	require.True(t, l.Valid)
	require.Len(t, l.Diagnostics, 1)
}
