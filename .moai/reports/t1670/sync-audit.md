auditor-model: claude-opus-5-5

---
verdict: FAIL
audited_sha: a2055ca463d4bbabad5070da6185ec24ebd26618
card: t1670
base: 67e462e80 (local main)
blocking_count: 1
required_backend: codex
required_backend_status: fail
engine_receipt: rcpt-1308cccab8ab13cc777166a9
engine: moai/audit_multi, gates from the tree configuration (no gates argument), plan_source=config
judging_build: MCP server / installed moai v3.2.0-rc.36-t1822c commit ee981a932 (dirty), an ANCESTOR of a2055ca46 — build lag, see Gaps 3
---

# 카드 t1670 동기화 감사 — evidence-only 착지 주장에 대한 적대적 검증

Overall Verdict: FAIL — overall score 72/100

## 종합 판정

FAIL이다. 코드 자체는 맞다. 현재 트리에서 재검증과 `git merge --ff-only`가 같은 `factory.RunIntegrationMutation` 구역 안에 있고, F1 same-SHA abort 소유권도 닫혀 있다. 그런데 이 카드의 첫 번째 완료 조건인 "재검증과 이동이 같은 락 구간에서 일어남을 시험으로 확인"은 충족되지 않았다. 레인이 근거로 든 `TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand`는 재검증을 잠금 밖으로 빼낸 변형(mutant A)에서도 통과한다. 이 감사에서 직접 돌려 확인했고, 필수 백엔드 codex도 별도의 overlay로 같은 결과를 냈다(verdict fail, P2). 따라서 "코드 변경 없음, evidence-only 착지"는 맞는 처리가 아니다. 판별력 있는 시험 하나(시험용 seam 한 줄 포함)를 추가해야 한다.

`local_main_resync.go:228`(t1616 감사 3f3504fb7이 따로 적은 잔여)은 F2와 같은 창이다. 별개의 창이 아니다. 근거는 아래 증거 E1에 있다.

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 55/100 | FAIL | 코드 경로는 맞다(E3, E6). 다만 완료 조건 1을 증명한다고 인용된 시험이 mutant A를 잡지 못한다. 재검증 이름이 붙은 cli 시험 14건이 모두 mutant A에서 통과한다(E5). 필수 통과 차원이므로 이 차원만으로 전체가 FAIL이 된다 |
| Security (25%) | 95/100 | PASS | 카드 차분은 `.moai/reports/t1670/evidence.md` 30줄 추가뿐이고(`git diff --stat 67e462e80 HEAD`), 실행 코드 변경은 0줄이다. 비밀값이나 입력 표면 변경 없음 |
| Craft (20%) | 70/100 | FAIL | 이름 붙은 시험은 모두 통과한다(E2, E4). 하지만 시험의 판별력이 부족하다. 경쟁 획득은 재검증이 끝난 뒤의 seam에서만 시도되므로, 시험이 고정하는 것은 "seam부터 이동까지가 잠금 안"이라는 성질뿐이다 |
| Consistency (15%) | 85/100 | PASS | 커밋 제목이 Conventional Commit 형식이고 `(card t1670)`를 달고 있다. 카드 id가 들어간 증거 경로를 쓴다. 덮는 변경이 merge 커밋 안에서 들어왔다는 점은 선택 항목 F3으로 따로 적었다 |

가중 점수: 0.40×55 + 0.25×95 + 0.20×70 + 0.15×85 = 72.5에서 내림해 72/100. 평면 가중 방식이다. 필수 통과 차원인 Functionality가 실패했으므로 점수와 상관없이 FAIL이다.

## 결함 목록 (structured defect-list)

- **F1** [High] [blocking] `internal/cli/local_main_resync_gap_test.go:22-34`(seam 호출 지점 `internal/cli/local_main_resync.go:245`) — 시험이 재검증 위치를 판별하지 못한다. 경쟁자의 `--force` 획득은 재검증이 이미 끝난 `localMainResyncBeforeFastForward`에서만 시도된다. 그래서 `localMainResyncRevalidate`를 `RunIntegrationMutation` 앞(잠금 밖)으로 옮기고 seam과 merge만 잠금 안에 남긴 변형도 통과한다. 그 변형에서는 재검증이 끝난 뒤 잠금을 잡기 전까지 경쟁자가 창을 가져갈 수 있다. 이것이 바로 F2 창이다. 레인 evidence의 `.moai/reports/t1670/evidence.md:11`("F2 test — re-validation and move in the same lock section")은 과대 주장이다.
  필요한 수정(최소):
  1. `internal/cli/local_main_resync.go`에 seam 하나를 추가한다. 예: `var localMainResyncDuringRevalidate = func(repoRoot string) {}`. `localMainResyncRevalidate` 안에서 소유자·lease 판독 뒤, 마지막 `return nil`(현재 `:327`) 직전에 호출한다.
  2. `TestLocalMainResyncRivalTakeoverDuringRevalidateCannotLand`를 추가한다. 기존 gap 시험과 같은 모양으로, 그 seam에서 `factory.AcquireIntegrationLock(..., true)`를 시도하고 `rivalErr != nil`, re-sync 성공, `HEAD == target`을 단언한다.
  3. 새 시험이 트리에서는 PASS이고, 아래 mutant A overlay(`-overlay`)에서는 FAIL임을 관측해 evidence.md에 기록한다.
  4. evidence.md:11의 주장을 고친다.

  이것이 남은 유일한 차단 항목이다.
- **F2** [Low] [optional] 덮는 변경(`factory.RunIntegrationMutation` 호출)은 merge 커밋 `cd19e5494`("merge: absorb main ec193e4777 into WT-moai-update-repair (card t1619)")에서 처음 나타나고, 두 부모 `8b0beadf1`·`ec193e477` 어디에도 없다(E7). 동작에는 영향이 없다. 추적성 문제로만 적어 둔다. 고칠 것 없음, 기록만.
- **F3** [Low] [optional] 경쟁 획득은 mutation lock의 busy 예산 전체를 기다린 뒤에 거절된다(gap 시험 8.80s). F1 수정으로 같은 모양의 시험이 하나 더 생기면 cli 시험 시간이 약 9s 늘어난다. 내부 짧은 예산 seam이 있으면 그것을 쓰라는 제안일 뿐이다.
- **F4** [Info] [optional] 감사 도중 이 트리의 `.moai/reports/t1670/evidence.md`에 커밋되지 않은 수정이 생겼다. "Stage record"와 "Lease expiry" 두 줄이다. HEAD는 `a2055ca46` 그대로였다. 내용으로 보아 트리를 소유한 레인 세션의 기록이고, 다른 세션의 커밋은 아니다. 이 감사는 그 파일을 커밋하지 않았다. 판정은 커밋 `a2055ca46`의 내용을 기준으로 한다.

## 1. Claim

- C1. `local_main_resync.go:228`(3f3504fb7 기준)은 F2와 같은 창이다. 별개의 잔여가 아니다.
- C2. 현재 트리(a2055ca46, 코드는 67e462e80과 같음)에서 재검증과 ff 이동은 같은 mutation 구역 안에 있다. F2의 코드 결함은 닫혔다.
- C3. F1 same-SHA abort 소유권은 닫혔다. abort는 이 호출이 exit 1을 받았을 때만, 같은 mutation 구역 안에서 실행된다.
- C4. 완료 조건 1("같은 락 구간을 시험으로 확인")은 충족되지 않았다. 인용된 시험은 재검증을 잠금 밖으로 옮긴 변형을 잡지 못한다.
- C5. 따라서 evidence-only 착지는 맞지 않다. 판별 시험 하나를 구현해야 한다. 판정은 FAIL이다.

## 2. Evidence

**E1 — `:228`의 정체(C1).** `git show 3f3504fb7:internal/cli/local_main_resync.go`의 해당 줄:

```
226		// The re-validation (card t1616, F2 and F5), immediately before the move. The
227		// holder check and the clean check ran earlier, and the fetch ran in between.
228		if err := localMainResyncRevalidate(git, repoRoot, sessionID, ref, head); err != nil {
229			return "", err
230		}
231		// The --no-overwrite-ignore flag is the second guard (plan §B4). Autostash is
232		// disabled so the verb never stashes another session's changes (plan §B5).
233		if _, err := git("-c", "merge.autoStash=false", "merge", "--ff-only", "--no-overwrite-ignore", "-q", baseline); err != nil {
```

t1616 보고서에서 출처를 찾은 결과(`grep -rn "local_main_resync.go:22[0-9]" .moai/reports/t1616/`):

```
.moai/reports/t1616/card-review.md:9:1. **[P1]** `internal/cli/local_main_resync.go:228`. 소유권 재검증과 fast-forward(`merge --ff-only`)를 같은 공유 mutation lock 아래에 묶지 않았다. ...
.moai/reports/t1616/card-review.md:34:  1. **[P1]** `internal/cli/local_main_resync.go:228`. ... 1차 발견 1과 같은 창이며, SPEC progress §E.4의 수용 잔여 **F2**다. 후속 카드 **t1670** 범위다(리더 지시).
.moai/reports/t1616/sync-audit-d409307cc.md:122:| F2 | ACCEPTED-RESIDUAL | ... | 재검증 `internal/cli/local_main_resync.go:228`과 `merge --ff-only` `:233`이 잠금 구역 밖에 있다. ...
```

`:228`은 codex 카드 리뷰 P1이 가리킨 위치이고, 같은 문서가 F2와 같은 창이라고 명시한다. 3f3504fb7 감사가 셋을 따로 적은 것은 출처별로 나열했기 때문이다. 별개의 창이 아니다. 같은 시점의 무시 바이트 중첩 탐침(3f3504fb7 `:215-225`, 현재 `:221-231`)은 잠금 밖에서 돌지만 잔여 목록에 오른 적이 없다. ff 경로에서는 `--no-overwrite-ignore`가 두 번째 방어선이 된다(Residual-risk 참고).

**E2 — 이름 붙은 cli 시험(원본 트리).** `TMPDIR=/private/tmp/t70-audit GOTMPDIR=/private/tmp/t70-audit go -C <tree> test -count=1 -v -run 'TestLocalMainResync' ./internal/cli/` → `exit=0`:

```
--- PASS: TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand (8.80s)
--- PASS: TestLocalMainResyncFastForwards (0.41s)
--- PASS: TestLocalMainResyncRefusesDiverged (0.40s)
--- PASS: TestLocalMainResyncAheadIsNoop (0.37s)
--- PASS: TestLocalMainResyncPreservesIgnoredFile (0.41s)
--- PASS: TestLocalMainResyncHolderCallKeepsWindow (0.85s)
--- PASS: TestLocalMainResyncTakeoverDuringFetchRefuses (0.48s)
--- PASS: TestLocalMainResyncPostMergeWritesHold (0.56s)
--- PASS: TestLocalMainResyncKeepsCardMetadata (0.43s)
--- PASS: TestLocalMainResyncLateDirtyPrimaryRefuses (0.51s)
--- PASS: TestLocalMainResyncFetchPinsFetchedCommit (0.72s)
--- PASS: TestLocalMainResyncForeignFetchDoesNotMoveBaseline (0.61s)
--- PASS: TestLocalMainResyncHoldWriteFailureKeepsWindow (1.17s)
--- PASS: TestLocalMainResyncPostFastForwardLateTxtRefuses (0.49s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	17.228s
```

**E3 — 현재 코드(C2).** `internal/cli/local_main_resync.go:238-256`:

```
238		var revalidateErr, moveErr error
239		if lockErr := factory.RunIntegrationMutation(repoRoot, func() error {
240			if revalidateErr = localMainResyncRevalidate(git, repoRoot, sessionID, ref, head); revalidateErr != nil {
241				return nil
242			}
...
245			localMainResyncBeforeFastForward(repoRoot)
246			_, moveErr = git("-c", "merge.autoStash=false", "merge", "--ff-only", "--no-overwrite-ignore", "-q", baseline)
247			return nil
248		}); lockErr != nil {
```

`RunIntegrationMutation`은 `withIntegrationLockMutation`을 그대로 부른다(`internal/factory/integration_lock_mutation.go:95-97`). 경쟁 경로 `AcquireIntegrationWindow`도 같은 구역을 잡는다(`internal/factory/integration_lock.go:369`, `if mutErr := withIntegrationLockMutation(projectRoot, func() error {`). `ReleaseIntegrationLock`도 마찬가지다(`:545`). 그러니 잠금을 지키는 행위자는 재검증과 이동 사이에 끼어들 수 없다.

**E4 — F1 시험(원본 트리).** `go -C <tree> test -count=1 -v -run '<6 names>' ./internal/factory/` → `exit=0`:

```
--- PASS: TestMergeStepPreexistingMergeHeadRefusesWithoutAbort (0.35s)
--- PASS: TestMergeStepRaceMergeHeadAfterProbeIsNotAborted (0.41s)
--- PASS: TestMergeStepRaceSameShaMergeHeadAfterProbeIsNotAborted (0.47s)
--- PASS: TestMergeStepStartFailureSameShaMergeIsNotAborted (0.44s)
--- PASS: TestMergeStepStartFailureWithoutMergeHeadRefusesWithoutAbort (0.42s)
--- PASS: TestMergeStepHookStoppedOwnMergeIsAbortedCause6 (0.70s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/factory	3.203s
```

이름 위치는 레인이 적은 대로다: `integration_merge_step_overlap_test.go:415, 447, 486, 552`(`grep -n "^func Test"`).

**E5 — 변이 시험(C4, 결정 근거).** 원본 블록(`:238-248`)을 바꾼 overlay 세 개를 `go test -overlay`로 실행했다. 생성 스크립트와 overlay JSON은 scratchpad의 `mkmut.py`, `mut/overlay_{A,B,C}.json`에 있다.

- mutant A: 재검증은 잠금 밖, seam과 merge는 잠금 안. → `mutant A exit=0`
  ```
  --- PASS: TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand (8.76s)
  --- PASS: TestLocalMainResyncTakeoverDuringFetchRefuses (0.47s)
  ok  	github.com/modu-ai/moai-adk/internal/cli	10.178s
  ```
  같은 변형으로 `-run 'TestLocalMainResync|TestIntegrationResync|Resync'` 전체를 돌린 결과: `mutant A family exit=0`, `--- PASS` 14건, `ok  	github.com/modu-ai/moai-adk/internal/cli	16.023s`. 이 변형을 잡는 시험은 하나도 없다.
- mutant B: 재검증과 seam은 잠금 밖, merge만 잠금 안. → `mutant B exit=1`
  ```
  local_main_resync_gap_test.go:38: a rival --force takeover landed between the ownership re-check and the fast-forward (re-sync err = integration resync: local main fast-forwarded b24dc638f535 -> f5acde165fa4 to origin/main, but releasing the window failed:
  --- FAIL: TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand (0.45s)
  ```
- mutant C: 잠금 없음(t1616 시점의 모양). → `mutant C exit=1`
  ```
  local_main_resync_gap_test.go:38: a rival --force takeover landed between the ownership re-check and the fast-forward (re-sync err = integration resync: local main fast-forwarded 591e901b98b5 -> 8360e7ac1469 to origin/main, but releasing the window failed:
  --- FAIL: TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand (0.41s)
  ```

결론: 이 시험은 "seam부터 merge까지가 잠금 안"이라는 성질만 고정한다. "재검증과 이동이 같은 구역"이라는 성질은 고정하지 않는다.

**E6 — F1 abort 위치(C3).** `internal/factory/integration_merge_step.go:415`에서 `sectionErr := withIntegrationLockMutation(in.Root, func() error {`로 구역이 시작된다. `mergeInSection`은 그 안의 `:532`에서 정의되고, 호출은 `:684`와 `:706` 두 곳뿐이며 둘 다 구역 안이다. abort 판단은 `:616-654`에 있다. 다른 SHA를 가리키는 MERGE_HEAD면 거절한다(`:621-623`). exit 1이 아니면 MERGE_HEAD가 있을 때 거절하고(`:634-642`), 시작조차 못 한 호출이면 abort 없이 거절한다(`:643-645`). `git merge --abort`(`:654`)에 도달하는 것은 exit 1인 자기 호출뿐이다. `mergeCallDidNotStop`(`:1199-1212`)에서 `exit.ExitCode == 1`일 때만 `false`를 돌려준다.

**E7 — 덮는 변경의 출처(F2).** `git -C <tree> grep -c 'factory.RunIntegrationMutation' 8b0beadf1 ec193e477 cd19e5494 -- internal/cli/local_main_resync.go`의 출력은 `cd19e5494:internal/cli/local_main_resync.go:1` 한 줄이다. 두 부모에는 없다. `git merge-base --is-ancestor cd19e5494 HEAD` → `ancestor=0`.

**E8 — 교차 모델(필수 codex).** `moai verify audit-plan --project-root <tree>` 결과: `"config_status": "ok"`, `"model": "codex"`, codex `"gate": "required"`, claude·glm `"off"`, `"cross_model_active": true`, `"enforced_required": ["codex"]`. 이어서 gates 인자 없이 `audit_multi`(target baseBranch)를 호출한 결과의 요약 멤버:

```
overall_verdict: fail
gate_unmet: (멤버 없음)
plan_source: config
per_backend_verdicts: [{backend: codex, gate: required, verdict: fail}]
residual_risk_note: required-backend FAIL: codex
audit_receipt: rcpt-1308cccab8ab13cc777166a9
build_commit: ee981a932
```

codex의 지적은 P2 한 건이다. "기존 테스트가 재검증의 잠금 범위를 구별하지 못하므로 evidence-only 완료 판정은 성립하지 않음 — .moai/reports/t1670/evidence.md:11". codex 자체 overlay 실행 결과는 `--- PASS: TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand (8.85s)`였고, 원본 gap 시험과 F1 시험 4건도 통과했다고 보고했다. 이 감사의 판정(F1 차단)과 codex의 판정이 같은 결함에서 일치하므로 불일치는 없다. 이 판정은 오케스트레이터가 `moai verify audit-plan --result-file`로 수렴 점검을 통과시키기 전까지 최종이 아니다.

## 3. Baseline-attribution

- 측정 대상 트리: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1670`, HEAD `a2055ca463d4bbabad5070da6185ec24ebd26618`(첫 명령과 이 파일을 쓰기 직전에 `git rev-parse HEAD`로 확인). 착수 시 `git status --short` 출력은 비어 있었다. 카드 차분은 `git diff --stat 67e462e80 HEAD` → `.moai/reports/t1670/evidence.md | 30 ++++++++++++++++++++++++++++++`, `1 file changed, 30 insertions(+)`. 그러므로 Go 코드는 base 67e462e80과 바이트 단위로 같다.
- Go 시험은 모두 `go -C <tree> test`로 이 트리의 소스를 컴파일해 실행했다(설치된 바이너리를 쓰지 않음). `TMPDIR=GOTMPDIR=/private/tmp/t70-audit`이고 `$HOME` 아래가 아니다. 변이 시험은 트리 파일을 고치지 않고 `-overlay`로만 바꿨다.
- 3f3504fb7 시점의 내용은 `git show 3f3504fb7:<path>`로 읽었다(`git cat-file -t 3f3504fb7` → `commit`).

## 4. Gaps

1. 도구 거절 2회(§3.1). 셸 변수와 heredoc을 섞은 변이 생성 명령, 그리고 반복문으로 돌린 변이 실행이 worktree 격리 가드에 거절됐다. 대신 스크립트를 파일로 쓰고 명령을 하나씩 따로 실행했다. 측정 내용은 바뀌지 않았다. 세 번째 거절(`git show`를 반복문에 넣은 형태)은 `git -C <tree> grep -c <rev>...` 한 줄로 바꿔 같은 질문에 답했다.
2. "잠금을 지키지 않는 외부 행위자" 경합 시험은 쓰지도 돌리지도 않았다. 잠금을 무시하는 행위자를 상대로 "창 재현 0"을 증명하는 시험은 원리상 만들 수 없다. 레인의 이 판단에는 동의하지만, 관측한 것은 아니다.
3. 판정 엔진의 빌드 시차(§2.2). MCP 서버와 설치된 `moai`는 `v3.2.0-rc.36-t1822c` commit `ee981a932`(dirty)이고, `git merge-base --is-ancestor ee981a932 HEAD` → `exit=0`, 즉 HEAD의 엄격한 조상이다. `audit-plan`과 `audit_multi`는 이 오래된 빌드가 판정했다. 결정 근거인 Go 시험은 트리에서 빌드했으므로 이 시차의 영향을 받지 않는다.
4. codex에 넘긴 focus 문구가 mutant A 가설을 명시했다. codex는 자체 overlay로 독립적으로 재현했지만, 질문 자체가 유도적이었다는 점에서 독립성이 약해졌다.
5. `-race` 실행과 패키지 커버리지 측정은 하지 않았다(코드 변경 0줄인 카드).
6. 감사 도중 생긴 evidence.md의 미커밋 두 줄(F4)은 판정 대상이 아니며, 내용 검증도 하지 않았다.

## 5. Residual-risk

- F1 수정 전까지는 재검증을 잠금 밖으로 빼는 회귀가 생겨도 cli 시험 14건이 모두 통과한다. F2 창이 조용히 다시 열릴 수 있다.
- t1616 F1의 비원자적 abort 창(이 감사의 결함 F1과는 다른 항목): exit 1로 멈춘 자기 merge와 `git merge --abort`(`:654`) 사이에, 잠금을 잡지 않는 행위자가 통합 worktree에서 편집하면 abort가 그 편집을 지운다. MoAI 행위자는 모두 같은 mutation 구역을 거치므로 닿을 수 없다. 프로토콜 밖에서만 생기는 수용 잔여이고, git에는 원자적 abort가 없다.
- 무시 바이트 중첩 탐침(`:221-231`)은 잠금 밖에서 돈다. 탐침 뒤에 생긴 무시 파일은 ff 경로의 `--no-overwrite-ignore`가 막는다는 것이 전제다. 그 전제는 `integration_merge_step.go:594-599` 주석의 probe 기록("ff rc=1 refused")에 기대며, 이 감사에서 race 형태로 재현하지는 않았다. `TestLocalMainResyncPreservesIgnoredFile`는 미리 있던 파일만 다룬다.
- 잠금 안의 재검증이 깨끗하다고 본 뒤 프로토콜 밖 행위자가 primary를 더럽히면, ff 뒤의 status 점검(`:276-283`, F5)이 hold로 잡는다. 사후에 잡는 것이고 사전 방지는 아니다.

## 감사 이력

- 1차(이 파일): HEAD a2055ca46, FAIL. 차단 F1 1건(판별 시험 부재), 선택 F2~F4. 재감사는 F1 차분(새 seam과 시험, mutant A에서 FAIL 관측, evidence.md:11 정정)으로 범위를 좁힌다.
