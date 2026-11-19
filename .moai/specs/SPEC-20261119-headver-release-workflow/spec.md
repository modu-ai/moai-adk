---
id: SPEC-20261119-headver-release-workflow
title: "A release workflow that numbers, names and checks a HeadVer release, and never publishes one"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261117-headver-ci-number
---

# A release workflow that numbers, names and checks a HeadVer release, and never publishes one

## Intent

카드 C-67 의 seq 3 을 이어받은 카드 C-93 이에요. 사용자가 2026-11-18T22:03:35Z 에 D4·D5·D6 을 골랐어요(C-67 note 2245): 릴리스 `{build}` 는 릴리스를 만드는 workflow 하나의 첫 시도 run number(D5), 릴리스 tag 는 `moai-adk-v{head}.{yearweek}.{build}` 이고 접미사 없이 RC 는 GitHub pre-release 로 표시(D6), 로컬 빌드는 작은 계수기를 유지하고 둘째 `--version` 줄로 구분(D4).

이 SPEC 은 그 결정을 `.github/workflows/release.yml` 로 옮겨요. workflow 는 사람이 `release/**` 브랜치를 push 할 때만 돌고, 번호와 tag 이름을 정하고, 그 번호로 binary 를 빌드해 `--version` 이 tag·릴리스 노트와 같은지 확인하고, 사람이 tag 를 올릴 때 쓸 노트를 job 요약에 남겨요. tag 를 올리는 일과 릴리스를 게시하는 일은 workflow 가 하지 않아요(되돌릴 수 없는 일이라 사용자가 매번 확인, `SKILL.md` case 2).

## Non-goals

- 실제 tag push, 릴리스 게시, tag ruleset(`moai-cli-v*` → `moai-adk-v*`) 변경. 사용자가 매번 고르는 일이에요(guarded).
- 게시 job, 산출물 업로드, attestation, cargo-dist 생성 workflow. 사용자가 환경·ruleset 을 정한 뒤 후속 카드에서 해요(Decisions D-1, `docs/release-attestation.md` 6 의 승인 항목 1·2·6).
- `src/**`(빌드 번호를 새기는 코드는 이미 있어요: `build.rs`, `src/headver.rs`, `src/local_build.rs`), `install/**`, `Cargo.toml` 의 `version`, `HEADVER`.
- 로컬 빌드와 릴리스를 `update` 가 어떻게 줄 세우는지 바꾸기: 숫자 순서를 그대로 두고 사실을 시험으로 고정해요(D-6, 리더 결정 note 2260). 종류 표시로 순서를 바꾸는 B 는 필요하면 새 카드예요.
- `install/**` 고치기(설치 스크립트가 `moai-adk-v` tag 를 읽게 하기)와 `tests/install_script_test.py` 의 문구 단언 바꾸기. 문서 `docs/release-checklist.md` 는 리더가 C-67 을 풀어 준 뒤(2026-11-19) 이 카드가 고쳤고, 그 시험이 고정한 문구(`moai-cli-v4.2647.12`, `사용자 결정 D5·D6` 등)는 그대로 둬요.

## Requirements

- REQ-1: The release workflow shall start only from a push to a branch named `release/<40 lowercase hex>` or `release/rc/<40 lowercase hex>` whose name is the pushed commit, and from no other event. · failure: a name of any other shape, or one that names another commit, stops the plan before any build.
- REQ-2: The plan shall take `{build}` from the workflow's `run_number` and stop when the run is not the first attempt (`run_attempt` is not `1`) or when the run number is not a decimal number from 1 to 4294967295 without a sign or leading zero. · failure: a re-run keeps `run_number`, so it is refused and the person pushes a new release branch.
- REQ-3: The plan shall take `{head}` from the tracked `HEADVER` file of the pinned commit (0 to 9999, no sign, no leading zero) and `{yearweek}` from the ISO week of one UTC clock reading, and hand that reading on as the build's `SOURCE_DATE_EPOCH`.
- REQ-4: IF a tag `moai-adk-v<version>` already exists, or an existing `moai-adk-v<head>.<yearweek>.<build>` tag has a build number that is not below this run's, THEN the plan shall stop. · failure: two releases never share a version, and the build number only goes up, also when the workflow's counter restarted.
- REQ-5: The release tag shall be `moai-adk-v{head}.{yearweek}.{build}` with no suffix, a `release/rc/…` branch marking the release as a pre-release.
- REQ-6: The build job shall build with `MOAI_BUILD_NUMBER` set to the plan's build number and `MOAI_LOCAL_BUILD` set nowhere, failing unless the first `--version` line of both binaries is `moai-cli <version> (rust)` and the second begins `CI build <build> (MOAI_BUILD_NUMBER), built`.
- REQ-7: The build job shall fail unless the tag is `moai-adk-v` followed by the version the binaries report, writing the release note (tag, version, commit, build source, pre-release mark) to the job summary only after that check.
- REQ-8: A local build shall stay distinguishable from a release build by the second `--version` line (`local build` against `CI build`), with the order in which `moai update` reads a local build and a release pinned by tests: only the numbers `(head, yearweek, build)` decide, so a local build below the installed release needs `--allow-older`, and a local build of the same week with a higher number reads as newer and updates without it.
- REQ-9: The workflow shall hold a read-only token and no secret, with no step creating a tag, a release, an upload or an API call.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): 번호 매기기 — 브랜치 모양, 첫 시도만, run number 와 `HEADVER` 의 범위, 주차 경계(2026-12-31, 2027-01-01, 2027-01-04), 같은 tag 와 낮거나 같은 build 거절, 다른 tag 계열은 막지 않음.
  verify: `python3 -m unittest tests.release_workflow_test.PlanTest`
- AC-2 (REQ-5, REQ-6, REQ-7): tag 모양과 pre-release 표시, 빌드한 `--version` 두 줄·tag·노트가 일치, 어긋나면 노트 없이 실패.
  verify: `python3 -m unittest tests.release_workflow_test.PlanTest.test_a_release_candidate_branch_is_a_pre_release_with_the_same_tag_shape tests.release_workflow_test.VerifyTest`
- AC-3 (REQ-6, REQ-8): 로컬과 릴리스 구별 — workflow 가 `MOAI_LOCAL_BUILD` 를 쓰지 않고, 로컬 빌드 줄을 가진 binary 를 거절하고, 이미 있는 Rust 시험이 두 줄을 고정하고, 새 `update_headver` 시험 둘이 순서(로컬이 더 낮으면 `--allow-older`, 같은 주에 로컬 번호가 더 높으면 더 새 빌드)를 고정.
  verify: `python3 -m unittest tests.release_workflow_test.WorkflowShapeTest.test_the_build_is_numbered_by_the_plan_and_never_a_local_build tests.release_workflow_test.VerifyTest.test_a_local_build_line_fails`; `cargo test --locked --offline --lib headver::`; `cargo test --locked --offline --bin moai version_flag_reports_the_headver_build`; `cargo test --locked --offline --test update_headver`
- AC-4 (REQ-1, REQ-9): 게시하지 않음 — `push` 만, 읽기 전용 token, secret·tag·release·업로드·API 호출 없음, 값은 `env` 로만 스크립트에 들어감.
  verify: `python3 -m unittest tests.release_workflow_test.WorkflowShapeTest`
- AC-5 (REQ-1…REQ-9): 이웃 시험과 범위.
  verify: `python3 -m unittest tests.release_workflow_test tests.github_assets_test tests.ci_scope_test`; `git diff --name-only origin/v2026.11.0...HEAD` lists only the Scope paths

## Scope

- write: `.github/workflows/release.yml`, `tests/release_workflow_test.py`, `tests/update_headver.rs` (two new tests only, added by the leader's decision note 2260), `docs/release-attestation.md`, `.moai/specs/SPEC-20261119-headver-release-workflow/spec.md`; `docs/release-checklist.md` (released by the leader after C-67 ended, 2026-11-19, the HeadVer section only), and, once C-76's pull request has merged, `tests/github_assets_test.py` (the exact list of workflow files in `WorkflowTest.workflows()`)
- read-only: `build.rs`, `src/headver.rs`, `src/local_build.rs`, `.github/workflows/ci.yml`, `.github/workflows/candidate.yml`, `skills/moai/references/internal/versioning.md`, `docs/release-tag-ruleset-moai-cli-v.md`, `AGENTS.local.md`
- forbidden: any tag push or release, the tag ruleset, `src/**`, `install/**`, `Cargo.toml` `version`, `HEADVER`, the real HOME, the network

## Plan

1. 시험 먼저: `tests/release_workflow_test.py` — workflow 모양, `plan` 스크립트(임시 git 저장소와 tag), `verify` 스크립트(가짜 binary). `release.yml` 이 없어서 전부 실패하는 것(RED)을 봐요.
2. `release.yml`: `plan` job(브랜치 모양, 첫 시도, 번호, tag 검사, 한 번 읽은 시계) → `build` job(`MOAI_BUILD_NUMBER`·`SOURCE_DATE_EPOCH` 로 빌드, `verify` 로 두 binary 의 `--version` 과 tag 비교, 노트).
3. `docs/release-attestation.md` 에 사람이 따르는 순서(7절)를 더해요. 시험은 이 문서를 읽지 않아요(`ci_scope` 거부 목록이 이 카드 밖이라서 `rg` 확인 명령으로 검증해요).
4. 이웃 시험, 실제 빌드로 `verify` 를 한 번 돌려 보기, mutation.

## Risks

- 설치 스크립트(`install/install.sh` 의 `TAG_PREFIX`, `install/install.ps1` 의 `$TagPrefix`)는 릴리스를 찾을 때와 내려받을 때 `moai-cli-v` 접두사만 써서(`--version` 으로 고정해도), D6 의 `moai-adk-v*` tag 로 게시한 릴리스를 찾지도 내려받지도 못해요. `install/**` 는 이 SPEC 이 쓰지 않아요 → 첫 게시 전에 사용자가 접두사를 맞추는 방법(설치 스크립트가 `moai-adk-v` 를 읽게 하기, 또는 tag 를 `moai-cli-v` 로 되돌리기)과 tag ruleset 을 함께 골라야 해요. 체크리스트 (a-2)와 `docs/release-attestation.md` 7.3 에 적었어요.
- `run_number` 는 workflow 별 계수기라 `release.yml` 을 지웠다 다시 만들면 1 부터 다시 시작해요. REQ-4 가 "이미 있는 tag 의 build 보다 커야 한다"로 막아요. tag 를 지우면(ruleset 이 아직 없어요) 이 보호도 약해져요 → 사용자의 tag ruleset 결정이 따라와야 해요.
- plan 이 읽는 tag 는 checkout 이 가져온 것이에요. plan 과 사람이 tag 를 올리는 사이에 다른 사람이 같은 tag 를 올리면 이 workflow 는 모르니, 사람의 순서(7절)가 올리기 직전에 원격 tag 를 다시 읽게 해요.
- 한 번의 Linux 빌드만 확인해요. 버전은 플랫폼과 무관하게 `build.rs` 가 새기지만, 산출물 자체는 이 workflow 가 만들지 않아요(Non-goals).
- GitHub 의 `workflow_dispatch` 는 기본 브랜치에 파일이 있어야 하고 이 저장소의 기본 브랜치(`main`)에는 workflow 가 없어서 쓰지 않아요(`candidate.yml` 머리말이 2026-11-16 에 읽은 문서를 적어요; 이 카드에서 문서를 다시 읽지는 않았어요).
- 한계(D-6): 로컬 빌드 번호와 릴리스 번호는 같은 주에 숫자만으로 줄 세워져요. 릴리스 계수기는 1 부터, 로컬 계수기는 지금 6 이라서 같은 주에 릴리스 `4.2648.1` 이 설치된 호스트의 로컬 빌드 `4.2648.7` 은 더 새 빌드로 읽혀 `--allow-older` 없이 update 돼요. 설치 기록에 종류가 없어서 `update` 는 둘을 구별하지 못해요. 사용자의 선택(로컬은 작은 계수기 유지)은 그대로예요.

## Decisions

| id | decision | why | alternative |
| --- | --- | --- | --- |
| D-1 | workflow 는 tag 를 올리거나 게시하지 않아요; 사람이 job 요약의 노트로 올려요 | 되돌릴 수 없는 외부 기록이라 사용자가 매번 확인(case 2)하고, 게시 환경(승인자)과 `moai-adk-v*` ruleset 이 아직 없어서 게시 job 은 확인 없이 돌 수 있어요 | 환경 승인자와 ruleset 을 사용자가 정한 뒤 게시 job 을 후속 카드로 |
| D-2 | `release/<sha>`·`release/rc/<sha>` 브랜치 push 가 시작 조건 | `workflow_dispatch` 는 기본 브랜치의 파일이 필요해요; `candidate.yml` 과 같은 방식 | 기본 브랜치에 workflow 를 두기(저장소 보호 설정 결정) |
| D-3 | pre-release 표시는 브랜치 모양(`release/rc/`)이 정해요 | 사람이 push 하는 이름이라 워크플로 입력이 없어도 매번 사람의 선택이에요; tag 에는 접미사가 없어요(D6) | 별도 입력(dispatch 가 필요) |
| D-4 | `{build}` = 첫 시도의 `run_number`, 기존 tag 의 build 보다 커야 함 | D5 + `versioning.md`(재실행은 `run_number` 를 유지) | 저장소 변수 계수기(D5 에서 고르지 않음) |
| D-5 | 플랫폼 하나(Linux x86_64)로 빌드해 버전만 확인 | 버전 새기기는 플랫폼과 무관, 산출물은 이 SPEC 밖 | 세 플랫폼 빌드(산출물 카드) |
| D-6 | 로컬 빌드와 릴리스의 update 순서는 지금처럼 숫자만(선택 A), 사실을 `tests/update_headver.rs` 의 시험 둘로 고정하고 한계를 문서와 Risks 에 적어요 | `compare`(`src/headver.rs`)는 `(head, yearweek, build)` 만 봐요; 리더 결정 note 2260(질문 note 2258, 분류 none, 되돌리기 가능; 리더 정정: D4 의 "로컬 빌드는 --allow-older 로만" 설명은 로컬 번호가 더 낮을 때만 참) | B: 로컬 빌드를 같은 주 릴리스보다 낮게 읽도록 `compare`·설치 기록에 종류를 더하기(src 변경, 새 카드로 사용자에게) |

## Result

구현 에이전트가 채웠어요. 감사가 다시 확인해요. 기준 tip `b04e815c39`, macOS arm64, 이 worktree, 빌드 폴더 `~/.moai/build/moai-adk-93`, `CARGO_INCREMENTAL=0`, cargo cache `/tmp/moai-rc-cargo-agent-3`. 측정 시각 2026-11-18T22:32Z 까지.

RED(시각은 재지 않았어요): `release.yml` 이 없는 채로 `python3 -m unittest tests.release_workflow_test` 는 27개 시험이 전부 오류(`FileNotFoundError`). 통과는 `release.yml` 을 쓴 뒤 27개 OK. `update_headver` 의 새 시험 둘은 현재 동작을 고정하는 시험이라 처음부터 통과해요(D-6 은 코드를 바꾸지 않아요).

- AC-1: met — `python3 -m unittest tests.release_workflow_test.PlanTest` 13개 OK(브랜치 모양 11가지, 재실행 6가지, run number 경계 13가지, `HEADVER` 경계 12가지, 주차 경계 6일, 같은 tag·같거나 낮은 build 거절 4개, 다른 tag 계열 무시). mutation 6개(재실행 허용, 같은 build 허용, commit 확인 삭제, 같은 tag 허용, build 증가 검사 삭제, `HEADVER` 앞자리 0 허용)를 모두 시험이 잡았어요.
- AC-2: met — `PlanTest.test_a_release_candidate_branch_is_a_pre_release_with_the_same_tag_shape` 와 `VerifyTest` 6개 OK. 실제 빌드 `MOAI_BUILD_NUMBER=7 SOURCE_DATE_EPOCH=1795521600 cargo build --release --locked --offline`(exit 0)의 `moai-cli`·`moai` 가 `moai-cli 4.2648.7 (rust)` / `CI build 7 (MOAI_BUILD_NUMBER), built 2026-11-24 UTC (ISO week 2026-W48)` 이고, 그 binary 로 `verify` 스크립트가 exit 0 과 노트를 내고 tag 번호를 틀리게 주면 exit 1(`expected 'moai-cli 4.2648.8 (rust)'`). mutation 3개(rc 표시 빼기, 둘째 줄 검사 삭제, tag 계열 바꾸기)를 잡았어요.
- AC-3: met — `WorkflowShapeTest.test_the_build_is_numbered_by_the_plan_and_never_a_local_build` 와 `VerifyTest.test_a_local_build_line_fails` OK(`local build`·`unnumbered build`·다른 번호·꼬리 글자·빈 줄 거절); `cargo test --locked --offline --lib headver::` 11 passed, `--bin moai version_flag_reports_the_headver_build` 1 passed(번호 없는 빌드에서). `cargo test --locked --offline --test update_headver` 는 `MOAI_LOCAL_BUILD=7` 와 `MOAI_BUILD_NUMBER=7` 로 각각 5 passed(새 시험 둘 포함, `skipped` 줄 0개, 번호 없는 빌드에서는 첫 시험이 건너뛰므로 번호를 준 빌드로 확인). mutation: `MOAI_BUILD_NUMBER` 를 `MOAI_LOCAL_BUILD` 로 바꾸기와 plan 의 시계를 넘기지 않기를 잡았어요. 숫자 순서는 바꾸지 않았고 한계는 Risks 와 `docs/release-attestation.md` 7.3 에 있어요(D-6). `src/headver.rs` 의 `compare` 를 일부러 깨는 mutation 은 이 카드의 쓰기 범위 밖이라 돌리지 않았어요(not_run).
- AC-4: met — `WorkflowShapeTest` 8개 OK(`push` 만, 읽기 전용 token, secret·id-token 없음, 금지어 `gh release|git tag|git push|upload-artifact|attest|gh api …` 없음, `${{` 가 `run:` 에 없음, `concurrency` 없음, 모든 `uses` 가 ci.yml 의 같은 commit). mutation 3개(`pull_request` 트리거 추가, 쓰기 token, `gh release create` 단계)를 잡았어요. `actionlint .github/workflows/release.yml` exit 0.
- AC-5: partly — `python3 -m unittest tests.release_workflow_test tests.ci_scope_test` 46 OK; `cargo test --locked --offline --test release_suite` 4 passed(문서 7절 추가 뒤); `git diff --check` exit 0; `rg -c` 로 문서 7절 세 문구 3. **`python3 -m unittest tests.github_assets_test` 는 지금 21개 실패**(`WorkflowTest.workflows()` 의 workflow 파일 목록 단언이 `release.yml` 을 모름, 21개가 모두 `Lists differ … 'release.yml'`). 목록에 `release.yml` 한 줄만 더한 메모리 안 사본으로는 39개 모두 OK 였어요. 그 파일은 아직 이 카드가 쓰지 않아요: 리더 답(2026-11-19)이 C-76 의 PR 이 병합된 뒤에 넘긴다고 했고, 그때까지 PR 을 열지 않아요(not_run). `docs/release-checklist.md` 는 같은 날 리더가 C-67 을 풀어 줘서 이 카드가 고쳤어요(아래 줄). `shellcheck` 는 로컬에 없어 CI 결과로 확인해야 하고, 실제 GitHub 실행(push 한 release 브랜치)은 하지 않았어요(not_run).
- AC-5 체크리스트: met — `docs/release-checklist.md` (a-2)의 옛 "not_run — 첫 HeadVer 릴리스의 번호와 tag" 줄을 규칙(D5·D6, `release.yml`)과 "실제 실행·첫 tag 는 not_run" 두 줄로 바꾸고, tag 접두사 줄에 D6 의 `moai-adk-v` 와 설치 스크립트 접두사 문제를 적었어요. `tests/install_script_test.py` 가 고정한 문구는 그대로라 `python3 -m unittest tests.install_script_test.HeadVerDocsTest tests.install_script_test.HeadVerTextTest` 2개 OK, `tests.support_matrix_test tests.release_workflow_test tests.ci_scope_test` 51개 OK, `git diff --check` exit 0, `moai spec check` 오류 0·경고 0(예전에 걸리던 req_gears 7건을 REQ 하나에 `shall` 하나로 고침). **새로 알게 된 사실**: `install/install.sh` 와 `install/install.ps1` 은 릴리스 찾기와 내려받기에 `moai-cli-v` 접두사만 써서(`grep -n TAG_PREFIX install/install.sh` 의 279·348행), `moai-adk-v*` 로 게시하면 설치 스크립트가 찾지 못해요. `install/**` 는 이 SPEC 의 범위 밖이라 고치지 않았고 Risks·7.3·체크리스트에 적었어요(사용자 선택, 새 카드).
