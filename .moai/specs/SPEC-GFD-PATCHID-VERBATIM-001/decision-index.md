# decision-index — SPEC-GFD-PATCHID-VERBATIM-001

> `interview.decision_gate: on` (.moai/config/sections/interview.yaml) — plan-phase 조립 중 표면화된 결정 중 인터뷰에서 운영자가 정하지 않은 것의 원장. 상태 축은 이 원장에 없다 — SPEC 의 생애는 spec.md 단 하나가 운반한다.

### Q1: 계층 2(누적 patch-id)의 비교 모드는 무엇인가?

Label: FOUNDER
Class: implementation-level
Authority anchor: 해당 없음 — FOUNDER (이 질문을 정한 설정 키·헌장 조항·완료 SPEC 의 HISTORY/Amendments 행이 없다. 전임 SPEC-GITHUB-FLOW-DEFAULT-001 은 in-progress 이고 --stable 을 고른 행위 자체를 결정 행으로 기록하지 않았다)
Why unresolved: 공백 정규화 결함은 카드 t1561 RED 로 처음 관측됐고, 대안 셋(--verbatim 전환 / --stable 유지+공백 정규화 전처리 / 계층 2 폐기·계층 3 의존) 중 무엇을 취할지 운영자가 정한 바가 없다. 배차 지시는 --verbatim 을 CANDIDATE 로 준 것까지다.
Default: git patch-id --verbatim 양면 적용 + 지원 감지 시임 (rule: smaller user-visible surface — 단일 플래그 교체에 감지 시임 하나가 붙는 최소 변경 면; 전처리안은 diff 스트림 변형 면이 넓고 계층 폐기안은 호출자 계약을 흔든다)
Alternate: --stable 유지 + 패치 공백 정규화 전처리
Operator verdict: DEFAULT-APPLIED 2026-10-06T21:36:06Z manager-spec (card t1561 lane)

### Q3: 세션 종료 ②팔(`git cherry` 동치)의 수리 형상은 무엇인가?

Label: FOUNDER
Class: implementation-level
Authority anchor: 해당 없음 — FOUNDER (제2 지점은 plan-audit 1회차에 처음 표면화됐고, 이를 정한 등재 원천이 없다)
Why unresolved: cherry 의 내부 patch-id 동치는 정규화 모드를 인자로 받지 않아 "복속"이 불가능하다 — 팔을 공유 verbatim 동치 술어로 교체하거나, 팔을 cannot-answer 전용으로 강등해 ③에 위임하는 두 형상이 갈린다. 강등은 개별 커밋이 upstream 에 각자 존재하는 카드의 확인력을 잃는다(기존 확인 시맨틱 후퇴).
Default: worktree 패키지에 커밋별 공백 충실 동치 술어를 공개해 ②팔이 그것으로 `git cherry` 를 대체 (rule: preserves current behavior — 커밋별 동치 확인 시맨틱을 공백 충실로 유지; undo 는 이 SPEC 자체 revert 한 번)
Alternate: ②팔을 cannot-answer 전용으로 강등 (확인력 상실 — 기존에 확인되던 착지가 preserve 로 바뀐다)
Operator verdict: DEFAULT-APPLIED 2026-10-06T22:57:03Z manager-spec (card t1561 lane)

### Q2: 로컬 git 이 공백 충실 모드를 지원하지 않을 때 계층 2 는 무엇으로 답하는가?

Label: FOUNDER
Class: implementation-level
Authority anchor: 해당 없음 — FOUNDER (미지원 git 환경의 계층 2 거동은 이번에 처음 다뤄지는 상태다 — 이를 정한 등재 원천이 없다)
Why unresolved: fail-closed(cannot answer → preserve, 계층 3 계속)와 가용성 유지(--stable 조용히 하락)가 갈린다. 전자는 오래된 git 에서 보수적으로 보존하고 후자는 결함의 재현 경로를 남긴다.
Default: cannot answer 오류 → 호출자 preserve, 계층 3 이 판정 (rule: preserves current behavior — 파일 헤더와 전임 REQ-GFD-002 가 유지해 온 "답할 수 없는 계층은 결코 착지가 아니다" 현행 계약)
Alternate: --stable 로 조용히 하락 (가용성 유지, 결함 재현 경로)
Operator verdict: DEFAULT-APPLIED 2026-10-06T21:36:06Z manager-spec (card t1561 lane)
