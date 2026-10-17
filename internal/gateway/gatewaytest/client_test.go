package gatewaytest_test

import (
	"context"
	"errors"
	"testing"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/gateway/gatewaytest"
	"google.golang.org/grpc/codes"
)

func TestNew_DelegatesToFake(t *testing.T) {
	c := gatewaytest.New()
	c.AddWorkspace(&v1.Workspace{Name: "ws1"})
	ws, err := c.Workspaces().Get(context.Background(), "ws1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if ws.Name != "ws1" {
		t.Fatalf("name = %q", ws.Name)
	}
}

func TestWithServices_Overrides(t *testing.T) {
	want := &v1.ServiceEndpoint{}
	called := false
	stub := &gatewaytest.ServiceStub{GetFn: func(ctx context.Context, ws, sb, name string) (*v1.ServiceEndpoint, error) {
		called = true
		if ws != "w" || sb != "s" || name != "n" {
			t.Errorf("args = %q %q %q", ws, sb, name)
		}
		return want, nil
	}}
	c := gatewaytest.New(gatewaytest.WithServices(stub))
	got, err := c.Services().Get(context.Background(), "w", "s", "n")
	if err != nil || got != want || !called {
		t.Fatalf("got %v, %v, called=%v", got, err, called)
	}
	_, err = c.Services().List("w", "s")
	if !errors.Is(err, gatewaytest.ErrUnimplemented) {
		t.Fatalf("List err = %v", err)
	}
	var se *v1.StatusError
	if !errors.As(err, &se) || se.Code != v1.ErrorUnimplemented {
		t.Fatalf("not an unimplemented StatusError: %v", err)
	}
}

func TestProviders_WrapperInjectsProfiles(t *testing.T) {
	profiles := &gatewaytest.ProfileStub{}
	refresh := &gatewaytest.RefreshStub{}
	c := gatewaytest.New(gatewaytest.WithProfiles(profiles), gatewaytest.WithRefresh(refresh))
	c.AddProvider("ws1", &v1.Provider{Name: "p1"})
	if c.Providers().Profiles() != v1.ProfileInterface(profiles) {
		t.Fatal("Profiles() is not the injected stub")
	}
	if c.Providers().Refresh() != v1.RefreshInterface(refresh) {
		t.Fatal("Refresh() is not the injected stub")
	}
	p, err := c.Providers().Get(context.Background(), "ws1", "p1")
	if err != nil || p.Name != "p1" {
		t.Fatalf("Get = %v, %v", p, err)
	}
}

func TestStaticPager(t *testing.T) {
	p := gatewaytest.StaticPager([]string{"a", "b"}, "tok")
	page, err := p.NextPage(context.Background())
	if err != nil || page == nil || len(page.Items) != 2 || page.NextPageToken != "tok" {
		t.Fatalf("page = %+v, %v", page, err)
	}
	// A continuation token is followed like a real pager: the follow-up page is empty and terminal.
	page, err = p.NextPage(context.Background())
	if err != nil || page == nil || len(page.Items) != 0 || page.NextPageToken != "" {
		t.Fatalf("second = %+v, %v", page, err)
	}
	if page, err = p.NextPage(context.Background()); err != nil || page != nil {
		t.Fatalf("third = %+v, %v", page, err)
	}
	last := gatewaytest.StaticPager([]string{"z"}, "")
	if page, _ = last.NextPage(context.Background()); page == nil || len(page.Items) != 1 {
		t.Fatalf("last first = %+v", page)
	}
	if page, err = last.NextPage(context.Background()); err != nil || page != nil {
		t.Fatalf("last second = %+v, %v", page, err)
	}
	e := gatewaytest.StatusErr(codes.FailedPrecondition, "x")
	if e.Code != v1.ErrorConflict || e.GRPCCode != int32(codes.FailedPrecondition) {
		t.Fatalf("StatusErr = %+v", e)
	}
}

func TestWithSandboxOverrides(t *testing.T) {
	stub := &gatewaytest.SandboxStub{GetFn: func(ctx context.Context, ws, name string) (*v1.Sandbox, error) {
		return &v1.Sandbox{Name: "stubbed"}, nil
	}}
	c := gatewaytest.New(gatewaytest.WithSandboxOverrides(stub))
	c.AddSandbox("w", &v1.Sandbox{Name: "real"})
	got, err := c.Sandboxes().Get(context.Background(), "w", "real")
	if err != nil || got.Name != "stubbed" {
		t.Fatalf("Get = %+v, %v", got, err)
	}
	// Unset methods fall back to the fake.
	list, err := c.Sandboxes().ListAll(context.Background(), "w")
	if err != nil || len(list) != 1 || list[0].Name != "real" {
		t.Fatalf("ListAll = %+v, %v", list, err)
	}
}
