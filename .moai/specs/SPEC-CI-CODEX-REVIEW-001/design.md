# design.md — SPEC-CI-CODEX-REVIEW-001

## Delta of plan revision 7 (read this first; a delta auditor needs only this section)

Plan revision 7 is an in-place amendment (spec.md `## Amendments`, Amendment 2, version 0.6.0 to 0.7.0). Authority:
the leader's ruling and the operator follow-up of 2026-10-19 (board and messages relayed by the leader). Cause: the
native Codex review is one **issue comment** of the Codex app identity, edited in place until it ends, carrying
machine markers; the pull-request `reviews` list stays empty for a clean review (`research.md` E37 to E46). The
revision-6 reader looked for an APPROVED review object, so every real non-draft pull request would end NOT RUN and
every automated merge would be withheld for good. Motivating incident, one line: PR #1841 merged at
2026-10-19T05:13:31Z while its Codex reviews finished at 05:14:11Z and 05:15:04Z (`research.md` E41), so a
`running` or absent review must never be merged over, on any base. **Frozen and untouched:** the exit classes 0, 1,
10, 11 and their mapping through T-CHAIN (script exit, step exit, job result, outputs), REQ-CCR-009, 013 to
019 in meaning (REQ-CCR-011 gains only the no-branch-deletion clause, Δ25), the files step and the tamper guard (§3.4), one value per output key, the draft job-fails rule, the
deferred gates, KD-1 (card t1874), and every decision-index row not named below. Every clause this revision
changed is listed here with an id; text elsewhere that restates a listed clause names its id.

| Id | Where | Before | After | Ruling item |
|----|-------|--------|-------|-------------|
| Δ12 | spec REQ-CCR-001 to 005; spec §3 terms; design §3.1, §3.2 | The verdict carrier was a pull-request review of the identity in state APPROVED, newest bound review, supersession of a change request | The carrier is the identity's single state comment. Pass only when BOTH legs (`code_review`, `security_review`) read `completed`, `commit_sha` equals the head, and `has_unresolved_findings` is exactly `false`; `running` is pending; a status other than running or completed, a stale `commit_sha`, unresolved findings, malformed marker JSON, a duplicated marker, a marker outside its position, and more than one state comment fail closed; author bound by id 199175422 and type Bot, never by login text or body | comment protocol; pass condition; fail-closed list; author binding |
| Δ13 | spec REQ-CCR-006, 007, 008, 010, 012; design §3.3, §5 | Lists were files, reviews, review comments; a review absent at the window end was NOT RUN; the recheck compared the selected review | A fourth list, the PR's issue comments, with count proof against the PR object's `comments`; `running` waits inside the window and, on the last pass, fails (`review-not-complete:running`); no state comment at the window end stays exit 11 `not-run:review-absent`; the recheck re-reads the comments and requires the same comment id, `updated_at`, parsed fields, and head; the merge step reads the head again immediately before the call | comment protocol; pending rule; recheck; addition 2 |
| Δ14 | spec §1, §2 item 1, REQ-CCR-006; acceptance AC-CCR-019 (rows added to the existing criterion) | The scope clause named no base; the workflow has no base-branch condition | The gate applies to every merge base, an unprotected integration base such as `release/main-batch-YYYYMMDD` included; no base is exempt. All five measured pull requests have the base `release/main-batch-20261019` (E40), a branch with `protected: false` and no rules (E43) | addition 1 |
| Δ15 | acceptance AC-CCR-001 to 008, 010, 012, 019 to 023 (same identifiers, 25 criteria, none added or removed) | Fixtures were synthetic review objects with APPROVED, COMMENTED, CHANGES_REQUESTED | Fixtures are modelled on the real comment objects (table "Real-sample provenance" below); the author object fields are taken from the real samples (login `chatgpt-codex-connector[bot]`, id 199175422, type Bot); the refuse list of addition 2 is pinned row by row | fixtures from real bodies; addition 2 |
| Δ16 | spec §5 DS-1, DS-2, DS-3; decision-index Q20, FU-10 | DS-1 unobserved; Q20 and FU-10 awaited the clean form | DS-1 observed on #1835, #1836, #1837, #1838, #1841 (and #1842), not on #1828 itself, not on a draft. DS-2 stays open: no automatic re-review on a push was observed (#1836 stayed at the older head). DS-3 gains the open question whether `@codex review` edits the comment or creates a second one. Q20 and FU-10 are closed (superseded by the comment contract) | DS-1 observed elsewhere; DS-2 open |
| Δ17 | research E41 and its intro paragraph; spec §1 | none | Motivating incident preserved: PR #1841 merged at 05:13:31Z, its code review completed at 05:14:11Z and its security review at 05:15:04Z; the merge preceded the review and is not retroactively a pass | addition 3 |
| Δ18 | spec REQ-CCR-001, 004; design §3.2 steps 5 to 13; research E42, U15, U16; decision-index Q24 | (new) | A leg passes only with `completed`, `commit_sha` equal to the head, and `has_unresolved_findings` exactly `false`; null is not false. The `codex-security-review:v1` marker must stand once, read `completed`, and carry the head in `headSha`; a running security marker is pending, any other disagreement refuses. Samples: #1842, security leg `completed` with null fields beside a running security marker; the #1842 head move (a state recorded for the old head never counts for the new one); #1839 and #1840, a security leg that is all null including `status` (E42, E43) | addition 4 (#1842) |
| Δ19 | design §5 `T-CHAIN` rows 1 to 9 (same numbers, sub-rows 2a to 2i added) | Rows named review-object reasons (`changes-requested`, `review-not-conclusive`, `review-entry-invalid` of a review) | One table now carries the comment state in its first column: comment state, gate exit and reason, step exit, job result, `ok`, `state`, annotation. Exit classes, step exits, job results, and output values are unchanged for every path | one table merged into T-CHAIN |
| Δ20 | design §3.3 allow-list; acceptance AC-CCR-006 | `fetch-pages` accepted `pulls/N/(files\|reviews\|comments)`; `issues/` was a forbidden token | Adds exactly `issues/N/comments`; the workflow's own `gh api` calls are still the pull request object only; `issues/` is allowed in that one listing and nowhere else | issue-comments endpoint |
| Δ21 | design "Real-sample provenance" (below); acceptance fixture shape paragraph | none | Each fixture row names the real comment it is modelled on and what is synthetic | provenance |
| Δ22 | decision-index Q6, Q19, Q20 (restated), Q23, Q24, Q25 (new) | Q6, Q19, Q20 fixed the review-object carrier and APPROVED | Q6 and Q19 restate the carrier; Q20 superseded; Q23 (review objects and review comments only refuse), Q24 (the security marker), Q25 (issue-comment count noise) added with defaults | scope |
| Δ23 | plan.md M6; plan §A, §B, §F | M1 to M5 landed on the card branch | M6 re-implements the reader, the consumer wiring, and the tests to the revision-7 contract; the gate script, `auto-merge.yml`, and their tests change in the run phase, not in this revision | scope |
| Δ24 | design §6 (threat rows and residual risks 8 and 9), §11 | Residual risk assumed the artifact form unknown | Adds: a user with write access can edit the bot's comment; a re-request that creates a second state comment withholds the pull request; four of five measured comments were created after the merge | scope |
| Δ25 | spec REQ-CCR-020 and REQ-CCR-021 (new), REQ-CCR-011 and REQ-CCR-012 (reworded), §3 terms, §5 G9, §6 exclusions; design §5 (step `author`, `T-CHAIN` rows A1 to A3, the merge call, the allow-list), §6, §11; acceptance AC-CCR-006, 011, 012, 019 (rows added to existing criteria); decision-index Q26, Q27 | The workflow merged a pull request of any author, a bot and a fork-derived one included, and its merge call passed `--delete-branch` (E45); the closed text said bot pull requests are withheld unless the native review reviews them | Only a pull request authored by GoosLab (id 6458969, type User, all three together), from a branch of the base repository, is ever merged by automation, and only through every other gate; every other pull request (fork, bot, another author) needs the user's test confirmation and explicit approval and is never merged by automation; the workflow enforces this through a step `author` that records `eligible`, and the tests pin it. The merge call carries no `--delete-branch` and no step deletes a ref; `delete_branch_on_merge` stays an operator decision (G9) | leader addition 2 (policy alignment) |
| Δ26 | research E44; spec §5 DS-3; design provenance and reader rows 32 to 35; acceptance AC-CCR-001, 002, 004, 005 (subtests added) | Markers of one comment were assumed to agree; the effect of `@codex review` was unobserved | One re-request on #1842 was observed (read at 05:36Z): the state comment was edited in place, only the code leg moved to the new head, a marker-less receipt comment was added, the security leg stayed `completed` with a null commit, and the security marker kept the old head with status `running`: not a pass (stale). The receipt is ignored (no marker), a second comment carrying the marker still refuses. #1833 is a real findings case on the current head | leader addition (re-request observation) |
| Δ27 | spec REQ-CCR-004, REQ-CCR-005, §1, DS-1; design §3.1 (constant row), §3.2 (step 7a, step 9 note, decision rows 36 to 40), `T-CHAIN` row 2i, provenance rows `F-failed`, `F-failure-notice`; acceptance AC-CCR-004, 005, 019, 020 (rows added); research E46, U14; decision-index Q29 | A leg `failed` fell into the generic refusal `review-not-complete:<leg>:failed`, and its finding flags, read first, would have produced `finding:unresolved:<leg>` | A leg whose `status` is `failed` is a review run that failed: neither a pass nor a finding. It is refused with its own reason `review-failed:<leg>` as a gate refusal (step exit 1, job failure), checked right after the stale check and before any finding flag is read, also while the other leg is `running`, and the annotation names the re-request that starts a new run (`@codex review` for `code_review`, `@codex security review` for `security_review`, as the comment's own text names them) | leader addition (#1844 failed state) |
| Δ28 | acceptance AC-CCR-022 (class), AC-CCR-019 group E, design §6 residual 7 | AC-CCR-022 was release-blocking on a cell that only checks a renamed row; group E claimed an `updated_at` change "inside the recheck" is detected; residual 7 started after the recheck's second pull request read | AC-CCR-022 is a regression-guard (12 release-blocking, 13 regression-guard); group E states that only changes between the pass's listing and the recheck's re-listing are detected, and residual 7 starts the undetected window at the recheck's last list read | plan audit iteration 2 (N1, N2) |

**Correction of revision 8 (plan audit iteration 1 of revision 7, FAIL 0.69; each defect verified against the text and the raw files).** D1: REQ-CCR-001 now says once
that a review object or review comment can only refuse (REQ-CCR-005, term, AC-CCR-005, Q23 already said so). D2: AC-CCR-019 splits its fixtures by job result (gate
refusals and recheck races: failure; author policy: `eligible=false`, job success; merge-step guards: job success). D3 and D11: the class lists of acceptance.md are
re-derived from the cells that reproduce on this tree; cells R1 to R17 are history (ledger status); new live cells R24 to R27. D4: the recheck re-lists the issue comments
and compares the issue-comment count and the comment's id, `updated_at`, and markers, and fails on any difference (§5; Q25 restated; the "next pass" rationale is withdrawn).
D5: REQ-CCR-012 is scoped to `files` and `review`. D6: the #1842 move is worded as observed (the code leg named the new head at 05:29:23Z; the move was observed by the
05:32:18Z re-check); `F-moved-head` is labelled as a composed fixture. D7: the running #1841 snapshot is on disk at `/tmp/adk-resume-20261019-0510/pr1841-check.json`
(outside the directories the audit searched) and carries a running security marker; a running comment without a security marker was not observed, and the reader now
waits in that case (steps 6, 10, 13; an AC row). D8: stale old-contract text rewritten (§6 row, plan §A, §B, §G, decision-index Q19). D9: AC-CCR-019 states the `gh pr checks` stub and
pins the Q27 consequence. D10: DoD renumbered, gates G1-G9. D12: `related_specs` kept and annotated (the manager-spec agent definition lists it; the schema table does not), the REQ
order is by section and stated in spec §3.

**Correction of revision 9 (plan audit iteration 2 of revision 7, FAIL 0.88; revision 8 audited).** N1: AC-CCR-022 becomes a regression-guard (its RED cell R25 only checks a renamed row; the draft
behaviour is already implemented and tested), so the classes are 12 release-blocking and 13 regression-guard. N2: AC-CCR-019 group E states the exact timing, and residual 7 starts the
undetected window at the recheck's last list read. N3: the missing plan-time mutant reds stay a stated Gap. New input: a `failed` leg (PR #1844, E46) is a distinct refusal, `review-failed:<leg>`
(step 7a, row 2i, Δ27).

**Real-sample provenance.** Raw REST objects of five pull requests, read 2026-10-19, and the coordinator's audit (`research.md` E37 to E46; machine-local,
gitignored, public GitHub data). Only marker lines and table rows are quoted anywhere in the SPEC.

| Fixture family | Modelled on | Taken from the sample | Synthetic |
|----------------|-------------|-----------------------|-----------|
| clean pass (`F-clean`) | #1835 comment 6259114985, #1837 comment 6259113921, #1838 comment 6259347980, #1841 comment 6260188619 | author fields (login `chatgpt-codex-connector[bot]`, id 199175422, type Bot), the three marker lines in the leading block, both legs `completed`, `commit_sha` equal to the head, flags `false`, security marker `completed` with `headSha` | the head is the sample's head; the table rows are copied in shape only |
| running (`F-running`) | #1841, the saved GraphQL snapshot `/tmp/adk-resume-20261019-0510/pr1841-check.json` (machine-local, outside the card tree; the time 05:13:48Z comes from the observer's report, the file carries none; the REST comment later carries id 6260188619) | both legs `running`, `commit_sha` set, `has_any_findings` and `has_unresolved_findings` null, security marker `running` and present | the author object is the REST one of the clean samples; a variant with no security marker is synthetic (never observed, D7) |
| unresolved finding (`F-unresolved`) | #1836 comment 6259116571 | code_review `has_unresolved_findings` true, security_review false, both legs `completed`, `commit_sha` `e657c6b0401a3d5c3e92204c8382d9080c617f50` (head set equal to it) | none |
| stale (`F-stale`) | the same #1836 comment against the saved pull head `4cee5eea1ca9de02c6a7aef7409fbd2db80a2353` | stale-by-commit: both legs bound to `e657c6b040`, head `4cee5eea1c` | none |
| corroborating finding (`F-corrob`) | #1836 review 5566687300 (COMMENTED, body marker `codex-review-receipt:v1`) and review comment 4310160416 (commit_id equals original_commit_id `e657c6b040`) | author fields, state, commit binding | the head is set to `e657c6b0401a3d5c3e92204c8382d9080c617f50` to make them bound |
| completed leg with null fields (`F-null-leg`) | #1842 comment 6260255148 (REST id relayed; snapshot `pr1842-0526.json`), head `45b2dbf682515dd8f4d8d67ce0ae7ac2fa8328a5` | code_review `completed` bound to the head, flags false; security_review `completed` with `commit_sha`, `has_any_findings`, `has_unresolved_findings` null; security marker `running` | none |
| both running (`F-running-both`) | #1833 comment 6260252011, head `4bbd2ed113485c5bed3ea92299ff9b9754ca55d1` (read in the coordinator's audit, E43) | both legs `running`, unresolved null | none |
| security leg all null (`F-null-security-leg`) | #1839 comment 6259931511 and #1840 comment 6259960025 (E43) | code_review `completed` bound to the head, flags false; security_review `commit_sha`, `status`, and `has_unresolved_findings` all null | the security marker is set to `completed` for the head to isolate the refusal (it is not stated in the audit) |
| moved head (`F-moved-head`, a composed fixture) | #1842 comment 6260255148 as read at 05:24Z (code leg bound to `45b2dbf682515dd8f4d8d67ce0ae7ac2fa8328a5`) combined with the head `e7cceb2cb5d6e1cd9dbd2f465c4b9d5a1d631f21` observed by the 05:32:18Z re-check (E43); the comment itself was never read in that combination | code leg bound to the old head, security leg null, security marker `running` with `headSha` the old head | the head under review is the new SHA |
| mixed SHAs after a re-request (`F-mixed-sha`) | #1842 comment 6260255148 as read at 05:36Z (E44), head `e7cceb2cb5d6e1cd9dbd2f465c4b9d5a1d631f21` | code leg bound to the new head, security leg `completed` with null `commit_sha`, security marker `headSha` `45b2dbf682515dd8f4d8d67ce0ae7ac2fa8328a5` with status `running` | none |
| receipt comment (`F-receipt`) | #1842 comment 6260377784 (same author, no state marker; body not read) | author fields, absence of both markers | the body is synthetic prose; a variant that quotes the state marker prefix is synthetic |
| findings on the current head (`F-findings-1833`) | #1833 comment 6260252011 (E44) and the COMMENTED review 5567280902 of the identity bound to the head | code leg `completed`, bound to the head, `has_unresolved_findings` true; security leg and security marker clean for the head; the review's author fields, state, and commit binding | none |
| failed run (`F-failed`) | #1844 comment 6260734432, capture `pr1844-comments.json` observed 2026-10-19T06:14:10Z (E46) | `code_review` `failed` with `has_any_findings` true and `has_unresolved_findings` true, bound to `16d3d25ebfa5e235845babeea7c4c1b6e174ebe0` (the head); `security_review` completed, false, false, same head; security marker completed with the same `headSha`; the table row of the failed leg reads trigger "Manual request" | the composite `F-failed-running` (code `failed`, security leg and marker `running`) is synthetic: the leader first described it from an earlier read the lane did not capture; variants with a null `commit_sha`, a failed security leg, and an older head are synthetic |
| failure notice (`F-failure-notice`) | #1844 comment 6260920127 (the identity's own marker-less comment created 06:13:27Z; body paraphrased, not quoted: a notice that the review went wrong and may be retried by commenting) | author fields, absence of both markers | none |
| author policy (`F-author`) | the author objects of the coordinator's audit: GoosLab id 6458969 type User; dependabot[bot] id 49699333 type Bot (E43) | login, id, and type | the head repository and base repository ids (a fork has different ones, a deleted fork has none) and the lookalikes (login `GoosLab` with another id, id 6458969 with type Bot) are synthetic |
| other-author review (noise) | CodeRabbit review 5567249506 (COMMENTED, id 136622811, type Bot) and its inline comment on #1842 (E43) | author fields, state, commit binding | none |
| other authors (noise) | #1835 comments 6258776199 (`vercel[bot]` id 35613825), 6258776508 (`coderabbitai[bot]` id 136622811), 6258780085 (`github-actions[bot]` id 41898282) | author fields only | bodies are empty or synthetic |
| forged and tampered (`F-forged`, `F-tamper`) | none real | the author fields of the real identity with ONE field changed (login without `[bot]`, id other than 199175422, type `User`, user null); a marker line altered, duplicated, moved out of the leading block, or quoted inside prose | all of it |

**Not observed, so not claimed** (each is an item of `research.md` E39 and U13 to U16): the real run of this reader on #1828
itself (a draft; silent) and on any pull request; the whole `running` to `completed` transition (the two ends were read at
different times on #1841, `running` only in a saved snapshot); DS-2 (no automatic re-review on a push; one sample); the
effect of `@codex review` on the comment; the value set of `status`; a pull request opened non-draft; the default-branch run of
`auto-merge.yml`; whether `CI`, and so auto-merge, fires for a pull request into a `release/*` base.

**Conflicts between the leader's brief and the existing SPEC, or inside the brief, named and resolved as stated (not guessed).**
(a) The brief's fact (2), every final comment has `commit_sha` equal to the head "at that time", and fact (3), #1836's comment differs from
the final head, are read together as: equal to the head the sample saved for the four clean pull requests, `e657c6b040` against the saved
`4cee5eea1c` for #1836. (b) The shipped invariant "the script names no `body` key" (AC-CCR-001) contradicts a comment reader; it is replaced by
"the body is read for identity comments only, and only the marker lines" (Δ12). (c) The closed exclusion "reading the review from an issue
comment is out of scope" and the forbidden token `issues/` are reversed for the one listing (Δ20). (d) The brief leaves open what happens when a
leg is still `running` on the last pass: chosen as gate exit 1 `review-not-complete:running`, a failed step (it is neither absent nor a pass). (e)
The brief's addition 4 (#1842) supersedes the first default of Q24 (the security marker unread). (f) A completed leg with null fields beside a
`running` security marker is classed pending (it waits, then fails on the last pass), not refused at once: the brief says it must refuse a pass, and
a transient state is a wait. (g) The brief keeps reviews and review comments only as corroboration; kept as refusal-only inputs (Q23). (h) An issue-comment count
that moves during one pass is retried, not a failure, because other bots post comments constantly (Q25). (i) Addition 1 (the gate applies on every
base, an unprotected integration base included) meets the shipped `checks` step, which reads `gh pr checks --required` exiting 1 on an unprotected base as
a lookup failure and withholds (E43, E45): the SPEC keeps that step unchanged (Q22) and records the consequence, and Q27 names the alternate for
the leader. (j) The leader's author policy is not in the shipped workflow at all (E45); it is a new requirement, REQ-CCR-020, not a restatement.
(k) The leader's wording "unresolved 0" is read as `has_unresolved_findings` exactly `false` on both legs.

## Delta of plan revision 6 (read this first; a delta auditor needs only this section)

Plan revision 6 is an in-place amendment (spec.md `## Amendments`, version 0.5.0 to 0.6.0). Authority: the
leader's ruling, board record d-20261018T152011Z-8447 (supersedes d-20261018T151824Z-db98; ruling and scope unchanged), after the delegated Codex observer's reproduction of the `review` step
(`research.md` E33, E34): a review failure ended the step with exit status 0 and wrote `ok=false` / `state=fail`
only to the step's outputs, and a draft wrote the key `state` twice. Every clause this revision changed is
listed here with an id; nothing else in the five artifacts was changed except text that restates a listed
clause (each restatement names its id). **Frozen and untouched:** REQ-CCR-001 to 006, 009 to 011, 013 to 017;
the gate script's behavior and its exit classes 0, 1, 10, 11 (§3); the recognition contract (§3.1); the exit-0
convention of the steps `pr`, `head`, `checks`, and the merge step; every decision-index row Q1 to Q20 except the text named below.

| Id | Where | Before | After | Ruling item |
|----|-------|--------|-------|-------------|
| Δ1 | spec REQ-CCR-018 (new); design §5 table `T-CHAIN` | The `review` step withheld with exit 0 on every path; a refusal was a green job whose only trace was `ok=false` | One table: gate script exit code, to step exit status, to job result, to output values. Every path other than a clean pass ends the step non-zero, so the step conclusion and the job conclusion are failures and the merge step does not run. Widened by Δ11 to the `files` step | 1 |
| Δ2 | spec REQ-CCR-008, spec §2 item 4 and §3 term "NOT RUN"; design §1, §5 | NOT RUN was one result with two causes, neither a failure nor a pass, ending the step with exit 0 | At the consumer the two causes split. Draft: single state `not-run:draft`, step exit 11, job result not success. Non-draft, review absent at the window's end: a failure (`state=fail`, step exit 1, reason `not-run:review-absent`). The gate script's exit class 11 keeps both causes | 2, 3 |
| Δ3 | spec REQ-CCR-019 (new); design §5 `T-OUTPUTS` | Both new steps wrote `ok=false` and `state=fail` first, then appended `ok=true` / `state=pass` or `state=not-run` | Each step records each output key at most once, on its single terminating path; `state` takes one of `pass`, `fail`, `not-run:draft`; an unrecorded key reads as not-pass | 4 |
| Δ4 | spec REQ-CCR-007, REQ-CCR-012 | Refusal and unreadable-at-expiry "withhold" | The same results, now stated as a failed step (REQ-CCR-018); REQ-CCR-012's "failing result" for a head that moved is a non-zero `review` step | 1 |
| Δ5 | acceptance AC-CCR-020 to 025 (new); AC-CCR-007, 008, 012, 019 (changed rows); matrix, counts, ledger R13 to R16 | 19 criteria; two rows asserted exit status 0 (`not_run_step_exit_status_0_with_notice_annotation`, `fail_step_exit_status_0_with_warning_annotation`), and rows of AC-CCR-012 (`files_step_*_not_ok`) asserted `ok` only | 25 criteria; those two rows are replaced; every step-level row asserts the step exit status and the derived job result, not only the outputs | 5 |
| Δ6 | spec §4 (new bullet), §5 G6 acceptance condition, design §11, §12, acceptance §6, §7 | "Remote verified" and the draft observation were worded without the job conclusion | The remote conclusion of the job is UNOBSERVED; a draft pull request's CI success cannot show that the amended `auto-merge.yml` runs on the default branch (`workflow_run` runs the default-branch copy); that is observable only on the next pull request after the change lands, and until then the remote state is "unverified" | 6, G6 |
| Δ7 | spec §5 DS-1, DS-2 rows; research E33, U11, U12 | DS-1 and DS-2 mixed the artifact form, the draft gate observation, and the ready-for-review transition | DS-1: the review actually runs on a draft, and a clean result stays as an APPROVED review (unobserved, not success). DS-2 (leader's item): whether a review runs again after ready-for-review or after a push. The gate's behavior on a draft head moved into the post-landing job observation of Δ6 | DS-1, DS-2 |
| Δ8 | spec §7 KD-1, decision-index "Follow-up candidates" and "Known debt" | KD-1 recorded as an unissued candidate | KD-1 is transferred to card t1874 (blocked until t1868 is merged-local and DS-1 is observed on the G6 draft pull request) | KD-1 |
| Δ9 | decision-index Q21, Q22 (new), Q9 and Q15 (restated) | none | Q21: how a draft's job result is not success (default: the step exits 11; alternate: a pre-job skips the job). Q22: the other steps keep their exit-0 withhold | scope |
| Δ10 | spec §6 (two new Out of Scope bullets) | none | Exit status of the steps `pr`, `head`, `checks`, and the merge step is unchanged; a two-job split is not built | scope |
| Δ11 | leader scope addition, board record d-20261018T155527Z-4e9e (a delegated Codex observer run, not an operator run and not a new approval): spec REQ-CCR-018 (both steps), §2 item 5, §6; design `T-CHAIN` rows 10 to 13, §5; acceptance AC-CCR-025 (new), AC-CCR-012 `files_step_*` rows, AC-CCR-019 chain rows; research E36 | Revision 6 first limited the non-zero rule to `review`; the `files` step, which runs first and gates `review`, still ended exit 0 on a missing gate script, a refusal, or a gate-path change (measured, E36), so a failure caught there read as a green job | The rule covers BOTH steps this SPEC added and the chain's conclusion: a failure or a missing gate script in `files` ends that step and the job as not-success; the tests run the chain `files`, `review`, merge through failure paths, gate-script-missing included, and assert each step's exit status and the job result. The earlier steps `pr`, `head`, `checks` stay out of scope. Consequence: a pull request that changes a gate path now shows a failed job as well as a withheld merge | chain |

**Correction (D1 to D4), plan delta audit iteration 6 (FAIL 0.80), one correction commit approved by board
record d-20261018T160907Z-9211; wording, mutant design, and AC structure only, no new contract, identifiers
unchanged.** D1: the `files` step description of §5 now agrees with `T-CHAIN` rows 10 to 13 (exit 1 on every
failure, 0 on pass only). D2: the AC-CCR-019 gate-path mutant is redesigned as MF and traced (§5 "wiring" bullet
and `acceptance.md` §2.5); the executed red belongs to the run phase's M5 mutant run. D3: the review-step
reproduction is attributed to the delegated Codex observer, and the effective ruling is cited as board record
d-20261018T152011Z-8447 (supersedes d-20261018T151824Z-db98; ruling and scope unchanged). D4: Q21 records that the
red job cost covers every non-pass, not only drafts. Correction (a), board record d-20261018T161841Z-894f: the AC-CCR-019 review-term-drop mutant is reclassified as equivalent and covered structurally by AC-CCR-012; two wrapped attribution wordings and the Q21 authority anchor are fixed; no mutant, harness behavior, id, or contract text was added. Exception correction (D8, D9), board record d-20261018T162703Z-95a3: the Q21 authority anchor is paraphrased from boards 8447 and 4e9e without quotation, the job-level-skip example is no longer attributed to a board record, and acceptance §6 item 2 carries the review-term-drop mutant in its equivalent-mutant carve-out.

Conflicts with the existing SPEC found while applying the ruling, resolved as the Delta states: (a) the
revision-4 rule "NOT RUN is neither a failure nor a pass, ends the step with exit 0" is withdrawn for the
`review` step (Δ1, Δ2); (b) the revision-4 DS-2 wording is replaced by the leader's DS-2 and its draft-gate
half moves to the post-landing observation (Δ7); (c) the "every guard step ends with exit 0" convention of §5
now holds for `pr`, `head`, `checks`, and the merge step only; the `files` step follows the `review` step (Δ11).

How the requirements in `spec.md` are met. File and script names appear here and in `plan.md`, not in
`spec.md`. Evidence ids (E*, U*) refer to `research.md`. Plan revision 7 replaces the review reader of revisions 3 to 6 (a pull-request review object in state APPROVED) with the
comment protocol (§3.1, §3.2, `T-CHAIN`); where §3 and §5 below describe an APPROVED review, the revision-7 text of §3.1, §3.2,
and `T-CHAIN` governs. Plan revision 4: the reviewer is the ChatGPT
subscription's native Codex GitHub review. Revision 4 changes the review reader (§3.2: APPROVED is the only
clean state, one supersession rule, body text never read), names one NOT RUN result (§3.2, §5), binds
changed-file entries to the verified head and reads the head before and after each list (§3.4, §5), and adds
the merge decision path as a criterion of its own (§5, AC-CCR-019). The keyed `openai/codex-action` workflow,
its verdict job, its schema, and its notice job of revisions 1 and 2 are not designed here (§4). The
doctrine rewrite of revision 1 stays out of the card (spec §6, §7). Revision 6 (the Delta above) turns every
non-pass of the `review` step into a failed step and a failed job (§5, `T-CHAIN`), splits NOT RUN by cause at
the consumer, and makes each output key single-valued (`T-OUTPUTS`); the gate script itself is unchanged.

## 1. Shape of the solution

One script and one edited workflow. The reviewer is outside the repository; the gate only reads what it
leaves on the pull request and refuses everything it cannot recognise.

```
review starts (open for review, draft marked ready, or an `@codex review` comment; a draft gets none)
   chatgpt-codex-connector[bot] writes ONE issue comment and edits it in place until the review ends   (E37)

[auto-merge.yml]  (workflow_run of CI, default-branch copy; every merge base, none exempt)
   step "files":  read the PR object (head, changed_files); every page of the changed-file list, count-checked
                  and bound entry by entry to the verified head; read the PR object again;
                  any change under .github/workflows/ or to the gate script, an unbound or unprovable list,
                  or a moved head -> withhold
   step "review": per pass: read the PR (draft, head, comments, review_comments), every page of the issue comments
                  (count-checked against `comments`), of the reviews, and of the review comments (count-checked
                  against `review_comments`), read the PR again, then read-native-review
        (every line below is the gate script's result; the step's own exit status, job result, and outputs for
         each are the one table T-CHAIN of section 5)
        draft                                             -> exit 11 not-run:draft: state not-run:draft, step exit 11
        exactly one state comment of the identity, both legs completed, bound to H, has_unresolved_findings
        exactly false, security marker completed for H, no corroborating finding on H
                                                          -> exit 0, then the pre-merge recheck of the SAME comment:
                                                             step exit 0; the merge step reads the head again before the call
        no state comment of the identity, window open     -> wait (exit 10), not a terminating path
        no state comment, window closed                   -> exit 11 not-run:review-absent: a FAILURE, step exit 1
        a leg (or the security marker) running, window open -> wait (exit 10, pending:running)
        a leg running on the last pass                    -> exit 1 review-not-complete:running: step exit 1
        a leg or the security marker bound to another commit -> stale-review (exit 1): step exit 1 at once
        any other status, null leg fields, unresolved findings, a corroborating finding, a malformed, duplicated or
        misplaced marker, two state comments, a marker disagreement -> fail (exit 1): step exit 1 at once
        a list that stays unreadable to the end of the window -> fail review-unreadable: step exit 1
```

Identity and commit are the only keys. Nothing a user can write (an issue comment, a label, review body
text) and no review by another author or app is read as a verdict, the gate never infers approval from
text, and time is only ever a bound that withholds: no elapsed time and no absence is a pass.

## 2. Components and owners

| Component | Path | Milestone |
|-----------|------|-----------|
| Gate script (three subcommands) | `scripts/ci/codex-review-gate.sh` | M1 (read-native-review), M2 (fetch-pages, gate-files-changed) |
| Consumer | `.github/workflows/auto-merge.yml` (edited) | M3 |
| Retirements | `.github/workflows/claude.yml`, `.github/workflows/review-quality-gate.yml` (deleted) and their two liveness entries (removed) | M4 |
| CodeRabbit off | `.coderabbit.yaml` (two flags) | M4 |
| Contract guard | `internal/config` test over `required-checks.yml`; `git diff` preserve proof | M4 |

No workflow is added, no secret is referenced, and no component is added under `internal/cli`; no non-test
Go source file changes. The script is plain `bash` plus `jq` and stays compatible with bash 3.2 (no
associative arrays, no `mapfile`) because the tests also run on darwin. The script is the only code that
talks to the GitHub API (through `gh api`, GET only), so a test replaces `gh` with a stub on `PATH` and
exercises every call.

## 3. Gate script contract

Invocation: `scripts/ci/codex-review-gate.sh <subcommand> [flags]`. A refusal prints one line to stderr
beginning `CODEX_GATE_REFUSED: <reason>`; stdout carries the human-readable summary. Exit codes are the
contract the workflow branches on; a caller treats every exit it does not recognise as a refusal. Exit
classes: **0** pass, **1** fail, **10** pending (no review yet, wait window open), **11** NOT RUN. The four are
pairwise distinct. At the script level NOT RUN is one exit class with exactly two causes, `not-run:draft` and
`not-run:review-absent`; both 1 and 11 withhold the merge, and the consumer never reads 11 as 0. **Revision 6
left this script contract untouched; revision 7 replaces the recognition contract and the reasons of `read-native-review` (§3.1, §3.2) and adds the
issue-comment listing to `fetch-pages` (§3.3), and keeps the four exit classes, the subcommand names, `gate-files-changed`, and the consumer mapping**; what changed is the
consumer's reading of the exit classes: the `review` step turns every exit other than 0 into a non-zero step
exit, the draft cause keeps its own state, and the review-absent cause is a failure (§5, `T-CHAIN`).

### 3.1 The recognition contract (constants in the script)

| Constant | Value | Evidence |
|----------|-------|----------|
| Identity login | `chatgpt-codex-connector[bot]` | E27, E37 |
| Identity numeric user id | `199175422` | E27, E37 |
| Identity account type | `Bot` | E27, E37 |
| Recognised verdict carrier | an issue comment (`issues/N/comments`) of the identity whose body carries the state marker prefix | E37 (revision 7); a review object is not a carrier |
| State marker | the literal prefix `<!-- codex-review-state:v1 ` on a whole line of the leading block, closed by ` -->` at the end of that line, between them one JSON object | E37 |
| Security marker | the literal prefix `<!-- codex-security-review:v1 ` on a whole line of the leading block, closed by ` -->`, one JSON object | E37, E42 |
| Leading block | the run of lines from line 0 of the body on that each begin with `<!--` and end with `-->` (observed: lines 0 to 2) | E37 |
| Fields read | state marker: `code_review` and `security_review`, each `status`, `commit_sha`, `has_unresolved_findings`; security marker: `status`, `headSha`; comment: `id`, `user.login`, `user.id`, `user.type`, `body`, `updated_at`. Nothing else: `has_any_findings`, `blockingSeverityThreshold`, `mergeGateEnabled`, `pullRequestNumber`, `repository`, the summary marker, the table, the prose, and the reactions are never read | E37, E42 |
| Pass | both legs: `status` `completed`, `commit_sha` equals the head (40 hex, case-insensitive), `has_unresolved_findings` is the boolean `false`; the security marker `completed` with `headSha` equal to the head; exactly one state comment; no corroborating finding | ruling; E42 |
| Failed run | any leg with `status` `failed` (a review run that failed): refused as `review-failed:<leg>` right after the stale check, before any finding flag is read; the flags of a failed leg are never interpreted | E46 |
| Pending | any leg or the security marker with `status` `running` (the observed null fields while running are not an error) | ruling; E39 |
| Corroborating finding | a review comment of the identity whose `commit_id` or `original_commit_id` equals the head (or carries neither field), or a review of the identity bound to the head (`commit_id` equals the head) in state COMMENTED or CHANGES_REQUESTED; it can only refuse | ruling ("may only corroborate"); Q23 |
| Changed-file entry binding | for an entry whose `status` is not `removed`, the `ref` value in the query of `contents_url` equals the verified head SHA | E31, U10 |
| Allowed list endpoints | `repos/<owner>/<repo>/pulls/<N>/(files\|reviews\|comments)` and `repos/<owner>/<repo>/issues/<N>/comments`, and nothing else | REQ-CCR-006, AC-CCR-006 |
| Gate paths | any path under `.github/workflows/`; `scripts/ci/codex-review-gate.sh` | spec §3 terms |

An artifact is recognised as the identity's only when login, id, and type all match: a lookalike login, the same login with another id, a
`User` account of that login, or a null `user` is not the identity (the `[bot]` suffix cannot be registered by a user account, and the
numeric id is immutable). A comment of any other author is ignored entirely, whatever markers its body carries: the exclusion is by author and
never by body. Exactly one identity is recognised. A marker string inside a user's comment is therefore never read, and a marker string inside
the identity's own comment but outside the marker position (in the prose, a table cell, the details block) is a refusal, not a second reading
(`state-marker-misplaced`, `security-marker-misplaced`). The contract is stated from five real comments (E37 to E39) and one inconsistent one
(E42); the value set of `status` beyond `running` and `completed` is unknown (U14), so any other value refuses.

### 3.2 `read-native-review --issue-comments I --reviews F --comments G --head-sha S --draft true|false [--wait-closed true|false]`

`I`, `F`, and `G` are NDJSON files (one object per line) written by `fetch-pages` (§3.3): the pull request's issue comments, reviews, and review
comments. Read-only. `--wait-closed` (default `false`) says that the caller's wait window has ended; it changes two results only (steps 4 and 10).
Checked in this order; the first matching step decides:

0. **Flags.** `--draft` or `--wait-closed` neither `true` nor `false` → exit 1 `draft-flag-unusable` / `wait-closed-flag-unusable`; `S` not 40
   hexadecimal characters → exit 1 `head-sha-unusable`.
1. `--draft true` → exit 11 `not-run:draft`, before any file is read (REQ-CCR-008).
2. `I`, `F`, or `G` missing, or a line that is not a JSON object → exit 1 `input-invalid`.
3. **Candidates**: issue-comment entries whose `user.login`, `user.id`, and `user.type` all equal the identity AND whose string `body` contains the
   state marker prefix at least once. Entries of any other author, and comments of the identity without the prefix, are ignored entirely. A
   candidate lacking an integer `id` or a string `updated_at`, or two candidates with the same `id` → exit 1 `state-comment-invalid:<id>`.
4. **Absence.** No candidate: `--wait-closed false` → exit 10 `pending:no-state-comment`; `--wait-closed true` → exit 11 `not-run:review-absent`
   (REQ-CCR-008).
5. **One comment.** More than one candidate → exit 1 `state-comment-duplicated:<ids>`; the gate never selects by recency (REQ-CCR-004).
6. **Parse the markers** of the single candidate. State marker: the prefix must occur exactly once in the body, else exit 1
   `state-marker-duplicated`; that occurrence must be a whole line of the leading block, else exit 1 `state-marker-misplaced`; the text between the
   prefix and the closing ` -->` must be one JSON object with object-valued `code_review` and `security_review`, else exit 1
   `state-marker-malformed`. Security marker: zero occurrences → recorded as absent, decided at steps 10 and 13 (a leg still `running` waits; otherwise exit 1
   `security-marker-missing`); two or more → `security-marker-duplicated`; one
   outside the leading block → `security-marker-misplaced`; not one JSON object → `security-marker-malformed`.
7. **Stale.** Any leg whose `commit_sha` is a non-null string that is not equal to `S`, or a security marker whose `headSha` is a non-null string
   that is not equal to `S`, → exit 1 `stale-review:<commit prefix>` (REQ-CCR-002; a comment exists, so this is a failure and never pending
   and never NOT RUN, whatever `--wait-closed` says). A `commit_sha` or `headSha` that is neither a string nor null → exit 1
   `review-entry-invalid:<leg>:commit_sha`.
7a. **Failed run.** A leg whose `status` is `failed` → exit 1 `review-failed:<leg>`, whatever its `has_any_findings` and `has_unresolved_findings` read (a failed run is not a finding), whatever the
    other leg reads (the comment cannot pass without a new run, so the gate does not wait), and also when its `commit_sha` is null; the annotation names the re-request that starts a new
    run (`@codex review` for `code_review`, `@codex security review` for `security_review`, the triggers the comment's own text names, E46; that the second one re-runs the security leg was
    not observed). A `failed` leg bound to another commit never reaches this step: step 7 refuses it as stale.
8. **Corroborating findings** (refusal only): a review comment of the identity whose `commit_id` or `original_commit_id` equals `S` → exit 1
   `finding:<id>`; one carrying neither field → exit 1 `finding-unbound:<id>`; a review of the identity with `commit_id` equal to `S` in state
   COMMENTED or CHANGES_REQUESTED → exit 1 `finding-review:<id>`. Artifacts bound only to another commit are stale findings of an older head and are
   ignored. No review or review comment can pass anything (REQ-CCR-005).
9. **Unresolved findings reported by the marker** (a `failed` leg never reaches this step, step 7a). A leg with `has_unresolved_findings` equal to `true` → exit 1
   `finding:unresolved:<leg>` (checked before the pending rule so a known finding refuses at once).
10. **Pending.** Any leg `status` equal to `running`, or the security marker `status` equal to `running` (a security marker that is not present yet does not stop the wait
    while a leg is `running`: the order is wait, never refuse, because a pending state cannot pass and the marker may be written later; a running comment without the
    security marker was never observed, so this is a decision, not a measurement): `--wait-closed false` → exit 10
    `pending:running`; `--wait-closed true` → exit 1 `review-not-complete:running`. A `completed` leg with null fields beside a `running` security
    marker (E42) lands here, never at a pass.
11. **Status.** A leg `status` other than `completed` (any other string, null, or missing) → exit 1 `review-not-complete:<leg>:<status>`.
12. **Binding of a completed leg.** A leg whose `commit_sha` is null → exit 1 `review-not-bound:<leg>`; a leg whose `has_unresolved_findings` is not
    a boolean (null included; null is not false) → exit 1 `review-entry-invalid:<leg>:has_unresolved_findings`.
13. **Security marker agrees.** A security marker recorded as absent at step 6 → exit 1 `security-marker-missing` (reached only when no leg is `running`); `status` other than `completed` (a running one was handled in step 10) → exit 1 `security-marker-disagrees:<status>`;
    a null `headSha` → exit 1 `security-marker-disagrees:no-head`.
14. Otherwise exit 0 and print `native-review-clean: comment <id> updated <updated_at> code_review:<S prefix>:completed:unresolved=false
    security_review:<S prefix>:completed:unresolved=false security-marker:<S prefix>:completed head <S prefix> markers <compact JSON of the state marker>
    <compact JSON of the security marker>`; the recheck of §5 compares this line byte for byte.

**The one rule, in a sentence.** The result is clean only when the identity's single state comment shows both legs `completed`, each bound to the
head, each `has_unresolved_findings` exactly false, and the security marker agrees, with no corroborating finding; the decision reads author,
marker position, marker fields, and the corroborating artifacts, never the prose, the table, the reactions, or any review object as a pass.

**Reader decision table** (H is the head, O another commit; "window" is the `--wait-closed` value; the full chain from each row to the step exit, the job
result, and the outputs is `T-CHAIN` of §5):

| # | Comment state of the identity (fixture family) | Result | Reason |
|---|------------------------------------------------|--------|--------|
| 1 | both legs completed, bound to H, flags false, security marker completed for H (`F-clean`) | exit 0 | `native-review-clean` |
| 2 | the same, header and table prose saying anything, summary marker absent or altered | exit 0 | prose and the summary marker are never read |
| 3 | both legs `running`, null fields, security marker `running` (`F-running`, `F-running-both`), window open | exit 10 | `pending:running` |
| 4 | the same, window closed | exit 1 | `review-not-complete:running` |
| 5 | no comment of the identity, window open / closed | exit 10 / exit 11 | `pending:no-state-comment` / `not-run:review-absent` |
| 6 | code leg completed bound to H, security leg completed with null `commit_sha` and flags, security marker running (`F-null-leg`), window open | exit 10 | `pending:running` |
| 7 | the same with the security marker completed | exit 1 | `review-not-bound:security_review` |
| 8 | both legs completed bound to O (`F-stale`) | exit 1 | `stale-review:<O prefix>` |
| 9 | one leg bound to H, the other to O | exit 1 | `stale-review:<O prefix>` |
| 10 | security marker `headSha` O, legs bound to H | exit 1 | `stale-review:<O prefix>` |
| 11 | a leg `has_unresolved_findings` true, bound to H (`F-unresolved`) | exit 1 | `finding:unresolved:code_review` |
| 12 | a leg completed with `has_unresolved_findings` null | exit 1 | `review-entry-invalid:<leg>:has_unresolved_findings` |
| 13 | a leg `status` `cancelled`, an unknown string, null, or missing | exit 1 | `review-not-complete:<leg>:<status>` |
| 14 | a leg object missing (`code_review` or `security_review`) | exit 1 | `state-marker-malformed` |
| 15 | marker JSON altered: truncated, trailing text, a key of the wrong type | exit 1 | `state-marker-malformed` |
| 16 | the state marker twice in one body | exit 1 | `state-marker-duplicated` |
| 17 | the state marker inside the identity's prose or table, beside or instead of the real one | exit 1 | `state-marker-misplaced` / `state-marker-duplicated` |
| 18 | two comments of the identity carrying the marker, one clean | exit 1 | `state-comment-duplicated:<ids>` |
| 19 | a user's comment (login lookalike, same login with id other than 199175422, type `User`, user null) carrying a clean marker, no real comment | window open: exit 10; closed: exit 11 | `pending:no-state-comment` / `not-run:review-absent` |
| 20 | the same user comment beside a real `running` comment | exit 10 | the user's comment is never a candidate |
| 21 | a user's comment carrying a clean marker beside a real stale comment | exit 1 | `stale-review:<O prefix>` (the forgery fills nothing) |
| 22 | security marker duplicated, misplaced, or malformed; or missing while no leg is `running` | exit 1 | `security-marker-duplicated` / `-misplaced` / `-malformed` / `security-marker-missing` |
| 23 | security marker `completed` while a leg reads running | exit 10 | `pending:running` |
| 23a | a leg `running` and the security marker not present (never observed; the order of steps 6 and 10), window open / closed | exit 10 / exit 1 | `pending:running` / `review-not-complete:running` |
| 24 | clean comment plus a review comment of the identity bound to H, or a COMMENTED or CHANGES_REQUESTED review of the identity bound to H (`F-corrob`) | exit 1 | `finding:<id>` / `finding-review:<id>` |
| 25 | clean comment plus an APPROVED, DISMISSED, or PENDING review of the identity, or a finding bound to O only | exit 0 | review objects and old findings neither pass nor refuse |
| 26 | clean comment, no `reviews` and no review comments (the four clean real pull requests) | exit 0 | the normal case |
| 27 | pull request is a draft | exit 11 | `not-run:draft` |
| 28 | a list cannot be proven complete: a page error, a count mismatch, a full 30th page, a repeated entry | `fetch-pages` exits 1 and removes the file; the consumer counts the pass as pending and, if still unreadable at the end of the window, reports a failure `review-unreadable` | `page-error`, `count-mismatch`, `page-ceiling`, `page-duplicate` |
| 29 | the API silently shortens a list with no error and no count to compare (the reviews list) | not detectable | residual risk, §6 (U9) |
| 30 | a state recorded for the old head against the new head: code leg bound to `45b2dbf682515dd8f4d8d67ce0ae7ac2fa8328a5`, security leg null, security marker `running` with that `headSha`, head `e7cceb2cb5d6e1cd9dbd2f465c4b9d5a1d631f21` (`F-moved-head`) | exit 1 | `stale-review:45b2dbf682` (stale precedes pending) |
| 31 | code leg completed bound to H, security leg with `status`, `commit_sha`, and flags all null (`F-null-security-leg`), security marker not `running` | exit 1 | `review-not-complete:security_review:null` |
| 32 | code leg bound to H, security leg `completed` with null `commit_sha`, security marker `headSha` O with status `running` (`F-mixed-sha`, the #1842 state after the re-request) | exit 1 | `stale-review:45b2dbf682` (the security marker names another commit; stale precedes pending) |
| 33 | a clean state comment plus a second comment of the identity with no state marker (`F-receipt`) | exit 0 | the receipt is not a candidate |
| 34 | a second comment of the identity that carries the state marker prefix (a receipt quoting the marker) beside the real one | exit 1 | `state-comment-duplicated:<ids>`; with no real marker, `state-marker-misplaced` |
| 35 | code leg `has_unresolved_findings` true bound to H, security clean, plus a COMMENTED review of the identity bound to H (`F-findings-1833`) | exit 1 | `finding-review:<id>` (step 8); without the review object, `finding:unresolved:code_review` |
| 36 | code leg `failed` with both finding flags true, bound to H; security leg completed, false, false; security marker completed for H (`F-failed`, #1844) | exit 1 | `review-failed:code_review` (never `finding:unresolved:code_review`) |
| 37 | code leg `failed`, security leg and security marker `running` (`F-failed-running`), window open | exit 1 | `review-failed:code_review` at once, no wait |
| 38 | security leg `failed`, code leg completed and clean | exit 1 | `review-failed:security_review` |
| 39 | a leg `failed` with a null `commit_sha` | exit 1 | `review-failed:<leg>` |
| 40 | a leg `failed` bound to O (the head moved) | exit 1 | `stale-review:<O prefix>` (stale precedes the failed check) |

Rows 1 to 27 and 30 to 40 are fixtures of AC-CCR-001 to AC-CCR-005 and AC-CCR-008; row 28 is AC-CCR-007 and AC-CCR-010; row 29 is stated, not tested. Exit
classes: 0 pass, 1 fail, 10 pending, 11 NOT RUN; pairwise distinct. The conditions the operator named map as follows, each never a pass: **missing
review** → exit 10 while the window is open, then exit 11 `not-run:review-absent`; **running** → exit 10 while the window is open, then exit 1
`review-not-complete:running`; **missing output** (a missing leg, a null or missing field, a missing security marker) → exit 1; **failed review**
(an unresolved finding, a status other than completed, a corroborating finding) → exit 1; **stale SHA** (a leg or the security marker bound to another
commit) → exit 1 `stale-review` (`acceptance.md` §3.1).

**Residual shape risk, stated and fail-closed.** If the native review adds a `status` value or another marker version, the reader refuses (a marker
`v2` carries another prefix, so the comment is not a candidate and the pull request ends `not-run:review-absent`); the amendment of the contract is
then a follow-up, never a runtime option. If the review reports a clean result in a form that is not the state comment, nothing passes.

### 3.3 `fetch-pages --endpoint E --out F [--expect-count N]`

The only subcommand that reads from the API, and the only place pagination lives. `E` is a repository
relative API path whose response is a JSON array; `F` receives one JSON object per line.

0. **Endpoint allow-list.** `E` must match `^repos/[A-Za-z0-9._-]+/[A-Za-z0-9._-]+/(pulls/[0-9]+/(files|reviews|comments)|issues/[0-9]+/comments)$`
   (revision 7 adds the issue-comment listing, the verdict carrier); anything else → exit 1 `endpoint-not-allowed:<E>` before any `gh` call, so no
   check-run, status, label, or reaction endpoint can be read through the script (AC-CCR-006).
1. For page P = 1, 2, … it runs `gh api "<E>?per_page=100&page=<P>"`. A non-zero exit of `gh` refuses with
   `page-error:<P>`; a response that is not an array of objects refuses with `page-invalid:<P>`.
2. A page with fewer than 100 entries ends the collection. A page with exactly 100 entries asks for the next
   page, so a listing whose size is an exact multiple of 100 ends on an empty page.
3. The collection never asks past page 30. When page 30 is full (3000 entries) it refuses with
   `page-ceiling`: the files listing's own limit is 3000 entries (U3), so a full 30th page cannot be proven
   complete and is never read as complete.
4. **Count proof.** With `--expect-count N` (a non-negative integer, else `expect-count-unusable`) a
   collected count different from `N` refuses with `count-mismatch:<collected>:<N>`. The consumer passes the
   pull request's own `changed_files` for the files listing, `review_comments` for the review-comment
   listing, and `comments` for the issue-comment listing (equal to the list length on all five measured pull requests, E37); the two counts agree with the paginated lists on every pull request measured (E22, E23). The
   reviews listing has no count on the pull request object, so it is collected without a count proof and its
   completeness rests on rules 1-3 and 5 (U9).
5. **Entry identity.** An entry's key is its integer `id` (issue comments, reviews, review comments) or its string `filename`
   (changed files); an entry with neither refuses `page-invalid:<P>`, and a key that repeats anywhere in the
   collection refuses `page-duplicate:<key>`. A repeat is the signature of a list that shifted between two
   page reads (an entry deleted and another added); a count comparison alone cannot see that case.
6. On any refusal `F` is removed, so a caller that ignores the exit status finds no list to read. On success
   it prints `fetched:<count> count-proof:<yes|none>`.

Why an explicit loop rather than `gh api --paginate`: the files listing returns 30 entries per page by
default (E22), a piped `--paginate` hides a mid-pagination failure unless every pipeline sets `pipefail`,
and a stub `gh` in a test cannot reproduce `--paginate`'s own behavior. The explicit loop uses only the
documented `per_page` and `page` parameters and is exercised offline page by page (AC-CCR-010).

### 3.4 `gate-files-changed --file FILES --head-sha S`

`FILES` is the NDJSON stream of the pull request's changed-file objects written by `fetch-pages`. Fields
read: `filename`, `status`, `contents_url`, and, when present, `previous_filename`. Read-only. Checked in
this order:

0. `S` that is not 40 hexadecimal characters, or no `--head-sha` → exit 1 `head-sha-unusable`.
1. `FILES` missing or empty → exit 1 `file-list-unusable`.
2. A line that is not an object with a string `filename` → exit 1 `file-entry-invalid`.
3. 3000 or more entries → exit 1 `file-list-at-api-ceiling` (defense in depth; `fetch-pages` refuses first).
4. Any `filename` or `previous_filename`, compared lower-cased, that starts with `.github/workflows/` or
   equals `scripts/ci/codex-review-gate.sh` → exit 1 `gate-path-changed:<path>`. This covers any added,
   edited, deleted, or renamed workflow file, not only `auto-merge.yml`, because a workflow added by a pull
   request could merge itself or act with the workflow token.
5. **Binding to the verified head.** Every entry whose `status` is not `removed` (a missing `status` counts
   as not removed) must carry a string `contents_url` whose `ref` query value equals `S` (case-insensitive);
   otherwise → exit 1 `file-entry-unbound:<filename>`. An entry of a list read while the head was another
   commit `T` carries `T`, so a list read in the window of a head that flipped from `H` to `T` and back is
   refused instead of being judged on `T`'s files (D14). Removed entries are not bound: their `contents_url`
   points at the base commit (E31), which is the named residual of §6.
6. Otherwise exit 0.

A pull request that changes a gate path is merged by a human; the human merge path is unaffected (§11).

## 4. What is not built: no review workflow

The keyed design of revisions 1 and 2 — a `pull_request` workflow running `openai/codex-action` with an
`OPENAI_API_KEY` secret, a verdict job publishing a check run named `Codex Review`, a verdict schema, a
key-presence step, SHA-pinned actions, and a keyless notice job for drafts and forks — is rejected
(decision-index Q1): the operator's existing subscription already reviews every pull request, a key would
add a secret, a cost, and a gate (G1) for a capability the repository already has. Consequences the
corrected design draws:

- No workflow is added, so there is no new liveness entry (the census needs only the two removals), no
  `MOAI_*` live-axis re-measure, and no `actionlint` target other than `auto-merge.yml`.
- No secret is referenced by anything this SPEC adds. Fork pull requests expose no secret because there is
  none to expose; the repository rule that its workflows use no `pull_request_target` family trigger for
  anything this SPEC touches is pinned by AC-CCR-013.
- No artifact is uploaded, downloaded, or cached; the lists the gate reads are temporary files in
  `$RUNNER_TEMP` of a single job (AC-CCR-013).
- The forged-carrier surface of the keyed design (a same-named check run from another workflow) does not
  exist: the verdict is an artifact authored by an external app identity that a repository workflow cannot
  impersonate.

## 5. Consumer changes (`auto-merge.yml`)

- Permissions are unchanged (`contents: write`, `pull-requests: write`); reading reviews, review comments,
  and changed files needs only `pull-requests`. No other permission is added.
- The CodeRabbit step (status wait, comment read, risk ladder, `AUTOMERGE_MAX_MERGE_RISK`) is deleted.
- **Step conclusion convention (revision 6, Δ1 to Δ3, widened by the leader's scope addition, board record
  d-20261018T155527Z-4e9e).** The steps `pr`, `head`, `checks`, and (revision 7) `author` keep the convention of the closed SPEC:
  they end with exit 0, an annotation, and an output, so a withheld merge by one of them is a green job whose
  log and annotation say what happened. **The two steps this SPEC adds, `files` and `review`, do not.** A
  failure in either — a missing gate script included — is a failure of the step and of the job: each step ends
  with exit 0 on its pass alone, and every other terminating path ends it non-zero. (Measured: the `files` step
  body run without the gate script exits 0 with `ok=false state=fail`, research E36.) For `review`: the step
  ends with exit 0 on the clean pass alone, and every
  other terminating path ends it non-zero, so the step conclusion and the job conclusion are failures and the
  merge step (whose `if` has no status function, so it implies `success()`) does not run. The step is not
  `continue-on-error` and nothing exempts it from the job's conclusion (REQ-CCR-018). Each of the two new
  steps records each output key once, on its single terminating path, with no default written first
  (REQ-CCR-019). The single table:

  **`T-CHAIN` — comment state, to gate script exit, to step exit, to job result, to outputs** (revision 7 folds the reader's decision into this one table; every terminating path of the
  two steps this SPEC adds: rows 1 to 9 are the `review` step, rows 10 to 13 the `files` step; the gate
  script's exits are unchanged)

  | # | Comment state / terminating path | Gate exit (reason) | Step exit | Job result | `ok` | `state` | Annotation |
  |---|-----------------------------------|--------------------|-----------|------------|------|---------|------------|
  | 1 | Clean state comment (reader rows 1, 2, 26 of §3.2): both legs `completed`, bound to the head, `has_unresolved_findings` false, security marker `completed` for the head, no corroborating finding; and the pre-merge recheck re-reads the comments and reads the same comment id, `updated_at`, fields, head, and review-comment count | 0, then 0 with the same summary line | 0 | success; the merge step may run | `true` | `pass` | none beyond the clean summary line |
  | 2a | Stale: a leg or the security marker bound to another commit (reader rows 8 to 10, 21) | 1 `stale-review:<prefix>` | 1 | failure | `false` | `fail` | `::error::native review gate: <reason> - merge withheld` |
  | 2b | Findings: `has_unresolved_findings` true; a corroborating review comment or COMMENTED or CHANGES_REQUESTED review (rows 11, 24) | 1 `finding:unresolved:<leg>`, `finding:<id>`, `finding-unbound:<id>`, `finding-review:<id>` | 1 | failure | `false` | `fail` | `::error::` with the reason |
  | 2c | Not complete or not bound: a leg `status` other than `completed` and other than `failed` (unknown, null, missing); a completed leg with a null `commit_sha`; a non-boolean `has_unresolved_findings` (rows 7, 12, 13) | 1 `review-not-complete:<leg>:<status>`, `review-not-bound:<leg>`, `review-entry-invalid:<leg>:<field>` | 1 | failure | `false` | `fail` | `::error::` with the reason |
  | 2d | Tampered or malformed: marker JSON altered, a leg missing, a marker duplicated or outside its position, a security marker malformed, duplicated, or misplaced, or missing while no leg is `running` (rows 14 to 17, 22) | 1 `state-marker-malformed`, `-duplicated`, `-misplaced`; `security-marker-missing`, `-malformed`, `-duplicated`, `-misplaced` | 1 | failure | `false` | `fail` | `::error::` with the reason |
  | 2e | Marker disagreement: the security marker reads another status than `completed` while the legs are completed and nothing is running | 1 `security-marker-disagrees:<status>` | 1 | failure | `false` | `fail` | `::error::` with the reason |
  | 2f | More than one state comment of the identity, or an identity candidate without an `id` or `updated_at` (row 18) | 1 `state-comment-duplicated:<ids>`, `state-comment-invalid:<id>` | 1 | failure | `false` | `fail` | `::error::` with the reason |
  | 2g | A leg or the security marker still `running` on the last pass (row 4) | 1 `review-not-complete:running` | 1 | failure | `false` | `fail` | `::error::native review gate: review-not-complete:running - merge withheld` |
  | 2h | Unusable input or flag (`input-invalid`, `draft-flag-unusable`, `wait-closed-flag-unusable`, `head-sha-unusable`) | 1 | 1 | failure | `false` | `fail` | `::error::` with the reason |
  | 2i | Failed run: a leg `status` `failed`, whatever its finding flags read and whatever the other leg reads (rows 36 to 39 of §3.2) | 1 `review-failed:<leg>` | 1 | failure | `false` | `fail` | `::error::native review gate: review-failed:<leg> - merge withheld; re-request the failed leg (code_review: @codex review; security_review: @codex security review)` |
  | 3 | The gate exits a code that is none of 0, 1, 10, 11 | other | 1 | failure | `false` | `fail` | `::error::` naming `gate-exit-<n>` |
  | 4 | Non-draft, no state comment of the identity when the window ends (reader row 5) | 11 (`not-run:review-absent`) | 1 | failure | `false` | `fail` | `::error::native review gate: not-run:review-absent - no state comment observed for the pull request` |
  | 5 | Draft pull request | 11 (`not-run:draft`) | 11 | failure; never success (decision-index Q21) | `false` | `not-run:draft` | `::notice::NOT RUN: draft - not a review result and never a pass; the merge is withheld` |
  | 6 | The head or the review-comment count moved before or during the list reads (`head-moved`, `head-moved-during-read`, `review-comments-moved-during-read`) | not reached | 1 | failure | `false` | `fail` | `::error::` with the reason |
  | 7 | The pre-merge recheck differs (`recheck-unreadable`, `head-moved-before-merge`, `review-comments-moved-before-merge`, `issue-comments-moved-before-merge` (the issue-comment count differs between the recheck's two reads, or the re-listed issue comments disagree with it), `review-changed-before-merge` (the summary line differs: comment id, `updated_at`, or any field, a `running` result included), `selected-review-changed-before-merge` (another state comment is selected)) | 0, then another exit or a different summary | 1 | failure | `false` | `fail` | `::error::` with the reason |
  | 8 | A list is unreadable on the last pass, so the loop ends with no stop (`review-unreadable`) | not reached or `fetch-pages` refused | 1 | failure | `false` | `fail` | `::error::` with the reason |
  | 9 | The gate script is missing (`gate-script-missing`) | not run | 1 | failure | `false` | `fail` | `::error::` with the reason |
  | — | Gate exit 10 with the window open (not a terminating path): `pending:no-state-comment`, `pending:running`; or an issue-comment count that moved between the two reads of one pass (note `issue-comments-moved`, decision-index Q25) | 10 | none: sleep 15 seconds, next pass | — | not recorded | not recorded | `native review not observed yet (n/60)` or `native review running (n/60)` |

  | 10 | `files`: the changed-file list passes (every read agrees, `fetch-pages` 0, `gate-files-changed` 0) | 0, 0 | 0 | success; the `review` step may run | `true` | `pass` | none |
  | 11 | `files`: the gate script is missing (`gate-script-missing`) | not run | 1 | failure | `false` | `fail` | `::error::changed-file check: gate-script-missing - merge withheld` |
  | 12 | `files`: the pull request is unreadable or the head or count moved (`files-unreadable`, `head-moved:<prefix>`, `files-unreadable-after-read`, `head-moved-during-read`, `count-moved-during-read`) | not reached | 1 | failure | `false` | `fail` | `::error::` with the reason |
  | 13 | `files`: `fetch-pages` or `gate-files-changed` refuses (`page-error`, `count-mismatch`, `gate-path-changed:<path>`, `file-entry-unbound:<name>`, any refusal) or exits a code it does not recognise | 1 or other | 1 | failure | `false` | `fail` | `::error::` with the reason |

  Step `author` (REQ-CCR-020; it keeps the exit-0 convention, Q26, so these are not failures of the chain):

  | # | Terminating path of `author` | Step exit | Job result | `eligible` | Annotation |
  |---|------------------------------|-----------|------------|------------|------------|
  | A1 | GoosLab (id 6458969, type User), head repository equals the base repository | 0 | success; the later steps run | `true` | none |
  | A2 | Any other author, a bot, a fork or a deleted fork, a lookalike (`GoosLab` with another id, the id with another type) | 0 | success; `checks`, `files`, `review`, and the merge step are skipped | `false` | `::notice::` needs the user's test confirmation and explicit approval, never merged by automation |
  | A3 | The pull request object is unreadable or empty (`author-unreadable`) | 0 | success; the later steps are skipped | `false` | `::warning::author-unreadable - merge withheld` |

  After a `files` failure the `review` step is skipped (its `if` requires `files.ok == 'true'` and, being
  status-function-free, implies `success()`), so the job result is the `files` failure and no review is read.
  The job result is "failure" because a step ended non-zero without `continue-on-error`; nothing in this table
  says what GitHub reports on a runner (research U11: unobserved). Rows 2 to 4, 6 to 9, and 11 to 13 are the
  failures the rulings name; row 5 is the draft that must not read as a pass; rows 1 and 10 together are the
  only success path of the chain (`files` then `review`, then the merge step). The earlier steps `pr`, `head`,
  and `checks` and the merge step keep their exit-0 convention (Out of Scope).

  **`T-OUTPUTS` — the output keys** (REQ-CCR-019)

  | Step | Key | Allowed values | Recorded |
  |------|-----|----------------|----------|
  | `files` | `ok` | `true`, `false` | once, on the step's terminating path |
  | `files` | `state` | `pass`, `fail` | once, with `ok` (`true`/`pass` or `false`/`fail`) |
  | `review` | `ok` | `true`, `false` | once, on the terminating path; `true` only on row 1 |
  | `review` | `state` | `pass`, `fail`, `not-run:draft` | once, with `ok` (`true`/`pass`, `false`/`fail`, `false`/`not-run:draft`) |

  A key the step never recorded (a step killed by the job timeout) reads empty, and every consumer compares
  `ok` with `'true'`, so an unrecorded key is not-pass. No default (`ok=false`, `state=fail`) is written ahead
  of the terminating path, because a later append of the same key is the duplicate the delegated Codex observer's
  reproduction showed.
- Three steps this SPEC adds run in this order: `author` (after the head re-verification, before the required-check observation), then, after the
  required-check observation and before the merge, `files` and `review`:
  0. **Step `author`** (name `Decide author eligibility`; revision 7, REQ-CCR-020). Its `if` requires the head re-verification to have passed. It
     reads the pull request object with one `gh api` call (`repos/$REPO/pulls/$PR_NUMBER`, the call form already allowed) and records the single
     output `eligible`, once: `true` only when `user.login` equals `GoosLab`, `user.id` equals `6458969`, `user.type` equals `User`, `head.repo`
     is not null, and `head.repo.id` equals `base.repo.id`; `false` on every other result, an unreadable or empty object included (so a bot,
     a fork, a deleted fork, another author, and a lookalike with the right login and another id are all `false`). It ends with exit 0 in every
     case (the convention of `pr`, `head`, and `checks`; decision-index Q26), with a `::notice::` saying that the pull request needs the user's
     test confirmation and explicit approval and is never merged by automation, and a `::warning::` naming `author-unreadable` when the object
     could not be read. The `if` of `checks`, `files`, `review`, and the merge step each carry `steps.author.outputs.eligible == 'true'`, so
     removing the term from any one is caught structurally (AC-CCR-012) and by behaviour where it changes a merge (AC-CCR-019). Because
     `workflow_run.pull_requests` can be empty for a fork, the earlier skip step may end such a run first; that is not relied on (E45).
  1. **Step `files`** (name `Collect and check the changed-file list`). Its `if` requires the head
     re-verification and the required-check observation to have passed. (a) It reads the pull request object
     with one `gh api` call — `head.sha` and `changed_files`; an unreadable or empty read, or a `head.sha`
     other than the verified head SHA of step `head`, ends the step with `ok=false` (`files-unreadable`,
     `head-moved:<prefix>`). (b) It runs `fetch-pages` on the files listing with `--expect-count` set to that
     `changed_files`. (c) It reads the pull request object again: a `head.sha` other than the verified SHA
     ends the step with `ok=false` `head-moved-during-read`, and a `changed_files` other than the first
     read's with `count-moved-during-read`. (d) It runs `gate-files-changed --file F --head-sha <verified
     SHA>`. It writes `ok=true` only when (a) to (d) all pass; any other exit writes `ok=false` with the
     reason (each key once, `T-OUTPUTS`); the step ends with exit 1 on every failure path (`T-CHAIN` rows 11 to 13) and with exit 0 only on `ok=true` (row 10). A
     missing script (the consumer's checkout is the default branch) is a refusal, not a pass.
  2. **Step `review`** (name `Read the native Codex review`). Its `if` additionally requires
     `steps.files.outputs.ok == 'true'`. It loops at most 60 times, 15 seconds apart, with no sleep after the
     60th pass (the bound the CodeRabbit step used; time is only this bound). Each pass: (a) read the pull
     request object (`draft`, `head.sha`, `comments`, `review_comments`; a failed or empty read counts as pending
     with the note `unreadable`); (b) **if `draft` is `true`, call `read-native-review --draft true` at once,
     before any list is fetched**, so a draft ends as `not-run:draft` whatever a list fetch would return; (c) a
     `head.sha` other than the verified SHA ends the step in failure, `head-moved` (`T-CHAIN` row 6); (d) run
     `fetch-pages` on `issues/<N>/comments` with `--expect-count` set to `comments`, on `.../reviews` (no count),
     and on `.../comments` with `--expect-count` set to `review_comments` (a refusal counts as pending with the
     note `unreadable:<reason>`); (d2) **read the pull request object again, after the list reads and before the
     gate script runs**: a `head.sha` other than the verified SHA ends the step in failure,
     `head-moved-during-read` (the lists just read belong to another head and are never judged; `T-CHAIN` row 6),
     a `review_comments` other than the first read's ends it in failure `review-comments-moved-during-read`, a
     `comments` other than the first read's makes the pass pending with the note `issue-comments-moved` (other
     bots post comments constantly and the count carries no verdict; decision-index Q25), and a failed or empty
     second read counts as pending with the note `unreadable`; (e) run `read-native-review` with the verified head
     SHA of step `head`, the `draft` flag, and `--wait-closed true` on the 60th pass only and `false` on every
     earlier pass. Mapping of the exit (the full table is `T-CHAIN`): **0** → the pre-merge recheck, then step
     exit 0 with `ok=true`, `state=pass` (row 1). **10** → sleep 15 seconds, next pass. **11** → stop at once: for
     a draft, `state=not-run:draft` and step exit 11 (row 5); for a non-draft pull request, `state=fail` and step
     exit 1 with the cause `not-run:review-absent` (row 4). **1, or any exit not listed** → `ok=false`,
     `state=fail`, the reason, step exit 1, stop at once (rows 2a to 2i and 3). If the loop ends with no stop —
     only possible when the last pass was unreadable or pending for a moved count — the step ends in failure,
     `review-unreadable: <last note>` (row 8): a list that could not be proven complete is a refusal
     (REQ-CCR-010), never a pass. A leg still `running` on the 60th pass is gate exit 1
     `review-not-complete:running`, produced by the script, and absence on the 60th pass is gate exit 11
     `not-run:review-absent`, produced by the script and not by the loop, so there is one place where each
     is decided. **The recheck (REQ-CCR-012; revisions 7 and 8, because the bot edits the comment in place and a second state comment can appear inside the interval).**
     Before `ok=true`, the step reads the pull request object (its `comments` and `review_comments` counts and its head), re-lists the issue comments, the
     reviews, and the review comments (the issue-comment and review-comment lists each count-proofed against that first read), reads the pull request object
     once more, and runs `read-native-review` again with `--wait-closed false`. It requires, and fails the step on any difference (`T-CHAIN` row 7): the same head
     in both reads, the same `comments` count in both reads (so a comment added after the list collection is caught by the second read, and one added before
     it by the count proof), the same `review_comments` count, and the identical summary line of exit 0 (comment id, `updated_at`, and the compact JSON of both
     markers), so a comment edited or replaced in the interval, a second state comment, or a result that is no longer clean ends the step as a failure. The
     loop's own passes still retry a moved issue-comment count (Q25); only the recheck refuses it, because there is no later pass.
- **The merge step reads the head again immediately before the merge call** (`gh pr view ... headRefOid`, unchanged since the
  closed SPEC) and withholds the merge when it differs from the verified head; `--match-head-commit` then pins the call itself. Every
  leg's `commit_sha` and the security marker's `headSha` were compared with that same verified head, so they equal the head read right
  before the call or the merge is withheld; a state recorded for an old head (the #1842 move) never counts for the new one.
  Revision 7 makes this a criterion row (AC-CCR-019 `head_moves_between_recheck_and_merge_call_zero_merge_calls`). The scope
  has no base-branch term: the workflow's `if` names none, so the review gate applies to every merge base, an unprotected
  `release/*` integration base included (AC-CCR-019 `release_base_*` rows).
- The merge step's `if` replaces `steps.coderabbit.outputs.ok == 'true'` by
  `steps.author.outputs.eligible == 'true' && steps.files.outputs.ok == 'true' && steps.review.outputs.ok == 'true'`; the existing terms, the
  `--match-head-commit`, and the deadline re-evaluation are untouched; the merge call drops `--delete-branch` (REQ-CCR-021); the `release/*`
  merge-flag logic (which reads the head branch name) is untouched.
- The checkout step keeps no `ref`. The step named `Check all required CI checks passed` keeps its name so
  `TestAutoMergeRequiredChecks` still finds it.
- **The `gh` allow-list is exact** (REQ-CCR-006, AC-CCR-006). The workflow makes exactly these calls and no
  others: `gh pr view "$PR_NUMBER" --json headRefOid` (steps `head` and merge), `gh pr checks "$PR_NUMBER"
  --required --json name,state,bucket` (step `checks`, the unchanged CI observation), `gh pr merge
  "$PR_NUMBER" … --match-head-commit …` (merge; no `--delete-branch`, REQ-CCR-021), and `gh api` calls whose only path is the
  pull request object `repos/$REPO/pulls/$PR_NUMBER` (steps `author`, `files`, and `review`); the gate script makes
  exactly one `gh` call form, `gh api "<E>?per_page=100&page=<P>"`, where `E` passed the endpoint check of
  §3.3. The retained `gh pr` forms observe CI check state and the head; none derives a review verdict.
- **The wiring and the merge decision are criteria.** AC-CCR-012 parses the real `auto-merge.yml`, asserts the
  step order, the step conditions, and the merge condition, and runs the two real steps against a stub `gh` and
  a real or stub script copy, so a deleted guard step, an ignored exit code, or a merge that does not depend on
  the gate turns a named subtest red. AC-CCR-019 runs the chain `head → checks → files → review → merge
  condition → merge step` end to end on fixtures with a stub `gh` that records every call: `gh pr merge` is
  called exactly once, pinned to the verified head, on the clean fixture, and zero times when the gate exits 1
  or 11, the window expires, the exit is unknown, the head moved, or the deadline passed — so a merge step
  that never merges, or one that merges past a refusal, turns a named subtest red. Removing the `files` term
  alone from the `review` step's `if`, or from the merge `if`, is not observable by behaviour under the
  revision-6 rules: a failed `files` step skips every later step whose `if` has no status function (the
  implied `success()`), so the merge stays unreachable and the recorded merge calls are 0 in the original and
  the mutant. Those removals are pinned structurally by AC-CCR-012 (`review_step_depends_on_files_ok`,
  `merge_if_requires_files_and_review_ok`). AC-CCR-019's own mutant for the gate-path row, **MF**, drops the
  `files` term from both conditions and leads both with `always()` (a status function, which lifts the implied
  `success()`); that changes behaviour, and the step-by-step trace is in `acceptance.md` §2.5 "Mutant MF
  trace". The harness's `if:` evaluator admits an optional leading `always()` for exactly this purpose.
- **Step exit and job result are criteria, not log text** (revision 6, Δ5). The step harness of the tests
  records, for every step it runs, the exit status and the output file, and derives the job result from the
  executed steps: failure when a step that is not `continue-on-error: true` exits non-zero, and the later steps
  whose `if` has no status function skipped; success otherwise. AC-CCR-020 to AC-CCR-023 assert the step exit
  status and the derived job result on every terminating row of `T-CHAIN` and the output keys of `T-OUTPUTS`;
  AC-CCR-024 pins the harness itself (a comparison that matches the outputs but not the exit status, or the exit
  status but not the job result, must fail) and rejects `continue-on-error` and status-function overrides on the
  `review` step. The derived job result is a rule over the workflow's own structure; what GitHub reports on a
  runner is a separate, unobserved fact (spec §4, research U11).

## 6. Threat model and fail-closed matrix

| Input or condition | What a forged or stale input could do | Control | Criterion |
|--------------------|----------------------------------------|---------|-----------|
| **A comment by any user carrying the marker text** (issue comment, review body, review comment) | A forged clean state read as the review | The verdict carrier is a comment whose author's login, id, and type all equal the Codex identity; any other author's comment is never a candidate, whatever markers it carries; no prose of any artifact is parsed | AC-CCR-003, AC-CCR-006 |
| **A comment by a different author or app** (another bot such as `coderabbitai[bot]`, `vercel[bot]`, a human) | A clean-looking state from the wrong author | Login, numeric id, and account type must all equal the Codex identity; another author's artifact is ignored in every state | AC-CCR-003 |
| **A lookalike**: the same login without `[bot]`, the same login with another id, a User account of that login | Impersonation | The three-part identity; `[bot]` logins cannot be registered by users; the id is immutable | AC-CCR-003 |
| **A state comment bound to another commit** (older head, prefix only, empty, null) | Pass a review written for earlier code | The `commit_sha` of each leg and the `headSha` of the security marker must equal the full head; a non-null other value is `stale-review`; null is bound to nothing | AC-CCR-002 |
| **A push after the review started** (no automatic re-review on a push was observed, DS-2) | The old state reads as current | Same binding: the moved head has no bound state → `stale-review`, withheld at once; fresh review is DS-3 | AC-CCR-002, AC-CCR-007 |
| **A marker altered, duplicated, or planted in the identity's own prose or table; two state comments** | A tampered or ambiguous state read as clean | One state marker on a whole line of the leading block, parsing to the required shape; a second occurrence, a misplaced one, or a second state comment refuses, and no recency is used | AC-CCR-004 |
| **A leg `completed` with null fields, a security marker that disagrees, an unknown `status`** (E42) | `completed` alone, or two legs reading `completed`, read as a pass | Each leg needs `completed`, a `commit_sha` equal to the head, and `has_unresolved_findings` exactly false; the security marker must read `completed` for the head; a running state waits and fails on the last pass | AC-CCR-004 |
| **A review `running` or absent when the merge is attempted** (the incident of E41) | A merge over an unfinished review, on any base | `running` waits and fails on the last pass; absence waits and ends exit 11; the merge step re-reads the head before the call; no base is exempt | AC-CCR-007, AC-CCR-008, AC-CCR-019, AC-CCR-021 |
| **A pull request of a bot, a fork, or another author** (dependabot, a fork-derived branch, someone else) | Automation merges a change that only the user's test confirmation and explicit approval may merge (E45: the shipped workflow had no author rule) | Step `author` records `eligible=true` only for GoosLab (id 6458969, type User) with a head repository equal to the base repository; every later step and the merge `if` require it | AC-CCR-012, AC-CCR-019 |
| **Automatic branch deletion** | A merge that deletes a branch the operator did not decide to delete | The merge call carries no `--delete-branch` and no step deletes a ref; the repository setting stays an operator decision (G9) | AC-CCR-006, AC-CCR-011, AC-CCR-019 |
| **Markers of one comment carrying different SHAs** (the #1842 re-request, E44) | The code leg for the new head passes while the security marker still names the old head | Every leg's `commit_sha` and the security marker's `headSha` must equal the head; any other value is `stale-review`; a null security `commit_sha` is never bound | AC-CCR-002, AC-CCR-004 |
| A finding: `has_unresolved_findings` true, or a review comment or a COMMENTED or CHANGES_REQUESTED review of the Codex identity on the head | Merge past a reported problem | Refuse; severity is not parsed; a review object or review comment can only refuse | AC-CCR-005 |
| **PR adds or edits any workflow, or the gate script** | The PR changes the gate for itself, or adds a workflow that merges itself or acts with the workflow token | Tamper guard withholds on ANY file under `.github/workflows/` and on the gate script; the consumer runs from the default branch, never the PR ref | AC-CCR-009, AC-CCR-011 |
| **A gate-path edit past the first page of the changed-file list** | The edit on page 2 or later is never read (default page size 30, E22) | `fetch-pages` collects every page at 100 per page, requires the collected count to equal `changed_files`, refuses at the 3000 ceiling | AC-CCR-010 |
| **The head flips from H to T and back while the changed-file list is read** | The list of T (a safe change) is judged and H (a workflow change) is merged | Entries that are not removals must carry `ref` = the verified head SHA in `contents_url`; the head is read before and after the list and must equal the verified SHA | AC-CCR-009, AC-CCR-012 |
| **A list API fails mid-pagination, is truncated, or shifts between pages** | A partial list read as complete | A page error refuses and removes the output; count proof for files and review comments; a repeated entry refuses; the step records `ok=true` only on exit 0 | AC-CCR-010, AC-CCR-012 |
| **The consumer calls the guard and ignores the result, or never calls it** | The guard exists but binds nothing | Structural test of the real step order, exit-code handling, and merge condition, with mutants | AC-CCR-012 |
| **The merge step never merges, or merges past a refusal** | A `&& false` in the merge condition, an early `exit 0`, an unbounded deadline, or a dropped term in the condition goes unnoticed | The chain from the head step to the merge step runs on fixtures; `gh pr merge` is called once, pinned to the verified head, on the clean fixture and zero times on every refusal | AC-CCR-019 |
| **The gate is pointed at another data source** | A verdict read from a check run, a status, a label, a reaction, a review object, or prose | The workflow's `gh` calls are an exact allow-list; `fetch-pages` accepts only the files, reviews, review-comments, and issue-comments listings of a pull request, and the one verdict carrier is the state marker of the identity's comment (the issue-comment listing exists for that marker only) | AC-CCR-006 |
| Draft pull request | A work-in-progress merged, or a draft observation read as a review result or as a green job | Exit 11 at once, whatever artifacts exist: state `not-run:draft`, step exit 11, job result never success; the message says gate path only | AC-CCR-008, AC-CCR-022 |
| **Elapsed time or the absence of a review read as success** | A review that never came is treated as "nothing to fix" | Time only bounds a wait; the end of the window is a failed step `not-run:review-absent` (gate exit 11) or `review-unreadable`, never a pass; the expiry fixtures assert that no merge happens | AC-CCR-007, AC-CCR-008, AC-CCR-019, AC-CCR-021 |
| Label, check run, commit status, reaction, or review object read as a verdict | Pass the consumer without a state comment | The consumer reads only the four listings; labels, statuses, check runs, and reactions are never read (census) | AC-CCR-006 |
| `pull_request_target` or a secret in a workflow this SPEC edits | Secrets with attacker-controlled code | The `workflow_run` trigger alone, no secret reference other than the workflow token (none added) | AC-CCR-013 |
| Auto-merge checks out the PR ref | Attacker code runs with a write token | Checkout has no `ref` | AC-CCR-011 |
| **Artifact or cache poisoning** | Not applicable: nothing is uploaded, downloaded, or cached; the lists are temporary files of one job | No surface exists; a test finds no artifact or cache step | AC-CCR-013 |
| A review arrives after the consumer's window | A late review never read | Withhold as a failed step `not-run:review-absent`; re-trigger paths (accepted limitation AL-1) | AC-CCR-007, AC-CCR-008, AC-CCR-021 |
| **A review failure the step swallows** (revision 6: the step wrote `ok=false` and exited 0) | A refused review, a changes-requested review, or a draft shows as a green job, so a human or a tool reading the job conclusion reads a pass | Every non-pass path ends the step non-zero (`T-CHAIN`), the step is not `continue-on-error`, and the tests assert step exit status and derived job result, not only the outputs | AC-CCR-020, AC-CCR-024 |
| **A key written twice in a step's output** (revision 6: `ok=false` then `ok=true`, `state=fail` then `state=not-run`) | A reader that takes the first value, or a platform that takes the last, reads a different verdict than the step's own terminating path | Each key recorded once on the terminating path, no default first (`T-OUTPUTS`) | AC-CCR-023 |

**Residual risk that no mechanical check closes** (each stated, none a pass-by-absence):

1. The native review's own judgment: a model can miss a defect or be manipulated by pull-request content, and a state comment
   that reports both legs completed with no unresolved finding passes whatever the prose says (a contract choice: the prose is
   unobserved structure and a parser would be a guess).
2. A clean result in a form other than the state comment, or a marker `v2`, a new `status` value, or a changed field name, is
   never recognised, so the gate may withhold forever; the observation of the new form and an amendment of the contract are the
   remedy (U14), never a runtime option.
3. The reviews listing has no count on the pull request object (U9): a silent shortening by the API with no
   error, which could hide a later review, is not detectable; the page rules (a page error refuses, a full 30th
   page refuses, a repeated entry refuses) are the only controls.
4. The changed-file binding does not cover removed entries (their `contents_url` points at the base commit,
   E31): a head that flipped from `H` to a `T` whose list holds only removals with the same count as `H`'s, and
   back, within one list read, would be judged on `T`. It needs push access and precise timing; the before and
   after head reads and the count proof narrow it, and nothing else closes it.
5. **CI-input files outside the tamper guard.** The guard covers `.github/workflows/` and the gate script.
   `.github/test-input-filters.yml` is read by the pinned `dorny/paths-filter` step (`ci.yml:76-79`) and decides
   `go_code`, which gates the test jobs (`ci.yml:106`); `scripts/ci/retry-flaky.sh` and
   `scripts/ci-census/test-census.sh` process the test results (`ci.yml:222-226`). A pull request can reduce
   what the required checks verify without touching a gate path. This is a pre-existing CI-integrity limit,
   not the review gate; widening the guard is a separate decision (follow-up FU-11).
6. The human merge path is as unguarded as it is today (E6: no review check in protection) — gates G3 and G4
   are the lever and stay with the operator.
7. A few seconds remain between the recheck's last list read and the merge call itself: an in-place edit of the state comment, or a dismissal, that lands in that interval is not detected (the lists
   were read before it, neither the issue-comment count nor a review count moves on an edit, and the step holds no later value to compare); the recheck detects an edit only between the pass's
   listing and its own re-listing (AC-CCR-019 group E). The merge is pinned to the head with `--match-head-commit`, so the code under review cannot change in the interval.
8. **The comment body is editable by a user with write access.** The bot edits its own comment, and a maintainer can edit
   it too; the author field stays the bot's. The gate cannot tell the two edits apart. A well-formed edit that keeps both markers
   consistent passes. The author binding keeps a pull-request author without write access, a fork author, and every other bot out; a
   user with write access can already merge. The recheck of the same `updated_at` narrows the window but cannot close it.
9. **A re-request may create a second state comment.** If `@codex review` after a head change creates a second comment instead of
   editing the first (U13; one re-request on #1842 edited the state comment in place and added a marker-less receipt comment, E44), the gate fails closed on more than one state comment and the pull request is withheld for
   a human merge. Choosing the newest by recency is rejected by the ruling.
10. **The merge may precede the comment.** On four of the five measured pull requests the state comment was created after the pull
   request had already merged (E41). A merge by a human or by anything other than this gate is outside what the gate controls; the
   gate only guarantees that its own merge never happens over a running or absent review.

## 7. Retirement set and census rule

Deleted: `claude.yml`, `review-quality-gate.yml`. Edited: `.coderabbit.yaml` (`reviews.auto_review.enabled:
false`, `chat.auto_reply: false`; every other key kept so a later re-enable is a two-line revert). Census
rule used by AC-CCR-014: no file under `.github/workflows/` contains `anthropics/claude-code-action` or a
top-level `name:` of `Claude Code Review`, and none triggers on `check_run`. The remote-only
`claude-code-review.yml` satisfies none of that locally; the census exists so that merging it back in turns
the suite red before the merge (E5). The deletions and the two manifest removals (§8) move in the same
milestone because an existing census test fails on any half-done state (E12).

## 8. Liveness manifest

`.moai/guards/liveness.yaml` is read by an existing test that requires the manifest and the files
`.github/workflows/*.yml|*.yaml` to be equal as sets in both directions, with no duplicate locator
(`internal/guardstate/census_test.go:49-91`, E12). Because this SPEC adds no workflow, the only effect is
the removal: deleting `claude.yml` and `review-quality-gate.yml` (M4) turns that test red until the same
milestone removes their entries (lines 48 and 146 at the measured tree). The gate's continued-firing signal
is the existing `auto-merge.yml` entry (line 34), which this SPEC leaves in place. The new criterion test
lives beside the census in `internal/guardstate` and reuses its repo-manifest loader.

## 9. Doctrine is left unchanged (known debt KD-1)

The deployed doctrine, templates, projections, and the lane notice that describe the CodeRabbit read are
not edited by this SPEC (leader ruling (b)). Nothing in M1-M4 depends on them: no requirement or criterion
claims they are updated, and no code path reads them. The sentences that go stale, the tests that pin them,
the interim effect, and the exit condition are in `spec.md` §7 (KD-1) and `research.md` E20; the follow-up
rows are in `decision-index.md` § Follow-up candidates. The follow-up rewrite would state the predicate "a
pull request's review is read from the single state comment of the Codex app identity, both legs `completed` and bound to the
current `headRefOid`, no unresolved finding, the security marker agreeing, and no corroborating finding on it", which the
consumer of §5 implements (the comment form was read on five pull requests, `research.md` E37).

## 10. Test layout

All new tests are `_test.go` files, none under `internal/template/templates/`, so the neutrality guard does
not scan them.

| File | Tests |
|------|-------|
| `internal/template/codex_review_gate_test.go` | `TestCodexGateReadNativeReviewAccepts`, `TestCodexGateReadNativeReviewRefusesStale`, `TestCodexGateReadNativeReviewIgnoresForged`, `TestCodexGateReadNativeReviewRefusesIncomplete`, `TestCodexGateReadNativeReviewRefusesFinding`, `TestCodexGateReadNativeReviewPending`, `TestCodexGateReadNativeReviewDraftNotRun` |
| `internal/template/codex_review_paging_test.go` | `TestCodexGateFetchPagesCompleteness`, `TestCodexGateFetchPagesRefusesDisallowedEndpoint`, `TestCodexGateFilesChangedRefusals` |
| `internal/template/auto_merge_codex_test.go` | `TestAutoMergeReadsOnlyAllowedReviewArtifacts`, `TestAutoMergeWaitLoopIsBounded`, `TestAutoMergeNotRunWithholds`, `TestAutoMergeCheckoutAndGuardsUnchanged`, `TestAutoMergeWiresGuardAndGate`, `TestAutoMergeAddsNoSecretTriggerOrArtifact`, `TestAutoMergeMergeDecisionPath`; revision 6 adds `TestAutoMergeReviewFailureFailsStepAndJob`, `TestAutoMergeReviewAbsentIsFailure`, `TestAutoMergeDraftIsNotRunNeverSuccess`, `TestAutoMergeGuardStepsOutputOneValuePerKey`, `TestAutoMergeStepHarnessObservesExitAndJobResult` |
| `internal/template/review_retirement_test.go` | `TestNoClaudeReviewWorkflows`, `TestCodeRabbitDisabledAndUnread` |
| `internal/config/required_checks_review_guard_test.go` | `TestRequiredChecksCarryNoCodexReviewCheckUntilConfirmed` |
| `internal/guardstate/census_review_test.go` | `TestCensusOmitsRetiredReviewWorkflows` |

Fixtures are inline strings and `t.TempDir()` files built in the shape of the real REST issue-comment object (E37, the
provenance table of the Delta) and the documented review and review-comment objects (E23); the test function names and
subtests are listed in `acceptance.md`; the tests copy the real script and run it with `bash`. The tests that exercise
`fetch-pages` and the two auto-merge steps put a stub `gh` first on `PATH` that serves fixture pages by the
`page` parameter, can fail a chosen page, and a no-op `sleep`, the pattern `TestAutoMergeRequiredChecks`
already uses. `TestAutoMergeMergeDecisionPath` additionally records every `gh` call of the stub, evaluates
the `if:` conditions with a small evaluator limited to the grammar the workflow uses
(`steps.<id>.outputs.<name> == '<literal>'` terms joined by `&&`, optionally led by the status function `always()`, which the harness reads as the exemption from the implied `success()`; a condition outside that grammar fails the
test, so a `&& false` or any other new shape cannot pass unnoticed), and runs the real merge step with the
`merge_deadline` environment value set to the past or the future. **Revision 6:** the step harness shared by
these tests returns, for each run of a step, the exit status, the output file, and the job result it derives
from the executed steps (`§5`, "Step exit and job result are criteria"), and every terminating-row assertion
compares all three; a row that compares only the outputs is the defect the amendment closes (AC-CCR-024). The
two rows of the closed SPEC that asserted exit status 0 (`not_run_step_exit_status_0_with_notice_annotation`,
`fail_step_exit_status_0_with_warning_annotation`) are replaced, not kept beside the new rows. A missing `bash` or `jq` fails the test
under linux and darwin; a windows run skips with the stated platform reason only (the consumer runs on
ubuntu, as the existing auto-merge test notes).

## 11. Known consequences the operator should see in the draft pull request

- **Until a pull request shows a state comment that this reader passes, every merge by automation is withheld.** The gate counts
  one thing as clean: the identity's single state comment with both legs completed and bound to the head, no unresolved finding,
  the security marker agreeing, and no corroborating finding (§3.2). That form was read on five real pull requests (E37), but the
  reader itself has never run on a real pull request; its first observable run is after landing (step B of §12). Human merges
  are unaffected.
- The pull request that introduces the gate cannot exercise it: `workflow_run` runs the default branch's
  `auto-merge.yml`, which does not yet carry the gate. It is merged by its owner (gate G6). The first
  pull request that can show the gate working is a later one, after landing. The G6 draft pull request's CI
  success proves nothing about the amended workflow on the default branch (spec §5, G6 acceptance condition);
  the remote conclusion of the `review` step and of the job is UNOBSERVED until then (research U11), and the
  remote state is "unverified".
- **Every pull request whose native review does not pass now shows a failed `Auto Merge` job** (revision 6, unchanged in revision 7).
  Before the amendment a withheld merge was a green job; now a refusal, an absent review, and a draft each
  end the `review` step non-zero, so the job is red on any pull request the gate does not pass. That is the
  ruling's intent (a failure must not read as a pass) and a visible cost: `Auto Merge` is not a required
  check (research E6), so the red blocks no human merge, but it appears on the pull request and on every
  draft. A job-level skip for drafts would need a second job (decision-index Q21, alternate).
- **Every pull request that adds, changes, deletes, or renames any file under `.github/workflows/`, or the
  gate script, is withheld from auto-merge** and needs a human merge.
- **A push after the review started leaves it withheld.** No automatic review on a push was observed (DS-2), so a moved head has
  no bound state comment → `stale-review`. A fresh review is the leader's `@codex review` request (DS-3), not a deliverable; whether it
  edits the comment or adds a second one is unobserved (AL-2, residual 9).
- A draft pull request is recorded as `not-run:draft` (step exit 11, job result never success) and withheld.
  A draft gets no native review by the leader's ruling (no controlled observation of a draft exists, U4); whether a review runs
  again after a push (DS-2) is unobserved (U6). Nothing here depends on either answer.
- **AL-1 (a late review is not picked up).** If a review finishes after the 15-minute window (the five measured comments finished
  1 to 11 minutes after their creation, E39: observations, not a bound), the merge stays withheld and the `review` step has failed
  (`not-run:review-absent`) until auto-merge next runs; the re-trigger paths are re-running `CI` for the head
  commit (U7) or re-running the failed Auto Merge job (U12). Fail closed. Pinned: AC-CCR-008
  `review_absent_window_closed_exit_11`, AC-CCR-021 `review_absent_window_closed_step_exit_1_job_failure`,
  AC-CCR-007 `expiry_never_merges`.
- **Only GoosLab's non-fork pull requests are ever auto-merged (revision 7, REQ-CCR-020).** Dependabot and every other bot, fork, or other
  author's pull request is never merged by automation, whatever its native review says; it needs the user's test confirmation and explicit
  approval. The earlier statement that bot-authored pull requests are withheld unless the native review reviews them is withdrawn. The
  `Auto Merge` job stays green with a notice for such a pull request (decision-index Q26).
- **The workflow requests no branch deletion (REQ-CCR-021).** The repository setting `delete_branch_on_merge` (relayed as `true`, not read by
  the lane) is a security setting and an operator decision item (G9); while it is true GitHub itself deletes the head branch after a merge.
- **On an unprotected base the shipped `checks` step withholds** (`gh pr checks --required` exits 1 there, E43, and the step reads that as a lookup
  failure), so no pull request into `release/main-batch-*` is auto-merged today whatever its review says; this SPEC leaves that step unchanged
  (decision-index Q22, Q27).
- The `@claude` responder disappears from issues and pull requests.
- Until the follow-up lands, the deployed doctrine still describes CodeRabbit as the review source (known
  debt KD-1, `spec.md` §7).

## 12. Remote observation path, the two states, and the ordered steps

This section answers "did a real review run" and defines the wording. It is a reading procedure for the
leader, not a criterion this lane can run: this lane makes no remote write and produces no remote data. Every
command below is a GET.

**States and the only permitted wordings.**

| State | Wording | Reached by | What it verifies |
|-------|---------|------------|------------------|
| L | `landed locally` (with the merged-local SHA) | this lane | every criterion passes on the merged-local tree |
| draft observation | `remote observed: gate path NOT RUN (not-run:draft)` | whoever reads a **draft** pull request's auto-merge result | the gate path only; never that a review ran |
| review-absent observation | `remote observed: review step failed, review absent` | whoever reads a **non-draft** pull request's Auto Merge run whose `review` step failed with `not-run:review-absent` | the gate path only; the absence of a review is not a verdict, never a pass |
| R | `remote verified: review ran and CI green for <SHA>` | whoever reads a **non-draft** pull request's data | a native review by the Codex identity bound to `<SHA>` was observed AND the CI of `<SHA>` is green |

A draft never reaches `remote verified`. A real review alone, or CI alone, is not `remote verified`, and
nothing is called complete before both are confirmed. The bare phrase "remote verified" is not used without
`review ran and CI green`.

**Read-only command shape (GET only; record, do not run any write).** For the pull request `<N>` and its
exact head commit `H`:

```
gh api "repos/modu-ai/moai-adk/pulls/<N>"                                  # head.sha = H; draft; changed_files; review_comments
gh api "repos/modu-ai/moai-adk/pulls/<N>/reviews?per_page=100&page=<P>"     # PR reviews: user.login, user.id, user.type, commit_id, state, submitted_at
gh api "repos/modu-ai/moai-adk/pulls/<N>/comments?per_page=100&page=<P>"    # review comments: user.*, commit_id, original_commit_id
gh api "repos/modu-ai/moai-adk/issues/<N>/comments?per_page=100&page=<P>"   # issue comments: the state comment of the identity is the verdict carrier (E37)
gh api "repos/modu-ai/moai-adk/commits/<H>/check-runs?per_page=100&page=<P>"# check runs (never a verdict; read to learn the form, DS-1)
gh api "repos/modu-ai/moai-adk/commits/<H>/status"                          # combined CI status of H, with the check runs above
```

Read further `page` values until fewer than 100 entries return. Record each artifact's `user.login`,
`user.id`, `user.type`, and `commit_id`; the gate's own read is the verdict on the review: run
`read-native-review` on the fetched lists (a lane or leader may write them to files). The worktree guard
refuses `gh api --jq` with object construction, so read the JSON or use a plain `select` filter.

**Ordered steps (leader-coordinated; the lane never performs a remote write).** Two observations must not be
confused (spec §5, G6 acceptance condition): the Codex app's behavior, readable on the G6 draft pull request
itself because it does not depend on this SPEC's workflow (step A), and the amended workflow's behavior, which
is observable only on a pull request after the change lands on the default branch, because `workflow_run` runs
the default branch's copy of `auto-merge.yml` (step B). Until step B exists the remote conclusion of the
`files` step, the `review` step, and the job is UNOBSERVED and the remote state is "unverified".

A. *The Codex app on the G6 draft pull request (DS-1, DS-2; workflow-independent).* Read the pull request
   object, `.../reviews`, `.../comments`, `.../issues/<N>/comments`, and the check runs (GETs above): does a
   review appear on a draft, under which author login, numeric id, and type, bound to which `commit_id`, and
   is its state comment the one this reader passes (DS-1: observed on other pull requests, E37)? After the ready-for-review transition (a remote
   write by the leader) and after a later push, re-read the same GETs: does a review run again, bound to the
   new head (DS-2; the settings relayed by the leader are in research E33)? Run `read-native-review` on the
   fetched lists (`--wait-closed true`) for the verdict the gate would give. Nothing here shows the workflow.
B. *The amended workflow, after landing.* The leader opens a later draft pull request (G6 coordination)
   whose base is a default branch that already carries the amendment. Read the pull request object
   (`draft: true`); the `CI` run of its head commit (`actions/workflows/ci.yml/runs?head_sha=<H>&event=pull_request`);
   the `Auto Merge` runs created after that run completed (`actions/workflows/auto-merge.yml/runs?event=workflow_run`;
   a `workflow_run` run carries the default branch's commit as its own `head_sha`, so match by time and by the
   log, which the `head` step prints as `CI tested : <sha>`); that run's job conclusion, the conclusion of each
   step, and the log line `NOT RUN: draft`. Wording: `remote observed: gate path NOT RUN (not-run:draft)`. For a
   non-draft pull request with no review at the window's end: `remote observed: review step failed, review
   absent`. Whether `CI`, and so auto-merge, fires for a draft head is unconfirmed (U8): if no run exists the
   gate path is `not observed`, a gap, not a pass. Whether re-running a failed Auto Merge job re-reads a late
   review is U12.
C. *CI.* Read the CI of `H` (`commits/<H>/status` and the check runs). Only with a bound review observed (step
   A) and CI green for the same `H` of a non-draft pull request is the wording `remote verified: review ran and
   CI green for <SHA>`.
D. *Fresh review after a later push* (DS-3): how a new review would be requested is an operator or leader
   action outside this lane; until it is done, the moved head is `stale-review` and the `review` step fails.
