# `check`

operation `check.affected` · 코드 `src/ops/check.rs`, `src/check_affected.rs`, `src/check_affected/reuse.rs` · SPEC-20261120-check-affected, SPEC-20261120-merge-reuse

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
- 돌리기: 체크아웃이 범위의 head 이고 추적 파일이 고쳐지지 않았을 때만 돌아요. 어느 한 단계가 실패해도 나머지를 끝까지 돌리고, 단계마다 로그를 빌드 폴더(호출한 쪽의 `CARGO_TARGET_DIR`, 없으면 `<저장소>/target`)의 `check-affected-<head 12자>/`(`--out` 으로 바꿈)에 남겨요. 파이썬 단계도 같은 폴더의 `debug/moai-cli` 를 가리켜서, 바로 앞 `build_cli` 가 만든 바이너리를 써요. `--dir` 은 저장소 안의 어느 폴더든 되고, 바뀐 경로는 `--no-renames -z` 로 읽어서 이름이 바뀐 파일은 옛 경로도 세고 한글 이름도 따옴표 없이 읽어요.
- 답의 `summary` 한 줄: `check-affected base=<12> head=<12> targets=<n> py=<n> scans=<none|listed:n|unreadable> <단계>=<종료 코드>… result=PASS|FAIL`. 같은 이름 단계(`bash_n:…`)는 첫 실패 코드 하나로 모아요. 시작하지 못한 프로그램은 127, 슬롯이 단계를 시작하지 않으면 슬롯의 코드(125 거절, 75 기계가 바쁨)가 남고 결과는 FAIL 이에요. `result=PASS` 는 모든 종료 코드가 0 일 때만이에요.
- 거절(아무것도 돌지 않음): 범위가 `<base>..<head>` 꼴이 아님, base·head 를 읽을 수 없음, 체크아웃이 head 가 아님, 추적 파일이 고쳐짐.

하지 않는 것: PASS 를 선언(독립 감사와 CI 는 그대로), 시험이나 도구를 고치거나 설치, 한 시험 안의 케이스 고르기, 실제 HOME 읽기.

## `check reuse` (SPEC-20261120-merge-reuse) · 코드 `src/check_affected/reuse.rs`

`moai check reuse --base <OLD> --head <후보 head> --tip <지금 끝>`(또는 `--merge <머지 커밋>`: 부모가 정확히 둘이고, 첫째가 base·둘째가 head)는 감사를 마친 병합 후보의 증거가 **로컬 통합 가지의 끝이 움직인 뒤에도 맞는지** git 만 읽고 판단해요. 끝이 OLD 에서 움직이면 ff 가 안 되고 레인이 처음부터 다시 병합·검사하는데, 움직인 부분이 후보와 상관없으면 그 검사는 다시 할 필요가 없어요. `--dir` 은 저장소(기본: 현재 폴더)예요. 저장소 안의 하위 폴더를 줘도 최상위 폴더를 줄 때와 같은 답을 내요.

| `decision` | 뜻 | 해야 할 일 |
| --- | --- | --- |
| `fast_forward` | 끝이 OLD 그대로 | 증거 그대로, ff 로 올려요 |
| `already_merged` | 후보 head 가 이미 끝 안에 있음 | 올릴 게 없어요 |
| `reuse` | 끝만 앞으로 갔고, 병합이 깨끗하고, 움직인 경로가 후보의 경로와 겹치지 않고, `tools/repo_scans.json` 의 어느 항목도 한쪽은 규칙 파일을 바꾸고 다른 쪽은 그 범위 안 경로를 바꾸지 않았고, 후보에 대한 `check affected` 선택이 옛 트리와 끝 위 병합 트리에서 같거나(`selection_is_equal`) 더 작음(`selection_is_covered`) | 증거를 다시 쓰고 `reasons` 를 기록해요 |
| `rerun` | 새로 영향받은 검사가 있음 | `rerun` 의 것만 다시 돌려요: `rust`·`python`(선택에 새로 들어온 대상, 또는 움직인 경로의 영역이 고르는 대상 중 후보의 선택에 든 것), `steps`(움직인 경로가 후보의 경로와 겹칠 때만: `diff_check`·`fmt`·`clippy`·`test_lib`·`test_bins`, 겹친 파일의 `ops_rustfmt`·`bash_n:<경로>`), `scan_checks`(한쪽이 어떤 항목의 규칙 파일을, 다른 쪽이 그 범위 안 경로를 바꿨을 때 그 항목의 검사 명령, 이유 `scan_rules_cross_scope`, `moai repo-scans --refs` 의 `remerge` 와 같은 규칙) |
| `rerun_all` | 판단할 수 없음: `merge_conflict`, `base_not_ancestor_of_tip`, 읽을 수 없는 사실(`unreadable: …`), 읽을 수 없는 스캔 목록(`scan_table_unreadable: …`) | 모르면 전부 다시 돌려요 |

답에는 `reasons`(고정 순서), `rerun`, `overlap`(겹친 경로), `base`·`head`·`tip`(전체 sha), `candidate_paths`·`delta_paths`(개수), `selection.old`·`selection.new`(`rust`·`python`·`scans`), `merge_tree`(끝 위 병합 트리의 id, 병합이 깨끗할 때)가 있고, 같은 입력은 같은 답을 내요.

쓰기: 아무 ref·색인·작업 파일도 움직이지 않아요. `git merge-tree --write-tree` 가 만드는 트리 객체만 남아요(어느 커밋도 가리키지 않는 객체). 인자가 빠졌거나, 리비전이 아니거나, `-` 로 시작하거나, `--merge` 가 두 부모의 병합이 아니면 거절해요(결정을 내놓지 않음).

하지 않는 것: 병합 실행(잠금 아래 `git merge --ff-only`)과 노트 쓰기, 큐 읽기, 짝(`pair=`)·의존(`after=`) 묶기, 검사 실행. `reuse` 는 "선택이 같고 겹침이 없을 때"라는 규칙 아래에서 옛 증거가 유효하다는 뜻이지, 병합한 트리를 시험했다는 뜻이 아니에요: 서로 다른 영역이 합쳐져서만 깨지는 검사는 리더의 끝 점검이 잡아요. 이름 규칙 밖의 변경(`Cargo.toml`·`Cargo.lock`·`src/lib.rs`)은 `check affected` 와 같은 범위만 닿아요: 겹치는 파일이 없으면 `--lib`·`--bins`·`clippy`·`fmt` 증거는 다시 돌리지 않아요.
