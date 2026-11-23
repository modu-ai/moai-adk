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

첫째, `moai update` 의 확인 적용은 라우팅 층을 다시 계획한 뒤 "확인한 호스트 수와 미리보기의 줄 수가 같을 때만" 계획 digest 를 비교해요(`src/update.rs` 의 옛 `routing_hosts.len() == fresh.routing.hosts.len()`). 라우팅 계획의 줄은 호스트가 아니라 설정 폴더마다 하나라서, `CLAUDE_CONFIG_DIR` 로 프로필 줄이 붙으면 줄은 둘이고 호스트는 하나예요.

이 SPEC 의 첫 판은 "그래서 두 수가 영원히 달라 재확인이 꺼진다"고 적었는데 그 전제가 틀렸어요(독립 감사 B1). 확인한 호스트 목록은 미리보기 줄에서 `gate_on_skills` 가 줄마다 호스트 이름을 하나씩 담아 만들어서(`src/update.rs` 의 `for h in &layer.routing.hosts { gate.routing.push(h.host.clone()) }`), 프로필 줄이 있으면 `[claude, claude]` 로 길이 2 예요. 옛 비교(줄 수 2 대 2)는 맞았고, 첫 판의 고침이 길이 2 와 서로 다른 호스트 수 1 을 비교해서 정작 프로필 줄이 있을 때 재확인을 껐어요. 그래서 이 SPEC 의 고침은 확인한 쪽과 미리보기 쪽을 모두 서로 다른 호스트 이름의 수로 세어 비교해요. 줄이 몇 개든, 한 호스트가 몇 번 담기든 같은 답이 나와요.

둘째, 소비자 폴더의 `rules` 링크가 상대 경로(`../.claude/rules`)일 때 링크 위치를 기준으로 풀어서 `linked` 로 보는 처리(`src/home_routing/rules.rs` 의 `not_reaching`, `dir.join(target)`)에 테스트가 없어요. 감사의 뮤턴트 A3(그 `join` 을 뺌)가 살아남았어요. 이 SPEC 은 그 뮤턴트를 죽이는 테스트를 더해요.

적용 중에 라우팅 계획이 바뀌는 일은 명령줄 테스트로 만들 수 없어요(앞선 부분의 적용이 라우팅 파일을 건드리지 않아요). 그래서 두 가지를 단위 테스트로 고정해요. 하나는 비교 결정으로, 적용이 실제로 넘기는 모양(`gate_on_skills` 가 만든 호스트 이름 목록)을 입력으로 써요. 다른 하나는 적용의 라우팅 단계 전체(`routing_part`: 다시 계획 → 비교 → 거절 또는 적용)로, 미리보기 뒤에 홈이 바뀐 상태에서 돌려 거절과 "쓴 것 없음"을 보고, 같은 계획이면 적용됨을 대조로 봐요. 비교를 부르는 줄을 지우면 이 단계 테스트가 빨개져요.

## Non-goals

- 에이전트 층과 스타일 층의 같은 비교는 바꾸지 않아요(각각 호스트마다 한 줄이거나 이미 `count_hosts` 로 세요).
- 라우팅 계획의 줄 구조, 소비자 줄 규칙, 링크를 만드는 일은 그대로예요.

## Requirements

- REQ-1: When the update applies a confirmed routing plan, the update shall compare the confirmed plan digest with the fresh one whenever the distinct host names of the confirmed hosts equal the distinct host names of the confirmed preview.
- REQ-2 (REQ-1): IF the two digests differ, THEN the update shall refuse the routing part with the reason that the routing plan changed after the confirmed preview and write nothing.
- REQ-3: While the confirmed preview has more than one row for one host, the comparison of REQ-1 shall count that host once.
- REQ-4: When a profile's `rules` is a relative link that reaches the default folder's `rules` from the link's own folder, the preview shall report the profile's rows as `linked`. · failure: none — a link that does not reach the default folder is `not_linked` by the existing rule of the parent SPEC and stays as it is.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): unit tests of the routing re-check say "changed" for a preview with a profile row (two rows, one host) when the confirmed names are the ones `gate_on_skills` gives for it (`[claude, claude]`) and the digest differs, say "unchanged" for the same digest, and skip the comparison when fewer hosts were confirmed than the preview holds; a test of the apply's routing step (`routing_part`) refuses a plan that moved after the preview and writes nothing, and applies the same plan unmoved; the tests fail when the confirmed side is counted by length against the previewed distinct hosts, and when the apply step no longer calls the comparison · verify: `cargo nextest run --lib -E 'test(routing_recheck) | test(the_routing_part_)'`
- AC-2 (REQ-4): a test links a profile's `rules` to `../.claude/rules` and sees both profile rows `linked`; the test fails when the relative target is no longer joined to the link's folder · verify: `cargo nextest run --test claude_md_home_cli -E 'test(target_a_relative_)'`
- AC-3 (REQ-1, REQ-2, REQ-3, REQ-4): the neighbouring update and routing tests still pass · verify: `cargo nextest run --test claude_md_home_cli --test host_home_cli --test update_user_layer_cli --test home_routing_cli`

## Scope

- write: `.moai/specs/SPEC-20261123-rules-install-followups/**`, `src/update.rs`, `src/home_routing/rules.rs` (only when a test needs a seam; no behaviour change), `tests/claude_md_home_cli.rs`
- read-only: `tests/host_home_cli.rs`, `tests/update_user_layer_cli.rs`, `src/home_routing.rs`, `src/home_style.rs`
- forbidden: any other path, the real HOME and the user's real Claude config folder

## Decisions

- D-1: The decision is a function `routing_recheck_moved(confirmed: &[&str], preview, fresh) -> bool` next to `count_hosts` in `src/update.rs`. It takes the confirmed host names themselves (what `gate_on_skills` gives, one name per routing row), not a count, so the caller cannot hand it a count of the wrong kind. The routing step of the apply is the function `routing_part`, so a test can run plan, compare and apply together.
- D-2: Both sides are counted as sets of host names: the confirmed names and the previewed rows' hosts. Rows of one host are not always adjacent and a host may be named once per row; the style helper `count_hosts` dedups only neighbours and stays as it is.
- D-3: No command-line test can change the routing plan between the preview and the apply deterministically (the auditor's harness uses a racing process). The guard against the apply not comparing is the unit test of `routing_part` on a home that changed after the preview (mutant: the comparison made unreachable). That `apply` calls `routing_part` at all is one line read in the diff, and is also seen by the race harness run written in the Result.

## Risks

- 비교를 다시 켜면 적용 중에 라우팅 계획이 실제로 바뀐 드문 경우 라우팅 부분이 거절돼요. 그것이 의도예요: 확인한 미리보기와 다른 것을 쓰지 않아요. 이전에는 이 경우 조용히 새 계획을 적용했어요.

## Plan

1. SPEC (this file), then `moai spec check`.
2. RED: the unit tests `routing_recheck_*` in `src/update.rs` and `target_a_relative_*` in `tests/claude_md_home_cli.rs` (green on the current code, red on the `join` mutant). After audit #394 B1: the test that feeds the re-check the host list the apply really passes (red on the first fix), and the test of the whole routing step `routing_part` on a home that changed after the preview.
3. GREEN: the re-check counts both sides as sets of host names and the apply's routing step calls it.
4. Mutants: the confirmed side counted by length, the `join` removed, the comparison made unreachable in the routing step; each run, then restored. One run of the auditor's race harness against the repaired binary in throwaway homes.
5. Neighbours: the AC-3 tests, fmt, clippy, `git diff --check`, `moai trailer-check`, one bare `check affected`.

## Result

Base `bba86859fc` (the integration tip). Commits on `spec/rules-install-followups-20261123`: the SPEC, the RED tests with the behaviour-preserving function, the first fix, the commit of its Result, and, after audit #394 (verdict FAIL, B1), the RED tests of the repair, the repair and the commit of this Result.

The first fix rested on a wrong premise, written in the first Intent: it compared the length of the confirmed host list (2 for a preview with a profile row, because `gate_on_skills` names the host once per routing row) with the number of distinct previewed hosts (1), so with a profile row the re-check stayed off, which the audit reproduced in a throwaway home (a user file removed after the preview was then written). The earlier row-count comparison was right (2 against 2). The repair counts both sides as sets of host names.

- AC-1: met. RED (repair): `cargo nextest run --lib -E 'test(routing_recheck) | test(the_routing_part_)'` ran 6 tests on the first fix with the routing step moved into `routing_part`, 5 passed and `routing_recheck_reads_the_hosts_the_apply_gate_gives_for_a_profile_row` failed (`confirmed ["claude", "claude"] for 2 rows`). GREEN: after the repair the same filter plus `a_skill_copy_that_did_not_apply_gets_no_agents_and_no_section` ran 7 tests, 7 passed. Mutants, each restored by hand afterwards (`git status --short` showed no tracked change): the confirmed side counted by `confirmed.len()` against the previewed distinct hosts, 2 of 6 failed (`routing_recheck_sees_a_changed_digest_beside_a_profile_row`, `routing_recheck_reads_the_hosts_the_apply_gate_gives_for_a_profile_row`); the comparison in `routing_part` made unreachable (`if false && …`), 1 of 6 failed (`the_routing_part_refuses_a_plan_that_moved_after_the_preview_and_writes_nothing`: the routing step answered `applied` and wrote `rules/moai.md` and `rules/moai/flow.md`, as in the audit's race).
- Race run (not a test, the harness is timing based): the audit's `race-rm.sh`, copied with the paths of this worktree, 8 tries against the repaired binary in throwaway homes under `env -i` with `CLAUDE_CONFIG_DIR` set to a profile folder: every try said `routing_result.state` `refused` with `written` empty (`refused` 8).
- AC-2: met. `cargo nextest run --test claude_md_home_cli -E 'test(target_a_relative_)'` passed on the current code (1 test: the relative `../.claude/rules` link reads `linked` before and after the apply, and `../other` reads `not_linked`). With `dir.join(target)` replaced by `target` in `src/home_routing/rules.rs` the same test failed at the profile row (`left: "not_linked", right: "linked"`, with the detail "is a link to ../.claude/rules"); the file was restored by hand and `git status --short` showed no tracked change.
- AC-3: met. `cargo nextest run --lib --test claude_md_home_cli --test host_home_cli --test update_user_layer_cli --test home_routing_cli` with the filter of the four test binaries plus `routing_recheck`, `the_routing_part_` and `user_layer_tests`: 103 run, 103 passed.
- Not pinned by a test (D-3): that `apply` calls `routing_part` at all (one line in `src/update.rs`, read in the diff and seen by the race run above). A mutant that skips `routing_part` and calls the routing apply directly was not run and would survive the unit tests.
- Not done and not in scope: the agent-layer comparison at the same place (`agent_hosts.len() == fresh.agents.hosts.len()`) has one row per host and was left as it is.
