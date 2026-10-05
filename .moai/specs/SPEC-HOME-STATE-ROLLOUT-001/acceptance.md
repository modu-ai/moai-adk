# acceptance.md — SPEC-HOME-STATE-ROLLOUT-001

## §A Acceptance Contract

모든 시나리오는 독립적으로 판정한다. Plan phase는 구현 test 파일을 소유하지 않으므로
모든 AC를 `regression-guard`로 채택하고 현재 PASS를 주장하지 않는다. Manager-develop은
각 행의 named test/validator를 production 변경보다 먼저 작성해 RED를 보존하고 같은 exact
command가 test 실행 수 1 이상으로 GREEN이 되기 전 milestone을 닫지 않는다. 각 command는
다른 AC와 selector를 공유하지 않으며 comment/string 존재만 검사하지 않는다. Evidence
ledger는 current HEAD, command, executed-test count, exit code와 stdout/stderr 원문을 보존한다.
Live rollout은 AC-HSR-021의 단일 gate 외의 경로로 실행하지 않는다.

## §B Given–When–Then Scenarios

- **AC-HSR-001** (`REQ-HSR-001`, `regression-guard`) — Given source가 있고 target이 없는 격리된 `MOAI_HOME`, When `moai migrate home-state`를 flag 없이 실행, Then exit 0이고 census만 출력하며 source/target/backup/marker의 실행 전후 inventory가 같다.
- **AC-HSR-002** (`REQ-HSR-002`, `regression-guard`) — Given 동일 fixture, When dry-run 실행, Then canonical root, project key, source/target, active census, logical count, integrity와 정확한 `not-applicable (no runtime producer)`를 출력한다.
- **AC-HSR-003** (`REQ-HSR-003`, `regression-guard`) — Given active registry 하나가 unreadable 또는 indeterminate, When `--apply` 실행, Then non-zero이며 target/backup/marker는 변하지 않고 차단 census를 보고한다.
- **AC-HSR-004** (`REQ-HSR-004`, `regression-guard`) — Given determinate zero-active census와 valid source, When apply가 첫 target mutation에 도달, Then 그 전에 private backup/manifest, SHA-256 검증과 restore probe가 성공한다.
- **AC-HSR-005** (`REQ-HSR-005`, `regression-guard`) — Given nonempty target이 source와 한 row라도 다름, When dry-run과 apply 실행, Then divergence를 보고하고 apply는 target을 byte-identical하게 보존한다.
- **AC-HSR-006** (`REQ-HSR-006`, `regression-guard`) — Given target write 또는 parity fault, When apply 종료, Then 성공을 보고하지 않고 불완전 target을 채택하지 않으며 source와 backup은 intact이다.
- **AC-HSR-007** (`REQ-HSR-007`, `regression-guard`) — Given successful apply, When filesystem을 다시 읽음, Then source와 backup/manifest가 모두 남아 있고 source delete 또는 global clean이 없다.
- **AC-HSR-008** (`REQ-HSR-008`, `regression-guard`) — Given source/target logical parity와 integrity `ok`, When apply를 다시 실행, Then mutation 없는 no-op이고 digest와 count가 같다.
- **AC-HSR-009** (`REQ-HSR-009`, `regression-guard`) — Given primary checkout와 linked worktree, When 양쪽에서 dry-run, Then project key, target DB, barrier path가 동일하다.
- **AC-HSR-010** (`REQ-HSR-010`, `regression-guard`) — Given marker active 또는 unreadable, When SessionStart, Factory registration, MCP start를 각각 시도, Then SessionStart stdout은 유효한 단일 JSON이며 `continue:false`와 nonempty `stopReason`을 포함하고 host prompt/tool 처리는 0건이다. 다른 두 경로도 registration/serve 전 중단하며 census가 증가하지 않는다. Warning-only, exit-2-only, `systemMessage`-only mutant는 실패해야 한다.
- **AC-HSR-011** (`REQ-HSR-011`, `regression-guard`) — Given start와 migration이 동일 admission lock에 경쟁, When cleanup-guaranteed concurrency test를 100회 실행, Then start 완전 등록 뒤 migration 거부 또는 marker 기록 뒤 start 거부만 발생하며 동시 진입은 0건이다. Marker check와 register 사이 lock을 푸는 mutant는 실패해야 한다.
- **AC-HSR-012** (`REQ-HSR-012`, `regression-guard`) — Given dead-owner marker와 unreadable marker, When runtime start 또는 apply 재시도, Then 자동 삭제 없이 fail-closed하고 recovery metadata를 출력한다.
- **AC-HSR-013** (`REQ-HSR-013`, `regression-guard`) — Given 실제 v1 DDL에 pending, parseable-`claimed_at` claimed, NULL/invalid-`claimed_at` claimed row가 함께 있음, When v2 open, Then pending과 payload/token provenance는 보존되고 parseable row expiry가 transaction 안에서 backfill되며 정해진 만료 뒤 정확히 한 consumer가 reclaim한다. NULL/invalid row는 `claimed`를 유지하면서 legacy-recovery flag/reason으로 조회되고 정상 claim에서 제외되며 index/schema version도 존재한다. 허용된 기존 status 이외의 값을 쓰거나 NULL을 expired로 간주하거나 조용히 숨기는 mutant는 실패해야 한다.
- **AC-HSR-014** (`REQ-HSR-014`, `regression-guard`) — Given 오래된 expired claimed와 더 최신 pending, When 여러 consumer가 claim, Then 하나만 최신 pending을 받고 pending 소진 뒤 하나만 expired row를 새 token으로 reclaim한다.
- **AC-HSR-015** (`REQ-HSR-015`, `regression-guard`) — Given token A row가 token B로 reclaim, When A가 finish, Then CAS 0-row/stale이며 B의 status/owner가 같다. Finish predicate에서 token을 제거한 mutant는 실패해야 한다.
- **AC-HSR-016** (`REQ-HSR-016`, `regression-guard`) — Given injection 성공 뒤 consumed CAS 전 crash, When 만료 후 reclaim, Then 같은 payload가 다시 반환되고 verdict가 이 경계를 at-least-once duplicate로 명시한다.
- **AC-HSR-017** (`REQ-HSR-017`, `regression-guard`) — Given 두 project의 named profile, When leases 기록, Then 모두 `~/.moai/run/profile-leases.db`에 있고 project Factory DB에는 없으며 DB/WAL/SHM은 0600이다.
- **AC-HSR-018** (`REQ-HSR-018`, `regression-guard`) — Given POSIX exec, direct Claude, non-exec child start fixtures, When 각각 시작, Then launcher는 exec/start 전 provisional lease를 만들고 SessionStart가 enrich하거나 direct lease를 만든다. Non-exec pause 지점에서 parent 종료와 child identity 기록 사이 concurrent cleaner를 실행해도 profile 삭제는 0건이며 parent-token CAS 뒤 child fingerprint가 보존된다. Grace 만료·fingerprint unreadable·CAS 충돌은 indeterminate이다. Provisional row를 parent death만으로 stale 처리하는 mutant는 실패해야 한다.
- **AC-HSR-019** (`REQ-HSR-019`, `regression-guard`) — Given normal SessionEnd, killed PID, reused PID/different start, unreadable fingerprint, When reconcile, Then release/stale/indeterminate를 구분한다.
- **AC-HSR-020** (`REQ-HSR-020`, `regression-guard`) — Given aged/over-cap profile과 live 또는 indeterminate lease, When `moai clean --home --force`, Then profile 파일 삭제는 0건이고 skip reason이 있다. Stale-only는 기존 allowlist 조건에서만 후보이다.
- **AC-HSR-021** (`REQ-HSR-021`, `regression-guard`) — Given 승인된 current project와 pre-apply validator set AC-001..021/023..025, When 정확히 `moai migrate home-state --apply --verified-live` 실행, Then 같은 process가 current HEAD에서 각 validator의 executed-test count >0과 GREEN, strict lint, race/coverage/vet/native, Windows disposition 및 fresh determinate zero-active census를 검증하고 HEAD+census fingerprint에 묶인 internal nonce를 in-memory CAS로 한 번 소비한다. 그 다음 exclusive admission marker를 첫 mutation으로 설치해 SessionStart/Factory/MCP 신규 진입을 차단하고, backup을 첫 data mutation으로 생성·검증한 뒤 data apply한다. Missing/failing/zero-test validator, HEAD/census change, injected stale/tampered/replayed nonce와 bare apply는 marker/backup/target inventory를 byte-identical하게 보존하며 non-zero이다. Prior 74/75와 external ledger는 authorization input이 아니다.
- **AC-HSR-022** (`REQ-HSR-022`, `regression-guard`) — Given AC-021 apply가 종료되어 authorization phase가 닫힘, When 독립 post-apply `TestHomeStateVerdictEvidenceValidator`와 source/target readback을 실행, Then validator는 AC-001..021/023..025 evidence와 post-apply readback만 입력으로 받아 fresh source/target count와 digest, 양쪽 integrity `ok`, source/backup 존재, 24개 선행 AC의 command/output/exit/HEAD, race, coverage ≥85%, vet, native, Windows disposition과 신규 test skip 0건을 확인한다. External run harness는 validator 종료 뒤 AC-022의 command/stdout/exit/HEAD를 ledger에 append하고 독립 sync audit가 전체 25개 completeness를 확인한다. AC-022 결과는 AC-022 자신의 입력이나 pre-apply authorization으로 사용되지 않는다.
- **AC-HSR-023** (`REQ-HSR-023`, `regression-guard`) — Given mid-apply crash marker, verified backup, incomplete new target, When `moai migrate home-state recover --migration-id <id> --backup-id <id>`, Then live owner 또는 indeterminate fingerprint는 mutation 전 거부하고, dead owner에서는 manifest/hash/restore probe 후 incomplete target만 격리해 source 또는 backup과 parity/integrity를 복구하고 marker를 마지막에 제거한다. Marker를 먼저 지우는 mutant는 실패하며 반복 실행은 no-op이다.
- **AC-HSR-024** (`REQ-HSR-024`, `regression-guard`) — Given completed apply와 선택한 verified backup, When `moai migrate home-state rollback --migration-id <id> --backup-id <id>`, Then pre-existing target을 timestamped quarantine으로 보존하고 temporary restore 검증 뒤 atomic replace하며 post-restore parity/integrity 뒤 marker를 마지막에 제거한다. Hash mismatch나 indeterminate owner는 mutation 전 거부하고 반복 실행은 no-op이다.
- **AC-HSR-025** (`REQ-HSR-025`, `regression-guard`) — Given v2 upgrade가 NULL/invalid lease의 legacy row를 `claimed`+legacy-recovery flag로 식별, When live owner, indeterminate fingerprint, dead owner, unknown-owner+zero-active census fixtures에서 `moai factory handoff recover-resume`를 실행, Then live/indeterminate는 row mutation 없이 거부되고 dead/verified-unknown fixture의 승인된 requeue만 `id+claimed+expected-token+flag` CAS로 기존 `pending` status가 된다. 이후 정상 claim 한 건만 새 token/expiry로 payload/provenance를 받고 old token finish는 거부된다. `--decision fail`은 같은 CAS 조건으로 기존 `failed` status가 되며 전달되지 않는다.

## §C Adoption and Evidence Ledger

Plan phase에 production test를 추가할 권한이 없으므로 AC-001..025를 모두
`regression-guard / future green-path`로 채택한다. 아래 command는 각각 하나의 named Go test
또는 validator만 선택하는 독립 command다. Manager-develop은 대응 production 변경 전에 그
test를 작성하고 `-v` 출력의 `=== RUN <exact-name>` 한 건 이상과 의도한 RED를 기록해야 한다.
`[no tests to run]` 또는 source-text 존재 검사만으로는 adoption evidence가 아니다.

| AC | Named test or validator | Exact independent command |
|----|-------------------------|---------------------------|
| 001 | `TestHomeStateDryRunNoMutation` | `go test ./internal/cli -run '^TestHomeStateDryRunNoMutation$' -count=1 -v` |
| 002 | `TestHomeStateDryRunReport` | `go test ./internal/cli -run '^TestHomeStateDryRunReport$' -count=1 -v` |
| 003 | `TestHomeStateApplyCensusFailClosed` | `go test ./internal/cli -run '^TestHomeStateApplyCensusFailClosed$' -count=1 -v` |
| 004 | `TestHomeStateBackupBeforeWrite` | `go test ./internal/cli -run '^TestHomeStateBackupBeforeWrite$' -count=1 -v` |
| 005 | `TestHomeStateRefusesDivergentTarget` | `go test ./internal/cli -run '^TestHomeStateRefusesDivergentTarget$' -count=1 -v` |
| 006 | `TestHomeStateApplyFaultPreservesSource` | `go test ./internal/cli -run '^TestHomeStateApplyFaultPreservesSource$' -count=1 -v` |
| 007 | `TestHomeStateApplyPreservesSourceAndBackup` | `go test ./internal/cli -run '^TestHomeStateApplyPreservesSourceAndBackup$' -count=1 -v` |
| 008 | `TestHomeStateApplyIdempotentNoOp` | `go test ./internal/cli -run '^TestHomeStateApplyIdempotentNoOp$' -count=1 -v` |
| 009 | existing `TestResolveTodoQueueRoot_WorktreeConvergesOnPrimary` | `go test ./internal/kanban -run '^TestResolveTodoQueueRoot_WorktreeConvergesOnPrimary$' -count=1 -v` |
| 010 | `TestHomeStateBarrierAdmissionHaltsAllHosts` | `go test ./internal/cli -run '^TestHomeStateBarrierAdmissionHaltsAllHosts$' -count=1 -v` |
| 011 | `TestHomeStateStartVsMigrateSerialized` | `go test -race ./internal/cli -run '^TestHomeStateStartVsMigrateSerialized$' -count=1 -v` |
| 012 | `TestHomeStateCrashMarkerFailsClosed` | `go test ./internal/cli -run '^TestHomeStateCrashMarkerFailsClosed$' -count=1 -v` |
| 013 | `TestFactoryV1ClaimedRowsUpgradeToV2` | `go test ./internal/homestate -run '^TestFactoryV1ClaimedRowsUpgradeToV2$' -count=1 -v` |
| 014 | `TestResumeLatestPendingThenExpiredReclaim` | `go test -race ./internal/homestate -run '^TestResumeLatestPendingThenExpiredReclaim$' -count=1 -v` |
| 015 | `TestResumeFinishRejectsABAToken` | `go test ./internal/homestate -run '^TestResumeFinishRejectsABAToken$' -count=1 -v` |
| 016 | `TestResumeInjectionCrashIsAtLeastOnce` | `go test ./internal/homestate -run '^TestResumeInjectionCrashIsAtLeastOnce$' -count=1 -v` |
| 017 | `TestProfileLeasesAreGlobalAndPrivate` | `go test ./internal/homestate -run '^TestProfileLeasesAreGlobalAndPrivate$' -count=1 -v` |
| 018 | `TestProfileLeaseLifecycleAndNonExecCleanerRace` | `go test -race ./internal/cli -run '^TestProfileLeaseLifecycleAndNonExecCleanerRace$' -count=1 -v` |
| 019 | `TestProfileLeaseReconcilePIDFingerprint` | `go test ./internal/homestate -run '^TestProfileLeaseReconcilePIDFingerprint$' -count=1 -v` |
| 020 | `TestCleanHomeSkipsLiveAndIndeterminateProfiles` | `go test ./internal/cli -run '^TestCleanHomeSkipsLiveAndIndeterminateProfiles$' -count=1 -v` |
| 021 | `TestHomeStateVerifiedLiveGateCannotBypassOrReplay` | `go test ./internal/cli -run '^TestHomeStateVerifiedLiveGateCannotBypassOrReplay$' -count=1 -v` |
| 022 | `TestHomeStateVerdictEvidenceValidator` | `go test ./internal/cli -run '^TestHomeStateVerdictEvidenceValidator$' -count=1 -v` |
| 023 | `TestHomeStateRecoverCrashMarkerSafely` | `go test ./internal/cli -run '^TestHomeStateRecoverCrashMarkerSafely$' -count=1 -v` |
| 024 | `TestHomeStateRollbackVerifiedBackup` | `go test ./internal/cli -run '^TestHomeStateRollbackVerifiedBackup$' -count=1 -v` |
| 025 | `TestResumeLegacyIndeterminateOperatorRecovery` | `go test ./internal/cli -run '^TestResumeLegacyIndeterminateOperatorRecovery$' -count=1 -v` |

### §C.1 Existing green-path baseline for AC-HSR-009

Current HEAD `6ea69661c`에서 실제 존재하는 selector를 실행했고 test 실행 수는 1이다.

```text
command: go test ./internal/kanban -run '^TestResolveTodoQueueRoot_WorktreeConvergesOnPrimary$' -count=1 -v
exit: 0
stdout:
=== RUN   TestResolveTodoQueueRoot_WorktreeConvergesOnPrimary
--- PASS: TestResolveTodoQueueRoot_WorktreeConvergesOnPrimary (0.13s)
PASS
ok  github.com/modu-ai/moai-adk/internal/kanban  0.600s
```

AC-HSR-021의 pre-apply set은 AC-001..021과 AC-023..025의 named validator이며 AC-022를
명시적으로 제외한다. AC-HSR-022의 post-apply validator 이름은
`TestHomeStateVerdictEvidenceValidator`로 고정한다. 이는 apply 종료 후 AC-001..021/023..025의
exact command, HEAD, executed-test count, exit, verbatim output, RED→GREEN pair와 live readback만
검사한다. External run harness가 종료된 AC-022 결과를 append한 뒤 독립 sync audit가 전체
25개 completeness를 확인하므로 validator는 자신의 미래 결과를 입력으로 요구하지 않는다.

## §D Edge Cases

- Source absent + target absent: dry-run no-op, apply도 mutation 없음.
- Source absent + equivalent target: already migrated로 보고하되 provenance gap을 명시.
- Source integrity failure: backup/target write 전 중단.
- Same PID/different start: stale; process start unreadable: indeterminate.
- Claim expiry와 new save 동시 발생: 최신 eligible pending 우선.
- NULL/invalid legacy `claimed_at`: `claimed`+legacy flag로 식별하고 explicit pending/failed CAS 전에는 전달하지 않는다.
- Worktree path가 symlink 포함: canonical key와 barrier는 primary와 동일.
- Marker removal 실패: 성공이 아니며 이후 start는 계속 fail-closed.

## §E Traceability Matrix

| Requirement | Acceptance |
|-------------|------------|
| REQ-HSR-001..009 | AC-HSR-001..009 |
| REQ-HSR-010..012 | AC-HSR-010..012 |
| REQ-HSR-013..016 | AC-HSR-013..016 |
| REQ-HSR-017..020 | AC-HSR-017..020 |
| REQ-HSR-021..025 | AC-HSR-021..025 |

Machine-readable 1:1 mappings:

- AC-HSR-001 maps REQ-HSR-001
- AC-HSR-002 maps REQ-HSR-002
- AC-HSR-003 maps REQ-HSR-003
- AC-HSR-004 maps REQ-HSR-004
- AC-HSR-005 maps REQ-HSR-005
- AC-HSR-006 maps REQ-HSR-006
- AC-HSR-007 maps REQ-HSR-007
- AC-HSR-008 maps REQ-HSR-008
- AC-HSR-009 maps REQ-HSR-009
- AC-HSR-010 maps REQ-HSR-010
- AC-HSR-011 maps REQ-HSR-011
- AC-HSR-012 maps REQ-HSR-012
- AC-HSR-013 maps REQ-HSR-013
- AC-HSR-014 maps REQ-HSR-014
- AC-HSR-015 maps REQ-HSR-015
- AC-HSR-016 maps REQ-HSR-016
- AC-HSR-017 maps REQ-HSR-017
- AC-HSR-018 maps REQ-HSR-018
- AC-HSR-019 maps REQ-HSR-019
- AC-HSR-020 maps REQ-HSR-020
- AC-HSR-021 maps REQ-HSR-021
- AC-HSR-022 maps REQ-HSR-022
- AC-HSR-023 maps REQ-HSR-023
- AC-HSR-024 maps REQ-HSR-024
- AC-HSR-025 maps REQ-HSR-025

## §F Quality Gates

- Functionality: AC-HSR-001..025 모두 PASS.
- Tested: 모든 named selector가 test 1건 이상을 실행하고 올바른 이유의 RED와 대응 GREEN, 새 skip 0건.
- Security: containment, private modes, parameterized SQL, symlink/PID reuse PASS.
- Consistency: linked worktree와 primary의 key/path/barrier 동일.
- Recovery: backup restore, crash recovery, rollback, reclaim, marker-last PASS.
- Trackability: card `t592`, SPEC ID, current HEAD와 report path가 evidence에 연결됨.

## §G Definition of Done

- SPEC/AC가 구현되고 독립 sync audit가 merge-blocking defect 0을 판정한다.
- Live rollout gate가 current evidence를 소비하고 source를 보존한 parity readback을 남긴다.
- Search producer와 receiver exactly-once dedupe는 미구현으로 명시된다.
- `moai clean --home --force`를 live rollout 일부로 실행하지 않는다.
