# t653 — M6 증거 취합 (2026-09-14)

## 측정 baseline

- worktree `.claude/worktrees/t653`, branch `WT-gateway-as4-resume`
- 이 문서의 커밋 목록: `14dba89c5`(M3) → `e45f50a8d`(M4) → `64885fa06`(M5)
- 흡수 기준: HEAD `92db2cfe7` (= M1/M2 커밋 + develop `033323529` 병합)

## 마일스톤 커밋

| 마일스톤 | 커밋 | RED 증거 | GREEN 증거 |
|---|---|---|---|
| M3 idle 모델 변경 | `14dba89c5` | `m3-model-red.log` | `m3-model-green.log`, `m3-model-gateway.log` |
| M4 compaction 정상 요약 turn | `e45f50a8d` | `m4-compact-red.log` | `m4-compact-green-receipt.log`, `m4-compact-codexbridge.log` |
| M5 경계 fork | `64885fa06` | `m5-fork-red.log` | `m5-fork-green.log` |

## 최종 검증 (M6, 이 트리에서 실측)

| 검증 | 명령 | 관측 결과 |
|---|---|---|
| 변경 패키지 테스트 | `go test ./internal/codexbridge/ ./internal/gateway/receipt/ ./internal/gateway/conversation/ -count=1 -cover` (기존 환경 실패 4건 skip) | 3패키지 `ok` (`m6-coverage.log`) |
| gateway 패키지 | `go test ./internal/gateway/ -count=1 -cover -skip TestAppServerSubprocessHTTPToolContinuation` | `ok ... coverage: 91.7%` |
| 커버리지 | 위 `-cover` 실행 | codexbridge 83.1% / receipt 88.9% / conversation 80.3% / gateway 91.7% |
| Windows 교차 컴파일 | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 (`WINDOWS BUILD OK`) |
| lint | `golangci-lint run --timeout=5m` | `0 issues.` exit 0 (`m6-lint.log`) |
| gofmt | `gofmt -l internal/codexbridge internal/gateway` | 빈 출력 |

## 사전 존재 환경 실패 (이 diff와 무관 — baseline 트리에서 동일 재현 확인)

- `internal/codexbridge` 4건: `TestAuditRealTransportEOFWakesBridge`, `TestAuditCancelBeforeRealStartReplyInterruptsTurn`, `TestCanceledStartBeyondReplyDeadlineStillInterrupts`, `TestAuditCanceledRPCDoesNotExhaustSharedTransport` — `lifecycle_subprocess_test.go`, "app server start failed". 커밋 트리(작업 변경 없는 `git archive HEAD` 추출)에서 동일 실패 재현으로 귀속.
- `internal/gateway` 1건: `TestAppServerSubprocessHTTPToolContinuation` — `appserver_integration_test.go:71`, 동일한 "app server start failed". 커밋 트리 추출본에서 동일 재현으로 귀속.
- 원인: 하위 프로세스 기동이 이 샌드박스 환경에서 실패(python3 fake-codex 스크립트 기동). 원인 규명은 이 카드 범위 밖 — CI(원격)가 전체 스위트 판정의 주체다.

## AS 대응 (자동화 부분)

- **AS-010** (M2, 선행): resume 경로·거절군 — 이미 착지. 이번 M3-M6에서 재검증됨(패키지 테스트 회귀 0).
- **AS-011** (M3): waiting 전환 거절, 전환 후 핀 대조(신모델 성공/구모델 ErrScope), turn/start가 신모델 운반, gateway 측 bare `gpt-6` ErrUnknownModel + 타 provider adapter 진입 거절(`ErrManagedAuthority`). 실제 turn model 일치(GPT-6 Astra/5.6 선택 ID)는 실증.
- **AS-012** (M4): PostCompact 대조기(exact digest, scope/epoch), 중복/지연(=stale)/유실(=gap) 재시작 재생 거절, substring-only 거절, `Manifest.Rebase`·`Store.Rebase` public history 재설정(영속성 포함), bridge 측 compact RPC 0회 + 서버 주도 compaction 알림 무시. 실제 Claude 압축 요청·응답·후행 hook 수집은 실증.
- **AS-013** (M5): 경계 fork 전체(경계 초과 유입 차단, 경계 체인 보존, 거절군 — 미지/영 경계/사이클/불일치 링크/미지 원본), fork 자식의 신규 thread 기동 + inherited prefix, 자식 배리어의 신규 프로세스 resume. `--fork-session` 실분기·병렬/중첩 자식 격리·native fork 양성은 실증(후자는 M1 전제 실패 NOT-RUN — AS-013 전체 지원 완료는 계속 보류).

## 설계 판정 기록

1. **compaction은 engine 변경 0건으로 분류된다** — 정상 turn/completed 경로가 그대로 요약 turn을 운반하고, RPC id 없는 서버 주도 알림은 기존 read/Step 루프가 무시한다(테스트로 고정). 별도 RPC나 PreCompact 분류는 도입하지 않았다(계약 금지).
2. **고리 걸린 Previous 링크는 체인 뿌리다** — 이 원장은 Previous가 manifest에 없는 접두사를 가리키는 것이 정상(core_test 자체가 그렇게 구성). ChainTo는 미해결 링크에서 걷기를 멈추고, 거절은 미지/영 경계·사이클·불일치 링크에 국한한다.
3. **`Manager.Fork`(tip 분기)도 ChainTo로 정렬** — 계층 간 의미 일치(라우처 --fork-session = lastTurnId 분기).
