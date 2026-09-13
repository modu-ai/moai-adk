# t707 follow-up RED 설계 요지 (seam 판정 완결분)

> 작성: lane-5, 2026-09-14. 출처: t707-run 에이전트의 읽기전용 코드·증거 분석(정지 전 최종 보고).
> 후속 카드 배차 시 이 파일을 프롬프트에 그대로 인용할 것. 트리: WT-edit-replay-chain, 선행 커밋 1b4547e3a(t707 본커밋).

## 1. 확정된 배경 사실

- **바이너리 패리티**: 서빙 바이너리 7a7a08f20과 HEAD 0c32a15b2 사이 6커밋(t835·t836·t702)은 internal/gateway·gateway_factory 변경 0 — 라이브 실패는 현재 HEAD 코드의 행동이다. 구-코드 분기는 소멸.
- **checkObserved 조기 반환은 무혐의(코드 순서 증명)**: 4개 응답 경로(Response / Stream / SubscriptionStream / SubscriptionResponse) 모두 `c.response(r)`(내부 Publish)가 클라이언트 관측 첫 바이트에 선행하고, 스트림 경로는 History 활성 시 publishOutput 완료까지 emit을 억제한다. upstreamSend는 스트림 시작 전 5xx/transport만 재시도하고 시작된 스트림은 재구동하지 않는다(upstream.go 17-19). 따라서 Publish 실패 시 클라이언트는 턴을 받기 전에 400/502 — "클라이언트가 미기록 턴을 보유"하는 경로가 없다.
- **실제 시나리오(확정)**: 중단된 1차 시도는 죽은 파이프로 emit이 실패해 completed 처리가 중단되거나, completed 이후 첫 emit 실패로 "미사용 후보"만 남긴다. 후자는 무해 설계이며 81f68955의 14:07:19 fallback 통과 케이스가 이 형상이다.
- **81f68955 :2787-2841 정독**: 14:01:33경 발사된 agent_summary 스트림이 무이벤트로 끊김(14:02:31) → non-stream fallback(14:02:31.020) → 응답 도착(14:02:52.474) → Edit 실행(opaque call_oVnR…) → 서브에이전트 다음 요청(14:02:52.685) 400. 즉 장애 창구간에 **같은 패밀리 루트에 두 요청(서브에이전트 본요청 + 요약 포크)이 동시 in-flight**이었고 포크는 스트림 1차+fallback 2차 — 한 세션 루트에 3개 발행 경로가 교차했다.
- **좁혀진 원인 후보**: "미기록 턴"이 아니라 **동일 (prefix, previous) 공유에 의한 required/empty 충돌** 또는 **포크 발행 경계의 체인 교차**.
- 키잉 금지: Edit/Write 종류·toolu_moai_v1_ 형태로 키잉하지 말 것 — 사례에서 둘 다 갈리고 opaque 래핑은 로그 전체에 보편.

## 2. RED 설계 3종 — `internal/gateway/translate/replay_seam_test.go`

1. **TestReplayFallbackRetryAfterAbandonedAttempt** — 1차 시도 완료·폐기(클라이언트 미수신) → 동일 요청 재시도 → 재시도 턴 발행·재생 통과 + 후보 2개 확인. 현재 코드 PASS 예상(무해 증명). 실패 시 seam 결함 확정.
2. **TestReplayConcurrentPublishDistinctTurns** — 동일 스토어 뿌리에 두 고루틴 동시 Publish → flock+CAS로 양쪽 착지·양 브랜치 재생 통과. 실패 시 결함 확정.
3. **(추가 권고) 포크 체인 교차 변형** — 요약 포크가 서브에이전트 체인 위에서 발행하는 형상(포크 history = 서브에이전트 [0..k] + 포크 턴)에서 서브에이전트 재생이 살아남는지 — checkObserved의 required-충돌 룰이 포크 후보를 오조준하는지 검사. t653이 착지한 receipt ForkAt/ChainTo(`internal/gateway/receipt/core.go`)와 접점이 있으니 참고.

## 3. 함께 갈라낼 것

- 81f68955 rows 5069-5241 · 5555-5767 · 14:07:19 분기 정독은 미완(후속 세션 몫).
- 관측 1줄(Check 거절 시 최초 미일치 관측 인덱스·뿌리 여부·items)은 `.moai/reports/t707/observability.patch`에 준비돼 있음(컴파일 미검증 표기 유지) — t837과 겹치면 이 1줄만.
- 라이브 계측(MOAI_RECEIPT_DEBUG=1)은 카드 t838 소관 — 본 RED 스위트와 독립 병행 가능.

🗿 MoAI
