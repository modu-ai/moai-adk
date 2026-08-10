---
id: SPEC-KANBAN-BOARD-001
title: "Research — measurements underlying the six-column kanban board model"
version: "0.4.0"
status: draft
created: 2026-08-10
updated: 2026-08-11
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, board, research, measurements, evidence"
tier: L
---

## §A. What this file is, and how to read it

Every measured fact `spec.md`, `plan.md`, and `design.md` rely on, recorded as **command → observed output → what it establishes**, so a later reader re-runs rather than re-derives. It is a Tier L artifact added at v0.2.0 with the promotion.

Two disciplines bind it. **Measurements are time-stamped by their tree, not trusted forever** — every count below is a fact about this repository at authoring time, and a run-phase that finds a different number trusts its own measurement and records the delta. And **an output is recorded even when it contradicts the expectation that motivated the command**; §H exists because two did.

Sections §B through §I were run at v0.2.0 authoring time on 2026-08-10; sections §L through §O were run at v0.3.0 authoring time on 2026-08-11 and are the measurements behind that revision's eight repairs. Two of them contradicted the expectation that motivated the command and are recorded as discrepancies (§L.2, §M) — the second half of §A's discipline, exercised twice more. Unless noted, the working directory is the primary checkout `/Users/goos/MoAI/moai-adk-go`; the worktree runs are from `/Users/goos/.moai/worktrees/kanban`. Git version: `git version 2.50.1 (Apple Git-155)`.

---

## §B. The git-common-dir behavior in both checkouts

This is the measurement that repaired `spec.md` §A.3 and `REQ-KB-005`. The primary case is recorded first because it is the one that breaks.

**Primary checkout** — `/Users/goos/MoAI/moai-adk-go`:

```
$ git rev-parse --git-common-dir
.git

$ git rev-parse --path-format=absolute --git-common-dir
/Users/goos/MoAI/moai-adk-go/.git

$ git rev-parse --absolute-git-dir
/Users/goos/MoAI/moai-adk-go/.git

$ git rev-parse --path-format=absolute --git-dir --git-common-dir
/Users/goos/MoAI/moai-adk-go/.git
/Users/goos/MoAI/moai-adk-go/.git
```

**Worktree** — `/Users/goos/.moai/worktrees/kanban`:

```
$ git rev-parse --git-dir
/Users/goos/MoAI/moai-adk-go/.git/worktrees/kanban

$ git rev-parse --git-common-dir
/Users/goos/MoAI/moai-adk-go/.git

$ git rev-parse --path-format=absolute --git-dir --git-common-dir
/Users/goos/MoAI/moai-adk-go/.git/worktrees/kanban
/Users/goos/MoAI/moai-adk-go/.git

$ git rev-parse --absolute-git-dir
/Users/goos/MoAI/moai-adk-go/.git/worktrees/kanban

$ git --version
git version 2.50.1 (Apple Git-155)
```

**What this establishes.** The bare `--git-common-dir` form returns a **repository-relative** path (`.git`) in the primary checkout and an absolute one from a worktree. A rule phrased as "the parent of the directory reported by `git rev-parse --git-common-dir`" therefore resolves correctly from every worktree and incorrectly from the primary checkout, where "the parent of `.git`" is not a path at all and the resolution collapses onto the current directory. `spec.md` §A.3 and `REQ-KB-005` carried exactly that phrasing at v0.1.0.

It also establishes the shape of the failure: it is **asymmetric**, and the asymmetry favors the wrong outcome. Every worktree exercise passes, and worktrees are where the feature is developed. This is why `AC-KB-002` requires the primary-checkout run as its positive control.

The two absolute forms agree in both checkouts, and the combined probe returns both paths in argument order — confirming that `--path-format=absolute` applies to every path flag that follows it, which is why one call serves both.

**Discrepancy note.** The equality of `--git-dir` and `--git-common-dir` in the primary-checkout combined probe is also the primary-versus-worktree discriminant `internal/hook/branch_guard.go` uses; the measurement above therefore doubles as confirmation that the discriminant still behaves as that file documents.

---

## §C. The existing resolution in service

```
$ grep -n 'path-format=absolute\|absolute-git-dir\|git-common-dir' internal/hook/branch_guard.go
5:// (a) the invocation occurs in the primary checkout (git-dir == git-common-dir),
27:// --path-format=absolute so the dispatcher's fallback path is exercised).
158:// (absolute git-dir == absolute git-common-dir). Returns (false, error) on any
162://   Primary path (git 2.31+, March 2021): --path-format=absolute for both
164://   --absolute-git-dir + cwd-normalized --git-common-dir. The fallback decision
175:	// is expensive (issue #1225 — TestBranchGuard_Latency). `--path-format=absolute`
176:	// applies to every following path flag, so --git-dir and --git-common-dir
178:	out, err := runGitRevParse(projectDir, "--path-format=absolute", "--git-dir", "--git-common-dir")
187:	// Fallback: --absolute-git-dir + cwd-normalized --git-common-dir. The bare
188:	// --git-common-dir form returns a repo-relative path (.git) on older git,
190:	absGitDir, err := runGitRevParse(projectDir, "--absolute-git-dir")
194:	relCommon, err := runGitRevParse(projectDir, "--git-common-dir")
```

The two anchors, verbatim:

```
$ sed -n '178p' internal/hook/branch_guard.go
	out, err := runGitRevParse(projectDir, "--path-format=absolute", "--git-dir", "--git-common-dir")

$ sed -n '190p' internal/hook/branch_guard.go
	absGitDir, err := runGitRevParse(projectDir, "--absolute-git-dir")
```

**What this establishes.** The correct resolution already exists in this repository, carries the git 2.31 flag floor and the older-git fallback, and documents at line 188 the very behavior §B measured — that the bare form returns a repo-relative `.git`. `REQ-KB-005` reuses it rather than re-deriving it, because the naive re-derivation *is* the defect §B records.

---

## §D. The gitignore lines that make a worktree's state private

```
$ grep -n '\.moai/state' .gitignore
205:# Nested SPEC-local .moai/state/ (e.g. .moai/specs/SPEC-*/.moai/state/) is caught by the ** glob.
207:**/.moai/state/
275:.moai/state/
```

**What this establishes.** `.moai/state/` is ignored at the repository root (line 275) and at every nested depth (line 207), with line 205 recording the intent. An ignored directory is not shared between checkouts, so each worktree carries its own copy of anything beneath it — the first of the three grounds on which `design.md` §B.1 rejects a per-tree board, and the reason `plan.md` §B.3 warns that no git-based check can observe the board's runtime state.

---

## §E. The tracked-file count under `.moai/specs/`

```
$ git ls-files .moai/specs/ | wc -l
    2528
```

**What this establishes.** `.moai/specs/` is git-tracked at scale, so a frontmatter write is a committed change. Combined with `enforce_admins: true` on this repository (main-direct push blocked), six column transitions per card would require six pull requests — the second ground of `design.md` §B.1.

**Discrepancy.** `spec.md` §A.3 and `plan.md` §E carried **2,536** at v0.1.0. The measurement above is **2,528**, taken from this worktree at v0.2.0 authoring time. Both files are corrected to 2,528. The delta is 8 files and does not move the argument — the ground is the order of magnitude, not the exact count — but the stale figure is replaced rather than left, since a number nobody re-measures is a number nobody can trust.

---

## §F. The status-value census

```
$ grep -rlE '^status: planned\s*$' .moai/specs/ | wc -l
      17

$ grep -rlE '^status: archived\s*$' .moai/specs/ | wc -l
      31
$ grep -rlE '^status: superseded\s*$' .moai/specs/ | wc -l
      10
$ grep -rlE '^status: rejected\s*$' .moai/specs/ | wc -l
       1

$ grep -rlE '^status: (archived|superseded|rejected)\s*$' .moai/specs/ | wc -l
      42

$ grep -rlE '^status: (planned|archived|superseded|rejected)\s*$' .moai/specs/ | wc -l
      59
```

**What this establishes.** `status: planned` is not hypothetical: **17** SPEC files carry it. It is a member of the canonical 8-value enum and appeared in **none** of the six rows of the v0.1.0 compatibility table, which made every one of those 17 SPECs illegal in all six columns and, by `REQ-KB-008`, permanently undispatchable — a table omission presenting as a board-wide failure. `spec.md` §A.2 and §A.4 now admit it in `backlog` and `plan`, and `AC-KB-021` decides it.

A further **42** files carry one of the three out-of-lifecycle terminals. These are handled by a different rule — no card is created for them at all (`spec.md` §A.2) — because `done` means *worked and finished* and a rejected SPEC was never worked. Stated as a property so their absence from the table is not read as the `planned` defect a second time.

**Discrepancy.** The revision brief anticipated **41** files across the four values. The measured figures are **42** for the three terminals and **59** for all four including `planned`; no run of these commands produces 41. All prose uses the measured values.

**A caveat on the census, recorded because it bounds how far these numbers may be pushed.** A frontmatter-anchored count over `.moai/specs/` also sweeps template scaffolds and prose that happen to begin a line with `status:`. An unanchored variant of the same census surfaces non-enum values — `backfilled`, `audit-ready`, `plan_ready`, `research-complete`, and several free-text lines — which are outside the canonical enum entirely and are *not* the board's problem: the board reads the enum, and a value outside it falls into `REQ-KB-008`'s illegal-pair path by construction. The counts above use the anchored form (`^status: <value>\s*$`) and are therefore a floor on real SPECs, not a ceiling on matching lines.

---

## §G. The two lock substrates, and the atomic-write primitive

### G.1 The in-process one, and its own comment forbidding promotion

```
$ sed -n '12,26p' internal/lockfile/lockfile_windows.go
// @MX:NOTE: [AUTO] Windows in-process-mutex limitation preserved verbatim from the
// pre-migration internal/cli/team_spawn_lock_windows.go (SPEC-AGENT-TEAM-RETIRE-001
// REQ-ATR-001 — behavior preservation is the contract; do NOT silently "upgrade"
// this to LockFileEx).
//
// fileLocks holds in-process mutexes keyed by absolute file path.
// Windows lacks portable advisory file locks (no fcntl/flock equivalent in stdlib),
// so we fall back to process-local mutexes. This means:
//   - Concurrent writes within the SAME process are serialized (safe for tests and
//     tmux teammates that run as separate processes within the same OS user session)
//   - Concurrent writes across DIFFERENT OS processes are NOT protected
//     (acceptable limitation: ClaimTask is primarily exercised by tmux-based team
//     workflows, which are macOS/Linux-only; Windows users run solo mode)
var fileLocks = map[string]*sync.Mutex{}
```

**What this establishes.** `internal/lockfile` is a `map[string]*sync.Mutex` on Windows whose own comment states that cross-process writes are **not** protected and explicitly forbids upgrading it. Sessions are distinct OS processes, so it would hold nothing in production while passing every same-process test. This is the measured reason `REQ-KB-019` selects the other family, and the reason `spec.md` §D.6 and `AC-KB-019` demand **separate processes** rather than goroutines — a goroutine test would pass against this implementation and measure only the harness.

### G.2 The cross-process one

```
$ sed -n '1,10p' internal/spec/lock.go
// Package spec — cross-platform per-SPEC file lock primitive.
//
// SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 REQ-LSG-010 + AC-LSG-010 + AC-LSG-021 (NFR-LSG-005):
// `moai spec close SPEC-XXX` acquires .moai/state/spec-close-<SPEC-ID>.lock to
// prevent concurrent close operations on the same SPEC. Lock scope is per-SPEC,
// not global (different SPECs may close concurrently).
//
// Cross-platform: Unix uses flock(2) advisory lock; Windows uses atomic-create-file
// pattern (O_CREATE|O_EXCL). Per CLAUDE.local.md §14, no naked syscall in body —
// the platform impl lives in lock_unix.go / lock_windows.go.
```

**What this establishes.** `internal/spec/lock.go` is genuinely cross-process on both platforms — `flock(2)` on Unix, `O_CREATE|O_EXCL` atomic create on Windows — with the platform bodies split out. It is the family `SPEC-KANBAN-WORKTREE-001` `REQ-KW-013` already selected, so `REQ-KB-019` reuses that selection rather than re-deciding it, and no third mechanism enters the repository.

Its scope is per-SPEC by construction, which is the point `design.md` §C.3 turns on: the *pattern* is reused, the *scope* is not. A board mutation takes a board-wide lock; a card's holder assignment keeps its card-scoped one.

### G.3 The atomic-write primitive and a working same-directory caller

```
$ sed -n '1,13p' internal/atomicfile/replace.go
// Package atomicfile provides a cross-platform atomic file-replacement
// primitive for the write-temp-then-rename idiom used across MoAI state files.
//
// On POSIX, rename(2) atomically replaces the destination even while other
// processes hold it open, so Replace is a direct os.Rename passthrough. On
// Windows the same call fails when any handle is still open on the
// destination — see replace_windows.go for the platform-specific handling.
//
// Callers keep owning temp-file creation and cleanup; Replace covers only the
// final rename step so it can be dropped into existing atomic writers without
// changing their error-handling shape.
package atomicfile
```

```
$ grep -n 'CreateTemp\|os.Rename' internal/verify/store.go
69:	tmp, err := os.CreateTemp(dir, ".snapshot-*.tmp")
89:	if err := os.Rename(tmpName, finalPath); err != nil {
```

with `dir` established four lines earlier as the target's own directory:

```
$ sed -n '61,64p' internal/verify/store.go
	dir := filepath.Join(projectRoot, SnapshotDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("snapshot mkdir %s: %w", dir, err)
	}
```

**What this establishes.** The write-temp-then-rename primitive already exists with the Windows edge case handled, and `internal/verify/store.go` `Save` is a working same-directory caller to copy: `os.CreateTemp(dir, ...)` creates the temp file **in the target's own directory**, and the rename follows. `REQ-KB-018` reuses both rather than authoring a third writer, and the same-directory property it mandates is the one this caller already demonstrates — which is also why `AC-KB-018` checks it statically, since a temp file elsewhere still renames and would still look correct.

---

## §H. The sole-writer absence

Measured across all three sibling SPECs at v0.1.0, before this revision:

```
$ grep -rc 'sole writer\|single writer' SPEC-KANBAN-BOARD-001 SPEC-KANBAN-WORKTREE-001 SPEC-KANBAN-BOOTSTRAP-001
SPEC-KANBAN-BOARD-001/progress.md:0
SPEC-KANBAN-BOARD-001/plan.md:0
SPEC-KANBAN-BOARD-001/acceptance.md:0
SPEC-KANBAN-BOARD-001/spec.md:0
SPEC-KANBAN-WORKTREE-001/plan.md:0
SPEC-KANBAN-WORKTREE-001/acceptance.md:0
SPEC-KANBAN-WORKTREE-001/spec.md:0
SPEC-KANBAN-WORKTREE-001/progress.md:0
SPEC-KANBAN-BOOTSTRAP-001/plan.md:0
SPEC-KANBAN-BOOTSTRAP-001/research.md:0
SPEC-KANBAN-BOOTSTRAP-001/spec.md:0
SPEC-KANBAN-BOOTSTRAP-001/progress.md:0
SPEC-KANBAN-BOOTSTRAP-001/design.md:0
SPEC-KANBAN-BOOTSTRAP-001/acceptance.md:0

$ grep -c 'atomic' SPEC-KANBAN-BOARD-001/spec.md
0
```

**What this establishes.** The ownership rule the predecessor's `REQ-KM-044` carried survives in **no** file of **any** of the three SPECs the split produced. Fourteen files, zero occurrences. Nor did any write-atomicity requirement exist — `atomic` appeared nowhere in `spec.md`. This is the F1 defect: `design.md` §B.1 rejected the `column:` storage mechanism, which says nothing about ownership, and ownership was deleted with it.

The sibling's card-scoped lock, for contrast, does exist and is unaffected:

```
$ grep -n 'REQ-KW-013' SPEC-KANBAN-WORKTREE-001/spec.md
195:**REQ-KW-013** — The system shall serialize holder mutation for a card — the read of the
holder, the decision, and the write — beneath an advisory file lock scoped to that card, …
225:- A third locking mechanism. REQ-KW-013 reuses an existing cross-process pattern; …
247:### D.2 The lock is checked for contention, not for existence (REQ-KW-013)
289:- `internal/spec/lock.go`, `lock_unix.go`, `lock_windows.go` — the cross-process per-scope
lock pattern REQ-KW-013 reuses.
```

**What this establishes.** A lock exists in the sibling family, and its scope is a **single card** — stated in its own text. It is correct for holder assignment and insufficient for a board mutation, which is the distinction `design.md` §C.3 draws and `AC-KB-019` decides by transitioning two **different** cards concurrently.

Re-running the first command after this revision returns non-zero counts for this SPEC's `spec.md`, `plan.md`, `acceptance.md`, and `design.md`; `plan.md` §C command 8 records it as the run-phase check.

---

## §I. The disclaimer pair that produced the gap

```
$ grep -n 'names no actor' SPEC-KANBAN-BOARD-001/spec.md          # at v0.1.0
199:- Who moves a card between columns, and by what message. This SPEC defines which moves
are legal; it names no actor and no transport.
```

and, in the sibling's §C:

```
- Deciding whether a card *may* move between columns. This SPEC decides who is told about
a card and by what message; the admission decision is the board's.
```

**What this establishes.** Each SPEC disclaimed the actor toward the other, and neither claimed it — being *told* about a card is not *writing* the card. The result was not a boundary but a hole with a boundary drawn around each side. `spec.md` §C is rewritten at v0.2.0 so this SPEC claims write authority (`REQ-KB-017`) while still naming no transport and electing no lead.

---

## §L. The shape of the symbol REQ-KB-005 reuses, and its line anchors

```
$ grep -n 'func isPrimaryCheckout' internal/hook/branch_guard.go
167:func isPrimaryCheckout(projectDir string) (bool, error) {

$ sed -n '178p;190p' internal/hook/branch_guard.go
	out, err := runGitRevParse(projectDir, "--path-format=absolute", "--git-dir", "--git-common-dir")
	absGitDir, err := runGitRevParse(projectDir, "--absolute-git-dir")
```

**What this establishes.** The resolution `REQ-KB-005` reuses is **unexported**, lives in package `hook`, and returns `(bool, error)` — a discriminant, not a path. So v0.2.0's instruction to "reuse rather than re-derive" mandated something the code shape does not permit: there is no exported path-resolving helper to call, and copying is what the same sentence forbids. `REQ-KB-005` now takes the extraction disposition of `SPEC-KANBAN-WORKTREE-001` `REQ-KW-018`.

The three line anchors above (167, 178, 190) are **recorded here and cited nowhere normative**, which is the v0.3.0 change. v0.2.0 put lines 178 and 190 into `REQ-KB-005`'s own text; all three were still correct when re-measured for this revision, and that is precisely the problem — a figure that happens to be right is not a figure anyone re-checks. Under this file's §A discipline they are re-measured at run-phase like every other number, and a requirement citing `isPrimaryCheckout` by name survives the next edit to that file.

### L.1 The extraction target, chosen by reading rather than by name

```
$ head -3 internal/core/git/doc.go
// Package git provides Git repository operations for MoAI-ADK.
//
// It implements three main interfaces:

$ grep -rn 'internal/hook\|internal/cli' internal/core/git/ | wc -l
       0
```

**What this establishes.** `internal/core/git`'s declared scope is git repository operations — which a git-directory resolution is — and it imports neither consumer, so the extraction creates no cycle. The target was selected by reading its `doc.go`, not by resemblance of name: the sibling's audit caught `SPEC-KANBAN-WORKTREE-001` naming `internal/worktree`, whose `doc.go` reads "working tree state guard primitives" and which that SPEC's §C excluded.

### L.2 Why the existing caller's green suite is not evidence for this consumer

`isPrimaryCheckout`'s body compares the two resolved paths for equality (`strings.TrimSpace(lines[0]) == strings.TrimSpace(lines[1])` on the primary path; `filepath.Clean(absGitDir) == filepath.Clean(absCommon)` on the fallback). An equality is **insensitive to an offset shared by both operands**: a normalization error shifting both identically leaves the discriminant correct and every existing test green. The board consumes the *parent of one operand* as a path, where the same error is a board root one directory wrong and silently so.

A note on scope, recorded because it corrects the premise this measurement was taken to confirm. The fallback's normalization — `filepath.Join(projectDir, relCommon)` when `relCommon` is not absolute — was expected to be unsound as a path resolver. It is not: `runGitRevParse` invokes `git -C projectDir`, so a relative result is relative to `projectDir` and the join is the correct anchor. What survives, and is the real finding, is the paragraph above: the caller's test surface **cannot detect** a resolver-level error even in principle, so the borrowed code carries no evidence about its correctness in this use. `AC-KB-002` therefore judges the resolved path against a separately recorded value.

### L.3 The fallback is never entered by the criterion, and the repository already knows how to force it

```
$ grep -n 'var execCommand' internal/hook/branch_guard.go
28:var execCommand = exec.Command

$ sed -n '24,27p' internal/hook/branch_guard.go
// execCommand is the package-level indirection over exec.Command. Tests inject
// a mock runner here (AC-WBG-005: simulates an older-git host rejecting
// --path-format=absolute so the dispatcher's fallback path is exercised).
// Restore via t.Cleanup in every test that swaps it.

$ sed -n '104,110p' internal/hook/branch_guard_test.go
// --path-format=absolute code path exits non-zero (older-git host), the
// dispatcher INSIDE isPrimaryCheckout falls back to --absolute-git-dir +
// cwd-normalized --git-common-dir. The mock is injected via the package-level
// execCommand indirection — direct invocation of the fallback is INSUFFICIENT
// (vacuous pass; bypasses the dispatcher).
```

**What this establishes.** On git 2.50.1 (§B) the primary probe always succeeds, so `AC-KB-002` as written at v0.2.0 exercised the fallback **zero times** while requiring it. The repository already carries the forcing mechanism and, in its own words, records that invoking the fallback directly is a *vacuous pass*. `AC-KB-002`'s second half reuses that mechanism rather than inventing one.

---

## §M. The stale-lock gap, and its platform asymmetry

```
$ sed -n '3,7p' internal/spec/lock_windows.go
// SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 — Windows per-SPEC close lock.
// Windows lacks fcntl-style advisory flock; we use atomic-create-file (O_CREATE|O_EXCL)
// per design.md §D.2 fallback. Stale lock detection (PID + timestamp embedded) is a
// post-MVP enhancement; M1 leaves stale-lock cleanup as a known-issue requiring
// manual `del .moai/state/spec-close-*.lock`.

$ grep -n 'Close releases the flock' internal/spec/lock_unix.go
21:	// Close releases the flock atomically.

$ ls .moai/state/spec-close-*.lock | wc -l
      14
```

**What this establishes.** The lock family `REQ-KB-019` reuses performs **no stale-lock detection** on its Windows substrate, by its own header, and the documented remedy is a manual `del`. A `lead` killed while holding the board-wide lock would therefore block every future board mutation with no defined operation able to change the answer — `spec.md` §A.6's brick, re-created by the revision that removed it. `REQ-KB-023` closes it, porting the shape of `SPEC-KANBAN-WORKTREE-001` `REQ-KW-014`.

**The gap is platform-asymmetric, and the asymmetry is the reason it survived review.** The Unix substrate holds `flock(2)` on an open descriptor and releases it on `Close`, which the kernel performs at process exit — so a killed holder there leaves a *file* and no *lock*. The fourteen orphaned zero-length `spec-close-*.lock` artifacts measured above, the oldest from May, are exactly that: inert. On Windows the substrate is `O_CREATE|O_EXCL` with no release path, so the artifact **is** the lock. An implementer developing on macOS cannot reproduce the defect, and a test suite run there passes.

**Discrepancy.** The revision brief stated that a killed `lead` "leaves an artifact blocking **every** future board mutation" without qualification. Measured, that is true on Windows and false on Unix. The requirement is unchanged — the Windows substrate is shipped and supported — but the unqualified claim is replaced by the measured one, and `AC-KB-023` now records the platform its result was obtained on, since a Unix-only pass satisfies its first observation trivially.

---

## §N. The path collision between the board and the session record

```
$ grep -n 'stateDirSegments' internal/factory/record.go
41:// stateDirSegments is the record's home beneath the project root, matching the
43:var stateDirSegments = []string{".moai", "state", "factory"}
106:	segments := append([]string{projectRoot}, stateDirSegments...)

$ ls .moai/state/ | grep -c '^kanban'
       0
```

**What this establishes.** `SPEC-KANBAN-RENAME-001` `REQ-KR-009` places the **session record** at `.moai/state/kanban/` "through the package's existing path-segment constant" — the constant above, joined beneath `projectRoot`, which inside a worktree is that worktree's own root. The record is deliberately per-tree: session-scoped, best-effort, and `REQ-KR-010` records that an orphan is inert.

v0.2.0 put the **board state** in the same directory while resolving it at the primary checkout. One name, two occupants, two contradictory resolution rules, and no SPEC in the family stating the coexistence — and an implementer following `REQ-KR-009`'s reuse instruction would have landed the board per-worktree, which is `plan.md` AP-1 reached by obeying a sibling requirement. The board moves to `.moai/state/kanban-board/`; the second command establishes that no entry of either name exists under `.moai/state/` today, so the new name collides with nothing.

---

## §O. The branch-side `status` read

```
$ git show origin/docs/SPEC-CODEX-PHASE2-001-fork-resolution:.moai/specs/SPEC-CODEX-PHASE2-001/spec.md | grep -m1 '^status:'
status: draft

$ git show spec-kanban:internal/factory/record.go | head -1
// Package factory implements the state record that carries Factory Mode's
```

Both runs from the primary checkout `/Users/goos/MoAI/moai-adk-go`. The first reads a remote-tracking ref; the second reads a **local** branch that a live worktree currently holds.

**What this establishes.** A card's `status` can be read from the card's branch with no checkout, no worktree, and — for a local branch — no fetch, because refs are shared across every checkout of one repository. This is the mechanism `REQ-KB-020` rests on.

**And the reason it is needed at all**, which no SPEC in the family had stated: `status` transitions are written on the card's branch (`spec.md` §A.2: `draft → in-progress` by manager-develop on the first run-phase commit), inside a worktree that `SPEC-KANBAN-WORKTREE-001` `REQ-KW-005` keeps for the card's `run`, `review`, and `sync` sessions and `REQ-KW-007` holds until **both** its pull requests have merged. So the primary checkout's copy reads `draft` for the whole interval the card sits in `run`, `(run, draft)` is outside `spec.md` §A.4's table, and `REQ-KB-008` refuses to dispatch — every card, on the normal path. A grep across all four SPECs in the family for a statement of the read location returns nothing.

### O.1 The declared dependency graph, and the cycle this consumption briefly created

```
$ for d in SPEC-KANBAN-BOARD-001 SPEC-KANBAN-WORKTREE-001 SPEC-KANBAN-BOOTSTRAP-001 SPEC-KANBAN-RENAME-001; do
    printf '%-28s ' "$d"; grep -m1 '^dependencies:' $d/spec.md || echo "(no dependencies: key)"; done

SPEC-KANBAN-BOARD-001        dependencies: [SPEC-KANBAN-RENAME-001]
SPEC-KANBAN-WORKTREE-001     dependencies: [SPEC-KANBAN-RENAME-001, SPEC-KANBAN-BOARD-001]
SPEC-KANBAN-BOOTSTRAP-001    dependencies: [SPEC-KANBAN-RENAME-001, SPEC-KANBAN-BOARD-001, SPEC-KANBAN-WORKTREE-001]
SPEC-KANBAN-RENAME-001       (no dependencies: key)
```

**What this establishes.** The four-SPEC graph is **acyclic**, with `SPEC-KANBAN-RENAME-001` as the sole root and the topological order `RENAME → BOARD → WORKTREE → BOOTSTRAP`. Six edges, no back-edge.

**Discrepancy, and it was this document's own.** Between the first and second drafts of v0.3.0 this SPEC's line read `dependencies: [SPEC-KANBAN-RENAME-001, SPEC-KANBAN-WORKTREE-001]`, which with the sibling's unchanged line formed a two-node cycle. The promotion was made because `REQ-KB-020` began consuming `REQ-KW-003`, on the reasoning that a requirement consuming a sibling's requirement is a dependency. That reasoning is wrong, and this SPEC had already applied the correct one two requirements earlier: `REQ-KB-017` consumes `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006` and that sibling stayed in `related_specs:` for exactly this reason.

The distinction the field turns on is what must **land** first:

| Consumption | Kind | Needs the sibling's code? |
|---|---|---|
| `REQ-KW-002` → this SPEC's holder field | landing | yes — until the field exists, `REQ-KW-011` has nothing to release |
| `REQ-KB-020` → `REQ-KW-003`'s identification rule | contract | no — the rule is readable from that document today |
| `REQ-KB-017` → `REQ-KS-006`'s role declaration | contract | no — same |

Only the first imposes an ordering, so only the first is declared. The cycle was found by the sibling's v0.3.0 author, who declined to resolve it by deleting the sibling's own edge — that edge records a real prerequisite — and recorded the analysis in its `spec.md` §A.4.0 and `research.md` §H.2, with `AC-KW-001` written to observe the mutual declaration and noted as **failing** at that time. Reversing the promotion here is what makes that observation pass; `AC-KB-022` carries the mirror observation on this side.

### O.2 How many branches a card actually has

§O establishes that a card's `status` **can** be read from a branch. It does not establish that a card has one branch, and it does not. Measured on the same tree:

```
$ git branch -a --format='%(refname:short)' | sed 's|^origin/||' | sort -u | wc -l
158
```

Matching each of those 158 deduplicated names against every SPEC identifier appearing on a branch, by `REQ-KW-003`'s exact-token rule — the segment after the type prefix begins with the identifier, and the next character is end-of-segment or a hyphen — gives **29 identifiers on branches, of which 3 carry two or more**:

```
SPEC-CODEX-PHASE2-001: 3   docs/…-fork-resolution, docs/…-m0-close, feat/…-run
SPEC-NAVIGATOR-SYNC-003: 3 plan/…, feat/…, sync/…
SPEC-PROJECT-NAVIGATOR-004: 2 fix/…, chore/…-sync-sha-backfill
```

Their `status` values, read the way §O reads them, and the two facts that decide which one the board should believe:

| Card | Branch `status` values | Live worktree | `origin/main` |
|---|---|---|---|
| `SPEC-CODEX-PHASE2-001` | `draft`, `draft`, `in-progress` | **yes** — `feat/…-run` at `~/.moai/worktrees/spec-codex-phase2` | `draft` |
| `SPEC-NAVIGATOR-SYNC-003` | `draft`, `in-progress`, `completed` | no | `completed` |
| `SPEC-PROJECT-NAVIGATOR-004` | `completed`, `completed` | no | `completed` |

**What this establishes.** Three things, and the third is the one the requirement rests on.

First, the disagreement is real and permanent: `SPEC-NAVIGATOR-SYNC-003` reads `draft`, `in-progress` and `completed` simultaneously, and nothing will reconcile them, because no rule in this family deletes a card's branches — `REQ-KW-004`'s prohibition is scoped to a *mismatched* branch during a creation refusal, and `REQ-KW-007` removes the worktree only. A grep across the sibling for a branch-deletion rule returns those two and nothing else.

Second, the type prefix is not a stage ladder. Only one of the three cards carries a `plan`/`feat`/`sync` triple; the others carry `docs`/`docs`/`feat` and `fix`/`chore`. A most-advanced-stage tiebreak has nothing to order two of the three cards by, quite apart from `REQ-KW-019` forbidding a tiebreak at all.

Third — and this is what makes worktree liveness the right selector rather than merely a workable one — **the one card whose `origin/main` value is stale is the one card with a live worktree**, and both cards whose `origin/main` value is current have none. `SPEC-CODEX-PHASE2-001` is mid-run: its tree holds `feat/…-run` at `in-progress` while `main` still reads `draft`. The other two are merged, and their `main` value equals their most advanced branch. The interval in which the primary checkout is wrong is exactly the interval in which a worktree exists to be observed, which is what `REQ-KB-020` now keys on.

One further shape is present and is the reason `REQ-KB-024` exists: a worktree can report no branch at all. Measured, `git worktree list` shows `.claude/worktrees/rc-build` in detached `HEAD`. An implementation that falls back to a branch **search** when the report comes back empty lands directly in the multiplicity above.

---

## §J. Out of Scope

### Out of Scope — what this file does not measure

- The runtime behavior of any board implementation. Nothing here exists yet; every measurement is of the repository, not of the feature.
- Performance of the lock or the atomic write. `plan.md` records that `internal/hook/branch_guard.go` carries a latency concern for its probe (line 175, issue #1225); no latency budget is set by this SPEC.
- The correctness of `SPEC-KANBAN-RENAME-001`'s rename mapping. `plan.md` §C command 1 measures only whether it landed.

### Out of Scope — measurements deferred to run-phase

- Each mirrored template pair's byte-identical-versus-sanitized classification. It is **time-varying** and is re-measured at run-phase per `spec.md` §D.9, never read from a document.
- The board root's resolved value on any machine other than this one. `AC-KB-002` compares against a value recorded at run-phase, not against the paths in §B.

---

## §K. Cross-references

- `spec.md` §A.2, §A.3, §A.4, §A.4a, §A.6, §A.7, §D.5 … §D.8, §D.11 … §D.13 — the requirements and verification surfaces these measurements support.
- `design.md` §B, §C, §D, §E.3 — the decisions each measurement forced.
- `plan.md` §B.4 … §B.8, §C — the run-phase re-measurement commands, including §C commands 9 through 13 added at v0.3.0.
- `acceptance.md` AC-KB-002, AC-KB-009, AC-KB-012, AC-KB-017 … AC-KB-024 — the criteria that consume them.
