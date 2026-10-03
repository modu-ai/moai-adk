---
id: SPEC-PREFIX-DIET-001
title: "세션 시작 prefix 다이어트 2단계 — 출력 스타일 파일 축약과 에이전트 설명 상한"
version: "0.5.0"
status: in-progress
created: 2026-10-03
updated: 2026-10-03
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/template/templates/.claude/output-styles/moai, internal/template/templates/.claude/agents/moai, internal/template"
lifecycle: spec-anchored
tags: "prefix-diet, session-start, output-style, agent-description, first-turn-tokens, template-first, binding-ledger"
tier: M
amendment_of: SPEC-PREFIX-DIET-001
related_specs: [SPEC-ALWAYS-LOADED-BUDGET-001, SPEC-ALWAYS-LOADED-DIET-002, SPEC-ALWAYS-LOADED-HEADROOM-001]
---

# SPEC-PREFIX-DIET-001

## HISTORY

| 날짜 | 버전 | 변경 | 작성자 |
|---|---|---|---|
| 2026-10-03 | 0.5.0 | **두 번째 in-place amendment(`completed → in-progress`, `amendment_of` 유지, 구조화 기록은 `## Amendments` 두 번째 항목; 카드 t1450, 리더 지시).** 실행 단계 repair-2(커밋 76e480974 + adf727ffb)가 codex 교차 모델 이의 뒤 리더 명령으로 moai-easy 단위 셋(0073, 0163, 0181)을 verbatim 복원하고 남은 `dropped` 행의 자유 서술 `survivor` 필드를 비웠다. 크기: moai.md 61,362 / moai-easy.md 23,586(이전 23,036) / moai-learn.md 28,517. `dropped` 49행 합 6,888 UTF-16(이전 52행 / 7,438), `rewrite` 행 0. 첫 턴 토큰 154,235 → 151,873(−2,362, −1.53%, 깨끗한 트리 3회; 앞선 −2,562 / −1.66% 는 대체됨). 원장 테스트: `survivor_file` 또는 `survivor_anchor` 가 비면 `DROPPED_SURVIVOR` 가 발화하고 `SURVIVOR_UNRESOLVED` 는 그대로. REQ-PFD-002·003 과 §B 원장 필드 문면 개정, 번호 변경 없음, REQ 13 · AC 16. | manager-spec |
| 2026-10-03 | 0.4.0 | **In-place amendment of the `completed` SPEC (카드 t1450, 리더 결정; `completed → in-progress`, `amendment_of` 자기참조, 구조화 기록은 `## Amendments`).** sync-audit FAIL 81.7(`.moai/reports/t1450/sync-audit.md`)이 사용자 지시의 유일 운반자였던 `dropped` 단위 둘(`moai-easy-0183`, `moai-learn-0122`)을 찾았고 원장 테스트가 비어 있지 않은 `survivor` 문자열이면 무엇이든 받았음이 드러나, 실행 단계가 수리했다. 엄격 재검 77행 중 25개 단위 verbatim 복원, `dropped` 52행(7,438 UTF-16). REQ-PFD-003 은 `survivor_file`+`survivor_anchor` 해석 규칙(`SURVIVOR_UNRESOLVED`)을 갖도록 개정(새 REQ 없음). REQ-PFD-002 는 수리 뒤 크기·결과로 개정(moai.md 61,362 / moai-easy.md 23,036 / moai-learn.md 28,517, 첫 턴 154,235 → 151,673, −2,562, −1.66%; 앞선 −3,178/−2.06% 는 수리 전 값이라 대체됨). `surface_guard.py` 는 실행 단계 점검기라는 범위 문구와 `plan.md` 초안 목표 대체 표기 반영. 번호 변경 없음, REQ 13 · AC 16. | manager-spec |
| 2026-10-03 | 0.3.1 | 실행 단계가 `progress.md` §E.2 에 남긴 SPEC 문면 공백 4건을 닫는 사후 수정(카드 t1450, 리더 지시, status 는 `in-progress` 그대로라 `## Amendments` 절은 두지 않고 버전·HISTORY 만 올림). D1 REQ-PFD-013 의 허용목록과 AC-PFD-012 양성 대조 목록에 `catalog-hashes` 표면 추가 — `surface_guard.py` 구현 그대로(편집한 에이전트의 `hash:` 줄만, 출력 스타일은 무변경). D2 AC-PFD-015 가 어느 REQ 에도 묶이지 않는 이유를 §E 에 명시(문서 위생 게이트, 동작 요구 아님). D3 §E 의 AC 개수 오기(015)를 016 으로 정정. REQ-PFD-002 파일별 목표를 실행 때 잰 `rationale`+`example` 합계로 낮춘 값(moai-easy 21,350 / moai-learn 27,010 / moai 61,149, 파일 전체 UTF-16)으로 문면에 반영하고 실측 결과(첫 턴 154,235 → 151,057, −3,178, −2.06%)를 기록. 요구사항·AC 번호 변경 없음, REQ 13 · AC 16. | manager-spec |
| 2026-10-03 | 0.3.0 | plan-audit 2회차 FAIL 0.88(MP-8, `.moai/reports/t1450/plan-audit-iter2.md`) 수리(리더가 3회차 델타를 허용). D1 AC-PFD-004·012 의 RED-now 칸이 초록 조건과 같은 출력이었으므로 두 AC 를 **RG 로 재분류**하고 릴리스 차단에서 뺐다. D2 점검기에서 `skillListingBudgetFraction` 키 예외를 없앴다 — 설정 파일 두 개를 허용목록에서 제거해 **어떤 설정 변경도 위반**이다. 점검기의 ruff 지적(E401·E741)을 동작 변경 없이 수리하고 PASS/FAIL 출력을 재관측했다. 에이전트 설명을 줄이면 `make build` 의 `agents-emit-check` 가 golden 해시 불일치로 실패함을 직접 재현(관측)했고, M5 에 `make agents-emit` 단계와 생성 TOML(`internal/template/templates/.codex/agents/moai/*.toml`)의 허용목록 등재(REQ-PFD-011·013)를 넣었다. AC-PFD-016(agents-emit-check 회귀 가드) 추가. REQ 13 · AC 16. | manager-spec |
| 2026-10-03 | 0.2.0 | plan-audit 1회차 FAIL 0.78(`.moai/reports/t1450/plan-audit-iter1.md`)과 리더 판정(미션 계약 `11c79e1a`) 반영. D1 계수 단위를 출력 스타일 **파일 전체 UTF-16** 하나로 통일(REQ-PFD-001·002, §A.2). D2 리더 결정: `skillListingBudgetFraction` 은 **변경하지 않는다** — 키는 `0.02` 그대로 두고 가드하며, `0.01` 측정은 운영자 결정 항목으로만 기록한다(항목 (c) 는 에이전트 설명 상한뿐). D3 리더 결정: `[HARD]` 줄은 `dropped`/`verbatim` 만 — 압축 재작성 없음, 원장에 `rewrite` 행 0(REQ-PFD-003). D4 단위별 구속 토큰 개수 검사와 추출기 명세를 이 SPEC 에 둔다(§B, REQ-PFD-004). D5 제외를 경로 허용목록 점검기(`surface_guard.py`)와 표면별 명령으로 강제(REQ-PFD-013). D6 §D.3 lint 기록 정정. D7 AC-PFD-007 이 날짜·해시·언어 편향까지 덮음. D8 `SPEC-ALWAYS-LOADED-BUDGET-001` 은 이 트리에 없어 `git show WT-always-loaded-budget:<경로>` 로만 인용. 카드의 에이전트 설명 합계 18.2K 를 실측 11,155(템플릿)/11,146(로컬)로 정정. REQ 13 · AC 15 로 정리. | manager-spec |
| 2026-10-03 | 0.1.0 | 최초 작성. 카드 t1450(클래스 C, Tier M). 기준 트리 `5d5ff1aae`(= develop `2b9e4a4d0` + 카드 t1449 병합). | manager-spec |

---

## §A. 배경

Claude Code 세션은 사용자 입력 전에 이미 큰 prefix 를 싣는다. 카드 t1450 은 그중 출력 스타일 파일과 에이전트 설명 표면을 줄이는 일을 맡는다. 이 SPEC 은 그 카드의 (b) 출력 스타일 축약과 (c) 에이전트 설명 상한만 다룬다. (c) 의 스킬 목록 예산 값 변경은 리더가 일 항목에서 뺐다(§A.1).

### A.1 실측 증거 (이 SPEC 의 근거 전부)

측정 명령(단일 실행씩, 기준 트리 = `5d5ff1aae`):

```
claude -p ok --output-format json --model claude-opus-5-5 --settings <disableAllHooks 를 담은 파일>
```

| 조건 | 첫 턴 입력 토큰 | 기준 대비 | 이 SPEC 에서의 지위 |
|---|---|---|---|
| 기준(현 배포 설정: `outputStyle=MoAI-Easy`, `skillListingBudgetFraction=0.02`) | 154,219 | — | 기준 |
| `outputStyle=default` | 143,186 | −11,033 | (b) 의 **상한**, 달성값 아님 |
| `skillListingBudgetFraction=0.01` | 147,025 | −7,194 (−4.7%) | **운영자 결정 항목**, 작업 아님 |

- `outputStyle=default` 의 −11,033 은 스타일 본문을 통째로 없앤 값이다. 이 SPEC 은 줄일 뿐 없애지 않는다. 배포 기본 스타일은 `MoAI-Easy` 한 개이며 prefix 에 실리는 파일은 선택된 스타일 하나뿐이라, `moai.md`·`moai-learn.md` 축약은 기본 설정 사용자의 첫 턴 토큰을 바꾸지 않는다.
- `skillListingBudgetFraction=0.01` 은 스킬 목록을 잘라 낸다. 값을 바꾸는 것은 스킬 발견 동작을 바꾸는 운영자의 판단이므로 이 SPEC 은 키를 `0.02` 로 **그대로 두고** 지킨다(REQ-PFD-009). 측정치는 운영자가 나중에 결정할 때 쓰도록 여기에 남길 뿐이다.
- 세 수치는 각각 한 번 잰 값이라 실행 간 분산을 모른다. 훅을 끈 측정이라 SessionStart 훅 주입 맥락은 포함하지 않는다.

### A.2 이 SPEC 자신이 잰 크기 (앵커 `5d5ff1aae`)

**계수 단위는 출력 스타일 파일 전체의 UTF-16 코드 단위다.** 이유: 런타임이 싣는 것이 파일이고, REQ-PFD-013 의 점검기가 frontmatter 변경을 막으므로 파일 변화량이 곧 본문 변화량이라 한 단위로 충분하다. 본문만 세는 단위를 따로 두면 frontmatter 파서가 테스트에 들어가고, 한 SPEC 안에 두 단위가 생긴다.

| 대상 | 파일 전체 UTF-16 | (참고) 코드포인트 | `[HARD]` 줄 수 |
|---|---|---|---|
| `output-styles/moai/moai.md` | 62,593 | 62,470 | 89 |
| `output-styles/moai/moai-easy.md` | 29,243 | 29,181 | 33 |
| `output-styles/moai/moai-learn.md` | 28,517 | 28,418 | 24 |
| 에이전트 12개 `description:` 블록 합(템플릿) | 11,155 | — | — |
| 에이전트 12개 `description:` 블록 합(로컬 사본) | 11,146 | — | — |

- 카드·리더가 인용한 62,470 등은 코드포인트 기준이다. plan-audit 가 잰 본문만의 값(62,219 / 28,740 / 28,123)과도 다르다 — 이 SPEC 의 상수·AC·REQ 는 위 표의 파일 전체 값 하나만 쓴다.
- **카드의 에이전트 설명 합계 18.2K 는 실측 11,155(템플릿)/11,146(로컬)로 정정한다.** 18.2K 가 무엇을 더한 수치인지는 재현되지 않았다. 큰 순서(템플릿): `manager-lead` 2,182 · `manager-docs` 1,552 · `manager-develop` 1,094 · `manager-spec` 1,022.
- 에이전트 본문은 템플릿과 로컬 사본이 12개 중 10개에서 이미 다르다. `manager-git`·`manager-spec` 은 `description:` 블록도 다르다(템플릿 553·1,022, 로컬 533·1,033). 이 불일치는 이 SPEC 이 만든 것이 아니며 §D 가 처리 방식을 정한다.

### A.3 소스 오브 트루스 (키별)

| 대상 | 원본 | 로컬 사본과의 관계 |
|---|---|---|
| 출력 스타일 파일 3개 | `internal/template/templates/.claude/output-styles/moai/*.md` | 로컬 `.claude/output-styles/moai/*.md` 가 파일 전체 동일해야 한다(현재 동일) |
| `skillListingBudgetFraction` | `internal/template/templates/.claude/settings.json.tmpl`(414행, 리터럴 `0.02`) | 로컬 `.claude/settings.json`(410행)은 렌더된 사본이며 파일 전체는 다르다 — **키 단위로만** 대조한다. 이 SPEC 은 두 파일을 수정하지 않는다 |
| `outputStyle` | 같은 `settings.json.tmpl` 418행 | 값 `MoAI-Easy` 불변(REQ-PFD-013) |
| 에이전트 `description:` | `internal/template/templates/.claude/agents/moai/*.md` | 편집한 에이전트의 `description:` 블록만 로컬과 같게 맞춘다. 본문의 기존 불일치는 그대로 둔다 |

### A.4 선행 관계

- 이 트리는 카드 t1449 의 설정 변경을 이미 병합했다(`5d5ff1aae`). 이 SPEC 은 그 위에 쌓는다.
- `SPEC-ALWAYS-LOADED-BUDGET-001`(카드 t1469) 은 **이 트리에 없다**. 브랜치 `WT-always-loaded-budget` 에만 있고, 이 문서는 그 SPEC 을 `git show WT-always-loaded-budget:.moai/specs/SPEC-ALWAYS-LOADED-BUDGET-001/spec.md` 로만 인용한다. 구속 원장 방식은 그 SPEC 에서 빌렸지만 단위 추출기·계수 규칙은 의존하지 않고 이 SPEC 에 직접 정의한다(§B).

---

## §B. 용어와 추출기 명세

| 용어 | 뜻 |
|---|---|
| 앵커 | 전/후 비교 기준 커밋 `5d5ff1aae`. 실행 시작 때 `git rev-parse` 로 재확인한다 |
| 계수 단위 | 출력 스타일 파일 전체의 UTF-16 코드 단위(JavaScript 문자열 길이와 같은 단위) |
| 출력 스타일 본문 | 두 번째 `---` 이후. 원장 단위는 본문에서만 뽑는다. frontmatter 는 편집하지 않는다(REQ-PFD-013) |
| 단위 | 본문 안의 문단. 빈 줄이 아닌 줄로 시작해 다음 제목(수준 무관) 직전, 또는 빈 줄 뒤 줄이 그 단위의 목록 항목·표 행·코드 펜스가 아닌 지점에서 끝난다. 그 단위가 바로 이어서 여는 목록·표·코드 펜스는 빈 줄 하나까지 사이에 두고 그 단위에 속한다. 코드 펜스 내부는 불투명하다(안의 `#` 줄·빈 줄·구속 토큰은 단위를 열거나 끊지 않는다). 표는 행 하나가 한 단위다 |
| 구속 토큰 | `[HARD]`, `MUST NOT`, `MUST`, `shall `. 대소문자 구분, 겹치지 않게 센다 — `MUST NOT` 을 먼저 세고 `MUST` 는 바로 뒤가 ` NOT` 이 아닌 경우만 센다. 코드 펜스 안의 토큰도 그 단위의 토큰이다 |
| 종류 | `binding`(구속 토큰을 **하나라도** 가진 단위 — 이 분류는 추출기가 기계적으로 정하며 작성자가 바꿀 수 없다), `normative`(토큰은 없지만 의무·금지·행동 지시를 담은 단위, 작성자 분류), `rationale`(이유·배경), `example`(예시·견본) |
| 동결 단위 | 이 SPEC 이 바이트 단위로 건드리지 않는 단위 — REQ-PFD-005 |
| 구속 원장 | 동결 단위를 뺀 앵커 본문 단위마다 한 행. 필드: `id`, `kind`, `source`(파일·제목), `before_text`, `after_text`, `treatment`(`verbatim` / `dropped` 둘뿐), `survivor_file`+`survivor_anchor`(`dropped` 에만: 같은 정보가 남은 단위가 있는 파일과, 그 파일에 실제로 들어 있는 앵커 문자열 — 테스트가 해석한다; 자유 서술 `survivor` 필드는 없고, 두 포인터 중 하나라도 비면 `DROPPED_SURVIVOR`), `note`. 머리에 `anchor` SHA. 테스트 고정물로 `internal/template/testdata/output_style_ledger.json`(최상위 객체의 `rows` 배열)에 커밋하며 배포되지 않는다 |
| 첫 턴 입력 토큰 | `claude -p` JSON 결과의 `usage` 안 `input_tokens` + `cache_creation_input_tokens` + `cache_read_input_tokens` 합. 리더의 154,219 와 같은 정의인지는 M0 앵커 재현으로 확인한다 |

---

## §C. 요구사항 (GEARS)

### C.1 출력 스타일 파일 축약 — 카드 항목 (b)

- **REQ-PFD-001** (Ubiquitous) — The template test suite shall carry a test named `TestOutputStylesCharBudget` that reads each of the three deployed output-style files, fails when a file's whole-file UTF-16 code-unit length exceeds that file's budget constant, and logs one `output-style=<name> <size>` line per file on every run.
- **REQ-PFD-002** (Event-driven) — When milestone M0 records the per-file reduction targets in `progress.md` §E.2, each budget constant shall be lowered to at most the anchor whole-file size minus that file's target in the milestone that shrinks the file, no constant shall ever exceed its file's anchor size, and when a file's droppable total (the UTF-16 length of its `rationale` and `example` units, measured in M0) is smaller than the target reduction, the target shall be lowered to that total and reported to the leader instead of being met by rewriting. The run phase first lowered the targets to the measured droppable totals (the plan drafts of 21,000 / 20,000 / 45,000 were not achievable without rewriting `[HARD]` lines, which the leader forbade, D3). After sync-audit FAIL 81.7 the first repair restored 25 of 77 dropped units, so the current budget constants equal the repaired whole-file UTF-16 sizes: moai.md 61,362, moai-easy.md 23,586, moai-learn.md 28,517 (moai-easy.md was 23,036 after the first repair; before repair 61,149 / 21,350 / 27,010; anchor 62,593 / 29,243 / 28,517), with 49 `dropped` rows totalling 6,888 UTF-16 units (was 52 / 7,438) and zero `rewrite` rows; repair-2 restored moai-easy units 0073, 0163 and 0181 verbatim. Measured outcome (REQ-PFD-012 protocol, three clean-tree runs): first-turn input tokens 154,235 → 151,873 (−2,362, −1.53%); the earlier −2,562 / −1.66% (first repair) and −3,178 / −2.06% (before any repair) are superseded.
- **REQ-PFD-003** (Ubiquitous) — The binding ledger shall carry one row per non-frozen unit of the three anchor bodies with treatment `verbatim` or `dropped` only, shall contain zero `rewrite` rows, and a ledger test shall fail when an anchor unit has no row, when any row's treatment is anything other than `verbatim` or `dropped`, when a unit that the extractor classifies `binding` is labeled any other kind, when a `binding` or `normative` row is `dropped`, when a `dropped` row lacks a reason in `note`, when a `dropped` row has an empty `survivor_file` or `survivor_anchor` (`DROPPED_SURVIVOR`; the free-text `survivor` field no longer exists on dropped rows) or has a pair that does not resolve (the file is readable and contains the anchor; the test emits `SURVIVOR_UNRESOLVED` otherwise, and mutation subtests with a missing anchor and a missing file shall fail), or when a `verbatim` row's after-text differs from its before-text or is absent from the deployed file.
- **REQ-PFD-004** (Ubiquitous) — The ledger test shall also compare the extractor's binding-token counts (§B), per unit and per file: for every `verbatim` row the after-text count shall equal the before-text count for each token, a `dropped` row's before-text shall carry zero binding tokens, and the per-file totals after the change shall equal the anchor totals, so that no binding token is lost; no unit of any kind shall be relocated to a skill, rule, companion, or other on-demand surface.
- **REQ-PFD-005** (Unwanted) — The change shall not modify any byte of the frozen handoff units — in `moai.md` the sections titled `### Session Boundary Handoff [HARD]` and `### Session Handoff [HARD]`, in `moai-easy.md` the section titled `### Banner 7 — Picking Up Next Time (Session Handoff)` — and a frozen-unit test shall compare the sections' hashes against a fixture recorded at the anchor.
- **REQ-PFD-006** (Ubiquitous) — Every row of the Localization table in each of the three files, including all four locale cells (en, ko, ja, zh), and the Localization Contract's translate-list, keep-verbatim list, anti-pattern, and 46-column banner-width standard shall remain present with unchanged text, and a parity test shall compare the table rows with an anchor fixture cell by cell.
- **REQ-PFD-007** (Ubiquitous) — Every content change shall originate under `internal/template/templates/`, be embedded through `make build`, and leave `.claude/output-styles/moai/` byte-identical to its template counterpart within the same SPEC run, with the existing `TestOutputStyles*` tests still passing.
- **REQ-PFD-008** (Unwanted) — The three output-style files shall not gain SPEC IDs, card ids, internal dates, commit hashes, or language-specific text beyond the anchor counts (SPEC-ID or card-id mentions `moai.md` 15 / `moai-easy.md` 0 / `moai-learn.md` 0; dates 0 / 0 / 1; hex hashes 0 / 0 / 0; named-language mentions 3 / 0 / 1).

### C.2 에이전트 설명 상한과 스킬 목록 키 가드 — 카드 항목 (c)

- **REQ-PFD-009** (Unwanted) — The change shall not alter the value of `skillListingBudgetFraction` (`0.02` in `settings.json.tmpl` and in `.claude/settings.json`) nor modify either settings file; the `0.01` measurement of §A.1 is recorded as an operator decision item only.
- **REQ-PFD-010** (Ubiquitous) — The template test suite shall carry a test named `TestAgentDescriptionBudget` that sums the UTF-16 length of the `description:` block of every agent under `internal/template/templates/.claude/agents/moai/` and fails when the sum exceeds the total constant or any single description exceeds the per-agent cap constant, naming the offending file and size; both constants are recorded in `progress.md` §E.2 at M0 and shall be below the anchor values (sum 11,155; largest 2,182).
- **REQ-PFD-011** (Unwanted) — A shortened agent description shall not remove the agent's phase or role statement, its invocation trigger, or any of its `NOT for:` clauses (anchor per-file counts: 1 for ten agents, 2 for `manager-spec`, 3 for `super-advisor`), `TestAgentFrontmatterAudit` shall still pass, and every edited agent's generated Codex TOML under `internal/template/templates/.codex/agents/moai/` shall be regenerated with `make agents-emit` so that `agents-emit-check` (part of `make build`) passes.

### C.3 측정과 경계

- **REQ-PFD-012** (Ubiquitous) — The run phase shall measure the first-turn input tokens with the identical command of §A.1 run from the worktree root, three runs per condition, first reproducing the anchor baseline (M0) and then measuring after each milestone that changed a file and on the final tree, recording each run's verbatim `usage` JSON and the summed figure in `progress.md` §E.2, and when none of the three M0 anchor reproductions falls within 1% of 154,219 the run phase shall stop before any edit and report the three values to the leader.
- **REQ-PFD-013** (Unwanted) — The change shall not touch any path outside the allowlist of `surface_guard.py` — the three output-style files and their local mirrors (frontmatter unchanged), the twelve agent definitions and their local mirrors (body and every frontmatter field other than `description:` unchanged), the generated Codex agent TOMLs under `internal/template/templates/.codex/agents/moai/`, this SPEC's tests and fixtures, this SPEC's directory and `.moai/reports/t1450/`, and the `catalog-hashes` surface — `internal/template/catalog.yaml`, where `make build` rewrites exactly one `hash:` line (a 64-hex value) per edited agent template and nothing for output styles, and the guard accepts only changed `hash:` lines whose entry `path:` is an agent template `.md` under `.claude/agents/moai/` that is itself changed in the same diff, reporting `VIOLATION catalog-change-beyond-edited-agent-hashes` for any other changed line — and therefore shall not change either settings file (the guard allowlists neither, so any settings change is a violation, the `skillListingBudgetFraction` line included), the `outputStyle` value, any `SKILL.md`, any rule, `CLAUDE.md`, `AGENTS*`, or hook code; the guard shall exit 0 with `surface-guard=PASS` on the run-phase tree and exit 1 on a positive-control mutant. The guard is a run-phase guard, hand-run (no CI runs it): it has no end-ref option (it diffs the working tree and untracked files against one base SHA), so the sync-phase deliverables `CHANGELOG.md` and the SPEC artifacts are outside its scope and its `VIOLATION outside-allowlist CHANGELOG.md` after the sync commit is expected; to re-run it on the run range, use a worktree checked out at `d5e4536a5` with `python3 .moai/specs/SPEC-PREFIX-DIET-001/surface_guard.py 5d5ff1aae`.

---

## §D. 제약

- 계수 단위는 출력 스타일 파일 전체의 UTF-16 코드 단위 하나다. 리더가 인용한 코드포인트 수와 같은 수로 비교하지 않는다.
- **구속 줄은 재작성하지 않는다**(리더 결정 D3: 의미 보존을 기계로 검증할 수 없고, `SPEC-ALWAYS-LOADED-BUDGET-001` 의 운영자 판정은 그 SPEC 한정이었다). 축약은 `rationale`·`example` 단위의 `dropped` 로만 한다. `binding` 단위는 `verbatim` 으로만 남는다. 삭제는 생존 단위 참조가 있을 때만 허용한다(예: `moai-easy.md` §10 의 배너 예시가 §7 의 견본과 같은 구조를 되풀이하면 §7 을 생존 단위로 가리킨다).
- `rationale`·`example` 이라는 분류는 작성자가 한다. 작성자가 `normative` 단위를 `rationale` 로 분류해 지우는 오분류는 기계가 못 잡는다 — `dropped` 행 전부를 `progress.md` §E.2 에 목록으로 반출해 sync-audit 와 리더가 훑는다(잔여 위험으로 남는다).
- 동결 단위의 앞뒤 단위를 고치다가 동결 단위 안으로 번지는 변경을 만들지 않는다. 동결 단위 해시는 그 단위 자체 텍스트만 덮는다.
- 에이전트 본문과 로컬 사본 사이의 기존 불일치는 이 SPEC 이 고치지 않는다. 설명을 줄이는 에이전트는 `description:` 블록만 템플릿과 같게 맞춘다. 로컬 쪽이 더 새롭다면 그 문구를 템플릿에 먼저 가져온 뒤 줄인다.
- 근거 없는 목표치를 쓰지 않는다. 파일별 축약 목표와 에이전트 설명 상한은 M0 의 측정 뒤 `progress.md` §E.2 에 기록한 값이 구속한다. `plan.md` §B 의 숫자는 초안이다.
- 범위 판정(점검기·`git diff`)은 병합 전 평가 전용이다. develop 을 흡수하면 기준 SHA 를 읽는 시점에 `git merge-base develop HEAD` 로 다시 구해 기록한다(리터럴 핀 금지 — `.claude/rules/local/gitflow-lane-protocol.md` §8). 병합 뒤에는 쓰지 않는다.
- Codex 하네스는 `.claude/output-styles/`·`.claude/settings.json`·`.claude/agents/` 를 쓰지 않고 `AGENTS.md` 를 읽는다(`AGENTS.md` 서두의 하네스 계약). 이 SPEC 은 `AGENTS.md` 를 건드리지 않으므로 영향 밖이라고 읽었다. 실측은 하지 않았다(§H).

## §E. 수용 기준

인수 조건 전체는 `acceptance.md` 에 있다(AC-PFD-001 ~ AC-PFD-016, Tier M 상한 16 이내; 요구사항도 REQ-PFD-001 ~ REQ-PFD-013 으로 상한 16 이내). AC-PFD-015 는 의도적으로 어느 REQ 에도 묶이지 않는다 — SPEC 문서 자체의 위생 게이트(lint 오류 0, `progress.md` §E 4절)라서 관측 가능한 동작 요구가 아니며, 요구를 지어내 억지로 매핑하면 거짓 추적성이 생긴다(`acceptance.md` 의 REQ 칸이 `(plan 단계 산출물)` 로 표기). 각 AC 는 실행 명령, 기대 출력, exit 코드를 갖고, 전/후 첫 턴 토큰 측정 프로토콜은 `acceptance.md` §D.2 에 있다.

---

## §F. 범위 밖

### Out of Scope — 카드 t1469 소관 (post-3.2 보류)

- (a) 역할 한정 규칙 주입 — `kanban-dispatch`, `cross-session-messaging`, `goal-directive`, `moai-mcp-tools` 분할과 SessionStart 훅 주입.
- (d) `AGENTS.md` 본문 편집과 규칙 쪽 중복 제거.
- `.claude/rules/**`(템플릿 `internal/template/templates/.claude/rules/**` 포함), `CLAUDE.md`(템플릿 `internal/template/templates/CLAUDE.md` 포함), `AGENTS.md`·`AGENTS.md.tmpl`·`AGENTS.local.md` 의 어떤 편집도 이 SPEC 에서 하지 않는다.
- 출력 스타일의 Session Handoff 절(`moai.md` §6 `Session Boundary Handoff` 와 §8 `Session Handoff`, `moai-easy.md` 의 Banner 7). 리더가 §6 포함을 확정했다. 이 절들은 `workflow/session-handoff.md` 의 렌더 면이며 그 규칙을 카드 t1469 가 다시 쓴다.

### Out of Scope — 설정과 동작 변경

- `skillListingBudgetFraction` 값 변경과 스킬 목록 예산 조정 일체(리더 결정 D1 — 스킬 목록이 잘리는 것은 운영자의 판단). 스킬 발견 품질 측정도 하지 않는다.
- 배포 기본 `outputStyle` 값(`MoAI-Easy`) 변경과 `settings.json.tmpl`·`.claude/settings.json` 의 모든 수정, 모델·effort 설정, 훅이 주입하는 SessionStart 맥락.
- 스킬 SKILL.md 본문·frontmatter `description` 축약.
- 에이전트 본문(frontmatter 이후)·`description:` 외 frontmatter 필드·에이전트 목록·도구 구성 변경, 에이전트 본문의 템플릿↔로컬 기존 불일치 해소.
- Localization 표 4개 로케일 문구의 개정과 다른 로케일 추가.
- `[HARD]` 줄의 압축 재작성(리더 결정 D3).

### Out of Scope — 체계

- 구속 의무의 삭제, 그리고 원장 행 없이 이루어지는 축약.
- v3.2.0 릴리스 배치 결정. 착지 시점은 리더가 정하며, 이 SPEC 은 prompt cache 무효화 때문에 배치 경계 착지를 요구할 뿐이다(`plan.md` §D).

---

## §G. 관련 관계

- `SPEC-ALWAYS-LOADED-BUDGET-001`(카드 t1469, 보류, **이 트리에 없음**) — 상시 규칙 표면을 맡는다. 이 SPEC 은 출력 스타일·에이전트 설명 표면을 맡아 겹치지 않는다. 인용 경로: `git show WT-always-loaded-budget:.moai/specs/SPEC-ALWAYS-LOADED-BUDGET-001/spec.md`. t1469 가 세션 핸드오프 규칙을 다시 쓰면 동결 단위 고정물과 출력 스타일 렌더 절을 그 카드가 함께 갱신한다.
- `SPEC-ALWAYS-LOADED-DIET-002` — 구속 줄 이동·재작성 동결(REQ-ALD2-002·003). 이 SPEC 은 구속 줄을 재작성하지도 옮기지도 않으므로 그 동결과 충돌하지 않는다.

## §H. Gaps (미검증, 정직하게)

- 생존 포인터는 앵커가 그 파일에 **있음**을 증명할 뿐 그 앵커가 떨어뜨린 지시를 **담고 있음**은 증명하지 않는다. 남은 `dropped` 52행은 검토자 판단에 기댄다(잔여 위험으로 기록).

- §A.1 의 세 토큰 수치는 단일 실행이며 분산을 모른다.
- 에이전트 설명 축약의 토큰 효과는 측정한 적이 없다. 설명 합 11.1K 는 대략 3K 토큰 안팎이라 상한이 낮다(환산은 추정).
- `skillListingBudgetFraction=0.01` 이 스킬 발견 품질에 주는 영향은 측정하지 않았다(운영자 결정 항목).
- `rationale`/`example` 오분류(`normative` 단위를 지우는 경우)는 기계가 못 잡는다 — 반출 목록의 사람 검토로만 막는다.
- 구속 토큰 개수 검사는 토큰 손실만 잡는다. 주체·예외 조항의 손실은 `[HARD]` 줄을 재작성하지 않으므로(`verbatim` 만) 발생하지 않는다는 구조적 이유로 막는다.
- Codex 등 다른 하네스가 영향 밖이라는 판단은 `AGENTS.md` 계약 문면을 읽은 것이며 측정이 아니다.
- 점검기의 양성 대조를 이 plan 에서 관측했다(모두 관측 뒤 `git checkout --` 로 복원, `git status --short` 로 확인) — (a) `docs` 표면 금지 시 exit 1, (b) 에이전트 본문 한 줄 덧붙임 → `VIOLATION agent-body-or-nondescription-frontmatter-changed`, exit 1, (c) `.claude/settings.json` 의 `skillListingBudgetFraction` 0.02→0.01 → `VIOLATION outside-allowlist .claude/settings.json`, exit 1, (d) 출력 스타일 frontmatter `name:` 변이 → `VIOLATION frontmatter-changed`, exit 1. 허용목록 밖 새 파일 변이는 별도로 관측하지 않았다((c) 가 같은 `outside-allowlist` 분기를 탄다).
- `agents-emit-check` 실패는 이 plan 에서 재현했다(관측): `manager-todo.md` 의 `description:` 첫 줄에 `(probe)` 를 덧붙이자 `AGENTEMIT_UPDATE= go test ./internal/template/agentemit/... -run '^TestGoldenCommittedArtifactsMatchEmission$' -count=1 -v` 가 `.codex/agents/moai/manager-todo.toml: committed artifact differs from emission (sha256 mismatch)`, `--- FAIL: TestGoldenCommittedArtifactsMatchEmission `, `FAIL` 이고, `make agents-emit` 뒤 같은 명령이 `--- PASS`. `make agents-emit` 은 `internal/template/templates/.codex/agents/moai/manager-todo.toml` 한 파일만 바꿨다. 변이는 복원했다. 로컬 `.codex/agents/moai` 디렉터리는 이 트리에 없다. 12개 전체·실제 축약 편집에서의 생성 파일 집합은 관측하지 않았다(실행 단계 몫).

---

## Amendments

### Amendment 0.4.0 (2026-10-03, card t1450)

- prior_completed_version: 0.3.1
- prior_completed_sha: b7ea0d823 (the sync commit that carried `sync_commit_sha`)
- rationale: sync-audit (`.moai/reports/t1450/sync-audit.md`, FAIL 81.7, receipt rcpt-8cf2e184c500ae809e6bd8cd) found two `dropped` units that were the sole carrier of user-facing instructions (`moai-easy-0183`, `moai-learn-0122`), and the ledger test accepted any non-empty `survivor` string. The run-phase repair restored 25 units verbatim after a strict sweep of all 77 dropped rows; dropped rows are now 52 (7,438 UTF-16 units total); the ledger test requires `survivor_file` + `survivor_anchor` that resolve (`SURVIVOR_UNRESOLVED` otherwise), with RED-first mutation subtests.
- scope: REQ-PFD-002 (sizes, constants, measured outcome 154,235 → 151,673, −2,562, −1.66%), REQ-PFD-003 (survivor resolution rule), REQ-PFD-013 (guard scope statement), §B ledger fields, §H residual risk; `acceptance.md` AC-PFD-002 and `plan.md` D4 supersession note. No requirement added or renumbered (REQ 13 / AC 16). Residual risk: a survivor pointer proves the anchor exists in the named file, not that it carries the instruction; that stays reviewer judgement.

### Amendment 0.5.0 (2026-10-03, card t1450)

- prior_completed_version: 0.4.0
- prior_completed_sha: 6b2fe3a1f454 (the current `sync_commit_sha` in `progress.md` §E.4)
- rationale: after the codex cross-model objections the leader ordered a conservative restore. Run repair-2 (commits 76e480974 + adf727ffb) restored moai-easy units 0073, 0163 and 0181 verbatim and cleared the stale free-text `survivor` prose of the remaining dropped rows.
- scope: REQ-PFD-002 (moai.md 61,362 / moai-easy.md 23,586 / moai-learn.md 28,517 whole-file UTF-16; 49 dropped rows totalling 6,888 units, was 52 / 7,438; zero rewrite rows; first-turn tokens 154,235 → 151,873, −2,362, −1.53%, superseding −2,562 / −1.66%), REQ-PFD-003 and the §B ledger field description (`DROPPED_SURVIVOR` fires when `survivor_file` or `survivor_anchor` is empty; `SURVIVOR_UNRESOLVED` unchanged). No requirement added or renumbered (REQ 13 / AC 16).
