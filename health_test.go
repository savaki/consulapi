package consulapi

import (
	"context"
	"encoding/hex"
	"log"
	"math/rand"
	"testing"
	"time"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func guid() string {
	data := make([]byte, 16)
	_, err := r.Read(data)
	if err != nil {
		log.Fatalln(err)
	}
	return hex.EncodeToString(data)
}

func TestHealth(t *testing.T) {
	var (
		ctx       = context.Background()
		health    = NewHealth()
		agent     = NewAgent()
		serviceID = guid()
		checkID   = guid()
		name      = guid()
		port      = 8080
	)

	registration := AgentServiceRegistration{
		Kind: ServiceKindTypical,
		ID:   serviceID,
		Name: name,
		Port: port,
		Check: &AgentServiceCheck{
			CheckID:                        checkID,
			TTL:                            "5s",
			DeregisterCriticalServiceAfter: "5s",
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

	err = agent.UpdateTTL(ctx, StatusPass, checkID, "ok")
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}

	entries, err := health.Connect(ctx, name, true)
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}
	if got, want := len(entries), 1; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}

	service := entries[0].Service
	if got := service.ID; got == "" {
		t.Fatalf("got %v; want not blank", got)
	}
	if got := service.Service; got == "" {
		t.Fatalf("got %v; want not blank", got)
	}
	if got := service.Address; got == "" {
		t.Fatalf("got %v; want not blank", got)
	}
	if got := service.Port; got == 0 {
		t.Fatalf("got %v; want != 0", got)
	}
}
