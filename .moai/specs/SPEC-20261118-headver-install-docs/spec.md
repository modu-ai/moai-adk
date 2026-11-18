---
id: SPEC-20261118-headver-install-docs
title: "The install scripts and the release docs name the HeadVer version scheme"
status: draft
revision: 1
created: 2026-11-18
---

# The install scripts and the release docs name the HeadVer version scheme

## Intent

카드 C-67의 마지막 조각 중 seq 4만이에요(리더 note 2075·2170): `install/install.sh`·`install/install.ps1`의 도움말·주석과 `docs/release-checklist.md`·`docs/local-rc1-kit.md`가 새 버전 체계 HeadVer `4.{yearweek}.{build}`를 말하게 해요. 설치 스크립트는 이미 `4.2647.12` 같은 글자를 버전으로 받아요(숫자로 시작하는 SemVer 모양). 그래서 코드 동작은 바꾸지 않고, 사용자가 보는 도움말 예시(옛 `2026.11.0`)와 채널 설명(옛 "release candidates", 하이픈 접미사 가정)을 고치고, 동작을 시험으로 고정하고, 두 문서에 HeadVer 절을 더해요.

## Non-goals

- seq 3: 릴리스 workflow, 태그 번호, CI 번호 규칙은 건드리지 않아요. 사용자 결정 D5·D6(첫 HeadVer 릴리스의 번호·태그, CI 번호 정렬 D4)을 기다려요. 이 SPEC의 `## Result`는 seq 3을 withdrawn으로 두지 않고 "이 SPEC 밖"이라고 적어요.
- 설치 스크립트의 판단 규칙(어떤 릴리스를 고르는지) 바꾸기. Atom 피드로 돌아갈 때 HeadVer 후보를 가려내지 못하는 점은 시험으로 사실만 고정하고 Risks에 적어요.
- `docs/release-checklist.md`의 rc.1 기록과 "릴리스 게시 절차"의 옛 번호 예시는 기록이라 그대로 둬요.
- 실제 HOME에 설치 스크립트를 돌리지 않아요. 시험은 임시 HOME과 file:// 가짜 릴리스만 써요.

## Requirements

- REQ-1: The install scripts shall show a HeadVer version in the `--version` help example.
- REQ-2: The install scripts shall describe the `rc` channel as taking releases marked as pre-releases.
- REQ-3: When a HeadVer version is given with `--version`, with or without the `v` or `moai-cli-v` prefix, the install script shall place it in the version folder named by that version. · failure: none — a text that is not version-shaped is refused by the existing check, which this SPEC does not change
- REQ-4: The install script shall take a release that the release list marks as a pre-release only on the `rc` channel.
- REQ-5: When the release list cannot be read and the release feed is used, the install script shall take the newest `moai-cli-v*` release. · failure: none — the feed carries no pre-release mark, so there is nothing to refuse on
- REQ-6: The release checklist shall state the HeadVer form, the tag form, how a HeadVer pre-release is marked, what the install scripts do with it, and that the first release number waits for the user's decisions D5 and D6.
- REQ-7: The local RC1 kit shall state how a HeadVer version installs and that the install scripts are tested only against a temporary HOME.

## Acceptance

- AC-1 (REQ-1, REQ-2): `install.sh --help` shows `--version 4.2647.12`, no `2026.11.0`, and the pre-release wording for `--channel`.
  verify: `python3 -m unittest tests.install_script_test.InstallShTest.test_help_examples_and_channel_name_the_headver_scheme`
- AC-2 (REQ-1, REQ-2): the same text is in `install/install.ps1`, read as text and also run when `pwsh` exists.
  verify: `python3 -m unittest tests.install_script_test.HeadVerTextTest tests.install_script_test.InstallPs1Test.test_help_examples_and_channel_name_the_headver_scheme`
- AC-3 (REQ-3, REQ-4): HeadVer versions install into `versions/4.2647.12`, and the channel follows the release list's pre-release mark.
  verify: `python3 -m unittest tests.install_script_test.InstallShTest.test_headver_versions_install_and_the_channel_follows_the_prerelease_flag`
- AC-4 (REQ-5): with only the feed, the newest `moai-cli-v*` release is taken.
  verify: `python3 -m unittest tests.install_script_test.InstallShTest.test_the_feed_has_no_prerelease_mark_so_the_newest_headver_release_is_taken`
- AC-5 (REQ-6): `docs/release-checklist.md` has the HeadVer section.
  verify: `python3 -m unittest tests.install_script_test.HeadVerDocsTest.test_release_checklist_names_headver`
- AC-6 (REQ-7): `docs/local-rc1-kit.md` has the HeadVer install section.
  verify: `rg -c "^## 8\. 설치 스크립트로 받는 HeadVer 버전|install\.sh --version 4\.2647\.12|\*\*임시 HOME\*\*에서만 시험해요" docs/local-rc1-kit.md` prints 3 (matching lines)
- AC-7 (REQ-1…REQ-7): the neighbour checks and the scope.
  verify: `python3 -m unittest tests.install_script_test tests.github_assets_test tests.ci_scope_test`; `git diff --name-only origin/v2026.11.0...HEAD` lists only the Scope paths

## Scope

- write: `install/install.sh`, `install/install.ps1`, `docs/release-checklist.md`, `docs/local-rc1-kit.md`, `tests/install_script_test.py`, `tests/github_assets_test.py`, `.moai/specs/SPEC-20261118-headver-install-docs/spec.md`
- read-only: `src/headver.rs`, `skills/moai/references/internal/versioning.md`, `.github/workflows/**`, `AGENTS.local.md`
- forbidden: `.github/workflows/**`, any tag, the real HOME, the network

## Plan

1. 시험 먼저: 도움말·채널 설명(sh와 ps1), HeadVer 설치와 채널 고르기, 피드 폴백, 두 문서의 문장. 아직 없는 문구 때문에 도움말 시험과 문서 시험이 실패하는 것을 보고, 설치·채널 고르기 시험은 지금 동작을 지키는 시험이에요(바로 통과).
2. 두 스크립트의 도움말 예시·채널 줄·주석을 고쳐요(판단 코드는 그대로).
3. 두 문서에 HeadVer 절을 더해요.
4. 이웃 시험과 `pwsh` 시험을 돌려요. `shellcheck`는 로컬에 없어서 CI 결과로 확인해요.

## Risks

- Atom 피드로 돌아가면 HeadVer 후보와 정식판을 가릴 수 없어서 stable 채널이 후보를 받을 수 있어요(AC-4가 사실만 고정). REST 목록을 읽지 못하는 환경에서만 생겨요. `--version`으로 고정하면 피할 수 있다고 문서에 적어요. 규칙을 바꾸는 일은 prerelease 표시 방식의 사용자 결정이 따라와야 해서 이 SPEC 밖이에요.
- 도움말 글자가 바뀌어서 도움말 문구를 고정한 다른 시험이 깨질 수 있어요(`test_help_lists_the_go_options_and_unknown_options_fail`은 옵션 이름만 봐요). AC-7로 확인해요.

## Decisions

| id | decision | why | alternative |
| --- | --- | --- | --- |
| D-1 | 설치 스크립트의 판단 코드는 바꾸지 않아요 | `4.2647.12`는 이미 받아들여지고, REST 목록의 prerelease 표시로 채널이 갈려요 | 피드 폴백에서 후보를 추정(표시 방식의 사용자 결정이 먼저) |
| D-2 | rc.1 기록(옛 번호 예시)은 문서에 남겨요 | 리허설 기록이라 지우면 근거가 사라져요 | 모두 새 번호로 고쳐 쓰기 |
| D-4 | `docs/local-rc1-kit.md`는 시험으로 읽지 않고 `rg` 확인 명령으로 검증해요 | 문서를 읽는 시험은 `tools/ci_scope.py` 거부 목록에 그 경로가 있어야 하고(`ci_scope_test`가 요구해요) 그 파일은 이 카드 범위 밖이에요 | 범위를 넓혀 거부 목록에 한 줄 추가 후 시험으로 고정(후속) |
| D-3 | seq 3은 Result에서 "이 SPEC 밖"으로 적어요 | 리더 지시(사용자 결정 대기) | withdrawn |

## Result

C-67 seq 3 (release workflow, tag numbers, CI number rule) is outside this SPEC: it waits for the user's decisions D5 and D6 and nothing under `.github/` or any tag was touched.

- AC-1: met — RED first: `InstallShTest.test_help_examples_and_channel_name_the_headver_scheme` failed on the old text (`'--version 4.2647.12' not found`), passes after the edit; reverting the example in `install/install.sh` makes it fail again (mutant M2 caught).
- AC-2: met — `HeadVerTextTest` reads both scripts as text (RED first: 2 failures, `install.sh` and `install.ps1`); the pwsh run of the same help test passes on this Mac (`InstallPs1Test`, 3 new tests OK); reverting the ps1 channel wording fails `HeadVerTextTest` (mutant M3 caught).
- AC-3: met — `test_headver_versions_install_and_the_channel_follows_the_prerelease_flag` passes on sh and pwsh (stable takes `4.2646.9`, rc takes `4.2647.12`, `v` and `moai-cli-v` prefixes land in `versions/4.2647.12`); removing the pre-release check from `pick_version` fails it (mutant M1 caught). It pins existing behaviour, so it was green before the edit.
- AC-4: met — `test_the_feed_has_no_prerelease_mark_so_the_newest_headver_release_is_taken` passes on sh and pwsh; a `pick_version` that skips `4.*` tags fails it (mutant M5 caught). The limitation (the feed cannot tell a HeadVer candidate) is in `docs/release-checklist.md` (a-2) and in the script comments.
- AC-5: met — `HeadVerDocsTest.test_release_checklist_names_headver` (RED first, then green); dropping the "Atom 피드" sentence fails it (mutant M4 caught).
- AC-6: met — `rg -c` with the three patterns of the verify line prints 3 on `docs/local-rc1-kit.md` (section 8). Not pinned by a test: a test reading that path needs `tools/ci_scope.py`, outside the Scope (D-4).
- AC-7: met — `python3 -m unittest tests.github_assets_test tests.ci_scope_test tests.support_matrix_test` 65 OK; `tests.install_script_test` 120 run, 1 failure, 13 skipped: the failure is `InstallPs1Test.test_the_locale_is_read_from_lc_all_then_lc_messages_then_lang`, which fails the same way on the unchanged primary checkout (`1d49d55062`, `LANG=en_US.UTF-8`), so it is this Mac's pwsh locale, not this change; `git diff --name-only origin/v2026.11.0...HEAD` lists only the six Scope paths. `shellcheck` is not installed locally: the CI result is the check, not run here.
