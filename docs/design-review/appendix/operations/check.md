# `check`

operation `check.affected` · 코드 `src/ops/check.rs`, `src/check_affected.rs`, `src/heavy_slot.rs` · SPEC-20261120-check-affected, SPEC-20261122-check-affected-one-slot

`check affected <base>..<head>` 는 바뀐 경로에서 그 변경이 깨뜨릴 수 있는 시험을 고르고, 검사 단계를 돌리고, 한 줄로 알려요. 레인이 READY 를 쓰기 전과 리더가 병합한 head 를 확인할 때 같은 규칙을 써요. 범위는 점 둘(`<base>..<head>`)이고, 바뀐 경로는 `<base>...<head>`(merge-base 기준)라서 옛 tip 위에서 만든 head 도 자기 변경만 골라요.

## 고르는 규칙

| 바뀐 것 | 고르는 것 |
| --- | --- |
| `src/<x>.rs` · `src/<dir>/…` · `src/ops/<y>.rs` | 영역은 `x` · `dir` · `y`. 시험 대상 이름이 영역의 **첫 낱말**(`_` 앞)과 같거나 `<첫 낱말>_` 로 시작하는 Rust 대상 전부, 같은 규칙의 파이썬 `tests/<첫 낱말>_*_test.py` |
| `tests/<이름>.rs` · `tests/<이름>_test.py` | 그 대상 자신. `tests/<폴더>/…`(공용 도우미)는 그 폴더를 `mod <폴더>;` 로 쓰는 모든 Rust 대상 |
| `tests/fixtures/<폴더>/…` | 이름이 그 폴더 이름과 같거나 `<폴더>_` 로 시작하는 Rust·파이썬 대상 |
| 무엇이 바뀌었든 | `tools/repo_scans.json` 의 `checks` 가 말하는 `--test <이름>` 과 `tests.<이름>` 전부. 파일이 없으면 `scans=none`, 읽을 수 없으면 `scans=unreadable` 이고 `tests/` 의 모든 시험을 골라요 |
| 무엇이 바뀌었든 | `cfg(target_os = "<이 호스트>")` 를 쓰는 Rust 대상, `--lib`, `--bin moai-cli` |

단계(이 차례, 해당할 때만): `diff_check`(`git diff --check <base>...<head>`), `fmt`(`cargo fmt --all -- --check`), `ops_rustfmt`(바뀐 `src/ops/*.rs` 에 `rustfmt --edition 2021 --check`), `bash_n:<경로>`(바뀐 `*.sh` 마다 `bash -n`), `clippy`(`--all-targets -- -D warnings`), `test`(라이브러리·`moai-cli` 바이너리·고른 Rust 대상 전부를 한 번에: `cargo nextest --version` 이 0 으로 끝나면 `cargo nextest run --locked --offline --no-fail-fast --lib --bin moai-cli --test <대상>…`, 아니면 같은 인자의 `cargo test` 한 번, 대상을 줄이지 않음. `--bins` 는 쓰지 않아요: `Cargo.toml` 이 `test = false` 라고 한 바이너리까지 시험 하네스를 지어서 같은 시험이 두 번 돌아요), `build_cli`(파이썬 대상만 있고 Rust 시험 대상이 없을 때 `cargo build --bin moai-cli`; Rust 대상이 있으면 `test` 가 같은 바이너리를 짓기 때문에 계획하지 않아요), `python`(`python3 -m unittest tests.<이름>…`, `CARGO_TARGET_DIR`·`MOAI_CLI_BIN` 지정).

무거운 자리(SPEC-20261122-check-affected-one-slot): 첫 cargo 단계(`clippy`) 앞에서 무거운 자리 **하나**를 잡아 마지막 cargo 단계가 끝날 때까지 쥐고, `python` 전에 놓아요. 자리 폴더는 `--slot-dir`(없으면 기계의 기본 폴더), 한도는 `heavy-slot-acquire` 와 같은 규칙(자리 옆 `limits.json`, `--max 4`·`--min-free-gb 50` 은 더 엄할 때만 이김)이에요. 줄 서는 동안의 `heavy-slot: waiting …` 줄과 자리를 잡은 줄, 거절 이유는 로그 폴더의 `heavy-slot.log` 에 남고 stdout 에는 JSON 답만 나와요. cargo 단계는 자리를 쥔 채 직접 돌고 `MOAI_HEAVY_SLOT_FD=<fd>:<자리 파일>` 을 물려받아요. 그래서 이미 자리를 쥔 `heavy-slot-acquire` 안에서 같은 자리 폴더로 돌리면 새로 줄 서지 않고 물려받은 자리를 다시 써요(열린 파일의 장치·inode 가 그 폴더의 자리 파일과 같을 때만). 자리를 얻지 못하면(125 거절, 75 기계가 바쁨, 2 자리 폴더 없음) 그 코드가 모든 cargo 단계(`test`·`build_cli` 포함)의 종료 코드가 되고 cargo 는 하나도 시작하지 않아요(nextest 확인도 자리를 잡은 뒤에만 해요). 호출한 쪽이 `CARGO_BUILD_JOBS` 를 주지 않았으면 cargo 단계에 `max(2, 논리 CPU 수 ÷ 실제 자리 수)` 를 넣고, 주었으면 그 값을 그대로 넘겨요.

`heavy-slot-acquire` 의 줄은 도착 순서예요: 누구든 먼저 `waiting/<나노초 20자리>-<pid>` 파일을 만들어 잠그고, 자기보다 먼저 온 **살아 있는**(잠긴) 기다림의 수가 빈 자리 수보다 적을 때만 자리를 잡아요. 잠기지 않은 기다림 파일(죽은 기다림)은 세지 않아요.

## 쓰는 법과 답

- `--list`: 선택과 단계만 보여 주고 아무것도 돌리지도 쓰지도 않아요(`writes: false`).
- 돌리기: 체크아웃이 범위의 head 이고 추적 파일이 고쳐지지 않았을 때만 돌아요. 가벼운 단계(`diff_check`, `fmt`, `ops_rustfmt`, `bash_n:…`: 첫 무거운 단계보다 앞의 단계)는 하나가 실패해도 서로 멈추지 않고 모두 돌지만, 그중 하나라도 실패하면 무거운 단계와 `python` 은 돌지 않고 거기서 멈춰요(SPEC-20261122-check-light-stop). 무거운 단계끼리는 하나가 실패해도 나머지를 끝까지 돌려요. 단계마다 로그(와 `heavy-slot.log`)를 빌드 폴더(호출한 쪽의 `CARGO_TARGET_DIR`, 없으면 `<저장소>/target`)의 `check-affected-<head 12자>/`(`--out` 으로 바꿈)에 남겨요. 파이썬 단계도 같은 폴더의 `debug/moai-cli` 를 가리켜서, Rust 시험 대상이 골라졌으면 `test` 가, 아니면 `build_cli` 가 만든 바이너리를 써요. `--dir` 은 저장소 안의 어느 폴더든 되고, 바뀐 경로는 `--no-renames -z` 로 읽어서 이름이 바뀐 파일은 옛 경로도 세고 한글 이름도 따옴표 없이 읽어요.
- 답의 `summary` 한 줄: `check-affected base=<12> head=<12> targets=<n> py=<n> scans=<none|listed:n|unreadable> <단계>=<종료 코드>… [stopped_at=<단계>] result=PASS|FAIL`. 가벼운 단계가 실패해서 멈췄을 때만 `stopped_at` 이 `result=` 바로 앞에 붙고(첫 실패한 가벼운 단계의 이름), 돌지 않은 단계는 줄에 낱말이 없어요. 같은 이름 단계(`bash_n:…`)는 첫 실패 코드 하나로 모아요. 시작하지 못한 프로그램은 127, 자리를 얻지 못하면 모든 cargo 단계에 자리의 코드(125 거절, 75 기계가 바쁨)가 남고 결과는 FAIL 이에요. 시험 단계는 낱말 하나 `test=<코드>` 예요. 답에는 `runner`(`nextest`·`cargo-test`, 자리를 얻지 못했거나 무거운 단계 전에 멈췄으면 null)가 있고, `steps` 의 각 기록에는 `command`(실제 실행한 프로그램과 인자)가 있어요. `result=PASS` 는 모든 종료 코드가 0 일 때만이에요.
- 거절(아무것도 돌지 않음): 범위가 `<base>..<head>` 꼴이 아님, base·head 를 읽을 수 없음, 체크아웃이 head 가 아님, 추적 파일이 고쳐짐.

하지 않는 것: PASS 를 선언(독립 감사와 CI 는 그대로), 시험이나 도구를 고치거나 설치, 한 시험 안의 케이스 고르기, 실제 HOME 읽기.
