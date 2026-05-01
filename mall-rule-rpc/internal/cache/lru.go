package cache

import (
	"sync"

	"github.com/expr-lang/expr/vm"
	lru "github.com/hashicorp/golang-lru/v2"
)

// ProgramCache is a thread-safe LRU keyed by `ruleId:version`.
type ProgramCache struct {
	mu sync.Mutex
	c  *lru.Cache[string, *vm.Program]
}

func NewProgramCache(size int) (*ProgramCache, error) {
	if size <= 0 {
		size = 4096
	}
	c, err := lru.New[string, *vm.Program](size)
	if err != nil {
		return nil, err
	}
	return &ProgramCache{c: c}, nil
}

func (p *ProgramCache) Get(key string) (*vm.Program, bool) {
	return p.c.Get(key)
}

func (p *ProgramCache) Add(key string, prog *vm.Program) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.c.Add(key, prog)
}

func (p *ProgramCache) Purge() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.c.Purge()
}
