# Acceptance Criteria: @SPEC:LANG-FIX-001

## Overview

This document defines the acceptance criteria and test scenarios for the language localization system fix. All criteria must be met before the SPEC can transition from draft to completed status.

---

## Acceptance Criteria by Requirement

### Python Layer Acceptance

#### AC-PY-001: Configuration Path Reading Works

**Given**: User has `.moai/config.json` with nested language structure
```json
{
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어"
  }
}
```

**When**: `phase_executor.py` reads configuration

**Then**:
- [ ] `CONVERSATION_LANGUAGE` variable equals "ko"
- [ ] `CONVERSATION_LANGUAGE_NAME` variable equals "한국어"
- [ ] No errors or warnings logged
- [ ] Template substitution works: `{{CONVERSATION_LANGUAGE}}` → "ko"

**Test Case**: `test_phase_executor_reads_nested_language_config()`

---

#### AC-PY-002: Legacy Config Migration Works

**Given**: User has legacy flat config:
```json
{
  "conversation_language": "ja",
  "locale": "ja"
}
```

**When**: Migration logic runs during config initialization

**Then**:
- [ ] Config is converted to nested structure
- [ ] `CONVERSATION_LANGUAGE` equals "ja"
- [ ] Original value is preserved
- [ ] No data loss
- [ ] Subsequent reads work correctly

**Test Case**: `test_config_migration_preserves_language()`

---

#### AC-PY-003: Default Language is English

**Given**: User config has no language section

**When**: `phase_executor.py` reads configuration

**Then**:
- [ ] `CONVERSATION_LANGUAGE` defaults to "en"
- [ ] `CONVERSATION_LANGUAGE_NAME` defaults to "English"
- [ ] No errors, graceful fallback
- [ ] System continues to function

**Test Case**: `test_default_language_is_english()`

---

### Command Layer Acceptance

#### AC-CMD-001: 0-project Command Passes Language to project-manager

**Given**: User runs `/alfred:0-project` with Korean language set
```json
{"language": {"conversation_language": "ko"}}
```

**When**: Command invokes project-manager agent

**Then**:
- [ ] Task() call includes `conversation_language: "ko"` in prompt
- [ ] Task() call includes `language_name: "한국어"` in prompt
- [ ] Prompt explicitly states "generate output in conversation_language"
- [ ] Agent receives both parameters

**Test Case**: `test_0_project_passes_language_to_agent()`
**Validation**: Grep command files for `conversation_language` parameter in Task()

---

#### AC-CMD-002: 1-plan Command Passes Language to spec-builder

**Given**: User runs `/alfred:1-plan "Feature Name"` with Korean set

**When**: Command invokes spec-builder agent (both STEP 1 and STEP 2)

**Then**:
- [ ] Both Task() calls (line ~143 and ~190) include language config
- [ ] Prompt states "All SPEC documents must be written in conversation_language"
- [ ] Each file location (spec.md, plan.md, acceptance.md) mentions language requirement

**Test Cases**:
- `test_1_plan_step1_passes_language()`
- `test_1_plan_step2_passes_language()`

---

#### AC-CMD-003: 2-run Command Passes Language to tdd-implementer

**Given**: User runs `/alfred:2-run SPEC-XXX` with Spanish set

**When**: Command invokes tdd-implementer agent

**Then**:
- [ ] Task() call includes language config
- [ ] Prompt acknowledges user's language selection

**Test Case**: `test_2_run_passes_language_to_agent()`

---

#### AC-CMD-004: 3-sync Command Passes Language to doc-syncer

**Given**: User runs `/alfred:3-sync` with Japanese set

**When**: Command invokes doc-syncer agent

**Then**:
- [ ] Task() call includes language config
- [ ] Documentation generation respects language setting

**Test Case**: `test_3_sync_passes_language_to_agent()`

---

### Agent Layer Acceptance

#### AC-AGENT-001: All 12 Agents Have Language Section

**Given**: All agent files need review

**When**: Agent files are examined

**Then**:
- [ ] All 12 agents have "🌍 Language Handling" section
- [ ] Section is identical across all agents
- [ ] Section appears after agent persona, before other sections
- [ ] Section includes output language rule with examples

**Test Cases**:
- `test_all_agents_have_language_section()`
- `test_language_sections_are_consistent()`

**Agents to Check**:
1. ✓ project-manager.md
2. ✓ spec-builder.md
3. ✓ implementation-planner.md
4. ✓ tdd-implementer.md
5. ✓ doc-syncer.md
6. ✓ debug-helper.md
7. ✓ quality-gate.md
8. ✓ tag-agent.md
9. ✓ trust-checker.md
10. ✓ git-manager.md
11. ✓ cc-manager.md
12. ✓ skill-factory.md

---

#### AC-AGENT-002: Language Section States Output Rules

**Given**: Agent receives prompt with `conversation_language="ko"`

**When**: Agent reads language handling section

**Then**:
- [ ] Agent sees rule: "If conversation_language='ko' → Generate output in Korean"
- [ ] Agent sees example showing Korean text in output
- [ ] Agent sees rule: "Always use English for code and @TAGs"
- [ ] Agent has clear, unambiguous instructions

**Test Case**: `test_language_section_contains_output_rules()`

---

### Template Layer Acceptance

#### AC-TEMPLATE-001: Command Templates Match Active Files

**Given**: All 4 command files modified with language parameters

**When**: Templates in `src/moai_adk/templates/.claude/commands/alfred/` examined

**Then**:
- [ ] Template 0-project.md matches active 0-project.md
- [ ] Template 1-plan.md matches active 1-plan.md
- [ ] Template 2-run.md matches active 2-run.md
- [ ] Template 3-sync.md matches active 3-sync.md
- [ ] Difference is only in variable placeholders ({{PROJECT_NAME}}, etc.)

**Test Case**: `test_command_templates_match_active_files()`

**Validation**: `diff` comparison between active and template files

---

#### AC-TEMPLATE-002: Agent Templates Match Active Files

**Given**: All 12 agent files modified with language sections

**When**: Templates in `src/moai_adk/templates/.claude/agents/alfred/` examined

**Then**:
- [ ] All 12 agent templates match their active counterparts
- [ ] Language sections are identical
- [ ] No unintended modifications in templates

**Test Case**: `test_agent_templates_match_active_files()`

**Validation**: `diff` comparison for all 12 agents

---

### Configuration Layer Acceptance

#### AC-CONFIG-001: config.json Template Has Nested Structure

**Given**: New project initialization

**When**: `config.json` is generated from template

**Then**:
- [ ] Config includes nested `language` section
- [ ] `language.conversation_language` exists (e.g., "ko")
- [ ] `language.conversation_language_name` exists (e.g., "한국어")
- [ ] Other config sections unchanged

**Test Case**: `test_config_template_has_nested_language_section()`

---

#### AC-CONFIG-002: Legacy Config Migrates Correctly

**Given**: Existing project with flat config:
```json
{"conversation_language": "ko", "locale": "ko"}
```

**When**: User initializes project with updated MoAI-ADK

**Then**:
- [ ] Config is automatically migrated to nested structure
- [ ] No manual intervention required
- [ ] Language setting preserved exactly
- [ ] All other config values preserved

**Test Case**: `test_legacy_config_auto_migrates()`

---

### Integration & E2E Acceptance

#### AC-INT-001: Korean User Complete Workflow

**Given**: New project, user selects Korean (ko)

**When**: User executes full workflow:
```bash
/alfred:0-project
/alfred:1-plan "사용자 인증 시스템"
/alfred:2-run SPEC-AUTH-001
```

**Then**:
- [ ] **After 0-project**:
  - config.json: `conversation_language: "ko"`
  - product.md, structure.md, tech.md contain Korean text

- [ ] **After 1-plan**:
  - `.moai/specs/SPEC-AUTH-001/spec.md` generated
  - spec.md title in Korean: "# @SPEC:AUTH-001: JWT 기반 인증..."
  - Requirements section in Korean: "## 요구사항"
  - Requirement text in Korean: "시스템은 사용자 인증 기능을 제공해야 한다"

- [ ] **After 2-run**:
  - Code generated in English (internal requirement)
  - Code comments may be in Korean (per conversation_language)
  - Implementation completes without errors

**Test Case**: `test_korean_user_complete_workflow()`

**Validation Steps**:
```bash
# Check config
cat .moai/config.json | grep "conversation_language" # Should show "ko"

# Check SPEC content
cat .moai/specs/SPEC-AUTH-001/spec.md | head -20 # Should have Korean text

# Check no hardcoded English in spec
grep -i "requirement\|ubiquitous\|event-driven" .moai/specs/SPEC-AUTH-001/spec.md # Should be in Korean
```

---

#### AC-INT-002: Japanese User Complete Workflow

**Given**: New project, user selects Japanese (ja)

**When**: User executes:
```bash
/alfred:0-project
/alfred:1-plan "ユーザー認証"
```

**Then**:
- [ ] product.md in Japanese
- [ ] SPEC document in Japanese
- [ ] Requirements in Japanese: "システムはユーザー認証機能を提供する必要があります"
- [ ] All user-facing output in Japanese

**Test Case**: `test_japanese_user_complete_workflow()`

---

#### AC-INT-003: Spanish User Complete Workflow

**Given**: New project, user selects Spanish (es)

**When**: User executes:
```bash
/alfred:0-project
/alfred:1-plan "Sistema de autenticación"
```

**Then**:
- [ ] All documents in Spanish
- [ ] SPEC requirements in Spanish
- [ ] No English text in user-facing content

**Test Case**: `test_spanish_user_complete_workflow()`

---

#### AC-INT-004: English User Workflow (Regression Test)

**Given**: User selects English (en) - existing behavior

**When**: User executes full workflow

**Then**:
- [ ] All output in English (no change from before)
- [ ] Backward compatibility maintained
- [ ] No regressions in existing functionality

**Test Case**: `test_english_user_complete_workflow()`

---

### Documentation Layer Acceptance

#### AC-DOC-001: CLAUDE.md Updated

**Given**: Implementation complete

**When**: CLAUDE.md reviewed

**Then**:
- [ ] "🌍 Alfred's Language Boundary Rule" section accurate
- [ ] Example shows real implementation (not aspirational)
- [ ] Known Limitations section added
- [ ] Config schema documentation referenced
- [ ] No outdated information

**Test Case**: Manual review
**Validation**: Compare with implementation

---

#### AC-DOC-002: Language Config Schema Documented

**Given**: `.moai/memory/language-config-schema.md` created

**When**: File is reviewed

**Then**:
- [ ] Complete config structure shown
- [ ] All supported languages listed
- [ ] Default values documented
- [ ] Migration instructions included
- [ ] Examples provided for each language

**Test Case**: Manual review

---

## Quality Gates

### Code Quality
- [ ] All Python code passes linting (ruff, mypy)
- [ ] Test coverage ≥ 85% for language-related code
- [ ] No hardcoded English in localization-critical paths
- [ ] No TODOs or FIXMEs left in code

### Testing
- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] All E2E tests pass for 3+ languages
- [ ] No flaky tests
- [ ] Test execution time < 2 minutes

### Documentation
- [ ] All requirements traced to test cases
- [ ] All test cases documented in acceptance.md
- [ ] No broken links in documentation
- [ ] Examples runnable and verified

### Compliance
- [ ] All 40 files modified as planned
- [ ] No unintended modifications
- [ ] Template files match active files
- [ ] Git history clean (no intermediate commits left behind)

---

## Acceptance Sign-Off Checklist

**Implementation Complete When All Below Are ✓**:

- [ ] **Phase 1**: Python fixes + unit tests pass
- [ ] **Phase 2**: Config template updated + migration logic works
- [ ] **Phase 3**: All 12 agents have language sections
- [ ] **Phase 4**: All 4 commands pass language parameters
- [ ] **Phase 5**: Integration and E2E tests pass
- [ ] **Phase 6**: Documentation updated + validation checklist complete
- [ ] **Quality**: Code coverage ≥ 85%, all tests green
- [ ] **Validation**: All 40 files modified correctly
- [ ] **Sign-off**: @GOOS or authorized reviewer approves

---

## Regression Prevention

**Critical Workflows to Verify No Breaking Changes**:

### Personal Mode (Existing Behavior)
- [ ] Project initialization still works
- [ ] SPEC creation still works
- [ ] TDD implementation still works
- [ ] Document sync still works

### Team Mode (Existing Behavior)
- [ ] GitHub Issues still created
- [ ] Draft PRs still created
- [ ] Branch naming conventions preserved
- [ ] GitFlow still works

### Multi-language Mode (New Behavior)
- [ ] Korean documents generated in Korean
- [ ] Japanese documents generated in Japanese
- [ ] Spanish documents generated in Spanish
- [ ] English mode unchanged

---

## Success Criteria Summary

**Functional Success**:
```
Korean User: Project → SPEC → Code
  ✓ product.md in Korean
  ✓ SPEC document in Korean
  ✓ Code in English (internal)

Japanese User: Project → SPEC → Code
  ✓ product.md in Japanese
  ✓ SPEC document in Japanese
  ✓ Code in English (internal)

Spanish User: Project → SPEC → Code
  ✓ product.md in Spanish
  ✓ SPEC document in Spanish
  ✓ Code in English (internal)
```

**Quality Success**:
- Test coverage ≥ 85%
- All 40 files modified
- Documentation accurate
- No regressions

**Timeline Success**:
- Completion within 6-7 hours
- All phases executed in order
- No blockers or major issues

---

**Acceptance Status**: Ready for implementation and testing
**Sign-off Target**: @GOOS
**Review Cycle**: After all tests green
