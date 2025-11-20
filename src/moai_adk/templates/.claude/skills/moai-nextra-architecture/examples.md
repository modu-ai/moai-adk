# Nextra Architecture Examples

Practical implementations for building documentation sites with Nextra (Next.js + MDX framework).

---

## Project Setup

### Complete Nextra Project Initialization

```bash
# Create new Next.js project with TypeScript
npx create-next-app@latest my-docs --typescript --eslint --tailwind --app=false --src-dir=false

cd my-docs

# Install Nextra dependencies
npm install nextra nextra-theme-docs

# Install additional dependencies
npm install sharp  # For optimized images
```

### Configuration Files

```javascript
// next.config.js
const withNextra = require("nextra")({
  theme: "nextra-theme-docs",
  themeConfig: "./theme.config.tsx",
  defaultShowCopyCode: true,
  flexsearch: {
    codeblocks: true,
  },
  staticImage: true,
});

module.exports = withNextra({
  reactStrictMode: true,
  images: {
    unoptimized: false,
  },
  // For static export (GitHub Pages, Netlify, etc.)
  // output: 'export',
  // images: { unoptimized: true },
});
```

```typescript
// theme.config.tsx
import React from "react";
import { DocsThemeConfig } from "nextra-theme-docs";
import { useRouter } from "next/router";

const config: DocsThemeConfig = {
  logo: (
    <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
      <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
        <path d="M12 2L2 7v10c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V7l-10-5z" />
      </svg>
      <span style={{ fontWeight: "bold", fontSize: "18px" }}>
        My Documentation
      </span>
    </div>
  ),

  project: {
    link: "https://github.com/myorg/my-project",
  },

  chat: {
    link: "https://discord.gg/myproject",
  },

  docsRepositoryBase: "https://github.com/myorg/my-project/tree/main/docs",

  footer: {
    text: (
      <span>
        MIT {new Date().getFullYear()} ©{" "}
        <a href="https://mycompany.com" target="_blank">
          My Company
        </a>
        .
      </span>
    ),
  },

  sidebar: {
    titleComponent({ title, type }) {
      if (type === "separator") {
        return (
          <div style={{ fontWeight: "bold", marginTop: "1rem" }}>{title}</div>
        );
      }
      return <>{title}</>;
    },
    defaultMenuCollapseLevel: 1,
    toggleButton: true,
  },

  toc: {
    backToTop: true,
    float: true,
  },

  editLink: {
    text: "Edit this page on GitHub →",
  },

  feedback: {
    content: "Question? Give us feedback →",
    labels: "feedback",
  },

  navigation: {
    prev: true,
    next: true,
  },

  darkMode: true,

  primaryHue: 210, // Blue theme

  head: () => {
    const { asPath } = useRouter();
    const url = `https://docs.mycompany.com${asPath}`;

    return (
      <>
        <meta property="og:url" content={url} />
        <meta property="og:title" content="My Documentation" />
        <meta
          property="og:description"
          content="Comprehensive documentation for My Project"
        />
        <link rel="icon" href="/favicon.ico" />
      </>
    );
  },

  useNextSeoProps() {
    const { asPath } = useRouter();
    if (asPath !== "/") {
      return {
        titleTemplate: "%s – My Docs",
      };
    }
    return {
      titleTemplate: "My Documentation",
    };
  },
};

export default config;
```

---

## Content Structure

### Folder Organization

```
my-docs/
├── pages/
│   ├── _app.tsx              # App wrapper (optional customization)
│   ├── index.mdx             # Homepage
│   ├── _meta.json            # Root navigation
│   │
│   ├── getting-started/
│   │   ├── _meta.json
│   │   ├── installation.mdx
│   │   ├── quickstart.mdx
│   │   └── configuration.mdx
│   │
│   ├── guides/
│   │   ├── _meta.json
│   │   ├── authentication.mdx
│   │   ├── deployment.mdx
│   │   └── best-practices.mdx
│   │
│   └── api-reference/
│       ├── _meta.json
│       ├── functions.mdx
│       ├── classes.mdx
│       └── types.mdx
│
├── components/              # Custom React components
│   ├── DemoButton.tsx
│   └── CodeSandbox.tsx
│
├── public/                 # Static assets
│   ├── images/
│   └── favicon.ico
│
└── styles/                 # Custom styles
    └── globals.css
```

### \_meta.json Configuration

```json
// pages/_meta.json (root navigation)
{
  "index": {
    "title": "Home",
    "type": "page",
    "display": "hidden",
    "theme": {
      "breadcrumb": false,
      "footer": true,
      "sidebar": false,
      "toc": false,
      "pagination": false
    }
  },
  "getting-started": {
    "title": "Getting Started",
    "type": "page"
  },
  "guides": {
    "title": "Guides",
    "type": "page"
  },
  "api-reference": {
    "title": "API Reference",
    "type": "page"
  },
  "---": {
    "type": "separator",
    "title": "Resources"
  },
  "changelog": {
    "title": "Changelog ↗",
    "type": "page",
    "href": "https://github.com/myorg/my-project/releases",
    "newWindow": true
  }
}
```

```json
// pages/getting-started/_meta.json
{
  "installation": "Installation",
  "quickstart": "Quick Start",
  "configuration": "Configuration"
}
```

---

## MDX Components & Syntax

### Advanced MDX Page Example

````mdx
---
title: Authentication Guide
description: Learn how to implement authentication in your application
---

import { Tabs, Callout, Steps, Cards, Card } from "nextra/components";
import { DemoButton } from "../../components/DemoButton";

# Authentication Guide

<Callout type="info" emoji="ℹ️">
  This guide covers both JWT and OAuth2 authentication patterns.
</Callout>

## Overview

Authentication is essential for securing your application. We support two main approaches:

<Cards>
  <Card title="JWT Tokens" href="#jwt-authentication" />
  <Card title="OAuth2" href="#oauth2-flow" />
</Cards>

---

## JWT Authentication

<Steps>

### Install Dependencies

<Tabs items={["npm", "yarn", "pnpm"]}>
  <Tabs.Tab>```bash npm install jsonwebtoken bcrypt ```</Tabs.Tab>
  <Tabs.Tab>```bash yarn add jsonwebtoken bcrypt ```</Tabs.Tab>
  <Tabs.Tab>```bash pnpm add jsonwebtoken bcrypt ```</Tabs.Tab>
</Tabs>

### Create JWT Utility

```typescript filename="utils/jwt.ts" showLineNumbers {5-7, 12-14}
import jwt from "jsonwebtoken";

const JWT_SECRET = process.env.JWT_SECRET!;

export function generateToken(userId: string): string {
  return jwt.sign({ userId }, JWT_SECRET, { expiresIn: "7d" });
}

export function verifyToken(token: string): { userId: string } | null {
  try {
    const payload = jwt.verify(token, JWT_SECRET);
    return payload as { userId: string };
  } catch (error) {
    return null;
  }
}
```
````

### Protect API Routes

```typescript filename="middleware/auth.ts"
import { verifyToken } from "../utils/jwt";

export async function requireAuth(req, res, next) {
  const token = req.headers.authorization?.replace("Bearer ", "");

  if (!token) {
    return res.status(401).json({ error: "No token provided" });
  }

  const payload = verifyToken(token);

  if (!payload) {
    return res.status(401).json({ error: "Invalid token" });
  }

  req.userId = payload.userId;
  next();
}
```

</Steps>

<Callout type="warning" emoji="⚠️">
  Never commit your `JWT_SECRET` to version control. Use environment variables!
</Callout>

---

## Try it Now

Click the demo button to test authentication:

<DemoButton text="Authenticate" />

---

## Next Steps

- [User Management](/guides/user-management)
- [Authorization & RBAC](/guides/authorization)
- [Security Best Practices](/guides/security)

````

---

## Custom Components

### Interactive Demo Button

```typescript
// components/DemoButton.tsx
import { useState } from 'react';

export function DemoButton({ text }: { text: string }) {
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<string | null>(null);

  const handleClick = async () => {
    setLoading(true);
    setResult(null);

    try {
      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 1000));
      setResult('✓ Success! Token generated.');
    } catch (error) {
      setResult('✗ Error occurred.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="my-4 p-4 border border-gray-300 dark:border-gray-700 rounded-lg">
      <button
        onClick={handleClick}
        disabled={loading}
        className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {loading ? 'Processing...' : text}
      </button>
      {result && (
        <p className="mt-2 text-sm text-gray-700 dark:text-gray-300">{result}</p>
      )}
    </div>
  );
}
````

### Mermaid Diagram Component

```typescript
// components/MermaidDiagram.tsx
import { useEffect, useRef } from "react";
import mermaid from "mermaid";

export function MermaidDiagram({ chart }: { chart: string }) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (ref.current) {
      mermaid.initialize({ startOnLoad: false, theme: "neutral" });
      mermaid.render("mermaid-diagram", chart).then(({ svg }) => {
        if (ref.current) {
          ref.current.innerHTML = svg;
        }
      });
    }
  }, [chart]);

  return <div ref={ref} className="my-4" />;
}
```

**Usage in MDX**:

```mdx
import { MermaidDiagram } from "../../components/MermaidDiagram";

<MermaidDiagram
  chart={`
graph TD
    A[User] -->|Login| B[Auth Server]
    B -->|Generate JWT| C[Client]
    C -->|API Request + Token| D[API Server]
    D -->|Verify Token| E[Protected Resource]
`}
/>
```

### Code Sandbox Embed

```typescript
// components/CodeSandbox.tsx
export function CodeSandbox({ sandboxId }: { sandboxId: string }) {
  const url = `https://codesandbox.io/embed/${sandboxId}?fontsize=14&hidenavigation=1&theme=dark`;

  return (
    <iframe
      src={url}
      style={{
        width: "100%",
        height: "500px",
        border: 0,
        borderRadius: "4px",
        overflow: "hidden",
      }}
      title="CodeSandbox Demo"
      allow="accelerometer; ambient-light-sensor; camera; encrypted-media; geolocation; gyroscope; hid; microphone; midi; payment; usb; vr; xr-spatial-tracking"
      sandbox="allow-forms allow-modals allow-popups allow-presentation allow-same-origin allow-scripts"
    />
  );
}
```

---

## Internationalization (i18n)

### Multi-Language Configuration

```javascript
// next.config.js (with i18n)
const withNextra = require("nextra")({
  theme: "nextra-theme-docs",
  themeConfig: "./theme.config.tsx",
});

module.exports = withNextra({
  i18n: {
    locales: ["en", "ko", "ja", "zh"],
    defaultLocale: "en",
  },
});
```

### Language-Specific Content Structure

```
pages/
├── en/
│   ├── _meta.json
│   ├── index.mdx
│   └── getting-started.mdx
├── ko/
│   ├── _meta.json
│   ├── index.mdx
│   └── getting-started.mdx
└── ja/
    ├── _meta.json
    ├── index.mdx
    └── getting-started.mdx
```

### Language Switcher in Theme Config

```typescript
// theme.config.tsx (i18n support)
import { useRouter } from "next/router";

const config: DocsThemeConfig = {
  // ... other config

  i18n: [
    { locale: "en", text: "English" },
    { locale: "ko", text: "한국어" },
    { locale: "ja", text: "日本語" },
    { locale: "zh", text: "中文" },
  ],

  head: () => {
    const { locale } = useRouter();
    return (
      <>
        <meta httpEquiv="Content-Language" content={locale} />
      </>
    );
  },
};
```

---

## Search Integration

### Flexsearch (Built-in)

```javascript
// next.config.js
const withNextra = require("nextra")({
  theme: "nextra-theme-docs",
  themeConfig: "./theme.config.tsx",

  // Enable Flexsearch
  flexsearch: {
    codeblocks: true,
  },
});

module.exports = withNextra();
```

### Algolia DocSearch

```typescript
// theme.config.tsx
const config: DocsThemeConfig = {
  // ... other config

  search: {
    component: null, // Disable default search
  },
};

// Install: npm install @docsearch/react
```

```typescript
// pages/_app.tsx
import { DocSearch } from "@docsearch/react";
import "@docsearch/css";

export default function App({ Component, pageProps }) {
  return (
    <>
      <DocSearch
        appId="YOUR_APP_ID"
        indexName="YOUR_INDEX_NAME"
        apiKey="YOUR_SEARCH_API_KEY"
      />
      <Component {...pageProps} />
    </>
  );
}
```

---

## Deployment

### Static Export (GitHub Pages, Netlify)

```javascript
// next.config.js (static export)
const withNextra = require("nextra")({
  theme: "nextra-theme-docs",
  themeConfig: "./theme.config.tsx",
});

module.exports = withNextra({
  output: "export",
  images: {
    unoptimized: true, // Required for static export
  },
  // For GitHub Pages subdirectory
  // basePath: '/my-docs',
  // assetPrefix: '/my-docs',
});
```

```bash
# Build static site
npm run build

# Output will be in `out/` directory
# Deploy `out/` to GitHub Pages, Netlify, Vercel, etc.
```

### Vercel Deployment

```json
// vercel.json
{
  "buildCommand": "npm run build",
  "outputDirectory": ".next",
  "framework": "nextjs"
}
```

### GitHub Actions CI/CD

```yaml
# .github/workflows/deploy.yml
name: Deploy Nextra Docs

on:
  push:
    branches: [main]

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: "18"
          cache: "npm"

      - name: Install dependencies
        run: npm ci

      - name: Build
        run: npm run build

      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./out
```

---

## Custom Styling

### Global CSS Customization

```css
/* styles/globals.css */

/* Custom color scheme */
:root {
  --nextra-primary-hue: 210deg;
  --nextra-primary-saturation: 100%;
}

/* Custom code block styling */
code {
  @apply rounded-md px-1 py-0.5 bg-gray-100 dark:bg-gray-800 font-mono text-sm;
}

/* Custom callout styling */
.nextra-callout {
  @apply my-4 p-4 rounded-lg border-l-4;
}

.nextra-callout.info {
  @apply bg-blue-50 dark:bg-blue-900/20 border-blue-500;
}

.nextra-callout.warning {
  @apply bg-yellow-50 dark:bg-yellow-900/20 border-yellow-500;
}

/* Custom table styling */
table {
  @apply w-full border-collapse my-4;
}

th {
  @apply bg-gray-100 dark:bg-gray-800 font-semibold p-2 text-left;
}

td {
  @apply border-t border-gray-200 dark:border-gray-700 p-2;
}
```

---

**See also**: [SKILL.md](./SKILL.md) for Nextra architecture overview and best practices
