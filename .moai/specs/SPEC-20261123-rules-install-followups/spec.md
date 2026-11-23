---
id: SPEC-20261123-rules-install-followups
title: "the update re-checks the routing plan by distinct hosts, and a relative rules link is pinned by a test"
status: draft
revision: 1
created: 2026-11-23
parent: SPEC-20261122-claude-md-home-routing
---

# the update re-checks the routing plan by distinct hosts, and a relative rules link is pinned by a test

## Intent

공통 rules 설치(`SPEC-20261122-claude-md-home-routing`)의 독립 감사가 남긴 비차단 두 가지를 고쳐요.

첫째, `moai update` 의 확인 적용은 라우팅 층을 다시 계획한 뒤 "확인한 호스트 수와 미리보기의 줄 수가 같을 때만" 계획 digest 를 비교해요(`src/update.rs` 의 `routing_hosts.len() == fresh.routing.hosts.len()`). 그런데 라우팅 계획의 줄은 호스트가 아니라 설정 폴더마다 하나라서, `CLAUDE_CONFIG_DIR` 나 `--config-dir` 로 프로필 줄이 붙으면 줄은 둘이고 호스트는 하나예요. 두 수가 영원히 달라서 이 재확인이 꺼져요(전체 digest 검사는 남아 있지만, 적용 중에 라우팅 계획이 바뀐 것을 잡는 마지막 문이 열려 있어요). 줄 수 대신 서로 다른 호스트 이름의 수로 비교해서 다시 켜요.

둘째, 소비자 폴더의 `rules` 링크가 상대 경로(`../.claude/rules`)일 때 링크 위치를 기준으로 풀어서 `linked` 로 보는 처리(`src/home_routing/rules.rs` 의 `not_reaching`, `dir.join(target)`)에 테스트가 없어요. 감사의 뮤턴트 A3(그 `join` 을 뺌)가 살아남았어요. 이 SPEC 은 그 뮤턴트를 죽이는 테스트를 더해요.

적용 중에 라우팅 계획이 바뀌는 일은 명령줄 테스트로 만들 수 없어요(앞선 부분의 적용이 라우팅 파일을 건드리지 않아요). 그래서 비교 결정을 작은 순수 함수로 빼고 그 함수를 단위 테스트로 고정해요. 이 함수를 적용 경로가 부르는지는 코드 읽기와 뮤턴트로 확인해요.

## Non-goals

- 에이전트 층과 스타일 층의 같은 비교는 바꾸지 않아요(각각 호스트마다 한 줄이거나 이미 `count_hosts` 로 세요).
- 라우팅 계획의 줄 구조, 소비자 줄 규칙, 링크를 만드는 일은 그대로예요.

## Requirements

- REQ-1: When the update applies a confirmed routing plan, the update shall compare the confirmed plan digest with the fresh one whenever the number of confirmed hosts equals the number of distinct host names in the confirmed preview.
- REQ-2 (REQ-1): IF the two digests differ, THEN the update shall refuse the routing part with the reason that the routing plan changed after the confirmed preview and write nothing.
- REQ-3: While the confirmed preview has more than one row for one host, the comparison of REQ-1 shall count that host once.
- REQ-4: When a profile's `rules` is a relative link that reaches the default folder's `rules` from the link's own folder, the preview shall report the profile's rows as `linked`. · failure: none — a link that does not reach the default folder is `not_linked` by the existing rule of the parent SPEC and stays as it is.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): unit tests of the routing re-check decision say "changed" for a preview with a profile row (two rows, one host) and a different digest, say "unchanged" for the same digest, and skip the comparison when fewer hosts were confirmed than the preview holds; the test with the profile row fails when the decision compares row counts · verify: `cargo nextest run --lib -E 'test(routing_recheck)'`
- AC-2 (REQ-4): a test links a profile's `rules` to `../.claude/rules` and sees both profile rows `linked`; the test fails when the relative target is no longer joined to the link's folder · verify: `cargo nextest run --test claude_md_home_cli -E 'test(target_a_relative_)'`
- AC-3 (REQ-1, REQ-2, REQ-3, REQ-4): the neighbouring update and routing tests still pass · verify: `cargo nextest run --test claude_md_home_cli --test host_home_cli --test update_user_layer_cli --test home_routing_cli`

## Scope

- write: `.moai/specs/SPEC-20261123-rules-install-followups/**`, `src/update.rs`, `src/home_routing/rules.rs` (only when a test needs a seam; no behaviour change), `tests/claude_md_home_cli.rs`
- read-only: `tests/host_home_cli.rs`, `tests/update_user_layer_cli.rs`, `src/home_routing.rs`, `src/home_style.rs`
- forbidden: any other path, the real HOME and the user's real Claude config folder

## Decisions

- D-1: The decision is a function `routing_recheck_moved(confirmed_hosts, preview, fresh) -> bool` next to `count_hosts` in `src/update.rs`, not a closure inside the apply, so a unit test can reach it. The apply calls it with the same arguments as before.
- D-2: Distinct host names are counted with a set, because rows of one host are not always adjacent; the style helper `count_hosts` dedups only neighbours and stays as it is.
- D-3: No command-line test can change the routing plan between the preview and the apply, so the guard against the apply not calling the function is a mutant (the call removed) and the code read, both written in the Result.

## Risks

- 비교를 다시 켜면 적용 중에 라우팅 계획이 실제로 바뀐 드문 경우 라우팅 부분이 거절돼요. 그것이 의도예요: 확인한 미리보기와 다른 것을 쓰지 않아요. 이전에는 이 경우 조용히 새 계획을 적용했어요.

## Plan

1. SPEC (this file), then `moai spec check`.
2. RED: the unit tests `routing_recheck_*` in `src/update.rs` (they need the function, so they fail to compile first; a stub that compares row counts is committed with them to show the red as a failing assertion) and `target_a_relative_*` in `tests/claude_md_home_cli.rs` (green on the current code, red on the `join` mutant).
3. GREEN: the function counts distinct hosts and the apply calls it.
4. Mutants: row-count comparison, the `join` removed, the apply call removed; each run, then restored.
5. Neighbours: the AC-3 tests, fmt, clippy, `git diff --check`, `moai trailer-check`, one bare `check affected`.

## Result

Base `bba86859fc` (the integration tip). Commits on `spec/rules-install-followups-20261123`: the SPEC, the RED tests with the behaviour-preserving function, the fix, and the commit of this Result.

- AC-1: met. RED: with `routing_recheck_moved` still comparing row counts, `cargo nextest run --lib -E 'test(routing_recheck)'` ran 4 tests, 3 passed and `routing_recheck_sees_a_changed_digest_beside_a_profile_row` failed (`assertion failed: routing_recheck_moved(1, &previewed, &plan(["/h/.claude", "/p"], "new"))`). GREEN: after the distinct-host count the same filter plus `user_layer_tests` ran 8 tests, 8 passed. The row-count comparison is the stub of the RED run itself, so it is the mutant that this test kills.
- AC-2: met. `cargo nextest run --test claude_md_home_cli -E 'test(target_a_relative_)'` passed on the current code (1 test: the relative `../.claude/rules` link reads `linked` before and after the apply, and `../other` reads `not_linked`). With `dir.join(target)` replaced by `target` in `src/home_routing/rules.rs` the same test failed at the profile row (`left: "not_linked", right: "linked"`, with the detail "is a link to ../.claude/rules"); the file was restored by hand and `git status --short` showed no tracked change.
- AC-3: met. `cargo nextest run --test claude_md_home_cli --test host_home_cli --test update_user_layer_cli --test home_routing_cli`: 93 run, 93 passed.
- Not pinned by a test (D-3): that the apply calls `routing_recheck_moved` at all. No command-line run can change the routing plan between the preview and the apply, so a mutant that removes the call survives; it was not run, and the call is on one line of `src/update.rs` read in the diff.
- Not done and not in scope: the agent-layer comparison at the same place (`agent_hosts.len() == fresh.agents.hosts.len()`) has one row per host and was left as it is.
