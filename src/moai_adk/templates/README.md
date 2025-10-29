# MoAI-ADK Project Templates

This directory contains template files that are copied to new projects when users run `moai-adk init`.

## Directory Structure

```
templates/
├── .claude/              # Claude Code configuration
│   ├── agents/          # Alfred sub-agents (12 specialists)
│   ├── commands/        # Slash commands (/alfred:0-4)
│   ├── skills/          # 55 reusable knowledge capsules
│   └── hooks/           # Session lifecycle hooks
├── .moai/               # MoAI-ADK configuration
│   ├── config.json     # Project settings (language, mode, owner)
│   ├── docs/           # Internal documentation
│   ├── memory/         # Session context persistence
│   ├── reports/        # Quality and sync reports
│   └── specs/          # SPEC documents directory
├── .github/             # GitHub Actions workflows
│   └── workflows/      # CI/CD automation
├── CLAUDE.md            # Project guidance for Claude Code
└── .gitignore          # Git ignore patterns
```

---

## Template Components

### 1. `.claude/` - Claude Code Configuration (2.0 MB)

The complete Alfred SuperAgent system:

- **Agents** (12 specialists): spec-builder, tdd-implementer, doc-syncer, tag-agent, trust-checker, debug-helper, implementation-planner, project-manager, quality-gate, git-manager, cc-manager, skill-factory
- **Commands** (4 workflow commands): `/alfred:0-project`, `/alfred:1-plan`, `/alfred:2-run`, `/alfred:3-sync`
- **Skills** (55 knowledge capsules): Foundation (5), Essentials (4), Alfred (7), Domain (7), Language (18), CC (14)
- **Hooks**: SessionStart, PreToolUse, PostToolUse lifecycle guards

### 2. `.moai/` - MoAI-ADK Configuration (120 KB)

Project configuration and memory:

- **config.json**: Language settings, project owner, team mode
- **memory/**: Persistent session context and state
- **docs/**: Internal guides and strategies
- **reports/**: Sync analysis and quality reports
- **specs/**: SPEC document directory structure

### 3. `.github/` - GitHub Workflows (80 KB)

Continuous integration and deployment:

- **workflows/**: Pre-configured GitHub Actions for testing, linting, type checking
- **ISSUE_TEMPLATE/**: Standard issue templates

### 4. `CLAUDE.md` - Project Guidance (15 KB)

The master instruction document for Claude Code:

- **Variable substitution**: `{{PROJECT_OWNER}}`, `{{CONVERSATION_LANGUAGE}}`, `{{CODEBASE_LANGUAGE}}`
- **Customizable**: User-facing project instructions in user's language

---

## How Templates Are Used

### Initialization Flow

When a user runs `moai-adk init`, the `TemplateProcessor` class performs:

1. **Template Discovery**: Locates `src/moai_adk/templates/` via package path resolution
2. **Directory Copy**: Copies `.claude/`, `.moai/`, `.github/` to target project
3. **File Copy**: Copies `CLAUDE.md`, `.gitignore` individually
4. **Variable Substitution**: Replaces template placeholders with user values:
   - `{{PROJECT_OWNER}}` → User's configured name
   - `{{CONVERSATION_LANGUAGE}}` → User's language (Korean, Japanese, etc.)
   - `{{CONVERSATION_LANGUAGE_NAME}}` → Language display name
   - `{{CODEBASE_LANGUAGE}}` → Detected or specified project language

### Current Template Processor Logic

**File**: `src/moai_adk/core/template/processor.py`

**Methods**:
- `_copy_claude()`: Copy Alfred system (agents, commands, skills, hooks)
- `_copy_moai()`: Copy MoAI-ADK configuration and structure
- `_copy_github()`: Copy GitHub Actions workflows
- `_copy_claude_md()`: Copy and substitute variables in CLAUDE.md
- `_copy_gitignore()`: Copy Git ignore patterns

---

## @TAG System & Traceability

### What Are @TAG Markers?

**@TAG markers** are MoAI-ADK's core traceability feature, linking:
- `@SPEC:ID` → Requirements
- `@TEST:ID` → Test cases
- `@CODE:ID` → Implementation
- `@DOC:ID` → Documentation

**Example**:
```python
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001/spec.md | Chain: AUTH-001
def authenticate_user(username, password):
    """Authenticate user credentials (SPEC-AUTH-001)."""
    # Implementation...
```

### How @TAGs Work in New Projects

**Day 1**: Your project starts with **0 @TAGs**
- No setup required, no validation needed
- Tags are created automatically via Alfred commands

**Day 30**: Tags grow naturally
- `/alfred:1-plan "User auth"` → creates `@SPEC:AUTH-001`
- `/alfred:2-run SPEC-AUTH-001` → creates `@TEST:AUTH-001`, `@CODE:AUTH-001`
- `/alfred:3-sync` → creates `@DOC:AUTH-001`

**Simple & Automatic**: @TAGs are added by Alfred agents, not by you.

### TAG Validation (Advanced, Optional)

**MoAI-ADK framework** uses TAG validation (82 files, complex chains)
**Your new project** does NOT need TAG validation

**When you might want TAG validation**:
- Project has 100+ TAGs
- Team has 10+ developers
- Need automated quality gates

**How to add TAG validation** (if needed):
- See MoAI-ADK documentation: "Advanced: TAG Validation Setup"
- Manual installation from framework source
- Optional plugin (future): `moai-adk plugin install tag-validation`

---

## Maintenance Guidelines

### Syncing Skills to Templates

**Script**: `scripts/sync_allowed_tools.py` (Korean)

This script synchronizes `.claude/skills/` to `src/moai_adk/templates/.claude/skills/` to ensure new projects receive the latest skill versions.

**When to run**:
- After adding new skills
- After updating existing skill content
- Before releasing a new version

### Template Versioning

Templates are versioned with the MoAI-ADK package. When releasing:

1. Update template files in `src/moai_adk/templates/`
2. Run skill sync script
3. Update this README if structure changes
4. Test with `moai-adk init` on a clean directory
5. Document breaking changes in CHANGELOG.md

### Testing Template Changes

**Integration tests**: `tests/integration/test_phase_executor.py`

```bash
# Test project initialization
pytest tests/integration/test_phase_executor.py::test_phase_0_initialization

# Test full workflow
pytest tests/e2e/test_e2e_workflow.py
```

---

## Related Documentation

- **CLAUDE.md**: Main project guidance (repository root)
- **CLAUDE-AGENTS-GUIDE.md**: Alfred team structure (repository root)
- **CLAUDE-RULES.md**: Development rules and conventions (repository root)
- **Language Configuration**: `.moai/memory/language-config-schema.md`

---

## FAQ

### Q: What is the difference between MoAI-ADK framework and user projects?

**A**:
- **MoAI-ADK framework** = The tool itself (this repository, needs complex validation)
- **User projects** = Projects created by `moai-adk init` (your apps, simple start)

Think of it like Ruby on Rails:
- Rails framework has complex CI/CD infrastructure
- `rails new my-app` creates a simple, clean project

### Q: Will my project have @TAG markers?

**A**: Yes! @TAG markers are **automatic and core** to MoAI-ADK workflow:
- Created automatically by `/alfred:1-plan`, `/alfred:2-run`, `/alfred:3-sync`
- Provide traceability (link requirements → tests → code → docs)
- No validation needed for small/medium projects (simple grep is enough)

### Q: When do I need TAG validation?

**A**: Only for **mature, large-scale projects**:
- 100+ SPEC documents
- 10+ team members
- Complex TAG chains requiring automated integrity checks

For most projects, @TAG markers alone are sufficient.

### Q: Will old projects break when templates are updated?

**A**: No. Templates only affect **new** projects created via `moai-adk init`. Existing projects are not modified unless explicitly updated via `moai-adk update`.

### Q: How do I customize templates for my organization?

**A**: Fork MoAI-ADK and modify files in `src/moai_adk/templates/`. Maintain your fork's template sync script to merge upstream updates.

---

## Design Philosophy

### Framework vs User Projects

**What MoAI-ADK framework needs** (this repository):
- ✅ Complex TAG validation (82 files, orphan/duplicate detection)
- ✅ Advanced CI/CD workflows
- ✅ Quality gates for framework development

**What new user projects need** (`moai-adk init`):
- ✅ Alfred SuperAgent (workflow orchestration)
- ✅ @TAG markers (traceability)
- ✅ Basic CI/CD (testing, linting)
- ❌ NO complex TAG validation (not needed until project matures)

**Goal**: Start simple, grow as needed.

---

## Change History

| Date | Version | Changes |
|------|---------|---------|
| 2025-10-29 | 0.8.3 | Removed TAG validation system from templates (design clarification) |
| 2025-10-29 | 0.7.0 | Language localization complete (5 languages supported) |
| 2025-10-16 | 0.6.0 | Initial template structure with Alfred system |

---

**Last Updated**: 2025-10-29
**Maintained By**: MoAI-ADK Core Team
**Design Principle**: Start simple, scale as needed
