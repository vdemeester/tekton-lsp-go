# Tekton Language Server Protocol (LSP) - Go Implementation

> A Language Server Protocol implementation for Tekton YAML files providing intelligent IDE features: diagnostics, completion, hover, navigation, formatting, and code actions.

[![CI](https://github.com/vdemeester/tekton-lsp-go/actions/workflows/ci.yaml/badge.svg)](https://github.com/vdemeester/tekton-lsp-go/actions/workflows/ci.yaml)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.25-blue.svg)](https://go.dev/)
[![LSP](https://img.shields.io/badge/LSP-3.16-green.svg)](https://microsoft.github.io/language-server-protocol/)

A port of the [Rust implementation](https://github.com/vdemeester/tekton-lsp) to Go for better alignment with the Tekton ecosystem and easier contribution to the tektoncd organization.

## Features

| Feature | Description |
|---------|-------------|
| **Diagnostics** | Validates Pipeline/Task structure, required fields, unknown fields |
| **Completion** | Context-aware field suggestions for Pipeline, Task, Step, Metadata |
| **Hover** | Documentation for 30+ Tekton fields with markdown formatting |
| **Go-to-definition** | Jump from `taskRef`/`pipelineRef` to the referenced resource |
| **Document symbols** | Outline view of Pipeline tasks, Task steps, params |
| **Formatting** | Consistent YAML indentation (configurable) |
| **Code actions** | Quick fixes: add missing fields, remove unknown fields |

## Quick Start

**Prerequisites:**
- Go 1.25+
- C compiler (for tree-sitter CGO bindings)

**Build:**
```bash
go build -o tekton-lsp ./cmd/tekton-lsp
```

**Run (stdio, default):**
```bash
./tekton-lsp
```

**Run (TCP):**
```bash
./tekton-lsp --tcp --address :7999
```

**Test:**
```bash
go test -race ./...
```

## Editor Setup

### VS Code

Add to `.vscode/settings.json`:
```json
{
  "yaml.customTags": [],
  "tekton-lsp.path": "/path/to/tekton-lsp"
}
```

### Neovim (nvim-lspconfig)

```lua
vim.api.nvim_create_autocmd("FileType", {
  pattern = "yaml",
  callback = function()
    vim.lsp.start({
      name = "tekton-lsp",
      cmd = { "/path/to/tekton-lsp" },
    })
  end,
})
```

## Architecture

```
pkg/
├── server/       # GLSP server — 9 LSP handlers
├── parser/       # tree-sitter YAML → AST with positions
├── cache/        # Thread-safe document cache
├── validator/    # Pipeline/Task structure validation
├── completion/   # Schema-based context-aware completions
├── hover/        # Field documentation (30+ entries)
├── definition/   # taskRef/pipelineRef → definition resolution
├── symbols/      # Document outline extraction
├── formatting/   # YAML reformatting via yaml.v3
└── actions/      # Quick fix code actions
test/
└── integration/  # LSP protocol tests (14 end-to-end tests)
```

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for details.

## Technology Stack

- **LSP Framework**: [GLSP](https://github.com/tliron/glsp) — complete LSP 3.16+ with stdio/TCP transports
- **YAML Parser**: [tree-sitter](https://github.com/tree-sitter/go-tree-sitter) — incremental parsing, precise positions, error recovery
- **Formatting**: [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) — consistent indentation
- **Testing**: 86 tests (72 unit + 14 integration) with `-race`

## Why Go?

- Better alignment with Tekton ecosystem (tektoncd is Go-centric)
- Easier contribution to tektoncd organization
- Leverages Go community familiarity
- ~10MB binary (minimal deps, no k8s dependency)

## Contributing

1. Fork the repository
2. Create a feature branch
3. **Write tests first** (TDD)
4. Implement features
5. Run `go test -race ./...`
6. Submit a pull request

## License

Apache License 2.0 — See [LICENSE](LICENSE) for details.
