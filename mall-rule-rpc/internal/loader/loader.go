package loader

import (
	"context"
	"fmt"

	rcache "mall-rule-rpc/internal/cache"
	"mall-rule-rpc/internal/engine"
	"mall-rule-rpc/internal/model"

	"github.com/expr-lang/expr/vm"
	"golang.org/x/sync/singleflight"
)

type Loader struct {
	rules model.RuleModel
	cache *rcache.ProgramCache
	sf    singleflight.Group
}

func New(rules model.RuleModel, c *rcache.ProgramCache) *Loader {
	return &Loader{rules: rules, cache: c}
}

// LoadById fetches the rule by id, compiles its expression, and caches it.
// Returns the compiled program plus the persisted Rule row.
func (l *Loader) LoadById(ctx context.Context, id int64) (*vm.Program, *model.Rule, error) {
	r, err := l.rules.FindOne(ctx, uint64(id))
	if err != nil {
		return nil, nil, fmt.Errorf("rule %d: %w", id, err)
	}
	prog, err := l.compileWithCache(r)
	if err != nil {
		return nil, r, err
	}
	return prog, r, nil
}

func (l *Loader) compileWithCache(r *model.Rule) (*vm.Program, error) {
	key := fmt.Sprintf("%d:%d", r.Id, r.Version)
	if prog, ok := l.cache.Get(key); ok {
		return prog, nil
	}
	out, err, _ := l.sf.Do(key, func() (any, error) {
		if prog, ok := l.cache.Get(key); ok {
			return prog, nil
		}
		prog, err := engine.Compile(r.Expression)
		if err != nil {
			return nil, fmt.Errorf("compile rule code=%s: %w", r.Code, err)
		}
		l.cache.Add(key, prog)
		return prog, nil
	})
	if err != nil {
		return nil, err
	}
	return out.(*vm.Program), nil
}
