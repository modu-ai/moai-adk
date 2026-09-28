# t1256 homestate 커버리지 기준 재측정

## Claim

원래 merge-base `b59a5d69c`의 `internal/homestate` 커버리지 수치는 69.0%로 다시 출력됐다. 현재 로컬 `develop` 병합 커밋 `543d102f5`에서는 같은 패키지 테스트가 통과하고 커버리지가 85.0%로 출력됐다. 기준 측정의 테스트가 실패했으므로 이 숫자만으로 동일 조건의 녹색 비교라고 주장하지 않는다.

## Evidence

현재 `develop`에서 직접 실행:

```text
$ go test -count=1 -cover ./internal/homestate
ok  github.com/modu-ai/moai-adk/internal/homestate  44.390s  coverage: 85.0% of statements
```

Python `tempfile.TemporaryDirectory(dir="/tmp")`에 `git archive b59a5d69c`를 추출해 그 디렉터리에서 직접 실행:

```text
$ go test -count=1 -cover ./internal/homestate
--- FAIL: TestTempDiscriminantParity (0.32s)
    --- FAIL: TestTempDiscriminantParity/non-temp_working_directory (0.00s)
        temp_parity_test.go:87: premise: kanban.TempOriginReason(.) = true, want false
FAIL
coverage: 69.0% of statements
FAIL github.com/modu-ai/moai-adk/internal/homestate 18.735s
```

임시 디렉터리는 Python 컨텍스트 종료 시 자동 정리됐다.

## Baseline-attribution

두 명령은 이번 작업에서 직접 실행했다. 기준 소스는 고정 커밋 `b59a5d69c`; 현재 소스는 로컬 `develop` `543d102f5`다. 기존 t1256 진행 기록의 M5 수치도 homestate 69.0%였으나, 그 기록은 이번 재측정의 대체 근거로 쓰지 않았다.

## Gaps

- 기준 테스트의 실패 원인은 기준 소스를 `/tmp`에 추출했을 때 `non-temp_working_directory` 전제가 성립하지 않기 때문이다. 실패한 테스트까지 포함해 커버리지 프로필은 계산됐지만, 성공한 기준 스위트의 수치는 관측하지 못했다.
- 현재 `develop`에는 t1256 이후 다른 카드의 변경도 포함된다. 현재 85.0%를 t1256 단독 변경의 효과로 귀속하지 않는다.
- 원격에 아직 병합 커밋을 push하지 않았으므로 이 새 트리의 CI 판정은 없다.

## Residual-risk

원래 merge-base를 `/tmp` 밖의 격리된 체크아웃에서 실행하면 커버리지 수치가 달라질 수 있다. 현재 트리에서 새로 발견된 기능 회귀는 없지만, 기준 실패를 해결한 동일 조건 비교는 별도 검증이 필요하다.
