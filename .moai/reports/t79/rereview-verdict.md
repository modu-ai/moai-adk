# t79 Re-Review Verdict — PASS (F1 해소)

- Card: t79 — glm_task MCP 도구군. 최초 판정 FAIL(차단 F1 — `review-verdict.md` 참조)
- Fix commit: `d7f9f3b3a` (delta `9865e87ed..d7f9f3b3a`, 3파일 +169/−29, 부모 = 9865e87ed ✓)
- Reviewer session: release-v311 (2acd4be4), 2026-08-17

## Claim (주장)

F1 수정 4점이 세트로 납입되어 `DefaultGLMTaskTimeout`이 양쪽 arm의 실질 상한이 되었고,
audit 경로(120s)는 무변경이며, 증명 테스트가 상수의 지배를 입증한다.

## Evidence (증거) — reviewer-executed (diff 전문 독회 + 재실행)

| Fix point | 관찰 |
|---|---|
| ① 태스크 전용 seam | `glmTaskHTTPClient glmHTTPDoer = &http.Client{}` (glm_task.go:84 — 클라이언트 Timeout 없음); `callGLMTask`가 이 seam 사용(:294); **mcp_glm.go는 fix에 미포함** → audit `glmHTTPClient` 120s 무변경 (grep 지도로 확인) |
| ② 양쪽 arm 상한 | sync arm에 `context.WithTimeout(ctx, config.DefaultGLMTaskTimeout)` 신설 + `defer cancel()`; background 기존 랩 유지(코멘트 정정). `errors.Is(err, context.DeadlineExceeded)` → `glmTaskTimeoutMessage()`("timed out after X") — 구조화 실패 결과, IsError 아님(테스트 단언) |
| ③ 코멘트 정정 | 헤더("no second HTTP client" 삭제→seam 분리 명시), 잘못된 "120s는 커넥션 하나" 주석 교체, defaults.go 문서 재작성 — 전부 실제 동작과 일치 |
| ④ 증명 테스트 | `withShortGLMTaskTimeout(100ms)` + ctx 블로킹 doer. **제 재실행: SyncDeadlineGovernsCall PASS 0.10s / BackgroundDeadlineRecordsFailed PASS 0.11s** — 구동 시간 자체가 100ms 상수 발화의 물증(120s 지배였으면 120s, 무상한이었으면 행). 기존 테스트 전부 `withGLMTaskSeams`로 이전 — callGLMTask가 실제 읽는 seam을 찌름(실네트워크 우회 지속) |

| Check | Command | Observed |
|---|---|---|
| 테스트 | `go -C <t80> test ./internal/cli/ -run 'GLMTask' -v` | 14/14 PASS (증명 2건 포함) |
| 테스트 | `-run 'GLM\|TestMoaiMCPServer\|TestAC_C' -count=1` | ok 1.484s |
| RED | /tmp/t80-red-f1.txt | 정당 클래스(신규 심볼 `glmTaskHTTPClient` 미정의 빌드 실패 — behavioral RED는 무상한 sync에서 프레임워크 타임아웃까지 행하므로) |
| Ride-along | `git log release/v3.1.1..WT-t80` | 3커밋(baseline 머지 + 구현 + fix) — 정합 |

## 잔여 경고 해소 확인

최초 판정이 전달한 "클라이언트 Timeout만 올리면 sync는 상한을 아예 잃는다" — seam과
sync 랩이 **세트로** 납입되어 해소 (감사 리포트 residual-risk 조항 폐쇄).

## Gaps (미검증)

- 구현측 보고 2건 수용(보수적 방향): 호출자 ctx의 더 짧은 deadline에도 glm_task 상수명
  언용(codex 대칭) · audit seam 고정 테스트 부재(기존 withGLMSeams 감사 테스트가 암묵
  커버 — 드리프트 시 실네트워크로 실패).
- golangci-lint 0 주장 인용(재실행 안 함, CI 판정).

## Residual-risk (잔여 위험)

- 종전 F2–F8(비차단, codex 상속 대칭)은 그대로 — 양 패밀리 공통 후속 카드 권장.
- 레코드의 Error에 타임아웃 메시지가 남는 것은 설계된 패턴(fail-open 명명 규약).

**Verdict: PASS** — release/v3.1.1 no-ff 통합 진행 가능.
