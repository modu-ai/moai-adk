# SPEC-CI-CODEX-REVIEW-001 — progress

status: in-progress
card: t1868 (lane-22, run tn1p4e) · tree .moai/worktrees/t1868 · branch WT-ci-codex-review · card base 4f9dba14bf · revision 5 (post-ceiling correction) follows audited revision 4 at 51993bd7d9
evidence: this file + .moai/reports/t1868/ (card record: progress.md; verdict.md, plan-audit*.md, sync-audit*.md are auditor/leader-owned)

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_revision: 6
revised_from: board d-20261018T152011Z-8447 (supersedes d-20261018T151824Z-db98; ruling and scope unchanged) and d-20261018T155527Z-4e9e (in-place amendment of the closed SPEC; revision 5 was audited as rcpt-44d9b0e5a5829de2c78400bd)
plan_complete_at: 2026-10-18
tier: L (19 requirements, 25 criteria at revision 6; ceilings 25 and 25)
artifacts: spec.md · plan.md · acceptance.md · design.md · research.md · decision-index.md · progress.md
notes: Plan revision 6 (amendment, spec.md `## Amendments`, 0.5.0 to 0.6.0): a failure in the `files` or `review` step ends that step and the job non-zero; a draft keeps the single state `not-run:draft` and never reads as success; one value per output key; tests assert step exit status and the derived job result; remote job conclusion UNOBSERVED (G6 acceptance condition); KD-1 transferred to t1874; DS-1 and DS-2 restated. 19 requirements, 25 criteria. Delta: design.md `Delta of plan revision 6`. §E.2 to §E.4 are untouched and describe the closed revision-5 run; M5 (plan.md) is the amendment's run work. The plan-audit verdict is auditor-owned. Plan revision 5 is the post-ceiling correction under the leader's standing rule C (board record
d-20261018T133012Z-5eb0): one correction commit for the three blockers of plan audit iteration 3 (FAIL 0.85;
receipts rcpt-44d9b0e5a5829de2c78400bd and rcpt-8b713fb407597f85c325b2fe), then one delta audit, named
plan-audit-iter4.md, which is auditor-owned and not written here. D24: REQ-CCR-003 excludes by author and no
body text is read. D25: the `review` step reads the head again after its list reads. D26: the `files` term of
the merge `if` is pinned structurally by AC-CCR-012 and AC-CCR-019 names a non-equivalent mutant. Requirement
and criterion identifier sets are unchanged (17 and 19); Q20 and FU-10 are untouched; D27-D36 are optional and
left as they are. Closure table with file:line: `.moai/reports/t1868/progress.md`.
Revision 4 note, kept for lineage: plan revision 4 after plan audit iteration 2 (FAIL 0.81; receipts
rcpt-df0227b9e68727bfeefccd65 and rcpt-196d8f94f5876f4814d6e5de), the last normal round. The design direction is unchanged: the reviewer is the
ChatGPT subscription's native Codex GitHub review and the gate reads a pull-request review by the Codex app
identity (login chatgpt-codex-connector[bot], user id 199175422, type Bot) bound to the exact head commit. What
changed is the review reader: only the newest bound review in state APPROVED is clean (COMMENTED, DISMISSED,
PENDING, and review body text never are), a CHANGES_REQUESTED review is superseded only by a later APPROVED
(D12, D13); NOT RUN is one named result for a draft and for a review absent at the end of the wait window, an
unreadable list is a failure (D17); changed-file entries are bound to the verified head and the head is read
before and after (D14); the merge decision path is a criterion of its own, AC-CCR-019 (D15); the `gh`
allow-list is exact (D16); AL-1 is an unverified expectation (D18); CI-input files outside the tamper guard are
named residual risk and follow-up FU-11 (D19); a repeated list entry refuses (D20); research cross-references
are corrected (D21); REQ-CCR-003, 013, 014, 015, 016, 017 are in positive form (D22); the dated DS-1
observation of 2026-10-18 is recorded in the DS-1 row (D23). The native review's artifact stays UNCONFIRMED:
the gate withholds every automated merge unless an APPROVED review by the identity is bound to the head; the
deferred read-only steps DS-1, DS-2, and DS-3 are leader-coordinated, never deliverables. Closure table of
D12-D23 with file:line: `.moai/reports/t1868/progress.md`. Boundary ruling (b), the follow-up candidates, and
KD-1 are kept; gates G2-G8 stay deferred and no requirement or criterion depends on any gate or step. The
plan-audit verdict is auditor-owned and is not written here.
audit_ready: true

## §E.2 Run-phase Evidence

Run owner: manager-develop (cycle_type=tdd) for lane-22, card t1868, run tn1p4e, Class C, Tier L. Tree
`/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1868`, branch `WT-ci-codex-review`, kickoff HEAD `b94983c323`. Plan
artifacts were not edited (spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md); this file
(§E.2, §E.3 only) and the card record `.moai/reports/t1868/progress.md` are the only SPEC-side edits.
Verbatim evidence (every RED, GREEN, mutant, and landing output): `.moai/reports/t1868/run-evidence.md`
(committed; cited below as `EV:<line range>`; every mutant block starts with a `### <id>` heading).

State: the run phase is complete on the card branch. **Not landed locally** (no merged-local SHA exists; the
lane merges). **Remote verified: not reached by this lane.** No key, secret, comment, review request, app or
settings change, branch-protection call, push, pull request, or any remote write was made; remote access was
read-only `gh api` GETs only: the five of the pre-flight (`EV:20-36`), one of the remote-only
`claude-code-review.yml` content for the M4 mutant (`EV:1891-1960`), and the real-API smoke (`EV:2414-2447`).

Commits in order (explicit pathspec, one pathspec set per commit; HEAD and branch re-read before each):

| # | SHA | Subject |
|---|-----|---------|
| 1 | `3dcb8e6c64` | test(t1868): M1 RED read-native-review reader tests, seven tests, 71 subtests |
| 2 | `0c4f8cb026` | feat(SPEC-CI-CODEX-REVIEW-001): M1 gate script read-native-review reader |
| 3 | `561678e801` | test(t1868): M2 RED fetch-pages and gate-files-changed tests, three tests, 44 subtests |
| 4 | `17d23ad732` | feat(t1868): M2 gate script fetch-pages and gate-files-changed |
| 5 | `d6f0328aed` | test(t1868): M3 RED auto-merge consumer tests, seven tests, one subcommand-aware stub gate (its message says "59 subtests" in error: 59 is the count of `--- FAIL` lines at RED; the file holds 74 subtests) |
| 6 | `38969f7f8b` | feat(t1868): M3 auto-merge wires the changed-file guard and the native review gate, CodeRabbit step removed |
| 7 | `8412e93ffc` | test(t1868): M4 RED retirement and census tests plus the required-checks Codex guard |
| 8 | `23e6c61fb1` | feat(t1868): M4 retire claude.yml and review-quality-gate.yml with their liveness entries, CodeRabbit auto review and reply off |

Every RED commit precedes the implementation commit of its milestone in the commit graph (verification-claim-integrity
section 2.3). Each RED test file set was committed alone and red; it compiled and passed `go vet` (exit 0) before its commit, so the
tree builds at every commit.
The AC-CCR-017 guard (a regression-guard, green on the baseline by nature) shares commit 7 with the M4 RED tests.

Files changed by the run (12, the plan's set; `git diff --name-only b94983c323 23e6c61fb1`): new (7)
`scripts/ci/codex-review-gate.sh`, `internal/template/codex_review_gate_test.go`,
`internal/template/codex_review_paging_test.go`, `internal/template/auto_merge_codex_test.go`,
`internal/template/review_retirement_test.go`, `internal/guardstate/census_review_test.go`,
`internal/config/required_checks_review_guard_test.go`; edited (3) `.github/workflows/auto-merge.yml`,
`.coderabbit.yaml`, `.moai/guards/liveness.yaml`; deleted (2) `.github/workflows/claude.yml`,
`.github/workflows/review-quality-gate.yml`. No non-test Go file, no template file, no `internal/template/templates/`
path changed, so no `make build` was needed (`EV:2038-2109` lists the changed paths).

### Run-phase AC matrix (E1)

Selectors are name-exact. "named PASS" is the count of acceptance-named subtests found as `--- PASS` lines in the
landing re-measure (`EV:2147-2389`); the swept-set check found every acceptance-named subtest present, none missing
(`EV:2390-2410`). Zero `--- SKIP` lines in any run. RED/GREEN/mutant line ranges are milestone blocks of `run-evidence.md`:
M1 RED `EV:58-148`, GREEN `EV:149-232`, mutants `EV:233-780`; M2 RED `EV:783-842`, GREEN `EV:843-895`, mutants
`EV:896-1103`; M3 RED `EV:1113-1184`, GREEN `EV:1185-1288`, mutants `EV:1289-1785`; M4 RED `EV:1788-1808`, GREEN
`EV:1809-1829`, mutants `EV:1830-1960`, guardstate and config `EV:1961-2037`; flips and contract commands `EV:2038-2109`.

| AC | Status | Command | RED (reason, source) | GREEN | Mutants made red |
|----|--------|---------|----------------------|-------|------------------|
| AC-CCR-001 | PASS | `go test -count=1 -v -run '^TestCodexGateReadNativeReviewAccepts$' ./internal/template/` | 11 FAIL lines (parent + 10 rows), message `gate script absent: .../scripts/ci/codex-review-gate.sh` (R1), commit 3dcb8e6c64 | 10 named PASS, commit 0c4f8cb026 | M1-A1 (unconditional exit 0 that prints the clean line: Accepts stays green, AC-002..005 red), A2, A3, A4, A5 |
| AC-CCR-002 | PASS | `... '^TestCodexGateReadNativeReviewRefusesStale$'` | 7 FAIL lines (R1) | 6 named PASS | M1-B1, B2, B3, B4 |
| AC-CCR-003 | PASS | `... '^TestCodexGateReadNativeReviewIgnoresForged$'` | 11 FAIL lines (R1) | 10 named PASS | M1-C1 .. C6 (C5 is the revision-4 literal REQ-CCR-003 reading, red on `identity_approved_body_says_approved_not_excluded`) |
| AC-CCR-004 | PASS | `... '^TestCodexGateReadNativeReviewRefusesIncomplete$'` | 25 FAIL lines (R1) | 24 named PASS | M1-D1 (the revision-3 reader, 8 of 8 named rows red), D2 .. D7 |
| AC-CCR-005 | PASS | `... '^TestCodexGateReadNativeReviewRefusesFinding$'` | 9 FAIL lines (R1) | 8 named PASS | M1-E1 .. E4 |
| AC-CCR-006 | PASS | `... '^(TestAutoMergeReadsNoCommentOrStatusVerdict\|TestCodexGateFetchPagesRefusesDisallowedEndpoint)$'` | M3 RED: 5 of 8 rows red (R2, R3: `forbidden token "issues/"`, `"/status"`, `"statuses"`); M2 RED: 5 of 5 endpoint rows red (`unknown-subcommand:fetch-pages`) | 8 + 5 named PASS | M2-B1; M3-A1, A2, A3 |
| AC-CCR-007 | PASS | `... '^(TestCodexGateReadNativeReviewPending\|TestAutoMergeWaitLoopIsBounded)$'` | script rows red (R1); loop rows 10 of 10 red against the unchanged workflow (R2) | 2 + 10 named PASS | M1-F1; M3-A4 .. A8 |
| AC-CCR-008 | PASS | `... '^(TestCodexGateReadNativeReviewDraftNotRun\|TestAutoMergeNotRunWithholds)$'` | script rows red (R1); step rows 11 of 11 red (R2) | 10 + 11 named PASS | M1-F2, F3, F4; M3-A9, A10; M3-W7a (the exit-11 killer, D39) |
| AC-CCR-009 | PASS | `... '^TestCodexGateFilesChangedRefusals$'` | 22 FAIL lines, `unknown-subcommand:gate-files-changed` (M2 RED) | 21 named PASS | M2-C1 .. C6 (C6 is the "list read by pull request number, no entry binding" mutant) |
| AC-CCR-010 | PASS | `... '^TestCodexGateFetchPagesCompleteness$'` | 19 FAIL lines, `unknown-subcommand:fetch-pages` | 18 named PASS | M2-A1 (first page only, count proof intact), A2 (first page only, no count proof: the audit's counterexample), A3 .. A7 |
| AC-CCR-011 | PASS (regression-guard) | `... '^TestAutoMergeCheckoutAndGuardsUnchanged$'` and `'^TestAutoMergeRequiredChecks$'` | none by nature (green on the baseline: 6 PASS at M3 RED) | 5 named PASS; `TestAutoMergeRequiredChecks` 17 `--- PASS` lines (parent + 16 subtests) | M3-R1 (checkout ref), M3-R2 (`--match-head-commit` removed), both recorded red verbatim |
| AC-CCR-012 | PASS | `... '^TestAutoMergeWiresGuardAndGate$'` plus `grep -c` of the three subcommand names | R4, R5, R6 flipped: counts 0 at the plan pin, now 1, 2, 3 (exit 0); 23 of 24 rows red at M3 RED (only `permissions_unchanged` green) | 24 named PASS | every mutant of the criterion: M3-W1, W2, W3, W4, W5, W5b, W6, W7a, W7b, W8, W9 (structural), W10, W11a, W11b, W11c |
| AC-CCR-013 | PASS (regression-guard) | `... '^TestAutoMergeAddsNoSecretTriggerOrArtifact$'`; `git diff --quiet --diff-filter=A <A> HEAD -- .github/workflows` | none by nature | 4 named PASS; command exit 0; control lists the two deleted workflows (`EV:2067-2109`) | M3-R3, R4, R5 recorded red; the command form's observed failure on a tree known to add a workflow: exit 1 (`EV:2067-2109`) |
| AC-CCR-014 | PASS | `... '^TestNoClaudeReviewWorkflows$'` | 4 of 7 rows red (R7, R10: `claude.yml runs anthropics/claude-code-action`, `review-quality-gate.yml triggers on check_run`, both files exist) | 7 named PASS (19 workflow files on disk) | M4-S1 (claude.yml restored), M4-S2 (the remote-only claude-code-review.yml merged back in, content read by GET), M4-S4 |
| AC-CCR-015 | PASS | `... '^TestCodeRabbitDisabledAndUnread$'` | 2 of 6 rows red (R8: `reviews.auto_review.enabled = true`, `chat.auto_reply = true`) | 6 named PASS | M4-C1, C2, C3, S3 (file deleted instead of disabled) |
| AC-CCR-016 | PASS (regression-guard) | `git merge-base main HEAD` -> `4f9dba14bf2834aa08bea79f9e72c2a7e75ab18e`; `git diff --quiet <A> HEAD -- <the seven paths>` | none by nature; observed red R9: `git diff --quiet 4f9dba14bf 6dde1c1c1b -- .github/required-checks.yml` exit 1 | exit 0; control `git diff --name-only <A> HEAD` lists 25 paths (`EV:2067-2109`) | R9 (the probe fails on a tree known to differ) |
| AC-CCR-017 | PASS (regression-guard) | `go test -count=1 -v -run '^TestRequiredChecksCarryNoCodexReviewCheckUntilConfirmed$' ./internal/config/` | none by nature (6 PASS at baseline) | 6 named PASS | `Codex Review` added to main contexts, then to auxiliary, on COPIES of the contract (CCR_GUARD_ROOT seam; the real file untouched), both red naming gate G4 (`EV:1996-2037`) |
| AC-CCR-018 | PASS | `... '^TestCensusOmitsRetiredReviewWorkflows$' ./internal/guardstate/` plus the three `TestCensus_*` | 3 FAIL lines (R10, R11: `the manifest still lists the retired workflow`) | 2 named PASS + 3 census PASS | M4-L1, L2 (entries left in place), M4-S1c (file restored: `disk\manifest`) |
| AC-CCR-019 | PASS | `... '^TestAutoMergeMergeDecisionPath$'` | R12 flipped (0 -> 1, exit 0 now); 5 of 12 rows red at M3 RED (the 7 green rows are vacuous on the unchanged workflow, which merges nothing; the paired positive control `clean_fixture_merges_once_with_head_pin` is red) | 13 named PASS | M3-D1, D2, D3, D4, D5 (D38 form), W8; D5b and the W9 control are not mutants (see D38, D26 below) |

Counts at the landing re-measure: internal/template 239 `--- PASS`, 0 `--- FAIL`, 0 `--- SKIP` (7 + 3 + 7 + 2 + 1 + 1 selectors
named); internal/guardstate 6 PASS; internal/config 9 PASS (`EV:2110-2146`).

### Carried debts disposed in this run (plan-audit-iter5, dispose_in=run)

- **D38.** The audit's mutant "pinning `--match-head-commit` to another variable" is recorded as "pinning it to a constant
  other than the verified head SHA": M3-D5 replaces `"$TESTED_SHA"` with a 40-zero constant and turns
  `MergeDecisionPath/clean_fixture_merges_once_with_head_pin` red (`EV:1635-1645`). The variable variant
  `"$HEAD_NOW"` was run as M3-D5b and stays green (`EV:1646-1653`): `$HEAD_NOW` equals `$TESTED_SHA` at the call
  (the merge step returns before the call otherwise), so it is behaviourally identical and not a mutant. The form is
  also written in the mutant comment above `TestAutoMergeMergeDecisionPath` in `auto_merge_codex_test.go`.
  `acceptance.md` was not edited.
- **D39.** The exit-11 half of the AC-CCR-012 mutant "treating exit 11 or an unknown exit as pass" is killed by the
  AC-CCR-008 row `TestAutoMergeNotRunWithholds/exit_11_withholds_merge` (M3-W7a: `exit_11_withholds_merge` and
  `exit_11_never_reaches_merge_condition` red, `EV` block `### M3-W7a`); the unknown-exit half is killed by
  `TestAutoMergeWiresGuardAndGate/review_step_unknown_exit_not_ok` (M3-W7b, which also reddens
  `MergeDecisionPath/review_unknown_exit_zero_merge_calls`). The row carries a comment naming D39 in the test file.
- **D40.** The Definition of Done is the binding mutant list: every mutant named by AC-CCR-012 and AC-CCR-019 has a
  recorded red (M3-W1 .. W11c, W5b and M3-D1 .. D5, W8); the `files` term removal is the structural mutant of
  AC-CCR-012 (M3-W9: only `merge_if_requires_files_and_review_ok` red) and the control run shows
  `gate_path_change_zero_merge_calls` stays green under it (`EV:1772-1785`), the combined `files`-dependency mutant
  M3-D4 being the one that reddens the gate-path row. The shorter `plan.md` enumeration was not followed.
- **D41.** One subcommand-aware stub gate script (`cgGateStub` in `auto_merge_codex_test.go`) dispatches by subcommand
  (`fetch-pages`, `gate-files-changed`, `read-native-review`), takes its exit codes from `STUB_FETCH_EXITS`,
  `STUB_FILES_EXITS`, `STUB_REVIEW_EXITS`, and appends one line per invocation to the same ordered `CALL_LOG` that the stub `gh` and the stub
  `sleep` write to, so "after the second head read" is decidable by position; the dispatch table is the comment at the
  top of the file. `review_step_head_moved_during_read_not_ok` asserts two pull-object reads, both `fetch-pages`
  calls between them, and no `read-native-review` call at all; mutant M3-W5b (the second read made but checked after
  the verdict call) is red on that clause.
- **Q20 / FU-10** (dispose_in=sync) were not touched.

### Other results

- `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go vet ./internal/template/
  ./internal/config/ ./internal/guardstate/` exit 0; `gofmt -l` on the six new test files prints nothing;
  `/Users/goos/go/bin/actionlint .github/workflows/auto-merge.yml` exit 0 (v1.7.10) (`EV:2110-2116`).
- Tool provenance: tests ran under macOS `/bin/bash` 3.2.57(1)-release and jq-1.8.1 (the only bash on PATH). The
  installed moai build was not used by any measurement; no moai verb was run.
- Real-API smoke (read-only GETs, `EV:2414-2447`): the committed script, through the real `gh`, collected the 1029
  changed files of PR 1826 over 11 pages with the count proof agreeing (`fetched:1029 count-proof:yes`), and
  `gate-files-changed` refused it with `gate-path-changed:.github/workflows/ci.yml` (that pull request does change a
  workflow); on PR 1823 (a CodeRabbit review only, no review of the Codex identity) `read-native-review` exited 10
  `pending:no-review` with the window open and 11 `not-run:review-absent` with it closed; `fetch-pages` on an
  issue-comment endpoint exited 1 `endpoint-not-allowed`. This is not evidence about the native Codex review, whose
  artifact has not been observed (DS-1).
- Checkout carries no `ref` and the merge call keeps `--match-head-commit "$TESTED_SHA"` (AC-CCR-011 rows PASS).
- E3 coverage and E4 subagent-boundary grep are not applicable: no non-test Go file changed and no instruction file
  was added; the new code is bash, YAML, and `_test.go` files.
- Doctrine, templates, `AGENTS.md`, role-core rules, the ledger, hook tests, the lane notice, and `required-checks.yml`
  were not touched; known debt KD-1 and accepted limitations AL-1, AL-2 stand as written in `spec.md` 7.

### Gaps (not observed in this run; the full list is also in the card record)

- No full-package run of `internal/template`, `internal/cli`, `internal/hook`, `internal/guardstate`, or `internal/config`
  (forbidden by the lane's slot discipline). The landing re-measure that needs a slot: the whole `internal/template`
  package once (`go test -count=1 ./internal/template/`), under the heavy-run slot.
- The criteria were evaluated on the card branch, not on the merged-local tree; Definition of Done State L item 1 and
  the landing-time AC-CCR-016 command on the absorbed `main` (left end `A` changes after the absorb-`main` merge) are the lane's to run.
- The native review's artifact form stays unobserved (DS-1); every fixture is synthetic in the documented REST shape.
  Under the shipped contract every automated merge stays withheld until a real APPROVED review by the identity is observed.
- Not run: golangci-lint, `moai spec lint`, the `internal/cli` live-axis test, any test under another jq version
  (jq 1.7 on the ubuntu runner) or bash 5; `make build` (nothing templated changed).
- `spec.md` `status: draft` was not advanced to `in-progress`: the run instruction freezes spec.md (the plan-artifact
  hash binds it); the transition is left for the leader or manager-spec.
- Plan section F names milestones M1..M4; the instruction text mentions "plan M5". The deletions and manifest removals were done in
  M4, in one commit (plan section B.1).
- 2026-10-23 trusted identity check: `gh api users/chatgpt-codex-connector[bot]` and the reverse lookup `user/199175422` both returned login chatgpt-codex-connector[bot], id 199175422, type Bot (observed by the lane and by the leader).

### Sync-phase correction (ruling B)

Authority: leader board d-20261018T151713Z-ad89 (option B): the required codex leg's FAIL on the sync audit
(`.moai/reports/t1868/sync-audit.md`, F1 codex P1) is fixed in this card, not disposed as the auditor's own
PASS-WITH-DEBT; a delta sync audit at the new HEAD follows. No REQ or AC identifier was added; only
`design.md` (one sentence in section 5, residual item 7 in section 6) was edited among the plan artifacts.

| Finding | Fix | Commit |
|---|---|---|
| F1 (codex P1, `auto-merge.yml:304`) the second PR read compared only `head.sha`; a review comment added or an approval dismissed after the list reads, head unchanged, still reached a merge call | the second read also compares `review_comments` with the first collection's count (`review-comments-moved-during-read`); a clean pass is rechecked before `ok=true`: read the PR, re-collect the reviews and the comments (`--expect-count "$RC_COUNT"`), read the PR again (head and count must match), run `read-native-review` on the fresh lists, require exit 0 and the identical summary line (same selected review); any change ends the step `ok=false`, `state=fail`. Exit mapping 0/1/10/11, fail-closed rule, `--match-head-commit` pin, and the gate-script subcommands are unchanged; no subcommand added | `205cc4c8fc` |
| D-2 comment at `auto-merge.yml:256` | now says the window is bounded by the 17-minute deadline and that a long `checks` wait or the 20-minute job timeout can shorten it, in which case the merge is withheld | `205cc4c8fc` |
| D-3 CHANGELOG | the COMMENTED sentence names `coderabbitai[bot]` (and one GitHub Advanced Security review), not the Codex app; one sentence added that the review step rechecks review comments and the selected approval review before the merge; nothing else changed | `205cc4c8fc` |

Commits: `7278fd4c75` test(t1868): ruling B RED, pre-merge review recheck rows and a stub gh with per-collection pages
(committed alone, before the fix); `205cc4c8fc` fix(t1868): ruling B, the fix, D-2, D-3, design sentences.
Tests: new top-level `TestAutoMergeReviewRecheckBeforeMerge` in `internal/template/auto_merge_codex_test.go` (7
subtests, real gate script, real step, whole chain; the stub `gh` in `codex_review_paging_test.go` now serves the pages of
generation N of a list when `$FIXTURES/<kind>/gN/page-P.json` exists, counted per page-1 read, so a second collection can
differ); two existing count assertions moved by the one extra recheck call (`pending_keeps_waiting` 4 to 5,
`fetch_failure_counts_pending` 1 to 2).

RED (committed pre-fix workflow, 0:33, load 27.03 at start):
```
$ go test -count=1 -v -p 4 -run '^(TestAutoMergeReviewRecheckBeforeMerge|TestAutoMergeWaitLoopIsBounded)$' ./internal/template/   exit=1
    --- FAIL: TestAutoMergeReviewRecheckBeforeMerge/later_changes_requested_on_recheck_not_ok (2.04s)
    --- FAIL: TestAutoMergeReviewRecheckBeforeMerge/second_read_review_comments_count_differs_not_ok (2.07s)
    --- FAIL: TestAutoMergeReviewRecheckBeforeMerge/approval_dismissed_on_recheck_not_ok (2.11s)
    --- FAIL: TestAutoMergeReviewRecheckBeforeMerge/new_finding_comment_on_recheck_not_ok (2.13s)
    --- FAIL: TestAutoMergeReviewRecheckBeforeMerge/selected_approval_replaced_on_recheck_not_ok (2.14s)
    --- PASS: TestAutoMergeReviewRecheckBeforeMerge/unchanged_lists_merge_once_after_recheck (2.45s)
    --- FAIL: TestAutoMergeReviewRecheckBeforeMerge/recheck_review_comments_count_differs_not_ok (2.45s)
    --- FAIL: TestAutoMergeWaitLoopIsBounded/fetch_failure_counts_pending (1.35s)
    --- FAIL: TestAutoMergeWaitLoopIsBounded/pending_keeps_waiting (2.15s)
FAIL	github.com/modu-ai/moai-adk/internal/template	32.844s
  auto_merge_codex_test.go:1497: ok="true" state="pass", want not-true/fail      (each new failing-path row: the unfixed step passes)
  auto_merge_codex_test.go:713: reads=1 sleeps=2, want 2 (1 pass + 1 pre-merge recheck) and 2
  auto_merge_codex_test.go:655: read-native-review ran 4 times, want 5 (4 passes + 1 pre-merge recheck)
```
GREEN (fix applied; the 12 auto-merge and paging selectors; load 38.32 at its start):
```
$ go test -count=1 -v -p 4 -run '^(TestAutoMergeReviewRecheckBeforeMerge|TestAutoMergeWaitLoopIsBounded|TestAutoMergeWiresGuardAndGate|TestAutoMergeNotRunWithholds|TestAutoMergeMergeDecisionPath|TestAutoMergeReadsNoCommentOrStatusVerdict|TestAutoMergeCheckoutAndGuardsUnchanged|TestAutoMergeAddsNoSecretTriggerOrArtifact|TestAutoMergeRequiredChecks|TestCodexGateFetchPagesCompleteness|TestCodexGateFetchPagesRefusesDisallowedEndpoint|TestCodexGateFilesChangedRefusals)$' ./internal/template/
exit=0; --- PASS 153, --- FAIL 0, --- SKIP 0
ok  	github.com/modu-ai/moai-adk/internal/template	65.004s
    --- PASS: TestAutoMergeReviewRecheckBeforeMerge/{second_read_review_comments_count_differs_not_ok, recheck_review_comments_count_differs_not_ok, approval_dismissed_on_recheck_not_ok, later_changes_requested_on_recheck_not_ok, new_finding_comment_on_recheck_not_ok, selected_approval_replaced_on_recheck_not_ok, unchanged_lists_merge_once_after_recheck}   (7 of 7)
    --- PASS: TestAutoMergeWaitLoopIsBounded/pending_keeps_waiting, /fetch_failure_counts_pending
$ go vet -p 4 ./internal/template/        go vet exit=0
$ actionlint .github/workflows/auto-merge.yml        actionlint exit=0
```
Mutants (each applied to the committed fix by `mut.py`, selector `^TestAutoMergeReviewRecheckBeforeMerge$`, one Go call at a
time, `git checkout -- .github/workflows/auto-merge.yml` after each, `git status --short` empty before the next; load 28.93
at the first call):
```
RB-1 no recheck at all (the pre-correction step)        6 of 6 failing-path rows red (the step merges once in each):
     recheck_review_comments_count_differs_not_ok, later_changes_requested_on_recheck_not_ok, approval_dismissed_on_recheck_not_ok,
     selected_approval_replaced_on_recheck_not_ok, new_finding_comment_on_recheck_not_ok, second_read_review_comments_count_differs_not_ok
RB-2 only the second-read count compare removed         red: second_read_review_comments_count_differs_not_ok (reason assertion: the
     pre-merge block still withholds, zero merges; the row names review-comments-moved-during-read)
RB-3 only the recheck call removed                      5 of 5 rows red: approval_dismissed_..., selected_approval_replaced_...,
     recheck_review_comments_count_differs_..., later_changes_requested_..., new_finding_comment_...   (the second-read count row stays green)
RB-4 recheck without re-collecting the lists            4 of 4 rows red: approval_dismissed_..., later_changes_requested_...,
     new_finding_comment_..., selected_approval_replaced_...
RB-5 recheck without the same-selected-review compare   red: selected_approval_replaced_on_recheck_not_ok
RB-6 both recheck count compares removed                red: recheck_review_comments_count_differs_not_ok
each: "path clean after restore: True (status='', before='')"
```
Every required row (count differs on the second read, approval DISMISSED, a later CHANGES_REQUESTED, a new finding comment)
has a mutant that merges: RB-1 for all, RB-3 and RB-4 for the list-borne ones. Mutant evidence is in this record; the raw
runner output was in scratch and is not committed.

Gaps of this round: (1) the GREEN run started at load 38.32, above the 32 admission limit (it ended with no wall-clock failure;
recorded, not read as a defect); (2) `go vet` was started right after GREEN without a load re-check (load 50.14 at the end of
GREEN, 47.49 after vet; exit 0); (3) `go build ./...` was not run in this round: load read 33.19 after the mutants, above the
limit, so it stays unrun (the changed files are test sources and a workflow, and `go vet` of the template package compiled
them); (4) the earlier M3 mutant spec for `MAX_WAIT` (A4) matches text this round changed and was not rerun; (5) the new rows
run the real gate script and a stub `gh`; the workflow has still never run on a runner; (6) the whole-package
`internal/template` run, the guardstate and config selectors, and the delta sync audit are not part of this round and are
the lane's; (7) the few seconds between the recheck and the merge call remain (design section 6, item 7).

### Contract-fix implementation (board 8447/4e9e)

Authority: leader boards 8447/4e9e (plan audit iter9 PASS 0.97 @`728f0a6e6c`). The files and review steps of
`.github/workflows/auto-merge.yml` now implement the audited contract (design.md section 5, T-CHAIN and T-OUTPUTS): a
failure in either step ends the step non-zero and fails the job; a non-draft pull request with no review at the end of the
window is a failure (`not-run:review-absent`, unreadable lists `review-unreadable`); a draft ends the review step with exit 11
and the single state `not-run:draft` (the job fails, never success); each step records `ok` and `state` once through one
`conclude` function; the merge runs only after files and review are ok. The gate script is unchanged (exit class 11 for both
NOT RUN causes). No SPEC artifact was edited in this round.

Commits: `f5d71f31e5` test(t1868): M5 RED (tests alone, before the fix); `a1b9242d6f` fix(t1868): M5 the fix plus the CHANGELOG
sentence (plan section B item 16).

**Admission label correction (board 8c1e, leader message of 2026-10-19).** The RED run (`f5d71f31e5`, started 16:42:46Z) and
the GREEN run (`a1b9242d6f`, started 16:43:55Z) both ran while landing slot t1818 was held (16:29:35Z to 16:59:35Z) and real
template Go was running (16:42:54Z). Load and memory alone were read before each of them (RED load 21.17, GREEN load 30.62,
mem free 83%); the slot state was not read. They are functional records only, NOT admission-compliant evidence.

RED (committed pre-fix workflow, tests only; functional record): `exit=1`, 69 `--- FAIL` and 50 `--- PASS` lines. Representative
failing rows, verbatim:
```
auto_merge_codex_test.go:1841: review step: step exit status 0, want 1      (9 rows)
auto_merge_codex_test.go:2278: files step: step exit status 0, want 1       (8 rows)
review_clean: the review step output is [ok=false state=fail ok=true state=pass], want exactly one ok line and one state line
review_draft: ... [ok=false state=fail state=not-run]
state="not-run", want not-run:draft
job="success" merge calls=[], want failure and none
```
GREEN (fix applied; functional record): `exit=0`, 211 `--- PASS`, 0 FAIL, 0 SKIP,
`ok  github.com/modu-ai/moai-adk/internal/template 52.210s`; subtests of the six new tests 13/6/8/7/6/12, `MergeDecisionPath` 14,
`NotRunWithholds` 9. `actionlint .github/workflows/auto-merge.yml` exit 0 before the GREEN commit.

Observer evidence (not lane measurement): the leader's consumer replay
`/tmp/adk-ci-followthrough-20261018/consumer-replay-164554/result.json` (observed 2026-10-18T16:45:55Z at source HEAD
`a1b9242d6f`, workflow sha256 `91f3a6ed...`, gate sha256 `2d07cdc4...`) ran the committed step bodies with the real gate script
and a local fixture `gh`, under `bash -eo pipefail`: changes-requested exit 1, outputs `ok=false` `state=fail` once each;
draft exit 11, outputs `ok=false` `state=not-run:draft` once each; clean-positive exit 0, `ok=true` `state=pass`; files with a
missing gate script exit 1 `ok=false` `state=fail`; review with a missing gate script exit 1 `ok=false` `state=fail`. Its own gap
line: the actual GitHub step and job conclusion and the native review delivery are not observed. DS-1 stays unobserved.

**M5 mutants (window of 2026-10-19, slot go-test-heavy held, lane-22 holder per the leader).** Each mutant: edit applied to the
committed file by `t1868-window/mut2.py`, ONE `go test -count=1 -json -p 4 -run <selector> ./internal/template/` call, file
restored with `git checkout --`, `git status --short` empty after each (all twelve; first line of `moai slot status` read as
`slot go-test-heavy: held` before every call, the holder name is not on that line). Selector for every mutant: the six new tests
(`ReviewFailureFailsStepAndJob`, `FilesStepFailureFailsStepAndJob`, `DraftIsNotRunNeverSuccess`, `ReviewAbsentIsFailure`,
`GuardStepsOutputOneValuePerKey`, `StepHarnessObservesExitAndJobResult`) plus `NotRunWithholds`, `WiresGuardAndGate`,
`MergeDecisionPath`, name-exact (`^(...)$`). Failure names were read from the JSONL by the runner (Action `fail` with a Test).
Raw JSONL: `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/207b7b17-aa81-4311-8e47-c01448800dee/scratchpad/t1868-window/m5-mutant-NN.jsonl`
(machine-local scratch, not a citation target). Counts are `Action` pass/fail events carrying a Test name (parent tests
included); every call exited 1 and ended `FAIL github.com/modu-ai/moai-adk/internal/template ~25s`.

| # | Mutant | reading before the call (UTC, load, mem free) | pass/fail | rows that went red (key rows) |
|---|---|---|---|---|
| 0 | MF: files term dropped from the review `if` and the merge `if`, both led by `always()` | 17:31:01Z, 21.77, 83% | 87/21 | `MergeDecisionPath/gate_path_change_zero_merge_calls` (`gh pr merge was called 1 times, want 0: [gh pr merge 7 --squash --match-head-commit 0123456789abcdef0123456789abcdef01234567 --delete-branch]`; original: 0 merge calls), `WiresGuardAndGate/review_step_depends_on_files_ok`, `WiresGuardAndGate/merge_if_requires_files_and_review_ok`, `StepHarnessObservesExitAndJobResult/guard_steps_have_no_continue_on_error_or_status_override`, `MergeDecisionPath/page_failure_zero_merge_calls`, `MergeDecisionPath/if_expression_outside_grammar_fails_test`, `MergeDecisionPath/refusal_fixtures_job_result_failure_clean_fixture_success`, and 10 `FilesStepFailureFailsStepAndJob` rows |
| 1 | EQ: `steps.review.outputs.ok == 'true'` dropped from the merge `if` | 17:31:33Z, 28.91, 83% | 104/4 | `WiresGuardAndGate/merge_if_requires_files_and_review_ok` and `NotRunWithholds/exit_11_withholds_merge`; no `MergeDecisionPath` merge-call row red (observed, not forced; the structural `WiresGuardAndGate` row and the existing `NotRunWithholds/exit_11_withholds_merge` row are the ones that see this mutant) |
| 2 | X0: `conclude` ends exit 0 on both steps (the committed pre-correction behaviour) | 17:32:14Z, 29.75, 83% (not admitted at 17:32:03Z 32.99 and 17:32:06Z 32.35: no call) | 64/44 | `ReviewFailureFailsStepAndJob/changes_requested_step_exit_nonzero_job_failure` (and 10 more rows of that test), `FilesStepFailureFailsStepAndJob/gate_path_changed_step_exit_nonzero_job_failure` (and 8 more), `ReviewAbsentIsFailure/review_absent_window_closed_step_exit_1_job_failure`, `.../lists_unreadable_at_expiry_step_exit_1_review_unreadable`, `DraftIsNotRunNeverSuccess/draft_step_exit_11_job_not_success`, `.../draft_with_clean_approved_review_still_not_success`, `NotRunWithholds/exit_11_withholds_merge`, 10 `WiresGuardAndGate` rows, 2 `MergeDecisionPath` rows |
| 3 | COE-review: `continue-on-error: true` on the review step | 17:32:53Z, 31.96, 83% (17:32:43Z 34.57 not admitted) | 87/21 | `ReviewFailureFailsStepAndJob/review_step_has_no_continue_on_error`, `.../changes_requested_step_exit_nonzero_job_failure` and 9 more, `ReviewAbsentIsFailure/review_absent_window_closed_step_exit_1_job_failure`, `DraftIsNotRunNeverSuccess/draft_step_exit_11_job_not_success`, `StepHarnessObservesExitAndJobResult/guard_steps_have_no_continue_on_error_or_status_override`, `MergeDecisionPath/refusal_fixtures_job_result_failure_clean_fixture_success` |
| 4 | COE-files: `continue-on-error: true` on the files step | 17:33:35Z, 30.28, 83% (17:33:22Z 32.99, 17:33:26Z 32.35 not admitted) | 92/16 | `FilesStepFailureFailsStepAndJob/files_step_has_no_continue_on_error`, `.../gate_path_changed_step_exit_nonzero_job_failure` and 8 more, `MergeDecisionPath/gate_script_missing_zero_merge_calls`, `.../refusal_fixtures_job_result_failure_clean_fixture_success`, `StepHarnessObservesExitAndJobResult/guard_steps_have_no_continue_on_error_or_status_override` |
| 5 | DUP-review: `ok=false` and `state=fail` written ahead of the terminating path (review step) | 17:34:05Z, 29.35, 83% | 79/29 | `GuardStepsOutputOneValuePerKey/review_step_clean_pass_each_key_once`, `.../review_step_changes_requested_each_key_once`, `.../review_step_draft_each_key_once`, `.../review_step_review_absent_each_key_once`, `.../output_pairs_within_allowed_set`, `DraftIsNotRunNeverSuccess/draft_state_output_is_single_value_not_run_draft` and `.../draft_ok_output_false`, `ReviewAbsentIsFailure/review_absent_state_fail_not_not_run`, 11 `ReviewFailureFailsStepAndJob` rows including `clean_pass_step_exit_0_job_success`, `StepHarnessObservesExitAndJobResult/harness_records_exit_status_and_outputs_per_step` |
| 6 | DUP-files: the same pre-write in the files step | 17:35:28Z, 29.79, 83% (17:34:38Z 39.69, 17:34:42Z 38.44, 17:34:51Z 38.08 not admitted) | 93/15 | `GuardStepsOutputOneValuePerKey/files_step_clear_each_key_once`, `.../files_step_refusal_each_key_once`, `.../output_pairs_within_allowed_set`, 8 `FilesStepFailureFailsStepAndJob` rows, `StepHarnessObservesExitAndJobResult/harness_records_exit_status_and_outputs_per_step` |
| 7 | DRAFT-OK: a draft mapped to `ok=true` | 17:36:44Z, 25.57, 83% (17:35:57Z 35.44 not admitted) | 101/7 | `DraftIsNotRunNeverSuccess/draft_ok_output_false`, `.../draft_step_exit_11_job_not_success`, `GuardStepsOutputOneValuePerKey/output_pairs_within_allowed_set`, `NotRunWithholds/exit_11_never_reaches_merge_condition` |
| 8 | CLEAN1: a clean pass that exits 1 | 17:37:44Z, 29.98, 83% | 94/14 | `ReviewFailureFailsStepAndJob/clean_pass_step_exit_0_job_success`, `FilesStepFailureFailsStepAndJob/clean_chain_step_exits_0_one_merge_job_success`, `MergeDecisionPath/clean_fixture_merges_once_with_head_pin`, `.../release_branch_clean_fixture_uses_merge_flag`, `.../refusal_fixtures_job_result_failure_clean_fixture_success`, `ReviewAbsentIsFailure/review_absent_window_open_keeps_waiting_no_output`, `WiresGuardAndGate/review_step_clean_ok_true_state_pass`, `StepHarnessObservesExitAndJobResult/harness_records_exit_status_and_outputs_per_step` |
| 9 | MISSING: a missing gate script read as a pass (files step) | 17:38:44Z, 31.27, 83% | 103/5 | `FilesStepFailureFailsStepAndJob/gate_script_missing_files_step_exit_nonzero_job_failure`, `.../chain_gate_script_missing_job_failure_zero_merges`, `WiresGuardAndGate/files_step_missing_script_not_ok` |
| 10 | H1 (test-side): a comparison on outputs only (exit and job result checks removed from `amCompare`) | 17:40:08Z, 28.47, 83% (17:39:40Z 32.41 not admitted) | 105/3 | `StepHarnessObservesExitAndJobResult/harness_comparison_rejects_exit_mismatch_with_matching_outputs`, `.../harness_comparison_rejects_job_result_mismatch_with_matching_exit` |
| 11 | H2 (test-side): the job-result derivation ignores a non-zero exit (`&& false`) | 17:41:39Z, 29.05, 83% | 74/34 | `StepHarnessObservesExitAndJobResult/harness_job_result_failure_when_a_step_exits_nonzero`, `.../harness_job_result_success_only_when_all_steps_zero_or_skipped`, plus every exit-status job-failure row of the five behaviour tests |

No mutant stayed green. MF turned `gate_path_change_zero_merge_calls` red as required. The first H2 attempt (17:41:08Z, load
27.23) was a malformed mutant (`if false {` left `e` unused: `internal/template/auto_merge_codex_test.go:444:9: declared and not
used: e`, `FAIL ... [build failed]`, no test ran) and is not a result; it was corrected to the compilable `&& false` form and run
once (row 11). Its JSONL was overwritten by the corrected run.

Gaps of this round: (1) the slot owner is not on the first line of `moai slot status`, so "held by lane-22" rests on the
leader's message; (2) RED and GREEN ran while t1818 held the landing slot (label above); (3) the mutants ran the nine-test
selector set, not the whole package; (4) the workflow has never run on a GitHub runner, so a real step or job conclusion and
DS-1 are unobserved; (5) the final-HEAD scoped verification (landing selector list, vet, build, actionlint) is not part of the
commit that carries this record; whatever of it ran is recorded in a following docs-only commit.

**Final-HEAD scoped verification (window of 2026-10-19, tree `3109d8fd8`, the record commit; later commits are docs only).**
Every call: one Go call at a time, `-p 4`, `-count=1`, raw JSONL (or text for `-list`) redirected to
`/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/207b7b17-aa81-4311-8e47-c01448800dee/scratchpad/t1868-window/<name>.out`
(machine-local scratch). Expected names were fixed first with `go test -list` once per package: template 28 names (the 21 of the
landing list above plus `TestAutoMergeReviewRecheckBeforeMerge`, `TestAutoMergeReviewFailureFailsStepAndJob`,
`TestAutoMergeReviewAbsentIsFailure`, `TestAutoMergeDraftIsNotRunNeverSuccess`, `TestAutoMergeGuardStepsOutputOneValuePerKey`,
`TestAutoMergeStepHarnessObservesExitAndJobResult`, `TestAutoMergeFilesStepFailureFailsStepAndJob`), all 28 listed; guardstate 4 of 4;
config 3 of 3.

| Call | reading before it (UTC, load, mem free, slot line) | exit | result, from the JSONL |
|---|---|---|---|
| `go test -list` template | 17:43:45Z, 20.98, 83%, `slot go-test-heavy: held` | 0 | 28 names, `ok internal/template 0.381s` |
| `go test -json` template selector (28 names) | 17:43:53Z, 20.52, 83%, held | 0 | 305 pass, 0 fail, 0 skip (parents and subtests), `ok  github.com/modu-ai/moai-adk/internal/template 32.867s` |
| `go test -list` guardstate | 17:45:07Z, 23.97, 83%, held (17:44:33Z 36.71 not admitted) | 0 | 4 names, `ok internal/guardstate 0.286s` |
| `go test -json` guardstate (4 names) | 17:45:12Z, 26.70, 83%, held | 0 | 6 pass, 0 fail, 0 skip, `ok  github.com/modu-ai/moai-adk/internal/guardstate 0.095s` |
| `go test -list` config | 17:45:16Z, 25.04, 83%, held | 0 | 3 names, `ok internal/config 0.286s` |
| `go test -json` config (3 names) | 17:45:21Z, 25.52, 83%, held | 0 | 9 pass, 0 fail, 0 skip, `ok  github.com/modu-ai/moai-adk/internal/config 0.159s` |
| `go vet ./internal/template/ ./internal/config/ ./internal/guardstate/` | 17:45:27Z, 23.95, 83%, held | 0 | no output |
| `go build ./...` | 17:45:33Z, 26.44, 83%, held | 0 | no output |
| `actionlint .github/workflows/auto-merge.yml` | no Go call, no reading | 0 | no output |

The whole-package `internal/template` run was not run (landing slot). Gaps: the slot holder name is not on the first status line;
the pass counts include parent tests, so they are not the 239 the earlier note expected from `-v` lines (the selector now carries
seven more tests); `go build ./...` ran on this machine only (darwin), no cross-platform build.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-18
run_commit_sha: 23e6c61fb1   # last implementation commit; the record commit that carries this file follows it
run_status: complete-on-card-branch   # not landed locally; remote verified: not reached by this lane
ac_pass_count: 19
ac_fail_count: 0
ac_not_run_count: 0
regression_guard_acs: [AC-CCR-011, AC-CCR-013, AC-CCR-016, AC-CCR-017]   # no RED-now by nature; reds recorded as mutants
preserve_list_post_run_count: 0   # git diff --quiet <A> HEAD over the seven AC-CCR-016 paths: exit 0 (EV:2067-2109)
l44_pre_commit_fetch: not-performed   # the run reads no remote ref; HEAD and branch re-read before every commit instead
l44_post_push_fetch: not-applicable   # no push was made
new_warnings_or_lints_introduced: none-observed   # go vet, gofmt -l, actionlint exit 0; golangci-lint not run (Gap)
cross_platform_build:
  linux_or_darwin_native: "go build ./... exit 0"
  windows_amd64: "GOOS=windows GOARCH=amd64 go build ./... exit 0"
total_run_phase_files: 12
m1_to_mN_commit_strategy: "per milestone: RED test commit, then implementation commit (8 commits); records in a ninth commit"
mutants_recorded: 93   # runs: M1 30+1, M2 13, M3 5+31+1, M4 5+5, AC-017 2; 91 are mutants and red, 2 are expected-green controls (M3-D5b equivalent variant, M3-W9-control)
debts_disposed: [D38, D39, D40, D41]
audit_ready: true
```

Every one of the 19 acceptance criteria has recorded evidence above; the items the lane still owes at landing (the
whole-package `internal/template` run under the slot, the merged-local criteria re-check, AC-CCR-016 on the absorbed
`main`) are listed under Gaps and are not acceptance evidence of this run.

## §E.4 Sync-phase Audit-Ready Signal

audit_ready: true

```yaml
sync_phase_complete_at: 2026-10-18
sync_commit_sha: pending-backfill   # a commit cannot cite its own hash (spec-workflow Sync (close)); backfilled with the real SHA in a following commit
sync_status: completed   # documentation and SPEC close only; NOT landed locally and NOT remote verified (see Gaps)
ac_source: .moai/specs/SPEC-CI-CODEX-REVIEW-001/acceptance.md   # tier L
changelog_entry_position: "CHANGELOG.md [Unreleased] > Changed, first entry"
frontmatter_status_transitions:
  spec_md: "draft -> completed (single sync commit; updated: 2026-10-18 was already current, unchanged)"
  plan_md: "no frontmatter updated field and no status field; untouched"
  acceptance_md: "no frontmatter updated field and no status field; untouched"
```

**Sync record** (tree HEAD `ccaccff7fe` before the sync commit; every command below was run in this sync pass)

- **Status transition.** The run phase did not move `spec.md` off `draft` (the plan-artifact hash froze it), so the `draft -> in-progress` step was never recorded; this single sync commit carries the terminal `completed` value, as the 3-phase close assigns to the sync owner. Only the `status:` line of `spec.md` and the `status:` line at the top of this file changed; no other part of `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, or `decision-index.md` was edited, and §E.1 to §E.3 are untouched.
- **B12 self-test.** (1) pre-emission `grep -c 'SPEC-CI-CODEX-REVIEW-001' CHANGELOG.md` -> 0 before the entry. (2) AC count from `acceptance.md` (tier L) with the reserved-token counter, run inline: stdout `19`, stderr `live=19 excluded=0 ambiguous=0`, exit 0, matching the entry's "19 acceptance criteria"; the 17 requirements counted from `spec.md` (`REQ-CCR-001..017`; `REQ-CCR-019` occurs only in the HISTORY row for plan revision 2 as a retired identifier). (3) every repository path the entry cites was checked with `ls`: all exist except `.github/workflows/claude.yml` and `.github/workflows/review-quality-gate.yml`, which the entry says are deleted and which `ls` reports absent.
- **Figures in the entry** were taken from §E.2 and §E.3 of this file (239 / 6 / 9 `--- PASS`, 93 mutant runs with 91 red and 2 controls, the run commit range `3dcb8e6c64`..`23e6c61fb1`), and the gate behaviour was read from `scripts/ci/codex-review-gate.sh` and `.github/workflows/auto-merge.yml` in this tree, not from a summary.
- **MX validation (sync sub-step).** No Go source file changed in the run (bash, YAML and `_test.go` only), so no exported identifier was added and no `@MX` annotation is owed; none was added.
- **Doc-surface decision.** No shipped documentation was edited (the card boundary names doctrine, `AGENTS.md`, and template wording as known debt KD-1). `grep -rnE 'review-quality-gate|claude\.yml|claude-code-review'` over `README*.md`, `docs-site`, `.claude/rules`, and `.moai/docs` (`*.md`) matched nothing. The CodeRabbit-naming sentences that the run left stale are listed in the sync report and in `spec.md` §7 KD-1; the lines observed in this tree: `.claude/rules/moai/workflow/factory-dispatch-core.md:46-50`, `factory-dispatch.md:146-150`, `factory-dispatch-gates.md:10-14`, `factory-dispatch-mechanics.md:154-155`, `factory-dispatch-detail.md:86,100,112`, `factory-dispatch-cards.md:32`, `.claude/agents/moai/manager-lead.md:224,236`, `AGENTS.md:200-203`, `internal/hook/session_start_factory_i18n.go:140,200,260,319`, plus the template mirrors under `internal/template/templates/` and the `.moai/policies/workflow/` projections.
- **Lint.** `moai spec lint SPEC-CI-CODEX-REVIEW-001`: exit 0, `0 error(s), 0 warning(s)`, one INFO `OwnershipTransitionUnmeasured` (the plan commit `8cd1d0d51d` carries no `Authored-By-Agent` trailer). `moai spec lint --strict SPEC-CI-CODEX-REVIEW-001`: exit 0, identical output. Both were run with a binary built from this tree by `go build -o <scratchpad>/moai-t1868 ./cmd/moai` (it prints the default `moai-adk v3.1.3`; no commit is baked in, so the judging build is identified by this build step against HEAD `ccaccff7fe`).
- **Gaps.**
  - Remote verification is not reached: the workflow has never run on a GitHub runner, no real review by `chatgpt-codex-connector[bot]` has been observed (DS-1), and no CI run of the SHA exists. Under the shipped contract every automated merge is withheld until a real `APPROVED` review bound to the head SHA is observed; Q20 and FU-10 stay open.
  - Not landed locally: no merged-local SHA exists at this commit; the lane's landing re-measure (whole `internal/template` package once under the heavy-run slot, AC-CCR-016 on the absorbed `main`) and the sync audit come after this commit. No test, `go build`, or `actionlint` was run in this sync pass.
  - The lineage act named in `spec.md` (marking `SPEC-V3R6-CODERABBIT-ADOPTION-001` partially superseded) was not performed: it edits another SPEC outside this commit's pathspec and is left to the leader or `manager-spec`.
  - KD-1 and FU-1..FU-9 are untouched, by the card boundary.
  - The `spec_audit` MCP call (`filter_spec=SPEC-CI-CODEX-REVIEW-001`, `project_root` = this tree, run on the working tree after the edits) returned `modern_era_clean: 1`, `grandfathered: 0`, one INFO `EraAutoDetected` (`H-4 (§E.2 + §E.4 + sync_commit_sha)`), no drift finding. The running MCP server is build `ee981a932`, an ancestor of this HEAD (`git merge-base --is-ancestor ee981a932 HEAD` exit 0), so it judged the tree with an older build; whether `internal/spec` changed in that range was not checked, and the reading is advisory.
- **Residual risk.** `sync_commit_sha` stays a placeholder until the backfill commit lands, so `moai spec audit` may read this SPEC as owing a backfill meanwhile. `status: completed` records the 3-phase close of this card's documentation and SPEC work; it is not a statement that the Codex review gate is active on GitHub.

## §G Override and Refusal Record

- 2026-10-18T13:54:17Z SPEC-CI-CODEX-REVIEW-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 5 >= tier ceiling 3 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1868/.moai/reports/t1868/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1868/.moai/reports/t1868/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1868/.moai/reports/t1868/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1868/.moai/reports/t1868/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1868/.moai/reports/t1868/plan-audit.md
- 2026-10-18T13:55:11Z SPEC-CI-CODEX-REVIEW-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 5 >= tier ceiling 3 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1868/.moai/reports/t1868/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1868/.moai/reports/t1868/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1868/.moai/reports/t1868/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1868/.moai/reports/t1868/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1868/.moai/reports/t1868/plan-audit.md

## Repair run (t1868 audit debt F1-F3, F5, F6)

Authority: operator ruling that all six optional findings of `.moai/reports/t1868/sync-audit-1.md` (PASS-WITH-DEBT 0.91, receipt `rcpt-d931c7048c8d65433ee30425`, tree `af27f744c5`) are repaired on this card; F4 (re-trigger after the review window) is not part of this run. No gate decision changed: every merging path still needs the head-bound trusted Codex summary, the clean receipt, and the separate functional and security checks; every hold path still exits 0. `.github/required-checks.yml` is unchanged (`git diff --stat origin/main -- .github/required-checks.yml` printed nothing, exit 0). Baseline tree for this run: `af27f744c5`, clean.

### Repairs

| Finding | Repair | Files |
|---|---|---|
| F1 | The `pr` step computes `active`: `false` when the number is empty, the state is `closed`, or `merged` is `true`; otherwise `true`. Checkout, head, author, checks, files, review and merge all require `steps.pr.outputs.active == 'true'`. The old "Skip" step is kept as a log-only notice with the inverse condition. | `.github/workflows/auto-merge.yml` |
| F2 | The merge step takes `PR_NUMBER` through its `env:`; no `run:` script carries a `${{ }}` expression any more. | `.github/workflows/auto-merge.yml` |
| F3 | The merge step writes one line to `$GITHUB_STEP_SUMMARY` for each outcome, always with the head SHA: `merge held: review window elapsed before merge`, `merge held: review window elapsed during head verification`, `merge not executed: head moved` (tested and current head), `merge failed: gh pr merge exit N (reason)`, `merge requested`. Exit codes unchanged. | `.github/workflows/auto-merge.yml` |
| F5 | `TestNoHardcodedContexts` lost the target `internal/cli/github_init.go` (deleted in `7d94cbc545`; grep found no file now carrying its init-time role, so no replacement target) and an unreadable scan target now fails the test instead of logging a skip. | `internal/config/required_checks_test.go` |
| F6 | The existing `Lint` job in `ci.yml` gains a hard step `shellcheck scripts/ci/codex-review-gate.sh`. shellcheck is not installed here (no system package installed); the script was reviewed by hand against the shellcheck rules (SC2016 on single-quoted jq programs is exempt for assignments and for `jq` arguments; SC2155 absent because no declaration carries an assignment; every expansion quoted; `read -r` throughout; no `$?` after a pipeline). | `.github/workflows/ci.yml` |

Design decision on F1: the audit suggested `steps.pr.outputs.state == 'open'`. That condition would stop every real merge if the `workflow_run` payload's `pull_requests[0]` element omits `state` and `merged` (from memory of GitHub's event documentation the element carries id, number, url, head, base only; that was not re-read and no live payload was observed in this run, see Gaps). The `active` output therefore turns the job off only on a positive signal and never reads an absent field as "not open"; an open pull request with the fields absent behaves as before (control `state_absent_payload`).

Harness changes in `internal/template/auto_merge_codex_test.go` (support only, no assertion weakened): `GITHUB_STEP_SUMMARY` is set per step; the renderer resolves `github.event.workflow_run.pull_requests[0].<field>` from `amCtx.prEvent`; the seeded `pr` outputs gain `active: "true"` (what the `pr` step now produces for an open pull request).

### RED on the unrepaired tree (tree `af27f744c5`, new tests only)

Command: `go test ./internal/template -run 'TestAutoMergeSkipsPullRequestThatIsNotOpen|TestAutoMergeNeverInterpolatesExpressionsIntoScripts|TestAutoMergeSummarisesEveryMergeOutcome|TestCILintJobShellchecksGateScript' -count=1 -v` -> exit 1, `FAIL github.com/modu-ai/moai-adk/internal/template 3.192s`. (The F6 test was renamed afterwards to `TestAutoMergeGateScriptIsShellcheckedInLint` so the lane's selector sweeps it.) Assertion lines, verbatim from the run log:

- F1: `auto_merge_repair_test.go:71: step "Verify PR head still matches the CI-tested commit" does not require steps.pr.outputs.active == 'true': ""` (also checkout, author, checks, files, review, merge); `auto_merge_repair_test.go:108: steps ran for a pull request that is not open: [{key:head code:0 coe:false} {key:author code:0 coe:false} {key:checks code:0 coe:false} {key:files code:0 coe:false} {key:review code:0 coe:false} {key:merge code:0 coe:false}]` (closed and merged cases; the number-less case stopped at `files`). On the unrepaired workflow a closed pull request with a number runs the whole chain up to the merge step in the harness; GitHub's server side refusing the merge of a closed pull request was not exercised. Controls `open` and `state_absent_payload` PASS on the unrepaired tree, as they must.
- F2: `auto_merge_repair_test.go:139: step "Auto merge PR" interpolates an expression into its script`; `auto_merge_repair_test.go:149: merge step env PR_NUMBER = ""`; `auto_merge_repair_test.go:164: the pull request number was executed as shell code` (the marker file was created by the hostile number `7"; : > <marker>; X="`).
- F3: `auto_merge_repair_test.go:225: summary lines [], want exactly one` (and the same at lines 230, 244, 249, 254, one per outcome subtest).
- F6: `auto_merge_repair_test.go:297: the Lint job has no shellcheck step`.

### GREEN after the repair

Same command (F6 test under its new name) -> exit 0, `ok github.com/modu-ai/moai-adk/internal/template 2.658s`; 19 new run entries, all `--- PASS`.

### Mutants (each: revert one repair in the working tree, run the four new tests, restore; restore verified byte-identical with `cmp`)

| Mutant | Result |
|---|---|
| `ACTIVE=false` -> `ACTIVE=true` in the `pr` step | red (4 subtests: closed, merged, merged-flag, no-number) |
| `active` term dropped from the merge `if` / the head `if` / the checkout `if` | red (static pin; head also red behaviourally) |
| `PR_NUMBER="${{ ... }}"` restored in the merge script | red (static sweep and hostile-number subtest) |
| each of the five summary lines deleted, one at a time | red, exactly the matching outcome subtest |
| head SHA dropped from the `merge requested` line | red |
| shellcheck step deleted / wrong script path / `|| true` appended | red (3 of 3) |

The F5 mutant: re-adding the deleted target `github_init.go` makes `TestNoHardcodedContexts` fail with `scan target ... github_init.go is unreadable: ... no such file or directory`; reverted.

### Swept counts

- Baseline, lane selector `go test ./internal/template -run 'Codex|AutoMerge|Retire|Review|InputFilter|BindingLedger|Parity' -count=1 -v` on `af27f744c5`: 611 `=== RUN`, 611 `--- PASS`, 0 `--- FAIL`.
- After the repair, same selector: 630 `=== RUN`, 630 `--- PASS`, 0 `--- FAIL` (611 + 19 new). The F6 test sweeps one `lint` job with its steps (non-empty asserted); the F1 static pin sweeps exactly 7 guarded steps; the F2 sweep covers at least 6 run scripts.
- Wider scoped runs, all green: `go test ./internal/config/... -count=1` (config, atomicfile, toolpolicy `ok`); `go test ./internal/guardstate/... ./internal/guardliveness/... -count=1` (`ok`); `go test ./internal/template -run 'Release|DotGithub|BranchProtection|Cutover|Pipefail|Provenance|Shard|WorkflowFiles|Workflow' -count=1` (`ok`); `go test ./internal/cli -run 'WorkflowSyncRef|CodexLiveAxis' -count=1` (`ok`).

### Verification (this run, this tree, logs under `.moai/state/verify-t1868/`)

- `go build ./...` -> exit 0.
- `go vet ./internal/config/... ./internal/template/...` -> exit 0.
- `bash -n scripts/ci/codex-review-gate.sh` -> exit 0.
- `actionlint .github/workflows/auto-merge.yml .github/workflows/ci.yml` (binary `/Users/goos/go/bin/actionlint`) -> exit 0, no output.
- `gofmt -l internal/template internal/config` -> no output.

### Gaps and residual risk of this run

- shellcheck was not run (not installed; not installed by this run). The F6 step is the first execution of shellcheck on the script, on the first CI run of this branch; a style-level finding I could not predict would turn the required `Lint` check red there. The step runs at the default severity on purpose.
- The `workflow_run` payload shape (whether `pull_requests[0]` carries `state` and `merged`) was not observed on a live run; the `active` computation is built to be correct either way.
- Neither workflow change has run on a GitHub runner; the summary lines and the skip are proven in the offline harness only.
- The two-second real-time delay in `review_window_elapsed_during_head_check` depends on wall-clock seconds; a very slow machine would take the earlier hold path, which carries the same summary prefix and still passes.
- Both edited workflow files are under `.github/workflows/` and need operator review before landing (`factory-dispatch-cards.md` § Workflow-file changes).
