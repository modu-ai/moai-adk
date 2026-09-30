# decision-index.md — SPEC-CODEX-GATE-SCOPE-001

`interview.decision_gate: on` — 카드 t1383 조립 중 표면화된 결정 중 레인 창의 운영자 방향이 직접 확정하지 않은 것. `Operator verdict` 는 작성 시점에 비어 있고 킥오프 게이트가 읽는다. 라벨 어휘: DECIDED / POLICY-COVERED / EVIDENCE-NEEDED / FOUNDER. **Q2·Q4·Q5 는 2026-10-01 운영자 확정(팩토리 리더 경유 AskUserQuestion)으로 verdict 가 채워졌다.**

## Q1: 레인 판별의 1차 신호를 카드 브랜치(WT- 접두사) 검출로 고정하는가?

- Label: POLICY-COVERED
- Authority anchor: `.claude/rules/moai/workflow/kanban-dispatch.md` § Isolation([HARD] 카드 워크트리 브랜치는 `WT-` 접두사 + 서술 slug) · `.claude/rules/local/gitflow-lane-protocol.md` §1
- Why unresolved: 커밋된 규율이 WT- 불변식 자체는 그대로 정한다. "게이트의 판별 입력으로 그 불변식을 쓴다"는 연결만 이 SPEC 이 새로 만드는 것이라 규율이 직접 답하지 않는다 — 불변식이 답하는 축은 이미 덮였으므로 POLICY-COVERED 다.
- Operator verdict:

## Q2: 세션 env 라벨을 판정 입력에서 제외하고 관측 맥락으로만 남기는가?

- Label: FOUNDER
- Authority anchor: 없음 — 근거가 라이브 큐 카드 t1373 본문(미커밋)에만 존재한다
- Why unresolved: /clear 뒤 env 라벨 잔존·retire 된 런의 고아 라벨 실측이 env 단독 판별의 신뢰성을 무너뜨리지만, 그 실측이 커밋된 어느 문서에도 없다. t1373 이 착지해 근거가 커밋되면 POLICY-COVERED 로 재판정 가능하다. 킥오프에서 리드·운영자 확정이 필요하다 — 결함을 강등하지 않는다.
- Operator verdict: **CONFIRMED**(2026-10-01, 팩토리 리더 경유 운영자 승인) — env 라벨을 판정 입력에서 제외하고 관측 맥락으로만 남긴다.

## Q3: 카드 diff 의 기저를 매 평가 시점의 `git merge-base develop HEAD` 로 다시 구하는가?

- Label: POLICY-COVERED
- Authority anchor: `.claude/rules/local/gitflow-lane-protocol.md` §8([HARD] CARD_BASE 재구성 — 핀 금지, 병합 전 평가 전용)
- Why unresolved: 규율이 레인 검증 판정식에 대해 그대로 답한다 — 게이트가 같은 식을 쓰는 것은 그 적용일 뿐 새 결정이 아니다.
- Operator verdict:

## Q4: 카드 diff 를 "커밋분(merge-base..HEAD) + 카드 트리 미커밋분"의 합집합으로 정의하는가?

- Label: FOUNDER
- Authority anchor: 없음 — 합집합 정의 자체의 선행 권위 없음(§8 은 기저 산출 규율만 커버한다)
- Why unresolved: 합의 형태는 리드 지시(카드 본문 R1 — 미커밋 큐)가 정했으나 지시문은 커밋된 권위가 아니다. 측정 형태(단일 `git diff <merge-base>`)는 plan.md §B.1 이 근거와 함께 제안한다 — 킥오프 확정 대상.
- Operator verdict: **CONFIRMED**(2026-10-01, 팩토리 리더 경유 운영자 승인) — 카드 diff = 커밋분(merge-base..HEAD) ∪ 카드 트리 미커밋분(runtime-managed 접두어 제외)의 합집합으로 확정.

## Q5: 두 신호가 모두 없는 세션을 트리 스코프로 떨어뜨리는가 (긍정 증거 활성화)?

- Label: FOUNDER
- Authority anchor: 없음 — "현행 동작 유지" 방향은 REQ-MCP-012 fail-open 정신과 같은 방향이지만, 스코프 클래스의 페일오픈 방향을 정한 커밋 문서는 없다
- Why unresolved: plan.md §B.2 가 (가) 트리 스코프를 (나) 카드 스코프의 조용한 무력화 실패 모드 대비 근거와 함께 제안한다 — 제안이지 확정이 아니다. 킥오프에서 확정한다.
- Operator verdict: **CONFIRMED**(2026-10-01, 팩토리 리더 경유 운영자 승인) — 식별 불가 세션은 트리 스코프로 페일오픈(긍정 증거 활성화)으로 확정.
