---
id: SPEC-20261123-check-precise-selection
title: "check affected: a precise lane selection, the full set once per train candidate, and the tests that pin skill and doc text"
status: draft
revision: 1
created: 2026-11-23
---

# check affected: a precise lane selection, the full set once per train candidate, and the tests that pin skill and doc text

## Intent

`check affected` picks one set for every caller: every target the scan list names, every consumer
of `tests/common`, every OS-tagged test. A lane's READY therefore carries the weight of a merge
train candidate run. The user decided on 2026-11-21 (test-load redesign
`~/.moai/reports/moai-adk/test-load-redesign-2026-11-21.md` §S3): the lane runs a precise
selection, the train runs the full set once per candidate, the answer says which one ran, and a
MERGE-READY accepts only the full run's PASS. Two follow-ups were added to this card later: a
change to skill or doc text selects the tests that read that text (leader note 5115, two misses of
tests that pin workflow sentences), and a range whose base lies behind the merge base with the integration tip is
reported, because it re-tests the tip's own changes (leader note 5395, the 343-target run).

What is kept from the earlier, unlanded branch `spec/check-two-modes-20261121` (12376225e5,
`SPEC-20261121-check-two-modes`): its three lane rules (scan scope, `tests/common` sub-modules, OS
area), the full set as the default with `--lane` choosing the precise set (its D-5, so every
existing caller and test keeps today's set), the `selection` answer field, and its fixture cases.
What changes: its tests called the library (`select_mode`, `Mode`), which the one-slot card is
reshaping now; these fixtures drive the command (`--list`) in temporary repositories instead, so
they hold whatever the library looks like after that card lands.

## Non-goals

- The per-item map of `tests/common/mod.rs` and a sub-command map of `src/main.rs` (report §S3b);
  a lane change to `tests/common/mod.rs` keeps selecting every consumer.
- Reusing a receipt by tree hash across a tip move (leader D-b; another card).
- The workflow guidance of the lane and leader modes: the cards that hold those files update them;
  this SPEC writes the command's operations page only.
- The ff gate: it keeps requiring `result=PASS` of a full run.
- A `selection` word on the `summary` line (its word order is pinned by existing tests) and the
  MERGE-READY full-run wording of the lane and leader guidance: here the answer's `selection` field
  and the operations page name the selection, and the cards that hold the guidance files write the
  rest. A card criterion that names the summary line or the MERGE-READY wording stays open until
  they land.

## Requirements

- REQ-1: The bare command shall select today's full set, so the full set contains every set `--lane` selects for the same range.
- REQ-2: When `--lane` is given, a scan of `tools/repo_scans.json` shall add its targets only when a changed path matches that scan's `scope` or is one of its `rules` files. · failure: none — a scan without a readable scope is left out of the lane and kept in the full set
- REQ-3: When `--lane` is given and only `tests/common/<sub>.rs` changed in `tests/common`, the selection shall hold the consumers whose source names `common::<sub>` and no other consumer. · failure: none — a change to `tests/common/mod.rs` keeps every consumer
- REQ-4: When `--lane` is given, an OS-tagged test shall be selected only when its name's first word names the area of a changed path. · failure: none — the full set keeps every OS-tagged test of the host
- REQ-5: The answers of `--list` and of a run shall name the selection in a `selection` field that reads `lane` or `full`.
- REQ-6: When a changed path under `skills/` or `docs/` appears literally in a Rust test file, the selection shall include that test in both selections. · failure: none — a test file that cannot be read adds nothing by this rule
- REQ-7: When the base of the range is a strict ancestor of the merge base of the head and the ref `--integration-ref` names, the answer shall carry a `range_note` that names that merge base and the count of extra changed paths. · failure: none — an unreadable ref gives no note and the selection is unchanged
- REQ-8: The operations page of the command shall describe the two selections, the `selection` field, the text-pin rule, the helper path-import rule and the range note.
- REQ-9: When a changed path is under `tests/common/`, the selection shall include, in both selections, the tests whose source holds `common/<rest>` for that path or, for a file directly in `tests/common/`, declares a module named like its stem. · failure: none — a test that builds the path at run time is not seen

## Acceptance

- AC-1 (REQ-1, REQ-2): a scan whose scope misses the diff is left out with `--lane` and kept without it, a scan without a scope is left out with `--lane`, a scan whose rule file changed is kept with `--lane`, and the full set contains the lane set · verify: `cargo test --locked --offline --test check_affected_modes_cli scan`
- AC-2 (REQ-3): a change to `tests/common/home.rs` selects the `home` consumer and not the `cli` consumer with `--lane`, and both without it · verify: `cargo test --locked --offline --test check_affected_modes_cli common`
- AC-3 (REQ-4): with `--lane` a host-OS test whose first word misses the diff areas is left out and one whose first word matches is kept · verify: `cargo test --locked --offline --test check_affected_modes_cli os`
- AC-4 (REQ-5): `selection` reads `lane` with `--lane` and `full` without it, in `--list` · verify: `cargo test --locked --offline --test check_affected_modes_cli selection`
- AC-5 (REQ-6): a change to `skills/moai/workflows/x.md` or `docs/a.md` selects the test whose source names that path, with and without `--lane` · verify: `cargo test --locked --offline --test check_affected_modes_cli text`
- AC-6 (REQ-7): a range from an old fork point after the tip was merged in carries `range_note` with the merge base and the extra count, and a range from the merge base carries none · verify: `cargo test --locked --offline --test check_affected_modes_cli range`
- AC-7 (REQ-8): the operations page names `--lane`, `selection`, the text-pin rule, the path-import rule and `range_note` · verify: `cargo test --locked --offline --test check_affected_modes_cli page`
- AC-8: the existing tests of the command pass unchanged · verify: `cargo test --locked --offline --test check_affected_cli`
- AC-9 (REQ-9, REQ-3): a change to `tests/common/spawn_cap.rs` selects the tests that import it by `#[path = "common/spawn_cap.rs"]` (also under another module name) or declare `mod spawn_cap;`, in both selections, and with `--lane` no `mod common` consumer that does not name the file · verify: `cargo test --locked --offline --test check_affected_modes_cli common_path`
- AC-10 (REQ-9, REQ-3): a change to a deeper helper file `tests/common/deep/x.rs` keeps every `mod common` consumer with `--lane`, adds the test that imports that file by path, and leaves the importers of other helper files out · verify: `cargo test --locked --offline --test check_affected_modes_cli common_deeper`

## Scope

- write now: this SPEC folder, the new `tests/check_affected_modes_cli.rs`
- write after the one-slot card lands and gives them back: `src/check_affected.rs`, `src/ops/check.rs`, `docs/design-review/appendix/operations/check.md`
- read-only: `tools/repo_scans.json`, `tests/check_affected_cli.rs` (held by the one-slot card), `tests/common/**`
- forbidden: card numbers in the SPEC id, the branch, commits and file text; the workflow guidance files (held by other cards)

## Plan

1. S1 fixtures (now, leader decision 6896): `tests/check_affected_modes_cli.rs` builds a temporary repository per case and runs `check affected --list` with and without `--lane`; every case fails on the tip by assertion or by the refused `--lane` flag.
2. S2 code (after the one-slot card lands): the selection takes a lane flag; the scan scope, common sub-module and OS area rules apply to the lane; the text-pin rule applies to both; `--lane` on the command; `selection` and `range_note` in the answers.
3. S3: the operations page; the measured lane and full target counts on this repository's last ten READY ranges go into the Result.

## Risks

- A scan scoped to all of `src/` still joins most lane runs; the Result records the measured drop.
- A test whose name does not say its area is missed by a lane run on another OS; the full train run catches it, and the MERGE-READY rule keeps that run the gate.
- The text-pin rule reads test sources for literal paths; a test that builds the path at run time is not seen (the doc-reference scan does the same today); the rule reads only changes under `skills/` and `docs/`.
- A change outside every scan's scope and rule files drops every scan target from the lane set: `README.md`, `Cargo.toml`, `build.rs`, `src/main.rs`, `REVIEW.md`, `CLAUDE.md` and `tools/repo_scans.json` itself are in no scope, so a lane run on such a change keeps only the area, helper and text-pin tests; the full train run catches it.
- A test that gets a `tests/common` helper by `common::*` only (a glob import) is not seen by the sub-module rule when only that sub-module changes; the full train run catches it.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The bare command keeps the full set; `--lane` picks the precise set | earlier branch D-5: every existing caller and test keeps its set, and the MERGE-READY rule reads the bare run | flip the default and add `--train` |
| D-2 | The fixtures drive the command, not the library | the one-slot card is reshaping `src/check_affected.rs`; the command's `--list` answer is the stable surface (leader decision 6896: fixtures first, no src write) | call the library once it settles |
| D-3 | The field is `selection: lane|full` beside the existing `mode: list|run` | `mode` already names list or run in the answer | rename both together |
| D-4 | The range note warns and never refuses | a refusal would stop the merge-candidate runs that need the wide range on purpose; the note makes the cost visible (planning default) | refuse unless a flag allows it |
| D-5 | The helper path-import rule (REQ-9) applies to both selections | `--lane` alone would put tests in the lane that the full set lacks, breaking REQ-1; the full set missed the `#[path]` importers too (the independent audit found it with a broken `GoFx::new`) | limit it to `--lane` and give up REQ-1 |

## Result

구현 에이전트가 2026-11-23T05:03Z 에 채우고, 독립 감사 FAIL(막는 것 하나: `#[path]` 로 불러 쓰는 도우미)을 고친 수리가 2026-11-23T06:00Z 에 다시 채웠어요. 감사가 다시 확인해요. 수리 뒤 코드 head `18922b6ee2`(base `6817eb69ad` = NeXT 병합 지점; 수리 전 코드 head `6a9e022ab5`), 이 worktree 의 `target/`, macOS arm64, 무거운 자리 `--max 4`, `CARGO_HOME=/tmp/moai-rc-cargo-c280`.

RED(`71f31169fc`): 9개 모두 단언 실패. GREEN: `cargo test --offline --locked --test check_affected_modes_cli` 9 passed(`target/c280/green1.log`).

수리 RED(head `8c9102061d` 위에 시험만 더함, `target/c280r/red.log`): 새 시험 둘이 단언 실패(`lane keeps path_user: {}`, `lane keeps deep_user: {"alpha_cli"}`). 수리 GREEN: `cargo nextest run --test check_affected_modes_cli` 11 passed(`target/c280r/green2.log` 는 lib 까지 1734 passed).

- AC-1: met — `scan_scope_decides_the_lane_and_the_full_set_keeps_every_scan`, `scan_rule_file_keeps_its_scan_in_the_lane` 통과(영수증의 같은 head 실행에서도 통과)
- AC-2: met — `common_sub_module_change_selects_only_its_users_in_the_lane`, `common_mod_change_keeps_every_consumer_in_the_lane` 통과
- AC-3: met — `os_tagged_test_joins_the_lane_by_area_only` 통과
- AC-4: met — `selection_field_names_lane_or_full` 통과
- AC-5: met — `text_change_selects_the_tests_that_name_the_file` 통과
- AC-6: met — `range_from_an_old_fork_point_carries_a_note` 통과
- AC-7: met — `page_describes_the_two_selections` 통과
- AC-8: met — `check_affected_cli` 와 `check_affected_one_slot` 전부 통과(수정 없음): 수리 뒤 head 에서 `moai check affected 6817eb69ad..18922b6ee2 --max 4`(감싸지 않음, 한 번)가 `targets=52 scans=listed:37 diff_check=0 fmt=0 ops_rustfmt=0 clippy=0 test=0 result=PASS`, nextest 2461 passed, 0 skipped(`target/c280r/receipt/05-test.log`, 수리 전 head 는 2459 passed)
- AC-9: met — `common_path_import_keeps_its_users_in_both_selections` 통과(`#[path = "common/spawn_cap.rs"]`·다른 모듈 이름·`mod spawn_cap;` 가 두 선택에 들고, 레인에는 파일 이름을 안 부르는 `mod common` 소비자가 안 들어요). 실제 기록으로도 확인: `check affected a53c9c7edd^..a53c9c7edd --list --lane`(`tests/common/go_project_real_run.rs` 변경)에 `go_project_real_run_cli`·`go_settings_cli`·`go_state_preview_cli`·`project_update_hooks_cli`·`update_end_to_end_cli` 가 들어요
- AC-10: met — `common_deeper_path_keeps_every_consumer_and_its_path_users` 통과(더 깊은 경로 `tests/common/deep.rs/x.rs`: 레인에도 `mod common` 소비자 유지, 그 파일을 경로로 부르는 시험 추가, 다른 도우미의 경로 시험은 빠짐)

뮤턴트 10개(지킴 코드를 하나씩 뺌) 모두 KILLED(`target/c280/mutants.log`): 스캔 scope 지킴(M1)·rules 파일 일치(M2)·`common` 하위 모듈 갈래(M3)·`mod.rs` 전체 유지(M4)·호스트 OS 지킴(M5)·글 고정(M6)·`selection` 낱말(M7)·`range_note` 조상 지킴(M8)·`range_note` 꺼짐(M9)은 `check_affected_modes_cli` 의 9개 중 하나 이상이 잡고, 식별자 전체 일치(M10)는 모듈 단위 시험이 잡아요. 수리의 새 규칙은 뮤턴트 4개가 모두 KILLED(`target/c280r/mut-*.log`, 소스 복원 확인): `common/<경로>` 글자 갈래를 뺌(MA, 2개 실패)·`mod <이름>` 갈래를 뺌(MB, 1개)·레인에만 적용(MC, 2개; 레인이 전체의 부분집합이어야 한다는 단언이 잡음)·`["common", file]` 갈래를 `["common", file, ..]` 로 넓힘(MD, 1개; 감사의 O2). 수리 전 뮤턴트 10개는 이 수리에서 다시 돌리지 않았어요(고친 것만 다시 확인).

측정(`target/debug/moai-cli check affected <병합>^1..<병합> --list`, NeXT 최근 첫 부모 병합 10개, 수리 뒤 새 바이너리): 대상 수 합계 전체 1304 → 레인 831(−36%). 병합별 전체/레인(`95a3a0e1df`부터 `0dfdfb31dc` 순): 80/69, 385/385, 49/21, 79/67, 90/62, 390/129, 48/5, 54/25, 71/41, 58/27. 앞의 측정과 겹치는 여섯 병합(90/62 에서 58/27 까지)은 값이 같아요. 앞 측정의 합계(992 → 447, 열 개가 달랐음)와 직접 견주지 않아요.

실행 안 함·한계: Linux·Windows, `summary` 줄에 `selection` 낱말은 넣지 않았어요(답의 필드만; 줄의 낱말 순서를 고정한 시험을 건드리지 않으려고). 레인 선택은 `tools/repo_scans.json` 자체와 스캔 범위 밖 변경(`README.md`, `Cargo.toml`, `build.rs`, `src/main.rs`, `REVIEW.md`, `CLAUDE.md`)에서 스캔 대상을 모두 빼요(Risks 에 적음); `common::*`(glob)로만 하위 모듈을 쓰는 시험은 하위 모듈 변경에서 빠져요. 모두 전체 선택(열차 실행)이 잡아요. 이 SPEC 의 `status` 는 draft 그대로예요(승인은 이 레인이 하지 않아요). 워크플로 가이드(mode-agent/leader)와 MERGE-READY 문구는 다른 카드가 쥔 파일이라 이 카드에서 바꾸지 않았어요: 그래서 카드 기준 1 의 "요약 줄이 모드를 말함·MERGE-READY 는 전체 실행" 중 요약 줄 낱말과 MERGE-READY 문구는 이 head 에서 되지 않았고(답의 `selection` 필드와 도움말 페이지만 있어요), 그 기준은 일부만 met 예요; SPEC 은 이를 Non-goals 로 두었어요. 수리 뒤 head 의 점검은 수리 커밋 두 개의 코드와 시험 위에서 돌렸고, 이 Result 커밋은 SPEC 파일만 바꿔요(`diff --check`만 다시 돌림).
