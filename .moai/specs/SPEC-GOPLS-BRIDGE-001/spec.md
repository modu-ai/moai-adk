---
id: SPEC-GOPLS-BRIDGE-001
version: "1.0.0"
status: draft
created: "2026-04-11"
updated: "2026-04-11"
author: GOOS
priority: P1
issue_number: 0
phase: "Phase 3 - Quality Infrastructure"
module: "internal/lsp/gopls/, internal/loop/, internal/hook/quality/"
estimated_loc: 400
dependencies: []
lifecycle: spec-anchored
tags: gopls, lsp, json-rpc, go-toolchain, quality-gate
---

# SPEC-GOPLS-BRIDGE-001: Go-only gopls Subprocess Bridge

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-04-11 | 1.0.0 | Initial draft (Path C secondary, Go-specific runtime) |

---

## Overview

| Field | Value |
|-------|-------|
| SPEC ID | SPEC-GOPLS-BRIDGE-001 |
| Title | Go-only gopls Subprocess Bridge |
| Status | Draft |
| Priority | P1 (Path C secondary) |
| Research Source | R1/R4/C2 reports (2026-04-11 comparative audit) |

---

## Problem Statement

`internal/loop/go_feedback.go` currently uses only `go test` and `go vet` CLI output. It does not collect LSP diagnostics from gopls, staticcheck-via-gopls analyses, or inlay type hints. As a result, Ralph Engine decides purely on exit codes and integer counts — losing the rich severity/source metadata LSP provides.

The previous audit (report A4) confirmed:
- `internal/cli/deps.go` injects `nil` as LSP client (PR #623 added a warn log, but the nil remains)
- `internal/lsp/` contains only interfaces, no client implementation
- `GoFeedbackGenerator` (shipped in commit `f6a6f2cc`) uses `exec.Command("go", "test", ...)` only

Per Path C (chosen in the LSP strategy decision), moai-adk-go's **local development** needs a minimal Go-only runtime LSP client via gopls subprocess. This is separate from SPEC-LSP-CORE-002 (full multi-language LSP client based on charmbracelet/powernap).

---

## Goal

Implement a **self-contained, dependency-free gopls subprocess bridge** that:

1. **Spawns gopls** on demand (lazy initialization)
2. **Speaks LSP via stdio JSON-RPC 2.0** using hand-rolled framing (no external library)
3. **Collects diagnostics** via `textDocument/publishDiagnostics` notifications
4. **Feeds into Ralph** via enhanced `GoFeedbackGenerator`
5. **Is opt-in** (default disabled per lsp.yaml; enabled by user choice)
6. **Is language-scoped**: Go-only — does NOT attempt to be a general LSP client

---

## Requirements (EARS Format)

### Subprocess Lifecycle

**REQ-GB-001**: The system SHALL spawn `gopls` as a child process when the user is working in a Go project (detected via `go.mod` in project root).

**REQ-GB-002**: When `gopls` is not found in PATH, the system SHALL log a `slog.Warn` with install hint `go install golang.org/x/tools/gopls@latest` and return `(nil, nil)` from the factory (no error, no client).

**REQ-GB-003**: The gopls subprocess SHALL be started lazily on the first diagnostics request, not eagerly at session start.

**REQ-GB-004**: The system SHALL terminate the gopls subprocess gracefully via LSP `shutdown`/`exit` sequence with a 5-second timeout; on timeout, SIGKILL is sent.

**REQ-GB-005**: When the gopls subprocess crashes mid-session, the system SHALL log the crash and return empty diagnostics for subsequent requests (fail-open, no retry in this SPEC).

### LSP Protocol Handshake

**REQ-GB-010**: The system SHALL send `initialize` request with `rootUri` set to the project root and `ClientCapabilities.textDocument.publishDiagnostics.relatedInformation: true`.

**REQ-GB-011**: The system SHALL send `initialized` notification after receiving the `initialize` response.

**REQ-GB-012**: The initialization SHALL timeout after 30 seconds; on timeout, the subprocess is killed.

**REQ-GB-013**: The system SHALL send `initializationOptions` with `staticcheck: true` to enable staticcheck analyses via gopls.

### Diagnostic Collection

**REQ-GB-020**: When a Go file is opened via `textDocument/didOpen`, the system SHALL wait up to 5 seconds (default) for `textDocument/publishDiagnostics` notifications and return the collected diagnostics.

**REQ-GB-021**: The system SHALL debounce diagnostic notifications for a configurable window (default 150 ms) to capture all related diagnostics before returning.

**REQ-GB-022**: The debounce window, initialization timeout, and request timeout SHALL be configurable via `.moai/config/sections/lsp.yaml` `gopls_bridge.timeouts.*` fields.

**REQ-GB-023**: Each returned `Diagnostic` SHALL include `severity` (error/warning/info/hint), `source` (e.g., `"compiler"`, `"staticcheck"`), `code`, `message`, and `range` fields.

### JSON-RPC Framing

**REQ-GB-030**: The system SHALL frame messages using the LSP `Content-Length` header convention over stdio: `Content-Length: N\r\n\r\n<json>`.

**REQ-GB-031**: The message reader SHALL parse headers until the double `\r\n`, read exactly `N` bytes, and unmarshal the JSON payload.

**REQ-GB-032**: The message writer SHALL serialize the JSON payload, compute `Content-Length`, and write the framed message atomically.

**REQ-GB-033**: Request-response correlation SHALL use the `id` field; a pending requests map correlates incoming responses with callers.

**REQ-GB-034**: Notifications (no `id` field) SHALL be dispatched to a registered handler without affecting pending requests.

### GoFeedbackGenerator Integration

**REQ-GB-040**: `internal/loop/go_feedback.go` `GoFeedbackGenerator.Collect()` SHALL optionally augment its `Feedback` result with gopls diagnostics when the bridge is enabled.

**REQ-GB-041**: The `Feedback` struct SHALL gain a new field `Diagnostics []Diagnostic` to carry LSP-collected diagnostics.

**REQ-GB-042**: Ralph Engine's `ClassifyFeedback` SHALL be updated to inspect `Feedback.Diagnostics` for severity/source-based classification when the slice is non-empty.

### Configuration

**REQ-GB-050**: The system SHALL read gopls bridge settings from `.moai/config/sections/lsp.yaml` `lsp.servers.go` entry (matching the language-neutral template structure).

**REQ-GB-051**: The master switch `lsp.enabled: true` SHALL be required to activate the bridge; when false (default), `GoFeedbackGenerator` falls back to the CLI-only path (current behavior).

**REQ-GB-052**: The bridge SHALL respect `lsp.discovery.on_missing: warn_and_skip` policy when gopls is not installed.

### No External Dependencies

**REQ-GB-060**: The JSON-RPC framing and LSP protocol handling SHALL be implemented without external libraries (no `go.lsp.dev/jsonrpc2`, no `powernap`).

**REQ-GB-061**: The implementation SHALL use only `encoding/json`, `bufio`, `os/exec`, `context`, `sync`, and `log/slog` from the Go standard library.

**REQ-GB-062**: Rationale: (a) go.lsp.dev/jsonrpc2 is pre-v1.0 (last release 2022), (b) powernap ties us to Charm ecosystem, (c) Go-only bridge keeps moai-adk binary dependency-free and small.

---

## Architecture

### Package Layout

```
internal/lsp/gopls/
├── bridge.go        # Bridge type: subprocess lifecycle + public API
├── protocol.go      # JSON-RPC framing (Content-Length reader/writer)
├── messages.go      # LSP message types (initialize, didOpen, publishDiagnostics)
├── handler.go       # Notification handler registry + pending request map
├── config.go        # Config loader from .moai/config/sections/lsp.yaml
└── bridge_test.go, protocol_test.go, handler_test.go
```

### Key Types

```go
// internal/lsp/gopls/bridge.go
type Bridge struct {
    cmd          *exec.Cmd
    stdin        io.WriteCloser
    stdout       io.ReadCloser
    writer       *Writer          // framed JSON-RPC writer
    reader       *Reader          // framed JSON-RPC reader
    nextID       atomic.Int64
    pending      sync.Map         // id → chan *Response
    diagnostics  chan DiagnosticEvent
    shutdown     chan struct{}
    config       *Config
}

func NewBridge(ctx context.Context, projectRoot string, cfg *Config) (*Bridge, error)
func (b *Bridge) GetDiagnostics(ctx context.Context, filePath string) ([]Diagnostic, error)
func (b *Bridge) Close(ctx context.Context) error
```

### Integration Point

```go
// internal/loop/go_feedback.go (updated)
type GoFeedbackGenerator struct {
    projectRoot string
    bridge      *gopls.Bridge  // nil when disabled
}

func (g *GoFeedbackGenerator) Collect(ctx context.Context) (*Feedback, error) {
    fb := &Feedback{
        TestsFailed: g.runGoTest(ctx),
        LintErrors:  g.runGoVet(ctx),
    }
    if g.bridge != nil {
        diags, _ := g.bridge.GetDiagnostics(ctx, g.projectRoot)
        fb.Diagnostics = diags  // new field
    }
    return fb, nil
}
```

---

## Non-Goals

- **Multi-language support**: this SPEC is Go-only by design. Other languages are handled by SPEC-LSP-CORE-002 + SPEC-LSP-MULTI-006.
- **Hover, goto-definition, rename**: only diagnostics collection is in scope.
- **Pull diagnostic model**: this SPEC uses push model (publishDiagnostics notifications) because gopls supports it reliably. SPEC-LSP-CORE-002 can add pull support.
- **gopls embedding as library**: R4 research confirmed gopls internal packages are unstable; we use subprocess.
- **Aggregation across servers**: that is SPEC-LSP-AGG-003's responsibility.

---

## References

- Phase 1 research reports R1 (LSP protocol), R4 (Go ecosystem)
- Phase 2 report C2 (crush manager-coordinator pattern)
- [LSP 3.17 specification](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/)
- [gopls v0.20.0 release notes](https://go.dev/gopls/release/v0.20.0)
- [JSON-RPC 2.0 spec](https://www.jsonrpc.org/specification)
- CLAUDE.local.md Section 22 (Template Language Neutrality) — gopls bridge is scoped to local dev, does NOT become a template default
