# Card t695 — Verdict

- card: t695 ([긴급] GPT 모델 4종 400 전수 조사)
- branch: WT-gpt-model-400-sweep (base: origin/develop 74d872aaf)
- evidence: .moai/reports/t695/{verdict,matrix,investigation}.md
- date: 2026-09-13, lane-6

## Claim (주장)

1. 게이트웨이의 effort 정책 거절이 t695의 원인이다 — 모델 불문, 허용값({high}+GPT 한정 {medium}) 밖(low·xhigh·max) 전부가 사유 은폐된 400을 냈다.
2. 수리 후 Claude Code가 내보낼 수 있는 모든 (모델 × effort) 조합이 실제 게이트웨이 요청으로 정상 응답한다.
3. translate 오류의 사유 은폐(openAITranslationError)가 제거돼, 이후 400은 실제 거절 사유를 본문에 실어 보낸다.
4. t695는 스폰 표면의 400(카드 t672 계열)을 흡수하지 않는다 — 별개 기제 2종(수신 확인됨)으로 분해돼 t672 소관으로 넘긴다.

## Evidence (증거 — 이번 run에서 직접 관측)

**통제 실험 (D2, 계측 빌드, 실제 상류 요청, 상세: investigation.md §2)**

```
sol+low   → 400 "unsupported effort" (3ms)   ← 로컬 거절, 상류 미도달
luna+low  → 400 "unsupported effort" (0ms)
luna+high → 200 실응답 (2542ms)
sol+medium → 200 / sol+effort 없음 → 200
```

**상류 수용 범위 (임시 허용 확장 빌드, 커밋 전 원복)** — low·medium·high·xhigh: 4모델 전부 수용. max: sol 4/4·terra 4/4·luna 6/7(1건은 probe 부산물)·**astra 0/4 거절** — 모델별 거절. → 수리: 통과 허용 {low,medium,high,xhigh,max} + astra+max → xhigh 맵핑 (investigation.md §3).

**전수 매트릭스 (D4, 실제 게이트웨이 요청, 최소 토큰)** — 수리 후 **57셀 전부 초록, 적색 0**: 일반 20/20 + 도구 포함 20/20 + 스폰 형상 8/8 (게이트웨이 HTTP 직접) + e2e 8/8 (실 `moai gpt` 런처 + 실 Claude Code 클라이언트: luna/astra/terra+low·sol+xhigh 신규, 동일 모델 이어쓰기, 동일 패밀리 전환 sol→terra, 교차 패밀리 →astra, 서브에이전트 스폰). 판별자: `claude-*` 모델 → 404 사유 있는 응답(마스크드 아님). 전체 표: matrix.md.

**운영자 실측 (수리 전, 프로덕션 게이트웨이 — matrix.md §A에 타임스탬프와 함께 기록)**
astra/sol/terra+low → 400, terra 기본 effort → 성공, xhigh → 400, luna+low(6d392874) → 400, sol+high(lane 8/9/10) → 정상.

**회귀 테스트 (D5)** — RED→GREEN: 수리 전 openai.go 복원 상태에서 `TestOpenAITranslationErrorSurfacesReason` FAIL, 수리 후 PASS. `go vet ./internal/gateway/...` clean. `golangci-lint --new-from-rev=74d872aaf internal/gateway/...` → 0 issues. `go test ./internal/cli/` → `ok 1276.059s EXIT=0`. 유일 실패 `TestAppServerSubprocessHTTPToolContinuation`은 pristine HEAD 74d872aaf에서 동일 실패(git archive 재현) — 기존 환경성 실패, 본 수리와 무관.

**레인 오케스트레이터 표본 재검증 (수리 에이전트 보고 별도)**
`go vet ./internal/gateway/... && go test ./internal/gateway/ -run TestOpenAITranslationErrorSurfacesReason -count=1 ./internal/gateway/translate/ -count=1` → `ok github.com/modu-ai/moai-adk/internal/gateway 0.463s`. 커밋 4개(e8e24ad63·72d34f7e1·76ef6e30a·02f7d2dba), diff 8파일(gateway 소스 5 + 증거 3)로 범위 적정 확인.

## Baseline-attribution (baseline 귀속)

- 대상 트리: 브랜치 `WT-gpt-model-400-sweep` @ 02f7d2dba (verdict 커밋 시 갱신), base origin/develop 74d872aaf — 이번 run, 이 워크트리에서 측정.
- 상류 수용 범위는 z.ai 실요청(2026-09-13) 기준 — 문서 추정 아님.

## t672 흡수 판정 (D6)

**t695는 어느 쪽도 흡수하지 않는다.** 스폰 표면 400은 기제 2종으로 분해:

1. **receipt/이력 축 (실사유형)** — lane-8의 "conversation history changed…" 400은 HistoryReplayError(receipt_history.go:18, 게이트웨이 자체 문구)로, 카탈로그 해석 **이후에만** 나온다. wire 모델이 opus/sonnet이면 먼저 404(사유 있음)가 나오므로(매트릭스 §F 판별자), 실제 재현은 GPT wire 모델 요청이 부모 이력 재생에서 receipt 검사에 걸린 형태. 모델 인자가 이력 재생 포크를 강제하는 선행 조건일 가능성 — **t672 검증 범위**.
2. **마스크드형 (@t688-advisor 2건, 08:30:01/08:30:27Z)** — 미특정 translate 경로 거절. D1 착지로 다음 발생 시 본문이 스스로 사유를 밝힌다.

부수 관측(Residual 기록): 0efb66f7(lane-8) 502(08:42:43Z) 이후 본 대화 전면 탈동조화(09:13-09:18 family 400 연속, sol→terra 전환 후에도 마스크드 400) — receipt 사슬 복구 정책도 t672 범위 후보. SubagentStart 컨텍스트에 완료된 SPEC-ARTIFACT-STATELESS-001(t357) 주입 — 별도 이슈, 리드 통보.

## Gaps (미검증 — 명시적)

1. astra+max 거절의 상류 상태/본문 — 어댑터가 비-200을 가려 관측 불가. 맵핑은 4/4 거절 대 xhigh 200 대조에 근거.
2. 라이브 스폰 요청의 정확한 모델 id — 클라이언트 로그에 요청 본문이 없어 추출 불가. 계측 스폰 셀과 기제 분석으로 대체.
3. 이름 붙은 팀메이트 스폰 형상 — 이 레인에서 재현 불가. 라이브 증거 + 기제 분석으로만 커버.
4. e2e 셀이 실제 MOAI_HOME을 사용(auth store가 symlink 거부, 토큰 복사 금지) — 자기 세션의 새 family만 생성, 기존 family 미조작.

## Residual-risk (잔여 위험)

- astra+max → xhigh 맵핑은 상류 정책 변경 시 재검증 필요 (상류가 max를 수용하면 맵핑 제거 가능).
- 마스크드형 스폰 400의 근본 원인 미특정 — D1으로 다음 발생 시 식별 가능.
- receipt/이력 축(②③계열)은 현재 develop에서 여전히 살아 있음 — t672 착지 전까지 스폰 실패·세션 탈동조화 재발 가능.
- 상류가 특정 effort 값을 거절하기 시작하면 통과 허용 정책이 400(사유 표시됨)으로 드러난다 — 은폐보다 낫지만 모니터링 필요.

## 통합 참고

- 병합 창에서 흡수 필요: origin/develop 이동(4c427fecb, t681·t680) — 레인이 창 안에서 로컬 develop 흡수 후 병합 트리 재측정.
- 워크트리 유지(미푸시 브랜치의 유일본).
