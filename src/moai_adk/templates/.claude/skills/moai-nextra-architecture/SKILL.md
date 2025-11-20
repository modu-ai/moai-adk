---
name: moai-nextra-architecture
version: 4.0.0
updated: 2025-11-20
status: stable
description: Nextra (Next.js Docs Framework) architecture, MDX, and theme config
allowed-tools: [Read, Bash, WebSearch, WebFetch]
---

# Nextra Architecture Expert

**Next.js-Based Documentation Framework**

> **Focus**: MDX Authoring, File-Based Routing, Theme Configuration  
> **Stack**: Next.js 14, Nextra 3, MDX 3, Storybook (optional)

---

## Overview

Nextra is a static site generator built on Next.js, optimized for documentation with zero-config MDX and automatic routing.

### Core Features

- **File-System Routing**: Automatic route generation from folder structure.
- **MDX Support**: Markdown + React components (`.mdx` files).
- **Theme System**: Pre-built themes (docs, blog) or custom.
- **i18n**: Built-in internationalization support.
- **Search**: Built-in Algolia/Flexsearch integration.

---

## Implementation Patterns

### 1. Installation & Setup

```bash
npm install next nextra nextra-theme-docs react react-dom
```

**`next.config.js`**:

```javascript
const withNextra = require("nextra")({
  theme: "nextra-theme-docs",
  themeConfig: "./theme.config.tsx",
});

module.exports = withNextra();
```

**`theme.config.tsx`** (Core Configuration):

```typescript
import { DocsThemeConfig } from "nextra-theme-docs";

const config: DocsThemeConfig = {
  logo: <span>My Docs</span>,
  project: {
    link: "https://github.com/user/project",
  },
  docsRepositoryBase: "https://github.com/user/project/tree/main/docs",
  footer: {
    text: "© 2025 My Company",
  },
  sidebar: {
    defaultMenuCollapseLevel: 1,
  },
  darkMode: true,
};

export default config;
```

### 2. Content Structure

```
pages/
├── index.mdx          # Homepage
├── _meta.json         # Navigation metadata
├── getting-started/
│   ├── installation.mdx
│   └── quickstart.mdx
└── api-reference/
    ├── functions.mdx
    └── classes.mdx
```

**`_meta.json`** (Navigation Config):

```json
{
  "index": "Introduction",
  "getting-started": "Getting Started",
  "api-reference": "API Reference"
}
```

### 3. MDX Components & Syntax

**Tabs, Callouts, Code Blocks**:

````mdx
---
title: Quick Start Guide
---

import { Tabs, Callout } from "nextra/components";

# Quick Start

<Callout type="info">This guide assumes Node.js 18+ is installed.</Callout>

<Tabs items={["npm", "yarn", "pnpm"]}>
  <Tabs.Tab>```bash npm install my-package ```</Tabs.Tab>
  <Tabs.Tab>```bash yarn add my-package ```</Tabs.Tab>
  <Tabs.Tab>```bash pnpm add my-package ```</Tabs.Tab>
</Tabs>

## Features

```typescript showLineNumbers {3-5}
function hello(name: string) {
  // Highlighted lines
  console.log(`Hello, ${name}!`);
  return true;
}
```
````

````

### 4. Custom Components

Embed React components in MDX:

```tsx
// components/DemoButton.tsx
export function DemoButton({ text }: { text: string }) {
  return <button className="btn-primary">{text}</button>;
}
````

```mdx
import { DemoButton } from "../components/DemoButton";

# My Page

Click the button below:

<DemoButton text="Try it now!" />
```

### 5. Localization (i18n)

Create separate folders per locale:

```
pages/
├── en/
│   └── docs.mdx
├── ko/
│   └── docs.mdx
└── _meta.json  # Define i18n settings
```

**`next.config.js`** (i18n):

```javascript
const withNextra = require("nextra")({
  theme: "nextra-theme-docs",
  themeConfig: "./theme.config.tsx",
  i18n: {
    locales: ["en", "ko", "ja"],
    defaultLocale: "en",
  },
});

module.exports = withNextra();
```

---

## Best Practices

1.  **Consistent Frontmatter**: Always define `title` and optional `description` in MDX.
2.  **\_meta.json for Structure**: Don't rely on alphabetical ordering; use `_meta.json`.
3.  **Component Imports**: Keep custom components in `/components` for clarity.
4.  **Static Export**: For GitHub Pages, set `output: 'export'` in `next.config.js`.

---

## Validation Checklist

- [ ] **Routing**: Does folder structure match navigation?
- [ ] **\_meta.json**: Are all pages listed?
- [ ] **Frontmatter**: Every MDX has `title`?
- [ ] **Code Blocks**: Language hints added (```typescript)?
- [ ] **Dark Mode**: Theme toggle configured?

---

## Related Skills

- `moai-lib-shadcn-ui`: Component library for MDX
- `moai-domain-frontend`: Next.js patterns
- `moai-project-documentation`: Documentation standards

---

**Last Updated**: 2025-11-20
