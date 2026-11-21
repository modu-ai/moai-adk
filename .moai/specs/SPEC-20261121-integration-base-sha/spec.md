---
id: SPEC-20261121-integration-base-sha
title: "M-Flow lanes start from the exact SHA of the local integration branch, never from origin; a missing local ref stops with the next step"
status: draft
revision: 1
created: 2026-11-21
---

# M-Flow 기준은 로컬 통합 브랜치의 정확한 SHA

## Interview

- Q1: 무엇이 문제였나요? → A: 사용자가 확인한 현장 점검(2026-11-21)에서 Codex CLI 화면 13개(리더 1 + 레인 12)를 다른 프로젝트에서 보니, 같은 프로젝트의 로컬 `main`(64ad8c93f)과 `origin/main`(5271432ee)이 갈라져 있었는데 어떤 레인의 워크트리 HEAD 와 다른 레인의 기록은 `origin/main` 을 썼고 다른 두 레인은 로컬 `main` 을 썼어요. 레인마다 출발점이 달랐어요.
- Q2: 어디까지 고치나요? → A: M-Flow 는 로컬 통합 브랜치가 기준이에요(PR 은 정식 릴리스 때만). 그래서 ADK 가 워크트리·브랜치를 만들거나 확인할 때 프로필의 `work_flow.integration_branch` 의 **로컬 ref** 를 정확한 SHA 로 풀고, 그 SHA 를 기록하고, 로컬 ref 가 없으면 origin 으로 대신하지 않고 멈춰요. 그 프로젝트와 살아 있는 레인 설정·보호 설정은 건드리지 않아요.
- Q3: Codex 자체 경로는요? → A: Codex 에는 새 워크트리가 어느 ref 에서 시작할지 정하는 설정이 없어요(리더의 조사, 2026-11-19, Codex 공식 문서: 앱은 "Below the composer, choose the Git branch to base the worktree on" 로 화면에서 고르고, CLI `--worktree` 는 값을 받지 않고, config 문서에 시작점 키가 없음). 강제할 수 있는 건 ADK 경로뿐이라, M-Flow 안내는 ADK 경로만 이름 붙여요.
- Q4: 겹치는 기존 SPEC 은요? → A: `SPEC-20261119-spec-take-profile-base`(통합 브랜치 이름을 기본 기준으로 읽음)와 `SPEC-20261120-mflow-integration-branch-choice`(통합 브랜치를 사용자와 고름)를 확장해요. 새 기능을 중복해서 만들지 않고, 위 둘이 정한 "이름" 기준을 "정확한 SHA, 로컬 ref 만" 으로 좁혀요.

## Intent

M-Flow 의 병합 기준은 하나의 로컬 통합 브랜치 팁이에요. 그런데 도구가 기준을 **이름**으로만 넘기면(`main`, `origin/main`) 갈라진 저장소에서 레인마다 다른 커밋에서 출발해요. 이 SPEC 은 기준을 두 단계로 고정해요: (1) 읽기 전용 명령 `moai integration-base` 가 프로필의 통합 브랜치를 `refs/heads/<branch>` 로만 풀어 40자리 SHA 를 돌려주고, 로컬 ref 가 없으면 원격 추적 ref 를 읽지 않고 다음 단계를 말하며 멈춰요. (2) 브랜치를 만드는 `moai spec take` 는 그 SHA 에서 시작하고 SHA 를 `branch.<name>.moai-base-sha` 로 기록해요. 레인의 워크트리가 정말 그 SHA 를 품고 있는지는 `moai integration-base --worktree <path>` 가 읽어서 보여 줘요. 마지막으로 레인 시작 안내가 m-flow 에서 `origin/<기준 브랜치>` 와 비교하라고 한 문장을 로컬 기준으로 고치고, Codex 앱의 워크트리 기능은 시작점을 정할 수 없다는 사실과 ADK 경로를 함께 적어요.

## Non-goals

- 그 현장 프로젝트와 살아 있는 레인의 설정·보호 설정을 바꾸는 것.
- Codex 설정(`config.toml`)에 시작점 키를 만드는 것(그런 키가 없어요), Codex 앱·CLI 자체를 바꾸는 것.
- `bundle-*` 명령의 기본 팁(`origin/<통합 브랜치>`, PR 을 만드는 쪽의 기준)을 바꾸는 것. 그건 PR 기준이라 로컬 기준과 다른 문제예요.
- m-flow 가 아닌 흐름(github-flow, trunk-based, git-flow)이나 프로필이 없는 프로젝트의 동작을 바꾸는 것(지금처럼 현재 브랜치 기준).
- 통합 브랜치를 새로 고르거나 만드는 것(`work-flow-plan` 몫), 로컬 ref 가 없을 때 대신 만들어 주는 것.

## Requirements

- REQ-1: When `moai integration-base <project>` runs on a project whose profile records the flow `m-flow` and an integration branch, the command shall print the branch, the 40-hex commit id of `refs/heads/<branch>` and `source: "local"`, and write nothing.
- REQ-2 (REQ-1): IF `refs/heads/<branch>` does not exist in the project, THEN the command shall refuse with the error kind `not_found`, name the missing local branch and the next step, and read no remote-tracking ref.
- REQ-3: Where the profile records no `m-flow` flow (another flow, no flow, or no profile), the command shall print `policy: "none"` with no commit id and exit 0.
- REQ-4: When `moai spec take` starts a fresh SPEC branch on an m-flow project without `--base` and without a recorded base, it shall start the branch from the commit id of `refs/heads/<integration branch>`, record it as `branch.<name>.moai-base-sha`, and report `base_sha` with the base source `local-integration` · failure: none
- REQ-5 (REQ-4): IF `refs/heads/<integration branch>` does not exist, THEN `spec take` shall refuse with the message of REQ-2 and change nothing, with no branch, no recorded base and no fallback to the current branch, a remote-tracking ref or a tag.
- REQ-6: When `moai integration-base <project> --worktree <path>` runs, the command shall print the worktree's `HEAD`, `contains_base` (whether the local integration commit is an ancestor of `HEAD`), the commit id recorded for the worktree's branch as `recorded_sha` (null when none), and `matches_recorded` (whether that commit id equals the local integration commit) · failure: none
- REQ-7: Where the m-flow guidance says how a lane gets its worktree, it shall name only the ADK path (`moai spec take`, and `moai worktree-apply` with the commit id `moai integration-base` prints) and say that the Codex app's own worktree creation and `--worktree` have no start-point setting.
- REQ-8: When the agent-mode start step tells a lane to compare its worktree with its base, the step shall name the local integration branch and `moai integration-base --worktree`, not `origin/<base branch>`, as the m-flow base · failure: none

## Acceptance

- AC-1 (REQ-1…REQ-3): temp-repo tests: with a local `main` that differs from `origin/main`, the command prints the local commit id and `source: "local"` and not the remote one; with the local branch deleted (only `origin/<branch>` present) it stops with `not_found`, the missing branch and the next step, and prints no commit id; with a flow that is not `m-flow`, no flow and no profile it prints `policy: "none"` and exits 0; nothing is written (`TempHome` snapshot equal).
  verify: `cargo test --offline --locked --test integration_base_cli`
- AC-2 (REQ-4…REQ-5): temp-repo tests: a fresh take on an m-flow project without `--base` creates the branch at the local integration commit (not at `origin`), writes `branch.<name>.moai-base-sha` with that commit id and reports `base_sha` and `local-integration`; a missing local integration branch refuses and leaves the branch list, the config and the checkout as they were; a re-take with a recorded base keeps it; a project whose flow is not m-flow keeps the earlier name-based profile base, and the current branch when the profile names none.
  verify: `cargo test --offline --locked --test spec_take_cli`
- AC-3 (REQ-6): temp-repo tests: a worktree made from the local integration commit has `contains_base: true` and `matches_recorded: true`; a worktree made from a diverged `origin/<branch>` has `contains_base: false`; a worktree whose branch has no recorded commit has `recorded_sha: null`.
  verify: `cargo test --offline --locked --test integration_base_cli`
- AC-4 (REQ-7…REQ-8): the named guidance sentences are present in the shipped workflow text and the start step no longer names `origin/<base branch>` as the m-flow base.
  verify: `cargo test --offline --locked --test integration_base_cli`
- AC-5 (REQ-1): the command is a file-op with its operations document, help row, example and goldens, and the skill package manifest matches the edited workflow text.
  verify: `cargo test --offline --locked --bin moai-cli operations_doc_matches`; `cargo test --offline --locked --bin moai-cli help_table_names_every_command_once`; `cargo test --offline --locked --bin moai-cli every_command_has_examples`; `cargo test --offline --locked --test help_layout`; `cargo test --offline --locked --test product_cli skill_package`
- AC-6 (REQ-1…REQ-8): at most five mutants of the new code, each turning a test red, are restored.
  verify: manual: the mutant table in the Result.

## Scope

- write: `.moai/specs/SPEC-20261121-integration-base-sha/**`, `src/ops/integration_base.rs`, `docs/design-review/appendix/operations/integration_base.md`, `tests/integration_base_cli.rs`, `src/spec_take.rs`, `tests/spec_take_cli.rs`, `src/help.rs`, `tests/fixtures/help/all.txt`, `tests/fixtures/help/all-60.txt`, `tests/fixtures/help/all-80.txt`, `tests/fixtures/help/all-dumb.txt`, `skills/moai/workflows/mode-agent.md`, `skills/moai/references/source-manifest.json`
- read-only: `src/profile.rs` (`read_profile`), `src/work_flow.rs`, `src/worktree.rs`, `src/ops/spec_take.rs`, `skills/moai/workflows/project.md`, `skills/moai/workflows/mode-leader.md`, `tests/common/home.rs`, `tests/common/git.rs`, `.moai/specs/SPEC-20261119-spec-take-profile-base/spec.md`, `.moai/specs/SPEC-20261120-mflow-integration-branch-choice/spec.md`
- forbidden: `src/profile.rs`, `src/ops/bundle.rs` and `src/bundle/**` (the PR tip stays `origin/<branch>`), every Codex or Claude settings file, the real HOME, the network, the project this was found in
- 읽은 사실(2026-11-21, 이 tip `041536fe07`): `src/spec_take.rs:351-374` 는 프로필 통합 브랜치를 기준 **이름**으로 받아 `rev-parse --verify <name>^{commit}` 로만 확인하고 `git switch -c <branch> <name>` 으로 시작하며 이름(`branch.<name>.moai-base`)만 기록해요. 이름이 태그나 다른 ref 로 풀릴 수 있고 SHA 는 안 남아요. `src/worktree.rs:615-619` 의 `worktree-apply` 는 청구 기록의 `base_sha`(40자리)로 `git worktree add` 를 해서 SHA 기준은 이미 있어요. `skills/moai/workflows/mode-agent.md` 3단계(state)는 "`HEAD` 를 `origin/<base branch>` 와 비교"라고 써서 m-flow 에서는 갈라진 기준을 가리킬 수 있어요. 번들 명령은 `--tip` 기본값을 `origin/<통합 브랜치>` 로 풀어요(`SPEC-20261120-mflow-integration-branch-choice` REQ-7, PR 기준이라 이 SPEC 의 범위 밖).
- 예약 상태: `src/ops/spec_take.rs`·`src/profile.rs`·`skills/moai/workflows/mode-agent.md` 등은 다른 카드가 쥐고 있어요(2026-11-21 읽음). 이 SPEC 은 `src/profile.rs` 를 쓰지 않고, 쓰기가 필요한 공용 파일(`src/help.rs`, `mode-agent.md`, `src/spec_take.rs` 의 한 곳)은 보유 카드가 반환하는 대로 짧게 잡아요.

## Plan

1. **S1 해석 명령 `integration-base`(읽기 전용)**: RED — `tests/integration_base_cli.rs` 를 임시 저장소(로컬/원격 갈라짐, 로컬 ref 없음, m-flow 아님·프로필 없음)로 먼저 쓰고, 워크트리 확인(`--worktree`)과 읽기 전용 스냅샷을 더해요. GREEN — `src/ops/integration_base.rs`(해석은 `refs/heads/<branch>` 에 `rev-parse --verify --quiet --end-of-options` 로만, 원격 추적 ref 는 읽지 않음). 운영 문서 조각, 도움말 행·예시·golden 4개.
2. **S2 `spec take` 가 SHA 로 시작·기록**: `src/spec_take.rs` 의 프로필 기준 분기를 같은 해석으로 바꾸고 `branch.<name>.moai-base-sha`·결과의 `base_sha`·출처 `local-integration` 을 더해요. 임시 저장소 시험 `tests/spec_take_cli.rs`.
3. **S3 안내 문장**: `mode-agent.md` 의 state 단계를 로컬 기준 + `integration-base --worktree` 로 바꾸고, ADK 경로와 "Codex 앱 워크트리는 시작점을 못 정함"을 한 곳에 적어요. 매니페스트를 다시 생성해요. 안내 문장 스캔 시험은 `tests/integration_base_cli.rs` 에 둬요.
4. 변형 5개 이하, 이웃 검사(fmt, clippy 전체, `--lib`·`--bins`, `repo_scans` 목록, 파이썬 `tests.repo_scans_test tests.source_manifest_test`, `spec check`, `trailer-check`)를 돌려 Result 에 적어요.

## Decisions

각 결정은 (무엇 · 선택지 · 기본값 · 되돌릴 수 있음 · 범주 · 기한 · 결정자) 순서예요. 모두 이 초안의 제안이고 승인은 SPEC 이 병합될 때 사용자·리더가 골라요.

- D1 (제안, 되돌릴 수 있음, 범주 none): 명령 이름은 평평한 파일 명령 `integration-base`(`src/ops/<word>.rs` 한 파일이 하위 명령이 되는 구조). 대안: `work-flow-plan` 에 SHA 를 더하기(그 명령은 후보 전체를 읽고 설명하는 계획 명령이라 한 값만 필요한 호출에는 무거워요). 기본값: 새 명령. 되돌리기: 이름만 바꾸면 돼요(배포 전).
- D2 (제안, 되돌릴 수 있음, 범주 none): "Codex 진입과 워크트리 생성"을 ADK 가 레인에게 워크트리·브랜치를 만들어 주거나 확인하는 단계 — `spec take`, `worktree-apply`, 레인 시작 안내 — 로 읽어요. Codex 앱·CLI 가 스스로 만드는 워크트리는 시작점 설정이 없어서(Q3) 강제하지 못하고, 안내가 ADK 경로를 이름 붙여요. 다르게 읽어야 하면 알려 주세요(리더).
- D3 (제안, 되돌릴 수 있음, 범주 none): 기록은 git 설정 `branch.<name>.moai-base-sha`(이미 있는 `moai-base` 옆). 대안: 파일. 기본값: git 설정, 이유: 이미 `spec take` 가 쓰고 `spec` 결과에 읽는 자리라서요.
- D4 (제안, 되돌릴 수 있음, 범주 none): 로컬 ref 가 없을 때 다음 단계 문장은 "로컬 브랜치 `<branch>` 가 없어요. 어느 커밋에서 시작할지 직접 정해 `git branch <branch> <commit>` 로 만들거나, 프로필의 통합 브랜치를 바꾼 뒤 다시 실행하세요. origin 으로 대신하지 않아요." 처럼 만들어 주는 명령은 알리지 않아요(원격 커밋을 고르는 건 사용자의 판단). 기본값: 안내만.
- D5 (제안, 되돌릴 수 있음, 범주 none): `bundle-*` 의 `origin/<통합 브랜치>` 팁은 그대로예요(Non-goals). 이유: 번들은 정식 릴리스 PR 을 만드는 쪽의 기준이고, 이 SPEC 은 레인이 일하는 로컬 기준만 고쳐요.

## Risks

- 로컬 ref 가 계획과 적용 사이에 움직일 수 있어요. 이 SPEC 은 SHA 를 기록하는 것으로 그 시점의 출발점을 남기고, 움직였는지는 `integration-base --worktree` 의 `recorded_sha` 와 `matches_recorded` 로 확인하게 해요(막지는 않아요).
- 태그·원격 추적 ref 이름이 통합 브랜치 이름과 같을 때 `rev-parse <name>` 은 먼저 맞는 쪽으로 풀려요. 해석은 항상 `refs/heads/<branch>` 로만 해서 피하고, 시험에 같은 이름의 태그와 원격 추적 ref 를 둬요.
- 공용 파일(`src/help.rs`, `mode-agent.md`, `src/ops/spec_take.rs`)을 다른 카드가 쥐고 있어요. 그 파일을 못 쓰는 동안 S1(새 파일만)을 먼저 하고, 나머지는 반환되는 대로 잡아요. 순서가 어긋나면 S2·S3 가 뒤로 밀릴 수 있어요.
- 현장의 Codex 화면은 다른 프로젝트이고 이 SPEC 은 그 프로젝트를 읽지도 고치지도 않아요. 그 레인들이 왜 다른 기준을 썼는지는(앱 UI 선택인지 도구 기본값인지) 이 SPEC 이 알아내지 못했고, 안내가 ADK 경로를 이름 붙이는 것으로만 대응해요.
- `integration-base` 가 m-flow 가 아닌 프로젝트에서 `policy: "none"` 만 말하고 멈추므로, 그런 프로젝트의 기준은 이전처럼 사람이 정해요.

## Result

(구현 뒤에 적어요.)
