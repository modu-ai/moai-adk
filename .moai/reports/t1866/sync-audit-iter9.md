auditor-model: claude-sonnet-5-5
verdict: PASS-WITH-DEBT
audited_sha: 723ab083521080f925928ceec7f384c333722804

# t1866 sync audit, round 9 (delta audit of the reduced card)

Score: Functionality 92, Security 95, Craft 88, Consistency 85. Overall PASS-WITH-DEBT: codex required gate passed (receipt rcpt-63171f237583bf8da0ec8916); the only defects are stale prose in progress.md (MINOR/SHOULD-FIX, non-blocking).

audit_multi result digest (orchestrator to check): overall_verdict=pass, gate_unmet=none, plan_source=config, per_backend_verdicts: codex gate=required verdict=pass. This verdict is not final until the orchestrator's `moai verify audit-plan --result-file` check passes.

## Claim

C1. Scope matches the keep list and nothing from the remove list remains. C2. Kept items (K1-K4) behave as before the reduction and are pinned by tests that would fail on regression. C3. No leftover references to removed symbols; CHANGELOG is accurate; progress receipt ids are correct; no template regeneration is needed. C4. Touched packages build and vet clean.

## Evidence

R1 launcher:
- `git diff 4f9dba14bf HEAD --stat -- internal/cli/launcher.go` -> no output (empty).
R2 deleted tests:
- `ls internal/cli/launcher_permission_default_test.go internal/cli/launcher_permission_resume_test.go` -> `No such file or directory` for both.
R3 schema/wizard:
- `git diff 4f9dba14bf HEAD --stat` lists `internal/settings/schema_sections.go` (a report-format comment only), NOT `schema.go`, `accessors_test.go` or `profile_setup_nested_test.go`, so those equal main. `grep -n defaultPermissionMode internal/settings/schema.go` -> `304:const defaultPermissionMode = "acceptEdits"`. `profile_setup.go` diff adds only `defaultModel = "sonnet[1m]"`, `defaultEffortLevel = "medium"` and the two empty-value fallbacks; `defaultPermissionMode = "acceptEdits"` unchanged. `TestPermissionModeNormalizeAcceptEdits` (profile_setup_nested_test.go:83) and `TestNormalizePermissionMode` (accessors_test.go:115) exist.
R4 CHANGELOG: `git diff 4f9dba14bf HEAD -- CHANGELOG.md` is one added entry covering (1) GLM effort low=max/fable=high, (2) report format artifact, (3) profile wizard `sonnet[1m]` and `medium`; no `auto`, launcher, GLM-resume or malformed-settings text.
Leftover grep: `grep -rnE 'launchDefaultPermissionMode|dropAutoDefaultModeFromSettingsLocal|profileModeUnset|acceptEditsPermissionMode|DefaultGLMEffort' --include='*.go' --include='*.yaml' --include='*.md' . | grep -v '^./.moai/reports'` -> only `DefaultGLMEffort*` hits (defaults.go, schema_sections.go:175-188, launch_session_defaults.go:123-143, tests); zero hits for the four removed symbols.
Template: `git diff 4f9dba14bf HEAD --stat -- internal/template` -> `glm_tier_defaults_test.go`, `templates/.moai/config/sections/llm.yaml` (low "max", fable "high"), `templates/.moai/config/sections/report.yaml` (`format: artifact`). The templates are read through the embedded tree; the template test below compares the embedded file with the Go constants and passes, so no generated file needs regeneration.
Receipts in progress.md vs files: round 6 `rcpt-813169891395590312ae33f6` and `rcpt-ef9584ba0990325e17cfe714` (iter6.md:84,33), round 7 `rcpt-c68710bc814236dc0a67251f` (iter7.md:70), round 8 `rcpt-9833ef82b66be72a725eb4d3` (iter8.md:88): all four match the progress.md Reduction section (line 278).
Builds and tests (each prefixed with the `unset MOAI_KANBAN_*` trio for tests):
- `go build ./...` -> no output.
- `go vet ./internal/cli ./internal/settings ./internal/config ./internal/core/project ./internal/web ./internal/template` -> no output.
- `go test ./internal/settings ./internal/cli/wizard ./internal/core/project ./internal/config -count=1` -> `ok .../settings 0.466s`, `ok .../cli/wizard 4.861s`, `ok .../core/project 2.358s`, `ok .../config 6.112s`.
- `go test ./internal/cli -run '^(TestProfileSetup|TestPermissionMode|TestInit_|TestUpdate_|TestInitialProfileResult_|TestLaunch)' -count=1` -> `ok  github.com/modu-ai/moai-adk/internal/cli  124.447s`.
- `go test ./internal/web -run 'GLMEffortTier|Schema' -count=1` -> `ok  .../internal/web 1.081s`.
- `go test ./internal/template -run 'GLMTierDefaults' -count=1 -v` -> `--- PASS: TestGLMTierDefaultsTemplateMatchesGoDefaults`.
- `gofmt -l internal/cli internal/settings internal/config internal/core internal/web` -> no output.
Test strength for K items: K1 `defaults_test.go:403-404` literals `"max"` / `"high"`, session_defaults_test, glm_tier_defaults_test, web glm_tier_test; K2 wizard `questions_test.go`/`wizard_test.go` default `artifact` and `initializer_test.go` report.format; K3 `TestInitialProfileResult_ShippedLaunchDefaults` uses literals `sonnet[1m]`/`medium` and `StoredLaunchChoicesWin`; K4 `shipped_defaults_test.go` unchanged by the reduction (grep for Skip: none). The only assertion removed in the reduction is the profile permission-default `auto` assertion (R3, by design). The revert commit `723ab08352` stat: launcher.go, two deleted tests, profile_setup.go, its tests, schema.go, accessors_test.go, CHANGELOG, progress.md only.
GLM-protection: with launcher.go equal to main and the schema/wizard permission default back to `acceptEdits`, no card code touches the permission flag for any backend. A new profile can store `auto` only via the pre-existing picker value list (`schema.go:251` contains "auto", unchanged from main), exactly as before the card; the GLM launch path is byte-identical to main, aside from the intended K1 default `DefaultGLMEffortLow`/`Fable` values read by `launch_session_defaults.go` (an intended behavior change).
Codex gate: `mcp__moai__audit_multi` (one call) -> `"overall_verdict":"pass"`, codex `"gate":"required","verdict":"pass"`, no findings, `"audit_receipt":"rcpt-63171f237583bf8da0ec8916"`, `"plan_source":"config"`.

## Baseline-attribution

Tree HEAD `723ab083521080f925928ceec7f384c333722804` (branch `WT-10-18-class`), clean at start. Comparison base for card scope `4f9dba14bf` (absorbed main) and `ae34f2683`. Judging build: the moai MCP server/binary `ee981a932` (built 2026-10-15), an ancestor of HEAD per the tool's `build_lag` notice, so moai CLI/MCP output (audit-plan, audit_multi) was produced by an older build than the tree; Go build/vet/test used the tree's sources directly via the Go toolchain, so those results are not subject to that lag.

## Gaps

- No whole-package run of `internal/cli`, `internal/web`, `internal/template`, `internal/config` was done for web/template/cli (named patterns only, as instructed); `internal/config`, `internal/settings`, `internal/cli/wizard`, `internal/core/project` ran whole.
- No guard refusal on the Go commands. One compound git+moai command at the start was refused by the worktree guard ("too complex to verify"); I split it into plain commands.
- Not run: mutation check that a regressed default fails each K test (judged by reading literal assertions); no live Claude/GLM launch.
- `git log --no-merges ae34f2683..HEAD --grep=t1866` used to separate the card from main; scope judged mainly through `git diff 4f9dba14bf HEAD --stat` (31 files, all card or report files).
- The moai binary lags HEAD (see Baseline-attribution); the audit-plan verb output matched the expected plan shape, so the lag did not affect gating.

## Residual-risk

- GLM users who never set a tier effort now launch the `low` tier with `max` effort and the `fable` tier with `high` (intended K1 change; costs/latency shift).
- The init wizard now pre-selects `artifact`; where the Artifact tool is absent the report skill falls back to `html+md` (documented in CHANGELOG, not re-observed).
- The follow-up card t1870 re-introduces the Claude `auto` default; it must re-solve the F14/F15/F17 class of defects.

## Findings

- F1 [MINOR] [optional] `.moai/reports/t1866/progress.md:81-85` ("Corrections made during the card") still describes `auto` -> empty schema normalization and `TestPermissionModeNormalizeDefaultAuto`, which no longer exist, and is not marked superseded; progress.md:93 (Gaps) mentions "the launcher change was observed through the exec seam", also stale. R4 asked that earlier auto sections be marked superseded; these two blocks were missed. Required fix: add a one-line "Superseded by Reduction" note under both.
- F2 [MINOR] [optional] `.moai/reports/t1866/progress.md:99-101` Residual-risk is marked superseded for auto bullets, fine; but Gaps line 89-92 ("Whole-package runs observed...") predates the reduction and could mislead; no action needed beyond F1's note.
- No BLOCKING or SHOULD-FIX findings. Verified: R1 empty diff, R2 files absent, R3 restored, R4 CHANGELOG accurate, receipts correct, build/vet/tests green.

## Iteration history

- Round 1-5: see earlier files (round 4 and 5 PASS before the auto default exceptions).
- Round 6: FAIL, receipts rcpt-813169891395590312ae33f6 and rcpt-ef9584ba0990325e17cfe714. Round 7: FAIL, rcpt-c68710bc814236dc0a67251f. Round 8: FAIL, rcpt-9833ef82b66be72a725eb4d3.
- Round 9 (this file): PASS-WITH-DEBT at `723ab08352`, receipt rcpt-63171f237583bf8da0ec8916; debt = F1/F2 stale progress prose.

AUDIT-VERDICT: PASS-WITH-DEBT spec=t1866 receipts=rcpt-63171f237583bf8da0ec8916
