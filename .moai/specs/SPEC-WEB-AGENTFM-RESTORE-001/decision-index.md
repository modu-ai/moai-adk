# decision-index.md — SPEC-WEB-AGENTFM-RESTORE-001

`interview.decision_gate: on` (`.moai/config/sections/interview.yaml:6`). 계획 조립 중 드러났으나 인터뷰에서 운영자가 확정하지 않은 결정만 행으로 올린다. SPEC 본문(spec.md §C)과 plan.md §D에 각 결정의 설계 처분은 이미 기록되어 있다 — 본 파일은 그 처분이 운영자 비준을 아직 받지 못했음을 상태 없이 추적한다(상태 축은 spec.md frontmatter 단일 소유).

## Q1: 템플릿 `llm.yaml`에 `profile`·`agent_overrides` 키를 다시 수재하는 것이 REQ-AMI-013 위반을 정당하는가?

Label: FOUNDER
Authority anchor: (해당 없음 — SPEC-AGENT-MODEL-INHERIT-001 REQ-AMI-013은 현재 금지 방향이며, 본 행은 그 금지를 되돌리는 행위 자체가 의제다)
Why unresolved: 카드는 "설정 스키마·저장 경로·UI 노출" 복원을 지시하지만, 저장 경로가 쓰는 키를 배포 템플릿이 다시 실어 나르는 것(REQ-AFR-008)까지 카드가 승인하는지는 명시되지 않았다. 측정 사실: `template.ShippedRetiredModelKeys`가 템플릿 수재 키를 strip에서 자동 제외하므로(§D.1), 수재하지 않으면 매 update가 콘솔이 쓴 키를 벗겨낸다 — 수재가 사실상 유일한 정합 option이지만, 템플릿 표면 변화는 사용자 프로젝트에 가는 변화라 운영자 비준 대상이다.
Operator verdict: 승인 (2026-10-02 운영자, lane AskUserQuestion 라운드) — 배포 템플릿 `llm.yaml`에 `profile`·`agent_overrides`(빈 기본값) 재수재를 비준한다. REQ-AMI-013 수정안이 이 승인에 동반된다 — 수정안 사슬에서 해당 금지 조항을 대체(supersede)한다. §C 기본 처분과 일치 — 설계 변경 없음.

## Q2: 복원된 `llm.agent_overrides`를 스폰 시점에 소비(런처가 에이전트별 model/effort를 고정)하는 계약을 언제, 어떤 SPEC으로 만드는가?

Label: FOUNDER
Authority anchor: (해당 없음 — 소비자 코드는 `3fa8bd2ab`에서 소멸했고 상속 기본을 규정하는 SPEC-AGENT-MODEL-INHERIT-001은 이 질문에 침묵한다)
Why unresolved: 본 SPEC은 콘솔 저장 표면까지만 복원하고 런타임 소비를 Out of Scope로 명시했다(spec.md §Out of Scope — 런타임 소비). 그러면 오버라이드는 "저장되지만 스폰에 영향 없는" 상태가 되는데, 이 공백을 후속 SPEC으로 언제 메울지(또는 영구 저장 전용으로 둘지)는 제품 방향 결정으로 운영자 몫이다.
Operator verdict: 후속 카드 지금 발행 (2026-10-02 운영자, lane AskUserQuestion 라운드) — 스폰-소비(spawn-consumption) 계약 SPEC의 후속 카드를 리더 큐에 발행한다. 본 SPEC은 런타임 소비를 Out of Scope로 유지하며, 그 공백은 무기한 방치가 아니라 발행된 후속 카드 t1421(선행: 본 SPEC 착지·운영자 승인 리더 세션 재확인)로 경계 진다.

## Q3: 복원 표면의 배치를 llm (3rd Party LLM) 패널 서브섹션으로 확정하는가?

Label: FOUNDER
Authority anchor: (해당 없음 — 13탭 계약은 `internal/web/tab_layout_test.go` `wantTabOrder`와 `primary_surface_test.go` `>13<`이 규정하지만, 어느 패널에 놓을지는 어느 커밋 문서도 정하지 않는다)
Why unresolved: 디스패치 지시는 "기존 FieldDef 서브섹션 기계(또는 유사 마커 패턴) 공유 선호"까지이고 설계는 이를 따랐으나(spec.md §C-3), 그 기계를 놓을 패널 위치(llm 패널 vs workflow 패널 vs 14번째 탭)는 판단 여지가 남는다. 설계는 파일 정합성(llm.yaml 저장)을 근거로 llm 패널을 골랐으나, 운영자가 티어 차트 옆 workflow 패널 배치를 원하면 M4 이전에 수정이 가장 싸다.
Operator verdict: llm 패널 서브섹션 확정 (2026-10-02 운영자, lane AskUserQuestion 라운드) — 설계 초안(llm 패널 서브섹션)을 그대로 확정한다. 설계 변경 불요; 판정만 기록.

## Q4: 복원된 `llm.agent_overrides`의 스폰-소비를 어떤 계약 형태로 이 SPEC의 v0.3.0 수정안에 넣는가?

Label: FOUNDER
Authority anchor: 카드 t1421 본문 — "에이전트별 model/effort 고정을 실제 서브에이전트 스폰에 반영하는 계약 설계" (운영자가 큐에 발행한 요청 원문 — Q2가 발행한 후속 카드 자체). 기계 근거는 커밋 트리로 검증 가능: `internal/config/types.go:319`(스폰-경로 소비자 0), `internal/hook/agent_model_guard.go`(observe/advise 2층, 구 차단층 의도 제거).
Why unresolved: 카드는 소비 계약의 방향만 지시하고 계약 형태(상시 소비 vs 옵트인, 소비 주체, 키 위치)는 열어 두었다. 상시 소비는 REQ-AFR-002(상속 기본)와 충돌하고, 전역 env 핀(`CLAUDE_CODE_SUBAGENT_MODEL`)은 [1m] 자격 문제와 전역 핀 문제를 동반한다(model-policy.md § Inherit-by-Default Convention) — 형태 선택은 운영자 몫이다.
Operator verdict: **옵트인 고정** (2026-10-03 운영자, lane AskUserQuestion 라운드, 카드 t1421) — `llm` 섹션 단일 명시 키(기본 off)를 켠 세션만 스폰 시 소비하고, 오케스트레이터가 Agent() 호출에 설정된 model/effort를 전달한다(유일 승인 소비 경로). 키 없는 세션은 상속 기본 유지. Q2의 기존 판정(후속 카드 발행, 2026-10-02)은 유효 — 본 카드가 그 후속이며 이 공백을 소진한다. plan.md §I 설계와 일치 — 설계 변경 없음.
