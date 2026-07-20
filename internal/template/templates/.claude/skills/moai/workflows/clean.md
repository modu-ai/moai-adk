---
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
  agents: ["manager-develop"]
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

## Pipeline Contract (Agentless Classification)

<!-- @MX:NOTE - Agentless fixed-pipeline classification; localize→repair→validate contract. See spec-workflow.md#subcommand-classification. -->

This subcommand is classified as **Agentless fixed-pipeline**.
It executes a deterministic 3-phase contract: **localize → repair → validate**.

- **Phase mapping**: localize ← Phase 1+2; repair ← Phase 4; validate ← Phase 5+5.5
- **No LLM-driven control flow**: Agent() invocations exist for executor delegation within phases (e.g., a per-spawn `Agent(general-purpose)` refactoring specialist for removal, per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C) but never select the next phase.
- **No-op exit**: When the localize phase finds zero targets, the pipeline exits with status `no-op` and exit code 0, skipping repair and validate.
- **Fail-fast**: When repair encounters an unresolvable error, the pipeline terminates and reports the error. There is no multi-agent fallback.
- **`--mode` flag handling**: Any `--mode` flag passed to this subcommand is ignored. The system logs `MODE_FLAG_IGNORED_FOR_UTILITY` at info level and proceeds with the fixed pipeline.
- **Repeatability**: Even when the parent invocation supplies `--mode loop`, the pipeline runs once per command invocation. Re-entry requires explicit user re-invocation.

See [Subcommand Classification matrix](../../rules/moai/workflow/spec-workflow.md#subcommand-classification) for the full pipeline-vs-multi-agent contract.

## Phase 1: Static Analysis Scan

[HARD] Delegate static analysis AND usage-graph analysis (Phases 1+2) to ONE combined per-spawn `Agent(general-purpose)` refactoring specialist (refactoring whitelist + ANALYZE-PRESERVE-IMPROVE instructions per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C). Inject `At start, invoke Skill("moai-workflow-ddd") for the ANALYZE-PRESERVE-IMPROVE cycle.` (per `.moai/config/sections/delegation.yaml`; skill-routing.md §1). The two phases share an identical whitelist and role, so a single combined analysis spawn performs Phase 1 then Phase 2 in one delegation — replacing the former two-spawn chain.

Language-specific dead code detection (all 16 supported languages, equal treatment):

| Language | Detection Tool |
|----------|----------------|
| Go | `go vet ./...` for unused variables, `staticcheck` for unused functions/types, `deadcode` tool |
| Python | `vulture` for dead code detection, `autoflake` for unused imports |
| TypeScript | `ts-prune` for unused exports, ESLint `no-unused-vars` |
| JavaScript | ESLint `no-unused-vars`, `depcheck` for unused dependencies |
| Rust | `cargo clippy` for dead code warnings, `cargo udeps` for unused dependencies |
| Java | `PMD` UnusedPrivateField/UnusedLocalVariable rules |
| Kotlin | `detekt` unused-code rules |
| C# | Roslyn unused-member analyzers via `dotnet format analyzers` |
| Ruby | `rubocop` Lint/UselessAssignment cop |
| PHP | `phpstan` dead-code rules |
| Elixir | `mix xref unreachable`, unused-variable compile warnings |
| C++ | `clang-tidy` unused-code checks, `cppcheck --enable=unusedFunction` |
| Scala | `scalafix` unused-imports rule, `scalac -Ywarn-unused` |
| R | `lintr` unused-object checks |
| Flutter | `dart analyze` unused_element/unused_import lints |
| Swift | `periphery` for unused code detection |

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

Performed by the SAME combined analysis spawn as Phase 1 (no separate Agent() spawn — see Phase 1).

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
- See .claude/rules/moai/workflow/mx-tag-protocol.md for tag type definitions

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

<!-- @MX:WARN @MX:REASON - Phase 4 delegates to a per-spawn general-purpose refactoring specialist for the *executor* role. Do NOT extend this delegation to choose between Phase 4 and Phase 5; that would violate the agentless-pipeline contract. -->

[HARD] Delegate safe removal AND test verification (Phases 4+5) to ONE combined per-spawn `Agent(general-purpose)` refactoring specialist — a single removal+verification spawn performs Phase 4 then Phase 5 in one delegation, replacing the former separate removal and verification spawns. Inject `At start, invoke Skill("moai-workflow-ddd") for the ANALYZE-PRESERVE-IMPROVE cycle.` (per `.moai/config/sections/delegation.yaml`; skill-routing.md §1).

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

Performed by the SAME combined removal+verification spawn as Phase 4 (no separate Agent() spawn — see Phase 4).

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

## Phase 6: MX Tag Cleanup

Executes orchestrator-direct (no Agent() spawn) — the worst-case spawn count for this workflow stays at 2 (the combined 1+2 analysis spawn and the combined 4+5 removal+verification spawn).

After verified removals:
- Remove @MX tags from deleted code
- Update @MX:ANCHOR fan_in counts if callers were removed
- Demote @MX:ANCHOR to @MX:NOTE if fan_in drops below 3
- Generate MX tag change report

## Phase 7: Report

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

- Phase 1+2: ONE combined per-spawn `Agent(general-purpose)` refactoring specialist (static analysis + usage graph analysis — spawn 1 of 2)
- Phase 3: MoAI orchestrator (user approval via AskUserQuestion)
- Phase 4+5: ONE combined per-spawn `Agent(general-purpose)` refactoring specialist (safe removal + test verification — spawn 2 of 2)
- Phase 6: MoAI orchestrator (MX tag cleanup, orchestrator-direct — no spawn)

## Execution Summary

1. Parse arguments (extract flags: --dry, --safe-only, --file, --type, --aggressive)
2. Delegate static analysis scan + usage graph analysis (Phases 1+2) to ONE combined per-spawn `Agent(general-purpose)` refactoring specialist (spawn 1 of 2)
3. Cross-check candidates against @MX tags (MX Tag Cross-Check)
4. Classify results (Confirmed Dead, Test-Only, Likely Dead, False Positive)
5. If --dry: Display analysis results and exit
6. Present removal plan to user via AskUserQuestion
7. Delegate safe removal + test verification (Phases 4+5) to ONE combined per-spawn `Agent(general-purpose)` refactoring specialist (spawn 2 of 2)
8. Clean up @MX tags for removed code orchestrator-direct (Phase 6)
9. TaskCreate/TaskUpdate for all candidates
10. Report results with next step options

---

Version: 1.1.0
Updated: 2026-02-25. Added MX Tag Cross-Check in Phase 2 for pre-removal safety validation.
