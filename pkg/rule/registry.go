package rule

import (
	"fmt"
	"sync"
)

var (
	globalRegistry = &Registry{
		rules: make(map[string]Rule),
	}
)

type Registry struct {
	mu    sync.RWMutex
	rules map[string]Rule
}

func GlobalRegistry() *Registry {
	return globalRegistry
}

func Register(r Rule) {
	globalRegistry.Register(r)
}

func (reg *Registry) Register(r Rule) {
	reg.mu.Lock()
	defer reg.mu.Unlock()
	name := r.Name()
	if _, exists := reg.rules[name]; exists {
		panic(fmt.Sprintf("glint: duplicate rule registration: %s", name))
	}
	reg.rules[name] = r
}

func (reg *Registry) Get(name string) (Rule, bool) {
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	r, ok := reg.rules[name]
	return r, ok
}

func (reg *Registry) All() []Rule {
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	out := make([]Rule, 0, len(reg.rules))
	for _, r := range reg.rules {
		out = append(out, r)
	}
	return out
}

func (reg *Registry) Names() []string {
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	out := make([]string, 0, len(reg.rules))
	for name := range reg.rules {
		out = append(out, name)
	}
	return out
}
