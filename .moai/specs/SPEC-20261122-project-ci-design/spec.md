---
id: SPEC-20261122-project-ci-design
title: "The project setup asks a few plain questions and plans CI files that are written only on the user's pick"
status: draft
revision: 2
created: 2026-11-22
approval: {by: user, at: 2026-11-22, basis: "AskUserQuestion selection r2: 승인 (Recommended)", digest: "sha256:9ff103df18e59a04f1daa214ed26b6f7aa789faf99fcf1bfeb407b257513e961"}
parent: SPEC-20261121-flow-skeleton
---

# The project setup asks a few plain questions and plans CI files that are written only on the user's pick

## Intent

`/moai project` 를 쓰는 사람 중에는 CI(코드를 올릴 때마다 자동으로 도는 검사)를 모르는 비개발자가 많아요. 그래서 설정 단계가 쉬운 질문 여섯 개 안팎으로 그 프로젝트에 맞는 CI 를 설계하고, 파일은 사용자가 고른 뒤에만 `.github/` 에 쓰도록 정해요. 사용자 제안(2026-11-22, 리더가 채택). 일반 제품 설계예요: 이 저장소 자신의 `.github` 는 건드리지 않아요. 이 조각은 조사와 SPEC 만 쓰고 코드는 뒤 조각이 이어요.

이 SPEC 은 **무엇을 묻고, 어떻게 계획하고, 언제 쓰는가** 를 정해요. 만들어지는 워크플로가 **무엇을 담는가**(트리거, 비용, 안전 규칙)는 자식 SPEC `SPEC-20261122-project-ci-workflows` 가 정해요. 한 SPEC 이 요구사항 열두 개를 넘지 않게 나눈 거예요.

워크플로 파일은 저장소의 가드 설정이므로 미리보기와 사용자의 선택 없이는 한 줄도 쓰지 않아요. 코드에는 이미 "계획 → 다이제스트 → 사용자 선택 → 적용, 기준선과 3-way 로 사용자 수정은 덮지 않음" 틀이 있어요(`github-plan`/`github-apply`). 프로필에는 프로젝트의 검사 명령을 담는 자리가 아직 없어요.

### Research

공식 문서를 통째로 읽은 기록이에요(읽은 시각은 UTC, 2026-11-22, 에이전트 4명이 같은 날 읽음). 원문은 모두 docs.github.com 이에요. 두 SPEC 이 같이 써요.

| 주제 | 읽은 페이지 | 설계에 쓴 사실 |
| --- | --- | --- |
| 권한 | workflow syntax, GITHUB_TOKEN 튜토리얼과 개념, 저장소·조직 설정 (00:53–00:54Z) | `permissions:` 에 하나라도 적으면 나머지는 모두 `none`. 새 저장소 기본 토큰은 contents·packages 읽기 전용이지만 설정이 허용형인 저장소도 있어서 생성 파일은 최상위에 `permissions: contents: read` 를 명시해요. |
| 동시 실행 | workflow syntax, control-workflow-concurrency (00:53Z) | 같은 그룹에 새 실행이 오면 대기 중인 실행을 대체해요. 돌고 있는 실행까지 끊으려면 `cancel-in-progress: true`. 워크플로 수준 그룹 식은 `github`·`inputs`·`vars` 만 쓰고, `github.head_ref` 는 PR 에만 있어 대체값이 필요해요. |
| 필터와 필수 검사 | workflow syntax, events (00:53Z) | 경로·브랜치 필터로 건너뛴 워크플로의 검사는 "Pending" 으로 남아 그 검사를 필수로 둔 PR 의 병합을 막아요. `branches`/`branches-ignore`, `paths`/`paths-ignore` 는 각각 함께 못 써요. |
| 예약 실행 | events (00:53Z) | UTC 기준, 최소 5분 간격, 기본 브랜치에서만, 매시 정각에 지연·누락될 수 있고, 공개 저장소는 활동이 60일 없으면 자동으로 꺼져요. |
| 포크 PR | events, 저장소 설정, 포크 승인 (00:53–00:54Z) | 포크 PR 은 `GITHUB_TOKEN` 외 시크릿이 없고 토큰은 읽기 전용이며 첫 기여자는 승인이 필요할 수 있어요. `pull_request_target` 은 기본 브랜치 맥락에서 쓰기 토큰으로 돌아 PR 코드를 검사하는 용도로 쓰면 위험해요. |
| 한도 | limits, workflow syntax (00:53Z) | 작업 `timeout-minutes` 기본 360, 호스티드 최대 6시간, 매트릭스 256개, `fail-fast` 기본 true, 작업 `if` 는 `matrix` 를 못 읽고 `runner.os` 는 `runs-on` 에 못 써요. |
| 비용 | billing/github-actions, 러너 가격표, hosted-runners, limits (00:53Z) | 포함량은 분이 아니라 금액(Free 12달러, Pro·Team 18달러, Enterprise 300달러, 문서가 "변경될 수 있다" 고 적음). 단가는 분당 Linux 2코어 0.006, Windows 0.010, macOS 0.062 달러이고 작업마다 올림. 공개 저장소의 표준 러너는 무료·무제한이에요. 예산은 웹 설정에서 사용자가 직접 만들고 "Stop usage when budget limit is reached" 를 켜야 멈춰요. 기본 예산값과 예산용 `gh` 명령은 문서에 없어요. |
| 캐시 | dependency-caching, manage-caches (00:53–00:54Z) | 저장소당 10GB, 7일 동안 안 쓰면 삭제. PR 이 만든 캐시는 그 PR 에서만 다시 읽혀요. 기본 브랜치에서 `push`·`schedule` 등으로 만든 캐시는 다른 브랜치가 읽어요. |
| 보안 | secure-use, script-injections, securely-using-pull_request_target, 시크릿, 환경, Dependabot (00:53–00:55Z) | 완전한 커밋 SHA 로 고정이 기본 권고이고(락파일은 더 강하지만 아직 일부 저장소 기능), `${{ }}` 를 `run:` 에 직접 쓰지 말고 `env:` 를 거쳐요. 시크릿은 이름만 다루고(`gh secret set NAME`) 구조화된 값을 시크릿으로 쓰지 않아요. 배포 승인자는 환경 설정이고 Free·Pro·Team 은 공개 저장소에서만 쓸 수 있어요. `dependabot.yml` 의 `github-actions` 생태계가 고정된 SHA 를 갱신해요(같은 줄의 버전 주석도 갱신). |

문서에서 확인하지 못한 것(설계가 단정하지 않아요): `cancel-in-progress: false` 의 동작을 풀어 쓴 문장, `pull_request_target` 에서 시크릿 사용 가능 여부, `ISSUE_TEMPLATE`·`dependabot.yml`·`CODEOWNERS` 를 GitHub 가 읽는다는 설명, 날짜가 붙은 플랫폼 요금, 호스티드 러너의 대기 시간 한도. 보존 정책 페이지의 "Starting October 1, 2026" 는 지난 날짜인데도 미래형이라 오래된 문구로 보고 쓰지 않아요.

코드와 이 저장소 읽은 기록(기준 `012e9470cb`): `src/github_templates.rs`(528줄, 계획·적용·기준선·3-way), 얇은 래퍼 둘, `tools/local_check.sh`(468줄), `tests/github_cli.rs`·`tests/github_profile_cli.rs`, `src/profile.rs`(1992줄), `src/project_update.rs`(1817줄), `src/project_detect.rs`(1084줄), `src/check_affected.rs`(905줄), `skills/moai/workflows/project.md`, `skills/moai/references/internal/interview.md`, 이 저장소의 `.github` 16개 파일, SPEC-20261122-remote-copy-slim.

## Non-goals

- 이 저장소 자신의 `.github/` 변경: 제품 설계만 정해요. 이 저장소의 원격 사본은 SPEC-20261122-remote-copy-slim 이 맡아요.
- 만들어지는 워크플로의 내용(트리거, 비용 통제, 안전 규칙, 배포): 자식 SPEC `SPEC-20261122-project-ci-workflows` 가 정해요.
- GitHub 이외의 CI(GitLab 등)와 자체 호스트 러너: 감지되면 "지원하지 않음" 으로 알리고 아무것도 쓰지 않아요.
- 브랜치 보호·필수 검사·환경·예산·시크릿 같은 저장소 설정을 도구가 바꾸는 일: 설정 화면의 순서를 글로만 알려 줘요.
- 이미 있는 이슈 양식·라벨 생성(`github-plan`)의 변경.
- 사용자의 실제 시크릿·키·HOME 접근.
- 코드와 테스트: 이 조각은 쓰지 않아요(조각 A 이후).

## Requirements

- REQ-1: When the profile holds no CI answers, the setup interview shall ask at most six plain-language questions, each a selection with the recommended answer first and a one-line explanation of every technical word. · failure: none — a question that the repository already answers is confirmed and not asked, and an unanswered question is recorded as held and asked again on the next run.
- REQ-2: The setup interview shall confirm the detected languages, package managers and check commands first and then ask what the project is, who uses it on which operating systems, whether and where it deploys, whether the repository is public or private with the monthly budget, and how strict the checks are. · failure: none — a detected fact is shown with its source and the user can correct it, and nothing detected is asked as an open question.
- REQ-3: Each interview answer shall change at least one line of the planned workflow text. · failure: none — a question whose answer changes no planned line is removed from the question set.
- REQ-4: The planned workflow shall run exactly the check commands that the profile declares, in the declared order. · failure: none — a command that the profile does not declare never appears in a run step, and a declared command that the plan lacks is a finding.
- REQ-5: The project's local check shall run the same declared commands in the same order as the planned workflow. · failure: none — two different lists fail the comparison test, and a project that has no declared commands gets no planned check and a plain explanation.
- REQ-6: The preview shall show an estimated monthly cost in minutes and dollars, computed from the answers and a dated price table, before anything is written. · failure: none — when the estimate exceeds the stated budget the preview says so and offers the cheaper option first, and a public repository's standard runners are shown as free.
- REQ-7: The CI plan shall be shown as a preview that explains every job in plain words, and no file is written until the user picks it. · failure: none — a cancelled, empty or unanswered pick writes nothing and is reported as open.
- REQ-8: The CI apply shall keep an existing workflow file that has no recorded baseline or that the user changed since the last apply, and it is shown as a merge or a kept file. · failure: none — nothing is overwritten silently, a planned file carries its own name prefix so that it collides with no user file by default, and a user file is listed as preserved.
- REQ-9: When the project changes, the update plan shall re-plan the CI files from the saved answers without asking again. · failure: none — a changed answer goes back through the interview, and a generated file that the user edited is kept and shown as a conflict.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): in a temp project for each of four profiles (web app, API, CLI, library) the interview asks at most six questions, confirms the detected facts first, and changing any one answer changes the planned workflow text.
  verify: a Rust test with four fixture projects that varies each answer and compares the planned text — red first in the code slice
- AC-2 (REQ-4, REQ-5): the commands in the planned run steps equal the profile's declared commands, and the local check reads the same list.
  verify: a Rust test that reads the profile, the planned workflow and the local check's command list and compares all three — red first
- AC-3 (REQ-6): the preview prints minutes and dollars from a dated price table, flags an estimate above the budget, and shows a public repository as free.
  verify: a Rust test with fixed inputs that reads the printed estimate against a hand-computed value — red first
- AC-4 (REQ-7, REQ-8): the plan writes nothing before the pick, and an existing workflow file is kept and shown as conflict or preserved.
  verify: a Rust test with a pre-existing workflow, a user-edited generated file and an empty pick, reading the folder after each — red first
- AC-5 (REQ-9): the update plan re-plans the CI files from the saved answers, asks nothing, and keeps an edited file as a conflict.
  verify: a Rust test on the update plan with saved answers and a changed project — red first
- AC-6: the wider READY set passes.
  verify: `cargo test --locked --offline --bins` and `--lib`, and `moai spec check SPEC-20261122-project-ci-design`

## Scope

- write: `.moai/specs/SPEC-20261122-project-ci-design/**` and `.moai/specs/SPEC-20261122-project-ci-workflows/**`
- later slices on this card (joined by card scope when their turn comes): a profile key for the saved CI answers and one for the declared checks (`src/profile.rs`, its pinned key tests), lint and format detection (`src/project_detect.rs`), the CI plan and apply operations with their own baseline manifest (a new module and two ops files), the update step (`src/project_update.rs`), the question rows and the final-report row (`skills/moai/workflows/project.md`, `skills/moai/references/internal/interview.md`, with the skill source manifest), and their Rust tests
- read-only: this repository's `.github/**`, `src/github_templates.rs` and its tests, `tools/local_check.sh`, `SPEC-20261122-remote-copy-slim`
- forbidden: edits to this repository's `.github/**`; any write to a real repository's settings (branch protection, required checks, environments, budgets, secrets); any secret value in a plan, a file or a test; pushes; a card number in a SPEC id, a branch, a commit or a template

## Interview

- Q1: CI 설계 SPEC 둘(project-ci-design, 자식 project-ci-workflows)을 적힌 그대로 승인할까요? 승인하면 요구사항·수락 기준·범위의 다이제스트가 기록돼요. → A: 승인 (Recommended)

## Plan

1. This slice: both SPECs, committed on the branch, with a READY note that carries the head and the checked base only. The user's own pick approves them; nothing here is self-approved.
2. Slice A: the profile keys for the saved answers and the declared checks (additive, with the pinned key-list tests), and lint and format detection, with red tests first (REQ-4).
3. Slice B: the CI plan and apply operations with the preview, the estimate and the pick rule, red tests first (REQ-3, REQ-6 to REQ-8). The generated content comes from `SPEC-20261122-project-ci-workflows`.
4. Slice C: the update step and the interview rows with the question bank text (REQ-1, REQ-2, REQ-9).
5. Slice D: the local check that reads the declared list and the comparison test (REQ-5).

### Design

**Questions** (at most three per call, at most two calls, every row a selection with the recommended answer first, every technical word explained once):

| Id | Asks | Options | Changes in the planned file |
| --- | --- | --- | --- |
| `ci_stack` | confirm the detected languages, package managers and the build, test, lint and format commands | each detected value first (recommended), the usual command of the language when nothing was found | the run steps, the cache and the language setup step |
| `ci_kind` | what the project is | web app, API, command-line tool, library, mobile app | which checks come first (build and tests; for a library also a package check), and whether a deploy question follows |
| `ci_users` | who uses it and on which systems | only me or a Linux server (recommended), also Windows users, also Mac users | the extra build jobs and their per-minute price lines |
| `ci_deploy` | does it publish or deploy anywhere | no (recommended), yes with a name for where | no deploy job, or one deploy job with a named environment and the secret names it needs |
| `ci_money` | is the repository public or private, and the monthly Actions budget | public (free, recommended when it is), private with a small, medium or custom amount | the estimate line, the nightly run, the macOS offer |
| `ci_strict` | block merges when checks fail, or only report | only report (recommended for a first run), block merges | the stable job name, no path filter, and the settings steps shown as text |

**Declared checks.** The profile gets an ordered `checks` list of `{id, command}` items that come from the confirmed detection (build and test today, lint and format after the detection slice). A command is one line, with no `${{`, no secret-shaped word and a length cap, and it is shown in the preview before it is saved. The planned run steps are those commands and nothing else, and the project's local check (the generic one the later slice adds beside the repository-specific `tools/local_check.sh`) reads the same list. This repository's `tools/local_check.sh` is the example that proves the idea: it is the one set that CI also runs.

**Estimate.** For each planned job: minutes per run (a default per project kind, shown in the preview and correctable) rounded up to a whole minute, times runs per month (commits to the integration branch in the last 30 days, at least 20, shown and correctable), times the per-minute price from a data file that carries the date it was read. The preview prints one number with a plus-or-minus range, labelled as an estimate. A public repository shows zero for standard runners and says that larger runners are billed. The included monthly amount depends on the plan and the docs say it can change, so the preview names the billing page and never hard-codes an allowance. The budget the user states is compared with the estimate only; the tool cannot set a budget, so the preview gives the web path (Billing, Budgets and alerts, New budget, "Stop usage when budget limit is reached").

**Where it plugs in.** A separate pair of operations with its own plan, digest, permission reference, receipt and baseline manifest (the same engine as `github-apply`: create, merge, unchanged, conflict, adopt baseline), wired in as one more install step of `project-update-plan` so `moai update` re-plans it from the saved answers. It is not a section of `github-plan`, because that plan is gated by the Issues opt-in and its five files are pinned by tests.

## Risks

- 프로필에 검사 명령 자리가 없어서(`commands_fact` 는 `detected` 옆에만 있고 저장하지 않아요) 조각 A 가 먼저 와야 해요. 이 SPEC 은 그 자리를 새 키로 정해요.
- 제품에는 지금 일반 프로젝트용 로컬 검사가 없어요(`check affected` 와 `tools/local_check.sh` 는 이 저장소 전용). REQ-5 의 "같은 명령" 은 조각 D 가 일반 로컬 검사를 만든 뒤에야 증명할 수 있어요.
- 가격표와 한도는 문서가 "변경될 수 있다" 고 적어요. 가격표 데이터에 읽은 날짜를 두고 미리보기가 그 날짜를 보여 주지만, 날짜가 오래되면 추정이 틀려요.
- 사용자 프로젝트의 쓰기 가드(`hook_guards.rs`)는 `.github/workflows` 를 막지 않아요. 그래서 "사용자 선택 뒤에만 쓴다" 는 규칙을 계획과 다이제스트 흐름에 두어요.
- 질문이 여섯 개를 넘기 쉬워요. 감지된 사실은 묻지 않고 확인만 하고, 답이 파일을 바꾸지 않는 질문은 REQ-3 으로 지워요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | CI 는 자기 계획·적용 명령과 자기 기준선 매니페스트를 갖는 별도 작업이고, `project-update-plan` 의 설치 단계 하나로 붙어요. `github-plan` 의 한 절이 아니에요. | `github-plan` 은 이슈 옵트인에 묶이고 파일 5개를 테스트가 고정해요(코드 읽기, 2026-11-22). | 카드 열림 |
| D-2 | 프로필에 저장된 답을 담는 키와 선언된 검사 명령 목록 키를 더해요(둘 다 additive). 명령은 확인된 감지 결과에서 오고 한 줄로만 받아요. | `project_detect` 읽기: 명령은 `detected` 밖에만 있고 린트·포맷은 감지하지 않아요. | 카드 열림 |
| D-3 | 가격표는 읽은 날짜가 붙은 데이터 파일이고 플랜별 포함 금액은 하드코딩하지 않아요. | 비용 페이지 읽기: 포함량이 금액이고 변경될 수 있다고 적혀 있어요. | 카드 열림 |
| D-4 | `moai update` 는 질문을 다시 하지 않고 저장된 답과 새 감지 결과로 다시 계획만 해요. 답을 바꾸려면 `/moai project` 로 돌아가요. | `project_update` 읽기: 업데이트 경로에는 질문을 다시 하는 코드가 없어요. | 카드 열림 |
| D-5 | 도구는 브랜치 보호·필수 검사·환경 승인자·예산·시크릿을 한 번도 쓰지 않고 설정 화면의 순서를 글로만 알려 줘요. | 이 카드 본문 규칙 4, 설정 페이지 읽기 | 카드 열림 |
| D-6 | 질문 여섯 개: 감지 확인, 종류, 사용 환경, 배포, 공개 여부와 예산, 엄격함. 일 흐름은 묻지 않고 프로필의 `work_flow` 에서 읽어요. | 이 카드 본문 규칙 1·2, `project.md` 의 `work_flow` 질문 | 카드 열림 |

## Result

(status draft — written when the slice lands)
