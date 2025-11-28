#!/usr/bin/env python3
"""
Parse README.ko.md into PART A-D structure and analyze content.
Generates JSON structure mapping README sections to Nextra pages.
"""

import json
import re
from pathlib import Path
from typing import Dict, List, Tuple


def extract_section_headers(content: str) -> List[Tuple[str, int, str]]:
    """Extract all ## headers with their line numbers and content."""
    lines = content.split('\n')
    headers = []

    for i, line in enumerate(lines):
        if line.startswith('## '):
            header_text = line.replace('## ', '').strip()
            headers.append((header_text, i, line))

    return headers


def categorize_parts(headers: List[Tuple[str, int, str]]) -> Dict[str, List[Dict]]:
    """Categorize headers into PART A, B, C, D based on content."""

    parts = {
        'PART_A_GETTING_STARTED': [],
        'PART_B_CORE_CONCEPTS': [],
        'PART_C_ADVANCED_LEARNING': [],
        'PART_D_ADVANCED_REFERENCE': []
    }

    # Define patterns for each part
    part_a_keywords = ['소개', '설치', '빠른 시작', 'Introduction', 'Installation', 'Quick Start']
    part_b_keywords = ['SPEC', 'EARS', 'Alfred', '에이전트', '커맨드', 'Workflow', 'Commands']
    part_c_keywords = ['에이전트 가이드', '스킬', '조합 패턴', 'Agent Guide', 'Skills', 'Patterns']
    part_d_keywords = ['고급', '설정', 'MCP', 'FAQ', 'Advanced', 'Configuration', 'Reference']

    for header, line_num, line in headers:
        categorized = False

        # Check PART A
        for keyword in part_a_keywords:
            if keyword.lower() in header.lower():
                parts['PART_A_GETTING_STARTED'].append({
                    'title': header,
                    'line_num': line_num,
                    'raw': line
                })
                categorized = True
                break

        if categorized:
            continue

        # Check PART B
        for keyword in part_b_keywords:
            if keyword.lower() in header.lower():
                parts['PART_B_CORE_CONCEPTS'].append({
                    'title': header,
                    'line_num': line_num,
                    'raw': line
                })
                categorized = True
                break

        if categorized:
            continue

        # Check PART C
        for keyword in part_c_keywords:
            if keyword.lower() in header.lower():
                parts['PART_C_ADVANCED_LEARNING'].append({
                    'title': header,
                    'line_num': line_num,
                    'raw': line
                })
                categorized = True
                break

        if categorized:
            continue

        # Check PART D (default)
        for keyword in part_d_keywords:
            if keyword.lower() in header.lower():
                parts['PART_D_ADVANCED_REFERENCE'].append({
                    'title': header,
                    'line_num': line_num,
                    'raw': line
                })
                categorized = True
                break

        if not categorized:
            # Default to PART D
            parts['PART_D_ADVANCED_REFERENCE'].append({
                'title': header,
                'line_num': line_num,
                'raw': line
            })

    return parts


def generate_mdx_pages_map(parts: Dict[str, List[Dict]]) -> Dict[str, List[Dict]]:
    """Generate mapping of README sections to Nextra pages."""

    pages_map = {
        'getting-started': [],
        'core-concepts': [],
        'workflows': [],
        'advanced': [],
        'reference': []
    }

    # PART A -> getting-started pages
    for header in parts['PART_A_GETTING_STARTED']:
        title = header['title']
        if '소개' in title or 'Introduction' in title:
            pages_map['getting-started'].append({
                'page': 'overview.mdx',
                'title': 'MoAI-ADK 개요',
                'source': title
            })
        elif '설치' in title or 'Installation' in title:
            pages_map['getting-started'].append({
                'page': 'installation.mdx',
                'title': '설치 및 설정',
                'source': title
            })
        elif '빠른 시작' in title or 'Quick Start' in title:
            pages_map['getting-started'].append({
                'page': 'quickstart.mdx',
                'title': '5분 빠른 시작',
                'source': title
            })

    # PART B -> core-concepts pages
    for header in parts['PART_B_CORE_CONCEPTS']:
        title = header['title']
        if 'SPEC' in title and 'EARS' in title:
            pages_map['core-concepts'].append({
                'page': 'spec-format.mdx',
                'title': 'SPEC과 EARS 포맷',
                'source': title
            })
        elif 'Alfred' in title or '에이전트' in title:
            pages_map['core-concepts'].append({
                'page': 'agents.mdx',
                'title': 'Mr.Alfred와 에이전트',
                'source': title
            })
        elif 'Workflow' in title or '워크플로우' in title:
            pages_map['core-concepts'].append({
                'page': 'workflow.mdx',
                'title': 'SPEC-First TDD 워크플로우',
                'source': title
            })
        elif '커맨드' in title or 'Commands' in title:
            pages_map['core-concepts'].append({
                'page': 'commands.mdx',
                'title': '/moai:0-3 핵심 커맨드',
                'source': title
            })

    # PART C -> advanced pages
    for header in parts['PART_C_ADVANCED_LEARNING']:
        title = header['title']
        if '에이전트' in title and '가이드' in title:
            pages_map['advanced'].append({
                'page': 'agents-guide.mdx',
                'title': '26개 에이전트 상세 가이드',
                'source': title
            })
        elif '스킬' in title:
            pages_map['advanced'].append({
                'page': 'skills-library.mdx',
                'title': '22개 스킬 라이브러리',
                'source': title
            })
        elif '조합' in title or 'Patterns' in title:
            pages_map['advanced'].append({
                'page': 'patterns.mdx',
                'title': '고급 조합 패턴',
                'source': title
            })

    # PART D -> reference pages
    for header in parts['PART_D_ADVANCED_REFERENCE']:
        title = header['title']
        if 'TRUST' in title:
            pages_map['reference'].append({
                'page': 'trust5-quality.mdx',
                'title': 'TRUST 5 품질 보증',
                'source': title
            })
        elif '설정' in title or 'Configuration' in title:
            pages_map['reference'].append({
                'page': 'advanced-config.mdx',
                'title': '고급 설정',
                'source': title
            })

    return pages_map


def analyze_readme(readme_path: str) -> Dict:
    """Analyze README and generate structure mapping."""

    with open(readme_path, 'r', encoding='utf-8') as f:
        content = f.read()

    # Extract headers
    headers = extract_section_headers(content)

    # Categorize into PARTS
    parts = categorize_parts(headers)

    # Generate pages mapping
    pages_map = generate_mdx_pages_map(parts)

    # Calculate statistics
    total_lines = len(content.split('\n'))
    code_blocks = len(re.findall(r'```', content))
    tables = len(re.findall(r'\|.*\|', content))
    images = len(re.findall(r'!\[.*\]\(.*\)', content))
    links = len(re.findall(r'\[.*\]\(.*\)', content))

    analysis = {
        'file': str(readme_path),
        'total_lines': total_lines,
        'total_headers': len(headers),
        'statistics': {
            'code_blocks': code_blocks,
            'tables': tables,
            'images': images,
            'links': links,
            'headers_by_level': {
                'h2': len([h for h in headers])
            }
        },
        'parts': {
            'PART_A_GETTING_STARTED': {
                'count': len(parts['PART_A_GETTING_STARTED']),
                'headers': parts['PART_A_GETTING_STARTED']
            },
            'PART_B_CORE_CONCEPTS': {
                'count': len(parts['PART_B_CORE_CONCEPTS']),
                'headers': parts['PART_B_CORE_CONCEPTS']
            },
            'PART_C_ADVANCED_LEARNING': {
                'count': len(parts['PART_C_ADVANCED_LEARNING']),
                'headers': parts['PART_C_ADVANCED_LEARNING']
            },
            'PART_D_ADVANCED_REFERENCE': {
                'count': len(parts['PART_D_ADVANCED_REFERENCE']),
                'headers': parts['PART_D_ADVANCED_REFERENCE']
            }
        },
        'pages_mapping': pages_map,
        'estimated_mdx_pages': sum(len(v) for v in pages_map.values())
    }

    return analysis


def main():
    """Main execution."""
    readme_path = Path(__file__).parent.parent / 'README.ko.md'

    if not readme_path.exists():
        print(f"Error: README.ko.md not found at {readme_path}")
        return

    print(f"Analyzing {readme_path}...")
    analysis = analyze_readme(str(readme_path))

    # Output JSON
    output_path = Path(__file__).parent / 'readme-analysis.json'
    with open(output_path, 'w', encoding='utf-8') as f:
        json.dump(analysis, f, ensure_ascii=False, indent=2)

    print(f"Analysis complete! Results saved to {output_path}")

    # Print summary
    print("\n=== README Analysis Summary ===")
    print(f"Total lines: {analysis['total_lines']}")
    print(f"Total headers: {analysis['total_headers']}")
    print(f"Code blocks: {analysis['statistics']['code_blocks']}")
    print(f"Tables: {analysis['statistics']['tables']}")
    print(f"Images: {analysis['statistics']['images']}")
    print(f"Links: {analysis['statistics']['links']}")

    print("\n=== PART Breakdown ===")
    for part_name, part_data in analysis['parts'].items():
        print(f"{part_name}: {part_data['count']} sections")

    print(f"\n=== Estimated MDX Pages: {analysis['estimated_mdx_pages']} ===")
    for section, pages in analysis['pages_mapping'].items():
        if pages:
            print(f"\n{section}/ ({len(pages)} pages)")
            for page in pages:
                print(f"  - {page['page']}: {page['title']}")


if __name__ == '__main__':
    main()
