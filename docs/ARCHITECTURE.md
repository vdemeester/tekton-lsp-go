# Architecture

## Overview

tekton-lsp-go is a Language Server Protocol (LSP) implementation for Tekton YAML files, written in Go.

## Technology Stack

- **LSP Framework**: [GLSP](https://github.com/tliron/glsp) - Language Server Protocol SDK for Go
  - Provides JSON-RPC 2.0 server with stdio, TCP, WebSocket transports
  - Complete LSP 3.16+ protocol implementation
  - Handler-based API

- **YAML Parser**: [tree-sitter](https://github.com/tree-sitter/go-tree-sitter)
  - Incremental parsing for efficient updates
  - Precise position tracking (line/column)
  - Error recovery for partial/invalid YAML
  - Grammar: [tree-sitter-yaml](https://github.com/tree-sitter-grammars/tree-sitter-yaml)

- **Validation**: Custom Tekton validation + [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3)
  - Tree-sitter for AST and positions
  - yaml.v3 for structure validation and formatting
  - Tekton-specific semantic validation

## Project Structure

```
tekton-lsp-go/
├── cmd/tekton-lsp/            # Binary entry point
│   └── main.go                # CLI flags, server startup
│
├── pkg/                       # Public API packages
│   ├── server/                # LSP server (GLSP handlers)
│   │   ├── server.go          # Server creation, handler wiring
│   │   ├── lifecycle.go       # initialize, initialized, shutdown
│   │   ├── document.go        # didOpen, didChange, didClose
│   │   ├── diagnostics.go     # publishDiagnostics
│   │   ├── completion.go      # textDocument/completion
│   │   ├── hover.go           # textDocument/hover
│   │   ├── symbols.go         # textDocument/documentSymbol
│   │   ├── formatting.go      # textDocument/formatting
│   │   ├── definition.go      # textDocument/definition
│   │   └── actions.go         # textDocument/codeAction
│   │
│   ├── parser/                # YAML parsing (tree-sitter)
│   │   ├── parser.go          # ParseYAML, tree-sitter wrapper
│   │   └── ast.go             # Document, Node, Range, Position
│   │
│   ├── cache/                 # Thread-safe document cache
│   │   └── cache.go           # Insert/Get/Update/Remove/AllParsed
│   │
│   ├── validator/             # Tekton validation
│   │   └── validator.go       # Pipeline/Task/metadata validation
│   │
│   ├── completion/            # Context-aware completions
│   │   ├── provider.go        # Complete(), context detection
│   │   └── schemas.go         # Field schemas per context
│   │
│   ├── hover/                 # Hover documentation
│   │   ├── provider.go        # Hover(), node lookup
│   │   └── docs.go            # 30+ field documentation entries
│   │
│   ├── definition/            # Go-to-definition
│   │   └── provider.go        # taskRef/pipelineRef resolution
│   │
│   ├── symbols/               # Document outline
│   │   └── provider.go        # DocumentSymbols(), AST → outline
│   │
│   ├── formatting/            # YAML formatting
│   │   └── formatter.go       # Format() via yaml.v3
│   │
│   └── actions/               # Code actions (quick fixes)
│       └── provider.go        # Add missing / remove unknown
│
└── test/
    ├── integration/           # LSP protocol tests (end-to-end)
    │   ├── lsp_client.go      # JSON-RPC client over stdio
    │   └── integration_test.go # 14 protocol tests
    └── fixtures/              # Sample YAML files
        ├── pipeline.yaml
        └── task.yaml
```

## Component Interaction

```
┌─────────────────────────────────────────────┐
│              LSP Client                      │
│         (VS Code, Neovim, etc.)              │
└────────────────┬────────────────────────────┘
                 │ JSON-RPC (stdio/TCP)
                 ▼
┌─────────────────────────────────────────────┐
│          GLSP Server (pkg/server/)          │
│  9 handlers: completion, hover, symbols,     │
│  formatting, definition, codeAction,         │
│  diagnostics, didOpen/Change/Close           │
└────────────┬────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────┐
│         Document Cache (pkg/cache/)         │
│  Thread-safe, stores content + parsed AST    │
└────────────┬────────────────────────────────┘
             │
    ┌────────┼────────┬────────┐
    ▼        ▼        ▼        ▼
┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐
│ Parser │ │Validator│ │Compltn │ │ Hover  │
│tree-   │ │Tekton  │ │Schema  │ │ Docs   │
│sitter  │ │checks  │ │context │ │ lookup │
└────────┘ └────────┘ └────────┘ └────────┘
             ┌────────┐ ┌────────┐ ┌────────┐
             │Symbols │ │ Format │ │Actions │
             │outline │ │yaml.v3 │ │quickfix│
             └────────┘ └────────┘ └────────┘
             ┌────────┐
             │ Defn   │
             │taskRef │
             └────────┘
```

## Feature Implementation Status

All 9 LSP features implemented with TDD:

| Phase | Feature | Package | Tests | Status |
|-------|---------|---------|-------|--------|
| 1 | Tree-sitter YAML parsing | `pkg/parser` | 8 | ✅ |
| 1 | Document cache | `pkg/cache` | 9 | ✅ |
| 1 | Document sync | `pkg/server` | 2 | ✅ |
| 2 | Validation & diagnostics | `pkg/validator` + server | 17 | ✅ |
| 3 | Completion | `pkg/completion` + server | 10 | ✅ |
| 4 | Hover | `pkg/hover` | 8 | ✅ |
| 5 | Document symbols | `pkg/symbols` | 5 | ✅ |
| 5 | Go-to-definition | `pkg/definition` | 4 | ✅ |
| 6 | Formatting | `pkg/formatting` | 4 | ✅ |
| 6 | Code actions | `pkg/actions` | 5 | ✅ |
| 7 | Integration tests | `test/integration` | 14 | ✅ |
| | **Total** | | **86** | ✅ |

## Design Decisions

### Why GLSP?
- Complete LSP protocol coverage
- Built-in transport layer (stdio, TCP, WebSocket)
- Simple handler-based API
- Production-proven (Puccini, zk)
- Active maintenance

### Why tree-sitter?
- Incremental parsing (fast updates)
- Precise position tracking
- Error recovery for partial/invalid YAML
- Used by many editors (Neovim, Emacs)

### Why minimal Tekton types?
- Avoid heavy kubernetes dependencies
- Fast builds (~10MB binary vs ~50MB+ with k8s)
- Custom validation for LSP needs
- Can add tektoncd/pipeline later if beneficial

## Dependencies

### Core
- `github.com/tliron/glsp` - LSP framework
- `github.com/tree-sitter/go-tree-sitter` v0.24.0 - Tree-sitter bindings
- `github.com/tree-sitter-grammars/tree-sitter-yaml` - YAML grammar
- `gopkg.in/yaml.v3` - YAML formatting

### Testing
- `github.com/stretchr/testify` - Test assertions

### Build
- Go 1.25
- C compiler (CGO for tree-sitter)

## References

- [LSP Specification](https://microsoft.github.io/language-server-protocol/)
- [GLSP Documentation](https://pkg.go.dev/github.com/tliron/glsp)
- [Tree-sitter Documentation](https://tree-sitter.github.io/)
- [Original Rust Implementation](https://github.com/vdemeester/tekton-lsp)
