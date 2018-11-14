package connect

import (
	"context"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/savaki/consulapi"
	"google.golang.org/grpc/naming"
)

type HealthAPI interface {
	Connect(ctx context.Context, service string, passing bool) ([]consulapi.HealthServiceEntry, error)
}

type watcher struct {
	ctx     context.Context
	cancel  context.CancelFunc
	client  HealthAPI
	service string
	logf    func(format string, args ...interface{})
	debugf  func(format string, args ...interface{})

	mutex    sync.Mutex
	previous []consulapi.HealthServiceEntry
}

func (w *watcher) poll() ([]*naming.Update, error) {
	ctx, cancel := context.WithTimeout(w.ctx, 30*time.Second)
	defer cancel()

	services, err := w.client.Connect(ctx, w.service, true)
	if err != nil {
		return nil, err
	}
	sort.Slice(services, func(i, j int) bool { return services[i].Service.ID < services[j].Service.ID })

	w.mutex.Lock()
	updates := w.makeUpdates(w.previous, services)
	w.previous = services
	w.mutex.Unlock()

	return updates, nil
}

func (w *watcher) Next() ([]*naming.Update, error) {
	for {
		updates, err := w.poll()
		if err == nil {
			w.debugf("found updates for service, %v - %#v", w.service, updates)
			return updates, nil
		}

		select {
		case <-w.ctx.Done():
			return nil, w.ctx.Err()
		case <-time.After(15 * time.Second):
		}
	}
}

func (w *watcher) makeUpdates(previous, latest []consulapi.HealthServiceEntry) []*naming.Update {
	var updates []*naming.Update

	// adds
addLoop:
	for _, l := range latest {
		for _, p := range previous {
			if l.Service.ID == p.Service.ID {
				continue addLoop
			}
		}

		w.logf("consul resolver: adding endpoint, %v:%v, to service, %v", l.Service.Address, l.Service.Port, w.service)
		updates = append(updates, &naming.Update{
			Op:   naming.Add,
			Addr: hostAndPort(l.Service.Address, l.Service.Port),
		})
	}

	// deletes
deleteLoop:
	for _, p := range previous {
		for _, l := range latest {
			if p.Service.ID == l.Service.ID {
				continue deleteLoop
			}
		}

		w.logf("consul resolver: removing endpoint, %v:%v, from service, %v", p.Service.Address, p.Service.Port, w.service)
		updates = append(updates, &naming.Update{
			Op:   naming.Delete,
			Addr: hostAndPort(p.Service.Address, p.Service.Port),
		})
	}

	return updates
}

func (w *watcher) Close() {
	w.cancel()
}

type resolver struct {
	client  HealthAPI
	service string
	logf    func(format string, args ...interface{})
	debugf  func(format string, args ...interface{})
}

func (r *resolver) Resolve(_ string) (naming.Watcher, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &watcher{
		ctx:     ctx,
		cancel:  cancel,
		client:  r.client,
		service: r.service,
		logf:    r.logf,
		debugf:  r.debugf,
	}, nil
}

func NewResolver(client HealthAPI, service string, opts ...ResolverOption) naming.Resolver {
	options := resolverOptions{
		logf:   func(format string, args ...interface{}) {},
		debugf: func(format string, args ...interface{}) {},
	}

	for _, opt := range opts {
		opt(&options)
	}

	return &resolver{
		client:  client,
		service: service,
		logf:    options.logf,
		debugf:  options.debugf,
	}
}

func hostAndPort(host string, port int) string {
	return host + ":" + strconv.Itoa(port)
}
