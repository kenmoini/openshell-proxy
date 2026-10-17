package gatewaytest

import (
	"context"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
)

// SandboxStub overrides individual v1.SandboxInterface methods. Each method
// calls its Fn field when set and otherwise falls back to the embedded
// interface (the fake's sandboxes unless set explicitly).
type SandboxStub struct {
	v1.SandboxInterface

	CreateFn           func(ctx context.Context, workspace, name string, spec *v1.SandboxSpec, labels map[string]string, opts ...v1.CreateOptions) (*v1.Sandbox, error)
	GetFn              func(ctx context.Context, workspace, name string) (*v1.Sandbox, error)
	ListFn             func(workspace string, opts ...v1.ListOptions) (*v1.Pager[*v1.Sandbox], error)
	ListAllFn          func(ctx context.Context, workspace string, opts ...v1.ListOptions) ([]*v1.Sandbox, error)
	StopFn             func(ctx context.Context, workspace, name string) (*v1.Sandbox, error)
	StartFn            func(ctx context.Context, workspace, name string) (*v1.Sandbox, error)
	DeleteFn           func(ctx context.Context, workspace, name string, opts ...v1.DeleteOptions) (*v1.DeletionResult, error)
	AttachProviderFn   func(ctx context.Context, workspace, sandboxName, providerName string, expectedResourceVersion uint64) (*v1.AttachProviderResult, error)
	DetachProviderFn   func(ctx context.Context, workspace, sandboxName, providerName string, expectedResourceVersion uint64) (*v1.DetachProviderResult, error)
	ListProvidersFn    func(workspace, sandboxName string, opts ...v1.ListOptions) (*v1.Pager[*v1.Provider], error)
	ListAllProvidersFn func(ctx context.Context, workspace, sandboxName string, opts ...v1.ListOptions) ([]*v1.Provider, error)
	WaitReadyFn        func(ctx context.Context, workspace, name string, opts ...v1.WaitOptions) (*v1.Sandbox, error)
	WaitStoppedFn      func(ctx context.Context, workspace, name string, opts ...v1.WaitOptions) (*v1.Sandbox, error)
	WatchFn            func(ctx context.Context, workspace, name string, opts ...v1.WatchOptions) (v1.WatchInterface[*v1.Sandbox], error)
	GetLogsFn          func(ctx context.Context, workspace, sandboxName string, opts ...v1.LogOption) (*v1.LogResult, error)
}

var _ v1.SandboxInterface = (*SandboxStub)(nil)

func (s *SandboxStub) Create(ctx context.Context, workspace, name string, spec *v1.SandboxSpec, labels map[string]string, opts ...v1.CreateOptions) (*v1.Sandbox, error) {
	if s.CreateFn != nil {
		return s.CreateFn(ctx, workspace, name, spec, labels, opts...)
	}
	return s.SandboxInterface.Create(ctx, workspace, name, spec, labels, opts...)
}

func (s *SandboxStub) Get(ctx context.Context, workspace, name string) (*v1.Sandbox, error) {
	if s.GetFn != nil {
		return s.GetFn(ctx, workspace, name)
	}
	return s.SandboxInterface.Get(ctx, workspace, name)
}

func (s *SandboxStub) List(workspace string, opts ...v1.ListOptions) (*v1.Pager[*v1.Sandbox], error) {
	if s.ListFn != nil {
		return s.ListFn(workspace, opts...)
	}
	return s.SandboxInterface.List(workspace, opts...)
}

func (s *SandboxStub) ListAll(ctx context.Context, workspace string, opts ...v1.ListOptions) ([]*v1.Sandbox, error) {
	if s.ListAllFn != nil {
		return s.ListAllFn(ctx, workspace, opts...)
	}
	return s.SandboxInterface.ListAll(ctx, workspace, opts...)
}

func (s *SandboxStub) Stop(ctx context.Context, workspace, name string) (*v1.Sandbox, error) {
	if s.StopFn != nil {
		return s.StopFn(ctx, workspace, name)
	}
	return s.SandboxInterface.Stop(ctx, workspace, name)
}

func (s *SandboxStub) Start(ctx context.Context, workspace, name string) (*v1.Sandbox, error) {
	if s.StartFn != nil {
		return s.StartFn(ctx, workspace, name)
	}
	return s.SandboxInterface.Start(ctx, workspace, name)
}

func (s *SandboxStub) Delete(ctx context.Context, workspace, name string, opts ...v1.DeleteOptions) (*v1.DeletionResult, error) {
	if s.DeleteFn != nil {
		return s.DeleteFn(ctx, workspace, name, opts...)
	}
	return s.SandboxInterface.Delete(ctx, workspace, name, opts...)
}

func (s *SandboxStub) AttachProvider(ctx context.Context, workspace, sandboxName, providerName string, expectedResourceVersion uint64) (*v1.AttachProviderResult, error) {
	if s.AttachProviderFn != nil {
		return s.AttachProviderFn(ctx, workspace, sandboxName, providerName, expectedResourceVersion)
	}
	return s.SandboxInterface.AttachProvider(ctx, workspace, sandboxName, providerName, expectedResourceVersion)
}

func (s *SandboxStub) DetachProvider(ctx context.Context, workspace, sandboxName, providerName string, expectedResourceVersion uint64) (*v1.DetachProviderResult, error) {
	if s.DetachProviderFn != nil {
		return s.DetachProviderFn(ctx, workspace, sandboxName, providerName, expectedResourceVersion)
	}
	return s.SandboxInterface.DetachProvider(ctx, workspace, sandboxName, providerName, expectedResourceVersion)
}

func (s *SandboxStub) ListProviders(workspace, sandboxName string, opts ...v1.ListOptions) (*v1.Pager[*v1.Provider], error) {
	if s.ListProvidersFn != nil {
		return s.ListProvidersFn(workspace, sandboxName, opts...)
	}
	return s.SandboxInterface.ListProviders(workspace, sandboxName, opts...)
}

func (s *SandboxStub) ListAllProviders(ctx context.Context, workspace, sandboxName string, opts ...v1.ListOptions) ([]*v1.Provider, error) {
	if s.ListAllProvidersFn != nil {
		return s.ListAllProvidersFn(ctx, workspace, sandboxName, opts...)
	}
	return s.SandboxInterface.ListAllProviders(ctx, workspace, sandboxName, opts...)
}

func (s *SandboxStub) WaitReady(ctx context.Context, workspace, name string, opts ...v1.WaitOptions) (*v1.Sandbox, error) {
	if s.WaitReadyFn != nil {
		return s.WaitReadyFn(ctx, workspace, name, opts...)
	}
	return s.SandboxInterface.WaitReady(ctx, workspace, name, opts...)
}

func (s *SandboxStub) WaitStopped(ctx context.Context, workspace, name string, opts ...v1.WaitOptions) (*v1.Sandbox, error) {
	if s.WaitStoppedFn != nil {
		return s.WaitStoppedFn(ctx, workspace, name, opts...)
	}
	return s.SandboxInterface.WaitStopped(ctx, workspace, name, opts...)
}

func (s *SandboxStub) Watch(ctx context.Context, workspace, name string, opts ...v1.WatchOptions) (v1.WatchInterface[*v1.Sandbox], error) {
	if s.WatchFn != nil {
		return s.WatchFn(ctx, workspace, name, opts...)
	}
	return s.SandboxInterface.Watch(ctx, workspace, name, opts...)
}

func (s *SandboxStub) GetLogs(ctx context.Context, workspace, sandboxName string, opts ...v1.LogOption) (*v1.LogResult, error) {
	if s.GetLogsFn != nil {
		return s.GetLogsFn(ctx, workspace, sandboxName, opts...)
	}
	return s.SandboxInterface.GetLogs(ctx, workspace, sandboxName, opts...)
}
