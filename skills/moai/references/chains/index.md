# Chains

When to read: after `routes.json` picked a route, before the parent session starts any agent.

A chain is the order in which the parent session uses agents for one kind of request. The parent session starts each agent in turn and hands the next one the previous result — the plan table, the test list, the verdict — as its input. An agent never starts another agent. Each agent reads the internal references of its own `reads:` list (`references/agents/<role>.md`); the last column names only what the parent session reads first.

| Request | Route | Agents in order | Parent reads first |
| --- | --- | --- | --- |
| Build something new | PLAN, RUN | moai-plan, moai-impl, moai-audit, moai-docs | internal/decide |
| Fix a bug | RUN | moai-research, moai-impl, moai-audit | internal/decide |
| Review a pull request | AUDIT | moai-audit | internal/review |
| Research or compare | READ | moai-research | none |
| Check a page or do a web task | VERIFY | none | none |
| Set up a project | PROJECT | none | internal/interview |
| Update documents and Issues after a merge | SYNC | moai-docs | internal/decide |
| Keep working toward a goal | MODE-AGENT | none | internal/git-safety |

Rules for every chain:

- The agent that implemented a change never audits it. `moai-audit` runs in a separate context on the exact head, and the parent session, not the agent, writes the verdict.
- A bug fix starts with a test that fails for the reported reason (`moai-impl` reads `internal/test-design`).
- `moai-research` fans out: the parent may start several at once, each on one question, and merges their answers with the sources and read times they report.
- A choice the request did not make is sorted with `internal/decide` before anyone is asked.
- MODE-AGENT picks one card at a time and runs the chain of that card's request; its own rules are in `workflows/mode-agent.md`.
- When the host cannot start agents, the parent session follows the same order itself and keeps the implementing and auditing contexts apart.
