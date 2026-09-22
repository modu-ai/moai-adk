# t167 — 카드 판정 기록 (verdict — 2026-08-22)

worktree t167 / branch `WT-diagram-profile-import`
(origin/main + origin/release/v3.1.3 병합 기반 — t165 내용 포함 트리).

## plan

SPEC-DIAGRAM-PROFILE-IMPORT-001 (Tier M, 16 REQ/16 AC — 상한): 1차 FAIL
0.84(증거 grep 부정확[naive 형태 오탐 6+2행]·Step-0 무이주 문장 미조정) →
판별 grep 재진술(수정 전 실행 관찰로 전부 0 확인) + REQ-13 호출자 명시 예외
1문장(AC-IMP-007 diff 스코프 검증) + D3-D5 → 재감사 **PASS 0.92** (0.84→0.92)
+ R1/R2 터치업(renumber 포인터 6건·§B1 문구). 커밋 03d77e4c3 → baef137ec →
70b69732d(§F serial).

핵심 설계: (a) design-dna에 diagram-profiles 참조 — `.design-dna/` 프로젝트
루트 저장(ManagedCleanTargets 7뿌리 밖 생존, 감사관 코드 인증), 마커 우선,
confirm-before-overwrite (b) 임포터 2종 = 절차형 옵트인 참조(스크립트 아님),
소스 비신뢰·충실도 원장·one-home 일괄 대체. t166 호환 = 착지 품질 수치 기준
계약 + AC-VERIFY-001 run-phase 대조. 아이콘 정규화 B-10 지연(기존 세트가 이미
24×24 currentColor).

## run (리드 판정 PASS)

AC 16/16: 15 PASS + AC-VERIFY-001 **PASS-WITH-GAP**(t166 미착지 재실측 —
기하 패턴 0, SVG07x/08x 0; 폴백대로 갭 기록, 진단 코드 미조작). 커밋 M1
28e7d76c5(프로파일 132행+SKILL 포인터) · M2 dd9487719(임포터 258행+번들
표+Step-0 예외) · M3 e61a7397a(catalog 재생성+neutrality 수리+§E) · M3b
6eb9e700c. 3플랫폼 빌드 0 · lint 0 · strict leak ok(예시 날짜 리터럴 1건
인런 수리 — 플레이스홀더로) · 미러 5/5 바이트 동일.

## sync + 종결

경량 close: CHANGELOG Added + t165 phase 라벨 정정 라이더(v3.2→v3.1.3) +
`completed` 전환 + §E.4(sync_commit_sha ed78b5c0e, 백필 6f0fe34e8).

## 통합 시점 항목 (리드 소관 목록)

- AC-VERIFY-001 갭 = t166이 release에 착지한 뒤 배치 시점에 두 카드가 모두
  있는 트리에서 VERIFY 대조 재실행으로 닫음 (t82 통합 측정·t175 GEM-008과
  같은 목록)
- 참고: R3(감사 NOTE — spec §5 비중복 괄호 서술 시점) 무행동
