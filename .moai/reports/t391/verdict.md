# t391 — 사전 판독 (plan-phase 진입 전)

- 카드: t391 (`moai codex` 가 론처가 아니라 상태 판독기다)
- 워크트리: `.claude/worktrees/t391`, 브랜치 `WT-codex-launch-verb`
- 측정 트리: `e79272713` (= `origin/develop`, 진입 시점 `git rev-list --count --left-right origin/develop...HEAD` → `0	0`)
- 판독 시점: 2026-08-31

이 문서는 판독만 담는다. 설계·구현은 plan 이후 산출물이 가진다.

## 1. Claim — 무엇을 주장하는가

| # | 주장 |
|---|---|
| C1 | 맨몸 `moai codex` 가 기동하지 않는 것은 결함이 아니라 **기록된 판정**이다. `plan.md` §B 가 (a) 기동 / (b) 리드아웃+명시기동 두 안을 놓고 (b)를 택했다 (2026-08-24 리드 판정). 이번 카드는 그 판정을 뒤집는 일이다. |
| C2 | (b)의 기각 근거였던 "실수로 현재 세션이 codex 에 넘어갈 위험"은 **현행 구현에서 무게가 다르다**. cc/glm 은 `syscall.Exec` 로 프로세스를 교체하지만, codex 는 SPEC 0.8.0 개정에서 `os/exec` 자식 + 종료코드 전파로 갔다. 자식이 끝나면 셸로 돌아온다. |
| C3 | 카드 (2)항 "환경 적응 대칭성"은 **REQ-CL-013 과 충돌한다**. settings 정리·프로필 적용은 codex 로 옮길 수 없다. |
| C4 | `codexInitOfferGate` 는 두 launch 동사의 **단일 통과 지점**이라, 맨몸을 `cli` 로 옮기면 INIT-001 게이트가 그대로 상속된다 (신규 설계 대상이 아니라 상속 확인 대상). |
| C5 | `codexDirectLaunch` 는 자식에게 `Env` 를 지정하지 않는다 — CODEX_HOME 을 해석해 리드아웃에 쓰면서도 자식에게 명시 전달하지 않는다. |
| C6 | 세 론처를 한 자리에서 대조하는 시험은 **있다**. 다만 대조 항목이 그룹 소속·GroupID·spawn 진단 바이트뿐이고, **"인자 없는 호출이 기동으로 이어지는가"를 단언하는 셀은 없다**. |

> C6 은 이전 세션에서 리드에 "지금은 없음"으로 보고했던 항목의 **정정**이다. 교차 시험 파일은 존재하며, 없는 것은 그 파일 자체가 아니라 그 파일이 다루지 않는 축이다.

## 2. Evidence — 잰 명령과 관측

측정은 전부 이 트리(`e79272713`)에서 다시 수행했다. 이전 세션의 판독을 옮겨 적지 않았다.

**C1** — `.moai/specs/SPEC-CODEX-LAUNCHER-001/plan.md` §B:

> **판정: (b)** — 리드 판정 (2026-08-24). REQ-CL-002 를 그 방향으로 재기술하고 AC-CL-002 의 조건절을 해소했다.

기각된 (a)의 서술과 (b)의 채택 근거("실수로 현재 세션이 codex 에 넘어갈 위험이 없다")가 같은 절에 있다.

`spec.md:113` REQ-CL-002:

> The bare `moai codex` command shall print the readiness readout and exec nothing; launching shall require an explicit verb — `cli` … `status` shall be accepted as an explicit alias of the bare readout form.

**C2** — `internal/cli/launch_exec_posix.go:24-26`:

```go
func execOrSpawnClaude(claudeBin string, args, env []string) error {
	return syscall.Exec(claudeBin, args, withSessionPID(env, os.Getpid()))
}
```

`spec.md:33` (HISTORY 0.8.0):

> **교체를 버리고 전 플랫폼 공통 `os/exec` + 종료코드 전파** 로 통일한다 (운영자 판정).

구현에서 확인: `codexDirectLaunch` 는 `exec.Command` 자식을 만들고 `codexPropagateLaunchError` 가 `*exec.ExitError` 의 코드를 그대로 되돌린다 (`internal/cli/codex_launcher.go:265-300`).

**C3** — `spec.md:133` REQ-CL-013:

> The system shall not mutate `.claude/settings.local.json`, Claude profile state, or any file under `CODEX_HOME` on any verb — the launcher reads state and execs; it does not write.

**C4** — `grep -rn 'codexInitOfferGate' internal/cli/*.go | grep -v _test` → 2행: 정의 1건(`codex_init.go:152`), 호출 1건(`codex_launcher.go:256`). 호출 지점은 `runCodexLaunch` 안이고, `runCodex` 는 `kind.launches()` 인 경우에만 그리로 간다 (`codex_launcher.go:203-205`). 즉 `cli`/`app` 두 동사가 같은 함수를 지난다.

**C5** — `grep -n 'c.Env\|Env =' internal/cli/codex_launcher.go` → 출력 0행. `codexDirectLaunch` 는 `c.Dir`/`c.Stdin`/`c.Stdout`/`c.Stderr` 만 지정한다. CODEX_HOME 해석은 `resolveCodexHomeDir` (`internal/cli/mcp_codex.go:1682`) 가 리드아웃 경로에서만 수행한다.

**C6** — `grep -rln 'ccCmd\|glmCmd' internal/cli/*_test.go | xargs grep -ln 'codexCmd\|runCodex'` → `internal/cli/codex_launcher_test.go` 1건. 그 파일이 실제로 대조하는 것:

- L228-232: `--help` 의 launchers 블록에 `cc`/`glm`/`cg`/`codex` 가 모두 있는가
- L235: `codexCmd.GroupID != ccCmd.GroupID`
- L532-545 (`TestCodexSpawn_TmuxAbsentDiagnosticBytes`): tmux 부재 시 codex 와 `cc --spawn` 의 진단 바이트 동일성

세 셀 어디에도 "인자 없이 부르면 기동하는가"를 세 론처에 걸쳐 단언하는 것은 없다.

참고로 `internal/cli/codex_launcher_cross_test.go` 는 이름과 달리 론처 간 대조가 아니라 **표면 간**(리드아웃 커맨드 ↔ `codex_setup` MCP 응답) sentinel 전파 시험이다.

## 3. Baseline-attribution — 무엇에 대고 쟀는가

모든 인용은 `.claude/worktrees/t391` 트리, HEAD `e79272713`. 진입 직후 `git rev-list --count --left-right origin/develop...HEAD` = `0	0` 로 `origin/develop` 과 동일함을 확인했다. 파일 좌표(`file:line`)는 이 트리 기준이며, 트리가 움직이면 다시 재야 한다.

브랜치는 `worktree-t391` 로 생성된 뒤 `WT-codex-launch-verb` 로 개명했다. 같은 이름의 낡은 브랜치(`59e898b31`)가 남아 있었고, `git merge-base --is-ancestor` 로 `origin/develop` 의 조상임을 — 즉 고유 커밋 0건임을 — 확인한 뒤 삭제했다.

## 4. Gaps — 관측하지 않은 것

| # | 미관측 |
|---|---|
| G1 | 설치본(`~/go/bin/moai`)의 `--help` 문자열을 이 세션에서 재측정하지 않았다. 카드가 인용한 v3.1.2 문구는 리드의 측정이며 내 것이 아니다. |
| G2 | 빌드·테스트 0건. `go build`/`go test` 를 이 트리에서 돌리지 않았다. |
| G3 | REQ-CL-013 이 다른 Codex SPEC 의 수용 칸(AC)에도 걸려 있는지 훑지 않았다. plan 에서 그 범위를 확인해야 한다. |
| G4 | `-k`/`-f` 를 codex 에 여는 문제는 **읽은 어떤 것도 답하지 않는다**. 열린 질문이며 이 카드 범위 밖(운영자 판정 2026-08-31). |
| G5 | `codex app` 경로의 실제 기동 거동은 관측하지 않았다 (codex 바이너리를 실행하지 않았다). |

## 5. Residual-risk — 관측했는데도 틀릴 수 있는 것

- C2 는 "셸로 돌아온다"를 **코드 형태**로 논증했을 뿐 실제 왕복을 관측하지 않았다. `spec.md` 0.8.0 자신이 tty 왕복을 CI 미관측 Gap 으로 기록하고 있다 — 같은 Gap 을 물려받는다.
- C6 의 "없다"는 `ccCmd`/`glmCmd` 식별자를 참조하는 테스트 파일 집합에 대한 부재 주장이다. 식별자를 쓰지 않고 문자열이나 다른 경로로 같은 축을 단언하는 시험이 있다면 이 주장은 좁다.
- C4 의 상속은 **호출 그래프**로 확인한 것이지, 맨몸을 `cli` 로 옮긴 뒤의 거동을 관측한 것이 아니다. AC 로 단언해야 한다(맨몸 + 미배선 + 비대화형 = 프롬프트 0, 기동 0).

## 6. 개정 대상 범위 (운영자 판정 2026-08-31 반영)

- 개정 대상은 **REQ-CL-002 하나**. settings 정리·프로필 적용은 요구에서 제외(C3). REQ-CL-013 은 유지.
- 남는 작업 4항: (1) 맨몸 기본 동사를 `cli` 로, `status` 는 명시 별칭으로 보존 (2) CODEX_HOME 자식 명시 전달(C5) (3) `-w` 워크트리 경로 정규화 — codex 에는 현재 `-w` 자체가 없다(`codex_launcher.go` 에 `-w`/worktree 매치 0건); `cc.go:220-228` 의 `resolveWorktreeL2Path` + `normalizeWorktreeFlag` 가 대응 지점 (4) 세 론처 맨몸 규약을 한 자리에서 대조하는 시험(C6).
