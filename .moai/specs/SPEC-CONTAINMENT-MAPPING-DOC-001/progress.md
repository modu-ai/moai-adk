# Progress: SPEC-CONTAINMENT-MAPPING-DOC-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-16
tier: M (spec.md, plan.md, acceptance.md)

### Plan fix round 2 record (card t1760)

Scope: the two blocking findings N1 and N2 of the round-2 plan audit (verdict FAIL, 0.875), plus optional O1, O2, O3. Dispositions: `plan.md` section J. Round-1 blockers D1 and D2 were not touched.

Tree: HEAD `1b2319571070c4d3b60d270037d2f90f7d7ffa8d`, branch `WT-10-15-release-3`; the working tree was that HEAD plus uncommitted edits confined to this SPEC directory. Tools used: `git`, `grep` (in this shell a function over ugrep; its version string was not observed), `awk`, and the CN-4 awk program extracted from `.claude/agents/moai/plan-auditor.md` lines 477-545. The only project tooling used is the `mcp__moai__spec_audit` call at the end; its judging build is the MCP server `v3.2.0-rc.36-t1822c`, commit `ee981a932` (server banner), and `git merge-base --is-ancestor ee981a932 HEAD` printed exit 1, so that commit is not an ancestor of this tree's HEAD `1b2319571` and the build is not a lagging ancestor.

Claim 1 (N1): E10a no longer fails on tracked card evidence.

```text
$ git diff --quiet 629bb51cda73fc70422769988b8c28182c9bd2d5 -- . ':!.moai/docs/multi-agent-containment-mapping.md' ':!.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001'
E10a-current exit=1
$ git diff --quiet 629bb51cda73fc70422769988b8c28182c9bd2d5 -- . ':!.moai/docs/multi-agent-containment-mapping.md' ':!.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001' ':!.moai/reports'
E10a-with-reports-excluded exit=0
$ git status --porcelain --untracked-files=all -- . ':!.moai/docs/multi-agent-containment-mapping.md' ':!.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001' ':!.moai/reports'
E10b-new exit=0        (stdout empty)
$ git diff --quiet 46960ac7d8043f1bee3f352688cceba8fb405738 -- . ':!.moai/docs/multi-agent-containment-mapping.md' ':!.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001' ':!.moai/reports'
E10a-new vs older base 46960ac7d exit=1
$ git ls-files .moai/reports/t1760
.moai/reports/t1760/plan-audit.md
.moai/reports/t1760/progress.md
```

E10b probe: with an untracked `.moai/docs/zz-e10-probe.md` present, the new E10b printed `?? .moai/docs/zz-e10-probe.md` (exit 0); the file was then removed with `rm` and `git status --porcelain --untracked-files=all` printed nothing (exit 0). This created and deleted one file outside the SPEC directory; no tracked file was touched.

Claim 2 (N2): E08 now counts the date shapes AC-CMM-008 promises. The exact fix-round-2 command is in `acceptance.md` (ledger entry E08); each run below used it with only the file argument changed.

```text
24   NEW on e08_hit.txt (24-line must-count) exit=0
0    NEW on e08_zero.txt (26-line must-stay-zero) exit=1
18   NEW on r2/e08_hit_new.txt (18-line new date forms) exit=0
0    NEW on r2/e08_zero_new.txt (21-line extra must-stay-zero) exit=1
3    NEW on r2/e08_three.txt (october 16 / OCTOBER 16 / 16 October) exit=0
0    NEW on f_pass.md (Korean passing fixture) exit=1
1    NEW on r2/f_mut_dayfirst.md (passing fixture + one 16 October line) exit=0
0    NEW on r2/e08_limits.txt (7 shapes expected uncounted) exit=1
4    OLD (fix-round-1 command) on r2/e08_hit_new.txt exit=0
0    OLD (fix-round-1 command) on r2/e08_three.txt exit=1
2    E08-new on the absent document exit=2   (stderr: ugrep: warning: .moai/docs/multi-agent-containment-mapping.md: No such file or directory)
```

The 2 on the absent-document line is the command's exit code, not a count. `e08_zero.txt` holds 26 lines by `wc -l` (the audit text said 25).

Claim 3 (O3): the CN-4 verb now reads the plan. `awk -f cn4.awk plan.md acceptance.md` printed `COLLECTED: 4 milestones in plan order (M1 M2 M3 M4), 12 exit bindings, 8 ordering candidates`, exit 0, no `CONFLICT:` line. The eight candidates were not each re-read in this round; the audit's earlier hand-read of the same clause family found none ordering work across milestones.

Claim 4: frontmatter and era unchanged in kind. `mcp__moai__spec_audit` (project_root = this worktree, filter SPEC-CONTAINMENT-MAPPING-DOC-001) returned `modern_era_clean: 1`, one INFO `EraAutoDetected` (H-5). ID check: `[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` printed `PASS`.

Gaps: the final document does not exist, so E01 to E09 were observed red on the absent document and on constructed fixtures only (fixtures live in session scratch, not committed); the post-commit run of E10a and E10b is reported in the hand-off message because a commit cannot cite its own hash here; no `audit_multi` or `spec lint` run was made in this round (the independent re-audit is the next step); the fix-round-1 tracked-file probe for E10a and E10b was not re-run; the 8 CN-4 candidates were not re-read one by one.

Residual risk: E08 counts lines and enumerates date shapes, so spellings outside the list in `acceptance.md` (Known limit) pass unseen; a change under `.moai/reports/` is invisible to E10 by design.

### Plan fix round 3 record (card t1760)

Scope: the two blocking findings N3 and N4 of the round-3 plan audit (verdict FAIL, 0.90; the required codex leg failed with two P2 findings the auditor reproduced), plus a disposition for the two optional findings. Dispositions: `plan.md` section J.1. N1, N2, and the round-1 blockers D1 and D2 were not touched, and the E08 command text was not changed.

Tree: HEAD `04890fc1336a8d4b4843be7208afbdbdb3ee8150`, branch `WT-10-15-release-3`, working tree clean at the start; the edits are confined to this SPEC directory. Tools used: `git`, `grep` (a shell function over ugrep here; its version string was not observed), `sh`/`eval`, `awk` (CN-4 program from `.claude/agents/moai/plan-auditor.md` lines 477-545). The only project tooling is the `mcp__moai__spec_audit` call; its judging build is the MCP server `v3.2.0-rc.36-t1822c`, commit `ee981a932` (server banner), and `git merge-base --is-ancestor ee981a932 HEAD` printed `is-ancestor exit=1`, so that commit is not an ancestor of this tree's HEAD.

Claim 1 (N3): the scope guard now sees another card's tracked evidence. Real tree, narrowed pathspec:

```text
E10a-narrow exit=0
E10b-narrow exit=0        (stdout empty)
E10a-broad (round-2 form ':!.moai/reports') exit=0
E10a-narrow vs older base 46960ac7d exit=1
E10a-no-reports-exclusion control exit=1
git ls-files .moai/reports | wc -l    -> 1789
```

Throwaway repository in session scratch (base commit `1203114fe5bd16b10fca5abd31150f723a3dcbc8`; tracked `.moai/reports/t9999/verdict.md`, `.moai/reports/t1760/plan-audit.md`, `src/keep.txt`), the real tree untouched:

```text
clean tree, narrow E10a                                   exit=0
delete .moai/reports/t9999/verdict.md: narrow E10a        exit=1
delete .moai/reports/t9999/verdict.md: broad E10a         exit=0   (the round-2 mutant)
delete .moai/reports/t9999/verdict.md: narrow E10b        stdout " D .moai/reports/t9999/verdict.md", exit=0
delete .moai/reports/t9999/verdict.md: broad E10b         stdout empty, exit=0
delete .moai/reports/t1760/plan-audit.md: narrow E10a     exit=0   (own evidence, invisible by design)
add untracked .moai/reports/t9999/new.md: narrow E10b     stdout "?? .moai/reports/t9999/new.md", exit=0
```

The probe files were restored with `git checkout -- <path>` inside the throwaway repository only. Post-edit, on the real tree: E10a exit 0, E10b stdout empty exit 0, and `git status --short` listed only the three SPEC files modified by this round before the progress edit (`acceptance.md`, `plan.md`, `spec.md`).

Claim 2 (N4): AC-CMM-008 now says what E08 matches. The command was run unchanged (taken verbatim from the ledger by a scratch script). Fixtures: the 24-line and 26-line sets were rebuilt from the lists in `acceptance.md` (the earlier scratch copies had been overwritten); the others are the round-2 files.

```text
count=24 exit=0 file=e08_hit24.txt        (24-line must-count)
count=0  exit=1 file=e08_zero26.txt       (26-line must-stay-zero)
count=18 exit=0 file=e08_hit_new.txt      (18 new date lines)
count=0  exit=1 file=e08_zero_new.txt     (21 further must-stay-zero lines)
count=3  exit=0 file=e08_three.txt        (october 16 / OCTOBER 16 / 16 October)
count=2  exit=0 file=e08_may.txt          (8-line May fixture; May 4 and MAY 4 are the counted lines per the audit, not re-derived line by line)
count=0  exit=1 file=e08_mixedmay.txt     (mAy 4, maY 4, 4 MAy, MaY 4, mAY 4, 4 mAy)
count=1  exit=0 file=e08_range.txt        ("sections 10-12 of the guide")
```

Claim 3: `awk -f cn4.awk plan.md acceptance.md` after the edits printed `COLLECTED: 4 milestones in plan order (M1 M2 M3 M4), 12 exit bindings, 8 ordering candidates`, exit 0, no `CONFLICT:` line. The 8 candidates were listed but not each re-read in this round.

Claim 4: `mcp__moai__spec_audit` (project_root = this worktree, filter this SPEC) returned `modern_era_clean: 1`, `total_specs: 1`, one INFO `EraAutoDetected` (H-5). ID check on `SPEC-CONTAINMENT-MAPPING-DOC-001` against `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` printed `PASS`.

Gaps: the deliverable document does not exist, so E01 to E09 remain observed red on the absent document and green only on constructed fixtures; the fix-round-1 and fix-round-2 E10 probes carried in the ledger were not re-run; the `.moai/lessons-inbox.jsonl` tool-failure rows were not inspected, though no command was refused in this round; no `audit_multi` or `spec lint` run was made (the independent re-audit is the next step); the post-commit E10a and E10b run is reported in the hand-off message because a commit cannot cite its own hash here; optional finding O5 (sync-phase template-mirror check) was not taken.

Residual risk: E08 still counts lines and enumerates date shapes, so mixed-case May, a `10-12` section range (false red), and the shapes listed under Known limit are outside what the command decides; a change under `.moai/reports/t1760/` is invisible to E10 by design; if the run phase merges `main` into the card branch, E10a (fixed base) would read red at the run-phase tip (inferred, not tested).

### Plan fix round 4 record (card t1760)

Scope: the one blocking finding N5 of the round-4 plan audit (verdict FAIL, 0.92; the required codex leg failed with one P2 finding the auditor reproduced), plus the two optional findings, labelled O-a and O-b in `plan.md` section J.2. Dispositions: `plan.md` section J.2. N3, N4, N1, N2, and the round-1 blockers D1 and D2 were not touched; the E08, E10a, and E10b command texts were not changed (`git diff -U0` of `acceptance.md` has no changed `command:` line for them: three `grep -c -F` counts, each 0).

Tree: HEAD `5a8e8dfd5731edfb767c2e468eb82a5b7c268f1d`, branch `WT-10-15-release-3`, working tree clean at the start; the edits are confined to this SPEC directory. Tools used: `git`, `awk` (CN-4 and traceability programs from the earlier scratch copies `r3_cn4.awk` and `r3_trace.awk`, extracted in round 3 from `.claude/agents/moai/plan-auditor.md`; not re-extracted this round), `cp`, `sed`. The only project tooling is the `mcp__moai__spec_audit` call; its judging build is the MCP server `v3.2.0-rc.36-t1822c`, commit `ee981a932` (server banner), and `git merge-base --is-ancestor ee981a932 HEAD` printed `is-ancestor exit=1`, so that commit is not an ancestor of this tree's HEAD.

Claim 1 (N5): a new file under another card's report directory is not seen by the scope guard in this repository, and the criterion now says so. Real tree:

```text
$ git check-ignore -v .moai/reports/t9999/new.md
.gitignore:235:.moai/reports/*	.moai/reports/t9999/new.md
check-ignore exit=0
$ git ls-files .moai/reports/t1760 | wc -l          -> 5
$ git ls-files .moai/reports | wc -l                -> 1792
real-tree E10a exit=0
real-tree E10b exit=0        (stdout empty)
```

Throwaway repository in session scratch with a copy of the real `.gitignore` (base commit `bbbb14105fee0500d89ad6d7a7e971b980071e0e`; tracked, force-added `.moai/reports/t9999/verdict.md`, `.moai/reports/t1760/plan-audit.md`, plus `src/keep.txt`), the real tree untouched; the narrowed E10a and E10b were run unchanged:

```text
clean tree                                       E10a exit=0   E10b stdout empty
new untracked t9999/new.md (ignored)             E10a exit=0   E10b stdout empty   (check-ignore in that repo: .gitignore:235:.moai/reports/*  .moai/reports/t9999/new.md, exit 0)
t9999/verdict.md modified (appended line)        E10a exit=1   E10b " M .moai/reports/t9999/verdict.md"
t9999/verdict.md deleted                         E10a exit=1   E10b " D .moai/reports/t9999/verdict.md"
t9999/forced.md created and `git add -f`         E10a exit=1   E10b "A  .moai/reports/t9999/forced.md"
```

The fix-round-3 throwaway repository (base `1203114`) was re-inspected: no `.gitignore` (`test -f .gitignore` exit 1) and `git check-ignore -v .moai/reports/t9999/new.md` exit 1, which is why its `??` line did not transfer. Two probe pairs were issued in the same turn against the same scratch repository, so the first clean-tree line was re-taken on its own after removing the new file (exit 0, stdout empty).

Claim 2 (O-a): the two stale counts are replaced by statements that name their command and tree; the measured values at the tree above are 5 and 1792.

Claim 3: `awk -f r3_cn4.awk plan.md acceptance.md` after the edits printed `COLLECTED: 4 milestones in plan order (M1 M2 M3 M4), 12 exit bindings, 10 ordering candidates`, exit 0, no `CONFLICT:` line. The candidate count rose from 8 to 10 because the new prose carries ordering words; the ten printed lines were read on screen, and none orders work across milestones except the Definition of Done line for AC-CMM-010 (run-phase tip, before sync), which agrees with M4. The traceability verb with `-v accstate=read` printed `COLLECTED: 11 REQ definitions (acceptance input: read)`, exit 0, no `UNCOVERED:` or `ORPHAN:` line (a first invocation without `-v accstate` printed an empty input state and was discarded).

Claim 4: `mcp__moai__spec_audit` (project_root = this worktree, filter this SPEC) returned `modern_era_clean: 1`, `total_specs: 1`, one INFO `EraAutoDetected` (H-5). ID check on `SPEC-CONTAINMENT-MAPPING-DOC-001` against `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` printed `PASS`.

Gaps: the deliverable document does not exist, so E01 to E09 were not re-run in this round (their commands and recorded red are unchanged); the fix-round-1 and fix-round-2 E10 probes carried in the ledger were not re-run; only the `.moai/reports/*` ignore rule was probed, so the claim that other ignored paths behave the same is a git-behaviour expectation, not an observation; the `.moai/lessons-inbox.jsonl` tool-failure rows were not inspected, though no command was refused in this round; no `audit_multi` or `spec lint` run was made (the independent re-audit is the next step); optional finding O-b (merge of `main` into the card branch) was not measured; the post-commit E10a and E10b run is reported in the hand-off message because a commit cannot cite its own hash here.

Residual risk: a new local file under `.moai/reports/` (any card's directory) or under any other ignored path passes E10a and E10b by design (Known limit 2 in AC-CMM-010); a change under `.moai/reports/t1760/` is invisible by design; if the run phase merges `main` into the card branch, E10a (fixed base) would read red at the run-phase tip (inferred, not tested, carried in `plan.md` section J.1 and J.2); E08 still counts lines and enumerates date shapes, so the shapes listed under its Known limit are outside what the command decides.

## §E.2 Run-phase Evidence

_pending run-phase_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_
