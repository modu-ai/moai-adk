---
id: SPEC-20261120-statusline-layout
title: "Statusline three lines: card counts on line 1, repository with forge counts opening line 2, Claude Code version ending line 3"
status: draft
revision: 1
created: 2026-11-20
---

# Statusline three lines: card counts on line 1, repository with forge counts opening line 2, Claude Code version ending line 3

## Intent

사용자 지시(2026-11-20, 리더 대화)로 상태줄 세 줄의 배치를 사용자가 적은 예시로 바꿔요:

```
🏷️ Leader | 🧠 Opus 5.5 | 🤔 high·t | 🗂️ 90 ▶️ 19 🔃 18 ✅ 75 | 🗿 v4.2647.10
📡 modu-ai/moai-adk, ⬇️ 0 🔀 0 | 🅱️ v2026.11.0 +10
🔋 CTX: ██░░░░░░░░ 22% | 🔋 Week: ██████░░░░ 60% | 🤖 v2.1.325
```

지금(`src/statusline/render.rs` 의 `info_line`·`git_line`·`bars_line`, 카드 C-137 뒤 기준): 줄 1 은 태그·모델·effort·♻️ 캐시·🤖 Claude Code 버전·🗿 MoAI 버전, 줄 2 는 git 숫자·카드 숫자·저장소·가지·🔗 PR, 줄 3 은 워크트리·CTX·Week 이에요. 바꿈: (1) 카드 숫자 묶음은 줄 1 로 올라가 effort 와 🗿 사이에 서요. (2) 줄 2 는 저장소로 시작해 이름 바로 뒤 쉼표와 git 숫자가 한 구간에 붙어요(`📡 <repo>, ⬇️ n 🔀 n`), 그다음 가지, 그다음 PR 이에요. (3) 🤖 Claude Code 버전은 줄 3 끝으로 내려가요. (4) ♻️ 캐시 구간은 사용자 예시에 없어서 기본 배치에서 빠지고, 옵션 이름 `cache_hit` 은 그대로 받지만 아무것도 보여 주지 않아요(session_time·output_style 의 선례와 같아요). (5) 워크트리는 있을 때 줄 3 을 여는 그대로고, 아무것도 없는 줄은 여전히 빠져요, 원장을 읽을 수 없으면 카드 숫자는 그냥 안 나와요.

바뀌지 않는 것: 각 숫자의 뜻과 아이콘과 색, 칸 규칙(모든 아이콘 뒤 한 칸, 묶음 안 항목 사이 한 칸, 묶음 사이 ` | `, 카드 C-137 결정 D-1), 줄이지 않는 동작(`COLUMNS` 을 읽지 않음), 세그먼트 옵션 이름, 손끊김(handoff) 경고가 CTX 막대에 붙어 있는 것, 원장 숫자 규칙과 `🔃` 세 항목 꼴.

읽어서 확인한 사실(명령으로):
- 골든 도구 `tools/statusline_golden.py` 는 `tests/fixtures/statusline-claude/` 만 쓰고, 도구 머릿 주석 6행이 `tests/fixtures/statusline/` 를 "읽기 전용 보관본"이라 명시해요. 그래서 보관본은 이번에도 쓰지 않아요(D-4).
- 캐시 구간(♻️ 와 ⏱️·🧊)은 #2756 AC-8 의 사용자 선택(2026-11-13)으로 들어왔어요 — 그때는 원했지만 이번 예시에는 없어요. 카드 지시대로 이 사실을 리더 노트로 알려요(기본은 제거).
- C-137 의 골든 재생성 때 케이스 파일 4개(b08·b09·b13·b14)의 세 변형만 바뀌었고 `cases.json` 은 그대로였어요; 골든 파일 안 기록 행은 빌드한 바이너리 버전을 적어서, 워크트리 빌드로 재생성하면 팁과 다르게 됩니다(되돌려 놓는 절차를 Plan 4 에 둬요).

## Non-goals

- 숫자 값·아이콘·색·테마, 세그먼트 옵션 이름과 설정 문법 바꾸기.
- 원장 읽기 규칙(C-121 대기 규칙 포함)이나 stdin 해석 바꾸기.
- `tests/fixtures/statusline/` 옛 Go 기록과 `tools/statusline_golden.py`·`tools/claude-statusline.sh` 고치기.
- 40칸 줄임 같은 새 동작 만들기(지금 없어요).

## Requirements

- REQ-1: Line 1 shall read, joined by the line separator, the session tag, the model, the effort group, the card counts group and the MoAI version group in exactly that order.
- REQ-2: The card counts group shall keep its shapes on line 1: each number with its icon, one space between items, `🔃` only when the ledger notes are readable, zeros with a readable ledger, nothing without one.
- REQ-3: The renderer shall show nothing for cache data: with `cache_hit` accepted in a config the `♻️`, warm-timer and cold marks never appear on any line.
- REQ-4: The Claude Code version group `🤖 v<version>` shall stand at the end of line 3, after the worktree label and the usage bars.
- REQ-5: Line 2 shall open with one segment that starts `📡 <owner>/<name>` and carries a comma right after the repository name followed by the forge counts `⬇️ <issues> 🔀 <prs>` whenever the forge counts are shown, and the repository name alone when they are not, each part keeping its own color.
- REQ-6: The branch group `🅱️ <branch>` with its behind and dirty counts shall follow the repository segment, with the pull-request group `🔗 #<n> <state>` keeping the last place of line 2 and its hyperlink behavior.
- REQ-7: Every item that a line leaves out today shall stay left out for the same reasons together with its separator, and so stays out every line left with nothing in it.
- REQ-8: The worktree label shall keep opening line 3 when a worktree is present, and the usage bars keep their battery icons, gradient bars and handoff warnings.
- REQ-9: The renderer shall keep one space after every icon, one space between the items of a group, and ` | ` between the groups of a line.
- REQ-10: The renderer shall give the same bytes under `COLUMNS` 40 and `COLUMNS` 80 for every line.
- REQ-11: Every `tests/statusline*.rs` target and the python fixtures test shall pass with the unit tests, the integration pins and the golden files carrying the new layout.
- REQ-12: The statusline documentation shall describe the three lines in the new order and carry the golden example lines it cites verbatim.

## Acceptance

- AC-1 (REQ-1…REQ-10): a unit test builds a status with every segment present at the values of the user's example and the plain output is exactly those three lines; further unit tests pin each line's segment order by icon, the no-cache rule, the repository-alone form without forge counts, and the Claude Code version at the end of line 3.
  verify: `cargo test --locked --offline --lib statusline::render`
- AC-2 (REQ-2, REQ-5, REQ-7, REQ-11): the regenerated goldens agree with the binary; every changed golden file is read and each one changes only what this layout change touches; `cases.json` is unchanged.
  verify: `cargo test --locked --offline --test statusline_golden`
  manual: the diff of every changed file under `tests/fixtures/statusline-claude/golden/` is read, and provenance-only diffs are put back from the tip.
- AC-3 (REQ-2, REQ-3, REQ-4, REQ-7): without a ledger no card icon appears on line 1; with cache data present no `♻️` appears; in a main checkout line 3 opens with the CTX bar; without a git branch line 2 shows only the pull request or nothing and the output keeps two lines.
  verify: `cargo test --locked --offline --lib statusline::render --test statusline_cards`
- AC-4 (REQ-9, REQ-10): the full three-line output is byte-equal under `COLUMNS` 80 and 40, plain and truecolor, in the tests that pin widths today.
  verify: `cargo test --locked --offline --test statusline_cards --test statusline_tag_order`
- AC-5 (REQ-11): every statusline rendering target passes: `statusline_cards`, `statusline_golden`, `statusline_tag_order`, `statusline_worktree_label`, `statusline_git_env`, `statusline_counts_refresh`, `statusline_home_state`, `statusline_guard_false_positive`, and `python3 -m unittest tests.statusline_fixtures_test`.
  verify: `cargo test --locked --offline --test statusline_cards --test statusline_golden --test statusline_tag_order --test statusline_worktree_label --test statusline_git_env --test statusline_counts_refresh --test statusline_home_state --test statusline_guard_false_positive`
- AC-6 (REQ-12): the documentation items for the three statusline lines state the new order and spacing, name `cache_hit` as accepted but silent, and cite the golden example lines verbatim.
  verify: `python3 -I -c` check that the cited golden lines appear in the document and the removed shapes do not.

## Scope

- write: `src/statusline/render.rs`, `tests/statusline_*.rs`, `tests/statusline_fixtures_test.py`, `tests/fixtures/statusline-claude/**` (`golden/**` and, only if the writer touches it, `cases.json`), `docs/rust-product-path.md` (statusline items 1–3 only), `.moai/specs/SPEC-20261120-statusline-layout/**`
- read-only: `tests/fixtures/statusline/**` (historical Go archive), `tools/statusline_golden.py`, `tools/claude-statusline.sh`, `src/statusline/{state,segments,counts,config,input,mod,color,godecode}.rs` except where a compile demands a one-line touch, which is then named in the Result
- forbidden: the real HOME and ledger, the network, a change of the counts, their markers or the option names

## Plan

1. SPEC first (this file); link it to the card; the cache finding goes to the leader as a note (D-1 basis).
2. RED: rewrite the whole-line pins of the renderer unit tests and `tests/statusline_cards.rs` to the new shapes; they fail on the current code.
3. GREEN: `info_line` drops the cache segment and the Claude Code version and gains the card counts group; `git_line` merges the repository and the forge counts into one segment with the comma and drops the card counts; `bars_line` gains the Claude Code version at the end; the doc comment of `render()` states the new three lines.
4. Regenerate the goldens with `python3 -I tools/statusline_golden.py --bin <fresh binary> --write`; read the diff of every changed file; put provenance-only changes back from the tip.
5. Run every `tests/statusline*.rs` target, the renderer unit tests, the python fixtures test, then `--lib` whole, `--bins`, fmt, ops rustfmt, clippy, the scan tests, `spec check`, trailer-check; mutants (at most five) after committing.
6. READY note with every exit code; audit by the leader (tier 1).

## Risks

- A golden case is three files; a regenerated file that changes something besides the layout lines is a bug of this change, so every diff is read.
- The provenance line of a regenerated golden names the building binary, so worktree builds produce a version diff that is not this change; those lines are put back from the tip (C-137 saw the same).
- The comma joins two colored parts inside one segment; with the forge counts absent it must not appear, and under every color profile the bytes stay the plain bytes plus escapes.
- Removing the cache display touches a test (`cache_timer_worktree_and_pr_segments`) that pins the warm and cold marks; it now pins that they never appear, which loses a pin rather than gaining one.
- Cases whose goldens should not change (no ledger, no versions, forge-less) exist; a diff there is a bug.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `cache_hit` stays accepted in configs and shows nothing; cache state and its parser are untouched | the user's layout has no `♻️`; the card says the option name stays accepted, like `session_time` and `output_style`; the 2026-11-13 user choice that added it is reported to the leader by note | one `segs.push` restores the segment |
| D-2 | the comma and the forge counts join the repository segment only when the forge counts show today (github on, available, not suppressed); otherwise the name stands alone | the forge counts keep their own availability rules; the example shows both together | none needed |
| D-3 | the Claude Code version keeps its `claude_version` gate and stands after the usage bars on line 3 | the user's example ends line 3 with `🤖 v2.1.325`; a line with only the version is still one non-empty line | move the push back to `info_line` |
| D-4 | `tests/fixtures/statusline/` stays a read-only archive even though the card names the "statusline" goldens | the golden tool's own header says so and `--write` touches only `statusline-claude` (read at the tool, line 6) | none |

## Result

(작성 예정 — 구현과 검사 뒤)
