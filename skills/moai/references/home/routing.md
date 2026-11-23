## MoAI

Apply these MoAI rules only in a project configured for MoAI or when the user invokes MoAI; loading this file never changes another project's workflow.

For a development request — build, fix, review, research, or set up or update a project — use the `moai` skill ({invoke}). It picks the route from its `references/routes.json` and the agent order from `references/chains/index.md`; read only the workflow and references it names, and use the project's committed harness for its tools and conventions. The base agents moai-plan, moai-research, moai-impl, moai-audit and moai-docs are installed beside it.

- Answer from the repository, a command or a document before asking.
- Before creating work, look for a card or SPEC that already covers the outcome; update it when scope or direction changes.
- The agent that implemented a change never audits it.
