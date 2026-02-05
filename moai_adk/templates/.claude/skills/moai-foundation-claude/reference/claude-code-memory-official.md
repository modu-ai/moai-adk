# Claude Code Memory System - Official Documentation Reference

Source: https://code.claude.com/docs/en/memory

## Key Concepts

### What is Claude Code Memory?

Claude Code Memory provides a hierarchical context management system that allows agents to maintain persistent information across sessions, projects, and organizations. It enables consistent behavior, knowledge retention, and context-aware interactions.

### Memory Architecture

Three-Tier Hierarchy:

1. Enterprise Policy: Organization-wide policies and standards
2. Project Memory: Project-specific knowledge and context
3. User Memory: Personal preferences and individual knowledge

Memory Flow:

```
Enterprise Policy → Project Memory → User Memory
 (Highest) (Project) (Personal)
 ↓ ↓ ↓
 Overrides Overrides Overrides
```

## Memory Storage and Access

### File-Based Memory System

Memory File Locations:

- Enterprise: `/etc/claude/policies/` (system-wide)
- Project: `./CLAUDE.md` (project-specific)
- User: `~/.claude/CLAUDE.md` (personal preferences)
- Local: `.claude/memory/` (project metadata)

File Types and Purpose:

```
Project Root/
 CLAUDE.md # Main project memory (highest priority in project)
 .claude/memory/ # Structured project metadata
 execution-rules.md # Execution constraints and rules
 agents.md # Agent catalog and capabilities
 commands.md # Command references and patterns
 delegation-patterns.md # Agent delegation strategies
 token-optimization.md # Token budget management
 .moai/
 config/ # Configuration management
 config.json # Project settings
 cache/ # Memory cache and optimization
```

### Memory Import Syntax

Direct Import Pattern:

```markdown
# In CLAUDE.md files

@path/to/import.md # Import external memory file
@.claude/memory/agents.md # Import agent reference
@.claude/memory/commands.md # Import command reference
@memory/delegation-patterns.md # Relative import from memory directory
```

Conditional Import:

```markdown
# Import based on environment or configuration

<!-- @if environment == "production" -->

@memory/production-rules.md

<!-- @endif -->

<!-- @if features.security == "enabled" -->

@memory/security-policies.md

<!-- @endif -->
```

## Memory Content Types

### Policy and Rules Memory

Execution Rules (`memory/execution-rules.md`):

```markdown
# Execution Rules and Constraints

## Core Principles

- Agent-first mandate: Always delegate to specialized agents
- Security sandbox: All operations in controlled environment
- Token budget management: Phase-based allocation strategy

## Agent Delegation Rules

- Required tools: Task(), AskUserQuestion(), Skill()
- Forbidden tools: Read(), Write(), Edit(), Bash(), Grep(), Glob()
- Delegation pattern: Sequential → Parallel → Conditional

## Security Constraints

- Forbidden paths: .env\*, .vercel/, .github/workflows/secrets
- Forbidden commands: rm -rf, sudo, chmod 777, dd, mkfs
- Input validation: Required before all processing
```

Agent Catalog (`memory/agents.md`):

```markdown
# Agent Reference Catalog

## Planning & Specification

- spec-builder: SPEC generation in EARS format
- plan: Decompose complex tasks step-by-step

## Implementation

- ddd-implementer: Execute DDD cycle (ANALYZE-PRESERVE-IMPROVE)
- backend-expert: Backend architecture and API development
- frontend-expert: Frontend UI component development

## Usage Patterns

- Simple tasks (1-2 files): Sequential execution
- Medium tasks (3-5 files): Mixed sequential/parallel
- Complex tasks (10+ files): Parallel with integration phase
```

### Configuration Memory

Settings Management (`config/config.json`):

```json
{
  "user": {
    "name": "Developer Name",
    "preferences": {
      "language": "en",
      "timezone": "UTC"
    }
  },
  "project": {
    "name": "Project Name",
    "type": "web-application",
    "documentation_mode": "comprehensive"
  },
  "constitution": {
    "test_coverage_target": 90,
    "enforce_tdd": true,
    "quality_gates": [
      "test-first",
      "readable",
      "unified",
      "secured",
      "trackable"
    ]
  },
  "git_strategy": {
    "mode": "team",
    "workflow": "github-flow",
    "auto_pr": true
  }
}
```

### Process Memory

Command References (`memory/commands.md`):

```markdown
# Command Reference Guide

## Core MoAI Commands

- /moai:0-project: Initialize project structure
- /moai:1-plan: Generate SPEC document
- /moai:2-run: Execute DDD implementation
- /moai:3-sync: Generate documentation
- /moai:9-feedback: Collect improvement feedback

## Command Execution Rules

- After /moai:1-plan: Execute /clear (mandatory)
- Token threshold: Execute /clear at >150K tokens
- Error handling: Use /moai:9-feedback for all issues
```

## Memory Management Strategies

### Memory Initialization

Project Bootstrap:

```bash
# Initialize project memory structure
/moai:0-project

# Creates:
# - .moai/config/config.yaml
# - .moai/memory/ directory
# - CLAUDE.md template
# - Memory structure files
```

Manual Memory Setup:

```bash
# Create memory directory structure
mkdir -p .claude/memory
mkdir -p .moai/config
mkdir -p .moai/cache

# Create initial memory files
touch .claude/memory/agents.md
touch .claude/memory/commands.md
touch .claude/memory/execution-rules.md
touch CLAUDE.md
```

### Memory Synchronization

Import Resolution:

```python
# Memory import resolution order
def resolve_memory_import(import_path, base_path):
 """
 Resolve @import paths in memory files
 1. Check relative to current file
 2. Check in .claude/memory/ directory
 3. Check in project root
 4. Check in user memory directory
 """
 candidates = [
 os.path.join(base_path, import_path),
 os.path.join(".claude/memory", import_path),
 os.path.join(".", import_path),
 os.path.expanduser(os.path.join("~/.claude", import_path))
 ]

 for candidate in candidates:
 if os.path.exists(candidate):
 return candidate
 return None
```

Memory Cache Management:

```bash
# Memory cache operations
claude memory cache clear # Clear all memory cache
claude memory cache list # List cached memory files
claude memory cache refresh # Refresh memory from files
claude memory cache status # Show cache statistics
```

### Memory Optimization

Token Efficiency Strategies:

```markdown
# Memory optimization techniques

## Progressive Loading

- Load core memory first (2000 tokens)
- Load detailed memory on-demand (5000 tokens each)
- Cache frequently accessed memory files

## Content Prioritization

- Priority 1: Execution rules and agent catalog (must load)
- Priority 2: Project-specific configurations (conditional)
- Priority 3: Historical data and examples (on-demand)

## Memory Compression

- Use concise bullet points over paragraphs
- Implement cross-references instead of duplication
- Group related information in structured sections
```

## Memory Access Patterns

### Agent Memory Access

Agent Memory Loading:

```python
# Agent memory access pattern
class AgentMemory:
 def __init__(self, session_id):
 self.session_id = session_id
 self.memory_cache = {}
 self.load_base_memory()

 def load_base_memory(self):
 """Load essential memory for agent operation"""
 essential_files = [
 ".claude/memory/execution-rules.md",
 ".claude/memory/agents.md",
 ".moai/config/config.yaml"
 ]

 for file_path in essential_files:
 self.memory_cache[file_path] = self.load_memory_file(file_path)

 def get_memory(self, key):
 """Get memory value with fallback hierarchy"""
 # 1. Check session cache
 if key in self.memory_cache:
 return self.memory_cache[key]

 # 2. Load from file system
 memory_value = self.load_memory_file(key)
 if memory_value:
 self.memory_cache[key] = memory_value
 return memory_value

 # 3. Return default or None
 return None
```

Context-Aware Memory:

```python
# Context-aware memory selection
def select_relevant_memory(context, available_memory):
 """
 Select memory files relevant to current context
 """
 relevant_memory = []

 # Analyze context keywords
 context_keywords = extract_keywords(context)

 # Match memory files by content relevance
 for memory_file in available_memory:
 relevance_score = calculate_relevance(memory_file, context_keywords)
 if relevance_score > 0.7: # Threshold
 relevant_memory.append((memory_file, relevance_score))

 # Sort by relevance and return top N
 relevant_memory.sort(key=lambda x: x[1], reverse=True)
 return [memory[0] for memory in relevant_memory[:5]]
```

## Memory Configuration

### Environment-Specific Memory

Development Environment:

```json
{
  "memory": {
    "mode": "development",
    "cache_size": "100MB",
    "auto_refresh": true,
    "debug_memory": true,
    "memory_files": [
      ".claude/memory/execution-rules.md",
      ".claude/memory/agents.md",
      ".claude/memory/commands.md"
    ]
  }
}
```

Production Environment:

```json
{
  "memory": {
    "mode": "production",
    "cache_size": "50MB",
    "auto_refresh": false,
    "debug_memory": false,
    "memory_files": [
      ".claude/memory/execution-rules.md",
      ".claude/memory/production-policies.md"
    ],
    "memory_restrictions": {
      "max_file_size": "1MB",
      "allowed_extensions": [".md", ".json"],
      "forbidden_patterns": ["password", "secret", "key"]
    }
  }
}
```

### User Preference Memory

Personal Memory Structure (`~/.claude/CLAUDE.md`):

```markdown
# Personal Claude Code Preferences

## User Information

- Name: John Developer
- Role: Senior Software Engineer
- Expertise: Backend Development, DevOps

## Development Preferences

- Language: Python, TypeScript
- Frameworks: FastAPI, React
- Testing: pytest, Jest
- Documentation: Markdown, OpenAPI

## Workflow Preferences

- Git strategy: feature branches
- Code review: required for PRs
- Testing coverage: >90%
- Documentation: comprehensive

## Tool Preferences

- Editor: VS Code
- Shell: bash
- Package manager: npm, pip
- Container: Docker
```

## Memory Maintenance

### Memory Updates and Synchronization

Automatic Memory Updates:

```bash
# Update memory from templates
claude memory update --from-templates

# Synchronize memory across team
claude memory sync --team

# Validate memory structure
claude memory validate --strict
```

Memory Version Control:

```bash
# Track memory changes in Git
git add .claude/memory/ CLAUDE.md
git commit -m "docs: Update project memory and agent catalog"

# Tag memory versions
git tag -a "memory-v1.2.0" -m "Memory version 1.2.0"
```

### Memory Cleanup

Cache Cleanup:

```bash
# Clear expired cache entries
claude memory cache cleanup --older-than 7d

# Remove unused memory files
claude memory cleanup --unused

# Optimize memory file size
claude memory optimize --compress
```

Memory Audit:

```bash
# Audit memory usage
claude memory audit --detailed

# Check for duplicate memory
claude memory audit --duplicates

# Validate memory references
claude memory audit --references
```

## Advanced Memory Features

### Memory Templates

Template-Based Memory Initialization:

```markdown
<!-- memory/project-template.md -->

# Project Memory Template

## Project Structure

- Name: {{project.name}}
- Type: {{project.type}}
- Language: {{project.language}}

## Team Configuration

- Team size: {{team.size}}
- Workflow: {{team.workflow}}
- Review policy: {{team.review_policy}}

## Quality Standards

- Test coverage: {{quality.test_coverage}}%
- Documentation: {{quality.documentation_level}}
- Security: {{quality.security_level}}
```

Template Instantiation:

```bash
# Create memory from template
claude memory init --template web-app --config project.json

# Variables in project.json:
# {
# "project": {"name": "MyApp", "type": "web-app", "language": "TypeScript"},
# "team": {"size": 5, "workflow": "github-flow", "review_policy": "required"},
# "quality": {"test_coverage": 90, "documentation_level": "comprehensive", "security_level": "high"}
# }
```

### Memory Sharing and Distribution

Team Memory Distribution:

```bash
# Export memory for team sharing
claude memory export --team --format archive

# Import shared memory
claude memory import --team --file team-memory.tar.gz

# Merge memory updates
claude memory merge --base current --update team-updates
```

Memory Distribution Channels:

- Git Repository: Version-controlled memory files
- Package Distribution: Memory bundled with tools/libraries
- Network Share: Centralized memory server
- Cloud Storage: Distributed memory storage

## Best Practices

### Memory Organization

Structural Guidelines:

- Keep memory files focused on single topics
- Use consistent naming conventions
- Implement clear hierarchy and relationships
- Maintain cross-references and links

Content Guidelines:

- Write memory content in clear, concise language
- Use structured formats (markdown, JSON, YAML)
- Include examples and use cases
- Provide context and usage instructions

### Performance Optimization

Memory Loading Optimization:

- Load memory files on-demand when possible
- Implement caching for frequently accessed memory
- Use compression for large memory files
- Preload critical memory files

Memory Access Patterns:

- Group related memory access operations
- Minimize memory file loading frequency
- Use memory references instead of duplication
- Implement lazy loading for optional memory

### Security and Privacy

Memory Security:

- Never store sensitive credentials in memory files
- Implement access controls for memory files
- Use encryption for confidential memory content
- Regular security audits of memory content

Privacy Considerations:

- Separate personal and project memory appropriately
- Use anonymization for sensitive data in shared memory
- Implement data retention policies for memory content
- Respect user privacy preferences in memory usage

This comprehensive reference provides all the information needed to effectively implement, manage, and optimize Claude Code Memory systems for projects of any scale and complexity.
