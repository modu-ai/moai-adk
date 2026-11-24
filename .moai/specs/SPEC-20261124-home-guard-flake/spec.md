---
id: SPEC-20261124-home-guard-flake
title: "HOME 가드가 부하 아래에서 테스트가 읽지 않은 HOME 읽기를 테스트 탓으로 실패시키는 일을, 먼저 증거로 가려내요"
status: draft
revision: 1
created: 2026-11-24
---

# HOME 가드가 부하 아래에서 테스트가 읽지 않은 HOME 읽기를 테스트 탓으로 실패시키는 일을, 먼저 증거로 가려내요

## Interview

- none: 리더 배차(카드 본문)가 요구를 정했어요 — 가드를 느슨하게 하지 않고, 진짜 HOME 쓰기는 계속 실패시키며, 원인을 재현으로 찾아요. 사용자 인터뷰는 없었어요.

## Intent

`tools/test_home_guard.sh`(SPEC-20261123-test-home-isolation)가 통과한 테스트를 `HOME-GUARD: touched its HOME` 으로 실패시키는 일이 2026-11-24 에 세 레인에서 다시 나왔어요(카드 노트 8260, 8267, 8269). 같은 테스트는 혼자 돌리거나 같은 head 로 한 번 더 돌리면 통과해요. 레인마다 `check affected` 를 한 번 더 돌려 넘기고 있어서, 레인 수만큼 시간이 새요.

지금 아는 것과 모르는 것을 나눠 적어요. 읽은 명령과 숫자는 이 SPEC 의 Result 에 옮겨요(`target/` 의 로그는 카드가 열려 있는 동안만 남아요).

아는 것 (이 실행에서 직접 읽고 잰 것):

- 실제 실패 기록 하나가 남아 있어요: 다른 레인의 `check affected` 로그 `05-test.log` 의 두 줄이에요. 두 줄 모두 찍힌 항목이 정확히 decoy 폴더, `.moai`, `.moai/user-assets.json` 셋뿐이에요. `language.yaml`, `settings.json`, `.claude` 는 찍히지 않았어요.
  - `language_setting` 의 `the_rerun_keeps_the_stored_conversation_language` (0.448 s)
  - `review_section_cli` 의 `review_md_writes_default_sourced_decisions_but_not_hold`
- 두 테스트 모두 기준선(`baseline.tsv`)에 없고, 두 테스트 파일은 `moai` 자식마다 `.env("HOME", <자기 임시 폴더>)` 를 줘요(`tests/language_setting.rs` 의 `run`, `tests/review_section_cli.rs` 의 `cli`). 그러니 테스트와 그 자식이 가드의 decoy 를 읽을 길이 코드에 없어요. 읽은 쪽은 테스트 바깥이에요.
- `user-assets.json` 을 여는 코드는 `home_state::home_go_markers` 하나예요. `home_root::path` 가 HOME 의 상태 폴더를 정할 때마다 불러요. 읽힌 모양이 그 코드가 남기는 모양(폴더 둘 읽기 + 파일 하나 읽기)과 같아요.
- 바깥 읽기는 아주 드물어요. 한 번의 `check affected` 테스트 단계에서 4083개 중 2개예요.

모르는 것:

- 읽은 쪽이 누구인지. 원인은 uncertain 이에요.

재현 시도는 전부 0건이었어요. 이 숫자는 "없다"가 아니라 "이 방법으로는 못 잡았다"예요.

| 시도 | 횟수 | 부하 평균 | 거짓 실패 |
| --- | --- | --- | --- |
| 실패한 테스트 한 개를 가드(observe)로 반복, 병렬 8 | 40 | 44 | 0 |
| 아무것도 안 하는 프로세스(8 s)를 가드로 반복, 병렬 20 | 60 | 37~45 | 0 |
| 가드와 같은 모양의 decoy 800개를 20 s 두고 비교(`.metadata_never_index` 없음) | 800 | 42~105 | 0 |
| 같은 모양 800개, `.metadata_never_index` 있음 | 800 | 42~105 | 0 |
| 이전 레인: 아무것도 안 하는 프로세스를 가드로 반복 | 300 | - | 0 |
| 이전 레인: 실패 이력이 있는 바이너리 5개(59개 테스트)를 nextest 모양으로 | 295 | 65~95 | 0 |

배제한 후보 (각각 근거가 있어요):

- Spotlight 같은 시스템 색인기: 가만히 둔 decoy 1600개에 읽기 0건, 마커 유무와 무관해요.
- 테스트 자신이나 그 `moai` 자식: 위에 적은 대로 코드에 읽을 길이 없고, 실패한 테스트를 40번 단독으로 돌려도 읽지 않아요.
- `hygiene.rs`/`worktree_report.rs` 의 임시 폴더 훑기: 최상위만 보거나 크기만 재고, `user-assets.json` 을 열지 않아요.
- `ps` 로 프로세스 표를 읽는 `home_state::process_command_lines`: 명령줄 인자만 읽고 환경 변수 HOME 은 안 읽어요.

여기서 코드를 고치면 원인이 아닌 곳을 고치게 돼요. 거짓 실패가 한 번 더 나올 때 그 한 번이 원인을 알려주도록, 증거를 남기는 장치를 먼저 만들어요. 가드의 판정(통과·실패·종료 코드)은 바꾸지 않아요.

## Non-goals

- 가드를 느슨하게 하는 일. 진짜 HOME 읽기·쓰기·삭제는 계속 exit 1 이에요. 판정 규칙, 기준선, `enforce` 기본값은 이 SPEC 이 건드리지 않아요.
- 실패한 테스트를 가드가 다시 돌리는 일(다시 돌리면 부작용이 있는 테스트가 있어요).
- 원인을 모른 채 읽기를 눈감아 주는 규칙(예: 폴더만 찍히면 통과). 이 SPEC 의 증거가 원인을 가리킨 뒤 SPEC 개정(revision 2)이나 새 SPEC 으로 다뤄요.
- Windows 호스트. 가드가 `sh` 를 쓰므로 unix 만이에요.

## Requirements

<!-- This SPEC adds evidence only and no blocking check: the guard's verdict stays as it is, so there is no observe step to record. -->
- REQ-1: When the wrapper prints a `HOME-GUARD: touched its HOME` line, it shall append one `HOME-GUARD-DETAIL:` line per hit entry to the file named by `MOAI_HOME_GUARD_LOG`, with the entry's access time and modification time in seconds with a fraction.
- REQ-2 (REQ-1): The `HOME-GUARD-DETAIL:` lines shall carry the time the wrapper started the test and the time it compared the HOME, in seconds with a fraction, so a read can be placed before the test, during it or after it.
- REQ-3 (REQ-1): The `HOME-GUARD-DETAIL:` lines shall carry, for each other folder named `moai-test-home.*` under `TMPDIR`, the access time of its `.moai/user-assets.json`, read with `stat` only.
- REQ-4: The wrapper shall never list, open or read any entry inside another guard folder.
- REQ-5: The wrapper shall keep every exit code and every `HOME-GUARD:` line it printed before this change.
- REQ-6 (REQ-1): IF a `HOME-GUARD-DETAIL:` line cannot be written, THEN the wrapper shall ignore the failure and keep the test's result.
- REQ-7: Where `perl` is not on the host, the wrapper shall write the times in whole seconds.

## Acceptance

- AC-1 (REQ-1, REQ-2): a test that reads the decoy `user-assets.json` under a log file gives a `HOME-GUARD:` line and one `HOME-GUARD-DETAIL:` line for each hit entry, each with an access time and a modification time that has a fraction, and the test's start time is not later than the comparison time. · verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-2 (REQ-3, REQ-4): with another guard folder planted beside the test's, the detail line names it with its `user-assets.json` access time, and the planted folder's own access times are the same after the wrapper ran as before. · verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-3 (REQ-5): every test of `tests/test_home_guard.rs` that existed before this change passes unchanged: a decoy read, a write and a removal still exit 1 in the default mode, observe keeps the test's exit code, a listed test is tolerated as before. · verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-4 (REQ-6): a log path that cannot be written changes neither the exit code nor the printed `HOME-GUARD:` line. · verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-5 (REQ-7): with `perl` hidden from `PATH` the detail line holds whole seconds and the verdict is the same. · verify: `cargo nextest run --locked --offline --test test_home_guard`

## Scope

- write: `tools/test_home_guard.sh`, `tests/test_home_guard.rs`, `.moai/specs/SPEC-20261124-home-guard-flake/spec.md`
- read-only: `src/`, `tests/fixtures/test_home_guard/baseline.tsv`, `.config/nextest.toml`
- forbidden: the baseline's lines (it never grows), the `enforce` default, the decoy files and their old access time, any rule that excuses a hit

## Plan

1. RED: the five acceptance tests in `tests/test_home_guard.rs` fail on the current wrapper.
2. `tools/test_home_guard.sh` records the start time and, after the verdict, writes the detail lines; the verdict code is not touched.
3. Guard mutants, one line each in the READY note: removing the detail write fails AC-1; reading a planted peer with `ls` or `find` instead of `stat` fails AC-2; changing an exit code fails AC-3.
4. `moai check affected` once at the head; the receipt goes in the READY note.
5. Later, with no new card: the next real `touched its HOME` that is not a real read leaves the detail lines in the log; the reader is read from them and revision 2 fixes the source.

## Risks

- A planted peer check that opens a file would change the peer's access time and make the guard cause the very flake it measures. REQ-4 and AC-2 pin that.
- The detail lines say who did not read, not who did. If the instants of several guard folders cluster at one moment the reader swept the temporary folder; if only one folder moved the reader knew its path. Neither names a program; it narrows the search.
- `perl` is not on every host: REQ-7 keeps the lines but with coarser times.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Evidence first: this revision changes no verdict. | proposed by the lane, awaiting the leader (LANE-BLOCKED kind=decision); default if the leader does not answer: this option | drop the detail lines |
| D-2 | A control folder (an unreachable twin of the decoy whose read marks a hit as an outsider's) is not part of this revision. | proposed: it would change what passes when a real read and an outside read coincide, a contract change that needs the cause first | add it in revision 2 after the evidence names the reader |
