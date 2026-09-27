# t1218 — census-check.sh 를 CI 에 연결

## Claim

1. `scripts/ci-census/census-check.sh`(census 스크립트의 자체 검사)는 CI와 Makefile 어디에도 연결돼 있지 않았다. `test-census.sh`는 스트림을 읽을 수 있으면 항상 0으로 끝나는 보고 도구다. 그래서 census의 출력이 틀려도 CI는 초록이다. 변이 두 개로 이 공백을 재현했다.
2. `ci.yml`의 `test` 잡에 `census-check.sh`를 도는 단계를 테스트 단계 바로 앞에 추가했다. 같은 두 변이를 이 단계가 이제 잡는다.

## Evidence

기준: 작업트리 `.claude/worktrees/t1218`, 브랜치 `WT-census-check-wire`. t1215 병합분을 포함한 로컬 develop `5ac030965`을 앞으로 감기로 흡수한 뒤 측정했다.

**연결 전 상태**

```
grep -rn 'census-check' .github/workflows Makefile
(출력 없음)
```

CI의 census 호출은 `test` 잡의 테스트 단계 안에 있는 `bash scripts/ci-census/test-census.sh test-stream.json` 한 줄뿐이다. 이 단계는 `go test`의 종료 코드(`exit $rc`)를 돌려준다.

**기준선** — 깨끗한 트리: `bash scripts/ci-census/census-check.sh` → `census-check: PASS`.

**변이 1 — skip 필터 분리(픽스처가 막으려고 만든 형태)** — 스크래치 사본의 `test-census.sh` 141행 `select(.Action=="skip")`을 `select(.Action=="skip" and (.Test // null) != null)`로 바꿨다.

| 실행 | 결과 |
|---|---|
| 변이된 `test-census.sh` 를 픽스처에 (현재 CI 단계와 같은 호출) | exit 0, `NOTHING RAN   example.com/censusfix/beta` 줄이 사라짐 → CI 단계는 초록 |
| 변이된 트리에서 `census-check.sh` | exit 1, diff `-NOTHING RAN   example.com/censusfix/beta` |

**변이 2 — notice 절의 `%` 이스케이프 제거** — 171행의 `sed 's/%/%25/g'`를 `cat`으로 바꿨다.

| 실행 | 결과 |
|---|---|
| 변이된 `test-census.sh` 를 픽스처에 | exit 0, `SPEC-FIXTURE-100%25` 가 `SPEC-FIXTURE-100%` 로 출력됨 → CI 단계는 초록 |
| 변이된 트리에서 `census-check.sh` | exit 1, diff 가 notice 줄 하나를 가리킴 |

**독립 감사 재측정** — 감사자가 m1·m2와 totals 변이 m4(`failed=` 제거)를 다시 돌렸다. `census-check.sh`는 셋 모두 exit 1이었다. Actions 셸 흉내(`bash --noprofile --norc -eo pipefail`)에서도 결과가 같았고, actionlint v1.7.10도 rc=0이었다.

**연결** — `ci.yml`의 `test` 잡에 `Check test census against its fixture` 단계(`run: bash scripts/ci-census/census-check.sh`)를 넣었다. `Run tests with coverage` 바로 앞, 인덱스 8이다. 이 잡은 이미 `scripts/ci-census/**` 경로 필터에 걸려 있어, census만 바꾼 PR에서도 실행된다.

```
ruby -ryaml … → idx=8 next=Run tests with coverage (fast — no race detector) run=bash scripts/ci-census/census-check.sh
actionlint .github/workflows/ci.yml → 출력 없음(exit 0)
```

변이 1을 담은 스크래치 트리에서 이 단계의 명령과 같은 스크립트를 돌리면 exit 1이다. 깨끗한 트리에서는 PASS다(위 기준선). GitHub Actions는 단계 본문을 `bash -e`로 돌리므로 종료 코드 1은 곧 단계 실패다.

## Baseline-attribution

작업트리 `.claude/worktrees/t1218`, HEAD `5ac030965` 위의 변경, 이번 실행에서 관측했다. 변이는 모두 세션 스크래치패드의 사본에만 넣었고, 추적 파일은 건드리지 않았다.

## Gaps

- 이 변경으로 실제 CI 실행이 어떻게 되는지는 보지 않았다(push·CI 요청 금지). 로컬 측정과 actionlint만 근거로 삼았다.
- `bash -e -o pipefail` 형태 그대로는 격리 가드가 실행을 막아 평범한 `bash`로 돌렸다. 종료 코드 1이 `-e`에서 단계 실패가 된다는 것은 GitHub Actions의 기본 셸 동작에 근거한 추론이다.
- CI 러너(ubuntu-latest)에 `jq`가 있다는 것은 확인하지 않았다. 기존 테스트 단계가 이미 같은 스크립트(`test-census.sh`, jq 필요)를 부르고 있어 같은 조건이다. jq가 없으면 `census-check.sh`는 exit 2로 끝나므로 단계는 실패한다(초록으로 지나가지는 않는다).
- `release-pr-multi-os.yml`의 census 호출에는 이 검사를 붙이지 않았다. 카드가 「한 잡」에 연결하라고 했고, 같은 스크립트를 PR마다 `test` 잡이 검사한다.
- Makefile에는 연결하지 않았다.

## Residual-risk

- `census-check.sh`는 픽스처 하나의 출력을 통째로 비교한다. 픽스처가 담지 않은 형태로 census가 망가지면 이 검사도 놓친다. 독립 감사(`audit.md` F1)가 실제로 살아남는 totals 변이 두 개를 찾았다.
  - passed 집계에서 `.Test != null` 조건 제거: 픽스처에 패키지 수준 pass 이벤트가 없어서 차이가 나지 않는다.
  - nothing-ran 조건을 `== null`에서 `!= null`로 뒤집기: 픽스처에서 skip 테스트 수와 테스트 없는 패키지 수가 둘 다 1이라 서로 뒤바뀌어도 같다.
  두 변이 모두 census와 `census-check.sh`가 exit 0으로 끝나 CI를 통과한다. 고치려면 픽스처를 보강하고 `expected.txt`를 갱신해야 한다. 이 카드 범위 밖이라 후속 후보로 남긴다.
- 새 단계가 실패하면 같은 잡의 뒤 단계가 돌지 않는다. 테스트 단계뿐 아니라 Codecov 업로드도 건너뛴다. `needs: test`인 `test-integration`(3개 OS) 잡도 돌지 않는다(감사 F2). census가 망가진 PR에서는 이 결과들을 같은 실행에서 볼 수 없다.
- 픽스처 검사는 `test` 잡이 도는 ubuntu에서만 돈다. release 워크플로의 macOS·Windows 레그에서만 드러나는 census 결함은 잡지 못한다(감사 F4).
