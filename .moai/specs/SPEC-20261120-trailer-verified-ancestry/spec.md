---
id: SPEC-20261120-trailer-verified-ancestry
title: "trailer-check: the Moai-Verified rule covers the descendants of its start commit, not the commits listed after it"
status: draft
revision: 1
created: 2026-11-20
---

# trailer-check: the Moai-Verified rule covers the descendants of its start commit, not the commits listed after it

## Intent

리더 관찰(2026-11-20, 로컬 통합 34, 범위 `1c9f1a1f95..d0c6f0cdc2`)이에요. `trailer-check` 가 38개 커밋 중 9개를 `missing_verified` 로 실패시켰는데, 아홉 개 모두 `Moai-Verified` 줄을 쓰기 전에 갈라져 나온 한 카드 브랜치의 커밋이었고, 그 브랜치 자신의 검사는 통과했어요. 원인은 규칙의 시작이 "범위에서 그 줄을 가진 첫 커밋"이고 그 뒤에 나열된 커밋을 모두 "시작 이후"로 읽는 데 있어요. 병합된 옛 가지의 커밋은 목록에서 시작 커밋 뒤에 나오지만 시작 커밋의 자손이 아니에요. 이 SPEC 은 "시작 이후"를 목록 순서가 아니라 조상 관계(시작 커밋의 자손인가)로 바꿔서 거짓 실패를 없애고, 줄을 가진 커밋의 자손이 줄을 빼먹는 경우는 그대로 실패시켜요.

## Non-goals

- `--verified-since` 가 범위 밖 커밋을 가리킬 때 범위 전체를 덮는 동작을 바꾸는 일: 그대로 둬요.
- `Moai-Verified` 값의 종류나 철자 검사(`malformed_verified`)를 바꾸는 일.
- 카드 번호 규칙(`--card-numbers-since`)의 시작 판정을 바꾸는 일: 이 SPEC 은 `Moai-Verified` 규칙만 다뤄요.
- 병합 커밋을 검사 대상에 넣는 일: 병합 커밋은 계속 목록에서 빠지고, 조상 관계만 병합을 지나가요.

## Requirements

- REQ-1: The `Moai-Verified` rule of `trailer-check` shall require the line on a commit only when a commit of the checked range that carries the line is an ancestor of it. · failure: none — a test merges an old side branch whose commits lack the line into a range that has carrying commits and expects no finding.
- REQ-2: The `Moai-Verified` rule shall keep failing a commit that descends from a carrying commit of the range and lacks the line, on the main line and on a side branch that forked after the carrying commit. · failure: none — a test reads the failing findings for both shapes.
- REQ-3: When `--verified-since` names a commit inside the range, the rule shall require the line only on the descendants of that commit. · failure: none — a test names a start with an old side branch beside it and expects the named commit itself to stay exempt.
- REQ-4: When `--verified-since` names a commit outside the range, the rule shall keep covering the whole range. · failure: none — the existing cut-at-the-rule-start test keeps passing.
- REQ-5: The `trailer-check` result shall keep reporting `verified_from` as the index of the rule's start in the oldest-first list of the range's commits. · failure: none — the existing cut-at-the-rule-start test reads that index.

## Acceptance

- AC-1 (REQ-1): 병합된 옛 가지(줄 없음)의 커밋은 `missing_verified` 를 받지 않아요.
  verify: `cargo test --locked --offline --test trailer_verified_cli trailer_check_old_side_branch`
- AC-2 (REQ-2): 줄을 가진 커밋의 자손이 줄이 없으면 본선에서도, 그 뒤에 갈라진 가지에서도 실패해요.
  verify: `cargo test --locked --offline --test trailer_verified_cli trailer_check_descendants_still_fail`
- AC-3 (REQ-3, REQ-4, REQ-5): `--verified-since` 가 범위 안이면 자손에만 요구하고, 범위 밖이면 범위 전체를 덮고, 기존 시작 판정 테스트가 그대로 통과해요.
  verify: `cargo test --locked --offline --test trailer_verified_cli`

## Scope

- write: `.moai/specs/SPEC-20261120-trailer-verified-ancestry/spec.md`, `src/trailers.rs`, `tests/trailer_verified_cli.rs`
- read-only: `src/ops/trailer_check.rs`, `src/bundle.rs`, `tests/trailer_check_cli.rs`
- forbidden: 실제 HOME, 사용자의 전역 git 설정, 네트워크

## Plan

1. 테스트 먼저: 옛 가지 병합 테스트와 자손 실패 테스트를 쓰고, 옛 가지 테스트가 지금 코드에서 실패(RED)하는 것을 읽어요.
2. `src/trailers.rs` 에서 `git rev-list --parents --reverse --topo-order <범위>` 한 번으로 부모 관계를 읽고, 시작 후보(줄을 가진 커밋 또는 `--verified-since` 커밋)에서 내려오는 자손 표시를 한 번 훑어 계산해요. 병합 커밋은 표시만 전달하고 검사 목록에는 없어요.
3. 판정 함수는 커밋마다 "자손인가" 표시를 받고, 기존 인덱스 판정은 표시가 없는 호출자를 위해 얇은 래퍼로 남겨요.
4. 기존 `trailer_verified_cli`, `trailer_check_cli`, 번들 검사 테스트를 모두 돌려요.

## Risks

- 범위에 없는 부모는 표시를 전달하지 못해요 → 범위 안 커밋만으로 판정하는 것이 "범위 안 시작" 규칙의 뜻이고, 기존 테스트(`c2..HEAD`)가 이 읽기를 고정해요.
- 정렬이 같은 초의 커밋을 섞을 수 있어요 → 판정은 목록 순서에 기대지 않고 부모 관계만 읽어요. 테스트는 커밋 날짜를 고정해서 옛 코드의 거짓 실패가 재현되게 해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 시작 후보는 범위 안에서 줄을 가진 모든 커밋이고, 한 커밋이 후보의 자손이면 규칙 아래에 있어요 | 리더가 정한 방향: 가지마다 그 가지의 첫 줄을 기준으로 판정(카드 노트) | 옛 인덱스 판정으로 돌려요 |
| D-2 | 부모 관계는 `git rev-list --parents` 한 번으로 읽어요 | 커밋마다 `merge-base --is-ancestor` 를 부르면 범위가 클 때 느려요 | 커밋별 호출로 바꿔요 |

## Result

(READY 전에 채워요.)
