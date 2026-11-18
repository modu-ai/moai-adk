---
id: SPEC-20261118-platform-targets
title: "Release targets are the three supported platforms; Windows test plan (slice 2)"
status: draft
revision: 1
created: 2026-11-18
parent: SPEC-20261117-supported-platforms
---

# Release targets are the three supported platforms; Windows test plan (slice 2)

## Intent

카드 C-89(사용자 결정 2026-11-17: "windows, macOS 실리콘, linux_x86_64 이렇게 제공하자. claude code, codex에서 지원하는 환경으로
동일하게 지원하자.")의 남은 조각이에요. 조각 1(`SPEC-20261117-supported-platforms`)은 CI target, 설치 스크립트, 지원 문서를 세 플랫폼에
맞췄고, 릴리스(dist) target은 D5로 남겼어요. 이 조각은 두 가지를 해요.

1. `dist-workspace.toml`의 `targets`를 세 플랫폼 triple로 늘려요. 설치 스크립트는 이미 세 triple을 고르고(`install/install.sh`,
   `install/install.ps1`) 자산 이름은 cargo-dist 규칙(`moai-cli-<triple>.tar.xz`, Windows는 `.zip`)을 따라요. 릴리스 자산이 없어
   Linux·Windows 설치가 내려받기에서 멈추던 틈(조각 1의 Risks)이 이 변경으로 닫혀요.
2. Windows에서 시험을 돌리는 계획을 적어요(카드의 seq 3). 계획만이고 CI에서 Windows 시험을 돌리는 일은 후속 조각이에요.

카드 C-89 기준 상태(리더 대조 note 2076, tip 1d49d55062): seq 4(문서)는 이미 충족, seq 1의 target 정리는 조각 1이 했고, 남은 것은
seq 1의 Windows 시험 실행, seq 2의 dist target, seq 3의 계획이에요.

## Non-goals

- `windows-build` job이 시험을 실행하게 하는 일. Windows 기계에서 아직 한 번도 돌려 보지 않은 시험을 병합 뒤 job에 넣으면 그 job이
  원인 모를 이유로 붉어지므로, 먼저 한 번 돌려 기준선을 재야 해요(후속 조각, 아래 Plan 3).
- 병합 큐(`merge_group`)와 그 target 구성.
- 릴리스 워크플로(`candidate*.yml`)의 구조, 서명, 태그와 공개(릴리스는 사용자 선택).
- `install/` 스크립트 변경(이미 세 triple을 골라요).

## Requirements

- REQ-1: The release configuration `dist-workspace.toml` shall list exactly the targets `aarch64-apple-darwin`, `x86_64-unknown-linux-gnu` and `x86_64-pc-windows-msvc`, each once.
- REQ-2: The release configuration shall name no Intel macOS and no Linux arm64 target.
- REQ-3: The release configuration shall include every target triple that `install/install.sh` or `install/install.ps1` selects.
- REQ-4: The installers `install/install.sh` and `install/install.ps1` together shall select every target of the release configuration.
- REQ-5: The change shall carry its own SPEC at `.moai/specs/SPEC-20261118-platform-targets/spec.md`, tracked by git in the same commit, with the Windows test plan in its Plan section.

## Acceptance

- AC-1 (REQ-1, REQ-2): `targets`가 세 triple이고 중복이 없으며 Intel macOS·Linux arm64가 없어요.
  verify: `python3 -I tests/platform_targets_test.py PlatformTargetsTest.test_dist_targets_are_exactly_the_three_supported_platforms PlatformTargetsTest.test_no_target_is_listed_twice PlatformTargetsTest.test_no_dropped_platform_is_a_dist_target`
- AC-2 (REQ-3, REQ-4): 두 설치 스크립트가 고르는 triple은 모두 dist target이고, 둘이 합쳐 모든 dist target을 골라요.
  verify: `python3 -I tests/platform_targets_test.py PlatformTargetsTest.test_each_installer_picks_only_triples_that_dist_builds PlatformTargetsTest.test_the_installers_between_them_pick_every_dist_target`
- AC-3 (REQ-5): SPEC이 git에 추적되고 SPEC 검사기가 받아들여요.
  verify: `git ls-files .moai/specs/SPEC-20261118-platform-targets`; `moai spec lint SPEC-20261118-platform-targets`
- AC-4 (REQ-1…REQ-5): 이웃 검사가 통과해요(`include` 설정, 릴리스 예행, 설치 스크립트, 워크플로 자산).
  verify: `python3 -I tests/dist_include_test.py`; `python3 -I tests/release_rehearsal_test.py`; `python3 -I tests/install_script_test.py`; `python3 -I tests/github_assets_test.py`

## Scope

- write: `dist-workspace.toml`, `tests/platform_targets_test.py`, `.moai/specs/SPEC-20261118-platform-targets/spec.md`
- read-only: `install/install.sh`, `install/install.ps1`, `.github/workflows/ci.yml`, `tests/*.rs`
- forbidden: 릴리스 태그·공개, 실제 HOME, 네트워크 쓰기

## Plan

1. 시험 먼저: `tests/platform_targets_test.py`(RED: 현재 `targets`는 `aarch64-apple-darwin` 하나라 5개 중 4개 실패).
2. `dist-workspace.toml`의 `targets`를 세 triple로.
3. Windows 시험 계획(카드 seq 3). tip 1d49d55062에서 잰 값이에요(글자 찾기로 센 값이라 가설이고, 판정은 Windows에서 실제로 돌려야 해요).
   - `tests/*.rs` 294개 중 파일 전체가 `#![cfg(unix)]`인 것 101개, `cfg(unix)`가 어디든 있는 것 163개, Windows용 `cfg`가 있는 것 2개
     (`grep -lE '^#!\[cfg\(unix\)\]' tests/*.rs | wc -l` 등).
   - 파일 전체가 `cfg(unix)`인 101개의 이유(글자 신호 기준): Unix API와 POSIX 도구를 둘 다 쓰는 것 67, Unix API(`std::os::unix`,
     symlink, `PermissionsExt`, pty 등)만 22, POSIX 도구(`sh`, `bash`, `chmod`, `/dev/null`)만 10, 신호 없음 2.
     → "이유가 있는" 쪽은 99개이고, 신호 없는 2개가 실수로 막힌 후보예요(파일을 열어 확인할 것).
   - 첫 조각(후속): 파일 전체가 `cfg(unix)`가 아닌 193개와 `--lib`·`--bin moai` 단위 시험을 `windows-build` job이 아니라
     `workflow_dispatch` 한 번으로 돌려 통과/실패 기준선을 먼저 잰다. 실패 원인을 "Unix 기능이 필요함(`cfg(unix)`로 옮김)"과
     "실수(고침)"로 나누고, 기준선이 초록이 된 뒤에야 `windows-build`에 시험 단계를 더한다.
   - 둘째 조각: 신호 없음 2개와 부분 `cfg(unix)` 파일의 Windows 대응을 하나씩 정한다.

## Risks

- dist target을 늘려도 릴리스 워크플로가 Linux·Windows 자산을 실제로 만드는지는 이 조각이 보여 주지 않아요. 이 저장소의 워크플로는
  `ci = []`(cargo-dist CI 생성 없음)이고 릴리스 공개는 사용자 선택이라, 자산 생성은 첫 릴리스 예행에서 확인해야 해요.
- Windows 분기(`install.ps1`의 `$IsWin` 경로)와 Windows 시험은 Windows 기계에서 실행해 본 적이 없어요(조각 1의 Risks 그대로).
- 위 Windows 숫자는 글자 찾기로 센 값이라 가설이에요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | dist `targets`를 세 triple로 늘리고, Windows 시험 실행은 후속 조각으로 미룬다 | 사용자 결정 2026-11-17(세 플랫폼); 로컬에서 Windows 시험을 돌릴 수 없어 기준선 없이 job에 넣으면 post-merge `windows-build`가 붉어질 위험 | `targets`를 이전 값 한 줄로 되돌리기 |
| D-2 | 조각 1의 D5("릴리스 target은 별도 선택")는 사용자 결정 원문("이렇게 제공하자")과 리더 지시(note 2080)로 풀린 것으로 본다 | 카드 C-89 seq 2, 리더 note 2076·2080 | `targets` 되돌리기(공개 전이라 외부 영향 없음) |
| D-3 | 시험은 TOML과 설치 스크립트의 triple 대입 줄만 읽는다(cargo-dist 실행 없음) | 오프라인·결정적 검사 | 없음 |

## Result

- AC-1: met — `python3 -I tests/platform_targets_test.py` 5개 OK. RED(변경 전): 5개 중 4개 실패(`targets`가 `aarch64-apple-darwin` 하나).
- AC-2: met — 같은 파일의 설치 스크립트 시험 두 개 OK. RED에서 `install.sh`는 `{x86_64-unknown-linux-gnu, aarch64-apple-darwin}`, `install.ps1`은 세 triple을 고르는데 dist는 하나만 만든다고 실패.
- AC-3: met — `moai-cli spec check SPEC-20261118-platform-targets`(tip 빌드) `errors:0 warnings:0`. 처음엔 REQ-3이 두 요구를 한 문장에 담아 `req_gears` 오류 1건이라 REQ-3·REQ-4로 나눔.
- AC-4: met — `dist_include_test` 4개, `release_rehearsal_test` 5개, `github_assets_test` 39개, `ci_scope_test` 19개 OK. `install_script_test`는 111개 중 실패 1, skip 7: 실패한 시험은 `InstallPs1Test.test_the_locale_is_read_from_lc_all_then_lc_messages_then_lang`이고 이 변경을 stash로 치운 tip에서도 같은 시험이 실패해요(조각 1 Result가 적은 로컬 환경 시험과 같음). 전체 실행의 실패 1건이 이 시험인지는 이름으로 직접 대조하지 않았어요.
- 돌리지 않은 것: Windows 기계와 cargo-dist 실행(Non-goals), 릴리스 예행.
