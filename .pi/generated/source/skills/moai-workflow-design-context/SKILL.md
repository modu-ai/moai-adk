---
name: moai-workflow-design-context
description: >
  Loads human-authored design briefs from .moai/design/ (research, system, spec,
  pencil-plan) and injects them into /moai design workflow context with priority
  truncation when token budget is exceeded.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-04-20"
  tags: "design, context, attach, auto-load, brand, design-brief"

progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

triggers:
  keywords: ["design context", "attach design", "design brief", ".moai/design"]
  agents: ["expert-frontend"]
  phases: ["design"]
---

# moai-workflow-design-context

Loads human-authored design brief documents from `.moai/design/` and injects a
token-capped consolidated context block into the orchestrator prompt before
`expert-frontend` or `moai-domain-brand-design` is invoked.

This skill is called automatically during Phase B2.5 of the `/moai design` workflow
when `design_docs.auto_load_on_design_command` is `true`, and can also be invoked
standalone with an explicit `dir` argument.

---

## Quick Reference

Priority order (REQ-2, AC-4): `spec > system > research > pencil-plan`

Bare-token to filename mapping:
- `spec` → `.moai/design/spec.md`
- `system` → `.moai/design/system.md`
- `research` → `.moai/design/research.md`
- `pencil-plan` → `.moai/design/pencil-plan.md`

Token budget default: `20000` (from `design.yaml design_docs.token_budget`)

Token estimation algorithm (REQ-5): `estimated_tokens = ceiling(char_count / 4) * 1.10`

Truncation order when over budget (REVERSE priority): drop `pencil-plan` first, then `research`, then `system`; always preserve `spec`.

---

## Implementation Guide

### Step 1: Resolve Configuration

Read `design_docs` from `.moai/config/sections/design.yaml`. If `design_docs` key
is absent, use compiled-in defaults and log:

```
design_docs not configured — using defaults
```

Defaults:
- `dir`: `.moai/design`
- `auto_load_on_design_command`: `true`
- `token_budget`: `20000`
- `priority`: `[spec, system, research, pencil-plan]`

If invoked standalone with an explicit `dir` argument, override `dir` with the
provided value. All Read tool calls must target paths under `<dir>/`, not the
default `.moai/design/`.

### Step 2: Check Directory Existence

Use Glob to check whether `<dir>/` exists. If the directory does not exist, return
gracefully with an empty context block and log:

```
design docs not initialized — run /moai init or SPEC-DESIGN-DOCS-001 to create
```

Return:

```markdown
## Design Context (from .moai/design/)
```

(header only, no content lines)

### Step 3: Auto-Load Gate (REQ-1, REQ-9)

When invoked from Phase B2.5 (not standalone), check
`design_docs.auto_load_on_design_command`. If `false`, skip and return an empty
context block. Do not auto-invoke; respond only to explicit standalone calls with
`dir`.

### Step 4: Parallel File Discovery and Read (REQ-11, AC-13)

Issue all Read tool calls for candidate files as a **single batched parallel
tool-call set** in the same orchestration turn. The candidates are the bare tokens
from the priority array, mapped to `<dir>/<token>.md`.

Example parallel Read calls (four files, one turn):
- Read `<dir>/spec.md`
- Read `<dir>/system.md`
- Read `<dir>/research.md`
- Read `<dir>/pencil-plan.md`

If a file does not exist, treat as not present (not an error). If a file cannot be
read due to a permission error or corruption, add the token name to the warnings
array and continue with remaining files (partial success semantics, REQ-14, AC-15).

### Step 5: Filter _TBD_ Files (REQ-3, AC-5)

For each successfully read file, check whether the content contains exclusively
`_TBD_` markers (i.e., no user-authored sections — all non-empty lines are `_TBD_`
or YAML/Markdown structural lines produced by scaffolding with no user content).

A file is considered `_TBD_`-only when every meaningful content line matches:
- Blank lines or lines containing only `_TBD_`
- Markdown headings with no body text beyond scaffold headings
- The scaffold comment pattern (lines starting with `<!--` or `>`)

If `_TBD_`-only, skip the file and log:

```
skip: <token> — _TBD_ only
```

### Step 6: Token Budget Enforcement (REQ-4, REQ-5, AC-9)

After filtering, compute the estimated token count for each remaining file:

```
estimated_tokens(file) = ceiling(len(content) / 4) * 1.10
```

Include files in priority order (`spec` first) until the cumulative
`estimated_tokens` would exceed `token_budget`.

When adding the next file would exceed the budget:

1. Drop the **lowest-priority** candidate first (REVERSE priority: `pencil-plan`
   before `research` before `system`; never drop `spec`).
2. Retry with remaining files.
3. If a **single file alone** exceeds the remaining budget after all higher-priority
   files are included, truncate that file at the nearest `##` or `###` section
   boundary before the budget is exceeded, then append a trailing marker:

```
> truncated: <filename> at char_offset=N
```

where `N` is the character offset of the truncation point within the original file.

### Step 7: Build Output Block (REQ-6, REQ-12, AC-6, AC-14)

Construct the consolidated context block. The **first non-empty line must be exactly**:

```
## Design Context (from .moai/design/)
```

For each included file, prepend a citation line:

```
> source: .moai/design/<filename>
```

then append the file content (or truncated content with the trailing marker).

If all candidate files were skipped (`_TBD_`-only) and no content was added, emit
the header and log (REQ-15, AC-16):

```
design docs present but all are _TBD_ — no content loaded
```

### Step 8: Return Result

Return the consolidated context block to the orchestrator for injection into the
next subagent prompt.

If any warnings were collected (unreadable files), include a warnings section after
the context block:

```
> warnings: [<token1> unreadable: <reason>, <token2> unreadable: <reason>]
```

---

## Output Format Contract

```markdown
## Design Context (from .moai/design/)

> source: .moai/design/spec.md
<content of spec.md>

> source: .moai/design/system.md
<content of system.md>

> source: .moai/design/research.md
<content of research.md — or truncated with marker>

> truncated: research.md at char_offset=12345
```

When `.moai/design/` is absent or all files are `_TBD_`:

```markdown
## Design Context (from .moai/design/)
```

---

## Edge Cases

### All Files Are _TBD_ (REQ-15, AC-16)

Directory exists but every candidate is `_TBD_`-only. Return header with no
`> source:` lines and log: `design docs present but all are _TBD_ — no content loaded`

### Missing design_docs Configuration (REQ-16, AC-17)

`design.yaml` has no `design_docs` key. Use compiled-in defaults and log:
`design_docs not configured — using defaults`

### Standalone Invocation with Custom dir (REQ-10, AC-12)

When invoked with `dir=/path/to/alt-design/`, all Read tool calls must target
paths under `/path/to/alt-design/`. Zero reads should target `.moai/design/` in
the project root.

### Partial Read Failure (REQ-14, AC-15)

When one file is unreadable (permission error, corruption), add it to warnings,
continue with remaining files, and include available `> source:` sections in the
output.

---

## Works Well With

- `moai-domain-brand-design`: Receives this skill's context block before brand
  design generation
- `expert-frontend`: Primary downstream consumer of injected context
- `moai-workflow-design-import`: Parallel path A workflow (Claude Design bundle)
- `moai-workflow-gan-loop`: Uses design context as quality evaluation reference

---

REQ coverage: REQ-1 through REQ-16
SPEC: SPEC-DESIGN-ATTACH-001
Version: 1.0.0
