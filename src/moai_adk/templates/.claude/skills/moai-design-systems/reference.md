# Design Systems Technical Reference

Comprehensive reference material for design system development, covering W3C DTCG 2025.10, WCAG 2.2, and tooling ecosystem.

---

## W3C DTCG 2025.10 Specification

**Official Spec**: https://tr.designtokens.org/format/
**Status**: First stable release (October 2025)

### Token Types
- `color`: Color values
- `dimension`: Spacing, sizing, typography
- `fontFamily`: Font stacks
- `fontWeight`: Font weights
- `duration`: Animation timing

### Theme Support
Use `$extensions` for mode-specific values (light/dark):

```json
{
  "background": {
    "$value": "#ffffff",
    "$extensions": {
      "mode": { "dark": "#1a1a1a" }
    }
  }
}
```

---

## WCAG 2.2 Accessibility

### Contrast Requirements
- **AA Normal text**: 4.5:1
- **AA Large text**: 3:1
- **AAA Normal text**: 7:1
- **AAA Large text**: 4.5:1

### Keyboard Navigation
- Tab order for all interactive elements
- Escape key dismisses modals
- Arrow keys navigate composites
- Enter/Space activate controls

### ARIA Patterns
- `role="button"`, `role="dialog"`, `role="tab"`
- `aria-label`, `aria-describedby`, `aria-invalid`
- `aria-live` for dynamic updates
- `aria-selected`, `aria-expanded` for state

---

## Figma MCP Server

**Guide**: https://help.figma.com/hc/en-us/articles/32132100833559

### Capabilities
- Code generation (React/Vue/TypeScript)
- Design token extraction
- Component metadata extraction
- Layout information parsing

### Server Types
- **Desktop**: `http://127.0.0.1:3845/mcp` (no rate limits)
- **Remote**: `https://mcp.figma.com/mcp` (limited calls/month)

---

## Style Dictionary 4.0

**Website**: https://styledictionary.com

- DTCG 2025.10 compatible
- Multi-platform output (CSS, JS, Android, iOS)
- Custom transforms support
- TypeScript type generation

---

## Storybook 8

**Website**: https://storybook.js.org

- Auto-generates docs from TypeScript types
- Built-in a11y addon for accessibility testing
- Component Story Format (CSF3)
- Visual regression with Chromatic

---

## Tools Compatibility Matrix

| Tool | Version | DTCG Support |
|------|---------|--------------|
| DTCG | 2025.10 | N/A |
| Style Dictionary | 4.0+ | ✅ Full |
| Figma MCP | Latest | ✅ Variables |
| Storybook | 8.x | N/A |
| React | 18+ | N/A |
| TypeScript | 5.0+ | N/A |
