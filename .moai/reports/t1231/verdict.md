# t1231 — census 픽스처 보강: 같은 이름의 테스트를 다른 패키지에, 이벤트는 정렬되지 않은 순서로

## Claim

1. t1228 감사가 남긴 생존 변이 다섯 개를 이번 카드 기준 트리에서 재현했다. 모두 `census-check.sh`를 통과했다.
   - F2 변이 y6·y7·y8: passed·skipped·failed 합계를 테스트 이름(`.Test`)만으로 센다.
   - F3 변이 y4·y5: skip 행과 실패 목록에서 `sort -u`를 뺀다.
2. 픽스처에 실패하는 패키지 zeta를 새로 넣었다. zeta에는 통과·skip·실패 테스트를 하나씩 두고, 세 테스트 모두 alpha에 이미 있는 이름을 그대로 쓴다. zeta 이벤트는 alpha 이벤트 사이에 끼워 넣었다. 병렬 실행이 내는 순서처럼 정렬되지 않은 순서다. 기대 출력은 손으로 고쳤다.
3. 보강 후 새 변이 다섯 개와 기존 변이 일곱 개(x5·x6·x7·m1·m2·m3·m5), 모두 열두 개가 FAIL로 잡힌다. 깨끗한 census는 PASS다.

## Evidence

기준은 작업트리 `.claude/worktrees/t1231`, 브랜치 `WT-census-final`, HEAD `044fb91c7`(로컬 develop)이다. 변이는 모두 세션 스크래치패드의 사본에만 넣었다. 스크립트가 사본마다 원본 `test-census.sh`와 `diff`해서 바뀐 줄 수를 센다(`changed_lines=1`). 기존 변이 일곱 개는 t1228에서 만든 변이 파일을 재사용했다. `test-census.sh`는 그 뒤로 바뀌지 않았고, 원본과의 차이가 1줄인 것도 같은 스크립트로 다시 확인했다.

**변이 정의** (`scripts/ci-census/test-census.sh`)

| 이름 | 위치 | 변이 |
|---|---|---|
| y4 | 145행 skip 행 | `| sort -u` → `| cat` |
| y5 | 98행 실패 목록 | `| sort -u)"` → `| cat)"` |
| y6 | 151행 passed | 집계 키 `"\(.Package)\t\(.Test)"` → `.Test` |
| y7 | 152행 skipped | 같음 |
| y8 | 154행 failed | 같음 |
| x5·x6·x7 | 151·152·154행 | 집계 키 → `.Package` (t1223 감사 F1, t1228) |
| m1 | 141행 | skip 판정 필터 분리 (t1218) |
| m2 | 171행 | notice의 `%` 이스케이프 제거 (t1218) |
| m3 | 151행 | passed에서 `.Test != null` 조건 제거 (t1223) |
| m5 | 153행 | nothing-ran 조건 뒤집기 (t1223) |

**재현 — 보강 전 픽스처(`044fb91c7`)**

명령: `bash scratchpad/t1231-mut.sh scripts/ci-census scratchpad/t1231/before`

```
y4 changed_lines=1 census-check=PASS
y5 changed_lines=1 census-check=PASS
y6 changed_lines=1 census-check=PASS
y7 changed_lines=1 census-check=PASS
y8 changed_lines=1 census-check=PASS
x5 changed_lines=1 census-check=FAIL
x6 changed_lines=1 census-check=FAIL
x7 changed_lines=1 census-check=FAIL
m1 changed_lines=1 census-check=FAIL
m2 changed_lines=1 census-check=FAIL
m3 changed_lines=1 census-check=FAIL
m5 changed_lines=1 census-check=FAIL
clean: census-check: PASS (census output matches /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1231/scripts/ci-census/testdata/expected.txt)
```

**보강** — `fixture.jsonl`에 이벤트 17줄을 추가했다.
- alpha `TestAlphaPasses` 통과 이벤트 바로 뒤, alpha `TestAlphaSkips`보다 앞에 zeta 블록 14줄을 넣었다: `start`, 그리고 `TestAlphaSkips`(run, output 2, skip), `TestAlphaFails`(run, output 3, fail), `TestAlphaPasses`(run, output 2, pass).
- 파일 끝(epsilon 패키지 pass 뒤)에 zeta 패키지 수준 이벤트 3줄을 넣었다: output `FAIL`, output `FAIL\t…zeta\t0.210s`, 패키지 fail.

이 배치가 두 결함을 가른다.
- **F2**: 이름만으로 세면 zeta의 세 테스트가 alpha의 같은 이름과 겹쳐 하나씩 사라진다.
- **F3**: zeta는 사전순으로 alpha·epsilon보다 뒤인데 스트림에서는 앞에 나온다. 그래서 정렬을 빼면 zeta 행이 제자리를 벗어난다.

`expected.txt`는 census를 돌려 만들지 않았다. 픽스처 이벤트에서 따져 손으로 고쳤다.
- `FAILED        example.com/censusfix/zeta  TestAlphaFails` 행과 본문 세 줄을 alpha의 FAILED 행들 뒤에 추가했다.
- `SKIPPED TEST  example.com/censusfix/zeta  TestAlphaSkips` 행을 epsilon skip 행들 뒤에 추가했다.
- 합계를 `packages=6 passed=4 skipped=4 nothing-ran=1 failed=3 build-failed=1`로 고쳤다.
- zeta에는 실패 테스트가 있으므로 FAILED PKG 행은 생기지 않는다.

`census-check.sh` 머리 주석에 9번 형태를 추가하고 「eight」를 「nine」으로 고쳤다.

손 편집에서 한 번 실수가 있었다. 첫 실행에서 깨끗한 census가 FAIL했다. `FAILED PKG` 행의 공백 네 칸을 세 칸으로 줄이는 편집 실수였고, 원인을 확인해 고쳤다. 아래는 고친 뒤 모든 변이를 **다시** 돌린 결과다. 실수가 있던 상태의 변이 결과는 판정에 쓰지 않았다.

**보강 후**

명령: `bash scratchpad/t1231-mut.sh scripts/ci-census scratchpad/t1231/after`

```
y4 changed_lines=1 census-check=FAIL
y5 changed_lines=1 census-check=FAIL
y6 changed_lines=1 census-check=FAIL
y7 changed_lines=1 census-check=FAIL
y8 changed_lines=1 census-check=FAIL
x5 changed_lines=1 census-check=FAIL
x6 changed_lines=1 census-check=FAIL
x7 changed_lines=1 census-check=FAIL
m1 changed_lines=1 census-check=FAIL
m2 changed_lines=1 census-check=FAIL
m3 changed_lines=1 census-check=FAIL
m5 changed_lines=1 census-check=FAIL
clean: census-check: PASS (census output matches /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1231/scripts/ci-census/testdata/expected.txt)
```

**FAIL 사유** — 명령 `bash scratchpad/t1231-why.sh`. 각 변이의 census-check diff 가운데 판정을 가른 줄이다.

출력 원문(`grep -E '^[-+][^-+]'`로 거른 줄, 변이당 앞 6줄):

```
== y4
+SKIPPED TEST  example.com/censusfix/zeta  TestAlphaSkips
-SKIPPED TEST  example.com/censusfix/zeta  TestAlphaSkips
== y5
+FAILED        example.com/censusfix/zeta  TestAlphaFails
+  === RUN   TestAlphaFails
+      zeta_test.go:12: expected ok, got timeout
+  --- FAIL: TestAlphaFails (0.00s)
-FAILED        example.com/censusfix/zeta  TestAlphaFails
-  === RUN   TestAlphaFails
== y6
-=== totals: packages=6 passed=4 skipped=4 nothing-ran=1 failed=3 build-failed=1 ===
+=== totals: packages=6 passed=3 skipped=4 nothing-ran=1 failed=3 build-failed=1 ===
== y7
-=== totals: packages=6 passed=4 skipped=4 nothing-ran=1 failed=3 build-failed=1 ===
+=== totals: packages=6 passed=4 skipped=3 nothing-ran=1 failed=3 build-failed=1 ===
== y8
-=== totals: packages=6 passed=4 skipped=4 nothing-ran=1 failed=3 build-failed=1 ===
+=== totals: packages=6 passed=4 skipped=4 nothing-ran=1 failed=2 build-failed=1 ===
== x5
-=== totals: packages=6 passed=4 skipped=4 nothing-ran=1 failed=3 build-failed=1 ===
+=== totals: packages=6 passed=3 skipped=4 nothing-ran=1 failed=3 build-failed=1 ===
== x6
-=== totals: packages=6 passed=4 skipped=4 nothing-ran=1 failed=3 build-failed=1 ===
+=== totals: packages=6 passed=4 skipped=3 nothing-ran=1 failed=3 build-failed=1 ===
== x7
-=== totals: packages=6 passed=4 skipped=4 nothing-ran=1 failed=3 build-failed=1 ===
+=== totals: packages=6 passed=4 skipped=4 nothing-ran=1 failed=2 build-failed=1 ===
== m1
-NOTHING RAN   example.com/censusfix/beta
== m2
-::notice title=AC snapshot absent::.moai/specs/SPEC-FIXTURE-100%25/acceptance.md: HALT AC-FIXTURE-001 AC-FIXTURE-002 (matched by the AC corpus glob but absent from the AC snapshot - reported, not failed)
+::notice title=AC snapshot absent::.moai/specs/SPEC-FIXTURE-100%/acceptance.md: HALT AC-FIXTURE-001 AC-FIXTURE-002 (matched by the AC corpus glob but absent from the AC snapshot - reported, not failed)
== m3
-=== totals: packages=6 passed=4 skipped=4 nothing-ran=1 failed=3 build-failed=1 ===
+=== totals: packages=6 passed=5 skipped=4 nothing-ran=1 failed=3 build-failed=1 ===
== m5
-=== totals: packages=6 passed=4 skipped=4 nothing-ran=1 failed=3 build-failed=1 ===
+=== totals: packages=6 passed=4 skipped=4 nothing-ran=3 failed=3 build-failed=1 ===
```

y4·y5는 zeta 행의 위치만 바뀌고 합계는 그대로다. y6~y8·x5~x7·m3·m5는 자기 합계 칸 하나만 바뀐다. m1·m2는 행 하나만 바뀐다.

## Baseline-attribution

기준 커밋은 `044fb91c7`(로컬 develop)이고 작업트리는 `.claude/worktrees/t1231`이다. 위 결과는 모두 이 트리에서 이번 실행으로 관측했다.
- 「보강 전」: 픽스처를 고치기 전 트리, 곧 `044fb91c7` 그대로다.
- 「보강 후」: 편집 실수를 고친 트리다. 여기에 판정한 커밋이 `528e9cb2b`이다.
- 감사 반영 후: `census-check.sh` 주석(감사 F1·F2)만 고친 트리에서 같은 스크립트를 다시 돌렸다. 결과는 위 「보강 후」 블록과 같았다. 변이 12개 모두 FAIL, 깨끗한 census PASS, `bash -n` 통과. 이 트리는 감사 반영 커밋에 담긴다.

## Gaps

- 실제 CI 실행은 보지 않았다(push·CI 요청 금지).
- 새 zeta 이벤트는 기존 이벤트와 같은 모양을 따랐다. 실제 `go test -json` 스트림과 다시 대조하지는 않았다.
- 저작자 쪽에서는 `count()` 안의 `sort -u`(148행)를 빼는 변이와 `pkg_failed`·`pkgs_with_failed_tests`의 `sort -u`를 빼는 변이를 돌리지 않았다. 이번 범위는 F2·F3이 지목한 y4·y5까지다. 이 변이들은 독립 감사가 n1~n3으로 돌렸다(아래 「한계」).
- t1228 감사의 y3(`-le`→`-lt`, 실패 51건 이상 필요)와 y9(packages를 `start` 이벤트로 세기, 등가 변이)는 다시 시험하지 않았다.

## Residual-risk

- 픽스처 하나로 모든 변이를 구별할 수는 없다. 이번 보강은 감사가 지목한 다섯 변이를 닫았을 뿐이다.
- 운영 지시에 따라 이 카드로 census 변이 대조 연작(t1215·t1218·t1223·t1228·t1231)을 닫는다. 독립 감사에서 새로 살아남는 변이가 나오더라도 후속 카드로 잇지 않고, 아래 「한계」에 기록만 한다.

## 독립 감사 반영

**판정: sync-auditor PASS-WITH-DEBT, 종합 93.0.** 차원별 점수는 Functionality 97, Security 98, Craft 85, Consistency 93이다. 전문은 `audit.md`에 있다.

감사가 직접 재현한 결과는 다음과 같다.
- y4~y8은 `044fb91c7`에서 생존하고 `528e9cb2b`에서 FAIL이다.
- 기존 변이 일곱 개는 양쪽 모두 FAIL이다.
- 합계는 census를 거치지 않고 jq로 따로 세어 `expected.txt`와 같음을 확인했다.

찾은 결함 네 건은 모두 Low이고 optional이다.

| 번호 | 내용 | 처리 |
|---|---|---|
| F1 | 머리 주석 7번의 "this fixture has two such packages"가 zeta 추가로 틀린 문장이 됐다. skip 테스트가 있는 패키지는 이제 셋이다 | **수리.** 문장을 alpha·epsilon·zeta 셋과 beta 하나의 비교로 고쳤다 |
| F2 | 9번의 "as a parallel run emits them"이 실측과 다르다. 실측에서는 패키지마다 한 덩어리로, 끝난 순서대로 나온다. "a second failing package"도 모호하다 | **수리.** 문구를 「alpha보다 사전순으로 뒤인 zeta의 이벤트가 alpha의 skip·fail보다 먼저 온다」로 고치고, 실제 병렬 실행의 모양은 괄호로 따로 밝혔다. 픽스처 이벤트 배치는 바꾸지 않았다. 판정에 필요한 성질(zeta 행이 alpha 행보다 먼저 옴)은 실측 모양에서도 성립하기 때문이다(감사 확인) |
| F3 | 판정 사유를 `…`로 줄인 요약 표로만 제시했고, 기준 귀속에 판정 커밋이 빠졌고, `count()` 줄 번호가 틀렸다 | **수리.** 사유 표를 출력 원문으로 바꾸고 `528e9cb2b`를 귀속에 넣었다. 줄 번호는 157 → 148로 고쳤다 |
| F4 | 동작이 달라지는 생존 변이 4개(n3·n4·n8·n9) | **한계로만 기록한다**(아래) |

## 한계 — 연작 종결에 따라 잇지 않는다

운영 지시에 따라 census 변이 대조 연작(t1215·t1218·t1223·t1228·t1231)을 이 카드로 닫는다. 아래 생존 변이는 후속 카드로 만들지 않고 여기에 기록만 한다.

동작이 실제로 달라지지만 현재 픽스처로는 드러나지 않는 변이다(`test-census.sh` HEAD 줄 번호, 감사 `audit.md` 표 기준).

| 변이 | 위치 | 내용 | 드러나는 조건 |
|---|---|---|---|
| n3 | 122행 | `pkg_failed`의 `sort -u`→`cat` | FAILED PKG 대상 패키지가 둘 이상이고 순서가 섞여 있을 때 |
| n4 | 84행 | `build_failed`의 `sort -u`→`cat` | BUILD FAILED 패키지가 둘 이상이고 순서가 섞여 있을 때 |
| n8 | 90행 | build-output 필터에서 `.ImportPath==$p` 제거 | build-fail 패키지가 둘 이상일 때(본문이 서로 섞인다) |
| n9 | 126행 | `grep -Fxq`→`grep -Fq`(부분 일치) | 한 패키지 경로가 실패 테스트가 있는 다른 패키지 경로의 접두어일 때 |

등가이거나 사실상 등가라서 구별할 필요가 없는 변이는 n2(소속 검사에만 쓰는 목록의 정렬), n13(`sort -u`→`sort`, 한 실행에서 같은 skip 이벤트는 한 번뿐), n14(`-Fxq`→`-xq`, `.`이 임의 문자에 맞는 경우만 차이)다.

n6(실패 본문 필터에서 `.Package==$p` 제거)은 기준 픽스처에서는 생존했고, 이번 보강으로 FAIL이 됐다. 의도하지 않았지만 이번 보강이 함께 잡은 변이다.

픽스처 하나로 모든 변이를 구별할 수는 없다. 위 네 개를 잡으려면 패키지 수준 실패와 빌드 실패를 각각 둘 이상 두고, 접두어가 겹치는 패키지 경로도 하나 두어야 한다. 그렇게 하면 픽스처가 다시 커진다. 연작을 여기서 닫는 것은 운영자의 판단이다.
