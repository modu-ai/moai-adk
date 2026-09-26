---
id: SPEC-AUTONOMY-CLOSURE-001
title: "Contract-based autonomy A4 — closure report, second-review record, human verdict, and stop before push (moai contract report)"
version: "0.1.0"
status: draft
created: 2026-09-26
updated: 2026-09-26
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
| 0.1.0 | 2026-09-26 | manager-spec | Initial plan-phase draft (card t1237, AUTONOMY-A4). Consumes the A1 contract schema at the tip of branch `WT-contract-schema` (read at `67a2f55cb`, SPEC-AUTONOMY-CONTRACT-001 v0.5.0; v0.5.1 pending) and the A2 escalation record format at commit `8c9ee29b7` (SPEC-AUTONOMY-ESCALATION-001 v0.3.0 §I). Lead amendment 2026-09-26: the stop before push when the second review was not performed is owned here. Operator re-decision relayed by lead: kickoff decider is `llm` (default) or `llm+jev`; the Kickoff section is tolerant of the decider value set (REQ-CLOSURE-010). |

## §A. User Story

As the operator of a MoAI project running SPECs under a signed contract, I want one page per card
that tells me — from evidence files, not from an agent's word — whether every acceptance criterion
has evidence, whether any invariant or ownership boundary was crossed, what new API surface appeared,
what the implementing agent concluded, what a second, different model concluded, and what was not
checked at all, so that I can accept, reject, or send the contract back for amendment without
reading the diff; and I want a push of the integration branch to stop mechanically when the second
review my contract requires was not performed.

A4 is the closing step of the contract-autonomy epic. A1 defines the contract and its signature; A2
detects contract departures and writes escalation records; A3 rewires the gates. A4 reads all of
those and adds three things of its own: the **closure report**, the **second-review record**, and
the **stop before push**.

## §B. Scope

In scope:

- `moai contract report <card-id>`: a one-page closure report rendered as Markdown plus a JSON twin,
  derived only from files and git.
- The second-review record: `audit_multi` gains an optional card argument; when given, the server
  appends a record binding the review to the card, the audited HEAD, and the signed contract digest.
- `moai contract verdict <card-id> <accept|reject|amend-contract>`: the human verdict, recorded
  only on a human path.
- Push readiness: a PreToolUse guard on pushes of the integration branch under contract mode, and
  `moai contract push-check` exposing the same evaluation with stable reason codes and exit codes.
- Plan-audit verdict binding to a report file, rendered in the closure report (A1 Forward Note —
  Verdict Binding).
- Auditor instruction text so that a contract-mode second review passes the card argument.

### Out of Scope — Contract schema, signing, and revocation (A1, A3)

- The `contract.yaml` schema, `moai contract sign|show|verify`, the kickoff receipt format and its
  validator are A1. A4 reads them; it adds no contract field.
- `moai contract revoke`, moai-issued kickoff receipts, and the moai-owned signing-event store are A3.

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
   `accept`, `reject`, or `amend-contract`.

### §C.2 Operator decision A-Q3 and the lead amendment

A second-model review is required by default (`workflow.autonomy.contract.second_review: required`,
A1 configuration). When it has not been performed, the run stops before push. The lead amendment of
2026-09-26 assigns that stop to this card. Where both codex and GLM are unavailable, the second
review is recorded as not performed; it is never silently skipped.

### §C.3 `push-develop` activation

A1 ships `push_develop: false` as the template default and states that the `push-develop` action is
not activated until A4 has landed. A4 satisfies that condition by landing the push readiness guard;
from that landing on, every observed push of the integration branch under contract mode is
evaluated. A4 adds no configuration key that disables the guard; the only switches are
`workflow.autonomy.mode` and `workflow.autonomy.contract.second_review`.

### §C.4 Where A4's files live, and why not `verdict.md`

The lead design names `.moai/reports/<card>/verdict.md` as the closure report location. In this
tree that name is already the lead's hand-authored final verdict, the evidence path the kanban
protocol fixes per card (381 files measured, `research.md` §B.6). A generated report written there
would overwrite a human-authored record. A4 therefore writes `closure-report.md` and
`closure-report.json` beside it and never touches `verdict.md` (confirmation requested,
`plan.md` § Open Questions OQ-2).

The kickoff receipt is read from the location A1 fixes, `.moai/specs/<SPEC-ID>/kickoff-receipt.json`
(A1 design § Kickoff Receipt), not from `.moai/reports/<card>/`.

### §C.5 Everything is derived; nothing is claimed

Every state the report renders, and every readiness decision, is computed by Go code from files and
git. No agent-supplied flag, message, or return value is an input. A missing input renders as
`not observed` or `not recorded` and is listed under "Not performed" — never as agreement, never as
a pass, and never omitted.

### §C.6 Guided mode

The distributed template keeps `mode: guided`. Under `guided`, no existing command, hook output, or
MCP tool output changes. The new commands (`report`, `verdict`, `push-check`) are additive and
available in every mode; the push guard is inert under `guided`.

## §D. Requirements (GEARS)

### D.1 Closure report

- **REQ-CLOSURE-001** (Event-driven) — When `moai contract report <card-id>` runs, the report
  generator shall resolve the card's SPEC ID from the queue store, build the report for that SPEC's
  contract at the current HEAD, write `closure-report.md` and `closure-report.json` atomically under
  `.moai/reports/<card-id>/` of the current tree, print the Markdown path, and exit 0; when the card
  is unknown, the card has no SPEC ID, or the SPEC has no `contract.yaml`, it shall write nothing and
  exit 2 naming the missing input.
- **REQ-CLOSURE-002** (Ubiquitous) — The closure report shall carry, in this order, the sections
  Summary, Kickoff, Contract Reconciliation, Invariants, Ownership, New APIs, Escalations, First
  Verdict, Second Verdict, Plan-Audit Binding, Not Performed, Residual Risk, and Human Verdict; the
  Markdown and JSON forms shall carry the same values; and the JSON form shall carry a
  `schema_version` of 1.
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
  `invariant-violation` escalation record points at that invariant; `not observed` for a
  `constitution:` invariant, for an invariant an escalation record lists as not observed, and for any
  invariant while the escalation detector was not armed for the contract in force; otherwise
  `no violation recorded`.
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
  kind and, for a receipt signature, the configured decider as the receipt records it, every recorded
  answer with its decider, answer, and confidence (including a Jev attempt recorded outside the
  decision list), whether a fallback happened with its recorded reason, and the A1 verify status of
  the receipt. The section shall display any decider token and any number of answers verbatim, shall
  ignore receipt fields it does not know, and shall state `missing` where the signature method is
  `receipt` and the receipt file is absent. Receipt content shall be displayed only and shall not be
  an input to any readiness decision.
- **REQ-CLOSURE-011** (Ubiquitous) — The Plan-Audit Binding section shall state `bound` when a
  plan-audit report file — the one named by the kickoff receipt's `inputs.plan_audit_report` with a
  matching SHA-256, otherwise the highest-numbered `.moai/reports/plan-audit/<SPEC-ID>-review-<n>.md`
  — ends with an audit verdict line naming this SPEC and a verdict equal to the contract's
  `plan_audit.verdict`; `mismatch` when such a line names a different verdict or the receipt-named
  file's hash differs; and `self-reported` when no such file or line exists.

### D.2 Second review

- **REQ-CLOSURE-012** (Event-driven) — When `audit_multi` is called with a `card_id` argument, the
  MCP server shall append one second-review record to
  `<project_root>/.moai/reports/<card-id>/second-review.jsonl` carrying the card ID, the card's SPEC
  ID, the signed contract digest (empty when the contract is absent or unsigned), the HEAD of the
  audited tree, each backend's name, gate, and verdict, the participant count, the disagreement flag,
  the audit receipt ID, the build commit, and the recording time; and when the argument is absent,
  the tool's output and side effects shall be byte-identical to the pre-change behavior.
- **REQ-CLOSURE-013** (Ubiquitous) — The report generator and the push readiness evaluator shall
  treat the second review as performed for a given HEAD and contract digest only when the latest
  second-review record whose HEAD and contract digest equal them carries at least one codex or GLM
  backend verdict of `pass` or `fail`; the second-review verdict shall be `fail` when any such backend
  verdict is `fail` and `pass` otherwise; every other state shall be `not performed` with exactly one
  cause from `no-record`, `stale-head`, `contract-changed`, `no-second-model`, or `unbound`.
- **REQ-CLOSURE-014** (Ubiquitous) — The Second Verdict section shall show the second-review state,
  each counted backend and its verdict, the contract's `review.second_model`, whether the performing
  backend differs from it, and the effective `second_review` policy; where the state is not performed,
  the section shall render the literal text `NOT PERFORMED` followed by the cause.

### D.3 Stop before push

- **REQ-CLOSURE-015** (Capability gate + event) — Where `workflow.autonomy.mode` is `contract`, when a
  Bash tool call pushes the integration branch, the PreToolUse hook shall evaluate push readiness for
  every signed contract of a non-terminal SPEC that lists `push-develop` and whose SPEC directory or
  `ownership.write` paths are changed by a non-merge commit in the range from the remote integration
  ref to the local HEAD, and shall deny the call with a reason beginning `CLOSURE_PUSH_STOP:` followed
  by the sorted, de-duplicated readiness codes when any contract is not ready.
- **REQ-CLOSURE-016** (Ubiquitous) — The push readiness evaluator shall report a contract not ready
  with every applicable code from the closed set: `contract_invalid` (A1 verify is not
  `signed-valid`), `closure_report_missing` (no closure report JSON for the card), and, where
  `second_review` is `required`, `second_review_not_performed`, `second_review_stale` (performed for a
  HEAD that a later in-range non-merge commit touching the contract's paths superseded), and
  `second_review_failed`; plus the human-verdict codes fixed by `plan.md` § Open Questions OQ-1.
- **REQ-CLOSURE-017** (Event-detected) — When, under contract mode, the push range, the card of an
  in-range contract, or that card's evidence directory cannot be determined, the push readiness
  evaluator shall report `push_check_undetermined` and the hook shall deny the push; it shall never
  allow a push on an undetermined result.
- **REQ-CLOSURE-018** (Capability gate) — Where `second_review` is `advisory` or `off`, the push
  readiness evaluator shall not report a second-review code; under `advisory` the report shall still
  render `NOT PERFORMED` with a warning line, and under `off` the Second Verdict section shall state
  `not required`.
- **REQ-CLOSURE-019** (Event-driven) — When `moai contract push-check` runs, it shall perform the same
  evaluation as the hook for the current tree's push range and exit 0 when every in-range contract is
  ready, 1 when any is not ready (printing each SPEC ID with its codes), and 2 on a usage or I/O
  error; under `guided` it shall print one line stating that the check is inactive and exit 0.

### D.4 Human verdict

- **REQ-CLOSURE-020** (Event-driven) — When `moai contract verdict <card-id> <accept|reject|amend-contract>`
  runs, the verdict recorder shall refuse without writing (exit 1) when an agent-environment marker
  from A1's closed marker set is present, when standard input is not an interactive terminal, when
  the typed confirmation differs from the displayed token, or when the card has no closure report;
  otherwise it shall append one record to `.moai/reports/<card-id>/closure-verdict.jsonl` carrying
  the verdict, an optional note, the operator name and email from git configuration, the UTC
  recording time, the SHA-256 of the current `closure-report.json`, and the method `interactive-tty`.
- **REQ-CLOSURE-021** (Unwanted) — The PreToolUse hook shall not allow a Bash tool call that invokes
  `moai contract verdict`; it shall deny it in every mode with a reason beginning
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
- **REQ-CLOSURE-024** (Ubiquitous) — The report generator, the push readiness evaluator, and the
  verdict recorder shall read a card's evidence files from `.moai/reports/<card-id>/` of the card's
  worktree first and of the primary checkout second, per file, and the report shall show the source
  path of each evidence file it read.
- **REQ-CLOSURE-025** (Ubiquitous) — Template content added by this SPEC — auditor instructions
  directing a contract-mode second review to pass the card argument, and command help text — shall
  contain no SPEC ID, card identifier, date, operator-decision identifier, or commit SHA, and shall
  favor no programming language.

## §E. Non-Functional Constraints

- The report is deterministic: two runs over the same files and HEAD produce byte-identical JSON
  except the `generated_at` field.
- The push guard runs only under contract mode and only for a push of the integration branch; its
  git work is bounded by a timeout, and a timeout is `push_check_undetermined` (REQ-CLOSURE-017).
- Every file A4 writes is written atomically (report pair) or appended with a single write per line
  (`.jsonl` records).
- A4 imports A1's verification core for every contract fact and A2's record reader for every
  escalation fact; it re-implements neither schema.

## §F. Assumptions

- A1 (SPEC-AUTONOMY-CONTRACT-001) and A2 (SPEC-AUTONOMY-ESCALATION-001) land on the integration
  branch before this SPEC's run phase; A2b (card t1245) exposes a push-command classifier, or A4
  carries its own (plan.md §B).
- `progress.md` keeps the §E.2 AC matrix row shape `| <AC-ID> | <status> | <evidence> |` and the §E.3
  fenced YAML block, as measured in this tree (research.md §B.5).
- Plan-audit reports keep the `AUDIT-VERDICT:` last-line format the audit receipt store parses.

## §G. References

- `design.md` — report layout, record shapes, readiness algorithm, evidence resolution, package
  layout.
- `research.md` — measured codebase facts and the dependency snapshot.
- `plan.md` — milestones, risks, open questions. `acceptance.md` — acceptance criteria.

## §H. Residual Risk

- **Local files can be forged.** Anyone who can write the working tree can hand-write a
  second-review record, a closure report, or a verdict line. A4 ensures that forgery leaves the same
  traces A1 accepts (git history, the audit receipt store for codex participation) and that a
  missing record stops the push; it does not prove a record's author.
- **Only Bash-tool pushes are observed.** A terminal push, a push through another shell tool, or a
  push with hooks disabled is not stopped by the PreToolUse guard (this repository sets
  `core.hooksPath` to `/dev/null`, research.md §B.8).
- **Staleness is path-based.** A later commit that touches a contract's paths makes its second review
  stale even when another card made the commit; this errs toward stopping.
- **Class-4 comparison is a heuristic** (A2 C6); an empty New APIs list is evidence of no detected
  addition, not of no addition.
