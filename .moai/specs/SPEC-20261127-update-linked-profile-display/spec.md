---
id: SPEC-20261127-update-linked-profile-display
title: "moai update shows a Claude config folder that links into the default folder as the default folder and says plainly what is blocked"
status: draft
revision: 2
created: 2026-11-27
---

# moai update shows a Claude config folder that links into the default folder as the default folder and says plainly what is blocked

## Intent

Orca 세션은 `CLAUDE_CONFIG_DIR` 로 프로필 폴더를 가리키고, 그 폴더의 `rules`, `output-styles`, `skills`, `agents`, `settings.json`, `CLAUDE.md` 는 `~/.claude` 의 같은 이름으로 가는 링크예요(이 세션의 프로필 폴더를 읽기 전용으로 본 결과, 2026-11-27). 사용자가 이 프로필에서 `moai update` 를 보면 routing 이 'moai.md up to date · moai.md skipped' 로, output style 두 줄이 'state root is not of this generation' 으로 막혀 보였고, 실제로는 쓸 것이 없었어요(카드 본문).

NeXT 끝(`cedcd1f234`)을 코드와 기존 테스트로 읽은 결과(조사 에이전트 보고, 2026-11-27 10:24Z)는 이래요. 카드가 본 증상의 원인은 이미 풀려 있어요. `rules` 가 `~/.claude/rules` 로 가는 링크이면 routing 행은 `linked` 이고 화면에 'moai.md linked' 로 나와요(`tests/update_conflict_baseline_cli.rs` 가 지켜요). `output-styles` 폴더 링크와 `settings.json` 링크가 기본 폴더로 가면 기록 이름은 `claude.json` 이고 두 대상은 `unchanged` 예요(`tests/update_linked_config_files.rs` 의 `a_config_file_linked_into_the_default_folder_is_planned_and_written_there`). 실제 프로필 모양에서 막힘이 남는다는 증거는 아직 없어요(실행하지 않았어요, unverified).

남은 것은 보여 주는 쪽이에요. (1) `linked` 프로필 행이 기본 행과 같은 이름 'moai.md' 로 두 번 나와 둘이 다른 파일처럼 읽혀요. (2) `not_linked` 행은 경고지만 낱말이 '건너뜀' 이고 이유 문장은 화면에 안 나와요(`h["section"]["detail"]` 를 읽지 않아요). (3) 아직 막히는 경우(진짜 둘째 폴더, 또는 `output-styles` 가 링크가 아닌 프로필)의 문구는 'state root', 'generation', 'moved state folder' 같은 안쪽 말로 쓰였고 `CLAUDE_CONFIG_DIR` 를 빼고 돌리는 길은 말하지 않아요. (4) 안내문과 요약이 쓰기가 닿는 `~/.claude` 가 아니라 프로필 폴더 경로를 적고, 모든 입구가 링크인데도 "프로필에 쓴다"는 경고를 내요.

이 SPEC 은 이 네 가지만 고쳐요. 쓰는 동작, 기록 이름, 계획 JSON, 영수증에 저장되는 값은 바꾸지 않아요. 읽는 쪽은 `moai update` 를 돌리는 사용자, 이 변경을 구현하는 레인, 리뷰하는 세션이에요.

## Non-goals

- 기록 이름(`record_name`), 링크를 따라가는 규칙(`host_home::place_in`, `home_style::linked_file`), 쓰기와 되돌리기예요. 지금 그대로예요(카드 기준 2 는 이 SPEC 의 AC-5 가 고정 테스트로 지켜요).
- 계획 JSON 의 값(`config_dir`, `file`, `path`)과 영수증에 저장되는 값(`config_folder`)이에요. 바꾸면 다이제스트와 되돌리기의 짝 맞추기가 달라져요. 화면과 안내문의 글자만 바꿔요.
- 링크를 만들거나 바꾸는 일, HOME 상태 폴더를 `.moai-cli` 에서 `.moai` 로 옮기는 일이에요.
- 진짜 `~/.claude`, 사용자의 Orca 프로필 폴더, 진짜 HOME 을 읽거나 쓰는 테스트예요. 모든 테스트는 임시 HOME 만 자식 프로세스에 줘요.
- 진짜 둘째 폴더(프로필이 링크가 아니라 자기 파일을 가진 경우)를 막는 일을 풀어 주는 일이에요. 막는 일은 그대로 두고 말만 바꿔요.

## Requirements

- REQ-1: WHEN a Claude config folder other than the default folder has a routing row `linked` (its `rules` is a link to the default folder's `rules`), the update preview shall list that folder's routing file as the default folder's entry with the note `(이 프로필은 링크로 읽어요)` (English: `(this folder is read through a link)`) and keep the mark `Ok`, so that one file is not shown as two.
- REQ-2 (REQ-1): IF the default folder's routing row is a conflict, is blocked or is not `Ok`, THEN the update preview shall show the profile's row with the default row's mark and words and without the note, so that a problem on the one real copy is never shown as fine.
- REQ-3: WHEN a profile row is `not_linked`, the update preview shall show it with the mark `Warn`, its file name, the reason and the next step taken from the row's detail, in place of the word 'skipped'.
- REQ-4 (REQ-3): IF a `not_linked` row carries no detail, THEN the preview shall show a plain sentence that moai keeps one copy in the default folder and writes none in this folder, so that the row is never a bare 'skipped'.
- REQ-5: WHEN an output style target is blocked because its folder is a separate second folder, the update preview shall give a text that names the blocked file and the way out (run `moai update` without `CLAUDE_CONFIG_DIR`, or link this folder's `output-styles` to the default folder's) in words that use none of 'state root', 'generation' and 'moved state folder'.
- REQ-6 (REQ-5): IF the output style target is blocked for another reason (a link that is refused, an unreadable file, a default folder that is a link), THEN the update preview shall keep that reason and its next step instead of the REQ-5 text, so that no refusal is described wrongly.
- REQ-7: WHEN `CLAUDE_CONFIG_DIR` names a folder whose entries that this update writes (routing, skills, agents, output style, settings) are all links into the default folder, the update output shall name the default folder's paths for those entries and give no warning that the update writes into the profile folder.
- REQ-8 (REQ-7): IF any entry that the update writes is a real file or folder inside the profile folder, THEN the preview shall name that real place and keep the warning for it, so that the person always sees where a write lands.
- REQ-9: WHEN the update plans, applies or reverts for a profile that links into the default folder, the files written, the record names, the plan digests and the receipt values shall be the same as before this change. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2): 화면 골든 테스트에서 `rules` 링크 프로필의 `linked` 행이 기본 행의 'moai.md' 하나에 `(이 프로필은 링크로 읽어요)` 를 붙이고 마크가 `Ok` 이며, 기본 행이 conflict 이면 메모가 없고 마크가 같이 `Warn` 이에요 · verify: `cargo nextest run --locked --offline --test update_tux_golden -E 'test(a_linked_profile_routing_row_is_the_default_entry_with_a_note)'`
- AC-2 (REQ-3, REQ-4): `not_linked` 행이 `Warn` 마크와 이유·다음 걸음 문장으로 나오고 '건너뜀'·'skipped' 낱말이 없으며, detail 이 없는 행은 한 문장의 일반 안내로 나와요 · verify: `cargo nextest run --locked --offline --test update_tux_golden -E 'test(a_not_linked_profile_row_shows_its_reason_as_a_warning)'`
- AC-3 (REQ-5, REQ-6): 진짜 둘째 폴더가 이전 세대 HOME 에서 막힐 때 문구가 세 안쪽 말을 안 쓰고 `CLAUDE_CONFIG_DIR` 없이 돌리는 길과 링크 거는 길을 말하며, 링크 거절·기본 폴더가 링크인 경우의 막힘 문구는 그 이유 그대로예요 · verify: `cargo nextest run --locked --offline --test home_style_cli -E 'test(a_blocked_second_folder_names_the_way_out_without_internal_words)'`
- AC-4 (REQ-7, REQ-8): 임시 HOME 에서 프로필의 모든 입구가 기본 폴더로 가는 링크일 때 미리보기 출력에 기본 폴더 경로가 나오고 프로필 경로와 '프로필에 쓰는' 경고가 없으며, 입구 하나가 프로필 안의 진짜 폴더이면 그 경로와 경고가 나와요 · verify: `cargo nextest run --locked --offline --test update_linked_config_files -E 'test(a_fully_linked_profile_is_shown_as_the_default_folder)'`
- AC-5 (REQ-9): 기존 테스트가 통과해요. 고치는 것은 옛 막힘 문구나 옛 행 글자를 단언하던 줄뿐이고(AC-3·AC-1·AC-2 가 새 글자로 바꿔요), 링크 프로필의 계획·적용·되돌리기, 진짜 둘째 폴더 막기, 기록 이름을 단언하는 줄은 그대로예요 · verify: `cargo nextest run --locked --offline --test update_linked_config_files --test home_style_cli --test claude_md_home_cli --test update_conflict_baseline_cli --test update_tux_golden --test desktop_folder_rules_cli`
- AC-6 (REQ-1…REQ-9): `moai check affected --card <base>..<head>` 가 PASS 예요 · verify: `moai check affected --card <base>..<head>`

## Scope

- write: `.moai/specs/SPEC-20261127-update-linked-profile-display/**`
- write (구현 슬라이스, 범위는 구현 때 카드에 넓혀요): `src/update_tux.rs` (routing 행과 style 행의 글자), `src/update.rs` (`claude_folders_notice`, `layer_summary` 의 글자), `src/home_style.rs` (막힘 문구 한 곳), `tests/update_tux_golden.rs`, `tests/update_linked_config_files.rs`, `tests/home_style_cli.rs`, `tests/desktop_folder_rules_cli.rs`
- read-only: `src/home_routing/rules.rs`, `src/host_home.rs`, `tests/claude_md_home_cli.rs`, `tests/update_conflict_baseline_cli.rs`
- forbidden: `CLAUDE.md`
- forbidden: `src/home_routing/**`
- forbidden: `src/host_home.rs`
- 그 밖의 금지(경로가 아니라 대상): 사용자의 진짜 기본 Claude 폴더와 Orca 프로필 폴더, 진짜 HOME, `record_name` 과 링크를 따라가는 규칙의 동작, 계획 JSON 과 영수증에 저장되는 값

## Plan

1. 슬라이스 1(이 SPEC): 이 문서만 커밋하고 `SPEC-READY` 로 리더에게 알려요. 승인이 필요한지는 리더가 정해요(화면 글자만 바꾸는 변경이라 tier 1 로 봐요, unverified).
2. 슬라이스 2: AC-1 부터 순서대로 RED 테스트를 먼저 쓰고(AC-1, AC-2, AC-3, AC-4), 하나씩 고쳐 GREEN 으로 만들어요. AC-5 의 고정 테스트는 고치기 전에 한 번 돌려 통과를 확인해요.
3. 화면 글자가 바뀌는 골든 파일과 이 글자를 읽는 테스트(`tests/areas.toml` 의 짝)를 같은 변경에서 맞춰요.

## Risks

- 'moai.md' 두 줄을 한 줄로 합치면 이 낱말을 세는 기존 골든 테스트가 깨질 수 있어요. 깨진 테스트의 이름을 READY 에 적고 같은 변경에서 맞춰요(`update_tux_golden`, `update_conflict_baseline_cli`).
- 프로필의 입구 일부만 링크인 혼합 모양이 있어요. REQ-8 이 진짜 입구는 경고를 남기게 하고 AC-4 가 한 입구를 진짜 폴더로 둔 경우를 지켜요.
- 이 SPEC 의 원인 진단은 코드와 테스트 읽기예요. 사용자의 실제 실행에서 아직 막힘이 보이면(Orca 프로필 모양이 다르거나 빌드가 달라서) 그 출력을 카드에 붙여 받아 REQ 를 고쳐요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 카드 기준 2(기록 이름)는 코드를 바꾸지 않고 고정 테스트(AC-5)로 지켜요 | NeXT 끝에서 링크 프로필의 계획·적용·되돌리기가 이미 `claude.json` 기록으로 통과해요(`tests/update_linked_config_files.rs`). 조사 보고 2026-11-27 | AC-5 의 목록에서 뺀 뒤 기준 2 를 다시 열어요 |
| D-2 | 계획 JSON 과 영수증의 값은 바꾸지 않고 화면·안내문의 글자만 바꿔요 | 값이 바뀌면 다이제스트와 되돌리기의 짝 맞추기가 달라져요(REQ-9) | 글자를 만드는 함수만 되돌려요 |
| D-3 | 프로필 `linked` 행의 메모 낱말은 카드가 적은 `(이 프로필은 링크로 읽어요)` 로 해요(영어는 `(this folder is read through a link)`) | 카드 기준 1 의 표현 | 낱말은 화면 글자 테이블 한 곳에 있어요 |
| D-4 | tier 는 1 로 다뤄요(화면 글자만 바꾸고 쓰기와 HOME 상태는 건드리지 않아요). 승인 블록 없이 구현하고 SPEC 은 draft 로 남겨요 | 리더 결정, 카드 노트 10784. 쓰기 경로, 계획 JSON 필드, 영수증 값, tier 2 바닥 경로를 건드리게 되면 멈추고 알려요 | 리더의 결정 | 리더의 결정 |
