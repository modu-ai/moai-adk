---
id: SPEC-20261120-mflow-integration-branch-choice
title: "M-Flow first setup picks the integration branch with the user"
status: draft
revision: 1
created: 2026-11-20
parent: null
---

# M-Flow first setup picks the integration branch with the user

## Intent

M-Flow 첫 설정(/moai project 인터뷰의 `work_flow` 단계)에서 통합 가지를 사용자와 함께 골라요. 인터뷰 안내(interview.md:34,43)는 이미 후보 규칙을 서술하지만 에이전트가 raw git 으로 직접 계산해 근거가 재현되지 않아요. 프로필에는 이미 `work_flow.integration_branch` 값 하나가 있고(src/profile.rs:177-181), `spec take` 가 그것을 기본 base 로 읽어요(src/spec_take.rs:350-371); 통합 묶음(bundle)의 `--tip` 기본은 이 저장소 이름에 묶인 `origin/v2026.11.0` 하드코딩이에요(src/ops/bundle.rs:51,78,115,145,204). 통합 안전 규칙 — 이미 포함된 head 건너뜀, 기존 가지·워크트리 미변경 — 는 bundle_build_cli.rs:493,503,602,721 로 이미 시험 고정돼 있어요. 이 조각(1단계)은 질문의 근거를 내보내는 읽기 전용 명령 `work-flow-plan` 을 더해요. 두 번째 저장 값(원격 최종 대상), 묶음 `--tip` 의 프로필 기본, 인터뷰 문구 갱신은 이 카드의 뒤 조각이에요: 그 파일들이 다른 카드의 예약(C-105 는 src/profile.rs, C-181 은 src/ops/bundle.rs, C-141 은 src/bundle.rs — 2026-11-20 세 번의 `card scope` 충돌로 확인, 노트 3990)과, 문구가 고치는 interview.md 의 sha256 을 원장(source-manifest.json)이 잠그는데 그 원장은 지금 acc-auto-link 카드가 예약 중이라 D-3 처럼 뒤 조각으로 가요.

## Non-goals

- `src/profile.rs`·`src/ops/bundle.rs`·`src/bundle.rs` 편집: 예약 충돌이라 뒤 조각.
- 통합 가지·워크트리·파일·ref 생성: 이 명령은 읽기 전용이고 앞으로도 그래요.
- 열린 풀 요청 대상 조사(네트워크): 인터뷰 에이전트 몫으로 안내만 남겨요.

## Requirements

- REQ-1: The `work-flow-plan` operation shall print, reading only, the repository's branch candidates with a `canonical` flag each, one recommended branch with a one-line reason, and the project profile's recorded work-flow state.
- REQ-2: When the profile already records `work_flow.integration_branch`, the recommendation shall be that recorded value with a reason saying it is the recorded answer. · failure: none — the recorded recommendation replaces the discovery rule of REQ-3.
- REQ-3: When the profile records no integration branch, the recommendation shall come from the default branch `origin/HEAD` names when it resolves and from the current branch otherwise, each with its own reason. · failure: none — the fallbacks keep the operation reading only and still print the candidates.
- REQ-4: The operation shall print a remote ref outside the standard branch namespaces — a pull-request candidate such as `refs/remotes/origin/pr/<n>` — as `canonical: false` with a `candidate` label, so it is visible but never recommended as the integration target. · failure: none — labeling only; a `canonical: false` ref is never recommended nor deleted.
- REQ-5: The operation shall report a recorded integration target it cannot integrate into with the safe choices — a `missing-target` state with the safe-choice wording when the recorded branch is absent from the repository, an `occupied-worktree` state naming the path when a worktree other than the project folder has it checked out — creating no branch, no worktree, no file and no ref itself. · failure: none — reporting and an unchanged ref state are the whole behavior.
- REQ-6: The operation's documentation fragment shall describe the read-only behavior and the states, its installed help beginning with `Reads only.`. · failure: none — wording only.
- REQ-7: The bundle operations shall resolve an unset `--tip` as the project profile's m-flow integration branch joined to `origin/`, keeping today's `origin/v2026.11.0` default when the profile records no integration branch or cannot be read. · failure: none — an unreadable profile keeps the old default, and an explicit `--tip` always wins.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): a fresh temp repository run recommends the current branch with the discovery reason; with `origin/HEAD` set it recommends the remote default with that reason; a re-run with `work_flow.integration_branch` recorded recommends the recorded value with the recorded-answer reason.
  verify: `cargo test --test work_flow_plan_cli`
- AC-2 (REQ-4): a `refs/remotes/origin/pr/<n>` ref prints `canonical: false` with the candidate label and is never recommended.
  verify: `cargo test --test work_flow_plan_cli`
- AC-3 (REQ-5): a missing recorded target reports `missing-target` with the safe-choice wording; a worktree occupying the recorded branch reports `occupied-worktree` naming its path; the branch list is identical before and after every run.
  verify: `cargo test --test work_flow_plan_cli`
- AC-4 (REQ-6): the documentation fragment exists and `work-flow-plan --help` reads "Reads only." first; the already-merged-head skip stays pinned by the existing bundle build tests.
  verify: `cargo test --test help_layout` (regenerated goldens)
- AC-5: the wider READY set around the change passes.
  verify: `cargo test --locked --offline --bins`
  verify: `cargo test --locked --offline --lib`
- AC-6 (REQ-7): with the profile recording an m-flow integration branch, `bundle ready` without `--tip` judges `origin/<branch>`; with no recorded branch it keeps `origin/v2026.11.0`.
  verify: `cargo test --test bundle_ready_cli`

## Scope

- write: `.moai/specs/SPEC-20261120-mflow-integration-branch-choice/spec.md`, `src/work_flow.rs`, `src/ops/work_flow_plan.rs`, `src/lib.rs`, `src/help.rs`, `docs/design-review/appendix/operations/work_flow_plan.md`, `tests/work_flow_plan_cli.rs`, `tests/fixtures/help/`
- read-only: `src/profile.rs`, `skills/moai/references/internal/interview.md`
- forbidden: 가지·워크트리·ref 생성, 네트워크 호출, 카드 번호의 출력 사용

## Plan

1. RED: work_flow_plan_cli 의 일곱 장면(첫 실행·원격 기본·재실행·candidate 라벨·missing·occupied·문서 pin)이 명령 부재로 깨져요(확인: 0 passed, 7 failed).
2. GREEN: src/work_flow.rs(후보·추천·상태 논리) + src/ops/work_flow_plan.rs(얇은 층) + lib/help 등록 + 문서 조각; 문서 pin 시험은 D-3 로 뒤 조각에서(6 장면으로 GREEN 확인).
3. help 골든 재생성(MOAI_HELP_GOLDEN_WRITE=1) → 비축소 READY 검사 → 감사는 리더에게 → READY. 변이는 뒤 조각과 함께.

## Risks

- 원격 기본 가지 판정이 `origin/HEAD` 없는 저장소에서 빠져요: 현재 가지 대체가 그 빈자리를 채워요(시험 첫 장면).
- 다른 카드가 같은 파일을 고치면 충돌해요: 예약된 파일은 이 조각이 만지지 않아요(D-1, D-3).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 조각 1 은 src/profile.rs·src/ops/bundle.rs·src/bundle.rs 를 제외했어요 — 예약 충돌(노트 3990). 조각 1b(리더 선택 (a), 노트 4048)는 bundle --tip 프로필 기본을 풀린 src/ops/bundle.rs 와 bundle_ready 시험에 착지해요; remote_target 필드는 profile.rs 의 카드, src/bundle.rs 는 자기 카드 대기. | 예약 충돌 세 번 관찰, 리더 결정 | 예약 풀림 |
| D-2 | 추천 이유는 로컬 git 근거만(네트워크 없음); 풀 요청 대상 조사는 인터뷰 에이전트 몫. | 읽기 전용·오프라인 원칙 | 문구 갱신 |
| D-3 | REQ-4·AC-4(interview.md 문구와 문서 pin 시험)는 뒤 조각으로 — 문구 편집이 원장이 잠근 sha256 을 바꾸고 원장은 지금 acc-auto-link 카드 예약 중. 문구 초안은 레인 보관(/tmp/c183-interview.patch). | repo_scans.json 이 `source_manifest_covers_every_package_file`(stale sha256 거절)을 돌림 | 원장 풀림 |

## Result

(status draft — written when the slice lands)
