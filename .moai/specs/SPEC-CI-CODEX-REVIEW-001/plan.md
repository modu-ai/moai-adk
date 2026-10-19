# plan.md — SPEC-CI-CODEX-REVIEW-001

**Plan revision 11 (amendment 4, 2026-10-20; incorporating corrections through revision 10).** The native Codex review is an issue-comment state protocol, not a pull-request review object (`research.md` E37 to E46), and the
leader's policy adds an author, fork, and bot rule and forbids automatic branch deletion. The amendment adds one milestone, **M6** (§F), which changes the gate script,
`auto-merge.yml`, and their tests, and nothing else; the delta is in `design.md` "Delta of plan revision 7" (Δ12 to Δ26). Revision-7 items that bind the run phase are marked
"rev 7" below. The text of §A to §G that describes an APPROVED review object is superseded where it conflicts with `design.md` §3.1, §3.2, and `T-CHAIN`.

**Revision-11 correction to M6.** The committed upper contract requires each completed leg to report both `has_any_findings: false` and `has_unresolved_findings: false`. The gate now refuses a completed leg with `has_any_findings: true` before a running peer can defer the verdict; failed-leg flags remain unread, and running remains pending unless another refusal applies. No requirement or criterion identifiers are added. `TestCodexNative1828ReceiptTypesAndHeadBinding` covers code-only, security-only, and both-leg reported findings, and `TestCodexNativeRunningSecurityLegWaits` covers refusal before a running security wait.

**Plan revision 6 (amendment).** M1 to M4 are complete on the card branch (HEAD `0955fb9058`); the amendment
under the leader's rulings (board records d-20261018T152011Z-8447 (supersedes d-20261018T151824Z-db98; ruling and scope unchanged) and d-20261018T155527Z-4e9e (a delegated Codex observer run, not a new approval)) adds one
milestone, **M5** (§F), which changes only `auto-merge.yml` and the auto-merge tests. The delta is in
`design.md` "Delta of plan revision 6". Revision-6 items that bind the run phase are marked "rev 6" below.

Implementation plan, plan revision 4 (the native-review design; boundary ruling (b) kept; revision 4 changes
the review reader, the NOT RUN result, the changed-file binding, and adds the merge-decision criterion). Priority
labels and ordering only; no time estimates. Line counts are estimates of changed lines per milestone, used to keep each
milestone small. Milestones are ordered by decision-reversibility: the decisions most likely to change (what
the gate recognises as a review, how lists are read) come first; the wiring and the mechanical retirement and
contract guards come last.

## §A Context

- Card t1868, run tn1p4e, Class C, P1. Plan phase authored in tree
  `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1868`, branch `WT-ci-codex-review`. Card base `4f9dba14bf`
  (equal to local `main` when the card branched); revision 4 is measured at HEAD `e11af67632`, which differs
  from the base only under this SPEC directory and the card record (`git diff --name-only 4f9dba14bf HEAD`
  lists 10 paths, all under those two directories). The run-phase writer is lane-22 alone;
  landing is to local `main` only (no push, no pull request, no tag). Evidence path `.moai/reports/t1868/`.
- SPEC artifacts (Tier L, 5 files plus `decision-index.md` and `progress.md`): `spec.md`, `plan.md`,
  `acceptance.md`, `design.md`, `research.md`.
- Tier classification: **L**. Why: the change set is about 2300 changed lines across the four milestones, tests
  dominating (over 1300), and it replaces the only enforcement point for pull-request review, a
  security-sensitive surface. It touches 12 files: 7 new (the gate script and six test files), 3 edited
  (`auto-merge.yml`, `.coderabbit.yaml`, `liveness.yaml`), 2 deleted. The ceilings hold: 17 requirements
  (ceiling 25), 19 criteria (ceiling 25). Revision 7: 21 requirements, 25 criteria (both at or under the ceiling of 25; no criterion is added).
- Reviewer (design correction, operator-confirmed; **revision 7 replaces the review-object reading below with the state-comment reading of design §3 and research E37**): the ChatGPT subscription's native Codex GitHub review,
  automatic, on pull-request open, for modu-ai/moai-adk and modu-ai/mo.ai.kr; no key, no secret, no workflow
  of our own. The gate reads the state comment of the Codex app identity (E27, E37) with both legs completed and bound to the head and passes on
  nothing else (revision 7 withdraws the review-object reading of revisions 3 to 6). The form was read on real pull requests (E37 to E46); DS-2 and the
  effect of re-requests stay open; the gate fails closed, treats no state comment as NOT RUN and never as a pass, uses
  time only as a bound that withholds, and the deferred steps DS-1 to DS-3 observe the form.
- Boundary (leader ruling (b), 2026-10-18): workflows, the review-gate script and its tests, `.coderabbit.yaml`,
  the `required-checks.yml` guard (the file is never edited), and the liveness manifest lines with the tests
  that read them. Replacement scope: ONLY GitHub CodeRabbit and the GitHub Claude review/Actions; local Claude
  and GLM sessions and the functional CI are untouched. Doctrine, role-core, templates, `AGENTS.md`,
  `manager-lead`, `tech.md`, the binding ledger, the hook notice, the runbook, and every other `Merge Risk`
  wording are follow-up candidates only (`decision-index.md` § Follow-up candidates) and known debt KD-1
  (`spec.md` §7). No milestone below touches them.
- Lineage: this SPEC reverses two choices of the completed SPEC-V3R6-CODERABBIT-ADOPTION-001 under the operator
  order of 2026-10-18; sibling SPEC-CI-VERDICT-INTEGRITY-001 owns the required-checks contract this SPEC leaves
  unchanged.
- File-overlap result against the other open cards (research E16): none of the files this SPEC edits is in the
  stated scope of t1865 or t1632. t1731's stated scope can touch `ci.yml`, `release-pr-multi-os.yml`, and
  `required-checks.yml`, which this SPEC only guards unchanged (AC-CCR-016). Rule recorded by the dispatch:
  this card goes first and t1731 absorbs it on resume.
- Plan-phase facts live in `research.md` (E1-E32, U1-U10); decisions in `decision-index.md`; the gate contract,
  threat model, and remote observation path in `design.md`.

## §B Known issues the run phase must expect

1. **The liveness manifest is census-coupled to the workflow directory** (research E12, E19). The existing
   `TestCensus_SetDifferenceEmptyBothDirections` (`internal/guardstate/census_test.go:49-91`) fails on any
   file in `.github/workflows/*.yml|*.yaml` with no manifest entry, any entry naming a file that is not on
   disk, and any duplicate locator. This SPEC adds no workflow, so only the two retired entries are removed,
   in the same milestone and commit as the two deletions (M4). A commit that deletes the workflows without
   the manifest turns an existing test red.
2. **The doctrine is left unchanged** (KD-1). Existing doctrine tests (`TestBindingLedgerIntegrity`,
   `TestInjectedCoreCarriesObligationElements`, the mirror, budget, and section-reference tests) keep
   asserting the unedited text; the run phase never edits a doctrine, template, `AGENTS.md`, ledger, or hook
   file "to keep them consistent" — that is the follow-up's work, and an edit here is scope drift.
3. **`claude-code-review.yml` is remote-only** (research E5). The local tree cannot delete it; the census test
   AC-CCR-014 exists so that a merge bringing it back turns red. The deletion itself is gate G6 (the leader's
   draft pull request).
4. **bash 3.2.** The gate script runs under macOS bash 3.2 in the tests and bash 5 on ubuntu: no associative
   arrays, no `mapfile`, no `${var,,}`; lowercase with `tr`.
5. **Worktree guard.** The session guard refuses compound git forms, `git -C` across trees, `sh`, and command
   substitution inside git arguments. Use plain separate commands; record any refusal under Gaps in
   `progress.md`.
6. **Heavy-run slot.** Use name-exact `-run` selectors for every milestone. A whole-package run of
   `internal/template` takes the heavy-run slot first (`.claude/rules/moai/workflow/resource-slot-lease.md`);
   the plan phase took none.
7. **(rewritten in rev 7; items 18 and 19) The native review's artifact was read, not assumed** (E37 to E46): one state comment of the identity. Fixtures are inline and
   modelled on the real samples; no test or report may claim that the reader has run on a real pull request (it has not); the gate passes a pull request only on the
   comment contract of REQ-CCR-001, and no milestone may widen it (a review object, `completed` alone, or marker text in another author's comment never passes).
8. **(still open in rev 7) Whether the native review runs on a draft, or at the ready-for-review transition, is unconfirmed** (U4; the trigger "Draft marked ready" is measured, E38).
   Nothing in M1-M4 depends on it: a draft is NOT RUN at the gate whatever exists (AC-CCR-008).
9. **`workflow_run` runs the default-branch copy of `auto-merge.yml`**: the introducing pull request cannot
   exercise the new gate, and a test cannot take a live `sleep 15` loop. The step tests put a stub `gh` and a
   no-op `sleep` first on `PATH`.
10. **(rev 7: a review starts on open-for-review, draft-marked-ready, or an `@codex review` comment; DS-2, DS-3) The review runs only when a pull request opens** (operator-confirmed contract, per-push excluded): a moved
    head is `stale-review` by design. The run phase never requests a review, comments, or re-opens anything
    (spec §4); a fresh review is the leader-coordinated step DS-3.
11. **Changed-file entry binding rests on a measured field** (E31, U10): for non-removed entries the
    `contents_url` carries `?ref=<head sha>`; removed entries carry the base SHA and are not bound. The statuses
    `copied`, `changed`, and `unchanged` were not observed; the reader treats any status other than `removed` as
    needing the binding, so an unobserved status fails closed. If the pre-flight re-read (§C) finds a
    non-removed entry without the `ref`, that is a blocker, not a reason to drop the binding.
12. **The merge-decision criterion needs a small `if:` evaluator** (AC-CCR-019). It supports only
    `steps.<id>.outputs.<name> == '<literal>'` terms joined by `&&` and fails the test on any other shape; do
    not widen it to make a mutant pass.

13. **(rev 6) Step exit status and job result.** A step that ends non-zero without `continue-on-error` fails
    the job, and a later step whose `if` has no status function is skipped through the implied `success()`;
    the harness derives the job result by that rule and never claims what GitHub reports (U11). The `review`
    and `files` steps must not get `continue-on-error`, `always()`, `failure()`, or `cancelled()`.
14. **(rev 6) One value per output key.** Write `ok` and `state` once, on the terminating path; do not write
    a default first. A single terminating function that appends both keys and exits with the given status is
    the shape (design `T-OUTPUTS`).
15. **(rev 6) Replace, do not keep, the two exit-0 rows.** `not_run_step_exit_status_0_with_notice_annotation`
    and `fail_step_exit_status_0_with_warning_annotation` assert the pre-ruling behavior (research E35) and are
    replaced by AC-CCR-020 to AC-CCR-022; the `not_run_message_neither_failure_nor_pass` row becomes
    `not_run_message_never_a_pass`; the `pending_keeps_waiting` and `fetch_failure_counts_pending` call counts
    (moved by ruling B) stay as they are.
16. **(rev 6) The CHANGELOG entry is the sync phase's.** Its sentence "reported as NOT RUN, which is neither a
    failure nor a pass, and the merge is withheld on that state as on a refusal" goes stale at M5; this plan
    does not edit `CHANGELOG.md`, and the sync phase of the amendment owes the correction.
17. **(rev 6) Do not widen.** The steps `pr`, `head`, `checks`, and the merge step keep their exit-0 convention
    (decision-index Q22); the draft's job result is "failure, never success" with one job (Q21); the gate
    script is untouched. A need to change any of these returns as a blocker.
18. **(rev 7) The reader decides from the state comment.** `body` is read only for comments of the identity and only the marker lines of the leading block; never the prose, the
    table, the summary marker, or a reaction. Stale precedes pending in the reader order (design §3.2, step 7 before step 10); changing the order changes the contract.
19. **(rev 7) Fixtures are inline and modelled on the real samples** (design "Real-sample provenance"): author fields from the real objects (login `chatgpt-codex-connector[bot]`, id
    199175422, type Bot), marker lines and table rows only; the lane's sample files are machine-local and gitignored, so a test carries its fixture inline and never reads them.
20. **(rev 7) The author, fork, and bot rule and the removal of `--delete-branch` are enforced by the workflow and pinned by tests** (REQ-CCR-020, REQ-CCR-021; AC-CCR-006, 012, 019). A
    sentence in a document that states the rule does not complete the card. The repository setting `delete_branch_on_merge` (G9) is an operator decision item and is not touched.
21. **(rev 7) On an unprotected base the shipped `checks` step withholds** (E43, E45; decision-index Q27). Do not change the step to make a pull request pass; record the consequence
    in the run evidence. The review gate, the `author` step, and the merge path are written and tested for every base (AC-CCR-019 base rows).
22. **(rev 7) `eligible` is one value, recorded once** (REQ-CCR-019 by analogy), and the `author` step ends with exit 0 in every case (Q26); the four later `if` carry the term.
23. **(rev 7) The sync phase of this amendment owes the CHANGELOG correction.** The entry of the publication tree describes the shipped (revision-6) reader; this plan did not re-read it
    for the correction and does not edit `CHANGELOG.md`.
24. **(rev 7) A moved head, a re-request, and mixed SHAs.** Every leg's `commit_sha` and the security marker's `headSha` must equal the head; the state of #1842 after its re-request is
    the fixture (`F-mixed-sha`). The merge step re-reads the head immediately before the call.

## §C Pre-flight (run once, before M1)

```
git rev-parse --short HEAD ; git branch --show-current          # expect the card branch, clean tree
git status --short                                               # expect empty
go build ./...                                                   # baseline compiles
test -x "$(command -v actionlint)" ; test -x "$(command -v jq)" ; command -v bash
gh api "users/chatgpt-codex-connector%5Bbot%5D"                  # re-read: id 199175422, type Bot (E27)
gh api "repos/modu-ai/moai-adk/pulls/1823/reviews"               # re-read the review fields: commit_id, state, submitted_at, user.id, user.type (E23)
gh api "repos/modu-ai/moai-adk/pulls/1823/comments"              # re-read the review comment fields: commit_id, original_commit_id, user.* (E23)
gh api "repos/modu-ai/moai-adk/pulls/1823/files"                 # re-read status and contents_url: non-removed entries carry ?ref=<head sha> (E31)
```

**(rev 7) Pre-flight for M6.** Re-read, by `gh api` GET: one issue comment of the identity on a pull request that has one (for example `repos/modu-ai/moai-adk/issues/1841/comments`): the author's login, id
199175422, type Bot, and the leading block with the three marker lines (E37); and the pull request object of a pull request of the eligible author: `user.login` `GoosLab`, `user.id` 6458969,
`user.type` `User`, and `head.repo.id` against `base.repo.id` (E43). If a re-read differs, that is a blocker.

If a `gh api` GET is refused or unavailable, stop with a blocker report: the recognition constants must not
carry an identity that was not re-read, and E27 and E23 are 2026-10-18 readings. If the bot id or type differs,
or a named field is absent from the re-read objects, or a non-removed changed-file entry carries no `ref` in
its `contents_url`, that is a blocker too. Baseline-red check: run these
existing tests once, name-exact, and record any pre-existing red as baseline so it is never counted as new:
`TestAutoMergeRequiredChecks` (`internal/template`); the three `TestCensus_*` tests (`internal/guardstate`);
`TestNoHardcodedContexts` (`internal/config`); `TestRequiredChecks_ReleaseGateSubsetOfMain` is run only where the ReleaseGate field exists on the tree; it is absent from the remote base, where the guard reads the optional release_gate YAML key directly through the helper readReleaseGate in internal/config/required_checks_review_guard_test.go and the rows not_in_release_gate and control_release_gate_reader_reads_the_key carry the same forbidden condition.

## §D Constraints (do not violate)

- Forbidden for the run phase, whatever a milestone seems to need: any push, pull request, tag, workflow
  dispatch, comment, review request, reopen or ready-for-review transition, secret or variable change,
  branch-protection call, ruleset change, app install or settings change; any edit of
  `.github/required-checks.yml`, `.github/branch-protection.json.gtmpl`, its template mirror, `ci.yml`,
  `codeql.yml`, `pr-multi-os-gate.yml`, `release-pr-multi-os.yml`; any key or secret. New keys, new app OAuth,
  persistent access, admin or security settings, and cost are not approved (gates G1-G8, deferred steps
  DS-1 to DS-3); local Claude and GLM sessions are unchanged.
- No workflow file is added (AC-CCR-013); the only workflow edits are `auto-merge.yml` (edited) and the two
  deletions.
- PRESERVE (byte-unchanged, AC-CCR-016): the seven paths of that criterion.
- Not this card's files: anything t1865, t1632, or t1731 own; every non-test `internal/cli` source file; the
  `internal/cli/testdata/codex-rollouts-t1171/` fixtures; and every doctrine, role-core, template,
  `AGENTS.md`, `manager-lead`, ledger, hook, `tech.md`, and runbook file named in the follow-up table.
- No non-test Go source file and no template file changes; the only Go files added are `_test.go`.
- No test skip, no timeout increase, no loosened assertion, no `|| true` over a verification. A missing `bash`
  or `jq` fails the gate tests.
- Stage by explicit pathspec only; re-read `git rev-parse --short HEAD` and `git branch --show-current`
  immediately before every commit. Every commit message carries `t1868` (traceability carrier) and ends with
  the `🗿 MoAI` trailer.
- Subagent boundary: no `AskUserQuestion`; a blocker returns as a Missing Inputs report.
- State wording: the run and sync reports say "landed locally" with the merged-local SHA and "remote verified:
  not reached by this lane"; they never say "complete", and never claim remote verification or the draft
  observation (the remote data for a pull request exists only after the leader's coordination).

## §E Self-verification protocol

Each milestone's report follows the five-section evidence format (Claim, Evidence, Baseline-attribution,
Gaps, Residual-risk) and carries:

- **RED first, verbatim.** The milestone's tests are written and run against the unchanged tree first; the
  failing output is quoted. A `[no tests to run]` result is not a RED. Tests that fail because a file they
  read is absent count only when the failure message names the absent file (the stated reason).
- **GREEN with the swept set.** The `-v` output lists every named `--- PASS:` line of the milestone's
  criteria; the count of named lines is stated.
- **Mutant protocol.** After the milestone's GREEN commit, apply each named mutant of its criteria to the
  committed file in the card worktree, run the criterion's test, quote the verbatim red, then restore with
  `git checkout -- <path>` and confirm `git status --short` is empty. For the four regression-guards the red
  is the adoption evidence.
- **Existing tests re-measured** (acceptance §5), name-exact, with the swept count stated.
- `go vet` and `gofmt -l` on the touched test files; `actionlint` on `auto-merge.yml` at M3.
- Progress record: `.moai/specs/SPEC-CI-CODEX-REVIEW-001/progress.md` §E.2 holds the run evidence;
  `.moai/reports/t1868/progress.md` holds the card record. The run phase never edits the plan artifacts; a
  needed change returns as a blocker for manager-spec.

## §F Milestones

Notation: files are new (N), edited (E), or deleted (D). "Re-measure" names the tests run at the milestone's
close; the landing re-measure is the union in § Landing scope.

### M1 — Native review reading (Priority High; landed with the revision-6 reader, replaced by M6 for the comment contract)

REQ/AC: REQ-CCR-001..005, REQ-CCR-008 (script half) / AC-CCR-001..005, AC-CCR-007 (script rows), AC-CCR-008
(script rows). Files (2, about 620 lines): N `scripts/ci/codex-review-gate.sh` (`read-native-review` only,
carrying the recognition constants of design §3.1 and the nine-step order and decision table of §3.2, with
`--wait-closed`), N `internal/template/codex_review_gate_test.go` (seven tests, subtests as acceptance.md
names them). RED first: write the seven tests; run name-exact against the absent script; quote the failures
naming the absent script (R1). GREEN: implement per design §3.1, §3.2. Exit: AC-CCR-001..005 swept; the
script rows of AC-CCR-007 and AC-CCR-008 swept; mutants recorded — including the revision-3 reader mutant of
AC-CCR-004 (newest decides, COMMENTED counts) and the body-text mutants; the four function-check conditions of
acceptance §3.1 each red under their mutant. Affected packages: `internal/template` (tests only). Re-measure:
the seven `^Test...$` selectors of `internal/template`.

### M2 — Complete paging and the tamper guard (Priority High)

REQ/AC: REQ-CCR-006 (script rows), REQ-CCR-009, REQ-CCR-010 / AC-CCR-006 (script rows), AC-CCR-009,
AC-CCR-010. Files (2, about 560): E `scripts/ci/codex-review-gate.sh` (adds `fetch-pages` with the endpoint
allow-list and the repeated-entry refusal, and `gate-files-changed --head-sha` with the entry binding), N
`internal/template/codex_review_paging_test.go` (`TestCodexGateFetchPagesCompleteness`,
`TestCodexGateFetchPagesRefusesDisallowedEndpoint`, `TestCodexGateFilesChangedRefusals`; the stub `gh` serves
pages and can fail one). RED first against the M1 script lacking the subcommands (the failure names the
unknown subcommand). Exit: swept; mutants recorded — including the first-page-only mutant on the 101-entry and
301-entry fixtures and the "list read by pull request number" mutant on the entry-binding fixtures. Affected
packages: `internal/template` (tests only). Re-measure: the three selectors plus M1's seven (the script is
shared).

### M3 — Auto-merge wiring (Priority High)

REQ/AC: REQ-CCR-006 (workflow rows), REQ-CCR-007, REQ-CCR-008 (step half), REQ-CCR-011, REQ-CCR-012,
REQ-CCR-013 / AC-CCR-006 (workflow rows), AC-CCR-007 (loop rows), AC-CCR-008 (step rows), AC-CCR-011,
AC-CCR-012, AC-CCR-013, AC-CCR-019. Files (2, about 1000 changed): E `.github/workflows/auto-merge.yml` (the
CodeRabbit step is deleted; the `files` step and the `review` step of design §5 are added, with the `state`
output, the NOT RUN notice, and the head reads; the merge step's `if` follows; permissions stay as they are), N
`internal/template/auto_merge_codex_test.go` (seven tests; stub `gh` that records calls, stub script, no-op
`sleep`, the restricted `if:` evaluator). RED first: the tests against the unchanged workflow (R2, R3, R4, R5,
R6, R12 are the stated reasons). Exit: swept; `TestAutoMergeRequiredChecks` still lists its 16 `--- PASS` rows;
`actionlint` exit 0; mutants recorded, including the AC-CCR-011 and AC-CCR-013 regression-guard reds, every
AC-CCR-012 mutant (deleted `files` step, ignored exit code, merge before the gate, a `review` step with no
head read after its list reads, merge `if` without the review term, merge `if` without the `files` term as
the structural pin), and every AC-CCR-019 mutant (`&& false` in the merge `if`, an early `exit 0`, a
hard-coded deadline, mutant MF (the `files` term dropped from the `review` step and the
merge `if`, both led with `always()`, traced in `acceptance.md` §2.5), a dropped `--match-head-commit`), each with its verbatim red; the dropped review term of the merge `if` is an
equivalent mutant covered structurally by AC-CCR-012 and has no AC-CCR-019 red. Affected
packages: `internal/template` (tests only). Re-measure: the seven new selectors plus
`^TestAutoMergeRequiredChecks$`.

### M4 — Retirement and contract guard (Priority High)

REQ/AC: REQ-CCR-014..017 / AC-CCR-014..018. Files (6, about 380 changed counting deletions): D
`.github/workflows/claude.yml`, D `.github/workflows/review-quality-gate.yml`, E `.coderabbit.yaml` (two
flags), E `.moai/guards/liveness.yaml` (remove two entries), N `internal/guardstate/census_review_test.go`
(`TestCensusOmitsRetiredReviewWorkflows`), N `internal/template/review_retirement_test.go`
(`TestNoClaudeReviewWorkflows`, `TestCodeRabbitDisabledAndUnread`), N
`internal/config/required_checks_review_guard_test.go`. The two deletions and the two manifest removals are
one commit (§B.1). RED first (R7, R8, R10, R11 are the stated reasons). Exit: swept; the AC-CCR-016 command exit
0 with its control; the regression-guard reds (AC-CCR-016 observed red, AC-CCR-017 mutant) recorded; the three
census tests pass. State at exit: landed locally; remote verified is not reached by this lane. Affected
packages: `internal/template`, `internal/config`, `internal/guardstate` (tests only; `required_checks*.go`
unchanged). Re-measure: the new selectors plus the three census tests, `^TestNoHardcodedContexts$`, and
`^TestRequiredChecks_ReleaseGateSubsetOfMain$` (only where the ReleaseGate field exists on the tree; it is absent from the remote base, where the guard reads the optional release_gate YAML key directly through the helper readReleaseGate in internal/config/required_checks_review_guard_test.go and the rows not_in_release_gate and control_release_gate_reader_reads_the_key carry the same forbidden condition).

### M5 — Step and job conclusion (rev 6, Priority High)

REQ/AC: REQ-CCR-007, REQ-CCR-008, REQ-CCR-012 (reworded), REQ-CCR-018, REQ-CCR-019 / AC-CCR-020 to AC-CCR-025
new; AC-CCR-007, 008, 012, 019 changed rows. Files (2, about 700 changed lines): E `.github/workflows/auto-merge.yml`
(the `files` and `review` steps: a terminating function that records `ok` and `state` once and exits with the
status of `T-CHAIN`; no default first; `::error::` annotations for failures; draft exit 11 with `state=not-run:draft`;
missing script a failure), E `internal/template/auto_merge_codex_test.go` (the step harness records exit status,
outputs, and the derived job result; five new tests named in `acceptance.md` §2.6; the two exit-0 rows replaced;
chain fixtures of AC-CCR-019 extended). The gate script, its tests, and every other file are untouched.
RED first: write the five new tests and the changed rows; run name-exact against the unchanged workflow and quote
the failures, which must be on the exit status or the job result (R13 to R17 are the stated reasons); a RED that
fails only because a helper is missing does not count. GREEN: edit the two steps. Exit: every named subtest of
AC-CCR-020 to AC-CCR-025 and of the changed rows swept; `TestAutoMergeRequiredChecks` still lists its 16 rows;
`actionlint` exit 0; mutants recorded with verbatim reds: the committed (exit-0) steps, `continue-on-error: true`
on each step, a duplicate-key pre-write, a comparison on outputs only, a draft mapped to `ok=true`, a clean pass
that exits 1, and the gate script missing read as a pass. Affected packages: `internal/template` (tests only).
Re-measure: the twelve auto-merge and paging selectors of the ruling-B round, plus the five new ones. State at
exit: landed locally is claimed only after the merge; the remote conclusion of the job is UNOBSERVED and the
remote state "unverified" (spec §4, §5).

### M6 — The comment contract and the author policy (rev 7, Priority High)

REQ/AC: REQ-CCR-001 to 008, 010, 011, 012 (reworded), REQ-CCR-020, REQ-CCR-021 (new) / AC-CCR-001 to 008, 010, 012, 019 to 023 rewritten and the new rows of AC-CCR-006, 012, and 019; no criterion
is added (25 criteria). Files (5, about 1500 changed lines, tests dominating): E `scripts/ci/codex-review-gate.sh` (the reader of design §3.1 and §3.2 with `--issue-comments`, the
`issues/N/comments` endpoint in `fetch-pages`), E `.github/workflows/auto-merge.yml` (the `author` step and the `eligible` term in the `if` of `checks`, `files`, `review`, and the merge step; the issue-comment collection with
`--expect-count comments`; the `running` note; the recheck of the same comment; the issue-comment-moved retry; the merge call without `--delete-branch`; the terminating function and `T-CHAIN`
unchanged), E `internal/template/codex_review_gate_test.go`, E `internal/template/codex_review_paging_test.go`, E `internal/template/auto_merge_codex_test.go` (the real-sample fixtures of the provenance
table; the base matrix; the author, fork, bot, lookalike, and branch-deletion rows). RED first: write the rewritten rows against the tree at `b2af3db9a1` and quote the failures, whose stated reasons are
R18 to R24, R26, and R27 (R25 is a test-addition cell outside the class claim; cells R1 to R17 no longer reproduce and bind nothing; a RED that fails only because a helper is missing does not count). GREEN: edit the gate script and the workflow. Exit: every named subtest of the criteria above swept (`-v` lists each
`--- PASS`); `TestAutoMergeRequiredChecks` still lists its 16 rows; `actionlint` exit 0; mutants recorded with verbatim reds: the revision-6 reader, a status-alone pass, null read as false, the
code leg alone, the newest of two comments, a whole-body marker scan, the security marker ignored, a recheck on the head only, a merge `if` naming a base, a merge step that does not re-read the head, a
state recorded for the old head counted, a recheck that does not compare the issue-comment count, a reader that reads the flags of a `failed` leg as findings or waits on it, a running comment with no security marker refused at once instead of waited on, an `author` step matching the login only, no head-repository comparison, `eligible=true` for an unreadable object, a merge call with `--delete-branch`, and
the MF-shaped mutant for the `eligible` term (the lone drop from the merge `if` is an equivalent mutant pinned structurally). The single writer of this milestone is the card's lane; this plan
revision writes none of it. Affected packages: `internal/template` (tests only); no non-test Go file and no template file changes. Re-measure: the seven reader selectors, the three paging selectors,
and the twelve auto-merge selectors of M1 to M5. State at exit: landed locally is claimed only after the merge; the real run of the revision-7 reader on a pull request, DS-2, the `running` to `completed`
transition, and whether `CI` fires for a pull request into a `release/*` base stay UNOBSERVED, and the remote state of the amended workflow is "unverified".

### Landing scope (plan §F rule: changed packages plus direct importers plus readers of changed non-Go files)

| Group | Packages | How |
|-------|----------|-----|
| Changed packages (tests only) | `internal/template`, `internal/config`, `internal/guardstate` | Name-exact selectors of every milestone, plus the whole `internal/template` package once at landing under the heavy-run slot |
| Direct importers | none to add: no non-test Go source file, no template file, and no catalog hash changes | `go build ./...` and `go vet` on the three packages; `git status --short` empty after `make build` (nothing templated changed) |
| Existing tests that read changed non-Go files | `internal/guardstate` (the three `TestCensus_*`: read `liveness.yaml` and the workflow directory), `internal/template` (`TestAutoMergeRequiredChecks`, `TestBranchProtectionParity`), `internal/config` (`required_checks_test.go`, `required_checks_release_gate_test.go`) | Name-exact selection (standing rule 35e9); record the selector list and swept counts in `progress.md` §E.2 |

Landing evidence also includes: AC-CCR-016's merge-base command and its control; `git status --short` empty
after `make build`; the list of changed files showing no path outside this SPEC's 12-file set. The final run
report says "landed locally" with the merged-local SHA on local `main` and "remote verified: not reached by this
lane" (acceptance §6, §7). The sync exit says the same two things in the same words.

## §G Anti-patterns

- Reading a draft as anything but NOT RUN, or a draft's gate-path observation as a review result (AC-CCR-008); reading a missing state comment as a pass, as a failure, or as
  pending forever instead of NOT RUN `review-absent` at the end of the window (AC-CCR-007, AC-CCR-008); reading a check run, a commit status, a reaction, a review object,
  or prose or a table row of any comment as a verdict, or a marker string in another author's comment as a state (AC-CCR-003, AC-CCR-006). The one verdict carrier is the
  state marker of the identity's single issue comment.
- Counting a review object (APPROVED included) as a pass, or letting a review object or review comment do anything but refuse (AC-CCR-001, AC-CCR-005); choosing among state
  comments by recency (AC-CCR-004); reading the body of any artifact but the identity's comments (AC-CCR-001); letting elapsed time produce a pass (AC-CCR-007, AC-CCR-019).
- Judging a changed-file list that is not bound to the verified head, or reading the head only once
  (AC-CCR-009, AC-CCR-012); a merge step that no criterion executes (AC-CCR-019).
- Matching the Codex identity by login alone or by id alone (AC-CCR-003); comparing the review's `commit_id` by
  prefix or not at all (AC-CCR-002).
- Reading page 1 of a list and deciding; ignoring the exit status of `gh`; skipping the count proof or the
  ceiling (AC-CCR-010).
- Calling the guard or the read and discarding its exit code, or omitting either from the merge condition
  (AC-CCR-012).
- Adding a workflow, a secret reference, a `pull_request_target` trigger, or an artifact or cache step
  (AC-CCR-013).
- Deleting the two workflows in one commit and removing their manifest entries in another: an existing census
  test is red in between (§B.1).
- Passing `[no tests to run]` as a pass; naming fewer subtests than the criterion lists.
- Editing `required-checks.yml` to "make a review visible" (AC-CCR-016, AC-CCR-017, gate G4).
- Editing doctrine, templates, `AGENTS.md`, the ledger, or the lane notice "while the files are open" (spec §6
  and §7; the follow-up table owns them).
- Requesting, commenting, re-opening, or changing any app setting to "get a review" (spec §4; DS-3 is the
  leader's).
- (rev 6) Asserting only the output file of a step and calling the failure covered; ending a failed gate with
  exit 0 and an annotation; a `continue-on-error` or a status-function override on a guard step; writing a
  default output first and appending the terminal value (AC-CCR-020 to AC-CCR-025).
- (rev 6) Reporting the remote conclusion of the job, or "active" or "passing" for the amended workflow, from the
  G6 draft pull request's CI (spec §5, G6 acceptance condition).
- (rev 7) Reading `completed` alone, or both legs reading `completed`, as a pass; reading null as false; letting a state recorded for another head count; merging while a review is `running` or
  absent, on any base (AC-CCR-002, 004, 007, 019).
- (rev 9) Reading the finding flags of a `failed` leg as findings, treating `failed` as pending, or letting a failed leg pass beside a completed one (AC-CCR-004, AC-CCR-005, AC-CCR-019).
- (rev 11) Treating a completed leg with `has_any_findings: true` and `has_unresolved_findings: false` as clean, or letting a running peer defer that completed finding (REQ-CCR-001, REQ-CCR-005, AC-CCR-005).
- (rev 7) Documenting the author, fork, and bot rule or the no-branch-deletion rule without the workflow enforcing it and the tests pinning it; touching `delete_branch_on_merge` (REQ-CCR-020,
  REQ-CCR-021, G9).
- Writing "complete" for a local landing, or claiming remote verification without the remote data of that SHA
  (§D State wording).
- Deleting `.coderabbit.yaml` instead of disabling its two flags (AC-CCR-015).

## §H Cross-references

- `spec.md` §3 (requirements), §5 (gates and deferred steps), §6 (exclusions), §7 (known debt, accepted
  limitations); `acceptance.md` (criteria, ledger, matrix, states); `design.md` (contract, threat model, test
  layout, observation path); `research.md` (evidence and unverified premises); `decision-index.md` (decisions,
  follow-up candidates, deferred gates).
- `.claude/rules/moai/development/verification-completeness.md` §1.1-§2.1 (RED-now, mutant, swept set);
  `.claude/rules/moai/core/verification-claim-integrity.md` §1-§3.
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier; `auto-semantics.md` §9.1 (the Kickoff
  transition this plan feeds).
- `.github/workflows/auto-merge.yml`, `internal/template/auto_merge_checks_test.go`,
  `internal/guardstate/census_test.go` (the existing patterns this SPEC follows).
