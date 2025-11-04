# Runtime Translation Flow: CompanyAnnouncements

**Document**: Runtime Translation Architecture for MoAI-ADK Language Support
**Status**: Complete
**Last Updated**: 2025-11-04

---

## Overview

MoAI-ADK uses a **single-source-of-truth English approach** with **runtime translation** to support any user-selected language. This document explains the complete flow from STEP 0 (language selection) through to Claude Code displaying announcements in the user's language.

---

## Architecture Design Principles

| Principle | Rationale |
|-----------|-----------|
| **Single Source of Truth** | Only English items stored in config.json; eliminates duplication and maintenance burden |
| **Runtime Translation** | Translations happen at runtime after user selects conversation_language, not at build/install time |
| **Zero Pre-translation** | No pre-translated copies maintained; supports unlimited languages automatically |
| **Any Language Support** | Not limited to predefined language list (en, ko, ja, zh, es, fr, etc.); ANY language supported |
| **Future-Proof** | New languages automatically supported via translation service without code changes |

---

## Complete Data Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│ USER INITIATES /alfred:0-project COMMAND                               │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ STEP 0: PROJECT INITIALIZATION (project-manager agent)                 │
│                                                                           │
│ Questions:                                                               │
│ 1. Project name, description, owner                                      │
│ 2. Technology stack and language                                         │
│ 3. ** LANGUAGE SELECTION **                                              │
│    - conversation_language = "ko" (Korean)                              │
│    - conversation_language_name = "한국어"                              │
│                                                                           │
│ Saved to: .moai/config.json → language.conversation_language            │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ VARIABLE MAPPING PHASE                                                   │
│                                                                           │
│ Template Variables Created:                                              │
│ - {{CONVERSATION_LANGUAGE}} = "ko"                                      │
│ - {{CONVERSATION_LANGUAGE_NAME}} = "한국어"                             │
│                                                                           │
│ Used in:                                                                 │
│ - 0-project.md STEP 2.1.2: Agent prompt language setting                │
│ - config.json: announcements.language field                             │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ BASE ENGLISH ANNOUNCEMENTS (from config.json)                           │
│                                                                           │
│ Source of Truth - Single English Version:                               │
│ [                                                                         │
│   "🎩 SPEC-First: Always define requirements as SPEC...",              │
│   "✅ TRUST 5 Principles: Test First, Readable, ...",                  │
│   "📝 TodoWrite Usage: Track all tasks and update...",                 │
│   "🌍 Language Boundary: Use conversation_language...",                │
│   "🔗 @TAG Chain: Maintain traceability SPEC→TEST→CODE→DOC",          │
│   "⚡ Parallel Execution: Independent tasks...",                       │
│   "💡 Skills First: Check appropriate Skill..."                        │
│ ]                                                                         │
│                                                                           │
│ Stored in: src/moai_adk/templates/.moai/config.json                    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ TRANSLATION PHASE (Alfred Translation Service)                          │
│                                                                           │
│ For each English announcement item:                                      │
│ INPUT:  "🎩 SPEC-First: Always define requirements as SPEC..."         │
│ SERVICE: translate(english_text → korean)                              │
│ OUTPUT: "🎩 SPEC-First: 구현 전에 항상 요구사항을..."                  │
│                                                                           │
│ Process:                                                                 │
│ 1. Read conversation_language from config.json                          │
│ 2. Create translation request with ALL 7 English items                  │
│ 3. Send to translation service (external API or local)                  │
│ 4. Receive translated array in target language                          │
│ 5. Validate translation quality (fallback to English if needed)         │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ SAVE TO .claude/settings.json                                           │
│                                                                           │
│ Generated File Content:                                                  │
│ {                                                                         │
│   "companyAnnouncements": [                                              │
│     "🎩 SPEC-First: 구현 전에 항상 요구사항을...",                     │
│     "✅ TRUST 5 Principles: Test First, Readable, ...",                │
│     "📝 TodoWrite Usage: 모든 작업을 추적하고...",                     │
│     "🌍 Language Boundary: conversation_language를...",                │
│     "🔗 @TAG Chain: SPEC→TEST→CODE→DOC 추적성을...",                 │
│     "⚡ Parallel Execution: 독립적인 작업은...",                       │
│     "💡 Skills First: 도메인 특화 작업은..."                          │
│   ]                                                                       │
│ }                                                                         │
│                                                                           │
│ Saved to: .claude/settings.json (local runtime configuration)           │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ CLAUDE CODE SESSION STARTUP                                              │
│                                                                           │
│ Claude Code initialization:                                              │
│ 1. Read .claude/settings.json                                            │
│ 2. Extract companyAnnouncements array                                    │
│ 3. Randomly select one item from the array                              │
│ 4. Display in Claude Code UI at startup                                 │
│                                                                           │
│ User sees (in their selected language):                                 │
│ ┌─────────────────────────────────────────────────────────────┐         │
│ │ 💡 Skills First: 도메인 특화 작업은 먼저 적절한 Skill을 확인 │         │
│ └─────────────────────────────────────────────────────────────┘         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## File Involvement Summary

| File | Role | Content | Timing |
|------|------|---------|--------|
| `.moai/config.json` | Stores selected language | `language.conversation_language = "ko"` | After STEP 0 |
| `src/moai_adk/templates/.moai/config.json` | Template source | English base announcements items | Package default |
| `src/moai_adk/templates/.claude/settings.json` | Template source | English base companyAnnouncements | Package default |
| `.claude/settings.json` | Runtime config | Translated companyAnnouncements | After translation |
| `.claude/commands/alfred/0-project.md` | Documentation | STEP 2.1.4 explains flow | Reference |

---

## Implementation Details

### Step 1: Language Selection (STEP 0)

**User Input**:
- Selects `conversation_language` from list (or enters custom)
- Examples: "ko" (Korean), "ja" (Japanese), "zh" (Chinese Simplified), etc.

**Result**:
```json
{
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어"
  }
}
```

### Step 2: Read Configuration

**Code Logic** (Alfred internally):
```python
import json
from pathlib import Path

config = json.loads(Path(".moai/config.json").read_text())
conversation_language = config["language"]["conversation_language"]
conversation_language_name = config["language"]["conversation_language_name"]
```

### Step 3: Get Base Announcements

**Source**: `config.json` → `announcements.items` array

**English base items** (7 total):
```json
{
  "announcements": {
    "enabled": true,
    "language": "{{CONVERSATION_LANGUAGE}}",
    "items": [
      "🎩 SPEC-First: Always define requirements as SPEC before implementation (/alfred:1-plan)",
      "✅ TRUST 5 Principles: Test First, Readable, Unified, Secured, Trackable",
      "📝 TodoWrite Usage: Track all tasks and update in_progress/completed status immediately",
      "🌍 Language Boundary: Use conversation_language for dialogs/documents, English for infrastructure",
      "🔗 @TAG Chain: Maintain traceability SPEC→TEST→CODE→DOC",
      "⚡ Parallel Execution: Independent tasks can run simultaneously (Task tool parallel calls)",
      "💡 Skills First: Check appropriate Skill first for domain-specific tasks"
    ]
  }
}
```

### Step 4: Translate Each Item

**Implementation**:
```python
# Pseudo-code - actual implementation depends on translation service

def translate_announcements(items, target_language):
    """
    Translate all announcement items to target language
    """
    translated_items = []

    for item in items:
        # Call translation service
        translated_item = translation_service.translate(
            text=item,
            source_language="en",
            target_language=target_language
        )
        translated_items.append(translated_item)

    return translated_items

# Usage
korean_announcements = translate_announcements(
    items=base_english_announcements,
    target_language="ko"
)
```

### Step 5: Save to .claude/settings.json

**Process**:
```python
settings = {
    "companyAnnouncements": translated_items
}

# Save to local settings
Path(".claude/settings.json").write_text(
    json.dumps(settings, ensure_ascii=False, indent=2)
)
```

**Result**:
```json
{
  "companyAnnouncements": [
    "🎩 SPEC-First: 구현 전에 항상 요구사항을 SPEC으로 정의하세요 (/alfred:1-plan)",
    "✅ TRUST 5 Principles: Test First, Readable, Unified, Secured, Trackable",
    "📝 TodoWrite Usage: 모든 작업을 추적하고 in_progress/completed 상태를 즉시 업데이트하세요",
    "🌍 Language Boundary: conversation_language를 대화/문서에, 영어를 인프라에 사용하세요",
    "🔗 @TAG Chain: SPEC→TEST→CODE→DOC 추적성을 유지하세요",
    "⚡ Parallel Execution: 독립적인 작업은 동시에 실행할 수 있습니다 (Task tool 병렬 호출)",
    "💡 Skills First: 도메인 특화 작업은 먼저 적절한 Skill을 확인하세요"
  ]
}
```

### Step 6: Claude Code Display

**When Claude Code starts**:
1. Loads `.claude/settings.json`
2. Reads `companyAnnouncements` array
3. Randomly selects one item
4. Displays to user in their selected language

---

## Example: Complete Korean Translation

### Input (STEP 0)

User selects: **한국어 (Korean)** with code **ko**

```json
{
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어"
  }
}
```

### Processing

Base English item:
```
🎩 SPEC-First: Always define requirements as SPEC before implementation (/alfred:1-plan)
```

Translation request:
```
translate(
  text="🎩 SPEC-First: Always define requirements as SPEC before implementation (/alfred:1-plan)",
  source="en",
  target="ko"
)
```

### Output

Translated to Korean:
```
🎩 SPEC-First: 구현 전에 항상 요구사항을 SPEC으로 정의하세요 (/alfred:1-plan)
```

### Display

Claude Code startup shows:
```
🎩 SPEC-First: 구현 전에 항상 요구사항을 SPEC으로 정의하세요 (/alfred:1-plan)
```

---

## Template Variable Substitution

### Variables Used

| Variable | Source | Value | Used In |
|----------|--------|-------|---------|
| `{{CONVERSATION_LANGUAGE}}` | STEP 0 selection | "ko" | config.json `announcements.language` |
| `{{CONVERSATION_LANGUAGE_NAME}}` | STEP 0 selection | "한국어" | .moai/config.json display |
| `{{AGENT_PROMPT_LANGUAGE}}` | STEP 2.1.2 selection | "english" or "localized" | 0-project.md prompt |

### Substitution Points

**In config.json**:
```json
{
  "announcements": {
    "language": "{{CONVERSATION_LANGUAGE}}"
  }
}
```

**In 0-project.md STEP 2.1.2**:
```
The `agent_prompt_language` is set to: {{AGENT_PROMPT_LANGUAGE}}
```

---

## Supported Languages

Not limited to this list (any language supported by translation service):

| Code | Language | Example Announcement |
|------|----------|---------------------|
| en | English | 🎩 SPEC-First: Always define requirements as SPEC... |
| ko | Korean | 🎩 SPEC-First: 구현 전에 항상 요구사항을 SPEC으로... |
| ja | Japanese | 🎩 SPEC-First: 常に要件をSPECとして定義してください... |
| zh | Chinese (Simplified) | 🎩 SPEC-First: 始终将需求定义为SPEC... |
| es | Spanish | 🎩 SPEC-First: Siempre define los requisitos como SPEC... |
| fr | French | 🎩 SPEC-First: Définissez toujours les exigences en tant que SPEC... |
| de | German | 🎩 SPEC-First: Definieren Sie Anforderungen immer als SPEC... |
| pt | Portuguese | 🎩 SPEC-First: Sempre defina requisitos como SPEC... |
| ru | Russian | 🎩 SPEC-First: Всегда определяйте требования как SPEC... |
| ar | Arabic | 🎩 SPEC-First: حدد دائماً المتطلبات كـ SPEC... |
| hi | Hindi | 🎩 SPEC-First: हमेशा आवश्यकताओं को SPEC के रूप में परिभाषित करें... |

---

## Key Design Decisions

### Why English Only in Source Files?

✅ **Advantages**:
- Single source of truth (no duplication)
- Standard practice for technical documentation
- Easy maintenance and consistency
- Global compatibility

❌ **Avoid**:
- Pre-translated versions in multiple languages
- Maintenance burden for each language variant
- Risk of inconsistency between versions
- Limited to pre-defined languages

### Why Runtime Translation?

✅ **Advantages**:
- Supports ANY language (unlimited scalability)
- No code changes needed for new languages
- Fresh translations always available
- Translation service can be updated independently

❌ **Avoid**:
- Hard-coded translations in source
- Build-time translation dependencies
- Limited to supported languages at compile time
- Stale translations

### Single vs. Multiple Translation Services

**Current Design**:
- Abstracted translation service
- Can use any provider (OpenAI, Google Translate, local model, etc.)
- Fallback to English if translation fails

**Benefits**:
- Flexibility to switch providers
- Cost optimization possible
- Resilience through fallbacks
- Framework-agnostic

---

## Error Handling & Fallbacks

### Translation Failure

If translation service is unavailable:
```python
try:
    translated = translate_service.translate(item, target_lang)
except TranslationError:
    # Fallback to English
    translated = item
    logger.warning(f"Translation failed for {target_lang}, using English")
```

### Invalid Language Code

If user selects unsupported language:
```python
if conversation_language not in SUPPORTED_LANGUAGES:
    conversation_language = "en"  # Fallback to English
```

### Missing Announcements

If config.json doesn't have announcements array:
```python
announcements = config.get("announcements", {}).get("items", [])
if not announcements:
    # Use hardcoded defaults
    announcements = [DEFAULT_SPEC_FIRST, ...]
```

---

## Integration with Alfred Workflow

### Phase 0: Project Initialization

- User selects `conversation_language` in STEP 0
- Alfred saves to `.moai/config.json`
- Announcements translation triggered automatically

### Phase 1: Specification

- STEP 2.1.2: Agent prompt language determined
- STEP 2.1.4: CompanyAnnouncements translated
- Documentation generated in user's language

### Phase 2: Implementation

- Sub-agents receive language parameter
- Code and docs respect user's language selection

### Phase 3: Sync

- Documentation verified and synchronized
- Language settings maintained throughout workflow

---

## Testing & Validation

### Test Cases

| Test | Input | Expected Output |
|------|-------|-----------------|
| English selected | conversation_language="en" | English announcements in .claude/settings.json |
| Korean selected | conversation_language="ko" | Korean announcements in .claude/settings.json |
| Japanese selected | conversation_language="ja" | Japanese announcements in .claude/settings.json |
| Translation fails | service down, conversation_language="ko" | Fallback to English announcements |
| Missing config | No language in config.json | Default to English |
| All 7 items translated | conversation_language="es" | All 7 Spanish announcements present |
| Unicode characters | conversation_language="ru" | Russian Cyrillic characters preserved |
| Emoji preservation | All translations | All emoji preserved in translations |

### Validation Checklist

- ✅ Single English source exists (config.json `announcements.items`)
- ✅ All 7 items present in base announcements
- ✅ Translation triggered after STEP 0 language selection
- ✅ Translated items saved to `.claude/settings.json`
- ✅ Claude Code displays translated announcement on startup
- ✅ Fallback to English if translation fails
- ✅ Unicode and emoji preserved
- ✅ No hardcoded pre-translations in code

---

## Future Enhancements

### Possible Improvements

1. **Batch Translation**
   - Send all 7 items in single API call
   - Better performance and cost efficiency

2. **Translation Caching**
   - Cache translations to avoid redundant API calls
   - Store in `.moai/cache/translations.json`

3. **Custom Translation Service**
   - Allow users to configure their own translation API
   - Support multiple providers via plugins

4. **Announcement Updates**
   - Version control announcements
   - Support for time-based rotation of announcements

5. **Quality Scoring**
   - Validate translation quality
   - User feedback on announcement translations

---

## References

- **STEP 0 Language Selection**: `.claude/commands/alfred/0-project.md`
- **Translation Documentation**: `STEP 2.1.4` - Variable Mapping & CompanyAnnouncements Translation
- **Base Configuration**: `src/moai_adk/templates/.moai/config.json`
- **Runtime Settings**: `.claude/settings.json`
- **Support Skills**: `Skill("moai-alfred-language-detection")`

---

**Document Status**: ✅ Complete
**Last Reviewed**: 2025-11-04
**Maintainer**: MoAI-ADK Project
