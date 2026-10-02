---
id: SPEC-AUTONOMY-BATCH-GATE-001
title: "Design — 배치 게이트 요약"
version: "0.3.0"
created: 2026-10-02
updated: 2026-10-02
author: manager-spec
tier: L
---

# SPEC-AUTONOMY-BATCH-GATE-001 — Design

> 이 문서는 설계 결정과 기각한 대안을 소유한다. 요구사항은 `spec.md` §C, 검증은 `acceptance.md`, 근거와 측정은 `research.md`가 정본이고 여기서는 되풀이하지 않는다. 열린 행(`decision-index.md`)에 걸린 결정은 *초안의 읽기*를 그대로 적되 판정이 아님을 표시한다. 시간 추정은 쓰지 않는다.

## D.1 해법의 모양과 Go 판단 (spec.md B.1, B.7)

**결정.** 운영자 형태로 남는 게이트 라운드를 *제시 단위*로 묶는다. 바꾸는 것은 (1) 정본 절 하나(`auto-semantics.md` §9.2), (2) 그 절을 가리키는 포인터와 낡은 표현 정합, (3) 런처가 SessionStart에 주입하는 리더 공지의 문장 하나, (4) 이들의 안전 불변식이 조용히 사라지지 않게 하는 반증 가능한 가드 테스트 둘이다.

**Go 판단.**
- 제품 Go 변경은 리더 공지 문자열에 한정된다(운영자 판정 Q5, 2026-10-02). 파일은 정확히 넷: `internal/hook/session_start_kanban.go`(칸반 리더 공지 조립), `session_start_kanban_i18n.go`(칸반 로케일 표), `session_start_factory.go`(팩토리 리더 공지 조립), `session_start_factory_i18n.go`(팩토리 로케일 표). 변경은 로케일 표에 문장 필드 하나와 네 로케일 값을 더하고, 조립부에서 그 필드를 *기존 블록 안의 한 줄*로 합치는 것이다. 새 블록·새 분기·새 환경 변수·새 호출은 없다.
- 테스트 Go 파일은 셋이다: 기존 `session_start_kanban_i18n_test.go`(필드 빈값 점검 표에 새 필드를 덧붙임), 신규 `session_start_leader_gate_notice_test.go`(두 리더 공지 × 네 로케일 문장 규칙과 변이 거부), 신규 `internal/template/batch_gate_summary_doctrine_test.go`(정본 절 앵커와 변이 거부). 셋 다 런타임 동작을 바꾸지 않는다.
- 다중 카드 적용은 `moai factory decide <card>...`가 이미 한다(순회 루프 `internal/cli/factory_card.go:1469-1491`). 한 호출은 gate/choice 하나만 받으므로 같은 게이트 행끼리만 묶는 규칙(REQ-BGS-001)과 맞는다. 결정 기록은 줄 기록이라 CLI 없이 쓴다(형제 SPEC도 같은 선). MCP `factory_decide`는 카드 1장 단위이고 레인 세션에 거부되지만 리더는 레인이 아니므로 카드별 호출이 가능하다(`internal/cli/mcp_factory_card.go:70-79,205-217`).

**기각한 대안.** 새 일괄 동사, 결정 기록 자동 기록기, 요약 검증기를 Go로 만드는 안 — 필요성은 Q8로 열려 있고 이 SPEC의 범위 밖이다. 이 SPEC은 배치 기구 자체의 Go 강제 장치를 만들지 않는다.

**잔여 위험.** 강제 장치가 없어 오케스트레이터 규율에 기댄다(spec.md G-3). 알림 문장은 리더에게 규칙의 존재를 상기시킬 뿐 준수를 강제하지 않는다.

**열린 행.** Q8 — 초안의 읽기는 "문서 수준 + 사후 sync 감사의 재독"이다. 이것은 판정이 아니다.

## D.2 배치 단위와 구성원 (REQ-BGS-001~003)

**결정.**
- 키는 **같은 게이트 행**이다. Kickoff 행과 sync 승인 행은 증거 기준이 다르므로(Kickoff는 감사 PASS와 해시 불변, sync 승인은 sync-auditor 판정과 증거 임계값 — `auto-semantics.md` §9) 한 질문에 섞이지 않는다.
- **대기 중인 행 전부를 스냅숏으로 제시**한다. 요약을 채우려고 준비된 카드를 붙잡지 않는다(REQ-BGS-002). 늦게 준비된 행은 다음 요약이나 개별 질문이다.
- 하나의 판정이 여러 카드의 공통 결정 주체일 때(제안서의 "카드 5장이 한 판정에 직렬 대기" 사례, `research.md` R6 #2)는 같은 질문을 카드 수만큼 묻지 않고 한 번 묻고 영향받는 카드를 전부 적는다(REQ-BGS-003).
- 묶는 주체는 운영자 대화를 쥔 세션(리더)뿐이다. 레인은 카드 1장만 쥐므로 교차 카드 배치를 만들지 않는다.

**기각한 대안.**
- 채워질 때까지 대기시키는 안 — 대기가 곧 지연이고, 제안서가 줄이려는 직렬 대기를 새로 만든다.
- 게이트 행이 다른 카드를 한 질문에 섞는 안 — 행마다 증거 기준이 다르다.
- 레인도 교차 카드 배치를 만드는 안 — 레인은 자기 카드의 질문 채널을 리더를 통해서만 쥔다(`agent-common-protocol.md` "Lane sessions are orchestrator-class").

**열린 행(초안의 읽기, 판정 아님).** Q1 — 키를 게이트 행으로 두는 것과 공통 판정 주체로 두는 것을 둘 다 담았다. 게이트 행만 키로 하면 한 판정이 여러 카드를 막는 경우가 빠지고, 판정 주체만 키로 하면 같은 게이트의 증거가 다른 카드의 다행 요약이 빠진다. 둘 다 두면 검토자가 한 번에 볼 면과 가드 앵커가 늘어난다. Q2 — 예약 행의 동일 판정 1회 질문은 제외했다.

## D.3 요약 서식, 단일 승인, 빼내기, 기록 서식 (REQ-BGS-004·005·006·009)

**서식.** 보고 → 질문 순서는 `askuser-protocol.md`의 Report-Before-Ask를 따른다. 행마다 카드 id, 게이트 행, SPEC id, 감사 판정 참조(*최종 반복* 판정, 반복 식별자, 점수, PASS 기준 대비 여유), 산출물 해시 불변 확인, 카드 자신의 결정 기록 참조, keep-set·리더 보유 권한 점검 결과와 분류 근거, `counter_refs=`를 싣는다. 반대 증거가 있는 행을 먼저 놓는다.

**keep-set 점검 필드.** 예약 규칙은 리더가 행을 올바르게 분류할 때만 성립하는데, 분류가 행에 안 보이면 오분류된 행이 옳은 행과 똑같이 보인다. 그래서 *승인 가능* 행도 "keep-set 세 범주와 리더 보유 권한 목록에 걸리지 않음 + 근거(게이트 행의 §9 처분과 증거)"를 한 칸으로 싣는다. 제외 행을 보고서에 나열할지는 Q3로 열려 있으나, 이 칸은 어느 답에서도 성립한다 — 나열하면 제외 행도 같은 칸에 사유가 오고, 나열하지 않으면 승인 가능 행의 분류 근거가 운영자가 볼 수 있는 유일한 제외 기준이다.

**질문 하나와 승인 범위.** 질문은 하나이고 승인은 *나열된 승인 가능 행에만* 미친다. 나중에 도착한 행, 예약 행, 차단 행은 덮지 않는다고 보고서가 진술한다(REQ-BGS-005). 권고 라벨은 `interview.recommendation_mode`의 현행 규약을 그대로 따른다(`pull` 동안 라벨 없음).

**빼내기 수단.** 질문 채널은 질문당 선택지를 최대 4개로 제한하고(`askuser-protocol.md` Structural Constraints) "Other" 자유 입력을 자동으로 덧붙인다(`askuser-protocol.md` Free-form Circumvention Prohibition). 그래서 행이 N개일 때 행마다 선택지를 두는 방식은 N>3에서 성립하지 않는다. 행 수에 의존하지 않는 수단은 자동으로 붙는 자유 입력에 운영자가 빼낼 카드 id를 적는 것이다(REQ-BGS-006). 이것은 선호가 아니라 상한에서 나온 귀결이다.
- 기각: 행마다 선택지(N>3 불가), 다중 선택 목록(선택지 4개 상한에 걸려 N>4 불가), 질문을 여러 개로 쪼개는 안("정확히 한 질문"을 깬다).

**기록 서식.** 승인된 행마다 `auto-semantics.md` §10의 한 줄 기록에 두 가지를 더한다. 예시(자리표시자만):

```text
decision record: decided_by=<runner+role> evidence_refs=<paths+verdict-ids> ladder_path=<gate-row-slug>;batch=<YYYYMMDDTHHMMSSZ> counter_refs=<ids or none searched=<what>>
```

- `<gate-row-slug>`는 §9.2 본문이 정하는 게이트 행 이름의 소문자·하이픈 토큰이다(값 안에 공백을 두지 않아 한 줄 `grep`이 토큰으로 갈라진다).
- `batch=<id>`의 `<id>`는 결정 질문을 낸 UTC 시각(`YYYYMMDDTHHMMSSZ`)이다. 한 요약의 모든 기록이 같은 값을 갖고, 요약마다 다르다. 가정: 질문 하나가 운영자 응답을 기다리는 동안 다른 질문이 나가지 않으므로 두 요약이 같은 초에 발행되지 않는다(질문 채널의 직렬성에 기댄 가정이며 측정하지 않았다).
- 이 식별자로 M-2(질문 한 번당 결정된 카드 수)를 기록에서 계산하고, sync 감사가 "승인이 정확히 나열된 행에만 미쳤다"를 재확인한다(G-5). 식별자가 없으면 한 요약의 기록들을 묶을 수 없다.
- 결정 기록은 디스크 결정 보드(`auto-semantics.md` §11, 홈 표면)에 줄로 쓴다. 기록 없는 행은 승인으로 치지 않는다.

**기각한 대안.** 카드 id나 요약 시퀀스 번호를 식별자로 쓰는 안 — 정본이 템플릿으로 배포되므로 카드 id를 못 쓰고(R-2), 번호는 보관소 없이 단조성을 보장할 수 없다.

## D.4 반대 증거 의무 (REQ-BGS-007·008)

**결정.** 행마다 "진행하지 말아야 할 가장 강한 증거"를 `counter_refs=`로 적는다. 후보 원천은 닫힌 목록이다: 감사 경고·기록된 부채, PASS 기준 대비 여유, 미해결 decision-index 행, 감사 교차의 의견 불일치, 열린 차단·대기 기록, 같은 요약 안의 다른 행과의 경로 겹침. 못 찾았으면 `counter_refs=none searched=<검색 내용>`이다. 검색 내용 없는 `none`은 반대 증거 진술로 치지 않는다.

**열린 행(초안의 읽기, 판정 아님).** Q7 — 초안은 닫힌 목록을 택했다. 이유는 초안의 것이다: 목록이 닫혀 있으면 필드가 검사 가능하고 `none searched=` 주장이 감사 가능하다. 이 이유는 판정이 아니며, 열린 목록과 "닫힌 목록 + 기타" 형태의 비용은 `decision-index.md` Q7이 대칭으로 적는다.

## D.5 예약·차단·contract 행과 기록 시점 재확인 (REQ-BGS-010·011·012·019)

**예약 목록의 출처 — 출하된 문서가 이미 싣는 것만.**
- keep-set 세 범주: 환경상 불가능(실행 자체가 못 가는 경우 — 자격 증명·실행기 부재), 운영자 보유 작업, 외부 공유 시스템을 건드리는 되돌릴 수 없는 조작(`auto-semantics.md:154-158`).
- 리더 세션이 쥐는 권한: `kanban-dispatch.md:106`이 "stay with the leader session"로 적는 여섯 항목 — 최종 PASS/FAIL 판정, 최종 병합 승인(`LEAD-MERGE-APPROVED`), 운영자 게이트, 카드 발행과 `done`(`moai gtd` 변경), CodeRabbit 슬롯 대기 판정, 세션 간 분쟁 조정. 이 줄을 이번 개정에서 직접 읽어 목록을 확인했다(`research.md` R4.5).
- 0.2.0까지 REQ-BGS-010에 있던 "user-facing behavior changes"는 뺐다. 이 문구는 로컬 유지보수 문서 `.moai/docs/jev-local-operations.md:30`에만 있고 출하되지 않는다. 문자대로 읽으면 대부분의 기능 SPEC Kickoff 행이 사용자 대면 동작을 바꾸므로 전부 예약되어 배치가 비고, 좁게 읽으면 리더마다 분류가 갈린다.
- 열린 행 Q6은 열린 채다: 3등급 개별 승인 규칙이 출하 문서에서 *어디에* 서는가(keep-set 정의 + 권한 불변식만, 또는 리더 보유 권한 목록까지 지명, 또는 새 정의). 초안의 읽기는 앞의 둘(keep-set 정의와 `kanban-dispatch.md:106`)이다. REQ-BGS-010, 이 절, Q6이 이 하나의 원천에 일치한다.

**차단 규칙 — 양성형.**
- 행은 (a) 최신(최종 반복) 감사 판정이 PASS, (b) 계획 단계가 audit-ready를 기록, (c) 판정 이후 계획 산출물 해시 불변, (d) 열린 차단 없음, 이 넷이 모두 성립할 때만 승인 가능하다. `auto-semantics.md` §9.1의 조건 그대로다.
- §9.1 문면은 "the independent plan-audit verdict is PASS (FAIL and INCONCLUSIVE stay hard blocks ...)"이고, PASS-WITH-DEBT와 BYPASSED는 그 파일에 나오지 않는다(`grep` 적중 없음, `research.md` R4.2). `plan-auditor.md:203,231`의 기계 판정 어휘는 `PASS|PASS-WITH-DEBT|FAIL`이고 `spec-workflow.md`의 생략 계약은 PASS를 "NOT FAIL, NOT INCONCLUSIVE, NOT BYPASSED"로 정의한다. §9.1이 이 둘에 침묵하므로 문면의 PASS에 맞춰 차단으로 다룬다. 이것은 문면에서 나온 사실이다.
- 최종 반복 결속: 감사 보고서는 반복마다 한 파일이고 "가장 최근 판정"은 최종 반복의 것이다(`spec-workflow.md` § Report Persistence, `runtime.ResolveLatestPlanAudit`). 이전 반복의 PASS를 인용하면 현재 산출물을 덮지 않는다.
- 기각: 부정 열거형(FAIL·INCONCLUSIVE·부재만 차단) — 그 문단은 PASS-WITH-DEBT와 BYPASSED를 통과시켜 §9.1과 §7("only a positive verdict proceeds")을 약화한다. 변이로 반증된다(acceptance.md AC-005).

**기록 시점 재확인(REQ-BGS-019).** 해시 불변과 차단 없음은 보고서를 그릴 때 확인된다. 운영자 응답이 올 때까지 시간이 흐르고, 배치는 서로 다른 시점에 읽은 N행을 한 답이 덮는다. 단일 카드 Kickoff도 같은 창이 있으나 배치가 그 창을 넓힌다. 그래서 응답을 받은 뒤 기록을 쓰기 직전에 행마다 판정과 해시를 다시 읽고, 표류한 행은 승인으로 기록하지 않고 거부한다.
- 기각: 렌더 시점 확인만 두는 안(창이 열려 있다).

**contract 모드(REQ-BGS-012).** `workflow.autonomy.mode: contract`에서 plan→run 게이트는 계약 서명(`moai contract kickoff-check`)이며 요약 행이 생기지 않는다(`contract-autonomy.md` § The signing gate).

## D.6 정본의 집과 표면 (REQ-BGS-013·014)

**집.** `auto-semantics.md`의 §9.1 바로 뒤, 새 절 `### 9.2 The batch gate summary`. 이유: 게이트 목록(§9)·결정 기록 서식(§10)·권한 게이트 불변식(§7)이 같은 파일이다. 그 파일은 `paths: ".claude/skills/moai-lane-watchdog/**"` 범위라 세션 선두에 상시 로딩되지 않는다(`.claude/rules/moai/workflow/auto-semantics.md:1-3`). 정본 본문이 세션 prefix를 키우지 않는다. 템플릿 쪽에 바이트 동일 미러를 둔다.

**표면.** 다른 모든 표면은 포인터만 지닌다.
- `kanban-dispatch.md`(상시 로딩, 리더가 매 턴 읽음): Boundaries "No gate bypass." 불릿 안에서 문장을 다듬어 §9.2를 가리킨다. 상시 로딩이므로 순증가는 `rule-authoring.md`의 1,000바이트 기준 아래여야 하고, 보존 구간("Promotion is the operator's act, always." → "The self-dispatch lane exception.")은 바이트 단위로 건드리지 않는다.
- `run.md`: 기존 줄(`:137`)의 "§9.1" 인용을 "§9.1–9.2"로 바꾸는 식의 줄 수 불변 편집.
- `spec-assembly.md:202-208`, `workflows/moai.md:144,240`: 낡은 "mandatory·score-independent" 표현을 §9.1·§9.2에 맞춰 다시 쓴다(REQ-BGS-014, 운영자 판정 Q4). 같은 줄 수로, 보존 문구와 사전 차이 줄은 그대로.

**기각한 대안.** `kanban-dispatch.md`에 본문을 두는 안(상시 로딩 비용, 보존 구간 제약), 새 규칙 파일(`paths:` 없으면 상시 로딩 비용이 늘고 `paths:`가 있으면 리더가 읽을 길이 없다), `askuser-protocol.md`(상시 로딩이고 질문 규약과 섞인다).

**범위 경계.** REQ-BGS-014는 위 두 파일만 덮는다. 같은 종류의 낡은 문구를 가진 다른 파일은 읽어서 분류했고(`plan.md` 표면 목록, `research.md` R4.4) 이 SPEC은 고치지 않는다. 넓힐지는 Q12다.

## D.7 리더 공지 문장 (REQ-BGS-016·017·018)

**결정.** 두 리더 공지에 **모두** 싣는다. 근거: 두 공지는 배타적이라 한쪽에만 두면 다른 쪽 리더가 문장을 받지 못하고(칸반 공지는 팩토리 환경에서 빈 문자열 — `session_start_kanban.go:50-52`, `TestKanbanNoticeSuppressedUnderFactoryEnv`), 두 리더 모두 운영자 대화를 쥔다. 레인·동반 세션은 카드 1장만 쥐므로 받지 않는다. 방출은 `startup`에서만 있고 resume·clear·compact·fork에서는 없다.

**사본과 로케일.** 에이전트용 사본(`additionalContext`)에는 영어 값이, 운영자용 사본(`systemMessage`)에는 설정된 `conversation_language`의 값이 실린다. 에이전트용 사본은 구성상 영어이므로(`session_start_kanban.go:17-19`) 로케일 격자는 리더 둘 × 로케일 넷, 곧 여덟 조합이고, 영어 렌더가 곧 에이전트용 사본이다. 알 수 없는 로케일은 영어로 폴백한다.

**문장의 모양.** 한 문장. 무엇 — 같은 운영자 형태 게이트에서 대기 중인 카드가 둘 이상이면 카드별 질문이 아니라 배치 게이트 요약 하나로 제시한다. 어디 — `.claude/rules/moai/workflow/auto-semantics.md` §9.2. 서식·제외·반대 증거 의무는 적지 않는다: 정본을 가리킬 뿐이다(REQ-BGS-016).

**제약.** 기존 테스트가 고정한 블록 구조·역할 용어·형식 문자열 금지·카드 id/SPEC ID/날짜 금지를 깨지 않는다(`plan.md` §D 제약 11). 포인터 토큰은 네 로케일에서 번역하지 않는 주소로 다룬다.

**기각한 대안.** 칸반 리더 공지에만 싣는 안(팩토리 리더 누락), 레인·동반 공지에 싣는 안(카드 1장만 쥔다), 새 블록을 만드는 안(`TestKanbanLeadNoticeBlockLayout`의 5블록 구조를 깬다), 정본 서식을 문장에 옮겨 적는 안(두 벌이 되는 순간 갈라진다).

**비용.** 리더 세션의 `startup`에서만 한 줄이 늘고 비리더 세션은 0바이트다. 리더는 `kanban-dispatch.md`를 이미 매 턴 읽으므로 의도된 중복이다(spec.md R-3). 렌더 길이의 전/후는 run-phase에서 공시한다(플랜 단계에서는 측정하지 않았다).

**열린 행.** Q10 — 문장은 "batch gate summary"라는 이름을 쓴다. 초안이 이 이름을 택한 이유는 초안의 것이다: 이 이름은 승인 수단이 아니라 제시 형식을 가리킨다. 이 이유는 판정이 아니며 이름이 바뀌면 문장과 가드 앵커가 같이 바뀐다.
