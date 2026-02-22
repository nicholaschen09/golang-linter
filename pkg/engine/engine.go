package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/nicholas/glint/pkg/config"
	"github.com/nicholas/glint/pkg/loader"
	"github.com/nicholas/glint/pkg/rule"
)

type Engine struct {
	cfg    *config.Config
	rules  []rule.Rule
	cache  *Cache
	runner *Runner
}

func New(cfg *config.Config, registry *rule.Registry) (*Engine, error) {
	var activeRules []rule.Rule
	needsTypes := false

	allRules := registry.All()
	sort.Slice(allRules, func(i, j int) bool {
		return allRules[i].Name() < allRules[j].Name()
	})

	for _, r := range allRules {
		rc, exists := cfg.Rules[r.Name()]
		if exists && !rc.Enabled {
			continue
		}
		if !exists && !cfg.EnableAll {
			continue
		}
		activeRules = append(activeRules, r)
		if r.NeedsTypeInfo() {
			needsTypes = true
		}
	}

	if len(activeRules) == 0 {
		return nil, fmt.Errorf("no rules enabled; enable rules in .glint.yml or use --enable-all")
	}

	cacheDir := cfg.Cache.Dir
	cache, err := NewCache(cacheDir, cfg.Cache.Enabled)
	if err != nil {
		return nil, fmt.Errorf("initializing cache: %w", err)
	}

	walker := NewWalker(activeRules)
	ruleSetKey := computeRuleSetKey(activeRules)

	_ = needsTypes

	runner := NewRunner(walker, cache, cfg.Concurrency, ruleSetKey)

	return &Engine{
		cfg:    cfg,
		rules:  activeRules,
		cache:  cache,
		runner: runner,
	}, nil
}

func (e *Engine) Run(ctx context.Context, patterns []string) ([]rule.Diagnostic, error) {
	needsTypes := false
	for _, r := range e.rules {
		if r.NeedsTypeInfo() {
			needsTypes = true
			break
		}
	}

	mode := loader.LoadSyntax
	if needsTypes {
		mode = loader.LoadTypes
	}

	result, err := loader.Load(patterns, mode, nil)
	if err != nil {
		return nil, fmt.Errorf("loading packages: %w", err)
	}

	return e.runner.Run(ctx, result.Packages)
}

func (e *Engine) ActiveRules() []rule.Rule {
	return e.rules
}

func (e *Engine) ClearCache() error {
	return e.cache.Clear()
}

func computeRuleSetKey(rules []rule.Rule) string {
	var names []string
	for _, r := range rules {
		names = append(names, r.Name())
	}
	sort.Strings(names)
	h := sha256.Sum256([]byte(strings.Join(names, ",")))
	return hex.EncodeToString(h[:8])
}
