# Pre-Paste Checklist

Before pasting prompt.md into Claude Design, verify each item below.

## Content Review

- [ ] Goal section accurately describes the page(s) you want designed (landing + one documentation page template)
- [ ] References section lists URLs you have personally checked and verified as relevant style references
- [ ] Brand Voice section reflects your actual brand voice. If it still shows the "default — please customize" header, edit it before pasting
- [ ] Acceptance Criteria section reflects your real quality bar, especially the accessibility, responsiveness, and information-architecture requirements
- [ ] Out of Scope section explicitly lists everything you do NOT want designed (no pricing pages, no auth flows, no animations on v1)

## Brand Voice Decision

The prompt.md currently uses the default brand voice template because brand files at `.moai/project/brand/` are not yet populated. Before pasting:

- [ ] Decision made: either (a) edit prompt.md Section 3 with actual brand voice, OR (b) populate the brand files first and regenerate the handoff package
- [ ] If proceeding with default brand voice, you accept that the design will reflect generic professional-technical aesthetics rather than a unique brand identity

## Project-Internal Cleanup (auto-verified)

- [ ] prompt.md contains no SPEC identifiers (no `SPEC-HARNESS-XXX`, no `SPEC-V3R4-*`)
- [ ] prompt.md contains no internal directory path references (no `.moai/`, no `.claude/`)
- [ ] prompt.md contains no internal command references (no `/moai`, no orchestrator agent names)
- [ ] prompt.md contains no idea-tracking identifiers (no `IDEA-NNN`)
- [ ] prompt.md reads as if written by a product designer with no knowledge of the project's internal structure

## Scope Verification

- [ ] Out of Scope section lists things you explicitly do NOT want included in v1
- [ ] Goal section is scoped to one or two specific page types (landing + documentation template), not the entire product surface
- [ ] You have decided whether dark mode is in scope for v1 or deferred (current prompt says dark mode is optional)

## Reference Quality

- [ ] All reference URLs in references.md are reachable and load correctly
- [ ] The competitor analysis URLs reflect actual competitors you want to differentiate from (not just URLs you happened to find)
- [ ] The visual inspiration URLs represent the aesthetic direction you genuinely want to pursue

## Session Readiness

- [ ] You have a claude.com account with Design access enabled
- [ ] You have reviewed proposal.md and understand which product capabilities this design supports
- [ ] You have reviewed ideation.md and accept the Verdict (Proceed) plus the identified Weaknesses
- [ ] You are prepared to iterate on the generated design (typically 2-3 rounds of refinement before final)

## Post-Design Steps

After completing the Claude Design session:

- [ ] Save the Claude Design output bundle to a local directory
- [ ] Review the output against the Acceptance Criteria checklist (acceptance.md)
- [ ] Note any acceptance gaps and decide whether to iterate in Claude Design or accept partial completion
- [ ] When satisfied, return to the orchestrator and proceed to the next workflow phase (project documentation generation or specification authoring)

## Open Issues Surfaced During Handoff Assembly

- The brand voice files exist as templates but contain only `_TBD_` placeholders. Brand identity has not yet been defined for this project. The prompt.md uses the default brand voice fallback. If you want a designed product with distinctive brand voice, run a brand interview before using this prompt.
- The five-layer safety architecture is the most distinctive design element and the prompt instructs Claude Design to render it as a single continuous visual. Verify in the generated design that this instruction was followed; if not, iterate.
- Mermaid diagram styling is called out in Acceptance Criteria but the actual diagram contents are not specified in this handoff. You will need to provide diagram contents (or accept Claude Design's placeholder examples) during the design session.
