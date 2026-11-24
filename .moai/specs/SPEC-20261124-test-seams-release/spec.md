---
id: SPEC-20261124-test-seams-release
title: "Four test-only seams stop being read by a release build"
status: approved
revision: 1
created: 2026-11-24
approval: {by: leader, at: 2026-11-24, basis: "leader decision note 8125 of card C-438 (ref note:438:93249569c03b6a81db69acade2e1161d, 2026-11-24T02:34:19Z, under the user's delegation of 2026-11-23T13:57Z, the leader's judgement and not a selection by the user) answering request note 8123 with option (a), approving revision 1 at head f2f532feed; not approved by the lane", digest: "sha256:5983ebc1c53b7c94248acb95eadb7de6ae2658a3d986ca51bc29dadf558b9582"}
---

# Four test-only seams stop being read by a release build

## Interview

- none: 별도 사용자 인터뷰는 없었어요. 요구사항은 카드 C-438 의 본문과 인수 기준 2개, 그리고 `AUDIT #318` 의 4번 지적이 정했어요.

## Intent

Related: SPEC-20261121-moai-name (D-20, 테스트 전용 환경 변수의 이름 바꾸기). 그 조각이 열다섯 변수의 이름을 `MOAI_TEST_*` 로 바꾸면서 `cfg(debug_assertions)` 없이 읽히는 네 개를 그대로 두었어요(결정 D-20). 다른 열한 개는 디버그 빌드에서만 읽혀요(`src/install.rs`, `src/update.rs`, `src/go_state.rs`, `src/heavy_slot.rs` 등). 이 네 개는 릴리스 바이너리도 읽어요:

- `MOAI_TEST_SESSION_REGISTRY_HOLD_MS` (`src/lane_registry.rs`): 세션 등록의 잠금을 최대 2초 붙잡아요.
- `MOAI_TEST_AGENT_SEQ_HOLD_MS` (`src/session.rs`): `agent-N` 번호 잠금을 최대 2초 붙잡아요.
- `MOAI_TEST_INIT_HOLD_MS` (`src/launcher_init.rs`): 첫 실행 초기화의 잠금을 최대 3초 붙잡아요.
- `MOAI_TEST_INIT_FAIL_AFTER_PROFILE` (`src/launcher_init.rs`): 값이 있으면 첫 실행 설정이 설정 파일을 쓴 직후 오류로 멈춰요.

사용자의 환경에 마지막 변수가 우연히 있으면 첫 실행 설정이 멈추고, 앞의 셋은 이유 없이 느려져요. 그래서 네 읽기를 다른 열한 개와 같은 방식(`#[cfg(debug_assertions)]`)으로 막아요. 같은 감사에서 나온 둘째 점은 옛 접두어 검사가 `tools/*.py` 를 안 보는 것이라, 그 검사의 범위에 `tools/` 의 `.py` 를 더해요.

원 출처: 카드 C-438 본문과 `AUDIT #318` (PASS, 착지됨) 의 4번 지적.

## Non-goals

- 변수 이름과 상수(`TEST_HOLD_VAR` 등)는 그대로 둬요: 통합 테스트가 `moai_cli::...::INIT_TEST_FAIL_VAR` 같은 이름으로 쓰므로 `pub const` 는 릴리스 빌드에도 남아요.
- 다른 열한 개 이음새와 그 테스트는 건드리지 않아요.
- SPEC-20261121-moai-name 안의 문장 하나(`heavy-slot-status` 라고 적은 곳이 실제 테스트에서는 `heavy-slot-acquire` 를 돌려요)는 고치지 않아요. 카드의 인수 기준에 없고, 승인된 SPEC 의 본문을 바꾸면 digest 와 승인이 다시 필요해서 리더 결정이 먼저예요(READY 노트에 남은 일로 적어요).
- 릴리스 프로필로 전체 테스트를 도는 일은 하지 않아요: 이 SPEC 은 릴리스 빌드가 읽지 않는다는 것을 소스 검사와 `cfg!` 로 가르는 테스트로 보여요.

## Requirements

- REQ-1: The launcher shall read `MOAI_TEST_SESSION_REGISTRY_HOLD_MS`, `MOAI_TEST_AGENT_SEQ_HOLD_MS`, `MOAI_TEST_INIT_HOLD_MS` and `MOAI_TEST_INIT_FAIL_AFTER_PROFILE` only in a debug build · failure: a debug build ignores one of them and the concurrency and interruption tests no longer hold the window they were written to hold.
- REQ-2 (REQ-1): IF a release build runs with one of the four variables set, THEN the launcher shall behave as if that variable were unset · failure: a release binary pauses on a hold variable or stops first-launch setup after the settings.
- REQ-3: When the repository test set runs, the earlier test seam prefix scan shall also read the `.py` files under `tools/` · failure: none — a `tools/` folder that is missing is read as an empty one.
- REQ-4: The gate test shall fail and name the file and line of every read of the four variables that is not directly under `#[cfg(debug_assertions)]` · failure: an ungated read of a seam variable reaches a release build with no failing test.

## Acceptance

- AC-1 (REQ-1, REQ-4): 네 변수의 읽기 네 곳(`src/lane_registry.rs` 하나, `src/session.rs` 하나, `src/launcher_init.rs` 둘)이 모두 `#[cfg(debug_assertions)]` 바로 아래에 있고, 한 곳이라도 그 줄이 없으면 `the_four_test_seams_are_read_only_in_debug_builds` 가 파일과 줄을 대며 실패해요. 처음에는 RED(현재 코드는 네 읽기 모두 막혀 있지 않아요) · verify: `cargo nextest run --test state_folder_literal_scan -E 'test(the_four_test_seams_are_read_only_in_debug_builds)'`
- AC-2 (REQ-2): 첫 실행에 `MOAI_TEST_INIT_FAIL_AFTER_PROFILE=1` 을 주면 디버그 빌드는 종료 코드 1 로 멈추고 `CLAUDE.md` 가 없고, 릴리스 빌드는 종료 코드 0 으로 끝나 `CLAUDE.md` 가 있어요(테스트가 `cfg!(debug_assertions)` 로 기대를 가르고, 디버그 쪽은 기존 중단 테스트가 이미 확인해요) · verify: `cargo nextest run --test launcher_auto_init -E 'test(the_init_fail_seam_is_honoured_only_by_a_debug_build)'`
- AC-3 (REQ-3): 임시 루트에 심은 `tools/x.py` 의 옛 접두어 줄은 히트로 잡히고 `tools/x.rs`·`tools/notes.md` 는 잡히지 않으며, 저장소 전체 검사는 통과해요. 처음에는 RED(`tools/` 를 읽지 않아요) · verify: `cargo nextest run --test state_folder_literal_scan -E 'test(the_seam_scan_finds_a_planted_line_and_skips_what_it_should) | test(no_source_or_test_spells_the_earlier_test_seam_prefix)'`
- AC-4 (REQ-1): 기존 동시성·중단 테스트(`a_launch_waits_for_a_first_launch_that_is_still_initializing`, `an_interrupted_first_launch_is_finished_by_the_next_launch`, `session_registry_cli` 와 `launcher_agent_numbering` 의 동시 실행 테스트)가 디버그 빌드에서 그대로 통과해요 · verify: `cargo nextest run --test launcher_auto_init --test session_registry_cli --test launcher_agent_numbering`

## Scope

- write: `.moai/specs/SPEC-20261124-test-seams-release/spec.md`
- write: `src/lane_registry.rs`, `src/session.rs`, `src/launcher_init.rs` (읽기 네 곳의 `#[cfg(debug_assertions)]`)
- write: `tests/state_folder_literal_scan.rs` (게이트 테스트와 `tools/` 검사), `tests/launcher_auto_init.rs` (릴리스 기대 테스트, 중단 테스트의 릴리스 건너뛰기)
- write (줄 임대): `tools/repo_scans.json` 의 항목 `earlier-test-seam-prefix-in-src-and-tests` 의 `scope` 한 줄 묶음
- read-only: `docs/rust-product-path.md`, `.moai/specs/SPEC-20261121-moai-name/spec.md`
- forbidden: every other path, `CLAUDE.md` included

## Plan

1. SPEC 초안을 커밋하고 `LANE-BLOCKED kind=decision` 노트로 리더 승인을 요청해요(스스로 승인하지 않아요).
2. RED: 게이트 테스트와 `tools/` 심기 검사, 릴리스 기대 테스트를 써서 실패(또는 디버그에서 이미 있는 동작의 고정)를 관찰해요.
3. GREEN: 네 읽기에 `#[cfg(debug_assertions)]` 를 붙이고, `seam_scan` 이 `tools` 의 `.py` 를 읽게 하고, 줄 임대로 `tools/repo_scans.json` 의 범위를 맞춰요.
4. 가드 뮤턴트: 읽기 한 곳의 `#[cfg(debug_assertions)]` 를 빼 게이트 테스트가 실패(exit 100)하는지 보고, `tools` 항목을 지워 심기 검사가 실패하는지 봐요. 각각 되돌려요.
5. `moai check affected` 를 맨몸으로 한 번 돌리고 READY 에 영수증을 실어요.

## Risks

- `#[cfg(debug_assertions)]` 를 `if let` 문 앞에 붙이면 릴리스 빌드에서 `ms` 와 `thread::sleep` 이 사라져 미사용 경고가 생길 수 있어요: 읽기와 쓰임이 한 블록 안이라 같이 사라지지만, 릴리스 프로필의 경고는 디버그 테스트로는 안 보여요. `moai check affected` 의 clippy 단계는 디버그 프로필이라, 릴리스 프로필 확인은 `cargo check --release` 를 따로 한 번 돌려요(Result 에 명령과 결과를 적어요).
- 게이트 테스트는 텍스트 검사라서 읽기를 다른 함수로 옮기면 놓칠 수 있어요: 읽기 네 곳의 개수를 함께 확인해 읽기가 없어지거나 옮겨지면 실패해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 읽기 앞에 `#[cfg(debug_assertions)]` 를 붙이는 기존 방식을 써요(`src/install.rs` 의 `probe_timeout` 과 같음), 함수 쌍으로 나누지 않아요 | 카드 본문이 "cfg(debug_assertions) 나 같은 기존 방식"을 정했고, 다른 열한 개 중 여럿이 이 방식이에요 | 읽기 앞의 속성 한 줄씩을 지워요 |
| D-2 | `pub const` 이름은 릴리스 빌드에도 남겨요 | 통합 테스트가 `moai_cli::launcher_init::INIT_TEST_FAIL_VAR` 같은 경로를 써요 | 상수를 `#[cfg(debug_assertions)]` 로 막고 테스트를 문자열로 바꿔요 |
| D-3 | `heavy-slot-status` 문장 수정은 이 SPEC 에서 하지 않아요 | 카드의 인수 기준에 없고 승인된 SPEC 의 digest 가 바뀌어요 | 리더가 별도 카드로 SPEC-20261121-moai-name 의 개정을 열어요 |

## Result

코드 커밋 14d309621d(테스트 먼저 8840c63af1), 2026-11-24 에 그 트리로 빌드해서 돌렸어요. 돌린 뒤 트레일러를 한 블록으로 모으려고 커밋 메시지만 다시 만들었고(옛 해시 f2f532feed·854d32918d·c29f488237·2c5b0c7456·5f241d1ae5, 트리는 `80316df5bd` 로 같고 `git diff` 가 비었어요), 아래 숫자는 옛 해시에서 같은 트리로 돌린 것이에요(로그는 `target/test-seams-release/`, 카드가 열려 있는 동안만 남아요). 모든 cargo 실행은 `moai heavy-slot-acquire --name test-seams-release --max 5` 안에서 `RUSTC_WRAPPER=sccache cargo nextest run --offline --locked` 로 했어요.

- AC-1: met. RED 먼저: 테스트만 커밋하고 코드는 그대로 둔 채 `--test state_folder_literal_scan` 을 돌렸어요. 10개 중 8개 통과, 2개 실패: `the_four_test_seams_are_read_only_in_debug_builds` 가 네 읽기(`src/lane_registry.rs`, `src/session.rs`, `src/launcher_init.rs` 둘)를 모두 짚었고, 다른 하나는 새 테스트가 상태 폴더 이름 리터럴을 한 번 더 써서 기준선이 18 대 17 이 된 것이라 그 줄을 지웠어요(exit 100). GREEN: 네 읽기 앞에 `#[cfg(debug_assertions)]` 를 붙인 뒤 네 테스트 파일(`state_folder_literal_scan`, `launcher_auto_init`, `session_registry_cli`, `launcher_agent_numbering`)을 돌려 88개 중 88개 통과(exit 0). 가드 뮤턴트 M1(`INIT_TEST_FAIL_VAR` 읽기 위의 `#[cfg(debug_assertions)]` 한 줄만 뺌, `git diff --numstat` 0 1): 게이트 테스트 실패, exit 100; `git checkout -- src/launcher_init.rs` 뒤 `git status --short` 에 변경이 없었어요.
- AC-2: met. 위험 경로를 릴리스 프로필로 직접 돌렸어요(`cargo nextest run --release`, 대신 쓴 단위 테스트가 아니라 실제 릴리스 바이너리): 같은 뮤턴트(cfg 한 줄 뺌)로는 `the_init_fail_seam_is_honoured_only_by_a_debug_build` 가 실패(exit 100, 릴리스 바이너리가 변수를 읽고 종료 코드 1 로 멈춰요), 커밋된 코드 head(14d309621d, 옛 해시 2c5b0c7456)로는 이 테스트와 `an_interrupted_first_launch_is_finished_by_the_next_launch`(릴리스에서는 건너뜀) 둘 다 통과(exit 0)해서 릴리스 빌드가 종료 코드 0 으로 끝나고 `CLAUDE.md` 가 생겨요. 릴리스 빌드 로그에 경고가 없었어요. 디버그 쪽은 위 88개 안의 기존 중단 테스트(`an_interrupted_first_launch_is_finished_by_the_next_launch`)가 확인해요. 나머지 세 변수(대기 시간 변수)는 릴리스 바이너리로 시간을 재어 보지 않았고, 소스 게이트 테스트(AC-1)가 읽기 위치로 보여요(not_run: 릴리스 바이너리의 실제 대기 시간).
- AC-3: met. 심기 테스트(`tools/x.py` 의 옛 접두어 줄이 `tools/x.py:2` 로 잡히고 `tools/x.rs`·`tools/notes.md` 는 안 잡혀요)와 저장소 전체 검사(`no_source_or_test_spells_the_earlier_test_seam_prefix`)가 통과해요(위 88개 안). RED 순서는 지키지 못했어요: `seam_scan` 변경과 심기 테스트를 같은 편집에 썼어요. 대신 가드 뮤턴트 M2(`("tools", &["py"][..]),` 한 줄만 뺌, `git diff --numstat` 0 1)가 심기 테스트를 실패(exit 100)시키고, 되돌린 뒤 트리가 깨끗했어요. `tools/repo_scans.json` 은 줄 임대로 고쳤어요: 항목 `earlier-test-seam-prefix-in-src-and-tests` 의 범위에 `tools/` 의 `.py` 를 더하고, 게이트 테스트를 `test-seams-read-only-in-debug-builds` 항목으로 올렸어요(JSON 이 읽히는 것까지는 확인했고 `check affected` 결과는 READY 노트에 있어요).
- AC-4: met. AC-1 의 GREEN 실행(88개, exit 0)에 `launcher_auto_init`, `session_registry_cli`, `launcher_agent_numbering` 의 기존 동시성·중단 테스트가 모두 들어 있어요.
- 안 돌린 것: 전체 `cargo test` 는 리더의 몫이라 돌리지 않았어요. `check affected` 영수증은 이 Result 를 쓸 때까지 돌리지 않았고 카드의 READY 노트에 있어요. Windows·Linux 결과는 말하지 않아요.
- 남은 일(카드 본문 둘째 점): `SPEC-20261121-moai-name` 이 `heavy-slot-status` 라고 적은 곳은 실제 테스트가 `heavy-slot-acquire` 를 돌려요. 승인된 SPEC 의 digest 가 바뀌어서 이 카드에서 고치지 않았어요(결정 D-3, 리더 노트 8125 확인 2).
