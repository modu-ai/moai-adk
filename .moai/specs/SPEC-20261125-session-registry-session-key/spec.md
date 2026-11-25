---
id: SPEC-20261125-session-registry-session-key
title: "Session registry keyed by session id, lane registration step, one name for both -a forms"
status: draft
revision: 1
created: 2026-11-25
---

# Session registry keyed by session id, lane registration step, one name for both -a forms

## Intent

사용자 결정 2026-11-25 (결정 로그 `11:27Z(11-25)` 줄)이에요. 세션 레지스트리 항목의 키는 호스트 세션 id 예요. 그래서 한 세션이 항목을 둘 갖지 않고, 다른 이름의 세션이 살아 있는 항목을 대신 쓰지도 않아요. 레인은 라우트 글의 단계로 등록·갱신·해제하므로 작업 중 한도가 살아 있는 레인을 세어요. `-a` 는 `agent-<name>` 과 `<project>-agent-<name>` 을 한 이름으로 봐요. 한도 숫자(레인 수 곱하기 2)는 바뀌지 않아요.

## Requirements

- REQ-1: When a session with a host session id registers, the registry shall write its entry to the file named by that session id, with the lane name kept inside the entry. · failure: REQ-2
- REQ-2 (REQ-1): IF another live session already holds the registered name, THEN the registry shall refuse the registration and leave both entries as they were.
- REQ-3: When a session registers again under another name, the registry shall replace that session's own entry so that one entry is left for the session. · failure: none — the replaced entry is the session's own, and no other session's entry is touched.
- REQ-4: The registry shall keep reading an entry whose file is named by the lane name. · failure: none — an entry that cannot be read is an error as it is today.
- REQ-5: When a lookup by name runs (`--show`, `--release`, `--lane-check` or a liveness lookup), the registry shall find the entry of either file form. · failure: none — a name with no entry answers `free` or `not_lane` as it does today.
- REQ-6: When the work-in-progress limit counts lanes, it shall count each live registered session once, whatever the form of its entry file. · failure: none — the multiple of 2 per lane and the limit off at 0 lanes stay as they are.
- REQ-7: The lane route text shall order the session registration, its renewal inside the lease and its release at the end as steps of the lane. · failure: none — a lane that skips a step is counted as before, by its live entries only.
- REQ-8: When `-a` picks a lane name, the launcher shall treat a live lock or a live lease of `agent-<name>` and of `<project>-agent-<name>` as one taken name. · failure: none — with every name taken the pick refuses as it does today.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5): a Rust test registers sessions with distinct host session ids in a temporary clone and shows the entry file name, the refusal of a name held by another live session with both entries unchanged, one entry after a rename, and the name lookups reading an entry of each form · verify: `cargo test --locked --test session_registry_owner_id`
- AC-2 (REQ-6, REQ-7): a Rust test shows the limit counting a session once for two entry forms and the lane route text holding the register, renew and release steps, with the multiple unchanged · verify: `cargo test --locked --test session_registry_owner_id --test lane_team_names`
- AC-3 (REQ-8): a Rust test shows the `-a` pick skipping a name taken in either form and returning the first name free in both · verify: `cargo test --locked --test lane_team_names`

## Scope

- write: `.moai/specs/SPEC-20261125-session-registry-session-key/**`
- write: `src/lane_registry.rs`
- write: `src/session.rs`
- write: `src/card/wip_limit.rs`
- write: `src/ops/session_register.rs`
- write: `src/leader_round/registration.rs`
- write: `skills/moai/workflows/work.md`
- write: `skills/moai/references/source-manifest.json` (one sha256 line, through a line lease)
- write: `tests/session_registry_owner_id.rs`
- write: `tests/session_registry_cli.rs`
- write: `tests/lane_team_names.rs`
- write: `tests/launcher_liveness_cli.rs`
- write: `tests/leader_registration_cli.rs`
- write: `tests/leader_round_cli.rs`
- write: `tests/lane_next_proposals_cli.rs`
- write: `tests/card/card_adopt_cli.rs`
- write: `tests/card/card_reclaim_early_cli.rs`
- write: `tests/card/orchestrate_cli.rs`
- write: `tests/card/card_wip_limit.rs`
- write: `tests/card/card_land_batch_cli.rs`
- write: `tests/card/card_leader_release_cli.rs` (the path helper only, after the NeXT tip merge)

## Plan

1. Write the failing tests: two live sessions of one name, a rename of one session, both name forms of `-a`, the limit count.
2. Key the entry file by the host session id, keep the name-keyed read path, and make the name lookups scan both forms.
3. Count distinct live sessions in the limit and make the `-a` pick read both name forms.
4. Add the register, renew and release steps to the lane route text.

## Result

- AC-1 met: `tests/session_registry_owner_id.rs` 의 `ac1_*` 테스트 17개 (세션 id 로 항목 파일 이름을 정해요; 잘못된 id, `-` 로 시작하는 id, id 없음은 `<name>.json` 을 지켜요; 다른 살아 있는 세션이 쥔 이름은 두 파일 형태 모두에서 거절되고 두 항목은 그대로예요; 끝난 항목은 교체돼요; 이름을 바꿔도 항목은 하나예요; 갱신은 파일을 그대로 둬요; `--continue` 는 항목을 새 세션 파일로 옮겨요; `--show`, `--release`, `--lane-check`, 살아 있음 조회, `session_entries`/`live_session_count` 가 두 형태를 읽어요; 깨지거나 맞지 않는 항목은 오류로 남고 파일도 남아요; 세션 id 로 된 리더 항목을 리더 등록이 읽어요). 코드 전 RED 는 14개 실패, 7개 통과(14개는 옛 이름 키 동작 때문에 실패)였고, 코드 뒤에는 21개 통과(`ac1` 17개와 `ac2` 4개)예요. `session_registry_cli` 32개, `session_lane_check_cli` 5개, `cargo test --lib lane_registry` 18개가 통과했어요.
- AC-2 met: `ac2_live_lanes_counts_a_session_once_for_two_entry_forms` (RED: left 2, right 1), `ac2_live_lanes_leaves_out_leaders_and_ended_leases`, `ac2_an_entry_without_a_session_counts_by_name`, `ac2_the_multiple_per_lane_is_still_two` (`DEFAULT_PER_LANE` 은 2 이고 0 레인은 한도 꺼짐 그대로예요). 라우트 글은 `tests/lane_team_names.rs` 의 `the_agent_workflow_orders_registration_renewal_and_release` 가 고정해요(갱신 줄은 `--mode agent` 를 지키고 Procedure 단계 키는 건드리지 않아요). `skills/moai/workflows/work.md` 에는 등록 문단 뒤에 새 문단이 하나 있고 새 번호 단계는 없어요.
- AC-3 met: `tests/lane_team_names.rs` 의 `ac3_*` 테스트 다섯 개 (선택이 짧은 이름 `agent-<name>` 의 잠금이나 살아 있는 임대가 있는 이름을 건너뛰고, 두 형태 모두 비어 있는 첫 이름을 돌려줘요; 끝난 임대와 쥐지 않은 잠금은 이름을 막지 않아요; 이름이 모두 차면 `--name` 을 가리키며 거절해요). 라우트 글 테스트를 더해 RED 는 6개 실패, 10개 통과였고 GREEN 은 16개 통과예요.
- 픽스처: 기존 테스트 파일 아홉 개가 세션 id 로 등록한 직후 `sessions/<name>.json` 을 열거나 덮어썼어요. 경로 도우미만 `lane_registry::entry_file` 과 `<name>.json` 대체 경로로 바꿨고 단언은 하나도 바꾸지 않았어요: `tests/launcher_liveness_cli.rs`, `tests/leader_registration_cli.rs`, `tests/leader_round_cli.rs`, `tests/card/card_adopt_cli.rs`, `tests/card/card_reclaim_early_cli.rs`, `tests/card/orchestrate_cli.rs`, `tests/card/card_wip_limit.rs`, `tests/card/card_land_batch_cli.rs` (RED: card 대상의 다섯 모듈에서 2, 6, 1, 39개 실패; GREEN: 7, 10, 19개 통과와 113개 통과), `tests/lane_next_proposals_cli.rs` (등록 도우미가 세션의 호스트 id 를 물려받아서 이제 `CLAUDE_CODE_SESSION_ID` 를 지워요; RED 1개 실패, GREEN 5개 통과). `leader_registration_cli` 의 `leader.json`/`boss.json` 부재 단언 둘은 `entry_file` 로 두 파일 형태를 물어요. `launcher_liveness_cli`, `leader_registration_cli`, `leader_round_cli`, card 대상(113개)의 nextest 네 번이 통과했어요. `cargo clippy --locked --all-targets -- -D warnings` 는 종료 코드 0 이에요.
- 동작 메모: 살아 있는 항목의 갱신은 파일을 그대로 둬요. 임대가 끝나 `<name>.json` 에 있는 항목은 `<세션 id>.json` 으로 다시 쓰고 옛 파일을 지워요(새 등록은 세션 자신의 파일을 써요). 세션은 이름을 바꿔도 항목 하나를 지키고, 새 항목을 쓴 뒤 자기 옛 항목을 지워요. 다른 세션의 살아 있는 항목은 지우지 않아요. `--continue` 의미와 리더 규칙은 그대로고 `.live` 잠금은 이름 키 그대로예요.
- 범위 목록에 없던 파일: 리더의 결정(카드 노트 9238, 노트 9231 의 답)으로 카드 범위를 넓혔어요. 처음에 `src/leader_round/registration.rs` 와 픽스처 일곱 개를, 같은 이유로 `tests/leader_round_cli.rs` 와 `tests/lane_next_proposals_cli.rs` 를 더했어요. 새 테스트 파일 이름은 `tests/session_registry_owner_id.rs` 예요. `.gitignore` 139번째 줄(`*_key.*`)이 계획한 이름 `tests/session_registry_session_key.rs` 를 무시하기 때문이에요. `## Scope` 의 쓰기 목록을 실제로 바뀐 파일로 고쳤어요(`git diff --stat 59ac3344cd..9bb3370639` 의 20개 경로).
- 돌리지 않은 것 (not_run): `moai check affected --card`(영수증은 READY 노트에 있어요), `--lib` 전체, `--bin moai` 와 `--bin moai-cli` 전체(도움말 골든 파일), `card` 대상 전체, `session_registry_owner_id`·`session_registry_cli`·`lane_team_names` 의 nextest, 가드 뮤턴트(tier 1 이라 요구하지 않아요).
- 남은 틈: `moai cc -a` 는 아직 `.live` 잠금만 쥐고 등록하지 않아요. 그래서 기준 2 는 라우트 글로 충족하고, 항목이 없는 런처 전용 레인은 한도에서 여전히 0 으로 세요(후속: `<project>-agent-<name>` 이나 `agent-<n>` 형태의 `<name>.live` 잠금을 세요). `session-register --release` 와 `--show` 는 여전히 인스턴스를 어떤 호출자에게나 넘겨요(그대로이고 이 SPEC 의 밖이에요).
- 확인한 개정: head `9bb3370639`, base `59ac3344cd`. 로그는 `target/b1`, `target/b2`, `target/b3-*.log` 에 있어요.
- 팁 재병합 (NeXT `64b948a105`, 일곱째 묶음 뒤): 같은 파일 `src/card/wip_limit.rs` 에서 SPEC-20261125-wip-limit-live-lock-lanes 의 `Lanes` 형 `live_lanes` 와 충돌했고, 그쪽 함수를 받아 세션 단위로 고쳤어요. `.live` 잠금과 항목은 이름으로 합쳐 한 레인이고(두 이름 모양 `agent-<name>`, `<project>-agent-<name>` 은 한 이름), 잠금 없는 살아 있는 항목은 호스트 세션 id 로 한 번만 세어요. `lane_registry::scan_registry` 는 `read_full` 대신 `read_any` 로 읽어요: 세션 id 로 된 항목 파일을 읽지 못해 `unreadable` 로 빼면 그 레인이 한도에서 사라지기 때문이에요. 테스트: `tests/session_registry_owner_id.rs` 의 `ac2_a_session_keyed_entry_is_counted_not_unreadable`, `ac2_a_held_lock_and_a_session_keyed_entry_are_one_lane`, `ac2_a_lock_alone_counts_once_for_either_name_form` (세 개 통과, 병합 전에 RED 로 돌리지는 않았어요: not_run). `ac2_*` 와 `card` 대상의 `lock_ac*`·`card_wip_limit` 31개와 `session_registry_owner_id` 7개가 병합 head 에서 통과했어요. 같은 병합의 첫 영수증은 FAIL 이었어요(2025개 중 3개 실패): 일곱째 묶음의 새 픽스처 `tests/card/card_leader_release_cli.rs` 가 세션 id 로 등록한 직후 `sessions/<name>.json` 을 열었어요(`ac8`, `ac9`, `ac10`, 파일 없음). 경로 도우미만 `lane_registry::entry_file` 로 바꿨고 단언은 바꾸지 않았어요. `card_leader_release_cli::` 6개가 통과했어요.
