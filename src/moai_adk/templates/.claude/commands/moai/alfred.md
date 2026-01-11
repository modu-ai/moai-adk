---
description: "Agentic AI automation - From SPEC to code with autonomous loop"
argument-hint: '"task description" [--loop] [--max N] | resume SPEC-XXX'
type: workflow
allowed-tools: Task, AskUserQuestion, TodoWrite, Skill, Glob, Bash
model: inherit
---

## Pre-execution Context

!git status --porcelain
!git branch --show-current

## Essential Files

@.moai/config/sections/ralph.yaml
@.moai/config/sections/git-strategy.yaml
@.moai/config/sections/quality.yaml
@.moai/config/sections/llm.yaml

---

# /moai:alfred - Agentic AI Autonomous Automation

## Core Principle: 완전 자율 자동화

사용자가 목표를 제시하면 AI가 스스로 계획하고 실행하고 완료합니다.

```
USER: "인증 기능 추가"
  ↓
AI: 탐색 → 계획 → 구현 → 검증 → 반복
  ↓
AI: 모든 이슈 해 solved
  ↓
AI: <promise>DONE</promise>
```

## Command Purpose

전체 MoAI 워크플로우를 자율적으로 실행:

1. **병렬 탐색** (Explore + Research 동시 실행)
2. **SPEC 생성** (사용자 승인 후)
3. **TDD 구현** (자동 반복 수정)
4. **문서 동기화**
5. **완료 마커 감지** (`<promise>DONE</promise>`)

Feature Description: $ARGUMENTS

## Quick Start

```bash
# 기본 자율 실행
/moai:alfred "JWT 인증 추가"

# 자동 루프 활성화 (최대 50회)
/moai:alfred "JWT 인증" --loop --max 50

# 병렬 탐색 + 자동 루프
/moai:alfred "JWT 인증" --loop --parallel

# 이어서 하기
/moai:alfred resume SPEC-AUTH-001
```

## Command Options

| 옵션 | 축약 | 설명 | 기본값 |
|------|------|------|--------|
| `--loop` | - | 자동 반복 수정 활성화 | ralph.yaml |
| `--max N` | --max-iterations | 최대 반복 횟수 | 100 |
| `--parallel` | - | 병렬 탐색 활성화 | 권장 |
| `--branch` | - | 기능 브랜치 자동 생성 | git-strategy |
| `--pr` | - | PR 자동 생성 | git-strategy |
| `--resume SPEC` | - | 이어서 하기 | - |

## Completion Promise (완료 마커)

AI가 작업을 완료했을 때 반드시 마커를 추가:

```markdown
## 완료

모든 구현 완료, 테스트 통과, 문서 업데이트. <promise>DONE</promise>
```

**마커 종류**:
- `<promise>DONE</promise>` - 작업 완료
- `<promise>COMPLETE</promise>` - 전체 완료
- `<ralph:done />` - XML 형식

## Agentic Autonomous Flow

```
START: /moai:alfred "task description"

PHASE 0: 병렬 탐색 (자율)
  ├── Explore Agent: 코드베이스 분석
  ├── Research Agent: 문서/이슈 검색
  └── Quality Agent: 현재 상태 진단
  ↓
통합 → 실행 계획 생성

PHASE 1: SPEC 생성
  └── EARS 형식으로 작성
  ↓
사용자 승인
  ↓
PHASE 2: TDD 구현 (자율 루프)
  │
  └── WHILE (issues_exist AND iteration < max):
       ├── 진단 (LSP + Tests + Coverage)
       ├── TODO 생성
       ├── 수정 실행
       ├── 검증
       └── 완료 마커 감지? → BREAK
  ↓
PHASE 3: 문서 동기화
  └── 마커 추가: <promise>DONE</promise>
  ↓
COMPLETE
```

## TODO-Obsessive Rule (자율 추적)

```
[HARD] TODO 관리 규칙

1. 즉시 생성: 이슈 발견 즉시 TODO 추가
2. 즉시 진행: 시작 전 in_progress 표시
3. 즉시 완료: 완료 후 completed 표시
4. 금지: 배치 완료 (여러 개 한꺼번에 완료 금지)
5. 완료 조건: 모든 TODO 완료 OR 완료 마커
```

## Output Format

### 실행 중

```markdown
## Alfred: Phase 2 (Loop 3/100)

### TODO Status
- [x] JWT 토큰 생성 구현
- [x] 로그인 엔드포인트 구현
- [ ] 토큰 검증 미들웨어 ← 진행 중

### Issues
- ERROR: src/auth.py:45 - undefined 'jwt_decode'
- WARNING: tests/test_auth.py:12 - unused 'result'

수정 중...
```

### 완료 시

```markdown
## Alfred: COMPLETE

### Summary
- SPEC: SPEC-AUTH-001
- Files: 8 files modified
- Tests: 25/25 passing
- Coverage: 88%
- Loops: 7 iterations

### Changes
+ JWT token generation
+ Login endpoint
+ Token validation middleware
+ Unit tests (12 cases)
+ API documentation

<promise>DONE</promise>
```

## LLM Mode

`llm.yaml` 설정에 따라 자동 분기:

| 모드 | Plan Phase | Run Phase |
|------|------------|-----------|
| opus-only | Claude (현재) | Claude (현재) |
| hybrid | Claude (현재) | GLM (worktree) |
| glm-only | GLM (worktree) | GLM (worktree) |

## Expert Delegation (단일 도메인)

단일 도메인 작업은 전문 에이전트에게 직접 위임:

```bash
# Alfred가 자동 판단
/moai:alfred "SQL 쿼리 최적화"

# → expert-performance 에이전트에게 직접 위임
# → SPEC 없이 즉시 구현
```

## Quick Reference

```bash
# 자율 실행 (기본)
/moai:alfred "task"

# 자동 루프 + 병렬
/moai:alfred "task" --loop --parallel

# 최대 반복 지정
/moai:alfred "task" --loop --max 50

# 브랜치 + PR
/moai:alfred "task" --branch --pr

# 이어서 하기
/moai:alfred resume SPEC-XXX
```

---

## EXECUTION DIRECTIVE

1. $ARGUMENTS 파싱
2. LLM 모드 감지 (llm.yaml)
3. 병렬 탐색 실행 (--parallel 또는 ralph.yaml)
4. 라우팅 결정 (전체 워크플로우 vs 전문가 위임)
5. 사용자 확인
6. 실행 및 자율 루프
7. 완료 마커로 종료

---

Version: 3.1.0
Last Updated: 2026-01-11
Core: Agentic AI Autonomous Automation
