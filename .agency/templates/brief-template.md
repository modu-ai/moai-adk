# BRIEF: {{BRIEF_TITLE}}

## 1. Metadata

| Field | Value |
|-------|-------|
| **ID** | BRIEF-YYYYMMDD-NNN |
| **Created** | YYYY-MM-DD |
| **Status** | draft / approved / in-progress / completed / archived |
| **Type** | landing-page / multi-page-site / web-app / component-library / redesign |
| **Priority** | low / medium / high / urgent |
| **Pipeline Mode** | default / step-by-step / team |
| **Framework** | next.js / react / vue / static |

---

## 2. Project Goal

### Business Objective

What is the business problem this project solves? What drives the need for this work?

> [Describe the business motivation and strategic context]

### Target Outcome

What does success look like when this project is delivered?

> [Describe the desired end state in concrete terms]

### Key Performance Indicators

How will success be measured? Define 2-4 measurable KPIs.

| KPI | Target | Measurement Method |
|-----|--------|--------------------|
| [e.g., Conversion Rate] | [e.g., > 3%] | [e.g., Analytics funnel tracking] |
| [e.g., Time on Page] | [e.g., > 2 min] | [e.g., Session analytics] |
| [e.g., Bounce Rate] | [e.g., < 40%] | [e.g., Page analytics] |

---

## 3. Target Audience

### Primary Persona

| Attribute | Description |
|-----------|-------------|
| **Name** | [Persona name] |
| **Role/Title** | [Job title or life role] |
| **Age Range** | [e.g., 25-40] |
| **Goals** | [What they want to achieve] |
| **Pain Points** | [What frustrates them today] |
| **Tech Comfort** | [low / medium / high] |
| **Decision Drivers** | [What influences their choices] |

### Secondary Persona (if applicable)

| Attribute | Description |
|-----------|-------------|
| **Name** | [Persona name] |
| **Role/Title** | [Job title or life role] |
| **Relationship to Primary** | [How they relate to the primary persona] |
| **Key Differences** | [How their needs differ] |

---

## 4. Brand and Tone

### Brand Voice

| Dimension | Position | Notes |
|-----------|----------|-------|
| **Formality** | casual / balanced / formal | |
| **Emotion** | reserved / warm / enthusiastic | |
| **Authority** | peer / advisor / expert | |
| **Humor** | none / subtle / playful | |
| **Pace** | deliberate / moderate / energetic | |

### Brand Keywords

Words that define the brand personality (pick 3-5):

> [e.g., innovative, trustworthy, human-centered, bold, precise]

### Color Palette

| Role | Color | Hex | Usage |
|------|-------|-----|-------|
| Primary | [Name] | #XXXXXX | [Main brand color, CTAs, key elements] |
| Secondary | [Name] | #XXXXXX | [Supporting elements, accents] |
| Accent | [Name] | #XXXXXX | [Highlights, interactive states] |
| Background | [Name] | #XXXXXX | [Page backgrounds] |
| Text | [Name] | #XXXXXX | [Body text] |

### Typography

| Role | Font | Weight | Usage |
|------|------|--------|-------|
| Heading | [Font name] | [e.g., 700] | [Headlines, section titles] |
| Body | [Font name] | [e.g., 400] | [Paragraph text, descriptions] |
| Accent | [Font name] | [e.g., 500] | [CTAs, labels, navigation] |

### Reference Sites

Sites that capture the desired look, feel, or functionality:

1. [URL] - What to reference: [specific element or feeling]
2. [URL] - What to reference: [specific element or feeling]
3. [URL] - What to reference: [specific element or feeling]

---

## 5. Content Requirements

### Pages / Sections

| # | Page/Section | Purpose | Priority |
|---|-------------|---------|----------|
| 1 | [e.g., Hero] | [Primary value proposition] | must-have |
| 2 | [e.g., Features] | [Product capabilities] | must-have |
| 3 | [e.g., Social Proof] | [Trust building] | should-have |
| 4 | [e.g., Pricing] | [Conversion] | must-have |
| 5 | [e.g., FAQ] | [Objection handling] | nice-to-have |
| 6 | [e.g., Footer] | [Navigation, legal] | must-have |

### Key Messages

The 3-5 core messages the audience must take away:

1. [Primary message - the one thing they must remember]
2. [Supporting message - builds credibility]
3. [Differentiator - why choose us over alternatives]
4. [Social proof - evidence of value]
5. [Call to action - the next step]

### Call-to-Action Strategy

| CTA | Location | Type | Copy |
|-----|----------|------|------|
| Primary CTA | [e.g., Hero, Pricing] | [button / link / form] | [e.g., "Start Free Trial"] |
| Secondary CTA | [e.g., Features] | [button / link] | [e.g., "Learn More"] |
| Tertiary CTA | [e.g., Footer] | [link] | [e.g., "Contact Sales"] |

---

## 6. Technical Constraints

### Platform

| Constraint | Requirement |
|------------|-------------|
| **Target Devices** | [desktop / mobile / both] |
| **Browser Support** | [modern only / IE11+ / specific browsers] |
| **Accessibility** | [WCAG 2.1 AA / AAA / none specified] |
| **Performance Budget** | [e.g., LCP < 2.5s, FCP < 1.8s] |
| **SEO Requirements** | [e.g., meta tags, structured data, sitemap] |

### Framework and Stack

| Layer | Technology | Version | Notes |
|-------|-----------|---------|-------|
| Framework | [e.g., Next.js] | [e.g., 15] | [App Router / Pages Router] |
| Styling | [e.g., Tailwind CSS] | [e.g., 4.0] | [With design tokens] |
| Components | [e.g., shadcn/ui] | [latest] | [Specific components needed] |
| Animation | [e.g., Framer Motion] | [latest] | [Scroll animations, transitions] |
| Icons | [e.g., Lucide] | [latest] | [Specific icon set] |

### Integrations

| Service | Purpose | Required |
|---------|---------|----------|
| [e.g., Analytics] | [e.g., Google Analytics 4] | yes / no |
| [e.g., CMS] | [e.g., Contentful] | yes / no |
| [e.g., Email] | [e.g., Resend] | yes / no |
| [e.g., Auth] | [e.g., Clerk] | yes / no |

### Hosting

| Aspect | Requirement |
|--------|-------------|
| **Platform** | [e.g., Vercel / Netlify / AWS] |
| **Region** | [e.g., US-East / Global CDN] |
| **Custom Domain** | [e.g., example.com] |
| **SSL** | [required / not required] |

---

## 7. Deliverables

Checklist of expected outputs from the pipeline:

- [ ] BRIEF document (this file, approved)
- [ ] Copy deck (all page text, headlines, CTAs)
- [ ] Design spec (layout, component hierarchy, design tokens)
- [ ] Working code
  - [ ] Page components
  - [ ] Reusable UI components
  - [ ] Global styles and theme configuration
  - [ ] Responsive layouts (mobile, tablet, desktop)
  - [ ] Animation and interaction implementations
- [ ] Asset inventory (images, icons, fonts referenced)
- [ ] Accessibility compliance report
- [ ] Performance audit results
- [ ] Deployment configuration

---

## 8. Evaluation Criteria

### Must-Pass Criteria

These criteria MUST all pass for the project to be accepted. Failure in any single criterion blocks approval regardless of other scores.

| # | Criterion | Threshold | Measurement |
|---|-----------|-----------|-------------|
| 1 | Brand Consistency | Voice, colors, and typography match brand context | Manual review against brand guidelines |
| 2 | Responsive Design | Functional on mobile (375px), tablet (768px), desktop (1280px) | Visual inspection at each breakpoint |
| 3 | Core Content Present | All must-have pages/sections implemented with copy | Content audit against Section 5 |
| 4 | CTA Functionality | Primary and secondary CTAs are visible and interactive | Click-through testing |
| 5 | No Broken Elements | No visual glitches, broken layouts, or console errors | Browser testing and dev tools audit |
| 6 | Accessibility Baseline | Keyboard navigable, sufficient color contrast, alt text present | Lighthouse accessibility score >= 90 |

### Nice-to-Have Criteria

These criteria improve the score but are not blocking. They contribute to the overall quality assessment.

| # | Criterion | Weight | Measurement |
|---|-----------|--------|-------------|
| 1 | Animation Quality | 0.15 | Smooth transitions, purposeful motion, no jank |
| 2 | Copy Excellence | 0.15 | Compelling, on-brand, free of errors |
| 3 | Performance | 0.15 | Lighthouse performance score >= 90 |
| 4 | Code Quality | 0.15 | Clean components, proper separation, no duplication |
| 5 | SEO Readiness | 0.10 | Meta tags, semantic HTML, structured data |
| 6 | Visual Polish | 0.15 | Spacing consistency, visual hierarchy, micro-interactions |
| 7 | Beyond Brief | 0.15 | Thoughtful additions that enhance the project without scope creep |

---

## Appendix: Pipeline Progress

This section is auto-populated during pipeline execution.

| Phase | Status | Agent | Started | Completed | Notes |
|-------|--------|-------|---------|-----------|-------|
| Planner | pending | planner | - | - | |
| Copywriter | pending | copywriter | - | - | |
| Designer | pending | designer | - | - | |
| Builder | pending | builder | - | - | |
| Evaluator | pending | evaluator | - | - | |
| Learner | pending | learner | - | - | |

### GAN Loop Log

| Iteration | Score | Pass | Key Feedback |
|-----------|-------|------|-------------|
| - | - | - | - |

### Final Score

| Category | Score | Status |
|----------|-------|--------|
| Must-Pass | -/6 | pending |
| Nice-to-Have | 0.00 | pending |
| **Overall** | **0.00** | **pending** |
