---
id: SPEC-MEM-SCOPE-001
plan_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Implementation Plan — SPEC-MEM-SCOPE-001

## 1. Overview

4-level memory scope (org/project/user/agent) + JSONL audit log + advisory file lock + 30-day rollback 메커니즘 신설. Anthropic Managed Agents Memory의 의미적 모델을 filesystem 인프라로 시뮬레이션.

## 2. Approach Summary

**전략**: Filesystem-Mounted-Simulation, JSONL-Append-Only, Cross-Platform-Lock.

1. 4 scope 디렉토리 구조 정의
2. `internal/memory/scope.go` — scope resolver (path → scope level 매핑)
3. `internal/memory/audit.go` — JSONL audit logger (append-only, atomic)
4. `internal/memory/lock.go` — flock (Unix) + LockFileEx (Windows)
5. `internal/memory/rollback.go` — 30-day audit replay → file 복원
6. `cmd/moai/memory.go` — list/read/write/rollback CLI
7. `.claude/rules/moai/core/memory-scope.md` 정책 문서

## 3. Milestones (Priority-based, no time estimates)

### M0 — Pre-flight (Priority: Critical)

- [ ] 기존 메모리 위치 verbatim 캡처 (research.md §2.1 표 검증)
- [ ] Go에서 flock 사용 가능 라이브러리 확인 (`golang.org/x/sys/unix.Flock`)
- [ ] Windows LockFileEx 호출 패턴 확인
- [ ] 기존 `~/.claude/projects/<hash>/memory/`와 conflict 없는지 검토
- [ ] backward compat 정책 검토 (`~/.claude/agent-memory/<agent>/`)

**Exit Criteria**: lock library 결정, backward compat 명확

### M1 — Scope Directory Structure (Priority: High)

- [ ] 디렉토리 생성 helper:
  - `~/.moai/org-memory/` 자동 생성 (사용자 첫 호출 시)
  - `<project>/.moai/memory/` 자동 생성
  - `<project>/.moai/memory/<agent-name>/` 자동 생성 (agent 호출 시)
  - `~/.claude/projects/<hash>/memory/` 변경 없음 (기존 유지)
- [ ] `cmd/moai/memory.go init` 서브커맨드: 디렉토리 일괄 생성
- [ ] Template-First: `internal/template/templates/.moai/memory/.gitkeep` 추가

**Exit Criteria**: 4 scope 디렉토리 생성 검증

### M2 — Scope Resolver (Priority: High)

- [ ] `internal/memory/scope.go`:
  - `func ResolveScope(path string) (Scope, error)` — 절대 경로 → scope enum
  - Scope enum: Org / Project / User / Agent
  - agent scope의 경우 owner agent 추출 (`<project>/.moai/memory/<agent-name>/...`)
- [ ] unit test: 각 scope별 path 5개씩 입력

**Exit Criteria**: 4 scope 모두 정확 resolve

### M3 — Audit Log (JSONL Append) (Priority: High)

- [ ] `internal/memory/audit.go`:
  - `type AuditEntry struct { Timestamp, Agent, Action, File, HashBefore, HashAfter, Scope }`
  - `func Append(scope Scope, entry AuditEntry) error` — atomic append
  - 파일 경로: `<scope-root>/audit.jsonl`
  - JSON Lines (한 줄 = 한 entry)
- [ ] atomic write: O_APPEND + flock LOCK_EX (또는 atomic rename pattern)
- [ ] privacy: 메모리 파일 내용은 hash만 저장, raw content 절대 미기록

**Exit Criteria**: 100 read/write/delete simulation에서 100/100 entries

### M4 — File Lock (Concurrency) (Priority: High)

- [ ] `internal/memory/lock.go`:
  - Unix: `unix.Flock(fd, LOCK_EX|LOCK_NB)` 또는 LOCK_SH
  - Windows: `LockFileEx(handle, LOCKFILE_EXCLUSIVE_LOCK, ...)` (또는 `golang.org/x/sys/windows`)
  - 5s timeout 후 error
- [ ] read는 LOCK_SH, write는 LOCK_EX
- [ ] defer unlock 표준 패턴

**Exit Criteria**: 10 goroutine concurrent write에서 0 data loss

### M5 — Rollback Logic (Priority: Medium)

- [ ] `internal/memory/rollback.go`:
  - `func Rollback(file string, toTimestamp time.Time) error`
  - audit log 읽기 → toTimestamp 시점의 file content 재구성
  - audit log replay 알고리즘: 가장 최근 write entry부터 역순으로
  - 30일 초과 시점은 reject (REQ-MS-014)
- [ ] rollback 자체도 audit entry로 기록 (action: "rollback")

**Exit Criteria**: 30일 이내 rollback 100% restore

### M6 — Org Scope Read-Only Enforcement (Priority: Medium)

- [ ] org scope (`~/.moai/org-memory/`) write 시도 시 admin check:
  - default: 모든 sub-agent는 read-only
  - admin write는 사용자 명시 호출 (`moai memory write --scope org --as admin`) 또는 환경변수
- [ ] write reject 시 명확한 error message

**Exit Criteria**: 비-admin agent의 org write 100% reject

### M7 — Audit Archive (30-day Rotation) (Priority: Medium)

- [ ] `internal/memory/archive.go`:
  - 30일 초과 audit entry → `<scope>/archive/audit-<YYYY-MM>.jsonl`
  - 사용자 명시 호출 (`moai memory archive`) 또는 hook 자동 (low priority)
- [ ] archive 후 main audit log는 30일 이내만 보유

**Exit Criteria**: archive 동작 검증

### M8 — CLI Commands (Priority: High)

- [ ] `cmd/moai/memory.go`:
  - `moai memory init` (M1)
  - `moai memory list [--scope <s>]`
  - `moai memory read <file>`
  - `moai memory write <file> [--scope <s>]`
  - `moai memory rollback <file> --to <ISO timestamp>`
  - `moai memory archive` (M7)
- [ ] `--scope`: org / project / user / agent
- [ ] verbose mode (-v)

**Exit Criteria**: 모든 서브커맨드 functional

### M9 — Policy Document + Templates (Priority: Medium)

- [ ] `.claude/rules/moai/core/memory-scope.md` 정책 문서 작성:
  - 4 scope 정의 + 의미
  - 각 scope의 owner / writer / reader 매트릭스
  - audit log 정책
  - rollback 정책
  - cross-scope read 정책
- [ ] CLAUDE.md §16 Context Search Protocol 또는 §13에 cross-ref
- [ ] Template-First 동기화

**Exit Criteria**: 정책 문서 + cross-ref

### M10 — Cross-Platform Validation (Priority: Medium)

- [ ] macOS/Linux flock 동작 검증
- [ ] Windows LockFileEx 동작 검증
- [ ] CI 3 runners PASS

**Exit Criteria**: 3 OS PASS

### M11 — Validation + Acceptance Sign-off (Priority: High)

- [ ] acceptance.md 시나리오 모두 PASS
- [ ] privacy: hash만 기록, content 누출 없음
- [ ] backward compat: 기존 user scope (`~/.claude/projects/<hash>/memory/`) 동작 유지
- [ ] plan-auditor PASS

**Exit Criteria**: 모든 acceptance + plan-auditor PASS

## 4. Technical Approach

### 4.1 디렉토리 매핑 (의사코드)

```go
func ResolveScope(absPath string) (Scope, error) {
    home, _ := os.UserHomeDir()
    if strings.HasPrefix(absPath, filepath.Join(home, ".moai", "org-memory")) {
        return Org, nil
    }
    if strings.HasPrefix(absPath, filepath.Join(home, ".claude", "projects")) {
        return User, nil
    }
    // project/agent 구분
    projectMemoryRoot := findProjectRoot(absPath) + "/.moai/memory"
    if strings.HasPrefix(absPath, projectMemoryRoot) {
        rel := strings.TrimPrefix(absPath, projectMemoryRoot+"/")
        if strings.Count(rel, "/") >= 1 {  // <agent-name>/file
            return Agent, nil
        }
        return Project, nil
    }
    return 0, fmt.Errorf("path %q not in any memory scope", absPath)
}
```

### 4.2 audit.jsonl entry 예시

```jsonl
{"ts":"2026-04-30T12:34:56Z","agent":"manager-ddd","action":"write","file":"lessons.md","hash_before":"abc123","hash_after":"def456","scope":"project"}
{"ts":"2026-04-30T12:35:01Z","agent":"manager-tdd","action":"read","file":"lessons.md","scope":"project"}
{"ts":"2026-04-30T12:35:30Z","agent":"manager-ddd","action":"rollback","file":"lessons.md","hash_before":"def456","hash_after":"abc123","scope":"project","rollback_to":"2026-04-30T12:34:00Z"}
```

### 4.3 lock 패턴 (Unix 예시)

```go
func WithExclusiveLock(file string, fn func() error) error {
    f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0644)
    if err != nil { return err }
    defer f.Close()
    if err := unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil {
        return fmt.Errorf("lock acquire: %w", err)
    }
    defer unix.Flock(int(f.Fd()), unix.LOCK_UN)
    return fn()
}
```

## 5. Risks and Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| audit log 무한 성장 | High | Medium | M7 30-day rotation + archive |
| flock NFS/SMB 미작동 | Medium | High | local filesystem 권장 명시 + warning 로그 |
| Windows LockFileEx complexity | Medium | High | well-tested library 사용 (`golang.org/x/sys/windows`) |
| rollback 시 audit log 무결성 | Medium | High | rollback도 audit entry (action: "rollback") |
| backward compat (기존 위치 사용자) | High | Medium | 기존 user scope 유지, agent scope 이전은 optional migration |
| 4-level 학습 곡선 | High | Medium | 정책 문서에 결정 트리 명시 |

## 6. Dependencies

- 선행 SPEC: 없음 (독립)
- 의존: `golang.org/x/sys/unix` (Unix flock), `golang.org/x/sys/windows` (Windows lock)
- sibling SPEC: SPEC-CONTEXT-INJ-001 (memory injection 정책 — scope 인지 필요)
- 도구: `make build`, plan-auditor

## 7. Open Questions Resolution

- **OQ1** (org admin 권한 in single-user): 사용자 명시 admin flag (`--as admin`) 또는 환경변수 (`MOAI_MEMORY_ADMIN=1`)
- **OQ2** (rollback destructive vs append-only revert): append-only revert (rollback도 audit entry, 원본 history 보존)
- **OQ3** (backward compat for `~/.claude/agent-memory/`): 기존 위치 유지, 이전은 optional `moai memory migrate-agent-scope` 명령
- **OQ4** (archive 위치): `<scope>/archive/audit-<YYYY-MM>.jsonl`
- **OQ5** (cross-scope read): default 허용 (특히 project/user/agent), org 만 read-only enforcement

## 8. Rollout Plan

1. M1-M8 구현 후 dogfooding: 본 프로젝트에서 `moai memory init` 실행
2. agent별 메모리 5개 시드 → audit log 검증
3. CHANGELOG + v2.x.0 minor release
4. backward compat 시나리오 (기존 사용자) 회귀 테스트

End of plan.md (SPEC-MEM-SCOPE-001).
