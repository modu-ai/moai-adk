# t370 — measurements (t354 수리 이후 잔여)

트리: `.claude/worktrees/t370`, 브랜치 `WT-race-residual`.
`git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t370`.
`git fetch origin develop` 후 `origin/develop` = `1e5199b88`.
수리 병합 `728f91006` 은 아래 다섯 head 전부의 조상 (`git merge-base --is-ancestor` 전부 exit 0).

## 측정원 — 로그가 아니라 `go test -json` 이벤트 스트림

`ci.yml` 의 `test` / `test-race` 잡이 각각 `test-stream.json.gz` 를 아티팩트로 올린다
(`test-stream-ci-test-ubuntu-latest` / `test-stream-ci-race-ubuntu-latest`, retention 7일).
`gh run download <id> -n <name>` → `gunzip` → `jq`. 로그 grep 보다 정밀하다 —
테스트별 elapsed 가 초 단위로 들어 있고, 실패 테스트 전수를 집합으로 뽑을 수 있다.

`d7010f86a`(run `33314551261`) 에는 test-stream 아티팩트가 **없다**:

```
$ gh api repos/modu-ai/moai-adk/actions/runs/33314551261/artifacts --jq '.artifacts[]|[.name,.expired]|@tsv'
moai-darwin-amd64  false      moai-linux-arm64  false     moai-windows-amd64  false
moai-linux-amd64   false      moai-darwin-arm64 false
```

바이너리 5개뿐 — 관측 배선이 이 head 이후에 들어왔다. 이 head 는 로그로만 판독 가능(별도 판독).

## 표 — Race Test 대 비-race Test, 같은 커밋 같은 러너 클래스

명령(각 head 반복):

```
gh run download <run-id> -n test-stream-ci-race-ubuntu-latest -D <dir>
gunzip -kf <dir>/test-stream.json.gz
jq -r 'select(.Action=="fail" and .Test!=null)|[.Package,.Test,.Elapsed]|@tsv' <dir>/test-stream.json | sort -u
grep -c "DATA RACE" <dir>/test-stream.json
jq -r 'select(.Test=="TestConcurrencyStress" and .Action=="output")|.Output' <dir>/test-stream.json
```

| head | CI run | Race Test `TestConcurrencyStress` | 실패 adds | `DATA RACE` | 비-race `Test (ubuntu)` 같은 테스트 | kanban 패키지 elapsed |
|---|---|---|---|---|---|---|
| `52c3fe590` | 33316389199 | **FAIL 2.15s** | 1/48 | **0** | **pass 0.76s** | 18.672 |
| `1728136c7` | 33317217710 | **FAIL 4.30s** | 7/48 | **0** | **pass 0.52s** | 17.612 |
| `d8a1a8e4e` | 33318836582 | **FAIL 1.97s** | 1/48 | **0** | **pass 0.53s** | 19.595 |
| `1e5199b88` | 33320417834 | **FAIL 2.88s** | 3/48 | **0** | **pass 0.54s** | 14.988 |
| `d7010f86a` | 33314551261 | FAIL (스트림 없음) | — | — | 잡 conclusion `success` | — |

`Test (ubuntu-latest)` 잡 conclusion 은 다섯 head 전부 `success`
(`gh api .../jobs --jq '.jobs[]|[.name,.conclusion]|@tsv'`).
즉 **실패는 `-race` 잡 전용**이다.

## 실패 테스트 전수 — 원 카드 질문 2

네 스트림의 `Action=="fail" and .Test!=null` 집합 전체:

- `52c3fe590`: **2건** — `internal/kanban TestConcurrencyStress` 2.15 · `internal/graph TestGitDiffNameCount_Predicate` 0.19
- `1728136c7`: 1건 — `TestConcurrencyStress` 4.30
- `d8a1a8e4e`: 1건 — `TestConcurrencyStress` 1.97
- `1e5199b88`: 1건 — `TestConcurrencyStress` 2.88

**`52c3fe590` 에는 두 번째 실패 테스트가 있다.** 리드가 미판독이라고 밝힌 두 head 중 하나이고,
`--log-failed` 3건 인용만으로는 안 보였을 항목이다. `TestGitDiffNameCount_Predicate` 는
기존 t352 소관 계열 — 이 카드에 귀속하지 않는다.

## `DATA RACE` — 직접 관측

```
$ grep -c "DATA RACE" <각 스트림>
0   0   0   0
```

네 head 전부 0. **data race 가 아니다** — 인용이 아니라 이번 실행의 측정이다.
(`d7010f86a` 는 스트림 부재로 미측정 → Gap.)

## 실패 문장 원문

```
backlog_concurrency_test.go:56: 7/48 adds failed under contention; first: mutate backlog
  /tmp/TestConcurrencyStress2606360510/001/backlog.json:
  lock /tmp/TestConcurrencyStress2606360510/001/backlog.lock: kanban board lock held   [1728136c7]
backlog_concurrency_test.go:56: 3/48 adds failed under contention; ... [1e5199b88]
backlog_concurrency_test.go:56: 1/48 adds failed under contention; ... [d8a1a8e4e]
backlog_concurrency_test.go:56: 1/48 adds failed under contention; ... [52c3fe590]
```

불변식 위반(id 충돌·lost update·last_seq) 문구는 어느 head 에도 없다 — 실패 지점은
`backlog_concurrency_test.go:56` 한 곳, 락 획득 실패뿐이다.

## 로컬 대조군 — darwin, 부하 생성 없음, 패키지 한정 직렬 1회

```
$ go test -json -race -count=1 ./internal/kanban/
rc=0
TestConcurrencyStress  pass  0.84
package                pass  23.495
```

## 러너 부하 상관 — 총 소요로는 안 갈린다

```
$ gh api .../jobs --jq '.jobs[]|select(.name=="Race Test")|[.started_at,.completed_at]|@tsv'
52c3fe590 14:16:18→14:25:31 (9m13s)   1728136c7 14:34:33→14:43:03 (8m30s)
d8a1a8e4e 15:09:58→15:19:15 (9m17s)   1e5199b88 15:43:39→15:53:02 (9m23s)
```

잡 총 소요 8m30s~9m23s, kanban 패키지 elapsed 15.0~19.6s — 넷이 비슷한데
실패 건수는 1·1·3·7 로 갈린다. **거친 러너 부하는 판별 인자가 아니다.**
갈리는 축은 `-race` 유무(위 표)이고, 그 안에서의 편차는 불공정 락의 꼬리다.

## 코드 — 수리가 실제로 바꾼 것 (`a680ea6e8`, `origin/develop` 현재값)

`internal/kanban/board_store.go`:

```go
boardLockSupportedWriters = 10
boardLockCIMutationCost   = 33 * time.Millisecond
boardLockHeadroom         = 5
boardLockWaitBudget = boardLockSupportedWriters * boardLockCIMutationCost * boardLockHeadroom  // 1.65s
boardLockWaitMin  =  5 * time.Millisecond
boardLockWaitMax  = 50 * time.Millisecond
boardLockWaitStep = 10 * time.Millisecond
```

`boardLockRetryWait(attempt)`: `ceil = 5ms + (attempt+1)*10ms`, `50ms` 상한,
반환 `5ms + rand.Int64N(ceil-5ms+1)`.

per-attempt 대기의 기댓값: attempt 0→10ms, 1→15, 2→20, 3→25, 4 이상→**27.5ms**(상한 50ms).

폐기된 값: `boardLockRetryDelay=25ms`, `boardLockRetries=40` → 41 sleep ≈ 1.025s.

## 산술 — 예산이 왜 모자라는가

테스트는 `writers=8 × addsPerWriter=6 = 48` 회 변이를 **하나의 flock 으로 직렬화**한다
(`backlog_concurrency_test.go`). 한 writer 가 최악으로 기다려야 하는 시간은
남은 직렬 작업량에 비례하므로 상한은 `48 × per-mutation cost` 에 가깝다.

관측에서 per-mutation cost 를 역산한다 (실패 writer 는 예산 1.65s 를 소진하고,
성공 변이들은 그 창과 겹쳐 진행되므로 `성공 변이 수 ÷ 테스트 elapsed`):

| head | elapsed | 성공 변이 | per-mutation cost (역산) |
|---|---|---|---|
| `d8a1a8e4e` | 1.97s | 47 | ≈ 42ms |
| `52c3fe590` | 2.15s | 47 | ≈ 46ms |
| `1e5199b88` | 2.88s | 45 | ≈ 64ms |
| `1728136c7` | 4.30s | 41 | ≈ 105ms |

대조: CI 비-race 0.52~0.76s ÷ 48 → **11~16ms**. 로컬 darwin -race 0.84s ÷ 48 → **17.5ms**.

예산이 버티려면 `48 × cost < 1.65s`, 즉 **cost < 34.4ms** 여야 한다.
sizing 입력값 `33ms` 가 바로 그 임계선이다 — 여유가 사실상 0이다.
그런데 CI `-race` 실측 cost 는 42~105ms 로 **임계선을 1.2~3.1배 넘는다**.

`headroom = 5` 가 5배 여유로 읽히지만 그렇지 않다. 곱해지는 항이
**직렬화되는 변이 수(48)** 가 아니라 **지원 레인 수(10)** 다.
`10 × 5 = 50 ≈ 48` 이라 우연히 "48회를 33ms 로 버티는 예산"과 같아졌을 뿐,
헤드룸은 이 환산에 이미 다 쓰였고 남는 마진이 없다.

그리고 `boardLockWaitBudget` 은 **컴파일 시점 상수**다. 요구 대기는 실행 머신 속도에
비례해 늘어나는데 예산은 늘지 않는다 — 한 번 관측한 33ms 머신에만 맞는 값이다.

## 수리가 자기 진단을 닫지 못한 지점

`a680ea6e8` 이 스스로 적은 기제는 "재시도 지연(25ms) < 변이 비용(33ms) 이라
깨어난 경합자가 거의 항상 잠긴 락을 본다" 였다. 수리 후 per-attempt 대기 기댓값은
**27.5ms**(상한 50ms) — 실측 변이 비용 42~105ms 보다 **여전히 작다**.
지터는 위상 고정(lockstep)을 깨지만 "깨어나면 잠겨 있다"는 성질 자체는 그대로다.
그래서 꼬리가 얇아졌을 뿐 사라지지 않았다 — 실패 건수가 1~7 로 흔들리고,
수리 이후에도 초록 run 이 존재한다(`51daada00`, run 33247558848, Race Test `success`).

## Gaps — 이번 실행에서 관측하지 않은 것

- ~~`d7010f86a` 미측정~~ → 아래 census 에서 로그로 메웠다:
  `gh run view 33314551261 --log-failed` (87,533 bytes) 에서
  `--- FAIL:` 1건(`TestConcurrencyStress`), `8/48 adds failed under contention`,
  `grep -c "DATA RACE"` → 0, `run_attempt` → 1
- 초록 race run `51daada00` 의 `TestConcurrencyStress` elapsed **미측정**
  (아티팩트 조회 실패 — `no artifact matches`). 초록일 때의 per-mutation cost 를 못 쟀다
- per-mutation cost 는 **역산**이다. 스트림에 변이 단위 타임스탬프가 없어 직접 측정이 아니다
- ubuntu-latest 러너의 vCPU 수 / `GOMAXPROCS` 를 로그에서 확인하지 않았다.
  "코어가 적어 `-race` 오버헤드가 커진다"는 설명은 **가설**이며 이번 실행의 측정이 아니다
- ~~발화율 미측정~~ → 아래 census 에서 잰다 (14 run 중 12 붉음)
- `-race` 잡의 실패 테스트 전수를, 스트림이 없는 9개 run 에 대해서는 `--log-failed` 의
  `--- FAIL:` 로만 셌다. `--log-failed` 는 실패 스텝만 담으므로 다른 스텝에서 난 실패는 안 보인다
- 재현 조건(질문 4)을 로컬에서 **재현하지 않았다**. 부하 생성 금지 지시에 따른 의도적 미측정

## Residual-risk

- 역산한 cost 는 실패 writer 의 대기와 성공 변이가 겹치는 정도에 따라 ±가 있다.
  결론(cost 가 34.4ms 임계를 넘는다)은 네 head 모두에서 같은 방향이지만, 개별 수치는 근사다
- `52c3fe590` 의 두 번째 실패를 t352 계열로 본 것은 종전 귀속에 기댄 것이고
  이번 실행에서 그 카드를 열어 확인하지 않았다
- 초록 run 이 존재하므로, 어떤 수리든 "한 번 초록"으로는 닫혔다고 말할 수 없다.
  판정에는 반복 관측이 필요하다

## 수리 이후 전수 census — 발화율

대상 선정: `gh run list --workflow=ci.yml --branch develop --limit 40` 에서
`cancelled` 를 뺀 run 중, head 가 수리 병합의 자손인 것.
자손 판정은 `git rev-list 728f91006..origin/develop` (230 커밋) 목록과의 대조 —
후보 9개 전부 이 목록에 있다(`grep -x -c` → 9). 여기에 이미 측정한 5개를 더해 **14 run**.

`Race Test` 잡 결론은 `gh api .../jobs --jq '.jobs[]|select(.name=="Race Test")|[.conclusion]'`.
실패 테스트는 스트림이 있으면 스트림, 없으면 `gh run view <id> --log-failed` 를 파일로 받아
`grep -o -- "--- FAIL: [A-Za-z0-9_]*" | sort -u`.
(스트림 아티팩트는 `52c3fe590`(08-30 14:15) 이후 run 에만 있다 — 관측 배선이 그때 들어왔다.)

| head | run | Race Test | `TestConcurrencyStress` | 같은 잡의 다른 실패 테스트 | `DATA RACE` |
|---|---|---|---|---|---|
| `c6aa61346` | 33173944485 | failure | **통과** | `TestGitDiffNameCount_Predicate` | 0 |
| `48d8ef4be` | 33175578950 | failure | FAIL | `TestBinaryLag_OneSeamServesBothSurfaces` | 0 |
| `1a635aea8` | 33188568464 | failure | FAIL | `TestGitDiffNameCount_Predicate` | 0 |
| `91ec14dbe` | 33219939019 | failure | FAIL | `TestBinaryLag_OneSeamServesBothSurfaces` | 0 |
| `a6bbbf82b` | 33244792917 | failure | FAIL | `TestBinaryLag_OneSeamServesBothSurfaces` | 0 |
| `51daada00` | 33247558848 | **success** | **통과** | — | — |
| `15453140a` | 33248639479 | failure | FAIL | 없음 | 0 |
| `ee50984ab` | 33250831813 | failure | FAIL | 없음 | 0 |
| `68ecbfe4a` | 33308726563 | failure | FAIL | 없음 | 0 |
| `d7010f86a` | 33314551261 | failure | FAIL **8/48** | 없음 | 0 |
| `52c3fe590` | 33316389199 | failure | FAIL 1/48 2.15s | `TestGitDiffNameCount_Predicate` | 0 |
| `1728136c7` | 33317217710 | failure | FAIL 7/48 4.30s | 없음 | 0 |
| `d8a1a8e4e` | 33318836582 | failure | FAIL 1/48 1.97s | 없음 | 0 |
| `1e5199b88` | 33320417834 | failure | FAIL 3/48 2.88s | 없음 | 0 |

**발화율: 14 run 중 12 붉음, 2 통과** (`51daada00`, `c6aa61346`).
`c6aa61346` 의 잡 붉음은 이 테스트가 아니라 `TestGitDiffNameCount_Predicate` 때문이다 —
잡 결론만 세면 13/14 이지만, **이 테스트 기준으로는 12/14** 다.

`run_attempt` 는 다섯 head 전부 **1** 이다
(`gh api .../runs/<id> --jq '.run_attempt'` → 1, 1, 1, 1, 1).
숨은 attempt-1 실패 없음.

`DATA RACE` 는 스트림·로그 통틀어 **전 run 0건**.

## 이웃 카드와의 관계 (관계만 — 흡수 판정 아님)

같은 `Race Test` 잡에서 함께 붉은 테스트가 둘 더 있다. 별개 계열로 보이며 이 카드에 귀속하지 않는다:

- `TestGitDiffNameCount_Predicate` (`internal/graph`) — `c6aa61346`·`1a635aea8`·`52c3fe590`. 종전 t352 소관
- `TestBinaryLag_OneSeamServesBothSurfaces` — `48d8ef4be`·`91ec14dbe`·`a6bbbf82b`. t326/t366 계열로 보임

t278(`TestConfigChange_RT005ReloadIntegration`)은 이번 14 run 어디에서도 실패로 나타나지 않았다.
다만 t278 이 적은 양상 — "로컬 darwin 초록 + CI ubuntu 붉음 + 수 초 내 실패" — 은
이 카드가 실측한 축(`-race` 유무로 갈리고, per-attempt 대기가 변이 비용보다 짧다)과 같은 모양이다.
공통 인자 조사(t278 조치 (1))가 이 카드의 측정을 재료로 쓸 수 있다.

## 병렬 판독 보고 도착 (지연) — 대조 및 정정

`Agent(Explore)` 6건의 보고가 위 측정을 마친 뒤 도착했다. 독립 판독이므로 대조했다.

### 일치 — 재측정으로 확인된 항목

- `run_attempt = 1`: 다섯 run 전부 (양쪽 독립 확인)
- `--- FAIL:` 개수: `52c3fe590` 만 **2**, 나머지 넷은 **1** (양쪽 일치)
- `DATA RACE` 0: 양쪽 전 run 일치
- `N/48` 문구: 8/48 · 1/48 · 7/48 · 1/48 · 3/48 (양쪽 일치)
- `Race Test` 가 유일 실패 잡인 run 들 (양쪽 일치)

### 보고가 메운 것 — 제가 못 쟀던 값

- `d7010f86a` 의 elapsed: **4.03s** (스트림이 없어 제 측정에는 없던 값)
- 스트림 없는 9 run 의 elapsed + `N/48`:
  `68ecbfe4a` 2.37s 2/48 · `ee50984ab` 2.32s 1/48 · `15453140a` 2.16s 1/48 ·
  `a6bbbf82b` 2.18s 1/48 · `91ec14dbe` 1.94s 1/48 · `1a635aea8` 2.41s 2/48 ·
  `48d8ef4be` 2.16s 1/48 · `c6aa61346` 해당 테스트 실패 없음(kanban 패키지 `ok`)

**결과: 실패한 14 run 중 12건의 elapsed 가 전부 1.94s ~ 4.30s.**
예산 1.65s 를 **하나도 빠짐없이** 넘는다. "실패 writer 가 예산을 소진하고 죽는다"는
산술이 4건이 아니라 12건 전부에서 성립한다.

### 정정 — 제 census 표의 열 하나가 부정확했다

제 census 표의 "같은 잡의 다른 실패 테스트" 열은 `--log-failed` 출력을 잡 구분 없이
읽어 만든 것이라, **다른 잡의 실패를 같은 잡으로 적었다.** 정확한 귀속:

| head | `Race Test` 잡의 실패 | `Test (ubuntu)` 잡의 실패 |
|---|---|---|
| `48d8ef4be` | `TestConcurrencyStress` 2.16s + `TestBinaryLag_OneSeamServesBothSurfaces` | `TestBinaryLag_OneSeamServesBothSurfaces` |
| `91ec14dbe` | `TestConcurrencyStress` 1.94s | `TestBinaryLag_OneSeamServesBothSurfaces` |
| `a6bbbf82b` | `TestConcurrencyStress` 2.18s | `TestBinaryLag_OneSeamServesBothSurfaces` |
| `1a635aea8` | `TestConcurrencyStress` 2.41s + `TestGitDiffNameCount_Predicate` 0.23s | 없음 |
| `c6aa61346` | `TestGitDiffNameCount_Predicate` 0.14s **만** | 없음 |
| `52c3fe590` | `TestConcurrencyStress` 2.15s + `TestGitDiffNameCount_Predicate` 0.19s | 없음 |

`TestBinaryLag_OneSeamServesBothSurfaces` 는 세 run 에서 **비-race 잡에서도** 붉다 —
`-race` 축과 무관한 별개 결함이라는 뜻이고, "실패는 `-race` 잡 전용"이라는 제 진술은
**`TestConcurrencyStress` 에 한정해서만** 참이다. 이 한정은 위 표 1에서 이미 그렇게
측정했으므로(다섯 head 의 비-race 잡에서 해당 테스트는 전부 pass) 결론은 바뀌지 않는다.

### 보고가 더한 진단 하나

`52c3fe590` 의 `TestGitDiffNameCount_Predicate` 실패 원문:

```
testing.go:1464: TempDir RemoveAll cleanup:
  unlinkat /tmp/TestGitDiffNameCount_Predicate1231336354/001/.git/objects: directory not empty
```

테스트 본문의 단언이 아니라 `t.TempDir()` **정리** 실패다. t352 의 브랜치명이
`WT-tempdir-cleanup-race` 인 것과 맞는다 — t352 귀속의 방증이지만,
이번 실행에서 그 카드를 열어 확인한 것은 아니다(Residual-risk 유지).

### 반환 채널 — 앞선 진술 정정

앞서 리드에 "에이전트 보고가 반환되지 않았다"고 적었다. **틀렸다** — 지연됐을 뿐
6건 전부 도착했다. 정확히 말하면: 에이전트가 `idle` 로 보인 시점과 보고가 이 세션에
도착한 시점 사이에 수 분의 간격이 있었고, 그 사이 재요청 메시지에도 응답이 없었다.
`idle` 을 완료 신호로 읽은 것이 오판이다.

다만 그 오판이 만든 중복은 낭비였을 뿐 오염은 아니다. 위 표 1~5의 수치는 전부
제 측정이고, 이 절에서 보고 인용임을 명시한 값(9 run 의 elapsed·`N/48`,
`d7010f86a` 4.03s, 잡 귀속 정정)만 보고 출처다.
