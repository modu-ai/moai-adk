# Research — SPEC-UTIL-003

> LSP Subprocess Hygiene (stderr drain + singleflight spawn + Diagnostic Alias)
> Release track: `release/v2.14.0` — Phase 3 — Utility Hardening
> Date: 2026-04-24
> Author: Wave v2.14 SPEC Writer

---

## 1. Problem Statement

SPEC-UTIL-003 closes the two P0 LSP findings from the Wave audits plus the internal-only portion of D2 (Diagnostic type unification):

- **IMP-V3U-002 — LSP subprocess stderr leak** (SYNTHESIS §3.1 D2-3; `a3-lsp-audit.md` §D2-3). `internal/lsp/core/client.go:213` constructs a `readWriteCloser` from `result.Stdout` and `result.Stdin` only. `result.Stderr` is returned by `subprocess.Launcher.Launch` (`launcher.go:146`) but is never attached to a reader. When a language server writes more than the OS pipe buffer (~64 KiB on Linux/macOS) to stderr — a routine occurrence for `gopls -rpc.trace` verbose logging or `pyright` large-project startup — the pipe buffer fills, the subprocess's `write(2)` blocks, and the LSP main loop deadlocks because the server cannot flush trace output between JSON-RPC responses. Symptom reported in field: client reports `context deadline exceeded` on `textDocument/diagnostic`, but `ps` shows the server parked in `D` (uninterruptible disk wait) on a `write` to stderr.

- **IMP-V3U-003 — LSP `getOrSpawn` concurrent client race** (SYNTHESIS §3.1 D10-1; `a3-lsp-audit.md` §D10-1). `internal/lsp/core/manager.go:332-363` holds `m.mu` while checking the cache, then releases `m.mu` before calling `c.Start(ctx)`. Two goroutines that both call `RouteFor` for the same language in the same ~100 μs window can both observe the cache miss, both insert freshly-constructed `client` pointers at `m.clients[language]`, and the loser's client never has `Start` called on it. The caller of the loser goroutine then invokes methods on a client whose `state` is still `StateSpawning` — race detector flags this as a write-after-write on `m.clients`, and end users see intermittent `lsp: client not started` errors on cold-path `textDocument/*` requests that arrive within one `RouteFor` tick of each other.

- **D2 Diagnostic type unification (internal-only alias scope)**. Three distinct `Diagnostic` structs exist today: `lsp.Diagnostic` (`internal/lsp/models.go:115`, `Severity` typed as `DiagnosticSeverity int`), `gopls.Diagnostic` (`internal/lsp/gopls/messages.go:130`, `Severity` typed as `DiagnosticSeverity int`), and `hook.Diagnostic` (`internal/lsp/hook/types.go:47`, `Severity` typed as `DiagnosticSeverity string`). The divergence forces `internal/hook/post_tool.go:481-498` to maintain a manual `convertHookDiagsToLSP` translation function mapping string severities (`"error"`, `"warning"`, ...) to int severities (`1`, `2`, ...). Under D2 the `lsp.Diagnostic` struct becomes the single source of truth, and consumer packages expose Go type aliases pointing at it so the translation layer can eventually be retired. The caveat surfaced during research: `hook.Diagnostic`'s JSON wire format uses string severity and is consumed by external hook consumers (see §5 Risk Assessment).

Cross-cutting context: all three items land in the same v2.14.0 release because they share the `internal/lsp/` subsystem and because the audit classifies each as Tier 1 Critical with effort XS-S. Per §0.5 D3 the stderr policy is `io.Discard` drain (simplest pattern, deadlock eliminated, debugging-insight trade-off accepted). Per §0.5 D2 the type unification is internal-only with Go type aliases (no wire format change in v2.14).

---

## 2. Current Implementation Analysis

### 2.1 `client.go:213-217` — stderr unclaimed

```
// internal/lsp/core/client.go
209    if result.Cmd != nil {
210        c.supervisor = subprocess.NewSupervisor(result)
211    }
212
213    stream := &readWriteCloser{
214        r: result.Stdout,
215        w: result.Stdin,
216        closers: []io.Closer{result.Stdin, result.Stdout},
217    }
218    c.tr = c.trFactory(stream)
```

The `readWriteCloser` bundles only Stdout (read-side) and Stdin (write-side). `result.Stderr` (typed `io.ReadCloser`, assigned in `launcher.go:146`) is never consumed after `Launch` returns. `subprocess.NewSupervisor` retains a reference to the `*exec.Cmd` for `Watch`/`Signal`/`Kill` but does not spawn a stderr pump.

Downstream effects:

- The OS-level kernel pipe buffer for stderr (`/proc/sys/fs/pipe-max-size` on Linux, default 64 KiB; macOS default 16 KiB) fills when the server emits `~64 KiB`+ of trace output.
- `write(2)` on the server side blocks in `D` state.
- The server's JSON-RPC write loop — which shares a single goroutine with its trace emitter on most reference implementations (gopls, pyright) — cannot emit the `textDocument/publishDiagnostics` notification the client is waiting for.
- From moai's perspective, `transport.CallWithTimeout` returns `context deadline exceeded` and `c.state` transitions to `StateDegraded`. The subprocess is not killed automatically because Supervisor's `Watch` only fires on exit.

### 2.2 `manager.go:332-363` — getOrSpawn locking pattern

```
// internal/lsp/core/manager.go
332    func (m *Manager) getOrSpawn(ctx context.Context, language string) (Client, error) {
333        m.mu.Lock()
334        existing, ok := m.clients[language]
335        if ok && existing.State() != StateShutdown {
336            m.mu.Unlock()
337            return existing, nil
338        }
...
346        c := m.clientFactory(sc)
347        m.clients[language] = c     // <-- half-ready client enters cache
348        m.mu.Unlock()
349
350        // Start 호출 (락 외부에서 실행 — 블로킹 가능)
351        if err := c.Start(ctx); err != nil {
352            m.mu.Lock()
353            if cur, still := m.clients[language]; still && cur == c {
354                delete(m.clients, language)
355            }
356            m.mu.Unlock()
357            return nil, fmt.Errorf(...)
358        }
359        return c, nil
360    }
```

Race scenario (timeline: G1 and G2 both call `RouteFor("go", ...)` at the same tick):

| Step | Goroutine | Action |
|------|-----------|--------|
| 1 | G1 | Acquires `m.mu`. `m.clients["go"]` empty. |
| 2 | G1 | `c1 := m.clientFactory(sc)`; `m.clients["go"] = c1`. |
| 3 | G1 | Releases `m.mu`. Begins `c1.Start(ctx)` (blocking). |
| 4 | G2 | Acquires `m.mu`. `m.clients["go"] == c1` but `c1.State() == StateSpawning` (not yet Ready). Condition `existing.State() != StateShutdown` is **true** (StateSpawning is not StateShutdown) — returns c1. |
| 5 | G2 | Callers of G2 receive c1, call `c1.OpenFile(...)` which dispatches through `c1.tr` — but `c1.tr == nil` until G1 completes Step 3. |

The guard `existing.State() != StateShutdown` is too permissive: it admits `StateSpawning` and `StateInitializing` as "reusable". The intended contract is "return only when the client is ready to serve LSP requests", which maps to `StateReady` or `StateDegraded` (degraded is still usable for partial capabilities).

An additional failure mode: if G2 arrives while G1 is *also* inside `c.Start`, and G1 hits a launch error at Step 3, then at Step 5 G2 is holding a pointer to a client that G1 is about to delete. Even if G2 checks the state again, the delete-then-use race is not bounded by the state machine.

Canonical fix pattern: `golang.org/x/sync/singleflight.Group.Do(language, func() (any, error) { ... })` — the first caller per key executes the function; subsequent callers block and receive the same result. This collapses the duplicate-factory race into a single execution per language.

### 2.3 Diagnostic type fragmentation

Three structs, three JSON shapes:

| Type | Location | Severity Field Type | JSON Severity | Consumer |
|------|----------|---------------------|---------------|----------|
| `lsp.Diagnostic` | `internal/lsp/models.go:115` | `DiagnosticSeverity` (int) | `1` `2` `3` `4` | core LSP client, transport layer |
| `gopls.Diagnostic` | `internal/lsp/gopls/messages.go:130` | `DiagnosticSeverity` (int, own declaration) | `1` `2` `3` `4` | gopls bridge path (SPEC-GOPLS-BRIDGE-001) |
| `hook.Diagnostic` | `internal/lsp/hook/types.go:47` | `DiagnosticSeverity` (string) | `"error"` `"warning"` `"information"` `"hint"` | hook JSON wire format, external consumers |

`gopls.Diagnostic` and `lsp.Diagnostic` have structurally identical layouts — same field names, same JSON tags, same int severity — making them eligible for a transparent Go type alias. The only latent cost is that `gopls.Range` and `gopls.Position` are declared in the `gopls` package (own `Range`/`Position` types); an alias on `Diagnostic` that uses `lsp.Range` would require the `gopls` package to either also alias `Range`/`Position` or rely on struct-literal compatibility. This research scopes the aliasing target accordingly (see §3).

`hook.Diagnostic` is **not** structurally compatible. Its JSON wire format uses string severity — hook consumers (`.claude/hooks/moai/*.sh`, external IDE plugins, SARIF exporters downstream of `collectDiagnosticsWithInstructionAndReturn`) rely on string severity values. A naive `type Diagnostic = lsp.Diagnostic` in the hook package would flip JSON severity from `"error"` to `1`, silently breaking wire format contracts. Per the v2.14.0 release plan §2.3 "Deferred items" rationale for IMP-V3U-008, field-level consolidation of `hook.Diagnostic` into `lsp.Diagnostic` requires a compatibility projection layer which is v2.16 types-domain scope.

---

## 3. Target State

### 3.1 stderr drain goroutine (REQ-UTIL-003-001 / 002)

After `NewSupervisor(result)` at `client.go:210`, spawn a detached goroutine that consumes `result.Stderr` into `io.Discard`:

```
if result.Stderr != nil {
    go func() {
        defer result.Stderr.Close()
        _, _ = io.Copy(io.Discard, result.Stderr)
    }()
}
```

Properties:

- The goroutine terminates naturally when the subprocess closes its stderr (on exit). `io.Copy` returns `nil` error; `result.Stderr.Close()` is idempotent against the already-closed pipe.
- `io.Discard` is the canonical `io.Writer` sink — zero allocations per write (see `io/io.go:624`).
- The `result.Stderr != nil` guard handles the test path where `launcher.Launch` synthesizes a `LaunchResult` without a real subprocess (`Cmd == nil`). Launcher code pairs the nil-Cmd case with nil-Stderr.
- No WaitGroup or context coordination is required. The goroutine is fire-and-forget by design; it does not participate in shutdown. Ownership is tied to the lifetime of `result.Stderr`, which is closed transitively by `Supervisor.Kill()`.

### 3.2 singleflight.Group guards clientFactory (REQ-UTIL-003-003 / 004 / 005 / 006)

Add `sf singleflight.Group` as a field on `Manager`:

```
type Manager struct {
    ...
    sf singleflight.Group  // NEW
}
```

Refactor `getOrSpawn` so the factory path is wrapped by `sf.Do(language, ...)`:

```
func (m *Manager) getOrSpawn(ctx context.Context, language string) (Client, error) {
    m.mu.Lock()
    if c, ok := m.clients[language]; ok && isReusable(c) {
        m.mu.Unlock()
        return c, nil
    }
    m.mu.Unlock()

    v, err, _ := m.sf.Do(language, func() (any, error) {
        m.mu.Lock()
        if c, ok := m.clients[language]; ok && isReusable(c) {
            m.mu.Unlock()
            return c, nil
        }
        sc, hasCfg := m.servers[language]
        if !hasCfg {
            m.mu.Unlock()
            return nil, fmt.Errorf(...)
        }
        m.mu.Unlock()

        c := m.clientFactory(sc)
        if err := c.Start(ctx); err != nil {
            return nil, fmt.Errorf(...)
        }

        m.mu.Lock()
        m.clients[language] = c
        m.mu.Unlock()
        return c, nil
    })
    if err != nil {
        return nil, err
    }
    return v.(Client), nil
}
```

Key changes from the current implementation:

- **Cache insertion deferred until after `Start` returns nil**. This eliminates the "half-ready client in cache" window. If Start fails, no cache entry is created, so no dangling client is observable.
- **`sf.Do(language, ...)` guarantees exactly one factory execution per key per in-flight call**. Concurrent callers for the same language block on the singleflight barrier; the first to arrive executes the factory, and the rest receive the same `(Client, error)` tuple.
- **`isReusable` helper** (or equivalent inline check): returns true when `c.State()` is `StateReady` or `StateDegraded`. `StateSpawning`, `StateInitializing`, `StateShutdown` are all non-reusable and trigger a fresh spawn.
- **Singleflight key is the language name**, not a computed hash. Collisions are impossible because `servers` is keyed on language.

### 3.3 gopls.Diagnostic type alias (REQ-UTIL-003-007 / 008 / 009)

`internal/lsp/gopls/messages.go` converts `type Diagnostic struct { ... }` to a type alias:

```
// Before
type Diagnostic struct {
    Range    Range              `json:"range"`
    Severity DiagnosticSeverity `json:"severity,omitempty"`
    Code     string             `json:"code,omitempty"`
    Source   string             `json:"source,omitempty"`
    Message  string             `json:"message"`
}

// After (single alias line + retain gopls.Range compatibility)
type Diagnostic = lsp.Diagnostic
```

Because `gopls.Range` and `gopls.Position` are currently declared in the `gopls` package with identical layouts to `lsp.Range` / `lsp.Position`, the alias requires either (a) aliasing `Range` and `Position` too, (b) keeping gopls-local `Range`/`Position` and accepting that `lsp.Diagnostic.Range` has type `lsp.Range` (not `gopls.Range`) in callers, or (c) leaving `gopls.Diagnostic` un-aliased for v2.14 and documenting the deferred scope. The SPEC scope lands on **(a) alias Range, Position, Diagnostic together** — three one-line `type X = lsp.X` declarations — because option (b) introduces a type mismatch that would ripple into `gopls/bridge.go` callers, and option (c) undershoots the D2 commitment.

No change to wire format: `lsp.Diagnostic` and the pre-alias `gopls.Diagnostic` have byte-identical JSON encodings. `encoding/json` respects the `json` struct tags on the aliased target type.

**`hook.Diagnostic` is explicitly out of scope for the alias** in v2.14 (see §5 Risk Assessment). The SPEC records the deferred item as "full field-level consolidation including `hook.Diagnostic` → `lsp.Diagnostic` → v2.16 types domain" and does not attempt partial aliasing that would break hook JSON wire format.

### 3.4 Invariants that MUST hold after the change

- `internal/hook/` JSON wire format for Diagnostic payloads is byte-identical before/after the change. (Verified by wire-format freeze test consuming a canonical hook JSON fixture.)
- `result.Stderr.Close()` is called exactly once during the subprocess lifetime (by the drain goroutine on copy completion, or transitively by Supervisor.Kill paths). No double-close panic.
- Concurrent `RouteFor(ctx, path)` calls for N goroutines on M distinct languages result in exactly min(N, M) factory invocations.
- `go test -race ./internal/lsp/...` is clean on the new test suite.
- No change to `Client` or `Manager` public API surface (method signatures, option constructors, exported struct fields).

---

## 4. External Research (A4 + SYNTHESIS)

### 4.1 `golang.org/x/sync/singleflight` canonical pattern (A4 T5)

The singleflight package ships in `golang.org/x/sync` (pinned at `v0.20.0` in `go.sum:88`). It is already a transitive dependency of the project via `go.mod:12` — no new import required at the module level.

Canonical usage per the Go standard library test corpus (e.g., `k8s.io/apimachinery/.../memcache.go`):

```
var sf singleflight.Group

func fetchOrCreate(key string) (*Thing, error) {
    v, err, _ := sf.Do(key, func() (any, error) {
        return create(key)
    })
    if err != nil {
        return nil, err
    }
    return v.(*Thing), nil
}
```

Semantics:

- **First caller per key executes the function**. Subsequent callers *with the same key* while the first call is in flight block on a condition variable internal to singleflight.
- **All waiting callers receive the same (value, error) tuple**. If the function returns `(nil, someErr)`, every waiter sees `someErr`. The `shared bool` third return value indicates whether the call was shared across multiple callers.
- **After the function returns**, subsequent callers with the same key execute the function again from scratch (no caching). This matches our use case: we want to retry on next call if Start failed.
- **Concurrent calls with different keys do not block each other**. Per-key serialization only.
- **Does not propagate context cancellation** — a canceled caller's context is ignored by the in-flight function. Call sites that need cancellation propagation must wrap the function body in a `ctx.Done()` select.

For our use case the lack of cancellation propagation is acceptable: `c.Start(ctx)` inside the factory already respects `ctx`. If the first caller's context is canceled mid-Start, `Start` returns an error and singleflight propagates that error to waiters. The waiters' own contexts are observed only on the *next* `sf.Do` call (i.e., the retry path).

### 4.2 powernap v0.1.4 subprocess hygiene (a4-external-best-practices.md T5; lsp-client.md)

`github.com/charmbracelet/x/powernap` is the upstream LSP client wrapper used under `internal/lsp/transport/`. The package (v0.1.4 pinned 2026-04-22) performs subprocess spawn via `os/exec` inside `powernap.lsp.Client.Start`. Per `a4-external-best-practices.md` §T5, the charmbracelet/crush reference implementation pairs the powernap client with a **user-maintained stderr drain** — powernap itself does not drain stderr. This SPEC's §3.1 fix mirrors the crush pattern, implemented at moai's integration boundary in `client.go:213` rather than inside powernap.

The Go 1.24+ subprocess hygiene baseline from the same research (`cmd.WaitDelay`, `Setpgid`, close-all-fds) is deliberately deferred to v2.15 subprocess hygiene baseline per release plan §2.3 IMP-V3U-021. v2.14 lands only the stderr drain, as the drain alone eliminates the immediate deadlock without requiring cross-platform build tags.

### 4.3 stderr buffering deadlock mechanics (a3-lsp-audit.md §D2-3)

The audit cites empirical evidence from Linux + macOS: writing more than `pipe-max-size` bytes (default 64 KiB Linux, 16 KiB macOS) to a stderr pipe with no reader attached causes the writing process to block in `write(2)`. gopls `-rpc.trace` emits ~2 KiB per JSON-RPC round-trip, so ~32 round-trips on Linux (or ~8 on macOS) fill the buffer. pyright's large-project startup log emits ~150 KiB in the first second, hitting the threshold immediately.

Mitigation options considered (a4 §T5):

1. **Consume to `io.Discard`** — what we ship. Simplest, lowest overhead, zero memory footprint. Trade-off: debugging insight is lost.
2. **Consume to ring buffer** — 64 KiB in-memory ring per client. Better debugging, zero disk. More code (~100 LOC ring buffer + tests). Deferred to v2.15+.
3. **Consume to `slog` with line discriminator** — preserves diagnostic value but couples LSP layer to slog. Deferred to v2.15+.
4. **File rotation to `~/.moai/logs/lsp-{lang}-stderr.log`** — best debugging, non-zero disk. Deferred to v2.15+.

Per §0.5 D3 the chosen path is option 1 for v2.14. Richer stderr handling is tagged as follow-on work in SPEC-UTIL-003's §6 dependencies section as a v2.15+ candidate.

### 4.4 Go type alias semantics (Go spec §Type declarations)

Go type aliases introduced in Go 1.9 have **identity semantics**: `type A = B` declares `A` as an alternate name for `B`. `A` and `B` are the **same type** — not just convertible, not just structurally identical. Consequences relevant to this SPEC:

- **Method sets are identical**. Any method defined on `B` is callable on a value typed `A` without a conversion.
- **Type assertions behave transparently**. `x.(A)` and `x.(B)` are interchangeable for any interface-typed `x`.
- **`reflect.TypeOf(a) == reflect.TypeOf(b)`** when `a` has type `A` and `b` has type `B`. Reflection-based code (including `encoding/json`'s struct tag reader) sees one type, not two.
- **Package-qualified names still differ in source**. Users read `gopls.Diagnostic` in imports but the underlying type is `lsp.Diagnostic`. godoc and IDE "go to definition" resolve to the aliased type.
- **`fmt.Sprintf("%T", x)` returns the canonical name** (the aliased target, `"lsp.Diagnostic"`), not the alias path. This is the only externally-visible reflection signal that changes; no production moai code currently formats Diagnostic types via `%T`.

For `hook.Diagnostic` the identity semantics are the blocker: aliasing `hook.Diagnostic` to `lsp.Diagnostic` would make the hook's JSON severity field `int` (from `lsp.DiagnosticSeverity`), flipping wire format. Go aliases cannot change the encoding/json representation of a type.

---

## 5. Risk Assessment

### R-UTIL-003-01 — hook.Diagnostic alias would break JSON wire format (resolved by scope restriction)

`hook.Diagnostic.Severity` is `string` JSON; `lsp.Diagnostic.Severity` is `int` JSON. Aliasing `hook.Diagnostic = lsp.Diagnostic` would silently flip the wire format for every hook JSON payload emitted by `internal/hook/post_tool.go:466` `emitToFeedbackChannel` and the feedback-channel consumer chain. External hook consumers that parse `"severity": "error"` would see `"severity": 1` and fail silently (string-typed deserializers in shell/Python would accept the integer as a string via JSON type coercion, but semantic comparisons against `"error"` would all return false).

**Mitigation**: This SPEC's scope does **not** alias `hook.Diagnostic`. Only `gopls.Diagnostic` is aliased in v2.14, plus `gopls.Range` and `gopls.Position` for cascade consistency. `hook.Diagnostic` consolidation (including wire format migration via a compatibility projection layer) is explicitly deferred to v2.16 types domain. The release plan §2.3 records this deferral.

### R-UTIL-003-02 — singleflight retains no result cache (acceptable)

`singleflight.Group.Do` does not memoize results; after the in-flight call completes, subsequent calls re-execute the function. This is intentional — our cache is `m.clients`, not the singleflight result. The semantics match: singleflight collapses *concurrent* spawns, `m.clients` handles *sequential* reuse. The two layers are complementary.

Residual risk: if factory execution is expensive and a burst of N callers arrives *sequentially*, each completing before the next arrives, all N pay the factory cost. Mitigation: `m.clients` cache handles this — after the first factory call writes to `m.clients`, subsequent callers hit the cache at the fast-path check before entering singleflight.

### R-UTIL-003-03 — singleflight blocking semantics (bounded)

`sf.Do` blocks the caller until the in-flight function returns. For our factory this means blocking until `c.Start(ctx)` completes — which can take several hundred ms on cold gopls startup. In practice this is desirable behavior (waiting callers get a working client instead of a half-initialized one), but it does mean that `RouteFor` latency for the second caller is "Start time" rather than "cache lookup time".

**Mitigation**: The factory respects `ctx`. If a waiting caller's context is canceled, they observe cancellation on their next `sf.Do` retry. No new deadlock risk because singleflight internally uses condition variables, not locks held across the function execution.

### R-UTIL-003-04 — io.Discard loses gopls debugging insight (trade-off accepted per D3)

Operators debugging gopls-specific misbehavior lose access to `-rpc.trace` output. Workarounds during debugging: manual `powernap` transport-level JSON-RPC logging via `WithLogger`, `gopls` serverside log file via gopls CLI flags (independent of moai's stderr handling), system-level `strace -f -e trace=write` attachment.

**Mitigation**: Document the trade-off in `.claude/rules/moai/core/lsp-client.md`. v2.15+ may introduce ring buffer or file rotation per §4.3 option 2/4.

### R-UTIL-003-05 — race detector noise from existing test patterns (mitigated by test design)

The existing `manager_test.go` has tests that spawn clients concurrently via a fake `clientFactory`. Some of these tests may have been passing only because the race was intermittent and the fake factory is instant (no Start blocking). After singleflight, some tests may block unexpectedly if they rely on both goroutines executing the factory.

**Mitigation**: Acceptance test `REQ-UTIL-003-011` explicitly verifies with race detector. Any pre-existing test depending on duplicate factory invocation must be updated to match the new semantics (single factory invocation per language per burst).

### R-UTIL-003-06 — gopls.Range / gopls.Position cascade

Aliasing `gopls.Diagnostic = lsp.Diagnostic` changes the type of its `Range` field from `gopls.Range` to `lsp.Range`. Callers that assign to `gopls.Diagnostic{Range: someGoplsRange}` compile-break unless `gopls.Range` is also aliased.

**Mitigation**: Alias all three (`Range`, `Position`, `Diagnostic`) simultaneously in the gopls package. Grep confirms `gopls.Range` and `gopls.Position` have no field-level drift from `lsp.Range` / `lsp.Position` (both use `Line int`, `Character int`, `Start Position`, `End Position` with identical JSON tags). Three alias declarations total.

---

## 6. References

### Source code (scanned)

- `/Users/goos/MoAI/moai-adk-go/internal/lsp/core/client.go` — lines 194-248 (Start method), 396-419 (readWriteCloser).
- `/Users/goos/MoAI/moai-adk-go/internal/lsp/core/manager.go` — lines 29-66 (Manager struct), 332-363 (getOrSpawn).
- `/Users/goos/MoAI/moai-adk-go/internal/lsp/models.go` — lines 32-48 (DiagnosticSeverity int), 114-130 (Diagnostic struct).
- `/Users/goos/MoAI/moai-adk-go/internal/lsp/gopls/messages.go` — lines 120-170 (PublishDiagnosticsParams, Diagnostic, Range, Position, DiagnosticSeverity int).
- `/Users/goos/MoAI/moai-adk-go/internal/lsp/hook/types.go` — lines 10-62 (DiagnosticSeverity string, Position, Range, Diagnostic).
- `/Users/goos/MoAI/moai-adk-go/internal/hook/post_tool.go` — lines 466-498 (emitToFeedbackChannel, convertHookDiagsToLSP).
- `/Users/goos/MoAI/moai-adk-go/internal/lsp/subprocess/launcher.go` — lines 47-150 (LaunchResult, Launch).

### Decisions and audits

- `/Users/goos/MoAI/moai-adk-go/docs/design/v2.14.0-release-plan.md` §0.5 D2 (Diagnostic unification internal-only), §0.5 D3 (stderr io.Discard), §2.2 UTIL-003 scope, §2.3 deferred items.
- `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a3-lsp-audit.md` — §D2-3 stderr leak, §D10-1 getOrSpawn race, full 7/10 audit score.
- `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/SYNTHESIS.md` — IMP-V3U-002, IMP-V3U-003 classification as P0 Critical.
- `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a4-external-best-practices.md` — §T3 singleflight canonical pattern, §T5 Go subprocess hygiene 2026.

### Dependency verification

- `go.mod:12`: `golang.org/x/sync v0.20.0` — direct dependency, singleflight package available.
- `go.sum:88-89`: verified hash for v0.20.0.
- `.claude/rules/moai/core/lsp-client.md`: powernap v0.1.4 pin, upgrade policy, co-existence with SPEC-GOPLS-BRIDGE-001.

### Sibling SPECs (same v2.14.0 release)

- SPEC-UTIL-001 — MX Validator Correctness + 16-Language Complexity (parallel).
- SPEC-UTIL-002 — ast-grep Integration Hardening + 5-Language Rule Seeding (parallel).

### Deferred items (explicit out-of-scope)

- **v2.15 subprocess hygiene baseline** (IMP-V3U-021): `cmd.WaitDelay`, `Setpgid`, close-all-fds cross-platform. Requires build-tag POSIX/Windows splits.
- **v2.16 types domain** (IMP-V3U-008): `hook.Diagnostic` field-level consolidation into `lsp.Diagnostic` with JSON wire format compatibility projection.
- **v3.0 breaking changes** (BC-V3R2-UTIL-001): `config.ResolveClientImpl()` default `"gopls_bridge"` → `"powernap_core"` flip.

---

Version: 0.1.0
Status: draft
Last Updated: 2026-04-24
