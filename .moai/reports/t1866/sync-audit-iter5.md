auditor-model: claude-sonnet-5-5

# t1866 sync-phase audit, round 5 (HEAD-confirmation delta after absorbing main; Class B, no SPEC)

verdict: PASS
audited_sha: a3dd18c5ac49e0813373e6ba2568c7c0bbe9afae

Overall Verdict: PASS. The PASS at `e6ba15deb0` still holds at the merge HEAD. Main's refactor kept the card's ordering intact, no card test became vacuous or wrong, the GLM effort defaults and the distributed llm.yaml/report.yaml are unchanged by main, and the required codex gate returned pass. No blocking finding.

Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1866`, HEAD `a3dd18c5ac49e0813373e6ba2568c7c0bbe9afae` (merge of main `4f9dba14bf` into the card branch), card base `ae34f268319ce91f028ee692f25c1bd476f9e6e5`. Scope: `git diff e6ba15deb0 HEAD` for the two files both sides touched, plus the greps of question (d).

## Audit plan and cross-model leg

`moai verify audit-plan --project-root <toplevel>` -> `config_status: ok`, `model: codex`, claude gate off, codex gate required, glm gate off, `cross_model_active: true`, `enforced_required: ["codex"]`.

`mcp__moai__audit_multi` called exactly once (`project_root` = toplevel, `target: baseBranch`, no `gates` argument, a focus string). Digest members for the orchestrator's result check:

- overall_verdict: pass
- gate_unmet: absent from the returned object
- plan_source: config
- per_backend_verdicts: backend codex, gate required, verdict pass (no findings, no next_steps)
- audit_receipt: rcpt-639481700c33236cce2a8ede
- build_lag: installed moai binary built from ee981a932, an ancestor of HEAD a3dd18c5a

Receipts produced by THIS audit (complete list): `rcpt-639481700c33236cce2a8ede`. Receipts of rounds 1 to 4 belong to those rounds and are not cited.

This verdict is not final until the orchestrator writes those members to `<toplevel>/.moai/state/audit-plan-result.json` and runs `moai verify audit-plan --project-root <toplevel> --result-file <that path>`.

## Findings

No blocking finding.

- F12 [optional] internal/cli/launcher.go:684 - No card test covers the claim "the GLM drop of auto precedes typed-argv consumption" on the GLM backend directly: a typed `--permission-mode auto` on a GLM launch is refused earlier in glm.go (line 360-363), so the order inside `runLaunchClaude` is protected by code reading and by the Claude-side test `TestLaunch_ExplicitPermissionModeAutoOnCommandLineIsForwardedOnClaude`, not by a GLM-side RED test. Pre-existing from the card (not introduced by the merge); no action required.
- F6 (round 1, cosmetic summary and GLM note) remains open, optional.
- F10, F11 (rounds 3 and 4) remain recorded, optional.

## Dimension scores (flat weighted mode; no `evaluator_mode: hierarchical` observed)

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 93/100 | PASS | 21 TestLaunch_ PASS lines, 0 FAIL, 0 SKIP at the merge HEAD; ordering read at launcher.go:684-734; codex gate pass |
| Security (25%) | 90/100 | PASS | Fail-closed GLM resume (launcher.go:730-734) intact; main's 7 defaults.go lines only add an audit-log path constant |
| Craft (20%) | 90/100 | PASS | settings, wizard, core/project packages ok; gofmt clean; coverage not measured |
| Consistency (15%) | 90/100 | PASS | gofmt -l internal/cli internal/config empty |

Must-pass firewall: Functionality and Security both clear their thresholds.

## Questions (a) to (d)

(a) Order after the merge. In `runLaunchClaude` the sequence at HEAD is: permMode from profile and settings (655-662), `permGLM` (666), default (670-673), the GLM drop `permGLM && permMode == "auto"` -> `permMode = ""` and `syncPermMode` drop (684-690), `cont` from settings (691), THEN `consumed := consumeLaunchArgs(extraArgs)` (702), then `if consumed.PermissionModeSet { permMode = consumed.PermissionMode }` (705-707) and `if consumed.Continue { cont = true }` (708-710), then the fresh-launch sync `!cont` (721) and the resume drop `cont && permGLM && settingsLocalPath != ""` (730-734). So: the GLM drop precedes the typed-argv consumption, a typed `--permission-mode auto` overrides afterwards (keeps its own refusal in glm.go:360-363; on the Claude backend the explicit argv mode wins), and `cont` is final before both the sync and the resume-drop blocks. `consumeLaunchArgs` (launch_view.go:210-297) handles `-b/--bypass` -> `bypassPermissions`, `--permission-mode <v>` and `--permission-mode=<v>`, `-c/--continue`, and breaks at `--` appending the remainder to PassThrough unconsumed. Tests that pin these: `TestLaunch_LegacyBypassWithEmptyProfileModeKeepsBypassAndWritesNoOverride`, `TestLaunch_ExplicitPermissionModeAutoOnCommandLineIsForwardedOnClaude`, `TestLaunch_SettingsChosenModeDoesNotPersistDefaultAuto`, `TestLaunch_GLMResumeFailsWhenStaleAutoCannotBeRemoved` (uses `[]string{"--continue"}`, so it exercises consumed.Continue), the `--continue` / `-c` table (resume_test.go:127-128). No test of the card uses `-b`, `--permission-mode=value` or `--` directly beyond what the above show; those paths are main's SPEC-LAUNCHER-ARGV-001 behavior and are read, not card-tested (see F12 and Gaps).

(b) Vacuous or wrong card tests. None found. All 21 PASS lines include the stale-auto failure test and the positive control with its `no_settings.local.json` subtest; the argv they pass reaches `consumeLaunchArgs`, so the tests run through main's new path, not the removed loop.

(c) GLM effort defaults. `grep -n DefaultGLMEffort internal/config/defaults.go` -> `DefaultGLMEffortLow = "max"` (line 364), `DefaultGLMEffortFable = "high"` (365). `git diff e6ba15deb0 HEAD -- internal/config/defaults.go` is exactly 7 added lines (`DefaultLeaderAnchorAuditRelPath`), no GLM touch. `git diff e6ba15deb0 HEAD --name-only | grep -i 'llm.yaml\|report.yaml\|defaults'` -> only `internal/config/defaults.go`; the distributed llm.yaml and report.yaml are not changed by main.

(d) Other contradictions in main. Files changed on both sides (card base to e6ba15deb0 vs e6ba15deb0 to HEAD): `CHANGELOG.md`, `internal/cli/launcher.go`, `internal/config/defaults.go` only. CHANGELOG diff on main's side adds no line touching the card's entry (the t1866 entry is still present at CHANGELOG.md:41). A `git grep` of non-test Go for `html+md`/`acceptEdits`/`defaultPermissionMode`/`DefaultGLMEffort` at HEAD shows only pre-existing uses (cc.go help text, init.go comment, profile_setup.go, settings/schema.go) that match the card's end state. `git diff e6ba15deb0 HEAD -G'html\+md|acceptEdits|defaultPermissionMode|DefaultGLMEffort'` over non-report files -> empty: main did not touch any of these values.

## Evidence-bearing report

### Claim
1. The ordering the card relies on is intact in `runLaunchClaude` after main's argv refactor.
2. No card test is vacuous or wrong after the refactor.
3. GLM effort defaults and distributed llm.yaml/report.yaml are unchanged by main.
4. Nothing else in main contradicts the card.

### Evidence
- `git rev-parse HEAD` -> `a3dd18c5ac49e0813373e6ba2568c7c0bbe9afae`; `git status --short` -> empty at start.
- `git diff e6ba15deb0 HEAD --stat -- internal/cli/launcher.go internal/config/defaults.go` -> `internal/cli/launcher.go | 142 +++++++-------------------------------------`, `internal/config/defaults.go | 7 +++`, `2 files changed, 30 insertions(+), 119 deletions(-)`.
- `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli -run '^TestLaunch_' -count=1 -v` -> `exit=0`; 21 `--- PASS` lines (13 top-level), 0 `--- FAIL`, 0 `--- SKIP`; `--- PASS: TestLaunch_GLMResumeFailsWhenStaleAutoCannotBeRemoved (0.00s)`; `--- PASS: TestLaunch_GLMResumeWithoutAutoStartsTheClaudeBinary (0.43s)`; `ok  	github.com/modu-ai/moai-adk/internal/cli	1.709s`.
- `go test ./internal/settings ./internal/cli/wizard ./internal/core/project -count=1` -> `ok .../internal/settings 0.786s`, `ok .../internal/cli/wizard 4.629s`, `ok .../internal/core/project 1.941s`.
- `gofmt -l internal/cli internal/config` -> no output.
- `grep -n "DefaultGLMEffortLow\|DefaultGLMEffortFable" internal/config/defaults.go` -> `364: DefaultGLMEffortLow    = "max"`, `365: DefaultGLMEffortFable  = "high"`.
- Code read: launcher.go lines 655-734 (order), launch_view.go lines 210-297 (`consumeLaunchArgs`/`walkLaunchArgs`), glm.go lines 360-363 (`auto mode is not available with GLM (third-party provider)`).
- `audit_multi` -> `overall_verdict: pass`, codex gate required verdict pass, `audit_receipt: rcpt-639481700c33236cce2a8ede`, `plan_source: config`.

### Baseline-attribution
All measurements ran in this session against HEAD `a3dd18c5ac49e0813373e6ba2568c7c0bbe9afae` in the worktree above; `go test` compiles the tree directly. The `moai` binary that served `audit-plan` and `audit_multi` was built from `ee981a932`, a strict ancestor of HEAD (server `build_lag` notice), so those two outputs were produced by an older build of the tool.

### Gaps
- No mutation test of the ordering was run this round (not in the request); the claim that the order matters on the GLM backend rests on code reading (F12).
- `-b`, `--permission-mode=value` and `--` handling were read in `walkLaunchArgs`, not run through dedicated tests.
- Whole-package runs of `internal/cli`, `internal/config`, `internal/template`, `internal/web` and the wide `^(TestLaunch|...)` run were not performed (forbidden or lane-observed: 463 PASS; this audit did not re-observe that figure).
- Coverage not measured; the built binary was not run.
- `Binding run conditions`: the card has no SPEC; the heading was not searched this round.
- No command was refused by a guard in this audit.
- The orchestrator's `audit_multi` result check has not run.

### Residual-risk
- Order protection on the GLM side rests on reading, not a RED test (F12).
- Main's `consumeLaunchArgs` now also consumes `-m` / `--model=`, which the card does not rely on.
- Existing installs keep old defaults after `moai update`.

## Iteration history

- Round 1: FAIL at `2eb7daee9b` (`sync-audit.md`), receipt `rcpt-265156e724a2ad7c21131413`, findings F1 to F6.
- Round 2: FAIL at `bece378100` (`sync-audit-iter2.md`), receipt `rcpt-b0662a488c1234ba9fdeb487`; F1 to F5 repaired; F7 blocking, F8 optional.
- Round 3: FAIL at `02bb6804fb` (`sync-audit-iter3.md`), receipt `rcpt-9a86f07371c17bb56bf21032`; F7, F8 fixed; F9 blocking via codex gate, F10 optional.
- Round 4: PASS at `e6ba15deb0` (`sync-audit-iter4.md`), receipt `rcpt-f8934fe53bf583fb3b0a9f75`; F9 closed, F10 recorded.
- Round 5 (this file): PASS at `a3dd18c5ac` (merge of main into the card), receipt `rcpt-639481700c33236cce2a8ede`; PASS of round 4 confirmed at the merge HEAD; optional F12 added.

## Recommendations

1. Land the card. F6 and F12 stay optional follow-ups.
