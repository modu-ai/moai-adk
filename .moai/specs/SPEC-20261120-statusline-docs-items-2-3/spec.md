---
id: SPEC-20261120-statusline-docs-items-2-3
title: "Statusline documentation: items 2 and 3 match the card counts and the worktree label"
status: draft
revision: 1
created: 2026-11-20
---

# Statusline documentation: items 2 and 3 match the card counts and the worktree label

## Intent

`docs/rust-product-path.md` 의 상태줄 설명 중 2번(저장소 줄)과 3번(사용량 줄)이 옛 모양을 말해요. 코드와 골든은 이미 바뀌었어요.

- 카드 숫자: 문서는 `🔥 N / 📥 M / ✅ D` 세 숫자(진행 = `active` + `review`)라고 해요. 코드는 `🗂️ 남은 / ▶️ 작업 중 / 🔃 통합 대기 / ✅ 완료` 네 숫자예요(C-121, SPEC-20261119-statusline-waiting-cards 의 D-5 가 이 문서 조각을 미뤘어요).
- 작업 폴더 이름표: 문서는 `🌳` 가 줄 2 끝(AC-9)에 있다고 해요. 코드는 줄 3(막대 줄) 맨 앞에 `🌳 WT-c<카드>-<이름>` 으로 써요(C-124, SPEC-20261119-statusline-worktree-label 의 D-8 이 이 문서 조각을 미뤘어요).

이 SPEC 은 그 두 문단을 현재 코드와 골든에 맞추는 문서 변경이에요. 코드·골든·시험은 건드리지 않아요.

## Non-goals

- 코드, 골든 파일, 시험의 변경.
- 1번(태그 줄) 문단과 문서의 다른 부분의 고쳐 쓰기.
- 상태줄의 새 기능이나 이름표 규칙 바꾸기.

## Requirements

- REQ-1: The statusline documentation shall describe the card counts of item 2 as `🗂️ <remaining> / ▶️ <working> / 🔃 <waiting> / ✅ <done>` and say what each of the four numbers counts.
- REQ-2: The statusline documentation shall state that a card in `review` counts as waiting and that an `active` card counts as waiting only when the newest of its notes that carries a marker is a waiting marker, with the markers listed.
- REQ-3: The statusline documentation shall state that, when the notes cannot be read within the render budget, the segment shows the three numbers `🗂️ / ▶️ / ✅` with `▶️` counting `active` plus `review` and with no `🔃`.
- REQ-4: The statusline documentation shall place the worktree segment at the front of the last line, the one that holds the `CTX:` and `Week:` bars, and say that item 2 no longer ends with it.
- REQ-5: The statusline documentation shall give the three shapes of the worktree segment, `🌳 WT-c<card>-<name>`, `🌳 WT-c<card>` and the worktree name alone, with the case that picks each.
- REQ-6: The statusline documentation shall state the rules of the worktree label: the lowest card number when several held reservations name the folder, the name taken from the branch and else the folder name, a recorded worktree path that is not absolute naming no folder, at most 32 characters cut with `…`, control and invisible format characters dropped, no segment in the main checkout and none while the `worktree` option is off.
- REQ-7: The statusline documentation shall carry, as its example, lines 2 and 3 of the golden `b14-card-counts-waiting` byte for byte.
- REQ-8: The statusline documentation shall name the tests and golden cases that pin the two items.

## Acceptance

- AC-1 (REQ-7): lines 2 and 3 of `tests/fixtures/statusline-claude/golden/b14-card-counts-waiting.plain.txt` each appear verbatim in the document.
  verify: `python3 -I -c "import pathlib; g=pathlib.Path('tests/fixtures/statusline-claude/golden/b14-card-counts-waiting.plain.txt').read_text().splitlines(); d=pathlib.Path('docs/rust-product-path.md').read_text(); assert g[1] in d and g[2] in d, [l for l in g[1:3] if l not in d]"`
- AC-2 (REQ-1, REQ-2, REQ-3): the old three-number text is gone and the new meaning is written once; the document carries neither `📥 M` nor `🔥 N` in the statusline section and names `🔃`, `LANE-READY-INTEGRATION`, `review`, and the three-number fallback.
  verify: `python3 -I -c "import pathlib,re; d=pathlib.Path('docs/rust-product-path.md').read_text(); s=d[d.index('The three lines follow'):d.index('Directory, PR review state')]; assert '🔥 N' not in s and '📥 M' not in s and '🔃' in s and 'LANE-READY-INTEGRATION' in s and 'review' in s"`
- AC-3 (REQ-4, REQ-5, REQ-6): item 2 does not list `🌳` among the repository-line segments and item 3 opens with the worktree segment; the document names `WT-c`, `32`, `…`, `absolute`, `worktree` option, and `main checkout`.
  verify: `python3 -I -c "import pathlib; d=pathlib.Path('docs/rust-product-path.md').read_text(); s=d[d.index('2. Repository:'):d.index('3. Usage')]; assert '🌳' not in s.split('🔗 #N state')[0]; t=d[d.index('3. Usage'):d.index('Directory, PR review state')]; assert all(k in t for k in ['WT-c','32','…','absolute','main checkout','worktree'])"`
- AC-4 (REQ-8): the statusline section names `tests/statusline_worktree_label.rs`, `b14-card-counts-waiting` and `tests/statusline_cards.rs`.
  verify: `grep -c "statusline_worktree_label.rs" docs/rust-product-path.md`
- AC-5: the document stays inside its CI rules and the repository SPEC check stays green.
  verify: `cargo test --locked --offline --test spec_cli the_repository_specs_pass_check`

## Scope

- write: `docs/rust-product-path.md` (items 2 and 3 of the statusline layout only), `.moai/specs/SPEC-20261120-statusline-docs-items-2-3/**`
- read-only: `src/statusline/render.rs`, `src/statusline/state.rs`, `tests/fixtures/statusline-claude/**`, `.moai/specs/SPEC-20261119-statusline-waiting-cards/spec.md`, `.moai/specs/SPEC-20261119-statusline-worktree-label/spec.md`
- forbidden: any code, golden or test file; the real HOME and ledger

## Plan

1. SPEC first (this file); the facts the text needs are read from the render code and the goldens at the local integration branch, not from the older SPEC prose, and listed under Decisions.
2. The document file is reserved by another card (the C-130 piece about recorded AC commands). The SPEC is written under a reservation that holds the SPEC folder only; the reservation widens to the document once that holder lets it go.
3. Edit item 2 (card counts, waiting rule, fallback, no `🌳`, option and tests) and item 3 (worktree segment first, shapes, rules), nothing else; put the two golden lines in as the example.
4. Run the verify commands above, `git diff --check`, the scan tests that read the document (`tests.ci_scope_test`, `doc_references`), the repository SPEC check, and the pre-READY set.
5. An independent audit reads the document against the render code and the golden line by line, then the READY line.

## Risks

- The document is also edited by another card; the two edits touch different paragraphs, so a merge of the local tip decides it, and a conflict is settled by keeping both texts.
- A number copied from a golden goes stale when the golden is re-recorded; only lines 2 and 3 are repeated (line 1 carries the version), and AC-1 fails on a drift.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | the example is lines 2 and 3 of `b14-card-counts-waiting`, not line 1 | line 1 carries the MoAI version, which the golden harness rewrites; lines 2 and 3 do not change with a release | repeat line 1 with `v<version>` |
| D-2 | the waiting rule is written from `src/statusline/state.rs` (`waiting_cards_with`, `marker_of`) and the render code, not copied from the SPEC | the code is what runs; the SPEC is a plan | none |
| D-3 | the markers are listed by what the code reads (`LANE-READY-INTEGRATION`, `LANE-DONE` with `audit requested` or `감사 요청`, a line starting `감사 요청`, `AUDIT` with `verdict: PASS` as waiting; `ACK `, `PLAN `, `AUDIT` with `verdict: FAIL` as working) | REQ-4 and REQ-5 of the waiting-cards SPEC, checked against `marker_of` before writing | none |
| D-4 | the document is written in the same English and style as the paragraph around it | one file, one voice; the change reads as a fix, not a rewrite | none |

## Result

(draft: filled in with evidence at RUN.)
