---
id: SPEC-20261117-supported-platforms
title: 지원 플랫폼 세 개 — CI target과 설치 스크립트, 지원 문서 (조각 1)
status: draft
revision: 1
created: 2026-11-17
parent: null
approval: {by: null, at: null, basis: null, digest: null}
---

## Intent

사용자 결정(2026-11-17, 선택 원문): "windows, macOS 실리콘, linux_x86_64 이렇게 제공하자. claude code, codex에서
지원하는 환경으로 동일하게 지원하자." 지원 플랫폼은 세 개예요.

- macOS arm64: `aarch64-apple-darwin`
- Linux x86_64: `x86_64-unknown-linux-gnu`
- Windows x86_64: `x86_64-pc-windows-msvc`

이 변경 전에는 CI가 병합 뒤에 Linux x86_64·arm64, macOS arm64·x86_64 네 target과 `windows-build`를 돌렸고,
후보 검증(`candidate`)은 다섯 target을 돌렸어요. 2026-11-09…17 조사에서 `macos-15-intel`과 `ubuntu-24.04-arm`이
CI job 시간의 46%였고, 다중 target run 206개 중 175개에서 마지막으로 끝났어요. 설치 스크립트는 Darwin·Linux의
x86_64·arm64를 모두 받아들였어요.

이 SPEC은 조각 1이에요: CI target, 설치 스크립트, 지원 문서를 세 플랫폼에 맞춰요.

## Non-goals

- 릴리스(dist) target 목록(`dist-workspace.toml`, 지금 `aarch64-apple-darwin`만)을 바꾸는 일. 릴리스 target은
  별도 사용자 선택이에요(아래 Decisions D5).
- Windows에서 시험을 실행하는 일과 `cfg(unix)` 시험 파일 정리(Windows 시험 계획, 후속 조각).
- 병합 큐(merge_group), `audit-gate` job, 위험 경로의 PR macOS 실행(CI 재설계의 다른 단계).
- `skills/moai/workflows/verify.md`의 CI target 문장, `src/flow.rs`·`src/lane_next.rs`의 주석·시험 자료에 남은 옛
  target 이름(이 조각의 쓰기 범위 밖, 아래 Result의 후속 목록).

## Requirements

요구 문장은 영어 GEARS로 써요(`src/spec.rs` `parse_gears`가 읽는 형식). `IF …, THEN …`은 원치 않는 동작이에요.

- REQ-1: When a pull request targets the integration branch, the `ci` workflow shall run the check set on `linux-x86_64` only.
- REQ-2: When a push merges into the integration branch, the `ci` workflow shall run the check set on `linux-x86_64` and `macos-arm64` and run the `windows-build` job.
- REQ-3: The workflows under `.github/workflows/` shall name no Intel macOS or Linux arm64 runner, target or nextest asset.
- REQ-4: The candidate verdict shall require exactly the targets `linux-x86_64`, `macos-arm64` and `windows-x86_64`.
- REQ-5: The aggregate check of the `ci` workflow shall keep the name `ci`.
- REQ-6: Every action a workflow uses shall stay pinned to a full commit SHA.
- REQ-7: When an installer runs on macOS arm64 (an x86_64 shell under Rosetta 2 included), Linux x86_64 or Windows x86_64, the installer shall select that platform's target triple.
- REQ-8: IF an installer runs on any other platform, THEN the installer shall exit 1 before any download or write.
- REQ-9: IF an installer stops for the platform, THEN its message shall name the platform and the three supported platforms.
- REQ-10: Each support document shall name the three supported platforms and state that Intel macOS and Linux arm64 are not supported.
- REQ-11: The change shall carry its own SPEC at `.moai/specs/SPEC-20261117-supported-platforms/spec.md`, tracked by git in the same commit.

실패 경로: REQ-7 → REQ-8, REQ-9. REQ-1·REQ-2 → 없음(한 target 실패는 기존 규칙대로 `ci`를 실패시켜요).
REQ-10의 "support document"는 `docs/support-matrix.md`, `docs/host-support-matrix.md`, `README.md`예요.

## Acceptance

검사는 이 worktree에서 Python 3.13(PyYAML)과 `pwsh`가 PATH에 있을 때 돌려요.

- AC-1 (REQ-1, REQ-2, REQ-5): PR 목록은 `linux-x86_64` 하나, 병합 뒤 목록은 `linux-x86_64`(ubuntu-24.04, 30분)와
  `macos-arm64`(macos-15, 45분), `windows-build`는 PR이 아닐 때만, 집계 job 이름은 `ci`예요.
  verify: `python3 -m unittest tests.github_assets_test.WorkflowTest.test_rc_ci_runs_the_supported_targets_and_fails_when_any_fails tests.github_assets_test.WorkflowTest.test_rc_ci_moves_the_full_matrix_after_merge_without_cancelling_it tests.github_assets_test.WorkflowTest.test_rc_ci_reports_the_rc_ci_check_on_prs_to_the_integration_branch tests.github_assets_test.WorkflowTest.test_rc_ci_post_merge_macos_waits_are_replaced_by_a_newer_merge`
- AC-2 (REQ-3): verify: `python3 -m unittest tests.github_assets_test.WorkflowTest.test_no_workflow_names_a_dropped_platform`;
  `grep -nE 'macos-15-intel|ubuntu-24.04-arm|macos-x86_64|linux-arm64|x86_64-apple-darwin|aarch64-unknown-linux-gnu' .github/workflows/*.yml`
  → 출력 없음, exit 1
- AC-3 (REQ-4): verify: `python3 -m unittest tests.candidate_summary_test tests.github_assets_test.WorkflowTest.test_candidate_runs_the_three_targets_and_one_verdict_needs_all_of_them tests.github_assets_test.WorkflowTest.test_candidate_summary_script_judges_the_same_three_targets`
- AC-4 (REQ-6): verify: `python3 -m unittest tests.github_assets_test.WorkflowTest.test_every_action_is_pinned_to_a_commit_sha`;
  `actionlint .github/workflows/*.yml` → exit 0
- AC-5 (REQ-7): 가짜 `uname`(과 Rosetta용 가짜 `sysctl`)으로 Darwin arm64·aarch64, Darwin x86_64(번역 중),
  Linux x86_64·amd64가 각 triple로 잡혀요. `install.ps1`은 같은 시험을 `pwsh`로 돌리고, Windows 분기는 스크립트의
  구문 트리에서 꺼낸 `Get-WindowsTarget`으로 AMD64·x64·x86(64비트 OS)을 확인해요.
  verify: `python3 -m unittest tests.install_script_test.InstallShTest.test_the_three_supported_platforms_are_detected tests.install_script_test.InstallPs1Test.test_the_three_supported_platforms_are_detected tests.install_script_test.InstallPs1Test.test_windows_takes_only_x86_64`
- AC-6 (REQ-8, REQ-9): Darwin x86_64(번역 아님, sysctl 응답 없음), Linux aarch64·arm64·i686, FreeBSD amd64는 exit
  1, HOME 트리 변화 없음, `Unsupported platform: <os> <arch>`와 지원 목록이 나와요. Windows ARM64·IA64·x86(32비트
  OS)는 `Get-WindowsTarget`이 `$null`이에요.
  verify: `python3 -m unittest tests.install_script_test.InstallShTest.test_an_unsupported_platform_stops_with_the_supported_list_and_writes_nothing tests.install_script_test.InstallPs1Test.test_an_unsupported_platform_stops_with_the_supported_list_and_writes_nothing tests.install_script_test.InstallPs1Test.test_windows_takes_only_x86_64`
- AC-7 (REQ-10): verify: `python3 -I tools/check_support_matrix.py docs/support-matrix.md docs/e2e-receipts/` → exit 0;
  `grep -c 'x86_64-pc-windows-msvc' docs/support-matrix.md docs/host-support-matrix.md README.md` → 각 1 이상;
  `python3 -m unittest tests/support_matrix_test.py tests/host_matrix_test.py tests/cargo_metadata_test.py tests/oss_docs_test.py`
- AC-8 (REQ-11): verify: `git ls-files .moai/specs/SPEC-20261117-supported-platforms` → 이 파일
- AC-9 (REQ-1…REQ-11): `IF`로 시작하지 않는 요구는 모두 `parse_gears`를 통과해요. verify: 커밋하지 않는 임시 시험
  파일로 이 파일의 REQ 줄을 `moai_cli::spec::parse_gears`에 넣어 봐요(결과는 Result).
- AC-10 (전체): verify: `python3 -m unittest tests/install_script_test.py tests/github_assets_test.py tests/candidate_summary_test.py tests/ci_scope_test.py tests/repo_scans_test.py tests/ci_install_ripgrep_test.py tests/local_check_test.py tests/self_check_test.py tests/host_live_harness_test.py tests/goal_block_test.py`;
  `bash -n install/install.sh`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`;
  `cargo test --locked --offline --bin moai`; `cargo test --locked --offline --test launcher_three_hosts`;
  `cargo test --locked --offline --test gitignore_block_cli`

## Scope

- write: `.github/workflows/ci.yml`, `.github/workflows/candidate.yml`, `.github/workflows/candidate-target.yml`,
  `install/install.sh`, `install/install.ps1`, `install/README.md`, `tools/candidate_summary.py`,
  `tests/github_assets_test.py`, `tests/candidate_summary_test.py`, `tests/install_script_test.py`,
  `docs/support-matrix.md`, `docs/host-support-matrix.md`, `README.md`,
  `.moai/specs/SPEC-20261117-supported-platforms/spec.md`
- read-only: `dist-workspace.toml`, `tools/ci_scope.py`, `tools/check_support_matrix.py`, `src/spec.rs`,
  `.moai/specs/SPEC-20261117-sdd-foundation/spec.md`
- forbidden: 릴리스 target 변경, 실제 HOME, 네트워크 쓰기(push, PR, Issue), 저장소 설정·ruleset

## Plan

1. RED: `tests/github_assets_test.py`(병합 뒤 target 둘, 후보 target 셋, 빠진 플랫폼 이름 금지),
   `tests/candidate_summary_test.py`(target 셋), `tests/install_script_test.py`(가짜 `uname`·`sysctl`로 지원·미지원
   플랫폼, `Get-WindowsTarget`). 새 단언에서 실패하는지 확인.
2. GREEN: workflow 세 개와 `tools/candidate_summary.py`; `install.sh` `detect_target`, `install.ps1`
   `Get-WindowsTarget`·`Get-UnixTarget`·`Get-Target`.
3. 문서: `docs/support-matrix.md`, `docs/host-support-matrix.md`, `README.md`, `install/README.md`.
4. AC-1…AC-10, 커밋 뒤 mutation, GEARS 임시 시험.

## Decisions

- D1 요구 문장은 영어 GEARS, 나머지 본문은 한국어, 제목은 영어(SDD 기반 SPEC과 같은 형식).
- D2 Rosetta 2: Apple silicon에서 x86_64 셸로 실행하면 `uname -m`이 `x86_64`라서, `sysctl -n
  sysctl.proc_translated`가 `1`이면 arm64로 봐요(Apple 문서 "About the Rosetta Translation Environment"가
  번역 중인 프로세스를 가리는 값으로 소개하는 `sysctl.proc_translated`; 이번에 웹에서 다시 읽지는 않았어요). 되돌리기: `install.sh`의 `if
  [ "$os" = "Darwin" ] …` 세 줄과 `install.ps1` `Get-UnixTarget`의 첫 줄, 시험의 해당 경우를 지워요.
- D3 Windows 판정을 `Get-WindowsTarget` 함수로 떼어 내 macOS의 `pwsh`에서도 시험해요(스크립트에 시험용 환경 변수를
  더하지 않으려고). ARM64 Windows에서 x64 에뮬레이션 프로세스가 `AMD64`로 보고하면 x86_64 바이너리를 받아요(에뮬레이션
  실행, 판정하지 않음).
- D4 지원하지 않는 플랫폼 메시지는 `Unsupported platform: <uname -s> <uname -m>`(원래 값)와 `Supported platforms:
  macOS arm64 (Apple silicon), Linux x86_64, Windows x86_64`예요. 예전 두 메시지(`Unsupported OS`, `Unsupported
  architecture`)는 하나로 합쳤어요.
- D5 릴리스(dist) target은 바꾸지 않아요: 릴리스 target은 별도 사용자 선택이에요. 그 전까지 Linux x86_64와 Windows
  x86_64 설치는 플랫폼 판정은 통과하지만 릴리스 자산이 없어 내려받기에서 멈춰요.
- D6 `docs/support-matrix.md`의 2026-11-09 CI 측정 표(Linux arm64·macOS x86_64 행)는 이력으로 남기고 그렇게 적었어요.
- D7 `tools/ci-install-ripgrep.sh`의 aarch64 Linux 분기는 그대로 둬요: workflow가 더는 그 runner를 쓰지 않지만
  스크립트 자체는 플랫폼 목록이 아니에요.
- D8 `Moai-Spec` 트레일러를 커밋에 넣어요. `Moai-Session`은 관문 조각(SDD-6)이 정할 때까지 넣지 않아요.

## Result

(구현 뒤 채워요.)
