# Research — SPEC-MEM-SCOPE-001 (Memory Scope Architecture)

**SPEC**: SPEC-MEM-SCOPE-001
**Wave**: 3 / Tier 2 (검증 통과 — 4-level scope + audit log)
**Created**: 2026-04-30
**Author**: manager-spec

---

## 1. 출처 (Anthropic 공식 자료)

### 1.1 Verbatim 인용

Anthropic blog "Claude Managed Agents Memory":

> "Stores can be shared across multiple agents with different access scopes."

> "All changes are tracked with a detailed audit log. Every read, write, and delete is timestamped and attributed to the agent that performed the action. You can roll back to an earlier version."

> "Scope hierarchy reflects who needs to know. Organization-level memories are read-only for individual agents but writable by an admin process. Project-level memories carry the team's institutional knowledge. Per-agent memories are private working notes."

### 1.2 검증 (claude-code-guide 검증 완료)

claude-code-guide 에이전트가 Claude Code memory model 기준으로 검증함. 결론:

- **호환성**: ⚠️ 부분 지원 — Managed Agents Memory는 Claude Code sub-agent context에 미적용. 본 SPEC은 **filesystem-mounted memory 시뮬레이션** 형태 (디렉토리 + 룰 + audit log)
- **표준 우회**: org/project/user/agent 4-level 디렉토리 + lock 메커니즘 + JSONL audit log 조합
- **권고 채택**: ACCEPT — Anthropic의 의미적 모델 + 본 프로젝트 filesystem 인프라 결합

---

## 2. 현재 상태 (As-Is)

### 2.1 기존 메모리 위치

| 위치 | 스코프 | 누가 쓰는가 | 관찰 |
|------|--------|-----------|------|
| `~/.claude/projects/<hash>/memory/` | per-project per-user | Claude Code agent | 표준 위치, 단일 user 한정 |
| `~/.claude/agent-memory/<agent-name>/` | per-agent (user) | 명시적 agent | 일부 agent body에 정의 |
| `<project>/.moai/specs/<ID>/progress.md` | per-SPEC | 작업 중 | 명시 작성 |
| `<project>/.moai/research/observations/` | design domain | learner | design 한정 |

**관찰**:
- org-wide (조직 단위) 메모리 부재
- project-wide (워크트리 무관 프로젝트 단위) 메모리 부재
- audit log 부재 (누가 언제 무엇을 썼는지 추적 불가)
- rollback 부재

### 2.2 권한 모델

현재 모든 메모리는:
- 파일시스템 권한 (POSIX rwx)
- 읽기/쓰기 분리 부재 (sub-agent가 메모리 디렉토리 보유 시 자유롭게 변경)

### 2.3 동시성 (concurrency)

여러 sub-agent가 동시 spawn되어 동일 메모리 파일에 쓸 경우:
- file lock 부재
- last-writer-wins (data loss 가능)

---

## 3. 격차 분석 (Gap Analysis)

| 영역 | 현재 (As-Is) | 목표 (To-Be) | 격차 |
|------|-------------|-------------|------|
| org scope | 부재 | `~/.moai/org-memory/` (read-only) | 신규 |
| project scope | 부재 | `<project>/.moai/memory/` (read-write) | 신규 |
| user scope | 존재 (`~/.claude/projects/<hash>/memory/`) | 유지 | 변경 없음 |
| agent scope | 부분적 (`~/.claude/agent-memory/`) | `<project>/.moai/memory/<agent-name>/`로 통일 | 위치 정상화 |
| audit log | 부재 | `<scope>/audit.jsonl` (hash only) | 신규 |
| CAS (Content-Addressable Storage) | 부재 | `<scope>/.cas/<hash>` blob store | 신규 (privacy ↔ reconstruct 양립) |
| rollback | 부재 | 30 days 이내 revert (audit hash → CAS lookup) | 신규 |
| 동시성 보호 | 부재 | file lock (advisory) | 신규 |

---

## 4. 코드베이스 분석 (Affected Files)

### 4.1 Primary 신규 대상

| 파일 | 수정 유형 | 변경 사유 |
|------|----------|----------|
| `internal/memory/scope.go` | 신규 | 4-level scope resolver |
| `internal/memory/audit.go` | 신규 | JSONL audit logger |
| `internal/memory/lock.go` | 신규 | advisory file lock (flock 기반) |
| `internal/memory/rollback.go` | 신규 | 30-day rollback 로직 |
| `cmd/moai/memory.go` | 신규 | `moai memory <list|read|write|rollback>` 명령 |
| `.claude/rules/moai/core/memory-scope.md` | 신규 | 정책 문서 |
| `internal/template/templates/.moai/memory/.gitkeep` | 신규 | 빈 디렉토리 유지 |

### 4.2 디렉토리 구조

```
~/.moai/org-memory/                       # org scope (read-only for agents, admin writable)
  ├── audit.jsonl                          # org-level audit log (hash only)
  └── .cas/                                # local-only blob store (NEVER exported)
      └── <sha256-hex>                     # raw content blob

<project>/.moai/memory/                    # project scope (read-write)
  ├── audit.jsonl                          # project audit log (hash only)
  ├── lessons.md                           # project-wide lessons
  ├── conventions.md                       # team conventions
  ├── .cas/                                # local-only blob store (NEVER exported)
  │   └── <sha256-hex>                     # raw content blob
  └── <agent-name>/                        # agent-specific subdirectory
      ├── audit.jsonl                      # agent audit log
      ├── .cas/                            # agent CAS
      └── notes.md                         # agent's private notes

~/.claude/projects/<hash>/memory/          # user scope (existing, unchanged)
  └── MEMORY.md
```

### 4.2.1 Privacy ↔ Reconstruct 분리 모델

| 저장소 | 내용 | 위치 | Export 가능 | 목적 |
|--------|------|------|-------------|------|
| `audit.jsonl` | hash + metadata only | scope root | YES (privacy-safe, 공유 가능) | 변경 이력 audit trail |
| `.cas/<hash>` | raw content blob | scope/.cas/ | NO (local-only) | rollback 시 content 복원 |

이중 저장으로:
- **Privacy 충족**: audit.jsonl에 raw content 미기록 → 외부 공유/export 안전
- **Reconstruct 충족**: rollback 시 audit hash → CAS blob lookup → file 복원 가능
- **Idempotent write**: 동일 content는 동일 hash → 단일 blob 재사용 (중복 저장 X)
- **GC**: 30일+ 미참조 blob만 삭제 (REQ-MS-019), 30일 이내 referenced blob 보호

### 4.3 audit.jsonl schema

```jsonl
{"ts":"2026-04-30T12:34:56Z","agent":"manager-ddd","action":"write","file":"lessons.md","hash_before":"abc123","hash_after":"def456","scope":"project"}
{"ts":"2026-04-30T12:35:01Z","agent":"manager-tdd","action":"read","file":"lessons.md","scope":"project"}
{"ts":"2026-04-30T12:35:30Z","agent":"manager-ddd","action":"delete","file":"obsolete.md","hash_before":"xyz789","scope":"agent","agent_target":"manager-ddd"}
```

### 4.4 lock 메커니즘

POSIX flock(2) 또는 Go `golang.org/x/sys/unix.Flock`:
- write 시 LOCK_EX (exclusive)
- read 시 LOCK_SH (shared)
- timeout: 5s (deadlock 방지)

---

## 5. 위험 및 가정

### 5.1 Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| audit log 무한 성장 | High | Medium | 30 day rotation + archive 정책 |
| flock이 NFS/SMB에서 미작동 | Medium | High | local filesystem 권장 명시 + warning |
| org-memory가 사용자별 다름 (실제 org-wide 아님) | High | Low | 명명: "user의 org-level 분리"로 정확히 표기 |
| rollback 시 audit log 무결성 깨짐 | Medium | High | rollback도 audit entry로 기록 (revert action) |
| Windows 호환 (flock 부재) | Medium | High | Windows는 LockFileEx 사용 |
| 4-level scope 학습 곡선 | High | Medium | 정책 문서 + 결정 트리 명시 |

### 5.2 Assumptions

- A1: 단일 머신/사용자 환경 가정 (true cross-org sync는 향후 SPEC)
- A2: filesystem POSIX flock 또는 Windows LockFileEx 가용
- A3: JSONL append-only 쓰기로 audit log 일관성 확보
- A4: rollback은 30일 이내 git-like history로 충분
- A5: 동시 쓰기는 sub-agent fan-out 시 발생 가능 (Wave 3 SPEC-PARALLEL-COOK-001 패턴 적용)

---

## 6. 측정 계획 (Baseline + Validation)

| Metric | Baseline 측정 방법 | 목표 |
|--------|-------------------|------|
| 4 scope 디렉토리 생성 | unit test | EXISTS |
| audit log 무결성 | 100 read/write/delete 시뮬레이션 | 100/100 entry 기록 |
| 동시 쓰기 안전성 | 10 goroutine concurrent write | data loss 0건 |
| rollback 정확도 | 30일 이내 entry revert | 100% restore |
| Windows 호환 | win runner CI | PASS |

---

## 7. 대안 검토 (Alternatives Considered)

| 대안 | 채택 여부 | 이유 |
|------|----------|------|
| SQLite 기반 메모리 저장소 | ❌ | 파일 기반이 git-friendly, audit + rollback git처럼 단순 |
| YAML audit log | ❌ | append-only는 JSONL이 적합 (한 줄 = 한 entry) |
| Managed Agents Memory 강제 | ❌ | Claude Code sub-agent에 미지원 |
| 3-level scope (org 제외) | ❌ | Anthropic 권고 정합성 우선, org는 single-user에서도 의미적 가치 |
| 5-level scope (session 추가) | ❌ | session은 user scope로 흡수 가능, 복잡도 비용 > 가치 |
| Audit log에 raw content 직접 기록 (CAS 미사용) | ❌ | privacy 위반 (audit log 공유 시 raw 누출), 또한 audit log 무한 성장 |
| Rollback scope를 metadata-only로 축소 (file content 복원 포기) | ❌ | rollback 의도 미충족 — 사용자는 file 복원을 요구 |
| CAS 대신 git-based snapshot (e.g., libgit2) | ❌ | git 의존성 추가, blob storage는 더 가벼운 SHA-256 keyed file store로 충분 |

---

## 8. 참고 SPEC

- SPEC-MEMO-001: 기존 memory 시스템
- SPEC-PERSIST-001: persistent state 관리 — scope 분리 의미 보강
- SPEC-CONTEXT-INJ-001 (이번 wave sibling): orchestrator의 메모리 주입 정책 — scope를 알아야 주입 우선순위 결정

---

## 9. Open Questions (Plan 단계 해결 대상)

- OQ1: org scope의 admin 권한 모델 (실질적으로 user 단일 환경에서 의미는?)
- OQ2: rollback이 destructive operation인가, append-only revert인가?
- OQ3: agent scope를 `~/.claude/agent-memory/`에서 `<project>/.moai/memory/<agent>/`로 이전 시 backward compat 정책?
- OQ4: audit log 보존 30일 이후 archive 위치 (`<scope>/archive/`)?
- OQ5: cross-scope read (e.g., manager-ddd가 manager-tdd의 agent-scope 읽기) 정책?

---

End of research.md (SPEC-MEM-SCOPE-001).
