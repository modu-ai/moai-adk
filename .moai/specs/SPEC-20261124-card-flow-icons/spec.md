---
id: SPEC-20261124-card-flow-icons
title: "One card-flow icon set across the status line, the progress board and the lane reports"
status: draft
revision: 1
created: 2026-11-24
---

# 카드 흐름 아이콘 한 묶음을 상태줄, 진행 보드, 레인 보고가 같이 써요

## Intent

사용자가 골랐어요(2026-11-24T10:49:29Z, 카드 C-466). 상태줄 `🗂️ 109 → 🟡 1 → 📨 2 → 🛬 2 → ✅ 55 | 🚧 6 🙋 96` 을 보고 "이모지를 좀더 직관적으로 바꾸자"고 했고, 선택지 넷 가운데 "일의 단계 비유"를 골랐어요. 한 장의 카드가 걷는 길과 그 옆의 상태는 이렇게 읽혀요.

| 키 | 아이콘 | 뜻 | 옛 아이콘 |
|---|---|---|---|
| todo | 📋 | 할 일(차례 대기) | 🗂️ |
| work | 🔨 | 작업 중(레인이 진행) | 🟡 |
| check | 👀 | 검사 중(READY 뒤 감사) | 📨 |
| ready | 🚀 | 출발 대기(감사 통과, 착지 대기) | 🛬 |
| done | ✅ | 완료 | ✅ |
| blocked | ⛔ | 막힘 | 🚧 |
| paused | ⏸️ | 보류(철회·합쳐짐) | (🙋 에 섞여 있었어요) |
| question | 🙋 | 사용자 답이 필요할 때만 | 🙋 |

지금은 같은 묶음이 세 곳에 따로 적혀 있어요. 상태줄 코드(`src/statusline/state.rs` 의 `Stage::icon`, `render.rs` 의 `flow_tails`, `width.rs` 의 `ICONS`), 출력 스타일(`skills/moai/references/home/output-style.md` 와 템플릿 셋)과 그 시험(`tests/style_banner_scan.rs`, `tests/home_style_ownership_cli.rs`)이에요. 레인 보고는 그 출력 스타일의 진행 보드를 그대로 써요. 이 SPEC 은 묶음을 한 표에 두고 셋이 그 표를 읽게 해서, 아이콘이 다시 바뀌어도 한 곳만 고치게 해요.

이 카드가 C-452(상태줄 수 고침)의 감사 후속도 받아요(N1, C-452 의 감사 노트 8519 와 리더 승인 노트 8518). C-452 는 수만 고쳤고 아이콘은 이 카드의 몫이에요. 그래서 `tests/statusline_counts_truth.rs` 의 아이콘을 새 묶음으로 바꾸고, `⏸️` 가 철회·합친 보류까지 세도록 `Flow` 에 필드를 더하고, C-452 SPEC 의 Supersedes 표에서 빠진 줄을 채워요.

### 읽은 것 (2026-11-24, NeXT `ff396855d9`)

1. 흐름 줄은 `render.rs` 의 `flow_tails` 가 만들어요. 항목 5개(`🗂️ 🟡 📨 🛬 ✅`)와 옆 항목 2개(`🚧 🙋`)가 문자열 상수로 박혀 있고, 내 카드의 `🎯 C-<n> <단계 아이콘>` 은 `Stage::icon`(`state.rs`)이 정해요. 폭 맞추기는 왼쪽 항목부터 빼요(`🗂️` 먼저, `✅` 마지막) — 이 규칙은 그대로 둬요.
2. 철회·합친 hold 카드는 C-452 가 통계문 1 에서 `withdrawn` 으로 읽어요(`FLOW_STATEMENTS[0]`). 지금은 세지 않고 건너뛰어요(`state.rs` 의 `"withdrawn" => continue`). ⏸️ 는 그 수를 세면 돼요.
3. `width::ICONS` 가 줄이 그릴 수 있는 아이콘 22개를 나열하고, `⬇️`·`🗂️` 같은 변이 선택자(U+FE0F) 아이콘 뒤에도 정확히 한 칸을 쓰는 시험이 그 표를 읽어요. `⏸️`(U+23F8 U+FE0F)도 같은 모양이라서 같은 규칙이에요.
4. 출력 스타일의 진행 보드는 열한 아이콘(`🙋 🔴 ✅ 🛬 🔍 📨 🟡 🔁 🚧 🗂️ ❌`)이에요. 템플릿 셋(`moai.md`, `moai-easy.md`, `moai-learn.md`)은 흐름 한 문장과 옆 항목 한 문장을 따로 적어요. 지금 시험은 `👀 ⛔ ⏸` 을 "묶음에서 빠진 아이콘"으로 막아요(`tests/style_banner_scan.rs` 의 `RETIRED_ICONS`, `tests/home_style_ownership_cli.rs` 의 `gone`) — 새 묶음이 셋을 되살리므로 그 목록이 바뀌어야 해요.
5. 출력 스타일의 렌더는 sha256 으로 고정돼 있어요(`CURRENT_RENDER_SHA256`, `EARLIER_RENDERS`, 이전 렌더는 `tests/fixtures/home-style/render-N.md`). 스타일 글이 바뀌면 지금 렌더가 `render-7.md` 로 들어가고 새 렌더가 고정돼요. 그래야 이미 설치된 사용자 스타일 파일이 충돌이 아니라 갱신으로 계획돼요(`SPEC-20261123-progress-board-flow-icons` REQ-2 와 같은 일이에요).
6. 아이콘이 나오는 골든은 `tests/fixtures/statusline-claude/golden/*` 이고, 흐름 줄이 있는 케이스만 바뀌어요. 한도 줄(`🔋 🪫`)과 첫 줄 아이콘은 그대로예요.

### 리더에게 묻는 것 (아래 Decisions, 기본값이 ★)

D-3(진행 보드에 남는 두 아이콘), D-4(⏸️ 의 수와 모양), D-7(C-452 SPEC 의 Supersedes 표를 어떻게 고칠지)이에요. 나머지는 카드와 사용자 결정을 그대로 옮긴 것이에요.

## Non-goals

- 한도 줄(`CW`·`5H`·`7D`, `🔋 🪫`), 첫 줄의 아이콘, 흐름 줄의 폭 규칙과 항목 순서의 변경.
- 카드 상태(원장의 `ready`, `active`, `review`, `hold`, `done`)나 `moai card` 명령, 단계 판정 규칙(`marker_of`, 임대, 철회 판정)의 변경 — 어느 카드가 어느 칸에 세어지는지는 그대로예요.
- 새 카드 상태나 새 숫자의 추가. `⏸️` 의 수는 C-452 가 이미 읽는 `withdrawn` 을 세는 것뿐이에요.
- 레인 보고 모양의 새 서식. 레인 보고는 진행 보드를 쓰므로 보드가 바뀌면 같이 바뀌어요.
- 옛 SPEC 본문을 고쳐 쓰는 일. 옛 SPEC 은 그때의 기록으로 남고 이 SPEC 의 Supersedes 표가 무엇이 바뀌었는지 적어요.

## Requirements

- REQ-1: The icon table `moai_cli::statusline::icons` shall hold the card-flow icon set once, as five flow steps `todo 📋`, `work 🔨`, `check 👀`, `ready 🚀`, `done ✅` and three beside-the-flow icons `blocked ⛔`, `paused ⏸️`, `question 🙋`, each with its key and its board meaning · failure: none — a second spelling of an icon in the status line code is found by the scan of AC-2
- REQ-2: When the flow line is drawn, the status line shall write the five flow items in the order `📋 <ready> → 🔨 <working plus stale> → 👀 <submitted> → 🚀 <audited plus landed> → ✅ <done today>` · failure: none — an item whose count is zero is left out with its arrow, as before
- REQ-3: When the flow line is drawn, the status line shall write the beside items after the flow, a bar apart, as `⛔ <blocked> ⏸️ <paused> 🙋 <question>` · failure: none — an item whose count is zero is left out, so `🙋` shows only when a user question exists
- REQ-4: The status line shall count as `paused` the hold cards that have a withdraw or fold event, and as `question` the hold cards that have none · failure: none — a card is in exactly one of the two, and a `paused` card is in no flow stage, as the withdrawn cards are today
- REQ-5: The status line shall draw the icon of the card in hand from the table: `📋` for a draft or ready card, `🔨` for a working or stale one, `👀` for a submitted one, `🚀` for an audited or landed one, `⛔` for a blocked one and `🙋` for a hold one · failure: none — every stage has an icon
- REQ-6: When the width is short, the status line shall drop the leftmost flow item first and keep the beside items and the card in hand to the end · failure: none — a ledger whose notes cannot be read gives the one form `📋 <ready> 🔨 <working>` without an arrow, never a guess
- REQ-7: The status line shall write exactly one space after every icon of the table, `⏸️` included · failure: none — the display width of `⏸️` is counted as the whole sequence
- REQ-8: The icon list `width::ICONS` shall list every icon of the table · failure: none — an icon of the table missing from the list fails the table test
- REQ-9: The progress board section of the shipped output style and the three shipped style templates shall spell the flow as one sentence and the beside icons as one sentence, both built from the table's meanings, with the sample lines of the board in the order `🙋 🔴 ✅ 🚀 👀 🔨 ⛔ ⏸️ 📋 ❌` · failure: none — a text that spells a sentence otherwise fails the scan of AC-3
- REQ-10: The shipped text of the status line documents and of the styles shall hold none of the retired icons `🗂️ 📨 🛬 🚧 🟡 🔍 🔁` · failure: none — the old icon text in a golden, a template, a document or a doc comment is found by the scan of AC-3
- REQ-11: When the shipped style text changes, the earlier renders list shall hold the outgoing render with the fixture `tests/fixtures/home-style/render-7.md` · failure: none — a home that holds the outgoing render plans as an update, never as a conflict
- REQ-12: The tests of the status line and of the style templates shall read their icons from the table of REQ-1 · failure: none — a test that spells an icon of the set by hand is found by review at the audit

## Acceptance

- AC-1 (REQ-1, REQ-7, REQ-8): the table holds eight different icons with their keys and meanings, all of them in `width::ICONS`, none of them one of the retired icons, and one space follows each in every drawn form · verify: `cargo nextest run --locked --offline --lib -E 'test(statusline::icons) | test(statusline::width)'`
- AC-2 (REQ-2, REQ-3, REQ-4, REQ-5, REQ-6): a ledger fixture with one card in every stage (and withdrawn, folded and plain hold cards) draws the line `📋 <n> → 🔨 <n> → 👀 <n> → 🚀 <n> → ✅ <n> | ⛔ <n> ⏸️ <n> 🙋 <n>` at the wide width, `🙋` is gone when no plain hold card exists, the items drop in the order of REQ-6 at each width boundary, my card shows the icon of its stage for all nine stages, and the 07:32Z fixture of C-452 draws its sums with the new icons (the sums stay the same) · verify: `cargo nextest run --locked --offline --test statusline_card_flow_icons --test statusline_counts_truth --test statusline_flow`
- AC-3 (REQ-9, REQ-10): the board section holds the ten sample lines in the order of REQ-9, the flow sentence and the beside sentence equal the strings built from the table, the three templates hold the same two sentences, and no retired icon is in the shipped styles, the status line documents, the agents, the skill references or the plain goldens · verify: `cargo nextest run --locked --offline --test style_banner_scan --test home_style_ownership_cli --test statusline_card_flow_icons`
- AC-4 (REQ-11): the earlier renders list holds the outgoing render with a `render-7.md` fixture, the new render's sha256 is pinned, a home that holds the outgoing render is planned as a `merge`, and an edited style file is still a `conflict` · verify: `cargo nextest run --locked --offline --test home_style_ownership_cli`
- AC-5 (REQ-2, REQ-3): the goldens for the flow line are rewritten with the new icons and read line by line, every other golden stays byte for byte, and the checksum line of `cases.json` follows the binary · verify: `cargo nextest run --locked --offline --test statusline_golden`
- AC-6 (REQ-12): the status line tests and the style tests build their expected icons from the table (a changed icon in the table fails no test but the table test itself) · verify: `cargo nextest run --locked --offline --test statusline_card_flow_icons -E 'test(the_tests_read_the_icons_from_the_table)'`
- AC-7: the Supersedes table of `SPEC-20261124-statusline-counts-truth` names the flow REQ-1 and AC-1, the bar-color REQ-1, the limits-left AC-7, the order-bars AC-4 and the ready-receipt-followup line it had wrong, and the SPEC check passes for both SPECs · verify: `moai spec check SPEC-20261124-card-flow-icons --project .` and `moai spec check SPEC-20261124-statusline-counts-truth --project .`

## Scope

- write: `.moai/specs/SPEC-20261124-card-flow-icons/**`, `src/statusline/icons.rs` (new), `src/statusline/mod.rs` (the one module line for icons), `src/statusline/state.rs`, `src/statusline/render.rs`, `src/statusline/width.rs`, `src/home_style.rs` (the earlier renders list only), `skills/moai/references/home` (its `output-style.md`), `skills/moai/references/source-manifest.json` (the sha256 of the style source), `templates/claude/output-styles/moai.md`, `templates/claude/output-styles/moai-easy.md`, `templates/claude/output-styles/moai-learn.md`
- write, tests: `tests/statusline_card_flow_icons.rs` (new), `tests/statusline_counts_truth.rs`, `tests/statusline_flow.rs`, `tests/statusline_cards.rs`, `tests/style_banner_scan.rs`, `tests/home_style_ownership_cli.rs`, `tests/fixtures/home-style/render-7.md` (new), `tests/fixtures/statusline-claude/golden/**`, `tests/fixtures/statusline-claude/cases.json` (the checksum line the golden tool writes), `tools/repo_scans.json` (the entry of the new scan test), and the files the render digest reaches that the test run names (`tests/fixtures/update_tux/*.json`, `tests/e2e_temp_home_cli.rs`, the source manifest), each added by a scope request when a run shows it
- write, the superseded lines: `.moai/specs/SPEC-20261124-statusline-counts-truth/spec.md` (the Supersedes table only, D-7) and the status line documents `docs/rust-product-path.md` (the flow-line item, by line lease: card C-449 holds the file)
- read-only: `src/ledger/**`, `src/card/**`, `tools/statusline_golden.py`
- forbidden: the real HOME and ledger, any write to a ledger from the status line, the limit-row code and goldens' limit lines, a new crate, `src/lib.rs`

## Plan

1. Red first. `tests/statusline_card_flow_icons.rs` holds the AC-2 ledger fixture, the AC-3 scans and the AC-6 test; the rewritten expectations of the existing tests fail on the current head because the line still draws `🗂️ 🟡 📨 🛬`. The table test of AC-1 fails because the table does not exist.
2. `src/statusline/icons.rs`: the table (`CARD_FLOW`, `BESIDE`, the board meanings) and `pub mod icons;` in `mod.rs`. `Stage::icon` and `flow_tails` read it; `Flow` gets `paused`, and statement 1's `withdrawn` branch counts it; `render.rs` writes `⏸️` and keeps `🙋` for the plain hold cards. `width::ICONS` lists the eight new icons and drops the retired ones; the doc comments that name the old icons follow.
3. The style text: the board section, the flow and beside sentences and the three templates, built from the table's meanings; `render-7.md` is a copy of the render before the change, `EARLIER_RENDERS` and `CURRENT_RENDER_SHA256` follow; the fixtures and manifests that carry the digest follow as the test run names them.
4. Goldens: `python3 -I tools/statusline_golden.py --write`, then `--check`, and every changed golden file is read; the flow-line icons are the only change in the plain goldens.
5. The Supersedes table of `SPEC-20261124-statusline-counts-truth` (D-7) and the flow-line item of `docs/rust-product-path.md` (by line lease).
6. guard-mutant per new guard test (a hand-spelled old icon in `flow_tails`, `paused` counted into `question`, a retired icon in a template), then `moai check affected` once, bare, for the READY note.

## Risks

1. **The golden churn is every flow line.** Only the icons and the `⏸️` item change; the sums do not. A golden that changes in any other way means the code changed more than the icons, and the audit reads the changed lines.
2. **`⏸️` is narrow by default with a variation selector.** Terminals differ in how wide they draw it. The width counts the whole sequence and one space follows it (REQ-7); a terminal that draws it one cell wide would show one cell of extra space, never a collision.
3. **`⏸️ <paused>` is a large number today.** The ledger holds ninety-odd withdrawn cards, so the line shows `⏸️ 96` where `🙋 96` was; the user asked for exactly that split (card C-466). The width forms keep it to the end like `⛔` (REQ-6), so a narrow terminal loses the flow items before it loses the pause count.
4. **The render digest reaches several fixtures.** A shipped style change moves the digest in the update fixtures and the e2e test; the scope request names each file the run shows, not a guess.
5. **The earlier retirement is reversed.** `SPEC-20261122-statusline-limits-left` REQ-8 and `SPEC-20261123-progress-board-flow-icons` retired `👀 ⛔ ⏸` and tests pin that; this SPEC brings them back as the user's pick, so those tests change on purpose (Supersedes).

## Decisions

Reversible, category `none`; the leader's delegation of 2026-11-23T13:57Z covers decisions of this kind. D-1, D-2 and D-5 are the user's pick and the card's text; D-3, D-4, D-6, D-7 and D-8 are this SPEC's reading and wait for the leader's check.

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | the set is 📋 🔨 👀 🚀 ✅ and ⛔ ⏸️ 🙋 with the meanings of the Intent table | the user's pick, 2026-11-24T10:49:29Z, card C-466 | change the table (one file) and the pinned render |
| D-2 | the table lives in `src/statusline/icons.rs` as `pub mod icons;` in `src/statusline/mod.rs`, not in `src/lib.rs` | `src/lib.rs` is a shared file held by line lease; `statusline` is already a public module the tests import (`moai_cli::statusline::width::ICONS`) | move the module |
| D-3 | ★ the board keeps `🔴` risk and `❌` failed or cancelled beside the set, and drops `🔁` repair and `🔍` in check (a failed check sends the card back to `🔨`; the audit is `👀`). Board order `🙋 🔴 ✅ 🚀 👀 🔨 ⛔ ⏸️ 📋 ❌`. Alternative: keep `🔁` too | `🔴` and `❌` are not card states and the user's pick did not replace them; `🔁` and `🔍` have a card-flow twin in the pick | add the line back to the table |
| D-4 | ★ `⏸️` counts the hold cards with a `card.withdraw` or `card.fold` event; `🙋` counts the hold cards with none; the order is `⛔ ⏸️ 🙋`; both stay to the narrowest form like `⛔`. Alternative: `⏸️` drops before the flow items at a narrow width | card C-466 criterion 4 and the ledger read of C-452 (REQ-8 there); one rule for the beside items | change `flow_tails` |
| D-5 | the card in hand shows `📋` (draft, ready), `🔨` (working, stale), `👀` (submitted), `🚀` (audited, landed), `⛔` (blocked), `🙋` (hold) | the same mapping as the flow items; hold in hand is a card that waits for the user's pick | change `Stage::icon` |
| D-6 | `⏸️` is written with U+FE0F, like `🗂️` before it | the user's pick showed `⏸` without the selector; it is written with U+FE0F only for the width count (emoji presentation, two cells), one rule for every narrow-by-default icon (REQ-7), approved by the leader in note 8537 | drop the selector and the width entry follows |
| D-7 | ★ the Supersedes table of `SPEC-20261124-statusline-counts-truth` is fixed in place as revision 3 (the table rows only), and this SPEC's own Supersedes table below carries the icon changes. The leader approves two digests. Alternative: leave the C-452 table as it is and put every missing row in this SPEC's table | the audit of C-452 (note 8519, N1; the leader's answer is note 8518) asked for the rows, and card C-466 criterion 4 names the C-452 table | `git revert` of the one commit |
| D-8 | the outgoing render becomes `render-7.md` and the new render is pinned in the same change | `SPEC-20261123-progress-board-flow-icons` REQ-2, the installer's ownership rule | n/a, forced by the installer test |

### Supersedes

This SPEC changes what earlier SPECs pinned; they stay as the record of what was true.

| earlier SPEC | requirement | what changes |
|---|---|---|
| `SPEC-20261122-statusline-limits-left` | REQ-6, AC-7, AC-8 (the flow line `🗂️ 🟡 📨 🛬 ✅ \| 🚧 🙋`, the width order) | the flow line is `📋 🔨 👀 🚀 ✅ \| ⛔ ⏸️ 🙋` in the same order and with the same width rule (REQ-2, REQ-3, REQ-6) |
| `SPEC-20261122-statusline-limits-left` | REQ-8 and the board order `🙋 🔴 ✅ 🛬 🔍 📨 🟡 🔁 🚧 🗂️ ❌` | the board order is `🙋 🔴 ✅ 🚀 👀 🔨 ⛔ ⏸️ 📋 ❌` (REQ-9, D-3) |
| `SPEC-20261123-progress-board-flow-icons` | REQ-3 (the eleven icons in flow order) | ten icons, the flow of REQ-9 |
| `SPEC-20261122-statusline-flow` | REQ-1 (nine stages), REQ-11 (the unsplit form `🗂️ D<n> R<n>`) | the stages are the same, the icons are the table's; the unsplit form is `📋 <ready> 🔨 <working>` (REQ-6) |
| `SPEC-20261124-statusline-counts-truth` | REQ-7 and REQ-8 (the `🟡` and `🙋` items) and AC-1, AC-3 (their lines) | the items are `🔨` and `⏸️`/`🙋`; the sums are the same (REQ-4) |

## Result

구현했어요(2026-11-24, 브랜치 `spec/card-flow-icons-20261124`, 코드·시험·골든·문서 커밋 `843a686f0b`). SPEC 개정 1 은 리더가 승인했어요(digest `4ddeecd9`, 선택 (가)(나)(다)는 모두 기본값, 범위는 노트 8534). 독립 감사는 이 레인이 하지 않아요(리더가 다른 세션에 맡겨요).

- red 증거: 구현을 먼저 한 실수가 있어서, 구현 diff 를 `target/card-flow-icons/impl.patch` 로 저장하고 `state.rs`·`render.rs` 만 커밋 전(HEAD) 판으로 되돌린 채 새 통합 시험 `statusline_card_flow_icons` 를 돌렸어요. 5개 중 4개 실패(exit 100, `target/card-flow-icons/red.log`)했고 표를 읽는 스캔만 통과했어요. 그 뒤 파일을 되돌리고 `cmp` 로 복원을 확인했어요.
- 시험 결과(작업 트리, 커밋 `843a686f0b` + 시험 한 줄 수정): `statusline`·`home_style` lib 시험과 통합 시험 8개 바이너리를 한 슬롯에서 돌려 339개 중 338 통과, 1 실패(`every_stage_draws_its_item_of_the_table` 의 기대 숫자 `🔨 3` 이 내 실수, 막힌 카드는 🔨 에 들지 않아요). 기대를 `🔨 2` 로 고친 뒤 `statusline_card_flow_icons` 5개 모두 통과했어요. 앞서 두 번의 "통과" 보고(244, 245개)는 `-E` 이름 필터가 통합 시험 바이너리를 빼서 lib 시험만 돈 것이라 취소해요.
- 골든: `tools/statusline_golden.py --write` 후 `--check` 로 114개 통과. 바뀐 파일은 `b13-card-counts`, `b14-card-counts-waiting` 의 plain·ansi·ansi256 여섯 개뿐이고, 옛 파일에 아이콘 치환(📋→🗂️ 🔨→🟡 👀→📨 🚀→🛬 ⛔→🚧)을 되돌려 적용하면 바이트까지 같은 것을 스크립트로 확인했어요(차이 0). plain 의 두 줄은 `📋 25 → 🔨 3 → 👀 2 → ✅ 12 | 🙋 5`, `📋 25 → 🔨 2 → 👀 3 → ✅ 12 | 🙋 5` 이에요. `cases.json` 은 바이너리 sha256 줄만 바뀌었어요. 이 골든 장부에는 철회된 hold 카드가 없어서 `⏸️` 는 골든이 아니라 통합 시험이 고정해요.
- 뮤턴트(레인이 직접, 커밋 뒤 하나씩 적용하고 `git checkout` 으로 되돌림, 같은 슬롯): M1 `Flow.paused` 를 올리지 않음 → exit 100(`a_withdrawn_hold_card_is_not_counted_as_waiting_for_the_user` 와 통합 시험), M2 `⏸️` 항목을 그리지 않음 → exit 100(`a_zero_count_hides_its_item` 등), M3 `PAUSED` 에서 U+FE0F 제거 → exit 100(`icons::tests` 둘, `width::tests`, `home_style_ownership_cli`), M4 템플릿 `moai.md` 에 `🚧` 추가 → exit 100(`no_shipped_text_holds_an_icon_of_the_earlier_card_flow`), M5 `EARLIER_RENDERS` 의 새 sha 한 자리 변경 → exit 100(`the_earlier_renders_are_pinned_and_never_hold_the_current_one`). 처음 돌린 M4·M5 는 이름 필터 탓에 통합 시험이 안 돌아 exit 0 으로 살아남았고, 필터를 고쳐 다시 돌려 잡았어요. 끝난 뒤 추적 파일은 이 SPEC 의 `spec.md` 만 바뀐 상태(내 편집)였어요.
- 문서: `docs/rust-product-path.md` 는 카드 C-449 가 예약 중이라 `moai card line-lease` 로 흐름 줄 항목(첫 줄 설명, 예시 한 줄, 아이콘 문장, 좁은 폭 문장, 비분할 모양)만 고쳤어요. 같은 파일의 옛 개수 묶음 설명(`🗂️ ▶️ 🔃 ✅`)은 흐름 줄이 대신한다고 적힌 옛 항목이라 그대로 뒀어요.
- `tools/repo_scans.json`: 처음에 새 시험 파일이 트리를 걷는다고 `repo_scans_table` 이 막았어요. 이 파일은 C-449 가 예약 중이고 내 예약 범위에도 없어서(잘못 한 번 쓴 것을 `git checkout` 으로 되돌렸어요, line-lease 는 C-449 때문에 거절) 쓰지 않고, 걷는 검사를 이미 등록된 `tests/style_banner_scan.rs`(범위 `templates/`, `skills/moai/`)로 옮기고 `statusline_card_flow_icons.rs` 는 이름으로 고른 파일만 읽게 바꿨어요. 이 SPEC 의 Scope 에 있던 `tools/repo_scans.json` 줄은 쓰지 않았어요.
- 해석: AC-3 의 "status line documents" 는 범위의 `docs/rust-product-path.md` 흐름 줄 항목까지만 고쳤고 스캔에는 넣지 않았어요(위의 옛 항목에 `🗂️` 가 남아 있어요). `tests/statusline_worktree_label.rs:165` 의 옛 아이콘은 범위 밖이라 그대로예요(리더가 요청하면 한 줄 요청으로 다뤄요).

- AC-1: met — `statusline::icons::tests` 와 `statusline::width::tests` 가 표의 아이콘 8개(+ 보드의 🔴 ❌)가 서로 다르고 모두 `width::ICONS` 에 있고 옛 아이콘이 없음을 고정해요. `⏸️` 의 폭도 한 곳에서 세요.
- AC-2: met — `tests/statusline_card_flow_icons.rs` 의 `every_stage_draws_its_item_of_the_table`(`📋 3 → 🔨 2 → 👀 1 → 🚀 1 → ✅ 2 | ⛔ 1 ⏸️ 2 🙋 2`), `the_question_item_shows_only_when_a_user_pick_is_owed`, `a_narrow_width_drops_the_road_and_keeps_what_stands_beside_it` 가 고정해요. 막힌 폭의 단계별 모양은 `flow_line_tests` 의 `every_form_is_exactly_its_string_and_changes_at_its_boundary` 가 고정해요.
- AC-3: met — `home_style_ownership_cli` 의 진행 보드 시험, `style_banner_scan` 의 두 시험(`the_three_style_templates_use_the_one_icon_set`, `no_shipped_text_holds_an_icon_of_the_earlier_card_flow`), `statusline_card_flow_icons` 의 plain 골든 시험이 고정해요(문서 항목은 스캔 밖, 위 해석 참고).
- AC-4: met — `home_style_ownership_cli` 의 `the_earlier_renders_are_pinned_and_never_hold_the_current_one`, `a_style_file_equal_to_an_earlier_shipped_text_is_updated`(M5 가 잡아요)와 `render-7.md`.
- AC-5: met — 위 골든 항목.
- AC-6: met — `statusline_card_flow_icons` 의 `the_tests_read_the_icons_from_the_table` 가 스캔하는 여섯 시험 파일에 새 아이콘 리터럴이 없어요(주석 줄은 제외).
- AC-7: met — `moai spec check` 가 두 SPEC 모두 오류 0(counts-truth 는 REQ 14개 경고 하나로 이미 있던 것), Supersedes 표에 flow REQ-1·AC-1·AC-3, order-bars AC-4, limits-left AC-7, bar-color REQ-1 줄을 더하고 ready-receipt-followup 줄을 고쳤어요. counts-truth 의 digest 는 개정 2 와 같은 `sha256:d92ac237c83882828ef28a17dec307f7b28a86b0fccf4a1ff8025f4de3dffcd6` 이에요(Supersedes 표는 digest 범위 밖).
