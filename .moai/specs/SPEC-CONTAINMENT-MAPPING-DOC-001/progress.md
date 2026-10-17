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

## §E.2 Run-phase Evidence

_pending run-phase_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_
