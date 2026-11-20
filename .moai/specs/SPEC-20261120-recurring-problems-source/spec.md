---
id: SPEC-20261120-recurring-problems-source
title: "AGENTS.local.md: a problem that happens twice is fixed by changing moai-adk, not by another note"
status: draft
revision: 1
created: 2026-11-20
---

# AGENTS.local.md: a problem that happens twice is fixed by changing moai-adk, not by another note

## Intent

사용자 지시(2026-11-20, 세션 병목 점검 뒤 리더 대화)예요. 같은 종류의 문제가 되풀이되면 메모를 더하는 게 아니라 moai-adk 의 소스(검사·시험·명령·가드·생성 지침)를 고치는 카드를 그 라운드에 열어요. 점검 보고서 `~/.moai/reports/moai-adk/session-bottleneck-review-2026-11-20.md` 가 읽은 사실이에요: 한 릴리스 PR 이 좁은 시험 선택 때문에 CI 에서 두 번 실패했고, 파일 예약 때문에 레인이 약 25번 막혔고, 커밋 메시지의 카드 번호는 감사에서만 발견됐어요. 이 SPEC 은 `AGENTS.local.md` 의 `## Leader supervision` 에 그 규칙을, `## Rule provenance` 에 출처 한 줄을 적고, 시험이 둘을 고정해요. 새 Rust 명령은 없어요.

## Non-goals

- 되풀이 횟수를 세는 도구를 만드는 일: 규칙은 라운드마다 원장 노트에서 세라고만 말해요. 도구는 따로 카드가 맡아요.
- 게이트·HOLD·사용자 규칙을 느슨하게 하는 일: 이 규칙은 그것들을 그대로 두고 소스를 고치는 길만 더해요.
- `AGENTS.md` 바꾸기: 규칙은 리더의 일이라 `AGENTS.local.md` 에만 적어요.

## Requirements

- REQ-1: The `AGENTS.local.md` section `## Leader supervision` shall state that when the same class of problem happens twice, the leader opens a card in that round that changes moai-adk itself. · failure: none — a test reads the section for the rule's sentence.
- REQ-2: The rule shall name the four recurrences it counts, a rule broken twice, one CI or integration failure cause twice, a lane block with one cause twice and a manual leader step repeated, and the ledger notes the count is read from. · failure: none — a test reads the rule for each of the four and for the note kinds.
- REQ-3: The rule shall say that the card follows the normal SPEC, audit and integration path and never relaxes a gate, a HOLD or a user rule. · failure: none — a test reads the rule for that sentence.
- REQ-4: The `AGENTS.local.md` section `## Rule provenance` shall carry a row for the rule that names the user instruction of 2026-11-20, the three causes and the review report path. · failure: none — a test reads the table row for each.
- REQ-5: The rule and its provenance row shall name no card number. · failure: none — a test scans both texts for the card-number shape.
- REQ-6: While the file holds the rule, `AGENTS.local.md` shall stay under the 32 KiB instruction budget. · failure: none — a test reads the file size.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): 규칙 글에 되풀이 한 번 더 일어나면 카드를 연다는 문장, 네 가지 되풀이, 원장 노트 종류, 게이트·HOLD·사용자 규칙을 느슨하게 하지 않는다는 문장이 있어요.
  verify: `cargo test --locked --offline --test recurring_problems_rule the_supervision_section_states_the_rule`
- AC-2 (REQ-4, REQ-5): 출처 표에 한 줄이 있고 날짜·세 원인·보고서 경로가 있으며, 규칙과 그 줄에 카드 번호가 없어요.
  verify: `cargo test --locked --offline --test recurring_problems_rule the_provenance_table_names_the_instruction_and_its_causes`
- AC-3 (REQ-6): 파일이 32 KiB 보다 작아요.
  verify: `cargo test --locked --offline --test repo_agents_md`; `cargo test --locked --offline --test recurring_problems_rule the_file_stays_inside_its_budget`

## Scope

- write: `.moai/specs/SPEC-20261120-recurring-problems-source/spec.md`, `tests/recurring_problems_rule.rs`, `AGENTS.local.md`
- read-only: `tests/repo_agents_md.rs`, `docs/agents-md-maintenance.md`
- forbidden: 실제 HOME, 사용자의 전역 git 설정, push 이외의 원격 쓰기

## Plan

1. 시험 먼저: `tests/recurring_problems_rule.rs` 시험 셋을 쓰고 실패(RED)를 읽어요.
2. `## Leader supervision` 끝에 규칙 한 항목을, `## Rule provenance` 표 끝에 한 줄을 써요. 파일 크기 예산(32 KiB, 남은 여유 1 KiB 남짓)에 맞게 짧게 써요.
3. 시험 셋과 `repo_agents_md`, `tools/repo_scans.json` 의 모든 명령, 영역 파이썬 시험, READY 전 6종을 돌려요.

## Risks

- 파일 여유가 1 KiB 남짓이라 글이 길어지면 `repo_agents_md` 의 크기 시험이 실패해요 → 규칙은 한 항목, 출처는 한 줄로 쓰고 크기 시험을 이 SPEC 의 시험에도 넣어요.
- 되풀이를 세는 기준이 흐리면 규칙이 안 지켜져요 → 네 가지 되풀이와 센 곳(원장 노트의 종류)을 글에 이름으로 적어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 규칙은 `AGENTS.local.md` 의 `## Leader supervision` 한 항목으로 적어요 | 사용자 지시 2026-11-20; 리더가 라운드마다 하는 일이라서요 | 항목과 출처 줄을 지워요 |
| D-2 | 출처 표에 규칙 한 줄을 더해요 | 이 파일의 다른 규칙이 모두 출처 줄을 가져요 | 줄을 지워요 |
| D-3 | 시험은 새 파일 하나이고 `repo_agents_md` 는 그대로 둬요 | 새 규칙의 문장만 고정하고 기존 시험의 범위는 넓히지 않아요 | 시험 파일을 지워요 |

## Result

(채워요: RED, GREEN, 변이, 이웃, Not run)
