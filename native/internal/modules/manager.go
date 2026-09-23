package modules

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// Manager 负责进程内模块的注册和生命周期管理。
type Manager struct {
	deps        Dependencies
	mu          sync.RWMutex
	lifecycleMu sync.Mutex
	modules     map[string]Module
	order       []string
	refreshMu   map[string]*sync.Mutex
	started     int
}

// NewManager 创建模块管理器。
func NewManager(deps Dependencies) *Manager {
	return &Manager{deps: deps, modules: make(map[string]Module), refreshMu: make(map[string]*sync.Mutex)}
}

// Register 注册模块并按注册顺序完成初始化。
func (m *Manager) Register(ctx context.Context, candidates ...Module) error {
	m.lifecycleMu.Lock()
	defer m.lifecycleMu.Unlock()
	m.mu.Lock()
	if m.started != 0 {
		m.mu.Unlock()
		return errors.New("cannot register modules after module startup")
	}
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		if candidate == nil {
			m.mu.Unlock()
			return errors.New("cannot register a nil module")
		}
		name := strings.TrimSpace(candidate.Name())
		if name == "" {
			m.mu.Unlock()
			return errors.New("cannot register a module with an empty name")
		}
		if _, exists := m.modules[name]; exists {
			m.mu.Unlock()
			return fmt.Errorf("module %q is already registered", name)
		}
		if _, exists := seen[name]; exists {
			m.mu.Unlock()
			return fmt.Errorf("module %q appears more than once", name)
		}
		seen[name] = struct{}{}
	}
	for _, candidate := range candidates {
		name := strings.TrimSpace(candidate.Name())
		m.modules[name] = candidate
		m.order = append(m.order, name)
		m.refreshMu[name] = &sync.Mutex{}
	}
	m.mu.Unlock()

	initialized := make([]Module, 0, len(candidates))
	for _, candidate := range candidates {
		if err := candidate.Init(ctx, m.deps); err != nil {
			for i := len(initialized) - 1; i >= 0; i-- {
				_ = initialized[i].Stop(ctx)
			}
			m.mu.Lock()
			for _, rollback := range candidates {
				name := strings.TrimSpace(rollback.Name())
				delete(m.modules, name)
				delete(m.refreshMu, name)
			}
			m.order = m.order[:len(m.order)-len(candidates)]
			m.mu.Unlock()
			return fmt.Errorf("initialize module %q: %w", candidate.Name(), err)
		}
		initialized = append(initialized, candidate)
	}
	return nil
}

// Start 按注册顺序启动所有模块；启动失败时回滚已启动模块。
func (m *Manager) Start(ctx context.Context) error {
	m.lifecycleMu.Lock()
	defer m.lifecycleMu.Unlock()
	m.mu.RLock()
	if m.started != 0 {
		m.mu.RUnlock()
		return errors.New("modules are already started")
	}
	ordered := make([]Module, 0, len(m.order))
	for _, name := range m.order {
		ordered = append(ordered, m.modules[name])
	}
	m.mu.RUnlock()

	started := make([]Module, 0, len(ordered))
	for _, candidate := range ordered {
		if err := candidate.Start(ctx); err != nil {
			var rollback []error
			for i := len(started) - 1; i >= 0; i-- {
				if stopErr := started[i].Stop(ctx); stopErr != nil {
					rollback = append(rollback, fmt.Errorf("rollback module %q: %w", started[i].Name(), stopErr))
				}
			}
			return errors.Join(append([]error{fmt.Errorf("start module %q: %w", candidate.Name(), err)}, rollback...)...)
		}
		started = append(started, candidate)
	}
	m.mu.Lock()
	m.started = len(started)
	m.mu.Unlock()
	return nil
}

// Stop 按启动顺序逆序停止模块。
func (m *Manager) Stop(ctx context.Context) error {
	m.lifecycleMu.Lock()
	defer m.lifecycleMu.Unlock()
	m.mu.RLock()
	ordered := make([]Module, 0, m.started)
	for i := 0; i < m.started; i++ {
		ordered = append(ordered, m.modules[m.order[i]])
	}
	m.mu.RUnlock()

	var stopErrors []error
	for i := len(ordered) - 1; i >= 0; i-- {
		if err := ordered[i].Stop(ctx); err != nil {
			stopErrors = append(stopErrors, fmt.Errorf("stop module %q: %w", ordered[i].Name(), err))
		}
	}
	m.mu.Lock()
	m.started = 0
	m.mu.Unlock()
	return errors.Join(stopErrors...)
}

// Refresh 串行执行指定模块的本地刷新请求。
func (m *Manager) Refresh(ctx context.Context, name string, req ModuleRefreshRequest) error {
	m.mu.RLock()
	module, lock := m.modules[name], m.refreshMu[name]
	m.mu.RUnlock()
	if module == nil || lock == nil {
		return nil
	}
	lock.Lock()
	defer lock.Unlock()
	if err := module.Refresh(ctx, req); err != nil {
		return fmt.Errorf("refresh module %q: %w", name, err)
	}
	return nil
}
