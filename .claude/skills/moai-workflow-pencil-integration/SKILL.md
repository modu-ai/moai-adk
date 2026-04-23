---
name: moai-workflow-pencil-integration
description: >
  Detects .pen files and pencil-plan.md in .moai/design/, loads Pencil MCP,
  executes batch operations with layout verification, archives screenshots.
  Invoked from /moai design Phase B2.6 when preconditions are met.
license: Apache-2.0
compatibility: Designed for Claude Code with Pencil MCP
allowed-tools: ToolSearch, Read, Write, Glob, mcp__pencil__batch_design, mcp__pencil__get_editor_state, mcp__pencil__snapshot_layout, mcp__pencil__get_screenshot, mcp__pencil__open_document, mcp__pencil__find_empty_space_on_canvas
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-04-20"
  tags: "pencil, design, mcp, batch, wireframe"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["pencil", ".pen", "batch design", "wireframe"]
  phases: ["design"]
---

# moai-workflow-pencil-integration

Pencil MCP integration skill for the `/moai design` Phase B2.6 workflow. Detects `.pen` files and `pencil-plan.md`, loads Pencil MCP tools via ToolSearch, parses DSL batch operations, executes them with layout verification, archives screenshots, and produces a run summary report.

This skill is invoked only when Phase B2.6 file/folder preconditions are satisfied. It is never invoked directly by the user.

---

## Quick Reference

### Entry Conditions (REQ-PENCIL-001)

Phase B2.6 checks these file/folder preconditions before invoking this skill:

1. `.moai/design/pencil-plan.md` exists.
2. At least one `.pen` file exists in `.moai/design/` or project root.

If either condition fails, Phase B2.6 skips gracefully without error and proceeds to Phase B3 (REQ-PENCIL-002). This skill is not invoked in that case.

When both conditions are met, this skill is invoked and must complete (success or structured error) before Phase B3 begins (REQ-PENCIL-003).

### Error Code Table (REQ-PENCIL-006, REQ-PENCIL-013, REQ-PENCIL-015, REQ-PENCIL-016)

| Code | Trigger | Recovery |
|---|---|---|
| PENCIL_MCP_UNAVAILABLE | ToolSearch with query "mcp__pencil" returns no matching tool schemas at runtime (REQ-PENCIL-013). | Return this code to the orchestrator. Phase B2.6 ends; orchestrator proceeds to Phase B3. Include Pencil MCP setup guidance: "Install and enable the Pencil MCP server. See Pencil MCP documentation for setup instructions." |
| PENCIL_CONNECTION_FAILED | `mcp__pencil__get_editor_state` returns a connection failure, or the open document name does not match the detected `.pen` file (REQ-PENCIL-006). | Return this code to the orchestrator. Phase B2.6 ends; orchestrator proceeds to Phase B3. Include user action guidance: "Restart Claude Desktop or open the correct `.pen` file in Pencil before retrying." |
| PENCIL_PLAN_SYNTAX_ERROR | `pencil-plan.md` contains tokens outside the DSL Grammar. Collect ALL errors before returning — do not halt on the first (REQ-PENCIL-016). | Return this code with a list of `{line_number, offending_text}` entries covering all discovered syntax errors. Halt execution. User must correct `pencil-plan.md` and re-run. |
| PENCIL_BATCH_FAILED | `mcp__pencil__batch_design` fails on the initial attempt AND the single retry (total 2 attempts) for a given batch (REQ-PENCIL-015). | Return this code with the failing batch index. Halt further batch execution. User must investigate Pencil state and re-run. |

---

## Implementation Guide

### Step 1: ToolSearch Preloading (REQ-PENCIL-004)

Before issuing any `mcp__pencil__*` call, invoke ToolSearch:

```
ToolSearch(query="mcp__pencil")
```

If ToolSearch returns no matching tool schemas, return `PENCIL_MCP_UNAVAILABLE` immediately. Do not attempt any `mcp__pencil__*` calls.

If ToolSearch returns one or more matching schemas, proceed to Step 2.

### Step 2: Editor State Verification (REQ-PENCIL-005)

Identify the `.pen` file detected by Phase B2.6 precondition check (the file found in `.moai/design/` or project root).

Call `mcp__pencil__get_editor_state` and inspect the response:

- If the response references the expected `.pen` file name: proceed to Step 3.
- If the response indicates a connection failure or references a different document: return `PENCIL_CONNECTION_FAILED` with recovery guidance. Do not proceed.

### Step 3: Parse pencil-plan.md (REQ-PENCIL-007, REQ-PENCIL-016)

Read `.moai/design/pencil-plan.md` and parse using the DSL Grammar below.

#### DSL Grammar

Acceptable operations (one per line):

```
I(parentId, { type: <"frame"|"text"|"rect"|"ellipse"|"line"|"group">, ...props })
M(nodeId, parentId, index)
R(nodeId)
```

Rules:
- `parentId`, `nodeId` are double-quoted string literals.
- `index` is a non-negative integer.
- `{...props}` is a JSON-serializable object literal per Pencil's `batch_design` schema.
- Line comments (`// ...`) and block comments (`/* ... */`) are permitted and ignored.
- Batch boundaries are delimited by `## Batch <N>` Markdown headings (N is a positive integer starting from 1).
- Each operation occupies exactly one line.
- Any token that does not match `I(...)`, `M(...)`, `R(...)`, a comment, or a batch heading is a syntax error.

**Syntax error collection** (REQ-PENCIL-016): Scan the entire file before returning. Collect every line containing an unrecognized token as `{line_number, offending_text}`. Do not halt on the first error. If any syntax errors are collected, return `PENCIL_PLAN_SYNTAX_ERROR` with the complete list.

**Batch ordering** (REQ-PENCIL-007): Extract batches in declaration order (Batch 1 through Batch N). Preserve this order in execution.

### Step 4: Sub-batch Splitting (REQ-PENCIL-008)

For each parsed batch, if the number of operations exceeds 25:

- Split into sub-batches of at most 25 operations each.
- Preserve the original operation order across sub-batches.
- Example: 30 operations → sub-batch A (ops 1-25) + sub-batch B (ops 26-30).

Sub-batches from the same batch are executed sequentially before moving to the next batch.

### Step 5: Batch Execution Loop (REQ-PENCIL-007, REQ-PENCIL-008, REQ-PENCIL-009, REQ-PENCIL-010, REQ-PENCIL-011, REQ-PENCIL-014, REQ-PENCIL-015)

For each batch (and each sub-batch within it):

#### 5a. Progress Update — Batch Start (REQ-PENCIL-014)

Emit a TaskUpdate marking the batch as started:

```
TaskUpdate: "Pencil batch <N> started (<op_count> operations)"
```

#### 5b. Execute Batch Design with Retry (REQ-PENCIL-015)

Call `mcp__pencil__batch_design` with the operations for this (sub-)batch.

If the call fails:
- Attempt exactly 1 retry (total 2 attempts).
- If the retry also fails: return `PENCIL_BATCH_FAILED` with the failing batch index. Halt all further execution.

If the call succeeds, capture the result. The `rootFrameId` for screenshot archival is the root frame node id returned by `batch_design` for this batch.

#### 5c. Layout Verification (REQ-PENCIL-009, REQ-PENCIL-010)

Immediately after a successful `batch_design` call, call:

```
mcp__pencil__snapshot_layout(problemsOnly=true)
```

If layout issues are reported:
- Do not proceed to the next batch or sub-batch.
- Return a structured error containing the problematic frame IDs from the snapshot result.
- Halt all further execution.

If no layout issues are reported, proceed to 5d.

#### 5d. Screenshot Archival (REQ-PENCIL-011)

Call `mcp__pencil__get_screenshot` and save the returned PNG to:

```
.moai/design/screenshots/frame-<rootFrameId>-<ISO8601-timestamp>.png
```

Where:
- `<rootFrameId>` is the root frame node id from the `batch_design` result for this batch.
- `<ISO8601-timestamp>` is the current UTC timestamp in compact format: `YYYYMMDDTHHMMSSZ` (e.g., `20260421T143022Z`).

Use the Write tool to save the PNG bytes to this path.

#### 5e. Progress Update — Batch Complete (REQ-PENCIL-014)

Emit a TaskUpdate marking the batch as complete:

```
TaskUpdate: "Pencil batch <N> complete — screenshot saved to .moai/design/screenshots/frame-<rootFrameId>-<timestamp>.png"
```

### Step 6: Screenshot Retention Cleanup (R-3 mitigation)

After all batches complete, list all files in `.moai/design/screenshots/` matching the pattern `frame-*-*.png`.

Sort by filename (ISO8601 timestamp ensures chronological order). Keep the 5 most recent files. Delete any older files using the Write tool (overwrite with empty is insufficient; use Bash if needed to remove files).

### Step 7: Summary Report (REQ-PENCIL-012)

Write a summary report to:

```
.moai/design/pencil-run-<ISO8601-timestamp>.md
```

The report must contain these sections:

```markdown
# Pencil Run — <ISO8601-timestamp>

## Applied Batches

<list each batch: batch number, operation count, sub-batch count if split>

## Screenshots Saved

<list each screenshot path saved during this run>

## Warnings

<list any non-fatal warnings collected during execution, or "None">
```

---

## Advanced Patterns

### Partial pencil-plan.md Structures

If `pencil-plan.md` contains batch headings with no operations between them, treat the batch as empty and skip it (emit no `batch_design` call, no TaskUpdate for that batch). Do not treat an empty batch as a syntax error.

### Connection Recovery

If `get_editor_state` indicates the Pencil app is not open (rather than a document mismatch), return `PENCIL_CONNECTION_FAILED` with specific guidance: "Open Pencil and load the `.pen` file before retrying."

### Large pencil-plan.md Files

Files with many batches (> 10) will generate many TaskUpdate events. This is expected and intentional per REQ-PENCIL-014 — it allows users to monitor long-running Pencil sessions and prevents timeout confusion.

---

## Works Well With

- `/moai design` Phase B2.6: The only orchestrator that invokes this skill.
- `moai-workflow-design-context`: Provides context about `.moai/design/` contents before this skill runs (Phase B2.5).
- `moai-workflow-gan-loop`: Receives Pencil wireframe artifacts for quality evaluation in Phase C.

---

REQ coverage: REQ-PENCIL-001 through REQ-PENCIL-016
SPEC: SPEC-DESIGN-PENCIL-001 v0.2.0
Version: 1.0.0
