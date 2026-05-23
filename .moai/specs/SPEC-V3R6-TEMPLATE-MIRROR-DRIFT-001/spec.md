---
id: SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001
title: "Template mirror drift cleanup: 4-file mechanical mirror parity (Sprint 7 entry)"
version: "0.1.0"
status: draft
created: 2026-05-24
updated: 2026-05-24
author: GOOS행님
priority: P3
phase: "v3.0.0"
module: "internal/template/templates/.claude"
lifecycle: spec-anchored
tags: "template-mirror, drift-fix, sprint-7-entry, tier-s, mechanical-cleanup"
---

# SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001 — Template mirror drift cleanup: 4-file mechanical mirror parity (Sprint 7 entry)

## §A. Why this SPEC

### §A.1 Problem statement

Sprint 2 P4 trio (IVB-001 + SARM-001 + TMC-001) closed on `38a638d3c` with **10 baseline failures** (6 categories) still present in the post-merge test surface. Sprint 7 entry SPEC is the **mechanical 4-mirror cleanup** of A1-A4 — the only category that can be cleared by byte-for-byte content overwrite without policy decisions.

The 4 template mirrors at `internal/template/templates/` lag behind their `.claude/` operational sources because intermediate SPECs modified the source but did not propagate to the template mirror. The mirrors are consumed by `moai init` and `moai update` deployment paths and by the mirror-invariant tests `TestRuleTemplateMirrorDrift` (3 of 4 failing) and `TestLateBranchTemplateMirror`. Without mirror sync, user projects that run `moai init` receive stale `.claude/` rules.

### §A.2 Drift evidence (verified 2026-05-24)

```
$ wc -c -l <source> <mirror>
  450  29363 .claude/rules/moai/workflow/spec-workflow.md
  428  26709 internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md
  372  15868 .claude/rules/moai/core/agent-common-protocol.md
  324  13599 internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md
  463  26469 .claude/agents/meta/plan-auditor.md
  433  24205 internal/template/templates/.claude/agents/meta/plan-auditor.md
  370  18602 .claude/rules/moai/core/hooks-system.md
  357  17857 internal/template/templates/.claude/rules/moai/core/hooks-system.md
```

| Source | Mirror | Source bytes | Mirror bytes | Delta |
|--------|--------|--------------|--------------|-------|
| spec-workflow.md | template/templates/.../spec-workflow.md | 29363 | 26709 | +2654 |
| agent-common-protocol.md | template/templates/.../agent-common-protocol.md | 15868 | 13599 | +2269 |
| plan-auditor.md | template/templates/.../plan-auditor.md | 26469 | 24205 | +2264 |
| hooks-system.md | template/templates/.../hooks-system.md | 18602 | 17857 | +745 |

### §A.3 Test signal (pre-fix FAIL evidence)

```
$ go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift' -v 2>&1 | tail -10
--- FAIL: TestRuleTemplateMirrorDrift (0.00s)
    --- PASS: TestRuleTemplateMirrorDrift/manager-develop-prompt-template.md
    --- PASS: TestRuleTemplateMirrorDrift/verification-batch-pattern.md
    --- PASS: TestRuleTemplateMirrorDrift/default.md
    --- PASS: TestRuleTemplateMirrorDrift/agent-teams-pattern.md
    --- PASS: TestRuleTemplateMirrorDrift/frontend.md
    --- PASS: TestRuleTemplateMirrorDrift/ci-watch-protocol.md
    --- FAIL: TestRuleTemplateMirrorDrift/plan-auditor.md
    --- FAIL: TestRuleTemplateMirrorDrift/agent-common-protocol.md
    --- FAIL: TestRuleTemplateMirrorDrift/spec-workflow.md
```

`TestRuleTemplateMirrorDrift/hooks-system.md` does NOT currently exist as a subtest — the entry must be added to `rule_template_mirror_test.go` `workflowOptMirroredPaths` allowlist (REQ-TMD-005).

### §A.4 L46 attribution (TEMPLATE-MIRROR-DRIFT-001 family)

The originating SPECs that added content to the sources without propagating to template mirrors are sibling SPECs in the SPEC-V3R5-* / SPEC-V3R6-* series (precise attribution per file is L46 follow-up; this SPEC clears the cascade without re-litigating the root cause). The TMC-001 precedent (Sprint 2 P4.3, merged `38a638d3c`) cleared `spec-assembly.md` as the smallest-first ordering tail; this Sprint 7 entry continues the cleanup with the 4 next-smallest mirrors.

## §B. What scope (and what is explicitly out-of-scope)

### §B.1 In-scope (5 files exactly — L40 Tier S envelope ≤5 files)

| # | Source | Target (mirror) | Operation | Expected delta |
|---|--------|-----------------|-----------|----------------|
| A1 | `.claude/rules/moai/workflow/spec-workflow.md` | `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` | byte-for-byte cp overwrite | mirror +2654 bytes / +22 lines |
| A2 | `.claude/rules/moai/core/agent-common-protocol.md` | `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` | byte-for-byte cp overwrite | mirror +2269 bytes / +48 lines |
| A3 | `.claude/agents/meta/plan-auditor.md` | `internal/template/templates/.claude/agents/meta/plan-auditor.md` | byte-for-byte cp overwrite | mirror +2264 bytes / +30 lines |
| A4 | `.claude/rules/moai/core/hooks-system.md` | `internal/template/templates/.claude/rules/moai/core/hooks-system.md` | byte-for-byte cp overwrite | mirror +745 bytes / +13 lines |
| A4b | `internal/template/rule_template_mirror_test.go` | (same file — registry add) | new entry `.claude/rules/moai/core/hooks-system.md` inserted into `workflowOptMirroredPaths` slice between line 48 (`agent-common-protocol.md`) and line 50 (`spec-workflow.md`) — within the `core/` group | +1 line + 1 comment line (~+2 lines net) |

Total scope: **5 files modified**, **0 files created**, **all PRESERVE list 11 dirty/untracked entries** unchanged. Net source delta = 0 (sources untouched per REQ-TMD-006); net mirror delta = +7932 bytes / +113 lines; net test-go delta = +2 lines.

### §B.2 Out of Scope (deferred to Sprint 8 or follow-up SPECs)

The following categories of baseline failures are explicitly NOT in scope. Each requires a policy decision (per CLAUDE.local.md §24 namespace policy or §21 dev-only) that this mechanical SPEC does not adjudicate:

- **B1 — Retirement assertion mismatches** (`TestRetirementCompletenessAssertion` × 2): requires retirement-tracking SPEC; not a mechanical mirror fix
- **B2 — Catalog / agent-folder drifts** (`TestBackwardCompatibility`, `TestAgentFrontmatterAudit`, `TestAllAgentsInCatalog`, `TestEmbeddedTemplates_AgentDefinitions`, `TestLoadCatalog`, `TestLoadEmbeddedCatalog_Success`): orthogonal subsystem, requires CLAUDE.local.md §24 namespace policy decision (e.g., `harness/` subdirectory boundaries)
- **B3 — Frontmatter audit drifts** (`TestAgentFrontmatterAudit`): requires schema reconciliation per `.claude/rules/moai/development/spec-frontmatter-schema.md` SSOT
- **B4 — Embedded catalog drifts** (`TestEmbeddedTemplates_AgentDefinitions`, `TestLoadEmbeddedCatalog_Success`): requires `make build` regeneration scope decision (per CLAUDE.local.md §2 Template-First Rule)
- **B5 — Other rule mirrors** (zone-registry.md, askuser-protocol.md, etc.) that may have drifted post-merge: requires future systematic sweep
- **B6 — Re-architecting the template-mirror invariant test itself**: `TestRuleTemplateMirrorDrift` and `TestLateBranchTemplateMirror` continue to operate as designed
- **F1 — Dev-only file confirmation** (CLAUDE.local.md §21 97/98/99 series): requires explicit user audit decision
- **Operational source modification**: REQ-TMD-006 forbids any change to the 4 `.claude/` sources; only mirrors and the test registry are modified

## §C. Decision rules

### §C.1 SSOT hierarchy

- `spec.md` is the canonical SSOT for REQ + AC. Every REQ-TMD ID is anchored here.
- `plan.md` is the derived implementation artifact (edit map + tier annotation). When the edit map conflicts with spec.md REQ wording, **spec.md wins** (L48 discipline).
- `acceptance.md` is the canonical AC enumeration (PASS/FAIL gates). REQs Covered column links each AC back to its REQ-TMD anchor in spec.md.
- `progress.md` is the runtime evidence log (4-phase Lifecycle Status table + Audit-Ready Signal).

### §C.2 Tier S minimal Section A-E justification

This SPEC qualifies for the Tier S minimal Section A-E variant per `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability:

- **Scope**: 5 files modified (4 mechanical mirror cp + 1 test registry add), ~+115 lines total, mechanical content overwrite — within Tier S envelope (≤300 LOC, ≤5 files, ≤1 milestone).
- **Risk**: minimal-to-moderate (4 of 5 files are mechanical content overwrite with zero behavior change; the 5th file `rule_template_mirror_test.go` adds 1 test entry — a single-line insertion into a slice literal that activates a new subtest case `TestRuleTemplateMirrorDrift/hooks-system.md`). No production code change; no API surface change; no behavior change in any runtime path.
- **Verification**: 4 PASS gates from `TestRuleTemplateMirrorDrift` (after registry add) plus `go vet` + `golangci-lint` zero-delta gates.

Per Section A-E variant precedent established by Sprint 2 P4 trio (IVB-001 `d3ed4727d`, SARM-001 `5e0dc6a9b`, TMC-001 `38a638d3c`), Tier S minimal retains the 4-artifact form (spec/plan/acceptance/progress) for traceability rather than the Tier S strict 2-artifact form. This is a documented pattern (L33 in MEMORY.md), not a deviation.

### §C.3 Mx Step C disposition

Per `.claude/rules/moai/workflow/mx-tag-protocol.md` §a:

- Mx Step C **SKIP** condition requires `.go` file count = 0 AND @MX tag count delta = 0
- This SPEC modifies `internal/template/rule_template_mirror_test.go` (1 `.go` file), so SKIP is NOT eligible
- Mx Step C disposition: **EVALUATE** — verify @MX tag count source vs mirror delta = 0 in the 4 `.md` files (no tag drift introduced) AND verify the 1 `.go` registry change does NOT introduce new high fan_in @MX:ANCHOR candidates
- progress.md §Mx-phase Audit-Ready Signal records `mx_disposition=EVALUATE` with rationale

## §D. Lint surface

- **spec.md MUST emit `✓ No findings`** when checked by `go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001/spec.md`. The Out-of-Scope section (§B.2 above) is canonical only in spec.md.
- **plan.md / acceptance.md / progress.md** MAY emit `MissingExclusions` ERROR when lint-checked individually. This is accepted derived-artifact lint surface per IVB-001/SARM-001/TMC-001 precedent (CI does not block on derived-artifact lint per `git-strategy.yaml` pattern). The canonical lint surface is `spec.md` only.
- Frontmatter canonical 12 fields enforced across all 4 artifacts per `.claude/rules/moai/development/spec-frontmatter-schema.md` SSOT. `tags:` MUST be CSV-string form (`tags: "a, b, c"`), not YAML array — array form causes `ParseFailure` in `internal/spec/lint.go` `SPECFrontmatter.Tags string yaml:"tags"` binding.

## §E. Sprint context

### §E.1 Sprint 7 entry position

| Phase | SPEC ID | Tier | Scope | Status |
|-------|---------|------|-------|--------|
| Sprint 2 P4.1 | SPEC-V3R6-I18N-VALIDATOR-BUDGET-001 | S minimal | i18n-validator budget 30s→35s | complete `d3ed4727d` |
| Sprint 2 P4.2 | SPEC-V3R6-SKILLS-AUDIT-RUN-MD-001 | S minimal | solo run.md test path | complete `5e0dc6a9b` |
| Sprint 2 P4.3 | SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001 | S minimal | spec-assembly.md mirror | complete `38a638d3c` |
| **Sprint 7 entry** | **SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001** | **S minimal** | **4-file mechanical mirror cleanup A1-A4** | **draft (this SPEC)** |

This SPEC continues the smallest-first ordering pattern by handling 4 mechanical mirrors that need no policy decision. After this SPEC completes 4-phase lifecycle (plan → run → sync → mx EVALUATE), 4 of 10 baseline failures are cleared (TestRuleTemplateMirrorDrift 3/3 + TestLateBranchTemplateMirror remains as before from TMC-001 closure).

### §E.2 Post-merge follow-up

Categories B1-F1 deferred to Sprint 8 (requires policy decisions). Sprint 7 then enters with the remaining backlog SPEC candidates per the operator's prioritization:
- CLI-INTEGRATION-001
- PROMPT-CACHE-001
- SPEC-ID-VALIDATION-001 (per L51 proposal)
- `.claude/` DRIFT 4 follow-up per user mid-plan audit Request #3

## §F. Requirements (EARS Format)

### REQ-TMD-001 (Ubiquitous, mandatory)

**WHEN** `/moai run SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001` executes M1, the mirror file `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` **shall** be byte-identical to the source `.claude/rules/moai/workflow/spec-workflow.md` after the fix commit (post-edit `diff <src> <mirror> | wc -l` = 0).

### REQ-TMD-002 (Ubiquitous, mandatory)

**WHEN** `/moai run SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001` executes M1, the mirror file `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` **shall** be byte-identical to the source `.claude/rules/moai/core/agent-common-protocol.md` after the fix commit.

### REQ-TMD-003 (Ubiquitous, mandatory)

**WHEN** `/moai run SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001` executes M1, the mirror file `internal/template/templates/.claude/agents/meta/plan-auditor.md` **shall** be byte-identical to the source `.claude/agents/meta/plan-auditor.md` after the fix commit.

### REQ-TMD-004 (Ubiquitous, mandatory)

**WHEN** `/moai run SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001` executes M1, the mirror file `internal/template/templates/.claude/rules/moai/core/hooks-system.md` **shall** be byte-identical to the source `.claude/rules/moai/core/hooks-system.md` after the fix commit (NEW mirror entry — currently NOT registered in `rule_template_mirror_test.go`).

### REQ-TMD-005 (Event-Driven, mandatory)

**WHEN** M1 executes, the test registry `internal/template/rule_template_mirror_test.go` `workflowOptMirroredPaths` slice **shall** be extended with a new entry `.claude/rules/moai/core/hooks-system.md` inserted within the `core/` group (between the existing `agent-common-protocol.md` entry at line 48 and the `spec-workflow.md` entry at line 50). This new entry activates the subtest case `TestRuleTemplateMirrorDrift/hooks-system.md` which validates REQ-TMD-004.

### REQ-TMD-006 (Unwanted Behavior, mandatory)

**While** manager-develop executes the fix, the 4 operational sources `.claude/rules/moai/workflow/spec-workflow.md`, `.claude/rules/moai/core/agent-common-protocol.md`, `.claude/agents/meta/plan-auditor.md`, `.claude/rules/moai/core/hooks-system.md` **shall not** be modified (no operational regression — sources remain byte-identical pre/post; `git diff HEAD..HEAD~1 -- <each source>` = 0 lines).

### REQ-TMD-007 (Unwanted Behavior, mandatory)

**WHILE** other dirty/untracked files exist in the working tree (PRESERVE list per L45 — 11 entries: `.claude/output-styles/moai/{einstein,moai}.md`, `.moai/config/sections/{git-convention,language,quality}.yaml`, `.moai/harness/usage-log.jsonl`, `.moai/harness/observations.yaml`, `.moai/research/v3.0-redesign-2026-05-23.md`, `internal/template/templates/.claude/output-styles/moai/{einstein,moai}.md`, `i18n-validator`), they **shall** remain unchanged across the fix commit.

### REQ-TMD-008 (State-Driven, mandatory)

**WHEN** `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift'` is invoked post-fix, ALL 4 subtests in the source-modified set (`spec-workflow.md`, `agent-common-protocol.md`, `plan-auditor.md`, `hooks-system.md`) **shall** PASS. The other pre-existing PASS subtests (`manager-develop-prompt-template.md`, `verification-batch-pattern.md`, `default.md`, `agent-teams-pattern.md`, `frontend.md`, `ci-watch-protocol.md`) **shall** remain PASS.

### REQ-TMD-009 (State-Driven, mandatory)

**WHEN** `go vet ./...` and `golangci-lint run --timeout=2m` are invoked post-fix, the baseline 0-issue state **shall** be preserved (no new vet warnings, no new lint findings introduced).

### REQ-TMD-010 (Event-Driven, L46 attribution discipline)

**WHEN** other template mirror tests are run post-fix (`TestLateBranchTemplateMirror` entries other than `spec-assembly.md` which was cleared by TMC-001, `TestBackwardCompatibility`, `TestAgentFrontmatterAudit`, `TestAllAgentsInCatalog`, `TestEmbeddedTemplates_AgentDefinitions`, `TestLoadCatalog`, `TestLoadEmbeddedCatalog_Success`, `TestRetirementCompletenessAssertion` × 2), the pre-existing baseline failures **shall** persist exactly as before — this SPEC fixes ONLY the 4 mirrors in §B.1 plus the test registry add for hooks-system.md; other drifts (categories B1-F1) are deferred to Sprint 8.

### REQ-TMD-011 (Optional, MAY)

The fix MAY use `git add -p` or path-specific `git add <mirror-path>` discipline to ensure ONLY the 5 modified files are staged (no accidental source or sibling staging). The minimal-change discipline supports either path-specific `git add` of just the 5 files (REQ-TMD-011 satisfied) or a wider `git add .` followed by un-staging unintended paths — the former is preferred for traceability.
