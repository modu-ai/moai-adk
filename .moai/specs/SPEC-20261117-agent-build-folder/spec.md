---
id: SPEC-20261117-agent-build-folder
title: 에이전트·감사 빌드 폴더를 작업 폴더 밖 한 곳으로 (C-82)
status: draft
revision: 1
created: 2026-11-17
parent: null
---

## Intent

에이전트와 감사 빌드는 지금 작업 폴더 안 `target/`에 쌓여요. 2026-11-17 조사에서 작업 폴더 `target/`
183개가 285.8 GiB였고, 테스트 빌드 하나(4.2 GiB) 중 2.9 GiB가 incremental 출력이었어요. 작업 폴더 안
빌드 출력은 파일 감시(fseventsd, CPU 약 200%, 메모리 4~7 GB)도 바쁘게 만들어요. 사용자 결정(2026-11-17,
선택 "빌드 폴더를 한 곳으로 모으기"): 빌드 폴더를 작업 폴더 밖 한 곳 `~/.moai/build/<project>-<작업 폴더
이름>/`(작업 폴더마다 하나, Spotlight 제외 위치)으로 모으고, `CARGO_INCREMENTAL=0`을 늘 켜고, 작업이 끝나면
그 폴더의 cargo 출력을 지워요. 남은 폴더는 읽기 전용 미리보기로 보여 주고 사용자가 고른 것만 지워요.

## Non-goals

- `moai lane-env`의 `/tmp/moai-rc-target-<lane>` 폴더와 `moai heavy-slot-acquire`의 경고 문장 바꾸기
  (`src/lane_env.rs`, `src/heavy_slot.rs`; 후속 카드).
- 열린 PR이 닫힘(closed)인지 판정: 미리보기는 로컬만 읽어서 작업 폴더가 없어졌거나 가지가 병합된 경우만
  후보로 봐요.
- 작업 폴더 안 기존 `target/` 정리(`worktree-gc` 몫), `moai` 명령으로 된 삭제 단계.
- `tools/candidate_summary.py`, `ci.yml`, `install/**`(다른 가지가 고쳐요).

## Requirements

- REQ-1: When an agent or audit build starts in the moai-adk repository, `tools/build_folder.py env` shall print `CARGO_TARGET_DIR` for `<HOME>/.moai/build/<project>-<worktree name>/` and `CARGO_INCREMENTAL=0`, and write the folder's owner record.
- REQ-2 (REQ-1): IF HOME is not absolute or not writable, or a link sits on the folder's path or on the record, THEN the tool shall print nothing on stdout, write nothing through the link and exit 1.
- REQ-3: The tool shall give each worktree its own folder name and never the same folder to two worktrees of one repository.
- REQ-4: When the task ends, `tools/build_folder.py clean` shall delete the cargo output of this worktree's folder and keep every other file.
- REQ-5 (REQ-4): IF the folder's owner record names another worktree, THEN the tool shall delete nothing and exit 1.
- REQ-6: When `moai optimize-plan` runs with a HOME, the preview shall list each folder under `<HOME>/.moai/build/` with its size and mark it `trim` only when its owner record names this repository and a worktree that is gone or whose branch is merged.
- REQ-7 (REQ-6): IF `<HOME>/.moai/build` or a folder under it is a link, THEN the preview shall list no row from behind the link and name it in `warnings` or as a `skip` row.
- REQ-8: When the user picks rows, `tools/build_folder.py remove` shall delete only folders that are `trim` rows of the given preview.
- REQ-9 (REQ-8): IF any picked name is not a `trim` row, is not one folder name, or is now a link, THEN the tool shall delete nothing and exit 1.
- REQ-10: The instructions agents and auditors read shall name the folder, `CARGO_INCREMENTAL=0`, the refusal path, the evidence rule and the end-of-task cleanup through `internal/build-folder`.
- REQ-11: The check scripts `tools/local_check.sh` and `tools/self_check.sh` shall set `CARGO_INCREMENTAL=0` for every build.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): 폴더 이름·내보내기 줄·소유 기록, 링크·쓸 수 없는 HOME·상대 HOME 거절이 시험으로 고정돼요.
  verify: `python3 -m unittest tests.build_folder_test`
- AC-2 (REQ-4, REQ-5): `clean`은 cargo 출력만 지우고 증거 파일을 남기며, 다른 작업 폴더의 기록이면 거절해요.
  verify: `python3 -m unittest tests.build_folder_test.CleanTest`
- AC-3 (REQ-6, REQ-7): 미리보기가 증명된 남은 폴더만 `trim`으로 보이고, 링크는 따라가지 않아요.
  verify: `cargo test --locked --offline --test optimize_plan_cli`
- AC-4 (REQ-8, REQ-9): `remove`는 미리보기의 `trim` 행만 지우고, 하나라도 맞지 않으면 아무것도 지우지 않아요.
  verify: `python3 -m unittest tests.build_folder_test.RemoveTest`
- AC-5 (REQ-10): 두 에이전트가 `internal/build-folder`를 읽고, 규칙 문장이 고정돼요.
  verify: `cargo test --locked --offline --test agent_defs_package`
- AC-6 (REQ-11): 두 검사 스크립트가 `export CARGO_INCREMENTAL=0`을 가져요.
  verify: `python3 -m unittest tests.build_folder_test.ScriptsTest`
- AC-7 (REQ-1…REQ-11): 이웃 검사가 모두 통과해요.
  verify: `cargo test --locked --offline --bin moai`; `cargo test --locked --offline --test home_isolation_scan`;
  `cargo test --locked --offline --test lane_rules_2879`; `cargo test --locked --offline --test mode_workflow_steps`;
  `cargo test --locked --offline --test mode_cards_53`; `cargo fmt --check`;
  `cargo clippy --locked --offline --all-targets -- -D warnings`;
  `python3 -m unittest tests.skill_references_test tests.self_check_test tests.local_check_test`

## Scope

- write: `tools/build_folder.py`, `tools/local_check.sh`, `tools/self_check.sh`, `tests/build_folder_test.py`,
  `tests/optimize_plan_cli.rs`, `tests/agent_defs_package.rs`, `tests/lane_rules_2879.rs`, `tests/mode_workflow_steps.rs`,
  `tests/mode_cards_53.rs`, `agents/impl.md`, `agents/audit.md`, `skills/moai/references/internal/build-folder.md`,
  `skills/moai/references/internal/git-safety.md`, `skills/moai/references/internal/self-check.md`,
  `skills/moai/references/source-manifest.json`, `skills/moai/workflows/mode-agent.md`, `skills/moai/workflows/verify.md`,
  `skills/moai/workflows/optimize.md`, `src/optimize.rs`, `.moai/specs/SPEC-20261117-agent-build-folder/spec.md`
- read-only: `src/worktree_gc.rs`, `src/lane_env.rs`, `src/heavy_slot.rs`, `src/ops/optimize_plan.rs`
- forbidden: `tools/candidate_summary.py`, `.github/workflows/ci.yml`, `install/**`, 실제 HOME, 네트워크 쓰기(push, PR, Issue)

## Plan

1. 미리보기: `optimize_plan_cli` 시험 먼저(RED) → `src/optimize.rs`에 `~/.moai/build/*` 행(GREEN).
2. 도구: `tests/build_folder_test.py` 먼저(RED) → `tools/build_folder.py`(GREEN), 두 검사 스크립트에 `CARGO_INCREMENTAL=0`.
3. 지시문: `internal/build-folder.md` 새로 쓰고, 두 에이전트의 `reads:`·경계·정리 규칙, git-safety, self-check, mode-agent(크기 그대로), verify, optimize가 그것을 가리키게; 고정 시험을 옮겨요.
4. 이웃 시험·형식·clippy·Python 검사, mutation 5개 이하.

## Risks

- `moai heavy-slot-acquire`가 `CARGO_TARGET_DIR`이 lane 폴더와 다르다고 경고해요 → 참고 문서 9번이 "예상된 경고"라고 말하고, 바꾸기는 후속 카드예요.
- 소유 기록이 없는 폴더는 지울 후보가 되지 않아요 → 사용자 프로젝트에서는 기록을 손으로 써야 후보가 돼요(참고 문서 3·4번).
- 미리보기는 PR이 닫혔는지 모르므로 병합되지 않고 닫힌 가지의 폴더는 남아요 → 작업 폴더를 지우면 `worktree_gone`으로 나와요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | 쓰기 범위를 카드 범위(`tools/**`, `internal/**`, `agents/**`, `tests/**`) 밖 `src/optimize.rs`로 넓힘 | 리더 결정(OPTIMIZE 코드 확장 허용) | 미리보기 행을 빼고 `src/optimize.rs` 되돌리기 |
| D2 | `mode-agent.md`·`verify.md`·`optimize.md`(workflows)도 고침, `mode-agent.md`는 줄이기만 | 리더 지시(heavy-checks 문장, 크기 그대로) | 세 파일 되돌리기 |
| D3 | 지울 후보는 소유 기록(`.moai-build.json`)이 증명할 때만; 이름 짐작은 하지 않음 | 구현 판단(안전: 증명 못 하면 남김) | 이름 접두사로 판정 |
| D4 | 작업 폴더 이름: `<project>-<worktree 폴더 이름>`, 주 checkout과 `<project>-`로 시작하는 이름은 그대로 | 구현 판단(`~/.moai/worktrees/<project>-<id>`와 겹침 방지) | 늘 접두사 붙이기 |
| D5 | 삭제 단계는 저장소 도구 `tools/build_folder.py remove`(미리보기 JSON의 `trim` 행만), `moai` 명령은 후속 | 범위(`src`는 optimize만) | `moai` apply 명령 추가 |
| D6 | 검사 스크립트는 `CARGO_INCREMENTAL=0`만 정하고 `CARGO_TARGET_DIR`은 호출자 것을 씀 | 구현 판단(CI 동작 그대로) | 스크립트가 폴더도 정하기 |
| D7 | 새 참고 문서 `internal/build-folder`를 `moai-impl`·`moai-audit`가 읽음 | 리더 지시("internal reference를 가리키기") | `reads:`에서 빼기 |

## Result

(채우기 전 — 구현 에이전트가 검사 뒤에 채워요.)
