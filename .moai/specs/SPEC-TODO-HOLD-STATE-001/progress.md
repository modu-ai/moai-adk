# SPEC-TODO-HOLD-STATE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
spec: SPEC-TODO-HOLD-STATE-001
phase: plan
plan_status: audit-ready
tier: M
measured_at:
  head: 8a969dfc0
  worktree: .moai/worktrees/t1308
  branch: WT-todo-hold-state
  date: 2026-09-29
artifacts:
  - spec.md
  - plan.md
  - acceptance.md
  - progress.md
requirements: 16
acceptance_criteria: 16
tier_ceiling_disposition: |
  plan-audit iter1 D2 (19/19 vs Tier M 16/16): resolved by CONSOLIDATION (auditor option a).
  Folds: REQ-THS-007→REQ-THS-006, REQ-THS-009→REQ-THS-008 (refusal halves absorbed as second
  When-clauses), REQ-THS-015→REQ-THS-011 (autodone covered by the actionable-surface enumeration);
  AC-THS-007→AC-THS-006, AC-THS-009→AC-THS-008, AC-THS-014→AC-THS-011(c, regression guard).
  Freed numbers (REQ/AC 007, 009, REQ 015, AC 014) are deliberate consolidation gaps recorded in
  spec.md HISTORY 0.2.0 — not authoring errors. Tier-up and split were rejected: the content fits
  M without the two additional Tier L artifacts.
lead_memo_survival:
  rebuild_safety_contract: "REQ-THS-002/003/004 — ids and content intact (AC-THS-002/003/004/019)"
  negative_predicate_sweep: "REQ-THS-011/012 — ids and content intact; mutation-drift AC = AC-THS-011 (id intact), autodone regression-guard row absorbed as (c)"
premise_corrections:
  - "카드 전제 「internal/kanban 이력상 ALTER TABLE 0건」 기각 — ensureLandingColumn (backlog_sqlite.go:439-455, t359)의 ADD COLUMN 선례 1건 존재. 마이그레이션 결정(리빌드+스탬프)은 불변."
  - "iter1 D1 정정 — todo_autodone.go:283 의 부정 disjunction 은 미래 상태를 삼키는 것이 아니라 이미 건너뛴다 (행동 정상·형태만 결함). 행동 red-now 는 todo.go:920 픽 게이트 (dropped 만 거절). autodone AC 는 green-at-M1 회귀 가드로 재분류."
lead_memos_folded:
  - "부정 술어 전수 전환 REQ-THS-011/012 (mutation-testable drift AC-THS-011)"
  - "리빌드 안전 계약 REQ-THS-002/003/004 (JSON→SQLite 규율 계승, SPEC-TODO-SQLITE-001)"
```

## §E.2 Run-phase Evidence

Run entry: plan-audit SKIP-ELIGIBLE (iter2 PASS 0.96 ≥ Tier M 0.80 + artifact hash
unchanged since d69b71c6d per plan-audit-codex.md; three-condition skip contract
satisfied, recorded in the run delegation prompt). Cycle: TDD (RED-GREEN-REFACTOR),
milestones M1-M5, one commit each. All evidence below measured in THIS run, against
THIS tree (worktree `.moai/worktrees/t1308`, branch `WT-todo-hold-state`), final HEAD
noted per row. Verbatim RED evidence: `.moai/reports/t1308/mutation-evidence.md` holds
the mutation runs; the RED-now captures are quoted in §E.2 rows E8 below.

### E1 — AC binary matrix (AC-THS-001..016; 007/009/014 are deliberate consolidation gaps)

| AC | Status | Verification command (verbatim) | Actual output (verbatim line) |
|----|--------|--------------------------------|-------------------------------|
| AC-THS-001 | PASS | `go test ./internal/kanban -run 'TestBacklogItemsTableAdmitsExactlyFourStates' -count=1` | `ok github.com/modu-ai/moai-adk/internal/kanban` |
| AC-THS-002 | PASS | `go test ./internal/kanban -run 'TestBacklogV1ToV2MigrationRoundTrip' -count=1` | `ok github.com/modu-ai/moai-adk/internal/kanban` |
| AC-THS-003 | PASS (natural-failure path; parity-failure path via Mutation 1 positive control) | `go test ./internal/kanban -run 'TestBacklogMigrationFailureLeavesOriginalFileUntouched' -count=1` | `ok github.com/modu-ai/moai-adk/internal/kanban` |
| AC-THS-004 | PASS | `go test ./internal/kanban -run 'TestBacklogMigrationStampAndRebuildAreOneTransaction' -count=1` | `ok github.com/modu-ai/moai-adk/internal/kanban` |
| AC-THS-005 | PASS | `go test ./internal/kanban -run 'TestBacklogOpenRefusesForeignSchemaVersion' -count=1` | `ok github.com/modu-ai/moai-adk/internal/kanban` |
| AC-THS-006 | PASS | `go test ./internal/cli -run 'TestTodoHold_MovesQueuedCardToHold|TestTodoHold_RefusesNonQueuedStates|TestTodoHold_ExpectMismatchRefuses' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli` |
| AC-THS-008 | PASS | `go test ./internal/cli -run 'TestTodoUnhold_ReturnsHeldCardToQueued|TestTodoUnhold_RefusesNonHeldStates' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli` |
| AC-THS-010 | PASS | `go test ./internal/cli -run 'TestTodoHold_ActorBoundaryLeasePathsCannotHold' -count=1` + `bin/moai factory assign --help | grep -ci hold` | `ok ...` / `0` |
| AC-THS-011 | PASS | `go test ./internal/cli -run 'TestTodoSelectionPredicatesPositivelyEnumerateStates|TestTodoFutureStateCardIsNeverSelectedByActionablePaths|TestTodoAutoDoneSkipsHeldCard' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli` |
| AC-THS-012 | PASS (pin at shared selection layer; named leaser functions not in tree — t1240/t1294 in flight) | `go test ./internal/cli -run 'TestTodoMachineLeaseSelectsOnlyQueued' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli` (5 shuffled orders) |
| AC-THS-013 | PASS | `go test ./internal/cli -run 'TestTodoNext_PickOnHeldCardRefused' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli` |
| AC-THS-015 | PASS | `go test ./internal/cli -run 'TestTodoListRendersHeldCardTruthfully' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli` |
| AC-THS-016 | PASS | `go test ./internal/statusline -run 'TestBacklogCountsForRootExcludesHeldCards' -count=1` | `ok github.com/modu-ai/moai-adk/internal/statusline` |
| AC-THS-017 | PASS | `go test ./internal/web -run 'TestTodoQueueRendersHeldCardState' -count=1` | `ok github.com/modu-ai/moai-adk/internal/web` |
| AC-THS-018 | PASS | `go test ./internal/cli -run 'TestTodoHoldDocumentedOnEverySurface' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli` |
| AC-THS-019 | PASS | `go test ./internal/kanban/... -count=1` (landing evidence + archive preserved in the roundtrip; sibling storage-freeze tests re-run unmodified-passing at the new schema) | `ok github.com/modu-ai/moai-adk/internal/kanban 181.026s` (full package) |

### E2 — builds

```
$ go build ./...                                    → exit 0
$ GOOS=windows GOARCH=amd64 go build ./internal/kanban ./internal/cli → exit 0
$ make build                                        → exit 0 (catalog hashes regenerated after the gtd.md edit)
$ go vet ./internal/kanban/... ./internal/cli/... ./internal/statusline/... ./internal/web/... → exit 0
```

### E3 — coverage (this run, this tree, HEAD bb6196f0f)

```
$ go test ./internal/kanban -cover -count=1                       (full package, slot-leased)
ok  github.com/modu-ai/moai-adk/internal/kanban 182.796s  coverage: 86.3% of statements
$ go test ./internal/cli -cover -count=1                          (full package, slot-leased)
coverage: 55.9% of statements
```

- kanban package: **86.3%** — the 85% gate holds.
- cli package-wide: 55.9% (pre-existing shape — the package spans every CLI
  subcommand). Touched-surface figures instead: `todo_hold.go` per-function
  100.0% / 95.7% / 95.7%; `assertRebuildParity` 100.0% (direct unit table over
  every tuple mismatch); `readItemsRows` 83.3%; `rebuildItemsTable` 53.7% — the
  uncovered statements are the per-step error-return paths (begin/read/drop/
  rename/index/commit faults) the SPEC requires no fault injection for; the
  load-bearing paths (happy rebuild, copy-failure abort, parity-failure abort)
  are covered.
- statusline / web: scoped runs `ok`; the pinned functions
  (`BacklogCountsForRoot`, `readTodoQueue`, `todoStateBadge` render) are the
  tested seams.

### E4 — subagent boundary

```
$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/kanban internal/cli internal/statusline internal/web | grep -v "_test.go" | grep -v "// "
```
New matches: **0**. Baseline attribution: the only matches are comment-continuation
prose lines in `harness.go`, `pr_watch_cmd.go`, `codex_*.go` (pre-existing at base
`c5ec67741`, files untouched by this SPEC) stating the boundary itself.

### E5 — lint (golangci-lint v2.1.6, CI판)

```
$ golangci-lint run ./internal/kanban/... ./internal/cli/... ./internal/statusline/... ./internal/web/...
0 issues.
```
NEW issues: 0. Pre-existing: 0 in the touched packages (the plan §E note's accepted
`-run` pattern warnings did not fire on v2.1.6).

### E6 — branch state

Branch `WT-todo-hold-state`, **NOT pushed** (lane discipline — the lead integrates).
Commits, in order: `e1d6b512e` (M1 storage), `966c821fc` (M2 verbs), `d6683ec5b` (M3
sweep), `634fc5c7f` (M4 display), `b365d9e82` (M5 docs), `9cefc4fbd` (catalog hash
cascade), `bb6196f0f` (coverage completion + gtd verb-parity declaration). Base:
`8a969dfc0` (develop tip, origin-reflected); plan close `c5ec67741` / `d69b71c6d`
precede M1 on this branch.

### E7 — mutations

`.moai/reports/t1308/mutation-evidence.md` — three mutations, each applied once,
evidenced, reverted; tree verified clean after reverts. Mutation 1 (parity removed +
row loss → silent loss caught by the roundtrip test) with its positive control (parity
intact → migration aborts, file preserved). Mutation 2 (pick-gate default removed →
future-state card admitted, drift test fails). Mutation 3 (autodone default removed →
held card archived, regression guard fails).

### E8 — RED evidence

Verbatim pre-GREEN outputs (measured in this run):

- Behavioral red-now (AC-THS-013 / AC-THS-011(b)) — pick of a held card BEFORE the gate
  fix, seeded via the store to isolate the gate from the verb:
  ```
  --- FAIL: TestTodoNext_PickOnHeldCardRefused (0.67s)
      todo_hold_test.go:210: picking a held card must be refused — the held card is not a pick candidate
  ```
- Verb-absence RED (AC-THS-006/008) — `todo hold` before M2:
  ```
  todo: "hold" is not a todo verb and "1" is a card reference — refusing to create a card named "hold 1".
  ```
- Migration-symbol-absence RED (AC-THS-001..005) — before M1:
  ```
  internal/kanban/backlog_hold_migration_test.go:449:2: undefined: backlogSchemaVersionOverride
  internal/kanban/backlog_hold_migration_test.go:449:33: undefined: backlogSchemaVersionV1
  FAIL github.com/modu-ai/moai-adk/internal/kanban [build failed]
  ```
- Form RED (AC-THS-011(a)) — the sweep test before M3 found the five negative filters
  (`todo.go` x2, `todo_autodone.go`, `todo_drop.go`, `goal.go`) and the two
  positive-control misses (`todo_autodone.go`, `factory_card.go`).
- Regression guards declared honestly, no RED fabricated: AC-THS-011(c)
  (autodone skip) and AC-THS-012 (lease pin) were GREEN on arrival by design; their
  green-for-the-right-reason evidence is the flip positive control inside each test plus
  Mutation 3 / the in-test mutation control. AC-THS-015..018 (display/docs) flipped at
  their milestones; their pre-milestone state was value-absence, pinned by the adopted
  cell table, not re-run as fake REDs.

### Baseline separation (B5)

Pre-existing failures: **0** observed — the Section C baselines were green before any
change (`go test ./internal/kanban/...` → `ok ... 222.131s`; `go test ./internal/cli
-run 'TestTodo|TestBacklog'` → `ok ... 387.808s`). Card-attributed NEW failures during
the run: 7 kanban + 3 cli, ALL schema/surface-freeze tests pinning the v1 contract the
SPEC redefines (schema-freeze, downgrade direction, version pins x3, CHECK tuple,
verb-surface additions, landed CHECK pin, gtd/todo verb-parity list) — each updated in
its milestone commit with the SPEC clause named. Final scoped suites: kanban `ok`,
cli families `ok ... 365.247s`, statusline `ok`, web `ok ... 14.870s`.
Pre-existing, NOT card-attributed: `TestRunInit_QuietWizardUnsetResolvesToDefaults` +
`TestRunInit_QuietWizardObserverDetectsNonDefault` (llm.yaml template-default
assertions — `git diff 8a969dfc0..HEAD --name-only | grep -cE "run_init|wizard|llm"`
→ 0: the tests' inputs are byte-identical between base and HEAD, so the failures
exist at base and are outside this card's scope; left untouched per scope
discipline).

## §E.3 Run-phase Audit-Ready Signal

```yaml
spec: SPEC-TODO-HOLD-STATE-001
phase: run
run_status: complete
tier: M
run_complete_at: 2026-09-29
run_commit_sha: pending-backfill-run
branch: WT-todo-hold-state
worktree: .moai/worktrees/t1308
base: 8a969dfc0
commits:
  - e1d6b512e  # M1 storage: hold state, v1→v2 rebuild, stamp "2", compat seam
  - 966c821fc  # M2 verbs: hold/unhold + held-card pick refusal
  - d6683ec5b  # M3 predicate sweep: positive enumeration + drift guards
  - 634fc5c7f  # M4 display truthfulness pins
  - b365d9e82  # M5 docs: gtd.md + todo.md + mirrors + published skill
  - 9cefc4fbd  # catalog.yaml hash cascade (Template-First)
  - bb6196f0f  # coverage completion + gtd verb-parity declaration
ac_pass_count: 16
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: skipped (lane-local worktree; no push per lane discipline)
l44_post_push_fetch: n/a (not pushed)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: pass
  windows_amd64: pass
total_run_phase_files: 24
m1_to_mN_commit_strategy: one commit per milestone (M1-M5) + one cascade chore + one coverage-completion test commit
evidence_paths:
  - .moai/reports/t1308/predicate-sweep-run.md
  - .moai/reports/t1308/mutation-evidence.md
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## 진행 기록

- 2026-09-29 (plan, card t1308, lane worker-72): Tier M artifact set authored in worktree
  `.moai/worktrees/t1308`. Measured basis exported to `.moai/reports/t1308/predicate-sweep.md`
  (state vocabulary, CHECK/DDL comment, schema_version machinery, predicate inventory S1-S13/D1-D4,
  actor precedent, doc surfaces). Two lead memos folded as REQs. One card premise falsified
  (ALTER TABLE 0건 → 1건 ADD COLUMN 선례), decision unchanged. Lint evidence: see
  `.moai/reports/t1308/lint.txt`.
- 2026-09-29 (plan-audit iter1 remediation, v0.2.0): FAIL 0.90 (2 blocking) — D1 autodone
  filter-direction 정정 (4개 표면: acceptance RED-now 셀·plan §A/§B.5·sweep S5·spec §A.1/§B.4,
  회귀 가드 재분류), D2 16/16 통합 처분 (위 tier_ceiling_disposition), D3 Event-detected →
  Event-driven 재표기 (REQ-THS-003/005/014), D4 P3 전제 인라인 인용+primary-checkout-local 표기.
  재측정: REQ 16 / AC 16, lint 0 error. 커밋 없음 — 리드 검토 후 커밋.
- 2026-09-29 (run 진입 게이트 처분, card t1308): **운영자 승인(리드 전달 2026-09-29):
  served_model_gate 해제 + GLM 채택, OVERTURN 시 제외.** 근거: 비GLM 독립 경로 3종이 구조적
  불가로 실측 — claude_audit CLAUDE_CAPACITY_UNAVAILABLE / codex_role_audit 서버 루트 결속
  (codex_audit_launch.go:309, worker-66 소스 확인) / worker-63 "opus" 배차도 GLM 서빙. 채택
  사슬: GLM iter2 PASS 0.96 + worker-63 무변경 재확인 + codex(GPT-6) RECONFIRM(샌드박스
  거부로 계측은 레인이 수행) + 레인 자체 무변경 계측(팁==d69b71c6d·이후 0커밋·4파일 체크섬
  일치 — plan-audit-codex.md). 카드 트리 workflow.yaml의 served_model_gate.enabled를 false로
  (이 트리만; 기존 거부 영수증 보존). OVERTURN 발생 시 채택 제외.
- 2026-09-29 (run, card t1308, lane worker-72): TDD 사이클 M1-M5 착지, 마일스톤당 1커밋
  (e1d6b512e→9cefc4fbd). 유일한 행동 red-now(todo.go:920 픽 게이트) RED 캡처 후 폐쇄;
  M3 스윕은 실행 HEAD에서 재산출 — 계획 시 인벤터리와 무드리프트, 신규 사이트 없음
  (t1240/t1294 미착지 확인). SPEC-required 테스트 갱신 9건: kanban 7(스키마 프리즈·다운그레이드
  방향 전환·버전 핀 3·CHECK 튜플·컬럼 순서 수렴), cli 2(동사 표면 선언·landed CHECK 핀).
  변이 3건(패리티 제거+행 손실 / 픽 게이트 default 제거 / autodone default 제거) 각각
  실패 출력 기록 후 복원 — mutation-evidence.md. 최종 스코프드 스위트 전부 green,
  lint 0, windows 빌드 pass, push 없음(리드 통합 대기).
