# SPEC Review Report: SPEC-TODO-IDENTITY-001

Iteration: 1/3
Verdict: FAIL
Overall Score: 0.69
Tier: M — PASS threshold 0.80

Reasoning context ignored per M1 Context Isolation. 이 감사는 `spec.md`, `plan.md`, `acceptance.md`의 최종 내용과 그 문서가 지목한 현재 트리 검증 명령만을 판정 입력으로 사용했다.

## Adversarial failure-mode inventory

감사 전에 다음 실패 형태를 기본 가설로 두고 전수 확인했다.

- REQ 번호 누락·중복·zero-padding 불일치
- REQ의 비-GEARS 문장 또는 verification-layer GWT와 requirement-layer GEARS 혼동
- canonical 12-field frontmatter 누락·타입 오류·금지 alias
- REQ 내부 HOW/구체 구현 고정
- 7 REQ와 7 AC 간 orphan·미커버·잘못된 참조
- 모호하거나 빈 selector, zero-test green, 잘못된 이유의 RED
- release-blocking AC의 RED-now command/stdout/exit/tree-SHA 누락
- legacy JSON·SQLite·active WAL 사이 migration/rollback 원자성 경계 누락
- in-process race와 cross-process writer 경쟁의 혼동
- public JSON null/key-presence 및 Add-return/runtime-link 검증의 공백
- 포함 범위와 exclusions의 충돌
- MX 계획의 비측정 가능성
- cross-SPEC retired/superseded/archived 참조 미조정
- `syscall` 도입 시 build constraint 또는 예외 누락
- 미해결 `[NEEDS CLARIFICATION]` 및 Implementation Kickoff Approval 상태 모순

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: requirement layer에는 `REQ-TID-001`부터 `REQ-TID-007`까지 순서대로 한 번씩 있다 (`spec.md:L32`, `L34`, `L36`, `L38`, `L40`, `L42`, `L44`). 구조 계수는 `REQ_COUNT=7`이었다.
- [PASS] MP-2 EARS/GEARS format compliance: 이 판정은 **requirement layer만** 대상으로 했다. REQ-001/003/004/005/006은 `When …, the system shall …`, REQ-002는 `While …, the system shall …`, REQ-007은 `The system shall …` 패턴이다 (`spec.md:L32-L44`). `AC-TID-*`의 Given-When-Then은 verification layer 형식으로 별도 평가했으며 MP-2에 불리하게 계산하지 않았다.
- [PASS] MP-3 YAML frontmatter validity: canonical 12 fields가 올바른 타입과 값으로 존재한다 (`spec.md:L2-L13`), optional `tier: M`도 유효하다 (`spec.md:L14`). 현재 명령 `moai spec lint .moai/specs/SPEC-TODO-IDENTITY-001 --json`은 exit 0, stdout `[]`을 반환했다. `related_specs`는 canonical dependency key가 아니지만 required-field/type 실패로 판정하지 않았다.
- [N/A] MP-4 Section 22 language neutrality: `module: "internal/kanban"` (`spec.md:L11`) 및 Go package/test 명령 (`plan.md:L25`, `L46-L49`)으로 한정된 single-language SPEC이다. 16개 언어를 열거해야 하는 template-bound universal tooling 범위가 아니다.
- [PASS] MP-5 D7 cross-SPEC reconciliation: verification verb는 실행 가능했다. `SPEC-TODO-IDENTITY-001`, `SPEC-TODO-RUNTIME-STORE-001`, `SPEC-TODO-UNIFIED-001`가 모두 존재하고 각각 `status=draft`였다. retired/superseded/archived 참조가 없어 D7 BLOCKING finding은 없다 (`spec.md:L15`, `L18`).
- [PASS] MP-6 D8 cross-platform discipline: `rg -q 'syscall' spec.md`의 관측 결과는 literal `syscall` 부재였고 wrapper 출력은 `PASS: syscall absent`였다. 따라서 D8 BLOCKING finding은 없다.
- [PASS] MP-7 clarification gate: Tier M의 필수 입력인 `plan.md`에서 `rg -n '\[NEEDS CLARIFICATION' ...`은 match 0, `RG_EXIT=1`이었다. Tier M에 `research.md`는 요구되지 않는다. 다만 marker 부재가 Implementation Kickoff Approval 통과를 뜻하지는 않는다.
- [FAIL] MP-8 RED-now cell re-execution: `acceptance.md:L21`은 “모든7AC가 종료필수”라고 분류하지만, AC별 command·verbatim stdout·exit code·tree SHA의 4요소를 갖춘 RED-now cell이 없다. `plan.md:L22-L37`의 historical ledger는 한 명령만 기록하며 작성 시 재실행 값이 아니라고 명시하고, AC-001/004/005 일부에만 대응하며 AC-002/003/006/007 증거가 없다고 스스로 제한한다. 현재 tree `a315dad9af0d3a0e04862e6106b3993d9a3812f7`에서 그 명령은 exit 1로 UUID 부재 RED를 재현했지만, AC-002 명령은 absence와 null을 함께 허용한 채 exit 0, AC-003 selector는 `Restore` 기존 테스트 하나만 잘못 선택해 UUID 부재로 exit 1, AC-006과 AC-007 selector는 각각 0 tests로 exit 0이었다. release-blocking AC 전체의 올바른 RED가 재현되지 않았으므로 score와 무관하게 FAIL이다.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|---|---:|---|---|
| Clarity | 0.50 | 0.50 | migration의 persistence 형태를 side table 또는 row column 중 구현 시점에 선택하도록 남겼고 (`plan.md:L14`), JSON/SQLite/live/archive/runtime을 “같은 SQL transaction”으로 다루는 원자성 경계가 명시되지 않았다 (`plan.md:L15`, `L55`). AC-006의 “두 실제 writer”가 goroutine, DB connection, process 중 무엇인지도 고정되지 않았다 (`acceptance.md:L13`). 합리적인 구현자가 서로 다른 구현·검증을 만들 수 있다. |
| Completeness | 0.75 | 0.75 | HISTORY (`spec.md:L20`), 목적/맥락 (`L26-L28`), requirements (`L30-L44`), acceptance (`L46-L60`), 구체 Out of Scope H3와 bullet (`L62-L69`), Tier M 3-artifact set은 존재한다. 그러나 모든 AC를 종료필수로 선언하고도 RED-now 4요소와 일부 핵심 검증 경계가 비어 있다 (`acceptance.md:L21`, `plan.md:L37`, `L41-L49`). |
| Testability | 0.50 | 0.50 | 7개 AC는 형식상 Given-When-Then (`acceptance.md:L3-L15`)이지만 AC-002 selector는 key-absent를 허용한 채 PASS했고, AC-003 selector는 Failure/Backfill이 아닌 Restore 하나를 선택했으며, AC-006/007은 zero-test green이었다. AC-006의 writer 격리 단위와 AC-007의 out-of-scope 권위 필드 관측 API도 불명확하다. |
| Traceability | 1.00 | 1.00 | `AC-TID-001`부터 `AC-TID-007`까지 각각 같은 번호의 존재하는 REQ를 정확히 한 번 참조한다 (`acceptance.md:L3-L15`); uncovered REQ와 orphan AC는 없다. 구조 계수는 REQ 7, inline AC 7, sibling acceptance AC 7이고 inline/sibling AC 본문은 byte comparison상 동일했다. |

산술 평균: `(0.50 + 0.75 + 0.50 + 1.00) / 4 = 0.6875`, 두 자리 반올림 `0.69`. Tier M threshold `0.80` 미달이며 MP-8 firewall 실패가 독립적으로 FAIL을 강제한다.

## Checklist Results

### Group 1 — YAML Frontmatter

- FC-ALL: PASS — 12 required fields는 `spec.md:L2-L13`, optional tier는 `L14`; lint exit 0/`[]`.

### Group 2 — Document Structure

- SC-1 HISTORY: PASS — `spec.md:L20-L24`.
- SC-2 WHY/Context: PASS — `spec.md:L26-L28`.
- SC-3 WHAT/Scope: PASS — identity 최소 단위 및 umbrella 비완료 경계가 `spec.md:L28`에 있다.
- SC-4 Requirements: PASS — `spec.md:L30-L44`, 7 entries.
- SC-5 Acceptance Criteria: PASS — `spec.md:L46-L60` 및 `acceptance.md:L3-L15`, 7 entries.
- SC-6 Out of Scope: PASS — 구체 H3와 4개 bullet이 `spec.md:L64-L69`에 있다.

### Group 3 — Requirements Quality

- RQ-1/RQ-2 numbering: PASS — 001…007, gap/duplicate 없음 (`spec.md:L32-L44`).
- RQ-3 outcome orientation: PARTIAL — 대체로 외부 동작이지만 REQ-003은 “하나의 transaction”이라는 HOW를 규정한다 (`spec.md:L36`).
- RQ-4 no implementation details: FAIL — REQ-003의 transaction은 원자적 결과가 아니라 구현 기법을 고정한다 (`spec.md:L36`).
- RQ-5 precise normative language: PASS — normative `shall`을 사용하고 should/may/weasel word가 없다 (`spec.md:L32-L44`).
- RQ-6 GEARS: PASS — MP-2와 동일한 requirement-layer 전수 판정.

### Group 4 — Acceptance Criteria Quality

- AC-1 GWT form: PASS — 7/7 Given-When-Then (`acceptance.md:L3-L15`).
- AC-2 binary testability: FAIL — AC-002/003/006/007의 현재 selector와 경계가 판정 가능한 전체 기준을 실행하지 못한다.
- AC-3 weasel words: PASS — `appropriate`, `adequate`, `reasonable`, `good`, `proper` 없음 (`acceptance.md:L3-L15`).
- AC-4 valid REQ references: PASS — 7/7 valid (`acceptance.md:L3-L15`).
- AC-5 REQ coverage: PASS — 7/7 covered (`acceptance.md:L3-L15`).
- AC-6 RED-now: FAIL — MP-8과 동일.

### Group 5 — Language Neutrality

- LN-1/LN-2: N/A — single-language Go package SPEC.
- LN-3: PASS by N/A precedent — `spec.md:L11`, `plan.md:L25`.

### Group 6 — Consistency

- CN-1 requirements contradiction: PASS — REQ끼리 직접 모순은 확인되지 않았다 (`spec.md:L32-L44`).
- CN-2 exclusions conflict: FAIL — AC-007은 owner·evidence·approval/receipt 판단을 실행하도록 요구하지만 (`acceptance.md:L15`), 해당 owner token/generation과 completion receipt 기능은 이 child에서 구현하지 않는다고 제외한다 (`spec.md:L66`). 기존 API에 대한 음성 회귀인지 신규 기능 검증인지 분리되지 않았다.
- CN-3 priority/tags coherence: PASS — P1, todo/identity/uuidv7/sqlite가 실제 범위와 일치한다 (`spec.md:L9`, `L13`, `L28`).

### Group 7 — D7 Cross-SPEC Reconciliation

- D7-1…D7-4: PASS — 세 reference 모두 존재하고 `draft`; terminal-status reconciliation 불필요.
- D7-5: PASS — not-found reference 없음.

### Group 8 — D8 Cross-Platform Discipline

- D8-1…D8-4: PASS — `syscall` literal 없음.

## Defects Found (structured defect-list)

D1. MP8-REDNOW-MISSING — `acceptance.md:L21`; `plan.md:L22-L49` — 종료필수 7 AC에 criterion별 RED-now 4요소가 없고, 현재 ledger는 일부 AC의 historical UUID-absence만 제공한다. AC-006/007 계획 selector는 현재 zero-test green이다. — Severity: critical — Class: blocking — Required fix: AC-001…007 각각에 현재 tree SHA, 단일 read-only command, verbatim stdout, exit code, 올바른 실패 이유를 기록하고 같은 tree에서 재실행해 실제 test count가 1 이상임을 보인다. AC-001/004/005를 하나의 selector로 묶을 경우에도 criterion별 어떤 test가 어떤 premise를 RED로 만들었는지 분리한다.

D2. KICKOFF-STATE-CONTRADICTION — `plan.md:L5`, `plan.md:L64`, `spec.md:L28` — plan은 “kickoff 승인 수령”과 “정식게이트 미통과”를 동시에 기록한다. canonical gate는 plan-audit 뒤의 별도 score-independent Implementation Kickoff Approval이므로 현재 artifacts만으로는 통과 증거가 없다. — Severity: major — Class: blocking — Required fix: pre-audit 범위 선택/의향과 post-audit Implementation Kickoff Approval을 분리해 명명하고, 현재 상태를 `formal gate pending` 하나로 일관되게 기록한다. 이 감사 FAIL 상태에서는 run 진입을 승인된 것으로 취급하지 않는다.

D3. AC002-VACUOUS-NULL-GUARD — `acceptance.md:L5`, `plan.md:L37`, `plan.md:L46` — AC-002는 key-present explicit null을 요구하지만 현재 지정 selector의 유일한 테스트는 `raw="" (absence and null both observationally allowed)`를 출력하고 PASS한다. — Severity: major — Class: blocking — Required fix: raw JSON object에서 `project_uuid`/`card_uuid` key presence를 별도로 assert하고 각 raw token이 literal `null`인지 검사하는 failing input을 작성한다. legacy JSON·SQLite·active WAL 및 live/archive 각각의 swept count를 출력한다.

D4. AC003-WRONG-REASON-SELECTOR — `acceptance.md:L7`, `plan.md:L47` — `^TestTodoIdentity.*(Failure|Backfill|Restore)`는 현재 `TestTodoIdentityREDLifetimeAcrossArchiveRestore` 하나만 선택하며 UUID 미발급으로 실패한다. 생성기 실패, 각 backfill write failure, rollback, backup restore를 전혀 실행하지 않아 AC-003의 올바른 RED가 아니다. — Severity: major — Class: blocking — Required fix: Failure·Backfill·Restore의 의도된 test 이름과 최소 실행 수를 명시하고, fault injection 도달 count 및 변경 0/hash 복원을 각각 실패시키는 known input을 기록한다.

D5. AC006-CONCURRENCY-MODEL-UNDEFINED — `acceptance.md:L13`, `plan.md:L16`, `plan.md:L48`, `plan.md:L60` — “두 실제 writer”의 격리 단위가 불명확하고 `go test -race`는 cross-process SQLite lost-update/lock semantics를 입증하지 않는다. 현재 selector는 0 tests다. — Severity: major — Class: blocking — Required fix: 두 goroutine이 아니라 필요한 실제 경계(별도 process 또는 최소 별도 DB connection/handle)를 명시하고 barrier, 성공/거절 요청 수, final rows/UUID linkage, lock/cancel error propagation을 검증한다. race detector는 process-local data race 보조 검사로 분리한다.

D6. MIGRATION-ATOMICITY-SUBJECT-UNDEFINED — `plan.md:L14-L16`, `plan.md:L53-L55`, `acceptance.md:L5-L7` — persistence layout을 구현 시점으로 미뤘고, legacy JSON·SQLite·active WAL·live/archive/runtime 중 무엇이 한 SQL transaction의 원자적 subject인지, bytes/hash가 main DB만인지 WAL/SHM까지인지 정의하지 않았다. 서로 다른 매체를 단일 SQL transaction으로 rollback한다는 해석은 성립하지 않을 수 있다. — Severity: major — Class: blocking — Required fix: source→target migration 경계, transaction에 포함되는 DB/표, 외부 JSON 처리 순서, WAL checkpoint/sidecar 관측, backup manifest와 hash 대상 파일 집합, 실패 후 복원 순서를 명시한다. side table/row column 선택은 사용자 질문이 아니라 기존 schema 근거에 따른 구현자 결정이라면 그 결정 규칙과 invariant를 적는다.

D7. AC007-EXCLUSION-BOUNDARY-AMBIGUOUS — `acceptance.md:L15`, `spec.md:L60`, `spec.md:L66`, `plan.md:L18` — AC-007은 owner·evidence·approval/receipt 판단을 실행하라고 하지만 그 기능 일부는 out of scope다. 기존 동작에 대한 음성 회귀인지 후속 기능의 선행 검증인지 구분되지 않아 구현자가 서로 다른 surface를 시험할 수 있다. — Severity: major — Class: blocking — Required fix: 이 child에 이미 존재하는 정확한 reader/decision API와 필드를 열거해 UUID time bits만 변조하는 음성 회귀로 한정하거나, out-of-scope 축은 후속 SPEC으로 이관한다.

D8. REQ003-HOW-IN-REQUIREMENT — `spec.md:L36` — “하나의 transaction”은 원자적 외부 결과가 아니라 구현 기법이다. — Severity: minor — Class: optional — Required fix: REQ는 “부분 변경이 관측되지 않는 원자적 처리”로 쓰고 transaction/lock 선택은 plan에 둔다.

D9. MX-VERIFICATION-NOT-CONCRETE — `plan.md:L57-L60` — NOTE/WARN/REASON 의도는 있으나 대상 anchor, 기대 count, 검증 명령이 없다. — Severity: minor — Class: optional — Required fix: 구현 후 측정할 exact tag anchors와 bounded `rg`/MX lint 명령 및 예상 count를 verification table에 추가한다.

## Recommendation

1. D1을 먼저 수정해 7개 release-blocking AC 모두의 RED-now 4요소와 올바른 실패 이유를 확보한다. zero-test/부분 selector 상태에서는 iteration 2로 넘기지 않는다.
2. D2의 gate 상태를 `formal Implementation Kickoff Approval pending after PASS`로 단일화한다.
3. D3·D4·D5의 selector를 premise별 독립 테스트로 나누고 실제 swept count를 명시한다.
4. D6에서 JSON/SQLite/WAL migration과 rollback의 원자성·hash subject를 확정한다.
5. D7에서 AC-007을 현재 child의 관측 가능한 API로 한정한다.
6. manager-spec 수정 후 iteration 2는 D1-D7 defect delta와 회귀만 재감사한다. 현재 verdict로 구현에 착수하지 않는다.

## Evidence Record

### Claim

현재 plan은 구조·GEARS·1:1 traceability는 갖췄지만, release-blocking RED-now와 핵심 migration/concurrency 경계가 구현 착수 판정을 지탱하지 못한다.

### Evidence

#### E1 — tree baseline and artifact hashes

Command:

```bash
git rev-parse --show-toplevel
git rev-parse HEAD
git branch --show-current
git status --short -- .moai/specs/SPEC-TODO-IDENTITY-001 .moai/reports/plan-audit
shasum -a 256 .moai/specs/SPEC-TODO-IDENTITY-001/spec.md .moai/specs/SPEC-TODO-IDENTITY-001/plan.md .moai/specs/SPEC-TODO-IDENTITY-001/acceptance.md internal/kanban/todo_identity_red_test.go
```

Observed stdout:

```text
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/todo-unified
a315dad9af0d3a0e04862e6106b3993d9a3812f7
WT-todo-unified
?? .moai/specs/SPEC-TODO-IDENTITY-001/
0003eb813b8af02918c0720fabefce7da951719442a0be1d6f63486d94b040e5  .moai/specs/SPEC-TODO-IDENTITY-001/spec.md
bc9db12c4a2c90883ffa05c62c9a9e1b1a8129f1e2ca3ef4120555cd2f26725b  .moai/specs/SPEC-TODO-IDENTITY-001/plan.md
dc1fb66b599c30bb7051de05594ef5bbe89b94977223477ae4192197967ede1e  .moai/specs/SPEC-TODO-IDENTITY-001/acceptance.md
b667cf6c4025e895fdb8b020ec722dd21f71f2a935c5c342404165e3e50a64e9  internal/kanban/todo_identity_red_test.go
```

Exit code: 0.

#### E2 — lint

Command:

```bash
moai spec lint .moai/specs/SPEC-TODO-IDENTITY-001 --json
```

Observed stdout:

```json
[]
```

Exit code: 0.

#### E3 — REQ/AC structure and mapping

Command:

```bash
awk '/^## Requirements/{f=1;next}/^## /{f=0}f&&/^- \*\*REQ-TID-[0-9][0-9][0-9]\*\*:/{n++}END{print "REQ_COUNT=" n+0}' .moai/specs/SPEC-TODO-IDENTITY-001/spec.md
awk '/^## Acceptance Criteria/{f=1;next}/^## /{f=0}f&&/^- \*\*AC-TID-[0-9][0-9][0-9]\*\*:/{n++}END{print "INLINE_AC_COUNT=" n+0}' .moai/specs/SPEC-TODO-IDENTITY-001/spec.md
rg -o '\(maps REQ-TID-[0-9]{3}\)' .moai/specs/SPEC-TODO-IDENTITY-001/acceptance.md | sort | uniq -c
```

Observed stdout:

```text
REQ_COUNT=7
INLINE_AC_COUNT=7
   1 (maps REQ-TID-001)
   1 (maps REQ-TID-002)
   1 (maps REQ-TID-003)
   1 (maps REQ-TID-004)
   1 (maps REQ-TID-005)
   1 (maps REQ-TID-006)
   1 (maps REQ-TID-007)
```

Exit code: 0.

#### E4 — D7/D8/clarification gate

Commands:

```bash
for sid in $(rg -o 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' .moai/specs/SPEC-TODO-IDENTITY-001/spec.md | sort -u); do f=".moai/specs/$sid/spec.md"; if test -f "$f"; then spec_state=$(rg -m1 '^status:' "$f" | cut -d: -f2 | tr -d ' '); echo "$sid status=$spec_state"; else echo "$sid NOT_FOUND"; fi; done
if rg -q 'syscall' .moai/specs/SPEC-TODO-IDENTITY-001/spec.md; then if ! rg -q '//go:build|cross-platform exemption|EXCL.*syscall' .moai/specs/SPEC-TODO-IDENTITY-001/spec.md; then echo 'BLOCKING: syscall without constraint/exemption'; else echo 'PASS: syscall constrained'; fi; else echo 'PASS: syscall absent'; fi
rg -n '\[NEEDS CLARIFICATION' .moai/specs/SPEC-TODO-IDENTITY-001/plan.md
```

Observed stdout:

```text
SPEC-TODO-IDENTITY-001 status=draft
SPEC-TODO-RUNTIME-STORE-001 status=draft
SPEC-TODO-UNIFIED-001 status=draft
PASS: syscall absent
```

The clarification `rg` emitted no stdout and exited 1 (zero matches). D7 and D8 wrappers exited 0.

#### E5 — current historical selector re-execution

Command:

```bash
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -run '^TestTodoIdentity' -count=1 -v -timeout=120s
```

Observed stdout/stderr:

```text
=== RUN   TestTodoIdentityREDAddReturnAndProjectIsolation
    todo_identity_red_test.go:63: first project: missing project_uuid after successful existing API
    todo_identity_red_test.go:64: first Add return: missing card_uuid after successful existing API
    todo_identity_red_test.go:65: first LoadPure item: missing card_uuid after successful existing API
    todo_identity_red_test.go:63: second project: missing project_uuid after successful existing API
    todo_identity_red_test.go:64: second Add return: missing card_uuid after successful existing API
    todo_identity_red_test.go:65: second LoadPure item: missing card_uuid after successful existing API
    todo_identity_red_test.go:73: GAP: issuance absent; cross-project UUID inequality and Add-return equality not reached
--- FAIL: TestTodoIdentityREDAddReturnAndProjectIsolation (0.21s)
=== RUN   TestTodoIdentityREDLifetimeAcrossArchiveRestore
    todo_identity_red_test.go:88: initial Add: missing card_uuid after successful existing API
    todo_identity_red_test.go:108: archived item: missing card_uuid after successful existing API
    todo_identity_red_test.go:119: restored/reopened item: missing card_uuid after successful existing API
    todo_identity_red_test.go:121: GAP: lifecycle APIs executed, but identity stability cannot be evaluated without issued UUID
--- FAIL: TestTodoIdentityREDLifetimeAcrossArchiveRestore (0.10s)
=== RUN   TestTodoIdentityREDRuntimeLinksActualCard
    todo_identity_red_test.go:165: runtime project: missing project_uuid after successful existing API
    todo_identity_red_test.go:166: runtime card: missing card_uuid after successful existing API
    todo_identity_red_test.go:167: runtime run: missing project_uuid after successful existing API
    todo_identity_red_test.go:168: runtime assignment: missing project_uuid after successful existing API
    todo_identity_red_test.go:169: runtime assignment: missing card_uuid after successful existing API
    todo_identity_red_test.go:171: GAP: actual runtime values read successfully, but UUID linkage equality not reached
--- FAIL: TestTodoIdentityREDRuntimeLinksActualCard (0.32s)
=== RUN   TestTodoIdentityLegacyPureReadDoesNotIssueOrWrite
    todo_identity_red_test.go:197: legacy project_uuid raw="" (absence and null both observationally allowed)
    todo_identity_red_test.go:197: legacy card_uuid raw="" (absence and null both observationally allowed)
--- PASS: TestTodoIdentityLegacyPureReadDoesNotIssueOrWrite (0.00s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/kanban 0.924s
FAIL
```

Exit code: 1.

#### E6 — AC-002 planned selector

Command:

```bash
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -run '^TestTodoIdentityLegacy' -count=1 -v -timeout=120s
```

Observed stdout/stderr:

```text
=== RUN   TestTodoIdentityLegacyPureReadDoesNotIssueOrWrite
    todo_identity_red_test.go:197: legacy project_uuid raw="" (absence and null both observationally allowed)
    todo_identity_red_test.go:197: legacy card_uuid raw="" (absence and null both observationally allowed)
--- PASS: TestTodoIdentityLegacyPureReadDoesNotIssueOrWrite (0.00s)
PASS
ok   github.com/modu-ai/moai-adk/internal/kanban 0.312s
```

Exit code: 0. Executed tests: 1. Verdict for AC-002 RED-now: FAIL, because the test explicitly permits the forbidden key-absence case.

#### E7 — AC-003 planned selector

Command:

```bash
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -run '^TestTodoIdentity.*(Failure|Backfill|Restore)' -count=1 -v -timeout=120s
```

Observed stdout/stderr:

```text
=== RUN   TestTodoIdentityREDLifetimeAcrossArchiveRestore
    todo_identity_red_test.go:88: initial Add: missing card_uuid after successful existing API
    todo_identity_red_test.go:108: archived item: missing card_uuid after successful existing API
    todo_identity_red_test.go:119: restored/reopened item: missing card_uuid after successful existing API
    todo_identity_red_test.go:121: GAP: lifecycle APIs executed, but identity stability cannot be evaluated without issued UUID
--- FAIL: TestTodoIdentityREDLifetimeAcrossArchiveRestore (0.22s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/kanban 0.776s
FAIL
```

Exit code: 1. Executed tests: 1. Verdict for AC-003 RED-now: FAIL, wrong-reason RED; Failure/Backfill/rollback premise는 실행되지 않았다.

#### E8 — AC-006 planned selector

Command:

```bash
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test -race ./internal/kanban -run '^TestTodoIdentity.*(Concurrent|Future|Retired)' -count=1 -v -timeout=120s
```

Observed stdout/stderr:

```text
testing: warning: no tests to run
PASS
ok   github.com/modu-ai/moai-adk/internal/kanban 1.653s [no tests to run]
```

Exit code: 0. Executed tests: 0. Verdict for AC-006 RED-now: FAIL; empty sweep reproduces nothing.

#### E9 — AC-007 planned selector

Command:

```bash
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -run '^TestTodoIdentity.*Timestamp' -count=1 -v -timeout=120s
```

Observed stdout/stderr:

```text
testing: warning: no tests to run
PASS
ok   github.com/modu-ai/moai-adk/internal/kanban 0.854s [no tests to run]
```

Exit code: 0. Executed tests: 0. Verdict for AC-007 RED-now: FAIL; empty sweep reproduces nothing.

### Baseline-attribution

- Repository/worktree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/todo-unified`
- Branch: `WT-todo-unified`
- Git tree pin: `a315dad9af0d3a0e04862e6106b3993d9a3812f7`
- Plan artifacts: untracked; exact SHA-256 values are E1에 기록했다.
- RED test source: untracked `internal/kanban/todo_identity_red_test.go`, SHA-256 `b667cf6c4025e895fdb8b020ec722dd21f71f2a935c5c342404165e3e50a64e9`.
- Tier threshold SSOT: `.claude/rules/moai/workflow/spec-workflow.md:L141-L145` — Tier M 0.80.
- Lint executable: `/Users/goos/go/bin/moai`; lint result는 이 installed binary에 귀속하며 current-tree Go source build라고 주장하지 않는다.

### Gaps

- 작성자의 reasoning, 이전 대화, 초안은 M1에 따라 읽지 않았다.
- 공식 post-audit Implementation Kickoff Approval transcript/receipt는 artifact set에 없었다. `plan.md:L64`의 “정식게이트 미통과”를 현재 문서 상태로 채택했다.
- 구현 코드의 정당성이나 최종 schema 선택은 감사하지 않았다. 테스트 명령은 현재 검증 instrument의 실행·선택 수·RED 이유를 확인하기 위해서만 실행했다.
- `audit_model`의 명시적 project setting을 찾지 못했고 이 세션에 `mcp__moai__audit_multi`가 제공되지 않아 cross-backend convergence는 수행하지 않았다. primary verdict는 이 독립 감사자가 소유한다.
- 호출자가 요청한 `.moai/reports/plan-audit/SPEC-TODO-IDENTITY-001-review-1.md`는 `.gitignore:L273`와 audit-artifact convention의 FORBIDDEN disposal path에 해당한다. 이 파일에는 쓰지 않았고, 추적 가능한 SPEC-scoped review stream인 현재 파일로 export했다.

### Residual-risk

- spec/test artifacts가 untracked이므로 이 보고서 이후 수정되어도 Git tree SHA는 바뀌지 않는다. 재감사 시 E1의 artifact/test SHA-256을 반드시 다시 비교해야 한다.
- 현재 UUID 부재 RED는 구현 전 상태를 보여줄 뿐 UUIDv7 version/lowercase, equality, lifetime, backfill, rollback, concurrency, future/retired rejection, timestamp non-authority의 올바른 RED를 대신하지 않는다.
- active WAL의 file-set/hash/checkpoint contract가 정해지지 않은 상태에서는 “bytes unchanged” PASS가 main DB만 측정한 불완전 관측일 수 있다.

## Operational Notes (unverified)

- `measured` — artifact drift를 재감사 전에 측정: E1의 `shasum -a 256 ...`을 다시 실행하고 세 digest 및 test digest를 이 보고서와 비교한다.
- `inferred` — `related_specs`가 dependency DAG를 기계적으로 만들지 확인: canonical dependency key가 필요한 경우 `moai spec lint ... --json` 결과뿐 아니라 parser의 `dependencies` field 소비를 측정한다. 근거 규칙은 frontmatter schema의 Optional Fields와 lint dependency rule이다.
