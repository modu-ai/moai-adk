# decision-index.md — SPEC-DOCS-DELEGATION-CWD-001

`interview.decision_gate: on` (.moai/config/sections/interview.yaml) — 레인 플로우라 인터뷰 라운드가 없었으므로, 계획 단계에서 드러난 미확정 결정을 이 표가 운영자에게 올린다. 상태 축 없음 — SPEC 수명은 spec.md frontmatter 만이 소유한다.

### Q1: 레인이 싱크 산출물을 직접 쓰는 경로를 흐름에서 완전히 배제하는 것이 맞는가?

Label: FOUNDER

Authority anchor: 없음 — 카드 t1387 본문이 제약("소유 매트릭스 유지")을 줬지만 카드 텍스트는 커밋된 권위 등록부 표면이 아니므로 이 표에 앵커로 인용하지 않는다.

Why unresolved: t1383 은 실제로 레인 직접 실행(소유 예외)으로 닫혔고, 이 SPEC 은 그 경로를 트리거+기록 의무가 붙는 최후 수단으로 격하한다. 배제가 맞다는 판단은 standing spawn authority(kanban-dispatch-detail.md § Factory in-lane 3-stage — tk8hce 결함)에 근거하지만, "예외 트리거를 몇 회 실패로 잡을지"와 함께 운영자가 비준할 운영 정책이다.

Operator verdict: **APPROVED (배제 비준)** — 레인의 싱크 산출물 직접 실행을 흐름에서 완전히 배제한다; 소유 예외는 REQ-DSC-008(2회 구조적 실패 + 거부 증거 트리거, 기록 의무 동반)로만 생존한다. 근거: operator, kickoff round 2026-10-01, card t1387 lane-1 pane direct answer.

### Q2: 소유 예외의 트리거 기준을 "2회 구조적 실패 + 거부 증거"로 고정하는 것이 맞는가?

Label: FOUNDER

Authority anchor: 없음 — 측정 사례가 t1383 한 건이라 기준값의 표본이 1이다.

Why unresolved: 1회 실패에 즉시 예외를 허용하면 standing spawn authority 가 무력화되고, 3회로 올리면 레인이 불필요하게 멈춘다. "2회 + 거부 증거"는 t1383 기록에서 읽어낸 잠정값일 뿐 운영자가 정한 값이 아니다.

Operator verdict: **CONFIRMED (2회 + 거부 증거 확정)** — 소유 예외의 트리거를 두 번의 구조적 실패 + 거부 증거로 확정한다. 이 판정은 위 Why 의 "잠정값" 표기를 대체한다 — 값은 이제 운영자 확정이고, 표본 1 유의점은 출처(provenance)로만 남는다. 근거: operator, kickoff round 2026-10-01, card t1387 lane-1 pane direct answer.

### Q3: 화해된 에이전트 브랜치의 원격 착지 뒤 처분은 기존 sweep 정책을 따르는가?

Label: POLICY-COVERED

Authority anchor: `.claude/rules/moai/workflow/worktree-integration.md` § Hoist a tree's evidence before disposing it — 동일 커밋 트리에서 직접 판독 확인(본 SPEC 작성 시점).

Why unresolved 아님: WT- 개명된 에이전트 브랜치의 착지 뒤 처분(hoist 전제 + sweep 후보 편입)은 기존 정책이 이미 답한다 — 새 정책 없이 REQ-DSC-004/007 이 그 정책을 인용할 뿐이다.
