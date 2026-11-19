---
id: SPEC-20261120-index-state-folders
title: "The code map leaves MoAI's own state folders out by default, and a go-state stop names the largest folders and the way to move reports out"
status: draft
revision: 1
created: 2026-11-20
---

# The code map leaves MoAI's own state folders out by default, and a go-state stop names the largest folders and the way to move reports out

## Intent

사용자 프로젝트(mo.ai.kr, 이전 세대 MoAI 를 쓰던 bun 모노레포)에서 `/moai project` 가 두 군데서 막혔어요.

- `index-build` 가 거절됐어요: 25,437개 > 16,384, 그중 `.moai/reports` 가 16,760개예요. 보고서는 코드가 아닌데 코드 지도의 개수에 들어갔어요.
- `go-state` 의 미리보기가 `.moai/` 를 그대로 두었어요: `.moai/` 아래 항목이 20만 개를 넘어 중첩 체크아웃 검사를 끝내지 못했고, 사유는 "폴더가 그대로 남는다"뿐이었어요. 보고서 폴더가 74GB 였어요.

C-58 이 이미 "추적 파일만 센다(gitignore 로 빠진 untracked 파일은 안 센다)"와 "한도를 넘으면 가장 큰 폴더를 알려 준다"를 넣었어요(`src/index.rs`). 이 SPEC 은 그 위에 둘을 얹어요.

1. 코드 지도는 기본으로 MoAI 자기 상태 폴더(`.moai/`, `.moai-cli/`) 안의 파일을 세지도 색인하지도 않아요. 보고서·기록은 코드가 아니고, 한도를 먼저 채우는 쪽이에요. 범위(`scope`)가 상태 폴더 안의 자리를 직접 이름 지으면 그때는 색인해요.
2. `go-state-plan` 이 `.moai/` 검사를 한도에서 멈출 때, 막힘 사유가 `.moai/` 아래 가장 큰 폴더를 항목 수와 함께 이름 짓고, 그중 `reports` 가 있으면 보고서를 옮기는 명령을 알려 줘요.

측정한 것(기준 가지 997becf25f 의 코드): `tracked_paths` 는 `git ls-files` 에서 `<state>/index/` 만 빼요. 그래서 git 이 추적하는 `.moai/reports` 는 세어져요. `untracked()` 는 현재 상태 폴더 하나(`state_root::dir`)만 빼요. 이전 세대 프로젝트의 상태 폴더는 `.moai-cli` 로 읽혀서 `.moai/reports` 의 untracked 파일이 목록에 들어가고, 16,384개를 넘으면 질의가 `unavailable` 이 돼요. `go_state` 의 중첩 체크아웃 검사(`nested_checkouts_under`)는 `.moai/` 전체를 걸어요.

## Non-goals

- `docs/index-generation.md` 의 개수 규칙 줄 고치기: C-58 이 그 파일을 쥐고 있어서 이 조각에 넣지 않아요. 후속으로 남겨요.
- go-state 의 중첩 체크아웃 걷기에서 `reports/` 를 건너뛰기: 그 걷기는 `.moai/` 를 통째로 옮겨도 되는지 판단하는 안전 검사라서 fail-closed 를 그대로 둬요.
- 추적 중인 보고서를 옮기거나 지우기: 보고서 이동 명령은 추적되지 않는 파일만 옮겨요. 추적 파일 제거는 사용자 선택이에요.
- 보고서 이동 명령 자체(`reports-move-plan`, `reports-move-apply`): SPEC-20261119-update-moves-reports 의 일이에요.
- `.gitignore` 에 걸리지만 추적되는 파일을 빼기: 커밋한 파일은 사용자의 선택이고, C-58 의 AC-1 은 "커밋한 vendor 폴더는 센다"를 못박았어요.

## Requirements

- REQ-1: Where a build has no scope, `index-build` shall leave the files under `.moai/` and `.moai-cli/` out of the file count and the index.
- REQ-2: Where a scope entry names a place inside `.moai/` or `.moai-cli/`, `index-build` shall index the tracked files under that entry.
- REQ-3: While the published generation's scope names no place inside `.moai/` or `.moai-cli/`, `index-query` shall leave the tracked and untracked files under those folders out of its freshness judgement.
- REQ-4: The state folders shall be matched on the first path component of the project only, so that `docs/.moai/x` and `.moai-notes/a.txt` still count.
- REQ-5: When `go-state-plan` stops checking `.moai/` at its entry limit, it shall name in the reason of the `.moai` item the largest folders under `.moai/` with their entry counts, largest first.
- REQ-6: Where `reports` is among the folders named, the reason shall give the command `moai reports-move-plan .` and the folder `~/.moai/reports/<project>/` the reports move to.
- REQ-7 (REQ-5): IF `.moai/` holds more entries than the limit THEN `go-state-plan` shall keep the `.moai` item at `leave` and move nothing.
- REQ-8: The reason shall show a folder that holds at most the entry limit with its exact count and a larger folder as `more than <limit> entries`.

## Acceptance

- AC-1 (REQ-1): in a repository whose `.moai/reports` and `.moai-cli/reports` each track 20,000 files next to three source files, `index-build` builds, `files` is 3 and the manifest lists no path under either folder; with 16384 source files beside them it builds and with 16385 it is refused.
  verify: `cargo test --locked --offline --test index_state_folders_cli tracked_reports_in_a_state_folder_do_not_count`; `cargo test --locked --offline --test index_state_folders_cli the_limit_counts_source_files_next_to_a_state_folder_full_of_reports`
- AC-2 (REQ-1, REQ-4): the tracked specs, config and project files under `.moai/` and `.moai-cli/` are left out while `.moai-notes/a.txt`, `docs/.moai/x` and `src/main.rs` are indexed.
  verify: `cargo test --locked --offline --test index_state_folders_cli the_state_folder_is_matched_on_the_first_component_only`
- AC-3 (REQ-2): a scope of `.moai/specs` indexes the tracked files under it and nothing else, a scope of `src` and `.moai/specs` indexes both, and an unscoped build still leaves the folder out.
  verify: `cargo test --locked --offline --test index_state_folders_cli a_scope_inside_a_state_folder_indexes_it`
- AC-4 (REQ-3): after a build, 17,000 untracked files under `.moai/reports` and a changed tracked file there leave `index-query` `fresh`, while 16,385 untracked files outside any state folder make it `unavailable`.
  verify: `cargo test --locked --offline --test index_state_folders_cli untracked_reports_in_a_state_folder_leave_a_query_fresh`; `cargo test --locked --offline --test index_state_folders_cli untracked_files_outside_a_state_folder_still_count`
- AC-5 (REQ-5, REQ-7, REQ-8): with an entry limit of 5, a `.moai/` whose `reports` folder holds more than 5 entries is left in place and its reason names `reports (more than 5 entries)`, `logs (3 entries)`, `cache (2 entries)`, `config (2 entries)` and `memory (1 entry)` in that order (largest first, equal counts by name); at most five folders are named; a folder of exactly 5 entries shows `5 entries` and one of 6 shows `more than 5 entries`; under the default limit the same folder still moves.
  verify: `cargo test --locked --offline --test go_state_blocked_folders_cli the_blocked_reason_names_the_largest_folders`; `cargo test --locked --offline --test go_state_blocked_folders_cli at_most_five_folders_are_named`; `cargo test --locked --offline --test go_state_blocked_folders_cli a_folder_at_the_limit_is_counted_and_one_over_is_more_than`; `cargo test --locked --offline --test go_state_blocked_folders_cli the_blocked_folder_stays_and_the_preview_writes_nothing`
- AC-6 (REQ-6): the reason gives `moai reports-move-plan .` and `~/.moai/reports/<project>/` when `reports` is among the named folders and gives neither when it is not.
  verify: `cargo test --locked --offline --test go_state_blocked_folders_cli the_reason_gives_the_report_move_only_when_reports_is_named`
- AC-7 (REQ-1…REQ-8): neighbours.
  verify: `cargo test --locked --offline --test index_scope_cli`; `cargo test --locked --offline --test index_cli`; `cargo test --locked --offline --test freshness_fixture`; `cargo test --locked --offline --test atlas_query_cli`; `cargo test --locked --offline --test go_state_cli`; `cargo test --locked --offline --test go_state_preview_cli`; `cargo test --locked --offline --lib`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/index.rs`, `src/go_state.rs`, `tests/index_state_folders_cli.rs`, `tests/go_state_blocked_folders_cli.rs`, `.moai/specs/SPEC-20261120-index-state-folders/spec.md`
- read-only: `src/state_root.rs`, `src/main.rs`, `tests/index_scope_cli.rs`, `tests/go_state_cli.rs`, `tests/common/**`
- forbidden: `docs/index-generation.md` (held by another card; follow-up), `skills/**`, the real HOME and ledger, the network, a change of `MAX_FILES` or of the entry limit

## Plan

1. 시험 먼저: 두 CLI 시험 파일을 쓰고 코드 없이 실패하는 것을 읽어요.
2. `src/index.rs`: 상태 폴더 두 이름과 첫 성분 판별, `tracked_paths` 와 `untracked` 에서 범위가 비었거나 상태 폴더 안을 이름 짓지 않을 때 빼기, git 쪽 출력이 4 MiB 를 넘지 않도록 같은 폴더를 git 에도 제외로 넘기기.
3. `src/go_state.rs`: `.moai/` 아래 첫 단계 폴더별 항목 수(한도까지), 막힘 사유 문장.
4. 변이 시험, 이웃 시험, 독립 감사.

## Risks

- 이 SPEC 이전에 만든 색인 generation 이 상태 폴더 파일을 기록하고 있으면, 다음 질의는 그 파일들이 사라진 것으로 보고 `stale` 이 돼요. 다시 빌드하면 새 id 로 풀려요(상태 폴더를 추적하는 저장소만 해당, 이관 단계는 없음).
- 사유가 이름 짓는 명령 `moai reports-move-plan` 은 SPEC-20261119-update-moves-reports 가 기준 가지에 들어와야 실제로 있어요. 그 전에 이 변경만 병합되면 사유가 아직 없는 명령을 가리켜요. 병합 순서는 그 SPEC 이 먼저예요.
- 폴더별 항목 수 세기는 폴더마다 한도(20만)까지 걸어요. 첫 단계 폴더가 많고 모두 크면 막힘 경로에서만 느려져요(미리보기의 막힘 사유를 만들 때만 돌아요).
- 범위가 상태 폴더 안을 이름 지어도 `untracked()` 는 현재 상태 폴더(`state_root::dir`)의 untracked 파일을 계속 빼요. 색인이 거기에 직접 쓰기 때문이고(기존 규칙), 그 폴더의 새 untracked 파일은 신선도에 안 들어와요.
- `git` 에 제외 경로를 넘길 때는 `--literal-pathspecs` 를 쓰지 않는 별도 명령을 써요. 그 명령에 넘기는 문자열은 이 코드의 고정 문자열뿐이고 호출자 텍스트는 없어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 상태 폴더는 프로젝트 루트의 `.moai/` 와 `.moai-cli/` 두 이름이에요. 프로젝트가 어느 세대인지와 상관없이 둘 다 빼요 | 이전 세대 프로젝트는 `.moai/`, 이전 Rust 세대는 `.moai-cli/` 를 써요. 현재 세대 판별(`state_root::dir`)만 쓰면 `.moai/reports` 가 샌 것이 이번 사건이에요 | 목록에서 이름을 빼요 |
| D-2 | 범위가 상태 폴더 안의 자리를 직접 이름 지을 때만 색인해요. 새 요청 필드는 만들지 않아요 | 카드가 "기본으로"라고 했고 `scope` 가 이미 있어요 | 필드를 따로 만들어요 |
| D-3 | 상태 폴더 전체를 빼요(`specs/` 포함). 보고서만 빼지 않아요 | 폴더 안에서 "프로젝트 문서"와 "기록"을 가르는 선이 세대마다 달라요. 한 줄로 설명되는 규칙이 예측 가능해요 | `specs/` 를 예외로 둬요 |
| D-4 | 사유는 `.moai/` 아래 첫 단계 폴더 최대 5개를 큰 순서로, 같은 수는 이름 순으로 이름 지어요. 각 폴더는 한도까지만 세고 넘으면 `more than <limit> entries` 로 써요(한도를 넘은 폴더끼리는 같은 수로 보고 이름 순이에요) | 색인 거절 문구(열 개, 같은 수는 경로 순)와 같은 모양이에요. 74GB 폴더를 끝까지 세지 않아요 | 개수를 바꿔요 |
| D-5 | 명령은 `moai reports-move-plan .` 이고, `reports` 가 이름 지어진 폴더 안에 있을 때만 적어요 | 보고서가 원인이 아닌데 보고서를 옮기라고 하면 잘못된 길이에요. 카드가 보고서 규칙을 `~/.moai/reports/<project>/` 로 정했어요 | 항상 적어요 |

## Result

아직 만들지 않았어요(초안).
