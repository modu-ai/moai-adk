---
id: SPEC-20261122-statusline-flow
title: "The status line shows the cards by the stage of their progress: my card first, what waits for me, what is blocked, then the flow and the limits"
status: draft
revision: 1
created: 2026-11-22
---

# 상태줄: 카드를 진행 단계 순서로 보여 주기 (내 카드 먼저)

## Intent

상태줄 3번째 줄은 지금 `🗂️ 200 ▶️ 15 🔃 15 ✅ 95 | 🪫 CTX 74% | 🔋 Week 21% (resets in 6d 21h)` 이에요. 숫자는 원장과 맞지만(측정 2026-11-21T23:2xZ, 아래 표) 사용자가 흐름을 한눈에 읽기엔 부족하다고 짚었어요. 리더 노트 5921 의 조사 결과가 이유 일곱 가지예요:

- (a) `✅` 는 평생 합계(95)라서 오늘 닫은 수(19)가 안 보여요.
- (b) `🗂️` 는 draft 43 + ready 150 + hold 7 을 한 숫자로 섞어요. hold 는 어디에도 따로 안 나와요.
- (c) `🔃` 는 "기다림"이고 제출됨과 감사 통과를 섞어요(아이콘은 사용자가 2026-11-19 에 골랐어요).
- (d) 잡힌 active 예약 26개 중 22개는 임대가 끝났는데도 `▶️` 로 세요.
- (e) 마커 규칙이 `MERGE-READY`·`LANE-BLOCKED`·`READY`·`RESUME` 첫 줄을 못 읽어요.
- (f) 모든 숫자가 모든 레인의 합이고 "내 것"은 줄 2 의 작업 폴더 이름표뿐이에요.
- (g) 제품 문서가 아직 카드 숫자가 줄 1 에 있다고 써요.

사용자 선택(2026-11-22, 노트 5922): 배치 B, **내 카드 먼저**. 레인 줄은 내 카드와 그 단계, 나를 기다리는 것(제출됨·감사 통과·착지됐지만 열린 것), 막힘과 보류, 흐름 숫자와 오늘 닫은 수, 맥락 숫자 순서예요. 리더 줄은 먼저 착지할 것(감사 통과·제출됨), 막힘과 착지됐지만 열린 것, 맥락이에요. 사용자가 더한 말: 주간 한도를 **7D**(7일 세션 한도)로 이름 붙여 맥락 숫자 옆에 보여 주세요. 모든 아이콘 옆에는 글자나 숫자가 따라붙고 색만으로 뜻을 나르지 않아요.

이 SPEC 은 코드가 없는 첫 조각이에요. 단계(stage)를 어떻게 세는지, 줄이 어떤 모양들을 갖는지, 너비를 어떻게 읽는지를 정하고, 코드 조각 셋의 순서와 시험·골든이 바뀌는 곳을 적어요.

### 지금 무엇을 세는지(조사 표 재현)

측정 도구와 시각: 조사 에이전트가 `sqlite3 -readonly ~/.moai/projects/moai-adk/ledger.db` 로 2026-11-21T23:21Z–23:24Z 에 읽었어요(노트 5921). 상태줄 코드 위치는 tip `7f1da593d1` 에서 읽었어요.

| 항목 | 코드가 세는 것 | 위치 | 그때 값 |
|---|---|---|---|
| `🗂️` | draft + ready + hold | `src/statusline/state.rs:943` (`remaining`) | 43 + 150 + 7 = 200 |
| `▶️` | (active + review) − 대기, 0 아래로 안 내려가요 | `state.rs:942,950-952,1127-1129` (`in_progress`, `split_in_progress`) | 15 |
| `🔃` | review 전부 + 새 마커 노트가 대기 마커인 active | `state.rs:1145` (`marker_of`), `1179` (`waiting_cards_with`) | 15 (13 + 2) |
| `✅` | done 평생 합계 | `state.rs:944` (`done`) | 95 |

내가 같은 원장에 새 규칙(아래 D-5·D-6)을 SQL 로 옮겨 센 값(2026-11-21T23:45:39Z, 이 폴더의 `stages.sql` 을 `sqlite3 -readonly ~/.moai/projects/moai-adk/ledger.db < stages.sql` 로 읽기만 했어요): A 5, B 4, D 43, E 6, H 7, R 149, S 13, W 3 — 합 230 은 `state <> 'done'` 인 카드 수 230 과 같고, 오늘(UTC) 닫은 수는 19, 평생 done 은 95 예요. 지금 줄의 `▶️` 15 와 `🔃` 15 는 새 규칙에서 W 3, E 6, S 13, A 5, B 4 로 갈라져요. 이 SQL 은 SPEC 의 규칙을 옮긴 것이라 상태줄 코드 값과 같은지는 코드 조각의 시험이 확인해요(이 조각에서는 코드를 돌리지 않았어요).

## Non-goals

- 줄 1 과 줄 2 의 모양, 줄 2 끝의 작업 폴더 이름표(`🌳 WT-c<card>-<name>`)는 그대로예요. 내 카드 그룹(🎯)과 이름표가 같은 카드를 두 번 말하는 것은 알고 두는 거예요(D-2).
- 원장 쓰기, 새 상태, 새 명령, 새 표. 상태줄은 읽기만 해요. 유형 있는 신호 줄(typed rows)과 `card inbox` 는 별도 카드가 해요.
- 5시간 창(`usage_5h`)을 보여 주는 것. 사용자가 고르지 않았어요.
- 색과 테마의 변경. 색은 항목 종류별 팔레트를 그대로 쓰고, 새 항목은 가까운 기존 색을 받아요.
- 파이썬 도구를 늘리는 것. 골든은 기존 `tools/statusline_golden.py --write` 로 다시 쓰고, 새 시험은 Rust 만이에요.

## Requirements

- REQ-1: The status line shall place every open card of the ledger in exactly one of the stages draft, ready, working, stale, submitted, audited, blocked, landed and hold so that the nine counts add up to the number of cards that are not done and a card in the ledger state `draft`, `ready` or `hold` is in the stage of the same name · failure: none — a ledger that cannot be read shows no stage (REQ-11)
- REQ-2: When a card is in the ledger state `active` or `review`, the status line shall take its stage from the newest of its notes whose first line starts with a marker of the widened rule (D-5) and from the ledger state when no note carries one · failure: none — a note that is no marker changes nothing
- REQ-3: When a card has a `card.landed` event and is not `done`, the status line shall place it in the stage landed whatever its notes say · failure: none — a card without the event keeps the stage of REQ-2
- REQ-4: When an `active` card would be working and no reservation of it is `held` or `held_for_review` with a lease that ends at or after the render's clock, the status line shall place it in the stage stale · failure: none — a live lease leaves the card working
- REQ-5: The status line shall count as done today the cards whose `closed_at` is at or after the start of the UTC day of the render's clock · failure: none — the lifetime total is no longer shown on the line
- REQ-6: When a reservation that is `held` or `held_for_review` names the session by the stdin session id or by its exact lane name or names the folder of the session's worktree, the status line shall open the lane line with `🎯 c<n> <letter>` for the lowest numbered such card · failure: none — a session that holds no card has no 🎯 group
- REQ-7: While the session is the leader, the status line shall open line 3 with `🎯 land A<n> S<n>` and follow it with `⛔ B<n> L<n>` · failure: none — the leader line has no flow counts and no held or stale items
- REQ-8: The status line shall write the groups of the lane line in the order mine, waits `👀 S<n> A<n> L<n>`, attention `⛔ B<n>` then `⏸️ H<n>`, flow `🗂️ D<n> R<n> W<n> E<n> ✓<n>`, context and seven-day with an icon and a letter or a number on every item · failure: none — a group a form leaves out takes its separator with it and the plain text equals the colored text with its escapes removed
- REQ-9: The status line shall write the seven-day window as `<battery> 7D <n>%` right after the context window with its reset countdown when the host gives a reset time · failure: none — a missing window is hidden and an unknown percentage says `7D n/a` as today
- REQ-10: When `COLUMNS` holds a whole number above zero, the status line shall write line 3 in the widest form whose display width is at most `COLUMNS` minus two and in the narrowest form when none fits · failure: none — a missing, empty, zero or non-numeric `COLUMNS` writes the widest form (today's no-shortening behavior) and lines 1 and 2 keep the same bytes at every width
- REQ-11: The status line shall read the stages read-only through the connection and the budget the card counts use and in no more than four statements · failure: none — when a read cannot finish within the budget the line shows only the ledger-state counts `🗂️ D<n> R<n>`, `▶️ P<n>` for active plus review and `⏸️ H<n>` with no split, never a guess, and the ledger file and the HOME tree stay the same bytes
- REQ-12: The statusline documentation shall describe line 3 in its forms with the legend of every letter, carry the golden lines it cites verbatim and say that `COLUMNS` is read for line 3 only · failure: none — a stale sentence is found by the doc test of AC-9

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): a fixture ledger holds one card for every edge of the stage rule, and the nine counts equal the open cards at every step.
  verify: `cargo test --offline --locked --lib statusline::state::stage_tests`
- AC-2 (REQ-2): the marker words of D-5 are pinned at their edges — each word alone, a first line only (a marker on line two is none), a note head cut at exactly 1000 characters, the newest marker winning over an older one of another kind, `RESUME` and `waiting:` and `LANE-HOLD` and `PAIR-` as no marker, `READY ` with its space against `READYX`.
  verify: `cargo test --offline --locked --lib statusline::state::marker_tests`
- AC-3 (REQ-5): the done-today count changes at the UTC midnight and not a second before.
  verify: `cargo test --offline --locked --lib statusline::state::done_today_tests`
- AC-4 (REQ-6, REQ-7): the real binary shows `🎯 c<n> <letter>` for a session id, a lane name and a worktree folder, the lowest card when two hold, no group without a card, and the leader group for a leader.
  verify: `cargo test --offline --locked --test statusline_flow my_card_`
- AC-5 (REQ-7, REQ-8): the golden lines of the full forms — lane with a card, lane without one, leader — are exactly the strings of the Layouts section, plain and with color.
  verify: `cargo test --offline --locked --lib statusline::render::flow_line_tests`
- AC-6 (REQ-9): the `7D` label, the icon switch at 70/71, the countdown, the hidden missing window and `7D n/a` are pinned.
  verify: `cargo test --offline --locked --lib statusline::render`
- AC-7 (REQ-10): at each ladder width the form changes exactly at the boundary — a width equal to the form's width plus two keeps it and one less drops to the next — `COLUMNS` of `abc`, empty, `0` and absent give the widest form, and lines 1 and 2 keep the same bytes at 40, 80 and 120.
  verify: `cargo test --offline --locked --test statusline_cards line_three_reads_columns_and_lines_one_and_two_do_not`
- AC-8 (REQ-11): a spent budget before or between the reads gives the state-only form, and a render leaves the ledger digest and the HOME tree as they were with at most four statements seen by a counting hook.
  verify: `cargo test --offline --locked --test statusline_flow reading_the_stages_writes_nothing_and_asks_four_times_at_most`
- AC-9 (REQ-12): the document test finds the legend, the cited golden lines, the `COLUMNS` sentence and none of the stale sentences (the card counts on line 1, `Week` as a label, "shortens nothing").
  verify: `cargo test --offline --locked --test statusline_numbers_only the_statusline_doc_describes_the_flow_line`
- AC-10 (REQ-1, REQ-8, REQ-9, REQ-10): every statusline target passes with the goldens rewritten and every changed golden file read.
  verify: `cargo test --offline --locked --test statusline_cards --test statusline_golden --test statusline_tag_order --test statusline_worktree_label --test statusline_git_env --test statusline_counts_refresh --test statusline_home_state --test statusline_guard_false_positive --test statusline_numbers_only`

### Layouts

Every example below is a fixture: `c12` is a made-up card number, and the numbers are the ones the research read (open cards at 2026-11-21T23:2xZ), so the lines can be measured. Widths are display columns from `width.py` of the research (an emoji with its variation selector counts 2); the Rust code measures with `unicode-width`, already in `Cargo.toml`.

**Lane line (the session holds card c12, working).** Items: 🎯 mine · 👀 `S` submitted `A` audited `L` landed but open · ⛔ `B` blocked · ⏸️ `H` hold · 🗂️ `D` draft `R` ready `W` working `E` stale-lease · `✓` done today · CTX · 7D.

| form | text | cols |
|---|---|---|
| L0 widest | `🎯 c12 W | 👀 S11 A2 L0 | ⛔ B3 ⏸️ H7 | 🗂️ D43 R150 W14 E22 ✓19 | 🪫 CTX 74% | 🔋 7D 21% (resets in 6d 21h)` | 108 |
| L1 no flow | `🎯 c12 W | 👀 S11 A2 L0 | ⛔ B3 ⏸️ H7 | 🪫 CTX 74% | 🔋 7D 21% (resets in 6d 21h)` | 81 |
| L2 no countdown | `🎯 c12 W | 👀 S11 A2 L0 | ⛔ B3 ⏸️ H7 | 🪫 CTX 74% | 🔋 7D 21%` | 62 |
| L3 no hold | `🎯 c12 W | 👀 S11 A2 L0 | ⛔ B3 | 🪫 CTX 74% | 🔋 7D 21%` | 56 |
| L4 narrowest | `🎯 c12 W | ⛔ B3 | 🪫 CTX 74% | 🔋 7D 21%` | 41 |

**Leader line.** Items: 🎯 land `A` audited `S` submitted · ⛔ `B` blocked `L` landed but open · CTX · 7D. Held and stale are not on it; the leader reads them from `moai card inbox`.

| form | text | cols |
|---|---|---|
| M0 widest | `🎯 land A2 S11 | ⛔ B3 L0 | 🪫 CTX 74% | 🔋 7D 21% (resets in 6d 21h)` | 69 |
| M1 narrowest | `🎯 land A2 S11 | ⛔ B3 L0 | 🪫 CTX 74% | 🔋 7D 21%` | 50 |

A session that holds no card and is not the leader writes the lane line without its 🎯 group (the `👀 S11 A2 L0 | ⛔ B3 | …` forms).

**Width rule.** `COLUMNS` is set by Claude Code to the current terminal width before the status line command runs; the stdin JSON has no width field (docs read 2026-11-22: "Claude Code sets these to the current terminal dimensions before running your script"; the page also says long lines "may get truncated or wrap awkwardly"). The budget is `COLUMNS − 2`. The forms are tried from the widest; the first whose measured width fits is written. Read at the ladder: 120 columns gives L0 (108 ≤ 118), 100 gives L1 (81 ≤ 98), 80 gives L2 (62 ≤ 78), 60 gives L3 (56 ≤ 58), 40 gives L4 (41 does not fit, the narrowest form is written anyway). The leader line gives M0 from 71 columns up (69 + 2) and M1 below (50).

**The stage letters** (the legend the document carries): D draft · R ready · W working · E stale (a working card whose lease ended or that nobody holds) · S submitted (the lane says it waits for the leader or an audit; also every `review` card) · A audited (an audit passed for the head and nobody landed it) · L landed but not closed · B blocked (the lane asked for a decision) · H hold · ✓ done today · P active plus review, unsplit (only in the fallback).

## Scope

- write: `.moai/specs/SPEC-20261122-statusline-flow/**` (this slice: `spec.md` and `stages.sql`). The code slices of the Plan write `src/statusline/{state,render,mod}.rs`, `tests/statusline_*.rs`, `tests/fixtures/statusline-claude/**` (cases, stdin, ledgers, goldens), `docs/rust-product-path.md` (the statusline section only) and the spec folders named under Supersedes only through their Result lines.
- read-only now: `src/statusline/**`, `src/ledger/schema.rs`, the statusline SPECs, `tools/statusline_golden.py`.
- forbidden: the real HOME and ledger (the research's `sqlite3 -readonly` reads are measurements, not tests), the network, any write to a ledger from the status line, a new crate.

## Plan

Each step is its own landing; the first two do not change what the user sees except the label.

1. **7D label.** `Week` becomes `7D` on line 3 and everywhere a test or document says it (render unit tests, `statusline_cards`, `statusline_numbers_only`, `statusline_tag_order`, the golden cases that carry `seven_day`, the document). Alone, so its golden diff is only that word.
2. **Stages in `state.rs`** (red tests first). `CardCounts` (used only inside `src/statusline/`) becomes a `Flow` holding the nine counts, done today and my card's stage; `marker_of` gets the widened words in the one existing function (no second parser); one read for the notes of `active` and `review` cards, one for the `card.landed` events, one for the reservations, one for the state counts and done today — four statements. Nothing is drawn differently yet except through step 3.
3. **The line.** `Context` gets `columns` (read from `COLUMNS`, absent or bad means none); `render.rs` builds the lane and leader lines as ordered groups, measures forms with `unicode-width` and picks by the width rule; the fallback form; goldens rewritten case by case with the existing switch and every changed file read (the cases with cards are `b13` and `b14`; every other line 3 changes only in what REQ-11 and the card group change); `tests/statusline_flow.rs` is the new real-binary target with the fixture ledgers written by the ignored fixture-writer test; the document and the legend.
4. Before step 3 starts: print `COLUMNS` and the usable width once in a real session with a probe command, to check the margin of two and the emoji widths on this terminal (see Risks). Not done in this slice.

Order of landing: 1, then 2, then 3 (step 4 first in 3). A step that finds this SPEC wrong writes a new revision before code.

### Supersedes

This SPEC changes what earlier SPECs pinned; they stay as the record of what was true and each is cited by the requirement it changes.

| earlier SPEC | requirement | what changes |
|---|---|---|
| `SPEC-20261120-statusline-layout` | REQ-9 (same bytes at 40 and 80 columns for every line) | only for line 3; lines 1 and 2 keep it (REQ-10) |
| `SPEC-20261120-statusline-layout` | REQ-2 (card counts `🗂️ ▶️ 🔃 ✅` open line 3) | the group is replaced by the groups of REQ-7 to REQ-9 |
| `SPEC-20261119-statusline-waiting-cards` | REQ-1 to REQ-5 (four numbers; waiting markers) | stages and the widened rule (D-5); `review` stays waiting as stage S |
| `SPEC-20261121-statusline-detached` | REQ-5 (`Week <n>%`) | the label is `7D` |
| `SPEC-20261120-statusline-docs-items-2-3` | AC-1 to AC-4 (its `check.py` and the example lines) | the document and its example follow step 3 |

## Risks

- **Width is a guess until observed.** The margin of two and the emoji widths (`⏸️` and `🗂️` are narrow-by-default characters that need the variation selector, so terminals differ) come from the host page and the research script, not from this terminal. Plan step 4 measures once; if a terminal draws wider, the margin becomes a constant to raise, not a redesign.
- **The marker words are a first-line convention.** The master SPEC of the flow skeleton lists five first-line parsers that disagree (`src/statusline/state.rs:1148` is one) and plans typed event rows to replace them; its AC-17 refuses a new parser. This SPEC adds words to the existing function only and says it is temporary: when typed submit, block and verdict rows exist, `marker_of` is replaced by a read of them. The leader note said the typed rows are "step 3"; the master SPEC numbers them step 4 (4a to 4f) and uses step 3 for admission and the limit. This SPEC follows the master SPEC and flags the mismatch.
- **`RESUME` is not a marker, on evidence.** The notes with that first word are 245; the newest ones say "weekly usage limit reached" (note 5925) and "returned the lane" (note 5915): they are written when work stops. Reading them as working would show a stopped card as W. `waiting:` (110 notes) is a lane waiting on its own running check, so it is working without a marker. Counts read from `sqlite3 -readonly` at 2026-11-21T23:44:48Z over all notes of all cards.
- **Session id equals the holder, by code reading only.** `card take` writes the holder from `CLAUDE_CODE_SESSION_ID` (`src/card/command.rs:275`) and the statusline already uses the stdin `session_id` for the goal file; that the two are the same string in a live session was not observed. Step 2 starts by checking it with the shipped binary; if they differ, REQ-7 falls back to the worktree folder, which exists and is tested.
- **Cost.** Four statements instead of two; the notes read already covers every note of every active card and now also of review cards (about 28 cards, bounded by the 1000-character cut). The 50 ms lock wait and the 800 ms budget stay. A timing measure on a copy of the real ledger goes in the Result of step 2, not in a test (a timing test would flake).
- **Letters need a legend.** A reader meeting `E22` for the first time cannot guess it. The legend is in the document and in the help of `moai statusline`; a test pins both against the letters the renderer writes.
- **Golden churn.** Step 1 changes every golden with a seven-day window; step 3 changes every line 3. The golden harness is Python (`tools/statusline_golden.py`); it is used as it is, not extended, until the Python removal ports it.

## Decisions

Reversible, category `none`, decider the leader (the user's pick of the layout is recorded as D-1). A step finding a better value writes a revision before code.

| Id | Decision | Basis | Undo |
|---|---|---|---|
| D-1 | layout B, my card first, with the 80/120-column variants; the pipeline strip and the worded stages of the research are not built | the user's pick 2026-11-22, note 5922 | a new revision choosing another form |
| D-2 | only line 3 changes; the worktree label stays on line 2 and the 🎯 group may repeat its card | the user asked for line 3; the label also carries the name and the host's fallback, and removing it was not asked | drop the label's card number |
| D-3 | line 3 reads `COLUMNS`; absent or unreadable means the widest form; margin 2; forms drop whole items in the order flow, countdown, hold, waits | the host sets `COLUMNS` and has no stdin width (read 2026-11-22); the user asked for 80 and 120 column variants; the widest form as fallback keeps today's behavior wherever width is unknown | one constant to ignore `COLUMNS` |
| D-4 | the label is `7D`, next to the context number; it stays hidden when the host gives no seven-day window | the user's note; the host gives `rate_limits` only to subscribers and only after the first reply (docs read 2026-11-22) | label back to `Week` |
| D-5 | the widened marker rule: submitted = first line starts `LANE-READY-INTEGRATION`, `MERGE-READY`, `READY ` (word and one space), `감사 요청`, or `LANE-DONE` with `audit requested` or `감사 요청`; audited = `AUDIT` with `verdict: PASS`; blocked = `LANE-BLOCKED`; working = `ACK `, `PLAN `, or `AUDIT` with `verdict: FAIL`; everything else, `RESUME`, `waiting:`, `LANE-HOLD` and `PAIR-` included, is no marker; the newest marker of a card wins; a card with no marker is working | notes 5921; the counts of the Risks; the note first-line prefixes read at 2026-11-21T23:44:48Z (`READY` 52, `MERGE-READY` 86, `LANE-BLOCKED` 325) | remove a word from the one function |
| D-6 | stale = a working card whose lease has ended or that has no live reservation; it is its own stage so `W` counts only cards someone holds; a submitted, audited or blocked card keeps its stage whatever its lease says, because it waits for someone else | finding (d): 22 of 26 held reservations had ended leases; the research row "stale" | count stale as working again |
| D-7 | done today = the UTC day of the render's clock | the ledger and every report use UTC; a rolling 24 hours is the alternative and avoids a reset at 09:00 in the user's zone | one expression |
| D-8 | my card = a held reservation by session id, else by exact lane name (`agent-<n>`), else by the label's worktree folder; the lowest number wins | the holder is the session variable today and the lane name after the master SPEC's step 3e | drop a source |
| D-9 | the leader line has no flow, hold or stale items; "the leader" is the existing `status.leader` (`state.rs:232`, `is_leader` at `state.rs:499`: a live `leader` entry of the lane registry for the stdin `session_name`) | the user's pick lists land, blocked, landed-but-open and context for it; `moai card inbox` reads the rest | add the items as a wide form |
| D-10 | letters D R W E S A L B H P and `✓`; English only; a legend in the document and the help | the user's rule: every icon keeps a letter or a number | rename a letter in one table |

## Result

이 head 는 SPEC 만이에요. 구현과 측정은 코드 조각에서 채워요. 이 조각에서 돌린 것: 위 조사 표의 원장 읽기(읽기 전용), 레이아웃 너비 측정(`width.py`), `spec check`. 돌리지 않은 것: 실제 Claude Code 세션에서의 `COLUMNS` 확인과 세션 id 대조(Risks 에 이유를 적었어요), 위 시험들(아직 없어요).
