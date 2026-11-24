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

(구현 뒤에 채워요.)
