auditor-model: lane-6 (orchestrator verification; class B card — no SPEC, lane verdict of record)

# Sync audit — card t1928 (residual --abbrev-ref sweep)

verdict: PASS
audited_sha: 34b2d134ff713fa8fef769e80c6295fdb6f7197d

Scope: the card body's completion verdict — enumeration disposition for every residual
--abbrev-ref site + regression on the applied fixes. Full detail in verdict.md (same tree).

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim
1. Two gate-input reads FIXED (internal/cli/integration.go currentBranch,
   internal/homestate/card_pr_states.go) with the symbolic-ref pattern (commit 34b2d134ff);
   reproductions RED 45eb364faa → GREEN.
2. Remaining enumerated sites carry recorded no-vulnerability rationales
   (hook_install.go:87 display-only; gitio.go:188 upstream query; session_worktree.go:768/870,
   session_worktree_automerge.go:311 display/lookup contexts).
3. Card diff confined to the enumerated files + tests (scope clean vs 7d7a06a825).

### Evidence (lane-observed this run, this tree)
```
git log --oneline 7d7a06a825..HEAD
  -> 45eb364faa test(card t1928): RED — same-name-tag reproductions for the two gate-input branch reads
  -> 34b2d134ff fix(card t1928): the two gate-input branch reads are symbolic-ref with refs/heads/ stripped
git status --porcelain -> (empty)
```
Delegate-observed (verbatim in report): affected-area slice pass 257s; TestResolveCardBranch ok
1.753s; TestMergeStep ok 40.143s; gofmt clean; vet exit 0.

### Baseline-attribution
Tree /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1928 at 34b2d134ff713fa8fef769e80c6295fdb6f7197d
(clean) — the audited content; the lane re-verified commit subjects, markers, and the clean tree
on disk this run.

### Gaps
- Full internal/cli package: three attempts (10m, 25m, 25m-at-recovered-load) each hit the
  go-test alarm with ZERO assertion failures — the package exceeds 25m on this machine under
  any observed condition. Affected-area slice green; repository-wide verdict is CI's, pending.

### Residual-risk
- Non-vulnerable rationales rest on caller-context reads; a future caller change could alter
  them — sites catalogued in verdict.md for any sweep follow-up.

AUDIT-VERDICT: PASS spec=card-t1928 receipts=lane-6-direct
