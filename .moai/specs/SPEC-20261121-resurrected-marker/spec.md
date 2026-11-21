---
id: SPEC-20261121-resurrected-marker
title: "A valid Rust marker keeps the project Rust-owned when git brings the Go markers back, and the update plan offers to keep them instead of stopping"
status: draft
revision: 1
created: 2026-11-21
---

# 되살아난 Go 표식이 있어도 유효한 Rust 표식이면 Rust 가 주인이고, update 계획은 멈추지 않고 표식 보관을 제안한다

## Interview

- Q1: 무엇이 문제였나요? → A: 사용자 지시(2026-11-21)로 조사한 보고서(`go-generation-update-2026-11-21.md` 4–8절)에서, Go 세대에서 옮긴 프로젝트가 `main` 을 동기화하자 git 이 Go 표식 파일(`manifest.json`, `config/config.yaml`, `config/sections/system.yaml`)을 되살렸어요. `.moai/moai.json`(Rust 표식)이 유효한데도 분류가 `mixed` 가 되어, 상태 폴더가 옛 `.moai-cli` 로 읽히고 모든 `card` 명령이 "project profile missing" 으로 실패했어요. update 계획도 `mixed` 를 명령 없는 `blocked` 로 막았어요.
- Q2: 어디서 그렇게 되나요? → A: `state_root::classify` 가 `("valid", false)` 를 `mixed` 로 돌려주고(`src/state_root.rs`), `state_root::dir` 는 `rust` 일 때만 `.moai` 를 골라요. `project_update` 의 go-state 단계는 `mixed|unknown|refuse` 를 `blocked` 로 바꿔요(`src/project_update.rs`).
- Q3: 무엇을 고치나요? → A: (1) 유효한 `moai.json` 이 있으면 되살아난 Go 표식이 있어도 상태 폴더는 `.moai` 예요(Rust 가 주인). 분류 이름 `mixed` 는 그대로 두어 go-state·home 쪽 동작이 변하지 않아요. (2) update 계획에 `go-residue` 단계를 두고, 그 틀에서는 `needs_decision` 으로 되살아난 표식을 나열하며 보관을 제안해요(`blocked` 아님). 보관 이동 자체는 다음 조각(적용·되돌리기)의 일이에요.
- Q4: 옛 이름 표식(`moai-cli.json`)이나 `moai.json` 이 없을 때는요? → A: 지금과 같아요. `moai.json` 없이 Go 표식만 있으면 `go`, 옛 이름이거나 유효하지 않은 표식이면 `mixed`/`refuse` 로 멈춰요(이 SPEC 은 그 경로를 건드리지 않아요).
- Q5: 이번에 안 하는 것은요? → A: 표식을 실제로 옮기는 것, 되살아난 Go 파일 222개 같은 표식 밖 잔여물 감지(앞 조각), 홈 쪽(`classify_home`) 판정, `update_tux` 의 화면 문구는 건드리지 않아요.

## Intent

Rust 표식이 유효한 프로젝트는 Rust 세대가 주인이에요. git 이 Go 파일을 되돌려 놓았다고 해서 프로젝트 상태를 못 읽거나 update 가 "아무것도 할 수 없다"고 막으면, 사용자는 손으로 표식을 지워야 해요. 이 SPEC 은 주인 판정을 표식 하나(유효한 `moai.json`)에 두고, 되살아난 표식은 보관 제안으로 다뤄요.

## Non-goals

- 표식 파일을 옮기거나 지우는 일(제안만, 적용·되돌리기는 다음 조각), 사용자 데이터를 읽거나 쓰는 일.
- 분류 이름·`classify` 반환값의 변경(`mixed` 는 그대로). go-state, home 쪽 코드, `update_tux`, `hook` 경로.
- 표식 밖 Go 잔여물의 감지(앞 조각이 같은 `go-residue` 단계에 키를 더해요).
- 옛 이름 표식(`moai-cli.json`)과 유효하지 않은 표식의 처리.

## Requirements

- REQ-1: When the project's `.moai/moai.json` is a valid marker and Go markers are present beside it, `state_root::dir` shall name `.moai` as the state folder · failure: none
- REQ-2 (REQ-1): IF the Rust marker is not a valid `moai.json`, THEN `state_root::dir` shall name the folder it named before this change.
- REQ-3: When a card command runs in a project whose Rust marker is valid and whose Go markers came back, the command shall read the project profile from `.moai/project.yaml` · failure: none
- REQ-4: When `project-update-plan` runs in a project whose Rust marker is valid and whose Go markers came back, the plan shall hold a step named `go-residue` in the state `needs_decision` that lists each Go marker with whether git tracks it and names the folder that would keep them · failure: none
- REQ-5 (REQ-4): IF the Rust marker is valid and Go markers came back, THEN the `go-state` step of the plan shall be `skip` with a reason that names `go-residue`.
- REQ-6: When the plan is made, the project shall stay byte for byte as it was, with no file, folder or time changed · failure: none
- REQ-7 (REQ-4): IF Go markers are present and no valid `moai.json` exists, THEN the classification shall stay `go`, with no `go-residue` step in the plan.
- REQ-8: While a resurrected marker is listed in `go-residue`, the plan's `all_unchanged` shall be false · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2): unit tests in `src/state_root.rs`: a project with a valid `moai.json` and each Go marker (`manifest.json`, `config/config.yaml`, `config/sections/system.yaml` with a version) has `dir` = `.moai`; without `moai.json`, with the older `moai-cli.json`, with an unreadable or newer `moai.json`, `dir` is as before; the verdict names (`mixed`, `go`, `refuse`) are unchanged.
  verify: `cargo test --offline --locked --lib state_root`
- AC-2 (REQ-3): `tests/state_root_resurrected_cli.rs` builds a temp project shaped like the field case (git repo, every file committed, `.moai/moai.json` valid, `.moai/project.yaml`, then `manifest.json` and `system.yaml` brought back by `git checkout`, no earlier state folder): `card list` and `card show` succeed, and the same project with the Go markers and no `moai.json` still fails the way it does now.
  verify: `cargo test --offline --locked --test state_root_resurrected_cli`
- AC-3 (REQ-4, REQ-5, REQ-8): the same file: `project-update-plan` reports state_root `mixed`, a `go-residue` step in `needs_decision` with the two markers, `tracked` true for both, and the keep folder; the `go-state` step is not `blocked`; `all_unchanged` is false.
  verify: `cargo test --offline --locked --test state_root_resurrected_cli`
- AC-4 (REQ-6, REQ-7): the same file: the project snapshot before and after the plan is equal; with Go markers and no `moai.json` the verdict is `go`, the plan has no `go-residue` step and the `go-state` step is what it was.
  verify: `cargo test --offline --locked --test state_root_resurrected_cli`
- AC-5: at most five mutants of the new code, each turning a test red, are restored.
  verify: manual: the mutant table in the Result.

## Scope

- write: `.moai/specs/SPEC-20261121-resurrected-marker/**`, `src/state_root.rs`, `src/project_update.rs`, `tests/state_root_resurrected_cli.rs`
- read-only: `src/go_state.rs`, `src/profile.rs`, `src/ledger/open.rs`, `tests/common/home.rs`
- forbidden: the real HOME, the project the field check came from, any move or delete of a marker
- 읽은 사실(2026-11-21, tip `53cce6c580`): `classify` 는 `("valid", false)` 를 `mixed`(`src/state_root.rs`)로 돌려주고, `dir` 는 `generation == "rust"` 일 때만 `.moai` 를 골라요. `project_update` 는 `mixed|unknown|refuse` 를 go-state 단계에서 `blocked` 로 바꿔요(`src/project_update.rs`, `go_state_step` 앞부분). `go-residue` 단계는 코드에 아직 없어요(`grep -rn go-residue src` 결과 없음). 홈 쪽 `classify_home` 과 `ledger::open::check_home_ready` 는 HOME 의 `.moai` 를 보는 별개의 판정이에요.

## Plan

1. **RED**: `tests/state_root_resurrected_cli.rs` 와 `state_root` 단위 시험을 먼저 써서 빨갛게 확인해요(AC-1~AC-4).
2. **GREEN**: `Classified::rust_owned()`(유효한 `moai.json` 이고 `generation` 이 `rust` 이거나 `mixed`)를 두고 `state_root::dir` 가 그것을 따르게 해요. `project_update::plan` 에 `go-residue` 단계를 더하고, 그 판정이 서면 `go-state` 단계를 `blocked` 대신 `skip` 으로 두어요.
3. 변형 5개 이하, 이웃 검사(`check affected`)를 돌려 Result 에 적어요.

## Decisions

각 결정은 (무엇 · 선택지 · 기본값 · 되돌릴 수 있음 · 범주 · 기한 · 결정자) 순서예요. 모두 제안이고 승인은 SPEC 이 병합될 때 사용자·리더가 골라요.

- D1 (제안, 되돌릴 수 있음, 범주 none): 새 분류 이름을 만들지 않고 `mixed` 를 그대로 둬요. 대안: 새 이름 `rust-resurrected`. 기본값: `mixed` 유지. 이유: `go_state`, `home_retire`, `update_tux` 가 `mixed` 를 이미 읽어요.
- D2 (제안, 되돌릴 수 있음, 범주 none): 주인 판정은 유효한 `moai.json` 하나예요. 옛 이름 `moai-cli.json` 은 주인으로 보지 않아요. 대안: 두 이름 모두. 기본값: `moai.json` 만. 이유: 카드 지시와 보고서 AC 가 `moai.json` 으로 적혀 있어요.
- D3 (제안, 되돌릴 수 있음, 범주 none): `go-residue` 단계를 이 SPEC 이 만들되 키는 `markers`·`keep_in` 만 둬요. 대안: 앞 조각이 올 때까지 기다림. 기본값: 지금 만듦. 이유: 인수 기준이 단계 이름을 요구하고, 앞 조각은 같은 단계에 키만 더해요.
- D4 (제안, 되돌릴 수 있음, 범주 none): 단계 상태는 `needs_decision` 이고 실행 명령은 붙이지 않아요(적용은 다음 조각). 대안: `ready` + 명령. 기본값: 명령 없음.

## Risks

- `state_root::dir` 를 바꾸면 쓰는 쪽(`profile::target`, 장부 경로)도 `.moai` 로 가요. 되살아난 표식이 있는 프로젝트에서 새 쓰기는 `.moai` 로 간다는 뜻이고, 이전에는 막혀 있었던 경로예요. 시험으로 고정해요.
- 앞 조각이 같은 `go-residue` 단계를 만들면 합칠 때 키 이름이 겹칠 수 있어요. 이 SPEC 의 키는 `markers`·`keep_in` 뿐이고, 합치기는 늦게 오는 쪽이 해요.
- `clone_root::dir` 와 `update_tux` 는 이 SPEC 에서 바뀌지 않아 되살아난 표식이 있는 프로젝트에서 그쪽은 이전처럼 `mixed` 로 읽어요.

## Result

전반(상태 폴더)은 끝났고, 후반(`go-residue` 단계)은 `src/project_update.rs` 를 다른 조각이 쥐고 있어 아직이에요. 아래는 이 레인이 직접 돌려 읽은 결과예요(2026-11-21, `--max 3`, `RUST_TEST_THREADS=4`).

- AC-1 met: `cargo test --locked --offline --lib state_root` → 26 passed (실행 줄 `test result: ok. 26 passed; 0 failed`, 단위 시험 `a_valid_rust_marker_owns_the_state_folder_when_go_markers_come_back` 포함).
- AC-2 met: `cargo test --locked --offline --test state_root_resurrected_cli` → 3 passed, 1 ignored (`card_commands_work_when_go_markers_came_back_beside_a_valid_marker`, `go_markers_without_a_valid_marker_keep_failing_the_way_they_do`, `the_plan_writes_nothing`).
- AC-3 not met yet: `the_plan_offers_to_keep_the_resurrected_markers_instead_of_stopping` 는 일부러 빨강이라 이유와 함께 `#[ignore]` 예요. 이름으로 돌린 결과(`-- --ignored`): `no go-residue` 로 실패해요. `go-residue` 단계가 계획에 생기면 이 표시를 지우고 이 줄을 met 로 바꿔요.
- AC-4 met for what exists: 계획 앞뒤 스냅샷이 같고(`the_plan_writes_nothing`), `moai.json` 없이 Go 표식만 있으면 지금처럼 실패해요(`go_markers_without_a_valid_marker_keep_failing_the_way_they_do`). `go-residue` 단계가 없다는 부분은 후반이 오면 같은 파일이 다시 확인해요.
- AC-5 met: 변형 3개(5개 이하), 모두 죽었고 복구 뒤 `git diff` 가 0바이트예요.

  | 변형 | 바꾼 곳 | 결과 |
  | --- | --- | --- |
  | M1 | `rust_owned`: 옛 이름 표식도 주인으로 봄 (`marker_file == Some("moai.json")` 조건 제거) | 단위 시험이 실패(25 passed, 1 failed) |
  | M2 | `dir`: 새 규칙을 쓰지 않고 `generation == "rust"` 로 되돌림 | 같은 단위 시험이 실패 |
  | M3 | `rust_owned`: `mixed` 를 주인으로 보지 않음 | 같은 단위 시험이 실패 |

  세 변형 모두 같은 단위 시험 하나가 잡았어요. CLI 시험은 단위 시험이 먼저 실패해 그 변형들에서는 돌지 않았어요.
- 돌리지 않은 것: 이 변경만의 독립 감사, 후반의 시험 3개(AC-3, AC-4 의 단계 부분), 전체 시험.
