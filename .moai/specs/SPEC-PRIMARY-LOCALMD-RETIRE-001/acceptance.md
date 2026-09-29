# acceptance.md — SPEC-PRIMARY-LOCALMD-RETIRE-001

> Stateless artifact (no `status:` frontmatter). Verification layer — every entry is a
> binary-testable `AC-PLR-*` Given/When/Then. Primary-side cells whose RED was measured by
> the delegating session (the worktree-session guard refuses cross-tree re-measurement from
> this worktree) carry that attribution explicitly and are re-executed at run entry (M1
> step 1) before any primary-side act — the RED must be re-observed on the tree it gates,
> in the run that acts on it.

## §D AC Matrix

| AC | Requirement(s) | Milestone | Priority | Verifies |
|----|----------------|-----------|----------|----------|
| AC-PLR-001 | REQ-PLR-003 | M1 | High | Pre-preserve capture exists, sha256 + provenance recorded |
| AC-PLR-002 | REQ-PLR-004 | M1 | High | Preserved copy byte-identical with pre-deletion working copy |
| AC-PLR-003 | REQ-PLR-005, REQ-PLR-006 | M1 | High | Exactly ONE new commit on local `main` deleting only `CLAUDE.local.md` (round-trip + staleness re-read exercised) |
| AC-PLR-004 | REQ-PLR-009 | M1 | High | Primary post-state: porcelain empty, file absent, ` M` marker gone |
| AC-PLR-005 | REQ-PLR-010 | M3 | High | Upward-traversal load source eliminated (mechanical absence required; probe optional) |
| AC-PLR-006 | REQ-PLR-011 | M2 | Medium | Develop-side §0.4 revision committed; §0.1–§0.3 byte-unchanged |
| AC-PLR-007 | REQ-PLR-001, REQ-PLR-002, REQ-PLR-012 | M1 pre / M3 | High | Ordering gates re-affirmed at run entry with recorded outputs |
| AC-PLR-008 | REQ-PLR-007, REQ-PLR-008 | M3 | Medium | No push and no working-copy-discard on the primary; divergence reported as `0 N` count on local `main` |

## §D.1 AC-PLR-001 — Pre-preserve capture (RED-now pinned)

- **Given** the primary checkout on branch `main` carries the working copy of
  `CLAUDE.local.md` (RED-now, delegation-measured this session at 68e37864a:
  `git -C /Users/goos/MoAI/moai-adk-go status --porcelain -- CLAUDE.local.md` → stdout
  ` M CLAUDE.local.md`, exit 0 — re-observed at run entry per M1; this agent could not
  independently re-measure, worktree-session guard refused cross-tree `git -C`).
- **When** the executing agent captures the working copy before any deletion and writes the
  sha256, date, card id, and source description to `progress.md` §E.2.
- **Then** `progress.md` contains a sha256 line (64 hex chars) for the captured copy and the
  four provenance fields, and the captured file exists at a recorded path.

## §D.2 AC-PLR-002 — Byte-identical preservation

- **Given** the AC-PLR-001 capture and the pre-deletion primary working copy.
- **When** `cmp` compares the capture against `.moai/state/retired/CLAUDE.local.md`
  (or `shasum -a 256` values are compared equal).
- **Then** exit 0 / identical digests. **RED-now**: `test -e <primary>/.moai/state/retired/CLAUDE.local.md`
  fails today (absent — observed: `.moai/state/retired` absent, this run, 68e37864a).
  **GREEN path**: M1 step "write preserved copy" flips it; passing output is exit 0 from
  `cmp` with both paths printed.

## §D.3 AC-PLR-003 — The ONE policy commit

- **Given** the primary checkout with the retirement executed.
- **When** `git -C <primary> log --oneline -1` and `git -C <primary> show --stat HEAD` run
  immediately after the commit (single invocations from the round-trip, not from the
  worktree).
- **Then** HEAD is exactly one new commit vs the pre-act HEAD; its stat shows
  `CLAUDE.local.md` deletion only (1 file changed, 1 deletion-class entry, no other paths);
  the message names `t1317` and `SPEC-PRIMARY-LOCALMD-RETIRE-001`. **RED-now**: primary
  HEAD at delegation-measured value with the file still tracked. **GREEN path**: M1's commit
  step flips it.

## §D.4 AC-PLR-004 — Primary post-state

- **Given** AC-PLR-003's commit landed.
- **When** three single-invocation checks run: (a) `git -C <primary> status --porcelain -- CLAUDE.local.md`
  → empty stdout, exit 0; (b) `test ! -e <primary>/CLAUDE.local.md` → exit 0;
  (c) the §D.1 command re-run → empty stdout.
- **Then** all three hold — the permanent ` M` marker and the working-copy load are gone.
  **RED-now**: (a) returns ` M CLAUDE.local.md`, (b) exit 1 (file present),
  (c) returns the marker line — delegation-measured pre-state. **GREEN path**: M1's commit.

## §D.5 AC-PLR-005 — Load elimination

- **Given** AC-PLR-004 holds.
- **When** the mechanical absence check re-runs from a fresh probe path
  (`test ! -e <primary>/CLAUDE.local.md` — the file upward traversal would load), AND the
  probe decision (session probe executed, or skipped with reason) is recorded in
  `progress.md` per REQ-PLR-010.
- **Then** the mechanical check passes and the progress.md record states which evidence
  shape was used. The baseline magnitude this eliminates: +20,855 tokens/session
  (t1279 probe 5: 62,209 vs probe 3: 41,354 first-turn input tokens,
  `.moai/reports/t1279/verdict.md` §① — carried citation; the run phase does NOT re-measure
  token deltas, it verifies the load source is gone).

## §D.6 AC-PLR-006 — Develop-side §0.4 refresh

- **Given** this worktree's branch (`WT-claudelocal-retire`) with the M2 edit.
- **When** (a-i) `grep -c "별도 전환 전까지 보존한다" AGENTS.local.md` → expect `0`
  (exit 1 — the un-revised preservation sentence is gone; RED-now: returns `1`, line 37,
  observed this run); (a-ii) `grep -c "t1317" AGENTS.local.md` → expect ≥ `1`
  (exit 0 — the completion record is present; RED-now: `0`); (b) `git diff <merge-base>..HEAD -- AGENTS.local.md`
  shows changes confined to the §0.4 region (single-sentence-scale, no §0.1–§0.3 hunks);
  (c) a commit on the branch carries the edit with card id in its message.
- **Then** (a-i), (a-ii), (b), and (c) all hold; the revision matches the file's Korean
  register and records that the 별도 전환 completed (card t1317 + the date). GREEN path: M2.

## §D.7 AC-PLR-007 — Ordering gates re-affirmed

- **Given** the run-phase entry (M1 pre-flight §C).
- **When** the two gate commands run and their outputs are recorded in `progress.md` §E.2:
  `grep '^status:' .moai/specs/SPEC-SESSION-DOUBLELOAD-001/spec.md` → `status: completed`;
  develop-side `git ls-files -- AGENTS.local.md CLAUDE.local.md` → `AGENTS.local.md` only.
- **Then** both outputs match the expected values. RED-now: both already PASS at
  68e37864a (observed this run) — this is a regression-guard criterion (the gates must be
  RE-observed in the run that acts, not carried), so it records the re-observation rather
  than claiming a flip. REQ-PLR-012 is this AC's failure path: **when** either gate reads
  anything other than its expected value, the run stops and a structured blocker report is
  returned instead of proceeding (evidenced by the §E.7 blocker report; see also the §D.11
  coverage rationale).

## §D.8 AC-PLR-008 — No push, no working-copy discard, divergence reported

- **Given** M1 and M2 are complete.
- **When** the completion report states: (a) no `git push` was executed by this lane;
  (b) local `main`'s unpushed-commit count vs `origin/main` (the `0 N` divergence-matrix
  row), read at completion time; (c) the new local-`main` commit SHA and WT-branch tip.
- **Then** the report carries all three; `git fetch`, `git push`, and the
  working-copy-discarding family — `git restore CLAUDE.local.md`, `git checkout -- CLAUDE.local.md`,
  `git reset`, `git stash` on the primary checkout (REQ-PLR-007) — never appear in the run
  transcript as lane-executed acts. RED-now: n/a (prohibition criterion — verified by
  absence of the acts plus the affirmative divergence report).

## §D.9 Edge cases

- **Foreign session touched the primary mid-act**: staleness rule (REQ-PLR-006) fires →
  stop, blocker report; the retirement retries only after the divergence is resolved.
- **Working copy differs from the develop-migrated content** (expected — the primary copy
  predates the §0 rewrite in places): irrelevant to correctness; the preservation is of the
  PRIMARY working copy as-is, not of canonical content. Note the difference in progress.md
  provenance if observed.
- **`.moai/state/retired/` path collision**: if a file already exists there (observed
  absent this run), suffix with the date (`CLAUDE.local.md.2026-09-29`) and record the
  chosen name in provenance.
- **Commit hook interference**: never `--no-verify`; a warn-only pre-commit result is
  normal (`manager-develop-prompt-template.md` §B9).

## §D.10 Quality gates

- TRUST 5 Tested/Readable/Unified/Secured/Trackable apply to the two write surfaces: the
  policy commit (Trackable — Conventional Commit, trailers, card id) and the §0.4 revision
  (Readable — Korean register parity with surrounding lines). No Go code changes → no
  coverage/lint gates. The develop-side batch's remote CI (`origin/develop`, lead-pushed)
  is the cross-tree verdict surface (git-flow lane protocol §4).

## §D.11 Definition of Done

All eight ACs PASS with baseline-attributed evidence in `progress.md` §E.2; the ONE policy
commit exists on local `main`; the §0.4 revision is committed on the WT branch; no
PRESERVE-list file modified; blocker count 0 (or all resolved); progress.md §E.3
audit-ready signal populated by manager-develop.

Coverage rationale (REQ-PLR-012, audit iter1 D1-3): REQ-PLR-012 is a failure-path criterion
with no happy-path flip — when any specified step fails it is evidenced by the §E.7 blocker
report (structured, per the subagent boundary); when no step fails it is evidenced by clean
§C pre-flight passage plus the blocker count above. Coverage is recorded here rather than
as a dedicated AC.
