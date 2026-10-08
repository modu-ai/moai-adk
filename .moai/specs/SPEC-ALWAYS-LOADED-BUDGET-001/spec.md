---
id: SPEC-ALWAYS-LOADED-BUDGET-001
title: "배포 표면 상시 로드 지시문 예산 — 역할 한정 규칙의 SessionStart 주입, 대형 규칙의 core+companion 분할, 배포 템플릿 기준 회귀 가드"
version: "0.9.0"
status: in-progress
created: 2026-10-03
updated: 2026-10-08
author: manager-spec
priority: P1
phase: "next release after v3.2.0"
module: "internal/template/templates/.claude/rules/moai, internal/template/templates/CLAUDE.md, internal/hook, internal/template"
lifecycle: spec-anchored
tags: "always-loaded, instruction-budget, role-gating, session-start, rule-split, regression-guard, template"
tier: L
related_specs: [SPEC-ALWAYS-LOADED-DIET-001, SPEC-ALWAYS-LOADED-DIET-002, SPEC-ALWAYS-LOADED-HEADROOM-001, SPEC-INSTRUCTION-BUDGET-SCOPE-001, SPEC-INSTRUCTIONS-BUDGET-001]
---

# SPEC-ALWAYS-LOADED-BUDGET-001

## HISTORY

| 날짜 | 버전 | 변경 | 작성자 |
|---|---|---|---|
| 2026-10-08 | 0.9.0 | REQ-ALB-022 중단이 Q6 이 약속한 「그 뒤 처분」으로 귀결되고, 운영자 처분 (a)(리더 전달, 2026-10-08)이 예산 상수를 규칙-only 축으로 재도출했다 — 재계획이 아니라 진행 중 run 의 표적 수정. 근거 수치 세 개: M3 실측 압축 바닥 122,000–126,000(종전 전면-표면 상수 115,000 위에서 측정 — 도달 불가), 규칙-only 소계 114,594(예산 테스트 실측 `deployed-surface-total=138278` − `AGENTS.md` 23,684), `AGENTS.md` 23,684(표면에는 있으나 이 SPEC 편집 범위 밖 — t1450 계정). 명시된 해석(운영자 의도가 다를 때 드러나도록 기록): 상수 = 규칙-only 소계 기준 150,000 코드 단위 — 배포 상시 규칙을 재고, 최상위 `AGENTS.md` 는 상수 축에서 제외해 별도 always-loaded-headroom(t1450) 계정에서 추적한다. 이것이 원 이슈 #1717/#1746 의 목표이자 런타임 1M 모델 한도다. REQ-ALB-002 를 제자리 재작성(같은 REQ id, 같은 Event-driven 분류, 재번호 없음 — 테스트는 모든 구성원을 계속 로그하고 가드 대상 수치와 실패 메시지는 규칙 축을 이름 댐), REQ-ALB-005 는 구조 그대로 도출만 교체, AC-ALB-002·003·006·007 초록 조건을 새 축으로 갱신, design §7 축 문장 갱신. M1 RED 증거(구 상수 115,000, 전면 합계 180,901)는 역사적 기록으로 남긴다. REQ·AC 25/25 유지. | manager-spec |
| 2026-10-08 | 0.8.0 | run M0 중간 수정(재계획 아님 — 진행 중인 run 이 측정한 사실이 한 요구사항의 전달 경로를 무효화했다). Q4 실측(`decision-index.md` Q4): SessionStart `additionalContext` 는 문자열당 10,000자로 제한되고 올릴 방법이 없으며, 한도를 넘는 출력은 런타임이 세션 디렉터리 파일로 저장해 경로 + 2,000자 미리보기를 전달한다(CC 2.1.89 문서화 동작). M0 원장은 `factory-dispatch.md` 역할 core 가 36행 전부 `binding`, 17,341 UTF-16 임을 재었다(`internal/template/testdata/binding_ledger.json`) — 기존 팩토리 맥락 생산자 `factoryHookContextLimit`(2,048)까지 합치면 같은 출력 안에 담으려면 축소율 약 0.46(STE 표본 0.746 대비)이 필요하고, REQ-ALB-025 가 역할 core 집합 축소를 금하므로 구식 REQ-ALB-010 의 2단 축소 사다리는 이 core 에 성립하지 않는다. 레인 결정(`progress.md` §F, 커밋 `4f43440b9`): 의도적 오버플로 파일 전달 — 훅은 역할 core 를 단 하나도 자르지 않고 그대로 내보내고 운영자 경고 + 에이전트용 안내 지시를 함께 내며, 런타임의 문서화된 초과 출력 동작이 전달 통로다. 오버플로 전달이 불가하거나 잘릴 때는 REQ-ALB-009 의 Read 지시 + 경고로 후퇴한다. REQ-ALB-010 을 제자리에서 재작성(같은 REQ id, 같은 Event-driven 분류, 재번호 없음), AC-ALB-012 초록 조건을 그에 맞춰 재작성, design §4 「실패와 크기」 크기 항을 새 경로로 갱신. REQ-ALB-011·AC-ALB-013 은 변동 없음 — 역할 가드는 훅 출력 수준에서 단정한다. REQ·AC 25/25 유지. | manager-spec |
| 2026-10-04 | 0.7.0 | plan-audit 5회차 FAIL 0.825(`.moai/reports/t1469/plan-audit-iter5.md`) 뒤 운영자 처분 Q10(a) 범위 축소(`decision-index.md` Q11). 원장 대상 파일을 상시 표면 파일로 되돌리고(기존 최상위 `paths:` companion 제외) N4-2 가 만든 companion 예외 행 규정을 삭제해 위치 어휘가 단일 규칙(`companion:` 은 `rationale` 행만)으로 회귀 — REQ-ALB-012·015·017 과 AC-ALB-021(1) 이 다시 일관된다. `factory-dispatch-detail.md` 분할을 후속 카드로 이동: REQ-ALB-003 을 상시 표면 구성원으로 축소하고 AC-ALB-004 초록을 빈 앵커 목록+변이 픽스처로 재작성, plan M3 무손실 조건 제거, design §6 정리. 제거 범위 기록은 `progress.md` §E.1, 목표 대조 재측정은 `research.md` §1.3. REQ·AC 25/25 유지. | manager-spec |
| 2026-10-04 | 0.6.0 | plan-audit 4회차 결함 수리(N4-1~N4-7, `.moai/reports/t1469/plan-audit-iter4.md`). N4-1: REQ-ALB-015 에 `always:` 표면 구성원 검사와 `companion:` 최상위 `paths:` 검사를 실패 조건으로 추가, AC-ALB-019 에 원본 잔류 변이 픽스처. N4-2: 원장 대상 파일을 「편집하는 배포 규칙 파일 전부(기존 `paths:` companion 포함)」로 넓히고, 원래 companion 에 있던 `binding`·`normative` 단위의 `companion:` 예외 행 규정. N4-3: AC-ALB-019 초록 조건을 「원장 대상 파일 전부의 모든 단위」로. N4-4: AC-ALB-015 RED-now 를 `/usr/bin/grep` 과 실제 출력 순서로. N4-5~7(옵션) 수리. 같은 착지에서 로컬 develop 흡수(`30ce3a02d`, t1399 개명·훅 제거)로 움직인 배포 파일명·진입점·측정치를 트리 `2771626b5` 기준으로 재고정(`research.md` §1.2). REQ·AC 25/25 유지. | manager-spec |
| 2026-10-03 | 0.5.0 | plan-audit 3회차 FAIL 0.80(`.moai/reports/t1469/plan-audit-iter3.md`) 뒤 리더가 델타 1회(iter4)를 허용(`decision-index.md` Q9). D3-1: RED-now 명령을 전체 경로로, 출력 축자. D3-2: 원장 범위를 「원장 대상 파일」(이 SPEC 이 편집하는 상시 파일 전부)의 모든 절로 넓힘. D3-3: 앵커 종류 고정과 종류 불일치·`companion:` 대조 실패 조건(REQ-ALB-015). D3-4: AC-ALB-021 면제와 조각 단위 정의. D3-5 진입점 식별자, D3-7 펜스 불투명 규칙, D3-8·D3-9 정정. REQ·AC 25/25 유지. | manager-spec |
| 2026-10-03 | 0.4.0 | plan-audit 2회차 FAIL 0.81(`.moai/reports/t1469/plan-audit-iter2.md`) 수리. N1: RB AC 의 RED-now 칸을 명령·stdout·exit 로 채우고 테스트 0개 실행 출력은 RED 근거에서 뺌. N2: 원장 행에 필수 `entry_points` 필드. N3·N6: 단위 경계를 §B 한 곳에서 정의하고, 구속 절 안의 토큰 없는 문단을 `normative`/`rationale` 행으로 원장에 넣음. N4: AC 기대값을 원장 앵커 기록에서 도출, M1 RED 는 M0 의 첫 흡수 뒤에 착지. N5: REQ-ALB-010 상한 대상을 최종 `additionalContext` 합본으로. N7·N9·N10 반영. REQ·AC 수는 25/25 그대로. | manager-spec |
| 2026-10-03 | 0.3.0 | plan-audit 1회차 FAIL 0.69(`.moai/reports/t1469/plan-audit-iter1.md`) 수리. D1: 원장 단위를 줄에서 구속 블록(연속 줄·하위 항목 포함)으로. D2: 런타임 분류 원천을 배포 규칙 안의 중립 영역 표지로(REQ-ALB-023), 원장은 테스트 고정물. D3: 무표지 리더급 진입점(manager-lead·deputy, moai-kanban-foreman, `--auto`)에 읽기 지시 표지와 가드(REQ-ALB-024), 분류 규칙(REQ-ALB-025). D4: 위치 어휘 `always:`/`role-core:`, role-core 행은 생성 함수 출력으로 대조. D5: 원장 앵커와 재고정 절차(`plan.md` §C.2). D6: 고정 렌더 입력과 렌더 독립 소계. D7: Q5 리더 결정 기록, REQ-ALB-014 예외 확정. D8·D9: AC 표에 명령·출력·exit·분류, 초록은 이름 붙은 PASS 줄. D10: REQ-ALB-010 을 When 으로, 초과 시 동작 명시. D11: 미러 등록(REQ-ALB-018). D12: 표지 레지스트리 도출(REQ-ALB-011). | manager-spec |
| 2026-10-03 | 0.2.0 | 운영자 판정 반영(리더 세션 `df44e022` 의 AskUserQuestion 답, `decision-index.md` Q1·Q2·Q6·Q7). Q1: 이 SPEC 에 한해 HEADROOM-001 의 구속 조항 동결을 해제 — 역할 주입과, 의미를 보존하는 압축 재작성 둘 다 허용. 조건은 감사자가 의미 보존을 확인할 수 있는 구속 원장(구속 줄마다 전/후)과 구속 의무 무손실. REQ-ALB-012·015·022 와 §D·§F·§G 를 그에 맞춰 고침. Q7: v3.2.0 에 넣지 않으며 v3.2.0 컷 뒤 develop 에 착지 — `phase` 변경. | manager-spec |
| 2026-10-03 | 0.1.0 | 최초 작성. 카드 t1469(GitHub #1717 + #1746, 클래스 C). 기준 트리 로컬 `develop` `b5815ca80`. 리드 실측(새 `moai init`, UTF-16 코드 단위) 191,386자를 이 plan 실행에서 템플릿 원본으로 교차 확인했다(`research.md` §1). 겹치는 운영자 카드 t1450 의 흡수·잔류 판정은 `plan.md` §B, 착지 순서 제약(t1453 M4 보다 먼저)은 `plan.md` §C 에 둔다. | manager-spec |

---

## §A. 배경

새로 `moai init` 한 프로젝트는 세션을 열 때마다 조건 없는 규칙 13개와 `CLAUDE.md`, 그리고 그 `@`-import를 합쳐 무조건 싣는다. 이 plan이 서는 트리(로컬 develop 흡수 `30ce3a02d` 뒤 `2771626b5`)에서는 규칙 13개(154,652자) + `CLAUDE.md`(15,541자) + `@`-import 원본 합(23,541자 — `AGENTS.md.tmpl` 21,860·`user.yaml` 224·`language.yaml` 1,457) = **193,734자**(UTF-16, 템플릿 원본 기준)다. plan 기준선 `b5815ca80` 에서 잰 191,386자는 이력이다(`research.md` §1.2). 카드가 인용한 Claude Code 한도는 1M 모델 150,000자, 200K 모델 120,000자, 파일당 40,000자다. 사용자는 세션마다 한도 초과 경고를 본다.

크기 상위는 `workflow/factory-dispatch.md`(구 `kanban-dispatch.md`) 27,423 · `core/agent-common-protocol.md` 18,421 · `core/askuser-protocol.md` 17,872 · `workflow/session-handoff.md` 15,833 · `core/verification-claim-integrity.md` 15,766 · `core/moai-constitution.md` 15,510 · `workflow/cross-session-messaging.md` 11,321 이다. `paths:` 로 한정된 `workflow/factory-dispatch-detail.md` 는 40,659자(§1 시점 43,138 — t1399 분할로 줄었으나 아직 초과)로 파일당 한도를 넘고, `workflow/spec-workflow.md` 도 40,052자로 초과에 들어왔다(`research.md` §1.2). 두 초과 파일의 처분은 이 SPEC 밖이다 — `factory-dispatch-detail.md` 분할은 Q10(a) 범위 축소로 후속 카드로 옮겨졌고(`decision-index.md` Q11), `spec-workflow.md` 는 리드가 별도로 처리한다.

기존 가드 두 개는 이 표면을 보지 못한다.

- `internal/config/token_budget_guard.go` 의 `AlwaysLoadedTokenBudget`(77,600 토큰, 바이트÷4 추정)은 **이 저장소의 도그푸드 규칙**을 잴 뿐, 사용자가 받는 배포 표면을 재지 않는다.
- `internal/hook/instructions_loaded.go` 의 `sessionCharBudget`(210,000)은 운영자 판정값이며(`SPEC-INSTRUCTIONS-BUDGET-001` HISTORY), 런타임 한도보다 높아 191k 표면에서 아무 신호도 내지 않는다. 이 상수를 바꾸려면 결정 기록이 필요하므로 이 SPEC 은 건드리지 않는다(§F).

선행 작업의 결론도 범위를 정한다. `SPEC-ALWAYS-LOADED-HEADROOM-001` 은 구속 조항 동결(`SPEC-ALWAYS-LOADED-DIET-002` REQ-ALD2-002·003 — 구속 줄을 companion·skill 로 옮기지도, 다시 쓰지도 않는다) 아래에서는 150,000자 달성이 구조적으로 불가능하다고 판정했고(`STRUCTURALLY-INFEASIBLE-UNDER-FREEZE`), 동결 해제 여부를 운영자 게이트로 남겼다. 이 SPEC 의 역할 한정 주입은 구속 줄을 상시 로드 표면 **밖**(역할 주입 경로)으로 옮기므로, 그 게이트와 정면으로 만난다. 운영자는 이 SPEC 에 한해 동결을 해제했다(2026-10-03, `decision-index.md` Q1) — 역할 주입과 의미 보존 압축 재작성을 모두 허용하되, 구속 원장으로 의미 보존을 감사할 수 있어야 하고 구속 의무는 하나도 잃지 않는다.

## §B. 용어

| 용어 | 뜻 |
|---|---|
| 배포 표면 | 내장 템플릿을 빈 디렉터리에 배포·렌더링해 얻은 트리에서, 세션 시작 때 무조건 실리는 파일 집합 — 최상위 `paths:` 키가 없는 `.claude/rules/**/*.md` 전부, 렌더링된 `CLAUDE.md`, 그리고 그 트리 안에서 해석되는 `CLAUDE.md` 의 `@`-import 전이 폐포 |
| 고정 렌더 입력 | 배포 표면을 잴 때 쓰는 렌더 입력: `UserName` 은 빈 문자열, `ConversationLanguage` 는 `en`. 그 밖의 입력은 `moai init` 기본값 |
| 계수 단위 | UTF-16 코드 단위(JavaScript 문자열 길이와 같은 단위). 한글·영문 모두 1, 보충 평면 문자는 2 |
| 역할 세션 | 런처가 칸반·팩토리의 리더 또는 레인으로 표시해 띄운 세션. 표시는 런처와 훅이 함께 읽는 단일 역할 표지 레지스트리로 판정한다 |
| 무표지 리더급 진입점 | 역할 표지 없이 리더급 의무를 질 수 있는 진입점 — `manager-lead` 에이전트(보좌 deputy 포함), `moai-factory-foreman` 스킬(구 `moai-kanban-foreman`, t1399 개명), `/moai gtd`·`/moai:todo` 의 `--auto` 경로 |
| 역할 한정 규칙 | 이 SPEC 의 대상인 `factory-dispatch.md`(구 `kanban-dispatch.md`, t1399 개명)와 `cross-session-messaging.md` |
| 역할 core | 역할 한정 규칙 본문 안에서 중립 영역 표지(`<!-- moai:role-core-start -->` … `<!-- moai:role-core-end -->`)로 둘러싼 구속 블록과 그 블록이 기대는 제목·표. 훅의 역할 core 생성 함수가 배포 트리의 파일만 읽어 만든다 |
| 상시 stub | 역할 한정 규칙마다 상시 표면에 남는 짧은 파일 — 역할 core 로 분류되지 않은 구속 블록 전부와 역할 경로 포인터 |
| 구속 토큰 | `[HARD]`, `MUST`, `MUST NOT`, `shall ` |
| 단위 경계 | 원장 단위는 원장 대상 파일 안의 문단이다. 한 단위는 빈 줄이 아닌 줄로 시작해, 다음 제목(수준 무관) 직전이나, 빈 줄 뒤에 이어지는 줄이 그 단위의 목록 항목·하위 항목·표 행·코드 펜스가 아닌 지점에서 끝난다. 그 단위가 바로 이어서 여는 목록·표·코드 펜스는 빈 줄 하나까지 사이에 두고 그 단위에 속한다. 코드 펜스 내부는 불투명하다 — 펜스 안의 `#` 줄, 빈 줄, 구속 토큰은 단위를 열거나 끊지 않고 제목으로도 보지 않는다. 이 정의가 유일한 정의이며 `design.md` 는 이를 인용만 한다 |
| 구속 블록 | 구속 토큰을 포함한 단위. 연속 줄과 하위 항목은 구속 토큰이 없어도 블록에 속한다 |
| 규범 문단 | 원장 대상 파일에서 구속 토큰이 없는 단위 가운데, 의무·금지·행동 지시를 담은 것(예: 앞 금지를 해석하거나 따르는 경로를 지시하는 문단, `goal-directive.md` 의 「Arming a goal does not authorize autonomous run-phase entry」 문단). M0 에서 원장 대상 파일의 토큰 없는 단위는 모두 `normative` 또는 `rationale` 로 분류되고, `rationale` 판정에는 이유 한 줄이 붙는다 |
| 원장 대상 파일 | 앵커 트리에서 이 SPEC 이 편집하는(분할·역할 한정·stub 생성·포인터화) 배포 상시 표면 파일 전부 — 최소 집합은 상시 규칙 13개와 `CLAUDE.md` 템플릿이며, M0 후보표에서 편집하지 않기로 한 파일만 원장 머리에 이유와 함께 제외를 적는다. 구속 토큰 유무와 무관하게 그 파일의 모든 절·모든 단위가 원장 대상이다. 최상위 `paths:` 를 가진 기존 companion 파일(`factory-dispatch-detail.md` 등)은 원장 대상이 아니다 — 그 분할은 Q10(a) 범위 축소(2026-10-04, `decision-index.md` Q11)로 후속 카드로 옮겨졌고, companion 은 `rationale` 이관의 대상이 될 뿐(REQ-ALB-017) 원장 행을 갖지 않는다 |
| 구속 원장 | 원장 대상 파일의 단위마다 한 행 — 블록 ID, 종류(`binding` / `normative` / `rationale`), 앵커 종류(M0 기준선 커밋에서 고정, `plan.md` §C.2 재고정 커밋에서만 바뀜), 출처(파일·제목), 변경 전 텍스트, 변경 후 텍스트, 변경 후 위치, 처리 방식(`verbatim` / `rewrite` / `relocated` — `relocated` 는 `rationale` 행만), 재작성 메모, `entry_points`(종류가 `binding`·`normative` 인 행은 필수: 그 단위가 묶을 수 있는 진입점마다 `{entry_point, delivery}`. `delivery` 는 `always` / `REQ-ALB-007` / `REQ-ALB-024` 이고, `entry_point` 식별자는 `REQ-ALB-007` 이면 역할 표지 레지스트리의 표지 이름, `REQ-ALB-024` 이면 표지를 단 배포 파일 경로, `always` 이면 `any-session`). 머리에 앵커 SHA 와 그 앵커에서 잰 기대값(렌더 독립 소계, 구성원 목록, 파일당 초과 파일)을 둔다. 테스트 고정물로 `internal/template/testdata/` 아래에 커밋되며 배포되지 않는다 |
| 위치 어휘 | 원장의 변경 후 위치는 `always:<배포 경로>`, `role-core:<역할 한정 규칙 경로>`, `companion:<배포 경로>` 셋 가운데 하나다. `always:` 의 경로는 예산 테스트가 도출한 배포 상시 표면 구성원(REQ-ALB-004)이어야 하고, `companion:` 의 경로는 최상위 `paths:` 를 가진 배포 파일이어야 한다. `companion:` 은 종류가 `rationale` 인 행에만 허용된다. 그 밖의 값은 원장 오류다 |
| 원장 앵커 | 원장 기준선을 뽑은 커밋. 원장 머리에 SHA 와 그 트리에서 잰 예산 테스트 기대값으로 기록한다. 재고정 절차는 `plan.md` §C.2 |

---

## §C. 요구사항 (GEARS)

### C.1 배포 표면 예산 가드

- **REQ-ALB-001** (Ubiquitous) — The template test suite shall carry a test named `TestDeployedAlwaysLoadedCharBudget` that deploys the embedded template set into a `t.TempDir()` project through the production deploy-and-render path with the fixed render inputs (§B), and sums the UTF-16 code-unit length of every file in the deployed surface.
- **REQ-ALB-002** (Event-driven) — When the rules-only subtotal of the deployed always-loaded surface exceeds the budget constant of 150,000 code units, the budget test shall fail with a message naming the rules-only subtotal and the axis it is measured over, the budget, the file count, and the five largest members with their sizes; the test shall also log one `deployed-surface-total=<N>` line and one `deployed-surface-member=<path> <size>` line per member — the root `AGENTS.md` instruction file included — on every run, while the figure the constant is guarded against is the rules-only subtotal alone. (Axis re-derived 2026-10-08 by operator disposition (a) — the disposal REQ-ALB-022's stop promised, `decision-index.md` Q6: the root `AGENTS.md`, 23,684 at the recorded measurement, sits on the measured surface but outside the constant's axis and is tracked on the separate always-loaded-headroom account, card t1450.)
- **REQ-ALB-003** (Event-driven) — When any deployed always-loaded rule file — a `.claude/rules/**/*.md` file carrying no top-level `paths:` key — exceeds 40,000 code units, the budget test shall fail naming that file and its size. (Narrowed 2026-10-04 by Q10(a) scope reduction: `paths:`-scoped companions left this SPEC's scope with the `factory-dispatch-detail.md` split moving to a follow-up card — `decision-index.md` Q11 — so a companion over 40,000 no longer fails this test; the two current companion breaches, `factory-dispatch-detail.md` 40,659 and `spec-workflow.md` 40,052, are disposed by the follow-up card and the leader's separate `spec-workflow.md` handling.)
- **REQ-ALB-004** (Ubiquitous) — The budget test shall derive the deployed surface mechanically from the deployed tree (frontmatter `paths:` inspection plus `@`-import closure), so that adding an unconditional rule or a new import grows the measured total, and adding a `paths:`-scoped rule leaves it unchanged.
- **REQ-ALB-005** (Ubiquitous) — The budget constant shall carry, next to its declaration, the measurement it was derived from: the runtime limit it sits under (the 150,000-char 1M-model limit, applied to the rules-only axis — the original issues' #1717/#1746 goal) and the operator disposition that re-derived it (disposition (a) of 2026-10-08, executing the REQ-ALB-022 stop — the measured compression floor 122,000–126,000 on the superseded full-surface constant 115,000; headroom 35,406 at the recorded rules-only 114,594).

### C.2 역할 한정 주입

- **REQ-ALB-006** (Ubiquitous) — The deployed always-loaded surface shall carry `factory-dispatch.md` and `cross-session-messaging.md` only as always-on stubs; each stub shall hold every binding block of its rule that is not classified role-core, and shall name the role-injection path.
- **REQ-ALB-007** (Event-driven) — When a SessionStart event with source `startup`, `clear`, or `compact` reaches a session that the role-marker registry identifies as a kanban or factory leader or lane, the SessionStart hook shall inject the role core of both role-gated rules into the session's additional context.
- **REQ-ALB-008** (Unwanted) — The SessionStart hook shall not inject a role core into a session that the role-marker registry does not identify as a leader or lane, nor on source `resume`.
- **REQ-ALB-009** (Event-driven) — When the hook cannot produce a role core for a role session (source file absent, unreadable, empty, or carrying no region markers), the hook shall emit an operator-visible warning and an agent-facing directive to read the full rule files by path, and shall not end the session start silently without either.
- **REQ-ALB-010** (Event-driven) — When the final additional context the SessionStart hook would emit — every existing producer's text (session attribution, kanban or factory notices, lane rules) plus the role core plus any directive — exceeds the size the runtime delivers intact (measured in M0 at 10,000 characters per `additionalContext` string, unraisable; `decision-index.md` Q4), the hook shall emit the role core intact — never truncating a unit — together with an operator-visible warning and an agent-facing directive stating that the runtime saves the intact output to a session-directory file and passes the file path plus a preview of its first 2,000 characters; when that overflow delivery is unavailable or would truncate, the hook shall emit the REQ-ALB-009 read-directive with the operator-visible warning instead.
- **REQ-ALB-011** (Ubiquitous) — The hook test suite shall carry a role-coverage guard that, for every leader and lane marker in the role-marker registry, asserts the injected context contains every role-core block, with the expected block set derived from the deployed rule files and the marker set derived from the registry rather than from a hand-written list.
- **REQ-ALB-023** (Ubiquitous) — The SessionStart hook shall build the role core solely from the deployed role-gated rule files, selecting the regions enclosed by the neutral region markers; the SPEC's binding ledger shall not be read at runtime.
- **REQ-ALB-024** (Ubiquitous) — Every deployed agent definition, skill, or workflow file that cites a role-gated rule shall carry the neutral marker `moai:role-rules-required` together with a binding directive to read both role-gated rule files in full before its first action, and a template test shall enumerate those files from the deployed tree and fail on any citing file without the marker.
- **REQ-ALB-025** (Ubiquitous) — A binding block or normative paragraph shall be classified role-core only when every entry point it can bind is covered by REQ-ALB-007 or REQ-ALB-024; otherwise it shall stay in the always-on stub.

### C.3 대형 규칙 분할과 구속 의무 보존

- **REQ-ALB-012** (Ubiquitous) — Every binding block and every normative paragraph of the ledger-scope files at the ledger anchor shall, after the change, be present in either the deployed always-loaded surface or a role core, either verbatim or as a meaning-preserving compressed rewrite recorded in the binding ledger; no binding obligation shall be dropped, and only units classified `rationale` may move to a companion.
- **REQ-ALB-013** (Unwanted) — The change shall not relocate any part of a binding block — token line, continuation line, or subordinate item — into a `paths:`-scoped companion, a skill, or any other on-demand surface.
- **REQ-ALB-014** (Unwanted) — The change shall not add a top-level `paths:` key to any rule whose trigger is a behavior rather than a file path, except the two role-gated rules, whose top-level `paths:` is a non-delivery placement while delivery is the injection of REQ-ALB-007 and REQ-ALB-024 (`decision-index.md` Q5).
- **REQ-ALB-015** (Ubiquitous) — The binding ledger shall carry one row per unit of every ledger-scope file at the ledger anchor, and a ledger test shall fail when a base-tree unit has no row, when a row's kind differs from its anchor kind, when a `binding` or `normative` row has a missing or empty `entry_points` field, when a `role-core:` row lists an entry point whose delivery is neither REQ-ALB-007 nor REQ-ALB-024, when a `REQ-ALB-007` entry point is not a marker in the role-marker registry or a `REQ-ALB-024` entry point is not a deployed file carrying the `moai:role-rules-required` marker, when a row's location falls outside the location vocabulary or a `binding` or `normative` row carries a `companion:` location, when an `always:` row's path is absent from the deployed-surface member list the budget test derives (REQ-ALB-004), when a `companion:` row's path is not a deployed file carrying a top-level `paths:` key, when an `always:` or `companion:` row's after-text is absent from that deployed file, or when a `role-core:` row's after-text is absent from the output of the hook's role-core builder run on the deployed tree.
- **REQ-ALB-016** (Ubiquitous) — Every `<file>.md § <heading>` cross-reference inside the deployed rules tree shall resolve to an existing heading after the split.
- **REQ-ALB-017** (Event-driven) — When a rule body is split, the companion shall receive only rationale, examples, incident records, and detail tables, and the always-loaded core shall keep a one-line pointer per relocated section.

### C.4 Template-First 와 중립성

- **REQ-ALB-018** (Ubiquitous) — Every content change shall originate under `internal/template/templates/`, be embedded through `make build`, and be mirrored to this repository's `.claude/rules/moai/` copy within the same SPEC run, with every changed or new rule pair registered in the rule-template mirror test.
- **REQ-ALB-019** (Unwanted) — The template text added or changed by this SPEC shall not carry SPEC IDs, card ids, internal dates, commit hashes, or text that favors one of the 16 supported programming languages.
- **REQ-ALB-020** (Ubiquitous) — Each role-gated rule's loading-scope note shall state the role-injection delivery path instead of declaring the file intentionally always-loaded.

### C.5 측정 기록

- **REQ-ALB-021** (Ubiquitous) — The run phase shall record in `progress.md` the deployed-surface total before and after the change, measured by the budget test itself, together with the first-turn context measurement of a fresh session before and after.
- **REQ-ALB-022** (Event-driven) — When the feasibility measurement of milestone M0 finds that the lowest total reachable under REQ-ALB-012 and REQ-ALB-013 — with role injection and meaning-preserving compressed rewriting both applied — exceeds the budget constant, the run phase shall stop before any rule edit and report the measured floor to the leader.

---

## §D. 제약

- 계수 단위는 UTF-16 코드 단위다. 기존 `instructions_loaded.go` 의 룬 계수와는 단위가 다르므로, 두 지표를 같은 수로 비교하지 않는다(한글 본문은 두 단위가 같고, 보충 평면 문자에서만 갈린다).
- 구속 블록과 규범 문단은 축자 이동하거나, 의미를 보존하는 압축 재작성만 할 수 있다. 재작성은 구속 원장의 전/후 행으로만 하며, 의무를 떨어뜨리는 재작성·병합·삭제는 허용하지 않는다(REQ-ALB-012·015, 운영자 판정 Q1).
- 역할 한정 규칙의 전체 본문은 기존 경로(`workflow/factory-dispatch.md`, `workflow/cross-session-messaging.md`)에 남긴다. 그 경로를 고정해 둔 테스트·스킬·에이전트 포인터가 많아서다(`research.md` §4).
- SessionStart 주입은 fail-open 이 아니라 fail-visible 이다. 역할 세션이 규칙 없이 조용히 시작하는 경로가 하나라도 남으면 안 된다(REQ-ALB-009).
- Codex 하네스는 `.claude/rules/` 를 읽지 않으므로 이 SPEC 의 영향 밖이다. `AGENTS.md` 의 자족성은 바뀌지 않는다.

## §E. 수용 기준

인수 조건 전체는 `acceptance.md` 에 있다(AC-ALB-001 ~ AC-ALB-025).

---

## §F. 범위 밖

### Out of Scope — 운영자 결정이 걸린 상수

- `internal/hook/instructions_loaded.go` 의 `sessionCharBudget`(210,000) 변경. 운영자 판정값이며 바꾸려면 결정 기록이 먼저 있어야 한다(`decision-index.md` Q3).
- `internal/config/token_budget_guard.go` 의 `AlwaysLoadedTokenBudget` 하향 조정. 도그푸드 표면을 재는 별개 가드이며, 이 SPEC 이 착지하면 여유가 커질 뿐 깨지지 않는다.

### Out of Scope — 카드 t1450 에 남는 항목

- 출력 스타일 본문 축약(t1450 (b)).
- 스킬 목록 예산(`skillListingBudgetFraction`)과 에이전트 설명 상한(t1450 (c)).
- `AGENTS.md` 본문 편집과 Codex 자족성 재설계(t1450 (d) 중 `AGENTS.md` 쪽). 규칙 쪽 중복 제거는 이 SPEC 이 맡는다(`plan.md` §B).

### Out of Scope — 문구와 체계

- 구속 의무의 삭제, 그리고 원장 행 없이 이루어지는 재작성·병합.
- v3.2.0 릴리스 배치. 이 SPEC 은 v3.2.0 컷 뒤 develop 에 착지한다(운영자 판정 Q7).
- `MEMORY.md` 등 사용자 소유 자동 메모리의 크기.
- github-flow 전환에 따른 develop 서술 갱신(카드 t1453 M4). 이 SPEC 이 먼저 착지하고 t1453 이 새 배치 위에서 문구를 고친다.

### Out of Scope — Q10(a) 범위 축소(2026-10-04)

- `workflow/factory-dispatch-detail.md`(40,659, 최상위 `paths:`)의 40,000 미만 분할. 후속 카드로 옮겨졌다(`decision-index.md` Q11) — 무손실 분할 검사와 companion 출처 앵커 원천 술어 설계를 포함한 기록은 `progress.md` §E.1.
- `workflow/spec-workflow.md`(40,052, 최상위 `paths:`)의 초과 처분. 리드가 별도로 처리하며 이 SPEC 이 건드리지 않는다.

---

## §G. 관련 관계

- `SPEC-ALWAYS-LOADED-DIET-002` REQ-ALD2-002·003 은 **그 SPEC 의 구현**에 걸린 동결이다. 이 SPEC 은 구속 블록의 어느 부분도 on-demand 표면으로 옮기지 않는다는 점(REQ-ALB-013)은 승계한다. 상시 표면에서 역할 주입 경로로 옮기는 것과 의미 보존 압축 재작성은 운영자가 이 SPEC 에 한해 허용했다(`decision-index.md` Q1, 운영자 답 2026-10-03). 허용 조건은 구속 원장과 의무 무손실이다.
- `SPEC-INSTRUCTION-BUDGET-SCOPE-001` 의 40,000자 파일당 원칙을 배포 상시 표면의 규칙 파일로 넓혀 기계 검사한다(REQ-ALB-003 — Q10(a) 축소 뒤 대상은 최상위 `paths:` 가 없는 규칙뿐이다).

## §H. 교차 참조

- 카드 t1469 · GitHub #1717 · #1746
- 카드 t1450(겹침, 관계 `conflicts` 기록됨) · t1453(착지 순서) · t1399(런처 역할 표면)
- `research.md` — 실측 원문과 명령
- `design.md` — 역할 주입 경로와 분할 배치
- `decision-index.md` — 결정 11건(해소 8: Q1·Q2·Q5·Q6·Q7·Q9·Q10·Q11, 범위 밖 1: Q3, 실측 대기 2: Q4·Q8)
