---
id: SPEC-V3R6-CONTEXT-GOV-AXIS-001
title: "Context-Governance Axis — eager-vs-on-demand weight recording + drift alarm"
version: "0.1.0"
status: implemented
created: 2026-06-18
updated: 2026-06-19
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/hooks/moai"
lifecycle: spec-anchored
era: V3R6
tags: "context, governance, observer, harness"
---

# SPEC-V3R6-CONTEXT-GOV-AXIS-001

> **Source**: applied analysis of `github.com/wquguru/harness-books` — book2 ch8.3 (signal dilution), diag-05 (three context-governance paths), appendix B.2 (compact/truncation/recovery trio invariant).
> **Sprint**: 15 (harness-books application cohort, priority P2b).
> **Tier**: M (standard — spec.md + plan.md + acceptance.md + progress.md §E skeleton).
> **Era**: V3R6 (subject to 4-phase lifecycle drift detection).

---

## §A. Problem Statement

As the MoAI-ADK harness matures, the path of least resistance is to keep stuffing more bootstrap / skill / identity / rule text into the eagerly-loaded context and lean on truncation after the fact ("inject first, rescue later"). The deeper cost — per book2 ch8.3 — is **signal dilution**: the model sees more, but is not necessarily clearer about which working semantics matter next.

moai-adk has token budgets for **skills** (skillListingBudgetFraction ~1%, compaction budget ~25K, progressive disclosure L1/L2/L3) but **NO equivalent budget or audit for the EAGERLY-loaded context**: `CLAUDE.md` + all auto-loaded `.claude/rules/moai/*.md` (~60 files) + output-style (`moai.md` 773 LOC) + auto-memory index (`MEMORY.md`). The existing `session-handoff.md` thresholds (1M=50% / 200K=90%) are **REACTIVE** (when to `/clear`), not **PREVENTIVE** (how much we should inject in the first place).

moai-adk cannot mechanically answer book2's diagnostic question: *"have we drifted from budgeted-working-memory toward inject-first?"* — because the eager-vs-on-demand context weight split is never recorded.

### A.1 The drift-risk surface

A LARGE control plane (`moai.md` 773 LOC + `CLAUDE.md` + ~60 rule files + ~50 skills) is exactly the surface where the inject-first drift accumulates. Rules get added over time; few get demoted to on-demand skills. Without a recording, the drift is invisible until the symptom test fires: *tokens burn fast, quality doesn't climb as context fattens.*

---

## §B. Scope

### B.1 In scope

1. **Recording**: extend the Layer A Observer hook pipeline to capture, per turn, the **eager-vs-on-demand context weight split** and append it to `.moai/harness/usage-log.jsonl`.
2. **Schema**: extend `usage-log.jsonl` (currently schema `v2`, per `internal/harness/types.go:16` `const LogSchemaVersion = "v2"` — already bumped from `v1` by SPEC-HARNESS-OUTCOME-CAPTURE-001 REQ-OC-010) with the new weight fields in a **backward-compatible** way (existing fields preserved; new fields additive).
3. **Drift alarm**: add a "Context-Governance Axis" section to `.moai/docs/harness-delivery-strategy.md` defining the monotonic-growth → Tier-2 demote proposal rule, cross-referencing book2 ch8.3 + diag-05.

### B.2 Source citations

- **book2 ch8.3** — signal dilution: *"the model sees more, but is not necessarily clearer about which working semantics matter next."*
- **diag-05** — three context-governance paths: (a) budgeted working memory (healthy) / (b) structured context units / (c) prompt-stacking (OpenClaw foil).
- **appendix B.2** — invariant: *"assert long session has compact/truncation/recovery trio."* (cross-reference only; this SPEC does not alter the trio).
- **symptom test** (book2 ch8.3): *tokens burn fast, quality doesn't climb as context fattens.* → operationalized as the mechanical drift gate in §B.1.3.

---

## §C. Requirements (GEARS)

> Subject convention: `<subject>` is generalized per GEARS (any noun). `<subject>` = "the observer hook pipeline", "the usage-log schema", "the drift alarm", as appropriate per requirement.

### REQ-CGA-001 — Eager-vs-on-demand weight recording

The observer hook pipeline **shall** record, per turn, an estimated **eager context weight** (comprising `CLAUDE.md` + auto-loaded `.claude/rules/moai/*.md` + output-style `moai.md` + auto-memory index `MEMORY.md`, measured in estimated tokens or bytes) AND an estimated **on-demand context weight** (invoked skills' bodies loaded this turn), appending both to `.moai/harness/usage-log.jsonl`.

### REQ-CGA-002 — Backward-compatible schema extension

**Where** the existing `usage-log.jsonl` schema is `v2` (verified baseline: `internal/harness/types.go:16` `const LogSchemaVersion = "v2"`, already bumped from `v1` by SPEC-HARNESS-OUTCOME-CAPTURE-001 REQ-OC-010; the live `.moai/harness/usage-log.jsonl` carries a mix of legacy `"v1"` and current `"v2"` lines — 508 `v1` lines + 100 `v2` lines at plan-phase 2026-06-18), the usage-log schema **shall** be extended with new eager-vs-on-demand weight fields such that existing fields are preserved, existing readers continue to parse old lines without error, and new fields default to a sentinel on lines written by older binaries. The `schema_version` string literal **shall** bump from `"v2"` to `"v2.1"` on lines written by the new binary (NOT `v1` → `v1.1` — that would be a regression below the already-current `v2` baseline), so a reader of an old line can branch on `schema_version` to distinguish two distinct cases: (a) `schema_version ∈ {"v1", "v2"}` → the line predates this SPEC (written by a binary that has no concept of weight fields — `v1` is the truly-original schema, `v2` is the SPEC-HARNESS-OUTCOME-CAPTURE-001 schema that added outcome-capture fields but still has no weight fields; both are weight-absent legacy); (b) `schema_version == "v2.1"` with weight fields set to the sentinel (`null`/`0`) → the line was written by the new binary but weight estimation was skipped (fail-open path fired). Without the bump these two cases are indistinguishable and a reader cannot tell "pre-SPEC binary, never measured" from "new binary, chose not to record". Old `"v1"` and `"v2"` lines parse unchanged under the extended reader (additive fields only).

### REQ-CGA-003 — Fail-open recording

**When** a weight-measurement failure is detected (file unreadable, stat error, token estimation exception), the observer hook pipeline **shall** emit a warn-and-continue log line to the hook stderr log (`$MOAI_HOOK_STDERR_LOG`) and **shall not** block the turn (exit 0, never exit 2). The drift-detection death-spiral avoidance property (per the P0 SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001 companion) **shall not** be violated.

### REQ-CGA-004 — Context-Governance Axis drift alarm

The drift alarm, documented in `.moai/docs/harness-delivery-strategy.md` "Context-Governance Axis" section, **shall** define a rule that surfaces a **Tier-2 harness-learning proposal** when eagerly-loaded weight grows monotonically across **N** consecutive sessions WITHOUT a corresponding skill-budget reduction (i.e. rules were added but not demoted to on-demand skills). The session-count threshold **N** **shall** be bounded to the documented range **N ∈ [3, 5]** (minimum 3, maximum 5), with the concrete value fixed inside that range in the doctrine section at run-phase M2. Bounding N to ≥ 3 ensures the alarm cannot fire trivially at N=1 (a single session's growth is noise, not drift); bounding to ≤ 5 keeps the alarm responsive (beyond 5 sessions of silent growth the signal has already cost real tokens). The run-phase M2 author records the chosen N value as a named constant at the top of the "Context-Governance Axis" section so it is visible to a reviewer and adjustable without re-spec'ing.

### REQ-CGA-005 — Source citation in drift alarm

The "Context-Governance Axis" section **shall** explicitly reference **book2 ch8.3** (signal dilution) and **diag-05** (three context-governance paths: budgeted working memory / structured context units / prompt-stacking foil) as the doctrinal basis for the drift alarm, so the rule is traceable to its external source rather than appearing as an arbitrary threshold.

### REQ-CGA-006 — Tier-system non-modification

**Where** the existing harness-learning Tier system (Tiers 1-4) is in force, the Context-Governance Axis drift alarm **shall** feed Tier-2 proposals as a new **drift SIGNAL** and **shall not** alter, renumber, or replace the existing Tier definitions. (Additive signal, not a tier-system rewrite.)

---

## §D. Constraints

- **DISCOVERY-FIRST** (per plan.md pre-flight): concrete file paths and line numbers cited in this SPEC are grounded in the discovery performed at plan-phase — the 4 observer hook wrapper filenames, the Go-side recording logic location, and the current `usage-log.jsonl` schema. The run-phase MUST re-verify these paths before editing (paths drift).
- **Fail-open is mandatory** (REQ-CGA-003). The observer hook must never block a turn even if weight-recording fails. Cross-reference: death-spiral avoidance from P0 SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001.
- **No tier-system rewrite** (REQ-CGA-006). This SPEC adds a drift SIGNAL feeding Tier-2 proposals; it does not restructure the Tier system.
- **Maintainer-local doc**: `.moai/docs/harness-delivery-strategy.md` is a maintainer-local strategy document (NOT a template-managed file per §15 language-neutrality / §25 template-isolation). The new section is therefore allowed to carry maintainer-specific context (Sprint 15 cohort, harness-books provenance).
- **Estimated weight is acceptable** (not exact). Token estimation may use a bytes/4 heuristic or equivalent. Exactness is NOT a requirement; monotonic-growth detection is.

---

## §E. Acceptance Criteria Matrix

> Full Given-When-Then scenarios live in `acceptance.md`. This matrix is the traceability summary.

| AC ID | REQ | Summary |
|-------|-----|---------|
| AC-CGA-001 | REQ-CGA-001 | Observer hook records eager-vs-on-demand context weight per turn into usage-log.jsonl |
| AC-CGA-002 | REQ-CGA-002 | Schema extension is backward-compatible (existing fields preserved, old lines parse) |
| AC-CGA-003 | REQ-CGA-003 | Observer hook remains fail-open (warn-and-continue on recording failure, never exit 2) |
| AC-CGA-004 | REQ-CGA-004 | harness-delivery-strategy.md documents Context-Governance Axis drift alarm with monotonic-growth → Tier-2 rule |
| AC-CGA-005 | REQ-CGA-005 | Drift alarm references book2 ch8.3 signal-dilution + diag-05 |
| AC-CGA-006 | REQ-CGA-006 | Existing harness-learning Tier system (1-4) unchanged; alarm is additive Tier-2 signal |
| AC-CGA-007 | (lint) | spec-lint clean (0 findings) on the new SPEC directory |

---

## §F. Discovery Artifacts (plan-phase ground truth)

Captured at plan-phase; run-phase MUST re-verify before editing (DISCOVERY-FIRST constraint).

### F.1 Layer A Observer hook wrapper files (4 discovered)

All four are thin bash wrappers that forward stdin JSON to `moai hook harness-observe*` and `exec` the Go binary. The actual recording logic lives in the **Go binary**, not in the `.sh` wrappers.

1. `.claude/hooks/moai/handle-harness-observe.sh` (1244 bytes) → `moai hook harness-observe`
2. `.claude/hooks/moai/handle-harness-observe-stop.sh` (1269 bytes) → `moai hook harness-observe-stop`
3. `.claude/hooks/moai/handle-harness-observe-subagent-stop.sh` (1314 bytes) → `moai hook harness-observe-subagent-stop`
4. `.claude/hooks/moai/handle-harness-observe-user-prompt-submit.sh` (1339 bytes) → `moai hook harness-observe-user-prompt-submit`

**Implication for run-phase**: the weight-recording logic is a **Go-side** change across three verified files (Glob/Read-confirmed at plan-phase 2026-06-18; run-phase MUST re-verify before editing per the DISCOVERY-FIRST constraint — paths drift):

- **`internal/harness/observer.go`** — the recorder. `RecordEvent` (L53, the original minimal recorder) and `RecordExtendedEvent` (L103, the full-Event entry point already used by the Stop/SubagentStop/UserPromptSubmit handlers). The new eager/on-demand weight population most naturally lands in `RecordExtendedEvent` or a sibling, since `RecordEvent` takes only `(eventType, subject, contextHash)` and cannot carry the new fields without a signature change.
- **`internal/cli/hook.go`** — the CLI dispatch. The four `runHarnessObserve*` handlers (L601 PostToolUse, L659 Stop, L741 SubagentStop, L895 UserPromptSubmit) wire hook stdin JSON → `RecordEvent`/`RecordExtendedEvent`. Weight estimation is invoked here (read source files, apply bytes/4 heuristic, populate the new Event fields) before handing the Event to the recorder.
- **`internal/harness/types.go`** — the schema-extension point. The `Event` struct (L65) is where the new eager/on-demand weight fields are added as additive `omitempty`-tagged fields, mirroring how the existing Stop/SubagentStop optional fields were added (REQ-HRN-OBS-003/005).

The `.sh` wrappers themselves need NOT be edited — they are generated forwarders (`# Generated by moai-adk` header, at risk of `moai update` overwrite) and are NOT the injection point for weight estimation.

### F.2 Current usage-log.jsonl schema (v2 — verified baseline)

Verified at plan-phase 2026-06-18 against `internal/harness/types.go:16` (`const LogSchemaVersion = "v2"`) and the live `.moai/harness/usage-log.jsonl` (508 lines `"v1"` + 100 lines `"v2"`). The `v1` → `v2` bump was performed by SPEC-HARNESS-OUTCOME-CAPTURE-001 REQ-OC-010 (comment at `internal/harness/types.go:12-15`). The `SchemaVersion` field is wired into the `Event` struct at `internal/harness/types.go:81-82`.

The log carries a MIX of legacy `v1` lines and current `v2` lines:

```json
{"timestamp":"...","event_type":"user_prompt","subject":"","context_hash":"","tier_increment":0,"schema_version":"v2","prompt_hash":"...","prompt_len":N,"prompt_lang":"en"}
{"timestamp":"...","event_type":"session_stop","subject":"","context_hash":"","tier_increment":0,"schema_version":"v2","last_assistant_message_hash":"...","last_assistant_message_len":N}
{"timestamp":"...","event_type":"subagent_stop","subject":"unknown","context_hash":"","tier_increment":0,"schema_version":"v2","agent_id":"..."}
```

Common fields across both `v1` and `v2`: `timestamp`, `event_type`, `subject`, `context_hash`, `tier_increment`, `schema_version` (`"v1"` on legacy lines, `"v2"` on current lines).

**Important for REQ-CGA-002**: BOTH `v1` and `v2` lines predate this SPEC and carry NO weight fields. The bump this SPEC introduces is `"v2"` → `"v2.1"` (NOT `v1` → `v1.1`). A post-extension reader MUST parse both legacy `v1` and legacy `v2` lines without crashing and treat BOTH as weight-absent (case (a) in REQ-CGA-002).

### F.3 harness-delivery-strategy.md

198 LOC, maintainer-local. §1-§7 structure ending with `## Sources` at line ~191. The new "Context-Governance Axis" section will be inserted BEFORE `## Sources` (or as §8 — run-phase decision).

---

## §G. History

| Version | Date | Author | Notes |
|---------|------|--------|-------|
| 0.1.0 | 2026-06-18 | manager-spec | Initial draft — Sprint 15 P2b cohort, harness-books ch8.3/diag-05 application. |

---

## §H. Cross-References

- `.moai/docs/harness-delivery-strategy.md` — target document for the Context-Governance Axis § (REQ-CGA-004/005).
- `.moai/harness/usage-log.jsonl` — schema extension target (REQ-CGA-001/002).
- `.claude/hooks/moai/handle-harness-observe*.sh` (4 files) — observer hook wrappers (REQ-CGA-001).
- `internal/harness/observer.go` (recorder: `RecordEvent` L53, `RecordExtendedEvent` L103 — re-verify at run-phase, line numbers drift), `internal/cli/hook.go` (CLI dispatch: `runHarnessObserve*` handlers L601/659/741/895), `internal/harness/types.go` (`Event` struct L65 — schema-extension point; `LogSchemaVersion = "v2"` constant at L16; `SchemaVersion` field wired at L81-82). Go side — run-phase MUST re-verify before editing (DISCOVERY-FIRST).
- P0 SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001 (Sprint 15 cohort companion) — death-spiral avoidance property that REQ-CGA-003 fail-open must preserve.
- `.claude/rules/moai/workflow/context-window-management.md` — REACTIVE thresholds (1M=50% / 200K=90%); this SPEC adds the PREVENTIVE complement.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12-field frontmatter SSOT.
- `.claude/rules/moai/workflow/lifecycle-sync-gate.md` — era V3R6 classification (this SPEC is V3R6, subject to 4-phase close).

---

## §X. Out of Scope (Exclusions)

### §X.1 Out of Scope — Auto-demotion of rules to on-demand skills

- This SPEC **surfaces** Tier-2 demote proposals via the drift alarm; it does **not** automatically execute demotion. Demotion remains a human-in-the-loop harness-learning decision (Tier-2 proposal → human approval → separate future SPEC executes the demotion).

### §X.2 Out of Scope — Token-budget enforcement / hard caps on eager context

- This SPEC records weight and raises an alarm; it does **not** enforce a hard cap on eagerly-loaded context (no "block the turn if eager weight > N tokens" gate). Enforcement is a possible future SPEC; this one is observability + alarm only.

### §X.3 Out of Scope — Rewriting the existing session-handoff thresholds

- The REACTIVE thresholds in `context-window-management.md` (1M=50% / 200K=90%) are NOT modified. This SPEC adds the PREVENTIVE complement (weight recording + drift alarm); the reactive `/clear` triggers remain canonical.

### §X.4 Out of Scope — Rewriting the harness-learning Tier system (1-4)

- Per REQ-CGA-006, the Tier system is untouched. The drift alarm feeds Tier-2 as a new signal; it does not add/rename/remove tiers.

### §X.5 Out of Scope — Exact token measurement

- Token estimation uses a bytes-based heuristic (e.g. bytes/4). Exact tokenizer-accurate measurement is out of scope; monotonic-growth detection does not require exactness.

### §X.6 Out of Scope — Per-skill weight breakdown

- The on-demand weight is recorded as an aggregate (sum of invoked skill bodies). Per-skill attribution (which specific skills contributed) is NOT recorded in this SPEC — deferred to a future SPEC if the drift alarm needs finer granularity.

### §X.7 Out of Scope — Compact/truncation/recovery trio modification

- Appendix B.2's trio invariant is cross-referenced as doctrinal context only. This SPEC does not modify the compaction, truncation, or recovery mechanisms themselves.
