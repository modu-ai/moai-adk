auditor-model: claude-opus-5-5
verdict: FAIL
audited_sha: 9aa451b107b6e59db0fad5a1e1c0d21950fcbef4

## Evaluation Report — 카드 t1670 sync-audit 3회차 (델타 재감사)

SPEC: SPEC-LOCAL-MAIN-FLOW-001
Overall Verdict: FAIL — overall score 82/100

감사 대상: 워크트리 `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1670`, 브랜치 `WT-t1616-10-10`, HEAD `9aa451b10`(전체 SHA는 위 `audited_sha`), 기준 local main `67e462e80`. 판정 대상 수리 커밋은 `afc00db6f`(시험 전용)이고, `afc00db6f..HEAD` 사이에는 `.moai/reports/t1670/` 문서 두 개만 바뀌었다.

### 요약

트립와이어 방식은 2회차에서 빠져 있던 mutant B(재검증과 merge를 두 구간으로 쪼갬)를 확실히 잡는다. 이 감사가 직접 만든 변형 다섯 개 가운데 A·B·C·E는 5/5 실패했고, 실제 트리는 5/5 통과했다. 시험 구조(양성 대조, 고루틴 정리, race)도 건전하다.

그런데 **소유권 읽기(`ReadIntegrationLock`)만 구간 앞으로 끌어내고 나머지 검사는 구간 안에 두는 변형(mutant D)** 은 새 시험과 기존 시험 모두 5/5 통과했다. 트립와이어는 재검증 **끝** seam에서 심어지므로, 그보다 앞서 일어나는 읽기의 위치는 볼 수 없다. F2가 막으려는 경합이 바로 "소유권을 읽은 뒤 이동하기 전에 경쟁자가 창을 가져가는" 것이라, 이 변형은 원래 결함을 그대로 되살린다. 필수 백엔드 codex도 독립적으로 같은 변형을 만들어 같은 결론(verdict fail, P2)을 냈다. 필수 게이트가 FAIL이므로 이 감사는 PASS를 낼 수 없다.

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%, must-pass) | 70/100 | FAIL | 실제 트리 `DuringRevalidate PASS=5 FAIL=0`, A/B/C `PASS=0 FAIL=5`, E는 두 시험 모두 `FAIL=5` (E1). 그러나 D는 `DuringRevalidate PASS=5 FAIL=0` / `BeforeFastForward PASS=5 FAIL=0` (E1, E2). codex required `fail` P2, 같은 변형 (E5). 필수 통과 차원이라 이것만으로 전체 FAIL |
| Security (25%, must-pass) | 95/100 | PASS | 델타는 시험 파일 하나(`git diff --stat afc00db6f HEAD`는 보고서 두 개만, E4). 트립와이어 쓰기 대상은 `t.TempDir()` 기반 fixture 안의 `.moai/state/integration-mutation.lock`. 운영 코드 변경 없음 |
| Craft (20%) | 88/100 | PASS | `-race -count=1 -run TestLocalMainResync`: PASS 17, FAIL 0, DATA RACE 0 (E3). `go vet` exit 0, `gofmt -l` 출력 없음, `golangci-lint --new-from-rev a535b88fc` `0 issues.` (E4). 고루틴은 WaitGroup으로 판정 전에 합류, 30초 상한과 stop 채널로 종료 보장 |
| Consistency (15%) | 90/100 | PASS | 기존 seam 복원(`t.Cleanup`) 방식과 같은 모양. 잠금 파일 이름을 시험에 하드코딩한 점은 선택 사항(F3) |

가중 합: 0.40×70 + 0.25×95 + 0.20×88 + 0.15×90 = 82.85 → 82 (필수 통과 차원 FAIL과 필수 게이트 FAIL로 전체 FAIL).

### Findings (structured defect-list)

- **F1 [P2] [blocking]** `internal/cli/local_main_resync_gap_test.go:121`(트립와이어를 심는 seam) 및 `internal/cli/local_main_resync.go:304`(소유권 읽기) — 트립와이어 구간은 `localMainResyncAfterRevalidate`(재검증 끝)에서 시작한다. 소유권 읽기만 `RunIntegrationMutation` 앞으로 옮기고 그 결과를 구간 안의 검사에 넘기는 회귀(mutant D)는 구간 안에서 잠금을 다시 잡지 않으므로 트립와이어가 그대로 남고, 경쟁자 고루틴도 구간 안에서 출발하므로 끼어들 틈이 없다. 실측: D에서 두 rival 시험 모두 5/5 PASS (E1, E2). codex는 같은 변형에서 `-run '^TestLocalMainResync'` 전체가 통과했고, 별도 probe에서 읽기 직후 경쟁자가 창을 가져갔는데도 fast-forward가 실행됐다고 보고했다 (E5, codex 보고이며 이 감사가 그 probe를 재현하지는 않았다). "빨리 실패하려고 잠금 전에 레코드를 먼저 읽는다"는 리팩터는 충분히 있을 법한 변경이라 회귀 고정 대상에 든다. 확신도 높음.
  - Required fix: 트립와이어와 경쟁자의 출발점을 **소유권 읽기 자체**에 묶는다. 재검증 함수 안의 seam은 읽기를 끌어낸 변형을 따라가지 못하므로(seam은 구간 안에 남는다) 위치만 바꾸는 것으로는 부족하다. 권장 방식: (1) `localMainResyncRevalidate`의 `factory.ReadIntegrationLock(repoRoot)` 호출을 패키지 변수 seam(예: `var localMainResyncReadWindow = factory.ReadIntegrationLock`)을 거치게 한다. (2) 시험에서 이 seam을 감싸 읽기 직전에 트립와이어를 심고, 같은 자리에서 경쟁자의 `--force` 탈취를 **동기적으로** 한 번 시도해 `factory.IsIntegrationLockBusy`로 거절되는지 단언한다. 잠금 밖에서 읽는 변형이면 이 탈취가 성공하므로 결정적으로 실패한다. (3) 기존 끝-seam 트립와이어 확인과 양성 대조는 그대로 둔다. 수리 후 mutant A·B·C·D·E 모두 실패하고 실제 트리는 통과하는지 `-count=5`로 재측정해 evidence에 남긴다. 운영 코드 변경은 seam 변수 정의와 호출 치환 한 줄 수준이다.
- **F2 [low] [optional]** `internal/cli/local_main_resync_gap_test.go:172` — 양성 대조는 Unix에서만 판별력이 있다. Windows 기판(`state_lock_windows.go`)은 해제 시 잠금 파일을 지우므로, 획득이 기록을 다시 쓰든 말든 트립와이어는 사라진다. 다만 Windows에서도 변형 탐지 자체는 유지된다고 코드 읽기로 판단했다: B는 구간 1 해제 때 파일이 지워지고 구간 2가 새로 만들며, A는 심은 파일이 O_EXCL 생성을 막아 busy로 거절돼 re-sync가 실패한다. Windows에서 실측하지 않았으므로 추론이다. 확신도 중간.
- **F3 [low] [optional]** `internal/cli/local_main_resync_gap_test.go:94` — `integration-mutation.lock`을 하드코딩해 `factory` 패키지의 비공개 상수와 이중 관리한다. 이름이 바뀌면 옛 경로에 심은 트립와이어가 남아 양성 대조가 크게 실패하므로 조용히 통과하는 경로는 없다. 시험 코드로서 허용 범위다. 확신도 높음.
- **F4 [info] [optional]** 경쟁자-HEAD 신호는 타이밍에 기대며, 시험 주석도 그렇게 적고 있다. 실측으로는 B에서 0/5 발화(트립와이어가 잡음), E에서 5/5 발화(트립와이어는 잡지 못하고 이 신호가 잡음). 두 신호가 서로 보완한다는 점은 확인됐다. 확신도 높음.

### Recommendations

- F1만 고친다. 범위는 시험 보강과 읽기 seam 한 줄이다. 4회차 재감사 범위는 F1 델타(읽기 seam + 시험 수정 + 변형 A–E 재측정)로 한정한다.
- F2–F4는 참고 사항이며 자동으로 수리 대상에 넣지 않는다.

---

## Claim

- C1. 실제 트리(HEAD `9aa451b10`)에서 새 시험은 `-count=5`로 5/5 통과하고, resync 시험 17건이 `-race` 아래 모두 통과한다.
- C2. 새 시험은 mutant A(재검증 전체를 구간 밖으로), B(두 구간으로 분리), C(재검증 결과를 구간 전에 계산하고 안에서만 확인)에서 5/5 실패한다. 실패 원인은 트립와이어 단언(177행)이다.
- C3. mutant E(재검증은 구간 안, merge는 구간 해제 뒤 잠금 없이)는 기존 시험(45행)과 새 시험의 경쟁자-HEAD 신호(188행)가 각각 5/5 잡는다.
- C4. mutant D(소유권 읽기만 구간 앞으로)는 두 rival 시험을 모두 5/5 통과한다. 따라서 완료 조건 1("재검증과 이동이 같은 락 구간임을 시험으로 확인")은 재검증 중 소유권 읽기 부분에 대해 고정되지 않았다.
- C5. 트립와이어는 빈 파일·쓰기 실패로 조용히 통과하지 않는다: 심기 실패는 `t.Fatalf`, 읽기 실패나 파일 부재는 `intactAtMove=false`로 실패한다. Unix의 "모든 획득이 기록을 다시 쓴다"는 계약은 `acquireStateLockImpl`의 `Ftruncate`+`Write`(flock 성공 뒤에만, 경쟁 패자는 쓰지 않음, `O_TRUNC` 없음)로 확인했고, 양성 대조가 그 계약이 깨지면 크게 실패시킨다.
- C6. 고루틴 정리는 건전하다: `wg.Add`는 seam 안 단일 호출, `halt`가 stop을 닫고 `wg.Wait`로 합류한 뒤 판정, `t.Cleanup(halt)`는 `sync.Once`로 중복 안전, 루프는 30초 상한. `-race`에서 DATA RACE 0.
- C7. 필수 백엔드 codex의 교차 감사 결과는 `fail`(gate required, P2, mutant D와 같은 변형)이다.
- C8. 카드 완료 조건: (1) 미충족(C4). (2) "외부 행위자 경합 재현 0회"는 잠금을 지키는 행위자에 한해 A·B·C·E 기준으로 충족하지만, D 형태의 회귀에 대해서는 고정되지 않았다. 잠금을 무시하는 행위자의 비원자적 abort 창은 이전 회차에서 수용된 잔여 위험 그대로다.
- C9. SPEC `progress.md`에 `Binding run conditions` 제목이 없다(`grep -n -i "binding run conditions"` 출력 없음). 처리할 구속 조건은 없다.

## Evidence

E1 — 변형별 `-count=5` 측정. 스크립트 `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/efe62e11-9343-48de-862c-29586e1c5b0d/scratchpad/run_mutants.sh`가 각 트리에서 `TMPDIR=GOTMPDIR=/private/tmp/t70-audit3/tmp go test ./internal/cli/ -run 'TestLocalMainResyncRivalTakeover' -count=5 -v`를 실행하고 `--- PASS`/`--- FAIL` 줄 수를 셌다. 변형 트리는 `git -C <worktree> archive -o /private/tmp/t70-audit3/head.tar HEAD`로 뽑아 `/private/tmp/t70-audit3/{A,B,C,D,E}`에 풀고 `mutate.py`로 바꿨다(시험 파일은 손대지 않음). 출력 원문:

```
real EXIT=0
  DuringRevalidate PASS=5 FAIL=0
  BeforeFastForward PASS=5 FAIL=0
A EXIT=1
  DuringRevalidate PASS=0 FAIL=5
  BeforeFastForward PASS=5 FAIL=0
B EXIT=1
  DuringRevalidate PASS=0 FAIL=5
  BeforeFastForward PASS=5 FAIL=0
C EXIT=1
  DuringRevalidate PASS=0 FAIL=5
  BeforeFastForward PASS=5 FAIL=0
D EXIT=0
  DuringRevalidate PASS=5 FAIL=0
  BeforeFastForward PASS=5 FAIL=0
E EXIT=1
  DuringRevalidate PASS=0 FAIL=5
  BeforeFastForward PASS=0 FAIL=5
```

실패 메시지 표본(A, B 동일 형태):

```
local_main_resync_gap_test.go:177: the mutation lock was acquired between the ownership re-check and the fast-forward (re-sync err = integration resync: local main fast-forwarded 093db31ee1b1 -> afb9459cd07e to origin/main, but releasing the window failed: release integration window is held by a different session: lane-rival (pid 83985) holds it (moai integration status reads it)): the re-validation and the move do not share one unbroken mutation section
```

E 표본:

```
local_main_resync_gap_test.go:45: a rival --force takeover landed between the ownership re-check and the fast-forward (...)
--- FAIL: TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand (0.47s)
local_main_resync_gap_test.go:188: a rival --force takeover landed while primary HEAD was still f6d91eac76eaac63410abfe00f443f57511b9f8c, before the fast-forward to 5d49b7d72e4b91175c43d1569d278ffc7d57214f (...)
```

`grep -c "landed while primary HEAD was still" E.log` → `5`. D 로그 끝: `PASS` / `ok  github.com/modu-ai/moai-adk/internal/cli 47.788s`.

E2 — 변형 diff(실제 `internal/cli/local_main_resync.go` 대비, `diff` 출력 원문):

```
# C
246a247
> 	precomputed := func() error { return localMainResyncRevalidate(git, repoRoot, sessionID, ref, head) }()
248c249
< 		if revalidateErr = localMainResyncRevalidate(git, repoRoot, sessionID, ref, head); revalidateErr != nil {
---
> 		if revalidateErr = precomputed; revalidateErr != nil {

# D
246a247
> 	mutDPreLock, mutDPreErr = factory.ReadIntegrationLock(repoRoot)
304c305
< 	lock, err := factory.ReadIntegrationLock(repoRoot)
---
> 	lock, err := mutDPreLock, mutDPreErr
481a483,485
>
> var mutDPreLock *factory.IntegrationLock
> var mutDPreErr error

# E
253,254d252
< 		localMainResyncBeforeFastForward(repoRoot)
< 		_, moveErr = git("-c", "merge.autoStash=false", "merge", "--ff-only", "--no-overwrite-ignore", "-q", baseline)
257a256,259
> 	}
> 	if revalidateErr == nil {
> 		localMainResyncBeforeFastForward(repoRoot)
> 		_, moveErr = git("-c", "merge.autoStash=false", "merge", "--ff-only", "--no-overwrite-ignore", "-q", baseline)
```

A와 B는 2회차 감사가 기록한 diff와 같은 모양이다(A: 재검증 호출을 `RunIntegrationMutation` 앞으로; B: 재검증만 담은 첫 구간 + `if revalidateErr != nil` 이후를 담은 둘째 구간).

E3 — race (`TMPDIR=GOTMPDIR=/private/tmp/t70-audit3/tmp go test -race -count=1 -run TestLocalMainResync -v ./internal/cli/`):

```
EXIT=0
PASS 줄 17 / FAIL 줄 0 / DATA RACE 0
--- PASS: TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand (9.89s)
--- PASS: TestLocalMainResyncRivalTakeoverDuringRevalidateCannotLand (0.75s)
ok  	github.com/modu-ai/moai-adk/internal/cli	21.354s
```

E4 — 형식·정적 검사·범위:

```
$ gofmt -l internal/cli/local_main_resync_gap_test.go internal/cli/local_main_resync.go
(출력 없음) GOFMT_EXIT=0
$ go vet ./internal/cli/
VET_EXIT=0
$ golangci-lint run --new-from-rev a535b88fc ./internal/cli/
0 issues.
$ git diff --stat afc00db6f HEAD
 .moai/reports/t1670/card-review.md | 2 +-
 .moai/reports/t1670/evidence.md    | 7 +++++++
$ git diff --stat 67e462e80 HEAD -- internal/
 internal/cli/local_main_resync.go          |   9 ++
 internal/cli/local_main_resync_gap_test.go | 156 +++++++++++++++++++++++++++++
 internal/cli/mcp_build_identity_test.go    |   9 +-
```

잠금 계약 코드 읽기: `internal/factory/state_lock_unix.go` `acquireStateLockImpl` — `unix.Open(... O_CREAT|O_RDWR|O_CLOEXEC|O_NOFOLLOW|O_NONBLOCK ...)`(O_TRUNC 없음) → `Flock(LOCK_EX|LOCK_NB)` 실패 시 쓰지 않고 반환 → 성공 시 `Ftruncate(fd, 0)` + `Write(fd, newLockOwnerRecord())`. `newLockOwnerRecord`의 `created_at`은 `time.Now().UTC()`라 1970 트립와이어와 같을 수 없다. `state_lock_windows.go`는 `O_CREATE|O_EXCL` 생성 후 기록, 해제 시 `os.Remove`.

E5 — 교차 모델 감사 (`moai verify audit-plan --project-root <tree>`: `config_status: ok`, `model: codex`, codex `gate: required`, claude/glm `off`, `cross_model_active: true`). `mcp__moai__audit_multi`(project_root=트리, target=baseBranch, card_id=t1670, gates 인자 없음) 결과 요약 멤버:

```
overall_verdict: fail
gate_unmet: (결과에 멤버 없음)
plan_source: config
per_backend_verdicts: [{backend: codex, gate: required, verdict: fail}]
audit_receipt: rcpt-0bfb550b1026971c5b6655ce
residual_risk_note: required-backend FAIL: codex
```

codex finding 원문 요지: "[P2] 소유권 읽기만 락 밖으로 이동하는 회귀는 감지하지 못한다 — internal/cli/local_main_resync_gap_test.go:121 … 전체 이동 A와 구간 분리 B는 현재 시험이 잡았으며, 원본의 집중 `-race` 검사도 통과했다."

이 판정은 오케스트레이터의 결과 확인(`moai verify audit-plan --result-file`)을 거치기 전까지 최종이 아니다.

## Baseline-attribution

- 모든 시험 측정은 이 실행에서 HEAD `9aa451b107b6e59db0fad5a1e1c0d21950fcbef4` 트리(및 그 `git archive`에서 뽑은 변형 사본)를 `go test`로 컴파일해 얻었다. 시험 바이너리는 측정 대상 트리에서 빌드됐다.
- 변형 사본의 시험 파일은 HEAD와 같다(`mutate.py`는 `internal/cli/local_main_resync.go`만 바꾼다).
- `moai verify audit-plan`과 MCP 서버는 설치된 빌드 `ee981a932`(HEAD의 조상)로 실행됐다. audit_multi 결과의 `build_lag`가 이를 명시한다. 감사 계획과 게이트 판정은 그 빌드가 읽은 것이다.

## Gaps

- Windows에서 시험을 실행하지 않았다. F2와 C5의 Windows 관련 판단은 `state_lock_windows.go` 코드 읽기에 근거한 추론이다. Windows 다리는 release multi-OS CI에서만 돈다.
- codex가 보고한 "읽기 직후 경쟁자가 창을 탈취했는데도 fast-forward가 실행됐다"는 별도 probe는 이 감사가 재현하지 않았다. mutant D가 시험을 통과한다는 사실(E1)은 직접 측정했다.
- `moai` CLI·MCP 서버 빌드가 HEAD보다 뒤처져 있다(`ee981a932`). 감사 계획 판독은 지연된 빌드의 출력이다. 이 카드의 델타는 그 도구 경로를 바꾸지 않는다.
- 전체 `internal/cli` 패키지 시험과 저장소 전체 시험은 실행하지 않았다(영향 범위 시험만 실행).
- 고루틴 누수는 goleak 같은 도구로 측정하지 않고, 코드 읽기와 `-race` 통과로 판단했다.

## Residual-risk

- F1을 고친 뒤에도 seam 변수를 우회해 `factory.ReadIntegrationLock`을 직접 호출하는 변형은 잡히지 않는다. 변형 시험은 그럴듯한 회귀를 고정할 뿐 완전하지 않다.
- 잠금을 지키지 않는 외부 행위자의 비원자적 abort 창은 이전 회차에서 수용된 잔여 위험으로 남는다.
- 경쟁자-HEAD 신호는 타이밍 의존이라, 부하가 높은 CI에서 E 형태 회귀를 놓칠 수 있다. E는 기존 시험(45행)이 결정적으로 잡으므로 실질 위험은 낮다.
- 실행 로그는 `/private/tmp/t70-audit3/logs/`(머신 로컬)에 있으며 인용 근거는 이 파일에 옮겨 적은 줄뿐이다. 그 디렉터리는 정리되면 사라진다.

## Iteration history

| 회차 | 감사 SHA | 판정 | 핵심 |
|---|---|---|---|
| 1 | `a2055ca46` | FAIL 72 | 시험이 재검증의 잠금 내 배치를 고정하지 못함(mutant A 통과) |
| 2 | `b79a70d97` | FAIL 78 | A는 잡지만 B(두 구간 분리) 통과 |
| 3 | `9aa451b10` | FAIL 82 | A·B·C·E는 잡지만 D(소유권 읽기만 구간 앞으로) 통과, codex required fail |
