---
id: SPEC-V3R3-STATUSLINE-FALLBACK-001
version: "0.1.0"
status: draft
created_at: 2026-05-10
updated_at: 2026-05-10
author: GOOS행님
priority: High
labels: [statusline, fallback, cli]
issue_number: null
---

# SPEC-V3R3-STATUSLINE-FALLBACK-001 (Compact)

## Requirements

### REQ-SF-001: Empty/Partial stdin Fallback
stdin JSON empty / `{}` / `null` / partial → project name(getcwd), git branch(filesystem), version(config), rate limits(API) 표시

### REQ-SF-002: Model Name Fallback
stdin model 누락 → `MOAI_LAST_MODEL` env → `~/.moai/state/last-model.txt` cache → 미표시. 정상 수신 시 cache write (atomic rename).

### REQ-SF-003: Workspace Directory Fallback
stdin workspace 없음 → `os.Getwd()` basename. stdin `cwd` ≠ `getcwd` 시 stdin 우선.

### REQ-SF-004: Cwd Guard Go Binary 흡수
`os.Getwd()` 실패/삭제 dir → `$HOME` Chdir → project name `~`. Shell wrapper cwd guard 제거 가능.

### REQ-SF-005: Cache File Write 안정성
부모 dir 자동 생성, atomic rename, write 실패 시 silent ignore.

## Acceptance Criteria (Key)

| ID | 시나리오 | 검증 |
|----|----------|------|
| AC-SF-001 | Empty stdin → project name + version 표시 | `echo "" \| moai statusline` |
| AC-SF-003 | `MOAI_LAST_MODEL` env → model name 표시 | env 설정 후 실행 |
| AC-SF-004 | Cache file → model name 표시 | file 생성 후 실행 |
| AC-SF-005 | 정상 stdin → cache file 생성 | file 내용 확인 |
| AC-SF-006 | Deleted cwd → HOME fallback + `~` | t.TempDir() 삭제 테스트 |
| AC-SF-008 | 정상 stdin → 기존 출력 무변경 | golden test 통과 |

## Exclusions

- settings.json `statusLine` key 소멸 원인 조사
- Git worktree branch 출력 불일치 수정
- Statusline UI 재설계
- 새로운 display mode
