# MoAI-ADK Commands/Agents/Skills 구조 개선 분석 보고서

> **공식 Claude Code Bundling Pattern 기반 최적화 제안**
>
> 분석 일자: 2025-10-19
> 분석자: @Alfred

---

## 📊 Executive Summary

### 핵심 발견사항

| 지표 | 현재 상태 | 문제점 | 영향 |
|-----|----------|-------|------|
| **단일 파일 크기** | 최대 1,177 라인 (cc-manager) | 컨텍스트 비효율 | 매번 전체 로드 필수 |
| **Bundling 패턴 사용** | 0개 / 49개 스킬 | 모듈화 부족 | 유지보수 어려움 |
| **대형 커맨드 파일** | 4개 (500~990 라인) | 섹션별 독립 관리 불가 | 확장성 저하 |
| **cc-manager 중복** | 템플릿 지침 중복 | 최근 스킬로 분리됨 | 일관성 문제 |

### 권장 조치

1. **즉시 실행** (High Priority):
   - cc-manager.md 리팩토링 (1,177 → ~300 라인)
   - 대형 스킬에 Bundling 패턴 적용

2. **단계적 실행** (Medium Priority):
   - 대형 커맨드 파일 분리 (0-project, 1-plan, 3-sync)
   - settings-template, plugins-template 스킬 생성

3. **장기 계획** (Low Priority):
   - 모든 스킬에 reference.md 추가 (고급 기능용)
   - 사용 사례별 가이드 문서 분리

---

## 🔍 현재 상태 분석

### 1. 파일 크기 분포

#### Commands

| 파일 | 라인 수 | 섹션 수 | 상태 | 비고 |
|-----|--------|---------|------|------|
| `0-project.md` | 990 | 19 | ❌ 과대 | 환경 분석, 인터뷰, 최적화 모두 포함 |
| `1-plan.md` | 657 | ? | ⚠️ 큼 | 분석 필요 |
| `3-sync.md` | 649 | ? | ⚠️ 큼 | 분석 필요 |
| `2-run.md` | 553 | ? | ⚠️ 큼 | 분석 필요 |
| `1-spec.md` | 31 | ? | ✅ 적정 | 별칭 파일 |
| `2-build.md` | 30 | ? | ✅ 적정 | 별칭 파일 |

**권장 크기**: 300 라인 이하 (공식 패턴 기준)

#### Agents

| 파일 | 라인 수 | 섹션 수 | 하위 섹션 | 상태 | 비고 |
|-----|--------|---------|-----------|------|------|
| `cc-manager.md` | 1,177 | 27 | 59 | ❌ 심각 | Commands/Agents/Skills/Plugins/Settings 모두 포함 |
| `git-manager.md` | 477 | ? | ? | ⚠️ 큼 | 분석 필요 |
| `project-manager.md` | 424 | ? | ? | ⚠️ 큼 | 분석 필요 |
| `tdd-implementer.md` | 409 | ? | ? | ⚠️ 큼 | 분석 필요 |
| `spec-builder.md` | 350 | ? | ? | ✅ 적정 | 경계선 |
| `trust-checker.md` | 331 | ? | ? | ✅ 적정 | 경계선 |

**권장 크기**: 300 라인 이하

#### Skills

| 파일 | 라인 수 | Bundling 사용 | 상태 | 비고 |
|-----|--------|--------------|------|------|
| `moai-cc-agent-template` | 897 | ❌ | ⚠️ 큼 | 템플릿 파일이라 예외 |
| `moai-cc-skill-template` | 606 | ❌ | ⚠️ 큼 | 템플릿 파일이라 예외 |
| `moai-cc-command-template` | 590 | ❌ | ⚠️ 큼 | 템플릿 파일이라 예외 |
| `moai-alfred-template-generator` | 516 | ❌ | ⚠️ 큼 | Bundling 후보 1순위 |
| `moai-alfred-feature-selector` | 440 | ❌ | ⚠️ 큼 | Bundling 후보 2순위 |
| 기타 언어/도메인 스킬 | 72~78 | ❌ | ✅ 적정 | 간결함 |

**현재 Bundling 사용률**: 0% (0개 / 49개 스킬)

---

## 📐 공식 패턴 vs MoAI-ADK 비교

### Claude Code 공식 Bundling Pattern

**이미지 예시**: `pdf/` 스킬

```
pdf/
├── SKILL.md          # 메인 가이드 (Overview + Quick Start)
├── reference.md      # 고급 참조 (pypdfium2, 상세 API)
└── forms.md          # 사용 사례 (PDF 양식 채우기)
```

**SKILL.md 내부 참조**:
```markdown
## Overview

This guide covers essential PDF processing operations using Python
libraries and command-line tools. For advanced features,
JavaScript libraries, and detailed examples, see ./reference.md.
If you need to fill out a PDF form, read ./forms.md and follow its
instructions.
```

**핵심 원칙**:
1. **SKILL.md**: 간결한 메인 가이드 (기본 사용법)
2. **reference.md**: 고급 기능, 상세 API, 추가 라이브러리
3. **{use-case}.md**: 특정 사용 사례별 가이드
4. **JIT Loading**: 필요할 때만 해당 파일 참조

### MoAI-ADK 현재 방식

**현재 구조**: 모든 스킬

```
moai-alfred-template-generator/
└── SKILL.md          # 모든 내용 포함 (516 라인)
```

**문제점**:
```markdown
## 전체 내용이 하나의 파일에

### 기본 사용법
...

### 고급 기능
...

### 예제 1
...

### 예제 2
...

### 트러블슈팅
...

### 성능 최적화
...
```

**비교 요약**:

| 항목 | 공식 패턴 | MoAI-ADK 현재 | 차이 |
|-----|----------|--------------|------|
| **파일 수** | 1 + N (bundling) | 1 (단일) | Bundling 미사용 |
| **메인 파일 크기** | ~200 라인 | 500~1,000 라인 | 2~5배 큼 |
| **컨텍스트 로딩** | JIT (필요 시) | 전체 로드 | 비효율적 |
| **모듈화** | 섹션별 독립 파일 | 모든 섹션 하나의 파일 | 유지보수 어려움 |
| **확장성** | 새 파일 추가 | 파일 길이 증가 | 확장 제한적 |

---

## 🎯 개선이 필요한 파일 목록

### Priority 1: 즉시 리팩토링 필요 (Critical)

#### 1. `cc-manager.md` (1,177 라인 → 300 라인)

**현재 포함 내용**:
- ✅ 에이전트 페르소나 (유지)
- ✅ 핵심 역할 (유지)
- ❌ 커맨드 표준 템플릿 (→ moai-cc-command-template 스킬로 이동 완료)
- ❌ 에이전트 표준 템플릿 (→ moai-cc-agent-template 스킬로 이동 완료)
- ❌ Skills 시스템 (→ moai-cc-skill-template 스킬로 이동 완료)
- ⚠️ Plugins 시스템 (→ moai-cc-plugin-template 스킬 생성 권장)
- ⚠️ 권한 설정 최적화 (→ moai-cc-settings-template 스킬 생성 권장)
- ✅ 표준 검증 체크리스트 (유지)
- ✅ 자동 검증 기능 (유지)

**리팩토링 계획**:

```
.claude/agents/alfred/cc-manager.md (1,177 → ~300 라인)
├─ 제거: 커맨드/에이전트/스킬 템플릿 (이미 스킬로 분리)
├─ 위임: Plugins → moai-cc-plugin-template 스킬
├─ 위임: Settings → moai-cc-settings-template 스킬
└─ 유지: 에이전트 페르소나, 핵심 역할, 검증 기능
```

**예상 효과**:
- 컨텍스트 비용 **74% 감소** (1,177 → 300 라인)
- 템플릿 스킬과 역할 명확히 분리
- cc-manager는 "검증 및 조율" 역할에만 집중

#### 2. `0-project.md` (990 라인 → 400 라인 + bundling)

**현재 포함 내용**:
- ✅ 커맨드 목적, 실행 흐름 (유지)
- ⚠️ STEP 1: 환경 분석 및 인터뷰 계획 (227 라인 → interview-guides.md)
- ⚠️ 프로젝트 유형별 인터뷰 가이드 (34 라인 → interview-guides.md)
- ⚠️ 오류 처리 (24 라인 → error-handling.md)
- ⚠️ STEP 3: 프로젝트 맞춤형 최적화 (107 라인 → optimization-guide.md)
- ✅ STEP 2: 프로젝트 초기화 실행 (유지)

**Bundling 계획**:

```
.claude/commands/alfred/0-project/
├── 0-project.md              # 메인 커맨드 (400 라인)
│   ├─ 커맨드 목적, 실행 흐름
│   ├─ STEP 2: 실행 로직
│   └─ 참조: "인터뷰 가이드는 ./interview-guides.md 참조"
├── interview-guides.md       # 인터뷰 전략 (261 라인)
│   ├─ 프로젝트 유형별 인터뷰 트리
│   ├─ 질문 템플릿
│   └─ 금지 사항
├── error-handling.md         # 오류 처리 (50 라인)
│   ├─ 백업 복구 시나리오
│   ├─ 파일 누락 처리
│   └─ 권한 오류 처리
└── optimization-guide.md     # 최적화 가이드 (200 라인)
    ├─ Phase 3 워크플로우
    ├─ feature-selector 통합
    └─ template-generator 사용법
```

**참조 방식**:
```markdown
## 🚀 STEP 1: 환경 분석 및 인터뷰 계획 수립

프로젝트 환경을 분석하고 체계적인 인터뷰 계획을 수립합니다.

**상세 가이드**: 프로젝트 유형별 인터뷰 전략 및 질문 트리는
`./interview-guides.md`를 참조하세요.

### 간략 실행 흐름
1. 환경 분석
2. 인터뷰 계획 수립
3. 사용자 승인

(상세 내용은 interview-guides.md에 위임)
```

**예상 효과**:
- 메인 파일 **60% 축소** (990 → 400 라인)
- 섹션별 독립 관리 가능
- 인터뷰 가이드 업데이트 시 메인 커맨드 영향 없음

---

### Priority 2: 단계적 리팩토링 (High)

#### 3. `moai-alfred-template-generator` (516 라인 → 200 + bundling)

**Bundling 계획**:

```
.claude/skills/moai-alfred-template-generator/
├── SKILL.md                  # 메인 가이드 (200 라인)
│   ├─ 스킬 목적
│   ├─ Quick Start
│   └─ 기본 사용법
├── reference.md              # 고급 기능 (200 라인)
│   ├─ 템플릿 엔진 상세
│   ├─ 변수 시스템
│   └─ 커스터마이징
└── examples.md               # 사용 사례 (116 라인)
    ├─ Python 프로젝트 예시
    ├─ TypeScript 프로젝트 예시
    └─ 멀티 언어 프로젝트 예시
```

#### 4. `moai-alfred-feature-selector` (440 라인 → 150 + bundling)

**Bundling 계획**:

```
.claude/skills/moai-alfred-feature-selector/
├── SKILL.md                  # 메인 가이드 (150 라인)
│   ├─ 스킬 목적
│   ├─ Quick Start
│   └─ 프로젝트 분류 체계
├── decision-tree.md          # 결정 트리 (150 라인)
│   ├─ 언어별 분류 규칙
│   ├─ 도메인별 분류 규칙
│   └─ 스킬 선택 매트릭스
└── recommendations.md        # 추천 로직 (140 라인)
    ├─ 프로젝트 유형별 권장 스킬
    ├─ 최소/표준/완전 세트
    └─ 성능 최적화 팁
```

#### 5. `git-manager.md` (477 라인 → 300 + bundling)

**Bundling 계획**:

```
.claude/agents/alfred/git-manager/
├── git-manager.md            # 메인 에이전트 (300 라인)
│   ├─ 에이전트 페르소나
│   ├─ 핵심 역할
│   └─ 워크플로우
├── commit-standards.md       # 커밋 표준 (100 라인)
│   ├─ TDD 단계별 커밋 메시지
│   ├─ Locale 기반 메시지
│   └─ 예시 및 템플릿
└── pr-workflow.md            # PR 워크플로우 (77 라인)
    ├─ Draft PR 생성
    ├─ PR Ready 전환
    └─ 자동 머지 규칙
```

---

### Priority 3: 장기 계획 (Medium)

#### 6. `1-plan.md`, `3-sync.md`, `2-run.md` (각 550~650 라인)

**분석 필요**: 파일 구조 상세 분석 후 Bundling 계획 수립

#### 7. 나머지 에이전트 파일

**project-manager.md** (424 라인), **tdd-implementer.md** (409 라인) 등:
- 300 라인 이상 파일들은 Bundling 패턴 적용 검토

---

## 💡 구체적 리팩토링 제안

### 제안 1: cc-manager 즉시 리팩토링

**현재 상태**:
- 1,177 라인
- Commands/Agents/Skills/Plugins/Settings 템플릿 모두 포함

**리팩토링 후**:
- 300 라인 (74% 감소)
- 템플릿은 모두 스킬로 위임
- cc-manager는 검증 및 조율만 담당

**실행 계획**:

```markdown
## cc-manager.md (리팩토링 후)

---
name: cc-manager
description: "Use when: Claude Code 파일 표준 검증 및 최적화가 필요할 때"
tools: Read, Grep, Glob, Bash
model: sonnet
---

# Claude Code Manager - 검증 및 조율

## 🎭 에이전트 페르소나
(유지)

## 🎯 핵심 역할

### 1. 표준 검증
- YAML frontmatter 검증
- 파일 명명 규칙 확인
- 권한 설정 검토

### 2. 템플릿 위임
- Command 파일 생성 → @moai-cc-command-template 스킬 호출
- Agent 파일 생성 → @moai-cc-agent-template 스킬 호출
- Skill 파일 생성 → @moai-cc-skill-template 스킬 호출
- Plugin 설정 → @moai-cc-plugin-template 스킬 호출 (예정)
- Settings 최적화 → @moai-cc-settings-template 스킬 호출 (예정)

### 3. 자동 검증
- 파일 생성 후 자동 검증 실행
- 표준 위반 시 수정 제안
- 체크리스트 자동 실행

## 🔍 표준 검증 체크리스트
(유지 - 300 라인 내에서)

## 🚨 자동 검증 및 수정 기능
(유지)
```

**제거할 섹션** (877 라인):
- 커맨드 표준 템플릿 지침 (→ moai-cc-command-template)
- 에이전트 표준 템플릿 지침 (→ moai-cc-agent-template)
- Skills 시스템 (→ moai-cc-skill-template)
- Plugins 시스템 (→ moai-cc-plugin-template 생성 예정)
- 권한 설정 최적화 (→ moai-cc-settings-template 생성 예정)

---

### 제안 2: Bundling Pattern 도입 (0-project)

**디렉토리 구조 변경**:

```
Before:
.claude/commands/alfred/0-project.md (990 라인)

After:
.claude/commands/alfred/0-project/
├── 0-project.md (400 라인)
├── interview-guides.md (261 라인)
├── error-handling.md (50 라인)
└── optimization-guide.md (200 라인)
```

**0-project.md (리팩토링 후)**:

```markdown
---
name: alfred:0-project
description: 프로젝트 문서 초기화
allowed-tools:
  - Read
  - Write
  - Task
---

# 📋 MoAI-ADK 0단계: 프로젝트 문서 초기화

## 🎯 커맨드 목적
(유지)

## 📋 실행 흐름
1. **환경 분석**: 프로젝트 유형 및 언어 자동 감지
2. **인터뷰 전략 수립**: 상세 가이드는 `./interview-guides.md` 참조
3. **사용자 확인**: 계획 검토 및 승인
4. **프로젝트 문서 작성**: product/structure/tech.md 생성
5. **오류 처리**: 문제 발생 시 `./error-handling.md` 참조

## 🚀 STEP 1: 환경 분석 및 인터뷰 계획

프로젝트 환경을 분석하고 체계적인 인터뷰 계획을 수립합니다.

**📖 상세 가이드**: 프로젝트 유형별 인터뷰 전략, 질문 트리,
금지 사항은 `./interview-guides.md`를 참조하세요.

### 1.0 백업 디렉토리 확인
(간략 요약, 상세는 error-handling.md)

### 1.1 프로젝트 환경 분석
(핵심 로직만 유지)

## 🚀 STEP 2: 프로젝트 초기화 실행
(핵심 로직 유지 - 대부분의 라인)

## 🚀 STEP 3: 프로젝트 맞춤형 최적화 (선택적)

**📖 상세 가이드**: Phase 3 워크플로우, feature-selector 통합은
`./optimization-guide.md`를 참조하세요.
```

**interview-guides.md** (새 파일):

```markdown
# 프로젝트 인터뷰 가이드

> `/alfred:0-project` 커맨드의 인터뷰 전략 상세 가이드

## 🎯 목적

프로젝트 유형별 최적 인터뷰 전략을 제공합니다.

## 📋 프로젝트 유형 분류

### 신규 프로젝트
...

### 기존 프로젝트
...

## 🌳 인터뷰 질문 트리

### Python 프로젝트
...

### TypeScript 프로젝트
...

## ⚠️ 금지 사항

절대 하지 말아야 할 작업:
- ❌ `.claude/memory/` 디렉토리에 파일 생성
- ❌ 날짜와 수치 예측
...
```

---

### 제안 3: 새 템플릿 스킬 생성

#### moai-cc-plugin-template

**목적**: Plugins 시스템 템플릿 제공 (cc-manager에서 분리)

```yaml
---
name: moai-cc-plugin-template
description: Claude Code MCP Plugin 설정 템플릿
model: haiku
allowed-tools:
  - Read
  - Write
---
```

**내용 출처**: cc-manager.md의 "Plugins 시스템" 섹션 (153 라인)

#### moai-cc-settings-template

**목적**: Claude Code settings.json 최적화 템플릿 제공

```yaml
---
name: moai-cc-settings-template
description: Claude Code settings.json 최적화 템플릿
model: haiku
allowed-tools:
  - Read
  - Write
  - Edit
---
```

**내용 출처**: cc-manager.md의 "권한 설정 최적화" 섹션 (104 라인)

---

## 📈 예상 효과

### 컨텍스트 효율

| 파일 | Before (라인) | After (라인) | 감소율 | 비고 |
|-----|--------------|-------------|--------|------|
| cc-manager.md | 1,177 | 300 | **74%** | 템플릿 스킬로 위임 |
| 0-project.md | 990 | 400 | **60%** | Bundling 적용 |
| template-generator | 516 | 200 | **61%** | Bundling 적용 |
| feature-selector | 440 | 150 | **66%** | Bundling 적용 |
| **전체 평균** | - | - | **65%** | 대형 파일 기준 |

### JIT Loading 효과

**Before** (Bundling 없음):
```
사용자: "커맨드 파일 생성 방법 알려줘"
Alfred: cc-manager.md 전체 로드 (1,177 라인)
       → Commands, Agents, Skills, Plugins 모두 로드
       → 실제 필요: Commands 섹션만 (66 라인)
       → 낭비: 1,111 라인 (94%)
```

**After** (Bundling 적용):
```
사용자: "커맨드 파일 생성 방법 알려줘"
Alfred: moai-cc-command-template 스킬 로드 (590 라인)
       → Commands 템플릿만 로드
       → 필요한 정보만 정확히 로드
       → 낭비: 0%
```

**컨텍스트 비용 절감**:
- cc-manager 참조 시: **74% 감소** (1,177 → 300)
- 특정 템플릿 필요 시: **95% 감소** (1,177 → 66) - 스킬 직접 호출

### 유지보수성 향상

| 항목 | Before | After | 개선 |
|-----|--------|-------|------|
| **섹션 독립 관리** | 불가 | 가능 | 파일별 독립 수정 |
| **확장성** | 파일 길이 증가 | 새 파일 추가 | 무한 확장 가능 |
| **충돌 위험** | 높음 | 낮음 | 파일별 독립 |
| **검색 효율** | 낮음 | 높음 | 파일명으로 빠른 찾기 |

### 개발자 경험 개선

**Before**:
```
개발자: "인터뷰 가이드 어디 있지?"
       → 0-project.md 990 라인 스크롤...
       → 227 라인 분량 찾음
       → 수정 후 전체 파일 다시 로드
```

**After**:
```
개발자: "인터뷰 가이드 어디 있지?"
       → interview-guides.md 직접 열기
       → 261 라인만 확인
       → 수정 후 해당 파일만 재로드
```

---

## 🚀 실행 계획

### Phase 1: 즉시 실행 (1주일)

**목표**: cc-manager 리팩토링 + 2개 템플릿 스킬 생성

| 작업 | 예상 시간 | 담당 | 우선순위 |
|-----|----------|------|----------|
| cc-manager.md 리팩토링 | 2시간 | Alfred | ⚠️ Critical |
| moai-cc-plugin-template 생성 | 1시간 | Alfred | 🔴 High |
| moai-cc-settings-template 생성 | 1시간 | Alfred | 🔴 High |
| cc-manager 테스트 | 1시간 | Alfred | 🔴 High |

**체크리스트**:
- [ ] cc-manager.md 1,177 → 300 라인으로 축소
- [ ] Commands/Agents/Skills 템플릿 섹션 제거
- [ ] moai-cc-plugin-template 스킬 생성
- [ ] moai-cc-settings-template 스킬 생성
- [ ] cc-manager에 스킬 위임 로직 추가
- [ ] 검증: cc-manager가 스킬을 올바르게 호출하는지 확인

### Phase 2: 단계적 실행 (2주일)

**목표**: 대형 커맨드/스킬에 Bundling 패턴 적용

| 작업 | 예상 시간 | 담당 | 우선순위 |
|-----|----------|------|----------|
| 0-project Bundling | 3시간 | Alfred | 🔴 High |
| template-generator Bundling | 2시간 | Alfred | 🟡 Medium |
| feature-selector Bundling | 2시간 | Alfred | 🟡 Medium |
| git-manager Bundling | 2시간 | Alfred | 🟡 Medium |

**0-project Bundling 상세**:
1. `.claude/commands/alfred/0-project/` 디렉토리 생성
2. `0-project.md` → 400 라인으로 축소
3. `interview-guides.md` 생성 (261 라인)
4. `error-handling.md` 생성 (50 라인)
5. `optimization-guide.md` 생성 (200 라인)
6. 메인 파일에 참조 링크 추가
7. 테스트: 커맨드 정상 작동 확인

### Phase 3: 장기 계획 (1개월)

**목표**: 나머지 대형 파일 Bundling + 표준화

| 작업 | 예상 시간 | 담당 | 우선순위 |
|-----|----------|------|----------|
| 1-plan, 3-sync, 2-run 분석 | 2시간 | Alfred | 🟢 Low |
| 나머지 커맨드 Bundling | 4시간 | Alfred | 🟢 Low |
| 나머지 에이전트 Bundling | 4시간 | Alfred | 🟢 Low |
| Bundling 가이드 문서 작성 | 2시간 | Alfred | 🟢 Low |
| 전체 검증 및 테스트 | 3시간 | Alfred | 🟢 Low |

---

## 📝 Bundling Pattern 가이드라인

### 언제 Bundling을 적용할까?

**Bundling 적용 기준**:
- ✅ 파일 크기 ≥ 500 라인
- ✅ 명확히 구분되는 섹션 3개 이상
- ✅ 일부 섹션만 자주 참조됨
- ✅ 고급 기능과 기본 기능이 혼재

**Bundling 불필요**:
- ❌ 파일 크기 < 300 라인
- ❌ 모든 섹션이 항상 필요함
- ❌ 섹션 간 강한 결합

### Bundling 파일 분류 기준

**공식 패턴 기반**:

| 파일명 | 용도 | 예시 |
|-------|------|------|
| `{name}.md` | 메인 가이드 (Overview + Quick Start) | SKILL.md, 0-project.md |
| `reference.md` | 고급 참조 (상세 API, 추가 라이브러리) | pdf/reference.md |
| `{use-case}.md` | 특정 사용 사례별 가이드 | pdf/forms.md, interview-guides.md |
| `examples.md` | 예제 모음 | 실전 사용 예시 |
| `troubleshooting.md` | 트러블슈팅 | 오류 처리, 문제 해결 |

### 참조 링크 작성법

**공식 패턴**:
```markdown
For advanced features, see ./reference.md.
If you need to fill out a PDF form, read ./forms.md and follow its instructions.
```

**MoAI-ADK 권장**:
```markdown
**📖 상세 가이드**: 프로젝트 유형별 인터뷰 전략은 `./interview-guides.md`를 참조하세요.

**고급 기능**: 템플릿 엔진 커스터마이징은 `./reference.md`를 참조하세요.

**트러블슈팅**: 오류 발생 시 `./troubleshooting.md`를 참조하세요.
```

### 디렉토리 구조 표준

**Commands**:
```
.claude/commands/alfred/{command-name}/
├── {command-name}.md       # 메인 커맨드
├── reference.md            # 고급 참조 (선택)
└── {use-case}.md           # 사용 사례 (선택)
```

**Agents**:
```
.claude/agents/alfred/{agent-name}/
├── {agent-name}.md         # 메인 에이전트
├── workflows.md            # 워크플로우 상세 (선택)
└── standards.md            # 표준 및 규칙 (선택)
```

**Skills**:
```
.claude/skills/{skill-name}/
├── SKILL.md                # 메인 스킬 (필수)
├── reference.md            # 고급 참조 (선택)
├── examples.md             # 사용 예시 (선택)
└── {use-case}.md           # 특정 사용 사례 (선택)
```

---

## 🔗 참고 자료

### Claude Code 공식 문서

- [Skills Documentation](https://docs.claude.com/en/docs/claude-code/skills)
- [Agent Skills Best Practices](https://docs.claude.com/en/docs/agents-and-tools/agent-skills/best-practices)
- [Bundling Additional Content](이미지 참조)

### MoAI-ADK 관련 문서

- `.moai/memory/development-guide.md` - MoAI-ADK 개발 가이드
- `.moai/memory/spec-metadata.md` - SPEC 메타데이터 표준
- `CLAUDE.md` - 프로젝트 루트 지침

---

## 📊 결론

### 핵심 발견

1. **Bundling Pattern 미적용**: 49개 스킬 중 0개가 Bundling 사용
2. **대형 파일 비대화**: cc-manager (1,177), 0-project (990), 1-plan (657) 등
3. **컨텍스트 비효율**: 매번 전체 파일 로드 필요
4. **템플릿 중복**: cc-manager와 템플릿 스킬 간 중복

### 권장 조치

**즉시 실행**:
- ✅ cc-manager.md 리팩토링 (1,177 → 300 라인, **74% 감소**)
- ✅ moai-cc-plugin-template 스킬 생성
- ✅ moai-cc-settings-template 스킬 생성

**단계적 실행**:
- ✅ 0-project Bundling (990 → 400 + bundling, **60% 감소**)
- ✅ 대형 스킬 Bundling (template-generator, feature-selector)

**예상 효과**:
- 컨텍스트 비용 **65% 감소** (대형 파일 기준)
- JIT Loading으로 **필요한 정보만 로드**
- 유지보수성 **대폭 향상**
- 확장성 **무한 확장 가능**

---

**최종 업데이트**: 2025-10-19
**작성자**: @Alfred
**다음 단계**: Phase 1 실행 (cc-manager 리팩토링)
