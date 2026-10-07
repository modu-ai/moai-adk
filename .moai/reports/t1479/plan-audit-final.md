auditor-model: claude-opus-5-5

# SPEC Review Report: SPEC-MERGE-WINDOW-QUEUE-001 (card t1479) — FINAL hunk re-read (Q22)
Verdict: FAIL (revised from PASS-WITH-DEBT after audit_multi round 2 — see Addendum)
Overall Score: 0.78
Scope: B1/B2 delta from .moai/reports/t1479/plan-audit-reread.md (HEAD 98e0296cb -> a4df5e8a0), Q22 convergence rule.
Reasoning context ignored per M1 Context Isolation.

verdict: FAIL
audited_sha: a4df5e8a0

## Addendum (supersedes the verdict below)
The Stop hook required a receipt, so I ran mcp__moai__audit_multi with project_root set to the agent tree. Result: overall_verdict fail, disagreement_flag true, required codex = fail, glm inconclusive (fail-open). No audit_receipt was returned, so receipts=none.
- NEW P1 (critical under Q22, data loss): spec.md:L182. Running `git merge --no-ff <pinned SHA>` silently overwrites an ignored local file in the integration worktree when the candidate starts tracking that path. Codex reproduced it in a temp repo: porcelain was empty before and after, exit code 0, the tree matched the record, and `runtime.local` was destroyed. The REQ-MWQ-017 clean checks are porcelain-based and cannot see ignored files. The integration worktree holds ignored runtime files (for example `settings.local.json`), so this is a real data-loss path that the SPEC's own gates pass as clean.
  Required fix (blocking): before `git merge`, the merge step lists the paths the pinned SHA adds relative to the integration tip and refuses with a new cause if any of them exists as an ignored or untracked file or directory in the integration worktree, leaving the file in place. Bump the cause count (thirteen) everywhere, and add an AC row: an ignored file collides with a candidate-tracked path → refused before `git merge`, file bytes preserved, window released.
- The P2 autostash finding duplicates O3.
- I did not reproduce the P1 myself. I judge it real on git semantics: git treats ignored files as expendable during merge and checkout.

## Claim
B1 and B2 are closed. No critical defect (unsafe develop move, data loss, or a contradiction that makes an AC unimplementable) was found. Three new non-critical findings from codex become run-phase obligations.

## Evidence
- B1 closed. spec.md:L171-173 REQ-MWQ-017: "apply the same card gate as `moai factory complete` (REQ-MWQ-019 step 1: card `merge-ready`, caller holds the card's unexpired lease, version as read) and require the requested card to equal the window record's card". All of this runs before `git merge` (L182). REQ-MWQ-018 cause 11 (L196-198) says "with the card and the integration branch untouched". AC-MWQ-018 rows 11a/11b/11c (acceptance.md:L148-150) test the foreign lease, not-merge-ready, and card != window card cases.
- B2 closed. Pre-merge clean check: spec.md:L173-174 says "`git status --porcelain` empty, no unmerged paths, no autostash residue". Post-merge check: L183 says "the integration worktree is clean again". Cause 12 covers a dirty tree before the merge (L198-199). Cause 8 covers a dirty tree after the merge, including autostash residue (L193-195), and gives hold plus a preserved commit (L200-204). AC rows 12 and 8b are at acceptance.md:L151-152.
- Counts agree. "twelve" appears at spec.md:L186 and L226, acceptance.md:L132 ("twelve codes; rows 11a-11c share code 11") and L179, and design.md:L77. Cause partition at L199-200: 1-6 and 9-12 happen before the merge, 7 and 8 after; together they are the full set of 12.
- Cross-model check: mcp__moai__codex_audit was run in adversarial mode with target baseBranch and project_root set to the agent tree. The verdict was "fail" with three P2 findings and no P1. The response carried no audit_receipt. Classification under Q22:
  1. P2 — adoption path (REQ-MWQ-019 step 2, spec.md:L222-226). When a post-adoption transition fails, the spec says "release the window", but the caller may not hold the window at that point. If a foreign holder has the window, foreign-release protection refuses the release. Develop never moves and no data is lost. AC-019 scenario 8 exercises only the merge-step path, so no AC becomes unimplementable. Classified NON-CRITICAL.
  2. P2 — `git status --porcelain` misses untracked files when `status.showUntrackedFiles=no` is set (codex reproduced this in a temp repo). REQ-MWQ-016 L164 already states "tracked and untracked", so the intent is clear and the fix belongs in the implementation's choice of flag. Classified NON-CRITICAL.
  3. P2 — AC-MWQ-018 row 8b. Its fixture ("conflicting local edit") would be refused by cause 12 before the merge. I confirmed this independently by reading L173-174 against L152. The row can still be implemented: a test seam can inject the residue after the pre-check passes. Classified NON-CRITICAL.

## Baseline-attribution
- Files were read directly from /Users/goos/MoAI/moai-adk-go/.claude/worktrees/agent-a1d44be9332845923/.moai/specs/SPEC-MERGE-WINDOW-QUEUE-001/ (spec.md L164-227, acceptance.md L114-183) and from decision-index/design via grep in this run.
- The codex judging build was commit 45600e4ee, which is an ancestor of HEAD a4df5e8a0 (build_lag reported). The review read SPEC markdown, so a lagging binary does not affect what was reviewed.

## Gaps
- `git -C <agent tree>` was refused by the worktree guard. I therefore did not observe the HEAD SHA a4df5e8a0 or the 98e0296cb..a4df5e8a0 diff stat myself; I took them from the dispatch. Files were read by absolute path instead.
- I did not run the CN-4, D7 or D8 verbs again. This was a delta-scoped final round, and MP-9 was not re-decided from the verb output.
- No audit_receipt was issued, so receipts=none.

## Residual-risk
- "version as read" in REQ-MWQ-017 does not say where the merge verb reads the version from. The run phase must pin it to complete's T14 read.
- The adoption path's hold/release ownership is underspecified (obligation O1).

## Run-phase obligations
- O1: Separate failure handling for the adoption path from the merge-step path. A post-adoption transition failure must not release or alter a window held by another session; release only when the caller is the holder. Add an AC/test that a foreign holder's window record bytes are unchanged.
- O2: Run the pre-merge and post-merge clean checks as `git status --porcelain --untracked-files=all`, which overrides `status.showUntrackedFiles`, and include a regression case with `status.showUntrackedFiles=no`.
- O3: Implement AC-MWQ-018 row 8b by injecting the dirty state or autostash residue through a seam after the cause-12 pre-check passes, not with a pre-existing local edit. Assert that `hold` is written before release.
- O4: Pin the source of "version as read" in the merge verb's card gate to the same read that REQ-MWQ-019 step 1 uses.

## Defects Found
D1. codex-P2-adopt — spec.md:L222-226 — adoption failure path releases a window it may not hold — Severity: major — Class: optional (run-phase O1) — Required fix: O1
D2. codex-P2-untracked — spec.md:L173 — porcelain check is config-sensitive — Severity: minor — Class: optional (O2) — Required fix: O2
D3. codex-P2-8b — acceptance.md:L152 — fixture timing collides with cause 12 — Severity: minor — Class: optional (O3) — Required fix: O3

## Regression Check
- B1: RESOLVED (spec.md:L171-173, L196-198; acceptance.md:L148-150)
- B2: RESOLVED (spec.md:L173-174, L183, L193-199; acceptance.md:L151-152)
