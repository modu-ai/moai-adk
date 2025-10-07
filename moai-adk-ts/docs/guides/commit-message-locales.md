# Commit Message Localization Guide

## Overview

MoAI-ADK supports locale-based commit message generation for TDD workflows. Commit messages are automatically generated in the language specified in your `.moai/config.json` file.

## Supported Locales

- **ko** - Korean (한국어)
- **en** - English
- **ja** - Japanese (日本語)
- **zh** - Chinese (中文)

## Configuration

### Setting Your Locale

Edit `.moai/config.json` in your project root:

```json
{
  "project": {
    "name": "my-project",
    "mode": "team",
    "locale": "ko",
    ...
  }
}
```

### Locale Priority

The system uses the following priority order:

1. **Project Config**: `.moai/config.json` → `project.locale`
2. **Environment Variable**: `MOAI_LOCALE`
3. **Default**: `en` (English)

## TDD Commit Message Templates

### Korean (ko)

```bash
🔴 RED: 테스트 설명
🟢 GREEN: 구현 설명
♻️ REFACTOR: 개선 설명
📝 DOCS: 문서 설명
```

### English (en)

```bash
🔴 RED: Test description
🟢 GREEN: Implementation description
♻️ REFACTOR: Improvement description
📝 DOCS: Documentation description
```

### Japanese (ja)

```bash
🔴 RED: テスト説明
🟢 GREEN: 実装説明
♻️ REFACTOR: 改善説明
📝 DOCS: ドキュメント説明
```

### Chinese (zh)

```bash
🔴 RED: 测试说明
🟢 GREEN: 实现说明
♻️ REFACTOR: 改进说明
📝 DOCS: 文档说明
```

## Usage Examples

### Using git-manager Agent

The `git-manager` agent automatically reads your locale from `.moai/config.json`:

```bash
# Korean locale example
@agent-git-manager "SPEC-AUTH-001에 대한 TDD 커밋 생성"

# Generated commits:
# 🔴 RED: 로그인 테스트 추가
# @TEST:AUTH-001-RED
#
# 🟢 GREEN: 로그인 기능 구현
# @CODE:AUTH-001-GREEN
#
# ♻️ REFACTOR: 로그인 코드 정리
# REFACTOR:AUTH-001-CLEAN
```

### Programmatic Usage

```typescript
import { getTDDCommitMessage, getTDDCommitWithTag } from '@moai-adk/git';

// Simple message
const message = getTDDCommitMessage('ko', 'RED', '로그인 테스트 추가');
// Result: "🔴 RED: 로그인 테스트 추가"

// With @TAG
const messageWithTag = getTDDCommitWithTag(
  'ko',
  'RED',
  '로그인 테스트 추가',
  'AUTH-001'
);
// Result: "🔴 RED: 로그인 테스트 추가\n\n@TEST:AUTH-001-RED"
```

### GitCommitManager Integration

```typescript
import { GitCommitManager } from '@moai-adk/git';

const manager = new GitCommitManager(config, '/path/to/project');

// Locale is automatically loaded from .moai/config.json
console.log(manager.getLocale()); // "ko"

// Create commits (automatically uses locale)
await manager.commitChanges('테스트 추가');

// Update locale if config changes
manager.updateLocale();
```

## TDD Workflow with Locales

### Complete Example (Korean)

**1. SPEC 작성** (`/alfred:1-spec`)
```bash
# .moai/config.json에서 locale: "ko" 확인
```

**2. TDD 구현** (`/alfred:2-build`)
```bash
# RED Phase
git add tests/
git commit -m "🔴 RED: 로그인 실패 테스트

@TEST:LOGIN-001-RED"

# GREEN Phase
git add src/
git commit -m "🟢 GREEN: 로그인 기능 구현

@CODE:LOGIN-001-GREEN"

# REFACTOR Phase
git commit -am "♻️ REFACTOR: 로그인 코드 개선

REFACTOR:LOGIN-001-CLEAN"
```

**3. 문서 동기화** (`/alfred:3-sync`)
```bash
git commit -m "📝 DOCS: 로그인 API 문서 업데이트

@DOC:LOGIN-001"
```

### Complete Example (English)

**1. SPEC Creation** (`/alfred:1-spec`)
```bash
# Check locale: "en" in .moai/config.json
```

**2. TDD Implementation** (`/alfred:2-build`)
```bash
# RED Phase
git add tests/
git commit -m "🔴 RED: add login failure test

@TEST:LOGIN-001-RED"

# GREEN Phase
git add src/
git commit -m "🟢 GREEN: implement login feature

@CODE:LOGIN-001-GREEN"

# REFACTOR Phase
git commit -am "♻️ REFACTOR: improve login code

REFACTOR:LOGIN-001-CLEAN"
```

**3. Documentation Sync** (`/alfred:3-sync`)
```bash
git commit -m "📝 DOCS: update login API documentation

@DOC:LOGIN-001"
```

## API Reference

### Functions

#### `getTDDCommitMessage(locale, stage, message)`

Generate a TDD commit message without @TAG.

**Parameters:**
- `locale: CommitLocale` - Target locale ('ko' | 'en' | 'ja' | 'zh')
- `stage: TDDStage` - TDD stage ('RED' | 'GREEN' | 'REFACTOR' | 'DOCS')
- `message: string` - Commit message content

**Returns:** `string` - Formatted commit message

#### `getTDDCommitWithTag(locale, stage, message, specId)`

Generate a TDD commit message with @TAG.

**Parameters:**
- `locale: CommitLocale` - Target locale
- `stage: TDDStage` - TDD stage
- `message: string` - Commit message content
- `specId: string` - SPEC ID (e.g., 'AUTH-001')

**Returns:** `string` - Formatted commit message with @TAG

#### `loadLocaleFromConfig(workingDir)`

Load locale from `.moai/config.json`.

**Parameters:**
- `workingDir: string` - Project directory path

**Returns:** `CommitLocale` - Validated locale (defaults to 'en')

#### `getLocaleWithFallback(workingDir)`

Get locale with complete fallback chain.

**Parameters:**
- `workingDir: string` - Project directory path

**Returns:** `CommitLocale` - Locale from config → env → default

## Best Practices

### 1. Set Locale During Project Initialization

```bash
moai-adk init
# Select your preferred locale during setup
```

### 2. Keep Locale Consistent Across Team

For team projects, document the locale choice in your README:

```markdown
## Development Setup

This project uses **Korean** commit messages.
Locale is configured in `.moai/config.json`:

\`\`\`json
{
  "project": {
    "locale": "ko"
  }
}
\`\`\`
```

### 3. Use Environment Variable for Personal Preference

Individual developers can override the project locale:

```bash
export MOAI_LOCALE=en
```

### 4. Validate Locale in CI/CD

Add a check in your CI/CD pipeline:

```yaml
- name: Validate locale
  run: |
    LOCALE=$(jq -r '.project.locale' .moai/config.json)
    if [ "$LOCALE" != "ko" ]; then
      echo "Error: Project locale must be 'ko'"
      exit 1
    fi
```

## Troubleshooting

### Commits Still in English

**Problem:** Commits are still generated in English despite setting locale to 'ko'.

**Solution:**
1. Check `.moai/config.json` exists and is valid JSON
2. Verify `project.locale` field is set correctly
3. Restart your development session
4. Check for `MOAI_LOCALE` environment variable override

### Locale Not Recognized

**Problem:** Invalid locale error or fallback to English.

**Solution:**
1. Ensure locale is one of: 'ko', 'en', 'ja', 'zh'
2. Check for typos in config file
3. Validate JSON syntax

### git-manager Not Using Locale

**Problem:** git-manager agent ignores locale setting.

**Solution:**
1. Ensure you're using the latest version of MoAI-ADK
2. Check that `.moai/config.json` is in the project root
3. Verify the agent has read access to the config file

## Migration Guide

### From English-only to Localized

If you have an existing project using English commits:

1. Add locale to `.moai/config.json`:
```json
{
  "project": {
    "locale": "ko"
  }
}
```

2. Update your git-manager agent (automatic in latest version)

3. New commits will use the specified locale

4. Old commits remain unchanged (no need to rewrite history)

## Related Documentation

- [Git Manager Agent](../agents/git-manager.md)
- [TDD Workflow](./tdd-workflow.md)
- [@TAG System](./tag-system.md)
- [MoAI Configuration](./configuration.md)
