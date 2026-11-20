---
id: SPEC-20261120-moai-layout
title: "One .moai/ layout rule for every project: tracked team files, ignored .moai/local/, project tools in .moai/tools/"
status: draft
revision: 3
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

이 SPEC 은 규칙의 **조각 1** 이에요: 규칙의 정의(코드 한 곳)와 관리되는 `.gitignore` 블록. 규칙의 나머지는 다음 조각의 SPEC 이 맡아요(SPEC 하나는 자기 AC 를 모두 `met` 으로 닫아야 해서 조각마다 SPEC 을 따로 둬요): 쓰는 쪽을 종류별로 `local/` 로 옮기는 SPEC 들, 만들어지는 `AGENTS.md` 와 스킬 지침이 `.moai/tools/` 를 말하게 하는 SPEC.

## Non-goals

- 쓰는 쪽을 `local/` 로 옮기지 않아요: 종류마다 다음 SPEC 이 해요(`receipts` → `index` → `state`·`forge/outbox` → `backups` 등). 이 SPEC 은 그 길(함수)만 줘요.
- 만들어지는 `AGENTS.md` 와 스킬 지침의 `.moai/tools/` 문구는 다음 SPEC 이 해요.
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
- REQ-11: The managed `.gitignore` block shall ignore the files `credentials.yaml`, `*.credentials`, `.env` and `.env.*` anywhere under the state folder, `tools/` and `evals/` included, with its lines after the allow-list.
- REQ-12: When a project update would lift the ignore from untracked files under the state folder's `tools/` or `evals/`, the project update plan shall name each such file as `now visible: <path>`, up to twenty, and count the rest in one line.
- REQ-13 (REQ-12): IF git is not available, THEN the project update plan shall name no file and say once that git was not consulted.
- REQ-14 (REQ-5): IF the new path or the earlier path is a link whose target is missing, THEN the state layout shall still return that path as the one that is there.

## Acceptance

- AC-1 (REQ-1, REQ-2): 레이아웃 정의가 추적 종류 열한 개를 이름으로 대고 나머지는 `local/` 아래라고 하며, 지금 있는 기계 출력 종류 목록을 갖고, 새 경로와 옛 경로를 한 함수씩 돌려줘요.
  verify: `cargo test --locked --offline --lib state_root::tests::the_layout_names_kept_kinds_and_local_paths`
- AC-2 (REQ-3): 새 프로젝트의 관리 블록은 `<state>/local/` 줄을 갖고 `tools/`·`evals/` 허용 줄을 `specs/` 뒤에 갖고, git 이 `tools/`·`evals/` 의 파일은 무시하지 않고 `local/` 안은 무시한다고 답해요.
  verify: `cargo test --locked --offline --test gitignore_block_cli the_block_ignores_local_and_keeps_tools_and_evals`
- AC-3 (REQ-4, REQ-8): 옛 블록이 있는 프로젝트의 업데이트 계획은 새 줄 일곱(`local/`, `tools/`·`evals/` 허용, 자격 증명·`.env` 넷)이 더해지는 `merge` 로 보여 주고, 적용 전에는 파일이 그대로예요.
  verify: `cargo test --locked --offline --test gitignore_block_cli an_earlier_block_is_shown_as_changed_and_left_until_applied`
- AC-4 (REQ-5, REQ-6, REQ-9, REQ-10): 읽는 길은 `local/<kind>` 를 먼저, 없으면 옛 경로를, 둘 다 없으면 없음을 돌려주고, 쓰는 길은 새 경로를 주어 옛 폴더를 건드리지 않아요.
  verify: `cargo test --locked --offline --lib state_root::tests::a_reader_falls_back_to_the_earlier_path_and_a_writer_uses_the_new_one`
- AC-5 (REQ-7): 호스트 소유 파일 네 곳은 추적 종류에도 기계 출력 종류에도 없어요.
  verify: `cargo test --locked --offline --lib state_root::tests::host_owned_files_are_not_in_the_layout`
- AC-6 (REQ-5, REQ-14): 끊어진 링크인 `local/<kind>` 는 있는 것으로 세어 옛 폴더로 건너뛰지 않고, 끊어진 옛 링크도 없음이 아니라 그 경로로 돌려줘요.
  verify: `cargo test --locked --offline --lib state_root::tests::a_broken_link_counts_as_the_path_being_there`
- AC-7 (REQ-11): git 이 `tools/`·`evals/`·그 밖의 상태 폴더 아래 어디서나 `credentials.yaml`·`*.credentials`·`.env`·`.env.*` 는 무시하고 `check.sh`·`env.sh`·`set.json` 은 무시하지 않으며, 그 줄들이 허용 줄 뒤에 와요.
  verify: `cargo test --locked --offline --test gitignore_block_cli the_block_ignores_credential_and_env_files_even_in_tools_and_evals`
- AC-8 (REQ-12): 업데이트가 `tools/`·`evals/` 의 무시를 푸는 프로젝트의 계획이 새로 보이게 될 파일마다 `now visible: <경로>` 를 말하고, 자격 증명·`.env` 파일은 말하지 않으며, 적용 뒤의 계획은 말하지 않아요.
  verify: `cargo test --locked --offline --test gitignore_block_cli the_plan_names_the_files_the_update_makes_visible`
- AC-9 (REQ-12): 사용자 자신의 규칙이 계속 무시하는 파일은 `now visible` 로 말하지 않아요.
  verify: `cargo test --locked --offline --test gitignore_block_cli a_file_the_users_own_rule_ignores_is_not_named_as_visible`
- AC-10 (REQ-12): 계획은 파일 스무 개까지 이름을 말하고 나머지는 한 줄로 세어요: 스무 개는 모두 말하고 스물한 개는 스무 개와 `and 1 more files` 줄이에요.
  verify: `cargo test --locked --offline --test gitignore_block_cli the_plan_counts_the_visible_files_past_twenty`
- AC-11 (REQ-13): git 이 없으면 계획은 파일 이름을 말하지 않고 git 을 보지 않았다는 한 줄을 한 번만 말해요.
  verify: `cargo test --locked --offline --test gitignore_block_cli no_repository_and_no_git_are_not_failures`

## Scope

- write: `.moai/specs/SPEC-20261120-moai-layout/spec.md`, `src/state_root.rs`, `src/agents.rs`(관리 블록: `STATE_KEPT` 와 `render_ignore`), `tests/gitignore_block_cli.rs`, 골든 `tests/fixtures/protocol/v1/current/16-agents-plan-create.json`·`17-agents-apply-ok.json`(블록 줄이 바뀌어서), `docs/design-review/appendix/state-layout.md`
- read-only: `src/layout.rs`(P09 홈 배치와는 별개예요), `src/go_state.rs`(옛 이름 목록)
- forbidden: 쓰는 쪽 모듈 바꾸기, 옛 경로 파일 이동·삭제, `.moai-cli` 이전 변경, 보고서 위치 변경, 안내 문구 바꾸기

## Plan

1. SPEC 을 커밋해요(이 문서). `spec check` 를 돌려요.
2. `src/state_root.rs` 에 `KEPT`(열한 종류)·`LOCAL_KINDS`·`HOST_OWNED`·`local_path`·`earlier_path`·`write_path`·`read_path`·`is_kept` 를 두고 단위 시험 셋(`the_layout_names_kept_kinds_and_local_paths`, `a_reader_falls_back_to_the_earlier_path_and_a_writer_uses_the_new_one`, `host_owned_files_are_not_in_the_layout`)으로 못 박아요.
3. `src/agents.rs` 의 `STATE_KEPT` 를 `state_root::KEPT` 로 바꾸고 블록에 `<state>/local/` 줄을 더해요(허용 줄 순서는 앞의 아홉 줄 그대로, `tools/`·`evals/` 가 `specs/` 뒤). `gitignore_block_cli` 시험 둘을 더하고, 줄 수를 못 박은 기존 시험 둘(12줄 → 15줄)을 고쳐요.
4. 골든 두 개를 `MOAI_UPDATE_FIXTURES=1` 로 다시 만들고 `git diff` 가 블록 줄과 해시뿐인지 봐요.
5. 확인: `moai check affected origin/v2026.11.0..HEAD`(내가 빌드한 moai), 변이(블록 줄 빼기·허용 줄 빼기·`read_path` 순서 바꾸기·`is_kept` 깨기)가 시험에 잡히는지, SPEC `spec check` 와 `--closure`.

## Risks

- 업데이트로 `tools/`·`evals/` 의 무시가 풀리면 그 안에 이미 있던 사용자 파일(예: `.moai/tools/.env`)이 git 에 새로 보여요(감사 N1). 계획이 `now visible: <경로>` 로 이름을 말하고, 자격 증명·`.env` 류는 블록의 새 줄이 계속 무시해요. 이름이 스무 개를 넘으면 앞 스무 개와 나머지 수예요. 키 파일(`*.key`, `id_rsa` 등)과 `secrets.*` 는 이 줄들이 막지 않아요: 목록에 더하는 것은 한 줄이에요.
- `.env.*` 줄은 `.env.example` 도 무시해요: 예제 파일을 `tools/` 에 두려는 사용자는 자기 `.gitignore` 에 `!.moai/tools/.env.example` 을 더해야 해요(안전한 쪽으로 틀려요).

- 관리 블록의 줄이 바뀌어서 이미 블록이 있는 프로젝트는 다음 업데이트 계획에서 "블록이 바뀐다"를 보게 돼요: 적용은 사용자가 계획을 확인한 뒤에만이에요. 골든 두 개(`16-agents-plan-create`, `17-agents-apply-ok`)와 `gitignore_block_cli` 의 줄 수 기대 둘이 함께 바뀌어요.
- 모든 상태가 이미 `<state>/*` 규칙으로 무시되고 있어서(허용 목록 밖은 무시) `local/` 줄은 읽는 사람에게 규칙을 말해 주는 줄이에요. 무시 결과는 달라지지 않아요.
- 함수만 있고 쓰는 쪽은 아직 옛 경로에 써요: 그 사이에 `local/` 은 비어 있고 `read_path` 는 옛 경로를 돌려줘요. 쓰는 쪽을 옮기는 SPEC 이 종류마다 되돌림·영수증 시험을 둬요.
- `LOCAL_KINDS` 는 지금 있는 종류의 목록이고 `go_state` 의 옛 이름 목록(`STATE_ENTRIES`)과 따로 있어요: 둘의 관계를 못 박는 시험은 `go_state` 를 쓰는 다음 SPEC 이 더해요(이 SPEC 은 그 파일을 쓰지 않아요).
- `.moai-cli` 와 `.moai` 가 함께 있는 프로젝트(이 저장소 포함)에서 `local/` 은 `state_root::dir` 가 정하는 쪽 아래예요. 이전 마무리는 이 SPEC 밖이에요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 추적 종류는 열한 개 이름을 코드 한 곳에 두고, 그 밖의 모든 종류는 `local/` 아래로 본다 | 사용자 결정: "추적은 팀 파일, 나머지는 이 기계의 것" — 목록이 닫혀 있어야 새 종류가 실수로 추적되지 않아요 | 목록에 이름 더하기 |
| D-2 | 새 경로·옛 경로·쓰기 경로·읽기 경로 함수를 한 모듈에서 주고 읽기는 새 경로 우선·없으면 옛 경로·둘 다 없으면 없음 | 사용자 결정: "옛 경로를 읽고 새 경로에 쓴다" | 함수가 옛 경로만 쓰게 하기 |
| D-3 | 관리 블록의 `tools/`·`evals/` 허용 줄은 `specs/` 뒤에 붙이고 `local/` 줄은 `<state>/*` 줄 바로 뒤에 둔다 | 앞의 아홉 줄 순서가 그대로여야 이미 쓰인 블록과 줄 단위로 비교돼요 | 세 줄 지우기 |
| D-4 | 쓰는 쪽 이동과 안내 문구는 SPEC 을 따로 둔다 | SPEC 하나는 자기 AC 를 모두 `met` 으로 닫아야 하고, 쓰는 곳이 서른 곳 넘고 각각 되돌림·영수증이 걸려 있어요 | SPEC 합치기 |
| D-5 | 이 저장소(moai-adk)의 루트 `tools/` 는 옮기지 않는다 | 사용자 결정(설계 D2): 제품 개발 코드이고 CI·검사 보호 경로가 가리켜요 | 해당 없음 |
| D-6 | 자격 증명·`.env` 줄(`credentials.yaml`, `*.credentials`, `.env`, `.env.*`)을 블록에 더하고 허용 줄 뒤에 둔다 | 리더 결정(감사 N2): 이 저장소의 `.gitignore` 와 Go 세대 템플릿이 이미 갖던 줄이고, `tools/`·`evals/` 허용 줄이 그 안의 `.env` 를 git 에 보이게 하지 않아야 해요 | 네 줄 지우기 |
| D-7 | 새로 보이는 파일은 계획이 `now visible: <경로>` 로 이름을 말한다(스무 개까지) | 리더 결정(감사 N1): 적용 전에 사용자가 어떤 파일이 git 에 나타날지 알아야 해요 | 노트 빼기 |

## Result

구현 끝(2026-11-20, 가지 `spec/20261120-moai-layout`, 조각 1). 쓰는 쪽 이동과 안내 문구는 뒤 SPEC 들이 맡아요(D-4). 변이 다섯(`target/c/mut174.py`, 결과 `target/c/layout2.sum`)은 모두 시험에 잡혔고 각 변이를 고친 뒤 파일은 그대로 되돌렸어요.

- AC-1: met — `src/state_root.rs` 의 `KEPT`(열한 종류)·`LOCAL_KINDS`·`local_path`·`earlier_path` 를 `cargo test --locked --offline --lib state_root::tests::the_layout_names_kept_kinds_and_local_paths` 가 못 박아요(통과). 변이 `evals/` → `evals-x/`(추적 목록 이름 바꾸기)와 `is_kept` 의 `==` → `!=` 에 이 시험이 실패했어요.
- AC-2: met — `cargo test --locked --offline --test gitignore_block_cli the_block_ignores_local_and_keeps_tools_and_evals` 통과(`git check-ignore` 가 `tools/`·`evals/` 파일은 무시하지 않고 `local/` 안은 무시한다고 답해요). 변이 "블록의 `local/` 줄 빼기"에 이 시험을 포함한 `gitignore_block_cli` 시험 넷이 실패했어요.
- AC-3: met — `cargo test --locked --offline --test gitignore_block_cli an_earlier_block_is_shown_as_changed_and_left_until_applied` 통과(옛 블록은 새 줄 셋이 더해지는 `merge`, 적용 전 파일 그대로). 두 골든(`16-agents-plan-create`, `17-agents-apply-ok`)은 `MOAI_UPDATE_FIXTURES=1` 로 다시 만들었고 `git diff` 가 블록 줄과 해시뿐임을 확인했어요.
- AC-4: met — `cargo test --locked --offline --lib state_root::tests::a_reader_falls_back_to_the_earlier_path_and_a_writer_uses_the_new_one` 통과(새 경로 우선·옛 경로·둘 다 없음=없음·쓰기는 새 경로로 옛 폴더 그대로). 변이 "`read_path` 의 순서 바꾸기"와 "`write_path` 가 옛 경로를 돌려주기"에 이 시험이 실패했어요.
- AC-5: met — `cargo test --locked --offline --lib state_root::tests::host_owned_files_are_not_in_the_layout` 통과(호스트 소유 네 곳이 추적·기계 출력 어느 목록에도 없음). 변이 `is_kept` 의 `==` → `!=` 에 이 시험이 실패했어요. `HOST_OWNED` 항목을 목록에 넣는 변이는 따로 돌리지 않았어요.
- 확인(2026-11-20): `moai check affected`(대상 179개) 는 fmt·clippy·ops rustfmt·`git diff --check`·파이썬·build_cli 가 0 이었고, `card::input::tests::secret_screen_time_on_one_long_word_has_a_bound` 와 `tests/home_retire_cli.rs` 의 `retire_keeps_the_go_markers_until_the_entries_in_use_are_free` 두 시험만 부하 때문에 시간 기준으로 실패했어요(단독 실행은 통과: 2.14 s, 61.5 s). 둘 다 시간 기준 시험을 작업량 기준으로 바꾸는 별도 변경의 범위예요.
