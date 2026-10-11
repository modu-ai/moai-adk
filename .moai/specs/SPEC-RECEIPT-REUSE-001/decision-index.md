# SPEC-RECEIPT-REUSE-001 — Decision Index

> `interview.decision_gate: on` (.moai/config/sections/interview.yaml) — plan-phase 조립에서 드러났으나 인터뷰에서 운영자가 정하지 않은 결정의 원장. 상태 축은 Stateless — SPEC 수명은 spec.md만 운반한다.

### Q1: 순차 재사용을 막는 기제는 무엇인가 — 단일-생존 종료-이벤트 경계(재범위), 인용-소비 원장, 인스턴스 신원 브리지, 앵커 전진 중
Label: FOUNDER
Class: implementation-level
Authority anchor: (해당 없음 — 선행 SPEC·설정·헌장 어느 것도 이 질문을 정하지 않는다. 모본 SPEC-CODEX-AUDIT-GATE-AXES-001은 이 가드를 창설했으나 인스턴스 경계 기제는 다루지 않는다.)
Why unresolved: anonymous 이벤트만으로는 겹침 구간의 start↔stop 귀속이 판정 불가능하다(iter1 D2 반례: A·B 시작, B가 r 주조, 무귀속 FAIL stop, C 시작, r 인용 PASS stop — 두 판독이 바이트 동일하게 상반 판정을 요구). 그래서 REQ-RR-001/REQ-RR-004가 단일-생존 종료 / 모호 종료 동결로 재범위됐고, 남는 기제 선택이 plan.md §D.3 표의 트레이드오프로 남는다.
Default: 재범위형 후보 1 — 단일-생존 종료-이벤트 경계(종료 시 생존 개시 1개일 때만 전진, 모호 종료는 동결) (rule: the option that preserves current behavior — 모호 종료 동결이 동시성 테스트가 못박은 t1544 의미를 문장 변경 없이 유지하는 유일한 후보)
Alternate: 후보 2 인용-소비 원장 (수락-선행 경로 한정 보조)
Operator verdict: DEFAULT-APPLIED 2026-10-07 manager-spec (iter2 수리 라운드 재적용)

### Q2: 재사용 거부의 원인 표기는 무엇인가 — 신규 전용 상수 vs 기존 원인 재사용
Label: FOUNDER
Class: implementation-level
Authority anchor: (해당 없음 — 원인 어휘 집합은 internal/auditreceipt/store.go가 소유하고 확장 정책을 명문화한 규칙은 없다.)
Why unresolved: 원인 상수 추가는 store.go와 거부 문구의 공개 어휘를 넓히는 선택이라 내부 전용으로 끝나지 않는다.
Default: 신규 전용 상수 추가 (rule: the option that preserves current behavior — 기존 원인들의 의미와 소비 규칙을 그대로 둔다)
Alternate: `CauseReceiptBeforeStart` 재사용 — 원인 구별 요구(REQ-RR-003)와 충돌해 기각 경향
Operator verdict: DEFAULT-APPLIED 2026-10-07 manager-spec

### Q3: 경계 기록의 수명은 무엇인가 — 세션-귀속 무-TTL vs 청소 주기
Label: FOUNDER
Class: implementation-level
Authority anchor: (해당 없음 — 세션 id 재사용 여부에 대한 프로젝트 규칙이 없다.)
Why unresolved: 기록 청소는 저장소 위생과 안전의 균형인데, 세션 id 재사용 관측이 없어 필요성이 측정되지 않았다.
Default: 세션-귀속, 별도 TTL 없음 (rule: the option with the smaller user-visible surface — 기록이 세션 id와 함께 무해화되고 청소 경로가 사용자에게 노출되지 않는다)
Alternate: TTL 기반 청소 — 세션 id 재사용이 관측될 때만 근거가 생긴다
Operator verdict: DEFAULT-APPLIED 2026-10-07 manager-spec
