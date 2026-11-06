auditor-model: claude-sonnet-5-5

# Sync audit iter6 (HEAD confirmation DELTA) — card t1834, SPEC-INTEGRATION-STALE-BASE-001

verdict: PASS
audited_sha: a63659551504a6f24efce731a484e541f91a48b1

Overall: PASS. The card's own product diff is the audited set of 15 non-.moai files, the only change to a card-owned file since the audited commit is a 4/4-line test-helper rename, and the codex gate (required) returned pass.

## Claim

1. HEAD is a63659551504a6f24efce731a484e541f91a48b1, the tree is clean, and the branch is 0 first-parent commits behind main.
2. The card's non-.moai diff against main is exactly the audited 15-file set; no file of main's landed work is altered by the card.
3. Against the audited commit ae092df3, no card-owned product file changed. The only card-owned change is the helper rename in internal/factory/base_freshness_test.go (4+/4-).
4. The base-freshness gate call sites and the config key are intact.
5. The required codex gate returned pass.

## Evidence

### E1 HEAD / status / behind-count
```
git rev-parse HEAD                      -> a63659551504a6f24efce731a484e541f91a48b1
git status --short                      -> (no output)
git rev-list --count --first-parent HEAD..main -> 0
git rev-parse --show-toplevel           -> /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1834
```

### E2 Card product diff vs main (`git diff --name-status main...HEAD -- . ':!.moai'`)
```
M	.claude/rules/moai/workflow/factory-dispatch-gates.md
M	CHANGELOG.md
M	internal/cli/factory_card.go
M	internal/cli/integration.go
A	internal/cli/integration_base_freshness.go
A	internal/cli/integration_base_freshness_test.go
M	internal/config/cache.go
M	internal/config/defaults.go
M	internal/config/testdata/shipped_key_inventory.yaml
M	internal/config/types.go
A	internal/config/workflow_base_freshness_gate_test.go
A	internal/factory/base_freshness.go
A	internal/factory/base_freshness_test.go
M	internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-gates.md
M	internal/template/templates/.moai/config/sections/workflow.yaml
```
Stat tail: `15 files changed, 2187 insertions(+), 13 deletions(-)`. The 15 paths match the audited set; no surprising path. Main's landed work (for example queue_location_test.go) is absent from the list, so it is not altered.

### E3 Delta against audited commit ae092df3
- `git diff ae092df3 HEAD -- internal/factory/base_freshness.go internal/cli/integration_base_freshness.go` -> empty (both product files byte-identical).
- `git diff ae092df3 HEAD --stat -- internal/cli/integration_base_freshness_test.go internal/config/workflow_base_freshness_gate_test.go internal/cli/integration.go internal/cli/factory_card.go internal/config/types.go internal/config/defaults.go`:
```
 internal/cli/factory_card.go | 6 +++---
 internal/config/defaults.go  | 7 +++++++
 2 files changed, 10 insertions(+), 3 deletions(-)
```
  integration.go, types.go, integration_base_freshness_test.go and workflow_base_freshness_gate_test.go are unchanged since ae092df3.
- The two changed shared files contain only main's other-card hunks: `Args: cobra.RangeArgs(1, 2)` and the `--to`/`--reassign` help text in factory_card.go, and `DefaultLeaderAnchorAuditRelPath` in defaults.go. No base-freshness hunk.
- `git diff ae092df3 HEAD -- internal/factory/base_freshness_test.go` -> the rename only: `snapshotTree` -> `baseFreshnessSnapshotTree` at 3 call sites plus the definition and its comment (4 insertions, 4 deletions).
- `grep -rn snapshotTree internal/factory` leaves only `internal/factory/queue_location_test.go` (main's helper, lines 19/108/121); the card's helper no longer shares the name, so the redeclaration clash is gone.

### E4 Gate call sites and key intact
```
internal/cli/integration.go:473:  if freshErr := acquireBaseFreshnessPrecondition(cmd, root, cardFlag, branch); freshErr != nil {
internal/cli/factory_card.go:2202: if freshErr := completeBaseFreshnessPrecondition(out, lockRoot, integTree, branch, card.WorktreePath, cardBranch, card.CardID); freshErr != nil {
internal/config/types.go:857:  MaxBehindLandings int  `yaml:"max_behind_landings"`
internal/config/defaults.go:1565: MaxBehindLandings: 0,
internal/config/testdata/shipped_key_inventory.yaml:2472: - path: "workflow.base_freshness_gate.max_behind_landings"
internal/template/templates/.moai/config/sections/workflow.yaml:262:        max_behind_landings: 0
```
EvaluateBaseFreshness and checkBaseFreshness live in the unchanged-since-ae092df3 product files (E3 empty diff).

### E5 Audit plan and gate
`moai verify audit-plan --project-root <tree>`: config_status ok; model codex; claude off, codex required (explicit), glm off; cross_model_active true; enforced_required ["codex"].

`mcp__moai__audit_multi` (project_root = tree, card_id = t1834, target = baseBranch, no gates argument), verbatim digest:
```
overall_verdict: pass
gate_unmet: (absent)
plan_source: config
per_backend_verdicts: [{backend: codex, gate: required, verdict: pass, summary: "Verdict: pass", findings: [], next_steps: []}]
disagreement_flag: null
participant_count: 1
fail_open_backends: []
audit_receipt: rcpt-d93ba0455b5be353b0cc3509
build_commit: ee981a932
```
This verdict is not final until the orchestrator's `moai verify audit-plan --result-file` convergence check passes.

## Baseline-attribution

- Tree measured: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1834, HEAD a63659551504a6f24efce731a484e541f91a48b1, in this run.
- Prior audited baseline: ae092df3adf82decccaa2e47a35a25ad9f963426 (round 4), compared by git diff in this run.
- Judging build: the codex gate ran in moai MCP server build_commit ee981a932 (built 2026-10-15T21:56:00Z).

## Gaps

- The build_lag field says the judging binary (ee981a932) is a strict ancestor of the tree HEAD a63659551, so the gate and `moai verify` outputs describe older code. The audited diff is read through git, which is unaffected, but the codex result itself was produced by that older build. No rebuilt binary was used.
- No tests were run, per instruction (no go build, no package or suite runs, no Windows build). The rename compiles only by inspection: no remaining `snapshotTree` in the card's test file, and main's helper keeps its own name. Compile and test status of internal/factory at HEAD is not observed by this audit.
- The 4/4-line rename diff and the other-card hunks were judged by reading the diff, not by executing code.
- AC-ISB-001..016 were not re-executed individually; their regression status rests on the empty product-file diff against the round-4 PASS commit and the codex pass.
- The convergence check (`--result-file`) is the orchestrator's step; not run here.

## Residual-risk

- A merge-time semantic clash between main's later landings and the card's gate that neither the diff nor the codex pass would show; only the lane's slot run (build and tests) can retire it.
- The older judging build could differ from the current tree in codex-gate handling.

## Iteration history

| Round | Receipt | Verdict | Audited commit |
|---|---|---|---|
| 1 | rcpt-c9bb6784d6ebb463b124e1d4 | FAIL | not re-read this round |
| 2 | rcpt-60312ee19c3179eea8af7e44 | FAIL | not re-read this round |
| 3 | rcpt-56f80c4d7a80842575127f29 | FAIL | not re-read this round |
| 4 | rcpt-8fdd95a2ca747abf709b8914 | PASS | ae092df3adf82decccaa2e47a35a25ad9f963426 |
| 5 (procedural delta) | rcpt-0a5d99b910ae00b2ddd53b41 | PASS | 30dff01635b992e760d875b31a61cd9f97394eed |
| 6 (this, HEAD confirmation) | rcpt-d93ba0455b5be353b0cc3509 | PASS | a63659551504a6f24efce731a484e541f91a48b1 |
