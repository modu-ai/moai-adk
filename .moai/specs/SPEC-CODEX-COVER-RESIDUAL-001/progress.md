# SPEC-CODEX-COVER-RESIDUAL-001 — Progress

Card: t519 · Branch: `WT-codex-cover-residual` · Plan-phase base: `bf779ecf2`

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_complete_at: 2026-09-07T08:54:01Z
plan_status: audit-ready
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
requirements: 10   # REQ-CCR-001..010
acceptance_criteria: 12   # AC-CCR-001..012
needs_clarification: 0
```

### Tier deviation from dispatch

The dispatch named Tier S; this SPEC is classified **M**. Grounds are mechanical and recorded in spec.md §A.1: the requested artifact set is the 3-file Tier M set, and the counts (10 REQ / 12 AC) exceed both Tier S ceilings of 8. The card text allowed "Tier S~M". Consequence: plan-audit PASS threshold is 0.80 and the Section A-E delegation template is required.

### Phase 1 SKIP rationale (no Socratic interview)

The Context-First Discovery interview was skipped. Intent clarity was already 100% at dispatch: the card named both target surfaces with file and line, supplied a measured per-function baseline with its command and verbatim output, enumerated the six test items, named the reusable seams, and pre-identified one vacuous mutant. None of the four ambiguity triggers fired — no unresolved referent, no multi-interpretable verb, no unclear boundary (the out-of-scope set is explicit), and no conflict with existing state. No `[NEEDS CLARIFICATION]` markers were emitted.

One judgment call was resolved without asking, because the card itself supplied the latitude: the Tier S-vs-M inconsistency above, resolved by the governing ceiling rule and surfaced rather than silently absorbed.

### Plan-phase verification performed

Read-only checks run by manager-spec on tree `bf779ecf2`:

| Check | Command | Observation |
|---|---|---|
| SPEC ID format | `[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS \|\| echo FAIL` | `PASS` |
| ID uniqueness | `ls -d .moai/specs/SPEC-CODEX-COVER-RESIDUAL-001` | `No such file or directory`, exit 1 — no collision |
| Axis 1 baseline | `grep runCodexReviewGate .moai/reports/t519/coverage-baseline-targets.txt` | `…codex_review_gate.go:183:		runCodexReviewGate			0.0%`, exit 0 |
| Axis 2 baseline | `grep mcp_codex.go:701 .moai/reports/t519/coverage-baseline-targets.txt` | `…mcp_codex.go:701:			pid					66.7%`, exit 0 |
| Codex RunE tests absent | `grep -rn 'func TestRunCodexReviewGate' internal/cli/` | no output, exit 1 |
| Control for the above | `grep -rn 'func TestRunMultiReviewGate' internal/cli/` | 3 rows from `multi_review_gate_wiring_test.go`, exit 0 |

The last two rows are a pair. The empty result alone would be consistent with a wrong grep form or a wrong path; the control run in the same form on the sibling gate prints three rows, so the absence is a real absence rather than a broken probe.

Source facts were verified by direct read, not inferred: `runCodexReviewGate` (codex_review_gate.go:183-202), `(codexSessionHandle).pid` (mcp_codex.go:701-706), `codexConnPID` (mcp_codex.go:411-416), `(*fakeCodexConn).pid` (codex_jobs_test.go:35), `newGateCmd` binding to `runMultiReviewGate` (multi_review_gate_wiring_test.go:104), and the one-for-one symmetry claim in that file's header (line 13).

**Not verified by manager-spec:** the coverage baseline itself was measured by the lane, not re-run here; it is cited with that provenance in spec.md §B.1. The 12-of-13 statement ceiling in spec.md §D.1 is a derivation from reading the source, not a measurement, and is labelled as such.

### spec lint

```
$ moai spec lint .moai/specs/SPEC-CODEX-COVER-RESIDUAL-001
```

Result recorded at run time by the lane — see §E.2. Known candidate defect from card t500: `moai spec lint` may emit a `ParseFailure` on an ID-form argument; if it does, the exact command and output are recorded here and the run continues (the defect is the linter's, not the SPEC's).

### Card Cross-Check

| Milestone | Card |
|---|---|
| M1 — runCodexReviewGate wiring tests | t519 |
| M2 — codexSessionHandle.pid nil-guard arm | t519 |
| M3 — close: re-measurement + evidence | t519 |

3 milestones → 1 card (t519). No milestone requires a new card.

### Plan-audit iteration 1 — fix round

Verdict: **FAIL score=0.775** (Tier M threshold 0.80; blocking 4, advisory 5). Report: `.moai/reports/t519/plan-audit-iter1.md`, audited on tree `faa76fa7e`. All seven must-pass criteria passed; the FAIL was driven by the aggregate score, and three of the four blocking findings were the same defect — a Given clause written from a sibling test's *intent* rather than its *text*, dropping a seam the sibling actually carries.

Every finding was re-verified against the source before editing rather than accepted on the report's word, because the auditor's own Gaps section states it executed no tests and its predictions are control-flow reasoning. The verification is recorded below alongside each fix.

| Finding | Artifact:line changed | What changed | Source verification |
|---|---|---|---|
| F1 | acceptance.md AC-CCR-004 Given + rationale; plan.md §F M1 step 5 | The fixture now swaps **all three** codex seams under one `t.Cleanup`, with the `codexLookPath` line copied verbatim; plan.md carries the block as a code fence. Added the reason: without it the test does a real PATH lookup and its verdict depends on the host. | `var codexLookPath = exec.LookPath` confirmed at mcp_codex.go:368; the three-seam precedent confirmed at codex_review_gate_test.go:152-164; `HandleCodexReviewGate` consults it at step 4 (codex_review_gate.go:78), before the session starts. |
| F2 | acceptance.md AC-CCR-003 Given + new rationale paragraph; ledger row M2; plan.md §F M1 step 4 | Added `withChangeDetector(t, true)` and stated why it must not be dropped as redundant: without it the mutated handler ALLOWs at step 3 on the non-git temp dir and M2 is vacuous. | The precedent pairing confirmed at codex_review_gate_test.go:40-46 (comment "even with changes present…"); the non-git-dir-yields-false assertion confirmed at codex_review_gate_test.go:312-314. |
| F3 | acceptance.md §B matrix row AC-CCR-002; ledger rows M1 and new M1b; plan.md §F M1 step 7 + M1 exit line | Split the row: M1 keeps the `fmt.Fprintf` deletion and is bound to AC-CCR-001 only; new **M1b** replaces codex_review_gate.go:188 with `return err` and is AC-CCR-002's adoption basis. | Confirmed by reading codex_review_gate.go:184-189 — `err` is in scope inside the `if err != nil` block, so the edit compiles; empty stdin does produce a non-nil `err` (`json.Unmarshal` on empty input), so M1b flips AC-CCR-002 as well as AC-CCR-001. |
| F4 | plan.md §A.4 table (rewritten, 12 rows); spec.md §E constraint 3 | Every Location cell corrected and the file count stated: the helpers live in **four** files. `stubCodexRunner` moved to `codex_rpc_error_test.go:29`, a file neither artifact had named. | All 12 locations grep-verified in this worktree: `withChangeDetector` codex_review_gate_test.go:28; `writeWorkflowYAML` :33 / `assertAllowJSON` :157 in multi_review_gate_wiring_test.go; `withCodexRunner` :56, `withCodexLookPath` :63, `fakeCodexSession` :74, `fakeCodexConn` :88, `withCodexSession` :108, `codexSessionScript` :123, `errFakeCodexCrash` :408 in mcp_codex_test.go; `stubCodexRunner` codex_rpc_error_test.go:29; `fakeCodexConnPID` codex_jobs_test.go:31. |

Advisories:

| Advisory | Disposition | Where |
|---|---|---|
| A1 — leave the ≥90.0% threshold alone | **Honoured** (it recommends changing nothing) | no edit; the threshold is untouched |
| A2 — no `t.Parallel()` in the six new tests | **Applied** | plan.md §D constraint 10, with the seam-race reason and the sibling-file evidence |
| A3 — M5a's RED arrives as a panic | **Applied** | acceptance.md AC-CCR-010, third bullet: run M5a in isolation and record the panic trace beside the `--- FAIL:` line |
| A4 — spec.md has no scope heading | **Declined.** The fix would edit spec.md prose, outside the acceptance.md / plan.md surface this round authorises, and scope is already carried precisely by §B.2's Disposition column and §F. Re-raise at iteration 2 if the auditor still finds it material. | no edit |
| A5 — `moai spec lint` fails on a directory argument | **Declined as a SPEC change.** It is a linter defect, already pre-recorded in §E.1 and reproduced by the auditor. Belongs in `/moai feedback`, not this SPEC. | no edit |

One defect neither the audit nor the original pass caught, found while applying the above and fixed in the same round: spec.md REQ-CCR-003 pointed at `§D.1` (the coverage-ceiling derivation) where it meant `§D.2` (the vacuous-mutant record). Corrected at spec.md:92.

Post-fix re-verification on this tree: `moai spec lint .moai/specs/SPEC-CODEX-COVER-RESIDUAL-001/spec.md` → `✓ No findings — all SPEC documents are valid`. Counts unchanged at 10 REQ / 12 AC; ledger grew from 9 rows to 10 with M1b. No production `.go` file touched.

## §E.2 Run-phase Evidence

All figures below were measured in the worktree `.claude/worktrees/t519`, branch
`WT-codex-cover-residual`. Milestone commits: M1 `4ed5011c4`, M2 `a26feb41e`, M3 (this record).
Every coverage and sweep figure was taken on `a26feb41e` — the tree after M1 and M2, before the M3
evidence commit. No figure is carried over from another tree or another run; the plan-phase baseline
(tree `bf779ecf2`) is cited only as the before-value it is.

The env scrub travels with every test command as one compound invocation:
`unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test …`.

### E.2.a AC matrix

| AC | Class | Status | Command | Observation |
|---|---|---|---|---|
| AC-CCR-001 | RB | **PASS** | `-run 'TestRunCodexReviewGate_InvalidStdinFailsOpen' -v` | `--- PASS: TestRunCodexReviewGate_InvalidStdinFailsOpen (0.00s)`; adopted via M1 (RED verbatim in the ledger) |
| AC-CCR-002 | RB | **PASS** | `-run 'TestRunCodexReviewGate_EmptyStdinFailsOpen' -v` | `--- PASS: TestRunCodexReviewGate_EmptyStdinFailsOpen (0.00s)`; adopted via M1b — M1 confirmed by measurement NOT to fire it, which is why the split was correct |
| AC-CCR-003 | RB | **PASS** | `-run 'TestRunCodexReviewGate_HappyPathAllow' -v` | `--- PASS: TestRunCodexReviewGate_HappyPathAllow (0.00s)`; adopted via M2, whose RED is the `t.Fatal` LookPath guard |
| AC-CCR-004 | RB | **PASS** | `-run 'TestRunCodexReviewGate_HandlerErrorFailsOpen' -v` | `--- PASS: TestRunCodexReviewGate_HandlerErrorFailsOpen (0.00s)`; adopted via M3a **and** M3b, both RED |
| AC-CCR-005 | RB | **PASS** | `-run 'TestRunCodexReviewGate_BlockVerdictPropagates' -v` | `--- PASS: TestRunCodexReviewGate_BlockVerdictPropagates (0.00s)`; adopted via M4 |
| AC-CCR-006 | RB | **PASS** | `-run 'TestCodexSessionHandlePid' -v` | `--- PASS: TestCodexSessionHandlePid (0.00s)`; adopted via M5a (panic RED, run isolated) and M5b |
| AC-CCR-007 | RB | **PASS** | `git diff --name-only bf779ecf2..HEAD` + `git status --short`, each filtered `grep '\.go$' \| grep -v '_test\.go$'` | both filters empty (rc=1). Fired via M6 first — see E.2.c |
| AC-CCR-008 | RB | **PASS** | `go tool cover -func` on `cover-after.out` | `runCodexReviewGate 92.3%` (≥ 90.0% required) and `mcp_codex.go:701 pid 100.0%` — see E.2.d |
| AC-CCR-009 | RB | **PASS** | six-name `-run` sweep, `grep -c '^--- PASS: Test'` | exactly `6`; no `no tests to run` / `no test files` / `FAIL` line — see E.2.e |
| AC-CCR-010 | RB | **PASS** | ledger at `.moai/reports/t519/mutants.md` | 8 mutants RED-then-GREEN, 2 deliberate non-firing probes recorded — see E.2.f |
| AC-CCR-011 | RG | **PASS** | `go vet`, `golangci-lint run`, `gofmt -l` | vet rc=0; lint `0 issues.`; gofmt lists nothing — see E.2.g |
| AC-CCR-012 | RG | **DEFERRED to sync** | not exercisable at run phase | The criterion inspects the tree after the sync-close commit, which does not yet exist. Its undecidable disposition is recorded in acceptance.md; it is NOT claimed PASS here. |

### E.2.b Pre-flight (tree `21a52d507`, before the first edit)

| Check | Observation |
|---|---|
| `git branch --show-current` | `WT-codex-cover-residual` |
| `git rev-parse --short HEAD` | `21a52d507` (descends from `bf779ecf2`) |
| `git status --short` | empty |
| `go build ./...` | rc=0 |
| `GOOS=windows GOARCH=amd64 go build ./...` | rc=0 |
| `go vet ./internal/cli/` | rc=0 |
| `golangci-lint run --timeout=3m ./internal/cli/` | `0 issues.` — the inherited lint baseline is CLEAN, so any finding at close would be NEW |
| `grep -rn 'func TestRunCodexReviewGate\|func TestCodexSessionHandlePid\|func newCodexGateCmd' internal/cli/` | no output, rc=1 |
| Control `grep -rn 'func TestRunMultiReviewGate' internal/cli/` | 3 rows from `multi_review_gate_wiring_test.go`, rc=0 |

The last two rows are read as a pair: the empty result alone would also be produced by a wrong grep
form or path, and the control run in the identical form on the sibling gate prints three rows, so
the absence is a real absence rather than a broken probe.

### E.2.c AC-CCR-007 — the union gate, adopted by firing it

Adopted via M6 rather than RED-now, because an absence guard is satisfied for free before any work
is done. Applied: `printf '\n' >> internal/cli/codex_review_gate.go`.

```
--- MUTANT M6 APPLIED: git status --short ---
 M internal/cli/codex_review_gate.go
--- union .go non-test filter ---
 M internal/cli/codex_review_gate.go
filter rc=0 (0 = violation detected = guard FIRED)
```

Reverted from the byte-exact backup; `internal/cli/codex_review_gate.go` re-hashed to
`9356669bdeb39f433897b7fc82c7e4cf036ca8ab24197301558b8c7bcedc3500`. The guard then read clean:

```
--- M6 REVERTED: union gate ---
status ends
[git diff --name-only bf779ecf2..HEAD]
.moai/reports/t519/coverage-baseline-run.log
.moai/reports/t519/coverage-baseline-targets.txt
.moai/reports/t519/mutants.md
.moai/reports/t519/plan-audit-iter1.md
.moai/reports/t519/plan-audit-iter2.md
.moai/specs/SPEC-CODEX-COVER-RESIDUAL-001/acceptance.md
.moai/specs/SPEC-CODEX-COVER-RESIDUAL-001/plan.md
.moai/specs/SPEC-CODEX-COVER-RESIDUAL-001/progress.md
.moai/specs/SPEC-CODEX-COVER-RESIDUAL-001/spec.md
internal/cli/codex_review_gate_wiring_test.go
internal/cli/mcp_codex_test.go
--- diff non-test .go filter (expect empty) ---
filter rc=1 (1 = no violation)
```

Every `.go` path in the union ends in `_test.go`. The non-`.go` paths are SPEC artifacts and
evidence files, which the criterion expects.

### E.2.d AC-CCR-008 — coverage on the final tree

Command (tree `a26feb41e`):

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -timeout 700s -coverprofile=.moai/state/verify/t519/cover-after.out ./internal/cli/
```

Verbatim: `ok  	github.com/modu-ai/moai-adk/internal/cli	448.176s	coverage: 81.0% of statements`, rc=0.
This run doubles as the full-package suite required by plan §F M3 step 4 — it is the whole package,
not a selector, so a separate re-run would measure the same thing twice.

Target rows from `go tool cover -func` (`.moai/reports/t519/coverage-after-targets.txt`):

```
github.com/modu-ai/moai-adk/internal/cli/codex_review_gate.go:183:		runCodexReviewGate			92.3%
github.com/modu-ai/moai-adk/internal/cli/mcp_codex.go:701:			pid					100.0%
```

| Function | Baseline (`bf779ecf2`) | Final (`a26feb41e`) | Threshold | Verdict |
|---|---|---|---|---|
| `runCodexReviewGate` | 0.0% | **92.3%** | ≥ 90.0% | PASS |
| `(codexSessionHandle).pid` | 66.7% | **100.0%** | 100.0% | PASS |

The measured 92.3% is exactly the spec.md §D.1 derived ceiling (12 of 13 statements). The
derivation was labelled a derivation, not a measurement; this run is what establishes it, and the
two agree.

Package-wide, recorded and NOT gated (spec.md §F): 80.9% → 81.0% (`go test` mode), 81.1% → 81.1%
(`-func` total). The two totals differ because they count differently; neither is reconciled into
the other. The package figure remains below the project's 85% target for reasons outside this
SPEC's scope.

**S1 documented-skip proof (REQ-CCR-006).** The uncovered statement is the `out == nil` defensive
arm, and the proof is a code reading, not tool silence. `HandleCodexReviewGate` declares
`allow := &hook.HookOutput{}` at codex_review_gate.go:67 — non-nil — and carries exactly seven
returns, every one of which returns a non-nil `*hook.HookOutput`:

```
69:		return allow, nil // (1) opt-in default-off
72:		return allow, nil // (2) loop prevention — never re-block an already-continuing turn
75:		return allow, nil // (3) self-gate — nothing reviewable ⇒ no false block
80:		return allow, nil // (4) fail-open: a missing reviewer must not trap the session
101:		return allow, rpcErr
104:		return &hook.HookOutput{
109:	return allow, nil // pass / inconclusive ⇒ ALLOW
```

Six return `allow`; the seventh returns a composite literal. No path returns `(nil, nil)`, so the
arm is unreachable through the real handler and reaching it would require a production seam, which
REQ-CCR-007 forbids. Corroborating (not primary) signal — the raw profile shows the arm's block at
zero while its neighbours are covered:

```
codex_review_gate.go:198.2,198.16 1 1
codex_review_gate.go:198.16,200.3 1 0
codex_review_gate.go:201.2,201.47 1 1
```

### E.2.e AC-CCR-009 — the six new tests actually ran

Log: `.moai/reports/t519/sweep-six-tests.log`, rc=0, tree `a26feb41e`.

```
=== RUN   TestRunCodexReviewGate_InvalidStdinFailsOpen
--- PASS: TestRunCodexReviewGate_InvalidStdinFailsOpen (0.00s)
=== RUN   TestRunCodexReviewGate_EmptyStdinFailsOpen
--- PASS: TestRunCodexReviewGate_EmptyStdinFailsOpen (0.00s)
=== RUN   TestRunCodexReviewGate_HappyPathAllow
--- PASS: TestRunCodexReviewGate_HappyPathAllow (0.00s)
=== RUN   TestRunCodexReviewGate_HandlerErrorFailsOpen
--- PASS: TestRunCodexReviewGate_HandlerErrorFailsOpen (0.00s)
=== RUN   TestRunCodexReviewGate_BlockVerdictPropagates
--- PASS: TestRunCodexReviewGate_BlockVerdictPropagates (0.00s)
=== RUN   TestCodexSessionHandlePid
--- PASS: TestCodexSessionHandlePid (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.258s
```

`grep -c '^--- PASS: Test'` → `6` (no `$` end anchor: a Go PASS line ends in ` (0.00s)`, so an
anchored pattern would report zero on a passing run). The negative sentinel grep
(`no tests to run|no test files|^--- FAIL|^FAIL`) returned rc=1 — none present. The count is the
criterion, not the exit code: a selector typo yields fewer than six PASS lines while still exiting 0.

### E.2.f AC-CCR-010 — the mutant ledger

Full ledger with verbatim RED and GREEN output per row: `.moai/reports/t519/mutants.md`.

| Mutant | Target | Verdict |
|---|---|---|
| M1 | codex_review_gate.go:186 | RED — AC-CCR-001 only |
| M1b | codex_review_gate.go:188 | RED — AC-CCR-001 + AC-CCR-002 |
| M2 | codex_review_gate.go:191 | RED — AC-CCR-003, via the `t.Fatal` guard |
| M3a | codex_review_gate.go:195 | RED — AC-CCR-004 stderr assertion |
| M3b | codex_review_gate.go:196 | RED — AC-CCR-004 fail-open assertion |
| M4 | codex_review_gate.go:201 | RED — AC-CCR-005 |
| M5a | mcp_codex.go:702-704 | RED — AC-CCR-006, arrives as a panic (run isolated) |
| M5b | mcp_codex.go:703 | RED — AC-CCR-006 both zero-arms |
| M6 | codex_review_gate.go (trailing newline) | FIRED — AC-CCR-007 absence guard |

Eight mutants went RED; M6 is an absence guard with no test verdict, adopted by observing
`git status --short` list the file. **Two mutants deliberately did NOT fire, and both are recorded
rather than dropped**, because a non-firing mutant maps the guard's boundary:

- **M3-vac** — pre-recorded vacuous in spec.md §D.2, and here confirmed by measurement rather than
  taken on trust: the test PASSES under it, because `HandleCodexReviewGate` returns its empty
  `allow` alongside the error so both forms serialize to `{}`. Not counted as adoption evidence;
  AC-CCR-004 rests on M3a and M3b.
- **M2-cf** — a counterfactual added during the run, not in the plan ledger. With M2 applied and
  `withChangeDetector(t, true)` removed from the AC-CCR-003 fixture, the test PASSES. This measures
  the claim acceptance.md AC-CCR-003 makes in prose — that the detector swap is what keeps M2
  detectable — instead of asserting it. The fixture line is load-bearing; dropping it as redundant
  would make the criterion's sole adoption mutant vacuous. Both files were restored and re-hashed
  before proceeding.

Production-source integrity across the whole ledger: `internal/cli/codex_review_gate.go` SHA256
`9356669bdeb39f433897b7fc82c7e4cf036ca8ab24197301558b8c7bcedc3500` and `internal/cli/mcp_codex.go`
SHA256 `2a2ad2a9fd3a84566a372b01b3367223cd37fa0e71136f9543e2827e528ad8da`, captured before the first
mutation and re-observed identical after every revert. `git status --short` was read clean of
production sources before each of the three commits.

### E.2.g AC-CCR-011 — quality gate on the final tree

| Gate | Command | Observation | Classification |
|---|---|---|---|
| vet | `go vet ./internal/cli/` | rc=0 | clean |
| lint | `golangci-lint run --timeout=3m ./internal/cli/` | `0 issues.` | clean; identical to the pre-flight baseline, so zero NEW findings |
| format | `gofmt -l internal/cli/` | lists nothing | clean |
| build | `go build ./...` / `GOOS=windows GOARCH=amd64 go build ./...` | rc=0 / rc=0 | clean |

The inherited lint baseline was itself `0 issues.`, so "no NEW finding" here is a stronger statement
than usual: there were no pre-existing findings to hide a new one behind.

### E.2.h Subagent boundary grep — reported as observed, NOT empty

```
grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli | grep -v _test.go | grep -v '// '
```

Returns **32 rows**, rc=0. This is reported as-is rather than as a clean empty result. None is an
invocation and none is attributable to this SPEC:

- The rows are cobra `Long:` help prose (`harness.go`, `pr_watch_cmd.go`, `harness_mute.go`,
  `harness_clusters.go`, `harness/{propose,execute,install}.go`), agent-lint fixture files under
  `testdata/`, and the agent-lint detector itself (`agentlint/agent_lint.go`), whose job is to
  find the literal string. The `grep -v '// '` filter does not strip raw-string help text or
  tab-indented comments, which is why they survive the filter.
- **Attribution**: `git diff --name-only bf779ecf2..HEAD | grep '\.go$' | grep -v '_test\.go$'`
  returns 0 rows, so no non-test `.go` file changed on this branch. Every one of the 32 rows
  therefore exists byte-identically on the base tree — inherited, not introduced here. The two Go
  files this SPEC touched are both `_test.go` and are excluded by the filter by construction.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-07
run_commit_sha: a26feb41e   # M2; the M3 evidence commit is this record's own commit
run_status: complete
ac_pass_count: 11           # AC-CCR-001..011
ac_fail_count: 0
ac_deferred_count: 1        # AC-CCR-012 — sync-phase only, undecidable disposition
preserve_list_post_run_count: 0   # zero PRESERVE-list files modified
mutants_red: 8
mutants_recorded_non_firing: 2    # M3-vac (pre-recorded vacuous), M2-cf (fixture counterfactual)
coverage:
  run_codex_review_gate: "92.3%"   # baseline 0.0%, threshold >= 90.0%
  codex_session_handle_pid: "100.0%"  # baseline 66.7%, threshold 100.0%
  package_recorded_not_gated: "81.0% (go test) / 81.1% (cover -func)"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: pass
  windows_amd64: pass
total_run_phase_files: 5    # 2 test files, 2 SPEC artifacts, 3 evidence files under .moai/reports/t519 (mutants.md + 2 coverage + 1 sweep log)
m1_to_mN_commit_strategy: "one commit per milestone; M1 4ed5011c4, M2 a26feb41e, M3 this commit; no push (lane integrates through the lead's window)"
pushed: false
```

### Gaps — what was NOT observed

- **AC-CCR-012** was not exercised. It inspects the tree after the sync-close commit, which does not
  exist yet. It is DEFERRED, not passed.
- **The full repository suite was not run locally** (`go test ./...`), by constraint: parallel lanes
  doing so drove machine load to 413 on 2026-08-15. Verification was scoped to `./internal/cli/`;
  the full-suite verdict belongs to CI on a clean host.
- **No `-race` run** was performed. The six new tests spawn no goroutines, but the package as a
  whole was not re-measured under the race detector on this tree.
- **The adjacent under-covered functions were not measured for improvement**:
  `reviewableFromPorcelain` (83.3%) and `readHookInput` (85.7%) are unchanged and were declined as
  scope creep (plan §A.6 D6).
- **Coverage was measured once**, not repeated, so run-to-run variance in the package figure is
  unquantified. The two target rows are deterministic (they are exercised by named tests), so this
  gap bears on the package figure only.

### Residual risk — what could still be wrong

- **The S1 skip rests on a code reading of today's handler.** If a future change makes
  `HandleCodexReviewGate` able to return `(nil, nil)`, the arm becomes reachable and the proof in
  E.2.d silently stops holding. Nothing mechanical guards that.
- **`(&fakeCodexConn{}).pid()` returns a constant**, so AC-CCR-006's third arm proves the interface
  dispatch reaches a `codexProcessConn`, not that a real subprocess pid is read correctly. The real
  path is `TestRealCodexConnPid`'s subject, not this test's.
- **The M1 wiring tests mutate package-level seams without `t.Parallel()`**, which is correct today,
  but a future author adding `t.Parallel()` to any of them would introduce an order-dependent flake
  that a green local run would not surface. The constraint is stated in the file header comment;
  no mechanical guard enforces it.
- **The 92.3% figure is per-function.** It says nothing about the package's 81% or about the 76
  functions still at 0.0% across `internal/cli`, which remain out of scope.

## §E.4 Sync-phase Audit-Ready Signal

_&lt;pending sync-phase&gt;_

- 2026-09-07 plan-audit iter-2 PASS 0.9375 (`.moai/reports/t519/plan-audit-iter2.md`); remaining minor F5 (acceptance.md:60 M1→M1b) and F6 (plan.md:37, spec.md:148 four→five files) fixed by the lane directly (one-word edits; distinct-file count re-verified = 5). Advisory A6 (withCodexSession overwrites codexLookPath — never combine it with a t.Fatal LookPath guard in one test) carried into the run-phase delegation as a constraint.

## §F Phase 4 Mode Selection

Decision: serial (Implementation Kickoff Approval obtained 2026-09-07 via lead; autonomous progression)

Input parameters: tier M · scope 2 files (1 new test file + 1 test-file addition) · domains 1 (Go test code, internal/cli) · language mix 100% Go test · concurrency benefit LOW (coding-heavy, sequential mutant discipline) · Agent Teams: not requested.

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | not a typo-level change; needs mutant RED/GREEN cycles and evidence capture |
| serial | **yes** | coding-heavy Tier M, single domain, 2 files — Anthropic coding-task parallelism caveat |
| fanout | no | 1 domain, 2 files — below the ≥3 domains / ≥10 files auto-select threshold |
| sweep | no | not mechanical bulk transformation |

Decision: serial

Justification: one manager-develop (cycle_type=tdd) drives M1 → M2 → M3 sequentially; each milestone depends on the previous mutant ledger state and the union gate, so concurrent spawns would race on the same package and the same evidence files. Implementation Kickoff Approval: obtained 2026-09-07 via the lead session (operator decision, autonomous progression). Logged at HEAD c6bf21a72 before the first run-phase spawn.
