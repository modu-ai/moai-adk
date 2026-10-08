---
id: SPEC-ALWAYS-LOADED-BUDGET-001
title: "design — 역할 주입 경로와 분할 배치"
version: "0.7.0"
created: 2026-10-03
updated: 2026-10-04
---

# design.md — SPEC-ALWAYS-LOADED-BUDGET-001

## §1. 세 층

| 층 | 언제 실리는가 | 무엇을 담는가 |
|---|---|---|
| 상시 core | 모든 세션, 세션 시작 시 | 역할 core 로 분류되지 않은 구속 블록 전부, 그 블록이 기대는 표·제목, 절마다 companion 포인터 한 줄 |
| 역할 core | 역할 세션: SessionStart(`startup`·`clear`·`compact`) 주입. 무표지 리더급 진입점: 정의 본문의 읽기 지시 | 역할 한정 규칙에서 영역 표지로 둘러싼 구속 블록과 그 블록이 기대는 제목·표 |
| companion | 경로 접촉 시(`paths:`) 또는 명시적 Read | 근거, 사례, 사건 기록, 상세 표 — 구속 블록의 어떤 부분도 없음 |

구속 블록은 상시 core 와 역할 core 에만 산다(REQ-ALB-012·013). 운영자 판정(Q1)에 따라 블록은 의미를 보존하는 범위에서 압축해 다시 쓸 수 있고, 그때마다 원장에 전/후를 남긴다.

## §2. 구속 블록과 분류

### 단위 경계

경계는 `spec.md` §B 「단위 경계」가 유일한 정의다. 여기서는 그 정의를 실제 본문에 적용한 결과만 적는다.

- `factory-dispatch.md`(구 `kanban-dispatch.md`) 「Dispatch format」: `[HARD]` 문단, 코드 펜스, 토큰 없는 `No explanatory prose`·`Ceiling: the block is at most 10 lines` 항목이 한 블록이다(펜스와 목록이 그 단위가 바로 연 하위 요소이므로).
- `cross-session-messaging.md` 「Rules」 절: `[HARD]` 문단 뒤 빈 줄을 두고 오는 「A `STOPPED_TEAMMATE_VIOLATION` deny is never a bug to route around …」 문단은 토큰이 없어 별도 단위가 된다. 원장 대상 파일 안에 있으므로 원장 행을 갖고, 행동 지시를 담으므로 `normative` 다. 따라서 companion 으로 갈 수 없다(REQ-ALB-012).
- `goal-directive.md` 는 구속 토큰 줄이 하나도 없지만 원장 대상 파일이다. 「Arming a goal does not authorize autonomous run-phase entry」·「Safety boundary unchanged」 문단은 `normative` 행이 되어 companion 으로 갈 수 없다.
- 원장 대상 파일 밖의 파일 — 최상위 `paths:` 를 가진 companion 을 포함해 — 는 원장 행을 갖지 않는다. 이 SPEC 이 companion 에 추가하는 내용은 `rationale` 뿐이므로(REQ-ALB-017) 구속 단위 기준선은 흔들리지 않는다.
- 앵커 종류는 M0 기준선 커밋에서 고정된다. `normative` 행을 `rationale` 로 바꾸려면 §C.2 재고정 커밋에서 이유와 함께 바꿔야 하고, 그 밖의 커밋에서 바꾸면 원장 테스트가 종류 불일치로 실패한다.

### 진입점 식별자

`entry_point` 는 `REQ-ALB-007` 이면 레지스트리의 표지 이름, `REQ-ALB-024` 이면 표지를 단 배포 파일 경로(예 `.claude/agents/moai/manager-lead.md`), `always` 이면 `any-session` 이다. 원장 테스트가 각 식별자를 레지스트리와 배포 트리에서 해석한다.

### 분류 규칙 (REQ-ALB-025)

`binding`·`normative` 행마다 그 단위가 묶을 수 있는 진입점을 `entry_points` 에 적고, 진입점마다 전달 근거(`always` / `REQ-ALB-007` / `REQ-ALB-024`)를 단다. 그 진입점 전부가 SessionStart 주입(REQ-ALB-007) 또는 읽기 지시 표지(REQ-ALB-024)로 덮일 때만 역할 core 다. 하나라도 덮이지 않으면 일반 블록으로 상시 stub 에 남는다. 판정이 갈리면 일반으로 둔다 — 상시 표면에 남겨서 잃는 것은 자수뿐이다.

## §3. 역할 한정 규칙의 배치

1. **전체 본문은 제자리에 남는다.** 경로를 고정한 소비자가 많으므로(`research.md` §4) 파일을 옮기지 않는다. 최상위 `paths:` 는 상시 표면에서 빼는 비전달 배치일 뿐이고, 전달은 §4 가 맡는다. 리더는 이것을 「`paths:` 만 붙이기」가 아니라고 결정했다(`decision-index.md` Q5).
2. **역할 core 영역은 본문 안 중립 표지로 표시한다.** `<!-- moai:role-core-start -->` … `<!-- moai:role-core-end -->`. 기존 `<!-- moai:evolvable-start -->` 표지와 같은 계열의 HTML 주석이라 렌더 결과에 보이지 않고, SPEC ID·카드 id 를 담지 않는다.
3. **상시 stub 이 새로 생긴다.** 일반 블록 전부와 역할 경로 포인터를 담는다.

## §4. 전달 경로

| 진입점 | 표지 | 전달 |
|---|---|---|
| 칸반·팩토리 리더·레인 세션 | 역할 표지 레지스트리의 환경 변수 | SessionStart 훅이 역할 core 주입 |
| `manager-lead` 에이전트(보좌 deputy 포함) | 없음 | 에이전트 정의의 `moai:role-rules-required` 표지 + 첫 행동 전 두 규칙 전체 Read 지시 |
| `moai-factory-foreman` 스킬(구 `moai-kanban-foreman`, t1399 개명) | 없음 | 스킬 본문 같은 표지·지시 |
| `gtd` 워크플로 `--auto`(`/moai:todo` 별칭 포함) | 없음 | 워크플로 본문 같은 표지·지시 |
| 위 목록에 없는 새 진입점 | — | 인용 파일 전수 가드(REQ-ALB-024)가 표지 누락을 잡고, 인용하지 않는 진입점은 분류 규칙이 블록을 일반으로 남긴다 |

에이전트·스킬 본문은 그 정의가 호출될 때 반드시 실리므로, 그 안의 읽기 지시는 환경 표지 없이도 결정적으로 도달한다. 전체 파일을 Read 하면 기존 companion 의 `paths:` 글롭(`**/factory-dispatch*.md`)이 걸려 companion 도 실린다.

### SessionStart 주입 시점

| source | 주입 | 이유 |
|---|---|---|
| `startup` | 한다 | 새 세션 |
| `clear` | 한다 | 대화가 비워져 이전 주입이 사라짐 |
| `compact` | 한다 | 요약이 이전 주입을 대체함(`research.md` §5) |
| `resume` | 하지 않는다 | 이전 전사가 그대로 복원됨 |

### 실패와 크기

- 역할 core 를 만들 수 없으면(파일 부재·읽기 실패·빈 본문·영역 표지 없음) 운영자용 `systemMessage` 경고와 에이전트용 Read 지시를 함께 낸다(REQ-ALB-009).
- 상한은 훅이 내보낼 최종 `additionalContext` 전체에 건다. 같은 출력에는 이미 세션 귀속문(`internal/hook/session_start.go` 의 귀속 줄), 칸반·팩토리 공지, 레인 규칙이 들어간다. 합본이 Q4 실측 한도를 넘으면 역할 core 를 그 `binding`·`normative` 단위와 Read 지시로 바꾼다. 그래도 넘으면 경고와 Read 지시만 낸다. 어떤 단위도 중간에서 자르지 않는다(REQ-ALB-010).

### 역할 표지 레지스트리

런처와 훅이 같은 표지 목록을 읽는다. 역할 가드는 이 레지스트리를 순회하므로, t1399 처럼 표지가 바뀌거나 늘어도 새 표지가 자동으로 검사 대상이 된다(REQ-ALB-011, 감사 D12).

## §5. 구속 원장

- 위치: `internal/template/testdata/` 아래 고정물. 배포되지 않으므로 SPEC ID·카드 id 를 담아도 템플릿 중립성과 충돌하지 않는다. 런타임은 읽지 않는다(REQ-ALB-023).
- 행: 블록 ID · 종류(`binding`/`normative`/`rationale`) · 출처(파일·제목) · 변경 전 텍스트 · 변경 후 텍스트 · 위치(`always:<배포 경로>` / `role-core:<역할 한정 규칙 경로>` / `companion:<배포 경로>` — 마지막은 `rationale` 행만) · 처리 방식(`verbatim`/`rewrite`/`relocated`) · 재작성 메모 · `entry_points`(`binding`·`normative` 필수) · (부록 행) `upstream-removed` 와 상류 SHA.
- 머리: 앵커 SHA, 그 앵커에서 잰 렌더 독립 소계와 구성원 목록(AC-ALB-002·003·004 의 기대값 원천).
- 대조: `always:` 행은 배포 트리의 해당 파일에서, `role-core:` 행은 배포 트리에 대해 훅의 역할 core 생성 함수를 돌린 출력에서 변경 후 텍스트를 찾는다. `paths:` 가 붙은 원본 파일에서 찾지 않는다(감사 D4). 위치 검사도 함께 한다 — `always:` 행의 경로는 예산 테스트가 도출한 배포 상시 표면 구성원 목록에 있어야 하고, `companion:` 행의 경로는 최상위 `paths:` 를 가진 배포 파일이어야 한다(REQ-ALB-015).
- 앵커와 재고정: `plan.md` §C.2.

## §6. 대형 규칙 분할

대상은 상시 core 로 남는 11개 규칙과 `CLAUDE.md` 템플릿이다. 기존 companion 이 있으면 그리로 옮기고, 수용량(40,000)이 모자라면 새 companion 을 만든다. 분할 순서는 M0 후보표가 정한다 — 비구속 본문이 많은 파일부터다.

Q10(a) 범위 축소(2026-10-04, `decision-index.md` Q11): 기존 `paths:` companion 의 분할은 이 SPEC 밖이다 — `factory-dispatch-detail.md`(40,659) 분할은 후속 카드로 옮겨졌고, `spec-workflow.md`(40,052) 는 리드가 별도로 처리한다. 이 SPEC 의 분할은 상시 표면 파일의 비구속 본문 이관뿐이다.

규칙 쪽에서 `AGENTS.md` 와 같은 내용을 되풀이하는 문단은, `AGENTS.md` 가 `CLAUDE.md` 의 `@`-import 로 늘 실린다는 사실에 기대어 포인터로 바꿀 수 있다. 그 문단이 구속 블록이면 의무가 `AGENTS.md` 에 같은 범위로 있다는 것을 원장 행에 적은 경우에만 바꾼다.

## §7. 예산 가드

`TestDeployedAlwaysLoadedCharBudget` 는 내장 템플릿을 고정 렌더 입력으로 `t.TempDir()` 에 배포·렌더링하고, 배포 트리에서 표면을 기계적으로 도출해 UTF-16 길이를 합한다. 매 실행마다 합계와 구성원별 크기를 로그로 남긴다. 렌더 입력에 따라 변하는 것은 `user.yaml`·`language.yaml`·`AGENTS.md` 뿐이므로, 사용자 표면과의 동치는 렌더 독립 소계(규칙 13개 + `CLAUDE.md`)와 구성원 목록으로 증명한다. 기대값은 고정 숫자가 아니라 원장 앵커에서 잰 값이다 — `b5815ca80` 에서 169,018, 로컬 develop `d7112d005` 에서 170,593 으로 이미 움직였다(`research.md` §1.1, AC-ALB-002).

## §8. 문서 정합

- 역할 한정 규칙의 「Loading scope: Intentionally always-loaded」 문구를 역할 주입 경로 설명으로 바꾼다(REQ-ALB-020).
- 상시 표면 구성을 서술하는 다른 문서가 두 파일을 상시 로드로 열거하면 함께 고친다.
