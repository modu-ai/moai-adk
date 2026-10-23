# acceptance.md — SPEC-CI-CODEX-REVIEW-001

Verification layer. Requirements live in `spec.md` §3 (GEARS); every entry here is a Given-When-Then
criterion, binary-testable, labelled `AC-CCR-NNN`. Evidence ids R* are the RED-now ledger below; E* and U*
are in `research.md`. Plan revision 4 keeps the criterion set of revision 3 and adds one criterion (the merge
decision path): 19 criteria for 17 requirements, contiguous. **Plan revision 6** (an in-place amendment, spec
`## Amendments`; delta in `design.md`) changes four criteria (AC-CCR-007, 008, 012, 019) and adds six
(AC-CCR-020 to AC-CCR-025): 25 criteria for 21 requirements, contiguous, at the ceiling of 25. Revision 6
criteria carry their own RED-now pin, tree `0955fb9058` (ledger R13 to R17). **Plan revision 7** (an in-place amendment, spec `## Amendments` Amendment 2; delta in `design.md`) rewrites
AC-CCR-001 to 008, 010, 012, 019 to 023 to the comment contract with unchanged identifiers: still 25 criteria for 21 requirements, none
added and none removed (ceilings 25 and 25). The revision-7 criteria carry the RED-now pin of tree `b2af3db9a1` (ledger R18 to R23). **Plan revision 8** (the correction after plan audit iteration 1 of revision 7) re-derives the class lists from the cells that
reproduce on the publication tree, splits the job-result groups of AC-CCR-019, and adds rows (recheck race, security-marker order, the `gh pr checks`
stub): still 25 criteria for 21 requirements. The
iteration-1 identifiers map to the new ones
in `.moai/reports/t1868/progress.md`; the keyed-action criteria (key presence, workflow triggers, pins,
verdict job, fork notice job, check-run binding) no longer exist because no such workflow is built.
**Plan revision 12** (the run-phase correction for exact-head security finding #4316515155; delta in `design.md`) updates
AC-CCR-012's recheck contract without adding or removing criteria: the fresh PR issue-comment count anchors the
complete recheck list and its closing PR read, while a count increase since the earlier clean pass alone is allowed.
**Plan revision 13** (Amendment 6, the operator and leader ruling of 2026-10-23, decision-index Q30) adds five lettered sub-criteria
inside existing logical criteria — AC-CCR-006a (the `auto-merge` status and `statuses: write`), AC-CCR-013a (the `issue_comment`
trigger and its job-level author gate), AC-CCR-013b (no pull-request code on the comment path), AC-CCR-013c (the same chain, one
evaluation, head re-confirmed, one concurrency group), AC-CCR-019a (non-zero held merge outcomes) — in §2.7, and amends AC-CCR-006,
007, 012, 013, and 019 where they pinned the old trigger, the old permission set, or an exit-0 merge hold. The logical criterion count
stays 25, the Tier L ceiling; the sub-criteria are pairs within one logical criterion each (the lettered-suffix convention of
acceptance bodies) and each carries its own two cells. Their RED-now pin is tree `41b7f29b7a` (ledger R28 to R35).

**Document pin.** Every RED-now observation below was re-measured on tree `e11af67632` (full
`e11af676320604bb1f7cf07f14aa457b3437a9ec`, branch `WT-ci-codex-review`, clean) on 2026-10-18 and binds every
criterion that carries no pin of its own. That tree differs from the card base `4f9dba14bf` only under
`.moai/specs/SPEC-CI-CODEX-REVIEW-001/` and `.moai/reports/t1868/` (`git diff --name-only 4f9dba14bf HEAD`
lists 10 paths, all under those two directories). The preserve proof of AC-CCR-016 is measured from the merge
base of `HEAD` and local `main`, read at measurement time (on this tree it prints
`4f9dba14bf2834aa08bea79f9e72c2a7e75ab18e`). A criterion-level pin, where present, wins. Cells R1 to R17 describe trees from before M1 to M5 landed and do not reproduce on the publication tree
("Ledger status" below); a criterion binds a RED-now cell only if that cell reproduces on a tree the criterion is audited on.

**Classes (re-derived in revision 8, AC-CCR-022 moved in revision 9).** *release-blocking* (12 criteria: AC-CCR-001 to 008, 010, 012, 019, 020): each carries at least one RED-now cell that
reproduces on the publication tree (R18 to R24, R26, R27) as a re-executable single-invocation command with stdout, exit code, and the tree pin, and a green path
that names the milestone that flips it (M6 for all of them). *regression-guard* (13 criteria: AC-CCR-009, 011, 013, 014, 015, 016, 017, 018, 021, 022, 023, 024, 025):
green on this tree by nature (M2 to M5 landed, so their older cells no longer reproduce, or the only red is a renamed row), so no RED-now exists; each names a mutant that makes it red, and the
run phase must record that red verbatim; no plan revision re-observed those mutant reds (a stated Gap). A regression-guard is never recorded as a release gate
(verification-completeness §2.1, the undecidable disposition).
**Revision 13 sub-criteria:** *release-blocking* AC-CCR-006a, 013a, 013c, 019a (RED-now R28 to R35 on tree `41b7f29b7a`, green path
M7); *regression-guard* AC-CCR-013b (the checkout already has no `ref` and no payload text is read, so no RED exists; its adoption
evidence is the mutant red M7 records).

**Swept-set rule.** Every Go-test criterion names its subtests. A run counts only when `-v` output lists
each named `--- PASS:` line; a green with `[no tests to run]` or fewer named lines is a Gap, never a pass.
The selectors are name-exact (`^...$`).

**Fixture shape.** Revision 7: the reader fixtures are modelled on the real REST objects read on 2026-10-19 (`design.md` "Real-sample
provenance"; `research.md` E37 to E46). An issue-comment fixture carries `id`, `user.{login,id,type}`, `body`, `created_at`,
`updated_at`; its body begins with the leading block of three marker lines (the summary marker, the state marker, the security marker),
followed by prose and a two-row table; the state marker carries a `code_review` and a `security_review` object with `commit_sha`,
`has_any_findings`, `has_unresolved_findings`, `status`; the security marker carries `blockingSeverityThreshold`, `headSha`,
`mergeGateEnabled`, `pullRequestNumber`, `repository`, `status`. A review carries `id`, `user.*`, `commit_id`, `state`, `submitted_at`;
a review comment carries `id`, `user.*`, `commit_id`, `original_commit_id`; a changed-file entry carries `filename`, `status`, and
`contents_url` with `?ref=<commit>` (E31). "The identity" below is login `chatgpt-codex-connector[bot]`, user id `199175422`, type `Bot`
(E27, E37). A fixture of the identity is the only artifact a test can make count. In the fixture names, `H` is the head under review
and `O` another commit. No fixture quotes more than the marker lines and the table rows of a real comment.

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

**Revision 7 ledger (tree `b2af3db9a1`, full `b2af3db9a1a7c3554110a3b4bd0dc588ed6d29ac`, branch `WT-ci-codex-publish`, clean).** Each
command is a single read-only invocation, run on this tree on 2026-10-19; a cell's pin wins over the document pin. Why each is red for the
right reason: the shipped reader and workflow lack the revision-7 feature that M6 adds, and no pre-existing unrelated file keeps them red.

```
R18 $ grep -c -F 'codex-review-state' scripts/ci/codex-review-gate.sh
    stdout: 0
    exit: 1   (the shipped reader knows no state marker)
R19 $ grep -c -F 'issues/' .github/workflows/auto-merge.yml
    stdout: 0
    exit: 1   (the workflow collects no issue-comment listing)
R20 $ grep -n -F 'APPROVED' scripts/ci/codex-review-gate.sh
    stdout: 16:#   0  pass      the newest native review bound to the head is APPROVED, no finding exists
            56:| ($bound | map(select(.state == "APPROVED") | .id) | max) as $ap
            73:  elif $n.state != "APPROVED" then {c: 1, r: ("review-not-complete:" + $n.state)}
            76:  else {c: 0, r: ("native-review-clean: review " + ($n.id | tostring) + " state APPROVED bound to " + $head[0:10])}
    exit: 0   (the shipped pass condition is an APPROVED review object)
R21 $ grep -c -F 'issues/' scripts/ci/codex-review-gate.sh
    stdout: 0
    exit: 1   (the `fetch-pages` allow-list has no issue-comment endpoint)
R22 $ grep -c -F -e '--delete-branch' .github/workflows/auto-merge.yml
    stdout: 1
    exit: 0   (the merge call requests branch deletion)
R23 $ grep -c -F '6458969' .github/workflows/auto-merge.yml
    stdout: 0
    exit: 1   (the workflow names no eligible author)
```

**Ledger status (revision 8; observed 2026-10-19 on the publication tree at `e9ed17271c`, by `grep` and `test` only, no Go run).** The cells of R1 to R17 were true on
the trees named for them and no longer reproduce, because M1 to M5 landed. Observed now: R1 exit 0 (cell: 1); R2 `0`, exit 1 (cell: 14); R3 `0`, exit 1 (cell: 1); R4 `1`, exit 0
(cell: 0, exit 1); R5 `3`, exit 0 (cell: 0, exit 1); R6 `5`, exit 0 (cell: 0, exit 1); R7 empty, exit 1 (cell: two lines); R8 `65:    enabled: false` and `74:  auto_reply: false` (cell:
true); R10 both `test -f` exit 1 (cell: exit 0); R11 `0`, exit 1 (cell: 2); R12 `1`, exit 0 (cell: 0, exit 1); R13 and R15 empty, exit 1 (cells: three lines); R16 `1`, exit 0 (cell:
0, exit 1); R17 `0`, exit 1 (cell: 1, exit 0). R9 and R14 were not re-run by this revision (R9 needs a git comparison against a moving tip, R14 prints other lines per the plan audit): a stated Gap.
They are history and bind no criterion. R18 to R23 were measured on `b2af3db9a1` and re-executed by the plan audit on `e9ed17271c` (the two trees differ only under this SPEC
directory); each reproduces.

```
R24 $ grep -c -F 'unresolved_findings_step_exit_nonzero_job_failure' internal/template/auto_merge_codex_test.go
    stdout: 0
    exit: 1   (the renamed AC-CCR-020 row is absent; M6 adds it)        tree e9ed17271cc743a78dc1e5b5d1d87b6b9a2f89ce
R25 $ grep -c -F 'draft_with_clean_state_comment_still_not_success' internal/template/auto_merge_codex_test.go
    stdout: 0
    exit: 1   (a test-addition cell, not a class cell: AC-CCR-022 is a regression-guard since revision 9; the renamed row is absent and M6 adds it)        tree e9ed17271cc743a78dc1e5b5d1d87b6b9a2f89ce
R26 $ grep -c -F 'clean_state_comment_f_clean' internal/template/codex_review_gate_test.go
    stdout: 0
    exit: 1   (the rewritten AC-CCR-001 reader row is absent)           tree e9ed17271cc743a78dc1e5b5d1d87b6b9a2f89ce
R27 $ grep -c -F 'author_step_present_after_head_before_checks' internal/template/auto_merge_codex_test.go
    stdout: 0
    exit: 1   (the author-step rows of AC-CCR-012 are absent)           tree e9ed17271cc743a78dc1e5b5d1d87b6b9a2f89ce
```

**Revision 13 ledger (tree `41b7f29b7a`, full `41b7f29b7aa4d9d6717dc8fbff38058561d81ebd`, branch `WT-codex-native-gate`, clean;
measured 2026-10-23 by manager-spec, each a single read-only invocation, exit code observed separately).** Why each is red for the
right reason: `auto-merge.yml` as landed has the `workflow_run` trigger alone, no comment path, no status write, and a merge step whose
held outcomes end exit 0; M7 adds exactly those, and no pre-existing unrelated file keeps any cell red.

```
R28 $ grep -c 'issue_comment' .github/workflows/auto-merge.yml
    stdout: 0
    exit: 1   (no issue_comment trigger: a late review is never re-evaluated, F4)
R29 $ grep -c -F 'github.event.comment.user.id' .github/workflows/auto-merge.yml
    stdout: 0
    exit: 1   (no job-level author gate on a comment)
R30 $ grep -c -F 'github.event.issue.number' .github/workflows/auto-merge.yml
    stdout: 0
    exit: 1   (neither a comment-path PR number nor a concurrency key for it)
R31 $ grep -c -F 'github.event_name' .github/workflows/auto-merge.yml
    stdout: 0
    exit: 1   (no event discrimination: the job condition and the review loop know one path)
R32 $ grep -n -A1 -F 'summary "merge not executed: head moved' .github/workflows/auto-merge.yml
    stdout: 470:            summary "merge not executed: head moved - PR #$PR_NUMBER left open, tested head ${TESTED_SHA:-<empty>}, current head ${HEAD_NOW:-<empty>}"
            471-            exit 0
    exit: 0   (a merge reached and not executed ends exit 0, F3/F1)
R33 $ grep -n -A1 -F 'summary "merge held: review window elapsed before merge' .github/workflows/auto-merge.yml
    stdout: 460:            summary "merge held: review window elapsed before merge - PR #$PR_NUMBER left open, head $TESTED_SHA"
            461-            exit 0
    exit: 0   (the deadline hold ends exit 0)
    $ grep -n -A1 -F 'summary "merge failed:' .github/workflows/auto-merge.yml
    stdout: 512:            summary "merge failed: gh pr merge exit $merge_rc (${MERGE_REASON:-no output}) - PR #$PR_NUMBER left open, head $TESTED_SHA"
            513-          fi
    exit: 0   (a failed merge call falls through to the step's end, whose last command is the summary echo: exit 0)
R34 $ grep -c -F 'statuses' .github/workflows/auto-merge.yml
    stdout: 0
    exit: 1   (no commit status is written)
R35 $ grep -c -F 'statuses: write' .github/workflows/auto-merge.yml
    stdout: 0
    exit: 1   (the permission is absent)
B1  $ git diff --quiet 8c36a7904bfe3e44c46b154a2839a6aedb719ef2 HEAD -- .github/required-checks.yml
    stdout: (empty)
    exit: 0   (baseline, not a RED: required-checks.yml equals origin/main, read as 8c36a7904b after `git fetch origin main`
               on 2026-10-23; a regression-guard row that must stay exit 0)
```

Tool provenance (verification-claim-integrity §2.2): the cells are `grep` and `git` measurements, not moai-tool output. The
`moai spec lint` run of this revision used the installed build `b1806498fe` (v3.2.0-rc.38); `git merge-base --is-ancestor
b1806498fe HEAD` exited 1, so that build is not a strict ancestor of the measured HEAD.

Supporting behavior, not a cell (the delegated Codex observer's reproduction `/tmp/adk-queue-repair-20261018/review-step-repro-v2/`,
the observer's `/tmp/adk-ci-followthrough-20261018/files-step-missing-script-repro/result.json`, and this lane's
re-run in its scratchpad, all machine-local; the deciding lines are carried in `research.md` E34 and E36):
changes-requested `exit 0`, outputs `ok=false state=fail`; draft `exit 0`, outputs `ok=false state=fail state=not-run`;
clean `exit 0`, outputs `ok=false state=fail ok=true state=pass`; `files` step without the script `exit 0`, outputs
`ok=false state=fail`.

R1 to R11 were re-run on `e11af67632` in revision 4 and print exactly what they printed on `ae494aedb2`
(revision 3); R12 is new in revision 4.

## 2. Criteria

### 2.1 Native review verification (M1; revision 7: the comment contract, M6)

Revision 7 rewrites AC-CCR-001 to AC-CCR-005 for the state comment (spec §3.1); identifiers and the count of criteria are
unchanged. Fixture families `F-*` and their real counterparts are in `design.md` "Real-sample provenance". The author object of
every identity fixture is the real one: login `chatgpt-codex-connector[bot]`, id `199175422`, type `Bot`. Revision-7 criteria carry
the RED-now pin of tree `b2af3db9a1` (ledger R18 to R23).

**AC-CCR-001** — REQ-CCR-001 — release-blocking
- Given the issue comments, reviews, and review comments of a pull request in which the identity's single state comment carries, in
  the leading block, a state marker with both legs `completed`, each `commit_sha` equal to the head, each
  `has_any_findings` and `has_unresolved_findings` both the boolean `false`, and a security marker `completed` with `headSha` equal to the head (`F-clean`,
  modelled on #1835, #1837, #1838, #1841), and no corroborating finding; and noise around it: comments of other authors
  (`vercel[bot]`, `coderabbitai[bot]`, `github-actions[bot]`, a human) carrying any text, a copy of a clean marker among them; an
  APPROVED, DISMISSED, or PENDING review of the identity; a finding bound only to an older head; prose and table rows saying
  anything; the summary marker absent or altered,
- When `read-native-review` runs with `--draft false` and either `--wait-closed` value,
- Then it exits 0 and prints `native-review-clean: comment <id> updated <updated_at> ...` (design §3.2 step 14); the noise changes
  nothing, because the decision reads the author, the marker position, the marker fields, and the corroborating artifacts only.
- Verify: `set -o pipefail && go test -count=1 -v -run '^TestCodexNative1828ReceiptTypesAndHeadBinding$' ./internal/template/ | grep -F -- '--- PASS: TestCodexNative1828ReceiptTypesAndHeadBinding '` must print the selected test's top-level PASS; Go's zero-test warning alone cannot satisfy this check.
  The selected test exercises current-head code/security receipt pairing, observed clean declarations, unresolved-summary refusal, untrusted marker handling, and malformed, stale, missing, wrong-type, duplicate, and ambiguous receipt rejection.
- Mutant: an unconditional pass is rejected by `actual_unresolved_review_remains_blocked`, `missing_security_receipt_is_rejected`,
  and `forged_code_receipt_identity_is_rejected`; an unconditional refusal is rejected by `security_receipt_newest_actual_clean_text`
  and `untrusted_marker_only_does_not_block_valid_native_receipts`. The selected clean path runs with empty formal-review and
  review-comment inputs, so an APPROVED review object cannot grant a pass.
- RED-now: R18 and R20 (tree `b2af3db9a1`): the shipped reader knows no state marker and requires APPROVED. Green path: M6 flips
  them; the new reader passes `F-clean`.

**AC-CCR-002** — REQ-CCR-002 — release-blocking
- Given state comments whose legs are bound to a commit other than the head: both legs (the #1836 comment against its saved final
  head `4cee5eea1c`, `F-stale`), the code leg only, the security leg only, the security marker's `headSha` only, a prefix of the head,
  an empty `commit_sha`, legs that are still `running` for an older commit, the re-request state of #1842 (`F-mixed-sha`: the code leg bound to the new head, the security leg null, the security marker naming the old
  head `45b2dbf682` with status `running`), the move of #1842 (a state recorded for
  `45b2dbf682515dd8f4d8d67ce0ae7ac2fa8328a5`, the head read as `e7cceb2cb5d6e1cd9dbd2f465c4b9d5a1d631f21`, `F-moved-head`), and a forged
  clean comment beside a real stale one,
- When `read-native-review` runs, with `--wait-closed false` and with `--wait-closed true`,
- Then it exits 1 with reason `stale-review:<commit prefix>`, never 0, never pending, and never NOT RUN.
- Verify: `go test -count=1 -v -run '^TestCodexGateReadNativeReviewRefusesStale$' ./internal/template/` lists `--- PASS` for
  `both_legs_bound_to_older_head_f_stale`, `markers_with_different_shas_f_mixed_sha_refused`, `code_leg_only_older`, `security_leg_only_older`, `security_marker_head_sha_older`,
  `commit_sha_prefix_only`, `commit_sha_empty_string`, `running_legs_bound_to_older_head_stale`,
  `forged_clean_comment_beside_real_stale_comment_still_stale`, `old_head_state_never_counts_for_new_head_f_moved_head`,
  `stale_window_closed_still_exit_1`.
- Mutant: a script that ignores `commit_sha`, compares a prefix, checks only the code leg, treats a stale state as pending, or maps
  a stale state to NOT RUN when the window is closed turns the matching row red.
- RED-now: R18. Green path: M6.

**AC-CCR-003** — REQ-CCR-003 — release-blocking
- Given artifacts that look like the state comment but are not authored by the identity, each carrying a clean marker (`F-forged`): a
  comment of `coderabbitai[bot]` (id 136622811), a human's comment, the same login without the `[bot]` suffix and the same id, the same
  login with an id other than 199175422, the same login with type `User`, and a comment whose `user` is null; and a review and a
  review comment of another author on the head,
- When `read-native-review` runs,
- Then none counts: with no real comment the result is exit 10 `pending:no-state-comment` (window open) or exit 11
  `not-run:review-absent` (window closed), never 0; a forged comment beside the identity's real `running` comment leaves the result
  pending; beside the real stale comment it leaves `stale-review`; beside the real clean comment it leaves exit 0 and is not read as a
  second state comment; and another author's review or review comment on the head neither blocks nor completes anything. The exclusion
  is by author only: the identity's own clean comment is not excluded by what its prose says.
- Verify: `go test -count=1 -v -run '^TestCodexGateReadNativeReviewIgnoresForged$' ./internal/template/` lists `--- PASS` for
  `other_bot_clean_marker`, `human_clean_marker`, `lookalike_login_without_bot_suffix_same_id`, `same_login_id_other_than_199175422`,
  `same_login_type_user`, `user_null`, `marker_text_in_user_comment_no_real_comment_pending`,
  `forged_beside_real_running_still_pending`, `forged_beside_real_stale_still_stale`,
  `forged_beside_real_clean_not_duplicate_exit_0`, `other_author_review_and_review_comment_ignored` (modelled on the CodeRabbit review 5567249506 and its inline comment on #1842).
- Mutant: a script that matches the login only turns `same_login_id_other_than_199175422` and `same_login_type_user` red; one that
  matches the id only turns `lookalike_login_without_bot_suffix_same_id` red; one that accepts any author turns
  `other_bot_clean_marker` red; one that selects candidates by marker text without the author filter turns
  `marker_text_in_user_comment_no_real_comment_pending` and `forged_beside_real_clean_not_duplicate_exit_0` red (the latter reads the
  forgery as a second state comment).
- RED-now: R18. Green path: M6.

**AC-CCR-004** — REQ-CCR-004 — release-blocking
- Given state comments of the identity in each of these forms: a leg `status` of `cancelled`, an unknown string, null, or
  missing; a leg `status` of `failed` (a review run that failed) with `has_unresolved_findings` true or false and the other leg completed, `running`, or
  failed, bound to the head, bound to an older head, or with a null `commit_sha` (`F-failed`, modelled on #1844; and the composite code `failed` with the
  security leg and marker `running`); a security leg whose `status`, `commit_sha`, and flags are all null beside a completed code leg (`F-null-security-leg`, modelled
  on #1839 and #1840); the `code_review` leg or the `security_review` leg missing; a completed leg whose `has_any_findings` is null,
  missing, the string `"false"`, or `0`; a completed leg whose `has_unresolved_findings` is null
  (the #1842 form), missing, the string `"false"`, or `0`; a completed leg whose `commit_sha` is null beside a security marker that
  reads `completed` (the #1842 form with the security marker `completed`); a security marker missing while no leg is `running`, duplicated, outside the
  leading block, or malformed, and one that reads a status other than `completed` while the legs are completed; marker JSON altered
  (truncated, text after the object, a key of the wrong type); the state marker twice in one body; the state marker quoted in the
  identity's own prose beside the real one, and in the prose only; two comments of the identity carrying the marker, one of them clean;
  a candidate without an `id`; a missing input file; and a line that is not an object,
- When `read-native-review` runs with `--wait-closed false` and `true`,
- Then every row exits 1 with its own named reason (design §3.2) and none exits 0: `review-failed:<leg>` (checked before any finding flag is read, so a failed leg's true flags never read as `finding:unresolved`), `review-not-complete:<leg>:<status>` and `review-not-complete:<leg>:<field>`,
  `review-not-bound:<leg>`, `security-marker-missing`, `-duplicated`,
  `-misplaced`, `-malformed`, `security-marker-disagrees:<status>`, `state-marker-malformed`, `state-marker-duplicated`,
  `state-marker-misplaced`, `state-comment-duplicated:<ids>`, `state-comment-invalid:<id>`, `input-invalid`. The gate never chooses
  between two state comments by recency, and `completed` alone, or both legs reading `completed`, never passes.
- Verify: `go test -count=1 -v -run '^(TestCodexGateReadNativeReviewRefusesIncomplete|TestCodexNative1828ReceiptTypesAndHeadBinding)$' ./internal/template/` lists `--- PASS` for
  `status_failed_review_failed_reason_f_failed`, `status_failed_security_running_refused_at_once`, `status_failed_security_leg_review_failed`,
  `status_failed_null_commit_sha_review_failed`, `status_failed_older_head_stale`, `status_failed_never_read_as_finding_nor_pass`, `status_cancelled_refused`, `status_unknown_string_refused`, `status_null_refused`, `status_missing_refused`, `security_leg_all_null_refused_f_null_security_leg`,
  `code_review_leg_missing_refused`, `security_review_leg_missing_refused`, `completed_with_has_unresolved_null_refused`,
  `completed_with_has_unresolved_missing_refused`, `completed_with_has_unresolved_string_false_refused`,
  `completed_code_missing_has_any_findings_refused`, `completed_security_missing_has_any_findings_refused`,
  `completed_code_null_has_any_findings_refused`, `completed_security_nonboolean_has_any_findings_refused`,
  `completed_leg_commit_sha_null_security_marker_completed_refused`, `completed_legs_without_security_marker_refused`,
  `security_marker_duplicated_refused`, `security_marker_misplaced_refused`, `security_marker_malformed_refused`,
  `security_marker_disagrees_legs_completed_marker_failed_refused`, `marker_json_truncated_refused`,
  `marker_json_trailing_text_refused`, `marker_key_wrong_type_refused`, `state_marker_duplicated_in_one_body_refused`,
  `state_marker_in_identity_prose_beside_real_refused`, `state_marker_in_identity_prose_only_refused`,
  `two_state_comments_one_clean_refused`, `receipt_quoting_marker_beside_real_refused` (a second comment of the identity that carries the state
  marker prefix), `candidate_without_id_refused`, `input_files_missing_refused`,
  `input_line_not_object_refused`.
- Mutant: **the revision-6 reader** (an APPROVED review object decides) turns the whole family red or vacuous; a reader that
  passes on `status` alone turns `completed_with_has_unresolved_null_refused` and
  `completed_leg_commit_sha_null_security_marker_completed_refused` red (null read as false); one that reads only the
  `code_review` leg turns `security_review_leg_missing_refused` red; one that reads the flags of a failed leg turns `status_failed_never_read_as_finding_nor_pass` red (it reports `finding:unresolved:code_review`); one that
  treats `failed` as pending and waits turns `status_failed_security_running_refused_at_once` red; one that lets a failed leg pass beside a completed leg turns `status_failed_review_failed_reason_f_failed` red; one that picks the newest of two comments turns
  `two_state_comments_one_clean_refused` red; one that scans the whole body for the marker without a position turns
  `state_marker_in_identity_prose_only_refused` red; one that ignores the security marker turns the six `security_marker_*` rows red.
- RED-now: R18 and R20. Green path: M6.

**AC-CCR-005** — REQ-CCR-005 — release-blocking
- Given a clean state comment of the identity plus, in turn: a completed leg with `has_any_findings` true and
  `has_unresolved_findings` false (`F-reported`, for code, security, and both legs); a leg with
  `has_unresolved_findings` true (`F-unresolved`, modelled on
  #1836), a review comment of the identity whose `commit_id` equals the head, whose `original_commit_id` equals the head, or that
  carries neither field, a COMMENTED review of the identity bound to the head (`F-corrob`, modelled on #1836 review 5566687300), and a
  CHANGES_REQUESTED review of the identity bound to the head; and, as controls, a comment of another author on the head, a review
  comment or review bound only to an older head, an APPROVED, DISMISSED, or PENDING review of the identity, and empty lists,
- When `read-native-review` runs,
- Then the first group exits 1 with `finding:reported:<leg>`, `finding:unresolved:<leg>`, `finding:<id>`,
  `finding-unbound:<id>`, or `finding-review:<id>`; a completed reported finding refuses before a pending peer leg,
  and unresolved findings still refuse while the state comment is running; the controls leave a clean comment at exit 0; and no review or review comment can
  turn a refusal into a pass.
- Verify: `go test -count=1 -v -run '^(TestCodexGateReadNativeReviewRefusesFinding|TestCodexNative1828ReceiptTypesAndHeadBinding|TestCodexNativeRunningSecurityLegWaits)$' ./internal/template/` lists `--- PASS` for
  `unresolved_code_leg_refused_f_unresolved`, `unresolved_security_leg_refused`, `failed_leg_true_flags_are_not_a_finding`, `review_comment_commit_id_equals_head_refused`,
  `review_comment_original_commit_id_equals_head_refused`, `review_comment_without_any_commit_field_refused`,
  `commented_review_bound_to_head_refused_f_corrob`, `changes_requested_review_bound_to_head_refused`,
  `finding_refused_while_running`, `findings_f_findings_1833_refused` (the real #1833 state with the COMMENTED review 5567280902 bound to the
  head), `completed_code_reported_finding_with_unresolved_false`,
  `completed_security_reported_finding_with_unresolved_false`,
  `completed_both_legs_report_findings_with_unresolved_false`,
  `completed_code_finding_refuses_before_running_security_wait`, `review_comment_by_other_author_ignored`, `finding_bound_to_older_head_ignored`,
  `approved_dismissed_pending_review_neither_pass_nor_refuse`, `comments_and_reviews_files_empty_clean`.
- Mutant: a script that never reads the review comments or reviews, counts artifacts of any author, treats an unbound comment as no
  finding, or reads `has_any_findings` true or `has_unresolved_findings` true as clean turns the matching row red.
- RED-now: R18 and R20. Green path: M6.

### 2.2 Auto-merge consumer (M2, M3)

**AC-CCR-006** — REQ-CCR-006 — release-blocking
- Given the auto-merge workflow and the gate script,
- When their text is scanned and the script's `fetch-pages` is run against disallowed endpoints,
- Then the `gh` calls are exactly the allow-list below and nothing else, neither file contains a forbidden token, `issues/` occurs
  only in the issue-comment listing, and `fetch-pages` refuses every endpoint outside the four listings.
  **Allow-list (exact), classified by `gh` group and subcommand and by the first `repos/...` path argument:**

  | Where | Allowed form | Used by |
  |-------|--------------|---------|
  | workflow | `gh pr view "$PR_NUMBER" --json headRefOid …` | steps `head` and merge |
  | workflow | `gh pr checks "$PR_NUMBER" --required --json name,state,bucket` | step `checks` (the unchanged CI observation; it reads check state for the CI gate, never a review verdict) |
  | workflow | `gh pr merge "$PR_NUMBER" … --match-head-commit …` with no `--delete-branch` (REQ-CCR-021) | merge step |
  | workflow | `gh api "repos/$REPO/pulls/$PR_NUMBER"` with no further path segment | steps `author`, `files`, and `review` (pull request object: `user`, `head.repo`, `base.repo`, `draft`, `head.sha`, `changed_files`, `comments`, `review_comments`) |
  | script | `gh api "<E>?per_page=100&page=<P>"` after `E` matched `^repos/<owner>/<repo>/(pulls/<N>/(files\|reviews\|comments)\|issues/<N>/comments)$` | `fetch-pages` |

  Any other `gh` group or subcommand (`gh issue`, `gh pr comment`, `gh pr review`, `gh pr edit`, `gh run`, `gh workflow`,
  `gh label`) or any other `gh api` path fails. **Revision 13:** the workflow (never the script) also makes exactly one write
  form, `gh api --method POST "repos/$REPO/statuses/$SHA"` with the context `auto-merge` (step `status`, AC-CCR-006a), and the
  token `statuses` is allowed only in that form and in the permission key `statuses: write`; `/status`, `commits/<sha>/status`,
  and any read of `statuses` stay forbidden. **Revision 13.2:** the workflow (never the script) also makes exactly one more read
  form, `gh api "repos/$REPO/actions/workflows/ci.yml/runs?head_sha=$VERIFIED_SHA&event=pull_request&per_page=100"` (step `checks`,
  comment path only, `design.md` item 4), admitted by `gh_invocations_match_allow_list_exactly` and
  `gh_api_paths_in_workflow_are_pull_object_only`; it needs the read-only permission `actions: read`. **Forbidden tokens** in either file (with that one exception): `/status`, `statuses`, `check-runs`,
  `check_runs`, `labels`, `reactions`, `--delete-branch`, `CodeRabbit`, `Merge Risk`, `AUTOMERGE_MAX_MERGE_RISK`; no step deletes a ref (`git push --delete`,
  `gh api -X DELETE`); the token `issues/` is allowed only
  as the endpoint `issues/<N>/comments` of the issue-comment listing (revision 7 reverses its earlier blanket ban).
- Verify: `go test -count=1 -v -run '^TestAutoMergeReadsOnlyAllowedReviewArtifacts$' ./internal/template/` lists `--- PASS` for
  `gh_invocations_match_allow_list_exactly`, `gh_api_paths_in_workflow_are_pull_object_only`,
  `retained_gh_pr_forms_view_checks_merge_only`, `no_forbidden_token`, `no_coderabbit_or_merge_risk`,
  `no_risk_ceiling_variable`, `no_label_read`, `scan_swept_at_least_6_gh_invocations`; and
  `go test -count=1 -v -run '^TestCodexGateFetchPagesRefusesDisallowedEndpoint$'
  ./internal/template/` lists `--- PASS` for `issue_comments_endpoint_accepted`, `issue_other_subpath_refused` (events, reactions,
  a single comment by id), `check_runs_endpoint_refused`, `commit_status_endpoint_refused`, `path_with_extra_segment_refused`,
  `files_reviews_comments_endpoints_accepted`.
- Mutant: re-adding `--delete-branch` or a ref deletion turns `no_delete_branch_flag` or `no_ref_deletion_call` red; a `gh api` of any path other than the pull request object in the workflow, a status read, a `gh pr comment`, a reactions
  read, or an `issues/` path other than the comment listing turns `gh_invocations_match_allow_list_exactly` (or the token row) red;
  a `fetch-pages` without the endpoint check turns the refusal rows red; a scan that matches nothing turns
  `scan_swept_at_least_6_gh_invocations` red.
- RED-now: R19 and R21 (tree `b2af3db9a1`: the workflow names no issue-comment listing and the script's allow-list has none) and, for the
  deletion rows, R22 (the merge call passes `--delete-branch`). Green path: M6.

**AC-CCR-007** — REQ-CCR-007 — release-blocking
- Given a state comment with both legs `running` (`F-running`, modelled on #1841; also the both-running shape of #1833), a comment
  that stays `running` on every pass, a comment that is `running` and turns `completed` on pass k, the #1842 form (a completed leg
  with null fields beside a running security marker, `F-null-leg`), both state legs completed clean while the current-head security
  marker still reads `running`, that state with a malformed completed code SHA (null, 10-character, or 40-character non-hex), a non-draft pull request whose state comment never appears, a stale state, a refusing read, a list
  fetch that fails on every pass, a review that finishes only after the last pass, the head moving
  during the loop, an issue-comment count that moves between the two reads of one pass, and an APPROVED review of the identity beside
  a running comment,
- When the auto-merge review step's loop is read and run against a stub `gh` and a no-op `sleep`,
- Then the loop runs at most 60 passes and does not sleep after the last; a `running` leg or current-head security marker and an absent
  state comment keep waiting; the last pass passes `--wait-closed true` and every earlier pass `false`; a leg or marker still `running`
  on the last pass ends as a failure named
  `review-not-complete:running` (step exit status non-zero, job failure, REQ-CCR-018); malformed SHA shape on a completed leg still refuses with
  `review-not-complete:<leg>:commit_sha` before the running-marker pending result; a refusal (including `stale-review`)
  withholds at once; a failing fetch counts as pending, and a list unreadable on every pass ends as a failure named
  `review-unreadable`, never as NOT RUN and never as success; a moved issue-comment count retries the pass and is never a failure; an
  APPROVED review never completes a running comment; a comment that turns `completed` is merged over only after the gate exits 0 and
  the recheck agrees; a review that finishes after the window is not read by this `workflow_run` evaluation (revision 13: the
  comment-path evaluation that the identity's comment edit starts reads it, AC-CCR-013c); no merge is authorized at the end of the window. The script
  half: no state comment yet is exit 10, and a `running` comment is exit 10 while the window is open and exit 1 when it is closed,
  never 0.
- Verify: `go test -count=1 -v -run '^TestCodexGateReadNativeReviewPending$' ./internal/template/` lists `--- PASS` for
  `empty_reviews_pending`, `empty_reviews_and_comments_pending`, `security_marker_running_legs_completed_pending`,
  `security_marker_running_legs_completed_window_closed_fails`, `security_marker_failed_legs_completed_refused`; and
  `go test -count=1 -v -run '^TestCodexGateReadNativeReviewPendingRefusesMalformedCompletedSHA$' ./internal/template/` lists `--- PASS` for
  `null`, `short`, and `non_hex_full_length`; and
  `go test -count=1 -v -run '^TestCodexNativeRunningSecurityLegWaits$' ./internal/template/` lists `--- PASS` for
  `running_without_receipt_waits`, `running_with_old_security_receipt_waits`,
  `running_without_receipt_at_window_end_fails`, `running_with_old_security_receipt_at_window_end_fails`,
  `running_unresolved_still_fails`, `running_current_finding_still_fails`, `running_malformed_summary_still_fails`,
  `running_stale_summary_still_fails`, `running_peer_failure_still_fails`; and `go test -count=1 -v -run
  '^TestAutoMergeWaitLoopIsBounded$' ./internal/template/` lists `--- PASS` for `loop_bound_at_most_60`, `pending_keeps_waiting`,
  `last_pass_passes_wait_closed_true`, `no_sleep_after_last_pass`, `exit_one_withholds_at_once`, `stale_withholds_at_once`,
  `fetch_failure_counts_pending`, `unreadable_to_end_of_window_fails_review_unreadable`, `review_arriving_after_window_not_read`,
  `expiry_never_merges`.
- Mutant: a current-head security marker still `running` while both legs read `completed` is refused at once instead of waited on (`security_marker_running_legs_completed_pending` red); a malformed completed code SHA is treated as pending (`TestCodexGateReadNativeReviewPendingRefusesMalformedCompletedSHA` red); a `running` comment whose security marker is not present yet refused at once instead of waited on (the order is wait, because a pending state cannot pass; never observed, a decision of design §3.2 step 10), a `running` or absent state read as success, expiry that authorizes a merge (`expiry_never_merges` red), a refusal read as
  pending, a stale state waited on for the whole window, an unbounded `while true`, a last pass that does not pass `--wait-closed true`,
  a running comment at the window end reported as NOT RUN or as a pass, an unreadable list at expiry reported as NOT RUN, or a
  moved issue-comment count that fails the step turns the matching row red.
- RED-now: R18 (and R20: the shipped reader has no `running` class). Green path: M6.
- Run-phase regression on `18d2a6d176d4a7ba544e2f58760ff5ba3c0f4778`: `GOMAXPROCS=2 TMPDIR=/tmp go test -count=1 -run '^TestCodexGateReadNativeReviewPending$' ./internal/template` failed on the current-head running-marker case (`exit 1`, `security-marker-disagrees:running`, expected exit 10) and on the closed-window reason (`security-marker-disagrees:running`, expected `review-not-complete:running`). The repaired selector passes both cases; exact output is in `/tmp/adk-issues-prs-20261019/new-lane-progress.md`.
- Additional run-phase regression: the supplied command in `/tmp/adk-issues-prs-20261019/after-18d2/working-running-marker-malformed-reproduction.json` returned exit 10 `pending:running` for code SHA `null`, 10-character `ab609cafbf`, and 40-character non-hex `z` repeated. `TestCodexGateReadNativeReviewPendingRefusesMalformedCompletedSHA` now pins all three as exit 1 `review-not-complete:code_review:commit_sha`; valid completed SHA with a running security marker still returns exit 10 while open and exit 1 on the closed pass.

**AC-CCR-008** — REQ-CCR-008 — release-blocking
- Given a draft pull request — with no state comment, with a clean state comment bound to the head, and with no input files at all —
  a non-draft pull request with a clean state comment, and a non-draft pull request with no state comment of the identity when the
  wait window is closed and when it is open,
- When `read-native-review` runs with `--draft true` and `--draft false` and with `--wait-closed true` and `false`, and the
  auto-merge review step's handling of each exit code is read,
- Then at the script level NOT RUN is one exit class with two causes: the draft cases exit 11 with `not-run:draft` whatever artifacts
  exist; the absent-comment case exits 10 `pending:no-state-comment` while the window is open and 11 `not-run:review-absent` when it
  is closed; a stale state and an unresolved finding still exit 1 with the window closed; the non-draft clean case exits 0; an
  invalid flag exits 1; the four classes 0, 1, 10, and 11 are pairwise distinct; and the step, on 11, withholds the merge, never
  treats 11 as success, reads `draft` from the pull request object, and does not fetch the lists for a draft. What the step records
  and how it ends for each cause is `T-CHAIN` rows 4 and 5 and criteria AC-CCR-021 (non-draft, a failure: step exit 1, `state=fail`,
  `::error::` naming `not-run:review-absent`) and AC-CCR-022 (draft: `state=not-run:draft`, step exit 11, a `::notice::` that says it
  is not a review result and never a pass, job result never success).
- Verify: `go test -count=1 -v -run '^TestCodexGateReadNativeReviewDraftNotRun$' ./internal/template/` lists `--- PASS` for
  `draft_true_exit_11`, `draft_true_with_clean_state_comment_still_11`, `draft_true_without_input_files_still_11`,
  `draft_false_clean_exit_0`, `draft_flag_invalid_exit_1`, `wait_closed_flag_invalid_exit_1`, `review_absent_window_open_exit_10`,
  `review_absent_window_closed_exit_11`, `unresolved_findings_window_closed_still_exit_1`, `four_classes_pairwise_distinct`; and
  `go test -count=1 -v -run '^TestAutoMergeNotRunWithholds$' ./internal/template/` lists `--- PASS` for `exit_11_withholds_merge`,
  `not_run_message_draft_names_cause`, `not_run_message_review_absent_names_cause`, `not_run_message_never_a_pass`,
  `not_run_message_gate_path_only`, `not_run_state_output_distinct_from_fail`, `exit_11_never_reaches_merge_condition`,
  `draft_read_from_pull_request_object`, `draft_skips_list_fetch_and_is_not_run`. A draft observation, and the absence of a state
  comment, can never be read as a review result or as a pass, and each says so (design §12).
- Mutant: a script that exits 0 for a draft with a clean state comment (a draft merged), one that exits 1 for a draft or for an absent
  comment at window close (NOT RUN indistinguishable from failure), one that stays at 10 at window close (absence never final), one
  that maps a stale state to 11, a loop that proceeds on 11, a `state` output of `fail` for a draft, or a message that claims a review
  result turns the matching row red.
- RED-now: R18 (the shipped reader reads no issue comment, so the absence of a state comment is not what it decides on). Green path: M6.

**AC-CCR-009** — REQ-CCR-009 — regression-guard
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
- RED-now: none on this tree. M2 landed the script and M3 the wiring; R4 and R1 no longer reproduce (ledger status). Under verification-completeness §2.1 this criterion is a regression-guard, not a release gate; its adoption evidence is the mutant red the run phase records (not re-observed by revision 8: a stated Gap).

**AC-CCR-010** — REQ-CCR-010 — release-blocking
- Given a stub `gh` that serves a pull request's lists (changed files, issue comments, reviews, review comments) in pages of 100 by
  the `page` parameter and can fail a chosen page,
- When `fetch-pages` collects a list and the reader of that list runs on the result, with the pull request's own count passed as the
  expected count where one exists (`changed_files`, `comments`, `review_comments`),
- Then a gate-path entry on a late page is found and refused — 101 entries with the gate path on page 2 and 301 entries with it on
  page 4 both exit 1 with `gate-path-changed:<path>` — while an unrelated 350-entry list is clear; the identity's state comment on
  page 2 or page 3 of the issue comments is read (exit 0 from the reader) and a finding on page 3 of the review comments is refused; a
  failing page refuses with `page-error:<P>` and leaves no output file; an invalid page refuses with `page-invalid:<P>`; a collected
  count different from the expected count refuses with `count-mismatch:<collected>:<expected>` (issue comments included); a
  non-integer expected count refuses; a list of 3000 entries (a full 30th page) refuses with `page-ceiling`; a list whose size is an
  exact multiple of 100 ends on an empty page and is collected whole; the reviews list, which has no count on the pull request object,
  is collected with `count-proof:none`; and an entry that repeats across pages (a review or comment `id`, or a `filename`) refuses with
  `page-duplicate:<key>` and removes the output.
- Verify: `go test -count=1 -v -run '^TestCodexGateFetchPagesCompleteness$' ./internal/template/` lists `--- PASS` for
  `single_page_clear`, `exact_200_ends_on_empty_page`, `entries_101_gate_file_on_page_2_refused`,
  `entries_301_gate_file_on_page_4_refused`, `entries_350_unrelated_clear`, `page_2_error_refuses_and_removes_out`,
  `page_invalid_json_refuses`, `fewer_entries_than_expected_refused`, `more_entries_than_expected_refused`,
  `expect_count_not_integer_refused`, `ceiling_3000_refused`, `reviews_without_count_proof_three_pages`,
  `review_comments_count_mismatch_refused`, `issue_comments_count_proof_yes`, `issue_comments_count_mismatch_refused`,
  `state_comment_on_page_2_of_issue_comments_is_read`, `finding_comment_on_page_3_refused`,
  `duplicate_review_id_across_pages_refused`, `duplicate_issue_comment_id_across_pages_refused`,
  `duplicate_filename_across_pages_refused`, `entry_without_id_or_filename_refused`.
- Mutant: an implementation that reads page 1 only exits 0 on `entries_101_gate_file_on_page_2_refused` and
  `entries_301_gate_file_on_page_4_refused`, sees no state comment in `state_comment_on_page_2_of_issue_comments_is_read`, and turns
  them red; one that ignores the exit status of `gh` or continues past a failed page turns `page_2_error_refuses_and_removes_out` red;
  one that never compares the collected count turns `fewer_entries_than_expected_refused` and `issue_comments_count_mismatch_refused`
  red; one with no 3000 ceiling turns `ceiling_3000_refused` red; one that stops at the first short page without an empty-page rule
  turns `exact_200_ends_on_empty_page` red when the stub serves a third non-empty page; one that accepts a repeated entry turns the
  three `duplicate_*` rows red (the stub serves a list that shifted by one between pages with the same total count).
- RED-now: R21 (the script's `fetch-pages` has no issue-comment endpoint). Green path: M6 (issue-comment listing); the paging rows of M2 are landed regression rows.

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
- Then the order is: head re-verification, then the `author` step, then the required-checks step, then `files`, then `review`, then the merge step; the
  `author` step reads the pull request object once and records `eligible` once (`true` only for login `GoosLab`, id 6458969, type `User`, with a head
  repository equal to the base repository; `false` for a bot, a fork, a missing head repository, another author, a lookalike, and an unreadable object),
  ends with exit status 0 in every case, and the `if` of the required-checks step, `files`, `review`, and the merge step each require
  `steps.author.outputs.eligible == 'true'`; the `files` step
  reads the pull request object before and after the list, passes the pull request's `changed_files` as the
  expected count, runs `fetch-pages` then `gate-files-changed` with the verified head SHA, and writes
  `ok=true` only when both exit 0 and the head read twice equals the verified SHA; the `review` step reads the
  pull request object on every pass, passes `comments` as the expected count for the issue-comment list and `review_comments` for the review-comment list,
  runs `fetch-pages` on the issue comments, the reviews, and the review comments, reads the pull request object again after those list reads
  (a head other than the verified SHA ends the step with `ok=false` `state=fail` and the lists are never
  judged), then runs `read-native-review` with the verified head SHA of step `head` and the `draft` flag, and
  writes `ok=true` only on exit 0 and after a recheck that reads a fresh PR object, requires its head to equal the verified SHA and its
  `review_comments` count to equal the count observed during the clean pass, captures its valid nonnegative `comments` count as the recheck
  anchor, re-lists the complete issue comments, reviews, and review comments with the lists count-proofed against that anchor, then reads the PR
  object again and requires the same head, the same anchored `comments` count, and the unchanged `review_comments` count; it reruns
  `read-native-review` over those complete lists and requires exit 0 with the identical clean summary (comment id, `updated_at`, and compact JSON
  of both markers). An issue-comment count increase since the earlier clean pass alone does not fail, but an incomplete or changing recheck list,
  a changed head or review-comment count, a second state comment, or any changed trusted verdict does; every other exit — 1, 11, and an exit the
  step does not recognise — leaves `ok` not `true` and ends the step with a non-zero exit status (REQ-CCR-018;
  the same holds for every non-pass path of the `files` step, including a missing script); a head other than the verified SHA ends either step with
  `ok=false`; the `review` step is skipped unless `files` passed; the merge step's `if` requires both
  `steps.files.outputs.ok == 'true'` and `steps.review.outputs.ok == 'true'` besides its existing terms; both
  steps run `scripts/ci/codex-review-gate.sh` from the checkout that has no `ref` (the default branch), and a
  missing script is not a pass; and the workflow permissions are exactly `contents: write`, `pull-requests: write`,
  (revision 13, REQ-CCR-025) `statuses: write`, and (revision 13.2) the read-only `actions: read` — the row `permissions_unchanged`
  keeps its name and is amended to that four-key set (AC-CCR-006a).
- Verify: `grep -c 'gate-files-changed' .github/workflows/auto-merge.yml`, `grep -c 'read-native-review'
  .github/workflows/auto-merge.yml`, and `grep -c 'fetch-pages' .github/workflows/auto-merge.yml` each print a
  count of at least 1 and exit 0 (R4, R5, R6 flipped; necessary, not sufficient); `grep -c -F '6458969' .github/workflows/auto-merge.yml` prints at
  least 1 and exits 0 (R23 flipped; necessary, not sufficient); and `go test -count=1 -v
  -run '^TestAutoMergeWiresGuardAndGate$' ./internal/template/` lists `--- PASS` for `files_step_present`,
  `review_step_present`, `order_checks_then_files_then_review_then_merge`,
  `files_step_depends_on_head_and_checks_ok`, `review_step_depends_on_files_ok`,
  `merge_if_requires_files_and_review_ok`, `steps_run_script_from_default_checkout`,
  `files_step_passes_changed_files_count`, `review_step_passes_review_comments_count`, `review_step_passes_comments_count`,
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
  `review_step_forged_comment_never_ok`,
  `author_step_present_after_head_before_checks`, `order_head_author_checks_files_review_merge`, `checks_files_review_and_merge_if_require_author_eligible`,
  `author_step_exit_0_in_every_case`, `author_goosLab_user_same_repo_eligible_true`, `author_other_login_eligible_false`,
  `author_lookalike_goosLab_other_id_eligible_false`, `author_right_id_type_bot_eligible_false`, `author_dependabot_bot_eligible_false`,
  `fork_head_repo_id_differs_eligible_false`, `head_repo_null_eligible_false`, `pr_object_unreadable_author_unreadable_eligible_false`,
  `author_output_recorded_once`,
  `review_step_recheck_requires_same_comment_id_and_updated_at`, `review_step_issue_comments_moved_counts_pending`,
  `review_step_issue_comments_moved_between_recheck_reads_step_exit_nonzero`, `review_step_second_state_comment_in_recheck_step_exit_nonzero`,
  `unrelated_issue_comment_added_before_recheck_does_not_block`, `missing_issue_comment_count_on_recheck_not_ok`, `malformed_issue_comment_count_on_recheck_not_ok`, `issue_comment_count_moved_during_recheck_not_ok`. The unrelated-comment row proves that a fresh count of 4 is accepted after the earlier pass read 3; the missing/malformed rows fail closed and the movement row proves R3/list/R4 count disagreement still fails. **Revision 6:** each `files_step_*_not_ok` and `review_step_*_not_ok`
  row also asserts a non-zero step exit status, and each `*_ok_true` row asserts exit status 0 (the full step
  and job assertions are AC-CCR-020 to AC-CCR-025).
- Mutant: deleting the `files` step turns `files_step_present` red; a `files` step that discards the exit code
  of the guard (`|| true`, or `ok=true` written regardless) turns `files_step_workflow_edit_not_ok` and
  `files_step_page_error_not_ok` red; a `files` step that reads the list by pull request number and never
  reads the head again turns `files_step_head_moved_during_read_not_ok` red; a `review` step that reads the
  head only before its list reads and never again after them turns
  `review_step_head_moved_during_read_not_ok` red (on that fixture it would pass the clean lists to the
  script, exit 0, and write `ok=true`); a recheck that does not use the fresh R3 count to verify the full issue-comment page list, or that
  fails to compare the R3 and R4 counts, turns the unrelated-comment and count-movement rows red; a `review` step that ignores the
  exit code turns `review_step_exit_1_not_ok_state_fail` red; a recheck that compares only the head (not the comment id and `updated_at`)
  turns `review_step_recheck_requires_same_comment_id_and_updated_at` red; a recheck that does not use the fresh R3 count to prove the full page list,
  or does not require R4 to retain the R3 count, turns the unrelated-comment and count-movement rows red (the loop's own pass keeps retrying a moved
  count, `issue_comments_moved_counts_pending_not_failure`); an `author` step that matches the login only turns
  `author_lookalike_goosLab_other_id_eligible_false` red, one that skips the head repository comparison turns
  `fork_head_repo_id_differs_eligible_false` red, one that reads `eligible=true` for an unreadable object turns
  `pr_object_unreadable_author_unreadable_eligible_false` red; dropping the `eligible` term from any of the four `if` turns
  `checks_files_review_and_merge_if_require_author_eligible` red; treating exit 11 or an unknown exit as pass turns
  `review_step_unknown_exit_not_ok` red (exit 11 is AC-CCR-008); a merge `if` without the `review` term turns
  `merge_if_requires_files_and_review_ok` red, and so does a merge `if` without the `files` term: that term
  is pinned here, structurally (the condition is parsed and must name both `steps.files.outputs.ok == 'true'`
  and `steps.review.outputs.ok == 'true'`), because AC-CCR-019 cannot observe its removal by behaviour (the
  `review` step already depends on `files`); placing the merge step before the gate steps turns
  `order_checks_then_files_then_review_then_merge` red; omitting the count arguments or the head SHA turns
  `files_step_passes_changed_files_count`, `review_step_passes_review_comments_count`, or
  `files_step_passes_verified_head_sha` red. The unit tests of the script alone (AC-CCR-001 to AC-CCR-010)
  cannot catch any of these, which is why this criterion exists; the merge decision itself is AC-CCR-019.
- RED-now: R19, R23, and R27 (the issue-comment collection, the `author` step, and the recheck of the same comment are absent from the workflow and its tests). R4 to R6 flipped at M3 and no longer bind. Green path: M6.

**AC-CCR-013** — REQ-CCR-013 — regression-guard
- Given `auto-merge.yml` after the edit and the set of workflow files added by this SPEC's commits,
- When the triggers, the secret references, and the steps are read,
- Then the triggers are exactly `workflow_run` and (revision 13) `issue_comment` with the types `created` and `edited` (the row
  `trigger_is_workflow_run_only` is amended to that two-trigger set; AC-CCR-013a), no trigger of the `pull_request_target` family appears, the only
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

**AC-CCR-014** — REQ-CCR-014 — regression-guard
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
- RED-now: none on this tree. M4 landed the retirement; R7 and R10 no longer reproduce. Under verification-completeness §2.1 this criterion is a regression-guard, not a release gate; its adoption evidence is the mutant red the run phase records (not re-observed by revision 8: a stated Gap).

**AC-CCR-015** — REQ-CCR-015 — regression-guard
- Given `.coderabbit.yaml`, the workflows, and `scripts/ci/`,
- When they are read,
- Then the file exists, `reviews.auto_review.enabled` is `false`, `chat.auto_reply` is `false`, and zero files
  under `.github/workflows/` or `scripts/ci/` mention CodeRabbit; both directories were swept non-empty.
- Verify: `go test -count=1 -v -run '^TestCodeRabbitDisabledAndUnread$' ./internal/template/` lists `--- PASS`
  for `config_present`, `auto_review_false`, `auto_reply_false`, `no_coderabbit_in_workflows`,
  `no_coderabbit_in_scripts_ci`, `sweep_is_nonempty`.
- Mutant: deleting the file, leaving `enabled: true`, or reading the CodeRabbit status in a workflow turns a
  row red.
- RED-now: none on this tree. M3 and M4 landed; R8 and R2 no longer reproduce. Under verification-completeness §2.1 this criterion is a regression-guard, not a release gate; its adoption evidence is the mutant red the run phase records (not re-observed by revision 8: a stated Gap).

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

**AC-CCR-018** — REQ-CCR-017 — regression-guard
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
- RED-now: none on this tree. M4 landed; R11 and R10 no longer reproduce. Under verification-completeness §2.1 this criterion is a regression-guard, not a release gate; its adoption evidence is the mutant red the run phase records (not re-observed by revision 8: a stated Gap).

### 2.5 Merge decision path (M3)

**AC-CCR-019** — REQ-CCR-012 — release-blocking
- Given the real `auto-merge.yml` parsed as YAML, a stub `gh` that records every call and answers the pull request object (`user`, `head.repo`, `base.repo`, `draft`,
  `head.sha`, the counts), the listings, `gh pr view`, `gh pr merge`, and **`gh pr checks`**, a stub gate script that exits a chosen code (and the real script for the
  end-to-end rows), the `merge_deadline` environment value set to the past or the future, and an evaluator for the `if:` conditions limited to the grammar the workflow
  uses (`steps.<id>.outputs.<name> == '<literal>'` terms joined by `&&`, optionally led by the status function `always()`, which the harness reads as the exemption from the
  implied `success()`; any other condition shape fails the test). **The `gh pr checks` stub:** for every fixture below the stub answers
  `gh pr checks "$PR_NUMBER" --required --json name,state,bucket` with a passing required list and exit 0, for the base `main` (the five required checks) and for the base
  `release/main-batch-20261019` alike; this is an assumption, and it means these rows pin the review, author, and merge path under that assumption. The one row that reflects
  the real answer on the unprotected base uses a stub that exits 1 with `no required checks reported` (E43), and pins the Q27 consequence.
- When the chain `head → author → checks → files → review → merge condition → merge step` is run on each fixture,
- Then the fixtures fall into five groups, each with its own merge-call count and job result:
  **A, the clean fixture** (GoosLab, id 6458969, type User, head repository equal to the base repository, a clean state comment): `gh pr merge` is called exactly once, with
  `--match-head-commit` equal to the verified head SHA, no `--delete-branch` (REQ-CCR-021), and the squash flag for a feature branch (the merge flag for a `release/` branch); the
  job result is success.
  **B, gate refusals** (zero merge calls; the `files` or `review` step ends non-zero, the later steps are skipped, the job result is failure): the gate exits 1; the gate exits 11
  for a draft; no state comment appears through the whole window (NOT RUN); a leg is still `running` on the last pass; the state is the #1842 form (a completed leg with null
  fields beside a running security marker); the state was recorded for the old head of the #1842 move (`45b2dbf682515dd8f4d8d67ce0ae7ac2fa8328a5`, composed fixture) while the head is
  now `e7cceb2cb5d6e1cd9dbd2f465c4b9d5a1d631f21`; the state is the #1842 re-request state with markers carrying different SHAs (`F-mixed-sha`); the state is the #1833 findings case
  (`F-findings-1833`); a leg `failed` (`F-failed`, #1844); a forged state comment of another author is the only one; the gate exits an unrecognised code; the state is stale; a gate path is changed; a list page fails;
  the gate script is missing.
  **C, the merge-step guards** (zero merge calls; revision 13: the merge step ends with exit 1 and the job result is **failure**,
  REQ-CCR-024 — formerly "the job result stays success", superseded by the 2026-10-23 ruling; the rows `head_moved_at_merge_step` and
  `deadline_expired` of `refusal_fixtures_job_result_failure_clean_fixture_success` and `merge_step_owned_guards_job_result_success`
  are amended to failure, AC-CCR-019a): the head differs at the merge step's own `gh pr view` (it moves between
  the review step's recheck and the merge call); the deadline has passed.
  **D, author policy** (zero merge calls; the `author` step exits 0 with `eligible=false`, the `checks`, `files`, `review`, and merge steps are skipped, and the job result is success):
  the author is a bot (dependabot, id 49699333, type Bot), another user, GoosLab's login with another id, id 6458969 with type Bot; the pull request is a fork (the head repository id
  differs from the base repository id) or has no head repository; the pull request object is unreadable.
  **E, recheck races** (zero merge calls; the `review` step ends non-zero, the job result is failure): a second state comment, or one more issue comment, appears between the two
  pull request reads of the recheck (the stub answers the second `comments` count differently); the state comment's `updated_at` or markers differ between the pass's
  listing and the recheck's re-listing (the stub answers the second issue-comment list call differently from the first). **Timing limit, stated exactly:** an in-place edit or a dismissal
  that lands after the recheck's last list read changes no count and no value the step still holds, so it is not detectable and no row claims it (design §6, residual 7).
  **Every fixture of A, B, D, and E runs twice**: with the base `main` and with the base `release/main-batch-20261019`, a branch that is not protected (`protected: false`, no rules,
  E43) and the base of the five measured pull requests; the verdicts are identical on both bases, a running or absent state comment yields zero merge calls on either, a clean one
  yields exactly one, and no base is exempt (REQ-CCR-006). **The unprotected-base row** (`release_base_checks_exit_1_zero_merge_calls`): with the stub exiting 1 on `gh pr checks`, the
  `checks` step withholds (exit 0, `should_merge=false`), `files` and `review` are skipped, `gh pr merge` is called zero times, and the job result is success; this pins the shipped
  consequence of decision-index Q27 and does not claim the production workflow can merge on that base.
- Verify: `go test -count=1 -v -run '^TestAutoMergeMergeDecisionPath$' ./internal/template/` lists `--- PASS` for
  group A: `clean_fixture_merges_once_with_head_pin`, `release_branch_clean_fixture_uses_merge_flag`, `release_base_clean_fixture_merges_once`, `author_goosLab_clean_fixture_merges_once`,
  `merge_call_has_no_delete_branch_flag`, `clean_fixture_job_result_success`;
  group B: `review_exit_1_zero_merge_calls`, `review_draft_exit_11_zero_merge_calls`, `review_absent_expiry_zero_merge_calls`, `review_unknown_exit_zero_merge_calls`,
  `stale_review_zero_merge_calls`, `gate_path_change_zero_merge_calls`, `page_failure_zero_merge_calls`, `gate_script_missing_zero_merge_calls`,
  `release_base_running_zero_merge_calls`, `release_base_absent_zero_merge_calls`, `review_running_last_pass_zero_merge_calls`,
  `null_leg_completed_security_marker_running_zero_merge_calls`, `old_head_state_zero_merge_calls_f_moved_head`, `mixed_sha_f_mixed_sha_zero_merge_calls`,
  `findings_f_findings_1833_zero_merge_calls`, `review_failed_f_failed_zero_merge_calls`, `forged_state_comment_zero_merge_calls`, `refusal_fixtures_job_result_failure`;
  group C: `head_moved_at_merge_step_zero_merge_calls`, `head_moves_between_recheck_and_merge_call_zero_merge_calls`, `deadline_expired_zero_merge_calls`,
  `merge_step_owned_guards_job_result_success`;
  group D: `author_dependabot_bot_clean_fixture_zero_merge_calls`, `author_other_user_clean_fixture_zero_merge_calls`, `author_lookalike_clean_fixture_zero_merge_calls`,
  `author_right_id_type_bot_clean_fixture_zero_merge_calls`, `fork_clean_fixture_zero_merge_calls`, `head_repo_null_clean_fixture_zero_merge_calls`,
  `author_pr_object_unreadable_zero_merge_calls`, `author_fixtures_job_result_success_eligible_false`;
  group E: `recheck_issue_comments_moved_zero_merge_calls`, `recheck_second_state_comment_zero_merge_calls`, `recheck_relist_updated_at_changed_zero_merge_calls`,
  `recheck_race_job_result_failure`;
  and `release_base_checks_exit_1_zero_merge_calls`, `if_expression_outside_grammar_fails_test`.
- Mutant (each is applied to the committed workflow and must turn the named row red): `&& false` appended to the merge `if`, or an `exit 0` placed before `gh pr merge`, turns
  `clean_fixture_merges_once_with_head_pin` red (zero calls where one is required; the first also turns `if_expression_outside_grammar_fails_test` red);
  `merge_deadline=9999999999` assigned inside the merge step turns `deadline_expired_zero_merge_calls` red; **dropping `steps.review.outputs.ok == 'true'` from the merge `if`
  is an equivalent mutant under the revision-6 rules** (a failed or non-passing `review` step skips the merge through the implied `success()`, so merge calls are 0 in the original
  and the mutant) and is covered by the structural checks of AC-CCR-012 (`review_step_depends_on_files_ok`, `merge_if_requires_files_and_review_ok`, and the review and `files`
  condition rows), not by a named AC-CCR-019 row; **mutant MF** (the `files` term removed from the `review` step's `if` and from the merge `if`, and both conditions led with
  `always()`) turns `gate_path_change_zero_merge_calls` red — the step-by-step trace is under "Mutant MF trace" below; removing the `files` term alone, from either condition, is an
  equivalent mutant under the revision-6 harness rules (a failed `files` step skips every later step whose `if` has no status function, so merge calls are 0 in the original and the
  mutant) and is pinned structurally by AC-CCR-012; dropping `--match-head-commit` or pinning it to another variable turns `clean_fixture_merges_once_with_head_pin` red; a merge `if` that
  names a base branch and skips the review terms on `release/*` turns `release_base_running_zero_merge_calls` red; a merge step that does not re-read the head before the call turns
  `head_moves_between_recheck_and_merge_call_zero_merge_calls` red; a reader that counts a state recorded for the old head turns `old_head_state_zero_merge_calls_f_moved_head` red, and one
  that reads the code leg alone turns `mixed_sha_f_mixed_sha_zero_merge_calls` red; a recheck that does not prove the complete issue-comment list against fresh R3 or does not compare R3 to R4 turns the unrelated-comment and count-movement rows red;
  dropping the `eligible` term from the merge `if` alone is an equivalent mutant by behaviour (the skipped guard steps leave the `files` and `review` outputs empty) and is pinned
  structurally by AC-CCR-012; a mutant that drops it from `checks`, `files`, `review`, and the merge `if` and leads them with `always()` (the shape of mutant MF) turns the `author_*`
  rows and the fork rows red; a merge call with `--delete-branch` turns `merge_call_has_no_delete_branch_flag` red; an `author` step that ends non-zero for an ineligible author turns
  `author_fixtures_job_result_success_eligible_false` red. This is the criterion that a workflow which never merges, or merges past a refusal, cannot pass; AC-CCR-011 and AC-CCR-012 check
  strings and steps and cannot.
- RED-now: R19, R20, R22, and R23 (the workflow collects no issue comments, the shipped reader requires APPROVED, the merge call passes `--delete-branch`, and no eligible author is
  named). R12 flipped at M3 and no longer binds. Green path: M6.

**Mutant MF trace (plan-time reasoning under the revision-6 harness rules; NOT executed. The executed red, with
its verbatim output, belongs to the run phase's M5 mutant run under the admission rule, and no report may say
it was observed before then).** Harness input, row `gate_path_change_zero_merge_calls`: pull request object
with head `H`, not a draft, `changed_files` listing one `.github/workflows/x.yml` entry bound to `H`; the issue-comment list holding the clean state comment of the identity bound to
`H`, the review and review-comment lists empty; the author eligible; `gh pr checks` all passing; `merge_deadline` in the future; `gh pr view` answering `H`; step order `head`,
`author`, `checks`, `files`, `review`, merge.
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
- Given the real `review` step run with each terminating condition of `T-CHAIN` rows 2a to 2i, 3, 6, 7, 8, 9 — a failed review run (`F-failed`), a state comment with
  unresolved findings, a corroborating finding comment, a stale state comment, a tampered marker, a leg still `running` on the last pass,
  an unknown gate exit, the head moved during the read, the review-comment count moved during the read, a pre-merge recheck that
  differs (another `updated_at`), lists unreadable to the end of the window, and a missing gate script — and with the clean pass of
  row 1,
- When the step runs through the harness,
- Then on every failure row the step exit status is non-zero (1), the derived job result is failure, `ok` is `false`, `state` is
  `fail`, the annotation is `::error::` naming the reason, and the merge step is skipped with zero `gh pr merge` calls; on the clean
  row the step exit status is 0, the job result is success, `ok=true`, `state=pass`; and the `review` step carries no
  `continue-on-error`.
- Verify: `go test -count=1 -v -run '^TestAutoMergeReviewFailureFailsStepAndJob$' ./internal/template/` lists `--- PASS` for
  `unresolved_findings_step_exit_nonzero_job_failure`, `finding_step_exit_nonzero_job_failure`,
  `stale_review_step_exit_nonzero_job_failure`, `tampered_marker_step_exit_nonzero`,
  `running_on_last_pass_step_exit_nonzero_job_failure`, `unknown_gate_exit_step_exit_nonzero`,
  `head_moved_during_read_step_exit_nonzero`, `review_comments_moved_during_read_step_exit_nonzero`,
  `review_failed_step_exit_nonzero_job_failure_annotation_names_re_request_target`, `recheck_mismatch_step_exit_nonzero`, `recheck_second_state_comment_step_exit_nonzero`, `unreadable_to_end_of_window_step_exit_nonzero`, `gate_script_missing_step_exit_nonzero`,
  `merge_step_skipped_after_review_failure`, `clean_pass_step_exit_0_job_success`, `review_step_has_no_continue_on_error`; and
  (revision 13.2, the header of `TestAutoMergeCommentPathRequiresCISuccess` cites this criterion) that function passes as listed
  under AC-CCR-013c.
- Mutant: the committed workflow (a `withhold` that ends `exit 0`) turns every failure row red on the exit status although its
  outputs read `ok=false state=fail`; `continue-on-error: true` on the step turns the job-result assertion of every failure row and
  `review_step_has_no_continue_on_error` red while the step exit status stays non-zero; a step that exits 1 on the clean pass turns
  `clean_pass_step_exit_0_job_success` red.
- RED-now: R24 (tree `e9ed17271c`: the rows renamed for the comment-contract fixtures are absent) and R18. The older cells R13 and R16 no longer reproduce (M5 landed). Green path: M6.

**AC-CCR-021** — REQ-CCR-008 — regression-guard
- Given a non-draft pull request with no state comment of the identity when the wait window ends (gate exit 11,
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
- RED-now: none on this tree. M5 landed; R13, R14, and R16 no longer reproduce. Under verification-completeness §2.1 this criterion is a regression-guard, not a release gate; its adoption evidence is the mutant red the run phase records (not re-observed by revision 8: a stated Gap).

**AC-CCR-022** — REQ-CCR-008 — regression-guard
- Given a draft pull request — with no state comment, with a clean state comment bound to the head, and with a
  state comment that reports unresolved findings —
- When the real `review` step runs through the harness,
- Then the step exit status is 11, the derived job result is failure and in no case success, `ok=false`,
  `state` is the single value `not-run:draft` (the key appears once; it is distinct from `pass` and from
  `fail`), the annotation is a `::notice::` that names the cause `draft`, says it is not a review result and
  never a pass, the review lists are not fetched, and no `gh pr merge` call is made; the state never reads
  as a review PASS.
- Verify: `go test -count=1 -v -run '^TestAutoMergeDraftIsNotRunNeverSuccess$' ./internal/template/` lists
  `--- PASS` for `draft_step_exit_11_job_not_success`, `draft_state_output_is_single_value_not_run_draft`,
  `draft_ok_output_false`, `draft_skips_list_fetch`, `draft_with_clean_state_comment_still_not_success`,
  `draft_annotation_notice_names_cause_and_says_not_a_pass`, `draft_zero_merge_calls`,
  `draft_state_distinct_from_fail_and_pass`.
- Mutant: the committed step (draft ends `exit 0`, `state=fail` then `state=not-run`) turns the first, second,
  and last rows red; mapping a draft to `state=pass` or `ok=true` turns `draft_ok_output_false` red;
  `continue-on-error` turns the job-result assertion red.
- RED-now: none on this tree. The draft path exits 11 before any list is read and is already implemented and tested (M5 landed; plan audit N1); R25 (the renamed row
  `draft_with_clean_state_comment_still_not_success` is absent) is a test-addition cell outside the class claim, because adding a row changes a name, not a behaviour. Under
  verification-completeness §2.1 this criterion is a regression-guard; its adoption evidence is the red of the named mutant (the committed draft ending `exit 0` with `state=fail` then
  `state=not-run`) recorded by the run phase (not re-observed by revision 9: a stated Gap).

**AC-CCR-023** — REQ-CCR-019 — regression-guard
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
- RED-now: none on this tree. M5 landed; R14 no longer reproduces. Under verification-completeness §2.1 this criterion is a regression-guard, not a release gate; its adoption evidence is the mutant red the run phase records (not re-observed by revision 8: a stated Gap).

**AC-CCR-024** — REQ-CCR-018 — regression-guard
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
- RED-now: none on this tree. M5 landed the harness; R16 no longer reproduces. Under verification-completeness §2.1 this criterion is a regression-guard, not a release gate; its adoption evidence is the mutant red the run phase records (not re-observed by revision 8: a stated Gap).

**AC-CCR-025** — REQ-CCR-018 — regression-guard
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
- RED-now: none on this tree. M5 landed; R15 and R16 no longer reproduce. Under verification-completeness §2.1 this criterion is a regression-guard, not a release gate; its adoption evidence is the mutant red the run phase records (not re-observed by revision 8: a stated Gap).

### 2.7 Comment-path re-evaluation and visible holds (plan revision 13; M7)

Lettered sub-criteria of existing logical criteria (count unchanged at 25). Each release-blocking sub-criterion cites its RED-now
cells from the revision 13 ledger (tree `41b7f29b7a`) and names the milestone that flips it. The test functions and rows named below
are new (M7); the amended existing rows are listed in `plan.md` M7.

**AC-CCR-013a** — REQ-CCR-022 (with REQ-CCR-013) — release-blocking
- Given the real `auto-merge.yml` parsed as YAML and an evaluator for its job-level `if` over a synthetic event context
  (`event_name`, `issue.pull_request`, `comment.user.id`, and the `workflow_run` fields),
- When the triggers and the job condition are read and evaluated for: an `issue_comment` `created` and `edited` by user id 199175422
  on a pull request; the same id on an issue that is not a pull request; another user, another bot (`coderabbitai[bot]`, dependabot id
  49699333), the login `chatgpt-codex-connector` with another id, and a null user, each on a pull request; and the unchanged
  `workflow_run` cases (event `pull_request` with conclusion `success`, and each other combination),
- Then the triggers are exactly `workflow_run` and `issue_comment` with the types exactly `created` and `edited`; the job runs for the
  identity's comment on a pull request and for a successful `pull_request` `workflow_run` only; every other case skips the job, so no
  step runs (zero step executions recorded by the harness) and no `gh` call is made.
- Verify: `grep -c 'issue_comment' .github/workflows/auto-merge.yml` prints at least 1 and exits 0, and `grep -c -F
  'github.event.comment.user.id' .github/workflows/auto-merge.yml` prints at least 1 and exits 0 (R28, R29 flipped; necessary, not
  sufficient); and `go test -count=1 -v -run '^TestAutoMergeCommentPathTrustGate$' ./internal/template/` lists `--- PASS` for
  `triggers_are_workflow_run_and_issue_comment_created_edited`, `codex_comment_on_pull_request_job_runs`,
  `codex_comment_edited_on_pull_request_job_runs`, `codex_id_on_plain_issue_job_skipped`, `other_user_comment_job_skipped`,
  `other_bot_comment_job_skipped`, `lookalike_login_other_id_job_skipped`, `null_user_job_skipped`,
  `skipped_job_records_zero_steps_and_zero_gh_calls`, `workflow_run_success_pull_request_job_runs`,
  `workflow_run_other_cases_job_skipped`.
- Mutant: a job `if` that matches the login instead of the id, or drops `issue.pull_request`, or a trigger list without `edited`, or
  with `pull_request_review` added, turns the matching row red.
- RED-now: R28 and R29 (tree `41b7f29b7a`: no `issue_comment` trigger, no comment author gate). Green path: M7 adds the trigger and
  the job condition; the rows above then pass.

**AC-CCR-013b** — REQ-CCR-022 (with REQ-CCR-011) — regression-guard
- Given the real `auto-merge.yml`,
- When every step's `uses`, `with`, `env`, and `run` text is read,
- Then the checkout step keeps no `ref` (on both paths); no expression names `github.event.comment.body`, `github.event.issue.title`,
  `github.event.issue.body`, `github.head_ref`, or the pull request's head ref or SHA as a checkout input; `github.event.issue.number`
  appears only as an `env` value or in the concurrency key, never inside a `run` script; no step runs a file other than
  `scripts/ci/codex-review-gate.sh` from the default-branch checkout; on the comment path the `pr` step reads `state`, `merged`, and
  `headRefOid` from the API.
- Verify: `go test -count=1 -v -run '^TestAutoMergeCommentPathRunsNoPullRequestCode$' ./internal/template/` lists `--- PASS` for
  `checkout_has_no_ref_on_both_paths`, `no_comment_body_or_issue_text_expression`, `issue_number_only_in_env_or_concurrency`,
  `no_head_ref_checkout_input`, `gate_script_only_from_default_checkout`, `comment_path_pr_step_reads_state_from_api`; and
  `TestAutoMergeNeverInterpolatesExpressionsIntoScripts` (landed by `41b7f29b7a`) still passes.
- Mutant: `ref: refs/pull/${{ github.event.issue.number }}/head` on the checkout, a `run` line that interpolates
  `${{ github.event.issue.number }}`, or an `env` value of `${{ github.event.comment.body }}` turns the matching row red. The red must
  be recorded verbatim at M7.
- RED-now: none (no RED exists: the checkout has no `ref` and no payload text is read on this tree by nature). Regression-guard, never
  recorded as a release gate (verification-completeness §2.1). Green path: stays green through M7.

**AC-CCR-013c** — REQ-CCR-023 — release-blocking
- Given the real `auto-merge.yml`, a stub `gh` that records every call, the real or a stub gate script, and the comment-path event
  context (`issue.number` 7, the identity's comment),
- When the chain is run on the comment path for: the clean fixture; the clean fixture with the head moved after the comment (the API
  head differs from the head the state comment names); a state comment with a leg `running`; no state comment; the draft fixture; and
  the concurrency key is read for both events,
- Then on the comment path the steps run in the same order and with the same conditions as on the `workflow_run` path; the clean
  fixture merges exactly once with `--match-head-commit` equal to the API-read head; the moved head yields zero merge calls and a
  failed `review` step (`stale-review`) or a held merge; the `review` step makes one pass with `--wait-closed false` and zero sleeps;
  (revision 13.1, decision-index Q31(a) reversed) a `running` leg or marker, and an absent state comment, are a "not yet" hold: the
  `review` step ends with exit 0, `ok` not `true`, the single state `pending`, zero merge calls, the job result is success, and the
  `auto-merge` status is `pending` with `review-not-complete:running` (or `review-not-complete:no-state-comment`); the draft ends
  `not-run:draft`; the concurrency group key resolves to the same `auto-merge-7` for a `workflow_run` of pull request 7 and for the
  identity's comment on pull request 7, and to a run-id key (`auto-merge-run-<run id>`) that differs from `auto-merge-7` for another
  author's comment on pull request 7; the group is declared at the job level, and `cancel-in-progress` is false.
- Verify: `grep -c -F 'github.event.issue.number' .github/workflows/auto-merge.yml` and `grep -c -F 'github.event_name'
  .github/workflows/auto-merge.yml` each print at least 1 and exit 0 (R30, R31 flipped; necessary, not sufficient); and `go test
  -count=1 -v -run '^TestAutoMergeCommentPathSameChainOnce$' ./internal/template/` lists `--- PASS` for
  `comment_path_same_step_order_and_conditions`, `comment_path_clean_fixture_merges_once_with_head_pin`,
  `comment_path_head_moved_after_comment_zero_merge_calls`, `comment_path_review_single_pass_wait_open_no_sleep`,
  `comment_path_running_is_not_yet_hold_exit_0_zero_merge_calls`, `comment_path_absent_review_is_not_yet_hold_exit_0_zero_merge_calls`,
  `comment_path_draft_not_run_zero_merge_calls`, `concurrency_key_same_for_both_paths`,
  `concurrency_key_unique_per_run_for_other_authors`, `concurrency_group_at_job_level`, `concurrency_cancel_in_progress_false`;
  and (revision 13.2, the comment path's current-head `CI` condition, `design.md` item 4) `go test -count=1 -v -run
  '^(TestAutoMergeCommentPathRequiresCISuccess|TestAutoMergeCommentPathCIInProgressHolds|TestAutoMergeCommentPathCIConditionMatchesWorkflowRunPath)$'
  ./internal/template/` lists `--- PASS` for every subtest of the three functions, among them
  `required_checks_pass_but_ci_concluded_non_success_zero_merge_calls`, `ci_run_listing_unreadable_zero_merge_calls`,
  `no_ci_run_for_the_verified_head_zero_merge_calls`, `latest_run_of_the_head_decides`, `workflow_run_path_makes_no_ci_read`,
  `every_unfinished_status_is_a_pending_hold`, and `ci_workflow_is_the_one_both_triggers_name`.
- Mutant: the 60-pass loop kept on the comment path turns `comment_path_review_single_pass_wait_open_no_sleep` red; a pass with
  `--wait-closed true` (a running review failing the job) turns `comment_path_running_is_not_yet_hold_exit_0_zero_merge_calls` red; a
  concurrency key without `github.event.issue.number` turns `concurrency_key_same_for_both_paths` red, and one that gives a foreign
  comment the pull request number turns `concurrency_key_unique_per_run_for_other_authors` red; a `head` step that takes the head from the
  event instead of the API, or a merge call pinned to another variable, turns `comment_path_head_moved_after_comment_zero_merge_calls`
  or the clean row red.
- RED-now: R30 and R31 (tree `41b7f29b7a`: no comment-path pull request number or concurrency key, no event discrimination). Green
  path: M7.

**AC-CCR-019a** — REQ-CCR-024 — release-blocking
- Given the real merge step run against a stub `gh` with the guards seeded as passed, and the chain harness of AC-CCR-019,
- When the merge step meets each outcome: merge requested; deadline passed before the merge call; deadline passed during the head
  re-read; head moved before the call; `gh pr merge` exits 1,
- Then "merge requested" ends the step with exit 0 and the job result success; each of the four other outcomes ends the step with
  exit 1 and the job result failure, with zero merge calls for the deadline and head rows and one failed call for the merge-failed
  row; every outcome still writes exactly one job-summary line naming the outcome and the head SHA; and the "not yet" holds — `pr`,
  `head`, `author`, `checks`, and (revision 13.1) the comment path's pending `review` pass — still end with exit 0 and a job that is
  not failed. Group mapping: non-zero = merge step reached and not merged (the four rows above) only; exit 0 = merge requested and
  every "not yet" hold.
- Verify: `go test -count=1 -v -run '^TestAutoMergeHeldMergeOutcomesFail$' ./internal/template/` lists `--- PASS` for
  `merge_requested_exit_0_job_success`, `window_elapsed_before_merge_exit_1_job_failure`,
  `window_elapsed_during_head_check_exit_1_job_failure`, `head_moved_exit_1_job_failure`, `merge_failed_exit_1_job_failure`,
  `each_outcome_one_summary_line_with_head_sha`, `earlier_holds_keep_exit_0`, `comment_path_running_hold_keeps_exit_0`; and the amended rows of
  `TestAutoMergeSummarisesEveryMergeOutcome` and `TestAutoMergeMergeDecisionPath` (`plan.md` M7) pass; and (revision 13.2, a head
  lookup that fails inside the merge step) `go test -count=1 -v -run '^TestAutoMergeMergeHeadLookupFailureFailsClosed$'
  ./internal/template/` lists `--- PASS` for `merge_step_alone_summary_line_exit_1_zero_merge_calls`,
  `whole_chain_status_failure_job_failure_zero_merge_calls`, and `moved_head_keeps_its_own_wording`.
- Mutant: the committed merge step (each held outcome `exit 0`, the merge-failed path falling through) turns the four exit-1 rows
  red; an `exit 1` on "merge requested" turns `merge_requested_exit_0_job_success` red; a deleted summary line turns
  `each_outcome_one_summary_line_with_head_sha` red.
- RED-now: R32 and R33 (tree `41b7f29b7a`: the head-moved and deadline holds end `exit 0`, and the merge-failed path ends on the
  summary echo with exit 0). Green path: M7.

**AC-CCR-006a** — REQ-CCR-025 — release-blocking
- Given the real `auto-merge.yml` and the chain harness with a stub `gh` that records every call,
- When the permissions are read, the `gh` invocations are classified, and the chain runs each `T-STATUS` outcome (merge requested; a
  merge step reached and not executed; a `review` refusal; review absent; the draft; `checks` withheld; an ineligible author; a `head`
  hold; a head read that failed; a failing status write),
- Then the permissions are exactly the four `contents: write`, `pull-requests: write`, `statuses: write`, and `actions: read`
  (revision 13.2: the read-only scope of the comment path's `CI` run listing, `design.md` item 4); the only status call is
  `gh api --method POST "repos/$REPO/statuses/$SHA"` and no status is read; for every outcome with a known evaluated head exactly one
  status is posted on that SHA with the context `auto-merge`, the state of `T-STATUS` (`success`, `failure`, or `pending`), and a
  description beginning with the `T-STATUS` text; no status is posted when the head is unknown; a failing status write ends the
  `status` step with exit 1 and does not change the merge-call count; and `.github/required-checks.yml` is byte-unchanged.
- Verify: `grep -c -F 'statuses' .github/workflows/auto-merge.yml` prints at least 1 and exits 0, and `grep -c -F 'statuses: write'
  .github/workflows/auto-merge.yml` prints 1 and exits 0 (R34, R35 flipped; necessary, not sufficient); `git diff --quiet
  8c36a7904bfe3e44c46b154a2839a6aedb719ef2 HEAD -- .github/required-checks.yml` exits 0 (B1, regression row; re-pinned to the
  `origin/main` tip read at measurement time if that ref moved, never compared against the moving name); and `go test -count=1 -v -run
  '^TestAutoMergePostsStatusContext$' ./internal/template/` lists `--- PASS` for
  `permissions_exactly_contents_pull_requests_statuses_write_actions_read`, `status_post_is_the_only_statuses_form`, `no_status_read`,
  `merge_requested_posts_success`, `merge_not_executed_posts_failure`, `review_refusal_posts_failure`,
  `review_absent_posts_failure`, `comment_path_running_posts_pending_review_not_complete_running`, `draft_posts_pending`, `checks_hold_posts_pending`, `ineligible_author_posts_pending`,
  `head_hold_posts_pending`, `unknown_head_posts_no_status`, `one_status_per_evaluation_on_evaluated_sha`,
  `context_is_auto_merge`, `description_names_reason`, `status_write_failure_step_exit_1_merge_unchanged`; and AC-CCR-016 and
  AC-CCR-017 still pass.
- Mutant: a status posted only on success (no `pending`/`failure`), a status step without `always()`, an `actions: write` or a
  missing `actions: read`, a `statuses: read` or a GET of
  `commits/<sha>/status`, or a status on the head read after the merge instead of the evaluated SHA turns the matching row red; adding
  `auto-merge` to `required-checks.yml` turns B1 and AC-CCR-017 red.
- RED-now: R34 and R35 (tree `41b7f29b7a`: no status write, no `statuses: write`). B1 is a baseline row, green by nature. Green
  path: M7.

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
| REQ-CCR-020 | AC-CCR-012, AC-CCR-019 | REQ-CCR-019 | AC-CCR-023 |
| REQ-CCR-021 | AC-CCR-006, AC-CCR-019 | REQ-CCR-022 | AC-CCR-013a, AC-CCR-013b |
| REQ-CCR-023 | AC-CCR-013c | REQ-CCR-024 | AC-CCR-019a |
| REQ-CCR-025 | AC-CCR-006a (with AC-CCR-016, AC-CCR-017 for the unchanged contract) | | |

Counts (revision 13): 25 requirements, 25 logical criteria with 5 lettered sub-criteria (ceilings 25 and 25; both counts are at
the ceiling, so a further requirement or logical criterion means tiering up or splitting). Revision 12 counts were 21 requirements
and 25 criteria. Every requirement maps to at least one criterion;
every criterion names one primary requirement in its heading, and AC-CCR-006, AC-CCR-012, and AC-CCR-019 also carry rows of REQ-CCR-020 and REQ-CCR-021; AC-CCR-016 and AC-CCR-017 share REQ-CCR-016, AC-CCR-012
and AC-CCR-019 share REQ-CCR-012, AC-CCR-008, AC-CCR-021, and AC-CCR-022 share REQ-CCR-008, and AC-CCR-020,
AC-CCR-024, and AC-CCR-025 share REQ-CCR-018. No criterion claims that doctrine is updated, and none depends on it (spec
§7, KD-1). No criterion depends on a deferred gate or on DS-1, DS-2, or DS-3.

### 3.1 The four function-check failures, each pinned

The operator named four conditions that must each be withheld — never a pass. Each is pinned by a criterion
and by a mutant that a weaker implementation would satisfy.

| Condition | Observable verdict | Pinned by | Mutant that turns it red |
|-----------|--------------------|-----------|--------------------------|
| Missing review | `read-native-review` exits 10 (`pending:no-state-comment`) while the window is open, never 0; at window close it exits 11 `not-run:review-absent` (the script's NOT RUN class; the `review` step ends in failure on it, never success, AC-CCR-021); the merge is withheld and never authorized | AC-CCR-007 (`empty_comments_pending`, `expiry_never_merges`), AC-CCR-008 (`review_absent_window_closed_exit_11`, `not_run_message_review_absent_names_cause`), AC-CCR-003 (forged artifacts do not fill the gap), AC-CCR-019 (`review_absent_expiry_zero_merge_calls`, `release_base_absent_zero_merge_calls`) | A script that reads an empty list as clean; a loop that merges on expiry; an absent comment left at 10 or read as success |
| Missing output | A leg or the `security_review` object missing; a leg `status` missing or null; a completed leg with a null `commit_sha` or a null, missing, or non-boolean `has_any_findings` or `has_unresolved_findings` (the #1842 and #1839 forms); a security marker missing: exit 1; a missing input file: `input-invalid`; a list that cannot be read at the end of the window: failure `review-unreadable` | AC-CCR-004 (`code_review_leg_missing_refused`, `security_review_leg_missing_refused`, `status_null_refused`, `completed_with_has_unresolved_null_refused`, `completed_code_missing_has_any_findings_refused`, `completed_security_missing_has_any_findings_refused`, `completed_code_null_has_any_findings_refused`, `completed_security_nonboolean_has_any_findings_refused`, `completed_leg_commit_sha_null_security_marker_completed_refused`, `security_leg_all_null_refused_f_null_security_leg`, `completed_legs_without_security_marker_refused`, `input_files_missing_refused`), AC-CCR-007 (`unreadable_to_end_of_window_fails_review_unreadable`) | A script that treats null as false, a missing field as pass, or reads only one leg |
| Failed review | A completed leg reports `has_any_findings` true even when `has_unresolved_findings` is false; a leg reports `has_unresolved_findings` true; a status other than `completed` (`failed`, `cancelled`, unknown); or a corroborating review comment or COMMENTED or CHANGES_REQUESTED review on the head: exit 1 | AC-CCR-004, AC-CCR-005 | A script that reads `completed` alone as a pass, treats reported findings as clean, or ignores the review comments |
| Stale SHA | Either leg or the security marker's `headSha` is bound to a commit other than the head (a push after the review started, the #1842 move): exit 1 `stale-review`, also when the stale state is still `running` | AC-CCR-002 (`both_legs_bound_to_older_head_f_stale`, `old_head_state_never_counts_for_new_head_f_moved_head`), AC-CCR-019 (`old_head_state_zero_merge_calls_f_moved_head`) | A script that ignores `commit_sha`, compares a prefix, checks only the code leg, or reads the security marker's `headSha` as optional |
| Running | A leg or the security marker reads `running`: exit 10 while the window is open, exit 1 `review-not-complete:running` on the last pass; never a pass, on any base | AC-CCR-007 (`running_both_legs_pending_f_running`, `running_window_closed_exit_1_not_complete`, `running_on_last_pass_fails_review_not_complete_running`), AC-CCR-019 (`review_running_last_pass_zero_merge_calls`, `release_base_running_zero_merge_calls`) | A loop that merges while running; a running state read as NOT RUN or as a pass |
| Tampered, duplicated, or misplaced marker | Marker JSON altered, the marker twice, the marker in the identity's prose or table, two state comments, a security marker that disagrees: exit 1 | AC-CCR-004 (`marker_json_truncated_refused`, `state_marker_duplicated_in_one_body_refused`, `state_marker_in_identity_prose_only_refused`, `two_state_comments_one_clean_refused`, `security_marker_disagrees_legs_completed_marker_failed_refused`) | A reader that scans the whole body without a position, picks the newest comment, or ignores the security marker |
| Wrong author | A comment of a `User` with the same login text, a different id than 199175422, a lookalike login, a null user, another bot, each carrying a clean marker: never read | AC-CCR-003 | A reader that matches the login only, the id only, or selects candidates by marker text |
| Head moved before the merge | The head read before the merge call differs from the verified head: zero merge calls | AC-CCR-019 (`head_moves_between_recheck_and_merge_call_zero_merge_calls`) | A merge step that does not re-read the head |

## 4. Edge cases

- A state comment edited in place between the review step's read and its recheck, or between the recheck and the merge: the recheck
  requires the same comment id, `updated_at`, and parsed fields, and the merge step re-reads the head before the call; a difference ends the
  step in failure or withholds the merge (AC-CCR-012, AC-CCR-019, `T-CHAIN` row 7).
- A push after the review started: the state is bound to the old head, so the moved head has no bound state → `stale-review`, withheld at
  once, also when the stale state is still `running`; the #1842 move is the real instance (AC-CCR-002, AC-CCR-019). A fresh review is the
  leader's `@codex review` request (DS-3, spec §5); whether it edits the comment or adds a second one is unobserved, and a second state
  comment fails closed (AC-CCR-004 `two_state_comments_one_clean_refused`).
- A forged comment, review, or marker text by another author or app, or a lookalike of the identity: never counted; with no real comment the
  read stays pending and the window ends in a failed step `not-run:review-absent` (AC-CCR-003, AC-CCR-008).
- A review `running` or absent when a merge is attempted, on `main` or on an unprotected integration base such as
  `release/main-batch-20261019`: zero merge calls (AC-CCR-007, AC-CCR-019); the incident of PR #1841 is the real instance.
- A completed leg with null fields beside a running security marker (#1842), a security leg with all fields null (#1839, #1840), a
  leg `completed` without a bound commit: never a pass (AC-CCR-004, AC-CCR-007).
- A gate-path edit past the first 100 changed files, a list API failure mid-pagination, a truncated or shifted list, a collected count
  different from the pull request's own count, and an entry of a head that moved: the merge is withheld (AC-CCR-010, AC-CCR-009,
  AC-CCR-012). A moved issue-comment count inside one pass is retried, never a failure (AC-CCR-007).
- An ordinary issue comment added between a clean pass and the pre-merge recheck does not alone withhold the merge: the complete page list
  must match the fresh R3 count and R4 must retain that same count. Movement during the recheck, a truncated list, or a changed trusted Codex
  artifact still withholds it (AC-CCR-012, REQ-CCR-012).
- A pull request that adds or changes any workflow file or the gate script: withheld for a human (AC-CCR-009).
- Draft pull request: `not-run:draft`, step exit 11, job result never success, withheld at once, whatever artifacts exist; its
  observation is `remote observed: gate path NOT RUN (not-run:draft)` at most, never `remote verified` (AC-CCR-008, AC-CCR-022, design
  §12). A non-draft pull request with no state comment at the end of the window is a failed `review` step with cause
  `not-run:review-absent`, observed as `remote observed: review step failed, review absent` at most (AC-CCR-021).
- A failure in the `files` step (a missing gate script, a gate-path change, a list that cannot be proven complete) ends that step and
  the job non-zero and skips the `review` and merge steps (AC-CCR-025); a failure in `review` ends it non-zero (AC-CCR-020); the earlier
  steps `pr`, `head`, `author`, `checks` keep the exit-0 withhold (Out of Scope) and post a `pending` `auto-merge` status (AC-CCR-006a);
  the merge step, reached and not executed, ends exit 1 (revision 13, AC-CCR-019a).
- A review that completes after a `workflow_run` evaluation's window or after the pull request becomes ready: the identity's comment
  edit starts a comment-path evaluation that reads it once through the same chain (AC-CCR-013a, AC-CCR-013c). A comment by anyone else
  starts a skipped job (AC-CCR-013a).
- The native review adds a `status` value or a marker `v2`: refused or not a candidate; the gate withholds until the contract is amended
  (design §3.2, §6, U14).
- A clean state comment whose prose says otherwise passes, because prose is never read: a documented residual risk, not a testable property
  (design §6). A user with write access can edit the bot's comment and the gate cannot tell (design §6, residual 8).
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
1. All 25 criteria and the 5 revision-13 sub-criteria pass on the merged-local tree with the swept-set rule satisfied.
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
5. The final run report lists gates G1-G9 and the deferred steps DS-1, DS-2, DS-3 (from `spec.md` §5) for the
   leader's draft pull request body, states known debt KD-1 and the accepted limitation AL-2 (AL-1 is superseded and resolved by
   REQ-CCR-022 and REQ-CCR-023, revision 13; the report names it as resolved, never as accepted), and
   names the merged-local SHA on local `main`.
6. The landing re-measure of `plan.md` §F ran and its output is recorded.
7. The report states that the remote conclusion of the `files` step, the `review` step, and the job is
   UNOBSERVED: the local criteria assert step exit statuses and a derived job result only (spec §4, G6
   acceptance condition), and the remote state of the amended workflow is "unverified".
8. (revision 7) The report states that the real run of the revision-7 reader on a pull request, the whole `running` to `completed`
   transition, DS-2, the effect of `@codex review`, and whether `CI` fires for a pull request into a `release/*` base are UNOBSERVED, and
   that the samples behind the fixtures are five completed comments, one running snapshot, and the #1842, #1839, #1840, and #1833 states
   (`research.md` E37 to E46), none of them a run of this gate.
9. (revision 7) The report states that the author, fork, and bot rule and the removal of `--delete-branch` are enforced by the workflow and
   pinned by the rows of AC-CCR-006, AC-CCR-012, and AC-CCR-019, not by a document; and that `delete_branch_on_merge` (G9) is an operator
   decision item this card did not change.
10. (revision 8) The report states that no regression-guard mutant red was re-observed by the plan phase (the run phase records them), and that the production
    workflow cannot merge on an unprotected base while the `checks` step is unchanged (decision-index Q27).
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
| Non-draft: a state comment bound to `H` that is not clean (a leg `running`, unresolved findings, another status, a disagreeing security marker) | `remote observed: review present, not clean (<state>)` — never a gate pass; the review ran, so the CI-and-review wording above is still the only route to `remote verified` |
| Non-draft: no review of the identity, or one bound to another commit | `remote observed: review step failed, review absent` or `remote observed: stale-review` — never verified |
