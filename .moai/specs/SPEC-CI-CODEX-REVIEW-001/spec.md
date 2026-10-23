---
id: SPEC-CI-CODEX-REVIEW-001
title: "CI review switch — fail-closed native Codex review gate replaces the CodeRabbit and GitHub Claude review surfaces"
version: "0.11.1"
status: in-progress
created: 2026-10-18
updated: 2026-10-23
author: "Goos Kim"
priority: P1
phase: "v3.2.0 target"
module: ".github/workflows"
lifecycle: spec-anchored
tags: "ci, review-gate, codex, coderabbit, github-actions, fail-closed, auto-merge"
tier: L
amendment_of: SPEC-CI-CODEX-REVIEW-001
related_specs: [SPEC-V3R6-CODERABBIT-ADOPTION-001, SPEC-CI-VERDICT-INTEGRITY-001]   # optional non-blocking references: listed as an optional field by the manager-spec agent definition, absent from the schema SSOT table (plan audit D12 recorded, kept)
---

# SPEC-CI-CODEX-REVIEW-001 — CI review switch

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-10-18 | manager-spec (card t1868, lane-22) | Initial plan-phase artifacts. Class C, Tier L. Operator order 2026-10-18 recorded in the card text: replace and verify the CodeRabbit and GitHub Claude review/Actions of modu-ai/moai-adk with the official Codex review. |
| 0.2.0 | 2026-10-18 | manager-spec (card t1868, lane-22) | Plan revision 2, before the first audit. (1) Shrunk to the leader's boundary ruling (b): the doctrine, role-core, template, hook, ledger, locale, and runbook work of the former REQ-CCR-019..021 and AC-CCR-021..024 is out of this card and recorded as follow-up candidates and known debt KD-1. (2) Operator directions relayed by the leader: fork pull requests get no secret and a not-run state that is neither failure nor pass, the four function-check failures pinned, the remote observation path named, and local landing separated from remote verification. (3) A measured coupling added: the guard-liveness census test. |
| 0.3.0 | 2026-10-18 | manager-spec (card t1868, lane-22) | Plan revision 3 after plan audit iteration 1 (FAIL 0.72, receipts rcpt-819813bcf0a72adc420a871d and rcpt-762f1a1b59db60680237fbaf) and the operator-confirmed design correction relayed by the leader. (1) **The reviewer is the ChatGPT subscription's native Codex GitHub review, not a keyed `openai/codex-action` workflow**: the key, the review workflow, the verdict job, the schema, the keyless notice job, and gate G1 are removed; the gate now reads the native review's pull-request review through the GitHub API, bound to the Codex app identity and the exact head commit. (2) Audit defects closed against the new design: forged verdict carrier → identity-and-SHA binding plus a tamper guard over every workflow file (D1); unpaginated lists → exhaustive paging with count proofs (D2); consumer wiring → a structural requirement and criterion (D3); two-state wording with a draft never reaching "remote verified" (D4, D5); accepted limitations (D6); fail-closed reading of an unverified artifact shape (D7); exact Q14 labels (D8); merge-base preserve proof (D9); GEARS keywords (D10); no artifact or cache (D11). (3) Identifiers were rebuilt: 17 requirements and 18 criteria, contiguous; the iteration-1 identifiers map to the new ones in `.moai/reports/t1868/progress.md`. |
| 0.4.0 | 2026-10-18 | manager-spec (card t1868, lane-22) | Plan revision 4 after plan audit iteration 2 (FAIL 0.81, receipts rcpt-df0227b9e68727bfeefccd65 and rcpt-196d8f94f5876f4814d6e5de), the last normal round. Changes are in the review reader, not in the design direction. (1) **What counts as clean** is one rule: the newest native review bound to the head must be in state APPROVED with a submission time, no finding may exist, and a CHANGES_REQUESTED review is superseded only by a later APPROVED; COMMENTED (with or without inline comments, whatever its body says), DISMISSED and PENDING are never clean and review body text is never read (D12, D13; REQ-CCR-001, 004). (2) **NOT RUN is one named result** with two causes, a draft and a review absent at the end of the wait window; an unreadable list at expiry is a failure, not NOT RUN (D17; REQ-CCR-007, 008). (3) Changed-file entries are bound to the verified head commit and the head is read before and after the list (D14; REQ-CCR-010, 012); the merge decision path is a criterion of its own (D15; AC-CCR-019); the `gh` allow-list is exact (D16; REQ-CCR-006). (4) REQ-CCR-003, 013, 014, 015 are restated in positive form (D22). (5) AL-1 is an unverified expectation (D18); CI-input files outside the tamper guard are named residual risk (D19); a repeated list entry refuses (D20); stale research cross-references corrected (D21); the dated DS-1 observation of 2026-10-18 is recorded in the DS-1 row (D23). Counts: 17 requirements, 19 criteria, contiguous. |
| 0.5.0 | 2026-10-18 | manager-spec (card t1868, lane-22) | Plan revision 5, the post-ceiling correction under the leader's standing rule C (one correction commit, one delta audit) after plan audit iteration 3 (FAIL 0.85, receipts rcpt-44d9b0e5a5829de2c78400bd and rcpt-8b713fb407597f85c325b2fe). Three local fixes, no design change, identifier sets unchanged: (D24) REQ-CCR-003 no longer excludes a review by its body text; an artifact is excluded because of its author and the gate reads no body text of any artifact. (D25) The `review` step reads the pull request's head after its list reads on every pass, so REQ-CCR-012, design §5, and AC-CCR-012 say the same thing. (D26) The `files` term of the merge `if` is pinned structurally by AC-CCR-012, and AC-CCR-019 names a non-equivalent mutant in its place. Counts: 17 requirements, 19 criteria, contiguous. |
| 0.6.0 | 2026-10-18 | manager-spec (card t1868, lane-22) | Plan revision 6, an in-place amendment of the closed SPEC (see `## Amendments`) under the leader's ruling, board record d-20261018T152011Z-8447 (supersedes d-20261018T151824Z-db98; ruling and scope unchanged), after the delegated Codex observer's reproduction of the `review` step: a review failure ended the step with exit status 0 and only wrote `ok=false` / `state=fail` to the outputs, and a draft wrote the key `state` twice. The delta is in `design.md` "Delta of plan revision 6" (Δ1 to Δ10). (Δ1, Δ11) REQ-CCR-018, new: every path of the `files` and `review` steps other than their pass ends non-zero, so the step and the job fail; one table maps gate exit code to step exit, job result, and outputs (`design.md` §5). (Δ2) REQ-CCR-008 and the NOT RUN term split by cause at the consumer: a draft keeps the state `not-run:draft` with a job result that is not success; a non-draft pull request with no review at the end of the window is a failure. (Δ3) REQ-CCR-019, new: one value per output key. (Δ4) REQ-CCR-007 and REQ-CCR-012 aligned. (Δ6) the remote conclusion of the job is UNOBSERVED and the G6 acceptance condition says why a draft pull request cannot show it. (Δ7) DS-1 and DS-2 restated with the leader's observation items. (Δ8) KD-1 transferred to card t1874. (Δ10) two exclusions. Identifiers REQ-CCR-018, REQ-CCR-019, and AC-CCR-020 to AC-CCR-025 are new; the same-numbered identifiers named in the 0.2.0 row belong to the superseded revision-1/2 numbering. Counts: 19 requirements, 25 criteria, contiguous (ceilings 25 and 25). The leader's scope addition (board record d-20261018T155527Z-4e9e, a delegated Codex observer run, not an operator run and not a new approval) was folded in before audit: the rule covers both added steps and the chain (Δ11, AC-CCR-025). |
| 0.7.0 | 2026-10-19 | manager-spec (card t1868, lane-22) | Plan revision 7, an in-place amendment (see `## Amendments`, Amendment 2) under the leader's rulings and operator follow-up of 2026-10-19 (board and messages relayed by the leader). The delta is in `design.md` "Delta of plan revision 7" (Δ12 to Δ26). **The native Codex review is an issue-comment state protocol, not a pull-request review object** (measured on five real pull requests, `research.md` E37 to E46): one issue comment by the identity (login `chatgpt-codex-connector[bot]`, id 199175422, type Bot), created when a review starts and edited in place until it ends, carrying the marker `<!-- codex-review-state:v1 {...} -->` with a `code_review` and a `security_review` leg; the `reviews` list stays empty unless there are findings. (Δ12) REQ-CCR-001 to 005 are replaced by the comment contract: pass only when both legs are `completed`, bound to the head, with `has_unresolved_findings` exactly false and the security marker `completed` for the head (null is not false; #1842 shows why); `running` is pending; everything else, malformed or duplicated markers, a marker outside its position, and more than one state comment fail closed; an APPROVED review object no longer passes anything, and a review or review comment of the identity can only refuse. (Δ13) REQ-CCR-006, 007, 008, 010, 012 follow (listings, pending, recheck of the same comment). (Δ14) The gate applies to every merge base, an unprotected integration base such as `release/main-batch-YYYYMMDD` included. (Δ15) AC rows rewritten to the real-sample fixtures; identifiers unchanged, the criterion count stays 25. (Δ16) DS-1 observed on other pull requests, not on #1828; DS-2 open; Q20 and FU-10 closed. (Δ17) Motivating incident: PR #1841 merged before its Codex review finished. (Δ25) Auto-merge policy alignment: REQ-CCR-020 (only a pull request authored by GoosLab, user id 6458969, type User, non-fork, non-bot is ever merged by automation) and REQ-CCR-021 (no automatic branch deletion, no `--delete-branch`) are new, enforced in the workflow and pinned by rows of existing criteria. (Δ26) The re-request observation (#1842: the state comment edited in place, a receipt comment added, markers carrying different SHAs) and the findings case of #1833 are fixtures. Counts: 21 requirements, 25 criteria, contiguous (ceilings 25 and 25). |
| 0.8.0 | 2026-10-19 | manager-spec (card t1868, lane-22) | Plan revision 8, the correction commit after plan audit iteration 1 of revision 7 (FAIL 0.69, receipt rcpt-218b45ffecb9ec590f621e17, verdict file `.moai/reports/t1868/plan-audit-rev7.md`, local only). No new contract beyond the audit's required fixes: (D1) REQ-CCR-001 states once that a review object or review comment can only refuse; (D5) REQ-CCR-012 is scoped to the `files` and `review` steps and leaves `author` to REQ-CCR-020; (D4) the recheck re-lists the issue comments and compares the issue-comment count and the comment's id, `updated_at`, and markers, and fails on any difference; (D7) a `running` comment waits even when the security marker is not yet present; (D6) the #1842 move is worded as the evidence shows; AC-CCR-019 splits its job-result groups (D2), states the `gh pr checks` stub (D9), and the class lists are re-derived from the cells that reproduce on this tree (D3, D11). The delta is in `design.md` "Correction of revision 8". Counts: 21 requirements, 25 criteria, contiguous. |
| 0.9.0 | 2026-10-19 | manager-spec (card t1868, lane-22) | Plan revision 9, the correction after plan audit iteration 2 of revision 7 (revision 8 audited: FAIL 0.88, receipt rcpt-cba480b75217f40e0e5cf040, verdict file `.moai/reports/t1868/plan-audit-rev8.md`, local only) and one new real sample. (N1) AC-CCR-022 is reclassified as a regression-guard (its draft behaviour is already implemented and tested; the renamed row is a test addition, not a flip): 12 release-blocking, 13 regression-guard. (N2) AC-CCR-019 group E and design residual 7 state the exact timing of the recheck: an in-place edit after the recheck's last list read is not detectable. (New input) PR #1844, comment 6260734432, observed 2026-10-19T06:14:10Z: a leg whose `status` is `failed` is a review run that failed, neither a pass nor a finding; REQ-CCR-004 and REQ-CCR-005 name it, the reader refuses it with `review-failed:<leg>` before any finding flag is interpreted (design §3.2 step 7a, T-CHAIN row 2i), and the fixture `F-failed` and the criteria rows pin it. No new requirement and no new criterion. Counts: 21 requirements, 25 criteria, contiguous. |
| 0.10.0 | 2026-10-20 | t1868 run-phase writer | Correction from the exact-head #1828 code review finding 4315924745 on `18d2a6d176d4a7ba544e2f58760ff5ba3c0f4778`: a current-head security marker still reading `running` after both state legs read `completed` is pending, as already stated by design §3.2 step 10 and AC-CCR-007, not a marker disagreement. REQ-CCR-004 and REQ-CCR-007 now state this transition explicitly. The follow-up found that a malformed SHA on a completed leg could reach pending first; the gate now refuses that shape before pending without weakening stale, failed, or unresolved checks. No requirement or criterion identifiers added; counts remain 21 requirements and 25 criteria. |
| 0.10.1 | 2026-10-20 | t1868 run-phase writer | Correction from exact-head #1828 finding 4316243010 and the upper contract in `AGENTS.md` §4: completed state legs must report both `has_any_findings: false` and `has_unresolved_findings: false`. The gate previously accepted `has_any_findings: true` when unresolved was false. REQ-CCR-001, 004, and 005, the reader design, and AC-CCR-001/005 now agree; three per-leg/both-leg controls refuse and a completed finding is not hidden by a running peer. Counts remain 21 requirements and 25 criteria. |
| 0.10.2 | 2026-10-20 | t1868 run-phase writer | Correction from exact-head #1828 security finding 4316515155: a pre-merge recheck no longer rejects an otherwise clean review solely because ordinary issue comments were added after the earlier clean pass. It snapshots the fresh PR issue-comment count, proves the complete paginated list against that count, checks the same count and head after collection, and reruns the trusted native-review gate; review-comment count, trusted identity, marker/state integrity, and clean summary equality remain enforced. Counts remain 21 requirements and 25 criteria. |
| 0.11.0 | 2026-10-23 | manager-spec (card t1868) | Plan revision 13, an in-place amendment (Amendment 6) under the operator and leader ruling of 2026-10-23 (decision-index Q30): audit findings F4 (a Codex review that completes after the merge step's window, or after the pull request becomes ready, is never re-evaluated) and F3/F1 (merge-step holds end exit 0 and look green) are fixes, not accepted limitations. New: REQ-CCR-022 (an `issue_comment` trigger, `created` and `edited`, admitted at the job level only for a pull request comment of user id 199175422; no pull-request code is checked out or run), REQ-CCR-023 (the same chain, one evaluation with no wait, head re-confirmed, one concurrency group per pull request), REQ-CCR-024 (a merge step that was reached and did not merge ends non-zero), REQ-CCR-025 (a commit status `auto-merge` on the evaluated head, `statuses: write`, no required check added). Reworded: REQ-CCR-007, 011, 013; §2 item 7; §5 stated fact and G6 condition; §6 Out of Scope "step conclusions"; §7 AL-1 superseded (resolved by REQ-CCR-022). The F1 `active` gating, the env-passed `PR_NUMBER`, the per-outcome job-summary line, and the shellcheck step of commit `41b7f29b7a` are already landed and are not re-specified as new work. Criteria: five lettered sub-criteria (AC-CCR-006a, 013a, 013b, 013c, 019a) inside existing logical criteria, because the logical criterion count is at the Tier L ceiling of 25. Counts: 25 requirements (ceiling 25), 25 logical criteria with 5 sub-criteria. No plan-audit re-entry (ruling); one head-bound sync re-audit at the final head checks this SPEC against the code. |
| 0.11.1 | 2026-10-23 | manager-spec (card t1868) | Revision 13.1, the lane ruling on the revision-13 defaults: decision-index Q31(a) is REVERSED — on the comment path a pending review pass (a `running` leg or marker, or no state comment yet) is a "not yet" hold with exit 0 and a `pending` `auto-merge` status (`review-not-complete:running`), never a failed job; only a merge step reached and not executed is non-zero. Q31(b) stands, clarified in design. The concurrency key is specified: the pull request number for the `workflow_run` path and the identity's comments, the run id for any other comment author. Reworded REQ-CCR-012, 018, 019, 023, 024; no requirement or criterion added. |

## Amendments

### Amendment 6 — late reviews are re-evaluated and held merges are visible (plan revision 13)

- rationale: the first sync audit (`.moai/reports/t1868/sync-audit-1.md`, PASS-WITH-DEBT 0.91, local only and gitignored) recorded F4, a review that completes after
  the merge step's window or after the pull request becomes ready is never re-evaluated, because the only trigger is a completed `CI`
  run; and F3/F1, a merge that was reached and not executed ends exit 0, so a hold reads as a green run. The operator and leader ruled
  on 2026-10-23 that both are fixes, withdrew the accepted-limitation option (AL-1), and approved the design of `design.md` "Delta of
  plan revision 13": an `issue_comment` trigger gated at the job level to the Codex app identity's id, a single evaluation with no wait,
  non-zero held outcomes of the merge step, and a commit status `auto-merge` with the one new permission `statuses: write`.
- scope: REQ-CCR-022 to 025 added; REQ-CCR-007, 011, 013 reworded; §2 item 7 added; §5, §6, §7 restated; AC-CCR-006, 007, 012, 013, 019
  amended and the sub-criteria AC-CCR-006a, 013a, 013b, 013c, 019a added; `plan.md` M7 added; `design.md` delta and §5, §6, §11 restated;
  decision-index Q22 superseded in part, Q30 and Q31 added. Not changed: the gate script's recognition contract and exit classes, `T-CHAIN`
  for the `files` and `review` steps, the author policy, `.github/required-checks.yml`, branch protection.
- evidence: RED-now ledger R28 to R35 of `acceptance.md` (tree `41b7f29b7a`): the workflow names no `issue_comment` trigger, no comment
  author gate, no comment-path pull request number, no event discrimination, no commit status and no `statuses: write`, and each held
  merge outcome ends `exit 0` or falls through to exit 0.
- plan audit: no re-entry, by the ruling; the sync re-audit bound to the final head is the check of this amendment against the code.

### Amendment 5 — unrelated issue comments do not block the pre-merge recheck (plan revision 12)

- rationale: exact-head security finding #4316515155 on `8989b7ebe83ffb1d0a1d922b3ee61d205c126783` showed that the recheck compared the current total issue-comment count with the count captured during the earlier clean pass. An unrelated external comment therefore withheld a merge before the trusted Codex artifacts were re-evaluated.
- scope: REQ-CCR-012, AC-CCR-012, design §3.2 and T-CHAIN row 7, decision-index Q25, and the M6 verification notes now use the fresh recheck PR count as the completeness anchor. The full paginated issue-comment list must match that count and the second PR read must still match it; the verified head and review-comment count remain fixed, and `read-native-review` must reproduce the same clean trusted summary. A count increase since the earlier clean pass alone is not a refusal. No identifiers or counts change.
- evidence: the exact 8989 recheck reproduction with issue comments 3→4 failed before the change with `RECHECK_REFUSED: issue-comments-moved-before-merge`; `TestAutoMergeReviewRecheckBeforeMerge/unrelated_issue_comment_added_before_recheck_does_not_block` passes after the change, `issue_comment_count_moved_during_recheck_not_ok` still refuses movement between the fresh count read and the second PR read, and missing/non-numeric R3 counts fail closed.

### Amendment 4 — completed legs cannot report findings (plan revision 11)

- rationale: exact-head code review finding #4316243010 on 9a22ef21ba948e40125034094994b281575230c7 identified that a completed state leg with `has_any_findings: true` and `has_unresolved_findings: false` could pass beside clean receipts. The committed upper contract in `AGENTS.md` §4 requires completed code review to report no findings or unresolved findings.
- scope: REQ-CCR-001, REQ-CCR-004, and REQ-CCR-005 plus design and AC-CCR-001/005 now require both completed-leg flags to be well-formed booleans and false for a clean result. A completed leg reporting findings refuses before a pending peer; failed-leg flags retain their existing precedence, and running legs remain pending unless another refusal applies. No identifiers or counts change.
- evidence: the preserved exact-9a22 three-case reproduction returned exit 0 before the fix; RED failed the code-only, security-only, and both-leg assertions. After the gate change, `TestCodexNative1828ReceiptTypesAndHeadBinding` refuses all three with `finding:reported:<leg>`, and `TestCodexNativeRunningSecurityLegWaits/completed_code_finding_refuses_before_running_security_wait` refuses the completed finding before pending.

### Amendment 3 — a current running security marker remains pending (plan revision 10)

- rationale: exact-head Codex code review finding 4315924745 on `18d2a6d176d4a7ba544e2f58760ff5ba3c0f4778` reproduced an exit-1 refusal when both state legs were completed and the current-head security marker alone still read `running`. Design §3.2 step 10 and AC-CCR-007 already say a current security marker in `running` waits; REQ-CCR-004 and REQ-CCR-007 did not state that case consistently.
- scope: clarify REQ-CCR-004 and REQ-CCR-007; no identifier or acceptance-count change. The gate returns exit 10 while the wait window is open and exit 1 `review-not-complete:running` on the last pass. Stale heads, malformed/failed states, unresolved findings, and missing completed-leg receipts remain refusals.
- evidence: the exact clean-state mutation failed before the fix with `security-marker-disagrees:running`; the new regression case is `TestCodexGateReadNativeReviewPending/security_marker_running_legs_completed_pending`. The supplied malformed-SHA controls for `null`, 10-character, and 40-character non-hex values initially returned `pending:running`; `TestCodexGateReadNativeReviewPendingRefusesMalformedCompletedSHA` now requires each to refuse with `review-not-complete:code_review:commit_sha`.

### Amendment 2 — the native review is read from its issue comment (plan revision 7)

- prior_completed_version: 0.6.0
- prior_completed_sha: pending-backfill   # a commit cannot cite its own hash (spec-workflow Sync (close)); backfilled with the real SHA in a following commit
  (the value `sync_commit_sha` still holds in `progress.md` §E.4; the revision-6 amendment did not re-close the SPEC, so the prior close is the same)
- rationale: the closed SPEC, and revision 6 after it, recognised the native review as a pull-request review
  object in state APPROVED. Measured on real pull requests (#1835, #1836, #1837, #1838, #1841; `research.md` E37 to E46) that object does not exist for a clean review: the review is one issue comment by the same bot identity,
  edited in place until it ends, and the `reviews` list stays empty. With the shipped reader every real non-draft
  pull request would end NOT RUN and every automated merge would be withheld for good. The motivating incident is
  PR #1841, merged while both Codex reviews were still running (`research.md` E41). The leader's ruling of
  2026-10-19 redesigns the contract around the comment protocol.
- scope: REQ-CCR-001 to 005 replaced; REQ-CCR-006, 007, 008, 010, 011, 012 reworded; REQ-CCR-020 and REQ-CCR-021 added;
  REQ-CCR-018, 019 and the remaining requirements untouched in meaning; the criteria AC-CCR-001 to 008, 010, 012, 019 to 023 rewritten
  to the new fixtures with unchanged identifiers (25 criteria, no new one); the §3 terms, §5 stated fact, DS-1,
  DS-2, DS-3, and §6 review-feature exclusions restated; Q20 and FU-10 closed, Q23 and Q24 added. The gate
  script, its tests, and `auto-merge.yml` as landed still implement the revision-6 reader; the code and test
  changes this amendment calls for are a run-phase milestone (`plan.md` M6; corrected by revision 8, see HISTORY), and so are the author and fork rule and the removal of `--delete-branch`, which the workflow must enforce and the
  tests must pin (a documented limit does not complete them). Not changed: the exit classes
  0, 1, 10, 11 and their mapping (T-CHAIN), the files step, the tamper guard, one value per output key,
  KD-1 (card t1874), the deferred gates.

### Amendment 1 — a review failure fails the `review` step and the job (plan revision 6)

- prior_completed_version: 0.5.0
- prior_completed_sha: pending-backfill   # a commit cannot cite its own hash (spec-workflow Sync (close)); backfilled with the real SHA in a following commit
  (the value `sync_commit_sha` holds in `progress.md` §E.4 at this revision; the commit that closed the SPEC
  and carried `status: completed` is `67a5149ce2`; the real SHA is owed by the backfill commit of that sync)
- rationale: the closed SPEC specified, in the "Step conclusion convention" of `design.md` §5, that every guard
  step ends with exit status 0 and that NOT RUN ends the step green. The delegated Codex observer's reproduction of the real
  `review` step body showed the consequence: a changes-requested review and a draft both ended the step with
  exit status 0 (a green job), a draft recorded `state` twice, and no criterion asserted an exit status or a job
  result. The leader's ruling (board record d-20261018T152011Z-8447 (supersedes d-20261018T151824Z-db98; ruling and scope unchanged)) makes a review failure a failure of the
  step and the job.
- scope: REQ-CCR-007, REQ-CCR-008, REQ-CCR-012 reworded; REQ-CCR-018 and REQ-CCR-019 added; AC-CCR-007, 008,
  012, 019 changed; AC-CCR-020 to 025 added; the G6 acceptance condition, DS-1, DS-2, and KD-1 restated; two
  exclusions added. The gate script, its exit classes, the recognition contract, and every other requirement
  are untouched. The code and test changes this amendment calls for are a run-phase milestone (`plan.md` M5);
  the implementation as landed at `0955fb9058` still has the defect.

Lineage note. This SPEC changes two choices that a completed SPEC recorded:
SPEC-V3R6-CODERABBIT-ADOPTION-001 §B/§C made CodeRabbit the single PR reviewer and retained the
`@claude` responder. The operator order of 2026-10-18 reverses both. Marking that SPEC partially
superseded is a sync-phase act and is not part of this plan's file set. This SPEC is distinct from
SPEC-CODEX-REVIEW-ASYNC-001, SPEC-CODEX-REVIEW-OWNERSHIP-001 and SPEC-CODEX-REVIEW-TARGET-001,
which concern the local `codex_review` tool and not repository CI.

## 1. Background and motivation

Facts below were read or measured in this tree (HEAD `8cd1d0d51d`, which differs from the card base
`4f9dba14bf` only under this SPEC directory and the card record; the revision-3 facts E22-E30 were
measured at HEAD `ae494aedb2`, and the revision-4 facts E31-E32 and the RED-now ledger at HEAD `e11af67632`;
both differ from the base only under this SPEC directory and `.moai/reports/t1868/`) or by read-only GET on
2026-10-18; every one carries its evidence id from `research.md` §3. The revision-6 facts E33 to E35 were
measured, or relayed and labelled as relayed, at HEAD `0955fb9058` (the run and sync phases landed on the card
branch; no merged-local SHA exists). The revision-7 facts E37 to E46 were read from the lane's saved raw REST
objects of 2026-10-19 (`.moai/reports/t1868/codex-samples/`, machine-local and gitignored, read-only; the bodies are public GitHub data) and
from the observer's report of the same day, with the publication tree at HEAD `b2af3db9a1`.

- `auto-merge.yml` withholds a merge unless a CodeRabbit commit status reaches `success` and a
  `Merge Risk:` comment line names the current head commit (E3). CodeRabbit skipped the large batch
  PR 1826 (`Too many files / Review skipped`, UNVERIFIED premise U1), so the only review gate the
  automation has can no longer reach a verdict on the PR shape this repository merges.
- `claude.yml` is the `@claude` mention responder (E4). `review-quality-gate.yml` parses a check run
  named `Claude Code Review` (E4). The workflow that produced that check run, `claude-code-review.yml`
  (remote workflow id 204797886), exists on `origin/main` and not in the local tree (E5); the
  remote and local histories diverged 10 versus 400 commits (E5).
- Main's measured required status checks are five and none is a review check (E6). Branch protection
  therefore never enforced any review; `auto-merge.yml` is the only enforcement point.
- **The reviewer.** The operator confirmed, in a logged-in browser session, the state of the existing
  ChatGPT subscription's native Codex GitHub review (relayed by the leader; this lane did not observe the
  settings page): for modu-ai/moai-adk and modu-ai/mo.ai.kr automatic review is ON for all pull requests,
  it runs **when a pull request opens**, and the personal-review credit is OFF. Per-push widening is
  excluded and settings changes are not retried by this card. The lane measured by GET that the app
  `chatgpt-codex-connector` (owner `openai`, id 1144995) exists, that its bot account is
  `chatgpt-codex-connector[bot]` with user id 199175422 and type Bot (E27), and that it is installed on
  the `modu-ai` organisation for all repositories and not suspended (E28).
- **The artifact (measured, revision 7).** Revisions 3 to 6 designed the reader before any native review had
  been seen (the 120 pull requests of 2026-10-18 carried none, E29, and a pull request 47 and 82 minutes old
  still had none, E32), and recognised a pull-request review object in state APPROVED. On 2026-10-19 the lane
  read the raw REST objects of five real pull requests (#1835, #1836, #1837, #1838, #1841; E37): the native
  review is **one issue comment** by the identity (login `chatgpt-codex-connector[bot]`, user id 199175422,
  type Bot), created when a review starts and edited in place until it ends; it carries the machine markers
  `<!-- codex-pull-request-review-summary -->`, `<!-- codex-review-state:v1 {...} -->` (a `code_review` and a
  `security_review` leg, each with `commit_sha`, `has_any_findings`, `has_unresolved_findings`, `status`), and
  `<!-- codex-security-review:v1 {...} -->`. The `reviews` list is empty on the four clean pull requests; a
  review object exists only beside findings (#1836: one COMMENTED review and one inline review comment of the
  identity). The values `running`, `completed`, and `failed` of `status` were read (E38, E39, E46); what other values exist is unknown. The gate therefore recognises exactly one carrier for a verdict, the state comment, passes on one
  condition (both legs `completed`, bound to the head, no unresolved finding), and fails closed on every
  other state (§3.1). One further sample shows why `completed` alone is not a pass: on #1842 the
  security leg reads `completed` with a null `commit_sha` and null finding fields while the security marker
  reads `running` (E42). Not observed: a pull request opened non-draft (all triggers read "Draft marked
  ready"), the whole `running` to `completed` transition (E39), a review after a push, and the behavior of a
  `@codex review` re-request (DS-2, DS-3).
- **Motivating incident (revision 7).** PR #1841 was merged at 2026-10-19T05:13:31Z; its Codex code review
  completed at 05:14:11Z and its security review at 05:15:04Z (the reviews started at 05:13:37Z and
  05:13:38Z after the draft was marked ready). The merge preceded the review, and it is not retroactively read
  as a pass. The automated merge must therefore never happen while a native review is `running` or absent for
  the head, on any base (`research.md` E41).
- **The author policy (relayed by the leader, 2026-10-19).** Only a pull request authored by GoosLab (user id 6458969, type User), from a
  branch of the base repository, and not by a bot, is ever merged by automation, and only conditionally: current-head CI, both review legs
  clean and bound to the head, no unresolved finding, the base's required checks satisfied. Every other pull request (a fork, a bot such as
  dependabot, id 49699333, type Bot, or another author) needs the user's test confirmation and explicit approval and is never merged
  by automation. Automation never deletes a branch. The repository setting `delete_branch_on_merge` is a security setting that stays an
  operator decision (G9). The shipped workflow enforces none of the author, fork, or bot rule and passes `--delete-branch`; REQ-CCR-020
  and REQ-CCR-021 make the workflow enforce them and the tests pin them (`research.md` E43, E45).
- **The gate has a base-branch scope, and it is every base.** All five measured pull requests have the base
  `release/main-batch-20261019`, an integration base without branch protection (`protected: false`, no rules; E40, E43); the review
  gate is the only enforcement point on it. A state recorded for an old head never counts for a new one: the #1842 head was observed to have moved by the coordinator's
  05:32:18Z re-check; its state comment, read at 05:24Z, named the old head, and read again at 05:36Z, after a re-request, its code leg named the
  new head while the security marker still named the old one (E43, E44). No base is exempt (REQ-CCR-006).
- **The trigger and the settings (relayed, revision 6).** The leader relayed a read-only observation of the
  native review's settings for this repository: the review trigger is "when a pull request is opened"; automatic
  review is on for all pull requests; thorough review is on; automatic security review is on; the minimum
  reported severity is High (E33; this lane did not observe the settings page, and changes no setting). What a
  review of a pull request with no finding at or above High leaves behind is unknown, and so is whether a review
  runs again after the ready-for-review transition or after a push (DS-1, DS-2, §5).
- **The consumer swallowed review failures (measured, revision 6).** The delegated Codex observer's reproduction of the real
  `review` step body against a stub `gh` and the real gate script showed that a changes-requested review and a
  draft both end the step with exit status 0, that a draft records the key `state` twice (`ok=false`,
  `state=fail`, `state=not-run`), and that a clean pass records every key twice; this lane re-measured the same
  three outcomes at HEAD `0955fb9058` (E34); a delegated observer found the same for the `files` step with
  the gate script absent (exit 0, `ok=false state=fail`, E36). The failure was visible only as a value in the step's outputs: the
  job concluded green, and no criterion asserted an exit status or a job result (E35). This revision closes that
  (REQ-CCR-018, REQ-CCR-019).
- The operator order is a replacement, not an addition. It also forbids new keys, cost, persistent app
  permission, branch-protection changes, and any loosening or widening of `required-checks.yml` before user
  confirmation. This SPEC adds no key and no secret and needs none; every user-held step stays a named
  confirmation gate (§5).

**Boundary of this card (leader ruling (b), 2026-10-18; replacement scope fixed again by the design
correction).** The replacement scope is ONLY GitHub CodeRabbit and the GitHub Claude review/Actions;
local Claude and GLM sessions and the functional CI are untouched. The audited scope is: the consumer
wiring in `auto-merge.yml` and the retirement of `claude.yml` and `review-quality-gate.yml` under
`.github/workflows`; the review-gate script and its tests; `.coderabbit.yaml`; a guard that
`.github/required-checks.yml` stays byte-unchanged and carries no Codex review check (this SPEC never edits
that file); and the `.moai/guards/liveness.yaml` entries that point at the deleted workflows, with the tests
that read that manifest. This SPEC adds no workflow. Doctrine, role-core rules, templates, `AGENTS.md`, the
hook lane notice, the binding ledger, and every other `Merge Risk` wording are not part of this card
(§6, §7).

## 2. Goal

After the run phase lands locally and the leader's draft pull request is merged by its owner:

1. For a non-draft pull request authored by the one eligible author (REQ-CCR-020), on every merge base (an unprotected integration base such as
   `release/main-batch-YYYYMMDD` included), `auto-merge.yml` merges only when the pull request carries a
   finished native Codex review — exactly one state comment authored by the Codex app identity whose marker
   reports both legs `completed`, bound to the exact head commit, with no unresolved finding, and no
   corroborating finding of the identity on that commit — in addition to the unchanged CI observations. A
   review that is `running` or absent for the head is never merged over. No comment, label, or review of any
   other author, no prose outside the marker, no review object of the identity, no third-party status is a
   verdict, and no elapsed time is a pass.
2. No CodeRabbit, Claude review, or `@claude` responder surface remains in the repository's workflows or
   under `scripts/ci/`, except the historical and fixture text named in §6. Doctrine surfaces are not changed
   by this card and are carried as known debt KD-1 (§7).
3. Every state the gate cannot verify — no state comment, a review of another commit, a leg that is not
   `completed`, an unresolved finding, a malformed, duplicated, or misplaced marker, more than one state
   comment, an unreadable or unprovably complete list — leaves the merge withheld with a named reason (a
   review still `running` is waited on inside the window); nothing is ever read as a pass by absence or by
   the passing of time.
4. A draft pull request is reported `not-run:draft` — a state of its own that is never a pass — and the
   `review` step and the job do not end in success; a non-draft pull request with no native review observed
   for its head commit when the wait window ends is a failure of the step and the job (cause
   `not-run:review-absent`). Auto-merge withholds the merge on both; a draft report verifies the gate path
   only. The leader's ruling is that a draft pull request gets no native review (a review starts on
   open-for-review, on draft-marked-ready, or on an `@codex review` comment: the trigger list the bot's own
   comment states, E37, together with the silence of #1828; no draft was observed being reviewed or not
   reviewed under a controlled read), and whether a review runs again after a push is an item to confirm by
   observation (§5, DS-1, DS-2).
5. A failure in either guard step this SPEC adds is a failure of that step and of the job: every path of the
   `files` and `review` steps other than their pass — a missing gate script included — ends with a non-zero
   exit status, so the later steps and the merge do not run and no failed gate reads as a green job
   (REQ-CCR-018). The remote conclusion of that job is UNOBSERVED until the leader-coordinated flow after
   landing (§4, §5 G6).
6. Automation merges only a pull request authored by GoosLab (user id 6458969, type User) from a branch of the base repository; a fork,
   bot, or other author's pull request needs the user's test confirmation and explicit approval and is never merged by automation, and
   automation never deletes a branch (REQ-CCR-020, REQ-CCR-021). The workflow enforces both rules and the tests pin them.
7. (revision 13) A native review that finishes after a `CI`-triggered evaluation ended — after its window, or after the pull request
   became ready — is evaluated again when the Codex app identity creates or edits its comment, through the same gates and without any
   pull-request code running (REQ-CCR-022, REQ-CCR-023); a merge that the automation reached and did not execute ends as a failed job
   (REQ-CCR-024); and every evaluation leaves a commit status `auto-merge` on the evaluated head naming its outcome, so a hold is never
   visible only as a green run (REQ-CCR-025).

## 3. Requirements (GEARS)

Notation: GEARS. Behavior is stated; file, script, and function names live in `design.md`. Identifiers are stable and are never renumbered:
REQ-CCR-018 to 025 stand in §3.2 beside the consumer requirements they concern, ahead of REQ-CCR-014 to 017 of §3.3 and §3.4, so the
document order follows the sections, not the numbers.
Terms: the **gate script** is the repository's single review-verdict checker; the **Codex app identity** is
the account with login `chatgpt-codex-connector[bot]`, numeric user id 199175422, and account type Bot (an
artifact is by the identity only when all three match); the **native review** is what that identity leaves on
the pull request, and its verdict carrier is the **state comment**: an issue comment authored by the identity
  whose body carries the **state marker**, the single line `<!-- codex-review-state:v1 {JSON} -->` standing in
  the leading block of the body (the run of lines from the first line on that each begin with `<!--` and end
  with `-->`); the JSON has two **legs**, `code_review` and `security_review`, each read for its `status`,
  `commit_sha`, `has_any_findings`, and `has_unresolved_findings` and for nothing else; the **security marker** is the line
`<!-- codex-security-review:v1 {JSON} -->` in the same leading block, read for its `status` and `headSha`
only; a leg is **bound** to a commit when its `commit_sha` equals that commit (a null `commit_sha` is bound
to nothing, and a null `has_any_findings` or `has_unresolved_findings` is not `false`); a **corroborating finding** is a review comment of the identity bound to the
head commit, or a review object of the identity bound to the head commit in state COMMENTED or
CHANGES_REQUESTED, and it can only refuse, never pass; the **eligible author** is the account with login `GoosLab`, numeric user id 6458969, and account type User (all three must
match), and an **eligible pull request** is one it authored whose head repository is the base repository (compared by repository id; a
missing head repository is a fork); the **gate paths** are every file under the workflows
directory and the gate script; **clean** is the one state REQ-CCR-001 defines; **pending** is a state
comment with a leg `running`; **NOT RUN** is the gate script's exit class 11, with exactly two causes: a
draft pull request, or no state comment observed for the head commit when the wait window ends. At the
consumer the two causes are not one result: a draft is the recorded state `not-run:draft` with a job result
that is never success, and a non-draft pull request whose window ended with no review is a failure
(REQ-CCR-008, REQ-CCR-018). The **step exit status** and the **job result** are what the workflow platform
reports for the `review` step and its job; **a pass** is the one case where the step exits 0.

### 3.1 Native review verification

- **REQ-CCR-001** (Ubiquitous) The gate script shall report a pull request's native review as clean only when
  exactly one state comment of the Codex app identity exists on the pull request, its state marker parses as
  one JSON object, and for each of the two legs the `status` is `completed`, the `commit_sha` equals the
  exact head commit under review, and both `has_any_findings` and `has_unresolved_findings` are the boolean
  `false` (a `completed` status alone, or both legs reading `completed`, is never enough, and null is not
  `false`), the
  security marker stands once, reads `completed`, and carries the head commit in `headSha`, and no
  corroborating finding exists on that commit; the gate script shall take this decision from the author of
  each artifact, the marker's three fields per leg, the security marker's two fields, and the
  corroborating artifacts alone, so that prose, table rows, the summary marker, and reactions, in either
  direction, change no pass, and so that no review object or review comment of any state can grant a pass: such an artifact can only add a
  refusal (REQ-CCR-005).
- **REQ-CCR-002** (Event-driven) When either leg of the state comment is bound to a commit other than the head
  commit under review — including a head that moved after the review started — the gate script shall refuse
  and name the stale commit, never read it as pending and never as NOT RUN.
- **REQ-CCR-003** (Event-driven) When a comment, review, or review comment is authored by any login, any
  numeric id, or an account type other than the Codex app identity's, the gate script shall exclude it from
  the verdict because of its author, in every state and whatever its body text says — a marker string inside
  it included — so that the verdict derives only from artifacts of the Codex app identity; the gate script
  shall read the body of one artifact kind only, the issue comments of the identity, and of such a body only
  the state marker (REQ-CCR-001).
- **REQ-CCR-004** (Event-driven) When the state comment is not clean — a leg `status` other than `completed`
  (`running` is pending, REQ-CCR-007; `failed` is a review run that failed, which is neither a pass nor a finding: the gate script refuses it with the reason `review-failed:<leg>`, as a gate refusal that fails the step and the job, before it interprets any finding flag of that leg, so the flags of a failed leg are never read as findings, and the reason names the re-request that starts a new run; any other value, a missing or null `status`, a missing leg, or a completed leg with a missing,
  null, or non-boolean `has_any_findings` or `has_unresolved_findings`, or a `commit_sha` that is not 40 hexadecimal characters
  refuses), a marker whose JSON is malformed or altered so that it no longer parses to the required shape, a
  marker string that stands outside the marker position of a comment of the identity, a body that carries the
  state marker or the security marker twice, a malformed or misplaced security marker, a security marker that is missing
  while no leg is `running` (a leg still `running` waits, REQ-CCR-007), a current-head security marker whose status is
  `running` (pending even if both legs already read `completed`, REQ-CCR-007), a security marker whose status is neither
  `completed` nor `running`, or a security marker whose `headSha` is not the head (the two markers disagree), or more
  than one state comment of the identity on the pull request — the gate script shall
  refuse and name the reason; no later comment, no prose, no review object, and no elapsed time clears it,
  and the gate script shall not choose among duplicates by recency.
- **REQ-CCR-005** (Event-driven) When a completed leg reports `has_any_findings` true, or a leg whose run did not fail
  reports `has_unresolved_findings` true, or a corroborating finding exists on the head commit, the gate script
  shall refuse and name it; when none exists, the findings check shall pass. A completed reported finding
  refuses even when `has_unresolved_findings` is false and before a pending peer leg can defer the decision.
  A review or review comment of the identity can only add a refusal and never grants a pass.

### 3.2 Auto-merge consumer

- **REQ-CCR-006** (Ubiquitous) Auto-merge shall authorize a merge only when the gate script's read of the
  pull request's native review exits 0 for the exact head commit, in addition to the unchanged CI
  observations, on every merge base of a pull request — a base without branch protection, such as an
  integration branch `release/main-batch-YYYYMMDD`, included, so that no base is exempt and a review that is
  `running` or absent for the head is never merged over on any base; it shall take the review verdict from
  that exit status alone, and shall read, for that verdict, only the pull request object and its
  changed-file, issue-comment, review, and review-comment listings; the `gh` calls the workflow and the gate
  script make shall be exactly the allow-list of `acceptance.md` AC-CCR-006.
- **REQ-CCR-007** (Event-driven) When no state comment is recognised yet, a state leg reports `running`, or the
  current-head security marker reports `running` (including when both state legs already read `completed`), auto-merge
  shall keep waiting inside a bounded window (at most 60 passes of 15 seconds; time
  appears only as this bound, and no elapsed time is ever a pass); when the read refuses, auto-merge shall
  withhold the merge at once and end the `review` step as a failure (REQ-CCR-018); when a leg or the current-head
  security marker is still `running` on the last pass, auto-merge shall report a failure named `review-not-complete:running`, a failed
  step and a failed job, never NOT RUN and never a pass; when a list the verdict needs stays unreadable to
  the end of the window, auto-merge shall withhold the merge and report a failure named
  `review-unreadable`, a failed step and a failed job, never NOT RUN. A review that finishes after the window
  expired is read by the comment-path evaluation that the Codex app identity's comment creation or edit starts
  (REQ-CCR-022, REQ-CCR-023); the bounded window above applies to the `workflow_run` path only (revision 13; the
  former "read only when auto-merge next runs", AL-1, is superseded).
- **REQ-CCR-008** (Event-driven) When the pull request is a draft, auto-merge shall record the single state
  `not-run:draft` — distinct in gate exit class (11), annotation, and recorded state from pass and from
  failure — shall withhold the merge, and shall end the `review` step with a non-zero exit status so that the
  job result is never success, whether or not any state comment exists on a draft; when the pull request is not
  a draft and the wait window ends with no state comment observed (gate exit class 11, cause
  `not-run:review-absent`), auto-merge shall report a failure — a failed step and a failed job — naming that
  cause, and shall withhold the merge; a draft report verifies the gate path only and is not a review result.
- **REQ-CCR-009** (Event-driven) When a pull request adds, changes, deletes, or renames any gate path,
  auto-merge shall withhold the merge for human decision.
- **REQ-CCR-010** (Event-driven) When a list the gate script decides on — changed files, issue comments,
  reviews, or review comments — cannot be proven complete and bound (a page cannot be read, the collected
  count differs from the
  pull request's own count where one exists, an entry repeats, the API's page ceiling is reached, the
  changed-file list is empty, or a changed-file entry that is not a removal is not bound to the verified head
  commit), the gate script shall refuse and, where the refusal comes from collecting a list, remove its output
  list.
- **REQ-CCR-011** (Ubiquitous) Auto-merge shall check out the default branch alone on both trigger paths (a pull-request ref is
  never checked out) and shall keep unchanged its head-commit re-verification, required-check observation,
  overall deadline, and head-pinned merge call, which requests no branch deletion (REQ-CCR-021).
- **REQ-CCR-012** (Ubiquitous) The auto-merge workflow shall run, in this order and before its merge step, the `author` step
  (REQ-CCR-020), a step that collects and checks the pull request's changed-file list (`files`), and a step that reads the native
  review (`review`); the `files` and `review` steps shall each read the pull request's head commit before and after their list reads
  (the review step on every pass that reads lists) and record a failing result when that head differs from the verified head commit, shall
  each record a passing result only when the corresponding gate-script call exits 0, and shall each treat any other exit, including one
  they do not recognise, as a withheld merge and a failed step (REQ-CCR-018; the one exception is the comment path's pending `review`
  pass, a "not yet" hold with exit 0, REQ-CCR-023); the `author` step is governed by REQ-CCR-020 alone (one
  read of the pull request object, no gate-script call, exit status 0); the `review` step shall recheck a clean pass before recording it,
  by taking the issue-comment count from the first pull-request read of that recheck, re-listing the complete issue-comment list, the reviews,
  and the review comments, then requiring the same head, that fresh issue-comment count, the unchanged review-comment count, and the same clean
  verdict of the same state comment (its id, its `updated_at`, and the compact JSON of both markers) from the second pull-request read and a
  repeated gate evaluation; shall end as a failed step on an unreadable or changing list, a changed head or review-comment count, or a changed
  trusted clean verdict. It shall not fail solely because the fresh issue-comment count differs from the earlier clean pass, because unrelated
  issue comments do not grant or refuse a Codex verdict; the complete list and trusted artifacts are still re-read before merge. The merge step
  shall run only when the `author`
  step recorded `eligible` true and both guard steps recorded a passing result, shall read the head commit again immediately before the
  merge call, and shall withhold the merge when that head differs from the verified head.
- **REQ-CCR-013** (Ubiquitous) The auto-merge workflow shall use exactly two triggers, `workflow_run` and
  `issue_comment` with the types `created` and `edited` (revision 13, REQ-CCR-022), and no trigger of the
  `pull_request_target` family, shall reference the workflow token as its only secret, and shall carry only steps that neither upload, download,
  nor cache an artifact; this SPEC's commits shall add zero workflow files.
- **REQ-CCR-018** (Event-driven) When either guard step this SPEC adds (`files`, `review`) ends any way other
  than its pass — the gate script refuses (exit 1), exits 11, or exits a code the step does not recognise; the
  pull request is unreadable or its head or count moved during or after the read; the `review` step's
  pre-merge recheck detects an incomplete/changing list, changed head or review-comment count, or a different trusted clean verdict;
  a list stays unreadable to the end of the window; or the gate
  script is missing — that step shall end with a non-zero exit status, so that the step conclusion and the job
  conclusion are failures and the later steps, the merge step included, do not run, and shall end with exit
  status 0 only on its pass — and, for `review` on the comment path only, on the pending "not yet" hold of REQ-CCR-023; the mapping from the gate script's exit code to the step's exit status, the job
  result, and the recorded output values, across both steps, shall be the single table of `design.md` §5
  (`T-CHAIN`), and the workflow shall mark neither step `continue-on-error` nor otherwise exempt from the
  job's conclusion.
- **REQ-CCR-019** (Ubiquitous) Each guard step this SPEC adds (`files` and `review`) shall record each of its
  output keys at most once, on its single terminating path, so that every key of a step's output carries
  exactly one value: `ok` is `true` or `false`; `state` is `pass` or `fail` for the `files` step and `pass`,
  `fail`, `not-run:draft`, or (comment path only, REQ-CCR-023) `pending` for the `review` step; a key the step never recorded reads as not-pass.
- **REQ-CCR-020** (Event-driven) When a pull request is not authored by the eligible author (login `GoosLab`, user id 6458969, type
  User, all three together), or its head repository is missing or is not the base repository, or its author's account type is Bot,
  auto-merge shall not merge it: a step that reads the pull request object once and records the single output `eligible` (`true` only when
  every condition holds, `false` on any other result, an unreadable object included) shall run before the check, `files`, and `review`
  steps, those steps and the merge step shall require `eligible` to be `true`, and the withheld pull request shall be annotated as needing
  the user's test confirmation and explicit approval; the step ends with exit status 0 like the `pr`, `head`, and `checks` steps (decision-index
  Q26). The author is bound by login, id, and type together, and an eligible pull request still merges only through every other gate of
  this SPEC: current-head CI and the required checks of its base, both review legs clean and bound to the head, no unresolved finding.
- **REQ-CCR-021** (Ubiquitous) The auto-merge workflow shall never delete a branch: its merge call shall carry no `--delete-branch` flag
  and no step shall delete a ref. The repository setting `delete_branch_on_merge` is a security setting that remains an operator decision
  (§5, G9) and is not changed by this SPEC.
- **REQ-CCR-022** (Event-driven) When an issue comment is created or edited, auto-merge shall evaluate a pull request on the
  **comment path** only when the comment belongs to a pull request (the event's `issue.pull_request` is present) and the comment's
  author has the numeric user id 199175422 (the Codex app identity), and shall take that decision in the job-level condition, so that a
  comment of any other author, or a comment on an issue that is not a pull request, starts a run whose job is skipped before any step
  runs, any token is used, or anything is checked out; on the comment path auto-merge shall take only the pull request number from the
  event payload, pass it to steps through the environment and never interpolate it into a script, read the head commit, the state, the
  merged and draft flags, and every listing from the API, read no comment body, title, or other text of the payload, check out the
  default branch alone, and run the gate script of that checkout — never a file, a ref, or a command of the pull request's head.
- **REQ-CCR-023** (Event-driven) When the comment path evaluates a pull request, auto-merge shall run the same steps in the same order
  with the same conditions as the `workflow_run` path (`pr`, `head`, `author`, `checks`, `files`, `review`, the merge step), shall take
  the commit under evaluation from the pull request's head as read from the API and bind every guard and the merge call to it, so that a
  head that moved after the identity's comment fails the binding and withholds; shall run the review read as one pass with the window
  open (`--wait-closed false`), with no wait and no sleep; when that pass is pending (gate exit 10: a leg or the current-head security
  marker still `running`, or no state comment yet), the `review` step shall end as a "not yet" hold — exit status 0, `ok` not `true`,
  the single state `pending`, the merge step not run, the job not failed, and the `auto-merge` status `pending` with the reason
  `review-not-complete:running` (or `review-not-complete:no-state-comment`) — and the identity's completion edit of its comment starts
  the next evaluation; every other result that is not clean is decided by `T-CHAIN` (revision 13.1, the lane ruling reversing
  decision-index Q31(a)); and auto-merge shall serialise evaluations in one job-level concurrency group whose key is the pull request
  number for the `workflow_run` path and for a comment of user id 199175422, and is unique per run (the run id) for a comment of any
  other author, so that at most one real evaluation of a pull request runs at a time, the two paths never both merge, and a run whose
  job the trust gate of REQ-CCR-022 skips can never share a group with, cancel, or replace a pending real evaluation.
- **REQ-CCR-024** (Event-driven) When the merge step runs (every guard passed) and does not merge — the overall deadline passed before
  the merge call or during the head re-read, the head read immediately before the call differs from the verified head, or `gh pr merge`
  exits non-zero — the merge step shall end with exit status 1, so that the job result is failure, and shall still write its one
  job-summary line naming the outcome and the head commit; the merge step shall end with exit status 0 only when `gh pr merge` exits 0.
  The steps `pr`, `head`, `author`, and `checks` keep exit status 0 when they withhold (the "not yet" holds), the comment path's pending
  `review` pass is a "not yet" hold with exit status 0 (REQ-CCR-023), and otherwise the `files` and `review` steps keep `T-CHAIN`; only
  the three outcomes above, where the merge step was reached and did not merge, are non-zero holds.
- **REQ-CCR-025** (Ubiquitous) Every evaluation of an open pull request whose evaluated head commit is known shall post exactly one
  commit status with the context `auto-merge` on that commit through the API, with the state `success` when the merge was requested
  (`gh pr merge` exited 0), `failure` when the merge step ran and did not merge or the `files` or `review` step ended non-zero other than
  `not-run:draft`, and `pending` for every other withheld outcome, and with a description naming the outcome's reason (`design.md`
  "Delta of plan revision 13", `T-STATUS`); the auto-merge workflow shall hold exactly the permissions `contents: write`,
  `pull-requests: write`, and `statuses: write` (one new scope), shall never read a commit status (the status is an output of the
  evaluation, never an input of the verdict), and shall end the status-writing step non-zero when the write fails, without changing the
  merge decision; the context shall not become a required check: `.github/required-checks.yml` and the branch protection stay
  byte-unchanged (REQ-CCR-016).

### 3.3 Retirement of the earlier reviewers

- **REQ-CCR-014** (Ubiquitous) The workflow directory shall hold zero workflows that run
  `anthropics/claude-code-action`, carry the name `Claude Code Review`, or trigger on a `check_run` event; the
  `@claude` responder workflow and the Claude review severity-parser workflow shall be removed.
- **REQ-CCR-015** (Ubiquitous) The CodeRabbit configuration file shall remain present with automatic review
  and automatic chat reply disabled, and the workflows under `.github/` and the scripts under `scripts/ci/`
  shall contain zero reads of a CodeRabbit status or comment.

### 3.4 Contract preservation and liveness

- **REQ-CCR-016** (Ubiquitous) The required-checks contract, the branch-protection template and its template
  mirror, and the workflows that publish or gate the measured required checks shall be byte-unchanged by this
  SPEC's own commits (measured from the merge base with local `main`, so another card's change absorbed from
  `main` is not read as this SPEC's), and the required, release-gate, and auxiliary lists shall hold zero
  Codex review checks until the required-checks confirmation gate (§5, G4) is passed.
- **REQ-CCR-017** (Ubiquitous) The guard-liveness manifest shall list exactly the workflows present in the
  workflow directory, so that the manifest and the workflow directory stay equal as sets after the removals.

## 4. Constraints

- Every deliverable is local and reversible: one revert of this SPEC's own commits restores the prior
  workflows, script set, liveness manifest, and CodeRabbit configuration (decision-index Q2).
- The run phase makes no remote write: no push, no pull request, no tag, no secret, no workflow dispatch,
  no comment, no review request, no branch-protection call, no app or settings change. Remote application is
  a leader-coordinated draft pull request (§5, G6), and every observation step on the remote is read-only
  (`gh api` GET).
- **Two states, worded separately.** *Landed locally* means this SPEC's commits are merged on local `main`
  (a merged-local SHA exists) and every criterion passes on that tree. *Remote verified* means both facts
  hold for one head commit `H` of a non-draft pull request, read by GET after the leader's coordination: a
  finished native review — the state comment of the Codex app identity, both legs bound to `H` — was
  observed (the gate script's own read of the fetched lists exits 0, or the findings are reported), AND the
  CI of `H` is green. A comment still `running` is not a finished review. A real
  review alone, or CI alone, is never "remote verified", and nothing is called complete before both are
  confirmed. This lane produces only the first state: no report of this lane says "complete", and
  "remote verified" is claimed only by the actor that read the remote data (`acceptance.md` §6 and §7).
- **A draft never reaches "remote verified".** On a draft pull request the observable thing is the gate
  path: auto-merge records `not-run:draft` for the head commit with a job result that is not success (gate
  exit class 11, REQ-CCR-008). That observation is worded `remote observed: gate path NOT RUN (not-run:draft)`;
  it never verifies that a review ran or what it returned, and required CI results on the same commit are
  reported as "required CI", never as a review verification. The leader-coordinated observation steps are
  DS-1 and DS-2 (§5) for the Codex app's behavior and the post-landing job observation of the G6 acceptance
  condition (§5) for the workflow. A non-draft pull request whose `review` step failed with cause
  `not-run:review-absent` is worded `remote observed: review step failed, review absent`, with the same
  limit.
- **The remote conclusion of the job is UNOBSERVED.** No GitHub-hosted run of the amended `auto-merge.yml`
  has been observed, so what the platform reports for the `review` step and its job — a failure for a refused
  or absent review, a result that is not success for a draft — is UNOBSERVED (research U11). The local
  criteria assert the step's exit status and a job result derived from the step statuses against a stub
  `gh`; they never claim what GitHub reports. The conclusion is first observable on the next pull request
  after the change lands on the default branch, through the leader-coordinated flow (§5, G6 acceptance
  condition); until then the remote state of the job is "unverified", never "verified" and never "passing".
- **Time is only ever a bound that withholds.** No requirement, step, or criterion turns elapsed time, a
  timeout, or the absence of a review into a pass or into a verified state: the wait window (at most 60
  passes of 15 seconds) and the existing 17-minute deadline only end a wait, and what ends it is a failure
  (`not-run:review-absent`, `review-unreadable`), `not-run:draft`, or a refusal. This lane never changes a setting, requests a review, or retries one; informing the operator is
  the leader's act.
- No step of this SPEC edits a non-test Go source file, a template file, or a distributed rule: the only Go
  files added are `_test.go` files, so no embedded template or catalog hash changes.
- The gate tests run offline: no network, no secret, no GitHub; a stub `gh` serves fixtures. Under linux and
  darwin a missing `bash` or `jq` fails the test rather than skipping it.
- `.github/` is not templated for any file this SPEC edits (research E9), so no template mirror applies to
  the workflows or the script.
- No time estimates, no test skips, no timeout increases.

## 5. Deferred confirmation gates and deferred observation steps

Each gate needs a user decision, a persistent app permission, a branch-protection change, a settings
change, a cost, or a change to `required-checks.yml` against the measured remote state. No requirement and
no acceptance criterion in this SPEC depends on a gate's or a step's answer; until a gate is passed the
review gate fails closed (merges withheld, never widened). Gate detail, confirmer, and interim state:
`decision-index.md` § Deferred confirmation gates. Local Claude and GLM sessions are unchanged. New API
keys, new app OAuth, persistent access, admin or security settings, branch-protection changes, and cost are
not approved and are never a deliverable of this card.

**Stated fact.** The gate passes only when the Codex app identity's single state comment reports both legs
`completed`, each bound to the head commit, each with `has_any_findings` and `has_unresolved_findings` false, and no corroborating
finding of the identity exists on that commit. No pull-request review object (an APPROVED one included), no
prose, no absence of a comment, a `running` leg, and no elapsed time is ever a pass. The form was read on
five real pull requests (DS-1, below), so the contract no longer rests on an unobserved artifact; what stays
unobserved is listed under DS-1 to DS-3 and in `research.md` E39. Until a pull request shows a state comment
that this reader passes, every merge by automation is withheld, and human merges are unaffected. Where a
later observation shows another `status` value or another marker version, the gate withholds until the
recognition contract is amended; that is a change of this contract, never a runtime option. The pull request
that introduces the gate cannot exercise it either: `workflow_run`, and (revision 13) `issue_comment`, run the
workflow file of the default branch, which does not yet carry the new gate (design §11).

| Gate | Needs | Interim state |
|------|-------|---------------|
| G1 | WITHDRAWN. The key-based review path (a provider API key as a repository secret) is not built; the native review needs no key | Not applicable |
| G2 | Change the native Codex review's settings or the app's permissions (per-push or draft review, a different repository selection, new OAuth, persistent access) | Not changed and not retried by this card; the operator-confirmed state stays as is |
| G3 | Add a Codex review check to main's branch-protection required contexts | Unchanged; the measured five required checks stay |
| G4 | Add a Codex review check to `required-checks.yml` | Unchanged; a permanent test (AC-CCR-017) blocks an unconfirmed edit |
| G5 | Uninstall or disable the CodeRabbit app at repository or organisation level | Repository config disables automatic review (REQ-CCR-015) |
| G6 | Apply this change remotely as a leader-coordinated draft pull request, including deletion of the remote-only `claude-code-review.yml` | A census test (AC-CCR-014) turns red if that file survives the merge; the G6 acceptance condition below says what that draft pull request can and cannot show |
| G7 | Delete the repository variable `AUTOMERGE_MAX_MERGE_RISK` | Unreferenced after the change; harmless |
| G8 | Delete the `CLAUDE_CODE_OAUTH_TOKEN` secret | Unreferenced after the change; harmless |
| G9 | Decide the repository setting `delete_branch_on_merge` (relayed by the leader as `true`; not read by this lane). It is a security setting and an operator decision item: this SPEC does not change it. While it is `true`, GitHub itself deletes a head branch after a merge, so removing `--delete-branch` from the workflow's own call (REQ-CCR-021) does not by itself stop deletion | Not changed; the workflow requests no deletion; an operator decision item, never a deliverable |

**G6 acceptance condition (the gap, stated once).** Auto-merge runs through the `workflow_run` trigger (and,
since revision 13, the `issue_comment` trigger, which likewise runs the default branch's workflow file) and
checks out the default branch, so it is the default branch's copy of `auto-merge.yml` that runs for any pull
request. The G6 draft pull request carries the amended workflow on its own branch only: its CI success proves
nothing about whether the amended workflow runs on the default branch, and that draft pull request cannot show
the `review` step's conclusion or the job result. What that draft pull request can show is the Codex app's own
behavior (DS-1, DS-2), which does not depend on this SPEC's workflow. The amended workflow's behavior on a real
pull request is observable only on the next pull request after the change lands on the default branch (the
leader-coordinated draft flow of `design.md` §12, step B), by read-only GETs of the Auto Merge run's job
conclusion and step log. Until that observation exists the remote state of the amended workflow is
"unverified": no report says "remote verified", "active", or "passing" for it, and nothing in this SPEC
depends on the observation (its local criteria are AC-CCR-001 to AC-CCR-025 on the merged-local tree).

| Step | What it observes (read-only GET; the lane never writes) | Status |
|------|--------------------------------------------------------|--------|
| DS-1 | The artifact the native review leaves on a pull request: which form, under which author login, numeric id, and type, bound to which commit; whether it matches the recognition contract of §3.1; whether a draft is reviewed; what a clean result looks like. A mismatch is an amendment of the contract, never a runtime option. | **OBSERVED on other pull requests (revision 7), NOT observed on #1828 itself, and not observed on a draft.** Dated fact (the lane's read of the raw REST objects saved on 2026-10-19, `research.md` E37 to E46; observer's report of the same day): on #1835, #1836, #1837, #1838, and #1841 the native review is one issue comment by login `chatgpt-codex-connector[bot]`, user id 199175422, type Bot, carrying the state marker; `created_at` is earlier than `updated_at` on all five (edited in place); the `reviews` list is empty on #1835, #1837, #1838, #1841; #1836 has one COMMENTED review of the identity (body marker `codex-review-receipt:v1`) and one inline review comment, and its state comment reports `has_unresolved_findings` true on the code leg. The final comment of each pull request reads both legs `completed`; on the four clean ones `commit_sha` equals the saved pull request head, on #1836 it is `e657c6b040` while the saved head is `4cee5eea1c`. All five trigger rows read "Draft marked ready"; a pull request opened non-draft was not observed. **Also read:** #1842 (E42), whose security leg reads `completed` with null fields while the security marker reads `running`; #1833 is read as both-`running` at 05:24Z and as completed with unresolved findings on the code leg at 05:36Z (E43, E44); #1844 shows a code leg `failed` (a review run that failed) with both finding flags true (E46). **Still unobserved:** the review on #1828 itself (draft; silent), a draft being reviewed or not under a controlled read, the whole `running` to `completed` transition (E39), the value set of `status`, and the effect of `@codex review` (DS-3). U2, U4, U5 |
| DS-2 | **Whether a Codex review runs again after a push** (the leader's item). The comment states three triggers (open for review, draft marked ready, an `@codex review` comment); the settings relayed by the leader: automatic review on for all pull requests, thorough review on, automatic security review on, minimum reported severity High (E33). Read-only GETs after a later push: the state comment's id, `updated_at`, and the `commit_sha` of each leg against the new head | **Open observation; no automatic re-review on push was observed.** The one sample with a moved head, #1836, keeps `commit_sha` `e657c6b040` in its state comment (last edit 2026-10-19T03:48:59Z) while another reviewer's review at 03:46:01Z is bound to the final head `4cee5eea1c` (E40). One sample is not a proof of absence. U6 |
| DS-3 | How a fresh review is requested for a head that moved after the review started: the leader re-requests with an `@codex review` comment (a remote write by the leader, outside this lane, never a deliverable). | **One re-request observed (#1842, read at 2026-10-19T05:36Z per the files' own `observed_at`; `research.md` E44).** After `@codex review` the existing state comment 6260255148 was edited in place: only `code_review.commit_sha` moved to the new head `e7cceb2cb5d6e1cd9dbd2f465c4b9d5a1d631f21`; a separate receipt comment 6260377784 of the same author was added, carrying no state marker; the `security_review` leg stayed `completed` with a null `commit_sha`, and the separate security marker still named the OLD head `45b2dbf682` with status `running`. That state is not a pass (REQ-CCR-001, REQ-CCR-002). Still unobserved: whether a later security review finishes for the new head, and a re-request that creates a second state comment (which the gate refuses, REQ-CCR-004). One sample. U13 |

## 6. Exclusions

### Out of Scope — remote and persistent changes

- Creating, rotating, or reading any key or secret; any cost-bearing change; any model-provider account
  step.
- Installing, connecting, uninstalling, or reconfiguring the Codex GitHub app or the CodeRabbit app, and
  changing the native review's settings (draft, per-push, repository selection, credit).
- Any branch-protection call or ruleset change, and any edit of `.github/required-checks.yml`,
  `.github/branch-protection.json.gtmpl`, or its template mirror.
- Pushing, opening, or merging a pull request, commenting on or requesting a review of a pull request,
  tagging, dispatching a workflow, or deleting the remote-only `claude-code-review.yml`; those belong to
  gate G6, the deferred steps, and the leader.

### Out of Scope — a review workflow of our own

- Any workflow that calls a review model, holds a key, publishes a verdict check run, or posts a review.
  The keyed `openai/codex-action` design of plan revisions 1 and 2 is rejected (decision-index Q1).

### Out of Scope — doctrine, role rules, and distributed templates

- Every tracked doctrine, role-prompt, template, projection, hook-notice, ledger, and operator-document
  surface that describes how a pull request's review is read: the factory dispatch rules and companions,
  `AGENTS.md` and its template, `manager-lead` and its template, the `.moai/policies/workflow/`
  projections, `.moai/project/tech.md`, the binding-ledger fixture, the hook role-core test, the
  four-locale lane notice, and the cutover runbook. They change the always-loaded surface and the
  template defaults; the operator's boundary was related docs, not role rules.
- They are recorded as follow-up candidates in `decision-index.md` § Follow-up candidates (recorded,
  not issued) and as known debt KD-1 (§7). No requirement or criterion of this SPEC claims that any of
  them is updated or depends on it.

### Out of Scope — files owned by other cards

- Files that cards t1865 (moai update ownership), t1632 (MCP rejection structuring), and t1731
  (GitHub Actions duplication, on hold) own: `ci.yml`, `codeql.yml`, `pr-multi-os-gate.yml`,
  `release-pr-multi-os.yml`, and anything under `internal/cli/update*` and `internal/mcp*`. This card
  goes first; t1731 absorbs any overlap on resume.
- Local GLM and local Claude sessions, `codex_review` and `claude_audit` tooling, and every non-test
  `internal/cli` source file.

### Out of Scope — review features not built now

- Reading the review from a check run, a commit status, a reaction (the thumbs-up of a clean review was not
  measured on the comment object), a pull-request review object, or any prose or table row of the comment:
  the verdict carrier is the state marker of the identity's state comment, cross-checked by the security
  marker, which can only refuse (revision 7, decision-index Q24). The `codex-pull-request-review-summary`
  marker is not read.
- Counting a pull-request review object of the identity in any state, an APPROVED one included, as a clean
  result: a review object or review comment can only refuse (REQ-CCR-005).
- Choosing among several state comments by recency, or reading a state comment that stands outside the
  marker position: both refuse (REQ-CCR-004).
- Checking the signature or provenance of the comment's text beyond the author binding: the body of a
  comment can be edited by a user with write access, which the gate cannot tell from the bot's own edit
  (design §6, residual risk 8).
- Parsing finding severity from a review comment: any finding withholds the merge for a human.
- Requesting a fresh review, re-running the native review, or any per-push review (DS-3, G2).
- Chunked or multi-pass review of very large diffs; whatever the native review does with a large pull
  request, the gate reads only what it leaves.
- Attesting the publisher of every other check run or comment in the repository; only the Codex app
  identity is bound.

### Out of Scope — what the author policy does not do

- Changing the repository setting `delete_branch_on_merge`, branch protection, a ruleset, or CODEOWNERS to express the author rule: the
  enforcement is the workflow's `eligible` output (REQ-CCR-020), and the setting is an operator decision item (G9).
- Collecting, recording, or checking the user's test confirmation or explicit approval for a fork, bot, or other author's pull request:
  that is the user's act outside the workflow; the workflow only never merges such a pull request.
- Changing how the `checks` step reads the required checks of an unprotected base (decision-index Q27): on a base with no required
  checks the shipped step withholds (E43), and this SPEC leaves that step as it is.

### Out of Scope — step conclusions other than the review step

- The exit status of the steps `pr`, `head`, `author`, and `checks`: they keep their existing convention (a
  withheld merge is exit status 0 with an annotation and an `ok`, `eligible`, or `should_merge` output) — these
  are the "not yet" holds, and since revision 13 each of them is also visible as a `pending` commit status
  `auto-merge` (REQ-CCR-025), never as a green run alone. **The merge step is no longer in this exclusion**
  (revision 13, operator ruling of 2026-10-23, decision-index Q30): a merge that the step reached and did not
  execute ends non-zero (REQ-CCR-024). Only `files` and `review` gain the one-value-per-key rule (REQ-CCR-019).
  Extending the non-zero rule to `pr`, `head`, `author`, and `checks` remains the alternate of decision-index
  Q22 and a separate ruling.
- A job-level skip for a draft (a pre-job that reads the draft flag and a `needs` condition on the job): not
  built; the alternate of decision-index Q21. The workflow stays one job.
- Any change to the gate script's behavior or exit classes (0, 1, 10, 11) and to its recognition contract.

### Out of Scope — what the comment path and the status context do not do

- Evaluating a pull request on a comment by any author other than user id 199175422, on a `pull_request_review` or
  `pull_request_review_comment` event, or on any other new trigger: the identity's verdict carrier is its issue comment, and it
  leaves no pull-request review for a clean result (research E37).
- Reading the comment body, the issue title or body, or any other text of the event payload; checking out or running anything of
  the pull request's head on either path.
- A wait loop on the comment path: it evaluates once, and the identity's next comment creation or edit is the next evaluation.
- Reading the `auto-merge` commit status, or any commit status, as an input; making `auto-merge` a required check; changing
  `.github/required-checks.yml`, branch protection, or a ruleset (gates G3 and G4 are unchanged).
- Re-requesting a review (`@codex review`) or re-running a workflow from the automation (DS-3 stays the leader's step).

### Out of Scope — historical and fixture text kept as is

- Completed SPEC bodies, the `CHANGELOG`, `internal/cli/testdata/codex-rollouts-t1171/**`,
  `internal/template/release_ci_gate_test.go` (its `CodeRabbit` row is an unrelated-app fixture), the
  explanatory comment at `scripts/release.sh:232` (outside `scripts/ci/`), and the explanatory CodeRabbit
  mention in `.moai/docs/batch-failure-bisect-runbook.md`.

## 7. Known debt

**KD-1 — the deployed doctrine still describes CodeRabbit as the review source — transferred to card
t1874.** Known and accepted by the leader ruling (b). Not repaired by this SPEC; no requirement or criterion
depends on it. Card t1874 owns the CodeRabbit rule replacement in `AGENTS.md` §4, the review-read contract, the
role cores, and the templates; it is blocked until t1868 is merged-local and DS-1 is observed on the G6 draft
pull request (leader ruling, board record d-20261018T152011Z-8447 (supersedes d-20261018T151824Z-db98; ruling and scope unchanged)). The rows below stay as the record of what
goes stale until t1874 lands; they are its input, not this card's work.

- Sentences that go stale once `auto-merge.yml` stops reading CodeRabbit and `.coderabbit.yaml` turns
  automatic review off (`research.md` E20 carries the file:line list): the `CodeRabbit is not read from
  gh pr checks` rule and its "counts only with combined-endpoint `success` + `Review completed` and a
  matching `Merge Risk:` line" predicate (`factory-dispatch-core.md:46-50`, `factory-dispatch.md:146-150`,
  `AGENTS.md` §4 paragraph at line 200), the deputy's "CodeRabbit two-condition READS"
  (`manager-lead.md:224, 236`), the endpoint-measurement companion (`factory-dispatch-gates.md:10-14`),
  the four-locale lane notice ("CodeRabbit via the commit-status API plus the Merge Risk line",
  `session_start_factory_i18n.go:140, 200, 260, 319`), and their template mirrors and
  `.moai/policies/workflow/` projections.
- Effect until the follow-up lands: the doctrine names a review source that no longer starts
  automatically and never names the native Codex review, so a reader applying it literally cannot
  satisfy its evidence predicate on a new pull request. The leader applies the ruling by judgment in
  the interim.
- Existing tests that pin the stale text stay unchanged and keep describing the unedited doctrine:
  `internal/hook/role_core_elements_test.go:51-52` (`Merge Risk:` and `/commits/{sha}/status`) and the
  binding-ledger rows read by `TestBindingLedgerIntegrity` (`binding_ledger_test.go:1131`). They were
  read, not run (plan-phase rule).
- Exit condition: card t1874 lands (the file groups FU-1 to FU-9 of `decision-index.md` § Follow-up
  candidates are its input), and the census command of `research.md` E13 reports no `Merge Risk` hit outside
  the historical allowlist. The card is issued as t1874; this plan does not issue or edit it.

### Accepted limitations (all fail closed; none is a defect to repair in this card)

**AL-1 — [SUPERSEDED 2026-10-23 — RESOLVED by REQ-CCR-022 and REQ-CCR-023, no longer an accepted limitation].** The operator and
leader ruled on 2026-10-23 (decision-index Q30) that a late review is a defect to fix in this card (audit finding F4), and withdrew the
accepted-limitation option. A review that finishes after the window, or after the pull request becomes ready, is now evaluated again
when the Codex app identity creates or edits its state comment (the comment path, REQ-CCR-022), through the same chain with no wait
(REQ-CCR-023); a merge reached and not executed ends non-zero (REQ-CCR-024) and every evaluation posts the `auto-merge` status
(REQ-CCR-025). What stays unobserved is the platform half: that GitHub delivers an `issue_comment` `edited` event for the bot's in-place
edit, observable only on the remote after landing (design "Delta of plan revision 13", residual R13-1). The history of the former entry
is kept below unchanged.

*(History, superseded)* **AL-1 — a late review is not picked up.** On the five measured pull requests the state comment was last
edited between 1 minute 14 seconds and 10 minutes 44 seconds after it was created (E39); on four of them
the comment was created 4 to 13 seconds after the pull request had already been merged (E41). That is five
observations, not a bound. If a review finishes after the wait window, the consequence is this: auto-merge waits at most 60 polls of 15
seconds (15 minutes) inside its 17-minute deadline and starts only when a `CI` run completes, so a review
that arrives after the window leaves the merge withheld and the `review` step failed (`not-run:review-absent`,
a failed job since revision 6) until auto-merge next runs. Re-trigger paths (operator actions outside this
lane): re-run the `CI` workflow run of the head commit, or re-run the failed Auto Merge job; that either
re-run fires auto-merge again, or re-reads the late review, is UNVERIFIED (research U7, U12). Human merges are
unaffected. The withholding itself is pinned by AC-CCR-008 (`review_absent_window_closed_exit_11`),
AC-CCR-021 (`review_absent_window_closed_step_exit_1_job_failure`), and AC-CCR-007 (`expiry_never_merges`).

**AL-2 — a head that moved after the review started, and pull requests opened before the review was
switched on.** A review starts on open-for-review, on draft-marked-ready, or on an `@codex review` comment;
no automatic review on a push was observed (DS-2), so a push after the review leaves no state comment bound to
the new head until the leader re-requests one, and a pull request opened before the contract has none at all. The gate withholds as stale (`stale-review`) or as a failed `review` step (`not-run:review-absent`). A fresh review is
requested only through the leader-coordinated step DS-3, never by this card. This is the common case for
any pull request that receives a later push; it is a stated consequence of the contract, and human merges
are unaffected.
