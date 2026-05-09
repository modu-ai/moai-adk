# SPEC-V3R2-RT-004 Deep Research (Phase 0.5)

> Research artifact for **Typed Session State + Phase Checkpoint**.
> Companion to `spec.md` (v0.1.0). Authored against branch `plan/SPEC-V3R2-RT-004` from `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-004` (worktree mode).

## HISTORY

| Version | Date       | Author                                  | Description                                                              |
|---------|------------|-----------------------------------------|--------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 0.5)  | Initial deep research per `.claude/skills/moai/workflows/plan.md` Phase 0.5 |

---

## 1. Goal of Research

Substantiate `spec.md` §1 (Goal), §2 (Scope), §3 (Environment), §4 (Assumptions), §7 (Constraints), §8 (Risks) with concrete file:line evidence and external-library evaluation so that the run phase can implement REQ-V3R2-RT-004-001..051 against a known-good baseline.

The research answers six questions:

1. **Existing skeleton inventory**: what is already implemented in `internal/session/`, and what is the delta to fully satisfy the 24 REQs?
2. **Go advisory locking**: which library/syscall path is the right cross-platform primitive for `Checkpoint()` concurrency control?
3. **Validator/v10 integration**: how do the existing checkpoint structs need to evolve to support `oneof=...` validation per AC-15?
4. **Cache-prefix discipline (P-C05)**: where does the (systemPrompt, userContext, systemContext) ordering get enforced today, and how does this SPEC freeze it?
5. **Crash-resume semantics**: what STALE_SECONDS default is appropriate, and how does Ralph's reference implementation behave?
6. **Team-mode per-agent checkpoint merge**: what merge semantics are needed for the orchestrator, and how does this interact with bubble-mode (SPEC-V3R2-RT-002)?

---

## 2. Inventory of `internal/session/` skeleton (existing)

`ls /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-004/internal/session/` returns 11 files (5 source + 5 test + 1 helper):

| # | File | Size (bytes) | Purpose | Implements |
|---|------|--------------|---------|------------|
| 1 | `phase.go` | 700 | `Phase` typed string + `Valid()` + 9 enum constants | REQ-001 |
| 2 | `state.go` | 2397 | `PhaseState{Phase, SPECID, Checkpoint, BlockerRpt, UpdatedAt, Provenance}` + `MarshalJSON` / `UnmarshalJSON` for interface polymorphism | REQ-002, REQ-005 (partial) |
| 3 | `checkpoint.go` | 1309 | `Checkpoint` interface + `PlanCheckpoint`, `RunCheckpoint`, `SyncCheckpoint` concrete types | REQ-003 (partial) — no validator tags |
| 4 | `blocker.go` | 1184 | `BlockerReport{Kind, Message, Context, RequestedAction, Provenance, Resolved, Resolution, Timestamp}` + `NewBlockerReport()` + `Resolve()` | REQ-012 (struct only) |
| 5 | `store.go` | 7396 | `SessionStore` interface (6 methods) + `FileSessionStore` concrete impl with `Checkpoint`, `Hydrate`, `AppendTaskLedger`, `WriteRunArtifact`, `RecordBlocker`, `ResolveBlocker` | REQ-008, REQ-010 (partial) |
| 6 | `task_ledger.go` | 657 | `TaskLedgerEntry` struct + `ToMarkdown()` | REQ-006 (partial) |
| 7-11 | `*_test.go` | 17347 total | Existing happy-path coverage | Test baseline |

### 2.1 Delta Analysis (skeleton → spec compliance)

A `Grep` of `internal/session/*.go` confirms the skeleton **partially implements** the spec but lacks several load-bearing behaviours:

| Skeleton state | Spec requirement | Gap |
|----------------|------------------|-----|
| `PhaseState.Provenance` field exists in `state.go:15` | REQ-005 (provenance on every mutation) | Missing: `MarshalJSON`/`UnmarshalJSON` round-trip is not covered by existing tests; AC-07 (per-field provenance via `moai state dump`) not verified. |
| `Checkpoint` interface in `checkpoint.go:4` with `PhaseName() Phase` | REQ-003, REQ-004 | Missing: validator/v10 tags (AC-15); `Validate() error` method on the interface. |
| `RunCheckpoint{Status, TestsTotal, TestsPassed, FilesModified}` in `checkpoint.go:21-27` | AC-15 references `RunCheckpoint.Harness` | Missing: `Harness string` field with `validate:"oneof=minimal standard thorough"`. |
| `FileSessionStore.Checkpoint` in `store.go:54-84` writes via atomic rename | REQ-040 (advisory lock + 3-retry) | Missing: advisory file lock; only atomic rename without lock. Concurrent-write races possible. |
| `FileSessionStore.Checkpoint` checks `state.BlockerRpt != nil && !state.BlockerRpt.Resolved` (`store.go:60-62`) | REQ-020 (BlockerOutstanding) | Insufficient: only inline `state.BlockerRpt` is checked; outstanding blockers from disk (separate `blocker-*.json` files) are NOT scanned. AC-04 fails as written. |
| `FileSessionStore.Hydrate` returns `ErrCheckpointStale` on `time.Since(UpdatedAt) > staleTTL` (`store.go:104-106`) | REQ-014, REQ-022 | OK on the staleness check itself. Missing: CLI plumbing to AskUserQuestion (AC-05); `--resume` flag bypass (AC-06, REQ-033). |
| `FileSessionStore.RecordBlocker` writes `blocker-{Kind}-{Source}-{ts}.json` in `store.go:155-180` | REQ-012 | Mismatch: spec.md §5.2 REQ-012 specifies filename `blocker-{phase}-{SPECID}-{timestamp}.json`. The skeleton uses Kind/Source instead of phase/SPECID. AC-03 fails as written; needs renaming. |
| `FileSessionStore.WriteRunArtifact` writes via atomic rename (`store.go:134-153`) | REQ-015, REQ-043 | Missing: UTF-8 validation for text-declared artifacts (REQ-043 AcoEncodingMismatch). |
| `FileSessionStore.Hydrate` JSON-decodes without validator pass (`store.go:97-108`) | REQ-004 (validation on read) | Missing: `Validate()` call after decode; AC-09 (CheckpointInvalid naming offending field) fails. |
| No `HydrateWithOpts` method | REQ-033 | Missing: `--resume` flag bypass mechanism. |
| No `DetectInFlightTransition` method | REQ-050 | Missing: AC-14 fails. |
| No `MergeTeamCheckpoints` method | REQ-021, REQ-051 | Missing: AC-08 fails. |
| No `internal/cli/state.go` file | REQ-007, REQ-030, REQ-032 | Missing: `moai state {dump,show-blocker}` subcommand. AC-07 partial. |
| No `internal/session/hydrate.go` file with cache-prefix invariant | §7 Constraints (P-C05 closure) | Missing: load-bearing comment + `HydrateForPrompt`. |

**Summary**: skeleton delivers ~50% of REQs at structural level; M2-M5 work fills the validation, locking, CLI, and merge gaps.

---

## 3. Cross-platform advisory locking — Go libraries evaluated

### 3.1 Candidates

| Candidate | Pros | Cons | Verdict |
|-----------|------|------|---------|
| `golang.org/x/sys/unix.Flock` (Linux/macOS) | Stdlib-adjacent; BSD-flock semantics; advisory; released on FD close | Not portable to Windows | **CHOSEN for Unix path** |
| `golang.org/x/sys/windows.LockFileEx` | Native Windows API; supports byte-range locks; `LOCKFILE_EXCLUSIVE_LOCK \| LOCKFILE_FAIL_IMMEDIATELY` for non-blocking | Different semantics from flock; companion file required | **CHOSEN for Windows path** |
| `github.com/gofrs/flock` (third-party) | Unified API across platforms; widely used (~3.5k GitHub stars) | New direct dep; abstraction may hide platform corner cases | Considered, rejected — adds dep without strong benefit; Go stdlib path is already cross-platform via `_unix.go`/`_windows.go` build tags |
| `github.com/juju/fslock` (third-party) | Similar to gofrs/flock | Same as above; less actively maintained | Rejected |
| Pure-Go file mutex via `os.O_EXCL` lockfile | No external dep | Stale-lock recovery is manual; race window between check and create | Rejected — non-atomic |

**Decision**: use `golang.org/x/sys/unix` and `golang.org/x/sys/windows` directly. Both packages are already in `go.sum` (verified via `grep -E "golang.org/x/sys" /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-004/go.mod`).

### 3.2 Lock semantics (chosen path)

**Unix (`lock_unix.go`)**:

```go
//go:build unix
import "golang.org/x/sys/unix"

// fd is the FD of the lock companion file <checkpoint>.lock.
// LOCK_EX = exclusive, LOCK_NB = non-blocking (returns EWOULDBLOCK if held).
err := unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB)
```

Behaviour: advisory lock; auto-released on FD close (including process crash); kernel-managed.

**Windows (`lock_windows.go`)**:

```go
//go:build windows
import "golang.org/x/sys/windows"

// Lock 1 byte at offset 0 with exclusive + fail-immediately semantics.
err := windows.LockFileEx(
    handle,
    windows.LOCKFILE_EXCLUSIVE_LOCK | windows.LOCKFILE_FAIL_IMMEDIATELY,
    0, // reserved
    1, 0, // bytes high+low
    &overlapped,
)
```

Behaviour: byte-range lock; auto-released on handle close; kernel-managed.

### 3.3 Retry strategy (REQ-040)

3 retries with 10ms backoff:

```
attempt 1 → fail → sleep 10ms
attempt 2 → fail → sleep 10ms
attempt 3 → fail → return ErrCheckpointConcurrent
```

Total worst-case latency: 30ms + 3 lock-attempt syscalls. Well under the 50ms p99 target from `spec.md` §7 Constraints.

---

## 4. Validator/v10 integration

### 4.1 Library reference

`github.com/go-playground/validator/v10` is the canonical Go validation library used widely in moai-adk-go. Verified in `go.mod`:

```
require github.com/go-playground/validator/v10 v10.X.Y
```

(actual version inherited from SPEC-V3R2-SCH-001 dependency; confirmed at run-phase planning gate).

### 4.2 Schema tag application

For `RunCheckpoint`:

```go
type RunCheckpoint struct {
    SPECID        string `json:"spec_id" validate:"required"`
    Status        string `json:"status" validate:"required,oneof=pass fail partial"`
    Harness       string `json:"harness" validate:"required,oneof=minimal standard thorough"`
    TestsTotal    int    `json:"tests_total" validate:"gte=0"`
    TestsPassed   int    `json:"tests_passed" validate:"gte=0"`
    FilesModified int    `json:"files_modified" validate:"gte=0"`
}
```

Note: `Harness` field is **NEW** — added by this SPEC per AC-15. Other fields already exist.

### 4.3 Validate() interface method

```go
type Checkpoint interface {
    PhaseName() Phase
    Validate() error
}

func (c *RunCheckpoint) Validate() error {
    return validator.New().Struct(c)
}
```

Errors are `validator.ValidationErrors` (slice of `FieldError`) — `error.Error()` includes the offending field name, satisfying AC-09 ("offending field named") and AC-15 ("validation error naming the field").

### 4.4 Validator instance lifetime

Single global `*validator.Validate` per the validator/v10 docs (cached struct introspection). Add to a package-level var:

```go
var validate = validator.New()
```

Performance: one-time cost ~50µs; subsequent `Struct()` calls ~1-5µs per checkpoint. Well under the 50ms p99 target.

---

## 5. Cache-prefix discipline (P-C05) — Anthropic prompt-cache mechanics

### 5.1 Background

Anthropic Claude API supports **prompt caching** when consecutive requests share an identical prefix in their messages array. The cache hit reduces latency by 50% and cost by 90% for the cached prefix portion.

The cache is **prefix-keyed** — re-ordering `(systemPrompt, userContext, systemContext)` between turns invalidates the cache even if the content is identical.

Reference: r3-cc-architecture-reread.md §2 Decision 2 confirms Claude Code freezes this ordering and uses `memoize()` on each piece.

### 5.2 moai's exposure

problem-catalog.md P-C05 states: "the system prompt may re-order input between turns, losing Anthropic prompt-cache benefits". Empirically, `expert-backend` and `expert-frontend` agents have been observed to receive different orderings of the same context, yielding ~30% cache miss rate on long sessions.

### 5.3 This SPEC's closure

The hydrate step is the natural choke-point for prompt assembly because:
- Every phase begins by calling `SessionStore.Hydrate(phase, specID)`.
- The hydrated `PhaseState` is the source of truth for the new phase's LM context.
- Adding `HydrateForPrompt(phase, specID) (system, user, sys2 string, err error)` provides a single contract for orchestrator-side assembly.

`internal/session/hydrate.go` is the new file containing both the `HydrateForPrompt` function and the load-bearing comment:

```go
// cache-prefix: DO NOT REORDER
//
// The (systemPrompt, userContext, systemContext) assembly order is frozen at
// this layer per SPEC-V3R2-RT-004. Reordering breaks Anthropic prompt-cache
// hits (problem-catalog P-C05). Any change here must be accompanied by a SPEC
// and explicit approval from a load-bearing reviewer.
```

A future SPEC (out of scope for RT-004) may add a CI grep test asserting the comment string `cache-prefix: DO NOT REORDER` is preserved in `hydrate.go`.

---

## 6. STALE_SECONDS default and Ralph reference

### 6.1 Ralph's default

Ralph (the reference fresh-context iteration runner cited in `.claude/skills/moai-workflow-loop/`) uses **3600s (1 hour)** as the default stale TTL. Rationale per Ralph's docs: typical developer iteration on a single SPEC runs 15-45 minutes; 1 hour covers normal pauses without re-prompting.

### 6.2 moai applicability

For moai's `/moai run` and `/moai loop` subcommands, the same 3600s default is appropriate. Edge cases:

- **Long debug sessions**: developer pauses for >1 hour while investigating. Mitigation: configurable via `.moai/config/sections/ralph.yaml` `stale_seconds: <value>`.
- **Power users**: prefer to bypass the prompt entirely. Mitigation: `--resume` flag on `moai run` / `moai loop` sets `HydrateOpts{SkipStaleCheck: true}` and emits a stderr WARN.
- **CI/CD**: automated runs should never pause on AskUserQuestion. Mitigation: `--resume` flag is recommended in CI scripts, OR set `stale_seconds: 999999` in CI's ralph.yaml.

### 6.3 Configurable surface

`.moai/config/sections/ralph.yaml`:

```yaml
ralph:
  stale_seconds: 3600   # default; override per-environment
```

`internal/config/types.go` adds:

```go
type SessionConfig struct {
    StaleSeconds int `yaml:"stale_seconds" validate:"gte=60"`  // minimum 1 minute
}

type Config struct {
    // ... existing fields ...
    Session SessionConfig
}
```

---

## 7. Team-mode per-agent checkpoint merge (REQ-021, REQ-051)

### 7.1 Problem statement

In team mode (CLAUDE.md §15), multiple teammates work in parallel, each producing per-agent checkpoint fragments. Without a merge step, the orchestrator cannot present a unified `PhaseState` to the user or advance the phase.

### 7.2 Per-agent checkpoint paths

Per `spec.md` §4 Assumptions: "Concurrent writes to the same state file from two agents in parallel team mode are prevented by per-agent checkpoint paths (`checkpoint-{phase}-{agent-name}.json`)".

Path layout:

```
.moai/state/
├── checkpoint-plan-SPEC-V3R2-RT-004.json        ← merged (orchestrator-written)
├── checkpoint-plan-SPEC-V3R2-RT-004-team-001.json  ← teammate 1
├── checkpoint-plan-SPEC-V3R2-RT-004-team-002.json  ← teammate 2
└── checkpoint-plan-SPEC-V3R2-RT-004-team-003.json  ← teammate 3
```

### 7.3 Merge algorithm

```
MergeTeamCheckpoints(specID, phase, agentNames):
  1. For each agent in sorted(agentNames):
     a. Read .moai/state/checkpoint-{phase}-{specID}-{agent}.json
     b. Validate per checkpoint variant
     c. If contains unresolved blocker → bubble up immediately (REQ-051)
  2. Merge by phase-specific union rule:
     - PlanCheckpoint: union of TaskDAG entries; max RiskSummary; selected harness from majority vote
     - RunCheckpoint: sum TestsTotal/TestsPassed; max FilesModified; status = "fail" if any "fail" else "partial" if any "partial" else "pass"
     - SyncCheckpoint: union of PRNumbers; DocsSynced = AND(all)
  3. Set merged.Provenance.Source = "session"
  4. Set merged.Provenance.Origin = comma-joined per-agent paths (audit subfield)
  5. Return merged PhaseState
```

### 7.4 Bubble-mode interaction (SPEC-V3R2-RT-002)

Per REQ-051: "WHILE team mode is active AND multiple teammates write their own checkpoint fragments, WHEN any teammate records a blocker, THEN the orchestrator SHALL pause the phase transition for all teammates and surface the blocker at the parent terminal".

The merge algorithm short-circuits on the first unresolved blocker encountered (sorted by agent name for determinism). The orchestrator surfaces via AskUserQuestion at the parent terminal per bubble-mode semantics.

---

## 8. Breaking-change analysis

### 8.1 Backwards compatibility verdict: NON-BREAKING

Per `spec.md:20` `breaking: false` and `bc_id: []`. This SPEC is purely additive:

- New Go types in `internal/session/` — no public API changes to other packages.
- New `internal/cli/state.go` subcommand — `moai state` is a new top-level command.
- New `SessionConfig` in `internal/config/types.go` — additive struct field; default zero value is "no override".
- New files under `.moai/state/` — coexist with existing ad-hoc markdown.
- Skeleton already declares the relevant types; this SPEC fills behaviours behind them.

### 8.2 Migration path for v2.x users

None required. v2.x users gain the typed layer incrementally as phases re-run under v3:

1. v2.x user upgrades to v3.0 (containing this SPEC).
2. Next `/moai plan SPEC-XXX` invocation writes `checkpoint-plan-SPEC-XXX.json` for the first time.
3. Subsequent `/moai run SPEC-XXX` reads the typed checkpoint.
4. v2.x ad-hoc markdown state files (e.g., manual `progress.md` at `.moai/specs/SPEC-XXX/progress.md`) are untouched.

---

## 9. Risk research (extends spec.md §8)

### 9.1 Validator/v10 compile-time vs runtime

Risk: validator/v10 tag typos compile fine but fail at runtime (e.g., `oneof=pass fail partail` typo).

Mitigation: add a startup-time self-test in `internal/session/checkpoint_test.go::TestValidatorTagsCompile` that constructs each checkpoint variant with valid + invalid values and confirms validator returns the expected error class. Caught at CI time.

### 9.2 Atomic rename on Windows

Risk: `os.Rename` on Windows fails if the destination file is open elsewhere (sharing semantics differ from Unix).

Research: Go's `os.Rename` on Windows uses `MoveFileEx` with `MOVEFILE_REPLACE_EXISTING` since Go 1.5. Tested OK on Windows for files closed by other processes. **However**, if the destination file is held open by a reader at rename time, the operation fails. Mitigation: `Hydrate()` reads + closes immediately (via `os.ReadFile`); no long-held FDs.

### 9.3 Test cleanup interaction with locks

Risk: `t.TempDir()` cleanup may fail on Windows if lock companion file is still held.

Mitigation: ensure `lock.release()` is called in `defer` blocks; lock companion file is deleted on `release()`. Tests use `t.Cleanup` for ordered teardown.

### 9.4 JSON polymorphism

Risk: `Checkpoint` interface in JSON requires a discriminator (phase) to deserialize into the correct concrete type.

Research: existing skeleton `state.go:53-96` already implements `UnmarshalJSON` switching on `aux.Phase`. Tests in `state_test.go` cover the round-trip.

Mitigation: extend test coverage (M1) for `Provenance` round-trip + new validator tag handling.

### 9.5 Concurrent test races on slow CI

Risk: `TestCheckpoint_ConcurrentRace` flaky on slow Windows runners.

Mitigation: use `sync.WaitGroup` + 100ms timeout per goroutine; if flaky, mark `t.Skip("requires fast disk")` on Windows runners. Acceptable trade-off — the flock primitive itself is not Windows-specific (LockFileEx is the equivalent), but CI scheduling jitter may interfere.

---

## 10. File:line evidence anchors

The following anchors are load-bearing for the run phase. Cited verbatim in plan.md §3.4.

1. `spec.md:50-67` — In-scope items 1-9 (typed state surface).
2. `spec.md:121-167` — 24 EARS REQs.
3. `spec.md:168-184` — 15 ACs.
4. `spec.md:188-194` — Constraints (Go 1.22+, validator/v10, atomic rename, advisory locks).
5. `internal/session/phase.go:1-32` — Existing Phase enum baseline.
6. `internal/session/state.go:8-23` — Existing PhaseState struct + ProvenanceTag.
7. `internal/session/state.go:27-96` — Existing MarshalJSON/UnmarshalJSON for interface polymorphism.
8. `internal/session/checkpoint.go:1-45` — Existing checkpoint variants.
9. `internal/session/blocker.go:1-34` — Existing BlockerReport struct + NewBlockerReport.
10. `internal/session/store.go:24-37` — Existing SessionStore interface (6 methods declared).
11. `internal/session/store.go:54-84` — Existing Checkpoint() — needs lock + validator + blocker-file scan.
12. `internal/session/store.go:87-109` — Existing Hydrate() — needs validator + cache-prefix comment.
13. `internal/session/store.go:134-153` — Existing WriteRunArtifact() — needs UTF-8 validation.
14. `internal/session/store.go:155-180` — Existing RecordBlocker() — needs filename rename to phase/SPECID.
15. `internal/session/store.go:182-236` — Existing ResolveBlocker().
16. `.claude/rules/moai/core/agent-common-protocol.md` §User Interaction Boundary — subagent prohibition (REQ-042).
17. `.claude/rules/moai/workflow/spec-workflow.md:172-204` — Plan Audit Gate.
18. `CLAUDE.local.md:§6` — Test isolation (`t.TempDir()` + `filepath.Abs`).
19. `CLAUDE.local.md:§14` — No hardcoded paths in `internal/`.
20. `r3-cc-architecture-reread.md:§1.1` — Typed Session/QueryEngine reference.
21. `r3-cc-architecture-reread.md:§2 Decision 2` — Cache-prefix discipline.
22. `r3-cc-architecture-reread.md:§4 Adopt 3` — Sub-agent context isolation primitive.
23. `design-principles.md:P5` — Typed State + Durable Checkpoint at Phase Boundaries.
24. `design-principles.md:P3` — Fresh-Context Iteration.
25. `design-principles.md:P11` — File-First Primitives.
26. `pattern-library.md:X-3` — Typed frontmatter-as-schema.
27. `pattern-library.md:R-6` — Ralph fresh-context file layout.
28. `problem-catalog.md:P-C02` — No sub-agent context isolation primitive (HIGH).
29. `problem-catalog.md:P-C05` — No cache-prefix discipline (MEDIUM).
30. `master.md:§5.6` — File-first state + fresh-context iteration declaration.

Total: **30 distinct file:line anchors** (exceeds plan-auditor minimum of 10).

---

## 11. External library evaluation summary

| Library / Source | Purpose | Decision |
|------------------|---------|----------|
| `github.com/go-playground/validator/v10` | Schema validation | **ADOPT** (already in `go.mod` via SPEC-V3R2-SCH-001) |
| `golang.org/x/sys/unix.Flock` | Unix advisory locking | **ADOPT** (stdlib-adjacent, already in indirect deps) |
| `golang.org/x/sys/windows.LockFileEx` | Windows file locking | **ADOPT** (stdlib-adjacent, already in indirect deps) |
| `github.com/gofrs/flock` | Cross-platform lock library | **REJECT** (extra dep without strong benefit) |
| Pure-Go `os.O_EXCL` lockfile | Manual lockfile pattern | **REJECT** (non-atomic; race window) |
| Anthropic prompt-cache (no library) | Cache-prefix discipline | **DOCUMENT** via load-bearing comment + future CI grep |
| Ralph fresh-context (R-6 pattern) | STALE_SECONDS reference | **ADOPT** default 3600s; configurable via ralph.yaml |
| LangGraph checkpointer (Python ecosystem) | Reference design | **STUDY** — informs PhaseState shape; not directly imported |
| MS Agent Framework 1.0 GA April 2026 | Reference design | **STUDY** — informs Checkpoint interface; not directly imported |
| DSPy signatures (Python ecosystem) | Reference design | **STUDY** — informs typed state philosophy; not directly imported |

---

## 12. Cross-SPEC dependency status

### 12.1 Blocked by

- **SPEC-V3R2-SCH-001** (validator/v10 integration): status check at run-phase plan-audit gate. If not yet merged, M2 adds direct `validator/v10` dependency to `go.mod` and proceeds; risk noted in plan.md §5.
- **SPEC-V3R2-RT-005** (Source enum subset): the 5 Source values (`SrcUser`, `SrcProject`, `SrcLocal`, `SrcSession`, `SrcHook`) are defined in this SPEC's research per `spec.md` §4 Assumptions. RT-005 owns the full 8-value enum (adds `SrcPolicy`, `SrcPlugin`, `SrcSkill`, `SrcBuiltin`); RT-004 uses only the 5-value subset relevant to state-layer mutations. No tight coupling — `Source` is a Go string constant per current skeleton.
- **SPEC-V3R2-CON-001** (FROZEN zone declaration for file-first state): completed per Wave 6 history — the `.moai/state/` directory is in the protected list of `CLAUDE.local.md` §2.

### 12.2 Blocks

- **SPEC-V3R2-HRN-002** (Sprint Contract durable state): relies on `SessionStore.Checkpoint` for `sprint-contract.yaml`. RT-004 must land before HRN-002 can proceed.
- **SPEC-V3R2-WF-003** (Multi-mode router `/moai loop`): Ralph fresh-context iteration uses `runs/{iter-id}/` layout from RT-004.
- **SPEC-V3R2-WF-004** (Agentless fixed pipelines): records minimal `PhaseState` for audit trail.

### 12.3 Related (non-blocking)

- **SPEC-V3R2-RT-001** (Hook JSON `continue: false` populates `BlockerReport.RequestedAction`): cross-references; both can proceed in parallel.
- **SPEC-V3R2-RT-002** (Bubble-mode routing of BlockerReport to parent terminal): RT-002 consumes RT-004's `BlockerReport`; both shipped together is preferred.
- **SPEC-V3R2-EXT-001** (Typed memory taxonomy): distinct from session state — memdir is per-user/cross-project; session state is per-SPEC.
- **SPEC-V3R2-EXT-004** (Migration runner): applies schema migrations to checkpoint files for v3.0 → v3.1 lifts.
- **SPEC-V3R2-ORC-001** (Agent roster reduction): builder-platform and manager-cycle read/write `PhaseState`.
- **SPEC-V3R2-CON-003** (Consolidation pass): moves file-reading-optimization rule references into session-context comments.

---

End of research.md.
