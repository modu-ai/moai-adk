auditor-model: glm-5.3-flash

# Independent sync audit — card t1928 (residual --abbrev-ref sweep)

verdict: PASS-WITH-DEBT
audited_sha: c4b5a878dd0ed21fed26422a8a00a918235b3f22

Independent audit by the sync-auditor subagent, 2026-10-23. The lane's own verdict
(.moai/reports/t1928/verdict.md, auditor-model: lane-6, PASS@34b2d134ff) is context, not
authority; every claim below rests on this session's own reads and runs.

This file carries THREE audit rounds of one session: ROUND 1 measured 7e27ce5aa7d1e6c52fe7da1fd1a493bb0fcaa65e
(FAIL — F1 blocking, merge-triple bare-name probes tag-vulnerable); ROUND 2 (§ Delta confirmation)
measured 48c9aaee429bd71a1530da25f644a4dfcd521bf5 (FAIL — F4, the qualified ref itself
fake-tag-shadowable); ROUND 3 (§ Delta confirmation round 3) measures the machine-line tree
c4b5a878dd… (PASS-WITH-DEBT — F1/F4 repaired and verified; recorded debts F5 + sibling
composed-string residuals + a coverage note). Each round's receipt ids are cited in its own
section and in the final verdict line.

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

## Delta confirmation — ROUND 2 (F1 repair re-audit)

audited_sha (this round): 48c9aaee429bd71a1530da25f644a4dfcd521bf5 — branch WT-abbrev-ref-card,
worktree clean (observed). Delta scope vs round 1: absorb merge e77101dde5 (origin/main
a1762cabe1 = tip, verified by `git rev-parse origin/main`; `git diff a1762cabe1 e77101dde5
--name-only` → exactly the card's 8 files — the merge introduced NOTHING beyond main), RED
76a1dfa23b (130 lines, internal/factorylane/merge_test.go), fix 6f1b8ead1e (merge.go +35/−11,
merge_test.go fixture key), docs 48c9aaee42 (F3 citation internal/cli→internal/factory corrected;
F2 known-issue note added to the lane verdict — PARTIAL: release.sh:142 and the template mirrors
are recorded, `internal/worktree/state_guard.go:107` is NOT named — F2 residue stands).

### Convergence result (delta, audit_multi with delta_base=7e27ce5aa7…)

- overall_verdict: fail
- gate_unmet: (empty — the required gate answered; its verdict is fail)
- plan_source: config
- per_backend_verdicts: codex / gate required / verdict fail (claude off, glm off)
- receipts: rcpt-a90446f0ea5ae36e7277fdf3
- residual_risk_note (verbatim): "delta_base 7e27ce5aa7d1e6c52fe7da1fd1a493bb0fcaa65e: no claude
  or glm leg produced a verdict on it; codex was scoped in prose only (not mechanically bounded);
  no second-review record and no session state written | required-backend FAIL: codex"

### Claim (delta)

1. The F1 repair qualifies the card branch at ALL FOUR probe sites with records keeping the
   display name — CONFIRMED by read of 6f1b8ead1e: `cardBranchRef(branch) = "refs/heads/" +
   branch` consumed at checkConflictFree's merge-tree, checkTreeIdentityPreMerge's merge-tree and
   `rev-parse …^{tree}`, and the re-measure candidate-tree read; `MergeCheckRun.Branch` remains
   `in.Branch` (construction line untouched).
2. The two new reproductions are non-vacuous and green — CONFIRMED: both fixtures build real
   repos, shadow the branch with a same-name tag AT DEVELOP'S TIP, and carry in-fixture
   non-vacuity assertions (merge-tree on the qualified ref exits 1 = real conflict;
   `rev-parse WT-card == rev-parse develop` = the tag really shadows). The conflict test's
   `FailedCondition == CheckConflictFree` assertion discriminates old from new behavior (under
   the old code the tag-shadowed probes failed a LATER condition, not (b)); the re-measure test
   asserts `got == rev-parse refs/heads/WT-card^{tree}`. Observed:
   `go test ./internal/factorylane/ -run 'TestEvaluateMergeTriple' -count=1 -v` → both new tests
   PASS alongside all 8 pre-existing triple tests; full `./internal/factorylane/ -count=1` → ok
   10.132s.
3. Scoped families green post-absorb — OBSERVED: TestFactoryMergeReady (5 tests, internal/cli
   7.929s), TestCurrentBranchSurvivesTagSharingTheBranchName ok, TestPROpenEdgeSurvives… ok,
   TestResolveCardBranch ok 3.520s, TestMergeStep ok 96.809s; go vet exit 0; gofmt clean;
   golangci-lint `0 issues.` Entry conditions re-checked before the runs: slot free, 1-min load
   17.47 (≤ 32), free memory 93% (≥ 50%).
4. The repair closes the same-name-tag bypass with no new defect — REFUTED at the required-gate
   level: the composed `refs/heads/<branch>` operand itself resolves through gitrevisions rules,
   and rule 3 (refs/tags/<name>) precedes rule 4, so a tag literally named `refs/heads/WT-card`
   redirects the QUALIFIED probes when the real branch is absent (and can shadow a packed-only
   branch, which misses rule 1's loose-file check). Mechanism verified by this session's read of
   merge.go:295 (the composed operand) against the documented resolution order; codex
   independently reproduced all_passed=false at the round-1 base vs all_passed=true at
   48c9aaee42. The SAME composed-string shape exists in the sibling fixes
   (integration_merge_step.go:289/300, base_freshness.go integrationBranchRef) — pre-existing
   class residual, not introduced by this repair.

### Findings (delta)

- F4 [Medium] [blocking-by-gate] internal/factorylane/merge.go:295 (and the sibling composed-ref
  sites) — the qualified `refs/heads/<branch>` operand is itself tag-shadowable by a tag named
  `refs/heads/<branch>` (gitrevisions rule 3 over rules 1/4 for absent or packed-only branches).
  The codex required gate returned FAIL on this P2; the verdict token below is gate-governed.
  My severity read: the trigger requires an adversarially-named tag (`refs/heads/…` as a TAG
  name) — near-adversarial-only, downstream-caught by the merge step's own probes — which is why
  the repair's form (mirroring the siblings) remains a net improvement. Required fix (per codex,
  which I endorse): verify the branch's existence and pin its SHA at the gate (`git show-ref
  --verify --hash refs/heads/<branch>`, probes consume the SHA), plus a no-branch-fake-tag-only
  regression test; the same residual applies to the sibling sites and belongs in the same sweep.
- F5 [Low] [optional] internal/factorylane/merge.go:295, 322 — the DEVELOP operand of both
  merge-tree probes is still consumed bare (`merge-tree --write-tree <developRef> <branchRef>`);
  developRef sources from config (bare "main") or --develop, and a tag named like the integration
  target shadows it. Unlike the card side this is fail-toward-wrong-base with a stale-ancestor
  tag and is caught downstream before any landing (the merge step qualifies both sides); the
  sibling convention (integration_merge_step.go:300 qualifies the integration side) argues for
  qualifying developRef here too. Reported as discretionary; the enumerated F1 delta is complete
  without it.
- F2 residue: the docs commit's known-issue note omits state_guard.go:107 (see Claim 3 context
  above); one-line completion when next touched.

### Gaps (delta)

- F4 has no live reproduction by this session (tag creation mutates the repo — prohibited;
  out-of-worktree git refused by the guard — refusals named). It rests on the documented
  gitrevisions order + my read of the composed operand + codex's reproduction.
- The delta HEAD's own CI runs were not yet observed: `gh pr checks 1892` returns the SAME run
  ids as the pre-delta head (39785728726…), i.e. the pre-delta results; the four Test/Race fails
  there are pre-existing on the main parent (round-1 attribution stands) and #1869 (absorbed)
  claims their repair — unverified by me.
- The fix's RED observation (76a1dfa23b message: "RED observed verbatim on both") is the lane's
  record; my own GREEN runs and the tests' discriminating assertions carry the non-vacuity
  burden (round-1 scratch-repo guard refusals apply unchanged).

### Baseline-attribution (delta)

All delta claims measured in THIS run against THIS tree: worktrees/t1928 at 48c9aaee42… (clean).
Round-1 scope (7e27ce5aa7…) was audited up to and including that SHA; the delta call passed
delta_base=7e27ce5aa7… per the tool contract. Tool provenance unchanged: judging build moai
8249bff8c3 — ancestor of both round trees (stated per the build_lag notice).

### Residual-risk (delta)

- The verdict token is GOVERNED by the required codex gate's FAIL on F4; a leader disposition
  that reclassifies F4 (adversarial-only trigger) still needs a conforming gate outcome —
  re-run audit_multi after the F4 repair lands, scoped by delta_base=48c9aaee42….
- F5 (bare develop operand) and the sibling composed-string residual remain open class surface
  in base_freshness.go and integration_merge_step.go.
- Participant count stayed 1 (codex only; claude/glm off) — no cross-backend agreement claim is
  possible; disagreement_flag null in both rounds.

## Delta confirmation round 3 — F4 repair re-audit

audited_sha (this round): c4b5a878dd0ed21fed26422a8a00a918235b3f22 — branch WT-abbrev-ref-card,
tree clean except this audit file (the one sanctioned writer). Delta vs round 2: e671bb5b38 (RED
— fake-tag reproduction, 56 lines, internal/factorylane/merge_test.go) + c4b5a878dd (fix —
branchTipSHA helper; merge.go +66/−20, merge_test.go fixture keys). HEAD is the fix commit; the
branch is unpushed (remote answers 422 No commit found for the SHA), so round-3 CI runs do not
exist yet.

### Convergence result (round 3, audit_multi with delta_base=48c9aaee42…)

- overall_verdict: pass
- gate_unmet: (empty)
- plan_source: config
- per_backend_verdicts: codex / gate required / verdict pass, findings empty (claude off, glm off)
- receipts: rcpt-f49524a64c0bfad878348db5
- residual_risk_note (verbatim): "delta_base 48c9aaee429bd71a1530da25f644a4dfcd521bf5: no claude
  or glm leg produced a verdict on it; codex was scoped in prose only (not mechanically bounded);
  no second-review record and no session state written"

### Claim (round 3)

1. The SHA pin closes the rule-3 path and no residual composed-string rev consumption of the
   branch name remains in the probes — CONFIRMED by read: `branchTipSHA` resolves the tip via
   `git show-ref --verify --hash refs/heads/<branch>` (an EXACT-ref-path lookup — `--verify`
   fails unless that precise ref exists, and a tag named `refs/heads/<branch>` lives under
   `refs/tags/…`, unreachable by it); all four sites then consume the returned SHA
   (`merge-tree --write-tree <develop> <tip>` in (b) and (c); `rev-parse <tip>^{tree}` in (c)
   and (d)). A SHA is not revision-rule-resolved, so neither bare-name (F1) nor fake-tag (F4)
   shadowing applies. The branch name survives only in `show-ref`'s exact-path argument, the
   fail-closed refusal text, and display strings (`<ref>@<sha>`; `MergeCheckRun.Branch` keeps the
   display name). A grep over merge.go confirms no other branch-name operand reaches a git probe.
   An unverifiable branch is a fail-closed refusal naming the exact check (observed in the new
   test's refusal path).
2. The fake-tag reproduction is non-vacuous and discriminating — CONFIRMED by read + run:
   `TestEvaluateMergeTripleSurvivesATagNamedLikeTheQualifiedBranchRef` builds a repo with NO
   WT-card branch and a tag literally named `refs/heads/WT-card` at develop's tip; its in-fixture
   asserts prove BOTH pre-fix behavior (`rev-parse refs/heads/WT-card` RESOLVES — to the tag,
   rule 3 — equal to develop's tip) and the pin's mechanism (`show-ref --verify --hash` refuses
   the fake tag), and the triple assertion requires `FailedCondition == CheckConflictFree` — the
   refusal lands at the FIRST probe, which only the pinned implementation produces (round-2 code
   probed the fake tag clean and either passed outright or failed a later condition).
   Observed: `go test ./internal/factorylane/ -run 'TestEvaluateMergeTriple' -count=1 -v` → all
   11 triple tests PASS including the new one (0.21s); full `./internal/factorylane/ -count=1` →
   ok 15.466s.
3. Scoped families green — OBSERVED: TestFactoryMergeReady (internal/cli, ok 6.169s),
   TestMergeStep (internal/factory, ok 67.006s); vet exit 0; gofmt clean; golangci-lint
   `0 issues.` Entry conditions: slot free, 1-min load 23.95 (≤ 32), free memory 93% (≥ 50%).
4. F2 naming complete at the card-record level — CONFIRMED with a placement nuance: the round-3
   fix commit message carries the full known-issue block (state_guard.go:107 — "auditor-verified
   not-vulnerable … required in the catalogue" — plus factory_card.go, integration_merge_step.go,
   scripts/release.sh:142, and both template mirrors); the lane verdict.md's Residual-risk note
   (48c9aaee42) carries release.sh + mirrors + the F5 develop-operand record but not yet the
   three Go names. The catalogue requirement is met by the commit-message record and this file;
   the verdict.md note is one edit behind, cosmetic.

### Findings (round 3)

- None blocking. Debt record (not new this round): F5 [Low][optional] — the develop operand of
  both merge-tree probes remains bare (config-sourced "main"); recorded in the lane verdict's
  Residual-risk per the minimal-prescription scope; a same-class sibling residual is recorded for
  factory_card.go and integration_merge_step.go (composed-string operands, pre-existing on main).
- Coverage note: the packed-only-branch variant of F4 (rule 1 misses a packed branch, rule 3
  fake tag wins) is declared inexpressible in the fixture — `git pack-refs --all` could in fact
  express it — but the SHA pin closes that arm mechanically (show-ref --verify reads packed
  refs; a SHA cannot be shadowed), so this is a test-coverage note for any future revert toward
  ref-spelling probes, not a defect in the fix.

### Gaps (round 3)

- CI on the round-3 HEAD is NOT observable: the branch is unpushed (remote 422 on the SHA) and
  `gh pr checks 1892` still shows the pre-delta run ids (39785728726…). The repository-wide test
  verdict remains CI's, post-landing. The four Test/Race fails visible there are the pre-delta
  results on the main parent (round-1 attribution stands; #1869, absorbed in round 2, claims
  their repair — unverified by me).
- F4's live reproduction and the RED observation remain the lane's/codex's records; this session
  re-verifies by mechanism read + discriminating GREEN runs (tag creation in this repo stays
  prohibited; out-of-worktree git stays refused).

### Baseline-attribution (round 3)

All round-3 claims measured in THIS run against THIS tree: worktrees/t1928 at c4b5a878dd… .
Round 2 was audited up to and including 48c9aaee42…; the delta call passed
delta_base=48c9aaee42…. Tool provenance unchanged: judging build moai 8249bff8c3 — ancestor of
every round's tree (stated per the build_lag notice).

### Residual-risk (round 3)

- Verdict token PASS-WITH-DEBT: the enumerated defects (F1, F4) are repaired and verified, the
  required codex gate passes on the delta; the carried debts are exactly the recorded ones — F5
  (bare develop operand), the sibling composed-string sites, and the packed-arm coverage note.
- A leader disposition that wants F5 or the sibling residuals fixed lands them as follow-up
  cards; nothing in this verdict blocks the card's merge-readiness beyond the factory's own
  gates.
- Participant count 1 (codex only) in all three rounds — no cross-backend agreement claim is
  possible; disagreement_flag null throughout.

AUDIT-VERDICT: PASS-WITH-DEBT spec=card-t1928 receipts=rcpt-990c49b70a8796df7ba249d0,rcpt-a90446f0ea5ae36e7277fdf3,rcpt-f49524a64c0bfad878348db5
