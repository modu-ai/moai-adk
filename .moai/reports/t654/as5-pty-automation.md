# t654 A5-M5 — 실제 PTY 실증 스윕 (자동화 부분) 증거 (AS-014·017·018·019)

## Claim (이 창의 몫 — 자동화 가능 부분)

subagent 네 슬롯 별칭의 제공자 경계 해석(AS-018의 자동화 축)과 미등록·타 제공자 ID 거절은 기존
계약 시험으로 고정돼 있으며, 이 창의 조립 변경(auth 표시·capability 재판정)이 그 계약을 깨지
않았음을 같은 시험 통과로 확인했다. 실제 PTY·실계정이 필요한 나머지는 전부 창 대기 Gap이다 —
카드 제약과 운영자 판정 3번(AS-017·019·021 창 대기 유지)을 따른다.

## Evidence

명령: `go test ./internal/cli/ -run 'TestGatewayProviderContract|TestGatewayExactModelContextAndGPTToolExposure|TestGatewayLaunchAssemblyCarriesAuthDisplay|TestGatewayNativeModelCapabilitiesPerProvider|TestGatewayAuthModesDualPath' -count=1 -v` (발췌)

```
--- PASS: TestGatewayProviderContractRejectsForeignModel
--- PASS: TestGatewayProviderContractPickerAndSlots
--- PASS: TestGatewayExactModelContextAndGPTToolExposure
--- PASS: TestGatewayLaunchAssemblyCarriesAuthDisplay
--- PASS: TestGatewayNativeModelCapabilitiesPerProvider
--- PASS: TestGatewayAuthModesDualPath
ok  	github.com/modu-ai/moai-adk/internal/cli
```

- AS-018 자동화 축(세션 catalog 안 해석·타 upstream 0·미등록 ID 자동 확장 없이 거절)은
  `TestGatewayProviderContractRejectsForeignModel`·`TestGatewayProviderContractPickerAndSlots`가
  소유한다(기존 시험 — 이 창에서 수정 없이 통과 재확인).
- ToolSearch hybrid의 기계 축은 t651 착지 시험(`internal/gateway/translate` 회귀)이 유지된다.

## Baseline-attribution

2026-09-14, worktree `.claude/worktrees/t654`, 최종 코드 상태에서 직접 실행.

## Gaps (창 대기 — 라이브 계측, 운영자 판정 3번)

| 항목 | 대기 사유 | 대기 명령(라이브 창에서) |
|---|---|---|
| AS-017 도구검색 실제 PTY (hybrid ToolSearch 제품 실행, `.moai/reports/t654/as5-pty-toolsearch.log` 증거) | 실계정 권한 — t851 gateway 400 해소 뒤 라이브 창 | 실제 launcher PTY에서 초기 도구+후발 도구 실행, gateway 요청 기록 대조 |
| AS-018 서브에이전트 실제 PTY (`.moai/reports/t654/as5-pty-subagent.log`) | 실계정 권한 | 실제 PTY에서 네 슬롯 별칭 유도 작업 실행 |
| AS-019 launcher 재개·`/model` 전환 실제 PTY — **launcher 측만** (Claude/GLM 재개·전환·provider 경계 + GPT 측 picker·인증 표시 AS-014까지). GPT thread 실세션 양성(AS-010·011·012)은 T21대로 t844 소관이며 흡수하지 않음 (`.moai/reports/t654/as5-pty-resume-model.log`) | 실계정 권한 | 실제 PTY에서 재개 플로우+`/model` 전환, 실제 turn 요청 model 대조 |
| AS-014 실제 PTY `/model` 전 표면(공유 설정 불변 포함) | 실계정 권한 | 실제 PTY에서 Default·현재 행·목록·직접 입력·Enter·s·재개 |

임시 Claude 직접 실험으로 제품 판정을 대체하지 않는다(카드 제약).

## Residual-risk

- 라이브 창이 열리기 전까지 세 launcher의 실세션 launch 경로는 자동화 시험 계약으로만 보호된다.
  대기 오류 게이트가 이를 1차 방어로 유지한다(A5-M1 대조군).
