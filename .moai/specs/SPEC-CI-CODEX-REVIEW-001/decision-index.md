# decision-index.md — SPEC-CI-CODEX-REVIEW-001

Stateless on the status axis — the SPEC lifecycle lives in `spec.md` alone.
Authored at plan phase (card t1868), revised twice before and twice after the first audit (plan revisions 3
and 4). Revision 4 (after plan audit iteration 2) added Q20 (what counts as clean), updated Q9, Q11, Q15, and
Q19 for the NOT RUN review-absent result and the APPROVED-only clean state, added follow-up FU-11, and moved
the DS-1 status to the dated fact in `spec.md` §5.
`interview.decision_gate: on` (`.moai/config/sections/interview.yaml`, line 6). Revision 3 (after plan audit
iteration 1 and the operator-confirmed design correction relayed by the leader) changed Q1 (the reviewer is
the native Codex review; the keyed action is the rejected alternate), Q5, Q6, Q9, Q13, Q14 (labels), Q15, and
added Q16-Q19; Q7, Q10, and Q11 existed only for the keyed action and are superseded; the deferred-gate table
withdrew G1 and added the deferred observation steps DS-1 to DS-3. **Revision 6** (an in-place amendment under
the leader's rulings of board records d-20261018T152011Z-8447 (supersedes d-20261018T151824Z-db98; ruling and scope unchanged) and d-20261018T155527Z-4e9e) added Q21 and Q22,
restated Q9 and Q15 for the consumer's split of NOT RUN, restated DS-1 and DS-2, and transferred KD-1 to card
t1874; the delta is in `design.md` "Delta of plan revision 6".

**Revision 7** (2026-10-19, an in-place amendment, Amendment 2) closed Q20 and FU-10, restated Q6 and Q19, and added Q23 to Q28 (corroboration, the security
marker, issue-comment noise, the `author` step's exit convention, the unprotected base, the author policy and branch deletion); DS-1, DS-2, DS-3 restated, G9 added; the
delta is in `design.md` "Delta of plan revision 7".

**Revision 13** (2026-10-23, an in-place amendment, Amendment 6) added Q30 (the operator and leader ruling on audit findings F4 and
F3/F1) and Q31 (the comment path's reading of a non-clean state and the status mapping), and superseded Q22 for the merge step; the
delta is in `design.md` "Delta of plan revision 13".

How to read this file. The card text, the operator-confirmed contract, and the repository settled most
questions; each settled one is a `### Q` row with its evidence and the rejected alternate. A product-level
row whose answer is the operator order recorded in the card or the correction relayed by the leader carries
`OPERATOR-SETTLED`, transcribed from that text — not a fresh operator reply. An implementation-level row
carries the `Default` selected by the published rule (preserve current behavior; else single-revert undo;
else smaller user-visible surface) and `DEFAULT-APPLIED`. Options that would violate a stated requirement (a
forgeable verdict carrier, a widening knob) are excluded before the rule ranks the rest. Items that need a
key, a cost, a persistent app permission, a settings change, a branch-protection change, or a
`required-checks.yml` change are NOT rows here: they are the deferred confirmation gates after the last row,
and no requirement or criterion depends on their answer.

### Q1: Which reviewer does the gate read — the native Codex GitHub review of the existing ChatGPT subscription, or a key-based `openai/codex-action` workflow of our own?

Label: FOUNDER
Class: product-level (changes which review source auto-merge trusts; until the artifact is observed no merge
is made by the automation)
Authority anchor: none in the committed register — the settlement is the operator-confirmed contract relayed
by the leader on 2026-10-18 (a message, not a committed artifact)
Why unresolved: revisions 1 and 2 built the key-based path because no key-free path was known; the operator
then confirmed in a logged-in browser that the subscription's native review is ON for modu-ai/moai-adk and
modu-ai/mo.ai.kr, for all pull requests, runs when a pull request opens, with the personal-review credit OFF
(per-push widening excluded; settings changes not retried by us)
Default: not applied (settled by the operator-confirmed contract)
Resolution: read the native review through the GitHub API. Evidence the lane measured by GET: the app
`chatgpt-codex-connector` (owner `openai`) exists and its bot user is `chatgpt-codex-connector[bot]`, id
199175422, type Bot (research E27); it is installed on the `modu-ai` organisation for all repositories and not
suspended (E28). Not observed: any artifact the app left on any of 120 recent pull requests (E29), so its
form is UNVERIFIED (U2) and the gate fails closed (design §3)
Rejected alternate: the key-based `openai/codex-action` workflow with an `OPENAI_API_KEY` secret (plan
revisions 1 and 2) — rejected: it adds a secret, a usage cost, and a deferred gate (G1) for a capability the
subscription already provides, and it adds a workflow whose verdict carrier needed the binding and
tamper-guard machinery of audit defect D1. A review of the Codex app path without our own workflow is the
smaller surface and one revert undoes it
Operator verdict: OPERATOR-SETTLED 2026-10-18 operator-confirmed contract relayed by the leader (native Codex
review ON, on pull-request open; replacement scope only GitHub CodeRabbit and the GitHub Claude
review/Actions), transcribed from the correction message

### Q2: Replace in one change, or run CodeRabbit and the native Codex review side by side until the native review is confirmed?

Label: FOUNDER
Class: product-level (changes auto-merge's default behavior: until the native review's artifact is observed
and recognised, no merge is made by the automation)
Authority anchor: none in the committed register — the order is in the card text only
Why unresolved: the published Default rule would rank side-by-side first (it preserves current behavior); the
card text says "Replace and verify the CodeRabbit and GitHub Claude review/Actions ... with the official Codex
review", which the rule does not override on its own
Default: not applied (the rule's pick is side-by-side; the card order differs, so the row is settled by the
order rather than by the rule)
Alternate: side-by-side — rejected because it keeps a third-party gate that cannot reach a verdict on the PR
shape this repository merges (research U1, E3), needs a repository variable as a switch (a widening path the
card forbids), and is not the replacement the card orders
Consequence stated, not hidden: until the artifact is observed (DS-1) every automated merge is withheld;
human merges are unaffected; one revert of this SPEC's commits restores the prior gate
Operator verdict: OPERATOR-SETTLED 2026-10-18 card t1868 order, scope clause "Replace and verify the
CodeRabbit and GitHub Claude review/Actions ... with the official Codex review" (transcribed from the card
text)

### Q3: Is the `@claude` responder workflow (`claude.yml`) removed?

Label: FOUNDER
Class: product-level (removes a user-facing capability: `@claude` replies on issues and pull requests)
Authority anchor: none for the decision; context only — SPEC-V3R6-CODERABBIT-ADOPTION-001 §B/§C
(`.moai/specs/SPEC-V3R6-CODERABBIT-ADOPTION-001/spec.md`) retained the responder when it made CodeRabbit the
single reviewer, so its row does not decide this question under the present order
Why unresolved: the responder is not a reviewer, but the card scope lists `claude.yml` among the files to
replace and the order covers "GitHub Claude review/Actions"
Default: not applied (settled by the order)
Alternate: keep `claude.yml` and remove only the review surfaces — rejected: it leaves
`anthropics/claude-code-action` and the `CLAUDE_CODE_OAUTH_TOKEN` secret in use, which the census of AC-CCR-014
forbids, and contradicts "Actions" in the order
Operator verdict: OPERATOR-SETTLED 2026-10-18 card t1868 order, scope list names `claude.yml` (transcribed from
the card text)

### Q4: What happens to `review-quality-gate.yml`?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: it parses a check run named `Claude Code Review` (research E4); once the Claude review
workflow is gone it can never fire usefully, but it could be repurposed
Default: delete it (rule: smaller user-visible surface — one fewer workflow; undo is a single revert)
Alternate: repurpose it to parse a review severity from a `check_run` trigger — rejected: it would duplicate
the gate script and add a `check_run`-triggered workflow with a write token
Operator verdict: DEFAULT-APPLIED 2026-10-18T11:53:23Z manager-spec (lane-22, card t1868)

### Q5: Which findings block?

Label: FOUNDER
Class: implementation-level
Authority anchor: none — the prior ceiling lives in `.github/workflows/auto-merge.yml` (lines 195-199, default
`low`), which is a workflow, not an operator setting in `.moai/config/sections/*.yaml`
Why unresolved: the native review's finding and severity vocabulary was never observed (research U5), so the
first rule (preserve the old "admit the two lowest rungs") cannot be applied to a vocabulary that is unknown
Default: any review comment by the Codex identity bound to the head commit, and any review by it in state
CHANGES_REQUESTED, withholds the merge for a human; no severity is parsed (rule: the first rule is
inapplicable; undo is a single revert in both options; smaller surface — no severity parser, no widening knob)
Alternate: parse a priority badge in the comment body and admit the lowest — rejected: it depends on an
unobserved text shape and is a widening path
Operator verdict: DEFAULT-APPLIED 2026-10-18T11:53:23Z manager-spec (lane-22, card t1868)

### Q6: Which fact does the consumer read as the verdict?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: candidates are a pull-request review, a review comment, an issue comment, a check run, a
commit status, or an uploaded artifact
Default: a pull-request review authored by the Codex identity, bound to the exact head commit by its
`commit_id` (excluded before ranking: an issue comment — it carries no commit binding and any user can post
the marker text; a commit status — any token with `statuses: write` can set it; a check run — nothing observed
shows the native app publishes one (research E29), and a same-named check run is forgeable by a repository
workflow; an uploaded artifact — none exists. Among the rest the rule picks the only form with a documented
commit binding: the review object carries `commit_id` (research E23, measured on a bot review))
Alternate: read an issue comment from the Codex bot — rejected until DS-1 shows that form; adding it would
need a commit marker the lane cannot verify
Operator verdict: DEFAULT-APPLIED 2026-10-18T11:53:23Z manager-spec (lane-22, card t1868)
Revision 7 (restated): the measured form (research E37) is the former alternate. The verdict carrier is the state marker of the identity's single issue comment;
the commit binding is the marker's `commit_sha` fields and the security marker's `headSha`, compared with the head, not a text guess; a review object is not a
carrier (it does not exist for a clean review) and can only refuse (Q23). The author binding of the first default (login, id, type) is unchanged.

### Q7: What privilege does the Codex step run with? — SUPERSEDED

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the question existed only for the key-based action (`sandbox`, `safety-strategy`); Q1 rejected
that action, so there is no Codex step and no privilege to choose
Default: not applied (superseded by Q1; no workflow of our own runs a review)
Alternate: none remaining
Operator verdict: OPERATOR-SETTLED 2026-10-18 superseded by the operator-confirmed native-review contract
(Q1)

### Q8: What becomes of `.coderabbit.yaml`?

Label: FOUNDER
Class: product-level (turns off the automatic PR review CodeRabbit gives contributors)
Authority anchor: none
Why unresolved: deleting the file returns the app to its defaults, which review (research E10); the card scope
names the file
Default: not applied (settled by the order)
Alternate: delete the file — rejected: it re-enables default review; uninstalling the app is gate G5
Resolution: keep the file; set `reviews.auto_review.enabled: false` and `chat.auto_reply: false`; leave every
other key so a re-enable is a two-line revert
Operator verdict: OPERATOR-SETTLED 2026-10-18 card t1868 order, scope list names `.coderabbit.yaml`
(transcribed from the card text)

### Q9: When does the review run, for whom, and what does the gate do about a head that moved or a draft?

Label: FOUNDER
Class: implementation-level
Authority anchor: none in the committed register — the contract is the operator-confirmed state relayed by the
leader on 2026-10-18 (a message, not a committed artifact)
Why unresolved: the native review runs "when a pull request opens"; whether it also runs on a draft, at the
ready-for-review transition, or after a push is not observed (research U4, U6), and per-push widening is
excluded by the operator
Default: the gate reads whatever exists and designs nothing on the answer: a draft ends as `not-run:draft`
(gate exit 11, step exit 11, job result never success) at once; a review bound to another commit is
`stale-review` (a moved head withholds); no review at window expiry is gate exit 11 `not-run:review-absent`,
which the consumer ends as a failed step (revision 6; never success); a fresh review after a push
is the leader-coordinated step DS-3 and never a deliverable
(rule: preserves current behavior — the old gate also withheld an unreviewed or moved head)
Alternate: widen the review to every push or to drafts — rejected here: it is a settings change (gate G2) the
operator excluded; revisions 1 and 2 had decided "a draft is never reviewed" for our own keyed workflow, and
with no key the cost reason ("reviewing drafts spends key cost") is moot, so that earlier leader ruling is
replaced only insofar as it concerned the keyed workflow
Facts recorded honestly: whether the native review runs on a draft, and what it leaves at the
ready-for-review transition, are items to confirm by observation (DS-2, `design.md` §12); nothing in the SPEC
depends on the answer
Operator verdict: OPERATOR-SETTLED 2026-10-18 operator-confirmed contract and leader correction (review on
pull-request open; per-push excluded; draft is an item to confirm by observation), transcribed from the
correction message

### Q10: What text reaches the Codex prompt? — SUPERSEDED

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the question existed only for the key-based action's prompt; Q1 rejected that action, so no
prompt of ours exists
Default: not applied (superseded by Q1)
Alternate: none remaining
Operator verdict: OPERATOR-SETTLED 2026-10-18 superseded by the operator-confirmed native-review contract
(Q1)

### Q11: What happens when a pull request is too large to review? — SUPERSEDED

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the question concerned the key-based action's size handling. With the native review the gate
reads only what the app leaves: a review that never arrives is NOT RUN `review-absent`, one bound elsewhere is
`stale-review`, and nothing is chunked or heuristic; CodeRabbit's skip also left the merge withheld (research
U1)
Default: not applied (no size heuristic and no chunking; superseded by Q1 and Q9)
Alternate: none remaining
Operator verdict: OPERATOR-SETTLED 2026-10-18 superseded by the operator-confirmed native-review contract
(Q1)

### Q12: Where does the gate logic live — a script, or inline workflow steps?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the existing auto-merge test extracts an inline step by name; `scripts/ci/` and
`scripts/release.sh` hold script precedent
Default: one script with three subcommands, run from the default-branch checkout (rule: smaller surface — one
implementation instead of inline copies, and a default-branch copy the pull request cannot edit for itself)
Alternate: inline steps extracted by the tests
Operator verdict: DEFAULT-APPLIED 2026-10-18T11:53:23Z manager-spec (lane-22, card t1868)

### Q13: Is the guard-liveness manifest in scope?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: it is not on the card's original file list, but `.moai/guards/liveness.yaml` names `claude.yml`
(line 48) and `review-quality-gate.yml` (line 146); deleting those workflows leaves entries for subjects that
can no longer fire, which the liveness read reports as stale (research E12)
Default: include it, as a consequence of the deletions (rule: preserves current behavior of the liveness read —
no false stale)
Alternate: leave it and open a follow-up card
Revision 3: the alternate is not available. The existing `TestCensus_SetDifferenceEmptyBothDirections`
(`internal/guardstate/census_test.go:49-91`) turns red on a half-done manifest (research E12 correction, E19),
so the removal is forced by the deletions. This SPEC adds no workflow, so it adds no manifest entry; the gate's
continued-firing signal is the existing `auto-merge.yml` entry
Operator verdict: DEFAULT-APPLIED 2026-10-18T11:53:23Z manager-spec (lane-22, card t1868); in scope by the
leader ruling (b) of 2026-10-18

### Q14: Does this card change the doctrine, template, projection, and session-start surfaces that describe the CodeRabbit read?

Label: FOUNDER
Class: implementation-level (settled as "not in this card": no template default and no always-loaded surface
changes here; revision 1 classed the then-planned rewrite product-level)
Authority anchor: none in the committed register — the settlement is the leader's ruling message of 2026-10-18,
transcribed from the revision brief (a message, not a committed artifact)
Why unresolved: the card scope says "related docs"; 25 tracked files carry the `Merge Risk` predicate
(research E13) — 3 are frozen test fixtures kept as is, 1 is `auto-merge.yml`, and 21 are doctrine,
role-prompt, template, projection, hook, test, and ledger files. Revision 1 planned to rewrite all 21 (about
30 files with mirrors, projections, a ledger fixture, and a Go string table) in five doctrine milestones,
which changes the always-loaded surface and the template defaults.
Options (the leader's wording): (a) include M6-M10; (b) shrink; (c) exclude templates and AGENTS.md from M6-M10.
Decision (recorded): the leader chose (b), shrink. The audited scope is the workflows under
`.github/workflows`, the review-gate script and its tests, `.coderabbit.yaml`, the `required-checks.yml` guard
(the file is never edited), and the liveness manifest lines with the tests that read the manifest. The
doctrine, role-core, companion, `AGENTS.md`, `manager-lead`, `tech.md`, ledger, hook-test, lane-notice, and
runbook work is out of this card; it is recorded in the Follow-up candidates table below and in known debt
KD-1 (`spec.md` §7).
Evidence: (1) the leader's ruling of 2026-10-18, option (b); (2) the operator's boundary, "related docs", not
role rules — from the card text and the leader's boundary list; (3) the measured consequence of leaving the
doctrine as is (research E20, `spec.md` §7 KD-1), which the leader accepted.
Rejected option (a), include M6-M10: it changes the always-loaded surface and the template defaults, which the
operator's "related docs" boundary did not name.
Rejected option (c), exclude templates and AGENTS.md from M6-M10: it still edits the deployed role rules the
ruling leaves alone, and the rule-template mirror test requires the deployed rule copies and their template
copies to be byte-identical unless a pair is declared an intentional fork (`internal/template/rule_template_mirror_test.go`
header, lines 4 and 39-52), so leaving the templates out means declaring a fork for each deployed rule, a
larger change than the rewrite it avoids.
Consequence stated, not hidden: until the follow-up lands the deployed doctrine still describes CodeRabbit as
the review source and existing doctrine tests still assert that text (known debt KD-1); no requirement or
criterion of this SPEC claims the doctrine is updated or depends on it.
Operator verdict: OPERATOR-SETTLED 2026-10-18 leader ruling (b), transcribed from the revision brief

### Q15: How is a draft pull request's state shown?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: a draft gets no merge and may or may not get a native review (research U4); "no result" could
read as a pass, as a failure, or as nothing, and the consumer must tell the three apart
Default: not applied (settled by the operator direction)
Resolution (revised in revision 6): the gate reads `draft` from the pull request object and exits 11 at once,
whatever reviews exist; the consumer records the single state `not-run:draft`, ends the `review` step with exit
11 so the job result is never success, and withholds the merge; a non-draft pull request with no native review
at the end of the wait window is the same gate exit 11 with cause `review-absent`, which the consumer reports
as a failure (step exit 1, `state=fail`); the draft report says it verifies the gate path only and is not a
review result; the gate's pass is 0, fail is 1, pending is 10, not run is 11 (design §3.2, §5 `T-CHAIN`). The keyless notice job of revisions 1 and 2 is gone with the keyed workflow
Alternate: ignore the draft flag and read reviews — rejected: a draft would be merged on a stale or partial
review, and a draft observation could be mistaken for a review result
Operator verdict: OPERATOR-SETTLED 2026-10-18 operator direction relayed by the leader (not run is neither a
failure nor a pass; the consumer withholds on not run), transcribed from the revision addendum; narrowed by the
leader's ruling of 2026-10-18 (board record d-20261018T152011Z-8447, which supersedes d-20261018T151824Z-db98): a non-draft NOT RUN is a failure and a
draft's job result is not success

### Q21: How does a draft pull request's job result avoid reading as success?

Label: FOUNDER
Class: implementation-level
Authority anchor: none in the committed register; the constraint is the leader's ruling, board record
d-20261018T152011Z-8447 (supersedes d-20261018T151824Z-db98; ruling and scope unchanged), supplemented by
d-20261018T155527Z-4e9e (a delegated Codex observer run, not an operator run, not a new approval). Paraphrase of
what those records say: a draft keeps its own state `not-run:draft` and is never success or a review PASS;
a non-draft NOT RUN (review-absent, review-unreadable) is failed, not success; a failure in either step this
card added ends that step and the job as not-success. The records name no mechanism for the draft's job result.
Why unresolved: the records fix the outcome (never success) and leave the mechanism open. The workflow_run payload carries no `draft` field, so a job-level skip needs a pre-job that
reads the pull request object and a `needs` condition on the job
Default: the `review` step ends a draft with exit 11, so the job result is failure and the state is
`not-run:draft`; the workflow stays one job
(rule: preserves current behavior, the workflow's one-job topology)
Alternate: a pre-job reads `draft` and the main job's `if` skips it for a draft, so the job result is skipped;
rejected as the default because it adds a job, an output plumbing step, a `needs` grammar for the tests'
`if:` evaluator, and a second place that reads the draft flag. Its visible gain is a grey skipped job instead
of a red one on every draft; the leader may choose it by a one-line ruling (spec §6 names it out of scope). The red-job cost is not
limited to drafts: every refused or unreviewed non-draft pull request (`T-CHAIN` rows 2 to 4 and 6 to 9) and every `files` failure (rows 11 to 13) also shows a red Auto Merge job, which the ruling accepts
Operator verdict: DEFAULT-APPLIED 2026-10-18T16:00:33Z manager-spec (card t1868, lane-22, plan revision 6)

### Q22: Do the earlier steps `pr`, `head`, `checks`, and the merge step also end non-zero when they withhold?

**[SUPERSEDED IN PART by Q30, 2026-10-23]** For the merge step the default below no longer holds: a merge the step reached and did not
execute ends non-zero (REQ-CCR-024). For `pr`, `head`, `checks` (and `author`, Q26) the default stands, and each of their holds is now
also a `pending` commit status `auto-merge` (REQ-CCR-025). The row is kept as written.

Label: FOUNDER
Class: implementation-level
Authority anchor: none in the committed register; the scope line is the leader's ruling, board record
d-20261018T155527Z-4e9e ("the earlier checks/head steps stay out of scope")
Why unresolved: the rulings cover the two steps this SPEC added (`files`, `review`) and the chain's conclusion;
they leave `pr`, `head`, `checks`, and the merge step as they are, so a head that moved after CI, a failed
required-checks lookup, or a passed deadline still ends a green job
Default: unchanged, exit 0 with an annotation and an output
(rule: preserves current behavior; the pre-existing convention)
Alternate: extend the non-zero rule to those steps; a separate ruling, because it changes which runs of the
existing, pre-SPEC steps read red
Operator verdict: DEFAULT-APPLIED 2026-10-18T16:00:33Z manager-spec (card t1868, lane-22, plan revision 6)

### Q16: How strongly is the Codex identity bound?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: a review's author can be matched by login, by numeric id, by account type, or by all three; the
forged-artifact risk of audit defect D1 is its successor here — a comment or review by any user, by another app,
or by a lookalike
Default: all three together — login `chatgpt-codex-connector[bot]`, user id 199175422, type Bot (rule: excluded
before ranking: login-only and id-only matching, each of which a lookalike fixture defeats (AC-CCR-003); among
the rest the three-part match is the only one that survives them, and the values were measured, research E27)
Alternate: login only — rejected: a renamed or lookalike account passes
Operator verdict: DEFAULT-APPLIED 2026-10-18T11:53:23Z manager-spec (lane-22, card t1868)

### Q17: Are paged lists read with an explicit page loop or with `gh api --paginate`?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the files listing returns 30 entries per page by default (research E22), so one call misses a
late page (audit defect D2); both forms exist
Default: an explicit loop in the gate script — `per_page=100`, page by page, a short page ends it, a full 30th
page refuses, a page error refuses, a count proof where the pull request object carries one (rule: the
`--paginate` form carries unverified premises — its behavior on a mid-pagination error and its shape in a
pipeline — which a stub `gh` in a test cannot reproduce; the loop has none)
Alternate: `gh api --paginate` to a file with an exit-status check
Operator verdict: DEFAULT-APPLIED 2026-10-18T11:53:23Z manager-spec (lane-22, card t1868)

### Q18: How wide is the tamper guard?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: a pull request can change the gate, or add a workflow that merges itself or acts with the
workflow token; the guard can name only the gate files or every workflow file
Default: any added, changed, deleted, or renamed file under `.github/workflows/`, and the gate script (rule:
the narrow form leaves a path where a pull request-added workflow acts for itself; undo is a single revert in
both options; the wide form is the smaller trust surface, and human merges are unaffected)
Alternate: only `auto-merge.yml` and the script — rejected: audit defect D1 showed a pull request can add an
unrelated workflow
Operator verdict: DEFAULT-APPLIED 2026-10-18T11:53:23Z manager-spec (lane-22, card t1868)

### Q19: What does the gate do when a recognised review is clean but its artifact form cannot be confirmed?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the native review's form is not observed (research E29, U2, U5); the gate must choose between
recognising several guessed forms and recognising one documented form
Default (superseded in revision 7, see the restatement below): recognise exactly one form — a pull-request review by the Codex identity bound by `commit_id`, counted
clean only in state APPROVED (Q20) — and fail closed on everything else; DS-1 observes the real form and any
mismatch becomes a follow-up (rule:
excluded before ranking: parsing issue-comment or review-body text, which has no documented commit binding and
would be pass-by-guess; among the rest, one form is the smaller surface)
Alternate: also accept an issue comment of the bot that names the head SHA — rejected until DS-1 shows that form
exists and what it names
Operator verdict: DEFAULT-APPLIED 2026-10-18T11:53:23Z manager-spec (lane-22, card t1868)
Revision 7 (restated): DS-1 showed the form on five real pull requests (E37 to E39). The gate recognises exactly one carrier, the state comment of the identity, and
fails closed on everything else, including a `status` value, a marker version, or a field the reader does not know (U14).

### Q20: Which state of the native review counts as clean until DS-1 shows what a clean native review looks like, and what supersedes a change request?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: plan audit iteration 2 found (D12) that a COMMENTED review bound to the head with no inline
comment read as clean whatever its body said — a body reporting findings, "could not review", or a
usage-limit notice all passed — and (D13) that a CHANGES_REQUESTED review followed by a COMMENTED review
passed, against REQ-CCR-001. The native review's clean form is unobserved (research E29, E32, U5), so the
reader cannot tell a clean COMMENTED review from a failed one without reading body text, whose shape is also
unobserved
Options: (a) APPROVED is the only clean state, the newest bound review decides, a CHANGES_REQUESTED review is
superseded only by a later APPROVED; (b) COMMENTED or APPROVED with no finding, newest decides (the revision-3
reader); (c) any bound CHANGES_REQUESTED is final, with no supersession
Default: (a) (rule: excluded before ranking: (b), which fails open on a body-only finding, a failure notice,
or a usage-limit notice, and lets a later COMMENTED mask a CHANGES_REQUESTED — it violates the fail-closed
direction; among (a) and (c) the rule's first step is inapplicable (no current behavior exists for this
reader); the second step ties (undo is a single revert of this SPEC's commits in both); the third picks the
smaller user-visible surface: (c) would leave a pull request permanently withheld after a legitimate
re-approval of the same head, (a) lets a later APPROVED of the same head clear it)
Alternate: (c), the strict reading — a CHANGES_REQUESTED review blocks the head whatever follows; simpler, and
withholds more, but a re-approval of the same head can never clear it
Consequence stated, not hidden: if the native review never submits an APPROVED review (its clean result is a
COMMENTED review, a comment, or a check run), the gate withholds every automated merge until follow-up FU-10
adapts the recognition contract after DS-1; human merges are unaffected. Body text is never read in either
direction, so an APPROVED review whose body says otherwise passes (design §6, residual 1)
Operator verdict: DEFAULT-APPLIED 2026-10-18T12:54:54Z manager-spec (lane-22, card t1868)
Revision 7: CLOSED, superseded by the comment contract (REQ-CCR-001, REQ-CCR-004). DS-1 showed what a clean native review looks like (research E37): a state comment, not a
review object; options (a), (b), and (c) all concern review objects and no longer apply. The consequence stated above (the gate withholds until FU-10 adapts the
contract) is discharged by this amendment.

### Q23: How do review objects and review comments of the identity count now that the state comment decides?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the leader's ruling says a COMMENTED review or an inline comment "may only corroborate findings, never grant a pass"; it does not say whether the
reviews listing is still read, or which review states refuse. A real review object exists only beside findings (#1836, #1833: COMMENTED, with an inline comment on #1836)
Default: keep the review-comment and review listings as refusal-only inputs: a review comment of the identity bound to the head, or a COMMENTED or CHANGES_REQUESTED review
of the identity bound to the head, refuses (`finding`, `finding-review`); an APPROVED, DISMISSED, or PENDING review neither passes nor refuses; nothing here ever grants a pass
(rule: preserves current behavior — the shipped reader already read both listings and refused on review comments and CHANGES_REQUESTED; undo is a single revert in every option)
Alternate: drop the reviews listing and read only review comments — a smaller surface, one endpoint fewer, but a COMMENTED review bound to the head with no inline comment would no longer corroborate
Alternate: read neither listing and trust the marker alone — the smallest surface; rejected as the default because it removes the only independent check on a marker that reads clean while a finding exists on the head
Operator verdict: DEFAULT-APPLIED 2026-10-19T05:46:35Z manager-spec (card t1868, lane-22, plan revision 7)

### Q24: Is the `codex-security-review:v1` marker read?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the first draft of revision 7 left it unread; the sample of #1842 (research E42, E44) shows the state marker's security leg reading `completed` (with null fields) while the
security marker reads `running`, and after a re-request the security marker still named the OLD head
Default: read it for `status` and `headSha` only (order, revision 8: a security marker that is absent while a leg is `running` waits, any other missing, malformed, duplicated, or misplaced one refuses; a running comment without the marker was never observed); it must stand once in the leading block, read `completed`, and carry the head in `headSha`; `running` is pending, any other disagreement refuses, a
headSha of another commit is stale (excluded before ranking: leaving it unread, which passes an unfinished security review; among the rest, two fields are the smaller surface)
Alternate: also compare `pullRequestNumber` and `repository` with the pull request — rejected here: no mismatch was ever observed and a copied comment is already excluded by the author binding
Operator verdict: DEFAULT-APPLIED 2026-10-19T05:46:35Z manager-spec (card t1868, lane-22, plan revision 7)

### Q25: Does a moving issue-comment count fail a pass?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the existing rule fails a pass when the review-comment count moves during the read; issue comments are written constantly by other bots (3 or 4 per measured pull request, research E37)
Default: a moved issue-comment count between the two reads of one review pass retries the pass (pending note `issue-comments-moved`). At the pre-merge recheck, R3 supplies a fresh valid issue-comment count; the complete paginated issue-comment list must match it, and R4 must retain that same count and the verified head. The gate then re-evaluates trusted artifacts and must return the identical clean summary. A count increase between the earlier clean pass and R3 alone is allowed because ordinary comments do not grant or refuse a Codex verdict. A count/list mismatch, movement between R3 and R4, or any change to trusted state remains a failure.
(rule: keep list completeness and the final observation window bound while ensuring unrelated comments do not block a clean review; a second or changed trusted state comment is still rejected by the gate and summary comparison)
Alternate: compare R3 with the earlier clean-pass issue-comment count — rejected because an unrelated external comment then blocks auto-merge before the trusted artifacts are re-read, as reproduced in finding #4316515155.
Operator verdict: APPLIED 2026-10-20, explicit follow-up to exact-head finding #4316515155; supersedes the 2026-10-19 pre-merge count-equality default. The within-pass retry behavior remains unchanged.

### Q26: What does the `author` step do to the job, and how is an ineligible pull request shown?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: REQ-CCR-020 fixes who may be auto-merged (Q28) and leaves open how the step ends for everyone else
Default: the step ends with exit 0, records `eligible=false` once, and annotates with a notice that the pull request needs the user's test confirmation and explicit approval; the later steps are skipped through their `if`, and the job stays green
(rule: preserves current behavior — the steps `pr`, `head`, `checks` withhold with exit 0, Q22; an ineligible pull request is a routing decision, not a gate failure)
Alternate: end the step non-zero — every dependabot, fork, or other-author pull request would show a red `Auto Merge` job, which reads as a failed gate when nothing failed
Operator verdict: DEFAULT-APPLIED 2026-10-19T05:46:35Z manager-spec (card t1868, lane-22, plan revision 7)

### Q27: What does the gate do on a base without branch protection, given the shipped `checks` step?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the leader's addition says the gate applies on every base (an unprotected `release/*` base included); the coordinator's audit records `gh pr checks --required` exiting 1 with `no required checks reported` on such a base (research E43) and the shipped `checks` step reads that as a lookup failure and withholds (E45), so no pull request into such a base is auto-merged today
Default: leave the `checks` step unchanged; the review gate, the `author` step, and the merge path apply to every base whenever the `checks` step passes (rule: preserves current behavior; Q22 keeps the earlier steps as they are)
Alternate: change the `checks` step to treat `no required checks reported` on a base with no protection as an empty requirement and rely on the triggering CI conclusion plus the review gate — a change to which merges the automation makes; a product-level decision for the leader, not taken here
Operator verdict: DEFAULT-APPLIED 2026-10-19T05:46:35Z manager-spec (card t1868, lane-22, plan revision 7)

### Q28: Which pull requests does auto-merge merge, and does it delete branches?

Label: FOUNDER
Class: product-level (changes a shipped command's default behavior: bot and other-author pull requests stop being merged by automation, and the merge no longer deletes the head branch)
Authority anchor: none in the committed register — the settlement is the leader's relayed current policy of 2026-10-19 (a message, not a committed artifact)
Why unresolved: the shipped workflow merges any author's pull request and passes `--delete-branch` (research E45)
Default: not applied (settled by the relayed policy)
Resolution: only a pull request authored by GoosLab (id 6458969, type User, all three), from a branch of the base repository, and not by a bot, is merged by automation, and only through every other gate; every other pull request needs the user's test confirmation and explicit approval and is never auto-merged; automatic branch deletion is forbidden (no `--delete-branch`, no ref-deleting step); the repository setting `delete_branch_on_merge` stays an operator decision (gate G9)
Alternate: keep merging bot pull requests once they have a clean native review (the closed text of design §11) — withdrawn by the policy
Operator verdict: OPERATOR-SETTLED 2026-10-19 leader's relayed current policy (author, fork, bot, and branch-deletion rules), transcribed from the leader's message

### Q29: How is a leg whose `status` is `failed` classed?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: PR #1844 (research E46) shows a code leg `failed` with both finding flags true and the head bound; the generic refusal `review-not-complete:<leg>:failed` would be reached only after the flags were read as findings
Default: a distinct refusal `review-failed:<leg>`, a gate refusal (step exit 1, job failure), checked right after the stale check and before any finding flag is read, also while the other leg is `running` and also when `commit_sha` is null; the annotation names the re-request (`@codex review` for `code_review`, `@codex security review` for `security_review`) (rule: excluded before ranking: reading the flags of a failed leg as findings, which reports a failed run as a finding, and treating `failed` as pending, which waits for a comment that cannot pass without a new run; among the rest, one reason and one row is the smaller surface)
Alternate: leave `failed` to the generic `review-not-complete:<leg>:<status>` — the same exit and job result, but the operator reads no re-request and a flags-first reader can misreport a finding
Operator verdict: DEFAULT-APPLIED 2026-10-19T06:19:40Z manager-spec (card t1868, lane-22, plan revision 9)

### Q30: Are audit findings F4 (a late review is never re-evaluated) and F3/F1 (merge-step holds end exit 0 and look green) accepted limitations or fixes?

Label: FOUNDER
Class: product-level (changes the shipped workflow's triggers, its permissions, the job result of a held merge, and adds a commit status every pull request shows)
Authority anchor: none in the committed register — the settlement is the operator and leader ruling of 2026-10-23, relayed by the leader in the delegation to manager-spec (a message, not a committed artifact)
Why unresolved: the first sync audit (`.moai/reports/t1868/sync-audit-1.md`, PASS-WITH-DEBT 0.91, local only and gitignored, never an authority anchor) found that a review finishing after the merge step's window, or after the pull request becomes ready, is never re-evaluated (the only trigger is a completed `CI` run; the SPEC had recorded this as accepted limitation AL-1), and that the merge step's holds end exit 0 (the SPEC had kept the merge step's exit-0 convention, Q22)
Default: not applied (settled by the ruling)
Resolution: both findings are FIXES on this card, not accepted limits; the accepted-limitation option (AL-1) is withdrawn. Approved: an `issue_comment` trigger (`created`, `edited`) admitted at the job level only for a pull request comment of user id 199175422 (`chatgpt-codex-connector[bot]`, type Bot), with no checkout or execution of pull-request code and the pull request number passed through `env`; the unchanged step chain bound to the API-read head, one concurrency group per pull request, one evaluation with no wait; non-zero merge-step outcomes when auto-merge was requested but not executed (deadline elapsed, head moved, `gh pr merge` failed); a commit status `auto-merge` (pending, failure, success) on the evaluated head on every evaluation, with the one new permission `statuses: write`; no branch-protection change, no required check, `.github/required-checks.yml` byte-identical to `origin/main`. A minimal SPEC amendment is approved with NO plan-audit re-entry; one head-bound sync re-audit at the final head checks this SPEC against the code (REQ-CCR-022 to 025, AC-CCR-006a, 013a, 013b, 013c, 019a)
Alternate: keep AL-1 and the merge step's exit-0 convention as accepted limitations — withdrawn by the ruling
Operator verdict: OPERATOR-SETTLED 2026-10-23 operator and leader ruling, transcribed from the leader's delegation message

### Q31: On the comment path, how is a state that is not clean (a `running` leg, no state comment) ended, what does "auto-merge requested" mean, and which commit status does each outcome get?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the Q30 ruling fixes one evaluation with no wait and the three status states (pending "not yet", failure "requested but not executed", success "merge requested"); it names no exit for a `running` state on that single pass, uses "auto-merge requested" without naming a field, and does not place the `files`/`review` refusals, the draft, or an ineligible author among the three states
Default: (a) the single pass is a last pass: `--wait-closed true`, so `running` is `review-not-complete:running` and an absent state comment is `not-run:review-absent`, both a failed `review` step per the unchanged `T-CHAIN`; the identity's next comment edit re-evaluates. (b) "Auto-merge requested" is the chain reaching the merge step with every guard passed; no GitHub `auto_merge` flag is required on either path. (c) `T-STATUS` of `design.md` "Delta of plan revision 13": `success` = merge requested; `failure` = merge step reached and not executed, or a `files`/`review` refusal other than `not-run:draft`; `pending` = draft, ineligible author, `checks` or `head` hold
(rule: preserves current behavior — `T-CHAIN` and the step conditions are unchanged, and a refusal that already fails the job keeps a failing signal)
Alternate: treat `running` on the comment path as a `pending` hold with step exit 0 — a change to `T-CHAIN` row 2g that the ruling did not make; or require GitHub's `auto_merge` flag on the comment path — a change to which pull requests the automation merges
Operator verdict: DEFAULT-APPLIED 2026-10-22T23:52:15Z manager-spec (card t1868, plan revision 13; no plan-audit re-entry by Q30, checked by the head-bound sync re-audit)
Lane ruling 2026-10-23 (revision 13.1), relayed by the coordinator: **default (a) SUPERSEDED — REVERSED** (the text above is kept as history): on the comment path a review that is not yet complete is a "not yet" hold, not a requested-but-not-executed outcome; the single pass runs with `--wait-closed false`, and a pending result (gate exit 10: a leg or the current-head security marker `running`, or no state comment yet) ends the `review` step with exit 0, state `pending`, the merge step skipped, the job not failed, and the `auto-merge` status `pending` with reason `review-not-complete:running` (or `review-not-complete:no-state-comment`); the bot's completion edit re-triggers a pass; only a merge step reached and not executed (head moved, window elapsed, `gh pr merge` failed) is non-zero. **Default (b) CONFIRMED:** "auto-merge requested" means the chain reached the merge step; the workflow reads no GitHub `auto_merge` flag and gates on the `author` (eligibility, same-repository) step and the CI, head, files, and review steps. **Default (c) stands**, with the pending row added to `T-STATUS`. Concurrency key specified (design item 4): the pull request number for the `workflow_run` path and the identity's comments, the run id for any other comment author.

## Follow-up candidates (recorded, not issued)

Revision 3 records these as candidates only; no card is issued by this plan. Each row names the files and the
reason. The reason for rows FU-1 to FU-8 is quoted from the leader's ruling: "always-loaded surface and
template defaults; the operator's boundary was related docs, not role rules". The doctrine predicate the
follow-up would state now names the native Codex review (design §9). **Transferred to card t1874 (revision
6):** the CodeRabbit rule replacement in `AGENTS.md` §4, the review-read contract, the role cores, and the
templates (rows FU-1 to FU-9 are its input) are card t1874's, blocked until t1868 is merged-local and DS-1 is
observed on the G6 draft pull request. This SPEC does not edit them.

| Id | Files (deployed, template mirror, and projection where each exists) | Reason |
|----|-----------------------------------------------------------------------|--------|
| FU-1 | Role-core rules: `.claude/rules/moai/workflow/factory-dispatch.md`, `factory-dispatch-core.md`, their `internal/template/templates/` mirrors, and the `.moai/policies/workflow/` projections | always-loaded surface and template defaults; the operator's boundary was related docs, not role rules |
| FU-2 | Companion doctrine: `factory-dispatch-detail.md`, `factory-dispatch-mechanics.md`, `factory-dispatch-gates.md` (deployed, template, and `.moai/policies/workflow/` copies), and the `kanban-dispatch*.md` projections | always-loaded surface and template defaults; the operator's boundary was related docs, not role rules |
| FU-3 | `AGENTS.md` and `internal/template/templates/AGENTS.md.tmpl` | always-loaded surface and template defaults; the operator's boundary was related docs, not role rules |
| FU-4 | `.claude/agents/moai/manager-lead.md`, its template mirror, and the template `.codex/agents/moai/manager-lead.toml` | always-loaded surface and template defaults; the operator's boundary was related docs, not role rules |
| FU-5 | `.moai/project/tech.md` (the "PR Reviewer" tooling row, line 157) | always-loaded surface and template defaults; the operator's boundary was related docs, not role rules |
| FU-6 | `internal/template/testdata/binding_ledger.json` (rows ALB-0284, ALB-0285 and two further rows) and the hook role-core test `internal/hook/role_core_elements_test.go` (rows at lines 51-52) | they pin the doctrine text, so they move with FU-1 to FU-4 and carry the same always-loaded and template-default reason |
| FU-7 | The four-locale lane notice `internal/hook/session_start_factory_i18n.go` (lines 140, 200, 260, 319) and any hook test added for it | always-loaded surface and template defaults; the operator's boundary was related docs, not role rules |
| FU-8 | Every other `Merge Risk` wording correction in the E13 census not listed above | always-loaded surface and template defaults; the operator's boundary was related docs, not role rules |
| FU-9 | The cutover runbook `.moai/docs/codex-review-cutover-runbook.md` and a test that it covers every gate (the former REQ-CCR-021) | not named by the ruling's audited-scope list and part of the out-of-card milestone M10; it changes neither the always-loaded surface nor a template default, so the leader may re-admit it into this card by a one-line ruling. The gates and deferred steps stay fully specified in `spec.md` §5 and the table below |
| FU-10 | CLOSED by revision 7: DS-1 showed the artifact the native review leaves (an issue comment with a state marker, research E37), and the recognition contract was rewritten to it (REQ-CCR-001 to 005). Nothing remains of the follow-up | answered by the amendment, not by a separate card |
| FU-11 | Widen the tamper guard to the CI-input files outside `.github/workflows/`: `.github/test-input-filters.yml` (read by the pinned paths-filter step, `ci.yml:76-79`; its `go_code` output gates the test jobs, `ci.yml:106`), `scripts/ci/retry-flaky.sh` and `scripts/ci-census/test-census.sh` (process the test results, `ci.yml:222-226`) | a pull request can reduce what the required checks verify without touching a gate path; a pre-existing CI-integrity limit outside the review gate (design §6, residual 5), and widening the guard is a separate decision that changes which pull requests auto-merge can merge |

Exit of the group: the census command of research E13 reports no `Merge Risk` hit outside the historical
allowlist (spec §7, KD-1 exit condition).

## Known debt

KD-1 — the deployed doctrine still describes CodeRabbit as the review source — **transferred to card t1874**
(blocked until t1868 is merged-local and DS-1 is observed on the G6 draft pull request). Full entry, the stale sentences
with file and line, the tests that pin them, the interim effect, and the exit condition (the follow-up card or
cards of the table above) are in `spec.md` §7 and `research.md` E20. The correction to the native Codex review
does not change KD-1; it changes only the predicate the follow-up would state (design §9).

## Deferred confirmation gates

Not decisions of this SPEC. Each is an operator-held step on the remote application (the keep-set category
"operator-held"); each leaves the review gate failing closed and none widens a gate. No requirement or
criterion depends on any gate's answer. New API keys, new app OAuth, persistent access, admin or security
settings, branch-protection changes, and cost are not approved; local Claude and GLM sessions are unchanged.

| Gate | Needs | Confirmer | Interim state | Reversal |
|------|-------|-----------|---------------|----------|
| G1 | WITHDRAWN: the key-based path is not built; no key, no secret, no cost | — | Not applicable | — |
| G2 | Change the native Codex review's settings or the app's permissions (per-push or draft review, repository selection, new OAuth, persistent access, credit) | Repository or organisation admin (operator) | Not changed, not retried by this card; the operator-confirmed state stays | Revert the setting |
| G3 | Add a Codex review check to main's branch-protection required contexts | Repository admin (operator) | Main keeps the five measured required checks (research E6) | Remove the context |
| G4 | Add a Codex review check to `.github/required-checks.yml` (and remove the guard test AC-CCR-017 in the same change) | Operator, with the leader | The file is byte-unchanged; AC-CCR-017 fails an unconfirmed edit | Revert the edit |
| G5 | Uninstall or disable the CodeRabbit app at repository or organisation level | Organisation owner (operator) | The repository config disables automatic review | Reinstall; flip the two flags |
| G6 | Apply this change remotely as a leader-coordinated draft pull request, including `git rm` of the remote-only `claude-code-review.yml` | Leader opens it; the operator or the leader's batch process merges it | Nothing is pushed by the lane; AC-CCR-014 turns red if the file survives the merge | Close the draft |
| G7 | Delete the repository variable `AUTOMERGE_MAX_MERGE_RISK` | Repository admin (operator) | Unreferenced after M3 | None needed |
| G8 | Delete the secret `CLAUDE_CODE_OAUTH_TOKEN` | Repository admin (operator) | Unreferenced after M4 | Recreate the secret |
| G9 | Decide the repository setting `delete_branch_on_merge` (relayed as `true`; not read by the lane): a security setting, an operator decision item; this SPEC records it and does not change it. While it is `true` GitHub deletes the head branch itself after a merge, so the workflow's removal of `--delete-branch` (REQ-CCR-021) does not stop deletion by itself | Repository admin (operator) | Not changed; the workflow requests no deletion | Revert the setting |

### Deferred observation steps (read-only on the remote; the lane never writes)

Not gates and not decisions: observations the leader coordinates after landing. No requirement or criterion
depends on any of them.

| Step | Observes | Coordinator | Status |
|------|----------|-------------|--------|
| DS-1 | The artifact form the native review leaves on a pull request | Leader | OBSERVED on #1835, #1836, #1837, #1838, #1841 and the samples of #1842, #1833, #1839, #1840 (research E37 to E46): one issue comment of the identity with the state marker; NOT observed on #1828 itself (a draft, silent) and not observed on a draft under a controlled read; the held fact is the DS-1 row of `spec.md` §5 |
| DS-2 | Whether a Codex review runs again after a push. Settings relayed from the leader's read-only observation: automatic review on for all pull requests, thorough review on, automatic security review on, minimum reported severity High (research E33) | Leader | Open observation; no automatic re-review on a push was observed (#1836 stayed at the older head; one sample), U6 |
| DS-3 | How a fresh review is requested for a head that moved after the review started: the leader's `@codex review` comment | Leader or operator, outside this lane | One re-request observed on #1842 (read at 05:36Z, research E44): the state comment was edited in place (only the code leg moved), a marker-less receipt comment was added, the security leg and marker stayed behind; U13. Never a deliverable |
