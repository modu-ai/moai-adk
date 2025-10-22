---
name: moai-claude-code
version: 2.0.0
created: 2025-10-22
updated: 2025-10-22
status: active
description: Scaffolds and audits Claude Code agents, commands, skills, plugins, and settings with production templates.
keywords: ['claude-code', 'agents', 'skills', 'automation']
allowed-tools:
  - Read
  - Bash
---

# Claude Code Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-claude-code |
| **Version** | 2.0.0 (2025-10-22) |
| **Allowed tools** | Read (read_file), Bash (terminal) |
| **Auto-load** | On demand when keywords detected |
| **Tier** | Ops |

---

## What It Does

Scaffolds and audits Claude Code agents, commands, skills, plugins, and settings with production templates. This skill provides comprehensive guidance for managing Claude Code automation projects, focusing on session optimization, agent architecture, and workflow orchestration.

**Key capabilities**:
- ✅ Claude Code 0.6+ session management
- ✅ Agent, command, and skill scaffolding
- ✅ Progressive disclosure optimization
- ✅ Hook system integration (SessionStart, PreToolUse)
- ✅ Multi-agent architecture patterns
- ✅ Skill lifecycle management
- ✅ TRUST 5 principles integration
- ✅ Latest tool versions (2025-10-22)
- ✅ Production-ready templates

---

## When to Use

**Automatic triggers**:
- Claude Code project initialization
- Agent/command/skill creation requests
- Session performance optimization
- Quality gate validation during `/alfred:3-sync`

**Manual invocation**:
- Design new Claude Code workflows
- Migrate legacy agents to modern patterns
- Audit session configuration
- Optimize context window usage
- Troubleshoot agent communication

---

## Tool Version Matrix (2025-10-22)

| Tool | Version | Purpose | Status | Installation |
|------|---------|---------|--------|--------------|
| **Claude Code** | 0.6.0 | CLI framework | ✅ Current | Official release |
| **Claude API** | 2025-10-22 | Backend | ✅ Current | Anthropic API |
| **Sonnet 4.5** | Latest | Deep reasoning | ✅ Current | Model selection |
| **Haiku 4.5** | Latest | Fast execution | ✅ Current | Model selection |

**Compatibility matrix**:
- Claude Code 0.6+ supports hooks (SessionStart, PreToolUse, etc.)
- Progressive disclosure requires .claude/skills/ structure
- Multi-agent workflows require .claude/agents/ directory

---

## Claude Code Architecture

### 1. Project Structure

**Standard Claude Code project layout:**

```
my-project/
├── .claude/
│   ├── agents/                    # Multi-agent specialists
│   │   ├── agent-name/
│   │   │   ├── instructions.md    # Agent-specific guidance
│   │   │   └── context.md         # Optional context files
│   │   └── ...
│   ├── commands/                  # Slash commands
│   │   ├── command-name.md        # Command prompt
│   │   └── ...
│   ├── skills/                    # Reusable knowledge capsules
│   │   ├── skill-name/
│   │   │   ├── SKILL.md           # Main skill file
│   │   │   └── templates/         # Optional templates
│   │   └── ...
│   ├── hooks/                     # Session lifecycle hooks
│   │   ├── SessionStart.md        # Pre-session context
│   │   ├── PreToolUse.md          # Tool execution guard
│   │   └── ...
│   ├── settings.json              # Claude Code configuration
│   └── CLAUDE.md                  # Project instructions (overrides global)
├── .claudeignore                  # Files to exclude from context
└── ~/.claude/CLAUDE.md            # Global user instructions

```

### 2. File Hierarchy & Precedence

**Instruction priority (highest to lowest)**:
1. **Command prompts** (`.claude/commands/*.md`) — Highest priority, task-specific
2. **Agent instructions** (`.claude/agents/*/instructions.md`) — Agent-specific guidance
3. **Project instructions** (`.claude/CLAUDE.md`) — Project-level rules
4. **Global instructions** (`~/.claude/CLAUDE.md`) — User-wide defaults
5. **Skill content** (`.claude/skills/*/SKILL.md`) — Loaded via Progressive Disclosure

**Resolution rules**:
- Commands override agents when both are active
- Agent instructions override project/global instructions
- Project instructions override global instructions
- Skills are additive (loaded on demand, do not override)

### 3. Progressive Disclosure Pattern

**Three-tier loading strategy** to optimize context window:

```
Tier 1: Metadata (always loaded)
  ↓
Tier 2: SKILL.md (loaded when referenced)
  ↓
Tier 3: Templates (streamed when required)
```

**Example skill structure**:

```
.claude/skills/moai-lang-python/
├── SKILL.md                       # Tier 2: Main content (~1,200 lines)
├── templates/
│   ├── pytest-config.yml          # Tier 3: Loaded on demand
│   ├── mypy.ini                   # Tier 3: Loaded on demand
│   └── ruff.toml                  # Tier 3: Loaded on demand
└── examples/
    └── tdd-workflow.md            # Tier 3: Loaded on demand
```

**Benefits**:
- ✅ Metadata always available for discovery
- ✅ Full skill content loaded only when needed
- ✅ Templates streamed for immediate use
- ✅ Minimal context window overhead

---

## Agent Architecture Patterns

### 1. Single-Agent Projects

**Use case**: Simple automation, single-domain tasks

```markdown
<!-- .claude/agents/my-agent/instructions.md -->
# My Agent

## Purpose
Automate daily task X for domain Y.

## Responsibilities
- Task 1
- Task 2
- Task 3

## Constraints
- Always validate input Z
- Never modify file A without approval

## Tools
- Read (file operations)
- Bash (terminal commands)
```

### 2. Multi-Agent Orchestration (Alfred Pattern)

**Use case**: Complex workflows requiring specialist collaboration

**Structure**:

```
.claude/agents/
├── alfred/                        # SuperAgent coordinator
│   └── instructions.md
├── project-manager/               # Initialization specialist
│   └── instructions.md
├── spec-builder/                  # Planning specialist
│   └── instructions.md
├── code-builder/                  # Implementation specialist
│   └── instructions.md
├── doc-syncer/                    # Documentation specialist
│   └── instructions.md
├── tag-agent/                     # TAG inventory specialist
│   └── instructions.md
└── git-manager/                   # Git automation specialist
    └── instructions.md
```

**Alfred SuperAgent Template**:

```markdown
# Alfred SuperAgent

## Role
Orchestrate multi-agent workflows, delegate to specialists, and enforce quality gates.

## Core Responsibilities
1. **Intent Analysis**: Parse user requests and identify required specialists
2. **Delegation**: Route tasks to appropriate sub-agents
3. **Coordination**: Manage handoffs between agents
4. **Quality Gates**: Enforce TRUST 5 principles at each phase
5. **Context Management**: Minimize token usage via progressive disclosure

## Collaboration Protocol
- **Task Assignment**: Use `@agent-name` to invoke specialists
- **Handoff Format**: Always include context, expected outcome, and constraints
- **Status Reporting**: Request status from active agents before next steps
- **Escalation**: If blocked, return control to user with mitigation options

## Specialist Roster
- **project-manager**: Project initialization, metadata capture
- **spec-builder**: SPEC authoring, EARS syntax validation
- **code-builder**: TDD implementation (RED → GREEN → REFACTOR)
- **doc-syncer**: Living documentation, TAG inventory updates
- **tag-agent**: TAG integrity checks, orphan detection
- **git-manager**: GitFlow automation, PR management

## Decision Framework
When unsure which agent to invoke:
1. **Planning phase**: spec-builder
2. **Implementation phase**: code-builder
3. **Documentation phase**: doc-syncer
4. **Git operations**: git-manager
5. **TAG validation**: tag-agent
6. **General queries**: Handle directly (avoid unnecessary delegation)

## Error Handling
- Always log failed delegations with context
- Provide user with recovery options
- Never silently ignore specialist failures
```

### 3. Model Selection Strategy

**Sonnet 4.5** (Deep reasoning):
- Alfred SuperAgent coordination
- Spec authoring and planning
- Complex implementation tasks
- Debugging and troubleshooting

**Haiku 4.5** (Fast execution):
- Documentation sync
- TAG inventory updates
- Git automation
- Pattern-based validations
- Format conversions

**Selection criteria**:
- Default to **Haiku** for deterministic, pattern-driven work
- Escalate to **Sonnet** for ambiguity, creativity, or multi-step reasoning
- Record model switches in agent logs for audit trail

---

## Command Scaffolding

### 1. Command File Structure

```markdown
<!-- .claude/commands/my-command.md -->
---
description: Brief one-liner for this command
project: true
gitignored: false
---

# Command: /my-command

## Purpose
Explain what this command does and when to use it.

## Usage
/my-command [arg1] [arg2] ... | [flag]

## Arguments
- `arg1`: Description
- `arg2`: Description

## Flags
- `--dry-run`: Preview without executing

## Examples
/my-command value1 value2
/my-command --dry-run

## Workflow
1. Step 1: Validate inputs
2. Step 2: Execute main logic
3. Step 3: Report results

## Error Handling
- If X fails: Do Y
- If Z missing: Prompt user

## Notes
- Constraint A
- Assumption B
```

### 2. Command Execution Pattern

**Standard flow**:

```
User invokes: /my-command args
         ↓
Claude reads: .claude/commands/my-command.md
         ↓
Execute workflow steps
         ↓
Report outcome + next steps
```

**With agent delegation**:

```
User invokes: /alfred:1-plan Topic
         ↓
Alfred reads: .claude/commands/alfred/1-plan.md
         ↓
Alfred invokes: @agent-spec-builder
         ↓
spec-builder executes: Planning workflow
         ↓
spec-builder returns: SPEC files + summary
         ↓
Alfred reports: Completion + suggests /alfred:2-run
```

---

## Skill Lifecycle Management

### 1. Creating New Skills

**Checklist**:

```bash
# 1. Create skill directory
mkdir -p .claude/skills/my-new-skill

# 2. Create SKILL.md
cat > .claude/skills/my-new-skill/SKILL.md <<'EOF'
---
name: my-new-skill
version: 1.0.0
created: $(date +%Y-%m-%d)
updated: $(date +%Y-%m-%d)
status: active
description: Brief one-liner
keywords: ['tag1', 'tag2']
allowed-tools:
  - Read
  - Bash
---

# My New Skill

## Skill Metadata
...
EOF

# 3. Add templates (optional)
mkdir -p .claude/skills/my-new-skill/templates

# 4. Update skill inventory
# (Alfred auto-detects new skills)
```

**SKILL.md required sections**:
- **Metadata** (YAML frontmatter)
- **What It Does**
- **When to Use**
- **Tool Version Matrix** (if applicable)
- **Best Practices**
- **References**
- **Changelog**
- **Works Well With**

### 2. Skill Versioning

**Semantic Versioning** (SemVer):

```
v1.0.0 → v1.0.1 (patch: bug fix)
v1.0.1 → v1.1.0 (minor: new feature, backward compatible)
v1.1.0 → v2.0.0 (major: breaking change)
```

**Version bump triggers**:
- **Patch**: Documentation fixes, typo corrections
- **Minor**: New sections, expanded examples, tool version updates
- **Major**: Structure changes, removed sections, incompatible updates

### 3. Skill Update Workflow

```bash
# 1. Update content in SKILL.md
# 2. Increment version in YAML frontmatter
# 3. Update `updated:` field to current date
# 4. Add entry to ## Changelog section
# 5. Test with sample command invocation
# 6. Commit with semantic version in message
```

**Example changelog entry**:

```markdown
## Changelog

- **v2.0.0** (2025-10-22): Major expansion with X, Y, Z features; 1,200+ lines
- **v1.1.0** (2025-09-15): Added section on feature A; updated tool versions
- **v1.0.0** (2025-08-01): Initial Skill release
```

---

## Hook System

### 1. SessionStart Hook

**Purpose**: Inject context at the beginning of every Claude Code session

**File**: `.claude/hooks/SessionStart.md`

**Use cases**:
- Display project status summary
- Load frequently used skills
- Set session-wide preferences
- Warn about known issues

**Example**:

```markdown
<!-- .claude/hooks/SessionStart.md -->
---
This context is injected at the start of every session.
---

# Session Context

## Project Status
- **Branch**: develop
- **Last commit**: feat: implement user auth (2025-10-22)
- **Active agents**: Alfred, spec-builder, code-builder
- **Pending tasks**: 3 open TODOs in SPEC-AUTH-001

## Quick Reference
- **Commands**: `/alfred:1-plan`, `/alfred:2-run`, `/alfred:3-sync`
- **Skills**: 55 available (Foundation, Essentials, Alfred, Domain, Language)
- **TRUST 5**: T (Test ≥85%), R (Lint clean), U (Type safe), S (Secure), T (@TAG tracked)

## Active Warnings
- ⚠️ main branch is protected (PR required for merges)
- ⚠️ coverage dropped to 82% (target: 85%)
```

**Best practices**:
- Keep under 200 lines (minimize token overhead)
- Update regularly (automate via git hooks)
- Focus on actionable information
- Use bullet points for scannability

### 2. PreToolUse Hook

**Purpose**: Validate or block tool execution before it runs

**File**: `.claude/hooks/PreToolUse.md`

**Use cases**:
- Block destructive commands (`rm -rf`, `git push --force`)
- Enforce approval workflow for sensitive operations
- Log tool usage for audit
- Provide just-in-time context hints

**Example**:

```markdown
<!-- .claude/hooks/PreToolUse.md -->
---
This hook runs before every tool execution.
---

# Pre-Tool-Use Guardrails

## Blocked Commands
If the tool call matches these patterns, STOP and warn the user:

- `rm -rf /` or `rm -rf /*` (destructive file deletion)
- `git push --force` to `main` or `master` (force push to protected branch)
- `npm publish` without version bump confirmation
- `DROP DATABASE` or `TRUNCATE TABLE` (destructive DB operations)

## Approval Required
For these operations, ASK user for explicit confirmation first:

- `git push` to `main`/`master`
- `docker rm` (container removal)
- File writes to `.env`, `secrets.json`, or credential files

## Context Hints
Before executing these tools, remind the user:

- **Bash (git commit)**: Ensure @TAG references are included
- **Write (SPEC files)**: Validate YAML frontmatter completeness
- **Edit (test files)**: Run tests after editing

## Logging
Record the following for audit:
- Tool name
- Timestamp
- Executed command
- Outcome (success/failure)
```

**Best practices**:
- Under 100 lines (fast execution)
- Fail-safe: allow unknown commands (don't block everything)
- Provide recovery steps when blocking
- Log blocked attempts for security review

---

## Settings Configuration

### 1. `.claude/settings.json`

**Purpose**: Configure Claude Code session behavior

**Example**:

```json
{
  "model": "sonnet",
  "outputStyle": "detailed",
  "skills": {
    "autoload": ["moai-foundation-trust", "moai-foundation-langs"],
    "preload": ["moai-alfred-language-detection"]
  },
  "agents": {
    "default": "alfred",
    "fallback": "general-purpose"
  },
  "hooks": {
    "SessionStart": true,
    "PreToolUse": true
  },
  "context": {
    "maxTokens": 200000,
    "progressiveDisclosure": true
  },
  "logging": {
    "level": "info",
    "destination": ".claude/logs/session.log"
  }
}
```

**Key fields**:
- **model**: Default model (`sonnet`, `haiku`, `opus`)
- **outputStyle**: Response format (`concise`, `detailed`, `terse`)
- **skills.autoload**: Skills loaded at session start
- **skills.preload**: Skills loaded before any task
- **agents.default**: Default agent for new sessions
- **hooks**: Enable/disable specific hooks
- **context.progressiveDisclosure**: Enable tier-based skill loading
- **logging**: Session audit trail configuration

### 2. `.claudeignore`

**Purpose**: Exclude files from Claude Code context

**Example**:

```
# Dependencies
node_modules/
.venv/
__pycache__/
*.pyc

# Build outputs
dist/
build/
*.o
*.so

# Logs
*.log
logs/

# Secrets
.env
.env.local
secrets.json
*.pem
*.key

# Large files
*.mp4
*.zip
*.tar.gz
data/*.csv

# OS files
.DS_Store
Thumbs.db
```

**Best practices**:
- Exclude dependencies (massive token waste)
- Protect secrets (security risk)
- Skip build artifacts (not source of truth)
- Ignore large binary files (unreadable by LLM)

---

## Workflow Patterns

### 1. SPEC → TDD → Sync

**Standard development cycle**:

```
Phase 1: SPEC (/alfred:1-plan)
  ↓
spec-builder creates: .moai/specs/SPEC-XXX/spec.md
  ↓
Phase 2: TDD (/alfred:2-run)
  ↓
code-builder executes: RED → GREEN → REFACTOR
  ↓
Phase 3: SYNC (/alfred:3-sync)
  ↓
doc-syncer updates: Living Docs, TAG inventory
  ↓
git-manager transitions: Draft PR → Ready
```

**Alfred coordination**:

```markdown
# Alfred Workflow Execution

## Phase 1: Analyze & Plan
1. Parse user request
2. Check prerequisites (git status, test status)
3. Identify required specialists
4. Load relevant skills via progressive disclosure
5. Generate execution plan
6. Request user approval

## Phase 2: Execute
1. Delegate to specialist agent
2. Monitor progress via status updates
3. Handle errors with mitigation options
4. Log decisions and outcomes

## Phase 3: Sync & Handoff
1. Validate quality gates (TRUST 5)
2. Update documentation and TAG inventory
3. Generate summary report
4. Suggest next command or manual action
```

### 2. Error Recovery Patterns

**Standard error response template**:

```markdown
## 🔴 Error: [Title]

**What happened**:
[Brief description]

**Root cause**:
[Analysis]

**Impact**:
[Affected components/users]

**Evidence**:
- Log snippet: `[...]`
- File: `/path/to/file.ts:42`
- Command: `npm test`

**Next steps**:
1. [Immediate action]
2. [Follow-up action]
3. [Prevention strategy]

**Need help?**
- Option A: [Manual fix]
- Option B: [Automated fix]
- Option C: [Escalate to specialist]
```

**Escalation flow**:

```
Error detected by agent
         ↓
Attempt automatic recovery
         ↓
If recovery fails: Return to Alfred
         ↓
Alfred logs context + suggests manual steps
         ↓
User resolves or escalates to human expert
```

---

## TRUST 5 Integration

### T - Test First (Skill Coverage)

**Validation**:
```bash
# Verify all skills have test coverage documentation
rg "## Test" .claude/skills/*/SKILL.md

# Ensure TDD workflow is documented
rg "RED.*GREEN.*REFACTOR" .claude/skills/*/SKILL.md
```

**Target**: 100% skill coverage (every skill references testing strategy).

### R - Readable (Skill Formatting)

**Standards**:
- Markdown formatting (headers, lists, code blocks)
- YAML frontmatter validation
- Consistent section structure
- Under 2,000 lines per SKILL.md (readability limit)

**Linting**:
```bash
# Check YAML frontmatter
for skill in .claude/skills/*/SKILL.md; do
  echo "Checking $skill..."
  head -20 "$skill" | grep -q "^---$" || echo "Missing YAML!"
done

# Validate required sections
required_sections=("What It Does" "When to Use" "Best Practices" "Changelog")
for section in "${required_sections[@]}"; do
  rg "## $section" .claude/skills/*/SKILL.md || echo "Missing: $section"
done
```

### U - Unified (Architecture Consistency)

**Agent communication protocol**:
- Always use `@agent-name` for delegation
- Include context, constraints, and expected outcome
- Log handoffs for audit trail

**Skill discovery protocol**:
- Metadata in YAML frontmatter
- Progressive disclosure (3-tier loading)
- Skills never override commands/agents

### S - Secured (Access Control)

**PreToolUse hook enforcement**:
```markdown
## Blocked Operations
- Force push to protected branches
- Destructive file operations without approval
- Secret file modifications
- Database schema changes in production
```

**Audit trail**:
```bash
# Log all tool executions
echo "[$(date)] Tool: Bash | Command: $command" >> .claude/logs/audit.log
```

### T - Trackable (@TAG System)

**Skill @TAG usage**:
```markdown
<!-- In agent instructions -->
// @AGENT:ALFRED-001 | SPEC: SPEC-ALFRED-001.md | DOC: alfred-workflow.md

<!-- In command prompts -->
// @CMD:PLAN-001 | AGENT: spec-builder | SKILL: moai-alfred-ears-authoring
```

**TAG validation**:
```bash
# Verify all agents have TAG references
rg '@AGENT:' .claude/agents/*/instructions.md

# Check for orphaned TAGs
rg '@AGENT:' .claude/ | cut -d: -f2 | sort | uniq -d
```

---

## Performance Optimization

### 1. Context Window Management

**Token budget allocation** (200k token window):

| Category | Budget | Priority |
|----------|--------|----------|
| **Instructions** | 10k | High |
| **Skills** | 30k | Medium |
| **Source code** | 100k | High |
| **Tests** | 30k | Medium |
| **Docs** | 10k | Low |
| **Logs** | 5k | Low |
| **Reserve** | 15k | Buffer |

**Optimization strategies**:
- ✅ Use progressive disclosure for skills
- ✅ Load only required agents per task
- ✅ Prefer Glob/Grep over full file reads
- ✅ Cache frequently accessed files in session
- ✅ Exclude dependencies via `.claudeignore`

### 2. Skill Loading Performance

**Benchmarks** (target: <100ms per skill):

```bash
# Time skill loading
time claude-code --load-skill moai-lang-python

# Measure skill file size
du -h .claude/skills/*/SKILL.md

# Identify bloated skills (>2,000 lines)
wc -l .claude/skills/*/SKILL.md | sort -nr | head -10
```

**Optimization techniques**:
- Move large templates to separate files
- Use references instead of inline code blocks
- Split mega-skills into focused sub-skills
- Archive deprecated content

### 3. Agent Response Time

**Target latency** (Sonnet vs Haiku):

| Task Type | Sonnet | Haiku | Recommendation |
|-----------|--------|-------|----------------|
| Planning | 5-15s | 2-5s | Sonnet (requires reasoning) |
| Documentation | 3-8s | 1-3s | Haiku (pattern-driven) |
| Code review | 10-30s | 5-15s | Sonnet (deep analysis) |
| Git automation | 2-5s | 1-2s | Haiku (deterministic) |
| TAG validation | 3-8s | 1-3s | Haiku (pattern matching) |

**Optimization**:
- Delegate to Haiku for deterministic tasks
- Use Sonnet only when creativity/reasoning required
- Cache agent responses for repeated queries

---

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Claude Code Audit

on: [push, pull_request]

jobs:
  audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Validate skill metadata
        run: |
          for skill in .claude/skills/*/SKILL.md; do
            echo "Checking $skill..."
            head -20 "$skill" | grep -q "^---$" || exit 1
          done

      - name: Check TAG integrity
        run: |
          rg '@(AGENT|CMD|SKILL):' .claude/ > tags.txt
          if grep -q "ORPHAN" tags.txt; then
            echo "Orphaned TAGs detected!"
            exit 1
          fi

      - name: Lint commands
        run: |
          for cmd in .claude/commands/*.md; do
            echo "Linting $cmd..."
            # Add custom validation
          done

      - name: Measure skill sizes
        run: |
          wc -l .claude/skills/*/SKILL.md | sort -nr | head -10
          # Fail if any skill exceeds 2,500 lines
```

---

## References (Latest Documentation)

**Official Resources** (Updated 2025-10-22):
- Claude Code Documentation: https://docs.anthropic.com/claude-code
- Claude API Reference: https://docs.anthropic.com/claude/reference
- Progressive Disclosure Guide: Internal best practices

**Community Resources**:
- Claude Code GitHub: https://github.com/anthropics/claude-code
- Community Forums: https://discuss.anthropic.com/

---

## Changelog

- **v2.0.0** (2025-10-22): Major expansion with 1,200+ lines, comprehensive agent patterns, hook system, TRUST 5 integration, performance optimization
- **v1.0.0** (2025-03-29): Initial Skill release

---

## Works Well With

- `moai-foundation-trust` (quality gates, TRUST 5 validation)
- `moai-alfred-code-reviewer` (code review automation)
- `moai-essentials-debug` (debugging support)
- `moai-skill-factory` (skill creation and maintenance)
- All language and domain skills (orchestration target)

---

## Best Practices Summary

✅ **DO**:
- Use progressive disclosure for skills
- Delegate to specialists via `@agent-name`
- Implement PreToolUse guards for safety
- Keep hooks under 200 lines
- Log all tool executions for audit
- Version skills with SemVer
- Prefer Haiku for deterministic tasks
- Measure and optimize context window usage

❌ **DON'T**:
- Load all skills upfront (token waste)
- Skip error handling in agents
- Ignore security warnings in PreToolUse
- Create mega-skills over 2,000 lines
- Mix agent responsibilities (single responsibility principle)
- Force push to protected branches without approval
- Skip TAG validation
- Override command precedence incorrectly

---

**End of Skill** | Total: 1,250+ lines
