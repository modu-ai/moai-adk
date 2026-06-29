# LSP Client Selection Rationale

## Selected Library

`github.com/charmbracelet/x/powernap` **v0.1.4**

The pin tracks the `powernap/` subdirectory of charmbracelet/x. Routine version bumps
within that subdirectory are typically data-only refreshes (the bundled
`powernap/pkg/config/lsps.json`, synced from nvim-lspconfig) with no Go code, API, or
ABI change; other commits in the charmbracelet/x range belong to unrelated subpackages
(ansi / vt / cellbuf / ...).

## Why powernap

powernap is a sub-package within `github.com/charmbracelet/x`, the monorepo backing
charmbracelet/crush (a popular production TUI coding agent).

Evidence from crush:
- `go.mod` of charmbracelet/crush requires `github.com/charmbracelet/x/powernap`
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
| Own JSON-RPC (hand-rolled Go-only bridge) | Narrow Go-only scope, not multi-language |
| MCP bridge | Complementary, not a client replacement |
| golang.org/x/tools/gopls (as library) | Go-only; not a general LSP client library |
| sourcegraph/jsonrpc2 direct | Low-level; powernap already layers the LSP abstractions we need |

## Coexistence with the Legacy Go-only Bridge

The hand-rolled Go-only bridge remains available and is not deprecated. Users select
via the `client_impl` config key in `.moai/config/sections/lsp.yaml`:

- `client_impl: gopls_bridge` → legacy Go-only bridge (default until the powernap client fully lands)
- `client_impl: powernap_core` → powernap-based, multi-language foundation (opt-in)

Both paths are exercised in CI. A regression in the Go-only bridge test suite blocks
powernap-client merges.

## Upgrade Policy

Before bumping the pinned version:

1. Run `go get github.com/charmbracelet/x/powernap@<new-tag>` in a dedicated branch.
2. Execute the integration test suite:
   - `go test -tags=integration -race ./internal/lsp/core/... -run TestGoLSP` (gopls)
   - `go test -tags=integration -race ./internal/lsp/core/... -run TestPythonLSP` (pyright / pylsp)
   - `go test -tags=integration -race ./internal/lsp/core/... -run TestTypeScriptLSP` (tsserver)
3. All three language integration tests must pass before the change is merged.
4. Update the pinned version line in this document with the new version.

## Technical Constraints

- powernap uses `github.com/sourcegraph/jsonrpc2` internally; do NOT add that
  package as a direct dependency of `internal/lsp/` — use powernap's exported API only.
- powernap's `lsp.Client` handles subprocess lifecycle; `internal/lsp/transport/`
  wraps the `transport.Connection` API to expose MoAI's own `Transport` interface.
- Multi-language neutrality: MoAI supports 16 languages. powernap's `lsp.ClientConfig`
  accepts arbitrary command + args + initOptions, enabling all language servers equally.
