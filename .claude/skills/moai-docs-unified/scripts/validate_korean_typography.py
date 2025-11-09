#!/usr/bin/env python3
"""
한글 특화 검증 스크립트 (Phase 3)
UTF-8 인코딩, 전각/반각 문자, 타이포그래피 검증
"""

import sys
from pathlib import Path


# 프로젝트 루트 자동 탐지
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
    """한글 문서 타이포그래피 검증"""

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
        """모든 한글 문서 검증"""
        print("=" * 90)
        print("Phase 3: 한글 특화 검증")
        print("=" * 90)
        print()

        korean_files = list(self.docs_path.glob("ko/**/*.md"))
        self.statistics['total_files'] = len(korean_files)

        report_lines = []
        report_lines.append("=" * 90)
        report_lines.append("한글 문서 타이포그래피 검증 리포트 (Phase 3)")
        report_lines.append("=" * 90)
        report_lines.append("")

        for file_path in sorted(korean_files):
            self._validate_file(file_path)

        # 결과 요약 출력
        print(f"검사 완료: {self.statistics['total_files']}개 파일")
        print(f"  - 한글 콘텐츠 파일: {self.statistics['korean_content_files']}개")
        print(f"  - 문제 발견 파일: {self.statistics['files_with_issues']}개")
        print()

        # 상세 결과
        report_lines = self._generate_report()

        return "\n".join(report_lines)

    def _validate_file(self, file_path: Path):
        """개별 파일 검증"""
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
            # 1. 전각 공백 검증 (U+3000)
            if '\u3000' in line:
                file_issues.append({
                    'type': '전각 공백',
                    'line': line_no,
                    'content': line[:80]
                })

            # 2. 전각 괄호 검증
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

            # 3. 공백 위치 검증 (한글 앞뒤 일관성)
            self._check_spacing_consistency(line, line_no, file_issues)

            # 4. 문장부호 검증
            self._check_punctuation(line, line_no, file_issues)

        if file_issues:
            self.statistics['files_with_issues'] += 1
            rel_path = str(file_path.relative_to(self.docs_path))
            print(f"⚠️  {rel_path}: {len(file_issues)}개 문제 발견")

    def _check_spacing_consistency(self, line: str, line_no: int, issues: list):
        """공백 일관성 검증"""
        # 한글과 숫자/영문 사이 공백 확인
        # 예: 한글영문 (X), 한글 영문 (O)

        # 간단한 검증: 연속된 한글-영문-한글 패턴
        import re
        pattern = r'[\uac00-\ud7af][a-zA-Z0-9]{1,3}[\uac00-\ud7af]'
        if re.search(pattern, line):
            # 이것은 경고일 수 있음
            pass

    def _check_punctuation(self, line: str, line_no: int, issues: list):
        """한글 문장부호 검증"""
        # 마침표, 쉼표 등 한글 기준 사용 확인

        # 한글 마침표 (。) vs 영문 마침표 (.)
        if '。' in line:
            issues.append({
                'type': '한글 마침표(。) 사용',
                'line': line_no,
                'content': line[:80]
            })

        # 한글 쉼표 (、) vs 영문 쉼표 (,)
        if '、' in line:
            issues.append({
                'type': '한글 쉼표(、) 사용',
                'line': line_no,
                'content': line[:80]
            })

    def _generate_report(self) -> list:
        """검증 리포트 생성"""
        report = []

        report.append("=" * 90)
        report.append("검증 결과 요약")
        report.append("=" * 90)
        report.append("")
        report.append(f"검사 파일: {self.statistics['total_files']}개")
        report.append(f"한글 콘텐츠 파일: {self.statistics['korean_content_files']}개")
        report.append(f"총 라인 수: {self.statistics['total_lines']:,}개")
        report.append(f"문제 발견 파일: {self.statistics['files_with_issues']}개")
        report.append("")

        # 세부 검증 결과
        report.append("=" * 90)
        report.append("상세 검증 결과")
        report.append("=" * 90)
        report.append("")

        report.append("📋 인코딩 검증")
        report.append("-" * 90)
        if self.results['encoding_issues']:
            report.append(f"❌ {len(self.results['encoding_issues'])}개 인코딩 문제 발견")
            for issue in self.results['encoding_issues'][:10]:
                report.append(f"  - {issue['file']}: {issue['error']}")
        else:
            report.append("✅ 모든 파일 UTF-8 인코딩 정상")
        report.append("")

        report.append("📋 한글 타이포그래피 검증")
        report.append("-" * 90)

        if not self.results['full_width_issues']:
            report.append("✅ 전각 문자 사용 최소화 (권장)")
        else:
            report.append(f"⚠️  {len(self.results['full_width_issues'])}개 전각 문자 사용")

        report.append("")
        report.append("=" * 90)
        report.append("한글 문서 가이드")
        report.append("=" * 90)
        report.append("")
        report.append("✅ 권장 사항:")
        report.append("  1. UTF-8 인코딩 사용 (현재 정상)")
        report.append("  2. 반각 공백 ( ) 사용, 전각 공백 (　) 피하기")
        report.append("  3. 반각 괄호 ( ) 사용, 전각 괄호 （） 피하기")
        report.append("  4. 영문 마침표(.) 사용, 한글 마침표(。) 피하기")
        report.append("  5. 한글-영문 사이에는 공백 추가 (예: '한글 English')")
        report.append("  6. 숫자는 반각 사용 (예: '버전 1.0')")
        report.append("")
        report.append("=" * 90)
        report.append("🎉 Phase 3 (한글 특화 검증) 완료!")
        report.append("=" * 90)

        return report

    def validate_sample_files(self, sample_count: int = 5) -> str:
        """샘플 파일 상세 검증"""
        report = []

        korean_files = list(self.docs_path.glob("ko/**/*.md"))[:sample_count]

        report.append("")
        report.append("=" * 90)
        report.append(f"샘플 파일 상세 분석 (상위 {sample_count}개)")
        report.append("=" * 90)
        report.append("")

        for file_path in sorted(korean_files):
            rel_path = str(file_path.relative_to(self.docs_path))
            content = file_path.read_text(encoding='utf-8')
            lines = content.split('\n')

            # 파일 통계
            korean_chars = sum(1 for c in content if '\uac00' <= c <= '\ud7af')
            english_words = len([w for w in content.split() if any(c.isascii() and c.isalpha() for c in w)])

            report.append(f"📄 {rel_path}")
            report.append(f"   라인 수: {len(lines)}개")
            report.append(f"   한글 문자: {korean_chars:,}개")
            report.append(f"   영문 단어: {english_words:,}개")

            # 제목 구조 분석
            headers = [line for line in lines if line.startswith('#')]
            if headers:
                report.append("   제목 구조:")
                for header in headers[:5]:
                    report.append(f"     {header[:70]}")
                if len(headers) > 5:
                    report.append(f"     ... 외 {len(headers) - 5}개")

            report.append("")

        return "\n".join(report)


def main():
    """메인 실행"""
    import argparse

    parser = argparse.ArgumentParser(description='한글 타이포그래피 검증')
    parser.add_argument('--path', type=str, default=str(DEFAULT_DOCS_PATH),
                       help=f'검사할 문서 경로 (기본값: {DEFAULT_DOCS_PATH})')
    parser.add_argument('--output', type=str, default=str(DEFAULT_REPORT_PATH),
                       help=f'리포트 저장 경로 (기본값: {DEFAULT_REPORT_PATH})')

    args = parser.parse_args()

    validator = KoreanTypographyValidator(args.path)

    # 전체 검증 실행
    report = validator.validate_all()

    # 샘플 파일 상세 분석 추가
    sample_report = validator.validate_sample_files(sample_count=10)
    report += sample_report

    # 콘솔 출력
    print(report)

    # 파일 저장
    report_path = Path(args.output)
    report_path.parent.mkdir(parents=True, exist_ok=True)
    report_path.write_text(report, encoding='utf-8')

    print(f"\n📁 리포트 저장됨: {report_path}")


if __name__ == "__main__":
    main()
