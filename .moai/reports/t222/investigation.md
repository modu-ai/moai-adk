# t222 — CI 비결정 실패 2건 조사

- 카드: t222 (Class B — 원인 미상, 조사 선행)
- 워크트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t222`
- 브랜치: `WT-ci-flaky-failures` (base `ec1837038`)
- 조사 일자: 2026-08-24

두 건 모두 **원인 후보를 실측으로 특정**했다. 재현은 여전히 안 되지만, 각 건마다
"관측 없이 짐작한 것"이 아니라 **명령을 돌려 본 결과**가 근거로 남아 있다.

---

## 0. 방법론 결함 하나 먼저 — 재실행이 실패를 지운다

`gh run list` 가 돌려주는 `conclusion` 은 **마지막 attempt** 의 결과다. attempt 1
이 실패하고 attempt 2 가 통과한 run 은 목록에서 `success` 로 보인다. 카드가 근거로
삼은 "최근 CI 60건 중 이 1건뿐" 은 그래서 **과소 계수**다.

실측 (최근 CI 100 run):

```
gh run list --limit 100 --workflow=ci.yml   →  run id 100개
각 id 에 대해 gh api …/actions/runs/<id> → run_attempt
```

`attempts.tsv` 기준 `run_attempt > 1` 인 run 4건:

| run | head | attempt | 현재 conclusion | attempt 1 실패 잡 |
|---|---|---|---|---|
| 32687843472 | 7ddeb8099 | 2 | **success** | Test (ubuntu-latest), Race Test |
| 32447398787 | ff75553ad | 2 | **success** | Test (ubuntu-latest) |
| 32445462169 | 6f138f04e | 2 | **success** | Test (ubuntu-latest) |
| 32429213275 | ecf9e3376 | 2 | failure | Test (ubuntu-latest) |

즉 **3건의 실패가 목록상 초록으로 보이고 있었다.** 앞으로 비결정 실패의 기저율을
셀 때는 `run_attempt` 를 함께 읽어야 한다.

---

## 1. 건 1 — TestBranchGuard_Latency ratio 1.82x

### 관측된 것

```
timing.go:209: checkBranchState: n=100 median=3.431ms p95=10.182ms worst=11.621ms
               avg=4.187ms | refUnit=1.884ms ratio=1.82x (maxUnits=1.50x, …)
```
(run 32687843472 attempt 1, 2026-08-24 — `failing-attempt1-excerpt.log`)

같은 계열 실패가 **처음이 아니다.** 2026-08-20 run 32429213275 attempt 1:

```
timing.go:209: paired-cpu-1x: n=20 median=3.536ms p95=7.187ms worst=9.187ms
               avg=3.704ms | refUnit=1.3ms ratio=2.72x (maxUnits=2.00x, …)
```

이쪽은 `internal/timing` **하네스 자신의 자가 테스트** `TestAssertPairedHealthyEndToEnd`
이고, ref 와 fn 이 **바이트 단위로 동일한 함수**다. 참 비율은 정의상 1.00x. 그런데
2.72x 가 나왔다 — 즉 **calibrated arm 은 회귀가 없는 코드에 대해 회귀를 보고한 전력이
있다.** `#1591` 커밋 메시지는 CI 에서 2.32x / 2.72x / 4.64x 를 관측했다고 기록한다.

### 왜 이게 이 카드에 결정적인가

실패 메시지는 "this is a code regression, not machine load (load inflates the
reference equally)" 라고 단언한다. 그 단언이 **자가 테스트에서 이미 반증됐다.**

`#1591` (bd463f55e) 이 진단한 기전은 정확했다: **부하의 계단(step) 변화**. 다만 그
수정은 `internal/timing` **패키지 내부의 병렬 측정 이웃을 제거**하는 방식이라,
계단의 발생원 하나만 막았고 추정기 자체는 그대로다. `internal/hook` 은 그 수정의
사정권 밖이다.

### 추정기의 결함 (기계적으로 시연됨)

`AssertPaired` 는 라운드마다 ref 1개 / fn 1개를 **교대로** 재고, 그러고 나서
`median(fn) / median(ref)` 를 계산했다. 이 나눗셈이 **짝을 버린다.** 두 median 은
서로 독립한 순서통계량이고, 두 계열이 계단을 **한 라운드 차이로** 넘으면(교대
순서가 교차점에서 정확히 그 오프셋을 만든다) 두 median 이 계단의 양쪽에 앉을 수
있다.

`internal/timing/paired_step_test.go` `TestCalibratedRatioSurvivesOffsetLoadStep`
가 이걸 합성 표본으로 시연한다 — 라운드별 참 비율이 **모든 라운드에서 1.00x** 인
계열인데:

```
offset-split@ 48  ratio-of-medians=1.00x   median-of-round-ratios=1.00x
offset-split@ 49  ratio-of-medians=1.89x   median-of-round-ratios=1.00x   ← 칼날
offset-split@ 50  ratio-of-medians=1.00x   median-of-round-ratios=1.00x
```

`1.89x` 는 CI 가 본 `1.82x` 와 크기가 일치한다(합성 계단비 3600/1900 = 1.89, CI
관측 3.431/1.884 = 1.82). 칼날은 100 라운드 중 1~2 라운드 폭이라 대략 1~2% —
"60 run 에 1건" 과 모순되지 않는다.

**주의(정직성):** 이 시연은 추정기가 그런 값을 **만들어낼 수 있음**을 기계적으로
보인 것이지, CI 에서 실제로 그렇게 됐다는 증명은 아니다. 다만 이것이 지금까지 나온
후보 중 **유일하게 시연된** 것이고, 회귀 가설은 카드가 이미 닫았다(같은 head 재실행
통과 = 결정적 회귀 아님).

### 계단의 발생원도 하나 찾았다 (§2와 같은 뿌리)

`newBranchGuardRepoFixture` 는 seed 커밋을 하나 만든다. 그 커밋이
`git maintenance run --auto --no-quiet --detach` 를 띄운다(§2 실측). **detach** 된
프로세스는 커밋이 반환된 뒤에도 살아서 CPU 를 쓰고, 그 직후 시작되는 100 라운드
측정 중에 끝난다 — 즉 **측정 자신이 자기 측정 창 안에 계단을 하나 공급하고 있었다.**

### 수정

1. `internal/timing`: `AssertPaired` 가 **라운드별 비율의 median** 을 강제하도록
   변경(`measurePaired` 가 세 번째 값으로 반환, `CheckRatio` 가 그 값을 받음).
   각 비율의 두 항은 마이크로초 간격으로 측정되므로 계단이 항 안에서 상쇄된다.
   `Assert`(비짝 형태)는 `ratioOfMedians` 로 종전 동작 유지.
2. `internal/hook`: 픽스처에 `gc.auto=0`, `maintenance.auto=false` — 자기 측정
   창 안의 계단 발생원 제거.
3. 회귀 테스트 2개: 계단 내성(위 시연 + 짝 안 쓴 추정기가 칼날에서 실제로
   트립하는지 확인하는 반증 arm) + **held-out**: 라운드마다 4배 비싼 fn 은 여전히
   트립해야 한다(`TestPairedRatioStillCatchesRealCostGrowth`).

### 검증

```
go test -count=1 ./internal/timing/                       → ok 0.886s
go test -count=1 -v -run TestBranchGuard_Latency ./internal/hook/
  timing.go:233: checkBranchState: n=100 median=43.324ms p95=77.767ms worst=122.142ms
                 avg=47.181ms | refUnit=43.37ms ratio=0.99x (maxUnits=1.50x, …)
  --- PASS (10.70s)
```

(로컬 부하가 높아 절대값이 43ms 로 부풀었는데 **비율은 0.99x** — 보정 arm 이
설계대로 작동한다는 뜻.)

---

## 2. 건 2 — TestReadCardStatus_DoesNotSearchBranchSet cleanup 실패

### 관측된 것

```
testing.go:1464: TempDir RemoveAll cleanup: unlinkat
  /tmp/TestReadCardStatus_DoesNotSearchBranchSet287648166/001/.git/objects: directory not empty
```

RemoveAll 이 objects/ 를 비운 뒤 rmdir 하려는 사이에 **누군가 다시 썼다**는 모양.
카드가 열어 둔 질문: **쓰는 주체 미특정.**

### 쓰는 주체 후보를 실측했다

`probe-gc.sh` — 픽스처와 같은 git 시퀀스(init → commit → branch/checkout/commit ×3)를
`GIT_TRACE=1` 로 돌린 결과:

```
4× trace: run_command: git maintenance run --auto --no-quiet --detach
git version 2.50.1 (Apple Git-155)
```

**커밋마다 detach 된 git 프로세스가 하나씩 뜬다.** 픽스처 repo 하나당 4개. 이들은
테스트가 기다린 git 명령이 반환된 **뒤에도 살아서** `.git` 을 만진다.

카드가 기각했던 근거(`gc.auto` 기본 임계 6700 loose object 에 한참 못 미침)는
**다른 기전을 겨냥한 것**이었다. `git maintenance run --auto` 는 먼저 detach 하고
나서 할 일이 있는지 판단한다 — 객체 수는 spawn 자체를 막지 못한다.

`probe-gc-off.sh` — `gc.auto=0` + `maintenance.auto=false` 를 걸면:

```
0 detached maintenance spawns
```

### 이 프로세스가 테스트 바이너리보다 오래 산다는 증거

같은 명령을 픽스처 수정 전후로 돌린 귀속 가능한 비교(같은 트리, 두 줄 config 차이):

| | 8회 반복 테스트 시간 합 | 패키지 wall time |
|---|---|---|
| 수정 전 (`repro-race8.log`) | ≈ 16.9s | **297.556s** |
| 수정 후 (`repro-race8-after.log`) | ≈ 13.9s | **15.284s** |

수정 전에는 테스트가 다 끝난 뒤 **약 280초의 정체**가 있었다. `go test` 는 테스트
바이너리의 stdout 파이프가 EOF 될 때까지 기다리는데, detach 된 자식이 그 fd 를
상속해 들고 있으면 정확히 이 모양이 된다. 수정 후 그 정체가 사라졌다.

**이것이 곧 cleanup 실패의 증명은 아니다.** 증명된 것은 (a) 테스트가 기다리지 않은
git 프로세스가 repo 를 만지며 살아 있다는 사실, (b) 그 프로세스가 테스트 바이너리
수명을 넘긴다는 사실 — RemoveAll 과의 경쟁이 성립하기 위한 **선행 조건 두 개**다.
어느 파일을 언제 썼는지는 관측하지 못했다.

### 수정

`specFixtureRepo` 에 `gc.auto=0`, `maintenance.auto=false`. 알려진 동시 writer
계열 전체를 제거하며, 테스트가 검증하는 어떤 동작도 건드리지 않는다.

### 재현 시도 (실패 — 정직한 음성)

```
go test -race -run TestReadCardStatus_DoesNotSearchBranchSet -count=8 ./internal/kanban/
  → 8/8 PASS (수정 전)
go test -race -count=1 ./internal/kanban/   → ok 38.199s (수정 후, 전체 패키지)
```

로컬에서 원 실패는 재현되지 않았다. 수정은 **재현으로 검증된 것이 아니라**, 측정된
동시 writer 를 제거한 것이다.

---

## 3. 남는 것 (Gaps / Residual risk)

- **CI 에서 실제로 무슨 일이 있었는지는 여전히 관측되지 않았다.** 두 건 다 로컬
  재현 실패. 위 수정은 "실측된 결함 있는 추정기" 와 "실측된 동시 writer" 를
  제거한 것이고, 그것이 CI 실패의 원인이었다는 증명은 아니다.
- **ubuntu CI 의 git 이 같은 detach 동작을 하는지 직접 관측하지 못했다.** git
  2.30+ 공통 동작이지만 CI 러너에서 `GIT_TRACE` 를 뜨워 확인하지는 않았다.
- **다른 픽스처에도 같은 detach 발생원이 남아 있다.** 이번 변경은 실패한 두 테스트의
  픽스처(`specFixtureRepo`, `newBranchGuardRepoFixture`)만 손댔다. 커밋을 만드는
  다른 테스트 픽스처는 그대로 — 별도 카드감(스윕 후보).
- **판정은 CI.** 로컬 초록은 조기 신호일 뿐이다.
- 재발 여부를 앞으로 볼 때는 §0 때문에 반드시 `run_attempt` 를 함께 읽을 것.

---

## 증거 파일

| 파일 | 내용 |
|---|---|
| `failing-attempt1-excerpt.log` | 원 실패 2건 + 2026-08-20 자가 테스트 실패의 CI 로그 발췌 |
| `attempts.tsv` | 최근 CI 100 run 의 `run_attempt` / conclusion (§0 근거) |
| `runids.txt` | 위 조회에 쓴 run id 목록 |
| `probe-gc.sh` | 커밋마다 detach 된 maintenance 가 뜨는 것을 보이는 프로브 |
| `probe-gc-off.sh` | 두 config 로 spawn 이 0 이 되는 것을 보이는 프로브 |
| `repro-race8.log` / `repro-race8-after.log` | 픽스처 수정 전/후 귀속 비교 |
| `repro-race20.log` | 20회 반복이 timeout 한 최초 실행(행이 아니라 느림) |
