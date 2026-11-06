auditor-model: claude-opus-5-5
verdict: FAIL
audited_sha: b79a70d9792a7926e50300c7f738715835bf54d1

## Evaluation Report — 카드 t1670 sync-audit 2회차 (델타 재감사)

SPEC: SPEC-LOCAL-MAIN-FLOW-001
Overall Verdict: FAIL — overall score 78/100

감사 대상: 워크트리 `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1670`, 브랜치 `WT-t1616-10-10`, HEAD `b79a70d97`, 기준 local main `67e462e80`. 판정 대상 수리 커밋은 `6a2d5cee1`이다.

### 요약

수리는 1회차에서 지적한 변형(mutant A: 재검증을 잠금 구간 밖으로 이동)을 정확히 잡아낸다. 실제 트리에서 PASS, mutant A에서 FAIL을 이 감사가 직접 재현했다. 운영 코드 변경은 아무 일도 하지 않는 seam 호출 한 줄뿐이고, binlag 좌표 수정(444/447)도 맞다.

그런데 두 번째 변형(mutant B)은 잡지 못한다. 재검증은 잠금 안에 두되, 재검증과 merge를 **별도의 `RunIntegrationMutation` 호출 두 번**으로 쪼갠 변형에서 새 시험과 기존 시험이 모두 통과했다. 이 변형은 구간 1이 끝나고 구간 2가 열리기 전 사이에 F2 창을 그대로 다시 연다. 카드의 첫 번째 완료 조건은 "재검증과 이동이 **같은** 락 구간에서 일어남을 시험으로 확인"이다. 지금 시험이 증명하는 것은 "재검증 시점에 잠금이 잡혀 있다"까지이고, "구간이 merge까지 끊기지 않는다"는 증명하지 못한다. 필수 백엔드 codex도 독립적으로 같은 변형을 만들어 같은 결론을 냈다(verdict fail, P2, receipt `rcpt-a9e7f20d1d938b840d05666e`). 필수 게이트가 FAIL이므로 이 감사는 PASS를 낼 수 없다.

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%, must-pass) | 60/100 | FAIL | 운영 코드 경로는 맞다(재검증과 merge가 한 구간 안에 있음, E1). mutant A는 이제 잡힌다(E2). 그러나 mutant B에서 `--- PASS: TestLocalMainResyncRivalTakeoverDuringRevalidateCannotLand (8.71s)` (E3) — 완료 조건 1이 여전히 시험으로 고정되지 않았다. 필수 통과 차원이라 이것만으로 전체가 FAIL이다 |
| Security (25%, must-pass) | 95/100 | PASS | 추가된 것은 패키지 내부 비공개 함수 변수(`localMainResyncAfterRevalidate`, 기본값 no-op) 하나다. 외부 입력, 비밀값, 셸 주입면이 없다. grep 결과 사용처는 정의(60/66행), 호출(335행), 시험(63/64/76행)뿐이다(E5) |
| Craft (20%) | 85/100 | PASS | resync 시험 15건 PASS / 0 FAIL(E1), `-race -count=3` PASS(E4), `gofmt -l` 출력 없음, `go vet` exit 0(E5). 다만 새 시험은 경쟁자 오류가 "busy"인지까지는 확인하지 않는다(F2, 선택 사항) |
| Consistency (15%) | 90/100 | PASS | seam 이름, @MX:NOTE/@MX:SPEC 주석, `t.Cleanup` 복원 방식이 기존 `localMainResyncBeforeFastForward` / `AfterFastForward` seam과 같은 모양이다 |

가중 합: 0.40×60 + 0.25×95 + 0.20×85 + 0.15×90 = 78.25 → 78.

### Findings (structured defect-list)

- **F1 [P2] [blocking]** `internal/cli/local_main_resync_gap_test.go:64` (및 `internal/cli/local_main_resync.go:246-257`) — 새 시험은 경쟁자의 `--force` 탈취를 seam 안에서 **동기적으로** 시도한다. 그래서 재검증 구간이 끝난 뒤 merge 전에 잠금을 풀었다 다시 잡는 회귀(mutant B)에서는 탈취 시도가 이미 끝난 상태라 빈틈에 경쟁자가 없다. 실측 결과 mutant B에서 두 rival 시험 모두 PASS(E3). codex도 같은 변형으로 같은 결과를 보고했다(E6). 확신도 높음.
  - Required fix: 경쟁자를 seam에서 **고루틴으로** 띄워서 잠금이 풀리는 즉시 탈취하도록 만들고, 탈취가 성공했다면 그 순간 HEAD가 이미 origin/main 대상 커밋이어야 한다고 단언한다. 즉 "탈취 성공 시점의 HEAD == target 이거나 탈취 실패"를 확인한다. 고루틴 종료를 채널로 기다린 뒤 판정한다. 실제 코드에서는 경쟁자가 merge가 끝날 때까지 flock에서 기다리므로 성공하더라도 HEAD가 이미 target이다. mutant B에서는 구간 1이 풀리는 순간 탈취해서 HEAD가 아직 base이므로 실패한다. 수리 후에는 mutant A와 mutant B 둘 다 FAIL하는지 측정해서 evidence에 남긴다. (주의: 실제 코드에서도 merge 이후 release 전에 경쟁자가 창을 가져갈 수 있으므로, re-sync 반환 오류에 대한 단언은 그 경우를 허용해야 한다.)
- **F2 [low] [optional]** `internal/cli/local_main_resync_gap_test.go:83` — `rivalErr == nil`만 확인한다. 입력 검증 실패 같은 다른 오류로도 시험이 통과할 수 있다. mutant A에서 같은 입력의 탈취가 실제로 성공했으므로(E2) 지금은 판별력이 있다. `factory.IsIntegrationLockBusy(rivalErr)` 단언을 더하면 의도가 더 분명해진다. 확신도 중간.
- **F3 [low] [optional]** 시험 시간 — 두 rival 시험이 각각 잠금 대기 예산을 다 쓰고 약 8.7초씩 걸린다(E1). 결정적이고 flaky하지 않지만(`-race -count=3` PASS, E4) 패키지 시험 시간에 약 17초를 더한다. 확신도 높음.

### Recommendations

- F1만 고치면 된다. 시험 하나를 위와 같이 고루틴 방식으로 바꾸거나 하나 더 추가하고, 운영 코드는 손대지 않는다(현재 운영 코드는 맞다).
- 3회차 재감사 범위는 F1 델타(시험 수정 + mutant A/B 재측정)로 한정한다.

---

## Claim

- C1. 실제 트리(HEAD `b79a70d97`)에서 새 시험 `TestLocalMainResyncRivalTakeoverDuringRevalidateCannotLand`와 resync 시험 15건 전체가 통과한다.
- C2. 새 시험은 mutant A(재검증을 잠금 구간 밖으로 이동)에서 실패한다. 1회차 blocker가 지적한 변형은 이제 잡힌다.
- C3. 새 시험은 mutant B(재검증은 잠금 안, 그러나 재검증과 merge를 서로 다른 두 `RunIntegrationMutation` 구간으로 분리)에서 **통과한다**. 기존 시험도 통과한다. 따라서 "같은 락 구간" 속성은 아직 시험으로 고정되지 않았다.
- C4. 운영 코드 변경은 기본값이 no-op인 seam 변수 정의와 `localMainResyncRevalidate` 성공 경로 끝의 호출 한 줄뿐이다.
- C5. binlag 좌표 444/447은 맞다. `TestAuditLagUsesBinlagSeam` PASS, 444행은 doc 언급, 447행은 `merge-base --is-ancestor` 호출이다.
- C6. 새 시험은 race detector 아래 3회 반복해도 안정적이다.
- C7. 카드 완료 조건: (1) "같은 락 구간을 시험으로 확인"은 **미충족**(C3). (2) "외부 행위자 경합 재현 0회"는 잠금을 지키는 행위자에 한해 충족한다. mutant A가 실패하고 실제 트리가 통과하므로 잠금을 지키는 경쟁자는 재검증과 merge 사이에 끼어들지 못한다. 잠금을 무시하는 행위자에 대한 비원자적 abort 창은 1회차에서 수용된 잔여 위험 그대로다.
- C8. 필수 백엔드 codex의 교차 감사 결과는 `fail`이다(gate required).

## Evidence

E1 — 실제 트리, resync 시험 전체 (`TMPDIR=/private/tmp/t70-audit2 GOTMPDIR=/private/tmp/t70-audit2 go test ./internal/cli/ -run TestLocalMainResync -count=1 -v > /private/tmp/t70-audit2/real.log 2>&1; echo EXIT=$?`):

```
EXIT=0
--- PASS: TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand (8.75s)
--- PASS: TestLocalMainResyncRivalTakeoverDuringRevalidateCannotLand (8.72s)
--- PASS: TestLocalMainResyncFastForwards (1.04s)
--- PASS: TestLocalMainResyncRefusesDiverged (0.39s)
--- PASS: TestLocalMainResyncAheadIsNoop (0.32s)
--- PASS: TestLocalMainResyncPreservesIgnoredFile (0.38s)
--- PASS: TestLocalMainResyncHolderCallKeepsWindow (0.77s)
--- PASS: TestLocalMainResyncTakeoverDuringFetchRefuses (0.41s)
--- PASS: TestLocalMainResyncPostMergeWritesHold (0.45s)
--- PASS: TestLocalMainResyncKeepsCardMetadata (0.34s)
--- PASS: TestLocalMainResyncLateDirtyPrimaryRefuses (0.41s)
--- PASS: TestLocalMainResyncFetchPinsFetchedCommit (0.60s)
--- PASS: TestLocalMainResyncForeignFetchDoesNotMoveBaseline (0.52s)
--- PASS: TestLocalMainResyncHoldWriteFailureKeepsWindow (0.42s)
--- PASS: TestLocalMainResyncPostFastForwardLateTxtRefuses (0.42s)
ok  	github.com/modu-ai/moai-adk/internal/cli	24.845s
```

PASS 15건, FAIL 0건.

E2 — mutant A (`/private/tmp/t70-audit2/mutA`, 이 감사가 직접 만듦: `git -C <worktree> archive -o /private/tmp/t70-audit2/head.tar HEAD` → 추출 → `localMainResyncRevalidate` 호출을 `RunIntegrationMutation` 앞으로 이동). `cd /private/tmp/t70-audit2/mutA && TMPDIR=... GOTMPDIR=... go test ./internal/cli/ -run 'TestLocalMainResyncRivalTakeover' -count=1 -v`:

```
EXIT=1
--- PASS: TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand (8.78s)
    local_main_resync_gap_test.go:83: a rival --force takeover landed right after the ownership re-check (re-sync err = integration resync: local main fast-forwarded ec20210697b2 -> 35aade310f59 to origin/main, but releasing the window failed: release integration window is held by a different session: lane-rival (pid 50769) holds it (moai integration status reads it)): the re-validation does not share the move's mutation section
--- FAIL: TestLocalMainResyncRivalTakeoverDuringRevalidateCannotLand (0.42s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	10.214s
```

mutant A diff (실제 파일 대비):

```
246a247
> 	revalidateErr = localMainResyncRevalidate(git, repoRoot, sessionID, ref, head)
248c249
< 		if revalidateErr = localMainResyncRevalidate(git, repoRoot, sessionID, ref, head); revalidateErr != nil {
---
> 		if revalidateErr != nil {
```

E3 — mutant B (`/private/tmp/t70-audit2/mutB`, 재검증은 첫 번째 `RunIntegrationMutation` 안, `localMainResyncBeforeFastForward` + merge는 두 번째 `RunIntegrationMutation` 안). 같은 명령:

```
EXIT=0
--- PASS: TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand (8.75s)
--- PASS: TestLocalMainResyncRivalTakeoverDuringRevalidateCannotLand (8.71s)
ok  	github.com/modu-ai/moai-adk/internal/cli	18.383s
```

mutant B diff (실제 파일 대비):

```
248c248,254
< 		if revalidateErr = localMainResyncRevalidate(git, repoRoot, sessionID, ref, head); revalidateErr != nil {
---
> 		revalidateErr = localMainResyncRevalidate(git, repoRoot, sessionID, ref, head)
> 		return nil
> 	}); lockErr != nil {
> 		return "", lockErr
> 	}
> 	if lockErr := factory.RunIntegrationMutation(repoRoot, func() error {
> 		if revalidateErr != nil {
```

원인: 경쟁자의 탈취는 seam 안에서 동기적으로 호출되므로 구간 1이 끝나기 전에 busy로 끝난다. 구간 1과 구간 2 사이의 빈틈에서는 아무도 탈취를 시도하지 않는다.

E4 — 안정성 (`TMPDIR=... GOTMPDIR=... go test ./internal/cli/ -race -run TestLocalMainResyncRivalTakeover -count=3`):

```
EXIT=0
ok  	github.com/modu-ai/moai-adk/internal/cli	56.256s
```

E5 — binlag·형식·사용처:

```
$ go test ./internal/cli/ -run TestAuditLagUsesBinlagSeam -count=1 -v   (TMPDIR/GOTMPDIR 동일)
--- PASS: TestAuditLagUsesBinlagSeam (0.20s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.921s

$ gofmt -l internal/cli/local_main_resync.go internal/cli/local_main_resync_gap_test.go internal/cli/mcp_build_identity_test.go
(출력 없음) GOFMT_EXIT=0
$ go vet ./internal/cli/
VET_EXIT=0

$ sed -n '444p;447p' internal/cli/local_main_resync.go
// included. `git merge-base --is-ancestor` exits 1 for "not an ancestor". Any
	_, err := git("merge-base", "--is-ancestor", a, b)

$ grep -n 'localMainResyncAfterRevalidate' internal/cli/*.go
internal/cli/local_main_resync.go:60:// localMainResyncAfterRevalidate is the end-of-re-validation test seam (card t1670,
internal/cli/local_main_resync.go:66:var localMainResyncAfterRevalidate = func(repoRoot string) {}
internal/cli/local_main_resync.go:335:	localMainResyncAfterRevalidate(repoRoot)
internal/cli/local_main_resync_gap_test.go:63:	prev := localMainResyncAfterRevalidate
internal/cli/local_main_resync_gap_test.go:64:	localMainResyncAfterRevalidate = func(repoRoot string) {
internal/cli/local_main_resync_gap_test.go:76:	t.Cleanup(func() { localMainResyncAfterRevalidate = prev })
internal/cli/mcp_build_identity_test.go:793:		// Re-measured at card t1670: the localMainResyncAfterRevalidate test
```

`git show 6a2d5cee1 -- internal/cli/local_main_resync.go`의 운영 코드 변경은 +8행(seam 정의와 주석) +1행(335행 호출)이 전부다.

E6 — 교차 모델 감사. `moai verify audit-plan --project-root <worktree>` 결과 `config_status: "ok"`, `model: "codex"`, codex `gate: "required"`, claude/glm `off`, `cross_model_active: true`. `mcp__moai__audit_multi`(project_root=worktree, target=baseBranch, card_id=t1670, gates 인자 없음) 결과 digest:

```
overall_verdict: fail
gate_unmet: (응답에 별도 항목 없음; residual_risk_note: "required-backend FAIL: codex")
plan_source: config
per_backend_verdicts: [{backend: codex, gate: required, verdict: fail}]
audit_receipt: rcpt-a9e7f20d1d938b840d05666e
```

codex finding 원문(발췌): "[P2] 재검증과 merge 사이에서 잠금을 해제하는 회귀를 테스트가 놓친다 — internal/cli/local_main_resync_gap_test.go:64 … 재검증과 merge를 각각 별도의 `RunIntegrationMutation` 호출로 분리한 변형에서도 두 rival 테스트가 모두 통과했다(`PASS`, 18.643s)."

이 판정은 오케스트레이터가 위 digest를 `.moai/state/audit-plan-result.json`에 기록하고 `moai verify audit-plan --result-file`로 확인하기 전까지 최종이 아니다. 다만 결과가 이미 fail이므로 그 확인이 판정을 PASS로 바꿀 수는 없다.

## Baseline-attribution

- 모든 Go 시험은 이번 실행에서 워크트리 HEAD `b79a70d9792a7926e50300c7f738715835bf54d1`(작업 트리 clean)에서 빌드됐다. mutant 두 개는 같은 HEAD의 `git archive`에서 각각 한 블록씩만 바꾼 사본이다(E2/E3 diff).
- 시험 바이너리는 `go test`가 트리에서 직접 빌드했으므로 §2.2 도구 출처 문제는 해당하지 않는다.
- `moai verify audit-plan`과 audit_multi 서버는 설치된 바이너리(`v3.2.0-rc.36-t1822c`, 빌드 커밋 `ee981a932`)로 실행됐다. audit_multi 응답의 `build_lag`는 `ee981a932`가 이 트리 HEAD의 조상이라고 보고한다. 이 바이너리는 감사 계획을 읽고 codex에 diff를 넘기는 데만 쓰였고, 판정 근거인 시험 결과에는 관여하지 않는다.

## Gaps

- 저장소 전체 시험은 돌리지 않았다(지시대로 영향받는 시험만 실행). `internal/cli`의 다른 시험과 다른 패키지는 이번 실행에서 관측하지 않았다.
- golangci-lint와 커버리지는 측정하지 않았다. 형식과 정적 검사는 `gofmt -l`과 `go vet`으로 대신했다.
- audit_multi 응답에 `gate_unmet` 필드가 따로 들어 있지 않았다. 필수 게이트 미충족은 `overall_verdict: fail`과 `residual_risk_note`로만 읽었다.
- 판정 통과 여부를 가리는 오케스트레이터의 `--result-file` 확인은 이 감사가 실행하지 않았다(읽기 전용 역할).
- 실행 도중 워크트리 격리 가드가 복합 명령 두 건을 거부했다(복잡한 `git archive` 파이프라인과, 인자에 `|`가 들어간 `go test -run` 패턴). 둘 다 단순한 개별 명령으로 나눠 다시 실행했고, 결과는 위 증거 그대로다. 측정을 원본 읽기로 대신한 부분은 없다.
- audit_multi가 `card_id`에 따라 `.moai/reports/t1670/second-review.jsonl`을 썼다. 이 파일은 이번 커밋에 넣지 않았다.

## Residual-risk

- 운영 코드 자체는 맞다(재검증과 merge가 한 구간 안). 지금 남은 위험은 앞으로 누군가 mutant B 같은 모양으로 리팩터링해도 시험이 잡지 못한다는 점이다.
- 잠금을 무시하는 외부 행위자에 대한 비원자적 `merge --abort` 창은 1회차에서 수용된 잔여 위험으로 그대로 남아 있다.
- 실제 코드에서도 merge가 끝난 뒤 window를 놓기 전에는 경쟁자가 창을 가져갈 수 있다. 이 경우 re-sync가 "releasing the window failed"로 끝난다. F1 수리 시험은 이 정상 경합을 실패로 오판하지 않도록 설계해야 한다.
- scratch 경로 `/private/tmp/t70-audit2/*.log`의 전체 로그는 머신 로컬이다. 판정을 가른 줄은 위 Evidence에 옮겨 두었고, 나머지는 보존되지 않는다.

### 반복 이력

| 회차 | 감사한 SHA | 판정 | 점수 | 차단 사유 |
|---|---|---|---|---|
| 1 | a2055ca46 | FAIL | 72 | 기존 시험이 mutant A(재검증을 잠금 밖으로 이동)를 잡지 못함 |
| 2 | b79a70d97 | FAIL | 78 | mutant A는 잡힘. mutant B(구간 둘로 분리)는 여전히 잡히지 않음. codex 필수 게이트 fail |
