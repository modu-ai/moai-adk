---
id: SPEC-20261117-headver-version
title: HeadVer 빌드 버전 — 카드 C-67 조각 1 (AC1·AC2)
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-sdd-foundation
---

## Intent

사용자 결정(2026-11-17): 이 세대의 개발 빌드는 HeadVer `{head}.{yearweek}.{build}`로 번호를 매기고,
다음 로컬 빌드(빌드 5)는 `2026.11.0-rc.2+local.5`가 아니라 `4.2647.5`로 보여야 해요. 지금은 빌드 번호를
`+local.<n>` 메타데이터로만 붙여서 `--version`, 상태줄, update의 버전이 옛 CalVer 모양이에요. 이 SPEC은
카드 C-67의 AC1(버전을 계산하는 함수 하나)과 AC2(그 버전을 쓰는 곳, 옛 rc 설치에서 깨끗하게 넘어오기)만
다뤄요.

## Non-goals

- `install/install.sh`·`install/install.ps1`, 워크플로, 태그 규칙, 릴리스 빌드 번호 찍기(카드 AC3·AC4):
  다른 에이전트가 카드 C-89로 `install/**`와 `ci.yml`을 고치는 중이에요.
- `Cargo.toml`의 패키지 버전 변경, 설치 기록(`install-record`)의 `version`·`channel`, 생성기의 `generator_version`.
- 버전만 바꾸는 커밋(HeadVer 규칙상 하지 않아요).

## Requirements

- REQ-1: When the build runs, the build script shall compute the version as `{head}.{yearweek}.{build}` from the one-line `HEADVER` file, the ISO-8601 week of the UTC build date and the build number. · failure: none — 실패 경로는 REQ-2·REQ-3이에요.
- REQ-2 (REQ-1): IF the `HEADVER` file is missing or is not one decimal number from 0 to 9999, THEN the build script shall stop the build with a message that names the file.
- REQ-3 (REQ-1): IF `MOAI_LOCAL_BUILD` is set to anything but a decimal number from 1 to 4294967295 without a sign or leading zero, THEN the build script shall stop the build.
- REQ-4: The version function shall take the ISO week-year and week of the UTC date, so the last days of December can belong to week 1 of the next year and the first days of January to week 52 or 53 of the year before.
- REQ-5: Where `MOAI_LOCAL_BUILD` is not set, the build shall use build number 0 and mark itself as an unnumbered build.
- REQ-6: The `--version` first line shall stay `<name> <version> (rust)`, and its second line shall say whether the build is a local build or an unnumbered build.
- REQ-7: The comparison function shall order every HeadVer version above every `YYYY.M.P[-rc.N]` version of the earlier scheme and order HeadVer versions by head, then yearweek, then build.
- REQ-8: The status line shall show its own version as the update target only when that version is newer than the template version.
- REQ-9: When `update-plan` finds a host receipt, the plan shall record the receipt's version as the source and this build's version as the target. · failure: none — 막는 경우는 REQ-10이에요.
- REQ-10 (REQ-9): IF the host receipt's version is newer than this build's version, THEN the plan shall block that host so the build never updates back.

## Acceptance

- AC-1 (REQ-1, REQ-4, REQ-5): 버전 함수가 연 경계(2026-12-31·2027-01-01·53주 해·2099·2100)에서 맞는 yearweek를 내요.
  verify: `cargo test --locked --offline --lib headver::`
- AC-2 (REQ-2, REQ-3): 머리 파일과 빌드 번호의 받는 값·거절하는 값이 경계(0/9999/10000, 1/4294967295/4294967296)에서 갈려요.
  verify: `cargo test --locked --offline --lib -- headver:: local_build::`
- AC-3 (REQ-6): `--version`의 첫 줄과 둘째 줄.
  verify: `cargo test --locked --offline --bin moai version_flag_reports_the_headver_build`; `cargo test --locked --offline --test cli_name_moai`
- AC-4 (REQ-7): 비교 순서(HeadVer > 옛 rc, head → yearweek → build, 메타데이터 무시, 모르는 모양은 비교하지 않음).
  verify: `cargo test --locked --offline --lib headver::`
- AC-5 (REQ-8): 상태줄은 더 새 바이너리만 업데이트 대상으로 보여요.
  verify: `cargo test --locked --offline --lib statusline::state::`
- AC-6 (REQ-9, REQ-10): 임시 HOME에서 옛 `2026.11.0-rc.2+local.4` 설치 기록 위로 이 빌드가 깨끗하게 update하고, 더 새 HeadVer 기록 위로는 막혀요.
  verify: `cargo test --locked --offline --test update_headver`
- AC-7 (REQ-1…REQ-10): 이웃 검사가 모두 통과해요.
  verify: `cargo test --locked --offline --bin moai`; `cargo test --locked --offline --test help_layout`;
  `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --test home_isolation_scan`;
  `cargo test --locked --offline --test statusline_golden`; `cargo fmt --check`;
  `cargo clippy --locked --offline --all-targets -- -D warnings`;
  `python3 -m unittest tests/repo_scans_test.py tests/ci_scope_test.py`

## Scope

- write: `HEADVER`, `build.rs`, `src/headver.rs`, `src/local_build.rs`, `src/lib.rs`, `src/main.rs`,
  `src/statusline/state.rs`, `src/update.rs`, `tests/update_headver.rs`, `tests/cli_name_moai.rs`,
  `docs/local-rc1-kit.md`, `AGENTS.local.md`, `.moai/specs/SPEC-20261117-headver-version/spec.md`
- read-only: `skills/moai/references/internal/versioning.md`, `AGENTS.md`, `src/install.rs`, `src/bin_names.rs`
- forbidden: `install/**`, `.github/workflows/**`, `Cargo.toml`의 `version`, 실제 HOME, 네트워크 쓰기

## Plan

1. 시험 먼저: `src/headver.rs`의 단위 시험(경계 날짜·비교·머리 파일), `--version` 시험, update 통합 시험.
2. `src/headver.rs`(빌드 스크립트와 라이브러리가 함께 씀), `build.rs`가 `MOAI_BUILD_VERSION`과 빌드 설명을 냄.
3. `--version` 둘째 줄, 상태줄 비교, update의 대상 버전과 되돌아가기 막기.
4. 이웃 시험, mutation, `MOAI_LOCAL_BUILD=5` 빌드의 `--version`.

## Risks

- 번호 없는 빌드(CI, `cargo build`)도 `4.<yearweek>.0`이 돼요. AC3가 릴리스 번호를 찍기 전에 릴리스를 만들면
  태그(`moai-cli-v2026.11.0-rc.N`)와 바이너리 버전이 달라요 → Decisions와 남은 일에 적었어요.
- 2100년에는 yearweek의 두 자리 연도가 `00`으로 돌아가요 → 그 전에 head를 올려야 순서가 지켜져요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | head는 저장소 뿌리의 한 줄 파일 `HEADVER`(지금 `4`), 사용자만 올림 | 카드 AC1, AGENTS.local.md | 다른 경로로 옮기기 |
| D2 | yearweek는 빌드 시각의 UTC ISO 주(`SOURCE_DATE_EPOCH`가 있으면 그 시각) | 사용자 결정, versioning.md | 지역 시각 |
| D3 | 빌드 번호: `MOAI_LOCAL_BUILD`, 없으면 0(번호 없는 빌드); 릴리스 번호는 AC3 | 카드 AC1 | 번호 없으면 옛 패키지 버전 |
| D4 | 로컬과 다른 빌드의 구분은 `--version` 둘째 줄(`local build 5 …`/`unnumbered build …`); 폴더·비교 키는 `4.2647.5` | 조정자 제안 | `+local` 메타데이터 |
| D5 | `Cargo.toml` 버전은 그대로: 버전을 파일에 적으면 버전만 바꾸는 커밋이 생겨요 | HeadVer 규칙 | AC3에서 다시 정하기 |
| D6 | 옛 CalVer(첫 수 2000 이상)는 모든 HeadVer보다 낮음; 모르는 모양은 비교하지 않음(막지 않음) | 카드 AC2 | 순서 표 바꾸기 |
| D7 | 2099는 `99xx`, 2100은 `00xx`(두 자리 연도 그대로, head로 순서를 지킴) | HeadVer README | 세 자리 연도 |

## Result

아직 비어 있어요(구현 뒤 채워요).
