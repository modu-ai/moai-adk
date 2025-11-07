# Documentation Scripts

This directory contains utility scripts for managing MoAI-ADK documentation.

## Available Scripts

### 📊 check_translation_status.py

Analyzes translation coverage across multiple languages (English, Japanese, Chinese) and generates status reports.

#### Usage

```bash
cd docs
python scripts/check_translation_status.py
```

#### Outputs

1. **JSON Report**: `.moai/reports/translation-status.json`
   - Machine-readable translation statistics
   - Missing files per language
   - Completion percentages

2. **Markdown Dashboard**: `src/translation-status.md`
   - Human-readable progress visualization
   - Progress bars for each language
   - List of missing files

#### Features

- Automatic discovery of base Korean documentation
- Comparison with translated versions (en, ja, zh)
- Progress visualization with Unicode progress bars
- Missing file tracking per language
- Timestamped reports for tracking changes over time

#### Example Output

```
============================================================
TRANSLATION STATUS SUMMARY
============================================================

Base Files (Korean): 61

English (en):
  Translated: 10/61
  Completion: 16.39%
  Missing: 51 files

Japanese (ja):
  Translated: 10/61
  Completion: 16.39%
  Missing: 51 files

Chinese (zh):
  Translated: 13/61
  Completion: 21.31%
  Missing: 48 files

============================================================
```

### 🔧 build_docs.py

Builds the documentation site using MkDocs Material.

#### Usage

```bash
cd docs
python scripts/build_docs.py
```

## CI/CD Integration

### GitHub Actions Workflow

Translation status is automatically updated via GitHub Actions:

- **Schedule**: Daily at 00:00 UTC
- **Triggers**: Push to main/develop branches (docs changes)
- **Manual**: Workflow dispatch available

See `.github/workflows/translation-status-update.yml` for details.

## Development

### Requirements

```bash
pip install -r docs/requirements.txt
```

### Testing Scripts Locally

```bash
# Test translation status checker
cd docs
python scripts/check_translation_status.py

# Test documentation build
python scripts/build_docs.py

# Serve documentation locally
mkdocs serve
```

## Contributing

See [Translation Contributing Guide](../src/guides/contributing-translations.md) for information about contributing translations.

---

**Generated with Claude Code**

Co-Authored-By: Alfred <alfred@mo.ai.kr>
