# Language Configuration Schema

## Overview

This document defines the complete language configuration structure for MoAI-ADK. The configuration controls which language is used for generating project documentation, SPEC documents, and user-facing content.

---

## Config Structure

### New Nested Structure (v0.7.0+)

```json
{
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어"
  }
}
```

**Location**: Top-level key in `.moai/config.json`

**Fields**:
- `conversation_language` (string): Language code (e.g., "ko", "en", "ja")
- `conversation_language_name` (string): Display name in that language (e.g., "한국어", "English")

### Complete Example

```json
{
  "_meta": {
    "tag_format_example": "@TYPE:DOMAIN-NNN"
  },
  "moai": {
    "version": "0.7.0"
  },
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어"
  },
  "project": {
    "name": "My Project",
    "mode": "personal",
    "language": "python"
  },
  "tags": {
    "auto_sync": true
  }
}
```

---

## Supported Languages

| Code | Language | Display Name |
|------|----------|--------------|
| en | English | English |
| ko | Korean | 한국어 |
| ja | Japanese | 日本語 |
| zh | Chinese | 中文 |
| es | Spanish | Español |

---

## Default Behavior

### When Language Config is Missing

If `.moai/config.json` lacks the `language` section:

```python
language_config = config.get("language", {})  # Returns empty dict
conversation_language = language_config.get("conversation_language", "en")  # Defaults to "en"
conversation_language_name = language_config.get("conversation_language_name", "English")
```

**Result**: System defaults to **English**

### When Language Config is Invalid

If the `language` value is not a dictionary:

```python
language_config = config.get("language", {})
if not isinstance(language_config, dict):
    language_config = {}  # Reset to empty dict
```

**Result**: System falls back to **English** with safe handling

---

## Migration from Legacy Config

### Legacy Flat Structure (v0.6.3 and earlier)

```json
{
  "conversation_language": "ko",
  "locale": "ko"
}
```

### Migration Process

Automatic migration function available in `src/moai_adk/core/config/migration.py`:

```python
from moai_adk.core.config import migrate_config_to_nested_structure

# Existing config with legacy structure
old_config = {
    "conversation_language": "ko",
    "locale": "ko"
}

# Migrate to new structure
new_config = migrate_config_to_nested_structure(old_config)

# Result
# {
#     "language": {
#         "conversation_language": "ko",
#         "conversation_language_name": "한국어"
#     }
# }
```

### Migration Guarantees

- ✅ **No Data Loss**: Original language setting preserved exactly
- ✅ **Automatic Mapping**: Language code automatically mapped to language name
- ✅ **Backward Compatible**: Old flat structure still works (with fallback reading)
- ✅ **Non-Breaking**: Existing projects continue to function during transition

---

## Usage in Code

### Reading Configuration

#### Python Code (Recommended)

```python
from moai_adk.core.config import (
    get_conversation_language,
    get_conversation_language_name
)

config = load_config(".moai/config.json")
language_code = get_conversation_language(config)  # "ko"
language_name = get_conversation_language_name(config)  # "한국어"
```

#### Direct Access (When Migration Complete)

```python
config = load_config(".moai/config.json")
language_code = config.get("language", {}).get("conversation_language", "en")
language_name = config.get("language", {}).get("conversation_language_name", "English")
```

### Template Variable Substitution

When generating files from templates, two variables are available:

```
{{CONVERSATION_LANGUAGE}}        → "ko"
{{CONVERSATION_LANGUAGE_NAME}}   → "한국어"
```

These are set by `phase_executor.py` during project initialization:

```python
context = {
    "CONVERSATION_LANGUAGE": language_config.get("conversation_language", "en"),
    "CONVERSATION_LANGUAGE_NAME": language_config.get("conversation_language_name", "English"),
}
processor.set_context(context)
```

---

## Impact on Generated Content

### Documents Generated in User's Language

When a user sets `conversation_language: "ko"`:

- ✅ `product.md` - Generated in Korean
- ✅ `structure.md` - Generated in Korean
- ✅ `tech.md` - Generated in Korean
- ✅ SPEC documents (spec.md, plan.md, acceptance.md) - Generated in Korean
- ✅ Project documents - Generated in Korean

### Internal Content Always in English

Regardless of `conversation_language`:

- ✅ Code comments - Always English
- ✅ Git commit messages - Always English
- ✅ @TAG identifiers - Always English
- ✅ YAML frontmatter - Always English
- ✅ Technical keywords - Always English

---

## Sub-agent Communication

When invoking sub-agents via `Task()`, the language configuration is passed:

```python
context = {
    "CONVERSATION_LANGUAGE": "ko",
    "CONVERSATION_LANGUAGE_NAME": "한국어",
}

task_prompt = f"""
LANGUAGE CONFIGURATION:
- conversation_language: {context['CONVERSATION_LANGUAGE']}
- language_name: {context['CONVERSATION_LANGUAGE_NAME']}

TASK: [Actual task description]
"""
```

**Rules**:
1. **Sub-agent input**: In user's configured language (passed directly by Alfred)
2. **Sub-agent output**: Generated in user's language based on parameter
3. **No inference**: Sub-agents NEVER guess language from prompt content
4. **Explicit Skill invocation**: Skills invoked using `Skill("skill-name")` syntax, NOT auto-triggered

---

## Validation & Testing

### Config Validation

Validate nested structure:

```python
def validate_language_config(config: dict) -> bool:
    language_config = config.get("language", {})
    if not isinstance(language_config, dict):
        return False

    conversation_language = language_config.get("conversation_language")
    conversation_language_name = language_config.get("conversation_language_name")

    return (
        conversation_language in ["en", "ko", "ja", "zh", "es"]
        and isinstance(conversation_language_name, str)
        and len(conversation_language_name) > 0
    )
```

### Unit Tests

Test file: `tests/unit/test_language_config.py`

Covers:
- Nested config reading (Korean, Japanese, Spanish)
- Default values when missing
- Error handling for malformed config
- Template variable substitution
- Migration from legacy structure

---

## File References

**Config files**:
- `.moai/config.json` - Active project configuration
- `src/moai_adk/templates/.moai/config.json` - Template with placeholders

**Code files**:
- `src/moai_adk/core/project/phase_executor.py` - Reads config during initialization
- `src/moai_adk/core/config/migration.py` - Migration utilities
- `src/moai_adk/core/template/processor.py` - Template variable substitution

**Test files**:
- `tests/unit/test_language_config.py` - Unit tests
- `tests/integration/test_command_language_flow.py` - Integration tests

---

## Troubleshooting

### Issue: Documents Generated in English Despite Korean Setting

**Check**:
1. `.moai/config.json` has nested `language` section
2. `conversation_language` is set correctly (e.g., "ko")
3. `phase_executor.py` is setting context variables correctly
4. Sub-agents are receiving language parameters in `Task()` calls

### Issue: Migration Fails

**Solution**:
1. Manually create `language` section in config.json:
```json
{
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어"
  }
}
```
2. Test with `python -c "from moai_adk.core.config import migrate_config_to_nested_structure; ..."`

### Issue: Template Substitution Not Working

**Check**:
1. Template processor context includes both variables:
   - `CONVERSATION_LANGUAGE`
   - `CONVERSATION_LANGUAGE_NAME`
2. Template files have correct placeholders: `{{CONVERSATION_LANGUAGE}}`
3. No typos in placeholder names

---

## Future Enhancements

Potential improvements for future versions:

- [ ] Environment variable override: `MOAI_CONVERSATION_LANGUAGE=ja`
- [ ] System locale detection as fallback
- [ ] Language pack system for custom mappings
- [ ] Performance caching of language config

---

**Last Updated**: 2025-10-29
**Status**: Complete - SPEC-LANG-FIX-001 Phase 2
