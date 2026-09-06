# t477 — binary-lag 가드 허용목록 누락 수리

- Card: **t477** (Class A · Tier S)
- Branch: `WT-binarylag-allowlist`
- Worktree: `.claude/worktrees/t477`
- Base: local `develop` @ `d7116400f` (ff-only, 카드 커밋 직전 기준)

## Claim

`internal/cli/binary_lag_test.go`의 `namesAddedAfterBaseline` 허용목록에
t466이 추가한 doctor 체크 이름 `"Hook Delivery"`가 등재되지 않아
`TestBinaryLag_DoctorCheckNameSetIsUnchanged`가 실패했다. 허용목록에
백틱 형태 엔트리 1줄을 추가해 수리했다.

## 핵심 발견 — 허용목록 키는 소스 텍스트다 (따옴표 포함)

`checkNamesFromSource`는 `exprSource`(`binary_lag_test.go:157`)로 registry
엔트리의 **이름 표현식 소스 텍스트**를 그대로 집합 원소로 삼는다. 따라서
두 모양이 공존한다:

| 등록 형태 | doctor.go 위치 | 집합 원소 | 허용목록 키 |
|---|---|---|---|
| 상수 식별자 | `doctor.go:224` `{hookWiringCheckName, ...}` | `hookWiringCheckName` | `"hookWiringCheckName"` |
| 문자열 리터럴 | `doctor.go:230` `{"Hook Delivery", ...}` | `"Hook Delivery"` (따옴표 포함) | `` `"Hook Delivery"` `` |

따옴표를 뺀 `"Hook Delivery"` 로 등재하면 **매치되지 않아 계속 RED**다.
이 함정은 뮤턴트 2로 실증했다(아래).

## Evidence

### 판정식 — 종료코드가 아니라 매치 수

```
$ grep -c '"Hook Delivery"' internal/cli/doctor.go
1
$ grep -c '^\t`"Hook Delivery"`:' internal/cli/binary_lag_test.go
1
$ sed -n '/^var namesAddedAfterBaseline/,/^}/p' internal/cli/binary_lag_test.go | grep -c ': *true,'
2
```

### RED (수리 전)

```
$ go test ./internal/cli/ -run TestBinaryLag_DoctorCheckNameSetIsUnchanged -count=1 -v
=== RUN   TestBinaryLag_DoctorCheckNameSetIsUnchanged
    binary_lag_test.go:205: this SPEC added doctor check name "Hook Delivery"; REQ-BLV-009 rewires the existing "Binary Freshness" item and registers no new name
--- FAIL: TestBinaryLag_DoctorCheckNameSetIsUnchanged (0.05s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.930s
```

실패 메시지가 이름을 `"Hook Delivery"` — 따옴표를 달고 — 출력한다.
집합 원소가 소스 텍스트라는 직접 증거다.

### GREEN (수리 후)

```
$ go test ./internal/cli/ -run TestBinaryLag_DoctorCheckNameSetIsUnchanged -count=1 -v
=== RUN   TestBinaryLag_DoctorCheckNameSetIsUnchanged
--- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged (0.05s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.918s
```

### 뮤턴트 검증 — 공허한 초록 아님 (2/2 RED)

**뮤턴트 1** — 추가한 줄을 통째로 제거 (`grep -c` = 0):

```
    binary_lag_test.go:215: this SPEC added doctor check name "Hook Delivery"; ...
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.934s
```

**뮤턴트 2** — 따옴표를 뺀 `"Hook Delivery": true` 로 등재:

```
    binary_lag_test.go:216: this SPEC added doctor check name "Hook Delivery"; ...
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.911s
```

뮤턴트 2가 RED라는 것이 백틱(따옴표 포함) 형태가 실제로 load-bearing임을 보인다.
두 뮤턴트 모두 원본 복원 후 GREEN 재확인.

### 패키지 범위 재측정

```
$ go vet ./internal/cli/...                          → exit 0, 출력 없음
$ gofmt -l internal/cli/binary_lag_test.go | wc -l   → 0
$ go test ./internal/cli/... -count=1 -timeout 900s  → exit 0
   ok 17 / FAIL 0
```

### 병합 트리 재측정 (develop 흡수 후)

카드 커밋 `b2f98f6aa` 이후 `develop`이 5커밋 앞서가 흡수했다.
병합 커밋 `3a09587b8` 트리에서 재측정:

```
$ go test ./internal/cli/... -count=1 -timeout 900s   → exit 0
   ok 17 / FAIL 0
```

이 회차의 최종 측정 트리는 `3a09587b8`이다. 창을 받을 시점에 `develop`이
더 움직였으면 재흡수 후 다시 잰다.

## Baseline-attribution

- 트리: `.claude/worktrees/t477`, 카드 커밋 직전 `d7116400f` (로컬 develop과 `git rev-list --count --left-right develop...HEAD` = `0 0`)
- baseline 블롭 `22f90b1c7:internal/cli/doctor.go` → `grep -c 'Hook Delivery'` = **0**
  (즉 `"Hook Delivery"`는 baseline 이후 추가가 맞고, 허용목록이 옳은 수리 지점이다.
  `lagBaselineSHA` 범프는 다른 드리프트까지 침묵시키므로 틀린 수리다 — 원 주석이 명시)
- 모든 수치는 이 트리에서 이 회차에 측정

## Gaps (관측하지 않은 것)

- **CI 판정 없음.** 로컬 darwin/arm64 단일 환경 측정이며, windows·linux 매트릭스는
  돌리지 않았다. push 금지 지시에 따라 CI 실행을 요청하지 않았다 — 판정은 리드의
  develop push가 일으키는 실행 몫이다.
- `./internal/cli/...` 밖 패키지는 재측정하지 않았다 (변경이 그 패키지 1파일에 국한).
- lane-10이 관측한 "이 레드가 사는 동안 `-cover` 출력이 억제된다"는 부수효과는
  **재현하지 않았다.** 카드 범위 밖이며, 수리로 해소됐는지는 미확인이다.
- `moai doctor` 런타임 동작은 관측하지 않았다 (테스트 파일 1개만 변경).

## Residual-risk

- 같은 함정이 재발할 수 있다: 앞으로 doctor 체크를 **문자열 리터럴로** 추가하는
  SPEC은 허용목록에 백틱 형태로 등재해야 하는데, 이 규칙은 이제 코드 주석에만
  적혀 있다. 기계적 가드는 없다.
- 허용목록은 baseline 고정 방식의 구조적 부채를 그대로 안고 있다. 이후 SPEC마다
  1줄씩 늘어난다 — 이번 수리는 그 설계를 바꾸지 않았다(범위 밖).

---

## 병합 트리 재측정 (창 2회차 — lane-14)

리드 호명 이후 `develop`이 `d7116400f` → `7896f30d7`(t237·t474 운반)로 움직여,
그 tip을 흡수한 뒤 **병합할 트리에서 다시** 측정했다.

- 흡수 대상: local `develop` @ `7896f30d7`
- 흡수 병합 커밋: `1987a51da` (`Merge branch 'develop' into WT-binarylag-allowlist (t477)`) — 충돌 0
- 측정 트리: `7f6072bbeb23c464e277c30e85292a0dc7ffef24` (HEAD `1987a51da`, 워킹트리 clean)

### 판별식 재측정 — 종료코드가 아니라 매치 수

```
$ grep -c '"Hook Delivery"' internal/cli/doctor.go
1
$ grep -c '^\t`"Hook Delivery"`:' internal/cli/binary_lag_test.go
1
$ sed -n '/^var namesAddedAfterBaseline/,/^}/p' internal/cli/binary_lag_test.go | grep -c ': *true,'
2
```

수치 3개 모두 창 1회차와 동일하다.

### 패키지 전수 — t237이 같은 패키지에 들어온 뒤에도 초록

```
$ go test ./internal/cli/... -count=1 -timeout 900s
... (전문: .moai/state/verify/t477/merged-cli.txt)
exit=0
$ grep -c '^ok ' merged-cli.txt   → 17
$ grep -c '^FAIL' merged-cli.txt  → 0
$ grep -c 'no test files' ...     → 0
```

t237(`hook_install_precommit.go`)과 t477(`binary_lag_test.go`)은 같은 패키지지만
파일이 겹치지 않아 충돌이 없었고, 합류 후에도 17/17이 유지된다.

### 귀속 — 측정 트리와 병합 트리의 델타

이 부록 커밋은 재측정 **뒤에** 얹혔다. 따라서 측정 트리(`7f6072bbe`)와
병합될 트리의 델타는 이 파일 하나(`.moai/reports/t477/verdict.md`)뿐이며
**Go 파일 델타는 0**이다 — 아래 `git diff --stat` 으로 증명한다.

### Gaps (변동 없음)

- CI 판정 없음 — windows/linux 매트릭스는 리드 push 몫. darwin 단일 환경 관측은 예상이지 판정이 아니다.
- `-cover` 억제 미재현 — 카드 범위 밖(별도 카드감).
