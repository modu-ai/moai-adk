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

- K4 (unchanged): the official settings reference says `effortLevel` does not accept `max`, but the launcher can pass `max`.
- K6: a profile `opus[1m]` now launches `--model claude-opus-5-5[1m]`, which needs Claude Code v2.1.280+. No version floor was added.
- The TUI wizard's effort empty option comes from `settings.EmptyLabelFor("effort_level")` and still reads "(runtime default)". REQ-OP55-007 covers only the web console option.

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

_<pending sync-phase>_
