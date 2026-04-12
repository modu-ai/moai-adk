# Frontend Evaluator Profile

UI/UX-focused evaluation with anti-AI-slop criteria for web frontend projects.

## Evaluation Dimensions

| Dimension | Weight | Pass Threshold |
|-----------|--------|----------------|
| Originality | 40% | No generic AI patterns detected |
| Design Quality | 30% | Coherent typography, color, and layout |
| Craft & Functionality | 30% | Accessibility (WCAG AA), responsive, interactive states present |

## Must-Pass Criteria

- WCAG AA compliance (contrast ratio, keyboard navigation, ARIA labels)
- Responsive breakpoints verified (mobile, tablet, desktop)
- All interactive elements have hover/focus/active states

## AI-Slop Detection (Penalize — Originality dimension)

The following patterns indicate generic AI-generated output. Each detected pattern reduces the Originality score. Three or more patterns trigger an automatic Originality FAIL.

- **Stock card layouts**: Card grids using default Bootstrap/Tailwind card component without custom design tokens or intentional layout variation
- **Default utility-only styling**: Entire design expressed solely through Tailwind or Bootstrap utility classes without any custom CSS variables or design tokens
- **Purple/blue gradient backgrounds**: Generic gradient hero sections (e.g., `from-purple-600 to-blue-500`) without brand color alignment or design justification
- **Generic placeholder text**: Literal placeholder strings such as "Lorem ipsum", "Welcome to our platform", "Your title here", "Click here to learn more" left in production code
- **Identical component structure**: Two or more unrelated page sections sharing exactly the same component layout without design rationale
- **Missing interactive states**: Interactive elements (buttons, links, form inputs, cards) that lack hover, focus, or active CSS state definitions

When 3 or more AI-slop patterns are detected: Originality dimension = FAIL.
When 1-2 AI-slop patterns are detected: Originality score capped at 0.50.

## Hard Thresholds

- WCAG AA compliance required (contrast, labels, keyboard nav)
- Responsive breakpoints must be tested (mobile, tablet, desktop)
- No accessibility violations = Craft & Functionality PASS condition

## Evaluation Focus

- Does the design feel intentional and unique?
- Are typography, color, and spacing choices coherent?
- Does the UI work correctly across device sizes?
- Is the interface accessible to users with disabilities?

## Scoring Rubric

### Originality (40%)

| Score | Description |
|-------|-------------|
| 1.00 | Design feels unique and intentional; custom design tokens; no AI-slop patterns detected |
| 0.75 | Minor generic elements present but overall design is distinctive and purposeful |
| 0.50 | 1-2 AI-slop patterns detected; design has some intentionality but generic areas present |
| 0.25 | 3+ AI-slop patterns detected or design is entirely generic (triggers Originality FAIL) |

### Design Quality (30%)

| Score | Description |
|-------|-------------|
| 1.00 | Coherent typographic hierarchy, consistent color application, intentional spacing rhythm |
| 0.75 | Generally coherent design; minor inconsistencies in spacing or color usage |
| 0.50 | Basic visual consistency; typography or color choices feel arbitrary |
| 0.25 | Incoherent visual design; no discernible hierarchy or rhythm |

### Craft & Functionality (30%)

| Score | Description |
|-------|-------------|
| 1.00 | WCAG AA passes; responsive at all breakpoints; all interactive states defined; no console errors |
| 0.75 | WCAG AA passes; minor responsive issues; most interactive states defined |
| 0.50 | Some accessibility violations; responsive issues at one breakpoint; some states missing |
| 0.25 | Multiple WCAG AA failures or broken responsive behavior (triggers Craft & Functionality FAIL) |
