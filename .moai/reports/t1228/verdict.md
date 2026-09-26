# t1228 — census 픽스처가 합계를 패키지 단위로 세는 변이를 구별하도록 보강

## Claim

1. t1223 감사(F1)가 찾은 변이 세 개(x5·x6·x7)는 이전 픽스처로 `census-check.sh`를 통과했다. 이를 재현했다. 세 변이는 passed·skipped·failed 합계를 (패키지, 테스트) 쌍이 아니라 패키지 단위로 센다. 원인은 이전 픽스처에 같은 결과를 낸 테스트가 한 패키지 안에 둘 이상인 경우가 없었다는 데 있다.
2. 픽스처를 보강했다. epsilon에는 통과 테스트와 skip 테스트를 하나씩 더 넣었다. alpha에는 실패 테스트를 하나 더 넣었다. 기대 출력(`expected.txt`)은 손으로 고쳤다.
3. 보강 후 세 변이와 기존 변이 m1·m2·m3·m5, 모두 일곱 개가 FAIL로 잡힌다. 깨끗한 census는 PASS다.

## Evidence

기준은 작업트리 `.claude/worktrees/t1228`, 브랜치 `WT-census-per-test`, 로컬 develop `c46b263d5`이다. 변이는 모두 세션 스크래치패드의 사본에만 넣었다. 사본마다 원본과 `diff`해서 바뀐 줄이 정확히 1줄씩인 것을 확인했다.

**변이 정의** (`scripts/ci-census/test-census.sh`)

| 이름 | 위치 | 변이 |
|---|---|---|
| x5 | 151행 passed | 집계 키 `"\(.Package)\t\(.Test)"` → `.Package` |
| x6 | 152행 skipped | 같음 |
| x7 | 154행 failed | 같음 |
| m1 | 141행 | skip 판정 필터 분리 (t1218) |
| m2 | 171행 | notice의 `%` 이스케이프 제거 (t1218) |
| m3 | 151행 | passed에서 `.Test != null` 조건 제거 (t1223) |
| m5 | 153행 | nothing-ran 조건 뒤집기 (t1223) |

**재현 — 보강 전 픽스처(`c46b263d5`)**

```
x5 → census-check: PASS (census output matches …)
x6 → census-check: PASS (census output matches …)
x7 → census-check: PASS (census output matches …)
```

**보강** — `fixture.jsonl`에 이벤트 7줄을 추가했다.
- alpha `TestAlphaFailsToo`: run, output, fail. alpha의 실패 테스트가 2개가 된다.
- epsilon `TestEpsilonPassesToo`: run, pass. epsilon의 통과 테스트가 2개가 된다.
- epsilon `TestEpsilonSkipsToo`: run, skip. epsilon의 skip 테스트가 2개가 된다.

`expected.txt`는 census를 돌려 만들지 않았다. 픽스처 이벤트에서 따져 손으로 고쳤다.
- `FAILED        example.com/censusfix/alpha  TestAlphaFailsToo` 행과 본문 한 줄 `      alpha_test.go:30: want nil error`를 추가했다.
- `SKIPPED TEST  example.com/censusfix/epsilon  TestEpsilonSkipsToo` 행을 추가했다.
- 합계를 `packages=5 passed=3 skipped=3 nothing-ran=1 failed=2 build-failed=1`로 고쳤다.

`census-check.sh` 머리 주석에는 8번 형태를 추가하고 「seven」을 「eight」로 고쳤다.

**보강 후**

```
bash scripts/ci-census/census-check.sh → census-check: PASS (census output matches …expected.txt)
```

| 변이 | census-check | 합계 줄(변이 출력) |
|---|---|---|
| x5 | FAIL | `passed=2` (정답 3) |
| x6 | FAIL | `skipped=2` (정답 3) |
| x7 | FAIL | `failed=1` (정답 2) |
| m1 | FAIL | (NOTHING RAN 행 누락) |
| m2 | FAIL | (notice `%25` 차이) |
| m3 | FAIL | `passed=4` (정답 3) |
| m5 | FAIL | `nothing-ran=2` (정답 1) |

변이마다 자기가 건드린 합계 칸 하나만 달라졌다. 다른 칸은 정답과 같다.

## Baseline-attribution

작업트리 `.claude/worktrees/t1228`, HEAD `c46b263d5` 위의 변경이다. 위 결과는 모두 이 트리에서 이번 실행으로 관측했다.

## Gaps

- 실제 CI 실행은 보지 않았다(push·CI 요청 금지).
- m1·m2는 FAIL 첫 줄만 확인했고, diff 본문은 다시 옮겨 적지 않았다.
- t1223 감사가 등가 변이로 분류한 x1(start 이벤트로 패키지 세기)과 x4(`FailedBuild`로 build-failed 세기)는 다시 시험하지 않았다.
- 추가한 이벤트는 기존 이벤트와 같은 모양을 따랐다. 이번 카드에서는 실제 `go test -json` 스트림과 다시 대조하지 않았다. t1223 감사가 epsilon 이벤트 모양을 실측으로 확인한 적은 있다.

## Residual-risk

- 추가한 통과·skip 이벤트에는 출력(`output`) 이벤트가 없다. 실제 스트림은 `=== RUN`과 `--- PASS`/`--- SKIP` 줄을 낸다. census는 통과·skip 테스트의 출력을 쓰지 않으므로 결과에는 영향이 없다고 본다. 측정하지는 않았다.
- 픽스처 하나로 모든 변이를 구별할 수는 없다. 이번 보강은 감사가 찾은 세 변이를 닫았을 뿐이다.
