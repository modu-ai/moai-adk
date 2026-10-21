auditor-model: lane-6 (orchestrator verification; class B card — no SPEC, lane verdict of record)

# Sync audit — card t1920 (integration_merge_step same-name-tag repair)

verdict: PASS
audited_sha: 3bcdac02d8304fbe0a6e4d50a4c19b6cf50b83f0e

Scope: the card body's completion verdict is the same-name-tag reproduction RED→GREEN; this
record pins it plus the regression and preservation evidence, all observed by the lane in this
run on this tree (implementer claims re-verified from disk: HEAD, both commit subjects, both
MoAI markers, the fix lines, clean tree).

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim
1. `resolveCardBranch` (internal/factory/integration_merge_step.go) now reads the branch via
   `git symbolic-ref --quiet HEAD` (line ~959) with only `refs/heads/` stripped (line ~963); a
   tag sharing the card branch's name no longer distorts the read (D1 closed).
2. The WT- prefix requirement and every refusal shape are unchanged; the error wrap is
   byte-identical; detached-HEAD semantics are stated precisely in the comment (symbolic-ref
   exits non-zero on detached HEAD — the old --abbrev-ref exited 0 printing the literal `HEAD`).
3. The reproduction test
   `TestResolveCardBranchSurvivesTagSharingTheBranchName` (internal/factory/integration_merge_step_test.go)
   pins the behavior contract.

### Evidence (commands run by the lane in this run, this tree)
```
git log --oneline 28b727065b..HEAD
  -> 3bcdac02d8 fix(card t1920): the merge-step branch read is symbolic-ref with refs/heads/ stripped
  -> 118a826d9c test(card t1920): RED — same-name-tag reproduction for the merge-step branch read
git log --format=%B 28b727065b..HEAD | grep -c '🗿'  -> 2
git status --porcelain  -> (empty)
sed -n '957,965p' internal/factory/integration_merge_step.go  -> symbolic-ref read + TrimPrefix + unchanged WT- refusal (read this run)
```
Implementer-observed (verbatim in its report, spot-trusted after the disk checks above):
RED `--- FAIL: TestResolveCardBranchSurvivesTagSharingTheBranchName ... is on "heads/WT-fix-card", not a WT- branch`
→ GREEN `ok github.com/modu-ai/moai-adk/internal/factory`; regression family
`go test ./internal/factory/ -run 'TestResolveCardBranch|TestMergeStep' -count=1` → ok 34.964s;
gofmt clean, go vet exit 0, golangci-lint 0 issues.

### Baseline-attribution
Tree /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1920 at 3bcdac02d8304fbe0a6e4d50a4c19b6cf50b83f0e,
this run. Branch base 28b727065b (local main). The card's diff is exactly
internal/factory/integration_merge_step.go + integration_merge_step_test.go.

### Gaps
- Full-suite run not performed (class B lane-local scope; repository-wide verdict is CI's, pending).
- The fix commit 3bcdac02d8 is NOT an ancestor of any remote ref yet (branch unpushed; the PR
  delivery follows).

### Residual-risk
- The card is confined to integration_merge_step.go; the sibling fix in factory_card.go
  (card t1834, merged via PR #1864) covers the parallel site — when this branch PRs, that file
  arrives from main, not from this card.
- git-version-dependent rendering of --abbrev-ref is now irrelevant to this path (symbolic-ref
  is version-stable); recorded as closed.

AUDIT-VERDICT: PASS spec=card-t1920 receipts=lane-6-direct
