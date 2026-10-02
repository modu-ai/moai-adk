# progress — SPEC-ASIDE-BROWSER-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-10-02 (iteration 2 revision; the iteration 1 signal carried 2026-10-02T11:52:52Z from the `spec_audit` tool, and this revision's validator run is recorded below)
- plan-audit: iteration 1 returned FAIL 0.75 (report `.moai/reports/t1439/plan-audit-iter1.md`, defects D1-D13); iteration 2 audit not yet run. This signal states that the revised artifacts pass the plan-phase validators below, not that an independent audit passed.
- tier: M · artifacts: spec.md (0.2.0), plan.md, acceptance.md, progress.md, plus spec-compact.md, research.md, decision-index.md (the last two beyond the Tier M set, by card request and by `interview.decision_gate: on`)
- requirements: 15 (REQ-ASB-001..015, budget 16) · acceptance criteria: 14 (AC-ASB-001..014, budget 16) · files: 15 (Tier M ceiling)
- measurement tree: `4bf547bcad7c155b1e91485921569db709ec3ac2` (worktree t1439, base develop); `git rev-parse HEAD` re-read at the start of the revision printed the same SHA
- SPEC ID check (iteration 1, Bash, verbatim): `[[ "SPEC-ASIDE-BROWSER-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` printed `PASS`

### Iteration 2 revision — what changed

- Operator verdicts recorded 2026-10-02 (decision-index Q1, Q2, Q3; spec.md § 3 R-1 to R-3): orchestrator-only execution, core skill tier, completely silent fallback.
- REQ changes: REQ-006 split into REQ-006 (write confirmation) and REQ-007 (no subagent invokes Aside); the old REQ-012 (e2e-tester delegated prompt) inverted into REQ-007 and REQ-010 (orchestrator runs every Aside step, bounded output); the fallback-note requirement removed from REQ-012; REQ-003 and REQ-014 stripped of process text (moved to plan.md); REQ-013 hedged (no claim about which side writes the screenshot bytes). 14 to 15 REQs, 12 to 14 ACs.
- Removed: the e2e-tester Aside recipe block, its template/root edit, and the delegated-prompt subtests. Kept: one negative sentence in the e2e-tester definition (template and root) and the same boundary in the skill, pinned by `subagent_never_invokes_*` subtests with deletion mutants. The e2e-tester core catalog hash and the Codex TOML therefore still change, so those steps stay.
- Audit defects closed: D1 (hash set now `{moai-ref-aside-browser, moai, e2e-tester}`), D2 (t1434 dependency paragraph in plan § A.3, lines 21/27/32 of the sibling verdict re-read), D3 (core tier; measured slim fact recorded in Q2 and research), D4 (`description_within_listing_cap` subtest), D5 (`:108` and `:119` carve-outs with subtests and a restore mutant), D6 (operator verdicts applied), D7 (commit order chosen: RED commit before GREEN commit), D8 (REQ-013 hedged), D9 (exact docs commands and the four-to-five numeral edit), D10 (neutrality grep over every authored template file with a 40-hex pattern; `e2e.md` baseline has two pre-existing hits), D11 (closed two-literal allow-list for `full-access`, `never forget` mutant, concrete advise-the-operator phrase), D12 (REQ split, process text moved), D13 (README and skill-guide listed out of scope; the new test's duplication justified and its settings-token half named as the non-duplicate part).
- New open decision-index row: Q8 (which e2e phases stay with the e2e-tester). New assumption: A-8.
- validator evidence for this revision (tree HEAD `4bf547bca`): `moai spec lint SPEC-ASIDE-BROWSER-001 --strict` → exit 0, last output line `✓ No findings — all SPEC documents are valid`; the orchestrator independently re-ran it after the revision and observed exit 0 with the same line (REQ rows = 15, AC headings = 14).

## §E.2 Run-phase Evidence

_pending run-phase (manager-develop)_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase (manager-develop)_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase (manager-docs)_
