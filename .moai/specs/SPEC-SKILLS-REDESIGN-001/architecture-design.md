# Commands → Sub-agents → Skills 아키텍처 설계 (공식 문서 기반)

**작성일**: 2025-10-20
**작성자**: @Alfred
**기준 문서**:
- https://docs.claude.com/en/docs/claude-code/skills
- https://www.anthropic.com/engineering/equipping-agents-for-the-real-world-with-agent-skills
- https://docs.claude.com/en/docs/claude-code/sub-agents

---

## ❌ 현재 구현 상태 (문제점)

### 1. Commands → Sub-agents 연결: ✅ 정상

**확인 결과**:
```bash
# Commands는 Task tool로 Sub-agents 호출
rg "subagent_type=" .claude/commands/alfred/*.md -c

# 결과: 모든 Commands가 Task tool 사용 확인됨
```

**예시**:
```markdown
# /alfred:1-plan 커맨드
Task(
    subagent_type="spec-builder",
    description="SPEC 문서 작성",
    prompt="..."
)
```

✅ **결론**: Commands는 공식 스펙대로 Task tool을 사용하여 Sub-agents를 명시적으로 호출합니다.

---

### 2. Sub-agents YAML Frontmatter: ✅ 정상

**확인 결과**:
```yaml
---
name: spec-builder
description: "Use when: EARS 방식의 SPEC 문서 작성이 필요할 때"
tools: Read, Write, Edit, MultiEdit, Bash, Glob, Grep, TodoWrite, WebFetch
model: sonnet
# ✅ skills: 필드 없음 (올바름)
---
```

**공식 스펙**:
- Sub-agents YAML frontmatter 필드: `name`, `description`, `tools`, `model`
- **`skills:` 필드는 공식 스펙이 아님**

✅ **결론**: Sub-agents의 YAML frontmatter는 공식 스펙을 정확히 따릅니다.

---

### 3. Sub-agents → Skills 연결: ❌ 문제 발견

**현재 상태**:
- Sub-agents의 **시스템 프롬프트 텍스트**에서 Skills를 **전혀 언급하지 않음**
- 예시:
  - `tdd-implementer.md`: Python TDD 구현 시 `moai-lang-python` 스킬 언급 없음
  - `implementation-planner.md`: SPEC 검증 시 `moai-foundation-specs` 스킬 언급 없음
  - `doc-syncer.md`: 문서 동기화 시 관련 스킬 언급 없음

**공식 스펙**:
```markdown
# ✅ 올바른 방법 (공식 스펙)
Sub-agents의 **시스템 프롬프트 텍스트**에서 Skills를 언급:

"Python 프로젝트에서 TDD 구현 시, **moai-lang-python 스킬**을 참고하세요.
이 스킬은 pytest, mypy, ruff, black 사용법을 제공합니다."
```

**Skills는 model-invoked**이므로:
- Claude가 description을 읽고 자동으로 필요 여부 판단
- Sub-agents가 **텍스트로 언급**하면 Claude가 해당 Skill을 로드

❌ **문제**: Sub-agents 시스템 프롬프트에 Skills 언급이 없어, Claude가 적절한 Skill을 사용하지 못함

---

### 4. Skills 구조: ✅ 정상

**확인 결과**:
```yaml
---
name: moai-lang-python
description: Python best practices with pytest, mypy, ruff, black, and uv package management
tier: 2  # MoAI-ADK 커스텀 필드 (분류용)
---
```

**공식 스펙**:
- 필수 필드: `name`, `description`
- `tier`는 MoAI-ADK 커스텀 필드 (내부 분류용, 공식 스펙 아님)

✅ **결론**: Skills YAML frontmatter는 공식 스펙을 따르며, 추가로 tier 필드를 사용합니다.

---

## ✅ 올바른 아키텍처 설계

### 공식 3-Layer 구조

```
┌─────────────────────────────────────────┐
│ Layer 1: Commands                       │
│ - 사용자 명시 호출 (/command-name)      │
│ - Task tool로 Sub-agents 호출 (명시적) │
└──────────────┬──────────────────────────┘
               │
               │ Task(subagent_type="name")
               │
               ↓
┌─────────────────────────────────────────┐
│ Layer 2: Sub-agents                     │
│ - 독립 컨텍스트 (메인과 분리)           │
│ - 시스템 프롬프트에서 Skills 언급 ✅    │
└──────────────┬──────────────────────────┘
               │
               │ "moai-lang-python 스킬을 참고하세요"
               │ (텍스트 언급, YAML 필드 아님)
               │
               ↓
┌─────────────────────────────────────────┐
│ Layer 3: Skills                         │
│ - Model-invoked (Claude 자동 판단)      │
│ - 시작 시 description만 pre-load        │
│ - 필요 시 전체 내용 JIT load            │
└─────────────────────────────────────────┘
```

---

### Sub-agents → Skills 연결 방법 (공식 스펙)

#### ❌ 잘못된 방법 (공식 스펙 아님)

```yaml
---
name: spec-builder
skills:  # ← 이것은 공식 스펙이 아님!
  - moai-foundation-specs
  - moai-foundation-ears
---
```

#### ✅ 올바른 방법 (공식 스펙)

**Sub-agents의 시스템 프롬프트 텍스트**에서 Skills를 언급:

```markdown
# TDD Implementer - TDD 실행 전문가

## 🔗 관련 스킬 (Skills)

**언어별 TDD 구현**:
- **Python 프로젝트**: `moai-lang-python` 스킬을 참고하세요 (pytest, mypy, ruff, black)
- **TypeScript 프로젝트**: `moai-lang-typescript` 스킬을 참고하세요 (Vitest, Biome)
- **Java 프로젝트**: `moai-lang-java` 스킬을 참고하세요 (JUnit, Maven/Gradle)
- **Go 프로젝트**: `moai-lang-go` 스킬을 참고하세요 (go test, gofmt)

**코드 품질 검증**:
- **리팩토링**: `moai-essentials-refactor` 스킬을 참고하세요
- **코드 리뷰**: `moai-essentials-review` 스킬을 참고하세요

Claude는 프로젝트 언어를 감지하여 적절한 스킬을 자동으로 사용합니다.
```

**핵심 원리**:
1. Sub-agents는 **텍스트로 관련 Skills를 언급**
2. Claude가 description을 읽고 **필요 시 자동 로드** (model-invoked)
3. YAML frontmatter에 `skills:` 필드는 **사용하지 않음**

---

## 📋 수정 계획

### 1단계: Sub-agents 시스템 프롬프트 수정

각 Sub-agent에 "관련 스킬" 섹션 추가:

| Sub-agent | 관련 Skills |
|-----------|------------|
| **spec-builder** | moai-foundation-specs, moai-foundation-ears, moai-lang-* |
| **tdd-implementer** | moai-lang-*, moai-essentials-refactor, moai-essentials-review |
| **implementation-planner** | moai-foundation-specs, moai-lang-* |
| **doc-syncer** | moai-foundation-specs, moai-essentials-review |
| **debug-helper** | moai-essentials-debug, moai-lang-* |
| **trust-checker** | moai-foundation-trust, moai-lang-* |

### 2단계: Skills 설명 최적화

Claude의 auto-discovery를 위해 Skills의 `description` 필드 최적화:

**현재**:
```yaml
description: Python best practices with pytest, mypy, ruff, black, and uv package management
```

**개선 (더 구체적)**:
```yaml
description: "Use when: Python TDD 구현 시. pytest 테스트, mypy 타입 체크, ruff 린트, black 포맷, uv 패키지 관리 가이드 제공"
```

**개선 포인트**:
- "Use when:" 패턴 추가 (호출 시점 명확화)
- 구체적인 도구 나열
- 한국어 설명 추가 (프로젝트 기본 언어)

### 3단계: 검증

**검증 방법**:
1. Sub-agents 파일 확인: 관련 스킬 섹션 존재 확인
2. Skills description 확인: "Use when:" 패턴 포함 확인
3. 통합 테스트: Commands → Sub-agents → Skills 호출 흐름 검증

---

## 🔗 통합 예시

### 예시 1: /alfred:2-run SPEC-AUTH-001 실행

**호출 흐름**:

```
1. /alfred:2-run 커맨드 실행
   ↓
2. implementation-planner Sub-agent 호출 (Task tool)
   → 시스템 프롬프트 읽기: "moai-foundation-specs 스킬을 참고하세요"
   → Claude가 moai-foundation-specs description 확인
   → "Use when: SPEC 검증 필요 시" → 로드 결정
   ↓
3. implementation-planner가 SPEC 분석
   → moai-foundation-specs 스킬의 EARS 검증 가이드 사용
   ↓
4. tdd-implementer Sub-agent 호출
   → 시스템 프롬프트 읽기: "moai-lang-python 스킬을 참고하세요"
   → Claude가 프로젝트 언어 감지 (Python)
   → moai-lang-python 스킬 로드
   ↓
5. tdd-implementer가 TDD 구현
   → moai-lang-python 스킬의 pytest, mypy 가이드 사용
```

### 예시 2: Sub-agent 시스템 프롬프트 템플릿

```markdown
---
name: tdd-implementer
description: "Use when: TDD RED-GREEN-REFACTOR 구현이 필요할 때"
tools: Read, Write, Edit, MultiEdit, Bash, Grep, Glob, TodoWrite
model: sonnet
---

# TDD Implementer - TDD 실행 전문가

## 🎯 핵심 역할
[기존 내용]

## 🔗 관련 스킬 (Skills)

**언어별 TDD 가이드**:
TDD 구현 시 프로젝트 언어에 맞는 스킬을 참고하세요:

- **Python**: `moai-lang-python` - pytest, mypy, ruff, black 사용법
- **TypeScript**: `moai-lang-typescript` - Vitest, Biome 사용법
- **Java**: `moai-lang-java` - JUnit, Maven/Gradle 사용법
- **Go**: `moai-lang-go` - go test, gofmt 사용법
- **Rust**: `moai-lang-rust` - cargo test, clippy, rustfmt 사용법

**코드 품질 스킬**:
- **리팩토링**: `moai-essentials-refactor` - 디자인 패턴, 코드 개선 전략
- **코드 리뷰**: `moai-essentials-review` - SOLID 원칙, 코드 스멜 감지
- **디버깅**: `moai-essentials-debug` - 스택 트레이스 분석, 오류 패턴 감지

Claude는 프로젝트 환경을 자동 감지하여 적절한 스킬을 로드합니다.

## 📋 워크플로우 단계
[기존 내용]
```

---

## 📊 수정 범위 요약

### 수정 대상 파일

**Sub-agents (19개)**:
```bash
.claude/agents/alfred/spec-builder.md
.claude/agents/alfred/tdd-implementer.md
.claude/agents/alfred/implementation-planner.md
.claude/agents/alfred/doc-syncer.md
.claude/agents/alfred/debug-helper.md
.claude/agents/alfred/trust-checker.md
.claude/agents/alfred/tag-agent.md
.claude/agents/alfred/git-manager.md
.claude/agents/alfred/cc-manager.md
.claude/agents/alfred/project-manager.md
.claude/agents/alfred/quality-gate.md
.claude/agents/alfred/backup-merger.md
.claude/agents/alfred/language-detector.md
.claude/agents/alfred/feature-selector.md
.claude/agents/alfred/document-generator.md
.claude/agents/alfred/project-interviewer.md
.claude/agents/alfred/template-optimizer.md
```

**Skills (46개)**: description 최적화 (선택적)

### 수정 내용

**필수**:
- ✅ Sub-agents에 "관련 스킬" 섹션 추가

**선택**:
- ⚠️ Skills description에 "Use when:" 패턴 추가

---

## 🎯 다음 단계

1. **Sub-agents 템플릿 작성**: 19개 Sub-agents의 "관련 스킬" 섹션 템플릿 생성
2. **일괄 수정**: 모든 Sub-agents에 관련 스킬 섹션 추가
3. **검증**: Commands → Sub-agents → Skills 호출 흐름 테스트
4. **문서화**: CLAUDE.md 및 development-guide.md에 공식 아키텍처 반영

---

**작성 완료**: 2025-10-20
**다음 작업**: Sub-agents 수정 계획 승인 대기
