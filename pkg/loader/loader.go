package loader

import (
	"fmt"

	"golang.org/x/tools/go/packages"
)

type LoadMode int

const (
	// LoadSyntax loads parsed AST but no type information.
	LoadSyntax LoadMode = iota
	// LoadTypes loads full type information in addition to AST.
	LoadTypes
)

type Result struct {
	Packages []*packages.Package
}

// Load loads Go packages at the given patterns. The mode controls
// whether type information is resolved — skipping it is significantly
// faster when only AST-level rules are active.
func Load(patterns []string, mode LoadMode, buildFlags []string) (*Result, error) {
	cfg := &packages.Config{
		BuildFlags: buildFlags,
	}

	switch mode {
	case LoadSyntax:
		cfg.Mode = packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedImports |
			packages.NeedCompiledGoFiles
	case LoadTypes:
		cfg.Mode = packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedImports |
			packages.NeedCompiledGoFiles |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedTypesSizes |
			packages.NeedDeps
	default:
		return nil, fmt.Errorf("unknown load mode: %d", mode)
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("loading packages: %w", err)
	}

	var errs []error
	for _, pkg := range pkgs {
		for _, e := range pkg.Errors {
			errs = append(errs, e)
		}
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("package errors: %v", errs)
	}

	return &Result{Packages: pkgs}, nil
}
