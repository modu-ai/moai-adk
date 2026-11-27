# `check`

operation `check.affected` · 코드 `src/ops/check.rs`, `src/check_affected.rs`, `src/check_fast.rs`, `src/check_fix.rs`, `src/heavy_slot.rs` · SPEC-20261120-check-affected, SPEC-20261122-check-affected-one-slot, SPEC-20261123-check-precise-selection, SPEC-20261125-check-fast-and-fix

`check affected <base>..<head>` 는 바뀐 경로에서 그 변경이 깨뜨릴 수 있는 시험을 고르고, 검사 단계를 돌리고, 한 줄로 알려요. 레인이 READY 를 쓰기 전과 리더가 병합한 head 를 확인할 때 같은 규칙을 써요. 범위는 점 둘(`<base>..<head>`)이고, 바뀐 경로는 `<base>...<head>`(merge-base 기준)라서 옛 tip 위에서 만든 head 도 자기 변경만 골라요.

## 고르는 규칙

| 바뀐 것 | 고르는 것 |
| --- | --- |
| `src/<x>.rs` · `src/<dir>/…` · `src/ops/<y>.rs` | 영역은 `x` · `dir` · `y`. 시험 대상 이름이 영역의 **첫 낱말**(`_` 앞)과 같거나 `<첫 낱말>_` 로 시작하는 Rust 대상 전부 |
| `tests/<이름>.rs` | 그 대상 자신. `tests/<폴더>/…`(공용 도우미)는 그 폴더를 `mod <폴더>;` 로 쓰는 모든 Rust 대상 |
| `tests/fixtures/<폴더>/…` | 이름이 그 폴더 이름과 같거나 `<폴더>_` 로 시작하는 Rust 대상 |
| 무엇이 바뀌었든 | `tools/repo_scans.json` 의 `checks` 가 말하는 `--test <이름>` 전부(`unittest` 검사의 `tests.<이름>` 은 읽지 않아요). 파일이 없으면 `scans=none`, 읽을 수 없으면 `scans=unreadable` 이고 `tests/` 의 모든 시험을 골라요 |
| 무엇이 바뀌었든 | `cfg(target_os = "<이 호스트>")` 를 쓰는 Rust 대상, `--lib`, `--bin moai-cli` |

단계(이 차례, 해당할 때만): `diff_check`(`git diff --check <base>...<head>`), `fmt`(`cargo fmt --all -- --check`), `ops_rustfmt`(바뀐 `src/ops/*.rs` 에 `rustfmt --edition 2021 --check`), `bash_n:<경로>`(바뀐 `*.sh` 마다 `bash -n`), `clippy`(`--all-targets -- -D warnings`), `test`(라이브러리·`moai-cli` 바이너리·고른 Rust 대상 전부를 한 번에: `cargo nextest --version` 이 0 으로 끝나면 `cargo nextest run --locked --offline --no-fail-fast --lib --bin moai-cli --test <대상>…`, 아니면 같은 인자의 `cargo test` 한 번, 대상을 줄이지 않음. `--bins` 는 쓰지 않아요: `Cargo.toml` 이 `test = false` 라고 한 바이너리까지 시험 하네스를 지어서 같은 시험이 두 번 돌아요). `test` 단계에만 시험 스레드 상한 4 가 걸려요(SPEC-20261120-test-parallelism-cap): nextest 면 `run` 바로 뒤에 `-j 4`, `cargo test` 면 그 단계 환경에 `RUST_TEST_THREADS=4`이고, 다른 단계에는 아무것도 붙지 않아요. 호출한 쪽이 `NEXTEST_TEST_THREADS`(nextest) 나 `RUST_TEST_THREADS`(`cargo test`) 를 주었으면 그 값을 그대로 두고 줄에도 그 값을 보여 줘요. 자리 폴더의 `limits.json` 에 `test_threads` 는 아직 쓰지 않아요: `heavy-slot-acquire` 가 모르는 키를 거절하는 한 그 키는 모든 무거운 실행을 막아요. 파이썬을 돌리는 단계(`python`)와 바이너리만 따로 짓는 단계(`build_cli`)는 없어요(SPEC-20261123-rust-scans-before-ready): 파이썬 시험 모듈은 고르지도 읽지도 않고, 남은 파이썬 시험은 CI 의 `unittest` 단계에서만 돌아요.

무거운 자리(SPEC-20261122-check-affected-one-slot): 첫 cargo 단계(`clippy`) 앞에서 무거운 자리 **하나**를 잡아 마지막 cargo 단계가 끝날 때까지 쥐고, 마지막 단계 뒤에 놓아요. 자리 폴더는 `--slot-dir`(없으면 기계의 기본 폴더), 한도는 `heavy-slot-acquire` 와 같은 규칙(자리 옆 `limits.json`, `--max 4`·`--min-free-gb 50` 은 더 엄할 때만 이김)이에요. 줄 서는 동안의 `heavy-slot: waiting …` 줄과 자리를 잡은 줄, 거절 이유는 로그 폴더의 `heavy-slot.log` 에 남고 stdout 에는 JSON 답만 나와요. cargo 단계는 자리를 쥔 채 직접 돌고 `MOAI_HEAVY_SLOT_FD=<fd>:<자리 파일>` 을 물려받아요. 그래서 이미 자리를 쥔 `heavy-slot-acquire` 안에서 같은 자리 폴더로 돌리면 새로 줄 서지 않고 물려받은 자리를 다시 써요(그 폴더의 `slot-<숫자>` 이고, 열린 파일의 장치·inode 가 그 자리 파일과 같고, 누군가 그 자리를 쥐고 있고, 그 설명자에 대한 비차단 배타 `flock` 이 성공할 때만 — `flock` 은 열린 파일 설명에 붙으니 쥔 쪽이 물려준 설명자만 성공하고, 스스로 새로 연 자리 파일은 줄을 서요). 자리 기록(`heavy-slot-status`)에는 이 실행의 명령 줄이 남아요. 자리를 얻지 못하면(125 거절, 75 기계가 바쁨, 2 자리 폴더 없음 또는 `--max 0`) 그 코드가 모든 cargo 단계(`clippy`·`test`)의 종료 코드가 되고 cargo 는 하나도 시작하지 않아요(nextest 확인도 자리를 잡은 뒤에만 해요). 호출한 쪽이 `CARGO_BUILD_JOBS` 를 주지 않았으면 cargo 단계에 `max(2, 논리 CPU 수 ÷ 실제 자리 수)` 를 넣고, 주었으면 그 값을 그대로 넘겨요.

`heavy-slot-acquire` 의 줄은 도착 순서예요: 누구든 먼저 `waiting/<나노초 20자리>-<pid>-<max>` 파일을 만들어 잠가요(끝이 자기 `--max`). 볼 때마다 자기 `0..max` 의 빈 자리를 낮은 것부터 늘어놓고, 자기보다 먼저 온 **살아 있는**(잠긴) 기다림마다 도착 순서로 그 기다림이 쓸 수 있는 가장 낮은 빈 자리를 하나씩 내준 뒤, 남은 자리가 있을 때만 그것을 잡아요. 그래서 먼저 온 좁은(`--max 1`) 기다림은 자기가 못 쓰는 slot 1 을 넓은 호출자에게서 막지 않아요. `<max>` 가 없는 이름(Python 쌍둥이, 옛 빌드)은 어느 자리든 쓸 수 있다고 봐요. 잠기지 않은 기다림 파일(죽은 기다림)은 세지 않고 그대로 둬요. 기다림 목록은 한 번 볼 때 한 번만 읽어요.

리더의 넓은 검사는 줄 맨 앞에 서요(SPEC-20261126-leader-slot-front): `--full` 로 돌리는 `check affected`(`card land-batch`·`card land-sync` 안의 호출 포함)가 `card land-batch` 와 같은 확인으로 등록된 리더 세션의 호출일 때만, 기다림 파일 이름의 시각을 나노초 대신 마이크로초로 써요(이름 모양 `<20자리>-<pid>-<max>` 는 그대로). 마이크로초 값은 문자열로 정렬하면 어떤 나노초 이름보다 앞이라 옛 빌드의 레인도 그 이름을 앞으로 세요. 앞자리끼리는 도착 순서예요. 확인은 호출의 세션 id(`CLAUDE_CODE_SESSION_ID`, 없으면 `CODEX_THREAD_ID`)가 살아 있는 `leader` 항목(모드 leader, lease 안)의 `session` 과 같은지 보는 것이에요. 세션 id 가 없거나, 항목이 없거나, 다른 세션이거나, 레지스트리를 못 읽으면 거절하지 않고 도착 순서로 서요. `--name leader`, 다른 플래그, `--full` 없는 실행, `heavy-slot-acquire` 는 앞에 서지 못해요. 슬롯 수, `limits.json` 한도, 거절 코드, CPU 규칙, 실행 중인 작업은 그대로이고 순서만 바뀌어요. 로그 폴더의 `heavy-slot.log` 에 `heavy-slot: queue front place (…)` 또는 `heavy-slot: queue arrival order (<이유>)` 한 줄이 남아요. 한계: 세션 id 는 `moai session-register --show` 로 읽혀서 값을 베껴 넣은 세션은 막지 못해요(이름·역할 말만 흉내 내는 것을 막아요).

## 두 가지 선택 (SPEC-20261123-check-precise-selection)

위의 "고르는 규칙" 표는 **전체 선택**(`selection: full`)이고, 맨 명령(`--lane` 없이)이 이것을 골라요. `--lane` 은 같은 범위에서 그 부분집합인 **레인 선택**(`selection: lane`)을 골라요: `--lane` 은 개발 중에 쓰는 부분 실행이에요. 그 요약 줄에는 `scans=` 바로 뒤에 `selection=lane` 낱말이 붙고(맨 명령 줄에는 없어요), `card note` 는 그 줄을 `LANE-READY-INTEGRATION`·`REPAIRED`·`READY ` 노트의 영수증(`summary.txt`)으로 받지 않아요(SPEC-20261124-ready-receipt-followup). READY head 에서 맨 명령을 한 번 돌린 `result=PASS` 가 영수증이고, 병합 후보 한 개마다 한 번 도는 열차 실행과 MERGE-READY 의 근거도 같은 맨 명령이에요. 레인 선택은 위 표의 세 규칙만 좁혀요.

`--card` 는 **카드 선택**(`selection: card`, SPEC-20261124-card-select-batch-land)을 골라요: 범위가 바꾸거나 더한 테스트(`tests/` 의 파일·고정 자료), 바뀐 소스 모듈의 이름을 부르는 테스트(테스트 본문이 `moai_cli::<모듈>` 이나 `<폴더>::<모듈>` 을 쓰거나 테스트 이름에 모듈 이름 낱말이 있는 것; 이름 낱말은 폴더 아래 모듈이면 `<폴더>_<모듈>`, 최상위 파일이면 모듈 이름이에요; 이름 낱말이나 본문 경로로 닿는 테스트 대상이 11개 이상인 모듈은 아무것도 말하지 않으므로 빼고 답의 `card_ignored_words` 에 낱말을 적어요), 규칙 파일이 바뀐 스캔의 테스트, 바뀐 `skills/`·`docs/` 경로를 부르는 테스트만 골라요. 영역 규칙(첫 낱말이 같은 테스트 전부)은 쓰지 않아요. 의존성(`Cargo.toml`·`Cargo.lock`), 빌드 설정(`build.rs`·툴체인·`.cargo/`), 공통 설정(`.moai/project.yaml`), 권한·훅 가드(`src/hook_*.rs` 등), 원장 저장층(`src/ledger/schema.rs`·`migrate.rs`·`open.rs`·`recover.rs`)이나 읽을 수 없는 스캔 표가 바뀌면 영향 범위가 분명하지 않으므로 전체 선택으로 돌아가고 답의 `card_wide` 가 이유와 경로를 말해요(그 줄에는 `selection=` 낱말이 없어요). 카드 선택으로 좁힌 줄에는 `scans=` 바로 뒤에 `selection=card` 가 붙고, `card note` 는 이 줄을 READY 영수증으로 받아요. `--lane` 과 함께 쓸 수 없어요. 넓은 확인은 `card land-batch` 가 묶음마다 한 번 돌려요.

| 규칙 | 전체 선택 | 레인 선택 (`--lane`) |
| --- | --- | --- |
| 스캔 (`tools/repo_scans.json`) | 모든 스캔의 `--test` 대상 | 바뀐 경로가 그 스캔의 `scope`(`prefix` 아래이고, `suffixes` 가 있으면 그 중 하나로 끝남)에 들거나 `rules` 파일 중 하나일 때만. `scope` 가 없거나 읽을 수 없는 스캔은 `rules` 파일이 바뀐 때 말고는 빠져요 |
| `tests/common/<이름>.rs` 가 바뀜 | `mod common;` 을 쓰는 모든 Rust 대상 | 소스에 `common::<이름>`(또는 `common::{<이름>, …}` 묶음의 항목)이 있는 대상만. `tests/common/mod.rs` 나 더 깊은 경로가 바뀌면 전체와 같아요 |
| `cfg(target_os = "<이 호스트>")` 시험 | 모두 | 영역 규칙(이름의 첫 낱말이 바뀐 경로의 영역과 같음)으로 고른 것만 |

도우미 경로 규칙(path-import rule)도 두 선택 모두에 같아요: 바뀐 경로가 `tests/common/` 아래이면 소스에 `common/<나머지 경로>` 가 글자 그대로 있는 대상(`#[path = "common/<이름>.rs"]` 로 불러 쓰는 시험, 다른 모듈 이름으로 불러도 같아요)과, `tests/common/<이름>.rs` 바로 아래 파일이면 `mod <이름>;` 을 선언한 대상을 골라요. `mod common;` 없이 도우미를 불러 쓰는 시험을 놓치지 않으려는 규칙이에요. 레인에서는 이 규칙이 고른 것에 위 표의 `common` 갈래가 더해져요.

레인 선택은 스캔 `scope` 와 `rules` 밖의 변경(`README.md`, `Cargo.toml`, `build.rs`, `src/main.rs`, `REVIEW.md`, `CLAUDE.md`, `tools/repo_scans.json` 자체)을 스캔 대상으로 이어 주지 않아요. 그런 변경은 전체 선택(열차 실행)이 잡아요.

글 고정 규칙은 두 선택 모두에 같아요: 바뀐 경로가 `skills/` 나 `docs/` 아래이고 그 경로가 어떤 `tests/*.rs` 소스에 글자 그대로 적혀 있으면 그 대상을 골라요(소스가 그 경로를 이름 그대로 부르는 시험, names that path). 경로를 실행 중에 조립하는 시험은 보이지 않아요.

`--list` 와 돌리기의 답에는 `selection`(`lane` 또는 `full`)과 `range_note` 가 있어요. `--integration-ref <ref>` 를 주었고 범위의 base 가 head 와 그 ref 의 merge-base 보다 엄격히 앞선 조상이면 `range_note` 는 `{merge_base, integration_ref, extra_changed}` 객체예요: 그 merge-base 와, 범위가 merge-base 에서 시작한 범위보다 더 가진 바뀐 경로 수예요(옛 분기점에서 tip 을 병합한 뒤 범위를 잡으면 tip 이 이미 가진 변경을 다시 시험해요). 알리기만 하고 막지 않아요. 플래그가 없거나 ref 를 읽을 수 없거나 범위가 merge-base 에서 시작하면 `range_note` 는 `null` 이고 선택은 그대로예요. `summary` 줄의 낱말과 순서는 바뀌지 않아요.

## 시험 영역 표 (SPEC-20261123-test-area-crates)

통합 시험(`tests/*.rs`)은 영역 크레이트 8개(`card` `update` `home` `project` `lane` `screen` `spec` `repo`)로 한 영역씩 옮겨져요. 어느 파일이 어느 영역인지는 `tests/areas.toml` 한 파일이 정해요: 파일 이름의 **첫 낱말**(첫 `_` 앞)을 영역에 적어 둔 낱말 목록에서 찾아요. 한 줄에 배열 하나이고(줄을 나누지 않아요: 파이썬 도구가 `tomllib` 없이 줄 단위로 읽어요), `[solo]` 의 파일(`set_var`·`remove_var`·`set_current_dir` 을 부르는 3개)은 영역에 들어가지 않고 따로 시험 대상으로 남아요.

- `tests/test_area_layout.rs` 가 지켜요: 모든 시험 파일의 첫 낱말이 표에 있어야 하고(새 첫 낱말은 같은 변경에서 표에 한 줄 더하기), 옮겨진 영역(`tests/<영역>/main.rs` 가 있는 영역)에는 그 영역의 낱말로 시작하는 최상위 `tests/*.rs` 가 남으면 안 돼요. `main.rs` 는 모듈 파일마다 `mod <이름>;` 을 정확히 한 번 적고, 모듈 안에서는 전역 상태를 바꾸는 호출을 하지 않아요.
- 새 최상위 `tests/*.rs` 파일은 같은 변경에서 두 곳에 맞춰요: 첫 낱말을 `tests/areas.toml` 의 `[areas]` 배열 한 줄에 적고, 상태 폴더 이름을 일부러 적는 파일(옛 배치 픽스처, 이전 세대 폴더, 마이그레이션)이면 `tests/fixtures/state_folder_literals/allow.tsv` 에 `경로<TAB>옛 이름 수<TAB>새 이름 수` 행을 더해요. `check affected` 는 범위가 새로 더한 최상위 `tests/*.rs` 마다 이 두 표를 무거운 단계보다 앞에서 읽는 가벼운 단계 `test_tables` 로 맞춰 봐요(SPEC-20261124-new-test-file-precheck): 어긋나면 그 단계가 1 로 끝나고, 로그의 줄마다 파일, 표, 더할 낱말이나 행을 말하고, 요약 줄은 `stopped_at=test_tables` 로 무거운 단계 없이 멈춰요. 두 표의 규칙은 풀지 않아요: 단계는 읽기만 하고, 더한 파일이 없거나 표 파일이 없는 저장소에서는 계획되지 않아요(`--list` 에도 없어요). 표 파일이 있는데 읽을 수 없으면(잘못된 UTF-8 등) 없는 표로 보지 않아요: 더한 파일이 있는 범위에서 `test_tables` 가 그 표의 경로와 읽기 오류를 한 줄로 말하고 멈춰요. 옮겨진 영역의 `tests/<영역>/*.rs` 모듈 파일과 고쳐서 문자열 수가 늘어난 기존 파일은 이 단계가 보지 않고 두 테스트가 그대로 잡아요. 같은 단계가 세 번째 표 `tools/repo_scans.json` 도 읽어요(SPEC-20261124-repo-scans-precheck): 새 파일이 저장소 트리를 걷거나(`CARGO_MANIFEST_DIR` 와 `read_dir(` 같은 호출), `include_str!`·`include_bytes!` 로 `tests/` 밖의 파일을 읽거나, 체크인된 파일의 경로를 문자열로 적고 열면, 그 파일을 스캔 항목의 규칙 파일로 등록해요(읽는 파일은 그 경로를 범위에 둔 `always` 항목과 `--test <이름>` 검사로). 임시 폴더나 이름 붙은 파일만 만지는 파일은 `tests/repo_scans_table.rs` 의 `EXEMPT`, `INCLUDE_STR_EXEMPT`, `FILE_READ_EXEMPT` 에 이유와 함께 적어요. 어느 쪽에도 없으면 `test_tables` 가 파일과 경로와 더할 곳을 한 줄로 말하고 멈춰요. `tools/repo_scans.json` 이 읽을 수 없거나 `scans` 목록이 아니면 같은 단계가 그 표를 없는 표로 보지 않고 한 줄로 말해요. 예외 표 셋은 `tests/repo_scans_table.rs` 글에서 읽기만 하고 풀지 않아요.
- 목록 파일(`tools/repo_scans.json`, `tools/self_check.sh`, `tools/ci_scope.py`)의 `--test <이름>` 과 `tests/….rs` 경로는 실제로 있는 대상이어야 해요. 영역 크레이트는 `--test <영역> <모듈>::<필터>` 꼴로 적어요.
- 시험 함수 이름은 그대로이고 옛 대상 이름이 모듈 경로가 돼요: `cargo test --test <옛 이름> <f>` → `cargo test --test <영역> <옛 이름>::<f>`.
- 지금 `check affected` 는 아직 영역을 몰라요(영역 크레이트 한 번 빌드와 모듈 필터는 테스트 스레드 상한 작업의 둘째 조각이 착지한 뒤의 조각이에요). 그래서 한 영역도 옮겨지지 않은 동안은 위의 선택 규칙 그대로예요.

## 쓰는 법과 답

- `--list`: 선택과 단계만 보여 주고 아무것도 돌리지도 쓰지도 않아요(`writes: false`).
- 돌리기: 체크아웃이 범위의 head 이고 추적 파일이 고쳐지지 않았을 때만 돌아요. 가벼운 단계(`diff_check`, `fmt`, `ops_rustfmt`, `bash_n:…`, 새 테스트 파일이 있을 때 `test_tables`: 첫 무거운 단계보다 앞의 단계)는 하나가 실패해도 서로 멈추지 않고 모두 돌지만, 그중 하나라도 실패하면 무거운 단계는 돌지 않고 거기서 멈춰요(SPEC-20261122-check-light-stop). 무거운 단계끼리는 하나가 실패해도 나머지를 끝까지 돌려요. 단계마다 로그(와 `heavy-slot.log`)를 빌드 폴더(호출한 쪽의 `CARGO_TARGET_DIR`, 없으면 `<저장소>/target`)의 `check-affected-<head 12자>/`(`--out` 으로 바꿈)에 남겨요. `--dir` 은 저장소 안의 어느 폴더든 되고, 바뀐 경로는 `--no-renames -z` 로 읽어서 이름이 바뀐 파일은 옛 경로도 세고 한글 이름도 따옴표 없이 읽어요.
- 답의 `summary` 한 줄: `check-affected base=<12> head=<12> targets=<n> scans=<none|listed:n|unreadable> [selection=lane|card] test-threads=<n> <단계>=<종료 코드>… [stopped_at=<단계>] result=PASS|FAIL`. 가벼운 단계가 실패해서 멈췄을 때만 `stopped_at` 이 `result=` 바로 앞에 붙고(첫 실패한 가벼운 단계의 이름), 돌지 않은 단계는 줄에 낱말이 없어요. 같은 이름 단계(`bash_n:…`)는 첫 실패 코드 하나로 모아요. 시작하지 못한 프로그램은 127, 자리를 얻지 못하면 모든 cargo 단계에 자리의 코드(125 거절, 75 기계가 바쁨)가 남고 결과는 FAIL 이에요. 테스트 단계는 낱말 하나 `test=<코드>` 예요. 답에는 `runner`(`nextest`·`cargo-test`, 자리를 얻지 못했거나 무거운 단계 전에 멈췄으면 null)가 있고, `steps` 의 각 기록에는 `command`(실제 실행한 프로그램과 인자)가 있어요. `result=PASS` 는 모든 종료 코드가 0 일 때만이에요.
- 거절(아무것도 돌지 않음): 범위가 `<base>..<head>` 꼴이 아님, base·head 를 읽을 수 없음, 체크아웃이 head 가 아님, 추적 파일이 고쳐짐.

하지 않는 것: PASS 를 선언(독립 감사와 CI 는 그대로), 시험이나 도구를 고치거나 설치, 한 시험 안의 케이스 고르기, 실제 HOME 읽기.

## `check fast` 와 `check fix` (SPEC-20261125-check-fast-and-fix)

`check affected` 의 첫 무거운 단계는 `check fast` 의 `ledger` 단계예요(Rust 저장소에서만, 그리고 장부 시험 파일이 있을 때만): 손으로 맞추는 표를 빠뜨리면 실패하는 시험(`repo_scans_table`, `state_folder_literal_scan`, `source_manifest_facts`, `rule_history_guard`, `test_area_layout`, `git_env_scrub`, `ac_command_lint` 의 둘, 라이브러리 시험 둘)만 한 번의 cargo 호출로 돌려요. 이 단계가 `clippy` 앞에 있어서, 실패하면 `clippy` 와 `test` 는 시작하지 않고 답의 `skipped` 에 올라가며 요약 줄은 `stopped_at=ledger` 예요. 바뀐 파일 중 마지막 바이트가 줄바꿈이 아닌 일반 텍스트 파일이 있으면 가벼운 단계 `endline` 이 그 파일 이름을 말해요(깨끗한 범위에서는 계획에도 요약 줄에도 나오지 않아요). 넓은 변경도 같은 `ledger` 단계를 담아요.

- `check fast <base>..<head>`: 명령을 돌려요. `git diff --check`, `cargo fmt --check`, 바뀐 `src/ops/*.rs` 의 `rustfmt --check`, 표 검사(`test_tables`), 끝줄 검사(`endline`), 장부 시험(`ledger`)을 **모두** 돌리고, 하나가 실패해도 멈추지 않아요. 답의 `failures` 에 실패한 단계마다 종료 코드, 로그 경로, 실패한 시험 이름이 들어 있고 `result` 는 `PASS` 나 `FAIL` 이에요. `FAIL` 이면 종료 코드 1 이에요(`check affected` 와 같은 규칙). 체크아웃이 범위의 head 이고 추적 파일이 고쳐지지 않았을 때만 돌아요. 로그는 `<빌드 폴더>/check-fast-<head 12자>/`(`--out` 으로 바꿈)에 남고, `summary.txt` 는 쓰지 않아요: READY 영수증은 `check affected` 만 써요. 옵션은 `--dir`, `--slot-dir`, `--max`, `--min-free-gb`, `--out`, `--integration-ref` 예요.
- `check fix <base>..<head>`: 파일을 써요. 사람이 고를 것이 없는 사실만 써넣어요: `skills/moai/references/source-manifest.json` 의 sha256 값, `--area <낱말>=<영역>` 으로 영역을 준 새 시험 파일 첫 낱말의 `tests/areas.toml` 줄(`--area` 는 되풀이할 수 있고 `=` 가 없는 값은 사용법 오류예요), `tools/repo_scans.json` 의 `always` 항목의 새 시험 대상(`test_targets` 목록의 이름, 목록이 없는 항목은 이름 필터가 없는 검사 줄의 `--test` 옵션 하나; `gate::registry_only_grew` 가 더하기로 읽는 글일 때만). 이름 필터가 있는 검사 줄은 바꾸지 않아요. 나머지(허용 목록 행, 면제 이유, 매니페스트 항목의 `origin`, 영역을 주지 않은 새 첫 낱말, 검사 줄이 모두 필터 줄인 항목)는 쓰지 않고 더할 줄을 그대로 `print` 에 출력해요. 파일이나 그 위 폴더가 링크이면 거기에는 쓰지 않고 경로를 말해요. 쓴 뒤 같은 판정을 다시 물어서 `result` 는 더 출력할 줄이 없을 때만 `PASS` 이고, 두 번째 실행은 `nothing_to_do` 가 `true` 예요. 쓴 파일은 레인이 한 커밋으로 더하고 다시 돌려요. 옵션은 `--dir`, `--area` 예요.
