---
name: moai-library-shadcn-cli-registry
description: shadcn/cli v4 commands, preset system, RTL support, migrations, and registry management
---

## CLI v4 Command Reference

### Initialization Commands

init: Scaffold a new project with shadcn/ui configuration. Supports flags for preset (bundled design system), template (framework selection), base (radix or base primitive library), rtl (right-to-left layout), and monorepo setup.

Available framework templates: Next.js, Vite, TanStack Start, React Router, Astro, Laravel.

### Component Management

add: Install components from the registry. Supports --dry-run (preview changes without writing), --diff (show registry updates), and --view (inspect component payload). Components auto-detect the project's base library (Radix or Base UI).

search: Find components in the registry by name or category.

docs: Fetch documentation, examples, and API references for a specific component.

### Project Information

info: Display project configuration including framework, Tailwind version, CSS variables, aliases, base library, icon library, installed components, and resolved paths. Outputs JSON when used with --json flag.

### Migration Commands

migrate radix: Convert individual @radix-ui/react-* package imports to the unified radix-ui package. Rewrites imports from `import * as DialogPrimitive from "@radix-ui/react-dialog"` to `import { Dialog } from "radix-ui"`.

migrate rtl: Convert physical CSS positioning classes to logical equivalents for RTL support. Transforms left-*/right-* to start-*/end-*, ml-*/mr-* to ms-*/me-*, pl-*/pr-* to ps-*/pe-*, text-left/text-right to text-start/text-end. Also handles animation classes and icon flip directives.

### Build and Distribution

build: Build registry items for distribution. Used when authoring custom component registries.

## Preset System

Presets encapsulate a complete design system configuration in a single string code. A preset bundles colors, theme configuration, icon library, fonts, and border radius settings.

Usage: Initialize with a preset code via the init command's --preset flag. Presets are generated through the shadcn/create builder interface with live preview. Projects can switch presets by re-running init with a different preset code.

Presets are AI-aware and work across Claude, Codex, v0, and Replit for consistent scaffolding.

## RTL Support

shadcn/ui provides first-class right-to-left support for Arabic, Hebrew, Persian, and other RTL languages. The CLI transforms CSS classes at install-time (not runtime), converting physical properties to logical equivalents.

Key transformations:
- Positioning: left-* becomes start-*, right-* becomes end-*
- Margin: ml-* becomes ms-*, mr-* becomes me-*
- Padding: pl-* becomes ps-*, pr-* becomes pe-*
- Text alignment: text-left becomes text-start, text-right becomes text-end
- Icons: Auto-flipped with rtl:rotate-180 utility
- Animations: slide-in-from-left becomes slide-in-from-start

Components with inline positioning (Tooltip, Popover, Select, Combobox, Context Menu, Dropdown Menu, Hover Card, Menubar) support side="inline-start" and side="inline-end" values that auto-adapt for LTR/RTL.

Use the migrate rtl command to convert existing projects to logical CSS properties.

## Unified Radix UI Package

Individual @radix-ui/react-* packages are replaced by a single radix-ui package. Import components directly: `import { Dialog } from "radix-ui"` instead of `import * as DialogPrimitive from "@radix-ui/react-dialog"`.

Benefits: Cleaner package.json, simpler imports, unified versioning.

Use the migrate radix command to automatically rewrite imports across the project.

## Registry System

The registry supports multiple item types for component distribution:

- registry:ui - Standard UI components
- registry:block - Full page blocks and layouts
- registry:base - Complete design system payload
- registry:font - Font configuration with provider, family, variable, and import metadata

Font registry items specify the font family, provider (e.g., google), import name, and CSS variable binding. Install fonts with the add command.

Registry items can be sourced from the official registry, third-party registries, or local files. The build command packages items for distribution.

## shadcn/skills Integration

The shadcn/skills system provides AI assistants with project context for component and registry workflows. Install with: pnpm dlx skills add shadcn/ui

When activated, it runs the info command with JSON output to inject project configuration into the AI context, including framework, Tailwind version, aliases, base library, icon library, installed components, and resolved paths.

This enables AI assistants to understand the project setup and generate correct component code, handle preset switching, and manage registry operations.

## MCP Server Integration

shadcn/ui provides a Model Context Protocol server for AI tool integration. The MCP server enables searching components, browsing the registry, and installing components from within AI assistants.

Setup instructions are available at ui.shadcn.com/docs/mcp.

## Framework Support

shadcn/cli v4 provides templates and full support for:

- Next.js (with App Router and dark mode)
- Vite (with React and dark mode)
- TanStack Start
- React Router
- Astro
- Laravel

Monorepo scaffolding is supported via the --monorepo flag during initialization.

---

Last Updated: 2026-03-28
Related: [Main Skill](../SKILL.md), [Components](shadcn-components.md), [Theming](shadcn-theming.md)
