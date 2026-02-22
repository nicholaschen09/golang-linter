# glint

A super fast Go linter built for speed. Glint uses single-pass AST walking with multi-rule dispatch, parallel package analysis, lazy type-checking, and file-level result caching to lint Go code significantly faster than traditional multi-pass approaches.

## Install

```bash
go install github.com/nicholas/glint/cmd/glint@latest
```

Or build from source:

```bash
git clone https://github.com/nicholas/glint.git
cd glint
go build -o glint ./cmd/glint
```

## Quick Start

```bash
# Lint the current package
glint run --enable-all ./...

# Generate a config file
glint init

# List available rules
glint rules
```

## Performance

Glint is fast because of its architecture:

| Technique | What it does |
|---|---|
| **Single-pass AST walk** | Parses each file once, walks the AST once, and dispatches to all matching rules per node via a `reflect.Type` lookup table |
| **Parallel analysis** | Fans out file analysis across all CPU cores using a bounded worker pool (`errgroup`) |
| **Lazy type-checking** | Only invokes `go/types` when at least one active rule needs type information; pure-AST rules skip it entirely |
| **File-level caching** | SHA-256 hashes each file and caches results to `~/.cache/glint/`; unchanged files are skipped on re-runs |
| **Arena allocation** | Uses `sync.Pool` for diagnostic slices to reduce GC pressure |

## Rules

### Bugs

| Rule | Severity | Description |
|---|---|---|
| `unchecked-error` | error | Ignoring returned error values |
| `nil-deref` | error | Dereference after type assertion or map lookup without ok check |
| `shadow-var` | warning | Variable shadowing in inner scopes |

### Style

| Rule | Severity | Description |
|---|---|---|
| `naming-convention` | warning | Exported names must be MixedCaps; enforces Go acronym conventions |
| `import-order` | info | Import grouping: stdlib, then external, then internal |
| `line-length` | warning | Lines exceeding 120 characters (configurable) |

### Performance

| Rule | Severity | Description |
|---|---|---|
| `prealloc-slice` | warning | Slices grown in loops without preallocation |
| `unnecessary-conversion` | warning | Redundant type conversions |

### Security

| Rule | Severity | Description |
|---|---|---|
| `hardcoded-secret` | error | Hardcoded passwords, API keys, tokens in string literals |
| `sql-injection` | error | String concatenation in SQL query functions |

## Configuration

Create a `.glint.yml` in your project root, or run `glint init` to generate one:

```yaml
rules:
  unchecked-error:
    enabled: true
    severity: error
  line-length:
    enabled: true
    severity: warning
    options:
      max: 120
  hardcoded-secret:
    enabled: true
    severity: error

cache:
  enabled: true
  dir: ~/.cache/glint

output:
  format: text   # text | json | sarif
  color: true

concurrency: 0   # 0 = runtime.NumCPU()
```

## CLI Reference

```
glint run [flags] [packages...]

Flags:
  -c, --config string      path to config file
  -f, --format string      output format: text, json, sarif
  -j, --concurrency int    worker count (0 = NumCPU)
      --enable-all         enable all rules regardless of config
      --no-cache           disable result caching

glint rules               list all available rules
glint init                generate a default .glint.yml
```

## Output Formats

**Text** (default) — human-readable colored output for terminals.

**JSON** — machine-readable array of diagnostics for editor integrations.

**SARIF** — Static Analysis Results Interchange Format for CI systems (GitHub Code Scanning, Azure DevOps).

## Adding Custom Rules

Implement the `rule.Rule` interface and register via `init()`:

```go
package myrule

import (
    "go/ast"
    "github.com/nicholas/glint/pkg/rule"
)

type MyRule struct{}

func (MyRule) Name() string            { return "my-rule" }
func (MyRule) Category() rule.Category { return rule.CategoryBugs }
func (MyRule) Severity() rule.Severity { return rule.SeverityWarning }
func (MyRule) Description() string     { return "My custom rule" }
func (MyRule) NeedsTypeInfo() bool     { return false }
func (MyRule) NodeTypes() []ast.Node   { return []ast.Node{(*ast.CallExpr)(nil)} }

func (MyRule) Check(ctx *rule.Context, node ast.Node) []rule.Diagnostic {
    // your logic here
    return nil
}

func init() {
    rule.Register(MyRule{})
}
```

For file-level rules (e.g., import ordering), also implement the `rule.FileRule` interface with a `CheckFile(ctx *rule.Context) []rule.Diagnostic` method.

## License

MIT
