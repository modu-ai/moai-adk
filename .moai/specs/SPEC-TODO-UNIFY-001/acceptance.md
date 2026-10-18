# Acceptance: SPEC-TODO-UNIFY-001

Verification layer. Each criterion is a binary-testable `Given ... When ... Then ...`; the GEARS requirements it verifies are in `spec.md`. Document-level pin: **every RED-now observation cited here was taken at tree `026f48658f`** and is carried by an entry of the evidence ledger in `research.md` §R.2 (the ledger is the carrier; a cell cites it by id). Observations made through the installed `moai` come from build `ee981a932-dirty`, a strict ancestor 516 commits behind that tree, and are labelled as such.

## §1 How a criterion is adopted

Each criterion is a pair of cells. The **RED-now cell** is the criterion observed red on the pre-implementation tree, with the reason it is red. The **green path cell** names the milestone that flips it and the command whose output becomes the pass. A criterion with one cell is unadopted.

Three classes, stated so no criterion borrows a stronger status than it has:

- **A, observed at plan time.** A read-only command already shows the pre-implementation state (ledger id cited). It is release-blocking.
- **B, to be observed at run time.** The criterion is a Go test that does not exist yet; plan time cannot run it (the card forbids `go test` here). It is **unadopted until the owning milestone's RED step records its failing output** (the failure must be the stated assertion, not a compile or fixture error) in `progress.md` §E.2. Once recorded it is release-blocking. Until then it is a pending criterion, not a pass.
- **G, regression guard.** Green by design today; it protects something this change must not move. It never counts as evidence that the work was done.

Every Go-test command below is a name-exact selector; the verdict is read only after the run prints at least one `--- PASS: <TestName>` line for each named test (an empty selection prints a pass-shaped result and proves nothing).

## §2 Criteria

Criteria are numbered in requirement order. The four cross-cutting guards (AC-TU-011, AC-TU-012, AC-TU-013, AC-TU-024) are listed last, after the three tracks.

### Track A, one canonical name

**AC-TU-001 (REQ-TU-002), class A.** *Given* the `moai gtd` command built at the tip of this branch and a temporary queue, *when* each verb of the `moai todo` tree is invoked through `gtd`, *then* stderr carries exactly one warning that names the `moai todo` form and contains the removal-release constant, and stdout and exit status equal the `moai todo` form's.
- Green: `go test ./internal/cli -run '^TestGTDAliasWarnsThenDelegates$' -count=1 -v` (M5).
- RED-now: E-05 (`gtd.go` carries no deprecation text, exit 1) and E-08 (the existing parity test requires equal stderr, so a warning cannot land without rewriting it); E-10 (installed build: `gtd list` prints no deprecation text). Red because the warning does not exist.
- Mutant: a warning written only by the parent command and not by a subverb fails the per-verb loop.

**AC-TU-002 (REQ-TU-003; REQ-TU-002 constant), class B.** *Given* the alias, *when* `gtd list --json` runs and when `gtd --help` runs, *then* stdout of the first parses as JSON with no warning text, the second carries no warning on either stream, and the removal-release constant is non-empty.
- Green: `go test ./internal/cli -run '^TestGTDAliasWarningStreams$' -count=1 -v` (M5).
- RED-now: recorded at M5 step 1 (the test is new). Mutant: a warning on stdout fails the JSON parse.

**AC-TU-003 (REQ-TU-001), class B.** *Given* one fixture queue, *when* the same mutating verb (`add`, `hold`) runs once through `todo` and once through `gtd` on copies of it, *then* the resulting stores are equal and the `gtd` tree has no verb whose `Use`, flags, or argument rule differ from the `todo` tree's.
- Green: `go test ./internal/cli -run '^TestGTDSharesTodoTree$' -count=1 -v` (M4, M5).
- RED-now: recorded at M4 step 1. Mutant: a copied handler that diverges fails the store comparison.

**AC-TU-004 (REQ-TU-004), class B.** *Given* the usage inventory and its disposition table, *when* the command trees are enumerated, *then* every verb marked ABSORB is registered on `todo` (and so reachable through `gtd`), every verb marked DELETE is registered on neither, and no stage verb is registered on `gtd` alone.
- Green: `go test ./internal/cli -run '^TestStageVerbDisposition$' -count=1 -v` (M4).
- RED-now: E-04 (`newGTDReflectCmd` is not registered by `todo.go`, exit 1) shows the "gtd alone" state exists today; the test's red is recorded at M4 step 1.

**AC-TU-005 (REQ-TU-005), class A.** *Given* the queue guidance corpus (workflow skill in both trees, router skill, READMEs, docs-site queue pages, help `Long` texts), *when* every `moai todo <word>` and `moai gtd <word>` token is resolved against its command tree, *then* every token resolves to an executable verb, and the scan reports at least the measured token floor.
- Green: `go test ./internal/cli -run '^TestGuidanceVerbsResolve$' -count=1 -v` (M4).
- RED-now: E-03 (six guidance lines name `moai todo capture|clarify|organize|reflect|engage|answer`) with E-04 and E-09 (installed build: `moai todo reflect` exits 1, `Unknown command "reflect" for "moai todo"`). Red because six named verbs do not exist on `moai todo`.
- Mutant: a scan that finds no tokens fails the floor assertion.

**AC-TU-006 (REQ-TU-006), class B.** *Given* each absorbed stage verb, *when* it is invoked in a lane session, *then* the guard's result equals the declared classification (read-only: admitted; mutating: refused with the standing refusal text), and a verb with no declared classification fails the test.
- Green: `go test ./internal/cli -run '^TestLaneClassifiesStageVerbs$' -count=1 -v` (M4).
- RED-now: recorded at M4 step 1. Observed input: in this lane session `moai gtd reflect --json` exited 1 with the lane-boundary refusal although `reflect` reads only.

**AC-TU-007 (REQ-TU-007), class A.** *Given* the template sources and the deployed tree, *when* the surface test runs, *then* `/moai:todo` dispatches `todo`, `/moai:gtd` dispatches `todo` and carries the one-line rename notice, the published skills equal the publisher output, the router skill lists `todo` canonical, `workflows/todo.md` exists and `workflows/gtd.md` does not in both trees, and the live-versus-template parity tests read the renamed path.
- Green: `go test ./internal/template ./internal/cli -run '^(TestTodoCanonicalSurfaceGolden|TestTodoSkillDoc.*|TestTodoSkillDocParity.*)$' -count=1 -v` (M6; confirm the two doc-test names with `git grep -n '^func Test' -- internal/cli/todo_skill_doc_test.go internal/cli/todo_skill_doc_parity_test.go` before the run).
- RED-now: E-13 (the current golden requires the `todo` command file to contain `arguments: ` `` `gtd` `` ` $ARGUMENTS`, the inverse of the requirement). Red because the surface test pins the old arrangement.

**AC-TU-008 (REQ-TU-008), class A.** *Given* the executable consumers, *when* `git grep -n -E "'gtd'|moai gtd" -- mods/moai-board/hooks/data.ts scripts/cutover-precheck.sh internal/cli/factory_card.go` runs, *then* it prints nothing and exits 1.
- RED-now: E-14 (8 lines, exit 0). Green path: M7.
- Gap: the board mod's JavaScript tests (`mods/moai-board/tests/**`) are not run by a Go selection; M7 records the runner used or lists them under Gaps.

**AC-TU-009 (REQ-TU-009, REQ-TU-010), class B.** *Given* a temporary tree holding one file per substitution class and one per excluded class, *when* the script runs in dry-run, then apply, then apply again, *then* dry-run lists exactly the substitutable files, apply rewrites them as the class table says, every excluded file is byte-identical, and the second apply reports zero replacements and changes no file.
- Green: `go test <host package> -run '^TestGTDToTodoSubstitute$' -count=1 -v` (M7). The host is a Go test in `internal/template` or `internal/cli` that executes the script against the temporary tree; M7 chooses it and names it in `progress.md`.
- RED-now: recorded at M7 step 1. Mutant: a script that rewrites bare `gtd` changes an excluded identifier file and fails the byte comparison.

**AC-TU-010 (REQ-TU-011), class A.** *Given* the tracked guidance corpus and the allowlist file, *when* the residual test runs, *then* no `moai gtd` or `/moai gtd` token appears outside the allowlist, the swept file count is at least the floor, and the planted-mutant fixture makes the check fail.
- Green: `go test ./internal/template -run '^TestNoResidualGTDSpelling$' -count=1 -v` (M8, M9; the host package is confirmed at M7).
- RED-now: E-01 (73 paths carry `moai gtd`, exit 0). Red because every one of them is residue until M8.

### Track B, the done path

**AC-TU-014 (REQ-TU-015), class B.** *Given* a lane session whose lane holds card X, X has landing evidence of the closed form, and a factory-linked record, *when* the lane runs `moai todo done X`, *then* the call is admitted, the card is archived, the factory row ends `done` under the OQ-5 receipt policy, and the log row's source names the lane own-card path.
- Green: `go test ./internal/cli -run '^TestTodoLaneOwnCardDone$' -count=1 -v` (M2).
- RED-now: recorded at M2 step 1. Observed input: E-07 (`done` appears in no allowlist map in `todo.go`, exit 1); the refusal itself was read from `todo.go` and was deliberately not executed against the live queue.

**AC-TU-015 (REQ-TU-016), class B.** *Given* the same fixture, *when* a lane runs `done` for a card held by another lane, for X before its landing evidence exists, or any of `add`, `drop`, `edit`, `hold`, `claim`, *then* each is refused with the standing lane-boundary text and the queue file is byte-identical afterwards.
- Green: `go test ./internal/cli -run '^TestTodoLaneDoneRefusalMatrix$' -count=1 -v` (M2).
- RED-now: recorded at M2 step 1. Mutant: a guard that admits `done` for any card fails the other-lane row; one that skips the evidence check fails the no-evidence row.

**AC-TU-016 (REQ-TU-017), class B.** *Given* an admitted own-card `done` on a factory-linked card, *when* the close runs, *then* it passes through the approval gate, the receipt condition is decided by the one receipt function the OQ-5 answer defines, the log row carries the new source label, and `moai todo undone` reverses the close.
- Green: `go test ./internal/cli -run '^TestTodoLaneDoneApprovalGate$' -count=1 -v` (M2).
- RED-now: recorded at M2 step 1.

**AC-TU-017 (REQ-TU-018), class A.** *Given* the transition table, *when* `git grep -n -E '"T18"|isReservedEdge|ErrReservedEdge' -- internal/homestate/card_transition.go` runs, *then* it prints nothing and exits 1; and the table test still reports 71 accepted and 370 refused; and both `merged-local → done` and `pushed → ci-green` fail with the undeclared-transition error.
- Green: the grep above (exit 1) plus `go test ./internal/homestate -run '^TestFR_AC005_TransitionTableEdgeCount$' -count=1 -v` (M1).
- RED-now: E-02 (six lines, exit 0) and E-12 (the pair-count pin, unchanged by design). Red because the row and the reserved branch exist.
- Conditional on OQ-6 (c): under "defer", this criterion is reported unmet and the card's removal item is recorded as undelivered.

**AC-TU-018 (REQ-TU-019), class B.** *Given* a merged-local card in a repository with a remote and one without, *when* the push gate chooses a target, *then* the first yields `pushed`, the second yields the edge or path OQ-6 selected, and neither yields an edge the table does not declare.
- Green: `go test ./internal/homestate ./internal/cli -run '^(TestPushGateTargetByRemote|TestPushGateDeclaredEdgesOnly)$' -count=1 -v` (M1).
- RED-now: recorded at M1 step 1. The existing `TestPushGateTargetByRemote` (`f7_gate_owner_admission_test.go:14`) is rewritten, not deleted.

**AC-TU-019 (REQ-TU-020), class B.** *Given* a fixture queue and factory record, *when* the lane own-card path and the leader `auto-done` scan each close a card with landing evidence, *then* every record OQ-7 names (both, under the alternate) ends `done` without an operator act and the log names the closing path; and the auto-done scan's three misfire guards still skip the cards they skipped before.
- Green: `go test ./internal/cli -run '^(TestTodoCloseBothRecords|TestAutoDone.*)$' -count=1 -v` (M2).
- RED-now: recorded at M2 step 1.

**AC-TU-020 (REQ-TU-021), class B.** *Given* twelve cards of which ten carry landing evidence of the closed form and two do not, *when* the closing acts run, *then* exactly the ten end `done` in both records and the two stay open; and, at sync, the live read-only listing of the ten most recent cards with landing evidence and their states is recorded with its command and output.
- Green: `go test ./internal/cli -run '^TestTenLandedCardsClose$' -count=1 -v` (M2); the live listing is a sync-phase record, not a test.
- RED-now: recorded at M2 step 1. The negative control (two cards stay open) is the clause that makes an always-close mutant fail.
- Conditional on OQ-5 (c): the criterion is reported unmet.

### Track C, the spec_id setter

**AC-TU-021 (REQ-TU-022), class B.** *Given* a `picked` card and an `assigned` card, both with empty spec_id, *when* the setter is invoked with a well-formed identifier, *then* the queue row and the factory row both carry it and one change event exists.
- Green: `go test ./internal/cli ./internal/homestate -run '^(TestTodoSpecSetter|TestCardSpecSetOnce)$' -count=1 -v` (M3).
- RED-now: E-06 (no verb named `spec` is registered, exit 1) is weak evidence (the verb name is OQ-8-dependent); the criterion's red is recorded at M3 step 1.

**AC-TU-022 (REQ-TU-023), class B.** *Given* the same cards, *when* the setter is invoked with a different identifier on a card that has one, on a `leased` or later card, on a terminal card, or with a malformed identifier, *then* each is refused and neither store changes; and an identical identifier succeeds as a no-op.
- Green: `go test ./internal/cli ./internal/homestate -run '^TestTodoSpecSetterRefusals$' -count=1 -v` (M3).
- RED-now: recorded at M3 step 1. Mutant: an overwrite-capable setter fails the different-identifier row.

**AC-TU-023 (REQ-TU-024), class B.** *Given* a card whose spec_id the setter wrote, *when* kickoff's SPEC lookup runs, *then* it resolves `<worktree>/.moai/specs/<spec_id>` as for a pick-time identifier; and the existing tests of `next --spec`, `factory assign --spec`, and `unpick` pass unchanged.
- Green: `go test ./internal/homestate ./internal/cli -run '^(TestKickoffAfterSpecSetter|TestTodoNextSpec.*|TestFactoryAssignSpec.*|TestTodoUnpick.*)$' -count=1 -v` (M3; confirm the three existing names with `git grep` before the run).
- RED-now: recorded at M3 step 1.

### Cross-cutting guards

**AC-TU-011 (REQ-TU-012), class G.** *Given* the branch tip and `CARD_BASE` (the value `git merge-base main HEAD` prints when measured, after absorbing `main`), *when* `git diff --name-only <CARD_BASE>..HEAD -- internal/factory/backlog_gtd_schema.go 'internal/factory/gtd_*.go'` runs, *then* it prints nothing; its control, the same range over `internal/cli/todo.go`, prints at least one path (a zero control reports "not measurable", not "unchanged").
- Not red now by design: it guards the OQ-3 default.

**AC-TU-012 (REQ-TU-013), class G.** *Given* the project build run on the final tree, *when* `git status --short` runs, *then* it lists only the intended change set (the regenerated embedded catalog and published skills included), and the template mirror and drift tests named at M6 pass.
- Not red now by design.

**AC-TU-013 (REQ-TU-014), class G.** *Given* the final `AGENTS.md`, *when* the Codex contract ceiling guard in `internal/config` runs, *then* it passes, and `wc -c AGENTS.md` stays at or below 24,576 (baseline E-11: 23,780).
- Green: `go test ./internal/config -run '^TestCodexContractByteCeiling$' -count=1 -v` (the test exists at `internal/config/token_budget_guard_test.go:84`; it measures the live `AGENTS.md` and the template mirror, so the substituted `AGENTS.md.tmpl` is covered).

**AC-TU-024 (all requirements), class G.** *Given* the change set, *when* the set of `t.Skip`, `t.Skipf`, and `t.SkipNow` call lines in the changed test files is listed at `CARD_BASE` and at the tip with `git grep -n -E 't\.Skip(f|Now)?\(' -- <changed test files>`, *then* the two lists are equal; and no test selector or invocation raises a timeout value relative to the baseline.
- Not red now by design.

## §3 Edge cases

- A `moai gtd` invocation whose first word is a mistyped verb followed by a card id: the pre-run hook fires before the mistyped-verb refusal, so the warning precedes the refusal; AC-TU-001's table carries one row for this case (exactly one warning, then the refusal, exit status equal to the `todo` form's).
- `moai gtd` with no arguments renders the queue (the bare parent form): one warning, then the queue.
- `--help` on a subverb of `gtd`: no warning (AC-TU-002).
- A lane session running `done` with a card id in the `t<n>` form and in the bare `<n>` form: both resolve to the same own card (the existing normalisation).
- A card with landing evidence recorded by the operator (`todo landed --sha`) versus a subject on the landed ref: both are the closed forms; the lane path accepts either.
- The setter on a card that was unpicked and re-picked: the unpick cleared the queue spec_id, so the setter fills it again (AC-TU-021/023).
- The substitution script on a path that contains a space or begins with a dash: the script quotes every path (AC-TU-009 fixture row).

## §4 Quality gate criteria

- TRUST 5 Tested: each milestone's new tests RED then GREEN with the output recorded; the changed packages' coverage does not fall below the baseline measured at the milestone's start; name-exact selectors for `internal/cli`.
- TRUST 5 Readable and Unified: `gofmt` clean on changed Go files; the repository's golangci-lint configuration run over the changed packages; shell script passes `sh -n` and the repository's shell lint if one exists.
- TRUST 5 Secured: the lane exception never reads a card identity from an environment variable alone (the held-card record decides); the setter validates the identifier before any write; the script reads only `git ls-files` output.
- TRUST 5 Trackable: every commit message names card t1641; the sync record carries the AC matrix with RED-now and GREEN cells and the Gaps list.
- Windows build: `GOOS=windows GOARCH=amd64 go build ./...` exits 0 on the final tree.

## §5 Definition of Done

1. All 24 criteria adopted (class B criteria carry a recorded RED), and every class A and B criterion is GREEN with its verbatim output.
2. OQ-1 to OQ-9 each have an operator verdict or a recorded `DEFAULT-APPLIED` line in `decision-index.md`, and the criteria conditional on an answer (AC-TU-017, AC-TU-020, AC-TU-004) state which answer they were measured under.
3. `moai spec lint SPEC-TODO-UNIFY-001` exits 0 with no finding, using a `moai` built from the tip of this branch and invoked by path, with that build's commit recorded next to the tree HEAD.
4. The landing note lists the Gaps: the unrun remainder of `internal/cli`, the board mod's JavaScript tests, the docs-site build if the generator is unavailable, and any criterion left pending.
5. No `t.Skip*` added, no timeout raised, no hand edit of a file the substitution script owns.
