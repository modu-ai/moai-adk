# SPEC-LOCAL-INSTR-RECEPTION-001 — acceptance criteria

> Authored by card t1290 (plan phase). All criteria are Given-When-Then. Release-blocking
> criteria carry the two-cell adoption discipline (`.claude/rules/moai/development/
> verification-completeness.md` §2): a RED-now cell (observed or, where plan-phase cannot
> execute it, first-act-of-M1 with the expectation stated and the measured precedent cited)
> and a green-path cell naming the milestone that flips it. 11 criteria; Tier L ceiling 25.

**Reading rule (inherited from the parent's audit history; extended to four signals per the
plan-audit D1 repair):** every probe leg is judged on four signals — **(1) the
file-presence precondition**: `git -C <probe-tree> show HEAD:AGENTS.local.md >/dev/null`
must exit 0 BEFORE the probe output is read; **(2)** the exit code; **(3)** the control
token; **(4)** the local token line. A leg whose tree lacks the file at HEAD is "measured
nothing" — the same standing as a leg whose control token is absent: never a reception
failure, never a pass. (Why the precondition comes first: with `worktree.baseRef` unset in
both settings files, launcher-created trees branch from `origin/HEAD` = `origin/develop`,
whose HEAD does not carry the fixture commit — probing such a tree prints control-present +
`LOCAL_AGENTS_TOKEN = NOT_PRESENT`, byte-identical to AC-LIR-005(b)'s expected negative, so
the raw output alone cannot separate a false reception failure from a real one.)

---

## §D.1 Acceptance criteria

### Reception gate (M1) — release-blocking

**AC-LIR-001 — `moai cc -w` reception.**
- **Given** a worktree created via `moai cc -w <probe-name>` **whose own HEAD carries the
  fixture** — a force-added `AGENTS.local.md` containing `LOCAL_AGENTS_TOKEN = <value>` at
  `git -C <probe-tree> show HEAD:AGENTS.local.md` (launcher-created trees branch from
  `origin/HEAD` = `origin/develop`, whose HEAD does not carry the fixture, so plan.md §D M1
  step 3 lands the fixture on the probe tree's branch before probing; the file-presence
  precondition — `git -C <probe-tree> show HEAD:AGENTS.local.md >/dev/null` exits 0 — passes
  before the probe output is read), **When** the Claude probe (plan.md §D M1 recipe) runs
  from that worktree's root, **Then** the probe exits 0, prints the `AGENTS.md` version line
  (control), and prints the `LOCAL_AGENTS_TOKEN` value — not `NOT_PRESENT`.
- RED-now: the same probe on today's geometry (no file) prints `NOT_PRESENT` for the token
  with the control present. Measured precedent: §M1a, `LOCAL_AGENTS_TOKEN = NOT_PRESENT`
  (`claude 2.1.283`, real nested worktree, gitignored fixture); static facts measured at
  `514ac7abe` (file absent, `.gitignore:275`). The leg's own RED is M1 step 1's first
  observation, recorded in progress.md §E.2 before the fixture lands.
- Green path: M1 steps 2-3 (fixture + leg).

**AC-LIR-002 — `claude --worktree` (native) reception.**
- **Given** a worktree created via Claude Code's native `claude --worktree` entry **whose
  own HEAD carries the fixture** — `git -C <probe-tree> show HEAD:AGENTS.local.md` resolves
  (fixture landed per plan.md §D M1 step 3; file-presence precondition passes before
  reading), **When** the probe runs from that worktree's root, **Then** the same four-signal
  outcome as AC-LIR-001 holds.
- RED-now / green path: same shape as AC-LIR-001. `.worktreeinclude`'s header claims this path
  copies gitignored files at creation; the force-tracked geometry makes that claim inert here
  (tracked files are never duplicated) — the leg measures the tracked-file-plus-import path on
  a native-created tree.

**AC-LIR-003 — `EnterWorktree` re-entry reception.**
- **Given** the AC-LIR-001 probe tree **whose HEAD carries the fixture** (the file-presence
  precondition passed in that leg), **When** the session re-enters it via
  `EnterWorktree(<path>)` and the probe runs, **Then** the same four-signal outcome holds.
  This leg is why the design prefers git-tracked delivery over creation-time copy: re-entry
  into an existing tree gets no copy step, only git's own checkout.

**AC-LIR-004 — `moai codex -w` child reception.**
- **Given** the same fixture, **When** a `moai codex -w` child session is launched and the
  Codex probe (plan.md §D M1, Codex recipe) runs in it, **Then** the token value appears in
  the child's model-visible input (the child answers the token question, or the recorded
  `developer_instructions` payload carries it), with the contract control present in the same
  measurement.
- RED-now precedent: §M1b measured Codex discovering zero local files by filename
  (`LOCAL count=0`, `codex-cli 0.157.0`) — the launcher path is the only Codex reception
  route, which is why the §8 claim is measured rather than cited.

**AC-LIR-005 — the probe can fail (negative control).**
- **Given** the probe has gone green on AC-LIR-001~004, **When** the probe runs (a) on the
  pre-fixture geometry (M1 step 1's recorded observation) and (b) in a worktree created from a
  base commit without the file — its HEAD lacks the fixture, the deliberate opposite of
  AC-LIR-001~003's file-presence precondition (signal 1) — **Then** in both it prints
  `NOT_PRESENT` for the local token while the control reads present. A probe green that has
  never seen its own red is uninterpreted output.

### Vacuous-green guard

**AC-LIR-006 — the migration may not pass while reception is unproven.**
- **Given** the close review of this SPEC, **When** progress.md §E.2 marks `AC-IFU-007` or
  `AC-IFU-024` PASS, **Then** the same record cites the evidence paths of `AC-LIR-001`~
  `AC-LIR-005`, each marked PASS with its probe output. A close record missing any of the five
  is invalid regardless of every other criterion's state. This is the criterion form of
  REQ-LIR-004 — the duty the parent kept in a plan bullet, promoted to something the close
  must read.

### Migration (M2) — moved from the parent, re-sited on the committed tree

**AC-IFU-007 — migrated content under the cap, measured from the committed tree.**
- **Given** M1 is fully green and `moai migrate local-instructions` has run on the
  develop-committed copy, **When** the milestone reads
  `git show HEAD:AGENTS.local.md | wc -m`, **Then** the value is at most 39,999, and the
  before-value (`git show develop:CLAUDE.local.md | wc -m`) was re-measured at this milestone
  and recorded beside it. The `git show` source is the iter4 D1 fix: the working copy is
  gitignored and passes vacuously while the merge lands nothing.
- RED-now: `git show HEAD:AGENTS.local.md | wc -m` fails today — the file is absent from
  `HEAD` (measured at `514ac7abe`: `ls AGENTS.local.md` → No such file; implied by
  `git ls-files` showing only `CLAUDE.local.md`, iter4 §2 D1). Red for the right reason: the
  migration has not run.
- Green path: M2. Baseline expectation (not a bound): before = 45,810 measured at `514ac7abe`
  (`git show develop:…` and `git show origin/develop:…` agree); re-measured at M2.

**AC-IFU-024 — §0 names the new file, read from the committed tree.**
- **Given** the migration has run, **When**
  `git show HEAD:AGENTS.local.md | grep -n -C1 'AGENTS.local.md'` runs, **Then** it matches in
  §0 (the canonical-copy discriminator names its own file). `git show` fails loudly if the
  file was never committed — the mutant AC-LIR-006 guards against (working-copy pass,
  committed-tree absence) is unreachable through this command.

**AC-LIR-007 — exactly one local instruction file in the committed tree.**
- **Given** the migration is complete, **When** `git ls-files AGENTS.local.md
  CLAUDE.local.md` runs, **Then** it lists `AGENTS.local.md` only, and
  `git show HEAD:CLAUDE.local.md` exits non-zero. RED-now: today's inverse —
  `git ls-files` lists `CLAUDE.local.md` only (iter4 §2 D1, measured). Green path: M2.

**AC-LIR-008 — operational procedure relocated to `.moai/docs/`.**
- **Given** the migration, **When** the relocated procedure file under `.moai/docs/` is read,
  **Then** it exists, carries the operational sections (gitflow chain, LSEL drain, Jev,
  kickoff autonomy — the procedure content of the migrated file), and the rules content that
  remains in `AGENTS.local.md` is unaltered in substance (§D Out of Scope: no doctrine
  re-decided). Judged by section-presence diff against the pre-migration file recorded at the
  milestone, not by character identity of relocated prose.

### Reception survival and documentation (M2/M3)

**AC-LIR-009 — reception survives the content swap; docs match the measured geometry.**
- **Given** M2 has replaced the fixture with the migrated content, **When** the M1 probe
  re-runs on one Claude-side leg and the documentation statement is read, **Then** the probe
  still reports the token (from the real file), and the worktree-reception sentences in
  `AGENTS.md` §8 and `CLAUDE.md` §18 no longer state the negative for the landed geometry —
  both the template copies and the root copies updated in the same change (Template-First;
  `make build` run; wording carries no card ids, SPEC ids, or internal dates).
- RED-now: `AGENTS.md:262` and `CLAUDE.md` §18 currently state the import-skip negative
  (measured at `514ac7abe`). Green path: M3.

---

## §D.2 Traceability

| Requirement | Criteria |
|---|---|
| REQ-LIR-001 | AC-LIR-001, AC-LIR-002, AC-LIR-003 |
| REQ-LIR-002 | AC-LIR-004 |
| REQ-LIR-003 | AC-LIR-005 |
| REQ-LIR-004 | AC-LIR-006 |
| REQ-IFU-021 | AC-IFU-007, AC-LIR-008 |
| REQ-IFU-022 | AC-IFU-024 |
| REQ-LIR-005 | AC-LIR-007 |
| REQ-LIR-006 | AC-LIR-009 |

Diff command (empty at close): none — the table above is hand-checked one-to-one; the close
re-reads it against §D.1 headings.

## §D.3 Close-time duties

1. Read the `no tests to run` marker on any selector-driven run (none planned; the discipline
   applies if one appears).
2. Re-run the `git ls-files` postcondition (AC-LIR-007) on the merge head, not the lane head —
   the absorb can change what develop carries.
3. Re-measure the before-value at M2 and record it beside the after-value (AC-IFU-007); never
   cite this document's 45,810 as the milestone's number.
4. Retain all probe outputs under `.moai/reports/t1290/` with the entry path, cwd, exit code,
   and verbatim output per leg — the reception evidence outlives the milestone and is what
   AC-LIR-006's close read consumes.
