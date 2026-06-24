# progress.md — SPEC-V3R6-ASKUSER-DECISION-MEMORY-001

> 본 파일은 plan-phase에서 §E 스켈레톤만 생성. §E.2/§E.3/§E.4 증거 콘텐츠는 run-phase(manager-develop) 및 sync-phase(manager-docs)에서 채운다. 본 에이전트(manager-spec)는 §E.1만 채운다.

---

## §A. 현재 상태

- **Phase**: plan-phase 완료
- **Status**: draft (frontmatter)
- **plan-auditor 독립 감사**: _<pending>_
- **Implementation Kickoff Approval**: _<pending>_

---

## §B. 산출물

| 파일 | 상태 |
|------|------|
| spec.md | 작성 완료 (plan-phase) |
| plan.md | 작성 완료 (plan-phase) |
| acceptance.md | 작성 완료 (plan-phase) |
| research.md | 작성 완료 (plan-phase) |
| design.md | 작성 완료 (plan-phase) |
| progress.md | 스켈레톤 (본 파일) |

---

## §C. 다음 단계

1. plan-auditor 독립 감사 (편향 방지)
2. Implementation Kickoff Approval (사용자 명시적 run-phase 진입 승인)
3. Pre-Spawn Sync Check (다중 세션 race 방지)
4. run-phase manager-develop 위임 (M1부터 순차)

---

## §D. PRESERVE-list (중단 시 복구용)

_<pending run-phase>_ — run-phase 진입 후 manager-develop이 채움.

---

## §E.1 Plan-phase Audit-Ready Signal

본 SPEC 디렉터리는 plan-phase 산출물 5종(spec/plan/acceptance/research/design) + 본 progress.md 스켈레톤으로 구성된다. SPEC ID 사전 작성 자체 점검 통과(`SPEC-V3R6-ASKUSER-DECISION-MEMORY-001` → 정준正규식 `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` PASS). frontmatter 12 정준 필드 + era: V3R6 + tier: M.

---

## §E.2 Run-phase Evidence

### M1 — preference 메모리 계층 (`internal/cli/preference/`)

**패키지 산출물**:
- `internal/cli/preference/entry.go` — Entry struct (7 필드) + typed Scope/Confidence enums + Validate()
- `internal/cli/preference/store.go` — Store interface (Upsert/Get/Query) + Tier enum + ErrInvalidEntry/ErrNotFound
- `internal/cli/preference/filestore.go` — fileStore 구현: 3-tier cascade (core.yaml / recall.jsonl / archival/), atomic upsert (temp+rename), core ≤4KB 강제 (demote on overflow), namespace 분리
- 테스트: `entry_test.go`, `store_test.go`, `store_helpers_test.go`, `filestore_coverage_test.go`, `error_paths_test.go`, `atomicwrite_test.go`

**테스트 결과** (`go test ./internal/cli/preference/...`):
```
ok  	github.com/modu-ai/moai-adk/internal/cli/preference	0.430s	coverage: 85.7% of statements
```

**레이스 검출** (`go test -race ./internal/cli/preference/...`):
```
ok  	github.com/modu-ai/moai-adk/internal/cli/preference	1.515s
```

**크로스 플랫폼 빌드**:
- `go build ./...` → exit 0
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0

**Lint** (`golangci-lint run --timeout=2m ./internal/cli/preference/...`):
- 0 issues (staticcheck SA9003 empty-branch 수정 후 clean)

**Subagent boundary grep** (C-HRA-008):
- `grep -rn 'AskUserQuestion' internal/cli/preference/ | grep -v _test.go | grep -v "// "` → 0 matches

**commit SHA**: M1 커밋 후 이 섹션에 백fill

---

## §E.3 Run-phase Audit-Ready Signal

### M1 AC PASS 매트릭스

| AC | Status | 검증 명령 | 실제 출력 |
|----|--------|-----------|-----------|
| AC-ADM-001 (upsert 멱등성 — replace not append) | PASS | `go test -run TestUpsert_Idempotent_ReplaceNotAppend ./internal/cli/preference/...` | `--- PASS: TestUpsert_Idempotent_ReplaceNotAppend (0.01s)` |
| AC-ADM-002 (네임스페이스 분리 — feedback vs user_decisions 독립 쿼리) | PASS | `go test -run TestNamespaceSeparation_UserDecisionsVsFeedback ./internal/cli/preference/...` | `--- PASS: TestNamespaceSeparation_UserDecisionsVsFeedback (0.01s)` |
| AC-ADM-003 (7-필드 스키마 + 누락 필드 거부) | PASS | `go test -run TestEntry_SevenFieldsPresent ./internal/cli/preference/... && go test -run TestEntry_Validate_RejectsMissingFields ./internal/cli/preference/...` | 2 PASS (8 subtests — missing fact/source_citation/valid_time/last_used/domain/decision_key + invalid scope/confidence 각각 거부) |
| AC-ADM-004 (3-tier cascade — core hit 시 recall/archival 미접근 + recall hit after core miss) | PASS | `go test -run TestCascade_CoreHitSkipsRecallAndArchival ./internal/cli/preference/... && go test -run TestCascade_RecallHitAfterCoreMiss ./internal/cli/preference/...` | 2 PASS (TierCore / TierRecall 반환 검증) |
| AC-ADM-NFR-002 (core ≤4KB 강제 + 초과 시 recall 강등) | PASS | `go test -run TestCoreSizeEnforcement_DemotesOnOverflow ./internal/cli/preference/...` | `--- PASS: TestCoreSizeEnforcement_DemotesOnOverflow (0.03s)` — 20 entries upsert 후 core.yaml ≤4096 bytes + 강등 엔트리 cascade 회수 |

**atomic upsert (SIGKILL 방어, AC-ADM-001 edge case)**: `TestUpsert_AtomicWriteSurvivesPartialState` — recall.jsonl write 후 임시 파일 잔류 0건 검증 PASS.

**coverage**: 85.7% (≥85% threshold PASS per acceptance.md §D.6 quality gate).

**Gaps (미검증 — verification-claim-integrity §3.4 준수)**:
- 다중 프로세스 동시 캡처 race (design.md §C cohabitation) — 본 패키지는 단일 프로세스 내 sequential 안전만 보장; cross-process 동시성은 M3 PostToolUse 캡처 훅 통합 시 advisory/fail-open (REQ-ADM-009)으로 처리됨. 본 M1 unit test는 cross-process 시나리오를 다루지 않음.
- `internal/hook/post_tool.go` user_decision_capture 서브파이프라인 배선 — M3 scope; 본 M1은 패키지만 제공.

**Residual-risk (잔여 위험 — verification-claim-integrity §3.5)**:
- core.yaml 4KB 강제가 매우 긴 단일 Fact(>4KB)에서 데드락 가능성 — 현재 `upsertToCore`는 `len(cf.Entries)==0` 조건으로 루프 탈출하므로 단일 엔트리가 4KB 초과해도 멈추지 않음 (core.yaml 자체가 4KB 초과 상태로 남음). 실제 선호 Fact는 수백 바이트 수준이므로 실용적 한계 내. 단일-엔트리 캡 도입은 complete-tier 후속 검토 권장.

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_ — manager-docs가 sync_commit_sha + CHANGELOG/README 업데이트 증거로 채움.
