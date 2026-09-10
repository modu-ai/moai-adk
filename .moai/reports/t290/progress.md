# 카드 t290 진행 기록 (lane-2)

- **배차 경로 변경 사유**: 지정 wt `.claude/worktrees/t290`이 이미 PR #1671 본체
  (`WT-glm-settings-rename`, 커밋 2건)로 점유 → 원본 보존 의무상 폐기 불가, 인접 신규 트리
  `t290-apply`로 우회. 리드 보고 완료(A안 확정 메시지에 같이 전달).
- **A안 확정**(리드): 패치로 진행, PR #1671은 리드가 superseded-by-t290으로 폐쇄 예정.
- **통합 락 불신 규율**(리드, t298 등록): `acquire` 기록 pid가 CLI 단명 프로세스라 status가
  항상 reclaimable로 보임 → 머지 전 리드에게 한 줄 알림 + "창 비었음" 확인 후 진입.

## 충돌 해소 결정 기록

3파일(각 1훙크) 전부 ours(develop·t289) 채택:
- closed_sets.go: `ValidGLMModels()` 반환 서열 flash-first 유지
- defaults.go: tier 기본값 4종 = DefaultGLM53Flash 바인딩 유지 + theirs의 max-only 설명 주석 흡수
- glm_tier_test.go: want 리스트 flash-first 유지 + theirs의 max-only 주석 한 줄 추가

근거: 카드 목적은 웹 콘솔 rename+lock 잔여 구출이며 기본값 재판정 아님. t289(#1668)는 리뷰를
거쳐 배송된 제품 진실 — theirs 수용은 미승인 회귀가 됨.

## 검증 요약

gofmt clean · config 3.1s · web 3.3s · cli 197.9s 전부 exit 0 · 윈도 크로스빌드 OK.
상세는 verdict.md.

## 현재 단계

⏳ 코드+테스트+증거 커밋 직후 — 리드의 primary 정리 확인 대기. 병합은 리드 공지 방식
(창 확인)으로 진행 예정.
