# t552 판정 — GH #1639 세 축은 이미 고쳐져 있다

- 카드: t552 (리드 발행 2026-09-08, t546 이슈 스윕 §D)
- 트리: `.claude/worktrees/t552`, 브랜치 `WT-gate-output-timeout`, base `d060e0d13` (로컬 develop)
- 측정 일자: 2026-09-10
- 바이너리: 이 워크트리에서 `make build` 로 만든 `bin/moai` (`BuildID=moai_cp/20260910_130400-14-gd060e0d13`)

## Claim

GH #1639 이 제기한 세 증상은 **현재 develop 트리에서 모두 재현되지 않는다.** 세 축 전부
`SPEC-GATE-THREE-AXES-001` (status: `completed`, updated 2026-08-27, 카드 t235 유래) 이 이미
닫았다. t552 는 t235 의 중복 카드이며, 새로 고칠 코드는 없다.

원인이 갈리는지 여부(카드가 요구한 분할 판정)도 답이 나온다 — **셋은 원인이 서로 달랐고**,
이미 축별로 다른 수리가 들어가 있다(M1 요약 방출 / M2 프로세스 그룹 종료 / M3 실행 락).
그래서 카드를 쪼갤 실익도 없다.

잔여 1건은 결함이 아니라 설계 선택이다 — §Residual-risk.

## Evidence

세 축을 각각 따로 재현했다. 픽스처는 세션 스크래치패드에 만들었고(`fx-a` Go 통과,
`fx-b` Go 타임아웃+자손, `fx-c` Node tier-(i)), 모든 실행은 `CLAUDE_PROJECT_DIR` 를 벗긴
단일 복합 호출로 돌렸다. 부하를 만드는 실행은 전부 바깥에서 `timeout` 으로 한정했다.

### 축 1 — 통과한 실행이 0바이트를 찍는가 (아니오)

```
$ unset CLAUDE_PROJECT_DIR && cd <fx-a> && bin/moai gate > out.txt 2> err.txt
exit=0
--- stdout bytes: 0 ---
--- stderr bytes: 853 ---
quality gate steps (4 configured):
  - go vet: executed in 835ms at 2026-09-10T13:46:03.940+09:00 — go vet ./...
  - typecheck: skipped — no default for this language; set gate.typecheck.command to enable one
  - golangci-lint: skipped — none of its config files exist in the project directory (...)
  - go test: executed in 1.882s at 2026-09-10T13:46:04.779+09:00 — go test ./...
```

통과한 실행이 853바이트를 낸다. 구성된 4스텝을 전부 이름 대고, 스텝마다 executed/skipped
결과·실측 소요·시각·**실제로 넘긴 명령줄**을 적는다. skipped 는 건너뛴 이유까지 적는다.

제보자의 실제 환경(bun+turborepo)에 해당하는 Node tier-(i) 경로도 같은 모양이다:

```
$ ... cd <fx-c> && bin/moai gate      # package.json scripts."test:run" 존재
exit=0   stderr bytes: 1044
  - npm run test:run: executed in 196ms at 2026-09-10T13:47:19.186+09:00 — npm run test:run
```

즉 리플레이(196ms)와 실제 실행(콜드)은 **실측 소요와 명령줄로 구분된다** — #1639 §1 이
없다고 지적한 바로 그 구분이다.

### 축 2 — 60s 설정인데 915s 생존하는가 (아니오)

`timeouts.test: 3` 으로 두고, 테스트가 stdout 을 물고 있는 자손(`sleep 300`)을 띄운 뒤
자기도 300초 자게 만든 픽스처:

```
$ ... cd <fx-b> && timeout 90 bin/moai gate
pre  sleep300 count: 0
1789015596.985111000        # 시작
exit=1
1789015600.616378000        # 종료 → 벽시계 3.63s
post sleep300 count: 0      # 자손도 죽었다

quality gate timed out: go test exceeded 3s; termination was signalled to the
step's whole process group, so processes the step started were signalled too
  - go test: executed in 3.003s at 2026-09-10T13:46:37.608+09:00 — go test ./...
```

설정 3s → 스텝 실측 3.003s, 게이트 전체 3.63s(3s 데드라인 + 2s `WaitDelay` 여유 안쪽),
자손 프로세스 0 생존. 카드가 [HARD] 로 요구한 "설정값을 읽는 경로 vs 적용하는 경로"의
분리 관측이 이 한 실행에 다 들어 있다 — 읽은 값(3s)이 보고문에 그대로 인용되고, 적용
결과(3.003s 에서 끊김 + 그룹 종료)가 별도로 관측된다.

구현 위치: `internal/hook/quality/step_process_group{,_unix,_windows}.go` 의
`isolateProcessGroup`(스폰 시 `Setpgid`) + `terminateProcessGroup`(`kill(-pid)`), 그리고
`gate.go:runStep` 의 `cmd.WaitDelay = stepWaitGrace`(2s). #1639 §2 가 추정한 수리
방향(프로세스 그룹 킬)과 같다.

### 축 3 — 수동 gate 가 직렬화되지 않는가 (아니오, 직렬화된다)

같은 체크아웃에서 gate 두 개를 0.3초 간격으로 띄웠다:

```
$ ... (bin/moai gate > o1 2> e1) & sleep 0.3; bin/moai gate > o2 2> e2
B exit=0
e2 첫 줄: gate-run lock: held by pid 44336 — waiting (budget 4m0s)
```

두 번째 실행이 락 대기에 들어가면서 **보유자 pid 와 대기 예산을 이름 대고 알린다.**
구현은 `internal/cli/gate_lock.go` + `internal/cli/gate.go`(`waitForGateLock`), #1639 §3 이
요청한 세 성질을 모두 갖췄다 — 한정된 대기(`gate.timeouts.lock_wait`), 예산 만료 시
비직렬 실행으로의 일방향 강등, 대기 고지. 락 결과는 종료코드에 관여하지 않는다.

### 회귀 스위트

```
$ go test ./internal/hook/quality/...          ok  63.732s
$ go test ./internal/cli/ -run 'GateLock|GateSummary|Gate_|Precommit' -count=1
                                                ok  97.446s
$ go test ./internal/hook/quality/ -run 'Summary|Termination|TimeoutAttribution' -v | grep -c '^--- PASS'   → 11
$ go test ./internal/cli/ -run 'GateLock|GateSummary' -v | grep -c '^--- PASS'                              → 6
```

셀렉터가 0매치로 조용히 초록이 되는 경우를 배제하려고 `--- PASS` 를 세었다: 각각 11건,
6건이 실제로 돌았다.

## Baseline-attribution

- 트리: `d060e0d13` (로컬 develop, 이 워크트리 HEAD). 측정 전 `git status --porcelain` 공백.
- 바이너리: 위 트리에서 이번에 빌드한 것 — 설치본(`~/go/bin/moai`)이 아니다.
- 모든 수치는 이 실행에서 관측한 값이며, SPEC progress.md 나 t235 보고서에서 옮겨온 값이
  아니다. SPEC 의 `status: completed` 는 **판정 근거가 아니라 단서로만** 썼다.

## Gaps

- Windows 경로(`step_process_group_windows.go`)는 이 트리에서 실행 관측하지 않았다 —
  darwin 에서만 쟀다. 크로스 플랫폼 판정은 CI 매트릭스 몫이다.
- 제보자가 겪은 **915초 생존 자체**는 재현하지 않았다. 그 수치는 수리 이전 버전(v3.1.2,
  `a1b1ca696`)의 관측이고, 이 카드가 답할 질문은 "지금 트리에서 재현되는가"였다.
  구버전에서의 재현은 시도하지 않았다.
- turbo 가 실제로 낀 실물 monorepo 는 쓰지 않았다. `fx-c` 는 tier-(i) 해석 경로를 타게
  하는 최소 픽스처다.
- 전체 스위트(`go test ./...`)는 로컬에서 돌리지 않았다(§4.1 규율). 건드린 코드가 없으므로
  회귀 위험 자체가 없다.

## Residual-risk

**#1639 §1 의 문자 그대로의 제안 하나는 구현되지 않았다.** 제보자는 "최소한 turbo 요약
줄을 `runStep` 밖으로 들고 나오라"고 했는데, 통과한 스텝의 stdout 은 지금도 버려진다
(`gate.go:runStep`, `if err == nil { return true, "" }`). `fx-c` 에서 심어둔
`CACHED-REPLAY-MARKER` 는 stdout/stderr 어디에도 나타나지 않았다(grep 카운트 0/0).

다만 그 제안이 **섬기려던 요구**(리플레이와 실제 실행의 구분)는 다른 방식으로 충족됐다 —
실측 소요 + 명령줄. SPEC 이 요구를 그렇게 정식화했고(§A "nothing names the command the
verdict was actually delegated to"), 스텝 stdout 을 통째로 흘리면 통과한 게이트가 수천 줄을
쏟는 반대쪽 문제가 생긴다. 그래서 이것은 미완의 결함이 아니라 설계 선택으로 본다. 제보자가
그 줄 자체를 원한다면 별도 카드 사안이지, 이 카드의 미결이 아니다.

## 권고

1. 카드 t552 는 **코드 변경 없이** 종결한다(중복: t235 → SPEC-GATE-THREE-AXES-001).
2. GH #1639 는 세 축 모두 수리 완료로 회신하고, 수리분이 실린 버전이 출시될 때 닫는다
   (현재 develop, v3.1.2 사용자에게는 아직 미출시). 회신에 위 축별 증거를 인용한다.
3. 이슈 스윕 절차 쪽 후속 — t546 스윕이 이미 닫힌 SPEC 을 교차확인하지 않고 카드를 냈다.
   스윕 단계에 "해당 이슈를 다룬 SPEC 이 이미 completed 인지" 확인을 넣는 것을 별도
   카드로 제안한다.
