---
id: SPEC-20261123-statusline-integration-distance
title: "The status line measures the integration branch distance with its own git call, whatever upstream the branch has"
status: draft
revision: 1
created: 2026-11-23
---

# The status line measures the integration branch distance with its own git call, whatever upstream the branch has

## Intent

상태 줄의 `🧭 <통합 브랜치> ↑n ↓m` 과 `🌳`/`🅱️` 뒤의 `↑n` 은 지금 가지가 통합 브랜치보다 몇 커밋 앞서고 뒤처졌는지 보여 줘요. 그런데 가지가 자기 upstream 을 가지고 있으면(`git push -u` 로 올린 레인 가지가 그래요) 숫자가 `↑0 ↓0` 으로 틀리게 나와요.

### 원인 (이 저장소의 팁 `a502f13229` 에서 읽고 git 2.54 로 확인했어요)

- `src/statusline/state.rs` 의 `status_args` 는 `git -c branch.<가지>.remote=. -c branch.<가지>.merge=refs/heads/<통합> status --porcelain --branch` 한 번으로 통합 브랜치 기준 숫자를 읽어요. upstream 이 없는 가지에서는 덮어쓴 두 값이 설정의 전부라서 맞아요.
- 그러나 `branch.<가지>.merge` 는 여러 값을 가질 수 있는 키라서 `-c` 가 기존 값을 바꾸지 않고 뒤에 더해요. 가지에 이미 `merge=refs/heads/<같은 이름>` 이 있으면 값이 둘이 되고, git 은 첫 값(저장소 설정의 값)을 upstream 으로 써요. `remote` 만 마지막 값(`.`)으로 바뀌니까 upstream 은 가지 자신이 돼요. `git status --porcelain --branch` 가 `## <가지>...<가지>` 를 찍고 숫자는 모두 0 이에요.
- 재현: 가지 `lane` 을 `next` 에서 따서 `branch.lane.remote=origin`, `branch.lane.merge=refs/heads/lane` 을 설정하고 `next` 보다 2 커밋 앞서고 1 커밋 뒤처지게 하면 `git rev-list --left-right --count next...lane` 은 `1 2` 인데 상태 줄은 `🧭 next ↑0 ↓0` 이에요.
- `-c` 로는 이미 있는 여러 값 키를 지울 수 없어요(`branch.<가지>.merge` 에는 `credential.helper` 처럼 빈 값으로 목록을 비우는 규칙이 없어요). 설정을 덮어쓰는 길은 고칠 수 없는 길이에요.

### 고치는 방향

통합 브랜치 거리는 upstream 설정에 기대지 않고 `git rev-list --left-right --count refs/heads/<통합>...HEAD` 로 따로 재요. 이 호출은 가지의 설정을 읽지 않아요. 더러움 숫자(`+n`)를 읽는 `git status --porcelain` 은 그대로 한 번 부르고 `-c` 덮어쓰기를 없애요.

## Non-goals

- 통합 브랜치 자신, 통합 브랜치를 모르는 경우, 분리된 `HEAD` 의 표시 변경. 이들은 지금처럼 git 자신의 upstream 숫자(또는 숫자 없음)를 보여요.
- 세션 캐시의 키·나이 규칙, 갱신 간격(10초), 렌더 예산의 변경.
- `gh` 새로 고침, 원격 추적 가지(`refs/remotes/...`) 기준 거리.
- 통합 브랜치가 로컬에 없을 때 원격 쪽 ref 로 대신 재는 길. 지금도 0 이고 계속 0 이에요.

## Requirements

<!-- This SPEC changes REQ-7 of SPEC-20261122-statusline-where from "at most one git process" to "at most two"; REQ-1 to REQ-6 here keep every other rule of that SPEC. -->
- REQ-1: While the checked-out branch is not the integration branch, both names are safe to hand to git and the local ref of the integration branch exists, the status line shall measure the branch against the integration branch with `git rev-list --left-right --count refs/heads/<integration>...HEAD` and use its two numbers as the commits ahead and behind, whatever `branch.<name>.remote` and `branch.<name>.merge` the branch has.
- REQ-2 (REQ-1): IF the `rev-list` call fails or the render deadline has passed, THEN the status line shall show zeros for the distance, count no process for it that was not started, and write no session cache record for that render.
- REQ-3: While the checked-out branch is the integration branch, no integration branch is known, the `HEAD` is detached or a name is not safe to hand to git, the status line shall show the numbers of the `git status --porcelain --branch` header as it did before this SPEC.
- REQ-4: The status line shall start the `git status` call without any `-c` override of a `branch.<name>.*` key.
- REQ-5: While the local ref of the integration branch does not exist, the status line shall show zeros for the distance and start no `rev-list` process · failure: none — the ref is read from the files before any process starts, so a missing ref costs no process.
- REQ-6: The status line shall start at most two git processes in one render and none while the session cache is fresh (REQ-6 of SPEC-20261122-statusline-where, unchanged) · failure: none — the second process is the `rev-list` of REQ-1 and starts only when REQ-1 applies.

## Acceptance

- AC-1 (REQ-1, REQ-4): a lane branch with its own upstream (`remote=origin`, `merge=refs/heads/<same name>`) that is 2 commits ahead and 1 behind the integration branch shows `🧭 next ↑2 ↓1` and `↑2` on the branch item; the test failed before the change with `↑0 ↓0` · verify: `cargo test --offline --locked --test statusline_integration_distance a_branch_with_its_own_upstream_is_measured_against_the_integration_branch`
- AC-2 (REQ-1): a branch without an upstream, a branch whose upstream names another branch, and a branch whose upstream names the integration branch itself all show the same `↑2 ↓1`; a branch level with the integration branch shows a real `↑0 ↓0` · verify: `cargo test --offline --locked --test statusline_integration_distance`
- AC-3 (REQ-3): the integration branch itself shows no `🧭` item and the branch item shows git's own numbers, with and without an own upstream · verify: `cargo test --offline --locked --test statusline_integration_distance the_integration_branch_itself_keeps_its_own_display`
- AC-4 (REQ-5): with the local ref of the integration branch missing the distance is `↑0 ↓0` and the render starts one process · verify: `cargo test --offline --locked --test statusline_integration_distance an_integration_branch_that_is_missing_locally_measures_nothing` and `cargo test --offline --locked --test statusline_git_env head_file_and_process_count`
- AC-5 (REQ-2, REQ-6): with the scripted `git` of `statusline_git_env` the render of a lane branch starts the `status` call and the `rev-list` call and no more; a second render within ten seconds starts none; a failed `rev-list` writes no cache record; a spent budget starts none · verify: `cargo test --offline --locked --test statusline_git_env`
- AC-6 (REQ-1, REQ-3, REQ-4): the arguments of both calls are the exact lists the unit tests pin (the integration branch only as the full ref `refs/heads/<name>`), a name git could misread never reaches git, and a name of 100 bytes does while one of 101 does not · verify: `cargo test --offline --locked --lib statusline::state::where_tests::`
- AC-7 (REQ-6): the measured cost of the added process and the render times before and after are in the Result · verify: the commands written in the Result

## Scope

- write: `.moai/specs/SPEC-20261123-statusline-integration-distance/**`, `src/statusline/state.rs`, `src/statusline/render.rs` (comments only, if a comment names the one process), `tests/statusline_integration_distance.rs` (new), `tests/statusline_git_env.rs` (the two cases that pinned the one process and the `-c` arguments)
- read-only: `src/statusline/segments.rs`, `src/profile.rs`, `.moai/specs/SPEC-20261122-statusline-where/spec.md`
- forbidden: the session cache rules, the `gh` refresh, any skill or workflow text, `tests/statusline_golden.rs` goldens

## Plan

1. RED: `tests/statusline_integration_distance.rs` against the real `git` (done first; AC-1 fails at the tip).
2. `src/statusline/state.rs`: `status_args` takes no names and returns the plain `status` list (with `--branch` only when no distance is measured); a new `distance_args` builds the `rev-list` list; `git_counts` starts the second process after the first, only when REQ-1 applies, and keeps the cache rule of "a git that cannot answer writes nothing".
3. Update the pinned argument lists and the scripted-git cases of `tests/statusline_git_env.rs`.
4. The edit of REQ-7 of SPEC-20261122-statusline-where is a note in its Result at the closure of this SPEC; this SPEC only states the new bound.
5. Measure the two calls on this repository and the whole render with the built binary (Result), run the real status line in a temporary repository for a branch with an upstream, a branch without one and the integration branch.

## Risks

- R-1: One more process per cache miss. A cache miss happens when the `HEAD` text, the integration commit, the branch commit or the index changes, or after ten seconds; a render with a fresh cache still starts none. The measured cost is in the Result. `git rev-list --left-right --count` reads no work tree and no index, so it is the cheaper of the two calls.
- R-2: The integration branch only on the remote (no local ref) shows `↑0 ↓0`, as it did when the override pointed at a missing ref (`[gone]`). REQ-5 keeps that and costs no process.
- R-3: `rev-list` on a very old fork point walks many commits. The render deadline (the existing `--render-budget-ms`) bounds the process, and a spent budget shows zeros (REQ-2).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A second process replaces the `-c` override. | `-c` cannot remove a value a multi-valued key already has, so no single `git status` can be pointed at the integration branch when the branch has its own upstream (reproduced with git 2.54). `status --porcelain=v2` also reads the branch's own upstream. `for-each-ref --format=%(ahead-behind:refs/heads/<integration>)` reads no upstream and works on the git Apple ships (Apple Git-157, git 2.54.0, run 2026-11-23T11:39Z: `NeXT 166 0`), but it gives only the two numbers: the dirty count still needs `git status`, so it is a second process too and costs no less than `rev-list`. | Go back to the override and accept the wrong numbers on a branch with an upstream. |
| D-2 | Always the second process when REQ-1 applies, not a "status first, second call only when the header looks wrong" check. | The check would read `git`'s header format to decide, a second path to test and to keep in step with git versions; the second call costs a few milliseconds and the cache hides it for ten seconds. | Add the header check back later if a measurement shows the cost matters. |
| D-3 | A failed `rev-list` shows zeros and writes no cache record. | The same rule the one process had: "a git that cannot answer gives zeros and writes nothing". A missing ref never reaches this branch (REQ-5). | Fall back to the numbers of the `status` header. |

## Result

Status: implemented on the branch `spec/20261123-statusline-integration-distance`; the SPEC front matter is not changed here (approval is recorded by the user's pick, not by a lane).

### 바뀐 것

- `src/statusline/state.rs`: `status_args` 의 `-c branch.<가지>.*` 덮어쓰기를 없앴어요. 새 `measure` 가 가지·통합 브랜치 이름과 통합 브랜치의 로컬 ref 유무로 `Upstream`, `Integration`, `Missing` 중 하나를 고르고, `git_counts` 가 `Integration` 일 때 `git status --porcelain` 다음에 `git rev-list --left-right --count refs/heads/<통합>...HEAD` 를 불러요. 답이 두 숫자가 아니거나 호출이 실패하면 거리는 0 이고 캐시를 쓰지 않아요.
- `tests/statusline_integration_distance.rs`(새 파일, 진짜 git·진짜 바이너리), `tests/statusline_git_env.rs`(스크립트 git 이 `rev-list` 에도 답하고 프로세스 수·인자를 새 규칙으로 고정), 단위 테스트는 `state.rs` 의 `where_tests`.

### 렌더당 git 프로세스 (REQ-6, 이 SPEC 이 SPEC-20261122-statusline-where 의 REQ-7 을 바꾸는 부분)

- 전: 캐시가 없을 때 한 프로세스(`status`). 후: 통합 브랜치와 거리를 재는 레인 가지에서만 두 프로세스(`status` 와 `rev-list`). 통합 브랜치 자신·통합 브랜치를 모름·분리된 `HEAD`·로컬 ref 없음은 그대로 한 프로세스 이하고, 캐시가 신선한 렌더는 0 이에요.
- 이유: D-1 — `-c` 로는 이미 있는 `branch.<가지>.merge` 값을 지울 수 없어서(git 2.54 에서 재현) 한 `git status` 로는 upstream 이 있는 가지의 통합 브랜치 거리를 읽을 수 없어요.
- 측정 (git 2.54.0 Apple Git-157, 이 저장소의 워크트리, 기계 부하 평균 약 60–76, 40회 중앙값; 명령은 스크래치패드의 `measure.zsh`: `git -C <워크트리> --no-optional-locks status --porcelain --branch` 와 `git -C <워크트리> rev-list --left-right --count refs/heads/v2026.11.0...refs/heads/spec/20261122-next-flow-hooks-r2` — 후자는 34 커밋 앞선 가지, 답 `0	34`): `status` 79.3 ms(다른 실행에서 92.7 ms), `rev-list` 34.4 ms. 그래서 캐시 미스 한 번에 약 34 ms 가 더해져요. 캐시 미스는 `HEAD`·통합 커밋·가지 커밋·인덱스가 바뀌거나 10초가 지난 뒤에만 와요.

### 실제 상태 줄 출력 (임시 저장소, 같은 스크립트로 전·후 한 번씩)

저장소: `next`(통합 브랜치)에서 딴 `lane-up`(자기 upstream `origin`/`refs/heads/lane-up` 설정, 2 커밋 앞서고 `next` 가 1 커밋 더 나아감)과 같은 모양의 `lane-plain`(upstream 없음). 진짜 git 은 `git rev-list --left-right --count next...lane-up` = `1 2`(뒤처짐 1, 앞섬 2). `+1` 은 임시 저장소의 추적 안 하는 `.moai-cli/project.yaml` 이에요.

| 가지 | 전(설치된 `moai-cli`, 팁 `a502f13229`) | 후(이 가지의 바이너리) |
| --- | --- | --- |
| `lane-up` (자기 upstream 있음) | `🧠 Opus \| 🅱️ lane-up +1 \| 🧭 next ↑0 ↓0` | `🧠 Opus \| 🅱️ lane-up ↑2 +1 \| 🧭 next ↑2 ↓1` |
| `lane-plain` (upstream 없음) | `🧠 Opus \| 🅱️ lane-plain ↑2 +1 \| 🧭 next ↑2 ↓1` | `🧠 Opus \| 🅱️ lane-plain ↑2 +1 \| 🧭 next ↑2 ↓1` |
| `next` (통합 브랜치 자신) | `🧠 Opus \| 🅱️ next +1` | `🧠 Opus \| 🅱️ next +1` |

### 검증 (이 가지의 머리에서)

- 빨강: 고치기 전 `cargo test --offline --locked --test statusline_integration_distance` 는 `a_branch_with_its_own_upstream_is_measured_against_the_integration_branch` 와 `an_upstream_that_names_another_branch_does_not_change_the_distance` 가 `🧭 next ↑0 ↓0` 으로 실패(7개 중 5 통과 2 실패)했어요.
- 초록: 같은 명령과 `--test statusline_git_env` 가 7 + 27 통과, `cargo test --offline --locked --lib statusline::` 212 통과, `cargo clippy --offline --locked --all-targets -- -D warnings` 와 `cargo fmt --check` 종료 0.
### 착지 뒤 정정 (2026-11-23)

- 화면: `SPEC-20261123-statusline-order-bars` REQ-11 이 첫 줄에서 `🧭`·`↑ ↓`·`+N` 을 뺐어요. 위 "실제 상태 줄 출력" 표는 착지 때의 값이고 지금 화면에는 없어요. 계산 코드(`measure`, `distance_args`, `git_counts` 의 `rev-list`)는 그대로 돌아요(order-bars Non-goals).
- 고친 것: 범위 뒤에 `--` 가 없어서 작업 트리에 `refs/heads/<통합>...HEAD` 라는 파일이 있으면 git 이 `ambiguous argument` 로 멈췄어요. 그러면 거리가 0 으로 정해지고 세션 캐시가 쓰이지 않아 렌더마다 git 이 두 번 돌아요. `distance_args` 끝에 `--` 를 더했어요(`fix(statusline-integration-distance)` 커밋 `173d271ac7`).
- 빨강: 고치기 전 `cargo nextest run --offline --locked --test statusline_integration_distance -E 'test(a_file_named_like_the_range_does_not_stop_the_distance_call)'` 가 exit 100(캐시 기록 없음). 초록: 고친 뒤 `--lib --test statusline_integration_distance --test statusline_git_env` 의 `statusline::state::`·두 테스트 파일 119개가 모두 통과, exit 0. 뮤턴트(`--` 를 뺌)에서 같은 새 테스트가 exit 100 으로 실패하고 되돌렸어요.
- 테스트 이름: 단언은 캐시 미스 한 번에 두 프로세스(2→4)인데 이름이 `starts_one_git` 이던 `session_cache_a_changed_integration_commit_starts_one_git` 을 `…starts_git_again` 으로 바꿨어요.

### 수용 기준별 결과

- AC-1: met — 착지 때 `a_branch_with_its_own_upstream_is_measured_against_the_integration_branch` 가 `↑2 ↓1` 로 통과했어요(위 "검증"). 숫자가 화면에서 빠진 뒤에는 이 테스트가 `a_branch_with_its_own_upstream_shows_the_integration_tip_and_no_distance` 로 바뀌어 첫 줄에 통합 항목과 가지만 있는 것을 고정하고, 거리 숫자 자체는 세션 캐시 기록(앞섬 2·뒤처짐 1)으로 `a_file_named_like_the_range_does_not_stop_the_distance_call` 이 봐요(upstream 이 있는 가지는 아니에요).
- AC-2: met — 착지 때 upstream 없음·다른 가지 upstream·통합 브랜치 upstream·같은 높이가 모두 같은 값으로 통과했어요. 지금 같은 사례들은 `a_branch_without_an_upstream_shows_the_same_line` 등 같은 파일의 테스트가 첫 줄이 같은 것을 고정해요.
- AC-3: withdrawn — 이 기준은 통합 브랜치 자신에서 `🧭` 가 없고 가지 항목에 git 자신의 숫자가 나오는 것이었는데, 그 숫자가 화면에서 빠져 볼 수 있는 출력이 없어요. 옛 테스트는 원격 없이 숫자 0 이어도 통과해서 실제 원격 사례를 보지 못했고(후속 지적), 철회로 정리했어요(리더 결정 2026-11-23T11:32Z). 통합 브랜치 자신의 첫 줄은 `the_integration_branch_itself_shows_both_items` 가 고정해요.
- AC-4: met — 착지 때 통과. 지금은 `statusline_integration_distance` 의 `an_integration_branch_that_is_missing_locally_says_n_a` 와 `statusline_git_env` 의 `a_missing_integration_ref_starts_no_distance_process` 가 고정해요.
- AC-5: met — 착지 때 `--test statusline_git_env` 통과. 지금 스크립트 git 의 `rev-list` 호출은 `rev-list --left-right --count refs/heads/next...HEAD --` 이에요.
- AC-6: met — 단위 테스트가 두 호출의 인자 목록을 그대로 고정하고, 지금 `rev-list` 목록은 끝에 `--` 가 있어요(`a_lane_branch_is_measured_against_the_integration_branch_by_its_own_call`, 위 119개 안).
- AC-7: met — 위 "측정"(착지 때)에 두 호출의 시간이 있어요. 이번 정정은 인자 하나를 더했을 뿐이라 다시 재지 않았어요(not_run).
