---
id: SPEC-20261124-test-cost-profile-git-path
title: "Test cost split: process start, build profile and git path measured, nothing adopted"
status: draft
revision: 1
created: 2026-11-24
---

# Test cost split: process start, build profile and git path measured, nothing adopted

## Intent

테스트 한 개의 평균 시간(약 0.77초 × 약 4000개)이 어디에 쓰이는지 재서, 프로세스 띄우기를 함수 호출로 바꾸는 일(in-process `cli()`)과 빌드 프로필 최적화, 진짜 git 경로 중 무엇이 20% 이상 줄이는지 가려요. 측정 카드라서 저장소 코드는 바꾸지 않고, 숫자와 방법과 한계를 아래 Result 에 남겨 다음 카드(C-473)가 다시 쓰게 해요.

## Requirements

- REQ-1: When a seeded sample of 30 tests is timed, the report shall give each test's wall time split into wrapper, test-binary start, moai start, moai body, git and the rest.
- REQ-2 (REQ-1): IF a sampled test cannot be timed with the logging shims, THEN the report shall name that test and the test that replaced it.
- REQ-3: When a build-profile variant or a git-path variant is compared with the base, the report shall use runs that alternate the variants over the same tests and give the sum, the ratio to the base and the compile cost.
- REQ-4 (REQ-3): IF a variant's runs did not alternate with the base or a repetition failed, THEN the report shall mark that variant's numbers as not usable.
- REQ-5: While no variant lowers the sum of the sampled wall times by 20 percent or more, the repository shall keep its code unchanged.

## Acceptance

- AC-1 (REQ-1, REQ-2): the Result gives the stage split of the 30 sampled tests, the share of each stage, and the one test that was replaced · verify: `moai spec check SPEC-20261124-test-cost-profile-git-path --project .`
- AC-2 (REQ-3): the Result gives, for each variant, the sums and the ratios to the base from alternating runs, and the compile time and disk of each variant · verify: `moai spec check SPEC-20261124-test-cost-profile-git-path --project .`
- AC-3 (REQ-4, REQ-5): every repetition passed and no variant reaches 20 percent, and the only file this branch changes is this SPEC · verify: `git diff --name-only NeXT...HEAD`

## Scope

- write: `.moai/specs/SPEC-20261124-test-cost-profile-git-path/spec.md`

## Plan

1. 표본 30개(seed 476)를 `moai`·`git` 앞에 끼운 시간 기록 shim 으로 재서 단계별로 쪼개요(1단계).
2. 빌드 프로필 3가지와 진짜 git 경로 1가지를 같은 표본에서 번갈아 재요(2단계).
3. 20% 에 못 미치면 코드는 되돌리고 이 문서에 숫자만 남겨요.

## Result

머리: 이 SPEC 만 바뀌는 문서 커밋이에요(코드 변경 없음). 측정은 2026-11-24 에 16 CPU Mac(arm64)의 디버그 빌드에서 했고, 부하(load1)는 12~50 이었어요. 측정 도구와 원자료는 레인 워크트리의 `target/c476/` 에 있고 카드가 닫히면 사라져요(표본 목록 `sample.tsv`, 시간 기록 shim `shim.rs`, 실행기 `runner.py`, 번갈아 재기 `ab2.py`). 카드 노트 8620(1단계), 8652(2단계, 되돌린 git 경로 패치 포함)에 같은 내용이 있어요.

### 1단계: 테스트 한 개의 시간을 쪼갠 표

방법: 239개 후보 파일(`MOAI_BIN` 이나 `common::cli` 를 쓰고 테스트가 3~40개이며 `[solo]`·card·ledger 가 아닌 것)에서 seed 476 으로 10개 파일을 고르고 파일마다 3개를 골라 30개를 만들었어요. 각 테스트를 nextest 와 같게 `tools/test_home_guard.sh <테스트 실행 파일> --exact <이름> --nocapture` 로 하나씩 3회 돌려 중앙값을 냈어요. `moai`, `moai-cli`, `git` 앞에 시간 기록 shim 을 끼워 프로세스마다 시작·끝 시각을 같은 시계로 남겼고, `moai` shim 은 자식 `PATH` 앞에 git shim 폴더를 넣어서 환경을 비운 테스트에서도 CLI 안의 git 호출이 잡혀요. 단계 정의는 이래요.

- wrapper: 빈 HOME 을 만들고 지우는 래퍼가 `true` 를 돌리는 시간.
- testbin: 테스트 실행 파일이 테스트 없이 떴다 끝나는 시간.
- moai_start: 최상위 `moai` 호출 수 × `moai --version` 시간(22ms). 다른 `moai` 안에서 뜬 `moai` 는 바깥 호출의 본체로 쳐요.
- moai_body: `moai` 수명 합 − moai_start − `moai` 안의 git.
- git: git 수명 합(직접 부른 것과 `moai` 안의 것).
- hops: shim 이 끼운 시간 추정(한 단계 7ms).
- rest: 나머지, 곧 테스트 코드의 픽스처·TempDir·단언.

바닥값: `moai --version` 22ms, shim 한 단계 7ms, 래퍼(`true`) 59ms.

| 묶음 | 합 | 평균 | 중앙값 | wrapper | testbin | moai_start | moai_body | git | hops | rest |
|---|---|---|---|---|---|---|---|---|---|---|
| 30개 전부 | 99.51s | 3.317s | 0.713s | 1.8% | 0.3% | 4.7% | 23.5% | 9.5% | 1.6% | 58.0% |
| 60초 이상치 1개를 뺀 29개 | 39.40s | 1.359s | 0.630s | 4.3% | 0.7% | 11.7% | 50.0% | 24.1% | 4.0% | 3.4% |

- 이상치는 `home_state_selective::an_open_file_keeps_its_entry_in_place` 예요(rest 약 56~57s, 60초 근처에서 끝나는 대기로 보이고 원인은 파지 않았어요). 이것과 `more_than_five_hundred_entries_block_the_plan`(15초, moai 본체 12.8s)은 C-473 이 맡아요.
- 테스트 1개당 평균 `moai` 호출 7.1회, git 호출 13.2회예요. `moai` 호출 1회는 평균 114ms(띄우기+본체), git 호출 1회는 평균 24ms(shim 7ms 포함)예요.
- 가장 가벼운 실제 명령(`apply_recover_partial` 3개)의 `moai` 호출은 33~38ms 라서 호출당 고정비는 22~33ms 예요. 7.3회 × 22~33ms 는 평균 1.36s 의 12~18% 라서, 호출을 함수 호출로 바꾸는 in-process `cli()` 로 줄일 수 있는 상한이 그 정도예요. 같은 영역을 3배 빠르게 하는 일은 닿을 수 없어요.
- 표본에서 바뀐 테스트: `update_moved_receipts` 는 테스트가 바이너리를 복사해서(복사본이 shim 이면 제품 바이너리가 아니에요) 표본에서 빼고 `project_scale_detect_lib` 로 바꿨어요.
- 별도 실측: `/usr/bin/git --version` 중앙값 20.1ms, 개발자 도구 폴더의 `git` 직접 호출 13.6ms (각 30회, 부하 31.5).

### 2단계: 빌드 프로필과 git 경로를 번갈아 잰 표

방법: 같은 표본에서 60초 대기 테스트 1개를 빼고 29개를, 변형 바이너리와 변형 테스트 실행 파일을 번갈아 바꿔 가며 테스트마다 5회씩 돌려 중앙값을 냈어요(래퍼를 거쳐 하나씩, shim 없이). `moai` 와 `moai-cli` 는 변형마다 한 파일로 두고 심볼릭 링크만 바꿔요. 변형:

- v1: `profile.dev.package."*".opt-level=2`(의존성만 최적화).
- v2: `profile.dev.opt-level=1`(제품 크레이트까지 전부; test 프로필은 dev 를 이어받아요. `[profile.test]` 만 따로 바꾼 실행은 하지 않았어요).
- v3: `profile.dev.package.libsqlite3-sys.opt-level=3` 와 `rusqlite` 3.
- n1: 기준선 + 진짜 git 경로(`tests/common/mod.rs` 의 `clean_env` 가 개발자 도구 폴더의 `git` 을 `PATH` 앞에 두고, `tests/common/git.rs` 가 그 `git` 을 경로로 부르고, 래퍼가 `DEVELOPER_DIR/usr/bin` 을 `PATH` 앞에 둬요). 노트 8652 에 패치가 있어요.

| 실행 | 변형 | 29개 합 | 기준선 대비 | `more_than_five_hundred_entries` 뺀 28개 합 | 기준선 대비 |
|---|---|---|---|---|---|
| B | base | 28.53s | 1.000 | 18.59s | 1.000 |
| B | v1 | 28.06s | 0.983 | 17.86s | 0.961 |
| B | v2 | 26.64s | 0.934 | 17.91s | 0.964 |
| B | v3 | 28.61s | 1.003 | 18.70s | 1.006 |
| A2 | base | 21.75s | 1.000 | 13.12s | 1.000 |
| A2 | v1 | 21.72s | 0.999 | 12.49s | 0.952 |
| C | base | 18.19s | 1.000 | 10.42s | 1.000 |
| C | n1 | 20.18s | 1.110 | 10.58s | 1.016 |

- 테스트별 중앙값 비(0.3초 넘는 테스트만): v1 0.95~0.97, v2 0.93~0.95, v3 1.00, n1 1.02~1.03. 모든 반복이 통과했어요.
- 가장 느린 `more_than_five_hundred_entries_block_the_plan` 은 base 9.9s, v1 10.2s, v2 8.7s, v3 9.9s 로 프로필을 바꿔도 안 줄어요. 반면 CPU 를 쓰는 `moai help --all` 10회 합은 base 0.16s, v1 0.09s, v2 0.06s, v3 0.18s 로 빨라져요(`time` 한 번씩, 거친 값). 그래서 테스트 시간은 `moai` 코드의 CPU 가 아니라 파일 입출력이나 자식 프로세스 같은 데서 써요. 그 안은 쪼개지 못했어요.
- 컴파일 시간과 디스크(`cargo test --no-run` 의 `Finished ... in`, 표본 크레이트 11개, 부하 12~27; 디스크는 `target/debug` KB):

| 변형 | 차가운 빌드 | sccache 적중 빌드 | 적중 빌드 후 디스크 |
|---|---|---|---|
| base | 46.2s (`SCCACHE_RECACHE=1`, 깨끗한 폴더) | 14.3s (깨끗한 폴더) | 672,300 (차가운 빌드 직후 808,604) |
| v1 | 1m24s (옛 산출물 위) | 18.3s (깨끗한 폴더) | 656,368 |
| v2 | 1m57s (옛 산출물 위) | 24.7s (깨끗한 폴더) | 565,344 |
| v3 | 1m01s (옛 산출물 위) | 12.4s (옛 산출물 위) | 깨끗한 폴더로는 안 쟀어요 |

  v1~v3 의 차가운 빌드는 기준선과 같은 조건이 아니라 참고용이에요.

판정: 20% 이상 줄이는 변형이 없어요(최대 7%). 그래서 프로필과 git 경로 변경은 남기지 않고 코드를 되돌렸어요. 남은 지렛대는 `moai` 본체 안쪽(약 50%)과 테스트 쪽 대기이고, C-473(느린 테스트 상위)이 이어 받아요.

### 측정 방법 (다음 측정이 다시 쓰는 교훈)

- 번갈아 재야 해요. 변형을 차례로 따로 재면 부하 흔들림이 변형 차이로 보여요. 이번 처음 시도에서 기준선 1회가 31.52s(부하 32), 같은 기준선 2회가 25.66s(부하 38)로 기준선끼리 19% 달랐고, 의존성 최적화는 0.60배로 보였지만 번갈아 재니 0.98배였어요. 비교는 한 실행 안에서 번갈아 잰 값만 써요.
- 첫 실행 비용을 빼야 해요. 번갈아 잴 때 바이너리를 매번 복사하면 새로 만든 파일의 첫 실행 검사 비용이 양쪽에 더해져서 차이가 가려졌어요. 변형마다 파일 하나를 두고 심볼릭 링크만 바꾸며, 재기 전에 한 번씩 미리 실행해요.
- 환경을 비우는 테스트(`clean_env`)는 `PATH` 와 로그 환경 변수를 잃어요. shim 은 자기 폴더에서 로그 경로를 찾고 자식 `PATH` 앞에 git shim 폴더를 넣어야 그 테스트의 git 호출이 잡혀요.
- 바이너리를 복사하는 테스트(`update_moved_receipts` 류)는 shim 으로 못 재요.
- 프로세스가 프로세스를 부르면 바깥 호출만 세요. 안쪽 `moai` 를 따로 세면 띄우기를 두 번 세요.
- 스택 샘플을 뜰 때는 프로세스 이름만으로 `pid` 를 고르지 말고 테스트가 띄운 프로세스인지 부모로 확인해요(이번에는 다른 세션의 `moai` 를 잡아서 샘플을 버렸어요).

### 한계 (not_run 포함)

- `moai` 본체 약 50% 안의 시간이 어디로 가는지(입출력, 자식 프로세스, clap 구성)는 못 밝혔어요(not_run).
- macOS arm64 디버그 빌드, 이 머신의 부하 12~50 에서만 쟀어요. Linux·Windows·릴리스 빌드는 안 쟀어요(not_run).
- 표본은 239개 후보 파일 중 10개 파일이라 전체 약 4000개를 대표한다고 말할 수 없어요.
- git 시간은 shim 이 `PATH` 로 잡은 호출만이라 직접 경로로 부른 호출은 빠졌을 수 있어서 하한이에요.
- CI 영향(캐시 키, 컴파일 시간)은 재지 않았어요. `[profile.test]` 만 따로 바꾼 변형은 재지 않았어요.

### 기준

- AC-1: met — 위 1단계 표와 비율(노트 8620), 바꾼 테스트 1개를 적었어요.
- AC-2: met — 위 2단계 표(실행 B, A2, C)와 컴파일 시간·디스크 표를 적었어요.
- AC-3: met — 모든 반복이 통과했고 최대 7% 라서 20% 에 못 미쳐요. 이 브랜치가 바꾸는 파일은 이 SPEC 뿐이에요(커밋 뒤 `git diff --name-only NeXT...HEAD` 로 읽어 확인해요).
