---
description: "Cancel autonomous loop with snapshot preservation"
argument-hint: "[--force] [--snapshot] | [--list]"
type: utility
allowed-tools: Task, AskUserQuestion, TodoWrite, Bash, Read, Write, Edit, Glob, Grep
model: inherit
---

## Pre-execution Context

!git status --porcelain
!cat .moai/cache/.moai_loop_state.json 2>/dev/null || echo "No active loop"

## Essential Files

@.moai/config/sections/ralph.yaml

---

# /moai:cancel-loop - Cancel Autonomous Loop

## Core Principle: 상태 보존

루프 취소 시 모든 진행 상태를 스냅샷으로 저장합니다.

```
CANCEL → SNAPSHOT → RECOVER 가능
```

## Command Purpose

활성 루프를 취소하고 복구 가능한 스냅샷을 생성:

1. 루프 중지
2. 스냅샷 저장
3. 상태 보고
4. 복구 옵션 제공

Arguments: $ARGUMENTS

## Quick Start

```bash
# 기본 취소 (확인 후)
/moai:cancel-loop

# 강제 취소
/moai:cancel-loop --force

# 스냅샷 저장
/moai:cancel-loop --snapshot

# 스냅샷 목록
/moai:cancel-loop --list
```

## Command Options

| 옵션 | 설명 |
|------|------|
| `--force` | 확인 없이 취소 |
| `--snapshot` | 스냅샷 저장 |
| `--keep` | state 파일 보존 |
| `--reason TEXT` | 취소 사유 |
| `--list` | 스냅샷 목록 |

## Output Format

### 기본 취소

```markdown
## Loop: Cancelled

### Status
- Iterations: 7/100
- Errors: 2 remaining
- Warnings: 3 remaining

### TODO
1. [ ] src/auth.py:67 - missing return
2. [ ] tests/test_auth.py:12 - unused var

### Snapshot
.moai/cache/ralph-snapshots/cancel-20240111-105230.json

### Recovery
/moai:loop --resume latest
```

### 스냅샷 목록

```markdown
## Snapshots: 2 available

1. cancel-20240111-105230 (7 iters, 2 errors)
2. cancel-20240110-154523 (12 iters, 0 errors)

### Recovery
/moai:loop --resume cancel-20240111-105230
```

### 활성 루프 없음

```markdown
## No Active Loop

### Snapshots Available
2 snapshots found. Use --list to view.

### Start New
/moai:loop
```

## Snapshot Location

```bash
.moai/cache/ralph-snapshots/
├── cancel-20240111-103042.json
├── cancel-20240111-105230.json
└── latest.json -> cancel-20240111-105230.json
```

## Recovery

```bash
# 최신 스냅샷
/moai:loop --resume latest

# 특정 스냅샷
/moai:loop --resume cancel-20240111-105230
```

## Quick Reference

```bash
# 취소
/moai:cancel-loop

# 강제 취소
/moai:cancel-loop --force

# 스냅샷 저장
/moai:cancel-loop --snapshot

# 목록
/moai:cancel-loop --list
```

---

## EXECUTION DIRECTIVE

1. 활성 루프 확인
2. 확인 요청 (--force 제외)
3. 스냅샷 저장 (--snapshot)
4. 상태 보고
5. 복구 옵션

---

Version: 2.1.0
Last Updated: 2026-01-11
Core: Agentic AI State Management
