# t688 게이트웨이 502 재발 기록

## 주장

재시작 뒤 t688 plan-phase의 `manager-spec` 호출에서 HTTP 502가 1회 다시 관측되었다. 같은 호출을 재시도하지 않았다.

## 증거

- 관측 시각: `2026-09-13T21:34:26+0900`
- 시각 근거: 실패한 agent output 기록의 파일 수정 시각
- conversation family: `8de36537-8e99-4b49-8a08-fd0c4e20b4d5`
- 단계: `manager-spec` plan-phase artifact authoring
- 모델: `gpt-5.6-sol`
- 오류: `API Error: 502 Bad Gateway`
- 오류가 가리킨 gateway: `127.0.0.1:53740`
- 재시작 신호에 기록된 gateway: pid `87995`, binary `9bfe424ee`
- 이번 재시작 뒤 이 lane에서 관측한 502 횟수: `1`
- 조치: 동일 호출 재시도 없음

Agent failure notification의 오류 본문:

```text
Agent terminated early due to an API error: API Error: 502 Bad Gateway. This is a server-side issue, usually temporary — try again in a moment. If it persists, check your inference gateway (127.0.0.1:53740). (error type server_error, HTTP 502, model sent to the API: gpt-5.6-sol)
```

## 기준 귀속

- worktree: `.claude/worktrees/t688`
- branch: `WT-graph-stamp-freshness`
- 시작 HEAD: `7097e6e21`
- 호출 전 SPEC ID 검사: `SPEC_ID_AVAILABLE`

## 미검증

- 502 응답의 실제 생성 주체가 gateway 자체인지 상류 서버인지는 이 기록만으로 판정하지 않았다.
- 실패 알림은 agent 내부 재시도 횟수를 제공하지 않는다. `횟수: 1`은 이 lane이 받은 502 실패 알림의 수다.
- 실패 뒤 `spec.md`만 생성되었고 `plan.md`, `acceptance.md`, `progress.md`는 존재하지 않았다. 문서 완성도와 전용 SPEC audit은 검증하지 않았다.
- 직전 조사 workflow 실패는 HTTP 502가 아니라 합성 agent의 180초 무진행 6회였으므로 이 502 횟수에 포함하지 않았다.

## 잔여 위험

부분 생성된 `spec.md`가 있으므로 다음 plan 재개에서는 새로 시작했다고 가정하지 말고 현재 파일을 먼저 읽어야 한다. 같은 gateway 상태에서 즉시 재호출하면 t697이 추적하는 반복 실패를 늘릴 수 있다.
