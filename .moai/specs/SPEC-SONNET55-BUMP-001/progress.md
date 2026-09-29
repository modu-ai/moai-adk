# SPEC-SONNET55-BUMP-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
spec: SPEC-SONNET55-BUMP-001
status: draft
plan_status: audit-ready
plan_complete_at: 2026-09-29
artifacts: [spec.md, plan.md, acceptance.md, research.md, progress.md]
tier: M
anchors_measured_at: a62a05764
card_anchor_corrections: 3  # spec.md §A.3
audit_iterations: 2         # iter-1 FAIL (0.825, D1 must-pass) -> fixes applied
open_research: [context-window figure per official docs — research.md §2]
needs_clarification: 0
served_model_gate_exception: operator approval (relayed by lead 2026-09-29) — served_model_gate disabled in this tree only + GLM-served plan-audit adopted; extension of the operator's existing approval to t1322 via the lead's pre-announced extension clause (7th unlock card; standing until t1323/t1324 structural fixes land)
auditor_serving_correction: both plan-audit reports name the OBSERVED serving model glm-5.3-flash in their first lines, matching the gate's own served-model declaration (expected opus, served glm-5.3-flash) — no spawn-injection mislabel to correct (contrast t1310). The plan-audit verdicts are valid and proceed under the GLM adoption. Run spawns stay model-inherited; SERVED_MODEL_VIOLATION re-occurrence must be reported, never routed around.
```

Plan-phase evidence: all file anchors in spec.md §A.1 were measured in this tree at
`a62a05764` (grep/sed reads of `model_policy.go`, `glm_effort_overlay.go`, `launcher.go`,
`validate.go`, `i18n.js`, `profile_setup_translations.go`, template hits, README/docs-site
counts). SPEC ID regex check executed: `SPEC-SONNET55-BUMP-001` PASS (the card-suggested
`SPEC-SONNET-55-BUMP-001` FAILS the `[A-Z][A-Z0-9]*` segment rule — digit-leading segment).

## §E.2 Run-phase Evidence

### t1315 merge-order record (AC-SSB-011)

`gh issue view 1730 --json state` → `{"state":"OPEN"}` (observed 2026-09-29, this worktree).
Direction: **this SPEC lands first; t1315 (GitHub #1730, live `expandModelString` resolution)
absorbs this SPEC's table contents** when it lands. This branch carries no change to
`expandModelString`'s mechanism (REQ-SSB-011 / plan.md §F); its branch rebases onto the promoted
table (`ModelIDSonnet55`) without semantic conflict.

### AC matrix (AC-SSB-001..012)

| AC | Status | Verification command | Observed output (this run, this tree) |
|---|---|---|---|
| AC-SSB-001 | PASS | `grep -n 'ModelIDSonnet55' internal/template/model_policy.go` | `56:const ModelIDSonnet55 = "claude-sonnet-5-5"` + `"sonnet":   ModelIDSonnet55,`; `grep -c '"sonnet":\s*"claude-sonnet-5"'` → `0` |
| AC-SSB-002 | PASS | `grep -n '"claude-sonnet-5":' internal/template/model_policy.go` | `98:	"claude-sonnet-5":   "sonnet", // superseded by ModelIDSonnet55` |
| AC-SSB-003 | PASS-WITH-DEBT | `go test -timeout 30m -count=1 ./internal/template/... ./internal/cli/... ./internal/web/... ./internal/hook/...` | template ok / web ok; cli + hook carry PRE-EXISTING failures identical to the pre-edit baseline (see Blockers); `launcher_test.go`, `served_model_test.go`, `served_model_stop_test.go` diff empty (`git diff --stat` → 0 lines) |
| AC-SSB-004 | PASS | `make agents-emit-check && make build` | both exit 0; `git status --porcelain -- internal/template/templates/.codex/` → empty |
| AC-SSB-005 | PASS | `go test -run 'TestGLMSlotForModel$' ./internal/template/` | ok — `claude-sonnet-5-5`, `claude-sonnet-5`, `sonnet[1m]` all `GLMSlotMedium` |
| AC-SSB-006 | PASS | `grep -rn '"Sonnet 5.5"' …` / `grep -rnE 'Sonnet 5([^.0-9]\|$)' …` | picker labels `"Sonnet 5.5"` in all 4 i18n locale blocks; residual-anchor grep → 0 hits in both files |
| AC-SSB-007 | PASS | `git diff --stat -- …/moai-foundation-cc/reference …/moai-foundation-core` | empty (0 lines); `git diff --name-only -- internal/template/templates \| grep -c model-policy` → 1 |
| AC-SSB-008 | PASS | `grep -c 'Sonnet 5.5' README.{ko,en,ja,zh}.md` | 1 / 1 / 1 / 1 (parity); `grep -c '54%±4' README.ko.md` → 1 (historical values intact) |
| AC-SSB-009 | PASS | `hugo --destination /tmp/...` + parity greps | build warning-free (0 warn/error lines), sitemap present, model-policy 4-locale heading parity 29/29/29/29, `Sonnet 5.5` present in all 4 locales (en 7 / ko 10 / ja 5 / zh 5 files) |
| AC-SSB-010 | PASS | `grep -c 'between_tools' <5 surfaces>` + research.md §2.1 | 1/1/1/1 (docs-site 4 locales) + 1 (template model-policy rule); §2.1 carries the official citation (1M stated, tables unchanged per D-5) |
| AC-SSB-011 | PASS | `grep -n '1730' progress.md` | t1315 record above (#1730 OPEN → this SPEC first) |
| AC-SSB-012 | PASS | `grep -rn '"claude-sonnet-5-5"' --include='*.go' internal \| grep -v _test.go \| grep -v model_policy.go` | 0 hits |

### RED-first evidence (E8, M1 — captured BEFORE the table flip, tree at 7a3d74abf)

```
--- FAIL: TestGLMSlotForModel (0.00s)
    glm_slot_test.go:41: GLMSlotForModel("claude-sonnet-5-5") = "", want "medium"
--- FAIL: TestNormalizeModel_Deprecated/deprecated/claude-sonnet-5-5 (0.00s)
    profile_setup_normalize_test.go:70: normalizeModel("claude-sonnet-5-5") = "", want "sonnet[1m]"
glm_slot_effort_test.go:56: glmSlotEffortForModel("claude-sonnet-5-5") = "", want "e-medium"
```

M2 cascade RED (observed before fixing): `console_ux_fix_test.go:199: i18n.js has 0 occurrences
of "f.model.opt.sonnet[1m]": "Sonnet 5", want 4 (one per locale, English-unified)` and the
profile-wizard golden frames (ko/ja/zh) diverging on the model row + PermAuto row.

### Verification summary

- `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (pre-flight AND post-M1).
- `golangci-lint run ./internal/template/... ./internal/cli/... ./internal/web/...` (v2.1.6, CI-pinned) → `0 issues.`
- Coverage (whole-package, this run): `internal/template` 83.8%, `internal/cli` 84.1%, `internal/web` 74.5%. The SPEC diff adds zero executable production statements (M1: a const + map rows; M2: string constants + tests; M3-M4: docs), so no coverage regression is attributable to this SPEC — the gate acceptance.md §D.3 applies ("no coverage regression in internal/template; alias table is data"). The three figures are the packages' pre-existing baselines, marginally under the 85% quality.yaml target — recorded as a Gap, not caused or fixable by this card.
- Reference mirrors (`moai-foundation-cc/reference/**`, `moai-foundation-core/**`): zero edits (AC-SSB-007).
- `launcher_test.go`, `served_model_test.go`, `served_model_stop_test.go`, `internal/spec/**`: zero edits (REQ-SSB-010, B10c).

### Blockers / pre-existing defects found (out of scope — no drive-by fixes)

1. **TestTodoVerbSurfaceZeroDelta (internal/cli) — deterministic, pre-existing at base `a62a05764`.**
   `todo_surface_test.go:161/177: verb "relate <a> <b> --relation <contains|absorbs|replaces|conflicts>" is GONE … "…|blocks|depends>" appeared and is not a declared addition` — t1309 added the blocks/depends vocabulary without updating the declared surface-additions list. Owner: t1309 follow-up.
2. **TestAC_CIG_009_WorkflowYamlCarriesKey (internal/hook) — local-config-dependent.**
   `commit_identity_guard_test.go:656: local workflow.yaml commit_identity_guard.enabled = false, want true` — the test asserts the LOCAL dev `.moai/config/sections/workflow.yaml` enables the guard; this tree's config has it false (§22 dev-settings intent / §2.3 post-update key reapplication domain). Owner: maintainer config, not this SPEC.
3. **TestLocalInstructions_UpdateDoctorPreserveFile (internal/cli) — order-dependent flake** (fails in full-package runs, passes in isolation).

All three predate this SPEC (identical failure sets in the pre-edit baseline run) and none touches
a sonnet surface. AC-SSB-003 is reported PASS-WITH-DEBT on this basis.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-29
run_commit_sha: 4ccb12e2c   # M4 (implementation complete); progress.md evidence commit follows
run_status: complete
ac_pass_count: 11
ac_fail_count: 0
ac_pass_with_debt_count: 1   # AC-SSB-003 (pre-existing cli/hook failures, documented above)
preserve_list_post_run_count: 0   # zero edits to mirrors, launcher_test, served_model fixtures, internal/spec
l44_pre_commit_fetch: not-run   # leaf agent in card worktree; no push path — lane discipline (gitflow §4)
l44_post_push_fetch: not-run   # push is lead-batched; no local push performed
new_warnings_or_lints_introduced: 0   # golangci-lint v2.1.6: 0 issues on changed packages
coverage_note: "template 83.8 / cli 84.1 / web 74.5 (whole-package, pre-existing baseline; zero executable statements added by this SPEC — no regression attributable)"
cross_platform_build.darwin: pass
cross_platform_build.windows: pass   # GOOS=windows GOARCH=amd64 go build ./... exit 0
total_run_phase_files: 51   # 6 (M1) + 6 (M2) + 10 (M3) + 29 (M4) files across 4 milestone commits
m1_to_mN_commit_strategy: per-milestone commits (M1 f97b9bb2b, M2 4f88099d5, M3 ad95368aa, M4 4ccb12e2c)
```


## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-29
sync_commit_sha: 6467daf0c   # backfilled in follow-up commit (placeholder was pending-backfill-sync — SHA placeholder backfill exemption, D3)
sync_status: complete
changelog_entry_position: "CHANGELOG.md [Unreleased] § Changed — SPEC-SONNET55-BUMP-001"
frontmatter_status_transitions.in-progress_to_implemented: sync commit (merged with completed — single 3-phase close, no separate Mx commit)
frontmatter_status_transitions.implemented_to_completed: sync commit (same)
canary_compliance_check.na: "no forward-looking policy defined by this SPEC"
mx_tag_check: "ModelAliasTable already carries @MX:ANCHOR/@MX:REASON; @MX:NOTE added on ModelIDSonnet55 (new exported const, MX gate: new exported surface should carry a NOTE naming the alias contract)"
b12_self_test_a: "grep -c 'SPEC-SONNET55-BUMP-001' CHANGELOG.md = 0 pre-emission (duplicate guard PASS)"
b12_self_test_b: "acceptance.md distinct AC ids = 12 (AC-SSB-001..012); CHANGELOG entry cites 12 acceptance criteria AC-SSB-001..012 — count match"
b12_self_test_c: "all file paths named in the CHANGELOG entry verified via ls (model_policy.go, glm_effort_overlay.go alias surface, web i18n, template model-policy, README 4-locale, docs-site)"
codemap_rotation: "no codemap covers internal/template/model_policy.go alias surface — no rotation owed (verified by absence of a model_policy codemap entry)"
```

## §F Phase 4 Mode Selection

```yaml
phase: run
spec: SPEC-SONNET55-BUMP-001
logged_at: 2026-09-29
logged_by: lane-orchestrator (card t1322)
input_parameters:
  tier: M
  scope_files: "~20+ (Go alias/core 3, web i18n+labels 2-3, templates 6-9, README 4, docs-site 8-16)"
  domain_count: 4   # Go code, web assets, templates, docs (ko-canonical + 3 derived locales)
  file_language_mix: "Go + JS + YAML/TOML templates + Markdown"
  concurrency_benefit: LOW   # M1 alias flip is the critical path; M2-M4 are mechanical derivatives
  agent_teams_prereqs: not-applicable (no explicit operator request)
mode_evaluation:
  direct: "not selected — multi-file, multi-domain, semantic changes"
  fanout: "not selected — coding-heavy work (Anthropic coding-task parallelism caveat); M2-M4 depend on the M1 alias flip"
  sweep: "not selected — mixed semantic+mechanical transforms across dependent files; not a single uniform rule"
  serial: "selected — single manager-develop spawn, milestones M1→M4 in order"
decision: serial
```

Justification: the SPEC's four milestones form a dependency chain — the alias-table
flip (M1) is the semantic core every later milestone derives from, and the mechanical
label/translation edits (M2-M4) gain nothing from concurrency that offsets reconciliation
cost in a single-writer tree. Serial with the Tier M Section A-E delegation template is
the Anthropic-recommended default for coding-heavy work; no `sweep` confirmation is owed
since sweep is not selected.

Phase-1 Plan Audit Gate skip is taken per the authoritative skip contract: (1) iter2
verdict PASS, (2) score 0.95 ≥ 0.80 Tier M threshold, (3) plan-artifact hash unchanged
since the audited commit `8253da6dc` (tree clean, zero commits since). Recorded here and
in the run delegation prompt Section A.
