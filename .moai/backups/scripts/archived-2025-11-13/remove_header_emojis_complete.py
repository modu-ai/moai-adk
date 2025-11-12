#!/usr/bin/env python3
"""
Complete Korean Documentation Emoji Removal and Material Icons Addition

This script:
1. Removes ALL emojis from H1-H4 headers in Korean docs
2. Adds Material Icons to H2/H3 sections (where appropriate)
3. Preserves emojis in code blocks, lists, and tables
4. Generates detailed processing report
"""

import re
import os
from pathlib import Path
from typing import Dict, List, Tuple

# Comprehensive emoji pattern
EMOJI_PATTERN = re.compile(
    r'[🎯🚀💻🎨📚🔧⚙️🛡️📈🌐🎓❌✅🔴🟢🟡🟠📋📄🧪📊🔗🎉📁🏗️💡⚡❓🐛✨♻️💼👥'
    r'🎭🎪🎬🎸🎵🎤🎧🎮🎲🃏🎰🎱🎳🎺🎻🪕🪘🔥📝🎁🌟⭐🔑🎈🎀🎊🎏🎐🎑🎒'
    r'🗂️📂📌📍📎📏📐📑📒📓📔📕📖📗📘📙🆘🏷️🔒🆕🔓]'
)

# Material icon mapping based on section content/keywords
ICON_MAP = {
    # Installation/Download
    'install': 'download',
    'download': 'cloud_download',
    'setup': 'build',

    # Quick Start
    'quick': 'flash_on',
    'start': 'rocket_launch',
    'begin': 'play_arrow',
    '시작': 'rocket_launch',
    '빠른': 'flash_on',

    # Concepts/Understanding
    'concept': 'lightbulb',
    'understand': 'info',
    'overview': 'dashboard',
    'intro': 'info',
    '개념': 'lightbulb',
    '개요': 'dashboard',
    '이해': 'info',

    # Workflow
    'workflow': 'autorenew',
    'process': 'sync',
    'pipeline': 'linear_scale',
    '워크플로': 'autorenew',
    '프로세스': 'sync',

    # Project Management
    'project': 'folder',
    'manage': 'dashboard',
    'config': 'settings',
    '프로젝트': 'folder',
    '관리': 'dashboard',
    '설정': 'settings',

    # SPEC Writing
    'spec': 'description',
    'requirement': 'assignment',
    '요구사항': 'assignment',
    'ears': 'description',

    # Testing
    'test': 'science',
    'verify': 'verified',
    'check': 'check_circle',
    '테스트': 'science',
    '검증': 'verified',

    # Code/Implementation
    'code': 'code',
    'implement': 'developer_board',
    'develop': 'computer',
    '코드': 'code',
    '구현': 'developer_board',

    # Troubleshooting
    'troubleshoot': 'build',
    'debug': 'bug_report',
    'error': 'error',
    'problem': 'help',
    '문제': 'help',
    '오류': 'error',
    '해결': 'build',

    # Deployment
    'deploy': 'cloud_upload',
    'release': 'rocket',
    'publish': 'publish',
    '배포': 'cloud_upload',
    '릴리즈': 'rocket',

    # Git
    'git': 'source_control',
    'branch': 'account_tree',
    'merge': 'merge_type',
    'commit': 'commit',

    # Documentation
    'doc': 'library_books',
    'guide': 'article',
    'tutorial': 'school',
    '문서': 'library_books',
    '가이드': 'article',
    '튜토리얼': 'school',

    # Architecture
    'architecture': 'architecture',
    'design': 'schema',
    'structure': 'account_tree',
    '아키텍처': 'architecture',
    '구조': 'schema',

    # Security
    'security': 'security',
    'secure': 'shield',
    'auth': 'vpn_lock',
    '보안': 'security',
    '인증': 'vpn_lock',

    # Performance
    'performance': 'speed',
    'optimize': 'tune',
    'perf': 'analytics',
    '성능': 'speed',
    '최적화': 'tune',

    # i18n
    'i18n': 'language',
    'translation': 'translate',
    'language': 'translate',
    '번역': 'translate',
    '다국어': 'language',

    # Extensions
    'extension': 'extension',
    'plugin': 'add',
    'hook': 'webhook',
    '확장': 'extension',

    # Feedback
    'feedback': 'feedback',
    'comment': 'comment',
    'review': 'rate_review',
    '피드백': 'feedback',
    '리뷰': 'rate_review',

    # Access/Permissions
    'permission': 'key',
    'access': 'person',
    'user': 'group',
    '권한': 'key',
    '접근': 'person',

    # Data
    'data': 'data_object',
    'database': 'storage',
    'db': 'database',
    '데이터': 'data_object',

    # Links/References
    'link': 'link',
    'reference': 'open_in_new',
    '참조': 'link',

    # Checklist
    'checklist': 'checklist',
    'todo': 'assignment',
    'task': 'assignment_turned_in',

    # Examples
    'example': 'code',
    'sample': 'terminal',
    '예시': 'code',
    '예제': 'terminal',

    # Advanced
    'advanced': 'school',
    'expert': 'library_books',
    '고급': 'school',

    # Steps/Phases
    'step': 'list',
    'phase': 'assignment',
    '단계': 'list',

    # Warning/Alert
    'warning': 'warning',
    'alert': 'priority_high',
    '경고': 'warning',

    # Success
    'success': 'check_circle',
    'complete': 'done_all',
    '완료': 'check_circle',

    # Features
    'feature': 'new_releases',
    'capability': 'stars',
    '기능': 'new_releases',

    # Goals/Objectives
    'goal': 'flag',
    'objective': 'track_changes',
    'target': 'gps_fixed',
    '목표': 'flag',
    '목적': 'gps_fixed',

    # Statistics/Metrics
    'statistics': 'analytics',
    'metrics': 'bar_chart',
    '통계': 'analytics',

    # Skills/Abilities
    'skill': 'psychology',
    'ability': 'stars',
    '스킬': 'psychology',
}


def detect_icon_for_section(header_text: str, next_paragraph: str = "") -> str:
    """Detect appropriate Material Icon based on header and content."""
    combined = (header_text + " " + next_paragraph).lower()

    for keyword, icon in ICON_MAP.items():
        if keyword in combined:
            return icon

    # Default icons by header level context
    if any(word in combined for word in ['list', 'table', 'summary', '목록', '표', '요약']):
        return 'list'
    if any(word in combined for word in ['tip', 'note', 'hint', '팁', '참고']):
        return 'lightbulb'

    return 'info'  # Default fallback


def is_in_code_block(lines: List[str], line_idx: int) -> bool:
    """Check if a line is inside a code block."""
    code_block_count = 0
    for i in range(line_idx):
        if lines[i].strip().startswith('```'):
            code_block_count += 1
    return code_block_count % 2 == 1


def process_file(file_path: Path) -> Dict:
    """Process a single markdown file."""
    stats = {
        'emojis_removed': 0,
        'icons_added': 0,
        'headers_processed': [],
        'errors': []
    }

    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            lines = f.readlines()

        new_lines = []
        i = 0

        while i < len(lines):
            line = lines[i]

            # Skip if in code block
            if is_in_code_block(lines, i):
                new_lines.append(line)
                i += 1
                continue

            # Check for H1-H4 headers with emojis
            header_match = re.match(r'^(#{1,4})\s+(.*)$', line)
            if header_match:
                header_level = header_match.group(1)
                header_content = header_match.group(2)

                # Remove emojis from header
                emojis_found = EMOJI_PATTERN.findall(header_content)
                if emojis_found:
                    clean_header = EMOJI_PATTERN.sub('', header_content).strip()
                    new_line = f"{header_level} {clean_header}\n"
                    new_lines.append(new_line)

                    stats['emojis_removed'] += len(emojis_found)
                    stats['headers_processed'].append({
                        'line': i + 1,
                        'level': len(header_level),
                        'original': header_content,
                        'cleaned': clean_header,
                        'emojis': emojis_found
                    })

                    # Add Material Icon for H2/H3 sections ONLY if next line isn't already an icon
                    if len(header_level) in [2, 3]:
                        # Check if next non-empty line already has Material Icon
                        next_line_idx = i + 1
                        while next_line_idx < len(lines) and not lines[next_line_idx].strip():
                            next_line_idx += 1

                        has_icon = False
                        if next_line_idx < len(lines):
                            if 'class="material-icons"' in lines[next_line_idx]:
                                has_icon = True

                        if not has_icon:
                            # Look ahead for context
                            next_paragraph = ""
                            for j in range(next_line_idx, min(next_line_idx + 5, len(lines))):
                                if lines[j].strip() and not lines[j].strip().startswith('#'):
                                    next_paragraph = lines[j]
                                    break

                            icon = detect_icon_for_section(clean_header, next_paragraph)
                            new_lines.append('\n')
                            new_lines.append(f'<span class="material-icons">{icon}</span> **주요 내용**\n')
                            new_lines.append('\n')
                            stats['icons_added'] += 1

                    i += 1
                    continue
                else:
                    new_lines.append(line)
            else:
                new_lines.append(line)

            i += 1

        # Write back to file
        with open(file_path, 'w', encoding='utf-8') as f:
            f.writelines(new_lines)

    except Exception as e:
        stats['errors'].append(str(e))

    return stats


def main():
    """Main execution function."""
    docs_dir = Path('/Users/goos/MoAI/MoAI-ADK/docs/src/ko')

    if not docs_dir.exists():
        print(f"Error: Directory not found: {docs_dir}")
        return

    # Find all markdown files
    md_files = list(docs_dir.rglob('*.md'))

    print(f"Found {len(md_files)} markdown files in {docs_dir}")
    print("=" * 80)

    total_stats = {
        'files_processed': 0,
        'total_emojis_removed': 0,
        'total_icons_added': 0,
        'file_details': []
    }

    for md_file in sorted(md_files):
        rel_path = md_file.relative_to(docs_dir)
        print(f"\nProcessing: {rel_path}")

        stats = process_file(md_file)

        if stats['errors']:
            print(f"  ERROR: {stats['errors']}")
        else:
            print(f"  Emojis removed: {stats['emojis_removed']}")
            print(f"  Icons added: {stats['icons_added']}")

            total_stats['files_processed'] += 1
            total_stats['total_emojis_removed'] += stats['emojis_removed']
            total_stats['total_icons_added'] += stats['icons_added']
            total_stats['file_details'].append({
                'file': str(rel_path),
                'emojis_removed': stats['emojis_removed'],
                'icons_added': stats['icons_added'],
                'headers': stats['headers_processed']
            })

    print("\n" + "=" * 80)
    print("FINAL SUMMARY")
    print("=" * 80)
    print(f"Files processed: {total_stats['files_processed']}")
    print(f"Total emojis removed: {total_stats['total_emojis_removed']}")
    print(f"Total Material Icons added: {total_stats['total_icons_added']}")

    # Generate detailed report
    report_path = Path('/Users/goos/MoAI/MoAI-ADK/.moai/reports/emoji_removal_complete_report.txt')
    report_path.parent.mkdir(parents=True, exist_ok=True)

    with open(report_path, 'w', encoding='utf-8') as f:
        f.write("=" * 80 + "\n")
        f.write("Korean Documentation Complete Emoji Removal Report\n")
        f.write("=" * 80 + "\n\n")

        f.write(f"Total Files Processed: {total_stats['files_processed']}\n")
        f.write(f"Total Emojis Removed: {total_stats['total_emojis_removed']}\n")
        f.write(f"Total Material Icons Added: {total_stats['total_icons_added']}\n\n")

        f.write("=" * 80 + "\n")
        f.write("File-by-File Details\n")
        f.write("=" * 80 + "\n\n")

        for detail in total_stats['file_details']:
            if detail['emojis_removed'] > 0 or detail['icons_added'] > 0:
                f.write(f"File: {detail['file']}\n")
                f.write(f"  Emojis Removed: {detail['emojis_removed']}\n")
                f.write(f"  Icons Added: {detail['icons_added']}\n")

                if detail['headers']:
                    f.write("  Headers Modified:\n")
                    for header in detail['headers']:
                        f.write(f"    Line {header['line']} (H{header['level']}): {header['original']} -> {header['cleaned']}\n")

                f.write("\n")

    print(f"\nDetailed report saved to: {report_path}")

    # Final validation
    print("\n" + "=" * 80)
    print("Running final validation...")
    print("=" * 80)

    remaining_emojis = 0
    for md_file in md_files:
        with open(md_file, 'r', encoding='utf-8') as f:
            lines = f.readlines()

        for i, line in enumerate(lines):
            if not is_in_code_block(lines, i):
                if re.match(r'^#{1,4}\s+', line):
                    emojis = EMOJI_PATTERN.findall(line)
                    if emojis:
                        remaining_emojis += len(emojis)
                        print(f"  WARNING: {md_file.relative_to(docs_dir)} line {i+1}: {emojis}")

    if remaining_emojis == 0:
        print("\nValidation PASSED: No emojis remaining in headers")
    else:
        print(f"\nValidation WARNING: {remaining_emojis} emojis still found in headers")


if __name__ == '__main__':
    main()
