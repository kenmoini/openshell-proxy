package gatewaytest

import (
	"context"
	"io"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
)

// ServiceStub is a function-field stub of v1.ServiceInterface.
type ServiceStub struct {
	ExposeFn  func(ctx context.Context, workspace, sandboxName, serviceName string, targetPort uint32, domain bool, opts ...v1.ExposeServiceOptions) (*v1.ServiceEndpoint, error)
	GetFn     func(ctx context.Context, workspace, sandboxName, serviceName string) (*v1.ServiceEndpoint, error)
	ListFn    func(workspace, sandboxName string, opts ...v1.ListOptions) (*v1.Pager[*v1.ServiceEndpoint], error)
	ListAllFn func(ctx context.Context, workspace, sandboxName string, opts ...v1.ListOptions) ([]*v1.ServiceEndpoint, error)
	DeleteFn  func(ctx context.Context, workspace, sandboxName, serviceName string, opts ...v1.DeleteOptions) (*v1.DeletionResult, error)
}

var _ v1.ServiceInterface = (*ServiceStub)(nil)

func (s *ServiceStub) Expose(ctx context.Context, workspace, sandboxName, serviceName string, targetPort uint32, domain bool, opts ...v1.ExposeServiceOptions) (*v1.ServiceEndpoint, error) {
	if s.ExposeFn == nil {
		return nil, ErrUnimplemented
	}
	return s.ExposeFn(ctx, workspace, sandboxName, serviceName, targetPort, domain, opts...)
}

func (s *ServiceStub) Get(ctx context.Context, workspace, sandboxName, serviceName string) (*v1.ServiceEndpoint, error) {
	if s.GetFn == nil {
		return nil, ErrUnimplemented
	}
	return s.GetFn(ctx, workspace, sandboxName, serviceName)
}

func (s *ServiceStub) List(workspace, sandboxName string, opts ...v1.ListOptions) (*v1.Pager[*v1.ServiceEndpoint], error) {
	if s.ListFn == nil {
		return nil, ErrUnimplemented
	}
	return s.ListFn(workspace, sandboxName, opts...)
}

func (s *ServiceStub) ListAll(ctx context.Context, workspace, sandboxName string, opts ...v1.ListOptions) ([]*v1.ServiceEndpoint, error) {
	if s.ListAllFn == nil {
		return nil, ErrUnimplemented
	}
	return s.ListAllFn(ctx, workspace, sandboxName, opts...)
}

func (s *ServiceStub) Delete(ctx context.Context, workspace, sandboxName, serviceName string, opts ...v1.DeleteOptions) (*v1.DeletionResult, error) {
	if s.DeleteFn == nil {
		return nil, ErrUnimplemented
	}
	return s.DeleteFn(ctx, workspace, sandboxName, serviceName, opts...)
}

// ConfigStub is a function-field stub of v1.ConfigInterface.
type ConfigStub struct {
	GetSandboxFn func(ctx context.Context, workspace, sandboxName string) (*v1.SandboxConfig, error)
	GetGatewayFn func(ctx context.Context) (*v1.GatewayConfig, error)
	UpdateFn     func(ctx context.Context, workspace string, update *v1.ConfigUpdate) (*v1.ConfigUpdateResult, error)
}

var _ v1.ConfigInterface = (*ConfigStub)(nil)

func (s *ConfigStub) GetSandbox(ctx context.Context, workspace, sandboxName string) (*v1.SandboxConfig, error) {
	if s.GetSandboxFn == nil {
		return nil, ErrUnimplemented
	}
	return s.GetSandboxFn(ctx, workspace, sandboxName)
}

func (s *ConfigStub) GetGateway(ctx context.Context) (*v1.GatewayConfig, error) {
	if s.GetGatewayFn == nil {
		return nil, ErrUnimplemented
	}
	return s.GetGatewayFn(ctx)
}

func (s *ConfigStub) Update(ctx context.Context, workspace string, update *v1.ConfigUpdate) (*v1.ConfigUpdateResult, error) {
	if s.UpdateFn == nil {
		return nil, ErrUnimplemented
	}
	return s.UpdateFn(ctx, workspace, update)
}

// PolicyStub is a function-field stub of v1.PolicyInterface.
type PolicyStub struct {
	GetDraftFn              func(ctx context.Context, workspace, sandboxName string, opts ...v1.GetDraftOption) (*v1.DraftPolicy, error)
	ApproveDraftChunkFn     func(ctx context.Context, workspace, sandboxName, chunkID, reviewToken string) (*v1.ApproveResult, error)
	RejectDraftChunkFn      func(ctx context.Context, workspace, sandboxName, chunkID, reason string) error
	ApproveAllDraftChunksFn func(ctx context.Context, workspace, sandboxName string, opts ...v1.ApproveAllOption) (*v1.ApproveAllResult, error)
	ClearDraftChunksFn      func(ctx context.Context, workspace, sandboxName string) (*v1.ClearResult, error)
	GetDraftHistoryFn       func(ctx context.Context, workspace, sandboxName string) ([]v1.DraftHistoryEntry, error)
	GetStatusFn             func(ctx context.Context, workspace, sandboxName string, opts ...v1.GetStatusOption) (*v1.PolicyStatusResult, error)
	ListFn                  func(workspace, sandboxName string, opts ...v1.ListPolicyOption) (*v1.Pager[v1.SandboxPolicyRevision], error)
	ListAllFn               func(ctx context.Context, workspace, sandboxName string, opts ...v1.ListPolicyOption) ([]v1.SandboxPolicyRevision, error)
	EditDraftChunkFn        func(ctx context.Context, workspace, sandboxName, chunkID string, proposedRule *v1.NetworkPolicyRule) error
	UndoDraftChunkFn        func(ctx context.Context, workspace, sandboxName, chunkID string) (*v1.UndoResult, error)
}

var _ v1.PolicyInterface = (*PolicyStub)(nil)

func (s *PolicyStub) GetDraft(ctx context.Context, workspace, sandboxName string, opts ...v1.GetDraftOption) (*v1.DraftPolicy, error) {
	if s.GetDraftFn == nil {
		return nil, ErrUnimplemented
	}
	return s.GetDraftFn(ctx, workspace, sandboxName, opts...)
}

func (s *PolicyStub) ApproveDraftChunk(ctx context.Context, workspace, sandboxName, chunkID, reviewToken string) (*v1.ApproveResult, error) {
	if s.ApproveDraftChunkFn == nil {
		return nil, ErrUnimplemented
	}
	return s.ApproveDraftChunkFn(ctx, workspace, sandboxName, chunkID, reviewToken)
}

func (s *PolicyStub) RejectDraftChunk(ctx context.Context, workspace, sandboxName, chunkID, reason string) error {
	if s.RejectDraftChunkFn == nil {
		return ErrUnimplemented
	}
	return s.RejectDraftChunkFn(ctx, workspace, sandboxName, chunkID, reason)
}

func (s *PolicyStub) ApproveAllDraftChunks(ctx context.Context, workspace, sandboxName string, opts ...v1.ApproveAllOption) (*v1.ApproveAllResult, error) {
	if s.ApproveAllDraftChunksFn == nil {
		return nil, ErrUnimplemented
	}
	return s.ApproveAllDraftChunksFn(ctx, workspace, sandboxName, opts...)
}

func (s *PolicyStub) ClearDraftChunks(ctx context.Context, workspace, sandboxName string) (*v1.ClearResult, error) {
	if s.ClearDraftChunksFn == nil {
		return nil, ErrUnimplemented
	}
	return s.ClearDraftChunksFn(ctx, workspace, sandboxName)
}

func (s *PolicyStub) GetDraftHistory(ctx context.Context, workspace, sandboxName string) ([]v1.DraftHistoryEntry, error) {
	if s.GetDraftHistoryFn == nil {
		return nil, ErrUnimplemented
	}
	return s.GetDraftHistoryFn(ctx, workspace, sandboxName)
}

func (s *PolicyStub) GetStatus(ctx context.Context, workspace, sandboxName string, opts ...v1.GetStatusOption) (*v1.PolicyStatusResult, error) {
	if s.GetStatusFn == nil {
		return nil, ErrUnimplemented
	}
	return s.GetStatusFn(ctx, workspace, sandboxName, opts...)
}

func (s *PolicyStub) List(workspace, sandboxName string, opts ...v1.ListPolicyOption) (*v1.Pager[v1.SandboxPolicyRevision], error) {
	if s.ListFn == nil {
		return nil, ErrUnimplemented
	}
	return s.ListFn(workspace, sandboxName, opts...)
}

func (s *PolicyStub) ListAll(ctx context.Context, workspace, sandboxName string, opts ...v1.ListPolicyOption) ([]v1.SandboxPolicyRevision, error) {
	if s.ListAllFn == nil {
		return nil, ErrUnimplemented
	}
	return s.ListAllFn(ctx, workspace, sandboxName, opts...)
}

func (s *PolicyStub) EditDraftChunk(ctx context.Context, workspace, sandboxName, chunkID string, proposedRule *v1.NetworkPolicyRule) error {
	if s.EditDraftChunkFn == nil {
		return ErrUnimplemented
	}
	return s.EditDraftChunkFn(ctx, workspace, sandboxName, chunkID, proposedRule)
}

func (s *PolicyStub) UndoDraftChunk(ctx context.Context, workspace, sandboxName, chunkID string) (*v1.UndoResult, error) {
	if s.UndoDraftChunkFn == nil {
		return nil, ErrUnimplemented
	}
	return s.UndoDraftChunkFn(ctx, workspace, sandboxName, chunkID)
}

// ProfileStub is a function-field stub of v1.ProfileInterface.
type ProfileStub struct {
	ListFn    func(workspace string, opts ...v1.ListOptions) (*v1.Pager[*v1.ProviderProfile], error)
	ListAllFn func(ctx context.Context, workspace string, opts ...v1.ListOptions) ([]*v1.ProviderProfile, error)
	GetFn     func(ctx context.Context, workspace, id string) (*v1.ProviderProfile, error)
	ImportFn  func(ctx context.Context, workspace string, items []v1.ProfileImportItem) (*v1.ImportResult, error)
	UpdateFn  func(ctx context.Context, workspace, id string, expectedResourceVersion uint64, item v1.ProfileImportItem) (*v1.UpdateResult, error)
	LintFn    func(ctx context.Context, workspace string, items []v1.ProfileImportItem) (*v1.LintResult, error)
	DeleteFn  func(ctx context.Context, workspace, id string, opts ...v1.DeleteOptions) (*v1.DeletionResult, error)
}

var _ v1.ProfileInterface = (*ProfileStub)(nil)

func (s *ProfileStub) List(workspace string, opts ...v1.ListOptions) (*v1.Pager[*v1.ProviderProfile], error) {
	if s.ListFn == nil {
		return nil, ErrUnimplemented
	}
	return s.ListFn(workspace, opts...)
}

func (s *ProfileStub) ListAll(ctx context.Context, workspace string, opts ...v1.ListOptions) ([]*v1.ProviderProfile, error) {
	if s.ListAllFn == nil {
		return nil, ErrUnimplemented
	}
	return s.ListAllFn(ctx, workspace, opts...)
}

func (s *ProfileStub) Get(ctx context.Context, workspace, id string) (*v1.ProviderProfile, error) {
	if s.GetFn == nil {
		return nil, ErrUnimplemented
	}
	return s.GetFn(ctx, workspace, id)
}

func (s *ProfileStub) Import(ctx context.Context, workspace string, items []v1.ProfileImportItem) (*v1.ImportResult, error) {
	if s.ImportFn == nil {
		return nil, ErrUnimplemented
	}
	return s.ImportFn(ctx, workspace, items)
}

func (s *ProfileStub) Update(ctx context.Context, workspace, id string, expectedResourceVersion uint64, item v1.ProfileImportItem) (*v1.UpdateResult, error) {
	if s.UpdateFn == nil {
		return nil, ErrUnimplemented
	}
	return s.UpdateFn(ctx, workspace, id, expectedResourceVersion, item)
}

func (s *ProfileStub) Lint(ctx context.Context, workspace string, items []v1.ProfileImportItem) (*v1.LintResult, error) {
	if s.LintFn == nil {
		return nil, ErrUnimplemented
	}
	return s.LintFn(ctx, workspace, items)
}

func (s *ProfileStub) Delete(ctx context.Context, workspace, id string, opts ...v1.DeleteOptions) (*v1.DeletionResult, error) {
	if s.DeleteFn == nil {
		return nil, ErrUnimplemented
	}
	return s.DeleteFn(ctx, workspace, id, opts...)
}

// RefreshStub is a function-field stub of v1.RefreshInterface.
type RefreshStub struct {
	GetStatusFn func(ctx context.Context, workspace, provider, credentialKey string) ([]*v1.RefreshStatus, error)
	ConfigureFn func(ctx context.Context, workspace string, config *v1.RefreshConfig) (*v1.RefreshStatus, error)
	RotateFn    func(ctx context.Context, workspace, provider, credentialKey string) (*v1.RefreshStatus, error)
	DeleteFn    func(ctx context.Context, workspace, provider, credentialKey string, opts ...v1.DeleteOptions) (*v1.DeletionResult, error)
}

var _ v1.RefreshInterface = (*RefreshStub)(nil)

func (s *RefreshStub) GetStatus(ctx context.Context, workspace, provider, credentialKey string) ([]*v1.RefreshStatus, error) {
	if s.GetStatusFn == nil {
		return nil, ErrUnimplemented
	}
	return s.GetStatusFn(ctx, workspace, provider, credentialKey)
}

func (s *RefreshStub) Configure(ctx context.Context, workspace string, config *v1.RefreshConfig) (*v1.RefreshStatus, error) {
	if s.ConfigureFn == nil {
		return nil, ErrUnimplemented
	}
	return s.ConfigureFn(ctx, workspace, config)
}

func (s *RefreshStub) Rotate(ctx context.Context, workspace, provider, credentialKey string) (*v1.RefreshStatus, error) {
	if s.RotateFn == nil {
		return nil, ErrUnimplemented
	}
	return s.RotateFn(ctx, workspace, provider, credentialKey)
}

func (s *RefreshStub) Delete(ctx context.Context, workspace, provider, credentialKey string, opts ...v1.DeleteOptions) (*v1.DeletionResult, error) {
	if s.DeleteFn == nil {
		return nil, ErrUnimplemented
	}
	return s.DeleteFn(ctx, workspace, provider, credentialKey, opts...)
}

// ExecStub is a function-field stub of v1.ExecInterface.
type ExecStub struct {
	RunFn         func(ctx context.Context, workspace, sandboxName string, command []string, opts ...v1.ExecOptions) (*v1.ExecResult, error)
	StreamFn      func(ctx context.Context, workspace, sandboxName string, command []string, opts ...v1.ExecOptions) (v1.ExecStream, error)
	InteractiveFn func(ctx context.Context, workspace, sandboxName string, command []string, cols, rows uint32, opts ...v1.ExecOptions) (v1.InteractiveSession, error)
}

var _ v1.ExecInterface = (*ExecStub)(nil)

func (s *ExecStub) Run(ctx context.Context, workspace, sandboxName string, command []string, opts ...v1.ExecOptions) (*v1.ExecResult, error) {
	if s.RunFn == nil {
		return nil, ErrUnimplemented
	}
	return s.RunFn(ctx, workspace, sandboxName, command, opts...)
}

func (s *ExecStub) Stream(ctx context.Context, workspace, sandboxName string, command []string, opts ...v1.ExecOptions) (v1.ExecStream, error) {
	if s.StreamFn == nil {
		return nil, ErrUnimplemented
	}
	return s.StreamFn(ctx, workspace, sandboxName, command, opts...)
}

func (s *ExecStub) Interactive(ctx context.Context, workspace, sandboxName string, command []string, cols, rows uint32, opts ...v1.ExecOptions) (v1.InteractiveSession, error) {
	if s.InteractiveFn == nil {
		return nil, ErrUnimplemented
	}
	return s.InteractiveFn(ctx, workspace, sandboxName, command, cols, rows, opts...)
}

// SSHStub is a function-field stub of v1.SSHInterface.
type SSHStub struct {
	CreateSessionFn func(ctx context.Context, workspace, sandboxName string) (*v1.SSHSession, error)
	RevokeSessionFn func(ctx context.Context, workspace, token string, opts ...v1.DeleteOptions) (*v1.DeletionResult, error)
	TunnelFn        func(ctx context.Context, workspace, sandboxName string, port uint32, opts ...v1.TunnelOption) (io.ReadWriteCloser, error)
}

var _ v1.SSHInterface = (*SSHStub)(nil)

func (s *SSHStub) CreateSession(ctx context.Context, workspace, sandboxName string) (*v1.SSHSession, error) {
	if s.CreateSessionFn == nil {
		return nil, ErrUnimplemented
	}
	return s.CreateSessionFn(ctx, workspace, sandboxName)
}

func (s *SSHStub) RevokeSession(ctx context.Context, workspace, token string, opts ...v1.DeleteOptions) (*v1.DeletionResult, error) {
	if s.RevokeSessionFn == nil {
		return nil, ErrUnimplemented
	}
	return s.RevokeSessionFn(ctx, workspace, token, opts...)
}

func (s *SSHStub) Tunnel(ctx context.Context, workspace, sandboxName string, port uint32, opts ...v1.TunnelOption) (io.ReadWriteCloser, error) {
	if s.TunnelFn == nil {
		return nil, ErrUnimplemented
	}
	return s.TunnelFn(ctx, workspace, sandboxName, port, opts...)
}

// TCPStub is a function-field stub of v1.TCPInterface.
type TCPStub struct {
	ForwardFn func(ctx context.Context, workspace, sandboxName string, port uint32, opts ...v1.ForwardOption) (io.ReadWriteCloser, error)
	ListenFn  func(ctx context.Context, workspace, sandboxName string, remotePort uint32, localPort uint32, opts ...v1.ListenOption) (v1.ForwardListener, error)
}

var _ v1.TCPInterface = (*TCPStub)(nil)

func (s *TCPStub) Forward(ctx context.Context, workspace, sandboxName string, port uint32, opts ...v1.ForwardOption) (io.ReadWriteCloser, error) {
	if s.ForwardFn == nil {
		return nil, ErrUnimplemented
	}
	return s.ForwardFn(ctx, workspace, sandboxName, port, opts...)
}

func (s *TCPStub) Listen(ctx context.Context, workspace, sandboxName string, remotePort uint32, localPort uint32, opts ...v1.ListenOption) (v1.ForwardListener, error) {
	if s.ListenFn == nil {
		return nil, ErrUnimplemented
	}
	return s.ListenFn(ctx, workspace, sandboxName, remotePort, localPort, opts...)
}
