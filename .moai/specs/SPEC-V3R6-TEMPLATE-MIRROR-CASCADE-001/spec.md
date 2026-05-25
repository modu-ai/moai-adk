---
id: SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001
title: "Template mirror cascade: spec-assembly.md drift fix"
version: "0.1.1"
status: completed
created: 2026-05-24
updated: 2026-05-25
author: GOOS행님
priority: P3
phase: "v3.0.0"
module: "internal/template/templates/.claude/skills/moai/workflows/plan"
lifecycle: spec-anchored
tags: "template-mirror, cascade, drift-fix, tier-s, sprint-2-p4-3"
---

# SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001 — Template mirror cascade: spec-assembly.md drift fix

## §A. Why this SPEC

### §A.1 Problem statement

The template mirror at `internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md` (25,939 bytes / 516 lines) lags behind the operational source at `.claude/skills/moai/workflows/plan/spec-assembly.md` (28,423 bytes / 548 lines) by **2,484 bytes / 32 lines**. The mirror predates the LEAN workflow's Phase 1.6 (Tier Judgment Socratic Question) addition; the template-mirror invariant test `TestLateBranchTemplateMirror/spec-assembly.md` fails as a pre-existing baseline failure attributable to the TEMPLATE-MIRROR-DRIFT-001 family.

### §A.2 Drift evidence (verified 2026-05-24)

```bash
$ wc -c -l .claude/skills/moai/workflows/plan/spec-assembly.md \
           internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md
     548   28423 .claude/skills/moai/workflows/plan/spec-assembly.md
     516   25939 internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md
```

Diff content (`diff <source> <mirror>` summary):

| Source line range | Content (missing from mirror) | Line count |
|-------------------|-------------------------------|------------|
| 29-46 | "Phase 1.6: Tier Judgment Socratic Question (LEAN Workflow)" intro + Skip condition + AskUserQuestion block (Options 1-4) | 18 lines |
| 48-61 | LOC threshold guidance + tier-conditional artifact set rules (Tier S/M/L) + anti-pattern note + cross-reference | 14 lines |
| Total | | **32 lines** |

Source `.claude/skills/moai/workflows/plan/spec-assembly.md` is the canonical (operational) version. Mirror `internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md` is the lagging (template-only) copy that ships with `moai init`.

### §A.3 Test signal (pre-fix FAIL evidence)

```
$ go test ./internal/template/ -run 'TestLateBranchTemplateMirror/spec-assembly' -v
=== RUN   TestLateBranchTemplateMirror/spec-assembly.md
    rule_template_mirror_test.go:182: RULE_TEMPLATE_MIRROR_DRIFT: source file
        .claude/skills/moai/workflows/plan/spec-assembly.md differs from its mirror
        at internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md
        (source 28423 bytes, mirror 25939 bytes); run 'cp <src> <mirror>' and stage both files
--- FAIL: TestLateBranchTemplateMirror/spec-assembly.md (0.00s)
FAIL    github.com/modu-ai/moai-adk/internal/template    0.544s
```

### §A.4 L46 attribution (TEMPLATE-MIRROR-DRIFT-001 family)

The originating SPEC that added Phase 1.6 documentation to the source — without propagating to the template mirror — is **SPEC-V3R5-WORKFLOW-LEAN-001** (Tier judgment Socratic question introduction). This is a TEMPLATE-MIRROR-DRIFT-001-family case. The TEMPLATE-MIRROR-DRIFT-001 master SPEC (which would systematically sweep all template mirror drifts) is deferred to Sprint 7+; this SPEC clears only the `spec-assembly.md` cascade as the smallest-first ordering tail of Sprint 2 P4.

## §B. What scope (and what is explicitly out-of-scope)

### §B.1 In-scope (1 file, mechanical overwrite)

- **Single file modified**: `internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md`
- **Operation**: complete content overwrite from `.claude/skills/moai/workflows/plan/spec-assembly.md` (byte-for-byte copy)
- **Expected delta**: +32 lines, +2,484 bytes in the mirror; **0 changes in the source**

### §B.2 Out of Scope (deferred to future SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001 master SPEC)

The following 17+ TEMPLATE-MIRROR-DRIFT-family pre-existing baseline failures and related concerns are explicitly NOT in scope:

- **Other rule/template mirror drifts** discovered by `TestRuleTemplateMirrorDrift` × 3 (e.g., zone-registry.md, manager-develop-prompt-template.md per RULES-PATH-SCOPE-001 §1.4.1 family) — covered by future master SPEC
- **Other LateBranchTemplateMirror entries** beyond `spec-assembly.md` — covered by future master SPEC
- **Catalog/agent-folder drifts** flagged by `TestBackwardCompatibility`, `TestAgentFrontmatterAudit`, `TestAllAgentsInCatalog`, `TestEmbeddedTemplates_AgentDefinitions`, `TestLoadCatalog`, `TestLoadEmbeddedCatalog_Success` — orthogonal subsystem, separate cleanup track
- **Retirement assertion mismatches** flagged by `TestRetirementCompletenessAssertion` × 2 — separate retirement-tracking SPEC
- **Operational source modification** — `REQ-TMC-002` forbids any change to `.claude/skills/moai/workflows/plan/spec-assembly.md`; only the mirror is updated
- **Re-architecting the template-mirror invariant test** itself — `TestLateBranchTemplateMirror` continues to operate as designed; this SPEC clears one of its current failing rows

## §C. Decision rules

### §C.1 SSOT hierarchy

- `spec.md` is the canonical SSOT for REQ + AC. Every REQ-TMC ID is anchored here.
- `plan.md` is the derived implementation artifact (edit map + tier annotation). When the edit map conflicts with spec.md REQ wording, **spec.md wins** (L48 discipline).
- `acceptance.md` is the canonical AC enumeration (PASS/FAIL gates). REQs Covered column links each AC back to its REQ-TMC anchor in spec.md.
- `progress.md` is the runtime evidence log (4-phase Lifecycle Status table + Audit-Ready Signal).

### §C.2 Tier S minimal Section A-E justification

This SPEC qualifies for the Tier S minimal Section A-E variant per `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability:

- **Scope**: 1 file modified, ~32 line additions, mechanical content overwrite — well within Tier S envelope (≤300 LOC, ≤5 files, ≤1 milestone).
- **Risk**: minimal (no production code change, no API surface change, no behavior change in any runtime path; the mirror file is consumed only by `moai init` and `moai update` deployment paths and by the template-mirror invariant test).
- **Verification**: single `go test ./internal/template/ -run 'TestLateBranchTemplateMirror/spec-assembly'` pass plus `wc -c` + `diff` evidence.

Per Section A-E variant precedent established by SPEC-V3R6-I18N-VALIDATOR-BUDGET-001 (Sprint 2 P4.1, merged `d3ed4727d`) and SPEC-V3R6-SKILLS-AUDIT-RUN-MD-001 (Sprint 2 P4.2, merged `5e0dc6a9b`), Tier S minimal retains the 4-artifact form (spec/plan/acceptance/progress) for traceability rather than the Tier S strict 2-artifact form. This is a documented pattern (L33 in MEMORY.md), not a deviation.

## §D. Lint surface

- **spec.md MUST emit `✓ No findings`** when checked by `go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001/spec.md`. The Out-of-Scope section (§B.2 above) is canonical only in spec.md.
- **plan.md / acceptance.md / progress.md** MAY emit `MissingExclusions` ERROR when lint-checked individually. This is accepted derived-artifact lint surface per IVB-001 and SARM-001 precedent (CI does not block on derived-artifact lint per `git-strategy.yaml` pattern). The canonical lint surface is `spec.md` only.
- Frontmatter canonical 12 fields enforced across all 4 artifacts. `tags:` MUST be CSV-string form (`tags: "a, b, c"`), not YAML array — array form causes `ParseFailure` in `internal/spec/lint.go` `SPECFrontmatter.Tags string yaml:"tags"` binding (B-1 historical IVB-001 finding).

## §E. Sprint context

### §E.1 Sprint 2 P4 trio position

| Phase | SPEC ID | Tier | Scope | Status |
|-------|---------|------|-------|--------|
| P4.1 | SPEC-V3R6-I18N-VALIDATOR-BUDGET-001 | S minimal | i18n-validator budget 30s→35s (clears LCL-001 AC-LCL-005 PASS-WITH-DEBT) | complete `d3ed4727d` |
| P4.2 | SPEC-V3R6-SKILLS-AUDIT-RUN-MD-001 | S minimal | solo run.md test path → workflows/run/phase-execution.md | complete `5e0dc6a9b` |
| **P4.3** | **SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001** | **S minimal** | **spec-assembly.md template mirror drift cascade** | **draft (this SPEC)** |

P4.3 is the smallest-first ordering tail of Sprint 2 P4. After this SPEC completes 4-phase lifecycle (plan → run → sync → mx; mx likely SKIP per `.claude/rules/moai/workflow/mx-tag-protocol.md` §a for template-only edits), Sprint 2 P4 trio closes.

### §E.2 Post-merge follow-up

Master SPEC `SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001` (systematic sweep of all 17+ remaining mirror drifts + catalog/agent-folder drifts) is deferred to Sprint 7. Sprint 2 closes with P4.3 merge; Sprint 7 entry point will be the master TEMPLATE-MIRROR-DRIFT-001 SPEC if user prioritizes it over other backlog SPECs (CLI-INTEGRATION-001 / PROMPT-CACHE-001).

## §F. Requirements (EARS Format)

### REQ-TMC-001 (Ubiquitous, mandatory)

**WHEN** `/moai run SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001` executes M1, the mirror file `internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md` **shall** be byte-identical to the source `.claude/skills/moai/workflows/plan/spec-assembly.md` after the fix commit (post-edit `diff <src> <mirror> | wc -l` = 0).

### REQ-TMC-002 (Unwanted Behavior, mandatory)

**While** manager-develop executes the fix, the operational source `.claude/skills/moai/workflows/plan/spec-assembly.md` **shall not** be modified (no operational regression — source remains byte-identical pre/post; `git diff HEAD~1..HEAD -- .claude/skills/moai/workflows/plan/spec-assembly.md` = 0 lines).

### REQ-TMC-003 (State-Driven, mandatory)

**WHEN** `go test ./internal/template/ -run 'TestLateBranchTemplateMirror/spec-assembly' -v` is invoked post-fix, the test **shall** PASS (the pre-fix `RULE_TEMPLATE_MIRROR_DRIFT` FAIL signal at `rule_template_mirror_test.go:182` is cleared).

### REQ-TMC-004 (Unwanted Behavior, mandatory)

**WHILE** other dirty/untracked files exist in the working tree (PRESERVE list per L45 — `.moai/config/sections/{git-convention,language,quality}.yaml`, `.moai/harness/usage-log.jsonl`, `.moai/harness/observations.yaml`, `.moai/research/v3.0-redesign-2026-05-23.md`, `i18n-validator`, plus the 2 emergent template mirrors `.claude/output-styles/moai/{einstein,moai}.md` and their template counterparts under `internal/template/templates/.claude/output-styles/moai/`), they **shall** remain unchanged across the fix commit.

### REQ-TMC-005 (State-Driven, mandatory)

**WHEN** `go vet ./...` and `golangci-lint run --timeout=2m` are invoked post-fix, the baseline 0-issue state **shall** be preserved (no new vet warnings, no new lint findings introduced).

### REQ-TMC-006 (Event-Driven, L46 attribution discipline)

**WHEN** other template mirror tests are run post-fix (`TestRuleTemplateMirrorDrift` × 3, `TestLateBranchTemplateMirror` entries other than `spec-assembly.md`, `TestBackwardCompatibility`, `TestAgentFrontmatterAudit`, `TestAllAgentsInCatalog`, `TestEmbeddedTemplates_AgentDefinitions`, `TestLoadCatalog`, `TestLoadEmbeddedCatalog_Success`, `TestRetirementCompletenessAssertion` × 2), the pre-existing baseline failures **shall** persist exactly as before — this SPEC fixes ONLY `spec-assembly.md` mirror; other drifts are deferred to the future master `SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001`.

### REQ-TMC-007 (Optional, MAY)

The fix MAY use `git add -p` or path-specific `git add <mirror-path>` discipline to ensure ONLY the mirror file is staged (no accidental source or sibling staging). The minimal-change discipline supports either path-specific `git add` of just the mirror (REQ-TMC-007 satisfied) or a wider `git add .` followed by un-staging unintended paths — the former is preferred for traceability.
