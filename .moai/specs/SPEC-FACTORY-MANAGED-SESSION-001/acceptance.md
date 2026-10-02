---
id: SPEC-FACTORY-MANAGED-SESSION-001
title: "acceptance.md — 인수 기준"
version: "0.1.1"
created: 2026-10-01
updated: 2026-10-01
author: GOOS (manager-spec)
tier: L
---

# acceptance.md — AC-MS-001..018

> 모든 AC는 기계 판정형. RED/GREEN 2칸은 run-phase에서 채운다(verification-completeness.md §2 — 채택 시 RED-now 셀 + green path 셀 페어). 검증 명령은 단일 호출형(파이프·`&&`·리다이렉션 없는 형태)을 원칙으로 하고, 워크트리 가드상 `go test ./...` 전체 대신 변경 영향 패키지 스위트를 돌린다(레인 로컬 검증 규율). `-run` 패턴은 전부 앵커형(`^…$`)이다.

## §1. AC 매트릭스

| AC | 요구 | 시나리오 (Given-When-Then) | 검증 |
|---|---|---|---|
| AC-MS-001 | REQ-MS-001 | Given 임시 projectRoot와 가짜 app-server 바이너리(re-exec) When 관리 Codex 런처가 기동하면 Then 루프백 WS에 토큰 인증으로 접속하고 `/readyz` 대기 후 `initialize`가 성공한다 | `go test ./internal/cli -run '^TestManagedCodexAppServerHandshake$'` |
| AC-MS-002 | REQ-MS-001 | Given 핸드셰이크 성공 When 스레드가 시작되면 Then 브로커에 스레드 ID가 SessionUUID인 엔드포인트가 등록된다 (status에 표시) | `go test ./internal/cli -run '^TestManagedCodexRegistersBoundPeer$'` |
| AC-MS-003 | REQ-MS-002 | Given launch-pending 등록 후 세션 기동 실패 When 런처가 종료하면 Then launch-pending 잔행 행이 0건이다 | `go test ./internal/cli -run '^TestManagedLaunchPendingRollback$'` |
| AC-MS-004 | REQ-MS-003 | Given 브로커에 수신 메시지 2건 When 관리 세션이 idle이 되면 Then claim 배치의 메타데이터(id/claim_token/kind/from/task_ref)만 다음 턴 프롬프트에 주입되고 본문 문자열은 프롬프트에 없다 | `go test ./internal/cli -run '^TestManagedInboxPromptMetadataOnly$'` |
| AC-MS-005 | REQ-MS-004 | Given claim된 메시지 When 세션이 body를 조회하면 Then claim token 없는 조회는 거부되고 토큰이 있는 조회만 본문을 돌려준다 | `go test ./internal/factorymsg -run '^TestReadBodyClaimToken$'` |
| AC-MS-006 | REQ-MS-004 | Given 처리 완료 When receipt를 기록하면 Then store status의 acknowledged 카운트가 증가하고 pending이 감소한다 | `go test ./internal/cli -run '^TestManagedReceiptAcknowledges$'` |
| AC-MS-007 | REQ-MS-005 | Given 관리 Claude/GLM 기동 When 사용자가 `-p`/`--input-format` 등 소유 플래그를 전달하면 Then 런처는 오류로 거부하고 stream-json 플래그를 강제한다 | `go test ./internal/cli -run '^TestManagedSessionOwnsStreamFlags$'` |
| AC-MS-008 | REQ-MS-006 | Given 연산자 입력 1건과 브로커 메시지 1건이 동시에 있을 When 이전 턴이 완료되면 Then 두 입력이 도착순 FIFO로 한 번에 한 턴씩 직렬 주입되고 busy 동안 새 claim이 없다 | `go test ./internal/cli -run '^TestManagedQueueSerializesOperatorAndInbox$'` |
| AC-MS-009 | REQ-MS-007 | Given 신규 관리 소스 파일들 When 라벨 생성 표면을 스캔하면 Then `agent-<n>`/`worker-<n>` 라벨 생성이 0건이고 lane 라벨만 만든다 | `grep -rnE 'fmt\.Sprintf\("(agent\|worker)-' internal/cli/managed_*.go` 0행 |
| AC-MS-010 | REQ-MS-014 · REQ-MS-015 | Given 신규 관리 소스 파일들 When F3 상징과 수신확인-판정 상징을 스캔하면 Then 병합 자동화/Decider/핸드오버/완료판정 상징(`merge-window`, `Decider`, `T29b`, `T29c`, 카드 상태 전이 호출)이 0건이다 (F3 비침범 — receipt는 배달 증거로만 존재) | `grep -rnE 'merge-window\|Decider\|T29b\|T29c\|handover' internal/cli/managed_*.go` 0행 |
| AC-MS-011 | REQ-MS-011 | Given Claude-only 브로커 run When `factory_msg_send`가 호출되면 Then 거부된다(네이티브 SendMessage 정책 유지) | `go test ./internal/cli -run '^TestFactoryMsgSendRejectsClaudeOnlyRun$'` (계승, 재작성 후에도 통과) |
| AC-MS-012 | REQ-MS-009 | Given 관리 Codex 기동 When 승인 인수를 검사하면 Then approve 인수는 소유 App Server 프로세스에만 붙고(TUI는 본 SPEC에서 미배달), 프로젝트 config 생성물의 `default_tools_approval_mode`는 `"writes"`로 불변이다 | `go test ./internal/codexwiring ./internal/cli -run '^(TestMoAIMCPApprovalArgsOnlyTargetMoAI\|TestConfigTomlWritesUnchanged)$'` |
| AC-MS-013 | REQ-MS-010 | Given 구식 프로젝트 전역 승인 잔재 When `moai doctor codex`가 실행되면 Then 경고를 내고 `.codex/config.toml`을 변경하지 않는다 | `go test ./internal/cli -run '^TestDoctorCodexWarnsStaleGlobalApproval$'` |
| AC-MS-014 | REQ-MS-012 | Given 신규 관리 파일들 When 크로스빌드+스캔하면 Then `GOOS=windows GOARCH=amd64 go build ./...`가 exit 0이고 신규 파일의 `syscall` 참조가 0건이다 | 크로스빌드 exit 0 + `grep -rn 'syscall\.' internal/cli/managed_*.go` 0행 |
| AC-MS-015 | REQ-MS-013 | Given 실제 제2호스트가 없는 CI When loopback 통합 테스트를 돌리면 Then 브로커 claim→메타데이터 주입→body 조회→receipt 왕복이 자기 재실행(re-exec) 가짜 세션으로 통과한다 | `go test ./internal/cli -run '^TestManagedSessionLoopbackRoundTrip$'` |
| AC-MS-016 | REQ-MS-013 | Given live 게이트 env 미설정 When 전체 패키지 스위트가 돌아가면 Then live 테스트는 skip으로 통과하고, env 설정 시에만 실세션 왕복을 실행한다 | `go test ./internal/cli -run '^TestManagedCodexFactoryBrokerLive$'` (env 없이 → skip 확인) |
| AC-MS-017 | REQ-MS-008 | Given run-phase 커밋 목록(카드 base = `f22e2d7ac`) When 브로커 저장소 경로의 커밋 열거를 돌리면 Then `internal/factorymsg/store.go` 변경 커밋이 0건이다 — API 부족 시 침묵 패치 대신 blocker로 보고했음을 기계 확인 | `git log --oneline f22e2d7ac..HEAD -- internal/factorymsg/store.go` 빈 출력 (base는 흡수 시 merge-base로 재산출 — gitflow-lane-protocol §8) |
| AC-MS-018 | REQ-MS-001 · REQ-MS-005 | Given factory 스탬프(`MOAI_KANBAN_ID` + `MOAI_FACTORY_WORKER` 또는 `MOAI_FACTORY_WORKERS`)가 있고 `MOAI_FACTORY_MANAGED`가 있을 때와 없을 때, Claude/GLM 런처와 평범한 `moai codex` 런처 양쪽에서 When 기동하면 Then 스위치(`1` 또는 `true`)와 스탬프가 모두 있을 때만 관리 소유자로 divert하고, 스탬프만 또는 스위치만이면 기존 exec 경로에 남으며, 두 값이 모두 있는 `moai codex -f lane`은 여전히 감독 루프로 가고 관리 seam은 호출되지 않는다 | `go test ./internal/cli -run '^(TestFactoryManagedRequested|TestManagedLaunchRequiresOptIn|TestManagedCodexLaunchRequiresOptIn|TestManagedSwitchDoesNotReachCodexLaneLoop)$'` |

## §2. 에지 케이스

- **launch-pending 경합**: 기동 중 다른 프로세스가 같은 레인으로 조인 시도 → `ErrEndpointLaunchPending` 응답, 롤백 후 재시도 가능.
- **owner 소멸**: 관리 세션 크래시 후 `PeerByOwner` 지문 불일치 → 스테일 엔드포인트 미사용, dead-letter 정책은 store 기존 동작 위임.
- **App Server 조기 종료**: WS 이벤트 채널 close → 세션 종료 오류로 승격, 소유 App Server 프로세스 정리(kill+Wait) 보장.
- **lease 만료 중 처리**: Claim lease 내 턴 미완료 → store의 lease 정책에 위임(관리 계층이 재구현하지 않음 — F3의 lease reaping과 혼동 금지).
- **busy 중 수신**: 턴 진행 중 도착 메시지는 다음 idle 폴링에서 claim — 즉시 인터럽트 금지.
- **`--` 이후 `--name`**: PR 리포트의 회귀(R2) — 레인 라벨은 `--` 앞에 배치하는 회귀 테스트 유지.

## §3. 품질 게이트

- TRUST 5 — Tested: 신규 패키지 코드 커버리지 85% 이상(`go test -cover ./internal/cli/...`, 관리 파일 대상 측정), Readable/Unified: 기존 `internal/cli` 스타일, Secured: 토큰 파일 0600·루프백 바인드·본문 미신뢰 문구, Trackable: Conventional Commit `feat(SPEC-FACTORY-MANAGED-SESSION-001): M<n>`.
- LSP 게이트: run 종료 시 0 error / 0 type-error / 0 lint-error.
- 회귀: `internal/cli`·`internal/kanban`·`internal/hook`의 Factory 스위트와 `internal/template` 미러 패리티 스위트 통과.

## §4. Definition of Done

1. AC-MS-001..018 전부 PASS (§E 매트릭스에 명령+출력 인용).
2. F3 비침범(AC-MS-010)·어휘(AC-MS-009)·zero-syscall(AC-MS-014)·store 무변경(AC-MS-017) 인용.
3. `GOOS=windows` 크로스빌드 exit 0.
4. 회귀 스위트(§3) 통과 — 기존 lane 어휘·조인·디스패치 동작 무손상.
5. CHANGELOG 항목(매니저-docs, B12 계수 규율) + progress.md §E.2/§E.3 증거 착지.
