---
id: SPEC-20261117-build-folder-windows-home
title: 빌드 폴더의 HOME을 Windows에서도 찾기, verify 문구를 작업 폴더 기준으로
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-slot-build-folder
---

## Intent

PR #3162(slot 빌드 폴더) 감사의 후속이에요.

- N1: Windows는 지원 대상이에요(C-89). `moai lane-env`와 `moai heavy-slot-acquire`는 HOME만 읽어서, HOME이 없는 Windows에서
  빌드 폴더를 정하지 못했어요. MoAI의 다른 곳(`key_vault::home`, 세션·launcher)이 쓰는 규칙 — HOME, 없으면 Windows에서
  USERPROFILE — 을 순수 함수 `key_vault::home_from`으로 꺼내 함께 쓰고, `tools/build_folder.py`도 같은 규칙을 따라요.
- N2: `workflows/verify.md`에 "one build folder per lane … never share it with another lane"이 남아 있었어요 → 작업 폴더 기준으로.
- N3: 이 작업 폴더의 기록(`.moai-build.json`)이 있는 다른 이름의 빌드 폴더도 "another worktree's build folder"로 경고했어요 → 기록이
  이 작업 폴더를 가리키면 경고하지 않아요.

## Non-goals

- `home_workdir::home`(Windows에서 HOME을 무시하는 규칙, `~/.moai/worktrees`)와 합치기. C-79 host_home 해석기는 아직 병합되지 않았어요.
- Windows에서 lane-env 전체 CLI 시험(시험 파일은 unix 전용); Windows 쪽은 순수 함수 시험과 `cfg(windows)` 시험 하나.

## Requirements

- REQ-1: When HOME is unset or empty on Windows, `lane-env` and `heavy-slot-acquire` shall take the build folder's home from USERPROFILE by the shared `key_vault::home_from` rule. · failure: none — REQ-2 covers the missing case.
- REQ-2 (REQ-1): IF neither HOME nor USERPROFILE gives an absolute folder, THEN `lane-env` shall refuse with stdout empty and exit 1.
- REQ-3: `tools/build_folder.py` shall read the home by the same rule. · failure: none — it refuses as before when none is set.
- REQ-4: `heavy-slot-acquire` shall not warn about a sibling build folder whose owner record names this worktree. · failure: none — a record of another worktree still warns.
- REQ-5: The VERIFY workflow shall state one build folder per worktree. · failure: none — text, pinned by AC-4.

## Acceptance

- AC-1 (REQ-1, REQ-2): HOME 없음·빈 값이면 Windows에서 USERPROFILE, HOME이 있으면 HOME, Windows가 아니면 USERPROFILE을 읽지 않음, 둘 다 없으면 두 이름을 말하며 거절, 상대 경로는 거절.
  verify: `cargo test --locked --offline --lib lane_env`
- AC-2 (REQ-3): Python 도구의 `home()`이 같은 다섯 경우에서 같게 판정해요.
  verify: `python3 -m unittest tests.build_folder_test`
- AC-3 (REQ-4): 이 작업 폴더를 가리키는 기록이 있는 형제 폴더는 경고 없음, 기록이 없으면 경고.
  verify: `cargo test --locked --offline --lib heavy_slot::build_folder_tests`
- AC-4 (REQ-5): verify.md가 "Use one build folder per worktree"와 "never share it with another worktree"를 말하고 "per lane"은 없어요.
  verify: `python3 -m unittest tests.check_pr_self_audit_test`; `python3 -I tools/source_manifest.py check`
- AC-5 (REQ-1…REQ-5): 이웃 검사가 통과해요.
  verify: `cargo test --locked --offline --test lane_env_cli`; `cargo test --locked --offline --test heavy_slot_cli`;
  `cargo test --locked --offline --test agent_defs_package`; `cargo test --locked --offline --test product_cli`;
  `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --bin moai`; `cargo fmt --check`;
  `cargo clippy --locked --offline --all-targets -- -D warnings`; `python3 -m unittest tests.repo_scans_test tests.ci_scope_test tests.source_manifest_test`

## Scope

- write: `src/key_vault.rs`, `src/lane_env.rs`, `src/heavy_slot.rs`, `src/main.rs`, `tools/build_folder.py`, `tests/build_folder_test.py`,
  `tests/check_pr_self_audit_test.py`, `skills/moai/workflows/verify.md`, `skills/moai/references/source-manifest.json`,
  `.moai/specs/SPEC-20261117-build-folder-windows-home/spec.md`
- read-only: `src/home_workdir.rs`
- forbidden: 실제 HOME, 네트워크 쓰기(push, PR, Issue)

## Plan

1. `key_vault::home`의 규칙을 `home_from`으로 꺼내고 `lane_env::home_dir`·`resolve`·`heavy_slot::Here::of`가 USERPROFILE도 받게.
2. Python `home()`을 같은 규칙으로; 시험.
3. N3 한 줄, N2 문장과 고정 시험; manifest write·check.
4. 시험·형식·clippy·Python, mutation 3개.

## Risks

- MoAI에 HOME 규칙이 두 가지 있어요(`key_vault`: HOME 먼저, `home_workdir`: Windows에서 USERPROFILE만). 빌드 폴더는 리더 지시대로
  `key_vault` 쪽을 따라요 → C-79 host_home 해석기가 병합되면 그쪽으로 맞춰요.
- tip `17c19a2759`의 manifest에 키가 두 번 있어요 → 이 가지의 `write`(tip 도구)가 하나로 합쳐요(#3175, C-92와 같은 결과).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | HOME 규칙은 `key_vault::home_from` 하나를 함께 씀(새 규칙 없음) | 리더 지시("다른 moai 코드의 해석을 재사용") | `home_workdir` 규칙으로 |
| D2 | Windows에서 HOME이 있으면 HOME, 없을 때만 USERPROFILE | `key_vault::home`과 같음; Git Bash가 HOME을 둠 | USERPROFILE 먼저 |
| D3 | N3은 기록이 이 작업 폴더를 가리킬 때만 경고를 뺌 | 작은 변경, 기록이 증거 | 이름만으로 판정 |
| D4 | `key_vault::home`은 이제 Windows에서 빈 USERPROFILE을 거절해요("HOME is not set"); 전에는 빈 경로를 돌려줬어요(감사 N-b) | 빈 값은 설정 안 됨과 같다는 HOME의 규칙과 맞춤 | 빈 USERPROFILE을 그대로 돌려주기 |

## Result

구현 에이전트가 2026-11-18에 채웠어요(UTC 2026-11-17 16:57–17:06Z). 감사가 다시 확인해요. 상태는 `draft` 그대로예요.
기준 tip `17c19a2759`(#3162 병합 포함). 빌드 폴더는 이 worktree의 `target/`, 무거운 작업 슬롯 안에서 cargo 한 번에 하나.
RED는 mutation으로 보였어요(새 시험은 바뀐 함수 모양을 써서 바뀌기 전 코드에서는 컴파일되지 않아요).

- AC-1: met — `--lib lane_env`의 `userprofile_stands_in_for_an_unset_home_on_windows_only`. mutation M1(공유 규칙이 USERPROFILE을
  안 읽음) KILLED. `cfg(windows)` 시험 `a_windows_userprofile_is_the_home`은 이 Mac에서 돌지 않아요(Windows CI 몫)
- AC-2: met — `tests.build_folder_test`의 `test_home_takes_userprofile_on_windows_only_when_home_is_unset`. mutation M2 KILLED
- AC-3: met — `heavy_slot::build_folder_tests::a_sibling_folder_recorded_for_this_worktree_is_its_own`. mutation M3 KILLED
- AC-4: met — `tests.check_pr_self_audit_test`(per worktree 고정, per lane 없음), manifest check 0(`write`가 tip의 중복 키도 하나로)
- AC-5: met — `lane_env_cli` 7, `heavy_slot_cli` 21, `agent_defs_package` 18, `product_cli` 75, `home_isolation_scan` 16, `spec_cli` 17,
  `--lib lane_env heavy_slot key_vault` 42, `--bin moai` 27; fmt 0, clippy 0; Python 5개 모듈 138개 OK; `moai spec check --all` 0

실행 안 함: Windows에서의 실제 실행(Windows CI 몫), 전체 `cargo test`.
