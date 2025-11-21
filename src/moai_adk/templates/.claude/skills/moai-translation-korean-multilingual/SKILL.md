---
name: moai-translation-korean-multilingual
description: Enterprise-grade technical document translation system for Korean↔English↔Japanese
---

## Quick Reference (30 seconds)

# Technical Document Translation System - Korean/English/Japanese

---

## Core Implementation

🎯 Quick Start

#
## Immediate Translation Setup





[\s\S]*?', preserve_code, text)
    preserved_text = re.sub(r'`[^`]+`', preserve_code, preserved_text)
    
    # Translate with GPT-4
    response = client.chat.completions.create(
        model="gpt-4o",
        messages=[
            {"role": "system", "content": f"""Technical translator for {source_lang} to {target_lang}.
            Preserve all technical terms, variable names, and formatting.
            Maintain formal technical documentation style."""},
            {"role": "user", "content": preserved_text}
        ],
        temperature=0.3  # Low temperature for consistency
    )
    
    # Restore code blocks
    translated = response.choices[0].message.content
    for placeholder, code in code_blocks.items():
        translated = translated.replace(placeholder, code)
    
    return translated

# Example usage
result = translate_technical_doc(
    "# API Documentation\n\npython\ndef hello():\n    return 'Hello'\n",
    source_lang="English",
    target_lang="Korean"
)








python
from typing import Dict, List, Optional, Tuple
from openai import OpenAI
import json
import re
from dataclasses import dataclass
from enum import Enum

class TranslationProvider(Enum):
    OPENAI_GPT4 = "openai_gpt4"
    OPENAI_GPT4_MINI = "openai_gpt4_mini"
    DEEPL = "deepl"  # Fallback option
    AZURE = "azure"  # Enterprise option

@dataclass
class TranslationConfig:
    """Configuration for translation engine"""
    provider: TranslationProvider
    source_lang: str
    target_lang: str
    domain: str = "technical"
    preserve_formatting: bool = True
    use_glossary: bool = True
    temperature: float = 0.3
    max_tokens: int = 4000

class TechnicalTranslationEngine:
    """Production-ready technical translation system"""
    
    def __init__(self, openai_key: str, glossary_path: Optional[str] = None):
        self.openai_client = OpenAI(api_key=openai_key)
        self.glossary = self._load_glossary(glossary_path) if glossary_path else {}
        self.placeholders = {}
        self.counter = 0
        
    def translate(
        self,
        text: str,
        config: TranslationConfig,
        context: Optional[str] = None
    ) -> Dict[str, any]:
        """Main translation method with provider selection"""
        
        # Step 1: Pre-process and preserve technical content
        processed_text, preservation_map = self._preserve_technical_content(text)
        
        # Step 2: Apply glossary if enabled
        if config.use_glossary:
            processed_text = self._apply_glossary_preprocessing(
                processed_text, 
                config.source_lang
            )
        
        # Step 3: Translate based on provider
        if config.provider in [TranslationProvider.OPENAI_GPT4, TranslationProvider.OPENAI_GPT4_MINI]:
            translation_result = self._translate_with_openai(
                processed_text, 
                config, 
                context
            )
        else:
            # Fallback or alternative providers
            translation_result = self._translate_with_fallback(
                processed_text,
                config
            )
        
        # Step 4: Restore technical content
        final_translation = self._restore_technical_content(
            translation_result['translation'],
            preservation_map
        )
        
        # Step 5: Validate translation quality
        validation_result = self._validate_translation(
            text,
            final_translation,
            config
        )
        
        return {
            'translation': final_translation,
            'provider': config.provider.value,
            'usage': translation_result.get('usage', {}),
            'validation': validation_result,
            'glossary_applied': len(self.glossary) if config.use_glossary else 0
        }
    
    def _preserve_technical_content(self, text: str) -> Tuple[str, Dict]:
        """Extract and preserve code blocks, URLs, and technical terms"""
        
        preservation_map = {}
        self.counter = 0
        
        # Preservation patterns in priority order
        patterns = [
            (r'[\s\S]*?', 'CODE_BLOCK'),      # Fenced code blocks
            (r'`[^`]+`', 'INLINE_CODE'),            # Inline code
            (r'https?://[^\s]+', 'URL'),            # URLs
            (r'!\[.*?\]\(.*?\)', 'IMAGE'),          # Markdown images
            (r'\$\$[\s\S]*?\$\$', 'MATH_BLOCK'),    # Math blocks
            (r'\$[^$]+\$', 'INLINE_MATH'),          # Inline math
            (r'\b[A-Z][A-Z0-9_]{2,}\b', 'CONSTANT'), # Constants (API_KEY, MAX_SIZE)
            (r'\b[a-z][a-zA-Z0-9]*(?:[A-Z][a-z0-9]*)+\b', 'CAMEL_CASE'), # camelCase
            (r'\b[a-z]+_[a-z_]+\b', 'SNAKE_CASE'),  # snake_case
        ]
        
        processed_text = text
        
        for pattern, content_type in patterns:
            def create_replacer(ctype):
                def replacer(match):
                    placeholder = f"__{ctype}_{self.counter}__"
                    preservation_map[placeholder] = match.group(0)
                    self.counter += 1
                    return placeholder
                return replacer
            
            processed_text = re.sub(pattern, create_replacer(content_type), processed_text)
        
        return processed_text, preservation_map
    
    def _restore_technical_content(self, text: str, preservation_map: Dict) -> str:
        """Restore all preserved technical content"""
        
        restored_text = text
        for placeholder, original_content in preservation_map.items():
            restored_text = restored_text.replace(placeholder, original_content)
        
        return restored_text
    
    def _translate_with_openai(
        self,
        text: str,
        config: TranslationConfig,
        context: Optional[str] = None
    ) -> Dict:
        """Translate using OpenAI GPT-4"""
        
        # Build system prompt with glossary
        system_prompt = self._build_system_prompt(config, context)
        
        # Select model based on config
        model = "gpt-4o" if config.provider == TranslationProvider.OPENAI_GPT4 else "gpt-4o-mini"
        
        response = self.openai_client.chat.completions.create(
            model=model,
            messages=[
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": f"Translate the following text:\n\n{text}"}
            ],
            temperature=config.temperature,
            max_tokens=config.max_tokens
        )
        
        return {
            'translation': response.choices[0].message.content,
            'usage': {
                'prompt_tokens': response.usage.prompt_tokens,
                'completion_tokens': response.usage.completion_tokens,
                'total_tokens': response.usage.total_tokens,
                'estimated_cost': self._calculate_cost(response.usage, model)
            }
        }
    
    def _build_system_prompt(self, config: TranslationConfig, context: Optional[str]) -> str:
        """Build comprehensive system prompt for translation"""
        
        lang_map = {
            'Korean': 'ko', 'English': 'en', 'Japanese': 'ja',
            'ko': 'Korean', 'en': 'English', 'ja': 'Japanese'
        }
        
        source = lang_map.get(config.source_lang, config.source_lang)
        target = lang_map.get(config.target_lang, config.target_lang)
        
        prompt = f"""You are a professional technical translator specializing in {source} to {target} translation.

CRITICAL RULES:
1. Preserve ALL placeholders exactly: __CODE_BLOCK_N__, __INLINE_CODE_N__, __URL_N__, etc.
2. Maintain markdown formatting structure
3. Use formal technical documentation style
4. Apply consistent terminology throughout
5. Do not translate technical terms that should remain in English

DOMAIN: {config.domain}
{f"CONTEXT: {context}" if context else ""}

GLOSSARY:
{json.dumps(self.glossary.get(f"{config.source_lang}_{config.target_lang}", {}), ensure_ascii=False)}

LANGUAGE-SPECIFIC GUIDELINES:
"""
        
        if target == "Korean" or target == "ko":
            prompt += """
- Use formal endings (습니다/합니다)
- Keep common technical terms in English (API, REST, JSON, etc.)
- Use consistent particles (은/는, 이/가, 을/를)
"""
        elif target == "Japanese" or target == "ja":
            prompt += """
- Use formal/polite form (です/ます)
- Maintain appropriate keigo level for technical documentation
- Use katakana for foreign technical terms
"""
        
        return prompt
    
    def _validate_translation(
        self,
        original: str,
        translation: str,
        config: TranslationConfig
    ) -> Dict:
        """Comprehensive translation validation"""
        
        issues = []
        
        # Check code block preservation
        original_code_blocks = len(re.findall(r'[\s\S]*?', original))
        translated_code_blocks = len(re.findall(r'[\s\S]*?', translation))
        
        if original_code_blocks != translated_code_blocks:
            issues.append({
                'type': 'code_block_mismatch',
                'severity': 'critical',
                'details': f"Original: {original_code_blocks}, Translated: {translated_code_blocks}"
            })
        
        # Check URL preservation
        original_urls = set(re.findall(r'https?://[^\s]+', original))
        translated_urls = set(re.findall(r'https?://[^\s]+', translation))
        
        if original_urls != translated_urls:
            issues.append({
                'type': 'url_mismatch',
                'severity': 'high',
                'missing': list(original_urls - translated_urls),
                'extra': list(translated_urls - original_urls)
            })
        
        # Check length ratio (Korean/Japanese typically 80-120% of English)
        length_ratio = len(translation) / len(original) if len(original) > 0 else 0
        
        if length_ratio < 0.5 or length_ratio > 2.0:
            issues.append({
                'type': 'extreme_length_difference',
                'severity': 'medium',
                'ratio': length_ratio
            })
        
        # Calculate quality score
        severity_scores = {'critical': 20, 'high': 10, 'medium': 5, 'low': 2}
        total_penalty = sum(severity_scores.get(issue['severity'], 5) for issue in issues)
        quality_score = max(0.0, 100.0 - total_penalty)
        
        return {
            'is_valid': len(issues) == 0,
            'quality_score': quality_score,
            'issues': issues
        }
    
    def _calculate_cost(self, usage, model: str) -> float:
        """Calculate API cost estimate"""
        
        # Pricing as of 2025-11 (USD per 1K tokens)
        pricing = {
            'gpt-4o': {'prompt': 0.0025, 'completion': 0.01},
            'gpt-4o-mini': {'prompt': 0.00015, 'completion': 0.0006}
        }
        
        model_pricing = pricing.get(model, pricing['gpt-4o-mini'])
        
        prompt_cost = (usage.prompt_tokens / 1000) * model_pricing['prompt']
        completion_cost = (usage.completion_tokens / 1000) * model_pricing['completion']
        
        return round(prompt_cost + completion_cost, 4)
    
    def _load_glossary(self, glossary_path: str) -> Dict:
        """Load glossary from JSON file"""
        
        import json
        try:
            with open(glossary_path, 'r', encoding='utf-8') as f:
                return json.load(f)
        except Exception as e:
            print(f"Warning: Could not load glossary: {e}")
            return {}
    
    def _apply_glossary_preprocessing(self, text: str, source_lang: str) -> str:
        """Apply glossary terms before translation"""
        
        # This is a simplified version - production would use more sophisticated matching
        return text
    
    def _translate_with_fallback(self, text: str, config: TranslationConfig) -> Dict:
        """Fallback translation method"""
        
        # Placeholder for other providers (DeepL, Azure, etc.)
        return {
            'translation': text,  # Return original as fallback
            'usage': {}
        }




python
from pathlib import Path
from typing import Dict, List, Optional
import re

class DocumentProcessor:
    """Process various document formats for translation"""
    
    def __init__(self):
        self.supported_formats = ['.md', '.txt', '.docx', '.pdf']
        
    def process_markdown_file(self, file_path: str) -> Dict:
        """Process Markdown file preserving structure"""
        
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        # Extract metadata if present
        metadata = {}
        if content.startswith(''):
            parts = content.split('', 2)
            if len(parts) >= 3:
                # Parse YAML front matter
                metadata = self._parse_yaml_metadata(parts[1])
                content = parts[2]
        
        # Split into logical sections
        sections = self._split_into_sections(content)
        
        return {
            'format': 'markdown',
            'metadata': metadata,
            'sections': sections,
            'original_content': content
        }
    
    def process_docx_file(self, file_path: str) -> Dict:
        """Process Word document using MarkItDown"""
        
        try:
            from markitdown import MarkItDown
            
            converter = MarkItDown()
            result = converter.convert(file_path)
            
            # Convert to markdown and process
            markdown_content = result.markdown
            sections = self._split_into_sections(markdown_content)
            
            return {
                'format': 'docx',
                'markdown': markdown_content,
                'sections': sections,
                'title': result.title if hasattr(result, 'title') else None
            }
        except ImportError:
            return {
                'error': 'MarkItDown not installed. Install with: pip install markitdown'
            }
    
    def _split_into_sections(self, content: str) -> List[Dict]:
        """Split content into translatable sections"""
        
        sections = []
        current_section = []
        in_code_block = False
        
        for line in content.split('\n'):
            # Track code blocks
            if line.strip().startswith(''):
                in_code_block = not in_code_block
            
            # Split on headers and empty lines (when not in code)
            if not in_code_block and (
                line.strip() == '' or 
                re.match(r'^#{1,6}\s', line)
            ):
                if current_section:
                    sections.append({
                        'content': '\n'.join(current_section),
                        'type': self._classify_section('\n'.join(current_section))
                    })
                    current_section = []
                if line.strip():  # Don't lose headers
                    current_section.append(line)
            else:
                current_section.append(line)
        
        # Add last section
        if current_section:
            sections.append({
                'content': '\n'.join(current_section),
                'type': self._classify_section('\n'.join(current_section))
            })
        
        return sections
    
    def _classify_section(self, content: str) -> str:
        """Classify section type for optimized translation"""
        
        if '' in content:
            return 'code_heavy'
        elif re.search(r'\|.*\|.*\|', content):
            return 'table'
        elif re.search(r'^\s*[-*+]\s', content, re.MULTILINE):
            return 'list'
        else:
            return 'text'
    
    def _parse_yaml_metadata(self, yaml_content: str) -> Dict:
        """Parse YAML front matter"""
        
        metadata = {}
        for line in yaml_content.strip().split('\n'):
            if ':' in line:
                key, value = line.split(':', 1)
                metadata[key.strip()] = value.strip()
        return metadata

    def save_translated_document(
        self,
        translated_sections: List[Dict],
        output_path: str,
        format: str = 'markdown'
    ):
        """Save translated document in specified format"""
        
        output_path = Path(output_path)
        output_path.parent.mkdir(parents=True, exist_ok=True)
        
        if format == 'markdown':
            content = '\n\n'.join(
                section['translated'] for section in translated_sections
            )
            output_path.write_text(content, encoding='utf-8')
        
        elif format == 'html':
            # Convert to HTML with bilingual display
            html_content = self._generate_bilingual_html(translated_sections)
            output_path.write_text(html_content, encoding='utf-8')
    
    def _generate_bilingual_html(self, sections: List[Dict]) -> str:
        """Generate HTML for side-by-side review"""
        
        html = """<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Translation Review</title>
    <style>
        body { font-family: 'Segoe UI', Arial, sans-serif; margin: 20px; }
        .container { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
        .source { background: #f0f8ff; padding: 15px; border-radius: 8px; }
        .target { background: #fff8f0; padding: 15px; border-radius: 8px; }
        .section { margin-bottom: 20px; }
        .meta { font-size: 12px; color: #666; margin-bottom: 10px; }
        pre { background: #f4f4f4; padding: 10px; overflow-x: auto; border-radius: 4px; }
        code { background: #f4f4f4; padding: 2px 5px; border-radius: 3px; }
        .quality-high { color: green; }
        .quality-medium { color: orange; }
        .quality-low { color: red; }
    </style>
</head>
<body>
    <h1>Technical Document Translation Review</h1>
    <div class="container">
        <div class="source">
            <h2>Original</h2>
"""
        
        for section in sections:
            quality_class = 'quality-high' if section.get('quality_score', 0) > 80 else 'quality-medium'
            html += f"""
            <div class="section">
                <div class="meta {quality_class}">
                    Type: {section.get('type', 'text')} | 
                    Quality: {section.get('quality_score', 'N/A')}%
                </div>
                <div>{self._markdown_to_basic_html(section.get('content', ''))}</div>
            </div>
"""
        
        html += """
        </div>
        <div class="target">
            <h2>Translation</h2>
"""
        
        for section in sections:
            html += f"""
            <div class="section">
                <div class="meta">Translated</div>
                <div>{self._markdown_to_basic_html(section.get('translated', ''))}</div>
            </div>
"""
        
        html += """
        </div>
    </div>
</body>
</html>"""
        
        return html
    
    def _markdown_to_basic_html(self, text: str) -> str:
        """Convert markdown to basic HTML for display"""
        
        # Code blocks
        text = re.sub(
            r'(.*?)\n(.*?)',
            r'<pre><code class="\1">\2</code></pre>',
            text,
            flags=re.DOTALL
        )
        
        # Inline code
        text = re.sub(r'`([^`]+)`', r'<code>\1</code>', text)
        
        # Headers
        for i in range(6, 0, -1):
            text = re.sub(f'^{"#" * i}\\s+(.+)$', f'<h{i}>\\1</h{i}>', text, flags=re.MULTILINE)
        
        # Bold and italic
        text = re.sub(r'\*\*([^*]+)\*\*', r'<strong>\1</strong>', text)
        text = re.sub(r'\*([^*]+)\*', r'<em>\1</em>', text)
        
        # Line breaks
        text = text.replace('\n', '<br>')
        
        return text




python
import sqlite3
import json
from typing import Dict, List, Optional
from datetime import datetime

class GlossaryManager:
    """Manage technical terminology for consistent translation"""
    
    def __init__(self, db_path: str = ".moai/translation/glossary.db"):
        self.db_path = db_path
        self.conn = sqlite3.connect(db_path)
        self._init_database()
        
    def _init_database(self):
        """Initialize glossary database schema"""
        
        self.conn.execute("""
            CREATE TABLE IF NOT EXISTS glossary (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                source_term TEXT NOT NULL,
                target_term TEXT NOT NULL,
                source_lang TEXT NOT NULL,
                target_lang TEXT NOT NULL,
                domain TEXT DEFAULT 'general',
                context TEXT,
                usage_count INTEGER DEFAULT 0,
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
                used_count INTEGER DEFAULT 0,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        """)
        
        # Create indexes for performance
        self.conn.execute("""
            CREATE INDEX IF NOT EXISTS idx_glossary_lookup 
            ON glossary(source_term, source_lang, target_lang, domain)
        """)
        
        self.conn.execute("""
            CREATE INDEX IF NOT EXISTS idx_memory_lookup 
            ON translation_memory(source_lang, target_lang, domain)
        """)
        
        self.conn.commit()
        
    def add_term(
        self,
        source_term: str,
        target_term: str,
        source_lang: str,
        target_lang: str,
        domain: str = "technical",
        context: str = None
    ):
        """Add or update a glossary term"""
        
        try:
            self.conn.execute("""
                INSERT OR REPLACE INTO glossary
                (source_term, target_term, source_lang, target_lang, domain, context, updated_at)
                VALUES (?, ?, ?, ?, ?, ?, ?)
            """, (
                source_term, target_term, source_lang, target_lang,
                domain, context, datetime.now()
            ))
            self.conn.commit()
        except Exception as e:
            print(f"Error adding term: {e}")
    
    def bulk_import_glossary(self, file_path: str, format: str = "json"):
        """Import glossary from file"""
        
        if format == "json":
            with open(file_path, 'r', encoding='utf-8') as f:
                glossary_data = json.load(f)
                
            for entry in glossary_data:
                self.add_term(**entry)
                
        elif format == "tsv":
            with open(file_path, 'r', encoding='utf-8') as f:
                for line in f:
                    if '\t' in line:
                        parts = line.strip().split('\t')
                        if len(parts) >= 4:
                            self.add_term(
                                source_term=parts[0],
                                target_term=parts[1],
                                source_lang=parts[2],
                                target_lang=parts[3],
                                domain=parts[4] if len(parts) > 4 else "technical"
                            )
    
    def export_for_openai(
        self,
        source_lang: str,
        target_lang: str,
        domain: str = None
    ) -> Dict[str, str]:
        """Export glossary in format suitable for OpenAI prompt"""
        
        query = """
            SELECT source_term, target_term 
            FROM glossary
            WHERE source_lang = ? AND target_lang = ?
        """
        params = [source_lang, target_lang]
        
        if domain:
            query += " AND domain = ?"
            params.append(domain)
            
        query += " ORDER BY usage_count DESC, source_term"
        
        cursor = self.conn.execute(query, params)
        glossary_dict = {row[0]: row[1] for row in cursor.fetchall()}
        
        # Update usage counts
        if glossary_dict:
            self.conn.execute("""
                UPDATE glossary 
                SET usage_count = usage_count + 1
                WHERE source_lang = ? AND target_lang = ?
            """, (source_lang, target_lang))
            self.conn.commit()
        
        return glossary_dict
    
    def add_to_memory(
        self,
        source_text: str,
        target_text: str,
        source_lang: str,
        target_lang: str,
        quality_score: float = None,
        domain: str = "technical"
    ):
        """Add translation to memory for future reference"""
        
        self.conn.execute("""
            INSERT INTO translation_memory
            (source_text, target_text, source_lang, target_lang, domain, quality_score)
            VALUES (?, ?, ?, ?, ?, ?)
        """, (source_text, target_text, source_lang, target_lang, domain, quality_score))
        self.conn.commit()
    
    def search_memory(
        self,
        source_text: str,
        source_lang: str,
        target_lang: str,
        threshold: float = 0.9
    ) -> Optional[str]:
        """Search translation memory for similar translations"""
        
        # Simple exact match for now
        cursor = self.conn.execute("""
            SELECT target_text, quality_score
            FROM translation_memory
            WHERE source_text = ? AND source_lang = ? AND target_lang = ?
            ORDER BY quality_score DESC, used_count DESC
            LIMIT 1
        """, (source_text, source_lang, target_lang))
        
        result = cursor.fetchone()
        
        if result:
            # Update usage count
            self.conn.execute("""
                UPDATE translation_memory
                SET used_count = used_count + 1
                WHERE source_text = ? AND source_lang = ? AND target_lang = ?
            """, (source_text, source_lang, target_lang))
            self.conn.commit()
            
            return result[0]
        
        return None
    
    def get_statistics(self) -> Dict:
        """Get glossary and memory statistics"""
        
        cursor = self.conn.execute("""
            SELECT 
                (SELECT COUNT(*) FROM glossary) as glossary_terms,
                (SELECT COUNT(*) FROM translation_memory) as memory_entries,
                (SELECT COUNT(DISTINCT domain) FROM glossary) as domains,
                (SELECT SUM(usage_count) FROM glossary) as total_glossary_usage,
                (SELECT SUM(used_count) FROM translation_memory) as total_memory_usage
        """)
        
        row = cursor.fetchone()
        
        return {
            'glossary_terms': row[0],
            'memory_entries': row[1],
            'domains': row[2],
            'total_glossary_usage': row[3] or 0,
            'total_memory_usage': row[4] or 0
        }




python
from typing import Dict, List, Tuple
import re
from difflib import SequenceMatcher

class TranslationValidator:
    """Comprehensive quality validation for translations"""
    
    def __init__(self):
        self.validation_rules = {
            'code_preservation': self._validate_code_blocks,
            'url_preservation': self._validate_urls,
            'variable_preservation': self._validate_variables,
            'markdown_structure': self._validate_markdown,
            'terminology_consistency': self._validate_terminology,
            'length_ratio': self._validate_length_ratio
        }
        
    def validate(
        self,
        original: str,
        translation: str,
        source_lang: str,
        target_lang: str,
        glossary: Dict = None
    ) -> Dict:
        """Run all validation checks"""
        
        issues = []
        passed_checks = []
        
        for check_name, check_func in self.validation_rules.items():
            result = check_func(original, translation, source_lang, target_lang)
            
            if result['passed']:
                passed_checks.append(check_name)
            else:
                issues.extend(result['issues'])
        
        # Check glossary compliance if provided
        if glossary:
            glossary_result = self._validate_glossary_compliance(
                translation, glossary, target_lang
            )
            if not glossary_result['passed']:
                issues.extend(glossary_result['issues'])
        
        # Calculate overall quality score
        quality_score = self._calculate_quality_score(issues)
        
        return {
            'is_valid': len(issues) == 0,
            'quality_score': quality_score,
            'passed_checks': passed_checks,
            'issues': issues,
            'recommendation': self._get_recommendation(quality_score, issues)
        }
    
    def _validate_code_blocks(
        self, 
        original: str, 
        translation: str,
        source_lang: str,
        target_lang: str
    ) -> Dict:
        """Validate code block preservation"""
        
        issues = []
        
        # Extract code blocks
        original_blocks = re.findall(r'[\s\S]*?', original)
        translated_blocks = re.findall(r'[\s\S]*?', translation)
        
        # Check count
        if len(original_blocks) != len(translated_blocks):
            issues.append({
                'type': 'code_block_count_mismatch',
                'severity': 'critical',
                'original_count': len(original_blocks),
                'translated_count': len(translated_blocks)
            })
        
        # Check content preservation
        for i, (orig, trans) in enumerate(zip(original_blocks, translated_blocks)):
            if orig != trans:
                # Extract just the code part for comparison
                orig_code = re.sub(r'^.*?\n', '', orig).rstrip('`')
                trans_code = re.sub(r'^.*?\n', '', trans).rstrip('`')
                
                if orig_code != trans_code:
                    issues.append({
                        'type': 'code_block_modified',
                        'severity': 'critical',
                        'block_index': i,
                        'preview': trans_code[:100] if len(trans_code) > 100 else trans_code
                    })
        
        return {
            'passed': len(issues) == 0,
            'issues': issues
        }
    
    def _validate_urls(
        self,
        original: str,
        translation: str,
        source_lang: str,
        target_lang: str
    ) -> Dict:
        """Validate URL preservation"""
        
        issues = []
        
        url_pattern = r'https?://[^\s<>"\']+'
        original_urls = set(re.findall(url_pattern, original))
        translated_urls = set(re.findall(url_pattern, translation))
        
        missing_urls = original_urls - translated_urls
        extra_urls = translated_urls - original_urls
        
        if missing_urls:
            issues.append({
                'type': 'missing_urls',
                'severity': 'high',
                'urls': list(missing_urls)
            })
        
        if extra_urls:
            issues.append({
                'type': 'extra_urls',
                'severity': 'medium',
                'urls': list(extra_urls)
            })
        
        return {
            'passed': len(issues) == 0,
            'issues': issues
        }
    
    def _validate_variables(
        self,
        original: str,
        translation: str,
        source_lang: str,
        target_lang: str
    ) -> Dict:
        """Validate variable name preservation"""
        
        issues = []
        
        # Pattern for common variable naming conventions
        patterns = [
            (r'\b[a-z][a-zA-Z0-9]*(?:[A-Z][a-z0-9]*)+\b', 'camelCase'),
            (r'\b[a-z]+_[a-z_]+\b', 'snake_case'),
            (r'\b[A-Z][A-Z0-9_]{2,}\b', 'CONSTANT')
        ]
        
        for pattern, var_type in patterns:
            original_vars = set(re.findall(pattern, original))
            translated_vars = set(re.findall(pattern, translation))
            
            missing_vars = original_vars - translated_vars
            
            if missing_vars and len(missing_vars) > len(original_vars) * 0.2:
                issues.append({
                    'type': f'missing_{var_type}_variables',
                    'severity': 'medium',
                    'count': len(missing_vars),
                    'sample': list(missing_vars)[:5]
                })
        
        return {
            'passed': len(issues) == 0,
            'issues': issues
        }
    
    def _validate_markdown(
        self,
        original: str,
        translation: str,
        source_lang: str,
        target_lang: str
    ) -> Dict:
        """Validate markdown structure preservation"""
        
        issues = []
        
        # Check heading structure
        original_headings = re.findall(r'^#{1,6}\s', original, re.MULTILINE)
        translated_headings = re.findall(r'^#{1,6}\s', translation, re.MULTILINE)
        
        if len(original_headings) != len(translated_headings):
            issues.append({
                'type': 'heading_structure_mismatch',
                'severity': 'medium',
                'original': len(original_headings),
                'translated': len(translated_headings)
            })
        
        # Check list structure
        original_lists = len(re.findall(r'^\s*[-*+]\s', original, re.MULTILINE))
        translated_lists = len(re.findall(r'^\s*[-*+]\s', translation, re.MULTILINE))
        
        if abs(original_lists - translated_lists) > original_lists * 0.2:
            issues.append({
                'type': 'list_structure_difference',
                'severity': 'low',
                'original': original_lists,
                'translated': translated_lists
            })
        
        return {
            'passed': len(issues) == 0,
            'issues': issues
        }
    
    def _validate_terminology(
        self,
        original: str,
        translation: str,
        source_lang: str,
        target_lang: str
    ) -> Dict:
        """Check terminology consistency"""
        
        issues = []
        
        # Common technical terms that should not be translated
        preserve_terms = ['API', 'REST', 'JSON', 'XML', 'HTTP', 'HTTPS', 'URL', 
                         'SQL', 'HTML', 'CSS', 'JavaScript', 'Python', 'Java']
        
        for term in preserve_terms:
            if term in original:
                # Check if term is preserved or properly handled
                if term not in translation and term.lower() not in translation.lower():
                    # May be translated, check if it's intentional
                    if target_lang in ['Korean', 'ko', 'Japanese', 'ja']:
                        # These languages might transliterate, which is okay
                        continue
                    else:
                        issues.append({
                            'type': 'technical_term_missing',
                            'severity': 'low',
                            'term': term
                        })
        
        return {
            'passed': len(issues) == 0,
            'issues': issues
        }
    
    def _validate_length_ratio(
        self,
        original: str,
        translation: str,
        source_lang: str,
        target_lang: str
    ) -> Dict:
        """Validate translation length is reasonable"""
        
        issues = []
        
        if len(original) == 0:
            return {'passed': True, 'issues': []}
        
        ratio = len(translation) / len(original)
        
        # Expected ratios based on language pairs
        expected_ratios = {
            ('English', 'Korean'): (0.8, 1.3),
            ('English', 'Japanese'): (0.9, 1.4),
            ('Korean', 'English'): (0.7, 1.2),
            ('Japanese', 'English'): (0.7, 1.1),
            ('Korean', 'Japanese'): (0.9, 1.1),
            ('Japanese', 'Korean'): (0.9, 1.1)
        }
        
        # Normalize language names
        source_normalized = source_lang if source_lang in ['English', 'Korean', 'Japanese'] else source_lang
        target_normalized = target_lang if target_lang in ['English', 'Korean', 'Japanese'] else target_lang
        
        min_ratio, max_ratio = expected_ratios.get(
            (source_normalized, target_normalized),
            (0.5, 2.0)  # Default fallback
        )
        
        if ratio < min_ratio or ratio > max_ratio:
            issues.append({
                'type': 'unusual_length_ratio',
                'severity': 'medium',
                'ratio': round(ratio, 2),
                'expected_range': f"{min_ratio:.1f}-{max_ratio:.1f}"
            })
        
        return {
            'passed': len(issues) == 0,
            'issues': issues
        }
    
    def _validate_glossary_compliance(
        self,
        translation: str,
        glossary: Dict[str, str],
        target_lang: str
    ) -> Dict:
        """Check if glossary terms are used correctly"""
        
        issues = []
        non_compliant_terms = []
        
        for source_term, expected_translation in glossary.items():
            # Simple check - can be enhanced with fuzzy matching
            if expected_translation not in translation:
                # Check if term should appear based on context
                # This is simplified - production would be more sophisticated
                non_compliant_terms.append({
                    'expected': expected_translation,
                    'source': source_term
                })
        
        if non_compliant_terms:
            issues.append({
                'type': 'glossary_non_compliance',
                'severity': 'low',
                'terms': non_compliant_terms[:5]  # Limit to 5 examples
            })
        
        return {
            'passed': len(issues) == 0,
            'issues': issues
        }
    
    def _calculate_quality_score(self, issues: List[Dict]) -> float:
        """Calculate overall quality score (0-100)"""
        
        if not issues:
            return 100.0
        
        severity_weights = {
            'critical': 25,
            'high': 15,
            'medium': 8,
            'low': 3
        }
        
        total_penalty = sum(
            severity_weights.get(issue['severity'], 5)
            for issue in issues
        )
        
        # Cap maximum penalty at 100
        total_penalty = min(total_penalty, 100)
        
        return max(0.0, 100.0 - total_penalty)
    
    def _get_recommendation(self, quality_score: float, issues: List[Dict]) -> str:
        """Get recommendation based on validation results"""
        
        if quality_score >= 95:
            return "Excellent translation quality. Ready for publication."
        elif quality_score >= 85:
            return "Good translation quality. Minor review recommended."
        elif quality_score >= 70:
            return "Acceptable quality. Review and fix identified issues."
        elif quality_score >= 50:
            return "Poor quality. Significant review and corrections needed."
        else:
            return "Critical issues detected. Consider re-translation."

    def generate_validation_report(
        self,
        validation_result: Dict,
        output_path: str = None
    ) -> str:
        """Generate detailed validation report"""
        
        report = f"""# Translation Validation Report


## Summary
- **Overall Quality Score**: {validation_result['quality_score']:.1f}/100
- **Validation Status**: {'✅ PASSED' if validation_result['is_valid'] else '❌ FAILED'}
- **Recommendation**: {validation_result['recommendation']}

---

## Additional Resources

- **Examples**: See [examples.md](examples.md) for practical code samples
- **API Reference**: See [reference.md](reference.md) for detailed documentation
