---
id: SPEC-20261120-moai-layout
title: "One .moai/ layout rule for every project: tracked team files, ignored .moai/local/, project tools in .moai/tools/"
status: draft
revision: 6
created: 2026-11-20
---

# One .moai/ layout rule for every project

## Intent

사용자 결정 2026-11-20(리더 대화, 선택): 모든 MoAI 프로젝트가 같은 `.moai/` 구조를 써요. 설계 보고서는 `~/.moai/reports/moai-adk/moai-layout-and-worktree-disk-design-2026-11-20.md` 4.1 이에요.

```
.moai/
  moai.json  project.yaml                 표지·프로필 (추적)
  config/ decisions/ docs/ specs/ manifests/ workflows/   (추적)
  tools/                                  프로젝트 전용 개발 지원 스크립트 (추적, 새로)
  evals/                                  평가 세트 (추적, 필요할 때)
  local/                                  이 기계의 것 (무시): receipts/ backups/ index/ logs/ state/ cache/ …
```

까닭: 무시되는 것은 worktree 에 복제되지 않아요(추적 파일은 worktree 를 만들 때마다 체크아웃돼요). 그리고 MoAI 와 에이전트가 만드는 개발 지원 파일이 프로젝트 루트(`scripts/` 등)에 프로젝트 코드와 섞이지 않아요. 호스트가 자리를 정한 파일(`AGENTS.md`, `.claude/`, `.codex/`, `.github/ISSUE_TEMPLATE/moai-*`)은 그 자리에 둬요.

이미 있는 프로젝트는 계속 돌아가요: 읽기는 옛 경로까지, 쓰기는 새 경로로.

이 SPEC 은 규칙의 **조각 1** 이에요: 규칙의 정의(코드 한 곳)와 관리되는 `.gitignore` 블록. 규칙의 나머지는 다음 조각의 SPEC 이 맡아요(SPEC 하나는 자기 AC 를 모두 `met` 으로 닫아야 해서 조각마다 SPEC 을 따로 둬요): 쓰는 쪽을 종류별로 `local/` 로 옮기는 SPEC 들. 개정 4(2026-11-24)는 같은 카드의 조각 3 을 이 SPEC 에 더해요: 만들어지는 관리 블록 한 줄과 `agents/impl.md` 한 줄이 `.moai/tools/` 를 말하게 하고(REQ-15, REQ-16), 업데이트 계획의 `now visible` 목록이 상태 폴더 아래 중첩 `.gitignore` 를 따르게 해요(REQ-17). 개정 5(2026-11-24)는 개정 4 감사(FAIL, 카드 노트 8184)의 수리를 더해요: 읽는 중첩 `.gitignore` 의 경계(64 KiB, 100개)와 링크 거절을 못 박고(REQ-17), 읽지 않은 `.gitignore` 아래 파일은 부모 규칙이 숨기더라도 이름을 말하게 하며(REQ-18), 자격 증명 목록에 `credentials.json` 을 더해요(REQ-11). 개정 6(2026-11-26)은 남은 카드 기준 둘을 더해요: 조사한 비밀 파일 이름을 블록에 더하고(REQ-19, `## Research`), 중첩 `.gitignore` 를 여는 순간에도 링크를 따르지 않는다는 것을 시험으로 못 박아요(D-11).

## Interview

- none: 사용자에게 묻지 않았어요. 설계는 2026-11-20 사용자 결정(카드 본문)을 따르고, 개정 4 의 선택은 카드 노트 8139, 개정 5 의 선택은 카드 노트 8184 의 리더 결정(사용자 위임 2026-11-23), 개정 6 의 선택(조사해서 이름을 더함)은 카드 노트 10187 의 리더 결정이에요.

## Non-goals

- 쓰는 쪽을 `local/` 로 옮기지 않아요: 종류마다 다음 SPEC 이 해요(`receipts` → `index` → `state`·`forge/outbox` → `backups` 등). 이 SPEC 은 그 길(함수)만 줘요.
- 안내 문구는 관리 블록 한 줄(REQ-15)과 `agents/impl.md` 한 줄(REQ-16)만 바꿔요. 스킬 본문(`skills/**`)과 그 밖의 에이전트·워크플로 문구, 사용자 홈 규칙 파일은 바꾸지 않아요: 다음 SPEC 이 해요.
- `.moai-cli` → `.moai` 이전 마무리(미리보기·토큰·되돌리기, 두 폴더 공존 해소)는 하지 않아요: 따로 다뤄요.
- 보고서·증거 위치는 정하지 않아요: 프로젝트 밖 `~/.moai/reports/<project>/` 이고 다른 SPEC 이 다뤄요.
- moai-adk 자체의 루트 `tools/`(제품 개발 코드)는 옮기지 않아요.
- 호스트 소유 파일의 자리는 바꾸지 않아요.
- 옛 경로의 파일을 지우거나 옮기지 않아요(읽기 호환만).

## Requirements

- REQ-1: The state layout shall name the kept kinds `project.yaml`, `moai-cli.json`, `moai.json`, `manifests`, `workflows`, `config`, `decisions`, `docs`, `specs`, `tools` and `evals`, and every other kind is local to the machine and lives under `local/`.
- REQ-2: The state layout shall give one function that returns the folder of a local kind under the state folder's `local/` and one that returns the folder where an earlier project kept it, so no writer spells either path itself.
- REQ-3: The managed `.gitignore` block shall ignore the state folder's `local/` explicitly and allow `tools/` and `evals/` beside the earlier kept list.
- REQ-4: When a project's managed `.gitignore` block lacks the new lines, the project update plan shall show the block as changed.
- REQ-5: When a reader looks for a local kind, the state layout shall return `local/<kind>` first and the earlier `<state>/<kind>` when the new one is absent.
- REQ-6: The state layout shall give writers the path `local/<kind>` for a local kind.
- REQ-7: The layout shall leave the host-owned files `AGENTS.md`, `.claude/`, `.codex/` and `.github/ISSUE_TEMPLATE/moai-*` where the host reads them.
- REQ-8 (REQ-4): IF the user has not applied the project update plan, THEN the project update shall leave the managed `.gitignore` block unchanged.
- REQ-9 (REQ-5): IF neither the new nor the earlier path exists, THEN the state layout shall report the kind as absent.
- REQ-10: The state layout shall give writers a path that leaves an earlier `<state>/<kind>` untouched.
- REQ-11: The managed `.gitignore` block shall ignore the files `credentials.yaml`, `credentials.json`, `*.credentials`, `.env` and `.env.*` anywhere under the state folder, `tools/` and `evals/` included, with its lines after the allow-list.
- REQ-12: When a project update would lift the ignore from untracked files under the state folder's `tools/` or `evals/`, the project update plan shall name each such file as `now visible: <path>`, up to twenty, and count the rest in one line.
- REQ-13 (REQ-12): IF git is not available, THEN the project update plan shall name no file and say once that git was not consulted.
- REQ-14 (REQ-5): IF the new path or the earlier path is a link whose target is missing, THEN the state layout shall still return that path as the one that is there.
- REQ-15: The managed block of the generated `CLAUDE.md` shall name `<state>/tools/` on one line, within the `BLOCK_MAX_LINES` lines of the block, as the place for the project's development scripts in the instruction languages `en` and `ko`.
- REQ-16: The `moai-impl` agent definition shall tell the agent to put a development-support script the task creates under `.moai/tools/`.
- REQ-17 (REQ-12): WHILE a `.gitignore` file under the state folder ignores a file that the project update would lift from the root rules, and that `.gitignore` is a regular file of at most 65536 bytes with no link on its path and is among the first 100 such files in path order, the project update plan shall not name that file as `now visible`.
- REQ-18 (REQ-17): IF a `.gitignore` file under the state folder is a link, has a link on its path, is larger than 65536 bytes, cannot be read, or is past the first 100, THEN the project update plan shall name every file below its folder as `now visible` even where a rule of a parent folder ignores it, and say once that a `.gitignore` was not read.
- REQ-19: The managed `.gitignore` block shall also ignore the files `.git-credentials`, `.netrc`, `_netrc`, `.pgpass`, `pgpass.conf`, `.pypirc`, `.credentials.json`, `credentials.tfrc.json`, `application_default_credentials.json`, `credentials.toml`, `.vault-token`, `.mylogin.cnf`, `master.key`, `keystore.properties`, `id_rsa`, `id_dsa`, `id_ecdsa`, `id_ecdsa_sk`, `id_ed25519`, `id_ed25519_sk`, `id_mldsa44_ed25519`, `*.ppk`, `*.tfstate`, `*.tfstate.*`, `.dev.vars`, `.dev.vars.*`, `.aws/credentials`, `.cargo/credentials`, `.gem/credentials` and `.streamlit/secrets.toml` anywhere under the state folder, `tools/` and `evals/` included, with its lines after the allow-list.

## Acceptance

- AC-1 (REQ-1, REQ-2): 레이아웃 정의가 추적 종류 열한 개를 이름으로 대고 나머지는 `local/` 아래라고 하며, 지금 있는 기계 출력 종류 목록을 갖고, 새 경로와 옛 경로를 한 함수씩 돌려줘요.
  verify: `cargo test --locked --offline --lib state_root::tests::the_layout_names_kept_kinds_and_local_paths`
- AC-2 (REQ-3): 새 프로젝트의 관리 블록은 `<state>/local/` 줄을 갖고 `tools/`·`evals/` 허용 줄을 `specs/` 뒤에 갖고, git 이 `tools/`·`evals/` 의 파일은 무시하지 않고 `local/` 안은 무시한다고 답해요.
  verify: `cargo test --locked --offline --test gitignore_block_cli the_block_ignores_local_and_keeps_tools_and_evals`
- AC-3 (REQ-4, REQ-8): 옛 블록이 있는 프로젝트의 업데이트 계획은 새 줄 서른여덟(`local/`, `tools/`·`evals/` 허용, 자격 증명·`.env` 서른다섯)이 더해지는 `merge` 로 보여 주고, 적용 전에는 파일이 그대로예요.
  verify: `cargo test --locked --offline --test gitignore_block_cli an_earlier_block_is_shown_as_changed_and_left_until_applied`
- AC-4 (REQ-5, REQ-6, REQ-9, REQ-10): 읽는 길은 `local/<kind>` 를 먼저, 없으면 옛 경로를, 둘 다 없으면 없음을 돌려주고, 쓰는 길은 새 경로를 주어 옛 폴더를 건드리지 않아요.
  verify: `cargo test --locked --offline --lib state_root::tests::a_reader_falls_back_to_the_earlier_path_and_a_writer_uses_the_new_one`
- AC-5 (REQ-7): 호스트 소유 파일 네 곳은 추적 종류에도 기계 출력 종류에도 없어요.
  verify: `cargo test --locked --offline --lib state_root::tests::host_owned_files_are_not_in_the_layout`
- AC-6 (REQ-5, REQ-14): 끊어진 링크인 `local/<kind>` 는 있는 것으로 세어 옛 폴더로 건너뛰지 않고, 끊어진 옛 링크도 없음이 아니라 그 경로로 돌려줘요.
  verify: `cargo test --locked --offline --lib state_root::tests::a_broken_link_counts_as_the_path_being_there`
- AC-7 (REQ-11): git 이 `tools/`·`evals/`·그 밖의 상태 폴더 아래 어디서나 `credentials.yaml`·`credentials.json`·`*.credentials`·`.env`·`.env.*` 는 무시하고 `check.sh`·`env.sh`·`set.json` 은 무시하지 않으며, 그 줄들이 허용 줄 뒤에 와요.
  verify: `cargo test --locked --offline --test gitignore_block_cli the_block_ignores_credential_and_env_files_even_in_tools_and_evals`
- AC-8 (REQ-12): 업데이트가 `tools/`·`evals/` 의 무시를 푸는 프로젝트의 계획이 새로 보이게 될 파일마다 `now visible: <경로>` 를 말하고, 자격 증명·`.env` 파일은 말하지 않으며, 적용 뒤의 계획은 말하지 않아요.
  verify: `cargo test --locked --offline --test gitignore_block_cli the_plan_names_the_files_the_update_makes_visible`
- AC-9 (REQ-12): 사용자 자신의 규칙이 계속 무시하는 파일은 `now visible` 로 말하지 않아요.
  verify: `cargo test --locked --offline --test gitignore_block_cli a_file_the_users_own_rule_ignores_is_not_named_as_visible`
- AC-10 (REQ-12): 계획은 파일 스무 개까지 이름을 말하고 나머지는 한 줄로 세어요: 스무 개는 모두 말하고 스물한 개는 스무 개와 `and 1 more files` 줄이에요.
  verify: `cargo test --locked --offline --test gitignore_block_cli the_plan_counts_the_visible_files_past_twenty`
- AC-11 (REQ-13): git 이 없으면 계획은 파일 이름을 말하지 않고 git 을 보지 않았다는 한 줄을 한 번만 말해요.
  verify: `cargo test --locked --offline --test gitignore_block_cli no_repository_and_no_git_are_not_failures`
- AC-12 (REQ-15): `en` 과 `ko` 로 만든 관리 블록은 `.moai/tools/` 를 말하는 줄을 한 줄 갖고 `scripts/` 와 프로젝트 루트를 스크립트 자리로 말하지 않으며, 줄 수가 `BLOCK_MAX_LINES`(15) 이내예요.
  verify: `cargo test --locked --offline --lib agents::tests::the_block_names_the_tools_folder_for_project_scripts`
- AC-13 (REQ-16): `agents/impl.md` 는 `.moai/tools/` 를 말하고 `scripts/` 와 프로젝트 루트를 스크립트 자리로 말하지 않아요.
  verify: `cargo test --locked --offline --lib agents::tests::the_impl_agent_names_the_tools_folder_for_project_scripts`
- AC-14 (REQ-17): 상태 폴더 아래 중첩 `.gitignore`(예: `.moai/tools/.gitignore` 의 `private/`)가 계속 무시하는 파일은 `now visible` 로 말하지 않고, 그 규칙이 없는 옆 파일은 말해요.
  verify: `cargo test --locked --offline --test gitignore_block_cli a_nested_gitignore_under_the_state_folder_keeps_a_file_out_of_the_visible_list`
- AC-15 (REQ-17, REQ-18): 크기가 65536 바이트인 중첩 `.gitignore` 는 읽혀서 그 규칙이 막는 파일을 말하지 않고, 65537 바이트인 것은 읽히지 않아 그 폴더 아래 파일을 말하고 한 줄로 알려요.
  verify: `cargo test --locked --offline --test gitignore_block_cli a_nested_gitignore_of_64_kib_is_read_and_one_byte_more_is_not`
- AC-16 (REQ-17, REQ-18): 중첩 `.gitignore` 는 경로 순서로 백 개까지 읽혀요: 백 개째 폴더의 규칙은 듣고, 백한 개째 폴더 아래 파일은 말하고 한 줄로 알려요.
  verify: `cargo test --locked --offline --test gitignore_block_cli the_hundredth_nested_gitignore_is_read_and_the_hundred_first_is_not`
- AC-17 (REQ-18): 링크인 중첩 `.gitignore`(프로젝트 밖의 `*.sh` 규칙을 가리킴)는 읽지 않아서 그 폴더의 `x.sh` 를 말하고 한 줄로 알려요.
  verify: `cargo test --locked --offline --test gitignore_block_cli a_nested_gitignore_that_is_a_link_is_not_read`
- AC-18 (REQ-18): 경로 중간의 폴더가 링크인 중첩 `.gitignore`(git 이 색인으로 아는 `tools/lp/.gitignore` 의 `lp` 가 프로젝트 밖을 가리키는 링크)는 읽지 않고 한 줄로 알려요.
  verify: `cargo test --locked --offline --test gitignore_block_cli a_nested_gitignore_behind_a_linked_folder_is_not_read`
- AC-19 (REQ-18): 읽을 수 없는(mode 000) 중첩 `.gitignore` 는 읽지 않아서 그 폴더 아래 파일을 말하고 한 줄로 알려요(유닉스에서만).
  verify: `cargo test --locked --offline --test gitignore_block_cli an_unreadable_nested_gitignore_is_not_read`
- AC-20 (REQ-18): 읽지 않은 자식 `.gitignore` 가 부정 규칙(`!keep.tmp`)을 가졌고 읽은 부모 규칙(`*.tmp`)이 그 파일을 숨길 때도 그 파일을 말하고, 적용 뒤 git 이 그 파일을 무시하지 않는다는 답과 같아요.
  verify: `cargo test --locked --offline --test gitignore_block_cli a_file_below_an_unread_gitignore_is_named_even_when_a_parent_rule_hides_it`
- AC-21 (REQ-19): git 이 `tools/`·`evals/`·상태 폴더 아래에서 REQ-19 의 이름마다 하나씩 만든 파일을 무시하고, 이웃 이름(`id_rsa.pub`, `deploy.pem`, `server.key`, `credentials`, `secrets.yaml`, `.npmrc`, `cert.p12`, `prod.tfvars`, `config.json`)은 무시하지 않으며, 그 줄들이 허용 줄 뒤에 와요.
  verify: `cargo test --locked --offline --test gitignore_block_cli the_block_ignores_the_researched_secret_names`
- AC-22 (REQ-17, REQ-18): 중첩 `.gitignore` 를 여는 순간에 파일이나 부모 폴더가 링크(프로젝트 안이나 밖을 가리킴)로 바뀌어 있으면 열지 않고, 그 자리의 파이프는 기다리지도 읽지도 않으며, 보통 파일은 열어 읽어요(시험 `a_file_swapped_for_a_link_inside_the_root_is_not_opened`, `a_file_swapped_for_a_link_to_outside_the_root_is_not_opened`, `a_parent_swapped_for_a_link_inside_the_root_is_not_opened`, `a_parent_swapped_for_a_link_to_outside_the_root_is_not_opened`, `a_pipe_is_neither_waited_on_nor_read`, `a_plain_nested_file_is_opened_and_read`, 유닉스에서만).
  verify: `cargo test --locked --offline --lib agents::tests::`

## Scope

- write: `.moai/specs/SPEC-20261120-moai-layout/spec.md`, `src/state_root.rs`, `src/agents.rs`(관리 블록: `STATE_KEPT` 와 `render_ignore`), `tests/gitignore_block_cli.rs`, 골든 `tests/fixtures/protocol/v1/current/16-agents-plan-create.json`·`17-agents-apply-ok.json`(블록 줄이 바뀌어서), `docs/design-review/appendix/state-layout.md`; 개정 4 가 더하는 것: `agents/impl.md`(Boundaries 한 줄). 블록 문구를 단언하는 `tests/language_setting.rs`·`tests/language_three_questions.rs`·`tests/product_cli.rs` 는 시험이 깨질 때만 한 줄씩 `moai card line-lease` 로 빌려요
- read-only: `src/layout.rs`(P09 홈 배치와는 별개예요), `src/go_state.rs`(옛 이름 목록)
- forbidden: 쓰는 쪽 모듈 바꾸기, 옛 경로 파일 이동·삭제, `.moai-cli` 이전 변경, 보고서 위치 변경, 관리 블록 한 줄과 `agents/impl.md` 한 줄 밖의 안내 문구(스킬 본문 `skills/**`, 다른 에이전트·워크플로 문구, 사용자 홈 규칙 파일) 바꾸기

## Plan

1. SPEC 을 커밋해요(이 문서). `spec check` 를 돌려요.
2. `src/state_root.rs` 에 `KEPT`(열한 종류)·`LOCAL_KINDS`·`HOST_OWNED`·`local_path`·`earlier_path`·`write_path`·`read_path`·`is_kept` 를 두고 단위 시험 셋(`the_layout_names_kept_kinds_and_local_paths`, `a_reader_falls_back_to_the_earlier_path_and_a_writer_uses_the_new_one`, `host_owned_files_are_not_in_the_layout`)으로 못 박아요.
3. `src/agents.rs` 의 `STATE_KEPT` 를 `state_root::KEPT` 로 바꾸고 블록에 `<state>/local/` 줄을 더해요(허용 줄 순서는 앞의 아홉 줄 그대로, `tools/`·`evals/` 가 `specs/` 뒤). `gitignore_block_cli` 시험 둘을 더하고, 줄 수를 못 박은 기존 시험 둘(12줄 → 15줄)을 고쳐요.
4. 골든 두 개를 `MOAI_UPDATE_FIXTURES=1` 로 다시 만들고 `git diff` 가 블록 줄과 해시뿐인지 봐요.
5. 확인: `moai check affected origin/v2026.11.0..HEAD`(내가 빌드한 moai), 변이(블록 줄 빼기·허용 줄 빼기·`read_path` 순서 바꾸기·`is_kept` 깨기)가 시험에 잡히는지, SPEC `spec check` 와 `--closure`.
6. (개정 4) `render_agents` 가 `{state}/tools/` 한 줄(`en`·`ko`)을 검사 줄 뒤에 더하고, `agents::tests::the_block_names_the_tools_folder_for_project_scripts` 가 두 언어의 줄 수(`BLOCK_MAX_LINES` 이내)와 `scripts/`·프로젝트 루트 문구 없음을 못 박아요. 블록을 단언하는 시험과 골든 16·17 은 한 줄·해시만 바뀌어요.
7. (개정 4) `agents/impl.md` Boundaries 에 한 줄을 더하고 `agents::tests::the_impl_agent_names_the_tools_folder_for_project_scripts`(`include_str!`)로 못 박아요.
8. (개정 4) `newly_visible` 이 상태 폴더 아래 중첩 `.gitignore` 파일을 읽어 임시 저장소의 같은 상대 경로에 두고 묻게 해요(링크는 따르지 않고, 크기 상한을 둬요). `a_nested_gitignore_under_the_state_folder_keeps_a_file_out_of_the_visible_list` 를 더해요.
9. (개정 4) 각 새 시험에 변이(블록 줄 빼기, `impl.md` 줄 빼기, 중첩 규칙 읽지 않기)를 한 번씩 돌려 실패하는지 보고, `check affected` 를 한 번 돌려요.
10. (개정 5) 시험 먼저: `gitignore_block_cli` 에 AC-15 에서 AC-20 의 시험과 `credentials.json` 줄 기대를 더하고 구현 전에 올바른 이유로 실패하는지 봐요.
11. (개정 5) `nested_ignore_files` 가 경로의 모든 조각이 링크가 아닌지 보고(마지막 조각만이 아니라), 읽지 않은 `.gitignore` 의 폴더 목록을 `newly_visible` 에 돌려줘 그 아래 후보를 규칙 판정과 상관없이 말하게 해요. `STATE_SECRETS` 에 `**/credentials.json` 을 더하고 골든 16·17 은 `MOAI_UPDATE_FIXTURES=1` 로 다시 만들어요.
12. (개정 5) 변이(상한 제거, 상한 100 → 2, `symlink_metadata` → `metadata`, 마지막 조각만 보기, 읽지 않은 폴더 아래 후보를 규칙에 맡기기, `credentials.json` 줄 빼기)를 시험이 잡는지 보고 `check affected` 를 한 번 돌려요.
13. (개정 6) 시험 먼저: 여는 단계의 단위 시험(링크로 바뀐 파일·부모 폴더, 파이프, 보통 파일)을 쓰고 이전 여는 방식에서 실패하는지 본 뒤 D-11 로 고쳐요.
14. (개정 6) 승인 뒤: `the_block_ignores_the_researched_secret_names` 와 줄 수 기대를 먼저 고쳐 실패를 보고, `STATE_SECRETS` 에 REQ-19 의 서른 줄을 더해요. 골든 16·17 은 `MOAI_UPDATE_FIXTURES=1` 로 다시 만들고 `git diff` 가 블록 줄과 해시뿐인지 봐요. `docs/design-review/appendix/state-layout.md` 를 맞춰요.
15. (개정 6) 변이(새 줄 빼기, 폴더 붙인 줄을 이름만으로 넓히기, 새 줄을 허용 줄 앞으로 옮기기)를 시험이 잡는지 보고 `check affected` 를 한 번 돌려요.

## Risks

- 업데이트로 `tools/`·`evals/` 의 무시가 풀리면 그 안에 이미 있던 사용자 파일(예: `.moai/tools/.env`)이 git 에 새로 보여요(감사 N1). 계획이 `now visible: <경로>` 로 이름을 말하고, 자격 증명·`.env` 류는 블록의 새 줄이 계속 무시해요. 이름이 스무 개를 넘으면 앞 스무 개와 나머지 수예요. `*.key`, `*.pem`, `*.p12` 같은 넓은 확장자와 `secrets.*` 는 이 줄들이 막지 않아요(`## Research`): 목록에 더하는 것은 한 줄이에요.
- (개정 6) `tools/`·`evals/` 에 같은 이름의 가짜 파일(시험 픽스처의 `id_rsa`, `x.tfstate` 등)을 일부러 두는 팀은 그 파일이 git 에 나타나지 않아요. 이미 추적 중인 파일은 그대로예요. 자기 `.gitignore` 의 관리 블록 뒤에 `!` 줄을 더하면 다시 보여요(관리 블록보다 앞의 `!` 줄은 블록이 이겨요).
- (개정 6) 대소문자가 다른 이름(`ID_RSA`, `X.PPK`)이 macOS·Windows 에서 걸리는지는 확인하지 않았어요.
- `.env.*` 줄은 `.env.example` 도 무시해요: 예제 파일을 `tools/` 에 두려는 사용자는 자기 `.gitignore` 에 `!.moai/tools/.env.example` 을 더해야 해요(안전한 쪽으로 틀려요).

- 관리 블록의 줄이 바뀌어서 이미 블록이 있는 프로젝트는 다음 업데이트 계획에서 "블록이 바뀐다"를 보게 돼요: 적용은 사용자가 계획을 확인한 뒤에만이에요. 골든 두 개(`16-agents-plan-create`, `17-agents-apply-ok`)와 `gitignore_block_cli` 의 줄 수 기대 둘이 함께 바뀌어요.
- 모든 상태가 이미 `<state>/*` 규칙으로 무시되고 있어서(허용 목록 밖은 무시) `local/` 줄은 읽는 사람에게 규칙을 말해 주는 줄이에요. 무시 결과는 달라지지 않아요.
- 함수만 있고 쓰는 쪽은 아직 옛 경로에 써요: 그 사이에 `local/` 은 비어 있고 `read_path` 는 옛 경로를 돌려줘요. 쓰는 쪽을 옮기는 SPEC 이 종류마다 되돌림·영수증 시험을 둬요.
- `LOCAL_KINDS` 는 지금 있는 종류의 목록이고 `go_state` 의 옛 이름 목록(`STATE_ENTRIES`)과 따로 있어요: 둘의 관계를 못 박는 시험은 `go_state` 를 쓰는 다음 SPEC 이 더해요(이 SPEC 은 그 파일을 쓰지 않아요).
- (개정 4) `agents/impl.md` 의 한 줄은 에이전트가 새로 만드는 개발 지원 스크립트에만 해당해요. 이 저장소의 루트 `tools/` 는 제품 개발 코드이고 그대로예요(D-5): 줄은 "이 작업이 새로 만드는 스크립트"로 한정해서 쓰고, 기존 `tools/` 파일을 옮기라는 뜻으로 읽히지 않게 해요.
- (개정 4) 중첩 `.gitignore` 를 임시 저장소에 옮겨 올 때 링크를 따라가면 프로젝트 밖 파일을 읽을 수 있어요: 링크인 `.gitignore` 와 링크 폴더는 읽지 않고, 한 파일 크기에 상한을 둬요.
- (개정 5) 링크 검사는 경로의 모든 조각에 해요: 색인에 있는 `tools/lp/.gitignore` 의 `lp` 가 나중에 프로젝트 밖을 가리키는 링크로 바뀌어도 그 파일은 읽지 않아요. 읽지 않은 `.gitignore`(링크, 링크 뒤, 65536 바이트 초과, 읽을 수 없음, 백 개 뒤)의 폴더 아래 파일은 부모 규칙이 숨기더라도 `now visible` 에 이름을 말해요(REQ-18): 읽지 않은 파일이 `!` 규칙으로 그 파일을 다시 보이게 할 수 있어서 이름을 빼면 적용 뒤 git 에 나타나는 파일을 계획이 못 말하기 때문이에요. 안전한 쪽(말하는 쪽)으로 틀리고, 값은 보이지 않아요.
- `.moai-cli` 와 `.moai` 가 함께 있는 프로젝트(이 저장소 포함)에서 `local/` 은 `state_root::dir` 가 정하는 쪽 아래예요. 이전 마무리는 이 SPEC 밖이에요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 추적 종류는 열한 개 이름을 코드 한 곳에 두고, 그 밖의 모든 종류는 `local/` 아래로 본다 | 사용자 결정: "추적은 팀 파일, 나머지는 이 기계의 것" — 목록이 닫혀 있어야 새 종류가 실수로 추적되지 않아요 | 목록에 이름 더하기 |
| D-2 | 새 경로·옛 경로·쓰기 경로·읽기 경로 함수를 한 모듈에서 주고 읽기는 새 경로 우선·없으면 옛 경로·둘 다 없으면 없음 | 사용자 결정: "옛 경로를 읽고 새 경로에 쓴다" | 함수가 옛 경로만 쓰게 하기 |
| D-3 | 관리 블록의 `tools/`·`evals/` 허용 줄은 `specs/` 뒤에 붙이고 `local/` 줄은 `<state>/*` 줄 바로 뒤에 둔다 | 앞의 아홉 줄 순서가 그대로여야 이미 쓰인 블록과 줄 단위로 비교돼요 | 세 줄 지우기 |
| D-4 | 쓰는 쪽 이동은 SPEC 을 따로 둔다 (개정 4: 안내 문구 중 블록 한 줄과 `agents/impl.md` 한 줄은 이 SPEC 이 맡는다, D-8) | SPEC 하나는 자기 AC 를 모두 `met` 으로 닫아야 하고, 쓰는 곳이 서른 곳 넘고 각각 되돌림·영수증이 걸려 있어요 | SPEC 합치기 |
| D-5 | 이 저장소(moai-adk)의 루트 `tools/` 는 옮기지 않는다 | 사용자 결정(설계 D2): 제품 개발 코드이고 CI·검사 보호 경로가 가리켜요 | 해당 없음 |
| D-6 | 자격 증명·`.env` 줄(`credentials.yaml`, `credentials.json`, `*.credentials`, `.env`, `.env.*`)을 블록에 더하고 허용 줄 뒤에 둔다 | 리더 결정(감사 N2, 개정 5 는 카드 노트 8184 의 N2): 이 저장소의 `.gitignore` 와 Go 세대 템플릿이 이미 갖던 줄이고, `tools/`·`evals/` 허용 줄이 그 안의 `.env` 를 git 에 보이게 하지 않아야 해요. 개정 5 는 이름 하나만 더했고, 다른 이름은 개정 6 이 조사해 더해요(D-10) | 다섯 줄 지우기 |
| D-7 | 새로 보이는 파일은 계획이 `now visible: <경로>` 로 이름을 말한다(스무 개까지) | 리더 결정(감사 N1): 적용 전에 사용자가 어떤 파일이 git 에 나타날지 알아야 해요 | 노트 빼기 |
| D-8 | 안내 문구는 관리 블록 한 줄과 `agents/impl.md` 한 줄로 최소로 하고, 중첩 `.gitignore` 보강(REQ-17)과 같은 SPEC 에서 닫는다 | 리더 결정(카드 노트 8139, 2026-11-24, 사용자 위임 2026-11-23): 카드 기준 2·3 을 이 카드에서 닫아요. 스킬 본문과 홈 규칙 파일은 건드리는 곳이 많아 따로 둬요. 블록 줄은 사실 한 줄이라 15줄 한도 안이에요 | 두 줄 지우기 |
| D-9 | 읽지 않은 중첩 `.gitignore` 의 폴더 아래 파일은 `now visible` 에 이름을 말한다(부정 규칙을 따지지 않는다) | 리더 결정(카드 노트 8184, 2026-11-24, 사용자 위임 2026-11-23): 읽지 못한 규칙은 알 수 없고, 말하는 쪽이 안전한 쪽이에요. 부정 규칙을 흉내 내면 읽지 못한 파일의 내용을 짐작하는 일이라 하지 않아요 | 목록 되돌리기 |
| D-10 | 쓸 수 있는 비밀(평문·되돌릴 수 있는 값·개인 키·토큰)을 담는 것이 용도이고 이름이 좁은 파일만 블록에 더한다. 일반 설정 파일, 암호화·해시된 파일, 여러 용도의 확장자, 정해진 이름이 없는 파일은 더하지 않는다 | 카드 기준 4 와 리더 결정(카드 노트 10187). 넓은 줄은 팀 파일을 조용히 git 밖에 둬요. 조사와 파일별 판단은 `## Research` 예요 | 서른 줄 지우기 |
| D-11 | 중첩 `.gitignore` 는 프로젝트 폴더 핸들에서 경로 조각마다 링크를 따르지 않고 열고(`cap-std`·`cap-fs-ext`), 열린 파일의 metadata 로 보통 파일·크기를 정한다. 앞서 본 경로 정보는 믿지 않는다 | 감사 N5(확인 뒤 여는 틈). 색인 읽기가 이미 쓰는 방식이에요(`docs/index-generation.md`). 마지막 조각만 `O_NOFOLLOW` 로 열면 부모 폴더 교체를 막지 못해요 | 이전 함수로 되돌리기 |

## Research

2026-11-26(UTC)에 두 번 읽었어요. 첫 번째는 조사 에이전트가 가져오기 도구(WebFetch)로 08:05-08:09 에 읽었고, 그 읽기는 원문 바이트와 체크섬을 남기지 않았어요. 두 번째는 레인이 아래 출처 서른일곱 곳을 08:32 에 다시 받아(모두 HTTP 200) 원문을 저장하고 sha256 을 적은 뒤, 근거로 쓴 짧은 문구가 저장한 원문에 있는지 `grep` 으로 확인했어요(빠진 문구 0). 이 절의 근거는 두 번째 읽기예요. 문구 확인은 글자 일치이고, 뜻의 판단은 첫 번째 읽기의 것이에요. `[번호]` 는 출처 목록의 번호예요.

더하는 이름(REQ-19):

| 이름 | 근거 | 출처 |
|---|---|---|
| `.git-credentials` | "stored in plaintext" | [01] |
| `.netrc`, `_netrc` | 암호가 든 파일 | [02] |
| `.pgpass`, `pgpass.conf` | 암호 파일 | [03] |
| `.pypirc` | "password in plain text" | [04] |
| `.credentials.json` | 로그인 자격 증명(`credentials.json` 줄에 안 걸려요) | [05] |
| `credentials.tfrc.json` | API 토큰을 "plain text" 로 저장 | [06] |
| `application_default_credentials.json` | 자격 증명 JSON | [07] |
| `credentials.toml`, `.cargo/credentials` | Cargo 토큰 파일(둘째는 확장자 없는 옛 이름, 폴더를 붙여 좁혀요) | [08] |
| `.vault-token` | 토큰 캐시 | [09] |
| `.mylogin.cnf` | 되돌릴 수 있게 가린 암호 | [10] |
| `master.key` | "Do not commit your master key" | [11] |
| `keystore.properties` | 저장소·키 암호 | [12] |
| `id_rsa`, `id_ecdsa`, `id_ed25519`, `id_ecdsa_sk`, `id_ed25519_sk`, `id_mldsa44_ed25519` | 개인 키 기본 이름(`.pub` 은 안 걸려요) | [13] |
| `id_dsa` | 옛 개인 키 이름 | [14] |
| `*.ppk` | PuTTY 개인 키(암호를 비우면 평문) | [15] |
| `*.tfstate`, `*.tfstate.*` | "plaintext file", 비밀 값 포함 | [16], [17] |
| `.dev.vars`, `.dev.vars.*` | "should not be committed to git" | [18] |
| `.aws/credentials` | 평문 자격 증명(폴더를 붙여 좁혀요) | [19] |
| `.gem/credentials` | API 키 파일 | [20] |
| `.streamlit/secrets.toml` | 비밀 파일, `.gitignore` 에 더하라고 함 | [21] |

더하지 않는 이름:

| 이름 | 까닭 | 출처 |
|---|---|---|
| `*.pem`, `*.key` | 공개 인증서·Keynote 문서도 같은 확장자 | [12], [22] |
| `*.p12`, `*.pfx`, `*.jks`, `*.keystore` | 인증서만 담기도 해요 | [23], [24] |
| `.npmrc`, `.yarnrc.yml`, `.envrc`, `.my.cnf`, `.terraformrc`, `terraform.rc`, `local.settings.json` | 일반 설정 파일이고 비밀은 있을 수도 있는 것 | [25], [26], [27], [28], [29], [30] |
| `.envrc.private` | 문서의 예시 이름일 뿐 | [27] |
| `.htpasswd` | 보통 해시로 저장해요 | [31] |
| `credentials`(이름만), `auth.json`, `config.json`, `kubeconfig`, 서비스 계정 키 JSON | 흔한 이름이거나 정해진 이름이 없어요 | [32], [33], [34], [35] |
| `secrets.yml`, `secrets.yaml`, `credentials.yml`, `*.tfvars` | 암호화·자리표시 파일이나 커밋하는 파일일 수 있고, 출처가 갈려요(`credentials.yml` 은 만드는 도구를 못 찾았어요) | [36], [37] |

확인하지 못한 것: Apple `.p8`, `*.keytab`, `*.publishsettings`, `.dockercfg`, OAuth `client_secret*.json` 은 이름·확장자의 근거를 읽지 못했어요. 대소문자가 다른 이름의 동작은 확인하지 않았어요. 새 줄을 git 이 실제로 무시하는지는 구현 시험(AC-21)이 확인해요.

출처 목록(읽은 시각 UTC 2026-11-26, sha256 은 받은 바이트):

| 번호 | 시각 | sha256 | URL |
|---|---|---|---|
| [01] | 08:32:30 | `da5a50fb3cca1ee64164cc56815c5cdecef409047ce59aa3a5a26275b07c6d39` | https://git-scm.com/docs/git-credential-store |
| [02] | 08:32:31 | `e5ac3761ede396cda32c424ada272a87d27a2a2c9a5fb6e46b5f55f62ed01581` | https://everything.curl.dev/usingcurl/netrc.html |
| [03] | 08:32:32 | `e289dffea8f87aa8b3d55de0a49d6872cd0b3c2e0815388e6f12b6992bd710a1` | https://www.postgresql.org/docs/current/libpq-pgpass.html |
| [04] | 08:32:33 | `dfb77e69745f63e44f78d72431f26a0a17c61db23396d58b9fdad142b682ac5e` | https://raw.githubusercontent.com/pypa/packaging.python.org/main/source/specifications/pypirc.rst |
| [05] | 08:32:33 | `b7f53a69dadcbea9010e802460e34782074fac94a42dbfcd5ce5a70d62be0fc0` | https://code.claude.com/docs/en/authentication.md |
| [06] | 08:32:33 | `13f16f40d549de43911c813108caeec8cd0f8895efd4bf19e9678ec4f8a6c90b` | https://developer.hashicorp.com/terraform/cli/commands/login |
| [07] | 08:32:33 | `f505a234bc2ab2ef3b411cb31309025298404badaff29bc2c9242cdf265ba09d` | https://docs.cloud.google.com/docs/authentication/application-default-credentials |
| [08] | 08:32:34 | `9f97fb6ee7ca9d709bc3e92388b7a1e19f8aab57dc618ee37f978c36cc210c29` | https://doc.rust-lang.org/cargo/reference/config.html |
| [09] | 08:32:34 | `ff2c5ae72a6f3664d4d663cc6156b8848b7cfc6711e7e10aff8959bd2a9696e7` | https://developer.hashicorp.com/vault/docs/commands/token-helper |
| [10] | 08:32:34 | `0e360db9d354203c95d13341293dbd47bf6d9fc036703743d4fc20962773a728` | https://dev.mysql.com/doc/refman/8.4/en/mysql-config-editor.html |
| [11] | 08:32:35 | `51ea8619645c15de9102da45cb54d34f8d4716dd33eab7af6a66ed14c24e6b4e` | https://guides.rubyonrails.org/security.html |
| [12] | 08:32:36 | `d4f27c6e3d26f81a2669d0646169957a061f49013c8da8ce2fbb0696bd41edeb` | https://developer.android.com/studio/publish/app-signing |
| [13] | 08:32:37 | `7381696a690d349ab50e6e86d4760838ed0e123a75d04061ad2a4847c27d7805` | https://man.openbsd.org/ssh-keygen |
| [14] | 08:32:39 | `51595de6f5a3dad74419527a9ee8f65689ec7311024f8b945e72b661765cd7d5` | https://man.openbsd.org/OpenBSD-7.4/ssh-keygen |
| [15] | 08:32:40 | `bfc2e1925729373bb6238065ac4967e8cbf73aea96782e713673a7e330af47af` | https://the.earth.li/~sgtatham/putty/0.83/htmldoc/Chapter8.html |
| [16] | 08:32:41 | `d25f382dc4413c188bd0cb82e7236384a515857c59a159e900e8e4e609df100f` | https://developer.hashicorp.com/terraform/language/state/sensitive-data |
| [17] | 08:32:41 | `4f4ee2ade7ef2653d739316a79c29d0c2a15b8846be364b60d80accf9eaf2763` | https://developer.hashicorp.com/terraform/language/style |
| [18] | 08:32:41 | `b3c7cbc9307a565f91df25a3017b3d369b63978aba5f3b996116c4b5dbe826a5` | https://developers.cloudflare.com/workers/configuration/secrets/ |
| [19] | 08:32:41 | `74ea2736ecd086a32ce2f9b2f6cd1b525bcca456ef8edf4cff00cfd6d5590ea8` | https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html |
| [20] | 08:32:42 | `9bfb188bd8dd37cfc4d3f26c11ee60ceba3386f91f7bb17a2a30d951c75b9c4e` | https://guides.rubygems.org/api-key-scopes/ |
| [21] | 08:32:42 | `c272bce8b83ea68cb60062852494ac9fe07baeda2243d11be9eba52324c1320c` | https://docs.streamlit.io/develop/concepts/connections/secrets-management |
| [22] | 08:32:43 | `e5abd2059a14f3e2e993f11da76933e5aa54ef4e7f6a15a4834a2313a269e779` | https://www.iana.org/assignments/media-types/application/vnd.apple.keynote |
| [23] | 08:32:43 | `21e7a2cdf1f758423db5f11b0d287fccc73a2ef285599de64bae923cdbed8c86` | https://www.rfc-editor.org/rfc/rfc7292 |
| [24] | 08:32:48 | `a31f29f3910b614080f18ed5fb4c21506a347ee0d35085d9b29ae015c1f9a093` | https://docs.oracle.com/en/java/javase/21/docs/specs/man/keytool.html |
| [25] | 08:32:48 | `81e06b9efe302a9c029150726fdb31a8119df11f981b43c8ae07096f1ada5a55` | https://docs.npmjs.com/cli/v11/configuring-npm/npmrc |
| [26] | 08:32:49 | `f323fb905bf175ab6a6f11e339e093bdc143a7948c383ed77eb940e0784dca08` | https://yarnpkg.com/configuration/yarnrc |
| [27] | 08:32:49 | `c43ac6b7d8ff4da490923094d6447cdaf3827f65d53bf901ae0082a0003b2d55` | https://direnv.net/man/direnv-stdlib.1.html |
| [28] | 08:32:49 | `565a30ea74a07a961d2dbd3ad44621bcfb75b59ec11584b7fa1a2cd49f48e53e` | https://dev.mysql.com/doc/refman/8.4/en/password-security-user.html |
| [29] | 08:32:50 | `06ba493b04e74e136d59e968214587bc017f45891f3d9a82c5e3072ce992f5c8` | https://developer.hashicorp.com/terraform/cli/config/config-file |
| [30] | 08:32:50 | `579576ebde5f232f63e4c3d85cd84852781d2eb3bc82f89afc6cf06f12f378c6` | https://learn.microsoft.com/en-us/azure/azure-functions/functions-develop-local |
| [31] | 08:32:50 | `c8c5385b4620da004a8e2ff91a14d7389a55897221c888b3f70d0c0a4f7721b3` | https://httpd.apache.org/docs/current/programs/htpasswd.html |
| [32] | 08:32:51 | `602b4d1011d89fcc3fc3424aff646f8e6eb1e7a0615dbf97437f9186e45593d5` | https://getcomposer.org/doc/articles/authentication-for-private-packages.md |
| [33] | 08:32:51 | `7ed98a37339f28ede26fdb5a98fac7216fd907d720a53a47786b3c2477082bde` | https://docs.docker.com/reference/cli/docker/login/ |
| [34] | 08:32:53 | `982236c48e2ab24ad2f7f7e5d8780403d3056fb7e3494173ed1da65b0d550954` | https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/ |
| [35] | 08:32:53 | `320449fa3b87ba1e2b3a78abb7b37c8c100384bfcdd8b571d95911a63b9ab2bf` | https://docs.cloud.google.com/iam/docs/keys-create-delete |
| [36] | 08:32:55 | `e37a3fcdad851726bfecb357669293fc4a40b8d91387d9d3923482c405fd961a` | https://docs.ansible.com/ansible/latest/vault_guide/vault.html |
| [37] | 08:32:55 | `ae754e7ea190151f1d6eb579484e08c24acabb1ca42f2517f44620757abc2cdd` | https://developer.hashicorp.com/terraform/language/values/variables |

## Result

구현 끝(2026-11-20, 가지 `spec/20261120-moai-layout`, 조각 1). 쓰는 쪽 이동과 안내 문구는 뒤 SPEC 들이 맡아요(D-4). 변이 다섯(`target/c/mut174.py`, 결과 `target/c/layout2.sum`)은 모두 시험에 잡혔고 각 변이를 고친 뒤 파일은 그대로 되돌렸어요.

감사 수리(개정 3, 2026-11-21): 감사가 막은 `read_path` 의 끊어진 링크 시험(B1)과 리더 결정 둘(N1 계획의 새로 보이는 파일 이름, N2 자격 증명·`.env` 무시 줄)을 같은 가지에서 고쳤어요. 변이 여덟(`target/c/mutfix.py`, 결과 `target/c/mutfix.sum`)을 모두 시험이 잡았고 각 변이 뒤 `git checkout -q -- src` 로 되돌렸어요.

- AC-1: met — `src/state_root.rs` 의 `KEPT`(열한 종류)·`LOCAL_KINDS`·`local_path`·`earlier_path` 를 `cargo test --locked --offline --lib state_root::tests::the_layout_names_kept_kinds_and_local_paths` 가 못 박아요(통과). 변이 `evals/` → `evals-x/`(추적 목록 이름 바꾸기)와 `is_kept` 의 `==` → `!=` 에 이 시험이 실패했어요.
- AC-2: met — `cargo test --locked --offline --test gitignore_block_cli the_block_ignores_local_and_keeps_tools_and_evals` 통과(`git check-ignore` 가 `tools/`·`evals/` 파일은 무시하지 않고 `local/` 안은 무시한다고 답해요). 변이 "블록의 `local/` 줄 빼기"에 이 시험을 포함한 `gitignore_block_cli` 시험 넷이 실패했어요.
- AC-3: met — `cargo test --locked --offline --test gitignore_block_cli an_earlier_block_is_shown_as_changed_and_left_until_applied` 통과(옛 블록은 새 줄 여덟이 더해지는 `merge`: `local/`, `tools/`·`evals/` 허용, 자격 증명·`.env` 다섯. 적용 전 파일 그대로). 두 골든(`16-agents-plan-create`, `17-agents-apply-ok`)은 `MOAI_UPDATE_FIXTURES=1` 로 다시 만들었고 `git diff` 가 블록 줄과 해시뿐임을 확인했어요.
- AC-4: met — `cargo test --locked --offline --lib state_root::tests::a_reader_falls_back_to_the_earlier_path_and_a_writer_uses_the_new_one` 통과(새 경로 우선·옛 경로·둘 다 없음=없음·쓰기는 새 경로로 옛 폴더 그대로). 변이 "`read_path` 의 순서 바꾸기"와 "`write_path` 가 옛 경로를 돌려주기"에 이 시험이 실패했어요.
- AC-5: met — `cargo test --locked --offline --lib state_root::tests::host_owned_files_are_not_in_the_layout` 통과(호스트 소유 네 곳이 추적·기계 출력 어느 목록에도 없음). 변이 `is_kept` 의 `==` → `!=` 에 이 시험이 실패했어요. `HOST_OWNED` 항목을 목록에 넣는 변이는 따로 돌리지 않았어요.
- AC-6: met — `cargo test --locked --offline --lib state_root::tests::a_broken_link_counts_as_the_path_being_there` 통과. 변이 "`symlink_metadata` → `metadata`"(감사의 M3)에 이 시험이 실패했어요.
- AC-7: met — `cargo test --locked --offline --test gitignore_block_cli the_block_ignores_credential_and_env_files_even_in_tools_and_evals` 통과(git `check-ignore` 가 `tools/.env`·`tools/credentials.yaml`·`tools/credentials.json`·`evals/api.credentials` 등은 무시하고 `check.sh`·`env.sh`·`set.json` 은 무시하지 않으며 줄이 마지막 허용 줄 뒤에 있음). 변이 "자격 증명 줄 렌더 빼기"와 "그 줄을 허용 줄 앞으로 옮기기"에 실패했어요.
- AC-8: met — `cargo test --locked --offline --test gitignore_block_cli the_plan_names_the_files_the_update_makes_visible` 통과(`evals/set.json`·`tools/check.sh` 만 `now visible`, `.env`·`credentials.yaml`·`token.credentials` 는 말하지 않음, 미리보기는 쓰지 않음, 적용 뒤 계획은 말하지 않음). 변이 "자격 증명 줄 렌더 빼기", "새 블록을 보지 않고 묻기", "규칙에 걸린 파일도 말하기"에 실패했어요.
- AC-9: met — `cargo test --locked --offline --test gitignore_block_cli a_file_the_users_own_rule_ignores_is_not_named_as_visible` 통과(`*.log` 규칙이 있는 `run.log` 는 말하지 않고 `run.sh` 만 말해요). 변이 "규칙에 걸린 파일도 말하기"에 실패했어요.
- AC-10: met — `cargo test --locked --offline --test gitignore_block_cli the_plan_counts_the_visible_files_past_twenty` 통과(스무 개는 스무 줄·`more files` 없음, 스물한 개는 스물한 줄에 마지막이 `and 1 more files under …`). 변이 "`>` → `>=`"와 "상한 20 → 21"에 실패했어요.
- AC-11: met — `cargo test --locked --offline --test gitignore_block_cli no_repository_and_no_git_are_not_failures` 통과(git 없음 줄 한 번). 변이 "이미 말한 노트를 또 더하기"에 실패했어요.
- AC-12: met — `cargo test --locked --offline --lib agents::tests::the_block_names_the_tools_folder_for_project_scripts` 통과(`en`·`ko` 블록이 `.moai/tools/` 줄을 한 줄 갖고 `scripts/`·`project root`·`프로젝트 루트` 가 없고 줄 수가 15 이내). 구현 전에 이 시험은 "도구 줄 0개"로 실패했어요. 변이 "영어 줄 빼기", "한국어 줄 빼기", "줄이 `scripts/` 를 말하기" 셋에 실패했어요.
- AC-13: met — `cargo test --locked --offline --lib agents::tests::the_impl_agent_names_the_tools_folder_for_project_scripts` 통과(`agents/impl.md` 에 `.moai/tools/` 가 한 번, `scripts/` 와 프로젝트 루트 문구는 없음). 구현 전에 실패(0 ≠ 1)했고 변이 "한 줄 빼기"에 실패했어요.
- AC-14: met — `cargo test --locked --offline --test gitignore_block_cli a_nested_gitignore_under_the_state_folder_keeps_a_file_out_of_the_visible_list` 통과(`tools/.gitignore` 의 `private/` 와 `tools/sub/.gitignore` 의 `*.tmp` 가 막는 파일은 말하지 않고 `check.sh`·`sub/y.sh` 는 말하며, 적용 뒤 git 의 답과 같아요). 구현 전에 실패(여섯 줄 대 네 줄)했고 변이 "중첩 파일을 쓰지 않기", "중첩 파일을 폴더 루트에 쓰기"에 실패했어요.
- AC-15: met — `cargo nextest run --test gitignore_block_cli a_nested_gitignore_of_64_kib_is_read_and_one_byte_more_is_not` 통과(65536 바이트는 읽혀 `x.tmp` 를 말하지 않고 안내 줄도 없으며, 65537 바이트는 읽히지 않아 `x.tmp` 를 말하고 안내 줄이 한 번). 감사가 살아남았다고 지적한 상한 제거(읽기 전 `len >` 검사와 읽은 뒤 길이 검사를 둘 다 지움)와 상한 값 올리기(64 KiB → 128 KiB)에 이 시험이 실패했어요.
- AC-16: met — `cargo nextest run --test gitignore_block_cli the_hundredth_nested_gitignore_is_read_and_the_hundred_first_is_not` 통과(백 개째 폴더는 규칙을 들어 아무것도 말하지 않고, 백한 개째 폴더는 그 안의 `x.tmp` 와 그 폴더의 `.gitignore` 를 말하며 안내 줄이 한 번). 변이 "상한 100 → 2"와 "읽지 않은 폴더 아래를 읽은 규칙에 맡기기"에 실패했어요.
- AC-17: met — `cargo nextest run --test gitignore_block_cli a_nested_gitignore_that_is_a_link_is_not_read` 통과(프로젝트 밖의 `*.sh` 규칙을 가리키는 링크를 읽지 않고 `x.sh` 를 말하며 안내 줄이 한 번). 변이 "링크 검사의 `symlink_metadata` → `metadata`"에 실패했어요.
- AC-18: met — `cargo nextest run --test gitignore_block_cli a_nested_gitignore_behind_a_linked_folder_is_not_read` 통과(`tools/lp` 가 프로젝트 밖을 가리키는 링크여도 그 안의 `.gitignore` 를 읽지 않고 안내 줄이 한 번). 변이 "마지막 경로 조각만 링크 검사"와 "`symlink_metadata` → `metadata`"에 실패했어요.
- AC-19: met — `cargo nextest run --test gitignore_block_cli an_unreadable_nested_gitignore_is_not_read` 통과(mode 000 파일은 읽지 않고 그 폴더 아래 파일을 말하며 안내 줄이 한 번, 유닉스). 변이 "읽을 수 없는 파일을 빈 규칙으로 읽기"에 실패했어요.
- AC-20: met — `cargo nextest run --test gitignore_block_cli a_file_below_an_unread_gitignore_is_named_even_when_a_parent_rule_hides_it` 통과(65537 바이트라 읽지 않은 자식의 `!keep.tmp` 가 있고 읽은 부모 `*.tmp` 가 숨기는 `keep.tmp` 를 말하며, 읽지 않은 폴더 밖의 `top.tmp` 는 읽은 규칙대로 말하지 않아요. 적용 뒤 git 은 `keep.tmp` 를 무시하지 않고 `top.tmp` 는 무시한다고 답해 계획과 같아요). 변이 "읽지 않은 폴더 아래를 읽은 규칙에 맡기기"에 실패했어요.
- AC-22: met — 커밋 `2fcd0acc2c` 의 `src/agents.rs` 와 같은 바이트(sha256 `bb3e024255f5bd5b…`)에서 `cargo nextest run --locked --offline --lib agents::` 18개 통과(2026-11-26, macOS). 고치기 전 여는 방식(경로 문자열로 열기)에서는 링크 시험 넷과 파이프 시험이 실패했어요(exit 100, `target/dana/red2.log`). 변이 "마지막 조각이 링크를 따르기", "부모 폴더가 링크를 따르기"(`open_dir_nofollow` → `open_dir`), "`O_NONBLOCK` 빼기", "보통 파일 확인 빼기"가 각각 시험 하나를 실패시켰어요(exit 100). Windows 와 Linux 는 돌리지 않았어요.
- 확인(개정 5 수리, 2026-11-24): 변이 여덟(`target/c174-repair/mutants/`, 요약 `mutants-summary.txt`)은 모두 시험이 잡았고(상한 제거, 상한 값 올리기, 상한 100 → 2, `symlink_metadata` → `metadata`, 마지막 조각만 보기, 읽지 않은 폴더 아래를 규칙에 맡기기, `credentials.json` 줄 빼기, 읽을 수 없는 파일을 빈 규칙으로 읽기) 각 변이 뒤 원본을 복사본과 `cmp` 로 되돌려 `git diff` 가 깨끗했어요. `gitignore_block_cli` 46개, `agents::` 단위 12개, `state_folder_literal_scan` 10개 통과. 골든 16·17 은 블록 줄(열여덟 → 열아홉)과 해시 셋만 바뀌었어요. 감사 N4(`newly_visible` 의 "안전한 쪽" 주석)는 같은 변경에서 REQ-18 에 맞게 고쳤고, N2(`credentials.json`)는 `STATE_SECRETS` 에 더했어요. 다른 키 파일 이름 조사는 카드 기준으로 남아 있어요. 전체 `moai check affected` 는 병합 준비 노트에 줄째로 적어요.
- 확인(개정 4, 2026-11-24): 변이 여섯은 모두 시험이 잡았고(`mut174.sum` 의 exit 101 여섯) 원본은 복사본과 `cmp` 로 같게 되돌렸어요. 골든 16·17 은 해시 셋만 바뀌었어요. `gitignore_block_cli` 40개, `product_cli` 74개, `agents::` 단위 12개 통과.
- 확인(2026-11-20): `moai check affected`(대상 179개) 는 fmt·clippy·ops rustfmt·`git diff --check`·파이썬·build_cli 가 0 이었고, `card::input::tests::secret_screen_time_on_one_long_word_has_a_bound` 와 `tests/home_retire_cli.rs` 의 `retire_keeps_the_go_markers_until_the_entries_in_use_are_free` 두 시험만 부하 때문에 시간 기준으로 실패했어요(단독 실행은 통과: 2.14 s, 61.5 s). 둘 다 시간 기준 시험을 작업량 기준으로 바꾸는 별도 변경의 범위예요.
- 확인(2026-11-21, 수리): `gitignore_block_cli` 38개 통과, `--lib` 1642개 통과, `product_cli` 79개 통과(골든 둘은 블록 줄 넷과 해시만 바뀜). 전체 `moai check affected` 는 병합 준비 노트에 줄째로 적어요.
