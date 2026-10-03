# t1489 판정서 — verify 하위 동사 등록 누락

## Claim
`internal/cli` 의 init 순서 의존 때문에 실제 CLI 에서 `moai verify sync-gate` 가 등록되지 않았다. `audit-plan`·`codex-review` 는 파일명이 `verify.go` 보다 앞이라 우연히 등록돼 있었다. `verifyExtraCommands` 슬라이스와 init append 세 곳을 없애고 `newVerifyCmd` 에서 직접 등록하도록 고쳤다. 동사 동작은 바꾸지 않았다. 커밋 `5db7be3c3`(브랜치 `WT-verify-subcommand-registration`).

## Evidence
RED — 수정 전 바이너리(`go build ./cmd/moai`, HEAD 19de344af) `verify --help` COMMANDS:
```
    record [--flags] ...
    check [--flags] ...
    run  -- <command...> [--flags] ...
    audit-plan [--flags] ...
    codex-review                    Run the codex review of uncommitted changes and record a receipt (Codex Stop chain)
```
(`sync-gate` 없음 — `verify sync-gate --help` 는 상위 verify 도움말을 출력한다.)

RED 테스트 `go test -count=1 -run TestVerifySubcommandsRegisteredOnRoot ./internal/cli/` (exit 1):
```
--- FAIL: TestVerifySubcommandsRegisteredOnRoot (0.00s)
    --- FAIL: TestVerifySubcommandsRegisteredOnRoot/sync-gate (0.00s)
        verify_registration_test.go:13: rootCmd.Find(verify sync-gate) = &{verify ...}, rest [sync-gate], err <nil>; want the sync-gate command
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.034s
```
GREEN (레인 env를 한 번의 복합 호출로 세척):
```
go test -count=1 -run Verify ./internal/cli/                                      ok  55.924s
go test -count=1 -run "AuditPlan|CodexReview|SyncGate|Receipt" ./internal/cli/   ok  97.790s
go test -count=1 -run "Help|Registered|Reachab" ./internal/cli/                   ok  20.782s
go vet ./internal/cli/                                                            vet-ok
golangci-lint v2.1.6 run ./internal/cli/...                                       0 issues.
```
수정 후 바이너리 `verify --help` 에 `sync-gate  Run the sync-phase quality gate checks and record a receipt (Codex Stop chain)` 가 나타난다. card-review(codex, scope=card): pass, 지적 0.

## Baseline-attribution
워크트리 `agent-a8aee07892878282b`, 기준 HEAD `19de344af`(로컬 develop과 동일). RED와 GREEN 모두 이번 실행에서 이 트리로 빌드·측정했다.

## Gaps
- 전체 `./internal/cli/` 패키지 스위트는 돌리지 않았다(셀렉터 세 개만 실행). 전체 판정은 CI 에 맡긴다.
- `.moai/project/codemaps/*.md` 에 `verifyExtraCommands` 레지스트리를 서술한 과거 기록이 남아 있다. 비테스트 `AddCommand(` 줄 수도 +2 늘었다. codemaps 는 고치지 않았다(범위 밖이며 다음 재생성 몫).

## Residual-risk
앞으로 verify 동사를 추가할 때 `newVerifyCmd` 에 등록하는 것을 빠뜨릴 수 있다. 다만 레지스트리가 없어졌으므로 누락은 조용한 init 순서 결함이 아니라 미사용 함수(lint)로 드러난다. 새 동사는 테스트 목록에도 직접 추가해야 한다.
