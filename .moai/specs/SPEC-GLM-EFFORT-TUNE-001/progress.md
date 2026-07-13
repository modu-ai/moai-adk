---
id: SPEC-GLM-EFFORT-TUNE-001
title: "GLM effort overlay configuration tune-up (P1/P2/P4) — progress"
version: "0.1.0"
status: in-progress
created: 2026-07-14
updated: 2026-07-14
author: manager-spec
priority: P2
phase: "v3.x config-tune"
module: "internal/template/glm_effort_overlay.go + .moai/config/sections/llm.yaml"
lifecycle: spec-anchored
tags: "glm, effort, overlay, config, template-mirror, reasoning-effort"
related_specs: [SPEC-MODEL-TIER-PLANTYPE-001]
---

# SPEC-GLM-EFFORT-TUNE-001 — progress.md

## §E.1 Plan-phase Audit-Ready Signal

| Field | Value |
|-------|-------|
| SPEC ID regex self-check | `SPEC-GLM-EFFORT-TUNE-001` → PASS (decomposition: SPEC ✓ · GLM ✓ · EFFORT ✓ · TUNE ✓ · 001 ✓) |
| Frontmatter 12-field schema | OK (id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags) |
| Out of Scope section | Present — 5 `### Out of Scope — <topic>` H3 sub-headings with `-` bullets (P3, tierProfiles, frontmatter, override tiers, per-spawn channel) |
| GEARS notation | 12 REQs (Ubiquitous / When / While / Where / event-detected); zero `IF/THEN` in NEW requirements |
| Line-number citation discipline | spec.md §B cites real lines read this session (glm_effort_overlay.go:26-33, 88-100, 102-110; model_policy.go:140-154, 294-323; glm_effort_overlay_test.go:114-135; agent-authoring.md:334-374) |
| Tier classification | M (standard) — cross-file Go + Go test + template mirror + config + comments |
| Artifact set | 6 files: spec.md, plan.md, acceptance.md, progress.md, design.md, research.md |
| plan-phase audit readiness | _<pending plan-auditor>_ |

## §E.2 Run-phase Evidence

### AC PASS/FAIL matrix

| AC | Sev | Status | Actual Output |
|----|-----|--------|---------------|
| AC-GET-001 (override set == {manager-develop}) | MUST | PASS | `grep -n '"builder-harness": true' glm_effort_overlay.go` → exit 1 (no match); `grep -n '"manager-develop": true'` → exactly 1 match at line 110 |
| AC-GET-002 (manager-develop → reasoning-max) | MUST | PASS | Go test `TestResolveGLMReasoning_CodingMaxOverride/manager-develop_input=low` PASS (override wins over thinking-off collapse) |
| AC-GET-003 (builder-harness → reasoning-high, make-or-break) | MUST | PASS | Go test `TestResolveGLMReasoning_CodingMaxOverride/builder-harness_input=high` PASS (`got.Name == "reasoning-high"`, NOT reasoning-max) |
| AC-GET-004 (test rename + cardinality) | MUST | PASS | `grep 'TestGLMCodingMaxOverrideAgents_ExactlyTwo'` → exit 1 (gone); `TestGLMCodingMaxOverrideAgents_ExactlyOne` present with `want := []string{"manager-develop"}` + `want exactly 1` |
| AC-GET-005 (doc comments updated) | MUST | PASS | `grep -n 'two code-producing\|{manager-develop, builder-harness}' glm_effort_overlay.go` → exit 1 (no match) |
| AC-GET-006 (full package test) | SHOULD | PASS | `go test ./internal/template/ -count=1` → `ok github.com/modu-ai/moai-adk/internal/template 1.143s` |
| AC-GET-007 (exposure block in BOTH llm.yaml) | MUST | _<pending M2>_ | — |
| AC-GET-008 (3-state vocabulary present) | MUST | _<pending M2>_ | — |
| AC-GET-009 (overlay = SSOT, no parallel path) | MUST | _<pending M2>_ | — |
| AC-GET-010 (honesty caveat present, no overclaim) | MUST | _<pending M2>_ | — |
| AC-GET-011 (mirror parity test passes / llm.yaml not covered) | SHOULD | _<pending M2>_ | KI-4 resolved: `rule_template_mirror_test.go` does NOT cover `.moai/config/sections/llm.yaml` (explicit "Out of scope" per CLAUDE.local.md §22); AC is INFO — both files still edited for consistency |
| AC-GET-012 (no config-loader CI-guard regression) | MUST | _<pending M2>_ | — |
| AC-GET-013 (overlay doc comments frame 3 states, not 2-tier) | MUST | _<pending M3>_ | — |
| AC-GET-014 (3-state in llm.yaml exposure) | MUST | _<pending M2/M3>_ | — |
| AC-GET-015 (docs-site grep evidence recorded) | MUST | _<pending M4>_ | — |
| AC-GET-016 (full repo test suite) | MUST | _<pending M5>_ | — |
| AC-GET-017 (lint + vet clean) | MUST | _<pending M5>_ | — |
| AC-GET-018 (make build succeeds) | MUST | _<pending M5 (post-M2)>_ | — |

### M1 commit (P1 override set change)

- Files: `internal/template/glm_effort_overlay.go` (map key removal + 3 doc-comment updates), `internal/template/glm_effort_overlay_test.go` (test rename + cardinality flip + 6 edge-case rows + AC-GET-003 behavioral assertion).
- TDD: RED confirmed (3 failing assertions — builder-harness still overridden + set cardinality 2≠1), GREEN confirmed (all PASS post-edit).
- Edge cases covered (acceptance.md §D.7): manager-develop low/max→max; builder-harness low→thinking-off, high→reasoning-high (AC-GET-003), xhigh→reasoning-max; manager-git low→thinking-off; manager-spec high→reasoning-high; super-advisor xhigh→reasoning-max.

## §E.3 Run-phase Audit-Ready Signal

_<pending M5>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
