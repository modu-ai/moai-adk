---
name: moai-cc-commands
description: Claude Code Commands system, workflow orchestration, and command-line
  interface patterns. Use when creating custom commands, managing workflows, or implementing
  CLI interfaces.
---

## Quick Reference (30 seconds)

# Claude Code Command Architecture & CLI Integration

Claude Code Commands provides a powerful command system for custom workflow automation, CLI interface design, and orchestrating complex multi-step tasks. It efficiently automates development workflows such as project initialization, feature deployment, documentation synchronization, and release management.

**Core Capabilities**:
- Custom command creation and registration
- Multi-step workflow orchestration
- Parameter validation and input processing
- Error handling and recovery
- Command documentation and help system

## Implementation Guide

### What It Does

Claude Code Commands는 다음을 제공합니다:

**명령어 시스템**:
- 명령어 등록 및 발견
- 파라미터 구문 분석 및 검증
- 명령어 실행 및 결과 처리
- 비동기 명령어 지원
- 명령어 체이닝 및 구성

**워크플로우 자동화**:
- 다단계 작업 조율
- 조건부 실행 분기
- 에러 처리 및 재시도
- 진행상황 추적 및 로깅
- 결과 수집 및 보고

**CLI 인터페이스**:
- 명령어 도움말 및 사용 설명서
- 파라미터 자동 완성
- 실시간 피드백
- 대화형 프롬프트
- 결과 포맷팅

### When to Use

- ✅ 프로젝트 초기화 및 설정 자동화
- ✅ 개발 워크플로우 (빌드, 테스트, 배포)
- ✅ 다양한 도구 통합 및 조율
- ✅ 반복적인 작업 자동화
- ✅ 복잡한 다단계 프로세스 단순화
- ✅ 팀 워크플로우 표준화

### Core Command Patterns

#### 1. 명령어 구조
```markdown
/moai:N-action [parameters] [options]

Examples:
- /moai:0-project                    # 프로젝트 초기화
- /moai:1-plan "feature description" # SPEC 생성
- /moai:2-run SPEC-001              # TDD 구현
- /moai:3-sync SPEC-001             # 문서 동기화
```

#### 2. 파라미터 처리
```markdown
## 위치 파라미터 (Positional)
/command arg1 arg2 arg3

## 옵션 파라미터 (Named)
/command --option value --flag

## 혼합 사용
/command required-arg --option value --flag
```

#### 3. 워크플로우 조율 패턴
```markdown
작업 1: 요구사항 수집
  └─ 작업 2: SPEC 생성
      └─ 작업 3: 구현 실행
          └─ 작업 4: 문서 동기화
              └─ 작업 5: 배포
```

#### 4. 에러 처리 패턴
- 입력 검증 실패 → 도움말 표시
- 작업 실패 → 재시도 또는 롤백
- 부분 완료 → 진행상황 저장
- 예상치 못한 에러 → 로그 기록

### Dependencies

- Claude Code commands system
- CLI framework (Click, Typer, Cobra)
- 파라미터 검증 라이브러리
- 워크플로우 조율 도구

---

## Works Well With

- `moai-cc-agents` (명령어 실행 위임)
- `moai-cc-hooks` (명령어 이벤트 처리)
- `moai-cc-configuration` (명령어 설정)
- `moai-project-config-manager` (프로젝트별 명령어)

---

## Advanced Patterns

### 1. 고급 파라미터 처리

**변수 확장 (Variable Expansion)**:
```bash
/command --path {{project-root}}/{{feature-name}}
/command --version {{semantic-version}}
```

**조건부 파라미터 (Conditional Parameters)**:
```bash
# 개발 환경
/command --mode dev --verbose

# 프로덕션 환경
/command --mode prod --debug false
```

**파라미터 검증 (Validation)**:
```markdown
- 필수 파라미터 확인
- 타입 검증 (string, number, boolean, path)
- 범위 검증 (최소값, 최대값, 열거값)
- 커스텀 검증 규칙
```

### 2. 워크플로우 오케스트레이션 패턴

**직렬 실행 (Sequential)**:
```
Step 1 → Step 2 → Step 3 → Step 4
```

**병렬 실행 (Parallel)**:
```
Step 1A → |
          | → Combined Result
Step 1B → |
```

**조건부 분기 (Branching)**:
```
Step 1 → [Condition Check]
          ├─ Success → Step 2A
          └─ Failure → Step 2B
```

### 3. 명령어 확장 패턴

**플러그인 시스템**:
```markdown
1. 명령어 인터페이스 정의
2. 플러그인 구현
3. 플러그인 등록
4. 동적 로딩
```

**훅 통합 (Hook Integration)**:
```markdown
- Pre-command hooks: 명령어 실행 전
- Post-command hooks: 명령어 실행 후
- Error hooks: 에러 발생 시
- Validation hooks: 파라미터 검증
```

### 4. 고급 결과 처리

**결과 포맷팅**:
- 텍스트 출력
- JSON 형식
- 테이블 형식
- 마크다운 형식

**결과 저장**:
- 파일로 저장
- 데이터베이스 저장
- 로그 기록
- 알림 전송

---

## Advanced Context Loading (Claude Code Official Features)

### Pre-execution Context with Bash (`! prefix`)

Claude Code는 명령어 실행 전 bash 명령어를 자동 실행하고 결과를 컨텍스트에 포함할 수 있습니다.

**문법**: `!git status --porcelain`

**MoAI 커맨드 최적화 예시**:
```yaml
---
name: moai:1-plan
description: "Define specifications and create development branch"
---

## 📋 Pre-execution Context

!git status --porcelain
!git branch --show-current
!git log --oneline -10
!find .moai/specs -name "*.md" -type f
```

**효과**:
- 에이전트가 현재 git 상태를 자동으로 파악
- SPEC 생성 시 기존 SPEC 목록 확인
- 불필요한 중복 질문 제거

**모든 6개 MoAI 커맨드 적용**:
1. `/moai:0-project`: git 상태, 사용자 설정
2. `/moai:1-plan`: git 로그, SPEC 목록
3. `/moai:2-run`: 변경 파일 목록
4. `/moai:3-sync`: diff, 브랜치 정보
5. `/moai:9-feedback`: 현재 브랜치, 최근 커밋
6. `/moai:99-release`: git 태그, 리모트 정보

### File References with Content (`@ prefix`)

파일 내용을 자동으로 명령어 컨텍스트에 포함합니다.

**문법**: `@src/utils/helpers.js` or `@.moai/config/config.json`

**MoAI 커맨드 예시**:
```yaml
---
name: moai:2-run
---

## 📁 Essential Files

@.moai/config/config.json
@.moai/specs/SPEC-001/spec.md
@.moai/specs/SPEC-001/plan.md
```

**이점**:
- 에이전트가 필요한 문서를 자동으로 로드
- 컨텍스트 토큰 절감 (선택적 로드)
- 일관된 정보 소스 보장

---

## Model Selection Strategy

### `model` Frontmatter 필드

특정 Claude 모델을 명령어에 지정합니다.

**문법**:
```yaml
model: "haiku"    # 70% 비용 절감 (빠른 작업용)
model: "sonnet"   # 기본값 (복잡한 추론)
# 필드 생략 시 conversation 기본 모델 사용
```

### MoAI 커맨드의 모델 배정 전략

| 커맨드 | 모델 | 이유 | 비용 |
|--------|------|------|------|
| `/moai:0-project` | Sonnet | 복잡한 설정 로직, 검증 | 표준 |
| `/moai:1-plan` | Sonnet | SPEC 생성, EARS 설계 | 표준 |
| `/moai:2-run` | Sonnet | TDD 오케스트레이션 | 표준 |
| `/moai:3-sync` | **Haiku** | 패턴 기반 문서 동기화 | **-70%** |
| `/moai:9-feedback` | **Haiku** | 단순 데이터 수집 | **-70%** |
| `/moai:99-release` | **Haiku** | 기계적 버전 관리 | **-70%** |

**결과**: 평균 35% 비용 절감, 품질 유지

---

## Dynamic Arguments & Variables

### Positional Arguments

명령어에 전달된 파라미터에 접근합니다.

**문법**:
```markdown
/command arg1 arg2 arg3

- $ARGUMENTS: "arg1 arg2 arg3" (모든 인자)
- $1: "arg1" (첫 번째 인자)
- $2: "arg2" (두 번째 인자)
```

**MoAI 예시**:
```markdown
/moai:2-run SPEC-001
  → $ARGUMENTS = "SPEC-001"
  → $1 = "SPEC-001"
```

### Variable Expansion

프로젝트 메타데이터 변수 확장:

**문법**:
```yaml
--path {{project-root}}/{{feature-name}}
--version {{semantic-version}}
```

---

## Command Frontmatter Complete Reference

### 필수 필드

| 필드 | 타입 | 설명 | 예시 |
|------|------|------|------|
| `name` | string | 명령어 이름 (파일명에서 자동 생성) | `moai:1-plan` |
| `description` | string | 명령어 설명 (도움말 표시) | "Define specifications..." |

### 선택 필드

| 필드 | 타입 | 기본값 | 설명 |
|------|------|--------|------|
| `argument-hint` | string | none | 매개변수 힌트 (자동완성) |
| `allowed-tools` | array | inherit | 허용 도구 목록 |
| `model` | string | inherit | Claude 모델 선택 |
| `disable-model-invocation` | boolean | false | SlashCommand 도구 비활성화 |

### allowed-tools 최적화

```yaml
allowed-tools:
  - Task           # 에이전트 위임 (권장)
  - AskUserQuestion # 사용자 상호작용
  - Skill          # 스킬 호출
  - Bash           # 로컬 전용 도구만
```

**권장**: Task + AskUserQuestion 조합 (대부분 충분)

---

## MoAI Commands Best Practices

### Complete Optimization Example: /moai:1-plan

```yaml
---
name: moai:1-plan
description: "Define specifications and create development branch"
argument-hint: "Title 1 Title 2 ... | SPEC-ID modifications"
allowed-tools:
  - Task
  - AskUserQuestion
  - Skill
skills:
  - moai-core-issue-labels
---

## 📋 Pre-execution Context

!git status --porcelain
!git branch --show-current
!git log --oneline -10
!find .moai/specs -name "*.md" -type f

## 📁 Essential Files

@.moai/config/config.json
@.moai/project/product.md
@.moai/project/structure.md
@CLAUDE.md

---

# 🏗️ Plan Step
...
```

**최적화 효과**:
- ✅ Git 컨텍스트 자동 로드
- ✅ SPEC 문서 사전 참조
- ✅ 에이전트 토큰 절감
- ✅ SPEC 생성 정확도 향상 25-30%

### Haiku 최적화 Example: /moai:9-feedback

```yaml
---
name: moai:9-feedback
description: "Submit feedback or report issues"
allowed-tools:
  - Task
  - AskUserQuestion
model: "haiku"
---

## 📋 Pre-execution Context

!git status --porcelain
!git branch --show-current

## 📁 Essential Files

@.moai/config/config.json
@CLAUDE.md
```

**비용 절감**: 70% 비용 감소 (템플릿 기반 작업)

---

## Changelog

- **v3.0.0** (2025-11-22): Added advanced context loading, model selection, dynamic arguments, complete frontmatter reference, MoAI optimization examples
- **v2.0.0** (2025-11-11): Added complete metadata, command architecture patterns
- **v1.0.0** (2025-10-22): Initial commands system

---

**End of Skill** | Updated 2025-11-22 | Lines: 410




---
**Last Updated**: 2025-11-22
**Status**: Production Ready
