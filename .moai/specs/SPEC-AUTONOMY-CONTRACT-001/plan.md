# plan.md — SPEC-AUTONOMY-CONTRACT-001

## §A. Context

A1 of the contract-autonomy epic (card t1234). It produces the contract artifact, its integrity, and
the sign/show/verify commands. Downstream A2/A3/A4 consume `design.md` § Contract Schema. Nothing in
this SPEC changes an existing gate: under A1 alone, Implementation Kickoff Approval behaves exactly
as today in both modes.

Tier L justification: a new package plus CLI, config, and template surfaces (~18 files including
tests), an estimated >1000 LOC with tests, a schema that three later SPECs depend on (stability
matters more than speed), and three independent domains (schema/integrity, CLI/TTY interaction,
configuration/template). Milestones >= 3 and files >= 10, so run-phase routes to manager-lead.

## §B. Known Issues and Open Decisions for plan-audit

- **D1 — AC counter port vs subprocess.** Verify must not spawn processes (A2 calls it from hooks), so
  the published awk counter is ported to Go and pinned by a full-corpus parity test. Two
  implementations of one measurement is a known hazard; the parity test is the mitigation.
- **D2 — `escalate_on` must be the full six.** Chosen so a contract cannot silently opt out of an
  escalation. Alternative (allow subsets) rejected as weakening A-Q3-style guarantees; revisit only if
  A2 finds a trigger inapplicable to some SPEC class.
- **D3 — Template defaults for `batch_sign` and `push_develop` are `false`.** The card fixes only
  `second_review: required` and `mode: guided`. `false` is chosen because a user project may have no
  integration branch; this repository sets both to `true` locally.
- **D4 — `review.second_model` value set is `codex | glm | none`.** `multi` (the `audit_multi` fan-out)
  is not included; adding it later is an additive schema change.
- **D5 — Human presence = interactive terminal + typed token.** Not an identity proof (spec.md §B).
- **D6 — Draft acceptance hash mismatch refuses signing** instead of overwriting, so a contract
  reviewed against older acceptance criteria cannot be signed silently.

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

`workflow.autonomy` struct, defaults, fail-safe reader for invalid `mode` / `second_review`, template
and local YAML. Covers REQ-CONTRACT-015, REQ-CONTRACT-016.

### M3 — Validation rules and verify core (Priority High)

Action vocabulary (allowed / forbidden / unknown), escalation completeness, ownership and invariant
well-formedness (constitution registry resolution), reobserve, budget, plan-audit verdict,
second-review and push-develop config coupling, reason-code collection. Covers REQ-CONTRACT-006,
007, 008, 017, 018, 020.

### M4 — Acceptance binding and AC counter port (Priority Medium)

LF/BOM-normalized hash, Go port of the published counter, full-corpus parity test. Covers
REQ-CONTRACT-005.

### M5 — Signing core (Priority Medium)

TTY refusal, typed confirmation, git identity and HEAD seams, sign-time measurement and refusals,
`--resign`, mission projection + `SealMissionContract`, batch mode with shared `batch_id`, atomic
writes. Covers REQ-CONTRACT-010..013, REQ-CONTRACT-019, REQ-CONTRACT-021.

### M6 — CLI wiring and output (Priority Low — mechanical)

`moai contract sign|show|verify`, `--json`, exit codes, guided-mode notice. Covers REQ-CONTRACT-009,
REQ-CONTRACT-014.

## §G. Risks

| Risk | Impact | Mitigation |
|---|---|---|
| AC counter port drifts from the published awk counter | Verify reports `ac_count_mismatch` on untouched files, or misses a real change | Full-corpus parity test in `./internal/spec` or `./internal/contract`; the SHA-256 is the primary tamper authority, the count is a secondary cross-check |
| Schema churn after A2-A4 start | Rework in three SPECs | `schema_version`; additive changes only; changes land as amendments to `design.md` § Contract Schema |
| An agent obtains a TTY (e.g. a tmux pane driven by an agent) | Agent-produced signature | Out of scope for A1 (documented); A2 may add a PreToolUse deny on `moai contract sign` |
| Line-ending differences across platforms | Spurious hash mismatch on windows checkouts | CRLF→LF and BOM normalization before hashing; cross-platform CI |
| Operators read the signature as authorizing Kickoff skip before A3 lands | Confusion | REQ-CONTRACT-019 notice; A1 changes no gate |

## §H. Cross-References

- `design.md` § Contract Schema, § Verify Reason Codes, § Signing Flow, § Mission-Contract Projection.
- `research.md` § Reuse Analysis.
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier.
