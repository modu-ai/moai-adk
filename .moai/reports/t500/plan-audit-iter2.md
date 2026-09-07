# SPEC Review Report: SPEC-CODEX-E2E-GUARD-001

Iteration: 2/2 (Tier M ceiling — final iteration)
Verdict: PASS
Overall Score: 1.00 (Tier M PASS threshold 0.80 — exceeded)

Auditor: plan-auditor (independent). Tree: `ace1c5440` (unchanged — SPEC artifacts remain
untracked, so the tree SHA did not move during repair), branch `WT-codex-e2e-guard`,
worktree `.claude/worktrees/t500`. Re-audit scope per the Retry Loop Contract: the iter-1
defect delta (D1-D4) + regression check + consistency sweep of the edited artifacts. M1
Context Isolation honored.

## Regression Check (Iteration 2)

Defects from iteration 1 (report: `.moai/reports/t500/plan-audit-iter1.md`):

- D1 (blocking) RED-now cells missing + non-discriminating extension-AC evidence —
  **RESOLVED**: acceptance.md:3-8 now states the two-cell discipline with the governing rule
  citation (verification-completeness.md §2); all six must-pass ACs carry RED-now cells with
  the four required elements (single-invocation read-only command / verbatim stdout / exit
  code as its own field / SHA pin — document-level pin at acceptance.md:8 plus per-cell
  "adopted at `ace1c5440`"); AC-CEG-005/006/007 green paths are now self-describing
  (`len(guardFiles) != 12` → `t.Fatalf`, acceptance.md:98, :123, :140; plan.md M3.1/M4.1)
  and AC-CEG-007 adds the positive-control canary (`"git"` literal at
  `codex_review_gate.go:129`, acceptance.md:140-142); AC-CEG-003's self-RED disposition is
  recorded with the half-paired baseline as the observable flipped state (acceptance.md:42-53)
  — coherent under §2: the mutant procedure is deterministic and re-executable, and the
  auditor-recommended "do not execute the mutant at plan phase" is honored. §D.3 gained the
  "green whose RED-now cell is missing is an unadopted criterion, not a pass" clause
  (acceptance.md:184-185) and the self-describing-green requirement (:186-187).
- D2 (optional) REQ-CEG-007 compound — **RESOLVED**: split into REQ-CEG-007 (Ubiquitous),
  REQ-CEG-009 (Where-gate), REQ-CEG-010 (Ubiquitous + shall-not) — spec.md:154-168, one
  pattern per REQ. §E table updated (spec.md:196).
- D3 (optional) undeclared `related_specs` — **RESOLVED**: field removed from frontmatter
  (spec.md:1-15 now carries exactly the 12 canonical fields + `tier: M`); cross-refs live in
  plan.md §H (verified: all four SPEC-CODEX-* + SPEC-INIT-HARNESS-PROMPT-001 at plan.md:171-177).
- D4 (optional) "an CheckOK" — **RESOLVED**: "a `CheckOK`" at spec.md:137 and
  acceptance.md:35. The only remaining literal matches are inside the HISTORY repair row
  (spec.md:275), which quotes what was changed — legitimate historical record, not residue.

## Repair verification (this run, this tree)

Commands re-executed by the auditor, with observed output:

| Cell | Command | Observed this run | Matches recorded cell |
|---|---|---|---|
| AC-CEG-001 RED-now | `go test ./internal/cli/ -run TestRunInit_ThenDoctorCodexWiringHealthy -count=1 -timeout 600s` | `ok … 0.846s [no tests to run]`, exit 0 | YES (duration varies; `[no tests to run]` + exit 0 exact) |
| AC-CEG-003 baseline | `go test ./internal/codexwiring/ -run TestStatusLineDefaultSubsetOfAllowlist -count=1` | `ok … 0.576s`, exit 0 | YES (duration varies) |
| AC-CEG-005 RED-now | `grep -n "codexSpecFiles = " internal/cli/codex_launcher_guards_test.go` | `30:var codexSpecFiles = …{2 files}`, exit 0 | YES byte-exact |
| AC-CEG-006 RED-now | `grep -n "req.Program" internal/cli/codex_launcher_guards_test.go` | lines 147/170/171, exit 0 | YES byte-exact |
| AC-CEG-007 RED-now | `grep -n "walk(codexCmd)" internal/cli/codex_launcher_guards_test.go` | `74:	walk(codexCmd)`, exit 0 | YES byte-exact |
| 12-count cross-check | `ls internal/cli/*codex*.go \| grep -v _test \| wc -l` | `12` | YES |
| Spec lint | `moai spec lint .moai/specs/SPEC-CODEX-E2E-GUARD-001/spec.md` | 0 findings, exit 0 | YES (auditor's own run) |

## Must-Pass re-checks (delta scope)

- MP-1: REQ-CEG-001..010 sequential, no gaps, no duplicates (spec.md §C sweep: all ten
  tokens present, unique). 10 ≤ Tier M ceiling 16. AC-CEG-001..007 unchanged, sequential.
- MP-2 (requirement layer, spec.md §C only): REQ-CEG-007/009/010 each match exactly one
  GEARS pattern; no informal language introduced. PASS.
- MP-3: frontmatter 12/12 canonical fields + `tier: M`; no snake_case aliases; undeclared
  field gone. Lint 0 findings (this run). PASS.
- MP-4: N/A (single-language SPEC) — unchanged. MP-5 D7: reference set unchanged, all
  `status: completed` (iter-1 measurement, references untouched by the repair; plan.md §H
  verified intact). MP-6 D8: syscall context unchanged. MP-7: no `[NEEDS CLARIFICATION]`
  markers in plan.md; research.md still absent (Tier M).

## Consistency sweep (edited artifacts)

- Traceability: 10/10 REQs covered (AC-CEG-007 covers 007/009/010 per the updated §E row);
  7/7 ACs trace to existing REQs; no orphans.
- plan.md M3.1/M4.1/M5.2 carry the length-pin + canary + swept-count evidence obligations,
  matching acceptance.md's cells — no dangling claims between plan and acceptance.
- HISTORY row appended (spec.md:275) accurately describing the D1-D4 repairs.
- Edit-mismatch residue check (reporter's disclosed transient): full re-read of spec.md
  shows intact structure — §A through HISTORY well-formed, all 10 REQ entries clean, tables
  well-formed, no duplicated fragments around REQ-CEG-002 (:135-138). No residue found.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 1.0 | 1.0 band | §B definitions intact; REQ split reduced per-entry density; one interpretation per requirement |
| Completeness | 1.0 | 1.0 band | Adoption layer now complete (two cells × six must-pass ACs, four elements each); all sections + 12/12 frontmatter; HISTORY repair row |
| Testability | 1.0 | 1.0 band | Self-describing greens (len==12 t.Fatalf + canary); binary ACs; §D.2.6 added the swept-count lifecycle edge case |
| Traceability | 1.0 | 1.0 band | 10/10 REQs ↔ 7 ACs, §E table updated for the split |

## Defects Found

D5. AC-CEG-004 RED-now stdout line numbers stale — acceptance.md:74-76 — The recorded
   verbatim grep output (`196:## §F …` / `198:### §F.1 …` / `213:### §F.2 …`) predates the
   repair's own insertion of REQ-CEG-009/010; re-running the cell's command today yields
   `201` / `203` / `218` (auditor re-ran it, exit 0). The tree-SHA pin does not disambiguate
   because the SPEC artifacts are untracked — the artifact moved under an unmoving SHA. The
   cell's substance is verified TRUE this run (§F.1/§F.2 exist; the sync verdict remains the
   unflipped state), and the command is re-runnable, so no adoption or red/green semantics
   are affected. — Severity: minor — Class: optional — Required fix: re-run the cell's grep
   and refresh the three line numbers. NOTE for the orchestrator: applying this fix edits
   acceptance.md, which breaks skip-eligibility condition 3 (artifact-hash) and forces a
   Phase-1 re-audit on the next run — the honest alternatives are (a) record this verdict
   as final with D5 as documented cosmetic debt, or (b) apply the fix and accept the
   re-audit. Either is defensible; (a) is cheaper.

No other defects found in the delta scope.

## Fail-open notes

- Cross-model second opinion: not re-invoked this iteration (iter-1 attempt returned
  `participant_count: 0` — no backend verdicts; recorded inconclusive there). Verdict
  authority: this in-session audit.

## Skip-eligibility status (Phase-1 Plan Audit Gate contract)

1. **Verdict is `PASS`** — YES (this iteration-2 verdict is the review stream's
   final-iteration verdict).
2. **Overall score ≥ Tier M threshold** — YES (1.00 ≥ 0.80; `SkipEligibleByScore` satisfied).
3. **Artifact-hash unchanged since this verdict** — TRUE AT RECORDING TIME. The three plan
   artifacts (spec.md / plan.md / acceptance.md) are uncommitted in the worktree as of this
   report; any subsequent edit — including the optional D5 line-number refresh — changes the
   hash and voids this condition. Record the skip decision (with the three conditions) in the
   run-phase delegation prompt per the contract, or fix D5 first and accept the re-audit.

Skip-eligibility governs Phase-1 verdict re-execution ONLY; Implementation Kickoff Approval
remains mandatory and is never auto-bypassed.
