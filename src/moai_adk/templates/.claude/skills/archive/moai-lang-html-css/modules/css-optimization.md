# HTML5 & CSS3 Performance Optimization

_Last updated: 2025-11-22_

## CSS Performance Optimization

### 1. Critical CSS Pattern
```html
<!-- ✅ Inline critical CSS -->
<head>
    <style>
        /* Above-the-fold styles only */
        body { font-family: Arial, sans-serif; }
        .hero { background: linear-gradient(...); }
        .header { display: flex; }
    </style>
    <!-- Load remaining CSS asynchronously -->
    <link rel="preload" href="styles.css" as="style" onload="this.onload=null;this.rel='stylesheet'">
    <noscript><link rel="stylesheet" href="styles.css"></noscript>
</head>
```

### 2. CSS Variables for Design Tokens
```css
/* ✅ Define design tokens once -->
:root {
    --color-primary: #007BFF;
    --color-secondary: #6C757D;
    --spacing-unit: 8px;
    --font-size-base: 16px;
    --transition-fast: 150ms;
}

/* ✅ Easy theming -->
body.dark-mode {
    --color-primary: #0D6EFD;
    --bg-color: #1a1a1a;
}

/* ✅ Use throughout -->
button {
    background-color: var(--color-primary);
    padding: var(--spacing-unit);
    transition: background-color var(--transition-fast);
}
```

### 3. GPU Acceleration
```css
/* ✅ Enable GPU acceleration */
.animated-element {
    transform: translateZ(0);
    will-change: transform;
}

/* ✅ Use transform instead of position -->
.slide {
    /* GOOD: GPU accelerated */
    animation: slide 1s;
}
@keyframes slide {
    from { transform: translateX(0); }
    to { transform: translateX(100px); }
}

/* ✅ Optimize animations */
.fade {
    animation: fade-in 0.3s ease-out;
}
@keyframes fade-in {
    from { opacity: 0; }
    to { opacity: 1; }
}
```

### 4. Responsive Images
```html
<!-- ✅ Responsive images with srcset -->
<picture>
    <source
        media="(max-width: 600px)"
        srcset="image-small.webp"
        type="image/webp"
    >
    <source
        media="(max-width: 1200px)"
        srcset="image-medium.webp"
        type="image/webp"
    >
    <img src="image-large.jpg" alt="Description">
</picture>

<!-- ✅ Lazy loading -->
<img
    src="image.jpg"
    alt="Description"
    loading="lazy"
>
```

## CSS Architecture

### BEM Methodology
```css
/* ✅ BEM: Block Element Modifier -->

/* Block: component itself */
.card { padding: 20px; }

/* Element: part of block */
.card__title { font-size: 24px; }
.card__content { margin-top: 10px; }

/* Modifier: variation -->
.card--featured { background: gold; }
.card__title--large { font-size: 32px; }
```

### Utility-First CSS (Tailwind Pattern)
```html
<!-- ✅ Compose styles from utilities -->
<div class="flex flex-col gap-4 p-6 bg-white rounded-lg shadow-md">
    <h2 class="text-2xl font-bold text-gray-900">Title</h2>
    <p class="text-gray-600">Content</p>
    <button class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">
        Action
    </button>
</div>
```

## Animation Best Practices

```css
/* ✅ Efficient animations -->
.loading {
    animation: spin 1s linear infinite;
}

@keyframes spin {
    to { transform: rotate(360deg); }
}

/* ✅ Avoid expensive properties -->
/* DON'T animate: box-shadow, border-radius, width/height -->
/* DO animate: transform, opacity */

.slide-in {
    animation: slide-in 0.3s ease-out forwards;
}

@keyframes slide-in {
    from {
        transform: translateX(-20px);
        opacity: 0;
    }
    to {
        transform: translateX(0);
        opacity: 1;
    }
}
```

## CSS Containment for Performance

```css
/* ✅ Contain layout changes -->
.widget {
    contain: layout style paint;
    /* Prevents affecting rest of page */
}

/* ✅ Content visibility -->
.off-screen {
    content-visibility: auto;
    /* Skips rendering until needed */
}
```

## Unused CSS Removal

```javascript
// Tools to identify unused CSS
// PurgeCSS, UnCSS, coverage tools in DevTools

// In webpack.config.js
const PurgecssPlugin = require('purgecss-webpack-plugin');

module.exports = {
    plugins: [
        new PurgecssPlugin({
            paths: glob.sync(`${path.join(__dirname, 'src')}/**/*`, {
                nodir: true,
            }),
        }),
    ],
};
```

## Dark Mode Implementation

```css
/* ✅ System preference -->
@media (prefers-color-scheme: dark) {
    :root {
        --color-text: #FFFFFF;
        --color-bg: #1a1a1a;
    }
}

/* ✅ User preference toggle -->
body.dark-mode {
    --color-text: #FFFFFF;
    --color-bg: #1a1a1a;
}

body.light-mode {
    --color-text: #000000;
    --color-bg: #FFFFFF;
}

/* ✅ Use variables -->
body {
    color: var(--color-text);
    background-color: var(--color-bg);
}
```

## Print Styles

```css
/* ✅ Optimize for printing -->
@media print {
    body { font-size: 12pt; }
    a { text-decoration: none; }
    a[href]:after { content: " (" attr(href) ")"; }
    .no-print { display: none; }
    h1, h2 { page-break-after: avoid; }
    img { max-width: 100%; }
}
```

---

**Last Updated**: 2025-11-22
**Related**: moai-lang-html-css/SKILL.md, modules/accessibility.md

