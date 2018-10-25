package connect

import (
	"time"

	"github.com/savaki/consulapi"
)

type registerOptions struct {
	agentOptions        []consulapi.Option
	healthCheckFunc     func() error
	healthCheckInterval time.Duration
}

type RegisterOption func(*registerOptions)

func WithConsulAddr(addr string) RegisterOption {
	return func(o *registerOptions) {
		o.agentOptions = append(o.agentOptions, consulapi.WithConsulAddr(addr))
	}
}

func WithHealthCheckFunc(fn func() error) RegisterOption {
	return func(o *registerOptions) {
		o.healthCheckFunc = fn
	}
}

func WithHealthCheckInterval(interval time.Duration) RegisterOption {
	return func(o *registerOptions) {
		o.healthCheckInterval = interval
	}
}
