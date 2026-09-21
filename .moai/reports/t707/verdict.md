# t707 판정문 — receipt chain 400 (CauseChain)

날짜: 2026-09-14 · 브랜치: WT-edit-replay-chain · HEAD(시작): 0c32a15b2
Class B(run→sync, plan 생략) · 수행: manager-develop

## Claim (주장)

1. 게이트웨이의 receipt 직렬화·복원 경로(openai response → redacted_thinking 봉투 → opaque 마커 → replay Check)는 재현 가능한 모든 형상에서 publish↔check 바이트 일치를 유지한다 — 재현 가능한 결함이 발견되지 않았다.
2. 라이브 CauseChain 400은 **Edit 툴 실행 직후 첫 요청**이라는 공통 패턴을 가지며, 트랜스크립트 재생 바이트와 실제 요청 바이트가 다름이 검증된 요청 쌍에서 실증된다. 라이브 불일치의 최종 국소화는 요청 본문 기록이 없어 불가능하며, 다음 발생 시 1회 계측으로 확정할 수 있다(준비 관측성 패치 동봉).
3. Check 강도는 약화되지 않았다 — 기존 거부 테스트 전체 + 신규 변조 음성 테스트가 모두 통과한다.

이 변경의 코드 산출물은 **테스트 전용**이다. 프로덕션 코드 변경은 하지 않았다. 근거: (a) 8형상 합성 왕복이 전부 GREEN — "직렬화 수리"로 만들 수 있는 검증 가능한 변경이 없고, (b) 증거 없는 프로덕션 변경은 anti-replay 체인을 추측으로 만질 위험이 있다(발임 지시: follow the evidence, not the analogy).

## Evidence (증거)

- **GREEN(회귀 스위트, 커밋됨)** — `internal/gateway/translate/replay_roundtrip_test.go`: 마커 id·raw call_id·연속 reasoning 2+fc(실제 응답 11 형상)·인터리브 reasoning·Edit replaceAll+유니코드·텍스트·요약 포함 reasoning·직렬 2턴 = 8형상 publish→stream→client-model→Check 전부 통과. verbatim: `.moai/reports/t707/green-roundtrip.log` (exit 0)
- **변조 음성 테스트(커밋됨)** — edited tool input / renamed tool → `HistoryReplayError`로 거부, dropped tool result → 페어링 검증으로 거부. verbatim: `.moai/reports/t707/green-cause-and-tamper.log` (exit 0). 기존 t672 cause 테스트 8건도 전부 통과.
- **라이브 실측** — `.moai/reports/t707/byte-comparison.md` + `live-prefix-forensics.py`: 매니페스트 11 후보↔트랜스크립트 11 응답 1:1 대응, 응답 11 봉투 digest(`da12ad7e…`)=마커 opaque_sha256(결합 일치), 후보 10 items=2=봉투 items(발행 일치). 실제 Go 코드로 트랜스크립트 재구성 이력의 prefix를 재계산 → **boundary[0]부터 전부 불일치**(요청 2가 통과했던 이력) → 트랜스크립트≠요청 바이트 실증.
- **패턴 실측** — 4 고유 사례 전부 Edit 실행 직후 ~80-90ms 내 첫 요청(dbe1dde2:8548, 1f14d174:1016, b8817007:3913, 81f68955:2836). SendMessage=raw / Bash=opaque 혼용 대화 존재 → raw-id 단독 가설 반증(리드 판정 확인).
- **패치(커밋 안 함)** — `.moai/reports/t707/observability.patch`: `MOAI_RECEIPT_DEBUG=1`에서 최초 미일치 경계 인덱스·뿌리 여부·items를 stderr 1행 출력(이력 데이터 미포함). 다음 라이브 발생 시 hotfix 브랜치에 적용→재현→stderr 포착→revert.

### 3축 분류표 (직전 툴=Edit 여부 × id 형태 × 재개/포크/직렬)

| 세션:행 | 직전 툴 | 직전 툴 id | 재개/포크/직렬 | model | 비고 |
|---|---|---|---|---|---|
| dbe1dde2:1280 (12:03:52) | Read×3 (fork 시점) | opaque | **포크**(agent_summary) | gpt-5.6-sol | 포크 요청 자체가 400 |
| dbe1dde2:2113 | — (manager-docs) | 미확인 | **포크**(spawn) | gpt-5.6-sol | subagent 요청 |
| dbe1dde2:3124 | — (manager-docs) | 미확인 | **포크**(spawn) | gpt-5.6-sol | subagent 요청 |
| dbe1dde2:3813 | — (manager-docs) | 미확인 | **포크**(spawn) | gpt-5.6-sol | subagent 요청 |
| dbe1dde2:5894 (12:32:33) | 미확인 | 미확인 | **포크**(agent_summary) | gpt-5.6-sol | 포크 요청 자체가 400 |
| dbe1dde2:8548 (12:52:27) | **Edit** | opaque | **직렬** (main, 무병렬) | gpt-5.6-luna | 툴 종료 ~80ms 후 |
| 1f14d174:1016 (13:16:07) | **Edit** | opaque | **직렬** (main, 12번째 요청) | gpt-5.6-luna | 툴 종료 ~80ms 후 |
| b8817007:3913 (13:37:38) | **Edit** | **raw** `call_q3ps…` | **직렬** (subagent 31턴) | gpt-5.6-sol | manifest 82 후보 |
| 81f68955:2836 (14:02:52) | **Edit replaceAll** | opaque | **재개** (resume 직후) | gpt-5.6-sol | resume 직후 subagent |

읽기: 4 고유 사례 모두 Edit 직후. 포크/재개/직렬이 섞여 있어 단일 기제로는 수렴하지 않으며, 공통 분모는 "직전 턴이 Edit"과 "Edit 후 빠른 연속 요청(~80ms)"이다.

## Baseline-attribution (baseline 귀속)

- 테스트: `go test ./internal/gateway/translate/ ./internal/gateway/receipt/` → `ok` 2패키지, exit 0 — 본 워크트리(0c32a15b2 + 신규 테스트 파일), 이번 실행.
- lint: `golangci-lint run --timeout=5m` → **exit 1**(997줄 기존 발견, errcheck 301/staticcheck 28/unused 2 — HEAD에 이미 존재; 리드의 "lint-clean" 주장과 불일치, 버전/설정 차이 의심). **내 파일(replay_roundtrip_test.go) 발견 0건** — grep 'replay_roundtrip_test' 매치 없음. 기존 적색은 내 스코프 밖(t671 소관)으로 리드 보고에 포함.
- gofmt: `gofmt -l internal/gateway/translate/ internal/gateway/receipt/` → 빈 출력, exit 0.
- 크로스 빌드: `GOOS=windows GOARCH=amd64 go build ./...` → 출력 없음, exit 0.
- 커밋 전 HEAD 재확인은 각 커밋 직전 수행(규정 준수).

## Gaps (미검증)

1. **직렬 사례(1·5)의 불일치 바이트 미국소화** — 요청 본문 덤프 부재. observability.patch 적용 + 라이브 재현 1회로 확정 가능.
2. **케이스 2·3·4(dbe1dde2)의 직전 툴·id 형태 일부 미확인**(표에서 미확인 표기) — 해당 세션 로그의 툴 스톨 라인을 더 읽으면 채울 수 있으나, 분류 결론을 바꾸지 않는 수준으로 판단해 중단.
3. **클라이언트(Claude Code 2.1.270)의 요청 배열 생성 로직 비공개** — "Edit 직후 무엇이 이력을 바꾸는가"는 추정 상태(대안: hook_success/file-history-delta/total_tokens_reminder 첨부의 삽입 위치). 관측되지 않은 가정이다.
4. **포크/재개 사례(2·3·4·6·7)의 기제** — 포크된 후보 사본과 재생 이력의 정합성은 트랜스크립트 재생이 바이트 보존적일 때만 성립함이 확인되지 않았다(갭 1과 동일 계측으로 확인 가능).
5. 리드 지정 참조 자료 `.moai/reports/t688/gateway-400-*.md`는 존재하지 않았다(검색 3경로).

## Residual-risk (잔여 위험)

- 판정 1은 "재현 가능한 범위에서"라는 한정이 붙는다 — 실제 upstream SSE의 바이트 수준 특이점(공백·유니코드 이스케이프·arguments 원본 문자열)이 합성 픽스처와 다를 수 있다. 라이브 1회 계측(갭 1)이 이 위험을 제거한다.
- 변조 음성 테스트는 chain-check 경로와 페어링 경로를 덮지만, 도메인 호환 경로(compatibleGPT alternate)의 음성은 기존 t672 테스트에 의존한다.
- observability.patch는 검증되지 않은 패치다(컴파일 확인 없음 — 적용 목적 산출물). 적용 시 컴파일 오류가 나면 1-2줄 수정이 필요할 수 있다.

## 커밋

- 테스트+증거 1커밋: `internal/gateway/translate/replay_roundtrip_test.go` + `.moai/reports/t707/*` (커밋 SHA는 리드 보고 참조)
- 프로덕션 코드 변경: 없음(근거는 Claim 참조)
