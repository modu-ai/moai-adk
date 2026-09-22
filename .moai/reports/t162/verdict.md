# t162 — checkBranchState 캘리브레이션 게이트 판정

카드: t162 (Class B) · 브랜치 `WT-guard-latency` · base `origin/release/v3.1.1 @ 76ef8a764`

## 1. 판정: 회귀 아님 — 측정 방법 결함

Release Verify (windows-latest) 의 `checkBranchState` ratio 1.51x(bound 1.50x) 는 코드 회귀가 아니라
**참조 단위를 측정하는 방식**이 만든 값이다. 근거 세 갈래가 같은 방향을 가리킨다.

### 근거 1 — 재실행이 통과했다

`76ef8a764` 로 재실행된 CI 에서 windows 레그가 통과했다. 코드는 그대로다.

```
$ gh pr checks 1585
Release Verify (windows-latest)   pass   18m24s   .../job/96269721929
Release Verify (macos-latest)     pass   10m57s
Release Verify (ubuntu-latest)    pass    9m19s
```

리드가 정한 분기 규칙("재현되면 회귀 / 통과하면 측정 방법 소관")에 따라 측정 방법 쪽으로 간다.
다만 통과 자체는 약한 근거다 — 통과한 실행의 실제 수치는 로그에 남지 않았다(§근거 3).

### 근거 2 — 이번 배치는 이 호출 경로를 건드리지 않았다

`checkBranchState` 가 호출하는 전부:

| 단계 | 비용 |
|---|---|
| `extractBranchStateCommand` | JSON unmarshal (µs) |
| `matchBranchStateCommand` | regex (µs) |
| `isExemptAgent` | `os.Getenv` 1회 |
| `resolveProjectRootFromInputOrEnv` | `input.CWD` 반환 (분기 1회) |
| `isPrimaryCheckout` → `git.ResolveGitDirs` | **`git rev-parse` spawn 1회** (primary path) |

`internal/session`(t144 resolveSessionPID 조상 탐색)도 `internal/cli`(t148 런처 env 스탬프)도 이 그래프에
없다. `ResolveGitDirs` 의 2-spawn fallback 은 `--path-format=absolute` 가 거부될 때만 타는데, windows-latest
의 git 은 이 플래그를 지원한다(지원하지 않았다면 ratio 는 ~2.0x 로 훨씬 크게 나왔을 것). 리드가 짚은 대로
"관계가 없으면 그 자체가 잡음 쪽 증거"다.

### 근거 3 — 같은 코드로 옛 방식의 ratio 가 0.67~1.03 사이를 오간다

이것이 결정적 근거다. darwin/arm64, 동일 트리, 동일 코드, 5회 연속 실행 —
**옛 측정 방식**(CombinedOutput 참조를 n=10 버스트로 측정 루프 *앞에서* 한 번에 가격 매김):

| # | median | refUnit | ratio |
|---|---|---|---|
| 1 | 24.355ms | 23.534ms | 1.03x |
| 2 | 24.031ms | 23.355ms | 1.03x |
| 3 | 25.570ms | 26.700ms | 0.96x |
| 4 | 25.488ms | **37.882ms** | **0.67x** |
| 5 | 24.386ms | 24.183ms | 1.01x |

4번 실행에서 refUnit 이 37.9ms 로 튀었다 — 부하 스파이크가 하필 참조 버스트 위에 내려앉은 것이다.
분모가 60% 부풀면 ratio 는 0.67 로 내려간다. **같은 메커니즘이 반대로 작동하면 1.51 이 된다**:
스파이크가 참조 버스트가 아니라 측정 루프 안에 내려앉으면 분자만 부푼다. windows 실패 표본이 정확히
그 모습이다 — `worst=305.258ms`(median 의 12.8배), avg 25.561ms 가 median 23.845ms 위로 끌려올라간
반면 refUnit 은 15.788ms 로 앉아 있다. 참조는 그 스파이크를 본 적이 없다.

`internal/timing` 패키지 문서가 내건 전제는 "부하가 양쪽을 똑같이 부풀리므로 비율은 제자리"인데,
참조를 t=0 에 버스트로 한 번만 재고 나면 그 전제가 성립하지 않는다.

## 2. 무엇을 고쳤나 — 임계값이 아니라 측정 방법

임계값 1.5x 는 **손대지 않았다**. 세 가지가 바뀌었다.

1. **같은 창(window)에서 잰다** — `timing.AssertPaired` 가 참조 1회와 측정 1회를 번갈아 실행한다.
   부하 스파이크가 분자·분모 양쪽에 동시에 얹힌다.
2. **표본 수를 맞춘다** — 참조 10개 대 측정 100개였다. 이제 양쪽 다 `Iterations` 개.
   분모가 분자보다 시끄러운 상태를 없앤다.
3. **참조가 측정 대상의 cost MIX 를 그대로 흉내낸다** — 참조가 `CombinedOutput()`(파이프 1개)이었는데
   실제 `runGitRevParse` 는 stdout/stderr 버퍼를 따로 두고 `cmd.Run()` 한다(파이프 2개). 참조를 같은
   형태로 맞췄다. darwin 에서는 차이가 1.004 배로 무시할 수준이지만(측정함), windows 파이프 비용은
   측정하지 못했다 — 패키지 문서가 이미 요구하는 규칙이라 맞추는 쪽이 옳다.

추가로 **순서 교대**: 홀수 회차는 측정을 먼저 돌린다. 한쪽이 늘 다른 쪽의 따뜻한 캐시를 물려받지 않게.

### 결과 (darwin/arm64, 5회 연속)

| 방식 | ratio 5회 | 폭 |
|---|---|---|
| 옛 방식 | 1.03 / 1.03 / 0.96 / **0.67** / 1.01 | **0.36** |
| 새 방식 | 0.99 / 1.00 / 1.00 / 1.00 / 0.99 | **0.01** |

같은 코드, 같은 머신, 같은 세션. 실행 간 폭이 36배 줄었고 중심이 1.00 에 앉는다.
1.5x 임계에 대한 여유가 0.7% 에서 50% 로 회복된다.

### 통과한 실행의 수치도 이제 남는다

지금까지 healthy ratio 를 아무도 본 적이 없었다. `t.Logf` 는 `-v` 없이 통과한 패키지에서 버려지고
CI 는 `go test -race -timeout 25m ./...` 로 돈다(`release-pr-multi-os.yml:156`). 그래서 **붉게 죽은 실행
하나가 플랫폼별 유일한 표본**이었고, 1.51 이 높은 값인지 원래 그런지 판단할 근거가 없었다.

`GITHUB_STEP_SUMMARY` 가 설정돼 있으면 캘리브레이션 줄 한 개를 잡 요약에 덧붙인다. 로컬에서는 무동작,
파일을 못 열면 조용히 넘어간다(측정 기록 실패가 통과한 검증을 붉게 만들면 안 된다). 다음 windows 초록
실행부터 healthy 수치가 쌓인다.

## 3. Linux 통과 / Windows 실패 비대칭

부하 스파이크 가설이 이 비대칭도 설명한다. 스파이크가 ratio 를 움직이는 크기는 `스파이크 크기 / 기준
비용`에 비례하는데, windows 러너는 git.exe spawn 이 비싼 대신(#1225 의 872ms 관측) 스케줄링·바이러스
검사 지터도 훨씬 크다. 실패 표본의 `worst/median = 12.8` 이 그 증거다. 같은 잡의 ubuntu/macos 레그는
median 대비 worst 가 그만큼 벌어지지 않는다.

단, 이건 **가설의 정합성**이지 측정된 사실이 아니다. windows 러너에서 옛 방식 대 새 방식을 나란히 잰
표본은 없다(§5).

## 4. 변경 파일

| 파일 | 변경 |
|---|---|
| `internal/timing/timing.go` | `AssertPaired` + `measurePaired` / `timeOne` / `summarize` / `report` / `publish` 추가, `Assert` 는 `report` 위임으로 동작 불변, 패키지 문서에 same-window·equal-N 규칙 추가 |
| `internal/timing/paired_test.go` | 신규 — equal-N, warmup 양쪽 폐기, 순서 교대, end-to-end |
| `internal/timing/publish_test.go` | 신규 — 잡 요약 기록, env 부재 무동작, 쓰기 실패 무해 |
| `internal/hook/pre_tool_branch_guard_integration_test.go` | 참조를 `runGitRevParse` 형태로 교체 + `AssertPaired` 로 전환 + 주석 갱신 |

`Assert` 의 시그니처와 동작은 그대로다 — `internal/harness/observer_test.go` 는 손대지 않았다(§5).

## 5. 검증 (Claim / Evidence / Baseline / Gaps / Residual)

**Claim** — 임계값을 건드리지 않고 측정 분산을 36배 줄였으며, 영향 패키지가 전부 통과한다.

**Evidence** — 트리 `WT-guard-latency`, base `76ef8a764`:

```
$ go test ./internal/hook/... ./internal/timing/ -count=1 -timeout 900s
ok  github.com/modu-ai/moai-adk/internal/hook            48.429s
ok  github.com/modu-ai/moai-adk/internal/hook/handoff     5.111s
ok  github.com/modu-ai/moai-adk/internal/hook/memo        0.514s
ok  github.com/modu-ai/moai-adk/internal/hook/memo/taxonomy 1.172s
ok  github.com/modu-ai/moai-adk/internal/hook/mx         21.100s
ok  github.com/modu-ai/moai-adk/internal/hook/mx/complexity 1.705s
ok  github.com/modu-ai/moai-adk/internal/hook/perf       25.401s
ok  github.com/modu-ai/moai-adk/internal/hook/quality    10.105s
ok  github.com/modu-ai/moai-adk/internal/hook/security    9.500s
ok  github.com/modu-ai/moai-adk/internal/hook/testutil     3.907s
ok  github.com/modu-ai/moai-adk/internal/hook/trace       1.171s
ok  github.com/modu-ai/moai-adk/internal/timing           2.661s

$ go test -race ./internal/timing/ ./internal/hook/ -run 'Timing|Assert|Paired|Publish|BranchGuard' -count=1
ok  github.com/modu-ai/moai-adk/internal/timing           1.398s
ok  github.com/modu-ai/moai-adk/internal/hook            14.632s

$ GOOS=windows go vet ./internal/hook/ ./internal/timing/   # exit 0
$ gofmt -l internal/timing internal/hook/pre_tool_branch_guard_integration_test.go   # 출력 없음
```

ratio 5회 비교는 §2 표. 옛 방식 수치는 폐기한 임시 테스트(`zz_oldform_test.go`, 옛 배치를 그대로 재현)로
같은 세션에서 측정했고, 측정 후 삭제했다.

**Baseline-attribution** — 전부 `76ef8a764` 위 이 트리에서, 이번 실행에 측정. 다른 시점·다른 SPEC 의
수치를 끌어오지 않았다.

**Gaps (관측하지 않은 것)**

- **windows 러너에서 새 방식을 돌린 적이 없다.** 폭 0.01 은 darwin/arm64 수치다. windows 에서 ratio 가
  1.00 근처에 앉는지는 다음 CI 가 처음 보여준다. push 는 카드 범위 밖이다.
- **windows 파이프 비용**(CombinedOutput 1개 대 split 2개)은 측정하지 못했다. darwin 에서 1.004 배임은
  측정했다(임시 실험, 폐기). windows 에서 이 항이 유의한지는 모른다.
- **실패한 windows 실행의 옛 방식 재현**을 하지 못했다 — 그 러너·그 부하는 재현 불가.
- 통과한 windows 실행의 실제 ratio 는 **끝내 알 수 없다**. 잡 요약 기록이 들어간 것이 그 때문이고,
  이 카드가 남기는 값은 다음 실행부터 생긴다.

**Residual-risk**

- 부하 스파이크가 아니라 windows 고유의 **체계적** 차이(예: split-pipe spawn 이 windows 에서만 비쌈)가
  1.51 의 진짜 원인이라면, 인터리빙만으로는 부족할 수 있다. 참조 MIX 를 맞춘 것이 그 경우를 겨냥한
  보험이지만 확인된 바 없다. 다음 windows 실행의 잡 요약 줄이 판정한다.
- `internal/harness/observer_test.go` 는 여전히 옛 배치(`Median` + `Assert`)를 쓴다. 같은 결함 유형이고
  같은 이력(ubuntu 2.56x~3.61x 팽창, job 95500006280)을 갖고 있다. 지금 통과 중이라 카드 범위를 넘겨
  건드리지 않았다 — **후속 카드 후보**.
- `AssertPaired` 는 참조를 100회 더 돌리므로 이 테스트 벽시계가 ~3.1s 에서 ~5.3s 로 늘어난다
  (windows 18분 잡 기준 무시 가능).

## 6. 남은 일

push 금지 카드라 커밋까지만 했다. 통합·push·PR 은 리드 몫.
