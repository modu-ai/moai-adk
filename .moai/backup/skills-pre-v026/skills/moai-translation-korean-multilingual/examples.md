# Technical Translation Examples

## Quick Start Examples

### Example 1: Simple Markdown Translation

```python
from openai import OpenAI

# Initialize
client = OpenAI(api_key="YOUR_API_KEY")

# Simple markdown document
document = """
# API Documentation

The REST API provides the following endpoints:

## Authentication

Use JWT tokens for authentication:

```python
headers = {
    'Authorization': f'Bearer {token}'
}
```

## Endpoints

- `GET /api/users` - List all users
- `POST /api/users` - Create new user
- `GET /api/users/{id}` - Get specific user
"""

# Translate to Korean
response = client.chat.completions.create(
    model="gpt-4o",
    messages=[
        {"role": "system", "content": """You are a technical translator.
        Translate from English to Korean.
        Preserve all code blocks and technical terms exactly.
        Use formal Korean (습니다/합니다)."""},
        {"role": "user", "content": document}
    ],
    temperature=0.3
)

translated = response.choices[0].message.content
print(translated)
```

**Output:**
```markdown
# API 문서

REST API는 다음 엔드포인트를 제공합니다:

## 인증

인증을 위해 JWT 토큰을 사용합니다:

```python
headers = {
    'Authorization': f'Bearer {token}'
}
```

## 엔드포인트

- `GET /api/users` - 모든 사용자 목록 조회
- `POST /api/users` - 새 사용자 생성
- `GET /api/users/{id}` - 특정 사용자 조회
```

### Example 2: Technical Book Chapter

```python
from pathlib import Path

# Load the complete translation system
engine = TechnicalTranslationEngine(openai_key="YOUR_KEY")
processor = DocumentProcessor()

# Read chapter file
chapter_content = """
# Chapter 3: Understanding Async/Await in Python

## Introduction

Asynchronous programming in Python has evolved significantly with the introduction 
of `async` and `await` keywords. This chapter explores how to leverage these 
features for better performance.

## Basic Concepts

### Coroutines

A coroutine is declared with the `async def` syntax:

```python
async def fetch_data(url):
    async with aiohttp.ClientSession() as session:
        async with session.get(url) as response:
            return await response.text()
```

### Event Loop

The event loop is the core of async execution:

```python
import asyncio

async def main():
    tasks = [fetch_data(url) for url in urls]
    results = await asyncio.gather(*tasks)
    return results

# Run the event loop
asyncio.run(main())
```

## Best Practices

1. **Don't block the event loop**: Avoid synchronous operations
2. **Use `asyncio.gather()` for concurrent tasks**
3. **Handle exceptions properly with try-except blocks**
"""

# Configure translation
config = TranslationConfig(
    provider=TranslationProvider.OPENAI_GPT4,
    source_lang="English",
    target_lang="Korean",
    domain="programming",
    use_glossary=True
)

# Translate with validation
result = engine.translate(chapter_content, config)

print(f"Quality Score: {result['validation']['quality_score']}%")
print("\nTranslated Content:")
print(result['translation'])
```

### Example 3: Batch Processing Multiple Files

```python
import asyncio
from pathlib import Path

async def translate_documentation():
    """Translate entire documentation folder"""
    
    pipeline = BatchTranslationPipeline({
        'openai_key': 'YOUR_API_KEY',
        'glossary_db': 'technical_glossary.db'
    })
    
    # Translate all markdown files
    results = await pipeline.translate_batch(
        input_dir='docs/en',
        output_dir='docs/ko',
        source_lang='English',
        target_lang='Korean',
        file_pattern='*.md'
    )
    
    # Print summary
    print(f"Translated {results['statistics']['successful']} files")
    print(f"Average quality: {results['statistics']['average_quality']:.1f}%")
    
    # List any failures
    if results['failed']:
        print("\nFailed translations:")
        for failure in results['failed']:
            print(f"  - {failure['file']}: {failure['error']}")
    
    return results

# Run batch translation
results = asyncio.run(translate_documentation())
```

## Advanced Examples

### Example 4: Custom Glossary Management

```python
# Initialize glossary manager
glossary = GlossaryManager(".moai/translation/glossary.db")

# Add technical terms
technical_terms = [
    ("authentication", "인증", "en", "ko", "security"),
    ("authorization", "권한 부여", "en", "ko", "security"),
    ("middleware", "미들웨어", "en", "ko", "web"),
    ("endpoint", "엔드포인트", "en", "ko", "api"),
    ("webhook", "웹훅", "en", "ko", "api"),
    ("payload", "페이로드", "en", "ko", "api"),
    ("callback", "콜백", "en", "ko", "programming"),
    ("async", "비동기", "en", "ko", "programming"),
    ("coroutine", "코루틴", "en", "ko", "programming")
]

for term in technical_terms:
    glossary.add_term(*term)

# Import from TSV file
glossary.bulk_import_glossary("extended_glossary.tsv", format="tsv")

# Export for use with translation
ko_glossary = glossary.export_for_openai("en", "ko", domain="programming")
print(f"Loaded {len(ko_glossary)} Korean terms")

# Use in translation
engine = TechnicalTranslationEngine(
    openai_key="YOUR_KEY",
    glossary_path="glossary.json"
)
```

### Example 5: Quality Validation Workflow

```python
# Complete validation example
validator = TranslationValidator()

original = """
# Database Configuration

Configure your PostgreSQL connection:

```yaml
database:
  host: localhost
  port: 5432
  name: myapp_db
  user: ${DB_USER}
  password: ${DB_PASSWORD}
```

Connection URL: `postgresql://user:pass@localhost:5432/myapp_db`
"""

translated = """
# 데이터베이스 설정

PostgreSQL 연결을 설정하세요:

```yaml
database:
  host: localhost
  port: 5432
  name: myapp_db
  user: ${DB_USER}
  password: ${DB_PASSWORD}
```

연결 URL: `postgresql://user:pass@localhost:5432/myapp_db`
"""

# Validate translation
validation_result = validator.validate(
    original,
    translated,
    source_lang="English",
    target_lang="Korean"
)

# Generate report
report = validator.generate_validation_report(
    validation_result,
    output_path="validation_report.md"
)

print(f"Validation Score: {validation_result['quality_score']:.1f}%")
print(f"Issues Found: {len(validation_result['issues'])}")

# Check specific validations
for check in validation_result['passed_checks']:
    print(f"✅ {check}")

for issue in validation_result['issues']:
    print(f"❌ {issue['type']}: {issue.get('details', 'N/A')}")
```

### Example 6: Bilingual Review Generation

```python
# Create side-by-side review HTML
processor = DocumentProcessor()

sections = [
    {
        'content': "# Introduction\n\nThis is the introduction.",
        'translated': "# 소개\n\n이것은 소개입니다.",
        'type': 'text',
        'quality_score': 95
    },
    {
        'content': "```python\ndef hello():\n    print('Hello')\n```",
        'translated': "```python\ndef hello():\n    print('Hello')\n```",
        'type': 'code_heavy',
        'quality_score': 100
    }
]

# Generate HTML review
html_content = processor._generate_bilingual_html(sections)

# Save review file
with open("translation_review.html", "w", encoding="utf-8") as f:
    f.write(html_content)

print("Review file generated: translation_review.html")
```

### Example 7: Translation Memory Usage

```python
# Setup with translation memory
glossary = GlossaryManager()

# Translate with memory caching
def translate_with_memory(text, source_lang="en", target_lang="ko"):
    # Check memory first
    cached = glossary.search_memory(text, source_lang, target_lang)
    
    if cached:
        print(f"Found in translation memory")
        return cached
    
    # Not in memory, translate
    engine = TechnicalTranslationEngine(openai_key="YOUR_KEY")
    config = TranslationConfig(
        provider=TranslationProvider.OPENAI_GPT4_MINI,  # Use cheaper model
        source_lang=source_lang,
        target_lang=target_lang
    )
    
    result = engine.translate(text, config)
    
    # Add to memory if good quality
    if result['validation']['quality_score'] > 80:
        glossary.add_to_memory(
            text,
            result['translation'],
            source_lang,
            target_lang,
            result['validation']['quality_score']
        )
        print(f"Added to translation memory (score: {result['validation']['quality_score']:.1f})")
    
    return result['translation']

# Example usage - second call will use cache
text = "The API returns a JSON response with status code 200."
translation1 = translate_with_memory(text)  # Translates via API
translation2 = translate_with_memory(text)  # Uses memory cache
```

### Example 8: Cost-Optimized Translation

```python
# Cost-aware translation
optimizer = CostOptimizer()

documents = [
    {"text": "Short simple text", "priority": "low"},
    {"text": "Complex technical documentation with code...", "priority": "high"},
    {"text": "Medium complexity document...", "priority": "medium"}
]

total_cost = 0

for doc in documents:
    # Estimate cost
    estimated_cost = optimizer.estimate_cost(doc['text'])
    
    # Select model based on priority
    if doc['priority'] == 'high':
        model = 'gpt-4o'
    else:
        model = optimizer.select_optimal_model(
            doc['text'],
            {'budget_conscious': True}
        )
    
    print(f"Document: {doc['text'][:30]}...")
    print(f"  Model: {model}")
    print(f"  Estimated cost: ${estimated_cost:.4f}")
    
    total_cost += estimated_cost

print(f"\nTotal estimated cost: ${total_cost:.2f}")
```

### Example 9: Word Document Translation

```python
# Translate Word documents using MarkItDown
from markitdown import MarkItDown

processor = DocumentProcessor()
engine = TechnicalTranslationEngine(openai_key="YOUR_KEY")

# Process Word document
doc_data = processor.process_docx_file("technical_manual.docx")

# Translate markdown content
config = TranslationConfig(
    provider=TranslationProvider.OPENAI_GPT4,
    source_lang="English",
    target_lang="Japanese",
    domain="technical_manual"
)

result = engine.translate(doc_data['markdown'], config)

# Save as translated markdown
with open("technical_manual_ja.md", "w", encoding="utf-8") as f:
    f.write(result['translation'])

# Or save as HTML for review
processor.save_translated_document(
    [{'translated': result['translation']}],
    "technical_manual_ja.html",
    format="html"
)
```

### Example 10: CLI Usage

```bash
# Single file translation
python translate.py input.md output.md -s en -t ko --api-key YOUR_KEY

# Batch translation with pattern
python translate.py docs/en docs/ko -s en -t ja --batch --pattern "*.md"

# With glossary
python translate.py input.md output.md -s en -t ko --glossary tech_terms.json

# Validation only
python translate.py translated.md --validate -s en -t ko

# Cost estimation
python translate.py input.md --estimate-cost -s en -t ko
```

## Integration Examples

### With FastAPI

```python
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

app = FastAPI()
engine = TechnicalTranslationEngine(openai_key="YOUR_KEY")

class TranslationRequest(BaseModel):
    text: str
    source_lang: str = "en"
    target_lang: str = "ko"
    use_glossary: bool = True
    validate: bool = True

class TranslationResponse(BaseModel):
    translation: str
    quality_score: float
    cost: float
    issues: list

@app.post("/translate", response_model=TranslationResponse)
async def translate(request: TranslationRequest):
    try:
        config = TranslationConfig(
            provider=TranslationProvider.OPENAI_GPT4,
            source_lang=request.source_lang,
            target_lang=request.target_lang,
            use_glossary=request.use_glossary
        )
        
        result = engine.translate(request.text, config)
        
        return TranslationResponse(
            translation=result['translation'],
            quality_score=result['validation']['quality_score'],
            cost=result['usage']['estimated_cost'],
            issues=result['validation']['issues']
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
```

### With Streamlit UI

```python
import streamlit as st

st.title("Technical Document Translator")

# Input section
source_lang = st.selectbox("Source Language", ["English", "Korean", "Japanese"])
target_lang = st.selectbox("Target Language", ["Korean", "English", "Japanese"])

input_text = st.text_area("Enter text to translate", height=200)

# Options
col1, col2 = st.columns(2)
with col1:
    use_glossary = st.checkbox("Use Glossary", value=True)
    validate = st.checkbox("Validate Quality", value=True)
with col2:
    model = st.radio("Model", ["GPT-4o", "GPT-4o-mini"])

# Translate button
if st.button("Translate"):
    with st.spinner("Translating..."):
        engine = TechnicalTranslationEngine(openai_key=st.secrets["OPENAI_KEY"])
        
        config = TranslationConfig(
            provider=TranslationProvider.OPENAI_GPT4 if model == "GPT-4o" else TranslationProvider.OPENAI_GPT4_MINI,
            source_lang=source_lang,
            target_lang=target_lang,
            use_glossary=use_glossary
        )
        
        result = engine.translate(input_text, config)
        
        # Display results
        st.subheader("Translation")
        st.text_area("Result", result['translation'], height=200)
        
        # Metrics
        col1, col2, col3 = st.columns(3)
        col1.metric("Quality Score", f"{result['validation']['quality_score']:.1f}%")
        col2.metric("Cost", f"${result['usage']['estimated_cost']:.4f}")
        col3.metric("Tokens", result['usage']['total_tokens'])
        
        # Validation issues
        if validate and result['validation']['issues']:
            st.warning(f"Found {len(result['validation']['issues'])} validation issues")
            for issue in result['validation']['issues']:
                st.error(f"{issue['type']}: {issue.get('details', 'N/A')}")
```

## Common Patterns

### Pattern: Preserving Markdown Tables

```python
table_content = """
| Feature | Description | Status |
|---------|-------------|--------|
| Authentication | JWT-based auth | ✅ Complete |
| Database | PostgreSQL integration | 🚧 In Progress |
| Caching | Redis support | ❌ Not Started |
"""

# Tables need special handling
result = engine.translate(table_content, config)
# Verify table structure is preserved
```

### Pattern: Handling Mixed Languages

```python
mixed_content = """
The function `getData()` returns a Promise.
韓国語のテキストも含まれています。
This is mixed with English text.
"""

# Detect and handle mixed languages appropriately
config.context = "Mixed language technical document"
```

### Pattern: Progressive Translation

```python
# For large documents, translate progressively
def progressive_translate(document_path):
    sections = processor._split_into_sections(document_content)
    translated_sections = []
    
    progress_bar = tqdm(sections, desc="Translating")
    for section in progress_bar:
        if section['type'] == 'code_heavy':
            # Skip or handle differently
            translated_sections.append(section)
        else:
            result = engine.translate(section['content'], config)
            section['translated'] = result['translation']
            translated_sections.append(section)
            
            progress_bar.set_postfix(quality=f"{result['validation']['quality_score']:.0f}%")
    
    return translated_sections
```

---

**Note**: Replace `YOUR_API_KEY` with actual API keys in production. Consider using environment variables or secure key management systems.
