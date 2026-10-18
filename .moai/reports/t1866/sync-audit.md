auditor-model: claude-sonnet-5-5

# t1866 sync-phase audit (Class B, no SPEC)

verdict: FAIL
audited_sha: 2eb7daee9b0be3baccf9072729d357dc38ea1963

Overall Verdict: FAIL (score about 0.76, harmonic of 70 / 80 / 78 / 80)
Cause: the required codex gate returned `fail` (receipt rcpt-265156e724a2ad7c21131413), and I independently confirmed by reading the code that both of its P2 findings describe real paths on which a GLM session can still end up in auto. Functionality is the must-pass dimension that fails.

Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1866`, HEAD 2eb7daee9b0be3baccf9072729d357dc38ea1963, base ae34f268319ce91f028ee692f25c1bd476f9e6e5, 28 files.

## Audit plan and cross-model leg

`moai verify audit-plan --project-root <toplevel>` returned: `config_status: ok`, `model: codex`, claude gate off, codex gate required, glm gate off, `cross_model_active: true`, `enforced_required: [codex]`.

`mcp__moai__audit_multi` (called once, `project_root` = toplevel, `target: baseBranch`, no `gates` argument). Digest members for the orchestrator's result check:

- overall_verdict: fail
- gate_unmet: not present in the returned object (the field was absent; overall_verdict is fail through the required-backend verdict, `residual_risk_note: "required-backend FAIL: codex"`)
- plan_source: config
- per_backend_verdicts: backend codex, gate required, verdict fail
- audit_receipt: rcpt-265156e724a2ad7c21131413
- build_lag reported by the server: the installed moai binary is built from ee981a932, an ancestor of this tree's HEAD; CLI/MCP output describes older code than this tree.

This verdict is not final until the orchestrator writes these members to `<toplevel>/.moai/state/audit-plan-result.json` and `moai verify audit-plan --project-root <toplevel> --result-file <that path>` returns `convergence_check.ok`. With `overall_verdict: fail` it cannot.

## Findings

- F1 [P2] [blocking] internal/cli/launcher.go (resume gate `if !cont && settingsLocalPath != ""`, about line 737, with `syncPermissionModeToSettingsLocal`, about line 1175) - A stale launcher-written `permissions.defaultMode: auto` survives into a GLM resume. A Claude launch with an unset profile now writes `defaultMode: auto` into `.claude/settings.local.json` on every fresh launch. A later GLM launch with `--continue` (or `DO_CLAUDE_CONTINUE=true`) skips the sync step entirely, so the file keeps `auto`, and the new GLM carve-out (which only clears the in-memory `syncPermMode`) never reaches the disk. The GLM launcher refuses `--permission-mode auto` precisely because auto is unsupported there, so the GLM session is left with the unsupported mode through the settings channel. This answers question (a) "is anything in the GLM path still able to receive auto from a profile or settings value": yes, through the settings file on resume. Confidence high on the path (read from code: line 737 gate, line 1175 writer); the runtime effect of Claude Code reading `defaultMode: auto` under a GLM model was not observed (Gap). Required fix: on the GLM backend, remove a `defaultMode` equal to `auto` from settings.local.json even when `cont` is true (or write the GLM-supported mode explicitly on the resume path), and add a test that launches Claude unset, then GLM with `--continue`, and asserts the file carries no `auto`.
- F2 [P2] [blocking] internal/cli/launcher.go:635-645 (step 4a backend resolution) with `applyGLMMode` at launcher.go:278-300 - Backend detection re-reads llm.yaml from disk (`resolveGLMBackendForLaunch` calls `template.IsGLMBackend`, which keys on `llm.team_mode`/`llm.mode`). `applyGLMMode` only warns when `persistTeamMode(root, "glm")` fails and continues, having already exported the GLM environment through `setGLMEnv`. If the config write fails (read-only config directory, `team_mode` still empty), the launch has GLM credentials in its environment but `glmBackend == false`, so `launchDefaultPermissionMode(false)` returns `auto` and `--permission-mode auto` is appended at launcher.go:779-780 to a GLM session. The mode the command already knows (`mode == "glm"` in `runUnifiedLaunch`, and `MOAI_LAUNCH_PROVIDER`) is not used for this decision. Codex reported this reproduced on a read-only config temp project; I confirmed the code path by reading and did not execute the reproduction (Gap). The same disk re-read pre-dated the card for model and effort selection, but the card is the first to hang a permission flag on it. Required fix: carry the command's resolved mode into `runLaunchClaude` (or OR it with the disk result) so the GLM carve-out does not depend on the save succeeding; add a test with an unwritable llm.yaml.
- F3 [P3] [optional] internal/cli/launcher_permission_default_test.go - The test file covers unset/stored-auto on Claude and GLM (4 tests, all pass, and the GLM cases would fail if the backend were not detected, so they are not vacuous). It does not cover the three paths the lead named: `DO_CLAUDE_BYPASS=true` with an empty profile mode (bypassPermissions, no settings.local override), a `--permission-mode auto` typed on the command line, and the resume path. By reading, legacyBypass is correct: `legacyBypass` is set, `permMode` stays `bypassPermissions`, the default is not applied, `syncPermMode` takes `prefs.PermissionMode` (empty), so the override is removed, not written. That reading is not backed by a test; only `mcp_doctor_coverage_test.go` touches `DO_CLAUDE_BYPASS` and it does not launch. Required fix: add the three tests (F1 and F2 tests would cover resume and provider).
- F4 [P3] [optional] internal/core/project/initializer.go:44 and :256-258, internal/cli/profile_setup.go:148 and :234 - Comments now state something false. initializer.go:44 says `(empty -> html+md default)` and :258 says `Empty resolves to html+md`, while the code at the fallback now resolves to `artifact`; :256 still says the template default is html+md. profile_setup.go:148 says the wizard "defaults to acceptEdits and normalizes that back to \"\" on save", and :234 says the empty permission mode "pre-selects acceptEdits"; the wizard now pre-selects auto and nothing normalizes it. Required fix: update the four comments.
- F5 [P3] [optional] internal/cli/launcher.go (`syncPermMode` derivation, about lines 668-672) - When the profile stores no mode but settings.local.json carries `DO_CLAUDE_PERMISSION_MODE` (say `plan`), the session runs `plan` via argv, but `syncPermMode` falls to `auto` on the Claude backend, so `defaultMode: auto` is persisted to settings.local.json. Before the card the same case removed the override. This is a persisted-posture widening for later direct `claude` runs in that project, for a user who did choose a mode (outside the profile). It is in the neighbourhood of what the CHANGELOG states but is not literally "profile that never chose a mode". Required fix: do not default `syncPermMode` when `settings["DO_CLAUDE_PERMISSION_MODE"] != ""`.
- F6 [P4] [optional] internal/cli/profile_setup.go:480 and the stderr note at launcher.go:678-681 - The profile summary shows `auto` for an unset profile regardless of backend, and the "not available on the GLM backend" note prints on every GLM launch for a stored auto. Cosmetic.

No defect found in: the ordering of the moved backend block (step 4a precedes every consumer of `glmBackend`, `glmModels`, `glmTierEffort`, `sessionRoot`; the step-6 block now only reads them; compiles and the launch tests pass), the explicit-argv path (argv parsing at launcher.go:703-727 overwrites `permMode` after the default and the GLM drop, so a typed `--permission-mode auto` keeps its own refusal in glm.go:352), the schema.go change (`Persist.Normalize` has no non-test caller: `git grep 'Normalize('` over non-test code returned only `internal/settings/schema.go:298`'s own function, matching the lead's statement), and the GLM effort constants (`DefaultGLMEffortLow = "max"`, `DefaultGLMEffortFable = "high"` in defaults.go:362-365, matching llm.yaml and the CHANGELOG).

## Dimension scores (flat weighted mode; no `evaluator_mode: hierarchical` observed)

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 70/100 | FAIL | F1, F2: GLM can still receive auto through the settings file (resume) and through argv when the team_mode save fails; codex required gate `fail`. 12 targeted tests pass. |
| Security (25%) | 80/100 | PASS | No widening beyond the CHANGELOG for profiles that never chose a mode; the explicit-bypass path unchanged; F5 persists auto for a settings-chosen mode. No secrets, no new input surface. |
| Craft (20%) | 78/100 | PASS | New tests are non-vacuous for the paths they cover (F3 lists the uncovered paths); four false comments (F4). Coverage was not measured (Gap). |
| Consistency (15%) | 80/100 | PASS | One stale-comment set (F4); the distributed template, Go constants, wizard label, web description and the schema constant agree. |

Must-pass firewall: Functionality failed, so the overall verdict is FAIL regardless of the other dimensions.

## Evidence-bearing report

### Claim
1. The launcher behaves as described on the Claude backend and for an unset profile on the GLM backend at argv level.
2. A GLM session can still receive `auto` (F1 via settings.local.json on resume, F2 via argv when the backend flag is lost).
3. The shipped-default edits are consistent across Go constants, the distributed templates, the wizard and the web fallback.
4. No silent migration of existing stored values.

### Evidence
- `git rev-parse --show-toplevel` -> `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1866`; `git rev-parse HEAD` -> `2eb7daee9b0be3baccf9072729d357dc38ea1963`.
- `git diff --stat ae34f2683..HEAD` -> `28 files changed, 544 insertions(+), 84 deletions(-)`.
- `moai verify audit-plan ...` -> `"cross_model_active": true`, codex `"gate": "required"`, `"enforced_required": ["codex"]`.
- `mcp__moai__audit_multi` -> `"overall_verdict":"fail"`, codex `"verdict":"fail"` with two P2 findings (launcher.go:643 and launcher.go:682), `"audit_receipt":"rcpt-265156e724a2ad7c21131413"`, `"plan_source":"config"`.
- `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli -run 'TestLaunch_|TestInitialProfileResult|TestShippedDefaults|TestProfileSetup' -count=1` -> `ok  	github.com/modu-ai/moai-adk/internal/cli	1.535s`; the `-v` run showed 12 `--- PASS` lines and no `--- FAIL` / `--- SKIP` lines.
- `go test ./internal/settings ./internal/cli/wizard ./internal/core/project -count=1` -> `ok .../internal/settings 0.763s`, `ok .../internal/cli/wizard 4.972s`, `ok .../internal/core/project 2.393s`.
- Code reading: launcher.go resume gate `if !cont && settingsLocalPath != ""` followed by `syncPermissionModeToSettingsLocal(settingsLocalPath, syncPermMode)`; `applyGLMMode` ends `if err := persistTeamMode(root, "glm"); err != nil { ... "Warning: failed to persist team mode" }` and continues; `resolveGLMBackendForLaunch` returns `template.IsGLMBackend(llm)` from the on-disk llm section; `IsGLMBackend` keys on `cfg.TeamMode` / `cfg.Mode`.
- Stale-text greps: `git grep -n 'html+md'` over internal/cmd/pkg/.claude/CHANGELOG, `git grep -n 'acceptEdits'` over internal (non-test). Hits that are still true: skill fallback text, wizard question comments, the settings schema comment. Hits that are false: the four in F4. `.moai/config/sections/report.yaml:2` (`format: html+md`) is this repository's own deployed config, not a distributed template, and is not a defect. `internal/template/templates/.claude/settings.json.tmpl:446` ships `"defaultMode": "default"` and `generic-patterns-guide.md` states "Template default: unset"; neither contradicts the launcher override, which is a settings.local.json value.
- `internal/config/testdata/shipped_key_inventory.yaml` lists `llm.glm.effort.low` with `class: W`, `evidence: reader`, with no value pinned; no stale default there.

### Baseline-attribution
All measurements ran in this session against HEAD 2eb7daee9b0be3baccf9072729d357dc38ea1963 in the worktree above. The `moai` binary that produced `audit-plan` and `audit_multi` output was built from ee981a932, a strict ancestor of the tree HEAD (the server's own `build_lag` notice says so); the `go test` runs compile the tree directly and are not affected by that lag. The audit-plan output describes the config of this tree, but the CLI's own code is older than the tree.

### Gaps
- No whole-package run of internal/cli, internal/config, internal/template, internal/web, internal/core/project (the lead forbade the heavy suites; the three small packages above were run).
- No landing re-measure (known gap in progress.md).
- The en/ja/zh docs-site trees were not searched (known gap); only the repository's own non-docs-site files were grepped.
- The web model/effort empty-option labels were left as is by the card and not reviewed.
- F1 and F2 were confirmed by code reading and by codex's reported reproduction; I did not execute either reproduction, because adding a test file would edit the tree.
- Not observed: what Claude Code does when `permissions.defaultMode: auto` is read in a GLM-env session (error, silent fallback, or accepted). F1's severity rests on the card's own premise that GLM refuses auto.
- Coverage percentages were not measured; the Craft score is not coverage-backed.
- No command was refused by a guard in this audit, and no guard record was touched.
- The orchestrator's `audit_multi` result check (`--result-file`) has not run; this audit is not final until it does.

### Residual-risk
- Even after F1 and F2 are fixed, the launcher persists `defaultMode: auto` into settings.local.json for every unset-profile Claude launch, so a direct `claude` run in that project after the launcher exits inherits auto. This is stated in the CHANGELOG, but it is a standing posture change.
- Existing installs keep the old defaults after `moai update` (the update test proves the preserve behavior), so the new defaults reach only fresh inits; that is consistent with no silent migration, but it means the CHANGELOG change affects new projects only for effort and report format.
- `normalizePermissionMode` is unwired; if a later card wires it, a stored `auto` would be written as an empty value, which the launcher resolves to auto on Claude and to nothing on GLM. That interaction is untested.

## Iteration history

Single sync-audit pass for this tree (HEAD 2eb7daee9b). Prior lane-side evidence: `.moai/reports/t1866/card-review.md` (advisory, codex, pass, no findings, static review only) did not surface F1 or F2.

## Binding run conditions

No `progress.md` heading named `Binding run conditions` was relied on for this verdict; the card has no SPEC and I did not find that heading when reading the lane's evidence record scope given in the dispatch. Not independently grep-verified (Gap).

## Recommendations

1. Fix F1 and F2 with tests (resume with stale auto under GLM; unwritable llm.yaml with GLM env present), then re-run the scoped tests and a delta re-audit limited to F1-F2.
2. Optionally fold in F3-F5 in the same commit (F4 is four comment edits; F5 is a one-condition change).
