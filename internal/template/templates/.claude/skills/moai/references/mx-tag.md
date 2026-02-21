---
name: moai-workflow-mx-tag
description: >
  @MX TAG annotation protocol reference for AI agent code context delivery.
  Provides detailed tag syntax grammar, fan-in analysis method, agent report
  format, and edge case handling. Used by manager-ddd and manager-tdd agents.
user-invocable: false
metadata:
  version: "2.5.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-22"
  tags: "mx, annotation, tag, context, invariant, danger, todo"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 3000

# MoAI Extension: Triggers
triggers:
  keywords: ["mx", "tag", "annotation", "anchor", "invariant", "context"]
  agents: ["manager-ddd", "manager-tdd", "manager-quality"]
  phases: ["run"]
---

# @MX TAG Reference -- Supplementary Details

For tag types, trigger conditions, lifecycle rules, when to add/update/remove tags, per-file limits, mandatory fields, and language settings, see the authoritative source: @.claude/rules/moai/workflow/mx-tag-protocol.md

This file contains ONLY supplementary operational details not covered in the protocol.

---

## Tag Syntax Grammar (Formal)

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

---

## Fan-In Analysis Method

Fan-in counting uses Grep-based reference analysis:

1. Extract function/method name from declaration
2. Execute `Grep(pattern="<function_name>", path=".", type="<lang>", output_mode="count")`
3. Subtract 1 for the declaration itself
4. The result is the approximate fan-in count

This is intentionally approximate. AST-level precision is not required for tagging threshold decisions. False positives (name collisions) are acceptable because ANCHOR tags are reviewed in reports.

---

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

---

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

**WHEN** a file matches a pattern in `.moai/config/sections/mx.yaml` exclude list: The agent does not add, modify, or validate @MX tags in that file.

### Broken SPEC Links

**WHEN** the ANALYZE phase detects an `@MX:SPEC: SPEC-XXX-000` reference where the SPEC file does not exist: Convert the tag to `@MX:LEGACY: true` and add an `@MX:TODO: Broken SPEC link, verify context`.

### Stale NOTE After Refactoring

**WHEN** a function signature changes: Re-review all @MX:NOTE tags on that function and update descriptions as needed.

---

Version: 2.5.0
Source: SPEC-MX-001
