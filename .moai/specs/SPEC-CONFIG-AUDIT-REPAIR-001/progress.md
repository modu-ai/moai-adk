# Progress — SPEC-CONFIG-AUDIT-REPAIR-001

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored 2026-07-25; amended to 0.2.0 same day after Implementation Kickoff: all 3 design decisions RESOLVED (D1 ast-grep gate RESTORE default-OFF — user-directed reversal of the removal recommendation; D2 tool-policy dev-only; D3 mcp-matrix reword) and plan-auditor SHOULD-FIX items folded (touch-point inventory, AC-CAR-002/005/007(c) tightening, LOW counting convention pinned in HISTORY). No [NEEDS CLARIFICATION] markers remain. Run-ready.

## §E.2 Run-phase Evidence

### M-2 — ast-grep gate restore (REQ-CAR-011/019/020/021)

| Item | Evidence |
|------|----------|
| gate.yaml loader (REQ-CAR-011) | `internal/config/loader_gate.go` + `Loader.Load` registration; `go test -run TestLoadGateSection ./internal/config/` → PASS (enable / default-off / partial-override) |
| Explicit default OFF (REQ-CAR-019) | `defaults.go NewDefaultGateConfig` sets `AstGrepGate.Enabled: false` explicitly; `TestLoadGateSection_DefaultOff` + `TestPreToolHandler_LoadGateConfig/astgrep_gate_default_off_via_config` PASS |
| Guard reachability (AC-CAR-011) | `TestQualityGate_Run_AstGrepGuardReachable` (zig marker + pure-Go suppression violation → Run returns false via gate.go astgrep branch) + `TestPreToolHandler_LoadGateConfig/astgrep_gate_enabled_via_config` (pre_tool mapping) — both PASS |
| V1 deletion (REQ-CAR-021) | `RunAstGrepGate` + `runSGConfig`/`runSGRule`/`parseSGScanOutput`/`astGrepScanMatch` deleted from `astgrep_gate.go`; V1 tests + `TestRunAstGrepGate_V1_V2_Equivalence` removed; `go build ./...` green |
| sgconfig `utils` phantom (REQ-CAR-020a) | `grep -n utils sgconfig.yml` → 0; ruleDirs now `[go, security]` (loadable curated set; demo-stub language dirs excluded pending ASTGREP-DOGFOOD-CLEANUP-001) |
| Config-mode loadability (REQ-CAR-020b) | go/ + security/ rules converted to sg-native `rule:` nested format; 2 sg-unparseable multi-node demo rules dropped (go-http-response-body-not-closed, go-mutex-not-deferred); manual run `sg scan --config .moai/config/astgrep-rules/sgconfig.yml --json internal/config/loader_gate.go` → exit 0, output `[]` |
| Graceful degrade without `sg` (REQ-CAR-020c) | `TestRunAstGrepGateV2_NoSgCLI` (PATH="") PASS — enabled gate skips scan, no hard failure |
| gate.yaml template+local | template-first: `internal/template/templates/.moai/config/sections/gate.yaml` + local mirror; `make build` exit 0; `TestStructYAMLSymmetry_Gate` + `TestAuditLoaderCompleteness` PASS |

### M-3 — H4+M3 distribution codification (REQ-CAR-013/014)

| Item | Evidence |
|------|----------|
| tool-policy graceful CLI (REQ-CAR-013) | `internal/cli/tool_policy.go`: absent-file guard prints one-line dev-only notice, exit 0 for both `list` and `build`; verified in clean `/tmp/tpclean` — `moai tool-policy list` → notice + exit=0, no stack |
| tool-policy dev-only declaration | CLAUDE.local.md §2 Local-Only Files entry added (`.moai/config/sections/tool-policy.yaml`) |
| mcp-matrix reword (REQ-CAR-014) | Template refs removed: `grep -rn 'mcp-matrix' internal/template/templates/` → 0 matches (project.md table row + doc-generation.md Step 3.6.2 generalized to matrix-optional wording, both trees); CLAUDE.local.md §2 entry added |
| Build + guards | `make build` exit 0; `go test ./internal/cli/ ./internal/template/...` green |

### M-4 — M7 test repair + registry reconciliation (REQ-CAR-009/010/007a/007b)

| Item | Evidence |
|------|----------|
| TestAuditParity repo-root fix (REQ-CAR-009) | `audit_test.go` resolves repo root via `runtime.Caller` (Skip → Fatalf); `go test -run 'TestAuditParity$' -v ./internal/config/` → `--- PASS: TestAuditParity`, no SKIP |
| Orphan reconciliation (REQ-CAR-010) | Repaired test exposed 7 orphan yamls (cache/db/feedback/mcp-matrix/observability/report/tool-policy) + 1 orphan struct (runtime). feedback → registry (loadFeedbackSection exists); runtime removed (RuntimeConfig loads from internal/runtime path, not sections/); remaining 6 registered as documented exceptions with rationale |
| Zero skips (AC-CAR-010) | `go test ./internal/config/... -v \| grep -c SKIP` → 0; package green |
| db comment fix (REQ-CAR-007a) | completeness-test allowlist db comment corrected — hook consumes migration_patterns via internal/cli/hook.go line-scan |
| constitution labeling (REQ-CAR-007b) | audit_registry exceptions relabeled loader-present (loader_constitution.go et al. via Loader.Load); entries retained for TestAuditParity_ExceptionsRespected compatibility |

### M-5 — Mechanical doc/rule sweep (REQ-CAR-001..008, LOW c/d/e)

| Item | Evidence |
|------|----------|
| H1 role_profiles (AC-CAR-001) | 3 rule files reworded (retired/historical wording only); grep shows no live-config phrasing; ORC_WORKTREE_REQUIRED row marked legacy-inert (sentinel const in agentlint retained) |
| H2 config.yaml → sections SSOT (AC-CAR-002) | `grep -rn '\.moai/config/config\.yaml'` over the 6 enumerated files → 0 matches; CLAUDE.local §5 version list repointed to sections/system.yaml.tmpl |
| Harness abort gate (AC-CAR-003) | harness.md gate now checks `.moai/config/sections/` directory existence (both trees) |
| M2 delivery keys (AC-CAR-004) | `grep -c spec_git_workflow delivery.md` → 5; `grep -n '[^_]git_workflow'` → 0 |
| M4 mcp-servers.yaml (AC-CAR-005) | repo grep (excl. this SPEC dir + reports + worktrees) → 0 matches |
| M6 (AC-CAR-006) | phase_weights → adaptation.iteration_limits (real design.yaml key); bare `quality.yaml development_mode` grep (excl. stale `.claude/worktrees/` copies) → 0 |
| LOW c/d/e (AC-CAR-007) | dynamic-workflows enum haiku-free at 87/91/95 sites; `grep -c haiku internal/config/defaults.go` = 2 (unchanged baseline); db.auto_sync described as map (auto_sync.enabled); `grep 'coverage_threshold: 0' quality.yaml` → 0 |
| Template-first parity (AC-CAR-008) | Every mirrored file edited template-first + local; `make build` exit 0; template neutrality/leak tests green. Parity-skip list (no mirror, by design): CLAUDE.local.md, MCP_OAUTH_SETUP.md, local quality.yaml, sgconfig/astgrep-rules |

### M-6 — Closeout (REQ-CAR-015/016/017/018) + collision fix

| Item | Evidence |
|------|----------|
| M1 db dependency (AC-CAR-015) | No auto_sync key implementation (db Go surfaces untouched); dependency on the parallel DB-removal track recorded in spec.md REQ-CAR-015 + plan.md §B; fallback (delete 3 dead keys) armed if that track stalls |
| tool-policy/lsp SSOT refs (AC-CAR-016) | CLAUDE.local §2 tool-policy entry; CLAUDE.md §6 names lsp.yaml as the LSP-gate SSOT (both trees); acknowledged-orphan inventory added in exactly one location (settings-management.md § Acknowledged config orphans, both trees) |
| gate.yaml DeprecatedPaths collision | `TestDeprecatedPaths_NoTemplateCollision` caught the stale v2 gate.yaml deprecation entry vs the newly-shipped template file (clean-reinstall-loop hazard, #1084 class); entry removed from internal/defs/dirs.go, count pins 40→39 / Category B 28→27 updated with rationale |
| Lint delta (AC-CAR-017) | `golangci-lint run` → `0 issues.` (fixed the one NEW errcheck my change introduced; pre-edit `go vet ./...` baseline exit 0, post exit 0) |
| Neutrality guards (AC-CAR-018) | `go test ./internal/template/...` green after every template edit |
| Full suite | `go test ./...` exit 0, 107 packages ok (evidence: .moai/state/verify/config-audit-repair-001/mv-test2.log); `GOOS=windows GOARCH=amd64 go build ./...` exit 0 |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-25
run_commit_sha: 850f579fb
run_status: audit-ready
ac_pass_count: 20
ac_fail_count: 0
ac_pass_with_debt: 1   # AC-CAR-020(b) — config-mode loads the curated go/security set (not the literal root file; same rules, see M-2 row)
preserve_list_post_run_count: intact (internal/cli/uikit, internal/cli/printer, internal/tui, SPEC-CLI-TUX-INIT-UPDATE-001 dir untouched)
new_warnings_or_lints_introduced: 0 (golangci-lint "0 issues.")
cross_platform_build:
  darwin: pass
  windows_amd64: pass
total_run_phase_files: 55
m1_to_mN_commit_strategy: per-milestone pathspec commits on feat/SPEC-CLI-TUX-INIT-UPDATE-001 (no push per session constraint)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-25
sync_commit_sha: pending-backfill-sync-config-audit-repair-001
sync_status: audit-ready
changelog_entry_position: "[Unreleased] > ### Fixed (top entry, above SPEC-SUBAGENT-NESTING-DOCTRINE-001)"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed"
  plan_md: "in-progress -> completed"
  acceptance_md: "in-progress -> completed"
docs_scope_decision: "README/docs-site untouched — repaired surfaces are internal harness docs (.claude/rules, .claude/skills, internal/config), not user-facing product docs"
residual_debt:
  - id: AC-CAR-020b
    note: "config-mode scan loads curated go/security rule set (sg-native format), not literal unconverted root go-hardcoding.yml; 2 sg-unparseable demo rules dropped"
    owner: "SPEC-ASTGREP-DOGFOOD-CLEANUP-001 (backlog, already related_specs)"
  - id: template-extra-config-yaml-refs
    note: "residual .moai/config/config.yaml-shaped references may remain in moai-foundation-core / moai-workflow-worktree template skill modules outside the 6 enumerated AC-CAR-002 sites"
    owner: "follow-up candidate, not yet a SPEC"
  - id: M1-db-yaml-dead-keys
    note: "auto_sync.{debounce_seconds,require_user_approval,excluded_patterns} intentionally not implemented; dependency on parallel DB-removal track (memory: project_db_retire_removal_approved.md), delete-3-keys fallback armed if that track stalls"
    owner: "DB subsystem removal track"
```
