---
id: SPEC-20261127-state-kept-narrowing
title: "The managed ignore block allows only the sub-paths of manifests/, workflows/ and decisions/ that Rust writes; the first-level KEPT names stay"
status: approved
revision: 1
created: 2026-11-27
approval:
  by: user
  at: 2026-11-27
  basis: "사용자가 리더 세션의 질문 도구에서 직접 고른 답(승인 노트 10722): \"블록 줄만 좁힘 (Recommended)\" (D-1 은 a, D-2 는 B); 이 레인이 기록 직전에 이 파일에서 moai spec show 로 approval_digest 를 다시 읽어 노트 10721 의 값과 같음을 확인했어요(선택은 리더가 옮긴 글이고 이 레인이 직접 보지는 않았어요)"
  digest: "sha256:18854a79f3f39196a0f1c36a33428f9e002fa66e49c1d73149fc12926cc14ad0"
---

# The managed ignore block allows only the sub-paths of manifests/, workflows/ and decisions/ that Rust writes; the first-level KEPT names stay

## Intent

관리 `.gitignore` 블록은 `state_root::KEPT`(`.moai/`의 첫 단계 이름 열한 개)를 폴더째 `!` 줄로 풀어요(`src/agents.rs`의 `render_ignore`, `STATE_KEPT`). 그래서 `manifests/`, `workflows/`, `decisions/` 안에 있는 파일은 Rust가 쓴 것이든 아니든 모두 git에 보여요. Rust가 이 폴더들에 쓰는 것은 정해져 있어요(아래 Facts). 이전 세대(Go) 프로젝트는 같은 이름 폴더 안에 다른 파일을 두어요: 예를 들어 카드 C-578이 밝힌 `.moai/workflows/*.md`(CLAUDE.md와 설정이 아직 읽는 파일)가 있어요. 이런 파일이 `workflows/` 허용 줄 때문에 추적되지 않은 새 파일로 보이고, `git add -A` 한 번에 커밋에 딸려 들어가요.

사용자 직접 선택(2026-11-27, 카드 C-300 노트 10517): "승인, 좁히기는 후속". 이 SPEC은 그 후속이에요. 허용 목록을 Rust 소유 하위 경로로 좁히는 안을 사실로 보여 주고, 사용자가 고른 뒤에만 코드를 바꿔요(카드 C-591 기준 1). 좁혀서 얻는 것은 크지 않아요: 이 SPEC의 기본안은 `workflows/`의 평평한 옛 파일과 `manifests/`·`decisions/`의 Rust가 안 쓰는 파일만 가려요. 그래서 "지금 그대로 둔다"(D-2의 A)도 정당한 선택이에요.

## Interview

- Q1: `STATE_KEPT` 좁히기를 C-300에 넣을까요, 후속으로 뺄까요? → A: "승인, 좁히기는 후속"(사용자 직접 선택 2026-11-27, 기록은 카드 C-300 노트 10517)
- Q2: 아래 Decisions의 D-1(`config/`·`docs/`)과 D-2(좁히는 범위)는 어느 쪽인가요? → A: "블록 줄만 좁힘 (Recommended)"(사용자 직접 선택 2026-11-27, 리더 세션의 질문 도구, 기록은 카드 C-591 노트 10722): D-1 은 a, D-2 는 B예요. 선택은 리더가 옮긴 글이고 이 레인이 직접 보지는 않았어요.

## Facts

2026-11-27, 통합 브랜치 `NeXT`(`8b6ab17634`)에서 코드를 읽어 확인한 것이에요.

| 이름 | Rust가 쓰는 것(코드 근거) | 이전 세대가 같은 이름에 두는 것 |
| --- | --- | --- |
| `manifests/` | `generation.json`, `claude-settings.json`, `profile-baseline.json`, `sources.json`, `ci.json`, `github.json`, `update-declined.json`(`src/generation.rs`, `src/hooks.rs`, `src/profile.rs`, `src/project_ci_plan.rs`, `src/github_templates.rs`, `src/project_update/declined.rs`) | 찾지 못했어요(확인 못 함) |
| `workflows/` | `workflows/<id>/SKILL.md`, `workflows/<id>/references/workflow.md`(`src/generation.rs:840-920`) | 평평한 `workflows/*.md`(C-578 본문) |
| `decisions/` | `decisions/<run>.jsonl`(`src/decision.rs:184`) | 찾지 못했어요(확인 못 함) |
| `config/`, `docs/` | 쓰는 곳이 없어요. Rust는 이 폴더를 만들지 않아요 | `config/config.yaml`, `config/sections/*.yaml`(Go 마커, `src/state_root.rs` `go_markers`), `docs/`(go-state `USER_DATA_FROM`이 백업에서 복원) |
| `project.yaml`, `moai-cli.json`, `moai.json` | 프로필과 세대 마커 | 같은 이름의 마커 |
| `specs/`, `tools/`, `evals/` | `specs/`는 SPEC이에요. `.moai/tools/`와 `.moai/evals/`에 Rust가 쓰는 것은 찾지 못했어요(레이아웃 SPEC REQ-1이 프로젝트 소유로 정해요). `src/review.rs`와 `src/repo_scans.rs`가 읽는 `tools/*.json`은 저장소 루트 `tools/`이고 `.moai/tools/`가 아니에요 | `specs/`는 `USER_DATA_FROM`에 있어요 |

`KEPT`를 읽는 곳은 넷이에요: 블록의 허용 줄(`src/agents.rs:686`), 분류기(`src/state_root.rs:507`, 첫 단계 이름이 모두 kept 또는 `local`이면 Rust 폴더), go-state의 `kept_name`(`src/go_state.rs:2410`, 적용 전에 있던 이름), `is_kept`(`src/state_root.rs:917`). 목록 자체는 `src/state_root.rs`의 단위 테스트 `the_layout_names_kept_kinds_and_local_paths`가 고정해요.

## Non-goals

- `KEPT`의 첫 단계 이름 열한 개는 바꾸지 않아요. `classify`와 go-state `kept_name`이 읽는 것은 그대로예요(D-2의 B).
- `config/`와 `docs/`의 허용 줄은 기본안에서 그대로 둬요(D-1의 a).
- go-state·go-residue의 이동 동작과 C-300의 leftovers 보관은 바꾸지 않아요.
- 이미 git이 추적하는 파일을 추적에서 빼지 않아요(REQ-5).
- 승인된 `SPEC-20261120-moai-layout`의 첫 단계 규칙(REQ-1)은 그대로고, 이 SPEC은 블록 줄(그 SPEC의 REQ-3)을 더 좁게 정할 뿐이에요.
- Windows·Linux 결과를 주장하지 않아요(Mac 픽스처만).

## Requirements

- REQ-1: The managed ignore block shall allow, under `manifests/`, only files with the extension `.json` directly in that folder; under `workflows/`, only sub-folders and what they hold; and under `decisions/`, only files with the extension `.jsonl` directly in that folder.
- REQ-2: The managed ignore block shall keep the whole-folder or whole-file `!` line of every other kept name (`project.yaml`, `moai-cli.json`, `moai.json`, `config/`, `docs/`, `specs/`, `tools/`, `evals/`), which is the default of D-1.
- REQ-3: `state_root::KEPT` shall stay the list of first-level names.
- REQ-4: The narrower allow lines shall be a second list of `state_root` that only the block reads, so `classify`, `go_state::kept_name` and `is_kept` return what they return today for every name.
- REQ-5: When git tracks a path under the state folder that the narrower block would ignore, the block shall carry an exact `!` line for that path, so no tracked file becomes ignored.
- REQ-6: When git tracks a path under the state folder that the narrower block would ignore, the project update plan shall name it as `now hidden: <path>` (up to twenty; the rest counted in one line), as REQ-12 of `SPEC-20261120-moai-layout` does for `now visible`.
- REQ-7: A test shall fail when a path that a Rust writer spells under `manifests/`, `workflows/` or `decisions/` is not allowed by the rendered block, so a new writer can never be silently ignored.
- REQ-8: The rendered block shall be tested against real git (`git check-ignore -v` in a temporary repository): the Rust paths of the Facts table are visible, a flat `workflows/w1.md`, a `manifests/x.txt` and a `decisions/x.md` are ignored, and a tracked flat `workflows/w1.md` stays tracked and visible through its exact `!` line.
- REQ-9: Planning and applying shall render the same block, and `ignore_block` stays the one function both call.
- REQ-10: When the managed block of a project already holds the new lines, an apply shall change nothing. · failure: none — the block is written from the same text every time, so a repeat has no step that can fail
- REQ-11 (REQ-5, REQ-6): IF git cannot list the tracked paths of the state folder, THEN the managed ignore block shall keep the whole-folder `!` lines of every kept name, as it does today.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): the allow lines of the block name exactly the narrower sub-paths of REQ-1 and the whole-folder lines of the other names, and `KEPT`, `is_kept`, the classifier verdicts and `kept_name` answer the same for every name as before · verify: `cargo test --locked --offline --lib state_root::tests::the_narrow_allow_list_leaves_the_first_level_names_alone`
- AC-2 (REQ-1, REQ-8): in a temporary git repository with the rendered block, the Rust paths are visible and the flat Go-shaped files are ignored · verify: `cargo test --locked --offline --test gitignore_block_cli the_block_allows_only_the_paths_rust_writes_under_manifests_workflows_decisions`
- AC-3 (REQ-5, REQ-6, REQ-8): a tracked flat `workflows/w1.md` and a tracked `manifests/x.txt` get exact `!` lines and stay visible, and the plan prints `now hidden: <path>` for each · verify: `cargo test --locked --offline --test gitignore_block_cli a_tracked_file_the_narrow_block_would_hide_keeps_an_exact_allow_line`
- AC-4 (REQ-7): the scan lists every `manifests/<name>`, `workflows/<id>/...` and `decisions/<run>` path the sources spell and finds each allowed by the block; a made-up writer path outside the lines fails it · verify: `cargo test --locked --offline --test gitignore_block_cli every_state_path_a_writer_spells_is_allowed_by_the_block`
- AC-5 (REQ-9, REQ-10): applying twice leaves the file byte-identical the second time, and plan and apply print the same block text · verify: `cargo test --locked --offline --test gitignore_block_cli the_narrow_block_is_the_same_in_plan_and_apply_and_idempotent`

## Scope

- write: `.moai/specs/SPEC-20261127-state-kept-narrowing/**`, `src/state_root.rs` (the allow list constant and its unit test only), `src/agents.rs` (`render_ignore`, `STATE_KEPT` use and `ignore_block` only), `src/project_update.rs` (the `now hidden` fact only), `tests/gitignore_block_cli.rs`, `tests/fixtures/protocol/v1/current/16-agents-plan-create.json`, `tests/fixtures/protocol/v1/current/17-agents-apply-ok.json` (the block lines change), `docs/design-review/appendix/state-layout.md`; only if a red run shows a pin flipping: `tests/project_update_plan_cli.rs`, `tests/update_end_to_end_cli.rs`, `tests/fixtures/state_folder_literals/allow.tsv`
- read-only: `src/go_state.rs`, `src/go_residue.rs`, `src/generation.rs`, `src/decision.rs`, `.moai/specs/SPEC-20261120-moai-layout/spec.md`
- forbidden: changing the `KEPT` names, any card number in shipped text, any write outside the project folder, real HOME or real project writes in tests, deleting any file, `skills/**`

## Plan

1. 사용자 승인 뒤 RED: `tests/gitignore_block_cli.rs`에 REQ-8의 진짜 git 테스트와 REQ-7의 스캔 테스트를 먼저 써요. 지금 블록은 폴더째 풀어서 평평한 `workflows/w1.md`가 보이므로 AC-2가 실패해야 해요.
2. `state_root.rs`에 허용 줄 목록을 `KEPT` 옆에 두어요(이름만 읽는 소비자는 그대로). `render_ignore`가 목록의 `!` 줄 앞에 하위 경로를 가리는 무시 줄을 넣어요. 깃 패턴의 순서(무시 → 다시 허용)와 `workflows/`는 "폴더만" 허용하는 끝 `/` 패턴으로 쓰고, 실제 줄은 진짜 `git check-ignore`로 확인해요.
3. `ignore_block`: 추적 중인데 새 블록이 가릴 경로를 `git ls-files`로 읽어 정확한 `!` 줄을 더하고 계획에 `now hidden`을 내요. 읽기 실패는 줄을 더하는 쪽(보이는 쪽)으로 실패해요.
4. 골든 둘(`16`, `17`)을 `MOAI_UPDATE_FIXTURES=1`로 다시 만들고 블록 줄과 `rendered_sha256`만 달라졌는지 읽어요. `state-layout.md`에 허용 줄 표를 더해요.
5. 검증: `cargo fmt --all`, `moai check affected --card` 맨몸으로, 새 테스트 파일이 없으면 `home_isolation_scan`·`state_folder`·`source_manifest`·`repo_scans_table`은 `check affected`가 고르는 대로 따라가요. 가드 테스트마다 뮤턴트 한 줄(`now hidden`의 정확한 `!` 줄, 하위 경로를 가리는 무시 줄).

## Risks

- 새 Rust 쓰기 경로가 `manifests/<하위폴더>/`처럼 허용 줄에 없는 모양이면 조용히 무시돼요. REQ-7의 스캔이 막지만, 문자열로 만든 경로는 스캔이 못 읽어요(확인된 한계예요).
- 이미 무시 줄이 있는 프로젝트의 블록은 다음 `moai update`가 다시 쓸 때까지 예전 줄이에요(블록은 커밋하는 파일이에요).
- `.gitignore`의 `!`와 `*` 순서가 틀리면 부모 폴더가 무시된 채 자식 `!`가 안 먹어요. 그래서 문자열이 아니라 진짜 git으로 테스트해요(REQ-8).
- `classify`와 `kept_name`은 건드리지 않아서 이 SPEC의 변경이 세대 분류에 영향을 주지 않아요. 영향이 있으면 기본안(B) 밖이에요.
- Go 세대 프로젝트에서 `manifests/`나 `decisions/`에 둔 옛 파일이 실제로 있는지는 확인하지 못했어요(Facts의 "찾지 못했어요").

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `config/`·`docs/`는 폴더째 허용을 그대로 둬요(a). 대안 b: 허용을 빼고, 이미 추적하는 파일만 정확한 `!` 줄로 지켜요 | 레이아웃 SPEC REQ-1이 팀 소유로 적었고 `docs`는 go-state가 복원해요. Rust 소유 하위 경로가 없어서 좁힐 기준이 없어요. 사용자 선택 필요 | 줄 두 개 |
| D-2 | 좁히는 범위: A 안 좁히고 문서만 남겨요 · B 블록 줄만 좁혀요(기본안) · C `KEPT` 자체를 하위 경로 목록으로 바꿔 분류기와 `kept_name`까지 | B는 소비자 넷 중 하나(블록)만 바뀌어서 카드 기준 1의 "일치"가 이름 목록 하나로 지켜져요. C는 세대 분류와 go-state 되돌리기를 같이 바꿔 위험이 커요. 사용자 선택 필요 | B→A는 블록 줄 되돌리기 |
| D-3 | 추적 중인 파일은 정확한 `!` 줄로 지키고 계획이 `now hidden`을 알려요(REQ-5, REQ-6) | 카드 기준 1의 "기존 추적 파일이 사라지지 않음", C-300 AC-14와 같은 방식 | 줄 규칙 |
| D-4 | 새 테스트 파일을 만들지 않고 `tests/gitignore_block_cli.rs`에 둬요 | 새 테스트 파일은 `home_isolation_scan`·`state_folder`·`source_manifest`·`repo_scans_table`을 더 건드려요. 레인 기본값 | 파일 분리 |

## Result

구현 커밋은 `ee4a7bb75f`예요. 승인은 사용자의 직접 선택(원장 노트 10722)이고 D-1 (a), D-2 B예요.

- AC-1: met — `state_root::tests::the_narrow_allow_list_leaves_the_first_level_names_alone`가 `KEPT_NARROW` 표(manifests `*.json`, workflows `*/`, decisions `*.jsonl`)와 `KEPT`의 열한 이름·`is_kept` 답이 그대로임을 고정해요. 통과했어요.
- AC-2: met — `the_block_allows_only_the_paths_rust_writes_under_manifests_workflows_decisions`가 진짜 `git check-ignore`로 Rust 경로는 보이고 평평한 Go 모양 파일은 무시됨을 확인해요. 통과했어요.
- AC-3: met — `a_tracked_file_the_narrow_block_would_hide_keeps_an_exact_allow_line`가 추적 중인 `manifests/x.txt`, `manifests/sub/y.json`, `workflows/w1.md`가 정확한 `!` 줄(윗 폴더 줄 포함)로 보이고 계획이 `now hidden:`을 내며 `git ls-files`에 남음을 확인해요. 통과했어요. 스무 개 초과는 `the_plan_counts_the_hidden_files_past_twenty`, 이전 폴더 `.moai-cli/`에서 옮겨 오는 파일은 `a_tracked_file_in_the_earlier_folder_stays_visible_after_the_move`가 확인해요.
- AC-4: met — `every_state_path_a_writer_spells_is_allowed_by_the_block`가 소스가 적는 경로를 읽어 모두 허용됨을 확인하고, 만든 경로 `manifests/sub/x.json`은 무시됨을 확인해요. 통과했어요.
- AC-5: met — `the_narrow_block_is_the_same_in_plan_and_apply_and_idempotent`가 계획과 적용이 같은 여섯 줄을 내고(허용 줄 → 무시 줄 → 다시 허용 순서) 두 번째 적용이 파일을 바꾸지 않음을 확인해요. 통과했어요.

실행: `gitignore_block_cli` 55개와 `state_root` 단위 테스트가 모두 통과했어요(슬롯 `--max 2`, nextest). 골든 `16`·`17`은 `MOAI_UPDATE_FIXTURES=1`로 다시 만들었고 블록 줄(50→56)과 `rendered_sha256`, `manifest_sha256`만 달라졌어요.

바뀐 다른 핀(이 카드 범위의 테스트 파일 안): 블록 줄 수 `51 added`→`57 added` 둘, 팀 파일 목록과 SPEC 테스트의 평평한 `workflows/run.md`·`decisions/D-1.md`를 새 모양(`workflows/run/SKILL.md`, `decisions/D-1.jsonl`)으로.

범위 밖 변경(승인된 Scope 글에 없는 경로, 리더 결정 10756·10758): 소스 스캔 테스트(REQ-7)가 `src/`를 읽어서 `tests/gitignore_block_cli.rs`가 읽기 표의 "읽는 테스트"가 됐어요. ① `tools/repo_scans.json`에 `reads-gitignore-block-cli` always 스캔 하나를 줄 리스로 더했어요(규칙 파일 이 테스트와 `tests/repo_scans_table.rs`, scope `src/`와 `.moai/evals/audit/cases-tuning.json`). ② `tests/repo_scans_table.rs`에 `FILE_READ_EXEMPT` 줄 하나(`tests/gitignore_block_cli.rs`와 임시 프로젝트용 이름 여덟 개에만). ③ `tests/fixtures/state_folder_literals/allow.tsv`의 두 행: `tests/gitignore_block_cli.rs` 226/81→272/92(이 카드가 새 테스트에 적은 글자), `tests/repo_scans_table.rs` 0/18→4/20(면제 줄이 적는 이름). 스캔 scope와 규칙은 느슨해진 데가 없어요.

SPEC에 없던 한 가지: 사용자가 쓴 규칙이 이미 `manifests/x.json` 같은 모양을 무시하면 그 폴더의 다시 허용하는 줄을 빼요(블록은 사용자 규칙의 효과를 바꾸지 않는다는 기존 약속, B1). 이를 `a_user_rule_keeps_its_effect_and_the_preview_names_it`가 고정해요. `src/project_update.rs`는 건드리지 않았어요.

guard-mutant: the_block_allows_only_the_paths_rust_writes_under_manifests_workflows_decisions | the `<state>/<folder>/*` ignore line of `render_ignore` | exit 100
guard-mutant: a_tracked_file_the_narrow_block_would_hide_keeps_an_exact_allow_line | the exact `!` line written for each `narrow.exact` path | exit 100
guard-mutant: without_git_the_block_keeps_the_whole_folder_lines | the git-missing fallback to the whole-folder block (`Ok(off)` made `Ok(on)`) | exit 100
guard-mutant: the_plan_counts_the_hidden_files_past_twenty | the `now hidden: and N more files` line | exit 100
guard-mutant: a_user_rule_keeps_its_effect_and_the_preview_names_it | the `skip_allow_back` check before the allow-back line | exit 100
guard-mutant: the_narrow_allow_list_leaves_the_first_level_names_alone | the `decisions` pattern `*.jsonl` of `KEPT_NARROW` changed to `*` | exit 100
guard-mutant: a_tracked_file_in_the_earlier_folder_stays_visible_after_the_move | the earlier-folder prefix in `narrowing` | exit 100
guard-mutant: every_state_path_a_writer_spells_is_allowed_by_the_block | the `workflows` pattern `*/` of `KEPT_NARROW` changed to `*.md` | exit 100
guard-mutant: the_narrow_block_is_the_same_in_plan_and_apply_and_idempotent | the allow-back line moved before the ignore line | exit 100

한계: 새 Rust 쓰기 경로가 허용 줄에 없는 모양이면 조용히 무시돼요(스캔은 문자열로 만든 경로를 못 읽어요). 이미 블록이 있는 프로젝트는 다음 `moai update`까지 예전 줄이에요.
