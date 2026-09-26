# plan.md — SPEC-AUTONOMY-CONTRACT-001

## §A. Context

A1 of the contract-autonomy epic (card t1234). It produces the contract artifact, its integrity, and
the sign/show/verify commands. Downstream A2/A3/A4 consume `design.md` § Contract Schema. Nothing in
this SPEC changes an existing gate: under A1 alone, Implementation Kickoff Approval behaves exactly
as today in both modes.

Tier L justification: a new verification package, a signing sub-package, CLI, config, and template
surfaces (~18 files including tests), an estimated >1000 LOC with tests, a schema that three later SPECs depend on (stability
matters more than speed), and three independent domains (schema/integrity, CLI/TTY interaction,
configuration/template). Milestones >= 3 and files >= 10, so run-phase routes to manager-lead.

## §B. Known Issues and Open Decisions for plan-audit

- **D1 — AC counter port vs subprocess.** Verify must not spawn processes (A2 calls it from hooks), so
  the published awk counter is ported to Go and pinned by a parity test over the full corpus plus one
  synthetic fixture per counter branch (prefix, `[RETIRED]`, `[REF]`, ambiguity, CRLF, BOM), with a
  positive control that the ambiguity fixture is ambiguous in both implementations. The test lives in
  `internal/spec` (beside the unexported extractor); `internal/contract` never imports `internal/spec`,
  so no import cycle arises.
- **D2 — `escalate_on` must be the full six.** Chosen so a contract cannot silently opt out of an
  escalation. Alternative (allow subsets) rejected as weakening A-Q3-style guarantees; revisit only if
  A2 finds a trigger inapplicable to some SPEC class.
- **D3 — Template defaults for `batch_sign` and `push_develop` are `false`.** The card fixes only
  `second_review: required` and `mode: guided`. `false` is chosen because a user project may have no
  integration branch; this repository sets both to `true` locally (asserted by AC-CONTRACT-020).
- **D4 — `review.second_model` value set is `codex | glm | none`.** `multi` (the `audit_multi` fan-out)
  is not included; adding it later is an additive schema change.
- **D5 — Human-presence checks raise the bar; they do not stop an agent.** Interactive terminal, typed
  token, and refusal on agent-environment markers (`CLAUDECODE`, `CLAUDE_CODE_SESSION_ID`). A pty
  wrapper and `unset` defeat them; the binding protection is the A3 precondition in spec.md §C.2 (a
  PreToolUse deny on `moai contract sign` from agent tool calls, owned by A2/t1235). No Codex marker is
  identifiable in the repository, so none ships (design.md § Agent-Environment Markers).
- **D6 — Unsigned-draft acceptance hash mismatch refuses signing** instead of overwriting, so a draft
  reviewed against older acceptance criteria cannot be signed silently. A **signed** contract whose
  acceptance changed is re-bound through `sign --resign` (REQ-CONTRACT-022) with the same human
  confirmation and an `old → new` summary.
- **D7 — Mission-validator projection deferred to A2.** Its only consumer is A2, and it needs a
  glob-to-prefix scope translation (design.md § Forward Note — Mission Projection). The withdrawn
  requirement's ID `REQ-CONTRACT-021` was reused for the agent-environment refusal so that the REQ
  sequence stays contiguous (MP-1) without a withdrawn-placeholder heading that lint would collect as a
  modality-less requirement.
- **D8 — Epic ordering and owners.** A1 → A2 → A3 (card t1236 carries t1235 as predecessor). A2
  (t1235) owns push-serialization enforcement and the PreToolUse deny on agent-invoked
  `moai contract sign`; both are hard preconditions for A3 activation. A4 (t1237) owns the
  "second review not performed → stop before push" rule and the performed record (spec.md §C.1).
- **D9 — Receipt signing path (lead-approved).** Non-human deciders sign non-interactively with a
  validated `kickoff-receipt.json`, only under `mode: contract`. A1 records `receipt.provenance: file`;
  autonomous Kickoff must not activate until `moai contract revoke` and moai-issued receipts exist (A3,
  spec.md §C.6). Residual risk: a Jev call cannot be proven (spec.md §H).

## §C. Pre-flight

- Confirm the published AC counter block is present exactly once in the manager-docs agent definition
  (the extraction helper already asserts this).
- Confirm `stdinIsTerminalFn` remains the TTY seam in `internal/cli`.
- Capture an LSP/lint baseline for `internal/config` and `internal/cli` before edits.

## §D. Constraints

- Template-First: the `autonomy` block is added to
  `internal/template/templates/.moai/config/sections/workflow.yaml` first, then mirrored locally, then
  `make build`. Template text carries no SPEC ID, card id, or date (neutrality guard).
- Verify: no subprocess, no network, no writes.
- All writes via `internal/atomicfile`.
- Tests use `t.TempDir()`; no test touches the real `.moai/specs/`.
- Scoped verification: `go test ./internal/contract/... ./internal/config/... ./internal/cli/... ./internal/spec/...`
  (the last because the parity test lives beside the corpus counter); CI runs the full suite.

## §E. Self-Verification (run-phase)

Each AC in `acceptance.md` carries its own command. The run is complete when every AC command passes
in the worktree and `moai spec lint` reports no error for this SPEC.

## §F. Milestones (ordered by decision-reversibility — highest-change-likelihood first)

### M1 — Schema types, strict decode, canonical digest (Priority High)

The data model A2-A4 depend on. Types for every section in `design.md` § Contract Schema, strict YAML
decode, the canonical digest, and golden fixtures (valid, unknown field, reordered sets, comment-only
change). Covers REQ-CONTRACT-001..004.

### M2 — Configuration keys and defaults (Priority High)

`workflow.autonomy` struct including `kickoff.{decider, jev_min_confidence, on_disagree}`, defaults,
fail-safe reader for every enum/range key, neutral template comments (no decision IDs, no
repository-specific statements), template and local YAML. Covers REQ-CONTRACT-015, REQ-CONTRACT-016.

### M3 — Validation rules and verify core (Priority High)

Action vocabulary (allowed / forbidden / unknown), escalation completeness, ownership and invariant
well-formedness (glob matcher; registry rule IDs passed in by the caller), non-empty actions,
reobserve, budget, plan-audit verdict, second-review and push-develop config coupling, reason-code
collection, derived sets (`effective_never`, `scratch`, `frozen_files`, `signable_contract_sha256`),
kickoff receipt decode, validator, and agreement rule. The core package imports neither `os/exec`, `net`, `internal/config`,
`internal/constitution` (the last two measured to pull in `net`), nor `internal/spec`. Covers
REQ-CONTRACT-006, 007, 008, 009, 017, 018, 020, 023, 025.

### M4 — Acceptance binding and AC counter port (Priority Medium)

LF/BOM-normalized hash, Go port of the published counter applied to normalized bytes, parity test
(corpus + per-branch fixtures) in `internal/spec`. Covers REQ-CONTRACT-005.

### M5 — Signing core (Priority Medium)

`internal/contract/sign`: agent-marker refusal, TTY refusal, typed confirmation, git identity and
HEAD seams, sign-time measurement and refusals, `--resign` re-binding with `old → new` summary, batch
mode with de-duplication and shared `batch_id`, atomic writes, Kickoff-neutral notice in both modes.
Receipt path (`--signer`/`--receipt`, `mode: contract` gate, provenance `file`). Closed sign refusal
codes. Covers REQ-CONTRACT-010..013, REQ-CONTRACT-019, REQ-CONTRACT-021, REQ-CONTRACT-022,
REQ-CONTRACT-024.

### M6 — CLI wiring and output (Priority Low — mechanical)

`moai contract sign|show|verify`, `--json`, exit codes; the CLI loads config and the constitution
registry and passes values into the core. Covers REQ-CONTRACT-014 (and the CLI half of 009).

## §G. Risks

| Risk | Impact | Mitigation |
|---|---|---|
| AC counter port drifts from the published awk counter | Verify reports `ac_count_mismatch` on untouched files, or misses a real change | Parity test (corpus + per-branch fixtures + ambiguity positive control) in `./internal/spec`; the SHA-256 is the primary tamper authority, the count is a secondary cross-check |
| Schema churn after A2-A4 start | Rework in three SPECs | `schema_version`; additive changes only; changes land as amendments to `design.md` § Contract Schema |
| An agent obtains a TTY with a one-line pty wrapper and unsets the markers | Agent-produced signature | A1 only raises the bar (spec.md §C.2); hard precondition for A3: a PreToolUse deny on `moai contract sign` from agent tool calls must exist before the signature replaces Kickoff |
| A receipt is hand- or agent-written | Autonomous Kickoff on a forged approval | A1 records `provenance: file`; §C.6 forbids activation until A3's moai-issued receipts and `revoke` exist; residual risk stated in spec.md §H |
| A3 lands before A2's enforcement | Signed `push-develop` authorizes unserialized pushes / pushes without second review | Binding ordering constraint in spec.md §C.1; until then the notice (REQ-CONTRACT-019) prints in both modes and Kickoff stays |
| Line-ending differences across platforms | Spurious hash mismatch on windows checkouts | CRLF→LF and BOM normalization before hashing; cross-platform CI |
| Operators read the signature as authorizing Kickoff skip before A3 lands | Confusion | REQ-CONTRACT-019 notice in both modes; A1 changes no gate |

## §H. Cross-References

- `design.md` § Contract Schema, § Verify Reason Codes, § Signing Flow, § Agent-Environment Markers, § Glob Semantics, § Package Layout, § Forward Note — Mission Projection.
- `research.md` § Reuse Analysis.
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier.
