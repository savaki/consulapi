package connect

import (
	"time"
)

type resolverOptions struct {
	logf func(format string, args ...interface{})
}

type ResolverOption func(*resolverOptions)

func WithLogger(logf func(format string, args ...interface{})) ResolverOption {
	return func(o *resolverOptions) {
		o.logf = logf
	}
}

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
