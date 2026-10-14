---
description: "Measure and govern always-loaded versus path-scoped MoAI rules"
paths: ".claude/rules/moai/**"
---

# Rule Loading Budget

New rules default to a `paths:` trigger. A rule may remain always-loaded only
when it applies to an arbitrary turn (for example shared-checkout safety,
user-language handling, or context-boundary recovery) and its frontmatter or
opening loading-scope note records that reason. Path-scoped rules keep the
dispatch contract and move long rationale/examples into a companion file.

For each change to `.claude/rules/moai/`, record the measured count and byte
total of always-loaded files, the path-scoped count, and the runtime
`InstructionsLoaded`/token observation when the host exposes it. A byte saving
without a reachability check is not a performance result. If the host does not
expose the runtime observation, report that gap rather than inventing a token
reduction.
