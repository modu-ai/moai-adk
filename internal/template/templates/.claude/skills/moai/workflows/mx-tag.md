---
name: moai-workflow-mx-tag
description: >
  @MX TAG annotation protocol for AI agent code context delivery.
  Defines 4 tag types (NOTE, WARN, ANCHOR, TODO), trigger conditions,
  lifecycle state machine, TDD/DDD workflow integration, and autonomous
  report generation. Used by manager-ddd and manager-tdd agents.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob Edit
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-20"
  modularized: "false"
  tags: "mx, annotation, tag, context, invariant, danger, todo"
  related-skills: "moai-workflow-ddd, moai-workflow-tdd, moai-workflow-testing"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["mx", "tag", "annotation", "anchor", "invariant", "context"]
  agents: ["manager-ddd", "manager-tdd", "manager-quality"]
  phases: ["run"]
---

# @MX TAG System -- Code-Level Annotation for AI Agent Context

Purpose: Enable AI agents to communicate code-level context, invariants, and danger zones between development sessions through structured comment annotations.

## Overview

The @MX TAG system is a code-level annotation protocol with 4 tag types that enables AI agents to understand code context autonomously. Tags are embedded as structured comments directly in source code.

**Core Principles:**
- Agents make all tagging decisions autonomously
- Humans only receive reports summarizing tag changes
- SPEC references are optional -- code exists without SPECs
- Tags support both greenfield (TDD) and brownfield (DDD) workflows

## Tag Syntax Grammar

```
mx_tag       := comment_prefix SPACE "@MX:" tag_type ":" SPACE description NEWLINE sub_lines*
tag_type     := "NOTE" | "WARN" | "ANCHOR" | "TODO"
description  := [auto_prefix] free_text
auto_prefix  := "[AUTO]" SPACE
sub_lines    := comment_prefix SPACE "@MX:" sub_key ":" SPACE sub_value NEWLINE
sub_key      := "SPEC" | "LEGACY" | "REASON" | "TEST" | "PRIORITY"
sub_value    := (SPEC: spec_id) | (LEGACY: "true") | (REASON: free_text) | (TEST: test_name) | (PRIORITY: priority_level)
spec_id      := "SPEC-" UPPER+ "-" DIGIT{3}
priority_level := "P1" | "P2" | "P3"
```

## Tag Types

### @MX:NOTE -- Context and Intent Delivery

**WHEN** to use: Magic constants, exported functions without godoc exceeding 100 lines, unexplained business rules.

**Syntax:**
```
// @MX:NOTE: [business logic / context / intent description]
// @MX:SPEC: SPEC-XXX-000    <- optional sub-line, only when SPEC exists
// @MX:LEGACY: true           <- optional, for pre-SPEC legacy code
```

**Constraints:**
- NOTE must be fully self-contained even without @MX:SPEC sub-line
- Author attribution: `[AUTO]` prefix for agent-generated tags

**Case Matrix:**
- **N1** (TDD + SPEC exists): Include `@MX:SPEC:` sub-line
- **N2** (TDD + no SPEC): Omit @MX:SPEC, description must be fully self-contained
- **N3** (DDD legacy + no SPEC): Use `@MX:LEGACY: true`, describe historical context
- **N4** (DDD legacy + retroactively created SPEC): Update @MX:LEGACY to @MX:SPEC

### @MX:WARN -- Danger Zone

**WHEN** to use: Goroutine without context.Context, cyclomatic complexity >= 15, global state mutation, if-branches >= 8.

**Syntax:**
```
// @MX:WARN: [specific danger -- what happens if modified incorrectly]
// @MX:REASON: [why it is dangerous -- MANDATORY for WARN]
// @MX:SPEC: SPEC-XXX-000    <- optional sub-line
```

**Constraints:**
- Hard limit: maximum 5 WARN tags per file
- @MX:REASON field is MANDATORY for every WARN
- Priority ordering P1-P5 when limit is reached; lowest priority tags are omitted

**Case Matrix:**
- **W1** (Concurrency danger): Describe goroutine lifecycle
- **W2** (Complexity danger): Include complexity metric
- **W3** (External side effect): Describe side effect scope
- **W4** (AUTO detected): `[AUTO]` prefix mandatory

### @MX:ANCHOR -- Invariant Contract

**WHEN** to use: Function with fan_in >= 5 callers, public API boundary, external system integration point.

**Syntax:**
```
// @MX:ANCHOR: [function/type signature with description]
// @MX:REASON: [why this cannot change -- MANDATORY for ANCHOR]
// @MX:SPEC: SPEC-XXX-000    <- optional sub-line
// @MX:TEST: TestFunctionName <- optional sub-line, preferred when test exists
```

**Constraints:**
- Hard limit: maximum 3 ANCHOR tags per file
- @MX:REASON field is MANDATORY for every ANCHOR
- Modifying ANCHOR-tagged code requires explicit mention in agent report
- ANCHOR tags are NEVER auto-deleted; demotion to NOTE requires report

**Case Matrix:**
- **A1** (High fan-in function): REASON: "fan_in=N, called from N files"
- **A2** (External system boundary): REASON: external protocol/contract
- **A3** (Stable contract, no SPEC): Omit @MX:SPEC, describe contract in ANCHOR
- **A4** (New TDD interface): Tests validate the contract

### @MX:TODO -- Incomplete Work

**WHEN** to use: Public function with no test file, SPEC requirement not implemented, error returned without handling.

**Syntax:**
```
// @MX:TODO: [what needs to be done + completion criteria]
// @MX:SPEC: SPEC-XXX-000    <- optional, use if SPEC requirement maps here
// @MX:PRIORITY: P1|P2|P3    <- optional (P1=immediate, P2=this sprint, P3=tech debt)
```

**Constraints:**
- P1 = immediate, P2 = this sprint, P3 = tech debt
- Lifecycle: Created in RED/ANALYZE phase, removed in GREEN/IMPROVE phase
- Escalation: If unresolved for > 3 iterations, TODO escalates to @MX:WARN

**Case Matrix:**
- **T1** (No test coverage): Describe what needs testing
- **T2** (SPEC requirement unimplemented): Use @MX:SPEC
- **T3** (Tech debt acknowledged): Describe the debt
- **T4** (AUTO detected missing test): `[AUTO]` prefix, suggest test path

## SPEC-ID Optionality

[HARD] SPEC References Are Fully Optional

The system treats `@MX:SPEC:` sub-lines as fully optional. Code is often generated without SPEC documents, and this is normal and accepted.

- **WHEN** a SPEC document exists for the annotated code: Include an `@MX:SPEC:` sub-line
- **WHEN** no SPEC document exists: Omit the `@MX:SPEC:` sub-line and ensure the primary tag description is fully self-contained
- The system does not force reverse SPEC creation as an anti-pattern

## Author Attribution

[HARD] AUTO Prefix for Agent-Generated Tags

- **WHEN** an agent autonomously generates an @MX tag: The tag description SHALL be prefixed with `[AUTO]`
- **WHEN** a developer manually writes an @MX tag: The tag description SHALL NOT include the `[AUTO]` prefix

## TDD Workflow Integration

### RED Phase Tag Protocol

**WHEN** entering the RED phase of TDD:
1. Write a failing test
2. Scan the function signature being tested
3. **IF** fan_in >= 5 for the target function: Add @MX:ANCHOR (OK without SPEC)
4. **IF** an @MX:TODO existed for the target function: The TODO becomes the test target and is removed after GREEN

### GREEN Phase Tag Protocol

**WHEN** entering the GREEN phase of TDD:
1. Write minimal implementation
2. **IF** complex logic is found in the implementation: Add @MX:WARN (@MX:SPEC optional)
3. **IF** business intent is non-obvious: Add @MX:NOTE (self-contained without SPEC)
4. Add @MX:SPEC sub-line ONLY if a SPEC exists AND directly maps to this function

### REFACTOR Phase Tag Protocol

**WHEN** entering the REFACTOR phase of TDD:
1. Re-validate all existing @MX tags after refactoring
2. **IF** fan_in changed for any ANCHOR-tagged function: Recalculate ANCHOR threshold
3. Generate an @MX Tag Report summarizing all tag changes

## DDD Workflow Integration

### ANALYZE Phase Tag Protocol

**WHEN** entering the ANALYZE phase of DDD:
1. Run the 3-Pass scan (see 3-Pass Fast Tagging Algorithm below)
2. Build fan-in map, detect goroutines, list magic constants
3. Generate a draft tag list with priority queue
4. Validate existing @MX tags for staleness (broken @MX:SPEC links converted to @MX:LEGACY + @MX:TODO)

### PRESERVE Phase Tag Protocol

**WHEN** entering the PRESERVE phase of DDD:
1. Write characterization tests
2. Add @MX:ANCHOR to functions covered by characterization tests
3. Add @MX:WARN to goroutines and high-complexity paths
4. Add @MX:LEGACY sentinel to pre-SPEC code
5. Remove @MX:TODO when a characterization test covers the behavior

### IMPROVE Phase Tag Protocol

**WHEN** entering the IMPROVE phase of DDD:
1. Make targeted changes respecting the WARN protocol (dangers are acknowledged)
2. Add @MX:NOTE for business logic exposed during improvement
3. **IF** a SPEC is retroactively created: Update @MX:LEGACY to @MX:SPEC
4. Log tag lifecycle transitions in the improvement report

## 3-Pass Fast Tagging Algorithm

**WHEN** encountering a legacy codebase with zero @MX tags during the DDD ANALYZE phase: Execute a 3-Pass fast tagging algorithm.

### Pass 1 -- Grep Full Scan (target: 10-30 seconds)

- Fan-in analysis via Grep (count function name references across files)
- Goroutine detection (search for `go func`, `go ` patterns)
- Magic constant detection (3+ digit numbers, decimal fractions in code)
- Exported functions without godoc

**Output:** Priority queue:
- P1: fan_in >= 5
- P2: goroutine or complexity >= 15
- P3: magic constants or no-godoc
- P4: no-test

### Pass 2 -- Selective Deep Read (P1 files only)

- Full file Read for each P1-priority file
- Generate accurate @MX:NOTE and @MX:ANCHOR descriptions from business context
- Understand goroutine lifecycle for accurate @MX:WARN descriptions

### Pass 3 -- Batch Edit

- One Edit call per file
- All tags for a given file are inserted in a single Edit operation

## Tag Lifecycle State Machine

### TODO Lifecycle

- **Created**: During RED/ANALYZE phase
- **Resolved**: When GREEN phase completes or test passes -- tag is REMOVED
- **Escalated**: When unresolved for > 3 iterations -- promoted to @MX:WARN

### ANCHOR Lifecycle

- **Created**: When fan_in >= 5 is detected
- **Updated**: When caller count is recalculated or SPEC is updated
- **Demoted**: When fan_in drops below 3 -- proposed demotion to @MX:NOTE (agent proposes, report documents the change)
- **Never auto-deleted**: ANCHOR tags are NEVER auto-deleted

### WARN Lifecycle

- **Created**: When danger is detected
- **Resolved**: When dangerous structure is improved -- tag is removable
- **Persistent**: When danger is structural -- tag is maintained

### NOTE Lifecycle

- **Created**: When context is needed
- **Updated**: When function signature changes -- content re-review triggered
- **Obsolete**: When code is deleted -- removed with the code

## Configuration (.mx.yaml)

The system supports a `.mx.yaml` configuration file at the project root:

```yaml
mx:
  version: "1.0"
  exclude:
    - "**/*_generated.go"
    - "**/vendor/**"
    - "**/mock_*.go"
  limits:
    anchor_per_file: 3
    warn_per_file: 5
  thresholds:
    fan_in_anchor: 5
    complexity_warn: 15
    branch_warn: 8
  auto_tag: true
  require_reason_for:
    - ANCHOR
    - WARN
```

**WHEN** `.mx.yaml` does not exist: Use the default values shown above.

**WHEN** `.mx.yaml` specifies `auto_tag: false`: Agents do not autonomously add tags, but still validate and report on existing tags.

## Fan-In Analysis Method

Fan-in counting uses Grep-based reference analysis:

1. Extract function/method name from declaration
2. Execute `Grep(pattern="<function_name>", path=".", type="<lang>", output_mode="count")`
3. Subtract 1 for the declaration itself
4. The result is the approximate fan-in count

This is intentionally approximate. AST-level precision is not required for tagging threshold decisions. False positives (name collisions) are acceptable because ANCHOR tags are reviewed in reports.

## Multi-Language Comment Syntax

The `@MX:` prefix pattern remains consistent across languages. Only the comment syntax varies:

| Language Family | Comment Prefix | Example |
|-----------------|---------------|---------|
| Go, Java, TypeScript, Rust, C/C++, Swift, Kotlin, Dart, Zig, Scala | `//` | `// @MX:NOTE: Rate limiter threshold` |
| Python, Ruby, Elixir | `#` | `# @MX:WARN: Thread-unsafe singleton` |
| Haskell | `--` | `-- @MX:ANCHOR: Parser combinator` |

## Agent Report Format

**WHEN** completing a DDD or TDD phase that involved @MX tag changes: Generate a report in the following format:

```markdown
## @MX Tag Report -- [Phase] -- [Timestamp]

### Tags Added (N new)
- FILE:LINE @MX:ANCHOR reason_summary [fan_in=N]
- FILE:LINE @MX:WARN reason_summary [concurrency]

### Tags Removed (N removed)
- FILE:LINE @MX:TODO -> resolved by [TestName]

### Tags Updated (N updated)
- FILE:LINE @MX:NOTE -> updated after signature change

### Attention Required
- FILE:LINE @MX:ANCHOR + @MX:TODO coexistence -> review needed
```

## Edge Cases

### Over-ANCHOR Prevention

**IF** a file would exceed the anchor_per_file limit (default: 3): Demote excess ANCHOR tags to @MX:NOTE based on lowest fan_in count.

### Over-WARN Prevention

**IF** a file would exceed the warn_per_file limit (default: 5): Keep only the P1-P5 highest priority WARNs and omit the rest.

### Stale Tag Detection

**WHEN** the ANALYZE phase runs: Re-validate fan-in counts for all existing @MX:ANCHOR tags and update or demote as needed.

### ANCHOR Security Exception

**IF** an ANCHOR-tagged function requires a security patch: Add `@MX:WARN: "ANCHOR breach for security"` and proceed with the modification, explicitly documenting the breach in the report.

### ANCHOR + TODO Coexistence

**WHEN** a function has both @MX:ANCHOR and @MX:TODO: This combination is valid and SHALL be highlighted in the report as "attention required."

### Auto-Generated File Exclusion

**WHEN** a file matches a pattern in `.mx.yaml` exclude list: The agent does not add, modify, or validate @MX tags in that file.

### Team Environment

**WHEN** operating in Agent Teams mode: @MX tag operations follow standard file ownership rules -- each teammate only modifies tags within their owned file patterns.

### Broken SPEC Links

**WHEN** the ANALYZE phase detects an `@MX:SPEC: SPEC-XXX-000` reference where the SPEC file does not exist: Convert the tag to `@MX:LEGACY: true` and add an `@MX:TODO: Broken SPEC link, verify context`.

### Stale NOTE After Refactoring

**WHEN** a function signature changes: Re-review all @MX:NOTE tags on that function and update descriptions as needed.

---

Version: 1.0.0
Source: SPEC-MX-001
