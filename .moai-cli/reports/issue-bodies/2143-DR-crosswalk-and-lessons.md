# [DR][07] 공통 부록 — 상태 어휘 대응표·경로 표·인터페이스/명령 대응표·시험 ID 대응 + §20 교훈 표 재작성안

## 배경
종합 보고서 §2 S4·S3·S11, §4 교차 규칙 1, §5 빠진 교훈 10건, §3 과거 설계 6편 행(`review-group4-past-designs.md`). 20편이 각자 다시 그리는 표를 **한 벌의 부록**으로 만들고, 상위 §20 "이전 설계 교훈" 표를 행별로 다시 쓴다.

## 범위
- `docs/design-review/appendix/state-vocabulary-crosswalk.md`: 문서별 상태 어휘([83f] 8, [25f] 10, [e08] 10, [3d9], [66a] 11, [0259] 14, [0d45] 3벌, Core Brain 3기계) ↔ RC1 어휘(receipt `src/update.rs`, admission `src/plan.rs`, claim `src/claim.rs` ACTIVE_STATES/outcome, sync `skills/moai/workflows/sync.md`, verify `verify.md`) — `docs/protocol-v1.md` crosswalk를 기준으로 확장
- `appendix/paths.md`: 홈·프로젝트·common-dir 경로 "RC1 출하 / 설계 목표(P09)" 두 열 + 용어 4종 정의
- `appendix/interfaces.md`: 문서 인터페이스명(`result.admit`/`evidence.admit`/`verify.accept`/`owner.handoff`/`decision.propose`/`sync.*`/`artifact.pin`/`dispatch.*`) ↔ RC1 OPERATIONS(`src/main.rs` 표 전수) + 내부 CLI 명령 4체계 대응
- `appendix/test-ids.md`: 문서 시험 ID 8체계 ↔ 상위 §17 15군 ↔ RC1 시험 파일/receipt
- `fixes/top-section20-lessons.md`: §20 표 각 행의 (계승/비계승 현재 문구 → 교체 문구 + 문서 줄 + "RC1 재채택" 열), 교훈 10건을 본문 어느 절에 넣을지와 완성문
- 모든 RC1 인용은 현재 tip에서 재확인

## 수용 기준
- AC-1: 4개 부록 모두 "문서 어휘 → RC1 어휘" 방향 전수(검토 파일이 열거한 어휘 누락 0)
- AC-2: RC1 열의 모든 항목이 소스 상수·표·문서에서 확인(가설 0)
- AC-3: §20 재작성안이 6개 문서 행 전부 + 교훈 10건 완성문
- AC-4: 개인 절대 경로·비밀 0건, check set 실측 exit

## 사용자 결정
- 2026-11-09 00:1x 보고서 컨펌
