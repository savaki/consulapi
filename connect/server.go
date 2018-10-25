package connect

import (
	"context"
	"encoding/hex"
	"io"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/savaki/consulapi"
)

const defaultHealthCheckInterval = 3 * time.Second

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func guid() string {
	data := make([]byte, 16)
	_, err := r.Read(data)
	if err != nil {
		log.Fatalln(err)
	}
	return hex.EncodeToString(data)
}

type closerFunc func() error

func (fn closerFunc) Close() error {
	return fn()
}

type config struct {
	service             string
	port                int
	healthCheckInterval time.Duration
	healthCheckFunc     func() error
	agent               *consulapi.Agent
}

func registerAndUpdate(ctx context.Context, config config) error {
	var (
		serviceID = guid()
		checkID   = guid()
	)

	registration := consulapi.AgentServiceRegistration{
		Kind: consulapi.ServiceKindTypical,
		ID:   serviceID,
		Name: config.service,
		Port: config.port,
		Check: &consulapi.AgentServiceCheck{
			CheckID:                        checkID,
			TTL:                            makeTTL(config.healthCheckInterval * 3),
			DeregisterCriticalServiceAfter: makeTTL(config.healthCheckInterval * 5),
		},
		Connect: &consulapi.AgentServiceConnect{
			Native: true,
		},
	}
	if err := config.agent.ServiceRegister(ctx, registration); err != nil {
		return err
	}
	defer config.agent.ServiceDeregister(context.Background(), serviceID)

	ticker := time.NewTicker(config.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			status := consulapi.StatusPass
			output := "ok"
			if err := config.healthCheckFunc(); err != nil {
				status = consulapi.StatusFail
				output = err.Error()
			}

			if err := config.agent.UpdateTTL(ctx, status, checkID, output); err != nil {
				return err
			}
		}
	}
}

func registerLoop(ctx context.Context, config config) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := registerAndUpdate(ctx, config); err != nil {
			log.Println(err)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(12 * time.Second):
			log.Println("retrying ...")
		}
	}
}

func Register(service string, port int, opts ...RegisterOption) (io.Closer, error) {
	options := registerOptions{
		healthCheckFunc:     func() error { return nil },
		healthCheckInterval: defaultHealthCheckInterval,
	}
	for _, opt := range opts {
		opt(&options)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		defer close(done)

		config := config{
			service:             service,
			port:                port,
			healthCheckInterval: options.healthCheckInterval,
			healthCheckFunc:     options.healthCheckFunc,
			agent:               consulapi.NewAgent(options.agentOptions...),
		}
		registerLoop(ctx, config)
	}()

	return closerFunc(func() error {
		cancel()
		<-done
		return nil
	}), nil
}

func makeTTL(d time.Duration) string {
	return strconv.FormatInt(int64(d/time.Second), 10) + "s"
}
