---
name: moai-foundation-uiux
description: Moai Design Systems - Professional implementation guide
version: 1.0.0
modularized: false
tags:
 - enterprise
 - systems
 - development
updated: 2025-11-24
status: active
---

## Quick Reference (30 seconds)

# Design Systems Development Skill

Enterprise design system implementation using W3C DTCG 2025.10 token standards, WCAG 2.2 accessibility compliance, and Figma MCP automation workflows.

Core Capabilities:
- W3C DTCG 2025.10 design token architecture
- WCAG 2.2 AA/AAA accessibility compliance
- Figma MCP integration for design-to-code workflows
- Atomic Design component patterns
- Storybook documentation and visual regression testing

When to Use:
- Setting up design token architecture for multi-platform projects
- Implementing accessible component libraries
- Automating design-to-code workflows with Figma MCP
- Building maintainable design systems with Storybook
- Ensuring color contrast compliance and semantic token naming

Module Organization:
- Core Concepts: This file (design tokens, DTCG standards, tool ecosystem)
- Components: [Component Architecture](component-architecture.md) (Atomic Design, component patterns, props APIs)
- Accessibility: [Accessibility WCAG](accessibility-wcag.md) (WCAG 2.2, testing, keyboard navigation)

Latest Standards (November 2025):
- DTCG Specification: 2025.10 (first stable version)
- WCAG Guidelines: 2.2 (AA: 4.5:1 text, AAA: 7:1 text)
- Figma MCP: Desktop + Remote server support
- Style Dictionary: 4.0 (DTCG-compatible)
- Storybook: 8.x (with Docs addon)

---

## Implementation Guide

### Design System Foundation - Three Pillars

1. Design Tokens (Single Source of Truth):
- Color, typography, spacing, borders, shadows
- Semantic naming: `color.primary.500`, `spacing.md`, `font.heading.lg`
- Multi-theme support (light/dark modes)
- Format: W3C DTCG 2025.10 JSON or Style Dictionary 4.0

2. Component Library (Atomic Design Pattern):
- Atoms → Molecules → Organisms → Templates → Pages
- Props API for reusability and composition
- Variant states: default, hover, active, disabled, error, loading
- Documentation: Storybook with auto-generated props/usage

3. Accessibility Standards (WCAG 2.2 Compliance):
- Color contrast: 4.5:1 (AA), 7:1 (AAA) for text
- Keyboard navigation: Tab order, focus management
- Screen readers: ARIA roles, labels, live regions
- Motion: `prefers-reduced-motion` support

### Tool Ecosystem

| Tool | Version | Purpose | Official Link |
|------|---------|---------|---------------|
| W3C DTCG | 2025.10 | Design token specification | https://designtokens.org |
| Style Dictionary | 4.0+ | Token transformation engine | https://styledictionary.com |
| Figma MCP | Latest | Design-to-code automation | https://help.figma.com/hc/en-us/articles/32132100833559 |
| Storybook | 8.x | Component documentation | https://storybook.js.org |
| axe DevTools | Latest | Accessibility testing | https://www.deque.com/axe/devtools/ |
| Chromatic | Latest | Visual regression testing | https://chromatic.com |

---

## Design Token Architecture (DTCG 2025.10)

### Token Structure - Semantic Naming Convention

```json
{
 "$schema": "https://tr.designtokens.org/format/",
 "$tokens": {
 "color": {
 "$type": "color",
 "primary": {
 "50": { "$value": "#eff6ff" },
 "100": { "$value": "#dbeafe" },
 "500": { "$value": "#3b82f6" },
 "900": { "$value": "#1e3a8a" }
 },
 "semantic": {
 "text": {
 "primary": { "$value": "{color.gray.900}" },
 "secondary": { "$value": "{color.gray.600}" },
 "disabled": { "$value": "{color.gray.400}" }
 },
 "background": {
 "default": { "$value": "{color.white}" },
 "elevated": { "$value": "{color.gray.50}" }
 }
 }
 },
 "spacing": {
 "$type": "dimension",
 "xs": { "$value": "0.25rem" },
 "sm": { "$value": "0.5rem" },
 "md": { "$value": "1rem" },
 "lg": { "$value": "1.5rem" },
 "xl": { "$value": "2rem" }
 },
 "typography": {
 "$type": "fontFamily",
 "sans": { "$value": ["Inter", "system-ui", "sans-serif"] },
 "mono": { "$value": ["JetBrains Mono", "monospace"] }
 },
 "fontSize": {
 "$type": "dimension",
 "sm": { "$value": "0.875rem" },
 "base": { "$value": "1rem" },
 "lg": { "$value": "1.125rem" },
 "xl": { "$value": "1.25rem" }
 }
 }
}
```

### Multi-Theme Support (Light/Dark Mode)

```json
{
 "color": {
 "semantic": {
 "background": {
 "$type": "color",
 "default": {
 "$value": "{color.white}",
 "$extensions": {
 "mode": {
 "dark": "{color.gray.900}"
 }
 }
 }
 }
 }
 }
}
```

### Style Dictionary Configuration (4.0+)

```javascript
// style-dictionary.config.js
export default {
 source: ['tokens//*.json'],
 platforms: {
 css: {
 transformGroup: 'css',
 buildPath: 'build/css/',
 files: [{
 destination: 'variables.css',
 format: 'css/variables'
 }]
 },
 js: {
 transformGroup: 'js',
 buildPath: 'build/js/',
 files: [{
 destination: 'tokens.js',
 format: 'javascript/es6'
 }]
 },
 typescript: {
 transformGroup: 'js',
 buildPath: 'build/ts/',
 files: [{
 destination: 'tokens.ts',
 format: 'typescript/es6-declarations'
 }]
 }
 }
};
```

---

## Figma MCP Integration Workflow

### Setup (Desktop Server)

```bash
# Install Figma desktop app
# Enable MCP server in settings
# Desktop server runs at http://127.0.0.1:3845/mcp
```

### MCP Configuration (Claude Desktop)

```json
{
 "mcpServers": {
 "figma": {
 "command": "figma-mcp",
 "args": [],
 "env": {
 "FIGMA_ACCESS_TOKEN": "your-figma-token"
 }
 }
 }
}
```

### Extracting Design Tokens from Figma

Workflow:
1. Create Figma Variables (Color, Typography, Spacing)
2. Use MCP Server to extract variables:
 - Select frame in Figma
 - Prompt: "Extract all design tokens from this frame"
 - MCP returns DTCG-compatible JSON
3. Transform to Code using Style Dictionary

### Component Code Generation

```
User Workflow:
1. Select component frame in Figma
2. Prompt: "Generate React component from this design"
3. MCP extracts:
 - Component structure
 - Applied design tokens
 - Layout properties (flex, grid)
 - Typography and spacing
4. Output: TypeScript React component with props
```

### Automation Pattern (Link-based workflow)

```typescript
// scripts/sync-figma-tokens.ts
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

async function syncFigmaTokens(figmaFileUrl: string) {
 // Use Figma MCP to extract tokens
 const { stdout } = await execAsync(
 `figma-mcp extract-tokens --url="${figmaFileUrl}"`
 );
 
 const tokens = JSON.parse(stdout);
 
 // Write to tokens directory
 await writeFile('tokens/color.json', JSON.stringify(tokens.color, null, 2));
 await writeFile('tokens/spacing.json', JSON.stringify(tokens.spacing, null, 2));
 
 // Run Style Dictionary build
 await execAsync('npm run tokens:build');
 
 console.log(' Design tokens synchronized from Figma');
}
```

---

## Best Practices

Design Token Architecture:
- Use semantic naming (`color.primary.500` not `color.blue`)
- Implement aliasing for themes (`{color.white}` references)
- Validate DTCG 2025.10 spec compliance
- Version tokens with semantic versioning
- Document token usage in Storybook

Component Development:
- Follow Atomic Design hierarchy (Atoms → Molecules → Organisms)
- Create variant-based props APIs (not separate components)
- Document all props with TypeScript types
- Write Storybook stories for all variants
- Test component accessibility with jest-axe

Accessibility:
- Verify 4.5:1 contrast for all text (WCAG AA)
- Implement keyboard navigation for all interactive elements
- Add ARIA labels to form fields and buttons
- Test with screen readers (NVDA, JAWS, VoiceOver)
- Support `prefers-reduced-motion`

---

## Advanced Patterns

### Type-Safe Design Tokens

```typescript
// scripts/generate-token-types.ts
import { readFileSync, writeFileSync } from 'fs';

interface DTCGToken {
 $value: string | number | string[];
 $type?: string;
 [key: string]: any;
}

function generateTypes(tokens: Record<string, any>, prefix = ''): string {
 let types = '';
 
 for (const [key, value] of Object.entries(tokens)) {
 if (value.$value !== undefined) {
 const tokenPath = `${prefix}${key}`.replace(/\./g, '-');
 types += `export const ${tokenPath} = '${value.$value}';\n`;
 } else {
 types += generateTypes(value, `${prefix}${key}.`);
 }
 }
 
 return types;
}

const colorTokens = JSON.parse(readFileSync('tokens/color.json', 'utf-8'));
const types = generateTypes(colorTokens.$tokens);
writeFileSync('src/tokens/colors.ts', types);
```

### Visual Regression Testing (Chromatic)

```bash
# Install Chromatic
npm install --save-dev chromatic
```

CI/CD Integration:
```yaml
# .github/workflows/chromatic.yml
name: Chromatic

on: [push]

jobs:
 chromatic:
 runs-on: ubuntu-latest
 steps:
 - uses: actions/checkout@v4
 with:
 fetch-depth: 0
 
 - uses: actions/setup-node@v4
 with:
 node-version: 20
 
 - name: Install dependencies
 run: npm ci
 
 - name: Run Chromatic
 uses: chromaui/action@v1
 with:
 projectToken: ${{ secrets.CHROMATIC_PROJECT_TOKEN }}
```

---

## When NOT to Use

- Simple static sites: Overkill for projects without complex UI requirements
- Rapid prototyping: Design systems add overhead during early exploration
- Single-use projects: Token architecture benefits long-term maintenance
- Non-web platforms: This skill focuses on web (React/Vue/TypeScript)

For these cases, consider:
- Plain CSS/Tailwind for static sites
- Component libraries (Material-UI, shadcn/ui) for rapid development
- Platform-specific design systems (iOS HIG, Material Design for Android)

---

## Works Well With

- [Component Architecture](component-architecture.md) - Component patterns and Atomic Design
- [Accessibility WCAG](accessibility-wcag.md) - WCAG 2.2 compliance and testing
- `moai-library-shadcn` - shadcn/ui component library
- `moai-code-frontend` - Frontend architecture patterns
- `moai-lang-unified` - TypeScript best practices

---

## Official Resources

- W3C DTCG: https://designtokens.org
- WCAG 2.2: https://www.w3.org/WAI/WCAG22/quickref/
- Figma MCP: https://help.figma.com/hc/en-us/articles/32132100833559
- Style Dictionary: https://styledictionary.com
- Storybook: https://storybook.js.org

---

Last Updated: 2025-11-21
Status: Production Ready
