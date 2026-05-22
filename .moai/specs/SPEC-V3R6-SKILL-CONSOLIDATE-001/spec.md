---
id: SPEC-V3R6-SKILL-CONSOLIDATE-001
title: "Skill 통폐합 5→2 — ci-loop / design / harness-patterns (-9.3K tokens)"
version: "0.1.0"
status: draft
created: 2026-05-22
updated: 2026-05-22
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/skills, internal/template/templates/.claude/skills, internal/template/catalog.yaml"
lifecycle: spec-anchored
tags: "skills, consolidation, token-economy, wave-1, v3.0.0, template-first"
tier: M
issue_number: null
depends_on: []
related_specs: [SPEC-V3R6-SKILL-COMPRESS-001, SPEC-V3R6-RULES-PATH-SCOPE-001, SPEC-V3R6-RULES-COMPRESS-001]
---

# SPEC-V3R6-SKILL-CONSOLIDATE-001: Skill 통폐합 5→2

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-22 | manager-spec | Initial draft — Wave 1 SPEC #3 (Layer 2 token-baseline 감축). Consolidate 7 source SKILL.md files into 3 unified skills: `moai-workflow-ci-loop` (ci-watch + ci-autofix), `moai-workflow-design` (design-import + design-context), `moai-harness-patterns` (harness-hook-ci + harness-workflow + harness-quality). Target: -9.3K tokens (~45% reduction from 6,914 → 3,800 words). Tier M (3 artifacts), ~20-25 cross-reference file impacts, mandatory Template-First Rule (local + internal/template/templates/ mirror), catalog.yaml 7-entry removal + 3-entry addition + 32-entry hash regen. |

## 1. Goal

v3.0 환골탈태 (`.moai/research/v3.0-design-2026-05-22.md` §Layer 2 line 185-211)의 핵심 목표 중 하나인 **Skill token-baseline 감축 -9.3K (~41%)** 를 달성한다. 현재 `.claude/skills/` 하위에 분산된 7개 source SKILL.md (6,914 words 합계) 를 3개 통합 SKILL.md (~3,800 words 합계) 로 재구성하여:

1. **인지 부하 절감**: 동일 도메인 (CI workflow / design workflow / harness patterns) 내 중복 정보 단일화
2. **Progressive Disclosure 효율 향상**: Level 2 트리거 매칭 시 fragmentation 제거 (3 skills → 1 skill body load)
3. **Cross-reference 일관성**: agents / rules / `moai/SKILL.md` 등 ~20 cross-ref 위치를 새 이름으로 정렬
4. **Template-First mirror 정합성**: local `.claude/skills/` + `internal/template/templates/.claude/skills/` 동시 적용

[ZONE:Evolvable] [HARD] 본 SPEC 의 *통합 메커니즘 + 절감 검증* 범위만 포함한다. 통합 후 body 의 상세 내용 작성 (workflow 단계 / examples / Red Flags / Verification 등 컨텐츠) 자체는 run-phase 작업이며 본 SPEC 의 AC는 통합 발생 여부 + token 감축량 + cross-ref 정합성 + template mirror 동등성 만 binary 로 검증한다.

## 2. Why

### 2.1 v3.0 Design 보고서 §Layer 2 직접 인용

`.moai/research/v3.0-design-2026-05-22.md` line 185-193:

> ### Layer 2: Skill 통폐합 + Body 압축 (152K → 90K, 41% off)
>
> **통폐합 (5 → 2)**:
>
> | 통합 전 (5 skills) | 통합 후 (2 skills) | 절감 |
> |---|---|---|
> | `moai-workflow-ci-watch` (961w) + `moai-workflow-ci-autofix` (1,507w) | `moai-workflow-ci-loop` (1,200w) | -3.8K |
> | `moai-workflow-design-import` (1,358w) + `moai-workflow-design-context` (1,049w) | `moai-workflow-design` (1,400w) | -3.0K |
> | `moai-harness-hook-ci` (655w) + `moai-harness-workflow` (666w) + `moai-harness-quality` (718w) | `moai-harness-patterns` (1,200w) | -2.5K |
> | **합계** | — | **-9.3K** |

> **Note**: 보고서 표현 "5 → 2"는 *카운트 절감 폭*을 의미하며 (`5+2+3 = 10 ≠ 5`), 본 SPEC 는 **7 source → 3 unified** 으로 정확히 명시한다 (소스 + 신규 skill 총 카운트는 `7-3 = 4` 감소).

### 2.2 사용자 §8 결정 (2026-05-22 AskUserQuestion 결과)

`v3.0-design-2026-05-22.md` §8 사용자 4 결정:

1. **6/15 이전 완료** — 본 SPEC P1 priority
2. **토큰 감축 우선** — Layer 2 가 가장 큰 단일 감축 레버 (-9.3K)
3. **Default profile** — `harness: standard` baseline 유지 (`thorough` 미강제)
4. **GLM 보수** — `claude` leader + selective glm 위임 (런타임 의사결정, 본 SPEC scope 외)

### 2.3 Baseline 측정 (2026-05-22)

`wc -w` 결과 (Pre-flight Check §G, step 3):

| Source skill | Words | 통합 대상 |
|---|---:|---|
| `moai-workflow-ci-watch/SKILL.md` | 961 | → `ci-loop` |
| `moai-workflow-ci-autofix/SKILL.md` | 1,507 | → `ci-loop` |
| `moai-workflow-design-import/SKILL.md` | 1,358 | → `design` |
| `moai-workflow-design-context/SKILL.md` | 1,049 | → `design` |
| `moai-harness-hook-ci/SKILL.md` | 655 | → `harness-patterns` |
| `moai-harness-workflow/SKILL.md` | 666 | → `harness-patterns` |
| `moai-harness-quality/SKILL.md` | 718 | → `harness-patterns` |
| **합계** | **6,914** | (7 source) |

Target post-consolidation: ≤ 3,800 words 합계 (ci-loop ≤1,500 / design ≤1,500 / harness-patterns ≤1,500, 각 300w 여유 포함).

### 2.4 Cross-reference 영향 범위 (Pre-flight Check §G, step 5)

7 source skill name 을 참조하는 위치 (stale worktrees 제외):

**Local files (`.claude/`)**:

| 카테고리 | 파일 (예시) | 영향 수 |
|---|---|---:|
| agents/harness/*-specialist.md | hook-ci-specialist / workflow-specialist / quality-specialist | 3 |
| rules/moai/workflow/*-protocol.md | ci-watch-protocol / ci-autofix-protocol | 2 |
| rules/moai/design/constitution.md | design skill 참조 | 1 |
| rules/moai/core/zone-registry.md | skill registry | 1 |
| skills/moai/SKILL.md | top-level skill orchestration | 1 |
| skills/moai/workflows/{fix,brain,design,sync/delivery}.md | workflow skill bodies | 4 |
| skills/moai-domain-{brand-design,design-handoff}/SKILL.md | sibling skill cross-ref | 2 |
| skills/moai-workflow-gan-loop/SKILL.md | sibling cross-ref | 1 |

**Template files (`internal/template/templates/`)**: 위 same 카운트 mirror (~15 files)

**Go test files**:

- `internal/design/dtcg/frozen_guard_test.go` (1 file) — design skill name reference 가능성 검증 필요

**Catalog**:

- `internal/template/catalog.yaml` 17 references (7 entries + name/path lines)

**Total cross-ref impact estimate**: ~30-35 file 위치 (local + template + Go test + catalog).

### 2.5 Cross-SPEC tension audit

| SPEC | 충돌 가능성 | 해결 |
|------|---|---|
| `SPEC-V3R6-HARNESS-RENAME-001` (MERGED) | 무관 — harness specialist 파일 rename 만, skill body 미수정 | ✅ 안전 |
| `SPEC-V3R6-AGENT-FOLDER-SPLIT-001` (run-COMPLETE) | 무관 — agent folder 이동, skill 미수정 | ✅ 안전 |
| `SPEC-V3R6-HOOK-CONTRACT-FIX-001` (Wave 0, plan-COMPLETE) | 무관 — `internal/hook/` scope | ✅ 안전 |
| `SPEC-V3R6-DOCS-USER-DRIFT-001` (Wave 0, plan-COMPLETE) | 무관 — docs-site sync workflow scope | ✅ 안전 |
| `SPEC-V3R6-SKILL-COMPRESS-001` (Wave 1 동시 계획) | **잠재 충돌** — 상위 5 무거운 skill 압축 vs 본 SPEC 7 통합. 통합 대상 7 skill 은 SKILL-COMPRESS scope 에 미포함 (testing/spec/project/design-handoff/meta-harness 가 압축 대상) | ⚠️ Scope 분리 확정 |
| `SPEC-V3R6-RULES-PATH-SCOPE-001` / `RULES-COMPRESS-001` (Wave 1) | 무관 — `.claude/rules/` scope (skill 비대상) | ✅ 안전 |
| `SPEC-V3R6-CATALOG-SSOT-001` (run-COMPLETE) | **연관** — catalog.yaml hash regen 자동화 이미 정착 (Makefile `gen-catalog-hashes.go --all`). 본 SPEC 의 catalog edit 도 자동 hash regen 활용 | ✅ 호환 |

## 3. EARS Requirements

### REQ-SC-001 (Ubiquitous, Source skill removal)

[ZONE:Frozen] [HARD] `/moai run` 완료 시점에 다음 7개 skill 디렉토리 (local `.claude/skills/` + template `internal/template/templates/.claude/skills/` 양쪽 모두) 가 존재하지 않아야 한다: `moai-workflow-ci-watch`, `moai-workflow-ci-autofix`, `moai-workflow-design-import`, `moai-workflow-design-context`, `moai-harness-hook-ci`, `moai-harness-workflow`, `moai-harness-quality`.

### REQ-SC-002 (Ubiquitous, Unified skill creation)

[ZONE:Frozen] [HARD] `/moai run` 완료 시점에 다음 3개 skill 디렉토리가 새로 존재해야 한다: `moai-workflow-ci-loop/SKILL.md`, `moai-workflow-design/SKILL.md`, `moai-harness-patterns/SKILL.md`. 각 SKILL.md 는 valid YAML frontmatter (skill name, description, paths 등 필수 필드) 를 포함해야 한다.

### REQ-SC-003 (Event-driven, Word-count budget)

WHEN 신규 통합 SKILL.md 의 `wc -w` 측정 시, 각 파일은 다음 상한을 준수해야 한다:

- `moai-workflow-ci-loop/SKILL.md` ≤ 1,500 words (목표 1,200, 여유 300)
- `moai-workflow-design/SKILL.md` ≤ 1,500 words (목표 1,400, 여유 100)
- `moai-harness-patterns/SKILL.md` ≤ 1,500 words (목표 1,200, 여유 300)
- **합계** ≤ 3,800 words (baseline 6,914 → target 3,800, **-3,114 words 절감 = -45%**)

토큰 추정: 약 -9.3K tokens (1 token ≈ 0.75 word 기준).

### REQ-SC-004 (Ubiquitous, Template-First mirror parity)

[ZONE:Frozen] [HARD] (CLAUDE.local.md §2 의무) `/moai run` 완료 시점에 다음 3 pair 의 SKILL.md 파일이 byte-identical 해야 한다:

- `.claude/skills/moai-workflow-ci-loop/SKILL.md` ↔ `internal/template/templates/.claude/skills/moai-workflow-ci-loop/SKILL.md`
- `.claude/skills/moai-workflow-design/SKILL.md` ↔ `internal/template/templates/.claude/skills/moai-workflow-design/SKILL.md`
- `.claude/skills/moai-harness-patterns/SKILL.md` ↔ `internal/template/templates/.claude/skills/moai-harness-patterns/SKILL.md`

검증: `diff -q <local> <template>` exit code 0.

### REQ-SC-005 (Ubiquitous, Catalog SSOT consistency)

[ZONE:Frozen] [HARD] `internal/template/catalog.yaml` 은 다음 조건을 모두 만족해야 한다:

1. 7개 old skill entry (`moai-workflow-ci-watch`, `moai-workflow-ci-autofix`, `moai-workflow-design-import`, `moai-workflow-design-context`, `moai-harness-hook-ci`, `moai-harness-workflow`, `moai-harness-quality`) 가 모두 제거됨
2. 3개 new skill entry (`moai-workflow-ci-loop`, `moai-workflow-design`, `moai-harness-patterns`) 가 추가됨 (path: `templates/.claude/skills/<name>/` 형식)
3. 전체 entry 의 hash 가 `Makefile` 의 `gen-catalog-hashes.go --all` 호출로 일관 재생성됨 (SPEC-V3R6-CATALOG-SSOT-001 인프라 활용)

### REQ-SC-006 (Event-driven, Cross-reference rename completeness)

WHEN `grep -rn "moai-workflow-ci-watch\|moai-workflow-ci-autofix\|moai-workflow-design-import\|moai-workflow-design-context\|moai-harness-hook-ci\|moai-harness-workflow\|moai-harness-quality" .claude/ CLAUDE.md internal/template/templates/.claude/ internal/design/dtcg/` 실행 시, **0 matches** 가 반환되어야 한다 (stale worktrees `.claude/worktrees/` 제외).

예외: 본 SPEC 자체 (`.moai/specs/SPEC-V3R6-SKILL-CONSOLIDATE-001/`), HISTORY 항목, memory 파일은 검색 범위 외.

### REQ-SC-007 (State-driven, Test-suite regression preservation)

WHILE 통합이 진행 중일 때, 다음 명령은 통합 전과 동일한 PASS/FAIL 상태를 유지해야 한다:

- `go test ./internal/template/... -run "TestAllSkillsInCatalog|TestAllAgentsInCatalog|TestManifestHashFormat"` — 통합 후 catalog.yaml 정합성 통과
- `go test ./internal/design/dtcg/... -run "TestFrozenGuard"` — design skill 참조 시 새 이름 정렬

NEW regression 0건 (pre-existing baseline failure 는 별도 SPEC 후속 처리, 본 SPEC scope 외).

### REQ-SC-008 (Ubiquitous, Skill body 최소 contract)

[ZONE:Evolvable] [HARD] 각 통합 SKILL.md 는 다음 섹션을 포함해야 한다 (Progressive Disclosure 표준):

1. YAML frontmatter (`name`, `description`, optional `paths`, optional `allowed-tools`)
2. `## Quick Reference` 또는 `# <Skill Name>` H1 (skill metadata 첫 페이지)
3. `## Implementation Guide` 또는 동등한 H2 (workflow / patterns 본문)
4. `## Works Well With` (related skills/agents/commands 명시) — 통합으로 인한 cross-ref 누락 방지
5. Source skill 흡수 출처 표기 — 본문 내 "absorbed from <old-skill-name>" 명시 또는 footer note (감사 추적성)

본 REQ 는 *구조 contract* 만 강제. 내용 품질 평가 (writing craft, examples 풍부도 등) 는 evaluator-active scope 이며 본 SPEC AC 외.

### REQ-SC-009 (Optional, Skill trigger keyword 비충돌)

WHERE 통합 skill 의 frontmatter `paths` 또는 trigger keyword 가 다른 sibling skill 과 의미적 충돌이 발생할 가능성이 있는 경우, 통합 skill 의 frontmatter description 에 충돌 회피 의도를 1줄 명시한다 (예: "Use this skill for CI loop workflow — NOT for general loop iteration patterns, see moai-workflow-loop").

검증: manual review (binary AC 외 — best-effort guidance).

## 4. Out of Scope

### 4.1 Out of Scope

본 SPEC 가 *명시적으로 제외* 하는 항목:

- **상위 5개 무거운 skill 압축**: `moai-workflow-testing` (≥5K words), `moai-workflow-spec`, `moai-workflow-project`, `moai-domain-design-handoff`, `moai-meta-harness` 의 body 압축은 **`SPEC-V3R6-SKILL-COMPRESS-001`** (Wave 1 sibling SPEC) 담당. 본 SPEC 는 통합 (5→2) 메커니즘만 다룸.
- **Skill body 내용 작성 자체**: 새 SKILL.md 의 *실제 워크플로우 단계 / examples / Red Flags / Verification* 내용 정의는 run-phase 작업이며 본 SPEC 의 AC 는 binary contract (existence + word count + cross-ref + template parity) 만 검증.
- **Foundation skill 분리**: `moai-foundation-core` (`moai-foundation-cc` 와 분리) 는 별도 SPEC (`SPEC-V3R6-FOUNDATION-SPLIT-001` 후보, Wave 미정).
- **Reference skill 정리**: `moai-ref-{api-patterns,git-workflow,owasp-checklist,react-patterns,testing-pyramid}` 5개 reference skill 은 본 SPEC 미대상. 향후 SPEC.
- **Domain skill 통폐합**: `moai-domain-{backend,frontend,database,brand-design,copywriting,design-handoff,ideation,research}` 8개 domain skill 은 본 SPEC 미대상. v3.1 또는 별도 wave.
- **Skill discovery 메커니즘 변경**: skill 자동 로딩 / 트리거 알고리즘 / Progressive Disclosure tier 정책은 본 SPEC scope 외.
- **`moai/SKILL.md` (top-level orchestrator skill) 압축**: cross-ref 갱신만 본 SPEC 에서 수행 (REQ-SC-006), body 압축은 SKILL-COMPRESS-001.
- **사용자 문서 (docs-site) 갱신**: skill 통합으로 인한 user-facing 문서 (`docs-site/content/{ko,en,ja,zh}/`) 변경은 `/moai sync` 단계 또는 별도 SPEC. 본 SPEC 는 `.claude/` + `internal/template/templates/.claude/` 만 다룸.
- **agent-memory / personal config 변경**: `.claude/agent-memory/` 내 stale 기록은 자동 제거 대상 아님 (정보 손실 위험).

### 4.2 Non-Goals (명시적 비목표)

- **Backward compatibility shim 제공 안 함**: 7 old skill name 으로 호출되던 코드/agent 가 *graceful fallback* 받지 않는다. Hard rename — 모든 cross-ref 가 새 이름으로 정렬되어야 한다 (REQ-SC-006). 이는 의도된 breaking change (v3.0.0 phase).
- **Skill version 번호 자체 변경 안 함**: 통합 skill 의 internal version 은 `0.1.0` 새로 시작. Source skill 의 version history 보존 의무 없음 (HISTORY 섹션에 absorption 출처만 1줄 명시).
- **Token 추정 정확도 보증 안 함**: -9.3K tokens 는 word count × 0.75 추정. 실제 Claude Code tokenizer 결과는 ±10% 변동 가능. 본 SPEC AC 는 word count (REQ-SC-003) 로 검증; token 측정 별도 도구 필요 시 future SPEC.

## 5. Risks

### R-SC-001: Template-First Mirror Drift (Severity: High, Likelihood: Medium)

**위험**: Local `.claude/skills/` 만 수정하고 `internal/template/templates/.claude/skills/` mirror 누락 → `TestRuleTemplateMirrorDrift` / `TestLateBranchTemplateMirror` CI BLOCKING 실패.

**Mitigation**:

- run-phase Section A 위임 prompt 에 Template-First Rule (CLAUDE.local.md §2) 명시 의무화
- 각 새 SKILL.md 생성 시 *반드시* MultiEdit 또는 동시 Write 로 local + template 동시 처리
- AC-SC-006 (diff -q) 가 BLOCKING gate 역할

**Reference**: V3R6 lesson #25 (`project_v3r6_template_mirror_drift_audit_2026_05_22` memory).

### R-SC-002: Cross-reference Rename 누락 (Severity: High, Likelihood: High)

**위험**: ~30 cross-ref 위치 중 일부 누락 → run-phase 종료 후에도 `grep` 매치 발견. 가장 흔한 누락 위치:

- `internal/design/dtcg/frozen_guard_test.go` (Go test, 종종 검색에서 제외됨)
- `agent-memory/manager-tdd/project_ciaut_wave2_complete.md` (사용자 의도와 무관한 잔여 ref 가능)
- agent frontmatter 내 `skills:` YAML array 의 entry

**Mitigation**:

- AC-SC-005 (grep 0 matches) 가 BLOCKING gate
- 검색 범위 명시: `.claude/`, `CLAUDE.md`, `internal/template/templates/.claude/`, `internal/design/dtcg/`
- agent-memory + `.claude/worktrees/` 는 명시적 제외 (stale, 정보 손실 회피)

### R-SC-003: Catalog Hash Drift (Severity: Medium, Likelihood: Medium)

**위험**: catalog.yaml 7 entry 제거 + 3 entry 추가 후 `TestManifestHashFormat` BLOCKING 실패. SPEC-V3R6-CATALOG-SSOT-001 의 Makefile `gen-catalog-hashes.go --all` 자동 hash regen 인프라가 정착했지만, *수동* edit 으로 hash 가 stale 해질 가능성.

**Mitigation**:

- `Makefile` `build` recipe 가 `gen-catalog-hashes.go --all` 을 prepend (SPEC-V3R6-CATALOG-SSOT-001 결과)
- run-phase 마지막에 `make build` 명시 실행 후 `go test ./internal/template/... -run TestManifestHashFormat` 검증 의무
- AC-SC-007 (test suite regression preservation) 의 일부

### R-SC-004: Body 부족으로 Workflow 정보 손실 (Severity: Medium, Likelihood: Medium)

**위험**: 통합 skill body 가 1,200 words 목표를 달성하려고 source 정보를 과도 압축 → 사용자/agent 가 필요한 workflow step 을 skill 본문에서 발견 못함.

**Mitigation**:

- REQ-SC-008 (최소 contract: Quick Reference + Implementation Guide + Works Well With) 가 구조 보장
- absorbed-from footer 가 감사 추적성 제공 (사용자가 원본 source 검색 가능)
- REQ-SC-003 의 상한 (1,500 words ≤) 은 *목표* 1,200 보다 여유, 정보 보존 우선
- Run-phase 작성자 (manager-develop 또는 orchestrator-direct) 의 craft 책임 — evaluator-active 가 후속 평가

**Note**: 본 risk 의 *완전* 방어는 본 SPEC scope 외 (binary AC 로 검증 불가). 사후 발견 시 separate SPEC 으로 body 보강.

### R-SC-005: Skill Trigger Keyword 충돌 (Severity: Low, Likelihood: Low)

**위험**: 새 통합 skill 의 `paths:` glob 또는 description keyword 가 기존 sibling skill (예: `moai-workflow-loop` vs `moai-workflow-ci-loop`) 과 의미적 충돌 → Claude Code skill auto-loader 가 잘못된 skill 매칭.

**Mitigation**:

- REQ-SC-009 (frontmatter description 에 의도 명시 권장)
- 새 skill name 자체가 충분히 specific (`ci-loop` vs `loop`, `design` 단어는 broad 하나 `moai-workflow-design` 풀네임 unique)
- manual review during run-phase

## Appendix A: 통합 매트릭스 (요약)

| 통합 후 Skill | Source Skills (Word Count) | Target Words | 절감 |
|---|---|---:|---:|
| `moai-workflow-ci-loop` | `ci-watch` (961) + `ci-autofix` (1,507) = 2,468 | ≤ 1,500 | **-968** |
| `moai-workflow-design` | `design-import` (1,358) + `design-context` (1,049) = 2,407 | ≤ 1,500 | **-907** |
| `moai-harness-patterns` | `hook-ci` (655) + `workflow` (666) + `quality` (718) = 2,039 | ≤ 1,500 | **-539** |
| **합계** | **7 source = 6,914 words** | **≤ 3,800 (~45% off)** | **-3,114 words ≈ -9.3K tokens** |

## Appendix B: Cross-Reference Update 표 (예시)

| 파일 (local + template 양쪽) | Old 참조 | New 참조 |
|---|---|---|
| `.claude/agents/harness/hook-ci-specialist.md` | `moai-harness-hook-ci` | `moai-harness-patterns` |
| `.claude/agents/harness/workflow-specialist.md` | `moai-harness-workflow` | `moai-harness-patterns` |
| `.claude/agents/harness/quality-specialist.md` | `moai-harness-quality` | `moai-harness-patterns` |
| `.claude/rules/moai/workflow/ci-watch-protocol.md` | `moai-workflow-ci-watch` | `moai-workflow-ci-loop` |
| `.claude/rules/moai/workflow/ci-autofix-protocol.md` | `moai-workflow-ci-autofix` | `moai-workflow-ci-loop` |
| `.claude/skills/moai/SKILL.md` | 7 names | 3 names |
| `.claude/skills/moai/workflows/fix.md` | `moai-workflow-ci-watch`, `moai-workflow-ci-autofix` | `moai-workflow-ci-loop` |
| `.claude/skills/moai/workflows/design.md` | `moai-workflow-design-import`, `moai-workflow-design-context` | `moai-workflow-design` |
| `.claude/skills/moai/workflows/brain.md` | (skill cross-ref) | new names |
| `.claude/skills/moai/workflows/sync/delivery.md` | (skill cross-ref) | new names |
| `.claude/skills/moai-domain-brand-design/SKILL.md` | design skill refs | `moai-workflow-design` |
| `.claude/skills/moai-domain-design-handoff/SKILL.md` | design skill refs | `moai-workflow-design` |
| `.claude/skills/moai-workflow-gan-loop/SKILL.md` | design skill refs | `moai-workflow-design` |
| `.claude/rules/moai/design/constitution.md` | `moai-workflow-design-import` | `moai-workflow-design` |
| `.claude/rules/moai/core/zone-registry.md` | (skill registry rows) | new names |
| `internal/template/catalog.yaml` | 7 entries (17 ref lines) | 3 entries |
| `internal/design/dtcg/frozen_guard_test.go` | (design skill name in frozen-zone test) | new name |
