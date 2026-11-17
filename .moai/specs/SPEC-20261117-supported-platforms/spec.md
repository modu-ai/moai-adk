---
id: SPEC-20261117-supported-platforms
title: 지원 플랫폼 세 개 — CI target과 설치 스크립트, 지원 문서 (조각 1)
status: approved
revision: 1
created: 2026-11-17
parent: null
approval: {by: user, at: 2026-11-17, basis: "in-session selection 2026-11-17 08:22:21Z: 세 SPEC 승인 — 감사 PASS b7f77edf35 내용 그대로", digest: "sha256:8263fe2cae97fe8a7a227021167de817f7d4a1a3fcdc42e6c2fc191edd3d17f2"}
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
- `src/lane_next.rs:3518`·`tests/gate_collect.rs:1848`의 시험 자료에 남은 옛 check 이름(동작과 무관, 후속).
  `skills/moai/workflows/verify.md`의 CI target 문장과 `src/flow.rs` 주석은 감사 뒤 범위에 넣었어요(D9).

## Requirements

요구 문장은 영어 GEARS로 써요(`src/spec.rs` `parse_gears`가 읽는 형식). `IF …, THEN …`은 원치 않는 동작이에요.

- REQ-1: When a pull request targets the integration branch, the `ci` workflow shall run the check set on `linux-x86_64` only. · failure: none — 한 target 실패는 기존 규칙대로 `ci`를 실패시켜요.
- REQ-2: When a push merges into the integration branch, the `ci` workflow shall run the check set on `linux-x86_64` and `macos-arm64` and run the `windows-build` job. · failure: none — 한 target 실패는 기존 규칙대로 `ci`를 실패시켜요.
- REQ-3: The workflows under `.github/workflows/` shall name no Intel macOS or Linux arm64 runner, target or nextest asset.
- REQ-4: The candidate verdict shall require exactly the targets `linux-x86_64`, `macos-arm64` and `windows-x86_64`.
- REQ-5: The aggregate check of the `ci` workflow shall keep the name `ci`.
- REQ-6: Every action a workflow uses shall stay pinned to a full commit SHA.
- REQ-7: When an installer runs on macOS arm64 (an x86_64 shell under Rosetta 2 included), Linux x86_64 or Windows x86_64, the installer shall select that platform's target triple.
- REQ-8 (REQ-7): IF an installer runs on any other platform, THEN the installer shall exit 1 before any download or write.
- REQ-9 (REQ-7): IF an installer stops for the platform, THEN its message shall name the platform and the three supported platforms.
- REQ-10: Each support document shall name the three supported platforms and state that Intel macOS and Linux arm64 are not supported.
- REQ-11: The change shall carry its own SPEC at `.moai/specs/SPEC-20261117-supported-platforms/spec.md`, tracked by git in the same commit.
- REQ-12: The verify workflow of the `moai` skill shall name `ci / linux-x86_64`, `ci / macos-arm64` and `windows-build` as the post-merge set.

실패 경로: REQ-7 → REQ-8, REQ-9. REQ-1·REQ-2 → 없음(줄 끝 `failure: none`).
REQ-10의 "support document"는 `docs/support-matrix.md`, `docs/host-support-matrix.md`, `README.md`예요.
REQ-7의 Rosetta 판정은 시스템의 `/usr/sbin/sysctl`만 물어요(PATH 앞의 `sysctl`은 묻지 않아요, D10).

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
- AC-5 (REQ-7): 가짜 `uname`으로 Darwin arm64·aarch64, Linux x86_64·amd64가 각 triple로 잡혀요. Apple
  silicon에서는 진짜 `arch -x86_64` 셸(진짜 `uname`과 `/usr/sbin/sysctl`)로 `install.sh`가 arm64를 잡아요.
  `install.ps1`은 같은 시험을 `pwsh`로 돌리고, 구문 트리에서 꺼낸 `Get-UnixTarget`(번역 중 Darwin x86_64 → arm64)과
  `Get-WindowsTarget`(AMD64·x64·x86(64비트 OS))을 확인해요.
  verify: `python3 -m unittest tests.install_script_test.InstallShTest.test_the_three_supported_platforms_are_detected tests.install_script_test.InstallShTest.test_an_x86_64_shell_under_rosetta_on_apple_silicon_is_arm64 tests.install_script_test.InstallPs1Test.test_the_three_supported_platforms_are_detected tests.install_script_test.InstallPs1Test.test_unix_targets_take_rosetta_as_arm64_and_nothing_else tests.install_script_test.InstallPs1Test.test_windows_takes_only_x86_64`
- AC-6 (REQ-8, REQ-9): Darwin x86_64(PATH 앞에 "번역 중"이라고 답하는 가짜 `sysctl`이 있어도), Linux
  aarch64·arm64·i686, FreeBSD amd64는 exit 1, HOME 트리 변화 없음, `Unsupported platform: <os> <arch>`와 지원 목록이
  나와요. Windows ARM64·IA64·x86(32비트 OS)는 `Get-WindowsTarget`이 `$null`이에요.
  verify: `python3 -m unittest tests.install_script_test.InstallShTest.test_an_unsupported_platform_stops_with_the_supported_list_and_writes_nothing tests.install_script_test.InstallPs1Test.test_an_unsupported_platform_stops_with_the_supported_list_and_writes_nothing tests.install_script_test.InstallPs1Test.test_the_rosetta_question_goes_to_the_system_sysctl tests.install_script_test.InstallPs1Test.test_windows_takes_only_x86_64`
- AC-7 (REQ-10): verify: `python3 -I tools/check_support_matrix.py docs/support-matrix.md docs/e2e-receipts/` → exit 0;
  `grep -c 'x86_64-pc-windows-msvc' docs/support-matrix.md docs/host-support-matrix.md README.md` → 각 1 이상;
  `python3 -m unittest tests/support_matrix_test.py tests/host_matrix_test.py tests/cargo_metadata_test.py tests/oss_docs_test.py`
- AC-8 (REQ-11): verify: `git ls-files .moai/specs/SPEC-20261117-supported-platforms` → 이 파일
- AC-9 (REQ-1…REQ-12): 저장소의 SPEC 검사기가 이 SPEC을 받아들여요(오류 0).
  verify: `moai spec check --all --project .` → exit 0; `cargo test --locked --offline --test spec_cli`
- AC-10 (REQ-1…REQ-12): 이웃 검사가 모두 통과해요. verify: `python3 -m unittest tests/install_script_test.py tests/github_assets_test.py tests/candidate_summary_test.py tests/ci_scope_test.py tests/repo_scans_test.py tests/ci_install_ripgrep_test.py tests/local_check_test.py tests/self_check_test.py tests/host_live_harness_test.py tests/goal_block_test.py tests/support_matrix_test.py tests/host_matrix_test.py tests/cargo_metadata_test.py tests/oss_docs_test.py`;
  `bash -n install/install.sh`; `sh -n install/install.sh`; `cargo fmt --check`;
  `cargo clippy --locked --offline --all-targets -- -D warnings`; `cargo test --locked --offline --bin moai`;
  `cargo test --locked --offline --test launcher_three_hosts`; `cargo test --locked --offline --test gitignore_block_cli`
- AC-11 (REQ-12): `skills/moai/workflows/verify.md`가 병합 뒤 실행을 `ci / linux-x86_64`, `ci / macos-arm64`,
  `windows-build`로 적고, skill 원본 manifest가 맞아요.
  verify: `grep -c 'ci / macos-arm64' skills/moai/workflows/verify.md` → 1 이상; `grep -c 'all four' skills/moai/workflows/verify.md` → 0;
  `python3 -I tools/source_manifest.py check` → exit 0; `cargo test --locked --offline --test agent_defs_package`

## Scope

- write: `.github/workflows/ci.yml`, `.github/workflows/candidate.yml`, `.github/workflows/candidate-target.yml`,
  `install/install.sh`, `install/install.ps1`, `install/README.md`, `tools/candidate_summary.py`,
  `tests/github_assets_test.py`, `tests/candidate_summary_test.py`, `tests/install_script_test.py`,
  `docs/support-matrix.md`, `docs/host-support-matrix.md`, `README.md`,
  `.moai/specs/SPEC-20261117-supported-platforms/spec.md`; 감사 뒤(D9) `skills/moai/workflows/verify.md`,
  `skills/moai/references/source-manifest.json`, `src/flow.rs`(주석만)
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
5. 감사 수리: tip 병합, SPEC을 `moai spec check` 형식에 맞춤, `/usr/sbin/sysctl` 절대 경로와 시험, 거절 exit 1
   고정, `verify.md`·`src/flow.rs` 문장.

## Risks

- 릴리스 자산이 아직 `aarch64-apple-darwin`뿐이라, Linux x86_64·Windows x86_64 사용자는 플랫폼 판정을 통과한 뒤
  내려받기에서 멈춰요(D5). 릴리스 target 선택 전에는 이 두 플랫폼에서 설치가 끝나지 않아요.
- PR CI는 여전히 Linux만 돌아서 macOS arm64 문제는 병합 뒤에야 보여요(CI 재설계의 병합 큐 단계가 다뤄요).
- Windows 분기(`Get-Target`의 `$IsWin` 경로)는 Windows 기계에서 실행하지 않았어요. 판정 함수만 구문 트리로 시험해요.
- 번역 중인(x86_64) Python으로 시험을 돌리면 "PATH 앞 가짜 `sysctl`" 경우가 진짜 `/usr/sbin/sysctl`의 `1`을 받아
  실패해요. 개발 Mac과 CI의 Python은 네이티브라 해당하지 않아요.

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
- D9 쓰기 범위 확대(리더 결정, 2026-11-17 감사 FAIL 뒤 수리 지시 N1): 카드 범위 밖이던
  `skills/moai/workflows/verify.md`(병합 뒤 target 문장)와 `src/flow.rs`(주석 "five targets")를 이 SPEC에서 고쳐요.
  skill이 바뀌어 `tools/source_manifest.py write`로 manifest를 다시 만들었어요. 되돌리기: 두 문장을 이전 커밋 값으로
  돌리고 manifest를 다시 써요.
- D10 Rosetta 판정은 `/usr/sbin/sysctl`을 절대 경로로 불러요(감사 N3). PATH 앞의 다른 `sysctl`이 판정을 바꾸지
  못해요. 수리 전 스크립트는 `PATH=/usr/bin:/bin`에서 `sysctl`을 못 찾아 진짜 Rosetta 셸을 거절했어요(아래 Result).
  되돌리기: `install.sh`·`install.ps1`의 `/usr/sbin/sysctl`을 `sysctl`로.
- D11 이 SPEC은 초안이라 앞머리에 `approval`이 없어요(검사기는 승인 뒤 상태에서만 요구해요). 승인은 사용자가 해요.

## Result

구현 에이전트가 2026-11-17에 채웠어요. SYNC와 AUDIT가 다시 확인해요. macOS arm64(개발 Mac), Python 3.13.1 +
PyYAML 6.0.3, `pwsh` 있음, `shellcheck` 없음. cargo는 설치된 slot이 거절해서(옛 swap 규칙, exit 75) slot 밖에서
한 번에 하나, `CARGO_BUILD_JOBS=4`, `CARGO_INCREMENTAL=0`, 빌드 폴더는 이 worktree의 `target/`.

커밋: 기준 `da4d9e3d10` 위에 `a0de2f9d18`(변경), 그 위에 이 Result.

| AC | 결과 | 증거 |
|---|---|---|
| AC-1 | 충족 | 4개 시험 OK. RED(변경 전): 병합 뒤 target 목록·macOS 공유 그룹 단언 실패 |
| AC-2 | 충족 | 시험 OK, `grep` → 출력 없음 exit 1. RED: `ci.yml`·`candidate.yml`에서 이름마다 실패(10건) |
| AC-3 | 충족 | OK. RED: `candidate_summary`가 다섯 target을 요구해 PASS 시험·CLI 시험 실패 |
| AC-4 | 충족 | 시험 OK, `actionlint` exit 0 |
| AC-5 | 충족 | sh·pwsh OK. 실제 기계: 네이티브 `Darwin arm64`(proc_translated 0)와 `arch -x86_64` 셸(`x86_64`, proc_translated 1) 모두 `Detected platform: aarch64-apple-darwin`, HOME 쓰기 없음. mutation: Rosetta 비교를 `"2"`로 → `Darwin x86_64 translated=1` 실패 |
| AC-6 | 충족 | sh·pwsh OK. RED: 옛 메시지(`Unsupported OS`/`architecture`)와 Intel·arm64 수용으로 실패. mutation: `Darwin/x86_64`를 받아들이면 2건 실패; `Get-WindowsTarget`에 ARM64를 더하면 `ARM64:1 -> aarch64-pc-windows-msvc`로 실패 |
| AC-7 | 충족 | `check_support_matrix` exit 0(지원 행 6 = passed receipt 6); 문서 시험 OK |
| AC-8 | 충족 | `git ls-files` → 이 파일 |
| AC-9 | 충족 | 임시 시험(`tests/zz_gears_probe_c89.rs`, 실행 뒤 삭제, 커밋 안 함): `PROBE total=11 gears=9 gears_ok=9 unwanted=2`. 처음 REQ-10은 주어에 쉼표가 있어 `text before the subject is not a Where, While or When clause`였고, 문장을 고친 뒤 ok |
| AC-10 | 일부 | Python 14개 파일 308개 중 307 통과, 1 실패(아래). `bash -n` 0, `cargo fmt --check` 0, clippy 0, `--bin moai` 0(27), `launcher_three_hosts` 0(4), `gitignore_block_cli` 0(32) |

AC-10의 실패 1건: `InstallPs1Test.test_the_locale_is_read_from_lc_all_then_lc_messages_then_lang`의
`{'LC_ALL': ''}` 경우가 `ko`를 받아요(기대 `en`). 지역 변수가 비면 `install.ps1`이 `Get-Culture`를 읽는데 이 Mac의
시스템 지역이 `ko_KR`이에요. 기준 커밋의 `install.ps1`로 같은 시험을 돌려도 같은 실패라서(이 변경과 무관, 기계
의존) 이 SPEC에서 고치지 않았어요. 후속으로 남겨요.

실행 안 함: Windows 분기의 실제 실행(Windows 기계 없음, `Get-WindowsTarget` 구문 트리 시험으로 대신), Linux 기계의
설치 실행, `shellcheck`(설치 안 됨, 시험은 skip), GitHub Actions의 실제 run(push 안 함), 전체 `cargo test`.

### 감사 수리 (2026-11-17 07:59Z–08:48Z)

독립 감사(head `4d128c7bdd`)가 FAIL이었어요: B1 — tip의 `moai spec check`가 이 SPEC에서 오류 9개
(`approval_field` ×4, `section_missing`, `ac_ref_unknown`, `req_failure_path` ×3)를 내서 병합 트리의
`tests/spec_cli.rs::the_repository_specs_pass_check`가 빨개져요. 비차단 N1·N3·N5도 함께 고쳤어요.

커밋: `4d128c7bdd` 위에 `dc3ee64e31`(origin/v2026.11.0 `c0b858de8c` 병합, `git merge-tree` 충돌 없음) →
`021c1aaba1`(수리) → 이 Result. 병합 커밋의 트레일러 블록에는 `Co-Authored-By`만 있어요(`Moai-Spec`이 따로 된
본문 문단에 들어감, amend 하지 않음).

- AC-9 (새 뜻, 검사기 통과): RED — 병합한 바이너리로 수리 전 SPEC을 검사하면 exit 1, 오류 9개(감사와 같은 줄).
  GREEN — `moai spec check --all --project .` exit 0, SPEC 3개 모두 오류 0·경고 0; `--test spec_cli` exit 0, 17 passed.
  (예전 AC-9인 GEARS 임시 시험 결과는 위 표에 그대로 있어요.)
- AC-5·AC-6 (N3, N5): RED — 수리 전 스크립트로 새 시험을 돌리면 4개 실패: PATH 앞 가짜 `sysctl`이 Intel Mac을
  arm64로 바꿈(sh·pwsh), 진짜 `arch -x86_64` 셸이 거절됨(시험 PATH `/usr/bin:/bin`에 `sysctl`이 없어서 — 수리 전
  스크립트의 실제 결함), ps1 본문에 `& sysctl`. GREEN — 모두 통과, 거절 시험은 exit 1을 고정해요. mutation:
  `install.sh`를 PATH의 `sysctl`로 되돌리면 2개 실패, `install.ps1`을 되돌리면 2개 실패.
- AC-11 (N1, D9): `verify.md`에 `ci / macos-arm64` 있음, `all four` 없음; `source_manifest.py write` →
  `workflows/verify.md` 바뀜, `check` exit 0; `--test agent_defs_package` exit 0, 17 passed.
- AC-10 다시: Python 14개 파일 312개 중 308 통과, 1 실패(위 지역 시험, 그대로), skip 3(`shellcheck` 2, pwsh Rosetta
  실행 1); `--bin moai` 0(27); `launcher_three_hosts` 0(4); `gitignore_block_cli` 0(32); `cargo fmt --check` 0;
  clippy 0; `actionlint` 0; `bash -n`·`sh -n` 0; `check_support_matrix` 0(6 = 6); manifest `check` 0.

후속(이 조각 밖): 릴리스 target 선택과 그 자산(N2, D5), 감사 N4·N6, Windows 시험 계획, 위 지역 시험,
`src/lane_next.rs:3518`·`tests/gate_collect.rs:1848`의 옛 check 이름(시험 자료).
