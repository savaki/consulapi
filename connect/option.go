package connect

import (
	"time"
)

type resolverOptions struct {
}

type ResolverOption func(*resolverOptions)

type serviceOptions struct {
	healthCheckFunc     func() error
	healthCheckInterval time.Duration
}

type ServiceOption func(*serviceOptions)

func WithHealthCheckFunc(fn func() error) ServiceOption {
	return func(o *serviceOptions) {
		o.healthCheckFunc = fn
	}
}

func WithHealthCheckInterval(interval time.Duration) ServiceOption {
	return func(o *serviceOptions) {
		o.healthCheckInterval = interval
	}
}
