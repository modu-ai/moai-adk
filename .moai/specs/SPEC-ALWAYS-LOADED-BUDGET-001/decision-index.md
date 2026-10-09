# decision-index.md — SPEC-ALWAYS-LOADED-BUDGET-001

Plan 단계에서 드러났으나 해소되지 않은 결정. 각 행은 무엇이 왜 미결인지만 적고, 선호 답을 담지 않는다.

### Q1: 상시 표면의 구속 줄을 역할 주입 경로로 옮기는 것이 `SPEC-ALWAYS-LOADED-HEADROOM-001` 이 운영자 게이트로 남긴 「구속 조항 동결 해제」에 해당하는가?

Label: RESOLVED (operator answer)
Authority anchor: Source: operator answer (AskUserQuestion, 2026-10-03, leader session `df44e022`)
Why unresolved: 동결(`SPEC-ALWAYS-LOADED-DIET-002` REQ-ALD2-002·003)은 그 SPEC 의 구현에 걸린 조항이고, 구속 줄의 on-demand 이동과 재작성을 금했다. 역할 주입 경로는 on-demand 가 아니라 역할 세션의 세션 시작 시점 경로지만, 일반 세션에서는 그 줄이 실리지 않는다. 이 이동이 동결의 취지 안인지 밖인지를 정한 기록이 커밋된 트리에 없다. 이 판단 전에는 M0 를 시작할 수 없다.
Operator verdict: RESOLVED — 이 SPEC 에 한해 HEADROOM-001 의 구속 조항 동결을 해제한다. 역할 주입(리더 전용 규칙을 SessionStart 로 리더·레인 세션에 전달)과 의미를 보존하는 구속 `[HARD]` 줄의 압축 재작성을 모두 허용한다. 조건: 감사자가 의미 보존을 확인할 수 있는 구속 원장(구속 줄마다 전/후), 구속 의무는 하나도 떨어뜨리지 않는다.

### Q2: 예산 상수를 115,000 으로 둘 것인가, 다른 값으로 둘 것인가?

Label: RESOLVED (operator answer)
Authority anchor: Source: operator answer (AskUserQuestion, 2026-10-03, leader session `df44e022`)
Why unresolved: 115,000 은 120,000 한도와 비다이어트 구간 증가 실측(`research.md` §3, 표본 두 구간)에서 나왔다. 그러나 구속 줄 축자 보존 아래 도달 가능한 하한은 아직 측정되지 않았고(`research.md` §2 는 가설), 하한이 115,000 을 넘으면 이 값은 착지할 수 없다. M0 하한 측정이 필요하다.
Operator verdict: RESOLVED — Q1 판정과 함께 115,000 목표를 유지한다. M0 는 여전히 하한을 재고, 압축 재작성으로도 닿지 않을 때만 멈춘다.

### Q3: `internal/hook/instructions_loaded.go` 의 `sessionCharBudget`(210,000)을 런타임 한도와 맞출 것인가?

Label: OUT OF SCOPE (operator answer)
Authority anchor: Source: operator answer (AskUserQuestion, 2026-10-03, leader session `df44e022`)
Why unresolved: 210,000 은 120,000·150,000 보다 높아 사용자가 런타임 경고를 받는 동안에도 훅 자문은 침묵한다. 이 SPEC 은 그 상수를 범위 밖으로 두었고, 바꾸려면 새 결정 기록이 필요하다.
Operator verdict: 이 SPEC 범위 밖으로 유지한다(`sessionCharBudget` 210,000 변경 없음). 결정 자체는 열린 채로 남는다.

### Q4: SessionStart 훅의 `additionalContext` 를 런타임이 잘리지 않고 전달하는 최대 크기는 얼마인가?

Label: EVIDENCE-NEEDED
Authority anchor: (없음)
Why unresolved: `factory-dispatch.md`(구 `kanban-dispatch.md`, t1399 개명) 전체 본문은 흡수 뒤 이 트리(`2771626b5`)에서 UTF-16 27,423이다(`research.md` §1.2). 런타임이 훅 출력을 일정 크기에서 자르거나 파일로 돌리는지는 이 plan 실행에서 관측하지 않았다. REQ-ALB-010 의 상한값이 이 측정에 걸려 있다.
Operator verdict: (미결 — 운영자 답 2026-10-03: run M0 에서 측정)

### Q5: 역할 한정 규칙의 전체 본문을 제자리에 두고 최상위 `paths:` 를 붙여 상시 표면에서 빼는 배치가 「`paths:` 만 붙이지 말라」는 지시와 양립하는가?

Label: RESOLVED (leader decision)
Authority anchor: Source: leader decision under mission contract `07d28c4b` (relayed with plan-audit iter1 defect D7, 2026-10-03)
Why unresolved: 배치 이유는 경로 고정 소비자 20여 개(`research.md` §4)다. 전달은 SessionStart 주입이 맡고 `paths:` 는 상시 표면 이탈만 담당하지만, 겉모양은 금지된 형태와 같다. 대안(본문을 `.claude/rules/` 밖으로 옮기기, 바이너리 내장 본문을 주입하기)은 경로 고정 소비자 전부를 함께 바꾸거나 `moai update` 갱신 경로에서 벗어난다.
Operator verdict: ACCEPTED — 역할 한정 규칙 본문은 최상위 `paths:` 를 단 채 제자리에 둘 수 있다. 이것은 비전달 배치이며, 전달은 SessionStart 주입과 무표지 진입점의 읽기 지시(REQ-ALB-007·024)가 맡고, 리더가 규칙 없이 뜨는 경로는 REQ-ALB-011·024 가드가 막는다. 따라서 「`paths:` 만 붙이기」에 해당하지 않는다.

### Q6: M0 하한이 115,000 을 넘을 때, 예산 초과 상태로 가드를 기록용으로 착지시킬 것인가, 카드를 멈출 것인가?

Label: RESOLVED (operator answer)
Authority anchor: Source: operator answer (AskUserQuestion, 2026-10-03, leader session `df44e022`)
Why unresolved: REQ-ALB-022 는 멈추고 보고하는 데까지만 정한다. 그 뒤 선택지는 예산 상향, 구속 줄 재작성 허용, 기록용 가드 착지 가운데 하나이며 모두 운영자 판단이다.
Operator verdict: RESOLVED — Q2 와 같은 답: 압축 재작성까지 적용한 하한이 115,000 을 넘을 때만 M0 에서 멈추고 보고한다. 그 뒤 처분은 다시 운영자에게 올린다.

### Q7: 이 SPEC 의 `phase` 를 `v3.2.0 target` 으로 둘 것인가?

Label: RESOLVED (operator answer)
Authority anchor: Source: operator answer (AskUserQuestion, 2026-10-03, leader session `df44e022`)
Why unresolved: v3.2.0 릴리스 미션이 진행 중이다. 이 카드가 그 배치에 들어가는지, 다음 릴리스로 가는지는 리드의 릴리스 범위 판단이다.
Operator verdict: RESOLVED — v3.2.0 에 넣지 않는다. `phase` 는 「next release after v3.2.0」, v3.2.0 컷 뒤 develop 에 착지한다.

### Q8: 역할 표지 집합을 무엇으로 볼 것인가 — 현재 `MOAI_FACTORY_*`·`MOAI_KANBAN_*` 인가, 카드 t1399 착지 뒤의 집합인가?

Label: EVIDENCE-NEEDED
Authority anchor: (없음)
Why unresolved: t1399 는 칸반 모드를 없애고 런처 진입을 `-f`/`-l` 로 바꾸는 중이다(M10 대기). 착지 시점의 표지 집합은 그 병합 뒤에야 관측할 수 있다.
Operator verdict: (미결 — 운영자 답 2026-10-03: run M0 에서 측정)

### Q9: plan-audit 가 Tier L 상한(3회)에서 FAIL 로 끝났을 때 추가 델타 회차를 둘 것인가?

Label: RESOLVED (leader decision)
Authority anchor: Source: leader decision under mission contract `07d28c4b` (relayed with plan-audit iter3, 2026-10-03; the same rule applied to t1399 and t1454)
Why unresolved: iter3 은 FAIL 0.80 으로 상한에 닿았다. 점수 하락은 회귀가 아니라 새로 발견된 결함 두 건 때문이었다. 상한 뒤 처분은 레인이 정할 수 없다.
Operator verdict: RESOLVED — 델타 1회(iter4)를 허용한다.

### Q10: plan-audit 4회차(추가 델타)도 FAIL 로 끝났을 때 SPEC 을 어떻게 처분할 것인가?

Label: RESOLVED (leader decision)
Authority anchor: Source: leader decision under mission contract `07d28c4b` (relayed with plan-audit iter4, 2026-10-03; same rule as t1356)
Why unresolved: iter4 는 FAIL 0.81(`.moai/reports/t1469/plan-audit-iter4.md`)로 두 번째 상한에 닿았다. 그 뒤 처분은 레인이 정할 수 없다.
Operator verdict: HOLD 2026-10-03 (second ceiling hit). Resume plan: (a) scope reduction — move the kanban-dispatch-detail.md split to a separate card, which removes N4-2; (b) fix N4-1 (always: location must be verified as a deployed always-loaded member), N4-3, N4-4 and optional N4-5..7; (c) adopt an STE-lite (ASD-STE100-inspired: one instruction per sentence, imperative, active voice, short sentences) compression method for the meaning-preserving rewrite, measured on one rule file first; then one delta audit.

### Q11: plan-audit 5회차(리더 승인 최종 재판독)도 FAIL 로 끝났을 때 N5-1 과 나머지 결함을 어떻게 처분할 것인가?

Label: RESOLVED (leader relay of operator disposition)
Authority anchor: Source: operator disposition relayed by the leader dispatch (Q10(a) scope reduction, 2026-10-04)
Why unresolved: iter5 는 FAIL 0.825(`.moai/reports/t1469/plan-audit-iter5.md`)로 N5-1 companion-예외 모순 클러스터를 냈고, 판정서 §Recommendation 의 세 처분(범위 축소 / PASS-with-debt / 운영자 재수리 승인) 가운데 선택은 레인이 정할 수 없다.
Operator verdict: RESOLVED — 2026-10-04 운영자 처분: Q10(a) 범위 축소를 채택한다. `factory-dispatch-detail.md` 분할은 후속 카드로 옮긴다(제거 범위 기록은 `progress.md` §E.1); STE-lite 는 run-phase 로 유보된다(Q10 (c) 항목 유지). 이 처분은 0.6.0 이 택한 split-유지 수리 경로(N4-2 원장 확대 + companion 예외 행)를 대체한다(supersedes). Removed scope → follow-up card candidates: (1) the `factory-dispatch-detail.md` split — UTF-16 40,659 > 40,000 per-file budget, live MUST obligations at `factory-dispatch-detail.md:L150`; the follow-up card needs the lossless-split check machinery (pre-split unit set vs post-split companion union) and the companion-origin anchor-source predicate design (anchor-source file carries top-level `paths:`, frozen at the anchor) that iter5 N5-1(b) found missing; (2) `spec-workflow.md` (40,052) is NOT a candidate here — the leader handles it separately (recorded as an excluded observation only).
