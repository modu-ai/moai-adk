# MoAI-ADK Documentation Site

<!-- @CODE:NEXTRA-INIT-001 @CODE:NEXTRA-CONFIG-001 - Project documentation and configuration -->

> **Next.js 14 + Nextra 4.0** based documentation site for MoAI-ADK
>
> Static-generated documentation platform with automatic deployment to Vercel

## Getting Started

### Prerequisites

- Node.js 20.x or higher
- npm 10.x or yarn 1.22.x+

### Installation

```bash
# Install dependencies
npm install

# Start development server
npm run dev
```

Visit [http://localhost:3000](http://localhost:3000) to view the site.

### Development

```bash
# Development server (with HMR)
npm run dev

# Production build
npm run build

# Preview production build
npm run start

# Lint check
npm run lint
```

### Project Structure

```
docs-site/
├── app/              # Next.js 14 App Router
├── pages/            # Nextra documentation pages
├── public/           # Static assets
├── theme.config.tsx  # Nextra theme configuration
├── next.config.js    # Next.js configuration
└── package.json      # Dependencies
```

### Features

- ✅ Next.js 14 App Router
- ✅ Nextra 4.0 Documentation Theme
- ✅ TypeScript Support
- ✅ Hot Module Replacement (HMR)
- ✅ Static Site Generation (SSG)
- ✅ Vercel Deployment Ready

### Deployment

This site is configured for Vercel deployment. Push to the main branch to trigger automatic deployment.

### License

Part of MoAI-ADK project.
