auditor-model: claude-sonnet-5-5[1m]
verdict: FAIL
audited_sha: 9d9d5fffa
score: 0.88

# SPEC Delta Re-read (final): SPEC-MERGE-WINDOW-QUEUE-001 (card t1479), v0.8.0

Reasoning context ignored per M1 Context Isolation. `git rev-parse --short HEAD` = `9d9d5fffa`, branch `WT-merge-window-queue`. Delta audited: `git diff d468ff19c 9d9d5fffa -- .moai/specs/SPEC-MERGE-WINDOW-QUEUE-001` (acceptance.md +8/-2, decision-index.md +15, design.md 1 line, plan.md 12 lines, progress.md +17, spec.md 20 lines). No earlier verdict was used as evidence; the prior FAIL file was read only for its defect list.

## Claim

1. D1 (critical, ancestor-path collision) is CLOSED for the shapes the SPEC names. The widened predicate refuses the ancestor case (13b) and the beneath case (13c), and my own scratch runs show plain `git merge --no-ff` destroys the ignored bytes in both (merge exit 0, status clean), so each refusal prevents a real data loss.
2. D2 (plan.md count and order) is CLOSED. "Thirteen" is consistent across the four artifacts; plan.md M5's pre-merge order matches REQ-MWQ-017 step for step.
3. One new blocking defect remains in the predicate's wording (D5 below): a directory-to-file type change over a tracked directory holding an ignored file. Plain merge destroys the ignored file (reproduced). Whether the predicate refuses it depends on whether "path" means tree entries or leaf entries; the SPEC does not say. The required codex gate returned FAIL on exactly this, and a backend FAIL is not downgraded here. No unsafe-develop-move or unimplementable-AC defect was introduced by the delta.

## Evidence

### (1) Reproduction against plain `git merge --no-ff` (Git 2.54.0 Apple Git-157, scratch repos under the session scratchpad, script `repro.sh`)

Command: `bash <scratchpad>/repro.sh`. Verbatim output:

```
git version 2.54.0 (Apple Git-157)
== 13b ancestor
status lines:        0
merge exit 0
drwxr-xr-x@ 3 goos  wheel  96 Oct  4 15:03 runtime.local
== 13c beneath
status lines:        0
merge exit 0
-rw-r--r--@ 1 goos  wheel  1 Oct  4 15:03 runtime.local
cat: runtime.local/keep: Not a directory
== 13d dir->file type change, dir tracked in both trees (codex P1)
status lines:        0
merge exit 0
-rw-r--r--@ 1 goos  wheel  1 Oct  4 15:03 runtime.local
ls: runtime.local/secret: Not a directory
```

- 13b: candidate adds `runtime.local/payload`; integration tree holds ignored regular file `runtime.local`. Merge exit 0, `runtime.local` is now a directory (file bytes gone). Predicate as worded: `runtime.local/payload` is an added path; its ancestor `runtime.local` exists as an ignored file, so spec.md REQ-MWQ-017 refuses. Closed.
- 13c: candidate adds file `runtime.local`; integration tree holds ignored directory with `keep`. Merge exit 0, directory replaced by file, `keep` gone. Predicate: `runtime.local` is added; `runtime.local/keep` is "beneath such a path" and exists, so refused. Closed.
- 13d (shape not in the SPEC): tip tracks `runtime.local/tracked`; candidate replaces the directory with a file `runtime.local`; integration tree holds ignored `runtime.local/secret`. Merge exit 0, `secret` destroyed (status clean beforehand).

Path-set probe for 13d before merge (`repro2.sh`, same scratch pattern), verbatim:

```
-- diff-filter=A (added leaf paths) tip..cand:
runtime.local
-- ls-tree -r leaf diff cand-only:
runtime.local
-- tree-entry (ls-tree -rt) diff cand-only:
-- end (before merge)
```

Reading: with leaf-entry path sets (`git diff --diff-filter=A`, `git ls-tree -r`), `runtime.local` is an added path and the "beneath" clause refuses 13d. With tree-entry path sets (`git ls-tree -rt`), the added-path set is empty and 13d passes. spec.md:183-185 defines added as "a path present in the former and absent from the latter" without saying which entry kind counts, so both implementations conform to the text.

### (2) Count consistency and order

`grep -niE "nine causes|twelve|eleven|ten causes|thirteen"` over spec.md, acceptance.md, design.md, plan.md (excerpts):
- spec.md:192 "of which there are thirteen"; spec.md:236 "distinct from the thirteen REQ-MWQ-018 codes".
- acceptance.md:132 "thirteen codes; rows 11a-11c share code 11 and rows 13a-13c share code 13"; acceptance.md:184 "differs from the thirteen".
- design.md:77 "assigns thirteen distinct values (cause 13: an added path, an ancestor path of an added path, or a p...".
- plan.md:93 "thirteen causes". No "nine", "twelve", "eleven" residue (spec.md:31-32 matches are HISTORY rows quoting audit scores, not counts).

Order, spec.md:167-190 vs plan.md:84-97 (M5): holder check (read only) -> renew/drops -> card gate (merge-ready, own unexpired lease, version as read; requested card equals window card) -> clean worktree -> branch resolve + pin -> record valid -> base equals tip -> pinned SHA differs from base -> descends from base -> tree identity -> landing check -> collision check -> `git merge --no-ff` -> merge-tree and clean verification -> release. plan.md lists the same sequence. Identical. Partition unchanged: causes 1-6 and 9-13 pre-merge/promote, 7-8 hold.

AC-MWQ-018 rows 13a/13b/13c exist (acceptance.md:152-154) and the thirteen-code statement now covers the shared code 13. Row 13b states the RED fixture steps (create ignored `runtime.local`, force-add `runtime.local/payload` on `cand`, merge seam instrumented) and a RED-now cell citing the prior audit's measured reproduction, pinned to tree `25b0eeec8`. That RED-now cell cites a prior report rather than a re-executable command; under verification-completeness.md §2.1 it is therefore not a re-executable release-blocking RED. I re-ran the equivalent myself (13b above) and the RED reproduced. MP-8 status: the criterion is acceptable as a regression-guard with my own reproduction on record; the run phase writes the real fixture first (progress.md, AC text), so this is optional.

### (3) New-defect scan on the delta
- No contradiction among REQ-MWQ-017/018, AC-MWQ-018 rows and plan M5. The check precedes `git merge`, so no develop move occurs on refusal.
- progress.md O2-extended (progress.md around line 64) says the untracked half of cause 13 is reachable only under `status.showUntrackedFiles=no`, while the base O2 (same file, lines 49-51) forces `--untracked-files=all`, which makes cause 12 catch it first. These two obligations conflict (codex P2; reproduced by codex, consistent with git semantics). Run-phase obligation, optional.
- MP-7: `grep -rn "NEEDS CLARIFICATION" plan.md research.md` returned nothing. PASS.

### Cross-model check (`mcp__moai__audit_multi`, project_root `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/agent-a1d44be9332845923`, target baseBranch, adversarial focus on the delta)

- codex: gate required, verdict FAIL. P1 `spec.md:184` "new path" definition misses a directory-to-file/link change over an existing directory (reproduced by codex on Git 2.54.0: merge exit 0 deleted an ignored `runtime.local/secret`). P2 `progress.md:64` extended O2 conflicts with base O2 (`--untracked-files=all` sends untracked collisions to cause 12). Receipt id: `rcpt-13c26682e63ffcbf523240eb`.
- claude, glm: not present in `per_backend_verdicts`; `participant_count: 1`. Only codex contributed. No cross-model agreement is claimed. `overall_verdict: fail`, `residual_risk_note: "required-backend FAIL: codex"`, `fail_open_backends: []`.
- build lag note from the tool: installed moai built from `0732cc699`, an ancestor of HEAD; output describes older code.

## Baseline-attribution

All observations were measured in this run against this worktree at HEAD `9d9d5fffa` (SPEC artifact bytes are those of the committed delta; `git status --short` shows only a pre-existing modification of `.moai/config/sections/workflow.yaml`, not a SPEC artifact). Reproductions ran on system git 2.54.0 in throwaway repos; they model the merge of the pinned SHA, not the moai binary. The auditing tool build (`0732cc699`) is an ancestor of HEAD, so the codex output is not evidence about code newer than that.

## Gaps

- Claude and GLM backends contributed no verdict through `audit_multi`; the cross-model check is single-backend (codex only).
- Full CN-4 / MP-9 and MP-8 verbs not re-run in full; delta-scoped pass. No milestone-order text changed except plan.md M5, which I compared by hand to spec.md (above).
- The first compound Bash command (scratch repro in one line) was refused by the worktree-isolation guard ("too complex to verify that it stays inside the worktree"); I reran it as a script file. The `Grep` tool was unavailable in this session; I used `grep` via Bash. The Read of the prior report and file greps were unaffected.
- No test of the symlink or case-insensitive-filesystem collision shapes (deliberately O5, not measured).
- Interpretation of "path" in the predicate is untested against the run-phase implementation, which does not exist yet.

## Residual-risk

Even with D5 fixed, collision shapes outside tree/leaf-entry reading (symlinks, case folding, submodule/gitlink paths, a candidate that deletes a tracked file inside a directory that also holds an ignored file) remain for the run phase to test against a varied fixture set; O5 covers part of that.

## Defects Found

D5. cause13-path-definition-dir-to-file — spec.md:183-185 (REQ-MWQ-017, "a path present in the former and absent from the latter"), spec.md:205-209 (cause 13), acceptance.md AC-MWQ-018 rows 13a-13c — the added-path set is defined on "paths" without stating whether directory tree entries count. Under a tree-entry reading, a candidate that replaces tracked directory `runtime.local/` with file `runtime.local` adds nothing, passes cause 13, and plain `git merge --no-ff` deletes an ignored `runtime.local/secret` (reproduced, exit 0). Under a leaf-entry reading the "beneath" clause refuses it. Same data-loss class as D1 on a narrower shape; backend fail from codex is not downgraded. — Severity: major — Class: blocking — Required fix (wording-only, no new REQ): define "path" in REQ-MWQ-017 as a leaf entry (file, symlink, or gitlink, i.e. the paths `git ls-tree -r` lists, equivalently `git diff --diff-filter=A` on the pinned SHA against the tip), and say that an entry which is a directory in one tree and a leaf in the other counts as added; add AC-MWQ-018 row 13d (tracked directory replaced by a file, ignored file inside the old directory; refused, bytes unchanged, merge seam zero calls, cause 13's code). If the leader prefers to carry this as a run-phase obligation instead, that is the leader's disposition, not this verdict's.

D6. o2-extended-conflict — progress.md O2-extended (line ~64) vs progress.md lines 49-51 (base O2) — O2 requires clean checks with `--untracked-files=all`, so a non-ignored untracked collision is always refused as cause 12 and cannot reach cause 13 under `status.showUntrackedFiles=no`; the extended obligation asks for a regression that expects cause 13 there. — Severity: minor — Class: optional — Run-phase obligation candidate: state the expected cause for the untracked-collision regression as cause 12 and test cause 13's non-ignored branch with a check-level unit test.

D7. row13b-red-now-cell-form — acceptance.md after AC-MWQ-018 (RED-now cell for row 13b) — the RED-now cell cites a prior audit report rather than a command with verbatim stdout, exit code and tree SHA. — Severity: minor — Class: optional — Run-phase obligation candidate: the run-phase writes the real fixture first and records its observed failure in progress.md §E.2; keep the cited report as provenance only.

## Regression Check

- D1 (critical, ancestor-path ignored-file overwrite): RESOLVED for the named shapes (13a exact, 13b ancestor, 13c beneath); evidence above.
- D2 (major, plan.md "nine causes"): RESOLVED; plan.md:93 "thirteen causes", order identical to REQ-MWQ-017.
- D3 (optional, untracked shadow): carried in progress.md O2-extended; see D6 for a conflict it introduces.
- D4 (optional, symlink/case): carried as O5; unchanged.

## Must-Pass Summary

- MP-1: PASS (REQ/AC counts 23/23 per progress.md; no count text changed; delta added no REQ).
- MP-5 / MP-6: not re-run for this delta (no cross-SPEC or `syscall` text added by the delta; unverified rather than passed).
- MP-7: PASS (no `[NEEDS CLARIFICATION]` in plan.md or research.md).
- MP-8: UNVERIFIED for the new row 13b cell (see D7); my own reproduction of the RED reproduces.
- MP-9: no ordering text changed besides plan.md M5, compared by hand: consistent.
- Required codex gate: FAIL (not downgraded).

## Recommendation

One wording sentence in REQ-MWQ-017 plus one AC row (13d) closes D5; D6/D7 are run-phase notes. The delta otherwise achieves what Q24 asked: D1 and D2 are closed and no new unsafe develop move or unimplementable AC was introduced.

## Operational Notes (unverified)

- Measure the added-path set the run phase will use: `git diff --name-only --diff-filter=A <base> <pinned-sha>` against a type-change fixture; expect it to list the file that replaced the directory. Status: assumption (single scratch fixture, measured above only for `--diff-filter=A` and `ls-tree -r`).
- Measure whether `git merge-tree --write-tree <base> <pinned-sha>` can predict the working-tree overwrite instead of hand-rolled path logic. Status: assumption; not run.

AUDIT-VERDICT: FAIL spec=SPEC-MERGE-WINDOW-QUEUE-001 receipts=rcpt-13c26682e63ffcbf523240eb
