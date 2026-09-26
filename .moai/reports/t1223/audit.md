# t1223 감사 — census 픽스처 totals 보강

- 대상: 커밋 `2990924ef` (브랜치 `WT-census-fixture-totals`), 기준 `6ae4583e6`
- 감사자: sync-auditor (독립·회의적 감사, 구현은 읽기 전용)
- 변이는 모두 스크래치 사본(`.../scratchpad/t1223-audit/mut/<트리>-<변이명>/`)에서만 만들었다. 추적 파일은 이 보고서 말고는 건드리지 않았다(`git status --short` 출력 없음).

## 판정

**PASS** — 차단 결함 없음. 조화평균 **93.6**

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 95 | PASS | 주장 1~5 모두 재현됨. 보강 전 m3·m5 exit 0, 보강 후 exit 1이고 합계 줄이 각각 `passed=3`, `nothing-ran=2`. m1·m2도 exit 1. 기대 출력은 독립 jq 계산과 일치 |
| Security (25%) | 98 | PASS | 테스트 데이터와 주석만 바뀜. 실행 경로·비밀·입력 처리 변경 없음. `test-census.sh`, `.github/` diff 0 |
| Craft (20%) | 88 | PASS | 목표 변이 두 개를 잡음. 다만 「패키지 단위로 세기」 변이 3개(x5~x7)는 여전히 살아남음(범위 밖, 참고) |
| Consistency (15%) | 94 | PASS | 필드 구성이 실제 `go test -json`과 이벤트 종류별로 같음. 타임스탬프 단조 증가, 그리스 문자 패키지 이름 관례 유지. 실제 스트림이 내는 출력 줄 3개가 빠졌지만 영향 없음을 측정으로 확인 |

조화평균: 4 / (1/95 + 1/98 + 1/88 + 1/94) = 93.6. must-pass 차원(Functionality, Security)은 모두 기준을 넘는다.

## 주장별 검증

하네스 `run.sh <base|head> <이름> <sed식|NONE>`은 해당 커밋의 `scripts/ci-census`를 `git archive`로 받은 tar를 새 디렉터리에 풀고, 사본의 `test-census.sh`에 sed 변이를 넣는다. 변이가 실제로 적용됐는지 `cmp`로 확인하고(적용 안 되면 exit 3), 그다음 `census-check.sh`를 실행한다.

### 주장 1 — 기준 커밋에서 m3·m5가 살아남는다: **확인**

```
$ bash run.sh base clean NONE
[base/clean] census-check exit=0
    census-check: PASS (census output matches …/mut/base-clean/base/scripts/ci-census/testdata/expected.txt)

$ bash run.sh base m3 '151s/ and (.Test \/\/ null) != null//'
[base/m3] mutated line(s):
    < passed=$(count 'select(.Action=="pass" and (.Test // null) != null) | "\(.Package)\t\(.Test)"')
    > passed=$(count 'select(.Action=="pass") | "\(.Package)\t\(.Test)"')
[base/m3] census-check exit=0
    census-check: PASS (…)

$ bash run.sh base m5 '153s/== null)/!= null)/'
[base/m5] mutated line(s):
    < nothing_ran=$(count 'select(.Action=="skip" and (.Test // null) == null) | .Package')
    > nothing_ran=$(count 'select(.Action=="skip" and (.Test // null) != null) | .Package')
[base/m5] census-check exit=0
    census-check: PASS (…)
```

### 주장 2 — 보강 후 깨끗한 census는 PASS, m3·m5·m1·m2는 FAIL: **확인**

```
$ bash run.sh head clean NONE
[head/clean] census-check exit=0
    census-check: PASS (…)

$ bash run.sh head m3 '151s/ and (.Test \/\/ null) != null//'
[head/m3] census-check exit=1
    -=== totals: packages=5 passed=2 skipped=2 nothing-ran=1 failed=1 build-failed=1 ===
    +=== totals: packages=5 passed=3 skipped=2 nothing-ran=1 failed=1 build-failed=1 ===

$ bash run.sh head m5 '153s/== null)/!= null)/'
[head/m5] census-check exit=1
    -=== totals: packages=5 passed=2 skipped=2 nothing-ran=1 failed=1 build-failed=1 ===
    +=== totals: packages=5 passed=2 skipped=2 nothing-ran=2 failed=1 build-failed=1 ===

$ bash run.sh head m1 '141s/select(.Action=="skip")/select(.Action=="skip" and (.Test \/\/ null) != null)/'
[head/m1] census-check exit=1
    -NOTHING RAN   example.com/censusfix/beta

$ bash run.sh head m2 '171s/sed .s\/%\/%25\/g./cat/'
[head/m2] census-check exit=1
    -::notice title=AC snapshot absent::.moai/specs/SPEC-FIXTURE-100%25/acceptance.md: HALT AC-FIXTURE-001 AC-FIXTURE-002 (…)
    +::notice title=AC snapshot absent::.moai/specs/SPEC-FIXTURE-100%/acceptance.md: HALT AC-FIXTURE-001 AC-FIXTURE-002 (…)
```

작업트리에서 직접 실행해도 PASS다.

```
$ bash scripts/ci-census/census-check.sh
census-check: PASS (census output matches /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1223/scripts/ci-census/testdata/expected.txt)
```

### 주장 3 — `expected.txt`가 올바른 census 출력과 일치한다: **확인 (독립 계산)**

census 스크립트를 재사용하지 않고, `jq -s`로 합계를 따로 계산했다.

```
$ jq -s -r '{packages: ([.[]|.Package|select(.!=null)]|unique|length), passed: ([.[]|select(.Action=="pass" and .Test!=null)|[.Package,.Test]]|unique|length), pkg_level_pass: …, skipped: …, pkgs_with_skipped_test: …, nothing_ran: …, failed: …, build_failed: …, skip_rows: …} | tostring' scripts/ci-census/testdata/fixture.jsonl
{"packages":5,"passed":2,"pkg_level_pass":["example.com/censusfix/epsilon"],"skipped":2,"pkgs_with_skipped_test":["example.com/censusfix/alpha","example.com/censusfix/epsilon"],"nothing_ran":["example.com/censusfix/beta"],"failed":1,"build_failed":["example.com/censusfix/delta"],"skip_rows":["example.com/censusfix/beta <none>","example.com/censusfix/alpha TestAlphaSkips","example.com/censusfix/epsilon TestEpsilonSkips"]}
```

- 합계 `packages=5 passed=2 skipped=2 nothing-ran=1 failed=1 build-failed=1`은 `expected.txt` 15행과 같다.
- skip 행 정렬(`NOTHING RAN beta` → `SKIPPED TEST alpha` → `SKIPPED TEST epsilon`)은 바이트 순서로 결정된다. `N`(0x4E) < `S`(0x53)이고, 같은 접두 뒤에서 `a` < `e`이므로 로케일과 관계없이 같은 순서가 나온다.
- epsilon에는 fail·build-fail 이벤트와 `absent-from-snapshot` 출력이 없다. 그래서 BUILD FAILED, FAILED, FAILED PKG, notice 절은 바뀌지 않는다. `git show 2990924ef`의 `expected.txt` diff도 SKIPPED 행 하나 추가와 합계 줄 교체뿐이다.
- 설계 의도도 수치로 확인된다. 패키지 수준 pass는 epsilon 1건이다(m3에서 passed 3). skip 테스트를 가진 패키지는 2개, 테스트 파일이 없는 패키지는 1개다(m5에서 nothing-ran 2).

### 주장 4 — 추가한 epsilon 이벤트가 실제 `go test -json` 모양이다: **확인 (누락 3줄, 무해)**

스크래치 모듈(`realmod/`, `go version go1.26.8 darwin/arm64`)에 통과 테스트 하나와 `t.SkipNow()` 테스트 하나를 가진 `epsilon`, 테스트 파일이 없는 `beta`를 만들었다.

실제 스트림의 이벤트 종류별 키(Time 제외)는 다음과 같다.

```
$ jq -c 'select(.Package=="example.com/censusfix/epsilon") | {a:.Action, test:(has("Test")), keys:(keys|map(select(.!="Time"))|sort), out:(.Output//null)}' real-stream.jsonl
{"a":"start","test":false,"keys":["Action","Package"],"out":null}
{"a":"run","test":true,"keys":["Action","Package","Test"],"out":null}
{"a":"output","test":true,"keys":["Action","Output","Package","Test"],"out":"=== RUN   TestEpsilonPasses\n"}
{"a":"output","test":true,"keys":["Action","Output","Package","Test"],"out":"--- PASS: TestEpsilonPasses (0.00s)\n"}
{"a":"pass","test":true,"keys":["Action","Elapsed","Package","Test"],"out":null}
{"a":"run","test":true,"keys":["Action","Package","Test"],"out":null}
{"a":"output","test":true,"keys":["Action","Output","Package","Test"],"out":"=== RUN   TestEpsilonSkips\n"}
{"a":"output","test":true,"keys":["Action","Output","Package","Test"],"out":"--- SKIP: TestEpsilonSkips (0.00s)\n"}
{"a":"skip","test":true,"keys":["Action","Elapsed","Package","Test"],"out":null}
{"a":"output","test":false,"keys":["Action","Output","Package"],"out":"PASS\n"}
{"a":"output","test":false,"keys":["Action","Output","Package"],"out":"ok  \texample.com/censusfix/epsilon\t0.088s\n"}
{"a":"pass","test":false,"keys":["Action","Elapsed","Package"],"out":null}
```

같은 질의를 픽스처에 돌린 결과다.

```
{"a":"start","test":false,"keys":["Action","Package"],"out":null}
{"a":"run","test":true,"keys":["Action","Package","Test"],"out":null}
{"a":"output","test":true,"keys":["Action","Output","Package","Test"],"out":"--- PASS: TestEpsilonPasses (0.00s)\n"}
{"a":"pass","test":true,"keys":["Action","Elapsed","Package","Test"],"out":null}
{"a":"run","test":true,"keys":["Action","Package","Test"],"out":null}
{"a":"output","test":true,"keys":["Action","Output","Package","Test"],"out":"--- SKIP: TestEpsilonSkips (0.00s)\n"}
{"a":"skip","test":true,"keys":["Action","Elapsed","Package","Test"],"out":null}
{"a":"output","test":false,"keys":["Action","Output","Package"],"out":"ok  \texample.com/censusfix/epsilon\t0.120s\n"}
{"a":"pass","test":false,"keys":["Action","Elapsed","Package"],"out":null}
```

이벤트 종류별 필드 구성은 같다. 패키지 수준 pass도 실제와 마찬가지로 `{Action, Elapsed, Package}`다. 픽스처에는 실제 스트림에 있는 출력 이벤트 3개(`=== RUN` 두 줄, 패키지 수준 `PASS\n`)가 빠져 있다. 이 누락이 결과를 바꾸는지 보려고, 픽스처 1~35행 뒤에 실제 epsilon 이벤트 12줄을 붙인 변형을 만들어 census를 돌렸다.

```
$ head -n 35 fixture.jsonl > fixture-realeps.jsonl
$ jq -c 'select(.Package=="example.com/censusfix/epsilon")' real-stream.jsonl >> fixture-realeps.jsonl
$ bash scripts/ci-census/test-census.sh fixture-realeps.jsonl > realeps-out.txt
$ diff scripts/ci-census/testdata/expected.txt realeps-out.txt
(출력 없음, exit 0)
```

차이가 없으므로 누락은 무해하다(F2).

실제 스트림에 census를 돌린 결과도 확인했다.

```
$ bash scripts/ci-census/test-census.sh real-stream.jsonl
=== test census ===
NOTHING RAN   example.com/censusfix/beta
SKIPPED TEST  example.com/censusfix/epsilon  TestEpsilonSkips
=== totals: packages=2 passed=1 skipped=1 nothing-ran=1 failed=0 build-failed=0 ===
```

### 주장 5 — 다른 소비처가 없고, 주석 수정이 정확하다: **확인**

```
$ git grep -n -e 'ci-census' -e 'fixture.jsonl' -e 'censusfix' 2990924ef
(발췌) .github/workflows/ci.yml:210:        run: bash scripts/ci-census/census-check.sh
       .github/workflows/ci.yml:230/312/425, release-pr-multi-os.yml:211 → test-census.sh test-stream.json (실제 스트림, 픽스처 아님)
       scripts/ci-census/census-check.sh:39: fixture="$here/testdata/fixture.jsonl"
       나머지는 .moai/reports·.moai/specs·CHANGELOG 의 서술
```

- 픽스처와 기대 출력을 읽는 곳은 `census-check.sh` 하나다. CI에서는 `ci.yml` 210행, `test` 잡의 `Check test census against its fixture` 단계에서 실행된다.
- 옛 합계 줄이나 「five shapes」를 인용한 문서도 찾아봤다. 결과는 progress.md의 `expected.txt` 경로 언급 한 곳뿐이었다(`git grep -n -e 'packages=4' -e 'five shapes' -e 'testdata/expected' 2990924ef -- .moai/specs/SPEC-CI-TEST-OBSERVABILITY-001 .github scripts Makefile`). 합계 값을 인용한 문서가 없으므로 낡은 서술로 남는 곳은 없다.
- 기준 커밋의 머리 주석은 1~6번 여섯 항목을 나열하면서 「all five shapes」라고 적고 있었다(`git show 6ae4583e6:scripts/ci-census/census-check.sh`로 확인). 새 주석은 항목 7개에 「seven」이라고 적어 개수가 맞는다. 7번 항목의 설명(패키지 수준 pass를 세면 passed가 커지고, skip 테스트를 가진 패키지를 nothing-ran으로 세면 nothing-ran이 커지는데, 그런 패키지는 2개이고 테스트 파일이 없는 패키지는 1개)은 위 m3·m5 diff와 정확히 맞는다.
- `bash -n scripts/ci-census/census-check.sh` 결과 출력이 없다(구문 정상). `git diff --stat 6ae4583e6 2990924ef -- scripts/ci-census/test-census.sh .github/`도 출력이 없다(census 본체와 워크플로는 바뀌지 않음).

## 추가 변이 탐색 (참고용 — 수정 요구 아님)

| 변이 | 내용 | census-check | 비고 |
|---|---|---|---|
| x1 | packages를 `start` 이벤트로만 셈 | **exit 0 (생존)** | 픽스처의 모든 패키지에 start가 있어 등가. 정상 스트림에서도 사실상 등가 |
| x2 | packages를 Test가 있는 이벤트로만 셈 | exit 1 (`packages=2`) | 잡힘 |
| x3 | failed에서 `.Test != null` 제거 | exit 1 (`failed=4`) | 잡힘 |
| x4 | build-failed를 `FailedBuild` 필드로 셈 | **exit 0 (생존)** | 픽스처에서 등가(delta 1건) |
| x5 | passed를 (패키지, 테스트) 대신 패키지로 셈 | **exit 0 (생존)** | 패키지마다 통과 테스트가 1개뿐 |
| x6 | skipped를 패키지로 셈 | **exit 0 (생존)** | 패키지마다 skip 테스트가 1개뿐 |
| x7 | failed를 패키지로 셈 | **exit 0 (생존)** | 실패 테스트가 alpha에 1개뿐 |
| x8 | skipped에서 `.Test != null` 제거 | exit 1 (`skipped=3`) | 잡힘 |

모든 변이는 `run.sh head <이름> '<sed식>'`으로 실행했고, 각 실행에서 변이 줄 diff와 exit 코드를 확인했다.

## 발견 사항 (구조화된 결함 목록)

- **F1** [Low] [optional] `scripts/ci-census/testdata/fixture.jsonl` — 한 패키지에 같은 결과의 테스트가 두 개 이상 있는 경우가 없다. 그래서 passed·skipped·failed를 (패키지, 테스트) 대신 패키지 단위로 세는 변이(x5·x6·x7)가 살아남는다. 이번 카드 범위(m3·m5) 밖이며, t1223이 만든 결함도 아니다. 신뢰도 높음(재현 완료). — 필요하다면: 후속 카드에서 epsilon에 통과 테스트를 하나 더 넣어 passed 기대값을 3으로 올린다. 같은 방식으로 skip 테스트와 실패 테스트를 한 패키지에 두 개씩 두면 x6·x7도 잡힌다. 이때 FAILED 절이 바뀌므로 기대 출력도 손으로 고쳐야 한다. x1·x4는 정상 스트림에서 등가 변이일 가능성이 높아 조치할 필요가 없다.
- **F2** [Info] [optional] `fixture.jsonl` 37~44행 — 실제 `go test -json`(go1.26.8)이 내는 `=== RUN` 출력 두 줄과 패키지 수준 `PASS\n` 출력 한 줄이 빠져 있다. 실제 이벤트로 바꿔 돌려도 census 출력이 `expected.txt`와 바이트 단위로 같으므로 영향은 없다. 신뢰도 높음(측정 완료). — 필요한 수정: 없음. 현실성을 맞추고 싶다면 세 줄을 추가해도 기대 출력은 그대로다.
- **F3** [Info] [optional] `.moai/reports/t1223/verdict.md` 잔여 위험 항목 — 「실제 패키지 수준 pass 이벤트의 필드 구성이 다를 수 있다」는 우려는 측정 결과 해소됐다. 실제 모양도 `{Action, Elapsed, Package}`로 같다. 레인이 공백으로 적어 둔 m1·m2 diff 본문도 이번 감사에서 재현했다. — 필요한 수정: 없음(기록용).

차단(blocking) 발견은 없다.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

- **Claim**: t1223의 주장 1~5는 모두 참이다. 판정은 PASS(93.6)다.
- **Evidence**: 위 각 절의 명령과 출력 원문.
- **Baseline-attribution**: 작업트리 `.claude/worktrees/t1223`, 커밋 `2990924ef`(감사 대상)와 `6ae4583e6`(기준). 두 커밋의 `scripts/ci-census`를 `git archive`로 받은 tar 사본에서, 이번 세션에 측정했다. 로컬 jq·BSD sed·bash(macOS)와 go1.26.8 darwin/arm64를 썼다.
- **Gaps**: CI(ubuntu, GNU sed/jq/sort)에서의 실행은 관측하지 않았다. push와 CI 요청은 금지돼 있다. 새 skip 행의 정렬 순서가 로케일에 따라 달라지지 않는다는 점은 바이트 비교로 추론했으며, GNU 환경에서 실제로 돌려 보지는 않았다. 변이 공간은 위 12개(m1·m2·m3·m5, x1~x8)만 탐색했다. 교차 모델 감사(codex/glm)는 실행하지 않았다.
- **Residual-risk**: 픽스처는 여전히 합성이다. 픽스처 하나로 모든 변이를 가를 수는 없으며, F1의 x5~x7이 그 실례다. 앞으로 census에 새 집계가 추가되면 그 집계를 가르는 형태가 픽스처에 없을 수 있다.
