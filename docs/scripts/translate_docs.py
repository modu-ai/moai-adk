#!/usr/bin/env python3
"""
Documentation Translation Script

This script translates Korean documentation files to English, Japanese, and Chinese.
It uses Claude API for high-quality translations while preserving markdown structure.
"""

import os
import json
from pathlib import Path
from typing import Dict, List, Set
import subprocess

# Language mappings
LANGUAGES = {
    "en": "English",
    "ja": "Japanese", 
    "zh": "Chinese (Simplified)"
}

def get_korean_files(docs_src: Path) -> Set[str]:
    """Get all Korean markdown files."""
    korean_files = set()
    ko_dir = docs_src / "ko"
    
    if not ko_dir.exists():
        return korean_files
    
    for md_file in ko_dir.rglob("*.md"):
        rel_path = md_file.relative_to(ko_dir)
        korean_files.add(str(rel_path))
    
    return korean_files

def get_translated_files(docs_src: Path, lang: str) -> Set[str]:
    """Get all translated files for a language."""
    translated = set()
    lang_dir = docs_src / lang
    
    if not lang_dir.exists():
        return translated
    
    for md_file in lang_dir.rglob("*.md"):
        rel_path = md_file.relative_to(lang_dir)
        translated.add(str(rel_path))
    
    return translated

def find_missing_files(docs_src: Path) -> Dict[str, List[str]]:
    """Find missing translation files."""
    korean_files = get_korean_files(docs_src)
    missing = {}
    
    for lang in LANGUAGES.keys():
        translated = get_translated_files(docs_src, lang)
        missing[lang] = sorted(korean_files - translated)
    
    return missing

def ensure_directory_structure(docs_src: Path, lang: str, file_path: str):
    """Ensure directory structure exists for translated file."""
    lang_dir = docs_src / lang
    target_file = lang_dir / file_path
    target_file.parent.mkdir(parents=True, exist_ok=True)

def translate_file_using_claude(ko_file: Path, target_lang: str, target_file: Path) -> bool:
    """
    Translate a file using Claude API.
    
    This is a placeholder - actual implementation would use Claude API.
    For now, we'll create a script that can be run with Claude's assistance.
    """
    # Read Korean content
    with open(ko_file, 'r', encoding='utf-8') as f:
        ko_content = f.read()
    
    # This would normally call Claude API
    # For now, we'll create a prompt file that can be used with Claude
    prompt_file = target_file.parent / f".translate_{target_lang}.md"
    
    prompt = f"""Translate the following Korean markdown document to {LANGUAGES[target_lang]}.

**IMPORTANT RULES:**
1. Preserve all markdown structure (headers, code blocks, links, tables)
2. Keep code blocks and technical terms unchanged
3. Maintain the same file structure and formatting
4. Translate only Korean text content
5. Keep all @TAG references unchanged
6. Preserve all file paths and URLs

**Source File:** {ko_file}
**Target Language:** {LANGUAGES[target_lang]}

**Content to Translate:**

{ko_content}

**Instructions:**
- Translate the content above to {LANGUAGES[target_lang]}
- Output ONLY the translated markdown content
- Do not include any explanations or comments
- Maintain exact markdown formatting
"""
    
    with open(prompt_file, 'w', encoding='utf-8') as f:
        f.write(prompt)
    
    print(f"Created translation prompt: {prompt_file}")
    return True

def main():
    """Main execution function."""
    docs_dir = Path(__file__).parent.parent
    docs_src = docs_dir / "src"
    
    print("=" * 60)
    print("DOCUMENTATION TRANSLATION ANALYSIS")
    print("=" * 60)
    
    # Find missing files
    missing = find_missing_files(docs_src)
    
    total_ko = len(get_korean_files(docs_src))
    print(f"\nTotal Korean files: {total_ko}")
    
    for lang, lang_name in LANGUAGES.items():
        missing_count = len(missing[lang])
        translated_count = total_ko - missing_count
        completion = (translated_count / total_ko * 100) if total_ko > 0 else 0
        
        print(f"\n{lang_name} ({lang}):")
        print(f"  Translated: {translated_count}/{total_ko}")
        print(f"  Missing: {missing_count}")
        print(f"  Completion: {completion:.1f}%")
        
        if missing_count > 0:
            print(f"\n  Top 10 missing files:")
            for file in missing[lang][:10]:
                print(f"    - {file}")
            if missing_count > 10:
                print(f"    ... and {missing_count - 10} more files")
    
    # Generate translation plan
    print("\n" + "=" * 60)
    print("TRANSLATION PLAN")
    print("=" * 60)
    
    plan_file = docs_dir / ".moai" / "reports" / "translation_plan.json"
    plan_file.parent.mkdir(parents=True, exist_ok=True)
    
    plan = {
        "total_korean_files": total_ko,
        "languages": {},
        "missing_files": missing,
        "translation_strategy": "Use Claude API for high-quality translations"
    }
    
    for lang in LANGUAGES.keys():
        plan["languages"][lang] = {
            "translated": total_ko - len(missing[lang]),
            "missing": len(missing[lang]),
            "completion": round((total_ko - len(missing[lang])) / total_ko * 100, 2) if total_ko > 0 else 0
        }
    
    with open(plan_file, 'w', encoding='utf-8') as f:
        json.dump(plan, f, indent=2, ensure_ascii=False)
    
    print(f"\nTranslation plan saved to: {plan_file}")
    print("\nNext steps:")
    print("1. Review the translation plan")
    print("2. Use Claude to translate missing files")
    print("3. Run this script again to verify completion")

if __name__ == "__main__":
    main()




