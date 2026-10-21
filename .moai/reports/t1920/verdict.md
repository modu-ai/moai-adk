# Verdict — card t1920 (integration_merge_step same-name-tag repair)

verdict: PASS
audited_sha: 3bcdac02d8c41eab42e286d9a48b2b3272e4368b

The lane's evidence-bearing completion verdict (class B card, no SPEC link — the factory
delivery reads this file). Written by lane-6 on 2026-10-22.

## Claim

1. `resolveCardBranch` reads the branch via `git symbolic-ref --quiet HEAD` with only
   `refs/heads/` stripped (internal/factory/integration_merge_step.go ~959/963) — a tag sharing
   the card branch's name no longer distorts the read (card t1834 verdict D1 closed at its
   remaining site).
2. The reproduction test
   `TestResolveCardBranchSurvivesTagSharingTheBranchName` went RED (the audit-observed
   `"heads/WT-fix-card", not a WT- branch` refusal) and is GREEN after the fix; the regression
   family (TestResolveCardBranch|TestMergeStep) is green; WT- refusal shapes unchanged.
3. The card's diff is exactly internal/factory/integration_merge_step.go +
   integration_merge_step_test.go (single-file scope, conflict-free with the parallel Cloud
   repair scope per the leader's measurement).

## Evidence

- RED (verbatim, implementer run, lane re-verified commit subjects + fix lines + clean tree on
  disk): `--- FAIL: TestResolveCardBranchSurvivesTagSharingTheBranchName (0.26s) ... is on
  "heads/WT-fix-card", not a WT- branch` → GREEN `ok ... internal/factory 0.547s`.
- Regression: `go test ./internal/factory/ -run 'TestResolveCardBranch|TestMergeStep'
  -count=1` → ok 34.964s (detached-HEAD / missing-dir / non-WT arms all PASS).
- Commits: 118a826d9c (RED) → 3bcdac02d8 (fix); both end with the MoAI marker; card t1920 in
  both subjects.
- Lint: gofmt clean, go vet exit 0, golangci-lint 0 issues (this tree, this run).

## Baseline-attribution

Tree /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1920 at
3bcdac02d8c41eab42e286d9a48b2b3272e4368b (clean), branch base 28b727065b. Full detail:
sync-audit.md beside this file.

## Gaps

- Full-suite run not performed (class B lane-local scope); repository-wide verdict is CI's,
  pending.

## Residual-risk

- The parallel factory_card.go site is owned by the merged t1834 change (PR #1864); this card
  deliberately does not touch that file (Cloud repair scope).
- Older git versions may render --abbrev-ref differently, but this path no longer reads it.

AUDIT-VERDICT: PASS spec=card-t1920 receipts=lane-6-direct
