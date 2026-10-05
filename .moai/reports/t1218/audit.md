# t1218 사후 감사 — census-check.sh 의 CI 연결

- 감사 대상: 커밋 `162d0874c` (기준 `5ac030965`), 브랜치 `WT-census-check-wire`
- 작업트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1218` (감사 중 HEAD `162d0874c` 고정, `git status --short` 비어 있음)
- 판정 프로필: 내장 기본값(Functionality 40 / Security 25 / Craft 20 / Consistency 15, 필수 통과 Functionality·Security)
- 변이는 모두 스크래치 사본(`.../scratchpad/t1218-audit/<변이명>/ci-census/`)에만 넣었다. 추적 파일은 이 보고서 외에 건드리지 않았다.

## 종합 판정: PASS-WITH-DEBT (조화평균 91.8)

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 88 | PASS | 레인이 든 변이 2개 재현, 연결 후 둘 다 적발. 추가 totals 변이 3개 중 1개 적발, 2개는 픽스처 한계로 생존(F1) |
| Security (25%) | 97 | PASS | 새 단계는 `${{ }}` 보간 없음, 비밀 미사용, 권한 `contents: read` 유지 |
| Craft (20%) | 88 | PASS | actionlint 통과, 주석 정확, 실행 0.33초. 판정서 잔여 위험이 구체성 부족(F2) |
| Consistency (15%) | 95 | PASS | `shell: bash`·선행 검사 단계 배치가 같은 잡의 기존 관례(sg·python3·exec.Command 검사)와 일치 |

조화평균: 4 / (1/88 + 1/97 + 1/88 + 1/95) = 91.8. 차단(blocking) 결함 없음. 부채는 F1(픽스처가 totals 변이 두 형태를 구별하지 못함)이며, 이것은 이번 카드가 만든 결함이 아니라 픽스처에 원래 있던 한계다.

## 주장별 검증

### 주장 1 — 연결 전에는 아무것도 census-check.sh 를 부르지 않았다: 확인

```
$ git grep -n 'census-check' 5ac030965 -- .github Makefile ; echo "rc=$?"
rc=1

$ git grep -ln 'census-check\|ci-census' 5ac030965 -- ':!.moai/specs' ':!.moai/reports'
5ac030965:.github/workflows/ci.yml
5ac030965:.github/workflows/release-pr-multi-os.yml
5ac030965:CHANGELOG.md
5ac030965:scripts/ci-census/census-check.sh
5ac030965:scripts/ci-census/test-census.sh
```

기준 트리 전체에서 `census-check.sh` 를 부르는 곳은 없다(워크플로 2곳은 `test-census.sh` 호출, 나머지는 스크립트 자신과 CHANGELOG). 기존 CI 단계는 `go test … || rc=$?; bash scripts/ci-census/test-census.sh test-stream.json; exit $rc` 형태로, 단계의 종료 코드는 `go test` 의 것이다. `test-census.sh` 는 변이 상태에서도 exit 0(아래 표)이므로 `-e` 아래에서도 단계를 멈추지 않는다.

### 주장 2 — 두 변이는 test-census.sh 를 0으로 두고 census-check.sh 를 1로 만든다: 확인, 추가 변이로 범위 확장

하네스 `run.sh <이름> <sed식>` 은 `scripts/ci-census` 를 스크래치로 복사해 변이를 넣고, 픽스처에 `test-census.sh` 를 돌린 종료 코드와 `census-check.sh` 의 종료 코드를 기록한다.

```
[base] census file identical to tracked
[base] test-census.sh exit=0
[base] census-check.sh exit=0
[m1-skipsplit] census file mutated: 2 changed lines
[m1-skipsplit] test-census.sh exit=0
[m1-skipsplit] census-check.sh exit=1
-NOTHING RAN   example.com/censusfix/beta
[m2-noescape] census file mutated: 2 changed lines
[m2-noescape] test-census.sh exit=0
[m2-noescape] census-check.sh exit=1
-::notice title=AC snapshot absent::.moai/specs/SPEC-FIXTURE-100%25/acceptance.md: HALT AC-FIXTURE-001 AC-FIXTURE-002 (...)
+::notice title=AC snapshot absent::.moai/specs/SPEC-FIXTURE-100%/acceptance.md: HALT AC-FIXTURE-001 AC-FIXTURE-002 (...)
[m3-passed-anytest] census file mutated: 2 changed lines
[m3-passed-anytest] test-census.sh exit=0
[m3-passed-anytest] census-check.sh exit=0
[m4-totals-drop-failed] census file mutated: 2 changed lines
[m4-totals-drop-failed] test-census.sh exit=0
[m4-totals-drop-failed] census-check.sh exit=1
-=== totals: packages=4 passed=1 skipped=1 nothing-ran=1 failed=1 build-failed=1 ===
+=== totals: packages=4 passed=1 skipped=1 nothing-ran=1 build-failed=1 ===
[m5-nothingran-swap] census file mutated: 2 changed lines
[m5-nothingran-swap] test-census.sh exit=0
[m5-nothingran-swap] census-check.sh exit=0
```

| 변이 | 내용 | test-census.sh | census-check.sh |
|---|---|---|---|
| m1 | 141행 skip 필터에 `and (.Test // null) != null` 추가 | 0 | **1** |
| m2 | 171행 `sed 's/%/%25/g'` → `cat` | 0 | **1** |
| m3 | 151행 passed 집계에서 `.Test != null` 조건 제거 | 0 | 0 (생존) |
| m4 | 157행 totals 줄에서 `failed=$failed` 제거 | 0 | **1** |
| m5 | 153행 nothing-ran 집계 `== null` → `!= null` | 0 | 0 (생존) |

m3·m5 의 생존은 동치 변이가 아니다. 픽스처 밖 스트림에서 실제로 출력이 달라진다.

```
$ jq -r 'select(.Action=="pass") | "\(.Package) \(.Test // "<pkg>")"' fixture.jsonl
example.com/censusfix/alpha TestAlphaPasses

# 테스트 pass 1건 + 패키지 수준 pass 1건
base: === totals: packages=1 passed=1 skipped=0 nothing-ran=0 failed=0 build-failed=0 ===
m3:   === totals: packages=1 passed=2 skipped=0 nothing-ran=0 failed=0 build-failed=0 ===

# 테스트 없는 패키지 1개 + skip 테스트 3건(패키지 2개)
base: === totals: packages=3 passed=0 skipped=3 nothing-ran=1 failed=0 build-failed=0 ===
m5:   === totals: packages=3 passed=0 skipped=3 nothing-ran=2 failed=0 build-failed=0 ===
```

픽스처에는 패키지 수준 pass 이벤트가 없고(모든 패키지가 fail·skip·build-fail), skip 테스트 1건과 테스트 없는 패키지 1개가 우연히 같은 수 1이라 두 변이가 보이지 않는다. 실제 CI 스트림에는 패키지 수준 pass 가 수백 건 있으므로 m3 는 실사용에서 passed 수를 크게 부풀리는데도 이번에 연결한 검사는 초록이다.

추가로 `count()` 의 `grep -c . || true` 를 `wc -l | tr -d " "` 로 바꾼 변이도 생존했지만(`census-check.sh exit=0`), 이것은 사실상 동치 변이라 결함으로 세지 않는다.

### 주장 3 — 새 단계의 위치와 트리거: 확인

```
$ git diff --stat 5ac030965 162d0874c
 .github/workflows/ci.yml       |  8 ++++++
 .moai/reports/t1218/verdict.md | 61 ++++++++++++++++++++++++++++++++++++++++++

$ actionlint .github/workflows/ci.yml ; echo "rc=$?"
rc=0
$ actionlint --version | head -1
v1.7.10

$ python3 (yaml.safe_load .github/workflows/ci.yml)
triggers: {'push': ['main', 'develop'], 'pull_request': ['main'], 'workflow_dispatch': None}
7 Assert zero exec.Command in MX validator (AC-UTIL-001-04) | if= None
8 Check test census against its fixture | if= None
9 Run tests with coverage (fast — no race detector) | if= None
test.if = needs.detect.outputs.go_code == 'true'
integration.needs = test
```

- 단계는 `test` 잡 인덱스 8, `Run tests with coverage` 바로 앞이다.
- `detect` 의 `go_code` 필터에 `scripts/ci-census/**` 와 `.github/workflows/ci.yml` 가 있어, census 만 바꾼 변경과 이번 워크플로 변경 모두 실제 `test` 잡을 돌린다(스킵 마커 잡이 아니다).
- 트리거는 develop·main push 와 main 대상 PR 이다. GitFlow 체인에서 리드의 develop 일괄 push 와 release→main PR 이 모두 이 잡을 지난다. develop 대상 PR 은 `pull_request` 트리거에 없지만, 이 저장소의 체인은 develop 으로 PR 을 내지 않으므로 공백이 아니다.

GitHub Actions 셸(`bash --noprofile --norc -eo pipefail`)을 흉내 낸 실행:

```
actions-shell step [base] exit=0
actions-shell step [m1-skipsplit] exit=1
actions-shell step [m2-noescape] exit=1
actions-shell step [m4-totals-drop-failed] exit=1
actions-shell step [m3-passed-anytest] exit=0
actions-shell step [m5-nothingran-swap] exit=0
census-check with jq absent exit=2
test-census: jq not found on PATH
census-check: census script exited non-zero
```

레인이 추론으로 남긴 두 가지(「종료 코드 1 은 곧 단계 실패」, 「jq 가 없으면 exit 2 로 실패」)를 로컬 재현으로 확인했다. jq 는 기존 테스트 단계의 `test-census.sh` 도 이미 요구하므로 새 가정이 아니다.

필수 검사 영향:

```
$ grep -n 'Test (ubuntu-latest)' .github/required-checks.yml
12:      - "Test (ubuntu-latest)"
24:      - "Test (ubuntu-latest)"
```

잡 이름 `Test (${{ matrix.os }})` 와 행렬이 그대로이고, `test-skip-marker` 는 `go_code != 'true'` 의 정확한 여집합이라 필수 컨텍스트 보고 방식은 바뀌지 않는다. census 가 깨지면 필수 검사 `Test (ubuntu-latest)` 가 빨개지는데, 이것이 카드가 원한 결과다.

### 배치 판단

단계를 테스트 앞에 둔 결과, census 가 깨진 실행에서는 `Run tests with coverage`, Codecov 업로드, 그리고 `needs: test` 인 `test-integration`(3-OS)이 모두 돌지 않는다. 대안과 비교하면 다음과 같다.

| 선택지 | 필수 검사에 걸리는가 | 테스트 결과 가시성 | 비용 |
|---|---|---|---|
| 테스트 앞(채택) | 걸린다 | census 가 깨진 실행에서는 안 보임 | 0.33초로 즉시 실패, 러너 시간 절약 |
| 테스트 뒤 + `if: success() \|\| failure()` | 걸린다 | 보인다 | census 가 깨져도 테스트 스위트를 다 돈다 |
| 별도 잡 | `required-checks.yml` 과 스킵 마커 여집합을 함께 손대지 않으면 걸리지 않는다 | 보인다 | 새 체크 이름 관리 부담 |

별도 잡은 필수 검사에 묶이지 않아 카드 목적을 약하게 만든다. 테스트 뒤 배치는 정보가 조금 더 많지만, census 결함은 census 를 고치면 끝나는 결함이라 같은 실행의 테스트 결과를 잃는 손실이 작다. 채택안은 같은 잡의 기존 선행 검사(sg·python3·exec.Command)와 같은 모양이다. 결함으로 보지 않고 참고 사항(F3)으로만 남긴다. 참고로 `if: always()` 는 체크아웃·셋업 실패에서도 돌아 잡음이 생기므로, 뒤로 옮긴다면 `success() || failure()` 가 맞다.

### release-pr-multi-os.yml 과 Makefile

release→main PR 은 `ci.yml` 의 `pull_request: [main]` 로도 돌기 때문에 `test` 잡의 census 검사를 이미 지난다. `release-pr-multi-os.yml` 에 같은 검사를 붙이면 ubuntu 에서는 중복이다. 다만 그 워크플로는 macOS·Windows 러너에서도 `test-census.sh` 를 돌리는데, 픽스처 검사는 ubuntu 에서만 돈다. BSD sed·Git Bash 차이로 census 가 특정 OS 에서만 틀리는 경우는 잡지 못한다(F4). Makefile 의 `ci-local` 은 언어별 범용 미러(`scripts/ci-mirror/run.sh`)라 `ci.yml` 단계를 복제하지 않으므로 여기에 연결하지 않은 것은 누락이 아니다.

## Findings (structured defect-list)

- **F1** [Medium] [optional] `scripts/ci-census/testdata/fixture.jsonl` — 픽스처가 totals 집계의 두 변이(m3: passed 에 패키지 수준 pass 포함, m5: nothing-ran 조건 반전)를 구별하지 못한다. 카드가 「notice 절 또는 totals」를 명시했으므로 totals 보호는 부분적이다. 신뢰도 높음(재현 완료). 이번 카드가 만든 결함은 아니다. — 필요한 수정: 후속 카드로 픽스처에 패키지 수준 `{"Action":"pass","Package":…}` 이벤트를 넣고, skip 테스트 수와 테스트 없는 패키지 수를 서로 다르게(예: 2 대 1) 만든 뒤 `expected.txt` 를 갱신한다. 수정 후 m3·m5 가 `census-check.sh exit=1` 이 되는지 확인한다.
- **F2** [Low] [optional] `.moai/reports/t1218/verdict.md:58-61` — 잔여 위험이 「픽스처가 담지 않은 형태」라는 일반론에 머물고, totals 변이는 하나도 시험하지 않았다. 또 단계 실패 시 건너뛰는 것이 테스트 단계뿐이라고 적었지만 Codecov 업로드와 `test-integration`(3-OS)도 함께 건너뛴다. 신뢰도 높음. — 필요한 수정: 잔여 위험에 m3·m5 생존 사실과 하류 잡 생략을 한 줄씩 덧붙인다(병합을 막을 사항은 아니다).
- **F3** [Info] [optional] `.github/workflows/ci.yml:208` — 테스트 앞 배치 때문에 census 가 깨진 실행에서는 테스트·커버리지·통합 테스트 결과를 같은 실행에서 볼 수 없다. 위 배치 판단대로 수용 가능. 신뢰도 중간. — 필요한 수정: 없음. 가시성을 우선한다면 테스트 뒤로 옮기고 `if: success() || failure()` 를 단다.
- **F4** [Low] [optional] `.github/workflows/release-pr-multi-os.yml:211` — 픽스처 검사는 ubuntu 에서만 돈다. macOS·Windows 에서만 드러나는 census 결함은 잡히지 않는다. 카드가 「한 잡」을 지정했으므로 범위 밖이다. 신뢰도 중간(OS 차이로 실제 결함이 나는지는 재지 않았다). — 필요한 수정: 필요하면 후속 카드에서 release 워크플로의 각 OS 레그에 `census-check.sh` 한 줄을 추가한다.
- **F5** [Info] [optional] `.github/workflows/ci.yml:206-207` — 주석의 「is the only thing that fails when it breaks」는 픽스처가 담은 형태에 한해 참이다(F1). 신뢰도 높음. — 필요한 수정: F1 을 고치거나, 주석에 「픽스처가 담은 형태에 대해」를 덧붙인다.

## Recommendations

- 이 커밋은 그대로 develop 병합 대상으로 둔다. 차단 결함이 없다.
- F1 은 별도 후속 카드로 올리는 것을 권한다. 이번 연결이 켠 검사의 실효 범위를 넓히는 가장 싼 수단이다.
- F2 는 판정서 보강 수준이며 병합 조건이 아니다.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

- **Claim**: 커밋 `162d0874c` 는 카드 목적을 달성했다. 연결 전 공백과 연결 후 적발을 모두 재현했다. 단 totals 보호는 픽스처가 구별하는 형태에 한정된다.
- **Evidence**: 위 각 절의 명령과 원문 출력.
- **Baseline-attribution**: 작업트리 `.claude/worktrees/t1218`, HEAD `162d0874c`, 기준 `5ac030965`, 이번 감사 실행에서 측정. 도구 버전은 `jq-1.8.1`, `actionlint v1.7.10`, macOS bash 이다.
- **Gaps**:
  - 실제 GitHub Actions 실행은 보지 않았다. Actions 셸은 로컬 `bash --noprofile --norc -eo pipefail` 로 흉내 냈다.
  - ubuntu-latest 러너 이미지에 jq 가 있다는 것은 직접 확인하지 않았다. 기존 테스트 단계가 이미 jq 에 의존하며, 없으면 exit 2 로 실패한다는 것만 확인했다.
  - macOS BSD sed 로 측정했다. GNU sed(ubuntu)에서 픽스처 출력이 같은지는 CI 실행 전까지 미관측이다. 다만 `census-check.sh` 는 기준선(깨끗한 트리)에서도 CI 에서 처음 도는 것이라, GNU 환경 차이가 있다면 첫 develop push 에서 초록으로 지나가지 않고 빨간불로 드러난다.
  - 교차 모델 감사(`audit_multi`/`codex_audit`/`glm_audit`)는 돌리지 않았다. 제공되는 대상(`uncommittedChanges`/`baseBranch`)이 이 단일 커밋만 겨누지 못해, 돌리면 이 카드와 무관한 develop↔main 차이 전체를 검토하게 된다.
- **Residual-risk**:
  - F1 의 두 totals 변이 형태는 여전히 CI 를 초록으로 통과한다.
  - 픽스처가 담지 않은 다른 스트림 형태(여러 notice 조합, 비정형 출력 등)에 대한 census 결함도 같은 이유로 놓칠 수 있다.
  - census-check 단계 실패 시 같은 실행의 테스트·커버리지·통합 테스트 결과가 사라진다.
