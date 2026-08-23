---
id: SPEC-TODO-ANALYSIS-001
title: "todo 큐 자동 분석 — 기계적 중복 탐지와 관계 기록, 그리고 큐레이션 금지선의 명문화"
version: "1.0.0"
status: completed
created: 2026-08-22
updated: 2026-08-23
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: internal/cli
lifecycle: spec-anchored
tags: "kanban, todo, backlog, analysis, doctrine"
tier: M
---

# SPEC-TODO-ANALYSIS-001 — todo 큐 자동 분석

## §A 배경 — 측정된 형태

### A.1 지금 있는 것

`moai todo`는 `origin/main@1519f2660` 기준으로 **9개 서브커맨드**를 갖는다 (`internal/cli/todo.go` `newTodoCmd` 의 `AddCommand` 호출로 확인):
`add`, `list`, `done`, `next`, `unpick`, `edit`, `move`, `drop`, `undrop`.

동사 부재는 이 카드의 범위가 아니다. `edit` / `move` 는 `internal/cli/todo_edit_move.go`, `drop` / `undrop` 은 `internal/cli/todo_drop.go` 에 이미 있다.

### A.2 실제로 없는 것

큐 경로 어디에도 **분석 계층이 없다**. `internal/cli/todo*.go` 와 `internal/kanban/backlog_store.go` 를 훑으면 유사도·중복·관계·정렬 로직은 한 줄도 없고, 대신 두 파일의 주석이 그것을 **명시적으로 거부**하고 있다:

- `internal/cli/todo_drop.go:13` — "no staleness heuristic, no duplicate detection, no absorption of one card into another … a doctrine change would have to come first."
- `internal/cli/todo_edit_move.go:5-7` — "Nothing here infers what a card should say or where it belongs — no analysis, no absorption, no silent promotion."

즉 이 기능은 "빠진 것"이 아니라 **의도적으로 배제된 것**이고, 코드 주석이 그 전제 조건까지 적어 두었다 — 독트린이 먼저 바뀌어야 한다.

### A.3 정면으로 충돌하는 독트린 두 곳

`.claude/skills/moai/workflows/todo.md` [HARD]:

> `edit`, `move`, `drop`, `undrop` 는 `add` 및 pick 과 똑같은 **운영자의 행위**다. … **추론된 우선순위로 절대 안 되고, 정리(tidy-up) 목적으로 안 되고, 한 카드를 다른 카드에 접어 넣으려고 안 되고, 카드가 오래돼 보인다고 안 된다.** 큐는 운영자의 의도를 기록할 뿐, 그것을 큐레이션하지 않는다.

`.claude/rules/moai/workflow/kanban-dispatch.md` § Entry into the board is an operator act:

> 리드는 큐의 유일한 생산자이며, 생산은 "발명이 아니라 번역"이다. 승격은 언제나 운영자의 행위이고, 리드는 **추론된 우선순위로 재정렬하지 않는다**.

카드가 요구한 세 동작이 이 금지 목록에 **이름으로** 들어 있다. 중복 제거 = 접어 넣기, 관계 판정 = 추론, 순서 정리 = 추론된 우선순위 기반 정리. 독트린은 이 기능을 허용하지 않는 정도가 아니라 **호명해서 금지**한다.

### A.4 저장소가 강제하는 제약

- 항목당 5필드(`id`, `text`, `added_at`, `spec_id`, `state`)는 SPEC-KANBAN-TODO-CLI-001 REQ-TODO-013 이 고정한 계약이다. 다만 그 SPEC의 §E는 **자기 범위**의 out-of-scope 선언이지 영구 동결이 아니고, 같은 REQ가 "additively" 변경을 허용한다 — `last_seq` 최상위 필드가 그 선례다.
- `drop` 은 새 필드 없이 텍스트 접두 마커(`[DROPPED — <reason>] `)로 사유를 기록한다. 스키마를 건드리지 않고 기록하는 선례가 이미 있다.
- `moai todo` 는 **절대 질문하지 않는다** (REQ-TODO-014, headless-safe). 따라서 "분석 결과를 운영자에게 물어본다"는 형태는 CLI 안에서 성립하지 않는다.

## §B 이 SPEC이 취하는 입장

운영자는 2026-08-19에 권고형이 아니라 **자동형**을 선택했다. 그 선택을 조용히 권고형으로 되돌리지 않는다. 대신 **자동화되는 것이 무엇인지**를 갈라서, 독트린이 지키려던 것을 실제로 지키면서 자동 형태를 유지한다.

### B.1 금지의 진짜 이유는 "변경"이 아니라 "안 보이는 변경"이다

독트린이 막는 손해는 두 겹이다. ① 운영자의 의도가 **파괴**되고, ② 파괴된 사실이 **보이지 않는다**. 되돌릴 수 있게 만들면 ①은 완화된다. 그러나 되돌리기는 **누군가 잘못을 알아챘을 때만** 발동한다. 알아챌 수 없는 변경에 대해서는 가역성이 아무것도 사 주지 못한다.

**운영자는 이 위험을 이미 검토하고 완화안까지 적어 두었다.** 카드 t119 본문:

> 오판 1건이 카드를 조용히 삼킬 수 있으므로 드롭은 되돌릴 수 있어야 한다(undrop 동사 존재).

완화안은 위험을 정확히 지목한다. 갈리는 지점은 위험의 식별이 아니라 완화의 **발동 조건**이다. `undrop` 은 실재하지만, 운영자가 "저 카드가 있어야 하는데 없다"고 먼저 알아채야 손이 간다. 흡수와 재정렬은 하필 그 알아챔을 없애는 변형이다 — 흡수된 카드는 애초에 없던 카드와 구별되지 않고, 재정렬된 큐는 운영자의 평상시 화면에서 자기가 정렬한 큐와 구별되지 않는다. 완화안이 요구하는 전제(누군가 알아챈다)를 변형 자체가 무너뜨리는 셈이다. 그러므로 완화안이 틀린 것이 아니라, 이 두 동작에 대해서만 **닿지 않는다**. 이 SPEC 은 완화안을 폐기하지 않는다 — `undrop` 은 그대로 남고, 운영자의 `drop` 은 여전히 되돌릴 수 있다. 그 위에 조건을 하나 더 얹을 뿐이다.

그래서 이 SPEC이 세우는 판정 기준은 가역성 단독이 아니라 **가역성 + 현저성(conspicuousness)** 이다:

> **운영자가 저작한 카드의 내용·순서·존재**에 대한 자동 변경은, 잘못된 결과가 **운영자가 의심하지 않아도 스스로 드러날 때에만** 허용된다.

범위 한정("운영자가 저작한 …")은 의도적이다. 기준이 지키려는 것은 운영자의 의도이지 저장소 내부 불변식이 아니다. 예컨대 `normalizeBacklogRecord`(`internal/kanban/backlog_store.go:326`)가 `last_seq` 를 최대 present id 까지 조용히 끌어올리는 수리는 카드의 내용·순서·존재를 하나도 바꾸지 않으므로 이 기준의 대상이 아니다. 한정어가 없으면 기준이 문자 그대로 기존 정상 동작까지 걸어 버린다.

이 기준으로 카드의 세 요구를 각각 재면 답이 갈린다.

| 동작 | 가역? | 현저? | 판정 |
|---|---|---|---|
| 정확 중복 add 를 **거절** | 아무것도 안 바꾸므로 자명 | **호출자가 exit code 또는 stderr 를 읽을 때에만** — 그때 운영자는 id 대신 에러를 받는다 | **조건부 자동 허용** (아래 잔여 위험) |
| 근접 중복을 **기록** | 소견 삭제로 가역 | 예 — `list` 가 매번 보여준다 | **자동 허용** |
| 한 카드를 다른 카드에 **접어 넣기** | `undrop` 으로 부분 가역 | **아니오** — 사라진 카드는 애초에 없던 카드와 구별되지 않는다 | **금지** |
| 순서 **정리** | 이전 순서를 알아야 가역 | **아니오** — 운영자의 기본 화면에서 구별되지 않는다 | **금지** |

순서 정리가 결정적이다. 순서는 큐가 우선순위에 대해 기록하는 **유일한 신호**이고(`internal/cli/todo_edit_move.go:99` "Order is the only thing the queue records about priority"), `moai todo list` 의 사람용 렌더는 `id`·`state`·`text` 3열만 찍는다(`internal/cli/todo.go:309`) — `added_at` 은 화면에 아예 없다. 파일 자체는 단조 증가하는 `t<N>` id 와 `added_at` 을 담으므로 **재정렬이 일어났다는 사실은 기계적으로 탐지된다**; 탐지되지 않는 것은 *누가·왜* 재정렬했는가이고, 운영자가 평상시 보는 화면에는 애초에 아무 흔적도 뜨지 않는다. 잘못된 재정렬이 알아채이지 않는 이유가 이것이고, 그래서 되돌리기 경로를 아무리 잘 만들어도 발동되지 않는다.

**잔여 위험 — 기준이 닿지 않는 호출 경로.** 정확 중복 거절의 현저성은 "호출자가 exit code 나 stderr 를 읽는다"는 전제 위에 선다. `manager-lead` 는 카드 id 를 받아야 디스패치가 되므로 실패를 놓칠 수 없지만, `§B.3` 이 요구하는 **에이전트 없는 순수 CLI 호출**(스크립트·파이프)에서 호출자가 종료 코드와 stderr 를 둘 다 버리면 카드는 조용히 유실된다. CLI 는 호출자가 자기 종료 코드를 읽게 만들 수 없다 — 이 경로를 덮는 보완 장치를 이 SPEC 은 만들지 않고 **명시적 잔여 위험으로 선언한다**. `--force` 는 이 경우의 탈출구가 아니라 의도적 중복 입력을 위한 것이다. 같은 위험이 `plan.md §D.3` 과 `§G` 에 기록돼 있다.

### B.2 그래서 "자동"인 것은 분석이지 변형이 아니다

- **분석은 자동으로 돈다** — `add` 할 때마다, 요청 없이. 운영자가 고른 자동 형태는 이렇게 보존된다.
- **분석이 일으키는 변형은 정확히 하나** — 정규화 후 텍스트가 완전히 같은 카드의 **입력 거절**. 이것은 기존 카드를 건드리지 않고, 큐 파일을 바이트 단위로 그대로 두며, 운영자에게 즉시 에러로 나타난다. "한 카드를 다른 카드에 접어 넣는" 행위가 아니라 **아무 카드도 만들지 않는** 행위다.
- **나머지 전부는 기록만 한다.** 근접 중복도, 네 관계도 카드를 바꾸지 않는다.

### B.3 판단하는 곳과 쓰는 곳은 다르다

에이전트가 판단하고 CLI가 쓴다. 두 계층은 신뢰 경계가 다르므로 능력도 다르다.

| 계층 | 주체 | 능력 | 사람이 CLI를 직접 쓸 때 |
|---|---|---|---|
| L1 기계 | Go CLI, 락 안 | 정규화 텍스트 비교뿐. 결정적, 모델 없음 | **작동한다** |
| L2 의미 | `relate` 를 호출하는 누구나 — 관행상 `manager-lead` (REQ-TA-008 은 행위자를 제약하지 않는다) | 네 관계 판정 | **작동하지 않는다** |

L2가 없는 호출에서도 큐가 "분석 완료"로 보이면 안 된다. 그래서 같은 쌍에 `source: agent` 소견이 하나도 없는 기계 소견은 `machine-only` 로 표시된다(REQ-TA-013).

**이 표식이 추적하는 것은 검토가 아니라 에이전트 출처 기록의 유무다.** 표식 이름을 그 의미에 맞춘 이유가 이것이다. REQ-TA-008 은 `relate` 에 행위자 제약을 두지 않는다 — `manager-lead` 든 다른 레인이든 스크립트든, `source: agent` 소견을 쓸 수 있는 자는 누구나 표식을 끈다. CLI 는 자기를 누가 호출했는지 강제할 수 없고(§B.1 의 정확 중복 거절이 부딪힌 것과 같은 한계다), 그래서 "검토됐다"를 증명하는 표식은 이 계층에서 만들 수 없다. 만들 수 있는 것은 "기계 소견만 있고 에이전트는 아무것도 기록하지 않았다"는 사실의 표시뿐이며, `machine-only` 는 정확히 그것만 말한다.

그 축소된 형태로도 표식은 제 몫을 한다. 막으려던 실패는 **분석되지 않은 큐가 깨끗한 큐로 위장하는 것**이고, 그 위장은 에이전트가 아무 기록도 남기지 않았을 때 일어난다 — 표식은 바로 그 경우를 잡는다. 잡지 못하는 것은 에이전트가 기록은 남겼으나 그 판단이 부실한 경우이고, 그것은 표식이 아니라 소견 본문을 읽어야 알 수 있다. 표식은 **읽을 것이 있는지**를 말하지, 읽은 것이 옳은지를 말하지 않는다.

### B.4 네 관계가 각각 무엇을 일으키는가

관계는 대칭이 아니고, 넷 중 셋은 **아무것도 일으키지 않는다**. 이 표가 그 답이다.

| 관계 | 출처 | 일으키는 일 |
|---|---|---|
| `duplicate` (정확) | 기계 | **입력 거절.** 이 기능 전체에서 유일한 변형 |
| `near-duplicate` | 기계 | 기록만. 해소는 운영자의 `drop` / `edit` |
| `contains` | 에이전트 | 기록만 |
| `absorbs` | 에이전트 | 기록만. 흡수는 독트린이 이름으로 금지한 바로 그 행위이므로 운영자의 행위로 남는다 |
| `replaces` | 에이전트 | 기록만 |
| `conflicts` | 에이전트 | 기록만. 해소안도 제시하지 않는다 — 충돌하는 두 카드가 둘 다 정당할 수 있다 |

## §C 요구사항 (GEARS)

### C.1 기계 분석 계층

- **REQ-TA-001** (Ubiquitous) The backlog store **shall** carry a mechanical text analyser that normalizes a card text (Unicode NFC → trim → collapse internal whitespace → case-fold) and classifies a candidate against every non-dropped card as `exact` (normalized texts equal), `near` (token-set Jaccard score ≥ 0.80 and < 1.0), or `none`.
- **REQ-TA-002** (Event-driven) **When** `moai todo add` runs, the command **shall** execute the mechanical analysis inside the same locked write that would append the card, before the append; **when** `moai todo analyze` runs, the command **shall** execute the same analysis across the whole queue and record findings without appending, removing, reordering, or editing any card, and **shall not** append a finding whose `{subject_id, related_id, relation, source}` tuple already exists, so two consecutive `analyze` runs leave the `findings` array length unchanged.
- **REQ-TA-003** (Event-driven) **When** the analysis classifies a candidate as `exact` against a card in state `queued` or `picked`, the command **shall** refuse the append, leave the queue file byte-identical, name the colliding card's id and text prefix on stderr, and exit non-zero.
- **REQ-TA-004** (Capability-gate) **Where** `--force` is passed to `moai todo add`, the command **shall** append the card despite an `exact` classification and **shall** record one `duplicate-forced` finding naming the colliding id.
- **REQ-TA-005** (Event-driven) **When** the analysis classifies a candidate as `near`, the command **shall** append the card with its text exactly as given, **shall** leave every field of the related card unchanged, and **shall** record one `near-duplicate` finding carrying the related id and the computed score.

### C.2 소견 레코드 (가산적 스키마)

- **REQ-TA-006** (Ubiquitous) The backlog record **shall** carry an additive top-level `findings` array — absent-tolerant on load, each entry holding `{subject_id, related_id, relation, source, score, note, at}` with `relation ∈ {duplicate-forced, near-duplicate, contains, absorbs, replaces, conflicts}` and `source ∈ {mechanical, agent}` — and the five-field per-item contract of REQ-TODO-013 **shall** remain unchanged.
- **REQ-TA-007** (Event-driven) **When** a card is removed by `moai todo done`, the command **shall** remove every finding naming that card in either the `subject_id` or the `related_id` position, so no finding can outlive its subject.

### C.3 에이전트 계층의 관계 기록

- **REQ-TA-008** (Event-driven) **When** `moai todo relate <a> <b> --relation <r> [--note <text>]` runs, the command **shall** append exactly one `source: agent` finding under the lock and **shall** change no field of card `<a>` or card `<b>`; **when** `moai todo unrelate <index>` runs, the command **shall** remove the addressed finding under the lock and **shall** change no card.
- **REQ-TA-009** (Unwanted) The `moai todo` command **shall not** drop, edit, move, reorder, absorb, or change the state of any card as a consequence of any recorded finding, any `relate` invocation, or any analysis result.
- **REQ-TA-010** (Unwanted) The mechanical analyser **shall not** run on any command other than `add` and `analyze`, so two consecutive `moai todo list` invocations with no intervening command **shall** render identical output over an identical file.

### C.4 운영자 가시성

- **REQ-TA-011** (Event-driven) **When** `moai todo list` runs and findings exist, the command **shall** print one indented finding line beneath each card named by a finding, carrying the relation, the counterpart id, the source, and the literal operator command that would act on it.
- **REQ-TA-012** (Event-driven) **When** `moai todo why <n>` runs, the command **shall** print every finding naming card `<n>`, and **shall** print an explicit no-findings line when none exist; **when** `moai todo list --json` runs, the command **shall** emit the `findings` array alongside `items`.
- **REQ-TA-013** (State-driven) **While** a card carries a `source: mechanical` finding with no `source: agent` finding naming the same pair — pairs are compared **unordered**, so `{a, b}` and `{b, a}` are the same pair — `moai todo list` **shall** mark that finding `machine-only`; the mark records the ABSENCE of an agent-sourced finding for the pair, not that any review took place.

### C.5 독트린 개정

- **REQ-TA-014** (Ubiquitous) `.claude/skills/moai/workflows/todo.md` and `.claude/rules/moai/workflow/kanban-dispatch.md`, together with their `internal/template/templates/` mirrors, **shall** state the analysis boundary explicitly: analysis runs automatically and records; it refuses the admission of an exact duplicate; it never folds, reorders, drops, or edits a card; and the four semantic relations cause nothing but a record.
- **REQ-TA-015** (Ubiquitous) The `moai todo` command surface **shall** remain headless and prompt-free across every new verb, preserving REQ-TODO-014 (structured stdout, human-readable stderr, exit 0/1/2, no user-question channel anywhere in `internal/cli`).

## §D 수용 기준

Tier M — 전체 AC 목록과 각 AC를 붉게 만드는 잘못된 구현은 `acceptance.md`.

## §E 범위 밖

### Out of Scope — 자동 순서 정리

- 큐 순서를 자동으로 바꾸는 기능은 **만들지 않는다**. §B.1 의 현저성 판정에서 탈락했다: 재정렬된 큐는 운영자가 그렇게 정렬한 큐와 구별 불가능하므로, 잘못된 재정렬은 되돌리기 경로가 있어도 발동되지 않는다.
- 순서를 바꾸는 유일한 경로는 기존 `moai todo move` — 운영자의 행위로 남는다.

### Out of Scope — 카드 흡수·병합

- 소견이 `absorbs` / `contains` / `replaces` 를 기록해도 어떤 카드도 자동으로 접히거나 병합되지 않는다. 흡수는 `drop` + `edit` 의 운영자 조합으로만 일어난다.
- `conflicts` 는 해소안을 제안하지도 않는다.

### Out of Scope — 의미 기반 중복 탐지의 CLI 구현

- CLI는 임베딩·모델 호출·외부 API를 쓰지 않는다. 기계 계층은 정규화 텍스트 비교로 한정한다.
- 텍스트가 다른 의미상 중복(`"Fix the auth bug"` vs `"Repair broken login"`)은 L2 에이전트만 판정하며, 에이전트 없는 호출에서는 탐지되지 않는다 — 그 사실은 `machine-only` 표시로 드러난다. 에이전트 없는 호출은 `source: agent` 소견을 만들지 않으므로 표식이 그대로 남는다.

### Out of Scope — 큐 스키마 재설계

- 버전 범프 없음. 항목당 5필드 계약 불변. 유일한 스키마 변경은 최상위 `findings` 배열의 가산적 추가(`last_seq` 선례).

### Out of Scope — 보드·SPEC 상태 연동

- `internal/kanban` 의 board 계열(`board.go`, `board_store.go`, `reconcile.go`)과 SPEC frontmatter 는 건드리지 않는다. 소견은 백로그 파일 안에서만 산다.

### Out of Scope — 소견의 원격 공유

- `backlog.json` 은 커밋되지 않는 프로젝트 로컬 파일이다. 소견을 커밋·동기화·PR에 싣는 경로는 만들지 않는다.

## §F 이력

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-08-22 | manager-spec | 초기 plan-phase 초안 (칸반 카드 t119). |
| 1.0.0 | 2026-08-23 | manager-develop | run + sync 완료. M3(기계 분석기 + `add` 통합) · M4(`relate`/`unrelate`/`why` + `list` 소견 렌더 + 가드 범위 확장) · M5(스킬 표면 갱신 + 템플릿 미러 + `make build`) 착지, docs-site 4로케일 반영. AC-TA-001~016 전부 PASS — 귀속은 `progress.md §E.3`. |
| 0.2.0 | 2026-08-23 | manager-spec | plan-audit iteration 1 결함 반영: §B.1 에 카드 t119 완화안 인용·반박(D10), 현저성 기준에 범위 한정 추가(D9), 정확 중복 거절 행에 조건 명시 + 잔여 위험 선언(D2), 순서 탐지 서술 정정(D8), 인용 줄번호 정정(D11), REQ-TA-002 재실행 멱등성(D4) · REQ-TA-013 무순서 쌍(D13) · REQ-TA-015 exit 0/1/2(D6). |

## §G 상호참조

- `.claude/skills/moai/workflows/todo.md` — 개정 대상 [HARD] 절.
- `.claude/rules/moai/workflow/kanban-dispatch.md` § Entry into the board is an operator act — 개정 대상.
- SPEC-KANBAN-TODO-CLI-001 — 락 기반 저장소, 5필드 계약(REQ-TODO-013), headless 계약(REQ-TODO-014).
- `internal/cli/todo_drop.go` / `internal/cli/todo_edit_move.go` — "독트린이 먼저 바뀌어야 한다"는 전제를 적어 둔 주석.
