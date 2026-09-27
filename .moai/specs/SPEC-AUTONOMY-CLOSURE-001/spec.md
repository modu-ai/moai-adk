---
id: SPEC-AUTONOMY-CLOSURE-001
title: "Contract-based autonomy A4 — closure report, second-review record, human verdict, and stop before push (moai contract report)"
version: "0.3.2"
status: in-progress
created: 2026-09-26
updated: 2026-09-27
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/closure, internal/cli, internal/hook"
lifecycle: spec-anchored
tags: "autonomy, contract, closure-report, second-review, push-stop, human-verdict, autonomy-a4"
tier: L
depends_on: [SPEC-AUTONOMY-CONTRACT-001, SPEC-AUTONOMY-ESCALATION-001]
related_specs: [SPEC-AUDIT-PARTICIPANT-COUNT-001, SPEC-CODEX-AUDIT-GATE-AXES-001, SPEC-MCP-WORKTREE-ROOT-001]
---

# SPEC-AUTONOMY-CLOSURE-001 — Closure Report and Stop Before Push (Autonomy A4)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-26 | manager-spec | Initial plan-phase draft (card t1237, AUTONOMY-A4). A1 read at `67a2f55cb` (v0.5.0), A2 escalation record format at `8c9ee29b7` (v0.3.0 §I). Lead amendment 2026-09-26: the stop before push when the second review was not performed is owned here. |
| 0.2.0 | 2026-09-26 | manager-spec | Plan-audit iteration 1 (FAIL 0.77) repairs D1-D18. A1 re-pinned to `65e0a9167` (v0.5.1): v0.5.1 receipt fields and `outcome` (D2), card taken from the signed contract `card` field (D3, D4). Push range taken from the pushed source ref with a fail-closed classifier (D5); one evidence-currency rule shared by the second review and the closure report (D6, D16); second-review record binds its reviewed scope (D7); command invariants never render as passed without evidence (D8); one card evidence home for writers and readers (D9); acceptance criteria cut to 25 (D10); plan-audit discovery covers both report streams (D11); residual-risk and guided-mode wording fixed (D13, D18); `audit_multi` card failure paths defined (D17); template markers named (D14). Lead decisions 2026-09-26 on OQ-1 and OQ-2 folded in (D1). |
| 0.3.0 | 2026-09-26 | manager-spec | Plan-audit iteration 2 (FAIL 0.84) repairs D19-D25: the push set no longer excludes terminal SPECs, so a card synced to `completed` before merge is still evaluated (D19); `--all` and `--mirror` are undetermined under contract mode everywhere (D20); `push_check_undetermined` exits 1 in `push-check` (D21); REQ-016 wording (D22); the second-review scope records the base the review backend resolved (D23); AC-010 states which receipt fixtures are written after signing (D24); A1 tip re-read (D25). |
| 0.3.1 | 2026-09-26 | manager-spec | Plan-audit iteration 3 (FAIL 0.87) blocker D26: push-set membership is own-card — a contract is a candidate only when a range commit changes its own SPEC directory, or it is non-terminal in the source commit and a range commit changes its governed paths; a closed card on the remote is not re-evaluated unless a range commit edits that closed SPEC's own directory. Optional D27 (scope base wording) and D28 (A1 tip row) folded in. 운영자 승인 4차 예외(D26 한정): the operator approved a one-time fourth plan-audit iteration limited to D26. |
| 0.3.2 | 2026-09-27 | manager-spec | Plan-audit iteration 4 (FAIL 0.87, binding, lead-side Opus) blocker D31 repaired via lead disposition (b): the membership rule is unchanged; the re-admission of a closed card by another card's SPEC-directory edit is now a named accepted residual in spec.md §H with its resolution procedure, plan.md R3 is conditioned accordingly, and AC-015 gains one pinning case. Optional D29 (research.md moving-ref tip sentence removed; re-measured and pinned at `2d7987c74`; `65e0a9167` stays the binding pin) and D33 (duplicate 0.3.1 HISTORY rows merged into one) folded in. D30 and D32 remain optional debt; delta-audit PASS 0.90; optional N1/N2 wording folded in. |

## §A. User Story

As the operator of a MoAI project running SPECs under a signed contract, I want one page per card
that tells me — from evidence files, not from an agent's word — whether every acceptance criterion
has evidence, whether any invariant or ownership boundary was crossed, what new API surface appeared,
what the implementing agent concluded, what a second, different model concluded, and what was not
checked at all, so that I can accept, reject, or send the contract back for amendment without
reading the diff; and I want a push of the integration branch to stop mechanically when the second
review my contract requires was not performed or failed.

A4 is the closing step of the contract-autonomy epic. A1 defines the contract and its signature; A2
detects contract departures and writes escalation records; A3 rewires the gates. A4 reads all of
those and adds three things of its own: the **closure report**, the **second-review record**, and
the **stop before push**.

## §B. Scope

In scope:

- `moai contract report <card-id>`: a one-page closure report rendered as Markdown plus a JSON twin,
  derived only from files and git.
- The second-review record: `audit_multi` gains an optional card argument; when given, the server
  appends a record binding the review to the card, the audited commit, the reviewed scope, and the
  signed contract digest.
- `moai contract verdict <card-id> <accept|reject|amend-contract>`: the human verdict, recorded
  only on a human path.
- Push readiness: a PreToolUse guard on pushes of the integration branch under contract mode, and
  `moai contract push-check` exposing the same evaluation with stable reason codes and exit codes.
- Plan-audit verdict binding to a report file, rendered in the closure report (A1 Forward Note —
  Verdict Binding).
- Auditor instruction text so that a contract-mode second review passes the card argument and runs
  after the last commit that changes the card's governed paths.

### Out of Scope — Contract schema, signing, and revocation (A1, A3)

- The `contract.yaml` schema (including its required `card` field), `moai contract sign|show|verify`,
  the kickoff receipt format and its validator are A1. A4 reads them; it adds no contract field.
- `moai contract revoke`, moai-issued kickoff receipts, the cross-check outcome rules, and the
  moai-owned signing-event store are A3.

### Out of Scope — Escalation detection and push serialization (A2, A2b)

- Detecting the six contract classes and the four operational trips, and the escalation record
  format, are A2. A4 lists records; it never writes, resolves, or deduplicates one.
- The `moai slot` lease on `push-develop` and the deny on agent-invoked `moai contract sign` are A2b.
  A4's push readiness is a separate predicate evaluated at the same point; it does not replace the
  lease.

### Out of Scope — Gate rewiring (A3)

- Making a signature replace Implementation Kickoff Approval, turning escalation trips into denials,
  and denying a push whose contract lacks `push-develop` are A3. A4 stops only on its own closed set
  of readiness reasons.

### Out of Scope — Integration and queue mutation

- A4 does not merge, does not mark a card done, and does not change any queue item. Integration and
  `done` stay the lead's acts.
- A4 does not write the lead's hand-authored `.moai/reports/<card-id>/verdict.md`.

### Out of Scope — Non-agent pushes and other shells

- A push typed by a human in a terminal, and a push issued through a tool other than the Bash tool,
  are not observed by the PreToolUse guard. `moai contract push-check` is available for manual use;
  wiring it into the distributed git `pre-push` hook is not part of this SPEC.

### Out of Scope — Documentation surfaces

- docs-site pages and README text for the new commands are sync-phase work for manager-docs.

## §C. Context and Decisions

### §C.1 Review layers (lead design)

1. **First verdict** — the implementing agent closes with evidence (`progress.md` §E.2 AC matrix and
   §E.3 audit-ready signal).
2. **Second verdict** — a different model reviews the same contract through `audit_multi` (codex or
   GLM backend). A Claude backend entry never counts as the second verdict.
3. **Human verdict** — the operator reviews the closure report, not the diff, and records one of
   `accept`, `reject`, or `amend-contract`. In autonomous mode the human review is post-hoc: the
   absence of a verdict does not hold a push, and reversal runs through `reject` or revocation.

### §C.2 Operator decision A-Q3, the lead amendment, and the lead decisions on push blocking

A second-model review is required by default (`workflow.autonomy.contract.second_review: required`,
A1 configuration). When it has not been performed, the run stops before push; the lead amendment of
2026-09-26 assigns that stop to this card. Where both codex and GLM are unavailable, the second
review is recorded as not performed; it is never silently skipped.

Lead decisions of 2026-09-26 (plan.md §H, OQ-1): under `required`, a second review that was
performed and returned `fail` also blocks the push, under its own reason code; a latest recorded
human verdict of `reject` or `amend-contract` blocks the push; no recorded human verdict does not.

### §C.3 `push-develop` activation

A1 ships `push_develop: false` as the template default and states that the `push-develop` action is
not activated until A4 has landed. A4 satisfies that condition by landing the push readiness guard;
from that landing on, every observed push of the integration branch under contract mode is
evaluated. A4 adds no configuration key that disables the guard; the only switches are
`workflow.autonomy.mode` and `workflow.autonomy.contract.second_review`.

### §C.4 File names, and why not `verdict.md`

The lead design names `.moai/reports/<card>/verdict.md` as the closure report location. That name is
already the lead's hand-authored final verdict, the evidence path the kanban protocol fixes per card
(381 files measured, `research.md` §B.6). Lead decision 2026-09-26 (plan.md §H, OQ-2): A4 writes
`closure-report.md` and `closure-report.json` beside it and never touches `verdict.md`.

The kickoff receipt is read from the location A1 fixes, `.moai/specs/<SPEC-ID>/kickoff-receipt.json`.

### §C.5 Everything is derived; nothing is claimed

Every state the report renders, and every readiness decision, is computed by Go code from files and
git. No agent-supplied flag, message, or return value is an input. A missing input renders as
`not observed` or `not recorded` and is listed under Not Performed — never as agreement, never as a
pass, and never omitted.

### §C.6 Guided mode

The distributed template keeps `mode: guided`. Under `guided`, no existing command, hook output, or
MCP tool output changes, except the PreToolUse deny of `moai contract verdict` (REQ-CLOSURE-021),
which applies in every mode. The new commands (`report`, `verdict`, `push-check`) are additive and
available in every mode; the push guard is inert under `guided`.

### §C.7 Definitions used by the requirements

- **Card of a contract** — the value of the contract's signed `card` field (A1 v0.5.1). A4 never
  derives a contract's card from the queue store.
- **Card evidence home** of card `C` — the linked worktree whose directory base name equals `C`, or
  the primary checkout when no such worktree exists. **Card evidence directory** —
  `<card evidence home>/.moai/reports/C/`. Every A4 writer writes there and every A4 reader reads
  there, whatever tree the command runs from; A2's escalation records and plan-audit reports are read
  from the same home.
- **Governed paths** of a contract — the paths matched by its `ownership.write` globs, excluding
  every path under `.moai/specs/<SPEC-ID>/`. Rationale: that SPEC's `contract.yaml` and
  `acceptance.md` are immutable after signing (A1 effective `never`) and are covered separately by
  the contract digest; its other files (`progress.md` evidence, `spec.md` status and SHA backfill)
  are lifecycle records the sync commit writes after the review.
- **Currency rule** — evidence recorded at commit `R` is **current** for commit `P` when `R` is `P`
  or an ancestor of `P` and no non-merge commit in `R..P` changes a governed path. The report uses the
  card evidence home's HEAD as `P`; the push readiness evaluator uses the pushed source commit.

## §D. Requirements (GEARS)

### D.1 Closure report

- **REQ-CLOSURE-001** (Event-driven) — When `moai contract report <card-id>` runs, the report
  generator shall resolve the card's SPEC ID from the queue store, require that the SPEC's contract
  names the same card in its `card` field, build the report at the card evidence home's HEAD, write
  `closure-report.md` and `closure-report.json` atomically into the card evidence directory whatever
  tree the command runs from, print the Markdown path, and exit 0; when the card is unknown, the card
  has no SPEC ID, the SPEC has no `contract.yaml`, or the contract names a different card, it shall
  write nothing and exit 2 naming the cause.
- **REQ-CLOSURE-002** (Ubiquitous) — The closure report shall carry, in this order, the sections
  Summary, Kickoff, Contract Reconciliation, Invariants, Ownership, New APIs, Escalations, First
  Verdict, Second Verdict, Plan-Audit Binding, Not Performed, Residual Risk, and Human Verdict; the
  Markdown and JSON forms shall carry the same values; the JSON form shall carry a `schema_version`
  of 1; and two builds over the same files and HEAD shall produce byte-identical JSON apart from the
  `generated_at` field.
- **REQ-CLOSURE-003** (Ubiquitous) — The report generator shall derive every rendered state from
  files and git only, and shall render every input it could not read or measure as `not observed` or
  `not recorded` and list it under Not Performed, never as a pass and never by omitting the row.
- **REQ-CLOSURE-004** (Ubiquitous) — The Contract Reconciliation section shall list one row per live
  acceptance-criterion ID of the SPEC's `acceptance.md` (A1 counter semantics), carrying the status
  and evidence text reported for that ID in the `progress.md` §E.2 AC matrix verbatim, or
  `not reported` where no row exists; shall list any ID reported in `progress.md` but absent from
  `acceptance.md` as `unknown`; and shall show the contract's recorded acceptance hash and AC count
  beside the measured ones together with the A1 verify state and reason codes.
- **REQ-CLOSURE-005** (Ubiquitous) — The Invariants section shall list every contract invariant with
  its kind and one result: `violation recorded (open)` or `violation recorded (resolved)` when an
  `invariant-violation` escalation record names that invariant; for a `frozen-files` invariant
  without such a record, `no violation recorded` while the escalation detector was armed for the
  contract in force and not disarmed, and `not observed` otherwise; and `not observed` for every
  `constitution:` invariant and for every command-kind invariant without such a record, because no
  file records that a command invariant ran and passed.
- **REQ-CLOSURE-006** (Ubiquitous) — The Ownership section shall list every `ownership-move`
  escalation record of the card and shall state the ownership-violation count as a number only while
  the escalation detector was armed for the contract in force and not disarmed; otherwise it shall
  state the count as `not observed`.
- **REQ-CLOSURE-007** (Ubiquitous) — The New APIs section shall list every addition, by kind and
  path, that the escalation detector's new-architecture-or-api comparison observes between the card
  base and the report HEAD, running that comparison at report time without writing any escalation
  record, together with the additions named by the card's `new-architecture-or-api` records; where
  the comparison cannot run, the section shall state `not observed` rather than an empty list.
- **REQ-CLOSURE-008** (Ubiquitous) — The Escalations section shall list every escalation record of
  the card (kind, class, status, occurrences, contract reference, decider), shall state whether the
  card is needs-decision under the A2 rule, and shall list a record it cannot parse as `unreadable`
  under Not Performed.
- **REQ-CLOSURE-009** (Ubiquitous) — The First Verdict section shall show `run_status`,
  `ac_pass_count`, `ac_fail_count`, and `run_commit_sha` from the `progress.md` §E.3 block verbatim,
  or `not recorded` for each absent field, and shall flag a mismatch when the reported pass and fail
  counts do not sum to the measured AC count.
- **REQ-CLOSURE-010** (Ubiquitous) — The Kickoff section shall show the signature method and signer
  kind and, for a receipt signature, the receipt's `requested_decider`, `effective_decider`,
  `fallback.applied` and `fallback.reason`, the `llm_answer` and `jev_answer` each with answer and
  confidence, the recorded `outcome`, and the A1 verify status of the receipt. The section shall
  display any decider token verbatim, shall list any receipt field it does not recognize under Not
  Performed rather than drop it, and shall state `missing` where the signature method is `receipt` and
  the receipt file is absent. Receipt content shall be displayed only and shall not be an input to any
  readiness decision.
- **REQ-CLOSURE-011** (Ubiquitous) — The Plan-Audit Binding section shall search, in the card
  evidence home, first the file named by the kickoff receipt's `inputs.plan_audit_report`, then the
  highest-iteration `plan-audit*.md` in the card evidence directory, then the highest-numbered
  `.moai/reports/plan-audit/<SPEC-ID>-review-<n>.md`, and shall state `bound` when the first file found
  ends with an audit verdict line naming this SPEC and a verdict equal to the contract's
  `plan_audit.verdict`; `mismatch` when that line names a different verdict or the receipt-named
  file's hash differs from the receipt; and `self-reported` when no file or no verdict line exists.

### D.2 Second review

- **REQ-CLOSURE-012** (Event-driven) — When `audit_multi` is called with a `card_id` argument, the
  MCP server shall append one second-review record to the card evidence directory's
  `second-review.jsonl` carrying the card argument, the SPEC ID and contract card the queue store and
  contract yield (empty when either is absent), the signed contract digest (empty when absent or
  unsigned), the audited commit, the requested `target`, the reviewed scope (the base branch and base
  commit resolved by the same base-resolution function the review backends use, head commit,
  changed-file count, diff digest), each backend's name, gate, and verdict, the participant count, the
  disagreement flag, the audit receipt ID, the build commit, and the recording time; when the append
  fails, the returned result shall carry a non-empty `second_review_record_error`; and when the
  argument is absent, the tool's output and side effects shall be byte-identical to the pre-change
  behavior.
- **REQ-CLOSURE-013** (Ubiquitous) — The report generator and the push readiness evaluator shall
  select, from the card's second-review records, the most recently recorded one that passes every
  filter in this order — bound (card argument equals the contract card, SPEC ID and contract digest
  non-empty), same contract (digest equals the current signed digest), scope covered (`target` is
  `baseBranch`, the scope's head equals the audited commit, and at least one file changed), second
  model (at least one codex or GLM verdict of `pass` or `fail`), and in history (audited commit is the
  evaluation commit or its ancestor) — and shall report the second review as `performed` with verdict
  `fail` when any counted backend verdict is `fail` and `pass` otherwise, then apply the currency rule
  (§C.7) and report `stale` when it fails; when no record passes, the state shall be `not performed`
  with the cause named by the first filter that removed the last remaining record (`no-record`,
  `unbound`, `contract-changed`, `scope-not-covered`, `no-second-model`, `not-in-history`).
- **REQ-CLOSURE-014** (Ubiquitous) — The Second Verdict section shall show the second-review state,
  each counted backend and its verdict, the contract's `review.second_model`, whether the performing
  backend differs from it, and the effective `second_review` policy; it shall render the literal text
  `NOT PERFORMED` followed by the cause for a not-performed state, `STALE` followed by the superseding
  commit for a stale state, and `FAILED` for a performed review with verdict `fail`.

### D.3 Stop before push

- **REQ-CLOSURE-015** (Capability gate + event) — Where `workflow.autonomy.mode` is `contract`, when a
  Bash tool call contains a `git push` whose destination is or may be the integration branch, the
  PreToolUse hook shall take the push range from the remote-tracking integration ref to the pushed
  source commit, read candidate contracts from the source commit's tree, evaluate push readiness for
  every signed contract that lists `push-develop` and is a candidate on an own-card basis — a
  non-merge commit in that range changes a path under the contract's own `.moai/specs/<SPEC-ID>/`
  (whatever the SPEC's `status` in the source commit, because the sync commit sets `completed` before
  the card is merged), or the SPEC is non-terminal in the source commit and a non-merge commit in that
  range changes one of its governed paths — so that a contract terminal on the remote integration ref
  whose SPEC directory is unchanged in the range is not a candidate, and deny the call with a reason
  beginning `CLOSURE_PUSH_STOP:` followed by the sorted, de-duplicated readiness codes when any
  candidate is not ready.
- **REQ-CLOSURE-016** (Ubiquitous) — The push readiness evaluator shall report a contract not ready
  with every applicable code from exactly this set: `contract_invalid` (A1 verify is not
  `signed-valid`); `closure_report_missing` (no closure report JSON in the card evidence directory);
  `closure_report_stale` (the closure report's recorded HEAD fails the currency rule for the pushed
  source commit); `human_verdict_reject` and `human_verdict_amend_contract` (the latest recorded human
  verdict is `reject` or `amend-contract`, whether current or stale); `push_check_undetermined`
  (REQ-CLOSURE-017); and, where `second_review` is `required`, `second_review_not_performed`
  (REQ-CLOSURE-013 state `not performed`), `second_review_stale` (state `stale`), and
  `second_review_failed` (state `performed` with verdict `fail`). The absence of a recorded human
  verdict shall produce no readiness code.
- **REQ-CLOSURE-017** (Event-detected) — When, under contract mode, a `git push` destination cannot be
  proven different from the integration branch (a bare `git push` without a resolvable upstream,
  `--all`, `--mirror`, a refspec with an unresolvable source, a push inside `sh -c`, command
  substitution, or a variable), or when the push range, a contract's card evidence home, or its
  evidence files cannot be determined, or a git call times out, the push readiness evaluator shall
  report `push_check_undetermined` and the hook shall deny the push; it shall never allow a push on an
  undetermined result.
- **REQ-CLOSURE-018** (Capability gate) — Where `second_review` is `advisory` or `off`, the push
  readiness evaluator shall not report a second-review code; under `advisory` the report shall still
  render the second-review state with a warning line, and under `off` the Second Verdict section shall
  state `not required`.
- **REQ-CLOSURE-019** (Event-driven) — When `moai contract push-check [<remote> <refspec>...]` runs, it
  shall perform the same evaluation as the hook for the given push (or, without arguments, for a push
  of the local integration branch to its upstream) and exit 0 when every in-range contract is ready;
  exit 1 when any contract is not ready or the push is undetermined, printing each SPEC ID with its
  codes, or `push_check_undetermined` with its cause when no SPEC can be named; and exit 2 only on a
  usage error (unknown flag, malformed argument). Under `guided` it shall print one line stating that
  the check is inactive and exit 0.

### D.4 Human verdict

- **REQ-CLOSURE-020** (Event-driven) — When `moai contract verdict <card-id> <accept|reject|amend-contract>`
  runs, the verdict recorder shall refuse without writing (exit 1) when an agent-environment marker
  from A1's closed marker set is present, when standard input is not an interactive terminal, when the
  typed confirmation differs from the displayed token, or when the card evidence directory holds no
  closure report; otherwise it shall append one record to the card evidence directory's
  `closure-verdict.jsonl` carrying the verdict, an optional note, the operator name and email from git
  configuration, the UTC recording time, the SHA-256 of the current `closure-report.json`, and the
  method `interactive-tty`.
- **REQ-CLOSURE-021** (Unwanted) — The PreToolUse hook shall not allow a Bash tool call whose command
  text invokes `moai contract verdict`; it shall deny it in every mode with a reason beginning
  `CLOSURE_VERDICT_HUMAN_ONLY:`. The report generator, the push readiness evaluator, and the MCP
  server shall not write `closure-verdict.jsonl`.
- **REQ-CLOSURE-022** (Ubiquitous) — The Human Verdict section shall show the latest recorded verdict
  with its operator and time, marked `current` when its report hash equals the SHA-256 of the report
  being rendered and `stale` otherwise, or `none recorded`.

### D.5 Mode, evidence location, and template

- **REQ-CLOSURE-023** (Capability gate) — Where `workflow.autonomy.mode` is absent, empty,
  unrecognized, or `guided`, the PreToolUse hook shall run no push readiness evaluation and spawn no
  subprocess for it, so that its output for every tool call other than `moai contract verdict` is
  byte-identical to the pre-change behavior.
- **REQ-CLOSURE-024** (Ubiquitous) — Every A4 writer (report pair, second-review record, human verdict
  record) and every A4 reader (report generator, push readiness evaluator, verdict recorder) shall
  resolve the card evidence directory by the single rule of §C.7, and the report shall show the path
  of every evidence file it read.
- **REQ-CLOSURE-025** (Ubiquitous) — Template content added by this SPEC — auditor instructions,
  placed between the literal lines `<!-- moai:closure-second-review:start -->` and
  `<!-- moai:closure-second-review:end -->`, directing a contract-mode second review to pass the card
  argument with target `baseBranch` after the last commit that changes the card's governed paths —
  shall contain no SPEC ID, card identifier, date, operator-decision identifier, or commit SHA, and
  shall favor no programming language.

## §E. Non-Functional Constraints

- The push guard runs only under contract mode and only for a push whose destination is or may be the
  integration branch; its git work is bounded by a timeout, and a timeout is `push_check_undetermined`.
- Every file A4 writes is written atomically (report pair) or appended with a single write per line
  (`.jsonl` records).
- A4 imports A1's verification core for every contract fact and A2's record reader for every
  escalation fact; it re-implements neither schema.

## §F. Assumptions

- A1 (SPEC-AUTONOMY-CONTRACT-001) and A2 (SPEC-AUTONOMY-ESCALATION-001) land on the integration
  branch before this SPEC's run phase; A2b (card t1245) exposes a push-command classifier, or A4
  carries its own (design.md §C.1).
- `progress.md` keeps the §E.2 AC matrix row shape `| <AC-ID> | <status> | <evidence> |` and the §E.3
  fenced YAML block, as measured in this tree (research.md §B.5).
- Plan-audit reports keep the `AUDIT-VERDICT:` last-line format the audit receipt store parses.

## §G. References

- `design.md` — report layout, record shapes, readiness algorithm, push classifier, package layout.
- `research.md` — measured codebase facts and the dependency snapshot.
- `plan.md` — milestones, risks, resolved decisions. `acceptance.md` — acceptance criteria.

## §H. Residual Risk

- **Local files can be forged.** Anyone who can write the working tree can hand-write a
  second-review record, a closure report, or a verdict line. A4 ensures that a missing record stops
  the push and that forgery leaves the traces A1 accepts (git history; the audit receipt store for
  codex participation); it does not prove a record's author.
- **The verdict-command deny is a convenience guard.** It is a string match on the Bash command, and
  a wrapper (`sh -c`, variable indirection, a script file) evades it. The enforcement of "only a human
  records a verdict" is the terminal and agent-marker check of REQ-CLOSURE-020.
- **A record states what was requested.** `target: baseBranch` records the review the caller asked
  for, and the scope records the base the server resolved for it; a backend that interprets the target
  differently from that resolution is not detected by A4.
- **Only Bash-tool pushes are observed.** A terminal push, a push through another shell tool, or a
  push with git hooks disabled is not stopped by the PreToolUse guard (this repository sets
  `core.hooksPath` to `/dev/null`, research.md §B.8).
- **Currency is path-based, candidacy is own-card.** While a card is a candidate in a push, another
  card's commit in the range that changes one of its governed paths makes its review and report stale;
  this errs toward stopping, and the remedy is a fresh second review and report of that card. A closed
  card is not re-evaluated because a later card touches its paths — provided the touch is not an edit
  of that card's own SPEC directory: a code-only commit that changes a closed card's governed paths
  without touching its SPEC directory is not attributed to that contract, and an edit of that SPEC
  directory is (the accepted residual below).
- **Accepted residual — another card's edit of a closed card's SPEC directory re-admits it.** The
  SPEC-directory branch of REQ-CLOSURE-015 is path-only by design (it never asks which card authored
  the commit), so a non-merge range commit from another card that edits a closed card's own
  `.moai/specs/<SPEC-ID>/` directory (a supersession marker, a HISTORY line) makes that closed
  contract a candidate again; if its evidence is not current, the push stops with that contract's
  readiness codes — typically `closure_report_stale` with `second_review_stale`, or
  `closure_report_missing` where the card evidence home holds no report. Resolution: recreate the
  card-named worktree at the pushed source commit (the local integration branch head being pushed),
  regenerate the closure report, re-run the second review, then push again.
- **Class-4 comparison is a heuristic** (A2 C6); an empty New APIs list is evidence of no detected
  addition, not of no addition.
