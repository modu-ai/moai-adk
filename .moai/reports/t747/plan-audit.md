# t747 plan-audit — SPEC-AC-ANCHOR-SCOPE-001 · iter 1/2

판정: **PASS** · 점수 **0.81** (Tier M PASS 임계 0.80) · 측정 트리: `WT-ac-anchor-scope` @ `8665f80b3` (이번 실행)
audit_model 미설정 확인(.moai/config grep 공백) — Claude 전용 감사. plan-auditor(opus/high) 수행.

## 필수 통과 (7/7)

- MP-1 REQ 번호 일관성 PASS — REQ-ACAS-001..006 연속·갭 없음 (spec.md:71,77,82,88,93,98)
- MP-2 GEARS 형식 PASS — 요구사항 레이어 기준 6개 모두 패턴 일치 (AC Given-When-Then은 검증 레이어)
- MP-3 YAML 프론트매터 PASS — 12 정식 필드·타입·거부 별칭 없음; `related_specs:`는 옵션 표 밖이나 코퍼스 350 SPEC 관행+린트 허용(정보)
- MP-4 언어 중립성 N/A — 단일 모듈 Go 파서 SPEC
- MP-5 D7 교차-SPEC 조정 PASS — 참조 5 SPEC-ID 전원 존재, status 충돌 없음
- MP-6 D8 교차 플랫폼 PASS — syscall 0
- MP-7 확인 게이트 PASS — NEEDS CLARIFICATION 마커 0, research.md 불요

## 카테고리 (0-1)

명확성 0.75 · 완전성 0.75 · 테스트 가능성 1.0 · 추적성 0.75

## 악의적 질문 6개 답변 요지

1. 두 축 인코딩 충실 — 표본 수준 클레임 전부 냉동 아티팩트에서 재확인(AC-COLLECTOR-ANCHOR 4·CC297 19·STATUS-AUTO 25; empty 1~28). 주의: "all 9 via fallback"은 프로브의 likely 근사 과장(D3).
2. 제어 집합 정의 밀폐 — 냉동 보완 ∪ 냉동 목록, 재파생 **정확히 129**(152 decl-bearing, 14∩9=0).
3. 혼합 문서 경계 교차 없음 — 지역 자격(영문+콜론형) 스코핑 확대는 제거 아님; 제어 메커니즘이 산문 흡수를 능동 포착.
4. fallback 교체 구현 가능 정확도 — return -1 기각 정당(9파일 앵커 없음으로 강등일 뿐), 수준 의미 fixture 고정.
5. 규약 깨끗 — 단, 커밋된 프로브 헤더가 "커밋 전 삭제"로 남아 있고 자신을 zz_ 이름으로 지칭(D4).
6. 마일스톤 건전(M1 RED→M2 좁음→M3 헐거움→M4 제어+산문→M5 문서), 비행 전 냉동 기준선 재유도 t528 학습 올바름.

## 무결성 재유도 (이번 실행)

860 ✓ 462 ✓ 152 ✓ 1405 ✓ 1240 ✓ 165 ✓ — 헤드라인 8수치 전부 spec.md §A 정확 재생산, 1405=1240+165 ✓.

## 결함 D1-D7 (전원 사소·선택적 — PASS 판정에 영향 없음)

- D1 spec.md:147-149 — ~129–142 수치의 plan.md 귀속 오류(grep 공백); 보완 재파생=129. 수정: 수치 귀속 정정 또는 냉동 목록 보완 유도 문구.
- D2 acceptance.md:22,111 — AC-747-007 고아(§D 추적). 수정: REQ-ACAS-007 승격 또는 §D.4 제외-추적 규약 문서화.
- D3 spec.md:43,55 — fallback 9파일 단정이 프로브 likely 근사 과장. 수정: "빈 vocab 절 앵커; fallback 메커니즘은 likely" 문구 또는 M1 프로브가 파일당 ≥2 vocab 제목 실확인.
- D4 probe 헤더 — "커밋 전 삭제" 문구+zz_ 자가명명이 커밋된 현실과 모순. 수정: 보존 결정 기록.
- D5 spec.md:121-124 — "~165 out-section" 산술 느슨(결함 파일 자체 decl 포함; 비결함 잔여는 더 작음). 수정: 비결함 수치 재계산 또는 의도 명시.
- D6 empty-anchor.txt — `_archive/` SPEC 포함이 본문 미언급. 수정: §A에 균일 처리 한 문장.
- D7 acceptance.md:41-47+plan.md:92 — AC-747-003 비교 도구(냉동 열) 미명명. 수정: 냉동 기준선 열 절 추가(zz_t528 패턴).

## 권고

**PASS.** D1·D3·D7을 실행 전에 수정하면 M1/M4 재해석 논쟁이 제거됨. Implementation Kickoff Approval 게이트는 본 판정의 영향 없음(리드 소관).

— 보고서 텍스트는 plan-auditor가 전달한 것을 lane이 `.moai/reports/t747/plan-audit.md`로 보존함(감사관은 하네스 규약상 파일 미작성).
