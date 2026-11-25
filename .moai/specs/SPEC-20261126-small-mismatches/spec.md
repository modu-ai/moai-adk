---
id: SPEC-20261126-small-mismatches
title: "Three small mismatches: a session-only trailer line, shipped texts that name only real commands, and a slot whose holder is gone"
status: draft
revision: 1
created: 2026-11-26
branch: spec/small-mismatches-20261126
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

- AC-1 (REQ-1): `moai trailers --no-spec` prints a `Moai-Session` line and no `Moai-Spec` line, and `--verified tests` adds the `Moai-Verified` line · verify: `cargo test --offline --locked --test small_mismatches ac1_`
- AC-2 (REQ-2, REQ-3): a call with none of `--spec`, `--no-spec`, `--list` is refused with a message that names `--no-spec`, and `--no-spec` with `--spec` or `--list` is refused · verify: `cargo test --offline --locked --test small_mismatches ac2_`
- AC-3 (REQ-4): the `trailers --help` text names `--no-spec` and says it is for a card with no SPEC · verify: `cargo test --offline --locked --test small_mismatches ac3_`
- AC-4 (REQ-5): every `moai` command named in an inline code span of `skills/moai/**/*.md`, `agents/*.md`, `CLAUDE.md`, `docs/lane-rules.md` and `docs/leader-rules.md` is a command of this build, a command with a subcommand group has the named subcommand, and the scan names the first text that cites a command the build does not have · verify: `cargo test --offline --locked --test small_mismatches ac4_`
- AC-5 (REQ-6): a slot file that records a dead pid and that no process holds is taken by the next acquire at once, and `heavy-slot-status` lists it as free · verify: `cargo test --offline --locked --test small_mismatches ac5_`
- AC-6 (REQ-6): a slot whose wrapper was killed by SIGKILL keeps held while its command runs and is free when the command ends (the landed tests) · verify: `cargo test --offline --locked --test heavy_slot_cli release_ac`

## Scope

- write: `.moai/specs/SPEC-20261126-small-mismatches/**`, `src/ops/trailers.rs`, `src/trailers.rs`, `tests/small_mismatches.rs`
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

(구현 뒤에 채워요.)
