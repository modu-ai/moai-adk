auditor-model: claude-opus-5-5

# SPEC Review Report: SPEC-MERGE-WINDOW-QUEUE-001 — hunk-limited re-read (card t1479)
Verdict: FAIL (hunk fix itself confirmed; the required codex gate returned fail with two new findings)
Scope: operator decision — one-line fix + re-read, not a full audit. Reasoning context ignored per M1 Context Isolation.
verdict: FAIL
audited_sha: 98e0296cb (full SHA not obtained — see Gaps)

## Claim
1. The iter5 P2 (pinned SHA == record base => "Already up to date", undefined outcome) is closed in the text.
2. The required cross-model gate (codex, adversarial, baseBranch) did not pass.

## Evidence (files read at the tree's working copy, HEAD stated by caller as 98e0296cb)
- spec.md:174-176 REQ-MWQ-017: "require the pinned SHA to differ from that absorbed commit (where they are equal there is nothing to merge, and the step refuses before invoking `git merge`)" — ordered after the tip-equality check and before the ancestry check. CONFIRMED.
- spec.md:181-192 REQ-MWQ-018: "of which there are ten ... (10) nothing to merge — the pinned SHA equals the record's absorbed commit, refused before `git merge` runs. Causes 1-6, 9 and 10 occur before the integration branch moves and promote the next live ticket on release". Distinct code, pre-merge class, window released. CONFIRMED.
- acceptance.md:147 AC-MWQ-018 row 10: "pinned SHA = record base = develop tip | none — `git merge` never invoked (merge seam records zero calls) | yes". Promotion "yes" consistent with rows 1-6, 9. acceptance.md:132 "(ten codes)". CONFIRMED.
- spec.md:215-219 REQ-MWQ-019: "When any state transition fails after the merge commit exists (including a card version or lease changed after step 1) ... distinct from the ten REQ-MWQ-018 codes." Widened as asked. CONFIRMED.
- Count words: design.md:77 "assigns ten distinct values (cause 10: pinned SHA equals `record.base` ...)"; acceptance.md:174 "the ten AC-MWQ-018 codes"; spec.md:182/219 "ten". grep for `\bnine\b` across spec/acceptance/design: 0 hits. CONSISTENT.
- No new internal contradiction found inside the edited hunks.
- codex_audit (adversarial, baseBranch, project_root=agent-a1d44be9332845923, no seeded hypotheses): verdict "fail".
  - [P1] spec.md:166 — the direct verb `moai integration merge --card` carries no card-state/lease/version gate (those live only in complete, REQ-019 step 1); a window holder with a valid record can merge a card leased by another lane or still in `sync-audit`. Codex reports an overlay test showing the transition API rejects those (`ErrLeaseHolder`, `ErrIllegalTransition`) — i.e. the verb path bypasses gates the system otherwise enforces.
  - [P2] spec.md:179 — success is judged by merge-commit tree equality only; no pre/post cleanliness check of the integration worktree. Codex reproduced in a temp repo: `merge.autostash=true` + uncommitted change → merge exit 0, tree matches record, but `UU file` remains; next lane would be promoted onto a conflicted tree.
  - No `audit_receipt` field on the result.

## Baseline-attribution
Spec text read from /Users/goos/MoAI/moai-adk-go/.claude/worktrees/agent-a1d44be9332845923 working copy in this run. Codex result from the moai MCP server build 45600e4ee (binary lag vs tree HEAD 98e0296cb reported by the tool itself; spec-text review is unaffected, but codex's overlay test ran against that tree's code, not the binary).

## Gaps
- `git diff 046d37f0e..98e0296cb` and `git rev-parse HEAD` were REFUSED by this session's worktree-isolation guard (cross-tree git). Fallback: read the edited regions directly; the hunks were located by content, not by diff. Edits outside these regions (if any) were not observed. HEAD SHA is the caller's figure, unverified.
- Codex's two reproductions were not re-run by me; I judged them on the spec text (both gaps are real in the text: REQ-017 lists no card gate; REQ-017/018 name no integration-worktree cleanliness precondition other than the post-abort check in causes 6/7).

## Residual-risk
- P1 and P2 are new surfaces, not regressions from this fix; they fall outside the operator's hunk scope but the codex gate is required, so they block.
- Iteration ceiling: this is past iter5; further rounds need an operator decision.

## Blockers
B1 (codex P1, critical/blocking): add card gates (merge-ready state, caller's own unexpired card lease, version, card == window's recorded card) to the REQ-MWQ-017 verb path before any merge, with an AC row.
B2 (codex P2, major/blocking): require the integration worktree clean (porcelain empty, no MERGE_HEAD) before `git merge`, and after the merge; a post-merge dirty state goes to the cause-8 class (keep commit, `hold`, release). Add AC rows; keep the cause count words in sync.

AUDIT-VERDICT: FAIL spec=SPEC-MERGE-WINDOW-QUEUE-001 receipts=none
