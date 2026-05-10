# Pre-Paste Checklist

Before pasting prompt.md into Claude Design, verify:

## Content Review

- [ ] Goal section accurately describes what you want designed (dashboard + daily check-in)
- [ ] References section has 3-5 URLs you have verified are still active and relevant
- [ ] Brand Voice section reflects your actual brand (not the default placeholder)
- [ ] Acceptance Criteria section reflects your real quality bar (WCAG 2.1 AA is non-negotiable)

## Completeness Check

- [ ] All 5 sections present in prompt.md: Goal, References, Brand Voice, Acceptance Criteria, Out of Scope
- [ ] Out of Scope section lists things you explicitly do NOT want (helps Claude Design stay focused)
- [ ] The "Goal" section is scoped to specific pages/views (dashboard + check-in, not the entire product)

## Internal Reference Cleanup (auto-verified by brain workflow)

- [ ] No SPEC- identifiers in prompt.md (e.g., SPEC-AUTH-001 would be inappropriate)
- [ ] No .moai/ path references in prompt.md
- [ ] No /moai commands in prompt.md
- [ ] No manager- or agent names in prompt.md

## Session Readiness

- [ ] You have a claude.com account with Claude Design access
- [ ] You have reviewed ideation.md (Lean Canvas) and proposal.md to confirm the design aligns with product direction
- [ ] You are ready to provide feedback on the generated design output

## After Design Session

Once Claude Design generates output:

1. Save the design artifacts to a local directory (e.g., `habitscope-design-v1/`)
2. Run the import command to process the handoff bundle:
   ```
   /moai design --path A --bundle habitscope-design-v1/
   ```
3. The imported artifacts will be available at `.moai/design/` for the frontend implementation phase
4. Proceed to:
   ```
   /moai project --from-brain IDEA-EXAMPLE
   ```
   to generate project documentation using the brain proposal as input
