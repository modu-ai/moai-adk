# Progress — SPEC-WEB-CONSOLE-001: MoAI Web Console (Settings CRUD)

## Run-phase summary

Run-phase implementation via `cycle_type=tdd` (RED-GREEN-REFACTOR). New Go package
`internal/web/` (loopback-only HTTP server + handlers + Host-check middleware +
`go:embed` assets) plus a thin `web` cobra subcommand registered in
`internal/cli/root.go`. Persistence and validation reuse the existing
`internal/profile` read/write/sync functions and canonical value lists — zero
parallel validation, zero direct YAML writes.

Tier: M. plan-auditor verdict: PASS 0.91. GATE-2: APPROVED.

## §E.2 Run-phase Evidence

### AC PASS/FAIL matrix (acceptance.md SSOT — 12 ACs)

| AC | Requirement | Status | Verification Command | Actual Output |
|----|-------------|--------|----------------------|---------------|
| AC-WC-001 | REQ-WC-001 | PASS | `go run ./cmd/moai web --help` | exit 0; help lists `--port` (int, default 8080) + `--no-open` (bool); `TestWebCmd_FlagsRegistered`/`TestWebCmd_HelpListsFlags` PASS |
| AC-WC-002 | REQ-WC-002 | PASS | `go test -run TestServer_LoopbackBindOnly ./internal/web` | listener bound to `127.0.0.1:<port>`; asserted `ip.IsLoopback()` true AND not `0.0.0.0`/unspecified — PASS |
| AC-WC-003 | REQ-WC-003 | PASS | `go test -run 'TestServer_GracefulShutdown' ./internal/web` | ctx-cancel AND real SIGTERM → `Shutdown(5s)` → returns nil within drain window — PASS |
| AC-WC-004 | REQ-WC-004 | PASS | `go test -run TestServer_AutoOpen ./internal/web` | `--no-open`→opener never called; open-failure→server keeps serving (GET / succeeds), returns nil — PASS |
| AC-WC-005 | REQ-WC-005 | PASS | `go test -run TestStaticAssetsServedFromEmbed ./internal/web` | `/static/style.css` + `/static/app.js` served from `go:embed` with correct Content-Type, no network fetch — PASS |
| AC-WC-006 | REQ-WC-006 | PASS | `go test -run TestIndexRendersPopulatedForm ./internal/web` | `GET /` renders pre-populated form via `ReadPreferences` (`value="Goos"`, `conversation_lang=ko` selected) — PASS |
| AC-WC-007 | REQ-WC-007 | PASS | `go test -run 'TestSaveValidRoundTrip\|TestGoldenPath_ReadWriteRoundTrip' ./internal/web` | valid POST → `WritePreferences` + `SyncToProjectConfig` → read-back confirms `ConversationLang=ko`, `StatuslineTheme=catppuccin-latte` persisted; language.yaml/statusline.yaml updated; no direct YAML write — PASS |
| AC-WC-008 | REQ-WC-008 | PASS | `go test -run 'TestSaveInvalid\|TestValidatePrefs' ./internal/web` | invalid PermissionMode/theme → 400, persisted state unchanged, per-field error rendered, reuses `IsValidPermissionMode` + canonical lists — PASS |
| AC-WC-009 | REQ-WC-009 | PASS | `go test -run 'TestHostCheck\|TestIsLoopbackHost' ./internal/web` | foreign Host on POST → 403 + state unchanged; loopback Hosts (127.0.0.1/localhost/[::1]) pass; GET not Host-gated — PASS |
| AC-WC-010 | REQ-WC-010 | PASS | `go test -run 'TestIndexNeutralDefaults\|TestIndexReadError\|TestGoldenPath_EmptyState' ./internal/web` | zero-value profile → neutral `(unset)` defaults form (not error); read error → inline error (never blank, never panic) — PASS |
| AC-WC-011 | REQ-WC-011 | PASS | `go test -run TestIndexProfileSelection ./internal/web` | >1 profile → selector + current marker + selectable; single profile → selector omitted; `?profile=` query selects profile — PASS |
| AC-WC-012 | REQ-WC-012 | PASS | `go test -run 'TestSaveScopeBoundary\|TestGoldenPath_ReadWriteRoundTrip' ./internal/web` | persistence calls exactly `[WritePreferences, SyncToProjectConfig]`; out-of-scope `workflow.yaml`/`harness.yaml`/`git-strategy.yaml` sentinels intact after WRITE — PASS (see Observation O1) |

### Edge cases (acceptance.md §Edge Cases)

| EC | Description | Status | Evidence |
|----|-------------|--------|----------|
| EC-1 | Port already in use → clear error, non-panic | PASS | `TestServer_PortConflictReturnsError` — bind error mentions "bind", no panic |
| EC-2 | Concurrent submits last-writer-wins | PASS (delegated) | Atomicity owned by existing `WritePreferences`/`SyncToProjectConfig`; web layer adds no shared mutable state |
| EC-3 | Deprecated `Bypass` → `PermissionMode` migration | PASS (delegated) | `ReadPreferences` performs the migration; web READ shows the migrated value |
| EC-4 | `StatuslineSegments` map round-trip | PASS | `TestSaveCustomSegmentsRoundTrip` — checked segments true, unchecked false (not dropped), all 15 keys present |
| EC-5 | Empty segments + absent statusline.yaml → defaults | PASS | `TestSaveNonCustomPresetLeavesSegmentsNil` — non-custom preset leaves segments nil so `syncStatusline` applies defaults |

### Quality gate results

| Gate | Result |
|------|--------|
| `go build ./...` | exit 0 |
| `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| `GOOS=linux GOARCH=amd64 go build ./...` | exit 0 |
| `go test ./...` (full suite) | all green (no FAIL) |
| `go test -race ./internal/web/...` | green (race-clean after listener mutex fix) |
| `go test -cover ./internal/web/...` | **90.5%** (≥ 85% threshold) |
| `golangci-lint run ./internal/web/... ./internal/cli/` | 0 issues |
| `go vet ./...` | clean |
| Subagent boundary grep (`AskUserQuestion`/`mcp__askuser` in `internal/web` + `internal/cli/web.go`, non-test, non-comment) | 0 matches |
| `moai web --help` | exit 0, lists `--port` + `--no-open` |

### Plan-auditor advisory D1 (MINOR, optional)

Implemented: `TestSaveSyncFailureSurfacesReadableError` asserts that a
`SyncToProjectConfig` failure after a successful `WritePreferences` surfaces a
readable partial-state error ("profile preferences saved, but project config
sync failed: ...") rather than silent success.

### MX tags placed (plan.md §G)

| Tag | Location | Status |
|-----|----------|--------|
| `@MX:ANCHOR` | `internal/web/server.go` `Run` (server entry, loopback-bind + graceful-shutdown invariant, fan_in≥3) | placed (+ `@MX:REASON`) |
| `@MX:WARN` | `internal/web/handlers.go` `handleSave` (sole on-disk settings mutation path) | placed (+ `@MX:REASON`) |
| `@MX:NOTE` | `internal/web/app.go` `hostCheckMiddleware` (deliberate minimal write-safety model) | placed |
| `@MX:TODO` | per-phase incompletion hooks | none unresolved (all phases passed) |

## Observations (run-phase, non-blocking)

### O1 — REQ-WC-012 scope-boundary nuance (existing-tool property, not a web-layer defect)

`SyncToProjectConfig` → `config.ConfigManager.Save()` atomically re-serializes the
sibling config sections it models (`user.yaml`, `language.yaml`, `quality.yaml`,
`git-convention.yaml`, `llm.yaml`) whenever an in-scope value changes. This is the
documented behavior of the existing canonical persistence path (`internal/config`),
not a behavior introduced by the web layer. The Console itself reads/writes ONLY
the named scope (profile preferences + user/language/statusline); it never
deliberately touches out-of-scope sections, and `workflow.yaml` / `harness.yaml` /
`git-strategy.yaml` are genuinely never written (verified byte-intact in
`TestGoldenPath_ReadWriteRoundTrip`). The sibling re-serialization of
`quality.yaml`/`git-convention.yaml`/`llm.yaml` round-trips their modeled fields
without data loss in a real project (no data outside the struct schema). REQ-WC-007
mandates reuse of `SyncToProjectConfig` (direct YAML writes are a forbidden
anti-pattern), so this transitive `Save()` behavior is the prescribed path. No
scope change is requested; documented here for sync-phase awareness.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-03
run_commit_sha: b1ab60454
run_status: implemented
ac_pass_count: 12
ac_fail_count: 0
edge_case_pass_count: 5
preserve_list_post_run_count: 0   # internal/profile + internal/config + internal/cli/profile_setup untouched
l44_pre_commit_fetch: "git fetch origin main → 0 0 (synced, worktree HEAD == origin/main ef9a619ad)"
l44_post_push_fetch: "n/a — orchestrator owns push timing (commits left local per delegation prompt)"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: exit-0
  windows_amd64: exit-0
  linux_amd64: exit-0
total_run_phase_files: 13   # 8 new source/asset + web_test.go + 4 web test files + 2 cli edits (web.go new, root.go via AddCommand, help.go) + spec.md frontmatter + progress.md
m1_to_mN_commit_strategy: "single M1 commit (Tier M, cohesive new package); status draft→in-progress on M1"
package_coverage_internal_web: "90.5%"
```

## Files created / modified (run-phase scope)

### Created
- `internal/web/server.go` — HTTP server entry (`Run`/`NewServer`/`ListenAndServe`), loopback bind, graceful shutdown, listener mutex (@MX:ANCHOR)
- `internal/web/browser.go` — cross-platform default-browser opener (no build tags)
- `internal/web/app.go` — routes, Host-check middleware (@MX:NOTE), loopback-host classification, error rendering
- `internal/web/handlers.go` — READ (`handleIndex`) + WRITE (`handleSave`, @MX:WARN) handlers, form binding, view-model
- `internal/web/validate.go` — canonical value lists + `validatePrefs` (reuses `profile.IsValidPermissionMode`)
- `internal/web/assets.go` — `go:embed` declaration + page-template parsing
- `internal/web/assets/page.html.tmpl` — HTML page shell (html/template)
- `internal/web/assets/style.css` — embedded CSS
- `internal/web/assets/app.js` — embedded vanilla-JS progressive enhancement
- `internal/web/server_test.go`, `handlers_test.go`, `integration_test.go`, `helpers_test.go`, `coverage_test.go` — tests
- `internal/cli/web.go` — thin `web` cobra subcommand
- `internal/cli/web_test.go` — CLI flag + NoAskUserQuestion guard tests

### Modified
- `internal/cli/root.go` — (via `newWebCmd()` `init()` `rootCmd.AddCommand`) — web subcommand registered
- `internal/cli/help.go` — added `moai web` row to the Launcher (런처) help group for discoverability
- `.moai/specs/SPEC-WEB-CONSOLE-001/spec.md` — frontmatter `status: draft → in-progress` (manager-develop owned transition)
- `.moai/specs/SPEC-WEB-CONSOLE-001/progress.md` — this file

## Sync-phase Audit-Ready Signal

```yaml
sync_commit_sha: (this commit)
sync_status: complete
changelog_entry_added: true
status_transition: "in-progress → implemented"
version_bump: "0.1.0 → 0.2.0"
sync_executor: orchestrator-direct
run_commit_sha_correction: "39649c6c (stale worktree SHA) → b1ab60454 (actual run commit on origin/main)"
sync_rationale: >
  Mother SPEC of the web-console-v3 cohort. Run (single cohesive M1) already on
  origin/main; close-tail only. Orchestrator-direct sync/Mx (active parallel sessions;
  Authored-By-Agent trailer omitted = legacy silent SKIP to avoid OwnershipTransitionInvalid).
```
