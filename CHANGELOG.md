# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/vdemeester/tekton-lsp-go/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/vdemeester/tekton-lsp-go/releases/tag/v0.1.0
