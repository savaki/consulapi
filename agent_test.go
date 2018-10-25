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
			CheckID:                        checkID,
			TTL:                            "15s",
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
