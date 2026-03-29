# AI-Assisted Frontend Patterns

Patterns for leveraging AI models in frontend development workflows, derived from OpenAI's GPT-5.4 frontend research (Mar 2026).

## Visual Reference Strategy

Provide reference screenshots or mood boards to help the model infer layout rhythm and typography scale. Generate a mood board or several visual options before selecting final assets.

Guide toward strong visual references by explicitly describing attributes: style, color palette, composition, and mood.

## Playwright Verification

Use Playwright for iterative inspection, validation, and refinement of AI-generated implementations:

- Inspect rendered pages across multiple viewports
- Navigate application flows to detect state issues
- Verify visual alignment with reference UI
- Validate functional completeness of generated components

Providing a Playwright tool significantly improves the likelihood of polished, functionally complete interfaces.

## React Patterns for AI-Generated Code

Modern React patterns to prefer in AI-assisted workflows:

- useEffectEvent: Prefer over useEffect with dependency workarounds
- startTransition: Use for non-urgent state updates to keep UI responsive
- useDeferredValue: Use for expensive renders that can lag behind user input
- Do not add useMemo/useCallback by default — follow the repo's React Compiler guidance

## Motion Design with Framer Motion

Framer Motion is the preferred animation library for React projects. Apply motion intentionally:

- Entrance sequences: Stagger children, fade-in with translate
- Scroll-linked effects: useScroll, useTransform for parallax
- Hover and reveal: whileHover, whileTap for micro-interactions
- Layout animations: layout prop for smooth reflows

Ship at least 2-3 intentional motions for visually-led work. Use motion to create presence and hierarchy, not noise.

## Reasoning Level Strategy

For simpler frontend tasks, lower reasoning levels often produce stronger results. The model stays fast, focused, and less prone to overthinking layout decisions.

Use higher reasoning for:
- Complex state management logic
- Multi-step form validation
- Architectural decisions (routing, data flow)

Use lower reasoning for:
- Landing page layouts
- Marketing pages
- Simple component styling
- Static content pages

## Content-First Development

Provide real copy, product context, or a clear project goal to improve frontend results. Generic placeholder text leads to generic-looking interfaces.

For product UI, default to utility copy:
- Prioritize orientation, status, and action
- Section headings describe what the area is or what the user can do
- Start with the working surface: KPIs, charts, filters, tables
