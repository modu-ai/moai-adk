# M2 맥락 장부 쓰기 전후 비교 (정제본)

원본: `.moai/reports/agentic-factory/m2-ctxledger-before-after.md` (로컬 전용). 계약: AF-EFFICIENCY-S52. 우월성 판정이 아니다.

조건: 같은 시험·같은 가짜 원격(`measure_ctxledger_test.go` @ `ecb19d2f99`), 서로 다른 텍스트 5회 갱신 뒤 같은 텍스트 2회 재제출. 후 코드는 lane-20 `8663240cd5`, 전 코드는 직전 논리의 복사본(별도 커밋 아님).

| 지표 | 전 | 후 | 상태 |
|---|---|---|---|
| 목록 조회 | 5 | 1 | 실측 |
| 총 호출 | 10 | 6 | 실측 |
| 같은 텍스트 재제출의 추가 호출 | 0 | 0 | 실측 |
| 댓글 1개·최신 텍스트·비밀 게시 0 | 유지 | 유지 | 실측(fixture) |
| 응답 유실 대사 | 유지 | 유지 | 실측 |
| 원격 댓글 삭제 뒤 첫 Flush | 댓글 1개로 복구 | 오류 없이 끝났으나 댓글 0개, 다음 Flush에서 복구 | 실측, 회귀 |
| 토큰 / 시간 / 구 Factory 대비 | | | UNKNOWN / NOT_RUN / 비교 불가 |

회귀 재현 시험: `real_ctxledger_staleid_red_test.go` (afred, 현재 FAIL). 한계: 페이지네이션·조건부 조회·실제 rate limit은 미측정.
