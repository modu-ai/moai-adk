---
id: SPEC-20261125-headver-weekly-release
title: "HeadVer 규칙 하나: 개발 빌드는 -dev, 릴리스는 배포 월요일의 주와 0부터의 순번"
status: approved
revision: 1
created: 2026-11-25
approval:
  by: user
  at: 2026-11-25
  basis: "사용자가 리더 대화의 선택지에서 직접 고른 답(2026-11-25T05:00:41Z, 카드 노트 8976): \"승인 (Recommended)\"; 리더가 기록 직전에 moai spec show 로 digest 를 다시 읽었어요"
  digest: "sha256:3939a62d296de5533d0060cf0d886c34fd66eb3780aaa702696e2899b27b5f4d"
---

# HeadVer 규칙 하나: 개발 빌드는 -dev, 릴리스는 배포 월요일의 주와 0부터의 순번

## Intent

사용자가 2026-11-25 에 버전 규칙을 정했어요(Decisions D-1…D-4, 말 그대로 인용). 개발 빌드는 `{head}.{yearweek}.{build}-dev` 로 `{build}` 가 계속 올라가고, 릴리스는 `{head}.{배포 월요일의 주}.{0부터의 순번}` 이며 릴리스 브랜치는 `release/v{head}.{yearweek}.{n}` 이에요. 이 규칙은 moai-adk 만이 아니라 HeadVer 를 고른 모든 프로젝트의 제품 기본값이에요.

지금은 셋이 이 규칙과 어긋나요. (1) 빌드가 접미사 없이 `4.2648.26` 처럼 보고해서 개발 빌드와 릴리스를 이름으로 구별하지 못해요(`build.rs`, `src/headver.rs`). (2) `.github/workflows/release.yml` 은 `release/<40hex>` 와 `release/rc/<40hex>` 만 받고 `{build}` 를 실행 번호로 매겨요(`SPEC-20261119-headver-release-workflow` D-2·D-4). (3) 출하 글(`skills/moai/references/internal/versioning.md` 34·38·42·67행, `interview.md` 의 `versioning_head` 행)과 `CLAUDE.md` 의 Flow 8·HeadVer 절은 "빌드 번호는 다시 시작하지 않는다" 와 "릴리스 브랜치는 `release/<40hex>`" 를 말해요.

이 SPEC 은 코드를 바꾸기 전에 남은 선택지(Options)를 사용자가 고르게 해요. 아래 REQ 는 각 선택지의 추천안(첫 번째)을 따른 초안이고, 사용자가 다른 안을 고르면 REQ 와 Plan 을 그에 맞춰 고친 뒤 다시 승인을 받아요.

## Interview

- Q1: 릴리스 브랜치 이름과 번호는 어떻게 정하는가? -> A: `release/v{head}.{yearweek}.{n}`, 배포하는 월요일의 주, 순번은 0 부터예요 (D-1; 사용자: "매주 월요일마다 배포하기 때문에 다음 주번호로 릴리즈를 만들고, 시작은 .0으로 한다.").
- Q2: 개발 빌드와 릴리스의 번호를 함께 세는가? -> A: 따로 세요 (D-2; 사용자: "개발은 4.2648.26-dev, 릴리스는 4.2649.0. 이렇게 가자").
- Q3: 이 규칙은 이 저장소만의 것인가? -> A: 제품 기본값이에요 (D-4; 사용자: "오늘 정하신 버전 규칙(개발 -dev, 릴리스는 주마다 0부터)는 범용적으로 이규칙 사용한다").
- Q4: Options O-1 부터 O-10 은 무엇으로 정했는가? -> A: 각 선택의 추천안이에요 (D-5). 사용자가 그 초안을 담은 이 개정을 승인했어요(카드 노트 8976: "승인 (Recommended)").

## Non-goals

- 실제 `release/**` 푸시, 태그 푸시, 릴리스 게시. 이 카드는 아무것도 원격에 쓰지 않아요.
- `release/**` 강제 푸시 금지 규칙 세트(ruleset)를 만드는 일: 원격 저장소 설정이라 사용자가 확인한 뒤 리더가 해요. 이 SPEC 은 그 규칙이 필요하다는 것과 워크플로 쪽 검사만 정해요.
- 설치 스크립트의 태그 접두사(`moai-cli-v` 와 `moai-adk-v`) 맞추기: `SPEC-20261119-headver-release-workflow` Risks 에 있는 열린 문제 그대로예요.
- SemVer·CalVer 규칙, 그리고 HeadVer 를 기본 선택으로 만드는 일(SemVer 가 기본이에요). 사용자 결정 D-4 는 "HeadVer 를 고른 프로젝트의 규칙" 으로 읽어요(리더 해석, 사용자 확인 전).
- `HEADVER`(head 번호)를 올리는 일과 `Cargo.toml` 의 `version`.

## Requirements

- REQ-1: While a build is not a release build, the build script shall report the version `{head}.{yearweek}.{build}-dev`, with `{build}` taken from `MOAI_LOCAL_BUILD`, from `MOAI_BUILD_NUMBER`, or 0 when neither is set.
- REQ-2: When `MOAI_RELEASE_VERSION` is set, the build script shall report that version without a suffix.
- REQ-3 (REQ-2): IF the release version's head differs from the `HEADVER` head, its yearweek is not four digits with an ISO week from 01 to 53, its serial is not a decimal number from 0 without a leading zero, or `MOAI_LOCAL_BUILD` or `MOAI_BUILD_NUMBER` is also set, THEN the build script shall stop the build.
- REQ-4: When two versions have the same three numbers, the version comparison shall rank the one without `-dev` above the one with it. · failure: none — a comparison of two read versions has no failing step
- REQ-5: When a branch named `release/v{head}.{yearweek}.{n}` is pushed, the release workflow shall compute the expected version from the `HEADVER` head of the pushed commit, the ISO week-year and week of the next Monday after one UTC clock reading (so that Monday 2026-12-28 gives 2653 and Monday 2027-01-04 gives 2701), and the lowest serial that no tag of that head and yearweek uses.
- REQ-6 (REQ-5): IF the branch name does not spell the expected version, or has the shape `release/<40 lowercase hex>` or any shape other than `release/v…` and `release/rc/v…`, THEN the release workflow shall stop before any build.
- REQ-7 (REQ-5): IF the pushed commit is not on the remote `NeXT` branch, THEN the release workflow shall stop before any build.
- REQ-8: When the plan job passes, the build job shall build with `MOAI_RELEASE_VERSION` set to the planned version.
- REQ-9 (REQ-8): IF the first `--version` line of a binary is not `moai-cli <version> (rust)` or `moai <version> (rust)`, the second does not begin `release build <n> (MOAI_RELEASE_VERSION), built`, or the tag is not `moai-adk-v<version>`, THEN the build job shall fail.
- REQ-10: While the workflow holds a read-only token and no secret, it shall create no tag, no release, no upload and no API call.
- REQ-11: While a project's profile names HeadVer, the shipped versioning and interview texts and the generated `Versioning:` line shall state the development form, the release form and the branch name of this SPEC and carry no moai-adk internal.
- REQ-12: While `CLAUDE.md` and the decisions log describe Flow 8 and the HeadVer rule, they shall state the same development form, release form and branch name as this SPEC.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): the build version and the `--version` second line follow REQ-1 and REQ-2 on fixtures and each refusal of REQ-3 is named · verify: `cargo test --locked --offline --lib headver::`
- AC-2 (REQ-4): the comparison puts a release above its `-dev` of the same numbers and keeps the numeric order for the rest, an installed folder named without `-dev` included · verify: `cargo test --locked --offline --test update_headver`
- AC-3 (REQ-5, REQ-6, REQ-7): the workflow's plan script run on temporary git repositories computes the expected name, accepts it, refuses a wrong head, week or serial, a pushed commit outside `NeXT` and the old shapes, and agrees with `headver::yearweek` on a table of boundary dates · verify: `cargo test --locked --offline --test card release_branch_version::`
- AC-4 (REQ-8, REQ-9, REQ-10): the workflow builds with `MOAI_RELEASE_VERSION`, checks the two `--version` lines and the tag, and holds no write permission, secret, tag, release or upload step · verify: `cargo test --locked --offline --test card -- release_branch_version::the_build_job_checks_both_lines_and_the_tag release_branch_version::the_workflow_publishes_nothing`
- AC-5 (REQ-11): the shipped texts and the generated line show the new rule and no moai-adk internal · verify: `cargo test --locked --offline --test versioning_cli`
- AC-6 (REQ-12): `CLAUDE.md` Flow 8 and the HeadVer section and the decisions log say the same rule · verify: `rg -c "release/v\{head\}\.\{yearweek\}\.\{n\}" CLAUDE.md docs/decisions-2026-11-08.md`
- AC-7 (REQ-1…REQ-12): each new guard is killed by a mutant, and the neighbour tests pass · verify: `moai check affected --card NeXT..HEAD`

## Scope

- write (this card now): `.moai/specs/SPEC-20261125-headver-weekly-release/**`
- write (after the user approves this SPEC, each by a scope request to the leader): `src/headver.rs`, `src/local_build.rs`, `build.rs`, `src/main.rs` (the `version_flag_reports_the_headver_build` test only), `src/versioning.rs` (the generated `Versioning:` line and its tests), `.github/workflows/release.yml`, `tests/card/release_branch_version.rs` (new; its first word `release` puts it in the moved `card` area crate) and the `mod release_branch_version;` line of `tests/card/main.rs`, `tests/bin_names_signature_cli.rs` (its release check runs the workflow's verify script), `tests/update_headver.rs`, `tests/versioning_cli.rs`, `tests/release_workflow_test.py` (its old-shape pins move into the new Rust test; the file is removed only if the leader agrees), `skills/moai/references/internal/versioning.md`, `skills/moai/references/internal/interview.md` (the `versioning_head` row: after the card that edits the interview file has landed, or by line lease), `docs/release-checklist.md`, `docs/release-attestation.md`, `docs/decisions-2026-11-08.md`
- write (tier 2, a card whose scope names it, with the user's SPEC approval): `CLAUDE.md` (Flow 8 and the HeadVer section), the workflow file above
- read-only: `HEADVER`, `install/**`, `.github/workflows/ci.yml` (it keeps `MOAI_BUILD_NUMBER`; the `-dev` suffix comes from the build script, so it needs no change), `tests/install_script_test.py`
- forbidden: any push, tag or release, the remote ruleset, `Cargo.toml` `version`, the real HOME, the network

## Plan

1. 사용자가 선택지를 고르면(아래 Options) REQ 를 고치고 `status` 를 사용자가 승인해요(레인은 승인을 기록하지 않아요).
2. 시험 먼저: `headver::` 단위 시험(`-dev`, `MOAI_RELEASE_VERSION`, 거절 사유), `update_headver`(같은 숫자에서 릴리스가 위, 접미사 없는 옛 설치 폴더), `release_branch_version`(워크플로의 plan 스크립트를 임시 git 저장소와 가짜 시계로 돌리고 경계 날짜 표를 `headver::yearweek` 와 맞춰요). 모두 RED 를 봐요.
3. `src/headver.rs`·`src/local_build.rs`·`build.rs`: `-dev` 접미사와 `MOAI_RELEASE_VERSION` 입력, `--version` 둘째 줄 `release build <n> (MOAI_RELEASE_VERSION)`, 비교의 같은 숫자 규칙.
4. `release.yml`: plan job 이 브랜치 이름의 기대 버전을 계산(시계는 한 번 읽어요), 푸시 커밋이 원격 `NeXT` 에 있는지 확인, build job 이 `MOAI_RELEASE_VERSION` 으로 빌드하고 두 줄과 태그를 확인해요. 게시하지 않아요(REQ-8).
5. 출하 글과 생성 줄(`versioning.md`, `interview.md`, `src/versioning.rs`), 문서(`release-checklist`, `release-attestation`), 결정 로그, `CLAUDE.md` 를 새 규칙으로 맞춰요.
6. 뮤턴트(REQ 마다 한 가드를 지워 시험이 죽는지), `moai check affected --card` 한 번.

## Options

선택마다 첫째가 추천이고 기본값이에요. 사용자가 고른 것은 Decisions 에 기록해요.

| Id | Question | Options (first = recommended = default) | Cost of the others | Reversible |
| --- | --- | --- | --- | --- |
| O-1 | 릴리스 후보(pre-release) 이름 | (a) `release/rc/v{head}.{yearweek}.{n}`: 같은 번호 공간, 태그에는 접미사 없이 GitHub pre-release 표시(`SPEC-20261119` D-3 계승). (b) 태그에 `-rc.N` 접미사 | (b) 는 `install.sh` 가 `-` 를 옛 스킴의 후보로 읽어 안정 채널에서 건너뛰는 규칙과 부딪혀요 | 예 |
| O-2 | 연말 주 경계 | (a) 배포 월요일의 날짜로 ISO 주-연도를 계산(2026-12-28 → 2653, 2027-01-04 → 2701). (b) 만드는 주 번호에 1 을 더하고 52·53 다음을 01 로 | (b) 는 해마다 52주인지 53주인지 따로 알아야 해요 | 예 |
| O-3 | 만드는 날이 월요일일 때 | (a) 항상 만든 날 다음 ISO 주의 월요일(지시의 "다음 주번호" 그대로: 월요일에 만들어도 다음 주). (b) 월요일이면 그날을 배포일로 | (b) 는 같은 월요일에 만든 릴리스의 주 번호가 두 가지 읽혀요 | 예 |
| O-4 | `-dev` 가 보이는 자리 | (a) 빌드 버전 자체: `--version` 첫 줄, 상태줄, 설치 기록, 설치 폴더 `versions/<버전>`, 영수증 모두. (b) `--version` 둘째 줄과 노트에만 | (b) 는 폴더 이름과 상태줄만 보고는 개발 빌드를 못 알아봐요 | 예 |
| O-5 | 같은 숫자의 개발 빌드와 릴리스, 접미사 없는 옛 설치 | (a) 숫자 순서 그대로, 같은 숫자일 때만 릴리스가 `-dev` 위. 옛 폴더(`4.2648.26`)는 접미사 없는 릴리스 모양으로 읽어요. (b) 같은 head·yearweek 에서는 릴리스가 항상 `-dev` 위. (c) 접미사를 계속 무시 | (b) 는 릴리스 뒤에 만든 더 새 개발 빌드가 더 낮게 읽혀요. (c) 는 같은 숫자의 둘을 구별 못 해요 | 예 |
| O-6 | 태그 꼴 | (a) `moai-adk-v{head}.{yearweek}.{n}` 그대로(브랜치 이름에서 `release/v` 를 뺀 것과 같아요). (b) `v{head}.{yearweek}.{n}` | (b) 는 태그 규칙 세트(`v*`, `moai-cli-v*`)와 설치 스크립트 접두사를 다시 정해야 해요 | 예 |
| O-7 | 배포 요일을 프로젝트마다 고르게 할지 | (a) 월요일로 고정(지시 그대로, 프로필 키를 더하지 않아요). (b) 프로필 `versioning` 에 배포 요일 키를 더해 월요일을 기본값으로 | (b) 는 프로필 스키마·검사·질문 행이 늘어요(필요해질 때 새 SPEC 으로 할 수 있어요) | 예 |
| O-8 | 릴리스 버전이 빌드로 들어가는 길 | (a) 새 환경 변수 `MOAI_RELEASE_VERSION`(`build.rs` 가 `HEADVER` 와 모양을 검사하고 둘째 줄을 `release build <n>` 으로). (b) `MOAI_BUILD_NUMBER` 에 0 을 허용하고 `SOURCE_DATE_EPOCH` 를 배포 월요일 0시로 | (b) 는 "built" 날짜가 실제 빌드 시각이 아니게 되고 `ci.yml` 의 실행 번호 빌드와 구별이 흐려요 | 예 |
| O-9 | 워크플로의 기대 버전 계산을 어디에 둘지 | (a) 워크플로의 bash plan 스크립트에 두고 Rust 시험이 그 스크립트를 임시 저장소에서 돌려 `headver::yearweek` 와 경계 날짜 표로 맞춰요(지금 방식 그대로, 새 명령 없음). (b) `moai` 하위 명령으로 만들어 워크플로가 그것을 빌드해 불러요 | (b) 는 plan job 이 먼저 빌드해야 해서 느리고 표면이 늘어요 | 예 |
| O-10 | 생성되는 `Versioning:` 줄(`src/versioning.rs`)과 옛 Python 워크플로 시험 | (a) 줄을 새 규칙으로 바꾸고, `tests/release_workflow_test.py` 의 옛 모양 단언은 Rust 시험으로 옮겨요. (b) 줄은 그대로 두고 출하 글만 고쳐요 | (b) 는 "생성되는 설정이 같은 규칙을 말한다"는 사용자 결정 D-4 와 맞지 않아요 | 예 |

## Risks

- 같은 주 안에서 릴리스 순번(0부터)은 개발 `{build}`(수십)보다 항상 작아서, 배포 월요일 이후 만든 개발 빌드는 O-5 (a) 에서 같은 주 릴리스보다 새로 읽혀요. 개발자 PC 에서 개발 빌드가 설치된 채 같은 주에 두 번째 릴리스가 나오면 `moai update` 가 `--allow-older` 를 요구해요. 사용자 배포는 릴리스만 설치하므로 영향이 없어요(옛 `SPEC-20261119` D-6 의 한계와 같은 종류예요). 시험으로 고정해요.
- `moai-adk-v*` 태그를 쓰는데 `install/install.sh` 는 `moai-cli-v` 접두사만 읽어요(옛 SPEC 의 열린 문제). 이 SPEC 은 고치지 않으므로 첫 게시 전에 사용자가 따로 골라야 해요.
- REQ-7 의 "원격 `NeXT` 에 있다"는 `NeXT` 가 원격에 올라가 있어야 성립해요. 지금 로컬 `NeXT` 가 원격보다 앞서 있어서(`remote_copy` 꺼짐) 첫 릴리스 전에 사람이 먼저 `NeXT` 를 올려야 해요: 원격 쓰기라 이 카드는 하지 않고, 워크플로는 원격에 없으면 멈춰요.
- `release/**` 강제 푸시 금지 규칙 세트는 아직 없어요. 만들기 전에는 같은 이름을 다른 커밋으로 덮어쓰는 것을 워크플로의 REQ-6(이미 쓰인 순번이면 기대 순번이 올라가 이름이 어긋나요)와 REQ-7 만 막아요.
- `-dev` 접미사는 이 저장소의 `compare` 가 이미 무시해서(`src/headver.rs` 의 `shape`) 상태줄·update 비교는 그대로 읽지만, 접미사 있는 폴더 이름이 새로 생겨요: 설치 폴더·영수증을 읽는 시험을 마지막 단계의 `moai check affected` 가 잡아요.
- `release.yml` 의 plan 스크립트는 지금 `python3 -c` 로 ISO 주를 계산해요(CI 러너 안). 새 plan 은 GNU `date` 로 계산해 Python 을 쓰지 않아요.
- 사용자 결정 D-4 는 리더가 "HeadVer 를 고른 프로젝트의 규칙" 으로 해석했어요(사용자 확인 전). 버전 체계의 기본 선택은 바뀌지 않아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 릴리스 브랜치는 `release/v{head}.{yearweek}.{n}`, `{yearweek}` 는 배포하는 월요일이 속한 ISO 주(만드는 시점의 다음 주), `{n}` 은 그 주의 릴리스 순번으로 0 에서 시작하고 같은 주 재배포는 1, 2 | 사용자 지시 2026-11-25T00:01:46Z 그리고 같은 날 두 번째 답(리더 대화, 이 SPEC 의 카드 본문): "release/v4.2648.0 형태로 다음 년주 릴리즈 버전으로 릴리즈 하도록 한다. 시작은 .0 으로 한다." 와 "매주 월요일마다 배포하기 때문에 다음 주번호로 릴리즈를 만들고, 시작은 .0으로 한다." | 새 사용자 결정으로만 |
| D-2 | 번호를 따로 세요: 개발 빌드는 계속 올라가는 `{build}` 에 `-dev`(예 `4.2648.26-dev`), 릴리스는 주마다 0 에서 시작하는 순번(예 `4.2649.0`) | 사용자 결정 2026-11-25T00:07:24Z(카드 note 8830): "개발은 4.2648.26-dev, 릴리스는 4.2649.0. 이렇게 가자" | 새 사용자 결정으로만 |
| D-3 | 안전장치 둘: (a) 릴리스 워크플로가 푸시된 커밋이 원격 `NeXT` 의 착지 커밋인지 확인하고 아니면 실패, (b) `release/**` 에 강제 푸시 금지 규칙(같은 이름을 다른 커밋으로 덮어쓰지 않고 순번을 올려 새로 내요) | note 8830 의 리더 제안에 사용자가 방향으로 동의. (a) 는 REQ-7, (b) 의 원격 규칙 세트 생성은 사용자 확인 뒤 리더가 해요(Non-goals) | (a) 는 REQ-7 삭제, (b) 는 규칙 세트 삭제 |
| D-4 | 이 규칙은 moai-adk 만이 아니라 제품 기본값: HeadVer 를 고른 프로젝트의 출하 글과 생성되는 설정이 같은 규칙을 말해요 | 사용자 결정 2026-11-25T00:49:09Z(note 8840): "오늘 정하신 버전 규칙(개발 -dev, 릴리스는 주마다 0부터)는 범용적으로 이규칙 사용한다" | 새 사용자 결정으로만 |
| D-5 | 릴리스 후보·연말 경계·`-dev` 위치·비교 규칙·태그 꼴·배포 요일·빌드 입력·계산 위치·생성 줄은 위 Options 의 추천안을 초안으로 삼아요 | 사용자가 고르기 전까지의 기본값이에요(레인이 정하는 것이 아니에요). 승인 기록은 사용자만 남겨요 | Options 에서 다른 안을 고르면 REQ·Plan 을 고쳐요 |

## Result

- AC-1 met: `cargo test --lib headver::` 15개 통과(`a_development_build_ends_in_dev`, `a_release_build_reports_its_version_without_a_suffix`, `a_wrong_release_version_stops_the_build`). 종단 시험(`src/main.rs` 의 버전 플래그 시험)도 `-dev` 와 `MOAI_RELEASE_VERSION` 두 모양을 읽어요.
- AC-2 met: `headver::tests::a_release_ranks_above_the_dev_build_of_the_same_numbers` 와 `cargo test --test update_headver` 6개(새 `an_installed_release_of_the_same_numbers_reads_newer_than_this_dev_build` 포함).
- AC-3 met: `cargo test --test card release_branch_version::` 의 plan 스크립트 시험 12개(이름·주·순번·연말 표·`NeXT` 조상·옛 모양 거절·잘못된 tag 거절). 주 계산 표는 `headver::yearweek` 와 맞춰요.
- AC-4 met: 같은 파일의 `the_build_job_checks_both_lines_and_the_tag`, `the_workflow_publishes_nothing`, `the_jobs_keep_their_settings`, 그리고 `tests/bin_names_signature_cli.rs` 3개. 옛 Python 시험의 단언은 모두 이 Rust 시험으로 옮겼고(대조는 그 시험 삭제를 감사한 노트에 있어요), Python 은 쓰지도 돌리지도 않았어요.
- AC-5 met: `cargo test --test versioning_cli` 15개와 `cargo test --lib versioning::` 22개, 출하 글 `skills/moai/references/internal/versioning.md` 와 소스 매니페스트 해시.
- AC-6 not_run: `CLAUDE.md` 와 `docs/decisions-2026-11-08.md` 는 다른 카드가 쥐고 있어 이 카드의 범위가 아니에요. 그 카드가 착지한 뒤의 조각으로 같은 규칙을 적어요(interview.md 는 그 파일의 카드가 착지한 뒤).
- AC-7 met: 가드 뮤턴트 12개를 한 슬롯에서 돌려 12개 모두 이름 지은 시험이 잡았어요(`KILLED`, 트리는 깨끗). 이웃 시험과 `moai check affected --card` 영수증은 READY 노트에 인용해요.
- 알려진 한계: `build.rs` 가 `MOAI_RELEASE_VERSION` 을 읽는 배선은 릴리스 workflow 의 실제 실행에서만 끝까지 지나가요(not_run; 계산은 `headver::release_build` 단위 시험이 덮어요). 같은 주의 개발 빌드가 릴리스보다 새로 읽히는 한계는 Risks 대로예요. `install/install.sh` 의 tag 접두사(`moai-cli-v`)는 바꾸지 않았어요. `src/lib.rs` 의 주석과 `docs/local-rc1-kit.md`, `skills/moai/workflows/sync.md` 에는 옛 규칙의 글이 남아 있고 이 카드의 범위 밖이에요.
