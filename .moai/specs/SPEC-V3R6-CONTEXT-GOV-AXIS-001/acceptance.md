# acceptance.md — SPEC-V3R6-CONTEXT-GOV-AXIS-001

> **Traceability**: each AC maps to one or more REQ-CGA-* in spec.md §C. Given-When-Then scenarios below.
> **Severity convention**: MUST-PASS (blocks close) / SHOULD-PASS (closes with debt) / INFO (observation only).

---

## §D. AC Matrix

| AC ID | Severity | REQ | Trace |
|-------|----------|-----|-------|
| AC-CGA-001 | MUST-PASS | REQ-CGA-001 | Observer records eager-vs-on-demand weight per turn |
| AC-CGA-002 | MUST-PASS | REQ-CGA-002 | Schema extension backward-compatible |
| AC-CGA-003 | MUST-PASS | REQ-CGA-003 | Fail-open on recording failure |
| AC-CGA-004 | MUST-PASS | REQ-CGA-004 | Drift alarm documented with monotonic → Tier-2 rule |
| AC-CGA-005 | MUST-PASS | REQ-CGA-005 | Drift alarm cites book2 ch8.3 + diag-05 |
| AC-CGA-006 | MUST-PASS | REQ-CGA-006 | Tier system (1-4) unchanged; additive signal |
| AC-CGA-007 | MUST-PASS | (lint) | spec-lint 0 findings |

---

## §D.1 AC-CGA-001 — Observer records eager-vs-on-demand weight per turn

**Given** the Layer A observer hook pipeline is registered (4 wrappers in `.claude/hooks/moai/handle-harness-observe*.sh` backing `moai hook harness-observe*`)
**When** a turn fires any observed event (`user_prompt`, `session_stop`, or `subagent_stop`)
**Then** a new line is appended to `.moai/harness/usage-log.jsonl` carrying:
- all existing fields (timestamp, event_type, subject, context_hash, tier_increment, schema_version — common to both legacy `v1` and current `v2` lines)
- a new `eager_context_weight` field (integer, estimated tokens/bytes of CLAUDE.md + auto-loaded rules + output-style moai.md + MEMORY.md)
- a new `on_demand_context_weight` field (integer, estimated tokens/bytes of invoked skill bodies this turn)
- a new `weight_unit` field (`"tokens"` or `"bytes"`)

**Evidence**: run-phase M3 produces a sample line from a live turn with all three new fields populated (non-sentinel).

---

## §D.2 AC-CGA-002 — Schema extension backward-compatible

**Given** a fixture of pre-extension `usage-log.jsonl` lines (captured at plan-phase 2026-06-18, carrying BOTH legacy `schema_version: "v1"` lines AND legacy `schema_version: "v2"` lines — neither has weight fields, since both predate this SPEC; the live `.moai/harness/usage-log.jsonl` carries 508 `v1` lines + 100 `v2` lines at discovery)
**When** the post-extension reader/parser processes those old lines
**Then**:
- no parse error is raised on EITHER legacy `v1` OR legacy `v2` lines
- old fields are preserved verbatim on both legacy schemas
- on the legacy lines, the new weight fields are **present-as-sentinel** — the reader treats the weight fields as `null` (tokens-unset) / `0` (integer-default) and does NOT crash. The reader distinguishes "pre-SPEC binary, never measured" (`schema_version ∈ {"v1", "v2"}` — both legacy, both weight-absent, sentinel) from "new binary, estimation skipped" (`schema_version == "v2.1"`, sentinel) by branching on `schema_version` per REQ-CGA-002.

Pinned to ONE behavior: **sentinel-on-old-lines** (`null` for the nullable field, `0` for the integer, as the Go zero-value / `omitempty`-absent rendering produces). The reader MUST parse BOTH legacy `v1` AND legacy `v2` lines without crashing — that is the non-negotiable binary gate. The fixture MUST include BOTH legacy schemas so the test exercises the dual-legacy parsing surface.

**Evidence**: run-phase M1/M3 unit test `TestParseOldLogLinesNoCrash` (or equivalent) passes against the plan-phase-captured fixture, asserting both (a) no parse error on `v1` lines AND `v2` lines, and (b) weight fields resolve to the sentinel when read back on BOTH legacy schemas.

---

## §D.3 AC-CGA-003 — Fail-open on recording failure

**Given** a simulated weight-measurement failure (e.g. `.moai/harness/usage-log.jsonl` temporarily unreadable, or stat error on a source file, or estimation exception)
**When** the observer hook attempts to record and the failure occurs
**Then**:
- a warning is emitted to `$MOAI_HOOK_STDERR_LOG` (default `$HOME/.moai/logs/hook-stderr.log`)
- the hook exits with code **0** (NOT 2)
- the turn is NOT blocked
- the death-spiral avoidance property (P0 SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001) is preserved

**Evidence**: run-phase M3 produces the stderr-log warning line + confirms exit code 0 under the simulated failure. No `exit 2` path exists in the weight-recording code.

---

## §D.4 AC-CGA-004 — Drift alarm documented with monotonic → Tier-2 rule

**Given** `.moai/docs/harness-delivery-strategy.md` has been extended with a "Context-Governance Axis" section
**When** a reader reviews that section
**Then** the section defines:
- a monotonic-growth rule: eagerly-loaded weight grows monotonically across **N** consecutive sessions, where **N** is a named constant fixed within the bounded range **N ∈ [3, 5]** (REQ-CGA-004; the section names the chosen value, recommend N = 3)
- the trigger condition: growth occurs WITHOUT a corresponding skill-budget reduction (rules added but not demoted to on-demand skills)
- the action: a **Tier-2 harness-learning proposal** is surfaced ("eager context grew X% — candidate rules to demote to on-demand skills")
- the threshold is operationalized from book2 ch8.3's symptom test (tokens burn fast, quality flat)

**Evidence**: the section exists in the file (run-phase M2 commit), grep-able for "Tier-2", "monotonic", "demote", and names a concrete N value within [3,5].

---

## §D.5 AC-CGA-005 — Drift alarm cites book2 ch8.3 + diag-05

**Given** the "Context-Governance Axis" section in harness-delivery-strategy.md
**When** a reader reviews the doctrinal basis
**Then** the section explicitly references:
- **book2 ch8.3** (signal dilution — "the model sees more, but is not necessarily clearer about which working semantics matter next")
- **diag-05** (three context-governance paths: budgeted working memory / structured context units / prompt-stacking foil)

**Evidence**: the section contains the literal citation strings "book2 ch8.3" and "diag-05" (grep-able).

---

## §D.6 AC-CGA-006 — Tier system (1-4) unchanged; additive signal

**Given** the existing harness-learning Tier system (Tiers 1-4) is in force prior to this SPEC
**When** the drift alarm fires
**Then**:
- the proposal is routed to **Tier-2** (the existing tier for proposals)
- no Tier is added, removed, or renumbered
- the alarm is documented as an additive **drift SIGNAL** feeding Tier-2, not a tier-system rewrite

**Evidence**: the "Context-Governance Axis" section explicitly states the Tier system is unchanged and this is an additive signal; the Tier definitions elsewhere in harness docs are byte-for-byte unchanged (diff confirms only the new section was added).

---

## §D.7 AC-CGA-007 — spec-lint 0 findings

**Given** the new SPEC directory `.moai/specs/SPEC-V3R6-CONTEXT-GOV-AXIS-001/` with spec.md + plan.md + acceptance.md + progress.md
**When** `moai spec lint` (or `moai spec audit --json`) is run
**Then** the SPEC produces **0 MUST-FIX findings** (0 findings total for a clean close; warnings acceptable only if explicitly justified as documented debt).

**Evidence**: run-phase M3 (or sync-phase) produces the `moai spec lint` output showing 0 findings for this SPEC ID.

---

## §I. Edge Cases

| ID | Edge case | Expected behavior |
|----|-----------|-------------------|
| EC-1 | A source file (e.g. a rule .md) is deleted between turns | Weight estimation skips the missing file; total eager weight decreases monotonically-correctly; no crash (fail-open). |
| EC-2 | `MEMORY.md` exceeds the 200-line/25KB index cap and is truncated | Weight reflects the truncated (loaded) portion, not the full file — consistent with what the model actually sees. |
| EC-3 | `moai.md` output-style is reloaded mid-session with a different LOC count | Per-turn weight reflects the count at that turn; drift alarm uses per-turn snapshots, not a session-average. |
| EC-4 | No skills are invoked in a turn | `on_demand_context_weight` = 0; valid line, not an error. |
| EC-5 | The Go binary is the `~/go/bin/moai` fallback (not in PATH) | Weight recording still works (the Go binary is the recording owner regardless of which wrapper path found it). |
| EC-6 | `usage-log.jsonl` does not exist yet (fresh project) | Hook creates it; first line carries the new fields. No dependency on pre-existing v1 lines. |

---

## §J. Quality Gates (Definition of Done)

- [ ] All 7 AC (AC-CGA-001..007) PASS with evidence
- [ ] `moai spec lint` 0 findings
- [ ] Unit tests pass: old-line parsing, new-line fields, fail-open on failure
- [ ] `harness-delivery-strategy.md` new section grep-able for "Tier-2", "monotonic", "book2 ch8.3", "diag-05"
- [ ] No `exit 2` path in the weight-recording code (grep-verified)
- [ ] Tier definitions byte-for-byte unchanged (diff-verified)
- [ ] 4-phase lifecycle: plan (this phase) → run (M1-M3) → sync → Mx close
