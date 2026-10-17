package gatewaytest

import (
	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/NVIDIA/OpenShell/sdk/go/openshell/v1/fake"
)

// Client embeds the SDK's in-memory fake and lets tests replace sub-interfaces.
type Client struct {
	*fake.Client

	exec      v1.ExecInterface
	services  v1.ServiceInterface
	config    v1.ConfigInterface
	policy    v1.PolicyInterface
	profiles  v1.ProfileInterface
	refresh   v1.RefreshInterface
	ssh       v1.SSHInterface
	tcp       v1.TCPInterface
	fakeOpts  []fake.ClientOption
	providers *providers
}

var _ v1.ClientInterface = (*Client)(nil)

// Option configures a Client.
type Option func(*Client)

// New builds a Client wrapping fake.NewClient.
func New(opts ...Option) *Client {
	c := &Client{}
	for _, o := range opts {
		o(c)
	}
	c.Client = fake.NewClient(c.fakeOpts...)
	c.providers = &providers{ProviderInterface: c.Client.Providers(), c: c}
	return c
}

func WithFakeOptions(opts ...fake.ClientOption) Option {
	return func(c *Client) { c.fakeOpts = append(c.fakeOpts, opts...) }
}
func WithExec(v v1.ExecInterface) Option        { return func(c *Client) { c.exec = v } }
func WithServices(v v1.ServiceInterface) Option { return func(c *Client) { c.services = v } }
func WithConfig(v v1.ConfigInterface) Option    { return func(c *Client) { c.config = v } }
func WithPolicy(v v1.PolicyInterface) Option    { return func(c *Client) { c.policy = v } }
func WithProfiles(v v1.ProfileInterface) Option { return func(c *Client) { c.profiles = v } }
func WithRefresh(v v1.RefreshInterface) Option  { return func(c *Client) { c.refresh = v } }
func WithSSH(v v1.SSHInterface) Option          { return func(c *Client) { c.ssh = v } }
func WithTCP(v v1.TCPInterface) Option          { return func(c *Client) { c.tcp = v } }

func (c *Client) Exec() v1.ExecInterface {
	if c.exec != nil {
		return c.exec
	}
	return c.Client.Exec()
}

func (c *Client) Services() v1.ServiceInterface {
	if c.services != nil {
		return c.services
	}
	return c.Client.Services()
}

func (c *Client) Config() v1.ConfigInterface {
	if c.config != nil {
		return c.config
	}
	return c.Client.Config()
}

func (c *Client) Policy() v1.PolicyInterface {
	if c.policy != nil {
		return c.policy
	}
	return c.Client.Policy()
}

func (c *Client) SSH() v1.SSHInterface {
	if c.ssh != nil {
		return c.ssh
	}
	return c.Client.SSH()
}

func (c *Client) TCP() v1.TCPInterface {
	if c.tcp != nil {
		return c.tcp
	}
	return c.Client.TCP()
}

// Providers returns a wrapper delegating to the fake except Profiles/Refresh.
func (c *Client) Providers() v1.ProviderInterface { return c.providers }

type providers struct {
	v1.ProviderInterface
	c *Client
}

func (p *providers) Profiles() v1.ProfileInterface {
	if p.c.profiles != nil {
		return p.c.profiles
	}
	return p.ProviderInterface.Profiles()
}

func (p *providers) Refresh() v1.RefreshInterface {
	if p.c.refresh != nil {
		return p.c.refresh
	}
	return p.ProviderInterface.Refresh()
}
