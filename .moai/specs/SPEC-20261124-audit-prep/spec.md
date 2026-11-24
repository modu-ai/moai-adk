---
id: SPEC-20261124-audit-prep
title: "감사가 카드 이름으로 자리를 잡고 head 에서 증분 빌드 한 번 하고 깨끗하게 끝나요"
status: approved
revision: 2
created: 2026-11-24
approval: {by: leader, at: 2026-11-23, basis: "leader decision note 7978 of card C-364 (2026-11-23T22:59:29Z, delegation of the user 2026-11-23T13:57Z, the leader's judgement and not a selection by the user) answering request note 7976 with option (a) for revision 2 at commit 163c2571d1b24fadc05ec2232fcec492b29393c9, on the condition that the whole audit_slot_cli passes on the same head receipt", digest: "sha256:da5fd1d18591333fb17e5eb397bada1130b9883d4a783da7093760d6f0b079c7"}
---

# 감사가 카드 이름으로 자리를 잡고 head 에서 증분 빌드 한 번 하고 깨끗하게 끝나요

## Interview

- none: 리더 배차(카드 C-364 와 리더 결정 노트 7953)가 요구사항을 전달했어요. 초안에 대한 리더 답이 노트 7966 이고(선택지 (b)), 별도 사용자 인터뷰는 없었어요.

## Intent

카드 C-364 와 리더 결정(2026-11-23T22:22Z, 사용자 위임 2026-11-23T13:57Z)이에요. 사용자가 감사 비용을 다시 짚었어요. 레인이 같은 head 를 이미 빌드했는데도 감사마다 빈 빌드 폴더에서 처음부터 다시 빌드했어요. 근거는 재설계 `~/.moai/reports/moai-adk/agentic-redesign-2026-11-22.md` §3D 와 점검 `audit-landing-process-review-2026-11-22.md` 의 A1, A2, A4, "감사가 레인 빌드를 이어받는 방법" 이에요.

지금은 리더가 손으로 해요. 감사 자리를 head 로 만들고, 레인 `target/` 을 `cp -c -R` 로 복제하고, 바뀐 대상만 그 `target/` 하나에서 증분 빌드하고, 합친 트리는 빌드하지 않아요. 이 SPEC 은 그 손 규칙을 카드 이름으로 부르는 명령 둘로 만들어요.

- `moai audit-prep <card> --head <40hex>` 는 카드를 위한 감사 자리를 잡고, head 로 detach 하고, 레인(또는 seed)의 빌드를 이어받고, 이름 붙인 테스트 대상을 증분 빌드 한 번 해요.
- `moai audit-done <card>` 는 카드의 자리를 마쳐요. 깨끗할 때만 풀고, 버리거나 강제하지 않아요.

이미 있는 것(NeXT `b21063a45d` 에서 읽음): `moai audit-slot prep|done` 이 `178fde8beb` 로 합쳐졌어요(SPEC-20261123-audit-slot, 감사 PASS 노트 7045). 자리 전환, `cp -c -R -p` 로 레인 `target/` 복제, 추적 소스 touch(레인의 뮤턴트 빌드가 head 로 통과하는 일을 막음), 자리당 holder 한 명, 링크·자기 작업트리 거부, `--discard` 정리를 이미 해요. 없는 것은 셋이에요. 빌드를 돌리지 않고, 자리와 holder 로 부르지 카드로 부르지 않고, 빌드한 테스트 실행 파일이 레인 경로가 아니라 감사 자리 경로를 읽는다는 것을 시험이 보이지 않아요. 이 SPEC 은 그 셋을 `audit-slot` 위에 더하고, 안전 코드를 복사하지 않아요.

읽는 사람: 감사를 배차하는 리더, 감사를 돌리는 감사자, 빌드의 원본이 되는 READY 작업트리의 레인이에요.

## Non-goals

- 카드마다 `.claude/worktrees/audit-<card>` 를 만들고 끝에 지우는 형태예요. 사용자가 승인한 재설계 §3A 는 `audit-1`..`audit-3` 을 고정해 다시 써요(D-1). 리더가 다른 형태를 고르면 코드 전에 이 SPEC 을 고쳐요.
- 컴파일 때 박히는 `env!` 경로를 고치는 일이에요. 지금 `tests/` 의 163개 파일이 `env!("CARGO_MANIFEST_DIR")` 또는 `env!("CARGO_BIN_EXE_*")` 를 읽어요(`df2cffd593` 의 이 워크트리에서 `grep -lE` 로 잰 값이고, 카드 본문은 135개라고 적었어요; 리더 노트 7966 이 163 을 기록해요). 그대로 두고, 다시 빌드한 테스트 실행 파일이 레인이 아니라 감사 자리를 읽는다는 것만 REQ-9 로 보여요.
- 레인과 감사가 nextest 아카이브를 공유하는 일이에요. 그 경로를 실행 시점에 읽게 바꾼 뒤에 해요.
- 판정 때 증거 로그를 복사하는 일, 자리를 지우는 일이에요. 자리는 고정이에요.
- 감사의 테스트나 뮤턴트를 돌리는 일이에요. `audit-prep` 은 빌드만 하고 돌리는 쪽은 감사자예요.
- head 의 `check affected` 를 다시 돌리는 일이에요. 감사자는 레인의 영수증을 인용해요.
- 이 저장소 밖의 호스트 변경이에요. `~/.claude` 와 `~/.moai` 의 실제 데이터는 건드리지 않아요.

## Requirements

- REQ-1: When `audit-prep` is given a card and a 40-hex head, the audit-prep command shall take a free audit slot with the card as holder and leave it detached at that head.
- REQ-2 (REQ-1): If the card already holds an audit slot, then the audit-prep command shall prepare that same slot again.
- REQ-3 (REQ-1): If every audit slot has a live holder other than the card, then the audit-prep command shall refuse, name every holder and run no build.
- REQ-4 (REQ-1): If the slot cannot be prepared, then the audit-prep command shall refuse with the reason, change nothing more and run no build.
- REQ-5: When the slot is prepared, the audit-prep command shall run one incremental `cargo test --no-run` of the test targets named with `--test`, or of the library when none is named, inside one heavy slot.
- REQ-6 (REQ-5): If more than two test targets are named, then the audit-prep command shall refuse before it changes anything.
- REQ-7: When the build ends, the audit-prep command shall report the units Cargo compiled, the units it found fresh and the exit code of the build. · failure: none — a build that fails is REQ-8, and the report is written in both cases.
- REQ-8 (REQ-5): If the build fails, then the audit-prep command shall exit non-zero, keep the slot held by the card and name the last lines of Cargo's output.
- REQ-9: When the slot's build came from a lane, the audit-prep command shall build every named target from the slot's own sources, so no compile-time path in a rebuilt test executable names the lane's folder. · failure: none — a compile error is REQ-8; a lane path found in a rebuilt executable fails AC-4.
- REQ-10: When `audit-done` is given a card, the audit-done command shall finish the slot that card holds like `audit-slot done` without discard: refuse and name the files when the slot holds a change, release it when it is clean, and keep its `target/`.
- REQ-11 (REQ-10): If the card holds no audit slot, then the audit-done command shall refuse and say so.
- REQ-12: The audit-done command shall accept no option that discards a change or forces a release.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): in a temporary repository with a lane at H, `audit-prep C-1 --head H --lane L` leaves one slot detached at H, held by `C-1`, and answers the slot name; a second call for `C-1` answers the same slot; three other holders make a call for `C-9` refuse naming all three; an unusable lane or a dirty slot is refused and no build runs. · verify: `cargo nextest run --locked --offline --test audit_prep_cli`
- AC-2 (REQ-5, REQ-6, REQ-7): a temporary crate with a dependency outside the repository, built in the lane and prepared into a slot, builds with one `audit-prep` call that names one test target; the answer lists the workspace units as compiled once each and every dependency as fresh; a third `--test` is refused before any change. · verify: `cargo nextest run --locked --offline --test audit_prep_cli`
- AC-3 (REQ-8): with a test target that does not compile, `audit-prep` exits non-zero, the slot is still held by the card, and the answer carries the compiler's last lines. · verify: `cargo nextest run --locked --offline --test audit_prep_cli`
- AC-4 (REQ-9): the lane builds a test executable that prints its `CARGO_MANIFEST_DIR`; after `audit-prep` it is rebuilt in the slot and prints the slot's folder, and the rebuilt executable holds no byte sequence of the lane's folder. · verify: `cargo nextest run --locked --offline --test audit_prep_cli`
- AC-5 (REQ-10, REQ-11, REQ-12): after a mutant changed a tracked file in the slot, `audit-done C-1` refuses, names the file and keeps the hold; once reverted it releases the slot, keeps `target/` and leaves an empty `status --porcelain`; a card without a slot is refused; `--discard` and `--force` are not accepted. · verify: `cargo nextest run --locked --offline --test audit_prep_cli`
- AC-6 (REQ-5): on a prepared slot a one-line mutant in the library recompiles only the workspace units, none of the dependencies; the real repository is not built by any test (no whole `cargo test`). · verify: `cargo nextest run --locked --offline --test audit_prep_cli`

## Scope

- write:
  - `.moai/specs/SPEC-20261124-audit-prep/spec.md`
  - `src/audit_prep.rs` (new)
  - `src/ops/audit_prep.rs` (new), the operation `audit.prep`, subcommand `audit-prep`
  - `src/ops/audit_done.rs` (new), the operation `audit.done`, subcommand `audit-done`
  - `src/lib.rs` (one `pub mod` line)
  - `tests/audit_prep_cli.rs` (new)
  - `docs/design-review/appendix/operations/audit_prep.md` (new)
  - `docs/design-review/appendix/operations/audit_done.md` (new)
  - `src/audit_slot.rs`: only the helpers the card-addressed commands share become `pub(crate)` (the holder lookup by name, the prep and done cores); no change of behavior, except the one the repair of the audit (note 8068, B4) added: `audit-slot prep` refuses a slot whose own `target/` is a link or a file before its first write, with or without `--lane` (`real_target`)
  - `src/help.rs`: the rows and examples, under a one-command line lease
  - `tests/fixtures/help/all.txt`, `all-60.txt`, `all-80.txt`, `all-dumb.txt`: regenerated with `MOAI_HELP_GOLDEN_WRITE=1`, one row each, under a recorded leader decision as for audit-slot
- read-only: everything else
- forbidden: `.github`, `.claude/settings.json`, `.moai/project.yaml`

## Plan

1. 이 SPEC 과 리더 승인(노트 7966). 끝났어요.
2. 코드 경로를 `SCOPE-REQUEST` 로 알려요. 공유 파일 `src/lib.rs`(C-322 가 쥠)와 `src/help.rs` 의 한 줄은 `card line-lease` 로 해요.
3. 테스트 먼저(RED): 임시 저장소와 임시 HOME 에서 `tests/audit_prep_cli.rs` 로 AC-1 ~ AC-6 을 써요.
4. `src/audit_prep.rs` 와 `src/ops/audit_prep.rs`, 도움말 행과 골든, 운영 문서를 써요.
5. 가드 뮤턴트를 돌려요. 새 가드 테스트마다 READY 에 한 줄씩 남겨요.
6. READY 에서 `moai check affected` 를 맨몸으로 한 번 돌리고, 같은 head 에서 `audit_slot_cli` 전체(승인 조건, 노트 7978)가 통과함을 영수증에 적고, head 와 base 를 적어 READY 를 내요.

## Risks

- 새 명령 둘과 `audit-slot` 이 한 일에 대한 명령 면이 둘이 돼요. 새 쪽은 옛 쪽의 함수만 부르고, 운영 문서가 자리 단위 일(리더의 `--discard`)은 `audit-slot` 이라고 알려요.
- 빌드가 실패하면 자리가 잡힌 채로 남아요. 감사자는 그래도 `audit-done` 을 돌려야 하고, 남은 뮤턴트는 일부러 `audit-done` 을 거부해요(REQ-10, REQ-12). 막힌 자리를 고치는 리더는 `audit-slot done --discard` 를 써요.
- 형제 크레이트도 더러우면 Cargo 가 워크스페이스 크레이트를 한 번 넘게 컴파일할 수 있어요. AC-2 와 AC-6 이 Cargo 의 JSON 메시지로 컴파일한 단위 수를 세니, 퇴행은 시간이 아니라 숫자로 보여요.
- 무거운 자리 때문에 `audit-prep` 이 기다릴 수 있어요. 답에 기다림을 적어요(`check affected` 처럼).
- 2026-11-22 감사에서 `RUSTC_WRAPPER` 와 증분 빌드가 같이 켜져 `sccache` 가 거부했어요(exit 101). 빌드는 `RUSTC_WRAPPER` 와 `CARGO_BUILD_RUSTC_WRAPPER` 없이 돌려요(D-3).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 두 명령은 고정 자리 `audit-1`..`audit-3` 의 카드 이름 앞문이고 `audit-slot` 의 코드를 불러요. `audit-<card>` 작업트리는 만들지 않아요. | 사용자가 승인한 재설계 §3A 가 감사 자리를 고정해 다시 쓰고, SPEC-20261123-audit-slot 이 합쳐져 감사를 통과했고, 작업트리 가족이 하나 더 생기면 링크·자기 작업트리·holder 안전 코드를 복사해야 해요. 배차의 `.claude/worktrees/audit-<card>` 는 리더 결정 노트 7966 이 이 결정으로 바꿔요. | 리더가 다른 형태를 고르면 코드 전에 SPEC 을 고쳐요. |
| D-2 | 카드 이름이 holder 이고(`C-364` 는 holder 이름 규칙을 통과해요), 자리는 `auto` 로 골라요. | 카드 하나, 감사 하나, 자리 하나예요. audit-slot 의 holder 기록이 이미 두 번째 살아 있는 holder 를 거부하고 죽은 holder 를 넘겨받아요. | 나중에 `--holder` 옵션을 더해요. |
| D-3 | 빌드는 `cargo test --no-run --locked --offline --message-format json` 이고 `--test NAME` 은 최대 둘, 없으면 `--lib` 하나예요. `RUSTC_WRAPPER` 를 지우고 무거운 자리 하나 안에서 자리의 `target/` 으로 돌려요. `CARGO_TARGET_DIR` 는 비워 둬요. | 카드 C-364(뮤턴트는 대상 1~2개만), 점검 A2, 2026-11-22 의 sccache·증분 충돌이에요. Cargo 의 JSON `compiler-artifact` 가 `fresh` 를 실어 줘요. | 인자 목록을 한 곳에서 바꿔요. |
| D-4 | 이름 붙이는 테스트 대상의 상한은 둘이에요. | 점검 A2: 뮤턴트당 대상 1~2개. | 상수를 올려요. |
| D-5 | `audit-done` 에는 버리기도 강제도 없고, 남은 뮤턴트는 `audit-done` 을 거부해요. | 배차: `--force` 없음. 재설계의 정리 조건(깨끗한 트리, 프로세스·클레임 없음). | 리더의 고침 길은 `audit-slot done --discard` 로 남아요. |
| D-6 | 카드 이름이 holder 의 전부예요. 두 명령은 holder 기록의 세션을 비워 두고, 세션이 달라도 같은 카드는 같은 holder 로 봐요. 이를 위해 `audit_slot::prep` 과 `done` 을 세션을 받는 몸통(`prep_with`, `done_with`)과 지금 세션을 넘기는 기존 함수로 나눠요. 기존 함수의 동작은 그대로예요. | 리더가 준비하고 감사 에이전트가 `audit-done` 을 치는 일이 흔한데, `audit-slot` 은 둘 다 세션이 있고 다르면 다른 holder 로 보고 거부해요. | 세션을 다시 기록하게 하고 호출부가 같은 세션에서 끝내요. |
| D-7 | 명령 둘은 연산 파일 둘이에요: `src/ops/audit_prep.rs`, `src/ops/audit_done.rs`. 연산마다 문서 조각 `operations/<word>.md` 가 있어야 해요. 두 파일은 `audit_prep` 의 함수만 불러요. | `build.rs` 는 파일 이름을 명령 이름으로 쓰고 별칭이 없어요. `src/main.rs` 의 시험이 연산 파일마다 문서 조각을 요구해요. 판 1 의 범위는 연산 하나와 두 행동으로 적어 이름 `audit-done` 을 만들 수 없었어요. | 한 연산 `audit-prep prep` / `audit-prep done` 으로 줄이면 파일 둘이 줄지만 배차의 이름 `audit-done` 은 없어져요. |
| D-8 | 컴파일 때 박히는 `env!` 경로는 여기서 고치지 않아요. 다시 빌드한 실행 파일이 감사 자리를 읽는다는 시험(REQ-9)을 둬요. `audit-slot prep` 이 이미 하는 추적 소스 touch 가 자리의 워크스페이스 단위를 다시 컴파일하게 해요. | 배차 범위. 잰 파일 수는 135가 아니라 163이고 리더 노트 7966 이 기록해요. | 나중 카드가 경로를 실행 시점으로 옮기고 그다음 nextest 아카이브를 공유해요. |

## Result

- 판 1 은 리더가 승인했어요(노트 7966, 요청 노트 7965 의 선택지 (b), 커밋 `eeba19c082`).
- 판 2: 구현에 들어가기 전에 `audit-done` 이 연산 파일과 문서 조각을 따로 요구한다는 것을 읽었어요(D-7). Scope 에 `src/ops/audit_done.rs` 와 `operations/audit_done.md` 를 더했고 요구사항과 인수 기준은 판 1 과 같아요. 리더가 승인했어요(노트 7978, 선택지 (a), 커밋 `163c2571d1`).
- 승인 조건(노트 7978): `audit_slot::prep` 과 `done` 의 동작이 그대로임을, `audit_slot_cli` 전체가 READY 영수증과 같은 head 에서 통과하는 것으로 보여요. 구현 커밋 `fd40aaee40` 에서 19개 전체가 통과했어요(그 실행은 nextest 41개 = `audit_prep_cli` 9 + `audit_slot_cli` 19 + `help_layout` 11 + 운영 문서 검사 2, 모두 통과). 처음 적은 17개는 잘못 센 숫자였어요. READY head 에서 다시 도는 것은 READY 노트의 영수증이 적어요.
- RED: 구현 전 `audit_prep_cli` 9개가 모두 `unknown command` 로 실패했어요(실행 9, 통과 0, 커밋 `8ddc017965`).
- AC-1: met — `prep_takes_a_slot_for_the_card_again_the_same_one_and_refuses_when_all_are_held`, `prep_prepares_the_slot_the_card_already_holds_not_the_first_free_one`, `prep_refuses_what_audit_slot_refuses_and_runs_no_build`.
- AC-2: met on macOS — `one_build_compiles_the_workspace_units_once_and_every_dependency_fresh`.
- AC-3: met — `a_build_that_does_not_compile_keeps_the_slot_held_and_names_the_errors`.
- AC-4: met on unix (run on macOS) — `a_rebuilt_test_executable_reads_the_slot_and_holds_no_byte_of_the_lane_folder`.
- AC-5: met — `audit_done_refuses_a_left_mutant_then_releases_and_takes_no_discard_or_force`, `the_card_is_the_holder_whatever_session_finishes_it`(D-6).
- AC-6: met on macOS — `a_one_line_mutant_on_a_prepared_slot_recompiles_only_the_workspace_units`.
- 수리로 바뀐 동작 하나(감사 노트 8068 의 B4, 노트 8103 의 R1): `audit-slot prep` 은 자리 자신의 `target/` 이 링크나 파일이면 첫 쓰기 전에 거부해요(`--lane` 유무와 관계없이, 코드 `real_target`, 시험 `prep_refuses_a_slot_whose_target_is_a_link_and_writes_nothing_through_it`). 위 Scope 의 "동작 변화 없음"은 이 한 가지를 빼고 맞아요. 운영 문서 `audit_slot.md` 에 같은 문장을 적었어요(글만 고침, 요구사항과 인수 기준은 그대로).
- 가드 뮤턴트: 새 가드 테스트마다 코드 한 줄을 떼어 `exit 100` 으로 죽였어요(10개). 처음엔 셋이 살아남았어요(`exit 0`): 테스트에 단언이 없던 둘(준비한 빌드의 `compiled_dependencies`, holder 기록의 세션)과 제가 맞지 않는 테스트로 돌린 하나(`MAX_TESTS`)예요. 단언을 더하고(`8ce2d5d537`) 맞는 테스트로 다시 돌려 모두 `exit 100` 이에요.
- 모르는 것(not_run): 멈춰 버린 cargo 자식에 대한 시간 제한은 이 조각에 없어요. `audit-prep` 은 cargo 가 끝날 때까지 무거운 자리를 쥐고 기다려요. Linux·Windows 결과는 맥 픽스처로 주장하지 않아요(AC-2 와 AC-6 은 macOS 에서만 돌고, 나머지는 unix 에서 돌지만 이 조각은 macOS 에서만 돌렸어요).
