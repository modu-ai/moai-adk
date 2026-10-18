# design.md — SPEC-CI-CODEX-REVIEW-001

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
`spec.md`. Evidence ids (E*, U*) refer to `research.md`. Plan revision 4: the reviewer is the ChatGPT
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
pull request opens  --(operator-confirmed contract: native review runs on open; not yet observed, DS-1)-->
   chatgpt-codex-connector[bot] leaves its result on the pull request   (form UNVERIFIED, U2)

[auto-merge.yml]  (workflow_run of CI, default-branch copy)
   step "files":  read the PR object (head, changed_files); every page of the changed-file list, count-checked
                  and bound entry by entry to the verified head; read the PR object again;
                  any change under .github/workflows/ or to the gate script, an unbound or unprovable list,
                  or a moved head -> withhold
   step "review": per pass: read the PR (draft, head, review_comments), every page of reviews and of review
                  comments (review comments count-checked), then read-native-review
        (every line below is the gate script's result; the step's own exit status, job result, and outputs for
         each are the one table T-CHAIN of section 5)
        draft                                             -> exit 11 not-run:draft: state not-run:draft, step exit 11
        newest review by the Codex identity bound to H is APPROVED, submitted, no finding on H, and no
        CHANGES_REQUESTED that a later APPROVED has not superseded
                                                          -> exit 0, then the pre-merge recheck: step exit 0, merge
        no review by the identity yet, window open        -> wait (exit 10), not a terminating path
        no review by the identity, window closed          -> exit 11 not-run:review-absent: a FAILURE, step exit 1
        review bound to another commit only               -> stale-review (exit 1): step exit 1 at once
        COMMENTED, DISMISSED, PENDING, CHANGES_REQUESTED, or a finding -> fail (exit 1): step exit 1 at once
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
leaves this script contract untouched** (same subcommands, same exit classes, same reasons); what changed is the
consumer's reading of the exit classes: the `review` step turns every exit other than 0 into a non-zero step
exit, the draft cause keeps its own state, and the review-absent cause is a failure (§5, `T-CHAIN`).

### 3.1 The recognition contract (constants in the script)

| Constant | Value | Evidence |
|----------|-------|----------|
| Identity login | `chatgpt-codex-connector[bot]` | E27 |
| Identity numeric user id | `199175422` | E27 |
| Identity account type | `Bot` | E27 |
| Recognised artifact form | a pull-request review (`pulls/N/reviews`) | E23 (commit binding is a field of this object), U2 |
| Clean state | `APPROVED` only. This is the interim contract until DS-1 shows what a clean native review looks like; FU-10 adapts it | spec §5 stated fact, Q20 |
| Never clean | `COMMENTED` (with or without inline comments, whatever its body says), `CHANGES_REQUESTED`, `DISMISSED`, `PENDING`, any other state | REQ-CCR-004 |
| Field never read | the review `body`: the script names no `body` key, so body text cannot change a verdict in either direction (AC-CCR-001 `script_never_references_body_field`) | REQ-CCR-001 |
| Changed-file entry binding | for an entry whose `status` is not `removed`, the `ref` value in the query of `contents_url` equals the verified head SHA | E31, U10 |
| Allowed list endpoints | `repos/<owner>/<repo>/pulls/<N>/(files\|reviews\|comments)` and nothing else | REQ-CCR-006, AC-CCR-006 |
| Gate paths | any path under `.github/workflows/`; `scripts/ci/codex-review-gate.sh` | spec §3 terms |

An artifact is recognised only when login, id, and type all match; a lookalike login, the same login with
another id, or a User account of that login is not the identity (the `[bot]` suffix cannot be registered by
a user account, and the numeric id is immutable). Exactly one identity is recognised: a second reviewer
(another bot, a human) is never read, in any state. The contract is a statement about the native review's
form that the lane could not observe (E29, U2, U5, and the dated DS-1 observation of spec §5); it is held in
one place so that DS-1 can confirm or refute it, and the gate fails closed on whatever does not match.

### 3.2 `read-native-review --reviews F --comments G --head-sha S --draft true|false [--wait-closed true|false]`

`F` and `G` are NDJSON files (one object per line) written by `fetch-pages` (§3.3): the pull request's
reviews and review comments. Fields read, and nothing else: from each review `id`, `user.login`, `user.id`,
`user.type`, `commit_id`, `state`, `submitted_at`; from each review comment `id`, `user.login`, `user.id`,
`user.type`, `commit_id`, `original_commit_id`. Read-only. `--wait-closed` (default `false`) says that the
caller's wait window has ended; it changes one result only (step 4). Checked in this order:

0. **Flags.** `--draft` or `--wait-closed` that is neither `true` nor `false` → exit 1 `draft-flag-unusable` /
   `wait-closed-flag-unusable`; `S` that is not 40 hexadecimal characters → exit 1 `head-sha-unusable`.
1. `--draft true` → exit 11, reason `not-run:draft`, before any file is read (REQ-CCR-008): a draft is never
   merged by automation, and whether the native review runs on a draft is unconfirmed (U4), so nothing here
   depends on it.
2. `F` or `G` missing, or a line that is not a JSON object → exit 1 `input-invalid`.
3. **Candidates**: review entries whose `user.login`, `user.id`, and `user.type` all equal the identity.
   Entries of any other author are ignored entirely, in every state and whatever their body says; the
   exclusion is by author, and no body text is read (REQ-CCR-003). A candidate lacking an integer `id`, a
   string `commit_id`, or a string `state` → exit 1 `review-entry-invalid:<id>`; two candidates with the same
   `id` → exit 1 `review-entry-invalid:duplicate-id:<id>`.
4. **Absence.** No candidate at all: with `--wait-closed false` → exit 10 `pending:no-review`; with
   `--wait-closed true` → exit 11 `not-run:review-absent` (REQ-CCR-008). A candidate exists → continue.
5. **Binding**: the bound reviews are the candidates whose `commit_id` equals `S` (40 characters,
   case-insensitive; a prefix or an empty string is not equal). If none is bound, whatever else exists →
   exit 1 `stale-review:<commit prefix of the highest-id candidate>` (REQ-CCR-002; a review exists, so this is
   a failure and never NOT RUN, whatever `--wait-closed` says).
6. **Supersession of a change request.** Let `C` be the highest `id` among bound reviews in state
   `CHANGES_REQUESTED` and `A` the highest `id` among bound reviews in state `APPROVED`. If `C` exists and
   `A` does not exist or `C` is greater than `A` → exit 1 `changes-requested`. A CHANGES_REQUESTED review is
   superseded only by a later bound APPROVED review; COMMENTED, DISMISSED, and time never supersede it
   (REQ-CCR-004).
7. **Findings**: any review comment by the identity whose `commit_id` or `original_commit_id` equals `S` →
   exit 1 `finding:<id>`; one that carries neither field → exit 1 `finding-unbound:<id>` (fail closed;
   REQ-CCR-005). A comment by the identity bound only to another commit is a stale finding of an older head
   and is ignored once a current review is clean.
8. **The newest bound review decides, and must be APPROVED.** `N` is the bound review with the highest `id`.
   `N` in state `COMMENTED` → exit 1 `review-not-conclusive:COMMENTED`; `N` in any other state but `APPROVED`
   (`DISMISSED`, `PENDING`, anything else) → exit 1 `review-not-complete:<state>`; `N` in state `APPROVED`
   with a `submitted_at` that is not a non-empty string → exit 1 `review-not-complete:no-submitted-at`
   (REQ-CCR-001, REQ-CCR-004). (`N` is never `CHANGES_REQUESTED` here: step 6 refused it.)
9. Otherwise exit 0 and print `native-review-clean: review <id> state APPROVED bound to <S prefix>`.

**The one rule, in a sentence.** The result is clean only when the newest bound review of the identity is an
APPROVED review with a submission time, no finding exists on the head, and any earlier CHANGES_REQUESTED has
been superseded by a later APPROVED. The decision reads state, binding, and finding comments; it never reads
body text, so a COMMENTED review does not become clean whatever its body says, and an APPROVED review does not
become unclean by its body.

**Decision table** (H is the head, O another commit; `(id,commit)`; "window" is the `--wait-closed` value):

| # | Reviews of the identity | Result | Reason |
|---|-------------------------|--------|--------|
| 1 | APPROVED(10,H), no finding | exit 0 | `native-review-clean` |
| 2 | APPROVED(10,H) whose body says "could not review" or carries a usage-limit notice | exit 0 | body never read; residual risk, §6 |
| 3 | COMMENTED(10,H), empty body, no inline comment | exit 1 | `review-not-conclusive:COMMENTED` |
| 4 | COMMENTED(10,H), body `LGTM` | exit 1 | `review-not-conclusive:COMMENTED` |
| 5 | COMMENTED(10,H), findings only in the body | exit 1 | `review-not-conclusive:COMMENTED` |
| 6 | COMMENTED(10,H), body is a usage-limit notice or says "could not review" | exit 1 | `review-not-conclusive:COMMENTED` |
| 7 | COMMENTED(10,H) with an inline comment bound to H | exit 1 | `finding:<id>` |
| 8 | APPROVED(10,H) then CHANGES_REQUESTED(11,H) | exit 1 | `changes-requested` |
| 9 | CHANGES_REQUESTED(10,H) then APPROVED(11,H), no finding | exit 0 | clean (superseded) |
| 10 | CHANGES_REQUESTED(10,H) then COMMENTED(11,H) | exit 1 | `changes-requested` |
| 11 | CHANGES_REQUESTED(10,H) then DISMISSED(11,H) | exit 1 | `changes-requested` |
| 12 | APPROVED(10,H) then COMMENTED(11,H) | exit 1 | `review-not-conclusive:COMMENTED` |
| 13 | APPROVED(10,H) then DISMISSED(11,H) | exit 1 | `review-not-complete:DISMISSED` |
| 14 | DISMISSED(10,H) then APPROVED(11,H), no finding | exit 0 | clean |
| 15 | APPROVED(10,O), then COMMENTED(11,H) | exit 1 | `review-not-conclusive:COMMENTED` (the approval belongs to another commit) |
| 16 | APPROVED(10,O) only | exit 1 | `stale-review:<O prefix>` |
| 17 | PENDING(10,H), or a state the gate does not know | exit 1 | `review-not-complete:<state>` |
| 18 | APPROVED(10,H), `submitted_at` null or empty | exit 1 | `review-not-complete:no-submitted-at` |
| 19 | APPROVED(10,H) by the identity and CHANGES_REQUESTED(11,H) by another login | exit 0 | another author is never read (REQ-CCR-003) |
| 20 | no review by the identity; another login's APPROVED(10,H) | window open: exit 10; window closed: exit 11 | `pending:no-review` / `not-run:review-absent` |
| 21 | no review at all | window open: exit 10; window closed: exit 11 | `pending:no-review` / `not-run:review-absent` |
| 22 | pull request is a draft | exit 11 | `not-run:draft` |
| 23 | a list cannot be proven complete: a page error, a count mismatch, a full 30th page, a repeated entry | `fetch-pages` exits 1 and removes the file; the consumer counts the pass as pending and, if still unreadable at the end of the window, reports a failure `review-unreadable` (not NOT RUN) | `page-error`, `count-mismatch`, `page-ceiling`, `page-duplicate` |
| 24 | the API silently shortens the reviews list with no error and no count to compare | not detectable | residual risk, §6 (U9) |
| 25 | CHANGES_REQUESTED(10,H) alone | exit 1 | `changes-requested` |

Rows 1-22 and 25 are fixtures of AC-CCR-001 to AC-CCR-005 and AC-CCR-008; row 23 is AC-CCR-007 and
AC-CCR-010; row 24 is stated, not tested.

Exit classes: 0 pass, 1 fail, 10 pending, 11 NOT RUN; pairwise distinct. The four conditions the operator
named map as follows, each never a pass: **missing review** → exit 10 while the window is open, then exit 11
`not-run:review-absent` (the script's NOT RUN class; the consumer ends the `review` step in failure on it,
§5 `T-CHAIN` row 4, never success); **missing output** (a
candidate lacking a required field) → exit 1 `review-entry-invalid`; **failed review** (not APPROVED, changes
requested, a finding) → exit 1; **stale SHA** (review bound to another commit) → exit 1 `stale-review`
(`acceptance.md` §3.1).

**Residual shape risk, stated and fail-closed.** If the native review reports findings only in a review's
body text, with no review comment, the gate does not read them as a finding (body text is never read) — but
such a review is a COMMENTED review, which is never clean, so it withholds. If the native review reports a
clean result as a COMMENTED review, a comment, or a check run, no clean state is ever recognised and the gate
withholds forever; DS-1 observes which, and FU-10 adapts the contract. Both are documented and neither is a
pass-by-absence.

### 3.3 `fetch-pages --endpoint E --out F [--expect-count N]`

The only subcommand that reads from the API, and the only place pagination lives. `E` is a repository
relative API path whose response is a JSON array; `F` receives one JSON object per line.

0. **Endpoint allow-list.** `E` must match `^repos/[A-Za-z0-9._-]+/[A-Za-z0-9._-]+/pulls/[0-9]+/(files|reviews|comments)$`;
   anything else → exit 1 `endpoint-not-allowed:<E>` before any `gh` call, so no issue-comment, check-run, or
   status endpoint can be read through the script (AC-CCR-006).
1. For page P = 1, 2, … it runs `gh api "<E>?per_page=100&page=<P>"`. A non-zero exit of `gh` refuses with
   `page-error:<P>`; a response that is not an array of objects refuses with `page-invalid:<P>`.
2. A page with fewer than 100 entries ends the collection. A page with exactly 100 entries asks for the next
   page, so a listing whose size is an exact multiple of 100 ends on an empty page.
3. The collection never asks past page 30. When page 30 is full (3000 entries) it refuses with
   `page-ceiling`: the files listing's own limit is 3000 entries (U3), so a full 30th page cannot be proven
   complete and is never read as complete.
4. **Count proof.** With `--expect-count N` (a non-negative integer, else `expect-count-unusable`) a
   collected count different from `N` refuses with `count-mismatch:<collected>:<N>`. The consumer passes the
   pull request's own `changed_files` for the files listing and `review_comments` for the review-comment
   listing; the two counts agree with the paginated lists on every pull request measured (E22, E23). The
   reviews listing has no count on the pull request object, so it is collected without a count proof and its
   completeness rests on rules 1-3 and 5 (U9).
5. **Entry identity.** An entry's key is its integer `id` (reviews, review comments) or its string `filename`
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
  d-20261018T155527Z-4e9e).** The steps `pr`, `head`, and `checks` keep the convention of the closed SPEC:
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

  **`T-CHAIN` — gate script exit, to step exit, to job result, to outputs** (every terminating path of the
  two steps this SPEC adds: rows 1 to 9 are the `review` step, rows 10 to 13 the `files` step; the gate
  script's exits are unchanged)

  | # | Terminating path | Gate exit | Step exit | Job result | `ok` | `state` | Annotation |
  |---|------------------|-----------|-----------|------------|------|---------|------------|
  | 1 | Clean pass, and the pre-merge recheck reads the same head, count, and selected review | 0, then 0 with the same summary | 0 | success; the merge step may run | `true` | `pass` | none beyond the clean summary line |
  | 2 | The gate refuses: `changes-requested`, `finding`, `finding-unbound`, `stale-review`, `review-not-conclusive`, `review-not-complete`, `review-entry-invalid`, `input-invalid`, an unusable flag | 1 | 1 | failure | `false` | `fail` | `::error::native review gate: <reason> - merge withheld` |
  | 3 | The gate exits a code that is none of 0, 1, 10, 11 | other | 1 | failure | `false` | `fail` | `::error::` naming `gate-exit-<n>` |
  | 4 | Non-draft, no native review observed when the window ends | 11 (`not-run:review-absent`) | 1 | failure | `false` | `fail` | `::error::native review gate: not-run:review-absent - no native review observed for the head commit` |
  | 5 | Draft pull request | 11 (`not-run:draft`) | 11 | failure; never success (decision-index Q21) | `false` | `not-run:draft` | `::notice::NOT RUN: draft - not a review result and never a pass; the merge is withheld` |
  | 6 | The head or the review-comment count moved before or during the list reads (`head-moved`, `head-moved-during-read`, `review-comments-moved-during-read`) | not reached | 1 | failure | `false` | `fail` | `::error::` with the reason |
  | 7 | The pre-merge recheck differs (`recheck-unreadable`, `head-moved-before-merge`, `review-comments-moved-before-merge`, `review-changed-before-merge`, `selected-review-changed-before-merge`) | 0, then another exit or a different summary | 1 | failure | `false` | `fail` | `::error::` with the reason |
  | 8 | A list is unreadable on the last pass, so the loop ends with no stop (`review-unreadable`) | not reached or `fetch-pages` refused | 1 | failure | `false` | `fail` | `::error::` with the reason |
  | 9 | The gate script is missing (`gate-script-missing`) | not run | 1 | failure | `false` | `fail` | `::error::` with the reason |
  | — | Gate exit 10 with the window open (not a terminating path) | 10 | none: sleep 15 seconds, next pass | — | not recorded | not recorded | `native review not observed yet (n/60)` |

  | 10 | `files`: the changed-file list passes (every read agrees, `fetch-pages` 0, `gate-files-changed` 0) | 0, 0 | 0 | success; the `review` step may run | `true` | `pass` | none |
  | 11 | `files`: the gate script is missing (`gate-script-missing`) | not run | 1 | failure | `false` | `fail` | `::error::changed-file check: gate-script-missing - merge withheld` |
  | 12 | `files`: the pull request is unreadable or the head or count moved (`files-unreadable`, `head-moved:<prefix>`, `files-unreadable-after-read`, `head-moved-during-read`, `count-moved-during-read`) | not reached | 1 | failure | `false` | `fail` | `::error::` with the reason |
  | 13 | `files`: `fetch-pages` or `gate-files-changed` refuses (`page-error`, `count-mismatch`, `gate-path-changed:<path>`, `file-entry-unbound:<name>`, any refusal) or exits a code it does not recognise | 1 or other | 1 | failure | `false` | `fail` | `::error::` with the reason |

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
- The two new steps run after the required-check observation and before the merge, in this order:
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
     request object (`draft`, `head.sha`, `review_comments`; a failed or empty read counts as pending with the
     note `unreadable`); (b) **if `draft` is `true`, call `read-native-review --draft true` at once, before
     any list is fetched**, so a draft ends as `not-run:draft` whatever a list fetch would return; (c) a
     `head.sha` other than the verified SHA ends the step in failure, `head-moved` (`T-CHAIN` row 6); (d) run `fetch-pages` on
     `.../reviews` (no count) and on `.../comments` with `--expect-count` set to `review_comments` (a refusal
     counts as pending with the note `unreadable:<reason>`); (d2) **read the pull request object again, after
     the list reads and before the gate script runs**: a `head.sha` other than the verified SHA ends the step
     in failure, `head-moved-during-read` (the lists just read belong to another head and are never judged;
     `T-CHAIN` row 6), and a failed or empty second read counts as pending with the note `unreadable`; (e)
     run `read-native-review` with the verified head SHA of step `head`, the `draft` flag, and `--wait-closed
     true` on the 60th pass only and `false` on every earlier pass. A pass therefore reads the head before and
     after its list reads, as the `files` step does (REQ-CCR-012); the draft path of (b) reads no list and so
     has nothing to bracket. Mapping of the exit (the full table is `T-CHAIN`): **0** → the pre-merge
     recheck, then step exit 0 with `ok=true`, `state=pass` (row 1). **10** → sleep 15 seconds, next pass.
     **11** → stop at once: for a draft, `state=not-run:draft` and step exit 11 (row 5); for a non-draft pull
     request, `state=fail` and step exit 1 with the cause `not-run:review-absent` (row 4) — the loop reads
     `draft` from the pull request object it just read, as before. **1, or any exit not listed** → `ok=false`,
     `state=fail`, the reason, step exit 1, stop at once (rows 2 and 3). If the loop ends with no stop — only
     possible when the last pass was unreadable — the step ends in failure, `review-unreadable: <last note>`
     (row 8): a list that could not be proven complete is a refusal (REQ-CCR-010), never a pass and, since
     revision 6, never a green step. Absence on the 60th pass is exit 11 `not-run:review-absent`, produced by
     the script and not by the loop, so there is one place where "no review" is decided. (Sync-phase
     correction, ruling B: the second pull request read also compares
     `review_comments` with the first collection's count, and a clean pass is rechecked before `ok=true` by
     re-collecting both lists between two further reads and requiring the same head, the same count, and the
     same selected review from `read-native-review`; any change ends the step as a failure.)
- The merge step's `if` replaces `steps.coderabbit.outputs.ok == 'true'` by
  `steps.files.outputs.ok == 'true' && steps.review.outputs.ok == 'true'`; the existing terms, the
  `--match-head-commit`, deadline re-evaluation, and `release/*` merge-flag logic are untouched.
- The checkout step keeps no `ref`. The step named `Check all required CI checks passed` keeps its name so
  `TestAutoMergeRequiredChecks` still finds it.
- **The `gh` allow-list is exact** (REQ-CCR-006, AC-CCR-006). The workflow makes exactly these calls and no
  others: `gh pr view "$PR_NUMBER" --json headRefOid` (steps `head` and merge), `gh pr checks "$PR_NUMBER"
  --required --json name,state,bucket` (step `checks`, the unchanged CI observation), `gh pr merge
  "$PR_NUMBER" … --match-head-commit … --delete-branch` (merge), and `gh api` calls whose only path is the
  pull request object `repos/$REPO/pulls/$PR_NUMBER` (steps `files` and `review`); the gate script makes
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
| **A comment by any user saying "approved"** (issue comment, review body, review comment) | Pass the consumer without a real review | Only a pull-request review by the Codex identity is read; issue comments are never read; no body text is parsed | AC-CCR-003, AC-CCR-006 |
| **A review by a different author or app** (another bot such as `coderabbitai[bot]`, a human) | A bound, APPROVED review from the wrong author | Login, numeric id, and account type must all equal the Codex identity; another login's review is ignored in every state | AC-CCR-003 |
| **A lookalike**: the same login without `[bot]`, the same login with another id, a User account of that login | Impersonation | The three-part identity; `[bot]` logins cannot be registered by users; the id is immutable | AC-CCR-003 |
| **A review of another commit** (older head, prefix only, empty) | Pass a review written for earlier code | `commit_id` must equal the full head SHA; a review bound only elsewhere is `stale-review`; an APPROVED review of another commit never counts | AC-CCR-002 |
| **A push after the pull request opened** (the review runs only on open) | The old review reads as current | Same binding: the moved head has no bound review → `stale-review`, withheld at once; fresh review is DS-3 | AC-CCR-002, AC-CCR-007 |
| **A COMMENTED review whose body says "LGTM", reports findings, or carries a usage-limit or "could not review" notice** | A review that reviewed nothing, or reported a problem only in prose, read as clean | Only state APPROVED is clean; COMMENTED (with or without inline comments) is `review-not-conclusive`; body text is never read (the script names no `body` key) | AC-CCR-004, AC-CCR-001 |
| **CHANGES_REQUESTED followed by a later COMMENTED, DISMISSED, or nothing** | The request is masked and the merge proceeds | Any bound CHANGES_REQUESTED refuses unless a later bound APPROVED supersedes it | AC-CCR-004 |
| Missing review; an incomplete one (`DISMISSED`, `PENDING`, no submission time); an entry lacking a required field | A review that never finished read as a pass | `pending` then a failed step `not-run:review-absent` at the end of the window; `review-not-complete`; `review-entry-invalid` | AC-CCR-004, AC-CCR-007, AC-CCR-008, AC-CCR-021 |
| A finding (review comment by the Codex identity on the head) or `CHANGES_REQUESTED` | Merge past a reported problem | Refuse; severity is not parsed, a human decides | AC-CCR-005 |
| **PR adds or edits any workflow, or the gate script** | The PR changes the gate for itself, or adds a workflow that merges itself or acts with the workflow token | Tamper guard withholds on ANY file under `.github/workflows/` and on the gate script; the consumer runs from the default branch, never the PR ref | AC-CCR-009, AC-CCR-011 |
| **A gate-path edit past the first page of the changed-file list** | The edit on page 2 or later is never read (default page size 30, E22) | `fetch-pages` collects every page at 100 per page, requires the collected count to equal `changed_files`, refuses at the 3000 ceiling | AC-CCR-010 |
| **The head flips from H to T and back while the changed-file list is read** | The list of T (a safe change) is judged and H (a workflow change) is merged | Entries that are not removals must carry `ref` = the verified head SHA in `contents_url`; the head is read before and after the list and must equal the verified SHA | AC-CCR-009, AC-CCR-012 |
| **A list API fails mid-pagination, is truncated, or shifts between pages** | A partial list read as complete | A page error refuses and removes the output; count proof for files and review comments; a repeated entry refuses; the step records `ok=true` only on exit 0 | AC-CCR-010, AC-CCR-012 |
| **The consumer calls the guard and ignores the result, or never calls it** | The guard exists but binds nothing | Structural test of the real step order, exit-code handling, and merge condition, with mutants | AC-CCR-012 |
| **The merge step never merges, or merges past a refusal** | A `&& false` in the merge condition, an early `exit 0`, an unbounded deadline, or a dropped term in the condition goes unnoticed | The chain from the head step to the merge step runs on fixtures; `gh pr merge` is called once, pinned to the verified head, on the clean fixture and zero times on every refusal | AC-CCR-019 |
| **The gate is pointed at another data source** | A verdict read from an issue comment, a check run, or a status | The workflow's `gh` calls are an exact allow-list; `fetch-pages` accepts only the files, reviews, and comments listings of a pull request | AC-CCR-006 |
| Draft pull request | A work-in-progress merged, or a draft observation read as a review result or as a green job | Exit 11 at once, whatever artifacts exist: state `not-run:draft`, step exit 11, job result never success; the message says gate path only | AC-CCR-008, AC-CCR-022 |
| **Elapsed time or the absence of a review read as success** | A review that never came is treated as "nothing to fix" | Time only bounds a wait; the end of the window is a failed step `not-run:review-absent` (gate exit 11) or `review-unreadable`, never a pass; the expiry fixtures assert that no merge happens | AC-CCR-007, AC-CCR-008, AC-CCR-019, AC-CCR-021 |
| Comment, label, check run, or commit status read as a verdict | Pass the consumer without any review | The consumer reads only reviews and review comments (census) | AC-CCR-006 |
| `pull_request_target` or a secret in a workflow this SPEC edits | Secrets with attacker-controlled code | The `workflow_run` trigger alone, no secret reference other than the workflow token (none added) | AC-CCR-013 |
| Auto-merge checks out the PR ref | Attacker code runs with a write token | Checkout has no `ref` | AC-CCR-011 |
| **Artifact or cache poisoning** | Not applicable: nothing is uploaded, downloaded, or cached; the lists are temporary files of one job | No surface exists; a test finds no artifact or cache step | AC-CCR-013 |
| A review arrives after the consumer's window | A late review never read | Withhold as a failed step `not-run:review-absent`; re-trigger paths (accepted limitation AL-1) | AC-CCR-007, AC-CCR-008, AC-CCR-021 |
| **A review failure the step swallows** (revision 6: the step wrote `ok=false` and exited 0) | A refused review, a changes-requested review, or a draft shows as a green job, so a human or a tool reading the job conclusion reads a pass | Every non-pass path ends the step non-zero (`T-CHAIN`), the step is not `continue-on-error`, and the tests assert step exit status and derived job result, not only the outputs | AC-CCR-020, AC-CCR-024 |
| **A key written twice in a step's output** (revision 6: `ok=false` then `ok=true`, `state=fail` then `state=not-run`) | A reader that takes the first value, or a platform that takes the last, reads a different verdict than the step's own terminating path | Each key recorded once on the terminating path, no default first (`T-OUTPUTS`) | AC-CCR-023 |

**Residual risk that no mechanical check closes** (each stated, none a pass-by-absence):

1. The native review's own judgment: a model can miss a defect or be manipulated by pull-request content, and
   an APPROVED review whose body says otherwise passes, because body text is never read (a contract choice:
   the body's shape is unobserved, and a parser would be a guess).
2. A clean result that is not an APPROVED pull-request review (a COMMENTED review, a comment, a check run) is
   never recognised, so the gate may withhold forever; DS-1 observes the form and FU-10 adapts the contract.
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
7. A few seconds remain between the review step's pre-merge recheck and the merge call itself: a review state
   change in that window, with the head unchanged, is not detected (the merge is pinned to the head with
   `--match-head-commit`, so the code under review cannot change in it); a dismissal also leaves no count
   on the pull request object, so only the re-collected reviews list can show it.

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
pull request's review is read from a pull-request review authored by the Codex app identity, bound to the
current `headRefOid`, newest, in state APPROVED, with no finding on it", which the consumer of §5 already
implements (once DS-1 confirms the form).

## 10. Test layout

All new tests are `_test.go` files, none under `internal/template/templates/`, so the neutrality guard does
not scan them.

| File | Tests |
|------|-------|
| `internal/template/codex_review_gate_test.go` | `TestCodexGateReadNativeReviewAccepts`, `TestCodexGateReadNativeReviewRefusesStale`, `TestCodexGateReadNativeReviewIgnoresForged`, `TestCodexGateReadNativeReviewRefusesIncomplete`, `TestCodexGateReadNativeReviewRefusesFinding`, `TestCodexGateReadNativeReviewPending`, `TestCodexGateReadNativeReviewDraftNotRun` |
| `internal/template/codex_review_paging_test.go` | `TestCodexGateFetchPagesCompleteness`, `TestCodexGateFetchPagesRefusesDisallowedEndpoint`, `TestCodexGateFilesChangedRefusals` |
| `internal/template/auto_merge_codex_test.go` | `TestAutoMergeReadsNoCommentOrStatusVerdict`, `TestAutoMergeWaitLoopIsBounded`, `TestAutoMergeNotRunWithholds`, `TestAutoMergeCheckoutAndGuardsUnchanged`, `TestAutoMergeWiresGuardAndGate`, `TestAutoMergeAddsNoSecretTriggerOrArtifact`, `TestAutoMergeMergeDecisionPath`; revision 6 adds `TestAutoMergeReviewFailureFailsStepAndJob`, `TestAutoMergeReviewAbsentIsFailure`, `TestAutoMergeDraftIsNotRunNeverSuccess`, `TestAutoMergeGuardStepsOutputOneValuePerKey`, `TestAutoMergeStepHarnessObservesExitAndJobResult` |
| `internal/template/review_retirement_test.go` | `TestNoClaudeReviewWorkflows`, `TestCodeRabbitDisabledAndUnread` |
| `internal/config/required_checks_review_guard_test.go` | `TestRequiredChecksCarryNoCodexReviewCheckUntilConfirmed` |
| `internal/guardstate/census_review_test.go` | `TestCensusOmitsRetiredReviewWorkflows` |

Fixtures are inline strings and `t.TempDir()` files built in the shape of the documented REST review and
review-comment objects (E23); the tests copy the real script and run it with `bash`. The tests that exercise
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

- **Until a real APPROVED review by the Codex identity is observed, every merge by automation is withheld.**
  The gate counts one thing as clean — an APPROVED pull-request review by the Codex identity bound to the head
  commit (§3.2); no review of any kind has been observed (E29, and the dated DS-1 observation of spec §5),
  so the gate has nothing to pass. If the native review's own clean result is another form, it stays
  withheld until FU-10 adapts the contract after DS-1. Human merges are unaffected.
- The pull request that introduces the gate cannot exercise it: `workflow_run` runs the default branch's
  `auto-merge.yml`, which does not yet carry the gate. It is merged by its owner (gate G6). The first
  pull request that can show the gate working is a later one, after landing. The G6 draft pull request's CI
  success proves nothing about the amended workflow on the default branch (spec §5, G6 acceptance condition);
  the remote conclusion of the `review` step and of the job is UNOBSERVED until then (research U11), and the
  remote state is "unverified".
- **Every pull request that has no APPROVED native review now shows a failed `Auto Merge` job** (revision 6).
  Before the amendment a withheld merge was a green job; now a refusal, an absent review, and a draft each
  end the `review` step non-zero, so the job is red on any pull request the gate does not pass. That is the
  ruling's intent (a failure must not read as a pass) and a visible cost: `Auto Merge` is not a required
  check (research E6), so the red blocks no human merge, but it appears on the pull request and on every
  draft. A job-level skip for drafts would need a second job (decision-index Q21, alternate).
- **Every pull request that adds, changes, deletes, or renames any file under `.github/workflows/`, or the
  gate script, is withheld from auto-merge** and needs a human merge.
- **A push after the pull request opened leaves it withheld.** The review runs only at open (the
  operator-confirmed contract, per-push widening excluded), so a moved head has no bound review →
  `stale-review`. A fresh review is the leader-coordinated step DS-3, not a deliverable (AL-2).
- A draft pull request is recorded as `not-run:draft` (step exit 11, job result never success) and withheld.
  Whether the native review runs on a draft, whether a clean result is an APPROVED review (DS-1), and whether
  a review runs again after the ready-for-review transition or a push (DS-2) are unconfirmed (U4, U6) and
  nothing here depends on the answer.
- **AL-1 (a late review is not picked up).** If a review arrives after the 15-minute window (that it arrives
  at all is an unverified expectation), the merge stays withheld and the `review` step has failed
  (`not-run:review-absent`) until auto-merge next runs; the re-trigger paths are re-running `CI` for the head
  commit (U7) or re-running the failed Auto Merge job (U12). Fail closed. Pinned: AC-CCR-008
  `review_absent_window_closed_exit_11`, AC-CCR-021 `review_absent_window_closed_step_exit_1_job_failure`,
  AC-CCR-007 `expiry_never_merges`.
- Dependabot and other bot-authored pull requests are withheld unless the native review reviews them; CodeRabbit
  used to review them (U8).
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
gh api "repos/modu-ai/moai-adk/issues/<N>/comments?per_page=100&page=<P>"   # issue comments (never a verdict; read to learn the form, DS-1)
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
   is a clean result an APPROVED pull-request review (DS-1)? After the ready-for-review transition (a remote
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
