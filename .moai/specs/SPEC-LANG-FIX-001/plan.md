# Implementation Plan: @SPEC:LANG-FIX-001

## Overview

This document outlines the 6-phase implementation strategy for fixing the language localization system in MoAI-ADK. The plan is designed to be executed sequentially, with some phases supporting parallel work.

**Total Estimated Time**: 6-7 hours

---

## Phase 1: Python Core Fixes (CRITICAL PATH - 1 hour)

**Objective**: Fix foundational Python configuration reading that blocks all language flow

### Phase 1.1: Fix phase_executor.py Config Path (30 minutes)

**File**: `src/moai_adk/core/project/phase_executor.py`

**Current Code (Line 154)**:
```python
"CONVERSATION_LANGUAGE": config.get("conversation_language", config.get("locale", "en")),
```

**Change Required**:
```python
# Add nested config reading
language_config = config.get("language", {})
context_vars = {
    # ... existing variables ...
    "CONVERSATION_LANGUAGE": language_config.get("conversation_language", "en"),
    "CONVERSATION_LANGUAGE_NAME": language_config.get("conversation_language_name", "English"),
}
```

**Testing**:
- [ ] Create unit test: `test_phase_executor_reads_nested_language_config()`
- [ ] Test default value when language section missing
- [ ] Test template variable substitution works

### Phase 1.2: Update processor.py Template Variables (20 minutes)

**File**: `src/moai_adk/core/template/processor.py`

**Requirement**: Ensure `{{CONVERSATION_LANGUAGE_NAME}}` is available in template substitution

**Change Required**:
```python
# Verify context includes both language variables
context = {
    "CONVERSATION_LANGUAGE": config.language.conversation_language,
    "CONVERSATION_LANGUAGE_NAME": config.language.conversation_language_name,
    # ... other variables ...
}
```

**Testing**:
- [ ] Template substitution integration test
- [ ] Verify all commands can access `{{CONVERSATION_LANGUAGE}}`

### Phase 1.3: Write Unit Tests (10 minutes)

**Location**: `tests/unit/test_language_config.py` (new)

```python
def test_reads_nested_language_config():
    """phase_executor reads from config.language.conversation_language"""

def test_defaults_to_english_when_missing():
    """When language config missing, defaults to 'en'"""

def test_includes_language_name():
    """CONVERSATION_LANGUAGE_NAME variable is included"""

def test_template_substitution():
    """{{CONVERSATION_LANGUAGE}} substitutes correctly"""
```

**Deliverable**: ✅ Phase 1 complete when all Python tests pass

---

## Phase 2: Configuration Updates (CONDITIONAL - 45 minutes)

**Objective**: Update configuration system to support nested language structure

**Prerequisite**: Phase 1 MUST be complete

### Phase 2.1: Create/Update config.json Template (20 minutes)

**File**: `src/moai_adk/templates/config.json.template`

**Required Structure**:
```json
{
  "project": {
    "name": "{{PROJECT_NAME}}",
    "owner": "{{PROJECT_OWNER}}",
    "mode": "personal"
  },
  "language": {
    "conversation_language": "{{SELECTED_LANGUAGE}}",
    "conversation_language_name": "{{LANGUAGE_NAME}}"
  },
  "localization": {
    "supported_languages": ["en", "ko", "ja", "zh", "es"]
  }
}
```

**Testing**:
- [ ] New installations generate correct nested config
- [ ] Config validation passes

### Phase 2.2: Add Migration Logic (20 minutes)

**File**: `src/moai_adk/core/config/migration.py` (new)

**Requirement**: Support legacy flat config.json during transition

```python
def migrate_config_to_nested_structure(config):
    """Convert flat config to nested language structure"""
    if "conversation_language" in config and "language" not in config:
        # Migrate flat structure to nested
        config["language"] = {
            "conversation_language": config.pop("conversation_language"),
            "conversation_language_name": config.pop("conversation_language_name", "English")
        }
    return config
```

**Testing**:
- [ ] Migration preserves language setting
- [ ] Existing projects continue to work
- [ ] No data loss during migration

### Phase 2.3: Create Config Schema Documentation (5 minutes)

**File**: `.moai/memory/language-config-schema.md` (new)

**Content**:
- Complete nested config structure
- Supported languages with codes
- Default values
- Migration instructions

**Deliverable**: ✅ Phase 2 complete when config tests pass

---

## Phase 3: Agent Instructions Update (PARALLEL - 1.5 hours)

**Objective**: Add language handling instructions to all 12 agents

**Prerequisite**: Phase 1 complete (Python fixes)

**Can Run Parallel With**: Phase 2 (independent)

### Phase 3.1: Add Language Section to 7 Missing Agents (1 hour)

**Agents to Update**:
1. debug-helper.md
2. quality-gate.md
3. tag-agent.md
4. trust-checker.md
5. git-manager.md
6. cc-manager.md
7. skill-factory.md

**Template to Add**:
```markdown
## 🌍 Language Handling

**IMPORTANT**: You will ALWAYS receive prompts in **English**, regardless of user's original conversation language.

Alfred translates user requests to English before invoking you. This ensures:
- ✅ Perfect skill trigger matching (English Skill descriptions match English prompts 100%)
- ✅ Consistent internal communication
- ✅ Global multilingual support (Korean, Japanese, Chinese, Spanish, etc.)

**Do not try to infer user's original language from your prompt.** Always work in English.

**Output Language Rule**: If prompt includes `conversation_language` parameter:
- `conversation_language="ko"` → Generate output in **Korean**
- `conversation_language="ja"` → Generate output in **Japanese**
- `conversation_language="zh"` → Generate output in **Chinese**
- `conversation_language="es"` → Generate output in **Spanish**
- Otherwise → Generate output in **English**

**Mixed Language Guidance**:
- ✅ ALWAYS user's language: User-facing content, document body, explanations
- ✅ ALWAYS English: Code, commit messages, @TAG identifiers, YAML frontmatter, technical keywords
```

**Process**:
- [ ] Add to each agent after existing sections
- [ ] Ensure positioning is consistent
- [ ] Review for clarity

**Testing**:
- [ ] All 7 agents have language section
- [ ] Language section identical across all agents

### Phase 3.2: Standardize 5 Existing Agent Sections (20 minutes)

**Agents to Review**:
1. project-manager.md
2. spec-builder.md
3. implementation-planner.md
4. tdd-implementer.md
5. doc-syncer.md

**Action**: Verify existing sections match template, update if format differs

**Testing**:
- [ ] All 5 agents have consistent language section
- [ ] Match new template format

### Phase 3.3: Mirror Changes to 12 Agent Templates (30 minutes)

**Files**: `src/moai_adk/templates/.claude/agents/alfred/*.md` (12 files)

**Process**:
- [ ] Copy all changes from active agents to templates
- [ ] Ensure exact match (use diff to verify)

**Testing**:
- [ ] Template comparison tests pass
- [ ] No drift between active and template files

**Deliverable**: ✅ Phase 3 complete when all 12 agents + 12 templates have language sections

---

## Phase 4: Command File Updates (SEQUENTIAL AFTER Phase 3 - 1 hour)

**Objective**: Add language parameters to all Task() invocations in commands

**Prerequisite**: Phase 3 MUST be complete (agents must be ready)

**Can Run Parallel With**: Phase 2 (independent)

### Phase 4.1: Update 4 Command Files (40 minutes)

**Files to Update**:
1. `.claude/commands/alfred/0-project.md` - project-manager invocation (~line 482)
2. `.claude/commands/alfred/1-plan.md` - spec-builder invocations (~lines 143, 190)
3. `.claude/commands/alfred/2-run.md` - tdd-implementer invocation
4. `.claude/commands/alfred/3-sync.md` - doc-syncer invocation

**Pattern for Each Task() Call**:

**Before**:
```markdown
Call the Task tool:
- subagent_type: "spec-builder"
- description: "Analyze the plan and establish a plan"
- prompt: "Please analyze the project document and suggest SPEC candidates..."
```

**After**:
```markdown
Call the Task tool:
- subagent_type: "spec-builder"
- description: "Analyze the plan and establish a plan"
- prompt: """You are spec-builder agent.

LANGUAGE CONFIGURATION:
- conversation_language: {{CONVERSATION_LANGUAGE}}
- language_name: {{CONVERSATION_LANGUAGE_NAME}}

CRITICAL INSTRUCTION:
All SPEC documents (spec.md, plan.md, acceptance.md) MUST be generated in the conversation_language.
If conversation_language is Korean (ko), ALL narrative text must be in Korean.
If conversation_language is Japanese (ja), ALL narrative text must be in Japanese.

TASK:
Please analyze the project document and suggest SPEC candidates...
"""
```

**Specific Locations**:
- **0-project.md**: Search for "project-manager" Task() invocation, add language config to prompt
- **1-plan.md**: Find STEP 1 and STEP 2 spec-builder invocations, add language config to both
- **2-run.md**: Find tdd-implementer invocation, add language config
- **3-sync.md**: Find doc-syncer invocation, add language config

**Testing**:
- [ ] Each command loads templates with {{CONVERSATION_LANGUAGE}}
- [ ] Variable substitution works correctly
- [ ] Task() calls include language configuration

### Phase 4.2: Mirror Changes to 4 Command Templates (20 minutes)

**Files**: `src/moai_adk/templates/.claude/commands/alfred/{0,1,2,3}-*.md` (4 files)

**Process**:
- [ ] Copy all changes from active commands to templates
- [ ] Ensure exact match at Task() invocation points

**Testing**:
- [ ] Template comparison tests pass

**Deliverable**: ✅ Phase 4 complete when all 4 commands + 4 templates updated with language parameters

---

## Phase 5: Integration Tests & E2E Testing (1 hour)

**Objective**: Verify entire language flow works end-to-end

**Prerequisite**: Phases 1-4 complete

### Phase 5.1: Write Integration Tests (30 minutes)

**Location**: `tests/integration/test_command_language_flow.py` (new)

```python
@pytest.mark.parametrize("language", ["ko", "ja", "zh", "es"])
def test_command_passes_language_to_agent(language, tmp_path):
    """All commands pass language parameter to agents via Task()"""
    # 1. Create config with specific language
    # 2. Run command
    # 3. Verify Task() call includes conversation_language parameter

def test_template_substitution_with_language():
    """{{CONVERSATION_LANGUAGE}} substitutes in commands"""

def test_legacy_config_migration():
    """Flat config converts to nested structure"""
```

**Coverage**:
- [ ] Each command (0, 1, 2, 3) passes language
- [ ] All supported languages
- [ ] Config migration path

### Phase 5.2: E2E Tests for Korean Workflow (20 minutes)

**Test Scenario**: `test_korean_user_complete_workflow()`

```gherkin
Feature: Korean user complete workflow
  Scenario: Initialize project, create SPEC, implement
    Given user selected Korean (ko) as conversation_language
    When user runs /alfred:0-project
    Then config.json includes conversation_language: "ko"
    And product.md, structure.md, tech.md are generated in Korean

    When user runs /alfred:1-plan "사용자 인증"
    Then SPEC directory created: .moai/specs/SPEC-AUTH-001/
    And spec.md contains Korean requirements text
    And plan.md contains Korean implementation plan
    And acceptance.md contains Korean acceptance criteria

    When user runs /alfred:2-run SPEC-AUTH-001
    Then TDD implementation creates English code with Korean comments
```

**Testing Tools**:
- [ ] File content verification (check Korean text present)
- [ ] Config validation
- [ ] Directory structure verification

### Phase 5.3: E2E Tests for Japanese & Spanish (10 minutes)

**Scenarios**:
- [ ] Japanese user workflow (verify Japanese text in outputs)
- [ ] Spanish user workflow (verify Spanish text in outputs)

**Deliverable**: ✅ Phase 5 complete when all integration and E2E tests pass

---

## Phase 6: Documentation & Final Validation (45 minutes)

**Objective**: Update documentation and validate complete implementation

**Prerequisite**: All previous phases complete

### Phase 6.1: Update CLAUDE.md (15 minutes)

**Files**:
- `CLAUDE.md`
- `src/moai_adk/templates/CLAUDE.md`

**Updates Required**:

1. **Update Section**: "🌍 Alfred's Language Boundary Rule" (currently lines 123-140)
   - Verify documentation matches actual implementation
   - Update any outdated claims about translation layers

2. **Add Section**: "⚠️ Known Limitations"
   ```markdown
   ## ⚠️ Implementation Status (v0.7.0)

   As of this version, language localization is FULLY IMPLEMENTED:
   - ✅ Python config reading (nested structure)
   - ✅ All commands pass language to agents
   - ✅ All agents support language output
   - ✅ Template variable substitution works
   - ✅ Legacy config migration supported
   ```

3. **Add Section**: "Configuration Reference"
   - Link to `.moai/memory/language-config-schema.md`

**Testing**:
- [ ] Documentation matches code
- [ ] No outdated claims remain
- [ ] Links are correct

### Phase 6.2: Create/Update Configuration Documentation (15 minutes)

**File**: `.moai/memory/language-config-schema.md` (if not created in Phase 2)

**Content**:
```markdown
# Language Configuration Schema

## Config Structure
```yaml
language:
  conversation_language: "ko"  # Language code
  conversation_language_name: "한국어"  # Display name
```

## Supported Languages
- en: English
- ko: 한국어 (Korean)
- ja: 日本語 (Japanese)
- zh: 中文 (Chinese)
- es: Español (Spanish)

## Migration from Legacy Config
...
```

### Phase 6.3: Final Validation Checklist (15 minutes)

**Validation Tasks**:
- [ ] All 40 files modified as planned
- [ ] Python tests pass (config reading)
- [ ] Integration tests pass (command→agent flow)
- [ ] E2E tests pass (complete workflows in 3+ languages)
- [ ] Template files match active files exactly
- [ ] Documentation updated and accurate
- [ ] No hardcoded English in localization-critical code

**Quality Gates**:
- [ ] Code coverage ≥ 85% for language code
- [ ] All tests pass
- [ ] No merge conflicts
- [ ] SPEC acceptance criteria met

**Deliverable**: ✅ Phase 6 complete and ready for `/alfred:3-sync`

---

## Dependency Graph

```
Phase 1: Python Fixes (CRITICAL - 1 hour)
  ↓ (blocks)
Phase 2: Config Updates (45 min)  ← Can start Phase 3 in parallel
  ↓                                   ↓
Phase 3: Agent Instructions (1.5 hours) ← depends on Phase 1
  ↓ (must complete before Phase 4)
Phase 4: Command Updates (1 hour) ← depends on Phase 3
  ↓ (all must complete before Phase 5)
Phase 5: Integration & E2E Tests (1 hour)
  ↓
Phase 6: Documentation & Validation (45 min)
```

**Optimal Execution**:
1. Start Phase 1 immediately (critical path)
2. While Phase 1 running, prepare Phase 2 changes
3. When Phase 1 done, start Phase 2 & 3 in parallel
4. When Phase 3 done, start Phase 4
5. When Phase 4 done, start Phase 5
6. When Phase 5 done, start Phase 6

**Total Time: 6-7 hours** (with parallel execution of Phases 2 & 3)

---

## Risk Mitigation

### Risk 1: Legacy Config Breaking
- **Mitigation**: Phase 2 includes migration logic
- **Testing**: Explicit legacy config test case

### Risk 2: Template Drift
- **Mitigation**: Phase 4 includes template comparison
- **Testing**: Template equivalence tests

### Risk 3: Sub-agent Instructions Inconsistency
- **Mitigation**: Phase 3 uses standardized template
- **Testing**: Language section comparison tests

### Risk 4: Incomplete Coverage
- **Mitigation**: Comprehensive file inventory in spec.md
- **Testing**: Phase 6 validation checklist covers all files

---

## Success Criteria

**Functional**:
- [ ] Korean user can create project and SPEC in Korean
- [ ] Japanese user can create project and SPEC in Japanese
- [ ] Spanish user can create project and SPEC in Spanish
- [ ] Config.json stores language in nested structure
- [ ] All 12 agents have language handling instructions
- [ ] All 4 commands pass language parameters

**Quality**:
- [ ] ≥ 85% test coverage for language code
- [ ] All unit, integration, E2E tests pass
- [ ] Documentation matches implementation
- [ ] No hardcoded English in critical paths

**Compliance**:
- [ ] SPEC acceptance criteria met
- [ ] All files modified as planned
- [ ] No breaking changes to existing functionality

---

## Next Steps

When implementation complete:
1. Run `/alfred:3-sync` to synchronize documentation
2. Create GitHub PR for code review
3. Merge to `develop` branch
4. Tag release as v0.7.0

---

**Plan Status**: Ready for Phase 1 execution
**Estimated Completion**: 6-7 hours from start
