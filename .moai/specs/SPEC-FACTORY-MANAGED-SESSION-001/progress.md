# progress.md — SPEC-FACTORY-MANAGED-SESSION-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: authored-pending-audit
- plan_complete_at: 2026-10-01 (plan-phase artifacts authored — plan-audit `--deep` 대기 중)
- artifacts: spec.md (REQ 15), plan.md (M1..M6), acceptance.md (AC 17), design.md (D-1..D-7), research.md (R1..R14)
- lint: `go run ./cmd/moai spec lint SPEC-FACTORY-MANAGED-SESSION-001` → `✓ No findings — all SPEC documents are valid` (1차 0 error/17 warning — Coverage 2·moving-ref 1·run-패턴 앵커 14 수리 후 재실행, 이번 런 2026-10-01)
- tree: WT-crosshost-rebuild @ f22e2d7ac (plan artifacts uncommitted — audit 통과 후 커밋)
- audit gate: plan-auditor `--deep` — 통과가 run 착수 조건(카드 t1375 절차)

## §E.2 Run-phase Evidence

### M1 — 관리 세션 코어 인터페이스 + Claude/GLM 소유자 (2026-10-01)

측정 기준: HEAD `63e433fe8` + M1 작업 변경분(커밋 직전 워킹 트리). 파일:
`internal/cli/managed_factory_session.go` (+test), `internal/config/defaults.go` 튜너블 3종.
M1 미커버 AC(AC-MS-001/002=M2, 011/012/013=M3·M4, 015/016=M5)는 대상 마일스톤에서 채움.

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-MS-003 | PASS | `go test ./internal/cli -run '^TestManagedLaunchPendingRollback$' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli` ( 롤백 후 launch-pending 행 0건 — 양성 대조로 직전 등록 행 관측 후 롤백) |
| AC-MS-004 | PASS | `go test ./internal/cli -run '^TestManagedInboxPromptMetadataOnly$' -count=1` | `ok …` (메타데이터 5종+run_id 포함, 본문 "SECRET_BODY" 부재, 도구명·untrusted 경고 포함) |
| AC-MS-006 | PASS | `go test ./internal/cli -run '^TestManagedReceiptAcknowledges$' -count=1` | `ok …` (acknowledged 1·pending 0, 위조 claim token 거부) |
| AC-MS-007 | PASS | `go test ./internal/cli -run '^TestManagedSessionOwnsStreamFlags$' -count=1` | `ok …` (소유 플래그 6형 거부+플래그명 오류, 강제 플래그 prefix 검증) |
| AC-MS-008 | PASS | `go test ./internal/cli -run '^TestManagedQueueSerializesOperatorAndInbox$' -count=1` | `ok …` (연산자 우선 큐 순서, busy 중 claim 0, claim 직후 즉시 배달) |
| AC-MS-009 | PASS | `grep -rnE 'fmt\.Sprintf\("(agent\|worker)-' internal/cli/managed_*.go` | 0행 (exit 1) |
| AC-MS-010 | PASS | `grep -rnE 'merge-window\|Decider\|T29b\|T29c\|handover' internal/cli/managed_*.go` | 0행 (exit 1) |
| AC-MS-014 | PASS | `GOOS=windows GOARCH=amd64 go build ./...` + `grep -rn 'syscall\.' internal/cli/managed_*.go` | `BUILD-WINDOWS-OK` + 0행 (exit 1) |
| AC-MS-017 | PASS | `git log --oneline f22e2d7ac..HEAD -- internal/factorymsg/store.go` + `git diff --stat internal/factorymsg/store.go` | 빈 출력 ×2 (merge-base = f22e2d7ac 재산출 일치 — 흡수 없음) |

품질 게이트 (M1):
- 커버리지: `go test ./internal/cli -run '^TestManaged|^TestClaimManagedFactoryInbox' -coverprofile=…` → 관리 파일 집계 **86.9% (139/160 statements)** — 85% 목표 초과. `internal/cli` 패키지 전체 집계는 미임대 관계상 GAP(CI 담당, t1338-M1..M5 선례).
- lint: `golangci-lint run --timeout=8m ./internal/cli/...` → `0 issues.` (베이스라인 동일; 중간 4건 errcheck·ST1005 신규 발견분은 전부 수리)
- 회귀: `go test ./internal/cli -run '^(TestFactory\|TestParseLauncherEntry\|TestSD_AC021)' -count=1` → `ok … 380.550s` (어휘 가드 TestSD_AC021 포함 녹색)
- race: `go test ./internal/cli -run 'TestManagedQueue' -race` → `ok … 2.982s` (픽스처 게이트 경합 1건 발견→턴 단위 바인딩으로 구조 수리 후 재측정)
- 브로커 기준선: `go test ./internal/factorymsg -count=1` → `ok … 219.753s`

E8 (TDD RED verbatim, 구현 전 캡처):
```
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/managed_factory_session_test.go:91:9: undefined: managedTurnQueue
internal/cli/managed_factory_session_test.go:129:24: undefined: driveManagedFactorySession
internal/cli/managed_factory_session_test.go:129:117: undefined: managedFactoryInboxPrompt
internal/cli/managed_factory_session_test.go:133:39: undefined: managedPrimingPrompt
...
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
```

M1 발견·수리 이력: (1) JSON 태그 버그(`Content` 필드 `json:"content"` 누락 — 펌프 테스트가 포착), (2) 드라이버 소유자 이중 Start("exec: already started" — 실세션 소유자 테스트가 포착, spawn 책임을 소유자로 정리), (3) claim 직후 큐 턴이 다음 폴 틱까지 대기하는 구조 결함(디버그 타임스탬프 계측으로 확정 — 큐 비었을 때만 자극 대기로 수리), (4) Close 비멱등(2회 호출 이중 close/Wait), (5) 테스트 픽스처 게이트 경합(-race 감지, 턴 단위 바인딩으로 구조 제거).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

**Kickoff gate record**: Implementation Kickoff Approval PASSED in-lane (2026-10-01, operator
answered the lane session's AskUserQuestion directly — the lead's dispatch-mandated operator
form): (1) cross-host scope definition CONFIRMED as same-host launcher/backend combos,
machine-to-machine out of scope; (2) run entry approved including the gorilla/websocket v1.5.3
dependency adoption (D-2); (3) progression axis = autonomous (goal armed after this log).
Plan-audit: --deep iter1 FAIL 0.90 (2 blocking: AC-count drift, boundary-grep symbol omission)
→ repaired (5 sites, one same-class self-caught at spec.md:118) → delta re-audit PASS 0.91
(.moai/reports/t1375/plan-audit.md + plan-audit-delta.md). Skip contract note: the delta pass
re-pinned the artifact hashes, so Phase 1 computes against the new set. Baseline divergence at
run entry: `0 1` (this branch's plan commit only).

**Input parameters**: tier L; scope ~8-12 files (managed layer 2 core files + tests, launcher
wiring 3-5, doctor/approval 2); domain count 1 (factory subsystem, Go-dominant); language mix
Go + 1 go.mod line; concurrency benefit LOW (coding-heavy); agent-teams prereqs not requested.

**Mode evaluation**:

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | multi-file feature work |
| fanout | no | coding-heavy — Anthropic coding-task parallelism caveat |
| sweep | no | semantic new code, inter-file dependency |
| serial | **YES** | per-milestone manager-develop spawns, Section A-E template |

**Decision: serial**

**Justification**: coding-heavy implementation of one cohesive subsystem (managed session
layer) across milestones M1-M6 that share one package graph and one writer tree — sequential
per-milestone delegation is the simpler correct envelope. Tier L auto-routing to manager-lead
was considered (the ≥3-milestone AND ≥10-file predicate is met on paper) and declined per the
§B.2 boundary default toward the simpler mode: no cross-domain fan-out is warranted, and the
delivery-only boundary (D-1) keeps every milestone inside one surface.
