# SPEC-MX-001: @MX TAG System -- MoAI eXtension Code Annotation System

| Field       | Value                                         |
|-------------|-----------------------------------------------|
| SPEC ID     | SPEC-MX-001                                   |
| Title       | @MX TAG System -- Code-Level Annotation for AI Agent Context |
| Status      | Planned                                       |
| Created     | 2026-02-20                                    |
| Author      | manager-spec                                  |
| Priority    | High                                          |
| Lifecycle   | spec-anchored                                 |
| Branch      | feature/2.5.0                                 |

---

## 1. Overview

### 1.1 Problem Statement

AI development agents (manager-ddd, manager-tdd) currently lack a structured mechanism to communicate code-level context, invariants, and danger zones between development sessions. When agents revisit files, they must re-analyze the entire context from scratch. There is no way to:

- Mark code locations where modifications carry high risk of regression
- Annotate implicit business rules embedded in code
- Track code that was written before SPEC documents existed (legacy code)
- Signal incomplete work that requires follow-up in subsequent iterations
- Preserve institutional knowledge about why code is structured in a particular way

This causes repeated analysis overhead, increased risk of breaking invariants, and loss of context across agent sessions.

### 1.2 Solution

The @MX TAG system is a code-level comment annotation system with 4 tag types that enables AI agents to understand code context, invariants, and danger zones autonomously. Tags are embedded as structured comments directly in source code, following a consistent syntax across all programming languages supported by MoAI-ADK.

Agents make all tagging decisions autonomously. Humans only receive reports summarizing tag changes.

### 1.3 Prior Art

AIDEV-NOTE system by Diwank Tomer introduced the foundational concept of AI-readable code annotations (`AIDEV-NOTE:`, `AIDEV-TODO:`, `AIDEV-QUESTION:`). The @MX system extends this with: structured sub-lines, TDD/DDD workflow integration, autonomous agent decision rules, lifecycle state machine, and per-file hard limits.

### 1.4 Scope

**In Scope:**
- 4 tag types: @MX:NOTE, @MX:WARN, @MX:ANCHOR, @MX:TODO
- Tag syntax, trigger conditions, and lifecycle rules
- TDD (RED-GREEN-REFACTOR) workflow integration
- DDD (ANALYZE-PRESERVE-IMPROVE) workflow integration
- 3-Pass Fast Tagging Algorithm for legacy bootstrap
- Configuration file (.moai/config/sections/mx.yaml) for project-level settings
- Agent autonomous report format
- MoAI skill definition (`moai-workflow-mx-tag`)
- Agent definition updates for manager-ddd and manager-tdd
- Workflow rules integration

**Out of Scope:**
- IDE plugins or editor integrations for @MX tags
- Web dashboard for tag visualization
- Cross-repository tag analysis
- Automated tag migration between programming languages
- Go binary changes to moai-adk-go (this is a template-only change)

---

## 2. Environment

### 2.1 Platform

- MoAI-ADK (Go Edition) v2.5.0+
- Claude Code v2.1.30+
- Supported programming languages: All 16+ languages supported by MoAI-ADK LSP integration (Go, Python, TypeScript, Java, Rust, C/C++, Ruby, PHP, Kotlin, Swift, Dart, Elixir, Scala, Haskell, Zig)

### 2.2 Integration Points

| System               | Integration Type   | Purpose                                    |
|----------------------|--------------------|--------------------------------------------|
| manager-ddd agent    | Skill preloading   | ANALYZE-PRESERVE-IMPROVE tag protocol      |
| manager-tdd agent    | Skill preloading   | RED-GREEN-REFACTOR tag protocol            |
| manager-quality      | Tag validation     | TRUST 5 traceability via @MX:ANCHOR       |
| .moai/config/sections/mx.yaml             | Configuration      | Per-project tag limits, thresholds, excludes|
| Workflow rules       | Rule loading       | @MX protocol as mandatory workflow rule    |

### 2.3 Dependencies

- No Go binary changes required. This is a template-only implementation.
- Depends on existing Grep, Read, Edit tools for tag scanning and insertion.
- Depends on existing fan-in analysis capability via Grep (reference counting).

---

## 3. Assumptions

- A1: Agents (manager-ddd, manager-tdd) have sufficient context window capacity to perform fan-in analysis via Grep before inserting ANCHOR tags.
- A2: The `@MX:` prefix will not conflict with existing annotations in user codebases. The prefix is sufficiently unique.
- A3: Code comments are preserved across formatting operations (gofmt, black, prettier, etc.) for all supported languages.
- A4: Agents can reliably count function references using Grep for fan-in analysis. Exact AST-level precision is not required; approximate grep-based counts are acceptable.
- A5: Users accept that agents will autonomously add, update, and remove @MX tags without prior approval. Humans are notified via reports only.
- A6: The `.moai/config/sections/mx.yaml` configuration file is optional. Sensible defaults are used when the file is absent.

---

## 4. Requirements

### 4.1 Tag Types

#### REQ-TAG-001: @MX:NOTE -- Context and Intent Delivery

**WHEN** an agent encounters a magic constant, an exported function without godoc that exceeds 100 lines, or an unexplained business rule, **THEN** the agent **shall** add an `@MX:NOTE` comment describing the context or intent.

Syntax:
```
// @MX:NOTE: [business logic / context / intent description]
// @MX:SPEC: SPEC-XXX-000    <- optional sub-line, only when SPEC exists
// @MX:LEGACY: true           <- optional, for pre-SPEC legacy code
```

Constraints:
- NOTE must be fully self-contained even without @MX:SPEC sub-line
- Author attribution: `[AUTO]` prefix for agent-generated tags, no prefix for developer-authored

Case Matrix:
- N1: TDD + SPEC exists -- include `@MX:SPEC:` sub-line
- N2: TDD + no SPEC -- omit @MX:SPEC, description must be fully self-contained
- N3: DDD legacy + no SPEC -- use `@MX:LEGACY: true`, describe historical context
- N4: DDD legacy + retroactively created SPEC -- update @MX:LEGACY to @MX:SPEC

#### REQ-TAG-002: @MX:WARN -- Danger Zone

**WHEN** an agent detects a goroutine without context.Context, cyclomatic complexity >= 15, global state mutation, or if-branches >= 8, **THEN** the agent **shall** add an `@MX:WARN` comment describing the specific danger and what happens if modified incorrectly.

Syntax:
```
// @MX:WARN: [specific danger -- what happens if modified incorrectly]
// @MX:REASON: [why it is dangerous -- MANDATORY for WARN]
// @MX:SPEC: SPEC-XXX-000    <- optional sub-line
```

Constraints:
- Hard limit: maximum 5 WARN tags per file
- @MX:REASON field is MANDATORY for every WARN
- Priority ordering P1-P5 when limit is reached; lowest priority tags are omitted

Case Matrix:
- W1: Concurrency danger (goroutine/channel) -- describe goroutine lifecycle
- W2: Complexity danger (branch/state) -- include complexity metric
- W3: External side effect (DB write, API call) -- describe side effect scope
- W4: AUTO detected -- `[AUTO]` prefix mandatory

#### REQ-TAG-003: @MX:ANCHOR -- Invariant Contract

**WHEN** an agent detects a function with fan_in >= 5 callers, a public API boundary, or an external system integration point, **THEN** the agent **shall** add an `@MX:ANCHOR` comment describing the function signature contract and why it cannot change.

Syntax:
```
// @MX:ANCHOR: [function/type signature with description]
// @MX:REASON: [why this cannot change -- MANDATORY for ANCHOR]
// @MX:SPEC: SPEC-XXX-000    <- optional sub-line
// @MX:TEST: TestFunctionName <- optional sub-line, preferred when test exists
```

Constraints:
- Hard limit: maximum 3 ANCHOR tags per file
- @MX:REASON field is MANDATORY for every ANCHOR
- Modifying ANCHOR-tagged code requires explicit mention in agent report
- ANCHOR tags are NEVER auto-deleted; demotion to NOTE requires report

Case Matrix:
- A1: High fan-in function -- REASON: "fan_in=N, called from N files"
- A2: External system boundary -- REASON: external protocol/contract
- A3: Stable contract, no SPEC -- omit @MX:SPEC, describe contract in ANCHOR
- A4: New TDD interface -- tests validate the contract

#### REQ-TAG-004: @MX:TODO -- Incomplete Work

**WHEN** an agent encounters a public function with no test file, a SPEC requirement that is not implemented, or an error returned without handling, **THEN** the agent **shall** add an `@MX:TODO` comment describing what needs to be done and the completion criteria.

Syntax:
```
// @MX:TODO: [what needs to be done + completion criteria]
// @MX:SPEC: SPEC-XXX-000    <- optional, use if SPEC requirement maps here
// @MX:PRIORITY: P1|P2|P3    <- optional (P1=immediate, P2=this sprint, P3=tech debt)
```

Constraints:
- P1 = immediate, P2 = this sprint, P3 = tech debt
- Lifecycle: Created in RED/ANALYZE phase, removed in GREEN/IMPROVE phase
- Escalation: If unresolved for > 3 iterations, TODO escalates to @MX:WARN

Case Matrix:
- T1: No test coverage -- describe what needs testing
- T2: SPEC requirement unimplemented -- use @MX:SPEC
- T3: Tech debt acknowledged -- describe the debt
- T4: AUTO detected missing test -- `[AUTO]` prefix, suggest test path

### 4.2 SPEC-ID Optionality

#### REQ-SPEC-001: SPEC References Are Fully Optional

The system **shall** treat `@MX:SPEC:` sub-lines as fully optional. Code is often generated without SPEC documents, and this is normal and accepted.

- **WHEN** a SPEC document exists for the annotated code, **THEN** the agent **shall** include an `@MX:SPEC:` sub-line referencing the SPEC ID.
- **WHEN** no SPEC document exists, **THEN** the agent **shall** omit the `@MX:SPEC:` sub-line and ensure the primary tag description is fully self-contained.
- The system **shall not** force reverse SPEC creation as an anti-pattern that this system explicitly avoids.

### 4.3 Author Attribution

#### REQ-AUTHOR-001: AUTO Prefix for Agent-Generated Tags

**WHEN** an agent autonomously generates an @MX tag, **THEN** the tag description **shall** be prefixed with `[AUTO]`.

**WHEN** a developer manually writes an @MX tag, **THEN** the tag description **shall not** include the `[AUTO]` prefix.

### 4.4 TDD Workflow Integration

#### REQ-TDD-001: RED Phase Tag Protocol

**WHEN** the agent enters the RED phase of TDD, **THEN** the agent **shall**:
1. Write a failing test
2. Scan the function signature being tested
3. **IF** fan_in >= 5 for the target function, **THEN** add @MX:ANCHOR (OK without SPEC)
4. **IF** an @MX:TODO existed for the target function, **THEN** the TODO becomes the test target and is removed after GREEN

#### REQ-TDD-002: GREEN Phase Tag Protocol

**WHEN** the agent enters the GREEN phase of TDD, **THEN** the agent **shall**:
1. Write minimal implementation
2. **IF** complex logic is found in the implementation, **THEN** add @MX:WARN (@MX:SPEC optional)
3. **IF** business intent is non-obvious, **THEN** add @MX:NOTE (self-contained without SPEC)
4. Add @MX:SPEC sub-line ONLY if a SPEC exists AND directly maps to this function

#### REQ-TDD-003: REFACTOR Phase Tag Protocol

**WHEN** the agent enters the REFACTOR phase of TDD, **THEN** the agent **shall**:
1. Re-validate all existing @MX tags after refactoring
2. **IF** fan_in changed for any ANCHOR-tagged function, **THEN** recalculate ANCHOR threshold
3. Generate an @MX Tag Report summarizing all tag changes

### 4.5 DDD Workflow Integration

#### REQ-DDD-001: ANALYZE Phase Tag Protocol

**WHEN** the agent enters the ANALYZE phase of DDD, **THEN** the agent **shall**:
1. Run the 3-Pass scan (REQ-SCAN-001)
2. Build fan-in map, detect goroutines, list magic constants
3. Generate a draft tag list with priority queue
4. Validate existing @MX tags for staleness (broken @MX:SPEC links converted to @MX:LEGACY + @MX:TODO)

#### REQ-DDD-002: PRESERVE Phase Tag Protocol

**WHEN** the agent enters the PRESERVE phase of DDD, **THEN** the agent **shall**:
1. Write characterization tests
2. Add @MX:ANCHOR to functions covered by characterization tests
3. Add @MX:WARN to goroutines and high-complexity paths
4. Add @MX:LEGACY sentinel to pre-SPEC code
5. Remove @MX:TODO when a characterization test covers the behavior

#### REQ-DDD-003: IMPROVE Phase Tag Protocol

**WHEN** the agent enters the IMPROVE phase of DDD, **THEN** the agent **shall**:
1. Make targeted changes respecting the WARN protocol (dangers are acknowledged)
2. Add @MX:NOTE for business logic exposed during improvement
3. **IF** a SPEC is retroactively created, **THEN** update @MX:LEGACY to @MX:SPEC
4. Log tag lifecycle transitions in the improvement report

### 4.6 3-Pass Fast Tagging Algorithm

#### REQ-SCAN-001: Legacy Codebase Bootstrap Scan

**WHEN** the agent encounters a legacy codebase with zero @MX tags during the DDD ANALYZE phase, **THEN** the agent **shall** execute a 3-Pass fast tagging algorithm.

**Pass 1 -- Grep Full Scan (target: 10-30 seconds):**
- Fan-in analysis via Grep (count function name references across files)
- Goroutine detection (search for `go func`, `go ` patterns)
- Magic constant detection (3+ digit numbers, decimal fractions in code)
- Exported functions without godoc
- Output: Priority queue -- P1 (fan_in>=5), P2 (goroutine/complexity>=15), P3 (magic/no-godoc), P4 (no-test)

**Pass 2 -- Selective Deep Read (P1 files only):**
- Full file Read for each P1-priority file
- Generate accurate @MX:NOTE and @MX:ANCHOR descriptions from business context
- Understand goroutine lifecycle for accurate @MX:WARN descriptions

**Pass 3 -- Batch Edit:**
- One Edit call per file
- All tags for a given file are inserted in a single Edit operation

### 4.7 Tag Lifecycle State Machine

#### REQ-LIFE-001: TODO Lifecycle

The @MX:TODO tag **shall** follow this lifecycle:
- **Created** during RED/ANALYZE phase
- **Resolved** when GREEN phase completes or test passes -- tag is REMOVED
- **Escalated** when unresolved for > 3 iterations -- promoted to @MX:WARN

#### REQ-LIFE-002: ANCHOR Lifecycle

The @MX:ANCHOR tag **shall** follow this lifecycle:
- **Created** when fan_in >= 5 is detected
- **Updated** when caller count is recalculated or SPEC is updated
- **Demoted** when fan_in drops below 3 -- proposed demotion to @MX:NOTE (agent proposes, report documents the change)
- ANCHOR tags are **NEVER auto-deleted**

#### REQ-LIFE-003: WARN Lifecycle

The @MX:WARN tag **shall** follow this lifecycle:
- **Created** when danger is detected
- **Resolved** when dangerous structure is improved -- tag is removable
- **Persistent** when danger is structural -- tag is maintained

#### REQ-LIFE-004: NOTE Lifecycle

The @MX:NOTE tag **shall** follow this lifecycle:
- **Created** when context is needed
- **Updated** when function signature changes -- content re-review triggered
- **Obsolete** when code is deleted -- removed with the code

### 4.8 Configuration

#### REQ-CONFIG-001: .moai/config/sections/mx.yaml Configuration File

The system **shall** support a `.moai/config/sections/mx.yaml` configuration file at the project root with the following structure:

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

**WHEN** `.moai/config/sections/mx.yaml` does not exist, **THEN** the system **shall** use the default values shown above.

**WHEN** `.moai/config/sections/mx.yaml` specifies `auto_tag: false`, **THEN** agents **shall not** autonomously add tags, but **shall** still validate and report on existing tags.

### 4.9 Agent Report

#### REQ-REPORT-001: Tag Change Report

**WHEN** an agent completes a DDD or TDD phase that involved @MX tag changes, **THEN** the agent **shall** generate a report in the following format:

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

### 4.10 Edge Cases

#### REQ-EDGE-001: Over-ANCHOR Prevention

**IF** a file would exceed the anchor_per_file limit (default: 3), **THEN** the agent **shall** demote excess ANCHOR tags to @MX:NOTE based on lowest fan_in count.

#### REQ-EDGE-002: Over-WARN Prevention

**IF** a file would exceed the warn_per_file limit (default: 5), **THEN** the agent **shall** keep only the P1-P5 highest priority WARNs and omit the rest.

#### REQ-EDGE-003: Stale Tag Detection

**WHEN** the ANALYZE phase runs, **THEN** the agent **shall** re-validate fan-in counts for all existing @MX:ANCHOR tags and update or demote as needed.

#### REQ-EDGE-004: ANCHOR Security Exception

**IF** an ANCHOR-tagged function requires a security patch, **THEN** the agent **shall** add `@MX:WARN: "ANCHOR breach for security"` and proceed with the modification, explicitly documenting the breach in the report.

#### REQ-EDGE-005: ANCHOR + TODO Coexistence

**WHEN** a function has both @MX:ANCHOR and @MX:TODO, **THEN** this combination is valid and **shall** be highlighted in the report as "attention required."

#### REQ-EDGE-006: Multi-Language Comment Syntax

The `@MX:` prefix pattern **shall** remain consistent across languages. Only the comment syntax varies:
- Go/Java/TypeScript/Rust/C/C++/Swift/Kotlin/Dart/Zig: `// @MX:`
- Python/Ruby/Elixir: `# @MX:`
- Haskell: `-- @MX:`

#### REQ-EDGE-007: Auto-Generated File Exclusion

**WHEN** a file matches a pattern in `.moai/config/sections/mx.yaml` exclude list, **THEN** the agent **shall not** add, modify, or validate @MX tags in that file.

#### REQ-EDGE-008: Team Environment

**WHEN** operating in Agent Teams mode, **THEN** @MX tag operations follow standard file ownership rules -- each teammate only modifies tags within their owned file patterns.

#### REQ-EDGE-009: Broken SPEC Links

**WHEN** the ANALYZE phase detects an `@MX:SPEC: SPEC-XXX-000` reference where the SPEC file does not exist, **THEN** the agent **shall** convert the tag to `@MX:LEGACY: true` and add an `@MX:TODO: Broken SPEC link, verify context`.

#### REQ-EDGE-010: Stale NOTE After Refactoring

**WHEN** a function signature changes, **THEN** the agent **shall** re-review all @MX:NOTE tags on that function and update descriptions as needed.

---

## 5. Specifications

### 5.1 Implementation Architecture

The @MX TAG system will be implemented entirely within the MoAI-ADK template system (no Go binary changes):

| Component                    | File Location (relative to `internal/template/templates/`) | Purpose                                    |
|------------------------------|------------------------------------------------------------|--------------------------------------------|
| MoAI Skill                   | `.claude/skills/moai/workflows/mx-tag.md`                  | Core @MX protocol, tag syntax, lifecycle rules, report format |
| DDD Agent Definition Update  | `.claude/agents/moai/manager-ddd.md`                       | Add `moai-workflow-mx-tag` to skills list  |
| TDD Agent Definition Update  | `.claude/agents/moai/manager-tdd.md`                       | Add `moai-workflow-mx-tag` to skills list  |
| Configuration Template       | `.moai/config/sections/mx.yaml` (project root template)                         | Default .moai/config/sections/mx.yaml with standard thresholds  |
| Workflow Rule                 | `.claude/rules/moai/workflow/mx-tag-protocol.md`           | @MX protocol rules loaded by all agents    |

### 5.2 Skill Structure: moai-workflow-mx-tag

The skill file will follow the standard MoAI skill authoring schema:

```yaml
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

progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

triggers:
  keywords: ["mx", "tag", "annotation", "anchor", "invariant", "context"]
  agents: ["manager-ddd", "manager-tdd", "manager-quality"]
  phases: ["run"]
---
```

### 5.3 Tag Syntax Grammar (Formal)

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
comment_prefix := "//" | "#" | "--"  (language-dependent)
```

### 5.4 Fan-In Analysis Method

Fan-in counting uses Grep-based reference analysis:

1. Extract function/method name from declaration
2. Execute `Grep(pattern="<function_name>", path=".", type="<lang>", output_mode="count")`
3. Subtract 1 for the declaration itself
4. The result is the approximate fan-in count

This is intentionally approximate. AST-level precision is not required for tagging threshold decisions. False positives (name collisions) are acceptable because ANCHOR tags are reviewed in reports.

### 5.5 Comment Syntax Mapping

| Language Family                  | Comment Prefix | Example                              |
|----------------------------------|---------------|--------------------------------------|
| Go, Java, TypeScript, Rust, C/C++, Swift, Kotlin, Dart, Zig, Scala | `//`  | `// @MX:NOTE: Rate limiter threshold` |
| Python, Ruby, Elixir             | `#`           | `# @MX:WARN: Thread-unsafe singleton` |
| Haskell                          | `--`          | `-- @MX:ANCHOR: Parser combinator`    |

### 5.6 Interaction with Existing MoAI Workflow

The @MX system does not replace or modify any existing workflow. It augments the DDD and TDD cycles with a new annotation layer:

```
/moai plan -> manager-spec (unchanged)
                |
/moai run  -> manager-ddd or manager-tdd
                |
                +-- ANALYZE/RED phase
                |     +-- @MX scan (new)
                |     +-- Tag validation (new)
                |
                +-- PRESERVE/GREEN phase
                |     +-- Tag insertion (new)
                |
                +-- IMPROVE/REFACTOR phase
                |     +-- Tag re-validation (new)
                |     +-- Tag report generation (new)
                |
/moai sync -> manager-docs (unchanged)
```

---

## 6. Out of Scope

- **OS-001**: IDE/editor integrations (VS Code extension, IntelliJ plugin) for @MX tag highlighting
- **OS-002**: Web dashboard or GUI for tag visualization and analytics
- **OS-003**: Cross-repository tag aggregation or analysis
- **OS-004**: Automated refactoring triggered by @MX tags (tags are informational only)
- **OS-005**: Go binary changes to `moai-adk-go` (all implementation is in templates)
- **OS-006**: Tag-based automated code generation
- **OS-007**: Integration with external issue trackers (JIRA, GitHub Issues) for @MX:TODO sync

---

## 7. Open Questions

- **OQ-001**: Should @MX:ANCHOR demotion (fan_in drops below threshold) require explicit human approval via the report, or can the agent autonomously demote? **Current decision**: Agent proposes, report documents the change. No blocking approval required.

- **OQ-002**: Should the 3-Pass scan target a specific file limit (e.g., top 20 P1 files) to prevent excessive token consumption on very large codebases? **Current decision**: Pass 2 reads only P1 files, naturally limiting scope. For codebases > 500 files, the agent may batch across multiple sessions.

- **OQ-003**: Should @MX:TODO with P1 priority trigger automatic test generation, or should it remain informational? **Current decision**: Informational only. The TODO serves as a marker for the next RED phase.

- **OQ-004**: How should @MX tags interact with code review tools (GitHub PR review)? Should tags be highlighted in PR descriptions? **Deferred**: This is a downstream concern for manager-git and manager-docs agents, not the core @MX system.

---

## 8. Traceability

| Requirement    | Acceptance Criteria | Plan Reference |
|----------------|---------------------|----------------|
| REQ-TAG-001    | AC-NOTE-001         | MILE-1         |
| REQ-TAG-002    | AC-WARN-001         | MILE-1         |
| REQ-TAG-003    | AC-ANCHOR-001       | MILE-1         |
| REQ-TAG-004    | AC-TODO-001         | MILE-1         |
| REQ-SPEC-001   | AC-SPEC-001         | MILE-1         |
| REQ-AUTHOR-001 | AC-AUTHOR-001       | MILE-1         |
| REQ-TDD-001    | AC-TDD-001          | MILE-2         |
| REQ-TDD-002    | AC-TDD-002          | MILE-2         |
| REQ-TDD-003    | AC-TDD-003          | MILE-2         |
| REQ-DDD-001    | AC-DDD-001          | MILE-2         |
| REQ-DDD-002    | AC-DDD-002          | MILE-2         |
| REQ-DDD-003    | AC-DDD-003          | MILE-2         |
| REQ-SCAN-001   | AC-SCAN-001         | MILE-3         |
| REQ-LIFE-001   | AC-LIFE-001         | MILE-1         |
| REQ-LIFE-002   | AC-LIFE-002         | MILE-1         |
| REQ-LIFE-003   | AC-LIFE-003         | MILE-1         |
| REQ-LIFE-004   | AC-LIFE-004         | MILE-1         |
| REQ-CONFIG-001 | AC-CONFIG-001       | MILE-1         |
| REQ-REPORT-001 | AC-REPORT-001       | MILE-2         |
| REQ-EDGE-001   | AC-EDGE-001         | MILE-1         |
| REQ-EDGE-002   | AC-EDGE-002         | MILE-1         |
| REQ-EDGE-003   | AC-EDGE-003         | MILE-2         |
| REQ-EDGE-004   | AC-EDGE-004         | MILE-2         |
| REQ-EDGE-005   | AC-EDGE-005         | MILE-1         |
| REQ-EDGE-006   | AC-EDGE-006         | MILE-1         |
| REQ-EDGE-007   | AC-EDGE-007         | MILE-1         |
| REQ-EDGE-008   | AC-EDGE-008         | MILE-2         |
| REQ-EDGE-009   | AC-EDGE-009         | MILE-2         |
| REQ-EDGE-010   | AC-EDGE-010         | MILE-2         |
