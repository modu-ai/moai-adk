---
id: SPEC-20261121-statusline-detached
title: 상태줄 detached 표시와 조회 실패 n/a, 그리고 사용자가 고른 2·3줄 숫자만 정리 (영어 전용 화면)
status: draft
revision: 1
created: 2026-11-21
---

## Intent

리더 결정(카드 노트 5373, 사용자 확인 2026-11-21T09:22:55Z): statusline의 `repo_groups`가
브랜치 칸이 빈 문자열인(detached HEAD) 때 저장소·GitHub 개수까지 통째로 숨기는 원인을 고쳐요.
지금 코드(`src/statusline/render.rs:376`)는 `branch.is_empty()`면 `None`을 돌려 2줄의
repo·forge·branch 세그먼트가 모두 사라져요. 규칙: 브랜치가 있으면 실제 이름, detached면 짧은
SHA, 저장소·remote·개수 조회 실패는 그 항목만 `n/a`(0으로 표시 금지), GitHub 줄은 항상 보임. 화면 글자는 영어만이에요(사용자 결정, 감사 노트 5573 B1 에서 리더가 확정): 모르는 값은 `미확인` 대신 `n/a` 예요.

사용자가 고른 2·3줄 정리(2026-11-20T21:5xZ mock-up 선택)와 그 뒤의 최종
사용자 확정(2026-11-21, 리더 정정: "그래프 제거하고 숫자만" 재확인 — 중간에 전달된
"막대 유지"는 오해였음): 2줄은
`📡 <repo>, ⬇️ n 🔀 n | 🅱️ <branch> ↑n ↓n +n | 🌳 <worktree>`(worktree 라벨이 3줄 끝에서
2줄 끝으로), 3줄은 `🗂️ n ▶️ n 🔃 n ✅ n | 🪫 CTX n% | 🔋 Week n% (resets in 6d 21h)` —
막대 그래프와 콜론 없이 숫자만, `(⚠️ /clear)`·`(🛑/clear!)` 안내 없음, Week에 리셋
카운트다운. 1줄은 그대로. `↑N`은 이미 착지된 SPEC-20261120-statusline-ahead의 동작이라
이 SPEC은 새로 요구하지 않아요.

관계: 착지된 레이아웃 작업은 SPEC-20261120-statusline-layout(draft)이고, 이 SPEC은 그 위의 수리와
재정리예요. `↑N ↓N` 과 숫자만 정리의 내용이 이 SPEC 하나로 실려요(리더 묶음 지시).

감사 FAIL 수리(노트 5573, 5734): 화면은 영어만이라 모르는 값은 `n/a` 로 쓰고(`📡 n/a`, `⬇️ n/a 🔀 n/a`,
`CTX n/a`, `Week n/a`), 퍼센트가 없는 Week 는 `0%` 가 아니라 `Week n/a`, 모르는 창에는 배터리 아이콘이
없어요. 두 창의 퍼센트는 0..100 으로 자르고, 막대 코드와 막대 낱말, 렌더러 주석의 카드 번호를 지우고,
제품 문서(`docs/rust-product-path.md`)를 숫자만 모습으로 고쳐요. 시험은 Rust 로만 써요(사용자 규칙: 새 파이썬
시험 없음, 파이썬 골든 도구는 늘리지 않아요).

## Requirements

- REQ-1: When git reports a repository and the branch name is empty (detached HEAD), the renderer shall collect the short HEAD SHA and show it in the branch slot — `🅱️ <short sha>` with the repository segment and its forge counts still shown. · failure: none — a missing SHA leaves the branch slot out and the repository segment stays (REQ-8).
- REQ-2: When git reports a repository but the repository name or its remote cannot be resolved, line 2 shall show `📡 n/a` in the repository slot instead of hiding the segment. · failure: none — a resolved name renders as the name; the unknown form is only for the failed lookup.
- REQ-3: When the forge counts cannot be read (`available` false), the counts shall render `⬇️ n/a 🔀 n/a` and never `⬇️ 0 🔀 0`, while a user switch that suppresses the counts keeps them hidden. · failure: none — a successful read renders the numbers; the unknown form is only for the failed read.
- REQ-4: When the worktree label is present, the renderer shall close line 2 with it after the branch slot and any pull-request link, leaving line 3 without a worktree segment. · failure: none — an absent label closes nothing; the main checkout has none.
- REQ-5: The context and weekly windows shall render as numbers only — `<icon> CTX <n>%` and `<icon> Week <n>%`, the battery icon switching to `🪫` above 70% and `🔋` otherwise, with no bar graph and no trailing colon.
- REQ-6: When the weekly window carries a reset time, the weekly segment shall append `(resets in <d>d <h>h)` at a day or more, `(resets in <h>h <m>m)` under a day and `(resets in <m>m)` under an hour, measured from the render's clock. · failure: none — without a reset time the segment ends at the percentage.
- REQ-7: The renderer shall not render the handoff hint `(⚠️ /clear)` or `(🛑/clear!)` on any line.
- REQ-8 (REQ-1): IF the branch name is empty and no short SHA is available, THEN line 2 shall still show the repository segment and its forge counts and omit only the branch slot.
- REQ-9: When a window's percentage is unavailable (the window is absent, its percentage is `null`, or its object is empty), the renderer shall show `CTX n/a` or `Week n/a` without a battery icon and never as `0%`. · failure: none — an available percentage renders with its icon.
- REQ-10: The renderer shall show a window's percentage no lower than 0 and no higher than 100.
- REQ-11: The renderer shall print only English text on every line of the screen.
- REQ-12: The statusline sources shall carry no bar-graph code or wording and no card number in the comments of `render.rs`.

실패 경로: REQ-1 → REQ-8, REQ-9.

## Acceptance

- AC-1 (REQ-1, REQ-8): verify: `cargo test --offline --locked detached_head_shows_the_short_sha_and_keeps_the_repository` — 가지가 빈 head_sha 붙은 상태의 2줄이 `📡 owner/repo, ⬇️ … 🔀 … | 🅱️ a1b2c3d …`를 말하고, SHA 없는 상태에선 repo 세그먼트만 남아요.
- AC-2 (REQ-2, REQ-3): verify: `cargo test --offline --locked lookup_failures_show_unknown_not_zero` — repo가 None이고 github가 unavailable인 2줄이 `📡 n/a … ⬇️ n/a 🔀 n/a`를 말하고 0을 갖지 않아요.
- AC-3 (REQ-4, REQ-5, REQ-6, REQ-7): verify: `cargo test --offline --locked the_user_example_gives_exactly_the_three_lines` — 사용자가 고른 예시(worktree c135, resets in 6d 21h)가 세 줄로 정확히 나오고, 막대·콜론·/clear 힌트가 어느 줄에도 없어요.
- AC-4 (REQ-6): verify: `cargo test --offline --locked the_reset_countdown_uses_the_three_formats` — 6d21h·5h30m·40m 입력이 세 포맷으로 나와요.
- AC-5 (REQ-1…REQ-8): verify: `cargo test --offline --locked --test statusline_golden` — 골든을 기존 `tools/statusline_golden.py --write` 스위치로 재생성하고 모든 변경 골든을 사람이 리뷰했어요(노트에 나열). 파이썬 도구는 늘리지 않아요.
- AC-6 (REQ-1…REQ-8): verify: `moai spec check SPEC-20261121-statusline-detached` → errors 0, warnings 0.
- AC-7 (REQ-9): verify: `cargo test --offline --locked an_unknown_window_says_n_a_without_a_battery_icon` — 모르는 CTX 와 퍼센트 없는 Week 가 `CTX n/a`·`Week n/a` 로 나오고 아이콘과 `0%` 가 없어요.
- AC-8 (REQ-10): verify: `cargo test --offline --locked window_percentages_stay_between_zero_and_one_hundred` — 음수와 100 초과 입력이 0 과 100 으로 나오고, 0·1·99·100·101 경계가 맞아요.
- AC-9 (REQ-11): verify: `cargo test --offline --locked no_line_of_the_screen_carries_korean_text` — 모르는 값·조회 실패·세 줄 전부를 그린 화면 어디에도 한글이 없어요.
- AC-10 (REQ-12): verify: `cargo test --offline --locked --test statusline_numbers_only the_renderer_sources_carry_no_bar_leftovers_and_no_card_numbers` — `render.rs`·`segments.rs` 에 막대 문자와 막대 낱말이 없고 `render.rs` 주석에 카드 번호가 없어요.
- AC-11 (REQ-5, REQ-7): verify: `cargo test --offline --locked --test statusline_numbers_only the_statusline_doc_describes_the_numbers_only_look` — 제품 문서의 3줄 설명이 숫자만 모습(막대·콜론·`/clear` 안내 없음, 리셋 카운트다운)을 말해요.
- AC-12 (REQ-9): verify: `cargo test --offline --locked --test statusline_numbers_only a_context_window_with_a_null_percentage_says_ctx_n_a` — 사용률이 `null` 이거나 비어 있는 `context_window` 가 stdin 으로 오면 실제 바이너리가 `CTX n/a` 를 그리고 `0%` 도 배터리 아이콘도 없어요. 단위 시험 `a_context_window_with_no_percentage_and_no_usage_is_unavailable` 가 읽기 값 쪽을 고정하고, 골든 사례 `a16-context-percentage-null` 이 화면을 고정해요.
- AC-13 (REQ-10): verify: `cargo test --offline --locked --lib huge_context_numbers_do_not_overflow` — 아주 큰 토큰 수·창 크기·사용률이 들어와도 디버그 빌드에서 죽지 않고 퍼센트가 0..100 에 머물러요.
- AC-14 (REQ-12): verify: `cargo test --offline --locked --test statusline_numbers_only the_statusline_test_files_carry_no_bar_wording` — 상태줄 시험 파일 머리말·주석에도 막대 낱말이 없어요.

## Scope

- `src/statusline/**`(render.rs, state.rs 수집, segments.rs), `tests/statusline_*.rs`(새 `statusline_numbers_only.rs` 포함),
  `tests/fixtures/statusline*/**`(재생성), `docs/rust-product-path.md`(3줄 설명), 이 SPEC.
- 이 카드의 이전 SPEC SPEC-20261120-statusline-layout은 그대로(착지된 작업의 기록).

## Plan

1. 시험 먼저: 유닛 4종을 목표 문장으로 RED(detached·미확인·사용자가 고른 예시·카운트다운 경계).
2. 수집: symbolic-ref 실패 시 rev-parse --short 한 번(GitStatus.head_sha).
3. 렌더: repo_groups의 None 원인 제거(REQ-1·2·3·8), git_line에 worktree 라벨 편입(REQ-4),
   bars_line 숫자화·카운트다운(REQ-5·6·7).
4. 골든 재생성 + 전수 리뷰, fmt/clippy/--max 2 슬롯 검사, spec check.
5. 감사 FAIL 수리: 영어 전용 `n/a`(REQ-2·3·9·11), 퍼센트 자르기(REQ-10), 막대 잔재·카드 번호 주석(REQ-12), 문서(AC-11). 시험 먼저(RED), 그다음 코드·문서.
6. 재감사 수리: 사용률이 `null` 인 컨텍스트 창이 `CTX 0%` 로 나오던 것(읽을 값이 없으면 읽기 값을 "없음"으로 돌려줘요, AC-12), 큰 수 곱셈 넘침(AC-13), 안 쓰는 `Theme.warn`·`Theme.alert` 삭제, 시험 파일의 막대 낱말과 SPEC 의 카드 번호 삭제(AC-14). 시험 먼저(RED), 그다음 코드, 골든 사례 하나를 데이터로 더해요(파이썬 도구는 그대로).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 모르는 창에는 배터리 아이콘을 붙이지 않고 `CTX n/a`·`Week n/a` 로 쓴다 | 감사 N2(모르는 CTX 에 아이콘 없음)와 N1(퍼센트 없는 Week 는 `Week n/a`) | 아이콘을 다시 붙이기 |
| D-2 | 렌더러 주석의 카드 번호는 `render.rs` 만 지운다 | 감사 N6 이 `render.rs` 줄을 지목했고, 다른 파일의 옛 주석까지 지우면 이 수리의 범위를 넘어요(후속) | 다른 파일도 지우기 |
| D-3 | 막대 잔재 시험은 막대 문자·`gradient`·`CTX bar` 같은 낱말 목록으로 고정한다 | 감사 N3 | 낱말 목록 줄이기 |

## Result

Filled in by the implementation lane on 2026-11-21T20:23:09Z (`date -u`); the audit checks it again. Branch `spec/20261121-statusline-detached-r2`, worktree `/Users/goos/.moai/worktrees/moai-adk/statusline-numbers`; the audited head of the first repair round was `c969197d2e` (audit FAIL, B1 and B2 and N1..N6). This round adds, on top of the local tip `6e1bedb4f3`: the tip merge `7a5ab2dc32`, this SPEC `3c279bd13f`, the red tests `b64000ae4d`, the code `f99c775900`, the goldens `6c8333bf03`, the product document `201182d73d` and one fix commit `1497aa98ee`. Evidence files are in this worktree's `target/notes/`.

Red first (`red.out`, the six integration targets, no name filter): the new `statusline_numbers_only` tests and the sibling suites failed as planned (cards 1, counts_refresh 1, git_env 2, numbers_only 2, worktree_label 4 failed; tag_order passed).

- AC-1: met — the unit tests of `statusline::` pass after the change: `cargo test --locked --offline --lib -- statusline::` 151 passed, 0 failed (`golden-train.out`, head `f99c775900`); the same tests are inside the passing `check affected` below.
- AC-2: met — `lookup_failures_show_unknown_not_zero` expects `📡 n/a, ⬇️ n/a 🔀 n/a | 🅱️ main ↑5 ↓2 +3` and passes; the four sibling suites that said the Korean word now say `n/a` and pass.
- AC-3: met — `the_user_example_gives_exactly_the_three_lines` passes unchanged in the `statusline::` run.
- AC-4: met — `the_reset_countdown_uses_the_three_formats` passes unchanged in the `statusline::` run.
- AC-5: met — the Rust test `statusline_matches_every_claude_golden` passes inside `check affected`; the goldens were rewritten with the existing `tools/statusline_golden.py --write` switch (`golden-write.out`, exit 0) and all 48 changed golden files were read: 16 cases times three variants (plain, ansi, ansi256): `a01-minimal`, `a03-context-null`, `a06-invalid-json`, `a09-top-level-null`, `a10-unpaired-surrogate`, `a12-invalid-utf8`, `a13-duplicate-top-level-keys`, `a15-case-folded-keys`, `b04-template-version`, `b05-task-state` lose the battery icon of the unknown CTX and say `CTX n/a`; `b01-git-clean`, `b02-git-dirty-ahead`, `d01-leader-mode`, `d02-leader-expired`, `d03-title-name`, `d04-agent-name` say `n/a` where they said the Korean word. In the colored variants only that icon and that word differ (the color codes are the same, checked on `a01-minimal.ansi` and by a mapped diff of all colored files). `cases.json` changed only by the recorded sha256 of the binary that wrote the goldens.
- AC-6: met — `spec check SPEC-20261121-statusline-detached` reports 0 errors and 0 warnings; `spec check --closure` leaves only `closure_status`.
- AC-7: met — `an_unknown_window_says_n_a_without_a_battery_icon` and `rate_windows_are_optional_and_an_unknown_window_says_n_a` pass; mutants M1 (the icon back on an unknown window) and M3 (a week without a percentage shows `0%` again) are killed by them.
- AC-8: met — `window_percentages_stay_between_zero_and_one_hundred` passes (-5, 0, 1, 99, 100, 101, 1000000 for Week and 0, 100, 101, 5000 tokens for CTX); mutant M2 (no clamp) is killed.
- AC-9: met — `no_line_of_the_screen_carries_korean_text` passes over four statuses in plain and colored output; mutant M4 (the unknown text Korean again) is killed by several tests; no mutant was written for this test alone.
- AC-10: met — `the_renderer_sources_carry_no_bar_leftovers_and_no_card_numbers` passes (no block characters, none of `gradient`, `10-cell`, `CTX bar`, `context bar`, `stops`, no whole word `bar` or `bars`, no card number in `render.rs`); mutant M5 (a card number back in a comment) is killed. The unused theme field `stops` is removed.
- AC-11: met — `the_statusline_doc_describes_the_numbers_only_look` passes; the doc test has no mutant of its own.

### Checks

`check affected` (the branch's own build, `--max 2`, `CLAUDE_CODE_SESSION_ID` unset): first run on head `201182d73d` FAILED (diff_check=2 for a blank line at the end of this SPEC, test_targets=101 for two of my own new tests: the doc test still looked for the old item title and `render.rs` still had a test named with the word bars); fixed in `1497aa98ee`. Second run, base `f566ebf268` (the tip when it started; the tip has moved since, see below), head `1497aa98ee`: `check-affected base=f566ebf268fd head=1497aa98eec8 targets=51 py=9 scans=listed:42 diff_check=0 fmt=0 clippy=0 test_lib=0 test_bins=0 test_targets=0 build_cli=0 python=0 result=PASS`. Five mutants, one at a time on the committed file with the bytes restored from memory (`target/notes/mutants.out`): 5 of 5 killed, last line `restored: clean = True`. `git diff --check 6e1bedb4f3 HEAD` is 0 and `cargo fmt --check` is 0.

Two commits of this round were made with the single-command pre-commit override `SKIP_MOAI_PRECOMMIT=1` because the shared hook ran a command the installed `moai` did not have (notes 5780 and 5784): `7a5ab2dc32` and `3c279bd13f`. No Go file was staged and the heavy gate was opted out, so no check that used to run was skipped. Every later commit went through the hook.

### Not run

- The statusline in a real Claude Code session with live data: not run.
- The old card-number comments in `state.rs`, `counts.rs` and the test files are not removed (D-2); `미확인` in `src/goal_block.rs`, `src/close_check.rs` and `tests/goal_block_cli.rs` belongs to other features and is untouched.
- `check affected` was not run again on a merge with the tip as it is now (`a01c444e9a`): the tip moved after the run by 15 files and none of them is a file of this change; the PASS above is cited for the merge (no overlap).
- A second run of the same tests on the same tree: not repeated, except the one rerun that the failed first run made necessary.
