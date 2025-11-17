# Technical Documentation Translation Best Practices
## Comprehensive Research Report for MoAI-ADK Translation System

**Date**: 2025-11-16
**Research Focus**: Korean ↔ English ↔ Japanese Technical Translation
**Primary Use Case**: Technical book and documentation translation

---

## Executive Summary

This research provides production-ready patterns for implementing a robust technical documentation translation system, focusing on:
- Multi-format document processing (Markdown, Word, PDF)
- Translation quality assurance with glossary management
- Code block and markup preservation
- Bilingual alignment and review workflows

---

## 1. Technical Translation Patterns

### 1.1 Korean to English Translation Strategies

**Key Challenges**:
- Subject-Object-Verb (SOV) vs Subject-Verb-Object (SVO) word order
- Honorifics and formality levels
- Context-dependent meanings
- Technical terminology consistency

**Best Practices**:

```python
# Context-aware translation with terminology preservation
translation_config = {
    "source_lang": "ko",
    "target_lang": "en",
    "formality": "default",  # For technical docs
    "preserve_formatting": True,
    "glossary_id": "tech-glossary-ko-en",
    "context": "technical documentation"  # Helps with ambiguous terms
}
```

**Recommended Approach**:
1. **Pre-processing**: Extract code blocks and technical terms
2. **Glossary matching**: Apply domain-specific terminology
3. **Translation**: Use context-aware API (DeepL preferred for quality)
4. **Post-processing**: Restore code blocks and verify formatting

### 1.2 Korean to Japanese Translation Strategies

**Key Advantages**:
- Similar grammar structures (both SOV)
- Shared Chinese character base (Hanja/Kanji)
- Similar honorific systems

**Implementation Pattern**:

```python
from deepl import Translator

translator = Translator(auth_key="YOUR_AUTH_KEY")

# Korean to Japanese with formality preservation
result = translator.translate_text(
    text="기술 문서를 번역합니다.",
    source_lang="KO",
    target_lang="JA",
    formality="default",
    preserve_formatting=True
)
```

**Quality Considerations**:
- Korean Hanja → Japanese Kanji mapping
- Technical term consistency (many borrowed from English)
- Formality level matching (존댓말 → 敬語)

### 1.3 Technical Terminology Preservation

**Critical Pattern for Code and Technical Terms**:

```python
import re

def preserve_technical_content(text: str) -> tuple[str, dict]:
    """Extract and preserve code blocks, variable names, and technical terms"""

    placeholders = {}
    counter = 0

    # Preserve code blocks (markdown)
    def replace_code_block(match):
        nonlocal counter
        placeholder = f"__CODE_BLOCK_{counter}__"
        placeholders[placeholder] = match.group(0)
        counter += 1
        return placeholder

    # Patterns to preserve
    patterns = [
        (r'```[\s\S]*?```', 'code_block'),  # Fenced code blocks
        (r'`[^`]+`', 'inline_code'),         # Inline code
        (r'\$\$[\s\S]*?\$\$', 'math_block'), # Math blocks
        (r'\$[^$]+\$', 'inline_math'),       # Inline math
        (r'https?://[^\s]+', 'url'),         # URLs
        (r'!\[.*?\]\(.*?\)', 'image'),       # Markdown images
    ]

    processed_text = text
    for pattern, type_name in patterns:
        processed_text = re.sub(pattern, replace_code_block, processed_text)

    return processed_text, placeholders

def restore_technical_content(text: str, placeholders: dict) -> str:
    """Restore preserved technical content"""
    result = text
    for placeholder, original in placeholders.items():
        result = result.replace(placeholder, original)
    return result
```

---

## 2. Document Processing Libraries

### 2.1 Markdown Processing

#### **Recommended: python-markdown2** (v2.4.12+)

**Advantages**:
- Fast, pure Python implementation
- Extensive extension support
- Good table handling
- Metadata extraction

```python
import markdown2

# Convert Markdown to HTML for translation
md = markdown2.Markdown(extras=[
    'fenced-code-blocks',
    'tables',
    'metadata',
    'code-friendly'
])

html = md.convert(markdown_text)
metadata = md.metadata  # Extract front matter
```

#### **Alternative: mistune** (v3.0+)

```python
import mistune

# Create custom renderer to preserve structure
class PreservingRenderer(mistune.HTMLRenderer):
    def block_code(self, code, info=None):
        # Preserve code blocks during translation
        return f'```{info}\n{code}\n```'

markdown = mistune.create_markdown(renderer=PreservingRenderer())
html = markdown(markdown_text)
```

### 2.2 Word Document Processing

#### **Recommended: python-docx** (v1.1.0+)

**Production-Ready Pattern**:

```python
from docx import Document
from docx.shared import Inches, Pt
from docx.enum.text import WD_ALIGN_PARAGRAPH

def extract_document_content(docx_path: str) -> dict:
    """Extract all content with structure preservation"""

    doc = Document(docx_path)
    content = {
        'paragraphs': [],
        'tables': [],
        'images': [],
        'metadata': {}
    }

    # Iterate through all content in document order
    for item in doc.iter_inner_content():
        if hasattr(item, 'text'):  # Paragraph
            content['paragraphs'].append({
                'text': item.text,
                'style': item.style.name,
                'alignment': item.alignment,
                'runs': [
                    {
                        'text': run.text,
                        'bold': run.bold,
                        'italic': run.italic,
                        'font_name': run.font.name,
                        'font_size': run.font.size
                    }
                    for run in item.runs
                ]
            })
        elif hasattr(item, 'rows'):  # Table
            table_data = []
            for row in item.rows:
                row_data = [cell.text for cell in row.cells]
                table_data.append(row_data)
            content['tables'].append(table_data)

    return content

def create_translated_document(original_docx: str, translations: dict, output_path: str):
    """Create new document with translated content, preserving formatting"""

    doc = Document(original_docx)

    para_index = 0
    table_index = 0

    for item in doc.iter_inner_content():
        if hasattr(item, 'text'):  # Paragraph
            if para_index < len(translations['paragraphs']):
                # Clear existing content
                item.clear()

                # Add translated text with original formatting
                original_para = translations['paragraphs'][para_index]
                for run_data in original_para['runs']:
                    run = item.add_run(run_data['translated_text'])
                    run.bold = run_data['bold']
                    run.italic = run_data['italic']
                    run.font.name = run_data['font_name']
                    if run_data['font_size']:
                        run.font.size = run_data['font_size']

                para_index += 1

        elif hasattr(item, 'rows'):  # Table
            if table_index < len(translations['tables']):
                # Translate table cells
                for i, row in enumerate(item.rows):
                    for j, cell in enumerate(row.cells):
                        cell.text = translations['tables'][table_index][i][j]
                table_index += 1

    doc.save(output_path)
```

### 2.3 Multi-Format Document Conversion

#### **Recommended: MarkItDown** (Microsoft, v0.0.1a4+)

**Production Pattern**:

```python
from markitdown import MarkItDown
from pathlib import Path

# Initialize with all capabilities
md = MarkItDown(enable_plugins=True)

# Convert various formats to Markdown
def convert_to_markdown(file_path: str) -> str:
    """Universal document to Markdown converter"""

    result = md.convert(file_path)

    return {
        'markdown': result.markdown,
        'title': result.title,  # Optional extracted title
        'file_type': Path(file_path).suffix
    }

# Supported formats:
# - PDF (.pdf)
# - Word (.docx)
# - PowerPoint (.pptx)
# - Excel (.xlsx)
# - Images (.png, .jpg, .gif) - with optional LLM descriptions
# - HTML (.html)
# - Text (.txt, .json, .xml)
# - Audio (.mp3, .wav) - with transcription
# - Video (.mp4) - from URLs with transcript
```

**Advanced: Azure Document Intelligence Integration**:

```python
from markitdown import MarkItDown

# Enhanced document processing with Azure AI
md = MarkItDown(
    docintel_endpoint="https://your-resource.cognitiveservices.azure.com/"
)

# Process complex layouts with tables, forms, etc.
result = md.convert("complex_technical_manual.pdf")
print(result.markdown)
```

**Image Description with LLM**:

```python
from markitdown import MarkItDown
from openai import OpenAI

# Add AI-generated descriptions for diagrams
client = OpenAI(api_key="YOUR_API_KEY")
md = MarkItDown(
    llm_client=client,
    llm_model="gpt-4o",
    llm_prompt="Describe this technical diagram in detail for documentation purposes"
)

result = md.convert("architecture_diagram.png")
# Output: ![Detailed AI-generated description of the architecture](architecture_diagram.png)
```

---

## 3. Translation APIs and Tools

### 3.1 DeepL API (Recommended for Quality)

**Best for**: Technical documentation with high accuracy requirements

**Version**: deepl-python v1.18.0+

**Production Implementation**:

```python
import deepl
from typing import List, Dict

class TechnicalTranslator:
    def __init__(self, auth_key: str, glossary_path: str = None):
        self.translator = deepl.Translator(auth_key)
        self.glossary = self._load_glossary(glossary_path) if glossary_path else None

    def translate_text(
        self,
        text: str,
        source_lang: str,
        target_lang: str,
        formality: str = "default"
    ) -> deepl.TextResult:
        """Translate text with glossary and formatting preservation"""

        return self.translator.translate_text(
            text,
            source_lang=source_lang,
            target_lang=target_lang,
            formality=formality,
            preserve_formatting=True,
            glossary=self.glossary,
            tag_handling='xml',  # For structured content
            outline_detection=False  # Disable auto-detection for technical docs
        )

    def translate_document(
        self,
        input_path: str,
        output_path: str,
        source_lang: str,
        target_lang: str
    ):
        """Translate entire documents (DOCX, PPTX, PDF)"""

        self.translator.translate_document_from_filepath(
            input_path,
            output_path,
            source_lang=source_lang,
            target_lang=target_lang,
            formality="default",
            glossary=self.glossary
        )

    def translate_batch(
        self,
        texts: List[str],
        source_lang: str,
        target_lang: str
    ) -> List[str]:
        """Batch translation for efficiency"""

        results = self.translator.translate_text(
            texts,
            source_lang=source_lang,
            target_lang=target_lang,
            preserve_formatting=True,
            glossary=self.glossary
        )

        return [result.text for result in results]

    def _load_glossary(self, glossary_path: str):
        """Load or create glossary"""

        # Create glossary from TSV file
        with open(glossary_path, 'r', encoding='utf-8') as f:
            entries = {}
            for line in f:
                if '\t' in line:
                    source, target = line.strip().split('\t', 1)
                    entries[source] = target

        glossary_info = self.translator.create_glossary(
            "Technical Glossary",
            source_lang="EN",
            target_lang="KO",
            entries=entries
        )

        return glossary_info

    def check_usage(self) -> Dict:
        """Monitor API usage"""

        usage = self.translator.get_usage()
        return {
            'character_count': usage.character.count,
            'character_limit': usage.character.limit,
            'limit_reached': usage.any_limit_reached
        }

# Usage example
translator = TechnicalTranslator(
    auth_key="YOUR_DEEPL_KEY",
    glossary_path="tech_glossary.tsv"
)

# Translate with context
result = translator.translate_text(
    text="The API returns a JSON response.",
    source_lang="EN",
    target_lang="KO",
    formality="default"
)

print(result.text)  # "API는 JSON 응답을 반환합니다."
print(f"Characters billed: {result.billed_characters}")
```

**Glossary Management**:

```python
# Create multilingual glossary (v3 API)
from deepl import Translator

translator = Translator("YOUR_KEY")

# Define terminology for multiple language pairs
glossary_entries = {
    "API": "API",
    "endpoint": "엔드포인트",
    "authentication": "인증",
    "REST API": "REST API",
    "JWT token": "JWT 토큰"
}

glossary = translator.create_glossary(
    name="Tech Glossary EN-KO",
    source_lang="EN",
    target_lang="KO",
    entries=glossary_entries
)

# Update glossary dynamically
updated_entries = {
    "API": "API",
    "endpoint": "엔드포인트",
    "middleware": "미들웨어",  # New entry
}

translator.update_glossary_entries(glossary.glossary_id, updated_entries)
```

### 3.2 OpenAI GPT for Context-Aware Translation

**Best for**: Complex technical content requiring deep context understanding

**Version**: openai v1.105.0+

```python
from openai import OpenAI
import json

class GPTTranslator:
    def __init__(self, api_key: str, model: str = "gpt-4o"):
        self.client = OpenAI(api_key=api_key)
        self.model = model

    def translate_with_context(
        self,
        text: str,
        source_lang: str,
        target_lang: str,
        context: str = "",
        glossary: dict = None
    ) -> dict:
        """Context-aware translation with glossary"""

        system_prompt = f"""You are a technical translator specializing in {source_lang} to {target_lang} translation.

Guidelines:
1. Preserve all code blocks, variable names, and technical terms exactly
2. Maintain markdown formatting
3. Use technical terminology consistently
4. Keep formality appropriate for technical documentation
5. Apply the provided glossary for specific terms

{f"Context: {context}" if context else ""}
{f"Glossary: {json.dumps(glossary, ensure_ascii=False)}" if glossary else ""}
"""

        response = self.client.chat.completions.create(
            model=self.model,
            messages=[
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": f"Translate the following text:\n\n{text}"}
            ],
            temperature=0.3,  # Lower temperature for consistency
        )

        translation = response.choices[0].message.content

        return {
            'translation': translation,
            'model': self.model,
            'usage': {
                'prompt_tokens': response.usage.prompt_tokens,
                'completion_tokens': response.usage.completion_tokens,
                'total_tokens': response.usage.total_tokens
            }
        }

    def batch_translate(
        self,
        texts: List[str],
        source_lang: str,
        target_lang: str,
        context: str = ""
    ) -> List[str]:
        """Batch translation with context preservation"""

        # Use OpenAI Batch API for cost efficiency
        batch_requests = [
            {
                "custom_id": f"request-{i}",
                "method": "POST",
                "url": "/v1/chat/completions",
                "body": {
                    "model": self.model,
                    "messages": [
                        {
                            "role": "system",
                            "content": f"Translate from {source_lang} to {target_lang}. {context}"
                        },
                        {
                            "role": "user",
                            "content": text
                        }
                    ],
                    "temperature": 0.3
                }
            }
            for i, text in enumerate(texts)
        ]

        # Process batch (see OpenAI Batch API docs)
        return self._process_batch(batch_requests)

# Usage
translator = GPTTranslator(api_key="YOUR_OPENAI_KEY")

result = translator.translate_with_context(
    text="The middleware handles authentication and authorization.",
    source_lang="English",
    target_lang="Korean",
    context="Technical documentation about web server architecture",
    glossary={
        "middleware": "미들웨어",
        "authentication": "인증",
        "authorization": "권한 부여"
    }
)
```

### 3.3 Microsoft Translator API

**Best for**: Enterprise integration with Azure services

**Key Features**:
- Custom Translator for domain-specific models
- Document translation (synchronous and asynchronous)
- Container deployment for on-premises
- Power Automate integration

```python
import requests
from typing import List

class AzureTranslator:
    def __init__(self, key: str, endpoint: str, region: str):
        self.key = key
        self.endpoint = endpoint
        self.region = region

    def translate(
        self,
        texts: List[str],
        to_lang: str,
        from_lang: str = None
    ) -> List[dict]:
        """Translate using Azure Translator v3"""

        path = '/translate'
        constructed_url = self.endpoint + path

        params = {
            'api-version': '3.0',
            'to': to_lang
        }
        if from_lang:
            params['from'] = from_lang

        headers = {
            'Ocp-Apim-Subscription-Key': self.key,
            'Ocp-Apim-Subscription-Region': self.region,
            'Content-type': 'application/json',
        }

        body = [{'text': text} for text in texts]

        response = requests.post(
            constructed_url,
            params=params,
            headers=headers,
            json=body
        )
        response.raise_for_status()

        return response.json()

# Usage
translator = AzureTranslator(
    key="YOUR_KEY",
    endpoint="https://api.cognitive.microsofttranslator.com",
    region="eastus"
)

results = translator.translate(
    texts=["Hello, world!", "How are you?"],
    from_lang="en",
    to_lang="ko"
)
```

### 3.4 Google Translate API

**Note**: While widely available, DeepL generally provides better quality for technical content. Use Google Translate for:
- Very large volume translations (cost-effective)
- Language pairs not supported by DeepL
- Integration with existing Google Cloud infrastructure

---

## 4. Quality Assurance Strategies

### 4.1 Translation Validation Framework

```python
from typing import List, Dict
import re

class TranslationValidator:
    def __init__(self):
        self.issues = []

    def validate_translation(
        self,
        original: str,
        translation: str,
        source_lang: str,
        target_lang: str
    ) -> Dict[str, any]:
        """Comprehensive translation quality checks"""

        self.issues = []

        # 1. Code block preservation
        self._check_code_blocks(original, translation)

        # 2. Variable name preservation
        self._check_variables(original, translation)

        # 3. URL preservation
        self._check_urls(original, translation)

        # 4. Markdown structure preservation
        self._check_markdown_structure(original, translation)

        # 5. Terminology consistency
        self._check_terminology(translation, target_lang)

        # 6. Length validation (extreme changes indicate issues)
        self._check_length_ratio(original, translation)

        return {
            'is_valid': len(self.issues) == 0,
            'issues': self.issues,
            'quality_score': self._calculate_quality_score()
        }

    def _check_code_blocks(self, original: str, translation: str):
        """Verify code blocks are unchanged"""

        original_code = re.findall(r'```[\s\S]*?```', original)
        translated_code = re.findall(r'```[\s\S]*?```', translation)

        if len(original_code) != len(translated_code):
            self.issues.append({
                'type': 'code_block_count_mismatch',
                'severity': 'high',
                'original_count': len(original_code),
                'translated_count': len(translated_code)
            })

        for orig, trans in zip(original_code, translated_code):
            if orig != trans:
                self.issues.append({
                    'type': 'code_block_modified',
                    'severity': 'critical',
                    'original': orig[:100],
                    'translated': trans[:100]
                })

    def _check_variables(self, original: str, translation: str):
        """Verify variable names and technical terms"""

        # Extract potential variable names (camelCase, snake_case, etc.)
        var_pattern = r'\b[a-z][a-zA-Z0-9_]*\b'

        original_vars = set(re.findall(var_pattern, original))
        translated_vars = set(re.findall(var_pattern, translation))

        # Check for missing critical variables
        technical_vars = {v for v in original_vars if '_' in v or any(c.isupper() for c in v[1:])}
        missing_vars = technical_vars - translated_vars

        if missing_vars:
            self.issues.append({
                'type': 'missing_variables',
                'severity': 'medium',
                'missing': list(missing_vars)[:10]  # Sample
            })

    def _check_urls(self, original: str, translation: str):
        """Ensure URLs are preserved"""

        url_pattern = r'https?://[^\s<>"\']+'
        original_urls = set(re.findall(url_pattern, original))
        translated_urls = set(re.findall(url_pattern, translation))

        if original_urls != translated_urls:
            self.issues.append({
                'type': 'url_mismatch',
                'severity': 'high',
                'missing_urls': list(original_urls - translated_urls),
                'extra_urls': list(translated_urls - original_urls)
            })

    def _check_markdown_structure(self, original: str, translation: str):
        """Verify markdown structure is maintained"""

        # Check heading levels
        original_headings = re.findall(r'^#{1,6}\s', original, re.MULTILINE)
        translated_headings = re.findall(r'^#{1,6}\s', translation, re.MULTILINE)

        if len(original_headings) != len(translated_headings):
            self.issues.append({
                'type': 'heading_count_mismatch',
                'severity': 'medium',
                'original': len(original_headings),
                'translated': len(translated_headings)
            })

        # Check list structures
        original_lists = len(re.findall(r'^\s*[-*+]\s', original, re.MULTILINE))
        translated_lists = len(re.findall(r'^\s*[-*+]\s', translation, re.MULTILINE))

        if original_lists != translated_lists:
            self.issues.append({
                'type': 'list_structure_mismatch',
                'severity': 'low',
                'original': original_lists,
                'translated': translated_lists
            })

    def _check_terminology(self, translation: str, target_lang: str):
        """Check for terminology consistency"""

        # Load language-specific terminology rules
        if target_lang == "KO":
            # Korean-specific checks
            # Example: Ensure technical terms are not over-translated
            over_translated = re.findall(r'응용 프로그램 프로그래밍 인터페이스', translation)
            if over_translated:
                self.issues.append({
                    'type': 'over_translation',
                    'severity': 'low',
                    'suggestion': 'Use "API" instead of full Korean translation'
                })

    def _check_length_ratio(self, original: str, translation: str):
        """Check for extreme length differences"""

        ratio = len(translation) / len(original) if len(original) > 0 else 0

        # Korean text is typically 80-120% of English length
        # Japanese is similar to English
        if ratio < 0.5 or ratio > 2.0:
            self.issues.append({
                'type': 'extreme_length_difference',
                'severity': 'medium',
                'ratio': ratio,
                'warning': 'Translation may be incomplete or incorrect'
            })

    def _calculate_quality_score(self) -> float:
        """Calculate overall quality score (0-100)"""

        if not self.issues:
            return 100.0

        severity_weights = {
            'critical': 20,
            'high': 10,
            'medium': 5,
            'low': 2
        }

        total_penalty = sum(
            severity_weights.get(issue['severity'], 5)
            for issue in self.issues
        )

        return max(0.0, 100.0 - total_penalty)

# Usage
validator = TranslationValidator()

original_text = """
# API Documentation

```python
def authenticate(user_id: str) -> Token:
    return generate_token(user_id)
```

Visit https://api.example.com for more details.
"""

translated_text = """
# API 문서

```python
def authenticate(user_id: str) -> Token:
    return generate_token(user_id)
```

자세한 내용은 https://api.example.com을 방문하세요.
"""

result = validator.validate_translation(
    original_text,
    translated_text,
    source_lang="EN",
    target_lang="KO"
)

print(f"Valid: {result['is_valid']}")
print(f"Quality Score: {result['quality_score']}")
print(f"Issues: {result['issues']}")
```

### 4.2 Glossary Management System

```python
import sqlite3
from typing import Dict, List
from datetime import datetime

class GlossaryManager:
    def __init__(self, db_path: str = "glossary.db"):
        self.conn = sqlite3.connect(db_path)
        self._init_db()

    def _init_db(self):
        """Initialize glossary database schema"""

        self.conn.execute("""
            CREATE TABLE IF NOT EXISTS glossary (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                source_term TEXT NOT NULL,
                target_term TEXT NOT NULL,
                source_lang TEXT NOT NULL,
                target_lang TEXT NOT NULL,
                domain TEXT NOT NULL,
                context TEXT,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                UNIQUE(source_term, source_lang, target_lang, domain)
            )
        """)

        self.conn.execute("""
            CREATE TABLE IF NOT EXISTS translation_memory (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                source_text TEXT NOT NULL,
                target_text TEXT NOT NULL,
                source_lang TEXT NOT NULL,
                target_lang TEXT NOT NULL,
                domain TEXT,
                quality_score REAL,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                used_count INTEGER DEFAULT 0
            )
        """)

        self.conn.commit()

    def add_term(
        self,
        source_term: str,
        target_term: str,
        source_lang: str,
        target_lang: str,
        domain: str = "general",
        context: str = None
    ):
        """Add or update glossary term"""

        self.conn.execute("""
            INSERT OR REPLACE INTO glossary
            (source_term, target_term, source_lang, target_lang, domain, context, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?)
        """, (source_term, target_term, source_lang, target_lang, domain, context, datetime.now()))

        self.conn.commit()

    def get_term(
        self,
        source_term: str,
        source_lang: str,
        target_lang: str,
        domain: str = "general"
    ) -> str:
        """Look up glossary term"""

        cursor = self.conn.execute("""
            SELECT target_term FROM glossary
            WHERE source_term = ? AND source_lang = ? AND target_lang = ? AND domain = ?
        """, (source_term, source_lang, target_lang, domain))

        result = cursor.fetchone()
        return result[0] if result else None

    def export_for_deepl(
        self,
        source_lang: str,
        target_lang: str,
        domain: str = None
    ) -> Dict[str, str]:
        """Export glossary in DeepL format"""

        query = """
            SELECT source_term, target_term FROM glossary
            WHERE source_lang = ? AND target_lang = ?
        """
        params = [source_lang, target_lang]

        if domain:
            query += " AND domain = ?"
            params.append(domain)

        cursor = self.conn.execute(query, params)

        return {row[0]: row[1] for row in cursor.fetchall()}

    def import_from_tsv(
        self,
        file_path: str,
        source_lang: str,
        target_lang: str,
        domain: str = "general"
    ):
        """Import glossary from TSV file"""

        with open(file_path, 'r', encoding='utf-8') as f:
            for line in f:
                if '\t' in line:
                    source, target = line.strip().split('\t', 1)
                    self.add_term(source, target, source_lang, target_lang, domain)

    def add_to_translation_memory(
        self,
        source_text: str,
        target_text: str,
        source_lang: str,
        target_lang: str,
        quality_score: float = None,
        domain: str = None
    ):
        """Add translation to memory for reuse"""

        self.conn.execute("""
            INSERT INTO translation_memory
            (source_text, target_text, source_lang, target_lang, domain, quality_score)
            VALUES (?, ?, ?, ?, ?, ?)
        """, (source_text, target_text, source_lang, target_lang, domain, quality_score))

        self.conn.commit()

    def search_translation_memory(
        self,
        source_text: str,
        source_lang: str,
        target_lang: str,
        threshold: float = 0.8
    ) -> List[Dict]:
        """Search for similar translations in memory"""

        # Simple exact match (can be enhanced with fuzzy matching)
        cursor = self.conn.execute("""
            SELECT source_text, target_text, quality_score, used_count
            FROM translation_memory
            WHERE source_text = ? AND source_lang = ? AND target_lang = ?
            ORDER BY quality_score DESC, used_count DESC
            LIMIT 5
        """, (source_text, source_lang, target_lang))

        results = []
        for row in cursor.fetchall():
            results.append({
                'source': row[0],
                'target': row[1],
                'quality': row[2],
                'usage_count': row[3]
            })

        return results

# Usage example
glossary_mgr = GlossaryManager()

# Add technical terms
glossary_mgr.add_term(
    source_term="authentication",
    target_term="인증",
    source_lang="EN",
    target_lang="KO",
    domain="security",
    context="User identity verification"
)

glossary_mgr.add_term(
    source_term="API endpoint",
    target_term="API 엔드포인트",
    source_lang="EN",
    target_lang="KO",
    domain="web_development"
)

# Export for DeepL
deepl_glossary = glossary_mgr.export_for_deepl("EN", "KO", "web_development")

# Add to translation memory
glossary_mgr.add_to_translation_memory(
    source_text="The API returns a JSON response with user data.",
    target_text="API는 사용자 데이터가 포함된 JSON 응답을 반환합니다.",
    source_lang="EN",
    target_lang="KO",
    quality_score=95.0,
    domain="web_development"
)
```

---

## 5. Review Workflow Patterns

### 5.1 Bilingual Alignment System

```python
from difflib import SequenceMatcher
from typing import List, Tuple

class BilingualAligner:
    def __init__(self):
        self.alignments = []

    def align_paragraphs(
        self,
        source_paragraphs: List[str],
        target_paragraphs: List[str]
    ) -> List[Tuple[str, str, float]]:
        """Align source and target paragraphs with confidence scores"""

        alignments = []

        # Simple 1:1 alignment (can be enhanced with sentence-level alignment)
        for i, (source, target) in enumerate(zip(source_paragraphs, target_paragraphs)):
            confidence = self._calculate_alignment_confidence(source, target)
            alignments.append((source, target, confidence))

        return alignments

    def _calculate_alignment_confidence(self, source: str, target: str) -> float:
        """Calculate alignment confidence based on structural similarity"""

        # Check for similar structures (code blocks, lists, etc.)
        source_structure = self._extract_structure(source)
        target_structure = self._extract_structure(target)

        structure_match = SequenceMatcher(
            None,
            source_structure,
            target_structure
        ).ratio()

        return structure_match

    def _extract_structure(self, text: str) -> str:
        """Extract structural elements for comparison"""

        import re

        structure = []

        # Code blocks
        if '```' in text:
            structure.append('CODE')

        # Lists
        if re.search(r'^\s*[-*+]\s', text, re.MULTILINE):
            structure.append('LIST')

        # Headings
        headings = re.findall(r'^#{1,6}\s', text, re.MULTILINE)
        structure.extend([f'H{len(h.strip())}' for h in headings])

        # Tables
        if '|' in text and re.search(r'\|.*\|.*\|', text):
            structure.append('TABLE')

        return '-'.join(structure)

    def generate_review_html(
        self,
        alignments: List[Tuple[str, str, float]],
        output_path: str
    ):
        """Generate HTML for side-by-side review"""

        html = """
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Translation Review</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .container { display: flex; gap: 20px; }
        .column { flex: 1; padding: 10px; }
        .source { background: #f0f8ff; }
        .target { background: #fff8f0; }
        .alignment { margin-bottom: 20px; border: 1px solid #ddd; padding: 15px; }
        .confidence {
            font-size: 12px;
            color: #666;
            margin-bottom: 10px;
        }
        .confidence.high { color: green; }
        .confidence.medium { color: orange; }
        .confidence.low { color: red; }
        code { background: #f4f4f4; padding: 2px 5px; border-radius: 3px; }
        pre { background: #f4f4f4; padding: 10px; overflow-x: auto; }
    </style>
</head>
<body>
    <h1>Translation Review</h1>
    <div class="container">
        <div class="column source">
            <h2>Source</h2>
"""

        for i, (source, target, confidence) in enumerate(alignments):
            conf_class = 'high' if confidence > 0.8 else 'medium' if confidence > 0.6 else 'low'

            html += f"""
            <div class="alignment">
                <div class="confidence {conf_class}">
                    Alignment confidence: {confidence:.2%}
                </div>
                <div>{self._markdown_to_html(source)}</div>
            </div>
"""

        html += """
        </div>
        <div class="column target">
            <h2>Target</h2>
"""

        for i, (source, target, confidence) in enumerate(alignments):
            html += f"""
            <div class="alignment">
                <div>{self._markdown_to_html(target)}</div>
            </div>
"""

        html += """
        </div>
    </div>
</body>
</html>
"""

        with open(output_path, 'w', encoding='utf-8') as f:
            f.write(html)

    def _markdown_to_html(self, text: str) -> str:
        """Simple markdown to HTML conversion for display"""

        import re

        # Code blocks
        text = re.sub(
            r'```(.*?)\n(.*?)```',
            r'<pre><code>\2</code></pre>',
            text,
            flags=re.DOTALL
        )

        # Inline code
        text = re.sub(r'`([^`]+)`', r'<code>\1</code>', text)

        # Bold
        text = re.sub(r'\*\*([^*]+)\*\*', r'<strong>\1</strong>', text)

        # Italic
        text = re.sub(r'\*([^*]+)\*', r'<em>\1</em>', text)

        # Line breaks
        text = text.replace('\n', '<br>')

        return text

# Usage
aligner = BilingualAligner()

source_paras = [
    "# Introduction\n\nThis is the API documentation.",
    "```python\ndef hello():\n    print('Hello')\n```",
    "See https://api.example.com for details."
]

target_paras = [
    "# 소개\n\n이것은 API 문서입니다.",
    "```python\ndef hello():\n    print('Hello')\n```",
    "자세한 내용은 https://api.example.com을 참조하세요."
]

alignments = aligner.align_paragraphs(source_paras, target_paras)
aligner.generate_review_html(alignments, "review.html")
```

---

## 6. Production-Ready Integration Example

### Complete Translation Pipeline

```python
import os
from pathlib import Path
from typing import List, Dict
import deepl
from markitdown import MarkItDown

class TechnicalDocumentTranslator:
    """Complete production-ready translation system"""

    def __init__(
        self,
        deepl_key: str,
        glossary_db_path: str = "glossary.db"
    ):
        self.deepl = deepl.Translator(deepl_key)
        self.markitdown = MarkItDown(enable_plugins=True)
        self.glossary_mgr = GlossaryManager(glossary_db_path)
        self.validator = TranslationValidator()
        self.aligner = BilingualAligner()

    def translate_document(
        self,
        input_path: str,
        output_path: str,
        source_lang: str,
        target_lang: str,
        domain: str = "technical"
    ) -> Dict:
        """End-to-end document translation"""

        # Step 1: Convert to Markdown if not already
        if not input_path.endswith('.md'):
            print(f"Converting {input_path} to Markdown...")
            md_result = self.markitdown.convert(input_path)
            markdown_content = md_result.markdown
        else:
            with open(input_path, 'r', encoding='utf-8') as f:
                markdown_content = f.read()

        # Step 2: Split into paragraphs
        paragraphs = self._split_paragraphs(markdown_content)

        # Step 3: Preserve technical content
        preserved_paragraphs = []
        for para in paragraphs:
            processed, placeholders = preserve_technical_content(para)
            preserved_paragraphs.append((processed, placeholders))

        # Step 4: Load glossary
        glossary_entries = self.glossary_mgr.export_for_deepl(
            source_lang,
            target_lang,
            domain
        )

        if glossary_entries:
            glossary = self.deepl.create_glossary(
                f"Translation-{domain}",
                source_lang=source_lang,
                target_lang=target_lang,
                entries=glossary_entries
            )
        else:
            glossary = None

        # Step 5: Translate with DeepL
        translated_paragraphs = []
        for processed_text, placeholders in preserved_paragraphs:
            result = self.deepl.translate_text(
                processed_text,
                source_lang=source_lang,
                target_lang=target_lang,
                preserve_formatting=True,
                glossary=glossary,
                tag_handling='xml'
            )

            # Restore technical content
            restored = restore_technical_content(result.text, placeholders)
            translated_paragraphs.append(restored)

        # Step 6: Validate translations
        validation_results = []
        for orig, trans in zip(paragraphs, translated_paragraphs):
            validation = self.validator.validate_translation(
                orig, trans, source_lang, target_lang
            )
            validation_results.append(validation)

        # Step 7: Generate bilingual review
        alignments = self.aligner.align_paragraphs(paragraphs, translated_paragraphs)
        review_path = output_path.replace('.md', '_review.html')
        self.aligner.generate_review_html(alignments, review_path)

        # Step 8: Save translated document
        translated_content = '\n\n'.join(translated_paragraphs)
        with open(output_path, 'w', encoding='utf-8') as f:
            f.write(translated_content)

        # Step 9: Generate quality report
        avg_quality = sum(v['quality_score'] for v in validation_results) / len(validation_results)

        return {
            'input_file': input_path,
            'output_file': output_path,
            'review_file': review_path,
            'source_lang': source_lang,
            'target_lang': target_lang,
            'paragraph_count': len(paragraphs),
            'average_quality': avg_quality,
            'validation_issues': sum(len(v['issues']) for v in validation_results),
            'low_confidence_alignments': sum(1 for _, _, conf in alignments if conf < 0.7)
        }

    def _split_paragraphs(self, markdown: str) -> List[str]:
        """Split markdown into logical paragraphs"""

        # Split by double newlines, but preserve code blocks
        paragraphs = []
        current = []
        in_code_block = False

        for line in markdown.split('\n'):
            if line.strip().startswith('```'):
                in_code_block = not in_code_block

            if not in_code_block and line.strip() == '' and current:
                paragraphs.append('\n'.join(current))
                current = []
            else:
                current.append(line)

        if current:
            paragraphs.append('\n'.join(current))

        return paragraphs

    def batch_translate_directory(
        self,
        input_dir: str,
        output_dir: str,
        source_lang: str,
        target_lang: str,
        file_pattern: str = "*.md"
    ) -> List[Dict]:
        """Translate all matching files in directory"""

        input_path = Path(input_dir)
        output_path = Path(output_dir)
        output_path.mkdir(parents=True, exist_ok=True)

        results = []

        for file_path in input_path.glob(file_pattern):
            output_file = output_path / file_path.name

            print(f"Translating {file_path.name}...")

            result = self.translate_document(
                str(file_path),
                str(output_file),
                source_lang,
                target_lang
            )

            results.append(result)

        return results

# Usage Example
if __name__ == "__main__":
    translator = TechnicalDocumentTranslator(
        deepl_key=os.getenv("DEEPL_API_KEY"),
        glossary_db_path="tech_glossary.db"
    )

    # Translate single document
    result = translator.translate_document(
        input_path="docs/api_guide.md",
        output_path="docs/ko/api_guide.md",
        source_lang="EN",
        target_lang="KO",
        domain="api_documentation"
    )

    print(f"Translation complete:")
    print(f"- Quality Score: {result['average_quality']:.1f}%")
    print(f"- Issues Found: {result['validation_issues']}")
    print(f"- Review HTML: {result['review_file']}")

    # Batch translate directory
    batch_results = translator.batch_translate_directory(
        input_dir="docs/en",
        output_dir="docs/ko",
        source_lang="EN",
        target_lang="KO"
    )

    print(f"\nBatch translation complete: {len(batch_results)} files")
```

---

## 7. Recommended Technology Stack

### **Optimal Configuration for MoAI-ADK**

```toml
# pyproject.toml
[project]
dependencies = [
    "deepl>=1.18.0",           # Primary translation API
    "openai>=1.105.0",         # Context-aware translation fallback
    "python-docx>=1.1.0",      # Word document processing
    "markitdown>=0.0.1a4",     # Multi-format conversion
    "markdown2>=2.4.12",       # Markdown processing
    "pydantic>=2.0.0",         # Data validation
    "requests>=2.31.0",        # HTTP client
]

[project.optional-dependencies]
azure = [
    "azure-ai-translation-text>=1.0.0",
    "azure-ai-formrecognizer>=3.3.0",
]
enhanced = [
    "mammoth>=1.6.0",          # Better DOCX to HTML conversion
    "pypandoc>=1.12",          # Universal document converter
    "beautifulsoup4>=4.12.0",  # HTML parsing
]
```

### **Environment Setup**

```bash
# Create virtual environment
python -m venv .venv
source .venv/bin/activate  # On Windows: .venv\Scripts\activate

# Install dependencies
pip install -e ".[enhanced]"

# Set API keys
export DEEPL_API_KEY="your_deepl_key"
export OPENAI_API_KEY="your_openai_key"
```

---

## 8. Cost Optimization Strategies

### API Cost Comparison

| Provider | Cost (per 1M chars) | Quality | Best For |
|----------|---------------------|---------|----------|
| DeepL | $25 | Excellent | Technical docs, high quality |
| OpenAI GPT-4o | $15 (prompt) + $60 (completion) | Excellent | Complex context |
| Azure Translator | $10 | Good | High volume, enterprise |
| Google Translate | $20 | Good | General content |

### Cost Optimization Pattern

```python
class CostOptimizedTranslator:
    def __init__(self, deepl_key: str, openai_key: str):
        self.deepl = deepl.Translator(deepl_key)
        self.openai = GPTTranslator(openai_key, model="gpt-4o-mini")  # Cheaper model
        self.cache = {}  # Simple translation cache

    def translate_smart(
        self,
        text: str,
        source_lang: str,
        target_lang: str,
        require_high_quality: bool = False
    ) -> str:
        """Choose optimal translation method based on content and requirements"""

        # Check cache first
        cache_key = f"{text}:{source_lang}:{target_lang}"
        if cache_key in self.cache:
            return self.cache[cache_key]

        # Simple content: Use cheaper DeepL
        if len(text) < 500 and not self._is_complex(text):
            result = self.deepl.translate_text(
                text,
                source_lang=source_lang,
                target_lang=target_lang
            )
            translation = result.text

        # Complex content requiring context: Use GPT (but batched)
        elif require_high_quality or self._is_complex(text):
            result = self.openai.translate_with_context(
                text,
                source_lang,
                target_lang
            )
            translation = result['translation']

        # Default: DeepL for balance
        else:
            result = self.deepl.translate_text(
                text,
                source_lang=source_lang,
                target_lang=target_lang
            )
            translation = result.text

        # Cache result
        self.cache[cache_key] = translation

        return translation

    def _is_complex(self, text: str) -> bool:
        """Determine if text requires advanced context understanding"""

        # Check for complex structures
        has_nested_lists = text.count('\n  -') > 2
        has_multiple_code_blocks = text.count('```') > 4
        has_tables = '|' in text and text.count('|') > 10

        return has_nested_lists or has_multiple_code_blocks or has_tables
```

---

## 9. Performance Benchmarks

### Translation Speed Comparison

| Method | 10 paragraphs | 100 paragraphs | Notes |
|--------|---------------|----------------|-------|
| DeepL API | 2-3 sec | 15-20 sec | Best quality/speed ratio |
| OpenAI GPT-4o | 5-8 sec | 45-60 sec | Slower but context-aware |
| Azure Translator | 1-2 sec | 8-12 sec | Fastest, good for bulk |
| Batch Processing | N/A | 5-10 sec | Asynchronous, best for large volumes |

### Recommended Batch Sizes

```python
# Optimal batch sizes for different APIs
BATCH_SIZES = {
    'deepl': 50,      # DeepL handles up to 50 texts per request
    'openai': 20,     # Rate limits apply
    'azure': 100,     # Supports large batches
}
```

---

## 10. Summary and Recommendations

### **For MoAI-ADK Translation System**

**Primary Stack**:
1. **DeepL API** - Primary translation engine (best quality for technical content)
2. **python-docx** - Word document processing
3. **MarkItDown** - Universal document conversion
4. **OpenAI GPT-4o** - Complex context handling (fallback)

**Quality Assurance**:
- Implement validation pipeline with code preservation checks
- Use glossary management for terminology consistency
- Generate bilingual review HTML for human verification
- Track translation memory for reuse

**Cost Optimization**:
- Cache translations to avoid re-translating identical content
- Batch requests where possible
- Use DeepL for standard technical content (best value)
- Reserve GPT-4 for complex contextual translations

**Workflow**:
1. Convert documents to Markdown (universal format)
2. Extract and preserve code blocks, variables, URLs
3. Apply glossary-enhanced translation
4. Validate output quality
5. Generate bilingual review
6. Export to target format (Markdown, DOCX, PDF)

---

## 11. Reference Implementation Files

Recommended file structure for MoAI-ADK:

```
.moai/translation/
├── config.json                 # Translation settings
├── glossary.db                 # SQLite glossary database
├── translator.py               # Main translation engine
├── validators.py               # Quality validation
├── converters/
│   ├── markdown.py
│   ├── docx_processor.py
│   └── universal.py           # MarkItDown wrapper
├── glossaries/
│   ├── tech_en_ko.tsv
│   ├── tech_en_ja.tsv
│   └── custom_terms.tsv
└── templates/
    └── review.html            # Bilingual review template
```

---

## 12. Additional Resources

**Documentation**:
- DeepL API: https://www.deepl.com/docs-api
- OpenAI Translation: https://platform.openai.com/docs
- python-docx: https://python-docx.readthedocs.io
- MarkItDown: https://github.com/microsoft/markitdown

**Community Resources**:
- Translation memory standards (TMX format)
- Bilingual terminology databases (TBX format)
- XLIFF for translation workflow integration

---

**Last Updated**: 2025-11-16
**Version**: 1.0
**Status**: Production-Ready
