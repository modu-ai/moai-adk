---
id: SPEC-20261126-small-mismatches
title: "Three small mismatches: a session-only trailer line, shipped texts that name only real commands, and a slot whose holder is gone"
status: approved
revision: 1
created: 2026-11-26
branch: spec/small-mismatches-20261126
approval:
  by: leader
  at: 2026-11-26
  basis: "리더 위임 승인 노트 9663 (사용자 위임 2026-11-25 ~14:50Z \"나머진 네가 알아서 추천 해주는데로 선택해서\", 리더 판단이고 사용자 직접 선택은 아님); 리더가 착지 직전 moai spec show 로 digest 를 다시 읽었어요"
  digest: "sha256:d5993a747fef86b28142dcbbe2cf4897fd575a5a77ff011a0f7f2294946fa1e8"
---

# Three small mismatches: a session-only trailer line, shipped texts that name only real commands, and a slot whose holder is gone

## Interview

- Q1: 무엇을 고치나요? → A: 이 카드의 본문이 옮긴 다른 저장소 리더 세션의 개선 요청 셋(요청 7, 8, 9)이에요. 사용자 말이 아니라 동료 세션의 보고예요(노트 8834).
- Q2: 어디까지 하나요? → A: 리더의 배정 글: 셋 가운데 tip 에서 이미 고쳐진 것은 증거를 붙여 met 로 찍고, heavy slot 은 죽은 pid 를 알아보고 비우기까지만 해요. 프로세스를 죽이는 일은 하지 않아요.

## Intent

세 가지가 어긋나 있다고 보고됐어요(요청 7, 8, 9). 보고한 저장소는 이 저장소가 아니라서 NeXT `9c8b7fb763` 에서 하나씩 다시 읽었어요.

- 7. `moai trailers` 는 `--spec` 이 없으면 거절해요(`trailers needs --spec …`). SPEC 없는 카드 작업에서는 `Moai-Session` 줄만 꺼낼 길이 없어요. 아직 열려 있어요.
- 8. 글이 부르는데 이 빌드에 없는 명령(`moai worktree …`, `moai version`)은 지금의 `skills/**`, `agents/**`, `CLAUDE.md`, `docs/lane-rules.md`, `docs/leader-rules.md` 에서 찾지 못했어요(`moai card hold`, `unhold` 는 있어요). 다만 글 전체를 훑어 부르는 명령이 실제로 있는지 보는 시험은 읽기 글 네 개와 감사 글 둘에만 있어요(`tests/e2e_temp_home_cli.rs`, `tests/audit_text_facts.rs`). 그래서 다시 어긋나도 잡히지 않아요.
- 9. 죽은 보유자가 쥔 heavy slot 은 `SPEC-20261125-heavy-slot-release-on-exit` 가 이미 고쳤어요. 래퍼가 끝나면 자기 슬롯을 풀고, SIGKILL 로 죽은 래퍼의 슬롯은 그 명령과 후손이 모두 끝나면 운영체제가 풀며, `heavy-slot-status` 는 죽은 pid 를 `alive: false`, 그 슬롯을 열고 있는 프로세스(`holders`)와 경고 한 줄로 보여요. 후손이 슬롯 파일을 여는 한 슬롯은 정말 잡혀 있어서, 그 프로세스를 끝내는 것은 주인의 몫이고 이 SPEC 은 하지 않아요.

읽는 사람: 이 변경을 판단하는 리더, 구현하는 레인, 독립 감사자.

## Non-goals

- 프로세스를 죽이거나 슬롯 파일을 손으로 지우는 일은 하지 않아요.
- 보고서 B(`card prereq` 의 선행 관계 지우기)와 C(카드 명령의 프로젝트 인자 모양)는 이 카드에 넣지 않아요. 둘 다 따로 카드를 받아야 하는 크기이고 노트 8834 가 레인이 정하라고 했어요.
- `moai trailers --spec` 의 거절, `moai trailer-check`, 게이트는 느슨해지지 않아요.

## Requirements

- REQ-1: When `moai trailers` is given `--no-spec`, the command shall print the `Moai-Session` line and no `Moai-Spec` line. · failure: none — without a host session id or outside a git worktree the result says so and prints no trailer line, as `--spec` does today.
- REQ-2: IF `moai trailers` is given neither `--spec`, `--no-spec` nor `--list`, THEN the command shall refuse with one message that names `--no-spec`.
- REQ-3: IF `--no-spec` is given with `--spec` or with `--list`, THEN the command shall refuse and print nothing.
- REQ-4: The `trailers` help shall say that `--no-spec` is the way to print the session line of a card with no SPEC.
- REQ-5: The repository shall carry a scan test that fails when a `moai` command named in the shipped texts is not a command of this build.
- REQ-6: When the recorded pid of a slot is dead and no process has the slot file open, the next `heavy-slot-acquire` shall take that slot without waiting. · failure: none — a slot that a live process still holds stays held and is never taken.

## Acceptance

- AC-1 (REQ-1): `moai trailers --no-spec` prints a `Moai-Session` line and no `Moai-Spec` line, and `--verified tests` adds the `Moai-Verified` line · verify: `cargo test --offline --locked --test repo_small_mismatches ac1_`
- AC-2 (REQ-2, REQ-3): a call with none of `--spec`, `--no-spec`, `--list` is refused with a message that names `--no-spec`, and `--no-spec` with `--spec` or `--list` is refused · verify: `cargo test --offline --locked --test repo_small_mismatches ac2_`
- AC-3 (REQ-4): the `trailers --help` text names `--no-spec` and says it is for a card with no SPEC · verify: `cargo test --offline --locked --test repo_small_mismatches ac3_`
- AC-4 (REQ-5): every `moai` command named in an inline code span of `skills/moai/**/*.md`, `agents/*.md`, `CLAUDE.md`, `docs/lane-rules.md` and `docs/leader-rules.md` is a command of this build, a command with a subcommand group has the named subcommand, and the scan names the first text that cites a command the build does not have · verify: `cargo test --offline --locked --test repo_small_mismatches ac4_`
- AC-5 (REQ-6): a slot file that records a dead pid and that no process holds is taken by the next acquire at once, and `heavy-slot-status` lists it as free · verify: `cargo test --offline --locked --test repo_small_mismatches ac5_`
- AC-6 (REQ-6): a slot whose wrapper was killed by SIGKILL keeps held while its command runs and is free when the command ends (the landed tests) · verify: `cargo test --offline --locked --test heavy_slot_cli release_ac`

## Scope

- write: `.moai/specs/SPEC-20261126-small-mismatches/**`, `src/ops/trailers.rs`, `src/trailers.rs`, `tests/repo_small_mismatches.rs`
- write (when the scan finds one): the shipped text that names a command the build does not have
- write (한 줄, line-lease): `tools/repo_scans.json`, `skills/moai/references/source-manifest.json`
- read-only: `src/heavy_slot.rs`, `tests/heavy_slot_cli.rs`, `tests/e2e_temp_home_cli.rs`, `tests/audit_text_facts.rs`
- forbidden: `CLAUDE.md` 의 규칙 변경, 게이트 완화, 실제 HOME, 네트워크, Python

## Decisions

- D-1: 세션 줄만 내는 길은 새 플래그 `--no-spec` 이에요. `--spec` 을 빼먹었을 때 조용히 `Moai-Spec` 줄이 사라지지 않게 하려고, 아무 플래그도 없는 호출은 지금처럼 거절하고 그 메시지가 `--no-spec` 을 알려요.
- D-2: 스캔은 만든 바이너리에 `moai <word> --help` 를 물어 명령이 있는지 봐요(`tests/e2e_temp_home_cli.rs` 의 방식과 같아요). 부분집합의 글만 보지 않고 위 AC-4 의 글 전부를 훑어요.
- D-3: heavy slot 은 코드를 바꾸지 않아요. 이미 있는 동작을 시험으로 못박고(AC-5), 후손이 잡고 있는 슬롯은 정말 잡혀 있다는 것을 SPEC 에 적어요.

## Plan

1. RED: `tests/small_mismatches.rs` 에 ac1_ … ac5_ 를 더해 실패를 기록해요.
2. GREEN: `--no-spec` 를 `src/ops/trailers.rs` 와 `src/trailers.rs` 에 더하고, 스캔이 찾은 글을 고쳐요.
3. 영수증: `cargo fmt --all`, `check affected --card`, READY.

## Risks

- R-1: 스캔이 처음 돌 때 글이 부르는 명령 가운데 이 빌드에 없는 것이 나올 수 있어요. 나오면 그 글을 이 카드의 범위에서 고쳐요(글만 고치고 명령은 만들지 않아요).
- R-2: 스캔은 만든 바이너리를 여러 번 불러서 시간이 걸려요. 명령 이름마다 한 번만 물어요.

## Result

구현 커밋 `bb46e9309b` 는 RED 커밋 `032a6271ea` 위에 있고 NeXT `9c8b7fb763` 에서 갈라졌어요. 아래는 이 실행에서 본 값이에요.

- RED: 구현 전 `cargo test --offline --locked --test repo_small_mismatches` 는 6개 가운데 3개가 실패했어요(ac1_, ac2_, ac3_, `--no-spec` 가 없었어요). 스캔(ac4_)과 죽은 pid 슬롯(ac5_)은 처음부터 통과했어요: 둘째와 셋째 어긋남은 tip 에서 이미 고쳐져 있었어요(기록 `target/small/red.log`).
- GREEN: 같은 실행이 6개 통과, 0개 실패예요(`target/small/green.log`).
- AC-1: met — `ac1_no_spec_prints_the_session_line_only`
- AC-2: met — `ac2_a_bare_call_names_no_spec_and_the_flag_conflicts_are_refused`
- AC-3: met — `ac3_the_help_says_no_spec_is_for_a_card_with_no_spec`
- AC-4: met — `ac4_every_command_the_shipped_texts_name_is_a_command_of_this_build`가 `skills/moai/**`, `agents/*.md`, `CLAUDE.md`, `docs/lane-rules.md`, `docs/leader-rules.md` 의 부르는 명령을 모두 만든 바이너리에 물어 없는 것이 하나도 없음을 읽어요. `ac4_the_scan_finds_a_command_the_build_does_not_have` 가 스캔의 읽는 법(말글·자리표시자는 명령이 아님, `moai version` 같은 없는 명령은 못 찾음)을 고정해요. 글을 고칠 일은 없었어요.
- AC-5: met — `ac5_a_slot_of_a_dead_pid_that_nobody_holds_is_taken_by_the_next_acquire`
- AC-6: met — `cargo test --offline --locked --test heavy_slot_cli release_ac` 가 6개 통과, 0개 실패(`target/small/slot.log`). SIGKILL 로 죽은 래퍼의 슬롯은 명령이 도는 동안 held 이고 끝나면 풀리며, 후손이 슬롯을 열고 있으면 `holders` 와 경고로 보여요.
- 코드를 고친 곳은 `--no-spec` 하나예요(`src/trailers.rs`, `src/ops/trailers.rs`). heavy slot 은 코드를 바꾸지 않았고 프로세스를 죽이는 일도 하지 않았어요.
- 보고서의 B(`card prereq` 의 선행 관계 지우기)와 C(카드 명령의 프로젝트 인자 모양)는 이 카드에 넣지 않았어요(크기가 달라 따로 카드가 필요해요).
- 남은 것: 독립 감사(tier 1)와 넓은 `check affected` 는 아직이에요. Windows 와 Linux 는 Mac 에서만 돌렸어요. 보고한 저장소(다른 저장소)의 환경은 재현하지 않았어요.
