---
id: SPEC-CODEX-VERDICT-SYNTH-001
title: "인수 기준 — 미인식 서식이 통과로 떨어지지 않는다"
version: "0.5.0"
created: 2026-08-24
updated: 2026-08-25
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: internal/cli
lifecycle: spec-anchored
tier: S
tags: "codex, verdict, property-based, mutant, version-drift"
---

# 인수 기준 — SPEC-CODEX-VERDICT-SYNTH-001

## §A 검증 규율 [HARD]

- 모든 AC 는 **Go 테스트 단언**으로 검증한다. grep 형태 AC 는 쓰지 않는다 — 대상이 문자열의 존재가 아니라 **채택된 값** 이기 때문이다.
- 실행: `go test ./internal/cli/... -timeout 600s`. `go test ./...` 금지.
- 각 AC 는 **자기가 죽이는 mutant** 를 함께 적는다. mutant 를 적지 못하는 AC 는 아무것도 관측하지 않는 AC 다.
- 기대값의 근거는 `.moai/reports/t229/premise-revision.md` §2 의 7행 실측표다.

## §B [HARD] 중심 AC 는 속성형이다 — 서식 열거형이 아니다

AC-CVS-001 은 **서식 목록이 아니라 속성**을 건다. 이유를 여기 적어 둔다.

서식을 열거하는 AC 는 이렇게 실패한다. AC 가 "`FAIL 0.75` 를 읽는다 / `Blocking` 표를 읽는다" 를 요구하면, 구현자는 정확히 그 두 서식을 읽는 정규식을 추가하고 AC 는 초록이 된다. 그런데 **그 AC 의 목록은 구현이 대상으로 삼은 목록과 같아졌으므로**, AC 는 구현을 검증한 것이 아니라 구현을 그대로 되풀이한 것이다. codex 가 다음 버전에서 세 번째 서식을 내면 같은 자리가 다시 뚫리고, 그때도 AC 는 초록이다.

그래서 걸어야 할 것은 **"임의의 미인식 입력이 `pass` 로 떨어지지 않는다"** 라는 속성이다. 서식 corpus 는 그 속성의 **증인**일 뿐이며, 요구사항 자체가 아니다.

판별 기준 하나: **corpus 에 케이스를 하나 더 넣는 일이 단언문 수정을 요구하면, 그 AC 는 속성형이 아니다.** 테스트는 corpus 를 순회하며 동일한 단언을 적용해야 한다.

### corpus 구성 원칙

- 구성원끼리 **서로 닮지 않아야** 한다. 한 정규식이 둘 이상을 동시에 덮으면 그만큼 증인 수가 줄어든다.
- 각 구성원은 판정 라벨(`Verdict:`)·점수 표기·심각도 태그 불릿을 **하나도** 담지 않는다(단, 점수 표기 구성원은 AC-CVS-002 전용으로 별도 취급).
- corpus 는 자유롭게 늘릴 수 있다. 늘리는 것이 AC 를 바꾸지 않는다.

### corpus 구성원 (초기 8건)

| # | 구성원 | 성격 |
|---|---|---|
| C1 | `Blocking` 행을 담은 마크다운 표 + `merge_status: blocked` | 실측표 4행 — **현재 `pass` 로 합성됨** |
| C2 | 번호 목록 형태의 지적 (`1. Missing input validation at api.go:31`) | 불릿이되 심각도 태그가 없음 |
| C3 | JSON blob (`{"result":"blocked","issues":[...]}`) | 구조화 출력이나 서식 미상 |
| C4 | 제목만 있는 본문 (`## Review Summary` 뒤 내용 없음) | 절단된 응답 |
| C5 | 산문 1줄 (`I walked the diff and moved on.`) | 실측표 5행 — **현재 `pass` 로 합성됨** |
| C6 | 한국어 본문 (`차단 사유 2건을 확인했습니다.`) | 비영어 — 영어 키워드 가정을 깬다 |
| C7 | 빈 문자열 | 응답 없음 |
| C8 | 점수 표기 `FAIL 0.75 / 1.00` + 차단 2건 | 실측표 2행 — AC-CVS-002 가 `fail` 로 별도 고정 |

## §B-2 [HARD] 조합 corpus — AC-CVS-006 의 증인

AC-CVS-006 도 §B 와 같은 이유로 **속성형**이다. 다만 §B 의 corpus 가 *서식* 을 증언한다면, 여기서는 **신호 집합의 조합** 을 증언한다.

규칙은 spec.md §A.5 의 **P-CONS** 다 — 채택값은 신호 **집합의 최댓값**이며, 순위는 `fail` > `inconclusive` > `pass`. 집합에는 순서가 없으므로 이 규칙은 신호가 몇 개로 늘든 그대로다.

**쌍을 열거하는 AC 로 쓰면 안 된다.** "명시 × 점수" 쌍만 거는 AC 는 그 쌍만 특수 분기로 처리한 구현에 통과당하고, 넷째 신호가 오는 날 같은 자리가 다시 뚫린다 — 오늘 이 저장소에서 반복적으로 겪은, 구멍을 막으면 한 겹 옆으로 옮겨 가는 바로 그 모양이다.

판별 기준은 §B 와 동일하다: **행을 하나 더 넣는 일이, 또는 신호를 넷째로 늘리는 일이 단언문 수정을 요구하면 그 AC 는 틀렸다.** 테이블 구동으로 쓰고 단언은 하나만 둔다.

### 조합 구성 원칙

- 각 행은 `(리뷰 본문, 그 본문이 산출하는 신호 집합, 기대 채택값)` 이다. 기대값은 언제나 집합의 최댓값이며 손으로 예외를 두지 않는다.
- **같은 집합을 다른 텍스트 순서로 담은 행을 최소 한 쌍 포함한다** — 순서 의존 구현을 가르는 유일한 증인이다.
- **`stated × scored` 가 아닌 쌍과 3-신호 행을 포함한다** — 조합 커버리지를 넓힌다. 단, **이 행들이 쌍 특수화 구현을 가르지는 못한다**(아래 각주).
- **신호가 갈리지 않는 행을 포함한다** — 무조건 보수적으로 답하는 구현을 가르는 증인이다.
- 넷째 신호가 생기면 행만 추가한다. 단언문도, 이 원칙 목록도 바뀌지 않는다.

### 조합 corpus 구성원 (초기 8행)

| # | 리뷰 본문이 담는 것 | 신호 집합 | 기대 채택값 | 이 행이 증언하는 것 |
|---|---|---|---|---|
| K1 | `Verdict: fail` → `PASS 0.95 / 1.00` | {fail, pass} | `fail` | 기본 세탁 반례 (mutant e) |
| K2 | `PASS 0.95 / 1.00` → `Verdict: fail` (K1 과 **같은 집합, 순서만 반대**) | {fail, pass} | `fail` | **순서 무관** (mutant f) |
| K3 | `Verdict: pass` → `FAIL 0.20 / 1.00` | {pass, fail} | `fail` | 반대 방향 세탁 |
| K4 | `Verdict: inconclusive` → `PASS 0.95 / 1.00` | {inconclusive, pass} | `inconclusive` | 순위 중간값 (`pass` 아님) |
| K5 | `PASS 0.95 / 1.00` + `- [P1] path traversal at fs.go:44` | {pass, fail} | `fail` | **`stated × scored` 가 아닌 쌍** (mutant g) |
| K6 | `Verdict: pass` + `INCONCLUSIVE 0.50 / 1.00` + `- [P2] weak hash at auth.go:88` | {pass, inconclusive, fail} | `fail` | **3-신호** (mutant g) |
| K7 | `INCONCLUSIVE 0.50 / 1.00` → `Verdict: pass` (K4 와 집합은 같고 **어느 신호가 어느 값을 담는지가 뒤바뀜**) | {inconclusive, pass} | `inconclusive` | **신호 역할 교환** — 값이 신호 종류에 묶여 있지 않음 |
| K8 | `Verdict: pass` → `PASS 0.99 / 1.00` (**갈리지 않음**) | {pass} | `pass` | 과잉 보수 방지 |

### [HARD] 각주 — 오늘 이 corpus 가 가르지 **못하는** 것 (감사 iter2 N1)

**쌍 특수화 구현은 오늘 어떤 행으로도 검출되지 않는다.** `stated × scored` 만 특수 분기하고 불릿을 마지막에 적용하는 구현은 K1~K8 **여덟 행 전부를 통과한다**.

구조적인 이유다. `codexFindingBullet` 은 `fail` **하나만** 기여하고, `fail` 은 P-CONS 순위의 최상단이다. 따라서 불릿을 마지막에 적용하는 것은 "집합의 최댓값을 취한다" 와 **항상** 결과가 같다. **신호가 셋이고 그중 하나가 `fail` 전용인 동안, 쌍 특수화와 일반 규칙은 모든 도달 가능한 입력에서 같은 값을 낸다** — 구별할 입력이 존재하지 않는다.

[HARD] **이 등가성은 조건부이며, 넷째 신호가 생기는 날 깨진다.** 특수 분기 **밖에서** `fail` 이 아닌 값(`pass` 또는 `inconclusive`)을 기여할 수 있는 신호가 하나라도 추가되면, 그 순간부터 쌍 특수화 구현은 일반 규칙과 다른 값을 내기 시작한다.

따라서 넷째 신호를 도입하는 사람에게 남기는 지시: **K5·K6 이 당신을 지켜준다고 읽지 말 것.** 그 보호는 오늘 존재하지 않으며, 넷째 신호를 추가할 때 비로소 검출 가능해진다 — 즉 그때 **쌍 특수화를 가르는 행을 새로 추가해야 한다.** 조건부로 참인 보호를 무조건으로 읽는 것은 이 저장소가 반복적으로 겪은 실패 형태다.

이 corpus 가 오늘 실제로 가르는 것은 **순서 의존 구현(mutant (e))** 과 **조기 반환형 특수화(mutant (g-2))** 이며, 그 이상은 주장하지 않는다.

## §C 인수 기준

### AC-CVS-001 — 미인식 서식은 하나도 pass 가 아니다 (REQ-CVS-001) [중심 AC · 속성형]

**Given** §B 의 corpus 구성원 각각을 리뷰 본문으로 놓았을 때
**When** `codexMethodTurnStart`(adversarial) 모드로 `synthesizeReviewOutput` 을 호출하면
**Then** **단 하나도** `Verdict == "pass"` 가 아니다.

corpus 를 순회하며 동일한 단언을 적용한다. 구성원을 추가해도 단언문은 바뀌지 않는다.

> **죽이는 mutant (a)**: 점수 표기 정규식을 하나 추가했으나 `verdict := "pass"` fall-through 를 남긴 구현. **AC-CVS-002 만으로는 이 mutant 가 통과한다** — 점수 표기는 읽으니까. 이 AC 가 없으면 구조적 원인(G3)이 그대로 남는다.
> **죽이는 mutant**: `verdict := "pass"` 를 유지한 현재 구현. 사전구현 트리에서 이 AC 는 **corpus 8건 전부(C1~C8)에서 실패해야 한다** — 감사 iter1 이 현재 트리에서 측정한 결과 C1~C8 이 **모두 `pass` 로 합성된다**(C8 은 AC-CVS-002 가 `fail` 로 추가 고정). RED 기대를 일부(C1·C5·C7)로만 적으면 DoD 의 독립성 검사가 실제보다 약해진다.

### AC-CVS-002 — 점수 표기는 fail 로 읽힌다 (REQ-CVS-002)

**Given** corpus 구성원 C8 — 줄 머리 `FAIL 0.75 / 1.00` 과 차단 2건, 심각도 태그 불릿은 없음
**When** adversarial 모드로 호출하면
**Then** `Verdict` 는 **`"fail"`** 이다 — `inconclusive` 도 `pass` 도 아니다.

동일 형태로 `PASS 0.88 / 1.00` → `"pass"`, `INCONCLUSIVE 0.50 / 1.00` → `"inconclusive"`.

**그리고** 산문 안의 `the suite reported PASS 12 times before the regression` 은 판정으로 읽히지 않는다(오탐 방지) — 이 본문은 corpus 규칙에 따라 `pass` 가 아니어야 한다.

> **죽이는 mutant**: 점수 표기를 인식하지 않는 현재 구현 (실측표 2행: `pass` 로 합성). 사전구현 트리에서 실패해야 한다.
> **죽이는 mutant**: 모든 것을 `inconclusive` 로 떨어뜨려 AC-CVS-001 만 만족시키는 구현 — 이 AC 의 `fail` 단언이 잡는다. 두 AC 가 서로를 구속한다.
> **죽이는 mutant**: 점수 표기 정규식을 `(?i)(pass|fail)\s+[\d.]+` 처럼 느슨하게 잡아 산문을 판정으로 읽는 구현.

### AC-CVS-003 — native 무불릿 정상 리뷰는 pass 로 보존 (REQ-CVS-004) [보고 정확성]

**Given** 리뷰 본문이 `The change introduces no blocking issues.` 이고 불릿도 판정 라벨도 없을 때
**When** `codexMethodReviewStart`(native review) 모드로 호출하면
**Then** `Verdict` 는 **`"pass"`** 다 (실측표 6행 = 보존 대상).

빈 문자열 입력에 대해서도 native 모드에서는 `"pass"` 를 유지한다.

> **죽이는 mutant (b)**: 기본값을 **양쪽 모드 모두** `inconclusive` 로 뒤집은 구현. 이 mutant 의 해악은 **차단이 아니라 보고 손실**이다 — 게이트는 `fail` 접두사에서만 차단하므로(`isBlockVerdict`, `codex_review_gate.go:116-117`; 종단 `return allow, nil // pass / inconclusive ⇒ ALLOW`, `:109`) 정상 변경이 막히지는 않는다. 대신 codex 가 실제로 "차단 사유 없음" 을 말한 경우가 **판정을 못 낸 경우와 구분되지 않게 되고**, 수렴 계층의 fail-open 폴백(required 전부 inconclusive → claude 판정, `mcp_convergence.go:126-129`)과 뒤섞이며, 기존 테스트가 고정한 `pass` 계약(`codex_review_rpc_test.go:119`)이 깨진다.
> **모드 배선 증인은 C5 로 한다 (도달 가능한 입력)**: C5(산문 1줄)가 adversarial 에서 `pass` 아님(AC-CVS-001), native 에서 `pass`(이 AC) — 같은 입력이 모드에 따라 갈린다는 것이 모드 구분이 실제로 배선됐다는 증거다. C7(빈 문자열)도 같은 대비를 보이지만 **프로덕션에서는 도달하지 않는다** — `runTurn` 이 빈 리뷰 텍스트를 합성기 호출 이전에 `inconclusive` 로 단락시킨다(`mcp_codex.go:702-703`). C7 의 native `pass` 핀은 기존 테스트 보존용으로 유지하되, 배선 증인으로는 쓰지 않는다.

### AC-CVS-004 — 신호가 갈리면 보수 채택 + 기록 (REQ-CVS-003)

**Given** 리뷰 본문이 `Verdict: pass` 를 명시하면서 동시에 `- [P1] SQL injection at db.go:12` 불릿을 담고 있을 때
**When** `synthesizeReviewOutput` 을 호출하면
**Then** 채택된 `Verdict` 는 **`"fail"`** 이고, **동시에** 불일치가 결과에 기록돼 있다(비어 있지 않은 기록 필드).

**Given** 위 결과가 codex 백엔드의 `PerBackendVerdict` 로 실려 `converge` 에 들어갔을 때
**When** 다른 백엔드는 모두 `pass` 일 때
**Then** `DisagreementFlag` 는 `true` 이고 `ResidualRiskNote` 가 그 내용을 담으며, `OverallVerdict` 는 기존 정책이 산출하던 값 그대로다 — `disagreement_flag` 는 새 차단 범주가 아니다.

> **죽이는 mutant (c)**: 보수 채택은 하되 **아무것도 기록하지 않는** 구현. 값은 맞지만 왜 그렇게 됐는지가 사라져 다음 사고의 진단 근거가 없어진다. 기록 필드 단언이 잡는다.
> **죽이는 mutant (d)** [카드 대표 mutant]: 불일치를 **감지는 하되 관대한 쪽을 채택하는** 구현. 단언이 "기록이 있다" 만이면 통과하므로 **채택된 `Verdict` 값 단언이 필수**다.
> **죽이는 mutant**: 기록 필드를 `ReviewOutput` 에만 두고 `PerBackendVerdict` 로 전달하지 않아 `converge` 가 영원히 못 보는 구현 (필드는 추가됐는데 배선이 없는 형태).
> **죽이는 mutant**: 불일치를 새 차단으로 승격시켜 `OverallVerdict` 를 `fail` 로 바꾸는 구현 — `OverallVerdict` 불변 단언이 잡는다.
>
> **AC-CVS-006 과의 관계** (감사 iter2 N4): 이 AC 의 본문(`stated: pass` × 불릿 `fail`)은 P-CONS 의 한 **사례**이며, 그 쌍은 §B-2 조합 corpus 에 별도 행으로 넣지 않는다(중복이 된다). 이 AC 가 AC-CVS-006 위에 얹는 고유 내용은 **불일치 기록**이다 — 채택값이 보수적인지는 AC-006 이, 그 사실이 결과에 남고 `converge` 까지 전달되는지는 이 AC 가 담당한다. 둘은 겹치지 않으며 합치지 않는다.

### AC-CVS-005 — 기존 동작 회귀 방어 (REQ-CVS-004)

**Given** `TestSynthesizeReviewOutput_FindingBulletsMapToFail`(`internal/cli/codex_review_rpc_test.go:114`) 의 4개 입력
**When** native review 모드를 명시해 호출하도록 확장된 뒤에도
**Then** 4건 모두 기존 기대값(`fail`/`fail`/`pass`/`pass`)을 유지한다.

**Given** `.moai/reports/t229/live-probe-body.txt` 를 **원문 그대로** 픽스처로 읽었을 때
**When** adversarial 모드로 호출하면
**Then** `Verdict` 는 `"inconclusive"` 다 (실측표 1행 — 이미 통과하는 회귀 고정 지점이며 새 검출이 목적이 아니다).

**Given** `codex_task` 경로가 판정어도 불릿도 없는 본문을 받았을 때
**When** 태스크가 완료되면
**Then** 반환된 `CodexTaskResult.Output` 은 본문 텍스트 그대로이며 이 SPEC 이전과 동일하다.

> **죽이는 mutant**: 기존 테스트를 삭제하거나 기대값을 새 동작에 맞춰 고쳐 회귀를 은폐하는 구현.
> **죽이는 mutant**: 모드 fall-through 변경이 `codex_task` 의 출력 경로까지 흘러들어 텍스트가 바뀌는 구현.

### AC-CVS-006 — 채택값은 신호 **집합**의 최댓값이다 (spec.md §A.5 P-CONS) [속성형 · M2 가드]

**Given** 조합 corpus(§B-2)의 각 행이 규정하는 리뷰 본문
**When** `synthesizeReviewOutput` 을 호출하면
**Then** 채택된 `Verdict` 는 그 행의 기대값과 같다. 기대값은 예외 없이 **그 본문이 산출하는 신호 집합의 가장 보수적인 원소**이며, 보수 순위는 `fail` > `inconclusive` > `pass` 다.

조합 corpus 를 순회하며 **단 하나의 단언**을 적용한다 — 행을 추가해도, 신호가 넷째로 늘어도, 단언문은 바뀌지 않는다.

이 AC 는 채택값만 단언하며 내부 구현 형태에 관해 아무것도 요구하지 않는다 — 순위 테이블이든 대입 순서이든 조건 분기이든, 결과가 P-CONS 를 만족하면 통과한다.

> **죽이는 mutant (e)** [D1 이 지목한 것]: M2 를 **가장 자연스러운 방식으로** 쓴 구현 — `codexScoredVerdict` 를 기존 대입 열에 한 줄 더 얹는 형태. plan.md §C.1 의 신호 순서(`stated` → `scored` → `bullet`) 아래에서 나중 대입이 앞선 보수적 값을 덮으므로 K1 이 **`pass`** 로 합성된다. 이 AC 가 없으면 그 구현이 **AC-CVS-001~005 를 전부 초록으로 통과한 채** 착지하며, 이 SPEC 이 없애려던 §0 위반을 이 SPEC 이 새로 만들어 낸다.
> **죽이는 mutant (f)** [순서 의존]: 신호를 텍스트 등장 순서나 검사 순서에 따라 다르게 처리하는 구현. **K1 과 K2 는 같은 신호 집합을 서로 다른 텍스트 순서로 담고 있으므로 반드시 같은 값을 내야 한다** — 순서에 민감한 구현은 여기서 갈라진다.
> **mutant (g) [쌍 특수화] — 오늘은 어떤 테스트로도 가려낼 수 없다. 이 AC 는 이것을 잡지 못한다.** `stated × scored` 조합만 특수 분기로 처리하고 불릿을 마지막에 적용하는 구현(감사 iter2 변종 `Vg1`)은 **K1~K8 8행 전부를 통과한다**. 구조적인 이유가 있다 — `codexFindingBullet` 은 `fail` 만 기여하고 `fail` 은 P-CONS 순위의 최상단이므로, 불릿을 마지막에 적용하는 것은 "집합의 최댓값" 과 **항상** 일치한다. 즉 **신호가 셋이고 그중 하나가 `fail` 전용인 동안, 쌍 특수화와 일반 규칙은 행동상 구별되지 않는다.** 오늘의 결함이 아니라 **넷째 신호가 생기는 날 드러날 잠재 위험**이며, 그 넷째 신호가 특수 분기 밖에서 `fail` 이 아닌 값을 기여할 수 있을 때 비로소 구별 가능해진다. K5·K6 이 이 변종을 잡는다는 앞선 서술은 **거짓이었고 여기서 철회한다**(감사 iter2 N1).
> **죽이는 mutant (g-2)** [조기 반환형 쌍 특수화]: 위와 달리 특수 분기에서 곧바로 반환해 불릿을 아예 보지 않는 구현(변종 `Vg2`). 이것은 오늘도 틀리며 **K6(3-신호)이 유일하게 잡는다**.
> **죽이는 mutant**: 보수 순위를 `pass` > `inconclusive` 로 잘못 매긴 구현 — K4·K7 이 잡는다.
> **죽이는 mutant**: 무조건 `fail` 을 내는 과잉 보수 구현 — K8(신호가 갈리지 않는 행)이 잡는다.
> **죽이는 mutant**: 불릿 적용을 명시 verdict 존재 여부에 걸어 둔 구현(변종 `Vh`) — K1·K2·K4·K5 가 잡는다.
>
> **이 AC 가 실제로 사 오는 것 (부풀리지 않고)**: (a) **mutant (e)** — 순서 의존 구현. 오늘 측정으로 검출 가능하며 M2 를 쓰는 가장 자연스러운 방식이므로, 이것이 이 AC 의 실질적 검출력이다. (b) 규칙을 **문서화된 속성(P-CONS)으로 고정**해, 넷째 신호가 들어오는 날 그것이 구현 습관이 아니라 **명시된 계약**을 상대하게 만든다. 그 이상은 주장하지 않는다.
>
> **[중요] 이 AC 는 현재 트리에서 부분적으로 RED 다.** 감사 iter2 실측: **K3 → `pass`(기대 `fail`)**, **K7 → `pass`(기대 `inconclusive`)** 두 행이 오늘 실패하고, K1·K2·K4·K5·K6·K8 여섯 행은 이미 통과한다. 원인은 점수 표기 신호가 아직 없어 **집합에서 관대한 쪽인 `stated` 값만 읽히기 때문**이며, M1 의 fall-through 수정 여부와 무관하다(두 행 모두 `stated` 가 매치되므로). 앞서 "RED 로 시작하지 않는다" 고 단정한 것은 **거짓이었고 여기서 철회한다**(감사 iter2 N2) — mutant (e) 증인인 K1·K2·K4 에 한해서만 참이었다.
> **[HARD] K3·K7 의 기대값은 P-CONS 에서 도출된 것이며, 관측된 동작에 맞춰 고쳐서는 안 된다.** 이 두 행이 붉은 것은 구현이 아직 규칙에 부합하지 않는다는 뜻이지, 행이 틀렸다는 뜻이 아니다. 기대값을 현재 동작으로 낮추면 K3·K7 의 검출력이 조용히 사라지며, 그것이 이 카드가 다루는 실패 형태 그 자체다.

## §D 품질 게이트

| 항목 | 기준 |
|---|---|
| 테스트 | `go test ./internal/cli/... -timeout 600s` 전량 통과 |
| 정적 분석 | `go vet ./internal/cli/...` 무경고 |
| 크로스 플랫폼 | `GOOS=windows go vet ./internal/cli/...` 통과 |
| fail-open | 어떤 AC 도 hard error 반환을 기대하지 않음 |

## §E Definition of Done

- [ ] AC-CVS-001 ~ AC-CVS-006 전부 통과
- [ ] **AC-CVS-001 이 AC-CVS-002 와 독립적으로** 사전구현 트리에서 실패했음을 RED 출력으로 기록 — 둘이 같은 mutant 만 잡으면 mutant (a) 가 빠져나간다
- [ ] AC-CVS-001 의 RED 를 **C1~C8 전부**에 대해 기록 (일부만 적으면 위 독립성 검사가 실제보다 약해진다)
- [ ] **AC-CVS-006 의 부분 RED 를 기록** — 착수 전 트리에서 **K3(`pass`, 기대 `fail`)·K7(`pass`, 기대 `inconclusive`) 두 행이 붉고** 나머지 여섯 행은 이미 초록임을 출력으로 남긴다. **이 두 행을 초록으로 뒤집는 것은 M2**(점수 표기 신호 도입)이며, M1 의 fall-through 수정은 두 행 모두 `stated` 가 매치되므로 영향이 없다. "가드 AC 이므로 RED 기대를 기록하지 않는다" 던 이전 지시는 철회됐다(감사 iter2 N2)
- [ ] **K3·K7 의 기대값을 관측 동작에 맞춰 낮추지 않았음을 확인** — 기대값은 P-CONS 에서 도출된 것이다. 붉은 행을 "고쳐서" 초록으로 만드는 것은 판별력을 지우는 것이며, 이 카드가 다루는 실패 형태 그 자체다
- [ ] **AC-CVS-006 이 M2 착지 직후 초록임을 별도로 확인** — M2 를 대입 한 줄 추가로 구현했다면 여기서 붉어진다
- [ ] AC-CVS-006 의 단언문이 **조합 corpus 를 순회하는 단일 단언**임을 코드로 확인 — 행 추가나 넷째 신호 도입이 단언문 수정을 요구하지 않아야 한다(§B-2 판별 기준)
- [ ] K1/K2 쌍이 **같은 값**을 내는지 확인 (순서 무관 증인) · **mutant (e) 가 AC-CVS-006 에서만 실패하고 AC-CVS-001~005 는 전부 통과함을 확인** (이것이 실제로 검증 가능한 독립성 주장이다. 종전 항목 "K5·K6 이 쌍 특수화 구현을 실제로 가르는지 확인" 은 **검증 불가이므로 철회한다** — §B-2 각주 참조)
- [ ] AC-CVS-001 의 단언문이 corpus 를 순회하는 형태임을 코드로 확인 (구성원 추가가 단언문 수정을 요구하지 않음)
- [ ] `TestSynthesizeReviewOutput_FindingBulletsMapToFail` 삭제되지 않음
- [ ] §D 품질 게이트 전량 통과 출력이 progress.md §E.2 에 인용됨
- [ ] 착지 후 리드에게 t234 (= GitHub #1632) 착수 가능 신호 전달
