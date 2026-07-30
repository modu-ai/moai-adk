# SPEC-CLIFIX-LINTER-STALE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- 2026-07-10: plan-phase artifact set authored (spec.md / plan.md / acceptance.md) by manager-spec from CLI audit 2026-07-10 §3 clusters 4/5 + §5 P3. Status: draft. Pending plan-audit. Sequenced after SPEC-CLIFIX-CONCURRENCY-001.

## §E.2 Run-phase Evidence

Run-phase implements M1-M4 (REQ-001/002/003/005/006/007) via TDD. All four
milestones landed green; AC matrix below.

### AC matrix (acceptance.md §D, 6 ACs after REQ-004 drop)

| AC | REQ | Status | Verification command + observed outcome |
|---|---|---|---|
| AC-LINT-001-001 | REQ-001 | PASS | `go test ./internal/cli/agentlint/ -run 'LR04\|DeadHook\|ParseFrontmatter' -count=1 -v` → PASS (TestParseFrontmatter_PopulatesHooksSkills, TestParseFrontmatter_SandboxScalarResilience, TestLR04_DeadHookFiringAndNonFiring, TestCheckDeadHooks, TestCheckDeadHooks_ViaLintAgentFile all PASS). Source grep: `agent_lint.go:19` `"gopkg.in/yaml.v3"`, `agent_lint.go:261` `return yaml.Unmarshal(data, v)`. Hooks/Skills/Sandbox populated. |
| AC-LINT-001-002 | REQ-002 | PASS | `go test ./internal/cli/agentlint/ -run 'WriteHeavyAgents\|LR05' -count=1` → ok. Negated grep on the writeHeavyAgents slice: the 4 archived names are absent from the slice at `agent_lint.go` writeHeavyAgents literal (drift guard TestWriteHeavyAgents_NoArchivedNames enforces). manager-develop fires LR-05 without worktree; expert-backend fixture does NOT. Residual: `researcher` survives in the pre-existing `canonicalEffortMatrix` history comment at line 620 (PRESERVE'd docblock, out of run-phase scope) — see §E.3 Gaps. |
| AC-LINT-001-003 | REQ-003 | PASS | `go test ./internal/cli/agentlint/ -run 'LR07\|DuplicateMandate\|LiveMirror' -count=1 -v` → TestLR07_LiveMirrorPairNoFinding PASS (0 findings), TestLR07_GenuineDuplicateStillFires PASS (≥1 finding), TestCheckDuplicateMandateBlocks PASS. |
| AC-LINT-001-005 | REQ-005 | PASS | `go run ./cmd/moai --help 2>&1 \| grep -q brain` → exit 1 (brain absent). `go test ./internal/cli/ -run 'HelpRegisteredCommands\|NoPhantomBrain' -count=1` → ok. Cobra-tree gate (`registeredRootSubcommands`) suppresses unregistered rows. |
| AC-LINT-001-006 | REQ-006 | PASS | `go test ./internal/cli/taskledger/ -run 'ClaimTaskValidate\|ClaimTaskPending' -count=1 -v` → TestClaimTaskValidate_NonexistentAndCompleted PASS (nonexistent + completed → error, ledger byte-size unchanged), TestClaimTaskValidate_PendingSucceeds PASS (pending → CLAIMED row appended). |
| AC-LINT-001-007 | REQ-007 | PASS | `go test ./internal/cli/agentlint/... ./internal/cli/taskledger/ -run 'LR04\|WriteHeavyAgents\|LR07\|ClaimTaskValidate' -count=1` → both packages ok. Each of the 4 previously-dead checks has a firing + non-firing case (see linter_stale_test.go + taskledger_test.go). |

### LR-07 baseline falsifiability (acceptance.md §D.5 / D8)

Baseline `moai agent lint --format=json` (pre-fix, ce1815b5f worktree HEAD):
`total: 24, {LR-05: 14, LR-08: 10}` — LR-07 count = 0 (no agent file currently
carries a Skeptical-Evaluator Mandate block, so the structural false-positive
is not observable on the live repo). Post-fix: LR-07 count remains 0. The
falsifiability delta is therefore demonstrated by the test fixtures
(TestLR07_LiveMirrorPairNoFinding: live/mirror pair → 0; TestLR07_GenuineDuplicateStillFires:
two distinct live same-name files → ≥1), not by the live repo.

### Coverage (scope packages)

- `internal/cli/agentlint`: 86.7% of statements (≥ 85% threshold).
- `internal/cli/taskledger`: 92.7% of statements.

### Build

- `go build ./...` → exit 0.
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (cross-platform OK).

### Lint

- `golangci-lint run --timeout=3m ./internal/cli/agentlint/... ./internal/cli/taskledger/...` → 0 issues. No NEW findings vs baseline.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-29
run_commit_sha: pending-backfill
run_status: implemented
ac_pass_count: 6
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: n/a (worktree, single session)
l44_post_push_fetch: pending-push
new_warnings_or_lints_introduced: 0
cross_platform_build:
  linux_darwin: pass
  windows_amd64: pass
total_run_phase_files: 5  # agent_lint.go, agent_lint_test.go, linter_stale_test.go (agentlint); taskledger.go, taskledger_test.go (taskledger); help.go, help_linter_stale_test.go (cli)
m1_to_mN_commit_strategy: per-milestone commits on feat/SPEC-CLIFIX-LINTER-STALE-001
```

### Gaps

- AC-LINT-001-002 file-wide neg-grep: `researcher` survives in the pre-existing
  `canonicalEffortMatrix` history comment at `agent_lint.go:620`. This comment is
  part of the PRESERVE'd canonicalEffortMatrix docblock (out of run-phase scope);
  the writeHeavyAgents slice itself is clean of all 4 archived names, which is
  the AC's actual target.
- Boundary grep (E4) returns pre-existing matches in `internal/cli/agentlint/`:
  these are the linter's own detection strings (`checkLiteralAskUserQuestion`
  scans agent files FOR the literal "AskUserQuestion" — it does not invoke the
  tool). Zero NEW AskUserQuestion references introduced by this SPEC
  (`git diff ce1815b5f -- ... | grep '^+' | grep -c AskUserQuestion` = 0).

### Residual-risk

- The yaml.v3 parser is stricter than the former hand-rolled parser; no NEW
  true findings surfaced on the live `.claude/agents/moai/` + template mirror
  scan (baseline 24 findings unchanged in rule distribution). A future agent
  with malformed frontmatter could now surface as a genuine parse error
  (triage per plan §G, not suppressed).
- LR-07 dedupe keys on path-mapping (base filename + live/mirror tree
  membership); a sanitized (non-byte-identical) mirror is still paired by path,
  so the content-hash fallback is not exercised by current fixtures.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
