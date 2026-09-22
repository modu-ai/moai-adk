# SPEC-WORKTREE-GC-001 — Compact (auto-generated digest)

> REQ + AC + 파일 + 제외 요약. 정본은 spec.md / plan.md / acceptance.md.

## Identity

- id: SPEC-WORKTREE-GC-001 · version 0.1.0 · status draft · tier M · phase "v3.2.0 target"
- OPERATIONS 카드: 제품 코드 없음. run 페이즈가 감사·정리 실행.

## Requirements (16 REQ, 5 modules)

- M1 인벤토리: REQ-WGC-001(측정 쌍), 002(제외 16트리), 003(측정 오염 금지)
- M2 판정: REQ-WGC-004(fetch 선행), 005(T1 = origin 조상), 006(T2 = 판정 기록), 007(T3 = 유지+보고), 008(dirty/lock/점유 예외)
- M3 반출: REQ-WGC-009(반출 선행 조건·rescue 목적지·인용 동반), 010(트리아지)
- M4 처치: REQ-WGC-011(T1 제거+ref 삭제), 012(T2 제거+ref 보존), 013(미해소 T2 → blocker)
- M5 검증: REQ-WGC-014(제거 행렬), 015(prune 뒤 + listing diff), 016(verdict.md 5-섹션)

## Acceptance Criteria (12)

AC-WGC-001 T1 종단 · 002 T2 ref 보존 · 003 T3 유지 · 004 제외 무결성+집산식 · 005 측정 쌍 · 006 더티 tracked · 007 lock 순서 · 008 세션 점유 · 009 미해소 T2 blocker · 010 반출 트리아지·목적지 · 011 prune 순서 · 012 판정 형식.

## Artifacts

- spec.md (정본 요구) · plan.md (M1-M5, 제약 C1-C12, REAPER 관계) · acceptance.md (AC 행렬) · spec-compact.md (본 문서) · progress.md (§E 신호)

## Exclusions (Out of Scope)

- Go 제품 코드 변경 없음
- CI·인프라 편집 없음
- primary 체크아웃·제외 16트리·타인 트리 기록 없음
- git push 없음(리드 일괄)
- plan 페이즈의 제거 실행 없음(run이 실행)

## Key baselines (2026-09-22 measured)

544 worktree rows · du 144,978,644 KB (~138.3 GB) · local develop 7f86971fc vs origin/develop f5fff2190 · 1차 정리로 9트리 이미 제거 · t1055에서 ~210디렉터리 구출 전례.
