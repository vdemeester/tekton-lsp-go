# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-03-09

### Added
- **Multi-document YAML support** — all LSP features (diagnostics, hover, symbols, go-to-definition, completion, code actions) now work across `---`-separated documents in a single file
- **Nested field validation** — unknown fields detected in params, steps, workspaces, results, and pipeline task entries (e.g. `def: bar` in a param flags as warning)
- **Emacs eglot setup** example in README

### Fixed
- `metadata.generateName` accepted as alternative to `metadata.name` (common in PipelineRun/TaskRun)
- Formatting no longer destroys multi-document YAML files

## [0.1.1] - 2026-03-09

### Fixed
- Return empty diagnostics array instead of nil, fixing Neovim crash on `publishDiagnostics`
- Diagnostics not updating on document change; formatting using wrong indentation

### Changed
- Improved Neovim editor setup example in README (smart auto-attach, root_dir, coexistence with yamlls)

## [0.1.0] - 2026-02-11

### Added
- **LSP server** with GLSP framework (stdio and TCP transports)
- **Tree-sitter YAML parser** with incremental parsing and error recovery
- **Thread-safe document cache** for open documents
- **Diagnostics & validation**
  - Pipeline: tasks required, non-empty, correct type, unknown fields
  - Task: steps required, non-empty, step image required
  - Metadata: name required
  - Param refs: `$(params.NAME)` validated against declared params
  - taskRef: name required
  - Duplicate task name detection
- **Context-aware completion** for Pipeline, Task, Step, Metadata fields
- **Hover documentation** for 30+ Tekton fields (markdown)
- **Go-to-definition** for `taskRef` and `pipelineRef` across workspace
- **Document symbols** (outline view) for Pipeline tasks, Task steps
- **YAML formatting** with configurable indentation
- **Code actions** (quick fixes) for missing and unknown fields
- **Workspace scanning** — indexes YAML files on initialization
- **101+ tests** (unit + integration) with race detector
- **CI/CD** with GitHub Actions (build, lint, test)

[Unreleased]: https://github.com/vdemeester/tekton-lsp-go/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/vdemeester/tekton-lsp-go/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/vdemeester/tekton-lsp-go/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/vdemeester/tekton-lsp-go/releases/tag/v0.1.0
