# `check`

operation `check.affected` · 코드 `src/ops/check.rs`, `src/check_affected.rs` · SPEC-20261120-check-affected

`check affected <base>..<head>` 는 바뀐 경로에서 그 변경이 깨뜨릴 수 있는 시험을 고르고, 검사 단계를 돌리고, 한 줄로 알려요. 레인이 READY 를 쓰기 전과 리더가 병합한 head 를 확인할 때 같은 규칙을 써요. 범위는 점 둘(`<base>..<head>`)이고, 바뀐 경로는 `<base>...<head>`(merge-base 기준)라서 옛 tip 위에서 만든 head 도 자기 변경만 골라요.

## 고르는 규칙

| 바뀐 것 | 고르는 것 |
| --- | --- |
| `src/<x>.rs` · `src/<dir>/…` · `src/ops/<y>.rs` | 영역은 `x` · `dir` · `y`. 시험 대상 이름이 영역의 **첫 낱말**(`_` 앞)과 같거나 `<첫 낱말>_` 로 시작하는 Rust 대상 전부, 같은 규칙의 파이썬 `tests/<첫 낱말>_*_test.py` |
| `tests/<이름>.rs` · `tests/<이름>_test.py` | 그 대상 자신. `tests/<폴더>/…`(공용 도우미)는 그 폴더를 `mod <폴더>;` 로 쓰는 모든 Rust 대상 |
| `tests/fixtures/<폴더>/…` | 이름이 그 폴더 이름과 같거나 `<폴더>_` 로 시작하는 Rust·파이썬 대상 |
| 무엇이 바뀌었든 | `tools/repo_scans.json` 의 `checks` 가 말하는 `--test <이름>` 과 `tests.<이름>` 전부. 파일이 없으면 `scans=none`, 읽을 수 없으면 `scans=unreadable` 이고 `tests/` 의 모든 시험을 골라요 |
| 무엇이 바뀌었든 | `cfg(target_os = "<이 호스트>")` 를 쓰는 Rust 대상, `--lib`, `--bins` |

단계(이 차례, 해당할 때만): `diff_check`(`git diff --check <base>...<head>`), `fmt`(`cargo fmt --all -- --check`), `ops_rustfmt`(바뀐 `src/ops/*.rs` 에 `rustfmt --edition 2021 --check`), `bash_n:<경로>`(바뀐 `*.sh` 마다 `bash -n`), `clippy`(`--all-targets -- -D warnings`), `test_lib`, `test_bins`, `test_targets`(고른 Rust 대상 전부, 줄이지 않음), `build_cli`(파이썬 대상이 있을 때 `cargo build --bin moai-cli`), `python`(`python3 -m unittest tests.<이름>…`, `CARGO_TARGET_DIR`·`MOAI_CLI_BIN` 지정). cargo 단계는 단계마다 `heavy-slot-acquire --max 4 --min-free-gb 50` 한 번을 잡아요(`--max`·`--min-free-gb`·`--slot-dir` 로 바꿈).

## 쓰는 법과 답

- `--list`: 선택과 단계만 보여 주고 아무것도 돌리지도 쓰지도 않아요(`writes: false`).
- 돌리기: 체크아웃이 범위의 head 이고 추적 파일이 고쳐지지 않았을 때만 돌아요. 어느 한 단계가 실패해도 나머지를 끝까지 돌리고, 단계마다 로그를 `<dir>/target/check-affected-<head 12자>/`(`--out` 으로 바꿈)에 남겨요.
- 답의 `summary` 한 줄: `check-affected base=<12> head=<12> targets=<n> py=<n> scans=<none|listed:n|unreadable> <단계>=<종료 코드>… result=PASS|FAIL`. 같은 이름 단계(`bash_n:…`)는 첫 실패 코드 하나로 모아요. 시작하지 못한 프로그램은 127, 슬롯이 단계를 시작하지 않으면 슬롯의 코드(125 거절, 75 기계가 바쁨)가 남고 결과는 FAIL 이에요. `result=PASS` 는 모든 종료 코드가 0 일 때만이에요.
- 거절(아무것도 돌지 않음): 범위가 `<base>..<head>` 꼴이 아님, base·head 를 읽을 수 없음, 체크아웃이 head 가 아님, 추적 파일이 고쳐짐.

하지 않는 것: PASS 를 선언(독립 감사와 CI 는 그대로), 시험이나 도구를 고치거나 설치, 한 시험 안의 케이스 고르기, 실제 HOME 읽기.
