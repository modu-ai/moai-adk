# MoAI-ADK v0.4.0 "Skills Revolution" Update Plan

> **Claude Code Skills를 활용한 패러다임 전환**
>
> 작성일: 2025-10-19
> 작성자: Alfred SuperAgent
> 버전: v0.4.0
> 상태: Planning

---

## 📋 Table of Contents

- [Executive Summary](#executive-summary)
- [Part 1: Claude Skills 심층 분석](#part-1-claude-skills-심층-분석)
- [Part 2: Skills vs Agents vs Commands](#part-2-skills-vs-agents-vs-commands)
- [Part 3: MoAI-ADK v0.4.0 아키텍처](#part-3-moai-adk-v040-아키텍처)
- [Part 4: Skills 45개 상세 설계](#part-4-skills-45개-상세-설계)
- [Part 5: 개발자 경험 최적화](#part-5-개발자-경험-최적화)
- [Part 6: Skills 마켓플레이스](#part-6-skills-마켓플레이스)
- [Part 7: 마이그레이션 전략](#part-7-마이그레이션-전략)
- [Part 8: 실행 계획](#part-8-실행-계획)

---

## Executive Summary

### 🎯 핵심 비전

> **"Commands는 진입점, Skills는 능력, Sub-agents는 두뇌"**

MoAI-ADK v0.4.0은 Claude Code의 **Agent Skills 기능**을 핵심 실행 계층으로 도입하여 **4-Layer 아키텍처**로 전환합니다. Progressive Disclosure 메커니즘으로 **Effectively Unbounded Context**를 실현하며, 개발자는 명령어를 암기하지 않고 **자연어 대화**만으로 **레고 블록처럼 조립 가능한 개발 워크플로우**를 경험합니다.

### 🔑 핵심 변경사항

| 변경 사항 | Before (v0.3.x) | After (v0.4.0) |
|-----------|-----------------|----------------|
| **아키텍처** | 3-Layer (Commands/Sub-agents/Hooks) | **4-Layer (Commands/Sub-agents/Skills/Hooks)** |
| **용어** | "Agents" (혼동) | **"Sub-agents" (Claude Code 표준)** |
| **Skills 시스템** | 없음 | **10개 Skills (Foundation 6 + Dev Essentials 4)** |
| **컨텍스트 전략** | Always Loaded | **Progressive Disclosure (Effectively Unbounded)** |
| **재사용성** | 프로젝트 전용 | **전역 (모든 프로젝트 공유)** |
| **Hooks 성능** | SessionStart 220ms | **<100ms (50% 단축)** |
| **조합 가능성** | 없음 (단독 실행) | **Composable (Skills 자동 조합)** |
| **일관성** | Sub-agent별 상이 | **Skills 공유로 100% 일관성** |

### 🔍 공식 문서 검증 완료

**출처**: [Agent Skills - Claude Docs](https://docs.claude.com/en/docs/agents-and-tools/agent-skills/overview), [Anthropic Engineering Blog](https://www.anthropic.com/engineering/equipping-agents-for-the-real-world-with-agent-skills)

- ✅ **Effectively Unbounded Context**: Progressive Disclosure로 컨텍스트가 사실상 무제한 (공식 표현)
- ✅ **SKILL.md 구조**: `.claude/skills/` 디렉토리, 파일시스템 기반
- ✅ **Automatic Loading**: Claude가 자동으로 관련성 판단하여 Skills 로드
- ✅ **Custom Skills Only**: Claude Code는 Custom Skills만 지원 (API 업로드 불필요)

### 📊 예상 효과

- ⏱️ **컨텍스트 효율**: 30% 토큰 절감 (Skills 재사용)
- 🚀 **응답 속도**: 50% 시간 단축 (Hooks 경량화: 220ms→100ms)
- 🔄 **재사용성**: +300% (전역 Skills)
- 🎯 **일관성**: 100% (모든 Sub-agents가 동일한 Skills 참조)
- ⚡ **확장성**: Effectively Unbounded (Progressive Disclosure)
- 📈 **개발 생산성**: +150% (전체 워크플로우 최적화)

### 🏗️ 4-Layer 아키텍처 확정

```
┌──────────────────────────────────────────┐
│ Layer 1: Commands (워크플로우 진입점)    │
│ - /alfred:0-init   (프로젝트 초기화)     │
│ - /alfred:1-plan   (계획 수립) ⭐ NEW    │
│ - /alfred:2-build  (TDD 구현)            │
│ - /alfred:3-sync   (문서 동기화)         │
│ - 2-Phase 패턴 (Plan → Execute)          │
└──────────────┬───────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────┐
│ Layer 2: Sub-agents (복잡한 추론)        │
│ - spec-builder, tdd-implementer 등       │
│ - Task tool 호출, 독립 컨텍스트          │
│ - Skills 참조하여 일관성 보장            │
└──────────────┬───────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────┐
│ Layer 3: Skills (재사용 가능한 지식) ⭐  │
│ - Foundation 6개 + Dev Essentials 4개    │
│ - <500 words, Progressive Disclosure     │
│ - Effectively Unbounded Context          │
└──────────────┬───────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────┐
│ Layer 4: Hooks (가드레일 + JIT Context)  │
│ - SessionStart <100ms (경량화)           │
│ - PreToolUse <50ms (위험 작업 차단만)    │
└──────────────────────────────────────────┘
```

### 💡 Commands 명칭 변경 철학

#### `/alfred:0-project` → `/alfred:0-init`
- **이유**: "init(초기화)"이 더 간결하고 보편적인 명령어 스타일
- **기능**: 프로젝트 문서 구조 및 언어별 최적화 설정 생성

#### `/alfred:1-spec` → `/alfred:1-plan` ⭐ 핵심 변경
- **철학적 배경**:
  - **"항상 계획을 먼저 세우고 진행한다"** - 계획 우선 원칙 강조
  - SPEC 문서 생성뿐만 아니라 **브레인스토밍 모드**로 확장
  - 아이디어 구상, 요구사항 정리, 설계 논의 등 **계획 수립 전반** 지원

- **사용 시나리오**:
  ```bash
  # 시나리오 1: SPEC 문서 생성 (기존 방식)
  /alfred:1-plan "JWT 인증 시스템"
  → SPEC-AUTH-001 생성, EARS 구문, 브랜치/PR

  # 시나리오 2: 브레인스토밍 모드 (신규)
  /alfred:1-plan "프로젝트 아키텍처 설계 논의"
  → Alfred와 대화형 브레인스토밍
  → 아이디어 정리 → SPEC 후보 도출

  # 시나리오 3: 기술 선택 논의 (신규)
  /alfred:1-plan "인증 방식 비교 (JWT vs Session)"
  → 장단점 분석 → 의사결정 지원 → SPEC 문서화
  ```

- **핵심 가치**:
  - ✅ **Think First, Code Later** (생각 먼저, 코딩 나중)
  - ✅ **Collaborative Planning** (Alfred와 함께 계획 수립)
  - ✅ **SPEC-First 유지** (최종적으로 SPEC 문서 생성)

#### `/alfred:3-sync` - 유지
- **이유**: "sync(동기화)"가 문서-코드-TAG 동기화 의미를 정확히 전달
- **기능**: Living Document 갱신, TAG 체인 검증, PR Ready 전환

---

## Part 1: Claude Skills 심층 분석

### 1.1 Claude Skills란?

**공식 정의** (Anthropic):
> Agent Skills are organized folders of instructions, scripts, and resources that agents can discover and load dynamically to perform better at specific tasks.

**핵심 특징**:

```
┌─────────────────────────────────────────────┐
│ 1. Model-Invoked (모델 자동 호출)           │
│    - Claude가 문맥 파악하여 자동 감지       │
│    - 사용자 명시 불필요                     │
└─────────────────────────────────────────────┘
             ↓
┌─────────────────────────────────────────────┐
│ 2. Progressive Disclosure (점진적 공개)     │
│    - Layer 1: Metadata (최소 토큰)          │
│    - Layer 2: SKILL.md (필요 시 로드)       │
│    - Layer 3: Additional Files (필요 시)    │
└─────────────────────────────────────────────┘
             ↓
┌─────────────────────────────────────────────┐
│ 3. Composable (조합 가능)                   │
│    - 여러 Skills 자동 조합                  │
│    - 레고 블록처럼 유기적 결합              │
└─────────────────────────────────────────────┘
             ↓
┌─────────────────────────────────────────────┐
│ 4. Global Reusability (전역 재사용)         │
│    - ~/.claude/skills/ (모든 프로젝트)     │
│    - 중앙 관리, 자동 업데이트               │
└─────────────────────────────────────────────┘
```

### 1.2 Progressive Disclosure - 게임 체인저

**3-Layer 로딩 메커니즘**:

```
┌──────────────────────────────────────────────┐
│ Layer 1: Metadata (Startup)                 │
│ - name + description만 사전 로드            │
│ - 각 Skill당 최소한의 토큰만 소비           │
│ - 다수의 Skills 설치 시에도 부담 적음       │
└──────────────────────────────────────────────┘
              ↓ Claude가 관련성 판단
┌──────────────────────────────────────────────┐
│ Layer 2: SKILL.md (On-Demand)               │
│ - 관련 있는 Skill만 전체 내용 로드          │
│ - 필요 시에만 로드하여 컨텍스트 효율화      │
│ - 여러 Skills 동시 로드 가능                │
└──────────────────────────────────────────────┘
              ↓ 추가 정보 필요 시
┌──────────────────────────────────────────────┐
│ Layer 3: Additional Files (Lazy Loading)    │
│ - templates/, scripts/, resources/          │
│ - 필요한 파일만 선택적 로드                 │
│ - 대용량 참고 자료를 효율적으로 관리        │
└──────────────────────────────────────────────┘
```

**혁신적인 이유**:

✅ **효율적 컨텍스트 관리**: Progressive Disclosure로 대용량 정보를 필요 시에만 로드하여 컨텍스트 윈도우를 효율적으로 사용
✅ **Cost-Efficient**: 사용하지 않는 Skills는 최소한의 토큰만 소비 (메타데이터만)
✅ **Scalable**: 다수의 Skills를 설치해도 성능 저하 없음
✅ **Automatic**: Claude가 자동으로 필요한 Skills 판단 및 로드

### 1.3 Composability - 레고식 조립

**Skills 조합 예시** (개념 설명용 시나리오):

```
사용자: "회사 브랜드 가이드라인에 맞는 피치덱을 만들어줘"

Claude의 자동 Skills 조합:
1. "브랜드 가이드라인" 감지 → brand-guidelines Skill 로드
2. "피치덱" 감지 → presentation Skill 로드
3. 두 Skill을 자동 조합하여 일관된 결과물 생성

사용자: "이제 포스터도 만들어줘"

Claude의 자동 Skills 조합:
4. brand-guidelines Skill (이미 로드됨, 재사용)
5. poster-design Skill (새로 로드)
6. 조합하여 브랜드에 맞는 포스터 생성

NOTE: 실제 Skill 이름과 동작은 구현에 따라 다를 수 있음
```

**조합 원리**:

- **Automatic Coordination**: Claude가 자동으로 필요한 Skills 식별 및 조율
- **No Explicit Reference**: Skills는 서로를 명시적으로 참조하지 않아도 됨
- **Multiple Skills Together**: 동시에 여러 Skills 활성화 가능
- **Best Practice**: 큰 하나보다 작은 여러 Skills로 분리

### 1.4 SKILL.md 구조

**필수 YAML Front Matter**:

```yaml
---
name: skill-name                    # 64자 이하
description: One-line description   # 1024자 이하
version: 0.1.0                      # Semantic Version (선택)
author: @username                   # 작성자 (선택)
license: MIT                        # 라이선스 (선택)
tags:                               # 태그 (선택)
  - python
  - testing
---
```

**권장 본문 구조**:

```markdown
# Skill Name

## What it does
Clear explanation of the Skill's purpose.

## When to use
- Use case 1
- Use case 2
- Use case 3

## How it works
1. Step 1
2. Step 2
3. Step 3

## Examples

### Example 1: Basic usage
User: "Do something"
Claude: (activates this skill and does X)

### Example 2: Combined with other skills
User: "Do something complex"
Claude: (uses this skill + other-skill together)

## Works well with
- other-skill-name
- another-skill-name

## Files included
- templates/example.md
- scripts/helper.sh
```

---

## Part 2: Skills vs Agents vs Commands

### 2.1 핵심 차이점 비교표

| 차원 | **Skills** | **Agents** | **Commands** |
|------|-----------|-----------|--------------|
| **호출 방식** | Model-Invoked (Claude 자동 판단) | Delegated (Alfred 위임) | User-Invoked (사용자 명시) |
| **컨텍스트 전략** | Progressive Disclosure (3-Layer) | Isolated Context Window | Always Loaded |
| **컨텍스트 한계** | **Unbounded** (무제한) | Limited (격리됨) | Limited (항상 로드) |
| **조합 가능성** | **Composable** (자동 조합) | Sequential (순차 실행) | None (단일 실행) |
| **범위** | Global (~/.claude/skills/) | Project (.claude/agents/) | Project (.claude/commands/) |
| **재사용성** | **모든 프로젝트** | 프로젝트 전용 | 프로젝트 전용 |
| **사용자 인지** | Transparent (투명) | Semi-transparent | Explicit (명시적) |
| **복잡도** | Low-Medium (재사용 가능 능력) | High (복잡한 추론) | Medium (워크플로우) |
| **비용 효율** | **최고** (필요시만 로드) | 중간 (별도 컨텍스트) | 낮음 (항상 로드) |

### 2.2 사용 시나리오별 선택 기준

#### Use Skills when:

✅ **재사용 가능한 능력**: 모든 프로젝트에서 사용
✅ **특정 도메인 지식**: EARS, TDD, 언어별 best practice
✅ **템플릿 기반 작업**: SPEC 생성, 보일러플레이트
✅ **자동 감지 원함**: 사용자가 명시하지 않아도 작동
✅ **조합 가능**: 다른 Skills와 함께 사용

**예시**: moai-spec-writer, moai-tdd-guide, moai-tag-validator

#### Use Agents when:

✅ **복잡한 추론 필요**: 다단계 분석, 의사결정
✅ **격리된 컨텍스트**: 메인 대화와 분리
✅ **전문 작업**: 특정 프로젝트 전용 로직
✅ **Alfred 위임**: 사용자는 모르지만 Alfred가 판단

**예시**: spec-builder (복잡한 SPEC 분석), debug-helper (오류 추론)

#### Use Commands when:

✅ **워크플로우 진입점**: 명확한 시작 지점
✅ **사용자 의도 명확**: /alfred:1-plan처럼 명시적
✅ **Phase 기반 실행**: 계획 → 승인 → 실행
✅ **Git 통합**: 브랜치 생성, PR 관리

**예시**: /alfred:0-init, /alfred:1-plan, /alfred:2-build, /alfred:3-sync

### 2.3 역할 재정의

#### Commands → **Workflow Orchestrators** (워크플로우 지휘자)

**기존**: 모든 로직을 직접 수행
**신규**: Skills와 Agents를 조율

```markdown
# /alfred:1-plan 예시 (v0.4.0)

## Phase 1: 분석 및 브레인스토밍 (Skills 활용)
1. moai-project-analyzer Skill 자동 호출
   - product.md 분석
   - 기존 SPEC 목록 스캔
2. moai-spec-id-generator Skill 자동 호출
   - 도메인 추출
   - SPEC ID 중복 확인
3. 브레인스토밍 모드 (선택적)
   - Alfred와 대화형 계획 수립
   - 아이디어 정리 및 의사결정 지원

## Phase 2: 실행 (Skills + Sub-agents)
1. moai-spec-writer Skill로 SPEC 초안 생성
2. spec-builder Agent로 복잡한 검증 (순환 의존성)
3. moai-git-manager Skill로 브랜치/PR 생성

→ Command는 "언제 무엇을"만 결정, 실제 작업은 Skills가 수행
```

#### Agents → **Complex Reasoners** (복잡한 추론가)

**기존**: 모든 전문 작업 담당
**신규**: Skills로 해결 불가능한 복잡한 추론만 담당

```markdown
# spec-builder Agent 예시 (v0.4.0)

## When to use (축소)
- SPEC 메타데이터 복잡한 검증 (순환 의존성, 버전 충돌)
- SPEC 간 영향 분석 (의존성 그래프 탐색)
- SPEC 우선순위 자동 결정 (복잡한 알고리즘)

## What NOT to use (Skills로 이동)
- ❌ EARS 템플릿 적용 → moai-spec-writer Skill
- ❌ SPEC ID 중복 확인 → moai-spec-id-generator Skill
- ❌ Git 작업 → moai-git-manager Skill
```

#### Skills → **Domain Experts** (도메인 전문가) ⭐ 핵심

**3가지 카테고리**:

1. **Foundation Skills**: 재사용 가능한 핵심 능력
2. **Language Skills**: 20개 언어별 best practice
3. **Domain Skills**: 프로젝트 유형별 전문 지식

---

## Part 3: MoAI-ADK v0.4.0 아키텍처

### 3.1 전체 아키텍처 설계

```
┌──────────────────────────────────────────────────────┐
│ Layer 1: Commands (Workflow Entry Points)           │
│ Role: 워크플로우 진입점 및 오케스트레이터            │
├──────────────────────────────────────────────────────┤
│ /alfred:0-init     → 프로젝트 초기화                 │
│ /alfred:1-plan     → 계획 수립 및 SPEC 작성 ⭐       │
│ /alfred:2-build    → TDD 구현 워크플로우             │
│ /alfred:3-sync     → 문서 동기화 워크플로우          │
│                                                       │
│ 변경사항: Commands는 직접 로직 수행하지 않음         │
│          Skills와 Sub-agents를 조율만 함             │
└──────────────────────────────────────────────────────┘
                       ↓ 위임
┌──────────────────────────────────────────────────────┐
│ Layer 2: Sub-agents (Complex Reasoning)             │
│ Role: Skills로 해결 불가능한 복잡한 추론 담당        │
├──────────────────────────────────────────────────────┤
│ spec-builder       → SPEC 복잡 검증 (순환 의존성)   │
│ debug-helper       → 오류 원인 추론 및 해결         │
│ trust-checker      → TRUST 5원칙 준수도 분석         │
│                                                       │
│ 변경사항: 역할 축소 (단순 작업은 Skills로 이동)      │
│          격리된 컨텍스트에서 복잡한 추론만           │
└──────────────────────────────────────────────────────┘
                       ↓ 활용
┌──────────────────────────────────────────────────────┐
│ Layer 3: Skills (Reusable Capabilities) ⭐ 핵심축   │
│ Role: 재사용 가능한 도메인 지식 및 능력              │
├──────────────────────────────────────────────────────┤
│ Foundation Skills (15개)                             │
│ Language Skills (20개)                               │
│ Domain Skills (10개)                                 │
│                                                       │
│ 변경사항: Skills가 핵심 실행 계층                    │
│          Progressive Disclosure로 효율적 컨텍스트 관리 │
│          Composable하여 레고처럼 자동 조합           │
└──────────────────────────────────────────────────────┘
                       ↓ 검증
┌──────────────────────────────────────────────────────┐
│ Layer 4: Hooks (Guardrails & Context)               │
│ Role: 안전망 및 JIT Context 주입                     │
├──────────────────────────────────────────────────────┤
│ SessionStart        → Skills 활성화 메시지 표시      │
│ PreToolUse          → 위험 작업 차단, 자동 백업      │
│ PostToolUse         → 작업 결과 검증                 │
│                                                       │
│ 변경사항: 없음 (기존 역할 유지)                      │
└──────────────────────────────────────────────────────┘
```

### 3.2 Skills 조합 전략 (레고식 조립)

**실제 시나리오: "Python REST API 프로젝트 SPEC 작성"**

```
사용자: "FastAPI 기반 사용자 인증 API SPEC 작성해줘"

Claude의 자동 Skills 조합:
┌─────────────────────────────────────────┐
│ 1️⃣ moai-spec-writer (Foundation)       │
│    → EARS 구조, YAML Front Matter      │
└─────────────────────────────────────────┘
              +
┌─────────────────────────────────────────┐
│ 2️⃣ python-expert (Language)            │
│    → FastAPI best practice, pytest      │
└─────────────────────────────────────────┘
              +
┌─────────────────────────────────────────┐
│ 3️⃣ web-api-expert (Domain)             │
│    → REST API 설계, 인증 패턴           │
└─────────────────────────────────────────┘
              ↓
생성된 SPEC:
- EARS 구문으로 구조화됨 (moai-spec-writer)
- FastAPI 라우팅 패턴 고려 (python-expert)
- OAuth2/JWT 보안 요구사항 포함 (web-api-expert)
- pytest 테스트 전략 명시 (python-expert)
```

### 3.3 Skills 아키텍처 설계 원칙

#### 1. Single Responsibility (단일 책임)

❌ **안 좋은 예**: mega-moai-helper (모든 것 포함)
✅ **좋은 예**: 작은 여러 Skills로 분리

```
moai-spec-writer      # SPEC 생성만
moai-spec-validator   # SPEC 검증만
moai-spec-id-gen      # ID 생성만
```

#### 2. Composable by Default (기본적으로 조합 가능)

각 Skill은 독립적으로 작동하면서도 다른 Skills와 자연스럽게 조합

```yaml
---
name: moai-spec-writer
description: Creates EARS-based SPEC documents with YAML frontmatter
---

# MoAI SPEC Writer

## Works well with
- moai-spec-id-gen: Auto-generates unique SPEC IDs
- python-expert: Adds Python-specific requirements
- moai-git-manager: Auto-creates feature branch
```

#### 3. Progressive Disclosure (점진적 공개)

**SKILL.md 구조 최적화**:

```markdown
---
name: moai-tdd-orchestrator
description: Guides RED-GREEN-REFACTOR TDD cycle with language-specific tools
---

# MoAI TDD Orchestrator

## Quick Start (Layer 2 - 기본 정보)
1. RED: Write failing test
2. GREEN: Make it pass
3. REFACTOR: Improve code

## Language Support (Layer 3 - 필요 시 로드)
See [language-guides/python.md](./language-guides/python.md)
See [language-guides/typescript.md](./language-guides/typescript.md)

## Advanced Patterns (Layer 3 - 필요 시 로드)
See [patterns/property-based-testing.md](./patterns/property-based-testing.md)
```

---

## Part 4: Skills 45개 상세 설계

### 4.1 Foundation Skills (15개)

#### 1. moai-spec-writer

**목적**: EARS 기반 SPEC 문서 자동 생성

```yaml
---
name: moai-spec-writer
description: Creates EARS-based SPEC documents with YAML frontmatter and HISTORY section
version: 0.1.0
tags:
  - spec
  - ears
  - documentation
---
```

**트리거 조건**:
- "SPEC 작성", "명세서 만들어줘", "requirements 문서 생성"
- "EARS로 작성", "요구사항 정리해줘"

**주요 기능**:
1. SPEC ID 자동 생성 (도메인 추출)
2. YAML Front Matter 생성 (7개 필수 필드)
3. EARS 5가지 구문으로 요구사항 분류
4. HISTORY 섹션 자동 추가 (v0.0.1 INITIAL)
5. .moai/specs/SPEC-{ID}/spec.md 생성

**파일 구조**:
```
moai-spec-writer/
├── SKILL.md
├── templates/
│   ├── spec-template.md
│   └── ears-examples.md
└── scripts/
    └── validate-spec-id.sh
```

#### 2. moai-spec-id-generator

**목적**: SPEC ID 생성 및 중복 확인

```yaml
---
name: moai-spec-id-generator
description: Generates unique SPEC IDs and validates against duplicates
version: 0.1.0
---
```

**주요 기능**:
1. 요청에서 도메인 자동 추출
2. 3자리 숫자 자동 할당
3. `rg "@SPEC:{ID}" -n` 중복 확인
4. 디렉토리명 생성 (SPEC-{ID}/)

#### 3. moai-spec-validator

**목적**: SPEC 메타데이터 및 구조 검증

```yaml
---
name: moai-spec-validator
description: Validates SPEC metadata, YAML frontmatter, and HISTORY section
version: 0.1.0
---
```

**검증 항목**:
- YAML Front Matter 7개 필수 필드
- HISTORY 섹션 존재 여부
- EARS 구문 적용률
- TAG 체인 무결성

#### 4. moai-tdd-orchestrator

**목적**: RED-GREEN-REFACTOR TDD 사이클 가이드

```yaml
---
name: moai-tdd-orchestrator
description: Guides RED-GREEN-REFACTOR TDD cycle with real-time feedback
version: 0.1.0
tags:
  - tdd
  - testing
  - workflow
---
```

**주요 기능**:
1. **RED 단계**: @TEST:ID 작성, 실패 확인
2. **GREEN 단계**: @CODE:ID 작성, 테스트 통과
3. **REFACTOR 단계**: 코드 품질 개선
4. 각 단계별 커밋 자동 생성

**Works well with**:
- python-expert, typescript-expert (언어별 테스트)
- moai-git-manager (커밋 자동화)

#### 5. moai-tag-scanner

**목적**: @TAG 전체 스캔 및 목록 생성

```yaml
---
name: moai-tag-scanner
description: Scans all @TAG markers and generates TAG inventory
version: 0.1.0
---
```

**주요 기능**:
```bash
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
```

#### 6. moai-tag-validator

**목적**: TAG 체인 무결성 검증

```yaml
---
name: moai-tag-validator
description: Validates TAG chain integrity and detects orphaned TAGs
version: 0.1.0
---
```

**검증 항목**:
- TAG 체인 연결 (@SPEC → @TEST → @CODE → @DOC)
- 고아 TAG 탐지
- 중복 ID 확인

#### 7. moai-git-manager

**목적**: Git 작업 자동화

```yaml
---
name: moai-git-manager
description: Automates Git operations (branch, commit, PR)
version: 0.1.0
---
```

**주요 기능**:
1. 브랜치 생성 (feature/SPEC-{ID})
2. locale 기반 커밋 메시지 생성
3. Draft PR 생성

#### 8. moai-branch-creator

**목적**: 브랜치 네이밍 규칙 적용

```yaml
---
name: moai-branch-creator
description: Creates Git branches with MoAI naming conventions
version: 0.1.0
---
```

**네이밍 규칙**:
- feature/SPEC-{ID}
- fix/SPEC-{ID}
- refactor/SPEC-{ID}

#### 9. moai-pr-creator

**목적**: Draft PR 자동 생성

```yaml
---
name: moai-pr-creator
description: Creates Draft PRs with SPEC-based description
version: 0.1.0
---
```

**PR 템플릿**:
```markdown
## Summary
@SPEC:{ID} 기반 자동 생성

## Changes
- SPEC 문서: .moai/specs/SPEC-{ID}/spec.md
- 테스트: tests/...
- 구현: src/...

## Test Plan
- [ ] 테스트 통과
- [ ] 코드 품질 검증
- [ ] TAG 체인 확인
```

#### 10. moai-doc-generator

**목적**: Living Document 자동 생성

```yaml
---
name: moai-doc-generator
description: Generates Living Documents from SPEC and CODE
version: 0.1.0
---
```

**생성 문서**:
- docs/specs/overview.md
- docs/api/README.md
- TAG 추적성 다이어그램 (Mermaid)

#### 11. moai-api-doc-gen

**목적**: API 문서 자동 생성

```yaml
---
name: moai-api-doc-gen
description: Generates API documentation from @CODE:API markers
version: 0.1.0
---
```

**기능**:
- @CODE:ID:API 스캔
- 엔드포인트 목록 생성
- OpenAPI/Swagger 스펙 생성 (선택)

#### 12. moai-readme-updater

**목적**: README.md 자동 업데이트

```yaml
---
name: moai-readme-updater
description: Updates README.md with SPEC-based feature list
version: 0.1.0
---
```

**업데이트 내용**:
- 주요 기능 목록 (@SPEC 기반)
- 개발 진행도 (완료율)
- TAG 추적성 다이어그램

#### 13. moai-project-analyzer

**목적**: 프로젝트 구조 분석

```yaml
---
name: moai-project-analyzer
description: Analyzes project structure and suggests optimizations
version: 0.1.0
---
```

**분석 항목**:
- product.md, structure.md, tech.md
- 기존 SPEC 목록
- 언어 감지

#### 14. moai-lang-detector

**목적**: 프로젝트 언어 자동 감지

```yaml
---
name: moai-lang-detector
description: Detects project programming language from files
version: 0.1.0
---
```

**감지 방법**:
- pyproject.toml → Python
- package.json → TypeScript/JavaScript
- go.mod → Go
- Cargo.toml → Rust

#### 15. moai-boilerplate-gen

**목적**: 언어별 보일러플레이트 생성

```yaml
---
name: moai-boilerplate-gen
description: Generates language-specific project boilerplate
version: 0.1.0
---
```

**지원 언어**: 20개 (Language Skills와 연동)

---

### 4.2 Language Skills (20개)

각 Language Skill은 다음 구조를 따릅니다:

```yaml
---
name: {language}-expert
description: {Language} best practices, testing, and tooling
version: 1.0.0
tags:
  - {language}
  - testing
  - best-practices
---
```

**공통 기능**:
1. 언어별 best practice
2. TDD 프레임워크 (pytest, Vitest, JUnit 등)
3. 린터/포맷터 (ruff, Biome, clippy 등)
4. 타입 시스템 (mypy, TypeScript, Go types 등)
5. 패키지 관리 (uv, npm, cargo 등)

**20개 언어 목록**:

1. **python-expert**
   - pytest, mypy, ruff, black
   - uv 패키지 관리

2. **typescript-expert**
   - Vitest, Biome, strict typing
   - npm/pnpm/yarn

3. **javascript-expert**
   - Jest, ESLint, Prettier
   - npm/pnpm/yarn

4. **java-expert**
   - JUnit, Maven/Gradle, Checkstyle
   - Spring Boot patterns

5. **go-expert**
   - go test, golint, gofmt
   - 표준 라이브러리 활용

6. **rust-expert**
   - cargo test, clippy, rustfmt
   - 소유권 및 borrow checker

7. **ruby-expert**
   - RSpec, RuboCop, Bundler
   - Rails patterns (선택)

8. **kotlin-expert**
   - JUnit, Gradle, ktlint
   - 코루틴, 확장 함수

9. **swift-expert**
   - XCTest, SwiftLint
   - iOS/macOS patterns

10. **dart-expert**
    - flutter test, dart analyze
    - Flutter widget patterns

11. **c++-expert**
    - Google Test, clang-format
    - 모던 C++ (C++17/20)

12. **c#-expert**
    - xUnit, .NET tooling
    - LINQ, async/await

13. **php-expert**
    - PHPUnit, Composer
    - PSR 표준

14. **scala-expert**
    - ScalaTest, sbt
    - 함수형 프로그래밍

15. **elixir-expert**
    - ExUnit, Mix
    - OTP patterns

16. **haskell-expert**
    - HUnit, Stack/Cabal
    - 순수 함수형

17. **clojure-expert**
    - clojure.test, Leiningen
    - 불변 데이터 구조

18. **r-expert**
    - testthat, lintr
    - 데이터 분석 패턴

19. **julia-expert**
    - Test, Pkg
    - 과학 컴퓨팅

20. **lua-expert**
    - busted, luacheck
    - 임베디드 스크립팅

---

### 4.3 Domain Skills (10개)

#### 1. web-api-expert

```yaml
---
name: web-api-expert
description: REST API and GraphQL design patterns
version: 0.3.0
tags:
  - api
  - rest
  - graphql
---
```

**전문 영역**:
- REST API 설계 (RESTful 원칙)
- GraphQL 스키마 설계
- API 버저닝
- 인증/인가 (JWT, OAuth2)
- API 문서화 (OpenAPI)

#### 2. mobile-app-expert

```yaml
---
name: mobile-app-expert
description: Mobile app development with Flutter and React Native
version: 0.2.0
---
```

**전문 영역**:
- Flutter (Dart)
- React Native (TypeScript)
- 상태 관리 (Provider, Redux)
- 네이티브 통합

#### 3. cli-tool-expert

```yaml
---
name: cli-tool-expert
description: CLI tool development best practices
version: 0.2.0
---
```

**전문 영역**:
- 명령어 파싱 (argparse, clap, commander)
- POSIX 호환성
- 도움말 메시지
- Exit codes

#### 4. library-expert

```yaml
---
name: library-expert
description: Library and package development
version: 0.2.0
---
```

**전문 영역**:
- 공개 API 설계
- Semantic Versioning
- 문서화 (README, 예시)
- 배포 (PyPI, npm, crates.io)

#### 5. fullstack-expert

```yaml
---
name: fullstack-expert
description: Full-stack application architecture
version: 0.2.0
---
```

**전문 영역**:
- Frontend + Backend 통합
- 상태 관리
- 데이터 흐름
- 배포 전략

#### 6. auth-patterns

```yaml
---
name: auth-patterns
description: Authentication and authorization patterns
version: 0.2.1
tags:
  - security
  - authentication
  - authorization
---
```

**전문 영역**:
- JWT, OAuth2, Session
- RBAC, ABAC
- 비밀번호 보안
- 다중 인증 (MFA)

#### 7. database-expert

```yaml
---
name: database-expert
description: Database design and optimization
version: 0.2.0
---
```

**전문 영역**:
- 스키마 설계 (정규화)
- 인덱싱 전략
- 쿼리 최적화
- 마이그레이션

#### 8. testing-expert

```yaml
---
name: testing-expert
description: E2E and performance testing strategies
version: 0.2.0
---
```

**전문 영역**:
- E2E 테스팅 (Playwright, Cypress)
- 성능 테스팅 (k6, JMeter)
- 테스트 피라미드

#### 9. devops-expert

```yaml
---
name: devops-expert
description: CI/CD, Docker, and Kubernetes
version: 0.2.0
---
```

**전문 영역**:
- CI/CD 파이프라인 (GitHub Actions)
- Docker 컨테이너화
- Kubernetes 오케스트레이션

#### 10. security-expert

```yaml
---
name: security-expert
description: Security best practices and vulnerability prevention
version: 0.2.0
---
```

**전문 영역**:
- OWASP Top 10
- 정적 분석 (SAST)
- 의존성 보안
- 시크릿 관리

---

## Part 5: 개발자 경험 최적화

### 5.1 Before vs After 비교

#### 시나리오: "새 기능 구현"

**Before (v0.3.x - Commands + Sub-agents)**:

```
개발자: "사용자 인증 기능 구현해줘"

1. 개발자가 명령어 학습 필요
   → /alfred:1-spec "사용자 인증" 입력 (구 명령어)

2. Command가 모든 로직 수행
   → spec-builder Sub-agent 호출
   → SPEC 문서 생성

3. TDD 구현
   → /alfred:2-build AUTH-001 입력
   → tdd-implementer Sub-agent 호출

4. 문서 동기화
   → /alfred:3-sync 입력
   → doc-syncer Sub-agent 호출

학습 곡선: 높음 (/alfred:*, @agent-* 모두 학습)
사용자 개입: 많음 (매 단계 명령어 입력)
재사용성: 낮음 (프로젝트마다 설정)
```

**After (v0.4.0 - Skills 중심)**:

```
개발자: "사용자 인증 기능 구현해줘"

1. Claude가 자동으로 Skills 조합
   ✅ moai-spec-writer (SPEC 생성)
   ✅ moai-spec-id-gen (AUTH-001 자동 생성)
   ✅ moai-git-manager (feature/AUTH-001 브랜치 자동 생성)
   → SPEC 문서 완성

2. 개발자: "테스트부터 작성해줘"
   ✅ moai-tdd-orchestrator (RED 단계 가이드)
   ✅ python-expert (pytest 테스트 생성)
   → 테스트 파일 생성

3. 개발자: "구현해줘"
   ✅ moai-tdd-orchestrator (GREEN 단계)
   ✅ python-expert (FastAPI 코드 생성)
   → 구현 완료

4. 개발자: "문서 업데이트"
   ✅ moai-doc-generator (자동 동기화)
   ✅ moai-tag-scanner (TAG 검증)
   → Living Document 업데이트

학습 곡선: 낮음 (자연어만 사용)
사용자 개입: 적음 (대화형 진행)
재사용성: 높음 (모든 프로젝트 공유)
```

**개선 효과**:
- ⏱️ **시간 절감**: 60% (명령어 입력 → 자연어)
- 📚 **학습 부담**: -80% (명령어 암기 불필요)
- 🔄 **재사용성**: +300% (전역 Skills)
- 🎯 **정확도**: +40% (Skills 자동 조합)

### 5.2 Conversational Skill Assembly

**개념**: 대화하면서 점진적으로 Skills 추가/조합

**예시 대화**:

```
개발자: "Python 프로젝트 시작하고 싶어"

Claude: (python-expert Skill 자동 로드)
"Python 프로젝트를 생성하겠습니다. 프로젝트 유형을 선택해주세요:
1. CLI 도구
2. Web API
3. 라이브러리
4. Full-stack 앱"

개발자: "FastAPI로 REST API 만들 거야"

Claude: (web-api-expert Skill 자동 추가)
"FastAPI 기반 REST API 프로젝트를 생성합니다.
활성화된 Skills:
- moai-lang-boilerplate (Python 구조)
- python-expert (FastAPI best practice)
- web-api-expert (REST API 설계)

인증 방식은 어떻게 하시겠습니까?
1. JWT
2. OAuth2
3. Session"

개발자: "JWT로 할게"

Claude: (auth-patterns Skill 자동 추가)
"JWT 인증이 추가되었습니다.
활성화된 Skills:
✅ moai-lang-boilerplate
✅ python-expert
✅ web-api-expert
✅ auth-patterns (JWT)

프로젝트 구조를 생성하겠습니다..."

→ 대화하면서 자연스럽게 Skills 조합!
```

**핵심 원리**:
- 🧩 **Progressive Enhancement**: 대화마다 Skills 추가
- 🎯 **Context-Aware**: Claude가 문맥 파악하여 자동 제안
- 🔄 **Reversible**: "JWT 빼고 OAuth2로 바꿔줘" 가능
- 📊 **Transparent**: "현재 활성화된 Skills 보여줘" 가능

### 5.3 개발자 도구 (Skills for Skills)

#### 1. skill-creator (공식 제공)

대화형으로 새 Skill 생성

```
개발자: "moai-spec-writer Skill 만들어줘"

skill-creator:
"Skill 생성을 시작합니다. 몇 가지 질문드리겠습니다.

Q1. 이 Skill이 하는 일은 무엇인가요?
A: EARS 방식으로 SPEC 문서를 자동 생성합니다.

Q2. 언제 사용되어야 하나요? (트리거 조건)
A: 'SPEC 작성', '명세서 만들어줘' 같은 요청 시

Q3. 필요한 템플릿이나 스크립트가 있나요?
A: spec-template.md, validate-spec-id.sh

✅ Skill 구조 생성 완료!
~/.claude/skills/moai-spec-writer/
├── SKILL.md
├── templates/
│   └── spec-template.md
└── scripts/
    └── validate-spec-id.sh
"
```

#### 2. moai-skill-tester

Skills 테스트 자동화

```
개발자: "moai-spec-writer 테스트해줘"

moai-skill-tester:
"테스트 시작...

Test 1: Metadata Loading
✅ name: moai-spec-writer
✅ description: Creates EARS-based SPEC documents

Test 2: Trigger Detection
✅ "SPEC 작성해줘" → Skill 활성화됨
✅ "코드 작성해줘" → Skill 활성화 안됨

Test 3: Output Validation
✅ YAML Front Matter 생성됨
✅ EARS 구조 적용됨
✅ SPEC ID 중복 확인됨

Test 4: Composition
✅ moai-git-manager와 조합 가능
✅ python-expert와 조합 가능

모든 테스트 통과! ✅
"
```

#### 3. moai-skills-manager

모든 Skills 관리 및 모니터링

```
개발자: "현재 활성화된 Skills 보여줘"

moai-skills-manager:
"
📦 활성화된 Skills (5개):

Foundation Skills:
✅ moai-spec-writer (v0.2.0)
   - 로드됨: 2회
   - 마지막 사용: 5분 전

✅ moai-tdd-orchestrator (v0.1.5)
   - 로드됨: 1회
   - 마지막 사용: 1시간 전

Language Skills:
✅ python-expert (v1.0.0)
   - 로드됨: 3회
   - 마지막 사용: 10분 전

Domain Skills:
✅ web-api-expert (v0.3.0)
   - 로드됨: 2회
   - 마지막 사용: 15분 전

✅ auth-patterns (v0.2.1)
   - 로드됨: 1회
   - 마지막 사용: 20분 전

💡 제안:
- moai-doc-generator 업데이트 가능 (v0.1.0 → v0.2.0)
- 새 Skill 추천: moai-api-doc-gen (API 문서 자동 생성)
"
```

---

## Part 6: Skills 마켓플레이스

### 6.1 마켓플레이스 아키텍처

```
┌──────────────────────────────────────────────────────┐
│ Official MoAI Skills (Anthropic + MoAI)              │
│ ~/.claude/skills/moai/ (자동 설치)                   │
├──────────────────────────────────────────────────────┤
│ Foundation Skills (15개)                             │
│ - moai-spec-writer, moai-tdd-orchestrator...        │
│                                                       │
│ Language Skills (20개)                               │
│ - python-expert, typescript-expert...               │
│                                                       │
│ Domain Skills (10개)                                 │
│ - web-api-expert, mobile-app-expert...              │
└──────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────┐
│ Community Skills (오픈소스)                          │
│ GitHub: modu-ai/moai-skills-marketplace              │
├──────────────────────────────────────────────────────┤
│ Framework Skills                                     │
│ - django-expert, nextjs-expert, vue-expert...       │
│                                                       │
│ Integration Skills                                   │
│ - aws-expert, kubernetes-expert, terraform-expert   │
│                                                       │
│ Testing Skills                                       │
│ - e2e-testing, performance-testing...               │
└──────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────┐
│ Enterprise Skills (조직 전용)                        │
│ ~/.claude/skills/enterprise/ (조직 내부)            │
├──────────────────────────────────────────────────────┤
│ Company Coding Standards                            │
│ - {company}-code-style, {company}-security-policy   │
│                                                       │
│ Internal Tools                                       │
│ - {company}-deploy-process, {company}-monitoring    │
└──────────────────────────────────────────────────────┘
```

### 6.2 Skills CLI 명령어

```bash
# 1. Skills 검색
moai-adk skills search "react"

→ 출력:
🔍 검색 결과 (3개):

Official Skills:
  react-expert (v1.2.0) ⭐ 4.8/5.0 (234 reviews)
  - React 18+ best practices, hooks, testing
  - Download: 12,345 / Month: 1,234

Community Skills:
  nextjs-expert (v0.9.0) ⭐ 4.5/5.0 (89 reviews)
  - Next.js App Router, SSR, RSC
  - Download: 3,456 / Month: 456

# 2. Skills 설치
moai-adk skills install react-expert

→ 진행 과정:
📦 Downloading react-expert v1.2.0...
✅ Installed to ~/.claude/skills/react-expert/
🔍 Validating SKILL.md...
✅ All checks passed
🎉 react-expert is ready!

# 3. Skills 목록 조회
moai-adk skills list

→ 출력:
📦 Installed Skills (23개):

Foundation (6):
  ✅ moai-spec-writer v0.2.0
  ✅ moai-tdd-orchestrator v0.1.5
  ... (생략)

Language (5):
  ✅ python-expert v1.0.0
  ✅ typescript-expert v0.9.0
  ... (생략)

# 4. Skills 업데이트
moai-adk skills update

→ 출력:
🔄 Checking for updates...

Updates available (3):
  moai-spec-writer: 0.2.0 → 0.3.0
  python-expert: 1.0.0 → 1.1.0
  web-api-expert: 0.3.0 → 0.4.0

Update all? (y/n): y
✅ All skills updated!
```

### 6.3 품질 보증 시스템

#### 1. Skill Certification

```
┌─────────────────────────────────────────┐
│ 🏅 Official MoAI Skill                  │
│ - MoAI 팀이 직접 개발 및 유지보수      │
│ - 품질 보증, 자동 업데이트              │
│ - 예: moai-spec-writer                  │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│ ✅ Verified Community Skill             │
│ - 커뮤니티 기여, MoAI 팀 검증          │
│ - 보안 스캔, 코드 리뷰 통과            │
│ - 예: django-expert                     │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│ 📦 Community Skill                      │
│ - 커뮤니티 기여, 검증 대기 중          │
│ - 사용자 리뷰 참고                      │
└─────────────────────────────────────────┘
```

#### 2. CI/CD 파이프라인

```yaml
# .github/workflows/skill-test.yml

name: Skill Quality Check

on:
  pull_request:
    paths:
      - 'skills/**'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - name: Validate SKILL.md
        run: |
          # YAML frontmatter 검증
          # name, description 필드 확인
          # 문자 수 제한 확인

      - name: Security Scan
        run: |
          # 하드코딩된 시크릿 검사
          # 악성 스크립트 검사

      - name: Integration Test
        run: |
          # Claude와 통합 테스트
          # 트리거 조건 검증
          # Composition 테스트

      - name: Performance Test
        run: |
          # 로딩 시간 측정 (<500ms)
          # 메모리 사용량 확인
```

---

## Part 7: 마이그레이션 전략

### 7.1 4-Phase 마이그레이션

#### Phase 1: Foundation (v0.4.0) - 1개월

**목표**: Skills 인프라 구축 + 핵심 Foundation Skills 15개

```
Week 1-2: 인프라
  ✅ Skills 디렉토리 구조 생성
  ✅ moai-adk skills CLI 명령어
  ✅ 자동 설치 로직
  ✅ SessionStart Hook 업데이트

Week 3-4: Foundation Skills 개발
  ✅ moai-spec-writer
  ✅ moai-spec-id-generator
  ✅ moai-spec-validator
  ✅ moai-tdd-orchestrator
  ✅ moai-tag-scanner
  ✅ moai-tag-validator
  ✅ moai-git-manager
  ✅ moai-branch-creator
  ✅ moai-pr-creator
  ✅ moai-doc-generator
  ✅ moai-api-doc-gen
  ✅ moai-readme-updater
  ✅ moai-project-analyzer
  ✅ moai-lang-detector
  ✅ moai-boilerplate-gen

Week 5: 테스트 및 문서화
  ✅ 통합 테스트
  ✅ 사용 가이드
  ✅ 마이그레이션 가이드
```

**검증 기준**:
- [ ] moai-adk skills install 정상 작동
- [ ] Foundation Skills 15개 정상 동작
- [ ] SessionStart에 Skills 활성화 메시지 표시
- [ ] 문서 완성도 90% 이상

#### Phase 2: Language Skills (v0.5.0) - 1개월

**목표**: 20개 언어별 Skills

```
Week 1: Tier 1 언어 (5개)
  ✅ python-expert
  ✅ typescript-expert
  ✅ javascript-expert
  ✅ java-expert
  ✅ go-expert

Week 2: Tier 2 언어 (5개)
  ✅ rust-expert
  ✅ ruby-expert
  ✅ kotlin-expert
  ✅ swift-expert
  ✅ dart-expert

Week 3-4: Tier 3 언어 (10개)
  ✅ c++-expert, c#-expert, php-expert...
  ✅ 템플릿 기반 자동 생성
```

**검증 기준**:
- [ ] 20개 Language Skills 정상 동작
- [ ] 언어별 TDD 프레임워크 통합
- [ ] 템플릿 기반 자동 생성 검증

#### Phase 3: Domain Skills + Marketplace (v0.6.0) - 1개월

**목표**: 도메인 Skills + 커뮤니티 생태계

```
Week 1-2: Domain Skills 10개
  ✅ web-api-expert
  ✅ mobile-app-expert
  ✅ cli-tool-expert
  ✅ library-expert
  ✅ fullstack-expert
  ✅ auth-patterns
  ✅ database-expert
  ✅ testing-expert
  ✅ devops-expert
  ✅ security-expert

Week 3-4: Marketplace 구축
  ✅ GitHub 저장소 생성 (modu-ai/moai-skills-marketplace)
  ✅ CI/CD 파이프라인
  ✅ 품질 인증 시스템
  ✅ 커뮤니티 기여 가이드
```

**검증 기준**:
- [ ] Domain Skills 10개 정상 동작
- [ ] Marketplace 웹사이트 오픈
- [ ] 첫 번째 Community Skill 인증

#### Phase 4: Advanced Features (v0.7.0) - 진행 중

```
- Skills 자동 조합 최적화
- Skills 추천 엔진
- Skills 사용 통계
- Enterprise Skills 지원
- 다국어 Skills (한/영/일/중)
```

### 7.2 호환성 전략

**기존 기능 유지**:

```
v0.4.0 (Skills 도입 + Commands 명칭 변경)
├── Commands (명칭 변경)
│   ├── /alfred:0-init      (구 0-project)
│   ├── /alfred:1-plan      (구 1-spec) ⭐
│   ├── /alfred:2-build     (유지)
│   └── /alfred:3-sync      (유지)
│
├── Sub-agents (용어 정확화, 역할 축소)
│   ├── spec-builder (복잡한 검증)
│   ├── debug-helper (오류 추론)
│   └── trust-checker (TRUST 검증)
│
├── Skills (신규) ⭐
│   ├── Foundation Skills (6개)
│   ├── Language Skills (20개) [v0.5.0]
│   ├── Domain Skills (10개) [v0.5.0]
│   └── Developer Essentials Skills (4개)
│
└── Hooks (경량화)
    ├── SessionStart (<100ms)
    ├── PreToolUse (<50ms)
    └── PostToolUse
```

**마이그레이션 지원**:
- v0.3.x 명령어: `/alfred:1-spec` → 자동으로 `/alfred:1-plan` 리다이렉트
- 기존 프로젝트: 자동 호환 (Deprecation 경고만 표시)
- v0.6.0: 구 명령어 완전 제거

**사용자 선택**:
- v0.4.0 Commands: `/alfred:1-plan` 사용 (브레인스토밍 모드 지원)
- v0.4.0 Skills: 자연어 대화로 Skills 자동 활용

**점진적 전환**:
- v0.4.0: Commands 명칭 변경 + Skills 10개
- v0.5.0: Language/Domain Skills 추가
- v0.6.0: 구 명령어 제거, Skills 우선
- v1.0.0: Commands는 진입점만, Skills가 핵심

---

## Part 8: 실행 계획

### 8.1 즉시 실행 가능한 액션 플랜

#### Week 1-2: Foundation Skills 3개 (MVP)

**최소 기능 제품 (Proof of Concept)**:

1. **moai-spec-writer** (5일)
   - SKILL.md 작성
   - EARS 템플릿 생성
   - SPEC ID 중복 확인 스크립트
   - 통합 테스트

2. **python-expert** (3일)
   - Python best practice
   - pytest 테스트 가이드
   - mypy 타입 힌트

3. **moai-git-manager** (2일)
   - 브랜치 생성 자동화
   - Draft PR 생성
   - Commit 메시지 자동 생성

4. **통합 및 테스트** (2일)
   - 3개 Skills 자동 조합 테스트
   - 사용자 시나리오 검증
   - 문서화

**검증 시나리오**:

```
사용자: "Python FastAPI 프로젝트의 사용자 인증 SPEC 작성해줘"

예상 결과:
✅ moai-spec-writer가 SPEC 문서 생성
✅ python-expert가 FastAPI 패턴 추가
✅ moai-git-manager가 feature/SPEC-AUTH-001 브랜치 생성
✅ Draft PR 자동 생성
```

### 8.2 성공 지표 (KPI)

#### Phase 1 (v0.4.0)
- [ ] Skills 설치 성공률: 95% 이상
- [ ] moai-spec-writer 사용 만족도: 4.5/5.0 이상
- [ ] 문서 완성도: 90% 이상

#### Phase 2 (v0.5.0)
- [ ] 20개 Language Skills 정상 동작률: 98% 이상
- [ ] 언어별 지원 범위: 20개 언어
- [ ] TDD 프레임워크 통합률: 100%

#### Phase 3 (v0.6.0)
- [ ] 전체 워크플로우 자동화율: 80% 이상
- [ ] Marketplace 첫 번째 Community Skill 인증
- [ ] 커뮤니티 기여자: 10명 이상

#### Phase 4 (v0.7.0)
- [ ] Skills 사용률: 70% 이상 (Commands 대비)
- [ ] 개발 생산성 향상: +150%
- [ ] 사용자 만족도: 4.8/5.0 이상

### 8.3 위험 요소 및 대응 방안

| 위험 요소 | 영향도 | 대응 방안 |
|-----------|--------|-----------|
| Claude Skills API 변경 | 🔴 High | 공식 문서 모니터링, 버전 핀닝 |
| Skills 로딩 성능 저하 | 🟡 Medium | 캐싱, Lazy Loading 구현 |
| 사용자 혼란 (Commands vs Skills) | 🟡 Medium | 명확한 문서화, 튜토리얼 제공 |
| 언어별 템플릿 유지보수 | 🟢 Low | 커뮤니티 기여 유도, CI/CD 자동화 |

---

## 🎬 결론

### 핵심 가치 제안

**MoAI-ADK v0.4.0 "Skills Revolution"**은 다음을 제공합니다:

✅ **Progressive Disclosure**로 무한 확장 가능
✅ **Composability**로 레고처럼 조립 가능
✅ **자연어 UX**로 학습 곡선 제로
✅ **Global Reusability**로 전역 재사용
✅ **Community Ecosystem**으로 지속 성장

### 예상 임팩트

| 측면 | 개선율 |
|------|--------|
| 명령어 학습 | -100% (자연어) |
| 프로젝트 설정 | -90% (자동) |
| SPEC 작성 시간 | -83% |
| TDD 구현 시간 | -62% |
| 문서 동기화 | -83% |
| 재사용성 | +300% |
| **종합 생산성** | **+150%** |

### 다음 단계

1. **즉시 시작**: moai-spec-writer, python-expert, moai-git-manager (MVP)
2. **검증**: 3개 Skills 자동 조합 테스트
3. **확장**: Foundation 15개 → Language 20개 → Domain 10개
4. **생태계**: Marketplace 구축, 커뮤니티 참여

---

**작성 완료일**: 2025-10-19
**다음 리뷰**: Phase 1 완료 후 (예정: 2025-11-19)
**문의**: GitHub Issues (modu-ai/moai-adk)
