package consulapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

var errUpdateTTL = errors.New("update ttl failed")

// ServiceKind is the kind of service being registered.
type ServiceKind string

const (
	// ServiceKindTypical is a typical, classic Consul service. This is
	// represented by the absence of a value. This was chosen for ease of
	// backwards compatibility: existing services in the catalog would
	// default to the typical service.
	ServiceKindTypical ServiceKind = ""

	// ServiceKindConnectProxy is a proxy for the Connect feature. This
	// service proxies another service within Consul and speaks the connect
	// protocol.
	ServiceKindConnectProxy ServiceKind = "connect-proxy"
)

type Status string

const (
	StatusPass Status = "passing"
	StatusWarn Status = "warning"
	StatusFail Status = "critical"
)

type AgentServiceCheck struct {
	CheckID                        string `json:",omitempty"`
	TTL                            string `json:",omitempty"`
	DeregisterCriticalServiceAfter string `json:",omitempty"`
}

// AgentServiceConnectProxyConfig is the proxy configuration in a connect-proxy
// ServiceDefinition or response.
type AgentServiceProxy struct {
	DestinationServiceName string
	DestinationServiceID   string `json:",omitempty"`
	LocalServiceAddress    string `json:",omitempty"`
	LocalServicePort       int    `json:",omitempty"`
}

type AgentServiceConnect struct {
	Native bool
}

type AgentServiceRegistration struct {
	Kind             ServiceKind          `json:",omitempty"`
	ID               string               `json:",omitempty"`
	Name             string               `json:",omitempty"`
	Tags             []string             `json:",omitempty"`
	Port             int                  `json:",omitempty"`
	Address          string               `json:",omitempty"`
	Check            *AgentServiceCheck   `json:",omitempty"`
	Connect          *AgentServiceConnect `json:",omitempty"`
	ProxyDestination string               `json:",omitempty"`
	Proxy            *AgentServiceProxy   `json:",omitempty"`
}

type Agent struct {
	client *client
}

func (a *Agent) ServiceRegister(ctx context.Context, registration AgentServiceRegistration) error {
	const path = "/v1/agent/service/register"

	if registration.Address == "" {
		registration.Address = a.client.hostAddr
	}
	if registration.Proxy != nil && registration.Proxy.LocalServiceAddress == "" {
		registration.Proxy.LocalServiceAddress = a.client.hostAddr
	}

	resp, err := a.client.Request(ctx, http.MethodPut, path, registration)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("register status code ->", resp.StatusCode)
	io.Copy(os.Stdout, resp.Body)

	return nil
}

func (a *Agent) ServiceDeregister(ctx context.Context, serviceID string) error {
	path := "/v1/agent/service/deregister/" + serviceID
	resp, err := a.client.Request(ctx, http.MethodPut, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (a *Agent) UpdateTTL(ctx context.Context, status Status, checkID, output string) error {
	input := struct {
		Status string
		Output string
	}{
		Status: string(status),
		Output: output,
	}

	path := "/v1/agent/check/update/" + checkID
	resp, err := a.client.Request(ctx, http.MethodPut, path, input)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errUpdateTTL
	}

	return nil
}

func NewAgent(opts ...Option) *Agent {
	client := newClient(opts...)
	return &Agent{
		client: client,
	}
}
