---
id: SPEC-20261118-card-ledger-wording
title: "Card ledger wording: agents, README and the decisions table say cards, not GitHub Issues"
status: draft
revision: 1
created: 2026-11-18
---

# Card ledger wording: agents, README and the decisions table say cards, not GitHub Issues

## Intent

사용자 결정(2026-11-16): GitHub Issues는 더 이상 작업 장부가 아니고, 프로젝트의 로컬 카드(`moai card`)가 작업 장부예요.
모드 파일은 이미 카드로 바뀌었지만, 기본 에이전트 파일(`agents/plan.md`, `agents/impl.md`, `agents/docs.md`)과 README는 아직 "Issue"를
작업 단위로 말하고, 결정 기록 표는 2026-11-10의 "Issue만 쓴다" 행(#2384)이 대체되었다는 줄이 없어요. 에이전트가 이 글을 그대로 읽고 따르므로,
글이 현재 결정과 어긋나면 일을 잘못된 단위로 해요. 이 SPEC은 그 글만 맞춰요. 코드 동작(claim guard, commit trailer)은 바꾸지 않아요.

## Non-goals

- claim guard와 commit trailer 코드 변경: 커밋 신원은 이미 `Moai-Spec`/`Moai-Session` 트레일러가 대신해요(결정 U-B, 2026-11-17).
- 생성되는 `AGENTS.md` 문구(`src/agents.rs`): 다른 카드가 같은 줄을 맡고 있어요.
- README 첫 문장, `Cargo.toml`, `README.ko.md`, `docs/llms.txt`: 첫 문장과 `Cargo.toml` description은 `tests/cargo_metadata_test.py`가 같다고 고정하고 있어서 한 묶음이에요. 이 SPEC의 쓰기 범위 밖이라 후속 SPEC으로 함께 고쳐요(Decisions D-3).

## Requirements

- REQ-1: The decisions table shall carry a row dated 2026-11-16 that records the local card ledger as the work ledger and names the 2026-11-10 forge-only row (#2384) as replaced.
- REQ-2: The three base agent files for plan and impl and docs shall name the unit of work as a card or a SPEC.
- REQ-3: The `RESULT` line of the three base agent files shall name a card, a SPEC id or a path.
- REQ-4: The README project facts list shall state that local cards are the work ledger.
- REQ-5: The README flow diagram and its first bullet shall name a card and a SPEC where they named a GitHub Issue.
- REQ-6: README.md shall keep saying that the merge gate is the `moai/audit` commit status. · failure: none — 이 줄은 바꾸지 않아요.

## Acceptance

- AC-1 (REQ-1): 결정 표에 2026-11-16 행이 있고 `#2384`와 "대체"를 함께 적어요.
  verify: `cargo test --locked --offline --test product_cli card_ledger_wording_decision_row_replaces_the_forge_only_row`
- AC-2 (REQ-2, REQ-3): 세 에이전트 파일에 `Issue`라는 단어가 없고 `RESULT` 줄은 `<card, SPEC id or path>` 꼴이에요.
  verify: `cargo test --locked --offline --test product_cli card_ledger_wording_base_agents_name_cards_not_issues`
- AC-3 (REQ-4, REQ-5, REQ-6): README 정보 목록은 로컬 카드를 작업 장부로 말하고 흐름도에 `GitHub Issue`가 없으며, `moai/audit` 병합 관문 줄은 그대로예요.
  verify: `cargo test --locked --offline --test product_cli card_ledger_wording_readme_says_cards_and_keeps_the_gate`
- AC-4 (REQ-1…REQ-6): 이웃 검사.
  verify: `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --test mode_cards_53`; `cargo test --locked --offline --test repo_agents_md`; `cargo fmt --check`

## Scope

- write: `agents/plan.md`, `agents/impl.md`, `agents/docs.md`, `README.md`, `docs/decisions-2026-11-08.md`, `tests/product_cli.rs`,
  `.moai/specs/SPEC-20261118-card-ledger-wording/spec.md`
- read-only: `skills/moai/workflows/mode-agent.md`, `tests/mode_cards_53.rs`, `README.ko.md`, `Cargo.toml`
- forbidden: 실제 HOME과 원장, 네트워크, `src/agents.rs`(다른 카드가 같은 줄을 맡아요)

## Plan

1. 시험 먼저: `tests/product_cli.rs`에 위 세 시험을 더하고 RED를 읽어요(글이 아직 옛말이라 셋 다 실패해야 해요).
2. `agents/*.md`, `README.md`, 결정 표 행을 고쳐 GREEN으로 만들어요.
3. 이웃 검사를 돌려요.

## Risks

- 다른 시험이 옛 문구를 고정하고 있을 수 있어요 → 이웃 검사 AC-4가 잡아요.
- 쓰기 범위 밖의 같은 문구(`Cargo.toml`, `README.ko.md`)가 남아 README와 어긋나요 → D-3의 결정 요청을 카드에 남겼고, 남은 곳은 보고서에 적어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 시험은 새 파일이 아니라 `tests/product_cli.rs`에 더해요 | 새 시험 파일은 이 작업의 쓰기 범위 밖 | 별도 파일로 옮기기 |
| D-2 | 카드 단위 `claim` 트레일러 줄은 더하지 않아요(원래 카드 항목의 "카드 trailer"는 넣지 않음) | 2026-11-17 결정 U-B: 커밋 신원은 `Moai-Spec`/`Moai-Session`이 대체 | 후속 SPEC에서 되살리기 |
| D-3 | README 첫 문장, `Cargo.toml:8`, `README.ko.md`는 이번에 고치지 않아요 | 첫 문장을 바꾸면 `tests/cargo_metadata_test.py`가 `Cargo.toml` description과의 불일치로 CI를 막았어요(PR 첫 head, 실행 44227800483); 쓰기 범위 밖이라 리더에게 물었고 답이 없어 되돌렸어요 | 리더가 범위를 늘리면 후속 SPEC에서 둘을 함께 고치기 |

## Result

- AC-1: met — `cargo test --locked --offline --test product_cli card_ledger_wording` (RED 먼저: 3 failed, exit 101, 글이 옛말이라). 결정 표에 11-16 행이 #2384를 대체한다고 적어요.
- AC-2: met — 같은 시험. `agents/plan.md`, `impl.md`, `docs.md`에서 `Issue`가 없어지고 `RESULT` 줄이 카드·SPEC id·경로를 말해요.
- AC-3: met — 같은 시험. README 흐름도·첫 불릿·정보 목록이 로컬 카드를 말하고 `moai/audit` 병합 관문 줄은 그대로예요.
- AC-4: met — `product_cli` 79 passed; `mode_cards_53` 9 passed; `repo_agents_md` 3 passed; `decision_append_cli` 17 passed; `cargo fmt --check` exit 0; `python3 -m unittest tests.ci_scope_test` OK.
- 남은 것: README 첫 문장과 `Cargo.toml:8`(시험이 둘을 같게 고정), `README.ko.md`, `docs/llms.txt`는 쓰기 범위 밖이라 그대로예요(D-3, 리더의 범위 결정 대기). SPEC은 draft예요: 승인은 사용자의 선택이라 이 레인이 하지 않았어요.
