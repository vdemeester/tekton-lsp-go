# Tekton Language Server Protocol (LSP) - Go Implementation

> A Go-based Language Server Protocol implementation for Tekton YAML files providing intelligent IDE features like diagnostics, completion, hover documentation, and navigation.

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.23%2B-blue.svg)](https://go.dev/)
[![LSP](https://img.shields.io/badge/LSP-3.17-green.svg)](https://microsoft.github.io/language-server-protocol/)

**Status**: 🚧 **Under Active Development** 🚧

This is a port of the [Rust implementation](https://github.com/vdemeester/tekton-lsp) to Go for better alignment with the Tekton ecosystem and easier contribution to the tektoncd organization.

## Roadmap

- [x] Project initialization
- [ ] **Phase 1**: LSP server scaffold with document management (Week 1)
- [ ] **Phase 2**: Diagnostics & validation (Week 2)
- [ ] **Phase 3**: Completion provider (Week 3)
- [ ] **Phase 4**: Hover documentation (Week 4)
- [ ] **Phase 5**: Navigation (go-to-definition, symbols) (Week 5)
- [ ] **Phase 6**: Advanced features (formatting, code actions) (Week 6)
- [ ] **Phase 7**: Testing & integration (Week 7-8)

See [implementation plan](docs/IMPLEMENTATION_PLAN.md) for details.

## Quick Start

**Prerequisites:**
- Go 1.23+
- C compiler (for tree-sitter CGO bindings)

**Build:**
```bash
go build -o tekton-lsp ./cmd/tekton-lsp
```

**Run:**
```bash
./tekton-lsp
```

## Architecture

```
tekton-lsp-go/
├── cmd/tekton-lsp/      # Binary entry point
├── pkg/
│   ├── server/          # GLSP server implementation
│   ├── parser/          # Tree-sitter YAML parser
│   ├── cache/           # Document caching
│   ├── workspace/       # Workspace indexing
│   ├── validator/       # Tekton validation
│   ├── completion/      # Completion provider
│   ├── hover/           # Hover provider
│   ├── definition/      # Go-to-definition
│   ├── symbols/         # Document symbols
│   ├── formatting/      # YAML formatting
│   └── actions/         # Code actions
└── internal/tekton/     # Tekton types
```

## Technology Stack

- **LSP Framework**: [GLSP](https://github.com/tliron/glsp) - Language Server Protocol SDK for Go
- **YAML Parser**: [tree-sitter](https://github.com/tree-sitter/go-tree-sitter) with [tree-sitter-yaml](https://github.com/tree-sitter-grammars/tree-sitter-yaml)
- **Validation**: Custom Tekton validation + gopkg.in/yaml.v3
- **Testing**: Standard library + testify

## Rust vs Go

This is a port from the feature-complete [Rust implementation](https://github.com/vdemeester/tekton-lsp). 

**Why Go?**
- Better alignment with Tekton ecosystem (tektoncd is Go-centric)
- Easier contribution to tektoncd organization
- Leverages Go community familiarity
- Potential to reuse tektoncd/pipeline types

**Trade-offs:**
- Larger binary size (~15-30MB vs ~5-10MB)
- Slightly higher memory usage (GC overhead)
- Same feature parity goal

## Contributing

Contributions welcome! This project is in active development. See the [implementation plan](docs/IMPLEMENTATION_PLAN.md) for current status and next steps.

### Development Workflow

1. Fork the repository
2. Create a feature branch
3. Write tests first
4. Implement features
5. Run `go test ./...`
6. Submit a pull request

## License

Apache License 2.0 - See [LICENSE](LICENSE) for details.

## References

- [Tekton Documentation](https://tekton.dev/)
- [Language Server Protocol](https://microsoft.github.io/language-server-protocol/)
- [GLSP Framework](https://github.com/tliron/glsp)
- [Tree-sitter](https://tree-sitter.github.io/)
- [Original Rust Implementation](https://github.com/vdemeester/tekton-lsp)

---

**Status**: Phase 1 - Foundation  
**Last Updated**: 2026-02-11
