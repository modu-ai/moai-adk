# MoAI-ADK Local Claude Code Development Guide

## Development Workflow

### 1.1 Work Location Rules

**All development work must be performed in the following location:**

```
/Users/goos/MoAI/MoAI-ADK/src/moai_adk/
├── .claude/                 # Claude Code configuration
├── .moai/                   # MoAI project metadata
├── templates/               # Project templates
└── [Other source code]
```

**After work, synchronize to local project:**

```
/Users/goos/MoAI/MoAI-ADK/
├── .claude/                 # Synchronized
├── .moai/                   # Synchronized
└── [Source code and documentation]
```

### 1.2 Development Cycle

```
1. Work in source project (/src/moai_adk/...)
   ↓
2. Synchronize to local project (./)
   ↓
3. Test and validate in local project
   ↓
4. Git commit (in local project)
```

---

## File Synchronization Rules

### 2.1 Synchronization Target Directories

**Areas requiring automatic synchronization:**

```
src/moai_adk/.claude/    ↔  .claude/
src/moai_adk/.moai/      ↔  .moai/
src/moai_adk/templates/  ↔  ./
```

### 2.2 Synchronization Exclusions (Local Only)

**Files that must NOT be synchronized:**

```
.claude/commands/moai/99-release.md          # Local release command only
.claude/settings.local.json                  # Personal settings
.claude/hooks/                               # Development hooks (forbidden in package)
.CLAUDE.local.md                             # This file
.moai/cache/                                 # Cache files
.moai/logs/                                  # Log files
.moai/config/config.json                     # Personal project settings
```

### 2.3 .moai/config Synchronization (Manual)

**Synchronization targets:**

| File/Folder              | Direction       | Description                        |
| ------------------------ | --------------- | ---------------------------------- |
| `presets/*.json`         | Local → Package | Deployment templates (manual copy) |
| `statusline-config.yaml` | Bidirectional   | Shared settings (only when needed) |
| `config.json`            | ❌ Forbidden    | User settings, do not synchronize  |

**Presets Synchronization (During package release):**

```bash
# Copy local presets to package
mkdir -p src/moai_adk/templates/.moai/config/presets
cp .moai/config/presets/*.json src/moai_adk/templates/.moai/config/presets/

# Verify changes
git status src/moai_adk/templates/.moai/config/presets/

# Commit
git add src/moai_adk/templates/.moai/config/presets/
git commit -m "config: Update preset templates"
```

**⚠️ Important Notes:**

- NEVER synchronize `config.json`
  - Local: User customized settings (38 lines)
  - Package: Base template (341 lines, all options included)
- Presets synchronization needed only before package release

---

## Code Writing Standards

### 3.1 Language Rules

**All code work:**

- ✅ **Write in English only**
- ✅ Variable names: camelCase or snake_case (per language convention)
- ✅ Function names: camelCase (JavaScript/Python) or PascalCase (C#/Java)
- ✅ Class names: PascalCase (all languages)
- ✅ Constant names: UPPER_SNAKE_CASE (all languages)

**Comments and documentation:**

- ✅ **All comments in English**
- ✅ JSDoc, docstrings, etc. all in English
- ✅ Commit messages: English (or mixed Korean + English with format: English)

**This file (CLAUDE.local.md):**

- ✅ **Written in English** (updated for consistency)
- ✅ Tracked in Git

### 3.2 Comment Standards (English)

- All code, output messages, and comments must be written in English

### 3.3 Forbidden Patterns

```python
# ❌ WRONG - Korean comments
def calculate_score():  # 점수 계산
    score = 100  # 최종 점수
    return score

# ✅ CORRECT - English comments
def calculate_score():  # Calculate final score
    score = 100  # Final score value
    return score
```

---

## Local-Only File Management

### 5.1 Local-Only File List

**Files that must NEVER be synchronized to the package:**

| File                  | Location                 | Purpose                 | Git Tracked |
| --------------------- | ------------------------ | ----------------------- | ----------- |
| `99-release.md`       | `.claude/commands/moai/` | Local release command   | ✅ Yes      |
| `CLAUDE.local.md`     | Root                     | Local development guide | ✅ Yes      |
| `settings.local.json` | `.claude/`               | Personal settings       | ❌ No       |
| `cache/`              | `.moai/`                 | Cache files             | ❌ No       |
| `logs/`               | `.moai/`                 | Log files               | ❌ No       |
| `config/config.json`  | `.moai/`                 | Personal settings       | ❌ No       |

### 5.2 Local Release Command

**.claude/commands/moai/99-release.md (Local only):**

```markdown
# Local Release Management

This command is only for local development and testing.
It manages MoAI-ADK package releases locally.

## Features

- Version management
- Pre-release testing
- Local deployment simulation
- Changelog generation

## Usage

> /moai:99-release

This command is NOT synchronized to the package.
```

---

### 6.3 Git Work Checklist

**Before committing:**

- [ ] All code written in English
- [ ] Comments and docstrings in English
- [ ] Local-only files not included
- [ ] Tests passing
- [ ] Linting passing (ruff, pylint, etc.)

**Before pushing:**

- [ ] Branch rebased to latest development version
- [ ] Commits organized into logical units
- [ ] Commit messages follow standard format

**Before PR:**

- [ ] Documentation synchronized
- [ ] SPEC updated (if needed)
- [ ] Changes documented

---

## Frequently Used Commands

### Synchronization

```bash
# Synchronize from source to local
bash .moai/scripts/sync-from-src.sh

# Synchronize specific directories only
rsync -avz src/moai_adk/.claude/ .claude/
rsync -avz src/moai_adk/.moai/ .moai/
```

### Validation

```bash
# Check code quality
ruff check src/
mypy src/

# Run tests
pytest tests/ -v --cov

# Validate documentation
python .moai/tools/validate-docs.py
```

---

## CLAUDE.md Writing and Maintenance Guide

### Overview

This guide describes how to write and maintain the CLAUDE.md file for MoAI-ADK.
This document is for developers working on the MoAI framework itself.

---

### The Nature of CLAUDE.md

**Important**: CLAUDE.md is **not code**. CLAUDE.md is **Alfred's fundamental execution directives**.

- ✅ **Purpose**: Orchestration rules for Claude Code agents
- ❌ **Not for**: User guides, implementation guides, tutorials
- 👥 **Audience**: Claude Code (agents, commands, hooks)
- ❌ **Not for**: End users

**CLAUDE.md vs. Other Documents**:

| Document           | Purpose                   | Audience          |
| ------------------ | ------------------------- | ----------------- |
| CLAUDE.md          | Alfred execution rules    | Agents/Commands   |
| README.md          | Project overview          | End users         |
| Skill SKILL.md     | Pattern/knowledge capsule | Agents/Developers |
| .moai/memory/\*.md | Execution rules reference | Agents/Developers |
| CLAUDE.local.md    | Local work guide          | Local developers  |

---

### 1. CLAUDE.md Structure Standards

All CLAUDE.md files MUST include the following 8 sections:

#### I. Purpose & Scope (Required)

```markdown
# [PROJECT]: Claude Code Execution Guide

**Purpose**: Super Agent Orchestrator execution manual for [PROJECT]
**Audience**: Claude Code (agents, commands), NOT end users
**Philosophy**: [Philosophy statement]
```

**Must include:**

- ✅ Clear purpose statement
- ✅ "Audience: Claude Code agents"
- ✅ "NOT for end users"
- ✅ Scope in/out specification

#### II. Core Principles (Required)

3-5 fundamental operating rules:

```markdown
## Core Principles

1. **[Principle name]** - Description
2. **[Principle name]** - Description
3. **[Principle name]** - Description
```

#### III. Configuration Integration (Conditional)

Connection to config.json:

```markdown
## Configuration Integration

Config fields this document reads:

- `github.spec_git_workflow` - Git workflow style
- `constitution.test_coverage_target` - Quality gate threshold

### Config Field Specification

**Field**: `github.spec_git_workflow`

- **Location**: config.json → github.spec_git_workflow
- **Type**: String (enum)
- **Possible values**: develop_direct, feature_branch, per_spec
- **Default value**: develop_direct
- **Priority**: Priority 1 (highest)
- **Impact**: Controls Git branch creation
```

#### IV. Auto-Trigger Rules (Conditional)

Conditions for automatic Agent/Command execution:

````markdown
## Agent: [AGENT_NAME] - Auto-Trigger Rules

### Trigger Activation Points

| Phase | Event | Condition | Config Field | Delegation Pattern |
| ----- | ------------ | ---- | ------------------------------ | --------- |
| PLAN | /moai:1-plan | Always | language.conversation_language | Direct call |
| RUN | /moai:2-run | Always | constitution.enforce_tdd | Direct call |

### Trigger Logic (Pseudo-code)

```python
def should_trigger(event, config):
    if event.type == "moai:1-plan":
        return True  # Always trigger
    elif event.type == "vague_request":
        return measure_clarity(event) < 70%
    return False
```

### Context to Pass

Pass the following information on trigger:

1. `user_request` - Original user request
2. `current_phase` - Current phase (PLAN/RUN/SYNC)
3. `config` - User config.json
4. `previous_results` - Previous phase results (if available)
````

#### V. Delegation Hierarchy (Required)

Which agent to call and when:

```markdown
## Delegation Hierarchy

- **spec-builder**: SPEC generation and analysis

  - Condition: /moai:1-plan execution
  - Context: User request + config

- **git-manager**: Git branch creation
  - Condition: spec_git_workflow != "develop_direct"
  - Context: SPEC ID + git config

### Delegation Error Handling

If git-manager call fails:

1. Log error
2. Present options to user via AskUserQuestion
3. Retry or skip based on selection
```

#### VI. Quality Gates (Required)

TRUST 5 or similar criteria:

```markdown
## Quality Gates (TRUST 5)

### Test-first

**Criterion**: ≥ 85% test coverage
**Validation**: pytest --cov=src/ | grep "Coverage"
**Failure**: Block PR, report coverage gaps

### Readable

**Criterion**: Clear naming (no ambiguous abbreviations)
**Validation**: ruff linter automatic check
**Failure**: Warning (not blocking)

### Unified

**Criterion**: Project pattern compliance (consistent style)
**Validation**: black, isort automatic check
**Failure**: Auto-format or warning

### Secured

**Criterion**: OWASP security check passed
**Validation**: security-expert agent review (required)
**Failure**: Block PR

### Trackable

**Criterion**: Clear commit messages + test evidence
**Validation**: Git commit message regex validation
**Failure**: Suggest message format
```

#### VII. Reference Documents (Required)

External document references:

```markdown
## Reference Documents

### Required References

- @.moai/memory/execution-rules.md - Execution constraints
- @.moai/memory/agents.md - Agent catalog
- @.moai/config/config.json - Config schema

### Recommended References

- Skill("moai-spec-intelligent-workflow") - SPEC decision logic
- Skill("moai-cc-configuration") - Config management
- @.moai/memory/token-optimization.md - Token budget
```

**Reference format (use exactly this format):**

- ✅ `@.moai/memory/agents.md` (File reference)
- ✅ `Skill("moai-cc-commands")` (Skill reference)
- ✅ `/moai:1-plan` (Command reference)
- ❌ `.moai/memory/agents.md` (Missing @)
- ❌ `moai-cc-commands` (Not wrapped in Skill())

#### VIII. Quick Reference & Examples (Required)

Real usage examples:

````markdown
## Example Scenario 1: Personal + develop_direct

**Configuration**:

```json
{
  "git_strategy": { "mode": "personal" },
  "github": { "spec_git_workflow": "develop_direct" }
}
```
````

**Expected behavior:**

- ✅ /moai:1-plan generates SPEC file
- ✅ git-manager not called
- ✅ Branch not created
- ✅ Can commit directly to current branch

---

### 2. Forbidden Content (What NOT to include in CLAUDE.md)

❌ **Never include:**

- ❌ User guides or tutorials
- ❌ Implementation code examples (flowcharts OK)
- ❌ Marketing language
- ❌ Duplicate content from Skills/memory/
- ❌ API implementation details (reference Skills instead)
- ❌ Hardcoded secrets or credentials

---

### 3. Writing Style Guidelines

#### Tone & Voice

- ✅ Direct, technical, clear
- ✅ Imperative: "Alfred MUST NOT directly execute tasks" (not passive)
- ✅ Completeness > brevity
- ✅ Define terms on first use

**Bad example:**

```text
Alfred는 아마도 작업을 실행해야 할 것 같습니다.
```

**Good example:**

```text
Alfred DOES NOT execute tasks directly. Alfred DELEGATES to specialized agents.
```

#### Technical Clarity

| Situation | Format |
| ------------------------ | ----------------------------- |
| Decision matrix (3+) | Use tables |
| Complex logic | ASCII flowchart or Pseudo-code |
| Config examples | Complete JSON/YAML blocks |
| Rules/constraints | Bullet lists |
| Sequential procedures | Numbered lists |

**When Pseudo-code is OK:**

```python
# OK: Shows decision logic
if config["spec_git_workflow"] == "develop_direct":
    TRIGGER_GIT_MANAGER = False
else:
    TRIGGER_GIT_MANAGER = True
```

**Implementation code should reference Skills:**

```markdown
# WRONG

def validate_configuration(config):
schema = ConfigSchema()
return schema.validate(config)

# RIGHT

Validation is handled in moai-cc-configuration Skill.
Details: @.moai/memory/configuration-validation.md
```

---

### 4. CLAUDE.md Validation Checklist

**Required checks before committing CLAUDE.md:**

- [ ] **Purpose clear**: Start with "Alfred's fundamental execution directives"
- [ ] **Audience stated**: "For Claude Code agents"
- [ ] **8 sections**: All included (or conditional sections justified)
- [ ] **No duplication**: No overlapping content with Skills/memory/
- [ ] **Valid config references**: All fields exist in schema
- [ ] **Accurate agent names**: Only agents that exist in .claude/agents/
- [ ] **Reference format**: `@.moai/` or `Skill()` format
- [ ] **Example validity**: JSON/YAML examples are syntactically correct
- [ ] **No secrets**: No API keys or credentials
- [ ] **Explicit end**: Ends with "For Claude Code execution"

---

### 5. Memory/Reference Document Standards

`.moai/memory/` document structure:

```markdown
# [Title]

**Purpose**: Single-line purpose (max 30 characters)
**Audience**: [Agents / Humans / Developers]
**Last Updated**: YYYY-MM-DD
**Version**: X.Y.Z

## Quick Reference (30 seconds)

One-paragraph summary. Agents read this section first.

---

## Implementation Guide (5 minutes)

Structured implementation instructions:

### Features

- Feature 1
- Feature 2

### When to Use

- Use in scenario 1
- Use in scenario 2

### Core Patterns

- Pattern 1
- Pattern 2

---

## Advanced Implementation (10+ minutes)

In-depth explanations, complex scenarios, edge cases

---

## References & Examples

Complete examples, code snippets, detailed references
```

---

### 6. Skill SKILL.md Standard

```markdown
---
name: moai-[domain]-[skill-name]
description: [One-line description - max 15 words]
---

## Quick Reference (30 seconds)

One paragraph.

---

## Implementation Guide

### Features

[Feature list]

### When to Use

[Use cases]

### Core Patterns

[Patterns and examples]

---

## Advanced Implementation (Level 3)

[Complex patterns, edge cases]

---

## References & Resources

[Complete API reference, examples, links]
```

**Skill Naming Conventions**:

```text
moai-cc-[feature-name]      # Claude Code related
moai-foundation-[concept]   # Shared concepts
moai-[language]-[feature]   # Language-specific features
```

Examples:

- moai-cc-commands (Claude Code commands)
- moai-foundation-trust (TRUST 5 framework)
- moai-lang-python (Python specific)

---

### 7. How Agents Read CLAUDE.md

Agents extract information in this order:

1. **What can I do?** (Permissions section)

   - Tool allow/block list
   - Maximum token budget
   - Execution constraints

2. **When do I run automatically?** (Auto-trigger section)

   - Trigger conditions
   - Event types
   - Config dependencies

3. **Who do I call?** (Delegation section)

   - Sub-agents to invoke
   - When to invoke each
   - Context to pass

4. **How do I know I succeeded?** (Quality gate section)
   - Pass criteria
   - Failure handling
   - Validation steps

---

### 8. Config Field Reference Pattern

Format to use when referencing config in CLAUDE.md:

```markdown
### Config: github.spec_git_workflow

**Field Path**: config.json → github → spec_git_workflow
**Type**: String (enum)
**Possible values**: develop_direct, feature_branch, per_spec
**Default value**: develop_direct
**Priority**: Priority 1 (highest)

**Impact**:

- Controls Git branch creation
- Determines git-manager auto-trigger
- Controls PHASE 3 execution

**Validation Rules**:

- Must be one of the enum values
- If missing: use default value develop_direct
- Invalid value: warn and use default

**Related Fields**:

- `git_strategy.mode` (fallback)
- `github.spec_git_workflow_configured` (validation flag)
```

---

### 9. Updates & Maintenance

#### Version Management

- Tag CLAUDE.md changes with semantic versioning
- Record major changes in root CHANGELOG.md
- Specify version in frontmatter if needed

#### Review Process (Before merging)

1. **Clarity review**: Read aloud (check for ambiguity)
2. **Agent testing**: Can agents clearly extract the rules?
3. **Config validation**: Do config references match schema?
4. **Reference check**: Do external references actually exist?
5. **Example validation**: Can examples run as-is?

#### Archiving Old Content

Old CLAUDE.md sections should be:

- Moved to `.moai/archive/CLAUDE.md.[date]`
- Keep current CLAUDE.md with only active rules

---

## Reference Materials

### Official Documentation

- [Claude Code Official Docs](https://code.claude.com/docs)
- [Claude Code CLI Reference](https://code.claude.com/docs/en/cli-reference)
- [Claude Code Settings Guide](https://code.claude.com/docs/en/settings)
- [MCP Integration Guide](https://code.claude.com/docs/en/mcp)

### MoAI-ADK Documentation

- [CLAUDE.md](./CLAUDE.md) - Claude Code execution guide
- [.moai/memory/](./. moai/memory/) - Reference documents
- [README.md](./README.md) - Project overview

### Related Skills

- `moai-cc-claude-md` - CLAUDE.md writing guide
- `moai-cc-hooks` - Claude Code Hooks system
- `moai-cc-skills-guide` - Skill development guide
- `moai-cc-configuration` - Configuration management guide

---

## Update History

| Date | Version | Changes |
| ---------- | ----- | --------- |
| 2025-11-22 | 1.0.0 | Initial creation |
| 2025-11-25 | 1.1.0 | English translation |

---

**Author**: GOOS
**Project**: MoAI-ADK
**Status**: ✅ Active Document

---

## Markdown Standards & Patterns

### CommonMark Compatibility Rules (Required)

**Rule**: Parentheses must be OUTSIDE bold markers for correct CommonMark rendering.

#### Forbidden Patterns

```markdown
❌ **Text(description)**next - Rendering fails
❌ **текст(описание)**next - Rendering fails
❌ **文本(描述)**next - Rendering fails
```

**Reason**: CommonMark treats `()` as punctuation. Parentheses inside bold markers `**` violate delimiter run rules, causing markup to not close properly.

#### Allowed Patterns

```markdown
✅ **Text**(description) - Recommended (no space)
✅ **Text** (description) - Allowed (with space)
✅ **текст**(описание) - Works in all languages
✅ **文本**(描述) - Works in all languages
```

### Implementation Rules

**When creating/modifying documents:**

```python
import re

# Validation: reject bad patterns
def validate_markdown_pattern(content: str) -> bool:
    bad_pattern = r'\*\*[^*]+\([^)]+\)\*\*[^\s*]'
    return not bool(re.search(bad_pattern, content))

# Normalization: fix spacing
def normalize_bold_parentheses(content: str) -> str:
    # **text** (desc) → **text**(desc)
    pattern = r'\*\*([^*]+)\*\*\s+\(([^)]+)\)'
    return re.sub(pattern, r'**\1**(\2)', content)
```

### Scope of Application

**Required for:**

- README.md (all language versions)
- API documentation
- Technical guides
- docs-manager agent output
- Document synchronization
- All markdown-related skills

**Validation Points:**

- Document creation: pre-validate
- Document synchronization: post-normalize
- Manual editing: markdownlint validation
- CI/CD: automated markdown validation

### Multilingual Support

This rule applies to **all languages** for consistent CommonMark rendering:

- English: `**Feature**(description)`
- Korean: `**기능**(설명)`
- Japanese: `**機能**(説明)`
- Chinese: `**功能**(描述)`
- Russian: `**Функция**(описание)`

**All languages follow the same pattern**: write in format `**Text**(details)` with no space between bold and parentheses.

---

## Important Notes

- `/Users/goos/MoAI/MoAI-ADK/.claude/settings.json` should always use substituted variable values
- Do NOT specify time periods like "4-6 hours investment" - avoid uncertain and unverified timeframes
- When `.claude` content changes, update `.moai/memory` files accordingly

---

## Document Synchronization Information

### Master-Replica Pattern

```
📄 Master (English)
   /src/moai_adk/templates/CLAUDE.md
        ↓ [Professional Translation]
   📄 Replica (English translation)
      ./CLAUDE.md (this file)
        ↓ [Git Pre-commit Hook]
   ✅ Auto-validation & Sync
```

### Synchronization Rules

1. **Master file changes**: Only modify templates/CLAUDE.md (English)
2. **Automatic sync**: Git pre-commit hook detects changes
3. **Translation validation**: Automatic translation quality check
4. **Replica update**: Root CLAUDE.md automatically updated
5. **Metadata**: Synchronization status auto-recorded

### Synchronization Tracking

| Item | Value |
| ---------- | --------- |
| Master Version | 2.2.0 |
| Translation Level | Professional |
| Sync Pattern | Master-Replica + Git Hook |
| Last Sync | 2025-11-25 |
| Next Sync Check | On next commit |
| Validation Status | ✅ Passed |

### Developer Guide

**When modifying master file:**

```bash
# 1. Modify only templates/CLAUDE.md
# 2. Run Git commit
git add src/moai_adk/templates/CLAUDE.md
git commit -m "docs: Update CLAUDE.md (master)"

# 3. Pre-commit hook will automatically:
#    - Generate English translation
#    - Update root CLAUDE.md
#    - Refresh metadata
#    - Run validation
```

**Do NOT modify this file (English replica) directly**
