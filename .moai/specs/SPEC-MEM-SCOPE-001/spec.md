---
id: SPEC-MEM-SCOPE-001
status: draft
version: "0.1.1"
priority: Medium
labels: [memory-scope, audit-log, rollback, concurrency, file-lock, wave-3, tier-2]
issue_number: null
scope: [internal/memory, cmd/moai, .claude/rules/moai/core, internal/template/templates/.moai/memory]
blockedBy: []
dependents: []
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
wave: 3
tier: 2
---

# SPEC-MEM-SCOPE-001: Memory Scope Architecture

## HISTORY

- 2026-04-30 v0.1.0: 최초 작성. Wave 3 / Tier 2. Anthropic Managed Agents Memory의 scope hierarchy + audit log 권고를 본 프로젝트 filesystem-mounted memory 모델에 흡수. 4-level scope (org/project/user/agent) + JSONL audit + 30-day rollback + flock concurrency 보호.
- 2026-04-30 v0.1.1: BLOCKING 결함 수정 — REQ-MS-008 (audit replay → reconstruct) 와 REQ-MS-016 (audit log content 미기록) 모순 해결. **Content-Addressable Storage (CAS)** 도입: audit.jsonl은 hash만 기록, content는 `<scope>/.cas/<hash>` 별도 저장소에 보관. Privacy (audit log에 raw content 없음) + Reconstruct (CAS lookup으로 hash → content 복원) 둘 다 충족.

---

## 1. Goal (목적)

본 프로젝트의 단일 user 스코프 메모리 시스템을 **4-level scope architecture** (org / project / user / agent)로 확장하고, 각 scope에 audit log + Content-Addressable Storage (CAS) + 동시성 보호 (file lock) + 30-day rollback 메커니즘을 추가한다. Anthropic Managed Agents Memory의 의미적 모델을 filesystem 인프라로 시뮬레이션하여 멀티 sub-agent 환경의 메모리 무결성을 보장한다.

> **Privacy ↔ Reconstruct 양립 설계 (CAS)**: audit log (`<scope>/audit.jsonl`)는 timestamp/agent/action/file/hash만 기록 (raw content 미기록 — privacy 준수). Content는 별도 CAS 저장소 (`<scope>/.cas/<hash>` blob)에 SHA-256 hash 키로 저장. Rollback 시 audit log의 hash → CAS blob lookup으로 file content 복원. 이중 저장으로 privacy (audit log to be shared/exported) + reconstruct (CAS는 local-only, never exported) 둘 다 만족.

### 1.1 배경

- Anthropic blog "Claude Managed Agents Memory": "Stores can be shared across multiple agents with different access scopes." / "All changes are tracked with a detailed audit log... You can roll back to an earlier version."
- 본 프로젝트의 메모리는 단일 user 스코프 (`~/.claude/projects/<hash>/memory/`) → org/project 분리 부재
- audit log + rollback 부재 → 누가 언제 무엇을 변경했는지 추적 불가
- 동시성 보호 부재 → 멀티 sub-agent fan-out 시 last-writer-wins data loss 위험

### 1.2 비목표 (Non-Goals)

- True cross-org sync (다른 머신/사용자 간 동기화 — 향후 SPEC)
- SQLite 또는 RDB 기반 메모리 저장소
- Web UI 또는 dashboard
- Real-time replication
- 자동 conflict resolution (수동 merge 우선)
- 30일 초과 rollback (archive로만 보존)

---

## 2. Scope (범위)

### 2.1 In Scope

- 4-level scope 디렉토리 구조 신설:
  - `~/.moai/org-memory/` (org scope, read-only)
  - `<project>/.moai/memory/` (project scope, read-write)
  - `~/.claude/projects/<hash>/memory/` (user scope, 기존 유지)
  - `<project>/.moai/memory/<agent-name>/` (agent scope)
- `internal/memory/scope.go` — scope resolver
- `internal/memory/audit.go` — JSONL audit logger (hash only, no content)
- `internal/memory/cas.go` — Content-Addressable Storage (SHA-256 keyed blob store at `<scope>/.cas/<hash>`)
- `internal/memory/lock.go` — flock(2) / LockFileEx 기반 advisory lock
- `internal/memory/rollback.go` — 30-day rollback 로직 (audit hash → CAS lookup → file restore)
- `cmd/moai/memory.go` — `moai memory <list|read|write|rollback>` 명령
- `.claude/rules/moai/core/memory-scope.md` 정책 문서
- audit.jsonl schema 정의 (timestamp, agent, action, file, hash_before, hash_after) — content 미기록
- CAS schema: `<scope>/.cas/<sha256-hex>` 파일에 raw content 저장, GC는 30+ days unreferenced blob 제거
- Cross-platform 호환 (macOS/Linux/Windows)
- Template-First 동기화

### 2.2 Exclusions (What NOT to Build)

- Cross-org / cross-machine 동기화
- SQLite 또는 RDB 메모리 저장소
- Web UI / dashboard
- Real-time replication
- 자동 conflict resolution
- 30일 초과 rollback (archive 정책으로 분리)
- Network filesystem (NFS/SMB) 공식 지원
- 메모리 내용 암호화 (별도 SPEC)

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.23+)
- 의존: POSIX flock(2) (macOS/Linux) 또는 LockFileEx (Windows)
- 영향 디렉터리: `internal/memory/`, `cmd/moai/`, `~/.moai/`, `<project>/.moai/memory/`, `.claude/rules/moai/core/`
- 영향 파일: scope 디렉토리들 + `audit.jsonl` 파일들

---

## 4. Assumptions (가정)

- A1: 단일 머신/사용자 환경 (true cross-org sync는 향후)
- A2: filesystem POSIX flock 또는 Windows LockFileEx 가용
- A3: JSONL append-only로 audit log 일관성 확보
- A4: 30일 rollback은 git-like history로 충분
- A5: 사용자가 `moai memory init`으로 명시적 디렉토리 초기화
- A6: org-memory는 admin write로만 변경 (single-user 환경에서는 사용자 자신)

---

## 5. Requirements (EARS Format)

### 5.1 Ubiquitous Requirements

- **REQ-MS-001**: THE MEMORY SYSTEM SHALL define 4 scope levels: org, project, user, agent.
- **REQ-MS-002**: THE MEMORY SYSTEM SHALL maintain a separate `audit.jsonl` file per scope, recording every read/write/delete operation.
- **REQ-MS-003**: THE AUDIT LOG ENTRY SHALL include timestamp (ISO 8601), agent identifier, action (read/write/delete), file path, and content hash (before/after for write/delete).
- **REQ-MS-003b**: THE MEMORY SYSTEM SHALL maintain a Content-Addressable Storage (CAS) at `<scope>/.cas/<sha256-hex>` storing raw blob keyed by SHA-256 hash, ensuring rollback can reconstruct file content from audit log hash references.
- **REQ-MS-004**: THE MEMORY SYSTEM SHALL provide rollback capability for changes within the past 30 days.

### 5.2 Event-Driven Requirements

- **REQ-MS-005**: WHEN a memory write occurs, THE SYSTEM SHALL acquire an exclusive file lock on the target file before writing AND SHALL release the lock after the write completes (success or failure).
- **REQ-MS-006**: WHEN a memory read occurs, THE SYSTEM SHALL acquire a shared file lock and release it after reading.
- **REQ-MS-007**: WHEN a memory write completes, THE SYSTEM SHALL (a) write raw blob to `<scope>/.cas/<new-hash>` (idempotent — skip if blob already exists with same hash), AND (b) append an audit entry containing hash_before and hash_after to the corresponding scope's `audit.jsonl`.
- **REQ-MS-008**: WHEN the user invokes `moai memory rollback <file> --to <timestamp>`, THE SYSTEM SHALL (a) scan audit log entries for the target file up to the requested timestamp, (b) identify the hash_after of the entry at or just before the timestamp, (c) lookup the corresponding blob in `<scope>/.cas/<hash>`, AND (d) restore file content from the CAS blob.

### 5.3 State-Driven Requirements

- **REQ-MS-009**: WHILE org scope memory is accessed by a non-admin agent, THE SCOPE SHALL be read-only (write attempt rejected with permission error).
- **REQ-MS-010**: WHILE a file lock is held by another process, THE WAITING PROCESS SHALL block up to 5 seconds before timing out with a clear error.

### 5.4 Conditional (WHERE / IF) Requirements

- **REQ-MS-011**: WHERE the agent scope is `<project>/.moai/memory/<agent-name>/`, THE OWNING AGENT SHALL have full read/write permission AND OTHER AGENTS SHALL have read-only access by default.
- **REQ-MS-012**: WHERE the operating system is Windows, THE LOCK MECHANISM SHALL use LockFileEx instead of flock.
- **REQ-MS-013**: IF the audit log is older than 30 days, THEN THE SYSTEM SHALL move stale entries to `<scope>/archive/audit-<YYYY-MM>.jsonl`.
- **REQ-MS-014**: WHERE the rollback target timestamp is older than 30 days, THE SYSTEM SHALL reject the rollback with "out of rollback window" error.

### 5.5 Unwanted (Negative) Requirements

- **REQ-MS-015**: THE MEMORY SYSTEM SHALL NOT permit cross-org synchronization (single-user / single-machine boundary).
- **REQ-MS-016**: THE `audit.jsonl` FILE SHALL NOT record memory file content directly (only timestamp/agent/action/file/hashes); raw content is stored exclusively in the CAS store at `<scope>/.cas/<hash>` (see REQ-MS-003b, REQ-MS-007).
- **REQ-MS-016b**: THE CAS STORE (`<scope>/.cas/`) SHALL NOT be exported, shared, or transmitted outside the local machine (privacy boundary — CAS is local-only; only audit.jsonl with hashes may be shared if explicitly approved).
- **REQ-MS-017**: THE ROLLBACK OPERATION SHALL NOT delete the audit log entries (rollback is also an audit entry of action "rollback").
- **REQ-MS-018**: THE MEMORY SYSTEM SHALL NOT silently drop write operations on lock timeout (must surface error to caller).
- **REQ-MS-019**: THE CAS GARBAGE COLLECTOR SHALL NOT delete blobs referenced by any audit entry within the past 30 days (rollback window protection).

---

## 6. Success Criteria (성공 기준)

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| 4 scope 디렉토리 생성 | unit test | EXISTS for all 4 |
| Audit log 무결성 | 100 read/write/delete simulation | 100/100 entries |
| CAS blob 저장 무결성 | 100 write SHA-256 검증 | 100/100 hash matches |
| CAS GC 보호 | 30-day window blob 보존 검증 | 0 false deletion |
| Concurrency safety | 10 goroutine concurrent write | 0 data loss |
| Rollback 정확도 (audit + CAS lookup) | within-30-day rollback test | 100% restore |
| Out-of-window rejection | rollback older than 30 days | rejected |
| Privacy: audit.jsonl content 미기록 | log scan for raw content | 0 violations |
| Privacy: CAS local-only | CAS path not in any export | 0 leaks |
| Windows 호환 | windows CI runner | PASS |
| `moai memory` CLI | sub-commands list/read/write/rollback | 4/4 functional |

---

## 7. Acceptance References

See `acceptance.md` for Given-When-Then scenarios and Definition of Done.

---

## 8. Constraints

- C1: Cross-org sync 금지 (single-user 한정)
- C2: 메모리 내용 암호화 안 함 (별도 SPEC)
- C3: NFS/SMB 공식 지원 안 함 (local filesystem 권장)
- C4: 모든 변경은 Template-First Rule 준수
- C5: 30일 rollback 윈도 hardcoded default; 사용자가 `.moai/config/sections/memory.yaml.rollback_days` 조정 가능

End of spec.md (SPEC-MEM-SCOPE-001 v0.1.0).
