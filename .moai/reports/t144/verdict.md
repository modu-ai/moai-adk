# t144 — active-sessions.json PID 결함 수정

카드: t144 (Class B — plan 생략, run → sync 인레인)
브랜치: `WT-t144` (base `4100d8767`)
근거 카드: `.moai/reports/glm-lane-concurrency-observations.md` §3.4 / §5-1

---

## 1. 주장 (Claim)

`.moai/state/active-sessions.json` 에 기록되는 `pid` 가 즉시 종료하는 훅 서브프로세스의 PID였다. 이제 장수 세션 프로세스(claude)의 PID를 기록한다. 등록 직후 `kill -0` 이 rc=0 을 돌려준다.

## 2. 증거 (Evidence)

### 2.1 결함의 위치 — 단일 지점

`internal/session/registry.go` `Registry.Register` 가 항목을 새로 append 할 때 `PID: os.Getpid()` 를 썼다. 이 함수의 유일한 프로덕션 호출 경로는 둘 다 단명 프로세스다:

- `internal/hook/session_start.go:1312` `runMultiSessionProtocol` → `reg.Register(...)` — `moai hook session-start` 서브프로세스
- `internal/cli/session.go:75` `RegisterSession(...)` — `moai session register` CLI 프로세스

즉 기록되는 PID 는 항상 밀리초 단위로 사라지는 프로세스의 것이었다.

### 2.2 workers.json 이 올바른 이유 (카드가 참조하라고 지시한 획득 경로)

`internal/kanban/factory_slots.go:109` `NewFactoryWorkerEntry` 도 똑같이 `os.Getpid()` 를 쓴다. 그런데 결과는 옳다 — 호출 지점(`internal/cli/factory.go:225`)이 런처 프로세스이고, 런처는 등록 직후 `syscall.Exec` 로 claude 로 **자기 자신을 대체**하기 때문이다(`internal/cli/launch_exec_posix.go`). 런처의 PID 가 그대로 세션의 PID 가 된다.

훅 서브프로세스에는 그런 행운이 없다. 자기가 태어난 조상 계보를 거슬러 올라가 찾는 수밖에 없다.

### 2.3 조상 계보의 실측 형태 (darwin)

임시 프로브로 실제 계보를 읽었다(`unix.SysctlKinfoProc("kern.proc.pid", …)`):

```
pid=26831 comm=".tmp-pidprobe" ppid=26821
pid=26821 comm="go"            ppid=26817
pid=26817 comm="zsh"           ppid=69111
pid=69111 comm="2.1.235"       ppid=68508
pid=68508 comm="zsh"           ppid=1530
```

관측 두 가지가 설계를 결정했다:

- **세션 프로세스의 이름은 `claude` 가 아닐 수 있다.** 위 계보에서 comm 은 `2.1.235` — 버전 이름 그대로다(설치 경로 basename). 그래서 목표 프로세스를 **이름으로 찾지 않는다**.
- 대신 **거쳐 지나갈 것들만 이름으로 건너뛴다.** 훅 래퍼는 셸 스크립트이고(`.claude/hooks/moai/handle-session-start.sh` 가 `exec moai hook session-start`), 런타임의 `sh -c` 와 래퍼 자신의 `exec` 가 접히는지에 따라 세션과 moai 사이에 셸이 0개일 수도 여러 개일 수도 있다. 셸과 `moai` 자신만 건너뛰면 남는 첫 조상이 세션이다.

### 2.4 실측 검증 — 배포 바이너리로 end-to-end

`make build` 후 셸에서 실행:

```
$ ../bin/moai session register 99999999-aaaa-bbbb-cccc-dddddddddddd NONE none --json
$ cat .moai/state/active-sessions.json
  "pid": 69111,
$ kill -0 69111 ; echo rc=$?
rc=0
$ ps -o pid=,ppid=,comm= -p 69111
69111 68508 claude
```

기록된 PID 69111 은 §2.3 계보에서 확인된 바로 그 claude 세션 프로세스다. 등록 직후 `kill -0` rc=0 — 카드가 요구한 회귀 단언이 실제 바이너리에서 성립한다.

### 2.5 회귀 테스트

`internal/session/session_pid_test.go` 7개. 계보 조회를 패키지 var 시임(`procInfo` / `pidIsAlive`)으로 뽑아 **합성 프로세스 트리**로 검증하므로 환경에 의존하지 않는다:

| 테스트 | 검증 대상 |
|---|---|
| `TestAncestorSessionPID_SkipsWrapperShells` | moai → bash → claude 계보에서 claude PID 반환 |
| `TestAncestorSessionPID_CollapsedChain` | 셸이 접혀 moai → claude 인 경우, 버전 이름 프로세스도 인식 |
| `TestAncestorSessionPID_Unresolvable` | 조회 불가 / init 도달 / 조상 사망 / 깊이 초과 4형태 모두 0 반환 |
| `TestSessionPIDFromEnv_RejectsUnusableValues` | 죽은 PID·음수·비정수·공백 override 거부 |
| `TestResolveSessionPID_PrefersEnvOverride` | 해석 순서 고정 |
| `TestResolveSessionPID_FallsBackToSelf` | 계보 조회 불가 플랫폼에서 기존 동작으로 폴백 |
| `TestRegister_RecordsLivePID` | **카드의 회귀 단언** — 등록 직후 기록된 PID 가 살아 있음 |

## 3. baseline 귀속

기준 트리: `WT-t144` @ base `4100d8767`, 변경 6파일.

| 명령 | 결과 |
|---|---|
| `go build ./...` | rc=0 |
| `go vet ./internal/session/... ./internal/config/...` | rc=0 |
| `GOOS=linux go vet ./internal/session/...` | rc=0 |
| `GOOS=windows go vet ./internal/session/...` | rc=0 |
| `go test ./internal/session/... ./internal/config/...` | ok (4패키지) |
| `go test ./internal/hook/... ./internal/kanban/...` | ok (11패키지) |
| `golangci-lint run internal/session/... internal/config/...` | 0 issues |
| `make build` | rc=0 |

크로스 플랫폼 vet 을 따로 돌린 이유: `proc_info_*.go` 3개가 빌드 태그로 갈라져 있어 darwin 빌드만으로는 linux·windows 분기가 **컴파일조차 되지 않는다**.

## 4. 미검증 (Gaps)

- **linux 계보 조회는 실기기 미검증.** `/proc/<pid>/stat` 파서는 `GOOS=linux go vet` 으로 컴파일만 확인했고 실제 리눅스 프로세스 트리에서 돌려보지 않았다. 파싱 자체는 표준 형식이고 마지막 `)` 기준 앵커로 comm 안의 공백·괄호를 견디게 짰지만, 실측은 아니다.
- **windows 는 설계상 미해결.** 조상 조회를 지원하지 않으므로 `os.Getpid()` 폴백 — 즉 windows 에서는 결함이 그대로 남는다. 다만 `anchor_pid_windows.go` 가 모든 PID 를 alive 로 보고하므로 liveness probe 자체가 windows 에서는 이미 보수적으로 무력화돼 있어, 이번 결함(살아있는 세션을 100% 사망 판정)은 애초에 windows 에서 발현하지 않는다.
- **전체 스위트 미실행.** 카드 지시대로 영향 패키지만 돌렸다(session/config/hook/kanban). 전 패키지 판정은 CI 몫.
- **훅 경로의 실제 부모가 무엇인지 직접 관측하지 않았다.** §2.4 검증은 셸에서 `moai session register` 를 직접 실행한 형태(moai → zsh → claude)다. 훅이 실제로 소환될 때의 계보(claude → sh → moai, 혹은 exec 로 접혀 claude → moai)는 래퍼 스크립트를 읽어 추론했을 뿐 런타임에서 찍어보지 않았다. 두 형태 모두 스킵 규칙이 덮도록 짰고 합성 트리 테스트가 둘 다 커버한다.

## 5. 잔여 위험

- **조상이 claude 가 아닐 수 있다.** `moai session register` 를 사용자가 터미널에서 직접 치면 계보가 moai → zsh → (claude 없이) Terminal.app 으로 이어질 수 있다. 그 경우 터미널 PID 가 기록된다 — 세션보다 오래 살므로 "살아 있음"으로 읽혀 격리 방향(보수적)으로 실패한다. 죽은 PID 를 기록해 "동시 세션 없음"으로 오독하던 종전 방향보다 안전한 오차다.
- **`MOAI_SESSION_PID` 는 아직 아무도 세팅하지 않는다.** 결정적 경로를 원하면 런처(`moai cc` / `glm` / `cg`)가 `syscall.Exec` 직전에 자기 PID 를 이 변수로 내보내면 된다 — workers.json 이 옳은 것과 같은 원리이고, 훅까지 상속된다. 이번 카드 범위에는 넣지 않았다(런처 5개 경로 변경 = 범위 확대). 후속 후보.
- **PID 재사용.** 세션이 죽고 OS 가 같은 PID 를 재발급하면 liveness probe 가 거짓 양성. `DefaultStaleMinutes` heartbeat 하한이 완화하지만 제거하지는 않는다 — 이번 변경 이전부터 있던 성질이다.

## 6. 범위 밖으로 둔 것

- `agent-common-protocol.md` § Pre-Edit Sync Check 의 stale-registry caveat 문안은 **고치지 않았다.** 그 서술("죽은 PID 는 무시, 불확정은 살아있는 것으로 취급")은 이번 수정 이후에도 그대로 옳다 — 결함은 문서가 아니라 배선이었다.
- Template-First 미러: 이번 변경은 Go 코드 전용이라 `internal/template/templates/` 아래 대응 파일이 없다. `make build` 는 규율대로 실행했다(rc=0).

---

## 변경 파일 (6)

| 파일 | 성격 |
|---|---|
| `internal/session/session_pid.go` | 신규 — 해석기(env override → 조상 탐색 → self 폴백) |
| `internal/session/proc_info_bsd.go` | 신규 — darwin/BSD sysctl 계보 조회 |
| `internal/session/proc_info_linux.go` | 신규 — `/proc/<pid>/stat` 계보 조회 |
| `internal/session/proc_info_other.go` | 신규 — 미지원 플랫폼 스텁 |
| `internal/session/session_pid_test.go` | 신규 — 회귀 테스트 7개 |
| `internal/session/registry.go` | `os.Getpid()` → `resolveSessionPID()` 1줄 |
| `internal/config/envkeys.go` | `EnvMoaiSessionPID` 상수 추가 |
