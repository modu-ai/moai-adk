# Checklist — MoAI Cockpit Design Handoff

A pre/post checklist for the user (you) when working with Claude Design on this brief.

---

## Before Pasting prompt.md into Claude Design

- [ ] Open claude.com/design (or the appropriate Claude Design entry point)
- [ ] Start a new design session
- [ ] Have the URL `https://api.anthropic.com/v1/design/h/-jhWdwlf2BjzXHPBHl4JxQ` ready (it is referenced inside prompt.md, but confirm Claude Design can fetch it)
- [ ] Decide upfront: do you want light mode first, dark mode first, or both in parallel?
- [ ] Decide upfront: 1440×900 desktop only, or also include 768px mobile?
- [ ] Have this folder bookmarked so you can quickly reference acceptance.md while reviewing output

## Pasting prompt.md

- [ ] Copy the entire content of `prompt.md` and paste as the first message in the design session
- [ ] If Claude Design asks for clarification, prefer answers that preserve the read-only invariant and the calm/ambient tone
- [ ] If Claude Design proposes adding interactive mutating buttons, push back and reference the prompt.md "Read-only Invariant" section

## During Generation

- [ ] If Claude Design generates a layout that buries the Workflow Tracker, request that it be moved to top/headline position
- [ ] If color contrast looks borderline, ask Claude Design to verify against WCAG 2.1 AA explicitly
- [ ] If badges use color-only encoding, ask Claude Design to add icon and text label
- [ ] If the design feels noisy or marketing-like, redirect to the visual reference anchors in context.md (Vercel, Linear, Stripe, Raycast)

## After Initial Generation

- [ ] Review against `acceptance.md` checklist sections A through I in order
- [ ] If any item in section B (Read-Only Invariant) fails, request regeneration of the offending panels
- [ ] If sections A, C, D, F, G all pass: download mockups, component sheet, token sheet
- [ ] If section H (Tone & Mood) feels off, request a tone adjustment with reference to context.md anchors

## Exporting from Claude Design

- [ ] Export full mockups as PNG (light + dark) at 2× resolution
- [ ] Export component spec sheet as PDF or PNG
- [ ] Export token sheet as JSON if Claude Design supports it (otherwise as a structured table)
- [ ] Save all exports to `.moai/design/` (or the appropriate downstream folder for your project's import workflow)

## Bringing the Design Back into the Project

- [ ] Run the project's design-import workflow (e.g., `/moai design --path A`) — this will read from `.moai/design/` and create a brief artifact
- [ ] Verify the imported tokens match what Claude Design exported
- [ ] Update `.moai/project/brand/visual-identity.md` with extracted tokens (replace `_TBD_` placeholders)
- [ ] Update `.moai/project/brand/brand-voice.md` if the design session surfaced voice/tone decisions

## Validating the Outcome

- [ ] Side-by-side: open prompt.md and the generated design, confirm every panel and constraint is addressed
- [ ] Run through `acceptance.md` one final time as a final pass/fail
- [ ] If any acceptance item fails, decide: regenerate (preferred), accept with note (acceptable for nice-to-haves), or escalate (for critical mismatches)
- [ ] Save the final approved design package back to `.moai/brain/IDEA-001/claude-design-handoff/output/` (create this folder)

## Next Steps After Design is Approved

- [ ] Run the project specification workflow (e.g., `/moai plan SPEC-V3R3-WEB-001`) — the spec author will use the approved design as visual reference
- [ ] In parallel, update project documentation (`product.md`, `structure.md`) with the brand tokens extracted during design
- [ ] Schedule a 1-week dogfooding period after first MVP release to validate the 30%+ CLI roundtrip reduction success metric
