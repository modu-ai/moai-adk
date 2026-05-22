---
id: SPEC-V3R6-SKILL-COMPRESS-001
title: "상위 5개 무거운 Workflow Skill Body 압축 (-17K tokens)"
version: "0.2.0"
status: implemented
created: 2026-05-22
updated: 2026-05-23
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/skills/moai-workflow-testing, .claude/skills/moai-workflow-spec, .claude/skills/moai-workflow-project, .claude/skills/moai-domain-design-handoff, .claude/skills/moai-meta-harness (+ template mirrors)"
lifecycle: spec-anchored
tags: "skills, compression, top-5, token-economy, progressive-disclosure, wave-1, v3.0.0"
tier: M
issue_number: null
depends_on: []
related_specs: [SPEC-V3R6-RULES-PATH-SCOPE-001, SPEC-V3R6-RULES-COMPRESS-001, SPEC-V3R6-SKILL-CONSOLIDATE-001]
---

# SPEC-V3R6-SKILL-COMPRESS-001: 상위 5개 무거운 Workflow Skill Body 압축

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-22 | manager-spec | Initial draft — Wave 1 SPEC #4 of 4. v3.0 환골탈체 §Layer 2 Top-5 Skill body compression (-17K tokens 절감). 5 skills (`moai-workflow-testing`, `moai-workflow-spec`, `moai-workflow-project`, `moai-domain-design-handoff`, `moai-meta-harness`) Level 2 body를 11,664w → 7,300w로 축소. Progressive Disclosure 3-Level 보존: Level 1 metadata (frontmatter) 불변 / Level 2 body 압축 / Level 3 references/ 디렉토리로 분리. Template-First Rule (CLAUDE.local.md §2) 동시 적용. Tier M (3 artifacts). |
| 0.2.0 | 2026-05-23 | manager-develop | Run-phase COMPLETE — DDD cycle (ANALYZE-PRESERVE-IMPROVE) executed. Aggregate compression 11,667w → 7,004w (-4,663w = -39.96%). Per-skill: testing 3153→1230w / spec 2394→1637w / project 2068→1389w / design-handoff 2042→1274w / meta-harness 2010→1474w. 16 new references/ files created (testing 5 + spec 3 + project 3 + design-handoff 2 + meta-harness 3). All 12 ACs PASS. Frontmatter byte-identity verified for all 5 skills (AC-SCM-011). Template mirror identical for all 5 SKILL.md + 16 references/ files (AC-SCM-008). Catalog hashes regenerated via `gen-catalog-hashes.go --all` (AC-SCM-010 TestAllSkillsInCatalog + TestManifestHashFormat PASS). Cross-platform builds PASS (linux + windows, AC-SCM-012). NEW lint regressions: 0. Naming convention Option A: new `references/` (plural), existing `reference/` (singular) preserved. Hybrid Trunk Tier M direct policy per CLAUDE.local.md §23. |

## 1. Goal

상위 5개 가장 무거운 workflow/domain/meta skill의 **Level 2 body**를 압축하여 v3.0 토큰 베이스라인을 **약 -17K tokens (≈ 37%)** 감축한다.

### 1.1 정량 목표

| Skill | 현재 (local 측정 baseline) | v3.0 목표 | 절감 |
|---|---:|---:|---:|
| `moai-workflow-testing` | 3,153w | 1,800w | -1,353w |
| `moai-workflow-spec` | 2,394w | 1,500w | -894w |
| `moai-workflow-project` | 2,068w | 1,200w | -868w |
| `moai-domain-design-handoff` | 2,039w | 1,400w | -639w |
| `moai-meta-harness` | 2,010w | 1,400w | -610w |
| **합계** | **11,664w** | **7,300w** | **-4,364w (≈ -17K tokens)** |

(token 추정: 영문 word 1 ≈ 1.3 token, 4,364w × ~3.9 = ~17K tokens. 실제 token 수는 tokenizer 의존이나 직선적 비례. §Layer 2 line 200-205의 v3.0 design 추정과 일치.)

### 1.2 비기능 목표

- Progressive Disclosure 3-Level **구조 보존**: Level 1 frontmatter 불변, Level 2 body만 축소, Level 3 references/ 분리.
- Trigger 키워드 **누락 0건**: 각 skill의 trigger keyword (`triggers.keywords`, `triggers.phases`, `triggers.agents`)는 frontmatter 불변이며, body에서도 keyword decision logic 보존.
- Template-First Rule (CLAUDE.local.md §2 HARD) 의무: local + `internal/template/templates/.claude/skills/<skill>/` byte-identical.

## 2. Context

### 2.1 v3.0 환골탈체 §Layer 2 인용

`.moai/research/v3.0-design-2026-05-22.md` line 200-205:

> §Layer 2 — 상위 5 skill 압축 (-17K tokens)
> - `moai-workflow-testing` (3,153w → 1,800w): pyramid 핵심만, references 분리
> - `moai-workflow-spec` (2,394w → 1,500w): EARS 핵심만, example 분리
> - `moai-workflow-project` (2,068w → 1,200w)
> - `moai-domain-design-handoff` (2,039w → 1,400w)
> - `moai-meta-harness` (2,010w → 1,400w)

### 2.2 §8 사용자 결정 반영

design.md §8 (4 결정 확정 2026-05-22):
- 6/15 이전 완료 → priority P1
- 토큰 감축 우선 → 본 SPEC가 Layer 2의 가장 큰 단일 감축 (-17K)

### 2.3 Progressive Disclosure 보존 의무

`.claude/rules/moai/development/skill-authoring.md` § Progressive Disclosure:

```
Level 1 (Metadata, ~100 tokens): frontmatter — 압축 영향 0
Level 2 (Body, ~5000 tokens): trigger 매칭 시 로드
Level 3 (Bundled, on-demand): references/, modules/ 등 별도 파일
```

본 SPEC는 Level 2를 ~5K target에 가깝게 축소하면서 Level 3로 분리되는 항목 (verbose references, examples, patterns)을 `<skill>/references/` 디렉토리로 외부화한다.

### 2.4 Template-First Rule (CLAUDE.local.md §2 HARD)

```
WHEN adding new files to .claude/skills/, THE template source MUST be updated FIRST.
make build embeds template files; local + template must remain byte-identical.
```

본 SPEC는 SKILL.md 수정 + 신규 Level 3 references/ 파일 추가 + template mirror 동시 적용 (5 SKILL.md × 2 + 5-15 Level 3 files × 2 = 20-40 file mirror pairs).

## 3. Stakeholders

| Role | Responsibility |
|---|---|
| manager-spec (this SPEC) | spec.md / plan.md / acceptance.md 작성 |
| manager-develop (next) | 5 skill body 축소 + Level 3 분리 + template mirror |
| Claude Code runtime | Level 1 frontmatter trigger 매칭 (불변), Level 2 body 로딩 |
| MoAI subagents (manager-develop, manager-quality 등) | trigger 매칭 후 skill body 활용 — 압축 후에도 핵심 decision logic 유지 의무 |
| Project users | `/moai run`, `/moai plan` 등 워크플로우 명령 사용자 — 절감된 토큰으로 더 큰 context 확보 |

## 4. EARS Requirements

### REQ-SCM-001 (Ubiquitous)
THE `moai-workflow-testing` skill Level 2 body word count SHALL be ≤ 2,000 words after compression (target 1,800w + 200w 여유).

### REQ-SCM-002 (Ubiquitous)
THE `moai-workflow-spec` skill Level 2 body word count SHALL be ≤ 1,700 words after compression (target 1,500w + 200w 여유).

### REQ-SCM-003 (Ubiquitous)
THE `moai-workflow-project` skill Level 2 body word count SHALL be ≤ 1,400 words after compression (target 1,200w + 200w 여유).

### REQ-SCM-004 (Ubiquitous)
THE `moai-domain-design-handoff` skill Level 2 body word count SHALL be ≤ 1,600 words after compression (target 1,400w + 200w 여유).

### REQ-SCM-005 (Ubiquitous)
THE `moai-meta-harness` skill Level 2 body word count SHALL be ≤ 1,600 words after compression (target 1,400w + 200w 여유).

### REQ-SCM-006 (Ubiquitous)
THE sum of 5 compressed skill Level 2 body word counts SHALL be ≤ 8,200 words (target 7,300w + 900w aggregate 여유).

### REQ-SCM-007 (Event-Driven)
WHEN any Level 2 body section is removed for compression, the removed content SHALL be relocated to a `<skill>/references/` Level 3 file OR explicitly justified as obsolete (with rationale in HISTORY / progress.md).

### REQ-SCM-008 (Ubiquitous, Trigger Preservation)
THE Level 1 frontmatter `triggers` block (keywords / phases / agents) of each of the 5 skills SHALL remain byte-identical between pre-compression and post-compression states.

### REQ-SCM-009 (Ubiquitous, Decision Logic Preservation)
For every `triggers.keywords` entry of each of the 5 skills, the Level 2 body SHALL retain at least one occurrence of either (a) the exact keyword, or (b) a synonym/decision-logic reference that allows the agent to recognize the trigger context.

### REQ-SCM-010 (Event-Driven, Template Mirror)
WHEN a SKILL.md file is modified or a Level 3 references file is created in `.claude/skills/<skill>/`, THE corresponding template mirror at `internal/template/templates/.claude/skills/<skill>/` SHALL be updated to byte-identical content in the same commit.

### REQ-SCM-011 (Event-Driven, Cross-reference Integrity)
WHEN a Level 3 references file is introduced, THE Level 2 body SHALL contain at least one cross-reference link to that file (relative path or markdown link), AND the link target SHALL resolve to an existing file.

### REQ-SCM-012 (State-Driven, Catalog Consistency)
IF the catalog registry (`internal/template/catalog.yaml`) includes any of the 5 affected skills, THEN the catalog hash entry MUST be regenerated (`go run gen-catalog-hashes.go --all`) such that `TestManifestHashFormat` and `TestAllSkillsInCatalog` continue to pass.

## 5. Exclusions

### 5.1 Out of Scope

- **다른 28+ skill 압축** (본 SPEC는 상위 5만; 그 외는 Wave 2/3 별도 SPEC 또는 deferred)
- **Skill 통폐합** (`SPEC-V3R6-SKILL-CONSOLIDATE-001` 담당 — agency/copywriter/designer/team-pattern-cookbook 등 흡수)
- **Skill 신규 생성** (본 SPEC는 압축만; 신규 skill 부재)
- **Level 1 metadata 수정**: 5 skill frontmatter 일체 불변 (REQ-SCM-008 강제). `name`, `description`, `allowed-tools`, `metadata.*`, `triggers.*` 모두 byte-identical
- **Foundation 분리 강화**: `moai-foundation-core` (~1,500w 목표), `moai-foundation-cc` (~1,000w 목표), `ref-agent-catalog` 신규 생성 등은 별도 SPEC (`SPEC-V3R6-FOUNDATION-SPLIT-001` 가칭, Wave 2 후보)
- **Rule 압축**: `.claude/rules/moai/` 압축은 `SPEC-V3R6-RULES-COMPRESS-001` 담당 (병렬 Wave 1 SPEC #2)
- **Rule path scope**: `SPEC-V3R6-RULES-PATH-SCOPE-001` 담당 (병렬 Wave 1 SPEC #1)
- **modules/ 디렉토리 정리**: testing skill의 `modules/automated-code-review`, `modules/ddd`, `modules/performance` 등 22개 sub-디렉토리 deduplication은 본 SPEC out of scope (압축 후 reference link만 정합성 보장; 실제 modules/ 콘텐츠 dedupe는 미래 SPEC)
- **CHANGELOG entry**: Tier M test/skill-only 변경이나 user-visible behavior change 동반 가능성 있음 — sync phase에서 결정 (현 단계 미정)

### 5.2 Non-goals

- Token tokenizer를 정확하게 사용한 token count 측정 (본 SPEC는 word count proxy 사용 — `wc -w` 기반)
- Skill 의미적 동등성 검증 자동화 (LLM 기반 semantic equivalence test — 향후 evaluator-active SPEC 후보)
- 5 skill의 functional regression test (Claude Code runtime trigger 매칭 통합 테스트는 본 SPEC에서 추가하지 않음 — 기존 unit test 보존)

## 6. Risks

### R-SCM-001 (Medium / High) — Trigger Decision Logic 누락
**Description**: Level 2 body 압축 과정에서 trigger keyword에 대응하는 decision logic 또는 핵심 워크플로우 절차가 함께 제거될 위험.
**Likelihood**: Medium (5 skill × 평균 8개 keyword = 40개 결정 경로 검토 필요).
**Impact**: High (agent가 trigger 매칭 후 skill body를 읽어도 행동 가이드 부재 → silent regression).
**Mitigation**:
- REQ-SCM-008 (frontmatter 불변) + REQ-SCM-009 (keyword 또는 synonym 보존) HARD 강제
- AC-SCM-007의 keyword presence grep 검증
- 압축 전후 diff에서 keyword/phrase loss 자동 탐지 스크립트 (manager-develop이 wave Pre-flight 도구로 활용)
- 압축은 한 번에 1 skill씩 (M2-M6 milestone 분리)으로 회귀 위험 격리

### R-SCM-002 (Low / Medium) — Level 3 Reference 분리 후 미등록
**Description**: 신규 `<skill>/references/*.md` 파일 생성 시 catalog 또는 frontmatter index에 등록 누락 → Claude Code가 Level 3를 발견하지 못해 dead reference.
**Likelihood**: Low (Level 3는 on-demand 로딩; 명시적 link만 있으면 작동).
**Impact**: Medium (사용자가 link click 시 404; 자동 로딩에는 영향 없음).
**Mitigation**:
- REQ-SCM-011 (cross-reference 무결성) AC-SCM-009로 검증 (`grep` 후 `test -f`)
- Level 3는 frontmatter 등록 불필요 — markdown link만 정합성 보장

### R-SCM-003 (Medium / High) — Template Mirror Drift
**Description**: Local skill만 압축하고 template mirror 누락 → `moai update` 시 사용자 프로젝트가 옛 무거운 body 받음 (Template-First Rule §2 위반).
**Likelihood**: Medium (반복 위반 사례 — V3R6-ABSORB-CLEANUP-001 / TEMPLATE-MIRROR-DRIFT 25-file audit 선례).
**Impact**: High (사용자 토큰 절감 효과 0; v3.0 환골탈체 가치 손실).
**Mitigation**:
- REQ-SCM-010 (template mirror byte-identical) HARD 강제
- AC-SCM-008 `diff -rq` 검증
- M2-M6 각 milestone 끝에 template mirror 동기화 의무
- TestRuleTemplateMirrorDrift 같은 자동 검증 (기존 CI guard 활용)

### R-SCM-004 (Low / Medium) — Cross-reference Dangling Links
**Description**: 압축 과정에서 Level 2 body가 외부 rule, skill, agent를 참조했던 링크가 깨지거나, 신규 Level 3 파일 path가 잘못 작성됨.
**Likelihood**: Low (markdown link broken은 grep으로 쉽게 탐지).
**Impact**: Medium (사용자가 dead link 클릭 시 frustration).
**Mitigation**:
- REQ-SCM-011 + AC-SCM-009
- 외부 참조 (`.claude/rules/...`, `.moai/...`) link 정합성 검증 별도 — 본 SPEC 본문에서 보존하는 reference는 변경 없음 (압축은 인용 단축이지 path 변경 아님)

### R-SCM-005 (Low / Low) — Sample Path 오인식
**Description**: design.md 추정 baseline (3,153w 등)이 본 SPEC 작성 직전 측정한 실제 baseline과 일치함을 사전 확인했으나, run-phase 시작 시점에 다른 변경 commit으로 인해 baseline 변동 가능성.
**Likelihood**: Low (Wave 0 머지 이후 5 skill 직접 수정 commit 없음 — git log 확인 의무).
**Impact**: Low (목표 word count는 절대값이므로 base가 변해도 final cap 유지하면 됨).
**Mitigation**:
- M1 baseline 재측정 step에서 사전 확인 의무
- target은 절대 cap (≤ 2,000w 등)이지 percentage가 아니므로 base shift 영향 없음

## 7. References

- [v3.0 환골탈체 design.md §Layer 2](.moai/research/v3.0-design-2026-05-22.md) line 200-205
- [skill-authoring.md § Progressive Disclosure](.claude/rules/moai/development/skill-authoring.md)
- [CLAUDE.local.md §2 Template-First Rule](CLAUDE.local.md)
- [SPEC-V3R6-HOOK-CONTRACT-FIX-001 — Wave 0 reference SPEC](.moai/specs/SPEC-V3R6-HOOK-CONTRACT-FIX-001/spec.md)
- [SPEC-V3R6-RULES-COMPRESS-001 — Wave 1 병렬 SPEC #2](.moai/specs/SPEC-V3R6-RULES-COMPRESS-001/spec.md) (parallel session)
- [SPEC-V3R6-SKILL-CONSOLIDATE-001 — Wave 1 병렬 SPEC #3](.moai/specs/SPEC-V3R6-SKILL-CONSOLIDATE-001/spec.md) (parallel session)
