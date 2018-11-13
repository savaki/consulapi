package connect

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/savaki/consulapi"
	"google.golang.org/grpc/naming"
)

type Mock struct {
	entries [][]consulapi.HealthServiceEntry
}

func (m *Mock) Connect(ctx context.Context, service string, passing bool) ([]consulapi.HealthServiceEntry, error) {
	if len(m.entries) == 0 {
		return nil, nil
	}

	head := m.entries[0]
	m.entries = m.entries[1:]
	return head, nil
}

func TestResolver(t *testing.T) {
	a := consulapi.HealthServiceEntry{
		Service: consulapi.HealthService{
			ID:      "a1",
			Service: "a2",
			Address: "a3",
			Port:    1,
		},
	}
	entries := [][]consulapi.HealthServiceEntry{
		{a},
		{},
	}
	m := &Mock{entries: entries}
	s := "blah"
	r := NewResolver(m, s, WithLogger(log.Printf))
	watcher, err := r.Resolve("")
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}

	// test 1 - add
	//
	got, err := watcher.Next()
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}

	want := []*naming.Update{
		{
			Op:   naming.Add,
			Addr: fmt.Sprintf("%v:%v", a.Service.Address, a.Service.Port),
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v; want %v", watcher, entries)
	}

	// test 2 - delete
	//
	got, err = watcher.Next()
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}

	want = []*naming.Update{
		{
			Op:   naming.Delete,
			Addr: fmt.Sprintf("%v:%v", a.Service.Address, a.Service.Port),
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v; want %v", watcher, entries)
	}

	// test 3 - no change
	//
	got, err = watcher.Next()
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %v; want 0", len(got))
	}
}
