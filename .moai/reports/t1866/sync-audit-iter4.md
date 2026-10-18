auditor-model: claude-sonnet-5-5

# t1866 sync-phase audit, round 4 (final delta; Class B, no SPEC)

verdict: PASS
audited_sha: e6ba15deb0cd332f41ef2030bf37623a3b55af48

Overall Verdict: PASS. F9 is closed: I reproduced the codex "child started before the removal check" overlay myself and the strengthened test fails on it. F10 is correctly recorded. No blocking finding; the required codex gate returned pass.

Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1866`, HEAD `e6ba15deb0cd332f41ef2030bf37623a3b55af48`, card base `ae34f268319ce91f028ee692f25c1bd476f9e6e5`. Scope: `git diff 02bb6804fb..HEAD` (3 files: test file, CHANGELOG one sentence, progress.md).

## Audit plan and cross-model leg

`moai verify audit-plan --project-root <toplevel>` -> `config_status: ok`, `model: codex`, claude gate off, codex gate required, glm gate off, `cross_model_active: true`, `enforced_required: ["codex"]`.

`mcp__moai__audit_multi` called exactly once (`project_root` = toplevel, `target: baseBranch`, no `gates` argument). Digest members for the orchestrator's result check:

- overall_verdict: pass
- gate_unmet: absent from the returned object
- plan_source: config
- per_backend_verdicts: backend codex, gate required, verdict pass (no findings)
- audit_receipt: rcpt-f8934fe53bf583fb3b0a9f75
- build_lag: the installed moai binary was built from ee981a932, an ancestor of HEAD e6ba15deb

Receipts produced by THIS audit (complete list): `rcpt-f8934fe53bf583fb3b0a9f75`. Earlier rounds' receipts belong to rounds 1 to 3 and are not cited.

This verdict is not final until the orchestrator writes those members to `<toplevel>/.moai/state/audit-plan-result.json` and runs `moai verify audit-plan --project-root <toplevel> --result-file <that path>`.

## Findings

- F9 CLOSED. The failing-path test now pins a marker-writing stub claude (`pinMarkerClaude`), asserts error naming file and `auto`, marker absent, exec seam not called, settings byte-unchanged. A positive control proves the mechanism.
- F10 [optional] RECORDED. A malformed `settings.local.json` stops a GLM resume with an error naming the file; the CHANGELOG sentence and progress record say so; no test (acceptable, optional).
- F11 [optional, not blocking] progress.md: the note that the exit-0 stub "covers ordering without the exit-1 variant" is true for the two mutations I ran (see (a)); it is an argument, and an exit-1 stub would add nothing the marker does not already give. No action required.
- F6 (round 1, cosmetic summary and GLM note) remains open, optional.

## Dimension scores (flat weighted mode; no `evaluator_mode: hierarchical` observed)

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 93/100 | PASS | Product code byte-identical since the last round; 21 TestLaunch_ PASS; required codex gate pass |
| Security (25%) | 90/100 | PASS | Fail-closed GLM resume unchanged; no product change |
| Craft (20%) | 90/100 | PASS | Test now pins the child-start ordering (RED on my overlay), positive control present; coverage not measured |
| Consistency (15%) | 90/100 | PASS | gofmt clean; helper style matches the file |

Must-pass firewall: Functionality and Security both clear their thresholds.

## Questions (a) to (e)

(a) Can the strengthened test pass if the child started before the removal check? No. Launcher read: the first child on a resume is `tryCmd.Run()` (launcher.go ~905), reached after the removal check at ~754 returns. The stub's marker is in `t.TempDir()` of `pinMarkerClaude`, a directory separate from the chmod-0555 `.claude`, so the stub can write it. `EnvClaudeBin` is pinned, so the child is the stub. The `err == nil` check is `t.Error` (not Fatal) but `t.Error` marks the test failed, so the guarantee holds; the later assertions still run. I ran the codex overlay myself in an extracted copy outside the tree (`git archive HEAD` into a scratchpad dir): the removal failure is recorded, `tryCmd.Run()` runs, then the error is returned. Result: `launcher_permission_resume_test.go:362: the session was started (the claude binary ran) with a stale auto still on disk`, `--- FAIL: TestLaunch_GLMResumeFailsWhenStaleAutoCannotBeRemoved`. That overlay returns the right error, so only the marker assertion catches it: the marker is what closes F9.

(b) Positive control meaningful? Yes. Both subtests (`defaultMode plan`, `no settings.local.json`) require nil from the resume AND the marker to exist. Under my mutated launcher the control still PASSED (`--- PASS: TestLaunch_GLMResumeWithoutAutoStartsTheClaudeBinary`), i.e. it is independent of the mutation and proves the marker mechanism observes a `--continue` start.

(c) `git diff 02bb6804fb HEAD -- internal/cli/launcher.go | wc -c` -> `0`; card-wide launcher diff unchanged at 91 insertions, 10 deletions. `git status --short` -> empty (the untracked `backup-rejection-*` dirs are not reported, left untouched).

(d) CHANGELOG sentence true: "A malformed `.claude/settings.local.json` now also stops a GLM resume with an error naming the file (before, the launch continued with a warning)." It matches `dropAutoDefaultModeFromSettingsLocal` returning the `readSettingsMap` error and the `cont && permGLM` return. Only that sentence changed in the CHANGELOG diff.

(e) Progress record: the RED output quoted matches the form I observed (my overlay differs slightly: the lane's mutation also returned nil, so it printed both lines; mine printed only the marker line, consistent with its Gaps note that the codex overlay was not the mutation used). "13 top-level PASS" -> I counted 13 `^--- PASS` (21 including subtests). The exit-0 note is accurate. Nothing false or overstated found. The "honest line" accepting the lane's mistake is correct.

## Evidence-bearing report

### Claim
1. F9 closed; the test fails against a launcher that starts the child before the removal check.
2. The positive control is meaningful and mutation-independent.
3. No product code changed; tree clean.
4. CHANGELOG and progress record are accurate.

### Evidence
- `git rev-parse HEAD` -> `e6ba15deb0cd332f41ef2030bf37623a3b55af48`; `git status --short` -> empty.
- `git diff 02bb6804fb..HEAD --stat` -> `progress.md | 17 ++`, `CHANGELOG.md | 2 +-`, `launcher_permission_resume_test.go | 72 +++-`; launcher.go diff byte count `0`.
- `go test ./internal/cli -run '^TestLaunch_' -count=1 -v` (kanban env unset) -> exit 0; 21 `--- PASS`, 0 `--- FAIL`, 0 `--- SKIP`; `--- PASS: TestLaunch_GLMResumeFailsWhenStaleAutoCannotBeRemoved (0.00s)`; control and both subtests PASS; `ok  github.com/modu-ai/moai/.../internal/cli 1.897s`.
- `go test ./internal/settings ./internal/cli/wizard ./internal/core/project -count=1` -> three `ok` lines (0.927s, 5.914s, 3.433s).
- `gofmt -l internal/cli` -> no output.
- Mutation (scratch copy, not the tree): `--- FAIL: TestLaunch_GLMResumeFailsWhenStaleAutoCannotBeRemoved (0.24s)` with `launcher_permission_resume_test.go:362: the session was started (the claude binary ran) with a stale auto still on disk`; `--- PASS: TestLaunch_GLMResumeWithoutAutoStartsTheClaudeBinary (0.46s)`.
- `audit_multi` -> `overall_verdict: pass`, codex gate required verdict pass, `audit_receipt: rcpt-f8934fe53bf583fb3b0a9f75`, `plan_source: config`.

### Baseline-attribution
Measured in this session against HEAD `e6ba15deb0`. The mutation ran on a `git archive HEAD` extraction of that commit with a two-hunk edit to launcher.go. The `moai` binary that served `audit-plan` and `audit_multi` was built from `ee981a932`, a strict ancestor of HEAD, so those two outputs were produced by an older build; `go test` compiles the tree directly.

### Gaps
- A first mutation command (extraction into a scratchpad directory preceded by `rm -rf`) was refused by the permission system: `Permission to use Bash with command S=... rm -rf $S; mkdir -p $S && ... git archive HEAD has been denied.` I did not work around the guard; I re-ran the extraction into a fresh, differently named empty scratchpad directory without the removal, a different command, not a bypass.
- Whole-package runs of `internal/cli`, `internal/config`, `internal/template`, `internal/web` were not run (forbidden; no slot acquired).
- Coverage not measured; the built binary was not run.
- The F10 malformed-settings stop is read, not run.
- `Binding run conditions`: `grep -n "Binding run conditions" .moai/reports/t1866/progress.md` returned nothing; the card has no SPEC, so none are recorded.
- The orchestrator's `audit_multi` result check has not run.

### Residual-risk
- The test depends on a non-root user (the precondition `t.Fatal`s otherwise, so it fails loudly rather than passing).
- A hand-written `auto` is removed on a GLM resume (accepted product behavior).
- Existing installs keep old defaults after `moai update`.

## Iteration history

- Round 1: FAIL at `2eb7daee9b` (`sync-audit.md`), receipt `rcpt-265156e724a2ad7c21131413`, findings F1 to F6.
- Round 2: FAIL at `bece378100` (`sync-audit-iter2.md`), receipt `rcpt-b0662a488c1234ba9fdeb487`; F1 to F5 repaired; F7 blocking, F8 optional.
- Round 3: FAIL at `02bb6804fb` (`sync-audit-iter3.md`), receipt `rcpt-9a86f07371c17bb56bf21032`; F7, F8 fixed; F9 blocking via codex gate, F10 optional.
- Round 4 (this file): PASS at `e6ba15deb0`, receipt `rcpt-f8934fe53bf583fb3b0a9f75`; F9 closed, F10 recorded.

## Recommendations

1. Land the card; F6 stays an optional follow-up.
