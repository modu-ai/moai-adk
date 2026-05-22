---
name: gears-paper-validation
description: GEARS paper validation artifact for SPEC-V3R6-GEARS-MIGRATION-001 M1 (run-phase first step per AC-GM-005)
created: 2026-05-22
updated: 2026-05-22
author: orchestrator (M1 paper validation)
related_spec: SPEC-V3R6-GEARS-MIGRATION-001
related_ac: AC-GM-005
---

# GEARS Paper Validation Report

> **M1 deliverable** for SPEC-V3R6-GEARS-MIGRATION-001 per `plan.md` §M1 (orchestrator-direct).
> **Acceptance gate**: `AC-GM-005` (run-phase first step, blocks M2 until verdict resolved).

## 1. Canonical Source URLs (verified via WebSearch + WebFetch 2026-05-22)

| # | URL | Type | Author | Date | Status |
|---|-----|------|--------|------|--------|
| 1 | https://dev.to/sublang/gears-the-spec-syntax-that-makes-ai-coding-actually-work-4f3f | DEV Community article (canonical) | Σ\* for SubLang | 2026-01-23 (edited 2026-04-05) | **PRIMARY** |
| 2 | https://medium.com/sublang/generalized-ears-the-ai-ready-spec-syntax-11ba36a37165 | Medium mirror (same content) | Alphabet / SubLang | 2026-01 | Mirror |
| 3 | sublang.ai | Original source domain referenced in DEV article | SubLang | n/a | Reference only |

**Blueprint citation context**: `.moai/research/v3-redesign-blueprint-2026-05-22.md` line 24 and 205 reference "EARS → GEARS evolution (DEV 2026)" without a URL. Run-phase first step has now resolved Open Question Q3 (spec.md §5) — the canonical reference is the DEV Community article above.

## 2. GEARS Acronym (verbatim from paper)

> **GEARS** = "Generalized Expression for AI-Ready Specs"
> Also expressed as: "generalized EARS or Generalized Expression for AI-Ready Specs"
> — DEV Community article §1 (verbatim quote)

## 3. GEARS Keyword Definitions (verbatim from paper)

The paper presents an **official syntax table** with the following verbatim definitions:

| Keyword | Verbatim definition from paper |
|---------|-------------------------------|
| **Where** | "Static precondition — configuration, feature flags, environment" |
| **While** | "Stateful precondition — a condition that must hold during execution" |
| **When** | "Trigger — the event that initiates behavior" |
| **shall** | "Required behavior — what the subject must do" |

## 4. Canonical GEARS Syntax Pattern (verbatim from paper)

```
[Where `<static precondition(s)>`]
[While `<stateful precondition(s)>`]
[When `<trigger>`]
The `<subject>` shall `<behavior>`
```

**Key properties from paper**:
- Square brackets `[]` denote **optional clauses** — all preconditions and triggers may be omitted.
- Multiple clauses may be **composed in a single requirement** (the paper's compound-clause feature).
- `<subject>` is **generalized** — replaces "the system" with any noun (system, component, service, agent, function, artifact).

**Example combining clauses (verbatim from paper)**:

> "Where the deployment is production, when a request fails, the service shall retry with exponential backoff."

## 5. IF/THEN Treatment (verbatim from paper)

> "The 'unwanted behavior' case deserves attention. EARS uses If...then to provide a visual signal that this is an edge case. GEARS drops this distinction. Structurally, error handling is just another trigger-response pair. The 'unwantedness' lives in the semantics, not the syntax. GEARS prioritizes AI processing over human visual scanning."

**Interpretation**: GEARS does NOT keep IF/THEN as a separate syntactic construct. Error handling is expressed as `When <event-detected>, the <subject> shall <action>`. The migration path from EARS `IF condition THEN action` to GEARS `WHEN condition-detected, ... shall action` is consistent with the paper's intent.

## 6. EARS vs GEARS Differences (verbatim from paper)

| Aspect | EARS | GEARS |
|--------|------|-------|
| Subject | "The system" (fixed) | `<subject>` — any noun (generalized) |
| Pattern count | 5 patterns (Ubiquitous, WHEN, WHILE, WHERE, IF/THEN) | 1 unified pattern with optional clauses |
| Error handling | Separate IF/THEN syntactic construct | Same trigger-response pattern (semantic distinction only) |
| WHERE semantics | "Feature exists" (narrowly framed) | Static precondition — configuration, feature flags, environment |
| WHILE semantics | "Conditional state" | Stateful precondition — runtime condition that may change |
| Test mapping | Indirect | Direct: Given → Where+While, When → When, Then → shall |

## 7. Test-Case Equivalence with Given-When-Then (verbatim from paper)

> "Test-case equivalence: The syntax maps directly to Given-When-Then, eliminating the need for separate specification and testing languages."

Mapping:
- **Given** → `Where` + `While` clauses
- **When** → `When` clause
- **Then** → `shall` clause

## 8. Side-by-Side Diff vs spec.md §1

| Item | spec.md §1 (current SPEC) | Paper (verbatim) | Status |
|------|---------------------------|------------------|--------|
| Acronym expansion | "Generative-AI-friendly Easy Approach to Requirements Syntax" | **"Generalized Expression for AI-Ready Specs"** | **❌ MISMATCH** — wrong expansion |
| Source attribution | "DEV 2026 proposal" (no URL) | DEV Community article 2026-01-23, author Σ\*/SubLang | ✅ Source verified |
| Ubiquitous notation | `The system shall <action>` (unchanged) | `The <subject> shall <behavior>` (generalized subject) | ⚠️ **PARTIAL MISMATCH** — spec.md keeps "The system"; paper generalizes |
| WHEN | `WHEN <event>, the system shall <action>` (unchanged) | Trigger clause: `When <trigger>` (composable with other clauses) | ✅ Functional match |
| WHILE | "promoted as primary state notation" | "Stateful precondition — a condition that must hold during execution" | ✅ Match (paper has more precise definition) |
| WHERE | "clarified — now also covers feature-flag gates and capability negotiation" | "Static precondition — configuration, feature flags, environment" | ✅ Match (paper is more concise) |
| IF/THEN | DEPRECATED — express as `WHEN <event-detected>` | Not "deprecated" syntactically — semantically absorbed into WHEN | ✅ Functional match — migration path aligns |
| Pattern structure | 4 separate keyword rows (Ubiquitous + WHEN + WHILE + WHERE) | **Single unified pattern** with optional `[Where] [While] [When]` clauses | ❌ **MISMATCH** — spec.md misses unified composition |
| Generalized subject | Not mentioned | **Headline feature** — subject is any noun | ❌ **MISMATCH** — spec.md misses entirely |
| Given-When-Then equivalence | Not mentioned | Major paper claim — Given→Where+While, When→When, Then→shall | ❌ **MISMATCH** — spec.md misses |

## 9. Verdict

**Verdict: MISMATCH**

**Summary**: The core migration decision (IF/THEN deprecation, WHEN/WHILE/WHERE retention, lint warning emission) is **paper-aligned**. The functional outcome of M2 (lint.go `LegacyEARSKeyword` finding for IF/THEN REQs) does not require amendment.

However, **spec.md §1 contains factual documentation defects** that WILL propagate into M3 (4-locale docs) and M4 (source comments) if uncorrected:

### Required Amendments to spec.md §1 (BEFORE M3 begins)

1. **Acronym expansion correction (BLOCKING for M3/M4 documentation)**:
   - Current: "Generative-AI-friendly Easy Approach to Requirements Syntax"
   - Replace with: "Generalized Expression for AI-Ready Specs" (verbatim paper definition)

2. **Reference URL addition (BLOCKING for M3 documentation citations)**:
   - Add canonical URL: https://dev.to/sublang/gears-the-spec-syntax-that-makes-ai-coding-actually-work-4f3f
   - Add author/date: Σ\*/SubLang, 2026-01-23 (edited 2026-04-05)

3. **Unified-pattern note (RECOMMENDED for M3 4-locale guide)**:
   - Add a §1.1 subsection documenting GEARS's compound-clause syntax: `[Where ...] [While ...] [When ...] The <subject> shall <behavior>`
   - This is informational; M2 lint behavior does not change.

4. **Generalized subject note (OPTIONAL, M3 documentation enhancement)**:
   - Note that GEARS allows any noun as subject; MoAI's 88 existing SPECs keep "The system" for backward compatibility (no migration required).

5. **Given-When-Then mapping note (OPTIONAL, M3 documentation enhancement)**:
   - Add a short paragraph mentioning that GEARS clauses map directly to Given-When-Then test syntax.

### M2 lint.go Code: NO CHANGE REQUIRED

The planned M2 implementation (per plan.md §3.1):
- `isLegacyEARSPattern()` detecting `IF ... THEN`
- Emitting `LegacyEARSKeyword` warning
- Source comment referencing the migration

remains **functionally correct** regardless of which acronym text appears in spec.md. The `isLegacyEARSPattern` regex does not depend on documentation text.

### M3 Docs and M4 Source Comments: AMENDMENT REQUIRED

Once amendments #1-#2 land on the plan-branch:
- M2 lint.go source comment can cite the corrected acronym + URL.
- M3 4-locale migration guide will document the correct acronym + reference paper.

## 10. Recommended Action

Per AC-GM-005 protocol:

> "If verdict == MISMATCH: a new commit on plan-branch amends spec.md §1 + plan.md M2 + acceptance.md AC-GM-003 fixture BEFORE M2 (lint.go edit) executes."

**Decision required from orchestrator**: Apply amendments #1 (acronym) and #2 (reference URL) NOW as a plan-branch commit before M2, OR proceed with M2 lint.go (functionally correct) and apply amendments at M3 entry (delayed correction).

Option B (delayed) carries the risk that M3 author cites the wrong acronym from spec.md without re-checking; Option A (immediate) is the AC-GM-005-prescribed flow.

## 11. References

- DEV Community article (canonical): https://dev.to/sublang/gears-the-spec-syntax-that-makes-ai-coding-actually-work-4f3f
- Medium mirror: https://medium.com/sublang/generalized-ears-the-ai-ready-spec-syntax-11ba36a37165
- EARS original (Mavin et al., 2009): https://alistairmavin.com/ears/
- LLM4RE Survey (arXiv 2509.11446) — projects LLM4RE will equal NLP4RE output by 2026
- Springer chapter on Advancing RE through Generative AI (2024): https://link.springer.com/chapter/10.1007/978-3-031-55642-5_6
- From Specifications to Prompts (arXiv 2408.09127) — Vogelsang on LLM prompting in RE

---

**End of validation report.**

Run-phase status: M1 COMPLETE with verdict MISMATCH. M2 lint.go is functionally unblocked but documentation amendments are recommended BEFORE M3 begins.
