# SPEC-WEB-CONSOLE-013 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready (plan-audit iter-1 PASS 0.91 skip-eligible; SHOULD-FIX D1-D7 반영 → v0.1.1)
- plan_authored_at: 2026-07-10
- artifacts: spec.md (24 REQ, GEARS, v0.1.1) / plan.md (§A-§H) / acceptance.md (25 AC + §D.1 증거 표) / progress.md
- depends_on gate: SPEC-WEB-CONSOLE-012 — 2026-07-10 authored (status: draft, Tier S); `completed`까지 run 진입 차단 (의도된 게이트)

## §E.2 Run-phase Evidence

Run-phase executed on worktree branch `worktree-agent-abd265494e4c263d7` (base main HEAD `99ad0a898`; L2/L3 opt-in NOT used — isolated agent worktree). Milestone commits are **worktree-branch SHAs** (rebased onto main and landed; SHAs below are main SHAs): M1 `83b72e036`, M2 `31e44513a`, M3 `c6a19dbc1`, M4 (this commit).

### AC Matrix (AC-WC13-001..032)

| AC | REQ | Status | Verification command | Actual output |
|----|-----|--------|----------------------|---------------|
| AC-WC13-001 | 001/002 | PASS | `grep -n '"handoff"\|"cache"' internal/settings/sectionroute.go` | `79:"handoff": RouteSeam` + `80:"cache": RouteSeam` |
| AC-WC13-002 | 002 | PASS | `awk '/func ExcludedSections/,/^}/' … \| grep -c '"cache"'` | `0` |
| AC-WC13-003 | 002 | PASS | `awk '/func SeamSections/,/^}/' … \| grep -c 'handoff\|cache'` | `1` (both on one line) |
| AC-WC13-004 | 003 | PASS | `awk '/sectionRootKeys = map/,/^}/' … \| grep -c '"handoff"\|"cache"'` | `2` |
| AC-WC13-005 | 004 | PASS | `go test ./internal/settings/ -run 'TestRouteForSection\|TestSeamSections\|TestExcludedSections\|TestWriteSectionViaSeam'` | `ok  internal/settings` (accept new scope + reject remaining excl group + db/research) |
| AC-WC13-006 | 005 | PASS | `grep -rn 'ConfigManager\|\.Save(' sectionwrite.go sectionvalues.go \| grep -iv 'seam\|comment\|//'` | (no match — no typed re-marshal) |
| AC-WC13-007 | 006 | PASS | `go test ./internal/settings/ -run TestCacheSeamPreservesUnexposedKeys` | `ok` (spec_ttl/min_cacheable_tokens/comments preserved) |
| AC-WC13-010 | 010 | PASS | `grep -c 'SectionHandoff\|SectionCache' internal/settings/schema.go` | `4` (SectionID decl + AllSections) |
| AC-WC13-011 | 010/012 | PASS | `go test ./internal/settings/ -run TestHandoffCacheFields` (derived count) | `ok` — handoff 2 (mode select, guide bool) + cache 2 (enabled bool, session_ttl select), all PersistSeam |
| AC-WC13-012 | 011 | PASS | `go test ./internal/web/ -run TestHandoffModeValidation` | `ok` — mode=bogus → 400 + file unchanged; mode=auto → 200 |
| AC-WC13-013 | 013 | PASS | `go test ./internal/settings/ -run TestSessionTTLClosedSetSymmetry` + config `TestValidSessionTTLs` | `ok` — settings option set ≡ config.ValidSessionTTLs() = {1h,5m,off} |
| AC-WC13-014a | 014 | PASS | `for k in sec.handoff sec.cache; do grep -c "\"$k\." i18n.js; done` | `8` each (2 keys × 4 locales) |
| AC-WC13-014b | 014 | PASS | `go test ./internal/web/ -run TestI18n` | `ok` — 4-locale parity |
| AC-WC13-015 | 015 | PASS | `awk '/func cacheFields/,/^}/' … \| grep -v '^[[:space:]]*//' \| grep -c 'spec_ttl\|min_cacheable_tokens'` | `0` |
| AC-WC13-016 | 016 | PASS | `go test ./internal/cli/ -run TestI18nKeySetParity` (schema_bridge_test.go UNCHANGED) | `ok` — PersistSeam predicate (`isWebOnlyKeyChipField`) scopes new fields out of TUI parity → no TUI code change (branch: PersistSeam exclusion) |
| AC-WC13-017 | 017 | PASS | new templ file check — only modelpolicy.templ added (M3), no per-section handoff/cache templ | generic fieldset pipeline reused; no bespoke handoff/cache template |
| AC-WC13-020 | 020 | PASS | `go test ./internal/web/ -run TestModelPolicyView` | `ok` — perfTier display + 3×12 table (3 mp-table blocks; S/M/L × plan/run/sync/mx) |
| AC-WC13-021 | 021 | PASS | `grep -c '<form\|<input\|<select\|hx-post' internal/web/modelpolicy.templ internal/web/modelpolicy_templ.go` | `0` + `0`; `TestModelPolicyView_ReadOnlyNoForm` + `_GETOnly` PASS |
| AC-WC13-022 | 021 | PASS | `grep -rn 'model_routing_profiles\|performance_tier' schema_sections.go sectionapply.go \| grep -v comment` | (no match — no persist binding); `TestModelPolicyView_NoPersistBinding` PASS |
| AC-WC13-023 | 022 | PASS | `go test ./internal/web/ -run TestModelPolicyView_LegacyFlatHidden` | `ok` — legacy flat model_routing + workflow_agents NOT rendered |
| AC-WC13-024 | 023 | PASS | `awk '/func agentSettingsFields/,/^}/' … \| grep -c 'workflow_agents'` | `0` (M5-a B1 retained — no FieldDef re-add) |
| AC-WC13-025 | 024 | PASS | `go test ./internal/web/ -run TestModelPolicyView_EmptyTier` | `ok` — empty performance_tier → "(runtime default: medium)" |
| AC-WC13-026 | 026 | PASS | `go test ./internal/web/ -run TestModelPolicyView_AbsentBlock` | `ok` — absent block → fallback state (all inherit/medium), no error |
| AC-WC13-027 | 025 | PASS | `go test ./internal/web/ -run TestModelPolicyView_I18nParity` | `ok` — mp.* keys 4-locale + model_policy vs performance_tier disambiguation note |
| AC-WC13-030 | 030 | PASS | `git diff --stat 99ad0a898..HEAD -- internal/statusline/` | (empty — no touch) |
| AC-WC13-031 | 031 | PASS | `git diff --stat 99ad0a898..HEAD -- internal/template/templates/` | (empty — no touch; template mirrors already carry all exposed keys) |
| AC-WC13-032a | 032 | PASS | `go build ./... && GOOS=windows GOARCH=amd64 go build ./...` | `darwin build exit=0` + `windows build exit=0` |
| AC-WC13-032b | 032 | PASS | `go test ./internal/settings/... ./internal/web/... ./internal/cli/... ./internal/config/...` | all `ok` (full touched-package suite) |
| AC-WC13-032c | 032 | PASS | `golangci-lint run --timeout=3m` | `0 issues.` (NEW 0; pre-flight baseline was also `0 issues.`) |

### Coverage (touched code — function-level)

| Function | File | Coverage |
|----------|------|----------|
| RouteForSection / SeamSections / ExcludedSections | sectionroute.go | 100% / 100% / 100% |
| WriteSectionViaSeam | sectionwrite.go | 90.0% |
| handoffFields / cacheFields | schema_sections.go | 100% / 100% |
| consoleTabs / schemaSectionMetas | schemaform.go | 100% / 100% |
| ValidSessionTTLs / stringSet | config/cache_config.go | 100% / 100% |
| modelPolicyPerfTiers/SpecTiers/Phases | modelpolicy.go | 100% each |
| buildModelPolicyView / handleModelPolicy / renderModelPolicy | modelpolicy.go | 86.7% / 75.0% / 71.4% |

Per-package totals: settings 88.7%, config 80.7% (whole-package — legacy untested surface dominates), web 70.0% (whole-package — legacy). All *modified* functions ≥ 71%; new M3 view + M1/M2 SSOT functions ≥ 86-100%.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-10
run_commit_sha: 819532241   # M4 run-completion commit (worktree branch; backfilled per spec-frontmatter-schema.md D3 exemption)
run_status: PASS
ac_pass_count: 29        # AC-WC13-001..032 including a/b sub-letters (25 AC IDs)
ac_fail_count: 0
preserve_list_post_run_count: 0   # statusline + template trees byte-unchanged (AC-030/031)
l44_pre_commit_fetch: not-applicable   # isolated worktree branch; no push per delegation constraint
l44_post_push_fetch: not-applicable    # push withheld per constraint (Do NOT push)
new_warnings_or_lints_introduced: 0    # golangci-lint 0 issues (baseline 0 issues)
cross_platform_build:
  darwin: exit-0
  windows: exit-0   # GOOS=windows GOARCH=amd64
total_run_phase_files: 21   # git diff --stat 99ad0a898..HEAD name-only
m1_to_mN_commit_strategy: per-milestone (M1 83b72e036 / M2 31e44513a / M3 c6a19dbc1 / M4 this)
known_debt:
  - internal/hook TestHookWrapper_TempFileCleanup — pre-existing intermittent flake (passes in isolation; out-of-scope per delegation)
  - B-1 (plan §B): internal/config/types.go PerformanceTier validator oneof=high|medium|low vs template ValidPerformanceTiers {max,medium,low} mismatch — pre-existing, NOT modified (Model Policy view shows raw value, did not block display); resolution out of scope
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-10
sync_commit_sha: a22f6cade   # backfilled — the sync commit that carried this §E.4 signal + CHANGELOG entry + frontmatter transition
sync_status: PASS
changelog_entry_position: "Unreleased > Changed (new section, top of file, immediately after '## [Unreleased]')"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed"
  plan_md: "no frontmatter (body-only status marker untouched per body-content ownership boundary)"
  acceptance_md: "no frontmatter (untouched per body-content ownership boundary)"
  progress_md: "n/a (progress.md carries no status: field)"
readme_check: "grep -n -i 'moai web|web console' README.md -> no matches; README documents no web-console config sections (consistent with SPEC-WEB-CONSOLE-012 precedent) -> no README edit required"
known_debt_recorded_in_changelog:
  - "internal/hook TestHookWrapper_TempFileCleanup — pre-existing intermittent flake under load, passes in isolation, out-of-scope"
  - "scripts/i18n-validator TestBudget_FullRepoScanWithin35Sec — 53.66s vs 35s budget under parallel load, package never touched by this SPEC"
lint_status: "golangci-lint run --timeout=2m -> 0 issues"
touched_package_tests: "all green (internal/settings, internal/web, internal/cli, internal/config)"
```

Sync commit lands on `main` (Route A — Hybrid Trunk main-direct, Tier M). CHANGELOG entry added under a new `### Changed` section at the top of `[Unreleased]` (per the repeated-header pattern already established in this file). spec.md frontmatter `status: in-progress -> completed` (+ `updated: 2026-07-10`, unchanged date since same-day close). plan.md/acceptance.md carry no YAML frontmatter (matching the SPEC-WEB-CONSOLE-012 precedent) — no frontmatter edit possible or made; body content untouched per the manager-docs body-content ownership boundary.
