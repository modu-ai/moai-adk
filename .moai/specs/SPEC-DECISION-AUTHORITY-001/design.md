---
id: SPEC-DECISION-AUTHORITY-001
title: "Design — decision-gate axis, index format, gate integration, verification design"
version: "0.1.0"
created: 2026-09-13
---

# design.md — SPEC-DECISION-AUTHORITY-001

Design constraints inherited verbatim and binding on every section below:

1. **Detect → Explain → Ask, but never decide.**
2. **An LLM "best practice" is not a policy.**
3. **When uncertain, escalate. Never downgrade.**

## §1 Layer map

| Layer | Change | Owner (run-phase) |
|---|---|---|
| Config | `interview.decision_gate` axis (`internal/config`) | M1 |
| Doctrine — flow owner | manager-spec body + clarity-interview (index authoring, labels, register, escalation) | M2 |
| Doctrine — gate | spec-assembly (kickoff presentation, verdict recording, zero-rows clause) | M3 |
| Distribution | template mirrors + `make build` | M4 |
| Verification | behavioral repro + sweeps | M5 |

No Go runtime change beyond `internal/config`. No hook. No pipeline stage. The gate
presentation is doctrine text — the orchestrator reads `decision-index.md` (or its absence)
and composes accordingly, exactly as it already does for the plan HTML report.

## §2 Config-axis design (M1)

```go
// internal/config/types.go — InterviewConfig gains:
DecisionGate string `yaml:"decision_gate"`

// Resolution — mirrors ResolvedRecommendationMode():
func (c InterviewConfig) ResolvedDecisionGate() string {
    if c.DecisionGate == "on" { return "on" }
    return "off"   // absent, empty, "off", or ANY unrecognized value
}
```

- Default: `DecisionGate: "off"` in `defaultInterviewConfig()`.
- Unknown values resolve to `off` (fail-safe: an unrecognized value can never activate the
  gate), and the raw value stays on the struct so the unrecognized setting is observable —
  the same property `ResolvedRecommendationMode()` carries.
- **Orthogonality (REQ-DA-018):** `ResolvedDecisionGate()` reads only `DecisionGate`;
  `ResolvedRecommendationMode()` reads only `RecommendationMode`. Neither resolver, neither
  default, and neither test references the other axis. A test asserts the independence
  explicitly (set one, leave the other absent, assert both resolve independently).
- Precedent compliance: identical wiring to `RecommendationMode`
  (`types.go:1360`, `defaults.go` default, dedicated test file) — measured this tree.

## §3 Index format design (M2)

`decision-index.md` — a plan-phase artifact, stateless (no `status:` field), authored only
while the gate is `on`. Fixed row shape (spec.md §B) so two authors produce the same file:

- **Q heading** — the decision as a question (the Detect step's output).
- **Label** — one of exactly `DECIDED` / `POLICY-COVERED` / `EVIDENCE-NEEDED` / `FOUNDER`
  (D3; no second vocabulary ever).
- **Authority anchor** — required for `DECIDED`/`POLICY-COVERED` only: `<file> §<section>`
  that exists in the committed tree. The anchor is the Explain step's core: it names WHERE the
  answer already lives.
- **Why unresolved** — one line: what the documents do not answer.
- **Operator verdict** — empty at authoring; filled at kickoff with `DECIDE` /
  `NEED_ANALYSIS` / `NEED_EVIDENCE` / `DEFER`.

Label routing algorithm (manager-spec body text, M2):

1. A prior completed SPEC's HISTORY/`## Amendments` row decides the identical question under
   identical conditions → `DECIDED` (anchor: that SPEC + row).
2. An explicit operator setting in `.moai/config/sections/*.yaml`, or a constitution clause,
   covers the question as written → `POLICY-COVERED` (anchor: file + section).
3. The decision needs data/measurements that do not exist yet → `EVIDENCE-NEEDED`.
4. Otherwise → `FOUNDER`. **An anchor that cannot be verified never produces 1 or 2** —
   condition 3 (escalate, never downgrade) makes `FOUNDER` the default for every doubt, which
   is the conservative direction: the cost of a false FOUNDER is one question the operator
   answers; the cost of a false DECIDED is a decision nobody made.

**Why the anchor is file+section, not a quote.** A quote rots with edits and cannot be
re-verified cheaply; a file+section address lets the operator (or a later session) resolve the
authority with one `git show`/Read, keeping the index honest without duplicating the register
into every SPEC.

## §4 Gate integration design (M3)

Composition point: the Implementation Kickoff Approval turn, beside the existing plan HTML
enrichment (spec-assembly Step 2.3.3a precedent):

- **Where** `decision_gate` resolves `on` AND `decision-index.md` exists → the orchestrator
  includes the index rows as additive prose context in the SAME turn the gate's
  `AskUserQuestion` fires. The gate's option structure, its `[HARD]` mandatory/score-
  independent clause, and its label clause are untouched (AC-DA-011 pins this).
- **Where** the index is absent or unreadable → skip silently (fail-open, exactly like the
  HTML report). No warning that blocks, no new gate, no retry.
- **Verdict recording**: after the gate, verdicts are written back per row. Product-level
  verdicts additionally surface the `product.md` reconciliation question with the ownership
  boundary named — the reconcile edit routes to manager-docs or to the operator's hand
  (§E.3 OD-2); manager-spec never writes `.moai/project/**`.
- **Zero rows** (or no index under `on`): the gate fires unchanged. Zero judgment points is a
  finding about this pass, not an approval (REQ-DA-019).
- **Pull composition**: the index rows never carry a preferred answer in either mode
  (REQ-DA-017); the gate question itself follows the landed `recommendation_mode` convention
  unchanged (label withheld under `pull`, per the `:217` clause SPEC-JFM already conditioned).

## §5 Verification design (M5)

Two axes, mirroring the sibling SPEC's measured approach:

- **Static axis (release-blocking)** — the acceptance.md sweeps: token-presence greps per
  surface with swept-count floors (§1.1 empty-sweep visibility: a zero-hit sweep is recorded
  with its exit code and is never read as a pass). Every RED-now cell is a single-invocation
  read-only command with verbatim stdout + exit code + tree SHA `62fbd6baf`, captured in the
  §E evidence ledger before implementation.
- **Behavioral axis (regression-guard, M5)** — two throwaway plan-phase passes under `/tmp`:
  off-mode asserts no index is created and the flow composes unchanged; on-mode asserts the
  index appears with the four labels and an unverifiable anchor routes to `FOUNDER`. The
  off-mode half is green-at-arrival by design — at base the flow cannot create an index, so
  no behavioral RED is producible at plan time — and is therefore classified regression-guard
  (AC-DA-019, the undecidable disposition of verification-completeness §2.1), with the M5
  repro as its completing act.

Regression-guard ACs (gate-count preserve, auditor-unchanged incl. byte-identity, off-mode
behavioral repro) are green-at-arrival by design; they are classified regression-guard — not
release-blocking — and pin `62fbd6baf`, per verification-completeness §2.1's undecidable
disposition.

## §6 Rejected alternatives (recorded for traceability)

- **Auditor-authored index** — rejected: report-contract change, no observable effect on the
  common paths (SPEC-JFM §A.3 measurement binds); deferred explicitly (§E.2).
- **Second human gate after plan approval** — rejected: the kickoff gate already sits at the
  decision point; a second gate is a pipeline-structure change and doubles operator cost
  without adding a decision surface.
- **Runtime PreToolUse observer for index authorship** — rejected: the index is authored by
  manager-spec as a file; file-existence and content sweeps verify it deterministically. An
  observer would gate nothing that a static check does not already reach (contrast SPEC-JFM,
  where the observed behavior is an ephemeral tool-call payload no file carries).
- **Embedding analysis in rows** — rejected: violates condition 1 (never decide) and
  recreates anchoring inside the artifact the pull convention just cleaned.
- **`DECIDED` rows citing `.moai/reports/**` evidence** — rejected on the measured criterion
  (research.md §7): untracked material is not authority; only committed artifacts are.
