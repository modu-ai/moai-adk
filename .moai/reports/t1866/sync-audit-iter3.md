auditor-model: claude-sonnet-5-5

# t1866 sync-phase audit, round 3 (delta; Class B, no SPEC)

verdict: FAIL
audited_sha: 02bb6804fbcb7762b3a7345e80bb44848db016ca

Overall Verdict: FAIL. The required codex gate returned `fail` (one P2 on the new test), and the instruction for this audit is that a required gate that returns fail is a FAIL. My own reading agrees the product code is correct and F7/F8 are fixed; the one open finding is test strength, and I only partly confirm it (see F9).

Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1866`, HEAD `02bb6804fbcb7762b3a7345e80bb44848db016ca`, card base `ae34f268319ce91f028ee692f25c1bd476f9e6e5`. Scope: `git diff bece378100..HEAD` (4 files, 77 insertions, 5 deletions) plus a regression read of the surrounding code.

## Audit plan and cross-model leg

`moai verify audit-plan --project-root <toplevel>` returned `config_status: ok`, `model: codex`, claude gate off, codex gate required, glm gate off, `cross_model_active: true`, `enforced_required: ["codex"]`.

`mcp__moai__audit_multi` was called exactly once (`project_root` = toplevel, `target: baseBranch`, no `gates` argument, a focus string). Digest members for the orchestrator's result check:

- overall_verdict: fail
- gate_unmet: absent from the returned object (the fail comes through the required-backend verdict, `residual_risk_note: "required-backend FAIL: codex"`)
- plan_source: config
- per_backend_verdicts: backend codex, gate required, verdict fail
- audit_receipt: rcpt-9a86f07371c17bb56bf21032
- build_lag reported by the server: the installed moai binary was built from ee981a932, an ancestor of this tree's HEAD 02bb6804f.

Receipts produced by THIS audit: `rcpt-9a86f07371c17bb56bf21032` (one; `audit_multi` was not called a second time). Earlier rounds' receipts (`rcpt-265156e724a2ad7c21131413`, `rcpt-b0662a488c1234ba9fdeb487`) belong to rounds 1 and 2 and are not cited here.

This verdict is not final until the orchestrator writes those members to `<toplevel>/.moai/state/audit-plan-result.json` and runs `moai verify audit-plan --project-root <toplevel> --result-file <that path>`. With `overall_verdict: fail` that check cannot pass.

## Findings

- F9 [P2] [blocking via the required codex gate; product-code impact none] internal/cli/launcher_permission_resume_test.go:345 - The test's "no exec" assertion observes the wrong seam for a resume. On `--continue` the first child start is `tryCmd.Run()` at launcher.go:905, and `execOrSpawnClaudeFunc` (the seam the test stubs, `called`) is reached only at launcher.go:924 after a resume exit-1 fallback. So `called == false` cannot detect a child started on the resume path. Codex reports an overlay that moved the removal check after the child start still passed. I could not execute that overlay (read-only audit), and with the shared `writeExecutable` stub (`exit 0`) a misordered check would return `nil` at launcher.go:906 before reaching the check, which the test's `err == nil` assertion would catch, so ordering is protected indirectly. The direct assertion the card brief asked for ("assert no exec") is nevertheless not proven. Required fix: have the stub claude binary (a separate writable temp dir) write a marker file when run, assert no marker on the failing path, and add a control test that a GLM resume with no `auto` and one with no `settings.local.json` creates the marker and returns nil. Use an exit-1 stub for the failing-path case so the assertion does not rest on the exit-0 shortcut.
- F10 [P4] [optional] internal/cli/launcher.go:755 - `dropAutoDefaultModeFromSettingsLocal` also returns the `readSettingsMap` parse/read error, so a malformed `.claude/settings.local.json` now stops a GLM resume (before, the launch continued with a warning). The error names the file, so it is fixable, and refusing is consistent with "cannot verify no auto". Mention it in the CHANGELOG if the lane wants it documented; not required.
- F6 (round 1, cosmetic summary and GLM note) remains deliberately open, optional.

Round-2 findings: F7 fixed (see (a) below), F8 fixed (see (c)).

## Dimension scores (flat weighted mode; no `evaluator_mode: hierarchical` observed)

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 86/100 | FAIL | F7 behavior verified correct by reading and test; required codex gate `fail` (F9) forces FAIL |
| Security (25%) | 88/100 | PASS | The drop touches only the exact value `auto`; the failure now fails closed on GLM only |
| Craft (20%) | 78/100 | PASS | New test proves the induced failure and error return, but not the child-start ordering (F9); coverage not measured |
| Consistency (15%) | 88/100 | PASS | Comment and CHANGELOG now agree with the code; gofmt clean |

Must-pass firewall: Functionality is held at FAIL by the required gate result, so the overall verdict is FAIL regardless of the other dimensions.

## Questions (a) to (d)

(a) Placement and normal resumes. The `return` is at launcher.go:754-757, after the argv loop (so `cont` is final) and before step 6, the `--continue` child `tryCmd.Run()` (905), the managed-launch divert (889) and the exec seam (924). Side effects that precede it: `profile.SyncToProjectConfig` (idempotent, pre-existing), the GLM env exported by the caller (process-level, the error exits the launcher), and no settings.local.json write (the drop only writes on `auto`, and on failure writes nothing, since the atomic temp-file swap failed). `dropAutoDefaultModeFromSettingsLocal` returns nil when the file is absent (`readSettingsMap` returns an empty map on `os.IsNotExist`, settings.go:26-27) and when `defaultMode` is not `auto`, so a normal GLM resume with no file or no `auto` still starts; I verified this by reading, and `TestLaunch_GLMResumeKeepsOtherDefaultModes` passes. There is no dedicated test for the no-file resume (Gap, part of F9's fix).

(b) Test non-vacuity. Proven by the test body and a run: `id -u` is 501 (non-root); the test chmods `.claude` to 0o555, has a creation-probe precondition that `t.Fatal`s (does not skip) when the directory is still writable, restores 0o755 in `t.Cleanup`, asserts the error is non-nil and names the file path and `auto`, and asserts the file is byte-identical. Without the drop it would hit `t.Fatal("... returned nil")`. As root it would fail on the precondition, not pass. Not proven: that no child started (F9).

(c) Text truth. The comment (launcher.go:750-753) and the CHANGELOG sentence no longer say "launcher-written"; the CHANGELOG now reads "a GLM resume (`--continue`) also removes a `permissions.defaultMode: auto` from `.claude/settings.local.json` (a failed removal stops the GLM launch with an error), while a Claude resume still does not rewrite that file", which matches the code (the error is returned only under `cont && permGLM`; Claude resumes and the fresh-launch sync stay warn-only at launcher.go:745-749).

(d) Contradiction of the rest of the card diff. None found: the `permGLM` computation, the fresh-launch sync, and `TestLaunch_GLMLaunchWithUnwritableConfigDoesNotReceiveAuto` are untouched by this commit and the whole `TestLaunch_` run stays green.

## Evidence-bearing report

### Claim
1. F7 is fixed: a failed stale-`auto` removal on a GLM resume returns an error naming the file before any child start; Claude resumes and the fresh sync remain warn-only.
2. F8 is fixed in the comment and the CHANGELOG.
3. The new test proves the induced failure, the error, and byte-unchanged file; it does not directly prove the child-start ordering (F9).
4. Nothing else in the card diff is contradicted.

### Evidence
- `id -u` -> `501`; `git rev-parse HEAD` -> `02bb6804fbcb7762b3a7345e80bb44848db016ca`; `git status --short` -> clean before this file was written.
- `git diff --stat bece378100..HEAD` -> `.moai/reports/t1866/progress.md | 15 +++`, `CHANGELOG.md | 2 +-`, `internal/cli/launcher.go | 9 ++--`, `internal/cli/launcher_permission_resume_test.go | 56 +++`.
- `moai verify audit-plan ...` -> `"cross_model_active": true`, codex `"gate": "required"`, `"enforced_required": ["codex"]`.
- `mcp__moai__audit_multi` -> `"overall_verdict":"fail"`, codex `"verdict":"fail"`, one P2 at launcher_permission_resume_test.go:345, `"audit_receipt":"rcpt-9a86f07371c17bb56bf21032"`, `"plan_source":"config"`.
- `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli -run '^TestLaunch_' -count=1 -v` -> `exit=0`; 18 `--- PASS`, 0 `--- FAIL`, 0 `--- SKIP`; `--- PASS: TestLaunch_GLMResumeFailsWhenStaleAutoCannotBeRemoved (0.00s)`; `ok  	github.com/modu-ai/moai-adk/internal/cli	1.471s`.
- `go test ./internal/settings ./internal/cli/wizard ./internal/core/project -count=1` -> `ok .../internal/settings 0.716s`, `ok .../internal/cli/wizard 5.338s`, `ok .../internal/core/project 2.758s`.
- `gofmt -l internal/cli` -> no output.
- Code reading: launcher.go:745-758 (sync and drop), 889-924 (managed divert, `tryCmd.Run()` at 905, exec seam at 924), 1190-1209 (`dropAutoDefaultModeFromSettingsLocal`); settings.go:22-37 (`readSettingsMap`), 67-97 (`mutateSettingsLocal`); claude_binary_test.go:22-28 (stub is `#!/bin/sh\nexit 0`).

### Baseline-attribution
All measurements ran in this session against HEAD `02bb6804fbcb7762b3a7345e80bb44848db016ca` in the worktree above; `go test` compiles the tree directly. The `moai` binary that served `audit-plan` and `audit_multi` was built from `ee981a932`, a strict ancestor of HEAD (server `build_lag` notice), so those two outputs describe this tree's config but were produced by an older build.

### Gaps
- Whole-package runs of `internal/cli`, `internal/config`, `internal/template`, `internal/web` were not run (forbidden for this audit; no slot acquired).
- Codex's misordering overlay was not reproduced (read-only audit); my statement that the exit-0 stub makes ordering regress visibly is by reading, not by execution.
- No test exists for a GLM resume with no `settings.local.json`; the no-op was verified by reading only.
- A malformed `settings.local.json` on GLM resume (F10) was read, not run.
- Coverage was not measured.
- The orchestrator's `audit_multi` result check has not run.
- No command was refused by a guard in this audit except my first Bash call, which combined git commands and a command substitution; the worktree guard answered `This session is isolated in the worktree /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1866, but this command names git in a form too complex to verify that it stays inside the worktree. Refusing to run it`. I split it into plain separate commands; nothing was worked around.

### Residual-risk
- Product code is correct as read, so the residual risk is a future refactor that moves the removal check after the resume child start without a test noticing (F9).
- A hand-written `auto` is removed on a GLM resume; a user who wants `auto` for Claude and runs GLM in the same project loses it.
- Existing installs keep old defaults after `moai update`.

## Iteration history

- Round 1: FAIL at `2eb7daee9b0be3baccf9072729d357dc38ea1963` (`.moai/reports/t1866/sync-audit.md`), receipt `rcpt-265156e724a2ad7c21131413`, findings F1 to F6.
- Round 2: FAIL at `bece378100542cfd91b0eb1cee5454a551131455` (`.moai/reports/t1866/sync-audit-iter2.md`), receipt `rcpt-b0662a488c1234ba9fdeb487`; F1 to F5 repaired; blocking F7, optional F8.
- Round 3 (this file): delta audit of `02bb6804fb`; F7 and F8 fixed; new F9 (test strength, blocking via the required codex gate) and optional F10; verdict FAIL, receipt `rcpt-9a86f07371c17bb56bf21032`.

## Binding run conditions

The card has no SPEC. I did not grep `.moai/reports/t1866/progress.md` for a `Binding run conditions` heading in this round (Gap).

## Recommendations

1. Fix F9 in the test only: a marker-writing stub claude in a separate writable temp dir, an exit-1 stub for the failing path, and control tests (no `auto`, no file) that create the marker and return nil. Then request a delta re-audit limited to F9.
2. Optionally document F10 in the CHANGELOG.
