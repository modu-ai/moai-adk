# t1228 감사 — census 픽스처에 같은 결과 테스트 쌍 추가

- 감사 대상: 커밋 `ff325ac4b` (기준 `c46b263d5`), 작업트리 `.claude/worktrees/t1228`, 브랜치 `WT-census-per-test`
- 감사자: sync-auditor (독립 감사, 구현 파일은 읽기만 함)
- 적용 규칙: `.claude/rules/moai/core/verification-claim-integrity.md` §1·§2·§3, `.claude/rules/moai/development/verification-completeness.md` §1.1(관측된 실패)·§2(변이 탐침)
- 변이는 모두 세션 스크래치패드(`.../scratchpad/t1228-audit/`)에 `git archive`로 받은 사본에만 넣었다. 추적 파일은 이 보고서 외에 건드리지 않았다.

## 판정

**PASS** — 조화평균 94.3

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 95 | PASS | 기준 트리에서 x5·x6·x7 생존 재현, 대상 트리에서 7개 변이 모두 FAIL·각자 한 칸만 변동, 기대 출력을 독립 도출해 바이트 일치 |
| Security (25%) | 100 | PASS | 테스트 픽스처와 주석만 바뀜. 비밀값 패턴 0건, 새 입력 경로 없음 |
| Craft (20%) | 88 | PASS | 새 이벤트의 필드 구성은 실제 스트림과 같다. 다만 실제 스트림이 늘 내는 `=== RUN`·`--- FAIL/PASS/SKIP` 출력 이벤트가 빠졌다(F1). 레인 판정서 증거가 생략부호로 요약됨(F4) |
| Consistency (15%) | 95 | PASS | 기존 픽스처 형식·이름 규칙·주석 구조를 따름. 커밋 메시지 규칙과 카드 id 표기 준수 |

조화평균: 4 / (1/95 + 1/100 + 1/88 + 1/95) = 94.3. must-pass 차원(Functionality·Security) 모두 통과. 차단(blocking) 발견 사항은 없다.

## 검증 1 — 기준 `c46b263d5`에서 x5·x6·x7 생존

하네스 `mutate.sh <트리>`는 사본의 `test-census.sh` 한 줄에 sed 변이를 넣고, 원본과 `diff`해 바뀐 줄 수를 센 뒤 `census-check.sh`를 돌린다. 변이식:

| 이름 | 행 | sed 식 |
|---|---|---|
| x5 | 151 | `s/\| "\\(\.Package)\\t\\(\.Test)"/\| .Package/` |
| x6 | 152 | 같음 |
| x7 | 154 | 같음 |
| m1 | 141 | `select(.Action=="skip")` → `select(.Action=="skip" and (.Test // null) != null)` |
| m2 | 171 | ` \| sed 's/%/%25/g'` 제거 |
| m3 | 151 | ` and (.Test // null) != null` 제거 |
| m5 | 153 | `== null) \| .Package` → `!= null) \| .Package` |

변이가 의도대로 들어갔는지는 같은 sed 식을 원본에 적용해 출력으로 확인했다(151·152·154행이 `| .Package)`로, 141행이 `.Test != null` 필터로, 171행이 `$(printf '%s' "$ln")`로 바뀜).

```
$ bash .../t1228-audit/mutate.sh .../t1228-audit/base
clean: census-check: PASS (census output matches .../base/scripts/ci-census/testdata/expected.txt) rc=0
expected totals: === totals: packages=5 passed=2 skipped=2 nothing-ran=1 failed=1 build-failed=1 ===
x5   changed_lines=1 rc=0 census-check: PASS (census output matches .../base.mut/scripts/ci-census/testdata/expected.txt)
x6   changed_lines=1 rc=0 census-check: PASS (...)
x7   changed_lines=1 rc=0 census-check: PASS (...)
m1   changed_lines=1 rc=1 census-check: FAIL — ...   diff: -NOTHING RAN   example.com/censusfix/beta|
m2   changed_lines=1 rc=1 census-check: FAIL — ...   diff: -...SPEC-FIXTURE-100%25/...|+...SPEC-FIXTURE-100%/...|
m3   changed_lines=1 rc=1 ...  totals: === totals: packages=5 passed=3 skipped=2 nothing-ran=1 failed=1 build-failed=1 ===
m5   changed_lines=1 rc=1 ...  totals: === totals: packages=5 passed=2 skipped=2 nothing-ran=2 failed=1 build-failed=1 ===
```

(경로만 `...`로 줄였다. 판정·합계 줄은 원문 그대로다.) 레인의 주장 1(세 변이가 이전 픽스처를 통과)이 재현된다.

## 검증 2 — 대상 `ff325ac4b`에서 깨끗한 census PASS, 7개 변이 모두 FAIL

```
$ bash .../t1228-audit/mutate.sh .../t1228-audit/head
clean: census-check: PASS (census output matches .../head/scripts/ci-census/testdata/expected.txt) rc=0
expected totals: === totals: packages=5 passed=3 skipped=3 nothing-ran=1 failed=2 build-failed=1 ===
x5   changed_lines=1 rc=1  totals: === totals: packages=5 passed=2 skipped=3 nothing-ran=1 failed=2 build-failed=1 ===
x6   changed_lines=1 rc=1  totals: === totals: packages=5 passed=3 skipped=2 nothing-ran=1 failed=2 build-failed=1 ===
x7   changed_lines=1 rc=1  totals: === totals: packages=5 passed=3 skipped=3 nothing-ran=1 failed=1 build-failed=1 ===
m1   changed_lines=1 rc=1  totals: (unchanged)  diff: -NOTHING RAN   example.com/censusfix/beta|
m2   changed_lines=1 rc=1  totals: (unchanged)  diff: -...SPEC-FIXTURE-100%25/...|+...SPEC-FIXTURE-100%/...|
m3   changed_lines=1 rc=1  totals: === totals: packages=5 passed=4 skipped=3 nothing-ran=1 failed=2 build-failed=1 ===
m5   changed_lines=1 rc=1  totals: === totals: packages=5 passed=3 skipped=3 nothing-ran=2 failed=2 build-failed=1 ===
```

x5·x6·x7·m3·m5는 diff가 합계 줄 하나뿐이고, 바뀐 칸은 각각 passed / skipped / failed / passed / nothing-ran 하나다. m1·m2는 합계를 바꾸지 않고 각자의 행 하나만 바꾼다. 레인 주장 3과 일치한다.

로케일 의존성도 확인했다(CHANGELOG에 `sort -u` 로케일 문제가 기록돼 있음). 새 행은 기존 행과 접두 관계라 정렬 결과가 로케일에 따라 달라지지 않는다.

```
$ LC_ALL=C bash scripts/ci-census/census-check.sh
census-check: PASS (census output matches /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1228/scripts/ci-census/testdata/expected.txt)
$ LC_ALL=en_US.UTF-8 bash scripts/ci-census/census-check.sh
census-check: PASS (census output matches /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1228/scripts/ci-census/testdata/expected.txt)
```

## 검증 3 — 기대 출력을 독립적으로 도출해 바이트 비교

`test-census.sh`의 본문을 쓰지 않고, 그 파일 머리의 OUTPUT CONTRACT 주석만 보고 jq 한 프로그램(`derive.jq`, `jq -rs`)으로 census 전체를 다시 썼다. 행 정렬은 jq `unique`(코드포인트 순), 본문은 출력 이벤트를 이어 붙여 줄마다 두 칸 들여쓴다.

```
$ jq -rs -f .../derive.jq scripts/ci-census/testdata/fixture.jsonl > .../derived.txt
$ cmp .../derived.txt scripts/ci-census/testdata/expected.txt
(출력 없음, exit 0)
$ shasum -a 256 .../derived.txt scripts/ci-census/testdata/expected.txt
2cd2669b91f4647d911dc14e8569dd55aa2c59b2e34580036072107431814bcc  .../derived.txt
2cd2669b91f4647d911dc14e8569dd55aa2c59b2e34580036072107431814bcc  .../expected.txt
```

도출기가 동어반복이 아님을 두 대조로 확인했다.

```
$ cmp .../derived-base.txt .../base/scripts/ci-census/testdata/expected.txt     # 양성 대조: 기준 픽스처 → 기준 기대 출력
(출력 없음, exit 0)
$ cmp .../derived.txt .../base/scripts/ci-census/testdata/expected.txt          # 음성 대조: 새 도출 → 옛 기대 출력
.../derived.txt .../base/.../expected.txt differ: char 327, line 9
```

들여쓰기와 순서는 `cat -vet`로 직접 봤다. FAILED 본문 줄은 `      alpha_test.go:30: want nil error$`(출력 자체의 4칸 + census 접두 2칸 = 6칸)이고, SKIPPED 행은 `TestEpsilonSkips` 다음에 `TestEpsilonSkipsToo`, FAILED 행은 `TestAlphaFails` 본문 3줄 다음에 `TestAlphaFailsToo`가 온다. 레인이 손으로 고친 기대 출력은 정확하다.

## 검증 4 — 새 이벤트가 실제 `go test -json` 모양인가

스크래치 모듈에 같은 이름의 테스트(통과 2·skip 2·실패 2)를 만들어 실제 스트림을 받았다(go test -C … -json -count=1 ./..., exit 1은 실패 테스트 때문). 새 테스트 한 건의 실제 이벤트:

```
{"Time":"2026-09-26T10:32:20.546087+09:00","Action":"run","Package":"example.com/censusreal/eps","Test":"TestEpsilonPassesToo"}
{"Time":"2026-09-26T10:32:20.546095+09:00","Action":"output","Package":"example.com/censusreal/eps","Test":"TestEpsilonPassesToo","Output":"=== RUN   TestEpsilonPassesToo\n"}
{"Time":"2026-09-26T10:32:20.546106+09:00","Action":"output","Package":"example.com/censusreal/eps","Test":"TestEpsilonPassesToo","Output":"--- PASS: TestEpsilonPassesToo (0.00s)\n"}
{"Time":"2026-09-26T10:32:20.546114+09:00","Action":"pass","Package":"example.com/censusreal/eps","Test":"TestEpsilonPassesToo","Elapsed":0}
```

실패 테스트도 `=== RUN   TestAlphaFailsToo`, 메시지, `--- FAIL: TestAlphaFailsToo (0.00s)` 세 출력 이벤트를 낸다. 픽스처의 새 이벤트는 다음과 같다.

```
$ jq -c 'select(.Test=="TestAlphaFailsToo" or .Test=="TestEpsilonPassesToo" or .Test=="TestEpsilonSkipsToo") | {Test, Action, keys: (keys|map(select(.!="Time")))}' scripts/ci-census/testdata/fixture.jsonl
{"Test":"TestAlphaFailsToo","Action":"run","keys":["Action","Package","Test"]}
{"Test":"TestAlphaFailsToo","Action":"output","keys":["Action","Output","Package","Test"]}
{"Test":"TestAlphaFailsToo","Action":"fail","keys":["Action","Elapsed","Package","Test"]}
{"Test":"TestEpsilonPassesToo","Action":"run","keys":["Action","Package","Test"]}
{"Test":"TestEpsilonPassesToo","Action":"pass","keys":["Action","Elapsed","Package","Test"]}
{"Test":"TestEpsilonSkipsToo","Action":"run","keys":["Action","Package","Test"]}
{"Test":"TestEpsilonSkipsToo","Action":"skip","keys":["Action","Elapsed","Package","Test"]}
```

이벤트마다 필드 구성은 실제와 같다. 다른 점은 출력 이벤트의 개수다. 통과·skip 테스트에는 출력 이벤트가 전혀 없고, 실패 테스트에는 메시지 한 줄뿐이다(`=== RUN`·`--- FAIL` 없음). 그래서 기대 출력의 `TestAlphaFailsToo` 본문은 1줄인데, 실제 스트림이면 3줄이다.

지금 census에 영향이 있는가: census가 출력 이벤트를 읽는 곳은 세 군데다. (a) 실패 테스트 본문은 `.Test==$t`로 거르므로 새 통과·skip 테스트와 무관하다. (b) FAILED PKG 본문은 `.Test == null`인 출력만 본다. (c) notice는 `absent-from-snapshot ` 포함 줄만 본다. 따라서 현재 경로에서 결과를 바꾸는 곳은 없다. 레인의 Residual-risk 판단은 측정으로 확인됐다.

앞으로 skip 사유를 출력하게 census를 바꾸면: `TestEpsilonSkipsToo`는 본문이 비고, 사유를 패키지 단위로 붙이는 잘못된 구현이라면 `TestEpsilonSkips`의 `--- SKIP` 줄이 `TestEpsilonSkipsToo` 아래에 붙어 기대 출력과 달라지므로 오히려 드러난다. 가려지는 결함은 찾지 못했다. 다만 실제로는 생길 수 없는 빈 본문이 그때 기대 출력에 고정될 수 있다(F1).

## 검증 5 — 다른 소비처와 주석 정확성

```
$ git grep -ln -e 'fixture.jsonl' -e 'testdata/expected.txt' -- ':!.moai/reports' ':!.moai/specs' ':!CHANGELOG.md'
scripts/ci-census/census-check.sh
```

픽스처를 읽는 곳은 `census-check.sh` 하나이고, 그 스크립트는 `ci.yml` 210행 `run: bash scripts/ci-census/census-check.sh`에서만 호출된다(`test` 잡). `.moai/reports`·`.moai/specs`·CHANGELOG의 언급은 과거 기록이며 실행 경로가 아니다.

주석: 목록 항목이 8개이고 머리글이 「all eight shapes」다. 8번 항목의 설명(epsilon의 통과 2·skip 2, alpha의 실패 2, 패키지 단위로 세면 더 작은 수)은 픽스처와 검증 2의 x5·x6·x7 결과(2<3, 2<3, 1<2)와 맞는다. 7번 항목의 「그런 패키지가 둘」(alpha·epsilon)도 새 픽스처에서 여전히 참이다.

```
$ bash -n scripts/ci-census/census-check.sh
(출력 없음, exit 0)
$ shellcheck scripts/ci-census/census-check.sh
(eval):1: command not found: shellcheck      # 도구 없음 → Gap
```

## 검증 6 — 추가 변이 (정보)

하네스 `mutate2.sh`로 9개를 더 시험했다. 기준·대상 모두에서 돌렸다.

| 이름 | 변이 | 기준 | 대상 |
|---|---|---|---|
| y1 | 107행 실패 본문 필터에서 `.Test==$t` 제거 | FAIL | FAIL |
| y2 | 63행 `MAX_FAILURE_BODIES` 기본값 50→1 | **PASS(생존)** | **FAIL** — `-      alpha_test.go:30: want nil error` / `+  (captured output omitted beyond the first 1 failures …)` |
| y3 | 105행 `-le`→`-lt` | 생존 | 생존 |
| y4 | 145행 skip 행 `sort -u`→`cat` | 생존 | 생존 |
| y5 | 98행 실패 목록 `sort -u`→`cat` | 생존 | 생존 |
| y6 | 151행 passed 키를 `.Test`만으로 | 생존 | 생존 |
| y7 | 152행 skipped 키를 `.Test`만으로 | 생존 | 생존 |
| y8 | 154행 failed 키를 `.Test`만으로 | 생존 | 생존 |
| y9 | 150행 packages를 `start` 이벤트로 세기 | 생존 | 생존 |

- y2는 이번 보강이 덤으로 잡게 된 변이다(한 패키지 실패 2건 덕분).
- y3은 실패가 51건 이상이어야 드러난다. 픽스처로 다루기엔 과하다.
- y4·y5가 사는 이유는 픽스처 이벤트가 이미 정렬된 순서로 놓여 있어서다. 실제 CI 스트림은 여러 패키지가 병렬로 섞여 나오므로 정렬 제거는 실제 결함이다(F3).
- y6·y7·y8은 x5·x6·x7의 반대쪽이다. 픽스처에 두 패키지에서 같은 이름을 쓰는 테스트가 없어서, 쌍의 패키지 성분을 버려도 통과한다(F2).
- y9는 t1223 감사가 등가 변이로 분류한 x1과 같다.

## 발견 사항

- **F1** [Low] [optional] 신뢰도 높음 — `scripts/ci-census/testdata/fixture.jsonl:29-31, 46-49` — 새 테스트 3건에 실제 스트림이 늘 내는 출력 이벤트가 없다. 통과·skip 테스트는 출력이 0건이고, `TestAlphaFailsToo`에는 `=== RUN`·`--- FAIL` 줄이 없다(검증 4의 실측과 대조). 현재 census 경로에는 영향이 없다. — 필요하다면: 각 테스트에 `=== RUN   <이름>`과 `--- PASS/SKIP/FAIL: <이름> (0.00s)` 출력 이벤트를 넣고, `expected.txt`의 `TestAlphaFailsToo` 본문을 3줄로 고친다.
- **F2** [Low] [optional] 신뢰도 높음 — 같은 파일 — 두 패키지에 같은 이름의 테스트가 없어 passed·skipped·failed를 `.Test`만으로 세는 변이(y6·y7·y8)가 산다. 실제 저장소에서는 `TestNew` 같은 이름이 여러 패키지에 흔하므로 실측될 수 있는 결함 모양이다. 카드 범위(패키지 단위 집계) 밖이다. — 필요하다면 후속 카드: 다른 패키지에 기존과 같은 이름의 통과·skip·실패 테스트를 하나씩 둔다. 이때 `census-check.sh` 8번 항목 설명(「counting packages instead prints a smaller number」)도 두 방향을 모두 적도록 넓힌다.
- **F3** [Low] [optional] 신뢰도 중간 — 같은 파일 — 이벤트가 이미 정렬 순서로 놓여 있어 `sort -u` 제거 변이(y4·y5)가 산다. — 필요하다면: 한 패키지의 이벤트를 다른 패키지 이벤트 사이에 끼워 병렬 실행처럼 섞는다.
- **F4** [Low] [optional] 신뢰도 높음 — `.moai/reports/t1228/verdict.md:28-30, 48` — 증거가 `census output matches …`처럼 생략부호로 잘려 있어 원문 출력이 아니다(verification-claim-integrity §3). 또 57-61행의 「변이마다 자기가 건드린 합계 칸 하나만 달라졌다」는 m1·m2에는 맞지 않는다. 두 변이는 합계를 바꾸지 않고 행을 바꾼다. 결론 자체는 이번 감사가 재현했다. — 필요하다면: 명령과 출력 원문을 붙이고, 해당 문장을 「x5·x6·x7·m3·m5는 합계 칸 하나만, m1·m2는 행 하나만 달라졌다」로 고친다.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

- **Claim**: 커밋 `ff325ac4b`는 카드 목적을 달성했다. 이전 픽스처에서 살던 x5·x6·x7이 이제 FAIL하고, 기존 m1·m2·m3·m5도 계속 FAIL하며, 깨끗한 census는 PASS다. `expected.txt`는 독립 도출과 바이트 단위로 같다.
- **Evidence**: 위 검증 1~6의 명령과 출력.
- **Baseline-attribution**: 이번 세션에서 `git archive`로 받은 `c46b263d5`·`ff325ac4b`의 `scripts/ci-census` 사본과, HEAD `ff325ac4b46a0a5045357af0135985850794ce07`이고 `git status --short`가 비어 있는 작업트리에서 측정했다. macOS의 bash·BSD sed·jq, 로컬 go 도구체인을 썼다.
- **Gaps**: CI 실행은 보지 않았다(push 전). `shellcheck`가 없어 정적 분석은 `bash -n`만 했다. 교차 모델 감사(`audit_multi`)는 돌리지 않았다. 레인이 쓴 변이 사본 자체는 보지 못했고, 같은 변이를 이번 감사가 따로 만들어 재현했다. 스크래치 하네스(`mutate.sh`, `mutate2.sh`, `derive.jq`)는 스크래치패드에만 있어 나중에 지워질 수 있다. 변이식은 이 보고서 표에 옮겨 두었다.
- **Residual-risk**: 픽스처 하나가 모든 변이를 구별하지는 못한다(F2·F3의 생존 변이). GNU sed와 리눅스 로케일에서의 census 출력은 CI의 `census-check` 단계가 처음으로 확인하게 된다.
