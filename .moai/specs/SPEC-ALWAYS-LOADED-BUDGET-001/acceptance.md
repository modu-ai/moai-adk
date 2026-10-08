---
id: SPEC-ALWAYS-LOADED-BUDGET-001
title: "acceptance — 배포 표면 상시 로드 지시문 예산"
version: "0.8.0"
created: 2026-10-03
updated: 2026-10-08
---

# acceptance.md — SPEC-ALWAYS-LOADED-BUDGET-001

## §D. AC 매트릭스

- 분류: **RB** = 릴리스 차단(새 의무 — RED-now 필수), **RG** = 회귀 가드(이미 성립하는 성질을 지키는 검사 — RED-now 를 요구하지 않음).
- RED-now 는 워크트리 `WT-always-loaded-budget`의 루트에서 실행한 단일 명령과 그 출력·exit 코드다. 2026-10-04 재측정 기준 트리는 `2771626b5`(로컬 develop 흡수 `30ce3a02d` 뒤 — 템플릿 바이트는 `b5815ca80` 에서 벗어났다, `research.md` §1.2). 표의 명령은 적힌 그대로 실행된다(경로 약어 없음). 출력은 축자이며, stdout 이 아닌 stderr 이면 그렇게 표기한다. 여러 파일을 잴 때의 RED-now 명령은 `/usr/bin/grep` 으로 적는다 — 이 환경의 `grep` 은 쉘 함수 래퍼라 다중 파일 출력 순서가 실행마다 달라질 수 있어서이고, 줄 대조는 순서를 무관하게 한다.
- 초록 조건에서 `go test` 판정은 언제나 `-v` 출력의 이름 붙은 `--- PASS: <테스트 이름> `(이름 뒤 공백 포함) 줄로 하고, `-run` 패턴은 `^…$` 로 앵커한다. 패키지 요약 `ok` 줄은 테스트가 0개여도 찍히므로(아래 「RED 근거」 항목의 `[no tests to run]` 판정이 그 실례) 근거가 되지 않는다.
- 원장 고정물 경로: `internal/template/testdata/binding_ledger.json`.
- RED 근거는 비어 있지 않은 실패 신호만 쓴다. 테스트가 0개 실행된 `go test` 출력(`[no tests to run]`, exit 0)은 아무것도 재현하지 않으므로 RED 근거가 아니다. 테스트 부재는 AC-ALB-001 의 grep 으로 보인다.
- **앵커 기대값**: AC-ALB-002·003·004·007 의 기대 숫자(렌더 독립 소계, 구성원 목록, 파일당 초과 파일)는 원장 머리에 기록된 앵커 측정값이다(`plan.md` §C.2 4단계). plan 단계 수치(`b5815ca80` 169,018, `d7112d005` 170,593, 흡수 뒤 `2771626b5` 170,193 — `research.md` §1·§1.1·§1.2)는 이력일 뿐 기대값이 아니다.

| AC | REQ | 분류 | 기준 | RED-now 명령 → stdout (exit) | 초록 조건 |
|---|---|---|---|---|---|
| AC-ALB-001 | REQ-ALB-001 | RB | 예산 테스트 존재 | `grep -rl TestDeployedAlwaysLoadedCharBudget internal/template` → (출력 없음) (exit 1) | 같은 명령이 테스트 파일 1개 출력 |
| AC-ALB-002 | REQ-ALB-001, REQ-ALB-004 | RB | 표면이 사용자 `moai init` 표면과 같음 | `grep -rl TestDeployedAlwaysLoadedCharBudget internal/template` → (출력 없음) (exit 1) — 측정할 테스트가 없음 | `go test ./internal/template/ -run '^TestDeployedAlwaysLoadedCharBudget$' -count=1 -v` 를 M1 RED 커밋(앵커 뒤 착지, 그 커밋과 앵커 사이 템플릿 diff 무출력)에서 실행할 때 `deployed-surface-member=` 줄이 원장 머리의 구성원 목록과 정확히 같고(plan 단계 참고: `b5815ca80` 트리에서는 규칙 13개, `CLAUDE.md`, `AGENTS.md`, `.moai/config/sections/user.yaml`, `.moai/config/sections/language.yaml` 이었다 — 기대값은 이 참고가 아니라 원장 머리다), 렌더 독립 구성원 합이 원장 머리의 소계와 같음 |
| AC-ALB-003 | REQ-ALB-002 | RB | 예산 초과 시 실패 | `grep -rl TestDeployedAlwaysLoadedCharBudget internal/template` → (출력 없음) (exit 1) | M1 RED 커밋의 같은 명령 출력에 `--- FAIL: TestDeployedAlwaysLoadedCharBudget `, `115000`, `deployed-surface-total=` 줄, 상위 5개 이름 |
| AC-ALB-004 | REQ-ALB-003 | RB | 상시 표면 파일당 40,000 초과 시 실패 | `grep -rl TestDeployedAlwaysLoadedCharBudget internal/template` → (출력 없음) (exit 1) — 파일당 검사가 없음. 참고: 이 트리(`2771626b5`)의 UTF-16 계수기 40,000 초과는 `workflow/factory-dispatch-detail.md` 40,659 와 `workflow/spec-workflow.md` 40,052 둘이지만, 둘 다 최상위 `paths:` 를 가진 파일이라 Q10(a) 축소 뒤 이 검사의 대상이 아니다(`research.md` §1.2·§1.3 — 앵커 기준 상시 표면 40,000 초과는 0개) | M1 RED 커밋의 같은 출력에 원장 머리가 기록한 상시 표면 초과 파일 목록이 그대로 기록되고(앵커 기준 빈 목록), 40,001 단위 이상을 넘긴 상시 규칙 픽스처 하위 테스트가 `--- FAIL` 과 그 파일명·크기를 냄 |
| AC-ALB-005 | REQ-ALB-004 | RB | 표면 도출 변이 검사 | `grep -rl TestDeployedAlwaysLoadedCharBudget internal/template` → (출력 없음) (exit 1) | 픽스처 하위 테스트 각각 `--- PASS`: 상시 규칙 추가 → 합계 증가, `paths:` 규칙 추가 → 불변, import 추가 → 증가. 하드코딩 목록 변이에서 `--- FAIL` |
| AC-ALB-006 | REQ-ALB-005 | RB | 상수 옆 근거 주석 | `grep -rl TestDeployedAlwaysLoadedCharBudget internal/template` → (출력 없음) (exit 1) — 테스트와 상수가 없음 | 상수 선언 바로 위 주석에 `120000` 과 증가 실측 근거 문구 |
| AC-ALB-007 | REQ-ALB-002 | RB | **최종 초록** | `grep -rl TestDeployedAlwaysLoadedCharBudget internal/template` → (출력 없음) (exit 1) — 가드가 없음. 표면 크기는 `research.md` §1.1(`d7112d005` 렌더 독립 소계 170,593 > 115,000) | 착지 트리에서 AC-ALB-002 명령 출력에 `--- PASS: TestDeployedAlwaysLoadedCharBudget ` |
| AC-ALB-008 | REQ-ALB-006, REQ-ALB-020 | RB | 역할 한정 2개가 상시 표면에 stub 로만 남음 | `/usr/bin/grep -c 'Intentionally always-loaded' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging.md` → `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:1` `internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging.md:1` (exit 0) | 같은 명령이 두 파일 모두 `:0`(exit 1), AC-ALB-007 실행의 `deployed-surface-member=` 줄에 두 전체 본문이 없고 stub 2개가 있음 |
| AC-ALB-009 | REQ-ALB-007 | RB | 역할 세션 주입 | `grep -l -e role-core -e factory-dispatch internal/hook/session_start_factory.go internal/hook/factory_messages.go internal/hook/subagent_start.go` → (출력 없음) (exit 1) — 구 `session_start_kanban.go` 는 t1399 M5b 에서 제거됨 | 훅 테스트 `--- PASS`: 레지스트리의 리더·레인 표지 × source `startup`/`clear`/`compact` 각각에서 주입 맥락에 역할 core 블록 |
| AC-ALB-010 | REQ-ALB-008 | RG | 비역할 세션·`resume` 무주입 | 현재 훅은 역할 core 를 주입하지 않음(AC-ALB-009 RED-now) | 훅 테스트 `--- PASS`: 표지 없는 환경 `startup` 과 역할 환경 `resume` 에서 역할 core 블록 0회 |
| AC-ALB-011 | REQ-ALB-009 | RB | 실패 가시화 | AC-ALB-009 와 같음(경로 부재) | 훅 테스트 `--- PASS`: 파일 부재·빈 파일·영역 표지 없음 픽스처 각각에서 `systemMessage` 경고와 Read 지시가 함께 있음 |
| AC-ALB-012 | REQ-ALB-010 | RB | 크기 상한 처리 — 최종 `additionalContext` 합본 기준 | AC-ALB-009 와 같음(경로 부재) | 훅 테스트 `--- PASS`: (a) 기존 맥락(세션 귀속문·공지) + core 합본이 10,000 이하 → 주입 맥락에 역할 core 전체가 그대로 실림; (b) 합본이 10,000 초과(core 단독은 이하인 경계 픽스처 포함) → 훅 출력이 역할 core 를 단 하나도 자르지 않고 전부 담고, 운영자 경고와 오버플로 안내 지시(런타임이 전체 출력을 세션 파일로 저장해 경로 + 2,000자 미리보기를 전달한다는 안내)를 함께 냄 — 파일 저장·경로 전달은 훅 테스트보다 상류 런타임 행위라 이 검사는 훅 출력만 단정; (c) 오버플로 전달 불가(시뮬레이션) → REQ-ALB-009 의 운영자 경고 + Read 지시. 모든 경우 잘린 단위 0 |
| AC-ALB-013 | REQ-ALB-011 | RB | 역할 가드 — 리더가 규칙 없이 뜨지 못함 | AC-ALB-009 와 같음(가드 부재) | 가드 `--- PASS`. 역할 core 블록 하나를 지운 변이 → `--- FAIL`. 레지스트리에 표지 하나를 더한 픽스처 → 가드가 그 표지도 검사(하위 테스트 이름에 표지 출현) |
| AC-ALB-014 | REQ-ALB-023 | RB | 역할 core 는 배포 파일에서만 만든다 | `grep -rl 'moai:role-core-start' internal/template/templates` → (출력 없음) (exit 1) | 같은 명령이 역할 한정 규칙 2개 출력. 테스트 `--- PASS`: 원장 고정물이 없는 `t.TempDir()` 배포 트리에서 생성 함수가 비어 있지 않은 역할 core 를 반환 |
| AC-ALB-015 | REQ-ALB-024 | RB | 진입점 — `manager-lead`(deputy 포함) | `/usr/bin/grep -c moai:role-rules-required internal/template/templates/.claude/agents/moai/manager-lead.md internal/template/templates/.claude/skills/moai-factory-foreman/SKILL.md internal/template/templates/.claude/skills/moai/workflows/gtd.md` → `internal/template/templates/.claude/agents/moai/manager-lead.md:0` `internal/template/templates/.claude/skills/moai-factory-foreman/SKILL.md:0` `internal/template/templates/.claude/skills/moai/workflows/gtd.md:0` (exit 1) | `internal/template/templates/.claude/agents/moai/manager-lead.md` 에 표지 1개 이상과 두 규칙 Read 지시, `make agents-emit` 뒤 `internal/template/agentemit` 테스트 `--- PASS` |
| AC-ALB-016 | REQ-ALB-024 | RB | 진입점 — `moai-factory-foreman` 루프(구 `moai-kanban-foreman`, t1399 개명) | AC-ALB-015 의 명령·출력(`internal/template/templates/.claude/skills/moai-factory-foreman/SKILL.md:0`) | `internal/template/templates/.claude/skills/moai-factory-foreman/SKILL.md` 에 표지와 Read 지시 |
| AC-ALB-017 | REQ-ALB-024 | RB | 진입점 — `/moai gtd`·`/moai:todo` `--auto` | AC-ALB-015 의 명령·출력(`internal/template/templates/.claude/skills/moai/workflows/gtd.md:0`) | `internal/template/templates/.claude/skills/moai/workflows/gtd.md` 의 `--auto` 절에 표지와 Read 지시 |
| AC-ALB-018 | REQ-ALB-024 | RB | 인용 파일 전수 가드 | AC-ALB-015 의 명령·출력(세 파일 모두 `:0`, exit 1 — 표지도 가드도 없음) | 가드 `--- PASS`: 배포 트리에서 두 규칙을 인용하는 agent·skill·workflow 파일을 모두 찾아 표지를 요구. 인용만 있고 표지 없는 픽스처 파일 → `--- FAIL` |
| AC-ALB-019 | REQ-ALB-012, REQ-ALB-015, REQ-ALB-025 | RB | 구속 원장 대조 | `ls internal/template/testdata/binding_ledger.json` → stdout 없음, stderr `ls: internal/template/testdata/binding_ledger.json: No such file or directory` (exit 1) | 원장 테스트 `--- PASS`: 원장 대상 파일 전부의 모든 단위(spec §B 단위 경계)에 정확히 1행, 위치가 어휘 안, `always:` 행은 배포 파일에서, `role-core:` 행은 생성 함수 출력에서 변경 후 텍스트 발견. 각각 `--- FAIL` 을 내야 하는 변이 픽스처: `binding`/`normative` 행의 `entry_points` 필드 누락, 빈 목록, `role-core:` 행의 진입점 하나를 `delivery: always` 로 바꾼 행(덮이지 않는 진입점), `binding` 행 하나의 위치를 최상위 `paths:` 를 가진 역할 한정 규칙의 전체 본문 경로로 바꾼 행(`always:` 표면 구성원이 아니므로 REQ-ALB-015 위반) |
| AC-ALB-020 | REQ-ALB-015, REQ-ALB-012 | RB | 단위 변이 — 연속 줄·규범 문단 삭제와 재분류가 판정을 바꾼다 | AC-ALB-019 의 명령·출력(원장·테스트 부재) | 하위 테스트 `--- PASS`: (a) 추출기는 `[HARD]` 문단 + 토큰 없는 하위 항목 픽스처를 한 블록으로, 하위 제목이 끼어든 픽스처는 그 제목에서 두 단위로 나누고, 코드 펜스 안에 `#` 줄·빈 줄·`MUST` 가 있는 픽스처는 한 단위로 둔다. (b) 변경 후 파일에서 어느 블록의 토큰 없는 연속 줄 하나를 지운 변이 → 원장 테스트 `--- FAIL` 이 그 블록 ID 를 이름 댐. (c) `cross-session-messaging.md` 의 `STOPPED_TEAMMATE_VIOLATION` 규범 문단을 지운 변이 → `--- FAIL`; 그 행의 종류를 `rationale` 로 바꾸고 companion 으로 옮긴 변이 → 앵커 종류 불일치로 `--- FAIL`. (d) `goal-directive.md` 의 「Arming a goal does not authorize autonomous run-phase entry」 문단을 지운 변이, 그리고 그 문단을 companion 으로 옮긴 변이 → 각각 `--- FAIL` |
| AC-ALB-021 | REQ-ALB-013 | RB | on-demand 표면에 구속 의무 조각 0 | AC-ALB-019 의 명령·출력 | 원장 테스트가 단정: (1) 종류가 `binding`·`normative` 인 행 가운데 위치가 `companion:` 이거나 skill 경로인 행 0 — `rationale` 행의 `companion:` 위치와, 최상위 `paths:` 를 가진 역할 한정 규칙 2개를 가리키는 `role-core:` 행은 이 단정의 대상이 아니다. (2) 조각 규칙: `binding`·`normative` 행의 변경 후 텍스트에서 공백만 다듬은 길이가 40 UTF-16 코드 단위 이상인 줄 각각이, 원장 대상이 아닌 배포 companion 파일과 배포 skill 파일 어느 곳에도 새로 나타나지 않음 — 앵커 트리의 같은 companion 에 이미 있던 줄은 제외 |
| AC-ALB-022 | REQ-ALB-014 | RG | 행동 트리거 규칙에 `paths:` 신설 없음 | `grep -l '^paths:' internal/template/templates/.claude/rules/moai/core/askuser-protocol.md internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md internal/template/templates/.claude/rules/moai/core/verification-claim-integrity.md internal/template/templates/.claude/rules/moai/core/moai-constitution.md internal/template/templates/.claude/rules/moai/core/moai-mcp-tools.md internal/template/templates/.claude/rules/moai/core/native-idiom-and-register.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/session-handoff.md internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging.md internal/template/templates/.claude/rules/moai/workflow/context-window-management.md internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md internal/template/templates/.claude/rules/moai/workflow/goal-directive.md internal/template/templates/.claude/rules/moai/workflow/cache-aware-execution.md` → (출력 없음) (exit 1) | 같은 명령이 `workflow/factory-dispatch.md`·`workflow/cross-session-messaging.md` 두 개만 출력하고, 그 둘은 AC-ALB-009·015 가 초록 |
| AC-ALB-023 | REQ-ALB-016, REQ-ALB-017 | RB | `§` 교차 참조 해석과 이관 포인터 | `grep -rl TestDeployedRuleSectionRefsResolve internal/template` → (출력 없음) (exit 1) | `go test ./internal/template/ -run '^TestDeployedRuleSectionRefsResolve$' -count=1 -v` 출력에 `--- PASS: TestDeployedRuleSectionRefsResolve `: 배포 규칙 트리의 `<file>.md § <heading>` 참조 미해석 0, M0 후보표의 이관 절마다 원래 core 에 포인터 줄 1개 이상 |
| AC-ALB-024 | REQ-ALB-018, REQ-ALB-019 | RB | Template-First·미러 등록·중립성 | `grep -c -e factory-dispatch -e cross-session-messaging internal/template/rule_template_mirror_test.go` → `1` (exit 0 — `cross-session-messaging-detail.md` 만 등록) | 같은 파일에 바뀌거나 새로 생긴 규칙 쌍 전부 등록, 미러 테스트·템플릿 중립성 가드·내부 내용 누출·카드 id 누출 테스트 각각 `--- PASS`, `make build` 뒤 `git status --porcelain` 에 미커밋 생성물 없음 |
| AC-ALB-025 | REQ-ALB-021, REQ-ALB-022, REQ-ALB-015 | RB | 측정 기록·중단 지점·원장 재고정 | `grep -c 'pending run-phase' .moai/specs/SPEC-ALWAYS-LOADED-BUDGET-001/progress.md` → `2` (exit 0 — §E.2·§E.3 미기록) | `progress.md` 에 M0 하한, 전후 합계, 첫 턴 맥락 실측, Q4·Q8 실측. 하한 > 115,000 이면 M1 이후 커밋 없이 리드 보고. `plan.md` §C.2 1단계의 diff 가 비지 않은 흡수마다 재고정 커밋이 그 흡수 뒤 첫 규칙 편집보다 앞섬(`git log --first-parent`); diff 가 빈 흡수는 재고정 커밋을 요구하지 않음; 원장 고정물 파일의 `git log` 대조에서 anchor kind 를 바꾼 커밋은 `plan.md` §C.2 재고정 커밋뿐임 |

## §D.1 시나리오 (Given-When-Then)

- **AC-ALB-002** — Given 원장 앵커 뒤에 착지하고 앵커와 템플릿 바이트가 같은 M1 RED 커밋, When 예산 테스트를 `-v` 로 돌리면, Then 로그의 구성원 목록과 렌더 독립 소계가 원장 머리의 앵커 측정값과 같다.
- **AC-ALB-020 (c)** — Given 원장과 착지 후 배포 트리, When `STOPPED_TEAMMATE_VIOLATION` 규범 문단을 지우거나, 그 행의 종류를 `rationale` 로 바꿔 companion 으로 옮기면, Then 원장 테스트가 각각 실패한다 — 앞은 행 텍스트 부재로, 뒤는 앵커 종류 불일치로.
- **AC-ALB-020 (d)** — Given 구속 토큰이 없는 `goal-directive.md`, When 그 안의 run-phase 진입 금지 문단을 지우거나 companion 으로 옮기면, Then 원장 테스트가 실패한다.
- **AC-ALB-007** — Given 착지 트리, When 같은 명령을 돌리면, Then `--- PASS: TestDeployedAlwaysLoadedCharBudget ` 가 출력된다.
- **AC-ALB-009** — Given 레지스트리의 레인 표지가 있는 환경, When SessionStart 가 source `compact` 로 들어오면, Then `additionalContext` 에 역할 core 블록이 다시 들어간다.
- **AC-ALB-011** — Given 역할 표지가 있고 `factory-dispatch.md` 가 없는 프로젝트, When SessionStart 가 들어오면, Then 운영자 경고와 Read 지시가 함께 나가고 조용한 성공은 없다.
- **AC-ALB-015** — Given 배포된 `manager-lead` 정의, When 에이전트 본문을 읽으면, Then 첫 행동 전에 두 규칙 전체를 Read 하라는 구속 지시와 표지가 있다. 보좌 deputy 는 같은 정의로 뜨므로 같은 지시를 받는다.
- **AC-ALB-018** — Given 두 규칙을 인용하지만 표지가 없는 스킬 픽스처, When 전수 가드를 돌리면, Then 가드가 그 파일을 이름 대며 실패한다.
- **AC-ALB-020** — Given 원장과 착지 후 배포 트리, When 어느 구속 블록의 토큰 없는 연속 줄 하나를 지우면, Then 원장 테스트가 그 블록 ID 를 이름 대며 실패한다.

## §D.2 경계 사례

- 보충 평면 문자(이모지 등): UTF-16 에서 2단위로 센다.
- 최상위가 아닌 들여쓴 `paths:`: 상시 로드로 판정한다(`instructions_loaded.go` 와 같은 규칙).
- 닫는 `---` 가 없는 frontmatter: 상시 로드로 판정한다.
- `CLAUDE.md` 의 `@AGENTS.local.md`: 새 init 트리에는 없으므로 건너뛴다.
- `@`-import 가 프로젝트 밖을 가리키면 따라가지 않는다.
- 코드 펜스: 펜스 내부는 불투명하다 — 펜스 안의 `#` 줄·빈 줄·구속 토큰은 단위를 열거나 끊지 않는다. 펜스 밖 하위 제목은 수준과 무관하게 단위를 끊는다(`spec.md` §B 단위 경계).
- 역할 core 블록이 0개인 역할 한정 규칙이 생기면 주입은 포인터만 담는다.

## §D.3 품질 게이트

- 변경 패키지만 테스트한다: `./internal/template/...`, `./internal/hook/...`, 에이전트 정의를 고쳤으면 `./internal/template/agentemit/...`. 전체 스위트는 CI 몫이다.
- `go vet` 와 CI 판 `golangci-lint` 0 경고.
- 새 테스트는 `t.TempDir()` 만 쓰고, 환경 변수는 `t.Setenv` 로 한정한다(병렬 테스트에서 역할 표지 변수 충돌 금지).

## §D.4 완료 정의

- RB 23개 초록, RG 2개(AC-ALB-010·022) 성립.
- 착지 트리에서 예산 테스트가 보고하는 합계가 115,000 이하.
- 구속 원장의 `rewrite` 행 전부가 감사에서 의미 보존으로 판정됨.
- `decision-index.md` 의 Q4·Q8 실측이 `progress.md` 에 기록됨.
