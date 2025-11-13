---
name: "moai-lang-html-css"
version: "4.0.0"
status: stable
description: "Enterprise Skill for advanced development"
allowed-tools: "Read, Bash, WebSearch, WebFetch"
---

# moai-lang-html-css

**HTML5 Semantic Markup & CSS Best Practices with Accessibility Focus**

> **Primary Agent**: ui-ux-expert
> **Secondary Agent**: frontend-expert
> **Version**: 1.0.0
> **Keywords**: html, semantic html, html5, css, styling, accessibility, a11y, wcag, responsive, semantics

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

**HTML Semantic Elements** form the foundation of accessible, SEO-friendly web pages:

| Element | Purpose | Use Case |
|---------|---------|----------|
| `<header>` | Page/section header | Logo, navigation, branding |
| `<nav>` | Navigation container | Menu, breadcrumbs, site links |
| `<main>` | Primary content | Main page content (only one per page) |
| `<section>` | Thematic content group | Articles, chapters, topics |
| `<article>` | Self-contained content | Blog posts, news articles, comments |
| `<aside>` | Tangentially related content | Sidebars, ads, related links |
| `<footer>` | Footer section | Copyright, links, metadata |

**CSS Best Practices**:
- ✅ Mobile-first responsive design (min-width breakpoints)
- ✅ CSS custom properties for design tokens
- ✅ Utility-first approach (Tailwind) or BEM methodology
- ✅ Avoid inline styles; use classes
- ✅ Minimize CSS specificity (avoid `!important`)

**WCAG 2.1 AA Accessibility Checklist**:
- 4.5:1 color contrast for text
- Keyboard navigation (Tab, Enter, Escape, Arrow keys)
- Focus indicators (2px solid outline)
- Semantic HTML (no `<div>` for buttons)
- ARIA labels where necessary

---

### Level 2: Practical Implementation (Common Patterns)

#### Pattern 1: Semantic HTML Page Structure

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Accessible Web Page</title>
  <link rel="stylesheet" href="styles.css">
</head>
<body>
  <!-- Page Header -->
  <header role="banner">
    <nav role="navigation" aria-label="Main navigation">
      <a href="/" class="logo">Brand</a>
      <ul>
        <li><a href="/about">About</a></li>
        <li><a href="/services">Services</a></li>
        <li><a href="/contact">Contact</a></li>
      </ul>
    </nav>
  </header>

  <!-- Skip Link (accessibility best practice) -->
  <a href="#main" class="sr-only">Skip to main content</a>

  <!-- Main Content -->
  <main id="main" role="main">
    <!-- Article Section -->
    <article>
      <h1>Article Title</h1>
      <p>Article introduction...</p>
      <section>
        <h2>Section Heading</h2>
        <p>Section content...</p>
      </section>
    </article>

    <!-- Sidebar -->
    <aside aria-label="Related content">
      <h3>Related Articles</h3>
      <ul>
        <li><a href="...">Related article 1</a></li>
        <li><a href="...">Related article 2</a></li>
      </ul>
    </aside>
  </main>

  <!-- Page Footer -->
  <footer role="contentinfo">
    <p>&copy; 2025 Company Name. All rights reserved.</p>
  </footer>
</body>
</html>
```

#### Pattern 2: Accessible Form with Labels

```html
<!-- WCAG 2.1 AA Compliant Form -->
<form method="POST" action="/submit">
  <fieldset>
    <legend>Contact Information</legend>

    <!-- Email Input with Label -->
    <div class="form-group">
      <label for="email">Email Address *</label>
      <input
        id="email"
        type="email"
        name="email"
        required
        aria-required="true"
        aria-describedby="email-help"
      >
      <small id="email-help">We'll never share your email</small>
    </div>

    <!-- Text Area -->
    <div class="form-group">
      <label for="message">Message *</label>
      <textarea
        id="message"
        name="message"
        rows="5"
        required
        aria-required="true"
      ></textarea>
    </div>

    <!-- Error Message (live region) -->
    <div role="alert" aria-live="polite" id="form-error">
      <!-- Error messages inserted here -->
    </div>

    <!-- Submit Button -->
    <button type="submit" class="btn btn-primary">
      Send Message
    </button>
  </fieldset>
</form>
```

#### Pattern 3: Responsive CSS with Custom Properties

```css
/* Design Tokens as CSS Custom Properties */
:root {
  /* Colors */
  --color-primary: #0ea5e9;
  --color-primary-dark: #0284c7;
  --color-secondary: #8b5cf6;
  --color-neutral-900: #111827;
  --color-neutral-500: #6b7280;
  --color-neutral-100: #f3f4f6;

  /* Typography */
  --font-family-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  --font-family-mono: "Monaco", "Menlo", "Ubuntu Mono", monospace;
  --font-size-sm: 0.875rem;
  --font-size-base: 1rem;
  --font-size-lg: 1.125rem;
  --font-size-xl: 1.25rem;
  --font-size-2xl: 1.5rem;
  --font-weight-normal: 400;
  --font-weight-medium: 500;
  --font-weight-bold: 700;

  /* Spacing (8px base unit) */
  --spacing-xs: 0.25rem;
  --spacing-sm: 0.5rem;
  --spacing-md: 1rem;
  --spacing-lg: 1.5rem;
  --spacing-xl: 2rem;
  --spacing-2xl: 3rem;

  /* Breakpoints */
  --breakpoint-sm: 640px;
  --breakpoint-md: 768px;
  --breakpoint-lg: 1024px;
  --breakpoint-xl: 1280px;

  /* Shadows */
  --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 6px rgba(0, 0, 0, 0.1);
  --shadow-lg: 0 10px 15px rgba(0, 0, 0, 0.1);
}

/* Base Styles */
html {
  scroll-behavior: smooth;
}

body {
  font-family: var(--font-family-sans);
  font-size: var(--font-size-base);
  color: var(--color-neutral-900);
  line-height: 1.5;
  -webkit-font-smoothing: antialiased;
}

/* Typography */
h1, h2, h3, h4, h5, h6 {
  font-weight: var(--font-weight-bold);
  line-height: 1.2;
}

h1 {
  font-size: var(--font-size-2xl);
  margin-bottom: var(--spacing-lg);
}

h2 {
  font-size: var(--font-size-xl);
  margin-bottom: var(--spacing-md);
}

p {
  margin-bottom: var(--spacing-md);
}

/* Button Component */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-sm) var(--spacing-md);
  font-weight: var(--font-weight-medium);
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  transition: all 0.2s ease;

  /* Focus indicator (WCAG AA) */
  outline: none;
}

.btn:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background-color: var(--color-primary);
  color: white;
}

.btn-primary:hover {
  background-color: var(--color-primary-dark);
}

/* Form Group */
.form-group {
  margin-bottom: var(--spacing-lg);
}

label {
  display: block;
  font-weight: var(--font-weight-medium);
  margin-bottom: var(--spacing-sm);
}

input, textarea, select {
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
  border: 1px solid var(--color-neutral-300);
  border-radius: 0.375rem;
  font-family: inherit;
  font-size: inherit;
}

input:focus, textarea:focus, select:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(14, 165, 233, 0.1);
}

input:invalid {
  border-color: #ef4444;
}

/* Responsive Grid (Mobile-First) */
.grid {
  display: grid;
  gap: var(--spacing-md);
}

@media (min-width: var(--breakpoint-md)) {
  .grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: var(--breakpoint-lg)) {
  .grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

/* Screen Reader Only (a11y) */
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border-width: 0;
}

/* Focus Visible (keyboard accessibility) */
:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}
```

---

### Level 3: Advanced Patterns (Expert Reference)

#### Advanced Pattern 1: Accessible Modal Dialog

```html
<div role="dialog"
     aria-modal="true"
     aria-labelledby="modal-title"
     aria-describedby="modal-description">
  <div class="modal-overlay" onclick="closeModal()"></div>
  <div class="modal-content">
    <button
      class="modal-close"
      aria-label="Close dialog"
      onclick="closeModal()">
      ✕
    </button>
    <h2 id="modal-title">Dialog Title</h2>
    <p id="modal-description">Dialog description...</p>
    <button class="btn btn-primary">Action</button>
  </div>
</div>

<style>
  [role="dialog"] {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 50;
  }

  .modal-overlay {
    position: absolute;
    inset: 0;
    background-color: rgba(0, 0, 0, 0.5);
  }

  .modal-content {
    position: relative;
    background: white;
    border-radius: 0.5rem;
    padding: 2rem;
    max-width: 500px;
    box-shadow: 0 20px 25px rgba(0, 0, 0, 0.15);
  }

  /* Focus trap for keyboard users */
  [role="dialog"]:focus-within .modal-content {
    outline: 2px solid var(--color-primary);
  }
</style>

<script>
  function closeModal() {
    const dialog = document.querySelector('[role="dialog"]');
    dialog.remove();
    // Restore focus to trigger element
    previousFocusElement?.focus();
  }

  // Focus trap (prevent Tab from leaving modal)
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') closeModal();
  });
</script>
```

#### Advanced Pattern 2: Skip Navigation (A11y Best Practice)

```html
<!-- Place at very beginning of <body> -->
<a href="#main" class="sr-only skip-link">Skip to main content</a>

<style>
  .skip-link {
    position: absolute;
    top: -40px;
    left: 0;
    background: var(--color-primary);
    color: white;
    padding: 0.5rem 1rem;
    text-decoration: none;
    z-index: 100;
  }

  .skip-link:focus {
    top: 0;
  }
</style>
```

---

## 🎯 Best Practices Checklist

### HTML
- ✅ Use semantic HTML5 elements (`<header>`, `<nav>`, `<main>`, `<section>`, `<article>`, `<aside>`, `<footer>`)
- ✅ Proper heading hierarchy (`<h1>` → `<h2>` → `<h3>`, never skip levels)
- ✅ Form labels explicitly connected with `<label for="id">` and `input id="id"`
- ✅ Alternative text for all images: `<img alt="description">`
- ✅ Landmark roles for screen readers: `role="banner"`, `role="main"`, `role="contentinfo"`

### CSS
- ✅ Mobile-first responsive design (start at mobile, add styles for larger screens)
- ✅ Use CSS custom properties for design tokens (colors, spacing, typography)
- ✅ Maintain 4.5:1 color contrast for normal text (WCAG AA)
- ✅ Visible focus indicators for keyboard users (`:focus-visible`, 2px outline)
- ✅ Use `rem` or `em` for sizing (not `px`) for better font scaling

### Accessibility (WCAG 2.1 AA)
- ✅ Color contrast: Text 4.5:1, UI components 3:1 (use WebAIM contrast checker)
- ✅ Keyboard navigation: All functionality accessible via Tab, Enter, Escape, arrow keys
- ✅ Focus visible: Clear focus indicators (2px solid outline preferred)
- ✅ Semantic HTML foundation: Proper heading levels, form labels, alt text
- ✅ ARIA only when semantic HTML insufficient: `aria-label`, `aria-describedby`, `role="alert"`
- ✅ Skip links: Allow skipping repetitive navigation

---

## 📚 Official References

- **W3C HTML5 Spec**: https://html.spec.whatwg.org/
- **WCAG 2.1 Guidelines**: https://www.w3.org/WAI/WCAG21/quickref/
- **MDN Web Docs - HTML**: https://developer.mozilla.org/en-US/docs/Web/HTML
- **MDN Web Docs - CSS**: https://developer.mozilla.org/en-US/docs/Web/CSS
- **WebAIM - Contrast Checker**: https://webaim.org/resources/contrastchecker/
- **WAI-ARIA Authoring Practices**: https://www.w3.org/WAI/ARIA/apg/

---

## 🔗 Related Skills

- `Skill("moai-lang-tailwind-css")` – Utility-first CSS framework
- `Skill("moai-lib-shadcn-ui")` – React components with Tailwind + Radix
- `Skill("moai-domain-frontend")` – Full frontend architecture
