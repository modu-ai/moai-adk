---
id: SPEC-ROLE-LOAD-PREDICATE-001
title: "Codex 역할 로드 판별식의 계약 중립화 (승계 노선)"
version: "0.5.0"
status: completed
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: High
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
tags: "codex, role-load, acceptance-criteria, live-measurement, predicate-design"
related_specs: "SPEC-DUAL-HARNESS-RECOVERY-001, SPEC-CODEX-AUDIT-READONLY-001"
tier: M
---

## HISTORY

### v0.5.0 (2026-09-26) — 열거된 결함 델타의 소진 (강화 아님)

iter-4 델타 감사 `FAIL 0.767`(0.65 → 0.68 → 0.75 → 0.752 → 0.767, critical 0, must-pass 7/7). 운영자가 「다섯 건만 수정하고 Kickoff」을 택했다. **이 판은 새 범위를 열지 않고 살아남은 기준을 강화하지 않는다** — 고친 다섯 건은 전부 문면이며, 판정식 형태가 바뀐 것은 둘(G3·G4)이고 그 실행 가능성은 iter-4 감사가 이미 측정했다. 전제(12/12)는 재측정하지 않았다.

- **G3** `AC-RLP-009`의 판정을 **네 단계**(1a git만 / 1b go만 / 2a git만 / 2b git 없는 판정)로 재분할했다. iter-3의 2단계도 이 워크트리 세션에서 거부됐고(감사자 3형태 실측), 방아쇠는 복잡도가 아니라 **한 호출 안의 `git` + 명령 치환 공존**이다. `plan.md` pre-flight #6·§H-3 F9 행(→**부분**)·`acceptance.md` §A의 [HARD] 항을 함께 고쳤다.
- **G4** 1단계 산출물 신선도(`fresh`·`ordered`)를 **판정 항으로** 넣었다(기계화 선택). **닫는 것은 다른 커밋 산출물의 재사용이고, 고의 위조는 닫지 못한다** — 두 변이를 실측해 `progress.md`에 적었고 잔여를 Residual-risk 5에 남겼다. 공짜가 아니다.
- **G1** §C 머리가 삭제된 `AC-RLP-010`을 현재 시제 판정자로 단언하던 문장을 **삭제 이전 사실**로 고쳤다. 같은 자리에 「기계가 지킨다고 주장하지 않는다」를 둔다.
- **G2** `progress.md`의 lint 행을 정정했다 — 선언은 **넷**이고 관측은 둘이며, 그 괴리의 원인은 이 SPEC 자신의 `[RETIRED]` 언급이 `REQ-RLP-006`·`008`의 커버리지 판독을 통과시키기 때문이다(iter-4 변이 대조). 경고 둘은 그대로 두고 `lint.skip`을 쓰지 않는다.
- **G5·G6** B.7을 B.8과 같은 모양으로 처분하고, §G 위험 표의 LIVE 범위 두 행을 「범위 밖」으로 고쳤다. iter-3의 「F6 정정 — 세 자리가 같은 이야기를 한다」는 그 시점의 과대 기재였고 **부분**으로 정정했다.
- **부채로 실어 보낸다:** G7(매달린 포인터 — 인접 처분 블록이 무장해제), G8(양성 대조 예시가 `superseded` SPEC — 오늘 건전, 선택 기준을 기록), G10(D9 행 대 집계 — 집계가 명시적으로 뒤집음). G9(iter 라벨 한 토큰)는 G3으로 같은 줄을 다시 쓰며 **지나가며 닫았다**.

### v0.4.0 (2026-09-26) — 축소 (subtractive)

plan-audit iter-3 `FAIL 0.75`(0.65 → 0.68 → 0.75, critical 0, must-pass 7/7 PASS)에 리드의 확장 종료 조건이 걸렸다. 4회차 강화는 없고, 감사자의 **범위 축소안**을 적용한다. 이 판은 **빼는 개정**이며 살아남는 기준을 강화하지 않는다.

- **절단 ①** `AC-RLP-005`(LIVE 14회 측정) → 후속 카드 후보. F3·F4가 함께 사라지고, 상한 증인 문제·감시 경로 동결의 되돌릴 수 없음·토큰 조기 언급 함정이 동시에 범위 밖으로 나간다. **잃는 것은 §F.1에 적는다 — 공짜가 아니다.**
- **절단 ②** `AC-RLP-010`·`AC-RLP-011` → **삭제.** `REQ-RLP-009`·`REQ-RLP-013`은 산문 의무로 남고 검증은 plan-audit·sync-audit의 **읽기**에 명시 위임한다. 사유는 §F.2 — 세 회차(D9 → E2 → F1) 내내 강화할 때마다 변이가 한 칸 옮겨갔을 뿐이다.
- **절단 ③** `AC-RLP-007`(선고정 조상 판정) → 후속 카드 후보. LIVE가 범위에 없으면 막을 대상이 없다. **알고리즘은 검증됐으므로 물려받되 다시 쓰지 않는다** — 여섯 케이스와 증거 경로를 §F.3이 적는다.
- **F5 정정** — 과대 기재 세 자리. (a) `REQ-RLP-015` 번호 위치 안내를 §B.1 머리에 **실제로 넣어** HISTORY의 주장을 참으로 만든다. (b) §H-2의 E2·E3 행을 「닫음」에서 정정한다. (c) iter-3 집계를 `RESOLVED 21 / PARTIAL 2 / UNRESOLVED 1`로 고친다.
- **F6 정정** — `plan.md` §G 위험 표의 스테일 잔여 행, §H D8 행의 폐기된 판정 형태 서술, §H 말미의 상반된 문장 셋을 `AC-RLP-007` 이관에 맞춰 함께 정리한다.
- **F7 처리** — `COUNT 15` 실측 단언을 걷어내고(수를 단언하지 않는다), 양성 대조 예시 경로를 `AC-` 부분 문자열이 없는 SPEC으로 바꿔 유령 AC를 제거한다.

### v0.3.0 (2026-09-26)
- plan-audit iter-2 `FAIL 0.68`(must-pass 7/7 PASS, Testability 0.50)에 대한 iter-3 개정. 리드가 Tier M 상한을 넘는 3회차를 인가했고, 조건은 「3회차 FAIL이면 확장을 끝내고 카드 범위를 줄인다」다.
- §A.4 분리(iter-2 E10): 네 인용은 **형제 무수정**의 근거이고, **영구 미충족**의 근거는 **iter-2 리드 결정**이다. iter-2의 "따라서"가 인용이 지지하지 않는 결론으로 이어졌던 것을 끊는다.
- §C.1·§C.2 고지 앵커에 설명을 같은 줄로 실어 내용 없는 변이가 판정을 통과하지 못하게 한다(iter-2 E2 — `AC-RLP-010`이 앵커 줄 길이·측정값·`확립한다` 면을 함께 본다).
- `AC-RLP-006`의 행동 필드 목록 출처를 LIVE 증거에서 **Go 타입 반사**로 옮겨 실제로 결정적이 되게 한다(iter-2 E5).
- `AC-RLP-007`을 위상 순서에서 **부모 그래프 조상 판정**으로 강화하고, iter-2가 남긴 잔여 사유(사실과 달랐다)를 제거한다(iter-2 E7).
- `AC-RLP-009` 재설계 — iter-2 판은 상호 배제된 두 항 때문에 구성상 통과할 수 없었다(iter-2 E1).
- `REQ-RLP-015`의 번호 위치에 대한 안내를 §B.1 머리에 둔다(iter-2 E19).

### v0.2.0 (2026-09-26)
- plan-audit iter-1 `FAIL 0.65`(`.moai/reports/t1171/plan-audit-iter1.md`)에 대한 iter-2 개정.
- **노선 확정 — 승계(Route A).** 형제 SPEC 두 개의 `acceptance.md`를 **고치지 않는다.** `AC-DHR-012`·`AC-DHR-023`은 영구 미충족으로 남고, 이 SPEC은 자기 AC로 판별식을 세운다. §A.4가 근거를 인용한다. iter-1의 clarification 표식(주제: 개정 대 승계)은 해소되어 네 아티팩트에서 제거됐다(MP-7).
- `priority: HIGH` → `High`(MP-3/D19). `title`·`version` 인용(D20).
- §A.2 파생 회복 주장 정정(D3): `AC-DHR-023`은 자기 nonce 항과 `ac023-evidence.json` 의존 때문에 회복하지 않는다. 승계 노선에서는 셋 다 이 SPEC의 완료 조건이 아니다.
- §A.1에 t1100 실행을 셋째 재현 사례로 추가(D23).
- §C.1/§C.2 고지를 **고유 앵커 문장**으로 재작성(D9). §C.5 신설 — 경로 A/B 분리의 기계 수단(D15).

### v0.1.0 (2026-09-26)
- 최초 저작. 카드 t1171(t1143 후속).

---

## §A 문제

### A.1 관측된 결함

`AC-DHR-012`의 (i) 절은 "역할 정의가 하위 에이전트에 실제로 로드되었는가"를 **역할별 nonce를 그대로 되돌려 달라는 요청에 모델이 그 nonce를 답했는가**로 판별한다. 이 판별식은 12개 역할 중 두 역할과 **계약상** 양립하지 않는다.

`manager-lead`와 `mission-governor`는 각자의 역할 정의에 따라 **작업을 지목하지 않은 위임을 거절하고 결정 객체만 답할 의무**가 있다(`manager-lead.toml` — `the delegation named no work — return the single line: LEAD BLOCKED`; `mission-governor.toml` — `Return exactly one decision object`). nonce만 되돌려 달라는 요청은 정의상 작업을 지목하지 않으므로, 역할 정의를 **읽고 따른** 에이전트는 nonce 대신 거절문을 답하는 것이 옳다. 즉 이 판별식은 그 두 역할에 대해 역할 로드가 아니라 **계약 준수 여부**를 재고 있고, 계약을 지킨 응답을 로드 실패로 기록한다.

**같은 실패가 세 번 재현됐다.**

| 실행 | 결과 | 근거 |
|---|---|---|
| t1100 LIVE (codex-cli 0.156.1) | 같은 두 역할만 `nonce_returned`가 빈 값, 나머지 10개는 왕복 성공 | `.moai/reports/t1100/ac012-evidence.json` |
| t1143 M4 실행 1 | 리드가 `INVALID — probe wording conflicts with the role contract`로 기록 | `.moai/reports/t1143/verdict.md` 원장 #13·#15 |
| t1143 M4 재실행 (문구를 12개 역할에 같게 적용) | `FAIL`. 10/12 nonce 회수, `manager-lead`는 `LEAD BLOCKED: The delegation named no work.`, `mission-governor`는 `blocker` 결정 JSON | 같은 판정서 원장 #27·#29, 결과줄 |

음성 대조(역할 미로드 → `false`)는 확인되었으므로 판별식의 판별력 자체는 살아 있다. 문제는 **판별 대상이 역할 로드가 아니라는 것**이다.

**서술 오차 한 건을 함께 기록한다.** `SPEC-DUAL-HARNESS-RECOVERY-001/acceptance.md:211`은 t1100 실행에 대해 "12개 역할 로드와 호출 수 14는 관측되었으나"라고 적었으나, 그 실행의 nonce 항 기준으로는 **10/12**였다. 그 주석이 파생 실패 범위 판단의 입력이므로 무해하지 않다 — 이 SPEC은 그 주석을 고치지 않되(§A.4), 그것을 근거로 쓰지 않는다.

### A.2 파생 실패의 범위 — 절반만 참이다

`AC-DHR-023`과 `AC-CAR-012b`는 `AC-DHR-012`와 같은 실행의 같은 `ac012-live.jsonl`을 읽고, 공통항 `([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0`을 포함한다. 그러나 **둘의 처지는 같지 않다.**

| AC | 자기 nonce 항 | 읽는 증거 파일 | 공유 테스트가 통과하면 회복하는가 |
|---|---|---|---|
| `AC-CAR-012b` | 없음 (route / used_spawn_agent / session_sandbox / probe_* / verdict_writer만) | `ac012-evidence.json`, `ac023-evidence.json` | 원리상 그렇다 |
| `AC-DHR-023` | **있음** — `.returned_contains_nonce==true` | `ac023-evidence.json` | **아니다** |

`AC-DHR-023`은 자기 판정식 안에 nonce 항을 갖고, 이 SPEC의 새 설계는 `ac023-evidence.json`을 생산하지 않는다. 따라서 "고치지 않아도 함께 회복된다"는 v0.1.0의 진술은 `AC-DHR-023`에 대해 **틀렸다**(plan-audit D3).

**더 앞선 사실 하나가 셋 모두를 덮는다.** 형제 세 판정 명령은 전부 `.moai/reports/t1100/`을 리터럴로 읽고, 그 디렉터리는 이 워크트리에 **없다**. 따라서 어떤 노선을 택해도 "형제 세 판정 명령을 원문 그대로 실행"은 이 트리에서 참을 낼 수 없다. 이것은 결함이 아니라 경계이며 §C.6이 한계로 적는다.

### A.3 이 SPEC이 정하는 것

역할 로드를 **계약과 무관하게** 판별하는 판별식을 정하고, 그 판별식이 판별력을 잃지 않음을 양·음 대조로 고정한다. 판별식은 LIVE 실행 **전에** 확정한다. 판별식은 **이 SPEC의 AC로** 세우고, 형제 SPEC의 문서는 건드리지 않는다(§A.4).

### A.4 노선 — 승계. 형제 문서는 고치지 않는다

[HARD] 이 SPEC은 `SPEC-DUAL-HARNESS-RECOVERY-001`과 `SPEC-CODEX-AUDIT-READONLY-001`의 어떤 파일도 **수정하지 않는다.** 근거는 두 SPEC에 네 자리로 기록된 리드 결정이다.

| 인용 위치 | 기록된 결정 |
|---|---|
| `.moai/specs/SPEC-CODEX-AUDIT-READONLY-001/acceptance.md:302` | "**카드 t1171로 이월 — FAIL. 판정식·기대값은 바꾸지 않음 (run 종료 뒤 리드 결정)**" + "아래 §B의 원문 블록은 바꾸지 않음" — **이 카드를 만든 바로 그 문장이다.** |
| `.moai/specs/SPEC-DUAL-HARNESS-RECOVERY-001/acceptance.md:211` | "카드 t1143으로 이관 — 이 SPEC 범위에서 충족되지 않음. **판정식과 기대값은 바꾸지 않음**" (`AC-DHR-012`) |
| `.moai/specs/SPEC-DUAL-HARNESS-RECOVERY-001/acceptance.md:414` | 같은 표식 (`AC-DHR-023`) |
| `.moai/specs/SPEC-DUAL-HARNESS-RECOVERY-001/acceptance.md:430` | "이 SPEC 안에서 이 기준은 PASS가 아니며, **판정식과 PASS 조건은 바꾸지 않는다**" (§C `AC-AGENT-01` 집계 행) |

**두 결론의 근거를 섞지 않는다.** iter-2 plan-audit E10이 지적했듯, 위 네 인용이 기록하는 것은 「**이 SPEC 범위에서** 충족되지 않음 / 판정식과 기대값은 바꾸지 않음 / 위 본문과 판정식은 t1143이 그대로 이어받는다」 — 즉 **이관과 무수정**이며, 영구성은 그 문장들이 말하는 바가 아니다. 그래서 근거를 갈라 적는다.

| 결론 | 근거 |
|---|---|
| **형제 SPEC의 파일을 고치지 않는다** | 위 네 인용. 네 자리 모두 "판정식·기대값을 바꾸지 않는다"를 기록한다 |
| **`AC-DHR-012`·`AC-DHR-023`은 영구 미충족으로 남는다** | **iter-2 리드 결정** — 노선을 승계로 확정하고 형제 두 SPEC의 `acceptance.md`를 건드리지 않기로 한 결정(iter-2 plan-audit 이후 리드 지시, 2026-09-26). 인용 넷이 아니라 이 결정이 영구성의 권한이다 |

이 SPEC은 `AC-DHR-012`·`AC-DHR-023`의 충족을 주장하지 않고, 그 둘의 판정식·기대값·본문을 고치지 않으며, `AC-CAR-012b`의 충족도 주장하지 않는다. 둘을 미충족으로 남기는 대가는 `AC-AGENT-01`(출처 SPEC §C)과 "이어받은 항목" 묶음(`SPEC-CODEX-AUDIT-READONLY-001` §D)이 그 SPEC들 안에서 PASS가 되지 않는다는 것이며, 그것은 이미 두 SPEC이 스스로 적어 둔 상태다.

승계 노선이 제자리 개정보다 싼 이유는 셋이다 — (1) `completed` 상태 SPEC 둘의 `completed → in-progress (amendment)` 전이(`amendment_of:` + HISTORY `## Amendments` + `prior_completed_sha`)를 피한다. (2) `AC-CAR-009`가 동결한 미러 블록과 HEAD 출처가 갈라지는 상태를 만들지 않는다. (3) §A.2가 보였듯 개정해도 `AC-DHR-023`은 회복하지 않으므로, 개정의 편익 자체가 기대보다 작다.

---

## §B 요구사항 (GEARS)

요구사항 문장은 GEARS 표기(영문)로 적고, 그 아래 한국어 주석이 근거와 범위를 설명한다. 식별자·경로·명령은 영문 원문이다.

**번호 배치.** 각 요구사항의 처분(이 카드가 판정하는지, 산문 의무인지, 후속 카드 후보인지)은 `acceptance.md` §A.1의 두 표가 적는다. 번호는 **재사용하지 않고 비우지도 않는다** — 이관·위임된 항목도 제자리에 남아 다음 독자가 결번을 의심하지 않게 한다.

### B.1 판별식의 성질

**`REQ-RLP-015`는 005와 006 사이가 아니라 이 묶음 끝에 있다** — 번호순이 아니라 주제순 배치이며, 선별(selection)이 판별식의 성질에 속하기 때문이다(iter-2 E19). 번호로 찾는 독자는 문서 중간이 아니라 이 절 끝에서 015를 만난다.

### REQ-RLP-001 — 로드 판별의 입력은 세션 기록뿐이다

The role-load predicate shall decide role load from the content of that role's own session record, and shall not accept the model's free-form reply text as an input to that decision.

모델 응답을 입력에서 빼는 것이 이 SPEC의 핵심이다. 응답을 입력으로 두는 한, 계약이 응답 형태를 규정하는 역할은 계약 준수와 로드가 구별되지 않는다.

### REQ-RLP-002 — 계약이 응답을 규정하는 역할도 같은 방식으로 판별한다

**Where** a role's contract obliges it to refuse a delegation that names no work, the role-load predicate shall decide that role's load by the same function it applies to every other role.

역할 이름으로 분기하는 예외 문구를 두지 않는다. 예외를 두면 그 예외 자체가 다음 번 판별력 상실의 자리가 된다.

### REQ-RLP-003 — 지문은 TOML 명세 파싱값에서 계산한다

The predicate shall compute the role-body fingerprint from the `developer_instructions` value obtained by a TOML-specification parse, and shall not use a regular-expression extraction of that value.

TOML은 여러 줄 문자열의 여는 구분자 직후 줄바꿈 하나를 제거하므로, 정규식 추출값은 파싱값보다 선행 바이트 하나가 많다. §C.4에 실측이 있다.

### REQ-RLP-004 — 버전이 어긋난 기대표는 거짓을 낸다

**When** the expectation table is built from a tree version other than the one the artifact under judgment carries, the predicate shall return false.

지문은 버전 귀속적이다. 이 성질은 결함이 아니라 판별력의 근거이며, §C.1이 그 한계를 함께 고지한다.

### REQ-RLP-005 — 적격하지 않은 기대표는 거짓을 낸다

**When** the set of role bodies does not carry as many distinct hashes as there are roles, the predicate shall return false.

본문이 겹치는 두 역할이 있으면 `matched_roles`가 하나로 좁혀지지 않는다. 표 적격성을 먼저 거른다.

### REQ-RLP-015 — 픽스처 선별은 표찰로 하고 본문 일치로 하지 않는다

The fixture-selection predicate shall select a session record by the presence of its `agent_role` label, and shall not use role-body agreement with the expectation table as a selection condition.

본문 일치를 선별 조건으로 쓰면 양성 갈래가 **구성상 항진**이 되고, 그래도 판정식은 통과한다(plan-audit D13). 원본 디렉터리는 기록 26개를 담고 그중 12개만 표찰을 가진다 — 선별은 그 표찰로만 한다.

### B.2 판별력과 고지

### REQ-RLP-006 — 판별식은 LIVE 실행 전에 확정된다 — **후속 카드 후보 (이 카드에 판정식 없음)**

The predicate shall be fixed in `acceptance.md` before any LIVE run begins, and shall not be loosened after a run result has been observed.

> **처분.** 이 요구사항의 유일한 판정식은 `AC-RLP-007`이었고 그것이 §F.3으로 이관됐다. LIVE가 이 카드의 범위 밖이므로 **막을 대상 자체가 없다** — 결과를 본 뒤 기준을 낮추는 일이 일어날 수 없다. 요구사항 문언은 후속 카드가 그대로 물려받도록 제자리에 남긴다. [HARD] 이 카드는 이 요구사항을 기계가 지킨다고 주장하지 않는다.

결과를 본 뒤 기준을 낮추는 개정은 금지한다. 이행은 커밋 순서로 판정하며, 그 판정이 무엇을 닫고 무엇을 닫지 않는지는 `acceptance.md` AC-RLP-007이 명시한다.

### REQ-RLP-007 — 모든 판별식은 양·음 대조를 갖는다

The verification suite shall pair every predicate with both a positive control and a negative control, and no predicate shall return true on an empty or absent input.

빈 결과집합 통과를 배제한다.

### REQ-RLP-008 — LIVE 상한은 인수 기준 안에 미리 적히고, 집행되는 상한과 선언은 구별된다 — **후속 카드 후보 (이 카드에 판정식 없음)**

**While** a LIVE measurement is open, every cap shall already be stated inside the acceptance criterion, and the criterion shall mark each cap as either mechanically enforced or a declaration that no judge verifies.

상한을 착수 뒤에 정하면 상한이 아니다. 그리고 집행되지 않는 상한을 집행되는 것처럼 적으면 그것이 공허 초록이 된다.

> **처분.** 유일한 판정식이 `AC-RLP-005`였고 그것이 §F.1로 이관됐다. 이 카드는 LIVE를 돌리지 않으므로 상한이 구속할 실행이 없다. 후속 카드는 이 문언과 함께 iter-3 F4·F10의 두 지적을 물려받는다 — 호출 수·호출당 상한에 **독립 증언이 없고**(생산자가 하나), 기준이 셋으로 갈리는데 요구사항 문언은 둘을 말한다. [HARD] 이 카드는 이 요구사항을 기계가 지킨다고 주장하지 않는다.

### REQ-RLP-009 — 판별식의 양면을 고유 문장으로 고지한다 — **산문 의무, 검증은 감사 읽기에 위임**

The SPEC shall state both what the predicate establishes and what it does not establish, as individually distinct sentences rather than as a repeated marker.

> **처분.** 요구사항은 살아 있고 §C.1·§C.2가 그것을 이행한다. 달라진 것은 **누가 확인하는가**다 — 문자열 판정식(`AC-RLP-010`)은 삭제되고, 충족 여부는 plan-audit·sync-audit이 §C를 **읽어** 판정한다. 사유는 §F.2. [HARD] 이 카드는 이 요구사항을 기계가 지킨다고 주장하지 않는다 — 그 주장을 뺀 것이 공허 초록을 없애는 방법이다.

한쪽 면만 적는 고지는 이 저장소에서 결함으로 본다. 그리고 같은 표식을 반복해 개수만 맞춘 고지는 고지가 아니다 — plan-audit D9가 내용 없는 표 행 변이로 v0.1.0의 판정식을 통과시켰다.

### B.3 행동 측정의 분리

### REQ-RLP-010 — 행동 탐침은 별개의 인수 기준으로만 존재한다

The behaviour probe, which measures whether the model read and followed the role instructions, shall exist only as an acceptance criterion separate from the load predicate, and shall not be the sole basis for a load claim.

경로 B를 폐기하지 않고 로드 판별의 자리에서 물러나게 한다.

### REQ-RLP-011 — 로드 판별 함수의 입력 타입은 행동 필드를 갖지 않는다

The load-predicate function shall accept an input type that carries no behaviour-derived field, so that reading a behaviour value inside that function is a compile error rather than a convention.

"코드 경로가 읽지 않는다"는 인쇄될 수 있을 뿐 보여지지 않는다(plan-audit D15). 구조적으로 막는 형태는 타입이며, 그 증언자는 컴파일러다.

### B.4 범위 보존

### REQ-RLP-012 — 형제 SPEC의 파일을 수정하지 않는다

This SPEC shall not modify any file under `.moai/specs/SPEC-DUAL-HARNESS-RECOVERY-001/` or `.moai/specs/SPEC-CODEX-AUDIT-READONLY-001/`.

승계 노선의 핵심 제약이며, v0.1.0의 "(i) 절 하나로 한정한다"를 대체한다. 전자는 여섯 자리에 걸친 편집면을 지목하지 못해 검증 수단이 없었고(plan-audit D5), 후자는 `git diff`로 곧바로 측정된다.

### REQ-RLP-013 — 영구 미충족을 문서에 명시한다 — **산문 의무, 검증은 감사 읽기에 위임**

The SPEC shall state that `AC-DHR-012` and `AC-DHR-023` remain permanently unmet, and shall cite the recorded lead decision by file and line.

> **처분.** 요구사항은 살아 있고 §A.4가 그것을 이행한다 — 네 인용은 제자리에 있고(세 회차 모두 감사자가 원문을 읽어 참임을 확인했다), 영구 미충족 선언과 그 권한 근거도 그대로다. 문자열 판정식(`AC-RLP-011`)만 삭제되고 충족 여부는 감사 읽기가 판정한다. 사유는 §F.2. [HARD] 이 카드는 이 요구사항을 기계가 지킨다고 주장하지 않는다.

기록된 결정과 충돌하는 노선을 침묵으로 취하는 것이 plan-audit D2였다. 인용은 §A.4에 있다.

### REQ-RLP-014 — 이 SPEC의 acceptance.md가 코퍼스에 들어가면 스냅숏을 같은 커밋에 재생성한다

**When** a commit adds or changes an `acceptance.md` that the AC corpus snapshot records, that same commit shall carry the regenerated snapshot.

빠뜨리면 develop CI가 떨어진다. 승계 노선에서는 형제 파일이 아니라 **이 SPEC 자신의** `acceptance.md`가 그 대상이다.

---

## §C 판별식이 확립하는 것과 확립하지 않는 것

이 절은 **REQ-RLP-009**의 이행이다. 각 항목은 **서로 다른 고유 문장**이다. 그 고유성은 iter-3판까지 `acceptance.md`의 문자열 판정식(구 `AC-RLP-010`)이 판정하려 한 대상이었으나, **그 판정식은 iter-4에서 삭제됐다** — 변이 통과가 실측됐기 때문이며(§F.2), 지금 고유성을 확인하는 것은 plan-audit·sync-audit의 읽기다. [HARD] 이 카드는 이 성질을 기계가 지킨다고 주장하지 않는다(§B `REQ-RLP-009` 처분 블록과 같은 문장).

### C.1 경로 A — 전송 지문 (Transport fingerprint)

역할의 세션 기록(rollout JSONL) 안에 `type == "response_item"`, `payload.role == "developer"`인 항목이 있고, 그 본문 텍스트의 sha256이 **그 역할의** `developer_instructions` 해시와 같은지를 본다.

**확립한다.** 역할 정의 본문의 **바이트**가 그 하위 에이전트의 컨텍스트에 들어갔다. 역할별로 서로 다른 본문이므로, 어느 역할의 본문이 들어갔는지까지 가른다.

확립하지 않는다 (1) — 읽었음: 모델이 그 지시를 읽었다는 것. 컨텍스트에 있는 것과 주의가 그리로 갔는 것은 다르다.

확립하지 않는다 (2) — 따랐음: 모델이 그 지시를 따랐다는 것. 이것은 §C.2의 별개 측정이다.

확립하지 않는다 (3) — 다른 TOML 필드: 역할 TOML의 `sandbox_mode`·`model_reasoning_effort`·`tools`가 적용되었다는 것. 본문 지문은 본문에 대해서만 말한다.

확립하지 않는다 (4) — 현행 트리와의 일치: 지문이 현행 워킹트리의 본문과 같다는 것. 판정 대상 산출물이 담은 버전과 같은지를 말할 뿐이다(REQ-RLP-004).

**측정된 근거(이 트리, 이 실행).** t1143 M4 재실행이 남긴 12개 하위 에이전트 세션 기록에서, 12개 역할 전부가 자기 역할 본문과 정확히 일치하는 `developer` 항목을 갖고 있었다 — `manager-lead`(27879 bytes, `714d3ca3…`)와 `mission-governor`(1724 bytes, `e3c7bded…`) 포함. nonce 판별식이 거짓을 낸 바로 그 실행에서 이 판별식은 12/12로 참이다. plan-audit iter-1이 별도 구현으로 같은 값을 독립 재측정했다.

### C.2 경로 B — 계약 호환 행동 탐침 (Behaviour probe)

역할 계약이 **허용하는** 형태의 작업을 주고, 그 역할 정의를 읽지 않으면 만들 수 없는 답을 요구한다.

**확립한다.** 모델이 역할 지시를 읽고 **썼다**(행동). 경로 A가 확립하지 않는 바로 그 면.

확립하지 않는다 (B1) — 원문 그대로의 전송: 본문이 바이트 그대로 전달되었다는 것. 요약·절단된 지시로도 맞는 답이 나올 수 있다.

확립하지 않는다 (B2) — 실패의 귀속: 실패가 로드 실패라는 것. 프롬프트 문구 결함과 로드 실패가 같은 출력으로 나타난다 — t1143에서 두 번 실현된 실패 계열이며, 이 SPEC이 존재하는 이유다.

### C.3 선택 기준 (둘 다 남기고, 역할을 가른다)

두 경로는 서로 다른 것을 재므로 어느 쪽도 다른 쪽을 포섭하지 않는다. 그래서 **둘 다 채택하되 질문으로 가른다.**

> **"정의가 에이전트에 닿았는가"는 경로 A가 판정한다. "에이전트가 정의대로 행동했는가"는 경로 B가 판정한다. 경로 B는 로드 주장의 유일 근거가 될 수 없다** (REQ-RLP-010).

기존 nonce 판별식은 경로 B의 한 사례이며, 그 프롬프트가 12개 역할 중 둘의 계약과 충돌했다. 이 SPEC은 경로 B를 폐기하지 않고 **로드 판별의 자리에서 물러나게** 한다.

### C.4 파서 선택은 임의가 아니다 (측정됨)

TOML 명세는 여러 줄 문자열에서 여는 구분자 직후의 줄바꿈 하나를 제거한다. 그래서 같은 파일의 `developer_instructions`를 정규식으로 뽑으면 TOML 파싱값보다 **선행 `\n` 하나가 더 붙는다.** 이 트리의 `plan-auditor.toml`에서 정규식은 61906 bytes / `517d444e…`, TOML 파싱은 61905 bytes / `d1bea71a…`였고, 세션 기록이 담은 본문은 **TOML 파싱값**과 일치했다. 한 바이트 차이가 판별식을 항상 거짓으로 만들므로, REQ-RLP-003은 편의 조항이 아니라 정확성 조항이다. plan-audit iter-1이 delta=1이고 그 바이트가 `b'\n'`임을 독립 확인했다.

### C.5 경로 A/B 분리는 무엇으로 집행되는가

분리를 **산문으로** 선언하는 것으로는 부족하다 — `behaviour`를 하위 객체로 내리는 것은 구조적 경계를 시사하되 읽기를 막지 않고, "코드 경로가 읽지 않는다"는 테스트가 출력할 수는 있어도 보여지지 않는다(plan-audit D15). 이 SPEC의 집행 수단은 셋이며 각각 다른 것을 담보한다.

| 수단 | 담보하는 것 | 증언자 | 담보하지 않는 것 |
|---|---|---|---|
| 로드 판별 함수의 입력 타입에 행동 필드가 없다(REQ-RLP-011) | 함수 본문에서 행동값을 읽는 것이 **컴파일 오류**다 | 컴파일러 | 호출자가 행동값을 보고 결과를 후처리하는 것 |
| 타입 반사 검사 — 입력 타입의 필드 이름에 `behaviour`·`nonce`가 없다 | 필드 추가로 우회하면 검사가 깨진다 | `AC-RLP-006` 판정식 | 다른 이름의 행동 필드를 넣는 것 |
| 행동 필드 **전수** 변이 — `behaviour` 하위 모든 필드를 각각 뒤집고 로드 결과 불변을 본다 | 저자가 고른 한 변이가 아니라 필드 집합 전체를 덮는다 | `AC-RLP-006` 변이 표 | `behaviour` 밖의 경로로 결합하는 것 |

세 수단을 합쳐도 "호출자 층의 후처리 결합"은 남는다. 그것은 이 SPEC이 닫지 못하는 잔여 위험이며, 산문으로 닫은 척하지 않는다.

### C.6 형제 판정 명령은 이 트리에서 실행 불가다 (경계)

`AC-DHR-012`·`AC-DHR-023`·`AC-CAR-012b`의 판정 명령은 전부 `.moai/reports/t1100/`을 리터럴 경로로 읽는다. 그 디렉터리는 이 워크트리에 **없다**. 리드 결정에 따라 그 증거를 이 트리로 복사하지 않으므로, 세 명령의 상태는 이 SPEC의 범위에서 **이 트리에서 실행 불가 (`.moai/reports/t1100/` 부재)** 로 기록되며, 실행 결과를 완료 조건으로 삼지 않는다. 경로를 인자화하지 않는 한 "원문 그대로 실행"은 성립하지 않는다는 사실 자체가 이 한계의 내용이다.

---

## §D 무엇을 만들지 않는가 (exclusions)

### Out of Scope — 형제 SPEC의 문서 개정
- `SPEC-DUAL-HARNESS-RECOVERY-001`·`SPEC-CODEX-AUDIT-READONLY-001`의 어떤 파일도 고치지 않는다(§A.4, REQ-RLP-012). `AC-DHR-012`·`AC-DHR-023`·`AC-CAR-012b`의 판정식·기대값·본문·미러 블록 전부 대상 밖이다.
- 두 SPEC의 `completed → in-progress (amendment)` 전이, `amendment_of:` 필드, HISTORY `## Amendments` 절, `prior_completed_sha` 기입. 승계 노선에서 이 절차는 발생하지 않는다.

### Out of Scope — 형제 증거의 이 트리 반입
- `.moai/reports/t1100/` 증거를 이 워크트리로 복사하는 일. 세 형제 판정 명령을 이 트리에서 실행 가능하게 만들지 않는다(§C.6).
- 형제 판정 명령의 경로 인자화.

### Out of Scope — LIVE 측정과 그것에 매달린 장치들
- 12개 역할에 대한 LIVE 14회 측정 자체(구 `AC-RLP-005`). 호출 상한·원장·`lane_turns`·양성 대조 탐침 파일도 함께 범위 밖이다.
- 선고정의 커밋 조상 판정(구 `AC-RLP-007`)과 그것에 붙은 감시 경로 동결, LIVE 표지 토큰과 그 조기 언급 금지 조항. LIVE가 없으면 막을 대상이 없다.
- 후속 카드 후보로 기록된다(§F.1·§F.3). **이 카드는 그 카드를 발행하지 않는다** — 큐 변경은 운영자 승인 뒤 리드가 한다.

### Out of Scope — 문서 서술을 문자열로 판정하는 일
- §C의 양면 고지가 내용을 담았는지, §A.4의 인용이 문서 서술과 일치하는지를 **판정식으로** 재는 일(구 `AC-RLP-010`·`AC-RLP-011`). 세 회차 측정 결과 문자열 검사가 결정할 수 있는 성질이 아니다(§F.2).
- `REQ-RLP-009`·`REQ-RLP-013`의 요구 자체는 범위 안에 남아 있다 — 빠진 것은 기계 판정뿐이고, 검증은 감사 읽기가 맡는다.

### Out of Scope — 모델 순응의 증명
- 모델이 역할 지시를 읽었는지, 이해했는지, 따랐는지를 기계로 증명하는 일. §C.2는 행동의 **표본 하나**를 재며 순응 일반을 증명하지 않는다.
- 역할 지시의 품질·표현·길이에 대한 판단.

### Out of Scope — sandbox 적용 경로의 재설계
- `sandbox_mode`가 하위 에이전트에 적용되지 않는 문제와 그 최상위 read-only 실행 경로. 이미 `SPEC-CODEX-AUDIT-READONLY-001`이 소유한다.
- `AC-CAR-012a`의 경로 유도 규칙과 launch record 스키마.

### Out of Scope — 과거 실행의 재판정
- t1143 M4의 두 실행(원장 #7-#34)과 t1100 실행의 결과를 새 판별식으로 다시 판정해 `FAIL`을 뒤집는 일. 결과를 본 뒤 기준을 바꿔 통과시키는 것이므로 REQ-RLP-006 위반이다. 그 실행의 세션 기록은 **판별식 설계의 픽스처**로만 쓴다.

---

## §E 측정 귀속

이 SPEC의 §A·§C가 인용한 수치의 출처.

| 주장 | 출처 |
|---|---|
| 두 역할이 세 실행에서 nonce 판별식에 걸림(t1100, t1143 ×2), 재실행 10/12, 음성 대조 확인 | `.moai/reports/t1100/ac012-evidence.json`; `.moai/reports/t1143/verdict.md` 원장 #13·#15·#27·#29, 결과줄 |
| 12개 역할 본문 지문이 12/12 자기 역할과 일치(`manager-lead`·`mission-governor` 포함) | `.moai/reports/t1143/m4-live-evidence/ac012-sessions/` 중 `agent_role` 표찰을 가진 12개 rollout × 이 워크트리(base `0356e8117`)의 `internal/template/templates/.codex/agents/moai/*.toml`. plan-audit iter-1이 독립 재측정 |
| 원본 디렉터리 기록 26개, 표찰 보유 12개, 나머지 14개는 부모 세션(표찰 없음, 어느 역할 본문과도 불일치) | 같은 디렉터리 전수. plan-audit iter-1 D13 |
| 다른 버전 트리 기대표 → 0/12, 역할 파일 부재 → 0/12 | 같은 rollout 집합에 primary 체크아웃 템플릿(역할 11개) 및 빈 디렉터리를 기대표로 적용. plan-audit iter-1이 두 트리에 걸쳐 바이트 동일 본문 0개임을 추가 확인 |
| 정규식 61906 `517d444e…` / TOML 파싱 61905 `d1bea71a…`, 차이는 선행 `b'\n'` 1바이트, 세션 기록은 파싱값과 일치 | 이 트리 `plan-auditor.toml`, `tomllib` 대 정규식 대조. plan-audit iter-1 독립 확인 |
| 역할 12개, 서로 다른 본문 12개 | 같은 트리의 `*.toml` 전수 |
| 두 역할의 거절 의무가 계약에 실재 | `manager-lead.toml`·`mission-governor.toml` 본문 |
| `.moai/reports/t1100/` 부재 | 이 워크트리 |
| LIVE 증거 경로가 gitignore 대상 | `.gitignore:235` `.moai/reports/*` — `git check-ignore -v`로 확인 |

`.moai/reports/t1143/**`와 `.moai/reports/t1100/**`는 primary 체크아웃의 미추적 경로다. run 단계는 판별식 픽스처를 이 워크트리 안으로 고정하고(`plan.md` §F M1), 그 사본의 sha256을 `progress.md`에 적는다.

---

## §F 후속 카드 **후보** (발행하지 않는다)

[HARD] 아래 셋은 **후보 기록**이다. 카드 발행은 큐 변경이며 운영자 승인 뒤 리드가 수행한다 — 이 SPEC은 후보를 적을 뿐 `moai gtd add`를 부르지 않는다.

### §F.1 후보 — LIVE 14회 측정 (구 `AC-RLP-005`, `REQ-RLP-006`·`REQ-RLP-008`을 함께 물려받는다)

**무엇을.** 격리된 임시 저장소·임시 `CODEX_HOME`에서 12개 역할에 하위 에이전트를 한 번씩 띄우고(작업 문구는 12개에 같게), 감사 역할 둘의 쓰기 시도를 더해 정확히 14회를 쓴다. 판별식은 이 카드가 착지시킨 것을 **고정된 것으로 받아** 쓴다.

**절단의 대가 — 공짜가 아니다.** 잃는 것은 **「새 판별식이 새 실행에서도 참인가」의 재확인** 하나다. 이 카드가 남기는 근거는 t1143 M4 재실행의 **실물 기록** 위에서의 12/12이며, 그것은 발견으로는 충분하지만 **재현성의 증거는 아니다.** 구체적으로 이 카드는 다음을 말하지 못한다 — (a) 다른 codex CLI 버전에서도 `developer` 항목이 같은 모양으로 남는가, (b) 새 실행에서도 12개 본문이 서로 다른 채로 전달되는가, (c) 감사 역할의 쓰기 거부가 이 판별식과 함께 여전히 관측되는가. 셋 다 후속 카드의 몫이고, 그때까지 이 SPEC의 주장은 **한 실행에 대한 것**이다.

**후속 카드가 선고정 문제를 겪지 않는 이유.** 판별식이 이미 착지해 있으므로 「결과를 본 뒤 기준을 낮춘다」의 대상이 존재하지 않는다 — 그래서 `REQ-RLP-006`의 게이트가 그 카드에서는 값이 있다.

**물려받는 미검증 항목.** iter-3 F4(호출당 상한의 `exit_code==124` 겹이 판정식에 없다)·F10(`REQ-RLP-008`은 둘로 말하는데 기준은 셋)·D12(호출 수·호출당 상한에 독립 증언 없음). 외부 증인을 가진 상한은 실행 전체 벽시계 하나다.

### §F.2 후보 — `REQ-RLP-009`·`REQ-RLP-013`의 검증 방식

**무엇을.** 두 요구사항은 이 카드에서 **산문 의무**로 남고 검증이 plan-audit·sync-audit의 읽기에 위임된다(§B의 두 처분 블록). 후보 작업은 그 위임을 감사 절차 쪽에 성문화할지 — 예컨대 감사 체크리스트 항목으로 두어 읽기가 빠지지 않게 할지 — 를 정하는 일이다.

**왜 판정식을 버렸는가 — 이것이 이 카드의 발견이다.** 세 회차 내내 강화할 때마다 변이가 한 칸 옮겨갔다.

| 회차 | 강화한 것 | 그 회차 감사자가 만든 변이 | 결과 |
|---|---|---|---|
| iter-1 D9 | `확립하지 않는다` 적중 수 | 같은 표식 반복 표 행 | 통과 |
| iter-2 E2 | 고유 앵커 여섯 + `sort -u` | 서로 다른 앵커 여섯, 내용 0 | 통과 |
| iter-3 F1 | 앵커 줄 길이 + `확립한다` 면 + 측정값 여섯 토큰 | 앵커 줄 **바이트 그대로** 보존, 측정값을 설명 없는 한 줄로 유지, `확립한다` 면은 남기고 본문을 공허하게, §C.3·§C.5·§C.6 삭제 | 통과 |
| iter-1~3 (인용) | 인용 존재 → 원문 대조 → 부정 문구 열거 | 인용 정확·원문 대조 통과·열거 문구 미사용, 그러면서 「그 결정문은 이 카드에 걸리지 않는다」·「형제 `acceptance.md`를 손질한다」를 단언 | 통과 |

**결론은 「네 번째 강화가 필요하다」가 아니다.** 문자열 검사는 **문장이 자리에 있는지**를 결정하고, 요구되는 성질은 **서술이 사실에 맞는지**다. 후자는 읽어야 알 수 있다. 덧붙여 iter-3이 지목한 구조적 한계: 내가 「실제 방어」로 지명했던 항(인용 위치의 줄을 직접 읽는 대조)은 **형제 파일**을 읽는데, `REQ-RLP-012`가 이 카드 범위에서 그 파일들을 불변으로 묶으므로 **그 항은 이 카드 안에서 결코 변하지 않는다** — 서술 정확성을 재는 항은 애초에 하나도 없었다.

### §F.3 후보 — 선고정 조상 판정 알고리즘 (구 `AC-RLP-007`)

**무엇을.** 판별식 확정 커밋이 LIVE 표지 커밋의 **진조상**임을 부모 그래프 도달성으로 판정한다. `git log --format='%H %P' --topo-order <리터럴 base>..HEAD`로 그래프를 얻고(리터럴 인자만 쓰므로 워크트리 가드를 통과한다), 각 LIVE 표지에서 부모를 거꾸로 걸어 조상 집합을 만들고, **모든 감시 커밋이 모든 LIVE 표지의 진조상**이며 한 커밋이 두 목록에 동시에 들지 않음을 본다.

[HARD] **다시 쓰지 말고 물려받는다.** 알고리즘은 iter-3 감사에서 여섯 케이스로 검증됐다 — 재유도는 낭비이고, 재작성은 검증을 버리는 일이다.

**상속 증거 경로: `.moai/reports/t1171/plan-audit-iter3.md`** (§4 「`AC-RLP-007` 조상 판정 — 여섯 케이스」). 검증된 여섯 케이스:

| # | 케이스 | 기대 |
|---|---|---|
| 1 | 선형 조상(감시가 LIVE의 조상) | `true` |
| 2 | 평행 가지(감시가 LIVE와 동시, 조상 아님) | `false` — 위상 순서로는 가릴 수 없던 경우 |
| 3 | 한 LIVE 표지의 조상이지만 다른 LIVE 표지의 조상은 아님 | `false` — 감사자가 추가한 케이스 |
| 4 | 같은 커밋이 감시와 LIVE 표지를 함께 담음 | `false` |
| 5 | 빈 입력 | `false` |
| 6 | **실제 병합 이력**(병합 둘을 통과) | `true` |

**함께 물려받는 미해결 항.** iter-3 F3 — M4가 제시한 복구 경로(LIVE 뒤 감시 파일을 고치면 그 실행을 `INVALID`로 기록하고 재실행)를 판정식에 넣으면 여전히 `false`가 된다. 즉 **동결을 깬 뒤 게이트를 되살리는 경로가 없다.** 후속 카드는 그 복구 경로를 설계하거나, 복구 불가를 명시하고 재실행 규정을 그에 맞게 고쳐야 한다.
