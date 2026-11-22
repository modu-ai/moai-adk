---
id: SPEC-20261122-project-ci-workflows
title: "The generated CI workflows are cheap and safe and follow the project's work flow"
status: draft
revision: 1
created: 2026-11-22
approval: {by: user, at: 2026-11-22, basis: "AskUserQuestion selection r1: 승인 (Recommended)", digest: "sha256:46f9aac969b0d45d8c2a7388e7269fda29f2776ca8ec157cdd234eb9e12aad7d"}
parent: SPEC-20261122-project-ci-design
---

# The generated CI workflows are cheap and safe and follow the project's work flow

## Intent

부모 SPEC `SPEC-20261122-project-ci-design` 가 정한 질문과 계획 뒤에 **실제로 만들어지는 `.github` 파일에 무엇이 들어가는가** 를 정해요. 근거가 된 공식 문서 읽기 기록(Research 표)은 부모 SPEC 에 있어요. 원칙은 네 가지예요: 프로젝트의 일 흐름을 따른다, 비용을 먼저 줄인다, 기본값이 안전하다, 사용자가 읽고 끌 수 있다. 이 조각은 SPEC 만 써요.

## Non-goals

- 질문, 미리보기, 선택, 기준선, 업데이트 재계획: 부모 SPEC 이 정해요.
- 이 저장소 자신의 `.github/` 변경.
- 저장소 설정(브랜치 보호, 필수 검사, 환경 승인자, 예산, 시크릿)을 도구가 쓰는 일.
- GitLab 같은 다른 CI.
- 코드와 테스트: 이 조각은 쓰지 않아요.

## Requirements

- REQ-1: The generated triggers shall follow the profile's work flow and integration branch. · failure: none — an unknown work flow value generates pull-request checks only and says so in the preview.
- REQ-2: When the work flow is m-flow, the generator shall add one check per landing batch that blocks no landing and one nightly full run at an off-peak minute that is not on the hour. · failure: none — a failed run only reports, and the preview says that a public repository's nightly run switches itself off after 60 days without activity.
- REQ-3: The generated workflow shall use one Linux runner by default, and another operating system appears only when an answer names it and then builds without running tests unless the user asks for them. · failure: none — macOS appears only after an explicit yes that shows its per-minute price.
- REQ-4: Every generated job shall carry a timeout, a dependency cache for the project's package manager and a concurrency group that lets a newer run replace a queued one. · failure: none — a job that lacks one of the three is a finding of the lint test.
- REQ-5: Every generated workflow file shall begin with a plain-words comment block that says what the file does and how to turn it off, and its token permissions are read-only at the top level, every action is pinned to a full commit SHA, event data reaches scripts only through environment variables and the pull_request_target trigger never appears. · failure: none — the lint test refuses to emit a file that breaks one of these rules.
- REQ-6: The generator shall write secret names only, with the steps to add each one in the repository settings, and never a secret value. · failure: none — a value that has the shape of a secret in a plan or a file is refused before anything is written.
- REQ-7: A deploy job shall appear in the generated files only after an explicit yes to deploying, and it references a named environment. · failure: none — the preview says that required reviewers on an environment are a repository setting that the tool cannot apply, and which plans offer it for private repositories.
- REQ-8: When the user chooses to block merges on failed checks, the generated pull-request check shall keep one stable job name and no path filter. · failure: none — the settings steps that require the check are shown as text and never applied, and a skipped workflow can no longer leave a required check pending.

## Acceptance

- AC-1 (REQ-1, REQ-2): each work flow value produces the triggers of the design table, and m-flow adds the landing-batch check and the off-peak nightly run, with no path filter on a blocking check.
  verify: a Rust test over the four work flow values reading the generated trigger text — red first in the code slice
- AC-2 (REQ-3, REQ-4): no non-Linux runner appears without an answer that names it, macOS never appears without its price shown, and every job has a timeout, a cache and a concurrency group.
  verify: the workflow lint test over the four profiles (web app, API, CLI, library) and the operating-system answers — red first
- AC-3 (REQ-5, REQ-6): every generated file has the plain comment block, top-level read-only permissions, full-SHA pins, no event data in a run value, no pull_request_target and no secret value.
  verify: the same lint test applied to every generated file of the four profiles, plus a negative case per rule — red first
- AC-4 (REQ-7): a deploy job exists only after the yes and names an environment.
  verify: a Rust test that plans with and without the yes and reads the files — red first
- AC-5 (REQ-8): the blocking choice keeps one stable job name and shows the settings steps as text without calling any settings interface.
  verify: a Rust test that reads the generated job name and the preview text, and a scan that no settings call exists in the module — red first
- AC-6: the wider READY set passes.
  verify: `cargo test --locked --offline --bins` and `--lib`, and `moai spec check SPEC-20261122-project-ci-workflows`

## Scope

- write: `.moai/specs/SPEC-20261122-project-ci-workflows/**`
- later slices on this card (joined by card scope when their turn comes): the embedded product templates for the four profiles (a new folder under `templates/`), the generator and the lint test in the CI module of the parent's slice B, and their Rust tests
- read-only: this repository's `.github/**` (the patterns worth reusing), `SPEC-20261122-remote-copy-slim`, `SPEC-20261122-project-ci-design`
- forbidden: edits to this repository's `.github/**`; a settings write of any kind; any secret value; pushes; a card number in a SPEC id, a branch, a commit or a template

## Interview

- Q1: CI 설계 SPEC 둘(project-ci-design, 자식 project-ci-workflows)을 적힌 그대로 승인할까요? 승인하면 요구사항·수락 기준·범위의 다이제스트가 기록돼요. → A: 승인 (Recommended)

## Plan

1. This slice: this SPEC, committed with the parent, with the same READY note.
2. Slice B2 (inside the parent's slice B): the four profile templates and the generator, with the lint tests red first (REQ-1 to REQ-8).
3. Slice B3: the dated price table the estimate reads (shared with the parent's REQ-6).

### Design

**Work flow mapping** (REQ-1, REQ-2). The profile has the `work_flow` object with `flow` and `integration_branch`; only `m-flow` has behaviour in code today and no `next-flow` value exists, so the mapping reads the four existing values:

| `flow` | Triggers | Concurrency | Extra |
| --- | --- | --- | --- |
| `m-flow` | push to the integration branch (the one name from the profile), no pull_request | one group per branch, queued runs replaced, a running run never cut | one nightly full run at an off-peak minute; a failed run reports and blocks no landing |
| `github-flow` | pull_request to the default branch, push to the default branch | per pull request number, a newer push cancels the older run on a pull request only | none |
| `trunk-based` | push to the trunk, pull_request when one is opened | same as `github-flow` | none |
| `git-flow` | pull_request to `develop`, `release/*` and the default branch, push to `develop` | same as `github-flow` | none |

**Generated files.** `.github/workflows/moai-check.yml` always; `.github/workflows/moai-nightly.yml` for m-flow; `.github/workflows/moai-deploy.yml` only after the deploy yes; `.github/dependabot.yml` for the `github-actions` ecosystem only when the user picks that item in the preview. The `moai-` prefix keeps a first run from colliding with a user's own `ci.yml`.

**Runners.** Linux runs the tests. Another system runs a build only: Windows when Windows users were named, macOS only after the price line was shown and accepted. The per-minute prices come from the dated data file of the parent's estimate.

**Safety rules** (REQ-5, REQ-6), each testable as text: a `permissions:` key at the top with `contents: read`; every `uses:` ends with a 40-hex SHA and a trailing version comment (the comment is what Dependabot updates; the docs only mention it in a Dependabot caveat); no `${{` inside any `run:` value; no `pull_request_target`; `timeout-minutes` on every job; `concurrency` on every workflow; a comment block as the first lines; no secret-shaped word; no `workflow_dispatch` input that reaches a shell. A pattern worth keeping from this repository's own `ci.yml`: a group such as the pull request number or the commit, `cancel-in-progress` only for pull requests, a cache that is saved only on a push, and a toolchain install loop that retries three times.

**Docs-only shortcut.** A blocking check never uses a `paths` filter. A small first job decides the scope inside the run (the way this repository's `tools/ci_scope.py` does), so the required job always reports.

## Risks

- 완전한 SHA 고정은 시간이 지나면 낡아요. 문서가 말하는 더 강한 방법인 의존성 락파일은 일부 저장소 기능이라 기본으로 쓰지 않고, `dependabot.yml` 을 사용자가 고를 수 있게 해요.
- 예약 실행은 활동이 없는 공개 저장소에서 60일 뒤 꺼져요. 미리보기와 파일 머리 주석이 이 사실을 알려요.
- 경로 필터로 건너뛴 워크플로는 필수 검사를 "Pending" 으로 남겨요. 그래서 막는 검사에는 경로 필터를 쓰지 않아요.
- `pull_request_target` 의 시크릿 접근 여부는 문서에서 확인하지 못했어요. 그래서 이 트리거는 아예 쓰지 않아요.
- 배포 승인자는 Free·Pro·Team 의 비공개 저장소에서 쓸 수 없어요(문서 두 페이지의 표현이 조금 달라 더 좁은 쪽을 따라요). 그런 저장소에서 "배포 yes" 를 고르면 승인 없이 도는 배포가 되므로 미리보기가 그 사실을 먼저 알려요.
- 일 흐름 이름은 아직 바뀌는 중이에요(NeXT). 이 SPEC 은 프로필의 기존 값 넷만 읽어서 이름이 바뀌어도 한 곳만 고치면 돼요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 생성 파일 이름은 `moai-` 접두어(`moai-check.yml`, `moai-nightly.yml`, `moai-deploy.yml`)예요. 사용자의 `ci.yml` 과 처음부터 부딪히지 않아요. | 이 카드 본문 규칙 5 | 카드 열림 |
| D-2 | Linux 가 테스트하고 다른 운영체제는 빌드만 해요. macOS 는 사용자가 가격을 보고 yes 한 경우만 넣어요. | 단가표(Linux 0.006, Windows 0.010, macOS 0.062 달러/분), SPEC-20261122-remote-copy-slim D-2 | 카드 열림 |
| D-3 | `dependabot.yml`(`github-actions` 생태계)은 미리보기의 별도 항목이고 사용자가 고를 때만 써요. | 보안 페이지 읽기: SHA 고정은 갱신이 필요하고 Dependabot 이 고정된 SHA 와 같은 줄의 버전 주석을 갱신해요. | 카드 열림 |
| D-4 | 일 흐름 매핑은 프로필의 기존 값 넷만 읽어요. NeXT-Flow 이름은 SPEC-20261122-remote-copy-slim 이 통합 브랜치 이름과 함께 한 곳에서 바꾸는 부분을 따라가요. | `work_flow` 읽기: 행동이 있는 값은 `m-flow` 뿐이고 `next-flow` 값은 코드에 없어요. | 카드 열림 |
| D-5 | 막는 검사는 안정된 작업 이름 하나를 갖고 경로 필터를 쓰지 않아요. | 이벤트 페이지 읽기: 건너뛴 워크플로의 필수 검사는 Pending 으로 남아 병합을 막아요. | 카드 열림 |
| D-6 | `pull_request_target` 은 쓰지 않아요. | 보안 페이지 읽기: 쓰기 토큰으로 돌고 PR 코드를 검사하면 위험해요. 시크릿 접근 여부는 확인하지 못했어요. | 카드 열림 |

## Result

(status draft — written when the slice lands)
