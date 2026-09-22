# t213 — `go test` TMPDIR 누수 (Class B, 재현 우선)

카드: t213 · 브랜치 `WT-gotest-tmp-leak` · base `origin/main f7d4b7824`

## 1. Claim

`$TMPDIR`에 쌓이던 세 갈래 누수 중 **두 갈래를 뿌리에서 닫았다.** 원인은 서로 다르고,
둘 다 리드가 제시한 (a)/(b)/(c) 중 **(c) 하위 프로세스** 계열이다 — 다만 기전이 서로 다르다.

| prefix | 개수(착수 시점) | 소유 패키지 | 근본 원인 | 이번 PR |
|---|---|---|---|---|
| `TestGateCmd_RunE_Behavior*` | 9,040 | `internal/hook/quality` | 게이트 스텝이 **자기 수트를 재귀 실행** | 닫음 |
| `go-build*` | 9,171 | 위와 동일 | 죽은 재귀 `go test` 1건당 1개 | 닫음(동일 뿌리) |
| `moai-cli-profiles-*` | 11,962 | `internal/cli` | 헬퍼 서브프로세스가 `os.Exit`로 TestMain 정리 우회 | 닫음 |

기존 41만 건 청소는 범위 밖(지시대로). 쌓이던 것을 멈추는 것만 했다.

## 2. Evidence

### 2-A. `TestGateCmd` — 게이트가 자기 수트를 재귀 실행했다

가설 검증은 **실행 중 프로세스 관측**으로 했다. 수정 전, 대상 테스트를 1회 돌리는 동안:

```
$ ps -o pid,etime,command -ax | grep cli.test
62542  01:49  …/b001/cli.test -test.paniconexit0 -test.run=TestGateCmd_RunE_Behavior …   ← 내가 띄운 것
65042  01:35  …/b001/cli.test -test.testlogfile=… -test.paniconexit0 -test.timeout=10m0s  ← 자식: -run 필터 없음 = 수트 전체
77729  00:01  (cli.test)                                                                  ← 손자
```

자식의 `-test.timeout=10m0s`는 `runGate`의 `context.WithTimeout(10*time.Minute)`와 일치한다.
`-test.run` 필터가 없다 = **`internal/cli` 수트 전체가 다시 돈다.** 손자까지 나온 시점에서 지수 증식.

기전: `internal/hook/quality`의 `runStep`이 `exec.CommandContext`를 만들면서 **`cmd.Dir`를 지정하지
않았다.** 모든 스텝의 인자가 cwd 상대(`vet ./...`, `run`, `test ./...`, 비-Go 15종도 동일)이므로,
자식은 **호출 프로세스의 cwd**를 물려받는다. 프로덕션에선 `moai` 바이너리의 cwd가 대개 프로젝트
루트라 잠복했다. `go test` 아래에서는 cwd가 시험 대상 패키지 디렉터리라, temp fixture를 가리킨
게이트가 **실제 저장소를 검사**했고, 그 스텝이 `go test ./...`일 때 자기를 호출한 수트를 다시 돌렸다.

누수는 그 결과다: 재귀 테스트 바이너리가 죽으면 `t.TempDir()` cleanup이 실행되지 않는다.
**정리 코드가 없어서가 아니라, 정리에 도달하지 못해서다.**

날짜 히스토그램이 이를 뒷받침한다 — 두 prefix가 거의 1:1로 겹친다(재귀 `go test` 1건 = go-build 임시
작업 디렉터리 1개; 정상 종료 시 제거된다. 툴체인 캐시는 `~/Library/Caches/go-build`에 따로 있다):

| 날짜 | `TestGateCmd*` | `go-build*` |
|---|---|---|
| 08-16 | 397 | 405 |
| 08-17 | 5,493 | 5,529 |
| 08-18 | 2,745 | 2,740 |
| 08-19 | 152 | 169 |
| 08-20 | 250 | 257 |

→ **`go-build*`는 툴체인 소관이 아니라 우리 소관이었다.** 리드 배차의 "우리 것인지 먼저 확인" 지시에
대한 답: 우리 것이다.

**수정**: `runStep`이 `resolveQualityProjectDir`(같은 패키지가 이미 파일 존재 검사에 쓰던 SSOT)로
`cmd.Dir`를 지정한다. 선행 SPEC(HOOK-CWD-LEAK-AUDIT)이 cwd 해석은 고쳤으나 **서브프로세스 실행부를
놓쳤다.**

**회귀 테스트** `TestRunStep_RunsInConfiguredProjectDir` — 문자열 grep이 아니라 실제 실행:
fixture에 `go vet`이 반드시 잡는 printf 불일치 파일(`gate_cwd_probe.go`, 이름이 저장소에 유일)을
심고, 스텝이 그 파일을 진단하는지 본다. cfg.ProjectDir를 지키면 실패하며 파일명을 뱉고, 다른 데서
돌면 그 파일을 아예 못 본다.

```
수정 전:  --- FAIL … runStep passed on a fixture whose only package fails `go vet`
수정 후:  --- PASS: TestRunStep_RunsInConfiguredProjectDir (0.51s)
```

**낙진 수정 1건**: `cmd.Dir`가 붙자 typecheck 테스트 5건이 깨졌다. fixture가 `go.mod`만 있고 `.go`
파일이 없어 `go vet ./...`가 `matched no packages`로 non-zero 종료했기 때문이다. 이 5건은 **버그
덕분에 통과하고 있었다**(실제 저장소를 검사해 초록). Go 스텝에 `sourceExts: [".go"]`를 붙여, 언어만
선언하고 코드는 아직 없는 스캐폴드가 첫 게이트에서 막히지 않게 했다 — Python 스텝이 이미 갖고 있던
정책과 동일하다.

### 2-B. `moai-cli-profiles-*` — 헬퍼가 `os.Exit`로 정리를 우회했다

리드가 전달한 t208 레인의 규명을 **재규명하지 않고 검증만** 했다. `internal/cli`의 5개 지점이
`os.Args[0]`을 재실행하고(`exitcode_guard_test.go:31`, `todo_test.go:150/333/662`,
`launch_session_pid_exec_posix_test.go:55`), 각 자식이 TestMain을 돌려 자기 샌드박스를 만든 뒤,
헬퍼 본문이 `os.Exit`으로 끝난다 → `main_test.go`의 `restoreProfileBaseDir()`에 도달하지 못한다.

**수정**: 자식이 자기 샌드박스를 만들지 않고 **부모 것을 물려받는다**(`MOAI_CLI_TEST_PROFILE_BASE`).
5개 지점 전부 `append(os.Environ(), …)`로 자식 env를 만들므로 **호출 지점 수정이 0건**이고, 앞으로
추가될 헬퍼도 자동으로 덮인다. 디렉터리 소유권은 그것을 만든 프로세스에 남는다.

**회귀 테스트** `TestProfileSandbox_HelperSubprocessLeavesNoBaseDir` — 자식에게 **전용 TMPDIR**을
주고 그 안만 검사한다. 전역 카운트를 쓰지 않으므로 다른 레인의 동시 실행이 결과를 흔들 수 없다.

```
수정 전:  --- FAIL … helper subprocess left 1 profile sandbox dir(s) behind in its TMPDIR
수정 후:  --- PASS: TestProfileSandbox_HelperSubprocessLeavesNoBaseDir (0.06s)
```

> 함정 회피: Go의 `os.MkdirTemp`는 `$TMPDIR`을 지킨다. 리드가 경고한 "BSD `mktemp`은 `$TMPDIR`을
> 무시한다"는 `mktemp` **바이너리** 이야기이며, 이 격리는 Go 런타임 경로라 무효화되지 않는다.

### 2-C. 실제 실행 기반 누수 단언 (전역 baseline)

```
# TestGateCmd 계열
before = 9,041
$ go test ./internal/cli/ -run TestGateCmd_RunE_Behavior -count=1     → ok, 1.53s (수정 전: 120s 타임아웃)
after  = 9,041                                                        → 신규 0

# moai-cli-profiles 계열 (t208 레인이 baseline 보존용으로 남겨둔 11,962 + 내 RED 실행 3)
before = 11,965
$ go test ./internal/cli/ -run 'Todo|ExitCod|ExecHelper|LaunchSessionPID' -count=1  → ok, 19.5s
after  = 11,965                                                       → 신규 0
```

부수 효과: `internal/hook/quality` 수트가 **22.1s → 6.9s**. fixture를 가리킨 게이트가 더 이상 실제
저장소를 훑지 않는다.

## 3. Baseline-attribution

- 트리: `WT-gotest-tmp-leak`, base `origin/main f7d4b7824`(워크트리 생성 시 `cd0cee1b8`).
- `$TMPDIR` = `/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T` (`printenv TMPDIR`).
- 계수는 전부 `timeout 300 find … -maxdepth 1 -name '<prefix>*'` + `find_rc=0` 확인.
  `timeout 60`은 41만 엔트리 스캔 중에 죽어 `wc -l`이 **조용히 0을 보고**한다 — 한 번 겪고 상향했다.
- 통과 기록: `go test ./internal/hook/quality/ -count=1` → `ok … 6.934s`.
  `go vet ./internal/cli/... ./internal/hook/quality/...` → 무출력.
  `gofmt -l` → 변경 4파일 모두 미출력(목록에 뜬 31개는 기존 미포맷 파일, 이번 변경과 무관).

## 4. Gaps (미검증)

- **`internal/cli` 전체 수트를 로컬에서 돌리지 않았다.** 착수 시점 load 29.10 — 여기서의 측정은
  코드가 아니라 머신을 재는 것이고, CLAUDE.local.md §4가 금지한다. 전 패키지 판정은 CI 몫.
- **darwin/windows 매트릭스 미검증.** `cmd.Dir` 지정은 플랫폼 중립이지만 실측은 darwin 1종뿐.
- **비-Go 15개 툴체인의 스텝은 실행하지 않았다.** `cmd.Dir` 수정은 모든 스텝에 동일하게 적용되나,
  실제 검증은 Go 스텝(vet)으로만 했다. `sourceExts` 추가는 Go 스텝에만 한정했다.
- **기존 41만 엔트리는 그대로다.** 지시대로 청소하지 않았다.

## 5. Residual-risk

- **`cmd.Dir` 지정으로 동작이 바뀌는 곳이 더 있을 수 있다.** 지금까지 cwd == 프로젝트 루트를
  암묵 전제하던 호출자가 있다면, 이제 `cfg.ProjectDir`(또는 `CLAUDE_PROJECT_DIR`)이 실제로
  구속한다. `ProjectDir`이 비면 해석 결과가 종전과 같으므로(cwd 폴백) 영향은 `ProjectDir`을
  설정한 경로에 한정된다. typecheck 테스트 5건이 바로 그 부류였고 — **버그에 기대 통과하던
  테스트가 더 남아 있을 수 있다.** CI가 판정한다.
- **`sourceExts` 스킵의 방향성**: `.go` 파일이 하나도 없는 모듈에서 vet/test가 스킵된다. 코드가
  없으니 검사할 것도 없다는 판단이지만, "검사가 돌지 않았다"를 "통과했다"로 읽히게 하는 형태다.
  기존 스킵과 마찬가지로 `slog.Warn`으로 보고된다.
- **`MOAI_CLI_TEST_PROFILE_BASE` 상속 범위**: 테스트 프로세스가 `os.Setenv`로 심으므로, 그
  프로세스가 띄우는 **모든** 자식이 값을 물려받는다. 테스트 전용 키이고 프로덕션 코드는 읽지
  않지만, 자식이 임의 프로그램일 경우 env에 낯선 키가 하나 늘어난다.
- **`moai-cli-profiles-*` 신규 누수 0**은 내가 돌린 테스트 집합에 한해 참이다. 아직 관측하지
  못한 다른 재실행 지점이 `os.Environ()`을 쓰지 않는다면 그 지점은 여전히 샌드박스를 새로 만든다
  (현재 5개 지점은 전부 `os.Environ()` 사용을 확인했다).
