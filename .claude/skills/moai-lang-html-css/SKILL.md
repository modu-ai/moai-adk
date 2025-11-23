---
name: moai-lang-html-css
description: HTML5 Semantic Markup & CSS3 with Accessibility and Responsive Design
version: 1.0.0
modularized: false
tags:
  - enterprise
  - css
  - html
  - development
  - programming-language
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: lang, css, moai, html  


## Quick Reference (30 seconds)

# HTML5 & CSS3 — Enterprise Web Markup & Styling

**Primary Focus**: HTML5 semantics, CSS3 layout/animation, accessibility (WCAG 2.1 AA)
**Best For**: Semantic markup, responsive design, accessible web applications
**Key Standards**: HTML5, CSS3, WCAG 2.1 AA, CSS Grid, Flexbox
**Auto-triggers**: HTML files, CSS files, web development, accessibility

| Standard | Version | Status |
|----------|---------|--------|
| HTML5 | 2024 | Active |
| CSS3 | Living Standard | Current |
| WCAG | 2.1 AA | Compliance |
| ES2024 | Latest | Modern |

---

## What It Does

HTML5 semantic markup and CSS3 styling with accessibility-first approach. WCAG 2.1 AA compliance, responsive design, and modern CSS features (Grid, Flexbox, Custom Properties).

**Key capabilities**:
- ✅ HTML5 semantic elements and ARIA labels
- ✅ CSS3 layout (Flexbox, CSS Grid, subgrid)
- ✅ Responsive design (mobile-first)
- ✅ WCAG 2.1 AA accessibility compliance
- ✅ CSS animations and transitions
- ✅ Custom properties and design tokens
- ✅ Cross-browser compatibility

---

## When to Use

**Automatic triggers**:
- HTML (.html, .htm, .ejs) files
- CSS (.css, .scss, .less) files
- Web layout and styling
- Accessibility audits

**Manual invocation**:
- Design semantic markup structure
- Implement responsive layouts
- Review accessibility compliance
- Optimize CSS performance

---

## Three-Level Learning Path

### Level 1: Fundamentals (See examples.md)

Core HTML5 & CSS3 concepts:
- **Semantic HTML**: Article, section, nav, header, footer, main
- **Forms**: Input types, labels, validation, accessibility
- **CSS Layout**: Flexbox basics, Grid basics, responsive units
- **Accessibility**: ARIA, focus states, color contrast
- **Responsive**: Media queries, mobile-first, viewport meta

### Level 2: Advanced Patterns (See modules/accessibility.md)

Production-ready enterprise patterns:
- **WCAG 2.1 AA Compliance**: Color contrast, keyboard navigation, screen readers
- **Advanced Layouts**: CSS Grid advanced, subgrid, container queries
- **Component Patterns**: Button states, form accessibility, accessible tables
- **ARIA Patterns**: Live regions, landmarks, custom widgets
- **Testing**: Accessibility auditing, keyboard testing, screen reader testing

### Level 3: Performance & Optimization (See modules/css-optimization.md)

Production optimization:
- **CSS Performance**: Minimization, critical CSS, unused CSS removal
- **Animation Performance**: GPU acceleration, will-change, containment
- **Design Tokens**: CSS variables, theming, dark mode
- **Cross-Browser**: Polyfills, fallbacks, feature detection
- **Bundle Optimization**: Purging unused CSS, tree-shaking

---

## Best Practices

✅ **DO**:
- Use semantic HTML elements
- Test with accessibility tools (axe, Lighthouse)
- Implement keyboard navigation
- Maintain 4.5:1 color contrast
- Use CSS Grid and Flexbox
- Define custom properties for design tokens
- Validate HTML and CSS

❌ **DON'T**:
- Use div for buttons/links (semantic violations)
- Ignore focus indicators
- Rely on color alone for information
- Use inline styles
- Create keyboard traps
- Use autoplaying audio/video
- Ignore ARIA warnings

---

## Tool Versions (2025-11-22)

| Tool | Version | Purpose |
|------|---------|---------|
| **HTML5** | 2024 | Markup |
| **CSS3** | Living | Styling |
| **Tailwind** | 4.0 | Utility CSS |
| **Sass** | 1.83 | CSS preprocessing |
| **PostCSS** | 8.4 | CSS transformation |

---

## Installation & Setup

```bash
# Validate HTML
npm install -D htmlhint

# Validate CSS
npm install -D stylelint

# Accessibility auditing
npm install -D axe-core

# CSS preprocessing (Sass)
npm install -D sass
```

---

## Works Well With

- `moai-domain-frontend` (React, Vue components)
- `moai-domain-ux` (UX/UI design principles)
- `moai-essentials-accessibility` (A11y compliance)

---

## Learn More

- **Practical Examples**: See `examples.md` for 20+ real-world patterns
- **Accessibility**: See `modules/accessibility.md` for WCAG compliance
- **CSS Optimization**: See `modules/css-optimization.md` for performance
- **MDN HTML**: https://developer.mozilla.org/en-US/docs/Web/HTML
- **CSS Tricks**: https://css-tricks.com/
- **WCAG 2.1**: https://www.w3.org/WAI/WCAG21/quickref/

---

## Changelog

- **v4.0.0** (2025-11-22): Modularized with accessibility and optimization modules
- **v3.0.0** (2025-11-13): HTML5 semantic refresh, CSS3 modern features
- **v2.0.0** (2025-10-01): WCAG 2.1 AA compliance framework
- **v1.0.0** (2025-03-01): Initial release

---

## Context7 Integration

### Related Libraries & Tools
- [HTML5](/whatwg/html): HTML Living Standard
- [CSS3](/w3c/css): CSS Specifications
- [Tailwind](/tailwindlabs/tailwindcss): Utility-first CSS framework
- [ARIA](/w3c/aria): Accessible Rich Internet Applications

### Official Documentation
- [MDN HTML Reference](https://developer.mozilla.org/en-US/docs/Web/HTML)
- [MDN CSS Reference](https://developer.mozilla.org/en-US/docs/Web/CSS)
- [WCAG 2.1 Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)
- [Tailwind Documentation](https://tailwindcss.com/docs)

---

**Skills**: Skill("moai-domain-frontend"), Skill("moai-essentials-accessibility")
**Auto-loads**: HTML/CSS files and web projects

