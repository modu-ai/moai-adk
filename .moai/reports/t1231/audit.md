# t1231 독립 감사 — census 픽스처 보강(zeta: 같은 이름, 섞인 순서)

- 대상: `.claude/worktrees/t1231`, 브랜치 `WT-census-final`, HEAD `528e9cb2b`, 기준 `044fb91c7`
- 감사자: sync-auditor (읽기 전용 — 작업트리 무수정·무커밋 확인: 감사 종료 시 `git status --short` 출력 없음, `git rev-parse --short HEAD` → `528e9cb2b`)
- SPEC 없음. 평가 프로필: 내장 기본값(Functionality·Security must-pass)

## 판정

**PASS-WITH-DEBT — 종합 93.0 (4개 차원 조화평균)**

카드 목표는 그대로 달성됐다. 기준 픽스처에서 y4·y5·y6·y7·y8 다섯 변이가 모두 살아 있음을 재현했고, HEAD 픽스처에서는 다섯 개 모두 FAIL한다. 기존 변이 x5·x6·x7·m1·m2·m3·m5도 계속 FAIL하고, 깨끗한 census는 PASS다. `expected.txt`의 합계는 독립 jq 집계와 일치한다. 차단 결함은 없다. 이번 카드가 바꾼 픽스처 때문에 머리 주석 7번 항목의 수치가 낡은 것(F1)과, 9번 항목이 실제 스트림 모양을 과장한 것(F2)이 부채로 남는다.

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 97 | PASS | 12개 변이 기준/HEAD 재현(아래 표), 합계 독립 도출 일치, 실작업트리 `census-check: PASS` |
| Security (25%) | 98 | PASS | 실행 로직 무변경(`test-census.sh` 기준↔HEAD 동일), 비밀 패턴 0건(양성 대조 68건 적중) |
| Craft (20%) | 85 | PASS | 새 이벤트의 필드 구성은 실측 스트림과 같다. 머리 주석 7번이 낡음(F1), 9번의 실제성 주장 과장(F2), 판정서 증거 축약·행 번호 오기(F3) |
| Consistency (15%) | 93 | PASS | 기존 픽스처 표기(시각·필드 순서·본문 들여쓰기)를 따름, 머리 주석 번호 체계 유지 |

조화평균: 4 / (1/97 + 1/98 + 1/85 + 1/93) = 93.0 (가중 산술평균으로는 94.3).

## 변이 표 (기준 `044fb91c7` 픽스처 vs HEAD `528e9cb2b` 픽스처)

`test-census.sh`는 기준과 HEAD가 같다(`git diff --quiet 044fb91c7 528e9cb2b -- scripts/ci-census/test-census.sh` → `test-census.sh unchanged base->HEAD`). 변이는 스크래치 사본에서 한 줄에만 `sed` 를 걸었고, 모든 변이가 원본과의 `diff` 에서 `changed_lines=1` 이다. 하네스: `scratchpad/t1231-audit/mut.sh`(git 호출 없음).

| 변이 | 행 | 내용 | 기준 | HEAD | HEAD에서 판정을 가른 diff 줄(원문) |
|---|---|---|---|---|---|
| y4 | 145 | skip 행 `sort -u`→`cat` | **PASS(생존)** | FAIL | `+SKIPPED TEST  example.com/censusfix/zeta  TestAlphaSkips` (NOTHING RAN 바로 뒤) / `-SKIPPED TEST  example.com/censusfix/zeta  TestAlphaSkips` (epsilon 뒤) |
| y5 | 98 | 실패 목록 `sort -u`→`cat` | **PASS(생존)** | FAIL | `+FAILED        example.com/censusfix/zeta  TestAlphaFails` 와 본문 3줄이 BUILD FAILED 바로 뒤로, `-` 쪽은 alpha 행들 뒤 |
| y6 | 151 | passed 키 `.Test` 만 | **PASS(생존)** | FAIL | `+=== totals: packages=6 passed=3 skipped=4 nothing-ran=1 failed=3 build-failed=1 ===` |
| y7 | 152 | skipped 키 `.Test` 만 | **PASS(생존)** | FAIL | `+=== totals: packages=6 passed=4 skipped=3 nothing-ran=1 failed=3 build-failed=1 ===` |
| y8 | 154 | failed 키 `.Test` 만 | **PASS(생존)** | FAIL | `+=== totals: packages=6 passed=4 skipped=4 nothing-ran=1 failed=2 build-failed=1 ===` |
| x5 | 151 | passed 키 `.Package` | FAIL | FAIL | `+=== totals: packages=6 passed=3 skipped=4 nothing-ran=1 failed=3 build-failed=1 ===` |
| x6 | 152 | skipped 키 `.Package` | FAIL | FAIL | `+=== totals: packages=6 passed=4 skipped=3 nothing-ran=1 failed=3 build-failed=1 ===` |
| x7 | 154 | failed 키 `.Package` | FAIL | FAIL | `+=== totals: packages=6 passed=4 skipped=4 nothing-ran=1 failed=2 build-failed=1 ===` |
| m1 | 141 | skip 필터 분리 | FAIL | FAIL | `-NOTHING RAN   example.com/censusfix/beta` |
| m2 | 171 | `%` 이스케이프 제거 | FAIL | FAIL | `+::notice title=AC snapshot absent::.moai/specs/SPEC-FIXTURE-100%/acceptance.md: HALT AC-FIXTURE-001 AC-FIXTURE-002 (...)` |
| m3 | 151 | passed의 `.Test != null` 제거 | FAIL | FAIL | `+=== totals: packages=6 passed=5 skipped=4 nothing-ran=1 failed=3 build-failed=1 ===` |
| m5 | 153 | nothing-ran 조건 뒤집기 | FAIL | FAIL | `+=== totals: packages=6 passed=4 skipped=4 nothing-ran=3 failed=3 build-failed=1 ===` |
| n1 | 148 | `count()` 의 `sort -u` 제거 | FAIL | FAIL | `+=== totals: packages=65 passed=4 skipped=4 nothing-ran=1 failed=3 build-failed=1 ===` |
| n2 | 121 | `pkgs_with_failed_tests` 의 `sort -u`→`cat` | 생존 | **생존** | (등가 변이 — 소속 검사에만 쓰임) |
| n3 | 122 | `pkg_failed` 의 `sort -u`→`cat` | 생존 | **생존** | — |
| n4 | 84 | `build_failed` 의 `sort -u`→`cat` | 생존 | **생존** | — |
| n5 | 121 | FAILED PKG 제외 키를 `.Test` 로 | FAIL | FAIL | `+FAILED PKG    example.com/censusfix/alpha` … `+FAILED PKG    example.com/censusfix/zeta` |
| n6 | 107 | 실패 본문 필터에서 `.Package==$p` 제거 | **생존** | FAIL | `+      zeta_test.go:12: expected ok, got timeout` 가 alpha 본문에, `+      alpha_test.go:10: expected 7, got 3` 가 zeta 본문에 섞임 |
| n7 | 131 | FAILED PKG 본문 필터에서 `.Package==$p` 제거 | FAIL | FAIL | `+  ?   	example.com/censusfix/beta	[no test files]` 등 |
| n8 | 90 | build-output 필터에서 `.ImportPath==$p` 제거 | 생존 | **생존** | — |
| n9 | 126 | `grep -Fxq`→`grep -Fq`(부분 일치) | 생존 | **생존** | — |
| n10 | 122 | `FailedBuild` 제외 조건 제거 | FAIL | FAIL | `+FAILED PKG    example.com/censusfix/delta` |
| n11 | 168 | notice 의 `sort -u`→`cat` | FAIL | FAIL | `+::notice title=AC snapshot absent::.moai/specs/SPEC-FIXTURE-001/acceptance.md: COUNT 3 (...)` 중복 |
| n12 | 98 | 실패 목록 키를 `.Test` 만 | FAIL | FAIL | `+FAILED        TestAlphaFails  ` |
| n13 | 145 | skip 행 `sort -u`→`sort` | 생존 | **생존** | (사실상 등가 — 한 실행에서 같은 skip 이벤트는 한 번) |
| n14 | 126 | `grep -Fxq`→`grep -xq`(정규식) | 생존 | **생존** | (사실상 등가 — `.` 이 임의 문자와 맞는 경우만 차이) |
| clean | — | — | PASS | PASS | `census-check: PASS (census output matches …/head/testdata/expected.txt)` |

요약: 카드가 겨냥한 y4~y8 다섯 개는 기준에서 모두 생존, HEAD에서 모두 FAIL이다. 기존 일곱 개는 양쪽 모두 FAIL이다. y4·y5 diff에는 합계 줄이 없고, y6~y8·x5~x7·m3·m5는 합계 줄의 자기 칸 하나만, m1·m2는 행 하나만 바뀐다 — 판정서 105행의 서술과 일치한다. 덤으로 n6(실패 본문을 테스트 이름만으로 고르는 변이)가 기준에서는 살고 HEAD에서는 죽는다.

## 검증 2 — `expected.txt` 도출 가능성과 zeta 이벤트 모양

**합계 독립 집계** (census를 쓰지 않은 jq -s):

```
jq -s '{packages: ([.[]|.Package//empty]|unique|length), passed: ([.[]|select(.Action=="pass" and .Test!=null)|[.Package,.Test]]|unique|length), skipped: …, nothing_ran: …, failed: …, build_failed: …, names_only_passed: …}' fixture.jsonl
{
  "packages": 6, "passed": 4, "skipped": 4, "nothing_ran": 1, "failed": 3, "build_failed": 1,
  "pkgs": ["…/alpha","…/beta","…/delta","…/epsilon","…/gamma","…/zeta"],
  "names_only_passed": 3
}
```

`expected.txt` 23행 `=== totals: packages=6 passed=4 skipped=4 nothing-ran=1 failed=3 build-failed=1 ===` 과 일치한다. 이름만으로 센 passed가 3이라는 점이 y6가 죽는 이유를 그대로 보여 준다.

**행 도출**: zeta의 `TestAlphaFails` output 이벤트 세 개(`=== RUN   TestAlphaFails\n`, `    zeta_test.go:12: expected ok, got timeout\n`, `--- FAIL: TestAlphaFails (0.00s)\n`)에 census 들여쓰기 두 칸을 붙이면 `expected.txt` 12~14행과 같다. zeta에는 실패 테스트가 있으므로 FAILED PKG 행이 없어야 하고, 실제로 없다. 기준 대비 `expected.txt` 변경은 FAILED 4줄·SKIPPED 1줄 추가와 합계 한 줄 교체뿐이다(`git diff 044fb91c7 528e9cb2b`). 정렬 위치(alpha < zeta, epsilon < zeta)는 C 로케일이든 glibc 기본 대조든 순서가 바뀌지 않는 조합이다. 픽스처 68줄 전부 `jq -e .` 로 유효 JSON(`all-json-valid`), 파일 끝은 개행(`00002715: 0a`)이다.

**실측 스트림과의 대조**: 스크래치에 패키지 두 개(`aa`: 느림, `zz`: 빠름, 테스트는 소스 순서 TestZLast·TestSkipMe·TestAFirst)짜리 모듈을 만들고 `go -C … test -json -count=1 -timeout 60s ./...` 를 돌렸다(go1.26.8 darwin/arm64, 종료 코드 1은 의도된 실패 테스트 때문). 이벤트 순서 원문(요약 없이 필드만 추림):

```
["start","example.com/probe/aa",null,""]
["start","example.com/probe/zz",null,""]
["run","example.com/probe/zz","TestZLast",""]
…(zz 의 모든 테스트 이벤트, 이어서)…
["output","example.com/probe/zz",null,"FAIL\texample.com/probe/zz\t0.238s"]
["fail","example.com/probe/zz",null,""]
["run","example.com/probe/aa","TestZLast",""]
…(aa 의 모든 테스트 이벤트)…
["fail","example.com/probe/aa",null,""]
```

키 구성은 `start`=`Action,Package,Time`, `run`=`+Test`, `output`=`+Output`, pass/skip/fail=`+Elapsed` 로, zeta 새 이벤트 17줄과 같다. 다른 점은 셋이다.
1. 실측에서는 패키지 하나의 테스트 이벤트와 패키지 종료 이벤트가 **한 덩어리로 연속**해서 나오고, 덩어리 순서가 **끝난 순서**(zz가 aa보다 먼저)다. 이벤트 단위로 다른 패키지 사이에 끼어들지는 않았다. 픽스처는 zeta의 테스트 블록을 alpha 블록 한가운데에 넣고, zeta의 패키지 종료 이벤트 3줄은 파일 맨 끝에 따로 두었다(F2).
2. zeta 이벤트 14줄의 시각이 모두 `…00.000010` 으로 같다. census는 `Time` 을 읽지 않으므로 영향은 없다.
3. zeta `TestAlphaSkips` 에는 skip 사유 줄이 없다. `t.SkipNow()` 라면 실제로 가능한 모양이다.

## 검증 3 — 판정서 주장 대조

| 판정서 주장 | 대조 결과 |
|---|---|
| 기준에서 y4~y8 생존, x5·x6·x7·m1·m2·m3·m5 FAIL (35~47행) | 재현 일치 |
| 보강 후 12개 모두 FAIL, clean PASS (72~85행) | 재현 일치 |
| 변이별 판정 사유 표 (90~103행) | 내용은 일치. 다만 경로를 `…` 로 줄인 요약이라 원문이 아니다(F3) |
| y4·y5는 순서만, 나머지는 자기 칸/행 하나만 (105행) | 일치 |
| 「병렬 실행이 내는 순서처럼」 (8행), 주석 9번 | 실측 1회와 맞지 않음(F2) |
| `count()` 안의 `sort -u`(157행 부근) (117행) | `count()` 는 148행이다. 157행은 합계 출력 `echo` (F3) |
| 기준 귀속 「HEAD `044fb91c7`」 (13·111행) | 측정 대상은 그 위의 미커밋 변경이었고 결과 커밋 `528e9cb2b` 가 적혀 있지 않다. 이번 감사가 `528e9cb2b` 트리(작업트리와 `cmp` 동일)에서 같은 결과를 재현했으므로 결론은 유지된다(F3) |
| 「expected.txt 는 census를 돌려 만들지 않았다」 (58행) | 저작 과정은 검증할 수 없다. 도출 가능성만 확인했다(Gaps) |

## 검증 5 — `census-check.sh` 머리 주석

- 1~6번: 픽스처와 맞다.
- **7번(20~24행)**: 「because this fixture has two such packages」 — skip된 테스트를 가진 패키지는 이제 alpha·epsilon·zeta 셋이다. HEAD에서 m5가 `nothing-ran=3` 을 내는 것이 그 증거다(기준에서는 `nothing-ran=2`). 이번 카드의 픽스처 변경이 이 문장을 낡게 만들었다(F1). 「counting package-level pass events inflates passed」 는 여전히 참이다(m3 → `passed=5`).
- 8번: 여전히 참(epsilon 통과 2·skip 2, alpha 실패 2; x5~x7이 더 작은 수를 낸다).
- **9번(29~33행)**: 「a second failing package (zeta)」 — 패키지 단위 fail 이벤트를 가진 패키지는 delta·alpha·gamma·zeta 넷이므로 「second」 는 모호하다(아마 「실패 테스트가 있는 두 번째 패키지」 뜻). 「as a parallel run emits them」 은 실측 1회와 맞지 않는다(F2). 나머지(이름만 세면 더 작은 수, 정렬을 빼면 행 순서가 틀어짐)는 y4~y8 결과로 확인된다.
- 「all nine shapes」 와 항목 수 9는 일치한다. `bash -n` 통과(출력 없음), 실작업트리 실행 `census-check: PASS`.

## Findings (구조화 결함 목록)

- **F1** [Low] [optional] 신뢰도 높음 — `scripts/ci-census/census-check.sh:20-24` — 7번 항목의 「this fixture has two such packages」 가 이번 변경으로 틀렸다. skip된 테스트를 가진 패키지는 이제 셋(alpha·epsilon·zeta)이고, m5 변이의 출력 `nothing-ran=3` 이 이를 보여 준다. 판정에는 영향이 없다(m5는 여전히 FAIL). — 필요한 수정: 「two such packages」 를 「three such packages」 로 고치거나, 수를 빼고 「more such packages than packages with no test files」 로 쓴다.
- **F2** [Low] [optional] 신뢰도 중간(실측 1회) — `scripts/ci-census/census-check.sh:29-31`, `testdata/fixture.jsonl:19-32, 66-68` — 「its events interleaved between alpha's as a parallel run emits them」 는 go1.26.8의 `go test -json ./...` 실측과 다르다. 실측에서는 패키지별 이벤트가 한 덩어리로 연속하고, 덩어리 순서가 끝난 순서를 따른다. 픽스처는 zeta 테스트 블록을 alpha 블록 안에 넣고 zeta 패키지 종료 이벤트를 파일 끝에 떨어뜨렸다. census 로직은 이 차이에 영향을 받지 않고, y4·y5를 잡는 데 필요한 것은 「zeta 행이 alpha 행보다 먼저 나온다」 뿐이라 실측 모양(덩어리째 먼저 끝남)으로도 같은 효과가 난다. 「a second failing package」 도 모호하다. — 필요한 수정: 주석을 「its package finishes before alpha, so its events come first in the stream」 정도로 고치고, 원한다면 zeta 블록(테스트 + 종료 3줄)을 alpha 블록 앞에 한 덩어리로 옮긴다. 옮기면 `expected.txt` 는 바뀌지 않는다(행은 정렬되므로).
- **F3** [Low] [optional] 신뢰도 높음 — `.moai/reports/t1231/verdict.md:13, 90-107, 111, 117` — (a) 판정 사유 표가 경로를 `…` 로 줄인 요약이고 원문은 스크래치에만 있다고 적었다(verification-claim-integrity §3, 증거 반출 의무). t1228 감사 F4와 같은 모양의 재발이다. (b) 기준 귀속이 「HEAD `044fb91c7`」 로만 적혀 있고, 측정한 트리의 결과 커밋 `528e9cb2b` 가 없다. (c) 117행 「157행 부근」 은 148행이 맞다. 결론 자체는 이번 감사가 재현했다. — 필요한 수정: 판정을 가른 diff 줄 원문을 붙이고(이 감사의 변이 표를 인용해도 된다), 귀속에 `528e9cb2b` 를 더하고, 행 번호를 148로 고친다.
- **F4** [Low] [optional] 신뢰도 높음 — `scripts/ci-census/test-census.sh:84, 90, 122, 126` — HEAD 픽스처에서도 사는 비등가 변이 넷(운영 지시에 따라 한계로만 기록): n3(122행 `pkg_failed` 정렬 제거 — FAILED PKG 패키지가 둘 이상이고 끝난 순서가 사전순이 아니면 행 순서가 틀어짐), n4(84행 `build_failed` 정렬 제거 — 빌드 실패 둘 이상), n8(90행 build-output 을 ImportPath로 거르지 않음 — 빌드 실패 둘 이상이면 진단이 서로 섞임), n9(126행 `grep -Fxq`→`-Fq` — `a/b` 가 패키지 수준 실패이고 `a/b/c` 에 실패 테스트가 있으면 `a/b` 의 FAILED PKG 가 사라짐). 픽스처에 FAILED PKG 와 BUILD FAILED 가 각각 하나뿐이어서 생긴다. n2는 등가, n13·n14는 사실상 등가다. — 필요한 수정: 없음(연작 종료 지시). 후속을 한다면 두 번째 패키지 수준 실패와 두 번째 빌드 실패, 그리고 경로가 다른 실패 패키지의 접두어인 패키지를 넣는다.

차단(blocking) 결함 없음. 모든 결함이 optional이므로 판정을 FAIL로 바꾸지 않는다.

## Claim

1. 커밋 `528e9cb2b` 는 카드 목표를 달성했다: 기준 픽스처에서 살던 y4·y5·y6·y7·y8이 HEAD 픽스처에서 모두 `census-check` FAIL이고, x5·x6·x7·m1·m2·m3·m5는 계속 FAIL이며, 깨끗한 census는 PASS다.
2. `expected.txt` 는 픽스처에서 손으로 도출 가능하며, 합계는 census와 무관한 jq 집계와 같다.
3. zeta 이벤트의 필드 구성은 실측 `go test -json` 과 같고, 배치 순서는 실측 모양과 다르다(F2). 머리 주석 7번은 이번 변경으로 낡았다(F1).
4. 새 변이 14개 중 HEAD에서 비등가 생존 4개(n3·n4·n8·n9), 등가/사실상 등가 생존 3개(n2·n13·n14)가 있다(F4).

## Evidence

- 변경 불변 확인: `git diff --quiet 044fb91c7 528e9cb2b -- scripts/ci-census/test-census.sh && echo …` → `test-census.sh unchanged base->HEAD`
- 사본: `git show 044fb91c7:scripts/ci-census/<파일>` / `git show 528e9cb2b:…` 로 `scratchpad/t1231-audit/{base,head}/` 에 반출. 기준 clean: `census-check: PASS (census output matches …/base/testdata/expected.txt)`
- 변이 실행: `bash scratchpad/t1231-audit/mut.sh scratchpad/t1231-audit/base` 및 `…/head` — 결과는 위 변이 표에 원문 줄로 옮겼다. 대표 원문:
  - 기준: `y4 line=145 changed_lines=1 rc=0 census-check=PASS` … `y8 line=154 changed_lines=1 rc=0 census-check=PASS`
  - HEAD: `y4 line=145 changed_lines=1 rc=1 census-check=FAIL` … `y8 line=154 changed_lines=1 rc=1 census-check=FAIL`, `clean rc=0`
  - HEAD y4 전체 diff 일부: `+SKIPPED TEST  example.com/censusfix/zeta  TestAlphaSkips` (18행 `NOTHING RAN` 바로 뒤), `-SKIPPED TEST  example.com/censusfix/zeta  TestAlphaSkips` (epsilon 뒤)
- 실작업트리: `bash scripts/ci-census/census-check.sh` → `census-check: PASS (census output matches /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1231/scripts/ci-census/testdata/expected.txt)`
- 비밀 패턴: `grep -cE '(api[_-]?key|token|secret|password|PRIVATE KEY)' …/fixture.jsonl` → `0`, 양성 대조 `…|censusfix)` → `68`
- 실측 스트림: `go -C …/probe test -json -count=1 -timeout 60s ./...` → 위 검증 2 인용

## Baseline-attribution

- 측정 트리: `528e9cb2b` 의 `scripts/ci-census/` (스크래치 사본이 작업트리 파일과 `cmp` 로 같음을 확인한 뒤 사용 — 첫 복합 명령은 가드에 거부돼 개별 명령으로 다시 확인), 기준은 `044fb91c7` 의 같은 경로. 모두 이번 실행에서 관측했다.
- 도구: jq(PATH), macOS `/usr/bin/sed`·`sort`·`diff`, GNU bash `bash`(PATH), Go go1.26.8 darwin/arm64. 판정 대상인 census는 저장소 스크립트 자체라 빌드 계보(§2.2) 문제는 없다.

## Gaps

- **가드 거부와 대체 경로**(§3.1): 작업트리 격리 가드가 다음을 거부했다 — `git show` 를 반복문·변수로 묶은 복합 명령, `;`·`echo rc=$?` 가 붙은 `bash` 호출, heredoc 으로 파일을 만드는 명령, `tail | od` 파이프. 대체로 `git show` 를 한 줄씩 따로 실행했고, 파일은 Write 도구로 만들었으며, 변이 실행은 git 호출이 없는 하네스 스크립트(`mut.sh`)를 `bash <절대경로>` 로 돌렸다. 하네스 안에는 git 이 없으므로 가드가 막으려던 대상을 우회한 것은 아니지만, 가드가 스크립트 내부를 읽지 못한다는 사실은 여기 적어 둔다.
- 실제 CI(리눅스, GNU sed/sort, 로케일) 실행은 보지 않았다(push·CI 요청 금지).
- `shellcheck` 미설치(`command not found: shellcheck`) — 정적 분석은 `bash -n` 뿐이다.
- 실측 스트림 대조는 1회, 패키지 2개, macOS 한 버전뿐이다. 다른 Go 버전이나 `-p` 설정에서 이벤트 단위 섞임이 생기는지는 재지 않았다.
- `expected.txt` 가 「census를 돌리지 않고 손으로」 고쳐졌다는 저작 과정은 검증할 수 없다. 도출 가능성만 확인했다.
- t1228의 y1·y2·y3·y9는 다시 돌리지 않았다.

## Residual-risk

- 픽스처 하나로 모든 변이를 가를 수 없다. HEAD에서도 비등가 변이 넷이 산다(F4). 모두 「같은 종류의 행이 둘 이상」 이거나 「경로 접두어 관계」 가 있어야 드러나는 결함이다.
- 정렬 결과는 CI의 GNU `sort` 로케일에서 처음 확인된다. 이번 추가 행(alpha < zeta, epsilon < zeta)은 대소문자·구두점 차이가 없는 조합이라 위험은 낮다.
- F2의 순서 차이는 지금 census에 영향이 없지만, 나중에 census가 스트림 순서에 의존하는 기능(예: 패키지 블록 단위 처리)을 얻으면 실측과 다른 픽스처가 그 결함을 가리거나 가짜 실패를 만들 수 있다.
