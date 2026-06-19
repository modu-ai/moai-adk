# Research — SPEC-V3R6-ORCH-IGGDA-001

> This document maps every existing component the IGGDA redesign touches, with file:line citations. Every claim in design.md §F (FROZEN-invariant analysis) is grounded here. Per `moai-workflow-spec/SKILL.md` § Red Flags, research.md is mandatory when the SPEC touches existing code/rules.

---

## §A — Existing-component mapping (file:line)

### §A.1 `run.md` (the run-phase autonomy sibling)

**File**: `.claude/skills/moai/workflows/run.md` (verified via `Read` this session, 172 lines total)

| Lines | Component | IGGDA relationship |
|-------|-----------|---------------------|
| `run.md:114` | Ordering invariant statement ("Implementation Kickoff Approval ... always cleared FIRST") | IGGDA's D2 safe-condition predicate AMENDS this ordering (Path B) |
| `run.md:116-171` | `## Run-phase Autonomy (/goal ac_converge)` section (the sibling's D2 deliverable) | IGGDA EXTENDS this with cross-phase chaining (Phase 1→2→3 auto-advance) + the recursive self-diagnosis loop |
| `run.md:120-126` | `### 1. Implementation Kickoff Approval ordering` subsection (the FROZEN gate) | IGGDA's D2 AMENDS lines 122 + 124 (the safe-condition predicate) — the HIGHEST-STAKES change |
| `run.md:122` | "[HARD] Before any run-phase autonomy ..., the orchestrator MUST have already obtained explicit Implementation Kickoff Approval approval." | The FROZEN invariant statement. Path B preserves the `AskUserQuestion` ISSUANCE but reduces its blocking weight under the safe-condition predicate. |
| `run.md:124` | "[HARD] Implementation Kickoff Approval is **score-independent** ... skip-eligibility applies ONLY to Phase 0.5 plan-auditor verdict re-execution — NOT to Implementation Kickoff Approval." | The score-independence claim. IGGDA's Path B does NOT make the gate score-dependent; it introduces a COMPOUND predicate (intent + audit + tier + keywords) that is independent of plan-auditor score. |
| `run.md:128-144` | `### 2. The ac_converge /goal condition` (the sibling's autonomy mechanism) | IGGDA's D4 Stop hook driver is the moai-native equivalent — the user does NOT author a `/goal` condition string. |
| `run.md:146-148` | `### 3. Transcript-measurability` | IGGDA inherits this verbatim — the Stop hook driver reads `progress.md` + `moai spec audit` (transcript-measurable), never infers from frontmatter. |
| `run.md:150-152` | `### 4. Semantic-failure escalation (HARD)` | IGGDA's REQ-IGGDA-016 inherits this verbatim — semantic failures NEVER auto-fixed. |
| `run.md:154-156` | `### 5. Non-substitution clause (HARD)` | IGGDA's REQ-IGGDA-027 inherits this verbatim — autonomy does NOT authorize destructive operations. |
| `run.md:158-160` | `### 6. Blocker reports, never user prompts` | IGGDA's REQ-IGGDA-013 + B11 inherit this verbatim — the Stop hook driver returns exit codes + JSON. |
| `run.md:162-164` | `### 7. Graceful degradation when /goal is unavailable` | IGGDA's REQ-IGGDA-012 inherits this verbatim. The `run.md:164` "autonomy-config follow-up" is DEFERRED (this SPEC's EX-1). |
| `run.md:166-171` | `### Cross-references (cite, do not restate)` | IGGDA extends this cross-reference list with the IGGDA pipeline definition. |

### §A.2 `orchestration-mode-selection.md` (Phase 0.95 mode catalog)

**File**: `.claude/rules/moai/workflow/orchestration-mode-selection.md` (verified via `grep` this session)

| Lines | Component | IGGDA relationship |
|-------|-----------|---------------------|
| `:14` | "[ZONE:Frozen] [HARD] All Phase 0.95 execution modes are strictly downstream of Implementation Kickoff Approval ... mandatory and score-independent" | The FROZEN rule. IGGDA's D1 + D2 AMEND this with the safe-condition predicate. The `[ZONE:Frozen]` marker is PRESERVED (§G Out of Scope). |
| `:33` | "Mode 5 is the **default fallback** ... Mode 6 (`workflow`) is the narrow high-volume-mechanical exception, selectable ONLY after Implementation Kickoff Approval has passed" | IGGDA's Phase 2 heavy work uses Mode 5 (default) or Mode 6 (mechanical fan-out); both require Implementation Kickoff Approval cleared (or auto-proceeded under Path B). |
| `:67` | Decision tree branch — Mode 6 checks Implementation Kickoff Approval already passed | IGGDA's safe-condition predicate feeds this check. |
| `:90` | Auto-mode pre-launch classifier (CC 2.1.178+) | IGGDA's Stop hook driver coexists with this platform-level classifier (complementary, not conflicting). |
| `:137-141` | Mode 6 preconditions table (Implementation Kickoff Approval passed + preferences collected + scope mechanical + Workflows available + selection logged) | IGGDA's Phase 2 Mode 6 path inherits these preconditions. |
| `:151-153` | "Mode 6 / `/goal` agents return blocker reports, never prompt the user" | IGGDA's REQ-IGGDA-013 + B11 inherit this verbatim. |
| `:196-213` | §E Anti-Patterns + §F Cross-References | IGGDA adds new anti-patterns (AP-IGGDA-M1-001 through M6-001) in plan.md §G. |

### §A.3 `runtime-recovery-doctrine.md` (the 5 circuit-breaker invariants)

**File**: `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` (verified via `grep` this session)

| Lines | Component | IGGDA relationship |
|-------|-----------|---------------------|
| `:1-6` | Header + policy-layer-only declaration | IGGDA's D3 + D4 are policy-layer + shell-hook; NO Go runtime creep (AP-RR-001 compliance). |
| `:14-18` | §1 withheld-recoverable error set `{PTL, max_output_tokens, media_size, compact-failure}` | IGGDA's REQ-IGGDA-011 (Stop hook driver Recovery-Signal Carve-Out) detects these signals. |
| `:28-35` | §2 4-rung cheapest-first recovery ladder | IGGDA's recursive self-diagnosis loop is rung 1 (in-turn self-correction) for mechanical failures; rung 4 (abort + preserve) on semantic failure or PTL. |
| `:51-61` | §3 5 circuit-breaker invariants | IGGDA's REQ-IGGDA-017 complies with ALL 5. The max-3-iteration bound IS invariant 1's projection. |
| `:53` | Invariant 1 — `MAX_CONSECUTIVE_AUTOCOMPACT_FAILURES=3` | IGGDA's REQ-IGGDA-014 max-3-iteration bound aligns with this verbatim. |
| `:55` | Invariant 2 — `hasAttemptedReactiveCompact` no-self-loop | IGGDA's recursive loop does NOT re-attempt the same DIAGNOSE-PATCH-VERIFY within one turn. |
| `:57` | Invariant 3 — `truncateHeadForPTLRetry` compact-can-PTL last-resort escape | IGGDA's recursive loop falls to rung-4 abort + preserve on PTL, NOT another patch. |
| `:59` | Invariant 4 — abort-closes-ledger | IGGDA's recursive loop persists state to `progress.md §E Recursive Self-Diagnosis Log` before session end. |
| `:61` | Invariant 5 — narrative-consistency (5-Section Evidence-Bearing Report) | IGGDA's recursive loop reports across compact/recovery boundaries via the 5-section format. |
| `:71-89` | §4 Recovery-Signal Carve-Out | IGGDA's REQ-IGGDA-011 + AC-IGGDA-014 implement this in the Stop hook driver. |
| `:113` | §6 agent consult-the-doctrine obligation | IGGDA's recursive-loop agent inherits this obligation. |
| `:122-124` | §7 anti-patterns (AP-RR-001 through AP-RR-006) | IGGDA's design.md §H + plan.md §G add IGGDA-specific anti-patterns but do NOT contradict these. |

### §A.4 `goal-directive.md` (`/goal` semantics)

**File**: `.claude/rules/moai/workflow/goal-directive.md` (cross-referenced from run.md:168, not re-read this session)

| Component | IGGDA relationship |
|-----------|---------------------|
| `/goal <condition>` session-scoped completion condition | IGGDA's D4 Stop hook driver is the moai-native equivalent — the user does NOT author a condition string. |
| Haiku evaluator judges the transcript only (does NOT run tools / read files) | IGGDA's Stop hook driver DIFFERS here — it DOES read `progress.md` + invoke `moai spec audit`. This is a deliberate design choice (the driver is a moai-layer hook, not the platform `/goal` evaluator). |
| `max N turns` bound | IGGDA inherits the bound concept (per-phase max turns); the bound value is per-phase (Phase 2 inherits `max 20 turns` from AUTONOMY-RUN-GOAL-001 REQ-ARG-008). |
| Clear-on-`/clear` | IGGDA inherits — a `/clear` resets the IGGDA pipeline state; the paste-ready resume (session-handoff.md) re-establishes it. |
| `/goal clear` on condition met | IGGDA's terminal gate (REQ-IGGDA-028) emits the equivalent clear signal when all 3 completeness conditions hold. |

### §A.5 `dynamic-workflows.md` (ultracode + Workflow primitive)

**File**: `.claude/rules/moai/workflow/dynamic-workflows.md` (cross-referenced from orchestration-mode-selection.md:211, not re-read this session)

| Component | IGGDA relationship |
|-----------|---------------------|
| Workflow primitive (v2.1.154+, 16 concurrent / 1000 total) | IGGDA's Phase 2 Mode 6 path uses this for genuinely-parallel mechanical fan-out. |
| No mid-run user input | IGGDA inherits — Phase 0 drains all preferences BEFORE any Workflow launch. |
| Implementation Kickoff Approval unaffected | IGGDA's D2 safe-condition predicate is the MOAI-NATIVE amendment; the Workflow primitive itself is unaffected. |
| `ultracode` session mode | IGGDA's Phase 2 MAY run under `ultracode` for autonomous fan-out; the paste-ready resume (session-handoff.md Block 1) re-sets `/effort ultracode` after `/clear`. |

### §A.6 `session-handoff.md` (paste-ready resume across `/clear`)

**File**: `.claude/rules/moai/workflow/session-handoff.md` (always-loaded per its header declaration)

| Component | IGGDA relationship |
|-----------|---------------------|
| 6-block paste-ready resume format | IGGDA's long-running autonomy crosses `/clear` boundaries; the resume format's Block 1 `ultrathink.` + Block 5 `/moai run` re-establishes the IGGDA pipeline state. |
| Block 1 `/effort ultracode` re-set line (purpose-conditional) | IGGDA's Phase 2 Mode 6 path triggers this line (the next SPEC's plan declares workflow fan-out). |
| Block 2 `source_session_id` | IGGDA's Stop hook driver logs the session_id for race attribution across `/clear` boundaries. |
| Cut-line markers (✂ U+2702 preserved verbatim) | IGGDA inherits — the resume format is unchanged. |

### §A.7 `verification-claim-integrity.md` (the no-unobserved-claim invariant)

**File**: `.claude/rules/moai/core/verification-claim-integrity.md` (loaded as project instructions this session)

| Component | IGGDA relationship |
|-----------|---------------------|
| §1.1 surface 3 — defect/success claims require the dedicated tool's output | IGGDA's REQ-IGGDA-010 mandates the Stop hook driver invoke `moai spec audit` (the dedicated tool), NEVER infer from frontmatter text. |
| §2 baseline-integrity attribution | IGGDA's recursive loop logs baseline-attribution in the 5-Section Evidence-Bearing Report (invariant 5 compliance). |
| §3 5-Section Evidence-Bearing Report Format | IGGDA's recursive loop inherits this format across compact/recovery boundaries. |
| §5 worked example — era=NONE implemented 29 SPECs mis-counted as "Mx-close debt" | IGGDA's design.md §F cites this as the cautionary tale for why the Stop hook driver MUST use `moai spec audit`, not frontmatter inference. |

### §A.8 `archived-agent-rejection.md` (the 12-agent migration table)

**File**: `.claude/rules/moai/workflow/archived-agent-rejection.md` (loaded as project instructions this session)

| Component | IGGDA relationship |
|-----------|---------------------|
| §C row #7 — `expert-backend` → `Agent(general-purpose, model: opus, tools: <backend whitelist>)` | IGGDA's recursive self-diagnosis sub-agent uses this per-spawn pattern (REQ-IGGDA-015). |
| §C row #2 — `manager-quality` → Stop hook enforcement | IGGDA's D4 Stop hook driver is the direct successor of this pattern (the hook replaces the phantom `manager-quality` spawn). |
| §D orchestrator recovery flow (ToolSearch → AskUserQuestion → re-delegate) | IGGDA inherits — the recursive loop's escalation path follows this flow. |

### §A.9 `CLAUDE.local.md §19.1` (the FROZEN invariant owner)

**File**: `CLAUDE.local.md` (loaded as user-private project instructions this session)

| Lines | Component | IGGDA relationship |
|-------|-----------|---------------------|
| `:702` | "§19.1 구현 착수 승인 (renamed from GATE-2) Mandatory Restoration (REQ-ATR-015 — SPEC-V3R6-AGENT-TEAM-REBUILD-001)" | The FROZEN invariant's canonical reference. IGGDA amends its BEHAVIOR (Path B) without removing the §19.1 section. |
| `:706` | "skip-eligible 0.90 autonomous bypass 정책의 적용 범위 ... 구현 착수 승인 (plan-to-implement HUMAN GATE)에는 적용되지 않는다" | IGGDA's safe-condition predicate does NOT use skip-eligibility; it uses a COMPOUND predicate (intent + audit + tier + keywords). The score-independence claim is PRESERVED. |
| `:714` | "위반 anti-pattern: Phase 0.5 verdict가 PASS skip-eligible (≥ 0.90)이라는 이유만으로 사용자 승인 없이 `/moai run`을 자율 시작하는 행위" | IGGDA's Path B does NOT violate this — the `AskUserQuestion` is STILL ISSUED (AC-IGGDA-004); the predicate is compound, not score-only. |

### §A.10 `moai spec audit` (the deterministic SPEC-compliance tool)

**File**: `internal/spec/audit.go` + `internal/spec/era.go` + `internal/spec/drift.go` (cross-referenced from lifecycle-sync-gate.md, not re-read this session)

| Component | IGGDA relationship |
|-----------|---------------------|
| `moai spec audit --json` | IGGDA's Stop hook driver invokes this. |
| `--filter-spec=<SPEC-ID>` flag | **Q4 open question** — does this flag exist today? If not, M5 adds it (minor Go wiring). Pre-flight check in plan.md §C resolves this. |
| `drift_findings[]` with severity (MUST-FIX / INFO) | IGGDA's REQ-IGGDA-028 requires 0 MUST-FIX for IGGDA-completeness. INFO findings do NOT block. |
| `EraAutoDetected` INFO finding | IGGDA's SPEC sets `era: V3R6` explicitly in frontmatter (H-override), suppressing this finding. |
| `OwnershipTransitionRule` | IGGDA's 3-phase close (M6) complies — `draft → in-progress` by manager-develop, `in-progress → implemented → completed` by manager-docs. |

---

## §B — FROZEN invariant lineage (traceability)

The Implementation Kickoff Approval invariant's lineage, traced through the rule system:

```
SPEC-V3R6-AGENT-TEAM-REBUILD-001 (REQ-ATR-015)
  │  canonical origin — Implementation Kickoff Approval mandatory restoration
  │  (originally "GATE-2", renamed to "Implementation Kickoff Approval")
  ▼
CLAUDE.local.md §19.1 (:702-714)
  │  local-development doctrine — the FROZEN invariant owner
  │  declares: score-independent, never auto-bypassed, mandatory AskUserQuestion
  ▼
.claude/rules/moai/workflow/orchestration-mode-selection.md:14
  │  [ZONE:Frozen] [HARD] — all Phase 0.95 modes downstream of Implementation Kickoff Approval
  ▼
.claude/skills/moai/workflows/run.md:122,124
  │  [HARD] — the gate is always cleared FIRST; score-independent
  │  (added by SPEC-AUTONOMY-RUN-GOAL-001 D2/D3 — the run-phase autonomy sibling)
  ▼
SPEC-AUTONOMY-RUN-GOAL-001 (status: completed)
  │  sibling — preserved the invariant verbatim under 6 HARD safety conditions C1-C6
  │  C1 = "Implementation Kickoff Approval mandatory, score-independent"
  ▼
SPEC-V3R6-ORCH-IGGDA-001 (this SPEC, status: draft)
     AMENDS the invariant under user-mandated Path B:
     - introduces the safe-condition predicate (REQ-IGGDA-004/005)
     - PRESERVES the AskUserQuestion issuance (AC-IGGDA-001, AC-IGGDA-004)
     - PRESERVES the [ZONE:Frozen] marker (AC-IGGDA-008)
     - CARVES OUT dangerous domains (REQ-IGGDA-005, design.md §F.5)
     - The amendment is REVERSIBLE (single rule-file edit restores Path A)
```

**Lineage summary**: the invariant has 4 canonical touchpoints (REQ-ATR-015 → CLAUDE.local.md §19.1 → orchestration-mode-selection.md:14 → run.md:122,124) plus 1 sibling preservation (AUTONOMY-RUN-GOAL-001 C1). IGGDA amends all 4 touchpoints consistently — the safe-condition predicate is documented in each location, cross-referenced to design.md §F.

---

## §C — Anthropic plan-editor mandate relationship

### §C.1 What the mandate is

Claude Code's Ctrl+G plan editor mandate is a platform-level mechanism (documented in Claude Code's official docs, not in moai-adk rules) that requires the user to review and approve a plan before the model enters implementation. The mandate fires on certain surfaces (Claude Code TUI, headless `-p` with specific flags).

### §C.2 What IGGDA's Path B does NOT claim

IGGDA's Path B does NOT claim to bypass the Anthropic plan-editor mandate. The mandate is a platform-level mechanism that fires regardless of MoAI's rule layer. If the mandate fires AFTER MoAI's auto-proceed, there is no conflict (the user gets a second chance via the platform gate).

### §C.3 What IGGDA's Path B DOES claim

Path B AMENDS MoAI's rule-layer Implementation Kickoff Approval (the moai-native analogue), reducing its per-run-blocking weight in safe domains. The Anthropic mandate (if it fires) still applies on top.

### §C.4 The intent-verification satisfaction argument

When the safe-condition predicate auto-proceeds, the user has ALREADY reviewed the plan via:
1. **Phase 0 Socratic interview** — multi-round intent collection to 100% clarity (condition (a)).
2. **Phase 1 plan-auditor verdict** — independent audit PASS surfaced to the user (condition (b)).

The Anthropic mandate's intent (user reviewed before implementation) is satisfied by the Socratic interview, which is UPSTREAM of the predicate. The mandate's value (catch mis-captured intent) is delivered by the Socratic interview's multi-round rigor, not by a single plan→run boundary confirmation.

### §C.5 Honest flag

The Anthropic plan-editor mandate's exact firing conditions are not fully documented in moai-adk's rule system. Two scenarios:
- **Mandate fires AFTER MoAI auto-proceed**: no conflict; user gets a second chance. Acceptable.
- **Mandate does NOT fire (some surfaces)**: the MoAI auto-proceed stands on its own; the Phase 0 Socratic interview is the sole intent verification. Acceptable because condition (a) requires 100% clarity.

This uncertainty is documented in design.md §F.6 as an honest flag. The plan-auditor should scrutinize whether this uncertainty is acceptable or whether the SPEC should be more conservative (e.g., require the mandate to fire as a precondition for auto-proceed — which would require platform detection that is out of scope for this SPEC).

---

## §D — Related SPEC survey

### §D.1 Direct predecessor/sibling

**`SPEC-AUTONOMY-RUN-GOAL-001`** (status: completed, origin: `aaf556119`-chain per memory)
- **Relationship**: the direct predecessor. Delivered run-phase autonomy (`/goal ac_converge` + Mode 6) under 6 HARD safety conditions C1–C6.
- **What IGGDA inherits**: the `ac_converge` condition structure, the semantic-failure escalation (run.md:152), the non-substitution clause (run.md:154-156), the blocker-report boundary (run.md:158-160), the graceful-degradation contract (run.md:162-164), C2–C6 safety conditions.
- **What IGGDA amends**: C1 (Implementation Kickoff Approval mandatory, score-independent) — transformed into the safe-condition predicate (Path B).
- **What IGGDA extends**: cross-phase chaining (Phase 1→2→3 auto-advance via D4 Stop hook driver), bounded recursive self-diagnosis loop (D3).

### §D.2 Dependency

**`SPEC-V3R6-WORKFLOW-EFFORT-MAP-001`** (status: completed, origin: `aab4c3297` per memory)
- **Relationship**: the purpose-driven model+effort selection SSOT for Workflow `agent()` calls.
- **What IGGDA uses**: the Phase 2 Mode 6 path's effort selection (per-purpose `(model, effort)` taxonomy) is governed by this SPEC. IGGDA does NOT re-derive effort selection; it references WORKFLOW-EFFORT-MAP-001.

### §D.3 Dependency

**`SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001`** (status: completed per memory `project_sprint15_cohort_p0p1a_closed_p2p3_handoff.md`)
- **Relationship**: the orchestrator interrupt ledger closure doctrine.
- **What IGGDA uses**: the Stop hook driver's `ledger_note` JSON field (REQ-IGGDA-013) is the ledger-closing artifact pattern from this SPEC. On a phase-transition block (exit 2), the `ledger_note` closes the open promise so the next turn does not proceed as if the phase transition succeeded.

### §D.4 Related (not dependency)

**`SPEC-V3R6-AGENT-TEAM-REBUILD-001`** (status: completed)
- **Relationship**: the origin of REQ-ATR-015 (Implementation Kickoff Approval mandatory restoration) — the FROZEN invariant this SPEC amends.
- **What IGGDA references**: REQ-ATR-015 is the lineage root (design.md §F, research.md §B).

### §D.5 Related (not dependency)

**`SPEC-V3R6-LIFECYCLE-REDESIGN-001`** (status: completed)
- **Relationship**: the 3-phase lifecycle (plan→run→sync) canonicalization.
- **What IGGDA uses**: the 3-phase close (M6) follows this SPEC's lifecycle. The progress.md §E structure is 4 sections (§E.1 Plan / §E.2 Run Evidence / §E.3 Run Audit-Ready / §E.4 Sync Audit-Ready).

### §D.6 Related (not dependency)

**`SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001`** (status: completed)
- **Relationship**: the close-subject full-ID mandate.
- **What IGGDA uses**: M6's close commit subject MUST be `chore(SPEC-V3R6-ORCH-IGGDA-001): ... 3-phase close` (NOT combined/abbreviated scope).

### §D.7 Forward-link (not yet authored)

**`SPEC-V3R6-IGGDA-PREFLIGHT-001`** (candidate, not yet authored)
- **Relationship**: the version-preflight follow-up (EX-1).
- **What it would deliver**: detect Claude Code runtime version + `/goal` availability + hook enablement, emit a structured signal. IGGDA v0.1.0 inherits the graceful-degradation contract (run.md:162-164) and assumes `/goal` is available.

### §D.8 Forward-link (not yet authored)

**`SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001`** (candidate, mentioned in runtime-recovery-doctrine.md §4)
- **Relationship**: the runtime-layer hook that parses `stopReason` to mechanically enforce the Recovery-Signal Carve-Out.
- **What it would deliver**: `stopReason` parsing in the Stop hook. IGGDA's D4 Stop hook driver does NOT parse `stopReason` in v0.1.0 (AP-RR-006 compliance — do not claim mechanical enforcement this layer cannot provide).

---

## §E — Keyword-list initial provenance

The safe-condition keyword list (design.md §F.3) is derived from:
1. **OWASP Top 10** — security domain keywords (auth, injection, xss, csrf, etc.).
2. **PCI-DSS** — payment domain keywords (card, pan, pci, etc.).
3. **Common production-incident triggers** — critical domain keywords (prod, migration, drop table, force-push, rm -rf).
4. **moai-adk's own CONST-V3R5-011/013** — destructive-path prohibitions (the recursive loop's forbidden-paths list, inherited verbatim).

The list is intentionally OVER-INCLUSIVE in v0.1.0 (per design.md §F.3). False-positives (forcing explicit-gate on a SPEC that is actually safe) are acceptable; false-negatives (auto-proceeding on a SPEC that is actually dangerous) are NOT.

---

## §F — Template-file impact inventory (for run-phase, NOT this plan-phase)

Per Template-First Rule (CLAUDE.local.md §2), this plan-phase authors SPEC artifacts ONLY. The template files this SPEC will touch in run-phase (enumerated for the run-phase plan):

| Template file | M milestone | Change |
|---------------|-------------|--------|
| `internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md` | M1 | D1 pipeline + D2 predicate (FROZEN-amend) |
| `internal/template/templates/.claude/skills/moai/workflows/run.md` | M1 (D2 amendment) + M2 (D3 NEW section) | Lines 120-126 amendment + NEW `## Recursive Self-Diagnosis Loop (bounded)` section |
| `internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md` | M2 | D3 reference injection |
| `internal/template/templates/.claude/hooks/moai/iggda-phase-driver.sh` | M3 | D4 NEW hook |
| `internal/template/templates/.claude/hooks/moai/handle-iggda-phase-driver.sh` | M3 | D4 NEW wrapper |
| `internal/template/templates/.claude/settings.json.tmpl` | M3 | D4 Stop hook registration |
| `internal/spec/audit.go` (if `--filter-spec` absent) | M5 | Additive Go flag |
| `internal/cli/spec_audit.go` (if `--filter-spec` absent) | M5 | Wire the flag |

**Neutrality CI guard**: every template edit MUST pass `TestTemplateNeutralityAudit` (no internal SPEC IDs / commit SHAs / macOS-bias paths / `feedback_` refs). The keyword list (design.md §F.3) is generic prose, compliant.

---

## §G — Open questions (Q1–Q4, traced to spec.md §F)

| Q | Question | Resolution path |
|---|----------|-----------------|
| Q1 | Keyword-list maintenance ownership | plan.md M2 + design.md §F.3 — maintained in orchestration-mode-selection.md §F.3 via SPEC amendment |
| Q2 | Tier L boundary correctness | plan-auditor review — REQ-IGGDA-005 currently forces explicit-gate for BOTH Tier L AND any-tier + security keywords (the OR is deliberate) |
| Q3 | Auto-proceed timeout value | design.md §F.4 — proposed 30 seconds, configurable via workflow.yaml |
| Q4 | `moai spec audit --filter-spec` flag existence | plan.md §C pre-flight — `grep -n "filter-spec\|FilterSpec" internal/spec/audit.go internal/cli/spec*.go` resolves this; if absent, M5 adds it |

---

## §H — Research completeness self-check

- [x] Every file:line citation in design.md §F is grounded in this research.md (§A).
- [x] The FROZEN invariant lineage is traced end-to-end (§B).
- [x] The Anthropic plan-editor mandate relationship is documented with honest flags (§C).
- [x] All related SPECs are surveyed with relationship + inheritance + amendment (§D).
- [x] The keyword-list provenance is documented (§E).
- [x] The template-file impact is inventoried for run-phase (§F).
- [x] All 4 open questions are traced to resolution paths (§G).

---

Version: 0.1.0 (plan-phase)
Status: Active — awaits plan-auditor independent audit
