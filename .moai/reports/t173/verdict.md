# t173 — moai update 청소 경로 링크 인식 plan-phase verdict

Class C (전체 3단계 진행). worktree t173 / branch `WT-clean-links`
(base origin/main @ 4b2f203fe).

- SPEC: `SPEC-CLI-CLEAN-SYMLINK-001` (Tier M, era V3R6, v0.1.1, status draft
  → run 진입)
- 커밋: `18f7cfc19`(측정 dossier 508행) · `075672146`(SPEC 4종) ·
  `31d338bc4`(감사 1차 수정 D1-D4 + 감사 보고서) · `cd5bfb60e`(2차 PASS
  감사 + §F) — 미푸시
- evidence: `.moai/reports/t173/measurements.md` + `.moai/specs/SPEC-CLI-CLEAN-SYMLINK-001/`(SPEC + audit-plan 2반복 보존)

## 판정 요약

| 항목 | 결과 |
|---|---|
| §A 실측 | **출하 결함 실측**(Run D): 관리 뿌리의 dangling 링크 → clean "Skipped (not found)" → deploy MkdirAll EEXIST → update 부분 파괴 중단 + 수동 rm 전까지 영구 재현. 라이브 링크는 중단 없으나 무인식(백업 0·출력 무언). t81 "Lstat 치환"은 4증거로 어느 트리에도 부재(기각된 제안) 확인 |
| plan-audit | **PASS 1.00 (iter-2, 최종)** — 1차 FAIL 0.92(하위 카운트 드리프트 + 계약 용어 미고정, 둘 다 1-2줄) → 수정 → 2차 전 항목 해소. must-pass 7/7 |
| 리드 비준 | FX-1 라이브 디렉터리 링크 = 제거+진행줄 가시화(MkdirAll fast-path 실측 근거) · FX-3b 글로브 dangling = 제거 — 2026-08-22 승인. 판별자 = "t81(가) §D에 기록된 한계, 후속 카드 후보(리드 큐)"로 정정 라우팅 |
| 설계 | 링크 전용 분기(mode&ModeSymlink를 IsDir 앞, Lstat) + 5형태 처분표 + both-pole AC + D2-D4 이관(형태 일치 [HARD]·3면 카운트 일치·비공허 단언) + t81(가) 교차 계약(REQ-CSL-009, "보유"=렌더링 목적지로 고정) |

## run 인계

- serial / manager-develop / cycle_type=tdd. M1(Run D → Go RED 전환 + 링크
  분기+처분) → M2(5형태 양극 fixture) → M3(계약 테스트) → M4(회귀+플랫폼
  t.Skip)
- [HARD] t81(가)와 같은 release/v3.1.3 착지
- 감사 residual: R1(REQ-CSL-009 §C 문구 미러링 — 미래 수정 시) ·
  run-phase 확인 목록(재실행 루프·파일 뿌리 dangling·비-darwin)

## Gaps

- 2차 감사 스코프 1.00은 델타 한정(1차 4건 해소 검증) — 전면 재감사 아님
- 감사관 M4 정직성: 리드 비준 원문 verbatim 대조는 불가(입력에 없음) —
  교차 산출물 정합성으로 대체 검증
- audit_multi 미실행(t171 워크트리 사각) — 방법론 노트에 기록
