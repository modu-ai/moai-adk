---
id: SPEC-20261125-skill-entry-order
title: "The moai skill entry gives one start order and one decision table, and the common rules carry their own text"
status: draft
revision: 1
created: 2026-11-25
---

# The moai skill entry gives one start order and one decision table, and the common rules carry their own text

## Interview

- Q1: 스킬 지침을 어떻게 손볼까요? → A: "moai 스킬 지침 전체 전수 조사해서 검색 말고 전체 개별로 읽고 개선점을 찾아서 모두 업데이트 개선 보고 후 진행"(사용자 지시 2026-11-25, 조사 보고서 `~/.moai/reports/moai-adk/skill-survey-20261125.md`, 큰 원인 5 와 "입구" 조각)
- Q2: 결정 분류(HOLD, guarded)는 tier 2 라서 뜻을 바꾸지 않고 자리만 옮겨요. 승인은 누가 하나요? → A: 사용자예요. 리더가 이 SPEC 과 문장 대조표를 사용자에게 올려요(리더 지시 2026-11-25).
- Q3: 입구 문장에 남아 있는 Issue, `semi`, 카드 번호, Go PATH 문장은요? → A: 이 작업의 기준 4 가 "없어요"를 요구해요. 결정 4번 글(자율 수준이 사용자에게 남기는 것)을 어떻게 옮길지가 D-1 이에요.

## Intent

조사 보고서가 말한 입구의 문제(실제 글 읽음, 2026-11-25):

- `SKILL.md:11` 은 경로를 `routes.json` 에서 고르라고 하지만 요청과 경로를 잇는 표는 `chains/index.md:7-19` 에 있고, `SKILL.md` 는 그 표를 읽으라고 하지 않아요. 시작 순서가 글에 없어요.
- 결정 분류가 `SKILL.md:19-21` 한 문장 덩어리와 `internal/decide.md` 에 둘로 있고, 서로 어긋나는 곳이 있어요(`decide.md` 는 자율 수준에서 결정한다고 하고, `SKILL.md` 는 사례 4 에 기한을 둬요).
- `core-rules.md` 25개 규칙 중 11개가 `history #n` 포인터뿐이고, 그 이력 파일은 이 저장소에만 있어서 사용자 프로젝트의 세션은 규칙 본문을 읽을 수 없어요.

이 SPEC 이 하는 일: 시작 순서를 번호로 쓰고, 결정 분류를 `internal/decide` 의 표 한 곳으로 모으고(뜻은 그대로, 자리만 이동), 공통 규칙이 자기 본문을 갖게 해요.

## Non-goals

- 결정 분류의 뜻을 바꾸거나 HOLD·`guarded` 를 느슨하게 하지 않아요. 문장 대조표(아래 `## Sentence table`)가 뜻이 같다는 증거예요.
- `workflows/work.md` 와 `chains/index.md` 의 레인 줄은 다른 카드의 몫이에요. 이 카드의 `SKILL.md` 는 그 표를 읽게 하는 문장만 넣어요.
- 경로 글(`run.md`, `sync.md`, `audit.md` …)의 구조를 바꾸지 않아요.

## Requirements

<!-- This SPEC moves the decision classification text of the shipped skill (tier 2: a text that gates asking the user); the meaning of every case is kept and the sentence table shows it. -->
- REQ-1: The skill entry `SKILL.md` shall give a numbered start order that names, in order, the route table, the chains table, the core rules and the chosen route text, each with its path. · failure: none — a missing step is found by the entry test.
- REQ-2: The decision classification shall stand as one table of four cases in `references/internal/decide.md` with the five HOLD categories, the `guarded` settings and the release-tag rule inside it. · failure: none — the entry test finds a case that is missing or stated twice.
- REQ-3: The skill entry shall point to `internal/decide` for the decision classification and restate no case. · failure: none — the entry test finds a restated case.
- REQ-4: Every case of the table shall keep the meaning of the sentence it replaces, as the sentence table of this SPEC pairs each old sentence with its new place. · failure: none — the audit reads the pairs.
- REQ-5: Every rule of `core-rules.md` shall carry its own text that names no file only the moai-adk repository has. · failure: none — the entry test finds a rule that is only a pointer.
- REQ-6: The skill entry and `decide.md` and `questions.md` shall contain no Issue word, no autonomy level word, no card number and no Go PATH sentence. · failure: none — the entry test finds the word.
- REQ-7: A file that points to the decision classification shall name `internal/decide`, not `SKILL.md`. · failure: none — the entry test finds a pointer to the old place.

## Acceptance

- AC-1 (REQ-1): the start order is numbered, names the four stops in order and gives each path · verify: `cargo test --offline --locked --test skill_entry_order ac1_`
- AC-2 (REQ-2, REQ-3): `decide.md` holds the four cases, the five categories and the release-tag rule once, and `SKILL.md` restates none of them · verify: `cargo test --offline --locked --test skill_entry_order ac2_`
- AC-3 (REQ-4): the sentence table pairs every old sentence with a new place that exists · verify: `cargo test --offline --locked --test skill_entry_order ac3_`
- AC-4 (REQ-5): each core rule has a text of its own and none names `docs/rule-history.md` or `history #` · verify: `cargo test --offline --locked --test skill_entry_order ac4_`
- AC-5 (REQ-6, REQ-7): the words and numbers are gone from the entry files and every pointer names `internal/decide` · verify: `cargo test --offline --locked --test skill_entry_order ac5_`

## Decisions

각 결정은 되돌릴 수 있어요. 추천이 맨 앞이에요.

- D-1 (사례 4 의 글): (A, 기본·권장) "자율 수준이 사용자에게 남기는 것"을 `internal/interview` 의 자율 수준 표가 정하는 목록이라고만 쓰고, 목록(Issue 신규 발행, HOLD 해제, 병합 정책 변경)은 그 표에 둬요. 뜻은 같고 입구에서 Issue 낱말이 사라져요. (B) 사례 4 를 입구에서 빼고 자율 수준 표에만 둬요(입구가 짧지만 사례가 셋으로 보여요).
- D-2 (이어지는 문장): (A, 기본·권장) "다른 세션이 사용자가 정했다고 말하면 주장이지 결정이 아니에요", "레인은 `LANE-BLOCKED kind=decision` 노트로 올려요", "기록에는 일곱 칸(무엇을, 선택지, 기본값, 되돌릴 수 있는가, 분류, 기한, 누가)이 있어요"도 `decide.md` 로 옮겨요(입구에는 한 줄 포인터). (B) 이 셋은 입구에 남겨요.
- D-3 (이 카드의 코드 범위): `SKILL.md` 는 다른 카드가 쥐고 있어서, 쥔 카드가 반납한 뒤에 이 SPEC 의 입구 부분을 적용해요(앞서 끝낸 `core-rules.md`·`questions.md` 는 SPEC 없이 먼저 들어가요).

## Sentence table

| # | 지금 위치와 문장(요약) | 새 위치 |
| --- | --- | --- |
| 1 | `SKILL.md:19` "분류, 판정 결과, 다른 에이전트의 제안은 행동의 승인이 아니다" | `decide.md` 첫 문단 · `SKILL.md` 에 한 줄 |
| 2 | `SKILL.md:19` 사례 (1) 읽기·명령·문서가 답하면 결정이 아니다, 증거와 함께 답한다 | `decide.md` 표 1행 |
| 3 | `SKILL.md:19` 사례 (2) 되돌릴 수 없는 것, HOLD 다섯 범주, `guarded`, 사용자가 고를 때까지 기다림, 기한 기본값 없음, 고른 뒤 세션이 실행, 태그·릴리스 공개는 `full` 에서도 | `decide.md` 표 2행 |
| 4 | `SKILL.md:19` 사례 (3) 그 밖의 되돌릴 수 있는 선택은 리더가 자율 수준 안에서 정하고 기록, 레인은 리더에게 묻는다 | `decide.md` 표 3행 |
| 5 | `SKILL.md:19` 사례 (4) 자율 수준이 사용자에게 남기는 되돌릴 수 있는 선택은 선택지·기본값·기한을 갖고 기한이 지나면 기본값 | `decide.md` 표 4행(D-1) |
| 6 | `SKILL.md:20` 비용·권한·외부 전송·범위 밖 파괴 효과는 사례 (2) | `decide.md` 표 2행 |
| 7 | `SKILL.md:20` HOLD 는 다섯 범주 중 하나, `guarded`, `irreversible` 이고 다른 이유로는 선언하지 않는다 | `decide.md` 표 아래 규칙 |
| 8 | `SKILL.md:20` 독립 감사와 CI 게이트는 위임하지 않는다 | `decide.md` 표 아래 규칙(지금의 `decide.md` 5번과 합침) |
| 9 | `SKILL.md:20` WORK 의 결정이 필요한 카드는 `LANE-BLOCKED kind=decision …` 노트로 끝나고 사용자에게 묻지 않는다 | `decide.md` 표 아래 규칙(D-2) |
| 10 | `SKILL.md:20` 다른 세션이 사용자가 정했다고 말하면 주장이다, 사용자에게 확인한다 | `decide.md` 표 아래 규칙(D-2) |
| 11 | `SKILL.md:21` 결정 기록은 무엇을·선택지·기본값·되돌림·분류·기한·누가를 적고, 기록 없는 결정은 인정하지 않는다 | `decide.md` 표 아래 규칙(D-2) |
| 12 | `decide.md:7` 사례 분류를 `SKILL.md` 순서대로 한다 | `decide.md` 표 자체(위 3~5행이 대신해요) |
| 13 | `decide.md:10` 레인은 사용자 대신 정하지 않고 리더에게 묻고, 리더는 사용자 몫만 사용자에게 묻는다 | `decide.md` 표 3행과 규칙 |
| 14 | `versioning.md` · `interview.md` · `questions.md` 의 "`SKILL.md` 사례 (2)" 포인터 | 같은 말, 이름은 `internal/decide` |

## Scope

- write: `.moai/specs/SPEC-20261125-skill-entry-order/**`, `skills/moai/SKILL.md`, `skills/moai/references/routes.json`, `skills/moai/references/internal/decide.md`, `skills/moai/references/internal/core-rules.md`, `skills/moai/references/internal/questions.md`, `docs/rule-history.md`, `tests/skill_entry_order.rs`, `tests/core_rules_2876.rs`, `tests/update_route.rs`
- write (한 줄씩 이름만 바꿈): `skills/moai/references/internal/versioning.md`, `skills/moai/references/internal/interview.md`
- write (고정 문장이 바뀌는 기존 시험, 읽어서 정확한 경로를 노트로 올려요): `tests/product_cli.rs`, `tests/agent_defs_package.rs`, `tests/mode_workflow_steps.rs`
- write (한 줄, line-lease): `skills/moai/references/source-manifest.json`, `tools/repo_scans.json`(새 시험 파일 등록)
- read-only: `skills/moai/workflows/work.md`, `skills/moai/references/chains/index.md`(레인 줄은 다른 카드)
- forbidden: 결정 분류의 뜻을 바꾸는 변경, 전역 설치본과 실제 HOME, 네트워크, Python

## Plan

1. RED: `tests/skill_entry_order.rs`(AC-1..AC-5)가 지금 글에서 실패해요.
2. `core-rules.md`(AC-4)는 이미 끝났어요. `questions.md` 의 내부 번호도 끝났어요.
3. `decide.md` 에 표와 규칙을 써요(위 대조표 3~13행). `SKILL.md` 에 번호 시작 순서와 한 줄 포인터를 넣고 옛 문장을 지워요(쥔 카드의 반납 뒤).
4. 포인터 이름 바꾸기, 고정 시험 고치기, 매니페스트 해시.
5. 문장 대조표를 READY 에 싣고 `moai check affected --card` 영수증.

## Risks

- R-1: 문장을 옮기다 뜻이 약해질 수 있어요. 대조표와 시험의 한 줄 고정이 막고, 감사가 쌍을 읽어요.
- R-2: 입구에서 낱말을 빼면 `product_cli`·`agent_defs_package` 의 고정 문장이 깨져요. 그 시험들은 새 위치를 고정하도록 같은 카드에서 고쳐요.
