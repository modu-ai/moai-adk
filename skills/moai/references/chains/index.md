# Chains

When to read: after `routes.json` picked a route, before the parent session starts any agent.

A chain is the order in which the parent session uses agents for one kind of request. The parent session starts each agent in turn and hands the next one the previous result — the plan table, the test list, the verdict — as its input. An agent never starts another agent. Each agent reads the internal references its installed agent file names first (the `reads:` list of its source, which ships outside this skill); the last column names only what the parent session reads first.

| Request | Route | Agents in order | Parent reads first |
| --- | --- | --- | --- |
| Build something new | PLAN, RUN, VERIFY, SYNC, AUDIT | moai-plan, moai-impl, moai-docs, moai-audit | internal/decide, internal/sdd |
| Fix a bug | PLAN, RUN, VERIFY, SYNC, AUDIT | moai-research, moai-plan, moai-impl, moai-docs, moai-audit | internal/decide, internal/sdd |
| Review a pull request | AUDIT | moai-audit | internal/review |
| Research or compare | READ | moai-research | none |
| Check a page or do a web task | VERIFY | none | none |
| Set up a project | PROJECT | none | internal/interview, internal/questions |
| Update documents and Issues after a merge | SYNC | moai-docs | internal/decide |
| Keep working toward a goal | MODE-AGENT | none | internal/git-safety |
| A test fails once and passes on a rerun of the same head (flaky) | VERIFY | none | internal/test-design |
| Repair after a FAIL verdict | RUN, VERIFY, SYNC, AUDIT | moai-impl, moai-audit | internal/review |

Rules for every chain:

- Every change that modifies the repository has a SPEC, and its order is PLAN → RUN → VERIFY → SYNC → AUDIT → merge (`internal/sdd`): SYNC writes the SPEC Result, documents and `@cc:` comments on the same branch, and AUDIT only checks them.
- The agent that implemented a change never audits it. `moai-audit` runs in a separate context on the exact head, and the parent session, not the agent, writes the verdict.
- A bug fix starts with a test that fails for the reported reason (`moai-impl` reads `internal/test-design`).
- A flaky test is classified from evidence only (the same head, the same job, the failing and the passing run) and recorded once per test until a change repairs it; a rerun that turns green is not a fix (`workflows/verify.md`).
- The re-audit after a FAIL goes to the first auditor, who judges the repair again and every changed criterion.
- Whoever wrote a change runs `internal/self-check` before asking for an audit or opening a pull request.
- Every commit for a SPEC carries the identity trailers of `internal/commit-trailers`, and every audit verdict names its auditor's session the same way.
- In a user project, RUN and VERIFY take the check commands, the AC-to-test tracing and the dependency existence check from `internal/quality`.
- `moai-research` fans out: the parent may start several at once, each on one question, and merges their answers with the sources and read times they report.
- A choice the request did not make is sorted with `internal/decide` before anyone is asked.
- A question about how a stateful part behaves, why a pipeline is slow, what of an old part to keep or what a refactor could break is looked at with `internal/lenses`, which reports and changes no code.
- The `versioning` rows of the PROJECT interview are asked with `internal/versioning`, read only when those rows are asked or a release version is computed.
- Any route that prepares, tags or publishes a release or release candidate runs `internal/versioning` `## At every release` first.
- MODE-AGENT picks one card at a time and runs the chain of that card's request; its own rules are in `workflows/mode-agent.md`.
- When the host cannot start agents, the parent session follows the same order itself and keeps the implementing and auditing contexts apart.
