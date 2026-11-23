---
id: SPEC-20261123-lane-env-own-target
title: "lane-env and heavy-slot-acquire leave the build folder to the worktree's own target/ and the optimize preview lists the leftover owner records"
status: draft
revision: 1
created: 2026-11-23
---

# lane-env and heavy-slot-acquire leave the build folder to the worktree's own target/ and the optimize preview lists the leftover owner records

## Intent

2026-11-22 에 사용자가 빌드 규칙을 바꿨어요. 빌드는 작업 폴더 안, 각 워크트리의 `target/` 에서 돌고(`CARGO_TARGET_DIR` 은 설정하지 않아요), 홈 아래 `~/.moai/build/<이름>/` 은 새 작업에 쓰지 않아요. 문서(`skills/moai/workflows/run.md`)와 `CLAUDE.md` 는 이미 이렇게 적고 있어요. 그런데 코드가 아직 옛 규칙이에요.

- `moai lane-env` 는 `CARGO_TARGET_DIR` 을 `~/.moai/build/<프로젝트>-<워크트리>` 로 내보내고, 그 폴더와 부모(`~/.moai`, `~/.moai/build`)를 만들고, 소유 표식 `.moai-build.json` 을 써요. 2026-11-23 에 `~/.moai/build` 에 항목이 136개 있고, 하나는 그날 아침 새로 생겼어요(143바이트짜리 표식 하나).
- `moai heavy-slot-acquire` 는 cargo 를 감쌀 때 `CARGO_TARGET_DIR` 이 옛 폴더가 아니면 "이 워크트리의 빌드 폴더는 `~/.moai/build/…` 이니 `lane-env` 를 먼저 돌려라" 하고 알려 줘요. 이 안내를 따르면 규칙과 반대로 홈에 쓰게 돼요.
- `moai optimize-plan` 은 홈 아래 옛 빌드 폴더를 읽기 전용으로 미리 보여 주지만, cargo 출력 없이 표식만 남은 폴더는 `not_cargo_output` 으로 건너뛰어서 "남은 표식" 으로 따로 보이지 않고, 복구 안내(`restore`)는 옛 규칙대로 그 폴더에 빌드하라고 해요.

이 SPEC 은 (1) `lane-env` 가 홈에 아무것도 만들지 않고 `CARGO_TARGET_DIR` 을 풀어 주게 하고, (2) `heavy-slot-acquire` 의 안내가 워크트리 자신의 `target/` 을 가리키게 하고, (3) 남은 표식을 읽기 전용 미리보기에서 한눈에 보이게 해요. 남은 항목을 실제로 지우는 일은 사용자가 고른 뒤에만 해요.

## Non-goals

- 남은 `~/.moai/build/*` 폴더나 표식을 지우지 않아요. 삭제는 사용자 선택(HOLD: deletion)이고, 이 SPEC 의 미리보기는 `writes: false` 예요.
- `src/worktree.rs` 의 `worktree-plan`/`worktree-apply` 가 `build_folder` 를 계획하고 만드는 일, `tools/build_folder.py` 와 그 Python 테스트는 바꾸지 않아요. 같은 옛 규칙의 다른 조각이라 따로 카드로 다뤄요(Result 의 후속).
- 카고 홈(`CARGO_HOME`)과 `CARGO_INCREMENTAL=0` 의 동작은 그대로예요.
- `CARGO_TARGET_DIR` 이 임의의 다른 폴더를 가리킬 때의 경고는 새로 만들지 않아요. 지금처럼 알려 줄 만한 경우만 알려요.

## Requirements

- REQ-1: When lane-env prints a lane's build environment, it shall print `unset CARGO_TARGET_DIR` and no `export CARGO_TARGET_DIR`, so cargo builds in the worktree's own `target/` · failure: none — a shell that evaluates the lines loses a value it already had.
- REQ-2: When lane-env runs inside a git worktree, it shall create nothing under HOME and need no usable HOME · failure: none — a folder outside a git worktree, a bad lane name or a relative base is refused with exit 1 and stdout empty, as before.
- REQ-3: When heavy-slot-acquire wraps cargo and CARGO_TARGET_DIR names the retired per-lane folder, a folder under the retired home build root `<HOME>/.moai/build` or a folder inside another worktree, it shall print one warning line on stderr that names the folder and the reason and says the worktree builds in its own `target/`, which unsetting the variable does · failure: none — a warning only, the wrapped command still runs, and an unset value, an empty value, this worktree's own `target/` and another command print nothing.
- REQ-4: When heavy-slot-acquire runs, it shall create nothing under `<HOME>/.moai/build` · failure: none — a HOME that is unset or not absolute only turns off the retired-home-folder reason of REQ-3.
- REQ-5: When optimize-plan lists the agent build folders under `<HOME>/.moai/build`, it shall mark a folder that holds only its owner record `marker_only` and offer it as a `trim` row with the record as its trim path when its worktree is gone or its branch merged · failure: none — a folder with other files, no record, another repository's record or a live worktree stays `skip` with its reason, and the plan writes nothing.
- REQ-6: When optimize-plan lists an agent build folder, it shall state in `restore` that nothing needs the folder because builds run in each worktree's own `target/` · failure: none — the row is the same one the preview already printed, only its text changes.

## Acceptance

- AC-1 (REQ-1): 임시 HOME 에서 `CARGO_TARGET_DIR` 이 옛 값으로 이미 정해진 셸이 `e=$(moai lane-env --lane agent-7 --base "$BASE") && eval "$e"` 를 하면 그 변수가 풀리고, 출력에 `export CARGO_TARGET_DIR` 줄이 없고, 카고 홈과 `CARGO_INCREMENTAL=0` 은 그대로예요 · verify: `cargo test --locked --offline --test lane_env_cli eval_unsets_cargo_target_dir_and_keeps_the_cargo_home`
- AC-2 (REQ-2): 워크트리·주 체크아웃·하위 폴더에서 돌린 뒤 임시 HOME 이 비어 있고, HOME 이 없어도 상대 경로여도 같은 출력으로 성공하고, git 밖 폴더나 나쁜 레인 이름·상대 base 는 stdout 이 빈 채 1 로 거절돼요 · verify: `cargo test --locked --offline --test lane_env_cli lane_env_makes_nothing_under_home` · verify: `cargo test --locked --offline --test lane_env_cli refusals_print_nothing_on_stdout`
- AC-3 (REQ-3): 옛 레인 폴더·홈 빌드 폴더(자기 것 포함)·다른 워크트리 안쪽 폴더는 한 줄 경고(폴더, 이유, `target/`, 변수 풀기)가 나오고 명령은 그대로 성공하고, 값이 없거나 비었거나 자기 `target/` 이거나 cargo 가 아닌 명령이면 아무 말도 없어요 · verify: `cargo test --locked --offline --test heavy_slot_cli a_cargo_run_is_warned_about_a_retired_or_foreign_build_folder` · verify: `cargo test --locked --offline --test heavy_slot_cli a_home_moai_link_is_followed_and_the_value_as_written_is_kept`
- AC-4 (REQ-4): 임시 HOME 에서 `heavy-slot-acquire` 로 cargo 를 감싸 돌려도(옛 홈 폴더를 값으로 줘도) HOME 아래에 아무것도 생기지 않고, HOME 이 없으면 홈 폴더 이유만 빠지고 다른 워크트리 이유는 그대로 나와요 · verify: `cargo test --locked --offline --test heavy_slot_cli heavy_slot_acquire_makes_nothing_under_home`
- AC-5 (REQ-5): 표식만 있는 폴더가 작업이 사라졌거나 병합된 때 `trim`(`marker_only: true`, trim 경로는 표식)이고, 작업이 살아 있거나 표식이 다른 저장소 것이거나 다른 파일이 있으면 `skip` 이에요. 미리보기는 아무것도 쓰거나 지우지 않고 `writes: false` 예요 · verify: `cargo test --locked --offline --test optimize_plan_cli a_folder_holding_only_the_owner_record_is_a_marker_only_trim_row`
- AC-6 (REQ-6): 모든 `agent_build_folder` 행의 `restore` 가 워크트리 자신의 `target/` 을 말하고 옛 `CARGO_TARGET_DIR` 안내가 없어요 · verify: `cargo test --locked --offline --test optimize_plan_cli agent_build_folder_rows_say_nothing_needs_the_folder`

## Scope

- write: `.moai/specs/SPEC-20261123-lane-env-own-target/**`, `src/lane_env.rs`, `src/heavy_slot.rs`, `src/optimize.rs`, `src/main.rs`(`lane-env` 도움말 문장), `tests/lane_env_cli.rs`, `tests/heavy_slot_cli.rs`, `tests/optimize_plan_cli.rs`
- write (병합 수리, 2026-11-23): 위에 더해 `.moai/specs/SPEC-20261120-test-parallelism-cap/spec.md`(REQ-4·AC-4·D5), `tests/never_index_test.rs`(머리 주석), `tests/next_flow_skill_scan.rs`(주석), `skills/moai/workflows/run.md`(경고 범위 한 문장), `docs/design-review/appendix/interfaces.md`(`lane-env` 문단)
- read-only: `src/worktree.rs`, `tests/worktree_build_root.rs`, `tools/build_folder.py`, `skills/moai/workflows/run.md`
- forbidden: 실제 HOME 에 쓰기, 홈 아래 옛 빌드 폴더 항목의 삭제, 새 Python 테스트

## Plan

1. RED: AC-1…AC-6 의 테스트를 먼저 쓰고 오늘 팁에서 실패하는 것을 확인해요(홈에 폴더가 생김, `export CARGO_TARGET_DIR` 이 나옴, 경고가 옛 폴더를 가리킴, 표식만 있는 폴더가 `skip`, `restore` 가 옛 안내).
2. GREEN: `src/lane_env.rs` 는 홈 쓰기(`prepare`, 소유 표식, 홈 확인)를 걷어내고 `unset CARGO_TARGET_DIR` 을 내보내요. `src/heavy_slot.rs` 의 `Here`·`build_folder_note` 는 워크트리 자신의 `target/` 을 기준으로 다시 써요. `src/optimize.rs` 는 `marker_only` 와 `restore` 문장을 더해요. `src/main.rs` 의 `lane-env` 도움말과 도움말 픽스처를 맞춰요.
3. 같은 이 테스트들과 옛 테스트(`lane_env_cli`, `heavy_slot_cli`, `optimize_plan_cli`, `worktree_build_root`)와 지킴 뮤턴트를 돌리고 마지막에 `moai check affected` 를 한 번 돌려요.

## Risks

- 레인이 새로 설치하기 전에는 옛 `lane-env` 가 계속 홈에 만들어요. 새 바이너리를 깔면 멈춰요.
- `CARGO_TARGET_DIR` 을 일부러 다른 곳으로 둔 호출자는 `lane-env` 를 평가하면 그 값을 잃어요. 규칙이 "설정하지 않는다" 라서 의도한 동작이에요.
- `src/worktree.rs` 의 `worktree-apply` 는 여전히 `<HOME>/.moai/build/<이름>` 폴더를 만들어요(같은 옛 규칙의 다른 조각). 이 SPEC 은 만지지 않아요.
- 이 SPEC 의 코드 경로에 지우기·프로세스 시작은 없어요. 미리보기는 읽기만 하고, 테스트가 쓰는 임시 HOME 밖에 쓰지 않아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `lane-env` 는 `CARGO_TARGET_DIR` 을 `<워크트리>/target` 으로 내보내지 않고 `unset` 해요. | `CLAUDE.md`("`<worktree>/target/` — do not set `CARGO_TARGET_DIR`")와 감사 2차 지적(`run.md` 는 변수를 비워 두기를 권해요)이에요. 내보내면 작업 폴더를 옮겨도 값이 따라가요. | 카드 열림 |
| D-2 | `heavy-slot-acquire` 의 경고 범위는 그대로 두고(옛 레인 폴더, 다른 워크트리 안쪽) 홈 빌드 폴더 전체를 옛 폴더로 더해요. 임의의 다른 폴더는 새로 경고하지 않아요. | 요청은 "옛 경로를 가리키는 것" 을 고치는 일이에요. 임의 폴더 경고는 새 동작이라 하지 않아요(Non-goals). | 카드 열림 |
| D-3 | 표식만 남은 폴더는 작업이 사라졌거나 병합된 때만 `trim` 으로 제안해요. | 이전 미리보기와 같은 증명 기준이에요. 살아 있는 작업의 폴더는 옛 바이너리가 아직 쓸 수 있어요. | 카드 열림 |
| D-4 | `src/worktree.rs` 와 `tools/build_folder.py` 는 이 SPEC 이 만지지 않고 후속으로 남겨요. | 카드의 쓰기 범위가 `lane_env`·`heavy_slot`·`optimize` 예요. | 후속 카드 |
| D-5 | `SPEC-20261120-test-parallelism-cap` REQ-4 의 `.metadata_never_index` 표식 중 빌드 폴더 쪽(`lane_env::prepare` 가 만들던 것)은 접고, 임시 HOME 쪽(`tests/common/home.rs`, `tests/never_index_test.rs`)만 남겨요. 홈 아래에는 표식도 만들지 않아요. 빌드 출력은 `.claude/worktrees/<이름>/target/` 에 있고, 그 자리는 앞의 점 때문에 Spotlight 색인 밖이에요(`SPEC-20261122-worktree-place` D-9, R-11). 같은 SPEC 이 이 표식은 최근 macOS 에서 더는 듣지 않는다고 적고 있어요. `target/` 안에 표식을 쓰는 안은 `lane-env` 가 일부러 하지 않는 쓰기를 새로 만들고 효과도 없어서 버렸어요. 이 머신(macOS, 2026-11-23T04:39Z)에서 `.claude/worktrees/<이름>/target/` 의 새 파일을 `mdls`·`mdfind -onlyin .claude` 로 봤을 때 색인 기록과 검색 결과가 없었어요(같은 날 저장소 루트 `Cargo.toml` 도 기록이 `(null)` 이라 이 한 번의 측정은 구별하는 증거가 못 돼요. 근거는 D-9 의 기록이에요). | `SPEC-20261122-worktree-place` D-9·R-11; 이 카드의 병합 수리(NeXT `6817eb69ad` 가 `prepare` 에 표식을 더해 `src/lane_env.rs` 가 충돌) | 나중 측정이 그 자리가 색인된다고 보이면 `target/` 이 만들어지는 자리에서 표식을 쓰는 카드를 따로 열어요 |
| D-6 | `heavy-slot-acquire` 는 홈 빌드 폴더를 링크를 푼 값과 쓴 그대로의 값(`.`·`..` 만 글자로 처리) 둘 다로 맞춰 봐요. 링크 때문에 한쪽만 맞는 값도 경고를 받고, 쓴 그대로 맞은 경우의 경고는 쓴 값과 쓴 뿌리를 보여 줘요. | 감사 비차단 N1·N2: 홈 안의 링크가 홈 밖을 가리키는 값은 풀린 값만 비교하면 놓쳐요. | 카드 열림 |

## Result

(status draft — the head that carries this Result is the one the READY names; 승인은 사용자 몫이라 상태는 draft 로 둬요)

- AC-1: met — `eval_unsets_cargo_target_dir_and_keeps_the_cargo_home` (옛 값을 쥔 셸이 eval 뒤 `unset`, `export CARGO_TARGET_DIR` 줄 없음, 카고 홈·`CARGO_INCREMENTAL=0` 그대로)
- AC-2: met — `lane_env_makes_nothing_under_home` (워크트리·주 체크아웃·하위 폴더, HOME 없음, 상대 HOME 에서 같은 출력, HOME 과 체크아웃 아래 새 항목 0), `refusals_print_nothing_on_stdout` (git 밖·나쁜 레인·상대 base 는 stdout 이 빈 채 1)
- AC-3: met — `a_cargo_run_is_warned_about_a_retired_or_foreign_build_folder` (옛 레인 폴더, 홈 빌드 폴더, 다른 워크트리 안쪽은 한 줄 경고, 값 없음·빈 값·자기 `target/`·상대 `target`·임의 폴더·cargo 아닌 명령은 조용함, HOME 없을 때 USERPROFILE 은 홈이 아님); 단위 테스트 `build_folder_tests`
- AC-4: met — `heavy_slot_acquire_makes_nothing_under_home` (HOME 이 있어도 없어도 임시 HOME 이 비어 있음)
- AC-5: met — `a_folder_holding_only_the_owner_record_is_a_marker_only_trim_row` (작업이 사라졌거나 병합됨 → `trim`·`marker_only: true`·trim 경로 `.moai-build.json`; 살아 있음·다른 저장소·다른 파일 → `skip`; 미리보기 전후 폴더 스냅숏이 같고 `writes: false`)
- AC-6: met — `agent_build_folder_rows_say_nothing_needs_the_folder`

증거 (이 가지의 팁에서 직접 돌린 것):

- 먼저 실패함: 새 테스트 8개(`eval_unsets…`, `lane_env_makes_nothing_under_home`, `lane_name_length_edge`, `the_comment_names_the_worktrees_own_target`, `a_cargo_run_is_warned…`, `heavy_slot_acquire_makes_nothing_under_home`, `a_folder_holding_only…`, `agent_build_folder_rows_say…`)가 고치기 전 코드에서 실패했어요.
- 뮤턴트 7개가 모두 죽었어요: `unset` 줄 삭제(lane_env_cli 2개 실패), `lane-env` 가 HOME 아래 폴더를 만듦(2개), 경고에서 홈 빌드 폴더 판정 삭제(heavy_slot_cli 2개), 경고에서 다른 워크트리 판정 삭제(1개), `Here::of` 가 홈 아래 폴더를 만듦(1개), `marker_only` 끄기(1개), `restore` 문장을 옛 안내로 되돌림(1개).
- 위험 경로 실행: 링크가 낀 임시 HOME(`.moai` 가 다른 폴더로의 링크)에서 빌드한 바이너리로 `lane-env`(링크를 따라 쓰지 않고 성공, 바깥 폴더 그대로), `heavy-slot-acquire`(링크 뒤 폴더를 값으로 줘도 경고만, 아무것도 안 만듦), `optimize-plan`(링크 행은 `skip: symlink`, 표식만 있는 폴더는 `trim` 과 `skip: worktree_live`, 실행 뒤 폴더 그대로)를 직접 돌렸어요.
- 실제 HOME 읽기 전용 미리보기(`optimize-plan`, `writes: false`): 홈 아래 에이전트 빌드 폴더 행 67개 중 표식만 남은 폴더가 12개(`trim` 2, `skip` 10)예요. 삭제는 하지 않았고 사용자 선택을 기다려요.

후속 (이 SPEC 이 만지지 않은 것):

- `src/worktree.rs` 의 `worktree-plan`/`worktree-apply` 가 아직 `<HOME>/.moai/build/<이름>` 을 계획하고 만들어요(`SPEC-20261121-shared-build-root` REQ-1·REQ-2, `tests/worktree_build_root.rs`).
- `tools/build_folder.py env` 와 `tests/build_folder_test.py` 는 옛 표식을 계속 써요(Rust 전용 테스트 카드가 정리해요).
- `skills/moai/workflows/update.md` 의 홈 빌드 폴더 문장은 `marker_only` 행을 아직 말하지 않아요.
- 레인이 새로 설치하기 전에는 옛 `lane-env` 가 계속 만들어요.
