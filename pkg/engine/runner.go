package engine

import (
	"context"
	"go/token"
	"os"
	"runtime"
	"sync"

	"github.com/nicholas/glint/pkg/rule"
	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/go/packages"
)

type Runner struct {
	walker      *Walker
	cache       *Cache
	concurrency int
	ruleSetKey  string
}

func NewRunner(walker *Walker, cache *Cache, concurrency int, ruleSetKey string) *Runner {
	if concurrency <= 0 {
		concurrency = runtime.NumCPU()
	}
	return &Runner{
		walker:      walker,
		cache:       cache,
		concurrency: concurrency,
		ruleSetKey:  ruleSetKey,
	}
}

type fileUnit struct {
	pkg      *packages.Package
	fileIdx  int
	filePath string
}

// Run analyzes all packages in parallel and returns collected diagnostics.
func (r *Runner) Run(ctx context.Context, pkgs []*packages.Package) ([]rule.Diagnostic, error) {
	var units []fileUnit
	for _, pkg := range pkgs {
		for i, f := range pkg.CompiledGoFiles {
			_ = f
			units = append(units, fileUnit{
				pkg:      pkg,
				fileIdx:  i,
				filePath: pkg.CompiledGoFiles[i],
			})
		}
	}

	var (
		mu       sync.Mutex
		allDiags []rule.Diagnostic
	)

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(r.concurrency)

	for _, u := range units {
		u := u
		g.Go(func() error {
			if gctx.Err() != nil {
				return gctx.Err()
			}

			src, err := os.ReadFile(u.filePath)
			if err != nil {
				return nil // skip unreadable files
			}
			fileHash := HashFile(src)

			if cached, ok := r.cache.Lookup(u.filePath, fileHash, r.ruleSetKey); ok {
				mu.Lock()
				allDiags = append(allDiags, cached...)
				mu.Unlock()
				return nil
			}

			if u.fileIdx >= len(u.pkg.Syntax) {
				return nil
			}

			rctx := &rule.Context{
				File:     u.pkg.Syntax[u.fileIdx],
				FileSet:  u.pkg.Fset,
				TypeInfo: u.pkg.TypesInfo,
				Pkg:      u.pkg.Types,
				FileHash: fileHash,
				FilePath: u.filePath,
			}

			diags := r.walker.Walk(rctx)

			r.cache.Store(u.filePath, fileHash, r.ruleSetKey, diags)

			mu.Lock()
			allDiags = append(allDiags, diags...)
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return allDiags, err
	}

	sortDiagnostics(allDiags)
	return allDiags, nil
}

func sortDiagnostics(diags []rule.Diagnostic) {
	for i := 1; i < len(diags); i++ {
		for j := i; j > 0 && lessPosition(diags[j].Pos, diags[j-1].Pos); j-- {
			diags[j], diags[j-1] = diags[j-1], diags[j]
		}
	}
}

func lessPosition(a, b token.Position) bool {
	if a.Filename != b.Filename {
		return a.Filename < b.Filename
	}
	if a.Line != b.Line {
		return a.Line < b.Line
	}
	return a.Column < b.Column
}
