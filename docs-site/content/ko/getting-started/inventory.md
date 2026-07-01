---
title: moai inventory 커맨드
weight: 25
draft: false
---

프로젝트의 활성 세션, 워크트리, 하네스를 조회하는 `moai inventory` 커맨드를 안내합니다.

{{< callout type="info" >}}
**한 줄 요약**: `moai inventory`는 현재 프로젝트의 모든 활성 자원(세션, 워크트리, 하네스)을 한눈에 조회합니다.
{{< /callout >}}

## 개요

`moai inventory`는 읽기 전용 명령어로, 현재 프로젝트 상태의 **통합 인벤토리**를 제공합니다.

### 조회 대상

| 자원 | 설명 | 위치 |
|------|------|------|
| **Active Sessions** | 현재 실행 중인 Claude Code 세션 | `.moai/state/active-sessions.json` |
| **Worktrees** | 프로젝트용 L2/L3 격리 브랜치 | `~/.moai/worktrees/<project>/` |
| **Harnesses** | 생성된 동적 에이전트 팀 | `.moai/harness/manifest.json` |
| **SPEC Progress** | 활성 SPEC의 진행 상태 | `.moai/specs/SPEC-*/progress.md` |

## 명령어 형식

```bash
moai inventory [options]
```

### 기본 사용

```bash
moai inventory
```

기본 텍스트 형식으로 인벤토리 출력합니다.

### JSON 형식 출력

```bash
moai inventory --json
```

구조화된 JSON으로 출력하여 자동 분석에 활용할 수 있습니다.

### 필터링

특정 자원 타입만 조회:

```bash
moai inventory --type sessions
moai inventory --type worktrees
moai inventory --type harnesses
moai inventory --type specs
```

### 상세 정보

각 자원의 추가 정보 포함:

```bash
moai inventory --verbose
moai inventory --verbose --json
```

## 텍스트 형식 출력

### 기본 출력 예시

```
MOAI Inventory for moai-adk-go
Project Root: /Users/goos/MoAI/moai-adk-go
Updated: 2026-07-01T10:15:00Z

========== ACTIVE SESSIONS ==========
Session ID                              Branch        SPEC ID            Status
edc25996-04cb-4139-b2f6-c2968e7337db    main          SPEC-DOCS-001      in-progress
a1b2c3d4-e5f6-7890-1234-567890abcdef    feat/auth     SPEC-AUTH-002      run-phase

========== WORKTREES ==========
Name                    Branch              Created        Status
SPEC-DOCS-001          docs/rebuild        2026-07-01     active
SPEC-AUTH-002          feat/auth            2026-07-01     active

========== HARNESSES ==========
Name                    Version    Teammates    Worktree Isolation    Status
backend-team            1.0.0      3            L1_optional           active
frontend-team           1.0.0      2            none                  active

========== ACTIVE SPECS ==========
SPEC ID                 Status          Phase      Owner           Progress
SPEC-DOCS-001          in-progress     run        manager-develop  M3/6
SPEC-AUTH-002          in-progress     run        manager-develop  M2/5
```

### 상세 정보 (`--verbose`)

```
========== ACTIVE SESSIONS (VERBOSE) ==========

Session: edc25996-04cb-4139-b2f6-c2968e7337db
  Created:     2026-06-29T14:30:00Z
  Last Update: 2026-07-01T10:15:00Z
  Branch:      main
  SPEC ID:     SPEC-DOCS-001
  Status:      in-progress (running M3)
  Context:     ~145K / 200K tokens (73%)
  Model:       claude-haiku-4-5
  Resume:      available (.moai/specs/SPEC-DOCS-001/progress.md)

========== WORKTREES (VERBOSE) ==========

Worktree: SPEC-DOCS-001
  Path:         ~/.moai/worktrees/moai-adk-go/SPEC-DOCS-001
  Base Branch:  main (origin/main)
  Created:      2026-07-01T08:00:00Z
  Session:      edc25996-04cb-4139-b2f6-c2968e7337db
  Files Modified: 7
  Files Created:  4
  Commits:       2
```

## JSON 형식 출력

### 스키마

```json
{
  "inventory": {
    "project_root": "/Users/goos/MoAI/moai-adk-go",
    "timestamp": "2026-07-01T10:15:00Z",
    "sessions": [...],
    "worktrees": [...],
    "harnesses": [...],
    "specs": [...]
  }
}
```

### Session 객체

```json
{
  "session_id": "edc25996-04cb-4139-b2f6-c2968e7337db",
  "created_at": "2026-06-29T14:30:00Z",
  "branch": "main",
  "spec_id": "SPEC-DOCS-001",
  "status": "in-progress",
  "context_usage": {
    "current": 145000,
    "total": 200000,
    "percentage": 72.5
  },
  "model": "claude-haiku-4-5",
  "resume_available": true
}
```

### Worktree 객체

```json
{
  "name": "SPEC-DOCS-001",
  "path": "~/.moai/worktrees/moai-adk-go/SPEC-DOCS-001",
  "base_branch": "main",
  "created_at": "2026-07-01T08:00:00Z",
  "session_id": "edc25996-04cb-4139-b2f6-c2968e7337db",
  "status": "active",
  "files_modified": 7,
  "files_created": 4,
  "commits": 2
}
```

### Harness 객체

```json
{
  "name": "backend-team",
  "version": "1.0.0",
  "created_at": "2026-07-01T10:00:00Z",
  "teammates": 3,
  "worktree_isolation": "L1_optional",
  "status": "active",
  "manifest_path": ".moai/harness/manifest.json"
}
```

### SPEC 객체

```json
{
  "spec_id": "SPEC-DOCS-001",
  "title": "Documentation v3 Rebuild",
  "status": "in-progress",
  "phase": "run",
  "current_milestone": 3,
  "total_milestones": 6,
  "owner": "manager-develop",
  "progress_file": ".moai/specs/SPEC-DOCS-001/progress.md",
  "created_at": "2026-06-20T09:00:00Z"
}
```

## 실용적인 사용 예시

### 1. 다중 세션 경합 감지

```bash
moai inventory --type sessions

# 출력에서 같은 SPEC을 다루는 세션 > 1개 감지 → 경합 위험
```

### 2. Worktree 정리 확인

```bash
moai inventory --type worktrees --verbose

# 오래된 worktree 확인 후 정리
moai worktree remove <name>
```

### 3. Harness 팀 목록 조회

```bash
moai inventory --type harnesses --json | jq '.inventory.harnesses[] | {name, teammates, status}'

# 예상 출력:
# {
#   "name": "backend-team",
#   "teammates": 3,
#   "status": "active"
# }
```

### 4. 활성 SPEC 진행률 추적

```bash
moai inventory --type specs | grep in-progress

# 현재 진행 중인 모든 SPEC 확인
```

### 5. 자동화 스크립트에서 사용

```bash
#!/bin/bash
# Worktree 자동 정리 스크립트

moai inventory --type worktrees --json | jq -r '.inventory.worktrees[] | select(.status == "stale") | .name' | while read name; do
  echo "Removing stale worktree: $name"
  moai worktree remove "$name"
done
```

## 출력 해석

### Status 필드

| Status | 의미 |
|--------|------|
| `active` | 현재 사용 중 |
| `idle` | 일시 중단 (세션이 명시적으로 일시중지 상태) |
| `stale` | 사용되지 않음 (7일 이상 미접속) |
| `error` | 오류 상태 (확인 필요) |

### Phase 필드

| Phase | 설명 |
|-------|------|
| `plan` | Plan 단계 실행 중 |
| `run` | Run 단계 실행 중 |
| `sync` | Sync 단계 실행 중 |
| `completed` | 완료 상태 |

## 관련 문서

- [SPEC 기반 개발](/workflow-commands/moai-plan) - SPEC 생명 사이클
- [Worktree 관리](/getting-started/worktree) - Worktree 격리 및 생명 사이클
- [Harness v4 Builder](/advanced/builder-agents) - 동적 팀 관리
- [CLI 참조](/getting-started/cli) - 다른 CLI 커맨드

{{< callout type="info" >}}
**팁**: `moai inventory`는 자동 정리 스크립트와 모니터링 대시보드에 활용할 수 있습니다. JSON 형식으로 자동 분석하면 프로젝트 상태를 항상 파악할 수 있습니다.
{{< /callout >}}
