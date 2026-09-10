# t343 — run-phase RED baseline, ledger re-measured on the CURRENT tree

The plan-phase ledger (`acceptance.md` §D.0, E-01..E-17) was measured on `a6bbbf82b`.
The run-phase tree is `15453140a` (`origin/develop` absorbed). A figure measured on
another tree is not a baseline for this one, so every RED-bearing entry was re-run
here, before any M1/M2/M3/M4 edit, and its output observed.

- HEAD: `15453140a`
- branch: `WT-red-now-threshold`
- worktree: `.claude/worktrees/t343`
- card: t343

## Re-measured entries

```
$ grep -c "RED-now cell content" .claude/rules/moai/development/verification-completeness.md
0
exit: 1

$ grep -c -e tense -e mood -e counterfactual -e "future.sense" .claude/rules/moai/development/verification-completeness.md .claude/agents/moai/plan-auditor.md
.claude/rules/moai/development/verification-completeness.md:0
.claude/agents/moai/plan-auditor.md:0
exit: 1

$ grep -c "regression-guard" .claude/rules/moai/development/verification-completeness.md
0
exit: 1

$ grep -c "MP-8" .claude/agents/moai/plan-auditor.md
0
exit: 1

$ ls internal/spec/red_now_cell_test.go
ls: internal/spec/red_now_cell_test.go: No such file or directory
exit: 1

$ grep -c "AC-6:" .claude/agents/moai/plan-auditor.md
0
exit: 1

$ grep -c "MOAI-REDNOW-BEGIN" .claude/agents/moai/plan-auditor.md
0
exit: 1

$ ls internal/spec/testdata/red_now/
ls: internal/spec/testdata/red_now/: No such file or directory
exit: 1

$ grep -c "MOAI-REDNOW-BEGIN" internal/template/templates/.claude/agents/moai/plan-auditor.md
0
exit: 1

$ grep -rl "red_now" internal/template/templates/
(empty stdout)
exit: 1
```

## Verdict

Every RED cited by a release-blocking criterion (E-01, E-03, E-04, E-05, E-06,
E-07, E-08, E-09) reproduces on `15453140a` with the same output and the same
exit code recorded at `a6bbbf82b`. The two regression-guard greens (E-02, E-11)
also reproduce unchanged.

## Gaps

- E-10, E-12..E-17 were not re-run here. E-12/E-13/E-14 count rows in this SPEC's
  own `acceptance.md`, which this run-phase does not edit; E-15/E-16/E-17 measure
  `internal/kanban` and `internal/hook/evidence_writer.go`, neither of which is in
  this SPEC's scope envelope. They are unmeasured on `15453140a` and are therefore
  neither confirmed nor refuted here.
- No corpus sweep over the other SPEC directories was run; this file measures only
  the entries this SPEC's criteria cite.
