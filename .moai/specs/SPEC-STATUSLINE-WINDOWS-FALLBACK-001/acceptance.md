# Acceptance Criteria — SPEC-STATUSLINE-WINDOWS-FALLBACK-001

All criteria are mechanically verifiable. Each AC names the exact command/assertion and the expected observable result. Paths are relative to repo root `/Users/goos/MoAI/moai-adk-go`.

## §A. Given-When-Then Scenarios

### Scenario 1 — Windows installer machine, Go absent (primary defect)

- **Given** a `TemplateContext` whose `GoBinPath` is a non-existent directory (e.g. the default `go env GOPATH/bin` on a Windows installer machine) AND whose `ResolvedMoaiPath` is the running binary's path (`%LOCALAPPDATA%\Programs\moai\moai.exe`),
- **When** `status_line.sh` is rendered from the template,
- **Then** the rendered script contains a guarded resolved-executable branch referencing `ResolvedMoaiPath` and does NOT contain any unguarded/baked `GoBinPath` branch — so the statusline resolves the real binary instead of the phantom path.

### Scenario 2 — Unix machine, binary on PATH (no regression)

- **Given** a Unix machine where `command -v moai` resolves the binary,
- **When** the rendered `status_line.sh` runs,
- **Then** Stage 1 (`command -v moai`) succeeds and `exec`s before any fallback — behavior identical to before the fix (no regression).

### Scenario 3 — `os.Executable()` error at init/update (graceful degradation)

- **Given** `os.Executable()` returned an error at init/update time, leaving `ResolvedMoaiPath` empty,
- **When** `status_line.sh` is rendered,
- **Then** the resolved-executable branch is OMITTED entirely (no empty-valued `exec` line) and the script falls through to the `$HOME`-relative guarded branches (REQ-SWF-005).

### Scenario 4 — Reproduction-first discipline

- **Given** the pre-fix template,
- **When** the M1 reproduction test runs against it,
- **Then** the test FAILS (asserting the phantom baked path is present / the guarded resolved-executable branch is absent), and after the M3 fix the same test PASSES (REQ-SWF-010, CLAUDE.md §7 Rule 4).

## §B. Mechanical AC Matrix

| AC ID | Bound REQ | Verification command | Expected result |
|-------|-----------|----------------------|-----------------|
| AC-SWF-001 | REQ-SWF-004 | `grep -n 'GoBinPath' internal/template/templates/.moai/status_line.sh.tmpl` | No match (template no longer bakes `.GoBinPath`) |
| AC-SWF-002a | REQ-SWF-002 | `grep -n 'ResolvedMoaiPath' internal/template/context.go` | Field declaration + `WithResolvedMoaiPath` option present |
| AC-SWF-002b | REQ-SWF-002, REQ-SWF-003 | `grep -n 'ResolvedMoaiPath' internal/template/templates/.moai/status_line.sh.tmpl` | Resolved-executable branch present, positioned after `command -v moai` and before `$HOME` branches |
| AC-SWF-003 | REQ-SWF-001, REQ-SWF-006 | Inspect rendered `status_line.sh`: every `exec ... moai statusline` line is preceded by `command -v moai` success OR a `[ -f "..." ]` guard | No `exec` line reachable via an unguarded non-existent path |
| AC-SWF-004 | REQ-SWF-005 | Unit test: render with `ResolvedMoaiPath=""` | Rendered output contains NO resolved-executable branch and NO empty-valued `exec`; `$HOME` branches present |
| AC-SWF-005 | REQ-SWF-005 | Unit test: render with `ResolvedMoaiPath="/some/path/moai"` | Rendered output contains `[ -f ".../moai" ]` guarded branch with that path (posix form) |
| AC-SWF-006 | REQ-SWF-002, REQ-SWF-003 | `grep -c 'WithResolvedMoaiPath' internal/core/project/initializer.go internal/cli/update.go internal/cli/update_clean_install.go` (sum across files) | ≥ 4 — all 4 `WithGoBinPath` injection sites wired (init ×1, update ×2 at `update.go:650`/`:688`, clean-install ×1) |
| AC-SWF-007 | REQ-SWF-007 | `GOOS=windows GOARCH=amd64 go build ./...` | Exit 0 (Windows cross-compile passes) |
| AC-SWF-008 | REQ-SWF-008 | `go test ./internal/template/... -run TestTemplateNeutralityAudit` | PASS (no internal-content / language-bias leak in `templates/`) |
| AC-SWF-009 | REQ-SWF-009 | `go test -cover ./internal/template/ ./internal/runtime/gobin/` then compare to baseline (template 84.6%, gobin 70.0%) | Changed-package coverage ≥ baseline; new logic directly asserted by tests |
| AC-SWF-010 | REQ-SWF-010 | Git history / run-phase evidence shows the M1 reproduction test failing on the pre-fix tree, then passing after M3 | RED-then-GREEN evidence present |
| AC-SWF-011 | REQ-SWF-001 (Template-First) | `make build` then `git status internal/template/embedded.go` | `embedded.go` regenerated to reflect the new template |
| AC-SWF-012 | full suite | `go test ./...` | All packages PASS |
| AC-SWF-013 | REQ-SWF-011 (RT-007-003 supersession) | `go test ./internal/template/ -run TestFallbackChainOrder` AND inspect `TestFallbackChainOrder` body (it MUST render the template with a known `ResolvedMoaiPath` context and assert on the rendered output) AND `grep -n 'GoBinPath' internal/template/hardcoded_path_audit_test.go` | `TestFallbackChainOrder` PASSES with a REAL render-and-assert body (renders `.moai/status_line.sh.tmpl` via `Renderer.Render` with a populated `ResolvedMoaiPath` and asserts the rendered output contains the four-stage ascending chain `command -v moai` → resolved-executable → `$HOME/go/bin/moai` → `$HOME/.local/bin/moai` in order, PLUS a negative assertion that the rendered output contains no `GoBinPath` token). A doc-comment-only edit, or a body that does not call `Renderer.Render` on the actual template, does NOT satisfy this AC — the pre-fix test was exactly such an inert empty-body stub and the AC is deliberately strengthened so the gate cannot regress to it. |

## §C. Edge Cases

- **Empty `GoBinPath` AND empty `ResolvedMoaiPath`**: script must still render valid bash that falls through to `$HOME` branches and finally `exit 0` (no dangling `if`).
- **`ResolvedMoaiPath` containing backslashes (Windows)**: `posixPath` must convert to forward slashes so the git-bash `[ -f ]` test works (AC-SWF-005 verifies posix form).
- **Path containing spaces** (`C:\Users\First Last\...`): the `[ -f "..."]` guard and `exec "..."` must keep the path double-quoted in the template.

## §D. Definition of Done

- [ ] AC-SWF-001 … AC-SWF-013 all PASS (mechanical evidence captured in progress.md §E.2/§E.3)
- [ ] M1 reproduction test written and observed to FAIL before the fix (AC-SWF-010)
- [ ] Template returned to §14 compliance — no baked `.GoBinPath` (AC-SWF-001)
- [ ] Resolved-executable branch present, guarded, conditionally omitted when empty (AC-SWF-002b/004/005)
- [ ] All 4 render call sites wired (AC-SWF-006)
- [ ] `TestFallbackChainOrder` UPDATED to new chain order, not silently broken (AC-SWF-013, RT-007-003 supersession)
- [ ] Windows cross-compile green (AC-SWF-007)
- [ ] Neutrality audit green (AC-SWF-008)
- [ ] Changed-package coverage ≥ baseline; new logic directly tested (AC-SWF-009)
- [ ] `make build` regenerated `embedded.go` (AC-SWF-011)
- [ ] Full test suite green (AC-SWF-012)
- [ ] Scope discipline: `git diff` shows NO change to `internal/runtime/gobin/resolver.go` `Detect` body and NO change to `internal/shell/env.go`

## §E. Quality Gate Criteria

- LSP: zero errors, zero type errors, zero lint errors (run-phase threshold).
- `golangci-lint run --timeout=2m` clean.
- No internal-content leak (`internal_content_leak_test.go` / neutrality audit green).
