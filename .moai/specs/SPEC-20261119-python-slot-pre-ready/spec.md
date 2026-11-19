---
id: SPEC-20261119-python-slot-pre-ready
title: "Whole Python unittest runs go through a python-only slot, and one command lists the whole-tree checks a head passes before READY"
status: draft
revision: 1
created: 2026-11-19
---

# Whole Python unittest runs go through a python-only slot, and one command lists the whole-tree checks a head passes before READY

## Intent

2026-11-19 에 레인 12개가 겹쳐 load average 가 100~140 까지 올랐어요(리더 관찰, 이 SPEC 에서 다시 재지 않음). cargo 는 `moai heavy-slot-acquire` 로 3개까지만 도는데, `python3 -m unittest discover` 전체 실행은 그 밖이라 레인마다 따로 돌았고 한 번에 38분이 걸렸어요. 같은 날 한 묶음이 몇 시간 READY 에 못 올라간 것은 전체 트리 점검이 잡았을 실패 6가지(fmt, SPEC Result 없음, source-manifest 항목 없음, repo_scans 표 항목 없음, 워크플로 파일 속 금지 문자열(`product_cli` 의 `skill_package`), windows 게이트) 때문이었어요. 조사 근거는 `tools/local_check.sh:377`(`step light unittest`), `tools/self_check.sh:202`(`--full-python` 도 `light`), 그리고 빌드한 `moai-cli` 가 없으면 `tests.repo_scans_test` 가 25개 중 19개를 헷갈리는 이유로 실패시킨다는 직접 실행(2026-11-19)이에요. 이 SPEC 은 (1) 파이썬 전체 실행을 파이썬 전용 슬롯(한 번에 2개)에 넣고, (2) `tools/self_check.sh --pre-ready` 가 READY 전에 돌릴 빠른 순서의 전체 트리 점검을 한 줄씩 보이게 해요.

## Non-goals

- cargo 슬롯(3칸)의 한도나 `src/heavy_slot.rs` 바꾸기 — 파이썬은 `--dir` 로 따로 두고 Rust 코드는 건드리지 않아요.
- `skills/` 아래 문구(self-check.md, verify.md, mode-agent.md)와 source-manifest — 다른 조각이 맡아요(리더 결정 2026-11-19).
- windows 형 검사를 스크립트 항목으로 두기 — 플랫폼 코드를 건드릴 때만 필요해서 문서 목록에만 둬요.
- 파이썬 시험 파일을 느림·빠름으로 나누거나 병렬로 돌리기 — 파일별 시간은 아직 재지 않았어요.

## Requirements

- REQ-1: When `tools/local_check.sh` runs its unittest step and `moai heavy-slot-acquire` exists, the script shall run the whole `unittest discover` inside the python slot, a slot folder of its own with at most `MOAI_PYTHON_SLOT_MAX` (default 2) runs at a time.
- REQ-2 (REQ-1): IF `moai heavy-slot-acquire` does not exist, THEN `tools/local_check.sh` shall run the unittest step directly as before.
- REQ-3: When `tools/self_check.sh --full-python` runs the whole `unittest discover`, it shall run it inside the python slot under the same folder and limit as REQ-1.
- REQ-4 (REQ-3): IF the python slot refuses to start (exit 75), THEN `tools/self_check.sh` shall print `slot-refused (not run)` for that item and count it as not_run, never as passed or failed.
- REQ-5: The python slot folder shall be the path in `MOAI_PYTHON_SLOT_DIR` when it is set and `${XDG_CACHE_HOME:-$HOME/.cache}/moai-cli/python-slots` otherwise, never the folder of the cargo slots.
- REQ-6: When `tools/self_check.sh --pre-ready` runs, it shall run only the whole-tree checks a head passes before READY, fast ones first, and print one `self-check: <item> exit=<code>` line per item and the usual summary line.
- REQ-7 (REQ-6): IF an item needs the built `moai-cli` and `MOAI_CLI_BIN` or `${CARGO_TARGET_DIR:-target}/debug/moai-cli` does not exist or is older than a file under `src/` or than `build.rs`, `Cargo.toml` or `Cargo.lock`, THEN the item shall print `not_run (no built moai-cli: <path>)` or `not_run (stale moai-cli: <path>)` and not run its test.
- REQ-8: When `--pre-ready` runs its items, the Python items shall run in the python slot and the cargo items in the heavy slot, as the other items of the script already do. · failure: none — 슬롯이 거절하는 경우는 REQ-4 의 not_run 규칙이 그대로 덮고, 슬롯 명령이 없으면 REQ-2 처럼 바로 돌아요.
- REQ-9 (REQ-6): IF `--pre-ready` is given together with `--full-python`, THEN `tools/self_check.sh` shall print the usage line and exit 2.
- REQ-10: Where `docs/ci-rc-ci-spec.md` describes `tools/local_check.sh`, it shall name the python slot, its two environment variables and the `--pre-ready` item list with each item's command.
- REQ-11: When `tests/local_check_test.py` and `tests/self_check_test.py` run, they shall pin the python-slot wrapping, the two environment defaults, the refusal handling, the `--pre-ready` item list and order, and the missing-binary rule, with stub tools in a temporary repository and a temporary HOME. · failure: none — 못박은 대상이 어긋나면 그 시험이 빨개질 뿐 따로 거부 경로가 없어요.
- REQ-12 (REQ-5): IF none of `MOAI_PYTHON_SLOT_DIR`, `XDG_CACHE_HOME` and `HOME` is set and non-empty, THEN each script shall print one line on its standard error for every Python item it runs and run that item directly, never in the folder of the cargo slots, and never stop on an unbound variable.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-5, REQ-12): local_check.sh 의 unittest 단계가 슬롯이 있으면 파이썬 슬롯 폴더와 한도 2 로 감싸이고, 없으면 그대로 돌아요. 폴더를 정할 수 없으면(HOME 없음) 안내 한 줄과 함께 바로 돌아요.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.local_check_test`
  verify: `LOCAL_CHECK_DRY_RUN=1 tools/local_check.sh`
- AC-2 (REQ-3, REQ-4, REQ-5, REQ-12): `--full-python` 도 같은 파이썬 슬롯 안에서 돌고, 슬롯 거부(exit 75)는 not_run 으로 읽히며, 폴더를 정할 수 없으면 안내 한 줄과 함께 바로 돌아요.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.self_check_test`
- AC-3 (REQ-6, REQ-8, REQ-9): `--pre-ready` 가 빠른 순서의 항목을 한 줄씩 내고, 파이썬 항목은 파이썬 슬롯·cargo 항목은 heavy 슬롯을 쓰며, `--full-python` 과 함께 주면 exit 2 예요.
  verify: `SELF_CHECK_DRY_RUN=1 tools/self_check.sh --pre-ready`
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.self_check_test`
- AC-4 (REQ-7): 빌드한 moai-cli 가 없으면 그 항목은 헷갈리는 실패가 아니라 not_run 줄을 내요.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.self_check_test`
- AC-5 (REQ-10): 문서가 파이썬 슬롯과 `--pre-ready` 목록을 명령과 함께 적어요.
  verify: `rg -n "python-slots|MOAI_PYTHON_SLOT_MAX|--pre-ready" docs/ci-rc-ci-spec.md`
- AC-6 (REQ-11): 두 시험 파일이 전부 통과해요.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.local_check_test tests.self_check_test`

## Scope

- write: `.moai/specs/SPEC-20261119-python-slot-pre-ready/spec.md`, `tools/local_check.sh`, `tools/self_check.sh`, `tests/local_check_test.py`, `tests/self_check_test.py`, `docs/ci-rc-ci-spec.md`
- read-only: `tools/heavy_check_slot.py`, `src/heavy_slot.rs`, `tests/heavy_check_slot_test.py`, `.github/workflows/ci.yml`, `skills/` 전부
- forbidden: 실제 HOME 과 원장, 네트워크, 실제 원격 저장소, 키, `skills/` 아래 파일

## Plan

1. 첫 RED: `tests/local_check_test.py` 와 `tests/self_check_test.py` 에 파이썬 슬롯 못박기와 `--pre-ready` 항목 목록 못박기를 먼저 쓰고 현재 `light` 단계 때문에 실패하는 것을 본다.
2. `tools/local_check.sh`: 새 단계 무게 `python` 을 더해 unittest 단계를 슬롯이 있으면 `--dir <python slot> --max <MOAI_PYTHON_SLOT_MAX>` 로 감싼다.
3. `tools/self_check.sh`: 같은 감싸기를 `--full-python` 항목에 적용하고, `--pre-ready` 묶음(diff-check, fmt, fmt-ops, manifest, spec-check, py-source-manifest, py-repo-scans, skill-package)과 빌드 없음 not_run 규칙을 더한다.
4. `docs/ci-rc-ci-spec.md` 의 `tools/local_check.sh` 절에 파이썬 슬롯과 `--pre-ready` 목록을 적는다.
5. 점검: 두 시험 파일, 두 스크립트의 dry-run, `git diff --check`, 빌드한 바이너리의 `spec check`.

## Risks

- 슬롯 폴더가 없을 때 `heavy-slot-acquire --dir` 가 폴더를 만들어 주는지는 확인했어요(2026-11-19 직접 실행: 없는 `a/python-slots` 가 만들어지고 `slot-0` 파일과 `waiting` 폴더가 생김). 그래서 스크립트는 `mkdir` 를 하지 않아요.
- 파이썬 시험 중 진짜 슬롯을 잡는 것은 `tests/heavy_check_slot_test.py:334` 하나이고 임시 `--dir` 를 써서 파이썬 슬롯과 부딪히지 않지만, `tests/local_check_test.py` 가 PATH 의 진짜 `moai` 를 쓰는지는 읽지 않았어요 — 첫 RED 에서 확인해요.
- `--pre-ready` 의 spec-check 항목은 `spec check --closure` 를 써요. 빌드한 바이너리로 확인했어요(2026-11-19 직접 실행): 초안 SPEC 은 `## Result` 가 있어도 `closure_status` 하나가 늘 나오고, `## Result` 가 없으면 `closure_result` 가 더해져요. 그래서 `closure_status` 만 통과로 읽고 다른 코드는 실패로 읽어요. 이 읽기는 오류 메시지의 `spec.md[:줄] <코드>` 모양에 기대요 — 그 모양이 바뀌면 코드를 못 읽어서 실패로 떨어져요(통과로 오해하지 않음). 시험은 가짜 `moai-cli` 의 출력을 쓰므로 실제 출력과의 일치는 이 직접 실행으로만 확인했어요.
- 항목 목록이 스크립트와 문서 사이에서 어긋날 수 있어서 시험이 목록과 순서를 못박아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 파이썬은 cargo 와 다른 슬롯 폴더에 한도 2 로 둔다(방법 A) | 리더 결정 2026-11-19 08:43Z; 파이썬 한 번(38분)이 cargo 슬롯 3칸 중 하나를 쥐지 않게 | 한도·폴더를 환경 변수로 바꾸거나 스크립트 래퍼 제거 |
| D-2 | READY 전 점검 목록은 `tools/self_check.sh --pre-ready` 로 둔다 | 리더 결정 2026-11-19 08:43Z; 기존 항목 이름·순서를 흔들지 않으려 새 플래그로 | 플래그와 시험 제거 |
| D-3 | windows 형 검사는 목록 문서에만 두고 스크립트 항목으로 만들지 않는다 | 플랫폼 코드를 건드릴 때만 필요하고 맥에서는 가짜 C 컴파일러가 있어야 해서 모든 PR 에 맞지 않아요 | 항목 추가 |

## Result

- AC-1: met — 2026-11-19 이 worktree 에서 `python3 -m unittest tests.local_check_test` 가 파이썬 슬롯 안에서 통과(`tests.self_check_test tests.local_check_test tests.repo_scans_test` 한 번에 Ran 130 tests, OK, 감사 수리 뒤). 새 10개가 기본 폴더·한도 2·`XDG_CACHE_HOME`·두 설정·빈 값·이름 접미사·거절(exit 75)·슬롯 명령 없음에 더해 HOME 없음/빈 HOME(안내 한 줄, 직접 실행, `unbound variable` 없음)·빈 `XDG_CACHE_HOME`·폴더 설정만 있는 경우를 못박아요. `LOCAL_CHECK_MODE=in-place LOCAL_CHECK_DRY_RUN=1 tools/local_check.sh` 의 unittest 단계 명령은 CI 못박기와 같은 `bash -c python3 -m unittest discover -v -s tests -p '*_test.py' …` 그대로예요(dry-run 은 슬롯 감싸기를 찍지 않으므로 감싸기 증거는 시험이에요).
- AC-2: met — `python3 -m unittest tests.self_check_test` 통과(위 130개 안, 새 PythonSlotTest 10개: 기본 폴더·한도, 터치한 파일 항목은 슬롯 없음, 두 설정, 빈 값, cargo 슬롯 폴더와 다름, 거절은 not_run, 슬롯 명령 없음, HOME 없음/빈 HOME, 빈 `XDG_CACHE_HOME`, 폴더 설정만).
- AC-3: met — `SELF_CHECK_DRY_RUN=1 tools/self_check.sh --pre-ready` 가 diff-check, fmt, fmt-ops, manifest, spec-check, py-source-manifest, py-repo-scans, skill-package 를 이 순서로 dry-run 줄로 찍어요. 시험 PreReadyTest 가 순서·슬롯 종류(파이썬 2·heavy 1)·`--full-python` 과 함께 주면 exit 2 를 못박아요. 진짜 실행(`tools/self_check.sh --pre-ready`, 빌드한 moai-cli, 진짜 슬롯)은 첫 번째(head d4b1a551cf)에 diff-check·fmt·fmt-ops·manifest·py-source-manifest·skill-package 가 exit=0 이었고, 이 SPEC 에 `## Result` 가 없던 `spec-check`(closure_result)와 `tests/self_check_test.py` 의 `rglob` 가 걸린 `py-repo-scans`(repo_scans 표 항목 없음) 두 개를 잡았어요 — 둘 다 이 목록이 잡으려던 종류의 실패라 고쳤어요. 고친 뒤 두 번째 실행(head 558cb9766b)은 여덟 항목이 모두 exit=0, `result=pass failed=none not_run=none` 이었어요. 그 뒤 감사 지적(HOME 없음) 수리를 더해 같은 실행을 다시 하고, 그 head 는 레인 보고 줄에 적어요.
- AC-4: met — `tests.self_check_test` 의 PreReadyTest 가 빌드 없음(`not_run (no built moai-cli: target/debug/moai-cli)`), 오래됨(`not_run (stale moai-cli: … is older than src/ops/a.rs)`), `MOAI_CLI_BIN` 지정, 바뀐 SPEC 없음 skip 을 못박아요. 실제 효과: 빌드 없이 돌린 `tests.repo_scans_test` 는 25개 중 19개가 `moai-cli not built` 로 실패했는데, 이 항목은 이제 그 경우 not_run 한 줄이에요.
- AC-5: met — `rg -n "python-slots|MOAI_PYTHON_SLOT_MAX|--pre-ready" docs/ci-rc-ci-spec.md` 가 31, 36, 54 줄을 찾아요(파이썬 슬롯 폴더와 두 환경 변수, 항목 표와 명령, `--full-python` 과 함께 줄 수 없음).
- AC-6: met — `python3 -m unittest tests.local_check_test tests.self_check_test` 통과(위 130개 안). 새 시험은 모두 37개이고, 처음 28개 중 26개는 예전 스크립트로 돌리면 실패했고(`git show` 로 꺼낸 첫 head 이전의 옛 `tools/self_check.sh`·`tools/local_check.sh` 로 확인: failures 9, errors 17), 변이 5개(자기 점검 파이썬 항목 슬롯 제거, 오래됨 판정 제거, `closure_status` 예외 제거, 두 플래그 거절 제거, local_check 의 unittest 슬롯 제거)는 모두 시험이 빨개졌어요.
- 감사 수리: 독립 감사(새 `moai-audit` 에이전트, head 558cb9766b)가 `HOME` 이 없으면 두 스크립트가 시작하자마자 `HOME: unbound variable` 로 죽는다고 지적했어요(B1, 재현: `env -i PATH=/usr/bin:/bin SELF_CHECK_DRY_RUN=1 bash tools/self_check.sh --full-python`). REQ-12 와 시험 6개(이전 커밋의 스크립트로는 6개 중 4개가 실패)로 고쳤어요. 같이 고친 메모: 오래됨 판정이 `Cargo.toml`·`Cargo.lock`·`build.rs`·`src/` 각각에서 N(오래됨)/N+1(새것)로 시험되고, 아직 커밋하지 않은 SPEC 과 SPEC 둘 중 하나만 실패하는 경우도 시험돼요. 문서에 한도는 `limits.json` 아래로만 낮출 수 있고 폴더는 절대 경로로 줘야 한다는 점을 적었고, `tools/self_check.sh` 머리 주석의 "both items" 를 바로잡았어요. 재감사(head d838586071)는 REQ-12 의 "cargo 슬롯 폴더로 되돌아가지 않는다"를 못박는 시험이 없다고 지적했어요(R1: 폴더 없음 가지를 cargo 슬롯 호출로 바꿔도 시험이 통과). HOME 없음 시험 둘이 어떤 슬롯 줄에도 `unittest` 가 없음을 단언하게 했고, 같은 변이(두 스크립트의 폴더 없음 가지를 `--dir` 없는 `heavy-slot-acquire` 호출로 교체)를 다시 돌리니 `PythonSlotTest` 와 `LocalCheckPythonSlotTest` 가 각각 10개 중 2개 실패(`FAILED (failures=2)`)했고 스크립트는 `cmp` 로 원상 복구했어요.
- 아직 안 한 것: `skills/` 아래 문구(self-check.md, verify.md, mode-agent.md)는 이 SPEC 밖(다른 조각). 파이썬 전체 실행 한 번의 소요는 다시 재지 않았어요(전체 discover 를 돌리지 않음). `tests/local_check_test.py` 가 PATH 의 진짜 `moai` 를 쓰는지 읽었어요: `FakeRepo` 는 `slot=True` 일 때만 가짜 `moai` 를 PATH 에 두고 그 밖에는 `moai` 가 없어요.
