package consulapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type HealthService struct {
	ID      string
	Service string
	Address string
	Port    int
}

type HealthServiceEntry struct {
	Service HealthService
}

type Health struct {
	client *client
}

func (h *Health) Connect(ctx context.Context, service string, passing bool) ([]HealthServiceEntry, error) {
	path := fmt.Sprintf("/v1/health/connect/%v?passing=%v", service, passing)
	resp, err := h.client.Request(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var entries []HealthServiceEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}

	return entries, nil
}

func NewHealth(opts ...Option) *Health {
	client := newClient(opts...)
	return &Health{
		client: client,
	}
}
