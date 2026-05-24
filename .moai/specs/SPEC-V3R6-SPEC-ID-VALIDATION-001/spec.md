---
id: SPEC-V3R6-SPEC-ID-VALIDATION-001
title: "manager-spec SPEC ID Regex Pre-Write Self-Check (L51 Implementation)"
version: "0.2.0"
status: implemented
created: 2026-05-24
updated: 2026-05-24
author: MoAI
priority: P1
phase: "v3.0.0"
module: "internal/template/templates/.claude/agents/core"
lifecycle: spec-anchored
tags: "manager-spec, spec-id, regex, lint, self-check, drift-prevention, v3r6, tier-s, sprint-7, iter2-bundled"
---

# SPEC-V3R6-SPEC-ID-VALIDATION-001 — manager-spec SPEC ID Regex Pre-Write Self-Check (L51 Implementation)

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-24 | MoAI | Initial creation (plan-phase). L51 lesson implementation derived from Sprint 7 TMC-001 plan-phase 5-incident SPEC ID drift root cause analysis. Tier S minimal Section A-E variant. |
| 0.2.0 | 2026-05-24 | MoAI | iter-2 bundled revision per user-approved plan-auditor recommendation. Bundle D1+D2+D3+D4: (D1) +REQ-SIV-008 +AC-SIV-006 frontmatter 9→12 schema fix; (D2) +REQ-SIV-009 +AC-SIV-007 rule_template_mirror_test.go allowlist enrollment for `manager-spec.md`; (D3) REQ-SIV-007 verification path disambiguation (diff -q canonical, test confirmatory); (D4) REQ-SIV-004 wording lock-in for deterministic AC-SIV-004 grep. REQ 7→9, AC 5→7, M1 files 2→3. Tier envelope unchanged (~80-130 LOC actual ≤ 300 cap). D5/D6/D7 deferred per audit recommendation. |

## §A. Identity & Background

### §A.1 L51 origin

L51 lesson was promoted during Sprint 7 TMC-001 plan-phase (2026-05-24) after the **5th consecutive** SPEC ID format drift incident (`L32` chain). Each incident propagated silently from `manager-spec` Write call through memory files, progress.md placeholders, and paste-ready resume messages until `spec-lint`'s `FrontmatterInvalid` ERROR detection caught the format violation at PR time. The proposed remediation: introduce a **regex pre-write self-check protocol** inside the `manager-spec` agent body, validating every SPEC ID against the canonical regex literal from `internal/spec/lint.go:573` BEFORE the file Write occurs.

### §A.2 Five-incident drift timeline (L32 chain)

| # | SPEC ID attempted | Failure mode | Resolution path | Date |
|---|-------------------|--------------|-----------------|------|
| 1 | `SPEC-V3R6-CHANGELOG-CLEANUP-001` (typo variants) | manager-spec body regex (single-segment `[A-Z][A-Z0-9]+`) failed to validate multi-segment IDs | orchestrator-direct fix-forward via Edit + spec-lint re-run | 2026-05-23 |
| 2 | `SPEC-V3R6-CLI-AUDIT-001` (sub-ID confusion) | AC sub-ID convention `XXX-NNNa/b` (acceptance criteria) bled over into SPEC ID literal | orchestrator-direct fix-forward + Edit replace_all | 2026-05-23 |
| 3 | `SPEC-V3R6-LCL-003` (acronym ambiguity) | manager-spec body regex single-segment failed on long composite domain names | orchestrator-direct fix-forward + AskUserQuestion canonical selection | 2026-05-23 |
| 4 | `SPEC-V3R6-SARM-001` (acronym) | spec-frontmatter-schema.md doc regex vs internal/spec/lint.go:573 actual regex drift | orchestrator-direct fix-forward + 4-file canonical rename | 2026-05-24 |
| 5 | `SPEC-V3R6-TMC-002a` → `-001` (digit-alpha suffix) | manager-spec applied literal paste-ready `-002a` → `FrontmatterInvalid` ERROR (`\d{3}$` digit-only anchor violation) | orchestrator-direct fix-forward via `mv` + Edit replace_all × 4 + spec-lint re-run | 2026-05-24 |

Cumulative cost: 5 retry rounds × ~3 Edit/mv operations each = ~15 reactive fixup operations that would have been prevented by a single pre-write regex verification step inside `manager-spec` itself.

### §A.3 Regex drift evidence (smoking gun) — and the second, deeper drift surfaced by plan-auditor iter-1

The L32 chain has not one but **two** drifts inside the same `manager-spec.md` agent body. Iter-1 of plan-auditor (D1 defect) surfaced the second drift, which is structurally identical in nature to the regex drift but affects the **frontmatter field schema** rather than the regex literal. Both drifts share the same root cause: `manager-spec.md` was updated independently from the canonical SSOT and never reconciled.

#### §A.3.1 Drift 1 — Regex literal divergence (original L51 motivation)

| Location | Regex literal | Semantic coverage |
|----------|---------------|-------------------|
| `internal/spec/lint.go:573` (SSOT, lint enforcer) | `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` | Multi-segment domain (`SPEC-V3R6-SPEC-ID-VALIDATION-001`), `*` allows 0-char segments, `\d{3}$` digit-only anchor |
| `.claude/agents/core/manager-spec.md:158` (agent body, advisory) | `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$` | **Single-segment only** (`SPEC-AUTH-001`), `+` requires ≥1 char per segment, mishandles `SPEC-V3R6-...` |

The manager-spec body's single-segment regex cannot validate multi-domain SPEC IDs like the present `SPEC-V3R6-SPEC-ID-VALIDATION-001` correctly. Even when manager-spec's pre-write checklist asks "Does `id` match the regex?", the manager applies a regex that is structurally incapable of recognizing the canonical multi-segment form — silently accepting malformed IDs (or rejecting valid ones) and pushing the detection burden downstream to spec-lint at PR time.

#### §A.3.2 Drift 2 — Frontmatter field schema divergence (9 vs 12) [iter-2 bundled per D1]

| Location | Field count | Field names | Status |
|----------|-------------|-------------|--------|
| `.claude/rules/moai/development/spec-frontmatter-schema.md` (SSOT) + `internal/spec/lint.go` `FrontmatterSchemaRule` | **12** required | `id`, `title`, `version`, `status`, `created`, `updated`, `author`, `priority`, `phase`, `module`, `lifecycle`, `tags` | Canonical, enforced by lint |
| `.claude/agents/core/manager-spec.md` (mirror pair: local + template) | **9** required (drift) | `id`, `version`, `status`, `created_at`, `updated_at`, `author`, `priority`, `labels`, `issue_number` | Stale, contradicts canonical |

Affected lines in `manager-spec.md` (verbatim — 8 lines + 1 checklist line on the operational mirror; identical lines on the template mirror):

- **L117**: `**spec.md**: YAML frontmatter (9 required fields, see schema below), HISTORY section, EARS requirements, exclusions.`
- **L125**: `[HARD] Every \`spec.md\` YAML frontmatter MUST contain ALL 9 required fields below. ...`
- **L132**: `created_at: YYYY-MM-DD                   # Required. ISO date. NEVER use \`created\` (legacy, rejected)`
- **L133**: `updated_at: YYYY-MM-DD                   # Required. ISO date. NEVER use \`updated\` (legacy, rejected)`
- **L136**: `labels: [domain1, domain2, ...]          # Required. YAML array of lowercase tags. ...`
- **L137**: `issue_number: null                       # Required. Integer | null. ...` (issue_number is Optional per SSOT — see Optional Fields table)
- **L151**: `- \`created\` → must be \`created_at\``  ← **inverts canonical: canonical SSOT says use `created` (no suffix)**
- **L152**: `- \`updated\` → must be \`updated_at\``  ← **inverts canonical: canonical SSOT says use `updated` (no suffix)**
- **L157**: `1. All 9 required fields present` (Pre-write validation step 1)
- **L162**: `6. \`labels\` is a YAML array (not comma-separated string)` (Pre-write validation step 6; canonical uses `tags:` comma-separated string)
- **L175-L176** (Verification Checklist): `- [ ] \`created_at\` / \`updated_at\` used (NOT \`created\` / \`updated\`)` + `- [ ] \`labels\` array present (non-empty unless documented reason)` — both reverse the canonical convention.

The L125 "rejection table" actively rejects the canonical field names (`created`, `updated`) as "legacy" while accepting the snake_case aliases (`created_at`, `updated_at`) that the canonical SSOT (`spec-frontmatter-schema.md` §Rejected Snake_Case Aliases) explicitly identifies as rejected.

#### §A.3.3 Why both drifts share the same root cause class

Both drifts are **agent-body-vs-SSOT documentation desynchronization**. The agent body was edited at some point against an older or hypothesized schema and never reconciled when the canonical SSOT settled on the current form. Lint enforcement (`internal/spec/lint.go` `FrontmatterSchemaRule` + `specIDPattern`) catches the result at PR time, but by then the malformed artifact has already propagated through manager-spec Write → progress.md → memory → paste-ready resume. Both drifts are **the L51 lesson chain's root cause** — not two independent bugs but two faces of the same maintenance gap. Behaviour #2 (Manage Confusion Actively) requires surfacing both, not deferring one as "out of scope" while patching the other. Iter-2 closes both in a single bundled M1.

### §A.4 Why agent-body discipline rather than lint-only enforcement

Spec-lint (`internal/spec/lint.go` `FrontmatterSchemaRule`) is already the SSOT enforcer and catches every malformed ID — but only at PR time, after manager-spec has already written the file, propagated through memory files, and emitted paste-ready resume messages. Fixing the agent body regex literal is the **earliest possible** detection point (shift-left): the SPEC ID is validated in the same agent turn that decides to write it, BEFORE any filesystem mutation. This complements lint enforcement (defense-in-depth) without replacing it.

### §A.5 Mirror parity context

The `manager-spec` agent body exists at two paths that MUST remain byte-identical (template-mirror invariant, enforced by `internal/template/rule_template_mirror_test.go`):

```
.claude/agents/core/manager-spec.md                              (operational source, 9158 bytes)
internal/template/templates/.claude/agents/core/manager-spec.md  (template mirror, 9158 bytes)
```

Run-phase MUST edit both files in lockstep (template-first principle per CLAUDE.local.md §2 [HARD] Template-First Rule). Editing only one half breaks the mirror invariant test.

## §B. Scope

### §B.1 In-scope (3 files, run-phase only) [iter-2: expanded from 2 to 3]

**Run-phase artifacts** (plan-phase produces 4 SPEC artifacts only; run-phase modifies the 3 files below):

- **PRIMARY (template SSOT)**: `internal/template/templates/.claude/agents/core/manager-spec.md`
- **MIRROR (local copy)**: `.claude/agents/core/manager-spec.md`
- **TEST ALLOWLIST**: `internal/template/rule_template_mirror_test.go` [iter-2 D2 addition]

**Edit content on the 2 manager-spec.md mirror files** (byte-identical change to both):

1. **Insert** a new subsection "SPEC ID Pre-Write Self-Check Protocol" under the existing Pre-Creation Validation gate (currently around L121-165 of manager-spec.md)
2. **Replace** the regex literal at L158 (currently `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$`) with the canonical literal `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (verbatim match with `internal/spec/lint.go:573`)
3. **Add** AC sub-ID convention clarification text contrasting acceptance criterion sub-IDs (`AC-XXX-NNNa/b`, e.g., `AC-V3R6-001a`/`AC-V3R6-001b`) against SPEC IDs (which MUST end in `\d{3}` digit-only anchor, never `\d{3}[a-z]`)
4. **Add** a regex-match decomposition print directive [iter-2 D4 wording lock-in]: when manager-spec is about to Write a SPEC ID, it MUST print each segment match verification to its response body. The output MUST use one of the literal strings `decomposition` or `segment match trace` and MUST end with `→ PASS` or `→ FAIL` (e.g., `decomposition: SPEC ✓ | V3R6 ✓ | SPEC ✓ | ID ✓ | VALIDATION ✓ | 001 ✓ → PASS`).
5. **Optionally** reference the 5 historical drift incidents in a non-AC explanatory footnote
6. **[iter-2 D1 bundled] Frontmatter schema substitution** — replace the embedded 9-field schema block + rejection table + checklist with the canonical 12-field schema. Specifically:
   - L117 "9 required fields" → "12 canonical fields" (literal text replacement)
   - L125 "ALL 9 required fields" → "ALL 12 canonical fields" (literal text replacement; rewrite the YAML schema block at L127-138 to enumerate `id`, `title`, `version`, `status`, `created`, `updated`, `author`, `priority`, `phase`, `module`, `lifecycle`, `tags`)
   - L132-133 `created_at:` / `updated_at:` → `created:` / `updated:` (canonical names; remove "NEVER use `created` (legacy, rejected)" wording entirely)
   - L136 `labels: [...]` → `tags: "tag1, tag2"` (comma-separated string per canonical schema)
   - L137 `issue_number: null` → relocate to the Optional Fields list (Optional per SSOT, not Required)
   - L151-152 rejection table inversion: replace `created → must be created_at` / `updated → must be updated_at` with the canonical inversions `created_at → must be created` / `updated_at → must be updated` (and add `labels → must be tags`)
   - L157 "All 9 required fields present" → "All 12 canonical fields present"
   - L162 `labels` YAML array check → remove (canonical schema uses `tags:` comma-separated string)
   - L175-176 Verification Checklist: replace `created_at` / `updated_at` references with `created` / `updated`; remove `labels` array check, add `tags:` comma-separated string check.

**Edit content on `internal/template/rule_template_mirror_test.go`** [iter-2 D2 addition, ~1-3 LOC]:

Add `.claude/agents/core/manager-spec.md` to the `lateBranchMirroredPaths` slice (preferred — it is a one-off file added by this SPEC) so that `TestLateBranchTemplateMirror/manager-spec.md` becomes a real subtest that fires on parity drift. Without this enrollment, the existing `TestRuleTemplateMirrorDrift` and `TestLateBranchTemplateMirror` are vacuously green on manager-spec.md edits (neither test enumerates the path; AC-SIV-005's "if listed" qualifier silently held). This addition converts that vacuous green into an active assertion.

### §B.2 Out of Scope (deferred to future SPECs) [iter-2: frontmatter drift removed from this list, now in scope]

The following concerns are explicitly NOT in scope for this SPEC:

- **Lint rule code-level enforcement enhancement**: No changes to `internal/spec/lint.go` `FrontmatterSchemaRule` or `specIDPattern`. The lint enforcer is already correct; this SPEC fixes only the agent body discipline.
- **Audit of other agent bodies**: No regex review for `manager-develop.md`, `manager-docs.md`, `expert-*.md`, `meta/*.md`, etc. Scope is `manager-spec` only.
- **TEMPLATE-MIRROR-DRIFT-001 family sweep**: No proactive scan or fix of other template-mirror invariant test failures. This SPEC addresses only the `manager-spec.md` pair plus the one-line test allowlist enrollment.
- **Skill body audits**: `.claude/skills/moai-workflow-spec/SKILL.md` and similar SKILL.md files referencing SPEC ID format are NOT in scope (they are advisory; the agent body is the authoritative writer).
- **Memory file propagation correction**: Past memory files containing drifted SPEC IDs (e.g., `project_sprint2_tmc001_plan_ready.md` mentions of `-002a`) remain as historical record. No retroactive memory edits.
- **CHANGELOG.md historical entry corrections**: No edits to past CHANGELOG.md entries that reference drifted SPEC IDs (if any).
- **Documentation regex consistency in `spec-frontmatter-schema.md`**: `spec-frontmatter-schema.md` documents `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$` (single-segment) in its "Field Reference" table but `internal/spec/lint.go:573` uses multi-segment. This 3-way regex documentation drift remains a separate follow-up SPEC.
- **Other manager-spec.md content drift**: Any other unrelated documentation drift in `manager-spec.md` (Step 6 Expert Consultation, Status Responsibility Matrix wording, etc.) is NOT addressed.

## §C. EARS Requirements

### REQ-SIV-001 (Ubiquitous)

The `manager-spec` agent body **shall** include a section titled exactly "SPEC ID Pre-Write Self-Check Protocol" before any SPEC ID Write actions. The section MUST be present in both `.claude/agents/core/manager-spec.md` (operational source) and `internal/template/templates/.claude/agents/core/manager-spec.md` (template mirror) byte-identically.

### REQ-SIV-002 (Ubiquitous)

The regex literal embedded in `manager-spec.md` body for SPEC ID validation **shall** be verbatim-identical (character-by-character) to the canonical regex literal in `internal/spec/lint.go:573`, which is `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`. The legacy single-segment literal `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$` MUST be removed and replaced.

### REQ-SIV-003 (Ubiquitous)

The `manager-spec` agent body **shall** enumerate the AC sub-ID convention (acceptance criteria MAY use `AC-XXX-NNNa/b` form, e.g., `AC-V3R6-001a` and `AC-V3R6-001b` as paired sub-criteria) and **shall** explicitly contrast it with the SPEC ID digit-only anchor requirement (SPEC IDs MUST end in `\d{3}` exactly, NEVER `\d{3}[a-z]`).

### REQ-SIV-004 (Event-Driven) [iter-2 D4 wording lock-in]

**When** `manager-spec` is about to invoke `Write` for a new SPEC document containing a SPEC ID in its frontmatter, the agent **shall** print a regex match decomposition to its response body, showing each domain segment verification result (e.g., `decomposition: SPEC ✓ | V3R6 ✓ | SPEC ✓ | ID ✓ | VALIDATION ✓ | 001 ✓ → PASS`). The self-check output **shall** use one of the literal strings `decomposition` or `segment match trace` and **shall** end with `→ PASS` or `→ FAIL` to enable deterministic AC-SIV-004 grep verification. If any segment FAILS, the agent **shall** halt the Write and return a structured blocker report to the orchestrator instead.

### REQ-SIV-005 (Optional)

`manager-spec` **may** reference the 5 historical drift incidents (CHANGELOG-CLEANUP-001, CLI-AUDIT-001, LCL-003, SARM-001, TMC-001) in a non-AC explanatory footnote within the new self-check section. This reference is informational only; absence does not constitute a SPEC failure. **No AC coverage** for this REQ per §C.1 decision rule below.

### REQ-SIV-006 (Unwanted Behavior — Negative)

`manager-spec` **shall not** use the legacy single-segment regex `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$` anywhere in its agent body after this SPEC is implemented. Any occurrence of the legacy literal is a regression.

### REQ-SIV-007 (Ubiquitous — Mirror Invariant) [iter-2 D3 disambiguation]

Both files (`.claude/agents/core/manager-spec.md` and `internal/template/templates/.claude/agents/core/manager-spec.md`) **shall** remain byte-identical after the edit. Mirror parity verification has two complementary paths:

- **Canonical (primary)**: `diff -q` between the source and mirror files. Empty output (exit code 0) is the SSOT signal of byte-equality. This is the deterministic, format-agnostic check.
- **Confirmatory (supplementary)**: `TestLateBranchTemplateMirror/manager-spec.md` in `internal/template/rule_template_mirror_test.go`. This test BECOMES ACTIVE only after REQ-SIV-009 enrolls `manager-spec.md` into the `lateBranchMirroredPaths` allowlist; before iter-2 it was vacuously absent ("if listed" qualifier).

AC-SIV-005 covers both paths (canonical `diff -q` + `go vet` + `golangci-lint`); AC-SIV-007 covers the test activation independently.

### REQ-SIV-008 (Ubiquitous — Frontmatter Schema Alignment) [iter-2 D1 bundled, NEW]

The `manager-spec` agent body **shall** reference exactly the 12 canonical frontmatter fields enumerated in `.claude/rules/moai/development/spec-frontmatter-schema.md` (`id`, `title`, `version`, `status`, `created`, `updated`, `author`, `priority`, `phase`, `module`, `lifecycle`, `tags`) and **shall not** reference snake_case aliases (`created_at`, `updated_at`, `labels`) as legacy, required, or otherwise instructional text. The literal phrases "9 required fields" and "ALL 9 required fields" **shall** be replaced by "12 canonical fields" or equivalent literal forms. The rejection table at L150-153 **shall** invert: canonical schema rejects `created_at`/`updated_at`/`labels` aliases (per spec-frontmatter-schema.md §Rejected Snake_Case Aliases), not the reverse.

### REQ-SIV-009 (Ubiquitous — Test Allowlist Enrollment) [iter-2 D2 bundled, NEW]

`internal/template/rule_template_mirror_test.go` **shall** enumerate `.claude/agents/core/manager-spec.md` in the appropriate allowlist (`lateBranchMirroredPaths` is the canonical placement — it is the slice for late-added one-off mirrors) such that `TestLateBranchTemplateMirror/manager-spec.md` becomes a real subtest that fires on parity drift. Without this enrollment, the existing test is vacuously green on manager-spec.md edits and AC-SIV-005's "if listed" qualifier hides the gap.

### §C.1 Decision rule (SSOT canonical) [iter-2 updated counts]

Per Sprint 2 P4 lessons (L48 spec.md SSOT canonical), REQ-SIV-005 is Optional MAY without AC coverage; AC-SIV-001..007 cover REQ-SIV-001..004 + REQ-SIV-006 (binary regression check via grep) + REQ-SIV-007 (mirror invariant via `diff -q` canonical, integrated into AC-SIV-005 quality gate) + REQ-SIV-008 (D1 frontmatter schema, AC-SIV-006) + REQ-SIV-009 (D2 test allowlist enrollment, AC-SIV-007). REQ count = **9**, AC count = **7** (1-to-1 mapping skips REQ-SIV-005 Optional + folds REQ-SIV-006/007 into existing ACs; REQ-SIV-008 and REQ-SIV-009 each get a dedicated AC).

## §D. Decisions & Rationale

### §D.1 Why agent-body level intervention vs lint-only enforcement

Spec-lint already catches every malformed SPEC ID at PR time (existing safety net, not weakening). The intervention point selection question is: **earliest detection** vs **simplest implementation**. The 5-incident L32 chain demonstrates that lint-only enforcement allows the malformed ID to propagate through:

1. `manager-spec` Write call (file created with bad ID)
2. `progress.md` Audit-Ready Signal (placeholder filled with bad ID)
3. Memory file emission (`project_sprint*_*.md` snapshot containing bad ID)
4. paste-ready resume message (next session resumes with bad ID baked in)
5. spec-lint detection at PR time (`FrontmatterInvalid` ERROR)

Adding a single pre-write self-check inside `manager-spec` body short-circuits steps 1-4 entirely. The trade-off is: agent-body discipline is advisory (relies on the agent following its own checklist), while lint enforcement is mechanical. Both layers are valuable — defense-in-depth.

### §D.2 Why Tier S minimal classification

This SPEC is **Tier S** (≤300 LOC, ≤5 files, ~30-80 line addition to agent body + 1-line regex literal swap × 2 mirror files). The Section A-E template applies in its minimal form per `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability. Tier S plan-auditor PASS threshold is 0.75 per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier.

Precedents in Sprint 2 P4 trio (IVB-001 + SARM-001 + TMC-001) and Sprint 7 entry TMD-001 all used Tier S minimal Section A-E with 4 artifacts (spec/plan/acceptance/progress) and achieved 1-pass success. This SPEC follows the same pattern.

### §D.3 Why mirror parity is required

The template-first principle (CLAUDE.local.md §2 [HARD]) mandates that any change to `.claude/agents/`, `.moai/`, or `.agency/` content must be propagated to `internal/template/templates/` BEFORE the local file is modified. The `internal/template/rule_template_mirror_test.go` test enforces this by comparing source and mirror byte-by-byte. Editing only one file would FAIL the mirror invariant test. Both files MUST be edited in lockstep.

### §D.4 Why a regex-decomposition print directive (REQ-SIV-004)

Merely embedding the correct regex in the agent body is insufficient — the agent must **demonstrate** that it applied the regex correctly. The print directive forces the agent to externalize its check in its response body, providing an auditable trail. If the agent skips the print, the orchestrator can detect the discipline lapse during code review and corrective re-delegation. This is similar to TDD's RED-phase explicit failure proof.

### §D.5 Why bundling frontmatter field-count drift (9 vs 12) in iter-2 was the correct call [iter-2 reversed from iter-1]

Iter-1 deferred the 9-vs-12 frontmatter drift as a follow-up SPEC, on the theory that it was a separate "documentation accuracy" concern orthogonal to the "semantic correctness" of the regex drift. Plan-auditor iter-1 (defect D1) rejected this framing on Behavior #2 grounds (Manage Confusion Actively): both drifts are agent-body-vs-SSOT desynchronization with identical root cause class, and surfacing only one while patching the other contradicts L51's stated remediation goal (shift-left detection of all manager-spec authorship drift).

Bundled iter-2 scope decision:

- **Cost of bundling**: +30-50 LOC inside the SAME manager-spec.md edit pair (8 affected lines per file × 2 files = ~16 line replacements + supporting structural changes). Total per-file edit grows from ~35-55 lines to ~65-100 lines — still well within the Tier S envelope (~80-130 LOC across 3 files ≤ 300 LOC cap).
- **Cost of NOT bundling**: separate Sprint 8+ SPEC with its own plan-phase + run-phase + sync-phase + mx-phase = 4 additional commits + plan-auditor cycle. The fix is mechanically trivial (literal text replacement) but the lifecycle overhead is fixed per SPEC.
- **Risk of NOT bundling**: until the field-count drift is fixed, manager-spec continues to instruct future authors to use `created_at`/`updated_at`/`labels` and reject `created`/`updated`/`tags` — producing exactly the kind of FrontmatterInvalid lint findings that L51 was promoted to prevent. Every new SPEC authored before the follow-up SPEC merges is at risk of inheriting the drift.

Conclusion: bundle in iter-2. The Tier S envelope holds; the L51 remediation is complete in a single SPEC.

### §D.6 Why adding REQ-SIV-009 (test allowlist enrollment) is structural, not cosmetic [iter-2 D2]

Plan-auditor iter-1 D2 surfaced that `TestLateBranchTemplateMirror` and `TestRuleTemplateMirrorDrift` do NOT enumerate `.claude/agents/core/manager-spec.md`. Iter-1's AC-SIV-005 quality gate cited the test with the qualifier "if listed" — meaning if it happened not to be listed (which it was not), the AC would pass vacuously. This is a test-vacuity defect: AC claims coverage that does not exist.

Two options were offered by plan-auditor:

- **Option A (chosen by user)**: Add `manager-spec.md` to the allowlist in iter-2 (+REQ-SIV-009 +AC-SIV-007). Cost: +1-3 LOC at the test file. Benefit: AC-SIV-005 confirmatory path becomes real; future manager-spec.md edits are guarded by CI.
- **Option B**: Drop the test reference from AC-SIV-005 entirely and rely only on `diff -q`. Cost: zero LOC change. Benefit: AC honesty. Cost: loses the CI guard for future regressions.

Option A was selected because the marginal cost is trivial (one slice entry) and the marginal benefit is durable (CI guard against future drift). Option B was rejected because losing the CI guard reopens the L51 problem space.

### §D.7 Why deferring D5/D6/D7 is acceptable [iter-2]

D5 (EARS Unwanted form syntactic precision), D6 (progress.md placeholder cosmetic polish), D7 (FAIL-path output ordering specification) are all MINOR cosmetic defects that do not affect:

- Lint correctness (REQ-SIV-006 + spec-lint guarantee binary regression detection)
- AC coverage (D5-D7 do not introduce gaps in REQ-to-AC mapping)
- Mirror parity (D5-D7 do not touch the manager-spec.md edit content)

Plan-auditor iter-1 recommendation explicitly deferred D5/D6/D7 as non-blocking polish. Iter-2 honors that recommendation to preserve scope discipline (Behavior #5) and Tier S envelope.

## §E. References

### §E.1 Code references

- `internal/spec/lint.go:573` — canonical regex literal `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (SSOT)
- `internal/spec/lint.go` `FrontmatterSchemaRule` — lint enforcement implementing REQ-SPC-003-006
- `internal/template/rule_template_mirror_test.go` — template-mirror invariant test (existing safety net)

### §E.2 Documentation references

- `.claude/rules/moai/development/spec-frontmatter-schema.md` — canonical 12-field frontmatter schema SSOT
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier S/M/L classification
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability — Tier S minimal template form
- `CLAUDE.local.md` §2 [HARD] Template-First Rule — mirror parity discipline

### §E.3 Memory references

- L51 lesson promotion: `project_sprint2_tmc001_plan_ready.md` (Sprint 7 TMC-001 plan-phase, 2026-05-24)
- L32 chain context: 5 incidents documented in `project_sprint2_*` memory files (2026-05-23..2026-05-24)

### §E.4 Confusion case (illustrative)

The retired `SPEC-RETIRED-DDD-001` case (if encountered in old memory) was confusingly close to the SPEC ID format but uses `RETIRED` as a domain segment — VALID per `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` because `RETIRED` matches `[A-Z][A-Z0-9]*` and `001` matches `\d{3}`. This is a useful case to include in the new self-check section as a "valid but unusual" example.
