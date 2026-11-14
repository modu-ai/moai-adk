# 새 명령 더하기 — 고칠 파일 (#2834)

병렬 PR 이 같은 줄을 고쳐 서로 충돌하지 않게, 새 명령 하나는 **자기 파일만** 더한다. `src/main.rs` 와 `tests/product_cli.rs` 는 건드리지 않는다.

## 고칠 곳

| 파일 | 무엇 | 다른 PR 과 같은 줄? |
| --- | --- | --- |
| `src/ops/<word>.rs` (새) | `pub struct Args`(clap 인자, `///` 줄이 도움말), `pub const OPERATION`(봉투의 operation 이름), `pub fn run(args: Args) -> Result<serde_json::Value, String>`. 파일 이름은 명령 이름의 `-` 를 `_` 로 | 아님(새 파일) |
| `docs/design-review/appendix/operations/<word>.md` (새) | 첫 줄 `` # `<word>` ``, 그다음 operation·코드 위치·동작 한 문단 | 아님(새 파일) |
| `src/help.rs` `GROUPS` | 자기 무리 목록에 `("<word>", false)` 한 줄 | 무리 안 한 줄 |
| `src/help.rs` `EXAMPLES` | `("<word>", &["<word> …"])` 한 줄(예시 1–3개). stdin 을 읽으면 예시 끝에 ` < request.json`. `help_rules::every_command_has_examples`·`every_example_parses`·`stdin_examples_match_the_commands_that_read_stdin` 이 확인(#2802 S1c) | 표 안 한 줄 |
| `tests/cli_<group>.rs` | 그 무리의 제품 CLI 시험. 공통 도우미는 `tests/common/cli.rs`(`cli`·`ok_result`·`profile`·`create`·`write_file`·`tree_state`)와 `tests/common/git.rs`(`git_in` 등) | 무리 파일 안 |

`build.rs` 가 `src/ops/*.rs` 를 모아 `Ops` 하위 명령으로 만들고, `src/main.rs` 는 그것을 `Command` 에 한 번 펼친다(`#[command(flatten)] Op(ops::Ops)`). 그래서 `Command` enum·`operation()` 표·dispatch `match` 에 줄을 더하지 않는다. 같은 모양의 두 가지(branch)가 각자 명령을 더해도 충돌 없이 합쳐지는 것은 `tests/hot_file_ops_merge.rs` 가 확인한다.

## stdin 을 읽는 명령

typed JSON 요청을 stdin 에서 읽으면 `src/main.rs` 의 `input()`(64 KiB 상한, 같은 오류 문구)을 `super::super::input()` 으로 부른다(`src/ops/<word>.rs` 는 `ops` 모듈 아래라 두 단계 위). 예: `src/ops/run_verify.rs`.

## 아직 이 틀로 못 오는 명령

오류가 `String` 이 아니라 타입 있는 `OpError`(`run_internal`)인 명령은 지금 `src/main.rs` 에 남는다. 옮기는 순서는 #2834 6869105247.

## 예 — R1 (#2834)

Checks 무리 7개(`review-package`·`review-check`·`audit-delta`·`audit-reduce`·`parity-check`·`plan-admit`·`run-verify`)를 이 틀로 옮겼다. 시험 쪽은 `tests/product_cli.rs` 의 identity 시험을 `tests/cli_diagnose.rs` 로 옮겼다. 도움말(`--help`·`-h`·`help --all`)·오류 봉투·종료 코드는 옮기기 전과 바이트가 같다.
