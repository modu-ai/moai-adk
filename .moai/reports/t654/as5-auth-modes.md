# t654 A5-M4 — 구독/API 이중 경로 검증 (자동화 부분) 증거 (AS-021 / AC-MG-020)

## Claim

GPT 세션 catalog는 managed 구독(PKCE) 인증만 선언한다 — API 과금 경로가 구조적으로 존재하지 않아
구독 실패 시 자동전환이 불가능하다. 구독 credential 부재는 명시 401이고 upstream 요청 계수는 0이다.
launcher 바인딩 생성 경로는 구독 토큰 저장소(`gateway-auth`)에 접근하지 않는다. 표시된 인증 방식은
catalog 선언에서 파생되어 실제 세션 env와 일치한다(A5-M1 조립 시험이 판정).

## Evidence

명령: `go test ./internal/gateway/ -run 'TestGatewaySubscriptionFailureNeverSwitchesToAPI' -count=1`
기준 트리: 커밋 예정 워킹 트리(M4 시험 파일). 출력:

```
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.409s
```

- 양성 대조군(유효 구독 credential → 200 + upstream 1회)과 음성(부재 → 401 + upstream 0)의 대비로
  0계수가 공허 통과가 아님을 보인다.
- 로그: `.moai/reports/t654/as5-m4-authmodes-green.log`

명령: `go test ./internal/cli/ -run 'TestGatewayAuthModesDualPath' -count=1` →
`ok  github.com/modu-ai/moai-adk/internal/cli  1.001s`

- 전제 단언: GPT catalog 4항목(빈 catalog 공허 통과 방지), 전 항목 AuthMethod == AuthPKCE.
- 토큰 저장소 무접근: 격리 `MOAI_HOME`에서 바인딩 생성 뒤 `<home>/gateway-auth` 부재 단언 +
  양성 대조군으로 `<home>/state/gateway-conversations` 생성 확인(환경 겹침 실효성 증명).

## Baseline-attribution

2026-09-14, worktree `.claude/worktrees/t654`, base `6de8dd489`에서 직접 실행.

## Gaps (창 대기 — 라이브 계측)

- 실제 PTY에서 구독/ API 두 모드를 각각 선택·실행하고 표시 일치를 관측(AS-021 본문) — 실계정 권한
  필요, t851 이후 라이브 창.
- API 키 모드의 실제 실행과 표시("api-key") — 이 트리는 API-key GPT route를 선언하지 않는다
  (구조적 금지). API 모드 선택 표면(`gpt login`의 API 키 선택)의 실세션 판정은 형제
  SPEC-MOAI-GPT-AUTH-001(제안) 소관이다.
- 구독 만료·한도 오류의 실제 유도와 그때의 API 과금 경로 계수 0 관측 — 라이브 창.
- 구독 출력이 서버 정책임의 실측 표시 — 라이브 창.

## Residual-risk

- `settings.local.json`에 남은 키로 gateway 우회가 일어나는 경로는 원리상 이 계수가 보지 못한다
  (AC-MG-022 검증 가능성 절과 동일 경계) — launch 정리는 AC-MG-018 (a)가 소관이다.
