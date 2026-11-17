---
id: SPEC-20261117-ac-verify-required
title: "검증 방법 없는 인수 기준이 있으면 완료를 거부"
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-sdd-foundation
---

## Intent

카드 C-86(2026-11-17 사용자 지시 "권고 말고 강제")의 두 번째 조각, 규칙 "AC verify 없는 완료 거부"예요. 지금 완료가 기록되는
곳은 두 군데예요: SPEC의 `status: completed`(와 `spec check --closure`)와 `moai card done`. 인수 기준(AC)은 `verify:` 명령,
`manual:` 확인, `unknown:` 이유 중 하나를 가져야 하고(`spec check`가 이미 강제), `unknown:`은 "아직 검증할 방법이 없음"이라
경고(`ac_unknown`)에 그쳐요. 그래서 검증할 방법이 없는 기준이 있어도 SPEC을 completed로 쓰거나 그 SPEC에 묶인 카드를 닫을 수
있었어요. 이 조각은 그 두 곳에서 그런 기준을 이름으로 짚어 거부해요.

## Requirements

요구 문장은 영어 GEARS예요(`src/spec.rs` `parse_gears`). `IF …, THEN …`은 원치 않는 동작이에요.

- REQ-1: When a SPEC has status `completed`, the spec check shall report an error that names every acceptance criterion without a `verify:` command or a `manual:` check. · failure: none
- REQ-2: While a SPEC is not `completed`, the spec check shall keep an `unknown:` criterion a warning.
- REQ-3: When `moai card done` closes a card that names a SPEC, the command shall first require that every acceptance criterion of that SPEC has a `verify:` command or a `manual:` check.
- REQ-4 (REQ-3): IF the named SPEC cannot be read, fails the spec check, or has a criterion without a way to verify it, THEN `moai card done` shall refuse with the card, the SPEC id and the criterion ids, and change nothing.
- REQ-5: A card that names no SPEC shall close under the earlier rules alone.

실패 경로: REQ-3 → REQ-4(괄호로 연결). REQ-1 → 없음(줄 끝 `failure: none`: 그 동작이 곧 거부예요).

## Acceptance

- AC-1 (REQ-1): completed SPEC에서 모든 기준이 `verify:`/`manual:`이면 오류가 없고(N), `unknown:` 하나면 그 id만, 둘이면 둘 다
  적은 `ac_unverified` 오류 하나예요(N+1). `--closure`도 같아요.
  verify: `cargo test --locked --offline --test ac_verify_required a_completed_spec_refuses_a_criterion_without_a_way_to_verify_it`
- AC-2 (REQ-2): draft·approved에서는 `ac_unknown` 경고뿐이에요.
  verify: `cargo test --locked --offline --test ac_verify_required before_completion_an_unknown_criterion_is_only_a_warning`
- AC-3 (REQ-3, REQ-4): `unknown:` 기준이 둘·하나인 SPEC에 묶인 카드는 그 id를 적고 거부되며 `review`로 남고, 모두 검증 가능하면 닫혀요.
  verify: `cargo test --locked --offline --test ac_verify_required card_done_waits_until_every_criterion_of_its_spec_can_be_verified`
- AC-4 (REQ-4): 깨진 SPEC, spec.md 없는 폴더, 없는 폴더도 거부돼요.
  verify: `cargo test --locked --offline --test ac_verify_required card_done_refuses_a_linked_spec_that_is_gone_or_broken`
- AC-5 (REQ-5): SPEC을 가리키지 않는 카드는 예전처럼 닫혀요.
  verify: `cargo test --locked --offline --test ac_verify_required a_card_without_a_spec_closes_as_before`;
  `cargo test --locked --offline --test card_work_cli`
- AC-6 (REQ-1…REQ-5): 저장소 SPEC과 바뀐 곳에 닿는 검사가 통과해요.
  verify: `moai spec check --all --project .`; `cargo test --locked --offline --test spec_cli`; `cargo test --locked --offline --test spec_doc`;
  `cargo test --locked --offline --test card_spec_cli`; `cargo test --locked --offline --test product_cli`;
  `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --bin moai`; `cargo fmt --check`;
  `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/spec_doc.rs`, `src/card/spec_link.rs`, `src/card/work.rs`, `tests/ac_verify_required.rs`,
  `docs/design-review/appendix/operations/card.md`, `.moai/specs/SPEC-20261117-ac-verify-required/spec.md`
- read-only: `src/ledger/**`, `tests/card_work_cli.rs`, `tests/card_spec_cli.rs`

## Plan

1. RED: `tests/ac_verify_required.rs` 5개(SPEC 본문은 메모리, 카드는 임시 HOME·프로젝트와 가짜 호스트 연결).
2. GREEN: `spec_doc::completion`(completed일 때 `ac_unverified`), `spec_doc::unverified`, `spec_link::done_refusal`,
   `work::done` 앞머리의 거부.
3. 운영 부록 `card.md`의 `card done` 행.
4. 검사, 커밋, mutation.

## Decisions

- D1 제안("`verify:` 없는 AC가 있으면 거부")을 "`verify:` 명령도 `manual:` 확인도 없는 AC"로 바꿨어요: 사람이 보는 확인(호스트
  화면 등)은 명령이 없고, `manual:`까지 막으면 가짜 명령을 부를 수 있어요. 거부되는 것은 `unknown:`(그리고 원래 오류인 태그
  없는 AC)예요. 더 좁히려면(`manual:`도 거부) 사용자 결정으로.
- D2 SPEC 쪽 규칙은 status `completed`에만 걸어요(제안의 "done/implemented"는 이 어휘에서 `completed`; Go 세대 `implemented`는
  `approved`로 읽혀 해당 없음). 저장소의 SPEC 12개는 completed가 없어 그대로 통과해요.
- D3 카드 쪽은 SPEC status와 상관없이 걸어요: 카드 완료가 곧 완료 기록이고, 지금 SPEC은 draft로 병합돼 사용자 승인을 기다려요.
  closure(`completed` + Result)까지 요구하면 승인 전 모든 묶인 카드가 막혀요.
- D4 SPEC 거부는 카드의 다른 조건(상태·기준·PR)보다 먼저 판단해요. 묶이지 않은 카드는 예전과 같은 순서예요.

## Result

구현 에이전트가 2026-11-17에 채웠어요. SYNC와 AUDIT가 다시 확인해요. 기준 `6be96b4fc7`(origin/v2026.11.0), 변경 커밋
`77302513fc`. 빌드 폴더는 이 worktree의 `target/`, `CARGO_INCREMENTAL=0`, `CARGO_BUILD_JOBS=4`, macOS arm64, cargo는
`moai heavy-slot-acquire` 슬롯 안에서 한 번에 하나.

- AC-1…AC-5: met — RED(구현 전): exit 101, 3 FAILED(completed SPEC에 오류 0개, card done이 닫힘), 2 통과(경고만, SPEC 없는
  카드). GREEN: exit 0, 5 passed. 처음 GREEN에서 AC-4가 실패했는데, spec.md만 지운 폴더는 오류 있는 SPEC으로 읽혀 "cannot be
  read"가 아니라 "spec check"로 거부됐어요; 시험을 폴더까지 지우는 경우와 나눠 고쳤어요(구현은 그대로).
- AC-6: met — `moai spec check --all` 0(SPEC 13개, errors 0, warnings 0); `spec_cli` 0(17); `spec_doc` 0(24); `spec_contract` 0(8);
  `card_cli` 0(14); `card_dispatch_cli` 0(25); `card_export_cli` 0(17); `card_import_cli` 0(9); `card_spec_cli` 0(11);
  `card_summary_cli` 0(16); `card_times_cli` 0(7); `card_work_cli` 0(24); `mode_cards_53` 0(9); `product_cli` 0(75);
  `home_isolation_scan` 0(16); `--bin moai` 0(27); `ac_command_lint` 0(5); `cargo fmt --check` 0(처음 1, 줄바꿈 Edit 뒤 0);
  clippy 0(처음 1: 시험 머리 문서 목록 다음 줄 들여쓰기, 빈 줄을 넣은 뒤 0); Python `repo_scans`·`ci_scope`·`skill_references`·
  `source_manifest` OK; `git diff --check` 0.
- mutation 3개(커밋 뒤, 각각 `git restore` 뒤 tree 깨끗): 규칙을 `approved`에 걸기 → 2 FAIL; `manual:`도 미검증으로 셈 →
  2 FAIL; card done이 SPEC 오류를 보지 않게 → 1 FAIL. 되돌린 뒤 5 passed.

실행 안 함: 전체 `cargo test`, Linux·Windows.
