---
id: SPEC-20261124-claude-md-product-next
title: "The project picks drop import and gain remove-originals: the code follows the user decision that moai uses only CLAUDE.md and CLAUDE.local.md"
status: draft
revision: 1
created: 2026-11-24
parent: SPEC-20261122-claude-md-layout
---

# Project picks: retire import, add remove-originals

## Intent

사용자 결정(2026-11-22, "AGENTS.md 제거해라. CLAUDE.md CLAUDE.local.md 만 사용한다", 범위는 제품과 이 저장소)은 두 가지를 정했어요. (1) moai 는 `AGENTS.md` 도 `@AGENTS.md` 줄도 쓰지 않아요. 그래서 `import` 선택이 없어져요. (2) 사용자의 기존 `AGENTS.md`·`AGENTS.local.md` 는 `move` 로 글을 옮긴 뒤, 따로 고르는 `remove-originals` 로만 지워요. 이 결정을 `SPEC-20261122-claude-md-project-picks` 개정 4 와 `SPEC-20261122-claude-md-layout` 개정 3 이 적었고 리더가 검토했어요. 그런데 코드는 아직 그대로예요(2026-11-24, 팁 `13e8f7d98c` 에서 읽음): `Pick::Import`, `Kind::Import`, 쓰기기의 선택 목록 `agents::PICKS` 의 `import`, 요청 스키마 enum 의 `import`, `update.md` 의 세 선택 문장이 남아 있고, `remove-originals` 는 SPEC 두 곳에만 있어요. 이 조각이 그 틈을 메워요. 카드의 기준 1(임시 프로젝트의 `AGENTS.md`·`AGENTS.local.md` 를 update 미리보기·선택으로 옮기고, 원본을 두고, 영수증으로 복구)을 선택 집합 `move`·`leave`·`remove-originals` 로 닫아요. 기준 2(출하 글이 `AGENTS.md` 를 moai 의 파일로 부르지 않는 저장소 글 시험)는 다음 조각이고, 이 조각이 `import` 글을 먼저 치워서 그 시험에 예외 줄이 덜 필요해요.

읽기 전용 조사(계획 에이전트, 2026-11-24)가 코드에서 찾은 사실 중 SPEC 을 바꾸는 것이에요.

- 옛 `import` 영수증을 "일반 복구로 되돌려요"라는 개정 4 의 문장은 틀려요. 엔진 영수증은 `receipts/<id>/receipt.json`(폴더) 에 있고, 일반 복구 `apply::recover` 는 `receipts/<id>.json` 만 읽어요(`src/apply.rs:372-384`, `src/agents_local_merge.rs:1478,1544`). 그냥 지우면 `import-line` 부분이 "receipt not found" 로 `kept` 에 떨어지고 `CLAUDE.md` 의 import 블록이 남아요. 되돌리기 쪽만 남겨요(D-1).
- `import` 는 선택 목록 `agents::PICKS`(`src/agents.rs:69-75`)와 스키마 enum(`schemas/protocol/v1/agents-request.schema.json:17`)에도 있어요. 개정 4 는 프로젝트 선택만 말해요.
- 옛 세대가 `AGENTS.md` 의 `moai:merged-local` 블록에 합친 개인 글은 어느 CLAUDE 파일에도 없는데도 `plan_for(Local)` 이 `unchanged` 를 돌려줘요(`src/agents_local_merge.rs:1209-1249`). "unchanged" 를 "찾았다"로 읽는 `remove-originals` 는 그 글을 잃게 해요(D-3).
- 지금의 `settled()` 는 글이 이미 옮겨진 프로젝트에 아무 선택도 주지 않아요. 새 선택은 그 자리를 채우는데, 미리보기가 "최신"으로 읽히는 조건 `all_unchanged`(`src/project_update.rs:1866-1870`)는 `unchanged`·`skip` 만 세요(D-4).

## Non-goals

- 저장소 글 시험(`AGENTS.md` 를 moai 의 파일로 부르는 글을 찾는 Rust 시험), `src/language.rs:93-94` 라벨, `src/generation.rs:1459` 안내문, `.gitignore` 허용 목록, `claude-doctor` 의 제안 문구, `src/scope.rs`·`src/gate.rs` 의 경로 목록. 모두 `SPEC-20261122-claude-md-text-and-ignore` 의 다음 조각이에요.
- `moai update` 의 `project_files` 층(`SPEC-20261122-claude-md-project-layer` 의 둘째 조각)과 옛 블록 제거의 확인 토큰. `src/update.rs` 는 뜨거운 파일이라 따로 해요.
- 하위 폴더 `AGENTS.md`(조정 A19). 이 조각은 프로젝트 루트의 두 파일만 다뤄요.
- 사용자의 실제 HOME 과 `~/.claude`. 시험은 임시 프로젝트와 임시 HOME 만 써요.

## Requirements

- REQ-1: The updater shall refuse the pick `import` as an unknown pick in the project picks, in the `picks` list of a generation request and in the request schema, and write nothing. · failure: none — the refusal writes nothing
- REQ-2: When the project holds `AGENTS.md` or `AGENTS.local.md` and the text of an original that holds user text is not yet in its CLAUDE file, the update preview shall list `move` (recommended) and `leave` and no `remove-originals`. · failure: none — a preview writes nothing
- REQ-3: When every original that holds user text has that text in its CLAUDE file and at least one original exists, the update preview shall list the pick `remove-originals` alone, with one command, one plan digest and the list of originals it would delete. · failure: none — a preview writes nothing; a project with no original or with nothing it may delete is offered no pick
- REQ-4: When the pick is `remove-originals`, the updater shall delete each original and keep its bytes, mode and sha256 in a private folder of the receipt of that pick. · failure: none — the refusals are REQ-5
- REQ-5 (REQ-4): IF an original is a link, lies below a link folder, is not valid UTF-8, is larger than 128 KiB or changed after the preview, THEN the updater shall delete no original and write nothing.
- REQ-6: The updater shall count the text of an original as found only when its CLAUDE file holds that text whole, `CLAUDE.md` for `AGENTS.md` and `CLAUDE.local.md` for `AGENTS.local.md`. · failure: none — a text that is not found makes the pick unavailable; a text that an earlier generation merged into `AGENTS.md`, a blank file and a text moai generated are settled by D-3 and D-5
- REQ-7 (REQ-3): IF a CLAUDE file of the project imports `@AGENTS.md` or is a link to `AGENTS.md`, THEN the update preview shall offer no `remove-originals`.
- REQ-8: When the receipt of `remove-originals` is reverted, the updater shall write each deleted original back with its bytes and mode. · failure: none — a file that exists again at that path is REQ-9
- REQ-9 (REQ-8): IF a file exists again at the path of a deleted original, THEN the updater shall leave that file untouched and name it in the answer.
- REQ-10: When `remove-originals` keeps a private copy, the updater shall add the ignore rule for the receipts folder in the same apply unless a rule of the project already ignores it. · failure: none — the rule is one line written through the receipt of the same pick
- REQ-11: When a receipt holds the part `import-line` of the retired pick, the revert command shall restore the file from that receipt and leave it untouched and named when it changed after the apply. · failure: none — only the revert side stays; nothing writes that part any more
- REQ-12: The shipped update workflow shall list `move`, `leave` and `remove-originals`, say that `remove-originals` is the one deletion of an original the user picks and that its receipt keeps the bytes, and keep the word `older` on every line that names an AGENTS file.

## Acceptance

- AC-1 (REQ-1, REQ-2): in a temporary project with `AGENTS.md` and `AGENTS.local.md` the preview lists `move` (recommended) then `leave` with one digest each, writes nothing, and `import` is refused as unknown by the project picks, by a generation request and by the schema, with no file written.
  verify: `cargo test --locked --offline --test claude_md_picks_cli picks_`
  verify: `cargo test --locked --offline --test agents_choices_step_cli schema_`
- AC-2 (REQ-3, REQ-6, REQ-7): after `move` the preview lists `remove-originals` alone with the originals it would delete; a project whose text sits only in the `moai:merged-local` block of `AGENTS.md`, whose `CLAUDE.md` imports `@AGENTS.md` or links to it, or whose text is not found gets no pick; a blank or generated `AGENTS.local.md` does not hold the pick back.
  verify: `cargo test --locked --offline --test claude_md_picks_cli settled_`
- AC-3 (REQ-4, REQ-5, REQ-10): the apply deletes both originals and nothing else, keeps their bytes and modes in an owner-only folder of the receipt, adds the receipts ignore rule, and a link original, a link folder, a file that is not UTF-8, a file above 128 KiB and a file changed after the preview each leave both originals and every other file byte-identical.
  verify: `cargo test --locked --offline --test claude_md_picks_cli remove_`
- AC-4 (REQ-8, REQ-9, REQ-11): the revert writes both originals back byte-identical with their modes, names and keeps a file recreated since, and still restores a receipt that holds an `import-line` part built by the earlier version.
  verify: `cargo test --locked --offline --test claude_md_picks_cli receipt_`
- AC-5 (REQ-4, REQ-5): a failure after the first original was deleted puts it back through the same receipt and names what it could not put back.
  verify: `cargo test --locked --offline --lib agents_local_merge::picks::tests`
- AC-6 (REQ-12): the update workflow lists the three picks, names the deletion pick and keeps the scan green.
  verify: `cargo test --locked --offline --test next_flow_skill_scan`
  verify: `cargo test --locked --offline --test source_manifest_facts`
- AC-7: neighbouring checks on the changed files.
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/agents_local_merge.rs`, `src/agents_local_merge/picks.rs`, `src/agents_local_merge/originals.rs` (new), `src/agents.rs`, `src/project_update.rs`, `src/ops/agents_merge_plan.rs`, `src/ops/agents_merge_apply.rs`, `src/ops/agents_merge_revert.rs`, `schemas/protocol/v1/agents-request.schema.json`, `tests/claude_md_picks_cli.rs`, `tests/claude_md_engine_cli.rs`, `tests/claude_md_layout_cli.rs`, `tests/agents_choices_step_cli.rs`, `tests/project_update_plan_cli.rs`, `skills/moai/workflows/update.md`, `skills/moai/references/source-manifest.json`, `.moai/specs/SPEC-20261124-claude-md-product-next/**`, `.moai/specs/SPEC-20261122-claude-md-project-picks/**`, `.moai/specs/SPEC-20261122-claude-md-layout/**`
- read-only: `src/apply.rs`, `src/claude.rs`, `src/update.rs`, `tests/next_flow_skill_scan.rs`
- forbidden: the real HOME, the network, `.claude/` of this checkout, `src/update.rs`, `src/help.rs`, `tools/repo_scans.json` (the next slice asks for it in turn)

## Plan

1. SPEC first: this SPEC, the picks SPEC revision 5 and the layout SPEC revision 5 agree on the picks, the revert side and the scope; the leader approves them before any code.
2. Red tests first: edit the 14 tests that name `import` (7 and the removal in `claude_md_picks_cli`, 2 in `claude_md_engine_cli`, 4 in `claude_md_layout_cli`, 1 in `agents_choices_step_cli`) and write the `remove_` and `settled_` tests; run them red.
3. Retire `import`: `Kind::Import`, `find_import`, `IMPORT_*` constants and notices, `Pick::Import`, `IMPORT_EFFECT`, `State.import`, the `import` id of `agents::PICKS` and of the schema enum, the writer warning that names it, the help text of the three `agents-merge-*` commands. Keep `imports_agents_md` (a user's own line counts as settled) and the revert side of an old `import-line` part under the prefix `agents-import-` (D-1).
4. `originals.rs`: the plan (per original: path, sha256, bytes, found in which CLAUDE file, verdict), the apply (re-read, compare sha256, write the private copy under `originals/<name>` of the receipt folder with owner-only bits, save the receipt, re-compare, remove the file; failure rolls back through `fail_with_rollback`), the revert (write back with the recorded mode, leave a recreated file and name it) and the receipts ignore line.
5. `picks.rs`: `Pick::ALL` is `[Move, Leave, RemoveOriginals]`; `settled()` splits into the question whether a `move` is needed and the question whether a removal may be offered; the generation request is built only for `move`; the `deletes` key lists the originals and `writes` stays the files a pick writes.
6. `project_update.rs`: the reason texts, the refused computation and the step state when only `remove-originals` is offered (D-4).
7. `update.md` and the source manifest hash (`tools/source_manifest.py`, run as the repository documents it).
8. One `moai check affected` at the READY head, unwrapped; the guard-mutant line for each new guard test (the digest compare before the delete, the all-or-nothing refusal, the `merged-local` not-found rule, the revert mode).

## Risks

- Deleting a user's file is the one irreversible-looking write of this family. The bytes, the mode and the sha256 are in the receipt, the apply compares the sha256 against the previewed bytes just before the delete, and the revert restores them. The compare and the delete are two calls; the existing engines accept that gap (`restore_file`, `src/agents_local_merge.rs:1566`), and a rename to a temporary name, a verify and an unlink would close it. Whether that is worth its code is the leader's to say; the default is the existing check-then-delete.
- A tracked `AGENTS.md` shows up as a git deletion after the pick, and another tool that reads `AGENTS.md` (Codex, Cursor) stops seeing it. The preview names both before the user picks.
- Text a person wrote inside the old managed section of `AGENTS.md` is not copied (layout D-5). It survives only in the receipt; D-6 warns and does not refuse.
- `read_checked` caps a read at 128 KiB, so a larger `CLAUDE.md` reads as no text and the pick is unavailable. That is the safe side; it is not tested as a feature.
- Hot files: `src/agents_local_merge.rs`, `src/agents.rs`, `src/project_update.rs` and `skills/moai/workflows/update.md` are edited by other open work. The merge conflict is the lander's; the card check names a holder when one exists.
- Files named in `Scope` that this card does not reserve are taken by the leader's word, as the earlier slices did; the READY quotes each `moai card check`.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The revert side of an old `import-line` part stays; the write side goes. `Kind::Import` is replaced by the prefix `agents-import-` and the allowed name `CLAUDE.md` passed to the existing revert. | An engine receipt is a folder `receipts/<id>/receipt.json` that the generic recovery does not read (`src/apply.rs:372-384`); a build earlier than this one may have written such a receipt. | Drop old import receipts and say so in the preview. |
| D-2 | `remove-originals` is all or nothing; the preview carries one verdict per original. | A half-deleted pair can leave the personal text with no home; the picks SPEC REQ-4 already says nothing is deleted when one original is refused. | Delete per original. |
| D-3 | "Found" means a text that sits in a CLAUDE file, whole: the marked block with a matching sha256, or the same lines pasted by hand. A plan action `unchanged` is not "found". When `AGENTS.md` holds a `moai:merged-local` block, that block's text must also be in `CLAUDE.local.md`, else the pick is unavailable. | `merged_into_agents_md` answers `unchanged` for text that is in no CLAUDE file (`src/agents_local_merge.rs:1209-1249`). | Count `unchanged` as found. |
| D-4 | When only `remove-originals` is offered the step reads `optional`: it prints the choice and the command, and `all_unchanged` counts it as nothing to do. `leave` is not listed there. | `needs_decision` would keep the preview from ever reading up to date for a user who keeps the originals; `leave` says Claude Code keeps reading `AGENTS.md`, which is false once `CLAUDE.md` exists. | Use `needs_decision`, or list `leave` too. |
| D-5 | A blank or moai-generated `AGENTS.local.md`, and an `AGENTS.md` that holds only the old managed block, hold no text to find; the first needs no `CLAUDE.local.md`, the second needs the managed block in `CLAUDE.md`. | The first-creation text is never copied (layout D-4); the old block is not copied (layout D-5). | Require a copy for every original. |
| D-6 | The preview warns that text a person wrote inside the old managed section of `AGENTS.md` is not copied and survives only in the receipt. It does not refuse. | Whether the manifest still carries a baseline for that section is not verified; a refusal would block every project that never edited it. | Refuse when no baseline matches. |
| D-7 | The private copies live in `originals/<name>` of the pick's own receipt folder with owner-only bits, never through `apply::save_before_bytes`; the apply adds the receipts ignore line. | `AGENTS.local.md` may hold secrets; the engines already create `before/` and `source/` that way. | Reuse the generic before-bytes store. |
| D-8 | `import` leaves `agents::PICKS` and the schema enum. A request that holds it is invalid input. | The writer pick and the project pick share the id; the user decision removes both. | Keep it in the schema only. |
| D-9 | The two readers of "imports AGENTS.md" (`agents_local_merge::imports_agents_md` and `claude::imports_agents`) stay as they are in this slice; the removal check uses the first, which is word-exact and honours `~~~` fences. | The second is shared with `claude-doctor`; unifying them is a change to that command. | Unify them in the text slice. |

## Result

(Filled when the work is done.)
