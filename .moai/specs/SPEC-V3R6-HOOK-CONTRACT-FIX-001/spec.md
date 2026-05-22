---
id: SPEC-V3R6-HOOK-CONTRACT-FIX-001
title: "WorktreeCreate/Remove Hook Contract Fix — Regression Guards + Working Tree Hygiene"
version: "0.1.0"
status: draft
created: 2026-05-22
updated: 2026-05-22
author: manager-spec
priority: Critical
phase: "v3.0.0"
module: "internal/hook, internal/cli, internal/template/templates/.claude/settings.json.tmpl, internal/template/templates/.claude/hooks/moai"
lifecycle: spec-anchored
tags: "hook, worktree, regression-guard, observability, wave-0, v3.0.0, critical"
tier: S
issue_number: null
depends_on: []
related_specs: [SPEC-V3R6-AGENT-FOLDER-SPLIT-001, SPEC-V3R6-HARNESS-RENAME-001]
---

# SPEC-V3R6-HOOK-CONTRACT-FIX-001: WorktreeCreate/Remove Hook Contract Fix

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-22 | manager-spec | Initial draft — Wave 0 SPEC #1. Lock in PR #1044 (commit `1b376eaef`) workaround as the canonical active-creator contract: cement regression guards (CI test against re-registration in settings.json + .tmpl), fix `internal/hook/.moai/` working-tree leak from subagent_stop.go observer fallback (REQ-HCF-005), close documentation drift (settings_test.go contract test + plain-text-stdout unit test in worktree_create_test.go / worktree_remove_test.go), and preserve handlers + shell wrappers + CLI subcommands as opt-in infrastructure. Tier S (~120 LOC, 6 files in-scope). |

## 1. Goal

Cement the **WorktreeCreate/WorktreeRemove active-creator contract** that PR #1044 (commit `1b376eaef`, MERGED 2026-05-22) introduced as a workaround. The PR removed the registration from `settings.json` and rewrote the documentation, but left the **handler code, shell wrappers, and CLI subcommands** in place as opt-in infrastructure with no regression guards. This SPEC adds those guards, fixes a related working-tree leak in `subagent_stop.go`, and locks in the contract so future contributors cannot silently re-introduce the `{}` regression.

**Active-creator contract** (the truth observed in `internal/cli/hook.go` `writeHookOutput()` lines 187-211):

> Claude Code v2.1.49+ parses WorktreeCreate / WorktreeRemove hook stdout as the **worktree directory path (plain text, not JSON)**. The dispatcher echoes `input.WorktreePath` unchanged; the Handle() methods return `&HookOutput{}` intentionally and the JSON HookOutput protocol is NOT used for these two events. Emitting a JSON object (e.g. `{}`) yields the runtime error "WorktreeCreate hook returned a path that is not a directory: `{}`".

## 2. Why

### 2.1 Status Quo Defect Surface

PR #1044 fix list (from memory `project_v3r6_worktree_hook_root_cause_fix_2026_05_22`):

1. `.claude/settings.json` + `internal/template/templates/.claude/settings.json.tmpl` — WorktreeCreate / WorktreeRemove keys removed (DONE)
2. `.claude/rules/moai/core/hooks-system.md` + `worktree-integration.md` — active-creator contract rewritten (DONE)
3. `docs-site/content/{ko,en,ja,zh}/advanced/hooks-reference.md` — 22→20 handler table corrected (DONE)
4. `internal/template/settings_test.go` — `expectedCount = 20` (DONE, line 512)

What PR #1044 did NOT do (residual scope, this SPEC):

| # | Residual | Risk if unaddressed |
|---|----------|---------------------|
| R1 | No CI guard against re-registering WorktreeCreate/WorktreeRemove in `settings.json` or its template | A future PR adding a "harmless logging hook" can silently re-trigger the `{}` regression |
| R2 | `worktree_create_test.go` + `worktree_remove_test.go` test ONLY the Handle() return value, NOT the writeHookOutput() plain-text contract | The stdout contract is implicit; refactor of `writeHookOutput()` cannot be caught |
| R3 | `internal/hook/.moai/harness/observations.yaml` working-tree leak (observed `git status`, 2026-05-22) — caused by `subagent_stop.go:204-210` `os.Getwd()` fallback when `input.CWD == ""` in unit tests | Test pollution; every `go test ./internal/hook/...` run when tests omit `input.CWD` creates a stray observations.yaml under the test's working directory |
| R4 | Handle() methods (`worktree_create.go`, `worktree_remove.go`) contain a stdout-contract explainer comment but no executable assertion that the contract STAYS this way | Documentation drift risk |
| R5 | Shell wrappers (`.claude/hooks/moai/handle-worktree-create.sh` + handle-worktree-remove.sh) + CLI subcommand `moai hook worktree-create` are preserved as opt-in infrastructure but un-tested for end-to-end stdout behavior | If a power user opts in by re-registering, no test catches the wrong-output failure mode |

### 2.2 v3.0 Design Report Wave 0 Position

`.moai/research/v3.0-design-2026-05-22.md` § Wave 0 (line 351-356):

> | Wave 0 SPEC | Tier | 핵심 |
> | --- | --- | --- |
> | `SPEC-V3R6-HOOK-CONTRACT-FIX-001` | S | WorktreeCreate `{}` 회귀 정정 — Critical |
>
> Wave 0 이 안 되면 Wave 1 진입 불가 (worktree hook 회귀가 manager-develop 호출 차단)

Wave 0 status (2026-05-22): PR #1044 already unblocked Wave 1 and Wave 2 (HARNESS-RENAME-001 + AGENT-FOLDER-SPLIT-001 are already in plan/run). This SPEC's role is **defensive consolidation**, not unblocking. Without it, the regression can re-emerge silently.

### 2.3 Cross-SPEC tension audit

- `SPEC-V3R6-HARNESS-RENAME-001` (MERGED 2026-05-22, PR #1043): independent — only renames agent files, does not touch `internal/hook/`.
- `SPEC-V3R6-AGENT-FOLDER-SPLIT-001` (plan-phase commit `58a235e06`): independent — only moves agent files between folders.
- `SPEC-V3R6-WORKTREE-HARD-LIST-ALIGN-001` (provisional, identified in memory `project_v3r6_agent_isolation_worktree_removal`): may overlap with R1 (settings.json HARD list alignment). Resolution: this SPEC owns the **regression guard** scope; HARD-LIST-ALIGN-001 owns the **HARD-rule list reconciliation** scope. If HARD-LIST-ALIGN-001 promotes, R1 stays here and HARD-LIST-ALIGN-001 must depend on this SPEC.

## 3. EARS Requirements

### REQ-HCF-001 (Ubiquitous, Active-creator contract preserved)

[ZONE:Frozen] [HARD] `internal/cli/hook.go` `writeHookOutput()` function MUST emit the WorktreeCreate / WorktreeRemove hook output as plain text consisting of `input.WorktreePath` followed by a single newline. When `input.WorktreePath == ""`, the function MUST emit nothing (empty stdout, no JSON, no newline-only). The JSON `HookOutput` protocol MUST NOT be used for `EventWorktreeCreate` or `EventWorktreeRemove`.

### REQ-HCF-002 (Ubiquitous, Handle() return discipline)

[ZONE:Frozen] [HARD] `worktreeCreateHandler.Handle()` and `worktreeRemoveHandler.Handle()` MUST return `(&HookOutput{}, nil)` on success (no Decision, no HookSpecificOutput, no stdout side effects from the Handle() body). All stdout emission for these two events is the dispatcher's responsibility per REQ-HCF-001.

### REQ-HCF-003 (Ubiquitous, settings.json non-registration)

[ZONE:Frozen] [HARD] Neither `.claude/settings.json` (local) nor `internal/template/templates/.claude/settings.json.tmpl` (template) MAY contain a top-level `hooks.WorktreeCreate` or `hooks.WorktreeRemove` entry. A CI guard test MUST fail the build when either key appears in either file.

### REQ-HCF-004 (Event-driven, Plain-text stdout regression test)

WHEN `go test ./internal/cli/...` runs, the test suite MUST include a test that:
- Invokes `writeHookOutput()` with `event = EventWorktreeCreate` and `input.WorktreePath = "/test/worktree/path"`,
- Captures stdout,
- Asserts the captured stdout equals `"/test/worktree/path\n"` exactly (plain text + single newline, NO JSON braces).

The symmetric assertion MUST exist for `event = EventWorktreeRemove`. Empty `input.WorktreePath` MUST be tested separately and MUST produce empty stdout.

### REQ-HCF-005 (Event-driven, observations.yaml path resolution)

WHEN `subagent_stop.go` `dispatchCapture()` resolves `obsPath`, the system MUST consult `$CLAUDE_PROJECT_DIR` environment variable BEFORE falling back to `os.Getwd()`. The resolution order is: `input.CWD` → `$CLAUDE_PROJECT_DIR` → `os.Getwd()`. This prevents the working-tree leak where unit tests inside `internal/hook/` cause `os.Getwd()` to return the package directory and seed `.moai/harness/observations.yaml` under the test's cwd.

### REQ-HCF-006 (Ubiquitous, observations.yaml leak cleanup)

[ZONE:Evolvable] [HARD] The current leaked file `internal/hook/.moai/harness/observations.yaml` (3025 bytes, observed in `git status` 2026-05-22) MUST be removed from the working tree as part of this SPEC's run-phase. The parent directories `internal/hook/.moai/harness/` and `internal/hook/.moai/` MUST also be removed. `.gitignore` MUST NOT be modified (the file was never tracked).

### REQ-HCF-007 (Ubiquitous, Shell wrapper + CLI preservation)

[ZONE:Frozen] [HARD] The following opt-in infrastructure MUST be preserved byte-identical:
- `.claude/hooks/moai/handle-worktree-create.sh`
- `.claude/hooks/moai/handle-worktree-remove.sh`
- Template mirrors at `internal/template/templates/.claude/hooks/moai/handle-worktree-create.sh` and `handle-worktree-remove.sh`
- CLI subcommands `moai hook worktree-create` and `moai hook worktree-remove` (registered at `internal/cli/hook.go:53` and the symmetric remove line)

These remain available for power users who explicitly opt into observer-style WorktreeCreate registration (e.g., custom CI orchestration). Removing them would break that opt-in path and is out of scope.

### REQ-HCF-008 (Ubiquitous, Documentation crosswalk consistency)

[ZONE:Evolvable] [HARD] The post-PR-#1044 state of the following documents MUST be verified consistent at run-phase entry:
- `.claude/rules/moai/core/hooks-system.md` (local + template mirror): active-creator contract present, 20-handler table
- `.claude/rules/moai/workflow/worktree-integration.md` (local + template mirror): §WorktreeCreate and WorktreeRemove Hooks section present
- `docs-site/content/{ko,en,ja,zh}/advanced/hooks-reference.md`: 20-handler table consistent across all 4 locales
- `internal/template/settings_test.go`: `expectedCount = 20` (line 512)

If any of these has drifted since PR #1044 merge, the drift MUST be corrected within this SPEC's run-phase. If no drift, AC-HCF-008 is a pass-by-inspection.

### REQ-HCF-009 (Optional, Opt-in re-registration guidance)

WHERE a future user opts into observer-style WorktreeCreate or WorktreeRemove registration by editing their `.claude/settings.json` directly, the project documentation (specifically `hooks-system.md` §WorktreeCreate and WorktreeRemove Hooks) MUST state:
- The user is opting into active-creator contract (their hook becomes the worktree creator)
- The user's hook MUST emit `input.WorktreePath` as plain text on stdout (not JSON)
- The user's hook MUST handle the empty-`input.WorktreePath` case (return empty stdout, not an error)

This is a documentation requirement; no code enforcement.

## 4. Constraints

### 4.1 Technical Constraints

- **Cross-platform**: `go build ./...` AND `GOOS=windows GOARCH=amd64 go build ./...` MUST both exit 0. No syscall-package usage expected.
- **Test discipline**: All new tests in `internal/cli/hook_test.go` (or new file) and `internal/hook/subagent_stop_test.go` MUST use `t.TempDir()` for any directory creation; MUST NOT touch real `$HOME`; MUST NOT depend on a real `$CLAUDE_PROJECT_DIR` (set via `t.Setenv`).
- **No mocks of code under test**: REQ-HCF-004 plain-text-stdout test asserts on real stdout via `os.Pipe()` + goroutine reader OR via dependency injection of `io.Writer` in `writeHookOutput()`. Mocking `writeHookOutput()` itself defeats the purpose. (Implementation choice deferred to run-phase.)
- **Idempotency**: REQ-HCF-006 working-tree leak cleanup MUST be idempotent — re-running the cleanup MUST NOT fail if the file is already absent.
- **Conventional Commits + `🗿 MoAI <email@mo.ai.kr>` trailer mandatory**

### 4.2 Cross-SPEC Constraints

- **No dependency on Wave 1/Wave 2 SPECs** — this SPEC operates on `internal/hook/` + `internal/cli/` + settings template only, none of which are touched by HARNESS-RENAME-001 or AGENT-FOLDER-SPLIT-001.
- **No conflict with `SPEC-V3R6-WORKTREE-HARD-LIST-ALIGN-001`** (provisional): this SPEC owns regression guards (CI test); HARD-LIST-ALIGN-001 owns rule-file HARD-list reconciliation. The two scopes are orthogonal.

### 4.3 PRESERVE Constraints (working tree hygiene per Known Issues B8)

Run-phase MUST NOT modify the following untracked or modified files outside this SPEC's scope:
- 2 modified docs-site files (`docs-site/hugo.toml`, `docs-site/layouts/_default/baseof.html`, `docs-site/layouts/partials/menu.html`) — modified per parallel work
- 4 untracked SPEC dirs (HARNESS-USER-AREA-RESOLUTION not yet created; will appear later)
- `.moai/research/moai-adk-current-state-2026-05-22.md` and `v3.0-design-2026-05-22.md` (input reports, PRESERVE)
- `docs-site/content/{ko,en,ja,zh}/book/` (untracked Hugo content)
- `docs-site/scripts/gen_menu.py`, `docs-site/static/book/`, `docs-site/data/menu/extra.yaml`, `docs-site/layouts/_default/redirect.html` (untracked Hugo extras)
- `.moai/harness/usage-log.jsonl` (runtime-managed)

Only `internal/hook/.moai/harness/observations.yaml` + its parent dirs are in scope for deletion (REQ-HCF-006).

### 4.4 Forbidden Operations

- `git reset --hard` (use `--keep` per CLAUDE.local.md §23.5)
- `--no-verify` on commit
- `--amend` (use new commits)
- force-push to main
- Modifying `internal/hook/worktree_create.go` or `internal/hook/worktree_remove.go` Handle() body (only the stdout-contract explainer comments may be polished; behavior MUST stay)
- Modifying `.claude/hooks/moai/handle-worktree-create.sh` or `handle-worktree-remove.sh` (REQ-HCF-007 PRESERVE)
- Removing `EventWorktreeCreate` or `EventWorktreeRemove` from `hook.EventType` enum (handler registration MUST stay; only settings.json activation MUST stay absent)
- AskUserQuestion call from any `internal/hook/` or `internal/cli/` code (C-HRA-008 boundary; subagent domain)

## 5. Scope and Out of Scope

### 5.1 In Scope

#### 5.1.1 New tests (regression guards)

1. `internal/cli/hook_test.go` (or new file `hook_writehookoutput_test.go`): `TestWriteHookOutput_WorktreeCreatePlainText`, `TestWriteHookOutput_WorktreeRemovePlainText`, `TestWriteHookOutput_EmptyWorktreePathProducesEmptyStdout` (REQ-HCF-001, REQ-HCF-004)
2. `internal/template/settings_test.go` (or new file `settings_no_worktree_keys_test.go`): `TestSettingsJsonHasNoWorktreeCreateKey`, `TestSettingsTmplHasNoWorktreeCreateKey` — fail when `WorktreeCreate` or `WorktreeRemove` appears as a top-level hooks key (REQ-HCF-003)

#### 5.1.2 Source modification (single targeted fix)

1. `internal/hook/subagent_stop.go` `dispatchCapture()` (lines 204-210): insert `$CLAUDE_PROJECT_DIR` lookup between `input.CWD` and `os.Getwd()` (REQ-HCF-005). Approximate diff: ~5 LOC added.

#### 5.1.3 Working-tree cleanup

1. `git rm -r internal/hook/.moai/` (the directory is untracked, so this is actually `rm -rf internal/hook/.moai/` since git rm requires tracked files). The acceptance test verifies absence (REQ-HCF-006, AC-HCF-006).

#### 5.1.4 Documentation crosswalk verification (no edits unless drift found)

1. Inspect `.claude/rules/moai/core/hooks-system.md` (local + template) — verify §WorktreeCreate and WorktreeRemove Hooks section present and 20-handler count consistent
2. Inspect `.claude/rules/moai/workflow/worktree-integration.md` (local + template) — same
3. Inspect 4 locales of `docs-site/content/*/advanced/hooks-reference.md` — verify 20-row handler table
4. Inspect `internal/template/settings_test.go:512` — verify `expectedCount = 20`

If no drift found, AC-HCF-008 is a pass-by-inspection (zero edits). If drift found, fix-up Edit operations are run-phase work.

### 5.2 Out of Scope

#### 5.2.1 Items explicitly out of scope

- **Subagent-boundary regression test** (`internal/hook/subagent_boundary_test.go`): broader C-HRA-008 enforcement is owned by a different SPEC. This SPEC only checks the boundary via grep at AC-HCF-010 (E4 deliverable).
- **Re-registration of WorktreeCreate / WorktreeRemove** for any purpose (logging, telemetry, observer pattern): explicitly forbidden by REQ-HCF-003. Power users may opt in by editing their personal `.claude/settings.local.json`; the project template stays clean.
- **Refactoring `writeHookOutput()` into a strategy pattern or per-event handler dispatch table**: out of scope. Simple in-line dispatch is the canonical form. Refactor request goes to a follow-up SPEC.
- **Removing handlers `worktreeCreateHandler` / `worktreeRemoveHandler`**: out of scope. They are dormant opt-in infrastructure (REQ-HCF-007 PRESERVE).
- **Removing shell wrappers `handle-worktree-create.sh` / `handle-worktree-remove.sh`**: out of scope (REQ-HCF-007 PRESERVE).
- **Removing CLI subcommands `moai hook worktree-create` / `moai hook worktree-remove`**: out of scope (REQ-HCF-007 PRESERVE).
- **CLAUDE_PROJECT_DIR fallback in other hook handlers** (compact.go, file_changed.go, instructions_loaded.go): out of scope. Only `subagent_stop.go` is in scope per REQ-HCF-005, because it is the documented leak source. A broader audit of `os.Getwd()` usage is a separate concern.
- **Updating `.moai/research/moai-adk-current-state-2026-05-22.md` § 6 Hooks** to reflect 20-handler count: out of scope. Research files are historical artifacts (CLAUDE.local.md §22-style hygiene).
- **Updating CLAUDE.md / CLAUDE.local.md / book/ Hugo content**: out of scope. Hugo updates flow through docs-only SPECs.

#### 5.2.2 Deferred Risks

- **Future PR adding a 21st hook (legitimate new event)**: the new CI guard test in §5.1.1 only checks WorktreeCreate / WorktreeRemove absence; it does NOT enforce the 20-handler count. Updating `settings_test.go:512 expectedCount = 20` is the existing mechanism. A new hook should bump that constant in the same PR.
- **`subagent_stop_test.go` does not cover the `$CLAUDE_PROJECT_DIR` fallback**: out of scope for this SPEC. Test will use `t.Setenv` to set the var explicitly; we are not asserting all branches of the resolution order, only that the var IS consulted before `os.Getwd()`. Branch coverage to 100% for `dispatchCapture()` is a separate concern.

## 6. Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| R-HCF-001: A future PR reintroduces `WorktreeCreate` to settings.json believing it's safe ("just observer logging") | M | H | REQ-HCF-003 CI guard test fails the build immediately. AC-HCF-003 enforces. |
| R-HCF-002: `writeHookOutput()` refactor accidentally changes plain-text to JSON for one event | L | H | REQ-HCF-004 plain-text-stdout regression test catches the change at PR CI time. AC-HCF-004 enforces. |
| R-HCF-003: `subagent_stop.go` `$CLAUDE_PROJECT_DIR` lookup interacts with existing tests that don't set the var | L | M | New test in `subagent_stop_test.go` uses `t.Setenv("CLAUDE_PROJECT_DIR", "")` to verify fallback to `os.Getwd()` still works. Existing tests inspected at AC-HCF-005 review. |
| R-HCF-004: `internal/hook/.moai/harness/observations.yaml` is currently un-tracked, so `git rm` fails. `rm -rf` works but cannot be undone from git history if accidentally targeting wrong path | L | M | Run-phase verification: `git ls-files internal/hook/.moai/` returns empty (untracked). Use `rm -rf internal/hook/.moai/` only after verifying. AC-HCF-006 final check `test ! -e internal/hook/.moai/`. |
| R-HCF-005: New CI guard test (REQ-HCF-003) false-positives on a legitimate string occurrence of "WorktreeCreate" inside a comment in settings.json (e.g., a generated banner) | L | L | Test parses the file as JSON (or regex with key-position anchor) rather than substring match. Implementation choice in run-phase. |
| R-HCF-006: Adding the `CLAUDE_PROJECT_DIR` lookup to `subagent_stop.go` changes the behavior in environments where the env var is set but points to a path that doesn't exist or is wrong | M | M | The lookup is **after** `input.CWD` (which is set by Claude Code in production) and **before** `os.Getwd()` (last-resort fallback). The change only affects the "input.CWD empty AND env var set" branch, which is exactly what we want to fix (tests). Production paths via `input.CWD` are unchanged. |

## 7. Dependencies

### 7.1 Explicit Dependencies

- **PR #1044** (commit `1b376eaef`, MERGED 2026-05-22): The active-creator contract documentation + settings.json deregistration. This SPEC builds defensively on top of that PR. If PR #1044 is reverted, this SPEC's run-phase MUST re-verify and possibly re-apply the deregistration as part of REQ-HCF-003 fix-up.

### 7.2 Cross-cutting

- **CLAUDE.local.md §2 Template-First Rule**: applies to `internal/template/templates/.claude/settings.json.tmpl` and `internal/template/templates/.claude/hooks/moai/*` mirrors
- **1-person OSS Hybrid Trunk policy** (CLAUDE.local.md §23): `auto_branch: true`, `auto_pr: true`, `branch_prefix: feat/SPEC-` for Tier S; main direct push allowed for trivial fix-ups

### 7.3 No follow-up SPECs required by this work

This SPEC is self-contained. The optional follow-up identified is `SPEC-V3R6-HOOK-OBS-PATH-AUDIT-001` (provisional) — a broader audit of `os.Getwd()` fallback usage across `internal/hook/`, deferred and not blocking this SPEC.

## 8. Success Metrics

### 8.1 Quantitative

- 1 new test function in `internal/cli/hook_test.go` (or new file) with ≥3 subtests (Create plain-text, Remove plain-text, empty-path)
- 1 new test function in `internal/template/settings_test.go` (or new file) asserting WorktreeCreate / WorktreeRemove absence in both local settings.json and template
- 1 modified function in `internal/hook/subagent_stop.go` (`dispatchCapture`) with ~5 LOC added (env-var lookup)
- 1 working-tree cleanup: `internal/hook/.moai/` removed (currently 3025 bytes + 1 dir + 1 subdir + 1 file)
- `go build ./...` exit 0, `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- `go test ./internal/cli/... ./internal/hook/... ./internal/template/...`: NEW regression 0 (baseline residual preserved per CLAUDE.local.md §23 hygiene)
- spec-lint NEW regression 0
- Total in-scope edit budget: ~120 LOC across 3-5 files (2 new test files or extensions + 1 source modification + working-tree cleanup)

### 8.2 Qualitative

- Future contributors physically cannot re-introduce `WorktreeCreate` to `settings.json` (or its `.tmpl`) without an immediate CI failure
- `writeHookOutput()` plain-text contract has an executable spec test that fails fast on refactor
- `internal/hook/.moai/` leak is eliminated and the root cause (subagent_stop.go fallback) is fixed at source
- Documentation crosswalk verified consistent post-PR-#1044

## 9. Edge Cases

### 9.1 Empty input.WorktreePath

When `input.WorktreePath == ""` for an `EventWorktreeCreate` event, `writeHookOutput()` MUST emit empty stdout (no newline, no JSON). REQ-HCF-001 wording is explicit; AC-HCF-001 has the exact byte-level assertion.

### 9.2 Missing $CLAUDE_PROJECT_DIR

When `input.CWD == ""` AND `$CLAUDE_PROJECT_DIR` is unset (env var absent), `subagent_stop.go` `dispatchCapture()` falls through to `os.Getwd()`. This preserves current behavior for environments where the env var is not exported. REQ-HCF-005 resolution order: `input.CWD` → `$CLAUDE_PROJECT_DIR` → `os.Getwd()`.

### 9.3 Concurrent hook invocation

WorktreeCreate / WorktreeRemove hooks are invoked sequentially by Claude Code runtime per agent spawn (no concurrent same-event invocation expected). Handler is stateless (no shared mutable state); concurrent invocation would be safe even if it happened. No mitigation needed.

### 9.4 settings.local.json bypass

`.claude/settings.local.json` is runtime-managed per CLAUDE.local.md §2 and is NOT covered by REQ-HCF-003. A power user can put `WorktreeCreate` registration in `settings.local.json` and the build will pass. This is intentional — `settings.local.json` is the explicit opt-in surface per REQ-HCF-007.

## 10. Pre-flight Baseline (2026-05-22)

### 10.1 File counts and key facts

- Branch: `main`, HEAD `58a235e06`
- `internal/hook/.moai/harness/observations.yaml`: 3025 bytes (created 2026-05-22 18:58, last mtime 2026-05-22 19:01)
- `internal/hook/.moai/` directory contains exactly 1 file: `harness/observations.yaml`
- `internal/hook/worktree_create.go`: 48 lines, Handle() at line 33, return `&HookOutput{}` at line 47
- `internal/hook/worktree_remove.go`: 49 lines, Handle() at line 34, return `&HookOutput{}` at line 48
- `internal/cli/hook.go`: `writeHookOutput()` at lines 187-211, plain-text branch at lines 198-205
- `internal/template/settings_test.go:512`: `const expectedCount = 20` (PR #1044 state)
- `.claude/settings.json` + `internal/template/templates/.claude/settings.json.tmpl`: 0 occurrences of `WorktreeCreate` or `WorktreeRemove` as hook keys (PR #1044 state, verified via grep)

### 10.2 Existing test coverage gaps (motivating this SPEC)

- `internal/hook/worktree_create_test.go` lines 8-67: tests Handle() return value only (no decision, no HookSpecificOutput). Does NOT test writeHookOutput().
- `internal/hook/worktree_remove_test.go` lines 8-67: same gap for Remove.
- `internal/cli/hook_test.go`: does NOT contain a writeHookOutput() unit test. Function tested only transitively via integration.
- `internal/template/settings_test.go`: counts total hooks (line 512) but does NOT assert specific event keys absent.

### 10.3 Cross-platform considerations

- No syscall package usage expected (REQ-HCF-005 fix uses only `os.Getenv`, which is fully cross-platform)
- Path separators: REQ-HCF-005 uses `filepath.Join` (already in subagent_stop.go:210) — no change to path-handling style
- No build tags required

### 10.4 Baseline test pass state (run-phase entry condition)

`go test ./internal/cli/... ./internal/hook/... ./internal/template/...` baseline at HEAD `58a235e06`: expected to pass cleanly (no Wave 2 SPEC has touched these packages yet). Baseline residual 3 FAIL from `TestRuleTemplateMirrorDrift` etc. (per `project_v3r6_template_mirror_drift_audit_2026_05_22`) is in `internal/template/` but unrelated to this SPEC's targets.

## 11. References

- v3.0 design report: `.moai/research/v3.0-design-2026-05-22.md` § Wave 0 (line 351-356)
- Baseline: `.moai/research/moai-adk-current-state-2026-05-22.md` § 6 Hooks (line 269), § 8 F6 (line 392)
- PR #1044 (commit `1b376eaef`): Active-creator contract + settings.json deregistration (auto-memory `project_v3r6_worktree_hook_root_cause_fix_2026_05_22`)
- Agent isolation removal (commit `25ee73039`): 5 agents removed `isolation: worktree` (auto-memory `project_v3r6_agent_isolation_worktree_removal`)
- Claude Code official hooks docs: `code.claude.com/docs/en/hooks` (active-creator semantic verified)
- Frontmatter schema: `.claude/rules/moai/development/spec-frontmatter-schema.md`
- Hybrid Trunk policy: CLAUDE.local.md §23
- Known Issues B3 (C-HRA-008 boundary): `.claude/rules/moai/development/manager-develop-prompt-template.md` § B3
- Known Issues B7 (observer.go / capture path resolution): same file § B7
- Known Issues B8 (working tree hygiene): same file § B8
