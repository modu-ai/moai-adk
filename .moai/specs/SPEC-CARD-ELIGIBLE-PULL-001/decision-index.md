# SPEC-CARD-ELIGIBLE-PULL-001 — decision index

`interview.decision_gate: on` — 조립 단계에서 노출된 결정 중 카드 t1839 본문이 확정짓지 않은 것을 한 행씩 기록한다. 이 인덱스는 status 축에 대해 무상태다(생명주기는 spec.md에만 산다).

### Q1: 플래그 OFF(기본값)에서 제네릭 암 (b)/(b2)는 오늘처럼 계속 임대하는가, 아니면 휴면하는가?

Label: FOUNDER
Class: product-level
Default: OFF = 오늘 동작 그대로 — 암 활성, 자격 게이트 비활성 (rule: preserves current behavior) — 판정 시점의 Default 규칙 적용 기록(기각됨)
Alternate: OFF = 제네릭 암 휴면 — 무소유 카드 임대 없음, 게이트는 ON에서만 암을 살림
Authority anchor: (작성 시점 — 커밋 트리 내 권위 원천이 이 질문을 담지 않았다. 해결 증거 쌍은 아래 판정 행에 기록: 카드 t1839 본문의 대우 + 리더 설계 문서 원칙 7/§3 목표 표)
Why unresolved: 카드 본문이 두 방향을 동시에 서술했다: 필수 동작 3번 "active ONLY when a self-dispatch config flag is ON"은 휴면을 시사하고, AC 방향 1번 "flag OFF — arms (b)/(b2) behave exactly as today"와 배달 제약 "flag OFF 동작은 현재와 byte-for-byte 동일"은 현행 유지를 서술한다 — 양립 불가. [철회] 해석 (i)의 닻이었던 "기존 테스트 무수정 GREEN"은 카드가 아니라 스폰 프롬프트 유래임이 확인되어 철회되었다.
Operator verdict: 해석 (ii) 채택 — 플래그 = 정책 노브(킬스위치): OFF(기본) ⇒ 암 (b)/(b2)는 자기 배정이 아닌 카드를 전혀 임대하지 않음(조용한 스킵); ON ⇒ 자격 게이트 뒤에서 활성. lane-resolved via decision ladder — disk evidence decisive (board/jev 불필요). 증거 쌍: (1) 카드 t1839 본문의 대우(contrapositive) "자기 배정이 아닌 카드는 … 통과할 때만, 자가 배차 설정이 켜진 경우에만 임대" ⇒ OFF ⇒ 제네릭 암 무임대; (2) 리더 설계 문서 원칙 7 + §3 목표 표 "generic arms only behind explicit eligibility filter (or removed)". decided_by: lane-17, tn1p4e 무인 위임 하에서; 리더 통보 완료. 이 행은 운영자 반환 리뷰를 위해 유지한다(삭제하지 않는다). SPEC 0.2.0이 이 판정으로 재작성되었다(REQ-CEP-001/002/004/006, AC-CEP-001 교체, plan.md §D 갱신 정책).
