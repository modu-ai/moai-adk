---
id: SPEC-20261117-slot-build-folder
title: lane-env와 heavy-slot이 C-82 빌드 폴더를 쓰게
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-agent-build-folder
---

## Intent

C-82가 에이전트·감사 빌드 폴더를 `~/.moai/build/<project>-<작업 폴더 이름>/`(작업 폴더마다 하나)로 옮겼지만,
`moai lane-env`는 아직 lane마다 하나인 `/tmp/moai-rc-target-<lane>`을 내보내고, `moai heavy-slot-acquire`는
그 폴더가 아니면(설정 안 함 포함) 늘 경고했어요. 사용자 요청(2026-11-17 08:22Z, "빌드 슬롯도 이번 버전에 맞게
개선"): `lane-env`가 C-82 폴더를 내보내고, heavy-slot은 그 폴더를 다른 작업 폴더와 함께 쓰거나 다른 작업 폴더
안을 가리킬 때만 경고해요.

## Non-goals

- 이미 있는 `/tmp/moai-rc-target-*` 폴더 지우기(`moai optimize-plan`의 빌드 루트 행이 다뤄요).
- `CARGO_HOME` 고르는 규칙(`--base` 아래 lane 것 → 공용 → 설정 안 함) 바꾸기.
- `tools/build_folder.py` 바꾸기, `AGENTS.md`·`AGENTS.local.md`, `install/**`, `ci.yml`.
- `mode-agent.md` 문장 늘리기.

## Requirements

- REQ-1: When `moai lane-env --lane <lane>` runs inside a git worktree, the command shall print `CARGO_TARGET_DIR` for `<HOME>/.moai/build/<name>` with the name `tools/build_folder.py env` gives that worktree, and `CARGO_INCREMENTAL=0`.
- REQ-2 (REQ-1): When `.moai`, `build`, the folder or its owner record is missing, the command shall make the folders with mode 0700 and write the owner record `tools/build_folder.py env` writes.
- REQ-3 (REQ-1, REQ-2): IF HOME is not absolute, the working folder is outside a git worktree, a link or a non-folder sits on the folder's path, or the owner record names another worktree, THEN the command shall print nothing on stdout, write nothing through a link and exit 1.
- REQ-4: The command shall give two worktrees two folders and one worktree the same folder for every lane. · failure: none — a pure naming rule; its refusals are REQ-3.
- REQ-5: When `moai heavy-slot-acquire` wraps `cargo` with a `CARGO_TARGET_DIR` that is shared with another worktree or lies inside another worktree, the command shall print one warning line on stderr that names this worktree's build folder.
- REQ-6 (REQ-5): IF `CARGO_TARGET_DIR` is unset, is this worktree's build folder or lies inside this worktree, or the command is not `cargo`, THEN the command shall print no build-folder warning.
- REQ-7: The documents that describe the two commands shall state the new folder and the new warning rule. · failure: none — text in `internal/build-folder` item 9, `workflows/verify.md` and the interface appendix, pinned by AC-5.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-4): 작업 폴더마다 다른 폴더, lane과 무관, 0700과 소유 기록, 한 줄 `eval` 결과가 시험으로 고정돼요.
  verify: `cargo test --locked --offline --test lane_env_cli`
- AC-2 (REQ-1, REQ-2): 같은 저장소·작업 폴더·HOME에서 `moai lane-env`와 `python3 -I tools/build_folder.py env`가 같은 폴더와 같은 기록을 만들고 서로의 기록을 받아들여요(주 checkout, `<project>-` 이름, 공백·한글 이름, 주 checkout 안 작업 폴더).
  verify: `cargo test --locked --offline --test lane_env_cli the_folder_and_record_match_tools_build_folder`
- AC-3 (REQ-3): 링크된 `.moai`·`build`·폴더, 파일인 폴더, 다른 작업 폴더의 기록, git 밖, 상대 HOME, 잘못된 lane·base가 stdout 없이 exit 1이고 링크 너머에 아무것도 쓰지 않아요.
  verify: `cargo test --locked --offline --test lane_env_cli refusals`; `cargo test --locked --offline --lib lane_env`
- AC-4 (REQ-5, REQ-6): 공유 폴더란 옛 lane 폴더 `moai-rc-target-*`, HOME의 `.moai/build` 아래 다른 작업 폴더의 폴더, 소유 기록이 다른 작업 폴더를 가리키는 폴더예요. 경고 여섯 경우와 조용한 여섯 경우(설정 안 함, 자기 폴더, 자기 `target/`, 상대 경로, 주 checkout 안 작업 폴더)와 cargo가 아닌 명령이 시험으로 고정돼요.
  verify: `cargo test --locked --offline --test heavy_slot_cli a_cargo_run_is_warned_about`; `cargo test --locked --offline --lib heavy_slot::build_folder_tests`
- AC-5 (REQ-7): 문서 문장이 고정돼요.
  verify: `cargo test --locked --offline --test lane_rules_2879`; `cargo test --locked --offline --test agent_defs_package`; `python3 -I tools/source_manifest.py check`
- AC-6 (REQ-1…REQ-7): 이웃 검사가 통과해요.
  verify: `cargo test --locked --offline --test heavy_slot_cli`; `cargo test --locked --offline --test mode_cards_53`;
  `cargo test --locked --offline --test mode_workflow_steps`; `cargo test --locked --offline --test home_isolation_scan`;
  `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --test git_env_scrub`;
  `cargo test --locked --offline --bin moai`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`;
  `python3 -m unittest tests.heavy_check_slot_test tests.build_folder_test tests.source_manifest_test tests.repo_scans_test tests.ci_scope_test`

## Scope

- write: `src/lane_env.rs`, `src/heavy_slot.rs`, `src/main.rs`, `tests/lane_env_cli.rs`, `tests/heavy_slot_cli.rs`,
  `tests/lane_rules_2879.rs`, `skills/moai/references/internal/build-folder.md`, `skills/moai/workflows/verify.md`,
  `skills/moai/references/source-manifest.json`, `docs/design-review/appendix/interfaces.md`,
  `.moai/specs/SPEC-20261117-slot-build-folder/spec.md`
- read-only: `tools/build_folder.py`, `src/optimize.rs`, `src/identity.rs`, `skills/moai/workflows/mode-agent.md`
- forbidden: `src/gate*`, `src/card*`, `src/headver.rs`, `src/update.rs`, `src/agents.rs`, `AGENTS.md`, `install/**`,
  `.github/workflows/ci.yml`, `skills/moai/references/internal/{sdd,code-atlas}.md`,
  `skills/moai/workflows/{read,plan,run,sync,audit}.md`, 실제 HOME, 네트워크 쓰기(push, PR, Issue)

## Plan

1. 시험 먼저(RED): `lane_env_cli` 새 시험 7개와 `heavy_slot_cli` 새 시험 1개를 바뀌기 전 코드로 돌려 의도한 단언에서 실패하는지 봐요.
2. `src/lane_env.rs`: Python과 같은 이름 규칙(`folder_name`), git 두 번(`rev-parse`), 폴더·기록 만들기(`prepare`), 단위 시험.
3. `src/heavy_slot.rs`: `Here`(작업 폴더·모든 worktree·자기 빌드 폴더)와 새 경고 규칙, 단위 시험; `src/main.rs` 연결.
4. 문서: build-folder 9번, verify 두 곳, 인터페이스 부록; 고정 시험 옮기기; manifest write·check.
5. 이웃 시험·형식·clippy·Python, mutation 3개 이하.

## Risks

- lane이 지금 쓰는 `/tmp/moai-rc-target-<lane>`은 이제 경고를 받아요 → 경고 줄이 `moai lane-env` 한 줄을 알려 줘요.
- 작업 폴더마다 의존성을 다시 빌드해서 lane 하나가 쓰는 디스크가 늘 수 있어요 → C-82 끝정리(`tools/build_folder.py clean`)와 `moai optimize-plan`이 다뤄요.
- Windows에서는 HOME이 보통 없어서 `lane-env`가 거절해요(Python 도구와 같음) → Windows 빌드 규칙은 이 SPEC 밖이에요.
- 이름 규칙이 Rust와 Python 두 곳에 있어요 → AC-2 시험이 같은 입력으로 둘을 대조해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | `lane-env`가 폴더(0700)와 소유 기록도 만듦(전에는 읽기만) | `internal/build-folder` 3번: 기록이 없으면 `optimize-plan`이 지울 후보로 보지 않음 | `prepare` 대신 경로만 계산 |
| D2 | 이름 규칙을 Rust(`lane_env::folder_name`)에 옮기고 Python과 대조 시험 | 리더 지시("한 곳의 규칙, 아니면 두 구현이 같다는 시험") | `lane-env`가 Python 도구를 부르기 |
| D3 | `CARGO_TARGET_DIR` 설정 안 함은 경고 없음 | 리더 지시(공유이거나 다른 작업 폴더 안일 때만) | 설정 안 함도 경고 |
| D4 | 옛 lane 폴더 `moai-rc-target-*`는 "공유"로 경고 | 그 폴더는 lane의 모든 작업 폴더가 함께 씀 | 이름 검사 빼기 |
| D5 | `mode-agent.md`는 그대로(47,962바이트) | heavy-checks 줄이 이미 "the worktree's build folder (`internal/build-folder`)"이고 `lane_rules_2879`가 `lane-env` 낱말을 막음 | — |
| D6 | `verify.md` 두 곳과 인터페이스 부록도 고침, `lane_rules_2879` 고정 문장을 옮김 | 바뀐 동작과 다른 문장이 남지 않게 | 세 문서 되돌리기 |
| D7 | heavy-slot은 `cargo`를 감쌀 때만 git을 불러요 | 다른 명령의 시작을 늦추지 않게 | 늘 부르기 |

## Result

(구현 뒤 채워요.)
