# LSP Client Selection Rationale

SPEC: SPEC-LSP-CORE-002
Decision Date: 2026-04-12

## Selected Library

`github.com/charmbracelet/x/powernap` **v0.1.4** (pinned 2026-04-22)

Previous pin: v0.1.3 (2026-04-12 → 2026-04-22). Upgrade merged in #679.
The v0.1.3 → v0.1.4 delta inside the `powernap/` subdirectory is limited to
`powernap/pkg/config/lsps.json` (+11/-7) — a data-only refresh synced from
nvim-lspconfig. No Go code, API, or ABI change. Other 28 commits in the
charmbracelet/x range belong to unrelated subpackages (ansi/vt/cellbuf/...).

## Why powernap

powernap is a sub-package within `github.com/charmbracelet/x`, the monorepo backing
charmbracelet/crush (23k+ GitHub stars as of 2026-04-12).

Evidence from crush (C2 validation):
- `go.mod` of charmbracelet/crush requires `github.com/charmbracelet/x/powernap v0.1.3`
- crush spawns and manages real language-server subprocesses in production
- powernap wraps `github.com/sourcegraph/jsonrpc2` (VSCode-compatible codec)
  and adds `Connection`, `Router`, and `Transport` abstractions purpose-built for LSP

Key capabilities confirmed in source code review:
- `transport.Connection`: managed JSON-RPC connection over `io.ReadWriteCloser`
- `transport.Connection.Call()` / `Notify()`: request / notification primitives
- `transport.Router`: per-method handler dispatch (requests + notifications)
- `lsp.Client`: full lifecycle (initialize → ready → shutdown)
- `lsp.ClientConfig`: command, args, root URI, InitOptions pass-through
- Subprocess launch via `os/exec` with stdio pipes

## Alternatives Considered

| Option | Reason Rejected |
|--------|----------------|
| Path A — own JSON-RPC (SPEC-GOPLS-BRIDGE-001 style) | Narrow Go-only scope, not multi-language |
| Path B — MCP bridge (SPEC-LSPMCP-001) | Complementary, not a client replacement |
| golang.org/x/tools/gopls (as library) | Go-only; not a general LSP client library |
| sourcegraph/jsonrpc2 direct | Low-level; powernap already layers the LSP abstractions we need |

## Coexistence with SPEC-GOPLS-BRIDGE-001 (REQ-LC-010)

SPEC-GOPLS-BRIDGE-001 (Path A, hand-rolled) remains available and is not deprecated.
Users select via `lsp.client` config key:
- `lsp.client: lsp-core` → this SPEC (powernap-based, multi-language, default)
- `lsp.client: gopls-bridge` → SPEC-GOPLS-BRIDGE-001 (opt-in, Go-only)

Both paths are exercised in CI. Regression in SPEC-GOPLS-BRIDGE-001 test suite blocks
SPEC-LSP-CORE-002 merges.

## Upgrade Policy (REQ-LC-001a)

Before bumping the pinned version:

1. Run `go get github.com/charmbracelet/x/powernap@<new-tag>` in a dedicated branch.
2. Execute integration test suite:
   - `go test -tags=integration -race ./internal/lsp/core/... -run TestGoLSP` (gopls)
   - `go test -tags=integration -race ./internal/lsp/core/... -run TestPythonLSP` (pyright / pylsp)
   - `go test -tags=integration -race ./internal/lsp/core/... -run TestTypeScriptLSP` (tsserver)
3. All three language integration tests must pass before the PR is merged.
4. Update the pinned version line in this document with the new version and date.

## Technical Constraints

- powernap uses `github.com/sourcegraph/jsonrpc2` internally; do NOT add that
  package as a direct dependency of `internal/lsp/` — use powernap's exported API only.
- powernap's `lsp.Client` handles subprocess lifecycle; `internal/lsp/transport/`
  wraps the `transport.Connection` API to expose MoAI's own `Transport` interface.
- Multi-language neutrality: MoAI supports 16 languages. powernap's `lsp.ClientConfig`
  accepts arbitrary command + args + initOptions, enabling all language servers equally.

---

Version: 1.0.0
Source: SPEC-LSP-CORE-002 Decision Point 1 (2026-04-12)
