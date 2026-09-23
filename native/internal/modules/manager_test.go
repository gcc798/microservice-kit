package modules

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
)

type managerTestModule struct {
	name       string
	deps       Dependencies
	initErr    error
	startErr   error
	stopErr    error
	refreshErr error
	events     *[]string
	mu         *sync.Mutex
}

func (m *managerTestModule) Name() string { return m.name }

func (m *managerTestModule) Init(_ context.Context, deps Dependencies) error {
	m.deps = deps
	m.record("init:" + m.name)
	return m.initErr
}

func (m *managerTestModule) Start(context.Context) error {
	m.record("start:" + m.name)
	return m.startErr
}

func (m *managerTestModule) Stop(context.Context) error {
	m.record("stop:" + m.name)
	return m.stopErr
}

func (m *managerTestModule) Refresh(context.Context, ModuleRefreshRequest) error {
	m.record("refresh:" + m.name)
	return m.refreshErr
}

func (m *managerTestModule) record(event string) {
	if m.events == nil || m.mu == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	*m.events = append(*m.events, event)
}

func TestManagerUsesExplicitDependenciesAndStopsInReverseOrder(t *testing.T) {
	var mu sync.Mutex
	events := make([]string, 0)
	deps := Dependencies{}
	first := &managerTestModule{name: "first", events: &events, mu: &mu}
	second := &managerTestModule{name: "second", events: &events, mu: &mu}
	manager := NewManager(deps)

	if err := manager.Register(t.Context(), first, second); err != nil {
		t.Fatal(err)
	}
	if first.deps != deps || second.deps != deps {
		t.Fatal("manager did not pass explicit dependencies")
	}
	if err := manager.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := manager.Stop(t.Context()); err != nil {
		t.Fatal(err)
	}
	want := []string{"init:first", "init:second", "start:first", "start:second", "stop:second", "stop:first"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestManagerRollsBackInitializationFailure(t *testing.T) {
	var mu sync.Mutex
	events := make([]string, 0)
	first := &managerTestModule{name: "first", events: &events, mu: &mu}
	second := &managerTestModule{name: "second", initErr: errors.New("init failed"), events: &events, mu: &mu}
	manager := NewManager(Dependencies{})

	if err := manager.Register(t.Context(), first, second); err == nil {
		t.Fatal("Register() succeeded")
	}
	want := []string{"init:first", "init:second", "stop:first"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if err := manager.Start(t.Context()); err != nil {
		t.Fatalf("Start() after rollback = %v", err)
	}
}

func TestManagerRefreshSerializesAndIgnoresUnknownModules(t *testing.T) {
	module := &managerTestModule{name: "known"}
	manager := NewManager(Dependencies{})
	if err := manager.Register(t.Context(), module); err != nil {
		t.Fatal(err)
	}
	if err := manager.Refresh(t.Context(), "unknown", ModuleRefreshRequest{}); err != nil {
		t.Fatal(err)
	}
	if err := manager.Refresh(t.Context(), "known", ModuleRefreshRequest{}); err != nil {
		t.Fatal(err)
	}
}
