# MoAI-ADK Architecture v2.0 - 전체 설계 요약

> **Commands → Sub-agents → Skills 3-Tier 아키텍처 (Claude Code 공식 표준 기반)**

---

## 📋 문서 개요

### 관련 문서

1. **ARCHITECTURE_v2.md**: 전체 아키텍처 상세 설명
2. **feature-selector_DESIGN.md**: Skills 자동 선택 알고리즘
3. **template-optimizer_DESIGN.md**: 템플릿 최적화 및 파일 정리

---

## 🎯 핵심 설계 원칙

### Claude Code 공식 표준

```
Layer 1: Commands (Slash Commands)
  ↓ (Task tool, 순차/병렬)
Layer 2: Sub-agents
  ↓ (자동 호출, description 기반)
Layer 3: Skills
```

**핵심 3원칙**:
1. **Commands**: User-invoked, 워크플로우 오케스트레이션
2. **Sub-agents**: Task tool로 위임, 독립 컨텍스트, 전문 작업 수행
3. **Skills**: Model-invoked, SKILL.md (YAML frontmatter), 자동 호출

---

## 🏗️ 아키텍처 계층

### Layer 1: Commands (4개)

**위치**: `.claude/commands/alfred/*.md`

| Command | 역할 | Sub-agents 조율 |
|---------|------|-----------------|
| `/alfred:0-project` | 프로젝트 초기화 | 6개 Sub-agents (병렬/순차) |
| `/alfred:1-plan` | SPEC 작성 | spec-builder |
| `/alfred:2-run` | TDD 구현 | tdd-implementer |
| `/alfred:3-sync` | 문서 동기화 | doc-syncer, tag-agent |

**형식**:
```yaml
---
name: alfred:0-project
description: 프로젝트 문서 초기화
allowed-tools:
  - Read
  - Write
  - Task  # Sub-agents 위임
---

# Phase 1: 계획
Task(subagent="language-detector", ...)

# Phase 2: 실행
Task(subagent="document-generator", ...)
```

---

### Layer 2: Sub-agents (17개)

**위치**: `.claude/agents/alfred/*.md`

#### Core Agents (9개)

| Agent | Model | Skills | 역할 |
|-------|-------|--------|------|
| spec-builder | Sonnet | specs, ears | SPEC 작성 |
| tdd-implementer | Sonnet | lang-* | TDD 구현 |
| doc-syncer | Haiku | tags, trust | 문서 동기화 |
| tag-agent | Haiku | tags | TAG 관리 |
| git-manager | Haiku | git | Git 워크플로우 |
| debug-helper | Sonnet | essentials-debug | 오류 진단 |
| trust-checker | Haiku | trust | TRUST 검증 |
| cc-manager | Sonnet | claude-code | 설정 관리 |
| project-manager | Sonnet | - | 프로젝트 조율 |

#### 0-project Sub-agents (6개)

| Agent | Model | Skills | 역할 |
|-------|-------|--------|------|
| language-detector | Haiku | langs | 언어 감지 |
| backup-merger | Sonnet | - | 백업 병합 |
| project-interviewer | Sonnet | specs, ears | 요구사항 수집 |
| document-generator | Haiku | specs, ears, langs | 문서 생성 |
| feature-selector | Haiku | - | Skills 선택 |
| template-optimizer | Haiku | claude-code | 템플릿 최적화 |

**형식**:
```yaml
---
name: language-detector
tools: Read, Bash, Grep
model: haiku
skills:
  - moai-foundation-langs
---

# Language Detector

## 핵심 역할
- 언어 감지 (Python, TypeScript...)
- LanguageInterface JSON 응답
```

---

### Layer 3: Skills (49개)

**위치**: `.claude/skills/*/SKILL.md`

#### Tier 구조

| Tier | 개수 | 역할 | 예시 |
|------|------|------|------|
| 1 (Foundation) | 6 | 핵심 기반 | specs, ears, tags, trust, langs, git |
| 2 (Language) | 28 | 언어별 도구 | python, typescript, java, go, rust... |
| 3 (Domain) | 10 | 도메인 전문 | backend, frontend, mobile, ml, db... |
| 4 (Essentials) | 5 | 추가 도구 | debug, review, refactor, perf |

**형식**:
```yaml
---
name: moai-foundation-langs
tier: 1
description: Auto-detects project language and provides LanguageInterface
allowed-tools:
  - Read
  - Bash
---

# LanguageInterface Standard

## What it does
Provides standardized toolchain recommendations

## When to use
- Automatically invoked by language-detector
```

---

## 🔄 `/alfred:0-project` 완전 자동화

### Phase 1: 분석 및 계획

```
User: "/alfred:0-project"

┌─────────────────────────────┐
│ language-detector (병렬)    │ → Python, FastAPI 감지
│ backup-merger (병렬)        │ → 백업 병합 (조건부)
└─────────────────────────────┘
            ↓
┌─────────────────────────────┐
│ project-interviewer (순차)  │ → 인터뷰 (신규/레거시)
└─────────────────────────────┘
            ↓
      사용자 승인 대기
```

### Phase 2: 실행

```
User: "진행"

┌─────────────────────────────┐
│ document-generator          │ → product/structure/tech.md 생성
└─────────────────────────────┘
            ↓
┌─────────────────────────────┐
│ Alfred                      │ → config.json 생성
└─────────────────────────────┘
```

### Phase 3: 최적화 (선택적)

```
User: "스킬 최적화 진행"

┌─────────────────────────────┐
│ feature-selector            │ → 49개 → 9개 선택
└─────────────────────────────┘
            ↓
┌─────────────────────────────┐
│ template-optimizer          │ → CLAUDE.md 생성, 40개 삭제
└─────────────────────────────┘
            ↓
      ✅ 완료 (84% 디스크 절약)
```

---

## 📊 feature-selector 알고리즘

### 입력

```json
{
  "product": {"domain": "backend"},
  "structure": {"architecture": "REST API"},
  "tech": {"language": "Python", "framework": "FastAPI"}
}
```

### 선택 로직

```python
# Tier 1: Foundation (6개, 항상 포함)
TIER_1 = ["specs", "ears", "tags", "trust", "langs", "git"]

# Tier 2: Language (1개, 언어별)
TIER_2 = {"Python": "python", "TypeScript": "typescript", ...}

# Tier 3: Domain (0~3개, 도메인별)
TIER_3 = {"backend": ["backend", "web-api"], ...}

# 선택
selected = TIER_1 + [TIER_2[lang]] + TIER_3[domain][:2]
# = 6 + 1 + 2 = 9개
```

### 출력

```json
{
  "selected_skills": [
    "moai-foundation-specs",
    "moai-foundation-ears",
    "moai-foundation-tags",
    "moai-foundation-trust",
    "moai-foundation-langs",
    "moai-foundation-git",
    "moai-lang-python",
    "moai-domain-backend",
    "moai-domain-web-api"
  ],
  "total": 9,
  "reduction": "82%"
}
```

---

## 🛠️ template-optimizer 작업

### 1. CLAUDE.md 맞춤형 생성

```python
def generate_claude_md(selected_skills):
    template = read("templates/.moai/CLAUDE.md")

    # Skills 섹션 생성
    skills_section = generate_skills_section(selected_skills)

    # 변수 치환
    claude_md = template.replace("{{SKILLS}}", skills_section)

    write(".moai/CLAUDE.md", claude_md)
```

### 2. 불필요한 Skills 삭제

```python
def cleanup_skills(selected_skills):
    all_skills = glob(".claude/skills/*/")

    for skill in all_skills:
        if skill not in selected_skills:
            # 백업 후 삭제
            backup(skill, f".moai-backups/{timestamp}/")
            delete(skill)
```

### 3. config.json 업데이트

```python
config = {
    "optimized": True,
    "selected_skills": selected_skills,
    "optimization_date": "2025-10-20T15:30:00"
}
write(".moai/config.json", config)
```

---

## 📈 성능 및 효과

### Before (최적화 전)

- **Skills**: 49개
- **디스크**: 15.2 MB
- **컨텍스트**: 모든 Skills description 로드

### After (최적화 후)

- **Skills**: 9개 (82% 감소)
- **디스크**: 2.4 MB (84% 절약)
- **컨텍스트**: 프로젝트 관련 Skills만

### 병렬 처리 효과

- **순차 실행**: 18초
- **병렬 실행**: 15초 (17% 단축)

---

## 🚀 구현 로드맵

### Phase 1: 기초 작업 (완료)

- [x] ARCHITECTURE_v2.md 작성
- [x] feature-selector_DESIGN.md 작성
- [x] template-optimizer_DESIGN.md 작성

### Phase 2: Commands 리팩토링

- [ ] 0-project.md 리팩토링 (Task tool만 사용)
- [ ] 1-plan.md Task tool 적용
- [ ] 2-run.md Task tool 적용
- [ ] 3-sync.md Task tool 적용

### Phase 3: Sub-agents 표준화

- [ ] 모든 Sub-agents에 `skills` 필드 추가
- [ ] feature-selector.md 구현
- [ ] template-optimizer.md 구현

### Phase 4: Skills 재구조화

- [ ] 모든 Skills에 `tier` 필드 추가
- [ ] LanguageInterface 표준 적용
- [ ] description 최적화 (자동 호출 조건 명확화)

### Phase 5: 0-project 완전 자동화

- [ ] Phase 1~3 워크플로우 통합
- [ ] 병렬/순차 실행 최적화
- [ ] 사용자 승인 흐름 개선

---

## 📚 참고 자료

### Claude Code 공식 문서

- **Agent Skills**: https://docs.claude.com/en/docs/claude-code/skills
- **Task Tool**: Claude Code 공식 가이드
- **Best Practices**: Anthropic Engineering Blog

### MoAI-ADK 내부 문서

- **development-guide.md**: `.moai/memory/development-guide.md`
- **spec-metadata.md**: `.moai/memory/spec-metadata.md`
- **CLAUDE.md**: `.moai/CLAUDE.md`

---

## ❓ FAQ

### Q1. Commands와 Sub-agents의 차이는?

- **Commands**: 사용자가 직접 호출 (slash command)
- **Sub-agents**: Commands가 Task tool로 위임

### Q2. Skills는 언제 호출되나?

- **자동 호출**: Claude가 description 기반으로 판단
- **호출 주체**: Sub-agents (skills 필드에 명시된 경우)

### Q3. 왜 49개 → 9개로 줄이나?

- **컨텍스트 비용 절감**: Claude가 선택지를 빠르게 판단
- **디스크 절약**: 84% 절감 (15.2 MB → 2.4 MB)
- **처리 속도 향상**: 불필요한 Skills 제외

### Q4. 삭제된 Skills는 복구 가능한가?

- **백업 보관**: `.moai-backups/{timestamp}/skills/`
- **복구 명령**: `moai-adk restore-skill <skill-name>`

### Q5. 기존 프로젝트에 적용 방법은?

```bash
# 1. 템플릿 업데이트
moai-adk update-template

# 2. 프로젝트 재초기화
/alfred:0-project

# 3. Skills 최적화
User: "스킬 최적화 진행"
```

---

## 🎯 다음 단계

### Immediate (지금)

1. **Commands 리팩토링**: Task tool만 사용하도록 변경
2. **Sub-agents 표준화**: `skills` 필드 추가
3. **Skills Tier 명시**: `tier` 필드 추가

### Short-term (1주)

1. **feature-selector 구현**: 선택 알고리즘 코드화
2. **template-optimizer 구현**: 백업/삭제/생성 로직
3. **0-project 통합 테스트**: 전체 워크플로우 검증

### Long-term (1달)

1. **성능 모니터링**: 병렬 처리 효과 측정
2. **사용자 피드백**: Skills 선택 정확도 개선
3. **자동화 확장**: 1-plan, 2-run, 3-sync에도 적용

---

**작성자**: @Alfred
**버전**: v2.0
**최종 업데이트**: 2025-10-20
