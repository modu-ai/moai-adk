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
| `4bea753cb` | v0.2.4 (manager-spec) | candidate (b) mechanism authored; AC-CPS-004/008 amended; AC-CPS-016 added |
| `861e08912` | M4 (b) | native format pin (`codexReviewSessionParams` → `developerInstructions`), native fall-through → `inconclusive`, pin recognition tests, stale no-signal-native-pass witnesses updated to the amended expectation |

**AC matrix.**

| AC | Status | Command | Observed |
|---|---|---|---|
| AC-CPS-004 (b) | **PASS** | `go test -count=1 -run 'TestCodexNative' ./internal/cli/` at `861e08912`; RED-now re-executed on `4bea753cb` (output identical to the `0ff644530` cell) → `m4-rednow.log` | All three Then clauses hold: (1) `TestCodexNativePinnedPassStaysPass` — pinned `Verdict: pass` body → `pass`/0, no contradiction; (2) `TestCodexNativeNoSignalBodyYieldsInconclusive` — three no-signal bodies → `inconclusive` (RED on the pre-GREEN tree: `Verdict = "pass", want "inconclusive"`, `m4-red.log`); (3) `TestCodex1718Fixtures` S1/S2/S2p rows unchanged (`fail`/3, `fail`/1, `fail`/1 on both paths). Pin carrier: `TestCodexNativeRequest_CarriesFormatPin` — native `thread/start` carries `developerInstructions` with the pin; `review/start` params stay schema-clean; adversarial `thread/start` carries none. Mutants M-A and M-B both executed live and killed (`m4-mutants.log`: M-A pinned-pass guard FAIL `got inconclusive/0, want pass/0`; M-B the no-signal bodies avoiding the word fail FAIL `Verdict = "pass", want "inconclusive"`). RED captured before GREEN in two stages: `undefined: codexNativeReviewFormatPin` build failure, then the carrier+downgrade assertion failures |
| AC-CPS-005 (c) | PASS | `go test -count=1 -v -run 'Contradiction' ./internal/cli/` at `562126b1f` (M2); re-run on HEAD `861e08912` → `m4-ac013.log`, 4/4 PASS | V8 body flagged on both paths; predicate table incl. the unmet-gate control; end-to-end control through `applyGateUnmet` not flagged — `ok` |
| AC-CPS-006 (a) | PASS | `-run 'TestCodex1718\|TestCodexWidened'` at `ce1df7f6f` (M2); re-run on HEAD `861e08912` → `ok` | exact count + severity/message/file/line for S1 (3), S2 and S2p (1) on both paths; partial drift → 2 and 0 survivors, exact count fails; `codex_findings_parse_test.go` (V1) unchanged and passing |
| AC-CPS-007 | PASS | `git diff 4bea753cb -- internal/cli/codex_findings_parse_test.go \| wc -l` (M4); `git diff e26633699 …` was `0` at M2/M3 | `0` |
| AC-CPS-008 | PASS | `-run TestGuard_` (M3; re-run on HEAD `861e08912` → `ok`) | guard fixture bodies state the pinned `Verdict: pass` line (scope note, v0.2.4); assertion unchanged — `pass`/0, no contradiction. M3 mutant "drop the verdict term from the predicate" made the guard fail (`guard-mutants.log`); M4's mutant M-A re-proved it live on the new bodies |
| AC-CPS-009 | PASS (keep) | `-run TestGuard_` (M3; re-run on HEAD → `ok`) | N1, N2 and the unrecognized AC-CVS-001 corpus members stay `inconclusive` on the adversarial path — unchanged by M4 (the downgrade now also covers native, which is AC-CPS-004's point, not a change to this handling) |
| AC-CPS-010 | PASS | `git diff 4bea753cb -- internal/cli/mcp_codex.go \| grep -cE '^[+-].*(NextSteps\|next_steps)'` (M4) → `0`; M2 form `git diff e26633699 …` was `0` (probe on `Contradiction` → `8`) | `0` |
| AC-CPS-011 | PASS | `go test -count=1 -v -run '^TestCodex1718Fixtures$' ./internal/cli/` at `6a89f3d83` (M2); structural properties + P10 mutants re-run on HEAD → `ok` (`m4-green.log`, `m4-blast2.log`) | exit 0, ten SYNTH rows (pre-M4 values at the time); ledger P1–P11 at their pass conditions (`ac011-ledger.log`) |
| AC-CPS-012 | PASS | same selector at `ce1df7f6f` (M2); re-run on HEAD → `ok`, S rows unchanged in verdict and findings (`m4-green.log`) | S1 `fail`/3, S2 `fail`/1, S2p `fail`/1 on both paths with exact content. RED before GREEN in `red-a.log`. **M4 before/after rows:** N1/N2 review/start `pass`/0 (before, `m4-rednow.log`) → `inconclusive`/0 (after, `m4-green.log`); turn/start rows unchanged |
| AC-CPS-013 | PASS (remeasured on HEAD under the v0.2.3 re-anchored wording) | `-run 'Contradiction'` on HEAD `861e08912` → `m4-ac013.log`, `--- PASS` ×4, exit 0 | `TestSynthesizeReviewOutput_V8ContradictionIsReported` holds: the V8-shaped body is reported self-contradictory on both paths; the S2p-at-`562126b1f` clause stands as recorded at M2. The M4 downgrade does not touch it — a V8 body states a blocking verdict, which `codexStatedVerdict` reads before any fall-through |
| AC-CPS-014 (d) | **NOT MEASURED — authorized, attempted, blocked** | `mcp__moai__codex_audit` (mode=adversarial, target=baseBranch, project_root=this worktree) on 2026-09-26 | authorized by the run-resume decision (§E.1, `f1bd21fc4`); attempt 1 blocked BEFORE any body was produced — account usage limit until 2026-09-28 14:37, and the running MCP server binary (`a8a9b9376`) predates the (d) pin commit `83046be7c` (pin string grep: 0 in binary tree vs 1 in card HEAD), so a call would exercise the unpinned request. Verbatim tool result + both preconditions recorded in `.moai/reports/t1203/run/ac014-attempt1.md`; retry path = post-integration rebuilt binary after quota reset |
| AC-CPS-015 | PASS | checks 1–4 (M1) | check 1 `1`/exit 0; D = `802ac54d2c17a6dc9afd07dfa44832389cdba944`; R = `6a89f3d8304ab37dc57a84b9ca17fe81062499c8`; `git merge-base --is-ancestor D R` exit 0; D ≠ R. Authorship remains reviewable, not mechanically proven |
| AC-CPS-016 (b, live) | **NOT MEASURED — pending its own operator authorization** | — | the in-tree half (AC-CPS-004) holds; the live native observation this criterion requires is a separate call from the authorized AC-CPS-014 adversarial call. External quota blocks until 2026-09-28 14:37 AND the acceptance.md Authorization clause requires its own operator grant; no live call was made |

**RED evidence (captured before each GREEN).** (c): build failure
(`undefined: flagVerdictFindingsContradiction`), then with a no-op stub
`--- FAIL` on the V8, predicate and fixture tests (`red-c.log`). (a): six
`--- FAIL`s, e.g. `S1.txt on turn/start: got inconclusive/0, want fail/3`
(`red-a.log`). (d): `prompt does not carry "Verdict: <pass|fail|inconclusive>"`
(`red-d.log`). M3 guards describe pre-existing behaviour and are shown live by
mutation instead (`guard-mutants.log`); recognizer narrowness was shown the same
way — relaxing the comma adjacency or the 판정 statement form fails N2 and the
negative cases, and admitting a hyphen after the bold severity word fails the
`**High-level**` negative (`recognizer-mutants.log`). M4: stage 1 build failure
(`undefined: codexNativeReviewFormatPin`), stage 2 — with only the pin constant
present — the carrier and downgrade assertion failures
(`native thread/start params carry no developerInstructions: map[]`;
`no-signal native body "I walked the diff and moved on.": Verdict = "pass",
want "inconclusive"` ×3) while the constants-sharing, recognition and
pinned-pass controls passed (`m4-red.log`).

**M4 pre-flight and feasibility (tree `4bea753cb`).** `go build ./...` exit 0;
`GOOS=windows GOARCH=amd64 go build ./...` exit 0;
`golangci-lint run --timeout=2m ./internal/cli/...` → `0 issues.`; the AC-CPS-004
RED-now command re-executed — `N1.txt review/start verdict=pass findings=0` and
the same for N2, turn/start rows `inconclusive` (`m4-rednow.log`), identical to
the amended cell's `0ff644530` output. **Carrier established, blocker path
cleared:** the review/start request itself has no instruction field
(`ReviewStartParams` = {delivery: ReviewDelivery|null (inline/detached — not an
instruction carrier), target, threadId}, measured on codex-cli 0.157.0
`codex app-server generate-json-schema`, `/tmp/codex-schema-m4/`), and the
target's `custom` variant is the forbidden substitution — so the pin rides the
session-level destination the review path already uses for the model
(REQ-CX2-002 precedent): `ThreadStartParams.developerInstructions` (measured
`string|null`, same schema run), populated by `codexReviewSessionParams` when
the method is `review/start` and forwarded by `openCodexSessionResolved`. The
review target and its variant are untouched; whether live codex honours a
thread-level pin is AC-CPS-016's live burden and is established by nothing
here.

**M4 stale-witness updates (within AC-CPS-004's scope).** Thirteen tests plus
four scripted-body tests asserted the pre-(b) native default this milestone
removes (unrecognized native body → `pass`), measured via the affected-family
selector before update (`m4-blast.log`, 13 `--- FAIL`) and after (`m4-blast2.log`,
`ok`): the fixture table's N1/N2 review/start rows (plan.md M4 step 4, before/
after rows recorded), AC-CPS-008's guard bodies stating the pin with the
assertion unchanged, AC-CVS-003's native arm and the mode-split witness
(renamed with the supersession recorded in their doc comments), the t551
AC-CBR-008 bound (deliberately lifted by v0.2.4; the pre-M4 value it had pinned
remains cited at `.moai/reports/t551/probe-synthesizer-20260908.txt`), the
characterization row, `realCleanReview` (now the pinned-pass form a real clean
review states under the pin), and the scripted native bodies of the audit-gate,
session-reuse and dispatch tests. `codex_findings_parse_test.go` is untouched
(AC-CPS-007: diff 0).

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

**Quality (M4, tree `861e08912`).** `go build ./...` exit 0 (`darwin_ok`);
`GOOS=windows GOARCH=amd64 go build ./...` exit 0 (`windows_ok`);
`golangci-lint run --timeout=2m ./internal/cli/...` → `0 issues.` (baseline 0;
one transient ST1018 from a literal U+200B written during test editing was
repaired to the escape form before commit). Parser/audit regression selector
`-run 'Codex|Synthesize|Review|Verdict|Convergence|Audit'` → `ok`, exit 0
(`m4-regress2.log`; the first pass surfaced the four scripted-body stale
witnesses named above, all fixed, none semantic). Selector-scoped coverage with
the M4 tests added
(`-run 'TestCodex1718|TestCodexWidened|Contradiction|TestCodexAdversarial|TestCodexNative|TestGuard_|TestSynthesizeReviewOutput|Verdict'`):
`codexReviewSessionParams` 100.0%, `codexUnrecognizedVerdict` 100.0%,
`synthesizeReviewOutput` 100.0%; `openCodexSessionResolved` 71.4% and
`runCodexReviewRPCResolved` 55.6% — the M4-touched lines in both (the
`developerInstructions` forwarding, the session-params injection) are on the
covered success path asserted by `TestCodexNativeRequest_CarriesFormatPin`; the
uncovered arms are pre-existing error paths this milestone does not touch
(`m4-coverage.log` + cover profile). E4 subagent boundary, scoped as in M2/M3
(`grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/mcp_codex*.go
internal/cli/codex_*.go | grep -v '_test.go' | grep -v '//'`): 0. Package-wide
raw grep shows 18 pre-existing doc/help-text mentions in files M4 does not
touch (`harness.go`, `agentlint`, …) — baseline, unchanged.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-26
run_commit_sha: 861e08912
run_status: partial — (a) (b) (c) (d) all implemented; AC-CPS-014 live observation authorized (§E.1) and attempted but blocked (account quota until 2026-09-28 14:37 + running server binary predates the (d) pin — .moai/reports/t1203/run/ac014-attempt1.md); AC-CPS-016 pending its own operator authorization
ac_pass_count: 11   # 004 005 006 007 008 009 010 011 012 013 (remeasured on HEAD per v0.2.3 re-anchoring) 015
ac_fail_count: 0
ac_blocked: []
ac_not_measured: [AC-CPS-014, AC-CPS-016]
ac_wording_change_needed: []
preserve_list_post_run_count: n/a (no PRESERVE list declared)
l44_pre_commit_fetch: not run (lanes do not push; lead batch-pushes develop)
l44_post_push_fetch: not applicable (no push)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: exit 0
  windows_amd64: exit 0
total_run_phase_files: 27
m1_to_mN_commit_strategy: one commit per milestone step, no push, no amend
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-26
sync_commit_sha: 2012431e2   # backfilled per the D3 placeholder exemption — the sync commit carrying the 3-phase close (short SHA, same form as §E.3 run_commit_sha; branch WT-codex-parser-shape)
sync_status: complete — CHANGELOG entry emitted under [Unreleased] ### Added (newest-first); no README / docs-site surface states the codex review verdict-synthesis semantics or the removed native-default-pass behavior (sweep: 4 README locales + docs-site codex surfaces — 0 rows), so CHANGELOG is the only doc change; spec.md frontmatter status in-progress → completed on the sync commit (updated: already 2026-09-26, value unchanged)
b12_self_test_a: pass — grep -c 'SPEC-CODEX-PARSER-SHAPE-001' CHANGELOG.md → 0 before emission (no duplicate entry)
b12_self_test_b: pass — acceptance.md live AC identifiers = 16 (AC-CPS-001..016, zero reserved-token exclusions); CHANGELOG entry states 16 criteria, 11 measured PASS, AC-CPS-014 / AC-CPS-016 explicitly NOT measured (deferred, regression-guard class) — no live observation claimed as passed
b12_self_test_c: pass — every path named in the CHANGELOG entry verified present: internal/cli/mcp_codex.go, internal/cli/testdata/codex-1718/ (N1, N2, S1, S2, S2p), .moai/reports/t1203/run/ac014-attempt1.md (read, cited as evidence)
changelog_entry_position: CHANGELOG.md [Unreleased] ### Added — first bullet (newest-first convention)
frontmatter_status_transitions:
  spec_md: in-progress → completed (the single sync commit; the implemented intermediate is merged per the 3-phase close)
  plan_md: n/a (status-stateless artifact)
  acceptance_md: n/a (status-stateless artifact)
  progress_md: n/a (body sections carry no status field)
canary_compliance_check:
  mx_tag_validation: pass — sync sub-step: all M2–M4 additions in internal/cli/mcp_codex.go are unexported single-caller functions (codexReviewSessionParams, codexUnrecognizedVerdict, flagVerdictFindingsContradiction) plus one struct field (ReviewOutput.Contradiction); no new exported symbol, no fan_in >= 3 symbol → no @MX additions required; existing tags untouched
  doc_surfaces_swept: README.md, README.ko.md, README.ja.md, README.zh.md, docs-site/content/{en,ko,ja,zh}/advanced/{codex-dual-harness,multi-model-audit}.md — none documents codex review verdict synthesis; the multi-model-audit inconclusive prose describes backend-gate convergence, a different layer, and remains accurate
```

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
