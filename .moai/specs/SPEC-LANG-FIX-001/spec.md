---
id: LANG-FIX-001
version: 0.0.1
status: closed
created: 2025-10-29
updated: 2025-10-29
author: @Alfred
priority: critical
category: system-architecture
labels: [localization, i18n, multi-language, bug-fix]
scope: complete-rewrite
---

# @SPEC:LANG-FIX-001: Complete Language Localization System Implementation

## HISTORY

### v0.0.1 (2025-10-29)
- **INITIAL**: Initial creation of comprehensive language localization fix specification
- **AUTHOR**: @Alfred
- **SCOPE**: Fix language settings being ignored across all commands, agents, and templates
- **CONTEXT**: Current implementation claims to support multi-language but doesn't pass language configuration to sub-agents, resulting in 100% English output regardless of user language setting

## Environment

### System Environment
- **MoAI-ADK Version**: 0.6.3
- **Python Version**: 3.11+
- **Supported Languages**: English (en), Korean (ko), Japanese (ja), Chinese (zh), Spanish (es)
- **Config Format**: JSON nested structure

### Technology Stack
- **Language Detection**: JSON config parsing
- **Template Engine**: Jinja2 with variable substitution
- **Sub-agent Communication**: Python Task() function with dictionary parameters

### Architecture
- **Three-layer model**: User Conversation / Internal Operations / Skills & Templates
- **Language Flow**: Config → Phase Executor → Task() → Sub-agents → Output

## Assumptions

1. **Existing Installations**: Legacy installations have flat config.json structure
2. **Backward Compatibility**: Must support both old and new config structures during migration
3. **Sub-agent Isolation**: Each sub-agent runs independently, requiring explicit language parameter
4. **Internal Language**: All internal prompts, skills, and technical documentation remain in English
5. **User Language Output**: SPEC documents, project documents, and user-facing content in user's language
6. **Template Processing**: MoAI-ADK uses Python to process templates before CLI execution

## Requirements

### Ubiquitous Requirements

- **REQ-U-001**: The system MUST read `conversation_language` from `.moai/config.json`
- **REQ-U-002**: The system MUST support at least 5 languages (English, Korean, Japanese, Chinese, Spanish)
- **REQ-U-003**: The system MUST preserve user's language choice throughout entire workflow
- **REQ-U-004**: The system MUST document complete language architecture for developers

### Event-Driven Requirements

- **REQ-E-001**: WHEN Alfred initializes a project, THEN it MUST prompt user to select language
- **REQ-E-002**: WHEN Alfred invokes a sub-agent, THEN it MUST pass `conversation_language` parameter
- **REQ-E-003**: WHEN a sub-agent generates documents, THEN it MUST check language parameter and output in specified language
- **REQ-E-004**: WHEN user runs `/alfred:1-plan`, THEN SPEC documents MUST be generated in user's language
- **REQ-E-005**: WHEN user runs `/alfred:0-project`, THEN project documents MUST be generated in user's language

### State-Driven Requirements

- **REQ-S-001**: WHILE language is set to Korean (ko) in config, ALL user-facing output MUST be in Korean
- **REQ-S-002**: WHILE language is set to Japanese (ja) in config, ALL user-facing output MUST be in Japanese
- **REQ-S-003**: WHILE language setting is missing/null, THE system MUST default to English (en)
- **REQ-S-004**: WHILE processing templates, THE system MUST have access to `{{CONVERSATION_LANGUAGE}}` variable

### Optional Requirements

- **REQ-OPT-001**: The system MAY support language selection via environment variables
- **REQ-OPT-002**: The system MAY provide language detection based on system locale (fallback)
- **REQ-OPT-003**: The system MAY cache language setting in memory for performance

### Constraints

- **REQ-C-001**: IF language parameter is missing, THEN system MUST NOT assume or infer language from prompts
- **REQ-C-002**: IF template variable `{{CONVERSATION_LANGUAGE}}` is not substituted, THEN command execution MUST fail with clear error
- **REQ-C-003**: IF config.json has legacy flat structure, THEN migration logic MUST convert to nested structure
- **REQ-C-004**: IF sub-agent receives language parameter, THEN it MUST generate output in that language (NOT guess from prompt)
- **REQ-C-005**: ALL internal prompts, skills, code comments, and technical documentation MUST remain in English

## Traceability (@TAG)

- **SPEC**: @SPEC:LANG-FIX-001
- **TEST**: tests/language/test_localization.py, tests/integration/test_command_language_flow.py
- **CODE**: src/moai_adk/core/project/phase_executor.py, src/moai_adk/core/template/processor.py
- **DOC**: CLAUDE.md (Language Architecture section), .moai/memory/language-config-schema.md

---

## Detailed Requirements Breakdown

### Python Layer (3 Requirements)

#### REQ-PY-001: Fix Configuration Path Reading (CRITICAL)
```python
# BROKEN (Current)
config.get("conversation_language", config.get("locale", "en"))

# FIXED (Required)
language_config = config.get("language", {})
language_config.get("conversation_language", "en")
```

**File**: `src/moai_adk/core/project/phase_executor.py:154`
**Impact**: CRITICAL - Blocks all language flow
**Test**: Unit test must verify nested config path reading

#### REQ-PY-002: Add Language Name Variable
```python
context = {
    "CONVERSATION_LANGUAGE": "ko",
    "CONVERSATION_LANGUAGE_NAME": "한국어"  # NEW
}
```

**File**: `src/moai_adk/core/template/processor.py`
**Impact**: HIGH - Enables localized command descriptions
**Test**: Template substitution verification

#### REQ-PY-003: Verify Template Substitution
**Requirement**: `{{CONVERSATION_LANGUAGE}}` must substitute correctly in all templates
**Test**: Integration test with actual template files

---

### Command Layer (4 Requirements)

#### REQ-CMD-001: Add Language Parameter to All Task() Calls
**Affected Files** (4 commands + 4 templates):
- `.claude/commands/alfred/0-project.md:482` → project-manager invocation
- `.claude/commands/alfred/1-plan.md:143, 190` → spec-builder invocations (2 places)
- `.claude/commands/alfred/2-run.md` → tdd-implementer invocation
- `.claude/commands/alfred/3-sync.md` → doc-syncer invocation
- `src/moai_adk/templates/.claude/commands/alfred/{0,1,2,3}-*.md` (same lines)

**Format**:
```markdown
Task(
    subagent_type="spec-builder",
    prompt="""
    LANGUAGE CONFIGURATION:
    - conversation_language: {{CONVERSATION_LANGUAGE}}
    - language_name: {{CONVERSATION_LANGUAGE_NAME}}

    TASK: [original prompt]

    OUTPUT REQUIREMENT:
    All SPEC documents must be written in conversation_language.
    """
)
```

**Test**: Command invocation test verifying parameter presence

#### REQ-CMD-002: Add Language Loading to Command Headers
**Requirement**: Each command header must explicitly document language support
**Files**: Same 4 command files + 4 templates

---

### Agent Layer (2 Requirements)

#### REQ-AGENT-001: Add Language Handling Section to 7 Missing Agents (CRITICAL)
**Agents missing language instructions**:
1. debug-helper.md
2. quality-gate.md
3. tag-agent.md
4. trust-checker.md
5. git-manager.md
6. cc-manager.md
7. skill-factory.md

**Standard Section to Add**:
```markdown
## 🌍 Language Handling

**IMPORTANT**: You will ALWAYS receive prompts in **English**, regardless of user's original conversation language.

**Do not try to infer language from your prompt.** Always process in English.

**Output Rule**: Check `conversation_language` parameter:
- If `conversation_language="ko"` → Output in Korean
- If `conversation_language="ja"` → Output in Japanese
- Otherwise → Output in English

**Mixed Language**: ALWAYS English for code, @TAGs, technical terms.
```

**Files**:
- `.claude/agents/alfred/{agent-name}.md` (7 files)
- `src/moai_adk/templates/.claude/agents/alfred/{agent-name}.md` (7 template files)

**Test**: Agent capability test verifying language section presence

#### REQ-AGENT-002: Standardize 5 Existing Agents
**Agents with existing sections** (need standardization):
1. project-manager.md
2. spec-builder.md
3. implementation-planner.md
4. tdd-implementer.md
5. doc-syncer.md

**Action**: Verify consistency, update if format differs

**Files**: Same as above (5 active + 5 template)

---

### Template Layer (2 Requirements)

#### REQ-TEMPLATE-001: Mirror All Command Changes
**Requirement**: All changes to `.claude/commands/alfred/*.md` must be mirrored to `src/moai_adk/templates/.claude/commands/alfred/*.md`

**Files**: 4 template files (one for each command)

**Test**: File comparison test ensuring templates match active files

#### REQ-TEMPLATE-002: Mirror All Agent Changes
**Requirement**: All changes to `.claude/agents/alfred/*.md` must be mirrored to `src/moai_adk/templates/.claude/agents/alfred/*.md`

**Files**: 12 template files (one for each agent)

**Test**: File comparison test for all agent templates

---

### Configuration Layer (2 Requirements)

#### REQ-CONFIG-001: Update config.json Template with Nested Language
**File**: `src/moai_adk/templates/config.json.template`

**Before**:
```json
{
  "conversation_language": "en",
  "locale": "en"
}
```

**After**:
```json
{
  "language": {
    "conversation_language": "en",
    "conversation_language_name": "English"
  },
  "project": {
    ...
  }
}
```

#### REQ-CONFIG-002: Add Config Migration Logic
**Requirement**: Support legacy flat config during transition
**Location**: `src/moai_adk/core/config/migration.py` (new or existing)

---

### Documentation Layer (2 Requirements)

#### REQ-DOC-001: Update CLAUDE.md Language Architecture
**File**: `CLAUDE.md` and `src/moai_adk/templates/CLAUDE.md`

**Changes**:
- Verify language architecture matches implementation
- Add "Known Limitations" section noting current status
- Add implementation roadmap

#### REQ-DOC-002: Create Language Configuration Schema
**File**: `.moai/memory/language-config-schema.md` (new)

**Content**:
- Complete config.json language section schema
- Supported languages and language codes
- Default behavior
- Migration guide from legacy config

---

## Implementation Constraints

### Order Dependencies
1. **Python fixes MUST come first** (Phase 1) - blocks all others
2. **Config updates MUST come second** (Phase 2) - depends on Python fixes
3. **Agent updates CAN be parallel** (Phase 3) - independent of each other
4. **Command updates MUST come after agent updates** (Phase 4) - agents must be ready
5. **Template updates MUST come last** (Phase 5) - after all active files finalized

### Testing Requirements
- Unit tests for Python config reading
- Integration tests for command → agent flow
- E2E tests for complete workflow in 3+ languages
- Template comparison tests to prevent drift

### Success Criteria
- [ ] Korean user workflow: new project → SPEC → English output
- [ ] Japanese user workflow: works in Japanese
- [ ] Spanish user workflow: works in Spanish
- [ ] All config reading tests pass
- [ ] All integration tests pass
- [ ] E2E tests for 3 languages pass
- [ ] Documentation updated and accurate

---

## Implementation Acceptance

This SPEC is ready for implementation when:
1. All requirements are understood and mapped to files
2. Test strategy is approved
3. Implementation order is confirmed
4. Team agrees to 6-7 hour timeline

**Status**: Ready for `/alfred:2-run SPEC-LANG-FIX-001` execution
