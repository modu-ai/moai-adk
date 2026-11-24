---
id: SPEC-20261123-statusline-order-bars
title: "The status line puts the versions on the flow line and brings the limit bars back"
status: draft
revision: 2
created: 2026-11-23
parent: SPEC-20261122-statusline-limits-left
---

# 상태줄: 버전을 흐름 줄 앞으로, 한도 줄에 막대 그래프 다시

## Intent

사용자 지시예요(2026-11-23, 리더 세션, 원문): `🤖 2.1.330 | 🗿 4.2648.20 | 🗂️ 176 → 🟡 6 → 📨 13 → 🛬 8 → ✅ 6 | 🚧 7 🙋 9 | 🎯 C-287 -> 이 배열로 다시 변경`, `🔋 CW 73% | 🔋 5H 49% (43m) | 🔋 7D 87% (3d 22h) -> 그래프 다시 추가`.

바꾸는 것은 둘이에요.

1. **둘째 줄(흐름 줄)의 순서**: `🤖 호스트 버전 | 🗿 moai 버전 | 🗂️ 흐름 | 🚧 🙋 | 🎯 카드`. 지금은 `🎯 <카드> <단계> | 🗂️ 흐름 | 🚧 🙋` 이고 두 버전은 셋째 줄 끝에 있어요. 두 버전이 셋째 줄에서 빠지고 둘째 줄 앞으로 와요.
2. **셋째 줄(한도 줄)**: 🔋 `CW`·`5H`·`7D` 만 남기고(`💰` 는 한도 정보가 없을 때만 나오던 그대로예요), 각 창의 값 옆에 막대 그래프를 다시 그려요. 막대 모양은 지워지기 전(`b6e8e9c977` 의 부모 커밋) 것을 따라요: 10칸, 채운 칸 `█` 와 빈 칸 `░`, 채운 칸 수는 `퍼센트 × 10 / 100` 내림, 위치는 라벨과 숫자 사이(`🔋 CW ███████░░░ 73%`), 콜론 없음. 숫자(남은 비율), 되돌림 시간, 배터리 아이콘 경계는 그대로예요.

**이 지시가 바꾸는 이전 결정**: `SPEC-20261121-statusline-detached` 의 REQ-5("숫자만, 막대 없음", 2026-11-21 사용자의 마지막 결정)는 이 SPEC 의 REQ-3 로 **대체**돼요. 그 SPEC 은 고치지 않고 이 SPEC 이 대체 관계를 적어요(REQ-7, 7 은 그 SPEC 의 REQ-7 과 무관). `SPEC-20261122-statusline-limits` 의 줄 3 REQ-1(항목 순서에 두 버전 포함)과 `SPEC-20261122-statusline-limits-left` 의 REQ-4(폭 사다리에서 버전이 가장 먼저 빠짐)도 이 SPEC 의 REQ-1·REQ-2·REQ-5 로 바뀐 부분만 대체돼요. 그 밖의 규칙(숫자는 남은 비율, `n/a`, 숨김 옵션, 한 칸 간격, 카운트다운 모양, 폭 예산)은 그대로예요.

개정 2(사용자 확정 지시, 같은 날 카드 노트의 리더 결정으로 전달): **첫 줄(누가·어디)** 도 같이 바꿔요. 순서와 표기는 `🧠 <모델> <effort>·t | 📡 <owner/repo>, ⬇️ <이슈> 🔀 <PR> | 🅱️ <로컬 통합 브랜치>(🏷️ <그 ref tip 7자리>) | 🌳 <체크아웃된 브랜치>` 예요. 🅱️ 는 설정된 로컬 통합 브랜치의 이름과 로컬 ref `refs/heads/<통합>` 의 tip 이고, 세션의 `HEAD`·원격 ref·태그가 아니에요. 🌳 는 작업 폴더 이름이 아니라 체크아웃된 브랜치 이름이고, detached `HEAD`(리더)에서는 항목을 지어내지 않고 빼요. 이 줄에서 `🧭`, ahead/behind(`↑ ↓`), 미커밋(`+N`) 표시가 빠져요. 이 개정은 `SPEC-20261122-statusline-where` 의 줄 1 항목 중 `🌳 <작업 폴더 이름>`, `🅱️ <브랜치>` 와 그 표시들, `🧭` 거리를, 그리고 방금 착지한 `SPEC-20261123-statusline-integration-distance` 가 화면에 내던 `🧭` 를 **대체**해요(그 SPEC 의 계산 코드는 이 SPEC 이 지우지 않아요, Non-goals).

## Non-goals

- 첫 줄의 `🧠` 모델·`📡` 저장소·`🔗` 요청 항목의 모양과 값(순서만 위 지시대로 놓여요), 카드 단계·흐름의 계산.
- 첫 줄에서 빠진 ahead/behind/미커밋 숫자와 `🧭` 거리를 계산하던 git 호출·캐시·`Where` 거리 코드를 지우는 일. 화면에서만 빼고 코드는 그대로 둬요. 쓰이지 않는 일을 없애는 것은 리더가 "이 카드 착지 뒤 다시 본다"고 한 다음 카드의 몫이에요.
- 흐름 줄의 항목과 아이콘.
- 막대의 그라데이션 색. 이전 막대는 채운 칸마다 초록→빨강 색을 칠했는데, 지금 막대는 **남은** 양을 그리므로 가득 찬 막대가 빨강으로 끝나면 뜻이 거꾸로예요(D-3). 색은 이 SPEC 이 새로 정한 두 가지뿐이에요.
- 제품 문서의 상태줄 설명 전체를 다시 쓰는 일. 이 변경이 거짓으로 만드는 문장만 고쳐요.

## Requirements

- REQ-1: The status line shall write its second line as the Claude Code version, the MoAI version, the card flow, the blocked and hold counts and the held card in this order · failure: none — an item without its input is left out with its separator, the line holds the versions alone when there is no flow to show and is left out only when nothing is left
- REQ-2: While the width budget is smaller than the second line, the status line shall drop the Claude Code version first, then the MoAI version with its update arrow, then the leftmost item of the flow, and keep the held card and the blocked and hold counts in every form · failure: none — the narrowest form is written when none fits
- REQ-3: The status line shall write each known limit window as `<battery> <label> <bar> <n>%` where `<n>` is the share left, the bar has ten cells, the first `<n>` times ten divided by one hundred (rounded down) cells are `█` and the others `░` · failure: none — the percentage is clamped to 0 to 100 so the bar never has fewer than zero or more than ten filled cells
- REQ-4 (REQ-3): IF the percentage of a window is unknown, THEN the status line shall write `<label> n/a` with no battery icon and no bar · failure: none — the screen never shows an empty placeholder bar or a guessed number
- REQ-5: While the width budget is smaller than the third line, the status line shall drop the reset countdowns first, then the bars, then the five-hour window and the cost, and keep `CW` and `7D` in every form · failure: none — the narrowest form is written when none fits
- REQ-6: The status line shall paint the filled cells of a bar in the base color of the line and the empty cells in the muted color so that the colored line minus its escapes equals the plain line · failure: none — without color the bar is the same text
- REQ-7: The status line shall keep the numbers as the share left, the battery icon boundary, the reset countdown forms, the `context`, `usage_5h`, `usage_7d`, `claude_version`, `moai_version` and `cards` options and one space after every icon as they were · failure: none
- REQ-8: The status line shall write its first line as the model with its effort, the repository with its forge counts, the local integration branch with the tip of its ref, the checked-out branch and the pull request in this order · failure: none — an item without its input is left out with its separator and no item is invented
- REQ-9: The status line shall write the integration branch item as `🅱️ <name>(🏷️ <seven characters>)` with the configured integration branch name and the first seven characters of the commit that the local ref `refs/heads/<name>` names, read from that ref and never from `HEAD` · failure: `🅱️ <name>(🏷️ n/a)` when the local ref cannot be read, never a guessed commit — a head at another commit, a lane branch or a detached `HEAD` never changes the seven characters
- REQ-10: The status line shall write the checked-out item as `🌳 <branch name>` with the branch the `HEAD` file names and not the name of the working folder · failure: none — a detached or unreadable `HEAD` leaves the item out and invents no name
- REQ-11: The status line shall write no distance item, no ahead or behind mark and no uncommitted mark on the first line · failure: none — the repository counts and the pull request item keep their text
- REQ-12: While the width budget is smaller than the first line, the status line shall drop the forge counts first, then the repository owner and the model version and the thinking mark and the request state, then the effort and the repository, then the commit of the integration item and the pull request, and keep the model and the integration branch name and the checked-out branch in every form · failure: none — the narrowest form is written when none fits

## Acceptance

- AC-1 (REQ-1): 흐름 줄이 `🤖 2.1.330 | 🗿 4.2648.20 | 🗂️ 176 → 🟡 6 → 📨 13 → 🛬 8 → ✅ 6 | 🚧 7 🙋 9 | 🎯 C-12 🟡` 모양이고, 흐름이 없으면 두 버전만, 버전이 없으면 흐름만 나오고, `claude_version`·`moai_version`·`cards` 옵션이 각자의 항목을 숨겨요 · verify: `cargo test --offline --locked --lib statusline::render::flow_line_tests::the_versions_open_the_flow_line_and_my_card_closes_it`
- AC-2 (REQ-2): 폭이 한 칸 모자랄 때마다 `🤖`, `🗿`(업데이트 화살표와 함께), `🗂️`, `🟡`, `📨`, `🛬`, `✅` 순서로 빠지고 내 카드와 `🚧 🙋` 는 끝까지 남아요. 모든 모양이 아래 Layouts 표의 문자열 그대로예요 · verify: `cargo test --offline --locked --lib statusline::render::flow_line_tests::every_form_is_exactly_its_string_and_changes_at_its_boundary`
- AC-3 (REQ-3): 막대 칸 수가 남은 비율 0·9·10·11·19·20·99·100 에서 0·0·1·1·1·2·9·10 이고(9→10 경계에서 한 칸이 생기고 99→100 에서 열 칸이 차요), 100 을 넘거나 0 보다 작은 입력도 0..10 칸이고, `🔋` 와 `🪫` 의 70/71 경계가 그대로예요 · verify: `cargo test --offline --locked --lib statusline::render::window_tests::the_bar_has_ten_cells_and_changes_at_each_tenth`
- AC-4 (REQ-3): 사용자의 예 — 27%·51%·13% 사용 — 가 `🔋 CW ███████░░░ 73% | 🔋 5H ████░░░░░░ 49% (43m) | 🔋 7D ████████░░ 87% (3d 22h)` 로 나와요 · verify: `cargo test --offline --locked --lib statusline::render::window_tests::the_row_shows_what_is_left_of_each_window`
- AC-5 (REQ-4): 알 수 없는 `CW`·`7D` 는 `CW n/a`·`7D n/a`, 아이콘과 막대 글자(`█`·`░`) 가 없고, 퍼센트가 없는 주간 창도 같아요 · verify: `cargo test --offline --locked --lib statusline::render::window_tests::unknown_windows_and_the_hide_list_stay_as_they_were`
- AC-6 (REQ-5): 한도 줄의 모든 모양이 Layouts 표의 문자열 그대로이고 경계 폭 N 에서 그 모양, N-1 에서 다음 모양이에요 · verify: `cargo test --offline --locked --test statusline_cards line_three_ladder_`
- AC-7 (REQ-6): 색 있는 줄에서 이스케이프를 뺀 글이 색 없는 줄과 같고, 채운 칸과 빈 칸이 서로 다른 색 구간이에요 · verify: `cargo test --offline --locked --lib statusline::render::window_tests::the_colored_bar_has_the_plain_text`
- AC-8 (REQ-1, REQ-3, REQ-5, REQ-8, REQ-11): 실제 바이너리의 출력이 골든 사례 전부(`COLUMNS` 넓음·좁음 포함)에서 고정된 문자열이에요 · verify: `cargo test --offline --locked --test statusline_golden`
- AC-9 (REQ-7): 숫자만 보여 주던 옛 단언(`statusline_numbers_only` 의 "막대 낱말 없음" 세 테스트)은 이 SPEC 의 막대 단언으로 바뀌고, 제품 문서가 새 줄 순서와 막대를 말해요 · verify: `cargo test --offline --locked --test statusline_numbers_only`

- AC-10 (REQ-8, REQ-9, REQ-10): 레인 가지에서 첫 줄이 `🧠 … | 📡 …, ⬇️ … 🔀 … | 🅱️ <통합>(🏷️ <tip 7자리>) | 🌳 <레인 가지>` 이고 `🧭`·`↑ ↓`·`+N` 이 없어요. 실제 git 저장소와 실제 바이너리로 확인하고, 7자리는 `git rev-parse --short=7 refs/heads/<통합>` 의 값이에요 · verify: `cargo test --offline --locked --test statusline_first_line the_lane_branch_shows_the_integration_tip_and_the_checked_out_branch`
- AC-11 (REQ-8, REQ-10): detached `HEAD`(리더)에서 첫 줄이 `🅱️ <통합>(🏷️ <tip 7자리>)` 로 끝나고 `🌳` 가 없고 `HEAD` 의 짧은 커밋이 나오지 않아요 · verify: `cargo test --offline --locked --test statusline_first_line a_detached_leader_shows_the_integration_item_and_no_checked_out_item`
- AC-12 (REQ-9): `HEAD` 가 통합 tip 이 아닌 다른 커밋에 있을 때(레인 가지가 앞서 있고 detached 가 뒤처진 두 경우 모두)도 7자리가 로컬 통합 ref 의 커밋이고, 느슨한 ref 파일이 없고 `packed-refs` 에만 있어도 같아요. tip 을 `HEAD` 에서 읽게 바꾼 뮤턴트가 이 테스트를 실패시켜요 · verify: `cargo test --offline --locked --test statusline_first_line the_tip_is_the_commit_of_the_integration_ref_and_never_the_one_of_head`
- AC-13 (REQ-9): 통합 브랜치 이름은 있는데 로컬 ref 가 없으면 `🅱️ <이름>(🏷️ n/a)` 이고 추측한 커밋이 없어요 · verify: `cargo test --offline --locked --test statusline_first_line an_integration_branch_without_a_local_ref_says_n_a`
- AC-14 (REQ-11): 앞선 커밋과 미커밋 변경이 있는 작업 트리에서도 첫 줄에 `🧭`·`↑`·`↓`·`+N` 이 없고, 저장소 숫자와 `🔗` 요청은 그대로예요 · verify: `cargo test --offline --locked --test statusline_first_line the_first_line_has_no_distance_and_no_ahead_behind_or_uncommitted_mark`
- AC-15 (REQ-12): 첫 줄의 모든 모양이 아래 Layouts 표의 문자열 그대로이고 경계 폭 N 에서 그 모양, N-1 에서 다음 모양이에요 · verify: `cargo test --offline --locked --lib statusline::render::line_one_form_tests`

### Layouts

첫 줄(레인 가지, 요청 있음; 값은 고정 입력):

| 모양 | 글 |
|---|---|
| L0 가장 넓음 | `🧠 Sonnet 5.5 high·t \| 📡 modu-ai/moai-adk, ⬇️ 1 🔀 0 \| 🅱️ v2026.11.0(🏷️ b3ba2b3) \| 🌳 spec/ready-mutant-line-20261123 \| 🔗 #1234 pending` |
| L1 숫자 없음 | `🧠 Sonnet 5.5 high·t \| 📡 modu-ai/moai-adk \| 🅱️ v2026.11.0(🏷️ b3ba2b3) \| 🌳 spec/ready-mutant-line-20261123 \| 🔗 #1234 pending` |
| L2 소유자·모델 버전·생각·요청 상태 없음 | `🧠 Sonnet high \| 📡 moai-adk \| 🅱️ v2026.11.0(🏷️ b3ba2b3) \| 🌳 spec/ready-mutant-line-20261123 \| 🔗 #1234` |
| L3 노력·저장소 없음 | `🧠 Sonnet \| 🅱️ v2026.11.0(🏷️ b3ba2b3) \| 🌳 spec/ready-mutant-line-20261123 \| 🔗 #1234` |
| L4 가장 좁음 | `🧠 Sonnet \| 🅱️ v2026.11.0 \| 🌳 spec/ready-mutant-line-20261123` |
| detached 리더 | `🧠 Sonnet 5.5 high·t \| 📡 modu-ai/moai-adk, ⬇️ 1 🔀 0 \| 🅱️ v2026.11.0(🏷️ b3ba2b3)` |
| 통합 ref 없음 | `🧠 Sonnet 5.5 high·t \| 📡 modu-ai/moai-adk, ⬇️ 1 🔀 0 \| 🅱️ v2026.11.0(🏷️ n/a) \| 🌳 spec/ready-mutant-line-20261123` |
| 통합 브랜치를 모름(프로필도 원격 기본 브랜치도 없음) | `🧠 Sonnet 5.5 high·t \| 📡 modu-ai/moai-adk, ⬇️ 1 🔀 0 \| 🌳 main` |

L4 의 `🅱️ v2026.11.0` 처럼 좁은 모양은 이름만 남기고 `(🏷️ …)` 가 빠져요.

아래는 둘째·셋째 줄이에요. 세 줄 모두 빈 값은 항목이 구분자와 함께 빠져요. 폭은 `unicode-width` 로 세요(테스트가 `display_width` 로 경계를 정해요). 값은 고정 입력이에요.

흐름 줄(보유 카드 `C-12`, 단계 `🟡`, 사용자의 예):

| 모양 | 글 |
|---|---|
| F0 가장 넓음 | `🤖 2.1.330 \| 🗿 4.2648.20 \| 🗂️ 176 → 🟡 6 → 📨 13 → 🛬 8 → ✅ 6 \| 🚧 7 🙋 9 \| 🎯 C-12 🟡` |
| F1 호스트 버전 없음 | `🗿 4.2648.20 \| 🗂️ 176 → 🟡 6 → 📨 13 → 🛬 8 → ✅ 6 \| 🚧 7 🙋 9 \| 🎯 C-12 🟡` |
| F2 버전 없음 | `🗂️ 176 → 🟡 6 → 📨 13 → 🛬 8 → ✅ 6 \| 🚧 7 🙋 9 \| 🎯 C-12 🟡` |
| F3 | `🟡 6 → 📨 13 → 🛬 8 → ✅ 6 \| 🚧 7 🙋 9 \| 🎯 C-12 🟡` |
| F4 | `📨 13 → 🛬 8 → ✅ 6 \| 🚧 7 🙋 9 \| 🎯 C-12 🟡` |
| F5 | `🛬 8 → ✅ 6 \| 🚧 7 🙋 9 \| 🎯 C-12 🟡` |
| F6 | `✅ 6 \| 🚧 7 🙋 9 \| 🎯 C-12 🟡` |
| F7 가장 좁음 | `🚧 7 🙋 9 \| 🎯 C-12 🟡` |
| 흐름이 없음 | `🤖 2.1.330 \| 🗿 4.2648.20` (좁으면 `🗿 4.2648.20`, 더 좁아도 마지막 모양) |
| 노트를 못 읽음(나누지 못한 흐름) | `🤖 2.1.330 \| 🗿 4.2648.20 \| 🗂️ 150 🟡 15` (모양은 버전만 줄어요) |

한도 줄(사용자의 예, 사용 27%·51%·13%):

| 모양 | 글 |
|---|---|
| Z0 가장 넓음 | `🔋 CW ███████░░░ 73% \| 🔋 5H ████░░░░░░ 49% (43m) \| 🔋 7D ████████░░ 87% (3d 22h)` |
| Z1 카운트다운 없음 | `🔋 CW ███████░░░ 73% \| 🔋 5H ████░░░░░░ 49% \| 🔋 7D ████████░░ 87%` |
| Z2 막대 없음 | `🔋 CW 73% \| 🔋 5H 49% \| 🔋 7D 87%` |
| Z3 가장 좁음 | `🔋 CW 73% \| 🔋 7D 87%` |
| 한도 정보가 없는 사용자(CW 사용 41%) | Z0·Z1 `🔋 CW █████░░░░░ 59% \| 💰 3.87`, Z2 `🔋 CW 59% \| 💰 3.87`, Z3 `🔋 CW 59%` |
| 알 수 없는 창 | `CW n/a \| 7D n/a` (모든 모양에서 막대 없음) |

## Scope

- write: `.moai/specs/SPEC-20261123-statusline-order-bars/**`, `src/statusline/**`, `tests/statusline*.rs`, `tests/fixtures/statusline/**`, `tests/fixtures/statusline-claude/**`, `docs/rust-product-path.md`
- read-only: `tools/statusline_golden.py` (골든 파일을 `--write` 로 다시 쓰고 `--check` 로 확인하는 데만 실행해요), 이 SPEC 이 말하는 이전 SPEC 들
- forbidden: `.claude/worktrees/**`(다른 레인의 작업 폴더), `.moai-cli/**`

## Plan

1. 테스트를 먼저 고쳐 red 를 봐요: 흐름 줄 순서(`flow_line_tests`), 막대 칸 수와 경계(`window_tests`), 알 수 없는 창, 한도 줄 사다리(`statusline_cards`), `statusline_numbers_only` 의 옛 단언.
2. `render.rs`: `Piece` 에 "앞 조각에 붙임" 표지를 더해 막대의 채운 칸과 빈 칸을 공백 없이 다른 색으로 이어요. `flow_forms` 는 두 버전을 앞에 두고 내 카드를 뒤로 보내며 버전부터 빠지는 사다리를 가져요. `line_three_forms` 는 버전을 빼고 막대 단계를 카운트다운과 5H 사이에 둬요.
3. 골든 파일을 도구로 다시 쓰고(`tools/statusline_golden.py --write`) 바뀐 plain 줄을 하나씩 읽어요. 검증은 Rust 테스트 `statusline_golden` 이 해요.
4. 제품 문서의 줄 순서와 막대 문장을 고치고 `statusline_numbers_only` 의 문서 테스트를 새 문장에 맞춰요.
5. 실제 바이너리 출력(넓은 폭·좁은 폭)을 Result 에 붙여요.
6. (개정 2) 새 파일 `tests/statusline_first_line.rs` 에 AC-10 ~ AC-14 의 테스트를 먼저 써서 red 를 봐요(임시 git 저장소: 통합 브랜치와 레인 가지가 서로 다른 커밋을 가리키게 만들고, `HEAD` 가 통합 tip 이 아닌 경우와 `packed-refs` 만 있는 경우를 포함). `render.rs` 의 `line_one_form_tests` 는 Layouts 표의 L0~L4 로 고쳐요.
7. (개정 2) `state.rs`: `read_git` 가 `ref_commit` 이 읽은 통합 ref 의 커밋을 `HEAD` 가 읽히든 말든 `GitRead` 에 담아 `Status` 로 나르고, `render.rs` 의 `line_one_form` 은 `where_items` 대신 `Status` 의 브랜치 이름·통합 이름·tip 으로 새 항목을 만들어요. `ICONS` 표에 `🏷️` 를 더해요. 뮤턴트(tip 을 `HEAD` 에서 읽기)를 한 번 넣어 AC-12 의 테스트가 실패하는 것을 보고 되돌려요.
8. (개정 2) 골든 파일을 다시 쓰고 줄 1 이 바뀐 것을 하나씩 읽어요. 줄 1 변화는 이 개정의 의도(항목 순서, `🅱️` 통합 항목, `🌳` 브랜치, `🧭`·`↑ ↓`·`+N` 없음)뿐이어야 해요.

## Risks

1. **`█`·`░` 의 폭.** 이 문자들은 동아시아 모호 폭(East Asian Ambiguous)이에요. `unicode-width` 는 1칸으로 세는데, 한국어 로케일의 일부 터미널은 2칸으로 그려요. 그러면 막대 줄이 센 것보다 넓어져 호스트가 줄을 자르거나 감쌀 수 있어요. 지워지기 전에도 같은 글자였고 사용자가 다시 요청했으니 그대로 쓰되, 문제가 보이면 폭이 좁은 모양(막대 없음)으로 내려가는 사다리가 안전판이에요.
2. **막대의 방향.** 숫자가 남은 비율이라서 막대도 남은 비율을 그려요(가득 = 넉넉). 지워지기 전 막대는 쓴 비율이었어요. 숫자와 막대가 같은 양을 가리키게 한 것이 이 SPEC 의 선택이고 D-1 에서 되돌릴 수 있어요.
3. **골든 파일 다수가 바뀌어요.** 두 버전의 이동과 막대로 plain 줄이 거의 모든 사례에서 바뀌어요. 바뀐 줄을 하나씩 읽고, 의도하지 않은 변화가 없는지(첫 줄과 카운트다운 숫자) 확인해요.
4. (개정 2) **tip 이 오래된 것.** `refs/heads/<통합>` 는 이 체크아웃의 공통 git 폴더에서 읽어요. 다른 작업 폴더가 통합 ref 를 올려도 `packed-refs` 와 느슨한 ref 파일을 매 렌더 읽으므로 같아요(캐시 키에도 통합 커밋이 있어요). 이 세션에서 `git fetch` 를 하지 않아 원격이 앞서 있어도 로컬 ref 의 tip 을 보여 줘요(지시가 로컬 ref).
5. (개정 2) **첫 줄의 `🌳` 뜻이 바뀌어요.** 지금까지 `🌳` 는 호스트의 작업 폴더 이름이었어요. 이 줄을 읽던 테스트·문서·`worktree` 옵션 설명이 같이 바뀌어야 하고, 작업 폴더 이름은 첫 줄에서 사라져요(흐름 줄의 내 카드 판별은 그 이름을 그대로 쓰므로 변하지 않아요).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 막대는 숫자와 같은 양, 곧 남은 비율을 그려요(채운 칸 = 남은 비율 × 10 / 100 내림) | 사용자: "숫자와 되돌림 시간은 그대로". 숫자가 남은 비율이라 쓴 비율 막대는 같은 줄에서 서로 어긋나요. 리더 결정 범위 안의 되돌릴 수 있는 선택(레인이 정하고 READY 에 적어요) | 채운 칸 계산에 `100 - left` 를 쓰는 한 줄 |
| D-2 | 내 카드는 단계 아이콘을 그대로 달고 `🎯 C-<n> <단계>` 로 맨 뒤에 가요 | 사용자의 예 `🎯 C-287` 에는 단계 아이콘이 없지만 카드는 순서만 바꾸라고 했고 단계 아이콘을 빼라는 말은 없었어요 | `MyCard` 조각에서 `stage.icon()` 을 빼는 한 줄 |
| D-3 | 막대 색은 채운 칸 기본 색, 빈 칸 흐린 색 두 가지뿐이고 그라데이션은 되살리지 않아요 | 남은 양을 그리는 막대가 초록에서 빨강으로 끝나면 뜻이 거꾸로예요. 색만이 신호가 되지 않게 `█`/`░` 모양이 이미 구분해요 | 테마에 `stops` 를 되돌리고 칸마다 색을 칠하기 |
| D-4 | 한도 줄의 사다리는 카운트다운 → 막대 → `5H`·`💰` 순서로 빠져요 | 지금 규칙(카운트다운이 먼저, 창은 나중)을 지키고 막대는 숫자를 꾸미는 장식이라 숫자가 있는 창보다 먼저 빠져요. 사용자 말: "좁은 폭에서 줄이는 순서는 지금 규칙을 따라요" | 막대 단계를 사다리의 다른 칸으로 옮기기 |
| D-5 | 흐름 줄의 사다리는 호스트 버전 → moai 버전 → 흐름의 왼쪽 항목 순서로 빠져요 | 두 버전은 지금도 사다리에서 가장 먼저 빠지는 항목이에요(`limits` SPEC 의 Z1·Z2) | 버전 조각을 사다리 끝으로 옮기기 |
| D-6 | 흐름이 없어도 버전은 둘째 줄에 남아요 | 버전은 원장과 무관한 항목이고 지금도 원장 없이 나와요 | 버전을 흐름이 있을 때만 쓰게 하기 |
| D-7 | `💰` 는 지금 규칙(한도 정보가 없을 때만, 셋째 줄)대로 그 줄에 남아요 | 카드가 한도 줄에서 빼라고 한 것은 두 버전뿐이에요 | `💰` 를 둘째 줄로 옮기기 |
| D-8 | `🔗` 요청 항목은 첫 줄 끝(`🌳` 뒤)에 남아요 | 지시가 빼라고 한 것은 `🧭`·ahead/behind·미커밋이고 요청은 말이 없어요. 요청은 체크아웃된 브랜치의 것이라 `🌳` 뒤가 자연스러워요 | `🔗` 를 줄에서 빼기 |
| D-9 | 통합 브랜치 이름은 지금 규칙 그대로(프로필의 `work_flow.integration_branch`, 없으면 원격 기본 브랜치)이고, 둘 다 없으면 `🅱️` 항목이 없고 `🌳` 만 남아요 | 지시의 "설정된 로컬 통합 브랜치". 일반 프로젝트에서는 `🌳 <브랜치>` 한 항목이 지금의 `🅱️ <브랜치>` 자리를 이어요 | 이름이 없을 때 `🅱️ <브랜치>` 로 되돌리기 |
| D-10 | 통합 ref 를 못 읽으면 `(🏷️ n/a)`, 너무 좁으면 `(🏷️ …)` 가 먼저 빠져요 | 프로젝트 규칙: 읽지 못하는 값은 `n/a` 이고 추측한 값은 없어요. 좁은 폭에서 `🏷️` 는 이름보다 덜 중요해요 | `(🏷️ …)` 를 사다리의 다른 칸으로 옮기기 |
| D-11 | 체크아웃된 브랜치가 통합 브랜치와 같은 이름이어도 두 항목을 다 써요 | 지시의 순서와 표기를 그대로 따라요. 이름이 같은 경우는 통합 브랜치가 보호돼 드물어요 | 같은 이름이면 `🌳` 를 빼기 |
| D-12 | 첫 줄에서 빠진 숫자를 계산하던 git 호출·캐시·`Where` 거리 코드는 지우지 않아요 | 리더가 지운 항목들을 다음 카드에서 다시 본다고 했어요 | 지우는 카드 |

## Result

구현했어요(2026-11-23, 브랜치 `spec/20261123-statusline-order-bars`, 개정 2 까지 구현한 코드 커밋 head `3d929c4a07d20359806e7cb6a6fdcf961d991bef`). 개정 1(둘째·셋째 줄)은 red 테스트 `b1f6fd1ae5` 를 먼저 커밋한 뒤 코드 `cf4ba47cd5` 로, 개정 2(첫 줄)는 SPEC 커밋 `1a5d439858` 위에 red 테스트 `352d54a5e1`·`d4f7303c9c` 를 먼저 커밋한 뒤 코드 `3d929c4a07` 로 구현했어요. 독립 감사(읽기 전용 `moai-audit`, tier 1)는 차단 결함 없이 끝났고 비차단 메모의 낡은 주석 2곳과 문서 1문장은 이어지는 커밋에서 고쳤어요(동작 변화 없음). 증거는 카드의 READY 노트에 있어요.

- red(개정 1, `b1f6fd1ae5`, 옛 코드): `cargo test --offline --locked --lib statusline::` exit 101, 181 passed 34 failed. `--test statusline_cards` exit 101, 14 passed 19 failed. `--test statusline_flow` 2 passed 3 failed. `--test statusline_numbers_only` 2 passed 3 failed. `--test statusline_worktree_label` 3 passed 4 failed. `--test statusline_counts_refresh` 8 passed 5 failed. 모두 exit 101.
- red(개정 2, `352d54a5e1`·`d4f7303c9c`, 옛 코드): `--lib statusline::` exit 101, 194 passed 23 failed. `--test statusline_first_line` exit 101, 0 passed 5 failed. `statusline_cards` 23 passed 10 failed. `statusline_git_env` 18 passed 9 failed. `statusline_integration_distance` 0 passed 7 failed. `statusline_worktree_label` 1 passed 6 failed. `statusline_flow` 4 passed 1 failed. `statusline_numbers_only` 4 passed 1 failed. 모두 exit 101이고, 실패는 동작 때문이에요(예: detached 사례가 `🅱️ 74aa922 +1` 을 써서 `🅱️ v2026.11.0(🏷️ 2894f90)` 와 달랐어요).
- green(head `3d929c4a07` 에서 레인이 `target/lane-notes/verify.sh` 로 다시 돌림, 로그 `target/lane-notes/verify.log`): lib 217 passed, `statusline_first_line` 5, `statusline_cards` 33 passed 1 ignored, `statusline_counts_refresh` 13, `statusline_flow` 5, `statusline_git_env` 27, `statusline_golden` 1, `statusline_guard_false_positive` 3, `statusline_home_state` 12, `statusline_integration_distance` 7, `statusline_numbers_only` 5, `statusline_tag_order` 3, `statusline_worktree_label` 7, `status_line_cli` 16 이 모두 exit 0, failed 0. `cargo clippy --offline --locked --all-targets -- -D warnings` exit 0, `cargo fmt --check` exit 0, `git diff --check` exit 0, `moai spec check` exit 0.
- 뮤턴트(레인이 직접 실행, `target/lane-notes/mutant.sh`): `read_git` 의 통합 tip 을 `refs/heads/<통합>` 대신 `HEAD` 가 가리키는 커밋에서 읽게 바꾸고 `cargo test --offline --locked --test statusline_first_line` 을 돌리니 exit 101, 0 passed 5 failed(다섯 테스트 모두 실패: `the_lane_branch_shows_the_integration_tip_and_the_checked_out_branch`, `a_detached_leader_shows_the_integration_item_and_no_checked_out_item`, `the_tip_is_the_commit_of_the_integration_ref_and_never_the_one_of_head`, `an_integration_branch_without_a_local_ref_says_n_a`, `the_first_line_has_no_distance_and_no_ahead_behind_or_uncommitted_mark`). `git checkout -- src/statusline/state.rs` 로 되돌린 뒤 `git diff -- src` 가 0줄이에요. 뮤턴트 뒤에는 코드를 되돌린 head `3d929c4a07` 에서 바이너리를 다시 빌드해 아래 출력을 만들었어요.
- 결정 D-1(막대는 남은 비율), D-2(내 카드는 단계 아이콘을 그대로 달아요), D-3(그라데이션 없이 두 색)은 위 Decisions 표에 이유와 되돌리는 방법이 있어요.

- AC-1: met — `statusline::render::flow_line_tests::the_versions_open_the_flow_line_and_my_card_closes_it` 가 가장 넓은 흐름 줄 `🤖 2.1.330 | 🗿 4.2648.20 | 🗂️ 176 → 🟡 6 → 📨 13 → 🛬 8 → ✅ 6 | 🚧 7 🙋 9 | 🎯 C-12 🟡`, `claude_version`·`moai_version`·`cards` 옵션이 각자의 항목과 구분자만 숨기는 것, 흐름이 없을 때 두 버전만(`🤖 2.1.330 | 🗿 4.2648.20`), 버전이 없을 때 흐름만 나오는 것을 고정해요.
- AC-2: met — `statusline::render::flow_line_tests::every_form_is_exactly_its_string_and_changes_at_its_boundary` 가 Layouts 표의 F0~F7 문자열 그대로와 각 모양의 경계 폭 N 에서 그 모양, N-1 에서 다음 모양(`🤖`, `🗿`, `🗂️`, `🟡`, `📨`, `🛬`, `✅` 순서로 빠지고 내 카드와 `🚧 🙋` 는 남음)을 고정해요.
- AC-3: met — `statusline::render::window_tests::the_bar_has_ten_cells_and_changes_at_each_tenth` 가 남은 비율 0·9·10·11·19·20·99·100 의 막대 `░░░░░░░░░░`·`░░░░░░░░░░`·`█░░░░░░░░░`·`█░░░░░░░░░`·`█░░░░░░░░░`·`██░░░░░░░░`·`█████████░`·`██████████`, 100 을 넘거나 0 보다 작은 입력의 0~10 칸 한정, 0~100 모든 값에서 열 칸 중 `left / 10` 칸이 찬 것, `🔋`/`🪫` 가 29 남음 이하에서 바뀌는 경계를 고정해요.
- AC-4: met — `statusline::render::window_tests::the_row_shows_what_is_left_of_each_window` 가 `🔋 CW ███████░░░ 73% | 🔋 5H ████░░░░░░ 49% (43m) | 🔋 7D ████████░░ 87% (3d 22h)` 와 5시간 창이 없는 줄, 콜론이 없는 것을 고정해요.
- AC-5: met — `statusline::render::window_tests::unknown_windows_and_the_hide_list_stay_as_they_were` 가 `CW n/a | 7D n/a` 에 아이콘과 `█`·`░` 가 없는 것과 `context`·`usage_5h`·`usage_7d` 숨김을, `tests/statusline_cards.rs` 의 `line_three_ladder_unknown_windows_never_draw_a_bar` 가 바이너리의 모든 폭에서 같은 것을 고정해요.
- AC-6: met — `tests/statusline_cards.rs` 의 `line_three_ladder_every_form_is_the_string_of_the_layouts_section` 이 한도 줄의 모든 모양의 문자열과 폭, 폭 N+2 에서 그 모양, N+1 에서 다음 모양을, `line_three_ladder_without_a_usable_columns_and_below_the_narrowest` 가 `COLUMNS` 없음과 가장 좁은 폭 아래의 모양을 고정해요.
- AC-7: met — `statusline::render::window_tests::the_colored_bar_has_the_plain_text` 가 두 테마(`catppuccin-mocha`, `catppuccin-latte`)에서 색 있는 줄에서 이스케이프를 뺀 글이 색 없는 줄과 같은 것과 채운 칸·빈 칸이 서로 다른 색 구간인 것을 고정해요.
- AC-8: met — `--test statusline_golden` 이 골든 파일 전부의 실제 바이너리 출력을 고정해요. `tools/statusline_golden.py --write` 로 골든 파일 108개를 다시 썼고 바뀐 plain 파일 36개를 옛 것과 하나씩 비교했어요(레인이 `git show` 로 옛 파일을 읽어 비교). 한도 숫자·카운트다운은 바뀌지 않았고, 두 버전이 둘째 줄로 옮겨졌고 셋째 줄에 막대가 생겼어요(개정 1). 개정 2 는 골든 파일 56개를 다시 썼고, 그중 plain 17개의 줄 1 이 바뀌었는데 `🧠 → 📡 → 🅱️(🏷️) → 🌳 → 🔗` 순서, `🅱️ <통합>(🏷️ <7자리>)` 모양, `🧭`·`↑ ↓`·`+N` 없음만 확인했어요(`target/lane-notes/golden_line1.py` 로 옛 `1a5d439858` 과 새 head 의 줄 1 을 모두 비교, 의도하지 않은 변화 없음, 줄 2·3 은 그대로). 하네스 사례들은 통합 브랜치를 모르므로 `🌳 main` 만 남는 줄이 많아요(D-9). 줄 1 이 없는 세 사례(`a06`, `a09`, `a15`)는 옛 첫 줄이던 한도 줄 자리에 버전 줄이 올라와 맨 윗줄이 `🗿 …` 로 달라졌고, 막대의 채운 칸 수는 모든 골든 줄에서 숫자의 `left × 10 / 100` 내림과 같아요. 버전 글자는 하네스가 읽은 빌드 버전(`4.2647.0` → `4.2648.0`)이라 다시 쓰면 함께 바뀌어요.
- AC-9: met — `tests/statusline_numbers_only.rs` 의 옛 "막대 낱말 없음" 세 테스트를 `the_renderer_draws_the_bar_and_carries_no_color_gradient_and_no_card_numbers`, `the_statusline_doc_describes_the_bars_and_the_line_order`, `the_statusline_tests_pin_the_bar` 로 바꿨고, `docs/rust-product-path.md` 가 새 줄 순서와 막대를 말해요(`--test statusline_numbers_only` 5 passed).
- AC-10: met — `tests/statusline_first_line.rs` 의 `the_lane_branch_shows_the_integration_tip_and_the_checked_out_branch` 가 실제 임시 git 저장소(통합 `v2026.11.0` 에서 딴 레인 가지가 두 커밋 앞서고 `HEAD` 는 그 가지)와 실제 바이너리로 첫 줄이 `🧠 Sonnet 5.5 high·t | 📡 modu-ai/moai-adk, ⬇️ 1 🔀 0 | 🅱️ v2026.11.0(🏷️ <tip>) | 🌳 spec/ready-mutant-line-20261123` 인 것을 고정해요. `<tip>` 은 테스트가 `git rev-parse --short=7 refs/heads/<통합>` 으로 읽고, `HEAD` 의 7자리와 다르다는 것도 확인해요.
- AC-11: met — `a_detached_leader_shows_the_integration_item_and_no_checked_out_item` 이 detached `HEAD`(통합 tip 과 다른 커밋)에서 첫 줄이 `🅱️ v2026.11.0(🏷️ <tip>)` 로 끝나고 `🌳` 도 `HEAD` 의 짧은 커밋도 없는 것을 고정해요.
- AC-12: met — `the_tip_is_the_commit_of_the_integration_ref_and_never_the_one_of_head` 가 레인 가지(앞섬)와 detached(뒤처짐) 두 경우, 그리고 느슨한 ref 파일을 `git pack-refs` 로 없앤 `packed-refs` 만 있는 경우에도 7자리가 통합 ref 의 커밋인 것을 고정해요. tip 을 `HEAD` 에서 읽는 뮤턴트에서 이 테스트가 실패했어요(위 뮤턴트 항목).
- AC-13: met — `an_integration_branch_without_a_local_ref_says_n_a` 가 로컬 ref 가 없는 통합 브랜치에서 `🅱️ <이름>(🏷️ n/a)` 를 고정해요.
- AC-14: met — `the_first_line_has_no_distance_and_no_ahead_behind_or_uncommitted_mark` 가 앞선 커밋과 미커밋 변경이 있는 작업 트리에서 첫 줄에 `🧭`·`↑`·`↓`·`+N` 이 없고 저장소 숫자와 `🔗` 가 그대로인 것을 고정해요.
- AC-15: met — `statusline::render::line_one_form_tests::every_form_is_exactly_the_string_of_the_layouts_section` 이 Layouts 표의 모양이 문자열 그대로이고 각 경계 폭에서 바뀌는 것을, `the_detached_missing_ref_and_unknown_integration_rows_are_the_strings_of_the_section` 이 detached·ref 없음·통합 모름 줄을, `the_first_line_has_no_distance_and_no_marks_in_any_form` 이 모든 모양에서 거리와 표시가 없는 것을 고정해요(`--lib statusline::` 217 passed 안에 있어요).

### 실제 바이너리 출력

**첫 줄(개정 2)**, 레인이 `target/lane-notes/first_line_output.sh` 로 만든 임시 `HOME` 과 임시 git 저장소(통합 브랜치 `v2026.11.0`, 프로필이 그 이름을 말함, 이슈 1·PR 0), `target/debug/moai-cli statusline --render-budget-ms 3600000`, `NO_COLOR=1`, 입력의 모델 `Sonnet 5.5 high`, 저장소 `modu-ai/moai-adk`. 각 사례 앞의 `integration tip` 은 `git rev-parse --short=7 refs/heads/v2026.11.0`, `HEAD` 는 `git rev-parse --short=7 HEAD` 이고 둘이 달라요(저장소는 매번 새로 만들어서 값은 실행마다 달라요). 모두 exit 0:

```
### A: lane branch
integration tip (7): 8c55a9b   HEAD (7): c037295   HEAD file: ref: refs/heads/spec/ready-mutant-line-20261123
--- COLUMNS=140
🧠 Sonnet 5.5 high·t | 📡 modu-ai/moai-adk, ⬇️ 1 🔀 0 | 🅱️ v2026.11.0(🏷️ 8c55a9b) | 🌳 spec/ready-mutant-line-20261123
--- COLUMNS=80
🧠 Sonnet | 🅱️ v2026.11.0(🏷️ 8c55a9b) | 🌳 spec/ready-mutant-line-20261123
--- COLUMNS=50
🧠 Sonnet | 🅱️ v2026.11.0 | 🌳 spec/ready-mutant-line-20261123
### B: detached leader
integration tip (7): 8c55a9b   HEAD (7): c037295   HEAD file: c037295e40cadb6dcd8e4ba8d70922e8364d1197
--- COLUMNS=140
🧠 Sonnet 5.5 high·t | 📡 modu-ai/moai-adk, ⬇️ 1 🔀 0 | 🅱️ v2026.11.0(🏷️ 8c55a9b)
--- COLUMNS=80
🧠 Sonnet 5.5 high·t | 📡 modu-ai/moai-adk | 🅱️ v2026.11.0(🏷️ 8c55a9b)
--- COLUMNS=50
🧠 Sonnet | 🅱️ v2026.11.0(🏷️ 8c55a9b)
```

**둘째·셋째 줄(개정 1)을 포함한 전체**, `target/debug/moai-cli statusline --render-budget-ms 3600000 --now-unix 1738429000`, 입력 `tests/fixtures/statusline/stdin/a02-doc-full.json`, 임시 `HOME` 과 git 이 아닌 임시 프로젝트(그래서 첫 줄에 `🅱️`·`🌳` 가 없어요), `NO_COLOR=1`, `COLUMNS` 140·80·60·45·30, 모두 exit 0:

```
--- COLUMNS=140
🧠 Opus high·t | 📡 anthropics/claude-code, ⬇️ n/a 🔀 n/a | 🔗 #1234 pending
🤖 2.1.90 | 🗿 4.2648.0
🔋 CW █████████░ 92% | 🔋 7D █████░░░░░ 59% (4d 23h)
[exit=0]
--- COLUMNS=80
🧠 Opus high·t | 📡 anthropics/claude-code, ⬇️ n/a 🔀 n/a | 🔗 #1234 pending
🤖 2.1.90 | 🗿 4.2648.0
🔋 CW █████████░ 92% | 🔋 7D █████░░░░░ 59% (4d 23h)
[exit=0]
--- COLUMNS=60
🧠 Opus high | 📡 claude-code | 🔗 #1234
🤖 2.1.90 | 🗿 4.2648.0
🔋 CW █████████░ 92% | 🔋 7D █████░░░░░ 59% (4d 23h)
[exit=0]
--- COLUMNS=45
🧠 Opus high | 📡 claude-code | 🔗 #1234
🤖 2.1.90 | 🗿 4.2648.0
🔋 CW █████████░ 92% | 🔋 7D █████░░░░░ 59%
[exit=0]
--- COLUMNS=30
🧠 Opus | 🔗 #1234
🤖 2.1.90 | 🗿 4.2648.0
🔋 CW 92% | 🔋 7D 59%
[exit=0]
```

원장이 있는 예(`tests/fixtures/statusline-claude/golden/b14-card-counts-waiting.plain.txt`, 이 골든의 하네스는 통합 브랜치를 모르므로 `🌳 main` 만 나와요):

```
🧠 Opus high·t | 📡 anthropics/claude-code, ⬇️ 46 🔀 0 | 🌳 main | 🔗 #1234 pending
🤖 2.1.90 | 🗿 4.2648.0 | 🗂️ 25 → 📨 3 → ✅ 12 | 🙋 5
🔋 CW █████████░ 92% | 🔋 7D █████░░░░░ 59% (4d 23h)
```

### 하지 않았거나 확인하지 못한 것

- 파이썬 테스트 모듈은 손으로 돌리지 않았어요(골든 도구는 `--write` 로만 썼고 검증은 Rust 테스트 `statusline_golden` 이 해요).
- 원격 CI 실행은 읽지 않았어요.
- Windows 와 Linux 에서는 돌리지 않았어요.
- 한국어 로케일 터미널에서 `█`/`░` 의 칸 폭은 보지 못했어요(위 Risks 1).
- 개정 2: "detached `HEAD` 가 `🌳` 를 지어내는" 뮤턴트는 제가 돌리지 않았어요(AC-11 의 테스트가 `🌳` 없음과 짧은 커밋 없음을 단언하는 것만 읽었어요). 돌린 뮤턴트는 "tip 을 `HEAD` 에서 읽기" 하나예요.
- 개정 2: 첫 줄에서 빠진 거리·ahead/behind·미커밋 계산(`where_items`, `Where`, `Marks`, git 숫자 캐시)은 지우지 않았어요(D-12). 화면에 안 나오는 코드가 남아 있고, 정리는 다음 카드의 몫이에요. `cargo clippy --all-targets -D warnings` 는 exit 0 이라 죽은 코드 경고는 없었어요.
- 통합 브랜치 tip 이 `0dfdfb31dc` 에서 `b3ba2b30d5` 로 다시 움직였지만 흐름 규칙대로 따라가지 않았어요. 합칠 때의 충돌은 착지하는 쪽이 다뤄요.
- 독립 감사는 첫 head `5e45be7bd4` 에서 받았고, 그 뒤 고친 주석·문서 변경분은 감사 에이전트가 변경분만 다시 읽어요(결과는 READY 노트). 감사 판정 줄은 리더가 기록해요.
