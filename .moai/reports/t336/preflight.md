# t336 pre-flight — integration lock acquire atomicity

Tree: `.claude/worktrees/t336`, branch `WT-integration-lock-atomic`, HEAD `15453140a` (== origin/develop, `0 0`).

## 1. Where the race window is (file:line, measured at 15453140a)

`internal/kanban/integration_lock.go`

- `AcquireIntegrationLock` L177-209 — the unserialized read-modify-write:
  - L186 `current, err := ReadIntegrationLock(projectRoot)`   ← READ
  - L190-200 held/stale/force decision                        ← MODIFY (decision)
  - L205 `writeIntegrationLock(path, &want)`                  ← WRITE
  Nothing holds an exclusive lock across L186..L205. Two processes can both read
  "free" (or both read the same stale record), both pass the decision, and both
  write. Last write wins; both callers are told they hold the window.
- `ReleaseIntegrationLock` L215-230 has the same shape (read L216 → holder check
  L220-225 → `os.Remove` L226) and the same gap.
- Secondary: `writeIntegrationLock` L257 uses a FIXED tmp path (`path + ".tmp"`).
  Two concurrent writers share one tmp file, so the interleaved
  `WriteFile`+`Rename` pair can publish a record neither caller composed. Under a
  mutation lock this cannot occur; a unique tmp (`os.CreateTemp`) is cheap
  defense-in-depth.

Header claim vs code: L18-20 states "the flock discipline is borrowed only to
serialize mutations of that record." No flock is taken anywhere in this file —
grep `flock` in `internal/kanban/integration_lock.go` returns only that prose.
The doctrine is written; the mechanism is absent.

## 2. Where the repair belongs relative to the PreToolUse guard

Two layers exist and they are independent:

- DENY layer — `internal/hook/integration_lock_guard.go` `checkIntegrationLock`
  (called from `internal/hook/pre_tool.go:551`), gated by
  `Workflow.IntegrationLock.Enabled`, distributed default `false`
  (`internal/config/defaults.go:846`, contract in `internal/config/types.go:630-639`).
  It is READ-ONLY on the record.
- RECORD layer — `internal/kanban/integration_lock.go`, which runs regardless of
  the flag (the flag gates only the deny, per the type doc).

The atomicity defect is in the RECORD layer, so the repair goes there and is
NOT flag-gated: a project with the deny off still writes the record, and a
double-hold corrupts coordination for everyone. Verdict: repair
`internal/kanban/integration_lock.go`; `internal/hook/**` and the config flag
are untouched.

## 3. Is it reproducible across processes?

Yes — the harness already exists and needs one new op, not new plumbing:

- `internal/kanban/kanban_helper_test.go` — child-process dispatcher keyed on
  `MOAI_KANBAN_HELPER` (ops at L23-111).
- `internal/kanban/board_lock_cross_test.go` — two goroutines fire two real
  child processes at the same tree and assert exactly-one-succeeds, with a
  positive control proving the refusal is conditional on the invariant rather
  than on concurrency itself (L41-105).

So the card is measurement-based, not code-reading-based. Bidirectional
regression is achievable as the lead required: RED = two live holders actually
observed on the unrepaired tree; GREEN = exactly one after the repair. A
one-sided test would not distinguish the repair from the lock being disabled.

Cleanup obligation: child processes are bounded by `t.Cleanup`-registered kills
or an external `timeout`; a trailing `kill` is not cleanup.

## 4. Substrate available for reuse (simplicity ladder step 2)

`internal/kanban/board_lock.go` + `board_lock_unix.go` (flock, non-blocking) +
`board_lock_windows.go` (atomic-create with a bounded transient retry and a
stale-clear path), themselves mirroring `internal/spec/lock*.go`. The
cross-process exclusion primitive exists in this package; nothing new is needed
at the substrate level.

Two shapes for the plan to weigh (the SPEC picks one):
- (a) reuse `AcquireBoardLock` as the mutation mutex — no new artifact, but
  couples two different scopes (whole-board card mutations vs the integration
  window) so unrelated operations block each other;
- (b) a dedicated short-lived mutation lock beside the record, reusing the same
  platform substrate — one more artifact, no cross-scope blocking.
Pre-flight leans (b); scope-correctness outranks artifact count here, and the
Windows stale-clear cost is the same either way.

## 5. Scope boundary vs t320

t320 (`moai integration release` returns ERROR with an empty message) sits in the
release OUTPUT layer — error classification and text — not in the record's
mutation path. Adjacent, non-conflicting. t336 MUST NOT change release's error
classification or message text; if t336 places release's read-modify-write under
the same mutation lock, t320's work lands on top unchanged. No merge conflict is
expected; the two touch different concerns in the same two functions.

## 6. What this pre-flight did NOT observe (gaps)

- No concurrent acquire was actually executed on this tree. The race is asserted
  from the code path (L186..L205 with no exclusion), which is a hypothesis until
  run-phase RED observes two live holders. Plan-phase must not claim it as
  measured.
- t320's empty-message defect was not reproduced here; the scope judgment above
  rests on reading the two call sites, not on running the failing case.
