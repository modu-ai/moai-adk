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

## Background Agent Execution

[HARD] Background subagents (`run_in_background: true`) MUST NOT perform Write/Edit operations.

Background agents auto-deny all non-pre-approved permission prompts because they cannot interact with the user. Even with `mode: "bypassPermissions"`, the background execution context does not fully inherit the parent session's permission allowlist.

Rules for agent spawning:
- **Read-only tasks** (research, analysis, review): `run_in_background: true` is safe
- **Write tasks** (implementation, refactoring, file creation): `run_in_background: false` required
- **Parallel writes needed**: Process directly from the main session, or use sequential foreground agents
- **Pre-approved writes**: Add path patterns to settings.json `permissions.allow` for background write support

Decision matrix:
- Agent reads files only → `run_in_background: true` (parallel, fast)
- Agent writes files → `run_in_background: false` (sequential, reliable)
- Multiple agents need to write different files → Use main session directly or foreground agents in sequence

## Tool Usage Guidelines

[HARD] Agents must follow tool usage patterns optimized for accuracy and efficiency.

### File Operations Pattern

Read-before-write rule:
- ALWAYS Read a file before using Edit on it
- Use Grep to locate specific line numbers before targeted Read with offset/limit
- Use Glob to discover files before reading — never guess file paths
- Prefer Edit over Write for existing files (sends only the diff, preserves context)

Path handling:
- Use absolute paths for all file operations
- Never construct paths from assumptions — verify with Glob or Bash `ls` first
- When working in worktrees, use project-root-relative paths for write targets

### Search Pattern

Progressive narrowing:
1. Glob to find candidate files by pattern
2. Grep with `files_with_matches` to narrow by content
3. Grep with `content` mode + context lines for detailed inspection
4. Read with offset/limit for full section understanding

Avoid:
- Reading entire large files when only a specific section is needed
- Using Bash grep/find when Grep/Glob tools are available
- Searching without file type filters when the target language is known

### Tool Selection by Task

| Task | Preferred Tool | Avoid |
|------|---------------|-------|
| Find files by name | Glob | Bash find, Bash ls |
| Search file contents | Grep | Bash grep, Bash rg |
| Read file contents | Read | Bash cat, Bash head |
| Modify existing file | Edit | Bash sed, Write (overwrites) |
| Create new file | Write | Bash echo/cat heredoc |
| Run system commands | Bash | — |
| Explore codebase | Agent(Explore) | Multiple sequential Grep calls |

### Error Recovery Pattern

When a tool call fails:
1. Read the error message carefully — diagnose root cause
2. Verify assumptions: does the file/path exist? (Glob check)
3. Try an alternative approach — do not retry the identical call
4. After 3 failures on the same operation, report the blocker

## Time Estimation

[HARD] Never use time predictions in plans or reports.
- Use priority labels: Priority High / Medium / Low
- Use phase ordering: "Complete A, then start B"
- Prohibited: "2-3 days", "1 week", "as soon as possible"
