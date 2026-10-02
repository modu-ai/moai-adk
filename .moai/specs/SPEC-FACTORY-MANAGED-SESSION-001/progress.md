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

### M2 — 관리 Codex App Server 세션 + websocket 의존 (2026-10-02)

측정 기준: HEAD `242ab5a2b` + M2 작업 변경분(커밋 직전 워킹 트리). 파일:
`internal/cli/managed_codex_factory.go` (+test), `internal/config/defaults.go` 튜너블 2종
(DefaultManagedCodexReadyTimeout/TurnTimeout), `go.mod`/`go.sum` (gorilla/websocket
v1.5.3 — Kickoff 승인 의존성), `internal/factorymsg/read_body_claim_token_test.go`
(AC-MS-005 명명 테스트), `internal/cli/mcp_factory_msg_test.go` (AC-MS-011 명명 테스트).
가짜 App Server는 re-exec 패턴(`codex_launcher_exec_posix_test.go` 선례)으로
`TestManagedCodexFakeAppServer` 헬퍼가 대기 — capability 토큰 불일치는 업그레이드 전
거부되므로 핸드셰이크 성공 자체가 토큰 인증의 증거다. 헤드리스 소유자로 구현 —
TUI 부착(plan.md M3/M4 배선)은 이후 마일스톤 표면이며 M2 AC 어느 행도 요구하지 않는다.

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-MS-001 | PASS | `go test ./internal/cli -run '^TestManagedCodexAppServerHandshake$' -count=1` | `ok … 1.708s` (루프백 WS 토큰 인증 접속·/readyz 대기·initialize·thread/start·thread/name/set·모델 전달 — 가짜 서버 로그로 면 단위 검증) |
| AC-MS-002 | PASS | `go test ./internal/cli -run '^TestManagedCodexRegistersBoundPeer$' -count=1` | `ok … 2.684s` (브로커 status에 lane-1 / bound / SessionUUID=fake-thread-1 — 런처가 BindLaunchPending으로 봉인) |
| AC-MS-003 | PASS | `go test ./internal/cli -run '^TestManagedLaunchPendingRollback$' -count=1` | `ok …` (M1 커밋 증거 재측정 — 기동 실패 후 launch-pending 잔행 0건) |
| AC-MS-004 | PASS | `go test ./internal/cli -run '^TestManagedInboxPromptMetadataOnly$' -count=1` | `ok …` (메타데이터만 주입·본문 부재 재측정) |
| AC-MS-005 | PASS | `go test ./internal/factorymsg -run '^TestReadBodyClaimToken$' -count=1` | `ok … 0.448s` (무토큰·위토큰 거부, 토큰만 본문 반환, 정산 후 재판정 불가 — store 기존 능력의 특성 테스트, 신규 store 코드 아님) |
| AC-MS-006 | PASS | `go test ./internal/cli -run '^TestManagedReceiptAcknowledges$' -count=1` | `ok …` (acknowledged 증가·pending 감소 재측정) |
| AC-MS-009 | PASS | `grep -rnE 'fmt\.Sprintf\("(agent\|worker)-' internal/cli/managed_*.go` | 0행 (exit 1) |
| AC-MS-010 | PASS | `grep -rnE 'merge-window\|Decider\|T29b\|T29c\|handover' internal/cli/managed_*.go` | 0행 (exit 1) |
| AC-MS-011 | PASS | `go test ./internal/cli -run '^TestFactoryMsgSendRejectsClaudeOnlyRun$' -count=1` | `ok … 2.021s` (SPEC 명명 테스트가 현재 트리에 없어 신설 — 거부 팔 + 등록 엔드포인트 소유자 성공 팔의 양면 구성) |
| AC-MS-014 | PASS | `GOOS=windows GOARCH=amd64 go build ./...` + `grep -rn 'syscall\.' internal/cli/managed_*.go` | `WIN=0` + 0행 (exit 1 — 주석의 리터럴까지 패러프레이즈 수리) |
| AC-MS-016 | PASS | `go test ./internal/cli -run '^TestManagedCodexFactoryBrokerLive$' -count=1 -v` | `--- SKIP: TestManagedCodexFactoryBrokerLive (0.00s)` (live 게이트 env 미설정 → skip — 합의된 CI 포지) |
| AC-MS-017 | PASS | `git log --oneline $(git merge-base develop HEAD)..HEAD -- internal/factorymsg/store.go` + `git diff --stat internal/factorymsg/store.go` | 빈 출력 ×2 (merge-base = `f22e2d7ac` — gitflow-lane-protocol §8 재산출 일치) |

품질 게이트 (M2):
- 커버리지: `go test ./internal/cli -run '^TestManaged' -coverprofile=…` → 관리 파일 2종
  집계 **85.2% (327/384 statements)** — 85% 목표 초과. `internal/cli` 패키지 전체
  집계는 미임대 관계상 GAP(CI 담당 — t1338/M1 선례). 미달 84.9%→85.2% 구간은
  `--model=` 빈 값 거부 1행 추가로 해소.
- lint: `golangci-lint run --timeout=8m ./internal/cli/... ./internal/config/... ./internal/factorymsg/...` → `0 issues.` (베이스라인 동일; 1회 "parallel golangci-lint is running" 점유 재시도 후 측정)
- race: `go test ./internal/cli -run 'TestManagedCodex' -race -count=1` → `ok … 9.178s` (data race 0)
- gofmt: 변경 파일 4종 `-l` 빈 출력
- 어휘 가드: `go test ./internal/cli -run '^TestSD_AC021' -count=1` → `ok … 1.274s`

E8 (TDD RED verbatim, 구현 전 캡처):
```
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/managed_codex_factory_test.go:190:12: undefined: newManagedCodexSession
internal/cli/managed_codex_factory_test.go:252:23: undefined: runManagedFactoryCodex
internal/cli/managed_codex_factory_test.go:300:13: undefined: managedLoopbackURL
internal/cli/managed_codex_factory_test.go:300:54: undefined: errManagedCodexNonLoopback
internal/cli/managed_codex_factory_test.go:305:13: undefined: managedLoopbackURL
internal/cli/managed_codex_factory_test.go:312:18: undefined: managedCodexDial
internal/cli/managed_codex_factory_test.go:312:79: undefined: errManagedCodexNonLoopback
internal/cli/managed_codex_factory_test.go:318:25: undefined: managedCodexOptions
internal/cli/managed_codex_factory_test.go:328:22: undefined: managedCodexOptions
internal/cli/managed_codex_factory_test.go:331:18: undefined: managedCodexOptions
internal/cli/managed_codex_factory_test.go:331:18: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
```
(AC-MS-005·011의 명명 테스트는 기존 능력의 특성 테스트 — 작성 시점에 이미 GREEN이며
RED 단계가 없다. 이는 TDD 위반이 아니라 계측 대상이 선행 능력이라는 뜻이다.)

M2 발견·수리 이력: (1) `managedCodexOptions`가 args[0](프로그램명)을 스캔 —
M1 규약(args[0] 미스캔) 위반으로 옵션 테스트가 포착, `i := 1` 시작으로 수리.
(2) spawn 실패 시 토큰 임시 디렉터리 누수 — Start의 spawn 실패 분기에서 즉시 제거.
(3) 주석의 `syscall.Exec` 리터럴이 AC-MS-014 기계 grep을 오염 — 패러프레이즈로 수리
(기계 판정형 AC는 주석 예외를 두지 않는다). (4) 진입점의 옵션 파싱을 브로커 등록
앞으로 이동 — 연산자 오류 거부가 상태를 남기지 않는다. (5) 세션 id 우선 경로가
시험 환경의 권위 세션 id를 집어 PID 경로를 가림 — `EnvClaudeCodeSessionID` 비움
(`TestFactoryMCPIdentityAttribution` 선례).

### M3 — launcher wiring (2026-10-02)

측정 기준: HEAD `1d17928a0`, 클린 트리. M3는 컨텍스트가 소진된 선행 워커가 남긴 미커밋 상태에서
복구되어 `1d17928a0`로 커밋되었고, 레인이 독립 검수 후 이번 실행에서 재측정했다 (아래 행은
전부 이 트리·이 실행의 관측값).

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| M3 launcher divert | PASS | `go test -count=1 -v -run 'TestManaged(Launch\|CodexLaunch)\|TestLaunchWithoutFactoryEnv\|TestCodexLaunchWithoutFactoryEnv' ./internal/cli/` | `--- PASS` 10건, `ok … 1.998s` (exit 0) |
| AC-MS-011 + `--` pass-through (R2) | PASS | `go test -count=1 -v -run '^TestFactoryMsgSendRejectsClaudeOnlyRun$\|TestParseCompanionLabelStopsAtPassThroughMarker' ./internal/cli/` | 두 테스트 `--- PASS`, `ok … 1.453s` (exit 0) |
| AC-MS-017 | PASS | `git diff 1d17928a0 -- internal/factorymsg/store.go \| wc -l` / `git diff f22e2d7ac..HEAD -- internal/factorymsg/store.go \| wc -l` | `0` / `0` |

### M4 — MCP approval scoping + doctor (2026-10-02)

측정 기준: HEAD `1d17928a0` + M4 작업 변경분(커밋 직전 워킹 트리). 파일:
`internal/cli/managed_codex_factory.go` (`factoryMoAIMCPApprovalArgs` + `managedCodexAppServerArgs`
추출, `Start`가 후자 사용), `internal/codexwiring/configtoml.go` (`StaleApprovalOverride`,
읽기 전용), `internal/cli/doctor_codex.go` (기존 "Codex Wiring" 점검에 finding 1종 추가 —
신규 점검이 아니므로 binary_lag 허용목록 변경 없음), 테스트
`internal/cli/managed_codex_approval_test.go`, `internal/codexwiring/configtoml_test.go`.
승인 인수는 소유 App Server 명령행에만 붙고 연산자 `-c` 인수가 뒤에 와서 우선한다.
TUI 부착은 M2와 마찬가지로 헤드리스 소유자라 이 트리에 소유 TUI 프로세스가 없다 — TUI 쪽
인수 전달은 TUI 소유 표면이 생기는 시점의 몫이며, 헬퍼는 그 경로가 재사용할 수 있게 분리했다.

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-MS-012 | PASS | `go test -count=1 -run '^(TestMoAIMCPApprovalArgsOnlyTargetMoAI\|TestConfigTomlWritesUnchanged\|TestDoctorCodexWarnsStaleGlobalApproval\|TestStaleApprovalOverride)$' ./internal/cli/ ./internal/codexwiring/` | `ok … internal/cli 1.268s` / `ok … internal/codexwiring 0.897s` (exit 0); project config 생성물 `default_tools_approval_mode = "writes"` 불변·`approve`/`factory_msg` 부재 단언 |
| AC-MS-013 | PASS | 위 동일 실행의 `TestDoctorCodexWarnsStaleGlobalApproval` | `--- PASS`: CheckWarn + 메시지에 approval·approve 명명 + `.codex/config.toml` 전후 바이트 동일 단언; 대조군(정규 config)은 approval finding 없음 |
| AC-MS-009 | PASS | `grep -rnE 'fmt\.Sprintf\("(agent\|worker)-' internal/cli/managed_*.go \| wc -l` | `0` |
| AC-MS-010 | PASS | M4 diff는 인수 빌더·읽기 전용 검사기·doctor finding뿐 (merge 자동화·Decider·T29b/c 코드 없음) | diff 검토 |
| AC-MS-014 | PASS | `GOOS=windows GOARCH=amd64 go build ./...` + `grep -c 'syscall\.' internal/cli/managed_*.go` | `win_exit=0`, 신규 `syscall.` 0건 |
| AC-MS-017 | PASS | `git diff 1d17928a0 -- internal/factorymsg/store.go \| wc -l` / `git diff f22e2d7ac..HEAD -- internal/factorymsg/store.go \| wc -l` | `0` / `0` |

품질 게이트 (M4):
- 커버리지: 신규 함수 `factoryMoAIMCPApprovalArgs` 100.0%, `managedCodexAppServerArgs` 100.0%,
  `StaleApprovalOverride` 93.3% (M4 테스트 한정 coverprofile). 패키지 전체 집계는 미임대
  관계상 GAP (CI/리드 슬롯 몫).
- codexwiring 패키지 전체: `go test -count=1 ./internal/codexwiring/` → `ok … 1.589s` (생성기 출력 불변).
- doctor 회귀: `go test -count=1 -run 'CheckCodexWiring|DoctorCodex|TestBinaryLag' ./internal/cli/` → `ok … 3.666s`.
- vet: `go vet ./internal/cli/ ./internal/codexwiring/` → exit 0.
- lint: `golangci-lint run --timeout=8m ./internal/cli/... ./internal/codexwiring/...` → `0 issues.`
- gofmt: 변경 파일 5종 `-l` 빈 출력.
- go.mod/go.sum: 변경 0.

E8 (TDD RED verbatim, 구현 전 캡처):
```
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/managed_codex_approval_test.go:24:10: undefined: factoryMoAIMCPApprovalArgs
internal/cli/managed_codex_approval_test.go:53:12: undefined: managedCodexAppServerArgs
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
```

M4 발견 이력: (1) 신규 doctor 점검이 아니라 기존 "Codex Wiring" 점검의 finding으로 구현 —
binary_lag 허용목록/TestBinaryLag 변경 불필요(`TestBinaryLag` 회귀 통과). (2) 기존 비정규
테이블 finding과 새 finding이 같은 config에서 함께 뜬다(중복이 아니라 서로 다른 사실 —
하나는 형태 드리프트, 하나는 잔재 지목). (3) M3 커밋의
`internal/cli/managed_factory_session_test.go`에 gofmt 지적(한 줄 goroutine 람다) 기존 존재 —
M4 범위 밖이라 미수정, 리드 판단 대기.

### M5 — loopback integration + regression + docs (2026-10-02)

측정 기준: HEAD `cea8c5a0b` (M4 tip), 클린 트리에서 시작. 파일:
`internal/cli/managed_loopback_test.go` (신규, `TestManagedSessionLoopbackRoundTrip` + 헬퍼
`runManagedLoopback`), `.moai/docs/factory-managed-session.md` (신규 한국어 안내). 프로덕션
코드 변경 없음. 루프백 테스트는 실제 `driveManagedFactorySession` 루프와 실제
`claimManagedFactoryInbox`/`managedFactoryInboxPrompt`를 임시 브로커 위에서 돌리고, 가짜 모델
세션이 프롬프트의 `message_id`/`claim_token`으로 `ReadBody` → `RecordDisposition` → `Receipt`를
수행한다. 2팔 대조: claim을 건너뛰는 변이, 프롬프트에 본문을 주입하는 변이가 각각 위반 목록으로
잡힌다.

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-MS-015 | PASS | `go test -count=1 -v -run '^TestManagedSessionLoopbackRoundTrip$' ./internal/cli` | real_path / mutant_skipping_the_claim / mutant_injecting_the_body 3 서브테스트 `--- PASS`, `ok … 2.541s`; `-race -count=5` → `ok … 12.029s` |
| AC-MS-016 | PASS | `go test -count=1 -run '^TestManagedCodexFactoryBrokerLive$' -v ./internal/cli` | `--- SKIP: TestManagedCodexFactoryBrokerLive (0.00s)` (live gate env not set), exit 0; M2 테스트 수정 없음 |
| AC-MS-009 | PASS | `grep -rnE 'fmt\.Sprintf\("(agent\|worker)-' internal/cli/managed_*.go \| wc -l` | `0` |
| AC-MS-014 | PASS | `grep -n 'syscall\.' internal/cli/managed_*.go \| wc -l` + `GOOS=windows GOARCH=amd64 go build ./...` | `0`, `win_exit=0` |
| AC-MS-017 | PASS | `git diff f22e2d7ac..HEAD -- internal/factorymsg/store.go \| wc -l` | `0` |

회귀 슬라이스 (모두 이 트리·이 실행):
- `go test -count=1 -run 'Factory|Managed|Kanban' ./internal/cli` (레인 env 스크럽 `unset … &&` 단일 호출) → `ok … 230.947s`, FAIL 0.
- `go test -count=1 ./internal/kanban` → `ok … 236.497s`.
- `go test -count=1 -run 'Factory' ./internal/hook` → `ok … 90.121s`.
- `go test -count=1 -run 'Parity|Mirror' ./internal/template` → `ok … 1.363s`.

품질 게이트 (M5):
- 커버리지 (managed 테스트 한정 coverprofile `-run '^TestManaged|^TestClaimManagedFactoryInbox|^TestMoAIMCP'`): `managed_factory_session.go` 84.8% (139/164), `managed_codex_factory.go` 85.0% (193/227), 합 84.9% (332/391 문장). 패키지 전체 집계는 GAP (CI/리드 슬롯 몫).
- vet: `go vet ./internal/cli` → exit 0. lint: `golangci-lint run --timeout=8m ./internal/cli/...` → `0 issues.`
- gofmt: 신규 `managed_loopback_test.go` 는 `gofmt -l` 빈 출력. `gofmt -l internal/cli/managed_*.go` 는 M3 커밋의 `managed_factory_session_test.go` 1건만 지목 (M4 기록과 동일한 기존 항목, M5 범위 밖이라 미수정).
- go.mod/go.sum: 변경 0.

E8 (TDD RED verbatim, 헬퍼 구현 전 캡처):
```
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/managed_loopback_test.go:21:10: undefined: runManagedLoopback
internal/cli/managed_loopback_test.go:28:10: undefined: runManagedLoopback
internal/cli/managed_loopback_test.go:37:10: undefined: runManagedLoopback
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
```

M5 발견 이력: (1) 레인 env(`MOAI_KANBAN_*`/`MOAI_FACTORY_*`)가 셸에 남아 있으면 `Factory|Managed|Kanban`
슬라이스에서 10건(`TestFactoryNext*`, `TestFR_AC024/025`, `TestTodoPickInFactory*` 등)이 거짓 적색으로
뜬다 — env 스크럽 후 FAIL 0 (기존 교훈 `lane env falsifies env-reading guard tests`와 동일). (2) 가짜
세션은 인프로세스다 — 재실행(re-exec) 자식은 브로커 호출 도구가 없어 왕복 증명에 쓸 수 없었고, 스트림 세션
fixture(sh)는 M1 테스트가 이미 다룬다. (3) 안내 문서 위치: `.moai/docs/`는 `todo-queue-storage.md` 등
운영자 안내 선례가 있는 디렉터리이며 이 문서는 템플릿 미러(`internal/template/templates/.moai/docs/`)
밖의 로컬 문서다(미러에는 6종만 존재). `.claude/rules`·`.claude/skills`는 건드리지 않았다.

M5 후속 (2026-10-02, 기준 HEAD `87d4990bb`, 클린 트리에서 시작):
- AC-MS-015 re-exec leg: `TestManagedSessionLoopbackRoundTrip`가 `in-process` / `re-exec` 두 세션 종류 × (real path, claim 건너뛰기 변이, 본문 주입 변이)로 확장됨. re-exec leg는 테스트 바이너리를 `TestManagedLoopbackChild`로 재실행(sh 래퍼 → `os.Args[0] -test.run=…`)해 실제 `managedStreamSession` 소유 하에 stream-json 자식으로 띄운다. 자식이 stdin의 메타데이터 프롬프트를 읽고 같은 디스크 브로커 저장소를 `factorymsg.Open`으로 열어 본문 조회·disposition·영수증을 직접 수행하며(`store.go` 변경 없음), 부모가 acknowledged 1 / pending 0 과 프롬프트의 본문 부재를 단언한다. 출력: `--- PASS: …/re-exec (2.33s)` 하위 3건 + in-process 3건, `ok … 4.950s`; `-race -count=3` → `ok … 24.757s`.
- RED verbatim (자식 지정을 없는 테스트명으로 바꾼 임시 변이, 확인 후 원복): `--- FAIL: …/re-exec/real_path`: `round trip violated: [session error: managed session output closed no inbox prompt was injected body lookup returned "" acknowledged=0 pending=1, want 1/0]`.
- 커버리지: `go test -count=1 -run '^TestManaged|^TestClaimManagedFactoryInbox|^TestMoAIMCP' -coverprofile=… ./internal/cli` → `managed_factory_session.go` 90.9% (149/164), `managed_codex_factory.go` 85.9% (195/227), 합 88.0% (344/391). 추가 파일 `managed_failure_paths_test.go`: 스트림 플래그 탈취 거부(생성 시점), Start 전 DeliverTurn 거부, pump의 쓰기 실패·stdout 실패·error result·oversize 라인, 프라이밍 실패, stdin EOF 후 폴링 지속, 런 없는 launch 거부(Claude/GLM), 소유자의 `--print` 거부, Codex readiness 루프백 가드·취소 컨텍스트.
- 불변식 재측정: store.go diff 0, vocabulary grep 0, `syscall.` 0, `win_exit=0`, go.mod/go.sum 변경 0, AC-MS-011(`TestFactoryMsgSendRejectsClaudeOnlyRun`)과 `--` 통과 테스트 PASS. lint `0 issues.`, vet exit 0.
- 기존 gofmt 지적: `internal/cli/managed_factory_session_test.go` (M1 커밋 `242ab5a2b`, 한 줄 goroutine 람다) — 이 후속에서도 미수정.
- 위 M5 본문의 "가짜 세션은 인프로세스" 발견 (2)와 gap 서술은 이 후속으로 대체됨: re-exec leg가 추가됐고 커버리지 gap은 닫혔다.

### M6 — run-phase closing matrix (2026-10-02)

측정 기준: HEAD `f8382647c`, 클린 트리, 카드 base `f22e2d7ac`. 모든 명령은 `for v in $(env | grep -E '^(MOAI_|CLAUDE_CODE_)' | cut -d= -f1); do unset $v; done &&` 로 레인 env를 전부 제거한 뒤 이 트리에서 이번 실행에 돌렸다(`go test -count=1 -v`의 최상위 `--- ` 줄 인용).

| AC | Status | Command | Verbatim result |
|----|--------|---------|-----------------|
| AC-MS-001 | PASS | `go test ./internal/cli -run '^TestManagedCodexAppServerHandshake$'` | `--- PASS: TestManagedCodexAppServerHandshake (0.31s)` |
| AC-MS-002 | PASS | `go test ./internal/cli -run '^TestManagedCodexRegistersBoundPeer$'` | `--- PASS: TestManagedCodexRegistersBoundPeer (0.76s)` |
| AC-MS-003 | PASS | `go test ./internal/cli -run '^TestManagedLaunchPendingRollback$'` | `--- PASS: TestManagedLaunchPendingRollback (0.85s)` |
| AC-MS-004 | PASS | `go test ./internal/cli -run '^TestManagedInboxPromptMetadataOnly$'` | `--- PASS: TestManagedInboxPromptMetadataOnly (0.41s)` |
| AC-MS-005 | PASS | `go test ./internal/factorymsg -run '^TestReadBodyClaimToken$'` | `--- PASS: TestReadBodyClaimToken (0.07s)` |
| AC-MS-006 | PASS | `go test ./internal/cli -run '^TestManagedReceiptAcknowledges$'` | `--- PASS: TestManagedReceiptAcknowledges (0.61s)` |
| AC-MS-007 | PASS | `go test ./internal/cli -run '^TestManagedSessionOwnsStreamFlags$'` | `--- PASS: TestManagedSessionOwnsStreamFlags (0.00s)` |
| AC-MS-008 | PASS | `go test ./internal/cli -run '^TestManagedQueueSerializesOperatorAndInbox$'` | `--- PASS: TestManagedQueueSerializesOperatorAndInbox (0.00s)` |
| AC-MS-009 | PASS | `grep -rnE 'fmt\.Sprintf\("(agent\|worker)-' internal/cli/managed_*.go \| wc -l` | `0` |
| AC-MS-010 | PASS | `grep -rnE 'merge-window\|Decider\|T29b\|T29c\|handover' internal/cli/managed_*.go \| wc -l` | `0` |
| AC-MS-011 | PASS | `go test ./internal/cli -run '^TestFactoryMsgSendRejectsClaudeOnlyRun$'` | `--- PASS: TestFactoryMsgSendRejectsClaudeOnlyRun (0.57s)` |
| AC-MS-012 | PASS | `go test ./internal/codexwiring ./internal/cli -run '^(TestMoAIMCPApprovalArgsOnlyTargetMoAI\|TestConfigTomlWritesUnchanged)$'` | `--- PASS: TestMoAIMCPApprovalArgsOnlyTargetMoAI (0.00s)`, `--- PASS: TestConfigTomlWritesUnchanged (0.00s)` (codexwiring 패키지는 `[no tests to run]` — 두 테스트는 cli 소속) |
| AC-MS-013 | PASS | `go test ./internal/cli -run '^TestDoctorCodexWarnsStaleGlobalApproval$'` | `--- PASS: TestDoctorCodexWarnsStaleGlobalApproval (0.01s)` |
| AC-MS-014 | PASS | `GOOS=windows GOARCH=amd64 go build ./...` + `grep -rn 'syscall\.' internal/cli/managed_*.go \| wc -l` | `win_exit=0`, `0` |
| AC-MS-015 | PASS | `go test ./internal/cli -run '^TestManagedSessionLoopbackRoundTrip$'` | `--- PASS: TestManagedSessionLoopbackRoundTrip (3.20s)` (in-process + re-exec 각 3팔) |
| AC-MS-016 | PASS (SKIP) | `go test -v ./internal/cli -run '^TestManagedCodexFactoryBrokerLive$'` | `--- SKIP: TestManagedCodexFactoryBrokerLive (0.00s)`, `live gate env (MOAI_FACTORY_LIVE_ROOT / MOAI_FACTORY_LIVE_RUN) not set`, exit 0 — 실세션 왕복은 GAP(아래) |
| AC-MS-017 | PASS | `git log --oneline f22e2d7ac..HEAD -- internal/factorymsg/store.go` + `git diff f22e2d7ac..HEAD -- internal/factorymsg/store.go \| wc -l` | 빈 출력, `0` |
| 회귀 (acceptance §3) | existing red (owner card named) | 레인 측정 — 결정 줄은 아래 "회귀 증거" 소절(이 파일), 디스크 사본 `.moai/reports/t1375/full-suite-evidence.md`(gitignore 대상 — 저장소 지침상 보고서는 원격에 올리지 않음) | 클린 env `./internal/cli` 전체 `FAIL … 2236.779s`, 고유 FAIL 이름 정확히 2개(`TestStopChainEffectParityGolden`, `TestSyncGateLanguageDetectionMatchesScript`), base `f22e2d7ac` 사본에서도 같은 2건. 소유 카드: t1390 / t1402 (리더 메시지 귀속, 본 에이전트 관측 아님) |

go.mod/go.sum: base 대비 변경은 M2의 승인된 `github.com/gorilla/websocket v1.5.3` 추가(`033221ced`)와 `jsonschema/v6`의 indirect 표기 해제뿐이다(`git log --oneline f22e2d7ac..HEAD -- go.mod go.sum` → `033221ced`). M3 이후 변경 없음.

`go run ./cmd/moai spec lint SPEC-FACTORY-MANAGED-SESSION-001` → `✓ No findings — all SPEC documents are valid`.

### 회귀 증거 (측정 주체: 레인 — manager-develop 미실행, 레인 scratch `/private/tmp/claude-501/t1375/`의 결정 줄을 M6에서 재독해 옮김)

`.moai/reports/*`는 gitignore(저장소 지침: 보고서는 원격 비전송)라 전체 서술은 로컬 디스크 `.moai/reports/t1375/full-suite-evidence.md`에만 있고, 커밋되는 인용 대상은 이 소절이다.

| 실행 | 범위·조건 | 관측 (scratch 파일) |
|---|---|---|
| (a) 1차 | `./internal/cli ./internal/codexwiring ./internal/factorymsg`, MOAI_KANBAN 계열 5종만 unset, 타 레인(go-test-cli-t1391)이 slot lease 보유 중인 부하 상태 | `full.txt`: `panic: test timed out after 30m0s`, `FAIL … internal/cli 1801.163s`, `ok … codexwiring 1.823s`, `ok … factorymsg 86.977s`, 고유 `--- FAIL` 254개(`failed.txt` 254행). 누수 env 이름(MOAI_FACTORY_ROLE/WORKER/WORKERS/AUTO_DISPATCH/CLEAR_POLICY, MOAI_AUTONOMY_TIER, MOAI_KANBAN_BACKEND, MOAI_PROJECT_DIR)은 레인 보고 — scratch에서 이름 목록은 미확인 |
| (b) 254개 재실행 | 모든 `MOAI_*`·`CLAUDE_CODE_*` 제거 | `rerun-failed.txt`: `FAIL … internal/cli 644.130s`, 고유 FAIL 2개(`still-failing.txt`), timeout panic 0 → 252 통과 |
| (c) 클린 env 전체 | `./internal/cli` | `full2.txt`: `FAIL … internal/cli 2236.779s`, `FULL2_EXIT=1`, 고유 FAIL = 아래 2개, timeout panic 0. `codex_sync_gate_test.go:115: languages: Go = [kotlin], script = [kotlin java]`; `codex_stop_chain_golden_test.go:146/160/180` (`Claude path decision = allow, want deny (the golden's premise does not hold)`, `premise: the first Claude run must block, got allow`) |
| (d) base 사본 | `f22e2d7ac` git archive(워크트리 아님), 클린 env, 같은 2개 테스트 | `base-two.txt`: `--- FAIL: TestStopChainEffectParityGolden (38.38s)`, `--- FAIL: TestSyncGateLanguageDetectionMatchesScript (0.86s)`, 같은 메시지, `FAIL … internal/cli 40.241s` |

결론: 252건은 레인 env 누수(측정된 원인), 남은 2건은 카드 base에도 있는 기존 적색. 소유 카드는 리더가 레인에 보낸 메시지의 귀속이며 본 에이전트의 관측이 아니다: `TestStopChainEffectParityGolden` = t1390(리더 로컬 develop `f3883ddb1`에서 수리·병합, 미push라 이 카드 base에 없음), `TestSyncGateLanguageDetectionMatchesScript` = t1402(발행됨, 의미 소유자는 그 plan에서 결정). 카드 변경 집합은 hook·gate·stop-chain 파일을 건드리지 않는다(`git diff f22e2d7ac..HEAD --name-only | grep -E '\.claude/hooks|sync-phase|stop|handle-'` → 0줄).

## §E.3 Run-phase Audit-Ready Signal

- run_status: run-complete-pending-audit (측정 기준 HEAD `f8382647c`, 이번 실행)
- 커버리지: `go test -count=1 -run '^TestManaged|^TestClaimManagedFactoryInbox|^TestMoAIMCP' -coverprofile=… ./internal/cli` → `managed_factory_session.go` 90.9% (149/164), `managed_codex_factory.go` 85.9% (195/227), 합 88.0% (344/391) — 목표 85% 충족. 패키지 전체 집계는 이 목록에 없다(GAP).
- vet: `go vet ./internal/cli ./internal/codexwiring ./internal/factorymsg ./internal/config` → exit 0.
- gofmt: `git diff f22e2d7ac..HEAD --name-only` 의 Go 파일 16개에 `gofmt -l` → 빈 출력.
- lint: `golangci-lint run --timeout=8m ./internal/cli/... ./internal/codexwiring/ ./internal/factorymsg/ ./internal/config/` → exit 0, `0 issues.`
- 비침범 grep: AC-MS-009/010/014/017 행(위 매트릭스) 전부 0. `git diff f22e2d7ac..HEAD --name-only | grep -E '\.claude/hooks|sync-phase|stop|handle-' | wc -l` → `0`.
- diff 범위: `git diff f22e2d7ac..HEAD --stat | tail -3` → `25 files changed, 4269 insertions(+), 2 deletions(-)`.
- 회귀 증거(측정 주체 = 레인, 본 에이전트 미실행): 아래 "회귀 증거" 소절 — 252건은 레인 env 누수(측정된 원인), 2건은 base에서도 실패하는 기존 적색(소유 t1390·t1402는 리더 메시지 귀속).
- GAP: (1) golangci-lint는 로컬 바이너리이며 CI(v2.1.6) 판 일치는 미확인. (2) 실제 Codex 왕복 미관측(live env 미설정). (3) `internal/kanban`·`internal/hook`·`internal/template` 스위트는 M5에 기록한 범위 한정 슬라이스로만 측정했고 M6에서 재실행하지 않았다. (4) 레인의 1차 전체 실행(a)은 다른 레인(go-test-cli-t1391)이 자기 slot lease를 쥔 부하 상태에서 돌았다. (5) 누수 env 이름 목록은 레인 보고이며 scratch에서 확인하지 못했다.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-10-02
sync_commit_sha: 578e0d8896a6238d9d110aa55cf702f85d11446e  # D3-exempt backfill: the placeholder in the sync commit replaced with the real SHA in a following commit (phase-owned field, manager-docs §E.4)
sync_status: complete
b12_self_test_a: pass  # pre-emission `grep -c 'SPEC-FACTORY-MANAGED-SESSION-001' CHANGELOG.md` = 0 before this sync commit (duplicate-entry guard)
b12_self_test_b: pass  # AC count: MOAI-AC-COUNTER against acceptance.md (tier L source) → live=17 excluded=0 ambiguous=0; the CHANGELOG entry cites the same 17 (AC-MS-001..017); milestones 6 (M1..M6, per the run commit list); managed-test coverage 88.0% (344/391) quoted from §E.3, not re-measured here
b12_self_test_c: pass  # file-path verification: all 9 implementation/doc paths cited in the CHANGELOG entry confirmed via ls against this tree (ls exit 0)
changelog_entry_position: "[Unreleased] > Added (first entry)"
frontmatter_status_transitions:
  in_progress_to_implemented: merged  # folded into the single sync commit per the 3-phase close
  implemented_to_completed: this commit  # spec.md frontmatter status+updated only; plan/acceptance/design/research are stateless on the status axis and carry no status field
canary_compliance_check:
  spec_body_untouched: true  # spec/plan/acceptance/design/research bodies byte-unchanged vs 677cb6919 (see Verification, diff below)
  runtime_files_untouched: true  # .moai/state, .moai/harness, .moai/cache untouched
recorded_by: manager-docs (sync phase, card t1375)
```

### What was synced

- `CHANGELOG.md` `[Unreleased] > Added`: one entry for the managed cross-host session layer. It states the one new dependency `github.com/gorilla/websocket v1.5.3` (`go.mod:37`, loopback App Server use only) and the delivery-only boundary (no merge automation, no card completion judgment).
- `spec.md` frontmatter: `status: in-progress → completed`, `updated: 2026-10-01 → 2026-10-02`.
- README / docs-site: no edit. The only docs-site surface naming the broker is `docs-site/content/en/guides/mcp-server.md:220-224` (and its ja/zh/ko twins), which lists the five `factory_msg_*` MCP tools; this card did not change that tool surface (`git diff f22e2d7ac..HEAD --name-only | grep -E 'mcp_server|docs-site|README' | wc -l` → `0`). The operator-facing guide `.moai/docs/factory-managed-session.md` already landed in M5 and is linked from the CHANGELOG entry, not duplicated.
- Codemaps refresh: not performed in this sync commit (outside the dispatched deliverables); listed under Gaps.

### Verification (this run, this tree: HEAD `677cb6919` + uncommitted sync edits; lane env scrubbed in the same compound call as each command)

| Check | Command | Verbatim result |
|---|---|---|
| SPEC lint | `go run ./cmd/moai spec lint SPEC-FACTORY-MANAGED-SESSION-001` | `✓ No findings — all SPEC documents are valid` |
| AC counter (B12) | MOAI-AC-COUNTER awk from `manager-docs.md` § B12, `AC_FILE=.moai/specs/SPEC-FACTORY-MANAGED-SESSION-001/acceptance.md` | stdout `17`, stderr `live=17 excluded=0 ambiguous=0`, exit 0 |
| Duplicate guard | `grep -c 'SPEC-FACTORY-MANAGED-SESSION-001' CHANGELOG.md` (before the edit) | `0` |
| SPEC-corpus test slice | `go test -count=1 -run '^(TestACCounterFullCorpusMatchesBaseline\|TestACCounterCorpusMutantIsDetected)$' -v ./internal/spec` | `--- PASS: TestACCounterFullCorpusMatchesBaseline (8.33s)`, `--- PASS: TestACCounterCorpusMutantIsDetected (0.01s)`, `ok  github.com/modu-ai/moai-adk/internal/spec 8.794s` |
| SPEC audit | `go run ./cmd/moai spec audit --json`, entries naming this SPEC | one entry: `finding_type: EraAutoDetected`, `severity: INFO`, `era: V3R6`, `heuristic_matched: H-4 (§E.2 + §E.4 + sync_commit_sha)`; no drift finding for this SPEC in the filtered output (`grep -B2 -A8` window, not a full-report parse) |
| store.go untouched | `git diff f22e2d7ac..HEAD -- internal/factorymsg/store.go \| wc -l` | `0` |
| Body untouched | `git diff 677cb6919 -- .moai/specs/SPEC-FACTORY-MANAGED-SESSION-001/<file>.md \| grep -E '^[-+]'` for plan/acceptance/design/research | empty for all four; for spec.md only `-status: in-progress` / `+status: completed` / `-updated: 2026-10-01` / `+updated: 2026-10-02` (plus the two `---`/`+++` header lines) |

### Gaps

- The live Codex round trip is unobserved: `TestManagedCodexFactoryBrokerLive` skips without `MOAI_FACTORY_LIVE_ROOT` / `MOAI_FACTORY_LIVE_RUN` (AC-MS-016 passes as a SKIP, per §E.2).
- The CI-parity lint version is unchecked: the `golangci-lint` figures in §E.3 come from the local binary, and its match to the CI v2.1.6 build was not verified. No lint was run in this sync phase.
- The two reds in the full `./internal/cli` run, `TestStopChainEffectParityGolden` (owner card t1390) and `TestSyncGateLanguageDetectionMatchesScript` (owner card t1402), are recorded in the 회귀 증거 subsection; their ownership comes from the leader's message and is not a sync-phase observation. No full-package suite was run in this sync phase.
- Codemaps (`.moai/project/codemaps/`) were not refreshed in this sync commit.
- The SPEC-corpus slice above is the AC-counter corpus test; no test dedicated to `status:` frontmatter of the real `.moai/specs` tree was found by grep, so `spec lint` on this SPEC is the status/frontmatter validity evidence.
- `sync_commit_sha` is a placeholder until the backfill commit lands.

### Residual-risk

- MCP approval scoping and loopback guards are covered by in-repo tests only; behavior against a real Codex App Server is unobserved (see Gaps).
- The coverage figure (88.0%) covers the managed test selection only, not the whole `internal/cli` package.
- A concurrent session pushing a competing change to the CHANGELOG `[Unreleased] > Added` block would conflict textually with the first-entry insertion; resolve at integration.

sync_status: audit-ready — this signal is ready for the independent sync audit (sync-auditor).

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

## §G Audit-repair delta 1 (2026-10-02, 기준 HEAD `623e4b15a`)

sync-audit FAIL 69의 F1·F6·F7·F10 대응(코드·테스트 한정). 이번 실행의 측정, 레인 env 전부 제거.

- F1 수리: `managedFactoryCodexLaunchFunc`/`runManagedFactoryCodex`에 `dir` 인수 추가, divert가 `req.Dir`을 넘기고(`req :=` 줄은 불변), 소유자가 App Server `cmd.Dir`과 `thread/start` cwd에 사용(빈 값이면 기존처럼 프로세스 cwd). Claude/GLM 시임은 같은 결함이 없다: 일반 Claude 런치도 `cmd.Dir`/Chdir 없이 프로세스 cwd로 exec하고(`launcher.go`에 launch Dir 설정 없음), `-w`는 claude 자신이 해소한다.
  - RED(구현 전, 빌드 실패): `managed_codex_factory_test.go:283:83: too many arguments in call to runManagedFactoryCodex … want (string, []string, []string, io.Reader)`.
  - GREEN: `TestManagedCodexLaunchCarriesLaunchDir`(하위 디렉터리 cwd→프로젝트 루트, `-w card`→워크트리 = 앵커 락 디렉터리), `TestManagedCodexOwnerUsesLaunchDir`(가짜 App Server가 `server-cwd`와 `thread-cwd`를 기록, 테스트 프로세스 cwd는 다른 디렉터리) PASS.
  - 변이(`go test -overlay`, 저장소 미변경): divert가 `""` 전달 / `cmd.Dir` 제거 / thread cwd가 dir 무시 / `session.dir = dir` 제거 → 각각 해당 테스트 `--- FAIL`.
- F6: 프로덕션 claim/prompt 배선을 이름 있는 `managedFactoryInboxWiring`으로 추출(두 소유자가 사용), 루프백 real arm이 이를 직접 사용. 변이 claim `return nil, nil` / 프롬프트에 본문 주입 → `TestManagedSessionLoopbackRoundTrip --- FAIL`(overlay). 잔여: 소유자 진입점의 호출 줄 자체(`claim, toPrompt := managedFactoryInboxWiring(...)`)를 건너뛰는 변이는 여전히 잡히지 않는다(바운드 피어가 필요한 진입점 단위 왕복 테스트 없음).
- F7: `TestManagedCodexTokenFileIsPrivate`(토큰 파일 0600, 디렉터리 그룹/기타 권한 0). 변이 `0o644` → `--- FAIL`(overlay).
- F10: 처분하지 않음(블로커급 보고). 구현은 도착순 FIFO(`managedTurnQueue.Next`)이고 AC-MS-008의 "연산자 우선"은 구현되지 않았다. overlay 프로브(저장소 미변경): claim이 도는 동안 도착한 연산자 줄이 채널에 대기 중일 때 `delivery order: [priming, "INBOX-BATCH", "OPERATOR-WAITING"]`.
- 검증: 관리 슬라이스 + M3/M4/M5 테스트 `ok … 18.889s`, 관리 파일 커버리지(statement-weighted) `managed_factory_session.go` 92.2% (153/166), `managed_codex_factory.go` 86.8% (198/228), 합 89.1% (351/394). vet exit 0, gofmt 빈 출력, lint `0 issues.`, store.go diff 0, vocab 0, syscall 0, windows exit 0, go.mod/go.sum 변경 0.

F10 처분 (2026-10-02, 기준 HEAD `175fa3398`): 측정 사실 — `managedTurnQueue.Next`는 도착순 FIFO(`turns[0]` pop)이고, claim이 도는 동안 도착한 연산자 줄은 overlay 프로브에서 `[priming, "INBOX-BATCH", "OPERATOR-WAITING"]` 순으로 inbox 뒤에 섰다. REQ-MS-006은 경합 없는 직렬화만 요구하므로 운영자 결정은 코드가 아니라 주석을 구현(FIFO)에 맞추는 것이다. `managed_factory_session.go`의 주석 3곳과 `TestManagedQueueSerializesOperatorAndInbox` 문서 주석 한 문장만 고쳤고 동작·단언·테스트 이름은 불변이다. AC-MS-008의 "연산자 우선" 문구는 SPEC 담당 워커가 정정한다(이 에이전트의 범위 밖).

AC-MS-008 FIFO 양방향 증명 (2026-10-02, 기준 HEAD `ba8c25396`): `TestManagedQueueSerializesOperatorAndInbox`에 큐·드라이버 두 수준의 거울 배치를 추가했다. 서브테스트: `queue serves arrival order: operator then inbox`, `queue serves arrival order: inbox then operator`, `driver claims only when idle and serves operator then inbox in arrival order`, `driver serves inbox then operator when the claim returns with an operator line already waiting`. overlay 변이(저장소 미변경): A = `Next`가 연산자 턴을 우선 → `inbox then operator`·`driver serves inbox then operator…` 서브테스트 `--- FAIL`; B = `Next`가 마지막 턴 선택 → 큐 두 서브테스트와 드라이버 거울 서브테스트 `--- FAIL`. 프로덕션 코드 변경 없음.
