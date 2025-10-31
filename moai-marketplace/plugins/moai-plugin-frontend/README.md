# Frontend Plugin

**Next.js 16 + React 19.2 scaffolding** — TypeScript strict typing, shadcn/ui components, Tailwind CSS 4, performance optimization, E2E testing.

## 🎯 What It Does

Build modern, performant frontend applications with Next.js and React:

```bash
/plugin install moai-plugin-frontend
```

**Automatically provides**:
- ⚛️ React 19.2 with use() hook and Server Components
- 🎨 shadcn/ui component library with Radix UI integration
- 🎯 Tailwind CSS 4 with design tokens and theming
- 🚀 Next.js 16 App Router with streaming and middleware
- ⚡ Performance optimization and Core Web Vitals
- 🧪 Playwright E2E testing and component testing

## 🏗️ Architecture

### 4 Specialist Agents

| Agent | Role | When to Use |
|-------|------|------------|
| **Frontend Architect** | App structure, routing, performance | Project setup |
| **Design System Manager** | Component library, theming, consistency | Design systems |
| **TypeScript Specialist** | Type safety, advanced patterns, validation | Code quality |
| **Performance Optimizer** | Optimization, Core Web Vitals, bundle size | Performance tuning |

### 5 Skills

1. **moai-lang-nextjs-advanced** — App Router, Server Components, middleware, streaming
2. **moai-lang-typescript** — Strict typing, generics, type guards, advanced patterns
3. **moai-design-shadcn-ui** — Component customization, Radix UI integration, accessibility
4. **moai-design-tailwind-v4** — Utility-first CSS, design tokens, performance engine
5. **moai-domain-frontend** — Frontend patterns, state management, accessibility

## ⚡ Quick Start

### Installation

```bash
/plugin install moai-plugin-frontend
```

### Use with MoAI-ADK

The frontend plugin provides agents for:

1. **Architecture** - Frontend Architect designs app structure
2. **Components** - Design System Manager builds components
3. **Typing** - TypeScript Specialist ensures type safety
4. **Performance** - Performance Optimizer tunes Core Web Vitals

## 📊 Project Structure

```
frontend-app/
├── app/                          # Next.js App Router
│   ├── layout.tsx               # Root layout
│   ├── page.tsx                 # Home page
│   └── [route]/page.tsx         # Dynamic routes
├── components/                   # React components
│   ├── ui/                      # shadcn/ui components
│   ├── features/                # Feature components
│   └── shared/                  # Shared components
├── lib/                         # Utility functions
│   ├── utils.ts
│   └── hooks.ts
├── styles/                      # Global styles
│   └── globals.css
├── types/                       # TypeScript types
│   └── index.ts
├── public/                      # Static assets
└── e2e/                        # Playwright tests
    └── example.spec.ts
```

## 🎨 Features

### React 19 & Next.js 16
- **Server Components** - Reduced client-side JavaScript
- **use() Hook** - Unwrap promises and context
- **App Router** - File-based routing with layouts
- **Streaming** - Incremental UI rendering
- **Middleware** - Request-level routing logic

### TypeScript
- **Strict Mode** - All flags enabled
- **Generics** - Reusable type-safe components
- **Type Guards** - Runtime type checking
- **Utility Types** - Advanced type operations
- **Path Aliases** - Clean import paths

### shadcn/ui Components
- **Customizable** - Use and customize any component
- **Radix UI** - Headless UI library
- **Accessible** - WAI-ARIA compliant
- **TypeScript** - Fully typed components
- **Dark Mode** - Built-in theme support

### Tailwind CSS 4
- **Performance Engine** - Optimized CSS output
- **Design Tokens** - Consistent spacing, colors, typography
- **Dark Mode** - Automatic theme switching
- **Responsive Design** - Mobile-first approach
- **Custom Utilities** - Extend for specific needs

### Performance
- **Image Optimization** - next/image component
- **Code Splitting** - Automatic chunk splitting
- **CSS-in-JS** - Style optimization
- **Bundle Analysis** - Identify bottlenecks
- **Core Web Vitals** - LCP, FID, CLS tracking

## 🚀 Typical Development Workflow

### Project Setup

```
Initialize Next.js project
    ↓
[Frontend Architect]
├─ Configure App Router
├─ Setup middleware
└─ Configure build optimization
    ↓
[Design System Manager]
├─ Setup shadcn/ui
├─ Configure Tailwind
└─ Create design tokens
    ↓
[TypeScript Specialist]
├─ Enable strict mode
├─ Create type definitions
└─ Setup path aliases
    ↓
[Performance Optimizer]
├─ Configure bundle analysis
├─ Setup performance monitoring
└─ Optimize images and fonts
```

### Component Development

```
Design component
    ↓
[Design System Manager]
├─ Select shadcn/ui base
├─ Customize styling
└─ Add theme support
    ↓
[TypeScript Specialist]
├─ Define props interface
├─ Add prop validation
└─ Create story file
    ↓
[Performance Optimizer]
├─ Analyze bundle impact
├─ Optimize for SSR
└─ Add lazy loading if needed
```

### Testing

```
Write component
    ↓
Create Playwright test
    ├─ Render component
    ├─ Test interactions
    └─ Verify a11y
    ↓
Run tests locally
    ↓
Deploy and monitor
```

## 📚 Skills Explained

### moai-lang-nextjs-advanced
Next.js 16 mastery:
- **App Router** - Modern file-based routing
- **Server Components** - Render on server by default
- **Streaming** - Incremental static regeneration
- **Middleware** - Request interception
- **Dynamic Routes** - Segment routing patterns

### moai-lang-typescript
TypeScript best practices:
- **Strict Typing** - All strict flags enabled
- **Generics** - Type-safe reusable code
- **Type Guards** - Discriminated unions
- **Utility Types** - Advanced type operations
- **Module Resolution** - Import organization

### moai-design-shadcn-ui
shadcn/ui component integration:
- **Customization** - Tailored component variants
- **Radix UI Foundation** - Accessible primitives
- **Dark Mode** - Theme switching
- **Composition** - Building complex UIs
- **Typography** - Text component variants

### moai-design-tailwind-v4
Tailwind CSS 4 mastery:
- **Utility Classes** - Rapid UI development
- **Design Tokens** - Consistent theming
- **Performance** - Optimized CSS output
- **Responsive Design** - Mobile-first approach
- **Plugin System** - Extended functionality

### moai-domain-frontend
Frontend architecture:
- **Component Architecture** - Atomic design patterns
- **State Management** - Context, Redux, Zustand
- **Performance** - Code splitting, lazy loading
- **Accessibility** - WCAG compliance
- **SEO** - Metadata, structured data

## 🔗 Integration with Other Plugins

- **Backend Plugin**: Consume APIs from FastAPI backends
- **DevOps Plugin**: Deploy to Vercel with CI/CD
- **Design Plugin**: Sync designs from Figma

## 🔧 Common Frontend Tasks

### Setup New Project
1. Initialize Next.js with TypeScript
2. Configure Tailwind CSS 4
3. Install shadcn/ui components
4. Setup project structure

### Build Components
1. Design component interface
2. Create shadcn/ui base
3. Add TypeScript types
4. Write Playwright tests

### Performance Tuning
1. Analyze bundle size
2. Optimize images
3. Code split by route
4. Monitor Core Web Vitals

### Deploy
1. Build production bundle
2. Run performance checks
3. Deploy to Vercel
4. Monitor performance

## 📖 Documentation

- Next.js: https://nextjs.org/docs
- React: https://react.dev
- shadcn/ui: https://ui.shadcn.com
- Tailwind CSS: https://tailwindcss.com
- TypeScript: https://www.typescriptlang.org

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## 📄 License

MIT License - See LICENSE file for details

---

**Created by**: GOOS
**Version**: 1.0.0-dev
**Status**: Development
**Updated**: 2025-10-31
