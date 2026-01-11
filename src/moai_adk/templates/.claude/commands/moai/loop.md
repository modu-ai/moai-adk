---
description: "Agentic autonomous loop - Auto-fix until completion marker"
argument-hint: "[--max N] [--auto] [--parallel] | --resume snapshot"
type: utility
allowed-tools: Task, AskUserQuestion, TodoWrite, Bash, Read, Write, Edit
model: inherit
---

## Pre-execution Context

!git status --porcelain
!git diff --name-only HEAD

## Essential Files

@.moai/config/sections/ralph.yaml

---

# /moai:loop - Agentic Autonomous Loop

## Core Principle: 완전 자율 반복 수정

AI가 스스로 이슈를 찾고 수정하고 완료할 때까지 반복합니다.

```
START: 이슈 감지
  ↓
AI: 수정 → 검증 → 반복
  ↓
AI: 완료 마커 추가
  ↓
<promise>DONE</promise>
```

## Command Purpose

LSP 오류, 테스트 실패, 커버리지를 AI가 자율적으로 수정:

1. **병렬 진단** (LSP + AST-grep + Tests 동시 실행)
2. **TODO 자동 생성**
3. **자율 수정** (Level 1-3)
4. **반복 검증**
5. **완료 마커 감지**

Arguments: $ARGUMENTS

## Quick Start

```bash
# 기본 자율 루프
/moai:loop

# 최대 50회 반복
/moai:loop --max 50

# 병렬 진단 + 자동 수정
/moai:loop --parallel --auto

# 스냅샷 복구
/moai:loop --resume latest
```

## Command Options

| 옵션 | 축약 | 설명 | 기본값 |
|------|------|------|--------|
| `--max N` | --max-iterations | 최대 반복 횟수 | 100 |
| `--auto` | --auto-fix | 자동 수정 활성화 | Level 1 |
| `--parallel` | - | 병렬 진단 실행 | 권장 |
| `--errors` | --errors-only | 에러만 수정 | 전체 |
| `--coverage` | --include-coverage | 커버리지 포함 | 85% |
| `--resume ID` | --resume-from | 스냅샷 복구 | - |

## Completion Promise (완료 마커)

AI가 모든 작업을 완료하면 마커를 추가:

```markdown
## 루프 완료

7번의 반복으로 5개 에러, 3개 경고 해결. <promise>DONE</promise>
```

**마커 종류**:
- `<promise>DONE</promise>` - 작업 완료
- `<promise>COMPLETE</promise>` - 전체 완료
- `<ralph:done />` - XML 형식

마커 없으면 계속 반복합니다.

## Autonomous Loop Flow

```
START: /moai:loop

PARALLEL 진단 (--parallel)
  ├── LSP: 에러/경고
  ├── AST-grep: 보안
  ├── Tests: 테스트
  └── Coverage: 커버리지
  ↓
통합 결과
  ↓
완료 마커 감지?
  ├── YES → COMPLETE
  ↓
조건 충족?
  ├── YES → "마커 추가 또는 계속?"
  ↓
TODO 생성 (즉시)
  ↓
수정 실행 (자율)
  ├── Level 1: 즉시 수정 (import, formatting)
  ├── Level 2: 안전 수정 (rename, type)
  └── Level 3: 승인 필요 (logic, api)
  ↓
검증
  ↓
max 도달? → STOP
  ↓
반복
```

## Parallel Diagnostics (병렬 진단)

```bash
# --parallel 없이 (순차)
LSP → AST-grep → Tests → Coverage
총 30초

# --parallel 사용 (병렬)
LSP ├─┐
     ├─→ 통합 → 8초 (3.75배 빠름)
AST ├─┤
    ├─┘
Tests ┤
       └─→ 3-4배 속도 향상
Coverage
```

## TODO-Obsessive Rule (자율 추적)

```
[HARD] TODO 관리 규칙

반복마다:
1. 즉시 생성: 이슈 → TODO
2. 즉시 진행: [ ] → [in_progress]
3. 즉시 완료: [in_progress] → [x]
4. 모두 완료: All [x] → 완료 마커
```

### TODO 예시

```markdown
## Loop 3/100

### Status
- Errors: 2
- Warnings: 5

### TODO
1. [x] src/auth.py:45 - undefined 'jwt_token'
2. [in_progress] src/auth.py:67 - missing return
3. [ ] tests/test_auth.py:12 - unused 'result'
```

## Auto-Fix Levels

| Level | 설명 | 승인 | 예시 |
|-------|------|------|------|
| 1 | 즉시 수정 | 불필요 | import 정렬, 공백 |
| 2 | 안전 수정 | 로그만 | 변수改名, 타입 추가 |
| 3 | 승인 필요 | 필요 | 로직 변경, API 수정 |
| 4 | 수동 필요 | 불가능 | 보안, 아키텍처 |

## Output Format

### 실행 중

```markdown
## Loop: 3/100 (parallel)

### Diagnostics (0.8s)
- LSP: 2 errors, 5 warnings
- AST-grep: 0 security issues
- Tests: 23/25 passing
- Coverage: 82%

### TODO
1. [x] src/auth.py:45 - undefined 'jwt_token'
2. [in_progress] src/auth.py:67 - missing return
3. [ ] tests/test_auth.py:12 - unused 'result'

수정 중...
```

### 완료 (마커 감지)

```markdown
## Loop: COMPLETE

### Summary
- Iterations: 7
- Errors fixed: 5
- Warnings fixed: 3
- Tests: 25/25 passing
- Coverage: 87%

### Files Modified
- src/auth.py (7 fixes)
- tests/test_auth.py (3 fixes)
- src/api/routes.py (2 fixes)

<promise>DONE</promise>
```

### 최대 반복 도달

```markdown
## Loop: MAX REACHED (100/100)

### Remaining
- Errors: 1
- Warnings: 2

### Options
1. /moai:loop --max 200  # 계속
2. /moai:fix --parallel  # 한 번만
3. 수동 수정
```

## State & Snapshot

```bash
# 상태 저장
.moai/cache/.moai_loop_state.json

# 스냅샷
.moai/cache/ralph-snapshots/
├── iteration-001.json
├── iteration-002.json
└── latest.json

# 복구
/moai:loop --resume iteration-002
/moai:loop --resume latest
```

## Cancellation

```bash
# 취소 (스냅샷 저장)
/moai:cancel-loop --snapshot

# 강제 취소
/moai:cancel-loop --force
```

## Quick Reference

```bash
# 자율 루프
/moai:loop

# 병렬 + 자동
/moai:loop --parallel --auto

# 최대 반복
/moai:loop --max 50

# 에러만
/moai:loop --errors

# 복구
/moai:loop --resume latest
```

---

## EXECUTION DIRECTIVE

1. $ARGUMENTS 파싱
2. 병렬 진단 실행 (--parallel)
3. 완료 마커 감지 확인
4. TODO 생성 (즉시)
5. 수정 실행 (--auto 레벨)
6. 검증
7. 반복 (max 미만)

---

Version: 2.1.0
Last Updated: 2026-01-11
Core: Agentic AI Autonomous Loop
