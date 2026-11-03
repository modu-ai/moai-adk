# M1 장부 쓰기 실측 요약 (정제본)

원본: `.moai/reports/agentic-factory/m1-ledger-write-measurement.md` (로컬 전용). 계약: AF-EFFICIENCY-S52. 이 요약은 우월성 판정이 아니다.

| 항목 | 값 | 상태 |
|---|---|---|
| 측정 코드 | `internal/agenticfactory/e2e/measure_ledger_test.go` @ `3523f103f1` | 실측 |
| 갱신 5회의 gh 호출 | 읽기 5 + 쓰기 5 = 10 | 실측 (가짜 gh, 댓글 1개) |
| 같은 갱신 2회 | GET 2, POST 1, PATCH 1, 댓글 1개 | 실측 |
| 응답 유실 뒤 재시도 | 댓글 1개, POST 1회 | 실측(표본 1) |
| 비밀 게시 / 쓰기 비활성 시 호출 | 0 / 0 | 실측(fixture) |
| 토큰 | 모델 호출 없음 | UNKNOWN(해당 없음) |
| 시간 | | NOT_RUN |
| 구 Factory 대비 | 대응 경로 미확인 | 비교 불가 |

발견: 쓰기마다 댓글 목록 조회가 선행되고, 본문이 같아도 PATCH를 보낸다. 안전은 유지된다. 개선 후보는 소유자(18/20) 몫이며 아직 적용하지 않았다.
한계: 실제 GitHub의 페이지네이션·rate limit·응답 크기는 반영되지 않았다.
