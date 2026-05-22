# SPEC-V3R6-HOOK-CONTRACT-FIX-001 — Acceptance Criteria

## Overview

This document enumerates binary PASS/FAIL acceptance criteria for SPEC-V3R6-HOOK-CONTRACT-FIX-001. Every AC has an explicit verification command whose exit code (or stdout grep) determines PASS/FAIL. No subjective judgments.

REQ↔AC mapping is provided in plan.md § Traceability.

## AC-HCF-001: writeHookOutput emits plain-text path for WorktreeCreate

**Maps to**: REQ-HCF-001

**Given** `internal/cli/hook.go` `writeHookOutput()` is invoked with `event = hook.EventWorktreeCreate` and `input = &hook.HookInput{WorktreePath: "/test/path"}`,

**When** the function returns,

**Then** captured stdout MUST equal exactly `"/test/path\n"` (10 bytes: 10 chars including the trailing newline, no JSON braces, no leading whitespace, no trailing characters beyond the single `\n`).

**Verification command**:
```bash
go test -run 'TestWriteHookOutput_WorktreeCreatePlainText' ./internal/cli/...
```

**Expected output**: `ok  github.com/modu-ai/moai-adk/internal/cli  <time>s` (exit 0).

**Negative assertion** (must NOT appear in test output): the string `"{}"` (empty JSON object) MUST NOT appear in any captured stdout of any subtest under `TestWriteHookOutput_WorktreeCreatePlainText`.

## AC-HCF-002: writeHookOutput emits plain-text path for WorktreeRemove

**Maps to**: REQ-HCF-001

**Given** `writeHookOutput()` is invoked with `event = hook.EventWorktreeRemove` and `input.WorktreePath = "/test/path"`,

**When** the function returns,

**Then** captured stdout MUST equal exactly `"/test/path\n"`.

**Verification command**:
```bash
go test -run 'TestWriteHookOutput_WorktreeRemovePlainText' ./internal/cli/...
```

**Expected output**: exit 0.

## AC-HCF-003: settings.json + .tmpl have no WorktreeCreate / WorktreeRemove top-level hooks key

**Maps to**: REQ-HCF-003

**Given** the current state of `.claude/settings.json` (local) and `internal/template/templates/.claude/settings.json.tmpl` (template),

**When** the new CI guard test runs,

**Then** the test MUST fail if either file contains a JSON path `$.hooks.WorktreeCreate` or `$.hooks.WorktreeRemove` (i.e., the key appears at the top-level `hooks` object).

**Verification command**:
```bash
go test -run 'TestSettingsJsonHasNoWorktreeCreateKey|TestSettingsTmplHasNoWorktreeCreateKey' ./internal/template/...
```

**Expected output**: exit 0 with both tests passing on current `main` HEAD.

**Negative regression assertion**: To validate the guard, manually inject a `"WorktreeCreate": []` entry into `.claude/settings.json` in a scratch branch, re-run the test, and confirm it fails. Revert before commit. (Documented in plan.md as a manual sanity check, not part of automated AC.)

## AC-HCF-004: writeHookOutput with empty WorktreePath produces empty stdout

**Maps to**: REQ-HCF-001, REQ-HCF-004, Edge Case 9.1

**Given** `writeHookOutput()` is invoked with `event = hook.EventWorktreeCreate` and `input.WorktreePath = ""`,

**When** the function returns,

**Then** captured stdout MUST equal exactly the empty string (0 bytes, no newline).

**Verification command**:
```bash
go test -run 'TestWriteHookOutput_EmptyWorktreePathProducesEmptyStdout' ./internal/cli/...
```

**Expected output**: exit 0.

**Symmetric assertion**: the same test (or sibling subtest) MUST cover `event = hook.EventWorktreeRemove` with empty WorktreePath. Same expected behavior.

## AC-HCF-005: subagent_stop.go dispatchCapture consults CLAUDE_PROJECT_DIR before os.Getwd

**Maps to**: REQ-HCF-005

**Given** a unit test that sets `input.CWD = ""`, sets `t.Setenv("CLAUDE_PROJECT_DIR", "/test/project")`, and invokes `subagentStopHandler.dispatchCapture(input)`,

**When** the capturer is constructed,

**Then** the constructed `capture.Config{ObservationsPath}` value MUST equal `filepath.Join("/test/project", ".moai", "harness", "observations.yaml")` (i.e., the env var was consulted, NOT `os.Getwd()`).

**Verification command**:
```bash
go test -run 'TestDispatchCapture_UsesClaudeProjectDirWhenCwdEmpty' ./internal/hook/...
```

**Expected output**: exit 0.

**Implementation note**: The test MUST verify the path resolution mechanism. A clean approach is to expose `dispatchCapture` (or a path-resolution helper) for testing, or inject a `capture.New` factory function via a package-level variable that the test can override. The exact mechanism is run-phase choice; the AC asserts the OUTCOME, not the wiring.

## AC-HCF-006: internal/hook/.moai/ directory does not exist post-run

**Maps to**: REQ-HCF-006

**Given** the run-phase has executed the cleanup step,

**When** the working tree is inspected,

**Then** `internal/hook/.moai/` directory MUST NOT exist, AND `git status` MUST NOT list any path under `internal/hook/.moai/` as modified or untracked.

**Verification command**:
```bash
test ! -e internal/hook/.moai/ && \
  ( git status --porcelain internal/hook/.moai/ 2>/dev/null | grep -q . ; [ $? -ne 0 ] )
```

**Expected output**: exit 0 (both conditions hold: directory absent AND no git status entries).

**Idempotency check**: Re-running the cleanup step (effectively `rm -rf internal/hook/.moai/`) MUST NOT fail when the directory is already absent. Verification:
```bash
rm -rf internal/hook/.moai/ && rm -rf internal/hook/.moai/ ; echo "idempotent exit=$?"
```
**Expected**: `idempotent exit=0`.

## AC-HCF-007: Shell wrappers + CLI subcommands preserved byte-identical

**Maps to**: REQ-HCF-007

**Given** the run-phase has completed,

**When** the following 4 files are diffed against HEAD `58a235e06`,

**Then** the diff MUST be empty:
- `.claude/hooks/moai/handle-worktree-create.sh`
- `.claude/hooks/moai/handle-worktree-remove.sh`
- `internal/template/templates/.claude/hooks/moai/handle-worktree-create.sh`
- `internal/template/templates/.claude/hooks/moai/handle-worktree-remove.sh`

**Verification command**:
```bash
git diff --stat 58a235e06 -- \
  .claude/hooks/moai/handle-worktree-create.sh \
  .claude/hooks/moai/handle-worktree-remove.sh \
  internal/template/templates/.claude/hooks/moai/handle-worktree-create.sh \
  internal/template/templates/.claude/hooks/moai/handle-worktree-remove.sh
```

**Expected output**: empty (zero changed files reported).

**CLI subcommand preservation**: `internal/cli/hook.go` line 53 (`{"worktree-create", "Handle worktree create event", hook.EventWorktreeCreate}`) and its symmetric `worktree-remove` line MUST be unchanged.

**Verification command (CLI subcommand)**:
```bash
grep -n '"worktree-create"\|"worktree-remove"' internal/cli/hook.go
```

**Expected output**: exactly 2 lines printed, matching the original lines (line numbers may shift slightly if other parts of the file are edited, which is acceptable; the literal strings MUST be present).

## AC-HCF-008: Documentation crosswalk consistency

**Maps to**: REQ-HCF-008

**Given** the run-phase has either verified or corrected the post-PR-#1044 docs state,

**When** the following 4 inspection points are checked,

**Then** each MUST pass:

(a) `.claude/rules/moai/core/hooks-system.md` (local) AND `internal/template/templates/.claude/rules/moai/core/hooks-system.md` (template) — both files contain the literal section heading `## WorktreeCreate and WorktreeRemove Hooks` or equivalent active-creator-contract heading. Local and template differ ONLY per CLAUDE.local.md §2 (template uses `.md` extension; both are markdown).

**Verification command (a)**:
```bash
grep -c 'WorktreeCreate and WorktreeRemove\|WorktreeCreate.*active.creator\|active.creator.*WorktreeCreate' \
  .claude/rules/moai/core/hooks-system.md \
  internal/template/templates/.claude/rules/moai/core/hooks-system.md
```

**Expected output**: both files produce match count ≥ 1.

(b) `.claude/rules/moai/workflow/worktree-integration.md` + template mirror — same heading present.

**Verification command (b)**:
```bash
grep -c 'WorktreeCreate and WorktreeRemove' \
  .claude/rules/moai/workflow/worktree-integration.md \
  internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md
```

**Expected**: both files match count ≥ 1.

(c) 4 locales of `docs-site/content/{ko,en,ja,zh}/advanced/hooks-reference.md` — WorktreeCreate / WorktreeRemove documented as **MoAI 기본 비등록 + active-creator contract** (PR #1044 invariant: event catalog row preserved + contract row added; settings.json registration deliberately absent — see AC-HCF-003).

**Verification command (c-1) — event catalog presence** (WorktreeCreate/WorktreeRemove rows preserved in 4 locales):
```bash
for locale in ko en ja zh; do
  f="docs-site/content/$locale/advanced/hooks-reference.md"
  wc=$(grep -c 'WorktreeCreate' "$f")
  wr=$(grep -c 'WorktreeRemove' "$f")
  echo "$locale WorktreeCreate=$wc WorktreeRemove=$wr"
done
```

**Expected output (c-1)**: every line MUST show `WorktreeCreate>=2 WorktreeRemove>=2` (event catalog row + active-creator contract row per locale). Counts < 2 in any locale means PR #1044's contract documentation regressed.

**Verification command (c-2) — canonical handlers still documented** (locale-agnostic, no wholesale deletion):
```bash
for locale in ko en ja zh; do
  f="docs-site/content/$locale/advanced/hooks-reference.md"
  kept=$(grep -cE '\bSessionStart\b|\bSessionEnd\b|\bStop\b|\bPreToolUse\b|\bPostToolUse\b' "$f")
  echo "$locale kept=$kept"
done
```

**Expected output (c-2)**: every line MUST show `kept >= 5` (each of the 5 canonical handler names appears at least once per locale). Guards against accidental wholesale handler-section deletion.

**Verification command (c-3) — settings.json registration absence cross-check** (negative invariant):
```bash
grep -c 'WorktreeCreate\|WorktreeRemove' \
  internal/template/templates/.claude/settings.json \
  .claude/settings.json 2>/dev/null
```

**Expected output (c-3)**: both files MUST show `:0` (no registered hook key). Cross-check with AC-HCF-003 which is the primary registration-absence guard.

**Rationale**: PR #1044 contract is layered — (i) event catalog preserves WorktreeCreate as a known Claude Code event, (ii) contract section documents MoAI's non-registration policy + active-creator semantic for opt-in users, (iii) settings.json physically omits the key. AC-HCF-008(c) verifies layer (i)+(ii) on doc consistency across 4 locales; AC-HCF-003 verifies layer (iii). The 3 commands form a binary discriminator: only the post-PR-#1044 correct state satisfies ALL three.

(d) `internal/template/settings_test.go:512` — `const expectedCount = 20`.

**Verification command (d)**:
```bash
grep -n 'expectedCount = 20' internal/template/settings_test.go
```

**Expected output**: exactly 1 line matching (line number ~512).

## AC-HCF-009: Documentation explicitly describes opt-in active-creator contract

**Maps to**: REQ-HCF-009

**Given** the documentation crosswalk in AC-HCF-008 has passed,

**When** `.claude/rules/moai/core/hooks-system.md` § WorktreeCreate and WorktreeRemove Hooks is inspected,

**Then** the section MUST contain BOTH of the following substrings (or semantically equivalent text in English):

1. The phrase `plain text` (case-insensitive) OR `stdout` paired with `worktree path` — describing the active-creator contract
2. The phrase `opt-in` OR `default off` OR equivalent describing the de-registered default state

**Verification command**:
```bash
section_start=$(grep -n 'WorktreeCreate and WorktreeRemove' .claude/rules/moai/core/hooks-system.md | head -1 | cut -d: -f1)
# Extract ~80 lines starting at the section heading
sed -n "${section_start},$(($section_start + 80))p" .claude/rules/moai/core/hooks-system.md > /tmp/wt-section.md
grep -i -E '(plain.text|stdout.*worktree.path|worktree.path.*stdout)' /tmp/wt-section.md
grep -i -E '(opt.in|default.off|de.registered|not.registered)' /tmp/wt-section.md
```

**Expected output**: both grep invocations produce match count ≥ 1.

## AC-HCF-010: C-HRA-008 subagent boundary (internal/hook, internal/cli)

**Maps to**: Constraint 4.4 (forbidden AskUserQuestion call)

**Given** the run-phase has completed,

**When** the in-scope code packages are grep'd for AskUserQuestion / mcp__askuser references,

**Then** the grep MUST produce zero matches (excluding test files and pure comments).

**Verification command**:
```bash
grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ internal/cli/hook.go | \
  grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"
```

**Expected output**: empty (zero non-comment matches in non-test files).

## AC-HCF-011: Cross-platform build PASS

**Maps to**: Constraint 4.1

**Given** the run-phase has completed,

**When** the cross-platform build is invoked,

**Then** both targets MUST exit 0.

**Verification commands**:
```bash
go build ./...
GOOS=windows GOARCH=amd64 go build ./...
```

**Expected output**: both commands exit 0 with no stdout/stderr output.

## AC-HCF-012: Full test suite NEW regression count is 0

**Maps to**: Constraint 4.1 + Success Metric 8.1

**Given** the run-phase has completed,

**When** the full test suite runs,

**Then** any test failure observed MUST be either:
- (a) The baseline residual 3 FAIL documented in memory `project_v3r6_template_mirror_drift_audit_2026_05_22` (`TestRuleTemplateMirrorDrift`, `TestLateBranchTemplateMirror`, `TestSkillsContainPlanAuditGateMarkers`), OR
- (b) Out-of-scope deferred residuals from prior Wave 2 SPECs (e.g., AGENT-FOLDER-SPLIT-001's `TestAllAgentsInCatalog` post-run state if AGENT-FOLDER-SPLIT-001 has been merged before this SPEC).

NO new failure introduced by THIS SPEC's edits.

**Verification command**:
```bash
go test ./internal/cli/... ./internal/hook/... ./internal/template/... 2>&1 | \
  grep -E 'FAIL|ok' | tail -30
```

**Expected**: any FAIL lines reference tests from the baseline residual list (a) or (b). Manual cross-check against the documented baseline list.

## AC-HCF-013: spec-lint NEW regression count is 0

**Maps to**: Success Metric 8.1

**Given** the run-phase has completed,

**When** `moai spec lint` (or equivalent) is invoked,

**Then** the count of `Error` level findings MUST be equal to or less than the pre-run baseline (no NEW Errors from this SPEC).

**Verification command**:
```bash
moai spec lint 2>&1 | grep -c 'Error:'
```

**Expected**: count ≤ baseline (capture baseline at run-phase Section C pre-flight, document in progress.md).

## AC-HCF-014: Handle() byte-identity (worktree_create.go + worktree_remove.go)

**Maps to**: Constraint 4.4 (Handle() body must not change)

**Given** the run-phase has completed,

**When** the two Handle() bodies are diffed against HEAD `58a235e06`,

**Then** the bodies (lines 33-48 of worktree_create.go and lines 34-49 of worktree_remove.go) MUST be byte-identical EXCEPT for optional in-scope comment polish that does NOT change runtime behavior.

**Verification command**:
```bash
git diff 58a235e06 -- internal/hook/worktree_create.go internal/hook/worktree_remove.go | \
  grep -E '^[+-]' | grep -v '^[+-]//' | grep -v '^[+-]\*' | grep -vE '^[+-]{3}'
```

**Expected output**: empty (no non-comment line changes) — OR exactly the no-op lines if the run-phase author chose pure comment refresh.

If any non-comment line of Handle() body changes, AC-HCF-014 FAILS regardless of test pass state.

## Quality Gate Criteria

This SPEC passes its quality gate when ALL of the following hold:

- AC-HCF-001 through AC-HCF-014: PASS (14/14)
- Cross-platform build: AC-HCF-011 PASS
- C-HRA-008 boundary: AC-HCF-010 zero matches
- Handle() byte-identity: AC-HCF-014 PASS
- 1-PR-cycle (Tier S): 1 plan PR (this SPEC's plan-phase) + 1 run PR (implementing the 5 changes)

## Definition of Done

The SPEC is **implemented** when:

1. All 14 acceptance criteria pass (binary PASS/FAIL matrix)
2. SPEC frontmatter `status: draft` → `status: implemented`, `version: 0.1.0` → `version: 0.2.0`
3. progress.md captures AC matrix, baseline residual list, build/lint baselines, and run-phase deviations (if any)
4. Run PR is in MERGED state on main
5. `internal/hook/.moai/` directory verified absent in main HEAD post-merge

The SPEC is **completed** when sync-phase has also merged (documentation crosswalk verified clean on main; this is the trivial step since AC-HCF-008 already verified consistency).
