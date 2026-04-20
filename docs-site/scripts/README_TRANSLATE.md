# Translation Script for MoAI-ADK Documentation

## Overview

The `translate.mjs` script translates Korean MDX content to English, Japanese, and Chinese while preserving code blocks, frontmatter structure, and technical terms.

## Location

`/Users/goos/MoAI/moai-docs/scripts/translate.mjs`

## Features

| Feature | Description |
|---------|-------------|
| Code Block Preservation | Does not translate content inside \`\`\` code fences |
| Frontmatter Handling | Translates title and description, preserves structure |
| Technical Terms | Preserves terms like SPEC, DDD, TRUST 5, MoAI, etc. |
| Import Statements | Preserves import statements unchanged |
| Meta File Generation | Creates _meta.ts files for each locale directory |
| Batch Processing | Translates multiple files in one run |
| Dry Run Mode | Preview changes without writing files |

## Usage

```bash
# Translate all Korean MDX files to all languages
node scripts/translate.mjs

# Translate only to English
node scripts/translate.mjs --target-locale en

# Dry run to see what would be translated
node scripts/translate.mjs --dry-run

# Translate specific directory
node scripts/translate.mjs --source-dir content/core-concepts --target-locale en

# Show help
node scripts/translate.mjs --help
```

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `--source-dir <dir>` | Source directory | content/ |
| `--target-locale <loc>` | Target locale: en, ja, zh | all |
| `--files <pattern>` | File pattern to translate | **/*.mdx |
| `--dry-run` | Show what would be translated without writing | false |
| `--api-key <key>` | Translation API key | null |
| `--help` | Show help message | - |

## Supported Locales

| Locale | Language | Directory |
|--------|----------|-----------|
| ko | Korean | content/ko/ |
| en | English | content/en/ |
| ja | Japanese | content/ja/ |
| zh | Chinese | content/zh/ |

## Preserved Technical Terms

The following terms are never translated:

- SPEC, DDD, TRUST 5
- MoAI, MoAI-ADK
- Claude Code, EARS, TDD
- OWASP, AST, LSP, MCP
- MDX, JSON, YAML, API, CLI
- And more...

## Code Block Languages Preserved

All code blocks with these language identifiers are preserved:

- Shell: bash, sh, zsh
- JavaScript/TypeScript: js, jsx, ts, tsx
- Python: py, python
- Go: go, golang
- Java, Kotlin, Scala
- C/C++/C#: c, cpp, csharp, cs
- Ruby, PHP, Elixir, R
- Swift, Dart
- Config: yaml, yml, json, toml
- Documentation: markdown, md
- Diagrams: mermaid
- And more...

## Architecture

The script follows this workflow:

```
1. Parse MDX Content
   ├─ Extract frontmatter (YAML)
   ├─ Extract import statements
   ├─ Extract code blocks
   └─ Identify prose content

2. Translation
   ├─ Preserve code blocks (no translation)
   ├─ Translate frontmatter fields (title, description)
   ├─ Translate prose content
   └─ Preserve technical terms

3. Output
   ├─ Create locale directories
   ├─ Write translated MDX files
   └─ Generate _meta.ts files
```

## Integration with Translation APIs

Currently, the script uses a placeholder translation function. To integrate with a real translation service, replace the `translateText` function with API calls to:

- Google Cloud Translation API
- DeepL API
- Azure Translator
- AWS Translate

Example integration pattern:

```javascript
async function translateText(text, sourceLocale, targetLocale, apiKey) {
  // Preserve technical terms
  const preserved = {};
  let tempText = text;

  for (const term of CONFIG.preserveTerms) {
    const placeholder = `__PRESERVE_${Math.random().toString(36)}__`;
    preserved[placeholder] = term;
    tempText = tempText.replace(new RegExp(`\\b${term}\\b`, 'g'), placeholder);
  }

  // Call translation API
  const translated = await callTranslationAPI(tempText, sourceLocale, targetLocale, apiKey);

  // Restore preserved terms
  for (const [placeholder, term] of Object.entries(preserved)) {
    translated = translated.replace(new RegExp(placeholder, 'g'), term);
  }

  return translated;
}
```

## File Structure After Translation

```
content/
├── ko/                    # Source (Korean)
│   ├── core-concepts/
│   │   ├── _meta.ts
│   │   ├── ddd.mdx
│   │   └── ...
│   └── ...
├── en/                    # Translated (English)
│   ├── core-concepts/
│   │   ├── _meta.ts
│   │   ├── ddd.mdx
│   │   └── ...
│   └── ...
├── ja/                    # Translated (Japanese)
│   └── ...
└── zh/                    # Translated (Chinese)
    └── ...
```

## Notes

- The script processes only files with `.mdx` extension
- Directories like `node_modules`, `.git`, `dist`, `build` are skipped
- Existing translated files will be overwritten
- Use `--dry-run` to preview changes before writing
