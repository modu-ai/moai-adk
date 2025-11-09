#!/usr/bin/env python3
"""Remove emojis from markdown files and replace with material icons."""

import re
from pathlib import Path

# Emoji to Material Icon mapping
EMOJI_MAPPING = {
    # Common emojis with their Material Icon equivalents
    "📋": ":clipboard:",
    "📚": ":books:",
    "🚀": ":rocket:",
    "🎯": ":bullseye:",
    "✨": ":sparkles:",
    "🔧": ":wrench:",
    "🌍": ":globe_with_meridians:",
    "🎨": ":artist_palette:",
    "🔐": ":lock:",
    "🎉": ":partying_face:",
    "📦": ":package:",
    "🔥": ":fire:",
    "👥": ":people_holding_hands:",
    "🤝": ":handshake:",
    "📖": ":books:",
    "🎭": ":performing_arts:",
    "✅": ":white_check_mark:",
    "📝": ":memo:",
    "⚡": ":zap:",
    "💡": ":bulb:",
    "🎓": ":graduation_cap:",
    "🔍": ":mag:",
    "📊": ":bar_chart:",
    "🗂️": ":card_index_dividers:",
    "🔗": ":link:",
    "⚙️": ":gear:",
    "🛠️": ":hammer_and_wrench:",
}


def remove_emojis_from_file(file_path):
    """Remove emojis from a markdown file."""
    with open(file_path, "r", encoding="utf-8") as f:
        content = f.read()

    original_content = content

    # Replace emojis
    for emoji, icon in EMOJI_MAPPING.items():
        content = content.replace(emoji, icon)

    # If content changed, write back
    if content != original_content:
        with open(file_path, "w", encoding="utf-8") as f:
            f.write(content)
        return True
    return False


def main():
    """Process all markdown files."""
    src_dir = Path("/Users/goos/MoAI/MoAI-ADK/docs/src")

    # Find all markdown files
    md_files = list(src_dir.rglob("*.md"))
    print(f"Found {len(md_files)} markdown files")

    modified_count = 0
    for md_file in sorted(md_files):
        if remove_emojis_from_file(md_file):
            print(f"  Modified: {md_file.relative_to(src_dir)}")
            modified_count += 1

    print(f"\nTotal modified: {modified_count} files")


if __name__ == "__main__":
    main()
