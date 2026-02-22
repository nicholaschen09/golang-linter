package engine

import (
	"go/ast"
	"reflect"
	"sync"

	"github.com/nicholas/glint/pkg/rule"
)

// Walker performs a single-pass traversal of an AST file, dispatching
// each node to only the rules that registered interest in that node type.
type Walker struct {
	dispatchTable map[reflect.Type][]rule.Rule
	fileRules     []rule.FileRule
	diagPool      sync.Pool
}

func NewWalker(rules []rule.Rule) *Walker {
	w := &Walker{
		dispatchTable: make(map[reflect.Type][]rule.Rule),
		diagPool: sync.Pool{
			New: func() any {
				s := make([]rule.Diagnostic, 0, 8)
				return &s
			},
		},
	}

	for _, r := range rules {
		if fr, ok := r.(rule.FileRule); ok {
			w.fileRules = append(w.fileRules, fr)
		}
		for _, nodeProto := range r.NodeTypes() {
			t := reflect.TypeOf(nodeProto)
			w.dispatchTable[t] = append(w.dispatchTable[t], r)
		}
	}

	return w
}

// Walk performs a single traversal of the file AST and returns all
// diagnostics produced by registered rules.
func (w *Walker) Walk(ctx *rule.Context) []rule.Diagnostic {
	poolVal, _ := w.diagPool.Get().(*[]rule.Diagnostic)
	if poolVal == nil {
		s := make([]rule.Diagnostic, 0, 8)
		poolVal = &s
	}
	buf := poolVal
	*buf = (*buf)[:0]
	defer w.diagPool.Put(buf)

	for _, fr := range w.fileRules {
		*buf = append(*buf, fr.CheckFile(ctx)...)
	}

	ast.Inspect(ctx.File, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		t := reflect.TypeOf(n)
		rules, ok := w.dispatchTable[t]
		if !ok {
			return true
		}
		for _, r := range rules {
			*buf = append(*buf, r.Check(ctx, n)...)
		}
		return true
	})

	out := make([]rule.Diagnostic, len(*buf))
	copy(out, *buf)
	return out
}
