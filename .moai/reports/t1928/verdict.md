# Verdict — card t1928 (residual --abbrev-ref sweep)

verdict: PASS
audited_sha: 34b2d134ff

The lane's evidence-bearing completion verdict (class B card, no SPEC link — the factory
delivery reads this file). Written by lane-6 on 2026-10-22.

## Claim

1. All enumerated residual `--abbrev-ref HEAD` sites were disposed: the two gate-input reads
   (internal/cli/integration.go currentBranch, internal/homestate/card_pr_states.go) are FIXED
   with the symbolic-ref pattern (commit 34b2d134ff); the remaining sites carry recorded
   no-vulnerability rationales (hook_install.go:87 display-only shell string with `unknown`
   fallback; gitio.go:188 upstream query — not HEAD identification; session_worktree.go:768/870
   and session_worktree_automerge.go:311 — branch display/lookup contexts where the caller
   tolerates any value; per the implementer's per-site report and the lane's reads).
2. The reproduction tests for the two fixed sites went RED (45eb364faa) and are GREEN after the
   fix (34b2d134ff).
3. The card's diff touches only the enumerated files + their tests (scope clean vs the absorbed
   main at 7d7a06a825).

## Evidence

- RED (verbatim, delegate run): `45eb364faa test(card t1928): RED — same-name-tag reproductions
  for the two gate-input branch reads` — reproductions for integration.go currentBranch and
  card_pr_states.go.
- GREEN fix: `34b2d134ff fix(card t1928): the two gate-input branch reads are symbolic-ref with
  refs/heads/ stripped`.
- Affected-area slice (delegate run): integration/acquire/base-freshness/merge-ready/
  settings-drift families → pass 257s; gofmt clean, vet exit 0.
- Regression families (delegate run): TestResolveCardBranch family ok 1.753s; TestMergeStep ok
  40.143s (t1920-era selectors still green on this tree).
- Full internal/cli package: THREE attempts (10m default, 25m, 25m at recovered load) each
  ended in the go-test alarm (601s / 1502.4s / 1501.7s) with ZERO `--- FAIL` assertion lines —
  see Gaps.

## Baseline-attribution

Tree /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1928 at 34b2d134ff (clean), branch base
28b727065b + origin/main absorbed at fc1c4349fa→7d7a06a825 lineage. The lane re-verified the
commit subjects, markers, fix lines, and clean tree on disk this run.

## Gaps

- **Full internal/cli package verdict: NOT observed.** Three attempts each hit the go-test
  alarm (10m default twice under load 400+; 25m once at recovered load ~31) with zero assertion
  failures before the alarm — the package's full duration on this machine exceeds 25 minutes
  under any observed condition. The affected-area slice (the families the fix can affect),
  the two reproductions, and the scoped regression families are all green; the repository-wide
  verdict is CI's on the PR and is PENDING. Recorded as a measurement-environment gap, not a
  code defect.
- One golangci-lint invocation was refused by a package-load deadline under load 475 (the
  stopped delegate's record); lint was measured clean earlier at lower load.

## Residual-risk

- The non-vulnerable rationales for session_worktree/automerge/gitio/hook_install sites rest on
  caller-context reads (the delegate's + the lane's); a future caller change could alter them —
  the sites are now catalogued here for any sweep follow-up.
- The two fixed reads change rendered branch values only in the same-named-tag pathology; other
  callers of the fixed functions see identical behavior for healthy trees (regression-pinned).

AUDIT-VERDICT: PASS spec=card-t1928 receipts=lane-6-direct
