auditor-model: claude-sonnet-5-5

# t1866 sync-phase audit, round 2 (delta; Class B, no SPEC)

verdict: FAIL
audited_sha: bece378100542cfd91b0eb1cee5454a551131455

Overall Verdict: FAIL (Functionality must-pass fails through the required codex gate)
Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1866`, HEAD `bece378100542cfd91b0eb1cee5454a551131455`, card base `ae34f268319ce91f028ee692f25c1bd476f9e6e5`. Delta scope: `git diff 2eb7daee9..HEAD` (6 files, 385 insertions, 13 deletions) plus a regression read of the card diff.

Summary: F1, F2, F3, F4 and F5 from round 1 are repaired and I could not break the repair on the paths I tried (a) to (e). The required codex gate nevertheless returned `fail` with one new P2 on the repair itself (F7 below). I confirmed the path by reading the code. It is the same invariant (a GLM session must not carry `auto`), reached through a failed cleanup instead of a skipped one. Under the required gate this cannot be PASS.

## Audit plan and cross-model leg

`moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1866` returned `config_status: ok`, `model: codex`, claude gate off, codex gate required, glm gate off, `cross_model_active: true`, `enforced_required: ["codex"]`.

`mcp__moai__audit_multi` was called exactly once (`project_root` = toplevel, `target: baseBranch`, no `gates` argument, a focus string). Digest members for the orchestrator's result check:

- overall_verdict: fail
- gate_unmet: absent from the returned object (overall_verdict is `fail` through the required-backend verdict, `residual_risk_note: "required-backend FAIL: codex"`)
- plan_source: config
- per_backend_verdicts: backend codex, gate required, verdict fail
- audit_receipt: rcpt-b0662a488c1234ba9fdeb487
- build_lag reported by the server: the installed moai binary was built from ee981a932, an ancestor of this tree's HEAD. CLI and MCP output describe older code than this tree.

Receipts produced by THIS audit: `rcpt-b0662a488c1234ba9fdeb487` (one; `audit_multi` was not called a second time). The round-1 receipt `rcpt-265156e724a2ad7c21131413` belongs to round 1 and is not cited here.

This verdict is not final until the orchestrator writes those members to `<toplevel>/.moai/state/audit-plan-result.json` and `moai verify audit-plan --project-root <toplevel> --result-file <that path>` passes. With `overall_verdict: fail` it cannot.

## Findings

- F7 [P2] [blocking] internal/cli/launcher.go:753-757 - The new stale-auto drop is warn-only. `dropAutoDefaultModeFromSettingsLocal` failing (for example a read-only `.claude` directory or an unwritable `settings.local.json`) prints `Warning: failed to clear stale auto permission mode` and the launch continues, so a GLM resume starts with `permissions.defaultMode: auto` still on disk. Codex reproduced it on `bece378100` with a temporary overlay test (`handoff provider=glm permission-arg="" persisted-default="auto"`, `launch returned err=<nil> launched=true`). I did not execute that reproduction; by reading, the gate is `if cont && permGLM && settingsLocalPath != ""` and the error branch only writes to stderr, so the path exists. The same fail-open shape already existed for the fresh-launch sync (`syncPermissionModeToSettingsLocal`), and it needs a double condition (GLM, resume, an unwritable settings file), so likelihood is low. It is blocking because it is the same invariant the card exists to hold, the required gate fails on it, and the fix is small. Confidence high on the path, not observed live. Required fix: on the GLM path return an error when the stale `auto` cannot be removed, or pass a GLM-supported mode explicitly on argv; add a regression test with an unwritable `settings.local.json` (the test must assert the failure was induced, as the F2 test does).
- F8 [P4] [optional] CHANGELOG.md (t1866 entry) and launcher.go:750-752 - The text says a GLM resume removes "a launcher-written `auto`". The code removes any `defaultMode` equal to `auto`, including one the user wrote by hand, because the file carries no provenance. Required fix (optional): say "a `permissions.defaultMode: auto`" instead of "launcher-written".
- F6 (round 1, cosmetic summary and GLM note) was deliberately left out by the lane. Still open, optional, not a blocker.

## Dimension scores (flat weighted mode; no `evaluator_mode: hierarchical` observed)

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 78/100 | FAIL | F1, F2, F5 behave as claimed (11 `TestLaunch_*` tests, 17 PASS lines including subtests); F7 leaves one GLM-carries-auto path; codex required gate `fail`. |
| Security (25%) | 82/100 | PASS | The drop touches only the exact value `auto`; no widening; F7 is a fail-open on an already-restricted double condition. |
| Craft (20%) | 85/100 | PASS | New tests are non-vacuous (precondition assertion on the induced save failure; no skips); coverage not measured (Gap). |
| Consistency (15%) | 86/100 | PASS | Four false comments corrected; the template `report.yaml` ships `format: artifact`, matching the fixed comments. |

Must-pass firewall: Functionality fails, so the overall verdict is FAIL regardless of the other dimensions.

## Questions (a) to (f), answered with what I observed

(a) Provider marker. `config.EnvMoaiLaunchProvider` is set in exactly one place, `launchClaudeForProvider` (launcher.go:231-248), which `runUnifiedLaunch` reaches at line 222 with `mode`. That is the only caller of `launchClaude`, and `launchClaude` is the only caller of `runLaunchClaude` (via `launchClaudeDefault`). Every entry that reaches `runLaunchClaude` passes through it: `moai glm` calls `unifiedLaunch(profileName, "glm", ...)` (glm.go:373); `resolveMode("")` returns `claude`. The set is unconditional (`os.Setenv`), so a stale process value is overwritten with `claude` on a Claude launch, and restored by the defer when the seam returns. `permGLM` therefore reads `glm` only when the command selected `glm`. Not observed: the `moai glm -f/-l` managed path (`managedFactoryLaunchFunc`, launcher.go:892) was read, not run; it is reached after `runLaunchClaude` has composed argv. Result: no hole found.

(b) Drop placement. The drop sits after the argv loop (launcher.go:702-737), where `cont` is final (settings plus `-c`/`--continue`), and before step 6 and the exec at launcher.go:923. It runs under `mutateSettingsLocal` (locked, atomic), reads first and returns without writing unless `defaultMode == "auto"`, and deletes only that key (and an emptied `permissions` object). `TestLaunch_GLMResumeKeepsOtherDefaultModes` confirms `plan` survives. A user-written `auto` is removed too (F8).

(c) Remaining GLM-receives-auto paths. Profile or settings value on a fresh launch: dropped (`permGLM && permMode == "auto"`), and the fresh sync writes `""`. Persisted file on resume: dropped by the new block, except when that write fails (F7). Failed `persistTeamMode` save: covered by `permGLM`. Argv typed on the command line keeps its own refusal in glm.go. `.claude/settings.json` project `defaultMode` is outside the card.

(d) Tests. `TestLaunch_GLMLaunchWithUnwritableConfigDoesNotReceiveAuto` calls `persistTeamMode(root, "glm")` before the launch and `t.Fatal`s when it succeeds, so it cannot pass for the wrong reason as root; it fails instead of skipping. `id -u` printed `501` (non-root), so the induced failure is real here. Without `permGLM` the test would see `glmBackend == false` (llm.yaml absent) and an `auto` argv, so it is not vacuous (the lane records the RED run; I did not rerun a mutated build, since I may not edit the tree). No `t.Skip` in the file; the run shows 0 SKIP. Bypass, typed `--permission-mode auto`, Claude resume, and settings-chosen mode each assert argv and the file.

(e) Comments and CHANGELOG. The four F4 comments now match the code (template `report.yaml:2` is `format: artifact`; the wizard pre-selects auto). The CHANGELOG sentence is true except for the word "launcher-written" (F8). The comment at launcher.go:674-675 says "the override is removed" while a non-empty profile mode would be synced, which is loose but not false for the case it describes.

(f) Earlier findings. F1 repaired (resume drop), F2 repaired (`permGLM`), F3 tests added, F4 comments fixed, F5 repaired (`settings["DO_CLAUDE_PERMISSION_MODE"] != ""`). F6 deliberately left. No earlier finding unaddressed except F6.

## Evidence-bearing report

### Claim
1. The repair commit closes round-1 F1 to F5 as claimed.
2. A GLM launch cannot receive `auto` through a profile, a settings value, a persisted file, or a failed team_mode save, except when the persisted file cannot be rewritten (F7).
3. The new tests are non-vacuous and non-skipping.
4. The comment and CHANGELOG edits are accurate apart from F8.

### Evidence
- `git rev-parse --show-toplevel` -> `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1866`; `git rev-parse HEAD` -> `bece378100542cfd91b0eb1cee5454a551131455`; `git status --short` printed nothing before this verdict file was written.
- `git diff --stat 2eb7daee9..HEAD` -> 6 files, `385 insertions(+), 13 deletions(-)`.
- `moai verify audit-plan ...` -> `"cross_model_active": true`, codex `"gate": "required"`, `"enforced_required": ["codex"]`.
- `mcp__moai__audit_multi` -> `"overall_verdict":"fail"`, codex `"verdict":"fail"`, one P2 at launcher.go:755, `"audit_receipt":"rcpt-b0662a488c1234ba9fdeb487"`, `"plan_source":"config"`.
- `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli -run '^TestLaunch_' -count=1 -v` -> `exit=0`; 17 `--- PASS`, 0 `--- FAIL`, 0 `--- SKIP`; top-level lines include `--- PASS: TestLaunch_GLMRemovesStaleAutoFromSettingsLocal (0.18s)`, `--- PASS: TestLaunch_GLMLaunchWithUnwritableConfigDoesNotReceiveAuto (0.01s)`, final `ok  	github.com/modu-ai/moai-adk/internal/cli	1.722s`.
- `go test ./internal/settings ./internal/cli/wizard ./internal/core/project -count=1` -> `ok .../internal/settings 1.088s`, `ok .../internal/cli/wizard 6.752s`, `ok .../internal/core/project 3.303s`.
- `id -u` -> `501`.
- `gofmt -l internal/cli internal/core/project` -> no output.
- `git grep -n 'format:' -- internal/template/templates/.moai/config/sections/report.yaml` -> `report.yaml:2:  format: artifact`.
- `git grep -n 'EnvMoaiLaunchProvider' -- '*.go'` (non-test) -> one setter (launcher.go:233, restored 239-241), readers at launcher.go:666, factory_quota.go:128, mcp_audit_multi.go:122, hook/gateway_guard.go:10, hook/session_start_record.go:106.
- Code reading: launcher.go:663-690 (`permGLM`), 745-757 (sync and drop), 1185-1209 (`dropAutoDefaultModeFromSettingsLocal`).

### Baseline-attribution
All measurements ran in this session against HEAD `bece378100542cfd91b0eb1cee5454a551131455` in the worktree above. `go test` compiles the tree directly. The `moai` binary that produced the `audit-plan` and `audit_multi` output was built from ee981a932, a strict ancestor of the tree HEAD (the server's `build_lag` notice says so), so those two tool outputs describe the config of this tree but were produced by older tool code.

### Gaps
- Whole-package runs of `internal/cli`, `internal/config`, `internal/template`, `internal/web` were not run (the lead forbade them; no slot acquired). Only `^TestLaunch_` and three small packages ran.
- F7 was confirmed by code reading and codex's reported overlay reproduction; I did not execute it, because that needs a test file and I may not edit the tree.
- The F2 test was not re-run against a build with `permGLM` reverted; its non-vacuity rests on reading plus the lane's recorded RED.
- What Claude Code does when it reads `permissions.defaultMode: auto` under a GLM environment was not observed.
- The managed factory launch path (`moai glm -f/-l`, `managedFactoryLaunch`) was read, not run.
- Coverage percentages were not measured; the Craft score is not coverage-backed.
- No landing re-measure; the en/ja/zh docs-site trees were not searched (both carried over from round 1).
- No command was refused by a guard in this audit, with one exception: my first Bash call combined two git commands and a command substitution, and the worktree guard answered `This session is isolated in the worktree /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1866, but this command names git in a form too complex to verify that it stays inside the worktree. Refusing to run it`. I split it into plain commands and passed the toplevel as a literal path; no guard record was touched.
- The orchestrator's `audit_multi` result check (`--result-file`) has not run.

### Residual-risk
- Even with F7 fixed, an unset-profile Claude launch persists `defaultMode: auto`, so a direct `claude` run afterwards inherits it (stated in the CHANGELOG, a standing posture change).
- A hand-written `auto` is removed on a GLM launch (F8); a user who wants `auto` for Claude and runs GLM in the same project loses it.
- Existing installs keep old defaults after `moai update`, so the new defaults reach fresh inits only.

## Iteration history

- Round 1: FAIL at `2eb7daee9b0be3baccf9072729d357dc38ea1963` (`.moai/reports/t1866/sync-audit.md`), score about 0.76, receipt `rcpt-265156e724a2ad7c21131413`, findings F1 to F6.
- Round 2 (this file): delta audit of repair commit `bece378100` at `bece378100542cfd91b0eb1cee5454a551131455`: F1 to F5 repaired; new blocking F7 from the required codex gate; verdict FAIL, receipt `rcpt-b0662a488c1234ba9fdeb487`.

## Binding run conditions

The card has no SPEC. I read `.moai/reports/t1866/progress.md` (diff since round 1 only) and found no heading named `Binding run conditions`. Not independently grep-verified over the whole file (Gap).

## Recommendations

1. Fix F7: make a failed stale-`auto` removal on the GLM path an error (or pass an explicit supported mode), with a regression test that asserts the induced write failure. Then request a delta re-audit limited to F7.
2. Optionally reword the CHANGELOG and the comment (F8) and fold in F6.
