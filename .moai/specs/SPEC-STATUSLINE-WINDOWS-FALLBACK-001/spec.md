---
id: SPEC-STATUSLINE-WINDOWS-FALLBACK-001
title: "Windows statusline phantom-path fallback repair"
version: "0.2.0"
status: completed
created: 2026-06-20
updated: 2026-06-21
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/template, internal/runtime/gobin"
lifecycle: spec-anchored
tags: "statusline, windows, template, gobin, bugfix"
issue_number: 1068
era: V3R6
tier: M
---

## HISTORY

- 2026-06-20: Initial draft. Origin: GitHub issue modu-ai/moai-adk#1068 — `moai init` / `moai update` generate `.moai/status_line.sh` whose Go-bin fallback bakes a non-existent `C:\Users\{user}\go\bin\moai` path on Windows installer machines, causing the Claude Code statusline to silently render empty.

## §A. Context and Problem Statement

The MoAI statusline wrapper (`.moai/status_line.sh`, generated from `internal/template/templates/.moai/status_line.sh.tmpl`) is responsible for forwarding Claude Code's stdin JSON to the `moai statusline` command. It uses a three-stage fallback chain to locate the `moai` binary:

1. `command -v moai` (PATH lookup)
2. A baked Go-bin path: `{{posixPath .GoBinPath}}/moai`
3. `$HOME/.local/bin/moai`

On a **Windows installer** machine (binary installed via `install.ps1` to `%LOCALAPPDATA%\Programs\moai\moai.exe`), all three stages fail:

- **Stage 1** fails because `install.ps1`'s `HKCU\Environment\Path` update does not propagate to an already-running Claude Code / terminal child process (the env change only affects newly-launched processes).
- **Stage 2** fails because the baked Go-bin path is a **phantom**: `gobin.Detect` (`internal/runtime/gobin/resolver.go`) returns `go env GOPATH/bin` (or `$HOME/go/bin`) **without any existence validation** (`os.Stat`). On a Windows installer machine Go is typically absent, yet `go env GOPATH` still returns its default value — so `gobin.Detect` returns `C:\Users\{user}\go\bin`, a directory that does not exist. The runtime `[ -f ]` guard then correctly finds no file there and falls through.
- **Stage 3** fails because `$HOME/.local/bin/moai` is the Unix installer location; the Windows installer uses `%LOCALAPPDATA%\Programs\moai\moai.exe`, which has no fallback branch.

The net result is `exit 0` (silent), and the statusline renders empty. Claude Code handles a missing statusline gracefully, so the user sees no error — only an absent statusline.

### §A.1 Verified Root Cause (three layers)

All three layers were confirmed by reading source, not inferred:

- **Layer A — detection has no existence validation.** `internal/runtime/gobin/resolver.go` `Detect()` fallback chain (`go env GOBIN` → `go env GOPATH/bin` → `$HOME/go/bin` → `""`) contains no `os.Stat` call. `go env GOPATH` returns the default value even when the directory does not exist; this is the actual mechanism (NOT a Git-Bash `$PATH` scan).
- **Layer B — the template bakes the unvalidated absolute path.** `internal/template/templates/.moai/status_line.sh.tmpl` lines 27-28 use `{{posixPath .GoBinPath}}/moai`. `posixPath` (`internal/template/renderer.go:27`) only converts `\` → `/`; it performs no validation. The value is injected at init/update time via `WithGoBinPath(...)` with no `os.Stat`. The runtime `[ -f ]` guard cannot rescue a phantom baked path.
- **Layer C — Windows-specific compounding.** `install.ps1:271` installs `moai.exe` to `%LOCALAPPDATA%\Programs\moai\`, while the Unix installer (`install.ps1:274`) uses `$HOME/.local/bin`. The fallback chain has a Unix `$HOME/.local/bin` branch but no Windows installer-location branch, and Stage 1 `command -v moai` fails due to non-propagated `HKCU\Environment\Path`.

### §A.2 Self-Inflicted Rule Violation

`internal/template/CLAUDE.md` §14 (HARD) states: generated shell-script fallbacks MUST use `$HOME` for portability and MUST NEVER use a baked `.GoBinPath` / `.HomeDir` absolute path. The current `status_line.sh.tmpl` Stage 2 violates this by baking `{{posixPath .GoBinPath}}`. This fix MUST bring the template back into §14 compliance.

### §A.3 Regression Surface

`gobin.Detect` carries `@MX:ANCHOR fan_in=2` — the helper itself has exactly 2 non-test callers (`internal/core/project/initializer.go:301` via `detectGoBinPath`, and `internal/cli/update.go:3054` via `detectGoBinPathForUpdate`). This fan_in count is DISTINCT from the count of `WithGoBinPath(...)` render call sites: there are **4** render sites that inject `GoBinPath` into a `TemplateContext` (init `initializer.go:284`; update `update.go:650` and `update.go:688` — two invocations in the same update function; clean-install `update_clean_install.go:276`). The fix MUST wire the new resolved-executable field at all **4** render sites (not 2), and `make build` MUST regenerate `internal/template/embedded.go`.

### §A.4 Existing Test Supersession — REQ-V3R2-RT-007-003

The pre-existing test `internal/template/hardcoded_path_audit_test.go:27-31` `TestFallbackChainOrder` was an **inert empty-body stub**: its body contained only a `// GREEN: ...` comment and zero assertions, so it trivially passed regardless of the template's actual output. Its doc comment (`REQ-V3R2-RT-007-003`, line 28) described the intended canonical order as `moai in PATH → $HOME/go/bin/moai → {{posixPath .GoBinPath}}/moai`, but the test never enforced it. This SPEC's fix REMOVES the baked `{{posixPath .GoBinPath}}/moai` branch (REQ-SWF-004), which **supersedes the REQ-V3R2-RT-007-003 fallback-chain-order contract**. The run-phase (commit `c4a42bc11`) converted the inert stub into a REAL assertion of the NEW canonical chain order — it now renders the template with a known context (`WithResolvedMoaiPath("/opt/moai/bin/moai")`) and asserts (a) a positive render-and-assert of the four-stage ascending index order (`command -v moai` → resolved-executable → `$HOME/go/bin/moai` → `$HOME/.local/bin/moai`) AND (b) a negative `must-be-gone` assertion that the rendered output no longer contains a `GoBinPath` token. The test MUST NOT be left as a doc-comment-only edit; see REQ-SWF-011 + acceptance.md AC-SWF-013 for the strengthened mechanical gate. The two sibling tests in the same file (`TestNoHardcodedAbsolutePath_StatusLine` line 17, `TestNoHardcodedAbsolutePath_HookWrappers` line 10) remain inert stubs that benefit from the fix in spirit but are not modified (removing the baked path strengthens, not weakens, the no-hardcoded-path invariant they assert in comment).

## §B. Stakeholders and Goals

| Stakeholder | Goal |
|-------------|------|
| Windows installer users | Statusline renders correctly after `moai init` / `moai update` |
| Unix (macOS/Linux) users | No regression — existing PATH / `$HOME/go/bin` / `$HOME/.local/bin` resolution preserved |
| Maintainer | Template returns to `internal/template/CLAUDE.md` §14 compliance |
| Issue #1068 reporter | A failing reproduction test precedes the fix (explicit maintainer request) |

## §C. Requirements (GEARS)

### Functional Requirements

- **REQ-SWF-001** (Ubiquitous): The statusline wrapper template **shall** resolve the `moai` binary via a fallback chain that contains no unvalidated baked absolute path — every non-`PATH` fallback branch shall either use a `$HOME`-relative path or guard the candidate with an existence check (`[ -f ]`) before `exec`.

- **REQ-SWF-002** (Event-driven): **When** the project is initialized or updated (`moai init` / `moai update`), the renderer **shall** expose the running binary's own resolved executable path (captured via `os.Executable()`) to the statusline template as a dedicated `TemplateContext` field.

- **REQ-SWF-003** (State-driven): **While** `command -v moai` does not resolve a `moai` binary on `PATH`, the statusline wrapper **shall** attempt the resolved-executable path (REQ-SWF-002) as the first fallback, before any `$HOME`-relative fallback branch.

- **REQ-SWF-004** (Ubiquitous): The statusline wrapper template **shall not** emit a `{{posixPath .GoBinPath}}/moai` branch or any other branch that bakes the `.GoBinPath` / `.HomeDir` template variable as an absolute path (returning the template to `internal/template/CLAUDE.md` §14 compliance).

- **REQ-SWF-005** (Capability gate): **Where** the resolved-executable `TemplateContext` field is empty (e.g. `os.Executable()` returned an error at init/update time), the renderer **shall** omit the resolved-executable branch from the generated script rather than emit an empty-valued `exec` line, and the wrapper **shall** fall through to the `$HOME`-relative branches.

- **REQ-SWF-006** (Event-driven): **When** the generated `status_line.sh` is rendered, the rendering **shall** produce a script in which every `exec ... moai statusline` line is reached only after an existence guard or a PATH-resolved match — so a non-existent candidate path never causes a phantom `exec` attempt.

### Non-Functional Requirements

- **REQ-SWF-007** (Ubiquitous): The fix **shall** preserve cross-platform correctness — `GOOS=windows GOARCH=amd64 go build ./...` shall pass, and path handling shall be correct on darwin / linux / windows (forward-slash POSIX form in the generated shell script via `posixPath`, since the wrapper runs under git-bash / WSL on Windows).

- **REQ-SWF-008** (Ubiquitous): The generated template output **shall** remain 16-language neutral and free of moai-adk internal-content leakage (no internal SPEC IDs, dates, commit SHAs, or OS-biased absolute paths baked into `internal/template/templates/**`), per CLAUDE.local.md §15 + §25.

- **REQ-SWF-009** (Ubiquitous): The changed code **shall** be covered by tests such that the diff-level coverage of the modified statements in `internal/template` and `internal/runtime/gobin` is not lower than the pre-change package baseline, and the new reproduction + fix logic shall carry direct test assertions (see acceptance.md for the mechanical coverage gate).

### Reproduction-First Requirement

- **REQ-SWF-010** (Event-driven): **When** the fix is implemented, a failing reproduction test **shall** have been written and observed to fail FIRST (asserting the rendered `status_line.sh` contains an unvalidated phantom Go-bin path / lacks a guarded resolved-executable branch), then made to pass by the fix — per CLAUDE.md §7 Rule 4 and the issue maintainer's explicit request.

- **REQ-SWF-011** (Ubiquitous): The fix **shall** UPDATE the existing `internal/template/hardcoded_path_audit_test.go` `TestFallbackChainOrder` test and its `REQ-V3R2-RT-007-003` doc comment to assert the NEW canonical fallback-chain order — `moai in PATH → resolved-executable (guarded, when present) → $HOME/go/bin/moai (guarded) → $HOME/.local/bin/moai (guarded)` — superseding the prior `PATH → $HOME/go/bin → {{posixPath .GoBinPath}}` order. The test **shall not** be left asserting the removed baked-`.GoBinPath` branch (which would silently fail at run-phase).

## §D. Constraints

1. **Template-First Rule** (CLAUDE.local.md §2 [HARD]): all template changes are made in `internal/template/templates/.moai/status_line.sh.tmpl` FIRST, then `make build` regenerates `internal/template/embedded.go`. Mirror parity (`embedded_mirror_test.go` where applicable) must hold.
2. **§14 HARD rule** (internal/template/CLAUDE.md): generated shell-script fallbacks use `$HOME`, never baked `.GoBinPath`.
3. **Cross-platform**: `GOOS=windows GOARCH=amd64 go build ./...` must pass; the generated script uses POSIX forward-slash form.
4. **16-language neutrality + internal-content isolation** (CLAUDE.local.md §15 / §25): no internal artifacts leak into `internal/template/templates/**`.
5. **Coverage** (CLAUDE.local.md §6): critical packages (template, cli) target ≥90%; at minimum the changed statements must not lower the package baseline (template baseline measured at 84.6%, gobin at 70.0% — see plan.md §B). The SPEC does NOT claim the packages already meet 90%; it requires the change to add direct coverage of the new logic and not regress the baseline.
6. **Scope discipline**: fix ONLY the statusline phantom-path defect. Do NOT refactor `gobin.Detect`'s unrelated callers. Note: `internal/shell/env.go` does NOT consume `gobin.Detect` output — its `AddGoBinPath` is a boolean option that, when true, appends the hardcoded literal `"$HOME/go/bin"` via `AddPathEntry(config.ConfigFile, "$HOME/go/bin")`. It references neither `gobin.Detect`, the `GoBinPath` template variable, nor any detected path; it is therefore unaffected by this fix regardless. `gobin.Detect` is out of scope because the fix operates at the template-render boundary (Option A `os.Executable()` + Option B removing the baked branch) and does not require touching `Detect`.

## §E. Solution Direction (WHAT, not HOW)

Combine issue option A (primary) and option B (defense-in-depth):

- **Option A (primary)** — capture the running binary's own path via `os.Executable()` at init/update time, expose it as a NEW `TemplateContext` field (working name `ResolvedMoaiPath`), and make it the FIRST fallback in the generated script after `command -v moai`. Because the binary running `moai init` / `moai update` IS the installed binary, this resolves to `%LOCALAPPDATA%\Programs\moai\moai.exe` on Windows installer machines, `$GOPATH/bin/moai` on `go install`, and the correct path on manual copy.
- **Option B (defense-in-depth)** — in the generated script, keep only `$HOME`-relative portable branches (`$HOME/go/bin/moai`, `$HOME/.local/bin/moai`) and guard each with `[ -f ]`. Remove the baked `{{posixPath .GoBinPath}}` branch (REQ-SWF-004).

The exact field name, the precise branch ordering, and whether an OS-aware Windows installer-location branch (issue option C) is added are HOW decisions deferred to plan.md / run-phase.

## §F. Exclusions

This section enumerates what this SPEC explicitly does NOT build. Items below are out of scope and, where they belong elsewhere, are routed to their correct home.

### Out of Scope — gobin.Detect refactor

- Adding `os.Stat` existence validation INSIDE `gobin.Detect` itself. The defect is fixed at the template-render boundary (Option A/B), NOT by changing the shared `gobin.Detect` contract. The right reason to exclude is that the fix (Option A `os.Executable()` + Option B removing the baked branch) does not require touching `Detect` at all. (Earlier drafts named `internal/shell/env.go` `AddGoBinPath` as a second `gobin.Detect` consumer whose semantics would change; that is incorrect — `env.go` does NOT consume `gobin.Detect` output. Its `AddGoBinPath` is a boolean toggle that appends the hardcoded literal `"$HOME/go/bin"` via `AddPathEntry(config.ConfigFile, "$HOME/go/bin")`, referencing neither `Detect`, `GoBinPath`, nor any detected path.) The conclusion (do not refactor `Detect`) is unchanged; the rationale is corrected.
- Changing the `gobin.Detect` fallback-chain order or removing any of its four stages.

### Out of Scope — installer (install.ps1) changes

- Fixing the `HKCU\Environment\Path` propagation in `install.ps1` so that `command -v moai` succeeds in already-running processes. The statusline fix makes the wrapper robust regardless of PATH propagation; repairing the installer's PATH-propagation behavior is a separate concern and is not built here.
- Adding a Windows installer self-test or post-install statusline verification.

### Out of Scope — PATH-construction call sites

- Modifying `internal/shell/env.go` (`AddGoBinPath`). This file does NOT consume `gobin.Detect` output; its `AddGoBinPath` is a boolean option that, when true, appends the hardcoded literal `"$HOME/go/bin"` via `AddPathEntry(config.ConfigFile, "$HOME/go/bin")`. It references neither the `GoBinPath` template variable nor any detected path, so it is unrelated to the statusline fallback baking concern and is explicitly preserved unchanged. (An earlier draft mischaracterized it as a `GoBinPath` consumer; that is corrected here.)

### Out of Scope — broader statusline behavior

- Statusline rendering content, padding, GLM env loading, or any behavior of `moai statusline` itself. Only the binary-resolution fallback chain of the wrapper script is in scope.
- Windows-native (PowerShell / `.cmd`) statusline wrapper. The wrapper remains a bash script (runs under git-bash / WSL on Windows); producing a non-bash variant is excluded.

## §G. Acceptance Criteria Summary

Full Given-When-Then scenarios and the mechanical AC matrix live in `acceptance.md`. The Definition of Done requires: failing reproduction test written and observed first (REQ-SWF-010); template §14 compliant (no baked `.GoBinPath`); resolved-executable branch present and guarded; cross-platform build green; coverage gate met; neutrality clean; `make build` regenerated `embedded.go`.

## §H. Cross-References

- GitHub issue: modu-ai/moai-adk#1068
- `internal/runtime/gobin/resolver.go` — Layer A (no existence validation)
- `internal/template/templates/.moai/status_line.sh.tmpl` — Layer B (baked path, lines 27-28)
- `internal/template/renderer.go:27` — `posixPath` helper
- `internal/template/context.go:10` — `TemplateContext` struct (new field target)
- `gobin.Detect` callers (fan_in=2): `internal/core/project/initializer.go:301` + `internal/cli/update.go:3054`
- `WithGoBinPath(...)` render sites (4, the wiring targets): `internal/core/project/initializer.go:284` + `internal/cli/update.go:650` + `internal/cli/update.go:688` + `internal/cli/update_clean_install.go:276`
- `internal/template/hardcoded_path_audit_test.go:27-31` — `TestFallbackChainOrder` (REQ-V3R2-RT-007-003 supersession target, §A.4 + REQ-SWF-011)
- `install.ps1:271` — Windows installer location (`%LOCALAPPDATA%\Programs\moai`)
- `internal/template/CLAUDE.md` §14 — `$HOME`-fallback HARD rule
- CLAUDE.local.md §2 (Template-First) / §6 (coverage) / §14 (hardcoding) / §15 (neutrality) / §25 (internal-content isolation)
