auditor-model: claude-sonnet-5-5[1m]
verdict: FAIL
audited_sha: d468ff19c
score: 0.80

# SPEC Delta Re-read: SPEC-MERGE-WINDOW-QUEUE-001 (card t1479)

Reasoning context ignored per M1 Context Isolation. HEAD is `25b0eeec8` (merge of develop on top, branch WT-merge-window-queue). SPEC-artifact commit audited: `d468ff19c` (delta from `a4df5e8a0`; `git log -- .moai/specs/SPEC-MERGE-WINDOW-QUEUE-001` shows d468ff19c as the latest SPEC commit).

## Claim

1. Cause 13 closes the original data-loss critical for the exact-path collision: the check sits before `git merge`, is read-only (bytes left in place), the window is released, and the cause has its own code. PARTIAL: it does not cover an ancestor-path collision (below).
2. "Thirteen" is consistent in spec.md, acceptance.md and design.md. It is NOT consistent in plan.md (M5 still says "nine causes" and its pre-merge check list omits cause 13). The pre-merge/post-merge partition is complete and consistent within spec.md.
3. The delta introduced one new critical defect of the same class as the one it fixes: a candidate that adds `runtime.local/payload` while the integration worktree holds an ignored regular file `runtime.local` passes cause 13 and `git merge` silently destroys the file. Reproduced by this auditor.

## Evidence

Delta (`git diff a4df5e8a0 d468ff19c --stat`): acceptance.md, decision-index.md, design.md, progress.md, spec.md (39 insertions, 7 deletions).

(1) Quoted text.
- spec.md REQ-MWQ-017: "require that no path the pinned SHA newly adds relative to the integration branch tip already exists in the integration worktree as an ignored or untracked file or directory (leaving such bytes untouched), run `git merge --no-ff <pinned SHA>`" — the check is ordered before the merge.
- spec.md REQ-MWQ-018: "(13) a path the pinned SHA adds already exists in the integration worktree as an ignored or untracked file or directory — refused before `git merge` with that file's bytes untouched. Causes 1-6 and 9-13 occur before the integration branch moves and promote the next live ticket on release"; the lead-in states "the merge step shall release the window and exit with a code distinct per cause".
- acceptance.md AC-MWQ-018 row 13: "pinned SHA adds `runtime.local`; the integration worktree holds an ignored `runtime.local` with known bytes | none — `git merge` never invoked | yes | the ignored file's bytes are unchanged (checksum equal before and after); window released". The row-level lead says each row has a distinct exit code.
- Gap in the wording: the check tests only whether the added path itself exists. Counter-case reproduced (scratch repo, below).

(2) Count consistency, `grep -niE "twelve|thirteen|nine causes|..."` over the four artifacts:
- spec.md:188 "of which there are thirteen"; spec.md:230 "distinct from the thirteen REQ-MWQ-018 codes".
- acceptance.md:132 "thirteen codes; rows 11a-11c share code 11"; acceptance.md:180 "differs from the thirteen".
- design.md:77 "assigns thirteen distinct values (cause 13: ...)".
- plan.md:90 (verbatim): "...tree identity, landing check — then `git merge --no-ff <sha>`, merge-tree verification, release; nine causes with distinct exit codes". Stale. It was already stale at twelve (a4df5e8a0 and earlier), and the same line omits causes 10-13 from the pre-merge check order.
- Row count in AC-MWQ-018: rows 1-10, 11a/b/c (one code), 12, 13, 8b (same code as 8) = 13 distinct codes. Consistent.
- Partition: pre-merge = 1-6, 9-13; post-merge hold = 7, 8 (spec.md REQ-MWQ-018). Complete, no overlap. Cause 13 in the REQ-MWQ-017 order falls after the landing check and before `git merge`, so "released and next ticket promoted" matches AC row 13 "C promoted: yes".
- Overlap note: an untracked (non-ignored) collision is already refused by cause 12 (`git status --porcelain` non-empty) first, so AC row 13 only exercises the ignored form, and cause 13's "untracked" half is reachable only under `status.showUntrackedFiles=no` (obligation O2).

(3) Ancestor-path counter-case, reproduced:
```
cd <scratch> && git init -b main; .gitignore="runtime.local"; commit; branch cand adds runtime.local/payload (git add -f); on main: printf SECRET > runtime.local; git merge --no-ff --no-edit cand
merge exit 0
ls -la runtime.local  ->  drwxr-xr-x ... (now a directory; file SECRET gone)
```
The merge exited 0 and replaced the ignored regular file with a directory. The cause-13 test as worded ("no path the pinned SHA newly adds ... already exists") checks `runtime.local/payload`, which does not exist, so it would pass and the data would be lost. Exact-path case (original critical) is fixed; this variant is not.

Cross-model check (`mcp__moai__audit_multi`, project_root = this worktree, target baseBranch, adversarial focus on the delta):
- codex: required gate, verdict FAIL. P1: ancestor-component collision not covered (spec.md:182); reproduction reported, matching mine. P2: plan.md:90 "nine causes" inconsistent with thirteen. Receipt id: `rcpt-cb17bce86499722ecce24590`.
- claude: not present in `per_backend_verdicts` (participant_count 1). glm: not present. Only codex returned a verdict; `overall_verdict: fail`, `residual_risk_note: "required-backend FAIL: codex"`. The other two backends contributed nothing, so no cross-model agreement is claimed.
- `build_lag`: installed moai binary built from `0732cc699`, an ancestor of HEAD; tool output reflects that build.

## Baseline-attribution

All figures measured in this run against this worktree at HEAD `25b0eeec8`; SPEC artifacts unchanged since `d468ff19c`. The scratch reproduction used the system git, in a throwaway repo under the scratchpad directory. The earlier verdict `plan-audit-final.md` was not relied on; I read only the task description for its summary.

## Defects Found

D1. cause13-ancestor-path — spec.md:182-184 (REQ-MWQ-017), spec.md:200-202 (REQ-MWQ-018 cause 13), acceptance.md AC-MWQ-018 row 13 — the cause-13 test covers only the added path itself, so a candidate adding `runtime.local/payload` over an ignored file `runtime.local` still silently destroys it (reproduced, merge exit 0). Same data-loss class as the critical the delta set out to close. — Severity: critical — Class: blocking — Required fix: widen the check to "no path the pinned SHA newly adds, nor any ancestor directory component of such a path, exists in the integration worktree as an ignored or untracked file or symlink (and no added path's existing descendant is an ignored file when the SHA adds a file there)"; add an AC-MWQ-018 row 13b (ancestor file collision: bytes unchanged, merge seam zero calls, tip unchanged, window released, cause 13's code).

D2. plan-cause-count — plan.md:90 — M5 says "nine causes" and its pre-merge order lacks causes 10-13 (pre-existing since the earlier counts, not introduced by d468ff19c, but the thirteen-count is now stated to be consistent project-wide) — Severity: major — Class: blocking (the count is a statement this SPEC makes about itself and the reviewer's question 2 requires consistency) — Required fix: change to "thirteen causes with distinct exit codes" and add card gate, clean check, nothing-to-merge and ignored/untracked-collision checks to the order.

D3. cause13-untracked-shadow — acceptance.md AC row 13 — the untracked half of cause 13 is shadowed by cause 12 and has no row. — Severity: minor — Class: optional — Run-phase obligation candidate: cover it together with O2 (`status.showUntrackedFiles=no`).

D4. cause13-symlink/case-insensitive — spec.md:182 — symlink and case-folding filesystem (macOS default) collisions are unspecified. — Severity: minor — Class: optional — Run-phase obligation candidate: use `git ls-files`/`lstat`-based existence test, not a case-sensitive path equality.

## Regression Check

Prior critical (exact-path ignored-file overwrite): RESOLVED for the exact-path form (REQ-MWQ-017/018 cause 13, AC row 13). O1-O4 recorded in progress.md §E.1 remain run-phase obligations (not re-audited here, not blockers per Q22).

## Must-Pass Summary

- MP-1: PASS (REQ-MWQ-017..021 retained, count text consistent for REQ/AC at 23/23 per progress.md; not re-enumerated in this delta).
- MP-9 / CN-4: not run in full (delta scope; no milestone-order text changed). UNVERIFIED for this delta.
- Required codex gate: FAIL (cannot be downgraded).
- Overall: FAIL on D1 (new critical in the delta's own fix) and the required codex FAIL; D2 is an independent blocking inconsistency.

## Recommendation

Fix D1 (and D2 in the same pass), then one narrow re-read of that delta. If D1 is fixed by wording alone, the re-read need cover only REQ-MWQ-017/018, AC row 13/13b, and plan.md:90.

## Gaps

- Claude and GLM backends returned no verdict through audit_multi; only codex contributed.
- Full CN-4/MP-8/MP-9 verbs were not run (delta-scoped pass).
- Reproduction used the scratch repo's git version, not the project's CI git.

## Residual-risk

Even after D1, git can lose ignored files through other paths (e.g. directory-over-file replaced by a deleted-then-added rename); the run phase should test the check against a randomized collision fixture set rather than the single row 13 case.

AUDIT-VERDICT: FAIL spec=SPEC-MERGE-WINDOW-QUEUE-001 receipts=rcpt-cb17bce86499722ecce24590
