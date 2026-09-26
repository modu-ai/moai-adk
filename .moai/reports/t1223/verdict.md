# t1223 — census 픽스처가 totals 변이 두 개를 구별하도록 보강

## Claim

1. t1218 감사(F1)가 찾은 totals 변이 두 개는 기존 픽스처에서 `census-check.sh`를 통과했다. 이를 재현했다.
   - m3: passed 집계에서 `.Test != null` 조건을 뺀 변이
   - m5: nothing-ran 조건을 `== null`에서 `!= null`로 뒤집은 변이
2. 원인은 픽스처에 있었다.
   - 패키지 수준 pass 이벤트가 없어서, m3가 그것까지 세도 수가 바뀌지 않았다.
   - skip 테스트를 가진 패키지와 테스트 파일이 없는 패키지가 둘 다 1개여서, m5가 둘을 뒤바꿔 세도 수가 같았다.
3. 픽스처에 패키지 `epsilon`을 추가하고 기대 출력(`expected.txt`)을 손으로 고쳤다. 이제 m3와 m5가 모두 FAIL로 잡힌다. t1218의 변이 m1·m2도 계속 잡히고, 깨끗한 census는 PASS다.

## Evidence

기준: 작업트리 `.claude/worktrees/t1223`, 브랜치 `WT-census-fixture-totals`, 로컬 develop `6ae4583e6`. 변이는 모두 세션 스크래치패드의 사본에만 넣었다.

**변이 정의** (`scripts/ci-census/test-census.sh`)

| 이름 | 위치 | 변이 |
|---|---|---|
| m3 | 151행 | `select(.Action=="pass" and (.Test // null) != null)` → `select(.Action=="pass")` |
| m5 | 153행 | `(.Test // null) == null` → `(.Test // null) != null` |
| m1 | 141행 | `select(.Action=="skip")` → `select(.Action=="skip" and (.Test // null) != null)` (t1218) |
| m2 | 171행 | `sed 's/%/%25/g'` → `cat` (t1218) |

**재현 — 보강 전 픽스처**

```
bash <scratch>/m3/census-check.sh → census-check: PASS (census output matches …)
bash <scratch>/m5/census-check.sh → census-check: PASS (census output matches …)
```

**보강** — `fixture.jsonl`에 `example.com/censusfix/epsilon` 패키지 이벤트 9줄을 추가했다.
- 통과 테스트 `TestEpsilonPasses`
- skip 테스트 `TestEpsilonSkips`
- 패키지 수준 `pass` 이벤트
이로써 skip 테스트를 가진 패키지는 2개(alpha, epsilon)가 되고, 테스트 파일이 없는 패키지는 1개(beta)로 남는다.

`expected.txt`는 census를 돌려 다시 만들지 않고 손으로 고쳤다. 검사 대상의 출력으로 기대값을 만들면 순환이 되기 때문이다.
- `SKIPPED TEST  example.com/censusfix/epsilon  TestEpsilonSkips` 줄을 추가했다.
- 합계 줄을 `packages=5 passed=2 skipped=2 nothing-ran=1 failed=1 build-failed=1`로 바꿨다.

`census-check.sh` 머리 주석의 픽스처 형태 목록에 7번 항목(epsilon)을 추가하고, 형태 수를 「five」에서 「seven」으로 고쳤다. 원래 여섯 항목을 나열하면서도 「five」라고 적혀 있었다.

**보강 후**

```
bash scripts/ci-census/census-check.sh → census-check: PASS (census output matches …expected.txt)
```

| 변이 | census-check | diff (합계 줄) |
|---|---|---|
| m3 | exit 1, FAIL | `-passed=2` / `+passed=3` |
| m5 | exit 1, FAIL | `-nothing-ran=1` / `+nothing-ran=2` |
| m1 | FAIL | (t1218과 같은 `NOTHING RAN` 줄 누락) |
| m2 | FAIL | (t1218과 같은 notice 줄 `%25` 차이) |

m3·m5의 diff는 보강 설계에서 예측한 값(passed 3, nothing-ran 2)과 정확히 같다.

**소비처** — 픽스처와 기대 출력을 읽는 곳은 `census-check.sh` 하나뿐이다(`grep -rn 'ci-census/testdata\|fixture.jsonl\|censusfix'`). 이 스크립트는 t1218이 `ci.yml` `test` 잡에 연결해 두었다.

## Baseline-attribution

작업트리 `.claude/worktrees/t1223`, HEAD `6ae4583e6` 위의 변경. 위 결과는 모두 이 트리에서 이번 실행으로 관측했다.

## Gaps

- 실제 CI 실행은 보지 않았다(push·CI 요청 금지).
- m1과 m2의 결과는 첫 줄(FAIL)만 확인했고, diff 본문은 다시 옮겨 적지 않았다.
- 이 네 개 말고 다른 변이 공간은 조사하지 않았다. 예를 들어 packages 집계, failed 집계, build-failed 집계를 바꾸는 변이는 시험하지 않았다.
- GNU sed와 GNU jq(CI ubuntu)에서 출력이 같은지는 재지 않았다. 추가한 이벤트는 기존 이벤트와 같은 모양이다.

## Residual-risk

- 픽스처는 합성 이벤트다. 독립 감사(`audit.md` F3)가 go1.26.8로 실제 스트림을 만들어 비교했다. 이벤트 종류별 필드 구성은 같았다. 패키지 수준 pass는 `{Action, Elapsed, Package}`였다. 실제 스트림에 있는 출력 이벤트 3줄(`=== RUN` 2줄, `PASS`)이 픽스처에는 없다. 그러나 실제 이벤트로 바꿔 돌려도 출력이 `expected.txt`와 바이트까지 같았다(F2).
- 픽스처 하나로 모든 변이를 구별하는 것은 원리적으로 불가능하다. 이번 보강은 감사가 찾은 두 변이를 닫았을 뿐이다. 독립 감사(F1)가 여전히 살아남는 변이 셋을 찾았다. passed·skipped·failed를 (패키지, 테스트) 쌍이 아니라 패키지 단위로 세는 변이다. 한 패키지 안에 같은 결과를 낸 테스트가 둘 이상 있는 경우가 픽스처에 없어서 구별되지 않는다. 이 카드 범위 밖이며 후속 후보다.
