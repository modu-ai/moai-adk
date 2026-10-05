# design.md — SPEC-HOME-STATE-ROLLOUT-001

## §A Design Summary

선택한 구조는 두 coordination domain을 분리한다.

```text
project-scoped                                  user-global
~/.moai/run/<project-key>/                      ~/.moai/run/
├── locks/home-state-migration.lock             └── profile-leases.db
└── home-state-migration.json
        │
        ├── moai migrate home-state
        ├── SessionStart admission
        ├── Factory worker admission
        └── MCP server admission

~/.moai/db/<project-key>/
├── todo/backlog.db
└── factory/factory.db
       └── resume_handoffs schema v2
```

프로젝트 이전과 runtime admission은 project key로 격리한다. 프로필 lease는 같은
Claude profile을 여러 프로젝트가 공유할 수 있으므로 전역 registry에 둔다.

## §B Alternatives Considered

| Option | Correctness | Complexity | Operability | Decision |
|--------|-------------|------------|-------------|----------|
| 각 subsystem이 자체 marker 확인 | 낮음: TOCTOU와 규칙 drift | 중간 | 낮음 | Reject |
| Factory DB 하나로 migration/profile까지 통합 | 프로젝트 밖 profile을 잘못 귀속 | 높음 | 낮음 | Reject |
| OS shared/exclusive lock만 사용 | crash marker/recovery 설명이 약함 | 중간 | 중간 | Reject |
| project lock + durable marker + global profile SQLite | 높음 | 중간 | 높음 | **Select** |

선택안은 기존 cross-platform lock 구현과 SQLite dependency를 재사용하면서, marker가
차단 이유와 crash recovery provenance를 남긴다.

## §C Migration Barrier Protocol

### C.1 Admission ordering

모든 시작 경로와 apply가 동일한 exclusive file-lock primitive를 짧게 공유한다.

```text
Runtime start                         Migration apply
-------------                         ---------------
acquire project lock                  pre-barrier read-only census
read marker                           acquire project lock
if present -> refuse                  validate HEAD/validators/census
register runtime identity             consume in-memory nonce CAS
release lock                          atomically write admission marker
                                      backup -> data apply -> verify
continue startup                      clear marker -> release lock
```

Runtime 쪽은 marker 확인과 registry 등록을 lock 안에서 끝낸다. Apply 쪽은 pre-apply 검증과
nonce CAS 뒤 같은 exclusive lock을 유지한 채 admission marker를 설치한다. Marker 설치 뒤에만
backup과 data apply를 시작하므로 신규 start는 backup/apply 전에 차단된다. 기존 runtime이
있거나 registry 하나라도 읽을 수 없으면 marker를 남기기 전에 apply를 중단한다. Marker가
생성된 뒤 실패하면 marker를 유지해 후속 시작을 fail-closed한다.

SessionStart는 경고 출력이나 exit 2에 의존하지 않는다. Marker가 active 또는 unreadable이면
stdout에 정확히 하나의 유효한 JSON object를 serialize하고 `continue:false`와 비어 있지 않은
`stopReason`을 포함한다. Host 통합 검증은 이 응답 뒤 prompt 제출과 tool dispatch가 0건임을
관측해야 한다. 가능하면 MoAI launcher도 Claude process 생성 전에 같은 admission을 수행하되,
direct Claude 경로의 안전성은 SessionStart universal halt가 담당한다.

### C.2 Marker content and permissions

Marker는 schema version, project key/root, migration ID, owner PID,
process-start fingerprint, started-at, source/target paths, phase를 기록한다. Directory는
0700, lock/marker는 0600이다. Marker 제거는 success verification 뒤 마지막 mutation이다.
Stale처럼 보여도 자동 삭제하지 않고 별도 recovery action이 backup/target/source를
검사한 뒤 해제하도록 한다.

## §D Home-state Migration Pipeline

```text
resolve canonical project
  -> census source/target/runtimes
  -> dry-run report and stop
  -> [--apply --verified-live] acquire exclusive barrier
  -> validate HEAD/pre-apply set + immediate census
  -> consume in-memory nonce CAS
  -> atomically install admission marker
  -> SQLite-consistent backup + manifest + hash + restore probe (first data mutation)
  -> refuse divergent nonempty target
  -> write/copy through existing Todo logical relocation seam
  -> compare logical digest/counts + integrity(source,target)
  -> clear marker + release
```

Backup은 열린 WAL 파일의 단순 복사가 아니라 SQLite-consistent snapshot을 사용한다.
Manifest는 backup 자체의 SHA-256뿐 아니라 source의 logical digest와 count를 담는다.
Target이 새로 생성된 run에서 실패하면 incomplete target artifacts를 제거한다. 사전에
존재하던 target은 절대 자동 삭제하거나 overwrite하지 않는다.

Idempotency는 marker 유무가 아니라 매 실행의 fresh logical parity로 판정한다. Search는
producer가 없으므로 path 생성, DB 생성, copy를 하지 않고 지정 문자열만 보고한다.

## §E Resume Handoff Schema v2

기존 table에 additive columns를 추가한다.

```sql
claim_expires_at    TEXT NULL
claim_owner_pid     INTEGER NULL
claim_owner_session TEXT NOT NULL DEFAULT ''
```

만료 후보를 조회할 index는 status와 expiry를 앞에 둔다. Schema upgrade는 한 transaction
안에서 column/index/meta version을 적용한다. v1 `pending` row는 status와 payload를 그대로
보존한다. v1 `claimed` row는 parse 가능한 `claimed_at`에 정책상 lease duration을 더해
`claim_expires_at`을 backfill하며 body, 기존 token, provenance를 보존한다. `claimed_at`이
비어 있거나 해석 불가능한 row는 upgrade transaction에서 기존 `claimed` status/token을
유지하고 additive legacy-recovery flag/reason을 설정한다. 정상 claim predicate는 이 flag가
없는 row만 허용하므로 NULL expiry가 영구히 숨거나 자동 소비되지 않는다. Handoff status는
기존 CHECK 집합인 pending, claimed, consumed, failed, expired, cleared만 사용한다.

Legacy row에는 신뢰할 owner identity가 없을 수 있다. `moai factory handoff recover-resume`는
Factory project barrier를 획득하고 active session/worker census가 determinate zero인지 확인한
후에만 unknown owner를 orphan으로 승인한다. Owner PID/fingerprint가 존재하면 dead가 확정된
경우에만 진행하고 live/indeterminate이면 거부한다. Recovery transaction은
`id + status='claimed' + expected original token + legacy flag`를 CAS 조건으로 삼는다.
`requeue`는 status를 pending으로 바꾸고 invalid timestamp/owner를 초기화한다. 이후 정상 claim이
새 token과 expiry를 발급한다. `fail`은 기존 terminal failed로 전이한다. 어느 전이든 기존
token을 사용하는 finish는 status/token predicate에서 거부된다.

Claim transaction은 다음 우선순위를 한 번에 적용한다.

1. 최신 `pending` row (`id DESC`)
2. pending이 없으면 `claim_expires_at <= now`인 최신 `claimed` row

선택된 row는 새 token, owner PID/session, expiry로 CAS 갱신한다. Finish는
`id + status='claimed' + claim_token`을 조건으로 하며 0-row update를 stale consumer로
반환한다. Crash가 external injection 뒤 DB finish 전에 발생하면 reclaim된 payload가
재주입될 수 있다. 이것이 at-least-once의 명시된 duplicate boundary다.

## §F Global Profile Lease Registry

Registry는 `~/.moai/run/profile-leases.db` 하나이며 named profile마다 복수 동시 lease를
허용한다. 최소 row는 lease token, profile canonical path/name, project key, PID,
process-start fingerprint, session ID, state(provisional/enriched/released), created/updated
timestamp를 가진다.

- Launcher: profile directory 확정 뒤, Claude exec/start 직전에 provisional lease 생성.
- SessionStart: profile path + PID/ancestry를 이용해 provisional row enrich. 매칭 row가
  없으면 direct-Claude row 생성.
- SessionEnd: session ID와 token에 맞는 lease release.
- Reconciler: PID dead 또는 같은 PID의 start fingerprint mismatch만 stale 확정.
- Cleaner: live 또는 indeterminate lease가 하나라도 있으면 profile 전체 skip.

POSIX exec는 PID를 보존한다. exec를 쓰지 않는 플랫폼에서는 provisional row의 token을
소유권 증표로 유지하고 `provisional(parent identity) -> transferring -> enriched(child identity)`
전이를 CAS로 수행한다. Child PID/fingerprint 확인과 row 갱신이 끝나기 전에는 launcher
identity가 사라져도 cleaner가 해당 row를 stale로 회수하지 않는다. 보정에는 bounded grace를
두되 grace 만료, child identity 판독 실패, CAS 충돌은 삭제 허가가 아니라 `indeterminate`로
귀결된다. SessionStart enrichment와 launcher child handoff가 경합해도 같은 token 또는
session identity를 조건으로 한 쪽만 최종 enrich하며 profile 보호는 끊기지 않는다.
Process-start fingerprint는 기존 OS별 process inspection pattern을 확장하고, 지원되지 않거나
권한 때문에 읽을 수 없으면 `indeterminate`로 처리한다.

## §G Integration Surfaces

| Surface | Reuse / Change |
|---------|----------------|
| `internal/homestate` | path SSOT, Factory schema migration, handoff CAS, global lease store |
| `internal/kanban` | existing Todo relocation/logical model, Factory worker admission |
| `internal/session` | active census, process liveness/fingerprint platform adapters |
| `internal/hook` | SessionStart admission/enrich, SessionEnd release, handoff reclaim |
| `internal/cli` | migrate child, launcher provisional lease, MCP start admission, clean filter |
| templates/docs | user-visible command and safety contract mirrors only where shipped |

## §H Rollback and Recovery

Operator entry point는 기존 migrate parent 아래의
`moai migrate home-state recover --migration-id <id> --backup-id <id>`와
`moai migrate home-state rollback --migration-id <id> --backup-id <id>`이다.

Recovery와 rollback은 모두 같은 project lock을 획득한 뒤 다음 precondition을 검사한다.

1. marker의 project key/root와 요청 migration ID가 일치한다.
2. marker owner PID와 process-start fingerprint가 live이면 거부한다. fingerprint가 읽히지
   않거나 owner 상태가 불확정이어도 거부한다.
3. 선택한 backup manifest/hash와 SQLite restore probe가 유효하고 source/target census가
   determinate하다.
4. 사전에 존재하던 target은 덮어쓰거나 삭제하지 않고 timestamped quarantine snapshot으로
   보존한다.

Mid-apply crash recovery는 새로 만든 불완전 target만 격리한 후 검증된 source 또는 backup으로
안전 상태를 복원한다. 성공 apply rollback은 선택한 backup을 임시 경로에 복원·검증한 뒤
atomic replace하고 post-restore logical parity와 `integrity_check=ok`를 다시 확인한다. 어느
경로든 marker 제거는 모든 검증과 directory sync 뒤 마지막 mutation이다. 검증 실패나 marker
제거 실패는 성공으로 보고하지 않는다. 이미 recovery/rollback이 끝난 동일 ID의 재실행은
상태를 재검증하고 mutation 없는 no-op을 보고한다.

Profile lease DB가 unreadable이면 clean은 관련 프로필을 모두 indeterminate로 보고 skip한다.
Handoff v2는 additive migration을 유지하되 구 binary downgrade는 보장하지 않는다.

## §I Executable Live Gate

Live apply의 유일한 operator entry point는
`moai migrate home-state --apply --verified-live`이다. Bare `--apply`, caller-provided success
ledger, hidden environment override 및 하위 copy helper 직접 호출은 live target mutation
authorization을 만들지 못한다.

Gate는 외부 서명 파일을 신뢰하지 않고 다음 순서를 지킨다.

### I.1 Pre-apply authorization — admission mutation 전

1. 실행 binary의 build commit과 repository current HEAD 일치를 확인하고 HEAD를 고정한다.
2. Pre-apply set인 AC-001..021/023..025의 named validator를 직접 실행하여 exact selector별
   `=== RUN` count가 1 이상이고 exit 0이며, strict lint/race/coverage/vet/native와 Windows
   disposition이 모두 유효한지 확인한다. Post-apply AC-022는 이 집합에 포함하지 않는다.
3. 사전에 생성된 project lock을 새 파일 생성 없이 획득하고 fresh determinate zero-active
   census와 source fingerprint를 읽는다. Lock이 없거나 읽을 수 없으면 mutation 없이 거부한다.
4. 외부 입력으로 받지 않은 migration ID/nonce를 생성해 current HEAD와 census/source
   fingerprint에 결합하고, same-process atomic state를 prepared→consumed로 CAS한다.
5. Missing/failed/zero-test validator, HEAD/census change, stale/tampered/replayed nonce 또는 bare
   apply를 거부한다. CAS 성공까지 marker, backup, target inventory는 byte-identical해야 한다.

### I.2 Admission marker → backup → data apply

6. Authorization CAS 성공 뒤 같은 exclusive lock 아래 admission marker를 첫 mutation으로
   원자 설치한다. 이때부터 SessionStart, Factory worker, MCP server 신규 admission이 차단된다.
7. Marker 설치 뒤 SQLite-consistent backup을 첫 data mutation으로 생성한다.
8. Backup ID/hash와 restore probe를 확인하고 authorization의 HEAD/census/source fingerprint 및
   migration ID에 결합한다. 실패하면 target을 생성하지 않고 source/backup을 보존하며, 이미
   설치한 marker는 제거하지 않아 후속 admission을 계속 차단한다.
9. Target data migration을 수행하고 logical parity/integrity를 확인한다.

### I.3 Post-apply verification — authorization input 아님

10. Apply가 끝난 뒤 별도 readback과 AC-022 `TestHomeStateVerdictEvidenceValidator`를 실행한다.
    이 validator의 입력은 AC-001..021/023..025 evidence와 post-apply readback뿐이다.
11. External run harness는 AC-022 종료 뒤 그 command/stdout/exit/HEAD를 ledger에 append한다.
    이후 독립 sync audit가 전체 25개 ledger completeness를 확인한다. AC-022는 자신의 아직
    생성되지 않은 결과를 입력으로 요구하지 않는다.
12. Post-apply 결과와 closure 판정은 verdict에 기록하지만 pre-apply CAS, nonce 또는 apply
    authorization에 역으로 입력하지 않는다.

Authorization은 same-process memory에서만 생성·소비되며 caller가 nonce나 PASS ledger를
제공할 입력 표면이 없다. 이전 output을 재제출해도 current process capability가 없으므로
apply할 수 없다. 동일 명령의 정상 재실행은 새 pre-apply authorization을 거친 뒤 logical
parity에 따라 no-op일 수 있지만 과거 nonce를 재사용하지 않는다.
