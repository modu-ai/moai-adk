---
id: SPEC-20261124-statusline-bar-color
title: "The status line limit bars paint their filled cells in a level color of the share left"
status: approved
revision: 1
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-23
  basis: "리더 결정 노트 7989 (2026-11-23T23:19:26Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 되돌릴 수 있음)가 개정 1 을 승인하고 D-1 을 옵션 B(초록 70 이상, 노랑 50~69, 주황 30~49, 빨강 30 미만, 빨강은 LOW_LEFT_PERCENT 상수)로 정했어요. 범위는 SCOPE-REQUEST 7988 그대로 넓혀졌어요. 리더가 노트에 적었고 레인은 그 노트를 인용만 해요. 사용자 직접 선택이 아니라 리더 판단이에요"
  digest: "sha256:c5eb45fd8511cde8d715b63a195212f91ba35575c86a9b676e803a3e06efef66"
---

# The status line limit bars paint their filled cells in a level color of the share left

## Intent

사용자 지시예요(2026-11-24, 카드 C-439: "그래프는 이전처럼 색상을 추가하자"). 상태줄 3번째 줄의 `CW`·`5H`·`7D` 막대(`🔋 CW ███████░░░ 73%`)는 지금 채운 칸이 줄의 글자색(`#CDD6F4`)이고 빈 칸만 흐린 색이에요. 예전에는 막대마다 테마의 네 색(초록, 노랑, 주황(peach), 빨강)이 있었어요(`git show b6e8e9c977^:src/statusline/render.rs` 의 `stops`; 그 커밋이 창을 숫자만으로 바꿨고 `cf4ba47cd5` 가 막대를 색 없이 되살렸어요). 이 SPEC 은 채운 칸에 **남은 비율에 따른 단계 색**을 되살려요. 창의 숫자가 이제 쓴 비율이 아니라 남은 비율이라서(`SPEC-20261122-statusline-limits-left`) 넉넉히 남았으면 초록, 얼마 안 남았으면 빨강이에요.

리더 결정 노트 7983(사용자 직접 선택, 4단계 색)이 정한 것: 한 막대는 한 색(채운 칸 전부), 빈 칸은 지금의 흐린 색, 색이 없으면 출력은 글자 그대로, 색은 카드 본문의 두 테마 값이에요. 예전 코드는 칸마다 네 색을 이어 붙인 그라데이션(gradient)이었지만 이번에는 막대 하나가 한 단계 색이에요.

### 읽고 잰 것 (2026-11-24, NeXT `67ba15244f`)

1. 지금 코드: `Window::known`(`src/statusline/render.rs`)이 채운 칸 조각에 `color: ""`(줄의 기본색)을 주고, 빈 칸 조각에 `t.muted` 를 줘요. 색 조각은 `Renderer::paint_form` 이 `Profile::paint` 로 써요. 색이 없는 프로필(`None`)이면 `paint` 가 글자만 돌려줘요(`Renderer::paint`).
2. `LOW_LEFT_PERCENT` 는 **30** 이에요(`render.rs` 25행, `100 - HIGH_USE_PERCENT`). 카드 노트 7983 은 빨강 기준을 이 값에 맞추되 "10 이 아니면 그 값을 쓰고 SPEC 에 적어요"라고 했어요. 그리고 노트가 쓴 단계표 50/25/10 의 `10` 은 이 값이 10 이라고 본 가정이에요.
3. 막대는 열 칸이고 채운 칸 수는 `left * 10 / 100`(내림)이에요. 남은 비율 0~9 는 채운 칸이 **없어요**(`the_bar_has_ten_cells_and_changes_at_each_tenth`). 그래서 "10 미만이 빨강"이면 빨강은 막대에 한 칸도 칠하지 못해요.
4. 색이 있는 기존 테스트: `window_tests` 의 색 테스트(채운 칸은 기본색, 빈 칸은 `muted`, 두 조각 사이에 틈 없음)가 새 규칙과 부딪혀요. 색 골든 파일 `tests/fixtures/statusline-claude/golden/*.ansi.txt`, `*.ansi256.txt` 중 막대(`█`)가 있는 것이 각각 21개예요(`grep -l`); `.plain.txt` 는 바뀌지 않아야 해요.

### 결정이 필요한 것 (D-1, 리더 승인 대기)

단계표의 숫자예요. 노트 7983 의 50/25/10 과 "빨강은 `LOW_LEFT_PERCENT`(30) 에 맞춰요"는 한꺼번에 이루어지지 않아요: 빨강이 30 미만이면 25~50 노랑과 10~25 주황이 빨강 아래로 들어가요.

- 옵션 A: 노트의 50/25/10 을 그대로 써요(초록 50 이상, 노랑 25~49, 주황 10~24, 빨강 10 미만). 막대 빨강은 한 칸도 칠하지 못하고(위 3번) 아이콘 `🪫` 은 30 미만부터라서 25~29 에서 `🪫` 와 노랑이 같이 나와요.
- 옵션 B (기본, 권장): 빨강을 `LOW_LEFT_PERCENT`(30) 미만으로 두어 `🪫` 와 같은 지점에서 바뀌고, 나머지 세 단계를 30~100 에 고르게 나눠요: 초록 70 이상, 노랑 50~69, 주황 30~49, 빨강 30 미만. 빨강이 1~2 칸에 칠해지고(남은 비율 10~29), 색 단계 경계(70, 50, 30)가 모두 0~100 안에서 눈에 보여요. 노트의 50 은 노랑 아래 경계로 남고 25·10 은 빠져요.
- 기본값: 옵션 B. 반대로 되돌리는 방법은 `render.rs` 의 상수 세 개를 바꾸고 테스트 기대값을 같이 바꾸는 것뿐이에요(구현 전이면 SPEC 개정). HOLD 범주도 `guarded` 도 아니고 이 저장소 안의 되돌릴 수 있는 선택이에요. 사용자가 고른 것은 "4단계 색"이고 숫자는 사용자가 쓴 것이 아니라 추천 선택지의 예시값이라서 리더 결정으로 정하기를 청해요.

## Interview

- Q1: 막대 한 개가 한 색인가요, 칸마다 다른 색인가요? → A: 한 막대가 한 색이에요(채운 칸 전부). 빈 칸은 지금의 흐린 색이에요(카드 노트 7983, 사용자 직접 선택 "4단계 색").
- Q2: 단계표 숫자는요? → A: 옵션 B 예요 — 초록 70 이상, 노랑 50~69, 주황 30~49, 빨강 30 미만(`LOW_LEFT_PERCENT`). 노트 7983 의 50/25/10 은 `LOW_LEFT_PERCENT` 를 10 으로 본 예시값이라서 바꿨어요(리더 결정 노트 7989).
- Q3: 색이 없는 환경은요? → A: 출력이 글자 하나까지 그대로예요(REQ-5, 노트 7983).
- Q4: 골든 파일은요? → A: `.ansi*` 골든 42개와 `cases.json` 의 체크섬 줄만 바뀌고 `.plain.txt` 는 하나도 안 바뀌어요(리더 조건, 노트 7989).

## Non-goals

- 칸마다 색을 이어 붙이는 옛 그라데이션(`interpolate`)을 되살리는 것: 이번에는 한 막대가 한 단계 색이에요.
- 막대의 머리(`🔋 CW`), 숫자(`73%`), 카운트다운, 배터리 아이콘의 색과 모양, 아이콘 경계(`LOW_LEFT_PERCENT`)의 변경.
- 막대 폭(`BAR_WIDTH`), 채운 칸 수의 식, 알 수 없는 창(`CW n/a`, 막대 없음)의 모양.
- 1·2번째 줄과 다른 세그먼트의 색, 새 테마, 색 프로필 고르기(`COLORTERM`)의 변경.

## Requirements

- REQ-1: When a color profile is on, the status line shall paint all filled cells of the bar of each known window `CW`, `5H` and `7D` in one level color chosen from the share left of that window · failure: none — a bar with no filled cell has no painted run of filled cells
- REQ-2: The status line shall choose the level color from the share left in four steps: green at `LEVEL_GREEN_FROM` (70) or more, yellow from `LEVEL_YELLOW_FROM` (50) up to 69, peach from `LOW_LEFT_PERCENT` (30) up to 49 and red below `LOW_LEFT_PERCENT` · failure: none — the steps follow the share left that the number shows and never the share used
- REQ-3: Each theme shall carry four level colors in the order green, yellow, peach and red: mocha `#A6E3A1` `#F9E2AF` `#FAB387` `#F38BA8` and latte `#40A02B` `#DF8E1D` `#FE640B` `#D20F39` · failure: none — an unknown theme name is mocha as before
- REQ-4: The status line shall keep the empty cells of a bar in the muted color and the head and the number and the countdown of a window in the base color of the line · failure: none — the two runs of a bar still touch with no space between them
- REQ-5: When no color profile is on, the status line shall write the same bytes as before this change with no escape sequence · failure: none — the colored line stripped of its escapes is the plain line at every width

## Acceptance

- AC-1 (REQ-1, REQ-3): 두 테마와 두 프로필(TrueColor, Ansi256)에서 `CW`·`5H`·`7D` 의 채운 칸 조각이 정확히 `profile.paint(<그 단계의 테마 색>, "█"×n)` 이고 빈 칸 조각이 `profile.paint(muted, "░"×m)` 이며 둘 사이에 틈이 없어요 · verify: `cargo test --locked --offline --lib statusline::render::window_tests::the_filled_cells_carry_the_level_color`
- AC-2 (REQ-2): 단계 경계가 N/N+1 로 고정돼요 — 남은 비율 70 은 초록, 69 는 노랑, 50 노랑, 49 주황, 30 주황, 29 빨강 — 기대값은 `render.rs` 의 상수에서 나오고, 0~100 의 모든 값에서 색 단계가 남은 비율이 줄 때 초록, 노랑, 주황, 빨강 순서를 거꾸로 가지 않아요 · verify: `cargo test --locked --offline --lib statusline::render::window_tests::the_level_color_changes_at_each_step_point`
- AC-3 (REQ-4): 머리, 숫자, 카운트다운은 칠해지지 않고(줄의 기본색), 채운 칸이 없는 막대(남은 비율 0~9)는 빈 칸 조각 하나만 흐린 색으로 나와요 · verify: `cargo test --locked --offline --lib statusline::render::window_tests::only_the_filled_cells_take_the_level_color`
- AC-4 (REQ-5): 색 없는 출력과 `.plain.txt` 골든이 이 변경 전과 글자 하나까지 같고, 색 있는 줄에서 이스케이프를 뺀 글자가 색 없는 줄과 같아요 · verify: `cargo test --locked --offline --lib statusline::render` 와 `git diff --name-only <base>..<head>` 에 `.plain.txt` 가 없음
- AC-5 (REQ-1, REQ-3): 색 골든 파일(`*.ansi.txt`, `*.ansi256.txt` 의 막대가 있는 21+21개)이 새 막대 색으로 다시 쓰였고 나머지 골든은 그대로예요 · verify: `cargo test --locked --offline --test statusline_golden`

## Scope

- write: `.moai/specs/SPEC-20261124-statusline-bar-color/**`, `src/statusline/render.rs`, `tests/fixtures/statusline-claude/golden/*.ansi.txt`, `tests/fixtures/statusline-claude/golden/*.ansi256.txt`, `tests/fixtures/statusline-claude/cases.json` (the binary checksum line the golden tool writes), `docs/rust-product-path.md` (the bar-color sentence of item 3 only)
- read-only: `src/statusline/color.rs`, `src/statusline/state.rs`, `tools/statusline_golden.py`, `tests/statusline_golden.rs`
- forbidden: `tests/fixtures/statusline-claude/golden/*.plain.txt`, `src/statusline/color.rs`, every file not listed under write

## Plan

1. SPEC 초안을 커밋하고 리더에게 D-1 승인을 청해요(스스로 승인하지 않아요). 범위 넓히기(SCOPE-REQUEST)도 같이 청해요.
2. RED: `window_tests` 에 새 테스트 세 개(AC-1~AC-3)를 먼저 쓰고 부딪히는 기존 색 단언을 새 규칙으로 고쳐요. 코드 없이 실패를 확인해요(`cargo nextest run --lib -E 'test(statusline::render)'`).
3. 구현: `Theme` 에 `level: [&str; 4]`(초록, 노랑, 주황, 빨강)를 더하고 두 테마에 값을 적어요. `LEVEL_GREEN_FROM`, `LEVEL_YELLOW_FROM` 상수와 `level_color(theme, left)` 를 더하고, `Window::known` 의 채운 칸 조각 색을 그 값으로 바꿔요.
4. 골든: `tools/statusline_golden.py --write` 로 색 골든을 다시 쓰고 `--check` 로 확인해요(도구 실행이고 직접 돌리는 테스트 모듈이 아니에요). 문서 한 문장을 맞춰요.
5. guard-mutant 둘(색 빼기, 단계 순서 뒤집기)을 복사본에서 한 번씩 돌려 `exit` 코드를 READY 에 적어요. 복구는 `git checkout` 이에요. 그다음 `moai check affected` 를 맨몸으로 한 번 돌려요.

## Risks

- 단계표 숫자(D-1)가 사용자의 기대와 다를 수 있어요. 결정이 SPEC 의 Decisions 에 남고 상수 세 개로 바꿀 수 있어요.
- 색 골든 42개와 `cases.json` 의 체크섬이 한꺼번에 바뀌어 diff 가 커져요. 바뀐 파일이 `.ansi*` 뿐이고 `.plain.txt` 가 하나도 없음을 `git diff --name-only` 로 확인해요.
- `docs/rust-product-path.md` 는 문서를 읽는 테스트가 있을 수 있어요. 고친 뒤 그 문서를 읽는 테스트 타깃을 `moai check affected` 가 고르는지 `--list` 로 봐요.
- Ansi256 프로필에서 네 색이 같은 256색 번호로 겹치면 단계가 구분되지 않아요. AC-1 이 두 프로필의 실제 `paint` 결과를 비교하고, 겹침이 있으면 Result 에 적어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 단계표는 옵션 B(초록 70 이상, 노랑 50~69, 주황 30~49, 빨강 30 미만; 빨강은 `LOW_LEFT_PERCENT`)로 해요. 리더 결정 노트 7989 가 승인했어요 | 노트 7983 은 빨강을 `LOW_LEFT_PERCENT`(=30, 10 이 아님)에 맞추라고 했고 50/25/10 은 그것과 모순이며 10 미만은 채운 칸이 없어요(읽고 잰 것 2, 3) | `render.rs` 의 `LEVEL_GREEN_FROM`·`LEVEL_YELLOW_FROM` 과 빨강 경계를 바꾸고 AC-2 기대값을 같이 바꿔요 |
