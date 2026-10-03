---
id: SPEC-PREFIX-DIET-001
title: "세션 시작 prefix 다이어트 2단계 — 출력 스타일 본문 축약, 스킬 목록 예산, 에이전트 설명 상한"
version: "0.1.0"
status: draft
created: 2026-10-03
updated: 2026-10-03
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/template/templates/.claude/output-styles/moai, internal/template/templates/.claude/settings.json.tmpl, internal/template/templates/.claude/agents/moai, internal/template"
lifecycle: spec-anchored
tags: "prefix-diet, session-start, output-style, skill-listing-budget, agent-description, first-turn-tokens, template-first"
tier: M
related_specs: [SPEC-ALWAYS-LOADED-BUDGET-001, SPEC-ALWAYS-LOADED-DIET-002, SPEC-ALWAYS-LOADED-HEADROOM-001]
---

# SPEC-PREFIX-DIET-001

## HISTORY

| 날짜 | 버전 | 변경 | 작성자 |
|---|---|---|---|
| 2026-10-03 | 0.1.0 | 최초 작성. 카드 t1450(클래스 C, Tier M). 리더 승인 범위 (b)·(c) 만 담는다. (a) 역할 한정 규칙 주입과 (d) AGENTS.md·규칙 중복 제거는 카드 t1469 소관이라 §F 에 제외로 못 박았다. 기준 트리는 `5d5ff1aae`(= develop `2b9e4a4d0` + 카드 t1449 병합). | manager-spec |

---

## §A. 배경

Claude Code 세션은 사용자 입력 전에 이미 큰 prefix 를 싣는다. 카드 t1450 의 목적은 그중 출력 스타일 본문과 스킬·에이전트 목록 표면을 줄이는 것이다. 이 SPEC 은 그 카드의 (b)·(c) 두 항목만 맡는다.

### A.1 실측 증거 (이 SPEC 의 근거 전부)

측정 명령(단일 실행씩, 기준 트리 = `5d5ff1aae`):

```
claude -p ok --output-format json --model claude-opus-5-5 --settings <disableAllHooks 를 담은 파일>
```

| 조건 | 첫 턴 입력 토큰 | 기준 대비 |
|---|---|---|
| 기준(현 배포 설정: `outputStyle=MoAI-Easy`, `skillListingBudgetFraction=0.02`) | 154,219 | — |
| `outputStyle=default` | 143,186 | −11,033 |
| `skillListingBudgetFraction=0.01` | 147,025 | −7,194 |

해석상 한계를 먼저 적는다.

- `outputStyle=default` 의 −11,033 은 **스타일 본문을 통째로 없앤** 값이다. 이 SPEC 은 본문을 줄일 뿐 없애지 않으므로 이 수치는 (b) 의 **상한**이지 달성값이 아니다. 배포 기본 스타일은 `MoAI-Easy` 한 개이며, 상시 prefix 에 실리는 본문은 선택된 스타일 하나뿐이다. 따라서 `moai.md`·`moai-learn.md` 를 줄여도 기본 설정 사용자의 첫 턴 토큰은 변하지 않는다.
- 위 세 수치는 각각 **한 번 잰 값**이며 실행 간 분산을 모른다. 훅을 끈 측정이므로 SessionStart 훅이 주입하는 맥락은 포함하지 않는다.

### A.2 이 SPEC 자신이 잰 크기 (이 plan 실행, 앵커 `5d5ff1aae`)

| 대상 | UTF-16 코드 단위 | (참고) 코드포인트 | `[HARD]` 줄 수 |
|---|---|---|---|
| `output-styles/moai/moai.md` | 62,593 | 62,470 | 89 |
| `output-styles/moai/moai-easy.md` | 29,243 | 29,181 | 33 |
| `output-styles/moai/moai-learn.md` | 28,517 | 28,418 | 24 |
| 에이전트 12개의 `description:` 블록 합(템플릿) | 11,155 | — | — |
| 에이전트 12개의 `description:` 블록 합(로컬 사본) | 11,146 | — | — |

- 출력 스타일 세 파일은 템플릿과 로컬 사본이 같다(`diff -rq` 무출력). 카드·리더가 인용한 크기(62,470 등)는 코드포인트 기준이라 UTF-16 값과 다르다. 이 SPEC 의 계수 단위는 UTF-16 코드 단위로 고정한다.
- 카드가 말한 에이전트 설명 합계 18.2K 는 **재현되지 않았다**. 이 plan 이 `description:` YAML 블록만 센 값은 11.1K 이며, 18.2K 가 무엇을 더한 수치인지 알 수 없다(Gap). 큰 순서: `manager-lead` 2,182 · `manager-docs` 1,552 · `manager-develop` 1,094 · `manager-spec` 1,022(템플릿).
- 에이전트 본문은 템플릿과 로컬 사본이 12개 중 10개에서 이미 다르다(`diff -rq`). `manager-git`·`manager-spec` 은 `description:` 블록 자체도 다르다(템플릿 553·1,022, 로컬 533·1,033). 이 불일치는 이 SPEC 이 만든 것이 아니며, §D 가 처리 방식을 정한다.

### A.3 소스 오브 트루스 (키별)

| 대상 | 원본 | 로컬 사본과의 관계 |
|---|---|---|
| 출력 스타일 본문 3개 | `internal/template/templates/.claude/output-styles/moai/*.md` | 로컬 `.claude/output-styles/moai/*.md` 가 전체 동일해야 한다(현재 동일) |
| `skillListingBudgetFraction` | `internal/template/templates/.claude/settings.json.tmpl`(414행, 리터럴 `0.02`, 템플릿 표현식 없음) | 로컬 `.claude/settings.json`(410행)은 렌더된 사본이다. 파일 전체는 다르다(`disableClaudeAiConnectors`, `enabledPlugins`, `refreshInterval` 값 등) — **키 단위로만** 대조한다 |
| `outputStyle` | 같은 `settings.json.tmpl` 418행 | 이 SPEC 은 값을 바꾸지 않는다(§F) |
| 에이전트 `description:` | `internal/template/templates/.claude/agents/moai/*.md` | 편집한 에이전트의 `description:` 블록만 로컬과 같게 맞춘다. 본문의 기존 불일치는 그대로 둔다 |

### A.4 선행 관계

- 이 트리는 카드 t1449(`WT-prefix-diet-account-inflow`)의 `settings.json` 변경을 이미 병합했다(`5d5ff1aae`). 이 SPEC 은 그 위에 쌓는다.
- `SPEC-ALWAYS-LOADED-BUDGET-001`(카드 t1469, 브랜치 `WT-always-loaded-budget`)은 아직 이 브랜치에 없다. 그 SPEC 의 구속 원장(구속 줄마다 전/후) 방식을 이 SPEC 이 출력 스타일에 맞게 축소해서 쓴다. 단위 경계 정의는 같은 취지로 §B 에 자급자족으로 적는다.

---

## §B. 용어

| 용어 | 뜻 |
|---|---|
| 앵커 | 이 SPEC 의 전/후 비교 기준 커밋 `5d5ff1aae`. 실행 시작 때 `git rev-parse` 로 같은 SHA 인지 다시 읽는다 |
| 계수 단위 | UTF-16 코드 단위(JavaScript 문자열 길이와 같은 단위) |
| 출력 스타일 본문 | `output-styles/moai/{moai,moai-easy,moai-learn}.md` 의 frontmatter 를 제외한 본문 |
| 단위 | 본문 안의 문단. 빈 줄이 아닌 줄로 시작해 다음 제목 직전, 또는 빈 줄 뒤 줄이 그 단위의 목록 항목·표 행·코드 펜스가 아닌 지점에서 끝난다. 코드 펜스 내부는 불투명하다(안의 `#` 줄·빈 줄·구속 토큰은 단위를 열거나 끊지 않는다). 표는 행 하나가 한 단위다 |
| 구속 토큰 | `[HARD]`, `MUST`, `MUST NOT`, `shall ` |
| 종류 | `binding`(구속 토큰 포함 단위, 연속 줄·하위 항목 포함), `normative`(토큰은 없지만 의무·금지·행동 지시를 담은 단위), `rationale`(이유·배경), `example`(예시·견본) |
| 동결 단위 | 이 SPEC 이 바이트 단위로 건드리지 않는 단위 — §F 와 REQ-PFD-005 |
| 구속 원장 | 동결 단위를 뺀 앵커 본문 단위마다 한 행: 단위 ID, 종류, 출처(파일·제목), 변경 전 텍스트, 변경 후 텍스트, 처리(`verbatim` / `rewrite` / `dropped`), 재작성 메모. `dropped` 는 `rationale`·`example` 에만 허용되며 이유와 생존 단위 참조(같은 정보가 남아 있는 단위)가 붙는다. 테스트 고정물로 `internal/template/testdata/output_style_ledger.json` 에 커밋하며 배포되지 않는다 |
| 첫 턴 입력 토큰 | `claude -p` JSON 결과의 `usage` 안 `input_tokens` + `cache_creation_input_tokens` + `cache_read_input_tokens` 합. 리더의 기준값 154,219 와 같은 정의인지는 M0 에서 앵커 재현으로 확인한다 |

---

## §C. 요구사항 (GEARS)

### C.1 출력 스타일 본문 축약 — 카드 항목 (b)

- **REQ-PFD-001** (Ubiquitous) — The template test suite shall carry a test named `TestOutputStylesCharBudget` that reads each of the three deployed output-style bodies, fails when a body's UTF-16 code-unit length exceeds that file's budget constant, and logs one `output-style=<name> <size>` line per file on every run.
- **REQ-PFD-002** (Event-driven) — When milestone M0 records the per-file reduction targets in `progress.md` §E.2, each budget constant shall be lowered to at most the anchor size minus that file's target in the milestone that shrinks the file, and no budget constant shall ever exceed its file's anchor size.
- **REQ-PFD-003** (Ubiquitous) — The binding ledger shall carry one row per non-frozen unit of the three anchor bodies, and a ledger test shall fail when an anchor unit has no row, when a `binding` or `normative` row has treatment `dropped`, when a `dropped` row lacks a reason or a surviving-unit reference, when a row's kind differs from its anchor kind, or when a `verbatim` or `rewrite` row's after-text is absent from the deployed body.
- **REQ-PFD-004** (Unwanted) — The change shall not drop, weaken, or relocate to a skill, rule, or any other on-demand surface any `binding` or `normative` unit, and a `rewrite` row shall preserve the unit's subject, its obligation strength (`[HARD]`, `MUST`, `MUST NOT`), and every stated exception.
- **REQ-PFD-005** (Unwanted) — The change shall not modify any byte of the frozen handoff units: in `moai.md` the sections titled `### Session Boundary Handoff [HARD]` and `### Session Handoff [HARD]`, and in `moai-easy.md` the section titled `### Banner 7 — Picking Up Next Time (Session Handoff)`; a frozen-unit test shall compare the sections' hashes against a fixture recorded at the anchor.
- **REQ-PFD-006** (Ubiquitous) — Every row of the Localization table in each of the three bodies, including all four locale cells (en, ko, ja, zh), and the Localization Contract's translate-list, keep-verbatim list, anti-pattern, and 46-column banner-width standard shall remain present with unchanged meaning, and a parity test shall compare the table rows with an anchor fixture cell by cell.
- **REQ-PFD-007** (Ubiquitous) — Every content change shall originate under `internal/template/templates/`, be embedded through `make build`, and leave `.claude/output-styles/moai/` byte-identical to its template counterpart within the same SPEC run, with the existing `TestOutputStyles*` tests still passing.
- **REQ-PFD-008** (Unwanted) — The text added or changed in the three bodies shall not carry SPEC IDs, card ids, internal dates, commit hashes, or text that favors one of the 16 supported programming languages, and the count of existing SPEC-ID or card-id mentions in `moai.md` (15 at the anchor) shall not increase.

### C.2 스킬 목록 예산과 에이전트 설명 상한 — 카드 항목 (c)

- **REQ-PFD-009** (Event-driven) — When milestone M1 begins, the value of `skillListingBudgetFraction` shall be fixed by an operator decision recorded in `progress.md` §E.2 (candidates: keep 0.02, set 0.01, or an intermediate value), and where no decision is recorded the key shall be left unchanged; the key's value shall be edited in `settings.json.tmpl` first and mirrored to `.claude/settings.json` at the key level.
- **REQ-PFD-010** (Where the value is lowered) — Where `skillListingBudgetFraction` is set below its anchor value, the run phase shall record before and after, in `progress.md` §E.2, the count of skill names a fresh session lists when asked for them, measured by the protocol in `acceptance.md` §D.2, together with the operator's disposition of any loss.
- **REQ-PFD-011** (Ubiquitous) — The template test suite shall carry a test named `TestAgentDescriptionBudget` that sums the UTF-16 length of the `description:` block of every agent under `internal/template/templates/.claude/agents/moai/`, and fails when the sum exceeds the total constant or any single description exceeds the per-agent cap constant, naming the offending file and size; both constants are recorded in `progress.md` §E.2 at M0 and shall be below the anchor values.
- **REQ-PFD-012** (Unwanted) — A shortened agent description shall not remove the agent's phase or role statement, its invocation trigger, or any of its `NOT for:` clauses (the per-file `NOT for:` count at the anchor is 1 for ten agents, 2 for `manager-spec`, 3 for `super-advisor`), and `TestAgentFrontmatterAudit` shall still pass.

### C.3 측정과 경계

- **REQ-PFD-013** (Ubiquitous) — The run phase shall measure the first-turn input tokens with the identical command of §A.1 run from the worktree root, three runs per condition, first reproducing the anchor baseline (M0) and then measuring after each of M1, the output-style milestones, the agent milestone, and the final tree, recording each run's verbatim `usage` JSON and the summed figure in `progress.md` §E.2.
- **REQ-PFD-014** (Event-driven) — When none of the three M0 anchor reproductions falls within 1% of 154,219, the run phase shall stop before any edit and report the three measured values to the leader.
- **REQ-PFD-015** (Unwanted) — The change shall not edit any path under the exclusions of §F, shall not change the value of the `outputStyle` key, and shall not change hook-injected session context.

---

## §D. 제약

- 계수 단위는 UTF-16 코드 단위다. 리더가 인용한 코드포인트 수와 같은 수로 비교하지 않는다.
- 구속 블록·규범 단위는 축자 유지하거나 의미를 보존하는 압축 재작성만 한다. `rationale`·`example` 만 삭제할 수 있고, 삭제는 생존 단위 참조가 있을 때만 허용한다(예: `moai-easy.md` §10 의 배너 예시가 §7 의 견본과 같은 구조를 되풀이하면 §7 을 생존 단위로 가리킨다).
- 동결 단위의 앞뒤 단위를 고치다가 동결 단위 안으로 번지는 변경을 만들지 않는다. 동결 단위 해시는 그 단위 자체 텍스트만 덮는다.
- 에이전트 본문과 로컬 사본 사이의 기존 불일치는 이 SPEC 이 고치지 않는다. 설명을 줄이는 에이전트는 `description:` 블록만 템플릿과 같게 맞춘다. 로컬이 템플릿보다 새로운 문구를 가졌는지는 M0 에서 `diff` 로 확인하고, 로컬 쪽이 더 새롭다면 그 문구를 템플릿에 먼저 가져온 뒤 줄인다.
- 근거 없는 목표치를 쓰지 않는다. 파일별 축약 목표와 에이전트 설명 상한은 M0 의 원장 하한 측정 뒤 `progress.md` §E.2 에 기록한 값이 구속한다. `plan.md` §B 의 숫자는 초안이다.
- Codex 하네스는 `.claude/output-styles/`·`.claude/settings.json`·`.claude/agents/` 를 쓰지 않고 `AGENTS.md` 를 읽는다(`AGENTS.md` 서두의 하네스 계약). 이 SPEC 은 `AGENTS.md` 를 건드리지 않으므로 영향 밖이라고 읽었다. 실측은 하지 않았다(§H Gaps).

## §E. 수용 기준

인수 조건 전체는 `acceptance.md` 에 있다(AC-PFD-001 ~ AC-PFD-016, Tier M 상한 16 이내; 요구사항도 REQ-PFD-001 ~ REQ-PFD-015 로 상한 16 이내). 각 AC 는 실행 명령, 기대 출력, exit 코드를 갖고, 전/후 첫 턴 토큰 측정 프로토콜은 `acceptance.md` §D.2 에 있다.

---

## §F. 범위 밖

### Out of Scope — 카드 t1469 소관 (post-3.2 보류)

- (a) 역할 한정 규칙 주입 — `kanban-dispatch`, `cross-session-messaging`, `goal-directive`, `moai-mcp-tools` 분할과 SessionStart 훅 주입.
- (d) `AGENTS.md` 본문 편집과 규칙 쪽 중복 제거.
- `.claude/rules/**`(템플릿 `internal/template/templates/.claude/rules/**` 포함), `CLAUDE.md`(템플릿 `internal/template/templates/CLAUDE.md` 포함), `AGENTS.md`·`AGENTS.md.tmpl`·`AGENTS.local.md` 의 어떤 편집도 이 SPEC 에서 하지 않는다.
- 출력 스타일의 Session Handoff 절(`moai.md` §6 의 `Session Boundary Handoff` 와 §8 의 `Session Handoff`, `moai-easy.md` 의 Banner 7). 이 절들은 `workflow/session-handoff.md` 의 렌더 면이며 그 규칙을 카드 t1469 가 다시 쓴다. §6 의 `Session Boundary Handoff` 는 같은 규칙의 5개 트리거 표를 되풀이하므로 §8 과 함께 동결한다(리더 지시는 §8 과 Banner 7 만 명시했고, §6 포함은 이 SPEC 이 보수적으로 넓힌 해석이다 — 리더 확인 대상).

### Out of Scope — 설정과 동작 변경

- 배포 기본 `outputStyle` 값(`MoAI-Easy`) 변경. 상한 측정(`outputStyle=default`)은 증거이지 설계 선택지가 아니다.
- `skillListingBudgetFraction` 외 설정 키, 모델·effort 설정, 훅이 주입하는 SessionStart 맥락.
- 스킬 SKILL.md 본문·frontmatter `description` 축약. 스킬 목록 크기는 예산 키로만 조절한다.
- 에이전트 본문(frontmatter 이후)과 에이전트 목록·도구 구성 변경, 에이전트 본문의 템플릿↔로컬 기존 불일치 해소.
- Localization 표 4개 로케일 문구의 개정과 다른 로케일 추가.

### Out of Scope — 체계

- 구속 의무의 삭제, 그리고 원장 행 없이 이루어지는 재작성·병합.
- v3.2.0 릴리스 배치 결정. 착지 시점은 리더가 정하며, 이 SPEC 은 prompt cache 무효화 때문에 배치 경계 착지를 요구할 뿐이다(`plan.md` §D).

---

## §G. 관련 관계

- `SPEC-ALWAYS-LOADED-BUDGET-001`(카드 t1469, 보류) — 상시 규칙 표면을 맡는다. 이 SPEC 은 출력 스타일·설정·에이전트 설명 표면을 맡아 겹치지 않는다. t1469 가 세션 핸드오프 규칙을 다시 쓰면 §C.1 의 동결 단위 고정물과 출력 스타일의 렌더 절을 그 카드가 함께 갱신한다.
- `SPEC-ALWAYS-LOADED-DIET-002` — 구속 줄 이동·재작성 동결(REQ-ALD2-002·003). 이 SPEC 의 압축 재작성은 출력 스타일 본문에 한정되고 구속 원장으로 의미 보존을 감사한다는 점에서 `SPEC-ALWAYS-LOADED-BUDGET-001` 의 운영자 판정 방식을 따른다. 출력 스타일에 대한 별도 운영자 판정이 있었는지는 확인하지 못했다(§H Gaps — plan-audit 의 쟁점이 될 수 있다).

## §H. Gaps (미검증, 정직하게)

- §A.1 의 세 토큰 수치는 단일 실행이며 분산을 모른다.
- 에이전트 설명 축약의 토큰 효과는 측정한 적이 없다. 설명 합 11.1K 는 대략 3K 토큰 안팎이라 상한이 낮다(토큰 환산은 추정이며 측정 아님).
- `skillListingBudgetFraction=0.01` 이 스킬 발견 품질에 주는 영향은 측정한 적이 없다.
- 카드의 에이전트 설명 합계 18.2K 는 재현하지 못했다.
- 출력 스타일 압축에 대해 운영자가 구속 줄 재작성을 허용했는지(`SPEC-ALWAYS-LOADED-BUDGET-001` 의 Q1 은 그 SPEC 한정)는 이 plan 에서 확인하지 못했다.
- Codex 등 다른 하네스가 영향 밖이라는 판단은 `AGENTS.md` 계약 문면을 읽은 것이며 측정이 아니다.
