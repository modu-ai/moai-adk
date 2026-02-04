# moai-library-nextra Reference

## API Reference

### Theme Configuration API

DocsThemeConfig Interface:
- `logo`: React node for site logo
- `logoLink`: URL for logo click
- `project`: Project link configuration
- `docsRepositoryBase`: Base URL for edit links
- `sidebar`: Sidebar configuration options
- `toc`: Table of contents options
- `footer`: Footer configuration
- `head`: Custom head elements
- `search`: Search configuration
- `useNextSeoProps`: SEO metadata function

Sidebar Options:
- `defaultMenuCollapseLevel`: Default collapse depth
- `toggleButton`: Show toggle button
- `autoCollapse`: Auto-collapse inactive sections

TOC Options:
- `backToTop`: Show back to top button
- `float`: Float table of contents
- `headingDepth`: Heading levels to include

### Navigation API (\_meta.js)

Page Configuration:
```javascript
{
  "page-name": "Display Title",
  "---": "",  // Separator
  "external-link": {
    "title": "Link Title",
    "href": "https://example.com",
    "newWindow": true
  }
}
```

Page Options:
- `title`: Display title in sidebar
- `type`: Page type (page, separator, menu)
- `display`: Display mode (normal, hidden, children)
- `theme`: Page-specific theme overrides

### MDX Component API

Built-in Components:
- `Callout`: Alert/info boxes (type: info, warning, error, default)
- `Tabs`: Tabbed content container
- `Tab`: Individual tab item
- `Cards`: Card grid container
- `Card`: Individual card item
- `Steps`: Step-by-step guide
- `FileTree`: File structure visualization

Callout Types:
- `info`: Blue informational box
- `warning`: Yellow warning box
- `error`: Red error/danger box
- `default`: Gray neutral box

### Search API

FlexSearch Configuration:
- `placeholder`: Search input placeholder
- `emptyResult`: Empty result message
- `loading`: Loading indicator
- `error`: Error message

Search Customization:
- Custom search endpoint
- Result filtering
- Keyboard shortcuts

---

## Configuration Options

### next.config.js

Nextra Plugin Configuration:
```javascript
const withNextra = require("nextra")({
  theme: "nextra-theme-docs",
  themeConfig: "./theme.config.tsx",
  staticImage: true,
  latex: true,
  flexsearch: {
    codeblocks: true
  },
  defaultShowCopyCode: true
});

module.exports = withNextra({
  // Next.js config
  reactStrictMode: true,
  images: {
    domains: ["example.com"]
  }
});
```

### theme.config.tsx

Full Configuration Example:
```typescript
const config: DocsThemeConfig = {
  // Branding
  logo: <span>My Documentation</span>,
  logoLink: "/",

  // Repository
  project: { link: "https://github.com/user/repo" },
  docsRepositoryBase: "https://github.com/user/repo/tree/main",

  // Navigation
  sidebar: {
    defaultMenuCollapseLevel: 1,
    toggleButton: true,
    autoCollapse: true
  },

  // Content
  toc: {
    backToTop: true,
    float: true
  },

  // Footer
  footer: {
    text: `${new Date().getFullYear()} My Project`
  },

  // SEO
  head: (
    <>
      <meta name="viewport" content="width=device-width, initial-scale=1.0" />
      <meta property="og:title" content="My Docs" />
      <meta property="og:description" content="Documentation" />
    </>
  ),

  useNextSeoProps() {
    return {
      titleTemplate: "%s - My Docs"
    };
  }
};
```

### Internationalization (i18n)

next.config.js i18n:
```javascript
module.exports = withNextra({
  i18n: {
    locales: ["en", "ko", "ja"],
    defaultLocale: "en"
  }
});
```

Language-specific \_meta:
```
pages/
  en/
    _meta.json
    index.mdx
  ko/
    _meta.json
    index.mdx
```

---

## Integration Patterns

### Project Structure Pattern

Recommended Directory Layout:
```
docs/
  pages/
    _app.tsx
    _meta.json
    index.mdx
    getting-started/
      _meta.json
      installation.mdx
      configuration.mdx
    guides/
      _meta.json
      basic.mdx
      advanced.mdx
    api/
      _meta.json
      reference.mdx
  public/
    images/
    og-image.png
  components/
    CustomComponent.tsx
  styles/
    custom.css
  theme.config.tsx
  next.config.js
  package.json
```

### Custom Component Integration

Creating Custom Components:
1. Create component in `components/` directory
2. Import in MDX files
3. Use standard React patterns

Global Component Registration:
```typescript
// pages/_app.tsx
import type { AppProps } from "next/app";
import { CustomComponent } from "../components/CustomComponent";

const components = {
  CustomComponent
};

export default function App({ Component, pageProps }: AppProps) {
  return <Component {...pageProps} components={components} />;
}
```

### Static Export Pattern

Static HTML Generation:
```javascript
// next.config.js
module.exports = withNextra({
  output: "export",
  images: {
    unoptimized: true
  },
  trailingSlash: true
});
```

Build Command:
```bash
npm run build
# Output in /out directory
```

### Theme Customization Pattern

CSS Variables Override:
```css
/* styles/custom.css */
:root {
  --nextra-primary-hue: 212deg;
  --nextra-primary-saturation: 100%;
  --nextra-content-width: 90rem;
}

.dark {
  --nextra-bg: 17 17 17;
}
```

Tailwind Integration:
```javascript
// tailwind.config.js
module.exports = {
  content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx}",
    "./theme.config.tsx"
  ]
};
```

---

## Troubleshooting

### Build Errors

Symptoms: Build fails, module not found

Solutions:
1. Clear `.next` cache: `rm -rf .next`
2. Reinstall dependencies: `rm -rf node_modules && npm install`
3. Check Next.js/Nextra version compatibility
4. Verify theme.config.tsx syntax

### MDX Parsing Errors

Symptoms: MDX syntax errors, component not rendering

Solutions:
1. Check JSX syntax in MDX files
2. Verify component imports are correct
3. Ensure no unescaped special characters
4. Check for unclosed tags

### Search Not Working

Symptoms: Search returns no results, index empty

Solutions:
1. Verify FlexSearch is enabled in next.config.js
2. Rebuild the search index: `npm run build`
3. Check for build errors in search indexing
4. Verify content files are in pages directory

### i18n Issues

Symptoms: Wrong language, missing translations

Solutions:
1. Verify locale configuration in next.config.js
2. Check \_meta.json exists for each locale
3. Confirm URL structure matches locale pattern
4. Verify default locale setting

### Styling Issues

Symptoms: Incorrect styles, theme not applying

Solutions:
1. Check CSS import order
2. Verify Tailwind configuration
3. Clear browser cache
4. Check for CSS specificity conflicts

---

## External Resources

### Official Documentation

- [Nextra Documentation](https://nextra.site/)
- [Next.js Documentation](https://nextjs.org/docs)
- [MDX Documentation](https://mdxjs.com/)
- [FlexSearch](https://github.com/nextapps-de/flexsearch)

### Deployment Guides

- [Vercel Deployment](https://vercel.com/docs)
- [GitHub Pages with Next.js](https://nextjs.org/docs/pages/building-your-application/deploying/static-exports)
- [Netlify Deployment](https://docs.netlify.com/)

### Theme Resources

- Nextra Theme Docs (documentation sites)
- Nextra Theme Blog (blog sites)
- Custom theme development guide

### Related Skills

- `moai-docs-generation`: Automated documentation generation
- `moai-workflow-docs`: Documentation workflow management
- `moai-domain-frontend`: Frontend development patterns
- `moai-library-mermaid`: Diagram integration

### Complementary Tools

- Mermaid.js: Diagram generation
- KaTeX: Math rendering
- Shiki: Code syntax highlighting
- next-sitemap: SEO sitemap generation

---

Version: 2.0.0
Last Updated: 2025-12-06
