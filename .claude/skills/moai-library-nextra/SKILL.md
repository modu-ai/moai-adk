---
name: moai-library-nextra
description: Enterprise Nextra documentation framework with Next.js. Use when building documentation sites, knowledge bases, or API reference documentation.
version: 2.1.0
modularized: true
allowed-tools: Read, Write, Edit, Grep, Glob
aliases:
 - moai-library-nextra
category: library
---

## Quick Reference (30 seconds)

Purpose: Build professional documentation sites with Nextra + Next.js.

Nextra Advantages:

- Zero config MDX (Markdown + JSX seamlessly)
- File-system routing (automatic routes)
- Performance optimized (code splitting, prefetching)
- Theme system (pluggable, customizable)
- i18n built-in (internationalization)

Core Files:

- `pages/` - Documentation pages (MDX)
- `theme.config.tsx` - Site configuration
- `_meta.js` - Navigation structure

## Implementation Guide (5 minutes)

### Features

- Nextra 3.x/4.x documentation framework architecture patterns
- Next.js 14/15 integration with optimal configuration
- Theme customization via theme.config.tsx or Layout props
- Advanced search with FlexSearch integration
- Internationalization (i18n) support
- MDX-powered content with React components
- App Router support (Nextra 4.x) with Turbopack compatibility

### When to Use

- Building documentation sites with modern React features
- Creating knowledge bases with advanced search capabilities
- Developing multi-language documentation portals
- Implementing custom documentation themes
- Integrating interactive examples in technical docs

### Core Patterns

Pattern 1: Nextra Project Setup

```bash
# Initialize Nextra docs site
npx create-nextra-app@latest my-docs --template docs

# Project structure
pages/
 _app.tsx (custom App component)
 index.mdx (home page)
 docs/
 guide.mdx
 api.mdx
 _meta.json (navigation config)
```

Pattern 2: Custom Theme Configuration

```typescript
// theme.config.tsx
export default {
 logo: <span>My Documentation</span>,
 project: { link: "https://github.com/user/repo" },
 docsRepositoryBase: "https://github.com/user/repo/tree/main",
 useNextSeoProps: () => ({
 titleTemplate: "%s – My Docs",
 }),
};
```

Pattern 3: MDX with React Components

```mdx
import { Callout } from "nextra/components";

# API Reference

<Callout type="info">This API requires authentication.</Callout>

<CustomCodeBlock language="typescript">// Your code here</CustomCodeBlock>
```

## Core Patterns (5-10 minutes)

### Pattern 1: Project Structure

Key Concept: Organize documentation files logically

Recommended Structure:

```
docs/
 pages/
 index.mdx # Homepage
 getting-started/
 _meta.js # Section config
 index.mdx
 installation.mdx
 guides/
 _meta.js
 basics.mdx
 advanced.mdx
 api/
 _meta.js
 reference.mdx
 public/ # Static assets
 theme.config.tsx # Main config
 next.config.js # Next.js config
 package.json
```

### Pattern 2: Theme Configuration

Key Concept: Customize site appearance and behavior

Essential Config:

```typescript
const config: DocsThemeConfig = {
 // Branding
 logo: <span>My Docs</span>,
 logoLink: "/",

 // Navigation
 project: { link: "https://github.com/..." },
 docsRepositoryBase: "https://github.com/.../tree/main",

 // Sidebar
 sidebar: {
 defaultMenuCollapseLevel: 1,
 toggleButton: true,
 },

 // Table of contents
 toc: { backToTop: true },

 // Footer
 footer: { text: "Built with Nextra" },
};
```

### Pattern 3: Navigation Structure (\_meta.js)

Key Concept: Control sidebar menu and page ordering

Example:

```javascript
// pages/guides/_meta.js
export default {
 index: "Overview",
 "getting-started": "Getting Started",
 basics: "Basic Concepts",
 advanced: "Advanced Topics",
 "---": "", // Separator
 faq: "FAQ",
};
```

### Pattern 4: MDX Content & JSX Integration

Key Concept: Mix Markdown with React components

Example:

```mdx
# My Documentation

<div className="bg-blue-100 p-4">
 <h3>Important Note</h3>
 <p>You can embed React components directly!</p>
</div>

## Code Examples

export const MyComponent = () => (
 <button onClick={() => alert("Clicked!")}>Click me</button>
);

<MyComponent />
```

### Pattern 5: Search & SEO Optimization

Key Concept: Make documentation discoverable

Config:

```typescript
// theme.config.tsx
const config: DocsThemeConfig = {
 // Enable search
 search: {
 placeholder: "Search docs...",
 },

 // SEO metadata
 head: (
 <>
 <meta name="og:title" content="My Documentation" />
 <meta name="og:description" content="Complete guide" />
 <meta name="og:image" content="/og-image.png" />
 </>
 ),

 // Analytics
 useNextSeoProps() {
 return {
 titleTemplate: "%s - My Docs",
 };
 },
};
```

---

## Advanced Documentation

This Skill uses Progressive Disclosure. For detailed patterns:

- [modules/configuration.md](modules/configuration.md) - Complete theme.config reference
- [modules/mdx-components.md](modules/mdx-components.md) - MDX component library
- [modules/i18n-setup.md](modules/i18n-setup.md) - Internationalization guide
- [modules/deployment.md](modules/deployment.md) - Hosting & deployment

---

## Theme Options

Built-in Themes:

- nextra-theme-docs (recommended for documentation)
- nextra-theme-blog (for blogs)

Customization:

- CSS variables for colors
- Custom sidebar components
- Footer customization
- Navigation layout

---

## Deployment

Popular Platforms:

- Vercel (zero-config, recommended)
- GitHub Pages (free, self-hosted)
- Netlify (flexible, CI/CD)
- Custom servers (full control)

Vercel Deployment:

```bash
npm install -g vercel
vercel
# Select project and deploy
```

---

## Integration with Other Skills

Complementary Skills:

- Skill("moai-docs-generation") - Auto-generate docs from code
- Skill("moai-workflow-docs") - Validate documentation quality
- Skill("moai-cc-claude-md") - Markdown formatting

---

## Version History

2.1.0 (2025-12-30)

- Updated configuration.md with complete Nextra-specific theme.config.tsx patterns
- Added Nextra 4.x App Router configuration patterns
- Updated version compatibility for Next.js 14/15
- Added Turbopack support documentation

2.0.0 (2025-11-23)

- Refactored with Progressive Disclosure
- Configuration patterns highlighted
- MDX integration guide

1.0.0 (2025-11-12)

- Nextra architecture guide
- Theme configuration
- i18n support

---

Maintained by: alfred
Domain: Documentation Architecture
Generated with: MoAI-ADK Skill Factory

---

## Works Well With

Agents:
- workflow-docs - Documentation generation
- code-frontend - Nextra implementation
- workflow-spec - Architecture documentation

Skills:
- moai-docs-generation - Content generation
- moai-workflow-docs - Documentation validation
- moai-library-mermaid - Diagram integration

Commands:
- `/moai:3-sync` - Documentation deployment
- `/moai:0-project` - Nextra project initialization
