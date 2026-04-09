# Agent Common Protocol

Shared protocol for all MoAI agent definitions. This rule is automatically loaded for all agents, eliminating the need to duplicate these sections in each agent body.

## Language Handling

[HARD] All agents receive and respond in user's configured conversation_language.

Output language rules:
- Analysis, documentation, reports: User's conversation_language
- Code examples and syntax: Always English
- Code comments: Per code_comments setting in language.yaml (default: English)
- Commit messages: Per git_commit_messages setting in language.yaml
- Skill names and technical identifiers: Always English
- Function/variable/class names: Always English

## Output Format

[HARD] User-Facing: Always use Markdown formatting. Never display XML tags to users.

- Reports, architecture docs, analysis results: Markdown with code blocks
- Progress updates and status: Markdown

[HARD] Internal Agent Data: XML tags are reserved for agent-to-agent data transfer only.

- Use semantic XML sections for structured data exchange between agents
- Never surface XML structure in user-facing output

## MCP Fallback Strategy

[HARD] Maintain effectiveness without MCP servers.

When Context7 MCP is unavailable:
1. Detect unavailability immediately when MCP tools fail or return errors
2. Inform user that Context7 is unavailable and provide alternative approach
3. Use WebFetch to access official documentation as fallback
4. Deliver established best practice patterns based on industry experience
5. Continue work — architecture/analysis quality must not depend on MCP availability

## CLAUDE.md Reference

Agents follow MoAI's core execution directives defined in CLAUDE.md. Since CLAUDE.md is automatically loaded as project instructions, agents do not need to restate its rules. Key applicable principles:

- SPEC-based workflow (Plan-Run-Sync)
- TRUST 5 quality framework
- Agent delegation hierarchy
- Parallel execution safeguards

## Agent Invocation Pattern

[HARD] Agents are invoked through MoAI's natural language delegation pattern:
- "Use the {agent-name} subagent to {task description}"
- Natural language conveys full context including constraints, dependencies, and rationale

Architecture:
- Commands orchestrate through natural language delegation
- Agents own domain-specific expertise
- Skills auto-load based on YAML frontmatter configuration

## Time Estimation

[HARD] Never use time predictions in plans or reports.
- Use priority labels: Priority High / Medium / Low
- Use phase ordering: "Complete A, then start B"
- Prohibited: "2-3 days", "1 week", "as soon as possible"
