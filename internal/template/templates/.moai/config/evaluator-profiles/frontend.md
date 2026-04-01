# Frontend Evaluator Profile

UI/UX-focused evaluation with anti-AI-slop criteria for web frontend projects.

## Evaluation Weights

| Dimension | Weight | Threshold |
|-----------|--------|-----------|
| Originality | 40% | No generic AI patterns detected |
| Design Quality | 30% | Coherent typography, color, and layout |
| Craft & Functionality | 30% | Accessibility (WCAG), responsive, usable |

## AI-Slop Detection (Penalize)

The following patterns indicate generic AI-generated output and should be penalized:
- Default Bootstrap/Tailwind card layouts without customization
- Generic purple/blue gradient backgrounds
- Stock hero sections with centered text + CTA button
- Identical component spacing without rhythm variation
- Default system fonts without typographic hierarchy
- Generic placeholder imagery descriptions

When AI-slop patterns are detected: Originality dimension = FAIL.

## Hard Thresholds

- WCAG AA compliance required (contrast, labels, keyboard nav)
- Responsive breakpoints must be tested (mobile, tablet, desktop)
- No accessibility violations = Craft PASS condition

## Evaluation Focus

- Does the design feel intentional and unique?
- Are typography, color, and spacing choices coherent?
- Does the UI work correctly across device sizes?
- Is the interface accessible to users with disabilities?
