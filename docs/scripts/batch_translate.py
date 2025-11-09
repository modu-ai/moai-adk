#!/usr/bin/env python3
"""
Batch Translation Script

This script reads Korean markdown files and creates translation prompts
for English, Japanese, and Chinese translations.
"""

import os
import json
from pathlib import Path
from typing import List, Dict

LANGUAGES = {
    "en": "English",
    "ja": "Japanese",
    "zh": "Chinese (Simplified)"
}

def get_korean_files(docs_src: Path) -> List[str]:
    """Get all Korean markdown files."""
    ko_dir = docs_src / "ko"
    if not ko_dir.exists():
        return []
    
    files = []
    for md_file in ko_dir.rglob("*.md"):
        rel_path = md_file.relative_to(ko_dir)
        files.append(str(rel_path))
    
    return sorted(files)

def get_translated_files(docs_src: Path, lang: str) -> set:
    """Get all translated files for a language."""
    lang_dir = docs_src / lang
    if not lang_dir.exists():
        return set()
    
    translated = set()
    for md_file in lang_dir.rglob("*.md"):
        rel_path = md_file.relative_to(lang_dir)
        translated.add(str(rel_path))
    
    return translated

def create_translation_prompt(ko_file: Path, target_lang: str, target_file: Path) -> str:
    """Create a translation prompt for Claude."""
    with open(ko_file, 'r', encoding='utf-8') as f:
        ko_content = f.read()
    
    lang_name = LANGUAGES[target_lang]
    
    prompt = f"""Translate the following Korean markdown document to {lang_name}.

**CRITICAL RULES:**
1. Preserve ALL markdown structure (headers, code blocks, links, tables, diagrams)
2. Keep ALL code blocks and technical terms UNCHANGED
3. Maintain the EXACT same file structure and formatting
4. Translate ONLY Korean text content
5. Keep ALL @TAG references unchanged (e.g., @SPEC:AUTH-001)
6. Preserve ALL file paths and URLs
7. Keep ALL emoji and icons as-is
8. Maintain ALL frontmatter (YAML) structure

**Source File:** {ko_file}
**Target Language:** {lang_name}
**Target File:** {target_file}

**Content to Translate:**

{ko_content}

**Instructions:**
- Translate the content above to {lang_name}
- Output ONLY the translated markdown content
- Do NOT include any explanations or comments
- Maintain EXACT markdown formatting
- Preserve ALL code blocks exactly as-is
"""
    
    return prompt

def main():
    """Main execution function."""
    docs_dir = Path(__file__).parent.parent
    docs_src = docs_dir / "src"
    
    print("=" * 60)
    print("BATCH TRANSLATION SETUP")
    print("=" * 60)
    
    korean_files = get_korean_files(docs_src)
    print(f"\nFound {len(korean_files)} Korean files")
    
    # Create translation prompts directory
    prompts_dir = docs_dir / ".moai" / "translation_prompts"
    prompts_dir.mkdir(parents=True, exist_ok=True)
    
    translation_tasks = []
    
    for lang_code, lang_name in LANGUAGES.items():
        translated = get_translated_files(docs_src, lang_code)
        missing = [f for f in korean_files if f not in translated]
        
        print(f"\n{lang_name} ({lang_code}):")
        print(f"  Translated: {len(translated)}/{len(korean_files)}")
        print(f"  Missing: {len(missing)}")
        
        for file_path in missing:
            ko_file = docs_src / "ko" / file_path
            target_file = docs_src / lang_code / file_path
            
            if not ko_file.exists():
                continue
            
            # Create prompt
            prompt = create_translation_prompt(ko_file, lang_code, target_file)
            
            # Save prompt
            prompt_file = prompts_dir / f"{lang_code}_{file_path.replace('/', '_')}.md"
            prompt_file.parent.mkdir(parents=True, exist_ok=True)
            
            with open(prompt_file, 'w', encoding='utf-8') as f:
                f.write(prompt)
            
            translation_tasks.append({
                "lang": lang_code,
                "lang_name": lang_name,
                "source": str(ko_file),
                "target": str(target_file),
                "prompt": str(prompt_file)
            })
    
    # Save task list
    tasks_file = prompts_dir / "translation_tasks.json"
    with open(tasks_file, 'w', encoding='utf-8') as f:
        json.dump(translation_tasks, f, indent=2, ensure_ascii=False)
    
    print(f"\n✅ Created {len(translation_tasks)} translation prompts")
    print(f"📁 Prompts saved to: {prompts_dir}")
    print(f"📋 Task list: {tasks_file}")
    print("\nNext steps:")
    print("1. Review translation prompts")
    print("2. Use Claude to translate files")
    print("3. Save translations to target directories")

if __name__ == "__main__":
    main()




