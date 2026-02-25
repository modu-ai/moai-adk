---
name: moai-workflow-clean
description: >
  Identify and safely remove dead code with test verification.
  Uses static analysis, usage graph analysis, and safe removal with rollback.
  Supports dry-run preview and file-targeted analysis.
  Use when removing unused code, cleaning up dead imports, or reducing codebase size.
user-invocable: false
metadata:
  version: "2.5.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-21"
  tags: "clean, dead-code, unused, refactoring, static-analysis"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 4000

# MoAI Extension: Triggers
triggers:
  keywords: ["clean", "dead code", "unused code", "dead-code", "remove unused"]
  agents: ["expert-refactoring", "expert-testing"]
  phases: ["clean"]
---

# Workflow: Clean - Dead Code Removal

Purpose: Identify and safely remove unused code through static analysis, usage graph traversal, and test-verified removal. Ensures no regressions are introduced.

Flow: Static Analysis -> Usage Graph -> Classification -> Safe Removal -> Test Verification -> Report

## Supported Flags

- --dry (alias --dry-run): Preview dead code without removing anything
- --safe-only: Only remove confirmed dead code (skip uncertain cases)
- --file PATH: Target specific file or directory for analysis
- --type TYPE: Focus on specific code type (functions, imports, types, variables, files)
- --aggressive: Include code with low usage (1 caller that is also dead)

## Phase 1: Static Analysis Scan

[HARD] Delegate static analysis to the expert-refactoring subagent.

Language-specific dead code detection:

- Go: `go vet ./...` for unused variables, `staticcheck` for unused functions/types, `deadcode` tool
- Python: `vulture` for dead code detection, `autoflake` for unused imports
- TypeScript/JavaScript: `ts-prune` for unused exports, ESLint `no-unused-vars`
- Rust: `cargo clippy` for dead code warnings, `cargo udeps` for unused dependencies

If --file flag: Limit scan to the specified file/directory.
If --type flag: Filter results to the specified code type only.

Scan Categories:

- Unused imports: Import statements with no references
- Unused variables: Declared but never read
- Unused functions: Defined but never called
- Unused types: Type definitions with no usage
- Unused files: Files with no incoming imports
- Dead dependencies: Packages installed but never imported

## Phase 2: Usage Graph Analysis

[HARD] Delegate usage graph analysis to the expert-refactoring subagent.

Build a usage graph to verify static analysis results:

- For each candidate: Grep all references across the codebase
- Check indirect usage (via interfaces, reflection, dynamic dispatch)
- Check test-only usage (used only in tests, not production code)
- Check conditional compilation (#ifdef, build tags, env-based imports)
- Check external usage (exported APIs that may be used by other projects)

Classification Results:

- Confirmed Dead: No references found anywhere in codebase
- Test-Only: Used only in test files (may indicate test-specific utilities)
- Likely Dead: Low confidence (dynamic usage possible)
- False Positive: Actually used (via reflection, plugins, external consumers)

MX Tag Cross-Check (Pre-Removal Safety):

After classification, cross-check all candidates against existing @MX tags:
- @MX:ANCHOR candidates: Reclassify from "Confirmed Dead" to "False Positive" (ANCHOR indicates high fan_in; dynamic or cross-module usage is likely)
- @MX:WARN candidates: Flag for manual review even if classified as "Confirmed Dead" (warned code may have hidden dependencies)
- @MX:NOTE candidates: Include the NOTE context in the removal plan for informed user decision
- @MX:TODO candidates: If TODO indicates pending work, reclassify as "Deferred" rather than dead
- This cross-check supplements the Phase 4 safety measure: "Never remove @MX:ANCHOR without explicit approval"
- See @.claude/rules/moai/workflow/mx-tag-protocol.md for tag type definitions

If --safe-only flag: Only proceed with "Confirmed Dead" items (after MX cross-check).
If --aggressive flag: Include "Likely Dead" items for removal (MX cross-check still applies).

## Phase 3: Removal Plan

Present removal plan via AskUserQuestion (unless --dry flag):

```markdown
## Dead Code Analysis Results

### Confirmed Dead (safe to remove)
- file.go: UnusedFunction (0 references)
- file.go: unusedVariable (0 references)
- unused_file.go: Entire file (0 imports)

### Test-Only Usage
- file.go: TestHelper (used in 2 test files only)

### Likely Dead (uncertain)
- file.go: MaybeUsed (1 reference in dead code chain)

### Summary
- Total candidates: N
- Safe to remove: N
- Lines to be removed: N
```

Options:

- Remove confirmed dead code (Recommended): Remove all items classified as "Confirmed Dead". This is the safest option with minimal risk of breaking anything. Tests will verify no regressions.
- Remove confirmed + test-only: Also remove test-only utilities that are no longer needed. Choose this for a more thorough cleanup.
- Review each item: Review each dead code candidate individually before deciding. MoAI will present them one by one for your approval.
- Cancel: Do not remove any code.

If --dry flag: Display analysis results and exit without removing anything.

## Phase 4: Safe Removal

[HARD] Delegate removal to the expert-refactoring subagent.

Removal Strategy:

1. Create removal order based on dependency graph (leaf nodes first)
2. For each removal:
   - Remove the dead code using Edit tool
   - Update any affected imports
   - Clean up empty files if all exports removed
3. After each batch of removals, run tests to verify

Safety Measures:

- Remove in reverse dependency order (callees before callers)
- Group related removals (function + its private helpers)
- Preserve @MX tags for remaining code (update if references change)
- Never remove code with @MX:ANCHOR tag without explicit approval

## Phase 5: Test Verification

[HARD] Delegate test verification to the expert-testing subagent.

After removals:
- Run full test suite: `go test -race ./...` (Go) or equivalent
- Verify no test failures
- Check that no new linting errors were introduced
- Confirm build succeeds

If tests fail:
- Identify which removal caused the failure
- Rollback that specific removal
- Mark the item as "False Positive" in the report
- Continue with remaining removals

## Phase 5.5: MX Tag Cleanup

After verified removals:
- Remove @MX tags from deleted code
- Update @MX:ANCHOR fan_in counts if callers were removed
- Demote @MX:ANCHOR to @MX:NOTE if fan_in drops below 3
- Generate MX tag change report

## Phase 6: Report

Display removal report in user's conversation_language:

```markdown
## Dead Code Removal Report

### Removed: N items (M lines)
- file.go: UnusedFunction (15 lines)
- file.go: unusedVariable (1 line)
- unused_file.go: Entire file deleted (120 lines)

### Kept (false positives): N items
- file.go: DynamicHandler (used via reflection)

### Test Results: PASS (all tests green)

### Codebase Reduction
- Files removed: N
- Lines removed: M
- Dependencies removed: K
```

Next Steps (AskUserQuestion):

- Commit changes (Recommended): Create a git commit with the dead code removal. The commit message will list all removed items for traceability.
- Run coverage analysis: Check if the removal affected test coverage. Dead test-only code removal may change coverage percentages.
- Review removed items: See the full diff of all removals for manual verification before committing.

## Task Tracking

[HARD] Task management tools mandatory:
- Each dead code candidate tracked as a pending task via TaskCreate
- Before removal: change to in_progress via TaskUpdate
- After verified removal: change to completed via TaskUpdate
- False positives marked as completed with note

## Agent Chain Summary

- Phase 1: expert-refactoring subagent (static analysis)
- Phase 2: expert-refactoring subagent (usage graph analysis)
- Phase 3: MoAI orchestrator (user approval via AskUserQuestion)
- Phase 4: expert-refactoring subagent (safe removal)
- Phase 5: expert-testing subagent (test verification)
- Phase 5.5: MoAI orchestrator or expert-refactoring (MX tag cleanup)

## Execution Summary

1. Parse arguments (extract flags: --dry, --safe-only, --file, --type, --aggressive)
2. Delegate static analysis scan to expert-refactoring subagent
3. Delegate usage graph analysis to expert-refactoring subagent
4. Cross-check candidates against @MX tags (MX Tag Cross-Check)
5. Classify results (Confirmed Dead, Test-Only, Likely Dead, False Positive)
6. If --dry: Display analysis results and exit
7. Present removal plan to user via AskUserQuestion
8. Delegate safe removal to expert-refactoring subagent
9. Delegate test verification to expert-testing subagent
10. Clean up @MX tags for removed code (Phase 5.5)
11. TaskCreate/TaskUpdate for all candidates
12. Report results with next step options

---

Version: 1.1.0
Updated: 2026-02-25. Added MX Tag Cross-Check in Phase 2 for pre-removal safety validation.
