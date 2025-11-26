#!/usr/bin/env python3
"""
Korean documentation markdown and Mermaid lint validation script
"""

import re
import sys
from collections import defaultdict
from pathlib import Path


# Auto-detect project root (based on pyproject.toml or .git)
def find_project_root(start_path: Path) -> Path:
    current = start_path
    while current != current.parent:
        if (current / "pyproject.toml").exists() or (current / ".git").exists():
            return current
        current = current.parent
    raise RuntimeError("Project root not found")

# Find project root
script_path = Path(__file__).resolve()
project_root = find_project_root(script_path.parent)
sys.path.insert(0, str(project_root))

# Default path configuration
DEFAULT_DOCS_PATH = project_root / "docs" / "src" / "ko"
DEFAULT_REPORT_PATH = project_root / ".moai" / "reports" / "lint_report_ko.txt"

class KoreanDocsLinter:
    def __init__(self, docs_path: str):
        self.docs_path = Path(docs_path)
        self.errors = []
        self.warnings = []
        self.info = []
        self.file_count = 0
        self.mermaid_blocks = 0

    def lint_all(self):
        """Validate all .md files"""
        md_files = sorted(self.docs_path.rglob("*.md"))
        self.file_count = len(md_files)

        print(f"Starting validation: {self.file_count} files")
        print("=" * 80)

        for md_file in md_files:
            self.lint_file(md_file)

        return self.generate_report()

    def lint_file(self, file_path: Path):
        """Validate individual file"""
        try:
            content = file_path.read_text(encoding='utf-8')
            rel_path = file_path.relative_to(self.docs_path.parent)

            # 1. Header structure
            self.check_headers(rel_path, content)

            # 2. Code blocks
            self.check_code_blocks(rel_path, content)

            # 3. Mermaid diagrams
            self.check_mermaid(rel_path, content)

            # 4. Links
            self.check_links(rel_path, content)

            # 5. Lists
            self.check_lists(rel_path, content)

            # 6. Tables
            self.check_tables(rel_path, content)

            # 7. Korean-specific checks
            self.check_korean_specifics(rel_path, content)

            # 8. Whitespace
            self.check_whitespace(rel_path, content)

        except Exception as e:
            self.errors.append({
                'file': file_path,
                'line': 'N/A',
                'type': 'file',
                'message': f'File read error: {str(e)}'
            })

    def check_headers(self, file_path, content):
        """Validate header structure"""
        lines = content.split('\n')
        prev_level = 0
        h1_count = 0
        h1_line = 0

        for i, line in enumerate(lines, 1):
            if match := re.match(r'^(#{1,6})\s+(.+)$', line):
                level = len(match.group(1))
                _title = match.group(2).strip()

                # Check for duplicate H1
                if level == 1:
                    h1_count += 1
                    h1_line = i
                    if h1_count > 1:
                        self.errors.append({
                            'file': file_path,
                            'line': i,
                            'type': 'header',
                            'message': f'Duplicate H1 (previous: line {h1_line}, current: line {i})'
                        })

                # Check for skipped header levels
                if prev_level > 0 and level > prev_level + 1:
                    self.warnings.append({
                        'file': file_path,
                        'line': i,
                        'type': 'header',
                        'message': f'Header level skipped: H{prev_level} → H{level}'
                    })

                prev_level = level

    def check_code_blocks(self, file_path, content):
        """Validate code block pairs"""
        lines = content.split('\n')
        in_code_block = False
        open_line = 0
        code_lang = ""

        for i, line in enumerate(lines, 1):
            if re.match(r'^```(\w+)?', line):
                if not in_code_block:
                    in_code_block = True
                    open_line = i
                    match = re.match(r'^```(\w+)?', line)
                    code_lang = match.group(1) if match.group(1) else "text"
                else:
                    in_code_block = False

        if in_code_block:
            self.errors.append({
                'file': file_path,
                'line': open_line,
                'type': 'code_block',
                'message': f'Unclosed code block (```{code_lang} not closed)'
            })

    def check_mermaid(self, file_path, content):
        """Validate Mermaid diagrams"""
        lines = content.split('\n')

        # Find Mermaid blocks
        i = 0
        while i < len(lines):
            if lines[i].strip() == '```mermaid':
                block_start = i
                block_lines = []
                i += 1

                # Collect until block end
                while i < len(lines) and lines[i].strip() != '```':
                    block_lines.append(lines[i])
                    i += 1

                if i >= len(lines):
                    self.errors.append({
                        'file': file_path,
                        'line': block_start + 1,
                        'type': 'mermaid',
                        'message': 'Unclosed Mermaid block'
                    })
                else:
                    self.mermaid_blocks += 1
                    block_content = '\n'.join(block_lines)
                    self.validate_mermaid_content(file_path, block_start + 1, block_content)
            i += 1

    def validate_mermaid_content(self, file_path, line_no, content):
        """Validate Mermaid block content"""
        if not content.strip():
            self.errors.append({
                'file': file_path,
                'line': line_no,
                'type': 'mermaid',
                'message': 'Empty Mermaid block'
            })
            return

        first_line = content.strip().split('\n')[0]

        # Supported diagram types
        valid_types = [
            'graph', 'sequenceDiagram', 'stateDiagram', 'stateDiagram-v2',
            'classDiagram', 'erDiagram', 'gantt', 'pie', 'flowchart'
        ]

        # Check diagram type
        has_valid_type = any(first_line.strip().startswith(t) for t in valid_types)

        if not has_valid_type and '%%{init:' not in first_line:
            self.warnings.append({
                'file': file_path,
                'line': line_no,
                'type': 'mermaid',
                'message': f'Unrecognized Mermaid diagram type: "{first_line[:50]}"'
            })

        # Check basic node definition patterns
        if 'graph' in first_line or 'flowchart' in first_line:
            nodes = set(re.findall(r'(\w+)[\[\(]', content))
            edges = set(re.findall(r'(\w+)\s*(?:-->|---|\.->|==>)', content))

            # Check for undefined node references (simple check)
            for edge_src in edges:
                if edge_src and edge_src not in nodes and not re.match(r'^[A-Z]+$', edge_src):
                    self.info.append({
                        'file': file_path,
                        'line': line_no,
                        'type': 'mermaid',
                        'message': f'Edge source not defined as node: {edge_src}'
                    })

    def check_links(self, file_path, content):
        """Validate links"""
        links = re.findall(r'\[([^\]]+)\]\(([^)]+)\)', content)

        for text, url in links:
            # Check relative path file existence
            if url.startswith('./') or url.startswith('../'):
                # Remove anchor
                file_url = url.split('#')[0] if '#' in url else url

                if file_url:  # Check relative paths only
                    try:
                        target_path = (file_path.parent / file_url).resolve()
                        # Check file existence
                        if not target_path.exists() and file_url:
                            self.warnings.append({
                                'file': file_path,
                                'line': 'N/A',
                                'type': 'link',
                                'message': f'Broken link: [{text}]({url})'
                            })
                    except Exception:
                        pass

    def check_lists(self, file_path, content):
        """Validate list format"""
        lines = content.split('\n')
        list_markers = set()

        for i, line in enumerate(lines, 1):
            # Extract list marker
            if match := re.match(r'^(\s*)([*\-+])\s+', line):
                indent, marker = match.groups()
                list_markers.add(marker)

                # Validate indentation (2 or 4 spaces)
                indent_len = len(indent)
                if indent_len > 0 and indent_len % 2 != 0:
                    self.info.append({
                        'file': file_path,
                        'line': i,
                        'type': 'list',
                        'message': f'Odd indentation: {indent_len} spaces'
                    })

        # Mixed marker usage
        if len(list_markers) > 1:
            self.info.append({
                'file': file_path,
                'line': 'N/A',
                'type': 'list',
                'message': f'Mixed list markers used: {", ".join(sorted(list_markers))}'
            })

    def check_tables(self, file_path, content):
        """Validate table format"""
        lines = content.split('\n')

        for i in range(len(lines) - 1):
            line = lines[i]

            # Table header pattern
            if '|' in line and line.strip().startswith('|') and '|' in lines[i + 1]:
                # Current line column count
                current_cols = len([c for c in line.split('|')[1:-1]])

                # Check if next line is separator
                next_line = lines[i + 1]
                if re.match(r'^\|[\s\-:|]+\|$', next_line):
                    sep_cols = len([c for c in next_line.split('|')[1:-1]])

                    if current_cols != sep_cols:
                        self.warnings.append({
                            'file': file_path,
                            'line': i + 1,
                            'type': 'table',
                            'message': f'Table column mismatch: {current_cols} vs {sep_cols}'
                        })

    def check_korean_specifics(self, file_path, content):
        """Korean-specific validation"""
        lines = content.split('\n')

        for i, line in enumerate(lines, 1):
            # Full-width space (Korean validation rule - PRESERVE)
            if '\u3000' in line:
                self.warnings.append({
                    'file': file_path,
                    'line': i,
                    'type': 'korean',
                    'message': '전각 공백 (U+3000) detected'
                })

            # Full-width parentheses (Korean validation rule - PRESERVE)
            if '（' in line or '）' in line:
                self.info.append({
                    'file': file_path,
                    'line': i,
                    'type': 'korean',
                    'message': '전각 괄호 usage detected'
                })

            # Full-width double quotes (Korean validation rule - PRESERVE)
            if '"' in line or '"' in line:
                self.info.append({
                    'file': file_path,
                    'line': i,
                    'type': 'korean',
                    'message': '전각 쌍따옴표 usage detected'
                })

    def check_whitespace(self, file_path, content):
        """Whitespace validation"""
        lines = content.split('\n')

        for i, line in enumerate(lines, 1):
            # Trailing whitespace
            if line.rstrip() != line:
                self.warnings.append({
                    'file': file_path,
                    'line': i,
                    'type': 'whitespace',
                    'message': f'Trailing whitespace ({len(line) - len(line.rstrip())} chars)'
                })

            # Tab character
            if '\t' in line:
                self.warnings.append({
                    'file': file_path,
                    'line': i,
                    'type': 'whitespace',
                    'message': 'Tab character detected (spaces recommended)'
                })

        # Check for newline at end of file
        if content and not content.endswith('\n'):
            self.info.append({
                'file': file_path,
                'line': 'EOF',
                'type': 'whitespace',
                'message': 'No newline at end of file'
            })

    def generate_report(self) -> str:
        """Generate validation report"""
        report = []

        # Header
        report.append("=" * 80)
        report.append("Korean Documentation Markdown and Mermaid Lint Validation Report")
        report.append("=" * 80)
        report.append("")

        # Statistics
        report.append("## Validation Statistics")
        report.append(f"- Files checked: {self.file_count}")
        report.append(f"- Mermaid blocks: {self.mermaid_blocks}")
        report.append(f"- Errors (Critical): {len(self.errors)}")
        report.append(f"- Warnings (High): {len(self.warnings)}")
        report.append(f"- Info (Low): {len(self.info)}")
        report.append("")

        # Classify by error type
        error_by_type = defaultdict(list)
        warning_by_type = defaultdict(list)
        info_by_type = defaultdict(list)

        for err in self.errors:
            error_by_type[err['type']].append(err)

        for warn in self.warnings:
            warning_by_type[warn['type']].append(warn)

        for inf in self.info:
            info_by_type[inf['type']].append(inf)

        # ERROR details
        if self.errors:
            report.append("## Errors (Critical - Immediate Fix Required)")
            report.append("")

            for error_type in sorted(error_by_type.keys()):
                errors = error_by_type[error_type]
                report.append(f"### {error_type.upper()} ({len(errors)} items)")
                for err in sorted(errors, key=lambda x: str(x['file'])):
                    line_info = f":{err['line']}" if err['line'] != 'N/A' else ""
                    report.append(f"  - {err['file']}{line_info}")
                    report.append(f"    {err['message']}")
                report.append("")

        # WARNING details
        if self.warnings:
            report.append("## Warnings (High Priority)")
            report.append("")

            for warning_type in sorted(warning_by_type.keys()):
                warnings = warning_by_type[warning_type]
                report.append(f"### {warning_type.upper()} ({len(warnings)} items)")

                # Group by file
                by_file = defaultdict(list)
                for warn in warnings:
                    by_file[warn['file']].append(warn)

                for file_path in sorted(by_file.keys()):
                    report.append(f"  {file_path}:")
                    for warn in by_file[file_path]:
                        line_info = f":{warn['line']}" if warn['line'] != 'N/A' else ""
                        report.append(f"    [{line_info}] {warn['message']}")
                report.append("")

        # INFO details
        if self.info:
            report.append("## Info (Low Priority - Optional)")
            report.append("")

            for info_type in sorted(info_by_type.keys()):
                infos = info_by_type[info_type]
                report.append(f"### {info_type.upper()} ({len(infos)} items)")

                # Group by file
                by_file = defaultdict(list)
                for inf in infos:
                    by_file[inf['file']].append(inf)

                for file_path in sorted(by_file.keys()):
                    count = len(by_file[file_path])
                    report.append(f"  {file_path} ({count} found)")
                report.append("")

        # Summary
        report.append("=" * 80)
        report.append("## Priority-based Recommendations")
        report.append("")

        if self.errors:
            report.append(f"**Priority 1 (Critical)**: {len(self.errors)} errors require immediate fix")
            report.append("")

        if self.warnings:
            report.append(f"**Priority 2 (High)**: {len(self.warnings)} warnings should be resolved")
            report.append("")

        if self.info:
            report.append(f"**Priority 3 (Low)**: {len(self.info)} info items for review")
            report.append("")

        report.append("=" * 80)

        return "\n".join(report)

# Execution
if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description='Korean documentation markdown lint validation')
    parser.add_argument('--path', type=str, default=str(DEFAULT_DOCS_PATH),
                       help=f'Documentation path to check (default: {DEFAULT_DOCS_PATH})')
    parser.add_argument('--output', type=str, default=str(DEFAULT_REPORT_PATH),
                       help=f'Report save path (default: {DEFAULT_REPORT_PATH})')

    args = parser.parse_args()

    linter = KoreanDocsLinter(args.path)
    report = linter.lint_all()
    print(report)

    # Save to file
    report_path = Path(args.output)
    report_path.parent.mkdir(parents=True, exist_ok=True)
    report_path.write_text(report, encoding='utf-8')
    print(f"\nReport saved: {report_path}")
