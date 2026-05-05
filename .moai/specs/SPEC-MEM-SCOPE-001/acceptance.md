---
id: SPEC-MEM-SCOPE-001
acceptance_version: "0.1.1"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Acceptance Criteria — SPEC-MEM-SCOPE-001

## Given-When-Then Scenarios

### Scenario 1: 4 scope 디렉토리 자동 생성

**Given** a fresh project directory
**And** `~/.moai/org-memory/` does not yet exist

**When** the user invokes `moai memory init`

**Then** the command SHALL create `~/.moai/org-memory/`
**And** the command SHALL create `<project>/.moai/memory/`
**And** the command SHALL preserve `~/.claude/projects/<hash>/memory/` (if exists)
**And** each scope directory SHALL contain an empty `audit.jsonl` file

---

### Scenario 2: Scope resolver — 경로 → scope 매핑

**Given** the 4 scope directories exist

**When** the resolver is invoked with each of the following paths:
- `~/.moai/org-memory/conventions.md`
- `<project>/.moai/memory/lessons.md`
- `~/.claude/projects/abc123/memory/MEMORY.md`
- `<project>/.moai/memory/manager-ddd/notes.md`

**Then** the resolver SHALL return Org, Project, User, Agent respectively
**And** the Agent scope resolution SHALL include owner agent name "manager-ddd"

---

### Scenario 3: Audit log entry on write (with CAS blob)

**Given** a project memory file `<project>/.moai/memory/lessons.md` with content hash "abc123"
**And** an agent "manager-ddd" updates the file with new content yielding hash "def456"

**When** the write operation completes

**Then** an entry SHALL be appended to `<project>/.moai/memory/audit.jsonl`
**And** the entry SHALL contain ts (ISO 8601), agent="manager-ddd", action="write", file="lessons.md", hash_before="abc123", hash_after="def456", scope="project"
**And** the entry SHALL NOT contain the file content (privacy)
**And** the file `<project>/.moai/memory/.cas/def456` SHALL exist
**And** the CAS blob SHALL contain the verbatim new content
**And** the SHA-256 of CAS blob SHALL equal "def456" (integrity verification)

---

### Scenario 4: 동시성 — 10 goroutine concurrent write

**Given** 10 goroutines each writing to the same memory file with unique content

**When** all goroutines invoke the write helper concurrently

**Then** each write SHALL acquire LOCK_EX before modifying the file
**And** writes SHALL complete sequentially (serialized via flock)
**And** the audit log SHALL contain exactly 10 entries
**And** the final file content SHALL be the content of one of the 10 writes (no torn writes)
**And** zero data loss SHALL occur (other 9 contents are recorded as previous hashes in audit)

---

### Scenario 5: 30일 이내 rollback (audit + CAS lookup)

**Given** a memory file modified 5 times over 10 days, with audit entries recorded
**And** for each write, the corresponding raw blob is stored in `<scope>/.cas/<hash_after>`
**And** the user wants to revert to the state at day 3

**When** the user invokes `moai memory rollback lessons.md --to <day-3 ISO timestamp>`

**Then** the command SHALL scan audit.jsonl for entries with file="lessons.md" up to day-3 timestamp
**And** the command SHALL identify hash_after of the latest entry at or just before day-3 timestamp
**And** the command SHALL read raw blob from `<scope>/.cas/<hash_after>` (CAS lookup)
**And** the command SHALL write blob content back to lessons.md
**And** the file content SHA-256 SHALL match the day-3 hash_after
**And** a new audit entry SHALL be appended with action="rollback", hash_before=<pre-rollback>, hash_after=<day-3 hash>, rollback_to=<target ISO>
**And** original audit history SHALL be preserved (rollback is append-only)
**And** original CAS blobs SHALL be preserved (rollback does not delete CAS entries)

---

### Scenario 6: 30일 초과 rollback → reject

**Given** the audit log has entries from 35 days ago

**When** the user invokes `moai memory rollback file.md --to <35-day-old timestamp>`

**Then** the command SHALL reject with error: "out of rollback window: target older than 30 days"
**And** the file SHALL NOT be modified
**And** no audit entry SHALL be appended

---

### Scenario 7: Org scope read-only enforcement

**Given** an agent (non-admin) invokes write on `~/.moai/org-memory/policy.md`

**When** the write helper runs

**Then** the operation SHALL be rejected with error: "org scope is read-only for non-admin agents"
**And** no audit entry SHALL be appended
**And** the file SHALL NOT be modified

---

### Scenario 8: Audit archive (30-day rotation)

**Given** the project audit log contains 200 entries spanning 45 days

**When** the user invokes `moai memory archive`

**Then** entries older than 30 days SHALL be moved to `<project>/.moai/memory/archive/audit-<YYYY-MM>.jsonl`
**And** the main `audit.jsonl` SHALL retain only entries within the last 30 days
**And** the archive directory SHALL be created if absent

---

### Scenario 9: Cross-platform — Windows lock

**Given** the codebase runs on Windows

**When** an agent acquires a write lock via `WithExclusiveLock`

**Then** the implementation SHALL use `LockFileEx` instead of POSIX flock
**And** the lock semantics SHALL match Unix flock (exclusive, 5s timeout)
**And** unit tests on Windows CI runner SHALL PASS

---

## Edge Cases

### EC-1: audit.jsonl missing
If `audit.jsonl` does not exist in a scope directory when the first write occurs, the system SHALL create it with empty content before appending the new entry.

### EC-2: Lock timeout (5s exceeded)
If the file lock cannot be acquired within 5 seconds, the operation SHALL fail with error "lock timeout: file <path> held by another process". No audit entry SHALL be appended.

### EC-3: NFS filesystem
If the memory directory is on NFS, flock semantics may be unreliable. The system SHALL log a warning at init time but continue. Documented as known limitation.

### EC-4: Backward compat — existing `~/.claude/agent-memory/`
Existing agent memories at `~/.claude/agent-memory/<agent>/` SHALL continue to function. Migration to `<project>/.moai/memory/<agent>/` SHALL be optional via `moai memory migrate-agent-scope` command.

### EC-5: Audit log size > 10MB
If `audit.jsonl` exceeds 10MB, the system SHALL warn the user and recommend running `moai memory archive`. The system SHALL NOT auto-archive without user invocation.

### EC-6: Concurrent rollback
If two agents attempt rollback simultaneously, the LOCK_EX serialization SHALL ensure sequential execution. The second rollback SHALL operate on the post-first-rollback state.

### EC-7: CAS blob missing (data corruption)
If rollback resolves an audit entry to hash X but `<scope>/.cas/<X>` is missing (manual deletion or filesystem corruption), the rollback SHALL fail with explicit error "CAS blob missing for hash <X>; rollback aborted; audit log unchanged". The file SHALL NOT be modified.

### EC-8: CAS GC during active rollback window
If the GC scheduler runs and a blob is referenced by any audit entry within the past 30 days, the GC SHALL skip that blob. Only blobs unreferenced for 30+ days qualify for deletion (REQ-MS-019).

### EC-9: Idempotent CAS write
If a write produces the same hash as an existing CAS blob (e.g., two agents write identical content), the second write SHALL be idempotent: the existing blob is reused, and only the audit entry differs. No duplicate blob storage.

---

## Quality Gate Criteria

| Gate | Threshold | Evidence |
|------|-----------|----------|
| 4 scope directories created | unit test | EXISTS for all 4 |
| Scope resolver accuracy | 4 sample paths | 4/4 correct |
| Audit log integrity | 100 op simulation | 100/100 entries |
| CAS blob integrity | 100 write SHA-256 verification | 100/100 hash matches |
| CAS GC 30-day protection | unreferenced/referenced blob test | 0 false deletion |
| CAS idempotent write | duplicate hash write test | 1 blob, 2 audit entries |
| Concurrency safety | 10 goroutine write | 0 data loss |
| Rollback within 30 days | audit + CAS lookup test | 100% restore (SHA-256 match) |
| Out-of-window rejection | older-than-30-days | rejected |
| CAS blob missing (EC-7) | manual delete + rollback | explicit error, file unchanged |
| Org read-only | non-admin write attempt | rejected |
| Privacy: audit.jsonl no content | log scan for content | 0 violations |
| Privacy: CAS local-only | export scan for `.cas/` paths | 0 leaks |
| Cross-platform | Windows CI | PASS |
| Template-First sync | clean | `make build` diff |
| plan-auditor | PASS | auditor report |

---

## Definition of Done

- [ ] All 9 Given-When-Then scenarios PASS
- [ ] All 9 edge cases (EC-1 to EC-9) documented and handled
- [ ] All 16 quality gate criteria meet threshold
- [ ] `internal/memory/{scope,audit,cas,lock,rollback,archive}.go` with >= 90% coverage
- [ ] `cmd/moai/memory.go` with subcommands: init, list, read, write, rollback, archive
- [ ] `.claude/rules/moai/core/memory-scope.md` policy document
- [ ] CLAUDE.md cross-reference (§13 or §16)
- [ ] Template-First sync at `internal/template/templates/.moai/memory/.gitkeep`
- [ ] Cross-platform CI (ubuntu/macos/windows) PASS
- [ ] Backward compat for `~/.claude/agent-memory/` verified
- [ ] `make build` regenerates embedded.go cleanly
- [ ] CHANGELOG.md updated under Unreleased
- [ ] privacy review: audit.jsonl hashes only (no raw content), CAS local-only (no export paths)
- [ ] CAS GC 30-day window protection verified
- [ ] plan-auditor PASS
- [ ] dogfooding: project memory used by at least one agent post-merge

End of acceptance.md (SPEC-MEM-SCOPE-001).
