---
name: advanced-deployment-patterns
parent: moai-library-nextra
description: Mobile optimization, accessibility, and SEO
---

# Module 3: Advanced Deployment & Patterns

## Mobile Optimization

### Mobile-First Layout

```tsx
// components/ui/MobileLayout.tsx
import { useState } from 'react'

export function MobileLayout({ children }) {
 const [sidebarOpen, setSidebarOpen] = useState(false)

 return (
 <div className="min-h-screen bg-background">
 {/* Mobile Header */}
 <header className="sticky top-0 z-50 border-b md:hidden">
 <div className="flex h-14 items-center px-4">
 <button
 onClick={() => setSidebarOpen(true)}
 className="mr-4 p-2 rounded-md hover:bg-accent"
 >
 <svg className="h-5 w-5" fill="none" stroke="currentColor">
 <path strokeLinecap="round" d="M4 6h16M4 12h16M4 18h16" />
 </svg>
 </button>
 <h1 className="text-lg font-semibold">Documentation</h1>
 </div>
 </header>

 <main className="flex-1 md:flex">
 <aside className="hidden md:block w-64 border-r">
 {/* Sidebar */}
 </aside>
 <div className="flex-1 px-4 py-6">
 {children}
 </div>
 </main>
 </div>
 )
}
```

## Accessibility (WCAG 2.1)

### Semantic HTML Structure

```tsx
// components/ui/AccessibleContent.tsx
export function AccessibleContent({ title, children }) {
 return (
 <main id="main-content" role="main" aria-labelledby="page-title">
 <header>
 <h1 id="page-title" className="sr-only">
 {title}
 </h1>
 </header>

 <div className="prose prose-lg max-w-none">
 {children}
 </div>

 <a
 href="#main-content"
 className="sr-only focus:not-sr-only focus:absolute focus:top-4"
 >
 Skip to main content
 </a>
 </main>
 )
}
```

### Keyboard Navigation

```tsx
// components/ui/KeyboardNavigation.tsx
import { useEffect } from 'react'

export function KeyboardNavigation() {
 useEffect(() => {
 const handleKeyDown = (event: KeyboardEvent) {
 if (event.key === 'Escape') {
 const modal = document.querySelector('[role="dialog"]')
 if (modal) {
 const closeButton = modal.querySelector('button[aria-label*="Close"]')
 closeButton?.dispatchEvent(new MouseEvent('click'))
 }
 }
 }

 document.addEventListener('keydown', handleKeyDown)
 return () => document.removeEventListener('keydown', handleKeyDown)
 }, [])

 return null
}
```

## SEO Optimization

### Dynamic SEO Implementation

```typescript
// lib/seo.ts
import type { Metadata } from 'next'

export function generateSEOMetadata({
 title,
 description,
 path = '',
 image,
 keywords = []
}: {
 title: string
 description: string
 path?: string
 image?: string
 keywords?: string[]
}): Metadata {
 const baseUrl = 'https://docs.example.com'
 const url = `${baseUrl}${path}`

 return {
 title,
 description,
 keywords: keywords.join(', '),
 openGraph: {
 type: 'article',
 locale: 'en_US',
 url,
 title,
 description,
 images: image ? [{
 url: image,
 width: 1200,
 height: 630,
 }] : [],
 },
 twitter: {
 card: 'summary_large_image',
 title,
 description,
 },
 alternates: {
 canonical: url,
 },
 robots: {
 index: true,
 follow: true,
 },
 }
}
```

## Production Deployment

```bash
# Build optimization
npm run build

# Static export
npm run export

# Deploy to Vercel
vercel --prod

# Deploy to Netlify
netlify deploy --prod --dir=out
```

---

References:
- [MDN Web Accessibility](https://developer.mozilla.org/en-US/docs/Web/Accessibility)
- [Google SEO Guide](https://developers.google.com/search/docs/fundamentals/seo-starter-guide)
