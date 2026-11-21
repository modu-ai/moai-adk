---
id: SPEC-20261122-remote-copy-slim
title: "The integration branch is copied to the remote after every landing and the remote runs only what the developer machine cannot"
status: draft
revision: 3
created: 2026-11-22
parent: SPEC-20261121-flow-skeleton
---

# The integration branch is copied to the remote after every landing and the remote runs only what the developer machine cannot

## Intent

흐름 재설계의 조각 2(마스터 SPEC-20261121-flow-skeleton). 사용자 결정(2026-11-21, 마스터 SPEC 의 Decisions 와 리더 기록 "the integration branch is pushed to origin after every landing"): 통합 브랜치는 착지 때마다 원격으로 복사되고, 원격 검사는 Linux 테스트와 Windows 빌드만 착지 배치마다 돌며 착지를 막지 않고, 실패한 원격 실행은 수리 카드를 만든다. 리더 조사(2026-11-21, 이 카드 본문): 원격 통합 브랜치는 2026-11-20 이후 멈춰 있고 로컬은 500커밋 이상 앞서며, 원격만 잡은 것은 Windows 컴파일 오류(15회)와 Linux 전용 실패(최대 7회), macOS 구간은 청구 가능 분의 87.7%(9,230개 작업 전체 기준)를 쓰고 macOS 개발기가 도는 것을 반복했다. 릴리스 워크플로는 `release/<40 hex>`·`release/rc/<40 hex>` 만 받는데 버전 문서가 정하는 `release/v{head}.{yearweek}.{build}` 브랜치도 그 트리거 패턴에는 걸려 이름 검사에서 거절돼요 — 두 철자가 부딪히는 것을 풀 되돌림 가능한 옵션으로 사용자 선택에 올려요. 이 조각은 SPEC 만 쓴다(코드는 착지 명령과 참조 가드가 착지된 뒤 이어진다). 워크플로 파일은 저장소의 가드 설정이므로 모든 워크플로 변경은 병합 전 사용자의 선택 목록에 올라간다.

## Non-goals

- 착지 명령 자체, 통합 참조(ref) 가드, 검사 선택(test selection), 호스트 어댑터 — 이 카드의 다른 조각과 별도 카드.
- macOS·arm64 원격 구간의 유지: 원격 슬림 워크플로에 두 구간이 다시 등장하지 않는다(개발기가 이미 반복하는 일).
- 원격 통합 브랜치로의 무엇이든 강제 푸시: 규칙상 금지이고 예외도 없다.
- 오래된 원격 뒤처리(500커밋 이상 뒤처진 현재 원격 분기의 재정렬)는 이 SPEC 이 정하는 fast-forward 전용 규칙 밖 — 첫 본격 복사 전 사용자 결정으로 기록(리스크 참조).
- 통합 브랜치 이름의 실제 값(고정 이름 제안)은 사용자 미결정 — 이 SPEC 은 이름을 정하지 않고 한 곳에서 읽는 규칙만 정해요.
- 릴리스 분기 철자 갈등의 최종 선택: 사용자 결정(2026-11-20, `release/v{head}.{yearweek}.{build}`)을 뒤집지 않고 옵션으로만 올려요.

## Requirements

- REQ-1: When a landing completes, the lander shall push the integration branch to the remote with a fast-forward-only push. · failure: none — a refused or failed push leaves the landing intact, is retried, and is reported in the landing result, and force is never used.
- REQ-2: The remote workflow that a push to the integration branch triggers shall run exactly two jobs: the Linux x86_64 tests and the Windows build. · failure: none — a macOS or arm64 leg on that trigger is a finding.
- REQ-3: The remote workflow shall collapse queued runs into one run per landing batch with one cache key shared across its jobs. · failure: none — the concurrency group and the shared cache key are read from the workflow text.
- REQ-4: The remote workflow shall retry the package-install and toolchain steps before it reports a failure. · failure: none — the retry is visible in the workflow text.
- REQ-5: The landing shall proceed with no landing-side step reading the remote run's result, whatever the remote run reports. · failure: none — a failed remote run blocks nothing and undoes nothing.
- REQ-6: When the newest remote runs of the integration branch are read at the next landing or on demand, the reader shall open one repair card per failed run naming the run, the failing step and the landed range. · failure: none — a run id maps to one open repair card, so the same run twice creates no second card.
- REQ-7: Every release branch name the versioning text orders shall be accepted by the release workflow's trigger and name check. · failure: none — the existing spellings `release/<40 hex>` and `release/rc/<40 hex>` keep working, and no working name is dropped.
- REQ-8: The integration branch's name shall be read from the project profile in one place for both the ci.yml trigger and the lander's push. · failure: none — renaming the branch touches one profile value only.

## Acceptance

- AC-1 (REQ-1): the lander's push after `card land` is fast-forward-only and a refused push leaves the landing intact and is reported.
  verify: `cargo test --locked --offline --test card_land_cli card_land_pushes_fast_forward_only_and_reports_a_refused_push` — red first in the code slice
- AC-2 (REQ-2): the workflow text on the integration-branch trigger names exactly a Linux x86_64 test job and a Windows build job.
  verify: the workflow lint test reading `.github/workflows/ci.yml` — red first
- AC-3 (REQ-3, REQ-4): the workflow text carries one concurrency group per landing batch, one shared cache key and retries on the package-install and toolchain steps.
  verify: the same workflow lint test — red first
- AC-4 (REQ-5): no landing-side step reads a remote run's result (the landing tests pass with the remote unreachable).
  verify: the landing tests pass with the remote unreachable — red first
- AC-5 (REQ-6): reading the newest remote runs opens exactly one repair card per failed run, naming run, step and range; the same run twice creates no second card.
  verify: the reader's repair-card test with a crafted failed-run record — red first
- AC-6 (REQ-7): the versioning text's branch spelling and the existing 40-hex spellings are all accepted by the release workflow, and no working name is dropped.
  verify: a text test reading `.github/workflows/release.yml` against every spelling the versioning text orders — red first
- AC-7 (REQ-8): renaming the integration branch is one profile value change; the ci.yml trigger and the lander's push read it from there.
  verify: a test that renames the profile value and reads both — red first
- AC-8: the wider READY set passes.
  verify: `cargo test --locked --offline --bins` and `--lib`

## Scope

- write: `.moai/specs/SPEC-20261122-remote-copy-slim/**`
- later slices on this card (joined by card scope when their turn comes): the lander's push, `src/ops/card_land.rs`, `.github/workflows/ci.yml`, `.github/workflows/release.yml`, the versioning document that orders `release/v{head}.{yearweek}.{build}`, the remote-run reader and its repair-card writer, their tests
- read-only: `src/state_root.rs`, `src/card/landed.rs`, `.moai/specs/SPEC-20261121-flow-skeleton/**`
- forbidden: pushes to the remote integration branch before the user has seen the first preview; workflow file edits before the user's pick; macOS or arm64 legs in the remote workflow; force pushes
- workflow changes for the user's pick before any merge: 1 the existing push trigger on the integration branch is reduced to exactly two jobs (the Linux x86_64 tests and the Windows build; the macOS and arm64 legs leave it); 2 `ci.yml` gains one run per landing batch (a concurrency group on the branch name so a newer push replaces a queued run); 3 `ci.yml` gains one shared cache key across the jobs of that trigger; 4 `ci.yml` gains retries on the package-install and toolchain steps; 5 `ci.yml` already also triggers on `pull_request` today — 옵션 (a) keep it for the formal release pull request only · 옵션 (b) reduce it likewise; 권장은 (a); 6 `.github/workflows/audit-gate.yml` and `check-guard.yml` stay pull-request-only and never run on the landing push; 7 the release-spelling collision is settled one way (options below) — 옵션 (a) the release workflow also accepts `release/v{head}.{yearweek}.{build}` and reads the commit from the push · 옵션 (b) the publishing trigger moves to a prefix of its own (for example `release-build/`); 권장은 (a) — it keeps the user's version-name decision of 2026-11-20; the existing `release/<40 hex>` and `release/rc/<40 hex>` names keep working either way. The integration branch's name is read from the project profile in one place (고정 이름 제안 — 사용자 미결정).

## Plan

1. This slice: this SPEC, committed on the branch, READY note with head and base (local v2026.11.0 end; a moved tip is not re-merged).
2. Code slice A (after the landing command and the ref guard land): the lander's fast-forward-only push with its red tests (REQ-1, AC-1).
3. Code slice B: the slim remote workflow with the workflow lint tests red first (REQ-2–REQ-4, AC-2·AC-3); the workflow changes above go to the user's pick before merge.
4. Code slice C: the remote-run reader and its repair-card writer with their red tests (REQ-6, AC-5).
5. Code slice D: the release-spelling settlement and the one profile place for the integration branch name (REQ-7–REQ-8, AC-6–AC-7).

## Risks

- 현재 원격 분기는 뒤처져 있고 그 참조를 가리키는 원격 이름이 없어요: fast-forward 전용 규칙상 첫 본격 복사는 원격이 조상임이 확인될 때만 가능하고, 어긋나 있으면 강제 푸시가 아니라 사용자 결정이 필요해요 — 첫 복사 전에 기록으로 남겨요.
- 워크플로 파일은 가드 설정이에요: 위 목록의 변경은 병합 전 사용자의 선택을 받아요(이 조각은 파일을 만지지 않아요).
- 원격 실행 기록을 읽는 방식은 미정이에요: gh 로 읽거나 다른 방법 — 어떤 reader 를 쓸지는 코드 조각의 열린 결정으로 남겨요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 통합 브랜치 원격 복사는 fast-forward 전용이고 강제 푸시는 없다 — 실패한 복사는 착지를 건드리지 않고 재시도·보고된다. | 사용자 결정(2026-11-21, 마스터 SPEC Decisions), 이 카드 본문 1 | 카드 열림 |
| D-2 | 원격 구간은 Linux x86_64 테스트와 Windows 빌드뿐이다 — macOS·arm64 는 원격에서 사라진다(측정: 9,230개 작업 전체의 87.7%, 원격만 잡는 것은 Windows 컴파일 15회·Linux 전용 최대 7회). | 리더 조사(2026-11-21), 사용자 결정 | 카드 열림 |
| D-3 | 실패한 원격 실행은 착지를 막지 않고, 다음 착지 때 또는 필요 시 읽은 최신 원격 실행마다 수리 카드 한 장을 만든다 — 카드를 만드는 주체는 원격 실행 기록을 읽는 reader 명령이고 랜더가 아니에요(랜더는 실행이 끝난 뒤에야 실패를 알 수 있어요). reader 의 방식(gh 등)은 열린 결정. | 이 카드 본문 3, 리더 리뷰 R3 | 카드 열림 |
| D-4 | 릴리스 철자 갈등은 사용자 선택으로 풀고 이 SPEC 은 편을 들지 않는다 — 옵션 (a) 워크플로가 버전 이름도 받고 푸시에서 커밋을 읽는다(권장; 2026-11-20 사용자 결정 유지) · 옵션 (b) 발행 트리거를 자체 접두어로 옮긴다 — 어느 쪽이든 기존 40-hex 이름은 계속 동작한다. | 리더 리뷰 2 R5, 사용자 결정(2026-11-20) | 카드 열림 |
| D-5 | 워크플로 변경은 병합 전 사용자의 선택 목록에 올라간다 — 이 조각은 워크플로 파일을 만지지 않는다. | 이 카드 본문 규칙 | 카드 열림 |

## Result

(status draft — written when the slice lands)
