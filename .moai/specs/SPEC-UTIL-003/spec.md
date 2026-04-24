---
id: SPEC-UTIL-003
title: "LSP Subprocess Hygiene (stderr drain + singleflight spawn + Diagnostic Alias)"
version: "0.1.0"
status: implemented
created: 2026-04-24
updated: 2026-04-24
author: Wave v2.14 SPEC Writer
priority: P0 Critical
phase: "v2.14.0 — Phase 3 — Utility Hardening"
module: "internal/lsp/core/, internal/hook/, internal/lsp/gopls/"
dependencies: []
related_problem: [IMP-V3U-002, IMP-V3U-003]
related_pattern: []
related_principle: []
related_decision: [D2, D3]
related_theme: "v2.14.0 Utility Hardening"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "lsp, subprocess, stderr, singleflight, diagnostic, v2.14"
---

# SPEC-UTIL-003: LSP Subprocess Hygiene — stderr drain + singleflight spawn + Diagnostic Alias

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.1 | 2026-04-24 | MoAI (release/v2.14.0) | Implementation complete, merged as commit `9dd6a5120`. AC coverage 12/12. `hook.Diagnostic` wire format frozen (IMP-V3U-008 deferred to v2.16). Coverage: lsp/core 92.9%, lsp/gopls 80.1%, hook 85.0%. TRUST 5: 5/5. |
| 0.1.0 | 2026-04-24 | Wave v2.14 SPEC Writer | Initial draft from v2.14.0 release plan §2.2 UTIL-003. Owns IMP-V3U-002 (stderr leak) + IMP-V3U-003 (getOrSpawn race) + D2 scope (internal Diagnostic type alias). |

---

## 1. Goal

Close the two P0 LSP findings from the v2.14 utility audits and land the internal-only portion of the Diagnostic type unification in a single non-breaking release. Specifically:

1. Eliminate the stderr buffer deadlock in `internal/lsp/core/client.go` (IMP-V3U-002) by attaching an `io.Discard` drain goroutine to `result.Stderr` immediately after subprocess supervisor creation.
2. Eliminate the concurrent `getOrSpawn` client race in `internal/lsp/core/manager.go` (IMP-V3U-003) by adding a `singleflight.Group` barrier that guarantees exactly one `clientFactory` invocation per language per in-flight spawn burst.
3. Land the Go type alias for `gopls.Diagnostic` (plus `gopls.Range` and `gopls.Position` cascade) pointing at `lsp.Diagnostic` / `lsp.Range` / `lsp.Position` so that moai has a single source of truth for the LSP bridge path (D2 internal-only scope).

All three changes are internal. No change to `Client` or `Manager` public API surface. No change to hook JSON wire format. No change to `.moai/config/sections/*.yaml` keys. No change to SARIF output. The release semver remains minor.

## 2. Scope

### In-scope

- **S-01 stderr drain**. `internal/lsp/core/client.go` gains a goroutine that runs `io.Copy(io.Discard, result.Stderr)` immediately after `subprocess.NewSupervisor(result)`. Guarded by `if result.Stderr != nil`. Goroutine closes the stderr reader on copy completion.
- **S-02 singleflight spawn**. `internal/lsp/core/manager.go` gains a `sf singleflight.Group` field on `Manager` and refactors `getOrSpawn` to wrap the cache-miss factory+Start path inside `m.sf.Do(language, func() (any, error) { ... })`. Cache insertion is deferred until after `c.Start(ctx)` returns nil.
- **S-03 gopls.Diagnostic alias**. `internal/lsp/gopls/messages.go` replaces the local `type Diagnostic struct { ... }`, `type Range struct { ... }`, and `type Position struct { ... }` declarations with three type alias declarations pointing at `lsp.Diagnostic`, `lsp.Range`, and `lsp.Position` respectively. `DiagnosticSeverity int` is also aliased to `lsp.DiagnosticSeverity`. All existing `gopls.*` usages compile without modification because type aliases have identity semantics.
- **S-04 wire-format freeze test**. A new test in `internal/hook/` locks the hook JSON output bytes before and after the SPEC change to prove no wire format drift.
- **S-05 race detector coverage**. `internal/lsp/core/manager_singleflight_test.go` runs under `-race` to prove the getOrSpawn race is eliminated.
- **S-06 stderr deadlock integration test**. `internal/lsp/core/client_stderr_drain_test.go` writes > 128 KiB to a mock subprocess stderr and asserts the client does not deadlock within a bounded wall clock.

### Out-of-scope (deferred)

- **v2.16 types domain**. `hook.Diagnostic` (string severity) to `lsp.Diagnostic` (int severity) consolidation requires a JSON wire format compatibility projection layer. Deferred to v2.16. Release plan §2.3 records this explicitly as IMP-V3U-008.
- **v3.0 breaking changes**. `config.ResolveClientImpl()` default value flip from `"gopls_bridge"` to `"powernap_core"` (BC-V3R2-UTIL-001) is breaking. Deferred to v3.0.
- **v2.15 subprocess hygiene baseline**. `cmd.WaitDelay`, `Setpgid`, close-all-fds, richer stderr handling (ring buffer, file rotation, structured logging). Cross-platform build tag work deferred to v2.15 IMP-V3U-021.
- **Non-gopls Diagnostic consumers**. The `internal/lsp/hook/` package's `Diagnostic` type remains unchanged. Any future unification of that type must go through the v2.16 wire format compatibility work.

## 3. Environment

### Current code state (scanned 2026-04-24)

- `internal/lsp/core/client.go:213-217`: `readWriteCloser` wraps only `result.Stdout` and `result.Stdin`; `result.Stderr` is returned by the launcher but never consumed.
- `internal/lsp/core/manager.go:332-363`: `getOrSpawn` holds `m.mu` while checking the cache, releases it, inserts the half-ready client at `m.clients[language]`, then calls `c.Start(ctx)` without protection from concurrent callers observing the non-ready client.
- `internal/lsp/models.go:115-130`: `lsp.Diagnostic` canonical type with `Severity DiagnosticSeverity` where `DiagnosticSeverity int` (SeverityError=1, SeverityWarning=2, SeverityInfo=3, SeverityHint=4).
- `internal/lsp/gopls/messages.go:130-170`: `gopls.Diagnostic` structurally identical to `lsp.Diagnostic` but declared as separate type; `gopls.Range`, `gopls.Position`, `gopls.DiagnosticSeverity` similarly duplicated.
- `internal/lsp/hook/types.go:12-62`: `hook.Diagnostic` with string `DiagnosticSeverity` (`"error"`, `"warning"`, ...). Wire format blocker for v2.14 alias — deferred.
- `go.mod:12` + `go.sum:88-89`: `golang.org/x/sync v0.20.0` already present as a direct dependency. singleflight package available without module changes.

### Hook JSON wire format contract

- Hook feedback channel consumer at `internal/hook/post_tool.go:466` `emitToFeedbackChannel` emits payloads containing `LSPDiagnostics: []lsp.Diagnostic` — already int severity. No change in scope.
- Hook teammate_idle path at `internal/hook/teammate_idle.go:180` consumes feedback diagnostics with severity numeric coercion. No change in scope.
- External hook consumers (shell wrappers in `.claude/hooks/moai/`, SARIF exporters, IDE plugins) parse the hook JSON produced by `systemMessage` rendering in `internal/hook/quality/lint_instruction.go`, not the feedback channel struct. The `hook.Diagnostic` string severity path is **untouched** by this SPEC.

### Dependency hardening references

- `.claude/rules/moai/core/lsp-client.md`: powernap v0.1.4 pin (2026-04-22), co-existence with SPEC-GOPLS-BRIDGE-001, quarterly upgrade cadence.
- `.moai/design/utility-review/a3-lsp-audit.md`: P0 classification for both findings, 7/10 audit score with MAINTAIN+targeted verdict.
- `.moai/design/utility-review/a4-external-best-practices.md` §T3 singleflight canonical pattern, §T5 Go subprocess hygiene 2026 baseline.

## 4. Assumptions

- A-01: `golang.org/x/sync v0.20.0` is already a direct dependency (verified in go.mod:12). No module-level change is introduced by importing `golang.org/x/sync/singleflight` from `internal/lsp/core/manager.go`.
- A-02: `powernap v0.1.4` is pinned and its `transport.Connection` does not internally drain subprocess stderr (verified against powernap source via lsp-client.md decision record). The drain responsibility stays at moai's integration boundary.
- A-03: `lsp.Diagnostic`, `lsp.Range`, `lsp.Position`, `lsp.DiagnosticSeverity` field layouts and JSON tags are frozen for v2.14. Any future type unification work is gated on this frozen contract.
- A-04: Hook JSON wire format consumed by external scripts under `.claude/hooks/moai/` uses string severity via `hook.Diagnostic` and is NOT aliased in v2.14. Consumers relying on `"severity": "error"` string values continue to work because `hook.Diagnostic` stays as-is.
- A-05: Integration tests for gopls, pyright, and tsserver can be executed under the `integration` build tag per `.claude/rules/moai/core/lsp-client.md` upgrade policy. CI provides at least one of these language servers.
- A-06: `io.Discard` is zero-allocation per write (Go standard library guarantee since Go 1.16) and imposes no measurable CPU overhead for the maximum observed stderr throughput (~150 KiB/s during pyright large-project startup).

## 5. Requirements (EARS)

### REQ-UTIL-003-001 — stderr drain goroutine

**The system shall** spawn a goroutine running `io.Copy(io.Discard, result.Stderr)` followed by `result.Stderr.Close()` immediately after `subprocess.NewSupervisor(result)` completes successfully in `internal/lsp/core/client.go`.

### REQ-UTIL-003-002 — stderr nil guard

**Where** `result.Stderr` is `nil` (test path with synthesized `LaunchResult` lacking a real subprocess), **the system shall** skip the drain goroutine and proceed to transport construction without error.

### REQ-UTIL-003-003 — singleflight.Group field

**The system shall** declare a `sf singleflight.Group` field on the `Manager` struct in `internal/lsp/core/manager.go`, with zero-value initialization via the existing `NewManager` constructor (no additional options API surface).

### REQ-UTIL-003-004 — getOrSpawn singleflight barrier

**When** `getOrSpawn(ctx, language)` observes a cache miss or a non-reusable existing client, **the system shall** invoke `clientFactory` and `c.Start(ctx)` inside `m.sf.Do(language, factoryFn)` such that concurrent callers for the same language block on the singleflight barrier until the first caller's factory function returns.

### REQ-UTIL-003-005 — exactly-once factory invocation per language per burst

**While** N concurrent `RouteFor` callers target the same language with an initially empty client cache, **the system shall** invoke `clientFactory` exactly once for that language; the remaining N-1 callers shall receive the same `(Client, error)` tuple produced by the first caller's factory execution.

### REQ-UTIL-003-006 — cache insertion gated on successful Start

**The system shall** insert the newly-constructed client into `m.clients[language]` only after `c.Start(ctx)` returns a nil error. **If** `c.Start(ctx)` returns a non-nil error, **then the system shall** leave `m.clients[language]` absent (not present, not set to nil) so that the next `RouteFor` call for the same language executes a fresh factory.

### REQ-UTIL-003-007 — gopls.Diagnostic type alias

**The system shall** declare `type Diagnostic = lsp.Diagnostic` in `internal/lsp/gopls/messages.go`, replacing the former local struct declaration. **The system shall** also declare `type Range = lsp.Range`, `type Position = lsp.Position`, and `type DiagnosticSeverity = lsp.DiagnosticSeverity` to preserve field-type compatibility for existing gopls callers.

### REQ-UTIL-003-008 — hook package Diagnostic unchanged

**The system shall** leave `internal/lsp/hook/types.go` `Diagnostic`, `DiagnosticSeverity`, `Range`, and `Position` declarations unchanged in v2.14. **The system shall not** introduce a type alias pointing `hook.Diagnostic` at `lsp.Diagnostic` because the hook JSON wire format severity is string-typed and the alias would silently flip wire format to int-typed.

### REQ-UTIL-003-009 — hook JSON wire format freeze

**The system shall** produce byte-identical hook JSON output (as emitted by `internal/hook/quality/lint_instruction.go` `FormatDiagnosticsAsInstructionWithFile` and `internal/hook/post_tool.go` `emitToFeedbackChannel`) for any input `[]hook.Diagnostic` slice before and after this SPEC lands. Wire format comparison shall be performed over a canonical fixture of Error / Warning / Information / Hint severities.

### REQ-UTIL-003-010 — stderr deadlock prevention

**When** a subprocess writes more than 128 KiB to stderr in a single burst without any reader draining, **the system shall** complete its next LSP request-response round-trip (e.g., `textDocument/diagnostic`) within a bounded wall clock of 5 seconds without observing a deadlock or `context deadline exceeded` error attributable to stderr back-pressure.

### REQ-UTIL-003-011 — race detector clean on concurrent getOrSpawn

**When** 16 goroutines concurrently invoke `Manager.RouteFor(ctx, path)` for the same language against an initially empty cache, **the system shall** execute under `go test -race ./internal/lsp/core/...` without reporting any data race on `m.clients`, `m.lastActivity`, or the returned client's internal state.

### REQ-UTIL-003-012 — no public API surface change

**The system shall not** add, remove, rename, or change the signature of any exported identifier (type, function, method, constant, variable) in packages `internal/lsp/core/`, `internal/lsp/gopls/`, or `internal/hook/`. The type aliases in `gopls.Diagnostic` / `gopls.Range` / `gopls.Position` / `gopls.DiagnosticSeverity` preserve identifier names and package paths; only the underlying declaration changes from struct/int to alias.

## 6. Acceptance Criteria

Each criterion below is written in Given-When-Then form and maps 1:1 to a REQ above. All criteria are observable via automated tests.

### AC-UTIL-003-001 (REQ-UTIL-003-001)

- **Given** a `subprocess.LaunchResult` with a non-nil `result.Stderr` reader.
- **When** `client.Start(ctx)` runs past `subprocess.NewSupervisor(result)`.
- **Then** a goroutine is observed (via `runtime.NumGoroutine` delta or test-visible instrumentation) that consumes `result.Stderr` into `io.Discard` and closes the reader on copy completion.

### AC-UTIL-003-002 (REQ-UTIL-003-002)

- **Given** a `subprocess.LaunchResult` with `result.Cmd == nil` and `result.Stderr == nil` (the existing test-only launcher path).
- **When** `client.Start(ctx)` runs.
- **Then** no drain goroutine is spawned, no `nil.Close()` panic occurs, and Start proceeds to transport construction normally.

### AC-UTIL-003-003 (REQ-UTIL-003-003)

- **Given** a newly-constructed `Manager` via `NewManager(cfg)` with no additional options.
- **When** the `Manager` struct is inspected (reflection or direct test-scope field read).
- **Then** the `sf singleflight.Group` field is present, zero-valued, and ready to accept `Do` calls.

### AC-UTIL-003-004 (REQ-UTIL-003-004)

- **Given** a `Manager` with an empty `clients` cache and a valid `servers["go"]` config.
- **When** two goroutines simultaneously call `getOrSpawn(ctx, "go")`.
- **Then** the second caller observably blocks on `sf.Do` until the first caller's factory function returns, as proven by wall-clock timing or by a test-visible counter on factory invocation.

### AC-UTIL-003-005 (REQ-UTIL-003-005)

- **Given** a `Manager` with an empty `clients` cache and a `clientFactory` wrapped to increment an atomic counter on each invocation.
- **When** 16 goroutines concurrently call `RouteFor(ctx, "file.go")` for the same language.
- **Then** the atomic counter shows exactly 1 factory invocation, and all 16 callers receive the same non-nil `Client` pointer and a nil error.

### AC-UTIL-003-006 (REQ-UTIL-003-006)

- **Given** a `Manager` with an empty `clients` cache and a `clientFactory` returning a client whose `Start(ctx)` returns a non-nil error.
- **When** `getOrSpawn(ctx, "go")` is called once.
- **Then** the function returns the Start error, `m.clients["go"]` is absent (`_, ok := m.clients["go"]; ok == false`), and a subsequent `getOrSpawn(ctx, "go")` call invokes the factory again from scratch.

### AC-UTIL-003-007 (REQ-UTIL-003-007)

- **Given** the refactored `internal/lsp/gopls/messages.go` with type alias declarations.
- **When** `go build ./internal/lsp/gopls/...` and `go build ./internal/...` run.
- **Then** the build succeeds with zero errors, `reflect.TypeOf(gopls.Diagnostic{}) == reflect.TypeOf(lsp.Diagnostic{})` evaluates to `true`, and `gopls.SeverityError == lsp.SeverityError` evaluates to `true`.

### AC-UTIL-003-008 (REQ-UTIL-003-008)

- **Given** the `internal/lsp/hook/types.go` file after this SPEC lands.
- **When** the file is diffed against the pre-SPEC baseline.
- **Then** no changes are present in the `Diagnostic`, `DiagnosticSeverity`, `Position`, or `Range` declarations, and `hook.DiagnosticSeverity("error")` continues to JSON-marshal to the string `"error"`.

### AC-UTIL-003-009 (REQ-UTIL-003-009)

- **Given** a canonical fixture `[]hook.Diagnostic` with one Error, one Warning, one Information, one Hint.
- **When** the fixture is rendered through `FormatDiagnosticsAsInstructionWithFile` and through `emitToFeedbackChannel` (with the feedback channel captured in-memory) before and after this SPEC lands.
- **Then** the produced byte strings are identical under `bytes.Equal` for both rendering paths.

### AC-UTIL-003-010 (REQ-UTIL-003-010)

- **Given** a mock subprocess that writes 128 KiB of bytes to its stderr before responding to any LSP JSON-RPC request.
- **When** a `core.Client` is constructed against the mock and `c.Start(ctx)` followed by `c.GetDiagnostics(ctx, "/fake/file.go")` runs.
- **Then** both calls return within 5 seconds of wall clock time with no `context deadline exceeded` error, confirming the drain goroutine prevents stderr buffer deadlock.

### AC-UTIL-003-011 (REQ-UTIL-003-011)

- **Given** the new `manager_singleflight_test.go` test file.
- **When** `go test -race -run TestGetOrSpawnConcurrent ./internal/lsp/core/` runs.
- **Then** the test passes with zero race detector warnings on `m.clients`, `m.lastActivity`, and on any internal field of the returned `Client`.

### AC-UTIL-003-012 (REQ-UTIL-003-012)

- **Given** the pre-SPEC exported-symbol surface of `internal/lsp/core/`, `internal/lsp/gopls/`, and `internal/hook/` packages (captured via `go doc` or `apigen` snapshot).
- **When** the post-SPEC surface is captured via the same tool.
- **Then** the two surfaces are byte-identical (same type names, same function signatures, same exported constants).

## 7. Constraints

- **C-01 Wire format immutability**. Hook JSON wire format output for any `[]hook.Diagnostic` input MUST be byte-identical before and after this SPEC. This is verified by REQ-UTIL-003-009 and AC-UTIL-003-009.
- **C-02 Race detector cleanliness**. `go test -race ./internal/lsp/...` and `go test -race ./internal/hook/...` MUST report zero warnings after this SPEC lands. This is verified by REQ-UTIL-003-011 and the standard test suite run in the v2.14 Phase 5 integration gate.
- **C-03 No new public API**. No exported identifier is added, removed, renamed, or has its signature changed. Type aliases preserve identifier names by construction (REQ-UTIL-003-012).
- **C-04 No new module dependency**. `golang.org/x/sync` is already a direct dependency; no new `go.mod require` line is added. The SPEC change touches `go.sum` only if the toolchain decides to re-verify (no hash change expected).
- **C-05 LSP integration coverage**. Per `.claude/rules/moai/core/lsp-client.md` upgrade policy the integration test suite MUST pass for at least one language server available on the CI runner (gopls, pyright, or tsserver). Failing integration tests block merge.
- **C-06 Shutdown safety**. The stderr drain goroutine MUST NOT block or be blocked by `Manager.Shutdown(ctx)`. Drain goroutine ownership is tied to the stderr pipe lifetime; Supervisor kill paths close stderr which causes `io.Copy` to return naturally.
- **C-07 Language neutrality**. The fix path treats Go, Python, TypeScript, Rust, and all 16 supported languages identically. No language-specific code is introduced; the stderr drain and singleflight barrier apply uniformly.

## 8. Risks & Mitigations

| Risk ID | Risk | Mitigation | Verification |
|---------|------|------------|--------------|
| RK-01 | hook.Diagnostic alias would silently flip JSON wire format severity from string to int | Scope explicitly excludes hook.Diagnostic; only gopls.Diagnostic is aliased in v2.14 | REQ-UTIL-003-008, REQ-UTIL-003-009, AC-UTIL-003-008, AC-UTIL-003-009 |
| RK-02 | singleflight blocks waiting caller until factory completes (cold-start latency ~300 ms) | Accepted trade-off; waiting callers previously received a half-ready client (correctness defect). Block time is bounded by Start timeout. | REQ-UTIL-003-004 (semantics documented) |
| RK-03 | io.Discard drain loses gopls debugging insight | Documented trade-off per D3 decision. v2.15+ may add ring buffer / file rotation. Operators can still use gopls CLI serverside log flags. | Release plan §0.5 D3 recorded |
| RK-04 | Existing tests may depend on duplicate factory invocation (race-condition-as-feature) | REQ-UTIL-003-005 explicit exactly-once semantics; test updates identified and tracked in plan.md | AC-UTIL-003-005 |
| RK-05 | gopls.Range / gopls.Position cascade required for gopls.Diagnostic alias | Alias all three (Range, Position, Diagnostic) + DiagnosticSeverity simultaneously; four one-line declarations preserve identity semantics for all callers | REQ-UTIL-003-007, AC-UTIL-003-007 |
| RK-06 | singleflight does not propagate context cancellation to waiters | Factory body respects ctx; cancelled waiters observe cancellation on next retry (acceptable for spawn path which is infrequent) | Documented in research.md §4.1 |
| RK-07 | Drain goroutine leak if subprocess never closes stderr | Supervisor.Kill() transitively closes the stderr pipe on shutdown; io.Copy returns and the goroutine exits naturally | AC-UTIL-003-001 (instrumentation-based verification) |

## 9. Dependencies

### Sibling SPECs (v2.14.0 Phase 3 parallel track)

- **SPEC-UTIL-001** — MX Validator Correctness + 16-Language Complexity. Parallel; independent module scope (`internal/hook/mx/`).
- **SPEC-UTIL-002** — ast-grep Integration Hardening + 5-Language Rule Seeding. Parallel; independent module scope (`internal/astgrep/`, `internal/hook/security/`).

No file-level overlap with UTIL-001 or UTIL-002 — parallel execution per release plan §5 Phase 3 gate.

### External dependencies

- `golang.org/x/sync/singleflight` (via `golang.org/x/sync v0.20.0`, already pinned in go.mod).
- `github.com/charmbracelet/x/powernap v0.1.4` (already pinned; drain pattern ships at moai's integration layer, not powernap's).
- Go standard library `io.Discard` (zero-allocation sink, stable since Go 1.16).

### Rule-level references

- `.claude/rules/moai/core/lsp-client.md` — powernap pin, upgrade policy, integration test matrix.
- `.claude/rules/moai/core/moai-constitution.md` — TRUST 5 quality gates (Tested, Readable, Unified, Secured, Trackable).
- `.claude/rules/moai/workflow/spec-workflow.md` — Plan-Run-Sync phase contract.

### Deferred to future releases

- **v2.15** — IMP-V3U-021 subprocess hygiene baseline (WaitDelay, Setpgid, close-fds, richer stderr handling).
- **v2.16** — IMP-V3U-008 hook.Diagnostic field-level consolidation with JSON wire format compatibility projection.
- **v3.0** — BC-V3R2-UTIL-001 `config.ResolveClientImpl()` default flip to `"powernap_core"`.

## 10. Traceability

### Requirement → Acceptance → Source

| REQ ID | AC ID | Source file(s) modified | Audit citation |
|--------|-------|--------------------------|----------------|
| REQ-UTIL-003-001 | AC-UTIL-003-001 | internal/lsp/core/client.go:~213 (post-Supervisor spawn) | IMP-V3U-002, SYNTHESIS §3.1 D2-3, a3-lsp-audit §D2-3 |
| REQ-UTIL-003-002 | AC-UTIL-003-002 | internal/lsp/core/client.go:~213 (nil guard) | IMP-V3U-002 (test-path safety) |
| REQ-UTIL-003-003 | AC-UTIL-003-003 | internal/lsp/core/manager.go:~40 (struct field) | IMP-V3U-003, SYNTHESIS §3.1 D10-1 |
| REQ-UTIL-003-004 | AC-UTIL-003-004 | internal/lsp/core/manager.go:332-363 (refactor) | IMP-V3U-003, a3-lsp-audit §D10-1 |
| REQ-UTIL-003-005 | AC-UTIL-003-005 | internal/lsp/core/manager.go:332-363 | IMP-V3U-003 (exactly-once semantics) |
| REQ-UTIL-003-006 | AC-UTIL-003-006 | internal/lsp/core/manager.go:332-363 | IMP-V3U-003 (cache-gating) |
| REQ-UTIL-003-007 | AC-UTIL-003-007 | internal/lsp/gopls/messages.go:130-170 (alias declarations) | D2 decision (release plan §0.5) |
| REQ-UTIL-003-008 | AC-UTIL-003-008 | internal/lsp/hook/types.go (unchanged) | D2 scope restriction (wire format freeze) |
| REQ-UTIL-003-009 | AC-UTIL-003-009 | new test internal/hook/wire_format_freeze_test.go | C-01, "breaking: false" gate |
| REQ-UTIL-003-010 | AC-UTIL-003-010 | new test internal/lsp/core/client_stderr_drain_test.go | IMP-V3U-002 deadlock proof |
| REQ-UTIL-003-011 | AC-UTIL-003-011 | new test internal/lsp/core/manager_singleflight_test.go | IMP-V3U-003 race detector |
| REQ-UTIL-003-012 | AC-UTIL-003-012 | new test internal/public_surface_test.go (or equivalent apigen) | C-03, semver-minor gate |

### Decision record mapping

- D2 (Diagnostic unification internal-only + Go type alias) → S-03, REQ-UTIL-003-007, REQ-UTIL-003-008.
- D3 (LSP subprocess stderr io.Discard drain) → S-01, REQ-UTIL-003-001, REQ-UTIL-003-002.

### Audit citation mapping

- A3 audit §D2-3 (stderr leak) → REQ-UTIL-003-001, REQ-UTIL-003-002, REQ-UTIL-003-010.
- A3 audit §D10-1 (getOrSpawn race) → REQ-UTIL-003-003, REQ-UTIL-003-004, REQ-UTIL-003-005, REQ-UTIL-003-006, REQ-UTIL-003-011.
- SYNTHESIS IMP-V3U-002 → REQ-UTIL-003-001, REQ-UTIL-003-002, REQ-UTIL-003-010.
- SYNTHESIS IMP-V3U-003 → REQ-UTIL-003-003..006, REQ-UTIL-003-011.

---

Version: 0.1.0
Status: draft
Author: Wave v2.14 SPEC Writer
Last Updated: 2026-04-24
