# A3 — LSP Subsystem Audit

## Executive Summary

- **LOC**: 18,718 (9 sub-packages, 73 Go source files + 26 test files)
- **Package map**: `lsp` (models), `transport`, `core`, `subprocess`, `gopls`, `hook`, `aggregator`, `cache`, `config`
- **powernap pin**: `github.com/charmbracelet/x/powernap v0.1.4`
- **Overall health**: **7/10**
- **All unit tests**: PASS (`go test -race ./internal/lsp/...` — clean)
- **go vet**: PASS (exit 0)

### Top 5 Issues

1. **[BUG] `getOrSpawn` concurrent client spawn** — two goroutines racing on the same language insert a half-initialized client (`StateSpawning`) into the cache before `Start()` returns; second caller receives that not-yet-ready client.
2. **[ARCH] `gopls/bridge.go` is an independent JSON-RPC stack** — 632-LOC hand-rolled Content-Length framing, PendingRegistry, NotificationDispatcher that fully duplicates the `transport` package; lsp-client.md selected powernap as the single wire layer but gopls bypasses it entirely.
3. **[LEAK] `LaunchResult.Stderr` is never consumed** — `client.Start()` builds `readWriteCloser` from only `Stdin`/`Stdout`; the subprocess stderr pipe is left open and unread, which can block the subprocess when its stderr buffer fills.
4. **[DUP] Three parallel diagnostic type systems** — `lsp.Diagnostic` (models), `hook.Diagnostic` (string severity), `gopls.Diagnostic` (in protocol.go); `loop/state.go` and `ralph/engine.go` still import `gopls.Diagnostic` directly.
5. **[DUP] Phase type defined twice** — `hook.PhaseType` and `core/quality.WorkflowPhase` are identical string constants (`"plan"`, `"run"`, `"sync"`) defined in separate packages with no cross-reference.

### Top 3 Critical Fixes

1. **Protect concurrent client spawn in `getOrSpawn`** with a per-language `sync.Once` or singleflight pattern.
2. **Drain stderr** in `client.Start()` with a background goroutine (`io.Copy(io.Discard, result.Stderr)`).
3. **Unify diagnostic types** — retire `hook.Diagnostic` and `gopls.Diagnostic`; canonicalize on `lsp.Diagnostic` from `models.go`.

### Verdict

**MAINTAIN with targeted fixes** — the powernap-based `core` path is architecturally sound and race-clean. The gopls bridge is a parallel universe that should be soft-deprecated in v3. The three critical bugs above can be fixed without a full rewrite.

---

## Package Dependency Graph (ASCII)

```
cmd / CLI
    │
    ├──► aggregator ──────────────────────────► cache
    │        │                                    │
    │        └──► core.Manager ◄── resilience     │
    │                  │                          │
    │        ┌─────────┘                          │
    │        ▼                                    │
    │    core.Client ──► transport (powernap) ◄───┘
    │        │               │
    │        │           subprocess.Launcher
    │        │           subprocess.Supervisor
    │        │
    │    core.{state,document,capabilities,queries,errors}
    │
    ├──► hook.{gate,tracker,fallback,diagnostics}
    │        │
    │        └──► resilience.CircuitBreaker
    │
    ├──► gopls.Bridge (INDEPENDENT JSON-RPC STACK)
    │        │
    │        └──► gopls.{Writer,Reader,PendingRegistry,
    │                     NotificationDispatcher,protocol,uri,config}
    │
    ├──► config.{loader,types}
    │
    └──► lsp.models (root package — shared types)
```

**Anti-pattern**: `gopls` package re-implements everything in `transport` without using powernap. `ralph/engine.go` and `loop/go_feedback.go` import `gopls.Diagnostic` directly, creating a second diagnostic type pathway.

---

## D1 — Transport Layer Correctness

### Powernap Integration

`transport/transport.go` wraps `pntr.NewConnection` (powernap) behind a `rpcConn` interface. The anti-pattern rule from `lsp-client.md` — "do NOT add sourcegraph/jsonrpc2 as a direct dependency" — is respected: the only powernap import is `github.com/charmbracelet/x/powernap/pkg/transport` at `transport.go:10`.

**Finding D1-1 (Low)**: `NewPowernapTransport` creates a `context.Background()` context for `pntr.NewConnection`. If the parent context is cancelled before the subprocess dies, the internal connection goroutine in powernap cannot be notified through this background context. The connection relies entirely on `stream.Close()` for teardown. This is acceptable but should be documented.

**Finding D1-2 (Low)**: `errorTransport.Close()` always returns `nil` even when the construction error was catastrophic. Callers can call `Close()` on a broken transport without knowing the root cause.

### Message Framing

All Content-Length framing is delegated to powernap. The `core` path has no manual framing code — clean by design.

**Concurrent pending-request management**: Delegated entirely to powernap. No manual in-flight map in `core`. Clean.

### Error Propagation

`CallWithTimeout` (transport/request.go) correctly distinguishes context timeout (`ErrRequestTimeout` sentinel) from protocol errors, allowing callers to branch. `WrapCallError` preserves the `%w` chain for `errors.Is/As`.

**Finding D1-3 (Medium)**: `isContextErr` checks `ctx.Err() != nil` first, then `errors.Is`. If the transport itself returns `context.DeadlineExceeded` before the context has expired (e.g., powernap internal timeout), the error is still wrapped as `ErrRequestTimeout`. This can cause false "timeout" classification for transport-layer errors.

---

## D2 — Client Lifecycle

### State Machine

`StateMachine` in `core/state.go` is clean: `sync.RWMutex`-guarded, explicit transition table, `ErrInvalidTransition` sentinel. Allowed transitions: `spawning → initializing → ready → degraded → shutdown`.

**Finding D2-1 (Medium — BUG)**: `client.Start()` transitions to `StateShutdown` on `ErrBinaryNotFound` (line 200). If `Start()` is retried (e.g., after binary install), `StateShutdown` is terminal and `Transition` will always fail — the client becomes permanently unusable. Callers must construct a new client after a binary-not-found failure, but this is not documented on the `Client.Start()` godoc.

### Initialize Handshake

`initialize()` (client.go:251) correctly sends `initialized` notification after the `initialize` response. `rootUri` and `workspaceFolders` are both populated when `cfg.RootDir != ""`. Comment notes gopls trusts `workspaceFolders` more reliably — good.

**Finding D2-2 (Low)**: Capability parse failure in `initialize()` degrades to empty capabilities with a Warn log, but state is still transitioned to `StateReady`. A server that sends malformed capabilities will appear ready but fail every capability check silently.

### Document Open/Close Tracking

`documentCache` (document.go) is correct: `openOrChange` sends `didOpen` first time, `didChange` on content change, no-op on same content. `reapIdle` sends `didClose` for inactive files.

**Finding D2-3 (Critical — BUG) — Stderr Leak**: `client.Start()` at line 213-217 builds `readWriteCloser{r: result.Stdout, w: result.Stdin}`. `result.Stderr` is assigned in `LaunchResult` (launcher.go:150) but never consumed. On Linux/macOS, when a subprocess writes more than 64KB to stderr (e.g., gopls verbose logging), the pipe buffer fills and the subprocess blocks on its next write, deadlocking the entire LSP client.

**Recommendation**: Add `go io.Copy(io.Discard, result.Stderr)` immediately after creating the supervisor in `client.Start()`. Use `subprocess.Supervisor` to close `Stderr` on `doneCh`.

---

## D3 — Multi-Language Support

### Supported Languages

`aggregator/aggregator.go:244-290` maps 16 extension groups to language identifiers: go, typescript, javascript, python, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, swift. This matches the FROZEN 16-language requirement.

`fallback.go:registerDefaultTools()` covers only 5 languages: python (ruff), typescript (tsc), javascript (eslint), go (go vet), rust (cargo clippy). Languages 6-16 have no fallback. This is acceptable per REQ-HOOK-162 ("no tool configured" → `ErrDiagnosticsUnavailable`).

### Graceful Degradation

`Manager.getOrSpawn` + `ErrBinaryNotFound` → warn_and_skip: clean. `Aggregator` circuit-breaker graceful-degradation: clean.

**Finding D3-1 (Medium)**: `detectLanguage` in `manager.go` and `detectLanguageFromPath` in `aggregator.go` are duplicate extension-to-language maps. When a new language is added, both must be updated. They have drifted slightly: `aggregator` maps `.dart` to `"flutter"` (correct FROZEN name); manager relies on `config.ServerConfig.FileExtensions` (data-driven). The `aggregator` copy is hardcoded and will diverge from `lsp.yaml` config.

**Recommendation**: Replace `aggregator.detectLanguageFromPath` with a call to `Manager.detectLanguage` via the `Router` interface, or expose a package-level `LanguageForExt(ext string) string` function from `config`.

### Feature Flag

`config.ServersConfig.ResolveClientImpl()` returns `"gopls_bridge"` as default. SPEC-LSP-CORE-002 selected `powernap_core` as the primary path, but the default is still the legacy bridge. This is an intentional coexistence configuration per lsp-client.md, but operationally means all deployments use the legacy bridge unless explicitly configured.

---

## D4 — Subprocess Management

### Launcher

`subprocess/launcher.go:Launch()` cleanly: resolves binary (primary + fallbacks), creates isolated stdin/stdout/stderr pipes, starts process. `ErrBinaryNotFound` + `InstallHintError` sentinel chain supports `errors.Is/As`.

### Supervisor

`supervisor.go:NewSupervisor` starts a goroutine that calls `cmd.Wait()` exactly once, stores `ExitEvent`, and closes `doneCh`. `Watch(ctx)` creates an independent per-caller channel — multiple watchers are safe.

**Finding D4-1 (High)**: `Supervisor.Kill()` uses `syscall.SIGKILL` directly. On Windows, `syscall.SIGKILL` is not defined in the same way as on POSIX systems. The `cmd.Process.Signal(syscall.SIGKILL)` call may be `SIGKILL = Signal(9)` via Go's `os.Kill` alias, but the `supervisor.go` comment says "SIGKILL" explicitly while the `Kill()` method calls `cmd.Process.Signal(syscall.SIGKILL)`. Go's `os.Process.Kill()` is cross-platform; using it instead of `syscall.SIGKILL` directly would be more portable.

**Recommendation**: Replace `syscall.SIGKILL` with `s.cmd.Process.Kill()` (uses `os.Process.Kill` which is cross-platform). Similarly in `client.Shutdown`, `syscall.SIGTERM` can be replaced with `os.Interrupt` for better Windows compatibility.

**Finding D4-2 (Medium)**: `Supervisor.Watch` creates a goroutine per call. If `Watch` is called many times without consuming the returned channel, goroutine count grows. `Watch` is used in `client.Shutdown` (single call), so this is low-risk today but is an API footgun.

---

## D5 — Hook Integration

### Quality Gate Phases

`gate.go:EnforcePhase` correctly implements the three-phase model: plan (baseline capture), run (error/type/lint thresholds + regression), sync (max_errors/max_warnings/require_clean_lsp).

**Finding D5-1 (Medium — Architecture)**: `parseQualityConfig` (gate.go:103-123) builds a flat `QualityGate` struct mixing Run `MaxErrors` with Sync `MaxWarnings` as a generic gate. This legacy path is used by `LoadConfig`/`CheckWithConfig`, which are still called from external hooks. Meanwhile `EnforcePhase` uses the full `loadModelsConfig()` path. Two parallel config-loading paths with different semantics exist in the same file.

**Finding D5-2 (Low — Bug)**: `DetectRegression` (gate.go:329-335):
```go
errorIncrease := current.Errors - baseline.Errors
if errorIncrease > errorThreshold {
```
When `current.Errors < baseline.Errors`, `errorIncrease` is negative. With default `errorThreshold = 0`, a negative value is not `> 0`, so improvement (fewer errors) correctly does not trigger regression. However, if the threshold is deliberately set to `-1` (intentionally allowing one error reduction), the integer subtraction could yield unexpected results. The function should document that negative `errorIncrease` is always safe.

### Regression Tracker

`tracker.go:GetBaseline` acquires `RLock` then calls `loadBaselineLocked()` which reads from disk. Disk reads under an RLock are acceptable only if `loadBaselineLocked` is idempotent for concurrent readers — it is (nil check then read). Safe.

**Finding D5-3 (Medium)**: `regressionTracker` stores the full diagnostic slice per file in the baseline JSON. A project with 500 files and 100 diagnostics each = 50,000 diagnostic structs serialized to disk on every `SaveBaseline` call. No pruning, compaction, or count-only option exists.

### Fallback Diagnostics

`fallback.go:execTool` at line 258 applies a **hardcoded 30-second timeout** regardless of the caller's context, duplicating/overriding the context timeout. If the caller passed a 5-second context, the tool can still run for 30 seconds.

**Recommendation**: Remove the `context.WithTimeout(ctx, 30*time.Second)` override; respect the caller's deadline directly. If a default is needed, apply it only when `ctx` has no deadline (`context.Background()`).

---

## D6 — Aggregator

### Architecture

`aggregator.go` correctly uses: `singleflight` for deduplication, per-language `resilience.CircuitBreaker` (lazy create with double-checked locking), TTL cache with `GetStale` fallback, per-query context timeout.

**Finding D6-1 (Low)**: `GetDiagnostics` applies `queryTimeout` on every call:
```go
qCtx, cancel := context.WithTimeout(ctx, a.queryTimeout)
```
This always creates a new context and goroutine infrastructure even for cache hits. The cache check at line 145 happens after the timeout context is created. Moving the cache check before `context.WithTimeout` would eliminate the overhead for hot paths.

**Finding D6-2 (Medium)**: `detectLanguageFromPath` (aggregator.go:244-290) is a 46-line hand-rolled extension parser that reimplements `filepath.Ext` plus a switch table. The existing `filepath.Ext` call could simplify this to ~10 lines:
```go
func detectLanguageFromPath(path string) string {
    switch filepath.Ext(path) {
    case ".go": return "go"
    ...
    }
}
```
The current implementation manually scans backwards for the last `.`, which is functionally equivalent to `filepath.Ext` but harder to read and maintain.

### Multi-client Deduplication

No issue with multi-language deduplication across clients — `singleflight` is keyed by URI (path), which is language-scoped by the router.

---

## D7 — gopls Bridge

### Architecture Assessment

`gopls/bridge.go` (632 LOC) is a complete, independent JSON-RPC stack:

| Component | gopls bridge | powernap core |
|-----------|-------------|---------------|
| Framing | `Writer`/`Reader` with Content-Length | powernap |
| Request tracking | `PendingRegistry` (channel map) | powernap |
| Notification dispatch | `NotificationDispatcher` | `NotificationRouter` |
| Circuit breaker | custom `cbMu`/`cbFailures` inline | `resilience.CircuitBreaker` |
| Subprocess mgmt | inline `cmd`, `stdin`, `stdout` | `subprocess.Launcher/Supervisor` |
| Retry logic | none | none |

This duplication violates lsp-client.md's intent: "sourcegraph/jsonrpc2: do NOT add as a direct dependency; use powernap's exported API only." The bridge avoids that specific anti-pattern (it implements its own framing rather than importing jsonrpc2 directly), but it nonetheless reimplements everything powernap provides.

**Finding D7-1 (High — Architecture)**: The bridge's inline circuit breaker (`cbMu/cbFailures/cbOpenUntil`) duplicates `resilience.CircuitBreaker`. When `resilience.CircuitBreaker` evolves (e.g., half-open retry improvements in #679), the gopls bridge's copy stays stale.

**Finding D7-2 (Critical — Semantic)**: `gopls/bridge.go:GetDiagnostics` sends `textDocument/didOpen` with `LanguageID: "go"` hardcoded (line 259):
```go
LanguageID: "go",
```
This is correct for gopls-only usage, but it signals the gopls bridge can never generalize. Any attempt to reuse this for TypeScript would need a different path, reinforcing duplication.

**Finding D7-3 (High)**: `gopls/bridge.go:Close` spawns a `go func() { done <- b.cmd.Wait() }()` at line 479. This goroutine leaks if `Bridge.Close` is called multiple times (e.g., panic recovery). `closeOnce` protects channel close but not the `cmd.Wait` goroutine.

### Should gopls Bridge Be Retired?

Per lsp-client.md: "Users select via `lsp.client` config key." The default is `gopls_bridge`. `internal/ralph/engine.go` and `internal/loop/` still import `gopls.Diagnostic` directly. Full retirement requires:

1. Migrating `ralph/engine.go` to use `lsp.Diagnostic` via aggregator
2. Migrating `loop/` to use the core Client path
3. Changing the default `ClientImpl` to `powernap_core`

**Recommendation for v3**: Soft-deprecate gopls bridge — add a `@MX:WARN DEPRECATED` on `NewBridge`. Plan removal after all callers are migrated to `aggregator.GetDiagnostics`.

---

## D8 — Testing Quality

### Coverage

| Package | Coverage |
|---------|----------|
| `lsp` (models) | 100.0% |
| `config` | 100.0% |
| `transport` | 97.0% |
| `aggregator` | 96.8% |
| `core` | 92.9% |
| `subprocess` | 89.5% |
| `hook` | 89.5% |
| `cache` | 87.9% |
| `gopls` | **80.6%** |

All packages exceed the 85% TRUST5 threshold except `gopls` (80.6%). The uncovered 19.4% includes error paths in `collectDiagnostics` and `readLoop`.

### Race Detector

`go test -race ./internal/lsp/...` — **CLEAN** (all pass). This is the most important health signal.

### Integration Tests

Integration tests use `//go:build integration` tag and `t.Skipf("...")` when binaries are missing — correct CI guard pattern. Tests are tagged for: Go (gopls), Python (pyright/pylsp), TypeScript (tsserver).

**Finding D8-1 (Medium)**: `gopls` package coverage at 80.6% is below the project threshold. The uncovered paths include: `collectDiagnostics` timeout branch when no events arrive, `readLoop` behavior after `b.shutdownCh` closes during active read, and `sendShutdown` response channel close path. These are the highest-risk paths in the async bridge.

**Finding D8-2 (Low)**: `tracker.go:compareDignostics` function name has a typo: `compareDignostics` (missing 'a'). This is exported-function adjacent (called via public interface) and creates confusion in test assertions.

---

## D9 — Documentation / Architecture

### lsp-client.md Alignment

lsp-client.md states the only powernap import should be via `powernap/pkg/transport.Connection`. The actual code imports `pntr "github.com/charmbracelet/x/powernap/pkg/transport"` at `transport.go:10` and nothing else from powernap — **aligned**.

lsp-client.md states gopls-bridge path remains opt-in via `lsp.client: gopls-bridge`. The config `ResolveClientImpl()` defaults to `"gopls_bridge"` — **misaligned**. The default should match lsp-client.md's intent that `lsp-core` (powernap) is the default.

### Per-package doc.go

Present in: `transport/doc.go`, `subprocess/doc.go`, `aggregator/doc.go`, `cache/doc.go`, `config/doc.go`, `core/doc.go`. Missing: **`hook/` package** has no `doc.go`. Given hook is the integration point with external hooks, this is a gap.

### Architecture Diagram

No architecture diagram exists in source. The dependency graph in this audit (D0) should be promoted to a `ARCHITECTURE.md` in `internal/lsp/`.

---

## D10 — Performance

### Diagnostics Polling

The core client uses LSP push model (publishDiagnostics). No polling — correct.

The gopls bridge uses a debounce-based collection window (`DebounceWindow` from config). Default not visible in this review, but the pattern is correct for LSP push.

### Server Startup Latency

`client.Start()` blocks the calling goroutine during subprocess launch + LSP initialize handshake. With gopls, this typically takes 200-800ms. `Manager.getOrSpawn` holds no lock during `Start()` — correct.

**Finding D10-1 (Medium)**: Because the client is inserted into the cache before `Start()` completes (see D1 concurrent spawn bug), a second concurrent caller for the same language will receive the in-progress client (State=Spawning) immediately. If that caller calls `OpenFile` before `Start()` completes (transport is nil), it will panic or receive a nil pointer dereference — `c.tr` is nil until `c.trFactory(stream)` is called.

**Finding D10-2 (Low)**: `DiagnosticCache` cleanup goroutine (cache/cache.go) runs every 10 seconds by default. For short-lived CLI invocations (< 10s), the cleanup goroutine may never fire. The cache should support `context.WithCancel`-based shutdown for clean CLI exit — this is provided via `cache.Stop()` but callers must remember to call it.

### Memory Footprint

`documentCache` stores full document text (`content string`). For large files (> 1MB), this doubles memory per open document. LSP servers typically only need incremental updates after the first `didOpen`, but the current implementation re-sends full text on every `didChange`. This is spec-compliant but memory-inefficient for large workspaces.

---

## Architectural Recommendations

**R1 [Critical]: Fix concurrent `getOrSpawn` spawn race**

Use per-language `singleflight.Group` or a separate "in-progress" map:

```go
// Replace the naive insert-before-Start pattern with singleflight
var sf singleflight.Group
// in getOrSpawn:
v, err, _ := m.sf.Do(language, func() (any, error) {
    c := m.clientFactory(sc)
    if err := c.Start(ctx); err != nil {
        return nil, err
    }
    m.mu.Lock()
    m.clients[language] = c
    m.mu.Unlock()
    return c, nil
})
```

This ensures exactly one `Start()` runs per language at a time, and all waiters get the same ready client.

**R2 [Critical]: Drain subprocess stderr**

In `client.Start()` after `NewSupervisor`:

```go
if result.Stderr != nil {
    go func() {
        io.Copy(io.Discard, result.Stderr)
        result.Stderr.Close()
    }()
}
```

**R3 [High]: Unify diagnostic types on `lsp.Diagnostic`**

Three parallel types (`lsp.Diagnostic`, `hook.Diagnostic`, `gopls.Diagnostic`) create serialization friction and mismatched severity constants (int vs string). Migrate:

1. `hook/types.go`: Replace `hook.Diagnostic` with `lsp.Diagnostic`; replace `hook.DiagnosticSeverity` (string) with `lsp.DiagnosticSeverity` (int).
2. `gopls/protocol.go`: `gopls.Diagnostic` → use `lsp.Diagnostic` directly.
3. Update `ralph/engine.go` and `loop/go_feedback.go` to import `lsp.Diagnostic`.

**R4 [High]: Soft-deprecate `gopls/bridge.go`**

Add deprecation marker and route `internal/loop/` and `internal/ralph/` through `Aggregator.GetDiagnostics`. Change `ResolveClientImpl()` default to `"powernap_core"`.

**R5 [Medium]: Unify phase type constants**

Merge `hook.PhaseType` and `core/quality.WorkflowPhase` into a single definition in `internal/lsp/hook/types.go` or a new `internal/lsp/phase` package. Both are string types with identical values.

**R6 [Medium]: Replace `detectLanguageFromPath` duplication**

Move the canonical extension-to-language map to `config` package as `config.LanguageForExt(ext string) string`. Both `manager.go` and `aggregator.go` should call it.

**R7 [Medium]: Fix `execTool` 30-second hardcoded timeout**

Replace:
```go
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
```
With:
```go
if _, ok := ctx.Deadline(); !ok {
    var cancel context.CancelFunc
    ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
}
```

**R8 [Low]: Cross-platform SIGKILL**

Replace `syscall.SIGKILL` in `supervisor.Kill()` with `s.cmd.Process.Kill()`.

---

## Proposed Package Restructure (v3)

No major restructure recommended. The `core` → `transport` → `subprocess` dependency chain is clean. Suggested targeted consolidations:

```
internal/lsp/
├── models.go          (canonical types — no change)
├── phase/             (NEW: extract PhaseType from hook, WorkflowPhase from quality)
├── config/            (add LanguageForExt helper)
├── transport/         (no change — powernap wrapper)
├── subprocess/        (no change)
├── core/              (fix getOrSpawn race + stderr drain)
├── aggregator/        (use config.LanguageForExt)
├── cache/             (no change)
├── hook/              (remove duplicate Diagnostic/Phase types)
├── gopls/             (DEPRECATED — migration target v3)
└── ARCHITECTURE.md    (NEW)
```

The `gopls` package should be preserved in a `gopls/` directory until migration is complete, then archived under `gopls/legacy/` with `//go:build ignore` stubs.

---

## Open Questions for v3 Architect

1. **Default `ClientImpl`**: lsp-client.md says `powernap_core` is the new default, but `ResolveClientImpl()` defaults to `gopls_bridge`. What is the rollout plan for switching the default?

2. **`getOrSpawn` caller contract**: Should `RouteFor` return a client in `StateSpawning` that the caller must wait on, or should `RouteFor` always block until `StateReady`? The current behavior returns a not-yet-ready client to the second concurrent caller.

3. **Stderr logging vs discard**: Should subprocess stderr be logged (via `slog`) or discarded? gopls emits structured log events on stderr during startup. Discarding loses valuable diagnostics; logging adds volume.

4. **`hook.Diagnostic` string severity**: `hook.Diagnostic` uses string severity (`"error"`, `"warning"`) for JSON serialization compatibility with hook payloads. If migrated to `lsp.Diagnostic` (int severity), hook JSON output changes. Is this a breaking change for existing hook consumers?

5. **Bridge retirement timeline**: Which SPEC governs the gopls bridge retirement? `ralph/engine.go` and `loop/go_feedback.go` depend on `gopls.Diagnostic` — who owns the migration?

---

## References (file:line)

| Finding | File | Line |
|---------|------|------|
| D1-1 Background context for powernap | `transport/transport.go` | 65 |
| D2-1 StateShutdown terminal after binary-not-found | `core/client.go` | 199-204 |
| D2-3 Stderr pipe not consumed | `core/client.go` | 213-217 |
| D3-1 Duplicate language detection maps | `core/manager.go:227`, `aggregator/aggregator.go:244` | — |
| D4-1 syscall.SIGKILL portability | `subprocess/supervisor.go` | 102 |
| D5-1 Two parallel config paths | `hook/gate.go` | 74-132 |
| D5-3 execTool 30s hardcoded timeout | `hook/fallback.go` | 258-260 |
| D6-2 detectLanguageFromPath manual scan | `aggregator/aggregator.go` | 244-293 |
| D7-1 Bridge inline circuit breaker | `gopls/bridge.go` | 21-26, 591-622 |
| D7-2 Hardcoded LanguageID "go" | `gopls/bridge.go` | 259 |
| D7-3 cmd.Wait goroutine leak on double-close | `gopls/bridge.go` | 477-479 |
| D8-1 gopls coverage below threshold | `gopls/*.go` | — |
| D8-2 Typo compareDignostics | `hook/tracker.go` | 147-148 |
| D10-1 nil transport race on concurrent spawn | `core/manager.go` | 332-363 |
| D10-2 Full document text in didChange | `core/document.go` | 98-110 |
| Phase type duplication | `hook/types.go:66`, `core/quality/trust.go:72` | — |
| Diagnostic type proliferation | `lsp/models.go:114`, `hook/types.go:46`, `gopls/protocol.go` | — |
| Default ClientImpl mismatch | `lsp/config/types.go` | 82-93 |

---

*Audit by: A3 — LSP Subsystem — 2026-04-23*
*Scope: 18,718 LOC, 9 packages, 73 source files*
*Tests: all pass, race-clean*
