auditor-model: glm-5.3-flash

# Independent sync audit — card t1928 (residual --abbrev-ref sweep)

verdict: FAIL
audited_sha: 7e27ce5aa7d1e6c52fe7da1fd1a493bb0fcaa65e

Independent audit by the sync-auditor subagent, 2026-10-23. The lane's own verdict
(.moai/reports/t1928/verdict.md, auditor-model: lane-6, PASS@34b2d134ff) is context, not
authority; every claim below rests on this session's own reads and runs.

## Convergence result (audit_multi, machine digest)

- overall_verdict: fail
- gate_unmet: (empty — the required gate answered; its verdict is fail)
- plan_source: config
- per_backend_verdicts: codex / gate required / verdict fail (claude off, glm off per the tree's plan)
- receipts: rcpt-990c49b70a8796df7ba249d0
- This verdict is NOT final until the orchestrator writes the digest to
  `.moai/state/audit-plan-result.json` and `moai verify audit-plan --result-file` passes.

Tool provenance (§2.2): judging build moai 8249bff8c3 (built 2026-10-19) — an ANCESTOR of the
audited tree HEAD 7e27ce5aa7 (stated per the build_lag notice; the audit backend reads the git
diff, not the binary's own compiled code). Tree HEAD measured: 7e27ce5aa7, branch WT-abbrev-ref-card,
worktree clean at audit start (git status empty, observed).

## Claim

1. The two fixed gate-input reads are correct in place, and the reproductions are non-vacuous
   and green — CONFIRMED for both sites.
2. The no-vulnerability rationales for the five enumerated remaining sites — CONFIRMED by
   caller-trace reads for all five.
3. The lane's enumeration is complete for live `--abbrev-ref` sites — REFUTED: at least one
   live Go site (internal/worktree/state_guard.go:107) is absent from the "five remaining live
   sites" claim (not-vulnerable on my read, but unenumerated).
4. The card closes the merge-ready path's exposure to the same-name-tag pathology — REFUTED
   (finding F1, corroborating the codex required-gate FAIL): the fixed read hands the triple a
   clean bare branch name, whose rev operations resolve tags-before-heads; the gate bypass moves
   from mislabel to misresolve.
5. Scope and absorb hygiene — CONFIRMED: 7d7a06a825..HEAD is exactly the 8 enumerated files
   (2 fix + 2 test + 3 t1928 reports); the absorb was a rebase onto origin/main 7d7a06a825
   (45eb364faa's sole parent), so nothing beyond origin/main entered the branch by construction.

## Evidence

Command → verbatim output, all observed this run in this tree:

- `git rev-parse HEAD; git branch --show-current; git status --short` →
  `7e27ce5aa7d1e6c52fe7da1fd1a493bb0fcaa65e` / `WT-abbrev-ref-card` / (empty).
- `git diff 7d7a06a825..HEAD --name-only` → exactly 8 files: card-review.md, sync-audit.md,
  verdict.md, internal/cli/integration.go, internal/cli/integration_target_test.go,
  internal/homestate/card_pr_states.go, internal/homestate/card_pr_states_test.go (plus nothing
  else). Parent check: `git log --format='%H %P' -1 45eb364faa` →
  `45eb364faa4e… 7d7a06a82585…` (single parent — rebase absorb, no merge commit).
- `go test ./internal/cli/ -run 'TestCurrentBranchSurvivesTagSharingTheBranchName' -count=1 -v` →
  `--- PASS: TestCurrentBranchSurvivesTagSharingTheBranchName (0.09s)` / `ok … internal/cli 0.933s`.
- `go test ./internal/homestate/ -run 'TestPROpenEdgeSurvivesTagSharingTheCardBranchName' -count=1 -v` →
  `--- PASS: TestPROpenEdgeSurvivesTagSharingTheCardBranchName (0.50s)` / `ok … internal/homestate 1.018s`.
- `go test ./internal/homestate/ -count=1` → `ok … 87.535s`.
- `go test ./internal/factory/ -run 'TestResolveCardBranch' -count=1 -v` →
  `--- PASS: TestResolveCardBranchRefusalsNameTheCard` (3 subtests) +
  `--- PASS: TestResolveCardBranchSurvivesTagSharingTheBranchName (0.21s)` / `ok … 1.173s`.
- `go test ./internal/factory/ -run 'TestMergeStep' -count=1` → `ok … 35.298s`.
- `go test ./internal/cli/ -run 'TestIntegration' -count=1` → `ok … 36.457s`.
- `go vet ./internal/cli/ ./internal/homestate/` → exit 0.
- `golangci-lint run internal/cli/... internal/homestate/...` → `0 issues.` exit 0.
- `gofmt -l` on the four touched files → empty, exit 0.
- Non-vacuity reads: both RED test bodies create a tag on the same commit as the checked-out
  branch and assert the plain branch name; the homestate fixture pushes the card branch before
  tagging, so the old composition `refs/remotes/origin/heads/WT-pr-card` cannot exist and the old
  read refuses — the assertion is load-bearing, not vacuous.
- Caller-tolerance reads: factory_merge.go:85 (triple input), integration.go:158-173 (window
  record, "honest unknown" doctrine in the function's own comment), integration_settings_drift.go:219/279
  (report text) — all tolerate `""` on detached HEAD; healthy-tree behavior unchanged.
- Rationale confirmations (reads): hook_install.go:87 → value feeds only the `printf >> log` line;
  session_worktree.go:768 and :870 → value compared only to `""`/`"HEAD"` (detached sentinels), the
  verdicts use `@{u}`/`--remotes`/`merge-base HEAD <ref>` directly; session_worktree_automerge.go:311 →
  value used as a rev operand (`develop..heads/<b>` resolves to refs/heads/<b> via gitrevisions
  rule 2, which precedes tags) and in a notice string; gitio.go:188 → `rev-parse --abbrev-ref
  branch@{upstream}` identifies the upstream, not HEAD (defect class is HEAD-specific).
- CI attribution: `gh pr checks 1892` → Test (ubuntu-latest) fail, Race Test ×3 fail,
  graph-freshness fail; Lint/Build (all 5 platforms)/Analyze-Go/CodeQL pass. `gh api …/commits/
  7d7a06a825…/check-runs` on the MAIN PARENT → the SAME four failures (Test ubuntu-latest,
  Race remainder, Race cli 1/4, Race cli 4/4). The repo-wide test red is pre-existing on
  origin/main and NOT attributable to this card's diff; the failing tests
  (TestFactoryNext*, TestCatalogHash*, TestBubblewrap_ExecHello, TestCodex*, TestTodoSelection…)
  all live outside the card's four touched files.

## Findings

- F1 [High] [blocking] internal/cli/factory_merge.go:85 → internal/factorylane/merge.go:290, 308,
  313, 214 — after the fix, `factory merge ready`'s currentBranch() fallback feeds the clean bare
  branch name into EvaluateMergeTriple, whose rev operations consume the bare name:
  `merge-tree --write-tree <develop> <branch>` (b and c) and `rev-parse <branch>^{tree}` (c and d).
  Git resolves a bare name tags-before-heads (gitrevisions rule 3 over rule 4), so a tag sharing
  the card branch's name makes conditions (b)/(c)/(d) probe the TAG's commits and trees — an
  unready merge (real branch conflicted, no re-measure record for its tree) can read all_passed
  on the tag's state. The card's fix cured the recorded-string distortion and the homestate
  composition, but on the merge-ready path the same pathology now defeats the probe rather than
  the label. Premise verified by this session's code-read of all four consumption sites; the
  codex backend independently reproduced allPassed=true (its finding P1). Note: pre-fix,
  `heads/<branch>` as a rev operand resolved via rule 2 to the actual branch — the probes were
  correct while the label was wrong; the fix inverted that. Required fix: the triple's git probes
  must consume `refs/heads/<branch>` (or the SHA resolved from it) while records keep the display
  name; add a triple-level same-name-tag regression test.
- F2 [Low] [optional] The lane verdict's "five remaining live abbrev-ref sites" enumeration is
  incomplete repo-wide: internal/worktree/state_guard.go:107 (live Go product code) is absent.
  My read: the value is stored in a Snapshot, compared symmetrically pre/post for BranchChanged,
  and displayed in reports — a mid-window same-name tag can only add a false-positive divergence
  alarm (fail-safe direction; HeadChanged is a redundant signal), so not-vulnerable — but it must
  be in the catalogue. Also outside the count: scripts/release.sh:142 (dev script; distorted value
  dies on `!= main` — fail-closed) and the template mirrors (templates/.git_hooks/pre-push:46 —
  same log-only content as the enumerated hook_install.go:87; worktree-state-guard.md:46 — doc).
  Required fix: none for safety; correct the enumeration in the card record or a follow-up sweep.
- F3 [Low] [optional] The lane verdict's Evidence cites "TestResolveCardBranch family ok 1.753s"
  against internal/cli — that selector matches ZERO tests in that package (`go test ./internal/cli/
  -run 'TestResolveCardBranch' -count=1` → `ok … 0.800s [no tests to run]`, observed this run);
  the family lives in internal/factory, where I re-ran it green. An empty-sweep green was cited
  as a regression-family pass (verification-completeness §1.1). Required fix: cite regression
  families from the package that owns their tests; treat `[no tests to run]` as a Gap, never a pass.

## Baseline-attribution

Every claim above was measured in THIS run against THIS tree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1928
at 7e27ce5aa7 (clean), branch base 28b727065b, main parent 7d7a06a825 (origin/main tip). Go entry
conditions observed before each Go command: `moai slot status --resource go-test-heavy` → free;
1-min load 9.87 (≤ 32); free memory 92% (≥ 50%). One Go invocation at a time throughout.

## Gaps

- F1 has no live reproduction by this session: creating a same-name tag mutates the repo's refs
  (prohibited: edit nothing, commit nothing), and the worktree guard twice refused scratch-repo
  git outside this tree (first the `rm -rf`-prefixed compound form, then the out-of-worktree git
  form — refusals named, not retried). F1 therefore rests on my code-read of the four consumption
  sites + documented gitrevisions resolution + codex's independent reproduction — not on an
  observed run of the bypass.
- The old-read distortion demonstration (abbrev-ref printing heads/<branch>) was not re-observed
  by me for the same guard reason; non-vacuity of the reproductions rests on the test bodies
  (read), my GREEN runs (observed), and the lane's RED record (attributed, not re-observed).
- The full internal/cli package was NOT re-run (three prior >25-minute alarms on this machine;
  the dispatch forbids re-attempt). The repository-wide test verdict is CI's and is currently RED
  on BOTH this PR and origin/main (pre-existing, attribution measured above) — the card cannot
  claim a green repo-wide verdict, and neither can main.
- The dispatch's output template says `spec=card-t1920`; the card audited is t1928. The verdict
  line below names card-t1928 (the audited card); the discrepancy is the dispatch's, flagged here.

## Residual-risk

- audit_multi residual_risk_note (verbatim): "required-backend FAIL: codex". disagreement_flag:
  null (single participant — codex was the only non-off gate; no cross-backend comparison was
  possible, so neither agreement nor disagreement is claimed).
- If the orchestrator's convergence check cannot corroborate the receipt, this verdict does not
  stand as the required-gate PASS/FAIL record.
- F1's exploitability requires a tag named exactly like a card branch to exist in the repo; this
  repo's tags are version-shaped today, so the path is latent, not active. It is the SAME
  contrivance the card treats as in-scope at the homestate site, so consistency demands the
  merge-ready path close it too.
- Pre-existing main CI red (4 checks) will ride this PR's merge; consumers of "CI green" evidence
  on main should not read this PR's red as new.

AUDIT-VERDICT: FAIL spec=card-t1928 receipts=rcpt-990c49b70a8796df7ba249d0
