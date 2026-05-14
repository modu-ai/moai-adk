# Design References

## Competitor Analysis

- https://github.com/revfactory/harness — the canonical reference framework; the design should explicitly improve on its GitHub-README-only documentation surface. Where revfactory's harness lives entirely inside GitHub markdown, the v2 harness deserves a proper documentation site that surfaces the safety stack and the principle hierarchy.
- https://learn.microsoft.com/en-us/agent-framework/overview/ — Microsoft Agent Framework documentation; study its information density and sidebar navigation, but improve on the visual restraint (MAF's interface is busy; the harness v2 design should feel calmer).
- https://blog.langchain.com/reflection-agents/ — LangChain's reflection-pattern explainer; reuse the diagram-led explanation style, especially for the self-critique loop visualization.

## Visual Inspiration

- https://anthropic.com — primary stylistic reference. Restrained typography, generous whitespace, principle-first hierarchy, monospace accents for technical terms. The way Anthropic communicates safety architecture (constitution, RLAIF, harmlessness) is a direct model for how the harness v2 should communicate its five-layer safety stack.
- https://linear.app — opinionated visual confidence; the design itself signals that the team has strong opinions on quality. Emulate the precise typographic rhythm and the restraint with color (single accent, dominant neutral grayscale).
- https://docs.astral.sh/uv/ — modern documentation site aesthetic for Rust/Python tooling. Clean code blocks, minimal chrome, fast feel, no marketing animations. The harness v2 docs should feel similarly substantive.
- https://www.cursor.com — developer-tool product landing energy; how to make autonomous capabilities feel trustworthy rather than threatening. Study how Cursor's hero communicates AI agency without triggering reader skepticism.

## UX Pattern References

- https://nextjs.org/docs — documentation sidebar pattern with collapsible sections, breadcrumbs at top, on-page table of contents on the right. The harness v2 docs should follow this three-column layout for desktop, collapsing to single-column on mobile.
- https://tailwindcss.com/docs — code-example presentation: inline copy-to-clipboard button, syntax highlighting, optional tabs for multiple language variants. The harness v2 docs will have many configuration examples and should follow this presentation standard.
- https://www.langchain.com/langgraph — diagram-driven feature explanations. The harness v2 should similarly anchor its key sections with inline mermaid diagrams rather than relying on prose alone.

## Documentation Tooling Examples

- https://docusaurus.io — full-featured documentation framework site, demonstrates good information architecture for technical products.
- https://vitepress.dev — minimal, fast documentation framework; closer to the aesthetic the harness v2 design should target than Docusaurus's more elaborate default theme.

Note: References focus on visual style and information architecture. The actual content (copy, diagrams, code examples) will be authored separately and is not constrained by these references.
