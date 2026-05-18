---
id: SPEC-V3R4-SPECLINT-DEBT-002
version: "0.2.0"
status: draft
created: 2026-05-16
updated: 2026-05-16
author: GOOS행님
priority: P1
tags: "v3r4, spec-lint, frontmatter, schema-canonical, plan-workflow, ssot, dual-schema-drift, foundation, compact"
issue_number: null
title: plan workflow Pre-Write Frontmatter Checklist를 lint.go canonical (12-field)에 정렬 — compact extract
phase: "v3.0.0 R4 — Foundation Cleanup"
module: ".claude/skills/moai/workflows/plan.md + .claude/skills/moai/team/plan.md + internal/spec/lint.go (조건부)"
dependencies: []
related_specs:
  - SPEC-V3R4-SPECLINT-DEBT-001
  - SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001
breaking: false
bc_id: []
lifecycle: spec-anchored
related_theme: "Foundation Cleanup — Schema SSOT 정립"
target_release: v2.20.0-rc1
---

# SPEC-V3R4-SPECLINT-DEBT-002 — Compact Extract

> **Note**: 본 파일은 spec.md + acceptance.md에서 EARS requirements와 acceptance scenarios만 추출한 auto-generated extract이다. 전체 컨텍스트는 spec.md, plan.md, acceptance.md 원본을 참조.

---

## EARS Requirements (5)

### REQ-SDBT-002-001 (Ubiquitous)

The `.claude/skills/moai/workflows/plan.md` § "Pre-Write Frontmatter Checklist" body **shall** list exactly the 12 required fields enforced by `internal/spec/lint.go` `FrontmatterSchemaRule.Check()` — `id`, `title`, `version`, `status`, `created`, `updated`, `author`, `priority`, `phase`, `module`, `lifecycle`, `tags` — in canonical order.

### REQ-SDBT-002-002 (Ubiquitous)

The plan.md "Rejected legacy aliases" section **shall** list `created_at:`, `updated_at:`, `labels:`, `spec_id:` as rejected (lint.go canonical 보호 방향). The previous direction (rejecting `created:`/`updated:`/`tags:`) **shall not** appear.

### REQ-SDBT-002-003 (Event-driven)

**When** manager-spec composes a new SPEC frontmatter using plan workflow body as guide, the agent **shall** include all 12 canonical required fields (`id`, `title`, `version`, `status`, `created`, `updated`, `author`, `priority`, `phase`, `module`, `lifecycle`, `tags`) and **shall not** add snake_case aliases (`created_at:`, `updated_at:`, `labels:`) as redundant duplicates.

### REQ-SDBT-002-004 (Event-driven)

**When** `.claude/skills/moai/workflows/plan.md` Pre-Write Frontmatter Checklist section is modified, `.claude/skills/moai/team/plan.md` **shall** have an equivalent Pre-Write Frontmatter Checklist section created (Case C: new section addition, not sync of existing). The section content MUST match solo plan.md verbatim with team-mode adaptations only where applicable.

### REQ-SDBT-002-005 (State-driven)

**While** the SSOT reference document `.claude/rules/moai/development/spec-frontmatter-schema.md` exists, the three schema enforcement layers — plan workflow body, `internal/spec/lint.go` `FrontmatterSchemaRule`, and plan-auditor schema check — **shall** all reference this document as the single source of truth via explicit cross-reference links.

---

## Acceptance Criteria (5)

### AC-SDBT-002-001 (manager-spec workflow 일관성) maps REQ-SDBT-002-001

**Given** `.claude/skills/moai/workflows/plan.md` § "Pre-Write Frontmatter Checklist"가 12-field canonical로 갱신되었다, AND 갱신된 plan workflow를 가이드로 manager-spec이 새 SPEC 작성을 시작한다.

**When** manager-spec이 plan.md 12-field 목록을 따라 frontmatter를 작성한다, AND 작성된 spec.md를 `moai spec lint --strict`로 검증한다.

**Then** 검증 결과 FrontmatterInvalid finding이 0건이어야 한다 (exit code 0).

---

### AC-SDBT-002-002 (snake_case rejection) maps REQ-SDBT-002-002

**Given** Wave 3 fixture `internal/spec/testdata/frontmatter-schema/invalid-snake-case-only/spec.md`가 snake_case alias만 보유 (`created_at:`, `updated_at:`, `labels:`)하고 canonical (`created:`, `updated:`, `tags:`)을 누락. 다른 9 required fields는 모두 정상값.

**When** fixture를 lint.go FrontmatterSchemaRule.Check()로 검증한다.

**Then** lint.go는 **정확히 3개 FrontmatterInvalid findings**를 보고해야 한다 — 각각 `created`, `updated`, `tags` 누락. unknown YAML keys는 struct decoder가 silently drop.

---

### AC-SDBT-002-003 (regression evidence — SDF-001 hotfix reproduction) maps REQ-SDBT-002-003

**Given** SDF-001 run-phase에서 hotfix commit `b2b7f32c7` (squash-merged; `git rev-parse b2b7f32c7`로 검증 가능)이 필요했다, AND 본 SPEC implementation이 완료되어 plan workflow가 12-field canonical로 정렬되었다.

**When** SDF-001과 동일한 시나리오를 재현한다 (manager-spec이 plan workflow를 충실히 따라 frontmatter 작성), AND 작성된 spec.md를 `moai spec lint --strict`로 검증한다.

**Then** hotfix commit이 더 이상 필요하지 않다 — manager-spec이 plan workflow 가이드만으로 12-field frontmatter를 작성하므로 lint pass. Binary check: `git log --since=2026-05-16 --grep "fix(spec).*frontmatter canonical" --oneline | wc -l` MUST = 0 (sync-phase baseline).

---

### AC-SDBT-002-004 (team/plan.md Pre-Write Checklist 신설 — Case C) maps REQ-SDBT-002-004

**Given** Plan-phase 실측: `.claude/skills/moai/team/plan.md` 존재 (8065 bytes, 2026-05-14) + Pre-Write Frontmatter Checklist 섹션 부재. team 모드도 SPEC 작성을 manager-spec에 위임.

**When** Wave 3 T-Wave3-001에서 team/plan.md의 delegation step 직후에 solo plan.md의 Pre-Write Frontmatter Checklist 섹션을 verbatim mirror하여 신설 (~75 LOC).

**Then** team/plan.md에 "Pre-Write Frontmatter Checklist" 섹션이 존재, 12 required fields + rejected aliases가 명시, solo plan.md와 의미적 동등.

---

### AC-SDBT-002-005 (SSOT cross-reference 3-layer) maps REQ-SDBT-002-005

**Given** Wave 2에서 `.claude/rules/moai/development/spec-frontmatter-schema.md` SSOT 문서가 신설되었다, AND 3개 enforcement layer가 모두 SSOT 문서를 cross-reference 한다.

**When** 코드베이스 grep으로 SSOT 파일명 발견 위치를 카운트한다.

**Then** `grep -rln "spec-frontmatter-schema.md"` 결과가 최소 3개 파일에서 매칭되어야 한다 (plan.md + lint.go + hub doc). SSOT 파일 자체도 존재.

---

## Out of Scope (Exclusions)

1. 197 기존 SPECs frontmatter 마이그레이션 — 51 dual-field `created_at:` + 53 `labels:`는 canonical과 dual-field 공존, lint regression 0건. opportunistic cleanup은 별도 SPEC 후보.
2. lint.go `FrontmatterSchemaRule` 12-field 의미적 변경 — canonical 유지가 본 SPEC의 전제
3. `title`/`phase`/`module`/`lifecycle` 의미적 검증 강화 — 별도 SPEC 후보
4. `issue_number` → lint.go required 추가 — optional 유지 권장
5. snake_case 표준화 — Python/Ruby 관례를 Go/YAML에 강제하지 않음
6. plan workflow body의 다른 section 변경 — Pre-Write Frontmatter Checklist에 한정
7. `dependencies`/`related_specs` 등 optional metadata 검증 강화 — 별도 rule로 이미 존재
8. SPEC frontmatter YAML parser 변경 (internal/spec/parse.go)
