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

### Delta revision (N1, N2, N6)

Scope: the three findings of `.moai/reports/t1439/plan-audit-iter2.md` (PASS-WITH-DEBT 0.81) that the leader approved for one extra iteration; N3, N4, N5, N7 and the P3 items were not touched. spec.md is now version 0.2.1; counts are unchanged (REQ rows 15, AC headings 14, files 15).

- N1: the Q1 carve-out now reaches every `e2e.md` site by role. The sites are enumerated by command (E16: 15 lines at baseline, current numbers `:36, :54, :90, :108, :119, :197, :219, :277, :328, :331, :332, :334, :342, :344, :345`) and listed with their roles in plan.md M3.2; the carve-out is an appended `(except Aside: ...)` parenthetical so stripping it restores each original line. REQ-ASB-010 and REQ-ASB-012 and spec.md § 1 now say "every site ... that delegates script creation, execution, or recording to the e2e-tester or runs the missing-toolchain sequence" instead of an enumerated count. A new subtest `e2e_every_site_carved_out` (AC-ASB-009 and AC-ASB-010; no new AC) fails when any matching line lacks `except Aside` or `silently`, has a minimum-count guard of 15 lines, and has per-site strip mutants (including `:331`, `:332`, `:342`, `:119`). The probe-table row was dropped: the Aside probe is a sentence in the new Aside paragraph so it sits outside the e2e-tester probe introduction at `:90`. `e2e-tester.md:90` (a generic missing-toolchain sentence in the agent definition) is deliberately not edited; the negative sentence already removes Aside from that agent.
- N2: plan.md § C now splits pre-flight: `aside --version` and the M3.0 screenshot-persistence measurement are run by the orchestrator (main session), with the operator's approval obtained through its question channel, and handed to manager-develop as run-prompt input; manager-develop records them in §E.2 and returns a blocker report if they are absent. spec.md § 3 (R-1, A-6), acceptance.md (AC-ASB-011 gate, Gaps), decision-index Q5, research.md § 5 and spec-compact.md carry the same wording.
- N6: plan.md and acceptance.md AC-ASB-003 now say 13 selectors, matching the command; the baseline run of that exact command printed 13 `--- PASS` lines, `ok  	github.com/modu-ai/moai-adk/internal/template	0.708s`, exit 0 (E11b).
- validator evidence: `moai spec lint SPEC-ASIDE-BROWSER-001 --strict` (binary built at plan time from the base tree; SPEC-only changes since) → exit 0, last output line `0 error(s), 0 warning(s)`. Tree HEAD at the run was `c9d4dd5d2` (the leader's commit of the SPEC artifacts on `4bf547bca`; `git diff --stat 4bf547bca HEAD` lists only files under `.moai/specs/SPEC-ASIDE-BROWSER-001/`). The run printed one `INFO OwnershipTransitionUnmeasured` row for spec.md (the commit carries no `Authored-By-Agent` trailer), which does not affect the exit status.

## §E.2 Run-phase Evidence

_pending run-phase (manager-develop)_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase (manager-develop)_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase (manager-docs)_
