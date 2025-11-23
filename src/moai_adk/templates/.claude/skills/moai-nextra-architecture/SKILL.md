---
name: moai-nextra-architecture
description: Enterprise Nextra documentation framework with Next.js
version: 1.0.1
modularized: true
---

## 📊 Skill Metadata

**Name**: moai-nextra-architecture
**Domain**: Documentation & Static Site Generation
**Freedom Level**: high
**Target Users**: Documentation architects, technical writers, developers
**Invocation**: Skill("moai-nextra-architecture")
**Progressive Disclosure**: SKILL.md (core) → modules/ (detailed configs)
**Last Updated**: 2025-11-23
**Modularized**: true

---

## 🎯 Quick Reference (30 seconds)

**Purpose**: Build professional documentation sites with Nextra + Next.js.

**Nextra Advantages**:
- Zero config MDX (Markdown + JSX seamlessly)
- File-system routing (automatic routes)
- Performance optimized (code splitting, prefetching)
- Theme system (pluggable, customizable)
- i18n built-in (internationalization)

**Core Files**:
- `pages/` - Documentation pages (MDX)
- `theme.config.tsx` - Site configuration
- `_meta.js` - Navigation structure

---

## 📚 Core Patterns (5-10 minutes)

### Pattern 1: Project Structure

**Key Concept**: Organize documentation files logically

**Recommended Structure**:
```
docs/
├── pages/
│   ├── index.mdx          # Homepage
│   ├── getting-started/
│   │   ├── _meta.js       # Section config
│   │   ├── index.mdx
│   │   └── installation.mdx
│   ├── guides/
│   │   ├── _meta.js
│   │   ├── basics.mdx
│   │   └── advanced.mdx
│   └── api/
│       ├── _meta.js
│       └── reference.mdx
├── public/                 # Static assets
├── theme.config.tsx        # Main config
├── next.config.js          # Next.js config
└── package.json
```

### Pattern 2: Theme Configuration

**Key Concept**: Customize site appearance and behavior

**Essential Config**:
```typescript
const config: DocsThemeConfig = {
  // Branding
  logo: <span>My Docs</span>,
  logoLink: '/',

  // Navigation
  project: { link: 'https://github.com/...' },
  docsRepositoryBase: 'https://github.com/.../tree/main',

  // Sidebar
  sidebar: {
    defaultMenuCollapseLevel: 1,
    toggleButton: true,
  },

  // Table of contents
  toc: { backToTop: true },

  // Footer
  footer: { text: 'Built with Nextra' },
};
```

### Pattern 3: Navigation Structure (_meta.js)

**Key Concept**: Control sidebar menu and page ordering

**Example**:
```javascript
// pages/guides/_meta.js
export default {
  'index': 'Overview',
  'getting-started': 'Getting Started',
  'basics': 'Basic Concepts',
  'advanced': 'Advanced Topics',
  '---': '', // Separator
  'faq': 'FAQ',
};
```

### Pattern 4: MDX Content & JSX Integration

**Key Concept**: Mix Markdown with React components

**Example**:
```mdx
# My Documentation

<div className="bg-blue-100 p-4">
  <h3>Important Note</h3>
  <p>You can embed React components directly!</p>
</div>

## Code Examples

export const MyComponent = () => (
  <button onClick={() => alert('Clicked!')}>
    Click me
  </button>
);

<MyComponent />
```

### Pattern 5: Search & SEO Optimization

**Key Concept**: Make documentation discoverable

**Config**:
```typescript
// theme.config.tsx
const config: DocsThemeConfig = {
  // Enable search
  search: {
    placeholder: 'Search docs...',
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
      titleTemplate: '%s - My Docs'
    }
  },
};
```

---

## 📖 Advanced Documentation

This Skill uses Progressive Disclosure. For detailed patterns:

- **[modules/configuration.md](modules/configuration.md)** - Complete theme.config reference
- **[modules/mdx-components.md](modules/mdx-components.md)** - MDX component library
- **[modules/i18n-setup.md](modules/i18n-setup.md)** - Internationalization guide
- **[modules/deployment.md](modules/deployment.md)** - Hosting & deployment

---

## 🎨 Theme Options

**Built-in Themes**:
- **nextra-theme-docs** (recommended for documentation)
- **nextra-theme-blog** (for blogs)

**Customization**:
- CSS variables for colors
- Custom sidebar components
- Footer customization
- Navigation layout

---

## 🚀 Deployment

**Popular Platforms**:
- **Vercel** (zero-config, recommended)
- **GitHub Pages** (free, self-hosted)
- **Netlify** (flexible, CI/CD)
- **Custom servers** (full control)

**Vercel Deployment**:
```bash
npm install -g vercel
vercel
# Select project and deploy
```

---

## 🔗 Integration with Other Skills

**Complementary Skills**:
- Skill("moai-docs-generation") - Auto-generate docs from code
- Skill("moai-docs-unified") - Validate documentation quality
- Skill("moai-cc-claude-md") - Markdown formatting

---

## 📈 Version History

**1.0.1** (2025-11-23)
- 🔄 Refactored with Progressive Disclosure
- ✨ Configuration patterns highlighted
- ✨ MDX integration guide

**1.0.0** (2025-11-12)
- ✨ Nextra architecture guide
- ✨ Theme configuration
- ✨ i18n support

---

**Maintained by**: alfred
**Domain**: Documentation Architecture
**Generated with**: MoAI-ADK Skill Factory
