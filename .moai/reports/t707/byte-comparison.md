# t707 케이스 5 실측 기록 — replayed request vs recorded receipt chain

세션: family `1f14d174-9f5c-4626-a0d0-82a2bc9afa62` (메인 스레드, gpt-5.6-luna, 2026-09-13 13:16:07.235Z 400)
측정 방법: 실제 게이트웨이 코드(`receipt.CanonicalPrefixes`, `opaque`, `translate.observations` 논리)를
워크트리에서 실행 + 트랜스크립트 재구성 비교 스크립트(`live-prefix-forensics.py`).

## 1. 매니페스트 (receipt/manifest.json, generation 12)

| idx | opaque_required | items | complete | prefix(12) | previous(12) |
|-----|-----------------|-------|----------|------------|--------------|
| 0..4 | true | 1 | true | 6cca4b35a520 … a4839496daeb | 체인 연결 확인 |
| 5 | true | 2 | true | a114f5b3124f | = cand4 prefix |
| 6 | true | 1 | true | b6dcf5ea301d | = cand5 prefix |
| 7 | true | 4 | true | 20eb1cbff123 | = cand6 prefix |
| 8..10 | true | 1/1/2 | true | f4865de012ba / ec0efc02d4ad / f786a13fefc1 | 체인 연결 확인 |

- 11 후보 = 트랜스크립트의 11개 응답(resp_0bd770f1b7778bdb016…)과 1:1 대응. 후보 10 = 응답 11(redacted(11846) + Edit tool_use).

## 2. 응답 11 봉투–마커 결합 검증 (트랜스크립트 rows 180/181)

- 봉투: `moai_opaque_v2_`, items 2(output_index 0·1, reasoning, content/summary 비어있음), public 1(function_call@2)
- 봉투 digest `da12ad7efe8859649d94ab802c38892fa018a243eabbf36ad7c9a071a1381703`
- Edit 마커 opaque_sha256 = 동일 값 → **결합 일치**(마커 위조 아님)
- 매니페스트 후보 10 items=2 = 봉투 items → **발행 일치**

## 3. 트랜스크립트 재구성 prefix 재계산 (실제 Go 코드)

- 재구성: user/assistant 행 36개, assistant는 응답 id별 병합(11 경계 — 요청 2의 성공이 병합을 증명)
- domain = `moai-history-v1\0<familyUUID>\0<scope>\0gpt-5.6\0`, scope = sha256("moai-openai-replay-account-v1\0"+account) (계정 값은 메모리에서만 사용, 출력 안 함)
- manifest session digest = sha256("1f14d174-…") 확인 → 세션/패밀리 UUID 식별 완료

| boundary | prefix 일치 | prev 일치 | 비고 |
|----------|------------|-----------|------|
| 0 | 불일치 (0f5c18bc0be4 vs 6cca4b35a520) | 일치(공 뿌리) | **검증된 요청 2가 통과한 바로 그 이력** |
| 1..10 | 전부 불일치 | 전부 불일치 | 파급 |

→ **트랜스크립트 재생 바이트 ≠ 실제 요청 바이트**가 검증된 쌍에서 실증됨. 트랜스크립트는 요청 본문의 충실한 기록이 아니므로(첨부·system-reminder·명령 메시지의 주입 위치가 기록과 다름), 트랜스크립트 기반 역산으로 라이브 불일치 바이트를 국소화하는 것은 불가능. 요청 본문 덤프는 어디에도 존재하지 않음(게이트웨이 덤프 기능 없음 확인).

## 4. 패턴 실측 (4개 세션 전체)

| 세션:행 | 실패 직전 툴 | 툴 종료→실패 요청 간격 | id 형태 |
|---|---|---|---|
| dbe1dde2:8548 | Edit (opaque) | ~80ms | opaque |
| 1f14d174:1016 | Edit (opaque) | ~80ms | opaque |
| b8817007:3913 | Edit (raw call_) | — | raw |
| 81f68955:2836 | Edit replaceAll (opaque) | ~90ms | opaque |

- 네 고유 사례 모두 **Edit 실행 직후 첫 요청**에서 발생. Read/Bash/SendMessage 직후 요청은 전부 통과.
- 같은 대화에서 섞인 id 형태(dbe1dde2: SendMessage=raw, Bash=opaque) → "raw id 단독" 가설은 기각(리드 판정과 일치).
- 직렬 사례(1·5)는 재시도 없음(API attempt 1/11에서 턴 종료, 세션 내 retry 흔적 0).
- autocompact level=ok 상수 — 컴팩트 재작성 없음.

## 5. 합성 재현 결과 (커밋된 회귀 스위트, 전부 GREEN)

마커 id / raw call_id / 연속 reasoning 2+fc(실제 응답 11 형상) / 인터리브 reasoning / Edit replaceAll+유니코드 / 텍스트 응답 / 요약 포함 reasoning / 직렬 2턴 + 변조 3종(거부 유지) — 8형상 전부 publish→check 바이트 일치. 즉 **재현 가능한 범위에서 게이트웨이 직렬화·복원 경로의 결함 없음**.
