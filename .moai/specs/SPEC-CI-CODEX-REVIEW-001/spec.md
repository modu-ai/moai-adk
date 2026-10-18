---
id: SPEC-CI-CODEX-REVIEW-001
title: "CI review switch — fail-closed native Codex review gate replaces the CodeRabbit and GitHub Claude review surfaces"
version: "0.6.0"
status: in-progress
created: 2026-10-18
updated: 2026-10-18
author: "Goos Kim"
priority: P1
phase: "v3.2.0 target"
module: ".github/workflows"
lifecycle: spec-anchored
tags: "ci, review-gate, codex, coderabbit, github-actions, fail-closed, auto-merge"
tier: L
amendment_of: SPEC-CI-CODEX-REVIEW-001
related_specs: [SPEC-V3R6-CODERABBIT-ADOPTION-001, SPEC-CI-VERDICT-INTEGRITY-001]
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

## Amendments

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
branch; no merged-local SHA exists).

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
- **The artifact is unverified.** None of the 120 most recent pull requests of the two repositories (60
  each) carries a review, a review comment, or an issue comment authored by the Codex app, or by any login
  naming codex, chatgpt, or openai (E29). Which artifact the native review leaves on a pull request, whether
  it runs on a draft, and what it leaves for a clean result are therefore UNVERIFIED (U2, U4, U5); a later
  dated observation (§5, DS-1, 2026-10-18) also found no Codex trace on a pull request about 47 minutes after
  it opened. The gate recognises exactly one form — a pull-request review — counts one state of it as clean
  (APPROVED), and fails closed on every other (§3.1).
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

1. For a non-draft pull request, `auto-merge.yml` merges only when the pull request carries a native Codex
   review — a pull-request review authored by the Codex app identity, bound to the exact head commit, in
   state APPROVED and newest among the identity's reviews of that commit, with no finding comment on that
   commit — in addition to the unchanged CI observations. No issue comment, label, review body text,
   third-party status, or review by any other author is a verdict, and no elapsed time is a pass.
2. No CodeRabbit, Claude review, or `@claude` responder surface remains in the repository's workflows or
   under `scripts/ci/`, except the historical and fixture text named in §6. Doctrine surfaces are not changed
   by this card and are carried as known debt KD-1 (§7).
3. Every state the gate cannot verify — no recognisable review, a review of another commit, a review that is
   not APPROVED, a finding, an unreadable or unprovably complete list — leaves the merge withheld with a
   named reason; nothing is ever read as a pass by absence or by the passing of time.
4. A draft pull request is reported `not-run:draft` — a state of its own that is never a pass — and the
   `review` step and the job do not end in success; a non-draft pull request with no native review observed
   for its head commit when the wait window ends is a failure of the step and the job (cause
   `not-run:review-absent`). Auto-merge withholds the merge on both; a draft report verifies the gate path
   only. Whether the native review runs on a draft, whether a clean result is an APPROVED review, and whether
   a review runs again after the ready-for-review transition or a push are items to confirm by observation
   (§5, DS-1, DS-2).
5. A failure in either guard step this SPEC adds is a failure of that step and of the job: every path of the
   `files` and `review` steps other than their pass — a missing gate script included — ends with a non-zero
   exit status, so the later steps and the merge do not run and no failed gate reads as a green job
   (REQ-CCR-018). The remote conclusion of that job is UNOBSERVED until the leader-coordinated flow after
   landing (§4, §5 G6).

## 3. Requirements (GEARS)

Notation: GEARS. Behavior is stated; file, script, and function names live in `design.md`.
Terms: the **gate script** is the repository's single review-verdict checker; the **Codex app identity** is
the account with login `chatgpt-codex-connector[bot]`, numeric user id 199175422, and account type Bot; a
**native review** is a pull-request review authored by that identity; it is **bound** to a commit when its
`commit_id` equals that commit; the **newest** native review bound to a commit is the one with the highest
numeric review `id`; a **finding** is a review comment authored by that identity and bound to the head
commit; the **gate paths** are every file under the workflows directory and the gate script; **clean** is
the one state REQ-CCR-001 defines; **NOT RUN** is the gate script's exit class 11, with exactly two causes: a
draft pull request, or no native review observed for the head commit when the wait window ends. At the
consumer the two causes are not one result: a draft is the recorded state `not-run:draft` with a job result
that is never success, and a non-draft pull request whose window ended with no review is a failure
(REQ-CCR-008, REQ-CCR-018). The **step exit status** and the **job result** are what the workflow platform
reports for the `review` step and its job; **a pass** is the one case where the step exits 0.

### 3.1 Native review verification

- **REQ-CCR-001** (Ubiquitous) The gate script shall report a pull request's native review as clean only when
  the newest native review bound to the exact head commit under review is in state APPROVED and carries a
  submission time, and no finding exists on that commit; the gate script shall take this decision from review
  state, binding, and finding comments alone, so that review body text, in either direction, changes no
  verdict.
- **REQ-CCR-002** (Event-driven) When every native review of the pull request is bound to a commit other
  than the head commit under review — including a head that moved after the pull request opened — the gate
  script shall refuse and name the stale commit.
- **REQ-CCR-003** (Event-driven) When a review, review comment, or issue comment is authored by any login, any
  numeric id, or an account type other than the Codex app identity's, the gate script shall exclude it from
  the verdict because of its author, in every state and whatever its body text says, so that the verdict
  derives only from reviews and review comments of the Codex app identity; the gate script shall read no body
  text of any artifact, so that a review of the Codex app identity is neither excluded nor demoted by what its
  body says (REQ-CCR-001).
- **REQ-CCR-004** (Event-driven) When the newest native review bound to the head commit is in any state other
  than APPROVED — COMMENTED (with or without inline comments, whatever its body says), CHANGES_REQUESTED,
  DISMISSED, PENDING, or an unknown state — or lacks a submission time or a required field, or when a native
  review in state CHANGES_REQUESTED is bound to the head commit and no later native review in state APPROVED
  is bound to the same commit, the gate script shall refuse and name the reason. A CHANGES_REQUESTED review is
  superseded only by a later APPROVED native review bound to the same commit; no other state, no body text,
  and no elapsed time supersedes it.
- **REQ-CCR-005** (Event-driven) When a finding exists on the head commit, the gate script shall refuse and
  name it; when none exists, the findings check shall pass.

### 3.2 Auto-merge consumer

- **REQ-CCR-006** (Ubiquitous) Auto-merge shall authorize a merge only when the gate script's read of the
  pull request's native review exits 0 for the exact head commit, in addition to the unchanged CI
  observations, shall take the review verdict from that exit status alone, and shall read, for that verdict,
  only the pull request object and its changed-file, review, and review-comment listings; the `gh` calls the
  workflow and the gate script make shall be exactly the allow-list of `acceptance.md` AC-CCR-006.
- **REQ-CCR-007** (Event-driven) When no native review is recognised yet, auto-merge shall keep waiting inside
  a bounded window (at most 60 passes of 15 seconds; time appears only as this bound, and no elapsed time is
  ever a pass); when the read refuses, auto-merge shall withhold the merge at once and end the `review` step
  as a failure (REQ-CCR-018); when a list the verdict needs stays unreadable to the end of the window,
  auto-merge shall withhold the merge and report a failure named `review-unreadable`, a failed step and a
  failed job, never NOT RUN. A review that arrives after the window expired is read only when auto-merge next
  runs (§7, AL-1).
- **REQ-CCR-008** (Event-driven) When the pull request is a draft, auto-merge shall record the single state
  `not-run:draft` — distinct in gate exit class (11), annotation, and recorded state from pass and from
  failure — shall withhold the merge, and shall end the `review` step with a non-zero exit status so that the
  job result is never success, whether or not any review exists on a draft; when the pull request is not a
  draft and the wait window ends with no native review observed for the head commit (gate exit class 11, cause
  `not-run:review-absent`), auto-merge shall report a failure — a failed step and a failed job — naming that
  cause, and shall withhold the merge; a draft report verifies the gate path only and is not a review result.
- **REQ-CCR-009** (Event-driven) When a pull request adds, changes, deletes, or renames any gate path,
  auto-merge shall withhold the merge for human decision.
- **REQ-CCR-010** (Event-driven) When a list the gate script decides on — changed files, reviews, or review
  comments — cannot be proven complete and bound (a page cannot be read, the collected count differs from the
  pull request's own count where one exists, an entry repeats, the API's page ceiling is reached, the
  changed-file list is empty, or a changed-file entry that is not a removal is not bound to the verified head
  commit), the gate script shall refuse and, where the refusal comes from collecting a list, remove its output
  list.
- **REQ-CCR-011** (Ubiquitous) Auto-merge shall check out the default branch alone (a pull-request ref is
  never checked out) and shall keep unchanged its head-commit re-verification, required-check observation,
  overall deadline, and head-pinned merge call.
- **REQ-CCR-012** (Ubiquitous) The auto-merge workflow shall run, in this order and before its merge step, a
  step that collects and checks the pull request's changed-file list and a step that reads the native review;
  each step shall read the pull request's head commit before and after its list reads (the review step on
  every pass that reads lists) and record a failing result when that head differs from the verified head
  commit; each shall record a passing result only when
  the corresponding gate-script call exits 0; each shall treat any other exit, including one it does not
  recognise, as a withheld merge (and each guard step, as a failed step, REQ-CCR-018); and the merge step
  shall run only when both steps recorded a passing result.
- **REQ-CCR-013** (Ubiquitous) The auto-merge workflow shall use the `workflow_run` trigger alone, shall
  reference the workflow token as its only secret, and shall carry only steps that neither upload, download,
  nor cache an artifact; this SPEC's commits shall add zero workflow files.
- **REQ-CCR-018** (Event-driven) When either guard step this SPEC adds (`files`, `review`) ends any way other
  than its pass — the gate script refuses (exit 1), exits 11, or exits a code the step does not recognise; the
  pull request is unreadable or its head or count moved during or after the read; the `review` step's
  pre-merge recheck differs from the clean pass; a list stays unreadable to the end of the window; or the gate
  script is missing — that step shall end with a non-zero exit status, so that the step conclusion and the job
  conclusion are failures and the later steps, the merge step included, do not run, and shall end with exit
  status 0 only on its pass; the mapping from the gate script's exit code to the step's exit status, the job
  result, and the recorded output values, across both steps, shall be the single table of `design.md` §5
  (`T-CHAIN`), and the workflow shall mark neither step `continue-on-error` nor otherwise exempt from the
  job's conclusion.
- **REQ-CCR-019** (Ubiquitous) Each guard step this SPEC adds (`files` and `review`) shall record each of its
  output keys at most once, on its single terminating path, so that every key of a step's output carries
  exactly one value: `ok` is `true` or `false`; `state` is `pass` or `fail` for the `files` step and `pass`,
  `fail`, or `not-run:draft` for the `review` step; a key the step never recorded reads as not-pass.

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
  native review authored by the Codex app identity and bound to `H` was observed (the gate script's own
  read of the fetched lists exits 0, or the findings are reported), AND the CI of `H` is green. A real
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

**Stated fact.** The gate passes only when the Codex app identity has submitted a pull-request review in
state APPROVED, bound to the head commit and newest among the identity's reviews of that commit, with no
finding on that commit; a CHANGES_REQUESTED review is superseded only by a later APPROVED. No other review
state (COMMENTED with or without inline comments, DISMISSED, PENDING), no review body text, no absence of a
review, and no elapsed time is ever a pass; this rule does not depend on the answer of DS-1 or on any
review's body. The native review has not been observed to submit any review on any pull request (the dated
DS-1 observation below), so until a real APPROVED review by the identity is observed every merge by
automation is withheld, and human merges are unaffected. If the native review's clean result is another form
(for example a COMMENTED review with no inline comment, or a comment), the gate keeps withholding until
follow-up FU-10 adapts the recognition contract after DS-1; that is a change of this contract, never a
runtime option. The pull request that introduces the gate cannot exercise it either: `workflow_run` runs the
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

**G6 acceptance condition (the gap, stated once).** Auto-merge runs through the `workflow_run` trigger and
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
| DS-1 | The artifact the native review leaves on a pull request: which form (pull-request review, review comment, issue comment, or check run), under which author login, numeric id, and type, bound to which commit; whether it matches the recognition contract of §3.1; and, from the leader's revision-6 wording, whether the review actually runs on a draft and whether a clean result stays as an APPROVED pull-request review (with the relayed minimum reported severity High, E33, a pull request with nothing at that level may leave a different form). A mismatch is a follow-up (adapt the contract, FU-10), never a change to this card. The Codex app's behavior does not depend on this SPEC's workflow, so it can be read on the G6 draft pull request itself | **UNCONFIRMED; no observation has found a review. DS-1 remains unobserved, which is not success.** Dated fact (observer: the leader, read-only `gh api` GETs on 2026-10-18 at 12:05Z, relayed to this lane; command shape: the pull request object, its `reviews`, its `issues/<N>/comments`, its `issues/<N>/timeline`, and `commits/<head>/check-runs`): modu-ai/mo.ai.kr#435 was created 2026-10-18T11:17:58Z, is not a draft, head 386f0d6c9c; about 47 minutes after creation there is no Codex trace — reviews 0, issue comments only from vercel[bot] and coderabbitai[bot], 0 codex, chatgpt, or openai actors in the timeline, head check runs only vercel and github-actions. The iteration-2 plan auditor re-censused 15 pull requests per repository, including mo.ai.kr#435 and #436, and found no Codex artifact. This lane re-read #435 (age 82 minutes) and #436 (created 11:54:19Z, age 45 minutes) at 2026-10-18T12:40:08Z with the same result. So the operator's "automatic review ON when a pull request opens" has not been observed to produce a review on that pull request (a delay is possible); the form stays UNCONFIRMED and is confirmed only when a real review is observed. The gate treats the lack of an observation as a failure (`not-run:review-absent`, REQ-CCR-008) and never as a pass; no setting is changed and no retry is made by the lane. Holder of this fact: this row; command-level record: `.moai/reports/t1868/progress.md` § DS-1 observation. U2, U5 |
| DS-2 | **Whether a Codex review runs again after the ready-for-review transition or after a push** (the leader's item). The repository's trigger is "when a pull request is opened"; the settings relayed by the leader from a read-only observation (this lane did not observe them): automatic review on for all pull requests, thorough review on, automatic security review on, minimum reported severity High (E33). Read-only GETs after the transition and after a later push: the pull request's reviews and review comments, each artifact's author login, numeric id, type, and `commit_id`, and which GET shows it (`design.md` §12). If no review appears bound to the new head, that is the stated consequence (AL-2) and DS-3 is the path to ask for one | **Open observation, not run.** U4, U6 |
| DS-3 | How a fresh review would be requested for a head that moved after the pull request opened (the review runs only at open; per-push review is excluded): an operator or leader action on the remote, outside this lane. Until then a moved head is withheld as stale | Not run; U6; a leader-coordinated request, never a deliverable |

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

- Reading the review from an issue comment, a check run, or a commit status, or from review body text: the
  recognised form is a pull-request review only, until DS-1 shows otherwise.
- Counting a COMMENTED review, with or without inline comments, or any non-review form as a clean result:
  APPROVED is the only clean state until DS-1 shows what a clean native review looks like (FU-10).
- Parsing finding severity from a review comment: any finding withholds the merge for a human.
- Requesting a fresh review, re-running the native review, or any per-push review (DS-3, G2).
- Chunked or multi-pass review of very large diffs; whatever the native review does with a large pull
  request, the gate reads only what it leaves.
- Attesting the publisher of every other check run or comment in the repository; only the Codex app
  identity is bound.

### Out of Scope — step conclusions other than the review step

- The exit status of the steps `pr`, `head`, `checks`, and the merge step: they keep their existing
  convention (a withheld merge is exit status 0 with an annotation and an `ok` or `should_merge` output).
  Only the two steps this SPEC adds, `files` and `review`, end non-zero on a failure (REQ-CCR-018) and gain the
  one-value-per-key rule (REQ-CCR-019). Extending the non-zero rule to the earlier steps is the alternate of
  decision-index Q22 and a separate ruling.
- A job-level skip for a draft (a pre-job that reads the draft flag and a `needs` condition on the job): not
  built; the alternate of decision-index Q21. The workflow stays one job.
- Any change to the gate script's behavior or exit classes (0, 1, 10, 11) and to its recognition contract.

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

**AL-1 — a late review is not picked up.** Unverified expectation, not a fact: that the native review arrives
some time after the pull request opens (neither whether it arrives at all nor its delay has been observed;
DS-1). If it arrives after the wait window, the consequence is this: auto-merge waits at most 60 polls of 15
seconds (15 minutes) inside its 17-minute deadline and starts only when a `CI` run completes, so a review
that arrives after the window leaves the merge withheld and the `review` step failed (`not-run:review-absent`,
a failed job since revision 6) until auto-merge next runs. Re-trigger paths (operator actions outside this
lane): re-run the `CI` workflow run of the head commit, or re-run the failed Auto Merge job; that either
re-run fires auto-merge again, or re-reads the late review, is UNVERIFIED (research U7, U12). Human merges are
unaffected. The withholding itself is pinned by AC-CCR-008 (`review_absent_window_closed_exit_11`),
AC-CCR-021 (`review_absent_window_closed_step_exit_1_job_failure`), and AC-CCR-007 (`expiry_never_merges`).

**AL-2 — a head that moved after the pull request opened, and pull requests opened before the review was
switched on.** The review runs only when a pull request opens (the operator-confirmed contract), so a push
after the open leaves no review bound to the new head, and a pull request opened before the contract has
none at all. The gate withholds as stale (`stale-review`) or as a failed `review` step (`not-run:review-absent`). A fresh review is
requested only through the leader-coordinated step DS-3, never by this card. This is the common case for
any pull request that receives a later push; it is a stated consequence of the contract, and human merges
are unaffected.
