# SPEC-CODEX-GATE-SCOPING-001 — decision index

> `interview.decision_gate: on` (`.moai/config/sections/interview.yaml:6`) → 인터뷰에서 운영자가 정하지 않은 결정의 행 목록. 라벨 어휘: DECIDED / POLICY-COVERED / EVIDENCE-NEEDED / FOUNDER 고정.

### Q1: 선행 SPEC-CODEX-GATE-SCOPE-001 §F가 운영자 확정(2026-10-01)으로 둔 「primary 체크아웃 → 트리 스코프 현행 유지」 페일오픈 팔의 반전을 운영자가 추인하는가?

Label: FOUNDER
Authority anchor: — (반전 대상의 확인은 선행 SPEC `.moai/specs/SPEC-CODEX-GATE-SCOPE-001/spec.md` §F에 있으나, 그것은 질문의 출처지 이 행을 결정하는 원장이 아니다. 반전의 측정 근거인 처분 기록 9건은 primary 체크아웃 `.moai/reports/t1395/` 아래 카드 표면이라 권위 원장 축이 아니다)
Why unresolved: 선행 확인이 운영자 확정이다. 카드 t1404 배차가 수정 방향의 작업은 위임했지만, 완료된 선행 SPEC의 운영자 확정 팔을 측정 기반으로 되돌리는 추인은 본 SPEC 본문만으로 갚아지지 않는다.
Operator verdict: DECIDED — 추인됨 (리더 판정 2026-10-03, 운영자 게이트급 결정은 리더가 증거로 정함 — kickoff-autonomy 독트린). 출처 2건: (a) 카드 t1404 본문 「③ 수리 방향(운영자 승인 10-02): primary 체크아웃 비카드 검토 제외」, (b) 10-03 미션 배차(리더 전달). 이 두 출처가 선행 SPEC-CODEX-GATE-SCOPE-001 §F 페일오픈 팔 반전의 운영자 추인이다. 기록: 레인, 2026-10-03.

### Q2: primary-scope 정책 기본값 skip이 plain-checkout 단일 사용자 배치의 비카드 검사를 잠근다 — 기본값의 근거가 본 저장소 레짐 측정뿐이고 다운스트림 영향은 미측정이다

Label: EVIDENCE-NEEDED
Authority anchor: —
Why unresolved: skip 기본값의 근거는 9블록/38발견/0생존(공유 primary 레짐 측정)이다. 게이트가 전체적으로 opt-in(enabled 기본 OFF)이라 피해 상한은 낮지만, enabled를 켠 plain-checkout 운영자의 비카드 검사가 조용히 사라지는 영향은 측정된 바 없다. 복원 키(REQ-CGSC-004)가 완화 축이다.
Operator verdict:

### Q3: Facet 2의 카드 제시 선택지(검사 범위 배제 vs 드리프트 원장 사전분류)에 본 SPEC이 두 팔을 함께 요구했다 — 한 팔로 좁힐지는 운영자가 좁힐 수 있다

Label: FOUNDER
Authority anchor: —
Why unresolved: 카드가 either/or로 방향을 제시했다. 본 SPEC은 자기게이트 배제 팔(REQ-CGSC-007)과 혼합 케이스 재분류 팔(REQ-CGSC-008)을 모두 요구했다 — 배제만으로는 혼합 케이스를 못 막고 재분류만으로는 codex 호출 비용이 남는다. 두 팔 유지와 한 팔 축소는 운영자 판단 축이다.
Operator verdict: DECIDED — 두 팔 모두 유지 (리더 판정 2026-10-03). 배제 REQ-CGSC-007 + 재분류 REQ-CGSC-008을 함께 유지한다 — 한 팔만으론 혼합 케이스가 뚫리거나 codex 호출 비용이 남는다는 본 SPEC 자체 분석을 따름. Q2(EVIDENCE-NEEDED)는 복원 키 REQ-CGSC-004와 함께 기록만 한다. 기록: 레인, 2026-10-03.
