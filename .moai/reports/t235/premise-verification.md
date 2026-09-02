# t235 — 카드 전제 실측 검증

- 측정 트리: `.claude/worktrees/t235` (branch `WT-gate-three-axes`)
- 측정 SHA: **294b4b6ab**
- 대상 카드: `moai todo` t235 (#1639)

## 결론

3축 전부 **현재도 성립**. 다만 카드가 인용한 줄번호와 lockfile 소비처 열거에 오차가 있어 아래로 정정한다.

## 축별 실측

### 축 1 — 통과 실행이 0바이트

| 카드 주장 | 실측 (294b4b6ab) | 판정 |
|---|---|---|
| `gate.go runStep` 성공 시 `err == nil → return true, ""` | `internal/hook/quality/gate.go:1020-1022` 동일 | 성립 |
| `Run`의 passReason이 skip 통지만 운반 (`:343-402`) | 실제 `Run`은 `gate.go:322-402`. passReason은 vet/typecheck/lint/ast-grep의 skip 통지를 누적하나, **test 스텝은 pass-side 출력을 아예 버린다**(`if ok, out := g.executeStep(...); !ok { return false, out }` — 성공 경로에서 `out` 미사용) | 성립 (카드보다 강함) |
| CLI가 빈 출력 시 무인쇄 (`cli/gate.go:69-80`) | `internal/cli/gate.go:69-81`. `output != ""`일 때만 stderr 출력 | 성립 |
| `resolveNodeTestStep` tier-(i)가 verdict를 `scripts.test:run`에 위임 (`gate.go:652-701`) | 실제 `gate.go:676-700`. `npm run test:run`으로 치환 확인 | 성립 |

정정: 파일 위치는 `internal/hook/quality/gate.go`, 줄번호는 카드 대비 약 +20~24행 이동.

### 축 2 — step timeout 미집행

| 카드 주장 | 실측 | 판정 |
|---|---|---|
| `CommandContext`가 직속 자식만 죽임 | `gate.go:1006` `exec.CommandContext(stepCtx, name, args...)`, `SysProcAttr` 미설정 | 성립 |
| quality 패키지에 Setpgid/killpg 0건 | `grep -rn 'Setpgid\|Killpg\|killpg\|SysProcAttr' internal/hook/quality/` → rc=1, 0건 | 성립 |
| 10분 외부 안전망이 Wait 봉쇄를 못 자름 | `internal/cli/gate.go:65` (카드는 `:57-59`) `context.WithTimeout(..., 10*time.Minute)` — `cmd.Run()`의 Wait 봉쇄에는 무력 | 성립 |

부하 의존 수치(915s 생존)는 본 저장소에서 **미재현**. 구조 사실만 확정 — 카드 서술과 동일.

기착지 픽스 주의: t218의 timeout 귀속 수정(`parentBinds` 분기, `gate.go:996-1002`)과 그 회귀 테스트 2건(`gate_timeout_attribution_test.go:44,67`)이 이미 존재. 축 2 변경은 이 귀속 의미를 보존해야 한다.

### 축 3 — 수동 gate 미직렬화

| 카드 주장 | 실측 | 판정 |
|---|---|---|
| `cli/gate.go`에 lock 0건 | 전문 215행 통독 — lock 획득 코드 0건 (`lock` 문자열 히트는 전부 주석 또는 `BlockOnError`) | 성립 |
| `internal/lockfile` 소비처 = glm_tools·settings·taskledger·**kanban board_lock** | 실제 import 3곳뿐: `internal/cli/settings.go:12`, `internal/cli/glm_tools.go:32`, `internal/cli/taskledger/taskledger.go:16`. **kanban board_lock은 소비처가 아니다** — `internal/kanban/board_lock.go:10-11`이 "internal/lockfile's in-process mutex is neither used nor upgraded"라고 명시하고 `internal/spec/lock.go` 계열 패턴을 재사용 | **부분 반증** |
| 업스트림에 직렬화 경로 자체가 없음 | 성립 | 성립 |

설계상 함의(카드 정정): 축 3이 요구하는 "bounded wait + 만료 시 무직렬 열화"에는 `internal/lockfile`(Lock/Unlock 2개, 블로킹 flock, Windows는 in-process mutex라 크로스프로세스 무보호)보다 **`internal/spec/lock.go` 계열이 더 맞는 기반**이다: `AcquireSpecCloseLock`이 경합 시 `ErrSpecCloseLockHeld`를 즉시 반환하는 try-semantics를 이미 갖고 있고(`internal/spec/lock.go:50`), Windows는 atomic-create(`O_CREATE|O_EXCL`)로 크로스프로세스가 실제로 보호된다. `internal/kanban/board_lock.go`가 같은 패턴에 stale-clear를 얹은 선례.

## 미검증 (Gaps)

- 하류(mo.ai.kr) 리포트 t324·t314의 수치는 본 트리에서 재현하지 않음 — 부하 의존.
- `moai gate` 실제 실행 시간(11m19s 등) 미재측정.
- #1631(t233, 같은 `executeStep`의 lint 축)은 본 카드 범위 밖 — 미착수 확인만 함.
