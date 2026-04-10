# SPEC-PENCIL-001: Pencil CLI Integration for .pen File Operations

- **Status**: Draft
- **Priority**: High
- **Created**: 2026-04-10
- **Author**: claude[bot] (analysis), mhsong55 (original research)
- **Issue**: modu-ai/moai-adk#618

---

## 1. Problem Statement

Pencil MCP has three verified structural limitations that prevent reliable multi-file .pen operations:

### 1.1 `filePath` Parameter Ignored in `batch_get` (Verified)

`mcp__pencil__batch_get` returns the active editor's file content regardless of the `filePath` value passed. When `topic-flow.pen` (111KB) is active in the editor, calling `batch_get(filePath: "docs/architecture.pen")` returns `topic-flow.pen` content instead. This makes independent multi-file access impossible via MCP.

### 1.2 `open_document` Fails to Switch Active File (Verified)

Calling `mcp__pencil__open_document` returns a success response ("Document opened") but does not actually change the editor's active file. Subsequent `get_editor_state()` and `batch_get()` calls continue referencing the previously active file.

### 1.3 Headless Environment Incompatibility (Estimated)

Pencil MCP depends on a running GUI editor (Desktop App or VS Code Extension). Headless Linux servers (CI/CD, remote development) cannot use Pencil MCP. This limitation is estimated but not yet tested on actual headless infrastructure.

### 1.4 Current Codebase Impact

The following files enforce MCP-only usage:

| File | Impact |
|------|--------|
| `expert-frontend.md:72` | `[HARD] Use Pencil MCP for all UI/UX design tasks` |
| `moai-design-tools/SKILL.md:89` | Lists "14 + 1 CLI-only" tools (inaccurate count) |
| `moai-design-tools/SKILL.md:109` | `export_nodes` incorrectly marked as "CLI only" |
| `moai-design-tools/SKILL.md:105-106` | Lists `get_style_guide` and `get_style_guide_tags` as separate tools (do not exist as separate MCP tools) |
| `moai-design-tools/reference/pencil-renderer.md:11` | "ALWAYS use Pencil MCP tools" |
| `CLAUDE.md:Section 12` | Lists Pencil as MCP-only |
| `settings-management.md` | Lists Pencil as MCP server |
| `designer.md:46` | References "Pencil MCP tools when available" |

### 1.5 Documentation Errors Found

| Item | Current Documentation | Actual Behavior |
|------|----------------------|-----------------|
| `get_style_guide` | Listed as separate tool | Does not exist as standalone tool; merged into `get_guidelines({ category: "style" })` |
| `get_style_guide_tags` | Listed as separate tool | Does not exist as standalone tool; output included in `get_guidelines()` |
| `export_nodes` | Marked "CLI-only" | Also available via MCP (`mcp__pencil__export_nodes` exists) |
| Tool count | "14 tools + 1 CLI-only" | Actually 18 MCP tools (verified by issue author) |

---

## 2. Verified Solution: Pencil CLI

Pencil CLI (`@pencil.dev/cli v0.2.4`) has been individually verified to reproduce all 18 MCP tool operations. The CLI operates in two modes:

- **Headless mode**: `pencil interactive --in file.pen --out output.pen` (no GUI required)
- **App mode**: `pencil interactive --app desktop` (real-time sync with Desktop editor)

### 2.1 CLI Tool Compatibility (18/18 Verified)

| # | MCP Tool | CLI Equivalent | Mode |
|---|----------|---------------|------|
| 1 | `get_editor_state` | `get_editor_state()` via interactive | headless + app |
| 2 | `open_document` | Replaced by `--in` flag | headless + app |
| 3 | `get_guidelines` | `get_guidelines()` via interactive | headless + app |
| 4 | `batch_get` | `batch_get()` via interactive | headless + app |
| 5 | `batch_design` (Insert) | `batch_design()` via interactive | headless + app |
| 6 | `batch_design` (Update) | `batch_design()` via interactive | headless + app |
| 7 | `batch_design` (Copy) | `batch_design()` via interactive | headless + app |
| 8 | `batch_design` (Replace) | `batch_design()` via interactive | headless + app |
| 9 | `batch_design` (Move) | `batch_design()` via interactive | headless + app |
| 10 | `batch_design` (Delete) | `batch_design()` via interactive | headless + app |
| 11 | `snapshot_layout` | `snapshot_layout()` via interactive | headless + app |
| 12 | `get_screenshot` | `get_screenshot()` via interactive | headless + app |
| 13 | `find_empty_space_on_canvas` | `find_empty_space_on_canvas()` via interactive | headless + app |
| 14 | `get_variables` | `get_variables()` via interactive | headless + app |
| 15 | `set_variables` | `set_variables()` via interactive | headless + app |
| 16 | `search_all_unique_properties` | `search_all_unique_properties()` via interactive | headless + app |
| 17 | `replace_all_matching_properties` | `replace_all_matching_properties()` via interactive | headless + app |
| 18 | `export_nodes` | `export_nodes()` via interactive | headless + app |

### 2.2 Capability Comparison

| Feature | Pencil MCP | CLI Headless | CLI `--app` |
|---------|-----------|-------------|-------------|
| Tool compatibility | 18/18 | 18/18 | 18/18 |
| Multi-file independent access | No | **Yes** | **Yes** |
| GUI-free environment | No (estimated) | **Yes** | No |
| Real-time editor sync | **Yes** | No | **Yes** |
| Claude Code native integration | **Yes** (direct MCP tool call) | No (Bash wrapping) | No (Bash wrapping) |
| Install/Auth | Automatic (bundled) | Separate npm + login | Separate npm + login |
| Execution speed | Fast (persistent conn) | Slow (process per call) | Medium |

---

## 3. Recommended Solution: Option B (Parallel Usage)

### 3.1 Rationale

After evaluating both options against the codebase impact and user workflows:

**Option A (Full Replacement) is NOT recommended** because:
- MCP provides native Claude Code integration with type-safe parameter validation and autocomplete
- MCP is faster for single-file operations (persistent connection vs process start/stop)
- Removing MCP tools from `expert-frontend.md` would degrade the single-file editing experience
- The `designer.md` agency agent references Pencil MCP for real-time design sync

**Option B (Parallel Usage) IS recommended** because:
- Preserves MCP advantages for single-file, GUI-connected workflows
- Adds CLI as a fallback for multi-file and headless scenarios
- Gradual adoption path — no breaking changes to existing workflows
- Agent can auto-select the best tool based on context
- Aligns with MoAI's MCP Fallback Strategy (agent-common-protocol.md)

### 3.2 Tool Selection Decision Tree

```
.pen file operation requested
  |
  +-- Is Pencil MCP available? (Desktop/VS Code running)
  |   |
  |   +-- YES: Single file operation?
  |   |   |
  |   |   +-- YES -> Use Pencil MCP (fastest, native integration)
  |   |   +-- NO (multiple files) -> Use Pencil CLI --app (multi-file + editor sync)
  |   |
  |   +-- NO: MCP not available
  |       +-- Use Pencil CLI headless (--in/--out)
  |
  +-- Is Pencil CLI installed?
      |
      +-- YES -> Proceed with CLI
      +-- NO -> Guide user to install: npm install -g @pencil.dev/cli && pencil login
```

---

## 4. Migration Plan

### Phase 1: Documentation Fixes (Priority: High)

Fix verified documentation errors in template files (single source of truth):

**File: `internal/template/templates/.claude/skills/moai-design-tools/SKILL.md`**
1. Remove `get_style_guide` and `get_style_guide_tags` from the tool list (they do not exist as separate MCP tools)
2. Remove "CLI only" annotation from `export_nodes` (it is available via MCP)
3. Update tool count from "14 + 1 CLI-only" to "18 tools"
4. Add CLI workflow section with tool selection guide

**File: `internal/template/templates/.claude/skills/moai-design-tools/reference/pencil-renderer.md`**
1. Remove references to non-existent `get_style_guide()` and `get_style_guide_tags()` tools
2. Expand CLI section to equal priority with MCP
3. Add multi-file workflow examples using CLI
4. Add headless environment usage guide

### Phase 2: Agent Rule Updates (Priority: High)

**File: `internal/template/templates/.claude/agents/moai/expert-frontend.md`**
1. Line 12: Remove `mcp__pencil__get_style_guide` and `mcp__pencil__get_style_guide_tags` from tools list (non-existent tools)
2. Line 72: Change `[HARD] Use Pencil MCP for all UI/UX design tasks` to `[HARD] Use Pencil MCP or Pencil CLI for all UI/UX design tasks`
3. Lines 74-78: Update the Pencil Design Workflow to include CLI alternative paths
4. Add Bash to tools list if not already present (needed for CLI wrapping)

**File: `internal/template/templates/.claude/agents/agency/designer.md`**
1. Line 46: Update "Pencil MCP tools when available" to "Pencil MCP or CLI tools when available"

### Phase 3: Skill Enhancement (Priority: Medium)

**File: `internal/template/templates/.claude/skills/moai-design-tools/SKILL.md`**
1. Add "Pencil CLI Workflow" section parallel to "Pencil MCP Workflow"
2. Add tool selection decision tree
3. Add CLI prerequisite check documentation
4. Update comparison.md to include CLI as third mode

### Phase 4: Infrastructure (Priority: Medium)

1. Add CLI installation check to session-start hook (optional, advisory only)
2. Add `pencil` to recommended tools in project setup documentation
3. Consider adding a `moai-tool-pencil-cli` skill for CLI-specific patterns

### Phase 5: Verification (Priority: Low)

1. Test headless environment on actual Linux server (estimated behavior, not yet verified)
2. Confirm with Pencil team whether `filePath` behavior in `batch_get` is a bug or by design
3. Confirm whether `open_document` state-switching failure will be fixed

---

## 5. CLI Command Mappings

### Bash Wrapper Pattern

All CLI operations follow this pattern:

```bash
# Single tool call
echo 'tool_name({ param1: "value1", param2: "value2" })
exit()' | pencil interactive --in input.pen --out output.pen

# Multiple tool calls in one session
echo 'batch_get()
get_variables()
get_screenshot({ nodeId: "rootId" })
exit()' | pencil interactive --in input.pen --out output.pen

# App mode (real-time editor sync)
echo 'batch_design({ operations: "U(\"nodeId\", { content: \"updated\" })" })
exit()' | pencil interactive --app desktop
```

### Per-Tool CLI Mapping

| MCP Tool Call | CLI Equivalent |
|--------------|----------------|
| `mcp__pencil__get_editor_state()` | `echo 'get_editor_state()\nexit()' \| pencil interactive --in file.pen --out /tmp/out.pen` |
| `mcp__pencil__open_document({ filePathOrNew: "path" })` | `pencil interactive --in path --out output.pen` (file specified at launch) |
| `mcp__pencil__batch_get({ patterns: [...] })` | `echo 'batch_get({ patterns: [...] })\nexit()' \| pencil interactive --in file.pen --out /tmp/out.pen` |
| `mcp__pencil__batch_design({ operations: "..." })` | `echo 'batch_design({ operations: "..." })\nexit()' \| pencil interactive --in file.pen --out output.pen` |
| `mcp__pencil__get_screenshot({ nodeId: "id" })` | `echo 'get_screenshot({ nodeId: "id" })\nexit()' \| pencil interactive --in file.pen --out /tmp/out.pen` |
| `mcp__pencil__snapshot_layout()` | `echo 'snapshot_layout()\nexit()' \| pencil interactive --in file.pen --out /tmp/out.pen` |
| `mcp__pencil__get_variables()` | `echo 'get_variables()\nexit()' \| pencil interactive --in file.pen --out /tmp/out.pen` |
| `mcp__pencil__set_variables({ variables: {...} })` | `echo 'set_variables({ variables: {...} })\nexit()' \| pencil interactive --in file.pen --out output.pen` |
| `mcp__pencil__get_guidelines({ topic: "code" })` | `echo 'get_guidelines({ topic: "code" })\nexit()' \| pencil interactive --in file.pen --out /tmp/out.pen` |
| `mcp__pencil__find_empty_space_on_canvas(...)` | `echo 'find_empty_space_on_canvas(...)\nexit()' \| pencil interactive --in file.pen --out /tmp/out.pen` |
| `mcp__pencil__search_all_unique_properties(...)` | `echo 'search_all_unique_properties({ parents: [...] })\nexit()' \| pencil interactive --in file.pen --out /tmp/out.pen` |
| `mcp__pencil__replace_all_matching_properties(...)` | `echo 'replace_all_matching_properties({ match: {...}, replace: {...}, parents: [...] })\nexit()' \| pencil interactive --in file.pen --out output.pen` |
| `mcp__pencil__export_nodes(...)` | `echo 'export_nodes(...)\nexit()' \| pencil interactive --in file.pen --out /tmp/out.pen` |

---

## 6. Updated Expert Rules for expert-frontend.md

### Current Rule (Line 72)

```
[HARD] Use Pencil MCP for all UI/UX design tasks.
```

### Proposed Rule

```
[HARD] Use Pencil MCP or Pencil CLI for all UI/UX design tasks.

Tool Selection:
- Single file + GUI editor available: Pencil MCP (native integration, fastest)
- Multiple files: Pencil CLI --app mode (multi-file independent access)
- Headless environment: Pencil CLI headless mode (--in/--out)
- MCP unavailable: Pencil CLI as fallback

Prerequisites for CLI: @pencil.dev/cli installed, pencil login completed.
```

### Current Workflow (Lines 74-78)

```
1. Initialize: get_editor_state -> open_document -> get_guidelines
2. Style Foundation: get_style_guide_tags -> get_style_guide -> set_variables
3. Design: batch_design -> snapshot_layout -> get_screenshot
4. Iterate: batch_get -> batch_design -> get_screenshot
5. Export: AI prompt to generate code
```

### Proposed Workflow

```
MCP Workflow (single file, GUI available):
1. Initialize: get_editor_state -> open_document -> get_guidelines
2. Style Foundation: get_guidelines({ category: "style" }) -> set_variables
3. Design: batch_design -> snapshot_layout -> get_screenshot
4. Iterate: batch_get -> batch_design -> get_screenshot
5. Export: AI prompt to generate code

CLI Workflow (multi-file or headless):
1. Initialize: pencil interactive --in file.pen --out output.pen
2. Style Foundation: get_guidelines({ category: "style" }) -> set_variables
3. Design: batch_design -> snapshot_layout -> get_screenshot
4. Iterate: batch_get -> batch_design -> get_screenshot
5. Save: save() -> exit()
```

### Tools List Update

Remove non-existent tools from the frontmatter tools list:
- Remove: `mcp__pencil__get_style_guide`, `mcp__pencil__get_style_guide_tags`
- Keep: All other `mcp__pencil__*` tools
- Add: `mcp__pencil__export_nodes` (was incorrectly excluded)

---

## 7. Rollback Strategy

### Immediate Rollback (Phase 1-2)

If CLI integration causes issues:
1. Revert `expert-frontend.md` tool selection rule to MCP-only
2. Revert `moai-design-tools/SKILL.md` to remove CLI workflow section
3. Documentation error fixes (Phase 1) should NOT be reverted as they correct factual inaccuracies

### Gradual Rollback (Phase 3-4)

1. Remove CLI workflow sections from skills
2. Remove CLI prerequisite checks from hooks
3. Keep MCP as sole tool path

### Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| CLI version incompatibility | Low | Medium | Pin to `@pencil.dev/cli ^0.2.4` |
| CLI auth token expiry | Medium | Low | Add `pencil status` check before operations |
| CLI process overhead | Medium | Low | Recommend `--app` mode for iterative work |
| MCP `filePath` bug gets fixed | Medium | Low | CLI integration remains valuable for headless |
| Agent selects wrong tool | Low | Low | Decision tree is deterministic |

---

## 8. Open Questions for Pencil Team

> **@pencil team**: We need confirmation on two behaviors observed in Pencil MCP:
>
> 1. **`batch_get` filePath parameter**: Is ignoring the `filePath` parameter (always returning active editor content) intentional behavior or a bug? If a bug, is there a fix timeline?
>
> 2. **`open_document` state switching**: Is the failure to actually switch the active editor file a known limitation, async timing issue, or a bug?
>
> These answers will determine whether MCP can eventually support multi-file workflows or whether CLI remains the permanent solution for that use case.

---

## 9. Acceptance Criteria

- [ ] AC-1: Documentation errors fixed (non-existent tools removed, correct tool count)
- [ ] AC-2: `expert-frontend.md` allows both MCP and CLI paths
- [ ] AC-3: `moai-design-tools/SKILL.md` includes CLI workflow section
- [ ] AC-4: Tool selection decision tree documented and deterministic
- [ ] AC-5: CLI command mappings for all 18 tools documented
- [ ] AC-6: Rollback procedure documented and tested
- [ ] AC-7: `designer.md` updated to reference CLI alongside MCP

---

## 10. Files to Modify (Summary)

All modifications follow the Template-First Rule (CLAUDE.local.md Section 2):

| Template File (Source of Truth) | Change |
|--------------------------------|--------|
| `internal/template/templates/.claude/agents/moai/expert-frontend.md` | Relax [HARD] rule, fix tools list, update workflow |
| `internal/template/templates/.claude/skills/moai-design-tools/SKILL.md` | Fix tool count/names, add CLI section |
| `internal/template/templates/.claude/skills/moai-design-tools/reference/pencil-renderer.md` | Fix tool references, expand CLI section |
| `internal/template/templates/.claude/agents/agency/designer.md` | Update Pencil reference |
| `internal/template/templates/.claude/skills/moai-design-tools/reference/comparison.md` | Add CLI mode to comparison |

Local copies (`.claude/`) must be synced after template changes via `make build && moai update`.

---

Version: 1.0.0 (Draft)
Source: modu-ai/moai-adk#618
