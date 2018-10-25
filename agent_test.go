package consulapi

import (
	"context"
	"testing"
)

func TestAgent(t *testing.T) {
	var (
		ctx       = context.Background()
		agent     = NewAgent()
		serviceID = guid()
		checkID   = guid()
		name      = guid()
		port      = 8090
	)

	registration := AgentServiceRegistration{
		Kind: ServiceKindTypical,
		ID:   serviceID,
		Name: name,
		Port: port,
		Check: &AgentServiceCheck{
			CheckID: checkID,
			TTL:     "15s",
			DeregisterCriticalServiceAfter: "15s",
		},
		Connect: &AgentServiceConnect{
			Native: true,
		},
		Proxy: &AgentServiceProxy{
			DestinationServiceName: name,
			DestinationServiceID:   serviceID,
			LocalServicePort:       port,
		},
	}
	err := agent.ServiceRegister(ctx, registration)
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}

	err = agent.UpdateTTL(ctx, StatusPass, registration.Check.CheckID, "ok")
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}

	err = agent.ServiceDeregister(ctx, registration.ID)
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}
}

func TestAgent_ConnectCARoots(t *testing.T) {
	ctx := context.Background()
	agent := NewAgent()
	out, err := agent.ConnectCARoots(ctx)
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}
	if got, want := len(out.Roots), 1; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}

	root := out.Roots[0]
	if got := root.ID; got == "" {
		t.Fatalf("got blank string; want not blank string")
	}
	if got := root.Name; got == "" {
		t.Fatalf("got blank string; want not blank string")
	}
	if got := root.SerialNumber; got == 0 {
		t.Fatalf("got 0; want not 0")
	}
	if got := root.SigningKeyID; got == "" {
		t.Fatalf("got blank string; want not blank string")
	}
	if got := root.NotBefore; got.IsZero() {
		t.Fatalf("got blank time.Time; want not blank time.Time")
	}
	if got := root.NotAfter; got.IsZero() {
		t.Fatalf("got blank time.Time; want not blank time.Time")
	}
	if got := root.RootCert; got == "" {
		t.Fatalf("got blank string; want not blank string")
	}
	if got := len(root.IntermediateCerts); got != 0 {
		t.Fatalf("got %v; want 0", got)
	}
	if got := root.Active; !got {
		t.Fatalf("got false; want true")
	}
	if got := root.CreateIndex; got == 0 {
		t.Fatalf("got 0; want not zero")
	}
	if got := root.ModifyIndex; got == 0 {
		t.Fatalf("got 0; want not zero")
	}
}

func TestAgent_ConnectCALeaf(t *testing.T) {
	ctx := context.Background()
	agent := NewAgent()
	out, err := agent.ConnectCALeaf(ctx, "my-service")
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}
	if got := out.SerialNumber; got == "" {
		t.Fatalf("got blank string; want not blank string")
	}
	if got := out.CertPEM; got == "" {
		t.Fatalf("got blank string; want not blank string")
	}
	if got := out.PrivateKeyPEM; got == "" {
		t.Fatalf("got blank string; want not blank string")
	}
	if got := out.Service; got == "" {
		t.Fatalf("got blank string; want not blank string")
	}
	if got := out.ServiceURI; got == "" {
		t.Fatalf("got blank string; want not blank string")
	}
	if got := out.ValidAfter; got.IsZero() {
		t.Fatalf("got blank time.Time; want not blank time.Time")
	}
	if got := out.ValidBefore; got.IsZero() {
		t.Fatalf("got blank time.Time; want not blank time.Time")
	}
	if got := out.CreateIndex; got == 0 {
		t.Fatalf("got 0; want not zero")
	}
	if got := out.ModifyIndex; got == 0 {
		t.Fatalf("got 0; want not zero")
	}
}

func TestAgent_ConnectAuthorize(t *testing.T) {
	ctx := context.Background()
	agent := NewAgent()

	leaf, err := agent.ConnectCALeaf(ctx, "client")
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}

	out, err := agent.ConnectAuthorize(ctx, AgentConnectAuthorizeRequest{
		Target:           "db",
		ClientCertURI:    leaf.ServiceURI,
		ClientCertSerial: "blah",
	})
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}
	if got := out.Authorized; !got {
		t.Fatalf("got false; want true")
	}
	if got := out.Reason; got == "" {
		t.Fatalf("got blank string; want not blank string")
	}
}
