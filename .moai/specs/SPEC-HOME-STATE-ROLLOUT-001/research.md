# research.md — SPEC-HOME-STATE-ROLLOUT-001

## §A Research Scope and Baseline

조사는 worktree `.claude/worktrees/t592`, branch `WT-home-state-rollout`, plan-time HEAD
`6ea69661c`에서 수행했다. 선행 홈 상태 구현은 commit `449b1c993`에 있으며 이 worktree는
이를 parent로 포함한다.

### Plan-time read-only live census

```text
command: sqlite3 'file:/Users/goos/MoAI/moai-adk-go/.moai/state/todo/backlog.db?immutable=1'
         'PRAGMA integrity_check; SELECT count(*) FROM items;'
output:
ok
75

command: existence checks
output:
~/.moai/backlog.db = ABSENT
~/.moai/db/moai-adk-go-1bd3d038/todo/backlog.db = ABSENT
```

선행 보고서의 직전 관측값은 `74`였다. 이번 읽기 전용 재측정은 `75`이므로 queue는
계획 중에도 변한다. 따라서 rollout AC는 74나 75를 hard-code하지 않고 **apply 직전
fresh count와 apply 뒤 target count**를 비교해야 한다.

## §B Existing Capability Map

### B.1 Home paths and databases

- `internal/homestate/paths.go`는 linked worktree를 Git common directory의 primary root로
  canonicalize하고 `~/.moai/db/<project-key>/{todo,factory}` 및
  `~/.moai/run/<project-key>`를 계산한다.
- `internal/homestate/factory.go`의 schema version은 현재 1이다. `resume_handoffs`에는
  `claim_token`, `claimed_at`은 있으나 claim expiry와 owner PID/session은 없다.
- Factory DB는 WAL, busy timeout, one-connection 및 DB/WAL/SHM 0600 보정을 이미 갖는다.

### B.2 Todo relocation

- `internal/kanban`은 source/target lock, legacy key/path adoption, SQLite logical store,
  incomplete target removal과 source preservation을 이미 구현한다.
- 이 seam은 migration engine이 별도 변환기를 만들지 않고 재사용할 가장 싼 경로다.
- 읽기 전용 census에는 WAL sidecar를 만들지 않는 immutable/read-only 접근이 필요하다.

### B.3 Runtime registries

- `internal/session/registry.go`는 project-local `active-sessions.json`과 advisory lock을
  제공한다. SessionStart hook subprocess PID가 아니라 실제 session PID를 기록한다.
- Factory workers는 project-scoped `factory.db`에 등록된다.
- MCP server는 serving 직전 project-local runtime stamp를 쓰지만 현재 stamp failure를
  fail-open한다. Migration admission은 별도 안전 경계이므로 marker/registry가 unreadable할
  때 동일한 fail-open 정책을 재사용하면 안 된다.

### B.4 Launcher and profile cleanup

- `internal/cli/launcher.go`는 profile directory materialization과 launch ledger 기록 뒤
  Claude process를 시작한다. POSIX에서는 exec 뒤 코드가 실행되지 않으므로 lease는 그 전에
  기록해야 한다.
- `internal/cli/clean_home.go`는 projects 180일, debug 30일, 5GiB cap과 carve-out을
  적용하지만 active profile 판정은 현재 프로세스의 `CLAUDE_CONFIG_DIR` 하나뿐이다.
- 전역 profile lease를 project Factory DB에 넣으면 동일 profile을 사용하는 다른 project의
  lease를 놓치므로 global registry가 필요하다.

### B.5 Handoff

- `SaveResume`은 새 pending 저장 시 기존 pending을 clear하고 latest-pending invariant를
  유지한다.
- `ClaimPendingResume`은 pending→claimed CAS로 단일 consumer를 고르지만 claimed row가
  중단되면 자동 재청구할 expiry가 없다.
- 현재 일반 status transition은 claim token을 최종 predicate에 포함하지 않는다. Expiry와
  reclaim을 추가하려면 token CAS가 없을 때 ABA가 발생한다.
- v1 `claimed` row에 nullable expiry column만 추가하면 SQL의 NULL 비교 특성 때문에 만료
  후보가 되지 않는다. `claimed_at` backfill 또는 명시적 indeterminate recovery가 필요하다.

## §C Root-cause Analysis

1. 왜 live migration이 아직 안 됐는가? 자동 relocation은 user-home mutation의 근거와
   backup 증거를 제공하지 못하기 때문이다.
2. 왜 census 두 번만으로 충분하지 않은가? 두 census 사이에 새 runtime이 등록될 수 있기 때문이다.
3. 왜 marker만으로 충분하지 않은가? start가 marker를 확인한 직후 migration이 marker를 만들고,
   start가 뒤늦게 등록할 수 있기 때문이다.
4. 왜 동일 lock이 필요한가? marker check와 start registration을 migration marker write와
   직렬화해야 TOCTOU가 닫히기 때문이다.
5. 왜 profile lease는 전역이어야 하는가? profile directory의 lifecycle이 project보다 넓기 때문이다.

## §D Assumptions and Resolutions

| Assumption | Evidence | Resolution |
|------------|----------|------------|
| Search DB도 옮겨야 한다 | path helper만 있고 이 scope의 runtime producer 없음 | not-applicable 보고, 구현 제외 |
| source count는 74로 고정 | fresh immutable query는 75 | dynamic pre-apply count를 기준으로 전환 |
| PID만 보면 stale 판정 가능 | PID reuse 가능 | process-start fingerprint를 함께 사용 |
| expired claim이면 exactly-once 가능 | injection↔finish 사이 crash 존재 | at-least-once 명시, receiver dedupe deferred |
| MCP stamp failure는 계속 fail-open 가능 | migration 중 신규 serving은 data race | migration admission에 한해 fail-closed |

## §E Evidence Gaps to Close in Run Phase

- 현재 OS별 process-start fingerprint API와 권한 실패 동작의 직접 측정.
- Windows에서 launcher가 exec인지 child-process start인지의 현재 실제 경로.
- SQLite-consistent backup primitive의 chosen driver behavior와 restore proof.
- Marker crash 뒤 operator recovery/rollback의 exit-code mapping과 atomic replace 증거.
- 실제 apply 직전의 active census와 source count; plan-time 값은 freshness가 만료된다.

## §G Plan Audit D1–D6 Resolution Record

| Audit defect | Resolution in this revision |
|--------------|-----------------------------|
| D1 | 모든 AC를 future green-path regression guard로 정직하게 분류하고 독립 named selector, 실행 수 계약, AC-009 실제 baseline, AC-022 validator를 고정했다. |
| D2 | v1 invalid/NULL row는 기존 claimed와 additive legacy flag로 식별하고 기존 pending/failed로만 CAS한 뒤 정상 claim하도록 고정했다. |
| D3 | SessionStart가 serialized `continue:false`를 반환하고 host 후속 처리가 0건인 계약을 고정했다. |
| D4 | recover/rollback entry point, owner refusal, restore, target 보존, marker-last 및 no-op을 정의했다. |
| D5 | pre-apply/nonce CAS → admission marker(first mutation) → backup(first data mutation) → data apply → non-self-referential AC-022/readback 순서로 고정했다. |
| D6 | parent-token CAS 기반 provisional→child handoff와 cleaner 경쟁 시 보호 계약을 정의했다. |

## §F Recommended Reuse Boundaries

- 홈 경로와 project key: `internal/homestate` SSOT 재사용.
- Todo 변환: 기존 backlog logical load/write/relocation 재사용.
- Cross-platform locking: session/kanban의 기존 build-tag lock pattern 재사용.
- Process identity: session PID/ancestry platform files를 확장하되 lease 전용 복제 금지.
- CLI tree: 기존 `migrate` parent 아래 child 추가; 별도 top-level command 금지.
