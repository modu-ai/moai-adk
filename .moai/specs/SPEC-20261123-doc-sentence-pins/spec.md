---
id: SPEC-20261123-doc-sentence-pins
title: "Two document sentences are pinned by tests: the leader condition of the next-card rule, and the cargo-only reading of the repo-scans sentence"
status: draft
revision: 1
created: 2026-11-23
---

# Two document sentences are pinned by tests: the leader condition of the next-card rule, and the cargo-only reading of the repo-scans sentence

## Intent

두 문장이 글에는 있는데 그 문장을 지우거나 모호하게 둬도 어떤 테스트도 빨개지지 않아요. 둘 다 이미 착지한 변경의 감사에서 나온 빈틈이고, 이 SPEC 은 그 둘만 닫아요. 재서 확인한 현재 상태는 아래와 같아요(2026-11-23, 이 가지의 `737df06fdd` 트리).

- 다음 카드 규칙의 리더 조건: `docs/lane-rules.md:8` 의 "…, and only when no leader is registered; with a registered leader the lane waits for that leader's dispatch." 절을 지워도 `lane_next_rules`·`repo_claude_md`·`lane_completion_handshake` 가 통과해요. 그 절을 읽는 테스트가 없어서예요. 레인이 리더 배차를 기다리는 근거가 이 한 절이라, 지워지면 레인이 리더 몰래 다음 카드를 받는 규칙이 돼요.
- repo-scans 문장: `skills/moai/references/flows/next-flow.md:56` 이 "`moai repo-scans --refs <tip> <head>` 가 고르는 테스트를 돌린다"고만 말해요. `moai repo-scans` 의 결과에는 Python 검사 줄도 들어 있어서(표에 그대로 남겨 둠), 이 문장만 읽은 세션이 그 줄까지 돌릴 수 있어요. 착지 명령은 `cargo` 로 시작하는 검사만 요구해요(`SPEC-20261123-land-repo-scans` D-7). 카드가 가리킨 줄 번호(`next-flow.md:43`)는 그 뒤 문단이 늘어서 56 행으로 밀렸어요.
- 크기: `next-flow.md` 는 7974 B 이고 한도는 8192 B 예요(`FLOW_REF_LIMIT`). 문장은 늘리지 않고 바꿔 써요.

## Non-goals

- 규칙 자체를 바꾸는 일: 리더 조건도, `cargo` 만 돌린다는 착지 규칙(`SPEC-20261123-land-repo-scans` D-7)도 그대로예요. 이미 있는 말을 고정하고 모호한 한 문장을 분명히 하는 일만 해요.
- `moai repo-scans` 의 출력이나 `tools/repo_scans.json` 표를 바꾸는 일(Python 줄은 표에 남아요), 착지 명령 코드.
- `docs/leader-rules.md` 와 `next-flow.md` 의 "Final flow" 문단(같은 뜻의 다른 문장)을 같이 고정하는 일. 필요하면 따로 카드로 해요.
- Python 테스트를 돌리는 일, 전체 `cargo test`.
- 이 변경이 더하는 문장에 카드·Issue·풀 리퀘스트 번호를 넣는 일(넣지 않아요).

## Requirements

- REQ-1: The lane rules shall carry the clause that a lane takes the next card only when no leader is registered and waits for a registered leader's dispatch.
- REQ-2: IF the leader clause is removed from the lane rules, THEN a test shall fail.
- REQ-3: The flow reference's repo-scans bullet shall tell the session that briefs an audit to run only the `cargo` checks of the scans `moai repo-scans` picks and to leave out a Python check line.
- REQ-4: IF the bullet does not name the cargo-only reading, THEN a test shall fail.
- REQ-5: The flow reference shall stay within its byte budget of 8192 bytes.
- REQ-6: The skill's source manifest shall match the files after the change.

## Acceptance

- AC-1 (REQ-1, REQ-2): the lane-rules next-card bullet carries both halves of the leader condition, and the same text without the clause is reported as a problem by the checker that the test uses · verify: `cargo nextest run --locked --offline --test lane_next_rules -E 'test(next_card_leader_clause)'`
- AC-2 (REQ-3, REQ-4): the repo-scans bullet says "run only their `cargo` checks" and "leaving out a Python line", and the same bullet without either phrase, or with the meaning reversed (run all their checks, including a Python line), is reported as a problem · verify: `cargo nextest run --locked --offline --test next_flow_skill_scan -E 'test(land_scans_the_flow_)'`
- AC-3 (REQ-5): the flow reference is 8192 bytes or less · verify: `cargo nextest run --locked --offline --test next_flow_skill_scan -E 'test(land_scans_the_flow_names_repo_scans_the_merge_tree_and_the_receipt_step)'`
- AC-4 (REQ-3, REQ-5): the shipped-text scans and the stage and size tests still pass on the changed text · verify: `cargo nextest run --locked --offline --test mode_workflow_steps --test next_flow_skill_scan --test source_manifest_facts --test skill_structure --test workflow_size_budget --test lane_next_rules --test repo_claude_md`
- AC-5 (REQ-6): the manifest matches the package · verify: `python3 -I tools/source_manifest.py check`

## Scope

- write: `.moai/specs/SPEC-20261123-doc-sentence-pins/**`, `docs/lane-rules.md` (only if a test needs a wording fix; the clause itself stays), `skills/moai/references/flows/next-flow.md` (the repo-scans bullet only), `tests/lane_next_rules.rs`, `tests/next_flow_skill_scan.rs`, `skills/moai/references/source-manifest.json` (written last with `tools/source_manifest.py write`)
- read-only: `tests/repo_claude_md.rs`, `tests/lane_completion_handshake.rs`, `.moai/specs/SPEC-20261123-land-repo-scans/spec.md`, `tools/repo_scans.json`, `src/card/land/scans.rs`
- forbidden: `CLAUDE.md`, `docs/leader-rules.md`, the Final flow paragraph of `next-flow.md`, any `src/` file

## Plan

1. RED first: add the two tests and run them on the tip. The lane-rules test passes on the real text and its mutant check proves the clause is read; the repo-scans test fails on the real bullet because it names neither `cargo` nor the Python line.
2. Rewrite the repo-scans bullet in place so that it names the cargo-only reading and the Python line without growing the file beyond the budget; keep every part the existing `land_scans_the_flow_names_repo_scans_the_merge_tree_and_the_receipt_step` test needs.
3. Run the five scans the dispatch names (`mode_workflow_steps`, `next_flow_skill_scan`, `source_manifest_facts`, `skill_structure`, `workflow_size_budget`) with the two lane tests, as one nextest filterset.
4. Write the manifest row with `tools/source_manifest.py write` and run its check.
5. Run guard mutants for each new test (remove the clause; remove the `cargo` word), restore, confirm a clean tree.
6. Ask for an independent audit; the implementer does not audit.

## Risks

- A shipped-text scan refuses a word of the new sentence (a Python word, a banned phrase): the sentence is run through `next_flow_skill_scan` before the first commit and reworded if it is refused.
- The byte budget: the new sentence replaces the old one in place and the size is measured after the edit (`wc -c`), not estimated.
- `source-manifest.json` may be held by another card: its row is written last, under a line lease if one is needed.
- The lane-rules test reads one bullet of `docs/lane-rules.md` by its first words; a rewording of that bullet by a later card must move the pin with a reason, not delete it.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The leader clause is pinned in `tests/lane_next_rules.rs`, which already tests next-card text | The card's write scope names it; no new test file | Move the test and its mutant check to another file |
| D-2 | The repo-scans bullet is changed in place, one bullet, no new line | `next-flow.md` is at 7974 of 8192 bytes; the user's dispatch says rewrite, do not grow | Restore the earlier bullet from git |
| D-3 | The `cargo`-only reading follows the landing command's own rule (`SPEC-20261123-land-repo-scans` D-7), so the guidance matches what `moai card land` requires | `CLAUDE.md` "Tests are Rust only"; D-7 of the land-repo-scans SPEC | Drop the words and the test |
| D-4 | The SPEC stays `draft`: only the user's selection approves it | `internal/sdd` item 6 | None needed |

## Result

2026-11-23 (UTC), 구현 레인이 채웠고 감사가 다시 확인해요. 가지 `spec/doc-sentence-pins-20261123` 은 통합 끝점 `737df06fdd` 에서 갈라졌어요. SPEC 과 RED `9a460dca6b`, 구현 `4192a0d92c`. 실행 기록은 이 워크트리의 `target/c400/` 의 `red2.out`·`green2.out`·`mutant1.out`·`mutant2.out` 이에요(카드가 열려 있는 동안만 둬요).

- 수리 (감사 `a6410159fd` 의 B1, 같은 가지 `771d3172c1`): 핀이 낱말 두 묶음만 봐서 56 행을 반대 뜻("run all their checks, not only their `cargo` checks, ... including a Python line")으로 바꿔도 통과했어요(M2 exit 0). 이제 `repo_scans_cargo_only_problems` 가 "run only their `cargo` checks" 와 "leaving out a python line" 을 통째로 고정하고, `land_scans_the_flow_cargo_only_mutants_are_caught` 에 뒤집은 문장 뮤턴트("meaning reversed")를 더했어요. 문장은 "run only" 를 "run all ... not only", "leaving out" 을 "including" 으로 바꿔 만들어서 문장을 다시 써도 따라가요. RED: 뮤턴트만 먼저 더하고 옛 핀으로 돌리면 "meaning reversed: the pin passes without the sentence" 로 exit 100.
- 계획과 달라진 것: 카드가 가리킨 `next-flow.md:43` 은 그 뒤 문단이 늘어서 56 행으로 밀려 있었어요. 수리에서 같은 불릿을 다시 써서 파일 크기가 8022 B 에서 7974 B 로, 원래 크기 그대로가 되었어요(`wc -c`).
- AC-1: met — `next_card_leader_clause_is_in_the_lane_rules` 와 `next_card_leader_clause_mutants_are_caught` 가 통과해요(절 지우기, 뒤쪽 반절 지우기, 앞쪽 반절 지우기, 불릿 지우기 네 뮤턴트를 모두 잡고 각 뮤턴트가 글을 실제로 바꾸는지도 확인해요). 가드 뮤턴트: `docs/lane-rules.md:8` 의 리더 절을 지우면 이 둘이 exit 100 으로 실패하고, 같은 때 `repo_claude_md`·`lane_completion_handshake` 는 통과해요(카드가 말한 빈틈). 복원했고 `git status` 가 깨끗해요.
- AC-2: met — `land_scans_the_flow_runs_cargo_checks_only` 와 `land_scans_the_flow_cargo_only_mutants_are_caught` 가 통과해요. RED 에서는 첫째가 "only their `cargo` checks"·"python line" 이 없다고 실패했어요. 가드 뮤턴트: next-flow 의 새 말을 되돌리면 둘 다 exit 100, 56 행을 반대 뜻으로 바꾸면 첫째가 exit 100(M1), 핀을 옛 낱말 둘로 약하게 하면 둘째가 exit 100(M2).
- AC-3: met — `next-flow.md` 는 7974 B 이고 `land_scans_the_flow_names_repo_scans_the_merge_tree_and_the_receipt_step` 와 새 테스트가 8192 B 이하를 확인해요.
- AC-4: met — `mode_workflow_steps`·`next_flow_skill_scan`·`source_manifest_facts`·`skill_structure`·`workflow_size_budget`·`lane_next_rules`·`repo_claude_md` 의 nextest 실행이 116 중 116 통과(exit 0).
- AC-5: met — `python3 -I tools/source_manifest.py write` 가 `references/flows/next-flow.md` 한 줄만 바꿨고 `check` 가 `ok: true`, `changed: []` 예요.
- 남은 일: 없음. `docs/leader-rules.md` 와 "Final flow" 의 같은 뜻 문장은 이 SPEC 의 범위 밖이에요.
