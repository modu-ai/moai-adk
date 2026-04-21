# Figma MCP Integration Guide

Figma MCP integration for fetching design context, creating designs, and managing design systems through the official Figma MCP remote server.

## Overview

Figma MCP provides direct access to Figma file data and design generation capabilities through the Model Context Protocol. The official Figma MCP server enables design context extraction, variable management, FigJam collaboration, and AI-powered design creation (Code-to-Canvas).

## Setup

### Installation

Install the official Figma MCP plugin:

```
claude plugin install figma@claude-plugins-official
```

No manual `mcpServers` configuration is required. The plugin handles connection automatically.

### Remote MCP Server

The official Figma MCP server is hosted at:

```
https://mcp.figma.com/mcp
```

### Authentication

1. Authenticate via the Figma MCP plugin prompt when first connecting
2. Your Figma account credentials are used for authorization (OAuth)
3. Access is scoped to files and projects your account can view

### Server Options

Two deployment modes are available:

**Remote MCP server (recommended):**
- Hosted at https://mcp.figma.com/mcp
- No Figma desktop app required
- Broadest feature set including write-to-canvas and Code-to-Canvas
- Recommended for most users

**Desktop MCP server:**
- Runs locally through the Figma desktop app
- Primarily for organizations and enterprises with specific requirements
- More limited feature set compared to remote server

## Figma MCP Tools Reference

All Figma MCP tools are exposed in Claude Code as `mcp__plugin_figma_figma__<name>` (sections below use the short names). Many tools accept optional `clientFrameworks` and `clientLanguages` string parameters used for telemetry only — pass `unknown` when uncertain. Node IDs are accepted as either `123:456` or `123-456` and are extractable from standard Figma URLs:

- `figma.com/design/:fileKey/:fileName?node-id=1-2` → `nodeId = "1:2"`, `fileKey = ":fileKey"`
- `figma.com/design/:fileKey/branch/:branchKey/:fileName` → pass `branchKey` as `fileKey`
- `figma.com/make/:makeFileKey/:makeFileName` → pass `makeFileKey` as `fileKey`

### Design Context and Reading

#### get_design_context

Primary tool for design-to-code workflows. Returns reference code, a screenshot, and contextual metadata for the node. The returned code is a reference to adapt to the target project — NOT final output:
- Required: `fileKey`, `nodeId`
- Optional: `excludeScreenshot` (screenshots strongly recommended), `forceCode` (always return code even if large), `disableCodeConnect`, `clientFrameworks`, `clientLanguages`
- Returns a code string plus a JSON of asset download URLs and (by default) a screenshot

```
get_design_context(fileKey, nodeId, { excludeScreenshot?, forceCode?, disableCodeConnect? }?) → { code, assets: { [assetUrl]: string }, screenshot? }
```

#### get_screenshot

Render a screenshot image of any Figma node (frame, component, page, etc.) or the currently selected node in the desktop app:
- Required: `fileKey`, `nodeId`
- No optional params
- Same URL parsing rules as `get_design_context`

```
get_screenshot(fileKey, nodeId) → image data
```

#### get_variable_defs

Resolve design-token variables referenced under a specific node's subtree (per-node, not per-file):
- Required: `fileKey`, `nodeId` — variables are returned for the subtree rooted at the node
- Optional: `clientFrameworks`, `clientLanguages`
- Returns a flat map `{ variableName: value }`, e.g. `{ "icon/default/secondary": "#949494" }`
- Covers color, typography, size, and spacing tokens referenced by the node

```
get_variable_defs(fileKey, nodeId) → { [variableName: string]: value }
```

#### get_metadata

Get XML-format structural metadata for a node or page. Prefer `get_design_context` for substantive design-to-code work:
- Required: `fileKey`, `nodeId` (page IDs such as `0:1` are accepted)
- Optional: `clientFrameworks`, `clientLanguages`
- Returns XML with node IDs, layer types, names, positions, and sizes only — no file-level metadata (name, description, timestamps, etc.)
- IMPORTANT: never call `get_metadata` on Figma Make files

```
get_metadata(fileKey, nodeId) → XML
```

#### get_libraries

Get the design libraries associated with a Figma file. Returns two lists:
- Subscribed libraries — libraries currently added to the file
- Available libraries — community UI kits and organization libraries available to add
- Each entry includes `name`, library `key`, `description`, and `sourceType`
- Use the returned library keys to scope `search_design_system` via its `includeLibraryKeys` parameter

```
get_libraries(fileKey) → { subscribed: [{ name, key, description, sourceType }, ...], available: [...] }
```

#### whoami

Get authenticated-user information and available plans:
- No parameters
- Returns the current user identity plus available plans; each plan's `key` is the `planKey` required by `create_new_file`
- Use for debugging authentication issues with other tools

```
whoami() → { user: { ... }, plans: [{ key, name, ... }] }
```

### Code Connect

#### get_code_connect_map

Retrieve existing Figma-node → codebase-component Code Connect mappings:
- Required: `fileKey`, `nodeId`
- Optional: `codeConnectLabel` — disambiguator when multiple mappings exist for the same node across languages/frameworks
- Returns a map keyed by `nodeId` with `{ codeConnectSrc, codeConnectName }` (source file path and exported component name)

```
get_code_connect_map(fileKey, nodeId, codeConnectLabel?) → { [nodeId]: { codeConnectSrc, codeConnectName } }
```

#### add_code_connect_map

Create a Code Connect mapping for a single Figma node → code component:
- Required: `fileKey`, `nodeId`, `source` (file path in codebase), `componentName`, `label`
- `label` enum (16 values): `React`, `Web Components`, `Vue`, `Svelte`, `Storybook`, `Javascript`, `Swift`, `Swift UIKit`, `Objective-C UIKit`, `SwiftUI`, `Compose`, `Java`, `Kotlin`, `Android XML Layout`, `Flutter`, `Markdown`
- Optional: `template` (executable JS template — promotes the record to `figmadoc`-type instead of a simple `component_browser` mapping), `templateDataJson` (metadata keys: `isParserless`, `imports`, `nestable`, `props`), `clientFrameworks`, `clientLanguages`
- For bulk mapping across many nodes, use `send_code_connect_mappings` instead

```
add_code_connect_map(fileKey, nodeId, source, componentName, label, { template?, templateDataJson? }?) → confirmation
```

#### get_code_connect_suggestions

Get AI-suggested Code Connect mapping candidates for a node:
- Required: `fileKey`, `nodeId`
- Optional: `excludeMappingPrompt` (return only a lightweight list of unmapped components), `clientFrameworks`, `clientLanguages`
- Workflow: call this → review suggestions with the user → persist via `send_code_connect_mappings`

```
get_code_connect_suggestions(fileKey, nodeId, { excludeMappingPrompt? }?) → { suggestions: [...] }
```

#### send_code_connect_mappings

Persist multiple Code Connect mappings in bulk after user approval:
- Required: `fileKey`, `nodeId`, `mappings` (array)
- Each mapping item: `{ nodeId, componentName, source, label, template?, templateDataJson? }` (same semantics as `add_code_connect_map`)
- Follow-up to `get_code_connect_suggestions`

```
send_code_connect_mappings(fileKey, nodeId, mappings) → confirmation
```

#### get_context_for_code_connect

Get structured Figma component metadata designed for authoring Code Connect template files (.figma.ts / .figma.js):
- Returns property definitions (with types and variant options) for the target component or component set
- Returns a descendant tree of instances and text nodes, each annotated with property references
- Required: `fileKey`, `nodeId`
- Optional: `clientFrameworks`, `clientLanguages` (telemetry only — pass `unknown` when uncertain)

```
get_context_for_code_connect(fileKey, nodeId, { clientFrameworks?, clientLanguages? }?) → { properties, variants, descendantTree }
```

### FigJam and Diagrams

#### get_figjam

Generate UI code for a FigJam node — FigJam-only, not a generic board reader:
- Required: `fileKey`, `nodeId` (use `0:1` for the root node)
- Optional: `includeImagesOfNodes` (default `true`)
- IMPORTANT: works only on FigJam files (`figma.com/board/...`), not on regular Figma design files

```
get_figjam(fileKey, nodeId, { includeImagesOfNodes? }?) → UI code
```

#### generate_diagram

Create a Mermaid.js diagram as a new FigJam file. Creates its own file — do NOT call `create_new_file` beforehand:
- Required: `name` (short human-readable title), `mermaidSyntax` (Mermaid.js code)
- Optional: `userIntent` (short description — telemetry only)
- Supported diagram types: `graph`, `flowchart`, `sequenceDiagram`, `stateDiagram`, `stateDiagram-v2`, `gantt`
- Not supported: class diagrams, timelines, Venn diagrams, ER diagrams, font changes, moving individual shapes
- IMPORTANT: after calling, you MUST surface the returned URL to the user as a markdown link

```
generate_diagram(name, mermaidSyntax, userIntent?) → { url, ... }
```

### Design Generation (Code-to-Canvas)

#### generate_figma_design

Multi-step Code-to-Canvas tool — captures, imports, or converts a web page or HTML into a Figma design. Works with localhost and external URLs (Remote MCP only):
- No required params; call first with no `outputMode` to receive capture instructions and output-mode options
- Follow-up params (once selected): `outputMode` (enum: `newFile` | `existingFile` | `clipboard`), `fileKey` (for `existingFile`), `fileName` + `planKey` (for `newFile`), `nodeId` (target inside `existingFile`), `captureId` (for polling)
- Workflow: initial call → choose `outputMode` → capture → poll with `captureId` every 5s (max 10 polls) until `status === "completed"`
- For LOCAL projects: identify the dev-server URL from the codebase first
- For EXTERNAL URLs: capture via Playwright MCP — do NOT use `open` with hash fragments
- For web apps, pair with `use_figma` + `search_design_system` to build the screen from design-system components, then delete this tool's capture (used only as a pixel-perfect layout reference)
- Each capture ID is single-use

```
generate_figma_design({ captureId?, fileKey?, fileName?, nodeId?, outputMode?, planKey? }?) → { captureId | fileKey | ... }
```

**Known Limitations:**
- Japanese text rendering may have issues
- Image dimensions may not exactly match specifications
- Available via Remote MCP server only (https://mcp.figma.com/mcp)

### Write and Create Tools

#### use_figma

Canonical Figma Plugin API executor — the primary tool for all Figma writes. Runs JavaScript in the Figma file context:
- Required: `fileKey`, `code` (JavaScript, max 50,000 chars), `description` (≤ 2000 chars summary of intent)
- Optional: `skillNames` (e.g. `"figma-use"` or `"figma-use,figma-generate-design"` — telemetry)
- MANDATORY PREREQUISITE: load the `figma-use` skill before calling (skipping causes hard-to-debug failures)
- Capabilities: create/edit/delete pages, frames, components, variants, variables, styles, text, images; set up design tokens; build variant systems; inspect node properties; fix layout/auto-layout issues
- Gotchas:
  - Inter font uses style `"Semi Bold"` with a space — not `"SemiBold"`; same for `"Extra Bold"`
  - Do NOT assign `figma.currentPage = page`; use `await figma.setCurrentPageAsync(page)`
- Before creating components, call `search_design_system` first and import matches via `importComponentByKeyAsync` / `importComponentSetByKeyAsync`

```
use_figma(fileKey, code, description, { skillNames? }?) → execution result
```

#### search_design_system

Search design libraries for matching components, variables, and styles:
- Required: `query`, `fileKey`
- Optional: `includeLibraryKeys` (scope to library keys from `get_libraries`), `includeComponents` (default `true`), `includeStyles` (default `true`), `includeVariables` (default `true`), `disableCodeConnect`
- Returns matching assets across all connected design libraries

```
search_design_system(query, fileKey, { includeLibraryKeys?, includeComponents?, includeStyles?, includeVariables?, disableCodeConnect? }?) → { components, styles, variables }
```

#### create_new_file

Create a new blank Figma Design or FigJam file in the authenticated user's drafts folder:
- Required: `fileName`, `planKey`, `editorType` (enum: `"design"` | `"figjam"`)
- `planKey` is obtained from `whoami()` — it is the `key` field on each plan entry. If the user has multiple plans, ask which team/organization to use
- Returns the new file key and URL

```
create_new_file(fileName, planKey, editorType) → { fileKey, url }
```

### Design System

#### create_design_system_rules

Return a prompt used by the agent to scaffold design-system rules for the current repository. This does NOT create rules inside Figma:
- No required params
- Optional: `clientFrameworks`, `clientLanguages` (telemetry)
- Output is a prompt string the agent feeds back into itself; pair with the `figma-create-design-system-rules` skill for the end-to-end flow

```
create_design_system_rules({ clientFrameworks?, clientLanguages? }?) → prompt
```

## Rate Limits

**Starter Plan / View or Collab seats on paid plans:**
- Limited to 6 tool calls per month

**Dev or Full seats (Professional/Organization/Enterprise plans):**
- Per-minute rate limits matching Figma REST API Tier 1

**Write-to-Canvas:**
- Currently free during beta period
- Will become a usage-based paid feature

## implement-design Workflow

When implementing a Figma design as code, follow this 7-step workflow:

### Step 1: Get Design Context

```
get_design_context(fileKey, nodeId)
```

Understand the component structure, layout relationships, and design intent before writing any code.

### Step 2: Get Visual Reference

```
get_screenshot(fileKey, nodeId)
```

Capture a screenshot to use as visual reference throughout implementation. Compare final code output against this image.

### Step 3: Extract Design Tokens

```
get_variable_defs(fileKey, nodeId)
```

Extract design variables referenced under the target node (colors, typography, spacing tokens). Use these values instead of hardcoding them. Variables are returned as a flat `{ variableName: value }` map.

### Step 4: Analyze Component Structure

Review the design context to identify:
- Component hierarchy and nesting
- Responsive behavior and breakpoints
- Interactive states (hover, active, disabled)
- Reusable sub-components

### Step 5: Generate React/Tailwind Code

Implement the component using extracted design context:
- Map Figma components to React components
- Apply design token values from get_variable_defs
- Use Tailwind classes for styling
- Handle responsive layouts with Tailwind breakpoints

### Step 6: Apply Design Tokens

Map extracted variables to Tailwind config:

```typescript
// tailwind.config.js — populated from get_variable_defs output
module.exports = {
  theme: {
    extend: {
      colors: {
        primary: {
          500: '#3B82F6',  /* Figma: colors.primary */
        }
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
      }
    }
  }
}
```

### Step 7: Verify Against Screenshot

Compare the implemented component against the screenshot from Step 2:
- Check visual accuracy (colors, spacing, typography)
- Validate responsive behavior at all breakpoints
- Verify interactive states

## Design Token Extraction

### Extracting Color Variables

```
vars = get_variable_defs(fileKey, nodeId)
// Returns a flat map, e.g.:
// { "color/primary": "#3B82F6", "color/primary-hover": "#2563EB", ... }
```

Map to CSS custom properties:
```css
:root {
  --color-primary: #3B82F6;
  --color-primary-hover: #2563EB;
}
```

### Extracting Typography Variables

```
vars = get_variable_defs(fileKey, nodeId)
// Returns a flat map with typography entries, e.g.:
// { "typography/heading-1": { fontFamily: "Inter", fontSize: 32, fontWeight: 700 } }
```

Map to Tailwind:
```typescript
module.exports = {
  theme: {
    extend: {
      fontSize: {
        'heading-1': ['2rem', { lineHeight: '1.2', fontWeight: '700' }],
        'heading-2': ['1.5rem', { lineHeight: '1.3', fontWeight: '600' }],
      }
    }
  }
}
```

## Design-to-Code Workflow Examples

### Example 1: Component Implementation

```
1. get_metadata(fileKey, pageNodeId) → Inspect page structure, identify target node
2. get_design_context(fileKey, nodeId) → Get reference code, screenshot, and contextual metadata
3. get_screenshot(fileKey, nodeId) → Capture visual reference (or reuse screenshot from get_design_context)
4. get_variable_defs(fileKey, nodeId) → Extract design tokens referenced under the node
5. Implement the component, adapting the reference code to your stack
6. Compare implementation screenshot with design screenshot
```

### Example 2: Design System Setup

```
1. get_libraries(fileKey) → Discover connected libraries (subscribed + available)
2. get_variable_defs(fileKey, rootNodeId) → Extract all design tokens under the root
3. Generate tailwind.config.js from the extracted token map
4. get_code_connect_suggestions(fileKey, componentNodeId) → Propose Figma → code component mappings
5. send_code_connect_mappings(fileKey, rootNodeId, approvedMappings) → Persist confirmed mappings in bulk
```

### Example 3: FigJam Workflow Import

```
1. get_figjam(fileKey, "0:1") → Generate UI code for the root of a FigJam file
2. Parse user journey from the returned UI code
3. Implement screens following the user flow
4. generate_diagram(name, mermaidSyntax) → Create updated architecture diagrams as new FigJam files
```

## Best Practices

### Design Context First

- Always call get_design_context before implementing any Figma design
- Use get_screenshot as a persistent visual reference checkpoint
- Extract all variables with get_variable_defs — never hardcode design values

### Token Naming Conventions

Use semantic naming when mapping Figma variables to code:
- `color.primary.500` instead of `blue`
- `spacing.md` instead of `16px`
- `font.heading.1` instead of `32px bold`

### Code Connect Usage

Register code-component mappings for full traceability:
- Use add_code_connect_map when implementing new components
- Use get_code_connect_map to discover if components are already mapped

### Design Verification

- Always verify implementation against get_screenshot output
- Use create_design_system_rules to document component usage patterns

## Error Handling

### Authentication Issues

```
Error: Not authenticated to Figma
Solution: Re-run: claude plugin install figma@claude-plugins-official
```

### File Not Found

```
Error: File not found or access denied
Solution: Verify fileKey is correct and your Figma account has file access
```

### Node Not Found

```
Error: Node ID not found in file
Solution: Use get_metadata to discover valid node IDs and page structure
```

### Code-to-Canvas Unavailable

```
Error: generate_figma_design not available
Solution: generate_figma_design requires the Remote MCP server (https://mcp.figma.com/mcp)
```

## Resources

- Figma MCP Remote Server: https://mcp.figma.com/mcp
- Figma MCP Developer Docs: https://developers.figma.com/docs/figma-mcp-server/
- Figma MCP Tools & Prompts: https://developers.figma.com/docs/figma-mcp-server/tools-and-prompts/
- Figma MCP GitHub Guide: https://github.com/figma/mcp-server-guide
- Figma Developers: https://www.figma.com/developers
- Design Tokens Format: https://designtokens.org/format/

---

Last Updated: 2026-04-21
Tool Version: Figma MCP (Official Remote Server, 18 tools — plugin v2.1.7, verified against live tool schemas)
