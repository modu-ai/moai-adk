---
id: SPEC-AUTONOMY-CONTRACT-001
title: "Contract-based autonomy A1 — contract schema, acceptance binding, and human signature (moai contract sign/show/verify)"
version: "0.4.1"
status: in-progress
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
| 0.2.0 | 2026-09-26 | manager-spec | Plan-audit iteration 1 (FAIL 0.74) revisions: mission-validator projection deferred to A2 (former REQ-021 withdrawn; ID 021 reused for agent-environment refusal); re-sign after acceptance change defined (REQ-022); human-presence claim restated; A2-before-A3 ordering recorded; Kickoff notice printed in both modes; verify isolated in its own package; second_review fallback, non-empty actions, and ownership coverage semantics specified. |
| 0.3.0 | 2026-09-26 | manager-spec | Plan-audit iteration 2 (FAIL 0.84) revisions and lead-approved schema additions. **ID reuse record:** AC-CONTRACT-024 was reused in 0.2.0 (was "mission-contract projection", now "re-sign after an acceptance change", mapped to REQ-CONTRACT-022); REQ-CONTRACT-021 was reused likewise. Both reuses happened before implementation; citations of REQ-CONTRACT-021 or AC-CONTRACT-024 in the iteration-1 plan-audit report refer to the withdrawn meanings. Other changes: enforcement owners aligned with cards (A2 = t1235 push serialization + sign deny; A4 = t1237 second-review stop; ordering A1 → A2 → A3); AC pass convention requires a `--- PASS:` line; kickoff receipt format, non-interactive receipt signing path, `workflow.autonomy.kickoff` config, post-signing immutability, `ownership.scratch`, and `frozen-files` definition added (REQ-CONTRACT-023..025); autonomous-Kickoff activation preconditions and residual-risk section added; `moai contract revoke` declared out of scope (A3). |
| 0.4.0 | 2026-09-26 | manager-spec | Plan-audit iteration 3 (FAIL 0.84) delta, lead-authorized iteration 4: signature seal and signature consistency rules (R1); Jev answers route to a human until A3 reconciles the display-only doctrine, and a Jev decider with `workflow.jev.enabled: false` is a config error (R2, §C.8); card-state assertions replaced by required-ordering statements, with push serialization and the sign deny owned by A2b (card t1245) (R3); `frozen_files` elements typed as globs (R4); jev confidence pinned in AC-016 (R5); push serialization re-specified as a `moai slot` lease on resource `push-develop` (`push_requires_window` renamed `push_requires_lease`); terminal contract state for `completed`/`archived` SPECs. |
| 0.4.1 | 2026-09-26 | manager-spec | Plan-audit iteration 4 text fixes (operator: text only): §C.6 and §H now state one consistent residual-risk statement (a full keyless re-seal is not caught by verify; `supersedes` is not a trace; traces are git history and, after A3, a store recording every signing event including human signatures); stale A2 attributions changed to A2b (t1245); REQ-015 states that the Jev-decider configuration error does not block loading or commands and refuses only the receipt signing path. |

## §A. User Story

As the operator of a MoAI project, I want to approve a SPEC's run once — by signing a machine-readable
contract that names what may be written, what must never be touched, which actions are allowed, what
budget applies, and which events force a stop — so that the run can proceed autonomously inside that
envelope instead of pausing for the per-SPEC Implementation Kickoff question, while any change to the
approved acceptance criteria is detected mechanically.

A1 is the foundation of the contract-autonomy epic. It delivers only the **contract artifact and its
integrity**: the schema, the acceptance binding, the signature (human, or receipt-based for an
automated decider), the kickoff receipt format and its validator, and the three read/sign commands.
The detectors and guards that act on a contract (A2, A2b), the gate rewiring that makes the signature
replace Kickoff (A3), and the closure report and second-review stop (A4) consume this SPEC's schema
and are separate SPECs.

## §B. Scope

In scope:

- The `contract.yaml` schema, stored at `.moai/specs/<SPEC-ID>/contract.yaml`, as a typed,
  machine-readable document with a schema version (the canonical draft lives in
  `design.md` § Contract Schema).
- A deterministic canonical digest of the contract body and a deterministic hash of the SPEC's
  `acceptance.md`, plus an acceptance-criterion count bound to the same file.
- A signature record embedded in the contract, the tamper semantics that invalidate it, and the
  re-sign path after a legitimate acceptance change.
- The kickoff receipt format (`kickoff-receipt.json`) and its validator, and a non-interactive signing
  path that accepts a validated receipt.
- The `moai contract sign`, `moai contract show`, and `moai contract verify` commands, including
  batch signing of several SPECs in one human confirmation.
- New configuration keys under `workflow.autonomy` in `.moai/config/sections/workflow.yaml`
  (template default `mode: guided`, `kickoff.decider: human`).
- The declarative representation of post-signing immutability, `ownership.scratch`, the `frozen-files`
  invariant, the push-serialization requirement, and the second-review requirement.

### Out of Scope — Escalation detectors and guards (A2, card t1235; A2b, card t1245)

- Acceptance-hash watching during a run, the ownership path check at PreToolUse (including the
  post-signing immutability of `contract.yaml` and `acceptance.md`), new-API detection through the
  code graph, and contradiction detection belong to A2 (t1235). **Push-serialization enforcement** and
  the **PreToolUse deny on agent-invoked `moai contract sign`** belong to A2b (t1245). A1 supplies the
  verify primitive and the schema those detectors read; it enforces nothing at tool-call time.
- A2 resolves which contract a detection applies to; contracts of terminal SPECs (REQ-CONTRACT-014)
  are excluded from that resolution and from signature resolution.
- The `workflow.autonomy.escalation.new_api_detector` key is not introduced by A1 (see §C.3).

### Out of Scope — Mission-validator projection (A2)

- Projecting a signed SPEC contract onto the `/moai goal --auto` mission contract so that the mission
  decision validator can judge individual operations is deferred to A2, its only consumer. The
  projection needs a glob-to-prefix scope translation that the mission validator's exact-or-prefix
  containment check requires. Forward note for A2: `design.md` § Forward Note — Mission Projection.
- A1 records no `mission_contract_sha256`; the contract digest is the sole tamper authority.

### Out of Scope — Revocation and receipt issuance (A3, card t1236)

- `moai contract revoke` is not part of A1; A3 owns it.
- Issuing kickoff receipts from moai itself — moai calling Jev directly and appending the Jev result
  together with the LLM decision record to a moai-owned append-only store — is A3. A1 validates a
  receipt **file**; it cannot establish who wrote it.

### Out of Scope — Gate rewiring (A3) and closure report (A4, card t1237)

- Rewriting skill and rule documents so that a valid contract signature replaces the Implementation
  Kickoff Approval gate is A3. Under A1 alone, the Kickoff gate is unchanged in every mode.
- The closure report consumed by the human reviewer (`review.human: closure-report`), the record that
  a second-model review was actually performed, and the **stop before push when it was not** are A4.
- Having manager-spec emit a draft `contract.yaml` during plan-phase is A3's documentation change; A1
  defines only what a valid draft looks like.

### Out of Scope — Factory record layer (F1)

- The factory redesign's `factory.db` card records are not implemented or modified here. §C.4 states
  how a card record would reference a signed contract; the storage itself is F1.

### Out of Scope — Cryptographic identity

- A1 does not sign with GPG, SSH, or Sigstore keys. The signature is an integrity-bound approval
  record; its human-presence checks (§C.2) raise the bar but are not an identity proof.

### Out of Scope — Documentation surfaces

- docs-site pages and README text for the new commands are sync-phase work for manager-docs.

## §C. Context and Decisions

### §C.1 Operator decisions (2026-09-26), owners, and epic ordering

- **A-Q1** — Signing the contract replaces the Implementation Kickoff Approval. A1 builds the
  signature; A3 (t1236) performs the replacement.
- **A-Q2** — Pushing `develop` is allowed inside a contract, but pushes are serialized. Lead decision
  (2026-09-26): the serialization mechanism is a `moai slot` lease on the resource `push-develop`, not
  the integration (merge) window — the push happens outside the merge window and is a different
  resource. A1 represents the permission (`push-develop` action plus the `push_develop` configuration
  switch) and the lease requirement (`push_requires_lease`, REQ-CONTRACT-018); **A2b (t1245) enforces
  serialization.**
- **A-Q3** — A second-model review is required by default; when it has not been performed, the run
  stops before any push. A1 represents the requirement (`second_review: required` plus a mandatory
  `review.second_model`, which names the planned reviewer). **A4 (t1237) owns the stop and the record
  that the review was performed** (as an amendment to `design.md` § Contract Schema or a run-time
  receipt of its choosing).
- **A-Q4** — The distributed template ships `mode: guided`.

**Required ordering: A1 → A2 (t1235) + A2b (t1245) → A3 (t1236).** A3 shall not activate "signature
replaces Kickoff" until A2 and A2b have landed — in particular A2b's push-serialization enforcement and
PreToolUse deny on agent-invoked `moai contract sign`. A4 (t1237) owns the stop before push when the
second review was not performed, and the `push-develop` action shall not be activated until A4 has
landed. Until then a signed contract and `mode: contract` confer no autonomy: the Kickoff gate stays in
force (REQ-CONTRACT-019). This is the ordering the epic requires; it is not a statement about the
current content of the queue.

### §C.2 Human presence

The human signature is meant to be produced by a human, and A1's checks **raise the bar for an agent
without stopping one**. The human path requires an interactive terminal on standard input, a typed
confirmation, and the absence of agent-harness environment markers (REQ-CONTRACT-021). None of these
is agent-proof: a one-line pseudo-terminal wrapper gives an agent's shell a terminal on standard input,
and an agent can unset environment variables.

**Hard precondition for A3.** Before A3 makes the signature replace Implementation Kickoff Approval, the
PreToolUse deny on `moai contract sign` issued from agent tool calls shall exist; A2b (t1245) owns it.

### §C.3 Why `new_api_detector` is not reserved in A1

The configuration loader ignores unknown keys, so A2 can add the key without a schema migration. A
key reserved by A1 would have no reader and no behavior, and its value set is A2's design decision.
A1 therefore ships `escalation.budget_default` only, which `sign` actually reads.

### §C.4 Relationship to the factory record layer (F1)

The contract file stays the single source of truth. A factory card record would carry a pointer only
— the SPEC ID, the signed contract digest, and the signing time — never a copy of the contract. A
reader that finds the pointer's digest different from the file's current signed digest treats the
card's contract reference as stale. Detail: `design.md` § F1 Reference Shape.

### §C.5 Relationship to existing SPECs

- **SPEC-GTD-AUTONOMY-001** built the sealed mission contract behind `/moai goal --auto`
  (`internal/mission`). A1 reuses its canonical-sealing technique (clone, sort, marshal, hash) for the
  contract digest, its integrity-digest receipt pattern for the signature, and `internal/atomicfile`
  for writes (`research.md` § Reuse Analysis). The mission-validator projection is deferred to A2.
- **SPEC-AUTONOMY-TIERS-001** owns `MOAI_AUTONOMY_TIER`, which sets permission-bundle and hook block
  strength. `workflow.autonomy.mode` is a different axis (who approves a run) with a disjoint value set
  (`guided | contract`); A1 adds no environment variable, so no name can collide.
- **SPEC-AUTONOMY-RUN-GOAL-001** defined run-phase `/goal ac_converge` wrapping with Kickoff
  preserved. A1 does not change that preservation; A3 revisits it.
- **SPEC-ACSNAPSHOT-COMMIT-GUARD-001** guards the corpus AC-count snapshot. A1's AC count uses the
  same counting semantics (the counter published in the manager-docs agent definition). A1 counts
  **normalized** bytes (LF, no BOM) while the snapshot guard counts raw bytes, so the two numbers are
  guaranteed equal only for LF, BOM-free files; AC-CONTRACT-006 measures the raw-versus-normalized
  difference on its CRLF fixture.

### §C.6 Autonomous Kickoff activation preconditions (binding)

A contract signed by a non-human decider (`llm`, `jev`, or `llm+jev`) shall not be used to activate
autonomous Kickoff until **both** hold: (a) `moai contract revoke` exists (A3, t1236), and (b) kickoff
receipts are issued and stored by moai itself — moai calls Jev directly and appends the result with the
LLM decision record to a moai-owned append-only store (A3). A receipt authored as agent-written JSON
shall not activate autonomous Kickoff. A1 provides only the receipt format, the validator, and a
signature that records the receipt's provenance as `file` (REQ-CONTRACT-024), so that A3 can refuse it.
The provenance, method, and signer kind are covered by the signature seal (REQ-CONTRACT-011): editing
any of them without recomputing the seal makes `verify` report `signature_seal_mismatch`. A full keyless
re-seal is not caught by `verify`, and A1 leaves no mechanical trace of it; `supersedes` is not a trace,
because the forger rewrites and re-seals it too. The only traces are git history (when the pre-forgery
contract was committed) and, after A3, the moai-owned store. Forward requirement for A3 (t1236): that
store shall record every signing event, including human signatures — not only receipt issuance.

### §C.7 Plan-audit verdict is self-reported

`plan_audit.verdict` is a value written into the draft; A1 checks only that it is `PASS` or
`PASS-WITH-DEBT`. It is not bound to an audit report in the human path, and `sign --resign` carries the
previous verdict forward (the summary shows it as carried). A3 and A4 shall bind the verdict to
evidence before relying on it. The receipt path binds a plan-audit report hash (REQ-CONTRACT-023), which
is the first such binding.

### §C.8 Jev is display-only until A3 reconciles the doctrine

The product ships the principle that a Jev answer is display-only — never an input to a completion
verdict, a merge approval, a queue mutation, or a gate — at
`internal/template/templates/.moai/config/sections/workflow.yaml:171-173`,
`.claude/rules/moai/core/moai-mcp-tools.md:75`, and
`.claude/rules/moai/core/moai-mcp-tools-catalogue.md:138`. The operator decision of 2026-09-26
authorizes a Kickoff-only exception, but A1 does not make it: the receipt validator treats **any** Jev
decision as unmeasured and routes it to a human (`receipt_requires_human`, REQ-CONTRACT-023), so `jev`
and `llm+jev` receipts cannot sign in A1. Forward note for A3 (t1236): amend the display-only statement
in all three locations, citing the operator decision, before any Jev answer gates Kickoff. Separately, a
Jev decider while the capability is off (`workflow.jev.enabled: false`) is a configuration error, not a
fallback (REQ-CONTRACT-015).

## §D. Requirements (GEARS)

### REQ-CONTRACT-001 — Contract location and schema version

The contract shall be a YAML document stored at `.moai/specs/<SPEC-ID>/contract.yaml`, carrying a
`schema_version` field and a `spec_id` field equal to the name of the SPEC directory that contains it.

### REQ-CONTRACT-002 — Required sections

The contract shall carry the sections `acceptance`, `invariants`, `ownership` (with `write`, `never`,
and an optional `scratch`), `approach`, `actions`, `reobserve`, `review`, `budget` (whose
`audit_retries` is the shared retry cap for plan-audit and sync-audit), `escalate_on`, and
`plan_audit`, with the field shapes and value sets defined in `design.md` § Contract Schema.

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
acceptance-criterion count produced by the project's published AC counter semantics, applied to the
same normalized content.

### REQ-CONTRACT-006 — Closed action vocabulary

The contract's `actions` list shall be non-empty and shall accept only `commit`, `worktree`,
`local-merge-develop`, and `push-develop`. When the list is empty, the verifier shall report
`actions_empty`; when the list names `push-main`, `merge-main`, `force-push`, `release-branch`, or
`release-pr`, the verifier shall report a forbidden action; when the list names any other token, the
verifier shall report an unknown action.

### REQ-CONTRACT-007 — Escalation set completeness

The contract's `escalate_on` list shall contain all six triggers — `acceptance-change`,
`invariant-violation`, `ownership-move`, `new-architecture-or-api`, `contradictory-evidence`, and
`irreversible-action` — and shall not contain any other token.

### REQ-CONTRACT-008 — Verify semantics

When `moai contract verify <SPEC-ID>` runs, the verifier shall report the contract valid only when the
contract decodes strictly, satisfies every schema rule, carries a signature whose recorded digest
equals the recomputed digest, carries a signature seal equal to its recomputation, satisfies the
signature consistency rules (method, signer kind, receipt presence, provenance), carries a
`signature.acceptance_sha256` equal to the measured acceptance hash, binds an `acceptance.md` whose
current hash and AC count equal the recorded ones, and — for a receipt-signed contract — finds the recorded receipt file with its recorded
hash; otherwise it shall report invalid with every applicable reason code from the closed set in
`design.md` § Verify Reason Codes.

### REQ-CONTRACT-009 — Verify is read-only and self-contained

The verifier shall not write any file, spawn any subprocess, or access the network, and shall exit
0 for a valid contract, 1 for an invalid contract, and 2 for a usage or I/O error. The verification
logic shall live in a package whose transitive imports include neither a process-spawning nor a
network package.

### REQ-CONTRACT-010 — Human signing path

Where the signer is human (no `--signer`, or `--signer human`), the signer shall refuse to sign and exit
non-zero without modifying any file when standard input is not an interactive terminal; while standard
input is an interactive terminal, the signer shall display the contract summary and shall sign only
after the operator types the confirmation token shown in the prompt.

### REQ-CONTRACT-011 — Signature record

When the signer signs a contract, it shall record, inside the contract's `signature` block, the signer
kind (`human`, `llm`, `jev`, or `llm+jev`), the operator name and email from git configuration, the
signing time in UTC RFC 3339 form, the HEAD commit at signing, the contract digest, the acceptance
hash, the signing method, for a receipt signature the receipt path, hash, and provenance, when it
replaces an earlier signature the digest that signature carried, and a seal — a SHA-256 over the
canonical form of every other signature field together with the contract digest — so that an edit to
any signature field is detectable.

### REQ-CONTRACT-012 — Sign-time binding and refusal

When the signer signs, it shall measure the acceptance hash and AC count and write them into the
contract, fill an absent `budget` from `workflow.autonomy.escalation.budget_default`, and write the
file atomically. When an unsigned draft already records an acceptance hash different from the
measured one, when the AC counter reports an ambiguous result or a count of zero, when
`plan_audit.verdict` is neither `PASS` nor `PASS-WITH-DEBT`, when git `user.name` or `user.email` is
empty, or when any verify rule other than the missing signature and the acceptance binding fails, the
signer shall refuse with a code from `design.md` § Sign Refusal Codes and leave the file unchanged.

### REQ-CONTRACT-013 — Batch signing

Where `workflow.autonomy.contract.batch_sign` is true and the signer is human, the signer shall accept
several SPEC IDs in one invocation, de-duplicate them, validate all of them before writing any, present
one combined summary, and sign all of them after a single confirmation, recording a shared batch
identifier in each signature. Where `batch_sign` is false, or the signer is not human, the signer shall
refuse an invocation naming more than one distinct SPEC ID.

### REQ-CONTRACT-014 — Show

When `moai contract show <SPEC-ID>` runs, the command shall print the contract's sections, its
signature state (`unsigned`, `signed-valid`, or `signed-invalid`), the verify reason codes, the
derived sets defined in `design.md` § Derived Fields, and a `terminal` flag that is true when the SPEC's
`spec.md` frontmatter `status` is `completed` or `archived` (such a contract is terminal and is excluded
from detection and signature resolution); with `--json` it shall print the same information
as a single JSON object whose field names are stable for downstream consumers.

### REQ-CONTRACT-015 — Configuration keys and defaults

The configuration shall expose `workflow.autonomy.mode` (`guided | contract`),
`workflow.autonomy.contract.batch_sign`, `workflow.autonomy.contract.second_review`
(`required | advisory | off`), `workflow.autonomy.contract.push_develop`,
`workflow.autonomy.escalation.budget_default` (`turns`, `operations`, `audit_retries`), and
`workflow.autonomy.kickoff.decider` (`human | llm | jev | llm+jev`),
`workflow.autonomy.kickoff.jev_min_confidence` (0.0–1.0), and `workflow.autonomy.kickoff.on_disagree`
(`human | reject`). When the section or a key is absent, the reader shall apply `mode: guided`,
`batch_sign: false`, `second_review: required`, `push_develop: false`, budget `60 / 40 / 2`,
`decider: human`, `jev_min_confidence: 0.50`, and `on_disagree: human`. When `mode`, `second_review`,
`decider`, `on_disagree`, or `jev_min_confidence` holds a value outside its set or range, the reader
shall fall back to `guided`, `required`, `human`, `human`, or `0.50` respectively and emit a warning
naming the key. When `decider` is `jev` or `llm+jev` while `workflow.jev.enabled` is false, the reader
shall report a configuration error naming both keys and shall not fall back; configuration loading
still succeeds and the error is surfaced, no command is blocked, and only the receipt signing path is
refused (`kickoff_decider_jev_disabled`).

### REQ-CONTRACT-016 — Template default and neutrality

The distributed template's `workflow.yaml` shall carry the `workflow.autonomy` section with
`mode: guided`, `second_review: required`, and `kickoff.decider: human`, and shall contain no SPEC ID,
card identifier, date, operator-decision identifier (`A-Q<n>`), or statement about this repository's
own local settings, in values or comments.

### REQ-CONTRACT-017 — Second-review representation

Where `workflow.autonomy.contract.second_review` is `required`, the verifier shall report a contract
whose `review.second_model` is `none` or empty as invalid. Where it is `advisory` or `off`, the
verifier shall accept `none`.

### REQ-CONTRACT-018 — Push-develop representation

Where `workflow.autonomy.contract.push_develop` is false, the verifier shall report a contract whose
`actions` list includes `push-develop` as invalid. The `push-develop` action shall denote a push of
the integration branch performed while holding a `moai slot` lease on the resource `push-develop`
(not the integration merge window); the verifier shall expose this as a derived
`push_requires_lease: true` field in `show --json` output for A2b to enforce.

### REQ-CONTRACT-019 — Kickoff-neutral notice in every mode

When the signer completes a signature on either signing path, it shall print a notice that the
signature does not yet replace Implementation Kickoff Approval, in both `guided` and `contract` modes,
until A3 amends this requirement. No existing workflow gate shall change in either mode as a result of
this SPEC.

### REQ-CONTRACT-020 — Ownership and invariant well-formedness

The verifier shall report a contract invalid when `ownership.write` is empty, when any `write`, `never`,
or `scratch` glob is absolute or contains a `..` segment, when the same glob string appears in both
`never` and `write` or in both `never` and `scratch`, when no `ownership.write` glob matches the path
`.moai/specs/<SPEC-ID>/contract.yaml` under the glob semantics defined in `design.md` § Glob Semantics,
or when a `constitution:<pattern>` invariant resolves to no rule in the constitution registry supplied
to the verifier.

### REQ-CONTRACT-021 — Agent-environment refusal (human path)

Where the signer is human, when `moai contract sign` runs in a process whose environment carries an
agent-harness marker from the closed marker set in `design.md` § Agent-Environment Markers, the signer
shall refuse to sign, name the detected marker, and exit non-zero without modifying any file.

### REQ-CONTRACT-022 — Re-sign after acceptance change

When `moai contract sign --resign` runs on a signed contract, the signer shall re-measure the
acceptance hash and AC count, show the recorded and newly measured values side by side together with
the carried plan-audit verdict, apply every check of the signing path in use (human or receipt), and on
success write the new acceptance binding and a new signature whose `supersedes` field carries the
previous signature's contract digest. When `--resign` runs on an unsigned contract, the signer shall
refuse with `not_signed` and leave the file unchanged.

### REQ-CONTRACT-023 — Kickoff receipt format and validation

The kickoff receipt shall be a JSON document with the fields defined in `design.md` § Kickoff Receipt,
carrying the signer kind, one decision entry per decider (answer `approve | reject | escalate`, a reason
that references contract lines, and a confidence), the hashes of its inputs (the signable contract
digest, `acceptance.md`, and the plan-audit report), and for a Jev decider the raw response and the
request hash. The receipt validator shall accept a receipt only when it decodes strictly, its input
hashes equal the current files' hashes, its signer kind equals `workflow.autonomy.kickoff.decider`, and
its decisions satisfy the agreement rule in `design.md` § Agreement Rule, under which any Jev decision
is treated as unmeasured and routes to a human until A3 reconciles the display-only doctrine (§C.8);
otherwise it shall return one
receipt refusal code from `design.md` § Sign Refusal Codes.

### REQ-CONTRACT-024 — Receipt signing path

Where the signer is `llm`, `jev`, or `llm+jev`, `moai contract sign --signer <kind> --receipt <path>`
shall run non-interactively, shall sign only while `workflow.autonomy.mode` is `contract` and the
receipt validator accepts the receipt, shall record the receipt path, its SHA-256, and the provenance
`file` in the signature, and shall otherwise refuse with a code from `design.md` § Sign Refusal Codes
without modifying any file.

### REQ-CONTRACT-025 — Post-signing immutability, scratch, and frozen files

The contract shall declare, and `show --json` shall report as derived sets, that once signed the SPEC's
`contract.yaml` and `acceptance.md` belong to the effective `never` set during a run even when an
`ownership.write` glob matches them; that `ownership.scratch` paths are exempt from ownership
escalation in addition to the detectors' fixed exemptions; and that the `frozen-files` invariant
denotes the union defined in `design.md` § Frozen Files, resolved from sources supplied to the verifier.

## §E. Non-Functional Constraints

- Verify completes without subprocesses or network access so that A2 can call it from a PreToolUse
  hook.
- Every file the signer writes is written atomically; a failed batch leaves every contract that was
  not yet written byte-identical.
- The acceptance hash and AC count are identical on darwin, linux, and windows for the same logical
  file content (line-ending normalization, REQ-CONTRACT-005).

## §F. Assumptions

- The published AC counter in the manager-docs agent definition remains the canonical AC-count
  semantics; A1's count is a port verified by a parity test against it (full corpus plus synthetic
  fixtures per counter branch).
- `git config user.name` and `user.email` are set where signing happens; when absent, signing refuses
  rather than recording an anonymous operator.
- Plan-audit verdicts use the `PASS` / `PASS-WITH-DEBT` / `FAIL` vocabulary.
- Jev answers carry a confidence in [0, 1]; the receipt stores the raw response so a reader can check
  the recorded confidence against it.

## §G. References

- `design.md` § Contract Schema — the schema draft forwarded to A2, A3, and A4.
- `research.md` § Reuse Analysis — what is reused from `internal/mission` and what is new.
- `plan.md` — milestones and risks. `acceptance.md` — acceptance criteria.

## §H. Residual Risk

- **A Jev call cannot be proven.** A1 validates a receipt file's shape, hashes, and agreement rule. It
  cannot prove that a Jev request was made or that the raw response came from Jev; a receipt can be
  written by hand or by an agent. This is why §C.6 forbids activating autonomous Kickoff on a
  file-provenance receipt.
- **Local single-user tamper-proofing is impossible.** Anyone who can write the working tree can edit
  and re-sign a contract, rewrite a receipt, or unset the agent markers. A1's goal is that forgery
  leaves a trace, not that forgery is prevented. An edit to any contract or signature field without
  recomputing the digest and the seal is detected by `verify`. A full keyless re-seal is not caught by
  `verify`, and A1 leaves no mechanical trace of it; `supersedes` is not a trace, because the forger
  rewrites and re-seals it too. The only traces are git history (when the pre-forgery contract was
  committed) and, after A3, the moai-owned store recording every signing event, including human
  signatures (§C.6).
- **The human path is not agent-proof** (§C.2) until A2b's (t1245) PreToolUse deny exists, and even then
  only for tool calls the hook observes.
