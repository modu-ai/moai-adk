# Web Interface Guidelines

Comprehensive web interface guidelines compliance checker from Vercel Labs. Review UI code for compliance with accessibility, performance, UX, and implementation best practices.

---

## Overview

The Web Interface Guidelines provide a comprehensive set of rules for building accessible, performant, and user-friendly web interfaces. These guidelines cover HTML structure, accessibility (a11y), forms, animation, typography, content handling, images, performance, navigation, touch interaction, layout, and theming.

### Guidelines Source

The latest guidelines are maintained at:
https://github.com/vercel-labs/web-interface-guidelines

### Usage Patterns

Use these guidelines when:
- Reviewing UI code for compliance issues
- Implementing new components or pages
- Conducting accessibility audits
- Optimizing web performance
- Ensuring consistent UX patterns

---

## HTML Structure

### Document Structure

- Use semantic HTML5 elements (`<nav>`, `<main>`, `<article>`, `<section>`, `<aside>`, `<footer>`)
- Include proper lang attribute on `<html>` element
- Ensure proper heading hierarchy (no skipped levels)
- Include skip link for main content
- Add `scroll-margin-top` on heading anchors for proper scroll position

Example:
```tsx
export default function Layout() {
  return (
    <html lang="en">
      <body>
        <a href="#main" className="sr-only focus:not-sr-only">
          Skip to main content
        </a>
        <Header />
        <main id="main" style={{ scrollMarginTop: 80 }}>
          {children}
        </main>
        <Footer />
      </body>
    </html>
  )
}
```

### Semantic HTML

Use appropriate semantic elements over generic divs:
- `<nav>` for navigation menus
- `<main>` for primary content
- `<article>` for self-contained content
- `<section>` for thematic grouping
- `<aside>` for tangentially related content
- `<header>` and `<footer>` for section headers/footers

---

## Accessibility (a11y)

### Focus States

- Interactive elements need visible focus: `focus-visible:ring-*` or equivalent
- Never `outline-none` / `outline: none` without focus replacement
- Use `:focus-visible` over `:focus` (avoid focus ring on click)
- Group focus with `:focus-within` for compound controls

Example:
```tsx
// Correct focus states
<button className="focus-visible:ring-2 focus-visible:ring-blue-500">
  Click me
</button>

// Compound control focus
<div className="focus-within:ring-2 focus-within:ring-blue-500">
  <input type="text" placeholder="Search..." />
  <button>Search</button>
</div>
```

### ARIA Attributes

- Use `aria-label` for icon-only buttons
- Use `aria-describedby` for additional context
- Use `aria-live` for dynamic content regions
- Use `aria-expanded` for toggle controls
- Use proper heading hierarchy before ARIA (use native HTML first)
- Ensure `role` is used only when necessary (prefer semantic HTML)

Example:
```tsx
// Icon button with aria-label
<button aria-label="Close dialog">
  <XIcon />
</button>

// Described by additional context
<input
  type="text"
  aria-describedby="password-hint"
/>
<p id="password-hint">Must be at least 8 characters</p>

// Live region for dynamic content
<div aria-live="polite" aria-atomic="true">
  {statusMessage}
</div>
```

### Keyboard Navigation

- All interactive elements must be keyboard accessible
- Provide visible focus indicators
- Support Tab, Enter, Escape, and Arrow keys where appropriate
- Ensure logical tab order
- Implement keyboard traps for modals

---

## Forms

### Input Best Practices

- Inputs need `autocomplete` and meaningful `name`
- Use correct `type` (`email`, `tel`, `url`, `number`) and `inputmode`
- Never block paste (`onPaste` + `preventDefault`)
- Labels clickable (`htmlFor` or wrapping control)
- Disable spellcheck on emails, codes, usernames (`spellCheck={false}`)

Example:
```tsx
<form>
  <label htmlFor="email">Email</label>
  <input
    id="email"
    type="email"
    name="email"
    autoComplete="email"
    inputMode="email"
    spellCheck={false}
    placeholder="you@example.com"
    required
  />
</form>
```

### Checkboxes and Radios

- Label + control share single hit target (no dead zones)
- Use proper grouping with `<fieldset>` and `<legend>`

Example:
```tsx
// Correct: label wraps control for single hit target
<label className="flex items-center gap-2 cursor-pointer">
  <input type="checkbox" name="subscribe" value="yes" />
  <span>Subscribe to newsletter</span>
</label>

// Grouping with fieldset
<fieldset>
  <legend>Notification preferences</legend>
  <label>
    <input type="radio" name="notifications" value="email" />
    Email
  </label>
  <label>
    <input type="radio" name="notifications" value="sms" />
    SMS
  </label>
</fieldset>
```

### Form Validation

- Submit button stays enabled until request starts; spinner during request
- Errors inline next to fields; focus first error on submit
- Placeholders end with `…` and show example pattern
- `autocomplete="off"` on non-auth fields to avoid password manager triggers
- Warn before navigation with unsaved changes

Example:
```tsx
function ContactForm() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [errors, setErrors] = useState<Record<string, string>>({})

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    const newErrors = validate(formData)
    setErrors(newErrors)

    if (Object.keys(newErrors).length > 0) {
      // Focus first error
      const firstErrorField = document.getElementById(Object.keys(newErrors)[0])
      firstErrorField?.focus()
      return
    }

    setIsSubmitting(true)
    try {
      await submitForm(formData)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      <div>
        <label htmlFor="email">Email</label>
        <input
          id="email"
          type="email"
          placeholder="you@example.com…"
          aria-invalid={errors.email ? 'true' : 'false'}
          aria-describedby={errors.email ? 'email-error' : undefined}
        />
        {errors.email && (
          <p id="email-error" className="error">
            {errors.email}
          </p>
        )}
      </div>
      <button type="submit" disabled={isSubmitting}>
        {isSubmitting ? <Spinner /> : 'Submit'}
      </button>
    </form>
  )
}
```

---

## Animation

### Performance Guidelines

- Honor `prefers-reduced-motion` (provide reduced variant or disable)
- Animate `transform`/`opacity` only (compositor-friendly)
- Never `transition: all`—list properties explicitly
- Set correct `transform-origin`
- SVG: transforms on `<g>` wrapper with `transform-box: fill-box; transform-origin: center`
- Animations interruptible—respond to user input mid-animation

Example:
```tsx
// Correct: explicit properties with reduced-motion support
const buttonVariants = cva({
  base: 'transition-transform duration-200',
  variants: {
    hover: {
      true: 'hover:scale-105'
    }
  }
})

function AnimatedButton() {
  return (
    <button className={buttonVariants({ hover: true })}>
      Click me
    </button>
  )
}

// Reduced motion query
@media (prefers-reduced-motion: reduce) {
  * {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}
```

---

## Typography

### Best Practices

- Use ellipsis (`…`) not three dots (`...`)
- Use curly quotes (`""`) not straight quotes (`""`)
- Non-breaking spaces for: `10 MB`, `⌘ K`, brand names
- Loading states end with ellipsis: `"Loading…"`, `"Saving…"`
- `font-variant-numeric: tabular-nums` for number columns/comparisons
- Use `text-wrap: balance` or `text-wrap: pretty` on headings (prevents widows)

Example:
```tsx
function Typography() {
  return (
    <div>
      <h1 className="text-wrap-balance">
        Headline that should not have widows
      </h1>
      <p>Loading…</p>
      <table>
        <td className="font-variant-numeric: tabular-nums">
          1,234,567
        </td>
      </table>
    </div>
  )
}
```

---

## Content Handling

### Text Containers

- Text containers handle long content: `truncate`, `line-clamp-*`, or `break-words`
- Flex children need `min-w-0` to allow text truncation
- Handle empty states—don't render broken UI for empty strings/arrays
- User-generated content: anticipate short, average, and very long inputs

Example:
```tsx
// Text truncation in flex container
<div className="flex min-w-0">
  <span className="truncate">{longTitle}</span>
</div>

// Line clamping
<p className="line-clamp-3">
  {longDescription}
</p>

// Empty state handling
function PostList({ posts }: { posts: Post[] }) {
  if (posts.length === 0) {
    return <EmptyState message="No posts yet" />
  }
  return posts.map(post => <PostCard key={post.id} post={post} />)
}
```

---

## Images

### Best Practices

- `<img>` needs explicit `width` and `height` (prevents CLS)
- Below-fold images: `loading="lazy"`
- Above-fold critical images: `priority` or `fetchpriority="high"`

Example:
```tsx
// Next.js Image component
import Image from 'next/image'

// Above-fold: priority
function Hero() {
  return (
    <Image
      src="/hero.jpg"
      alt="Hero image"
      width={1200}
      height={600}
      priority
    />
  )
}

// Below-fold: lazy loading
function Gallery() {
  return (
    <Image
      src="/photo.jpg"
      alt="Gallery photo"
      width={800}
      height={600}
      loading="lazy"
    />
  )
}
```

---

## Performance

### Optimization Guidelines

- Large lists (>50 items): virtualize (`virtua`, `content-visibility: auto`)
- No layout reads in render (`getBoundingClientRect`, `offsetHeight`, `offsetWidth`, `scrollTop`)
- Batch DOM reads/writes; avoid interleaving
- Prefer uncontrolled inputs; controlled inputs must be cheap per keystroke
- Add `<link rel="preconnect">` for CDN/asset domains
- Critical fonts: `<link rel="preload">` with `font-display: swap`

Example:
```tsx
// Virtual list for large datasets
import { useVirtualizer } from '@tanstack/react-virtual'

function VirtualList({ items }: { items: Item[] }) {
  const parentRef = useRef<HTMLDivElement>(null)

  const virtualizer = useVirtualizer({
    count: items.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 50,
  })

  return (
    <div ref={parentRef} className="h-96 overflow-auto">
      <div style={{ height: `${virtualizer.getTotalSize()}px` }}>
        {virtualizer.getVirtualItems().map((virtualItem) => (
          <div
            key={virtualItem.key}
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              width: '100%',
              height: `${virtualItem.size}px`,
              transform: `translateY(${virtualItem.start}px)`,
            }}
          >
            {items[virtualItem.index].name}
          </div>
        ))}
      </div>
    </div>
  )
}

// Preconnect for CDN domains
export default function RootLayout() {
  return (
    <html>
      <head>
        <link rel="preconnect" href="https://cdn.example.com" />
      </head>
      <body>{children}</body>
    </html>
  )
}
```

---

## Navigation & State

### URL State Management

- URL reflects state—filters, tabs, pagination, expanded panels in query params
- Links use `<Link>`/`<a>` (Cmd/Ctrl+click, middle-click support)
- Deep-link all stateful UI (if uses `useState`, consider URL sync via nuqs or similar)
- Destructive actions need confirmation modal or undo window—never immediate

Example:
```tsx
// URL state with nuqs
import { useQueryState } from 'nuqs'

function ProductList() {
  const [category, setCategory] = useQueryState('category')
  const [page, setPage] = useQueryState('page', { defaultValue: '1' })

  return (
    <div>
      <select value={category} onChange={(e) => setCategory(e.target.value)}>
        <option value="">All</option>
        <option value="electronics">Electronics</option>
      </select>
      <ProductGrid category={category} page={page} />
      <Link href={`?category=${category}&page=${Number(page) + 1}`}>
        Next page
      </Link>
    </div>
  )
}

// Destructive action with confirmation
function DeleteButton({ id }: { id: string }) {
  const [showConfirm, setShowConfirm] = useState(false)

  const handleDelete = async () => {
    await deleteItem(id)
    setShowConfirm(false)
  }

  return (
    <>
      <button onClick={() => setShowConfirm(true)}>Delete</button>
      {showConfirm && (
        <ConfirmDialog
          message="Are you sure you want to delete this item?"
          onConfirm={handleDelete}
          onCancel={() => setShowConfirm(false)}
        />
      )}
    </>
  )
}
```

---

## Touch & Interaction

### Touch Guidelines

- `touch-action: manipulation` (prevents double-tap zoom delay)
- `-webkit-tap-highlight-color` set intentionally
- `overscroll-behavior: contain` in modals/drawers/sheets
- During drag: disable text selection, `inert` on dragged elements
- `autoFocus` sparingly—desktop only, single primary input; avoid on mobile

Example:
```tsx
// Touch-friendly button
const button = cva({
  base: 'touch-action-manipulation -webkit-tap-highlight-color-transparent'
})

// Modal with overscroll containment
function Modal() {
  return (
    <div className="overscroll-behavior-contain">
      <div className="backdrop">...</div>
      <div className="content">...</div>
    </div>
  )
}
```

---

## Safe Areas & Layout

### Layout Guidelines

- Full-bleed layouts need `env(safe-area-inset-*)` for notches
- Avoid unwanted scrollbars: `overflow-x-hidden` on containers, fix content overflow
- Flex/grid over JS measurement for layout

Example:
```tsx
// Safe area handling for notched devices
function FullBleedLayout() {
  return (
    <div
      style={{
        paddingTop: 'env(safe-area-inset-top)',
        paddingBottom: 'env(safe-area-inset-bottom)',
        paddingLeft: 'env(safe-area-inset-left)',
        paddingRight: 'env(safe-area-inset-right)',
      }}
    >
      {children}
    </div>
  )
}
```

---

## Dark Mode & Theming

### Theme Implementation

- `color-scheme: dark` on `<html>` for dark themes (fixes scrollbar, inputs)
- `<body>` matches page background

Example:
```tsx
// Dark mode with color-scheme
export default function RootLayout() {
  return (
    <html className="dark" style={{ colorScheme: 'dark' }}>
      <body className="bg-gray-950 text-gray-50">
        {children}
      </body>
    </html>
  )
}
```

---

## Review Checklist

When reviewing UI code, check for:

### HTML & Structure
- [ ] Semantic HTML5 elements used appropriately
- [ ] Proper heading hierarchy (no skipped levels)
- [ ] Skip link for main content included
- [ ] `scroll-margin-top` on heading anchors

### Accessibility
- [ ] Visible focus states on all interactive elements
- [ ] `:focus-visible` used instead of `:focus`
- [ ] ARIA labels on icon-only buttons
- [ ] Keyboard navigation implemented
- [ ] Screen reader optimization applied

### Forms
- [ ] Inputs have `autocomplete` and meaningful `name`
- [ ] Correct `type` and `inputmode` used
- [ ] Paste not blocked
- [ ] Labels clickable (wrapping control or `htmlFor`)
- [ ] Spellcheck disabled on appropriate fields
- [ ] Checkboxes/radios have single hit target
- [ ] Inline error messages with focus management
- [ ] Placeholders end with `…`

### Animation
- [ ] `prefers-reduced-motion` honored
- [ ] Only `transform`/`opacity` animated
- [ ] Properties listed explicitly (no `transition: all`)
- [ ] Animations interruptible

### Typography
- [ ] Ellipsis (`…`) used instead of three dots
- [ ] Curly quotes used
- [ ] Non-breaking spaces for appropriate content
- [ ] Loading states end with `…`
- [ ] `tabular-nums` for number columns
- [ ] `text-wrap: balance` on headings

### Content
- [ ] Text containers handle overflow (`truncate`, `line-clamp`, `break-words`)
- [ ] Flex children have `min-w-0` for truncation
- [ ] Empty states handled gracefully

### Images
- [ ] Explicit `width` and `height` on images
- [ ] Lazy loading on below-fold images
- [ ] Priority on above-fold critical images

### Performance
- [ ] Large lists virtualized
- [ ] No layout reads in render
- [ ] DOM reads/writes batched
- [ ] Preconnect tags for CDN domains
- [ ] Font preloading for critical fonts

### Navigation
- [ ] URL state reflected in query params
- [ ] Links use proper anchor tags
- [ ] Deep linking implemented
- [ ] Destructive actions have confirmation

### Touch
- [ ] `touch-action: manipulation` applied
- [ ] `overscroll-behavior: contain` in modals
- [ ] `autoFocus` used sparingly

### Layout
- [ ] Safe area insets handled for notched devices
- [ ] Unwanted scrollbars prevented
- [ ] Flex/grid used over JS measurement

### Theming
- [ ] `color-scheme: dark` on html element for dark themes
- [ ] Body background matches page background

---

## Resources

- Official Repository: https://github.com/vercel-labs/web-interface-guidelines
- Related: WAI-ARIA Authoring Practices: https://www.w3.org/WAI/ARIA/apg/
- Related: WCAG 2.2 Quick Reference: https://www.w3.org/WAI/WCAG22/quickref/

---

Version: 1.0.0
Last Updated: 2026-01-15
Source: Vercel Labs Web Interface Guidelines
