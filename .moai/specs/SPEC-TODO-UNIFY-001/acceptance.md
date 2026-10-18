# Acceptance: SPEC-TODO-UNIFY-001

Verification layer. Each criterion is a binary-testable `Given ... When ... Then ...`; the GEARS requirements it verifies are in `spec.md`. Document-level pin: **every RED-now observation cited here was taken at tree `026f48658f`, except ledger entries E-15 to E-18, which carry their own pin `23e40eb7c4`**; that commit adds only plan artifacts and the audit report (`git diff --stat 026f48658f 23e40eb7c4 -- . ':!.moai/specs/SPEC-TODO-UNIFY-001' ':!.moai/reports/t1641'` printed nothing). The evidence ledger in `research.md` §R.2 carries every command, exit, and output; a cell cites it by id. Observations made through the installed `moai` come from build `ee981a932-dirty`, a strict ancestor 516 commits behind that tree, and are labelled as such.

Numbering note (revision 2): AC identifiers are stable across revisions where a criterion survives. Old AC-TU-021 and AC-TU-022 are merged into AC-TU-021; old AC-TU-023 is now AC-TU-022; AC-TU-023, AC-TU-025 are new; AC-TU-014 absorbs old AC-TU-019's closing-act clauses; AC-TU-019 is new (revocation); AC-TU-013 gained a wording clause.

## §1 How a criterion is adopted

Each criterion is a pair of cells. The **RED-now cell** is the criterion observed red on the pre-implementation tree, with the reason it is red. The **green path cell** names the milestone that flips it and the command whose output becomes the pass. A criterion with one cell is unadopted.

Three classes, stated so no criterion borrows a stronger status than it has:

- **A, observed at plan time.** A read-only command already shows the pre-implementation state (ledger id cited). It is release-blocking.
- **B, to be observed at run time.** The criterion is a Go test that does not exist yet; plan time cannot run it (the card forbids `go test` here). It is **unadopted until the owning milestone's RED step records its failing output** (the failure must be the stated assertion, not a compile or fixture error) in `progress.md` §E.2. Once recorded it is release-blocking. Until then it is a pending criterion, not a pass.
- **G, regression guard.** Green by design today; it protects something this change must not move. It never counts as evidence that the work was done.

Every Go-test command below is a name-exact selector that names tests which exist now or which the owning milestone creates at RED; the verdict is read only after the run prints at least one `--- PASS: <TestName>` line for each named test (an empty or partial selection prints a pass-shaped result and proves nothing).

## §2 Criteria

### Track A, one canonical name

**AC-TU-001 (REQ-TU-002), class A.** *Given* the `moai gtd` command built at the tip of this branch and a temporary queue, *when* each verb of the `moai todo` tree is invoked through `gtd` with valid arguments, *then* stderr carries exactly one warning whose first line begins `Warning: moai gtd has been renamed to moai todo`, that names the `moai todo` form and states removal one release after the release that first carries the warning, and stdout and exit status equal the `moai todo` form's.
- Green: `go test ./internal/cli -run '^TestGTDAliasWarnsThenDelegates$' -count=1 -v` (M5).
- RED-now: E-15 (no `renamed to moai todo` text in `gtd.go`; this is the proxy for the warning's lead phrase and it flips when the phrase lands), E-08 (the existing parity test requires equal stderr, so a warning cannot land without rewriting it), E-10 (installed build: `gtd list` prints no deprecation text). Red because the warning does not exist.
- Mutant: a warning written only by the parent command and not by a subverb fails the per-verb loop; a warning missing the removal statement fails the text assertion.

**AC-TU-002 (REQ-TU-003), class B.** *Given* the alias, *when* `gtd list --json` runs, when `gtd --help` runs, and when `gtd` is invoked with a mistyped verb, a missing required argument, and an unknown flag, *then* stdout of the first parses as JSON with no warning text, the help carries no warning on either stream, and each argument- or flag-refused invocation prints exactly the `todo` form's error (command name aside) with no warning and an exit status equal to the `todo` form's.
- Green: `go test ./internal/cli -run '^TestGTDAliasWarningStreams$' -count=1 -v` (M5).
- RED-now: recorded at M5 step 1 (the test is new). Basis for the refusal rows: cobra runs argument validation (`command.go:968`) before the persistent pre-run step (`:984`) and help before both (`:926-935`), so the warning's hook never runs for those invocations (`research.md` §R.11). Mutant: a warning on stdout fails the JSON parse; a warning printed from an argument-validation path fails the refusal rows.

**AC-TU-003 (REQ-TU-001), class B.** *Given* one fixture queue, *when* the same mutating verb (`add`, `hold`) runs once through `todo` and once through `gtd` on copies of it, *then* the resulting stores are equal and the `gtd` tree has no verb whose `Use`, flags, or argument rule differ from the `todo` tree's.
- Green: `go test ./internal/cli -run '^TestGTDSharesTodoTree$' -count=1 -v` (M4, M5).
- RED-now: recorded at M4 step 1. Mutant: a copied handler that diverges fails the store comparison.

**AC-TU-004 (REQ-TU-004), class B.** *Given* the command trees, *when* they are enumerated, *then* the `todo` tree carries exactly the 26 verbs it carries today plus `capture`, `clarify`, `organize`, `reflect`, `engage`, `answer` (32 in all), the `gtd` tree carries the same 32 and no other, no stage verb is registered on `gtd` alone, and the five-stage path (capture, clarify, organize, reflect incl. `--rebuild-projection`, engage with `--pick --dispatch`) run through the `todo` spelling produces a store equal to the same run through `gtd`.
- Green: `go test ./internal/cli -run '^(TestStageVerbsOnTodoTree|TestGTDAllTodoVerbsParity|TestGTDFiveStageCLIUsesSameSQLite)$' -count=1 -v` (M4; the second and third exist at `gtd_compat_test.go:49` and `:124`, the first is created at RED).
- RED-now: E-04 (`newGTDReflectCmd` is not registered by `todo.go`, so the "gtd alone" state exists today) and E-09 (installed build: `moai todo reflect` fails); the test's own red is recorded at M4 step 1.
- Mutant: a disposition that deletes any stage verb fails the exact-32 set; a verb registered but with a no-op handler fails the store-equality run; a verb left on `gtd` alone fails the no-gtd-only clause. The expected verb set is fixed in this criterion, not produced by the run phase.

**AC-TU-005 (REQ-TU-005), class A.** *Given* the queue guidance corpus (workflow skill in both trees, router skill in both trees, the four READMEs, the eight docs-site queue pages, help `Long` texts), *when* every `moai todo <word>` and `moai gtd <word>` token is resolved against its command tree, *then* every token resolves to an executable verb, and the scan reports at least 300 tokens across at least 17 files.
- Green: `go test ./internal/cli -run '^TestGuidanceVerbsResolve$' -count=1 -v` (M4).
- RED-now: E-03 (six guidance lines name `moai todo capture|clarify|organize|reflect|engage|answer`) with E-04 and E-09. Red because six named verbs do not exist on `moai todo`.
- Floors (measured, `research.md` §R.5b): 17 files carry 326 matching lines at tree `23e40eb7c4`, and tokens are at least lines; the floors 300 and 17 leave slack for edited wording. Mutant: a scan that finds no tokens, or reads one file, fails the floors.

**AC-TU-006 (REQ-TU-006), class B.** *Given* each stage verb, *when* it is invoked in a lane session, *then* `reflect` without `--rebuild-projection` is admitted, `reflect --rebuild-projection`, `capture`, `clarify`, `organize`, `engage`, and `answer` are refused with the standing refusal text, and a stage verb with no declared classification fails the test.
- Green: `go test ./internal/cli -run '^TestLaneClassifiesStageVerbs$' -count=1 -v` (M4).
- RED-now: recorded at M4 step 1. Observed input: in this lane session `moai gtd reflect --json` exited 1 with the lane-boundary refusal although `reflect` without the flag reads only (`research.md` §R.2 observation, §R.12 for the flag).

**AC-TU-007 (REQ-TU-007), class A.** *Given* the template sources and the deployed tree, *when* the surface test runs, *then* `/moai:todo` dispatches `todo`, `/moai:gtd` dispatches `todo` and carries the one-line rename notice, the published skills equal the publisher output (17 names), the router skill lists `todo` canonical and `gtd` alias, the queue workflow document at its unchanged path names `moai todo` canonical in its header, and the existing doc tests still pass.
- Green: `go test ./internal/template -run '^TestTodoCanonicalSurfaceGolden$' -count=1 -v` and `go test ./internal/cli -run '^(TestTodoSkillDocumentsHistoryVerb|TestTodoSkillDocumentsJevSource)$' -count=1 -v` (M6; the golden is created at RED by inverting `TestGTDCanonicalSurfaceGolden`, the two doc tests exist at `todo_skill_doc_test.go:17` and `todo_skill_doc_parity_test.go:19`).
- RED-now: E-13 (the current golden requires the `todo` command file to contain `arguments: ` `` `gtd` `` ` $ARGUMENTS`, the inverse of the requirement). Red because the surface test pins the old arrangement.

**AC-TU-008 (REQ-TU-008), class A.** *Given* the executable consumers, *when* `git grep -n -E "'gtd'|moai gtd" -- mods/moai-board/hooks/data.ts scripts/cutover-precheck.sh internal/cli/factory_card.go` runs, *then* it prints nothing and exits 1.
- RED-now: E-14 (8 lines, exit 0). Green path: M7.
- Gap: the board mod's JavaScript tests (`mods/moai-board/tests/**`) are not run by a Go selection; M7 records the runner used or lists them under Gaps.

**AC-TU-009 (REQ-TU-009, REQ-TU-010), class B.** *Given* a temporary tree holding one file per substitution class and one per excluded class (including `AGENTS.md`), *when* the script runs in dry-run, then apply, then apply again, *then* dry-run lists exactly the substitutable files, apply rewrites them as the class table says, every excluded file is byte-identical, and the second apply reports zero replacements and changes no file.
- Green: `go test ./internal/template -run '^TestGTDToTodoSubstitute$' -count=1 -v` (M7; the host package is `internal/template`, the test executes the script against the temporary tree).
- RED-now: recorded at M7 step 1. Mutant: a script that rewrites bare `gtd` changes an excluded identifier file and fails the byte comparison; one that rewrites `AGENTS.md` fails the excluded-class row.

**AC-TU-010 (REQ-TU-011), class A.** *Given* the tracked guidance corpus and the allowlist file, *when* the residual test runs, *then* no `moai gtd` or `/moai gtd` token appears outside the allowlist, the swept file count is at least 5,000, the paths that carry the spelling today and are not excluded are all among the swept set, and the planted-mutant fixture makes the check fail.
- Green: `go test ./internal/template -run '^TestNoResidualGTDSpelling$' -count=1 -v` (M8, M9); after M8 the paths E-01 prints must all be allowlisted lines or excluded classes (the test compares the two).
- RED-now: E-01 (73 paths carry `moai gtd`). Red because every one of them is residue until M8; E-01 stays non-empty after the work (allowlisted lines and recorded fixtures remain), so the criterion's own command is the test, not E-01.
- Floor (measured, `research.md` §R.5b): 16,151 tracked files, 9,090 in scope before the binary filter; the floor 5,000 is below that upper bound; the binary share was not measured.

### Track B, the done path

**AC-TU-014 (REQ-TU-015, REQ-TU-020), class B.** *Given* a lane session whose lane holds card X, X has landing evidence of the closed form, a factory-linked record in `merged-local`, and the three facts of REQ-TU-017, *when* the lane runs `moai todo done X`, and *when* the leader's `auto-done` scan closes a second such card, *then* each call is admitted, the factory row ends `done` (edge T29) and the queue row is archived, in that order, without an operator act, each log row names its closing path, and the auto-done scan's three misfire guards still skip the cards they skipped before.
- Green: `go test ./internal/cli -run '^(TestTodoLaneOwnCardDone|TestTodoCloseBothRecords)$' -count=1 -v` and the existing auto-done family read from `todo_autodone_test.go` at RED time (M2b).
- RED-now: recorded at M2b step 1. Observed input: E-07 (`done` appears in no allowlist map in `todo.go`); the refusal itself was read from `todo.go` and was deliberately not executed against the live queue.
- Mutant: a close that archives the queue first and requests the factory edge second fails the order assertion.

**AC-TU-015 (REQ-TU-016), class B.** *Given* the same fixture, *when* a lane runs `done` for a card held by another lane, for X before its landing evidence exists, or any of `add`, `drop`, `edit`, `hold`, `claim`, *then* each is refused with the standing lane-boundary text and the queue file is byte-identical afterwards.
- Green: `go test ./internal/cli -run '^TestTodoLaneDoneRefusalMatrix$' -count=1 -v` (M2b).
- RED-now: recorded at M2b step 1. Mutant: a guard that admits `done` for any card fails the other-lane row; one that skips the evidence check fails the no-evidence row.

**AC-TU-016 (REQ-TU-017), class B.** *Given* a factory-linked card, *when* the close reaches the approval gate with each of the eight combinations of the three facts (state `merged-local`; a committed `sync-audit.md` carrying `verdict: PASS` or `PASS-WITH-DEBT` whose `audited_sha` is an ancestor of the integration branch; a re-measure record keyed to the merge commit's tree that satisfies the merge gate), *then* exactly the combination with all three issues a receipt with issuer `auto` and the leader role marker, which the unchanged gate then verifies; each of the three two-of-three combinations, and every other combination, issues nothing, writes nothing, refuses the close with the receipt-absent text, and leaves both records unchanged; and rows with `verdict: FAIL`, an uncommitted verdict file, and an `audited_sha` that is not an ancestor each issue nothing.
- Green: `go test ./internal/homestate ./internal/cli -run '^(TestAutoReceiptThreeFacts|TestAutoReceiptIssuer)$' -count=1 -v` (M2a, M2b).
- RED-now: E-18 (no `auto` issuer and no revocation code exist in `internal/homestate` or `internal/cli`); the test's own red is recorded at M2a step 1.
- Mutant: an issuer that issues on two of the three facts fails each two-of-three row; one that issues for the wrong card or run fails the unchanged gate's binding rows; one that records a leader label instead of `auto` fails the issuer assertion.

**AC-TU-017 (REQ-TU-018), class A.** *Given* the transition table, *when* `git grep -n -E '"T18"|isReservedEdge|ErrReservedEdge' -- internal/homestate/card_transition.go` runs, *then* it prints nothing and exits 1; and the table test still reports 71 accepted and 370 refused; and `pushed → ci-green` fails with the undeclared-transition error, not the reserved-edge error.
- Green: the grep above (exit 1) plus `go test ./internal/homestate -run '^(TestFR_AC005_TransitionTableEdgeCount|TestReservedEdgeRemoved)$' -count=1 -v` (M1; the first exists at `fr_transition_test.go:127`, the second is created at RED).
- RED-now: E-02 (six lines, exit 0) and E-12 (the pair-count pin, unchanged by design). Red because the row and the reserved branch exist.

**AC-TU-018 (REQ-TU-019), class B.** *Given* a repository **without a remote** holding a `merged-local` card whose merge commit is reachable from the local integration branch and a leader-issued receipt bound to the card's current version, *when* the push gate runs, *then* the card reaches `done` with no push step and the receipt is re-stamped to the new version; and *given* the same card with the branch reset so the merge commit is unreachable, or with no receipt, *then* the request is refused and the card row is byte-identical; and *given* a repository **with** a remote, *then* T29 admits only when the merge commit is reachable from the remote-tracking landed ref, and the push gate still requests `pushed` for it.
- Green: `go test ./internal/homestate ./internal/cli -run '^(TestLandedEdgeT29NoRemote|TestLandedEdgeT29WithRemote|TestPushGateTargetByRemote|TestPushGateDeclaredEdgesOnly)$' -count=1 -v` (M1; the third exists at `f7_gate_owner_admission_test.go:14` and is rewritten, not deleted; the others are created at RED).
- RED-now: E-17 (no edge `T29` exists, so the no-remote completion rests on row `T18` alone); the RED order at M1 puts these tests first.
- Mutant: a renamed T18 (human decider required, refused while a remote exists, no merge-reachability check) fails the remote row and the unreachable-merge row; a T29 without the receipt gate fails the no-receipt row.

**AC-TU-019 (REQ-TU-025), class B.** *Given* a card holding an automatically issued receipt, *when* the leader revokes it, *then* the store records the revocation, the close is refused with the receipt-absent text, and a second automatic issuance for that card and run does not occur even with all three facts present; *and* a lane session running the revocation is refused; *and* a leader-issued receipt for another card is unaffected by the revocation of the first.
- Green: `go test ./internal/homestate ./internal/cli -run '^TestAutoReceiptRevocation$' -count=1 -v` (M2a).
- RED-now: E-18; the test's own red is recorded at M2a step 1. Mutant: a revocation that deletes nothing, or that lets issuance recur, fails the second-issuance row.

**AC-TU-020 (REQ-TU-021), class B.** *Given* twelve cards of which ten carry landing evidence of the closed form, the three facts, and a `merged-local` factory row, and two do not carry landing evidence, *when* the closing acts run, *then* exactly the ten end `done` in both records and the two stay open in both; and, at sync, the live read-only listing of the ten most recent cards with landing evidence and their states is recorded with its command and output.
- Green: `go test ./internal/cli -run '^TestTenLandedCardsClose$' -count=1 -v` (M2b); the live listing is a sync-phase record, not a test.
- RED-now: recorded at M2b step 1. The negative control (two cards stay open) is the clause that makes an always-close mutant fail.

**AC-TU-023 (REQ-TU-020 recovery clause), class B.** *Given* a card with landing evidence and the three facts, *when* a fault is injected after the factory edge commits and before the queue archive, *then* the act exits non-zero naming both records and their states (factory `done`, queue open), a repeat of the same act archives the queue row without a second factory edge request and exits 0, and a third run reports the card already closed without writing.
- Green: `go test ./internal/cli -run '^TestCloseFailureBetweenRecords$' -count=1 -v` (M2b).
- RED-now: recorded at M2b step 1. Mutant: a repeat that requests the factory edge again fails on the stale-version refusal row; an act that archives first leaves the factory row open and fails the state assertion.

### Track C, the spec_id setter

**AC-TU-021 (REQ-TU-022, REQ-TU-023), class B.** *Given* a `picked` card and an `assigned` card, both with empty spec_id in both carriers, *when* the leader runs the setter with a well-formed identifier, *then* the factory row and the queue row both carry it and one change event exists; and *given* the same cards, *when* the setter is invoked with a different identifier on a card that has one in either carrier, on a `leased` or later card, on a terminal card, with a malformed identifier, or from a lane session, *then* each is refused and neither carrier changes; and an identical identifier already in both carriers succeeds as a no-op.
- Green: `go test ./internal/cli ./internal/homestate -run '^(TestTodoSpecSetter|TestTodoSpecSetterLaneRefused|TestCardSpecSetOnce)$' -count=1 -v` (M3a, M3b).
- RED-now: E-06 (no verb named `spec` is registered) is weak evidence; the criterion's red is recorded at M3a step 1. Mutant: an overwrite-capable setter fails the different-identifier row; a setter that writes only the queue carrier fails the factory-row assertion.

**AC-TU-022 (REQ-TU-024), class B.** *Given* a card whose spec_id the setter wrote, *when* kickoff's SPEC lookup runs, *then* it resolves `<worktree>/.moai/specs/<spec_id>` as for a pick-time identifier; and the existing tests of `next --spec`, `factory assign --spec`, and `unpick` pass unchanged.
- Green: `go test ./internal/homestate ./internal/cli -run '^(TestKickoffAfterSpecSetter|TestTodoPickInFactoryRecordsCardAndEvent|TestTodoPickInFactoryCapturesLinkedWorktreeSpecAndHEAD|TestFactoryAssignReassignHappyPath|TestFactoryAssignReassignRefusals|TestTodoUnpick_RevertsPickedToQueued|TestTodoUnpick_RefusalsLeaveFileUntouched)$' -count=1 -v` (M3b; all but the first exist: `todo_test.go:121`, `:214`, `:1078`, `:1123` and `factory_assign_reassign_test.go:41`, `:70`).
- RED-now: recorded at M3b step 1 (the kickoff test is new).

**AC-TU-025 (REQ-TU-022 recovery clause), class B.** *Given* a `picked` card with empty carriers, *when* a fault is injected in the queue write, *then* both carriers are unchanged; *when* a fault is injected after the queue write and before the factory commit, *then* the queue carrier holds the identifier, the factory carrier is empty, the setter exits non-zero naming both carriers, a repeat invocation with the same identifier fills the factory carrier and exits 0, and a repeat with a different identifier in that half-written state is refused.
- Green: `go test ./internal/homestate ./internal/cli -run '^(TestCardSpecSetFailureBetweenCarriers|TestTodoSpecSetterPartialRepeat)$' -count=1 -v` (M3a, M3b).
- RED-now: recorded at M3a step 1. Mutant: a setter that rolls the queue write back on a later failure it cannot undo, or that leaves a divergent value in either carrier, fails the repeat rows.

### Cross-cutting guards

**AC-TU-011 (REQ-TU-012), class G.** *Given* the branch tip and `CARD_BASE` (the value `git merge-base main HEAD` prints when measured, after absorbing `main`), *when* `git diff --name-only <CARD_BASE>..HEAD -- internal/factory/backlog_gtd_schema.go 'internal/factory/gtd_*.go' internal/template/testdata/binding_ledger.json internal/cli/testdata` runs, *then* it prints nothing; and the behavioural checks pass: the `gtd-<16 hex>` item pattern still matches an id issued before the change and rejects a malformed one (`gtdIDPattern` in `internal/cli/gtd.go`), the answers path is still `.moai/gtd/answers` (`internal/cli/gtd_answer.go`), and the five rule files' `paths:` keys still name `workflows/gtd.md`. The diff's control, the same range over `internal/cli/todo.go`, prints at least one path (a zero control reports "not measurable", not "unchanged").
- Green: `go test ./internal/cli -run '^(TestGTDFiveStageCLIUsesSameSQLite|TestGTDCLIInputRefusalsDoNotFallThrough|TestGTDAnswerWritesAppendOnlyRecord|TestGTDAnswerRejectsInvalidID|TestGTDAnswerCardUntouched)$' -count=1 -v` for the item pattern and the answers path, and `go test ./internal/template -run '^TestWorkflowRulePathsPinned$' -count=1 -v` for the `paths:` keys (all six names exist: `gtd_compat_test.go:124` and `:164`, `gtd_answer_test.go:13`, `:43`, `:52`, `workflow_rule_paths_pinned_test.go:24`).
- Not red now by design: it guards the Q3 decision.

**AC-TU-012 (REQ-TU-013), class G.** *Given* the project build run on the final tree, *when* `git status --short` runs, *then* it lists only the intended change set (the regenerated embedded catalog and published skills included), and the template mirror and drift tests named at M6 pass.
- Not red now by design.

**AC-TU-013 (REQ-TU-014), class A for the wording clause, class G for the ceiling clause.** *Given* the final `AGENTS.md` and its template mirror, *when* the wording test runs, *then* the `moai todo` row names `moai todo` canonical and `moai gtd` a warning alias and contains no "fully supported"; and the Codex contract ceiling guard passes with `wc -c AGENTS.md` at or below 24,576 (baseline E-11: 23,780).
- Green: `go test ./internal/config ./internal/template -run '^(TestCodexContractByteCeiling|TestAgentsMDTodoRowWording)$' -count=1 -v` (M9; the first exists at `internal/config/token_budget_guard_test.go:84` and measures the live file and the template mirror, the second is created at RED).
- RED-now (wording clause): E-16 (both `AGENTS.md:344` and `AGENTS.md.tmpl:343` say `fully supported compatibility alias`). Mutant: a row that keeps "fully supported" or drops "warning alias" fails the wording test.

**AC-TU-024 (all requirements), class G.** *Given* the change set, *when* the set of `t.Skip`, `t.Skipf`, and `t.SkipNow` call lines in the changed test files is listed at `CARD_BASE` and at the tip with `git grep -n -E 't\.Skip(f|Now)?\(' -- <changed test files>`, *then* the two lists are equal; and no test selector or invocation raises a timeout value relative to the baseline.
- Not red now by design.

## §3 Edge cases

- A `moai gtd` invocation whose first word is a mistyped verb followed by a card id (`moai gtd lst 401`): the refusal comes from the command's argument validation, before the pre-run step, so no warning prints and the output equals the `todo` form's error (AC-TU-002 row).
- `moai gtd` with no arguments renders the queue (the bare parent form): one warning, then the queue.
- `--help` on a subverb of `gtd`: no warning (AC-TU-002).
- A lane session running `done` with a card id in the `t<n>` form and in the bare `<n>` form: both resolve to the same own card (the existing normalisation).
- A card with landing evidence recorded by the operator (`todo landed --sha`) versus a subject on the landed ref: both are the closed forms; the lane path accepts either.
- A card whose verdict file exists but is not committed, or whose `audited_sha` is not an ancestor of the integration branch: no receipt is issued and the card waits for the leader's manual approval (AC-TU-016 rows).
- A card whose factory row is `pushed`: the automatic receipt is not issued; the card closes only under a leader-issued receipt and its factory row stays `pushed` (the edge is card t1639's).
- The setter on a card that was unpicked and re-picked: the unpick cleared the queue spec_id, so the setter fills it again (AC-TU-021/022).
- The substitution script on a path that contains a space or begins with a dash: the script quotes every path (AC-TU-009 fixture row).

## §4 Quality gate criteria

- TRUST 5 Tested: each milestone's new tests RED then GREEN with the output recorded; the changed packages' coverage does not fall below the baseline measured at the milestone's start; name-exact selectors for `internal/cli`.
- TRUST 5 Readable and Unified: `gofmt` clean on changed Go files; the repository's golangci-lint configuration run over the changed packages; shell script passes `sh -n` and the repository's shell lint if one exists.
- TRUST 5 Secured: the lane exception never reads a card identity from an environment variable alone (the held-card record decides); the automatic receipt is issued only after the three facts are re-read at the moment of issuance; the setter validates the identifier before any write; the script reads only `git ls-files` output.
- TRUST 5 Trackable: every commit message names card t1641; the sync record carries the AC matrix with RED-now and GREEN cells and the Gaps list.
- Windows build: `GOOS=windows GOARCH=amd64 go build ./...` exits 0 on the final tree.

## §5 Definition of Done

1. All 25 criteria adopted (class B criteria carry a recorded RED), and every class A and B criterion is GREEN with its verbatim output.
2. Q1 to Q9 each carry a recorded decision in `decision-index.md` (Q5 marked `lane-judgment, leader-confirmed 10-18 (conditional)`), and no clarification marker remains in any plan artifact.
3. `moai spec lint SPEC-TODO-UNIFY-001` exits 0 with no finding, using a `moai` built from the tip of this branch and invoked by path, with that build's commit recorded next to the tree HEAD.
4. The landing note lists the Gaps: the unrun remainder of `internal/cli`, the board mod's JavaScript tests, the docs-site build if the generator is unavailable, and any criterion left pending.
5. No `t.Skip*` added, no timeout raised, no hand edit of a file the substitution script owns.
