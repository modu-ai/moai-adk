# progress.md — SPEC-MODEL-OPUS55-001

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifacts authored 2026-09-23 by manager-spec (card t1089), Tier M, 4 files (spec.md, plan.md, acceptance.md, progress.md), worktree `.claude/worktrees/t1089`, branch `WT-opus-55-default`, authored at HEAD `6e75b74db` (dispatched base `17f71a13d`).
- SPEC ID self-check (executed Bash, `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`): dispatched `SPEC-MODEL-OPUS-5-5-001` → `FAIL` (segment `5` does not start with a letter); corrected `SPEC-MODEL-OPUS55-001` → `PASS`; collision check `ls .moai/specs | grep -c 'OPUS55'` → `0`.
- Frontmatter: 12 canonical fields; `phase: "v3.2.0"` (release target, not a stage token); `status: draft`.
- Decisions: (a) `claude-opus-5` joins the deprecated-id map → `opus`; (b) no template `effortLevel` injection; (c) docs-site + READMEs split to a separate card — plan.md §C.
- Evidence persisted: `.moai/reports/t1089/p4-rednow-6e75b74db.txt` (probe P4 v1, 147 lines); `.moai/reports/t1089/p4v2-rednow-6e75b74db.txt` (probe P4 v2, 153 lines).
- Iteration-1 plan-audit (2026-09-23): FAIL 0.78 (Tier M 0.80), 2 blocking (D1, D2) + 13 other; report `.moai/reports/t1089/plan-audit-iter1.md`. Repaired in spec.md 0.1.1 scoped to D1–D15 only (no new REQs; REQ/AC counts stay 16/16).
- Open items: 0 open clarification markers.
- plan_status: audit-ready
- plan_complete_at: 2026-09-23
- Implementation Kickoff Approval: NOT requested, NOT granted.

## §E.2 Run-phase Evidence

Run by manager-develop (cycle_type=tdd) on 2026-09-23 in worktree `.claude/worktrees/t1089`, branch `WT-opus-55-default`. Start tree `f30d07ea7` (develop `08113ff0f` absorbed after the plan audit). Commits: M1 `06c5133dd`, M2 `7100be63b`, M3 `43aa64cba`, M4 = the commit carrying this section. Evidence files are under `.moai/reports/t1089/run-*.log`. The decisive lines are quoted below so the claims do not depend on a scratch path.

### E.2.0 Start-of-run drift check (plan probes re-run on `f30d07ea7`)

| Probe | Plan value (`6e75b74db`) | Measured on `f30d07ea7` | Drift |
|---|---|---|---|
| P4 v2 | 153 lines | 153 lines; `diff` against the sorted `p4v2-rednow-6e75b74db.txt` returned rc=0 (identical). Saved as `run-p4v2-start-f30d07ea7.txt` | none |
| P6 | 11 lines | 11 lines, the same file:line set as plan §B.7 | none |
| P5 | no output, rc=1 | no output, rc=1 | none |
| P7 | `0` | `0` | none |
| B.5 SAME pairs | 6 SAME | 6 SAME (`cmp -s`) | none |
| AC-011 | rc=1 | `grep -n -i effort …settings.json.tmpl`: no output, rc=1. The absorbed `settings.json.tmpl` change added no effort key | none |

Drift the probes did not see. There are three items.
1. **A third label-drift guard.** `TestModelPolicyLabels_AgreeWithProfileMatrix` (`internal/cli/profile_setup_schema_options_test.go:142`) derives `"Opus " + TrimPrefix(id, "claude-opus-")`, which gives `Opus 5-5`. Plan §E K7 named only two guards. It was fixed in M4 with the same hyphen→dot derivation.
2. **Two `internal/web` failures already present on the absorbed tree.** They are `TestDataI18nKeysSubsetOfDictionary` and `TestI18nKeySetParity`, and both involve the `factory_msg_*` MCP tool keys added by develop `internal/mcp/catalog.go` with no i18n entries.
3. **Two `internal/cli` failures already present on the absorbed tree.** They are `TestGLM_FactoryWorkerEntry` and `TestSessionPIDStamp_NotSetFromHooks`. One lane-environment leak also appeared: `TestCodexSpawn_RealAssemblyThroughStubTmux`. Evidence is in E.2.3.

### E.2.1 RED evidence (captured before GREEN)

- M1: `go test -count=1 -run 'TestModelOpusAliasTargetsOpus55|TestModelDeprecatedOpusIDsNormalizeToAlias' -v ./internal/template/` returned `model_policy_test.go:60: opus alias = "claude-opus-5", want "claude-opus-5-5"` and `FAIL` (`run-m1-red.log`). The AC-OP55-002 RED was taken after the alias moved and before the deprecated row existed: `model_policy_test.go:73: ModelAliasFromCanonicalID("claude-opus-5") = "claude-opus-5", want opus` and `glm_slot_test.go:39: GLMSlotForModel("claude-opus-5") = "", want "high"` (`run-m1-red2.log`).
- M2: guards and new tests were run against the old labels (`run-m2-red.log`):
  - `TestGetProfileText_OpusAliasValues` failed 20× ("should reference Opus 5.5").
  - `TestGetProfileText_RecommendationMarkers` failed 8× (for example `EffortLevelMedium "medium - 균형" lacks the (권장) marker`).
  - `TestModelPolicyDescsAgreeWithProfileMatrix` failed 12×.
  - `TestModelOptLabelsEnglishUnified` failed with `0 occurrences of "f.model.opt.opus[1m]": "Opus 5.5 (Recommended)"`.
  - `TestEffortOptRecommendationLabels` failed 8×.
- M4: the third guard went RED on the full `internal/cli` run: `profile_setup_schema_options_test.go:184: lang="en" label "High - Opus 5.5 (high~medium) + …" should name "Opus 5-5"`.

### E.2.2 AC matrix (tree `43aa64cba` + the M4 test fix, unless a row says otherwise)

| AC | Command (abridged; full form in acceptance.md) | Decisive output | Status |
|---|---|---|---|
| 001 | `grep -nE 'ModelIDOpus55 = "claude-opus-5-5"\|"opus":[[:space:]]+ModelIDOpus55' internal/template/model_policy.go` | `51:const ModelIDOpus55 = "claude-opus-5-5"` / `78:	"opus":     ModelIDOpus55,` | PASS |
| 002 | `grep -nE '"claude-opus-5":[[:space:]]+"opus",[[:space:]]*// superseded' internal/template/model_policy.go` | `98:	"claude-opus-5":     "opus", // superseded by ModelIDOpus55`; test `TestModelDeprecatedOpusIDsNormalizeToAlias` PASS | PASS |
| 003 | `grep -rnw 'ModelIDOpus5' internal --include='*.go'` | no output, rc=1 | PASS |
| 004 | P4 v2 (single `find … -exec awk`) | no output. Positive control `awk … internal/template/model_policy.go` → `1` | PASS |
| 005 | anchor/heading greps + `go test -run 'TestRegistrySyncGuard\|TestRegistrySyncMirrorsIdentical' -v ./internal/constitution/` | `opus-55…` `:2`/`:2`; `opus-5-48…` `:0`/`:0`; heading `:1`/`:1`; `--- PASS: TestRegistrySyncGuard`, `--- PASS: TestRegistrySyncMirrorsIdentical` (`run-ac005.log`) | PASS |
| 006a | P7 awk | `2` | PASS |
| 006b | `grep -cE 'Opus 5\.5[^\|]*medium\|medium[^\|]*Opus 5\.5' <9 files>` | constitution ×2 `:1`, agent-authoring ×2 `:1`, model-policy ×2 `:4`, dynamic-workflows ×2 `:1`, tech.md `:1` | PASS |
| 006c | P6 | no output | PASS |
| 006d | P5 | no output | PASS |
| 006e | negative / positive greps on prompting-best-practices (both copies) | negative `:0`/`:0` rc=1; positive `:1`/`:1` rc=0 | PASS |
| 007 | i18n greps + `go test -run 'TestModelOptLabelsEnglishUnified\|TestEffortOptRecommendationLabels' -v ./internal/web/` + `TestResolveLaunchEffort` | model label `4`, stale `0`; medium ×4 locales `1` each; runtime_default ×4 locales `1` each; `--- PASS` both web tests (`run-ac007-web.log`); `--- PASS: TestResolveLaunchEffort` with 5 subtests PASS (`run-ac009-reverted-cli.log`) | PASS |
| 008 | 8 per-locale `grep -cE` on `profile_setup_translations.go` | `1` ×8 | PASS |
| 009a/b/c | mutants M-a / M-b / M-c (`run-ac009-mutants.log`) | M-a: `lang="ko": ModelOpus "opus (Opus 5 , 적응형 사고)" should reference Opus 5.5`. M-b: `locale "ko" option "high": description does not name "Opus 5.5"`. M-c: both guards FAIL with `should reference Opus 6` / `does not name "Opus 6"`. Reverted trees → `ok` | PASS |
| 010 | `cmp` × 6 SAME pairs | all rc=0 | PASS |
| 011 | `grep -n -i effort internal/template/templates/.claude/settings.json.tmpl` | no output, rc=1 | PASS |
| 012 | `go test -count=1 ./internal/template/ ./internal/cli/wizard/ ./internal/web/ ./internal/settings/ ./internal/constitution/`; `go test -count=1 -timeout 25m ./internal/cli/` under slot `go-test-cli` | template/wizard/settings/constitution `ok`. web `FAIL`: the only failing tests are `TestDataI18nKeysSubsetOfDictionary` and `TestI18nKeySetParity`, and every failure line names a `factory_msg_*` key; the same two fail on a pristine `git archive f30d07ea7` extract (`run-baseline-web-f30d07ea7.log`). cli `FAIL` 1229s: 4 tests, attributed in `run-ac012-cli-attribution.log`. The one attributable to this card (the third guard) was fixed; its targeted run is `ok` | FAIL-attributed (pre-existing, not this card) |
| 013 | `git diff develop...HEAD -- internal/template/profile_matrix.go` | two hunks, comment lines only (`//   - Opus dominates …(measured on Opus 5)…`, `//   - \`xhigh\` is retired …: measured on Opus 5, …`); no `defaultProfileMatrix` cell changed | PASS |
| 014 | `git diff --name-only develop...HEAD -- CHANGELOG.md docs-site README*.md .moai/research .moai/docs .moai/reports ':!.moai/reports/t1089'` | no output. Control: the same range without pathspec lists 60 files. `.moai/specs` lists only `SPEC-MODEL-OPUS55-001/*` | PASS |
| 015 | `go test -run 'TestTemplateNoInternalContentLeak\|TestLanguageNeutrality\|TestLeakClassNoDateShaInDefaultTier' -v ./internal/template/` | three `--- PASS` lines; `ok` (`run-ac015.log`) | PASS |
| 016 | `make build`; `git diff --name-only develop...HEAD -- internal/template/templates/.claude/agents` | build rc=0 (`run-make-build.log`, catalog.yaml refreshed 3 hashes); agents diff empty, so `make agents-emit` is not required | PASS |

### E.2.3 Quality checks

- `go vet ./internal/template/ ./internal/cli/wizard/ ./internal/web/ ./internal/settings/ ./internal/constitution/` and `go vet ./internal/cli/` both returned no output.
- `GOOS=windows GOARCH=amd64 go build ./...` → rc=0.
- `golangci-lint run ./internal/template/... ./internal/cli/... ./internal/web/...` → rc=1 with `25 issues: errcheck: 25` (`run-lint.log`). All 25 are in `codex_launcher*_test.go`, `factory*.go`, `launch_exec_posix_rollback_test.go`, and `mcp_factory_msg.go`, and none of those files is in this card's diff. New findings: 0.
- `go test -cover ./internal/template/` → `coverage: 81.7% of statements`. A pristine `f30d07ea7` extract also reports `81.7%`. That extract lacks repo-root files, so some of its tests fail early, and the comparison is approximate. The package figure is below the 85% target both before and after this card.

### E.2.4 Plan-audit iter-2 minor findings — decisions

- **N2 (P6 false-fires on correct text).** Not changed in acceptance.md, because the AC body is manager-spec territory. The run worded around the probe instead, for example `- high: default on most effort-capable models (not Opus 5.5)`. That keeps the shapes P6 detects out of correct rewrites, and P6 now reads green for the right reason.
- **N3 (the `(superseded)` escape applies to the whole line).** Adopted the optional part. Row 7 (the vendor statement that low and medium are stronger on Opus 5) now carries the honest attribution `measured on Opus 5` instead of `(superseded)`. `(superseded)` is now used only where the id really is superseded: `claude-opus-5` in the fact line and in tech.md. The Known-limit note in AC-004 is not added (acceptance.md body) and is carried to sync.
- **N4 (the empty option names a field the web console does not show).** Adopted. Each locale's `opt.runtime_default` now names where the policy is set: "(model policy from moai profile setup, else Claude Code default: medium on Opus 5.5)".

### E.2.5 Residual risk

- K4: fixed. It is no longer residual; see E.2.6.
- K6: a profile `opus[1m]` now launches `--model claude-opus-5-5[1m]`, which needs Claude Code v2.1.280+. No version floor was added.
- The TUI wizard's effort empty option comes from `settings.EmptyLabelFor("effort_level")` and still reads "(runtime default)". REQ-OP55-007 covers only the web console option.

### E.2.6 K4 fix — `max` leaves the settings path (Kanban lead dispatch, tree `6cd26e504`)

- **Defect path.** The orchestrator measured it as follows:
  - The profile value `effort_level: max` passes through `resolveLaunchEffort` and `applyLaunchEffort`, which wrote `effortLevel: "max"` into the injected `--settings` payload.
  - Claude Code's settings key does not accept `max`. The model-config page says: "set `effortLevel` to `low`, `medium`, `high`, or `xhigh` … `max` isn't accepted as a level in either key".
- **Fix.**
  - `applyLaunchEffort` (`internal/cli/launch_effort_settings.go`) now returns `(payload, launchArgs)`. A resolved `max` never enters the payload and comes back as `[--effort max]`. The `@MX:NOTE` on that branch explains why.
  - low/medium/high/xhigh are unchanged and stay on the settings path.
  - `CLAUDE_CODE_EFFORT_LEVEL` is still not used, and `buildEnvForClaudeLaunch` is unchanged.
  - An operator-supplied `--effort` wins, so no second flag is added.
- **Argv injection sites.** Both paths append the flag after the settings pair, or on their own when the payload is empty:
  - `appendCrossSessionSettings` (`internal/cli/crosssession_settings.go`, general launch funnel `launcher.go:219`).
  - `prepareKanbanSettings` (`internal/cli/kanban_settings.go`, kanban/factory lanes in `cc.go` / `glm.go`).
- **GLM env path.** `buildEnvForGLMLaunch` is not touched.
- **RED** before the fix: `go test -count=1 -run 'TestLaunchEffortMax|TestLaunchEffortXHighStaysOnSettingsPath' -v ./internal/cli/` (`run-k4-red.log`):
  - `launch_effort_settings_test.go:228: settings effortLevel = max; max must never be written to the settings payload`
  - `launch_effort_settings_test.go:231: args = [-p dev --settings …/moai-crosssession-….json], want exactly one --effort max`
  - the same two failures on the kanban path (`:248`, `:251`)
  - `FAIL github.com/modu-ai/moai-adk/internal/cli`
- **GREEN.** `go test -count=1 -run 'TestLaunchEffort|TestApplyLaunchEffort|TestClaudeLaunchEnvPreservesInheritedEffort|TestResolveLaunchEffort|Kanban|CrossSession' -v ./internal/cli/` returned 46 `--- PASS` and `ok` (`run-k4-green.log`). Passing tests include:
  - `TestLaunchEffortMaxTravelsAsArgvOnGeneralInjection`
  - `TestLaunchEffortMaxTravelsAsArgvOnKanbanInjection`
  - `TestLaunchEffortXHighStaysOnSettingsPath`
  - `TestLaunchEffortMaxDefersToOperatorEffortFlag`
  - `TestApplyLaunchEffort/max_returns_launch_argv_and_stays_out_of_the_payload`
  - `TestClaudeLaunchEnvPreservesInheritedEffort`
- **Mutant.** Removing the operator-`--effort` guard makes `TestLaunchEffortMaxDefersToOperatorEffortFlag` fail with `args = [--effort low --effort max], want only the operator's --effort low`. After revert it returns `ok`.
- **Checks.** `go vet ./internal/cli/` returned no output. `golangci-lint run ./internal/cli/...` returned `0 issues.` with rc=0 (`run-k4-lint.log`).
- **Full `internal/cli`** (slot `go-test-cli`), unscrubbed lane env: `FAIL` 1108.8s, with only `TestCodexSpawn_RealAssemblyThroughStubTmux` and `TestCC_FactoryEntryThroughRunCC/-f_lane-2` (`AMBIGUOUS_FACTORY`) failing (`run-k4-cli-full.log`). Both pass once the lane `MOAI_*` env is unset in the same invocation (`run-k4-cli-scrubbed.log`). The full-package run was repeated with the scrubbed env in one compound invocation (`unset MOAI_AUTONOMY_TIER MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_KANBAN_BACKEND MOAI_KANBAN_SETTINGS_INJECTED MOAI_LAUNCH_PROVIDER MOAI_PROFILE_LEASE_TOKEN MOAI_SESSION_PID MOAI_CONFIG_SOURCE && go test -count=1 -timeout 25m ./internal/cli/`) and returned `ok  	github.com/modu-ai/moai-adk/internal/cli	1327.513s` (`run-k4-cli-full-scrubbed.log`). The slot was re-acquired for that run and released afterwards.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-23
run_commit_sha: pending-backfill   # M4 commit carries this section; M1 06c5133dd, M2 7100be63b, M3 43aa64cba
run_status: complete-with-attributed-failures
ac_pass_count: 15
ac_fail_count: 1   # AC-OP55-012: pre-existing web/cli failures on the absorbed tree; see E.2.2 row 012
preserve_list_post_run_count: 0   # AC-014 historical-surface diff empty
l44_pre_commit_fetch: not-run   # lane does not push; lead batch-pushes develop
l44_post_push_fetch: not-applicable
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: ok
  windows_amd64: ok
total_run_phase_files: "git diff --name-only f30d07ea7..<M4> (measured in the M4 report)"
m1_to_mN_commit_strategy: per-milestone commits M1..M4 on WT-opus-55-default, no push
```

## §E.4 Sync-phase Audit-Ready Signal

Sync by manager-docs on 2026-09-23, card t1089. Worktree `.claude/worktrees/t1089`, branch `WT-opus-55-default`. Sync started at HEAD `eb629efb5`; develop `1dbe5e2f3` had been absorbed at `9e6d57f4f`.

### E.4.1 AC-OP55-012 re-close (post-absorb; supersedes the E.2.2 row 012 FAIL)

The orchestrator re-measured this AC after the absorb. The manager-docs sync read every log below back from disk.

| Package set | Evidence file | Decisive line |
|---|---|---|
| template, wizard, settings, constitution | `run-ac012-pkgs.log` | `ok` ×4. web was the only FAIL in this run |
| web | `remeasure-ac012-web.log` | `ok  	github.com/modu-ai/moai-adk/internal/web	22.656s` |
| cli, targeted (5 tests) | `remeasure-ac012-cli.log` | `TestGLM_FactoryWorkerEntry`, `TestSessionPIDStamp_NotSetFromHooks`, `TestResolveLaunchEffort`, `TestModelPolicyLabels_AgreeWithProfileMatrix`, `TestGetProfileText_OpusAliasValues` all pass (13 `--- PASS` lines including subtests); last line `ok  	github.com/modu-ai/moai-adk/internal/cli	1.689s` |
| cli, full package with lane env scrubbed | `run-k4-cli-full-scrubbed.log` | `ok  	github.com/modu-ai/moai-adk/internal/cli	1327.513s` |
| lint on template, cli, web, settings | `remeasure-ac012-lint.log` | `0 issues.` |

Result: AC-OP55-012 is **PASS**, so the tally is **16/16 PASS**. The unscrubbed cli run had two failures caused only by env leakage (`TestCodexSpawn_RealAssemblyThroughStubTmux`, `TestCC_FactoryEntryThroughRunCC/-f_lane-2`). E.2.6 records both, and they are not attributed to this card.

AC-OP55-014 was re-run by manager-docs on HEAD `eb629efb5`: `git diff --name-only develop...HEAD -- CHANGELOG.md docs-site README.md README.ko.md README.ja.md README.zh.md .moai/research .moai/docs .moai/reports ':!.moai/reports/t1089'` gave empty stdout, so it is PASS.

### E.4.2 Sync decisions

- **CHANGELOG.md not touched.** REQ-OP55-015 and AC-OP55-014 list CHANGELOG.md as a historical surface this SPEC must not modify. Release notes for this change belong to the release lane, per the lead's decision on 2026-09-23.
- **README and docs-site not touched.** Card t1094 owns them (spec.md §C).
- **Status transition.** Only `spec.md` `status:` changed, `in-progress → completed`, on this sync commit. `updated:` was already `2026-09-23`. plan.md and acceptance.md have no status frontmatter.
- **Basis for `completed`.** acceptance.md §4 Definition of Done requires all sixteen ACs green with evidence in §E.2. Rows 001–011 and 013–016 are PASS in E.2.2, AC-012 is PASS in E.4.1, and the AC-009a/b/c mutant evidence is in `run-ac009-mutants.log`. The follow-up docs card from the DoD is t1094.
- **N2–N4 dispositions** (from E.2.4, unchanged):
  - N2: P6 was worded around, and acceptance.md is unchanged.
  - N3: `measured on Opus 5` attribution was adopted. The Known-limit note already exists at acceptance.md:61 ("a phrase shaped 'Opus 5-era' is not matched"), so no new note was added.
  - N4: adopted. `opt.runtime_default` names where the effort policy comes from.

### E.4.3 Residual risk

- **K4 is fixed** (`eb629efb5`, E.2.6). `max` is now passed as `--effort max` and is never written as settings `effortLevel`. Three residuals remain:
  - (a) A Claude Code session launched through GLM now also receives `--effort max` for a max profile. It used to receive settings `effortLevel`.
  - (b) An operator-supplied `--settings` still suppresses profile effort injection. This behavior predates this card.
  - (c) No live `claude --effort max` launch was run. The basis is the official model-config doc quote only.
- **K6.** A profile `opus[1m]` launches `--model claude-opus-5-5[1m]`, which requires Claude Code v2.1.280+. No version floor was added.
- **TUI wizard.** The wizard's effort empty option (`settings.EmptyLabelFor("effort_level")`) still reads "(runtime default)". This is outside REQ-OP55-007, which covers the web console only. It is a candidate for a follow-up card.

### E.4.4 Signal

```yaml
sync_complete_at: 2026-09-23
sync_commit_sha: pending-backfill   # a commit cannot cite its own hash
sync_status: complete
ac_pass_count: 16
ac_fail_count: 0
frontmatter_status_transitions:
  spec_md: in-progress -> completed
changelog_entry_position: none   # REQ-OP55-015 / AC-OP55-014 forbid CHANGELOG edits
b12_self_test_a: not-applicable   # no CHANGELOG emission
b12_self_test_b: "acceptance.md distinct AC ids = 16"
b12_self_test_c: not-applicable
docs_surfaces_touched: none   # README/docs-site -> card t1094
push: not-run   # lead batch-pushes develop
```
