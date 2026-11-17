---
id: SPEC-20261117-card-done-refusal-pin
title: "거부된 card done이 아무것도 바꾸지 않음을 바이트로 고정하고, 판정 순서를 문서에 적기"
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-ac-verify-required
---

## Intent

#3174(SPEC-20261117-ac-verify-required) 감사 PASS의 비차단 지적 두 가지를 고쳐요. N1: 거부된 `card done`의 "nothing was
changed"를 시험이 바이트로 고정하지 않아요. N3: `docs/design-review/appendix/operations/card.md`가 SPEC 검사가 카드의 다른
조건보다 먼저 판정된다는 순서를 적지 않아요.

## Requirements

요구 문장은 영어 GEARS예요. 코드 동작은 바꾸지 않고 시험과 문서만 더해요.

- REQ-1: IF `card done` is refused because of the card's SPEC, THEN the tool shall leave every byte under the HOME and the project unchanged.
  · failure: none
- REQ-2: The `card` operation page shall state that the SPEC check of `card done` runs before the card's other conditions.

## Acceptance

- AC-1 (REQ-1): 시험 하나가 거부 전과 후에 임시 HOME 아래(프로젝트 포함) 모든 파일의 경로·종류·바이트를 비교해 같고, SPEC
  조건이 맞으면 같은 비교가 달라져요(카드가 닫힘).
  verify: `cargo test --test ac_verify_required a_refused_card_done_changes_no_byte`
- AC-2 (REQ-2): `card.md`의 `card done` 행이 판정 순서를 적어요.
  verify: `grep -c 'SPEC 검사가 먼저' docs/design-review/appendix/operations/card.md` → `1`

## Scope

- write: `tests/ac_verify_required.rs`, `docs/design-review/appendix/operations/card.md`,
  `.moai/specs/SPEC-20261117-card-done-refusal-pin/spec.md`
- read-only: `src/card/work.rs`, `src/card/spec_link.rs`, `tests/common/home.rs`
- forbidden: 실제 HOME, 네트워크, `.github/workflows/audit-gate.yml`

## Plan

1. 시험: 기존 `World`로 SPEC 거부를 만들고, `TempHome::snapshot`에서 경로·종류·바이트만 비교해요(시각은 비교하지 않음).
2. 문서 한 문장.
3. 검사, 커밋, mutation 2개(거부 뒤 원장에 쓰기, 판정 순서 뒤로).

## Decisions

- D1 시각(mtime)은 비교하지 않아요. 원장을 쓰기로 여는 일이 폴더 시각을 바꿀 수 있고, 지적 N1은 바이트였어요.
- D2 코드를 고치지 않고 동작만 고정해요. 그래서 RED 대신 mutation으로 시험이 잡는지 보여요.

## Result

구현 에이전트가 채웠어요. 기준 `17c19a2759`(origin/v2026.11.0, #3174 병합 `224d4bfb76` 포함), 변경 커밋 `c105eae6ca`.

- AC-1: met — `ac_verify_required` 6 통과(새 `a_refused_card_done_changes_no_byte`). mutation(커밋 뒤, 각각 `git restore` 뒤
  tree 깨끗): 거부 길에서 프로젝트에 파일 하나를 남김 → 새 시험만 FAIL; SPEC 검사를 카드 조건 뒤로 → 3 FAIL.
- AC-2: met — grep `1`.
- 함께 돌림: `card_work_cli` 24, `product_cli` 75, `cargo fmt --check`, clippy `-D warnings`, `moai spec check --all`
  (21개, errors 0), Python `ci_scope`·`skill_references` 24 OK.

실행 안 함: 전체 cargo test, Linux CI.
