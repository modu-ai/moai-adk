---
id: SPEC-CODEX-REVIEW-TARGET-001
title: "수락 기준 — codex native review/start target 계약"
version: "0.3.0"
created: 2026-09-01
---

# SPEC-CODEX-REVIEW-TARGET-001 — 수락 기준

## §A 관측 규율 [HARD]

모든 AC 는 **직렬화된 JSON-RPC 요청 라인**(스텁 세션의 `sess.sent[2]`)을 관측 대상으로 삼는다. 스텁은 요청과 무관한 스크립트를 되돌려주므로(spec.md §A.5), 반환된 verdict 는 어떤 AC 의 근거도 되지 못한다.

`inconclusive` 를 통과로 세는 단언은 이 SPEC 이 고치려는 결함 그 자체이므로 금지한다.

## §B 스키마 required 집합 (판정 기준의 원천)

`.moai/reports/t399/schema/v2/ReviewStartParams.json` 에서 측정:

| variant | required | 분류 | 근거 |
|---|---|---|---|
| `uncommittedChanges` | `type` | 직렬화 가능 | 서버가 required 를 전부 채운다 |
| `baseBranch` | `branch`, `type` | 직렬화 가능 | 서버가 `branch` 를 해석한다(REQ-CRT-003) |
| `commit` | `sha`, `type` | 직렬화 불가 | 서버에 `sha` 원천이 없고 도구 표면도 받지 않는다 |
| `custom` | `instructions`, `type` | 직렬화 불가 | 서버에 `instructions` 원천이 없다 |

`분류` 열이 AC-CRT-006 / 006b 의 순회 대상을 가른다. 다섯째 variant 가 생기면 이 열에 값을 넣기만 하면 되고, 두 AC 의 단언 자체는 변하지 않는다.

## §C RED 확립 규율 [HARD]

AC-CRT-001 · 002 · 004 · 005 · 006b · 010 은 **변경 전 구현에서 실패해야 한다**. 실패는 서술이 아니라 **관측**으로 기록한다: 프로덕션 변경을 넣기 전에 신규 검사만 추가해 실행하고, `--- FAIL` 행을 담은 출력을 `.moai/reports/t399/red/` 아래에 남긴 뒤 그 경로를 progress.md §E.2 에 인용한다.

AC-CRT-010 의 RED 는 종류가 다르다: 계약 판독에서 유도한 예측이 아니라 **실 codex 가 요청을 거절하는 것을 관측한** 실패다. 그 거절 응답(JSON-RPC error 본문)을 그대로 남긴다.

셀렉터 0매칭이 초록으로 보이는 사고를 막기 위해, RED 실행은 `-v` 로 함께 돌려 `=== RUN` 행에서 대상 케이스가 실제로 실행됐음을 확인한다.

---

## §D AC 매트릭스

| AC | 요구 | RED 로 시작 |
|---|---|---|
| AC-CRT-001 | REQ-CRT-002 | 예 |
| AC-CRT-002 | REQ-CRT-003 | 예 |
| AC-CRT-003 | REQ-CRT-006 | 아니오 (회귀선, 초록으로 시작) |
| AC-CRT-004 | REQ-CRT-004 + REQ-CRT-008 | 예 |
| AC-CRT-005 | REQ-CRT-005 | 예 |
| AC-CRT-006 | REQ-CRT-001 (직렬화되는 variant) | 부분 (baseBranch 행만) |
| AC-CRT-006b | REQ-CRT-001 + REQ-CRT-005 (직렬화되지 않는 variant) | 예 |
| AC-CRT-007 | REQ-CRT-001 | 아니오 (기존 공허 검사 교정) |
| AC-CRT-008 | `§C 규율` (RED 확립 절차 — 특정 REQ 의 행동 검증이 아니다) | 해당 없음 |
| AC-CRT-009 | REQ-CRT-007 | 아니오 (도구 표면 서술) |
| AC-CRT-010 | REQ-CRT-002 (라이브 왕복 층) | 예 — 스키마상 거절이 **예측되며**, M2b 가 그 거절을 관측해 §C 규율대로 기록한다 |

---

### AC-CRT-001 — baseBranch 요청이 비어 있지 않은 branch 를 싣는다

**Given** 스크립트 스텁 codex 세션과 base 브랜치가 해석 가능한 프로젝트 루트가 주어지고,
**When** `handleCodexAudit` 가 `mode=native`, `target=baseBranch` 로 호출되면,
**Then** 3번째로 전송된 요청은 `review/start` 이고, 그 `params.target` 은 JSON 객체이며, `type == "baseBranch"` 이고 `branch` 가 존재하며 빈 문자열이 아니다.

### AC-CRT-002 — 해석 우선순위가 관측된다

순회 대상은 REQ-CRT-003 의 정렬된 사슬(plan.md §C)이다. `git_strategy.worktree_base_branch` 는 이 경로에서 읽히지 않으므로 어느 단계에도 나타나지 않는다(spec.md §A.7).

**Given** `origin/HEAD` 가 해석되는 프로젝트 루트가 주어지고,
**When** `mode=native`, `target=baseBranch` 로 감사가 실행되면,
**Then** 전송된 `target.branch` 는 `origin/` 접두사를 뗀 그 이름이다.

**And Given** `origin/HEAD` 가 부재하고 `main` 이 ref 로 해석되는 루트가 주어지면,
**Then** `target.branch` 는 `main` 이다 — 1단계 부재가 조용한 건너뜀이 되지 않는다.

**And** 각 단계는 이름을 반환하기 전에 그 이름이 해당 트리에서 ref 로 해석되는지 확인한다(REQ-CRT-003 후반절). 해석되지 않는 이름을 반환하는 구현은 이 AC 를 통과하지 못한다.

[HARD] 이 AC 는 `worktree_base_branch` 를 **읽지 않는 것**까지 함께 지킨다: 그 키가 설정돼 있고 그 값이 `origin/HEAD` 가 가리키는 이름과 **다른** 루트에서, 전송된 `target.branch` 는 여전히 `origin/HEAD` 쪽 이름이어야 한다. 이 트리에서는 두 값이 우연히 일치하므로(spec.md §A.7), 이 절을 검사하려면 픽스처에서 둘을 **갈라 놓아야** 한다 — 일치하는 루트에서는 두 설계가 구별되지 않는다.

### AC-CRT-003 — uncommittedChanges 요청 형태가 변하지 않는다 (회귀선)

**Given** 스크립트 스텁 codex 세션이 주어지고,
**When** `mode=native`, `target=uncommittedChanges` 로 감사가 실행되면,
**Then** 전송된 `params.target` 은 정확히 `type` 키 하나만 가진 객체이고 그 값은 `uncommittedChanges` 이며, `branch` / `sha` / `instructions` 중 어느 키도 존재하지 않는다.

이 AC 는 변경 전에도 초록이어야 한다. 초록이 아니면 회귀선이 아니라 이미 깨진 상태이며, 그 사실 자체를 먼저 보고한다.

### AC-CRT-004 — base 해석 실패가 조용히 다른 리뷰가 되지 않는다

**Given** base 브랜치를 해석할 수 없는 프로젝트 루트(원격 기본 head 부재 + `origin/main` 부재 + `main` 부재)가 주어지고,
**When** `mode=native`, `target=baseBranch` 로 감사가 실행되면,
**Then** `review/start` 요청은 **전송되지 않고**, `target.type == "uncommittedChanges"` 인 요청도 대신 전송되지 않으며, 원인이 아래 후보별로 지정된 **필드**에 실린다.

관측 필드는 plan.md §B 의 결정에 따라 하나로 고정된다. "반환 JSON 어딘가에 원인처럼 보이는 문자열이 있다"는 판정은 금지한다 — §A 의 관측 규율이 가장 헐거워지는 자리가 여기이기 때문이다.

| plan §B 후보 | 관측 필드와 단언 |
|---|---|
| **(가)** 원인을 명명한 `inconclusive` | `Verdict == "inconclusive"` **그리고** `Summary` 가 base 해석 불가를 명명 **그리고** 그 `Summary` 가 `"codex binary not found in PATH"` 와 **문자열로 서로 다르다** |
| **(나)** 도구 오류 | `res.IsError == true` **그리고** 그 오류 텍스트가 base 해석 불가를 명명 |

판정은 네 가지를 모두 본다: 전송 라인의 부재, 대체 요청의 부재, 지정된 필드의 값, 그리고 다른 fail-open 원인과의 구별 가능성. 구별 가능성을 빼면 이 SPEC 이 새 `inconclusive` 를 하나 늘리고 그것을 기존 것과 섞어버린다.

### AC-CRT-005 — required 필드를 채울 수 없는 variant 는 직렬화되지 않는다

**Given** `commit` 또는 `custom` 을 가리키는 bare string 이 요청 조립부에 도달하고,
**When** `review/start` 파라미터가 조립되면,
**Then** `sha` 없는 `{"type":"commit"}` 또는 `instructions` 없는 `{"type":"custom"}` 이 직렬화되지 않는다.

이 AC 는 두 variant 를 도구 enum 에 노출하는 것을 **요구하지 않는다**(spec.md §E). 요구하는 것은 미완성 객체를 내보내지 않는 것 하나다.

### AC-CRT-006 — 직렬화되는 variant 는 required 집합을 만족한다 (속성형)

**Given** §B 표에서 **직렬화 가능**으로 분류된 각 행(`uncommittedChanges`, `baseBranch`)에 대해 그 variant 를 유발하는 입력이 주어지고,
**When** 요청이 조립되면,
**Then** 직렬화된 `target` 객체의 키 집합은 그 variant 의 required 집합을 **포함한다**.

### AC-CRT-006b — 직렬화되지 않는 variant 는 target 객체를 남기지 않는다 (속성형, 짝)

**Given** §B 표에서 **직렬화 불가**로 분류된 각 행(`commit`, `custom` — 서버가 `sha` / `instructions` 를 채울 수 없다)에 대해 그 variant 를 가리키는 입력이 주어지고,
**When** 요청이 조립되면,
**Then** 그 variant 의 `target` 객체가 **출현하지 않으며**, 다른 variant 의 `target` 을 실은 `review/start` 가 **대신 전송되지도 않는다**.

두 번째 절이 없으면 이 AC 는 조용한 대체를 통과시킨다: `commit` 입력에 `{"type":"uncommittedChanges"}` 를 대신 내보내는 구현도 "`commit` 의 target 객체는 출현하지 않았다"를 만족한다. 그것은 AC-CRT-004 가 `baseBranch` 에 대해 이미 닫은 구멍(`target.type == "uncommittedChanges"` 인 요청도 대신 전송되지 않는다)을 `commit` / `custom` 에 대해서만 열어 두는 것이다.

가설이 아니다: `internal/cli/mcp_codex.go:1004` 의 default 분기가 이미 `uncommittedChanges` 를 반환한다. 첫 절만으로는 현행 동작에 대해 **붉지 않으며**, §D 매트릭스가 이 AC 를 "RED 로 시작: 예"로 등재한 것을 실제로 성립시키는 것은 두 번째 절이다.

두 AC 는 한 쌍이며 §B 표의 모든 행을 남김없이 덮는다. 하나로 합치면 직렬화되지 않는 두 행이 관측할 객체가 없어 **조용히 건너뛰고**, 그 0매칭이 초록으로 읽힌다(§A 가 금지하는 바로 그 사고). 쪼갠 형태는 다섯째 variant 가 와도 그것을 두 분류 중 하나에 넣기만 하면 되므로, 표만 늘고 단언은 변하지 않는다는 원래 의도를 그대로 보존한다.

[HARD] 두 AC 모두 **순회한 행의 수를 먼저 세고** 판정한다. 분류 결과가 빈 집합이면 그 AC 는 아무것도 단언하지 않은 것이며, 통과가 아니라 결함으로 보고한다.

### AC-CRT-007 — baseBranch 를 든 기존 검사가 공허하지 않다

**Given** `TestCodexAudit_AdversarialDispatchesTurnStart` 가 `target: baseBranch` 를 넘긴다는 사실이 주어지고,
**When** 그 검사가 실행되면,
**Then** 그 값이 전송되지 않는다는 사실이 그 검사 안에서 단언되거나(adversarial 은 target 을 싣지 않는다), 그 검사의 `target` 인자가 이 SPEC 이 새로 세운 native baseBranch 검사로 옮겨져, 저장소 안에서 `baseBranch` 를 든 codex 검사가 **아무것도 단언하지 않는 상태로 남지 않는다**.

### AC-CRT-008 — RED 가 관측으로 기록된다

**Given** 프로덕션 변경 이전의 트리가 주어지고,
**When** §C 가 RED 로 지정한 검사(AC-CRT-001 · 002 · 004 · 005 · 006b · 010)만 추가해 `-v` 로 실행하면,
**Then** 각 검사의 `=== RUN` 행이 출력에 나타나고(셀렉터 0매칭 아님), 각 검사에 대응하는 `--- FAIL` 행이 나타나며, 그 출력이 `.moai/reports/t399/red/` 아래 파일로 남아 progress.md §E.2 에서 경로로 인용된다.

관측된 실패가 없으면 그 검사는 이 결함을 겨누고 있지 않다는 뜻이다 — 통과시키지 말고 검사를 고친다.

### AC-CRT-009 — 도구 표면이 서버 해석을 서술한다

**Given** `codex_audit` 의 `target` 파라미터 설명이 주어지고,
**When** 이 SPEC 의 변경이 착지하면,
**Then** `baseBranch` 값의 설명이 **브랜치 이름을 서버가 해석한다**는 사실과 그 해석 원천을 서술하며, 호출자가 공급할 수 없는 값을 공급하도록 암시하지 않는다.

### AC-CRT-010 — 라이브 codex 가 baseBranch 요청을 거절하지 않는다

**Given** 실 codex 바이너리(PATH 존재 + `--version` 정상 — 부재 시 skip, `MOAI_SKIP_LIVE_CODEX=1` 로 opt-out; skip 조건 3종은 `internal/cli/codex_review_gate_live_test.go:33` 의 확립본을 재사용)와, base 브랜치 및 그로부터 갈라진 HEAD 를 가진 픽스처 git repo(`probeSeedRepo` 의 변형)가 주어지고,
**When** `mode=native`, `target=baseBranch`, `project_root=<픽스처>` 로 감사가 실행되면,
**Then** `review/start` 에 대한 응답이 JSON-RPC error 가 **아니고**(특히 missing field `branch` 계열이 아니고), 그 turn 이 `turn/started` 에 도달한다.

판정은 **거절 부재라는 양성 사실**이다. `inconclusive` 여부는 보지 않는다 — 그것이 이 카드가 고치려는 결함을 통과로 세는 방식이다.

이 AC 가 이 SPEC 안에 있어야 하는 이유:

- 카드의 [HARD] 회귀 문언("native 경로가 **실제로 codex 에 도달했는지** 단언할 것")이 지목하는 것이 정확히 이 층이다. 계약 층(AC-CRT-006)은 요청이 **스키마를 만족한다**는 것까지만 말하고, 그것이 곧 codex 의 수락이라는 것은 스키마 문서의 주장이지 이 SPEC 의 측정이 아니다.
- **현행 트리에서 붉을 것으로 예측된다**: 0.150.1 의 스키마가 `baseBranch` 에 `branch` 를 required 로 두므로 `{"type":"baseBranch"}` 는 거절될 **것으로 예측된다**. 이 예측의 근거는 스키마 판독이며, 그것이 관측이 아니라는 점이 M2b 를 두는 이유다 — 관측되면 이 SPEC 이 확보할 수 있는 RED 중 가장 강한 것이 된다(계약 판독이 아니라 실물 거절이므로). 아직 아무도 이 왕복을 돌리지 않았다.
- 라이브 왕복은 **재사용이지 발명이 아니다**(spec.md §G): 실 codex 세션에 `review/start` 를 보내는 선례와 skip 가드가 이미 저장소에 있다. §A.3 이 해석기에 적용한 재사용 규율을 검증 표면에도 적용한 결과다.
- 비용 논거가 성립하지 않는다: 스키마 위반은 turn 이 시작되기 전에 JSON-RPC 오류로 즉시 거절되고, 선례가 이미 `turn/started` 에서 세션을 끊어 리뷰 turn 전체를 청구서에서 뺀다.

부수적으로 이 AC 는 **지금 아무도 재지 않은 전제** 하나를 함께 잰다 — codex 가 `baseBranch` 값으로 로컬에 존재하지 않을 수도 있는 이름을 수용하는지(spec.md §F). REQ-CRT-003 의 ref 해석 가능성 요구가 서버 쪽 절반이고, 이 AC 가 codex 쪽 절반이다.

[HARD] 이 AC 의 skip 은 **통과가 아니다.** codex 부재로 skip 된 실행은 이 AC 에 대해 아무것도 관측하지 않은 것이며, DoD 판정에서 skip 은 미관측으로 기록한다.

---

## §E 완료 정의 (DoD)

1. AC-CRT-001 ~ 010 전부 초록. AC-CRT-010 이 codex 부재로 skip 됐다면 **미관측**으로 기록하며, 그 상태에서 DoD 를 충족으로 선언하지 않는다 — codex 있는 머신에서 한 번은 관측한다.
2. AC-CRT-008 의 RED 출력 파일이 존재하고 progress.md §E.2 가 그 경로를 인용한다. AC-CRT-010 의 RED(라이브 거절 관측)도 같은 자리에 남긴다.
3. 변경이 닿은 패키지(`internal/cli`)의 테스트가 초록 — 전 패키지 판정은 CI 몫.
4. `go vet` 초록 (darwin + windows 교차 컴파일 포함 — 두 관문은 서로 다른 것을 본다).
5. spec.md §E 의 범위 밖 항목 중 어느 것도 구현되지 않았음이 diff 로 확인된다 — 특히 `mcp_convergence.go` 무변경(t284 축 침범 금지).
6. **#1632 응답이 세 상태를 구별해 말한다**: (a) 이 카드가 닫은 것 — `mode=native, target=baseBranch` 가 더는 스키마 위반으로 거절되지 않는다 (b) 아직 열린 것 — "required native gate 가 조용히 아무것도 안 낸다"는 증상은 **다른 원인들로 남는다**(바이너리 부재, codex 오류, 그리고 이 SPEC 이 새로 만드는 base 해석 불가). 셋 다 `applyGateUnmet` 주석을 받되 verdict 는 `inconclusive` 이고, 그 값을 통과와 구별해 보이는 축은 t284 다 (c) 제보 #3(게이트 강제 여부)은 여전히 열린 **정책** 질문이다.

   6항이 DoD 에 있는 이유: 이 SPEC 자체는 오도하지 않지만(범위를 §E 에 명시했다), 착지 후 이슈에 "고쳤다"고만 답하면 **응답이** 오도한다. 오도가 발생할 수 있는 지점이 SPEC 이 아니라 응답이므로, 구속도 거기에 건다.

## §F 잔여 위험

- 스키마는 codex-cli 0.150.1 에서 측정됐고, AC-CRT-010 의 라이브 왕복도 실행 머신의 codex 판번호에서만 성립한다. 제보자가 쓴 0.149.0 과의 차이는 이 트리에서 관측 불가.
- AC-CRT-010 은 codex 가 요청을 **거절하지 않는다**는 것까지 잰다. 리뷰 결과의 품질이나 codex 가 실제로 어느 변경분을 비교했는지는 재지 않는다 — 선례와 같이 `turn/started` 에서 끊기 때문이다.
- §A.7 의 정렬 결정은 이 트리에서 두 후보가 우연히 같은 값을 내는 조건에서 내려졌다. 결정의 근거는 그 우연이 아니라 대칭성이지만, 두 값이 갈리는 트리에서의 동작은 이 SPEC 이 관측하지 않았다.
