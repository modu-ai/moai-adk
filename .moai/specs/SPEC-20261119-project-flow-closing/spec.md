---
id: SPEC-20261119-project-flow-closing
title: "The PROJECT route reports an ignored profile as a selection and ends with a closing commit selection"
status: draft
revision: 1
created: 2026-11-19
---

# The PROJECT route reports an ignored profile as a selection and ends with a closing commit selection

## Intent

카드 C-59(사용자가 mo.ai.kr 에서 `/moai project` 를 돌린 기록, 2026-11-16)의 남은 두 기준이에요. 리더 결정(note 2528, 2026-11-19): 기준 2(MCP 편집)는 하지 않고, 기준 1·3 은 tip 근거로 닫고, 기준 4·5 만 해요.
(4) profile 이 `.gitignore` 에 막혀 있어도 경로만 알려 주고 끝났어요. 제품 코드는 이미 `.moai/project.yaml` 을 git 에 남기는 관리 블록을 쓰고(#3009 S3b), 사용자 규칙이 그걸 막으면 업데이트 계획의 `.gitignore` 행 `detail` 에 `left ignored by your rule …: <경로>` 줄을 내보내요. 빠진 것은 그 사실을 "git 에 남길까/내 컴퓨터에만 둘까" 선택으로 바꾸는 워크플로 글이에요.
(5) 사용자 기록에서 라우트(route)가 커밋 안 된 삭제 190 개를 남긴 채 끝났고 닫는 선택이 없었어요(카드 원문 관찰 6). 라우트 끝에서 바뀐 경로를 단계별로 보여 주고 "커밋 / 그대로 두기" 를 고르게 해야 해요.
이 SPEC 은 `skills/moai/workflows/project.md` 의 글과 그 글이 말하는 git 동작을 고정하는 시험만 다뤄요. 새 Rust 명령은 없어요.

## Non-goals

- 기준 2(MCP plan/apply): 리더 결정 note 2528 로 이 카드에서 하지 않아요. `project.md` 의 "MoAI never creates or edits an MCP configuration" 는 그대로예요.
- 기준 1·3: tip 에 근거가 있어서 리더가 met 로 적어요(이 SPEC 의 쓰기가 아니에요).
- `.gitignore` 관리 블록(`src/agents.rs`)의 규칙 바꾸기: 사용자 규칙이 이기는 현재 동작을 그대로 둬요.
- 커밋을 push 하거나 PR 로 올리는 일: 라우트의 끝은 로컬 커밋까지예요.

## Requirements

- REQ-1: When the profile has been written by `project-apply`, the PROJECT workflow shall ask git whether the profile is tracked (`git ls-files --error-unmatch`) and, if it is not, whether git ignores it (`git check-ignore -q`).
- REQ-2: When git ignores the profile, the PROJECT workflow shall report the rule that does it (`git check-ignore -v`) and offer one selection, "Track it" (recommended) or "Keep it local".
- REQ-3: When the user picks "Track it", the PROJECT workflow shall add the one line `!<profile path>` at the end of the root `.gitignore`, after the named rule and outside any managed block, and ask `git check-ignore -q` again.
- REQ-4: When the user picks "Keep it local", the PROJECT workflow shall change nothing and say that a fresh clone will not have the profile. · failure: none — nothing is written, so there is no way for it to fail halfway.
- REQ-5: The PROJECT workflow shall end, after the Completion Check's offer of left-out steps has been answered and its steps have run, by listing the paths that `git status --porcelain=v1 -z --untracked-files=all` shows as changed, each under the apply or edit whose snapshot first showed it, and a path no snapshot gave to a step under "not from this route".
- REQ-6: When the list holds at least one path of this route, the PROJECT workflow shall offer one selection, "Commit these changes" (recommended) or "Leave them uncommitted".
- REQ-7: When the user picks the commit, the PROJECT workflow shall stage exactly the listed paths of this route with `git add -- <path>` (never `-A`, `.` or a path under "not from this route"), commit without pushing and with a message in the project's language setting, and report the commit id with `git reset --soft HEAD~1` as the way back.
- REQ-8: The PROJECT workflow shall list a path that has a secret-looking name anywhere in it (`.env*`, `*.key`, `*.pem`, `*credentials*`, `*secret*`) or that sits under `.moai-backup/`, `.moai/backups/`, `.moai/receipts/` or a `.before/` folder with its reason ("secret name" or "kept local") and never stage it.
- REQ-9 (REQ-2, REQ-6): IF a selection is cancelled, empty or unanswered, THEN the PROJECT workflow shall treat it as "Keep it local" or "Leave them uncommitted" and write nothing.
- REQ-10 (REQ-1, REQ-5, REQ-7): IF `git ls-files`, `git check-ignore`, `git status`, a snapshot, `git add` or `git commit` fails (for the first two, an exit other than 0 or 1), THEN the PROJECT workflow shall report git's message, ask nothing more about tracking or the closing commit, and name every path already staged without retrying in a loop.
- REQ-11 (REQ-3): IF git still ignores the profile after the `!` line is added, THEN the PROJECT workflow shall remove exactly that line and report the rule that still ignores it.
- REQ-12: The PROJECT workflow shall read the `git status --porcelain=v1 -z --untracked-files=all` list before the route's first apply and again after each apply and after the "Track it" edit, and give every new or changed path to the step that preceded that read, never taking paths from an apply's result.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-9, REQ-10, REQ-11): `project.md` 에 "Profile tracking" 항목이 있고 위 순서(tracked 확인 → ignored 확인 → 규칙 이름 → 선택 → 되돌리기 가능한 한 줄 → 다시 확인)와 git 오류·무응답·여전히 막힘의 처리를 말해요.
  verify: `cargo test --locked --offline --test project_flow_closing profile_tracking_step_states_the_checks_the_selection_and_the_undo`
- AC-2 (REQ-1, REQ-3): 글이 부르는 git 명령이 실제 git 에서 글대로 동작해요: tracked 파일은 `--error-unmatch` 0, 안 그러면 1; 규칙 `*.yaml` 은 profile 을 막고 `check-ignore -v` 가 그 규칙을 이름으로 주고; `/.moai/*` 아래에서는 `!` 한 줄이 풀고; 폴더 전체 규칙(`/.moai/`)에서는 풀지 못해 되돌리는 길이 필요해요.
  verify: `cargo test --locked --offline --test project_flow_closing the_git_commands_the_text_names_behave_as_the_text_says`
- AC-3 (REQ-5, REQ-6, REQ-7, REQ-8, REQ-9, REQ-10): `project.md` 의 마지막 단계 "Offer the closing commit" 가 `--untracked-files=all` 목록·단계별 묶음·"kept local"(백업·영수증 폴더)·선택·정확한 경로만 add·push 없음·되돌리기·비밀 이름(`.env*`, `*.key`, `*.pem`, `*credentials*`, `*secret*`, 경로 어디서든)·무응답·add/commit 실패·git 읽기 실패·프로젝트 언어 설정 커밋 메시지·Completion Check 제안 뒤의 순서를 말해요.
  verify: `cargo test --locked --offline --test project_flow_closing closing_commit_step_lists_by_apply_stages_only_those_paths_and_never_pushes`
- AC-4 (REQ-5, REQ-7): 글이 말하는 `status --porcelain=v1 -z` 읽기와 `git add -- <지워진 경로>` 가 실제 git 에서 삭제를 스테이징하고, 목록 밖의 변경(not from this route)은 스테이징하지 않아요.
  verify: `cargo test --locked --offline --test project_flow_closing staging_listed_paths_stages_deletions_and_leaves_other_changes_alone`
- AC-6 (REQ-5, REQ-8, REQ-12): `project.md` 의 Procedure 첫 문단 "Closing-commit snapshots" 가 첫 apply 전과 각 apply·"Track it" 편집 뒤의 목록 읽기, 새로 생기거나 상태가 바뀐 경로를 그 단계에 주기, 결과에서 경로를 읽지 않기, 이미 바뀐 경로를 어느 단계에도 주지 않기를 말해요. 실제 git 에서: 새 프로젝트의 기본 `status -z` 는 폴더를 한 항목(`?? .moai/`, `?? .moai-backup/`)으로 뭉치고 `--untracked-files=all` 은 파일마다 주며, 그 목록에 이름 규칙과 "kept local" 규칙을 적용하면 `.moai-backup/config/credentials.yaml`·`.moai/receipts/r.json` 은 빠지고 profile 만 스테이징돼요; 목록 차이가 각 단계의 새 경로를 정확히 주고 사용자의 먼저 바뀐 경로는 어느 단계에도 안 가요.
  verify: `cargo test --locked --offline --test project_flow_closing procedure_snapshots_the_status_around_every_apply`; `cargo test --locked --offline --test project_flow_closing untracked_folders_collapse_unless_every_file_is_asked_for`; `cargo test --locked --offline --test project_flow_closing snapshot_differences_give_each_new_path_to_the_step_that_wrote_it`
- AC-5 (REQ-1…REQ-12): 이웃 검사.
  verify: `cargo test --locked --offline --test agent_defs_package`; `cargo test --locked --offline --test update_route`; `cargo test --locked --offline --test project_decide_model`; `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --test ac_command_lint`; `cargo test --locked --offline --test git_env_scrub`; `cargo test --locked --offline --test product_cli source_manifest_covers_every_package_file`; `python3 tools/source_manifest.py check`; `cargo fmt --check`

## Scope

- write: `.moai/specs/SPEC-20261119-project-flow-closing/spec.md`, `tests/project_flow_closing.rs`, `skills/moai/workflows/project.md`, `skills/moai/references/source-manifest.json`
- read-only: `src/agents.rs`(관리 블록과 `left ignored by your rule` 줄), `tests/gitignore_block_cli.rs`, `skills/moai/references/internal/questions.md`
- forbidden: 실제 HOME, 사용자의 전역 git 설정, push, MCP 설정 파일

## Plan

1. 시험 먼저: `tests/project_flow_closing.rs` 의 글 시험 둘이 `project.md` 에 항목이 없어서 실패하는 것(RED)을 읽어요. git 동작 시험 둘은 임시 저장소에서 바로 통과해야 해요(글이 기대는 사실이므로).
2. `project.md` 의 단계 4 에 "Profile tracking" 항목을, 단계 8 로 "Offer the closing commit" 을 더해 GREEN 으로 만들고 매니페스트의 해시를 같은 커밋에서 바꿔요.
3. 이웃 검사(AC-5)를 돌려요.
4. 감사 FAIL(note 2666, B1·B2) 수리: 시험 먼저(RED: 글 시험 3개 실패, 새 git 시험 2개 통과) → Procedure 첫 문단과 단계 8·단계 4 의 `!` 줄 위치를 고쳐 GREEN → 변이 → 이웃 검사 → 같은 감사 에이전트 재감사.

## Risks

- C-64(route-end-offer)도 `project.md` 의 Completion Check 한 줄과 매니페스트의 `project.md` 해시를 바꿔요 → 이 가지는 C-64 의 head 위에 쌓아서(D-4) 해시 줄이 겹치지 않게 하고, 단계 8 은 그 줄과 떨어진 자리에 둬요.
- `!` 한 줄이 어떤 규칙에서는 풀지 못해요(부모 폴더 제외) → REQ-3 이 다시 확인과 되돌리기를 요구하고 AC-2 가 폴더 전체 규칙 경우를 직접 시험해요.
- 목록이 사용자의 다른 작업을 섞어 커밋해요 → REQ-5 가 "not from this route" 를 따로 두고 REQ-7 이 `-A`·`.` 를 금지해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | "Track it" 은 사용자 규칙을 지우지 않고 루트 `.gitignore` 의 끝(관리 블록 밖)에 `!<경로>` 한 줄을 더해요 | `src/agents.rs` 의 설명: 루트 `.gitignore` 의 `!` 줄이 `.git/info/exclude` 와 전역 제외 파일의 규칙보다 이겨요; 사용자 규칙 자체는 그대로; 감사(2026-11-19) 측정: 끝줄은 `check-ignore -q` 1, 생성 계획은 한 줄 병합(`ready`)이고 블록 안 줄은 어디에 둬도 `conflict` 는 안 났지만 블록 안의 줄이 나중 생성에도 남는지는 재지 못했어요 | 더한 한 줄을 지워요 |
| D-2 | 전역 제외 파일은 고치지 않아요 | 사용자 전역 설정은 `guarded` 예요(SKILL.md 사례 2) | 해당 없음 |
| D-3 | 라우트 끝 커밋은 로컬만이고 push 하지 않아요 | 되돌리기 쉬운 일만 세션이 해요; push 는 바깥으로 나가는 일이에요 | `git reset --soft HEAD~1` |
| D-5 | 변경 경로는 apply 결과가 아니라 단계 전후 `git status --porcelain=v1 -z --untracked-files=all` 목록의 차이로 단계에 줘요 | 감사 B2: `project-apply` 결과는 쓴 파일 이름을 하나도 안 주고 `go-settings-apply` 도 `.claude/settings.json` 을 이름으로 안 줘요; 감사 B1: 기본 `status` 는 untracked 폴더를 한 항목으로 뭉쳐 이름 검사가 안을 못 봐요 | 글의 Procedure 첫 문단과 단계 8 을 이전 문구로 되돌려요 |
| D-4 | 이 가지는 tip 이 아니라 C-64 의 LANE-READY head(`9078cccc41`, 가지 `spec/20261119-route-end-offer`) 위에 쌓아요 | 둘 다 매니페스트의 `project.md` 해시 한 줄을 바꿔서, tip 에서 갈라지면 묶음에서 그 줄이 충돌해요; C-64 가 먼저 묶이면 이 가지의 앞 두 커밋은 이미 있는 커밋이에요 | C-64 가 다시 만들어지면 이 가지를 새 head 로 옮겨요(커밋 둘만 다시 얹음) |

## Result

- AC-1: met — `cargo test --locked --offline --test project_flow_closing` `profile_tracking_step_states_the_checks_the_selection_and_the_undo` passes (RED 먼저: 4 중 2 failed, exit 101, `project.md` 에 항목이 없어서). `project.md` 단계 4 의 "Profile tracking" 이 tracked 확인 → ignored 확인 → 규칙 이름 → "Track it (recommended)"/"Keep it local" → 한 줄 추가와 다시 확인·되돌리기, git 오류(0도 1도 아님)·무응답·여전히 막힘의 처리를 말해요.
- AC-2: met — 같은 시험 파일 `the_git_commands_the_text_names_behave_as_the_text_says` passes (RED 때부터 통과: 글이 기대는 git 사실). 임시 저장소의 실제 git: untracked 는 `ls-files --error-unmatch` 1 과 `check-ignore -q` 1(stderr 없음), tracked 는 0, 규칙 `*.yaml` 은 `.gitignore:1:*.yaml<TAB>.moai/project.yaml` 로 이름이 나오고, `/.moai/*` 아래에서는 `!` 한 줄이 풀고, 폴더 전체 규칙 `/.moai/` 에서는 `!` 로도 못 풀며, 저장소 밖에서는 둘 다 0 도 1 도 아닌 종료와 메시지예요.
- AC-3: met — 같은 시험 파일 `closing_commit_step_lists_by_apply_stages_only_those_paths_and_never_pushes` passes. `project.md` 단계 8 "Offer the closing commit" 이 `git status --porcelain=v1 -z` 목록을 apply 별로 묶고("not from this route" 따로), "Commit these changes (recommended)"/"Leave them uncommitted", `git add -- <path>` 하나씩(`-A`·`.` 금지), push 안 함, `git reset --soft HEAD~1`, 비밀 이름 `.env*`·`*.key`·`*credentials*` 는 목록에만, add/commit 실패와 무응답 처리를 말해요.
- AC-4: met — 같은 시험 파일 `staging_listed_paths_stages_deletions_and_leaves_other_changes_alone` passes: `-z` 목록이 ` D`, ` M`, `??` 를 주고, 두 경로만 `git add --` 하면 `D`/`A` 가 스테이징되고 사용자의 수정 경로는 ` M` 으로 남아요.
- AC-5: met — 머리 `0e5bf8bc86` 에서(시험 파일의 문구 한 줄 추가 전): `agent_defs_package` 19, `update_route` 9, `project_decide_model` 4, `home_isolation_scan` 18, `ac_command_lint` 5, `git_env_scrub` 3, `test_isolation_rules` 3 passed, `product_cli source_manifest_covers_every_package_file` 1 passed, `python3 -I tools/source_manifest.py check` ok, `cargo fmt --check` 0, `cargo clippy --locked --offline --all-targets -- -D warnings` 0.

Mutation 5 mutants of `project.md` (committed file restored with `git restore` after each): M1 `never -A` → `also -A`, M2 `does not push` → `pushes`, M4 `*credentials*` → `*secret*`, M6 the "nothing more is asked" sentence dropped, M7 the unanswered-question rule replaced. First run 4 killed and M6 survived (the sentence was not pinned); the pin `Exit 0 means tracked and nothing more is asked` was added and the second run killed 5 of 5.
Not run: the full local `cargo test`; a real `/moai project` session reading these two passages (the text is an instruction to a session, and the tests pin its wording and the git behaviour it relies on, not how a session follows it); Linux; Windows. C-59 criteria 1 and 3 are linked to tip evidence by the leader's note, criterion 2 is not pursued (leader decision note 2528).
