# Translation Contributing Guide

Thank you for contributing to MoAI-ADK documentation translation! This guide will help you get started with translation work.

## 📊 Current Translation Status

To check the current translation status, see the [Translation Status Dashboard](translation-status.md).

## 🌍 Supported Languages

Currently supported languages:

- **English (en)** - English
- **Japanese (ja)** - Japanese
- **Chinese (zh)** - Chinese (Simplified)

## 🚀 Quick Start

### 1. Select a File to Translate

Check the list of missing files in the [Translation Status Dashboard](translation-status.md).

Files requiring translation are displayed for each language.

### 2. Understand File Structure

```
docs/src/
├── index.md                    # Korean (default)
├── getting-started/
│   ├── installation.md
│   └── quick-start.md
├── en/                         # English translation
│   ├── getting-started/
│   │   ├── installation.md
│   │   └── quick-start.md
├── ja/                         # Japanese translation
│   └── ...
└── zh/                         # Chinese translation
    └── ...
```

**Core Principles**:
- Korean originals: `docs/src/` root and subdirectories
- Translations: Maintain identical directory structure under `docs/src/{language-code}/`

### 3. Start Translation Work

#### Method A: Using GitHub Web UI

1. Find the file to translate on GitHub
2. Click the "Edit" button
3. Write your translation
4. Click "Propose changes"
5. Create a Pull Request

#### Method B: Using Local Environment

```bash
# 1. Fork and clone repository
git clone https://github.com/YOUR_USERNAME/moai-adk.git
cd moai-adk

# 2. Create translation branch
git checkout -b translate-ja-getting-started

# 3. Create translation file
# Example: Translate docs/src/getting-started/installation.md to Japanese
mkdir -p docs/src/ja/getting-started
cp docs/src/getting-started/installation.md docs/src/ja/getting-started/installation.md

# 4. Translate the file (open with your editor)

# 5. Check changes
python docs/scripts/check_translation_status.py

# 6. Commit and push
git add docs/src/ja/
git commit -m "docs: Add Japanese translation for installation guide"
git push origin translate-ja-getting-started

# 7. Create Pull Request on GitHub
```

## 📝 Translation Guidelines

### Terminology Consistency

Keep technical terms in their original form when possible, or translate and add the original in parentheses.

| Korean | English | Japanese | Chinese |
|--------|---------|----------|---------|
| SPEC | SPEC | SPEC | SPEC |
| TAG | TAG | TAG | TAG |
| Alfred | Alfred | Alfred | Alfred |
| 테스트 주도 개발 | Test-Driven Development (TDD) | テスト駆動開発 (TDD) | 测试驱动开发 (TDD) |
| 요구사항 | Requirements | 要件 | 需求 |
| 구현 | Implementation | 実装 | 实现 |

### Writing Style

- Maintain **polite and professional tone**
- Use **second person**: "you" (English), "당신" (Korean), "あなた" (Japanese), "您" (Chinese)
- Use **clear and concise expressions**

### Code Blocks

Keep code examples as-is without translation:

```python
# Keep code as-is (do not translate comments in code blocks)
def hello_world():
    print("Hello, World!")
```

### Links and References

- **Internal links**: Change to the translated language path if the page exists
  ```markdown
  <!-- Korean -->
  [설치 가이드](getting-started/installation.md)

  <!-- English -->
  [Installation Guide](en/getting-started/installation.md)
  ```

- **External links**: Change to the language-specific version if available

## ✅ Quality Checklist

After completing your translation, verify the following:

- [ ] **File structure**: Maintains identical directory structure to Korean original
- [ ] **File name**: Uses same filename as original
- [ ] **Markdown syntax**: No errors in headings, links, code blocks, etc.
- [ ] **Terminology consistency**: Key terms are translated consistently
- [ ] **Code preservation**: Code examples are kept as-is
- [ ] **Link validation**: Internal/external links work correctly
- [ ] **Local build test**: Verify rendering with `mkdocs serve`

## 🔍 Testing Your Translation

To check your translation results locally:

```bash
# 1. Install documentation dependencies
cd docs
pip install -r requirements.txt

# 2. Run MkDocs development server
mkdocs serve

# 3. Check in browser
# http://localhost:8000
```

## 🤝 Review Process

1. **Create Pull Request**: Submit PR after completing translation
2. **Automated validation**: CI/CD automatically validates syntax and links
3. **Review**: Maintainer or language-specific reviewer examines the work
4. **Revision request**: Apply feedback if needed
5. **Merge**: Merge to main branch after approval

## 📧 Contact

If you have questions or need help:

- **GitHub Issues**: [moai-adk/issues](https://github.com/modu-ai/moai-adk/issues)
- **GitHub Discussions**: [moai-adk/discussions](https://github.com/modu-ai/moai-adk/discussions)

## 🎖️ Contributors

Contributors to translations:

- The contributor list can be found at [Contributors](https://github.com/modu-ai/moai-adk/graphs/contributors).

---

**Thank you!** Your contributions help MoAI-ADK reach more users worldwide. 🌏
