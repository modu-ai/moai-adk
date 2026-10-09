# design.md — SPEC-MEMORY-FOLD-RENAME-RACE-001

Mechanism rationale for the fold store lock. This file records WHY the chosen mechanism is the lock shape below and what was rejected — the decision rows live in decision-index.md (Q1..Q5); this file carries the engineering reasoning.

## The chosen mechanism

A per-store advisory lock file, held for the whole `applyFold` span:

- **File**: `<store-dir>/.moai-fold.lock`, mode 0644, created on first acquire, never removed on release.
- **POSIX**: `unix.Open(path, O_CREAT|O_RDWR|O_CLOEXEC, 0o644)` then `unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB)` with a bounded retry loop (plan.md D-2). Pattern copied from `internal/sessionmsg/lock_unix.go` — including the `unacquiredFD = -1` sentinel lesson (fd 0 is a valid descriptor; a 0 sentinel turns `release()` into a silent no-op and leaks the lock for the process lifetime).
- **Windows**: `LockFileEx` with `LOCKFILE_EXCLUSIVE_LOCK|LOCKFILE_FAIL_IMMEDIATELY` on a byte-range of the lock file; pattern from `internal/sessionmsg/lock_windows.go`.
- **Span**: acquired at the top of `applyFold` (before the first `checkFoldUnchanged`), released by `defer`. One acquisition covers both renames (archive append; MEMORY.md rewrite) — premise P2: `applyFold` is the single write chokepoint.
- **One-line rationale**: an advisory flock is released by the kernel on process death, so the design needs no stale-lock sweep, no PID-file liveness probe, and no cleanup path — the two failure modes an O_EXCL claim file adds are absent by construction.

## Why the whole-apply span and not per-rename

A per-rename lock (acquire around each check→rename pair) would close the window too, but it re-acquires twice per apply, interleaves two folds' plan/re-check phases on one store (each still refusing on the byte comparison — correct but noisy), and gives fold-vs-fold no serialization beyond the bare window. The whole-apply span is one acquisition, strictly stronger (two folds on one store serialize entirely — the mechanism card t1595 item 8 independently needs), and simpler to reason about. Cost: slightly longer hold (the effective-state re-scan runs under the lock) — bounded by the same store's file count, and the close-path bound still expires the step (REQ-MRR-006).

## Rejected alternatives

| Alternative | Why rejected |
|---|---|
| `renameat2(RENAME_REPLACE)` / compare-and-swap rename | Not portable — Linux-only, absent on darwin and Windows; the repo ships both. No POSIX compare-and-swap rename exists. |
| Re-running `checkFoldUnchanged` "closer" to the rename | Narrows the window; never closes it. The card's five observations are exactly the residual of this approach. |
| O_EXCL claim-stamp file (`internal/verify/claim_lock.go` pattern) | Works, but a crashed holder leaves a stale claim that wedges every later fold until a staleness sweep (age/PID heuristics — more code, more failure modes) removes it. flock's kernel-owned lifetime is the simpler contract. Revisit only if a flock-free platform ever matters. |
| In-process `sync.Mutex` only | The defect class is CROSS-PROCESS (two `moai` invocations — lane sessions). A mutex never sees the other process. |
| New third-party dependency (e.g. go-flock) | The standard library + already-present `golang.org/x/sys` covers both platforms; the ladder puts an installed dependency's helper below reuse of the repo's own pattern. `go.mod` stays unchanged (REQ-MRR-007). |
| Blocking `LOCK_EX` (wait indefinitely) | An unbounded user-facing wait on the verb path, and a parked goroutine the close-path bound cannot interrupt (the step cannot observe `writesForbidden` while parked inside flock). The bounded NB-retry keeps every wait bounded and every refusal explicit. |

## Interaction with the existing check sequence

The lock is strictly additive: acquisition precedes the first `checkFoldUnchanged`; nothing between the existing checks moves (plan.md D-1). The `orderProbe` regression (`TestReviewArchiveUpdateDuringEffectiveScan`) observes the identical `effective-start` → `bytes-done` sequence after the fix — a deliberate regression guard that the fix did not reorder the MFB checks.

## Limits (restated, not re-decided)

- Non-cooperating writers (editors) keep the byte-comparison narrowing; the lock cannot protect what never acquires it (spec.md Out of Scope).
- Other memory verbs adopt the lock in a follow-up (decision-index Q5); today only fold applies cooperate, which is exactly the observed defect class.

## Promotion path

The lock is an unexported `internal/cli` type following the repo's copy-per-package precedent (`internal/session` → `internal/sessionmsg` → now `internal/cli`). If t1595 or another subsystem adopts it, promote the body to `internal/filelock/` at that time — a mechanical move, planned as contingency (i) in plan.md §A.2, not now.
