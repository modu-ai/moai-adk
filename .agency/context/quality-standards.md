<!-- 
  Agency Brand Context: Quality Standards
  
  This file is part of the Agency Brand Context (constitution).
  It is NOT auto-modified by the learner agent.
  Changes require manual editing by the user.
  It serves as the foundation for all agency skill evolution.
  
  Fill this file during the first `/agency brief` run via client interview.
-->

# Quality Standards

## Performance Targets

### Core Web Vitals

LCP (Largest Contentful Paint): _TBD_
<!-- Example: < 2.5s -->

INP (Interaction to Next Paint): _TBD_
<!-- Example: < 200ms -->

CLS (Cumulative Layout Shift): _TBD_
<!-- Example: < 0.1 -->

### Lighthouse Scores

Performance: _TBD_
<!-- Example: >= 90 -->

Accessibility: _TBD_
<!-- Example: >= 95 -->

Best Practices: _TBD_
<!-- Example: >= 90 -->

SEO: _TBD_
<!-- Example: >= 95 -->

### Bundle Size Budgets

Initial JS: _TBD_
<!-- Example: < 150KB gzipped -->

Per-Route JS: _TBD_
<!-- Example: < 50KB gzipped -->

Total CSS: _TBD_
<!-- Example: < 30KB gzipped -->

## Accessibility Requirements

WCAG Level: _TBD_
<!-- Example: WCAG 2.2 AA -->

Key Requirements:
- _TBD_
- _TBD_
- _TBD_
- _TBD_

<!-- Example:
- All interactive elements must be keyboard accessible
- Color contrast ratio minimum 4.5:1 for normal text, 3:1 for large text
- All images must have descriptive alt text
- Focus indicators must be visible and consistent
- Screen reader tested with VoiceOver (macOS) and NVDA (Windows)
- Reduced motion support via prefers-reduced-motion
-->

## SEO Requirements

Meta Tags: _TBD_
<!-- Example: title, description, og:image required on every page -->

Structured Data: _TBD_
<!-- Example: JSON-LD for Organization, Product, FAQ schemas -->

Sitemap: _TBD_
<!-- Example: auto-generated sitemap.xml, submitted to Google Search Console -->

Robots: _TBD_
<!-- Example: allow all public pages, block /admin/ and /api/ -->

Canonical URLs: _TBD_
<!-- Example: self-referencing canonical on every page -->

Page Speed: _TBD_
<!-- Example: all pages must load in under 3s on 3G connection -->

## Browser and Device Support

### Browsers

- _TBD_
- _TBD_
- _TBD_
- _TBD_

<!-- Example:
- Chrome (last 2 versions)
- Firefox (last 2 versions)
- Safari (last 2 versions)
- Edge (last 2 versions)
- No IE11 support
-->

### Devices

- _TBD_
- _TBD_
- _TBD_

<!-- Example:
- Mobile: 360px minimum width (iPhone SE)
- Tablet: 768px breakpoint
- Desktop: 1024px-1920px, max-width 1440px for content
-->

### Responsive Breakpoints

- _TBD_

<!-- Example:
- sm: 640px
- md: 768px
- lg: 1024px
- xl: 1280px
- 2xl: 1536px
-->

## Code Quality Standards

### Linting and Formatting

Linter: _TBD_
<!-- Example: ESLint 9 with strict TypeScript rules -->

Formatter: _TBD_
<!-- Example: Prettier with 2-space indent, single quotes, trailing commas -->

Pre-commit Hooks: _TBD_
<!-- Example: Husky + lint-staged for auto-formatting on commit -->

### Code Review

Review Required: _TBD_
<!-- Example: 1 approval required for all PRs to main -->

Auto-merge: _TBD_
<!-- Example: allowed for dependency updates (Dependabot/Renovate) -->

### Naming Conventions

Components: _TBD_
<!-- Example: PascalCase (UserProfile.tsx) -->

Utilities: _TBD_
<!-- Example: camelCase (formatDate.ts) -->

Constants: _TBD_
<!-- Example: UPPER_SNAKE_CASE (MAX_RETRY_COUNT) -->

CSS Classes: _TBD_
<!-- Example: Tailwind utilities, no custom BEM -->

## Testing Requirements

### Coverage Targets

Unit Tests: _TBD_
<!-- Example: >= 80% line coverage -->

Integration Tests: _TBD_
<!-- Example: critical user flows covered -->

E2E Tests: _TBD_
<!-- Example: top 5 user journeys covered with Playwright -->

### Testing Tools

Unit/Integration: _TBD_
<!-- Example: Vitest with React Testing Library -->

E2E: _TBD_
<!-- Example: Playwright -->

Visual Regression: _TBD_
<!-- Example: Chromatic (Storybook) or Percy -->

### Testing Rules

- _TBD_
- _TBD_
- _TBD_

<!-- Example:
- All new components must have at least one test
- All bug fixes must include a regression test
- E2E tests must run in CI before merge to main
- No skipped tests allowed in main branch
-->

## Security Requirements

### Authentication

- _TBD_

<!-- Example:
- All API routes must verify authentication
- Session tokens must expire after 24 hours
- Support MFA for admin accounts
-->

### Data Protection

- _TBD_

<!-- Example:
- All data in transit must use TLS 1.3
- PII must be encrypted at rest
- No sensitive data in URL parameters or client-side logs
-->

### Headers and Policies

- _TBD_

<!-- Example:
- Content-Security-Policy configured per environment
- X-Frame-Options: DENY
- Strict-Transport-Security enabled
- CORS restricted to known origins
-->

### Dependency Security

- _TBD_

<!-- Example:
- Automated vulnerability scanning via GitHub Dependabot
- No known critical vulnerabilities in production dependencies
- Lock file (pnpm-lock.yaml) committed and reviewed
-->

### Compliance

- _TBD_

<!-- Example:
- GDPR compliant (cookie consent, data deletion on request)
- SOC 2 Type II (if applicable)
- Privacy policy and terms of service pages required
-->
