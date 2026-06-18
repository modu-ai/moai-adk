# Progress — SPEC-WEB-CONSOLE-002

> S1 of the `web-console-v3-extension` cohort. Tier S. cycle_type=tdd.
> Run-phase evidence for the closure gate (`go test ./internal/web/... ./internal/cli/...`).

## §E.2 Run-phase Evidence

| AC | Status | Verification Command | Actual Output |
|----|--------|----------------------|---------------|
| AC-WC2-001 | PASS | `go test -run 'TestWebCmd_FlagsRegistered\|TestWebCmd_HelpListsFlags' ./internal/cli/` | `ok` — `--port` DefValue == "3041"; help documents 3041, no 8080 |
| AC-WC2-002a | PASS | `go test -run 'TestValidatePrefs_ModelField\|TestSaveInvalidModelRejected' ./internal/web/` | `ok` — bogus model → 400 + `FieldErrors["model"]` + no persistence; 6 canonical + empty accepted |
| AC-WC2-002b | PASS | `go test -run TestSaveValidModelPersisted ./internal/web/` | `ok` — `model=sonnet[1m]` → 200, persisted value `sonnet[1m]` |
| AC-WC2-003 | PASS | `go test -run 'TestValidatePrefs_EffortLevelField\|TestSaveInvalidEffortLevelRejected\|TestSaveValidEffortLevelPersisted' ./internal/web/` | `ok` — bogus `ultra` → 400; `xhigh` → 200 persisted; 5 canonical + empty accepted |
| AC-WC2-004 | PASS | `go test -run 'TestValidatePrefs_ModelPolicyField\|TestSaveInvalidModelPolicyRejected\|TestSaveValidModelPolicyPersisted' ./internal/web/` | `ok` — `template.IsValidModelPolicy` wired; bogus → 400; `medium` → 200 persisted; high/medium/low + empty accepted |
| AC-WC2-005 | PASS | `go test -run TestRenderModelEffortPolicyAreSelects ./internal/web/` | `ok` — model/effort_level/model_policy render as `<select>`; no `type="text"`; canonical options present; current values marked selected |
| AC-WC2-006 | PASS | `go test -run 'TestProfileText_ModelPolicyLabels\|TestProfileSetup_ModelPolicySelectPresent' ./internal/cli/` | `ok` — model_policy `huh.Select` in model-settings group (4-locale labels pre-existed); ModelPolicy wired into saved prefs |
| AC-WC2-007 | PASS | `go test -run TestSaveValidModelPolicyPersisted ./internal/web/` | `ok` — model_policy persists profile-only; no new config section, SyncToProjectConfig scope unchanged |
| AC-WC2-008 | PASS | `go test ./internal/web/...` | `ok` — full SPEC-WEB-CONSOLE-001 suite green; loopback/no-auth/Host-header/persistence-path tests unchanged |
| AC-WC2-009 | PASS | `go test ./internal/web/... ./internal/cli/...` | `ok` (all packages) — closure gate exit 0 |

**Invariant checks:**

| Invariant | Status | Evidence |
|-----------|--------|----------|
| Loopback-only bind unchanged | PASS | No `0.0.0.0` bind added; `TestIsLoopbackHost` unchanged |
| No-auth / no-token / no-session | PASS | No auth surface added in S1 |
| Host-header write-safety | PASS | Existing host-gate tests pass unchanged |
| Persistence via WritePreferences + SyncToProjectConfig only | PASS | No direct YAML marshal in web layer |
| No duplicate validation rule-set | PASS | model/effort mirror wizard SSOT with MX:NOTE; model_policy imports `template.IsValidModelPolicy` |
| No new config section for model_policy | PASS | SyncToProjectConfig scope untouched (REQ-WC2-007) |
| No template mirroring | PASS | web assets embedded via web package `go:embed`; no `internal/template/templates/` change |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-03
run_commit_sha: "3a050ba51"
run_commit_sha_note: "reconciled to integrated branch SHA (L_manager_docs_stale_worktree_sha); pre-integration L1 worktree SHA was 25dd5d6fe4211612f852e89bf5091299d4f7b791"
run_status: implemented
ac_pass_count: 9
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "0 0 (worktree branched from ad974fe5b; SPEC artifacts uncommitted in parent checkout)"
l44_post_push_fetch: "n/a — push deferred to orchestrator"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: pass
  windows: "GOOS=windows GOARCH=amd64 go build ./... exit 0"
total_run_phase_files: 8
m1_to_mN_commit_strategy: "single M1 commit (Tier S) — implementation + spec.md draft→in-progress + progress.md"
coverage_internal_web: "90.8%"
```

## §E.4 Notes

- **L1 worktree**: run-phase executed in Claude Code runtime L1 worktree at `.claude/worktrees/agent-a2644d34c55bcb8ec` (branch `worktree-agent-a2644d34c55bcb8ec`, base `ad974fe5b`). The SPEC plan-phase artifacts (`spec.md`/`plan.md`) were uncommitted working-tree files in the parent checkout; they were copied into the worktree for the unified M1 commit. Orchestrator must integrate the worktree commit into `docs/glm-webtool-routing-m1-m5` and reconcile the parent checkout's uncommitted SPEC artifacts.
- **Pre-existing failures (out of scope, verified on baseline without WC2 changes)**: `internal/template` `TestOutputStylesTemplateLiveParity` (einstein.md drift — `M ...einstein.md` present in working tree at session start, parallel-session owned) + `internal/hook` `TestHookWrapper_ValidJSON`/`TestHookWrapper_MoaiBinaryFallback` (~5.01s timeout flaky under parallel-suite contention; pass in isolation). Neither involves the 8 WC2 scope files.

### (Migrated from §E.5)

```yaml
sync_complete_at: 2026-06-03
sync_commit_sha: "76ad74f8e"
sync_status: implemented
doc_deliverables: "CHANGELOG.md [Unreleased] entry (Changed: port 3041 supersede + Added: validation parity + widget select-ification + TUI model_policy select); spec.md frontmatter in-progress→implemented"
readme_docs_site_scope: "n/a — moai web undocumented in README/docs-site (0 refs verified); web i18n/webfont deferred to cohort S3"
```

## §E.6 Mx-phase Audit-Ready Signal

```yaml
mx_complete_at: 2026-06-03
mx_commit_sha: "7778f6586"
mx_status: completed
four_phase_close: "plan (fad1be853) → run (3a050ba51, M1) → sync (76ad74f8e) → Mx (this transition)"
lifecycle_transition: "implemented → completed"
mx_tag_scope: "MX:NOTE present at internal/web model_policy bindForm/persist path per plan.md M5 (REQ-WC2-007 profile-only by design)"
cohort_position: "S1 of web-console-v3 cohort closed; S2 (8 missing v3 settings) / S3 (web i18n + webfont) / S4 (dead-config audit) remain"
```

