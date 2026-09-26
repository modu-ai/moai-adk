---
id: SPEC-AUTONOMY-CONTRACT-001
title: "Contract-based autonomy A1 — contract schema, acceptance binding, and human signature (moai contract sign/show/verify)"
version: "0.1.0"
status: draft
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/contract, internal/cli, internal/config, internal/template"
lifecycle: spec-anchored
tags: "autonomy, contract, signing, kickoff, acceptance-hash, config, autonomy-a1"
tier: L
related_specs: [SPEC-GTD-AUTONOMY-001, SPEC-AUTONOMY-TIERS-001, SPEC-AUTONOMY-RUN-GOAL-001, SPEC-ACSNAPSHOT-COMMIT-GUARD-001]
---

# SPEC-AUTONOMY-CONTRACT-001 — Contract Schema and Signing (Autonomy A1)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-26 | manager-spec | Initial plan-phase draft (card t1234, AUTONOMY-A1). Operator decisions A-Q1..A-Q4 (2026-09-26) recorded in §C. |

## §A. User Story

As the operator of a MoAI project, I want to approve a SPEC's run once — by signing a machine-readable
contract that names what may be written, what must never be touched, which actions are allowed, what
budget applies, and which events force a stop — so that the run can proceed autonomously inside that
envelope instead of pausing for the per-SPEC Implementation Kickoff question, while any change to the
approved acceptance criteria is detected mechanically.

A1 is the foundation of the contract-autonomy epic. It delivers only the **contract artifact and its
integrity**: the schema, the acceptance binding, the human signature, and the three read/sign
commands. The detectors that act on a contract (A2), the gate rewiring that makes the signature
replace Kickoff (A3), and the closure report (A4) consume this SPEC's schema and are separate SPECs.

## §B. Scope

In scope:

- The `contract.yaml` schema, stored at `.moai/specs/<SPEC-ID>/contract.yaml`, as a typed,
  machine-readable document with a schema version (the canonical draft lives in
  `design.md` § Contract Schema).
- A deterministic canonical digest of the contract body and a deterministic hash of the SPEC's
  `acceptance.md`, plus an acceptance-criterion count bound to the same file.
- A signature record embedded in the contract (who, when, which digest, which HEAD) and the tamper
  semantics that invalidate it.
- The `moai contract sign`, `moai contract show`, and `moai contract verify` commands, including
  batch signing of several SPECs in one confirmation.
- New configuration keys under `workflow.autonomy` in `.moai/config/sections/workflow.yaml`
  (template default `mode: guided`).
- The declarative representation of the push-serialization and second-review requirements inside the
  contract and configuration.

### Out of Scope — Escalation detectors (A2)

- Acceptance-hash watching during a run, the ownership path check at PreToolUse, new-API detection
  through the code graph, contradiction detection, and the enforcement of push serialization and of
  the second-review stop-before-push rule all belong to A2. A1 supplies the verify primitive and the
  schema those detectors read; it enforces nothing at tool-call time.
- The `workflow.autonomy.escalation.new_api_detector` key is not introduced by A1 (see §C.3).

### Out of Scope — Gate rewiring (A3) and closure report (A4)

- Rewriting skill and rule documents so that a valid contract signature replaces the Implementation
  Kickoff Approval gate is A3. Under A1 alone, the Kickoff gate is unchanged in every mode.
- The closure report consumed by the human reviewer (`review.human: closure-report`) is A4.
- Having manager-spec emit a draft `contract.yaml` during plan-phase is A3's documentation change; A1
  defines only what a valid draft looks like.

### Out of Scope — Factory record layer (F1)

- The factory redesign's `factory.db` card records are not implemented or modified here. §C.4 states
  how a card record would reference a signed contract; the storage itself is F1.

### Out of Scope — Cryptographic identity

- A1 does not sign with GPG, SSH, or Sigstore keys. The signature is an integrity-bound approval
  record whose human-presence property rests on an interactive-terminal requirement (§C.2). An actor
  with write access to the working tree and an interactive terminal can re-sign; defending against
  that actor is out of scope.

### Out of Scope — Documentation surfaces

- docs-site pages and README text for the new commands are sync-phase work for manager-docs.

## §C. Context and Decisions

### §C.1 Operator decisions (2026-09-26)

- **A-Q1** — Signing the contract replaces the Implementation Kickoff Approval. A1 builds the
  signature; A3 performs the replacement.
- **A-Q2** — Pushing `develop` is allowed inside a contract, but pushes are serialized the same way
  the merge window is. A1 represents the permission (`push-develop` action plus the
  `push_develop` configuration switch); A2 enforces serialization.
- **A-Q3** — A second-model review is required by default; when it has not been performed, the run
  stops before any push. A1 represents the requirement (`second_review: required` plus a mandatory
  `review.second_model`); A2 enforces the stop.
- **A-Q4** — The distributed template ships `mode: guided`.

### §C.2 Human presence

The signature is meaningful only if an agent cannot produce it. Agent tool calls run without an
interactive terminal, so signing requires an interactive terminal on standard input and a typed
confirmation. This is a presence check, not an identity proof (§B Out of Scope — Cryptographic
identity).

### §C.3 Why `new_api_detector` is not reserved in A1

The configuration loader ignores unknown keys, so A2 can add the key without a schema migration. A
key reserved by A1 would have no reader and no behavior — a switch that does nothing while appearing
to — and its value set (`graph | off`) is A2's design decision. A1 therefore ships
`escalation.budget_default` only, which `sign` actually reads.

### §C.4 Relationship to the factory record layer (F1)

The contract file stays the single source of truth. A factory card record would carry a pointer only
— the SPEC ID, the signed contract digest, and the signing time — never a copy of the contract. A
reader that finds the pointer's digest different from the file's current signed digest treats the
card's contract reference as stale. Detail: `design.md` § F1 Reference Shape.

### §C.5 Relationship to existing SPECs

- **SPEC-GTD-AUTONOMY-001** built the sealed mission contract behind `/moai goal --auto`
  (`internal/mission`). A1 reuses its canonical-sealing technique, its integrity-digest receipt
  pattern, and its action vocabulary through an explicit mapping; the SPEC contract projects onto a
  mission contract so A2 can reuse the mission decision validator (`research.md` § Reuse Analysis).
- **SPEC-AUTONOMY-TIERS-001** owns `MOAI_AUTONOMY_TIER`, which sets permission-bundle and hook block
  strength. `workflow.autonomy.mode` is a different axis (who approves a run) with a disjoint value set
  (`guided | contract`); A1 adds no environment variable, so no name can collide.
- **SPEC-AUTONOMY-RUN-GOAL-001** defined run-phase `/goal ac_converge` wrapping with Kickoff
  preserved. A1 does not change that preservation; A3 revisits it.
- **SPEC-ACSNAPSHOT-COMMIT-GUARD-001** guards the corpus AC-count snapshot. A1's AC count uses the
  same counting semantics (the counter published in the manager-docs agent definition) so the two
  numbers cannot disagree for the same file.

## §D. Requirements (GEARS)

### REQ-CONTRACT-001 — Contract location and schema version

The contract shall be a YAML document stored at `.moai/specs/<SPEC-ID>/contract.yaml`, carrying a
`schema_version` field and a `spec_id` field equal to the name of the SPEC directory that contains it.

### REQ-CONTRACT-002 — Required sections

The contract shall carry the sections `acceptance`, `invariants`, `ownership` (with `write` and
`never`), `approach`, `actions`, `reobserve`, `review`, `budget`, `escalate_on`, and `plan_audit`,
with the field shapes and value sets defined in `design.md` § Contract Schema.

### REQ-CONTRACT-003 — Strict decoding

When a contract contains a field the schema does not define, the contract reader shall reject the
contract as schema-invalid rather than ignore the field, so that no content can sit outside the
signed digest.

### REQ-CONTRACT-004 — Canonical digest

The contract digest shall be a SHA-256 over a canonical form of every contract field except the
`signature` block, such that reordering set-valued lists, changing comments, or changing YAML
formatting does not change the digest, and changing any field value does.

### REQ-CONTRACT-005 — Acceptance binding

The contract shall bind the SPEC's `acceptance.md` through a SHA-256 of its content (after
line-ending normalization to LF and removal of a leading UTF-8 byte-order mark) and through the
acceptance-criterion count produced by the project's published AC counter semantics.

### REQ-CONTRACT-006 — Closed action vocabulary

The contract's `actions` list shall accept only `commit`, `worktree`, `local-merge-develop`, and
`push-develop`. When the list names `push-main`, `merge-main`, `force-push`, `release-branch`, or
`release-pr`, the verifier shall report a forbidden action; when the list names any other token, the
verifier shall report an unknown action.

### REQ-CONTRACT-007 — Escalation set completeness

The contract's `escalate_on` list shall contain all six triggers — `acceptance-change`,
`invariant-violation`, `ownership-move`, `new-architecture-or-api`, `contradictory-evidence`, and
`irreversible-action` — and shall not contain any other token.

### REQ-CONTRACT-008 — Verify semantics

When `moai contract verify <SPEC-ID>` runs, the verifier shall report the contract valid only when the
contract decodes strictly, satisfies every schema rule, carries a signature whose recorded digest
equals the recomputed digest, and binds an `acceptance.md` whose current hash and AC count equal the
recorded ones; otherwise it shall report invalid with every applicable reason code from the closed
set in `design.md` § Verify Reason Codes.

### REQ-CONTRACT-009 — Verify is read-only and self-contained

The verifier shall not write any file, spawn any subprocess, or access the network, and shall exit
0 for a valid contract, 1 for an invalid contract, and 2 for a usage or I/O error.

### REQ-CONTRACT-010 — Human-presence signing

When `moai contract sign` runs without an interactive terminal on standard input, the signer shall
refuse to sign and shall exit non-zero without modifying any file. While standard input is an
interactive terminal, the signer shall display the contract summary and shall sign only after the
operator types the confirmation token shown in the prompt.

### REQ-CONTRACT-011 — Signature record

When the signer signs a contract, it shall record, inside the contract's `signature` block, the
signer's name and email from git configuration, the signing time in UTC RFC 3339 form, the HEAD
commit at signing, the contract digest, the signing method, and — when it replaces an earlier
signature — the digest that signature carried.

### REQ-CONTRACT-012 — Sign-time binding and refusal

When the signer signs, it shall measure the acceptance hash and AC count and write them into the
contract, fill an absent `budget` from `workflow.autonomy.escalation.budget_default`, and write the
file atomically. When the draft already records an acceptance hash different from the measured one,
when the AC counter reports an ambiguous result, when `plan_audit.verdict` is neither `PASS` nor
`PASS-WITH-DEBT`, or when any verify rule other than the missing signature fails, the signer shall
refuse and leave the file unchanged.

### REQ-CONTRACT-013 — Batch signing

Where `workflow.autonomy.contract.batch_sign` is true, the signer shall accept several SPEC IDs in one
invocation, validate all of them before writing any, present one combined summary, and sign all of
them after a single confirmation, recording a shared batch identifier in each signature. Where
`batch_sign` is false, the signer shall refuse an invocation naming more than one SPEC ID.

### REQ-CONTRACT-014 — Show

When `moai contract show <SPEC-ID>` runs, the command shall print the contract's sections, its
signature state (`unsigned`, `signed-valid`, or `signed-invalid`), and the verify reason codes; with
`--json` it shall print the same information as a single JSON object whose field names are stable
for downstream consumers.

### REQ-CONTRACT-015 — Configuration keys and defaults

The configuration shall expose `workflow.autonomy.mode` (`guided | contract`),
`workflow.autonomy.contract.batch_sign`, `workflow.autonomy.contract.second_review`
(`required | advisory | off`), `workflow.autonomy.contract.push_develop`, and
`workflow.autonomy.escalation.budget_default` (`turns`, `operations`, `audit_retries`). When the
section or a key is absent, the reader shall apply `mode: guided`, `batch_sign: false`,
`second_review: required`, `push_develop: false`, and budget `60 / 40 / 2`. When `mode` holds a value
outside its set, the reader shall fall back to `guided` and emit a warning.

### REQ-CONTRACT-016 — Template default

The distributed template's `workflow.yaml` shall carry the `workflow.autonomy` section with
`mode: guided` and `second_review: required`, and shall contain no SPEC ID, card identifier, or date.

### REQ-CONTRACT-017 — Second-review representation

Where `workflow.autonomy.contract.second_review` is `required`, the verifier shall report a contract
whose `review.second_model` is `none` or empty as invalid. Where it is `advisory` or `off`, the
verifier shall accept `none`.

### REQ-CONTRACT-018 — Push-develop representation

Where `workflow.autonomy.contract.push_develop` is false, the verifier shall report a contract whose
`actions` list includes `push-develop` as invalid. The `push-develop` action shall denote a push of
the integration branch performed while holding the integration window; the verifier shall expose
this as a derived `push_requires_window: true` field in `show --json` output for A2 to enforce.

### REQ-CONTRACT-019 — Guided-mode neutrality

While `workflow.autonomy.mode` is `guided`, the contract commands shall remain usable and the signer
shall print a notice that the signature does not replace Implementation Kickoff Approval in guided
mode; no existing workflow behavior shall change in either mode as a result of this SPEC.

### REQ-CONTRACT-020 — Ownership and invariant well-formedness

The verifier shall report a contract invalid when `ownership.write` is empty, when any ownership glob
is absolute or contains a `..` segment, when the same glob appears in both `write` and `never`, when
`ownership.write` does not cover the SPEC's own directory, or when a `constitution:<pattern>`
invariant resolves to no rule in the constitution registry.

### REQ-CONTRACT-021 — Mission-contract projection

The contract shall be projectable onto the existing sealed mission contract used by
`/moai goal --auto`, and the projection's sealed hash shall be recorded in the signature, so that a
downstream consumer can apply the mission decision validator to a signed SPEC contract without a
second schema.

## §E. Non-Functional Constraints

- Verify completes without subprocesses or network access so that A2 can call it from a PreToolUse
  hook.
- Every file the signer writes is written atomically; a failed batch leaves every contract that was
  not yet written byte-identical.
- The acceptance hash and AC count are identical on darwin, linux, and windows for the same logical
  file content (line-ending normalization, REQ-CONTRACT-005).

## §F. Assumptions

- The published AC counter in the manager-docs agent definition remains the canonical AC-count
  semantics; A1's count is a port verified by a parity test against it over the full SPEC corpus.
- `git config user.name` and `user.email` are set where signing happens; when absent, signing refuses
  rather than recording an anonymous signer.
- Plan-audit verdicts use the `PASS` / `PASS-WITH-DEBT` / `FAIL` vocabulary.

## §G. References

- `design.md` § Contract Schema — the schema draft forwarded to A2, A3, and A4.
- `research.md` § Reuse Analysis — what is reused from `internal/mission` and what is new.
- `plan.md` — milestones and risks. `acceptance.md` — acceptance criteria.
