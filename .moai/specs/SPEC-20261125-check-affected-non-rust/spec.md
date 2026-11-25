---
id: SPEC-20261125-check-affected-non-rust
title: "moai check affected runs the profile's declared checks when the repository has no Rust code, and answers not_run when it has none"
status: draft
revision: 1
created: 2026-11-25
---

# moai check affected runs the profile's declared checks when the repository has no Rust code, and answers not_run when it has none

## Intent

이 SPEC 의 카드(p1)는 다른 저장소(TypeScript 모노레포, `Cargo.toml` 없음)에서 레인이 `moai check affected` 를 돌리면 `result=FAIL stopped_at=fmt`(`cargo fmt: could not find Cargo.toml`)가 나서 READY 노트가 PASS 영수증 없음으로 거절되고 레인이 READY 를 못 낸다고 적어요. 그 저장소의 실측이라 이 저장소에서는 재현하지 않았어요(카드 본문).

읽어서 확인한 사실(코드와 명령 출력, 2026-11-24, NeXT `580d47c6ec`):

- `src/check_affected.rs` 의 `steps()` 는 저장소가 무엇이든 `diff_check`, `cargo fmt`, `clippy`, `cargo nextest run --lib --bin moai-cli` 를 고정으로 깔아요(1429-1499 줄). 임시 git 저장소(`README.md`, `package.json` 둘뿐)에서 `moai check affected HEAD~1..HEAD --list` 를 돌려 `rust_targets` 가 비었는데도 `cargo fmt`, `clippy`, `test` 단계가 나오는 것을 봤어요(재현, 이 레인).
- 프로필 `.moai/project.yaml` 에는 사람이 확인한 확인 명령 목록 `checks`(`[{id, command}]`, `src/profile.rs` 의 `checks`, 검증은 `project_ci::validate_checks`)가 이미 있어요. 생성되는 CI 워크플로가 같은 목록을 읽어요.
- READY 영수증 관문(`src/card/ready_receipt.rs` 의 `judge`)은 `result=PASS` 가 아닌 줄을 모두 "영수증 아님"으로 거절해요(`result=<other>` 라고 말해요). 그래서 이 SPEC 이 PASS 가 아닌 새 낱말을 영수증에 적어도 관문은 그것을 받지 않아요.
- 이 저장소의 `check affected` 테스트 고정물(`tests/check_affected_*.rs`)은 `Cargo.toml` 없이 `.rs` 파일만 둔 저장소를 Rust 저장소로 써요. 그래서 Rust 판정을 루트 `Cargo.toml` 하나로 하면 Rust 저장소의 선택이 바뀌어요.

읽는 사람: 레인, 이 변경의 감사자, Rust 가 아닌 저장소에서 MoAI 를 쓰는 사람.

## Non-goals

- Rust 저장소의 단계 선택·순서·요약 줄은 바뀌지 않아요(REQ-5). 이 SPEC 은 Rust 저장소에 프로필 `checks` 를 더하지 않아요.
- 프로필에 `checks` 를 쓰는 길(`/moai project` 인터뷰와 CI 계획)을 바꾸지 않아요. 이 SPEC 은 이미 있는 값을 읽기만 해요.
- 언어별 확인 명령을 MoAI 가 추측해서 정하지 않아요. 확인 명령은 프로필에 사람이 적은 것만 돌려요(하드코딩 금지).
- `--lane`·`--card` 선택, 스캔(`tools/repo_scans.json`), `card land-batch` 의 읽기를 바꾸지 않아요. 새 요약 줄 낱말 `result=not_run` 을 `land-batch` 가 받지 않는 것은 지금 코드 그대로예요(`result=PASS` 만 받아요).
- 루트가 아닌 폴더에만 `Cargo.toml` 이 있는 혼합 저장소(예: 하위 폴더의 크레이트)는 이 SPEC 이 다루지 않아요(Risks).

## Requirements

<!-- This SPEC adds no blocking check: it narrows which steps a non-Rust repository gets and adds one more result word. -->
- REQ-1: When the repository has no `Cargo.toml` at its top folder and its head tracks no `.rs` file, the `check affected` command shall plan no `fmt`, `ops_rustfmt`, `clippy` or `test` step. · failure: none — a repository without a declared check still gets the answer of REQ-3.
- REQ-2: When REQ-1 holds and the project profile declares `checks`, the command shall plan one step per declared check, in the declared order, that runs the check's command with `sh -c` in the repository's top folder as a heavy step. · failure: none — a check that fails gives the answer of REQ-4.
- REQ-3 (REQ-2): IF REQ-1 holds, no declared check can be read (no profile, a profile that cannot be read, or no `checks`) and every planned step exited 0, THEN the command shall answer `not_run` as the answer's `result` and `result=not_run` as the last word of its summary line, with `not_run.next` naming the profile key `checks` and the file `.moai/project.yaml`, never `PASS`.
- REQ-4 (REQ-2): IF a planned step exits non-zero, THEN the command shall answer `FAIL` as before, whatever REQ-3 says.
- REQ-5: When the repository has a `Cargo.toml` at its top folder or its head tracks a `.rs` file, the command shall plan and run the same steps and write the same summary line as before this SPEC. · failure: none — the steps and their failures are unchanged.
- REQ-6: The command shall write its summary line to `summary.txt` whatever the result word, so that a `not_run` run leaves a receipt that the READY note gate counts as no receipt.

## Acceptance

- AC-1 (REQ-1, REQ-2): 임시 저장소(`package.json` 과 `src/index.ts` 만, `Cargo.toml` 도 `.rs` 도 없음, 프로필 `checks` 둘)의 `--list` 는 `diff_check`, `check:lint`, `check:test` 순서로 세 단계만 내고 `fmt`·`clippy`·`test`·`ops_rustfmt` 가 없고 `rust_targets` 가 비어요; `check:` 단계는 `sh -c <명령>` 이고 heavy 예요 · verify: `cargo nextest run --locked --offline --test check_affected_non_rust -E 'test(a_repository_without_rust_plans_no_cargo_step)'`
- AC-2 (REQ-2, REQ-6): 같은 저장소에서 두 확인 명령이 0 으로 끝나면 `result` 가 `PASS` 이고 요약 줄이 `check=0 result=PASS` 로 끝나며 `summary.txt` 가 그 줄과 같고, 명령이 저장소 루트에서 돌았음을 로그가 보여요 · verify: `cargo nextest run --locked --offline --test check_affected_non_rust -E 'test(the_declared_checks_run_in_order_and_give_the_receipt)'`
- AC-3 (REQ-4): 둘째 확인 명령이 3 으로 끝나면 `result` 가 `FAIL`, 요약 줄이 `check=3` 을 말하고 `result=FAIL` 로 끝나며, 첫 명령이 실패해도 둘째 단계가 돌아요(앞 단계가 뒤를 막지 않아요, 지금 가벼운 단계 규칙 그대로) · verify: `cargo nextest run --locked --offline --test check_affected_non_rust -E 'test(a_failed_declared_check_fails_the_run)'`
- AC-4 (REQ-3, REQ-6): 프로필이 없거나, 읽을 수 없거나(깨진 파일), `checks` 가 없는 비 Rust 저장소의 실행은 `result` 가 `not_run` 이고 `PASS` 가 아니며, 요약 줄이 `result=not_run` 으로 끝나고, 답의 `not_run.next` 가 `checks` 와 `.moai/project.yaml` 을 말하고, `summary.txt` 가 그 줄이며, cargo 단계가 하나도 돌지 않아요 · verify: `cargo nextest run --locked --offline --test check_affected_non_rust -E 'test(no_declared_check_answers_not_run_with_the_next_step)'`
- AC-5 (REQ-4): `git diff --check` 가 실패하는 비 Rust 범위는 확인 명령이 없어도 `not_run` 이 아니라 `FAIL` 이에요 · verify: `cargo nextest run --locked --offline --test check_affected_non_rust -E 'test(a_failed_cheap_step_is_fail_not_not_run)'`
- AC-6 (REQ-5): `Cargo.toml` 만 있고 `.rs` 가 없는 저장소와 `.rs` 만 있고 `Cargo.toml` 이 없는 저장소는 `--list` 의 단계가 이전과 같이 `diff_check`, `fmt`, `clippy`, `test` 이고, 이 저장소의 `--list` 출력은 변경 전후가 같아요(READY 의 Result 에 두 출력의 비교를 적어요) · verify: `cargo nextest run --locked --offline --test check_affected_non_rust -E 'test(a_rust_repository_keeps_its_steps)'` and the existing `--test check_affected_cli --test check_affected_modes_cli --test check_affected_one_slot` still pass
- AC-7 (REQ-3): 요약 줄을 만드는 `summary_line` 은 `not_run` 이면 `result=not_run`, `PASS`·`FAIL` 은 이전 줄 그대로 만들어요 · verify: `cargo nextest run --locked --offline --lib -E 'test(summary_line_not_run)'`
- AC-8 (REQ-3): READY 관문 `judge` 는 `result=not_run` 줄을 영수증으로 받지 않아요 · verify: `cargo nextest run --locked --offline --lib -E 'test(a_not_run_summary_line_is_no_receipt)'`

## Scope

- write: `.moai/specs/SPEC-20261125-check-affected-non-rust/**`, `src/check_affected.rs`(Rust 저장소 판정, 비 Rust 단계, `not_run` 결과, 단위 테스트), `src/card/ready_receipt.rs`(AC-8 단위 테스트 한 개만, 코드는 바꾸지 않아요), `src/ops/check.rs`(도움말 한 문단), `tests/check_affected_non_rust.rs`
- read-only: `src/profile.rs`, `src/project_ci.rs`, `src/card/land/batch.rs`, `tests/check_affected_cli.rs`, `tests/check_affected_modes_cli.rs`, `tests/check_affected_one_slot.rs`
- forbidden: `CLAUDE.md`, `tools/repo_scans.json`, `tests/areas.toml`(새 파일의 첫 낱말 `check` 는 이미 spec 영역에 있어요), 실제 HOME, 프로필에 `checks` 를 쓰는 코드, 네트워크

## Plan

1. SPEC 초안을 커밋해요. 승인은 기록하지 않아요(리더가 노트로 해요). READY 에 digest 를 적어요.
2. RED: `tests/check_affected_non_rust.rs` 에 AC-1 ~ AC-6 의 테스트를 먼저 써서 현재 팁에서 옳은 이유(비 Rust 저장소에 `fmt` 가 나옴, 요약이 FAIL)로 실패하는 것을 봐요. AC-7·AC-8 의 단위 테스트도 먼저 써요.
3. GREEN: `Facts` 에 `stack`(Rust · 선언된 확인 · 선언 없음과 이유)을 더하고 `read_facts` 가 읽어요. `list`·`run` 의 세 호출 자리가 `steps()` 대신 `stack` 으로 가르는 한 함수를 불러요(`steps()` 는 그대로예요). `run` 은 비 Rust 에서 runner 를 묻지 않고, 모든 단계가 0 이고 확인이 선언되지 않았으면 `not_run` 으로 답해요. `summary_line` 의 결과 낱말이 세 가지가 돼요.
4. 확인: 새 테스트, 이웃 테스트(`check_affected_cli`, `check_affected_modes_cli`, `check_affected_one_slot`, `check_affected_card_select`, `check_affected_scans_cli`, `check_affected_tables_cli`, `check_affected_touch`), `cargo fmt`, 대상 한정 clippy, 가드 뮤턴트 한 줄씩, 위험 경로(`sh -c` 로 프로필 명령을 시작하는 자리)의 실제 실행(임시 프로젝트에서 멈추는 자식·복합 명령), `moai check affected` 를 READY 에서 한 번.
5. SYNC: Result 에 AC 별 met·not_run 을 적고 카드 기준 1~4 를 `moai card criterion` 으로 표시해요. 감사는 이 변경을 쓰지 않은 세션이 맡아요.

## Risks

- 프로필 `checks` 의 명령은 사람이 확인한 한 줄이고 `project_ci::command_ok`(제어 문자·비밀 모양 거절)를 이미 통과해요. 이 SPEC 은 그 명령을 `sh -c` 로 시작하는 새 자리를 만들어요: 프로필이 곧 실행 권한이라는 점은 CI 워크플로가 같은 목록을 `run:` 으로 돌리는 것과 같지만, 레인이 PR 의 프로필 변경을 돌릴 수 있다는 뜻이에요. 감사는 프로필이 바뀐 범위에서 이 점을 봐요.
- 혼합 저장소(루트 `Cargo.toml` 은 없고 하위 폴더에 크레이트나 `.rs` 가 있음)는 `.rs` 판정 때문에 Rust 저장소로 읽히고, 이전처럼 `cargo fmt` 가 실패해요. 이전과 같은 동작이라 나빠지지 않지만 풀리지도 않아요. 후속으로 둬요.
- `.rs` 파일이 없는 Rust 저장소(새 크레이트인데 `Cargo.toml` 만 커밋)는 루트 `Cargo.toml` 으로 Rust 로 읽혀요(AC-6).
- 새 결과 낱말 `not_run` 을 아는 곳은 READY 관문(`judge`, 모르는 낱말은 거절)과 `card land-batch`(`result=PASS` 만 받음)예요. `git grep` 으로 `result=` 를 읽는 곳을 찾아 이 둘과 `src/ops/card.rs` 의 도움말 한 줄뿐인 것을 확인해요(구현 중 한 번 더).
- 확인 명령이 오래 걸리면 heavy 슬롯을 오래 잡아요. 비 Rust 저장소의 확인은 지금 아무 슬롯도 쓰지 않았으니 같은 규칙(한 번에 한 레인)을 따르게 하는 쪽을 골랐어요(D-2).

## Decisions

승인 전 제안이에요. 사용자나 위임받은 리더가 고르면 revision 을 올려 `approval` 을 기록해요(레인은 승인을 기록하지 않아요).

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | "Rust 저장소"는 루트 `Cargo.toml` 이 있거나 head 가 `.rs` 파일을 하나라도 추적하는 저장소다. 대안: 루트 `Cargo.toml` 만 | 이 저장소의 고정물이 `Cargo.toml` 없이 `.rs` 만 두고 Rust 저장소로 쓰여서, `Cargo.toml` 하나로 판정하면 Rust 저장소의 선택이 바뀌어요(카드 기준 3). 카드 본문의 "Cargo.toml 과 Rust 타깃이 없으면"을 `.rs` 추적으로 읽었어요. 기본: 이대로 | 판정 함수 한 줄을 바꿔요 |
| D-2 | 선언된 확인은 heavy 단계로 돌려 기존 heavy 슬롯 아래서 차례로 돈다. 대안: 가벼운 단계 | 확인 명령은 빌드·테스트라 무거울 수 있어요; 한 번에 한 무거운 실행이라는 이 저장소의 레인 규칙을 따라요. 슬롯을 못 얻으면 기존대로 모든 heavy 단계가 그 종료 코드를 받고 아무것도 시작하지 않아요. 기본: heavy | `heavy: false` 로 바꿔요 |
| D-3 | 확인이 하나도 없으면 `PASS` 가 아니라 `not_run` 이고, 다른 단계가 실패하면 `FAIL` 이다. 영수증 파일은 쓰되 관문이 받지 않는다. 대안: 오류로 끝냄(영수증 없음) | 카드 본문: PASS 가 아니라 not_run 으로 답하고 프로필에 무엇을 적으면 되는지 말해요. 영수증을 쓰면 레인이 `result=not_run` 이라는 이유를 관문의 거절문에서 읽어요. 게이트는 느슨해지지 않아요(관문이 `result=PASS` 만 받는 것은 그대로). 기본: not_run | 오류로 바꿔요 |
| D-4 | 프로필이 읽히지 않으면(깨짐, 검증 실패) 확인이 없는 것으로 보고 `not_run` 의 이유에 그 오류를 적는다. 대안: 오류로 끝냄 | 읽을 수 없는 프로필을 추측으로 메우지 않고 PASS 로도 가지 않아요. 기본: not_run 과 이유 | 오류로 바꿔요 |

## Result

구현 에이전트가 채워요(이어받은 레인 agent-7, 2026-11-25T03:28Z 시계 기준). 감사가 다시 확인해요. SPEC revision 1, 승인 digest `sha256:8b859b883a82a780c2aa1e08f9ca404c1feeef73b11f3d22585cca2f068abb28`(리더 결정 노트 8890 의 인용, `moai spec show` 로 같은 값을 다시 읽었어요). macOS arm64, 이 워크트리, 기준은 NeXT `0befa9cba6` 를 한 번 머지한 `868df65825` 위예요.

- AC-1…AC-6 met: `tests/check_affected_non_rust.rs` 6개 통과(`cargo nextest run`, 로그 `target/c487/baseline-merged.log`: 합친 tip 에서 9개 통과 = 통합 6 + 단위 3; 이웃 8개 대상 `target/c487/neighbours.log` 는 67개 중 66개 통과, 1개 실패를 아래 회귀로 고쳤어요). 고치기 전 RED 는 이전 레인이 기록했어요(`target/c487/red.log`: 새 7개 중 4개가 맞는 이유로 실패).
- AC-7 met: `summary_line_not_run_pass_and_fail` 통과. AC-8 met: `a_not_run_summary_line_is_no_receipt` 통과(`src/card/ready_receipt.rs`, 코드는 안 바꿨어요).
- AC-6 의 `--list` 비교: 이 저장소의 `check affected 0befa9cba6..HEAD --list` 를 변경 전 빌드(`moai` 빌드 26)와 이 빌드로 돌려 `jq -S .result` 를 `diff` 하면 같아요(`runner` nextest, `rust_targets` 127, 단계 `diff_check fmt ops_rustfmt test_tables clippy test`).
- 이웃 테스트의 회귀 둘을 찾아 고쳤어요(REQ-5 의 가정이 틀린 두 고정물): `tests/check_affected_touch.rs` 의 `a_changed_link_is_skipped_and_named` 와 `tests/card/card_land_batch_cli.rs` 의 `scans_the_wide_run_of_this_program_covers_the_scans_and_lands_every_head` 는 `Cargo.toml` 도 `.rs` 도 없는 저장소를 Rust 로 썼어요. 비 Rust 로 읽혀 `not_run`(PASS 아님)이 나와 실패했고, 두 고정물에 루트 `Cargo.toml` 한 줄씩을 넣어 둘 다 통과해요(37개 통과, `target/c487/neighbours3.log`). Rust 판정(D-1)은 바꾸지 않았어요.
- 실제 Bun 실행(카드 기준 5): bun 1.4.4 가 있어요. 임시 프로젝트(`package.json`, `bun.lock`, `src/index.ts`, `src/index.test.ts`, 프로필 `checks` 둘 `bun test`·`bun run lint`)에서 `--list` 는 `diff_check`, `check:test`, `check:lint` 만 내고(`sh -c`, heavy, cargo 단계 없음), 실제 실행은 `bun test` 1 pass, 린트 번들 성공으로 `check-affected base=41ebdb18a398 head=ed94ca996af9 targets=0 scans=none test-threads=5 diff_check=0 check=0 result=PASS`, `summary.txt` 와 같아요. 임시 HOME 의 임시 장부(`ledger-register-apply`, 진짜 HOME 아님)에서 같은 헤드의 READY 노트가 수락됐어요(`card note` 응답 `ok: true`). 프로필에서 `checks` 를 뺀 다음 커밋은 `result=not_run`(`not_run.reason` = the profile declares no `checks`)이고 READY 노트는 `has no PASS check-affected receipt … result=not_run` 으로 거절되고 장부에는 수락된 노트 하나뿐이에요.
- 위험 경로 실제 실행(`sh -c` 로 프로필 명령을 시작하는 자리, 같은 임시 프로젝트): 복합 명령(`echo one && false || echo two`)은 `one`, `two` 를 쓰고, 따옴표와 파이프(`printf '%s\n' "a b" | wc -c`)는 그대로 돌고, 신호로 끝나는 명령(`kill -TERM $$`)은 `check=143`, 8초 걸리는 자식(`sleep 8`)은 12:10:50→12:10:58 동안 슬롯 하나를 잡았다가 풀었고(`heavy-slot.log`; 상태에 남은 항목은 다른 프로세스의 것), 앞 단계가 실패해도 뒤 단계(`exit 3`)가 이어서 돌아 `result=FAIL` 이에요. 시간 제한은 없어요: 멈추는 확인은 자식이 끝날 때까지 슬롯을 잡아요(cargo 단계와 같은 동작이라 이 SPEC 이 바꾸지 않았어요; 후속 후보).
- guard-mutant(셸과 cargo 로 한 슬롯에서 12개, 스크립트 `target/c487/mutants.sh`, 결과 `target/c487/mutants2.out`, 마지막 줄 `summary: 12 of 12 killed`, 이어서 `git status --short` 가 비었어요; 이전 레인의 Python 스크립트는 돌리지 않았어요):
  - guard-mutant: a_rust_repository_keeps_its_steps | `Cargo.toml` 판정을 지움 | exit 100
  - guard-mutant: a_rust_repository_keeps_its_steps | `.rs` 추적 판정을 지움 | exit 100
  - guard-mutant: no_declared_check_answers_not_run_with_the_next_step | 선언 없음이 PASS 로 답함 | exit 100
  - guard-mutant: a_failed_cheap_step_is_fail_not_not_run | 실패한 단계를 not_run 이 가림 | exit 100
  - guard-mutant: a_repository_without_rust_plans_no_cargo_step | 선언된 확인을 가벼운 단계로 | exit 100
  - guard-mutant: a_repository_without_rust_plans_no_cargo_step | 선언된 확인을 모두 건너뜀 | exit 100
  - guard-mutant: a_repository_without_rust_plans_no_cargo_step | 선언된 확인 순서를 뒤집음 | exit 100
  - guard-mutant: no_declared_check_answers_not_run_with_the_next_step | 읽을 수 없는 프로필을 빈 선언으로 읽음 | exit 100
  - guard-mutant: no_declared_check_answers_not_run_with_the_next_step | 다음 걸음 문구에서 프로필 파일 이름을 뺌 | exit 100
  - guard-mutant: summary_line_not_run_pass_and_fail | not_run 낱말을 PASS 로 | exit 100
  - guard-mutant: a_not_run_summary_line_is_no_receipt | 관문이 not_run 을 PASS 로 받음 | exit 100
  - guard-mutant: no_declared_check_answers_not_run_with_the_next_step | PASS 가 아니면 영수증 파일을 안 씀 | exit 100
- 알려진 한계: `moai check affected` 는 `not_run` 일 때도 종료 코드 0 이에요(READY 관문은 `result=not_run` 줄로 거절하고, 이 SPEC 이 종료 코드를 정하지 않았어요). 종료 코드로만 성공을 읽는 스크립트는 `not_run` 을 성공으로 읽을 수 있으니 후속에서 정해요.
- 첫 READY 영수증(`check affected --card`)은 `test_tables=1 stopped_at=test_tables` 로 FAIL 이었어요: 새 테스트 파일 `tests/check_affected_non_rust.rs` 가 상태 폴더 이름을 적는데 `tests/fixtures/state_folder_literals/allow.tsv` 에 행이 없었어요. 표가 알려 준 행(`0`, `3`)을 더했어요.
- outside-scope: tests/check_affected_touch.rs 1줄 추가, tests/card/card_land_batch_cli.rs 3줄 추가, tests/fixtures/state_folder_literals/allow.tsv 1줄 추가(카드 범위를 `card scope` 로 넓혔고 이유는 카드에 있어요).
- 돌리지 않은 것: 전체 `cargo test`(리더 몫), 다른 호스트의 실행, Windows 와 Linux. READY 영수증(`check affected --card`)은 이 Result 를 커밋한 뒤 한 번 내요.
