---
id: SPEC-20261117-headver-ci-number
title: HeadVer 빌드 번호 — CI 빌드와 릴리스(카드 C-67 조각 2, AC3)
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-headver-version
---

## Intent

조각 1(SPEC-20261117-headver-version)은 로컬 빌드에 `MOAI_LOCAL_BUILD`로 번호를 줬고, 번호 없는 빌드(CI·릴리스)는
`4.<yearweek>.0`이라 같은 주의 빌드가 모두 같은 버전이고 릴리스 tag(`moai-cli-v…rc.N`)와도 맞지 않아요. 이 조각은 카드 AC3
("릴리스 workflow가 build를 빌드 서버에서 찍고, 한 주의 두 빌드가 버전을 나누지 않고, 로컬 빌드가 릴리스 빌드와 구별된다")
중 사용자 답 없이 할 수 있는 것을 해요: CI 빌드가 `MOAI_BUILD_NUMBER`로 번호를 받고 `--version`이 CI 빌드라고 말해요. 릴리스
tag 모양과 릴리스 번호의 출처는 릴리스 정책이라 사용자 선택을 기다려요(Decisions D4–D6).

## Non-goals

- `install/**`(C-78이 고치는 중), 설치 기록의 버전(AC4).
- 릴리스 workflow 만들기와 tag ruleset 바꾸기(정식판 게시는 `docs/release-attestation.md`에서 HOLD).
- `Cargo.toml` 버전 바꾸기.

## Requirements

- REQ-1: Where `MOAI_BUILD_NUMBER` is set to a build number, the build shall use it as HeadVer's `{build}` and begin its `--version` note with `CI build <n> (MOAI_BUILD_NUMBER)`.
- REQ-2 (REQ-1): IF `MOAI_BUILD_NUMBER` is not a decimal number from 1 to 4294967295 without a sign or leading zero, THEN the build script shall stop the build with a message that names the variable.
- REQ-3 (REQ-1): IF both `MOAI_LOCAL_BUILD` and `MOAI_BUILD_NUMBER` are set, THEN the build script shall stop the build.
- REQ-4: When `ci.yml` builds, the job shall set `MOAI_BUILD_NUMBER` to the workflow run number on the run's first attempt and to nothing on a re-run. · failure: none — 다시 실행한 빌드는 번호 없는 빌드예요.
- REQ-5: The unnumbered build's note shall name both variables it did not get.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-5): CI 번호의 버전과 둘째 줄, 두 번호를 함께 주면 거절, 번호 없는 빌드의 둘째 줄.
  verify: `cargo test --locked --offline --lib -- headver:: local_build::`; `cargo test --locked --offline --bin moai version_flag_reports_the_headver_build`
- AC-2 (REQ-2): `MOAI_BUILD_NUMBER`의 받는 값과 거절하는 값(0/1, 4294967295/4294967296, `01`, `x`)과 거절 문장의 변수 이름.
  verify: `cargo test --locked --offline --lib local_build::tests::the_ci_number_has_the_same_bounds_and_names_its_variable`
- AC-3 (REQ-4): `ci.yml`의 `check`와 `windows-build` job이 첫 시도에서만 run number를 줘요.
  verify: `python3 -m unittest tests.github_assets_test`
  manual: `actionlint .github/workflows/ci.yml`
- AC-4 (REQ-1…REQ-5): 이웃 검사.
  verify: `cargo test --locked --offline --test update_headver`; `cargo test --locked --offline --test cli_name_moai`;
  `cargo test --locked --offline --test help_layout`; `cargo test --locked --offline --test product_cli`;
  `cargo test --locked --offline --bin moai`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `build.rs`, `src/headver.rs`, `src/local_build.rs`, `src/main.rs`, `.github/workflows/ci.yml`, `tests/github_assets_test.py`,
  `AGENTS.local.md`, `docs/local-rc1-kit.md`, `skills/moai/references/source-manifest.json`,
  `.moai/specs/SPEC-20261117-headver-ci-number/spec.md`
- read-only: `skills/moai/references/internal/versioning.md`, `docs/release-attestation.md`, `docs/release-tag-ruleset-moai-cli-v.md`
- forbidden: `install/**`, tag ruleset, 릴리스 게시, `Cargo.toml`의 `version`

## Plan

1. 시험 먼저: `headver::build_numbered`, `local_build::number_of`, `--version` 시험의 세 종류, workflow 고정 시험.
2. `build.rs`가 `MOAI_BUILD_NUMBER`를 읽고, `ci.yml` 두 job이 첫 시도에서만 run number를 줘요.
3. `AGENTS.local.md`, `docs/local-rc1-kit.md`에 한 줄씩.

## Risks

- `ci.yml`의 run number와 로컬 `MOAI_LOCAL_BUILD` 번호는 다른 계수기예요. 지금은 로컬 번호(한 자리)가 CI run number보다 훨씬
  작아 겹치지 않지만, 같은 주에 같은 숫자가 나오면 두 빌드가 같은 버전 문자열을 가져요(둘째 줄로만 구별) → D4.
- workflow마다 run number가 따로라서(`ci.yml`, `candidate.yml`) 두 workflow의 빌드가 같은 번호를 가질 수 있어요 → 이 조각은 `ci.yml`에만
  번호를 주고, 릴리스 빌드의 번호 출처는 D5.
- `MOAI_BUILD_NUMBER`가 run마다 바뀌어 `build.rs`가 매번 다시 돌아요(작업 공간 crate만 다시 빌드, 의존성 캐시는 그대로).
- cargo-dist는 tag의 버전과 패키지 버전이 맞기를 기대해요(`docs/release-attestation.md` 3.2): HeadVer tag로 릴리스하려면 D6의 결정이 먼저예요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | CI 번호는 새 변수 `MOAI_BUILD_NUMBER`(값의 규칙은 `MOAI_LOCAL_BUILD`와 같음); `MOAI_LOCAL_BUILD`를 CI에서 쓰지 않음 | `--version`이 로컬과 CI를 구별(카드 AC3) | 변수 하나로 합치기 |
| D2 | 두 변수를 함께 주면 빌드를 멈춤 | 추측하지 않음 | 하나를 우선 |
| D3 | `ci.yml`의 `check`·`windows-build`는 `run_attempt`가 1일 때만 `run_number`, 아니면 비움 | `internal/versioning`("a re-run keeps `run_number`") | 다시 실행도 번호 주기 |
| D4 | [NEEDS CLARIFICATION: 로컬 번호와 CI 번호가 같은 주에 같은 숫자일 때, 그리고 순서] (감사 N1) 수백 번대인 CI 빌드는 같은 주의 로컬 빌드보다 위로 정렬돼요(`4.2647.812` > `4.2647.5`): CI 빌드를 설치한 호스트에서는 update가 로컬 빌드를 "더 오래된 빌드"로 막아요(`--allow-older`가 필요). 지금 `ci.yml`은 빌드 산출물을 올리지 않아 CI 빌드가 설치될 길은 없어요. 권장: 로컬 빌드는 지금처럼 작은 계수기, 릴리스는 전용 계수기(D5) — 다른 선택: 로컬 빌드에 `-dev` 접미사, 로컬 번호를 CI보다 큰 범위에서 시작 | HeadVer 글의 `-dev` 관찰(규칙 아님) | — |
| D5 | [NEEDS CLARIFICATION: 릴리스 빌드의 `{build}` 출처] 권장: 릴리스 workflow 전용 계수기(첫 시도의 그 workflow run number, workflow 하나만 릴리스를 만듦) — 다른 선택: 저장소 변수에 둔 계수기, `ci.yml` run number 재사용 | 카드 AC3, `internal/versioning` 2단계 | — |
| D6 | [NEEDS CLARIFICATION: 릴리스 tag 모양] 권장: `moai-cli-v{head}.{yearweek}.{build}`(접미사 없음, RC는 GitHub pre-release 표시) — 다른 선택: 이번 세대 RC는 `moai-cli-v2026.11.0-rc.N` 유지, HeadVer는 다음 릴리스부터; cargo-dist의 tag·패키지 버전 일치 요구를 함께 정해야 함 | `internal/versioning` HeadVer pre-release 규칙, `docs/release-attestation.md` | — |

## Result

구현 에이전트가 채웠어요. 감사가 다시 확인해요. 기준 tip `17c19a2759`(중복 manifest 키는 `source_manifest.py write`로 고친 커밋
`90e22f933c` 먼저), 변경 커밋 `7b05e95f89`, 이 worktree의 `target/`, `CARGO_INCREMENTAL=0`, `CARGO_BUILD_JOBS=4`, macOS arm64.

RED(2026-11-17T16:59:44Z): `build_numbered`·`number_of`·`CI_ENV`가 없어 lib·bin 시험이 컴파일 실패(새 함수라 예상한 RED),
`github_assets_test`는 `MOAI_BUILD_NUMBER`가 없어 2개 오류. 통과 17:01:32Z.

- AC-1: met — `--lib headver:: local_build:: statusline:: update::` 141 passed, `--bin moai` 27(`version_flag_reports_the_headver_build`가 세 종류를 받음)
- AC-2: met — 같은 `--lib` 실행(`the_ci_number_has_the_same_bounds_and_names_its_variable`)
- AC-3: met — `github_assets_test` OK(`test_ci_builds_carry_the_run_number_on_the_first_attempt_only`), actionlint(ci·candidate 두 파일) 0
- AC-4: met — `update_headver` 3, `update_version_folders` 11, `cli_name_moai` 5, `help_layout` 10, `product_cli` 75, `home_isolation_scan` 16, `spec_cli` 17, `status_line_cli` 16, `statusline_tag_order` 1, `banner_golden` 12, 저장소 scan 대상 exit 0; fmt 0, clippy 0; Python 138 OK; manifest ok; `moai spec check --all` ok(이 SPEC 경고 3 = NEEDS CLARIFICATION 셋)
- 실제 빌드: `MOAI_BUILD_NUMBER=812 cargo build` → `moai-cli 4.2647.812 (rust)` / `CI build 812 (MOAI_BUILD_NUMBER), built 2026-11-17 UTC (ISO week 2026-W47), package 2026.11.0-rc.1`;
  `MOAI_BUILD_NUMBER=812 MOAI_LOCAL_BUILD=5 cargo build` → exit 101, "MOAI_LOCAL_BUILD and MOAI_BUILD_NUMBER are both set"

mutation 3개 모두 KILLED: CI 빌드가 local이라 말하기, 두 번호를 받기(로컬 우선), 다시 실행해도 run number 쓰기(처음 쓴 두 mutant 문장은
컴파일 실패·두 곳 일치로 다시 써서 돌림).

실행 안 함: GitHub에서 실제 CI 실행(push 없음), 릴리스 workflow(없음, D5·D6 대기), 전체 `cargo test`, Linux·Windows.
