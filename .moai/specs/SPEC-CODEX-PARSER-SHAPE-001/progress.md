# SPEC-CODEX-PARSER-SHAPE-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

- Card: t1053 · worktree `.claude/worktrees/t1053` · branch `WT-codex-parser-shape`
- Plan-phase artifacts authored: `spec.md`, `plan.md`, `acceptance.md` (Tier M)
- Evidence base: `.moai/reports/t1053/verdict.md` (tree `8b55fc8f0`). No
  measurement re-run; no figure introduced that is absent from that file.
- Status: `draft`.
- AC-CPS-001 / AC-CPS-002 (live-convention comparison): **SATISFIED 2026-09-21**,
  result same-shape. Record: `.moai/reports/t1053/live-convention-20260921.md`
  (tree `a5c3f5dc6`, codex-cli 0.155.1). The pre-run gate is no longer blocking;
  the remaining Kickoff decision is the §C candidate selection, which is the
  operator's.
- Plan-phase corrections after the live call (three defects codex found in this
  SPEC, each verified against the repository before acceptance): AC-CPS-001's
  no-findings edge case now stays OPEN instead of closing; candidate (c)'s
  predicate is scoped with `GateUnmet == ""` plus a mandatory control case
  (REQ-CPS-006a); AC-CPS-006 now requires exact count and content plus a
  partial-drift case.
- 2026-09-26 · v0.2.0 amendment (card t1203, GitHub #1718; worktree
  `.claude/worktrees/t1203`, branch `WT-codex-parser-shape`): the #1718 real case
  added from `.moai/reports/t1203/verdict.md` (tree `df526c9a9`) — spec.md §A.6,
  §A.4 conclusion scoped, §C #1718 coverage table, candidate (d), §C.1 conflict;
  REQ-CPS-012..014; AC-CPS-011..015. Status stays `draft`. The pending Kickoff
  decision now covers BOTH the §C candidate selection AND the REQ-CPS-010
  question (keep as written, or revise — spec.md §C.1, REQ-CPS-013); neither is
  taken here. The plan-audit verdict from before this amendment does not cover
  it; a fresh plan-audit is due.
- 2026-09-26 · v0.2.1 repair (card t1203): plan-audit iter-1
  (`.moai/reports/t1203/plan-audit.md`, FAIL 0.75) defects D1–D14 addressed in
  wording and structure only — no new measurement, no candidate chosen. REQ 15 /
  AC 15, unchanged. The REQ-CPS-010 question is now posed as "is the #1718
  adversarial outcome acceptable?" and excludes no candidate under either answer
  (spec.md §C.1). Both Kickoff decisions remain pending. Next: delta plan-audit
  (iter-2, the last under the Tier M ceiling) limited to D1–D14.
- 2026-09-26 · v0.2.2 repair (card t1203): plan-audit iter-2
  (`.moai/reports/t1203/plan-audit-iter2.md`, FAIL 0.83) — blocking N4-P4, N5,
  N1 and optional N2, N3, N6 addressed in wording and check commands only; no
  candidate chosen, REQ 15 / AC 15 unchanged. Iter-2 was the Tier M ceiling, so
  the next step is the orchestrator's escalation choice (PASS-with-debt, scope
  reduction, or an explicit extension), not an automatic re-audit. Both Kickoff
  decisions remain pending.
- 2026-09-26 · plan-audit iter-3 (card t1203; operator-authorized extension
  beyond the Tier M ceiling of 2): PASS-WITH-DEBT 0.89, report
  `.moai/reports/t1203/plan-audit-iter3.md`. Debt carried: N7 (P10 does not
  decide "prose states FAIL"; mutants M3/M4 pass P1..P11) — to be fixed before
  the first commit of the run's second milestone; N8, N9, N10 (minor).
- 2026-09-26 · §C candidate selection (operator, lane question channel): all
  four candidates — (a), (b), (c), (d) — selected.
- 2026-09-26 · Implementation Kickoff Approval: HELD by the operator. The run
  phase has NOT started; `status` stays `draft`. [SUPERSEDED — see the
  run-resume decision below. Run work subsequently recorded in §E.2 landed
  with no approval record observed for this line.]
- 2026-09-26 · spec v0.2.3 (manager-spec, card t1203): run-phase acceptance
  wording fixes applied — N7 (P10 bound to line 1 + P10b), N8, N9, N10,
  AC-CPS-013 re-anchored, AC-CPS-011 check 1 selector note; AC count 15
  unchanged; AC-CPS-004 / REQ-CPS-005 untouched.
- REQ-CPS-010 decision: keep — reason: the operator judged the #1718 adversarial outcome (`inconclusive` with an empty findings list for a body whose prose states FAIL) acceptable — source: operator answer via the card t1203 lane question channel on 2026-09-26, recorded in this progress.md §E.1

- 2026-09-26 · Run-resume decisions (operator, worker-63 lane question
  channel, this worktree at `52ab653c8`): (1) run continues — the remaining
  scope is candidate (b) and the AC-CPS-014 live observation only; this entry
  supersedes the HELD line above. (2) AC-CPS-004 (b): the native
  disambiguation mechanism is authored into the SPEC (manager-spec) and then
  implemented (manager-develop) — resolving the blocker reported in §E.2.
  (3) AC-CPS-014 (d): one live codex review call is authorized for the live
  observation.
- 2026-09-26 · spec v0.2.4 (manager-spec, card t1203): the candidate (b)
  native disambiguation mechanism authored per decision (2) of the run-resume
  entry above — resolving the AC-CPS-004 BLOCKED row of §E.2. REQ-CPS-005
  amended in place: the native review request pins its output format (the (d)
  family, REQ-CPS-012, applied to the native request), the no-recognized-
  signal fall-through on native downgrades to `inconclusive`, the target the
  request reviews is unchanged, and the live-honours-the-pin claim carries
  the same live-observation burden REQ-CPS-012 states. AC-CPS-004 amended in
  place with two-cell adoption: RED-now re-executed on this tree `0ff644530`
  via the committed fixture test (`N1`/`N2` review/start = `pass`/0 observed —
  the silent pass the mechanism removes; turn/start lines the positive
  control), green path = plan.md M4, mutant probe sharpened the criterion
  (M-A pinned-pass class, M-B prose-token keying). AC-CPS-016 added: the
  pinned native format observed in live output — regression-guard, pending
  its OWN operator authorization (the authorized AC-CPS-014 call is
  adversarial-only; the orchestrator surfaces the native-call question).
  AC-CPS-008 gains a scope note (its guard fixture bodies state the pinned
  line; assertion unchanged). REQ 15 / AC 16. Implementation is M4's
  (manager-develop); this amendment establishes nothing about whether live
  codex honours the pin.

## §E.2 Run-phase Evidence

Run by manager-develop (cycle_type=tdd), card t1203, branch
`WT-codex-parser-shape`, base `e26633699`. Local logs cited below live under
`.moai/reports/t1203/run/` (gitignored — local evidence, not committed); the
deciding lines are carried here.

**Pre-flight (tree `e26633699`).** `go build ./...` exit 0;
`GOOS=windows GOARCH=amd64 go build ./...` exit 0;
`golangci-lint run --timeout=2m ./internal/cli/...` → `0 issues.`. E-1718
re-measured with the probe from `repro/probe_test.go.txt` (temporary file,
deleted): body1 `inconclusive`/0 (turn/start), `pass`/0 (review/start); body2
the same; ctrlA `fail`/1 both; ctrlB `fail`/0 both — identical to E-1718.

**Commits.**

| SHA | Milestone | Content |
|---|---|---|
| `6a89f3d83` | M2 step 1 | status → in-progress; five sanitized fixtures; fidelity test; N7 P10 repair; §F |
| `562126b1f` | M2 (c) | `ReviewOutput.Contradiction` + `flagVerdictFindingsContradiction` |
| `ce1df7f6f` | M2 (a) | greeting-prefixed / localized verdict recognizers; bold-severity table-row and bullet findings |
| `83046be7c` | M2 (d) | adversarial prompt pins the output format |
| `54fe08135` | M3 | preserved-behaviour guards |

**AC matrix.**

| AC | Status | Command | Observed |
|---|---|---|---|
| AC-CPS-004 (b) | **BLOCKED** | — | No disambiguation mechanism is specified by the SPEC; (b) not implemented. Blocker report returned to the orchestrator |
| AC-CPS-005 (c) | PASS | `go test -count=1 -v -run 'Contradiction' ./internal/cli/` at `562126b1f` | V8 body flagged on both paths; predicate table incl. the unmet-gate control; end-to-end control through `applyGateUnmet` not flagged — `ok` |
| AC-CPS-006 (a) | PASS | `-run 'TestCodex1718\|TestCodexWidened'` at `ce1df7f6f` | exact count + severity/message/file/line for S1 (3), S2 and S2p (1) on both paths; partial drift → 2 and 0 survivors, exact count fails; `codex_findings_parse_test.go` (V1) unchanged and passing |
| AC-CPS-007 | PASS | `git diff e26633699 -- internal/cli/codex_findings_parse_test.go \| wc -l` | `0` |
| AC-CPS-008 | PASS | `-run TestGuard_` | clean native bodies stay `pass`/0, no contradiction; mutant "drop the verdict term from the predicate" makes the guard fail |
| AC-CPS-009 | PASS (keep) | `-run TestGuard_` | N1, N2 and the unrecognized AC-CVS-001 corpus members stay `inconclusive`; mutant "unrecognized adversarial → pass" makes the guard fail |
| AC-CPS-010 | PASS | `git diff e26633699 -- internal/cli/mcp_codex.go \| grep -cE '^[+-].*(NextSteps\|next_steps)'` | `0` (same form on `Contradiction` → `8`, so the probe fires); NextSteps guard + mutant "NextSteps nil" fails it |
| AC-CPS-011 | PASS | `go test -count=1 -v -run '^TestCodex1718Fixtures$' ./internal/cli/` at `6a89f3d83` | exit 0, `--- PASS: TestCodex1718Fixtures`, ten SYNTH lines matching E-1718 row for row (S1/S2 `inconclusive`/0 + `pass`/0; S2p `fail`/0 both; N1/N2 `inconclusive`/0 + `pass`/0). Ledger P1–P11 all at their pass conditions; sanitization grep empty exit 1, firing control prints the inserted line exit 0 (`ac011-ledger.log`) |
| AC-CPS-012 | PASS | same selector at `ce1df7f6f` | S1 `fail`/3, S2 `fail`/1, S2p `fail`/1 on both paths with exact content; N1/N2 unchanged; partial drift fails the exact count. RED before GREEN in `red-a.log` |
| AC-CPS-013 | PASS at `562126b1f`; see note | `-run 'Contradiction'` at `562126b1f` | S2p flagged on both paths; S1/S2 not flagged (`inconclusive`/0, `pass`/0). **Note:** (c) alone leaves the #1718 shapes at `inconclusive`/0 (turn/start) and `pass`/0 (review/start). With (a) also selected, S2p yields `fail`/1 from `ce1df7f6f` on, so on HEAD it is no longer contradictory — the criterion's first clause is not true on the final tree; wording change routed to manager-spec |
| AC-CPS-014 (d) | **NOT MEASURED** | — | requires a live codex call the operator has not authorized; no live call was made |
| AC-CPS-015 | PASS | checks 1–4 | check 1 `1`/exit 0; D = `802ac54d2c17a6dc9afd07dfa44832389cdba944`; R = `6a89f3d8304ab37dc57a84b9ca17fe81062499c8`; `git merge-base --is-ancestor D R` exit 0; D ≠ R. Authorship remains reviewable, not mechanically proven |

**RED evidence (captured before each GREEN).** (c): build failure
(`undefined: flagVerdictFindingsContradiction`), then with a no-op stub
`--- FAIL` on the V8, predicate and fixture tests (`red-c.log`). (a): six
`--- FAIL`s, e.g. `S1.txt on turn/start: got inconclusive/0, want fail/3`
(`red-a.log`). (d): `prompt does not carry "Verdict: <pass|fail|inconclusive>"`
(`red-d.log`). M3 guards describe pre-existing behaviour and are shown live by
mutation instead (`guard-mutants.log`); recognizer narrowness was shown the same
way — relaxing the comma adjacency or the 판정 statement form fails N2 and the
negative cases, and admitting a hyphen after the bold severity word fails the
`**High-level**` negative (`recognizer-mutants.log`).

**Offline observations beyond the ACs (local, tree `ce1df7f6f`+).** The raw
#1718 bodies now synthesize `fail`/3 (body1) and `fail`/1 (body2) on both paths
(`probe-raw-post-a.log`). Over the 142-body population: turn/start
`fail`/0 = 95, `fail`/>0 = 14, `inconclusive`/0 = 21, `inconclusive`/>0 = 2,
`pass`/0 = 10; review/start `fail`/0 = 95, `fail`/>0 = 14, `pass`/0 = 31,
`pass`/>0 = 2; contradiction flagged = 95 per path (`pop-post-ac.log`). Every
localized-label match was a statement form (`판정: FAIL` 78, `판정은 **FAIL` 24,
`판정: PASS` 4, three others) (`pop-signals.log`).

**Quality.** `go build ./...` exit 0 and `GOOS=windows GOARCH=amd64 go build
./...` exit 0 on `54fe08135`; `golangci-lint run --timeout=2m
./internal/cli/...` → `0 issues.` (baseline also 0). Selector-scoped coverage
(`-run 'TestCodex1718|TestCodexWidened|Contradiction|TestCodexAdversarial|TestGuard_|TestSynthesizeReviewOutput|Verdict'`):
every touched function in `mcp_codex.go` at 100.0%. Parser/audit regression
selector (`-run 'Codex|Synthesize|Review|Verdict|Convergence|Audit|…'`) `ok`
after each of (c), (a), (d). Full `go test -count=1 -cover ./internal/cli/`
**timed out** at 10m (`panic: test timed out after 10m0s`, running
`TestDoctorCmd_Execution`; zero `--- FAIL` lines) — a timeout, not a failure;
package-level coverage was therefore not measured.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-26
run_commit_sha: 54fe08135
run_status: partial — (a) (c) (d) implemented; (b) blocked pending an operator decision
ac_pass_count: 10   # 005 006 007 008 009 010 011 012 015, plus 013 at 562126b1f
ac_fail_count: 0
ac_blocked: [AC-CPS-004]
ac_not_measured: [AC-CPS-014]
ac_wording_change_needed: [AC-CPS-013]
preserve_list_post_run_count: n/a (no PRESERVE list declared)
l44_pre_commit_fetch: not run (lanes do not push; lead batch-pushes develop)
l44_post_push_fetch: not applicable (no push)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: exit 0
  windows_amd64: exit 0
total_run_phase_files: 13
m1_to_mN_commit_strategy: one commit per milestone step, no push, no amend
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

- Decision: **serial**
- Inputs: Tier M; one package (`internal/cli`) plus its `testdata/codex-1718/`
  fixtures and this SPEC's progress record; coding-heavy; one domain (the codex
  review parser).
- Rationale: coding-heavy single-package work — the milestones edit the same
  file (`internal/cli/mcp_codex.go`) in sequence and each depends on the
  previous commit's test state (Anthropic coding-task parallelism caveat), so
  parallel writers would contend for one file with no independent unit to give
  each.
