# t592 독립 sync audit — SPEC-HOME-STATE-ROLLOUT-001

## Evaluation Report

SPEC: `SPEC-HOME-STATE-ROLLOUT-001` v0.5.0  
Baseline: branch `WT-home-state-rollout`, HEAD `6ea69661c405c1b6a3ef38e598f5653a63cd7af0`, 현재 미커밋 diff  
Profile: `.moai/config/evaluator-profiles/default.md` (flat weighted-percentage)  
Overall Verdict: **FAIL — 35/100**  
Must-pass firewall: **Functionality 25/100 FAIL, Security 50/100 FAIL**

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|------:|---------|----------|
| Functionality (40%) | 25/100 | FAIL | 공식 25 selector는 exit 0이었으나 auditor overlay 10개가 exit 1로 실제 AC 위반을 재현했다. |
| Security (25%) | 50/100 | FAIL | `go mod verify`는 통과했지만 Factory admission 우회와 profile-clean 보호 공백이라는 High finding이 있다. |
| Craft (20%) | 25/100 | FAIL | coverage allowlist가 실제 변경 production 파일 13개를 제외한다. 보고된 85.047%는 전체 changed surface의 coverage가 아니다. |
| Consistency (15%) | 50/100 | FAIL | `gofmt -d`와 `git diff --check`는 깨끗했지만 tokenless legacy transition과 중복 SessionEnd release가 공존한다. |

가중 점수: `25×0.40 + 50×0.25 + 25×0.20 + 50×0.15 = 35.0`.

## Findings

- **F1 [High] [blocking]** `internal/cli/factory.go:340` — production Factory launcher의 `resolveFactoryWorkerName`이 `ClaimFactoryWorkerName`의 admission 오류를 삼키고 요청 label을 그대로 반환한다(`:341-344`). `cc.go:186`, `glm.go:241`은 그 값을 계속 launch에 사용하므로 active migration marker 아래에서도 lane 시작이 계속된다. **Confidence: High. Impact:** AC-HSR-010/011의 동일 장벽이 실제 Factory 진입점에서 무력화된다. **Required fix:** resolver가 오류를 반환하도록 signature를 바꾸고 cc/glm launch를 중단하며, marker 아래 실제 command가 child 시작과 worker 등록을 모두 0건으로 유지하는 integration test를 추가한다. **Merge-blocking: yes.**

- **F2 [High] [blocking]** `internal/cli/launcher.go:768` — named profile의 `--continue` 성공 경로는 Claude를 `tryCmd.Run()`으로 먼저 시작하고 `:775`에서 반환한다. provisional lease는 `:819-833`에서 그 뒤에 생성되므로 이 경로에서는 생성되지 않는다. **Confidence: High. Impact:** cleaner가 SessionStart 이전 경쟁 구간에서 사용 중 profile을 삭제할 수 있고 AC-HSR-018을 위반한다. **Required fix:** 최종 profile 확정 직후 모든 Claude start/exec 이전에 하나의 provisional lease를 생성하고 continue-success/fallback이 같은 token을 사용하게 한다. fake-Claude integration test로 child start 시점의 row를 확인한다. **Merge-blocking: yes.**

- **F3 [High] [blocking]** `internal/homestate/profile_lease.go:137` — fresh provisional lease도 parent PID가 dead로 보이면 grace/transfer 상태를 고려하지 않고 `LeaseStale`로 반환한다(`:156-183`). **Confidence: High. Impact:** parent 종료와 child identity CAS 사이의 정상 Windows handoff에서 cleaner 삭제가 가능하며 AC-HSR-018의 provisional grace 계약을 직접 위반한다. **Required fix:** explicit transfer deadline/grace 상태를 schema에 보존하고, transfer 완료·취소 전 또는 fingerprint 불확정 상태는 `LeaseIndeterminate`로 유지한다. parent-death/cleaner/child-CAS barrier test를 실제 동시 실행한다. **Merge-blocking: yes.**

- **F4 [High] [blocking]** `internal/homestate/handoff.go:281` — 기존 public `SetResumeStatus`는 token 없이 `id+status`만으로 claimed row를 consumed/failed로 바꿀 수 있다. `internal/hook/handoff/pending.go:288-310`도 token 생략 호출 시 현재 DB token을 다시 읽어 사용하므로 오래된 consumer가 새 claimant의 token을 대신 읽어 ABA 방어를 우회할 수 있다. **Confidence: High. Impact:** AC-HSR-015의 token CAS 불변식이 API 전체에서 성립하지 않는다. **Required fix:** claimed→terminal 전이의 tokenless API/fallback을 제거하고 모든 production/test caller가 원 claim token을 의무 전달하도록 한다. **Merge-blocking: yes.**

- **F5 [High] [blocking]** `internal/homestate/admission.go:89` — marker release closure가 `os.Remove` 오류를 버린다. apply는 `internal/cli/migrate_home_state.go:448,505`, rollback은 `:614-634`에서 그 void closure를 사용해 marker 제거 실패에도 성공을 반환할 수 있다. **Confidence: High. Impact:** command는 성공했지만 모든 신규 runtime이 영구 차단되는 상태가 되며 AC-HSR-012/023/024와 acceptance edge case를 위반한다. **Required fix:** clear 함수가 오류를 반환하게 하고 marker-last 제거 실패를 command 실패로 전파하며 recovery metadata를 출력한다. **Merge-blocking: yes.**

- **F6 [High] [blocking]** `internal/homestate/runtime_census.go:26` — census가 입력을 primary root로 canonicalize한 뒤 `primary/.moai/state/active-sessions.json` 하나만 읽는다(`:27-30`). linked-worktree 자체 registry의 live session은 누락된다. **Confidence: High. Impact:** false zero-active가 되어 live migration authorization이 발급될 수 있으므로 AC-HSR-003/009/021의 안전 전제를 깨뜨린다. **Required fix:** project에 속하는 primary/caller/worktree registry를 중복 제거하여 fail-closed로 읽고, worktree-local-only live PID fixture를 gate test에 추가한다. **Merge-blocking: yes.**

- **F7 [High] [blocking]** `internal/cli/migrate_home_state.go:380` — dry-run은 target 검사 전 즉시 반환하고 target divergence 로직은 apply-only 구간 `:432-441`에 있다. **Confidence: High. Impact:** operator가 dry-run에서 충돌을 보지 못하며 AC-HSR-005를 위반한다. **Required fix:** dry-run도 target integrity/logical digest를 읽어 equivalent/divergent/unreadable을 보고하되 inventory는 byte-identical하게 유지한다. **Merge-blocking: yes.**

- **F8 [High] [blocking]** `internal/cli/migrate_home_state.go:113` — `runNamedTests`는 `=== RUN` 문자열 존재와 process exit 0만 확인하고 `--- SKIP`을 거부하지 않는다(`:121-125`). **Confidence: High. Impact:** 필수 validator가 모두 skip이어도 verified-live nonce가 발급될 수 있어 AC-HSR-021의 GREEN/new skip 0건 조건을 우회한다. **Required fix:** `go test -json` event를 파싱해 exact test별 pass/fail/skip과 count를 검증하고 skip 하나라도 거부한다. **Merge-blocking: yes.**

- **F9 [High] [blocking]** `internal/cli/home_state_coverage.go:15` — 수동 `changedSurfaceFiles`가 현재 변경 production 파일 13개를 제외한다: `launch_exec_windows.go`, `launcher.go`, `mcp_server.go`, `factory.go`, `handoff.go`, `hook/handoff/pending.go`, `hook/handoff_inject.go`, `session_end.go`, `session_start.go`, `factory_slots.go` 및 Windows identity/lock 3개. **Confidence: High. Impact:** 가장 중요한 admission·lease·handoff 경로를 분모에서 뺀 채 85%를 주장하여 AC-HSR-021/022와 Craft hard threshold가 성립하지 않는다. **Required fix:** current HEAD 대비 production diff에서 coverage 대상 목록을 자동 산출하고, 누락/추가/rename을 fail-closed로 검증한다. 플랫폼별 파일은 명시적 disposition과 대응 test evidence를 요구한다. **Merge-blocking: yes.**

- **F10 [High] [blocking]** `internal/cli/migrate_home_state.go:206` — AC-HSR-022 validator는 임의의 nonempty `Command`/`Output`, caller-supplied `Executed`, `ExitCode`, `Head`만 신뢰한다(`:210-229`). production에서 ledger를 생성·persist·readback하는 caller도 없고 유일한 caller는 synthetic unit test다. **Confidence: High. Impact:** 실행하지 않은 24개 AC와 tooling을 문자열로 위조해 completeness PASS를 만들 수 있다. **Required fix:** immutable/current-HEAD evidence artifact를 실제 command runner가 생성하고, exact selector/JSON test event/output hash/exit/HEAD를 validator가 재검증하도록 연결한다. post-apply readback은 실제 source/target/backup에서 생성해야 한다. **Merge-blocking: yes.**

- **F11 [Medium] [blocking]** `internal/cli/clean_home.go:198,263` — live/indeterminate profile을 scanner에서 조용히 `continue`하고 `runCleanHome`은 `:527-534`에서 generic `nothing to clean`만 출력한다. **Confidence: High. Impact:** AC-HSR-020의 operator-visible skip reason이 없다. **Required fix:** protected profile과 live/indeterminate/error 이유를 scanner result에 포함해 dry-run/force 모두에서 삭제 0건과 이유를 출력하도록 한다. **Merge-blocking: yes.**

- **F12 [Medium] [blocking]** `.moai/reports/t592/verdict.md:1` — verdict 본문에 injection 성공 후 consumed CAS 전 crash가 중복 전달될 수 있다는 구체적인 at-least-once 경계가 없다. 해당 표현은 `progress.md`의 단어 수준 요약과 source comment에만 있다. **Confidence: High. Impact:** AC-HSR-016의 명시적 운영 계약이 충족되지 않는다. **Required fix:** verdict의 Residual-risk에 crash 경계, duplicate 가능성, receiver-side exactly-once 미구현을 명시한다. **Merge-blocking: yes.**

- **F13 [Low] [optional]** `internal/hook/session_end.go:70-77` — 동일 `OpenProfileLeases`/`ReleaseSession`/`Close` 블록이 두 번 연속 실행된다. **Confidence: High. Impact:** 기능 손상은 없지만 불필요한 SQLite open/write와 코드 중복이다. **Recommendation:** 한 블록만 남긴다. **Merge-blocking: no.**

## Recommendations

- F1–F10을 우선 수정한 뒤 동일 overlay mutant를 정식 regression test로 편입한다.
- F11–F12로 operator evidence 계약을 닫고 F13은 같은 파일 수정 시 최소 정리한다.
- live apply는 재감사 PASS 전까지 실행하지 않는다. 현재 결과는 실제 `~/.moai` 이전 승인이 아니다.

## Evidence-bearing record

### Claim

현재 구현은 공식 named selector가 GREEN이어도 10개의 독립 semantic mutant를 방어하지 못한다. Functionality와 Security must-pass가 모두 실패하므로 merge 및 live rollout을 차단한다.

### Evidence

#### 1. 공식 acceptance selector batch

```text
command: unset MOAI_HOME CLAUDE_PROJECT_DIR CODEX_HOME MOAI_PROFILE_LEASE_TOKEN CLAUDE_CONFIG_DIR && MOAI_HOME=$(mktemp -d /private/tmp/t592-func.XXXXXX) go test ./internal/cli ./internal/homestate ./internal/kanban -run '<25 exact selectors>' -count=1 -v
exit: 0
output:
--- PASS: TestHomeStateDryRunNoMutation
--- PASS: TestHomeStateDryRunReport
--- PASS: TestHomeStateApplyCensusFailClosed
--- PASS: TestHomeStateBackupBeforeWrite
--- PASS: TestHomeStateRefusesDivergentTarget
--- PASS: TestHomeStateApplyFaultPreservesSource
--- PASS: TestHomeStateApplyPreservesSourceAndBackup
--- PASS: TestHomeStateApplyIdempotentNoOp
--- PASS: TestHomeStateBarrierAdmissionHaltsAllHosts
--- PASS: TestHomeStateStartVsMigrateSerialized
--- PASS: TestHomeStateCrashMarkerFailsClosed
--- PASS: TestHomeStateVerifiedLiveGateCannotBypassOrReplay
--- PASS: TestHomeStateRecoverCrashMarkerSafely
--- PASS: TestHomeStateRollbackVerifiedBackup
--- PASS: TestHomeStateVerdictEvidenceValidator
--- PASS: TestProfileLeaseLifecycleAndNonExecCleanerRace
--- PASS: TestCleanHomeSkipsLiveAndIndeterminateProfiles
--- PASS: TestFactoryV1ClaimedRowsUpgradeToV2
--- PASS: TestResumeLatestPendingThenExpiredReclaim
--- PASS: TestResumeFinishRejectsABAToken
--- PASS: TestResumeInjectionCrashIsAtLeastOnce
--- PASS: TestProfileLeasesAreGlobalAndPrivate
--- PASS: TestProfileLeaseReconcilePIDFingerprint
--- PASS: TestResolveTodoQueueRoot_WorktreeConvergesOnPrimary
PASS
ok github.com/modu-ai/moai-adk/internal/cli 46.268s
ok github.com/modu-ai/moai-adk/internal/homestate 0.461s
ok github.com/modu-ai/moai-adk/internal/kanban 0.967s
```

#### 2. Auditor semantic overlay

Overlay 파일은 `/private/tmp`에만 만들었고 worktree source에는 추가하지 않았다.

```text
command: go test -overlay=/private/tmp/t592_overlay_cli.json ./internal/cli -run '^(TestAuditDryRunReportsDivergentTarget|TestAuditEvidenceValidatorRejectsFabricatedStrings|TestAuditRunNamedTestsRejectsSkip|TestAuditFactoryLauncherCannotSwallowBarrier|TestAuditContinueLaunchCreatesProvisionalLeaseBeforeStart|TestAuditCoverageListIncludesEveryChangedProductionFile)$' -count=1 -v
exit: 1
output:
--- FAIL: TestAuditDryRunReportsDivergentTarget
    AC-HSR-005: dry-run did not report divergent target
--- FAIL: TestAuditEvidenceValidatorRejectsFabricatedStrings
    AC-HSR-022: fabricated command/output strings were accepted
--- FAIL: TestAuditRunNamedTestsRejectsSkip
    AC-HSR-021: skipped validator was accepted as GREEN
--- FAIL: TestAuditFactoryLauncherCannotSwallowBarrier
    AC-HSR-010: production Factory launcher swallowed admission denial and returned a launchable lane label
--- FAIL: TestAuditContinueLaunchCreatesProvisionalLeaseBeforeStart
    AC-HSR-018: successful --continue started Claude before creating a provisional profile lease
--- FAIL: TestAuditCoverageListIncludesEveryChangedProductionFile
    AC-HSR-021 changed-surface coverage omitted production files: [internal/cli/launch_exec_windows.go internal/cli/launcher.go internal/cli/mcp_server.go internal/homestate/factory.go internal/homestate/handoff.go internal/hook/handoff/pending.go internal/hook/handoff_inject.go internal/hook/session_end.go internal/hook/session_start.go internal/kanban/factory_slots.go internal/homestate/admission_lock_windows.go internal/homestate/pid_state_windows.go internal/homestate/process_fingerprint_windows.go]
FAIL
```

```text
command: go test -overlay=/private/tmp/t592_overlay_homestate.json ./internal/homestate -run '^(TestAuditDeadProvisionalLeaseRemainsIndeterminateDuringGrace|TestAuditLegacyFinishWithoutTokenCannotBypassABA|TestAuditMarkerClearFailureIsObservable|TestAuditCensusIncludesLinkedWorktreeRegistry)$' -count=1 -v
exit: 1
output:
--- FAIL: TestAuditDeadProvisionalLeaseRemainsIndeterminateDuringGrace
    AC-HSR-018: fresh dead-parent provisional lease=stale, want indeterminate during transfer grace
--- FAIL: TestAuditLegacyFinishWithoutTokenCannotBypassABA
    AC-HSR-015: tokenless SetResumeStatus consumed a row reclaimed under token B
--- FAIL: TestAuditMarkerClearFailureIsObservable
    AC-HSR-012/023/024: marker clear failed silently and runtime stayed blocked
--- FAIL: TestAuditCensusIncludesLinkedWorktreeRegistry
    AC-HSR-003/009: worktree-local live registry was omitted: {ActiveSessions:0 ActiveFactoryWorkers:0 ActiveMCPServers:0 Fingerprint:0:0:0}
FAIL
```

#### 3. Security dimension

```text
command: go mod verify
exit: 0
output:
all modules verified

secret-probe:
no credential-shaped additions found

dynamic-sql-probe:
internal/cli/migrate_home_state.go:268: info, err := db.Query(`PRAGMA table_info(` + quoteID(name) + `)`)
internal/cli/migrate_home_state.go:293: rows, err := db.Query(`SELECT * FROM ` + quoteID(name) + ` ORDER BY ` + strings.Join(order, ","))
internal/cli/migrate_home_state.go:333: if _, err := db.Exec(`VACUUM INTO '` + escaped + `'`); err != nil {
```

동적 identifier는 double-quote escaping, `VACUUM INTO` path는 single-quote escaping을 적용하므로 이 세 행 자체를 injection finding으로 판정하지 않았다. Security FAIL은 F1/F2/F3의 실제 admission·deletion-safety 위반 때문이다.

#### 4. Craft와 Consistency dimension

```text
command: auditor changed-production-file coverage membership test
exit: 1
output:
AC-HSR-021 changed-surface coverage omitted production files: [13 files]
```

```text
command: gofmt -d <all changed Go files>; git diff --check
exit: 0
output:
gofmt-diff: empty
git-diff-check: empty
duplicate-session-release-probe:
71: _ = store.ReleaseSession(ctx, input.SessionID)
75: _ = store.ReleaseSession(ctx, input.SessionID)
```

### Baseline-attribution

모든 command와 source line은 2026-09-10 현재 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t592`, branch `WT-home-state-rollout`, HEAD `6ea69661c405c1b6a3ef38e598f5653a63cd7af0`와 그 위의 미커밋 구현 diff를 대상으로 측정했다. 테스트는 격리된 `MOAI_HOME`, `t.TempDir()` 및 `/private/tmp` overlay만 사용했다. 실제 `~/.moai`에는 apply/clean/write를 실행하지 않았다.

### Gaps

- 실제 live apply와 post-apply source/target readback은 실행하지 않았다. blocker가 존재하므로 실행하면 안 된다.
- 장시간 changed-surface coverage 재측정은 parent 요청에 따라 중단했다. 더구나 현재 allowlist가 13개 production 파일을 누락하므로 기존 85.047%도 전체 changed surface의 증거가 아니다.
- Windows 바이너리는 이 audit에서 실행하지 않았다. source의 Windows 경로는 coverage 누락 finding에 포함했다.
- 전체 repository suite와 integration-branch CI는 실행하지 않았다. 변경 영향 selector와 semantic mutants로 판정했다.

### Residual-risk

현재 발견 외에도 SQLite target의 `-wal`/`-shm` crash artifact, backup ID path normalization, marker 디렉터리 fsync/atomic durability는 추가 adversarial fixture가 필요하다. 다만 이미 12개 blocking finding이 있으므로 이 미검증 영역은 PASS로 간주하지 않는다.

## Iteration history

| Iteration | Baseline | Verdict | Blocking findings |
|-----------|----------|---------|------------------:|
| sync-audit 1 | `6ea69661c` + current diff | FAIL 35/100 | 12 |
| sync-audit 2 | `6ea69661c` + remediation diff | **FAIL 68/100** | **1** |
| sync-audit 3 | `6ea69661c` + F6-R2 remediation diff | **FAIL 95/100** | **1** |

---

## Iteration 2 — remediation delta re-audit

### Evaluation Report

SPEC: `SPEC-HOME-STATE-ROLLOUT-001` v0.5.0  
Baseline: branch `WT-home-state-rollout`, HEAD `6ea69661c405c1b6a3ef38e598f5653a63cd7af0`, current uncommitted remediation diff  
Overall Verdict: **FAIL — 68/100**  
Must-pass firewall: **Functionality 50/100 FAIL; Security 50/100 FAIL**

| Dimension | Score | Verdict | Evidence |
|-----------|------:|---------|----------|
| Functionality (40%) | 50/100 | FAIL | 보강 focused batch와 실제 Factory command overlay는 GREEN이지만, linked-worktree discovery 실패 반증 test가 live registry를 누락한 `0:0:0` census를 재현했다. |
| Security (25%) | 50/100 | FAIL | `go mod verify`는 `all modules verified`; credential/shell probe는 finding이 없었으나 false-zero census가 active runtime 중 migration authorization을 허용할 수 있는 High safety finding이다. |
| Craft (20%) | 100/100 | PASS | 자동 HEAD diff 기반 영구 validator가 `1198/1408 = 85.085%`를 직접 재현했다. |
| Consistency (15%) | 100/100 | PASS | `gofmt` clean, tokenless resume API 0건, SessionEnd duplicate release 제거를 확인했다. |

가중 점수: `50×0.40 + 50×0.25 + 100×0.20 + 100×0.15 = 67.5`, 반올림 `68/100`. Functionality/Security must-pass 실패로 overall FAIL이다.

### Findings

- **F6-R2 [High] [blocking]** `internal/homestate/runtime_census.go:36-45` — `git -C <canonical-root> worktree list --porcelain` 실패를 오류 없이 무시하고 primary/caller root만으로 determinate census를 만든다. 실제 git primary+linked worktree fixture에서 linked registry에 현재 live PID를 기록한 뒤 worktree inventory만 unavailable하게 만들자 `ReadRuntimeCensus(primary)`가 오류 대신 `{ActiveSessions:0 ActiveFactoryWorkers:0 ActiveMCPServers:0 Fingerprint:0:0:0}`을 반환했다. **Confidence: High. Impact:** AC-HSR-003/009/021의 “읽을 수 없으면 fail-closed” 조건을 깨고, linked lane이 실행 중인데도 verified-live apply nonce가 발급될 수 있다. **Required fix:** worktree inventory command 실패를 census error로 전파하고, malformed/empty porcelain 및 discovered root canonicalization 오류도 fail-closed 처리한다. 현재 auditor fixture를 영구 regression test로 편입한다. **Merge-blocking: yes.**

### Verified remediation non-findings

- F1: actual `runCC -f lane-7` command path가 claim error를 반환하고 backend launch capture는 nil이었다.
- F2: `--continue` fake child가 시작되는 시점에 lease DB와 token이 이미 존재했고 정확히 한 provisional row가 확인됐다.
- F3: dead provisional lease가 transfer deadline 동안 indeterminate로 유지됐다.
- F4: `SetResumeStatus`/tokenless `FinishClaim` source probe는 0건이며 ABA test가 통과했다.
- F5: marker clear failure가 error로 관찰되는 영구 test가 통과했다.
- F6: 정상 git inventory에서는 linked-worktree registry를 읽는다. 단, inventory 실패 경로는 F6-R2로 미완료다.
- F7: dry-run이 absent/equivalent/divergent/unreadable target 상태를 보고하면서 mutation하지 않는 test가 통과했다.
- F8: go-test JSON skip/fail/zero-event/duplicate-pass를 모두 거부했다.
- F9: production diff/platform disposition 자동 산출과 exact changed-line result `1198/1408 = 85.085%`를 재현했다.
- F10: evidence file O_EXCL/0400, output SHA-256, current HEAD, 실제 source/target/backup SQLite readback 검증 test가 통과했다.
- F11: dry-run/force 모두 protected profile 이유를 출력하고 삭제 0건을 보장하는 test가 통과했다.
- F12: `verdict.md:184`가 injection-success/consumed-CAS crash의 at-least-once duplicate와 receiver-side exactly-once 미구현을 명시한다.
- F13: SessionEnd의 중복 release 블록이 제거됐다.

### Evidence-bearing record

#### Claim

13개 기존 finding 중 12개는 기계적으로 닫혔고 F6 정상 경로도 보강됐지만, worktree inventory 자체가 실패할 때 census가 fail-open한다. 따라서 merge 및 live rollout은 계속 차단한다.

#### Evidence

```text
command: audit_home=$(mktemp -d) && unset MOAI_PROFILE_LEASE_TOKEN CLAUDE_PROJECT_DIR MOAI_KANBAN_ID MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && MOAI_HOME="$audit_home" go test ./internal/cli ./internal/homestate ./internal/hook -run '^(TestResolveFactoryWorkerName|TestCC_FactoryEntryThroughRunCC|TestContinueLaunchCreatesSingleProvisionalLeaseBeforeChildStart|TestProfileLease.*|TestDeadProvisionalLease.*|TestResume.*|TestAdmissionMarkerClearFailureIsObservable|TestRuntimeCensus.*|TestHomeStateDryRun.*|TestHomeStateRunNamedTestsRejectsSkipFailAndZeroJSONEvents|TestPersistedHomeStateEvidenceIsImmutableAndReadsRealDatabases|TestCleanHomeReportsProtectedProfileReasonInDryRunAndForce|TestSessionStartProfileLeaseDirectAndTokenFallback)$' -count=1 -v
exit: 0
output:
--- PASS: TestResolveFactoryWorkerName
--- PASS: TestCC_FactoryEntryThroughRunCC
--- PASS: TestHomeStateDryRunClassifiesEquivalentAndDivergentTargets
--- PASS: TestHomeStateRunNamedTestsRejectsSkipFailAndZeroJSONEvents
--- PASS: TestPersistedHomeStateEvidenceIsImmutableAndReadsRealDatabases
--- PASS: TestContinueLaunchCreatesSingleProvisionalLeaseBeforeChildStart
--- PASS: TestCleanHomeReportsProtectedProfileReasonInDryRunAndForce
--- PASS: TestAdmissionMarkerClearFailureIsObservable
--- PASS: TestResumeFinishRejectsABAToken
--- PASS: TestResumeInjectionCrashIsAtLeastOnce
--- PASS: TestDeadProvisionalLeaseRemainsIndeterminateUntilTransferDeadline
--- PASS: TestRuntimeCensusIncludesLinkedWorktreeLocalRegistry
--- PASS: TestRuntimeCensusRejectsCorruptRegistries
--- PASS: TestSessionStartProfileLeaseDirectAndTokenFallback
PASS
ok github.com/modu-ai/moai-adk/internal/cli 12.479s
ok github.com/modu-ai/moai-adk/internal/homestate 3.500s
ok github.com/modu-ai/moai-adk/internal/hook 0.837s
```

```text
command: MOAI_HOME=<temp> go test -overlay=/private/tmp/t592_iter2_cli_overlay.json ./internal/cli -run '^TestAuditActualCCFactoryLaunchPropagatesClaimError$' -count=1 -v
exit: 0
output:
=== RUN   TestAuditActualCCFactoryLaunchPropagatesClaimError
--- PASS: TestAuditActualCCFactoryLaunchPropagatesClaimError (0.28s)
PASS
ok github.com/modu-ai/moai-adk/internal/cli 1.229s
```

```text
command: MOAI_HOME=<temp> go test -overlay=/private/tmp/t592_iter2_homestate_overlay.json ./internal/homestate -run '^TestAuditRuntimeCensusFailsClosedWhenWorktreeDiscoveryUnavailable$' -count=1 -v
exit: 1
output:
=== RUN   TestAuditRuntimeCensusFailsClosedWhenWorktreeDiscoveryUnavailable
    t592_iter2_homestate_probe_test.go:23: worktree discovery failure accepted as determinate census: {ActiveSessions:0 ActiveFactoryWorkers:0 ActiveMCPServers:0 Fingerprint:0:0:0}
--- FAIL: TestAuditRuntimeCensusFailsClosedWhenWorktreeDiscoveryUnavailable (0.44s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/homestate 0.955s
FAIL
```

```text
command: audit_userhome=$(mktemp -d) && unset MOAI_HOME MOAI_PROFILE_LEASE_TOKEN CLAUDE_PROJECT_DIR MOAI_KANBAN_ID MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && HOME="$audit_userhome" go test ./internal/cli -run '^TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite$' -count=1 -v
exit: 0
output:
=== RUN   TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite
    home_state_coverage_test.go:127: auto-diff changed production coverage: 1198/1408 = 85.085%
--- PASS: TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite (87.32s)
PASS
ok github.com/modu-ai/moai-adk/internal/cli 88.294s
```

```text
command: go mod verify; OWASP shell/secret/dynamic-SQL rg probe
exit: 0
output:
all modules verified
internal/cli/migrate_home_state.go:466: rows, err := db.Query(`SELECT * FROM ` + quoteID(name) + ` ORDER BY ` + strings.Join(order, ","))
```

동적 SQL 행은 table/column identifier를 `quoteID`로 double-quote escaping한 내부 SQLite schema census이며 외부 raw identifier를 직접 실행하지 않아 finding으로 분류하지 않았다.

```text
command: gofmt -l <changed production files>; rg tokenless APIs; rg -n 'ReleaseSession' internal/hook/session_end.go
exit: 0
output:
GOFMT_OK
tokenless=0
session_end_ReleaseSession_calls=1
71:        _ = store.ReleaseSession(ctx, input.SessionID)
```

#### Baseline-attribution

모든 결과는 2026-09-10 현재 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t592`, branch `WT-home-state-rollout`, HEAD `6ea69661c405c1b6a3ef38e598f5653a63cd7af0`와 current remediation diff에서 이번 iteration에 직접 측정했다. HOME/MOAI_HOME은 임시 디렉터리로 격리했고 live `~/.moai`에는 apply/clean/write하지 않았다.

#### Gaps

- 실제 사용자 홈 `--apply --verified-live`와 post-apply AC-HSR-022 external append/독립 completeness는 parent 지시와 안전상 실행하지 않았다.
- Windows 바이너리 runtime은 실행하지 않았다. 기존 Windows compile/test evidence와 source disposition을 이번 focused audit에서 재실행하지 않았다.
- 전체 repository suite/CI는 실행하지 않았다. 변경 영향 focused tests와 adversarial overlays만 실행했다.

#### Residual-risk

F6-R2가 수정되기 전에는 second census도 같은 false-zero 결과를 반복할 수 있어 “두 번 읽음”이 안전성을 회복하지 못한다. 이 상태에서 live rollout은 실행하면 안 된다. Coverage는 85%를 1 statement만 넘으므로 remediation diff 뒤 exact validator를 다시 실행해야 한다.

### Recommendations

- `git worktree list --porcelain`의 non-zero/parse failure를 즉시 census error로 전파하고 linked-live fixture를 영구 test로 편입한다.
- 수정 후 F6-R2 delta test와 exact changed-line coverage를 재실행한 뒤 iteration 3 독립 판정을 받는다.
- PASS 전에는 live `~/.moai` apply를 실행하지 않는다.

---

## Iteration 4 — F14 closure

### Evaluation Report

SPEC: `SPEC-HOME-STATE-ROLLOUT-001` v0.5.0  
Baseline: branch `WT-home-state-rollout`, HEAD `6ea69661c405c1b6a3ef38e598f5653a63cd7af0`, current uncommitted test-only F14 closure diff  
Overall Verdict: **PASS — 100/100**  
Findings: **0**

| Dimension | Score | Verdict | Evidence |
|-----------|------:|---------|----------|
| Functionality (40%) | 100/100 | PASS | permanent primary+linked same-PID fixture가 양쪽 caller에서 count 1과 stable fingerprint를 검증한다. |
| Security (25%) | 100/100 | PASS | F6-R2 fail-closed와 verified-live error propagation은 iteration 3에서 GREEN이며 production 변경이 없다. |
| Craft (20%) | 100/100 | PASS | dedup guard branch가 focused coverprofile에서 실행됐고 exact changed-production coverage는 `1214/1425 = 85.193%`다. |
| Consistency (15%) | 100/100 | PASS | 새 영구 test는 gofmt clean이다. |

### Findings

없음. F14는 닫혔다.

### Evidence-bearing record

#### Claim

영구 regression test가 primary와 linked registry에 동일 live PID를 기록하고 두 진입점 모두 정확히 1건 및 동일 `1:0:0` fingerprint를 요구한다. dedup guard가 실제 실행되므로 `seenSessionPID` 제거 mutant는 이 assertion을 통과할 수 없다.

#### Evidence

```text
command: MOAI_HOME=<temp> go test ./internal/homestate -run '^TestRuntimeCensusDeduplicatesSameLivePIDAcrossPrimaryAndLinkedRegistries$' -count=1 -v
exit: 0
output:
=== RUN   TestRuntimeCensusDeduplicatesSameLivePIDAcrossPrimaryAndLinkedRegistries
--- PASS: TestRuntimeCensusDeduplicatesSameLivePIDAcrossPrimaryAndLinkedRegistries (0.86s)
PASS
ok github.com/modu-ai/moai-adk/internal/homestate 1.081s
```

```text
command: MOAI_HOME=<temp> go test ./internal/homestate -run '^TestRuntimeCensusDeduplicatesSameLivePIDAcrossPrimaryAndLinkedRegistries$' -count=1 -coverprofile=<temp>; rg 'runtime_census.go:(83|84|85|86)' <temp>
exit: 0
output:
github.com/modu-ai/moai-adk/internal/homestate/runtime_census.go:83.32,84.14 1 1
github.com/modu-ai/moai-adk/internal/homestate/runtime_census.go:86.5,88.46 3 1
MUTANT_GUARD_EXECUTED
```

```text
permanent assertions: internal/homestate/runtime_census_test.go:146-150
first.ActiveSessions == 1
second.ActiveSessions == 1
first.Fingerprint == "1:0:0"
second.Fingerprint == first.Fingerprint
```

```text
changed-production exact coverage, unchanged production diff baseline:
auto-diff changed production coverage: 1214/1425 = 85.193%
PASS
```

F14는 `_test.go`만 추가했다. `home_state_coverage.go`는 `_test.go`를 production denominator에서 제외하며 production source와 coverage validator는 iteration 3 exact 측정 이후 변경되지 않았다.

#### Baseline-attribution

focused test와 branch coverage는 현재 t592 worktree에서 직접 측정했다. Exact coverage는 동일 production diff에 대한 iteration 3의 직접 측정 `1214/1425 = 85.193%`를 연속 baseline으로 사용했다. F14 closure는 `runtime_census_test.go`만 추가했고 production denominator/validator는 동일함을 현재 diff에서 확인했다.

#### Gaps

- 실제 사용자 홈 apply와 post-apply AC-HSR-022는 실행하지 않았다.
- parent 지시에 따라 장시간 exact validator와 전체 suite는 다시 실행하지 않았다. production diff가 동일하므로 직전 direct exact measurement를 사용했다.

#### Residual-risk

실제 rollout 성공은 이 PASS에 포함되지 않는다. 적용 시 fresh zero-active census, current-HEAD one-shot authorization, backup/restore probe와 post-apply AC-HSR-022 독립 검증이 별도로 필요하다.

## Final iteration history

| Iteration | Verdict | Blocking findings |
|-----------|---------|------------------:|
| 1 | FAIL 35/100 | 12 |
| 2 | FAIL 68/100 | 1 |
| 3 | FAIL 95/100 | 1 |
| 4 | **PASS 100/100** | **0** |

---

## Iteration 3 — F6-R2 final focused re-audit

### Evaluation Report

SPEC: `SPEC-HOME-STATE-ROLLOUT-001` v0.5.0  
Baseline: branch `WT-home-state-rollout`, HEAD `6ea69661c405c1b6a3ef38e598f5653a63cd7af0`, current uncommitted remediation diff  
Overall Verdict: **FAIL — 95/100**  
Must-pass firewall: Functionality **PASS**, Security **PASS**. 명시된 permanent-regression coverage finding 1건이 blocking이므로 overall FAIL이다.

| Dimension | Score | Verdict | Evidence |
|-----------|------:|---------|----------|
| Functionality (40%) | 100/100 | PASS | F6-R2 영구 edge fixtures와 auditor dedup/verified-live overlays, F1–F13 smoke가 모두 GREEN이다. |
| Security (25%) | 100/100 | PASS | inventory blindness는 fail-closed하며 `go mod verify`는 `all modules verified`였다. |
| Craft (20%) | 75/100 | FAIL | exact diff coverage는 85.193%지만 cross-worktree duplicate PID 제거 mutant를 죽이는 영구 test가 없다. |
| Consistency (15%) | 100/100 | PASS | F6-R2 files는 gofmt clean이고 `git diff --check`도 clean이다. |

가중 점수: `100×0.40 + 100×0.25 + 75×0.20 + 100×0.15 = 95/100`.

### Findings

- **F14 [Medium] [blocking]** `internal/homestate/runtime_census_test.go:95-116` — linked-worktree 영구 test는 live PID를 linked registry 한 곳에만 기록한다. primary와 linked registry에 같은 PID가 동시에 존재하는 fixture가 없어 `runtime_census.go:71-86`의 `seenSessionPID` dedup을 제거해도 현재 영구 suite가 통과한다. Auditor overlay에서는 동일 PID를 양쪽 registry에 기록했을 때 `ActiveSessions == 1`로 올바르게 동작함을 확인했지만, parent가 요구한 “permanent tests cover ... dedup”은 충족되지 않았다. **Confidence: High. Impact:** 향후 regression 시 runtime count/fingerprint가 호출 위치·중복 registry에 따라 달라지고 census 보고 및 HEAD+census authorization binding이 불안정해진다. **Required fix:** `TestRuntimeCensusIncludesLinkedWorktreeLocalRegistry` 또는 별도 영구 test에서 primary+linked에 동일 live PID를 기록하고 정확히 1건으로 집계됨을 assertion한다. **Merge-blocking: yes.**

### Verified non-findings

- `git worktree list` command error는 `worktree census inventory` error로 전파된다.
- empty/malformed output, nonexistent path, canonical-root mismatch는 모두 fail-closed한다.
- linked worktree의 live registry는 production `ReadRuntimeCensus`에서 집계된다.
- cross-worktree duplicate PID는 production에서 1건으로 dedup된다.
- worktree inventory가 unavailable하면 production `prepareLiveAuthorization`이 authorization을 발급하지 않는다.
- F1–F13 focused smoke는 모두 GREEN이다.
- exact changed-production diff coverage는 `1214/1425 = 85.193%`로 재현됐다.

### Evidence-bearing record

#### Claim

F6-R2 runtime defect 자체는 수정됐고 verified-live까지 fail-closed한다. 다만 동일 PID cross-worktree dedup을 고정하는 영구 regression test가 없어 최종 merge gate는 아직 닫히지 않았다.

#### Evidence

```text
command: MOAI_HOME=<temp> go test ./internal/homestate ./internal/cli -run '^(TestRuntimeCensusFailsClosedWhenWorktreeInventoryUnavailable|TestRuntimeCensusFailsClosedOnMalformedOrUnavailableWorktreeInventory|TestRuntimeCensusIncludesLinkedWorktreeLocalRegistry|TestRuntimeCensusCountsLiveAndIgnoresProvablyDead|TestHomeStateAuthorizationPreparationFailsClosed)$' -count=1 -v
exit: 0
output:
--- PASS: TestRuntimeCensusFailsClosedWhenWorktreeInventoryUnavailable (0.58s)
--- PASS: TestRuntimeCensusFailsClosedOnMalformedOrUnavailableWorktreeInventory (0.23s)
    --- PASS: .../empty
    --- PASS: .../malformed
    --- PASS: .../missing-root
    --- PASS: .../canonical-mismatch
--- PASS: TestRuntimeCensusCountsLiveAndIgnoresProvablyDead (0.56s)
--- PASS: TestRuntimeCensusIncludesLinkedWorktreeLocalRegistry (0.61s)
ok github.com/modu-ai/moai-adk/internal/homestate 2.419s
--- PASS: TestHomeStateAuthorizationPreparationFailsClosed (0.40s)
ok github.com/modu-ai/moai-adk/internal/cli 2.381s
```

```text
command: MOAI_HOME=<temp> go test -overlay=/private/tmp/t592_iter2_homestate_overlay.json ./internal/homestate -run '^TestAuditRuntimeCensusDeduplicatesSamePIDAcrossWorktrees$' -count=1 -v
exit: 0
output:
=== RUN   TestAuditRuntimeCensusDeduplicatesSamePIDAcrossWorktrees
--- PASS: TestAuditRuntimeCensusDeduplicatesSamePIDAcrossWorktrees (0.62s)
PASS
ok github.com/modu-ai/moai-adk/internal/homestate 1.065s
```

```text
command: MOAI_HOME=<temp> go test -overlay=/private/tmp/t592_iter2_cli_overlay.json ./internal/cli -run '^TestAuditVerifiedLiveFailsClosedWhenWorktreeInventoryUnavailable$' -count=1 -v
exit: 0
output:
=== RUN   TestAuditVerifiedLiveFailsClosedWhenWorktreeInventoryUnavailable
--- PASS: TestAuditVerifiedLiveFailsClosedWhenWorktreeInventoryUnavailable (0.00s)
PASS
ok github.com/modu-ai/moai-adk/internal/cli 1.059s
```

```text
command: MOAI_HOME=<temp> go test ./internal/cli ./internal/homestate ./internal/hook -run '<concise F1-F13 smoke selector>' -count=1 -v
exit: 0
output:
--- PASS: TestResolveFactoryWorkerName
--- PASS: TestCC_FactoryEntryThroughRunCC
--- PASS: TestHomeStateDryRunClassifiesEquivalentAndDivergentTargets
--- PASS: TestHomeStateRunNamedTestsRejectsSkipFailAndZeroJSONEvents
--- PASS: TestPersistedHomeStateEvidenceIsImmutableAndReadsRealDatabases
--- PASS: TestContinueLaunchCreatesSingleProvisionalLeaseBeforeChildStart
--- PASS: TestCleanHomeReportsProtectedProfileReasonInDryRunAndForce
--- PASS: TestAdmissionMarkerClearFailureIsObservable
--- PASS: TestResumeFinishRejectsABAToken
--- PASS: TestResumeInjectionCrashIsAtLeastOnce
--- PASS: TestDeadProvisionalLeaseRemainsIndeterminateUntilTransferDeadline
--- PASS: TestSessionStartProfileLeaseDirectAndTokenFallback
PASS
ok github.com/modu-ai/moai-adk/internal/cli 9.964s
ok github.com/modu-ai/moai-adk/internal/homestate 0.510s
ok github.com/modu-ai/moai-adk/internal/hook 2.409s
```

```text
command: HOME=<temp> env -u MOAI_HOME go test ./internal/cli -run '^TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite$' -count=1 -v
exit: 0
output:
=== RUN   TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite
    home_state_coverage_test.go:127: auto-diff changed production coverage: 1214/1425 = 85.193%
--- PASS: TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite (87.08s)
PASS
ok github.com/modu-ai/moai-adk/internal/cli 88.026s
```

```text
command: go mod verify; gofmt -l internal/homestate/runtime_census.go internal/homestate/runtime_census_test.go; git diff --check
exit: 0
output:
all modules verified
GOFMT_OK
```

#### Baseline-attribution

모든 결과는 2026-09-10 현재 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t592`, branch `WT-home-state-rollout`, HEAD `6ea69661c405c1b6a3ef38e598f5653a63cd7af0`와 current remediation diff에서 직접 측정했다. 테스트는 임시 HOME/MOAI_HOME과 `/private/tmp` overlay만 사용했으며 live `~/.moai`에는 apply/clean/write하지 않았다.

#### Gaps

- 실제 사용자 홈 apply와 post-apply AC-HSR-022 external closure는 실행하지 않았다.
- Windows runtime 및 full repository CI는 이번 focused iteration에서 재실행하지 않았다.
- cross-worktree dedup은 auditor overlay로 실행됐지만 영구 repository test가 아니다.

#### Residual-risk

현재 production dedup은 정상이나 영구 test 부재로 회귀 방어가 없다. 이 test를 추가하면 남은 F14는 좁은 delta 재감사로 닫을 수 있다. Coverage 여유는 3 statement 수준이므로 test 추가 후 exact validator를 다시 실행해야 한다.

### Recommendations

- primary와 linked registry에 동일 live PID를 기록하고 `ActiveSessions == 1`을 검증하는 영구 test 하나를 추가한다.
- 해당 test 및 exact coverage를 재실행한 뒤 최종 재감사를 받는다.

---

## Authoritative final verdict

**PASS — 100/100, findings 0.** Iteration 4가 최종 판정이며 이전 FAIL 섹션은 감사 이력이다. F14 영구 same-PID primary+linked test는 GREEN이고 dedup guard branch가 coverprofile에서 실행됐다. 동일 production diff의 exact coverage는 `1214/1425 = 85.193%`다. 실제 live apply와 post-apply AC-HSR-022는 이 판정 범위 밖이며 실행하지 않았다.

---

## Iteration 5 — post-merge clean-git coverage regression

### Evaluation Report

SPEC: `SPEC-HOME-STATE-ROLLOUT-001` v0.5.0
Baseline: branch `WT-home-state-rollout`, HEAD `18f446d579dcd5f254ff9ef7c2be30366b72505c`, five-file uncommitted remediation set
Overall Verdict: **PASS — 100/100**
Findings: **0**

| Dimension | Score | Verdict | Evidence |
|-----------|------:|---------|----------|
| Functionality (40%) | 100/100 | PASS | committed change-set resolver matrix와 F1–F14 smoke가 모두 GREEN이다. |
| Security (25%) | 100/100 | PASS | stale/ambiguous/non-descendant/unreachable/invalid Git evidence가 fail-closed하며 외부 base 입력은 없다. |
| Craft (20%) | 100/100 | PASS | exact current union coverage `1196/1407 = 85.004%`, malformed/tampered profile 거부, self-recursion guard 확인. |
| Consistency (15%) | 100/100 | PASS | `go mod verify`, gofmt, `git diff --check`가 모두 통과했다. |

### Findings

없음.

### Claim

Resolver는 clean original tip, descendant remediation tip, ancestry-preserving merge HEAD, later unrelated commit 및 dirty production diff union을 처리한다. Audited path stale blob, duplicate marker, non-descendant remediation, unreachable marker, invalid Git/parentless marker/malformed evidence는 모두 거부한다.

### Evidence

```text
command: MOAI_HOME=<temp> go test ./internal/cli -run '^(TestCommittedCoverageChangeSet.*|TestParseChangedSurfaceCoverageRejectsMissingZeroAndTamperedProfiles|TestHomeStateChangedSurfaceCoverageConsumesFreshProfile)$' -count=1 -v
exit: 0
output:
--- PASS: TestCommittedCoverageChangeSetWorksInCleanRepositoryAndAfterUnrelatedCommit (1.33s)
--- PASS: TestCommittedCoverageChangeSetWorksAfterMergeCommit (1.28s)
--- PASS: TestCommittedCoverageChangeSetWorksCleanAfterRemediationCommit (1.64s)
--- PASS: TestCommittedCoverageChangeSetMergesDirtyProductionDiff (1.26s)
--- PASS: TestCommittedCoverageChangeSetRejectsStaleOrAmbiguousEvidence (3.82s)
    --- PASS: .../covered_path_changed_after_audited_tip
    --- PASS: .../duplicate_audit_marker
    --- PASS: .../audit_marker_not_reachable_from_head
    --- PASS: .../remediation_marker_is_not_a_descendant
    --- PASS: .../duplicate_remediation_marker
--- PASS: TestCommittedCoverageChangeSetRejectsInvalidGitAndMalformedEvidence (3.20s)
--- PASS: TestParseChangedSurfaceCoverageRejectsMissingZeroAndTamperedProfiles (0.00s)
--- PASS: TestHomeStateChangedSurfaceCoverageConsumesFreshProfile (8.39s)
PASS
ok github.com/modu-ai/moai-adk/internal/cli 23.468s
```

```text
command: HOME=<temp> env -u MOAI_HOME go test ./internal/cli -run '^TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite$' -count=1 -v
exit: 0
output:
=== RUN   TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite
    home_state_coverage_test.go:341: auto-diff changed production coverage: 1196/1407 = 85.004%
--- PASS: TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite (80.22s)
PASS
ok github.com/modu-ai/moai-adk/internal/cli 81.090s
```

```text
command: MOAI_HOME=<temp> go test ./internal/cli ./internal/homestate ./internal/hook -run '<F1-F14 concise smoke>' -count=1 -v
exit: 0
output:
PASS
ok github.com/modu-ai/moai-adk/internal/cli 9.902s
ok github.com/modu-ai/moai-adk/internal/homestate 1.821s
ok github.com/modu-ai/moai-adk/internal/hook 1.152s
```

```text
command: go mod verify; gofmt -l internal/cli/home_state_coverage.go internal/cli/home_state_coverage_test.go; git diff --check
exit: 0
output:
all modules verified
GOFMT_OK
external_base_inputs=0
self_recursion_guard:
329: if os.Getenv("MOAI_HOME_STATE_COVERAGE_CHILD") == "1" {
332: t.Setenv("MOAI_HOME_STATE_COVERAGE_CHILD", "1")
```

### Baseline-attribution

모든 command는 현재 t592 worktree의 HEAD `18f446d579dcd5f254ff9ef7c2be30366b72505c`와 다섯 uncommitted remediation files를 대상으로 실행했다. 테스트는 임시 HOME/MOAI_HOME만 사용했으며 실제 사용자 홈에는 apply하지 않았다.

### Integration contract

- 미래 remediation commit subject는 정확히 `fix(state): stabilize committed coverage evidence (t592)`여야 한다.
- 이 subject는 original evidence commit의 descendant인 단 하나의 reachable commit일 때 충분하다.
- Integration은 해당 commit object와 ancestry를 보존하는 merge 방식이어야 한다. squash/rebase로 evidence commit을 제거하거나 ancestry를 재작성하면 resolver가 의도적으로 fail-closed하므로 금지한다.
- 이 `sync-audit.md` 갱신은 별도 report diff다. 별도 문서 commit으로 처리해야 하며 audited production blob 또는 remediation evidence commit에 섞지 않는다.

### Gaps

- 실제 live apply와 post-apply AC-HSR-022는 실행하지 않았다.
- Integration branch에서 commit 이후 clean-git exact validator는 아직 실행되지 않았다. 현재 resolver의 clean/remediation/merge fixture와 dirty-union exact validator가 그 사전 조건을 검증했다.

### Residual-risk

Coverage 여유는 `0.004%`로 사실상 한 statement 미만이다. Production statement 한 개가 미커버 상태로 추가되면 gate가 실패하므로 remediation commit 직전과 ancestry-preserving integration 직후 exact validator를 다시 실행해야 한다.

---

## Iteration 7 — F15 closure authoritative verdict

Overall Verdict: **PASS — 100/100**

Findings: **0**

### Claim

F15는 닫혔다. Production coverage runner가 시작하는 다섯 child command 모두 `MOAI_HOME_STATE_COVERAGE_CHILD=1`을 받으며 coverage driver self-recursion이 차단된다. F1–F14 smoke도 회귀 없이 통과했다.

### Evidence

```text
command: MOAI_HOME=<temp> go test ./internal/cli -run '^TestCoverageRunnerSetsRecursionGuardOnEveryChild$' -count=1 -v
exit: 0
output:
=== RUN   TestCoverageRunnerSetsRecursionGuardOnEveryChild
--- PASS: TestCoverageRunnerSetsRecursionGuardOnEveryChild (1.80s)
PASS
ok github.com/modu-ai/moai-adk/internal/cli 2.790s
```

영구 fake-go test는 guard log의 entry가 정확히 5개인지 확인하고 각 값이 모두 `1`인지 반복 assertion한다.

```text
command: HOME=<temp> env -u MOAI_HOME go test ./internal/cli -run '^TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite$' -count=1 -v
exit: 0
output:
=== RUN   TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite
    home_state_coverage_test.go:341: auto-diff changed production coverage: 1197/1408 = 85.014%
--- PASS: TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite (78.98s)
PASS
ok github.com/modu-ai/moai-adk/internal/cli 79.895s
```

```text
command: MOAI_HOME=<temp> go test ./internal/cli ./internal/homestate ./internal/hook -run '<F1-F14 concise smoke>' -count=1
exit: 0
output:
ok github.com/modu-ai/moai-adk/internal/cli 9.315s
ok github.com/modu-ai/moai-adk/internal/homestate 3.830s
ok github.com/modu-ai/moai-adk/internal/hook 2.206s
```

### Baseline-attribution

현재 branch `WT-home-state-rollout`, HEAD `18f446d579dcd5f254ff9ef7c2be30366b72505c`와 uncommitted remediation diff에서 직접 측정했다. 임시 HOME/MOAI_HOME만 사용했고 live apply는 실행하지 않았다.

### Gaps

- 실제 live apply와 post-apply AC-HSR-022는 실행하지 않았다.
- Integration branch clean-git 검증은 ancestry-preserving merge 뒤 별도로 실행해야 한다.

### Residual-risk

Coverage 여유는 `0.014%`로 매우 작다. Production 변경이 추가되면 exact validator를 즉시 재실행해야 한다. 이 보고서 갱신은 별도 report diff이며 audited production blob에 포함하지 않는다.

---

## Iteration 6 — authoritative post-merge verdict

Overall Verdict: **FAIL — 71/100**
Findings: **1 blocking**

| Dimension | Score | Verdict | Evidence |
|-----------|------:|---------|----------|
| Functionality (40%) | 50/100 | FAIL | verified-live production coverage 호출이 coverage test를 한 단계 재호출한다. |
| Security (25%) | 100/100 | PASS | 재귀는 apply를 우회하지 않고 실패 시 fail-closed한다. |
| Craft (20%) | 75/100 | FAIL | 85.004% coverage는 통과하지만 child recursion control이 test caller에만 있다. |
| Consistency (15%) | 75/100 | FAIL | production runner와 test-only guard의 책임 위치가 어긋난 localized deviation이다. |

가중 점수: `50×0.40 + 100×0.25 + 75×0.20 + 75×0.15 = 71.25`, 반올림 `71/100`.

### Findings

- **F15 [Medium] [blocking]** `internal/cli/home_state_coverage.go:35-66`, `internal/cli/home_state_coverage_test.go:329-333` — recursion guard는 outer test 함수가 `t.Setenv`로만 설정한다. Production `validateLivePreApply → measureChangedSurfaceCoverage → runChangedSurfaceCoverageSuite` 경로는 env를 설정하지 않은 채 child `go test`를 시작하고, selector에 `TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite` 자체가 포함된다. 따라서 child test가 coverage suite를 한 단계 다시 실행한다. Fake `go` auditor fixture로 child env에 `MOAI_HOME_STATE_COVERAGE_CHILD=1`을 요구했을 때 production runner가 exit 92를 반환해 결함을 재현했다. **Confidence: High. Impact:** verified-live validation이 불필요한 중첩 full suite를 실행하여 4분 deadline/불안정 timing failure 가능성을 키운다. Apply는 fail-closed하므로 data corruption은 없지만 AC-HSR-021 live gate를 안정적으로 완료하지 못할 수 있다. **Required fix:** `runChangedSurfaceCoverageSuite`가 시작하는 모든 child command의 env에 recursion marker를 직접 주입하거나 selector에서 coverage driver test를 제외한다. Production entry에서 child가 재귀 driver를 실행하지 않는 영구 fake-runner test를 추가한다. **Merge-blocking: yes.**

### Evidence

```text
command: MOAI_HOME=<temp> go test -overlay=/private/tmp/t592_iter2_cli_overlay.json ./internal/cli -run '^TestAuditCoverageRunnerSetsChildRecursionGuard$' -count=1 -v
exit: 1
output:
=== RUN   TestAuditCoverageRunnerSetsChildRecursionGuard
    t592_iter2_cli_probe_test.go:50: coverage child launched without recursion guard
--- FAIL: TestAuditCoverageRunnerSetsChildRecursionGuard (0.33s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/cli 1.352s
FAIL
```

### Verified non-findings

- committed resolver의 clean original/remediation descendant/merge/later unrelated/dirty union matrix는 GREEN이다.
- stale blob, duplicate markers, non-descendant remediation, unreachable evidence, invalid Git 및 malformed profile은 fail-closed한다.
- 외부 base flag/env 입력은 0건이다.
- exact dirty-union coverage는 `1196/1407 = 85.004%`다.
- F1–F14 smoke는 regression 없이 GREEN이다.
- 미래 subject `fix(state): stabilize committed coverage evidence (t592)`는 단일 reachable descendant commit이며 ancestry-preserving merge될 때 충분하다. squash/rebase는 금지한다.

### Gaps and residual risk

실제 live apply와 post-apply AC-HSR-022는 실행하지 않았다. 이 report 갱신은 별도 diff이며 audited production blob에 섞으면 안 된다. F15 수정 뒤 exact coverage는 한 statement 여유도 없으므로 반드시 다시 측정해야 한다.

---

## Final authoritative closure

**PASS — 100/100, findings 0.** Iteration 7이 최종 판정이며 앞선 FAIL은 감사 이력이다. `TestCoverageRunnerSetsRecursionGuardOnEveryChild`는 다섯 child 모두의 guard를 확인해 통과했고, F1–F14 smoke도 통과했다. Exact coverage는 `1197/1408 = 85.014%`다. 실제 live apply와 post-apply AC-HSR-022는 실행하지 않았다.
