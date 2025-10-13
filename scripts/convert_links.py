#!/usr/bin/env python3
"""
MkDocs Link Converter

Converts VitePress-style links to MkDocs-compatible links.
@CODE:SPEC-DOCS-003 - Link Conversion
"""
import os
import re
import logging
from pathlib import Path

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def normalize_link(link: str, base_path: str) -> str:
    """
    Convert link to MkDocs compatible format:
    - Remove .vitepress prefix
    - Append .md if no extension
    - Normalize path
    """
    # Remove leading slash and .vitepress prefix
    link = re.sub(r'^/\.vitepress/', '', link)

    # Handle absolute paths
    if link.startswith('/'):
        link = link[1:]

    # Force .md extension if no extension
    if not os.path.splitext(link)[1]:
        link += '.md'

    return link

def convert_markdown_links(content: str) -> str:
    """
    Replace Markdown links in a file.

    Supported conversions:
    - /docs/foo → docs/foo.md
    - /guides/bar.md → guides/bar.md
    - External links preserved
    """
    # Markdown link patterns
    link_patterns = [
        # Inline links: [text](/path)
        r'\[([^\]]+)\]\(([^)]+)\)',
        # Reference links: [text][ref]
        r'\[([^\]]+)\]\[([^\]]+)\]'
    ]

    for pattern in link_patterns:
        def replace_link(match):
            text, link = match.groups()
            # Skip external links
            if link.startswith(('http://', 'https://', '#', 'mailto:')):
                return match.group(0)

            # Normalize link
            normalized_link = normalize_link(link, '')
            return f'[{text}]({normalized_link})'

        content = re.sub(pattern, replace_link, content)

    return content

def convert_files(base_dir: str):
    """
    Convert links in all Markdown files under base_dir.

    Args:
        base_dir (str): Base directory to start conversion
    """
    markdown_files = list(Path(base_dir).rglob('*.md'))
    converted_count = 0

    for file_path in markdown_files:
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()

            new_content = convert_markdown_links(content)

            if new_content != content:
                with open(file_path, 'w', encoding='utf-8') as f:
                    f.write(new_content)
                converted_count += 1
                logger.info(f'Converted: {file_path}')
        except Exception as e:
            logger.error(f'Error processing {file_path}: {e}')

    logger.info(f'Total files converted: {converted_count}')

def main():
    base_dir = os.path.join(os.path.dirname(__file__), '..', 'docs')
    convert_files(base_dir)

if __name__ == '__main__':
    main()