---
id: SPEC-ALWAYS-LOADED-DIET-001
title: "always-loaded 컨텍스트 표면 다이어트 + 캐시 규율 편입 + 재발 통제"
version: "0.1.1"
status: draft
created: 2026-08-16
updated: 2026-08-17
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".claude/rules/moai/workflow, internal/config, internal/template/templates/.claude/rules/moai"
lifecycle: spec-anchored
tags: "always-loaded, token-budget, context-diet, prompt-cache, rule-split, recurrence-control"
tier: M
era: V3R6
related_specs: [SPEC-V3R6-RULES-PATH-SCOPE-001, SPEC-TOKEN-EFFICIENCY-001]
---

# SPEC: always-loaded 컨텍스트 표면 다이어트 + 캐시 규율 편입 + 재발 통제

## 1. 문제 — 측정된 형태

Claude Code는 매 턴 always-loaded 컨텍스트 표면을 다시 주입하고, `/clear` 이후에는 그 전량을 다시 지불한다. 이 표면에는 이미 기계적 가드가 걸려 있다.

- `internal/config/token_budget_guard.go:26` — `const AlwaysLoadedTokenBudget = 75000`
- `internal/config/token_budget_guard_test.go:53` — `TestAlwaysLoadedTokenBudget`. 표면이 예산을 넘으면 빌드가 깨진다.
- 가드가 재는 슬롯은 4개다: `.claude/rules/moai/**` 중 top-level `paths:` frontmatter 키가 **없는** 규칙 파일 전체 + `CLAUDE.md` + `.claude/output-styles/moai/moai.md` + `MEMORY.md` head. 토큰 추정은 char/4.

2026-08-16 실측(가드 자체의 공식을 그대로 재현):

| 슬롯 | bytes |
|---|---|
| no-`paths:` 규칙 14개 | 210,753 |
| `CLAUDE.md` | 19,040 |
| `.claude/output-styles/moai/moai.md` | 65,251 |
| `MEMORY.md` head (리포 루트에 부재) | 0 |
| 합계 | 295,044 → 추정 73,761 토큰 |

**남은 여유: 1,239 토큰 ≈ 4,956 bytes (1.65%).** `go test ./internal/config/ -run TestAlwaysLoaded` 는 이번 세션에서 실제로 실행해 `ok github.com/modu-ai/moai-adk/internal/config 0.422s` 를 확인했다 — 지금은 통과한다.

문제는 여유의 폭이다. 가장 최근 always-loaded 규칙 추가 3건은 각각 6,200 bytes(`moai-mcp-tools.md`, 2026-08-13), 21,003 bytes(`kanban-dispatch.md`, 2026-08-14), 12,608 bytes(`cross-session-messaging.md`, 2026-08-15)였다. **셋 중 어느 것도 현재 여유 안에 들어가지 않는다.** 통상 크기의 규칙이 하나만 더 들어오면 빌드가 깨진다.

## 2. 이것은 첫 청소가 아니라 재발이다

`SPEC-V3R6-RULES-PATH-SCOPE-001`(status: implemented, created 2026-05-22)이 always-loaded 규칙을 9개에서 5개로 줄였다. 그 직후 커밋 `09d905bda57ed2537ac98aeefd22e2652719a482` 기준과 현재를 비교하면:

- 당시: 5개 파일, 53,299 bytes
- 현재: 14개 파일, 210,753 bytes — **3.95배, 약 3개월 만에 +157,454 bytes**
- 분해: 기존 5개의 자체 증가 +71,347 (45.3%) / 신규 9개 유입 +86,107 (54.7%)

원래 5개의 파일별 증가:

| 파일 | 당시 | 현재 | 증가 |
|---|---|---|---|
| `agent-common-protocol.md` | 15,868 | 39,455 | +23,587 |
| `askuser-protocol.md` | 11,603 | 30,154 | +18,551 |
| `session-handoff.md` | 8,963 | 23,251 | +14,288 |
| `context-window-management.md` | 4,328 | 13,009 | +8,681 |
| `moai-constitution.md` | 12,537 | 18,777 | +6,240 |

읽는 방식이 중요하다. 유입(54.7%)과 자체 증가(45.3%)가 거의 반반이므로, "새 파일을 막는" 통제만으로도 "기존 파일을 줄이는" 작업만으로도 재발을 못 막는다. 그리고 한 번의 정리는 그 자체로는 아무 것도 남기지 않는다 — 재발 통제 없는 다이어트는 위 표를 3개월 뒤에 다시 만들어낸다. 이 SPEC이 M3를 포함하는 이유다.

## 3. 요구사항 (GEARS)

각 요구사항 뒤 괄호는 GEARS 패턴을 가리킨다.

### 3.1 M1 — `kanban-dispatch.md` 분리

- **REQ-ALD-001** (Ubiquitous) — `kanban-dispatch.md` 스텁은 BINDING-EVERYWHERE로 분류된 7개 구간(제목+서문+Loading-scope, `Scope`, `Completion is read, never trusted`(하위 `CodeRabbit` 포함), `Isolation is entered, never provisioned`, `Verification load is lane-local`(하위 `env-isolated verification form` 포함), `Boundaries`, `Cross-references`+푸터)을 인라인으로 보유해야 한다. 분류 술어는 "이 구간이 kanban lead가 **아닌** 세션도 구속하는가"이며, 위 7개는 모두 참이다.
  술어는 **구간 단위가 아니라 문단 단위로** 적용한다. 그렇게 적용하면 LEAD-ONLY 구간 안에서 반례가 정확히 1건 나온다 — `Card classes` 안의 Class B 문단(분리 이전 원본 60행, **495 bytes 실측**)은 주어가 `the run session` 이라 술어가 참이다("the evidence that established the cause ... is written into the card's progress record, and the run session names that path in its completion report"). 따라서 이 문단도 스텁이 보유해야 하며, 자리는 STAY 구간 `Completion is read, never trusted` 다(주제도 그쪽이 맞다 — 증거를 읽는 규율). 재배치는 **줄 단위 통째 이동**이며 문단을 쪼개지 않는다: 쪼개면 원본의 그 줄이 어느 파일에도 그대로 남지 않아 내용 보존 판정(AC-ALD-006, 비어있지 않은 줄 집합 대조)이 거짓 실패한다.
- **REQ-ALD-002** (Ubiquitous) — 신설 `kanban-dispatch-detail.md` 는 LEAD-ONLY 6개 구간(`The board`, `Entry into the board is an operator act`, `Card classes`(단 REQ-ALD-001 이 지정한 Class B 문단 1줄 제외), `The dispatch cycle`(하위 `Dispatch language` 포함), `Review lens selection`, `The /clear handoff between phases`)을 받아야 하며, 스텁과 컴패니언의 합집합은 분리 이전 원본의 내용을 빠짐없이 덮어야 한다.
- **REQ-ALD-003** (Ubiquitous) — `kanban-dispatch-detail.md` 의 frontmatter `paths:` 는 **domain-keyed** 형태여야 한다. 즉 부모 규칙 파일만 가리키는 self-keyed(`**/kanban-dispatch*.md` 단독)가 아니라, 실제 kanban 작업 중에도 붙는 도메인 경로를 함께 키로 가져야 한다. `goal-directive-detail.md` 가 유일한 domain-keyed 선례다.
- **REQ-ALD-004** (Ubiquitous) — 분리는 `goal-directive.md` 선례의 3요소를 갖춰야 한다: (a) 스텁의 포인터 줄이 **옮겨진 구간을 이름으로 열거**하고 과업 형태의 로드 트리거를 진술한다, (b) 컴패니언이 자기 쪽에서 소유 경계를 선언한다, (c) 스텁 푸터의 버전 줄이 분리 사실을 기록한다.
- **REQ-ALD-005** (unwanted) — 컴패니언은 분리 과정에서 원본에 없던 내용을 새로 획득해서는 안 된다(`shall not`). 근거: `session-handoff-examples.md` 는 40,891 bytes 로 부모 `session-handoff.md`(23,251)보다 **크다** — 분리는 축소와 같지 않으며, 컴패니언은 적치장이 되기 쉽다.

### 3.2 M2 — G3~G7 규율을 `cache-aware-execution.md` 로 편입

편입 대상은 `cache-aware-execution.md`(현재 4,497 bytes)다. 이미 prompt-cache와 실행 순서 지시를 소유하고 있어 주제가 맞고, **새 always-loaded 파일을 만들지 않는다**는 점이 결정적이다.

- **REQ-ALD-006** (Ubiquitous) — 파일 내용을 프롬프트에 끌어올 때 파일명을 인용해 모델이 읽게 하는 대신 `@`-멘션으로 넘기는 규율, 그리고 `/context` 를 1회성 감사 수단으로 쓰는 규율을 HARD 지시로 보유해야 한다. (실측 갭: 두 규율 모두 `.claude/rules/` + `CLAUDE.md` 전체에서 매치 0개.)
- **REQ-ALD-007** (Ubiquitous) — 일반 명령 출력 규율을 보유해야 한다. 현재 `agent-common-protocol.md` 의 file-redirect 계약(50줄 / 2KB)은 **검증 배치에만** 걸려 있고, `BASH_MAX_OUTPUT_LENGTH`(30,000자 임계)는 규칙 파일에 0회 등장한다.
- **REQ-ALD-008** (Ubiquitous) — 상용 명령을 조용한 형태(quiet form)로 부르는 규율을 보유해야 한다. 실측 갭: 규칙 내 "quiet" 매치 2건은 모두 영어 형용사 용법("quiet and expensive", "the race is quiet")이며 규율이 아니다.
- **REQ-ALD-009** (Ubiquitous) — 세션 길이를 비용 축으로 다루는 진술을 보유해야 한다. `context-window-management.md` 는 `/clear` 가 always-loaded 프리픽스를 다시 지불한다는 사실은 이미 적고 있으나(line 47), **긴 세션 하나가 짧은 세션 여럿보다 싸다**는 진술은 어디에도 없다(매치 0개).
- **REQ-ALD-010** (Event-driven) — **When** 세션 도중 모델 또는 effort 를 바꾸려 할 때, 그 변경이 prompt cache 를 파기한다는 사실을 진술해야 한다(`MAX_THINKING_TOKENS` 는 규칙 파일에 0회 등장). 아울러 겉보기 충돌 2건을 같은 자리에서 해소해야 한다: `agent-common-protocol.md` § Per-Spawn Model Injection(스폰마다 모델을 명시하라)과 `cache-aware-execution.md` 지시 5(세션 모델을 상속하라)는 **서로 다른 축**이다 — 전자는 서브에이전트 컨텍스트, 후자는 메인 세션이다. 모순으로 읽히게 방치해서는 안 된다.
- **REQ-ALD-011** (Ubiquitous) — HARD 지시 본문만 `cache-aware-execution.md` 에 압축해 인라인으로 두고, 근거·수치·예시는 신설 `cache-aware-execution-reference.md` 로 보내야 한다. 신설 파일은 `paths:` 를 가져 always-loaded 표면에 계상되지 않아야 한다.
- **REQ-ALD-012** (Ubiquitous) — M2가 인용하는 캐시 배수·TTL(read 0.1x, write 최대 2x, output ≈ input 5x, TTL 구독 1시간 / API 키 5분)은 원문 인용이며 이번 세션에서 재측정하지 않았다. 문서에 **인용 출처값**임을 명시해야 하며 측정치로 서술해서는 안 된다.

### 3.3 M3 — 재발 통제

- **REQ-ALD-013** (Ubiquitous) — 다음 always-loaded 추가가 **자기 비용을 진술하도록** 강제하는 통제 장치 하나를 채택해 구현해야 한다. 채택된 통제는 **작성 규칙(문서 전용)** 이며, 신규 always-loaded 파일 생성과 기존 always-loaded 파일의 증가 **양쪽**에 진술 의무를 건다. 이 통제는 문서만으로 성립하고 Go 변경을 수반하지 않는다. 후보 3개(작성 규칙 / 예산 하향 / 여유 보고 가드)의 평가와 이 선택의 근거는 plan.md D1이 소유하며, 셋을 모두 구현하지는 않는다.
  적용 범위는 **가드가 세는 슬롯 4개 전부**다. 통제의 `paths:` 글롭이 `.claude/rules/moai/**` 의 규칙 파일 + `CLAUDE.md` + `.claude/output-styles/**` + `MEMORY.md` 에 모두 붙어야 한다. 네 번째 슬롯을 "리포 루트에 파일이 없어 기여 0" 이라는 **현재 시제 사실**로 면제해서는 안 된다 — 통제는 미래 시제이고, 가드는 이 슬롯을 조건부로(`[ -f MEMORY.md ] && …`) 세므로 파일이 생기는 순간 최대 `memoryHeadByteCap` 25,600 bytes(약 6,400 토큰)가 진술 의무 없이 표면에 편입된다. 그 값은 이번 다이어트의 최악 코너 여유 2,597 토큰보다 크다(감사 iter2 D3). 글롭에 13자를 더하는 비용이 0 이므로 면제할 이유가 없다. 규칙 파일만 키잉하면 도달률이 **71.4%** 에 그친다: `CLAUDE.md`(19,040)와 `.claude/output-styles/moai/moai.md`(65,251)의 합계 84,291 bytes 가 295,044 중 **28.6%** 이고, `moai.md` 는 단일 파일로는 표면 최대 기여자다(규칙 1위 `agent-common-protocol.md` 39,455 의 1.65배). 통제 파일 자신이 `paths:` 를 갖기 때문에 이 확장의 always-loaded 비용은 0 이다.
- **REQ-ALD-014** (Ubiquitous) — 그 통제는 "그 파일을 한 번도 필요로 하지 않는 세션이 지불하는 비용"을 명시적으로 다뤄야 한다. 진술 의무를 지는 대상은 규칙 파일에 한정되지 않고 REQ-ALD-013 이 열거한 가드 슬롯 4개 전부다. 근거: 현재 always-loaded 파일 중 어느 것의 정당화도 이 비용을 다루지 않는다. `session-handoff.md` 가 비용을 이름으로 부른다는 점에서 가장 가깝고, `native-idiom-and-register.md` 는 "영어 세션에 부담 zero"라고 적는데 이는 **동작에 대해서는 참이고 컨텍스트 바이트에 대해서는 거짓**이다.

### 3.4 전 마일스톤 공통

- **REQ-ALD-015** (Ubiquitous) — 이 SPEC의 구속 조건은 **순감소**다. M1은 표면을 줄이고 M2는 늘리므로, 종료 판정은 개별 마일스톤이 아니라 가드 여유의 전후 비교로 한다: 완료 후 여유가 baseline 1,239 토큰보다 **엄격히 커야** 한다.
- **REQ-ALD-016** (Ubiquitous) — 생성·수정된 모든 규칙 파일은 `internal/template/templates/.claude/rules/moai/` 아래에 미러돼야 하고, `make build` 를 돌려야 하며, 미러본은 중립성을 유지해야 한다(SPEC ID, REQ 토큰, 내부 날짜, 커밋 SHA 금지). always-loaded 14개는 이번 세션에 미러 존재를 확인했다.

### 3.5 Out of Scope

#### Out of Scope — 다른 다이어트 후보

- `agent-common-protocol.md`(39,455), `askuser-protocol.md`(30,154), `session-handoff.md`(23,251) 등 나머지 대형 always-loaded 파일의 분리. 사용자가 이번 범위를 `kanban-dispatch.md` 하나로 확정했다.
- `CLAUDE.md`(19,040)와 `.claude/output-styles/moai/moai.md`(65,251) **축소**. 가드 슬롯이지만 규칙 파일이 아니며 별도 소관이다. 범위 밖인 것은 **이번에 바이트를 줄이는 작업**뿐이다 — M3 재발 통제는 이 두 슬롯에도 붙는다(REQ-ALD-013). 통제와 축소는 다른 일이고, 통제의 사각지대는 이번 다이어트가 끝난 뒤 성장이 그 두 슬롯으로 흘러들어도 아무 신호가 나지 않음을 뜻하기 때문이다.
- `MEMORY.md` 정리. 리포 루트에 파일이 없어 현재 슬롯 기여가 0이다.

#### Out of Scope — 가드 자체의 재설계

- char/4 추정을 실제 tokenizer 로 교체하는 것. 가드는 상대 증가 감시용 트립와이어이며, 의존성 추가는 `token_budget_guard.go` 주석이 명시적으로 배제한다.
- `paths:` 글롭 부착 동작의 검증. 이는 Claude Code 런타임 소관이며 이 리포에 구현이 없다.

#### Out of Scope — 신규 always-loaded 파일

- M2가 새 always-loaded 규칙 파일을 만드는 선택지. 편입 대상을 기존 `cache-aware-execution.md` 로 고정한 것은 사용자 확정 사항이다.

#### Out of Scope — 이번에 내리지 않는 결정

- `AlwaysLoadedTokenBudget`(현 75,000)의 하향. 상수는 전 구간에서 유지하고, M4는 착지 후 여유를 측정해 기록만 한다. 그 기록값을 입력으로 삼아 상수를 낮출지는 **별도 백로그 카드**가 판단한다. 순감소 AC("완료 후 여유 > 완료 전 여유")의 기준이 되는 상수를 같은 SPEC 안에서 함께 움직이지 않기 위한 확정 사항이다(plan.md D4-1).
- M3 재발 통제의 Go 변경. 통제는 문서 전용이며 `internal/config` 의 가드 코드와 출력 형식은 손대지 않는다. 가드가 pass/fail 대신 현재 여유를 함께 보고하도록 바꾸는 안(plan.md D1 후보 C)은 채택하지 않는다(plan.md D4-2).

## 4. 사용자 확정 결정 (재논의 금지)

1. **하나의 SPEC, 세 마일스톤, 순감소를 구속 기준으로.** 두 방향(축소/증가)이 한 SPEC 안에서 상쇄되므로 AC는 "완료 후 여유 > 완료 전 여유"다.
2. **다이어트 대상은 `kanban-dispatch.md` 하나.** 다른 후보는 이번에 손대지 않는다.
3. **G3~G7은 `cache-aware-execution.md` 에 압축 HARD 지시로, 근거는 신설 `paths:`-스코프 컴패니언으로.** 새 always-loaded 파일은 만들지 않는다.

## 5. 미검증 항목 (Gaps)

이 SPEC의 근거 중 아래 항목은 **측정되지 않았거나 추정**이다. 덮지 않고 남긴다.

- **토큰 수치는 char/4 추정이다.** 가드 자신의 공식이며 `token_budget_guard.go` 주석이 실제 tokenizer 대비 ±약 15% 오차를 문서화하고 있다. `/context` 는 한 번도 실행하지 않았다.
- **`paths:` 글롭 부착 동작은 Claude Code 런타임 소관이며 이 리포에 구현이 없다.** self-keyed 대 domain-keyed 구분은 글롭 문자열을 읽고 **추론**한 것이지 실제 부착을 관측한 것이 아니다.
- **분리 후 되돌아오는 오버헤드(600~900 bytes)는 `goal-directive.md` 선례에서 뽑은 추정치**이며 측정치가 아니다. 포인터 주석 + 컴패니언 frontmatter 분량을 가리킨다.
- **M2의 캐시 배수와 TTL은 원문 인용이며 여기서 재현하지 않았다** (REQ-ALD-012가 이 사실의 문서 명기를 요구한다).
- **기존 5개 파일의 2026-05-22 시점 크기는 커밋 `09d905bda57ed2537ac98aeefd22e2652719a482` 기준 1회 측정치**다. 그 사이 어떤 커밋이 어떤 몫을 더했는지는 분해하지 않았다.
- **M3 진술 의무의 임계값 1,000 bytes 는 선택한 값이지 측정에서 유도한 값이 아니다.** 위 §2의 성장 수치(기존 5개 파일 합계 +71,347 bytes, 파일별 +6,240 ~ +23,587)에 대고 교정했을 뿐이며, 거기서 나온 "약 70회 발화"도 그 수치를 임계값으로 나눈 추정이다. 실제 발화 빈도는 앞으로의 편집 크기 분포에 달려 있어 관측되지 않았다.
- **1,000 bytes 단일 편집 임계는 그 아래로 반복되는 증분 성장을 통제하지 못한다.** 측정된 +71,347 bytes 는 999 bytes 편집 71회로 **발화 0회** 재현이 가능하다(파일별 23/18/14/8/6회). 위 "약 70회 발화" 추정은 모든 편집이 1,000 bytes **이상**이라는 미관측 가정 위에서만 성립한다. 그리고 §2가 성장의 45.3%를 "기존 파일의 자체 증가"로 분해하는데, 자체 증가야말로 작은 증분으로 도착할 가능성이 가장 높은 모드다. 누적 델타를 보는 2차 트리거를 두지 않는 것은 확정 사항이며(plan.md D4-3), 이 사각지대는 **감수하는 잔여 위험**이다.
- **`CLAUDE.md` 40,000자 규율이 기계적으로 강제되는지는 확인하지 않았다.** `.claude/rules/moai/development/coding-standards.md:25` 는 그 한도를 "any project-local instruction file that also loads in full at every session launch" 로 확장해 적고 있고, 그 기준을 그대로 대면 `moai.md`(65,251)는 25,251 초과다. 다만 같은 줄이 스스로를 "CI-**enforceable** heuristic"(강제 가능한 발견법)이라 부를 뿐 강제된다고 하지 않으며, `internal/config` 에서 40,000 리터럴은 무관한 `DefaultSyncTokens` 로만 나온다. 따라서 이는 **미검증 관찰**이지 작동 중인 통제가 아니다.
