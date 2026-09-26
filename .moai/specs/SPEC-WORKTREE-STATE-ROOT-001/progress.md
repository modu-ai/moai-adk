# SPEC-WORKTREE-STATE-ROOT-001 — Progress

Card: t1213 | Branch: WT-worktree-state-roots | Base: develop `c630de892` | Tier: M

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-26
tier: M
spec_version: "0.2.1"
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
spec_id_check: "Bash regex PASS on SPEC-WORKTREE-STATE-ROOT-001; directory absent from .moai/specs before creation"
baseline_tree: c630de892
premise_evidence: .moai/reports/t1213/remeasure-d18-d25.md
plan_audit: "iter-1 FAIL 0.71 (.moai/reports/t1213/plan-audit-iter1.md); 0.2.0 revision closes D1-D10 and D11-D14"
plan_audit_iter2: "FAIL 0.78 (.moai/reports/t1213/plan-audit-iter2.md); Tier M iteration cap reached"
operator_decision: "PASS-with-debt, 2026-09-26"
debt_conditions_applied: "N1-N3 applied in the 0.2.1 commit (N4 also applied; N5 open, optional)"
req_count: 16
ac_count: 16
kickoff_decisions: 5
unmeasured: "stdin cwd / CLAUDE_PROJECT_DIR delivered to a hook process in a worktree session (spec.md §5)"
```

## §E.2 Run-phase Evidence

Raw outputs are under `.moai/reports/t1213/run/` (gitignored, kept in this worktree).
Every `go test` call ran as one compound invocation prefixed with
`unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED &&`
(elided below). Optional M0 probe (plan.md §D): **skipped** — it gates no criterion.

### Commits (unpushed)

| SHA | Milestone |
|---|---|
| `e065ef1e9` | M1 start run phase — spec.md `draft → in-progress`, §F Mode Selection |
| `a086e87bd` | M1 RED tests (only `_test.go`): AC-WSR-001, -002, -006, -007, -008, -010, -014, -016 |
| `ab1cc2294` | M2 store-root answer (`auditreceipt.StoreRoot`), per-tree rejections, receipt guard (AC-007/008/009) |
| `c2024b4f0` | M3-M8 union catalogue, state writers/readers, `_root`, docs + template mirror, predecessor-test edits |
| `7c312eb22` | M9 store-root helper shape tests (coverage) |

### RED evidence (E8) — observed before GREEN

RED-first set, run on the RED commit's tree (`a086e87bd`, base production code):

```
$ go test ./internal/cli/ -run 'TestWSR0(01|02|06|10|14|16)' -count=1 -v   → exit 1
wsr_state_root_test.go:235: want exactly one receipt under P/.moai/state/audit-receipts/receipts, got 0: []
wsr_state_root_test.go:256: want exactly one convergence result for S under P's store, got 0: []
wsr_state_root_test.go:282: want two results for S in P's store (W fail, W2 pass), got map[]
wsr_state_root_test.go:288: 2a (cwd=W2): want block, got {} (detector calls [])
wsr_state_root_test.go:295: 2b (cwd=P): want block on the fail audit_multi wrote for W, got {} (detector calls [<tmp>/P])
row 2..18 (the 13 R=W rows): observed Wt, required Wr
spec_progress count = 1, want 2 ; spec_drift total_specs = 1, want 2 ; spec_audit total_specs = 1, want 2
W: want one MCP receipt with tree_root <tmp>/W under <tmp>/P/.moai/state, found 0
W: hook start marker not under <tmp>/P/.moai/state
W: CLI check "c-W" not recorded in the snapshot under <tmp>/P/.moai/state
W F-noid: codex_audit must carry a skip notice ... got {"verdict":"pass",...,"audit_receipt":"rcpt-..."}
W F-noid: moai verify record must fail naming the primary checkout, got err=verify record: verify key: rev-parse HEAD: ...
spec_progress on W must list SPEC-PRI-001 ; <tmp>/W/.moai must not exist (stat err=<nil>)
--- FAIL: TestWSR001 / TestWSR002 (cell1, cell2-store, cell2a, cell2b) / TestWSR006 / TestWSR010 / TestWSR014 / TestWSR016
$ go test ./internal/hook/ -run 'TestWSR00(7|8)' -count=1 -v   → exit 1
wsr_audit_receipt_tree_test.go:183: uncited PASS on W must be refused with AUDIT_RECEIPT_VIOLATION, got &{... Decision: Reason: ...}
wsr_audit_receipt_tree_test.go:187: want one start marker in P's store with tree_root W, got []
wsr_audit_receipt_tree_test.go:239: A1 and A3 uncited PASSes must be refused
wsr_audit_receipt_tree_test.go:258: spawn with cwd=W must be denied naming SPEC-X-001, got denied=false reason=""
```

Each fails on its named predicate, not on a build error. AC-WSR-006's observed per-row matrix on the RED
tree (lead-requested measurement): rows 1 → N; rows 3, 7-9, 19-27 → Pp; rows 2, 4-6, 10-18 → **Wt** (13 rows,
`codex calls=[] multi calls=[] multi block=false`), exactly the "Today" column. F-noid row 4 → N′ (PASS, preservation).

Criteria not in the RED-first set:

| AC | RED observation | How |
|---|---|---|
| AC-WSR-003, -005 | `verify_snapshot`/`verify_trend` return success on W; notices `""`; receipt + convergence under W; snapshot under W; `W/.moai` created | test written before the writer change; run at `ab1cc2294` + struct fields only |
| AC-WSR-004 | PASS on base (preservation criterion); its mutant is the discriminating input (below) | — |
| AC-WSR-009 | first stop not refused, re-entry silent, spawn allowed (`denied=false reason=""`), texts `""` | guard implementation preceded this test; RED observed by restoring the base `audit_receipt_guard.go` (`git show a086e87bd:…`), running, and restoring |
| AC-WSR-011, -012, -013 | no `source`, count 1, no `shadowed`, no `primary_catalogue`, no `sources`, old `worktree_warning` | union code preceded these tests; RED observed by disabling the catalogue/`_root` branches (base handler path) and restoring — test-after disclosed |
| AC-WSR-015 | 12 missing/stale doc assertions (section, shared description, 7 tool descriptions) | test run before any doc edit |

### Lead-requested: AC-WSR-002 cell 2b observed RED then GREEN, and mutant "P gate reads only S.json"

```
# RED, tree a086e87bd
$ go test ./internal/cli/ -run 'TestWSR002/cell2b' -count=1 -v   → exit 1
wsr_state_root_test.go:295: 2b (cwd=P): want block on the fail audit_multi wrote for W, got {} (detector calls [<tmp>/P])
wsr_state_root_test.go:297: 2b observed: output={}
--- FAIL: TestWSR002_ConvergenceStoreIdentityAndCoexistence/cell2b
# GREEN, HEAD c2024b4f0
$ go test ./internal/cli/ -run 'TestWSR002/cell2b' -count=1 -v   → exit 0
wsr_state_root_test.go:298: 2b observed: output={"decision":"block","reason":"multi-review-gate: cross-model disagreement (advisory, NOT a block): pass=[codex(required)] fail=[claude(required)]"}
--- PASS: TestWSR002_ConvergenceStoreIdentityAndCoexistence/cell2b
# Mutant B ("P gate reads only S.json": loadConvergenceResults accepts only <session>.json), on c2024b4f0
$ go test ./internal/cli/ -run 'TestWSR002' -count=1 -v   → exit 1
wsr_state_root_test.go:296: 2b (cwd=P): want block on the fail audit_multi wrote for W, got {} (detector calls [<tmp>/P])
--- PASS: cell1 ; --- PASS: cell2-store ; --- FAIL: cell2a ; --- FAIL: cell2b
```

(The block reason text is the convergence engine's pre-existing `residual_risk_note`; the decision is `block`.)

### Discriminating mutants — executed once each (HEAD `c2024b4f0`, restored after each run)

| Mutant | Criterion | Observed | Matches derivation |
|---|---|---|---|
| A: shared store keyed by session only (v0.1.0) | AC-WSR-002 | cell2-store `got map[W2:pass]` (one record); 2a `{}`; 2b `{}` | yes |
| B: P gate reads only `S.json` (N1) | AC-WSR-002 | cell2b `{}`, cell2a `{}`; cell1/cell2-store pass | yes |
| mapping applied to every linked worktree | AC-WSR-004 | WT: receipt, convergence, snapshot not under WT (both git runs) | yes |
| route opt-in flag only | AC-WSR-006 | `13 observed decoy-read, required Wr` | yes |
| route state read only | AC-WSR-006 | `13 observed Wt, required Wr` | yes |
| rejections without tree identity (v0.1.0, D1) | AC-WSR-007 | one untagged record remains; `spawn with cwd=P must be allowed, got denied` (cwd=W also denied) | yes |
| fail-closed at stop only (v0.1.0, D5) | AC-WSR-009 | `spawn must be denied ..., got denied=false reason=""` (both F-noid variants) | yes |
| primary catalogue only | AC-WSR-010 | count 1; totals 1; premise fails (`got map[SPEC-PRI-001:true]`) | yes |

AC-WSR-002's "D6" (convergence identity) is mutant A; "D1" (rejection identity) is the AC-WSR-007 mutant.

### Deviations / notes

- AC-WSR-007 ordering: A2's SubagentStart is handled before the `codex_audit` receipt it cites; the AC text
  lists the audit first, which the unchanged receipt-before-start check (`store.go` CheckCitedReceipts) would
  refuse. An auditor calls the audit during its run, so start-then-audit is the realistic order.
- AC-WSR-008 / -007 (hook package) write the P-store receipt with `auditreceipt.WriteReceipt` (the hook package
  cannot import `internal/cli`); the real `codex_audit` → guard path is covered end to end by AC-WSR-016.
- AC-WSR-014 row B: `moai verify record` on a non-git root fails at working-tree key computation on base and
  head alike (before any store is chosen); the MCP and hook writers are asserted there.
- For a root that is not config-orphaned, rejections now also carry `tree_root` (permitted by REQ-WSR-005);
  paths are byte-identical (legacy file names kept for a tree equal to its store).
- `audit_multi` with no `project_root` keeps writing `<session>.json` without `tree_root`; readers treat an
  untagged record as the store's own tree (REQ-WSR-003 legacy rule).

### E1 — AC matrix (GREEN, HEAD `c2024b4f0` for AC tests; package runs at `7c312eb22`)

| AC | Status | Command (`go test … -count=1 -v`) | Observed |
|---|---|---|---|
| AC-WSR-001 | PASS | `./internal/cli/ -run TestWSR001` | `--- PASS: TestWSR001_ReceiptInPrimaryStoreWithWorktreeIdentity` |
| AC-WSR-002 | PASS | `./internal/cli/ -run TestWSR002` | cell1, cell2-store, cell2a, cell2b all `--- PASS` |
| AC-WSR-003 | PASS | `./internal/cli/ -run TestWSR003` | both F-noid variants `--- PASS` |
| AC-WSR-004 | PASS | `./internal/cli/ -run TestWSR004` + predecessor run below | `--- PASS`; mutant kills it |
| AC-WSR-005 | PASS | `./internal/cli/ -run TestWSR005` | `--- PASS` |
| AC-WSR-006 | PASS | `./internal/cli/ -run TestWSR006` | 27 rows: 1 → N, 2/4-6/10-18 → **Wr** (13), 3/7-9/19-27 → Pp; F-noid row 4 → N′ |
| AC-WSR-007 | PASS | `./internal/hook/ -run TestWSR007` | `--- PASS: TestWSR007_RejectionsArePerTree` |
| AC-WSR-008 | PASS | `./internal/hook/ -run TestWSR008` | `--- PASS` |
| AC-WSR-009 | PASS | `./internal/hook/ -run TestWSR009` | 3 subtests `--- PASS` (incl. tracked-`.moai` WT: no fail-closed) |
| AC-WSR-010 | PASS | `./internal/cli/ -run TestWSR010` | `--- PASS` |
| AC-WSR-011 | PASS | `./internal/cli/ -run TestWSR011` | `--- PASS` |
| AC-WSR-012 | PASS | `./internal/cli/ -run TestWSR012` | both variants `--- PASS` |
| AC-WSR-013 | PASS | `./internal/cli/ -run 'TestWSR013\|TestConfigOrphanedWorktree_CatalogueWarning'` | both tests reported `--- PASS` (swept count 2); predecessor unedited |
| AC-WSR-014 | PASS | `./internal/auditreceipt/ ./internal/hook/ ./internal/cli/ -run TestWSR014` | auditreceipt `TestWSR014_StoreRootPerRoot` PASS, cli `TestWSR014_OneStoreRootEverywhere` PASS; hook package has no TestWSR014 (`[no tests to run]` — the hook writer is exercised from the cli test) |
| AC-WSR-015 | PASS | `./internal/cli/ -run TestWSR015` + `diff` of rule vs mirror | `--- PASS`; `cmp` identical |
| AC-WSR-016 | PASS | `./internal/cli/ -run TestWSR016` | `--- PASS` |

Predecessor tests (AC-WSR-004): exactly the two named tests were edited —
`TestConfigOrphanedWorktree_PrimaryGateEnforced` (premise now "`W/.moai` does not exist after the first call";
round 2 creates `W/.moai/state` itself) and `TestLinkedWorktree_DescriptionsStateTheGateSource` (the two removed
phrases replaced by "state is kept in the primary checkout's .moai/state" and "hook-side receipt guard").

### Full affected-package runs (HEAD `7c312eb22`, this worktree)

```
$ go test -cover ./internal/auditreceipt/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/auditreceipt	0.839s	coverage: 88.5% of statements
$ go test -cover ./internal/spec/ ./internal/verify/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/spec	110.655s	coverage: 90.6% of statements
ok  	github.com/modu-ai/moai-adk/internal/verify	2.762s	coverage: 84.6% of statements   (no source change in this package)
$ go test -cover ./internal/hook/... -count=1        (slot go-test-heavy held)
ok  	github.com/modu-ai/moai-adk/internal/hook	267.165s	coverage: 85.9% of statements
(10 sub-packages ok)
$ go test -timeout 25m -cover ./internal/cli/ -count=1   (slot go-test-heavy held)
ok  	github.com/modu-ai/moai-adk/internal/cli	1063.487s	coverage: 84.1% of statements
```

A first `./internal/cli/` run without `-timeout` hit the default 10 m limit (`panic: test timed out after
10m0s`, running test `TestResolveArmSessionID_MultiSessionWarns (0s)`, zero `--- FAIL` lines before it); the
package simply takes ~18 m here. Touched-function coverage (`-run 'TestWSR|TestConfigOrphaned|…'
-coverprofile`): new helpers 75-100 %, e.g. `unionAudit` 91.7 %, `loadConvergenceResults` 79.2 %,
`persistConvergenceResult` 88.9 %, `recordAuditReceipt` 78.6 %.

### E2 / E4 / E5

```
$ go build ./...                            → exit 0   (baseline 638458f08: exit 0)
$ GOOS=windows GOARCH=amd64 go build ./...  → exit 0   (baseline: exit 0)
$ golangci-lint run --timeout=5m ./internal/cli/... ./internal/hook/... ./internal/auditreceipt/... ./internal/verify/... ./internal/spec/...
0 issues.                                   (baseline 638458f08: 0 issues.)
$ go vet ./internal/cli/ ./internal/hook/ ./internal/auditreceipt/ ./internal/spec/ ./internal/verify/   → clean
$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook internal/cli internal/auditreceipt | grep -v _test.go | grep -v "// " | wc -l
34   (baseline `git grep` at 638458f08: 34; lines added by this run: 0 — help text and the tool-name check in pre_tool.go)
$ make build → exit 0 ; cmp rule vs template mirror → identical
```

### Gaps

- `internal/cli` package coverage 84.1 % was not compared against a base-tree measurement (an 18-minute run).
- The M0 live-payload probe was skipped, so which AC-WSR-006 row a real worktree session lands on stays unmeasured.
- The full repository suite was not run locally (CI on the develop push is the verdict).


## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-26
run_commit_sha: pending-backfill-run   # the commit carrying this block; last code commit 7c312eb22
run_status: audit-ready
ac_pass_count: 16
ac_fail_count: 0
red_first_witnessed: [AC-WSR-001, AC-WSR-002, AC-WSR-006, AC-WSR-007, AC-WSR-008, AC-WSR-010, AC-WSR-014, AC-WSR-016]
red_commit: a086e87bd
mutants_executed: 8
preserve_list_post_run_count: "2 predecessor tests edited (named in AC-WSR-004); no other existing test edited"
l44_pre_commit_fetch: not-run   # lane does not push; lead batch-pushes develop
l44_post_push_fetch: not-applicable
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: pass
  windows_amd64: pass
total_run_phase_files: 23   # git diff --name-only 638458f08..HEAD, incl. spec.md + progress.md
m1_to_mN_commit_strategy: "per-milestone commits, unpushed (e065ef1e9, a086e87bd, ab1cc2294, c2024b4f0, 7c312eb22)"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Input parameters: tier M; scope ~12-16 files (Go source + tests in internal/auditreceipt,
internal/hook, internal/cli, internal/spec, one rule file + template mirror); domains 3
(Go source, rule doc, template mirror); language mix mostly Go; concurrency benefit LOW
(coding-heavy, milestones depend on one shared store-root answer); Agent Teams not requested.

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | multi-file semantic change, not a typo |
| serial | yes | coding-heavy; M2-M7 all consume the M1 store-root function |
| fanout | no | not research-heavy; parallel writers would share `internal/cli` |
| sweep | no | not a uniform mechanical transform |

Decision: serial

Justification: every milestone routes through one store-root answer introduced first, so
the work is sequential by dependency; Anthropic's coding-task parallelism caveat applies.
Implementation Kickoff Approval was granted by the operator on 2026-09-26 with all five
plan.md §C decisions at their recommended defaults, progression autonomous.
