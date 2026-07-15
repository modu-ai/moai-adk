---
title: moai inventory 커맨드
weight: 10
draft: false
---

현재 프로젝트의 활성 세션, 워크트리, 하네스를 한눈에 조회하는 `moai inventory` 커맨드를 안내합니다.

{{< callout type="info" >}}
**한 줄 요약**: `moai inventory`는 현재 프로젝트의 활성 자원(세션, 워크트리, 하네스)을 읽기 전용으로 조회합니다. `--json` 으로 구조화된 출력을 받아 스크립트에 활용할 수 있습니다.
{{< /callout >}}

## 개요

`moai inventory`는 읽기 전용 명령어로, 병렬 세션과 워크트리를 여러 개 운영할 때 "지금 뭐가 돌아가고 있지?"를 한 번에 확인하는 통합 뷰를 제공합니다.

### 조회 대상

| 자원 | 설명 | 데이터 원본 |
|------|------|------------|
| **Sessions** | 활성 Claude Code 세션 | `.moai/state/active-sessions.json` |
| **Worktrees** | 프로젝트용 Git 워크트리 | Git worktree 목록 |
| **Harnesses** | 등록된 하네스 | `.moai/harness/` 매니페스트 |

## 명령어 형식

```bash
moai inventory [OPTIONS]
```

### 플래그

| 플래그 | 설명 |
|------|------|
| `--json` | 구조화된 JSON 출력 (머신 리더블) |
| `--project-root <path>` | 프로젝트 루트 경로 (기본값: 현재 디렉토리) |

이 명령어는 위 두 플래그만 지원합니다. 필터링이나 상세 모드 플래그는 없습니다 — 필요한 가공은 `--json` 출력을 `jq` 등으로 처리합니다.

## 기본 사용

```bash
moai inventory
```

텍스트 형식으로 세션·워크트리·하네스 요약을 출력합니다.

## JSON 형식 출력

```bash
moai inventory --json
```

구조화된 JSON으로 출력하여 자동 분석이나 CI 스크립트에 활용할 수 있습니다.

## JSON 스키마

`--json` 출력의 최상위 구조는 세 섹션으로 이루어집니다.

```json
{
  "sessions": { ... },
  "worktrees": { ... },
  "harnesses": { ... }
}
```

각 섹션은 `count`, `entries`, 그리고 선택적 `error` 필드를 가집니다.

### Session 항목

```json
{
  "session_id": "edc25996",
  "spec_id": "SPEC-DOCS-001",
  "phase": "run"
}
```

| 필드 | 설명 |
|------|------|
| `session_id` | 세션 ID (단축형, 첫 8자) |
| `spec_id` | 연결된 SPEC ID |
| `phase` | 현재 Phase (`plan`, `run`, `sync`, `mx`) |

### Worktree 항목

```json
{
  "branch": "feat/auth",
  "path": "/home/user/.moai/worktrees/project/SPEC-AUTH-001",
  "head": "a1b2c3d4"
}
```

| 필드 | 설명 |
|------|------|
| `branch` | 워크트리 브랜치명 |
| `path` | 워크트리 파일시스템 경로 |
| `head` | HEAD 커밋 해시 (단축형, 첫 8자) |

### Harness 항목

```json
{
  "name": "backend-team",
  "domain": "backend",
  "manifest_missing": false
}
```

| 필드 | 설명 |
|------|------|
| `name` | 하네스 이름 |
| `domain` | 하네스 도메인 |
| `manifest_missing` | 매니페스트 파일 누락 여부 (`true`면 설정 불완전) |

### 전체 출력 예시

```json
{
  "sessions": {
    "count": 2,
    "entries": [
      { "session_id": "edc25996", "spec_id": "SPEC-DOCS-001", "phase": "run" },
      { "session_id": "a1b2c3d4", "spec_id": "SPEC-AUTH-002", "phase": "plan" }
    ]
  },
  "worktrees": {
    "count": 1,
    "entries": [
      { "branch": "feat/auth", "path": "/home/user/.moai/worktrees/project/SPEC-AUTH-001", "head": "a1b2c3d4" }
    ]
  },
  "harnesses": {
    "count": 1,
    "entries": [
      { "name": "backend-team", "domain": "backend", "manifest_missing": false }
    ]
  }
}
```

## 실용적인 사용 예시

### 1. 다중 세션 경합 감지

같은 SPEC을 다루는 세션이 2개 이상이면 경합 위험이 있습니다.

```bash
moai inventory --json | jq '[.sessions.entries[] | .spec_id] | group_by(.) | map(select(length > 1))'
```

### 2. 활성 워크트리 브랜치 목록

```bash
moai inventory --json | jq -r '.worktrees.entries[].branch'
```

### 3. 매니페스트 누락 하네스 찾기

`manifest_missing: true` 인 하네스는 설정이 불완전한 상태입니다.

```bash
moai inventory --json | jq '.harnesses.entries[] | select(.manifest_missing)'
```

### 4. 현재 진행 중인 Phase 분포

```bash
moai inventory --json | jq '[.sessions.entries[].phase] | group_by(.) | map({phase: .[0], count: length})'
```

## 관련 문서

- [CLI 레퍼런스](./cli) — 전체 CLI 명령어
- [프로젝트 상태](./status) — `moai status` 명령어
- [SPEC 기반 개발](/ko/workflow-commands/moai-plan) — SPEC 생명 주기

{{< callout type="info" >}}
**팁**: `moai inventory --json` 은 모니터링 대시보드와 CI 스크립트에 활용할 수 있습니다. 읽기 전용 명령어이므로 안전하게 자동화할 수 있습니다.
{{< /callout >}}
