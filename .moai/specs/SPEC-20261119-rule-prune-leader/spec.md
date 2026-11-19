---
id: SPEC-20261119-rule-prune-leader
title: "Prune mode-leader.md and make the rule inventory name its enforcers row by row (C-103 last slice)"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261119-rule-prune
---

# Prune mode-leader.md and make the rule inventory name its enforcers row by row (C-103 last slice)

## Intent

카드 C-103 의 문서 조각(SPEC-20261119-rule-prune)은 `mode-agent.md`, `internal/core-rules.md`, `workflows/audit.md` 를 다뤘고 `mode-leader.md` 편집과 `source-manifest.json` 갱신을 마지막 조각으로 미뤘어요(C-99 가 두 파일을 예약해서, 리더 결정 2527·2535). 지금 `mode-leader.md` 는 예약이 없어요(`card show` 로 활성 카드 19개의 쓰기 범위를 읽어 확인). 카드 기준 2("강제되는 규칙은 하나도 안 빠진다, 살린 규칙마다 강제 시험 이름 또는 text only, 매니페스트와 크기 예산 시험 통과")는 열려 있어요.

이 조각이 하는 일: (1) `mode-leader.md` 의 규칙 가운데 근거가 하나뿐이고 시험·훅·게이트가 읽지 않는 문장을 `rule-history.md` 의 history 절로 옮기고 중복은 합쳐요. (2) `rule-history.md` 의 기계 목록 셋(audit, mode-agent, mode-leader)을 지금 팁의 글에 맞게 다시 쓰고, 행마다 그 줄을 읽는 시험 파일 이름 또는 `text only` 를 붙여요(지난 감사 N2 가 "파일 단위 일괄 문구"라고 지적했고 재감사에서 니트로 이월됐어요). (3) history 본문과 강제자 이름이 거짓이 되지 않게 지키는 시험 하나를 더해요(이월 니트 "history 본문 무가드"). (4) 마지막 텍스트 커밋에서 매니페스트 해시를 갱신해요.

측정(읽기 전용, 이 SPEC 이 쓰이는 시점의 팁 190de2d8d7; 스크래치 스크립트 `pinmap.py`: `tests/` 아래 `.rs`·`.py` 의 문자열 리터럴 가운데 20자 이상이 글 안에 그대로 있으면 그 문장은 "핀"): `mode-leader.md` 45,128B 에서 규칙 후보 줄 41개 중 36개가 핀이고, 문장 단위로는 후보 128문장 중 30문장(5,821B)을 어떤 리터럴도 읽지 않아요. 줄 단위 수치(안 읽히는 줄 5개 1,715B)는 긴 줄 하나가 문장 여럿을 품어서 과소평가해요. `mode-agent.md` 는 후보 167문장 중 30문장(4,432B)이 안 읽혀요. 그래서 옮길 후보는 `mode-leader.md` 의 그 30문장이고, 문장마다 코드·훅·게이트(`src/`, `tools/`, `.github/`)가 읽는지 사람이 읽어 판정해요. 30문장 중에는 "병합은 감사 게이트를 따른다", "소유권은 카드 예약에서만 온다" 같은 계약 문장이 있어서 기계 판정만으로 옮기지 않아요(결정 D-2).

## Non-goals

- 규칙의 의미 변경: 옮기기·합치기·강제자 표기만 해요(이 카드 SPEC 의 forbidden 그대로).
- `mode-agent.md`·`audit.md`·`core-rules.md` 의 편집: 목록만 다시 써요(기계 목록의 "그 줄을 읽는 시험" 칸을 채울 뿐 글은 그대로).
- C-96 기준 3(precheck 를 배차 첫 단계로): `moai card precheck` 가 팁에 없어서 이 조각이 글에 넣지 않아요(결정 D-4). C-96 이 그 명령을 내면 그때 `mode-leader.md` 에 한 줄을 넣는 일을 이월해 카드 노트에 적어요.
- `tests/orca_optional_3005.rs`: agent-3 의 C-120 이 `SRC_ORCA_FILES` 에 줄을 더하는 중이라 이 조각은 그 파일을 만지지 않아요. 옮기는 문장에 Orca 낱말이 있어도 `rule-history.md` 는 이미 허용 목록에 있어요.
- 목록 행의 글과 현재 파일의 일치 검사: 기계 목록은 날짜가 찍힌 기록이고 다른 카드가 워크플로를 계속 고쳐서, 일치 시험은 그들의 병합을 막아요(결정 D-3).

## Requirements

- REQ-1: When the inventory of `workflows/mode-leader.md`, `workflows/mode-agent.md` and `workflows/audit.md` is rewritten, `references/rule-history.md` shall hold one row per rule line of each file at the base tip, each row carrying the first words of the line, its source, and the names of the test files that read the line or the words `text only`. · failure: none — a row without the last field is REQ-9's to find.
- REQ-2 (REQ-1): When an inventory section is written, its heading shall carry the short sha of the base tip it was written from. · failure: none — a heading without a sha is REQ-9's to find.
- REQ-3: IF a sentence of `mode-leader.md` has no test literal of 20 characters or more inside it, no reader in `src/`, `tools/` or `.github/`, and at most one source, THEN it shall move word for word to a history section of `references/rule-history.md` and leave `mode-leader.md`, with a one-line reason.
- REQ-4: When a sentence of `mode-leader.md` repeats a rule that `SKILL.md`, `references/internal/core-rules.md` or `mode-agent.md` already holds, the leader file shall keep one sentence. · failure: none — a duplicate left in both files is the audit's to find (AC-3).
- REQ-5 (REQ-4): When two sentences are merged, the inventory shall name the sentence removed and where the kept one stands. · failure: none — a merge without that row is the audit's to find (AC-3).
- REQ-6: IF a sentence of `mode-leader.md` names a path, command or check that does not exist at the base tip, THEN it shall move to the history section too. · failure: none — a missing name is read from `moai help --all` and the tree, and found by the audit.
- REQ-7: When the slice completes, `references/rule-history.md` shall report the byte count of `mode-leader.md` against its base count. · failure: none — a missing report is the audit's to find (AC-1).
- REQ-8 (REQ-7): IF `mode-leader.md` did not get smaller, THEN `references/rule-history.md` shall name the rule that had to stay and why.
- REQ-9: When the tests run, `rule_history_guard` shall fail on a history entry whose text occurs verbatim in the live rule file its heading names as the source, on a test file name written in `rule-history.md` that exists neither under `tests/` nor as a path from the root, and on an inventory row or section heading that lacks the fields of REQ-1 and REQ-2. · failure: none — the failing test names the entry, the file or the row.
- REQ-10: IF `moai card precheck` does not exist at the base tip, THEN `mode-leader.md` shall not name it. · failure: none — `help --all` is the check; the carried criterion goes to the card note.
- REQ-11: When the last text commit is made, `skills/moai/references/source-manifest.json` shall carry the written hash of every package file this slice edited. · failure: none — `python3 -I tools/source_manifest.py check` and `tests.source_manifest_test` fail on a stale hash.

## Acceptance

- AC-1 (REQ-3, REQ-4, REQ-6, REQ-7, REQ-8): `mode-leader.md` is smaller than at the base tip by the bytes of the sentences moved (the report names both counts), every moved sentence is in the history section word for word, each with a reason, and no sentence a test literal reads was moved.
  verify: `cargo test --locked --offline --test mode_workflow_steps`
  verify: `cargo test --locked --offline --test rule_history_guard`
- AC-2 (REQ-1, REQ-2, REQ-9): the three mechanical inventories name, per row, the test files that read the line or `text only`, every named test file exists, and each section heading carries the short sha of the base tip.
  verify: `cargo test --locked --offline --test rule_history_guard every_inventory_row_names_its_enforcers_or_text_only`
  verify: `cargo test --locked --offline --test rule_history_guard every_test_file_named_in_the_history_exists`
- AC-3 (REQ-4, REQ-5, REQ-9): no history entry occurs verbatim in the live rule file it was moved from (a merged sentence may stay in the other file's copy), and a sentence that was merged names where the kept one stands.
  verify: `cargo test --locked --offline --test rule_history_guard no_history_entry_is_still_live`
- AC-4 (REQ-10): `moai help --all` of the built binary lists no `card precheck` while the leader file does not name it, and the carried criterion is on the card.
  verify: `cargo test --locked --offline --test mode_workflow_steps ops_named_by_the_mode_workflows_exist_with_their_flags`
- AC-5 (REQ-11): the manifest check passes on the head and on the trial merge with the local `v2026.11.0`.
  verify: `python3 -I tools/source_manifest.py check`
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.source_manifest_test tests.skill_references_test tests.repo_scans_test`
- AC-6 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9, REQ-10, REQ-11): the neighbour checks pass: the workflow size budget, the pins of the two mode files, the core-rules pins and the style checks.
  verify: `cargo test --locked --offline --test workflow_size_budget`
  verify: `cargo test --locked --offline --test core_rules_2876`
  verify: `cargo test --locked --offline --test orca_optional_3005`
  verify: `cargo fmt --check`

## Scope

- write: `skills/moai/workflows/mode-leader.md`, `skills/moai/references/rule-history.md`, `tests/rule_history_guard.rs`, `.moai/specs/SPEC-20261119-rule-prune-leader/spec.md`, `skills/moai/references/source-manifest.json` (the manifest while C-112 holds it is taken only in the last step)
- forbidden: a change of what a rule means, the real HOME, ledgers and keys, network writes, `tests/orca_optional_3005.rs` (C-120), the local `v2026.11.0` branch and the default checkout

## Plan

1. SPEC, then a RED test: `tests/rule_history_guard.rs` with the three checks of REQ-9; it fails first because no inventory row names its enforcers.
2. Evidence at the base tip (read-only, `target/c103l/`): the pin map of the three files (per line and per sentence), then for each of the 30 unpinned `mode-leader.md` sentences a read of `src/`, `tools/`, `.github/`, `SKILL.md`, `core-rules.md` and `mode-agent.md` for a reader or a duplicate, and a `git blame` for its source.
3. Judgment per sentence: move (REQ-3, REQ-6), merge (REQ-4) or stay as `text only` with a reason. The list goes to the card as a plan note before the edit.
4. Edit: `mode-leader.md` and `rule-history.md` (the three inventories rewritten with the per-row field, the new history section, the byte report).
5. Checks: the guard, the pins, `workflow_size_budget`, the manifest tests after the manifest step, the leader's seven checks.

## Risks

- Where a rule sentence ends is a judgment: the unit of the pin map is a sentence split after `. ` before a capital letter, a backtick, a bracket or `*`; the inventory keeps the line as its row so that it does not grow past the package file bound (65,536 bytes; `rule-history.md` is 17,376 now).
- A sentence no test reads may still be a contract: ownership only from a reservation, no merge without an independent audit, keep-set decisions to the user. These are read, not moved by the tool; a contract with a single copy stays marked `text only` with its reason (D-2).
- Other cards edit `mode-leader.md` while this slice runs (C-95, C-99, C-120): the tip moves, the branch merges the local `v2026.11.0` instead of rebasing, and the inventory heading names the tip it was written from.
- The manifest is reserved by C-112 now: the last step may have to wait for its release; if it does, the slice reports that as blocked and does not write READY.
- The guard test reads three rule files and `rule-history.md` by path: a rename of one of them fails it on purpose.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | the pin map unit is a sentence; the inventory row stays a line | lines over-count pins (36 of 41 lines, 98 of 128 sentences); a row per sentence would push `rule-history.md` near 65,536 bytes | row per sentence |
| D-2 | a mechanical hit is only a candidate: each of the 30 unpinned sentences is read for a code/hook/gate reader and a duplicate before it moves; a single-copy contract stays | the card says "no rule that a test, hook or gate enforces is dropped" and the audit gate (`.github/workflows/audit-gate.yml`) is a gate | move all 30 |
| D-3 | the guard checks history entries, test file names and row fields, not the mechanical rows' text against the live file | other cards edit the workflows; a text match would block their merges | text match |
| D-4 | `mode-leader.md` does not name `moai card precheck` until the command exists | C-96 criterion 3 moved here (note 2369) but `precheck` is absent at the tip (`git grep` of `src` at 190de2d8d7: no `card precheck`) | name it later, one line |
| D-5 | the moved sentences are listed on the card before the edit and the lane goes on; the leader reverses by the release pull request | lane rule: a reversible choice is the leader's to veto, not a wait | wait for an answer |

## Result

(written after the checks)
