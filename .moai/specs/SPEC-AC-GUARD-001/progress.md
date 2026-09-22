# Progress — SPEC-AC-GUARD-001

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifacts authored 2026-09-22 by manager-spec (card t1067), Tier M, 4 files (spec.md, plan.md, acceptance.md, progress.md), worktree `.claude/worktrees/t1067`, branch `WT-guard-rejected-ac`, base `0314801c2`.
- SPEC ID pre-write regex check: PASS (verbatim `PASS` output cited in the authoring session).
- Frontmatter validated against spec-frontmatter-schema.md § Canonical 12 Required Fields; `phase` carries release target `v3.1.0` (no lifecycle-stage token).
- Calibration dataset (spec.md §2) carried verbatim from the kickoff measurement; census predicates derive from it.
- Open items: 2 x [NEEDS CLARIFICATION] markers in plan.md §F (M3 exit-capture-only disposition rule; M4 skill-pointer scope) — to be resolved via orchestrator AskUserQuestion before Implementation Kickoff Approval.
- Note: the card dispatch referenced "progress.md §F.1"; the canonical plan-phase skeleton is this §E.1-§E.4 structure per the era-classification engine contract (§F.* markers misclassify eras) — skeleton emitted accordingly.
- Iteration-1 plan-audit (2026-09-22): FAIL, overall 0.75 harmonic (< Tier M 0.80); report `.moai/reports/t1067/plan-audit-iter1.md`. Repairs D1-D7 applied in one pass (spec.md 0.1.1 HISTORY row); iteration-2 re-audit is delta-scoped to D1-D7.
- plan_status: audit-ready
- plan_complete_at: 2026-09-22
- plan_status note: `audit-ready` CONFIRMED — iter-2 re-audit PASS 0.93 (report: .moai/reports/t1067/plan-audit-iter2.md); both [NEEDS CLARIFICATION] markers RESOLVED by operator 2026-09-22 (per-file leave verification; M4 pointer in-SPEC); Implementation Kickoff Approval granted same day.

## §E.2 Run-phase Evidence

### M1 — Verified full-corpus block-level census (2026-09-22, census base `873a19710c`)

- Disposition table: `.moai/reports/t1067/census-20260922.md` (working copy; §D.1-final home = this path per acceptance §D.1's "or `.moai/reports/t1067/`" clause).
- Headline: corpus 795 files; sweep union population **111** files (A 60 / B 26 / C 25 / D 55); dispositions **rewrite 52 · leave-verified 53 · leave-parse 6 · unprobeable 0** (two initial rewrite dispositions — INTEGRATION-LOCK-TARGET-SOURCE, UPDATE-CI-GUARD — were reversed to leave-verified on verbatim re-probe; the census report §4.1 records the correction).
- Fresh guard samples: ≥2 per family recorded (families A/B plus NEW families C-H the census surfaced: nested-git-`$()`-as-argument, git-piped-into-loop, loop-over-git-substitution, subshell-with-git, test-with-embedded-git-`$()`, assignment-`&&`-expansion, non-git-command-with-git-`$()`-argument, env-prefixed-git-`$()`).
- Measured passing shapes extend the calibration set: counter-terminated `$()` assignments (`wc -l`, `grep -c`, `awk`, `wc||echo`), echo-embedded `$()` (any terminator), comments/patterns naming git.
- Census tooling refusals: 3 recorded, each re-expressed as plain single invocations (report §3) — never bypassed.
- Proxy retired: AC-002 note records the numeric coincidence (verified census 55 vs retired proxy 55) with both predicates named (report §2).
- Closure: `comm` checks — sweeps ⊆ union (0 outside), refused ∩ union = 55, union − refused = 56 (report §5).

### M3 — Corpus disposition: verified-refusing blocks rewritten (2026-09-22)

- All 52 rewrite-dispositioned files' refusing blocks restated per the M2 convention (three-dot ranges, base recorded on its own line, `$?`-capture lines, redirect-to-file capture, git-free loops, `<PLACEHOLDER>` recorded-value convention per the WEB-CODEX-PANEL precedent); verification semantics preserved (observed stdout, exit code as own field, recorded base/tree values).
- AC-004: every rewritten file's decisive command re-executed in this worktree-isolated session in six grouped invocations, all without refusal (evidence: census report §6; HEAD at re-execution `dbe1a6941`).
- AC-005: leave files zero-diff verified mechanically — every git-modified acceptance.md ∈ rewrite set (`comm` check at commit time).
- AC-006: no verification relocated into a script file; where a loop needed per-item git calls, the loop was reduced to a documented per-item plain-command procedure (e.g. SPEC-ASTGREP-EDIT-001, SPEC-DOCSITE-E2E-001, SPEC-V3R6-RULES-PATH-SCOPE-001, SPEC-PHASE-FIELD-VALIDATION-001).
- Two census dispositions corrected on verbatim re-probe (INTEGRATION-LOCK-TARGET-SOURCE, UPDATE-CI-GUARD: actual compositions are guard-executable; replica over-mutation disclosed in census §4.1).
- New boundary observations recorded during M3: non-git `$()` in printf args passes; computed command NAME (`"$REPO/bin/moai"`) refuses even git-free; git command with `$(cat …)` argument refuses; `if ! git … | grep -q` passes; plain-grep-terminated `$()` assignment + later expansion passes; literal-path git chains (no `$()`/loops) pass even as long `&&` bundles.

## AC matrix (final, 2026-09-22)

| AC | Status | Evidence (command → observed output, this run, tree `a2e6763f4` unless noted) |
|----|--------|-------------------------------------------------------------------------------|
| AC-001 census + four-field evidence | PASS | census report §1-§4 (111 union rows; sweeps 60/26/25/55; `ls .moai/specs/*/acceptance.md \| wc -l` → `795`) |
| AC-002 census governs; proxy retired | PASS | report §2 AC-002 note (numeric-coincidence disclosure; proxy labeled by predicate+SHA wherever quoted) |
| AC-003 boundary map + authoring rule, C1/C2 parity, build green | PASS | C1/C2 `diff` → only 2 internal-trace hunks; C2 forbidden-class scan → 0 hits in added text; `make build` → exit 0 (twice: M2, M4) |
| AC-004 rewritten blocks executable, semantics preserved | PASS | census report §6 — six grouped re-execution invocations, all without refusal |
| AC-005 leave files zero diff | PASS | `comm` of git-modified acceptance.md set vs rewrite list at M3 commit: all modified files ∈ rewrite set (52/52); leave files untouched |
| AC-006 no script-file relocation | PASS | rewrites reduced to plain invocations or documented per-item plain procedures; zero `.sh` additions (progress.md M3 entry) |
| AC-007 census closure | PASS | report §5 — sweeps ⊆ union, partition 52/53/6, 0 unprobeable, re-verified at M3 commit time |

Gaps: none blocking. Residual risks: report §7 (version-specific guard behavior; leave-parse rows rest on the comment/pattern exemption measured via pass-controls 11/12).

## §E.3 Run-phase Audit-Ready Signal

run_status: complete
run_complete_at: 2026-09-22
run_commit_sha: a2e6763f4
run_commits: e8daa231c (M1) · dbe1a6941 (M2) · 874aaca79 (M3) · a2e6763f4 (M4)
census_base_sha: 873a19710cac6ec737f12bfaadf976857246c6ad
census_report: .moai/reports/t1067/census-20260922.md
ac_pass_count: 7
ac_fail_count: 0

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
