# SPEC-LSP-MULTI-006: Implementation Plan

## Phases

### Phase 1: Discovery Engine
- `internal/lsp/discovery/detector.go`: scan project root for marker files
- Match against `lsp.yaml.servers.*.project_markers`
- Return list of detected languages
- **LOC**: +200

### Phase 2: Binary Resolver
- `internal/lsp/discovery/resolver.go`: find language server binary
- Search order: project-local → PATH → fallback_binaries
- Return resolved path or error with install hint
- **LOC**: +150

### Phase 3: Per-Language Integration Tests
- 16 test files in `internal/lsp/core/integration/`:
  - `go_test.go` (gopls)
  - `python_test.go` (pylsp)
  - `typescript_test.go` (typescript-language-server)
  - `javascript_test.go` (typescript-language-server)
  - `rust_test.go` (rust-analyzer)
  - `java_test.go` (jdtls)
  - `kotlin_test.go` (kotlin-language-server)
  - `csharp_test.go` (omnisharp)
  - `ruby_test.go` (ruby-lsp)
  - `php_test.go` (phpactor)
  - `elixir_test.go` (elixir-ls)
  - `cpp_test.go` (clangd)
  - `scala_test.go` (metals)
  - `r_test.go` (R languageserver)
  - `flutter_test.go` (dart language-server)
  - `swift_test.go` (sourcekit-lsp)
- Each test: skip-if-missing + spawn + diagnostic assertion
- **LOC**: +2000 (~125 per test file)

### Phase 4: Doctor Subcommand
- `internal/cli/lsp_doctor.go`: `moai lsp doctor` command
- Detects languages, checks binaries, prints report
- Table output + JSON output via flag
- **LOC**: +300

### Phase 5: Capability Tracking
- `internal/lsp/core/capabilities.go`: capability map per client
- `Client.Capabilities() ServerCapabilities` method
- Used by downstream features (rename support check, etc.)
- **LOC**: +150

### Phase 6: Documentation
- `.claude/rules/moai/core/lsp-multi-language.md`: matrix reference
- README update with language installation instructions
- **LOC**: +100 docs

## Estimated LOC: +2900 total

## Dependencies

- **Hard**: SPEC-LSP-CORE-002 (powernap-based Manager), SPEC-LSP-AGG-003 (Aggregator)
- **Nice-to-have**: SPEC-LSP-QGATE-004 (for phase-aware integration)
