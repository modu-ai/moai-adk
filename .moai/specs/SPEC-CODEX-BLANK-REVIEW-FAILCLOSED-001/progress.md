# Progress — SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001

Card: t551 · Branch: `WT-audit-fail-open` · Plan-phase HEAD: `3ac58b5a1`

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M).
- SPEC ID regex self-check: executed as Bash, output `PASS`.
- ID uniqueness: no existing SPEC directory matches `BLANK` / `FAILCLOSED` / `FAIL-OPEN`.
- Every file:line citation in the artifacts was re-measured against this tree at
  HEAD `3ac58b5a1` before being written. Two citations supplied by the dispatch
  were corrected by measurement: the `runTurn` guard is at
  `internal/cli/mcp_codex.go:817` (not `:818`), and the pinned synthesizer case
  is at `internal/cli/codex_review_rpc_test.go:119` (not `:120`).
- `moai spec lint --strict .moai/specs/SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001/spec.md`
  → `✓ No findings — all SPEC documents are valid`, exit 0.
  The invocation covers the whole SPEC directory, not spec.md alone: an earlier
  run of the same command reported findings in `plan.md:5` and `acceptance.md:5`,
  which is the catch-all control showing the linter reaches all three artifacts.
  Two authoring defects were found and repaired by that run:
  (1) `tags:` was authored as a YAML sequence, but `internal/spec/lint.go:511`
      declares `Tags string` — corrected to a quoted CSV;
  (2) `status:` was present in `plan.md` and `acceptance.md`, which
      `ArtifactStatusFieldForbidden` rejects — lifecycle state lives in `spec.md`
      alone. Removed from both.
- Open decision carried into run-phase: `plan.md` §B.1 (option a vs option b for
  the pinned synthesizer test). Recommendation recorded; not settled.

### Plan-audit repair pass (FAIL 0.81 → repaired)

Six blocking findings addressed in one edit pass. Two were FALSE STATEMENTS I
authored; both were independently re-traced against the code in this tree before
rewriting, not accepted on report alone.

| Finding | Repair |
|---|---|
| D2 — AC-CBR-003's discriminating power overstated | Restated: `:1253` ALONE is load-bearing; it says nothing about `:1134`/`:1137`. Covers REQ-CBR-003 only. |
| D3 — REQ-CBR-001/002 had zero coverage | New AC-CBR-009 pins the OVERWRITE case (later blank clobbers earlier real, loop `:1125-1139`). |
| D4 — headline overstated the effect | §A and §C.3 rewritten: `inconclusive` does NOT close the gate hole (`default:` arm → claude anchor). The scoped win is an honest codex row, not a blocked gate. §E reconciled. |
| D5/D7 — old AC-CBR-009 vacuous-capable + mandated the helper | Demoted out of the AC set into `plan.md` §E as a checklist item; helper is now explicitly RECOMMENDED, and no AC mandates implementation shape. |
| D6 — orphan AC naming no REQ | Resolved by the demotion; the AC-CBR-009 slot now covers REQ-CBR-001/002. |
| D8/D9/D10 | `acceptance.md` §D→§A (2 sites); test range `:113-127`→`:114-126`; auditor report citations corrected to `:817`, `:1409`, `:276`/`:1446`, `:291`, `:1608`. |

Also corrected: the §D RED list now includes AC-CBR-009 and states why
AC-CBR-002/004/006/008 are excluded (preservation pins must be green both sides).

REQ↔AC mapping checked directly: REQ-CBR-009 IS referenced, by AC-CBR-007
(`acceptance.md:123`). The relayed `CoverageIncomplete` finding was an artifact
of a wrong lint invocation form and is not a real defect.

Process defect recorded (not mine to undo): the lint-repair rewrites of
2026-09-08 13:44:27-13:44:40 landed inside an audit window opened 13:36.

### Plan-audit iteration 2 (PASS 0.93) — final plan-phase edit

Tier M's 2-iteration audit ceiling is spent; these landed without re-audit.

| Finding | Repair |
|---|---|
| F1 (blocking) — AC-CBR-009 admitted a first-wins mutant | Added one `**And given**` clause pinning the BLANK-FIRST ordering for both item types. Criterion not restructured. |
| F2 — AC-CBR-003's RED reason unstated | §D now names what must be asserted beyond the verdict: the value returned by `bestCodexReviewText`, since the verdict half is already `pass` pre-repair and would be a vacuous green. |
| F5 | `codexFindingsOf` citation `:1465`→`:1473` (re-measured; `:1465` is the first line of its doc comment, not the func). |
| F6 | `acceptance.md:112`→`:123` (re-measured, not offset). |

F3 and F4 were classed optional and are deliberately NOT addressed — the AC set
is closed.

**Residual risk (F4, unpinned).** No criterion pins the PREFERENCE ordering in
`bestCodexReviewText` (`internal/cli/mcp_codex.go:1252-1257`) — that a non-blank
structured `exitedReviewMode.review` wins over a non-blank `agentMessage.text`.
An implementation that reversed that preference would satisfy every criterion in
this SPEC, because each one exercises a case where exactly one of the two is
non-blank. The preference is pre-existing behavior this SPEC does not modify, so
the exposure is a silent regression during run-phase editing of `:1253`, not a
defect introduced here. Run-phase should keep the preference intact when making
that site blank-aware.
- Unobserved items are recorded in `spec.md` §E Gaps, not asserted in the body.

**Plan-audit closed.** Iteration 1 FAIL 0.81 → iteration 2 PASS 0.93 (Tier M
threshold 0.80, delta +0.12, so the score-regression STOP clause did not fire).
Both verdicts are recorded at `.moai/reports/t551/plan-audit-verdict.md`
(iteration 2 appended at line 292, iteration 1 preserved).

## §F Phase 4 Mode Selection

**Decision: serial**

Input parameters — tier: M · scope: 1 implementation file
(`internal/cli/mcp_codex.go`) plus test files · domain count: 1 (Go source,
`internal/cli`) · file language mix: 100% Go · concurrency benefit: LOW.

| Mode | Selected | Rationale |
|---|---|---|
| `direct` | no | Not trivial — four coupled sites, a RED-first contract, and a nine-criterion acceptance suite. |
| `serial` | **yes** | Coding-heavy single-package work; one sequential `manager-develop` spawn per milestone. |
| `fanout` | no | Single domain, single file. Fan-out would add reconciliation cost with nothing to parallelize, and Anthropic's coding-task parallelism caveat points at `serial` for coding work regardless. |
| `sweep` | no | Far below the ~30-file mechanical threshold, and the transform is not uniform — four sites with different surrounding logic. |

Justification: every criterion for the alternatives fails on the same fact —
this is one file in one package, and the four edit sites are coupled (the guard's
correctness depends on the selection site, which depends on the collection
sites). Splitting coupled edits across concurrent agents would create exactly the
write race the concurrency safeguard forbids, for no wall-clock gain.

Implementation Kickoff Approval: PASSED (operator, this session). Progression
mode: autonomous — report once at completion with the evidence gathered.

## §E.2 Run-phase Evidence

Run-phase HEAD at entry: `3ac58b5a1` · branch `WT-audit-fail-open` · card t551.

### M1 — decisions settled and the characterization baseline

**plan.md §B.1 decision: OPTION (a).** The repair lands at the guard
(`internal/cli/mcp_codex.go:817`), the two collection sites (`:1134`, `:1137`),
and the selection site (`:1253`). `synthesizeReviewOutput` and
`codexUnrecognizedVerdict` are NOT touched, so the pinned case at
`internal/cli/codex_review_rpc_test.go:119` (`"": "pass"`) stays byte-identical
and untouched.

Rationale, stated as the plan states it: the defect is REACHABILITY, not
synthesis. The synthesizer's answer for an unrecognized body is a documented
mode-keyed policy (REQ-CBR-007) that this card preserves; the bug is that an
ABSENT body is routed to it at all. Repairing the routing removes the blank
body from the synthesizer's input set without moving the synthesizer's
contract. Option (b) was rejected because it edits a deliberately-placed record
and widens the blast radius to every caller of a function this card has not
surveyed, for a case the guard already makes unreachable.

**plan.md §B.2 decision: the helper IS adopted.** `codexReviewTextIsBlank`
(`internal/cli/mcp_codex.go`) is a one-line wrapper over
`strings.TrimSpace(s) == ""`, called at all four sites. The plan marks it
RECOMMENDED-not-mandated and notes it earns its keep only through the
miss-one-site argument — that argument is exactly this defect's shape (three
sites written with the same wrong test, a fourth added later), so the wrapper
is adopted. No acceptance criterion mandates it; it remains an implementation
choice.

**U+00A0 decision: IN SCOPE, and asserted.** `strings.TrimSpace` cuts on
`unicode.IsSpace`, which includes U+00A0, so a body of non-breaking spaces
alone is blank. `acceptance.md` §C requires the behavior be asserted either
way; `TestCodexBlankReview_BlankDiscriminator` pins it rather than leaving it
implicit.

**Characterization baseline (M1, measured BEFORE any repair).**
`internal/cli/codex_blank_review_characterization_test.go` pins the pre-repair
verdict + Summary of the review-text path per fixture class, and pins the
unavailable-backend path as the permanent AC-CBR-004 control. Command and
verbatim output:

```
$ go test -run 'TestCharacterize_UnavailableBackend|TestCharacterize_ReviewTextPathPreRepair' -v ./internal/cli/
=== RUN   TestCharacterize_UnavailableBackend
--- PASS: TestCharacterize_UnavailableBackend (0.00s)
=== RUN   TestCharacterize_ReviewTextPathPreRepair
    --- PASS: TestCharacterize_ReviewTextPathPreRepair/space-only (0.00s)
    --- PASS: TestCharacterize_ReviewTextPathPreRepair/newline-only (0.00s)
    --- PASS: TestCharacterize_ReviewTextPathPreRepair/mixed-whitespace (0.00s)
    --- PASS: TestCharacterize_ReviewTextPathPreRepair/exactly-empty (0.00s)
    --- PASS: TestCharacterize_ReviewTextPathPreRepair/real-clean-review (0.00s)
    --- PASS: TestCharacterize_ReviewTextPathPreRepair/finding-bullet (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.932s
```

Baseline-attribution: this run, this tree, HEAD `3ac58b5a1`. The three
whitespace rows pin `verdict="pass" summary=""` — the defect, observed here as
a green characterization rather than asserted from the plan-phase probe.

### M2 RED — the acceptance suite fails on the unrepaired code

The blankness seam (`codexReviewTextIsBlank`, `codexBlankReviewSummary`,
`blankReviewInconclusive`) was landed UNWIRED in the same commit as the
acceptance suite, deliberately: with the seam present the package compiles, so
the RED is a BEHAVIORAL failure on the unrepaired path rather than a build
error. A build error proves only that a symbol is missing.

Command, and the verbatim result (full output at
`.moai/reports/t551/red-evidence-20260908.txt`, which carries the HEAD SHA on
its last line):

```
$ go test -count=1 -v -run 'TestCodexBlankReview_' ./internal/cli/
=== RUN   TestCodexBlankReview_AC001_BlankBodyDoesNotSynthesizePass/space-only
    codex_blank_review_test.go:109: verdict = pass for a blank body — a review that produced no verdict was reported as one that found nothing wrong
    codex_blank_review_test.go:112: verdict = "pass", want "inconclusive"
    codex_blank_review_test.go:115: blank output must surface its cause alongside the fail-open struct
--- FAIL: TestCodexBlankReview_AC001_BlankBodyDoesNotSynthesizePass (0.00s)
    --- FAIL: .../space-only (0.00s)
    --- FAIL: .../newline-only (0.00s)
    --- FAIL: .../mixed-whitespace (0.00s)
    --- PASS: .../exactly-empty (0.00s)
--- PASS: TestCodexBlankReview_AC002_RealCleanReviewStillPasses (0.00s)
    codex_blank_review_test.go:154: bestCodexReviewText(blank, real) = "   \n", want the agent text "The change introduces no blocking issues." — a blank structured review must fall through
    codex_blank_review_test.go:172: summary = "", want "The change introduces no blocking issues." — the real agent text must be the SELECTED text
--- FAIL: TestCodexBlankReview_AC003_BlankStructuredReviewDoesNotShadowAgentMessage (0.00s)
--- PASS: TestCodexBlankReview_AC004_UnavailableBackendStillFailsOpen (0.00s)
    codex_blank_review_test.go:220: both states must be inconclusive; blank="pass" unavailable="inconclusive"
--- FAIL: TestCodexBlankReview_AC005_BlankAndUnavailableAreDistinguishable (0.00s)
--- PASS: TestCodexBlankReview_AC006_BlankBodyIsNeverFail (0.00s)
    codex_blank_review_test.go:266: verdict = "pass", want "inconclusive"
    codex_blank_review_test.go:269: gate_unmet is empty — a required gate whose backend produced no verdict text must be recorded as unmet
--- FAIL: TestCodexBlankReview_AC007_RequiredGateAnnotatesBlankOutput (0.00s)
--- PASS: TestCodexBlankReview_AC008_NonBlankUnrecognizedBodyUnchanged (0.00s)
    codex_blank_review_test.go:338: summary = "", want the surviving real text "The change introduces no blocking issues."
--- FAIL: TestCodexBlankReview_AC009_BlankItemDoesNotClobberRealOne (0.00s)
    --- FAIL: .../review:_real_then_blank (0.00s)
    --- PASS: .../review:_blank_then_real_(kills_first-wins) (0.00s)
    --- FAIL: .../agent:_real_then_blank (0.00s)
    --- PASS: .../agent:_blank_then_real_(kills_first-wins) (0.00s)
    codex_blank_review_test.go:372: state A verdict = pass, want NOT pass
    codex_blank_review_test.go:397: state A summary = "", want it to name blank review output
--- FAIL: TestCodexBlankReview_ThreeStateControlMatrix (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.867s
```

Baseline-attribution: this run, this tree, HEAD `a87a7d8a7` (the M1 commit —
the four sites still unrepaired).

Two properties of that RED are load-bearing:

- The preservation pins AC-CBR-002 / 004 / 006 / 008 are GREEN pre-repair, as
  `acceptance.md` §D requires. A pin that went red here would mean the suite
  was asserting a change the SPEC forbids.
- AC-CBR-009's two REVERSED orderings (blank arriving before real) PASS
  pre-repair. That is expected and is the criterion working as designed: those
  rows exist to kill a first-wins mutant, not to go red. The two
  real-then-blank rows are the ones that go red, and they are the ones the
  overwrite defect actually produces.

### M2 + M3 GREEN — the repair

All four sites now share ONE discriminator. Verbatim result (full output at
`.moai/reports/t551/green-evidence-20260908.txt`):

```
$ go test -count=1 -v -run 'TestCodexBlankReview_|TestCharacterize_UnavailableBackend|TestCharacterize_ReviewTextPath|TestSynthesizeReviewOutput_FindingBulletsMapToFail' ./internal/cli/
--- PASS: TestCharacterize_UnavailableBackend (0.00s)
--- PASS: TestCharacterize_ReviewTextPath (0.00s)
--- PASS: TestCodexBlankReview_BlankDiscriminator (0.00s)
--- PASS: TestCodexBlankReview_AC001_BlankBodyDoesNotSynthesizePass (0.00s)
--- PASS: TestCodexBlankReview_AC002_RealCleanReviewStillPasses (0.00s)
--- PASS: TestCodexBlankReview_AC003_BlankStructuredReviewDoesNotShadowAgentMessage (0.00s)
--- PASS: TestCodexBlankReview_AC004_UnavailableBackendStillFailsOpen (0.00s)
--- PASS: TestCodexBlankReview_AC005_BlankAndUnavailableAreDistinguishable (0.00s)
--- PASS: TestCodexBlankReview_AC006_BlankBodyIsNeverFail (0.00s)
--- PASS: TestCodexBlankReview_AC007_RequiredGateAnnotatesBlankOutput (0.01s)
--- PASS: TestCodexBlankReview_AC008_NonBlankUnrecognizedBodyUnchanged (0.00s)
--- PASS: TestCodexBlankReview_AC009_BlankItemDoesNotClobberRealOne (0.00s)
--- PASS: TestCodexBlankReview_ThreeStateControlMatrix (0.00s)
--- PASS: TestSynthesizeReviewOutput_FindingBulletsMapToFail (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.067s
```

The last line is the plan.md §B.1 check: the pinned case at
`internal/cli/codex_review_rpc_test.go:119` (`"": "pass"`) is UNTOUCHED and
still green, because option (a) left `synthesizeReviewOutput` alone.

### M4 — three-state control matrix and gate visibility

Observed in ONE run by `TestCodexBlankReview_ThreeStateControlMatrix`, which
asserts three DISTINGUISHABLE signals rather than three verdicts:

| State | Input | verdict | Summary |
|---|---|---|---|
| A blank | `" \n\t "` | `inconclusive` | `codex review output was blank: no verdict text was produced` |
| B reviewed, clean | `The change introduces no blocking issues.` | `pass` | the review prose verbatim |
| C backend unavailable | session start fails | `inconclusive` | `codex unavailable: codex session start failed: …` |

A and C share the verdict and are separated by the Summary alone; the test
compares the `verdict|summary` pair of all three and fails on any collision, so
a future edit that collapses A into C is caught by construction rather than by
a reviewer noticing. Gate visibility (REQ-CBR-009) is measured separately by
`TestCodexBlankReview_AC007_…`, which drives the real `codex_audit` tool against
a project declaring `workflow.audit.gates.codex: required` and observes a
non-empty `gate_unmet`.

### M5 — the spec.md §E convergence-layer gap, closed by measurement

`TestCodexBlankReview_M5_InconclusiveDoesNotBlockTheConvergenceLayer` runs the
pure `converge()` over a required claude `pass` + required codex `inconclusive`
and observes `overall_verdict = pass` with codex present in
`fail_open_backends`. Command and result:

```
$ go test -count=1 -v -run 'TestCodexBlankReview_M5_' ./internal/cli/
--- PASS: TestCodexBlankReview_M5_InconclusiveDoesNotBlockTheConvergenceLayer (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.882s
```

The plan-phase static trace is therefore confirmed BY EXECUTION rather than
inherited: this repair changes what codex reports about itself, not what the
convergence layer decides. The gate hole stays open and belongs to #1632 axis 3.

### Residual-site sweep (plan.md §E checklist, with its positive control)

The plan requires the sweep grep be shown to MATCH on the pre-repair tree
before a zero-match run is read as evidence — an unmatched grep is not evidence.
Both halves, run against the same pattern set:

```
$ git show 3ac58b5a1:internal/cli/mcp_codex.go > /tmp/t551-pre.go
$ /usr/bin/grep -n 'p\.Item\.Review != ""\|p\.Item\.Text != ""\|if review != ""\|reviewText == ""' /tmp/t551-pre.go
817:	if reviewText == "" {
1134:			if p.Item.Type == "exitedReviewMode" && p.Item.Review != "" {
1137:			if p.Item.Type == "agentMessage" && p.Item.Text != "" {
1253:	if review != "" {
exit=0

$ /usr/bin/grep -n '<same pattern>' internal/cli/mcp_codex.go
exit=1   (no output)

$ /usr/bin/grep -n 'codexReviewTextIsBlank' internal/cli/mcp_codex.go
356:// codexReviewTextIsBlank is the ONE emptiness discriminator the codex
367:func codexReviewTextIsBlank(s string) bool {
855:	if codexReviewTextIsBlank(reviewText) {
1184:			if p.Item.Type == "exitedReviewMode" && !codexReviewTextIsBlank(p.Item.Review) {
1187:			if p.Item.Type == "agentMessage" && !codexReviewTextIsBlank(p.Item.Text) {
1308:	if !codexReviewTextIsBlank(review) {
```

4 matched before, 0 after, and exactly 4 call sites of the single
discriminator. `/usr/bin/grep` is used explicitly: the shell's `grep` here is a
ugrep wrapper that can skip files silently, and a silent skip would make the
zero-match half meaningless.

### Full-suite, coverage, builds, lint (no `-run` narrowing)

```
$ go test -count=1 -coverprofile=/tmp/t551-cover.out ./internal/cli/...
ok  	github.com/modu-ai/moai-adk/internal/cli	500.411s	coverage: 81.4% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/agentlint	0.464s	coverage: 86.7% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/harness	9.026s	coverage: 80.9% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/pr	2.917s	coverage: 91.7% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/preference	1.537s	coverage: 85.8% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/printer	1.684s	coverage: 97.0% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/specid	5.579s	coverage: 100.0% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/taskledger	4.840s	coverage: 92.7% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/uikit	5.230s	coverage: 98.8% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/update	1.148s	coverage: 88.9% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	3.469s	coverage: 90.1% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/update/deploy	4.715s	coverage: 91.6% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	3.802s	coverage: 92.1% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/update/plan	6.262s	coverage: 95.0% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/update/report	1.757s	coverage: 92.9% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	8.803s	coverage: 92.0% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	9.788s	coverage: 87.1% of statements
exit=0
```

Per-function coverage of the functions this card added or edited, read from the
same profile:

```
$ go tool cover -func=/tmp/t551-cover.out | grep -E 'codexReviewTextIsBlank|blankReviewInconclusive|inconclusiveReview|bestCodexReviewText|runTurn|awaitCodexTurnReview'
mcp_codex.go:324:  inconclusiveReview             100.0%
mcp_codex.go:333:  inconclusiveReviewWithSummary  100.0%
mcp_codex.go:352:  blankReviewInconclusive        100.0%
mcp_codex.go:367:  codexReviewTextIsBlank         100.0%
mcp_codex.go:825:  runTurn                         85.7%
mcp_codex.go:1136: awaitCodexTurnReview            86.2%
mcp_codex.go:1307: bestCodexReviewText            100.0%
```

Builds and lint:

```
$ go build ./...                            → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...  → exit 0
$ go vet ./internal/cli/                    → exit 0 (no output)
$ golangci-lint run --timeout=5m ./internal/cli/...
internal/cli/todo.go:97:13: Error return value of `fmt.Fprintf` is not checked (errcheck)
1 issues:
* errcheck: 1
exit=1
```

The single lint issue is PRE-EXISTING and byte-identical to the baseline
captured before any edit in this card (Section C pre-flight, same command, same
one finding at `internal/cli/todo.go:97`). NEW issues introduced by this card: 0.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-08
run_commit_sha: pending-backfill-run
run_status: complete
ac_pass_count: 9
ac_fail_count: 0
preserve_list_post_run_count: 3   # synthesizeReviewOutput, codexUnrecognizedVerdict,
                                  # codex_review_rpc_test.go:119 — all untouched
l44_pre_commit_fetch: not-run     # lane does not push; no origin interaction this phase
l44_post_push_fetch: not-run      # no push performed (lead owns the develop push)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  host_darwin_arm64: pass
  windows_amd64: pass
total_run_phase_files: 3          # internal/cli/mcp_codex.go
                                  # internal/cli/codex_blank_review_test.go
                                  # internal/cli/codex_blank_review_characterization_test.go
m1_to_mN_commit_strategy: one commit per milestone group on WT-audit-fail-open
                          (M1 characterization, M2 RED, M2+M3 repair, M5
                          measurement, final evidence) — no amend, no
                          force-push, no push
```

### Gaps — what run-phase did NOT observe

- **No pre-repair package coverage baseline was measured.** The 81.4% figure for
  `internal/cli` is a post-repair reading with no attributable delta; this card
  adds only tests to that package, so the number cannot have fallen, but that is
  an argument, not a measurement. The per-function figures above ARE attributable
  measurements of the repaired sites.
- **No live `codex` binary was exercised**, exactly as in plan-phase. Every
  fixture drives the production `runCodexReviewRPC` path over a stubbed
  connection. `spec.md` §E's second gap therefore stands unchanged.
- **The real-world frequency of a whitespace-only body from codex-cli is still
  unmeasured** (`spec.md` §E third gap, unchanged). The defect is established by
  reachability, not by field incidence.
- **The subagent-boundary grep is not zero at the repository baseline.** The
  Section E form
  (`grep -rn 'AskUserQuestion' internal/cli | grep -v "_test.go" | grep -v "// "`)
  returns pre-existing matches: `internal/cli/testdata/agent_lint/*.md` fixtures
  and the help/doc strings of `harness.go`, `agentlint`, `pr_watch_cmd.go` and
  friends, all of which DESCRIBE the boundary rather than cross it. Scoped to the
  three files this card touched, the grep returns exactly one match — a doc
  comment at `mcp_codex.go:1596` stating the function never invokes it. Zero
  invocations were added.

### Residual risk

- The guard is now the only thing between a blank body and a `pass`, because
  option (a) deliberately left `synthesizeReviewOutput` lenient. A future caller
  reaching the synthesizer directly with a blank body would still get a `pass`.
  That is the acknowledged cost of option (a), recorded in plan.md §B.1 under
  *Against*, and it is bounded by the fact that the guard is the only production
  route into the synthesizer on this path.
- `strings.TrimSpace` decides blankness, so a body of exotic Unicode formatting
  characters that `unicode.IsSpace` does NOT classify as space (a zero-width
  space, U+200B, for instance) still counts as content and reaches the
  synthesizer. U+00A0 IS covered and is asserted; U+200B is not, and no
  criterion pins it.
- The repair does not close the gate hole (measured in M5, not assumed). A
  required codex gate still fails open to the claude anchor.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-08
sync_commit_sha: b6428562c   # the sync commit carrying the CHANGELOG entry, the frontmatter
                             # close, and this §E.4. It landed with the canonical
                             # `pending-backfill-sync` placeholder here — a commit cannot
                             # cite its own hash — and this line is the backfill (D3
                             # exemption, spec-frontmatter-schema.md § SHA placeholder
                             # backfill exemption)
sync_backfill_commit_sha: self   # this line's own commit, the immediate successor of
                                 # b6428562c in `git log`; it likewise cannot name itself
sync_status: completed
b12_self_test_a: pass    # pre-emission duplicate grep —
                         # `grep -c 'SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001' CHANGELOG.md`
                         # → 0 before writing (rc=1, no match)
b12_self_test_b: pass    # AC count match — 9 distinct AC identifiers in acceptance.md
                         # (AC-CBR-001..009), non-zero, equal to §E.3 ac_pass_count 9
b12_self_test_c: pass    # every path cited in the CHANGELOG entry verified with `ls`:
                         # 9 paths, all present (3 internal/cli sources +
                         # mcp_convergence.go + 5 under .moai/reports/t551/)
changelog_entry_position: "[Unreleased] › ### Fixed, topmost entry (immediately above SPEC-TODO-HOME-TEMP-GUARD-001)"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (single sync commit; updated: 2026-09-08, already current — no byte changed on that key)"
  plan_md: "no status field — not a transition target (ArtifactStatusFieldForbidden); updated already 2026-09-08"
  acceptance_md: "no status field — not a transition target; updated already 2026-09-08"
  progress_md: "no frontmatter — not a transition target"
canary_compliance_check: not_applicable   # this SPEC defines no forward-looking policy
                                          # for its own sync to test
verification_tree: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t551
verification_branch: WT-audit-fail-open
verification_head_at_measurement: b2f7fdd63
docs_site_surface: related-but-not-stale   # measured, not assumed — see the note below
```

### What this sync execution measured for itself

Two claims in the CHANGELOG entry were re-measured here rather than inherited from
the SPEC body, because both are the kind that decays:

| Claim | Command | Observed |
|---|---|---|
| The four call sites route through the one discriminator | `grep -n 'codexReviewTextIsBlank(' internal/cli/mcp_codex.go` | `:855`, `:1184`, `:1187`, `:1308`; definition at `:367` — matching the dispatch exactly |
| The pinned synthesizer test is byte-identical | `git diff --stat 3ac58b5a1..HEAD -- internal/cli/codex_review_rpc_test.go` | empty output, rc=0 |
| The convergence default arm still reaches the claude anchor | `sed -n '185,200p' internal/cli/mcp_convergence.go` | `default:` → `claudeVerdictOrDefault(verdicts, overallVerdictPass)` — the non-closure holds as written |

### Gaps — what this sync execution did NOT observe

- **No test, build, lint, or coverage command was re-run in this sync phase.** The
  verification figures in the CHANGELOG entry (9/9 AC, suite exit 0, lint delta 0,
  cross-platform build) are the run-phase measurements recorded in §E.2/§E.3 and
  the `.moai/reports/t551/` evidence files, cited as such. They are attributed to
  run-phase, not re-measured here, and this sync makes no independent claim about
  them.
- **No CI verdict exists.** This lane does not push, so nothing has been measured
  in a clean environment or on the windows/linux matrix. `GOOS=windows go build`
  compiles non-test code only — it establishes compilability, not that the tests
  compile or pass there.
- **`run_commit_sha` in §E.3 above is still the `pending-backfill-run` placeholder.**
  That field belongs to manager-develop's §E.3 surface and was not modified here;
  it is reported as an owed backfill rather than silently completed.
### docs-site: a related page exists, and it is reported rather than edited

docs-site was out of scope for this card by dispatch. It was searched anyway, so
that the scope call rests on a measurement instead of an assumption:
`/usr/bin/grep -rlni 'codex' docs-site/content/` returns **40** files (the plain
`grep` on this shell is a `ugrep` wrapper that can skip silently, so the absolute
path was used).

The one page that touches this contract is `advanced/multi-model-audit.md`
(4 locales). Two of its sentences were read in full:

- `:80` documents the diff-collection-failure cause of `inconclusive`. This card
  adds a **second** cause (a blank review body). The page does not present its
  causes as a closed set, so nothing on it becomes false — it becomes less
  complete.
- `:82` states that convergence reads a self-declared `inconclusive` as
  inconclusive and never synthesizes it into a PASS. That remains true of the
  backend's own recorded verdict, which is what the sentence is about, and this
  card does not touch convergence at all.

**Verdict: related, not stale — no page is edited.** Adding the second cause to
that page is a documentation-completeness follow-up, deliberately not folded into
a sync commit scoped to an internal classification change. Recorded here so the
next reader does not have to re-derive it.

- **README was not searched.** Only `docs-site/content/` was grepped; the four
  README locales were not, so no claim is made about them in either direction.
