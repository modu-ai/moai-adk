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

| 변이 | 판정을 가른 diff 줄 |
|---|---|
| y4 | `SKIPPED TEST  …/zeta  TestAlphaSkips` 행이 제자리(`-`)에서 빠지고 다른 위치(`+`)에 나타남 |
| y5 | `FAILED        …/zeta  TestAlphaFails` 행과 본문이 제자리에서 빠지고 alpha 행들 앞에 나타남 |
| y6 | `passed=3` (정답 4) |
| y7 | `skipped=3` (정답 4) |
| y8 | `failed=2` (정답 3) |
| x5 | `passed=3` (정답 4) |
| x6 | `skipped=3` (정답 4) |
| x7 | `failed=2` (정답 3) |
| m1 | `-NOTHING RAN   example.com/censusfix/beta` (행 누락) |
| m2 | notice의 `%25`가 `%`로 출력됨 |
| m3 | `passed=5` (정답 4) |
| m5 | `nothing-ran=3` (정답 1) |

y4·y5는 순서만 바뀌고 합계는 바뀌지 않는다. y6~y8·x5~x7·m3·m5는 자기 합계 칸 하나만 바뀐다. m1·m2는 행 하나만 바뀐다.

위 표의 `…`는 긴 경로(`example.com/censusfix`)를 줄인 표기다. 명령과 출력 원문은 스크래치패드 사본에서 다시 돌려 얻을 수 있다. 독립 감사도 이를 재현한다(`audit.md`).

## Baseline-attribution

작업트리 `.claude/worktrees/t1231`, HEAD `044fb91c7` 위의 변경이다. 위 결과는 모두 이 트리에서 이번 실행으로 관측했다. 「보강 전」은 픽스처를 고치기 전의 같은 트리이고, 「보강 후」는 편집 실수를 고친 뒤의 트리다.

## Gaps

- 실제 CI 실행은 보지 않았다(push·CI 요청 금지).
- 새 zeta 이벤트는 기존 이벤트와 같은 모양을 따랐다. 실제 `go test -json` 스트림과 다시 대조하지는 않았다.
- `count()` 안의 `sort -u`(157행 부근)를 빼는 변이와 `pkg_failed`·`pkgs_with_failed_tests`의 `sort -u`를 빼는 변이는 이번 범위(F2·F3가 지목한 y4·y5)에 넣지 않았고 돌리지 않았다.
- t1228 감사의 y3(`-le`→`-lt`, 실패 51건 이상 필요)와 y9(packages를 `start` 이벤트로 세기, 등가 변이)는 다시 시험하지 않았다.

## Residual-risk

- 픽스처 하나로 모든 변이를 구별할 수는 없다. 이번 보강은 감사가 지목한 다섯 변이를 닫았을 뿐이다.
- 운영 지시에 따라 이 카드로 census 변이 대조 연작(t1215·t1218·t1223·t1228·t1231)을 닫는다. 독립 감사에서 새로 살아남는 변이가 나오더라도 후속 카드로 잇지 않고, 아래 「한계」에 기록만 한다.

## 한계 (독립 감사 반영 전)

감사 결과를 반영하면서 이 절을 채운다.
