# t570 — lane-orchestrator independent verification

Second, independent observation of the run-phase claims. Measured by the lane
orchestrator, not by the implementing agent, against tree `a700e496c` in
`.claude/worktrees/t570`.

## Claim

The guard added by `c08d91168` actually fails when the conversion it guards is
removed. This re-derives the implementing agent's RED rather than reading its log.

## Evidence

Mutant M-1 injected by hand at `internal/cli/doctor_codex.go:837`
(`statPath = fromConfigPath(e.Path, configPathSeparator)` → `statPath = e.Path`),
tree confirmed clean (`git status --porcelain | wc -l` → `0`) immediately before
the injection.

```
$ go test ./internal/cli/ -run 'TestCodexStaleSkillFinding_(AbsoluteStatTargetIsConvertedForm|ExtendedLengthStatTargetIsConvertedForm|ClassificationPrecedesConversion)' -timeout 1200s -v
=== RUN   TestCodexStaleSkillFinding_AbsoluteStatTargetIsConvertedForm
    doctor_codex_path_guard_test.go:52: stat target = "/Users/u/skills/probe/SKILL.md", want the converted form "\\Users\\u\\skills\\probe\\SKILL.md"
    doctor_codex_path_guard_test.go:55: stat target = "/Users/u/skills/probe/SKILL.md" — the DECLARED form: the conversion is not applied
--- FAIL: TestCodexStaleSkillFinding_AbsoluteStatTargetIsConvertedForm (0.00s)
=== RUN   TestCodexStaleSkillFinding_ExtendedLengthStatTargetIsConvertedForm
    doctor_codex_path_guard_test.go:93: stat target = "//?/C:/Users/u/skills/probe/SKILL.md", want the converted extended-length form "\\\\?\\C:\\Users\\u\\skills\\probe\\SKILL.md"
--- FAIL: TestCodexStaleSkillFinding_ExtendedLengthStatTargetIsConvertedForm (0.00s)
=== RUN   TestCodexStaleSkillFinding_ClassificationPrecedesConversion
--- PASS: TestCodexStaleSkillFinding_ClassificationPrecedesConversion (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.843s
```

Reverted with `git checkout -- internal/cli/doctor_codex.go`; both post-conditions
re-measured in the same invocation:

```
$ git diff -- internal/cli/doctor_codex.go | wc -c
       0
$ git status --porcelain | wc -l
       0
```

Unmutated run, same selector, same tree, immediately before the injection:

```
--- PASS: TestCodexStaleSkillFinding_AbsoluteStatTargetIsConvertedForm (0.00s)
--- PASS: TestCodexStaleSkillFinding_ExtendedLengthStatTargetIsConvertedForm (0.00s)
--- PASS: TestCodexStaleSkillFinding_ClassificationPrecedesConversion (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.914s
```

Swept count is visible in both runs: three `=== RUN` lines, so neither `ok` nor
`FAIL` was returned over an empty selection.

## Baseline-attribution

Both runs: tree `a700e496c` on branch `WT-codex-doctor-guard`, worktree
`.claude/worktrees/t570`, this session, darwin. Base `a4855f0b2` == `origin/develop`
as read at card start.

Scope re-measured independently: `git diff a4855f0b2 -- internal/cli/doctor_codex.go`
is 0 bytes; `git diff --stat 076847aba..HEAD` touches one non-evidence source file,
`internal/cli/doctor_codex_path_guard_test.go`.

## Gaps — what this verification did NOT observe

- M-2, M-3 and M-4 were NOT re-injected here. Their records are the implementing
  agent's, read from the committed logs, not re-derived.
- The full-package run (`rc=0`, `internal/cli` 611.738s) was NOT re-executed; the
  figure is read from `full-package.log`, not measured in this run.
- Windows behaviour of the `\\?\` prefix is asserted by no run on any tree here.
  It is the stated motivation for the family, not a measured property.

## Residual risk

M-4 (`strings.ReplaceAll(e.Path, "/", "\\")` in place of the seam call) is MISSED
by this guard and is recorded as such in `run-evidence.md`. The guard pins the
converted VALUE under a pinned separator, not the USE of the `fromConfigPath` /
`configPathSeparator` seam. A separator-identity companion run would close it;
that is outside this card.
