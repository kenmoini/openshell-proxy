package dto

// ProfileDuration is a protobuf-style duration.
type ProfileDuration struct {
	Seconds int64 `json:"seconds"`
	Nanos   int32 `json:"nanos,omitempty"`
}

// ProfileFile is a file materialized into a sandbox by a profile.
type ProfileFile struct {
	Path    string `json:"path,omitempty"`
	Content string `json:"content,omitempty"`
	EnvVar  string `json:"envVar,omitempty"`
}

// NetworkEndpoint is a network endpoint a profile needs.
type NetworkEndpoint struct {
	Host     string `json:"host,omitempty"`
	Port     uint32 `json:"port,omitempty"`
	Protocol string `json:"protocol,omitempty"`
}

// NetworkBinary is a binary allowed to use a profile's endpoints.
type NetworkBinary struct {
	Path string `json:"path,omitempty"`
}

// ProfileDiscovery lists credentials discoverable from the environment.
type ProfileDiscovery struct {
	Credentials []string `json:"credentials,omitempty"`
}

// ProfileCredentialRefreshMaterial describes one input of a refresh strategy.
type ProfileCredentialRefreshMaterial struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Secret      bool   `json:"secret,omitempty"`
}

// ProfileCredentialRefreshOutput maps a refresh output to a credential.
type ProfileCredentialRefreshOutput struct {
	Output     string `json:"output,omitempty"`
	Credential string `json:"credential,omitempty"`
}

// ProfileCredentialRefresh declares how a credential is refreshed.
type ProfileCredentialRefresh struct {
	Strategy          string                             `json:"strategy,omitempty"`
	TokenURL          string                             `json:"tokenUrl,omitempty"`
	Scopes            []string                           `json:"scopes,omitempty"`
	RefreshBefore     *ProfileDuration                   `json:"refreshBefore,omitempty"`
	MaxLifetime       *ProfileDuration                   `json:"maxLifetime,omitempty"`
	Material          []ProfileCredentialRefreshMaterial `json:"material,omitempty"`
	AdditionalOutputs []ProfileCredentialRefreshOutput   `json:"additionalOutputs,omitempty"`
}

// TokenGrantAudienceOverride overrides the audience for matching requests.
type TokenGrantAudienceOverride struct {
	Host     string   `json:"host,omitempty"`
	Port     uint32   `json:"port,omitempty"`
	Path     string   `json:"path,omitempty"`
	Audience string   `json:"audience,omitempty"`
	Scopes   []string `json:"scopes,omitempty"`
}

// TokenGrantSubjectToken describes the subject token of a token exchange.
type TokenGrantSubjectToken struct {
	Source           string `json:"source,omitempty"`
	Credential       string `json:"credential,omitempty"`
	SubjectTokenType string `json:"subjectTokenType,omitempty"`
}

// CredentialTokenGrant configures token-grant credential acquisition.
type CredentialTokenGrant struct {
	TokenEndpoint       string                       `json:"tokenEndpoint,omitempty"`
	Audience            string                       `json:"audience,omitempty"`
	JWTSVIDAudience     string                       `json:"jwtSvidAudience,omitempty"`
	Scopes              []string                     `json:"scopes,omitempty"`
	CacheTTL            *ProfileDuration             `json:"cacheTtl,omitempty"`
	AudienceOverrides   []TokenGrantAudienceOverride `json:"audienceOverrides,omitempty"`
	ClientAssertionType string                       `json:"clientAssertionType,omitempty"`
	GrantType           string                       `json:"grantType,omitempty"`
	SubjectToken        *TokenGrantSubjectToken      `json:"subjectToken,omitempty"`
	RequestedTokenType  string                       `json:"requestedTokenType,omitempty"`
}

// ProfileCredential declares one credential of a provider profile.
type ProfileCredential struct {
	Name              string                    `json:"name,omitempty"`
	Description       string                    `json:"description,omitempty"`
	EnvVars           []string                  `json:"envVars,omitempty"`
	Required          bool                      `json:"required,omitempty"`
	StablePlaceholder bool                      `json:"stablePlaceholder,omitempty"`
	Secret            bool                      `json:"secret,omitempty"`
	Refresh           *ProfileCredentialRefresh `json:"refresh,omitempty"`
	AuthStyle         string                    `json:"authStyle,omitempty"`
	HeaderName        string                    `json:"headerName,omitempty"`
	QueryParam        string                    `json:"queryParam,omitempty"`
	PathTemplate      string                    `json:"pathTemplate,omitempty"`
	TokenGrant        *CredentialTokenGrant     `json:"tokenGrant,omitempty"`
}

// ProviderProfile is a provider type definition.
type ProviderProfile struct {
	ID               string              `json:"id"`
	DisplayName      string              `json:"displayName,omitempty"`
	Description      string              `json:"description,omitempty"`
	Category         string              `json:"category,omitempty"`
	Credentials      []ProfileCredential `json:"credentials,omitempty"`
	Files            []ProfileFile       `json:"files,omitempty"`
	Endpoints        []NetworkEndpoint   `json:"endpoints,omitempty"`
	Binaries         []NetworkBinary     `json:"binaries,omitempty"`
	InferenceCapable bool                `json:"inferenceCapable,omitempty"`
	Discovery        ProfileDiscovery    `json:"discovery,omitempty"`
	ResourceVersion  Uint64String        `json:"resourceVersion,omitempty"`
	Annotations      map[string]string   `json:"annotations,omitempty"`
	Source           string              `json:"source,omitempty"`
	Scope            string              `json:"scope,omitempty"`
}

// ProfileImportItem is one profile submitted for import, update or lint.
type ProfileImportItem struct {
	Profile ProviderProfile `json:"profile"`
	Source  string          `json:"source,omitempty"`
}

// ProfileImportRequest is the body of the import and lint operations.
type ProfileImportRequest struct {
	Items []ProfileImportItem `json:"items"`
}

// ProfileDiagnostic is one finding from profile validation.
type ProfileDiagnostic struct {
	Source    string `json:"source,omitempty"`
	ProfileID string `json:"profileId,omitempty"`
	Field     string `json:"field,omitempty"`
	Message   string `json:"message,omitempty"`
	Severity  string `json:"severity,omitempty"`
}

// ProfileImportResult is the outcome of an import.
type ProfileImportResult struct {
	Diagnostics []ProfileDiagnostic `json:"diagnostics"`
	Profiles    []ProviderProfile   `json:"profiles"`
	Imported    bool                `json:"imported"`
}

// ProfileUpdateResult is the outcome of a profile update.
type ProfileUpdateResult struct {
	Diagnostics []ProfileDiagnostic `json:"diagnostics"`
	Profile     *ProviderProfile    `json:"profile,omitempty"`
	Updated     bool                `json:"updated"`
}

// ProfileLintResult is the outcome of a lint.
type ProfileLintResult struct {
	Diagnostics []ProfileDiagnostic `json:"diagnostics"`
	Valid       bool                `json:"valid"`
}
