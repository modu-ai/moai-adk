---
id: SPEC-20261121-rust-only-inventory
title: "Every tracked Python file has a verdict (port, covered or drop) before any Python is removed"
status: draft
revision: 1
created: 2026-11-21
---

# 파이썬 파일별 판정 목록

## Intent

사용자 결정(2026-11-21, 사용자의 말): "모든 python테스트 코드는 더이상 사용하지 않고 테스트는 rust로만 사용", 그리고 리더의 조사와 계획에 "진행해". 목표는 이 저장소의 테스트와 도구를 Rust 로만 두는 거예요. 지우기 전에 파이썬이 지키던 것을 잃지 않으려고, 이 조각은 추적 중인 파이썬 파일 84개(`git ls-files -- '*.py'`, 26386줄)를 파일마다 판정해요.

파일마다 셋 중 하나예요. **port**: Rust 시험이나 `moai` 명령으로 옮겨요(대체 이름을 적어요). **covered**: Rust 쪽이 이미 같은 사실을 단정해요(이름을 적어요). **drop**: 대체 없이 지워요(잃는 보호를 적어요). 표는 `inventory.md` 에 있고, 이 파일은 요구사항과 확인 방법과 결정할 것을 담아요.

조사 중에 카드 본문의 가정을 넘는 사실 셋이 나왔어요. 첫째, Rust 시험 일곱 곳이 파이썬을 실제로 실행해요(`tests/statusline_golden.rs` 는 `tools/statusline_golden.py` 를 그대로 부르고, `tests/atlas_three_arm.rs`·`tests/search_atlas.rs`·`tests/ac_command_lint.rs`·`tests/heavy_slot_cli.rs`·`tests/lane_env_cli.rs`·`tests/home_state_plan.rs` 도 그래요). 그래서 도구를 지우기 전에 이 시험을 먼저 바꿔야 해요. 둘째, `src/heavy_slot.rs:1318` 이 `include_str!` 로 `tools/heavy_check_slot.py` 를 컴파일에 담아서 그 파일은 코드 변경 없이 지울 수 없어요. 셋째, CI 단계가 파이썬 도구 네 개(`ci_scope`, `check_guard`, `candidate_summary`, `ci_superseded`)와 인라인 `python3 -c` 여러 곳을 불러서, 카드 기준 2("검사 단계가 파이썬을 안 돌린다")는 카드의 조각 3 이 말한 여섯 도구만으로는 이뤄지지 않아요.

## Non-goals

- 파일 삭제와 검사 단계 변경. 이 조각은 아무것도 지우지 않고 `tools/local_check.sh`·`check affected`·CI 를 그대로 둬요.
- 시험을 Rust 로 옮기는 일. 조각 2 와 3 이 해요.
- 사용자의 파이썬 프로젝트를 다루는 제품 기능(`src/bundle/verify.rs`, `src/tdd.rs`, `src/project_detect.rs` 등)은 이 작업의 대상이 아니에요.
- 제품 코드(`src/**`)와 `tools/**`, `tests/**` 변경.

## Requirements

- REQ-1: The inventory shall give every tracked Python file one row that names what it protects and who calls it and a verdict of port or covered or drop and the replacing Rust test or moai command · failure: none — a file without a row is found by the check in AC-1
- REQ-2: The inventory shall list every Rust test that runs a Python file or reads its name or content with the action the deleting slice owes it · failure: none — a missing Rust file is found by the check in AC-2
- REQ-3: The inventory shall list the Python use that sits outside Python files in shell scripts and workflow files with a replacement direction · failure: none — a missing file is found by the check in AC-3
- REQ-4: The inventory shall give a risk order for the six workflow tools with one reason per tool · failure: none — a missing tool is found by the check in AC-4
- REQ-5: When a verdict is drop, the row shall name the protection that is lost or point to the tool row that names it · failure: none — an unexplained drop is found by the check in AC-5
- REQ-6: This slice shall delete no file and change no check step · failure: none — a changed path outside the SPEC folder is found by the check in AC-6
- REQ-7 (REQ-1): IF a verdict cannot be decided from the files alone, THEN the SPEC shall record a decision for the leader with its options and its default · failure: none — a decision without a default is found by the check in AC-7

## Acceptance

- AC-1 (REQ-1): 추적 중인 파이썬 파일 84개가 표에 정확히 한 번씩 있어요.
  verify: `git ls-files -- '*.py' | while read -r f; do n=$(grep -c "^| .$f. | [0-9]*줄" .moai/specs/SPEC-20261121-rust-only-inventory/inventory.md); [ "$n" = 1 ] || echo "$f $n"; done` 이 아무것도 내지 않아요.
- AC-2 (REQ-2): 파이썬을 실행하는 Rust 시험 파일이 모두 3-1 표에 있어요.
  verify: `git grep -l -E 'Command::new\("python' -- 'tests/*.rs' | while read -r f; do grep -q "$f" .moai/specs/SPEC-20261121-rust-only-inventory/inventory.md || echo "$f"; done` 이 아무것도 내지 않아요.
- AC-3 (REQ-3): 셸과 워크플로에서 파이썬을 부르는 파일이 모두 4절에 있어요.
  verify: `grep -l -E 'python3?\b' .github/workflows/*.yml tools/*.sh tools/smoke/run.sh scripts/benchmark.sh | while read -r f; do grep -q "$(basename "$f")" .moai/specs/SPEC-20261121-rust-only-inventory/inventory.md || echo "$f"; done` 이 아무것도 내지 않아요.
- AC-4 (REQ-4): 여섯 도구가 모두 순서표에 있어요.
  verify: `for t in mutate source_manifest merge_driver_manifest smoke build_folder heavy_check_slot; do grep -q "^| [1-5] | .*$t" .moai/specs/SPEC-20261121-rust-only-inventory/inventory.md || echo "$t"; done` 이 아무것도 내지 않아요.
- AC-5 (REQ-5): drop 행마다 잃는 보호나 도구 행 참조가 있어요.
  verify: `grep -E '^\| .* \| drop \|' .moai/specs/SPEC-20261121-rust-only-inventory/inventory.md | grep -v -e '잃는 보호' -e '도구와 함께' -e '쓸 곳이 없어요'` 이 아무것도 내지 않아요.
- AC-6 (REQ-6): 이 조각이 바꾼 경로는 SPEC 폴더 안뿐이고 지운 파일이 없어요.
  verify: `git diff --name-status $(git merge-base HEAD v2026.11.0)..HEAD | grep -v -E '^A	.moai/specs/SPEC-20261121-rust-only-inventory/'` 이 아무것도 내지 않아요.
- AC-7 (REQ-7): 결정 D1 부터 D7 까지 모두 기본값이 있어요.
  verify: `grep -c -E '^- D[1-7] .*기본값:' .moai/specs/SPEC-20261121-rust-only-inventory/spec.md` 가 7 을 내요.

## Scope

- write: `.moai/specs/SPEC-20261121-rust-only-inventory/**`
- read-only: 추적 중인 모든 파일(`git ls-files`), `tools/repo_scans.json`, `tools/local_check.sh`, `tools/self_check.sh`, `.github/workflows/*.yml`, `src/check_affected.rs`, `src/gate.rs`
- forbidden: 다른 모든 경로 쓰기, 실제 HOME, 네트워크 쓰기

## Plan

조각마다 SPEC 하나와 착지 하나예요. 이 SPEC 은 조각 1 이에요.

1. 조각 1(이 SPEC): 파일별 판정표와 결정 목록. 삭제 없음.
2. 조각 2: 파이썬에만 있던 보호를 Rust 시험으로 옮기고(문서·CI 글 점검, 설치 스크립트 점검), `tools/local_check.sh`·`tools/self_check.sh`·`check affected`·CI 에서 파이썬 단계와 `setup-python`·PyYAML 설치를 걷어요. CI 가 부르는 도구 네 개를 여기에 넣을지는 D3.
3. 조각 3: 여섯 워크플로 도구를 `moai` 명령으로, 도구마다 한 조각씩. 순서는 `inventory.md` 5절의 제안과 D1.
4. 조각 4: 조각 2·3 이 대체한 파이썬 시험·도구와 drop 으로 표시한 것을 지워요. 지우기 전에 3절의 Rust 쪽 줄(실행·`include_str!`·보호 경로 목록·문구 단정)을 먼저 바꿔요.

## Decisions

결정은 모두 되돌릴 수 있고 범주는 `none` 이에요. 결정자는 리더이고, 기한은 해당 조각을 시작하기 전이에요. 그때까지 답이 없으면 기본값으로 진행해요.

- D1 조각 3 의 순서와 단위. 선택지: (a) 위험 큰 것부터 `mutate` → `source_manifest`+`merge_driver_manifest` → `smoke` → `build_folder` → `heavy_check_slot`, (b) 값싼 것부터 거꾸로, (c) 매니페스트 도구 둘을 따로. 기본값: (a), 매니페스트 도구 둘은 서로를 불러서 한 조각.
- D2 drop 17개를 그대로 지울지. 선택지: (a) 목록대로 지움(필요하면 나중에 `moai` 명령으로 새로 만듦), (b) 일부를 port 로 바꿈. 기본값: (a).
- D3 CI 가 부르는 도구 네 개(`ci_scope`, `check_guard`, `candidate_summary`, `ci_superseded`)의 위치. 선택지: (a) 조각 2 에 포함, (b) 조각 2 와 3 사이에 조각 하나를 더 둠, (c) 조각 3 에 합침. 기본값: (b), 카드 기준 2 는 그 조각에서 이뤄져요.
- D4 워크플로·이슈 양식 YAML 점검 방식. 선택지: (a) 줄 단위 텍스트 점검으로 옮김(새 크레이트 없음), (b) YAML 크레이트 하나를 들임(`Cargo.lock` 에 없고 의존성 최소 원칙에 맞는 근거가 필요). 기본값: (a), 줄 단위로 표현이 안 되는 점검이 나오면 그때 이 결정으로 되돌아와요.
- D5 파이썬이 들어 있는 셸 스크립트 두 개(`scripts/benchmark.sh` 3곳, `tools/local-rc1-isolation.sh` 8곳). 선택지: (a) 한 줄짜리 JSON 읽기는 `moai` 출력으로 바꾸고 `benchmark.sh` 는 쓰는 사람이 있는지 확인한 뒤 결정, (b) 둘 다 지움, (c) 그대로 둠(이유를 적어 사용자 승인). 기본값: (a).
- D6 `check_guard` 의 신뢰 모델. 지금은 CI 가 베이스 커밋의 사본을 실행해서 PR 이 자기 점검을 약화하지 못해요. 선택지: (a) check-guard 잡이 베이스 커밋을 빌드해 그 `moai` 를 실행, (b) 설치된 릴리스 바이너리를 믿음, (c) 이 점검을 파이썬으로 남김(이유 기록). 기본값: (a).
- D7 Codex 지원 제거와 겹치는 파일(`tools/host_live_harness.py`, `tests/e2e_host_codex_verdict_test.py`, `tests/host_matrix_test.py` 의 Codex 행). 선택지: (a) 제거 작업이 먼저 정리하길 기다림, (b) 이 작업이 먼저 지움. 기본값: (a), 조각 4 시점에도 남아 있으면 그때 지워요.

## Risks

- 판정은 파일 머리말·시험 이름·호출 위치를 읽은 결과예요. 시험 본문을 사례별로 대조한 것이 아니라서, 조각 2·3 은 옮기기 전에 사례를 하나씩 대조해야 해요. 이 목록의 `covered` 두 줄(`tools/heavy_check_slot.py`, `tests/heavy_check_slot_test.py`)도 Rust 쪽에 같은 사례가 없는 것이 남아 있을 수 있어요.
- `build_folder.py clean` 의 Rust 대응을 찾지 못했어요. `src/` 를 `clean` 으로 훑은 결과가 없었다는 뜻이고, 다른 이름으로 있을 수 있어서 조각 3 에서 다시 확인해요.
- 파이썬 시험 1050개가 지워지는 일이라 조각마다 "시험 수가 줄어든 것이 보호를 잃은 것이 아님"을 대조표로 보여야 해요.

## Result

(감사 뒤에 채워요.)
