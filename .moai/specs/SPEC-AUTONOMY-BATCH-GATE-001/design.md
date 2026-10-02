---
id: SPEC-AUTONOMY-BATCH-GATE-001
title: "Design — 배치 게이트 요약"
version: "0.4.1"
created: 2026-10-02
updated: 2026-10-02
author: manager-spec
tier: L
---

# SPEC-AUTONOMY-BATCH-GATE-001 — Design

> 이 문서는 설계 결정과 기각한 대안을 소유한다. 요구사항은 `spec.md` §C, 검증은 `acceptance.md`, 근거와 측정은 `research.md`가 정본이고 여기서는 되풀이하지 않는다. 열린 행(`decision-index.md`)에 걸린 결정은 *초안의 읽기*를 그대로 적되 판정이 아님을 표시한다. 시간 추정은 쓰지 않는다.

## D.1 해법의 모양과 Go 판단 (spec.md B.1, B.7)

**결정.** 운영자 형태로 남는 plan→run Kickoff 라운드를 *제시 단위*로 묶는다. 바꾸는 것은 (1) 정본 절 하나(`auto-semantics.md` §9.2), (2) 그 절을 가리키는 포인터와 낡은 표현 정합, (3) 런처가 SessionStart에 주입하는 리더 공지의 문장 하나, (4) 이들의 안전 불변식이 조용히 사라지지 않게 하는 반증 가능한 가드 테스트 둘이다.

**Go 판단.**
- 제품 Go 변경은 리더 공지 문자열에 한정된다(운영자 판정 Q5, 2026-10-02). 파일은 정확히 넷: `internal/hook/session_start_kanban.go`(칸반 리더 공지 조립), `session_start_kanban_i18n.go`(칸반 로케일 표), `session_start_factory.go`(팩토리 리더 공지 조립), `session_start_factory_i18n.go`(팩토리 로케일 표). 변경은 로케일 표에 문장 필드 하나와 네 로케일 값을 더하고, 조립부에서 그 필드를 *기존 블록 안의 한 줄*로 합치는 것이다. 새 블록·새 분기·새 환경 변수·새 호출은 없다.
- 테스트 Go 파일은 셋이다: 기존 `session_start_kanban_i18n_test.go`(필드 빈값 점검 표에 새 필드를 덧붙임), 신규 `session_start_leader_gate_notice_test.go`(두 리더 공지 × 네 로케일 문장 규칙과 변이 거부), 신규 `internal/template/batch_gate_summary_doctrine_test.go`(정본 절 앵커와 변이 거부). 셋 다 런타임 동작을 바꾸지 않는다.
- `moai factory decide <card>...`는 순회 루프를 이미 갖지만(`internal/cli/factory_card.go:1469-1491`) factory decide 행은 이 SPEC의 요약 행이 아니라 개별 질문이다(REQ-BGS-001, §D.8). 결정 기록은 줄 기록이라 CLI 없이 쓴다(형제 SPEC도 같은 선).

**기각한 대안.** 새 일괄 동사, 결정 기록 자동 기록기, 요약 검증기를 Go로 만드는 안 — 필요성은 Q8로 열려 있고 이 SPEC의 범위 밖이다. 이 SPEC은 배치 기구 자체의 Go 강제 장치를 만들지 않는다.

**잔여 위험.** 강제 장치가 없어 오케스트레이터 규율에 기댄다(spec.md G-3). 알림 문장은 리더에게 규칙의 존재를 상기시킬 뿐 준수를 강제하지 않는다.

**열린 행.** Q8 — 초안의 읽기는 "문서 수준 + 사후 sync 감사의 재독"이다. 이것은 판정이 아니다.

## D.2 배치 단위와 구성원 (REQ-BGS-001~003)

**결정.**
- 요약의 범위는 **plan→run Kickoff 행 하나**다(§D.8 R1). 나머지 게이트 행(factory decide 행들, sync 차단 승인, 카드 선택)은 개별 질문이다 — 행마다 증거 기준이 다르다(Kickoff는 독립 감사 PASS와 해시 불변, sync 승인은 sync-auditor 판정과 증거 임계값 — `auto-semantics.md` §9).
- **대기 중인 행 전부를 스냅숏으로 제시**한다. 요약을 채우려고 준비된 카드를 붙잡지 않는다(REQ-BGS-002). 늦게 준비된 행은 다음 요약이나 개별 질문이다.
- 하나의 판정이 여러 Kickoff 행의 공통 결정 주체일 때는 같은 질문을 카드 수만큼 묻지 않고 한 번 묻고 영향받는 카드를 전부 적는다(REQ-BGS-003). 그 질문은 예약 행(REQ-BGS-010)과 차단 행(REQ-BGS-011)을 이름에 올리지 않는다 — 올리면 공유 판정 질문이 차단 규칙을 우회하는 길이 된다. 제안서의 "카드 5장이 한 판정에 직렬 대기" 사례(`research.md` R6 #2)는 Kickoff가 아니라 served-model 판정이었고, Kickoff가 아닌 게이트의 공유 판정은 이 SPEC에서 REQ-BGS-003이 덮지 않는다(Q13).
- 묶는 주체는 운영자 대화를 쥔 세션(리더)뿐이다. 레인은 카드 1장만 쥐므로 교차 카드 배치를 만들지 않는다.

**기각한 대안.**
- 채워질 때까지 대기시키는 안 — 대기가 곧 지연이고, 제안서가 줄이려는 직렬 대기를 새로 만든다.
- 다른 게이트 행을 한 요약에 섞거나 범위를 모든 게이트 행으로 넓히는 안 — 행마다 승인 가능성 술어가 달라야 하고 이 SPEC은 Kickoff의 술어만 정의한다(§D.8, Q13).
- 레인도 교차 카드 배치를 만드는 안 — 레인은 자기 카드의 질문 채널을 리더를 통해서만 쥔다(`agent-common-protocol.md` "Lane sessions are orchestrator-class").

**열린 행(초안의 읽기, 판정 아님).** Q1 — 요약(REQ-BGS-001)과 공유 판정 1회 질문(REQ-BGS-003)을 둘 다 담았다. 요약만 두면 한 판정이 여러 카드를 막는 경우가 빠지고, 공유 판정 질문만 두면 증거가 서로 다른 카드의 다행 요약이 빠진다. 둘 다 두면 검토자가 한 번에 볼 면과 가드 앵커가 늘어난다. Q2 — 예약 행의 동일 판정 1회 질문은 제외했다. Q13 — 범위 확대 여부.

## D.3 요약 서식, 단일 승인, 빼내기, 기록 서식 (REQ-BGS-004·005·006·009)

**서식.** 보고 → 질문 순서는 `askuser-protocol.md`의 Report-Before-Ask를 따른다. 보고서 머리는 plan→run Kickoff 행임을 적고, 나열된 카드마다 한 행에 카드 id, SPEC id, 독립 plan-audit 판정 참조(*최종 반복* 판정, 반복 식별자, 점수, PASS 기준 대비 여유), 산출물 해시 불변 확인, 카드 자신의 결정 기록 참조, keep-set·리더 보유 권한 점검 결과와 분류 근거, `counter_refs=`를 싣는다. 반대 증거가 있는 행을 먼저 놓는다.

**keep-set 점검 필드.** 예약 규칙은 리더가 행을 올바르게 분류할 때만 성립하는데, 분류가 행에 안 보이면 오분류된 행이 옳은 행과 똑같이 보인다. 그래서 *승인 가능* 행도 "keep-set 세 범주와 리더 보유 권한 목록에 걸리지 않음 + 근거(§9 처분과 증거)"를 한 칸으로 싣는다. 예약·차단 행을 표에 나열할지 표 밖에서 보고할지는 Q3로 열려 있으나(차단 행을 차단으로 보고하고 승인에서 빼는 것 자체는 요구다, REQ-BGS-011), 이 칸은 어느 답에서도 성립한다 — 표에 나열하면 그 행도 같은 칸에 사유가 오고, 표 밖에서 보고하면 승인 가능 행의 분류 근거가 운영자가 표에서 볼 수 있는 유일한 제외 기준이다.

**질문 하나와 승인 범위.** 질문은 하나이고(한 호출에 요약 질문은 하나뿐) 승인은 *나열된 승인 가능 행에만* 미친다. 나중에 도착한 행, 예약 행, 차단 행은 덮지 않는다고 보고서가 진술한다(REQ-BGS-005). 권고 라벨은 `interview.recommendation_mode`의 현행 규약을 그대로 따른다(`pull` 동안 라벨 없음).

**선호 배출 비약화(REQ-BGS-020).** Kickoff 질문은 카드별 선호도 함께 배출한다(`workflows/moai.md:144` "All user preferences (tier, mode preference, PR strategy, chain scope) are drained at this gate", `run.md:141`). 한 번의 일괄 승인은 N장의 선호 집합을 배출할 수 없으므로, 요약 승인은 이 조건을 대신하지 않는다: 승인된 카드마다 선호가 run 진입 전에 디스크에 있어야 하고 그 출처는 카드·SPEC 계약이거나 그 카드를 위한 운영자 대화다(`orchestration-mode-selection.md` §C.3). 선호가 디스크에 없는 카드는 요약 승인 뒤에도 그 카드의 선호 질문이 따로 필요하다 — 일괄화가 줄이는 라운드의 한계다(G-1).

**빼내기 수단.** 질문 채널은 질문당 선택지를 최대 4개로 제한하고(`askuser-protocol.md` Structural Constraints) "Other" 자유 입력을 자동으로 덧붙인다(`askuser-protocol.md` Free-form Circumvention Prohibition). 그래서 행이 N개일 때 행마다 선택지를 두는 방식은 N>3에서 성립하지 않는다. 행 수에 의존하지 않는 수단은 자동으로 붙는 자유 입력에 운영자가 빼낼 카드 id를 적는 것이다(REQ-BGS-006). 이것은 선호가 아니라 상한에서 나온 귀결이다.
- 기각: 행마다 선택지(N>3 불가), 다중 선택 목록(선택지 4개 상한에 걸려 N>4 불가), 질문을 여러 개로 쪼개는 안("정확히 한 질문"을 깬다).

**자유 입력의 읽기 규칙(REQ-BGS-006).** 단일 선택 질문에서 "Other"를 고르면 답은 입력한 글뿐이다(`askuser-protocol.md`는 자동 "Other"를 자유 입력 수단으로만 기술한다. 명시적 선택지와 "Other" 글을 한 답에 함께 싣는 인코딩은 실행해 보지 않았다 — UNVERIFIED). 그래서 자유 입력만으로 "나머지는 승인"이 전해져야 한다. 읽기 규칙은 질문 문구가 적는다: 적힌 카드 id는 모두 빼내고 나열된 나머지 승인 가능 행은 승인한다. 읽을 수 없는 글이나 보고서에 없는 카드 id는 아무것도 승인하지 않고 같은 질문을 다시 낸다 — 잘못 읽었을 때 승인을 미루는 쪽으로 실패한다. 기각: "적힌 id만 승인" 읽기 — 요구가 지정한 수단은 빼내기이고 요구 문장이 말하는 쪽은 적힌 id를 빼내는 읽기다.

**기록 서식.** 승인된 행마다 `auto-semantics.md` §10의 한 줄 기록에 둘을 더한다 — `ladder_path`에 `;batch=<id>`, §10의 세 필드 뒤에 `counter_refs=`. 줄은 공백으로 나뉜 `이름=값` 토큰이므로 값에 공백이 없다(한 줄 `grep`이 토큰으로 갈라진다). 예시(자리표시자만):

```text
decision record: decided_by=<runner+role> evidence_refs=<paths+verdict-ids> ladder_path=<gate-row-slug>;batch=<YYYYMMDDTHHMMSSZ> counter_refs=<ids>
decision record: decided_by=<runner+role> evidence_refs=<paths+verdict-ids> ladder_path=<gate-row-slug>;batch=<YYYYMMDDTHHMMSSZ> counter_refs=none searched=<hyphenated-search>
```

- `<gate-row-slug>`는 plan→run Kickoff 행 이름의 소문자·하이픈 토큰이며 값은 §9.2 본문이 정한다.
- 반대 증거가 없는 행의 기록도 `counter_refs=none searched=<토큰>`이다(REQ-BGS-008이 보고서와 기록 양쪽에 걸린다). `searched=`가 없는 `counter_refs=none`은 반대 증거 진술이 아니며 M-3이 센다.
- `batch=<id>`의 `<id>`는 결정 질문을 낸 UTC 시각(`YYYYMMDDTHHMMSSZ`)이고 한 요약의 모든 기록이 같은 값을 갖는다. 다른 요약과의 구별은 측정하지 않은 가정에 기대지 않는다: 한 호출에 요약 질문이 하나뿐이고(REQ-BGS-005), 결정 보드에 같은 값이 이미 있으면 다음 빈 초로 넘어간다(REQ-BGS-009). 점검 뒤에 쓰인 기록과의 충돌은 남는 위험이다(G-5).
- 이 식별자로 M-2(질문 한 번당 결정된 카드 수)를 기록에서 계산하고, sync 감사가 "승인이 정확히 나열된 행에만 미쳤다"를 재확인한다(G-5). 식별자가 없으면 한 요약의 기록들을 묶을 수 없다.
- 결정 기록은 디스크 결정 보드(`auto-semantics.md` §11, 홈 표면)에 줄로 쓴다. 기록 없는 행은 승인으로 치지 않는다.

**기각한 대안.** 카드 id나 요약 시퀀스 번호를 식별자로 쓰는 안 — 정본이 템플릿으로 배포되므로 카드 id를 못 쓰고(R-2), 번호는 보관소 없이 단조성을 보장할 수 없다.

## D.4 반대 증거 의무 (REQ-BGS-007·008)

**결정.** 행마다 "진행하지 말아야 할 가장 강한 증거"를 `counter_refs=`로 적는다. 후보 원천은 닫힌 목록이다: 감사 경고·기록된 부채, PASS 기준 대비 여유, 미해결 decision-index 행, 감사 교차의 의견 불일치, 열린 차단·대기 기록, 같은 요약 안의 다른 행과의 경로 겹침. 못 찾았으면 `counter_refs=none searched=<공백 없는 토큰>`이다(보고서와 기록 양쪽). `searched=`가 없는 `counter_refs=none`은 반대 증거 진술로 치지 않는다.

**열린 행(초안의 읽기, 판정 아님).** Q7 — 초안은 닫힌 목록을 택했다. 이유는 초안의 것이다: 목록이 닫혀 있으면 필드가 검사 가능하고 `none searched=` 주장이 감사 가능하다. 이 이유는 판정이 아니며, 열린 목록과 "닫힌 목록 + 기타" 형태의 비용은 `decision-index.md` Q7이 대칭으로 적는다.

## D.5 예약·차단·contract 행과 기록 시점 재확인 (REQ-BGS-010·011·012·019)

**예약 목록의 출처 — 출하된 문서가 이미 싣는 것만.**
- 목록의 단일 원천은 REQ-BGS-010이다. 원천 문서는 keep-set 정의(`auto-semantics.md:154-158`)와 `kanban-dispatch.md:106`이며, 그 줄의 원문은 `research.md` R4.5가 가진다. 여기서 되풀이하지 않는다. 목록의 "operator gates" 항목이 요약 모집단(운영자 형태 plan→run Kickoff 행)을 포함하지 않는다는 한정도 REQ-BGS-010이 가진다 — 한정이 없으면 모집단 전체가 예약되어 요약이 비기 때문이다. 운영자 형태 Kickoff 행이 예약되는 길은 keep-set 범주뿐이고, 그 밖의 행은 요약 행이다.
- 0.2.0까지 REQ-BGS-010에 있던 "user-facing behavior changes"는 뺐다. 이 문구는 로컬 유지보수 문서 `.moai/docs/jev-local-operations.md:30`에만 있고 출하되지 않는다. 문자대로 읽으면 대부분의 기능 SPEC Kickoff 행이 사용자 대면 동작을 바꾸므로 전부 예약되어 배치가 비고, 좁게 읽으면 리더마다 분류가 갈린다.
- 열린 행 Q6은 열린 채다: 3등급 개별 승인 규칙이 출하 문서에서 *어디에* 서는가(keep-set 정의 + 권한 불변식만, 또는 리더 보유 권한 목록까지 지명, 또는 새 정의). 초안의 읽기는 앞의 둘(keep-set 정의와 `kanban-dispatch.md:106`)이다. REQ-BGS-010, 이 절, Q6이 이 하나의 원천에 일치한다.

**차단 규칙 — 양성형.**
- 행은 (a) 최신(최종 반복) 독립 plan-audit 판정이 PASS, (b) 계획 단계가 audit-ready를 기록, (c) 판정 이후 계획 산출물 해시 불변, (d) 열린 차단 없음, 이 넷이 모두 성립할 때만 승인 가능하다. `auto-semantics.md` §9.1의 조건 그대로다. §9.1은 "the independent plan-audit verdict"라 쓴다 — 독립은 plan-auditor가 낸 판정을 뜻하고 계획 산출물을 쓴 세션의 자기 진술 PASS는 판정이 아니다(REQ-BGS-011, A44).
- 차단 행은 차단으로 *보고*하고 승인에서 뺀다(요구, REQ-BGS-011). 보고가 요약 표 안에 오는지 표 밖에 오는지는 Q3의 열린 부분이며 이 규칙은 어느 쪽에서도 성립한다(§D.8 R2).
- §9.1 문면은 "the independent plan-audit verdict is PASS (FAIL and INCONCLUSIVE stay hard blocks ...)"이고, PASS-WITH-DEBT와 BYPASSED는 그 파일에 나오지 않는다(`grep` 적중 없음, `research.md` R4.2). `plan-auditor.md:203,231`의 기계 판정 어휘는 `PASS|PASS-WITH-DEBT|FAIL`이고 `spec-workflow.md`의 생략 계약은 PASS를 "NOT FAIL, NOT INCONCLUSIVE, NOT BYPASSED"로 정의한다. §9.1이 이 둘에 침묵하므로 문면의 PASS에 맞춰 차단으로 다룬다. 이것은 문면에서 나온 사실이다.
- 최종 반복 결속: 감사 보고서는 반복마다 한 파일이고 "가장 최근 판정"은 최종 반복의 것이다(`spec-workflow.md` § Report Persistence, `runtime.ResolveLatestPlanAudit`). 이전 반복의 PASS를 인용하면 현재 산출물을 덮지 않는다.
- 기각: 부정 열거형(FAIL·INCONCLUSIVE·부재만 차단) — 그 문단은 PASS-WITH-DEBT와 BYPASSED를 통과시켜 §9.1과 §7("only a positive verdict proceeds")을 약화한다. 변이로 반증된다(acceptance.md AC-005).

**기록 시점 재확인(REQ-BGS-019).** 승인 가능 조건은 보고서를 그릴 때 확인된다. 운영자 응답이 올 때까지 시간이 흐르고, 배치는 서로 다른 시점에 읽은 N행을 한 답이 덮는다. 단일 카드 Kickoff도 같은 창이 있으나 배치가 그 창을 넓힌다. 그래서 응답을 받은 뒤 기록을 쓰기 직전에 행마다 REQ-BGS-011의 네 조건 전부와 REQ-BGS-010 분류(운영자 hold 포함)를 다시 읽고, 표류한 행은 승인으로 기록하지 않고 거부한다. 재확인 범위는 승인 가능 술어와 같아야 한다 — 더 좁으면 판정과 해시 사이에 열린 차단, 철회된 audit-ready 기록, 운영자 hold가 읽히지 않는다.
- 기각: 렌더 시점 확인만 두는 안(창이 열려 있다), 판정·해시만 다시 읽는 안(나머지 조건이 표류해도 거부 문장이 발화하지 않는다).

**contract 모드(REQ-BGS-012).** `workflow.autonomy.mode: contract`에서 plan→run 게이트는 계약 서명(`moai contract kickoff-check`)이며 요약 행이 생기지 않는다(`contract-autonomy.md` § The signing gate).

## D.6 정본의 집과 표면 (REQ-BGS-013·014)

**집.** `auto-semantics.md`의 §9.1 바로 뒤, 새 절 `### 9.2 The batch gate summary`. 이유: 게이트 목록(§9)·결정 기록 서식(§10)·권한 게이트 불변식(§7)이 같은 파일이다. 그 파일은 `paths: ".claude/skills/moai-lane-watchdog/**"` 범위라 세션 선두에 상시 로딩되지 않는다(`.claude/rules/moai/workflow/auto-semantics.md:1-3`). 정본 본문이 세션 prefix를 키우지 않는다. 템플릿 쪽에 바이트 동일 미러를 둔다.

**표면.** 다른 모든 표면은 포인터만 지닌다.
- `kanban-dispatch.md`(상시 로딩, 리더가 매 턴 읽음): Boundaries "No gate bypass." 불릿 안에서 문장을 다듬어 §9.2를 가리킨다. 상시 로딩이므로 순증가는 `rule-authoring.md`의 1,000바이트 기준 아래여야 하고, 보존 구간("Promotion is the operator's act, always." → "The self-dispatch lane exception.")은 바이트 단위로 건드리지 않는다.
- `run.md`: 기존 줄(`:137`)의 "§9.1" 인용을 "§9.1 and §9.2"로 바꾸는 식의 줄 수 불변 편집.
- `spec-assembly.md:202-208`, `workflows/moai.md:144,240`: 낡은 "mandatory·score-independent" 표현을 §9.1·§9.2에 맞춰 다시 쓴다(REQ-BGS-014, 운영자 판정 Q4). 같은 줄 수로, 보존 문구와 사전 차이 줄은 그대로.

**기각한 대안.** `kanban-dispatch.md`에 본문을 두는 안(상시 로딩 비용, 보존 구간 제약), 새 규칙 파일(`paths:` 없으면 상시 로딩 비용이 늘고 `paths:`가 있으면 리더가 읽을 길이 없다), `askuser-protocol.md`(상시 로딩이고 질문 규약과 섞인다).

**범위 경계.** REQ-BGS-014는 위 두 파일만 덮는다. 같은 종류의 낡은 문구를 가진 다른 파일은 읽어서 분류했고(`plan.md` 표면 목록, `research.md` R4.4) 이 SPEC은 고치지 않는다. 넓힐지는 Q12다.

## D.7 리더 공지 문장 (REQ-BGS-016·017·018)

**결정.** 두 리더 공지에 **모두** 싣는다. 근거: 두 공지는 배타적이라 한쪽에만 두면 다른 쪽 리더가 문장을 받지 못하고(칸반 공지는 팩토리 환경에서 빈 문자열 — `session_start_kanban.go:50-52`, `TestKanbanNoticeSuppressedUnderFactoryEnv`), 두 리더 모두 운영자 대화를 쥔다. 레인·동반 세션은 카드 1장만 쥐므로 받지 않는다. 방출은 `startup`에서만 있고 resume·clear·compact·fork에서는 없다.

**사본과 로케일.** 에이전트용 사본(`additionalContext`)에는 영어 값이, 운영자용 사본(`systemMessage`)에는 설정된 `conversation_language`의 값이 실린다. 에이전트용 사본은 구성상 영어이므로(`session_start_kanban.go:17-19`) 로케일 격자는 리더 둘 × 로케일 넷, 곧 여덟 조합이고, 영어 렌더가 곧 에이전트용 사본이다. 알 수 없는 로케일은 영어로 폴백한다.

**문장의 모양(이 문장의 집은 여기다 — `plan.md` M5는 가리키기만 한다).** 한 문장. 무엇 — plan→run Kickoff를 운영자 형태로 묻는 중 대기 카드가 둘 이상이면 카드별 질문이 아니라 배치 게이트 요약 하나로 제시한다. 이름 — `batch gate summary`를 네 로케일에 그대로 둔다(포인터 주소와 같은 취급; REQ-BGS-016, AC-015). 어디 — `.claude/rules/moai/workflow/auto-semantics.md` §9.2. 서식·제외·반대 증거 의무는 적지 않는다: 정본을 가리킬 뿐이다(REQ-BGS-016).

**제약.** 기존 테스트가 고정한 블록 구조·역할 용어·형식 문자열 금지·카드 id/SPEC ID/날짜 금지를 깨지 않는다(`plan.md` §D 제약 11). 포인터 토큰은 네 로케일에서 번역하지 않는 주소로 다룬다.

**기각한 대안.** 칸반 리더 공지에만 싣는 안(팩토리 리더 누락), 레인·동반 공지에 싣는 안(카드 1장만 쥔다), 새 블록을 만드는 안(`TestKanbanLeadNoticeBlockLayout`의 5블록 구조를 깬다), 정본 서식을 문장에 옮겨 적는 안(두 벌이 되는 순간 갈라진다).

**비용.** 리더 세션의 `startup`에서만 한 줄이 늘고 비리더 세션은 0바이트다. 리더는 `kanban-dispatch.md`를 이미 매 턴 읽으므로 의도된 중복이다(spec.md R-3). 렌더 길이의 전/후는 run-phase에서 공시한다(플랜 단계에서는 측정하지 않았다).

**열린 행.** Q10 — 문장은 "batch gate summary"라는 이름을 쓴다. 초안이 이 이름을 택한 이유는 초안의 것이다: 이 이름은 승인 수단이 아니라 제시 형식을 가리킨다. 이 이유는 판정이 아니며 이름이 바뀌면 문장·이름 토큰·가드 앵커가 같이 바뀐다.

## D.8 감사 2회차 오케스트레이터 판정 R1–R4 (기록)

1회차 판정 R1–R5(Tier L, 기준선 경로, 양성형 차단 규칙, 예약 목록 출처, Q8 정정)는 `plan.md` §I와 decision-index Q11이 가진다. 아래는 2회차 판정이다. 판정은 이 초안의 범위와 서식을 정하며, 달리 판정할 수 있는 것은 decision-index에 행으로 남겼다.

- **R1 (N2) 범위 = plan→run Kickoff 행.** 카드 본문이 "per-card Kickoff and plan-audit approvals"를 이름으로 든다. REQ-BGS-011의 승인 가능성 술어는 계획 감사 증거로만 정의되는데, `auto-semantics.md` §7은 plan→run Kickoff와 sync 차단 승인을 권한 게이트로 명명해 둘 다에 양성 판정을 요구하고 §9는 sync 승인을 sync-auditor 판정과 증거 임계값으로 판정한다. 그래서 계획 감사 네 조건을 sync 승인 행에 그대로 적용하는 문단은 sync-auditor FAIL 행을 통과시킬 수 있었다. 확대는 행마다 승인 가능성 술어를 새로 정의해야 하며 이 SPEC은 정의하지 않는다 — Q13(선호 답 없음). 귀결: REQ-BGS-001·003·004·005·006·011·019, AC-006, 앵커 A01·A39·A40이 Kickoff 범위로 쓰였고, factory decide 행들도 다른 게이트 행이므로 개별 질문이다(§D.1).
- **R2 (N6) Q3는 풀지 않는다.** REQ-BGS-011과 AC-011을 "차단 행을 차단으로 보고하고 승인에서 뺀다"로 다시 써서 차단 행이 표 안에 있든 표 밖에 있든 성립하게 했다. 보고 위치만 Q3의 열린 부분이다.
- **R3 (N1) M0 종료 조건.** 기준선 파일이 존재하고, 무시되지 않으며, 네 요소 레이블 줄을 싣고, 기준선 커밋이 그 파일 하나만 담는다. "B가 D의 조상"과 `-S` 순서는 D를 만드는 M2 뒤에만 판정되므로 끝 점검(DoD)으로 옮겼다. 같은 부류(마일스톤 종료가 뒤 마일스톤의 산출물에 기대는 경우)를 plan의 모든 마일스톤에서 점검했고 마일스톤마다 `Exit:` 줄을 달았다. 네 요소 레이블 줄은 N4(placeholder 파일 변이)에서 왔고 파일만 있으면 판정되므로 M0 종료에 든다.
- **R4 (N5) 왼쪽 끝은 읽는 시점의 merge-base.** `gitflow-lane-protocol.md` §8(리터럴 base SHA로 "이 카드가 무엇을 바꿨는가"를 재지 않는다)을 따른다. 병합 전에만 유효하다는 한계와 범위가 비지 않았다는 대조를 AC-016에 적었다. 리터럴 `c50da9c2f`는 `research.md` R5의 계획 시점 측정으로만 남는다.
