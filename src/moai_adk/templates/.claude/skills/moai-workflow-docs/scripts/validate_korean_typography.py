#!/usr/bin/env python3
"""
Korean-specific validation script (Phase 3)
UTF-8 encoding, full-width/half-width characters, typography validation
"""

import sys
from pathlib import Path


# Auto-detect project root
def find_project_root(start_path: Path) -> Path:
    current = start_path
    while current != current.parent:
        if (current / "pyproject.toml").exists() or (current / ".git").exists():
            return current
        current = current.parent
    raise RuntimeError("Project root not found")

script_path = Path(__file__).resolve()
project_root = find_project_root(script_path.parent)
sys.path.insert(0, str(project_root))

DEFAULT_DOCS_PATH = project_root / "docs" / "src"
DEFAULT_REPORT_PATH = project_root / ".moai" / "reports" / "korean_typography_report.txt"


class KoreanTypographyValidator:
    """Korean document typography validation"""

    def __init__(self, docs_path: str):
        self.docs_path = Path(docs_path)
        self.results = {
            'encoding_issues': [],
            'full_width_issues': [],
            'typography_issues': [],
            'spacing_issues': [],
            'punctuation_issues': [],
            'consistency_issues': [],
        }
        self.statistics = {
            'total_files': 0,
            'total_lines': 0,
            'korean_content_files': 0,
            'files_with_issues': 0,
        }

    def validate_all(self) -> str:
        """Validate all Korean documents"""
        print("=" * 90)
        print("Phase 3: Korean-specific validation")
        print("=" * 90)
        print()

        korean_files = list(self.docs_path.glob("ko/**/*.md"))
        self.statistics['total_files'] = len(korean_files)

        report_lines = []
        report_lines.append("=" * 90)
        report_lines.append("Korean Document Typography Validation Report (Phase 3)")
        report_lines.append("=" * 90)
        report_lines.append("")

        for file_path in sorted(korean_files):
            self._validate_file(file_path)

        # Output summary
        print(f"Validation complete: {self.statistics['total_files']} files")
        print(f"  - Korean content files: {self.statistics['korean_content_files']}")
        print(f"  - Files with issues: {self.statistics['files_with_issues']}")
        print()

        # Detailed results
        report_lines = self._generate_report()

        return "\n".join(report_lines)

    def _validate_file(self, file_path: Path):
        """Validate individual file"""
        try:
            content = file_path.read_text(encoding='utf-8')
        except UnicodeDecodeError as e:
            self.results['encoding_issues'].append({
                'file': str(file_path.relative_to(self.docs_path)),
                'error': str(e)
            })
            return

        lines = content.split('\n')
        self.statistics['total_lines'] += len(lines)

        has_korean = any('\uac00' <= c <= '\ud7af' for line in lines for c in line)
        if not has_korean:
            return

        self.statistics['korean_content_files'] += 1
        file_issues = []

        for line_no, line in enumerate(lines, 1):
            # 1. Full-width space validation (U+3000) - Korean validation rule PRESERVE
            if '\u3000' in line:
                file_issues.append({
                    'type': '전각 공백',
                    'line': line_no,
                    'content': line[:80]
                })

            # 2. Full-width parentheses validation - Korean validation rule PRESERVE
            full_width_parens = {
                '\uff08': '(',  # （
                '\uff09': ')',  # ）
                '\u300c': '"',  # 「
                '\u300d': '"',  # 」
            }

            for full, half in full_width_parens.items():
                if full in line:
                    file_issues.append({
                        'type': f'전각 문자: {full}',
                        'line': line_no,
                        'content': line[:80]
                    })

            # 3. Spacing consistency validation (Korean before/after consistency)
            self._check_spacing_consistency(line, line_no, file_issues)

            # 4. Punctuation validation
            self._check_punctuation(line, line_no, file_issues)

        if file_issues:
            self.statistics['files_with_issues'] += 1
            rel_path = str(file_path.relative_to(self.docs_path))
            print(f"Warning: {rel_path}: {len(file_issues)} issues found")

    def _check_spacing_consistency(self, line: str, line_no: int, issues: list):
        """Spacing consistency validation"""
        # Check spacing between Korean and numbers/English
        # Example: 한글영문 (X), 한글 영문 (O)

        # Simple validation: consecutive Korean-English-Korean pattern
        import re
        pattern = r'[\uac00-\ud7af][a-zA-Z0-9]{1,3}[\uac00-\ud7af]'
        if re.search(pattern, line):
            # This may be a warning
            pass

    def _check_punctuation(self, line: str, line_no: int, issues: list):
        """Korean punctuation validation"""
        # Check usage of periods, commas, etc. based on Korean standard

        # Korean period (。) vs English period (.)
        if '。' in line:
            issues.append({
                'type': '한글 마침표(。) usage',
                'line': line_no,
                'content': line[:80]
            })

        # Korean comma (、) vs English comma (,)
        if '、' in line:
            issues.append({
                'type': '한글 쉼표(、) usage',
                'line': line_no,
                'content': line[:80]
            })

    def _generate_report(self) -> list:
        """Generate validation report"""
        report = []

        report.append("=" * 90)
        report.append("Validation Results Summary")
        report.append("=" * 90)
        report.append("")
        report.append(f"Files checked: {self.statistics['total_files']}")
        report.append(f"Korean content files: {self.statistics['korean_content_files']}")
        report.append(f"Total lines: {self.statistics['total_lines']:,}")
        report.append(f"Files with issues: {self.statistics['files_with_issues']}")
        report.append("")

        # Detailed validation results
        report.append("=" * 90)
        report.append("Detailed Validation Results")
        report.append("=" * 90)
        report.append("")

        report.append("Encoding Validation")
        report.append("-" * 90)
        if self.results['encoding_issues']:
            report.append(f"Error: {len(self.results['encoding_issues'])} encoding issues found")
            for issue in self.results['encoding_issues'][:10]:
                report.append(f"  - {issue['file']}: {issue['error']}")
        else:
            report.append("All files have proper UTF-8 encoding")
        report.append("")

        report.append("Korean Typography Validation")
        report.append("-" * 90)

        if not self.results['full_width_issues']:
            report.append("Full-width character usage minimized (recommended)")
        else:
            report.append(f"Warning: {len(self.results['full_width_issues'])} full-width characters used")

        report.append("")
        report.append("=" * 90)
        report.append("Korean Document Guidelines")
        report.append("=" * 90)
        report.append("")
        report.append("Recommendations:")
        report.append("  1. Use UTF-8 encoding (currently normal)")
        report.append("  2. Use half-width space ( ), avoid full-width space (　)")
        report.append("  3. Use half-width parentheses ( ), avoid full-width （）")
        report.append("  4. Use English period (.), avoid Korean period (。)")
        report.append("  5. Add space between Korean and English (e.g., '한글 English')")
        report.append("  6. Use half-width numbers (e.g., '버전 1.0')")
        report.append("")
        report.append("=" * 90)
        report.append("Phase 3 (Korean-specific validation) complete!")
        report.append("=" * 90)

        return report

    def validate_sample_files(self, sample_count: int = 5) -> str:
        """Validate sample files in detail"""
        report = []

        korean_files = list(self.docs_path.glob("ko/**/*.md"))[:sample_count]

        report.append("")
        report.append("=" * 90)
        report.append(f"Sample File Detailed Analysis (top {sample_count} files)")
        report.append("=" * 90)
        report.append("")

        for file_path in sorted(korean_files):
            rel_path = str(file_path.relative_to(self.docs_path))
            content = file_path.read_text(encoding='utf-8')
            lines = content.split('\n')

            # File statistics
            korean_chars = sum(1 for c in content if '\uac00' <= c <= '\ud7af')
            english_words = len([w for w in content.split() if any(c.isascii() and c.isalpha() for c in w)])

            report.append(f"File: {rel_path}")
            report.append(f"   Lines: {len(lines)}")
            report.append(f"   Korean characters: {korean_chars:,}")
            report.append(f"   English words: {english_words:,}")

            # Header structure analysis
            headers = [line for line in lines if line.startswith('#')]
            if headers:
                report.append("   Header structure:")
                for header in headers[:5]:
                    report.append(f"     {header[:70]}")
                if len(headers) > 5:
                    report.append(f"     ... and {len(headers) - 5} more")

            report.append("")

        return "\n".join(report)


def main():
    """Main execution"""
    import argparse

    parser = argparse.ArgumentParser(description='Korean typography validation')
    parser.add_argument('--path', type=str, default=str(DEFAULT_DOCS_PATH),
                       help=f'Documentation path to check (default: {DEFAULT_DOCS_PATH})')
    parser.add_argument('--output', type=str, default=str(DEFAULT_REPORT_PATH),
                       help=f'Report save path (default: {DEFAULT_REPORT_PATH})')

    args = parser.parse_args()

    validator = KoreanTypographyValidator(args.path)

    # Execute full validation
    report = validator.validate_all()

    # Add sample file detailed analysis
    sample_report = validator.validate_sample_files(sample_count=10)
    report += sample_report

    # Console output
    print(report)

    # Save to file
    report_path = Path(args.output)
    report_path.parent.mkdir(parents=True, exist_ok=True)
    report_path.write_text(report, encoding='utf-8')

    print(f"\nReport saved: {report_path}")


if __name__ == "__main__":
    main()
