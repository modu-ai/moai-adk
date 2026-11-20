---
id: SPEC-20261120-trailer-verified-ancestry
title: "trailer-check: the Moai-Verified rule covers the descendants of its start commit, not the commits listed after it"
status: draft
revision: 1
created: 2026-11-20
---

# trailer-check: the Moai-Verified rule covers the descendants of its start commit, not the commits listed after it

## Intent

리더 관찰(2026-11-20, 로컬 통합 34, 범위 `1c9f1a1f95..d0c6f0cdc2`)이에요. `trailer-check` 가 38개 커밋 중 9개를 `missing_verified` 로 실패시켰는데, 아홉 개 모두 `Moai-Verified` 줄을 쓰기 전에 갈라져 나온 한 카드 브랜치의 커밋이었고, 그 브랜치 자신의 검사는 통과했어요. 원인은 규칙의 시작이 "범위에서 그 줄을 가진 첫 커밋"이고 그 뒤에 나열된 커밋을 모두 "시작 이후"로 읽는 데 있어요. 병합된 옛 브랜치의 커밋은 목록에서 시작 커밋 뒤에 나오지만 시작 커밋의 자손이 아니에요. 이 SPEC 은 "시작 이후"를 목록 순서가 아니라 조상 관계(시작 커밋의 자손인가)로 바꿔서 거짓 실패를 없애고, 줄을 가진 커밋의 자손이 줄을 빼먹는 경우는 그대로 실패시켜요.

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

- AC-1 (REQ-1): 병합된 옛 브랜치(줄 없음)의 커밋은 `missing_verified` 를 받지 않아요.
  verify: `cargo test --locked --offline --test trailer_verified_cli trailer_check_old_side_branch`
- AC-2 (REQ-2): 줄을 가진 커밋의 자손이 줄이 없으면 본선에서도, 그 뒤에 갈라진 브랜치에서도 실패해요.
  verify: `cargo test --locked --offline --test trailer_verified_cli trailer_check_descendants_still_fail`
- AC-3 (REQ-3, REQ-4, REQ-5): `--verified-since` 가 범위 안이면 자손에만 요구하고, 범위 밖이면 범위 전체를 덮고, 기존 시작 판정 테스트가 그대로 통과해요.
  verify: `cargo test --locked --offline --test trailer_verified_cli`

## Scope

- write: `.moai/specs/SPEC-20261120-trailer-verified-ancestry/spec.md`, `src/trailers.rs`, `tests/trailer_verified_cli.rs`
- read-only: `src/ops/trailer_check.rs`, `src/bundle.rs`, `tests/trailer_check_cli.rs`
- forbidden: 실제 HOME, 사용자의 전역 git 설정, 네트워크

## Plan

1. 테스트 먼저: 옛 브랜치 병합 테스트와 자손 실패 테스트를 쓰고, 옛 브랜치 테스트가 지금 코드에서 실패(RED)하는 것을 읽어요.
2. `src/trailers.rs` 에서 `git rev-list --parents --reverse --topo-order <범위>` 한 번으로 부모 관계를 읽고, 시작 후보(줄을 가진 커밋 또는 `--verified-since` 커밋)에서 내려오는 자손 표시를 한 번 훑어 계산해요. 병합 커밋은 표시만 전달하고 검사 목록에는 없어요.
3. 판정 함수는 커밋마다 "자손인가" 표시를 받고, 기존 인덱스 판정은 표시가 없는 호출자를 위해 얇은 래퍼로 남겨요.
4. 기존 `trailer_verified_cli`, `trailer_check_cli`, 번들 검사 테스트를 모두 돌려요.

## Risks

- 범위에 없는 부모는 표시를 전달하지 못해요 → 범위 안 커밋만으로 판정하는 것이 "범위 안 시작" 규칙의 뜻이고, 기존 테스트(`c2..HEAD`)가 이 읽기를 고정해요.
- 정렬이 같은 초의 커밋을 섞을 수 있어요 → 판정은 목록 순서에 기대지 않고 부모 관계만 읽어요. 테스트는 커밋 날짜를 고정해서 옛 코드의 거짓 실패가 재현되게 해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 시작 후보는 범위 안에서 줄을 가진 모든 커밋이고, 한 커밋이 후보의 자손이면 규칙 아래에 있어요 | 리더가 정한 방향: 브랜치마다 그 브랜치의 첫 줄을 기준으로 판정(카드 노트) | 옛 인덱스 판정으로 돌려요 |
| D-2 | 부모 관계는 `git rev-list --parents` 한 번으로 읽어요 | 커밋마다 `merge-base --is-ancestor` 를 부르면 범위가 클 때 느려요 | 커밋별 호출로 바꿔요 |

## Result

RED: 테스트 커밋 `d939b150db` 에서 `cargo test --test trailer_verified_cli` 가 5개 중 `trailer_check_old_side_branch` 와 `trailer_check_descendants_still_fail` 둘이 실패(3 통과, 2 실패)했어요(RED 실행 로그). 이어 수정 커밋 `3507b5302f`(`src/trailers.rs`: `verified_cover` 가 `git rev-list --parents --reverse --topo-order` 한 번으로 시작의 자손을 표시하고 병합 커밋을 지나 전파해요) 뒤에 같은 대상이 통과했고(GREEN 실행 로그: 5 통과), `trailer_check_cli` 7개도 통과했어요.

보강: 후속 테스트 둘(`b889ed2451`, `f02fb7403e`)이 병합을 지나는 자손, `--verified-since` 가 범위 안일 때의 사슬, 날짜가 어긋난 커밋을 더해요. 마지막은 시작 커밋의 자식에 더 새로운 두 번째 자식(브랜치)을 두어서, 날짜순으로 걸으면 부모가 자식 뒤에 나열되는 모양을 만들어요. 독립 감사의 지적으로 `--verified-since` 가 범위 안의 병합 커밋을 가리키는 테스트(`trailer_check_since_names_a_merge_commit`)를 더했어요: 병합 뒤의 커밋만 실패하고 병합 전에 갈라진 커밋은 요구받지 않아요. 이 병합 시작에서도 `verified_from` 은 0 이에요(병합 커밋은 목록에 없어서, 옛 코드와 같아요).

뮤턴트(`src/trailers.rs` 에 하나씩, `--test trailer_verified_cli` 만, 한 번에 한 슬롯, 매번 `git restore` + `touch` 로 복구하고 같은 파일임을 읽었어요): 9개 모두 적용돼 9개 모두 killed. M1 `--verified-since` 를 줄을 가진 집합으로 읽음, M2 사슬로 전파하지 않음, M3 범위 밖 시작이 아무것도 덮지 않음, M4 모든 커밋이 시작, M5 병합 커밋이 표시를 넘기지 않음, M7 부모-먼저 걷기에서 `--reverse` 제거, M8 범위 밖 시작을 알아보지 않음, M9 부모 하나(any)가 아니라 전부(all)를 요구. M6(`--topo-order` 제거)은 처음 실행에서 SURVIVED 였어요(테스트 그래프가 날짜순 걷기에서도 부모를 먼저 놓아서요); `f02fb7403e` 의 두 번째 자식 테스트를 더한 뒤 같은 뮤턴트만 다시 돌려 `trailer_check_since_inside_the_range` 가 실패해 killed 가 됐어요(기준 단계는 통과, 복구 확인). 뮤턴트 출력 로그 둘(전체 실행과 M6 재실행)은 세션 스크래치패드에 있어요.

부수 효과: `--verified-since` 가 범위 안 커밋을 가리킬 때도 이제 그 자손에만 줄을 요구해요(REQ-3). 전에는 목록 순서로 그 뒤의 커밋 모두에 요구했어요. `verified_from` 의 값과 의미는 그대로예요(REQ-5).

이웃과 READY 체인(head `8b54dea972` 기준, 모두 READY 체인 스크립트가 이 실행에서 남긴 종료 코드와 로그에서 읽었어요): 여섯 가지 가벼운 검사 `git diff --check` 0, `cargo fmt --check` 0, ops `rustfmt --edition 2021` 0, `cargo clippy` 0, `build-moai-cli` 0, `cargo-tests` 0, 파이썬 `unittest` 0, `trailer-check` 0. `tools/repo_scans.json` 41개 명령 모두 exit 0(`failed=0`). 영역 대상 11개(`trailer_*`, `bundle_*`, 그리고 `trailer-check`·`trailers::`·`Moai-Verified` 를 읽는 테스트 파일: bundle_body_cli, bundle_build_cli, bundle_check_cli, bundle_ready_cli, bundle_user_assets_cli, bundle_verify_cli, claim_verify_cli, sdd_guidance, spec_branch_name_first, trailer_check_cli, trailer_verified_cli) 128개 통과, exit 0. `--lib` 1641개 통과 exit 0, `--bins` 2개 대상 56개 통과 exit 0, `python3 -m unittest tests.local_check_test` 54개 OK exit 0. 병합 뒤 커밋은 이 SPEC 문서만 바꿔요.

Not run: 전체 `--test` 한 번에, 다른 카드가 쥔 파일을 읽는 테스트의 새 tip 위 재실행, CI macOS·Windows 시간. 이 변경은 새 프로세스 호출 방식을 더하지 않아요: `verified_cover` 는 `check_rules` 가 이미 쓰는 `git` 클로저(`crate::audit_delta::command`)를 받아 써요(`src/trailers.rs:402`, `src/trailers.rs:463`).

- AC-1: met — `trailer_check_old_side_branch` 가 RED(`d939b150db`)에서 실패하고 GREEN(`3507b5302f` 뒤)에서 통과해요. M3, M4, M8 이 이 테스트나 같은 줄의 시작 판정 테스트에서 killed 예요.
- AC-2: met — `trailer_check_descendants_still_fail` 이 RED 에서 실패하고 GREEN 에서 통과해요(본선과 갈라진 브랜치, 병합을 지난 자손). M2, M5, M9 가 이 테스트에서 killed 예요.
- AC-3: met — `trailer_check_since_inside_the_range` 가 범위 안 시작에서 자손에게만 줄을 요구해요(M1, M4, M6, M7 killed). 범위 밖 시작이 범위 전체를 덮는 동작은 `trailer_check_cuts_at_the_rule_start` 가 계속 통과하고 M3, M8 에서 실패해요.
