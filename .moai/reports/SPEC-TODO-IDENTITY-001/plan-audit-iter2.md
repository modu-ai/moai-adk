# SPEC Review Report: SPEC-TODO-IDENTITY-001

Iteration: 2/3
Verdict: FAIL
Overall Score: 0.75
Tier: M — PASS threshold 0.80
Previous iteration: FAIL 0.69 — score regression: no (`0.75 > 0.69`)

Reasoning context ignored per M1 Context Isolation. 이번 재감사는 iteration 1의 D1–D9 수정 delta, 그 수정으로 생긴 회귀, 현재 Tier M artifacts(`spec.md`, `plan.md`, `acceptance.md`), `red-baseline.md`, 그리고 문서가 지목한 현재 selector의 실행 결과만 사용했다.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: requirement layer는 `REQ-TID-001`…`REQ-TID-007`의 연속 7개이며 gap·duplicate·padding 불일치가 없다 (`spec.md:L32-L44`).
- [PASS] MP-2 EARS/GEARS format compliance: **requirement layer만** 판정했다. REQ-001/003/004/005/006은 `When …, the system shall …`, REQ-002는 `While …, the system shall …`, REQ-007은 `The system shall …` 패턴이다 (`spec.md:L32-L44`). verification layer의 Given-When-Then AC는 MP-2에서 감점하지 않았다.
- [PASS] MP-3 YAML frontmatter validity: canonical 12 required fields와 유효한 optional `tier: M`이 있다 (`spec.md:L2-L14`). `moai spec lint .moai/specs/SPEC-TODO-IDENTITY-001 --json`은 exit 0, stdout `[]`이었다.
- [N/A] MP-4 Section 22 language neutrality: `module: "internal/kanban"`인 single-language Go SPEC이다 (`spec.md:L11`; `plan.md:L43-L68`). universal/template-bound 다언어 표면이 아니다.
- [PASS] MP-5 D7 cross-SPEC reconciliation: `SPEC-TODO-IDENTITY-001`, `SPEC-TODO-RUNTIME-STORE-001`, `SPEC-TODO-UNIFIED-001`가 모두 존재하고 `status=draft`였다 (`spec.md:L15`, `L18`). terminal-status reference 및 D7 BLOCKING finding이 없다.
- [PASS] MP-6 D8 cross-platform discipline: `spec.md`에 literal `syscall`이 없다. D8 BLOCKING finding이 없다.
- [PASS] MP-7 clarification gate: Tier M 필수 입력인 `plan.md`에서 `[NEEDS CLARIFICATION` match는 0개(`rg` exit 1)였다. `research.md`는 Tier M 입력이 아니다. 이 결과는 kickoff 승인을 뜻하지 않는다.
- [PASS] MP-8 RED-now cell re-execution: `acceptance.md:L21`의 종료필수 7 AC 각각에 `plan.md:L41-L55`의 공통 command+literal selector, `plan.md:L73`의 document-level tree SHA 및 test SHA, `plan.md:L75-L252`/`red-baseline.md:L15-L192`의 stdout·exit·RED 이유가 있다. 현재 tree에서 7 selector를 독립 재실행했으며 AC별 top-level test 수는 1/3/2/1/2/3/1, 모두 exit 1이었다. zero-test·compile·setup·timeout 실패는 없었다. AC-006 race 보조 명령도 test 1개를 실행하고 UUID semantic assertion으로 exit 1이었으며 `WARNING: DATA RACE`는 없었다. MP-8은 command output 재현만 판정하며 command가 AC 전체 premise를 측정하는지는 보증하지 않는다; 그 별도 결함은 D10–D13에 기록한다.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|---|---:|---|---|
| Clarity | 0.75 | 0.75 | side-table DDL, media/transaction/WAL 경계, concurrency 최소 단위, kickoff state, MX anchor가 구체화됐다 (`plan.md:L7-L37`, `L59-L68`, `L255-L277`). 다만 AC-005의 “모호한 카드” 구성은 fixture까지 고정되지 않았고 (`acceptance.md:L11`, `plan.md:L279`), 일부 GREEN 계측 범위가 여전히 구현자 해석에 남았다. |
| Completeness | 0.75 | 0.75 | Tier M 3-artifact set, HISTORY/WHY/WHAT/REQ/AC/Out-of-Scope, 7 AC RED ledger, migration/rollback, concurrency, MX 계획은 모두 존재한다 (`spec.md:L20-L78`; `plan.md:L31-L279`). 하지만 문서가 직접 인정한 schema stamp/retry/ambiguous/combined sequence/OwnerLabel/provenance gap이 release-blocking AC instrument에 남아 있다 (`plan.md:L57`, `L279`; `acceptance.md:L19-L25`). |
| Testability | 0.50 | 0.50 | 모든 AC는 GWT이며 selector도 nonzero RED를 재현했지만, 7개 중 AC-001/003/005/007 verification instrument는 acceptance의 일부 clause를 전혀 assert하지 않는다. 이 상태의 future GREEN은 cross-project collision, retry/stamp rollback, ambiguous/state-authority, OwnerLabel/provenance 회귀를 놓칠 수 있다 (`todo_identity_red_test.go:L147-L189`, `L261-L321`, `L371-L450`, `L572-L615`). |
| Traceability | 1.00 | 1.00 | AC-001…007은 각각 존재하는 같은 번호 REQ를 정확히 한 번 매핑한다 (`acceptance.md:L3-L15`). orphan AC와 uncovered REQ가 없다. |

산술 평균: `(0.75 + 0.75 + 0.50 + 1.00) / 4 = 0.75`. Tier M threshold 0.80 미달이며 D10–D13의 blocking verification gaps가 남아 있어 FAIL이다. 이전 0.69보다 상승했으므로 STOP score-regression 신호는 없다.

## Regression Check

Defects from iteration 1:

- D1 MP8-REDNOW-MISSING — [RESOLVED]: 7개 AC별 selector·nonzero count·stdout·exit·tree/test pin이 생겼고 모두 현재 tree에서 semantic RED로 재현됐다 (`plan.md:L41-L252`; `red-baseline.md:L9-L219`). 단, criterion 전체 clause를 측정하지 않는 새 blocking gaps는 D10–D13으로 분리한다.
- D2 KICKOFF-STATE-CONTRADICTION — [RESOLVED]: 사용자 구현 요청/범위 선택과 Formal post-audit Implementation Kickoff Approval을 분리했고, 후자는 일관되게 pending이다 (`spec.md:L24-L28`; `plan.md:L3-L5`, `L277`; `progress.md:L3-L5`).
- D3 AC002-VACUOUS-NULL-GUARD — [RESOLVED]: JSON/SQLite/active-WAL 3개 테스트가 raw JSON key-present literal null을 요구하며, 현재 key absence로 각각 RED다 (`plan.md:L101-L127`; `todo_identity_red_test.go:L191-L255`). main/WAL bytes 및 DB/schema 보존 premise도 실행됐다.
- D4 AC003-WRONG-REASON-SELECTOR — [RESOLVED]: selector는 entropy와 backfill-trigger 2개 전용 테스트를 선택한다. 두 테스트는 reached=0, nil writer error, partial record change로 RED이며 기존 Restore test 오선택은 없다 (`plan.md:L129-L148`; `todo_identity_red_test.go:L261-L321`). 재시도/stamp 계측 누락은 D11로 분리한다.
- D5 AC006-CONCURRENCY-MODEL-UNDEFINED — [RESOLVED]: 동일 process의 별도 BacklogStore/SQLite handle 2개와 barrier가 최소 경계로 고정됐고, race는 보조 검사로 분리됐다 (`acceptance.md:L13`; `plan.md:L59-L68`; `todo_identity_red_test.go:L452-L503`).
- D6 MIGRATION-ATOMICITY-SUBJECT-UNDEFINED — [RESOLVED]: exact side-table DDL, SQLite transaction subject, JSON staging/publication, main/WAL/JSON hash set, SHM 제외, logical rollback과 physical-byte 차이가 정의됐다 (`plan.md:L11-L37`). backup-file restore는 AC-003 밖 비차단 후속으로 명시됐다 (`spec.md:L72-L74`; `acceptance.md:L25`).
- D7 AC007-EXCLUSION-BOUNDARY-AMBIGUOUS — [RESOLVED]: scope가 기존 `AddedAt`, `State`, `OwnerLabel`, provenance, `ReportedState`, `LoadPure`, `RecordFactoryCardState`의 음성 회귀로 한정됐고 receipt/approval/owner-token subsystem을 만들지 않는다고 분리했다 (`acceptance.md:L15`; `plan.md:L252`). OwnerLabel/provenance assertion 누락은 D13으로 분리한다.
- D8 REQ003-HOW-IN-REQUIREMENT — [RESOLVED]: REQ-003은 “하나의 transaction” 대신 부분 변경이 관측되지 않는 원자적 outcome으로 바뀌었다 (`spec.md:L36`). transaction은 plan에만 있다 (`plan.md:L25`).
- D9 MX-VERIFICATION-NOT-CONCRETE — [RESOLVED]: 5개 file/function anchor, exact marker, expected count=5, bounded `rg` command가 추가됐다 (`plan.md:L255-L273`).

## Defects Found (structured defect-list)

D10. AC001-CROSS-PROJECT-UNIQUENESS-UNMEASURED — `acceptance.md:L3`; `todo_identity_red_test.go:L147-L189` — AC-001은 독립 Git/비Git 프로젝트의 runtime start/assignment와 project/card 충돌 부재를 요구한다. 현재 테스트는 각 subtest의 project UUID를 외부에 보존·상호 비교하지 않고, card UUID도 프로젝트 간 비교하지 않으며, `RecordFactoryRunStart/Assignment`를 호출하지 않는다. 동일 project UUID가 두 프로젝트에 발급되거나 프로젝트 사이 card UUID가 충돌해도 future GREEN이 가능하다. — Severity: major — Class: blocking — Required fix: 두 fixture의 project UUID와 모든 card UUID를 공통 collection에 저장해 전체 cardinality/uniqueness를 assert하고, AC 문구에 남긴 runtime start/assignment를 실제 호출·검증하거나 중복인 그 When clause를 AC-005로만 이동한다.

D11. AC003-RETRY-AND-STAMP-UNMEASURED — `acceptance.md:L7`; `plan.md:L57`, `L279`; `todo_identity_red_test.go:L261-L321` — AC-003은 오류 제거 후 성공 재시도·반복 writer identity 보존과 schema stamp 부분 반영 0을 요구하지만 두 테스트 모두 fault 제거/재시도를 하지 않고 stamp/table creation rollback을 assert하지 않는다. fault handling만 구현하면 이 selector가 GREEN이면서 acceptance 후반이 깨질 수 있다. — Severity: major — Class: blocking — Required fix: entropy/trigger fault를 제거한 뒤 같은 fixture에서 재시도하고 non-null UUID 발급 및 반복 writer 동일 UUID를 assert한다. DDL 없는 legacy fixture의 table/stamp failure 지점을 추가하고 failure 후 table/version/record/row snapshot이 전부 원상태인지 검사한다.

D12. AC005-AMBIGUOUS-STATE-AUTHORITY-UNMEASURED — `acceptance.md:L11`; `plan.md:L29`, `L279`; `todo_identity_red_test.go:L371-L450` — AC-005는 missing/ambiguous 거절, `RecordFactoryCardState`, `reported_state` 비권위를 요구하지만 테스트는 live/archive assignment link와 missing card만 측정한다. ambiguous fixture 및 state API 호출/카드 State 불변 assertion이 없다. — Severity: major — Class: blocking — Required fix: 동일 local ID가 live/archive 또는 project 후보에서 모호해지는 정확 fixture를 만들고 error+mutation0을 assert하며, `RecordFactoryCardState(..., "completed", ...)` 후 `ReportedState`만 바뀌고 카드 `State`는 승격되지 않는지 검사한다.

D13. AC007-OWNERLABEL-PROVENANCE-UNMEASURED — `acceptance.md:L15`; `acceptance.md:L19`; `plan.md:L252`, `L279`; `todo_identity_red_test.go:L572-L615` — AC-007은 `OwnerLabel`과 provenance를 UUID timestamp로 덮어쓰지 않는다고 명시하지만 current test와 실행 로그는 `AddedAt`, `State`, `ReportedState` 3개만 검사한다. 문서가 이 GAP를 직접 인정한다. — Severity: major — Class: blocking — Required fix: fixture에 실제 `OwnerLabel`과 runtime provenance 값을 넣고 역순 UUIDv7 삽입 전후 deep equality를 assert하며 실행 로그의 `authority_fields` count에도 포함한다. 실제 타입에 해당 필드가 없다면 AC에서 존재한다고 가정하지 말고 현재 관측 가능한 필드로 줄인다.

## Recommendation

1. D10–D13의 누락 assertion을 기존 7 selector 안에 추가하고, 각 selector가 같은 tree에서 여전히 올바른 이유로 RED인지 재측정한다.
2. `red-baseline.md`의 test SHA, AC test count, stdout, exit를 갱신한다. untracked test이므로 Git tree SHA만으로는 pin되지 않는다.
3. iteration 3은 D10–D13 delta와 D1–D9 회귀만 확인한다. Formal post-audit Implementation Kickoff Approval은 PASS 뒤에도 별도 gate로 남는다.

## Evidence Record

### Claim

Iteration 1의 D1–D9 표면 수정은 모두 확인됐고 MP-8 RED re-execution도 통과했다. 그러나 4개 release-blocking AC의 verification instrument가 선언한 criterion 전체를 측정하지 않아 Tier M 구현 kickoff에 필요한 testability/completeness가 부족하다.

### Evidence

#### E1 — baseline and hashes

Commands:

```bash
git rev-parse --show-toplevel
git rev-parse HEAD
git branch --show-current
git status --short -- .moai/specs/SPEC-TODO-IDENTITY-001 .moai/reports/SPEC-TODO-IDENTITY-001 internal/kanban
shasum -a 256 .moai/specs/SPEC-TODO-IDENTITY-001/spec.md .moai/specs/SPEC-TODO-IDENTITY-001/plan.md .moai/specs/SPEC-TODO-IDENTITY-001/acceptance.md .moai/reports/SPEC-TODO-IDENTITY-001/red-baseline.md .moai/reports/SPEC-TODO-IDENTITY-001/plan-audit-iter1.md internal/kanban/todo_identity_red_test.go
```

Observed stdout:

```text
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/todo-unified
a315dad9af0d3a0e04862e6106b3993d9a3812f7
WT-todo-unified
 M internal/kanban/backlog_integrity_audit_test.go
 M internal/kanban/backlog_migrate.go
 M internal/kanban/backlog_pure_reader_test.go
 M internal/kanban/backlog_sqlite.go
 M internal/kanban/backlog_store.go
 M internal/kanban/factory_runtime.go
 M internal/kanban/factory_runtime_test.go
?? .moai/reports/SPEC-TODO-IDENTITY-001/
?? .moai/specs/SPEC-TODO-IDENTITY-001/
?? internal/kanban/todo_identity_red_test.go
?? internal/kanban/todo_runtime.go
?? internal/kanban/todo_runtime_safety_test.go
?? internal/kanban/todo_runtime_store_test.go
?? internal/kanban/todo_unified_red_test.go
19985e8f4eec6a0c7730994f67a947000bc3c20b1f33f0f7ad15f8bbe5d2e2e1  .moai/specs/SPEC-TODO-IDENTITY-001/spec.md
a83e36a11a02a008c6225585c9e6dfea222683cc14b24a16140a6a40d6bbf9cd  .moai/specs/SPEC-TODO-IDENTITY-001/plan.md
b934cd0056dbc30ecdfd1ffe41928dd7f0e7ba9686901259ef1fc1d9a380eee0  .moai/specs/SPEC-TODO-IDENTITY-001/acceptance.md
f1c3d7a879c435d9b74338f0cbc44574f3ae9ee748efdd9cf48f3db3cced91e3  .moai/reports/SPEC-TODO-IDENTITY-001/red-baseline.md
804360c55a928dcd75c7082f3ef2c7aad3882b57746cda462c2fc0fce293adc6  .moai/reports/SPEC-TODO-IDENTITY-001/plan-audit-iter1.md
d7e13ab73ada9b1aaba97cd2ba3412f50ab20acc55004a32ea5ac39317d8dfd7  internal/kanban/todo_identity_red_test.go
```

Exit code: 0.

#### E2 — lint, structure, clarification, D7, D8

Commands and observed output:

```text
$ moai spec lint .moai/specs/SPEC-TODO-IDENTITY-001 --json
[]
[exit 0]

$ rg -n '\[NEEDS CLARIFICATION' .moai/specs/SPEC-TODO-IDENTITY-001/plan.md
[no stdout]
[exit 1: zero matches]

$ D7 reference/status walk
SPEC-TODO-IDENTITY-001 status=draft
SPEC-TODO-RUNTIME-STORE-001 status=draft
SPEC-TODO-UNIFIED-001 status=draft
[exit 0]

$ rg -q 'syscall' .moai/specs/SPEC-TODO-IDENTITY-001/spec.md
[no stdout]
[exit 1: literal absent]
```

#### E3 — selector inventory

Command:

```bash
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -list '^TestTodoIdentityAC' -count=1
```

Observed stdout:

```text
TestTodoIdentityAC001GitAndNonGitUUIDv7
TestTodoIdentityAC002LegacyJSONExplicitNullNoMutation
TestTodoIdentityAC002LegacySQLiteExplicitNullNoMutation
TestTodoIdentityAC002LegacyActiveWALExplicitNullNoMutation
TestTodoIdentityAC003UUIDEntropyFailureRollsBack
TestTodoIdentityAC003BackfillFaultRollsBack
TestTodoIdentityAC004LifecycleStability
TestTodoIdentityAC005RuntimeLinksLiveAndArchived
TestTodoIdentityAC005MissingRuntimeCardMutationZero
TestTodoIdentityAC006ConcurrentDistinctHandles
TestTodoIdentityAC006FutureIdentityVersionFailClosed
TestTodoIdentityAC006RetiredRegression
TestTodoIdentityAC007TimestampNonAuthority
ok   github.com/modu-ai/moai-adk/internal/kanban 0.306s
```

Exit code: 0. Top-level tests listed: 13.

#### E4 — AC-001

Command:

```bash
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -run '^TestTodoIdentityAC001' -count=1 -v -timeout=120s
```

Observed stdout/stderr:

```text
=== RUN   TestTodoIdentityAC001GitAndNonGitUUIDv7
=== RUN   TestTodoIdentityAC001GitAndNonGitUUIDv7/git
    todo_identity_red_test.go:178: git project: missing project_uuid after successful writer
    todo_identity_red_test.go:179: git first Add: missing card_uuid after successful writer
    todo_identity_red_test.go:180: git second Add: missing card_uuid after successful writer
    todo_identity_red_test.go:181: git first LoadPure: missing card_uuid after successful writer
    todo_identity_red_test.go:182: git second LoadPure: missing card_uuid after successful writer
=== RUN   TestTodoIdentityAC001GitAndNonGitUUIDv7/non-git
    todo_identity_red_test.go:178: non-git project: missing project_uuid after successful writer
    todo_identity_red_test.go:179: non-git first Add: missing card_uuid after successful writer
    todo_identity_red_test.go:180: non-git second Add: missing card_uuid after successful writer
    todo_identity_red_test.go:181: non-git first LoadPure: missing card_uuid after successful writer
    todo_identity_red_test.go:182: non-git second LoadPure: missing card_uuid after successful writer
=== NAME  TestTodoIdentityAC001GitAndNonGitUUIDv7
    todo_identity_red_test.go:188: AC-TID-001 executed fixtures=2 git=1 non_git=1 cards=4
--- FAIL: TestTodoIdentityAC001GitAndNonGitUUIDv7 (0.57s)
    --- FAIL: TestTodoIdentityAC001GitAndNonGitUUIDv7/git (0.42s)
    --- FAIL: TestTodoIdentityAC001GitAndNonGitUUIDv7/non-git (0.14s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/kanban 1.134s
FAIL
```

Exit code: 1. Top-level tests: 1; subtests: 2. RED reproduced. Source inspection established D10.

#### E5 — AC-002

Command:

```bash
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -run '^TestTodoIdentityAC002' -count=1 -v -timeout=120s
```

Observed stdout/stderr:

```text
=== RUN   TestTodoIdentityAC002LegacyJSONExplicitNullNoMutation
    todo_identity_red_test.go:201: JSON project: project_uuid key absent, want key-present literal null
    todo_identity_red_test.go:202: JSON live: card_uuid key absent, want key-present literal null
    todo_identity_red_test.go:203: JSON archive: card_uuid key absent, want key-present literal null
    todo_identity_red_test.go:210: AC-TID-002 executed JSON=1 live=1 archive=1
--- FAIL: TestTodoIdentityAC002LegacyJSONExplicitNullNoMutation (0.00s)
=== RUN   TestTodoIdentityAC002LegacySQLiteExplicitNullNoMutation
    todo_identity_red_test.go:221: SQLite project: project_uuid key absent, want key-present literal null
    todo_identity_red_test.go:222: SQLite live: card_uuid key absent, want key-present literal null
    todo_identity_red_test.go:223: SQLite archive: card_uuid key absent, want key-present literal null
    todo_identity_red_test.go:230: AC-TID-002 executed SQLite=1 live=1 archive=1
--- FAIL: TestTodoIdentityAC002LegacySQLiteExplicitNullNoMutation (0.15s)
=== RUN   TestTodoIdentityAC002LegacyActiveWALExplicitNullNoMutation
    todo_identity_red_test.go:248: WAL project: project_uuid key absent, want key-present literal null
    todo_identity_red_test.go:249: WAL live: card_uuid key absent, want key-present literal null
    todo_identity_red_test.go:250: WAL archive: card_uuid key absent, want key-present literal null
    todo_identity_red_test.go:254: AC-TID-002 executed active_WAL=1 committed_snapshot=1 live=1 archive=1
--- FAIL: TestTodoIdentityAC002LegacyActiveWALExplicitNullNoMutation (0.14s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/kanban 1.406s
FAIL
```

Exit code: 1. Top-level tests: 3. Correct key-absence RED reproduced.

#### E6 — AC-003

Command:

```bash
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -run '^TestTodoIdentityAC003' -count=1 -v -timeout=120s
```

Observed stdout/stderr:

```text
=== RUN   TestTodoIdentityAC003UUIDEntropyFailureRollsBack
    todo_identity_red_test.go:277: AC-TID-003 entropy fault reached=0 writer_error=<nil>
    todo_identity_red_test.go:279: UUIDv7 entropy seam reached=0 want=1
    todo_identity_red_test.go:282: UUIDv7 entropy failure ignored
    todo_identity_red_test.go:285: entropy failure left partial change
--- FAIL: TestTodoIdentityAC003UUIDEntropyFailureRollsBack (0.15s)
=== RUN   TestTodoIdentityAC003BackfillFaultRollsBack
    todo_identity_red_test.go:314: AC-TID-003 backfill fault reached=0 writer_error=<nil> identity_rows=0
    todo_identity_red_test.go:316: writer missed backfill fault: reached=0 err=<nil>
    todo_identity_red_test.go:319: rollback failed: record_equal=false identity_rows=0
--- FAIL: TestTodoIdentityAC003BackfillFaultRollsBack (0.15s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/kanban 1.149s
FAIL
```

Exit code: 1. Top-level tests: 2. Required identity paths are absent, so fault reached=0 and partial record mutation produce RED. Source inspection established D11.

#### E7 — AC-004

Command:

```bash
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -run '^TestTodoIdentityAC004' -count=1 -v -timeout=120s
```

Observed stdout/stderr:

```text
=== RUN   TestTodoIdentityAC004LifecycleStability
    todo_identity_red_test.go:334: initial project: missing project_uuid after successful writer
    todo_identity_red_test.go:335: Add return: missing card_uuid after successful writer
    todo_identity_red_test.go:352: archive: missing card_uuid after successful writer
    todo_identity_red_test.go:360: restore: missing card_uuid after successful writer
    todo_identity_red_test.go:361: reopen project: missing project_uuid after successful writer
    todo_identity_red_test.go:368: AC-TID-004 executed mutate=1 archive=1 restore=1 reopen=1
--- FAIL: TestTodoIdentityAC004LifecycleStability (0.16s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/kanban 0.472s
FAIL
```

Exit code: 1. Top-level tests: 1. Correct identity-absence RED reproduced.

#### E8 — AC-005

Command:

```bash
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -run '^TestTodoIdentityAC005' -count=1 -v -timeout=120s
```

Observed stdout/stderr:

```text
=== RUN   TestTodoIdentityAC005RuntimeLinksLiveAndArchived
    todo_identity_red_test.go:398: project: missing project_uuid after successful writer
    todo_identity_red_test.go:399: live: missing card_uuid after successful writer
    todo_identity_red_test.go:400: archive: missing card_uuid after successful writer
    todo_identity_red_test.go:412: run: missing project_uuid after successful writer
    todo_identity_red_test.go:419: assignment t1: missing project_uuid after successful writer
    todo_identity_red_test.go:420: assignment t1: missing card_uuid after successful writer
    todo_identity_red_test.go:419: assignment t2: missing project_uuid after successful writer
    todo_identity_red_test.go:420: assignment t2: missing card_uuid after successful writer
    todo_identity_red_test.go:428: AC-TID-005 executed runs=1 assignments=2 live=1 archive=1
--- FAIL: TestTodoIdentityAC005RuntimeLinksLiveAndArchived (0.59s)
=== RUN   TestTodoIdentityAC005MissingRuntimeCardMutationZero
    todo_identity_red_test.go:449: AC-TID-005 regression guard executed missing=1 mutation_zero=1
--- PASS: TestTodoIdentityAC005MissingRuntimeCardMutationZero (0.30s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/kanban 1.845s
FAIL
```

Exit code: 1. Top-level tests: 2. Runtime UUID-link RED and missing-card regression PASS reproduced. Source inspection established D12.

#### E9 — AC-006 and race auxiliary

Commands:

```bash
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -run '^TestTodoIdentityAC006' -count=1 -v -timeout=120s
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test -race ./internal/kanban -run '^TestTodoIdentityAC006ConcurrentDistinctHandles$' -count=1 -v -timeout=120s
```

Observed stdout/stderr, primary selector:

```text
=== RUN   TestTodoIdentityAC006ConcurrentDistinctHandles
    todo_identity_red_test.go:483: concurrent Add: missing card_uuid after successful writer
    todo_identity_red_test.go:483: concurrent Add: missing card_uuid after successful writer
    todo_identity_red_test.go:495: concurrent project: missing project_uuid after successful writer
    todo_identity_red_test.go:497: stored t1: missing card_uuid after successful writer
    todo_identity_red_test.go:497: stored t2: missing card_uuid after successful writer
    todo_identity_red_test.go:500: conservation success=2 items=2 last_seq=2 project="" uuids=0
    todo_identity_red_test.go:502: AC-TID-006 executed handles=2 barrier=1 successes=2 cards=2 UUIDs=0
--- FAIL: TestTodoIdentityAC006ConcurrentDistinctHandles (0.21s)
=== RUN   TestTodoIdentityAC006FutureIdentityVersionFailClosed
=== RUN   TestTodoIdentityAC006FutureIdentityVersionFailClosed/LoadPure
    todo_identity_red_test.go:513: LoadPure accepted future identity version: &{Version:1 LastSeq:1 Items:[{ID:t1 Text:legacy AddedAt:2026-09-12T08:24:34Z SpecID:<nil> State:queued Landing:<nil>}] Findings:[] Archived:[] Runtime:{Runs:[] Assignments:[]}}
=== RUN   TestTodoIdentityAC006FutureIdentityVersionFailClosed/writer
    todo_identity_red_test.go:540: writer accepted/mutated future identity version err=<nil> unchanged=false before=1/1/0 after=2/2/0
=== NAME  TestTodoIdentityAC006FutureIdentityVersionFailClosed
    todo_identity_red_test.go:544: AC-TID-006 executed future_version operations=2
--- FAIL: TestTodoIdentityAC006FutureIdentityVersionFailClosed (0.35s)
=== RUN   TestTodoIdentityAC006RetiredRegression
    todo_identity_red_test.go:569: AC-TID-006 retired regression executed writers=2 mutation_zero=1
--- PASS: TestTodoIdentityAC006RetiredRegression (0.35s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/kanban 2.161s
FAIL
```

Primary exit code: 1. Top-level tests: 3. Semantic RED 2, regression PASS 1.

Observed stdout/stderr, race auxiliary:

```text
=== RUN   TestTodoIdentityAC006ConcurrentDistinctHandles
    todo_identity_red_test.go:483: concurrent Add: missing card_uuid after successful writer
    todo_identity_red_test.go:483: concurrent Add: missing card_uuid after successful writer
    todo_identity_red_test.go:495: concurrent project: missing project_uuid after successful writer
    todo_identity_red_test.go:497: stored t1: missing card_uuid after successful writer
    todo_identity_red_test.go:497: stored t2: missing card_uuid after successful writer
    todo_identity_red_test.go:500: conservation success=2 items=2 last_seq=2 project="" uuids=0
    todo_identity_red_test.go:502: AC-TID-006 executed handles=2 barrier=1 successes=2 cards=2 UUIDs=0
--- FAIL: TestTodoIdentityAC006ConcurrentDistinctHandles (0.28s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/kanban 1.913s
FAIL
```

Race auxiliary exit code: 1. Top-level tests: 1. No `WARNING: DATA RACE`; semantic UUID assertion failed, so this is not a race PASS.

#### E10 — AC-007

Command:

```bash
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -run '^TestTodoIdentityAC007' -count=1 -v -timeout=120s
```

Observed stdout/stderr:

```text
=== RUN   TestTodoIdentityAC007TimestampNonAuthority
    todo_identity_red_test.go:606: later UUID card: missing card_uuid after successful writer
    todo_identity_red_test.go:607: earlier UUID card: missing card_uuid after successful writer
    todo_identity_red_test.go:609: projection got="","" want="01890f3e-8b01-7000-8000-000000000002","01890f3e-8b00-7000-8000-000000000001"
    todo_identity_red_test.go:614: AC-TID-007 executed UUIDv7_timestamps=2 authority_fields=added_at,state,reported_state
--- FAIL: TestTodoIdentityAC007TimestampNonAuthority (0.26s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/kanban 0.610s
FAIL
```

Exit code: 1. Top-level tests: 1. Projection RED reproduced. Source/output establish D13.

#### E11 — existing core/runtime/retired regression selector

Command:

```bash
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -run '^(TestTodoRuntimeStoreFutureSchemaPreservesBytes|TestTodoRuntimeStoreWriterRejectsFutureTodoVersions|TestTodoRuntimeSafetyRetiredAndMissingCardRefuse)$' -count=1 -v -timeout=120s
```

Observed stdout/stderr:

```text
=== RUN   TestTodoRuntimeSafetyRetiredAndMissingCardRefuse
--- PASS: TestTodoRuntimeSafetyRetiredAndMissingCardRefuse (0.24s)
=== RUN   TestTodoRuntimeStoreFutureSchemaPreservesBytes
--- PASS: TestTodoRuntimeStoreFutureSchemaPreservesBytes (0.01s)
=== RUN   TestTodoRuntimeStoreWriterRejectsFutureTodoVersions
=== RUN   TestTodoRuntimeStoreWriterRejectsFutureTodoVersions/schema_version
=== RUN   TestTodoRuntimeStoreWriterRejectsFutureTodoVersions/runtime_schema_version
--- PASS: TestTodoRuntimeStoreWriterRejectsFutureTodoVersions (0.91s)
    --- PASS: TestTodoRuntimeStoreWriterRejectsFutureTodoVersions/schema_version (0.49s)
    --- PASS: TestTodoRuntimeStoreWriterRejectsFutureTodoVersions/runtime_schema_version (0.42s)
PASS
ok   github.com/modu-ai/moai-adk/internal/kanban 1.801s
```

Exit code: 0. Top-level tests: 3. This is separated regression evidence, not identity GREEN.

### Baseline-attribution

- Worktree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/todo-unified`
- Branch: `WT-todo-unified`
- Git HEAD/tree pin: `a315dad9af0d3a0e04862e6106b3993d9a3812f7`
- Current artifact SHA-256: spec `19985e8f…`, plan `a83e36a1…`, acceptance `b934cd00…`, RED baseline `f1c3d7a8…`, iter1 report `804360c5…`.
- Current untracked RED test SHA-256: `d7e13ab73ada9b1aaba97cd2ba3412f50ab20acc55004a32ea5ac39317d8dfd7`.
- The SPEC/test/report files are untracked. The tree SHA alone does not pin them; their SHA-256 values above are part of this baseline.
- Existing dirty production/runtime files were observed before this audit and were not modified by the auditor. Their correctness and provenance are outside this plan-delta verdict.

### Gaps

- 작성자 reasoning, 대화, 초안은 읽지 않았다.
- Full package suite, repository suite, coverage, production implementation correctness, DB migration execution, backup-file restore, cross-process contention은 실행하지 않았다. 이들은 plan-delta RED audit의 범위가 아니다.
- AC-003의 fault가 현재 reached=0이므로 실제 rollback 경로가 작동한다는 증거는 아니다. 이는 RED reason이며 GREEN claim이 아니다.
- AC-006 race command는 semantic assertion에서 실패했으므로 race-clean PASS가 아니다.
- Formal post-audit Implementation Kickoff Approval은 artifact상 pending이며 이 FAIL 판정으로 발급되지 않는다.

### Residual-risk

- D10–D13을 고치지 않으면 future GREEN이 criterion 일부를 측정하지 않은 채 release-blocking AC PASS로 오인될 수 있다.
- `uuid.SetRand`는 package-global seam이다. current AC-003 selector에는 parallel test가 없고 defer 복구가 있지만, 향후 parallel identity test 추가 시 process isolation이 필요할 수 있다.
- active WAL hash는 main/WAL만 포함하고 SHM을 제외한다. 이 범위는 plan에 명시됐지만, GREEN 단계에서 logical row/schema assertions도 함께 유지해야 한다.
- 별도 BacklogStore handle 2개는 같은 process 경계만 검증한다. cross-process correctness는 이 child의 완료 주장에 포함되지 않는다.

## Operational Notes (unverified)

- `measured` — iteration 3 drift 확인: E1의 SHA-256 5개와 test SHA를 다시 측정해 이 보고서 baseline과 비교한다.
- `inferred` — AC-001의 runtime start/assignment가 AC-005와 중복이라면 criterion 단순화를 검토한다. 판단 규칙: 한 AC의 When에 남긴 호출은 그 AC selector가 실행하거나 그 clause를 다른 AC로 이동해야 한다.
