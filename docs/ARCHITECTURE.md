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
  - yaml.v3 for structure validation
  - Tekton-specific semantic validation

## Project Structure

```
tekton-lsp-go/
├── cmd/
│   └── tekton-lsp/          # Binary entry point
│       └── main.go          # CLI and server initialization
│
├── pkg/                     # Public API packages
│   ├── server/              # LSP server implementation
│   │   ├── server.go        # Server creation and transport
│   │   ├── lifecycle.go     # LSP lifecycle (initialize, shutdown)
│   │   └── document.go      # Document sync (didOpen, didChange)
│   │
│   ├── parser/              # YAML parsing with tree-sitter
│   │   ├── parser.go        # Parser wrapper
│   │   ├── ast.go           # AST types
│   │   └── positions.go     # Position conversion (LSP ↔ tree-sitter)
│   │
│   ├── cache/               # Document caching
│   │   ├── document.go      # Document state
│   │   └── cache.go         # Document cache management
│   │
│   ├── workspace/           # Workspace indexing
│   │   ├── index.go         # Cross-file reference index
│   │   └── scanner.go       # Workspace scanner
│   │
│   ├── validator/           # Tekton validation
│   │   ├── validator.go     # Main validator
│   │   ├── pipeline.go      # Pipeline validation
│   │   ├── task.go          # Task validation
│   │   └── diagnostics.go   # Diagnostic conversion
│   │
│   ├── completion/          # Completion provider
│   │   ├── provider.go      # Completion logic
│   │   ├── schemas.go       # Tekton field schemas
│   │   └── context.go       # Context analysis (cursor position)
│   │
│   ├── hover/               # Hover provider
│   │   ├── provider.go      # Hover logic
│   │   └── docs.go          # Documentation lookup
│   │
│   ├── definition/          # Go-to-definition provider
│   │   ├── provider.go      # Definition resolution
│   │   └── references.go    # Reference tracking
│   │
│   ├── symbols/             # Document symbols provider
│   │   └── provider.go      # Symbol extraction
│   │
│   ├── formatting/          # YAML formatting provider
│   │   └── formatter.go     # Format logic
│   │
│   └── actions/             # Code actions provider
│       └── provider.go      # Quick fixes
│
├── internal/                # Private packages
│   └── tekton/              # Tekton types (may become separate module)
│       ├── types.go         # Common types
│       └── v1/              # Tekton v1 API
│           ├── pipeline.go  # Pipeline types
│           └── task.go      # Task types
│
└── test/
    ├── integration/         # Integration tests
    └── fixtures/            # Test YAML files
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
│                                              │
│  ┌──────────────┐   ┌──────────────┐       │
│  │  Lifecycle   │   │  Document    │       │
│  │   Handlers   │   │   Handlers   │       │
│  └──────────────┘   └──────────────┘       │
└────────────┬────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────┐
│         Document Cache (pkg/cache/)         │
│                                              │
│  ┌────────┐  ┌────────┐  ┌────────┐        │
│  │ Doc 1  │  │ Doc 2  │  │ Doc N  │        │
│  │ + AST  │  │ + AST  │  │ + AST  │        │
│  └────────┘  └────────┘  └────────┘        │
└────────────┬────────────────────────────────┘
             │
    ┌────────┼────────┐
    ▼        ▼        ▼
┌────────┐ ┌────────┐ ┌────────┐
│ Parser │ │Validator│ │Features│
│        │ │        │ │Provider│
│tree-   │ │Tekton  │ │- Hover │
│sitter  │ │Schema  │ │- Goto  │
│+ yaml  │ │Checks  │ │- Comp  │
└────────┘ └────────┘ └────────┘
```

## LSP Lifecycle

```
1. Client starts server
   ↓
2. Client → Server: initialize
   ↓
3. Server → Client: initialize result (capabilities)
   ↓
4. Client → Server: initialized
   ↓
5. Client → Server: textDocument/didOpen
   ↓
6. Server parses document, runs validation
   ↓
7. Server → Client: textDocument/publishDiagnostics
   ↓
8. Client → Server: textDocument/completion (on user input)
   ↓
9. Server → Client: completion items
   ↓
   ...
   ↓
10. Client → Server: shutdown
    ↓
11. Client → Server: exit
```

## Feature Implementation Status

### Phase 1: Foundation ✅ (In Progress)
- [x] LSP server scaffold with GLSP
- [x] Document lifecycle (didOpen, didChange, didClose)
- [ ] Tree-sitter YAML parser integration
- [ ] Document caching

### Phase 2: Diagnostics
- [ ] Tekton type definitions
- [ ] Pipeline validation
- [ ] Task validation
- [ ] Diagnostic publishing

### Phase 3: Completion
- [ ] Schema definitions
- [ ] Context analysis
- [ ] Completion provider

### Phase 4: Hover
- [ ] Documentation database
- [ ] Hover provider

### Phase 5: Navigation
- [ ] Workspace indexing
- [ ] Go-to-definition
- [ ] Document symbols

### Phase 6: Advanced
- [ ] YAML formatting
- [ ] Code actions

### Phase 7: Testing
- [ ] Integration tests
- [ ] Protocol tests
- [ ] VS Code extension

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
- Error recovery
- Used by many editors (Neovim, Emacs)

### Why minimal Tekton types?
- Avoid heavy kubernetes dependencies
- Fast builds (~10-15MB binary vs ~50MB+ with k8s)
- Custom validation for LSP needs
- Can add tektoncd/pipeline later if beneficial

## Performance Targets

Based on Rust implementation benchmarks:

| Operation | Target | Notes |
|-----------|--------|-------|
| didOpen | < 50ms | Parse + cache |
| didChange | < 10ms | Incremental parse |
| diagnostics | < 100ms | Validation |
| completion | < 50ms | Schema lookup |
| hover | < 20ms | Doc lookup |
| definition | < 50ms | Reference resolution |

## Dependencies

### Core
- `github.com/tliron/glsp` - LSP framework
- `github.com/tree-sitter/go-tree-sitter` - Tree-sitter bindings
- `gopkg.in/yaml.v3` - YAML validation

### Testing
- `github.com/stretchr/testify` - Test assertions

### Build
- Go 1.23+
- C compiler (for tree-sitter CGO)

## References

- [LSP Specification](https://microsoft.github.io/language-server-protocol/)
- [GLSP Documentation](https://pkg.go.dev/github.com/tliron/glsp)
- [Tree-sitter Documentation](https://tree-sitter.github.io/)
- [Original Rust Implementation](https://github.com/vdemeester/tekton-lsp)
