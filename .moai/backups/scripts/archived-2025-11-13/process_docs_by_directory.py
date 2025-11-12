#!/usr/bin/env python3
"""
Directory-by-Directory Korean Documentation Processing

Processes one directory at a time with detailed reporting.
"""

import re
import sys
from pathlib import Path
from typing import Dict, List

# Comprehensive emoji pattern
EMOJI_PATTERN = re.compile(
    r'[🎯🚀💻🎨📚🔧⚙️🛡️📈🌐🎓❌✅🔴🟢🟡🟠📋📄🧪📊🔗🎉📁🏗️💡⚡❓🐛✨♻️💼👥'
    r'🎭🎪🎬🎸🎵🎤🎧🎮🎲🃏🎰🎱🎳🎺🎻🪕🪘🔥📝🎁🌟⭐🔑🎈🎀🎊🎏🎐🎑🎒'
    r'🗂️📂📌📍📎📏📐📑📒📓📔📕📖📗📘📙🆘🏷️🔒🆕🔓]'
)

# Material icon mapping (comprehensive)
ICON_MAP = {
    # Core concepts
    '개요': 'dashboard', 'overview': 'dashboard',
    '목적': 'gps_fixed', 'goal': 'flag', 'objective': 'track_changes',
    '특징': 'stars', 'feature': 'new_releases',

    # Installation/Setup
    '설치': 'download', 'install': 'download',
    '설정': 'settings', 'config': 'settings', 'setup': 'build',
    '시작': 'rocket_launch', 'start': 'rocket_launch', '빠른': 'flash_on', 'quick': 'flash_on',

    # Documentation & Guides
    '문서': 'library_books', 'doc': 'library_books',
    '가이드': 'article', 'guide': 'article',
    '튜토리얼': 'school', 'tutorial': 'school',
    '예시': 'code', 'example': 'code', 'sample': 'terminal',
    '참조': 'link', 'reference': 'open_in_new',

    # Development & Code
    '코드': 'code', 'code': 'code',
    '구현': 'developer_board', 'implement': 'developer_board',
    '개발': 'computer', 'develop': 'computer',

    # Testing & Verification
    '테스트': 'science', 'test': 'science',
    '검증': 'verified', 'verify': 'verified', 'check': 'check_circle',

    # Workflow & Process
    '워크플로': 'autorenew', 'workflow': 'autorenew',
    '프로세스': 'sync', 'process': 'sync',
    '단계': 'list', 'step': 'list', 'phase': 'assignment',

    # Project Management
    '프로젝트': 'folder', 'project': 'folder',
    '관리': 'dashboard', 'manage': 'dashboard',

    # Architecture & Design
    '아키텍처': 'architecture', 'architecture': 'architecture',
    '구조': 'schema', 'structure': 'account_tree', 'design': 'schema',

    # Security
    '보안': 'security', 'security': 'security', 'secure': 'shield',
    '인증': 'vpn_lock', 'auth': 'vpn_lock',

    # Performance
    '성능': 'speed', 'performance': 'speed',
    '최적화': 'tune', 'optimize': 'tune', 'perf': 'analytics',

    # Data & Database
    '데이터': 'data_object', 'data': 'data_object',
    '데이터베이스': 'storage', 'database': 'storage', 'db': 'database',

    # Git & Deployment
    'git': 'source_control', 'branch': 'account_tree', 'merge': 'merge_type',
    '배포': 'cloud_upload', 'deploy': 'cloud_upload', 'release': 'rocket',

    # Troubleshooting
    '문제': 'help', 'problem': 'help', 'troubleshoot': 'build',
    '오류': 'error', 'error': 'error',
    '해결': 'build', 'debug': 'bug_report',

    # Skills & Abilities
    '스킬': 'psychology', 'skill': 'psychology',
    '통계': 'analytics', 'statistics': 'analytics', 'metrics': 'bar_chart',

    # Extensions & Plugins
    '확장': 'extension', 'extension': 'extension', 'plugin': 'add',
    'hook': 'webhook',

    # Translation & i18n
    '번역': 'translate', 'translation': 'translate',
    '다국어': 'language', 'i18n': 'language', 'language': 'translate',

    # Feedback & Review
    '피드백': 'feedback', 'feedback': 'feedback',
    '리뷰': 'rate_review', 'review': 'rate_review', 'comment': 'comment',

    # Access & Permissions
    '권한': 'key', 'permission': 'key', 'access': 'person',

    # Status & State
    '완료': 'check_circle', 'success': 'check_circle', 'complete': 'done_all',
    '경고': 'warning', 'warning': 'warning', 'alert': 'priority_high',

    # Lists & Organization
    '목록': 'list', 'list': 'list',
    '체크리스트': 'checklist', 'checklist': 'checklist',
    'todo': 'assignment', 'task': 'assignment_turned_in',

    # Learning & Understanding
    '개념': 'lightbulb', 'concept': 'lightbulb',
    '이해': 'info', 'understand': 'info', 'intro': 'info',
    '고급': 'school', 'advanced': 'school', 'expert': 'library_books',

    # SPEC & Requirements
    'spec': 'description', 'requirement': 'assignment',
    '요구사항': 'assignment', 'ears': 'description',

    # TDD Phases (specific)
    'red': 'cancel', 'green': 'check_circle', 'refactor': 'autorenew',
}


def detect_icon(header_text: str, next_content: str = "") -> str:
    """Detect appropriate icon based on header and content."""
    combined = (header_text + " " + next_content).lower()

    # Try exact keyword matching
    for keyword, icon in ICON_MAP.items():
        if keyword in combined:
            return icon

    # Fallback defaults
    if any(word in combined for word in ['tip', '팁', 'note', '참고', 'hint']):
        return 'lightbulb'
    if any(word in combined for word in ['table', '표', 'summary', '요약']):
        return 'list'

    return 'info'  # Ultimate fallback


def is_in_code_block(lines: List[str], idx: int) -> bool:
    """Check if line is inside code block."""
    count = sum(1 for i in range(idx) if lines[i].strip().startswith('```'))
    return count % 2 == 1


def process_file(file_path: Path) -> Dict:
    """Process single markdown file."""
    stats = {
        'emojis_removed': 0,
        'icons_added': 0,
        'headers_modified': [],
        'errors': []
    }

    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            lines = f.readlines()

        new_lines = []
        i = 0

        while i < len(lines):
            line = lines[i]

            # Skip code blocks
            if is_in_code_block(lines, i):
                new_lines.append(line)
                i += 1
                continue

            # Process headers
            header_match = re.match(r'^(#{1,4})\s+(.*)$', line)
            if header_match:
                level = header_match.group(1)
                content = header_match.group(2)

                # Find and remove emojis
                emojis = EMOJI_PATTERN.findall(content)
                if emojis:
                    clean_content = EMOJI_PATTERN.sub('', content).strip()
                    new_lines.append(f"{level} {clean_content}\n")

                    stats['emojis_removed'] += len(emojis)
                    stats['headers_modified'].append({
                        'line': i + 1,
                        'level': len(level),
                        'original': content,
                        'cleaned': clean_content
                    })

                    # Add Material Icons for H2/H3
                    if len(level) in [2, 3]:
                        # Check if next line already has icon
                        next_idx = i + 1
                        while next_idx < len(lines) and not lines[next_idx].strip():
                            next_idx += 1

                        has_icon = next_idx < len(lines) and 'class="material-icons"' in lines[next_idx]

                        if not has_icon:
                            # Get context
                            context = ""
                            for j in range(next_idx, min(next_idx + 3, len(lines))):
                                if lines[j].strip() and not lines[j].startswith('#'):
                                    context = lines[j]
                                    break

                            icon = detect_icon(clean_content, context)
                            new_lines.append('\n')
                            new_lines.append(f'<span class="material-icons">{icon}</span> **핵심 내용**\n')
                            new_lines.append('\n')
                            stats['icons_added'] += 1

                    i += 1
                    continue
                else:
                    new_lines.append(line)
            else:
                new_lines.append(line)

            i += 1

        # Write back
        with open(file_path, 'w', encoding='utf-8') as f:
            f.writelines(new_lines)

    except Exception as e:
        stats['errors'].append(str(e))

    return stats


def process_directory(base_dir: Path, target_subdir: str):
    """Process all markdown files in a specific subdirectory."""
    target_path = base_dir / target_subdir

    if not target_path.exists():
        print(f"Directory not found: {target_path}")
        return

    md_files = sorted(target_path.rglob('*.md'))

    print("=" * 80)
    print(f"Processing: {target_subdir}/ directory")
    print(f"Found {len(md_files)} markdown files")
    print("=" * 80)

    total_emojis = 0
    total_icons = 0

    for md_file in md_files:
        rel_path = md_file.relative_to(base_dir)
        print(f"\n  {rel_path}")

        stats = process_file(md_file)

        if stats['errors']:
            print(f"    ERROR: {stats['errors']}")
        else:
            print(f"    Emojis removed: {stats['emojis_removed']}")
            print(f"    Icons added: {stats['icons_added']}")

            total_emojis += stats['emojis_removed']
            total_icons += stats['icons_added']

            if stats['headers_modified']:
                print(f"    Headers modified: {len(stats['headers_modified'])}")

    print("\n" + "=" * 80)
    print(f"SUMMARY for {target_subdir}/")
    print("=" * 80)
    print(f"Files processed: {len(md_files)}")
    print(f"Total emojis removed: {total_emojis}")
    print(f"Total Material Icons added: {total_icons}")
    print("")


def main():
    """Main execution."""
    if len(sys.argv) < 2:
        print("Usage: python3 process_docs_by_directory.py <subdirectory>")
        print("\nAvailable subdirectories:")
        print("  reference")
        print("  guides")
        print("  getting-started")
        print("  tutorials")
        print("  contributing")
        print("  advanced")
        print("  troubleshooting")
        print("  root (for index.md, etc.)")
        sys.exit(1)

    base_dir = Path('/Users/goos/MoAI/MoAI-ADK/docs/src/ko')
    target_subdir = sys.argv[1]

    if target_subdir == 'root':
        # Process files in root ko/ directory
        md_files = [f for f in base_dir.glob('*.md')]
        print(f"Processing root files: {len(md_files)}")
        for md_file in md_files:
            print(f"\n  {md_file.name}")
            stats = process_file(md_file)
            print(f"    Emojis removed: {stats['emojis_removed']}")
            print(f"    Icons added: {stats['icons_added']}")
    else:
        process_directory(base_dir, target_subdir)


if __name__ == '__main__':
    main()
