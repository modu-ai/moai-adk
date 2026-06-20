# Implementation Plan — SPEC-STATUSLINE-WINDOWS-FALLBACK-001

> Tier M (standard). Bugfix touching ~5-6 files: 1 new `TemplateContext` field + option, 1 template, 4 render call sites (init + 2 update + clean-install), 1 existing test update (`TestFallbackChainOrder`), reproduction + fix tests, `make build`. No new subsystem, no architectural change (M, not L). More than 1-2 files (M, not S).

## §A. Context

GitHub issue modu-ai/moai-adk#1068: the generated `.moai/status_line.sh` bakes a phantom `C:\Users\{user}\go\bin\moai` path on Windows installer machines, so the Claude Code statusline silently renders empty. Root cause confirmed across three layers (spec.md §A.1); all verified against source, not inferred.

The fix is at the **template-render boundary**, NOT inside the shared `gobin.Detect` (which has a second legitimate consumer in `internal/shell/env.go`). Primary solution is issue option A (`os.Executable()` resolved path as first fallback), combined with option B (remove the baked branch, keep only `$HOME`-relative guarded branches).

## §B. Known Issues / Baselines (measured)

| Baseline | Command | Observed value |
|----------|---------|----------------|
| `internal/template` coverage | `go test -cover ./internal/template/` | 84.6% of statements |
| `internal/runtime/gobin` coverage | `go test -cover ./internal/runtime/gobin/` | 70.0% of statements |
| `gobin.Detect` fan_in (helper callers) | `grep -rn "gobin.Detect" --include="*.go"` (non-test) | 2 (init `initializer.go:301` + update `update.go:3054`) |
| `WithGoBinPath(...)` render sites (wiring targets — DISTINCT from fan_in) | `grep -n "WithGoBinPath" internal/core/project/initializer.go internal/cli/update.go internal/cli/update_clean_install.go` | **4**: `initializer.go:284`, `update.go:650`, `update.go:688`, `update_clean_install.go:276` |
| `os.Executable()` precedent | `grep -rn "os.Executable"` (non-test) | 3 existing usages (`internal/update/local.go:114`, `internal/cli/deps.go:283`, `internal/cli/update.go:451`) |
| `TestFallbackChainOrder` (RT-007-003 supersede target) | `grep -n "TestFallbackChainOrder\|RT-007-003" internal/template/hardcoded_path_audit_test.go` | lines 27-28 (must be UPDATED to new chain order, not silently broken) |

> **fan_in ≠ render-site count.** `gobin.Detect`'s `@MX:ANCHOR fan_in=2` counts the 2 callers of the `Detect` HELPER (`detectGoBinPath` in init, `detectGoBinPathForUpdate` in update). The number of `WithGoBinPath(...)` TemplateContext-injection sites is 4 — `update.go` invokes `WithGoBinPath` twice (lines 650 + 688) inside the update render function, and clean-install (`update_clean_install.go:276`) is a third file. All 4 injection sites must be wired with the new `WithResolvedMoaiPath(...)`; the 2 `detectGoBinPath*` helpers stay unchanged.

Honesty note: neither `internal/template` (84.6%) nor `internal/runtime/gobin` (70.0%) currently meets the §6 ≥90% critical-package target. This SPEC does NOT claim they do. The coverage gate (acceptance.md AC-SWF-009) requires the change to add direct coverage of the new logic and not regress the package baseline; lifting the whole package to 90% is a separate effort and is not in scope.

## §C. Pre-flight Checklist

- [ ] Read `internal/template/context.go` (struct + `With*` option pattern) before editing
- [ ] Read `internal/template/templates/.moai/status_line.sh.tmpl` before editing
- [ ] Read `internal/core/project/initializer.go` render call site + `internal/cli/update.go` + `internal/cli/update_clean_install.go`
- [ ] Confirm `os.Executable()` usage pattern from one of the 3 precedents
- [ ] Confirm `posixPath` is available in the template func map (`renderer.go:27`)

## §D. Constraints (binding)

1. Template-First Rule (CLAUDE.local.md §2): edit `templates/` then `make build`.
2. §14 HARD: generated shell fallbacks use `$HOME`, never baked `.GoBinPath`.
3. `GOOS=windows GOARCH=amd64 go build ./...` must pass.
4. 16-language neutrality + §25 internal-content isolation in `templates/`.
5. Coverage: changed statements ≥ baseline; new logic directly tested.
6. Scope: do NOT touch `gobin.Detect` internals or `internal/shell/env.go`.

## §E. Self-Verification Plan

Run-phase will verify via the canonical read-only batch (single-turn parallel):

```bash
# 1. Full test suite
go test ./...
# 2. Coverage (changed packages)
go test -cover ./internal/template/ ./internal/runtime/gobin/ ./internal/core/project/ ./internal/cli/
# 3. Cross-compile (Windows target)
GOOS=windows GOARCH=amd64 go build ./...
# 4. Template neutrality audit
go test ./internal/template/... -run TestTemplateNeutralityAudit
# 5. embedded.go regenerated + no drift
make build && git diff --exit-code internal/template/embedded.go || echo "embedded.go regenerated (expected on first build)"
# 6. Lint
golangci-lint run --timeout=2m
# 7. Rendered-output assertion (reproduction test target)
go test ./internal/template/... -run TestStatusline
```

## §F. Milestones (priority-ordered, no time estimates)

### M1 — Reproduction test (RED) [Priority: High]

Write a failing test (per REQ-SWF-010, CLAUDE.md §7 Rule 4) that renders `status_line.sh` from the template with a `TemplateContext` whose `GoBinPath` points at a non-existent directory, and asserts the rendered output does NOT contain an unvalidated `{{posixPath .GoBinPath}}`-style phantom branch AND DOES contain a guarded resolved-executable branch. Observe it FAIL against the current template. This anchors the bug before any fix.

- Files: `internal/template/*_test.go` (new test, e.g. `status_line_test.go` or extend existing renderer test)
- Exit: test compiles and FAILS for the documented reason.

### M2 — TemplateContext field + option [Priority: High]

Add the `ResolvedMoaiPath` field to `TemplateContext` (`internal/template/context.go`) and a `WithResolvedMoaiPath(path string) ContextOption`. Default empty string. Update `renderer_test.go` golden / context tests as needed.

- Files: `internal/template/context.go`, `internal/template/context_test.go` (or equivalent)
- Exit: field + option exist, unit-tested, package compiles.

### M3 — Template rewrite (§14 compliance + Option A/B) [Priority: High]

Rewrite `status_line.sh.tmpl` binary-resolution chain:
1. `command -v moai` (unchanged stage 1)
2. NEW: resolved-executable branch — guarded with `[ -f ]`, emitted ONLY when `{{if .ResolvedMoaiPath}}` is non-empty (REQ-SWF-005), using `{{posixPath .ResolvedMoaiPath}}` for forward-slash form (REQ-SWF-007)
3. `$HOME/go/bin/moai` (guarded `[ -f ]`, $HOME-relative per §14)
4. `$HOME/.local/bin/moai` (guarded `[ -f ]`, existing)
5. REMOVE the baked `{{posixPath .GoBinPath}}/moai` branch (REQ-SWF-004)
6. Final `exit 0`

Optionally (HOW decision, in-scope if judged necessary at run-phase) add a Windows installer-location branch (`%LOCALAPPDATA%`-derived). Default: rely on Option A's `os.Executable()` which already resolves the Windows installer path — keep the script minimal.

- Files: `internal/template/templates/.moai/status_line.sh.tmpl`
- Exit: template renders; M1 reproduction test now PASSES; §14 grep finds no `.GoBinPath` in the template.

### M4 — Wire ResolvedMoaiPath at all 4 render call sites [Priority: High]

At all 4 `WithGoBinPath(...)` injection sites, capture `os.Executable()` (handle the error by leaving the field empty per REQ-SWF-005) and additionally pass `WithResolvedMoaiPath(...)`:
- `internal/core/project/initializer.go:284` (init render)
- `internal/cli/update.go:650` (update render — site 1)
- `internal/cli/update.go:688` (update render — site 2, same update function)
- `internal/cli/update_clean_install.go:276` (clean-install render)

- Files: the 3 files above (4 injection sites — `update.go` has 2)
- Exit: every statusline render path supplies `ResolvedMoaiPath`; tests cover the empty-on-error path (REQ-SWF-005).

### M5 — Update TestFallbackChainOrder for new canonical chain order [Priority: High]

Update the existing `internal/template/hardcoded_path_audit_test.go` `TestFallbackChainOrder` (lines 27-31) and its `REQ-V3R2-RT-007-003` doc comment (line 28) to assert the NEW canonical fallback-chain order (REQ-SWF-011): `moai in PATH → resolved-executable (guarded) → $HOME/go/bin/moai (guarded) → $HOME/.local/bin/moai (guarded)`. The removed baked-`.GoBinPath` branch MUST NOT remain in the asserted order. This is an explicit supersession of RT-007-003's order contract (spec.md §A.4) — the test is UPDATED, never silently broken.

- Files: `internal/template/hardcoded_path_audit_test.go`
- Exit: `TestFallbackChainOrder` asserts the new order and PASSES; `grep -n "RT-007-003" ...` shows the updated comment.

### M6 — make build + neutrality + cross-compile [Priority: High]

Run `make build` to regenerate `internal/template/embedded.go`; run the neutrality audit; cross-compile for Windows. Confirm no internal-content leak (§25) and no language bias (§15).

- Files: `internal/template/embedded.go` (generated)
- Exit: §E batch all green; `embedded.go` reflects the new template.

### M7 — Full verification + coverage gate [Priority: High]

Run the full §E read-only batch. Confirm: full suite green (including the updated `TestFallbackChainOrder`), changed-package coverage ≥ baseline, Windows cross-compile green, lint clean.

- Exit: all acceptance.md ACs PASS; Definition of Done met.

## §G. Anti-Patterns to Avoid

- **Refactoring `gobin.Detect` to add `os.Stat`** — out of scope (spec.md §F); breaks `internal/shell/env.go` semantics.
- **Baking `%LOCALAPPDATA%` as an absolute path into the template** — would re-introduce a §14-style baked-path violation; rely on `os.Executable()` (Option A) instead.
- **Emitting an empty-valued `exec` line when `ResolvedMoaiPath` is empty** — REQ-SWF-005 requires conditional omission (`{{if .ResolvedMoaiPath}}`), not a blank `exec`.
- **Skipping the RED reproduction** — REQ-SWF-010 + maintainer request require the failing test first.
- **Forgetting one of the 4 render call sites** — `gobin.Detect` fan_in=2 (helper callers), but there are 4 `WithGoBinPath(...)` injection sites (init `initializer.go:284`, update `update.go:650` + `update.go:688`, clean-install `update_clean_install.go:276`); all 4 must additionally wire `WithResolvedMoaiPath(...)`. fan_in ≠ render-site count.
- **Silently breaking `TestFallbackChainOrder`** — removing the baked `.GoBinPath` branch supersedes RT-007-003's order contract (spec.md §A.4); the test MUST be UPDATED to the new order (M5), not left asserting the removed branch.

## §H. Cross-References

- spec.md §A.1 (three-layer root cause), §E (solution direction), §F (exclusions)
- acceptance.md (Given-When-Then + mechanical AC matrix)
- `internal/template/CLAUDE.md` §14 (the rule being restored)
- CLAUDE.local.md §2 / §6 / §14 / §15 / §25
