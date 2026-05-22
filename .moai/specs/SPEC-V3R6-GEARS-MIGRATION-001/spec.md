---
id: SPEC-V3R6-GEARS-MIGRATION-001
title: "EARS → GEARS Keyword Migration + Lint Rule Update + 4-locale Docs Sync"
version: "0.1.0"
status: draft
created: 2026-05-22
updated: 2026-05-22
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/spec, .moai/specs, docs-site/content"
lifecycle: spec-anchored
tags: "ears, gears, lint, migration, frontmatter, i18n, wave-6, v3.0.0"
tier: M
issue_number: null
depends_on: []
related_specs: [SPEC-V3R6-RULES-COMPLIANCE-001, SPEC-V3R6-V3-CUTOVER-001]  # forward-references — both planned Wave 6 SPECs, not yet authored as of 2026-05-22 (per blueprint:174)
---

# SPEC-V3R6-GEARS-MIGRATION-001 — EARS → GEARS Keyword Migration

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-22 | manager-spec | Initial draft — Wave 6 SPEC #2 per `.moai/research/v3-redesign-blueprint-2026-05-22.md` line 16 + 24 + 174 (user 2026-05-22 confirm). EARS → GEARS keyword migration: update `internal/spec/lint.go` `EARSModalityRule` (replace IF/THEN drop + WHERE clarification) + write user-facing migration guide in 4-locale docs-site + provide policy for the 88 existing SPECs (default: keep legacy keyword OK as backward-compat warning, NEW SPECs use GEARS). Tier M (3 artifacts) — touches lint engine + cross-cutting docs but no behavioral change to runtime workflow. Defers actual 88-SPEC bulk rewrite to follow-up sweep SPEC per AskUserQuestion default. |

## 1. Goal

Migrate MoAI-ADK's requirement notation from **EARS (Easy Approach to Requirements Syntax, Alistair Mavin 2009)** to **GEARS (Generalized Expression for AI-Ready Specs, DEV Community 2026-01-23)** as the canonical keyword set for new SPEC documents (`.moai/specs/SPEC-*/spec.md` REQ entries).

The migration **adds GEARS keyword support** to `internal/spec/lint.go` `EARSModalityRule` and **adds user-facing migration documentation** in `docs-site/content/{ko,en,ja,zh}/` so users can author SPECs in either notation. The migration **does NOT mass-rewrite the 88 existing SPECs** in this SPEC — backward compatibility is preserved via legacy-keyword acceptance (warning-only), and a separate sweep SPEC (`SPEC-V3R6-GEARS-SWEEP-001`, provisional) is queued for the actual bulk rewrite once v3.0.0 cutover is stable.

**Canonical GEARS keyword set** (from *GEARS: The Spec Syntax That Makes AI Coding Actually Work*, Σ\*/SubLang, DEV Community 2026-01-23 — verified by orchestrator web research 2026-05-22, see `.moai/research/gears-paper-validation.md`. Q3 RESOLVED.):

| Notation pattern | EARS (legacy) | GEARS (new) | Semantic |
|------------------|--------------|--------------|----------|
| Ubiquitous (always active) | `The system shall <action>` | `The system shall <action>` (unchanged) | Invariant |
| Event-driven (trigger) | `WHEN <event>, the system shall <action>` | `WHEN <event>, the system shall <action>` (unchanged) | Trigger-response |
| State-driven (continuous condition) | `WHILE <state>, the system shall <action>` | `WHILE <state>, the system shall <action>` (unchanged, **promoted** as primary state notation) | Continuous |
| Optional precondition (capability gate) | `WHERE <feature exists>, the system shall <action>` | `WHERE <precondition>, the system shall <action>` (**clarified** — now also covers feature-flag gates and capability negotiation, no longer purely "feature exists") | Conditional opt-in |
| Negative trigger (unwanted event) | `IF <undesired>, THEN the system shall <action>` | **DEPRECATED** — express as `WHEN <event-detected>, the system shall <action>` (event normalized to detection) | Trigger-response (no IF/THEN) |

The GEARS rationale (per the DEV 2026 paper):

1. **IF/THEN is ambiguous for LLM parsing** — IF can be read as either trigger (event) or guard (state), causing parser-to-parser disagreement. WHEN (trigger) + WHILE (state) are syntactically separable.
2. **WHERE was under-used** (0 of 88 MoAI SPECs use WHERE today) because the original "feature exists" framing was too narrow. Reframing WHERE as "precondition / capability gate" gives it real surface area.
3. **Ubiquitous + WHEN + WHILE + WHERE = 4-keyword closed set** is easier for LLM SPEC writers and SPEC parsers to internalize than EARS's 5-keyword open set with IF/THEN ambiguity.

### 1.1 GEARS Unified Pattern (paper feature — informational)

The GEARS paper introduces a **single unified syntax pattern** with optional clauses, replacing EARS's five separate patterns:

```
[Where <static precondition(s)>]
[While <stateful precondition(s)>]
[When <trigger>]
The <subject> shall <behavior>
```

Square brackets denote optional clauses. Multiple clauses may be composed in a single requirement.

**Example (verbatim from paper)**:

> "Where the deployment is production, when a request fails, the service shall retry with exponential backoff."

**Generalized subject**: GEARS replaces "the system" with `<subject>` — any noun (system, component, service, agent, function, artifact). MoAI-ADK's 88 existing SPECs retain "The system" for backward compatibility; new SPECs MAY use generalized subjects when more precise.

**Given-When-Then mapping** (paper feature): GEARS clauses map directly to Given-When-Then test syntax — `Given` → `Where`+`While`, `When` → `When`, `Then` → `shall`. This is informational; MoAI-ADK does not enforce this mapping at the lint layer.

This subsection is **documentation-only**; the lint engine (M2) accepts EARS-style separate keywords as well as GEARS compound clauses transparently. M3 4-locale docs should mention this unified pattern as part of the migration guide.

## 2. Why

### 2.1 Industry Finding (Blueprint § Industry Findings #5)

> "EARS → GEARS evolution (drops If/Then; clarifies where vs while for AI parsing)."
> — `.moai/research/v3-redesign-blueprint-2026-05-22.md` line 24

The MoAI-ADK lint engine (`internal/spec/lint.go` `EARSModalityRule`, lines 404-447) currently enforces the legacy EARS 5-keyword set including IF/THEN. The lint warns when WHEN/WHILE/WHERE/IF prefixes are present without a `SHALL` token. This rule has flagged real malformed REQs (see `TestEARSModalityRule_*` in `lint_test.go`) but does not differentiate trigger-IF from guard-IF, so it cannot help users author cleaner SPECs for the AI parsing case.

### 2.2 Distribution Audit (Pre-flight Grep)

The 88 existing SPECs use EARS keywords in this distribution (measured 2026-05-22 against `.moai/specs/`):

| Keyword pattern | File count | % of 88 |
|-----------------|-----------:|--------:|
| `The system shall <action>` (Ubiquitous) | 37 | 42% |
| `WHEN <event> THEN/SHALL` (Event-driven) | 36 | 41% |
| `WHILE <state> the system shall` (State-driven) | 2 | 2% |
| `WHERE <feature> the system shall` (Optional precondition) | **0** | 0% |
| `IF <condition> THEN` (Conditional/Negative) | 42 | 48% |

Total SPECs using any EARS keyword: 88. (Files often use multiple patterns; the per-keyword counts sum to >88.)

Observations:

- **`IF...THEN` is the most-used keyword (42 files)** — direct evidence that the EARS-to-GEARS deprecation needs a clear migration path before any bulk rewrite.
- **`WHERE` is unused (0 files)** — the EARS "feature exists" framing failed; GEARS's "precondition / capability gate" reframing has 0 inertia to overcome.
- **`The system shall` and `WHEN` are stable** — these carry over unchanged in GEARS, so 73 of 88 SPECs (83%) are already GEARS-compliant for their primary REQs and only need IF/THEN audit.

### 2.3 Cross-cutting Surface

- **`internal/spec/lint.go`** — `EARSModalityRule.isModalityMalformed()` checks WHEN/WHILE/WHERE/IF prefixes for `SHALL` token. Update to accept GEARS notation as the canonical form + emit a `Warning`-severity `LegacyEARSKeyword` finding (NOT error) for IF/THEN to give existing SPECs a graceful migration window.
- **`.claude/skills/moai-foundation-core/modules/spec-ears-format.md`** — Currently documents 5 EARS patterns. Add GEARS pattern table + deprecation note for IF/THEN.
- **`docs-site/content/{ko,en,ja,zh}/`** — 53 pages reference "EARS" today (pre-flight grep). Add a `## GEARS notation (v3.0.0+)` section to the SPEC-authoring guide page in each locale (path to be determined at run-phase, candidate: `workflow-commands/moai-plan.md` or a new `concepts/spec-authoring.md`).

### 2.4 Wave 6 Position + Dependency

Per blueprint § Wave 6 (line 172-176):

> Wave 6: Final Compliance + Release (Tier M)
> - SPEC-V3R6-RULES-COMPLIANCE-001 (M — paths CSV, Korean→EN, zone-registry CI guard)
> - **SPEC-V3R6-GEARS-MIGRATION-001 (M — EARS → GEARS 키워드 마이그레이션, lint.go rule 갱신, 88개 SPEC 후보 변환 가이드, 4-locale docs-site 반영) [추가 2026-05-22 — 사용자 결정 반영]**
> - SPEC-V3R6-V3-CUTOVER-001 (M — v2.20.0-rc1 → v3.0.0 release manifest)

Dependency: this SPEC **MUST land before SPEC-V3R6-V3-CUTOVER-001** so that v3.0.0 release notes can announce the GEARS-as-default policy. It is independent of SPEC-V3R6-RULES-COMPLIANCE-001 (different rule layer).

**Wave ordering caveat (user-acknowledged)**: per AskUserQuestion 2026-05-22, the user opted to plan Wave 6 GEARS-MIGRATION-001 **immediately** despite Wave 1~5 SPECs being incomplete (HARNESS-LEARNER-FIX-001 implemented + Wave 1 4 SPECs in plan-only state + Wave 2-5 not yet started). This SPEC's plan-phase artifacts therefore document the dependency without acting on it; **run-phase entry blocks** until plan-auditor confirms cross-Wave compatibility is acceptable (see acceptance.md AC-GM-007).

### 2.5 References (Industry + Blueprint)

| Reference | Citation | Purpose |
|-----------|----------|---------|
| EARS official | Mavin, A. (2009-ongoing). *Easy Approach to Requirements Syntax*. https://alistairmavin.com/ears/ | Baseline notation being migrated FROM |
| GEARS proposal (verified 2026-05-22) | Σ\*/SubLang, *GEARS: The Spec Syntax That Makes AI Coding Actually Work*, DEV Community 2026-01-23, https://dev.to/sublang/gears-the-spec-syntax-that-makes-ai-coding-actually-work-4f3f | Target notation being migrated TO |
| Voyager (auto-evolution context) | arXiv:2305.16291 | Self-evolution baseline (peripheral) |
| Reflexion (self-evolution context) | arXiv:2303.11366 | Reflection-based agent self-improvement (peripheral) |
| AgentDevel (LLM SPEC-handling context) | arXiv:2601.04620 | LLM-authored SPEC parsing study (peripheral) |
| Anthropic effective context engineering | https://www.anthropic.com/news/effective-context-engineering-for-ai-agents | LLM-friendly notation rationale (peripheral) |

Peripheral references are listed for completeness; the load-bearing citations for the migration decision are **EARS official guide** + **GEARS proposal**.

## 3. Requirements (GEARS-native — authored in target notation as a dogfooding demonstration)

### REQ-GM-001 (Ubiquitous)

The system **shall** treat both EARS legacy keywords (WHEN, WHILE, WHERE, IF...THEN, Ubiquitous "The system shall") and GEARS keywords (WHEN, WHILE, WHERE-reframed, Ubiquitous "The system shall"; IF...THEN deprecated) as valid REQ modality during the v3.0.0 backward-compatibility window.

### REQ-GM-002 (Event-driven)

**WHEN** `internal/spec/lint.go` `EARSModalityRule.Check()` encounters a REQ whose text matches the legacy `IF <condition> THEN` pattern, the system **shall** emit a `LegacyEARSKeyword` finding at `Severity: warning` with message `"REQ %s: GEARS migration: replace IF/THEN with WHEN/event normalization; see docs-site GEARS guide"`.

### REQ-GM-003 (Event-driven)

**WHEN** `internal/spec/lint.go` `EARSModalityRule.Check()` encounters a REQ in GEARS notation (`WHEN <event>` / `WHILE <state>` / `WHERE <precondition>` / `The system shall`), the system **shall** validate the `SHALL` token presence exactly as the legacy EARS validation and emit zero findings on well-formed REQs.

### REQ-GM-004 (State-driven)

**WHILE** the system is generating user-facing SPEC-authoring documentation in `docs-site/content/{ko,en,ja,zh}/`, the system **shall** present GEARS as the recommended notation for new SPECs and EARS as the supported-but-deprecated legacy notation.

### REQ-GM-005 (Optional precondition — dogfooding WHERE reframing)

**WHERE** the orchestrator detects that `.moai/research/gears-paper-validation.md` has been written by run-phase web-research step (Open Question Q3 resolved), the system **shall** record the canonical GEARS keyword set verbatim in `internal/spec/lint.go` source comments AND in `docs-site/content/{ko,en,ja,zh}/` migration guide.

### REQ-GM-006 (Event-driven — IF→WHEN dogfooding)

**WHEN** a future SPEC author writes a new REQ using the legacy `IF...THEN` pattern, the system **shall** surface the GEARS migration warning in `moai spec lint` output AND link to the docs-site migration guide URL.

### REQ-GM-007 (Ubiquitous)

The system **shall** preserve the existing 88 SPEC documents' REQ entries unchanged in this SPEC; bulk rewrite is the explicit responsibility of `SPEC-V3R6-GEARS-SWEEP-001` (provisional, queued separately).

### REQ-GM-008 (Event-driven)

**WHEN** `moai spec lint --strict` is invoked, the system **shall** escalate `LegacyEARSKeyword` warnings to errors (per existing `--strict` policy in `internal/spec/lint.go` `Report.HasErrors()`), giving CI authors the option to enforce GEARS-only for new branches.

### REQ-GM-009 (State-driven)

**WHILE** the v3.0.0 backward-compatibility window is active (defined as: from the v3.0.0 release through 6 months post-release, OR until `SPEC-V3R6-GEARS-SWEEP-001` completes the 88-SPEC bulk rewrite, whichever comes first), the system **shall** accept EARS legacy IF/THEN as warning-only and not error-out non-strict `moai spec lint` runs.

## 3.1 Out of Scope (REQ-GM-001 scope guard)

### 3.1.1 Out of Scope

- **Mass rewrite of 88 existing SPECs** — explicitly deferred to `SPEC-V3R6-GEARS-SWEEP-001` (provisional). REQ-GM-007 preserves them.
- **Removal of `IF` from the lint regex** — `IF` is preserved as a recognized prefix to enable the warning; only the *interpretation* changes (was: warn on missing SHALL; now: warn on legacy pattern + missing SHALL).

## 3.2 Out of Scope (REQ-GM-005 dogfooding scope guard)

### 3.2.1 Out of Scope

- **Live web-research of the GEARS paper** during plan-phase — deferred to run-phase entry (see Open Question Q3 + acceptance.md AC-GM-005). If the paper validation reveals notation that contradicts §1 Table, run-phase MUST re-open this SPEC for amendment before code-touch.

## 3.3 Out of Scope (REQ-GM-009 backward-compatibility scope guard)

### 3.3.1 Out of Scope

- **Defining the post-window cutover policy** (warning → error transition) — this is `SPEC-V3R6-V3-CUTOVER-001`'s decision after the 6-month window or sweep completion.

## 4. Non-Goals (SPEC-wide)

- **Changing requirement semantics**: This SPEC changes notation/lint rules only, not what requirements mean or how they are implemented.
- **Modifying the EARSModalityRule's Code() identifier**: The rule code stays `ModalityMalformed` to preserve historical lint-suppression annotations in existing SPECs. A new finding code `LegacyEARSKeyword` is introduced for the warning class.
- **Localizing GEARS keywords**: GEARS keywords (WHEN, WHILE, WHERE, SHALL) remain English in all 4 locales — only the surrounding migration-guide prose is localized. (Matches existing EARS approach.)
- **Auto-rewriting user SPECs**: No `moai spec migrate` codemod is in scope. Manual edits + warning surface only.

## 5. Open Questions (resolved at run-phase entry)

| ID | Question | Default if user does not answer | Owner |
|----|----------|-------------------------------|-------|
| Q1 | **88 SPECs bulk-rewrite policy** — perform during this SPEC (extends to Tier L) or defer to `SPEC-V3R6-GEARS-SWEEP-001`? | **Defer** (Tier M scope guard, REQ-GM-007). Sweep SPEC queued for post-cutover stability window. | user |
| Q2 | **`--strict` enforcement timing** — should `moai spec lint --strict` block on `LegacyEARSKeyword` immediately on v3.0.0 release, or wait for the 6-month window? | **Immediate** (REQ-GM-008). CI authors can opt-in by adding `--strict` to their pipelines. Non-strict default unchanged. | user |
| Q3 | **GEARS paper URL** — RESOLVED 2026-05-22 — canonical reference is *GEARS: The Spec Syntax That Makes AI Coding Actually Work*, Σ\*/SubLang, DEV Community 2026-01-23, https://dev.to/sublang/gears-the-spec-syntax-that-makes-ai-coding-actually-work-4f3f. Validation report at `.moai/research/gears-paper-validation.md` showed MISMATCH on acronym text — corrected in §1 (this amendment). Functional migration path (IF/THEN → WHEN warning) is paper-aligned. | n/a — resolved by M1 orchestrator-direct | orchestrator |
| Q4 | **4-locale docs-site target file** — add a new `concepts/spec-authoring.md` page (×4 locales) or amend the existing `workflow-commands/moai-plan.md` page (×4 locales)? | **Amend existing** (`workflow-commands/moai-plan.md` × 4) to minimize i18n surface; lower risk per `.moai/docs/docs-site-i18n-rules.md`. | user |

These Open Questions do NOT block plan-phase artifact creation; they block run-phase entry via acceptance.md AC-GM-005 + AC-GM-007.

## 6. Stakeholders + Impact

| Stakeholder | Impact | Mitigation |
|-------------|--------|------------|
| MoAI-ADK end users (SPEC authors) | Need to learn GEARS notation for new SPECs. Existing SPECs continue to lint clean. | 4-locale migration guide + `moai spec lint` warning that links to the guide URL. |
| MoAI-ADK CI integrators | `--strict` mode now blocks IF/THEN; pipelines using `--strict` may fail on existing SPECs touched after v3.0.0. | REQ-GM-009 6-month window. `--strict` adoption is opt-in. CI failure → ad-hoc IF→WHEN rewrite per SPEC, not bulk. |
| MoAI-ADK contributors writing new SPECs | New SPEC template (provisional, out of scope) will use GEARS. This SPEC's spec.md §3 dogfoods GEARS to demonstrate the notation. | This SPEC IS the template demonstrator. |
| Downstream tools parsing MoAI SPEC files | If they pattern-match on `IF .* THEN`, they may miss new GEARS-authored REQs. | Migration guide §"For tool authors" section. (Run-phase content decision.) |

## 7. Acceptance Criteria

See `acceptance.md`. Summary:

- AC-GM-001 ↔ REQ-GM-001 + REQ-GM-007: legacy EARS REQs continue to pass `moai spec lint` (non-strict)
- AC-GM-002 ↔ REQ-GM-002 + REQ-GM-006: IF/THEN REQs emit `LegacyEARSKeyword` warning
- AC-GM-003 ↔ REQ-GM-003: GEARS REQs (WHEN/WHILE/WHERE/Ubiquitous well-formed) pass lint
- AC-GM-004 ↔ REQ-GM-004: 4-locale docs-site GEARS guide present + cross-linked
- AC-GM-005 ↔ REQ-GM-005: run-phase first step writes `.moai/research/gears-paper-validation.md` before lint.go edit
- AC-GM-006 ↔ REQ-GM-008: `moai spec lint --strict` exits non-zero on `LegacyEARSKeyword`
- AC-GM-007 ↔ REQ-GM-007: 88 existing SPECs unchanged by this SPEC (diff verification)
- AC-GM-008 ↔ REQ-GM-009: 6-month window policy documented in `internal/spec/lint.go` source comment + 4-locale migration guide

## 8. Risks + Mitigations

See `plan.md` § Risks. Five risks identified:

1. **GEARS paper notation mismatches §1 Table** — mitigated by Q3 run-phase web-research first-step blocker.
2. **88-SPEC backward-compat regression** — mitigated by AC-GM-007 diff verification + REQ-GM-009 window.
3. **4-locale translation drift** — mitigated by `.moai/docs/docs-site-i18n-rules.md` §17.3 HARD obligation + Hugo build PASS verification.
4. **Downstream tool breakage from new `LegacyEARSKeyword` finding code** — mitigated by NOT changing `ModalityMalformed`; new code is additive.
5. **Cross-Wave ordering (this SPEC Wave 6 planned before Waves 2-5 complete)** — mitigated by plan-phase being read-only / docs-only at Wave 6 entry; run-phase blocked on plan-auditor cross-Wave compatibility check (AC-GM-007).

## 9. Implementation Approach (Pointer)

See `plan.md` for the 4-milestone plan (M1: web-research validation, M2: lint.go GEARS support, M3: 4-locale docs sync, M4: regression guards + CI integration), risk-controlled rollout, and direct-orchestrator-vs-manager-develop delegation rationale.

## 10. Cross-References

- Blueprint: `.moai/research/v3-redesign-blueprint-2026-05-22.md` line 16, 24, 174, 200-206
- Design report: `.moai/research/v3.0-design-2026-05-22.md` §Wave 6 (line 200-206)
- Lint engine: `internal/spec/lint.go` lines 404-447 (EARSModalityRule)
- Schema SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md`
- EARS pattern docs: `.claude/skills/moai-foundation-core/modules/spec-ears-format.md`
- i18n HARD rules: `.moai/docs/docs-site-i18n-rules.md`
- Wave 6 sibling: `SPEC-V3R6-RULES-COMPLIANCE-001`, `SPEC-V3R6-V3-CUTOVER-001`
- Follow-up sweep (provisional): `SPEC-V3R6-GEARS-SWEEP-001`
