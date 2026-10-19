# acceptance.md — SPEC-CI-CODEX-REVIEW-001

Verification layer. Requirements live in `spec.md` §3 (GEARS); every entry here is a Given-When-Then
criterion, binary-testable, labelled `AC-CCR-NNN`. Evidence ids R* are the RED-now ledger below; E* and U*
are in `research.md`. Plan revision 4 keeps the criterion set of revision 3 and adds one criterion (the merge
decision path): 19 criteria for 17 requirements, contiguous. **Plan revision 6** (an in-place amendment, spec
`## Amendments`; delta in `design.md`) changes four criteria (AC-CCR-007, 008, 012, 019) and adds six
(AC-CCR-020 to AC-CCR-025): 25 criteria for 19 requirements, contiguous, at the ceiling of 25. Revision 6
criteria carry their own RED-now pin, tree `0955fb9058` (ledger R13 to R17). The iteration-1 identifiers map to the new ones
in `.moai/reports/t1868/progress.md`; the keyed-action criteria (key presence, workflow triggers, pins,
verdict job, fork notice job, check-run binding) no longer exist because no such workflow is built.

**Document pin.** Every RED-now observation below was re-measured on tree `e11af67632` (full
`e11af676320604bb1f7cf07f14aa457b3437a9ec`, branch `WT-ci-codex-review`, clean) on 2026-10-18 and binds every
criterion that carries no pin of its own. That tree differs from the card base `4f9dba14bf` only under
`.moai/specs/SPEC-CI-CODEX-REVIEW-001/` and `.moai/reports/t1868/` (`git diff --name-only 4f9dba14bf HEAD`
lists 10 paths, all under those two directories). The preserve proof of AC-CCR-016 is measured from the merge
base of `HEAD` and local `main`, read at measurement time (on this tree it prints
`4f9dba14bf2834aa08bea79f9e72c2a7e75ab18e`). A criterion-level pin, where present, wins.

**Classes.** *release-blocking* (21 criteria: the 15 of revision 4 and AC-CCR-020 to AC-CCR-025): its RED-now cell is a re-executable single-invocation command
with stdout, exit code, and the tree pin, and a green path names the milestone that flips it.
*regression-guard* (4 criteria: AC-CCR-011, AC-CCR-013, AC-CCR-016, AC-CCR-017): green on the baseline by
nature, so no RED-now exists; each names a mutant that makes it red, and the milestone that adds it must
record that red verbatim. A regression-guard is never recorded as a release gate.

**Swept-set rule.** Every Go-test criterion names its subtests. A run counts only when `-v` output lists
each named `--- PASS:` line; a green with `[no tests to run]` or fewer named lines is a Gap, never a pass.
The selectors are name-exact (`^...$`).

**Fixture shape.** The review and review-comment fixtures follow the documented REST objects (E23): a review
carries `id`, `user.{login,id,type}`, `commit_id`, `state`, `submitted_at`; a review comment carries `id`,
`user.*`, `commit_id`, `original_commit_id`; a changed-file entry carries `filename`, `status`, and
`contents_url` with `?ref=<commit>` (E31). "The identity" below is login `chatgpt-codex-connector[bot]`, user
id `199175422`, type `Bot` (E27). A fixture of the identity is the only artifact a test can make count. In the
fixture names, `H` is the head under review and `O` another commit; `(id,commit)` names a review.

## 1. RED-now ledger (tree e11af67632)

```
R1  $ test -f scripts/ci/codex-review-gate.sh
    stdout: (empty)                                   exit: 1
R2  $ grep -c -i coderabbit .github/workflows/auto-merge.yml
    stdout: 14                                        exit: 0
R3  $ grep -c -F 'issues/$PR_NUMBER/comments' .github/workflows/auto-merge.yml
    stdout: 1                                         exit: 0
R4  $ grep -c 'gate-files-changed' .github/workflows/auto-merge.yml
    stdout: 0                                         exit: 1
R5  $ grep -c 'read-native-review' .github/workflows/auto-merge.yml
    stdout: 0                                         exit: 1   (the native review read is not wired)
R6  $ grep -c 'fetch-pages' .github/workflows/auto-merge.yml
    stdout: 0                                         exit: 1   (no paginated collection is wired)
R7  $ grep -rn 'anthropics/claude-code-action' .github/workflows
    stdout: .github/workflows/claude.yml:35:        uses: anthropics/claude-code-action@v1
            .github/workflows/claude.yml:47:          # See https://github.com/anthropics/claude-code-action/blob/main/docs/usage.md
                                                      exit: 0
R8  $ grep -n 'enabled:' .coderabbit.yaml
    stdout: 65:    enabled: true                      exit: 0
    $ grep -n 'auto_reply' .coderabbit.yaml
    stdout: 74:  auto_reply: true                     exit: 0
R9  $ git diff --quiet 4f9dba14bf 6dde1c1c1b -- .github/required-checks.yml
    stdout: (empty)                                   exit: 1   (6dde1c1c1b is the origin/main tip read at
                                                                 this tree; a tree known to differ: the probe can fail)
    $ git diff --quiet 4f9dba14bf HEAD -- .github/required-checks.yml
    stdout: (empty)                                   exit: 0   (the baseline: the probe is green)
R10 $ test -f .github/workflows/claude.yml
    stdout: (empty)                                   exit: 0   (present)
    $ test -f .github/workflows/review-quality-gate.yml
    stdout: (empty)                                   exit: 0   (present)
R11 $ grep -c -e 'claude.yml' -e 'review-quality-gate.yml' .moai/guards/liveness.yaml
    stdout: 2                                         exit: 0   (both retired workflows still listed)
R12 $ grep -c -F 'steps.review.outputs.ok' .github/workflows/auto-merge.yml
    stdout: 0                                         exit: 1   (the merge condition has no review term today:
                                                                 it reads steps.coderabbit.outputs.ok)
```

**Revision 6 ledger (tree `0955fb9058`, full `0955fb90585444f4a9679afe82d994d3f1b1a80b`, clean).** Each
command is a single read-only invocation; the criteria below cite these cells and the cell's pin wins over the
document pin.

```
R13 $ grep -n -B1 -A1 -F 'native review gate: $1 - withholding merge' .github/workflows/auto-merge.yml
    stdout: 263-          withhold() {
            264:            echo "::warning::native review gate: $1 - withholding merge"
            265-            exit 0
    exit: 0   (the review step ends a refusal with exit 0)
R14 $ sed -n '260,261p;353p' .github/workflows/auto-merge.yml
    stdout:           echo "ok=false" >> "$GITHUB_OUTPUT"
              echo "state=fail" >> "$GITHUB_OUTPUT"
                    echo "state=not-run" >> "$GITHUB_OUTPUT"
    exit: 0   (defaults written first, then a second `state`)
R15 $ grep -n -B1 -A1 -F 'changed-file check: $1 - withholding merge' .github/workflows/auto-merge.yml
    stdout: 194-          withhold() {
            195:            echo "::warning::changed-file check: $1 - withholding merge"
            196-            exit 0
    exit: 0   (the files step ends a refusal, gate-script-missing included, with exit 0)
R16 $ grep -c 'TestAutoMergeReviewFailureFailsStepAndJob' internal/template/auto_merge_codex_test.go
    stdout: 0
    exit: 1   (the new tests are absent)
R17 $ grep -c -F 'not_run_step_exit_status_0_with_notice_annotation' internal/template/auto_merge_codex_test.go
    stdout: 1
    exit: 0   (a committed row asserts the pre-ruling exit status 0)
```

Supporting behavior, not a cell (the delegated Codex observer's reproduction `/tmp/adk-queue-repair-20261018/review-step-repro-v2/`,
the observer's `/tmp/adk-ci-followthrough-20261018/files-step-missing-script-repro/result.json`, and this lane's
re-run in its scratchpad, all machine-local; the deciding lines are carried in `research.md` E34 and E36):
changes-requested `exit 0`, outputs `ok=false state=fail`; draft `exit 0`, outputs `ok=false state=fail state=not-run`;
clean `exit 0`, outputs `ok=false state=fail ok=true state=pass`; `files` step without the script `exit 0`, outputs
`ok=false state=fail`.

R1 to R11 were re-run on `e11af67632` in revision 4 and print exactly what they printed on `ae494aedb2`
(revision 3); R12 is new in revision 4.

## 2. Criteria

### 2.1 Native review verification (M1)

**AC-CCR-001** — REQ-CCR-001 — release-blocking
- Given reviews and review comments in which the identity's newest review bound to the head commit is
  APPROVED with a submission time, no review comment of the identity is bound to the head, and every earlier
  bound CHANGES_REQUESTED has a later bound APPROVED,
- When `read-native-review` runs with `--draft false`,
- Then it exits 0 and prints `native-review-clean`; noise does not change that: another bot's review, a
  human's comment, another login's CHANGES_REQUESTED on the head, a stale review of the identity for an older
  head, a stale finding bound to an older head, and the body text of the APPROVED review, whatever it says
  (the body is never read).
- Verify: `go test -count=1 -v -run '^TestCodexGateReadNativeReviewAccepts$' ./internal/template/` lists
  `--- PASS` for `approved_bound_clean`, `uppercase_sha_bound_clean`, `other_authors_noise_ignored`,
  `stale_review_then_current_approved_clean`, `stale_finding_older_head_ignored`,
  `latest_bound_review_decides_approved`, `changes_requested_then_approved_clean`,
  `dismissed_then_approved_clean`, `approved_body_text_never_read`, `script_never_references_body_field`
  (the last reads the script text: `.body` occurs in no jq filter).
- Mutant: a script that exits 0 unconditionally passes this criterion and fails AC-CCR-002 to AC-CCR-005; a
  script that always refuses fails this one. The pair makes both directions observable. A script that treats
  any bound CHANGES_REQUESTED as final, with no supersession, turns `changes_requested_then_approved_clean`
  red; one that reads body text to demote an APPROVED review turns `approved_body_text_never_read` red; any
  reference to the `body` field turns `script_never_references_body_field` red.
- RED-now: R1 — the script is absent, so the positive control cannot run. Green path: M1 flips it.

**AC-CCR-002** — REQ-CCR-002 — release-blocking
- Given reviews by the identity that are all bound to a commit other than the head under review (an older
  head, the head moved after the pull request opened, a prefix of the head, an empty `commit_id`), including an
  APPROVED review of an older head,
- When `read-native-review` runs, with `--wait-closed false` and with `--wait-closed true`,
- Then it exits 1 with reason `stale-review:<commit prefix>`, never 0, never pending, and never NOT RUN.
- Verify: `go test -count=1 -v -run '^TestCodexGateReadNativeReviewRefusesStale$' ./internal/template/` lists
  `--- PASS` for `approved_review_for_older_head`, `review_commit_prefix_only`, `review_commit_empty_string`,
  `review_for_other_commit_after_push`, `forged_bound_review_beside_real_stale_review_still_stale`,
  `stale_review_window_closed_still_exit_1`.
- Mutant: a script that ignores `commit_id` (any review of the identity passes), compares a prefix, treats a
  stale review as pending, or maps a stale review to NOT RUN when the window is closed turns the matching row
  red.
- RED-now: R1. Green path: M1.

**AC-CCR-003** — REQ-CCR-003 — release-blocking
- Given artifacts that look like a review but are not authored by the identity: a bound review in any state
  by another bot (`coderabbitai[bot]`, id 136622811), a human's review whose body says the change is approved,
  the same login without the `[bot]` suffix and the same id, the same login with another numeric id, the same
  login on a User account, and a review whose `user` is null,
- When `read-native-review` runs,
- Then none counts: with no review of the identity the result is exit 10 `pending:no-review`, never 0; a
  forged bound review beside the identity's real stale review does not turn that stale review into a pass;
  and another login's CHANGES_REQUESTED or COMMENTED review on the head neither blocks nor completes the
  identity's APPROVED review (another author is never read, in any state). The human's review is excluded
  because of its author, never because of its body: the identity's own APPROVED review whose body says the
  change is approved is not excluded and counts as clean beside it, so the body text of an artifact decides
  nothing in either direction (REQ-CCR-003, REQ-CCR-001).
- Verify: `go test -count=1 -v -run '^TestCodexGateReadNativeReviewIgnoresForged$' ./internal/template/` lists
  `--- PASS` for `other_bot_approved_bound`, `human_review_says_approved`,
  `lookalike_login_without_bot_suffix_same_id`, `same_login_other_id`, `same_login_type_user`,
  `review_user_null`, `forged_bound_review_beside_real_stale_review_not_pass`,
  `identity_approved_beside_other_login_changes_requested_clean`,
  `other_login_commented_never_completes_identity_review`,
  `identity_approved_body_says_approved_not_excluded` (the identity's APPROVED review with the body "approved"
  beside a human's review with the same body: exit 0, so exclusion follows the author alone).
- Mutant: a script that matches the login only turns `same_login_other_id` and `same_login_type_user` red; one
  that matches the id only turns `lookalike_login_without_bot_suffix_same_id` red; one that accepts any bound
  review of any author turns `other_bot_approved_bound` red; one that reads body text for "approved" turns
  `human_review_says_approved` red; one that excludes any review whose body text says the change is approved
  (the literal reading of the revision-4 REQ-CCR-003, which also drops the identity's own APPROVED review)
  turns `identity_approved_body_says_approved_not_excluded` red, because with its only candidate dropped it
  exits 10 `pending:no-review` where 0 is required; one that lets another login's CHANGES_REQUESTED block turns
  `identity_approved_beside_other_login_changes_requested_clean` red.
- RED-now: R1. Green path: M1.

**AC-CCR-004** — REQ-CCR-004 — release-blocking
- Given the identity's reviews bound to the head in each of these forms: COMMENTED with an empty body and no
  inline comment, with the body `LGTM`, with findings only in the body, with a "could not review" body, and
  with a usage-limit notice as the body; CHANGES_REQUESTED alone, after an APPROVED, followed by a COMMENTED, and followed by a DISMISSED; an
  APPROVED followed by a COMMENTED, and by a DISMISSED; an APPROVED review of another commit beside a COMMENTED
  review of the head; DISMISSED, PENDING, and an unknown state alone; an APPROVED review whose `submitted_at`
  is null or empty; an entry missing `commit_id`, `state`, or `id`; a repeated review `id`; an entry that is
  not an object; and a missing input file,
- When `read-native-review` runs,
- Then every row exits 1 with its own named reason and none exits 0: `review-not-conclusive:COMMENTED` for a
  newest COMMENTED review whatever its body; `changes-requested` while a CHANGES_REQUESTED review has no later
  APPROVED; `review-not-complete:<state>` for DISMISSED, PENDING, and unknown states;
  `review-not-complete:no-submitted-at`; `review-entry-invalid:<id>` and
  `review-entry-invalid:duplicate-id:<id>`; `input-invalid`. A CHANGES_REQUESTED review is superseded only by
  a later APPROVED review of the same commit (the supersession row is AC-CCR-001
  `changes_requested_then_approved_clean`). Decision table: design §3.2, rows 3-6, 8, 10-13, 15, 17, 18, 25.
- Verify: `go test -count=1 -v -run '^TestCodexGateReadNativeReviewRefusesIncomplete$' ./internal/template/`
  lists `--- PASS` for `commented_empty_body_refused`, `commented_body_lgtm_refused`,
  `commented_findings_only_in_body_refused`, `commented_could_not_review_body_refused`,
  `commented_usage_limit_notice_refused`, `changes_requested_alone_refused`, `approved_then_changes_requested_refused`,
  `changes_requested_then_commented_refused`, `changes_requested_then_dismissed_refused`,
  `approved_then_commented_refused`, `approved_then_dismissed_refused`,
  `approved_other_sha_and_commented_head_refused`, `state_dismissed`, `state_pending`, `state_unknown_value`,
  `approved_submitted_at_null`, `approved_submitted_at_empty`, `commit_id_missing`, `state_missing`,
  `id_missing`, `duplicate_review_id`, `review_entry_not_object`, `reviews_file_missing`,
  `comments_file_missing`.
- Mutant: **the revision-3 reader** (the newest bound review decides, and COMMENTED or APPROVED counts as
  complete) turns `commented_empty_body_refused`, `commented_body_lgtm_refused`,
  `commented_findings_only_in_body_refused`, `commented_could_not_review_body_refused`,
  `commented_usage_limit_notice_refused`, `changes_requested_then_commented_refused`,
  `approved_then_commented_refused`, and
  `approved_other_sha_and_commented_head_refused` red (this is the audit's "COMMENTED with no comment treated
  as clean" and "newest only" mutant). A reader that decides by the oldest bound review turns
  `approved_then_changes_requested_refused` red; one that reads body text to find a clean result turns
  `commented_body_lgtm_refused` red; one that treats DISMISSED or PENDING as complete, ignores
  `submitted_at`, or passes an entry with a missing field turns the matching row red.
- RED-now: R1. Green path: M1.

**AC-CCR-005** — REQ-CCR-005 — release-blocking
- Given a clean bound APPROVED review of the identity and a review comment by the identity whose `commit_id`
  equals the head, whose `original_commit_id` equals the head, or that carries neither field; and a COMMENTED
  review of the identity with an inline comment on the head,
- When `read-native-review` runs,
- Then it exits 1 with `finding:<id>` or `finding-unbound:<id>`; a comment by another author on the head, a
  comment bound only to an older head, and an empty comment list do not change a clean review's exit 0.
- Verify: `go test -count=1 -v -run '^TestCodexGateReadNativeReviewRefusesFinding$' ./internal/template/` lists
  `--- PASS` for `finding_commit_id_equals_head`, `finding_original_commit_id_equals_head`,
  `finding_without_any_commit_field`, `finding_by_other_author_ignored`,
  `finding_bound_to_older_head_ignored`, `comments_file_empty_clean`, `approved_review_with_finding_refused`,
  `commented_review_with_inline_finding_refused`.
- Mutant: a script that never reads the comments, counts only `commit_id`, counts comments of any author, or
  treats an unbound comment as no finding turns the matching row red.
- RED-now: R1. Green path: M1.

### 2.2 Auto-merge consumer (M2, M3)

**AC-CCR-006** — REQ-CCR-006 — release-blocking
- Given the auto-merge workflow and the gate script,
- When their text is scanned and the script's `fetch-pages` is run against disallowed endpoints,
- Then the `gh` calls are exactly the allow-list below and nothing else, neither file contains a forbidden
  token, and `fetch-pages` refuses every endpoint outside the three listings.
  **Allow-list (exact), classified by `gh` group and subcommand and by the first `repos/...` path argument:**

  | Where | Allowed form | Used by |
  |-------|--------------|---------|
  | workflow | `gh pr view "$PR_NUMBER" --json headRefOid …` | steps `head` and merge |
  | workflow | `gh pr checks "$PR_NUMBER" --required --json name,state,bucket` | step `checks` (the unchanged CI observation; it reads check state for the CI gate, never a review verdict) |
  | workflow | `gh pr merge "$PR_NUMBER" … --match-head-commit … --delete-branch` | merge step |
  | workflow | `gh api "repos/$REPO/pulls/$PR_NUMBER"` with no further path segment | steps `files` and `review` (pull request object: `draft`, `head.sha`, `changed_files`, `review_comments`) |
  | script | `gh api "<E>?per_page=100&page=<P>"` after `E` matched `^repos/<owner>/<repo>/pulls/<N>/(files\|reviews\|comments)$` | `fetch-pages` |

  Any other `gh` group or subcommand (`gh issue`, `gh pr comment`, `gh pr review`, `gh pr edit`, `gh run`,
  `gh workflow`, `gh label`) or any other `gh api` path fails. **Forbidden tokens** in either file: `issues/`,
  `/status`, `statuses`, `check-runs`, `check_runs`, `labels`, `CodeRabbit`, `Merge Risk`,
  `AUTOMERGE_MAX_MERGE_RISK`.
- Verify: `go test -count=1 -v -run '^TestAutoMergeReadsNoCommentOrStatusVerdict$' ./internal/template/` lists
  `--- PASS` for `gh_invocations_match_allow_list_exactly`, `gh_api_paths_in_workflow_are_pull_object_only`,
  `retained_gh_pr_forms_view_checks_merge_only`, `no_forbidden_token`, `no_coderabbit_or_merge_risk`,
  `no_risk_ceiling_variable`, `no_label_read`, `scan_swept_at_least_6_gh_invocations`; and `go test -count=1 -v
  -run '^TestCodexGateFetchPagesRefusesDisallowedEndpoint$' ./internal/template/` lists `--- PASS` for
  `issue_comments_endpoint_refused`, `check_runs_endpoint_refused`, `commit_status_endpoint_refused`,
  `path_with_extra_segment_refused`, `files_reviews_comments_endpoints_accepted`.
- Mutant: re-adding a comment-parse fallback (`gh api …/issues/…/comments`), a status read, a `gh pr comment`,
  or a `gh api` of any other path turns `gh_invocations_match_allow_list_exactly` (or the token row) red; a
  `fetch-pages` without the endpoint check turns the three refusal rows red; a scan that matches nothing turns
  `scan_swept_at_least_6_gh_invocations` red (at least the four retained `gh pr` forms and two `gh api` forms
  are swept).
- RED-now: R3 (the comment endpoint is read today) and R2. Green path: M2 for the script rows, M3 for the
  workflow rows.

**AC-CCR-007** — REQ-CCR-007 — release-blocking
- Given a native-review read that returns pending, a non-draft pull request whose review never appears, a
  stale review, a refusing read, a list fetch that fails on every pass, a review that appears only after the
  last pass, and the head moving during the loop,
- When the auto-merge review step's loop is read and run against a stub `gh` and a no-op `sleep`,
- Then the loop runs at most 60 passes and does not sleep after the last; pending keeps waiting; the last pass
  passes `--wait-closed true` and every earlier pass `false`; a refusal (including `stale-review`) withholds at
  once; a failing fetch counts as pending, and a list unreadable on every pass ends as a failure named
  `review-unreadable` (step exit status non-zero, job failure, REQ-CCR-018), never as NOT RUN and never as
  success; a review that appears after the window is not read; no merge is authorized at the end of the window. The script half: no review of the identity yet is
  exit 10, never 0.
- Verify: `go test -count=1 -v -run '^TestCodexGateReadNativeReviewPending$' ./internal/template/` lists
  `--- PASS` for `empty_reviews_pending`, `empty_reviews_and_comments_pending`; and `go test -count=1 -v -run
  '^TestAutoMergeWaitLoopIsBounded$' ./internal/template/` lists `--- PASS` for `loop_bound_at_most_60`,
  `pending_keeps_waiting`, `last_pass_passes_wait_closed_true`, `no_sleep_after_last_pass`,
  `exit_one_withholds_at_once`, `stale_withholds_at_once`, `fetch_failure_counts_pending`,
  `unreadable_to_end_of_window_fails_review_unreadable`, `review_arriving_after_window_not_read`,
  `expiry_never_merges`.
- Mutant: an absent read treated as success, expiry that authorizes a merge (`expiry_never_merges` red), a
  refusal read as pending, a stale review waited on for the whole window, an unbounded `while true`, a last
  pass that does not pass `--wait-closed true`, or an unreadable list at expiry reported as NOT RUN turns the
  matching row red.
- RED-now: R1 and R2. Green path: M1 for the script rows, M3 for the loop rows.

**AC-CCR-008** — REQ-CCR-008 — release-blocking
- Given a draft pull request — with no review, with a clean APPROVED review of the identity bound to the head,
  and with no input files at all — a non-draft pull request with a clean APPROVED review, and a non-draft pull
  request with no review of the identity when the wait window is closed and when it is open,
- When `read-native-review` runs with `--draft true` and `--draft false` and with `--wait-closed true` and
  `false`, and the auto-merge review step's handling of each exit code is read,
- Then at the script level NOT RUN is one exit class with two causes: the draft cases exit 11 with `not-run:draft` whatever artifacts exist; the
  absent-review case exits 10 `pending:no-review` while the window is open and 11 `not-run:review-absent`
  when it is closed; a stale review and a change request still exit 1 with the window closed; the non-draft
  clean APPROVED case exits 0; an invalid flag exits 1; the four classes 0, 1, 10, and 11 are pairwise
  distinct; and the step, on 11, withholds the merge, never treats 11 as success, reads `draft` from the pull request
  object, and does not fetch the review lists for a draft. **Revision 6:** what the step records and how it
  ends for each cause is `T-CHAIN` rows 4 and 5 and criteria AC-CCR-021 (non-draft, a failure: step exit 1,
  `state=fail`, `::error::` naming `not-run:review-absent`) and AC-CCR-022 (draft: `state=not-run:draft`, step
  exit 11, a `::notice::` that says it is not a review result and never a pass, job result never success);
  the closed SPEC's wording "records `state=not-run`, exits with status 0, neither a failure nor a pass" is
  withdrawn.
- Verify: `go test -count=1 -v -run '^TestCodexGateReadNativeReviewDraftNotRun$' ./internal/template/` lists
  `--- PASS` for `draft_true_exit_11`, `draft_true_with_clean_bound_review_still_11`,
  `draft_true_without_input_files_still_11`, `draft_false_approved_exit_0`, `draft_flag_invalid_exit_1`,
  `wait_closed_flag_invalid_exit_1`, `review_absent_window_open_exit_10`,
  `review_absent_window_closed_exit_11`, `changes_requested_window_closed_still_exit_1`,
  `four_classes_pairwise_distinct`; and `go test -count=1 -v -run '^TestAutoMergeNotRunWithholds$'
  ./internal/template/` lists `--- PASS` for `exit_11_withholds_merge`, `not_run_message_draft_names_cause`,
  `not_run_message_review_absent_names_cause`, `not_run_message_never_a_pass`,
  `not_run_message_gate_path_only`, `not_run_state_output_distinct_from_fail`,
  `exit_11_never_reaches_merge_condition`, `draft_read_from_pull_request_object`,
  `draft_skips_list_fetch_and_is_not_run`. This is the testable half of the NOT RUN rule: a draft observation,
  and the absence of a review, can never be read as a review result or as a pass, and each says so (design
  §12).
- Mutant: a script that exits 0 for a draft with a clean APPROVED review (a draft merged), one that exits 1 for
  a draft or for an absent review at window close (NOT RUN indistinguishable from failure), one that stays at
  10 at window close (absence never final), one that maps a stale review to 11, a loop that proceeds on 11, a
  `state` output of `fail` for a draft, or a message that claims a review result turns the matching row red.
- RED-now: R1 and R2. Green path: M1 for the script rows, M3 for the step rows.

**AC-CCR-009** — REQ-CCR-009 — release-blocking
- Given a changed-file list (one JSON object per line, as `fetch-pages` writes it) and the verified head SHA,
- When `gate-files-changed --file F --head-sha S` runs,
- Then it exits 1 when any `filename` or `previous_filename`, compared case-insensitively, is under
  `.github/workflows/` — any file there: the auto-merge workflow, an unrelated workflow, a newly added
  workflow with any name, a deleted one, a renamed one — or is the gate script; when the list is missing or
  empty, has an entry without a `filename`, or has 3000 or more entries; and when an entry that is not a
  removal carries a `contents_url` whose `ref` is not `S` or carries no `contents_url` (the list of a head
  that flipped from `H` to `T` and back: its entries carry `T`); and exits 0 for an unrelated list (a
  documentation file, a Go file) of up to 2999 entries whose non-removed entries all carry `ref=S`.
- Verify: `go test -count=1 -v -run '^TestCodexGateFilesChangedRefusals$' ./internal/template/` lists
  `--- PASS` for `workflow_auto_merge_edited`, `workflow_added_any_name`, `workflow_edited_other`,
  `workflow_deleted`, `workflow_case_variant_prefix`, `rename_into_workflows`, `rename_from_workflows`,
  `gate_script_edited`, `unrelated_files_clear`, `docs_and_go_files_clear`, `empty_list_refused`,
  `missing_file_refused`, `entry_without_filename_refused`, `list_of_3000_refused`,
  `list_of_2999_unrelated_clear`, `non_removed_entry_ref_other_commit_refused`,
  `non_removed_entry_without_contents_url_refused`, `entries_bound_to_head_clear`,
  `removed_entry_ref_not_checked_clear`, `ref_case_insensitive_equal_clear`, `head_sha_missing_refused`.
- Mutant: a guard that matches only `auto-merge.yml` and the script lets `workflow_added_any_name` and
  `workflow_edited_other` exit 0 and turns them red; ignoring `previous_filename`, a case-sensitive-only match
  (`workflow_case_variant_prefix`), reading an empty list as clear, or having no 3000 cap turns the matching
  row red. **A list read by pull request number, with no entry binding** (the audit's "H to T to H" mutant)
  exits 0 on `non_removed_entry_ref_other_commit_refused` and
  `non_removed_entry_without_contents_url_refused` and turns them red.
- RED-now: R4 (no tamper guard in the workflow) and R1. Green path: M2 for the script; that the workflow calls
  it, honours its exit code, and gates the merge on it is AC-CCR-012 (M3), which flips R4.

**AC-CCR-010** — REQ-CCR-010 — release-blocking
- Given a stub `gh` that serves a pull request's lists (changed files, reviews, review comments) in pages of
  100 by the `page` parameter and can fail a chosen page,
- When `fetch-pages` collects a list and the reader of that list runs on the result, with the pull request's
  own count passed as the expected count where one exists,
- Then a gate-path entry on a late page is found and refused — 101 entries with the gate path on page 2 and 301
  entries with it on page 4 both exit 1 with `gate-path-changed:<path>` — while an unrelated 350-entry list is
  clear; the identity's review on page 2 of the reviews list is read (exit 0) and a finding on page 3 of the
  comments list is refused; a failing page refuses with `page-error:<P>` and leaves no output file; an invalid
  page refuses with `page-invalid:<P>`; a collected count different from the expected count refuses with
  `count-mismatch:<collected>:<expected>`; a non-integer expected count refuses; a list of 3000 entries (a full
  30th page) refuses with `page-ceiling`; a list whose size is an exact multiple of 100 ends on an empty page
  and is collected whole; the reviews list, which has no count on the pull request object, is collected with
  `count-proof:none`; and an entry that repeats across pages (a review `id` or a `filename`) refuses with
  `page-duplicate:<key>` and removes the output.
- Verify: `go test -count=1 -v -run '^TestCodexGateFetchPagesCompleteness$' ./internal/template/` lists
  `--- PASS` for `single_page_clear`, `exact_200_ends_on_empty_page`,
  `entries_101_gate_file_on_page_2_refused`, `entries_301_gate_file_on_page_4_refused`,
  `entries_350_unrelated_clear`, `page_2_error_refuses_and_removes_out`, `page_invalid_json_refuses`,
  `fewer_entries_than_expected_refused`, `more_entries_than_expected_refused`,
  `expect_count_not_integer_refused`, `ceiling_3000_refused`, `reviews_without_count_proof_three_pages`,
  `review_comments_count_mismatch_refused`, `codex_review_on_page_2_is_read`, `finding_comment_on_page_3_refused`,
  `duplicate_review_id_across_pages_refused`, `duplicate_filename_across_pages_refused`,
  `entry_without_id_or_filename_refused`.
- Mutant: an implementation that reads page 1 only exits 0 on `entries_101_gate_file_on_page_2_refused` and
  `entries_301_gate_file_on_page_4_refused`, sees no review in `codex_review_on_page_2_is_read`, and turns
  them red (the first-page-only counterexample of the audit); one that ignores the exit status of `gh` or
  continues past a failed page turns `page_2_error_refuses_and_removes_out` red; one that never compares the
  collected count turns `fewer_entries_than_expected_refused` red; one with no 3000 ceiling turns
  `ceiling_3000_refused` red; one that stops at the first short page without an empty-page rule turns
  `exact_200_ends_on_empty_page` red when the stub serves a third non-empty page; one that accepts a repeated
  entry turns `duplicate_review_id_across_pages_refused` and `duplicate_filename_across_pages_refused` red (the
  stub serves a list that shifted by one between pages with the same total count).
- RED-now: R1 (the script, and so `fetch-pages`, is absent). Green path: M2.

**AC-CCR-011** — REQ-CCR-011 — regression-guard
- Given the auto-merge workflow after the edit,
- When its checkout step and its guard steps are read,
- Then the checkout has no `ref`, the head re-verification step, the step named `Check all required CI
  checks passed`, the `--match-head-commit` merge pin, and both deadline re-evaluations are all still
  present, and the existing required-checks test still passes.
- Verify: `go test -count=1 -v -run '^TestAutoMergeCheckoutAndGuardsUnchanged$' ./internal/template/` lists
  `--- PASS` for `checkout_has_no_ref`, `head_reverify_present`, `required_checks_step_name_present`,
  `match_head_commit_present`, `deadline_reeval_present`; and `go test -count=1 -v -run
  '^TestAutoMergeRequiredChecks$' ./internal/template/` lists its 16 `--- PASS` subtests.
- Mutant: `ref: ${{ github.event.workflow_run.head_sha }}` on the checkout, or removing `--match-head-commit`,
  turns a row red. The red must be recorded verbatim at M3.
- RED-now: none (green on the baseline by nature). Green path: stays green through M3.

**AC-CCR-012** — REQ-CCR-012 — release-blocking
- Given the real `auto-merge.yml` parsed as YAML, and its two new steps (`files`, `review`) extracted and run
  against a stub `gh` and either the real gate script copy or a stub script that exits a chosen code,
- When the step order, the step conditions, the permission set, the arguments passed, and the merge condition
  are read, and the steps are run,
- Then the order is: required-checks step, then `files`, then `review`, then the merge step; the `files` step
  reads the pull request object before and after the list, passes the pull request's `changed_files` as the
  expected count, runs `fetch-pages` then `gate-files-changed` with the verified head SHA, and writes
  `ok=true` only when both exit 0 and the head read twice equals the verified SHA; the `review` step reads the
  pull request object on every pass, passes `review_comments` as the expected count for the comments list,
  runs `fetch-pages` on reviews and comments, reads the pull request object again after those list reads
  (a head other than the verified SHA ends the step with `ok=false` `state=fail` and the lists are never
  judged), then runs `read-native-review` with the verified head SHA of step `head` and the `draft` flag, and
  writes `ok=true` only on exit 0; every other exit — 1, 11, and an exit the
  step does not recognise — leaves `ok` not `true` and ends the step with a non-zero exit status (REQ-CCR-018;
  the same holds for every non-pass path of the `files` step, including a missing script); a head other than the verified SHA ends either step with
  `ok=false`; the `review` step is skipped unless `files` passed; the merge step's `if` requires both
  `steps.files.outputs.ok == 'true'` and `steps.review.outputs.ok == 'true'` besides its existing terms; both
  steps run `scripts/ci/codex-review-gate.sh` from the checkout that has no `ref` (the default branch), and a
  missing script is not a pass; and the workflow permissions stay exactly `contents: write` and
  `pull-requests: write`.
- Verify: `grep -c 'gate-files-changed' .github/workflows/auto-merge.yml`, `grep -c 'read-native-review'
  .github/workflows/auto-merge.yml`, and `grep -c 'fetch-pages' .github/workflows/auto-merge.yml` each print a
  count of at least 1 and exit 0 (R4, R5, R6 flipped; necessary, not sufficient); and `go test -count=1 -v
  -run '^TestAutoMergeWiresGuardAndGate$' ./internal/template/` lists `--- PASS` for `files_step_present`,
  `review_step_present`, `order_checks_then_files_then_review_then_merge`,
  `files_step_depends_on_head_and_checks_ok`, `review_step_depends_on_files_ok`,
  `merge_if_requires_files_and_review_ok`, `steps_run_script_from_default_checkout`,
  `files_step_passes_changed_files_count`, `review_step_passes_review_comments_count`,
  `files_step_passes_verified_head_sha`, `permissions_unchanged`, `files_step_workflow_edit_not_ok`,
  `files_step_clear_list_ok_true`, `files_step_page_error_not_ok`, `files_step_missing_script_not_ok`,
  `files_step_head_moved_before_read_not_ok`, `files_step_head_moved_during_read_not_ok`,
  `files_step_count_moved_during_read_not_ok`, `review_step_clean_ok_true_state_pass`,
  `review_step_exit_1_not_ok_state_fail`, `review_step_unknown_exit_not_ok`, `review_step_head_moved_not_ok`,
  `review_step_head_moved_during_read_not_ok` (the stub `gh` answers the pull request object with the verified
  head on the first read of a pass and with another SHA on the second, beside lists that are the clean fixture
  for the verified head: `ok` is `false`, `state` is `fail`, and, the stub script recording each subcommand
  invocation, no `read-native-review` (verdict) invocation is recorded after the second head read observes the
  other SHA, while the `fetch-pages` invocations before that read are allowed and expected),
  `review_step_forged_review_never_ok`. **Revision 6:** each `files_step_*_not_ok` and `review_step_*_not_ok`
  row also asserts a non-zero step exit status, and each `*_ok_true` row asserts exit status 0 (the full step
  and job assertions are AC-CCR-020 to AC-CCR-025).
- Mutant: deleting the `files` step turns `files_step_present` red; a `files` step that discards the exit code
  of the guard (`|| true`, or `ok=true` written regardless) turns `files_step_workflow_edit_not_ok` and
  `files_step_page_error_not_ok` red; a `files` step that reads the list by pull request number and never
  reads the head again turns `files_step_head_moved_during_read_not_ok` red; a `review` step that reads the
  head only before its list reads and never again after them turns
  `review_step_head_moved_during_read_not_ok` red (on that fixture it would pass the clean lists to the
  script, exit 0, and write `ok=true`); a `review` step that ignores the
  exit code turns `review_step_exit_1_not_ok_state_fail` red; treating exit 11 or an unknown exit as pass turns
  `review_step_unknown_exit_not_ok` red (exit 11 is AC-CCR-008); a merge `if` without the `review` term turns
  `merge_if_requires_files_and_review_ok` red, and so does a merge `if` without the `files` term: that term
  is pinned here, structurally (the condition is parsed and must name both `steps.files.outputs.ok == 'true'`
  and `steps.review.outputs.ok == 'true'`), because AC-CCR-019 cannot observe its removal by behaviour (the
  `review` step already depends on `files`); placing the merge step before the gate steps turns
  `order_checks_then_files_then_review_then_merge` red; omitting the count arguments or the head SHA turns
  `files_step_passes_changed_files_count`, `review_step_passes_review_comments_count`, or
  `files_step_passes_verified_head_sha` red. The unit tests of the script alone (AC-CCR-001 to AC-CCR-010)
  cannot catch any of these, which is why this criterion exists; the merge decision itself is AC-CCR-019.
- RED-now: R4, R5, R6 (each count 0, exit 1 on the unchanged workflow). Green path: M3 flips all three and
  lists the subtests.

**AC-CCR-013** — REQ-CCR-013 — regression-guard
- Given `auto-merge.yml` after the edit and the set of workflow files added by this SPEC's commits,
- When the triggers, the secret references, and the steps are read,
- Then the only trigger is `workflow_run`, no trigger of the `pull_request_target` family appears, the only
  secret reference is the workflow token, no step uploads, downloads, or caches an artifact (no
  `actions/upload-artifact`, `actions/download-artifact`, `actions/cache`, no `cache:` input), and this SPEC
  adds no workflow file.
- Verify: `go test -count=1 -v -run '^TestAutoMergeAddsNoSecretTriggerOrArtifact$' ./internal/template/` lists
  `--- PASS` for `trigger_is_workflow_run_only`, `no_pull_request_target_family`,
  `secrets_only_workflow_token`, `no_artifact_or_cache_step`; and `git diff --quiet --diff-filter=A A HEAD --
  .github/workflows` exits 0 where `A` is the recorded merge base of AC-CCR-016, with the control `git diff
  --name-only --diff-filter=D A HEAD -- .github/workflows` listing the two deleted workflows (a vacuous empty
  range is "not measurable").
- Mutant: adding `pull_request_target`, `${{ secrets.X }}`, an `actions/cache` step, or a new workflow file
  turns the matching row or command red. The red must be recorded verbatim at M3. Observed failure of the
  command form on a tree known to add a workflow: `git diff --quiet --diff-filter=A 4f9dba14bf 6dde1c1c1b --
  .github/workflows` exits 1 (that range adds `claude-code-review.yml`) beside the baseline `git diff --quiet
  --diff-filter=A 4f9dba14bf HEAD -- .github/workflows` exit 0 (tree `e11af67632`). The deleted-workflow control
  holds at landing, after M4; at plan time it is empty.
- RED-now: none (green on the baseline by nature). Green path: stays green through M4.

### 2.3 Retirement (M4)

**AC-CCR-014** — REQ-CCR-014 — release-blocking
- Given every workflow file in `.github/workflows/`,
- When the census runs,
- Then zero of them contain `anthropics/claude-code-action`, zero have a top-level `name:` of `Claude Code
  Review`, zero trigger on `check_run`, and `claude.yml`, `review-quality-gate.yml`, `claude-code-review.yml`
  do not exist; the census swept at least 18 workflow files (21 on the measured tree, 19 expected after the
  change).
- Verify: `go test -count=1 -v -run '^TestNoClaudeReviewWorkflows$' ./internal/template/` lists `--- PASS` for
  `no_claude_code_action`, `no_workflow_named_claude_code_review`, `no_check_run_trigger`, `claude_yml_absent`,
  `review_quality_gate_yml_absent`, `claude_code_review_yml_absent`, `sweep_is_nonempty`.
- Mutant: restoring `claude.yml`, or merging the remote `claude-code-review.yml` back in, turns a row red; that
  second case is the one the census exists for (E5).
- RED-now: R7 and R10. Green path: M4.

**AC-CCR-015** — REQ-CCR-015 — release-blocking
- Given `.coderabbit.yaml`, the workflows, and `scripts/ci/`,
- When they are read,
- Then the file exists, `reviews.auto_review.enabled` is `false`, `chat.auto_reply` is `false`, and zero files
  under `.github/workflows/` or `scripts/ci/` mention CodeRabbit; both directories were swept non-empty.
- Verify: `go test -count=1 -v -run '^TestCodeRabbitDisabledAndUnread$' ./internal/template/` lists `--- PASS`
  for `config_present`, `auto_review_false`, `auto_reply_false`, `no_coderabbit_in_workflows`,
  `no_coderabbit_in_scripts_ci`, `sweep_is_nonempty`.
- Mutant: deleting the file, leaving `enabled: true`, or reading the CodeRabbit status in a workflow turns a
  row red.
- RED-now: R8 (`65:    enabled: true`, `74:  auto_reply: true`) and R2. Green path: M4 (the workflow reads
  end at M3; the config flags flip at M4).

### 2.4 Contract preservation and liveness (M4)

**AC-CCR-016** — REQ-CCR-016 — regression-guard
- Given the merge base `A` of `HEAD` and local `main`, read at measurement time by `git merge-base main HEAD`
  and recorded as a commit SHA (on the measured tree it prints `4f9dba14bf2834aa08bea79f9e72c2a7e75ab18e`,
  exit 0),
- When the three contract files and the four workflow files that publish or gate the measured required checks
  (`ci.yml`, `codeql.yml`, `release-pr-multi-os.yml` — which carries the job named `Release PR Multi-OS Gate`,
  line 275 — and its PR-side sibling `pr-multi-os-gate.yml`) are compared with the branch tip at landing,
- Then the comparison is empty, and a control shows the range itself is not empty.
- Verify: `git merge-base main HEAD` prints `A` (record it); then `git diff --quiet A HEAD --
  .github/required-checks.yml .github/branch-protection.json.gtmpl
  internal/template/templates/.github/branch-protection.json.gtmpl .github/workflows/ci.yml
  .github/workflows/codeql.yml .github/workflows/pr-multi-os-gate.yml .github/workflows/release-pr-multi-os.yml`
  exits 0; and the control `git diff --name-only A HEAD` lists at least one path (a vacuous empty range is
  "not measurable", never "no change"). Landing-time evidence; the left end of the diff is the recorded SHA
  `A`, never a branch name, and after the standard absorb-`main`-before-landing merge `A` is the absorbed
  `main` commit, so another card's change to `ci.yml` or `required-checks.yml` is not read as this SPEC's. The
  control reads 10 paths on the measured tree `e11af67632` (the SPEC files and the card record). "Measured
  state" here is the file at that merge base (E11, E26), not the remote branch protection (a moving system,
  E6, never a requirement premise). The criterion is evaluated before the merge into `main`: after it the
  merge base is the card tip and the range is empty by construction.
- Mutant: any edit to one of those paths exits 1. Observed red on a tree known to differ: R9 (exit 1 against
  `6dde1c1c1b`) beside the green baseline (exit 0).
- RED-now: none by nature; the observed failure is R9.

**AC-CCR-017** — REQ-CCR-016 — regression-guard
- Given `.github/required-checks.yml` parsed,
- When every branch's contexts, the release gate, and the auxiliary list are read,
- Then zero entries contain the word `codex` (case-insensitive) as a review check, `main` still lists `Lint`,
  and the failure message names confirmation gate G4.
- Verify: `go test -count=1 -v -run '^TestRequiredChecksCarryNoCodexReviewCheckUntilConfirmed$'
  ./internal/config/` lists `--- PASS` for `not_in_main`, `not_in_ci_candidate_key`, `not_in_release_key`,
  `not_in_release_gate`, `not_in_auxiliary`, `control_main_lists_lint`.
- Mutant: adding a `Codex Review` context to `branches.main.contexts` or `auxiliary` turns a row red; the run
  phase records that red verbatim at M4. The test is permanent: removing it is the user-confirmed change of
  gate G4.
- RED-now: none by nature.

**AC-CCR-018** — REQ-CCR-017 — release-blocking
- Given `.moai/guards/liveness.yaml` parsed by the repository's own manifest loader,
- When its entries are read after the two workflows are deleted,
- Then the entries name exactly the workflows on disk: none names `claude.yml` or `review-quality-gate.yml`,
  and the existing census still holds as a set equality in both directions (no entry names a file that is not
  on disk, none is missing).
- Verify: `go test -count=1 -v -run '^TestCensusOmitsRetiredReviewWorkflows$' ./internal/guardstate/` lists
  `--- PASS` for `claude_entry_absent`, `review_quality_gate_entry_absent`; and `go test -count=1 -v -run
  '^(TestCensus_SetDifferenceEmptyBothDirections|TestCensus_EveryEntryCarriesFiveFields|TestCensus_ReleaseOnlySubjectsDeclareTheirCondition)$'
  ./internal/guardstate/` lists a `--- PASS` line for each of the three names.
- Mutant: leaving either entry in place turns the matching row red (and the existing census reports
  `manifest\disk` once the file is deleted).
- RED-now: R11 (both entries present) and R10 (both files present). Green path: M4.

### 2.5 Merge decision path (M3)

**AC-CCR-019** — REQ-CCR-012 — release-blocking
- Given the real `auto-merge.yml` parsed as YAML, a stub `gh` that records every call and answers the pull
  request object, the listings, `gh pr view`, and `gh pr merge`, a stub gate script that exits a chosen code
  (and the real script for the end-to-end rows), the `merge_deadline` environment value set to the past or
  the future, and an evaluator for the `if:` conditions limited to the grammar the workflow uses
  (`steps.<id>.outputs.<name> == '<literal>'` terms joined by `&&`, optionally led by the status function
  `always()`, which the harness reads as the exemption from the implied `success()`; any other condition shape
  fails the test),
- When the chain `head → checks → files → review → merge condition → merge step` is run on each fixture,
- Then on the clean fixture `gh pr merge` is called exactly once, with `--match-head-commit` equal to the
  verified head SHA, `--delete-branch`, and the squash flag for a feature branch (the merge flag for a
  `release/` branch); and on each of these fixtures `gh pr merge` is called zero times: the gate exits 1, the
  gate exits 11 for a draft, no review appears through the whole window (NOT RUN), the gate exits an
  unrecognised code, the review is stale, the head differs at the merge step's own `gh pr view`, the deadline
  has passed, a gate path is changed, and a list page fails. **Revision 6:** the harness also records each step's exit status and the derived job
  result: on every refusal fixture above the job result is failure (the `files` step, the `review` step, or the
  gate-script-missing fixture ends non-zero and the later steps are skipped), on the clean fixture it is success,
  and the two merge-step-owned guards (the head differs at the merge step's own `gh pr view`, the deadline has
  passed) keep a success job with zero merge calls (Out of Scope of the amendment).
- Verify: `go test -count=1 -v -run '^TestAutoMergeMergeDecisionPath$' ./internal/template/` lists `--- PASS`
  for `clean_fixture_merges_once_with_head_pin`, `release_branch_clean_fixture_uses_merge_flag`,
  `review_exit_1_zero_merge_calls`, `review_draft_exit_11_zero_merge_calls`,
  `review_absent_expiry_zero_merge_calls`, `review_unknown_exit_zero_merge_calls`,
  `stale_review_zero_merge_calls`, `head_moved_at_merge_step_zero_merge_calls`,
  `deadline_expired_zero_merge_calls`, `gate_path_change_zero_merge_calls`, `page_failure_zero_merge_calls`,
  `gate_script_missing_zero_merge_calls`, `refusal_fixtures_job_result_failure_clean_fixture_success`,
  `if_expression_outside_grammar_fails_test`.
- Mutant (each is applied to the committed workflow and must turn the named row red): `&& false` appended to
  the merge `if`, or an `exit 0` placed before `gh pr merge`, turns `clean_fixture_merges_once_with_head_pin`
  red (zero calls where one is required; the first also turns `if_expression_outside_grammar_fails_test`
  red); `merge_deadline=9999999999` assigned inside the merge step turns `deadline_expired_zero_merge_calls`
  red; **dropping `steps.review.outputs.ok == 'true'` from the merge `if` is an equivalent mutant under the
  revision-6 rules** (a failed or non-passing `review` step skips the merge through the implied `success()`, so
  merge calls are 0 in the original and the mutant) and is covered by the structural checks of AC-CCR-012
  (`review_step_depends_on_files_ok`, `merge_if_requires_files_and_review_ok`, and the review and `files`
  condition rows), not by a named AC-CCR-019 row; **mutant MF** (the `files` term removed from the `review` step's `if` and from the merge `if`, and
  both conditions led with `always()`) turns `gate_path_change_zero_merge_calls` red — the step-by-step trace is
  under "Mutant MF trace" below; removing the `files` term alone, from either condition, is an equivalent mutant
  under the revision-6 harness rules (a failed `files` step skips every later step whose `if` has no status
  function, so merge calls are 0 in the original and the mutant) and is pinned structurally by AC-CCR-012
  (`review_step_depends_on_files_ok`, `merge_if_requires_files_and_review_ok`); dropping `--match-head-commit` or pinning it to another variable
  turns `clean_fixture_merges_once_with_head_pin` red. This is the criterion that a workflow which never merges, or
  merges past a refusal, cannot pass; AC-CCR-011 and AC-CCR-012 check strings and steps and cannot.
- RED-now: R12 (`grep -c -F 'steps.review.outputs.ok' .github/workflows/auto-merge.yml` prints 0, exit 1: the
  merge condition names no review term today, so no review refusal can reach the merge decision). Green path:
  M3.

**Mutant MF trace (plan-time reasoning under the revision-6 harness rules; NOT executed. The executed red, with
its verbatim output, belongs to the run phase's M5 mutant run under the admission rule, and no report may say
it was observed before then).** Harness input, row `gate_path_change_zero_merge_calls`: pull request object
with head `H`, not a draft, `changed_files` listing one `.github/workflows/x.yml` entry bound to `H`; reviews
list holding a clean APPROVED review of the identity bound to `H`, comments empty; `gh pr checks` all passing;
`merge_deadline` in the future; `gh pr view` answering `H`; step order `head`, `checks`, `files`, `review`, merge.
Harness rules used: a step ending non-zero without `continue-on-error` makes the job failure; a later step whose
`if` has no status function is skipped; `always()` lifts that skip and the remaining terms still decide.

| Stage | Original workflow | Mutant MF |
|-------|-------------------|-----------|
| `head`, `checks` | exit 0, `ok=true`, `should_merge=true` | same |
| `files` | `gate-files-changed` exits 1 (`gate-path-changed`): step exit 1, `ok=false`, `state=fail` (`T-CHAIN` row 13) | same |
| `review` `if` | `head.ok == 'true' && checks.should_merge == 'true' && files.ok == 'true'`: `files.ok` is `false`, so false (and the implied `success()` is false after the failed step): **skipped**, no output | `always() && head.ok == 'true' && checks.should_merge == 'true'`: `always()` lifts the skip, the remaining terms are true: **runs**; lists are the clean fixture, gate exit 0, recheck agrees: step exit 0, `ok=true`, `state=pass` |
| merge `if` | contains `steps.review.outputs.ok == 'true'`: empty, so false: **skipped** | `always() && head.ok && checks.should_merge && review.ok == 'true'`: true: **runs**; deadline in the future, `gh pr view` equals `H`: `gh pr merge ... --match-head-commit H` **called once** |
| Assertion `merge calls == 0` | 0: pass | **1: the named row turns red** |
| Job result | failure (from `files`) | failure (unchanged), so the job-result assertion does not flip; the merge-call assertion does |

The same mutant also turns red, by structure, `review_step_depends_on_files_ok`, `merge_if_requires_files_and_review_ok`
(AC-CCR-012) and `guard_steps_have_no_continue_on_error_or_status_override` (AC-CCR-024). The committed workflow
contains no `always()`, so the evaluator's added allowance changes no original result.

### 2.6 Step and job conclusion (plan revision 6; M5)

Every row of AC-CCR-020 to AC-CCR-025 runs the real step body of `auto-merge.yml` (the real gate script, a stub
`gh`, a no-op `sleep`) through the shared step harness and asserts **three observables together**: the step's
exit status, the job result the harness derives from the executed steps (failure when a step that is not
`continue-on-error: true` exits non-zero, and later steps with a status-function-free `if` skipped; success
otherwise), and the step's output-file content. The derived job result is a rule over the workflow's
structure; what GitHub reports on a runner is UNOBSERVED (spec §4, research U11). The table that every row
instantiates is `design.md` §5 `T-CHAIN`.

**AC-CCR-020** — REQ-CCR-018 — release-blocking
- Given the real `review` step run with each terminating condition of `T-CHAIN` rows 2, 3, 6, 7, 8, 9 — a
  changes-requested review, a finding comment, a stale review, a COMMENTED review, an unknown gate exit, the
  head moved during the read, the review-comment count moved during the read, a pre-merge recheck that differs,
  lists unreadable to the end of the window, and a missing gate script — and with the clean pass of row 1,
- When the step runs through the harness,
- Then on every failure row the step exit status is non-zero (1), the derived job result is failure, `ok` is
  `false`, `state` is `fail`, the annotation is `::error::` naming the reason, and the merge step is skipped
  with zero `gh pr merge` calls; on the clean row the step exit status is 0, the job result is success,
  `ok=true`, `state=pass`; and the `review` step carries no `continue-on-error`.
- Verify: `go test -count=1 -v -run '^TestAutoMergeReviewFailureFailsStepAndJob$' ./internal/template/` lists
  `--- PASS` for `changes_requested_step_exit_nonzero_job_failure`, `finding_step_exit_nonzero_job_failure`,
  `stale_review_step_exit_nonzero_job_failure`, `commented_not_conclusive_step_exit_nonzero`,
  `unknown_gate_exit_step_exit_nonzero`, `head_moved_during_read_step_exit_nonzero`,
  `review_comments_moved_during_read_step_exit_nonzero`, `recheck_mismatch_step_exit_nonzero`,
  `unreadable_to_end_of_window_step_exit_nonzero`, `gate_script_missing_step_exit_nonzero`,
  `merge_step_skipped_after_review_failure`, `clean_pass_step_exit_0_job_success`,
  `review_step_has_no_continue_on_error`.
- Mutant: the committed workflow (a `withhold` that ends `exit 0`) turns every failure row red on the exit
  status although its outputs read `ok=false state=fail`; `continue-on-error: true` on the step turns the
  job-result assertion of every failure row and `review_step_has_no_continue_on_error` red while the step exit
  status stays non-zero; a step that exits 1 on the clean pass turns `clean_pass_step_exit_0_job_success` red.
- RED-now: R13 and R16 (tree `0955fb9058`): the refusal path ends `exit 0`, and the tests are absent. Green
  path: M5.

**AC-CCR-021** — REQ-CCR-008 — release-blocking
- Given a non-draft pull request with no native review of the identity when the wait window ends (gate exit 11,
  `not-run:review-absent`), with the window still open, and with the lists unreadable at the end of the window,
- When the real `review` step runs through the harness,
- Then at the window's end the step exit status is 1, the job result is failure, `state=fail` (never
  `not-run:draft`, never a second key), the annotation is `::error::` naming `not-run:review-absent`, and no
  `gh pr merge` call is made; with the window open the step keeps waiting and records no output; with the lists
  unreadable the failure is named `review-unreadable`.
- Verify: `go test -count=1 -v -run '^TestAutoMergeReviewAbsentIsFailure$' ./internal/template/` lists `--- PASS`
  for `review_absent_window_closed_step_exit_1_job_failure`, `review_absent_state_fail_not_not_run`,
  `review_absent_annotation_error_names_cause`, `review_absent_zero_merge_calls`,
  `review_absent_window_open_keeps_waiting_no_output`, `lists_unreadable_at_expiry_step_exit_1_review_unreadable`.
- Mutant: a step that ends review-absent with exit 0 and `state=not-run` (the committed behavior) turns the first
  two rows red; one that reads review-absent as success turns `review_absent_zero_merge_calls` red.
- RED-now: R13, R14, R16. Green path: M5.

**AC-CCR-022** — REQ-CCR-008 — release-blocking
- Given a draft pull request — with no review, with a clean APPROVED review bound to the head, and with a
  review that CHANGES_REQUESTED —
- When the real `review` step runs through the harness,
- Then the step exit status is 11, the derived job result is failure and in no case success, `ok=false`,
  `state` is the single value `not-run:draft` (the key appears once; it is distinct from `pass` and from
  `fail`), the annotation is a `::notice::` that names the cause `draft`, says it is not a review result and
  never a pass, the review lists are not fetched, and no `gh pr merge` call is made; the state never reads
  as a review PASS.
- Verify: `go test -count=1 -v -run '^TestAutoMergeDraftIsNotRunNeverSuccess$' ./internal/template/` lists
  `--- PASS` for `draft_step_exit_11_job_not_success`, `draft_state_output_is_single_value_not_run_draft`,
  `draft_ok_output_false`, `draft_skips_list_fetch`, `draft_with_clean_approved_review_still_not_success`,
  `draft_annotation_notice_names_cause_and_says_not_a_pass`, `draft_zero_merge_calls`,
  `draft_state_distinct_from_fail_and_pass`.
- Mutant: the committed step (draft ends `exit 0`, `state=fail` then `state=not-run`) turns the first, second,
  and last rows red; mapping a draft to `state=pass` or `ok=true` turns `draft_ok_output_false` red;
  `continue-on-error` turns the job-result assertion red.
- RED-now: R14 and R16. Green path: M5.

**AC-CCR-023** — REQ-CCR-019 — release-blocking
- Given the real `files` step and the real `review` step run on a pass, on a refusal, on a draft (review), and
  on a review-absent window (review),
- When their output files are read after the run,
- Then every key appears exactly once, `ok` is `true` or `false`, `state` is within the allowed set of
  `T-OUTPUTS`, the pair is one of (`true`,`pass`), (`false`,`fail`), (`false`,`not-run:draft`), and no default is
  written before the terminating path.
- Verify: `go test -count=1 -v -run '^TestAutoMergeGuardStepsOutputOneValuePerKey$' ./internal/template/` lists
  `--- PASS` for `review_step_clean_pass_each_key_once`, `review_step_changes_requested_each_key_once`,
  `review_step_draft_each_key_once`, `review_step_review_absent_each_key_once`, `files_step_clear_each_key_once`,
  `files_step_refusal_each_key_once`, `output_pairs_within_allowed_set`.
- Mutant: pre-writing `ok=false` / `state=fail` and appending the terminal value (the committed behavior) turns
  every row red, the clean-pass row included (`ok` and `state` each appear twice).
- RED-now: R14 (and the clean-pass reproduction: `ok=false state=fail ok=true state=pass`). Green path: M5.

**AC-CCR-024** — REQ-CCR-018 — release-blocking
- Given the step harness of `internal/template/auto_merge_codex_test.go`,
- When its comparison and its job-result derivation are exercised on synthetic step results,
- Then a result whose outputs match but whose exit status differs fails the comparison; a result whose exit
  status matches but whose derived job result differs fails it; the job result is failure whenever a step that
  is not `continue-on-error` exits non-zero and success only when every executed step exits 0 or is skipped; and
  neither `files` nor `review` carries `continue-on-error` or a status function in its `if` that exempts it.
- Verify: `go test -count=1 -v -run '^TestAutoMergeStepHarnessObservesExitAndJobResult$' ./internal/template/`
  lists `--- PASS` for `harness_records_exit_status_and_outputs_per_step`,
  `harness_job_result_failure_when_a_step_exits_nonzero`,
  `harness_job_result_success_only_when_all_steps_zero_or_skipped`,
  `harness_comparison_rejects_exit_mismatch_with_matching_outputs`,
  `harness_comparison_rejects_job_result_mismatch_with_matching_exit`,
  `guard_steps_have_no_continue_on_error_or_status_override`.
- Mutant: a comparison that checks the outputs only turns the two `harness_comparison_rejects_*` rows red; a
  derivation that ignores the exit status turns `harness_job_result_failure_when_a_step_exits_nonzero` red.
- RED-now: R16 (the harness and its rows are absent). Green path: M5.

**AC-CCR-025** — REQ-CCR-018 — release-blocking
- Given the real chain `files` → `review` → merge condition → merge step run through the harness on failure
  paths of the `files` step (`T-CHAIN` rows 11 to 13): the gate script missing, a gate-path change, a page
  error, the pull request unreadable, the head moved before and during the read, the count moved during the
  read, and an unknown gate exit — and on the clean chain,
- When each step's exit status and the job result are read,
- Then on every `files` failure row the `files` step exit status is 1, `ok=false`, `state=fail`, the `review`
  step is skipped (not run, no `gh api` list call), the merge step is skipped with zero `gh pr merge` calls,
  and the job result is failure; with the gate script absent the same holds for the whole chain (the
  missing script never reads as a passing step); on the clean chain every step exits 0, `gh pr merge` is
  called once, and the job result is success.
- Verify: `go test -count=1 -v -run '^TestAutoMergeFilesStepFailureFailsStepAndJob$' ./internal/template/` lists
  `--- PASS` for `gate_script_missing_files_step_exit_nonzero_job_failure`,
  `gate_path_changed_step_exit_nonzero_job_failure`, `fetch_pages_error_step_exit_nonzero`,
  `pr_unreadable_step_exit_nonzero`, `head_moved_before_read_step_exit_nonzero`,
  `head_moved_during_read_step_exit_nonzero`, `count_moved_during_read_step_exit_nonzero`,
  `unknown_gate_exit_step_exit_nonzero`, `files_failure_skips_review_and_merge_steps`,
  `chain_gate_script_missing_job_failure_zero_merges`, `files_step_has_no_continue_on_error`,
  `clean_chain_step_exits_0_one_merge_job_success`.
- Mutant: the committed `files` step (`withhold` ends `exit 0`; a missing script ends `exit 0` with
  `ok=false state=fail`, R15) turns every failure row red on the exit status and the job result;
  `continue-on-error: true` on `files` turns the job-result assertions red; a chain runner that asserts the
  output file only passes on the committed workflow and is rejected by AC-CCR-024.
- RED-now: R15 and R16. Green path: M5.

## 3. REQ-to-AC matrix

| REQ | AC | REQ | AC |
|-----|----|-----|----|
| REQ-CCR-001 | AC-CCR-001 | REQ-CCR-010 | AC-CCR-010 |
| REQ-CCR-002 | AC-CCR-002 | REQ-CCR-011 | AC-CCR-011 |
| REQ-CCR-003 | AC-CCR-003 | REQ-CCR-012 | AC-CCR-012, AC-CCR-019 |
| REQ-CCR-004 | AC-CCR-004 | REQ-CCR-013 | AC-CCR-013 |
| REQ-CCR-005 | AC-CCR-005 | REQ-CCR-014 | AC-CCR-014 |
| REQ-CCR-006 | AC-CCR-006 | REQ-CCR-015 | AC-CCR-015 |
| REQ-CCR-007 | AC-CCR-007 | REQ-CCR-016 | AC-CCR-016, AC-CCR-017 |
| REQ-CCR-008 | AC-CCR-008, AC-CCR-021, AC-CCR-022 | REQ-CCR-017 | AC-CCR-018 |
| REQ-CCR-009 | AC-CCR-009 | REQ-CCR-018 | AC-CCR-020, AC-CCR-024, AC-CCR-025 |
| | | REQ-CCR-019 | AC-CCR-023 |

Counts: 19 requirements, 25 criteria (ceilings 25 and 25; the criterion count is at the ceiling, so a further
criterion means tiering up or splitting). Every requirement maps to at least one criterion;
every criterion maps to exactly one requirement; AC-CCR-016 and AC-CCR-017 share REQ-CCR-016, AC-CCR-012
and AC-CCR-019 share REQ-CCR-012, AC-CCR-008, AC-CCR-021, and AC-CCR-022 share REQ-CCR-008, and AC-CCR-020,
AC-CCR-024, and AC-CCR-025 share REQ-CCR-018. No criterion claims that doctrine is updated, and none depends on it (spec
§7, KD-1). No criterion depends on a deferred gate or on DS-1, DS-2, or DS-3.

### 3.1 The four function-check failures, each pinned

The operator named four conditions that must each be withheld — never a pass. Each is pinned by a criterion
and by a mutant that a weaker implementation would satisfy.

| Condition | Observable verdict | Pinned by | Mutant that turns it red |
|-----------|--------------------|-----------|--------------------------|
| Missing review | `read-native-review` exits 10 (`pending:no-review`) while the window is open, never 0; at window close it exits 11 `not-run:review-absent` (the script's NOT RUN class; the `review` step ends in failure on it, never success, AC-CCR-021); the merge is withheld and never authorized | AC-CCR-007 (`empty_reviews_pending`, `expiry_never_merges`), AC-CCR-008 (`review_absent_window_closed_exit_11`, `not_run_message_review_absent_names_cause`), AC-CCR-003 (forged artifacts do not fill the gap), AC-CCR-019 (`review_absent_expiry_zero_merge_calls`) | A script that reads an empty list as clean; a loop that merges on expiry; an absent review left at 10 or read as success |
| Missing output | A review of the identity lacking `commit_id`, `state`, `id`, or a submission time: exit 1 (`review-entry-invalid`, `review-not-complete:no-submitted-at`); a missing input file: `input-invalid`; a list that cannot be read at the end of the window: failure `review-unreadable` | AC-CCR-004 (`commit_id_missing`, `state_missing`, `id_missing`, `approved_submitted_at_*`, `*_file_missing`), AC-CCR-007 (`unreadable_to_end_of_window_fails_review_unreadable`) | A script that treats a missing field or file as pass |
| Failed review | A bound review that is `COMMENTED` (whatever its body), `DISMISSED`, `PENDING`, `CHANGES_REQUESTED` without a later APPROVED, or accompanied by a finding: exit 1 | AC-CCR-004, AC-CCR-005 | A script that treats those states as clean (the revision-3 reader) or ignores the comments |
| Stale SHA | The identity's review is bound to a commit other than the head (a push after the pull request opened): exit 1 `stale-review` | AC-CCR-002 | A script that ignores `commit_id` or compares a prefix |

## 4. Edge cases

- Re-review of the same head: the newest bound review (highest `id`) decides and must be APPROVED; an earlier
  CHANGES_REQUESTED is superseded only by a later APPROVED (AC-CCR-001 `latest_bound_review_decides_approved`,
  `changes_requested_then_approved_clean`; AC-CCR-004 `changes_requested_then_commented_refused`,
  `approved_then_changes_requested_refused`, `approved_then_commented_refused`).
- A push after the pull request opened: the review runs only at open, so the moved head has no bound review →
  `stale-review`, withheld at once (AC-CCR-002, AC-CCR-007); the old review belongs to the old commit and is
  never read as current. A fresh review is the leader-coordinated step DS-3 (spec §5).
- A forged review, comment, or approval text by another author or app, or a lookalike of the identity: never
  counted; with no real review the read stays pending and the window ends in a failed step `not-run:review-absent`
  (AC-CCR-003, AC-CCR-008).
- A gate-path edit past the first 100 changed files, a list API failure mid-pagination, a truncated or shifted
  list, a collected count different from the pull request's own count, and an entry of a head that moved: the
  merge is withheld (AC-CCR-010, AC-CCR-009, AC-CCR-012).
- A pull request that adds or changes any workflow file or the gate script: withheld for a human (AC-CCR-009).
- Draft pull request: `not-run:draft`, step exit 11, job result never success, withheld at once, whatever
  artifacts exist; its observation is `remote observed: gate path NOT RUN (not-run:draft)` at most, never
  `remote verified` (AC-CCR-008, AC-CCR-022, design §12). A non-draft pull request with no review at the end
  of the window is a failed `review` step with cause `not-run:review-absent`, observed as `remote observed:
  review step failed, review absent` at most (AC-CCR-021).
- A failure in the `files` step (a missing gate script, a gate-path change, a list that cannot be proven
  complete) ends that step and the job non-zero and skips the `review` and merge steps (AC-CCR-025); a
  failure in `review` ends it non-zero (AC-CCR-020); the earlier steps `pr`, `head`, `checks` keep the exit-0
  withhold (Out of Scope).
- The native review reports findings only in a review body, or a clean result as a COMMENTED review, a
  comment, or a check run: never read as a finding or as a pass; a COMMENTED review is never clean (AC-CCR-004),
  documented residual (design §3.2, §6, U5), observed at DS-1.
- An APPROVED review with no finding comment is a pass; its body is never read, so an APPROVED review whose
  body says otherwise also passes: a documented residual risk, not a testable property (design §6).
- The reviews list silently shortened by the API with no error: not detectable (U9, design §6 residual 3).

## 5. Quality gates

- `go vet ./internal/template/ ./internal/config/ ./internal/guardstate/` and `gofmt -l` on the new `_test.go`
  files report nothing.
- `actionlint .github/workflows/auto-merge.yml` exits 0 locally at M3 (a local evidence command; not wired into
  CI by this SPEC).
- `moai spec lint SPEC-CI-CODEX-REVIEW-001` and `--strict` exit 0 at plan close.
- Existing tests re-measured, name-exact, because a milestone's edits can reach them (research E19):
  `^TestAutoMergeRequiredChecks$` (`internal/template`, M3); the three `TestCensus_*` tests
  (`internal/guardstate`, M4); `^TestNoHardcodedContexts$` (`internal/config`, M4); `^TestRequiredChecks_ReleaseGateSubsetOfMain$` is run only where the ReleaseGate field exists on the tree; it is absent from the remote base, where the guard reads the optional release_gate YAML key directly through the helper readReleaseGate in internal/config/required_checks_review_guard_test.go and the rows not_in_release_gate and control_release_gate_reader_reads_the_key carry the same forbidden condition. A baseline-red among them is recorded at pre-flight and never counted as new.

## 6. Definition of Done — two states, never merged

**State L — landed locally (reachable by this lane).**
1. All 25 criteria pass on the merged-local tree with the swept-set rule satisfied.
2. Every regression-guard mutant red is recorded verbatim in `progress.md` §E.2, and so is the red of each
   mutant named by AC-CCR-012 and AC-CCR-019, each with the subtest it turns red; the removal of the `files`
   term alone from the `review` or merge `if`, and the removal of the `steps.review.outputs.ok` term from
   the merge `if`, are structural mutants of AC-CCR-012 (`review_step_depends_on_files_ok`,
   `merge_if_requires_files_and_review_ok`, and the review and `files` condition rows; both are equivalent
   mutants under the revision-6 rules, so no AC-CCR-019 red is owed and no obligation is lost), not of
   AC-CCR-019, whose gate-path row is killed by mutant MF (trace in §2.5), recorded verbatim by the run phase's
   mutant run.
3. AC-CCR-016 holds: `git merge-base main HEAD` was read and recorded as `A`, `git diff --quiet A HEAD --
   <the AC-CCR-016 paths>` exits 0, and the control `git diff --name-only A HEAD` lists at least one path.
4. No key, secret, comment, review request, app or settings change, branch-protection call, or remote write was
   made by the run; the run record states that.
5. The final run report lists gates G1-G8 and the deferred steps DS-1, DS-2, DS-3 (from `spec.md` §5) for the
   leader's draft pull request body, states known debt KD-1 and the accepted limitations AL-1 and AL-2, and
   names the merged-local SHA on local `main`.
6. The landing re-measure of `plan.md` §F ran and its output is recorded.
7. The report states that the remote conclusion of the `files` step, the `review` step, and the job is
   UNOBSERVED: the local criteria assert step exit statuses and a derived job result only (spec §4, G6
   acceptance condition), and the remote state of the amended workflow is "unverified".
The report states "landed locally" and the merged-local SHA. It does not state "complete", and it does not use
the words "remote verified".

**State R — `remote verified: review ran and CI green for <SHA>` (NOT reachable by this lane).** For one head
commit `H` of a **non-draft** pull request, read by GET after the leader's coordination (§7): a review by the
Codex app identity bound to `H` was observed (the gate script's read of the fetched lists, or the findings
reported) AND the CI of `H` is green. A real review alone, or CI alone, is not this state, and nothing is
called complete before both are confirmed. **A draft pull request never reaches this state**; its observation
is `remote observed: gate path NOT RUN (not-run:draft)`, which verifies the gate path only. The G6 draft pull
request's CI success cannot show that the amended `auto-merge.yml` runs on the default branch (spec §5, G6
acceptance condition): until the next pull request after landing is observed, the remote state of the
workflow is "unverified". This lane makes no remote write
and reads no remote data for the leader's draft pull request, so it states neither the draft observation nor
this state. The run exit and the sync exit each say "landed locally" and "remote verified: not reached by this
lane" in those words.

## 7. Remote observation — reading procedure, not a local criterion

The ordered, leader-coordinated steps and the read-only commands are in design §12. In summary, for the
exact head commit `H` of a pull request: step A (on the G6 draft pull request, workflow-independent) reads the
Codex-artifact GETs to learn whether anything appears on a draft (DS-1); step B (on the next pull request after
landing, a draft) reads the pull request object (`draft: true`), the Auto Merge run's job conclusion and its
`NOT RUN: draft` log line (the conclusion is UNOBSERVED until then); and the Codex-artifact GETs (`pulls/<N>/reviews`,
`pulls/<N>/comments`, `issues/<N>/comments`, `commits/<H>/check-runs`) to learn whether anything appears on a
draft (U4); step 2 (ready-for-review) re-reads the same GETs and records what appeared, under which author
(`user.login`, `user.id`, `user.type`), bound to which `commit_id`; step 3 reads the CI of `H`. Only with a
bound review observed AND CI green for the same `H` is the wording `remote verified: review ran and CI green
for <SHA>` used. A moved head is `stale-review` until a fresh review is requested (DS-3, an action outside this
lane).

| Observation | Wording that may be used |
|-------------|--------------------------|
| Draft: the pull request object says `draft: true` and the auto-merge run for its head logged `NOT RUN: draft` | `remote observed: gate path NOT RUN (not-run:draft)` |
| Draft: no auto-merge run exists for the head | `gate path not observed` (a gap, not a pass) |
| Non-draft: a review of the identity bound to `H`, no finding, and CI of `H` green | `remote verified: review ran and CI green for <SHA>` |
| Non-draft: a review bound to `H` with findings, or CI not green | `remote observed: review ran with findings` or `CI not green` — never verified |
| Non-draft: a review bound to `H` that is not APPROVED (COMMENTED, DISMISSED, PENDING, CHANGES_REQUESTED) | `remote observed: review present, not clean (<state>)` — never a gate pass; the review ran, so the CI-and-review wording above is still the only route to `remote verified` |
| Non-draft: no review of the identity, or one bound to another commit | `remote observed: review step failed, review absent` or `remote observed: stale-review` — never verified |
