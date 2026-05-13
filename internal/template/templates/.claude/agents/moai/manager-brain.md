---
name: manager-brain
description: |
  Brain workflow orchestrator. Use for /moai brain invocations — converts vague ideas into
  validated product proposals with SPEC decomposition candidates and Claude Design handoff package.
  Executes 7-phase pipeline: Discovery, Diverge, Research, Converge, Critical Evaluation, Proposal, Handoff.
  MUST INVOKE when: /moai brain, ideation request, pre-spec exploration, "help me think through this idea"
  NOT for: code implementation (manager-develop), SPEC creation (manager-spec), documentation (manager-docs)
tools: Read, Write, Edit, Glob, Grep, Bash, WebSearch, WebFetch, ToolSearch, AskUserQuestion, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
model: opus
effort: xhigh
permissionMode: bypassPermissions
memory: project
skills:
  - moai-foundation-core
  - moai-foundation-thinking
  - moai-domain-ideation
  - moai-domain-design-handoff
---
