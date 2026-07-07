# Acceptance — SPEC-CLI-UIKIT-KERNEL-001

Given-When-Then acceptance criteria for the uikit kernel extraction. Because this
is a behavior-preserving refactor of working code, the acceptance model is
CHARACTERIZATION: the existing test suite + cross-platform build + `moai --help`
output are the behavior contract M1 must preserve. The AC structure mirrors
SPEC-CLI-SUBPKG-SPLIT-001 acceptance.md §A-§C (AC-CSS-001..012); this file
specializes each AC for the single-cluster uikit case.

## §A. Per-Milestone Behavior-Preservation Gate (applies to M1)

### AC-CUK-001 — Full test suite green after M1 (REQ-CUK-004)
- **Given** a completed uikit extraction (4 source files moved, 2 type-dependency resolutions
  applied, all callers rewritten, all test files relocated),
- **When** `go test ./...` is run,
- **Then** it exits 0 with zero NEW failures vs the documented pre-flight baseline — pre-existing
  baseline failures (e.g. `TestRunHookEvent_ReadInputError`, `TestAuthoringDocHasEffortMatrix`
  per SPLIT-001 progress.md §E.2) are documented separately and NOT counted as regressions.
- **Verification**: `go test ./... 2>&1 | tail -10` → compare pass/fail set to `/tmp/cli-before.txt`;
  every NEW failure blocks M1.

### AC-CUK-002 — Cross-platform build green after M1 (REQ-CUK-005)
- **Given** a completed M1 (4 source files moved; no platform-tagged siblings expected, verified
  at run-phase),
- **When** the cross-platform build matrix runs,
- **Then** both invocations exit 0.
- **Verification**: `go build ./...` → exit 0 AND `GOOS=windows GOARCH=amd64 go build ./...` →
  exit 0.

### AC-CUK-003 — `moai --help` subcommand list unchanged (REQ-CUK-003)
- **Given** a completed M1,
- **When** `moai --help` is captured,
- **Then** the subcommand names, groups, and order are identical to the pre-M1 snapshot.
- **Verification**: `go run ./cmd/moai --help > /tmp/help-after.txt; diff /tmp/help-before.txt
  /tmp/help-after.txt` → empty diff.

### AC-CUK-004 — Public entry point unchanged (REQ-CUK-008)
- **Given** the uikit extraction,
- **When** `cmd/moai/main.go` is inspected,
- **Then** it still calls `cli.Execute()` unchanged and builds.
- **Verification**: `grep -n 'cli.Execute' cmd/moai/main.go` present; `go build ./cmd/moai` exit 0.

## §B. Structural Correctness Gates

### AC-CUK-005 — uikit is a proper leaf package (REQ-CUK-001, REQ-CUK-002)
- **Given** the extracted uikit package at `internal/cli/uikit/`,
- **When** the package is inspected,
- **Then** every file declares `package uikit`, AND the package imports NEITHER
  `github.com/modu-ai/moai-adk/internal/cli` NOR any `github.com/modu-ai/moai-adk/internal/cli/*`
  subpackage.
- **Verification**:
  - `head -1 internal/cli/uikit/*.go | grep -c 'package uikit'` = file count;
  - `grep -rn '"github.com/modu-ai/moai-adk/internal/cli"' internal/cli/uikit/` → no match;
  - `grep -rn '"github.com/modu-ai/moai-adk/internal/cli/' internal/cli/uikit/` → no match
    (the second grep catches any `internal/cli/<sibling>` import — the leaf contract forbids it).

### AC-CUK-006 — No import cycle (REQ-CUK-001)
- **Given** the uikit extraction,
- **When** the build runs,
- **Then** no import cycle exists.
- **Verification**: `go build ./...` exit 0 (Go reports import cycles at build time); additionally
  `go list -deps ./internal/cli/uikit/... | grep 'internal/cli'` returns only `internal/cli/uikit`
  itself (no parent, no sibling).

### AC-CUK-007 — Type-dependency co-location complete (REQ-CUK-007)
- **Given** the uikit extraction with cross-file type dependencies,
- **When** the moved source files are inspected,
- **Then** (a) `CheckStatus` type + its 3 consts (`CheckOK`/`CheckWarn`/`CheckFail`) live in
  `internal/cli/uikit/` (e.g. `uikit/types.go`); AND (b) the `profileSetupText` b-ii split is
  applied — BOTH `schemaFieldBridge` (L24-58) AND `schemaSegmentBridge` (L60-77) stay in package
  cli (e.g. in a new `package cli` file `schema_bridge_profile.go`); only `SchemaKeyToTUIField`
  + `FieldDefTUILabel` move to uikit; the resolution is documented in the M1 commit message
  (D5 iter-1 fix: iter-1 mentioned only `schemaFieldBridge`; both maps reference
  profileSetupText); AND (c) `doctor.go` no longer DEFINES `CheckStatus` (it imports
  `uikit.CheckStatus`); AND (d) no MOVED source file references an undefined symbol; AND
  (e) **D2 iter-1 BLOCKING fix**: `SettingsLocal` cycle is resolved by `settings.go STAYS in
  package cli` (option a) — `settings.go` does NOT move to uikit; uikit does NOT import package
  cli; `mutateSettingsLocal`/`writeFileAtomic`/`stripGLMCredsAndSetTeammateMode` remain
  package-cli-internal.
- **Verification**:
  - `grep -rn 'type CheckStatus' internal/cli/*.go internal/cli/uikit/*.go` → exactly 1 match,
    inside `internal/cli/uikit/`;
  - `grep -rn 'CheckStatus\|CheckOK\|CheckWarn\|CheckFail' internal/cli/doctor.go internal/cli/doctor_cache.go internal/cli/doctor_harness.go | head -20` →
    all references are `uikit.CheckStatus` / `uikit.CheckOK` / etc. (no bare type ref) —
    D1 iter-1 fix: doctor_cache.go + doctor_harness.go are now in the verification set;
  - `grep -rn 'profileSetupText' internal/cli/uikit/ internal/cli/schema_bridge.go internal/cli/profile_setup_translations.go` → consistent with the b-ii resolution (profileSetupText-referencing maps stay in package cli; uikit has zero profileSetupText refs);
  - `grep -rn '"github.com/modu-ai/moai-adk/internal/cli"' internal/cli/uikit/` → no match (cycle pre-check);
  - `ls internal/cli/uikit/settings.go 2>&1` → "No such file or directory" (D2-a: settings.go STAYS).

### AC-CUK-008 — Caller-rewrite completeness (REQ-CUK-006)
- **Given** a completed M1 with all caller rewrites applied,
- **When** the `package cli` non-test files are grepped for residual pre-move symbol references,
- **Then** NO residual reference exists — every caller that used `renderCard(...)` now uses
  `uikit.RenderCard(...)`, etc.
- **Verification**: for each MOVED symbol, `grep -rnE '\b<symbol>\b' internal/cli/*.go | grep -v
  _test.go | grep -v '^internal/cli/root.go.*AddCommand'` returns no match (the symbol no longer
  appears unqualified in `package cli`).
  - Concrete MOVED symbols to verify: `renderCard`, `renderKeyValue`, `renderKeyValueLines`,
    `renderStatusLine`, `renderSuccessCard`, `renderInfoCard`, `renderSummaryLine`, `RenderError`,
    `cardStyle`, `PrintBanner`, `PrintWelcomeMessage`, `printWelcomeMessage`,
    `schemaKeyToTUIField`, `fieldDefTUILabel`, `kvPair`, `CheckStatus` (when defined in
    doctor.go pre-move — post-move doctor.go / doctor_cache.go / doctor_harness.go / 5
    CheckStatus-bearing test files all use the `uikit.` qualifier per D1 iter-1 fix),
    `CheckOK`, `CheckWarn`, `CheckFail`.
  - **D2 iter-1 fix — NOT verified by this AC** (these symbols STAY in package cli under
    D2-a, so their unqualified references are EXPECTED): `mutateSettingsLocal`,
    `writeFileAtomic`, `stripGLMCredsAndSetTeammateMode`. The settings helpers' call sites
    (e.g. `launcher.go:206`) remain `mutateSettingsLocal(...)` / `stripGLMCredsAndSetTeammateMode`
    unqualified — they did not move. The run phase does NOT rewrite them.

### AC-CUK-009 — No AskUserQuestion introduced (subagent boundary, C-HRA-008, REQ-CUK-013)
- **Given** the extracted uikit package,
- **When** grepped,
- **Then** no `AskUserQuestion`/`mcp__askuser` call exists.
- **Verification**: `grep -rnE 'AskUserQuestion|mcp__askuser' internal/cli/uikit/ | grep -v
  _test.go | grep -v '^[^:]*:[0-9]*:[ \t]*//'` → no output.

### AC-CUK-010 — uikit exports the helpers its callers need (REQ-CUK-002)
- **Given** the extracted uikit package,
- **When** its exported surface is inspected,
- **Then** every helper referenced by any `package cli` caller (post-rewrite) is exported from
  uikit; helpers with zero external callers remain unexported.
- **Verification**: `go doc ./internal/cli/uikit` lists the exported helpers; cross-check against
  the caller-rewrite grep that every `uikit.<X>` reference resolves to an exported `X`.

## §C. Phasing Discipline Gates

### AC-CUK-011 — Single atomic commit for M1 (REQ-CUK-009)
- **Given** the M1 commit,
- **When** its diff is inspected,
- **Then** it touches exactly: the new `internal/cli/uikit/` directory (**3 moved source files**
  — render.go + banner.go + schema_bridge.go's helper portion under b-ii split; `settings.go`
  STAYS per D2 option a) + any new file like `types.go` for CheckStatus + the rewritten caller
  files (~12 production files + ≥10 test files per D3/D4 iter-1 fix, including doctor_cache.go +
  doctor_harness.go + the 5 CheckStatus-bearing test files) + the moved test files (e.g.
  `render_test.go` + the misc_coverage_test.go PrintWelcomeMessage block per design.md §D.10) +
  the `doctor.go`/`doctor_cache.go`/`doctor_harness.go` CheckStatus reference rewrites — NO
  second cluster's files (no migrate_agency.go logic change, no profile_setup.go move, no
  doctor.go cluster extraction, no update.go cluster extraction, NO settings.go move).
- **Verification**: `git show --stat <M1-sha>` shows one `internal/cli/uikit/` dir + the caller
  files; `git show <M1-sha> -- internal/cli/migrate_agency.go` shows ONLY the
  `RenderError` → `uikit.RenderError` rewrite (one-line change), NOT a cluster move;
  `git show <M1-sha> -- internal/cli/settings.go` shows NO changes (D2-a: settings.go STAYS);
  `git show <M1-sha> -- internal/cli/launcher.go` shows NO `mutateSettingsLocal`/`stripGLMCredsAndSetTeammateMode` rewrite (D2-a: these helpers stay in package cli).

### AC-CUK-012 — No functional change / no test weakening (REQ-CUK-011, REQ-CUK-012)
- **Given** the M1 diff,
- **When** reviewed,
- **Then** no logic edit beyond symbol re-scoping (export/unexport) + type co-location + import
  rewiring; no test deleted or skipped; test count is preserved (moved, not removed).
- **Verification**: `git show <sha>` — non-test production diffs are `package`/export/import
  lines + `uikit.` qualifier additions; test-file count before == after (moved, not deleted);
  `grep -rn 't.Skip' internal/cli/uikit/ internal/cli/*.go` unchanged vs baseline.

### AC-CUK-013 — Checkpoint stop-condition honored (REQ-CUK-010)
- **Given** the post-M1 checkpoint,
- **When** the marginal value of future kernel-dependent cluster SPECs (migrate/doctor/update)
  is judged insufficient,
- **Then** the work stops with M1 shipped and the decision is recorded in `progress.md §E.4` —
  no forced authoring of follow-up cluster SPECs.
- **Verification**: checkpoint decision recorded; if STOP chosen, no migrate/doctor/update SPEC
  is authored from THIS SPEC's scope (they remain independent future SPECs gated on their own
  coupling-resolution design).

## §D. Quality Gates (M1)

- `go test ./...` exit 0 (AC-CUK-001, allowing documented pre-existing baseline failures) — the
  binding gate.
- `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (AC-CUK-002).
- `go vet ./...` exit 0.
- `golangci-lint run --timeout=2m` — no NEW issues vs the pre-flight baseline (pre-existing
  baseline reported separately).
- Coverage: per-package coverage of the moved helpers (in `internal/cli/uikit/`) ≥ the pre-move
  coverage (tests moved intact, so coverage is preserved by construction — verify not regressed).

## §E. Definition of Done

M1 is DONE when:
1. AC-CUK-001..004 pass (behavior preserved: tests green, cross-platform build green,
   `moai --help` diff empty, public entry point unchanged).
2. AC-CUK-005..010 pass (proper leaf package, no import cycle, type-dependency co-location
   complete, caller-rewrite completeness, no AskUserQuestion, exports align with callers).
3. AC-CUK-011..012 pass (atomic single-M1 commit, no functional change, no test weakening).
4. `go vet` + `golangci-lint` clean of NEW issues.
5. The M1 commit is committed with a Conventional-Commit subject
   `refactor(SPEC-CLI-UIKIT-KERNEL-001): M1 extract uikit kernel to internal/cli/uikit`.

The SPEC as a whole is DONE when M1 ships AND the §F CHECKPOINT decision (AC-CUK-013) is
recorded — NOT necessarily when follow-up cluster SPECs are authored (those are independent
future SPECs, each with their own scope).
