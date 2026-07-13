# SPEC-CLI-TUX-V3-003 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-07-13

## §E.2 Run-phase Evidence

### Pre-flight baselines (2026-07-14, M3a entry)

| Check | Command | Result |
|-------|---------|--------|
| git sync | `git rev-list --count --left-right origin/main...HEAD` | `0 0` (synced, HEAD b45c552ea) |
| cross-platform build | `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` | darwin exit 0, windows exit 0 |
| guard 5-family anchor | `go test ./internal/template/ -run TestSplitHarnessNamespaceNoLeak` + `go test ./internal/merge/...` | both PASS |
| fmt.Printf baseline (AC-TUX3-020) | `grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' \| grep -v '_test.go' \| wc -l` | **38** (M3f target: below this) |
| bubbletea/bubbles v2 state (AC-TUX3-017) | `go list -m charm.land/bubbletea/v2 charm.land/bubbles/v2` | **already landed by SPEC-CLI-TUX-V3-002**: bubbletea/v2 v2.0.8, bubbles/v2 v2.1.1 (go.mod lines 6-7) |
| confirm.go @MX debt tags | `grep -n '@MX:DEBT\|@MX:CEILING\|@MX:UPGRADE' internal/merge/confirm.go` | 3 tags at lines 871-873 (M3e removal target; untouched this spawn) |

### M3a-1 Classification type contract (REQ-TUX3-001/002) — LANDED

New package `internal/cli/update/` with `class.go` + `class_test.go`:

- **`ChangeClass` enum** (4 values, single shared type for TUI preview + text fallback — REQ-TUX3-001):
  `ClassAdd` / `ClassUpdate` / `ClassPreserveUserOwned` / `ClassConflict`
- **`Classify(relPath, exists, conflict, isUserOwned)`** — priority order: user-owned > conflict > add(!exists) > update. The preserve class delegates entirely to the injected `UserOwnedPredicate` (REQ-TUX3-002 — no parallel heuristic). At M3a the predicate is injected (package `cli` passes `isUserOwnedNamespace`); M3d moves the predicate into the package for a same-package call.
- **`ChangeClass.String()`** labels — preserve class renders `"preserved (user-owned)"` (REQ-TUX3-014).
- **@MX:ANCHOR** on `Classify` (fan_in grows to ≥3 at M3c preview + M3e fallback + M3f enforcement coherence).

M3a gate (classification contract tests green):
```
go test ./internal/cli/update/... -count=1 -v
→ TestClassificationExhaustive PASS (10 subtests)
→ TestChangeClassString PASS
→ TestPreserveClassSource PASS
→ TestPredicateShared PASS
ok  github.com/modu-ai/moai-adk/internal/cli/update  0.252s
```

**AC-regex PASS-WITH-DEBT note (AC-TUX3-001/002 wording)**: acceptance.md shows `-run 'Classification\|ChangeClass'` and `-run 'PreserveClassSource\|PredicateShared'`. In Go RE2, `\|` is a literal-pipe escape (matches the `|` character), NOT alternation — so the regexes as written match zero test names. The intended alternation is plain `|`. My test names (`TestClassificationExhaustive`, `TestChangeClassString`, `TestPreserveClassSource`, `TestPredicateShared`) satisfy the INTENT with `-run 'Classification|ChangeClass'` / `-run 'PreserveClassSource|PredicateShared'` (verified green above). This is an acceptance.md wording fix owned by manager-spec (not a code defect). Recording here per the deferred-artifact obligation; no code action at M3a.

### M3a-2 update.go function→subpackage mapping table (50 funcs → 6 buckets)

Source: `grep -n '^func ' internal/cli/update.go` (3,276 LOC, 50 top-level funcs). Target subpackages per AC-TUX3-004: `internal/cli/update/{plan,backup,deploy,merge,report}` + root `update.go` retains cobra wiring/orchestration glue. Counts: ROOT=18, PLAN=8, BACKUP=8, DEPLOY=6, MERGE=9, REPORT=1 (total 50).

**ROOT (stays in update.go — cobra wiring + orchestration glue + utilities)**:
`validateUpdateFlags`:70, `init`:78, `runUpdate`:136, `applyUpdateTierProfile`:407, `shouldSkipBinaryUpdate`:453, `runBinaryUpdateStep`:482, `reexecNewBinary`:536, `runTemplateSync`:569, `runShellEnvConfig`:1595, `runInitWizard`:2645 (PRESERVE M2/-002), `applyWizardConfig`:2720 (PRESERVE M2/-002), `updateSettingsLocalEnv`:2928, `ensureGlobalSettingsEnv`:2938, `cleanLegacyHooks`:3064, `execCommand`:3160, `detectGoBinPathForUpdate`:3172, `resolveMoaiExecutable`:3181, `readHookOptInEnabled`:3193

**PLAN (→ update/plan — analysis/classification/namespace)**:
`classifyFileRisk`:1056, `determineStrategy`:1077, `determineChangeType`:1244, `analyzeFiles`:1260, `isUserAreaPath`:1313, `isUserOwnedNamespace`:1387 (ANCHOR predicate), `isMoaiManaged`:1488, `getProjectConfigVersion`:1639

**BACKUP (→ update/backup — config backup/restore + rotation)**:
`backupMoaiConfig`:1685, `saveTemplateDefaults`:1807, `cleanup_old_backups`:2068, `restoreMoaiConfig`:2132, `restoreMoaiConfigLegacy`:2239, `isSymlinkEntry`:2310, `restoreTargetContained`:2342, `parentChainContained`:2415

**DEPLOY (→ update/deploy — template deploy + migration + cleanup)**:
`runTemplateSyncWithReporter`:579, `runTemplateSyncWithProgress`:1005, `cleanMoaiManagedPaths`:1868, `migrateLegacyMemoryDir`:1983, `runAgencyMigrationAdapter`:2039, `scaffoldEvolutionDir`:3216

**MERGE (→ update/merge — 3-way + gitignore + analysis)**:
`mergeGitignoreFile`:1099, `mergeUserFiles`:1149, `buildMergeAnalysis`:1542, `analyzeMergeChanges`:1585, `mergeYAML3Way`:2488, `deepMerge3Way`:2511, `valuesEqual`:2571, `mergeYAMLDeep`:2584, `deepMergeMaps`:2602

**REPORT (→ update/report — guidance/preview render; grows at M3c/M3e)**:
`emitHooksReviewGuidance`:442 (single function; preview table + outcome card renderers land here at M3c/M3e)

### M3a-3 Namespace predicate caller map (PRESERVE safety — M3d move anchor)

- `isUserOwnedNamespace` (update.go:1387) — 3 call sites in `update_namespace_protect.go` (lines 153, 264, 329). Test files referencing it (M3d import-path adjustment): `update_namespace_hns_test.go`, `update_namespace_harness_v2_test.go`, `update_namespace_harness_v4_test.go`, `update_test.go`, `update_preserve_my_harness_test.go`, `update_evolution_test.go`.
- `isMoaiManaged` (update.go:1488) — 1 call site (update.go:1270).
- `isUserAreaPath` (update.go:1313) — 0 non-test callers (purely tested). Test refs: same 6 files above.
- `isUserOwnedNamespaceConservative` (protect.go:313) — stays in protect.go; calls `isUserOwnedNamespace` internally.

### M3b characterization safety net — LANDED (separate commit)

See `internal/cli/update/characterization_test.go`. Suite green on UNDECOMPOSED update.go (M3d anchor established).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

decided_at: 2026-07-14 (run kickoff)
orchestrator_session: 86ac7ac3-d1d9-43a4-8ec9-55732930b7b2
handoff_source_session: ec3aa70e-e1ac-4a68-ad48-02b61a1ee323

Input parameters:
- tier: L
- scope (file count): ~15+ (update.go decompose target + 5 new internal/cli/update/* subpackages + internal/merge/{confirm,confirm_test,confirm_coverage} 3 + go.mod/go.sum + characterization/preview/coverage tests)
- domains: 2 (internal/cli Go code + internal/merge)
- file language mix: Go 100%
- concurrency benefit: LOW (sequential dependency — M3b characterization MUST precede M3d decomposition; M3a classification contract gates M3c/M3e display layer)
- Agent Teams prereqs: N/A (Mode 3 RETIRED)

Mode evaluation:
- Mode 1 (trivial): not selected — Tier L 3,276 LOC decomposition + TUI, not trivial
- Mode 2 (background): not selected — write-capable coding work, blocking delegation
- Mode 3 (agent-team): not selected — RETIRED
- Mode 4 (parallel): not selected — coding-heavy + sequential milestone dependency (Anthropic coding-task parallelism caveat)
- Mode 5 (sub-agent): SELECTED
- Mode 6 (workflow): not selected — semantic/new-code/multi-rule work (decomposition + preview TUI + confirm.go v2 promotion), NOT mechanical-uniform transform

Decision: sub-agent (Mode 5)

Justification: Tier L coding-heavy refactoring + new TUI display layer with hard sequential dependencies — M3b characterization safety net MUST land before M3d decomposition, and M3a classification contract gates M3c/M3e. Per Anthropic's coding-task parallelism caveat ("most coding tasks involve fewer truly parallelizable tasks than research, and LLM agents are not yet great at coordinating in real time"), sequential sub-agent delegation per milestone is the safe default. The worst-case risk (namespace-protection regression = user-asset loss) demands the characterization-first safety net that only sequential DDD provides, not parallel fan-out. Methodology routing within the single manager-develop delegation: M3b/M3d = DDD (ANALYZE-PRESERVE-IMPROVE), M3c/M3e = TDD (RED-GREEN-REFACTOR), M3a/M3f = planning/gate. Route A (Hybrid Trunk main-direct) per CLAUDE.local.md §23 1-person OSS policy — direct-to-main push with CI 4-check protection; manager-develop performs commit + push directly.

Implementation Kickoff Approval: PASSED — user selected "run-phase 진입" 2026-07-14 (plan-audit 0.92 PASS, preconditions 4/4 verified).
