# progress.md — SPEC-SYNC-GATE-SKIP-SUBSHELL-001

## §E.1 Plan-phase Audit-Ready Signal

- SPEC authored 2026-10-02 at plan phase for card t1395; tier M; status `draft`.
- All six card premises verified against tree `89e3164aff20345618c5db053107b066adc98316` (spec.md §B, evidence file:line cited per premise).
- Observational RED executed NOW and RE-EXECUTED from the committed harness (`red-now-t1395.sh` in this SPEC directory): 3/3 defect cells CONFIRMED, 3/3 positive controls PASS, harness exit 0, both runs; `git rev-parse --short HEAD` = `89e3164af` before and after; gate sha256 prefix `19180598f67db114` unchanged (spec.md §B P4, acceptance.md §D.1).
- Static sweep discriminator proven on pre-fix commit: 4 hits working copy / 4 hits `git show 89e3164af:<template>` (acceptance.md §D.2).
- ID uniqueness: no prior SPEC carries the SYNC-GATE-SKIP-SUBSHELL domain (catalog grep, 1009 SPEC dirs).
- Decision gate `on` (`.moai/config/sections/interview.yaml:6`) → `decision-index.md` authored (1 row, FOUNDER).
- Plan-phase artifacts: spec.md, plan.md, acceptance.md, progress.md, decision-index.md.
- plan_status: audit-ready
- plan_complete_at: 2026-10-02
- plan_audit_verdict: PASS (score 0.96, Tier M threshold 0.80; trajectory 0.89 → 0.96 over 2 iterations, no STOP signal; iteration-1 findings D1-D4 were additive and all resolved)
- plan_audit_report: .moai/reports/t1395/plan-audit-iter2.md (full stream: plan-audit.md, plan-audit-iter2.md)
- plan_audit_sha: a43d04efeb3a4f28646676a0d20e11fab9293d05 (delta-verified: 9a6d03edd..a43d04efe touches exactly the 3 SPEC artifacts; the RED harness and both gate-script surfaces byte-unchanged)
- plan_artifact_hash: 04676e813b6ecd5c69ea0027573aa584ddd175887af6426292af78215dfa4547 (the auditor's hash bound to the verdict; a lane audit-cache ComputeHash probe returned d218cec2… — subject-set/algorithm difference between the two tools noted, the run-gate's own recompute owns the mechanical skip check)
- recorded_by: lane orchestrator (verdict landed after manager-spec's final fix turn; audit-ready signal derived from the iteration-2 verdict per the auditor's carry-forward instruction)

## §E.2 Run-phase Evidence

Run phase executed 2026-10-02 by manager-develop (cycle_type=tdd) on run branch
`WT-gate-skipped-tools-run` (isolated worktree; card branch `WT-gate-skipped-tools`
held locked by the lane at `df8af9045`). Baseline absorption merge `e709a2dab`
(origin/develop `6789de39b`) precedes M1 — rationale: AC-006's failstate family
was measurably RED at the card base 89e3164af (pre-existing t1396 fixture defect,
out of scope per spec.md §E; measured on a detached checkout of the base in this
worktree before any edit). Full command/verbatim-output ledger:
`.moai/reports/t1395/run-evidence.md` (lane-local; quoted extracts below).

### AC matrix (commands per acceptance.md §D)

| AC | Status | Actual output (verbatim, condensed — full in run-evidence.md) |
|----|--------|---------------------------------------------------------------|
| AC-001 | PASS | RED (pre-fix, e709a2dab content cf131089db9b7300): `CELL1 CONFIRMED: stdout EMPTY` / `CELL2 CONFIRMED: audit skipped_tools empty: skipped_tools=` / `CELL3 CONFIRMED: pass record written: b04b4d3a… pass cf131089db9b7300…` + 3/3 PC PASS, harness exit 0. GREEN (post-fix, 0267f855f): `CELL1 NOT OBSERVED: stdout={"systemMessage":"sync-phase quality gate: checker(s) dotnet absent — …"}`; audit file final line `…decision=allow dotnet build=0 (none)=0 skipped_tools= dotnet deps_modified=1 head=391ad3e4…`; `CELL3 CONFIRMED: pass record written` (t1385 preserved). Harness CELL2 prints CONFIRMED green because its detector greps `'skipped_tools= '` and the repaired value `" dotnet"` leads with a space — the harness's own captured audit file decides, and it carries the populated value. |
| AC-002 | PASS | `grep -cE "run_step [a-zA-Z]+ c[12] [^']*\|"` → template 0 (exit 1), local 0 (exit 1); positive control `git show 89e3164af:…` → 4. Pre-fix sweep during pre-flight also read 4 on both surfaces before the edit. |
| AC-003 | PASS | `go test ./internal/template/ -run '^TestSyncGateSkipNotice_ToolAbsentNotifies_AllFourLanguages$'`: RED 4/4 subtests FAIL pre-fix on exactly the two lost surfaces (audit value empty, stdout empty/no systemMessage/no tool name); GREEN post-fix `--- PASS` all four (csharp 1.34s, elixir 1.04s, flutter 1.02s, swift 0.78s). Negative control `TestSyncGateExitStatus_PassingCheckerAllows` green inside the AC-005 family run. In-test PC1 (lookPathIn) asserted per subtest. |
| AC-004 | PASS | `cmp` exit 0 post-fix ("BYTE-IDENTICAL post-fix-v2"); `make build` exit 0; `git status --porcelain` showed exactly the 3 expected modified files (no embed/catalog churn); `TestHookWrapperCopiesStayIdentical` PASS (incl. `sync-phase-quality-gate.sh` pair subtest). |
| AC-005 | PASS | `go test ./internal/template/ -run '^TestSyncGateExitStatus'` → `ok … 50.788s` (all subtests incl. FailingCheckerBlocks + PassingCheckerAllows for the four repaired languages). |
| AC-006 | PASS | `go test ./internal/hook/ -run '^TestSyncGateFailState'` → `ok … 157.550s` (green after the documented baseline absorption); AC-003 GREEN record assertion: pass record written on allow-with-skip; `git diff` over the template gate shows exactly 4 hunks, one line each, no hunks outside the four check lines. |
| AC-007 | PASS | `go test ./internal/template/ -run '^TestSyncGateMultiLanguage'` (batched with AC-008) → `ok … 25.180s`. |
| AC-008 | PASS | `go test ./internal/template/ -run '^TestSyncGateCpp'` (batched with AC-007) → `ok … 25.180s`. |

### Invariant rows

| Invariant | Status | Evidence |
|-----------|--------|----------|
| t602 exit-status family unmodified in intent | PASS | AC-005 family ok; the fixture file gained only the new sibling test + scrub helpers (existing three tests byte-untouched). |
| t1392 behaviors unregressed | PASS | AC-007 family ok. |
| t1385 gate-cache contract for genuinely-passing runs | PASS | AC-006 family ok + CELL3 still `pass` post-fix + silent-pass reuse path exercised by AC003_PassThenSameHeadStaysSilent (PASS in family run). |
| emit_gate_notice user-notification reachability (the repair's purpose) | PASS | CELL1 GREEN: stdout carries `{"systemMessage":"…checker(s) dotnet absent…"}`. |
| 7 safe `sh -c` internally-piped lines untouched | PASS | `git diff` = exactly 4 hunks at the four check lines (javac/kotlinc/scalac/ruby/php/g++/R lines byte-unchanged). |
| Local/template byte-identity | PASS | cmp exit 0 at pre-flight (pre-fix), post-fix, and post-merge. |
| Existing test suite never broken | PASS | All scoped families green at every committed state; the only transient RED was the new test against the pre-fix gate (TDD RED, never committed red). |

### TDD order

The behavioral test was authored and observed RED against the pre-fix gate BEFORE
the M1 pipeline removal (RED transcript quoted in run-evidence.md); M1 then landed
the fix (0267f855f) and M3 the GREEN test (7da55ba96). No implementation was
derived before its failing test.

### IN-2 deviation (acceptance-driven, documented)

The trailing `|| true` is dropped on the four repaired lines, against IN-2's
keep-default: the AC-002 discriminator matches any unquoted pipe after the slot
token, so a kept `|| true` still counts 4 post-fix (measured), and the frozen
Blocker AC demands 0. IN-2 itself records keep-vs-drop as behaviorally neutral.
The seven internally-piped checkers keep their `|| true` (untouched).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-02
run_commit_sha: 7da55ba96
run_status: complete
ac_pass_count: 8
ac_fail_count: 0
preserve_list_post_run_count: 7
preserve_list_note: "7 safe sh -c internally-piped checker lines byte-untouched (git diff = 4 hunks at the four check lines); t602/t1392/t1385 families green"
l44_pre_commit_fetch: "git fetch origin develop observed before the e709a2dab absorption merge (Fetch-HEAD update output recorded); branch/HEAD re-read immediately before each commit (staleness rule)"
l44_post_push_fetch: "not-applicable — run agent does not push; develop push is lane-leader-owned (2026-09-02 discipline)"
new_warnings_or_lints_introduced: 0
lint_toolchain: "golangci-lint v2.1.6 (CI pin) run ./internal/template/... -> 0 issues; go vet exit 0; gofmt clean"
coverage_note: "go test -cover ./internal/template/ -run TestSyncGate -> coverage: 0.0% of statements (no production Go delta in this SPEC — repair is bash + a Go test; behavioral coverage is AC-003's four-language sweep)"
cross_platform_build:
  darwin: "primary dev host — all runs executed here"
  windows: "not exercised locally; gate skips on GOOS=windows by design; CI owns the matrix"
  linux: "not exercised locally; CI owns the matrix"
total_run_phase_files: 5
total_run_phase_files_note: "template gate, local gate mirror, hook_gate_exit_status_test.go, spec.md (frontmatter status flip only), progress.md (this signal)"
m1_to_mN_commit_strategy: "pre-M1 baseline absorption merge e709a2dab (origin/develop 6789de39b, t1396 fixture alignment — AC-006 measurability, documented in the merge message) -> M1 0267f855f (4-line repair + mirror + spec status flip) -> M3 7da55ba96 (behavioral test) -> M5 this evidence commit"
baseline_red_note: "failstate family RED at card base 89e3164af measured before any edit; attributed to the pre-existing t1396 fixture defect (stub $1 vs go -C <root> subcommand); repaired upstream by e75f67ec1 and absorbed — not a run-phase regression"
harness_cell2_detector_note: "red-now-t1395.sh CELL2 detector substring-matches the repaired line (value ' dotnet' leads with a space); AC-001 GREEN CELL2 decided from the harness's own captured audit file carrying 'skipped_tools= dotnet' — no harness byte was modified (same-harness same-recipe discipline)"
plan_audit_gate_skip: "Phase-1 Plan Audit Gate skip taken per the authoritative skip contract: plan-auditor PASS 0.96 (iter2) + score >= 0.80 + plan-artifact hash 04676e81… unchanged since the verdict (run-phase §E.2/§E.3 edits and the spec.md status flip are outside the hashed set)"
```

## §E.4 Sync-phase Audit-Ready Signal

<pending sync-phase>

## §F Phase 4 Mode Selection

### Kickoff record (operator-held gate)

- The card dispatch (leader → lane, 2026-10-02) marked the plan→run Kickoff for OPERATOR DIRECT ANSWER — the keep-set operator form, not the autonomous transition.
- Operator answered via AskUserQuestion in the lane on 2026-10-02: **착수 (proceed to run phase)** with decision-index Q1 confirmed as **remove-pipe** (the recommended default; the journal-derived reconstruction alternative was declined — recorded here as the settled decision for Q1, which the decision-index row carried as FOUNDER).
- Gate evidence at ask time: plan-audit iter2 PASS 0.96 (≥ 0.80 Tier M), RED harness 3 defect cells + 3 positive controls reproduced at every run including the auditor's re-execution at the final commit, tree-sourced spec lint 0 errors, plan-artifact hash `04676e813b6ecd5c69ea0027573aa584ddd175887af6426292af78215dfa4547` fixed since the verdict.

### Input parameters

- tier: M · scope: 2 script surfaces (template original + local sync) + embedded regeneration + 1-2 Go test files · domain count: 2 (template/hook script + internal/template Go tests) · file language mix: bash + Go tests · concurrency benefit: LOW (small mechanical diff, dependency-ordered).

### Mode evaluation

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | multi-surface contract (template+local+embed+regression pins), not a one-liner |
| fanout | no | coding-heavy; single-file script surgery with regression families |
| sweep | no | mechanical but 4-line scale, far under any sweep threshold |
| agent-team | no | not operator-requested; experimental surface stays unselected |
| serial | **yes** | one manager-develop carries template+local+tests |

Decision: serial

Justification: the mechanical diff is 4 lines but the surface contract (Template-First dual file + embed regeneration + test fixture + two regression families) requires one actor holding the whole picture; a sub-agent-per-milestone fan-out would split that picture for no concurrency benefit.
