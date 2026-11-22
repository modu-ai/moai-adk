---
id: SPEC-20261122-statusline-flow
title: "The status line gives each of its three lines one role, and its flow line shows the cards by the stage of their progress, my card first"
status: draft
revision: 2
created: 2026-11-22
---

# 상태줄: 세 줄에 역할 하나씩, 흐름 줄은 진행 단계 순서로 (내 카드 먼저)

## Intent

**개정 2 (2026-11-22).** 개정 1 은 3번째 줄만 다뤘고, 그 첫 조각(이름표 `CW`·`7D`)은 이미 통합 브랜치에 들어갔어요. 이번 개정은 사용자가 고른 **줄 전체의 설계**를 담아요. 이 SPEC 은 줄 전체의 계약(역할, 단계 규칙, 흐름 줄, 너비 사다리, 아이콘 띄어쓰기, 문서)을 갖고, 1번째 줄과 3번째 줄의 내용은 아래 두 하위 SPEC 이 가져요. 한 SPEC 이 요구사항 12개를 넘으면 `spec check` 가 나누라고 하기 때문이에요.

| SPEC | 맡은 것 |
|---|---|
| `SPEC-20261122-statusline-flow` (이 문서) | 세 줄의 역할, 단계 규칙과 흐름 줄(2번째), 너비 사다리, 아이콘 띄어쓰기, 문서 |
| `SPEC-20261122-statusline-where` | 1번째 줄(누구이고 어디에 있나), git 읽기와 세션별 캐시, 리더 태그 삭제 |
| `SPEC-20261122-statusline-limits` | 3번째 줄(한도), 사용량 기록 옮기기, Orca 에 입력 넘기기, 생성 설정, 입력 읽기 |

**사용자 선택(2026-11-22, 리더 결정 기록 5922·5935·5942).**

- 배치 B, 역할 세 줄: 1번째 줄 누구이고 어디에 있나(모델, 저장소, 작업트리, 브랜치, 통합 브랜치와의 거리, PR), 2번째 줄 무엇을 하나(흐름 줄), 3번째 줄 한도(`CW`, `5H`, `7D`, 비용, 버전). 80칸·120칸 모양이 따로 있어요.
- 이름표는 `CW`(context window)와 `7D`(seven-day limit). `CW` 는 호스트가 준 사용 비율 그대로(`/context` 와 같은 숫자). `5H` 는 70% 를 넘을 때만. 비용은 `💰`, 한도 정보가 없을 때만.
- 그릴 때 읽기만 한다는 옛 규칙을 바꿔요: `$TMPDIR` 에 세션별 캐시를 쓰고 git 은 바뀔 때만 실행해요. 생성 설정에 `refreshInterval: 30`.
- moai 가 받은 입력을 Orca 의 상태줄 래퍼에도 넘겨요(이 컴퓨터 안 loopback). 고르지 않은 것(그대로 둬요): 이슈·PR 수 `⬇️ 🔀` 와 그 `gh` 갱신.
- 모든 아이콘 뒤에 글자나 숫자가 따라붙고 색만으로 뜻을 나르지 않아요. 아이콘 뒤에는 띄어쓰기를 둬서 글자가 붙지 않게 해요(D-6 에서 방식을 정해요).
- 개정 1 의 사용자 선택은 그대로예요: 흐름 줄은 내 카드, 나를 기다리는 것, 막힘·보류, 흐름 숫자와 오늘 닫은 수 순서.

**개정 1 에서 고치는 것(리더 조사, 노트 5934, 아홉 가지).** 어디서 닫는지 표로 둬요.

| # | 고칠 것 | 닫는 곳 |
|---|---|---|
| 1 | 너비 예산은 `COLUMNS` 에서 여유와 padding 을 빼야 해요 | REQ-9, D-5, 위험 3 |
| 2 | 쉬는 동안 갱신(`refreshInterval`)과 실행 한 번의 비용 예산이 없어요 | `limits` REQ-9, `where` REQ-5·6, REQ-11 |
| 3 | 7D 근거에 지출 한도(`spend_limit`)와 "초기화 시각이 지나면 창이 사라진다"가 빠졌어요 | `limits` REQ-3·5, 위험 4 |
| 4 | 세로 줄 수(`LINES`)를 쓰지 않아요 | D-8 |
| 5 | 내 카드는 가장 낮은 번호가 아니라 작업트리 일치, 그다음 살아 있는 최신 임대 | REQ-6, D-3 |
| 6 | 리더 줄은 레인 등록부에 기대는데 상주 리더를 없애기로 했어요 | D-7, `where` REQ-8 |
| 7 | 문서 시험이 지울 문구("shortens nothing")를 구간 표지로 써요 | REQ-12, 계획 5단계 |
| 8 | 너비 경계는 `unicode-width` 로 고정, 글자 폭이 터미널마다 달라요 | REQ-9, D-6, 위험 2 |
| 9 | 제품 문서 세 곳이 코드와 달라요(카드 줄 위치, 색 끄기 조건, "네트워크 없음") | REQ-12 |

### 지금 무엇을 세는지(조사 표)

측정 도구와 시각: `sqlite3 -readonly ~/.moai/projects/moai-adk/ledger.db`, 2026-11-21T23:21Z–23:24Z(리더 노트 5921), 내 재질의 2026-11-21T23:47Z(`stages.sql`). 상태줄 코드는 통합 브랜치 끝에서 읽었어요.

| 항목 | 코드가 세는 것 | 위치 | 그때 값 |
|---|---|---|---|
| `🗂️` | draft + ready + hold | `src/statusline/state.rs:943` (`remaining`) | 43 + 150 + 7 = 200 |
| `▶️` | (active + review) − 대기, 0 아래로 안 내려가요 | `state.rs:942,950-952,1127-1129` | 15 |
| `🔃` | review 전부 + 새 마커 노트가 대기 마커인 active | `state.rs:1145` (`marker_of`), `1179` (`waiting_cards_with`) | 15 (13 + 2) |
| `✅` | done 평생 합계 | `state.rs:944` (`done`) | 95 |

`stages.sql` 의 규칙(아래 D-1·D-2)으로 같은 원장을 센 값(23:47Z): A 5, B 4, D 43, E 6, H 7, R 149, S 13, W 3 — 합 230 은 `state <> 'done'` 인 카드 수와 같고, 오늘(UTC) 닫은 수 19, 평생 95. `▶️` 15 와 `🔃` 15 는 W 3, E 6, S 13, A 5, B 4 로 갈라져요. 이 SQL 은 규칙을 옮긴 것이고 코드 값과 같은지는 코드 조각의 시험이 확인해요.

### 그릴 때 드는 비용(리더 조사 실측, 보고서 `statusline-redesign-2026-11-22.md` 3장)

| 조건 | 중앙값 | p90 |
|---|---|---|
| git 없음 | 10.0 ms | 11.6 ms |
| 파일 200개 저장소, 브랜치 위(git 1번) | 49.0 ms | 56.7 ms |
| 같은 저장소, 분리된 HEAD(git 3번, 부하 높음) | 88.9 ms | 123.4 ms |
| 이 저장소 `git status` 하나 | 69.7 ms | 105.1 ms |

원장 읽기(`sqlite3 -readonly`, 같은 보고서의 조사 자료): 상태 수 2 ms, active·review 노트 19 ms, 노트 머리 9 ms, `card.landed` 이벤트 3 ms, 예약 0.3 ms — 네 번 따로 읽으면 약 33 ms, 개정 1 의 `stages.sql` 한 문장(CTE)은 약 115 ms 였어요. 그래서 이 SPEC 은 네 문장으로 나눠 읽어요(REQ-11). Claude Code 는 갱신을 300 ms 모아서 한 번 부르고 실행 중인 스크립트를 취소해요(공식 문서 `statusline.md`, 2026-11-21T23:52Z 받음).

## Non-goals

- 원장 쓰기, 새 상태, 새 명령, 새 표. 상태줄은 원장을 읽기만 해요(캐시 파일은 원장이 아니에요). 유형 있는 신호 줄(typed rows)과 `card inbox` 는 별도 카드가 해요.
- 리더 전용 줄. 상주 리더와 레인 등록부가 없어져요(D-7). 착지자 역할이 원장에 생기면 새 개정으로 그 줄을 정해요.
- 색과 테마의 변경. 새 항목은 가까운 기존 색을 받아요.
- 파이썬 도구를 늘리는 것. 골든은 기존 `tools/statusline_golden.py --write` 로 다시 쓰고, 새 시험은 Rust 만이에요.
- 세로 줄 수(`LINES`) 읽기(D-8).

## Requirements

- REQ-1: The status line shall place every open card of the ledger in exactly one of the stages draft, ready, working, stale, submitted, audited, blocked, landed and hold so that the nine counts add up to the number of cards that are not done · failure: none — a ledger that cannot be read shows no stage (REQ-11)
- REQ-2: When a card is in the ledger state `active` or `review`, the status line shall take its stage from the newest of its notes whose first line starts with a marker of the widened rule (D-1) and from the ledger state when no note carries one · failure: none — a note that is no marker changes nothing
- REQ-3: When a card has a `card.landed` event and is not `done`, the status line shall place it in the stage landed whatever its notes say · failure: none — a card without the event keeps the stage of REQ-2
- REQ-4: When an `active` card would be working and no reservation of it is `held` or `held_for_review` with a lease that ends at or after the render's clock, the status line shall place it in the stage stale · failure: none — a live lease leaves the card working
- REQ-5: The status line shall count as done today the cards whose `closed_at` is at or after the start of the UTC day of the render's clock · failure: none — the lifetime total is no longer shown
- REQ-6: The status line shall open its flow line with `🎯 c<n> <letter>` for the card whose held reservation names the session's worktree and else for the newest live held reservation of the stdin session id · failure: none — a session that holds no card has no 🎯 group
- REQ-7: The status line shall write the flow line as the groups mine, waits `👀 S<n> A<n> L<n>`, attention `⛔ B<n> ⏸️  H<n>` and flow `🗂️  D<n> R<n> W<n> E<n> ✅ <n>` in this order · failure: none — a waits group whose three counts are zero and an attention group whose counts are zero are left out with their separator
- REQ-8: The status line shall give its three lines the roles who and where, what and limits in this order · failure: none — a line with nothing to show is left out and the next one moves up
- REQ-9: The status line shall write each line in the widest of its forms whose display width is at most `COLUMNS` minus two minus the configured padding and in its narrowest form when none fits · failure: none — a missing, empty, zero or non-numeric `COLUMNS` writes the widest form and a width never changes the content of a form
- REQ-10: The status line shall write one space after every icon and two when the icon is narrow by default and carries a variation selector so that no text touches an icon · failure: none — the display width of an icon is counted as the `unicode-width` crate counts its whole sequence, never less
- REQ-11: The status line shall read the stages read-only in no more than four statements within the render budget · failure: none — a read that cannot finish shows only the ledger-state counts `🗂️  D<n> R<n>` and `W<n>` for active plus review unsplit, never a guess, and the ledger file stays the same bytes
- REQ-12: The statusline documentation shall describe the three lines, the legend of every letter and the width rule and carry the golden lines it cites verbatim · failure: none — a stale sentence is found by the doc test of AC-9

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): a fixture ledger holds one card for every edge of the stage rule, and the nine counts equal the open cards at every step.
  verify: `cargo test --offline --locked --lib statusline::state::stage_tests`
- AC-2 (REQ-2): the marker words of D-1 are pinned at their edges — each word alone, a first line only (a marker on line two is none), a note head cut at exactly 1000 characters, the newest marker winning over an older one of another kind, `RESUME` and `waiting:` and `LANE-HOLD` and `PAIR-` as no marker, `READY ` with its space against `READYX`.
  verify: `cargo test --offline --locked --lib statusline::state::marker_tests`
- AC-3 (REQ-5): the done-today count changes at the UTC midnight and not a second before.
  verify: `cargo test --offline --locked --lib statusline::state::done_today_tests`
- AC-4 (REQ-6): a session whose worktree matches a reservation shows that card even when its session id holds a newer one, a session without a worktree match shows the newest live reservation of its id, a card with an ended lease is never chosen, and a session with neither shows no group.
  verify: `cargo test --offline --locked --test statusline_flow my_card_`
- AC-5 (REQ-7): the flow line of every form in the Layouts section is exactly the string written there, plain and with color, the colored text minus its escapes equals the plain text, and a group of zero counts is left out.
  verify: `cargo test --offline --locked --lib statusline::render::flow_line_tests`
- AC-6 (REQ-8): the real binary writes the three lines in their roles at 120, 80 and 60 columns for a worktree session with a card, a session without a card and a generic project, and leaves a line out when it has nothing.
  verify: `cargo test --offline --locked --test statusline_flow three_lines_have_their_roles`
- AC-7 (REQ-9): at every form of every line the form changes exactly at the boundary — a budget equal to the form's width keeps it and one less drops to the next — with the padding subtracted, `COLUMNS` of `abc`, empty, `0` and absent giving the widest form, and the widths measured by `unicode-width`.
  verify: `cargo test --offline --locked --test statusline_cards width_ladder_`
- AC-8 (REQ-10, REQ-11): every icon of the icon table (D-6) is followed by the spacing its width class asks for, two spaces for a narrow one and one for a wide one, and a spent read budget gives the unsplit form with the ledger file unchanged.
  verify: `cargo test --offline --locked --test statusline_flow icon_spacing_and_read_budget`
- AC-9 (REQ-12): the document test finds the legend, the cited golden lines, the width rule and none of the stale sentences (the card counts on line 1, `Week` or `CTX` as a label, "shortens nothing", "no network").
  verify: `cargo test --offline --locked --test statusline_numbers_only the_statusline_doc_describes_the_three_lines`
- AC-10 (REQ-1, REQ-7, REQ-8, REQ-9, REQ-10): every statusline target passes with the goldens rewritten and every changed golden file read.
  verify: `cargo test --offline --locked --test statusline_cards --test statusline_golden --test statusline_tag_order --test statusline_worktree_label --test statusline_git_env --test statusline_counts_refresh --test statusline_home_state --test statusline_guard_false_positive --test statusline_numbers_only`

### Layouts

Every example is a fixture: `c12` is a made-up card number and the counts are the ones of the research (open cards at 2026-11-21T23:2xZ). Widths are display columns as the `unicode-width` crate of `Cargo.toml` counts them (measured 2026-11-22 with a probe built against the same crate; an emoji with its variation selector counts 2).

**The three lines at 120 columns (a worktree session holding a card, a subscriber).**

```
🧠 Opus 5.5 high·t | 📡 modu-ai/moai-adk | 🌳 statusline-flow ↑3 +2 | 🧭 next ↑3 ↓0 | 🔗 #12 pending
🎯 c12 W | 👀 S11 A2 L0 | ⛔ B3 ⏸️  H7 | 🗂️  D43 R150 W14 E22 ✅ 19
🪫 CW 74% | 🪫 5H 85% (1h 10m) | 🔋 7D 21% (6d 21h) | 🤖 2.1.330 | 🗿 4.2647.16
```

Widths 100, 67 and 79. At 80 columns line 1 drops to its 78-column form (`🧠 Opus high | 📡 moai-adk | 🌳 statusline-flow ↑3 +2 | 🧭 next ↑3 ↓0 | 🔗 #12`), line 2 stays whole and line 3 drops the versions (51). The forms of lines 1 and 3 are in `where` and `limits`.

**The flow line forms (this SPEC).**

| form | text | cols |
|---|---|---|
| P0 widest | `🎯 c12 W \| 👀 S11 A2 L0 \| ⛔ B3 ⏸️  H7 \| 🗂️  D43 R150 W14 E22 ✅ 19` | 67 |
| P1 no flow | `🎯 c12 W \| 👀 S11 A2 L0 \| ⛔ B3 ⏸️  H7` | 38 |
| P2 no hold | `🎯 c12 W \| 👀 S11 A2 L0 \| ⛔ B3` | 31 |
| P3 narrowest | `🎯 c12 W \| ⛔ B3` | 16 |
| Q0 no card | `👀 S11 A2 L0 \| ⛔ B3 ⏸️  H7 \| 🗂️  D43 R150 W14 E22 ✅ 19` | 56 |
| Q1 | `👀 S11 A2 L0 \| ⛔ B3 ⏸️  H7` | 27 |
| Q2 | `👀 S11 A2 L0 \| ⛔ B3` | 20 |
| Q3 narrowest | `⛔ B3` | 5 |

The forms drop groups from the right (flow, then hold, then waits), so the part the user asked for first stays longest. A session that holds a card whose stage is stale shows `🎯 c12 E`.

**Width rule.** `COLUMNS` and `LINES` are set by Claude Code to the terminal size before the command runs; the stdin JSON has no width field (docs read 2026-11-21T23:52Z; the page also says long output "may get truncated or wrap awkwardly" and notifications "may truncate your status line output"). The budget of a line is `COLUMNS − 2 − padding`, where `padding` is the `padding` value of `.moai/config/sections/statusline.yaml` (default 0; the generator writes the same value into the host's `statusLine.padding`). Each line picks its own form, so a narrow window shortens line 1 and line 3 without touching line 2. Read at the ladder, line 2 with a card: 120 and 80 columns give P0 (67 ≤ 78); 60 and 40 columns give P1 (67 > 58, and 38 ≤ 38 at 40); 36 columns gives P2 (31 ≤ 34); 30 columns gives P3 (31 > 28, the narrowest form is written anyway).

**The stage letters** (the legend the document carries): D draft · R ready · W working · E stale (a working card whose lease ended or that nobody holds) · S submitted (the lane says it waits for the leader or an audit; every `review` card too) · A audited (an audit passed for the head and nobody landed it) · L landed but not closed · B blocked (the lane asked for a decision) · H hold · ✅ n done today · `🎯 c<n> <letter>` my card and its stage.

### Supersedes

This SPEC changes what earlier SPECs pinned; they stay as the record of what was true, each cited by the requirement it changes.

| earlier SPEC | requirement | what changes |
|---|---|---|
| `SPEC-20261120-statusline-layout` | REQ-9 (same bytes at 40 and 80 columns for every line) | every line has width forms (REQ-9 here); the same bytes hold only when a form fits |
| `SPEC-20261120-statusline-layout` | REQ-2 (card counts `🗂️ ▶️ 🔃 ✅` open line 3) | the groups of REQ-6 and REQ-7 form line 2; line 3 holds the limits |
| `SPEC-20261119-statusline-waiting-cards` | REQ-1 to REQ-5 (four numbers; waiting markers) | stages and the widened rule (D-1); `review` stays waiting as stage S |
| `SPEC-20261121-statusline-detached` | REQ-5 (`CTX <n>%`, `Week <n>%`, `resets in`) | the labels are `CW` and `7D` (landed in step 1); the countdown loses `resets in` (`limits` D-4) |
| `SPEC-20261120-statusline-docs-items-2-3` | AC-1 to AC-4 (its `check.py` and the example lines) | the document and its example follow the steps of the Plan |

## Scope

- write: `.moai/specs/SPEC-20261122-statusline-flow/**` and the two child folders (this slice). The code slices of the Plan write `src/statusline/**`, `tests/statusline_*.rs`, `tests/fixtures/statusline-claude/**` (cases, stdin, ledgers, goldens), the statusline section of `docs/rust-product-path.md` and the SPEC folders named under Supersedes only through their Result lines.
- read-only now: `src/statusline/**`, `src/ledger/schema.rs`, the statusline SPECs, `tools/statusline_golden.py`.
- forbidden: the real HOME and ledger (the research's `sqlite3 -readonly` reads are measurements, not tests), the network in tests, any write to a ledger from the status line, a new crate.

## Plan

Step 1 (the `CW` and `7D` labels) landed. Each of the rest is its own landing with red tests first; the order puts what the user asked for first and moves the risky infrastructure behind it.

2. **Flow line.** `state.rs`: `CardCounts` becomes a `Flow` holding the nine counts, done today and my card with its stage; `marker_of` gets the widened words in the one existing function (no second parser); four statements (state counts and done today, notes of `active` and `review`, `card.landed` events, reservations). `render.rs`: the flow line as ordered groups. A shared width helper (`src/statusline/width.rs`: display width by `unicode-width`, the icon table with its spacing class, the form picker) that the other lines use later. No other line changes in this step except that the old card counts leave line 3.
3. **Line 1 and the cache** (`where`). Branch and SHA from `.git/HEAD`, the integration distance, the per-session cache, the tag and the lane registry gone.
4. **Line 3 and the host contract** (`limits`). `CW` as the host gives it, `5H` over 70, `💰`, versions, the usage record moved, the Orca pass-through, the generated settings and the tolerant input reader.
5. **Documents** (REQ-12): the three stale places and the example lines, in the same commit as the doc test's section marker. The old marker `shortens nothing` (`tests/statusline_numbers_only.rs:104`) ends the section the doc test reads and sibling tests read it too, so the new marker and the doc text change together and the sibling tests are run before the commit.
6. **Clean-up**: values nobody sees stop being collected (cache hit ratio, session time, output style, the goal file), config read from one path.

Before step 3 starts: print `COLUMNS` and the usable width once in a real session with a probe command and check the margin of two and the icon widths on this terminal (risk 2). Not done in this slice.

A step that finds this SPEC wrong writes a new revision before code.

## Risks

1. **The marker words are a first-line convention.** The master SPEC of the flow skeleton lists five first-line parsers that disagree (`src/statusline/state.rs:1148` is one) and plans typed event rows to replace them (its step 4; the leader's note said step 3, the master SPEC numbers admission and the limit as step 3); its AC-17 refuses a new parser. This SPEC adds words to the existing function only and says it is temporary: when typed submit, block and verdict rows exist, `marker_of` is replaced by a read of them. `RESUME` is not a marker, on evidence: of 245 notes with that first word the newest say "weekly usage limit reached" and "returned the lane", written when work stops; `waiting:` (110 notes) is a lane waiting on its own run, so it is working without a marker (counts read from `sqlite3 -readonly` at 2026-11-21T23:44:48Z over all notes).
2. **Width is a guess until observed.** The margin of two and the icon widths come from the host page and the crate, not from a live terminal. The crate counts an emoji with a variation selector as 2 columns, which is never less than what a terminal draws, so a line never ends up wider than counted; a terminal that draws such an icon in one column leaves the line a column shorter, which the two-space rule hides. The probe of the Plan measures once; if a terminal draws wider than counted, the margin is a constant to raise, not a redesign.
3. **Notifications share the line.** The host says notifications "may truncate your status line output" on narrow terminals, and the padding the user sets in the host's settings is not visible to the script. The margin of two and the `padding` option cover the defaults only.
4. **7D after the reset.** The host drops a window once its `resets_at` has passed, so `(0m)` never appears and an absent window is hidden (`limits` REQ-3); a user on a gateway with only a spend limit sees no `7D` and sees `💰` instead (`limits` REQ-5).
5. **Session id equals the holder, by code reading only.** `card take` writes the holder from `CLAUDE_CODE_SESSION_ID` (`src/card/command.rs:275`) and the statusline already uses the stdin `session_id` for the goal file; that the two are the same string in a live session was not observed. After `/clear` the id changes (the doc says it is unique per session), which is why REQ-6 matches the worktree first. Step 2 starts by checking both with the shipped binary; if they differ, REQ-6 keeps only the worktree match.
6. **Cost of the ledger read.** Four statements instead of two, about 33 ms measured on the real ledger against 115 ms for one statement; the 50 ms lock wait and the 800 ms budget stay. A timing measure on a copy of the real ledger goes into the Result of step 2, not into a test (a timing test would flake); the test counts statements instead.
7. **Letters need a legend.** A reader meeting `E22` cannot guess it. The legend is in the document and in the help of `moai statusline`; a test pins both against the letters the renderer writes.
8. **Golden churn.** Steps 2 to 4 change every golden; the golden harness is Python (`tools/statusline_golden.py`) and is used as it is until the Python removal ports it.

## Decisions

Reversible, category `none`, decider the leader (the user's picks are recorded in the Intent). A step finding a better value writes a revision before code.

| Id | Decision | Basis | Undo |
|---|---|---|---|
| D-1 | the widened marker rule: submitted = first line starts `LANE-READY-INTEGRATION`, `MERGE-READY`, `READY ` (word and one space), `감사 요청`, or `LANE-DONE` with `audit requested` or `감사 요청`; audited = `AUDIT` with `verdict: PASS`; blocked = `LANE-BLOCKED`; working = `ACK `, `PLAN `, or `AUDIT` with `verdict: FAIL`; everything else is no marker; the newest marker of a card wins; a card with no marker is working | notes 5921; the counts of risk 1; first-line prefixes read at 2026-11-21T23:44:48Z (`READY` 52, `MERGE-READY` 86, `LANE-BLOCKED` 325) | remove a word from the one function |
| D-2 | stale = a working card whose lease has ended or that has no live reservation; it is its own stage so `W` counts only cards someone holds; a submitted, audited or blocked card keeps its stage whatever its lease says | finding: 22 of 26 held reservations had ended leases | count stale as working again |
| D-3 | my card = a held reservation whose worktree is the session's worktree; else the newest live held reservation of the stdin session id; never the lowest number | note 5934 (measured: one session held an ended-lease card and two live ones) | order of the two sources |
| D-4 | done today = the UTC day of the render's clock | the ledger and every report use UTC; a rolling 24 hours is the alternative and avoids a reset at 09:00 in the user's zone | one expression |
| D-5 | the width budget of a line = `COLUMNS − 2 − padding`; each line picks its own form; absent or unreadable `COLUMNS` means the widest form | the host sets `COLUMNS` and has no stdin width; the user asked for 80 and 120 column forms | one constant |
| D-6 | icon spacing: one space after an icon that is wide by default, two after an icon that is narrow by default and written with a variation selector (`⏸️`, `🗂️`, `🅱️`, `⬇️`); the table of every icon with its class is in the code and pinned, and the wide-by-default replacements (`💤 📋 🌿 📥`) are listed in the table as the alternative, not built; width counted by `unicode-width` | note 5935; probe 2026-11-22: the crate gives the first character of each of those four width 1 and the whole sequence 2, and every other icon of the layouts width 2 | switch four icons and their spacing in one table |
| D-7 | no leader line, no leader tag and no lane registry read | note 5934 (the redesign removes the resident leader); revision 1 had a leader line and note 5930 asked for a one-function role read: with no leader line the function is not needed | a new revision when the ledger has a lander role |
| D-8 | `LINES` is not read; the status line stays three lines at most | the host decides what happens when the terminal has fewer rows; nothing in `LINES` changes what the three lines say | read it and drop line 3 first |
| D-9 | letters D R W E S A L B H and `✅ n`; English only; a legend in the document and the help | the user's rule: every icon keeps a letter or a number | rename a letter in one table |

## Result

이 개정은 SPEC 만이에요. 구현과 측정은 코드 조각에서 채워요. 돌린 것: 위 조사 표의 원장 읽기(읽기 전용), 레이아웃 너비 측정(`unicode-width` 로 만든 프로브, 2026-11-22), 아이콘 폭 표 측정, `spec check`. 돌리지 않은 것: 실제 Claude Code 세션에서의 `COLUMNS` 확인과 세션 id 대조(위험 2·5), 위 시험들(아직 없어요).
