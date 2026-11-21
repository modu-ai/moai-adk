---
id: SPEC-20261121-shared-build-root
title: 워크트리는 공유 빌드 루트에서 빌드하고 워크트리 안 target/ 은 보고한다 (test-load 보고서 S8)
status: draft
revision: 1
created: 2026-11-21
---

## Intent

test-load-redesign 보고서(`/Users/goos/.moai/reports/moai-adk/test-load-redesign-2026-11-21.md`
§6 S8, 2026-11-21에 읽음)의 과제예요: 70개 워크트리가 각자 `target/`을 쥐어 fseventsd 부하와
디스크를 낭비해요. 사용자 지시 2026-11-21T03:3xZ("근본원인 보고서의 모든 문제를 긴급 처리").

지금 코드에서 읽은 사실(로컬 통합 ref b4681b3a6f 워크트리에서 읽음):

- 공유 빌드 루트 기계는 이미 있어요: `src/lane_env.rs:301`의 `build_folder(home, name)` =
  `<HOME>/.moai/build/<name>`(카드 C-82, pub). heavy-slot은 이 폴더를 `CARGO_TARGET_DIR`로
  건네주고(`src/heavy_slot.rs:377,461`), `CARGO_TARGET_DIR`이 워크트리 안 `target/`이나 남의
  빌드 폴더를 가리킬 때 이름으로 경고해요. `src/worktree.rs:260`은 빌드 폴더 안에 워크트리를
  만드는 걸 거절해요.
- 갭: (1) `worktree-plan`의 출력(474-484 줄)이 이 워크트리의 빌드 폴더를 말하지 않아요.
  (2) `worktree-gc-plan`은 검사한 후보의 `trim_paths`·`preserved_in_target`은 내놓지만,
  "자기 `target/`을 쥔 워크트리"가 몇 개인지의 상시 요약이 없어요 — S8이 잰 70 → 0을 재는
  숫자가 어디에도 없어요.
- `owner.name`은 워크트리 폴더 이름이어서(`src/lane_env.rs:255-265`), 워크트리의 빌드 폴더는
  `<HOME>/.moai/build/<워크트리 폴더 이름>`이에요 — heavy-slot이 런타임에 계산하는 것과 같은 값.

관계: `src/heavy_slot.rs`는 카드 C-192가, `src/lane_env.rs`는 카드 C-227이 예약 중이라
(2026-11-21T03:3xZ 기준) 이 조각은 쓰지 않고 `lane_env::build_folder`를 호출로만 재사용해요.
경고 쪽은 이미 그 파일들에 있어요 — 이 조각은 보고(이름 대기·요약 집계)를 놓아요.

## Requirements

- REQ-1: When worktree-plan previews a worktree with a known HOME, the plan shall name the worktree's shared build folder — the `lane_env::build_folder` value `<HOME>/.moai/build/<worktree folder name>` — beside the path, leaving the field null without a known HOME. · failure: none — the plan writes nothing; an unknown HOME is today's existing refusal path for the default location.
- REQ-2: When worktree-apply creates the worktree, it shall create that build folder with mode 0700, refuse a link in its place, and report a creation failure in the answer as `build_folder_error` instead of failing the apply. · failure: none — a creation failure is itself part of the behavior above: it is reported in the answer, never silent.
- REQ-3: When worktree-gc-plan examines worktrees, the plan shall report `in_worktree_targets` — the count of examined worktrees holding their own `target/`, the summed trim bytes of those worktrees and up to five example paths — marking the sum with `bytes_partial` when a counted worktree was skipped before its bytes were measured.
- REQ-4 (REQ-3): IF no examined worktree holds a `target/`, THEN the plan shall report the count as 0 with an empty example list.

실패 경로: REQ-3 → REQ-4.

## Acceptance

- AC-1 (REQ-1): verify: `cargo test --offline --locked --test worktree_build_root` — 기본 위치의
  plan 답이 `build_folder`를 `<home>/.moai/build/<워크트리 폴더 이름>`으로 말해요.
- AC-2 (REQ-2): verify: `cargo test --offline --locked --test worktree_build_root` — apply 뒤에
  그 폴더가 0700으로 생기고, 그 자리에 심어 둔 링크는 `build_folder_error`로 답에 이름 올라요
  (apply 자체는 created).
- AC-3 (REQ-3, REQ-4): verify: `cargo test --offline --locked --test worktree_build_root` —
  `target/`을 심은 워크트리가 있는 저장소의 gc plan은 그 워크트리를 세고 예시 경로에 넣고,
  없는 저장소는 0과 빈 예시를 내놔요.
- AC-4 (REQ-1…REQ-4): verify: `./target/debug/moai spec check SPEC-20261121-shared-build-root`
  → errors 0, warnings 0.

## Scope

- 이 조각: `.moai/specs/SPEC-20261121-shared-build-root/**`, `src/worktree.rs`,
  `src/worktree_gc.rs`, `tests/worktree_build_root.rs`.
- 이어지는 조각: `src/heavy_slot.rs`(C-192), `src/lane_env.rs`(C-227) — 예약이 풀리면 그 소유
  카드와 조율. 워크트리 70개의 실제 청소는 gc-apply로 리더가(사용자 확인 절차).

## Plan

1. 시험 먼저: REQ-1·REQ-2·REQ-3·REQ-4의 시험을 tests/worktree_build_root.rs에 놓는다.
2. `worktree.rs`: plan 출력에 `build_folder`, apply가 폴더를 만들고 실패를 답에 말한다.
3. `worktree_gc.rs`: 상위 요약 `in_worktree_targets`(집계는 이미 잰 값의 재사용이라 예산 추가 없음).
4. 검사: 대상 시험 + fmt + clippy + spec check.
