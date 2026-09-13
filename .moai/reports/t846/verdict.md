# t846 판정문 — receipt seam RED 3종 (seam 무혐의 확정)

날짜: 2026-09-14 · 브랜치: WT-receipt-seam-red · HEAD(시작): fbcddfa0b (93ae49ce7 + t707 후속 RED 설계 브리프 체리픽)
Class B(run→sync, plan 생략) · 수행: manager-develop · 설계 출처: `.moai/reports/t707/followup-red.md`

## Claim (주장)

1. **RED 설계 3종 모두 현재 코드에서 PASS — seam 무혐의(GREEN) 확정.** 폐기된 1차 시도(클라이언트 미수신) 뒤 동일 요청 재시도, 동일 스토어 뿌리 위 동시 발행, 포크 체인 교차의 세 형상 어디에서도 재생 거절·후보 유실·required-충돌 오조준이 관측되지 않았다.
2. **Test 1 (TestReplayFallbackRetryAfterAbandonedAttempt): PASS — 무해 설계 증명.** 1차 시도는 클라이언트 단절 시나리오(seamDeadWriter)에서 스트림이 실패하지만 후보는 착지한다(발행이 첫 emit에 선행 — t707의 코드 순서 증명과 일치). 재시도 턴이 자기 후보를 발행하고, 매니페스트는 정확히 2후보, 재생 통과. 폐기 후보는 다른 prefix라 간섭하지 않는다.
3. **Test 2 (TestReplayConcurrentPublishDistinctTurns): PASS — flock+CAS 무유실 증명.** 3라운드 × 동시 2발행에서 매번 정확히 2후보 착지, 양 브랜치 재생 통과. -race 20회 반복 재검증에서도 20/20 통과.
4. **Test 3 (TestReplayForkChainCrossing): PASS — required-충돌 룰 오조준 없음.** 서브에이전트 체인(필수 2턴) 위에서 요약 포크가 발행해도, 포크 후보는 포크 고유 내용에서 파생된 (prefix, previous)를 가지므로 서브에이전트 경계와 충돌할 수 없다 — 포크 턴이 opaque reasoning을 갖는 변형(Required)과 갖지 않는 변형(빈 후보, checkObserved의 required-충돌 룰이 검사하는 유일한 클래스) 모두에서 서브에이전트 재생이 살아남았고, 포크 이후 서브에이전트가 체인을 계속 발행해도(4후보) 전체 재생이 통과했다.
5. **검증 중 발견한 간헐 실패 1건은 제품 결함이 아니라 테스트 작성 결함이었다** — 아래 Gaps/부록 참조. 제품 코드 변경은 없다(테스트 파일 + 증거만).

**GREEN 판정에 따라 라이브 국소화의 이관 카드는 t838(MOAI_RECEIPT_DEBUG=1 라이브 계측) 그대로 유효하다.** 본 스위트는 합성 재현 범위에서 seam의 무해를 증명했을 뿐, 라이브 400의 최종 국소화는 요청 본문 기록이 있는 라이브 계측 1회를 필요로 한다(t707 갭 1과 동일).

## Evidence (증거)

테스트 파일: `internal/gateway/translate/replay_seam_test.go` (신규, 테스트 전용 — 프로덕션 변경 0)

| 테스트 | EXPECTED | OBSERVED | 판정 |
|---|---|---|---|
| TestReplayFallbackRetryAfterAbandonedAttempt | PASS (무해 증명) | **PASS** (0.07s) | seam-innocent |
| TestReplayConcurrentPublishDistinctTurns | PASS (결함이면 확정) | **PASS** (0.15s, 3라운드) | seam-innocent |
| TestReplayForkChainCrossing/required fork turn | 미지 (측정) | **PASS** (0.19s) | seam-innocent |
| TestReplayForkChainCrossing/empty fork turn | 미지 (측정) | **PASS** (0.11s) | seam-innocent |

- **3종 + race**: `go test -count=1 -race -run 'TestReplayFallbackRetryAfterAbandonedAttempt|TestReplayConcurrentPublishDistinctTurns|TestReplayForkChainCrossing' -v ./internal/gateway/translate/` → `ok ... 1.746s`, exit 0. verbatim: `.moai/reports/t846/seam-tests-race.log`
- **변경 패키지 전체 + race**: `go test -count=1 -race ./internal/gateway/translate/` → `ok ... 3.498s`, exit 0. verbatim: `.moai/reports/t846/package-race.log`
- **lint**: `golangci-lint run --new-from-rev=93ae49ce7 --timeout=5m ./...` → `0 issues.`, exit 0. verbatim: `.moai/reports/t846/lint.log`
- **gofmt**: `gofmt -l internal/gateway/translate/` → 빈 출력, exit 0.
- **크로스 빌드**: `GOOS=windows GOARCH=amd64 go build ./...` → 출력 없음, exit 0.
- **안정성 재검증**: Test 2를 `-count=20 -race`로 20회 반복 → 20/20 통과. verbatim: `.moai/reports/t846/flake-post-fix-20x.log`

## Baseline-attribution (baseline 귀속)

- 모든 수치는 본 워크트리(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t846`, 브랜치 WT-receipt-seam-red, 커밋 fbcddfa0b + 본 테스트 파일 작업 사본)에서 2026-09-14 이번 실행으로 관측한 출력이다.
- lint 베이스는 디스패치 명시값 93ae49ce7(브랜치 분기점)을 사용했다. 참고: 실행 시점의 로컬 develop은 다른 레인 병합으로 241c9dbad로 전진해 있었으나 본 브랜치 트리와 무관하므로 분기점 베이스가 정확한 귀속이다(`git merge-base --is-ancestor 93ae49ce7 HEAD` 확인).
- Test 1의 시나리오 재현 수단: `seamDeadWriter`(모든 Write를 실패시키는 io.Writer) — 터미널 발행이 첫 클라이언트 관측 emit에 선행한다는 t707의 코드 순서 증명(stream.go: `c.response(r)` → `publishOutput=true` → emit)을 행동으로 고정한다.

## Gaps (미검증)

1. **라이브 CauseChain 400의 최종 국소화** — 본 카드 스코프 밖. 합성 재현 범위에서 seam은 무혐의이므로, 라이브 불일치의 바이트 국소화는 t838(MOAI_RECEIPT_DEBUG=1 계측)의 몫이다.
2. **t707 §3의 미완 독해**(81f68955 rows 5069-5241·5555-5767·14:07:19 분기) — 본 카드에서 다루지 않았다.
3. **포크-모델-스위치 교차**(gpt-5.6↔gpt-6-astra 혼용 대화에서 checkObserved의 alternate-domain required-충돌 경로) — 공개 API로 재현 가능한 설계 범위에서 벗어나 미측정. live 사례는 전부 단일 모델(gpt-5.6-sol)이므로 우선순위 낮음.
4. **동시 발행의 강제 오버랩** — start 게이트로 두 스트림을 동시 개시했지만 스케줄러가 직렬화할 가능성은 남는다(3라운드 × 20반복에서 전부 2후보 착지로 CAS 무유실은 확인).
5. **테스트 3 변형에서 빈 포크 턴의 alternate-domain 재생** — primary 도메인 재생만 확인했다.

## Residual-risk (잔여 위험)

- 합성 픽스처가 실제 upstream SSE의 바이트 수준 특이점(공백·유니코드 이스케이프·arguments 원본 문자열)을 재현하지 못할 수 있다 — t707과 동일한 한정. 라이브 계측이 이 위험을 제거한다.
- Test 2의 동시성은 race detector가 데이터 경합을 잡아주지만, flock 경합 경로의 타이밍 민감성(잠금 대기 중 context 취소 등)은 미검증 — store 단위 테스트의 소관.
- 테스트가 현재 행동을 고정하는 특성상, 미래에 required-충돌 룰이 포크 후보를 실제로 오조준하는 새로운 형상이 발견되면 Test 3는 그 결함을 잡지 못할 수 있다(측정 당시 무혐의만 증명).

## 부록 — 검증 중 발견·수정한 테스트 결함 (제품 결함 아님)

초기 구현에서 Test 2가 간헐 실패했다(`unpaired or duplicate tool result`, 20회 중 다수). 원인: 두 고루틴의 결과를 `results` 채널에서 **수신 순서**로 `assistants[g]`에 채워 브랜치 귀속이 어긋남 — assistant는 branch b인데 tool_result를 call_a로 붙이는 오귀속 덤프로 확정(`.moai/reports/t846/flake-pre-fix-repro.log`). outcome에 생산자 식별자를 실어 수정 후 20/20 통과. 이 과정에서 관측된 바: **두 동시 스트림의 발행·매니페스트 착지 자체는 한 번도 어긋나지 않았다**(실패 라운드에서도 항상 2후보) — flock+CAS의 무유실을 오히려 보강하는 관측이다. 진행 과정 기록: 메모리 교훈 "A/B 비교는 셀렉터 집합까지 고정해야 한다"와 동일 계열.

## 커밋

- 테스트 + 증거: `test(gateway): ...` (SHA는 리드 보고 참조) — 본 파일과 `.moai/reports/t846/*` 동봉
- 프로덕션 코드 변경: 없음

🗿 MoAI
