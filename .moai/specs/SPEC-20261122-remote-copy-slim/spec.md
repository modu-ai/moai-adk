---
id: SPEC-20261122-remote-copy-slim
title: "The integration branch is copied to the remote after every landing and the remote runs only what the developer machine cannot"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261121-flow-skeleton
---

# The integration branch is copied to the remote after every landing and the remote runs only what the developer machine cannot

## Intent

흐름 재설계의 조각 2(마스터 SPEC-20261121-flow-skeleton). 사용자 결정(노트 5867의 b·c): 통합 브랜치는 착지 때마다 원격으로 복사되고, 원격 검사는 Linux 테스트와 Windows 빌드만 랜딩 배치마다 돌며 랜딩을 막지 않고, 실패한 원격 실행은 수리 카드를 만든다. 리더 조사(2026-11-21, 이 카드 본문): 원격 통합 브랜치는 2026-11-20 이후 멈춰 있고 로컬은 500커밋 이상 앞서며, 원격만 잡은 것은 Windows 컴파일 오류(15회)와 Linux 전용 실패(최대 7회), macOS 구간은 청구 가능 분의 88퍼센트를 쓰고 macOS 개발기가 도는 것을 반복했다. 릴리스 워크플로는 `release/<40 hex>`·`release/rc/<40 hex>` 만 받지만 버전 문서는 `release/v{head}.{yearweek}.{build}` 을 적어 두어 둘이 어긋난다. 이 조각은 SPEC 만 쓴다(코드는 착지 명령과 참조 가드가 착지된 뒤 이어진다). 워크플로 파일은 저장소의 가드 설정이므로 모든 워크플로 변경은 병합 전 사용자의 선택 목록에 올라간다.

## Non-goals

- 착지 명령 자체, 통합 참조(ref) 가드, 검사 선택(test selection), 호스트 어댑터 — 이 카드의 다른 조각과 별도 카드.
- macOS·arm64 원격 구간의 유지: 원격 슬림 워크플로에 두 구간이 다시 등장하지 않는다(개발기가 이미 반복하는 일).
- 원격 통합 브랜치로의 무엇이든 강제 푸시: 규칙상 금지이고 예외도 없다.
- 오래된 원격 뒤처리(500커밋 이상 뒤처진 현재 원격 분기의 재정렬)는 이 SPEC 이 정하는 빠른 앞깔림 전용 규칙 밖 — 첫 본격 복사 전 사용자 결정으로 기록(리스크 참조).

## Requirements

- REQ-1: When a landing completes, the lander shall push the integration branch to the remote with a fast-forward-only push. · failure: none — a refused or failed push leaves the landing intact, is retried, and is reported in the landing result, and force is never used.
- REQ-2: The remote workflow that a push to the integration branch triggers shall run the Linux x86_64 tests and the Windows build with no macOS and no arm64 leg. · failure: none — the job matrix on that trigger names exactly those two.
- REQ-3: The remote workflow shall collapse queued runs into one run per landing batch with one cache key shared across its jobs. · failure: none — the concurrency group and the shared cache key are read from the workflow text.
- REQ-4: The remote workflow shall retry the package-install and toolchain steps before it reports a failure. · failure: none — the retry is visible in the workflow text.
- REQ-5: The landing shall proceed with no landing-side step reading the remote run's result, whatever the remote run reports. · failure: none — a red remote run blocks nothing and undoes nothing.
- REQ-6: When a remote run fails, the lander shall create one repair card naming the run, the failing step and the landed range. · failure: none — a run id maps to one open repair card, so the same run twice creates no second card.
- REQ-7: The release workflow and the versioning text shall accept one and the same integration-branch name that the project profile names in one place. · failure: none — both are checked against that single value.

## Acceptance

- AC-1 (REQ-1): the lander's push after `card land` is fast-forward-only and a refused push leaves the landing intact and is reported.
  verify: `cargo test --locked --offline --test card_land_cli card_land_pushes_fast_forward_only_and_reports_a_refused_push` — red first in the code slice
- AC-2 (REQ-2): the workflow text on the integration-branch trigger names exactly a Linux x86_64 test job and a Windows build job.
  verify: the workflow lint test reading `.github/workflows/ci.yml` — red first
- AC-3 (REQ-3, REQ-4): the workflow text carries one concurrency group per landing batch, one shared cache key and retries on the package-install and toolchain steps.
  verify: the same workflow lint test — red first
- AC-4 (REQ-5): no landing-side step reads a remote run's result (the landing code and its tests have no remote-result dependency).
  verify: the landing tests pass with the remote unreachable — red first
- AC-5 (REQ-6): a failed remote run creates exactly one repair card naming run, step and range; the same run twice creates no second card.
  verify: the lander's repair-card test with a crafted failed-run record — red first
- AC-6 (REQ-7): the release workflow's branch filter and the versioning text accept the same name read from one profile place.
  verify: a text test reading `.github/workflows/release.yml` and the versioning document — red first
- AC-7: the wider READY set passes.
  verify: `cargo test --locked --offline --bins` and `--lib`

## Scope

- write: `.moai/specs/SPEC-20261122-remote-copy-slim/**`
- later slices on this card (joined by card scope when their turn comes): `src/ops/card_land.rs`, `.github/workflows/ci.yml`, `.github/workflows/release.yml`, the versioning document that names `release/v{head}.{yearweek}.{build}`, the repair-card writer, its tests
- read-only: `src/state_root.rs`, `src/card/landed.rs`, `.moai/specs/SPEC-20261121-flow-skeleton/**`
- forbidden: pushes to the remote integration branch before the user has seen the first preview; workflow file edits before the user's pick; macOS or arm64 legs in the remote workflow; force pushes
- workflow changes for the user's pick before any merge: 1 `.github/workflows/ci.yml` gains a trigger on pushes to the integration branch with exactly two jobs (Linux x86_64 tests and the Windows build; no macOS and no arm64 leg); 2 `ci.yml` gains one run per landing batch (a concurrency group on the branch name so a newer push replaces a queued run); 3 `ci.yml` gains one shared cache key across the jobs of that trigger; 4 `ci.yml` gains retries on the package-install and toolchain steps; 5 `.github/workflows/audit-gate.yml` and `check-guard.yml` stay pull-request-only and never run on the landing push; 6 `.github/workflows/release.yml` accepts the same integration-branch name the profile names (today it accepts only `release/<40 hex>` and `release/rc/<40 hex>`); 7 the versioning text accepts the same name as the release workflow.

## Plan

1. This slice: this SPEC, committed on the branch, READY note with head and base (local v2026.11.0 end; a moved tip is not re-merged).
2. Code slice A (after the landing command and the ref guard land): the lander's fast-forward-only push with its red tests (REQ-1, AC-1).
3. Code slice B: the slim remote workflow with the workflow lint tests red first (REQ-2–REQ-4, AC-2·AC-3); the workflow changes above go to the user's pick before merge.
4. Code slice C: the repair-card writer with its red test (REQ-6, AC-5).
5. Code slice D: the single branch name in the release workflow and the versioning text (REQ-7, AC-6).

## Risks

- 현재 원격 분기는 뒤처져 있고 그 참조를 가리키는 원격 이름이 없어요: 빠른 앞깔림 전용 규칙상 첫 본격 복사는 원격이 조상임이 확인될 때만 가능하고, 어긋나 있으면 강제 푸시가 아니라 사용자 결정이 필요해요 — 첫 복사 전에 기록으로 남겨요.
- 워크플로 파일은 가드 설정이에요: 위 목록의 변경은 병합 전 사용자의 선택을 받아요(이 조각은 파일을 만지지 않아요).
- 원격 실행 기록을 읽는 방식(gh 대신 무엇으로 수리 카드를 만들지)은 코드 조각에서 결정 사실로 남겨요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 통합 브랜치 원격 복사는 빠른 앞깔림 전용이고 강제 푸시는 없다 — 실패한 복사는 착지를 건드리지 않고 재시도·보고된다. | 사용자 결정 b(마스터 노트 5867), 이 카드 본문 1 | 카드 열림 |
| D-2 | 원격 구간은 Linux x86_64 테스트와 Windows 빌드뿐이다 — macOS·arm64 는 원격에서 사라진다(측정: 청구 가능 분의 88퍼센트, 원격만 잡는 것은 Windows 컴파일 15회·Linux 전용 최대 7회). | 리더 조사 노트 5843, 사용자 결정 c | 카드 열림 |
| D-3 | 붉은 원격 실행은 랜딩을 막지 않고 수리 카드 하나를 만든다 — 같은 실행의 반복 실패는 카드를 늘리지 않는다. | 사용자 결정 c | 카드 열림 |
| D-4 | 통합 브랜치 이름은 프로필이 한 곳에서 읽히고 릴리스 워크플로와 버전 문서가 같은 이름을 받는다 — 이름 자체(고정 이름 제안)는 사용자 미결정이라 이 SPEC 은 이름을 정하지 않는다. | 이 카드 본문 5 | 카드 열림 |
| D-5 | 워크플로 변경은 병합 전 사용자의 선택 목록에 올라간다 — 이 조각은 워크플로 파일을 만지지 않는다. | 이 카드 본문 규칙 | 카드 열림 |

## Result

(status draft — written when the slice lands)
