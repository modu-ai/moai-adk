# Technical Translation System - API Reference

## Configuration Reference

### Environment Variables

```bash
# Required
export OPENAI_API_KEY="sk-..."          # OpenAI API key

# Optional
export DEEPL_API_KEY="..."              # DeepL API key (fallback)
export AZURE_TRANSLATOR_KEY="..."       # Azure Translator key
export TRANSLATION_DB_PATH=".moai/translation/glossary.db"
export TRANSLATION_CACHE_DIR=".moai/translation/cache"
```

### Configuration File Format

```json
{
  "translation": {
    "primary_provider": "openai_gpt4",
    "fallback_providers": ["deepl", "azure"],
    "default_source_lang": "en",
    "default_target_lang": "ko",
    "glossary": {
      "enabled": true,
      "database": ".moai/translation/glossary.db",
      "auto_update": true
    },
    "validation": {
      "enabled": true,
      "min_quality_score": 80,
      "strict_mode": false
    },
    "cache": {
      "enabled": true,
      "ttl_hours": 168,
      "max_size_mb": 500
    }
  }
}
```

## API Reference

### Core Classes

#### TechnicalTranslationEngine

```python
class TechnicalTranslationEngine:
    """Main translation engine with multi-provider support"""
    
    def __init__(self, openai_key: str, glossary_path: Optional[str] = None)
    
    def translate(
        text: str,
        config: TranslationConfig,
        context: Optional[str] = None
    ) -> Dict[str, any]
    
    def batch_translate(
        texts: List[str],
        config: TranslationConfig
    ) -> List[Dict[str, any]]
    
    def estimate_cost(text: str, model: str = 'gpt-4o') -> float
```

#### TranslationConfig

```python
@dataclass
class TranslationConfig:
    provider: TranslationProvider
    source_lang: str           # 'en', 'ko', 'ja' or full names
    target_lang: str           # 'en', 'ko', 'ja' or full names
    domain: str = "technical"  # Domain for terminology
    preserve_formatting: bool = True
    use_glossary: bool = True
    temperature: float = 0.3   # 0.0-1.0, lower = more consistent
    max_tokens: int = 4000
```

#### DocumentProcessor

```python
class DocumentProcessor:
    """Handle various document formats"""
    
    def process_markdown_file(file_path: str) -> Dict
    def process_docx_file(file_path: str) -> Dict
    def save_translated_document(
        translated_sections: List[Dict],
        output_path: str,
        format: str = 'markdown'  # 'markdown', 'html', 'docx'
    )
```

#### GlossaryManager

```python
class GlossaryManager:
    """Terminology and translation memory management"""
    
    def __init__(self, db_path: str = "glossary.db")
    
    def add_term(
        source_term: str,
        target_term: str,
        source_lang: str,
        target_lang: str,
        domain: str = "technical",
        context: str = None
    )
    
    def bulk_import_glossary(
        file_path: str,
        format: str = "json"  # 'json', 'tsv', 'csv'
    )
    
    def export_for_openai(
        source_lang: str,
        target_lang: str,
        domain: str = None
    ) -> Dict[str, str]
    
    def search_memory(
        source_text: str,
        source_lang: str,
        target_lang: str,
        threshold: float = 0.9
    ) -> Optional[str]
```

#### TranslationValidator

```python
class TranslationValidator:
    """Quality validation system"""
    
    def validate(
        original: str,
        translation: str,
        source_lang: str,
        target_lang: str,
        glossary: Dict = None
    ) -> Dict
    
    def generate_validation_report(
        validation_result: Dict,
        output_path: str = None
    ) -> str
```

## Language Codes

### Supported Languages

| Language | Code | Full Name | Notes |
|----------|------|-----------|-------|
| English | en | English | Default source |
| Korean | ko | Korean | Formal style default |
| Japanese | ja | Japanese | Polite form default |

### Language Pair Specifics

#### English → Korean
- Formal endings (습니다/합니다)
- Technical terms often kept in English
- Length ratio: 0.8-1.3x original

#### English → Japanese
- Polite form (です/ます)
- Katakana for foreign terms
- Length ratio: 0.9-1.4x original

#### Korean → English
- SOV to SVO conversion
- Honorifics removed
- Length ratio: 0.7-1.2x original

#### Korean ↔ Japanese
- Similar grammar structure
- Hanja/Kanji mapping
- Length ratio: 0.9-1.1x original

## Glossary Format

### JSON Format

```json
{
  "English_Korean": {
    "API": "API",
    "endpoint": "엔드포인트",
    "authentication": "인증",
    "authorization": "권한 부여",
    "middleware": "미들웨어",
    "database": "데이터베이스",
    "cache": "캐시",
    "token": "토큰"
  },
  "English_Japanese": {
    "API": "API",
    "endpoint": "エンドポイント",
    "authentication": "認証",
    "authorization": "認可",
    "middleware": "ミドルウェア"
  }
}
```

### TSV Format

```tsv
source_term	target_term	source_lang	target_lang	domain
API	API	en	ko	technical
endpoint	엔드포인트	en	ko	web
authentication	인증	en	ko	security
```

## Validation Rules

### Critical Checks
- Code block preservation (must be 100% identical)
- URL preservation (all URLs must be retained)
- Markdown structure (headers, lists maintained)

### High Priority
- Variable name preservation
- Technical term consistency
- Image/link references

### Medium Priority
- Length ratio within expected bounds
- List structure similarity
- Table formatting

### Low Priority
- Glossary compliance
- Style consistency
- Formatting preferences

## Performance Metrics

### Translation Speed

| Document Size | GPT-4o | GPT-4o-mini | Batch Mode |
|--------------|--------|-------------|------------|
| 1 page | 2-3s | 1-2s | N/A |
| 10 pages | 20-30s | 10-15s | 15-20s |
| 50 pages | 2-3min | 1-2min | 1-1.5min |
| 100 pages | 5-7min | 3-4min | 2-3min |

### Cost Estimates (USD)

| Document Size | GPT-4o | GPT-4o-mini | Savings |
|--------------|--------|-------------|---------|
| 1 page (~500 words) | $0.015 | $0.002 | 87% |
| 10 pages | $0.15 | $0.02 | 87% |
| 50 pages | $0.75 | $0.10 | 87% |
| 100 pages | $1.50 | $0.20 | 87% |

### Quality Scores

| Provider | Technical Accuracy | Consistency | Speed |
|----------|-------------------|-------------|-------|
| GPT-4o | 95% | 92% | Medium |
| GPT-4o-mini | 88% | 85% | Fast |
| DeepL | 93% | 90% | Fast |
| Azure | 85% | 82% | Very Fast |

## Error Codes

| Code | Description | Resolution |
|------|-------------|------------|
| E001 | Invalid API key | Check OPENAI_API_KEY environment variable |
| E002 | Rate limit exceeded | Implement retry with exponential backoff |
| E003 | Token limit exceeded | Split document into smaller sections |
| E004 | Unsupported language pair | Check supported languages list |
| E005 | Glossary database error | Check database path and permissions |
| E006 | Document format not supported | Convert to markdown first |
| E007 | Validation failed | Review validation report and fix issues |
| E008 | Network error | Check internet connection |
| E009 | Insufficient quota | Check OpenAI account balance |
| E010 | Cache corruption | Clear cache directory |

## Best Practices

### For Optimal Results

1. **Pre-process Documents**
   - Clean formatting inconsistencies
   - Mark code blocks properly
   - Validate markdown syntax

2. **Glossary Management**
   - Keep domain-specific glossaries
   - Update regularly based on feedback
   - Review auto-generated terms

3. **Quality Assurance**
   - Always validate critical documents
   - Review bilingual output for important content
   - Maintain translation memory

4. **Cost Optimization**
   - Use GPT-4o-mini for drafts
   - Enable caching for repeated content
   - Batch process when possible

### Common Patterns

#### Pattern: Technical Book Translation

```python
# 1. Setup
engine = TechnicalTranslationEngine(api_key)
processor = DocumentProcessor()
validator = TranslationValidator()

# 2. Process chapters
for chapter in chapters:
    doc = processor.process_markdown_file(chapter)
    
    config = TranslationConfig(
        provider=TranslationProvider.OPENAI_GPT4,
        source_lang="en",
        target_lang="ko",
        domain="programming"
    )
    
    result = engine.translate(doc['content'], config)
    
    # 3. Validate
    validation = validator.validate(
        doc['content'],
        result['translation'],
        "en", "ko"
    )
    
    if validation['quality_score'] > 85:
        processor.save_translated_document(
            result['translation'],
            output_path
        )
```

#### Pattern: API Documentation

```python
# Use consistent terminology
glossary = GlossaryManager()
glossary.bulk_import_glossary("api_terms.json")

config = TranslationConfig(
    provider=TranslationProvider.OPENAI_GPT4,
    source_lang="en",
    target_lang="ja",
    domain="api_documentation",
    use_glossary=True,
    temperature=0.2  # Very consistent
)
```

#### Pattern: Incremental Translation

```python
# Translate and save to memory for reuse
for section in sections:
    # Check cache first
    cached = glossary.search_memory(
        section, "en", "ko"
    )
    
    if not cached:
        result = engine.translate(section, config)
        
        # Save to memory if good quality
        if result['validation']['quality_score'] > 80:
            glossary.add_to_memory(
                section,
                result['translation'],
                "en", "ko",
                result['validation']['quality_score']
            )
```

## Integration Examples

### With MoAI-ADK

```python
from moai_adk import Task

# Delegate translation task
result = await Task(
    subagent_type="translation-expert",
    prompt=f"Translate technical documentation from {source_lang} to {target_lang}",
    context={
        "document_path": "docs/api.md",
        "glossary": "technical_terms.json",
        "quality_threshold": 85
    }
)
```

### With CI/CD Pipeline

```yaml
# .github/workflows/translate.yml
name: Translate Documentation

on:
  push:
    paths:
      - 'docs/en/**/*.md'

jobs:
  translate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Setup Python
        uses: actions/setup-python@v2
        with:
          python-version: '3.11'
      
      - name: Install dependencies
        run: |
          pip install openai markitdown
      
      - name: Translate to Korean
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
        run: |
          python translate.py docs/en docs/ko -s en -t ko --batch
      
      - name: Validate translations
        run: |
          python translate.py docs/ko --validate --min-quality 80
```

### With Web Service

```python
from flask import Flask, request, jsonify
import asyncio

app = Flask(__name__)
engine = TechnicalTranslationEngine(api_key)

@app.route('/translate', methods=['POST'])
def translate_endpoint():
    data = request.json
    
    config = TranslationConfig(
        provider=TranslationProvider.OPENAI_GPT4,
        source_lang=data['source_lang'],
        target_lang=data['target_lang']
    )
    
    result = engine.translate(
        data['text'],
        config
    )
    
    return jsonify({
        'translation': result['translation'],
        'quality_score': result['validation']['quality_score'],
        'cost': result['usage']['estimated_cost']
    })
```

## Troubleshooting

### Common Issues

#### Issue: Code blocks are being translated

**Solution**: Check preservation patterns are working:

```python
# Debug code preservation
preserved, map = engine._preserve_technical_content(text)
print(f"Preserved blocks: {len(map)}")
for placeholder, content in map.items():
    print(f"{placeholder}: {content[:50]}...")
```

#### Issue: Inconsistent terminology

**Solution**: Verify glossary is loaded:

```python
glossary_terms = glossary.export_for_openai("en", "ko")
print(f"Loaded {len(glossary_terms)} glossary terms")
```

#### Issue: High API costs

**Solution**: Optimize model selection:

```python
optimizer = CostOptimizer()
model = optimizer.select_optimal_model(text, {'budget_conscious': True})
print(f"Selected model: {model}")
print(f"Estimated cost: ${optimizer.estimate_cost(text, model):.4f}")
```

## Support & Resources

- **GitHub Issues**: Report bugs and request features
- **Documentation**: Full guides and tutorials
- **Community**: Discord/Slack channels
- **Enterprise Support**: Contact for SLA options

---

**Version**: 1.0.0
**Last Updated**: 2025-11-16
**API Compatibility**: OpenAI API v1, DeepL API v2
