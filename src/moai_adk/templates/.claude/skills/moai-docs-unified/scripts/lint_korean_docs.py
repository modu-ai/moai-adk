#!/usr/bin/env python3
"""
한국어 문서 마크다운 및 Mermaid 린트 검증 스크립트
"""

import re
import sys
from collections import defaultdict
from pathlib import Path


# 프로젝트 루트 자동 탐지 (pyproject.toml 또는 .git 기준)
def find_project_root(start_path: Path) -> Path:
    current = start_path
    while current != current.parent:
        if (current / "pyproject.toml").exists() or (current / ".git").exists():
            return current
        current = current.parent
    raise RuntimeError("Project root not found")

# 프로젝트 루트 찾기
script_path = Path(__file__).resolve()
project_root = find_project_root(script_path.parent)
sys.path.insert(0, str(project_root))

# 기본 경로 설정
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
        """모든 .md 파일 검증"""
        md_files = sorted(self.docs_path.rglob("*.md"))
        self.file_count = len(md_files)

        print(f"검사 시작: {self.file_count}개 파일")
        print("=" * 80)

        for md_file in md_files:
            self.lint_file(md_file)

        return self.generate_report()

    def lint_file(self, file_path: Path):
        """개별 파일 검증"""
        try:
            content = file_path.read_text(encoding='utf-8')
            rel_path = file_path.relative_to(self.docs_path.parent)

            # 1. 헤더 구조
            self.check_headers(rel_path, content)

            # 2. 코드 블록
            self.check_code_blocks(rel_path, content)

            # 3. Mermaid 다이어그램
            self.check_mermaid(rel_path, content)

            # 4. 링크
            self.check_links(rel_path, content)

            # 5. 리스트
            self.check_lists(rel_path, content)

            # 6. 테이블
            self.check_tables(rel_path, content)

            # 7. 한글 특화
            self.check_korean_specifics(rel_path, content)

            # 8. 공백
            self.check_whitespace(rel_path, content)

        except Exception as e:
            self.errors.append({
                'file': file_path,
                'line': 'N/A',
                'type': 'file',
                'message': f'파일 읽기 오류: {str(e)}'
            })

    def check_headers(self, file_path, content):
        """헤더 구조 검증"""
        lines = content.split('\n')
        prev_level = 0
        h1_count = 0
        h1_line = 0

        for i, line in enumerate(lines, 1):
            if match := re.match(r'^(#{1,6})\s+(.+)$', line):
                level = len(match.group(1))
                _title = match.group(2).strip()

                # H1 중복 확인
                if level == 1:
                    h1_count += 1
                    h1_line = i
                    if h1_count > 1:
                        self.errors.append({
                            'file': file_path,
                            'line': i,
                            'type': 'header',
                            'message': f'H1 중복 (이전: {h1_line}줄, 현재: {i}줄)'
                        })

                # 레벨 건너뛰기 확인
                if prev_level > 0 and level > prev_level + 1:
                    self.warnings.append({
                        'file': file_path,
                        'line': i,
                        'type': 'header',
                        'message': f'헤더 레벨 건너뛰기: H{prev_level} → H{level}'
                    })

                prev_level = level

    def check_code_blocks(self, file_path, content):
        """코드 블록 쌍 검증"""
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
                'message': f'코드 블록 미닫힘 (```{code_lang} 시작이 닫히지 않음)'
            })

    def check_mermaid(self, file_path, content):
        """Mermaid 다이어그램 검증"""
        lines = content.split('\n')

        # Mermaid 블록 찾기
        i = 0
        while i < len(lines):
            if lines[i].strip() == '```mermaid':
                block_start = i
                block_lines = []
                i += 1

                # 블록 끝까지 수집
                while i < len(lines) and lines[i].strip() != '```':
                    block_lines.append(lines[i])
                    i += 1

                if i >= len(lines):
                    self.errors.append({
                        'file': file_path,
                        'line': block_start + 1,
                        'type': 'mermaid',
                        'message': 'Mermaid 블록 미닫힘'
                    })
                else:
                    self.mermaid_blocks += 1
                    block_content = '\n'.join(block_lines)
                    self.validate_mermaid_content(file_path, block_start + 1, block_content)
            i += 1

    def validate_mermaid_content(self, file_path, line_no, content):
        """Mermaid 블록 내용 검증"""
        if not content.strip():
            self.errors.append({
                'file': file_path,
                'line': line_no,
                'type': 'mermaid',
                'message': 'Mermaid 블록이 비어있음'
            })
            return

        first_line = content.strip().split('\n')[0]

        # 지원 다이어그램 타입
        valid_types = [
            'graph', 'sequenceDiagram', 'stateDiagram', 'stateDiagram-v2',
            'classDiagram', 'erDiagram', 'gantt', 'pie', 'flowchart'
        ]

        # 다이어그램 타입 확인
        has_valid_type = any(first_line.strip().startswith(t) for t in valid_types)

        if not has_valid_type and '%%{init:' not in first_line:
            self.warnings.append({
                'file': file_path,
                'line': line_no,
                'type': 'mermaid',
                'message': f'Mermaid 다이어그램 타입 미확인: "{first_line[:50]}"'
            })

        # 기본 노드 정의 패턴 확인
        if 'graph' in first_line or 'flowchart' in first_line:
            nodes = set(re.findall(r'(\w+)[\[\(]', content))
            edges = set(re.findall(r'(\w+)\s*(?:-->|---|\.->|==>)', content))

            # 정의되지 않은 노드 참조 (간단한 검사)
            for edge_src in edges:
                if edge_src and edge_src not in nodes and not re.match(r'^[A-Z]+$', edge_src):
                    self.info.append({
                        'file': file_path,
                        'line': line_no,
                        'type': 'mermaid',
                        'message': f'엣지 시작점이 노드로 정의되지 않음: {edge_src}'
                    })

    def check_links(self, file_path, content):
        """링크 검증"""
        links = re.findall(r'\[([^\]]+)\]\(([^)]+)\)', content)

        for text, url in links:
            # 상대 경로 파일 존재 여부 확인
            if url.startswith('./') or url.startswith('../'):
                # 앵커 제거
                file_url = url.split('#')[0] if '#' in url else url

                if file_url:  # 상대 경로만 확인
                    try:
                        target_path = (file_path.parent / file_url).resolve()
                        # 파일 존재 여부 확인
                        if not target_path.exists() and file_url:
                            self.warnings.append({
                                'file': file_path,
                                'line': 'N/A',
                                'type': 'link',
                                'message': f'깨진 링크: [{text}]({url})'
                            })
                    except Exception:
                        pass

    def check_lists(self, file_path, content):
        """리스트 포맷 검증"""
        lines = content.split('\n')
        list_markers = set()

        for i, line in enumerate(lines, 1):
            # 리스트 마커 추출
            if match := re.match(r'^(\s*)([*\-+])\s+', line):
                indent, marker = match.groups()
                list_markers.add(marker)

                # 들여쓰기 검증 (2 또는 4 스페이스)
                indent_len = len(indent)
                if indent_len > 0 and indent_len % 2 != 0:
                    self.info.append({
                        'file': file_path,
                        'line': i,
                        'type': 'list',
                        'message': f'들여쓰기 홀수: {indent_len}개 스페이스'
                    })

        # 혼합 마커 사용
        if len(list_markers) > 1:
            self.info.append({
                'file': file_path,
                'line': 'N/A',
                'type': 'list',
                'message': f'혼합된 리스트 마커 사용: {", ".join(sorted(list_markers))}'
            })

    def check_tables(self, file_path, content):
        """테이블 포맷 검증"""
        lines = content.split('\n')

        for i in range(len(lines) - 1):
            line = lines[i]

            # 테이블 헤더 패턴
            if '|' in line and line.strip().startswith('|') and '|' in lines[i + 1]:
                # 현재 줄 칼럼 수
                current_cols = len([c for c in line.split('|')[1:-1]])

                # 다음 줄이 구분선인지 확인
                next_line = lines[i + 1]
                if re.match(r'^\|[\s\-:|]+\|$', next_line):
                    sep_cols = len([c for c in next_line.split('|')[1:-1]])

                    if current_cols != sep_cols:
                        self.warnings.append({
                            'file': file_path,
                            'line': i + 1,
                            'type': 'table',
                            'message': f'테이블 칼럼 불일치: {current_cols} vs {sep_cols}'
                        })

    def check_korean_specifics(self, file_path, content):
        """한글 특화 검증"""
        lines = content.split('\n')

        for i, line in enumerate(lines, 1):
            # 전각 공백
            if '\u3000' in line:
                self.warnings.append({
                    'file': file_path,
                    'line': i,
                    'type': 'korean',
                    'message': '전각 공백 (U+3000) 감지'
                })

            # 전각 괄호
            if '（' in line or '）' in line:
                self.info.append({
                    'file': file_path,
                    'line': i,
                    'type': 'korean',
                    'message': '전각 괄호 사용 감지'
                })

            # 전각 쌍따옴표
            if '"' in line or '"' in line:
                self.info.append({
                    'file': file_path,
                    'line': i,
                    'type': 'korean',
                    'message': '전각 쌍따옴표 사용 감지'
                })

    def check_whitespace(self, file_path, content):
        """공백 관련 검증"""
        lines = content.split('\n')

        for i, line in enumerate(lines, 1):
            # 줄 끝 공백
            if line.rstrip() != line:
                self.warnings.append({
                    'file': file_path,
                    'line': i,
                    'type': 'whitespace',
                    'message': f'줄 끝 공백 ({len(line) - len(line.rstrip())}개)'
                })

            # 탭 문자
            if '\t' in line:
                self.warnings.append({
                    'file': file_path,
                    'line': i,
                    'type': 'whitespace',
                    'message': '탭 문자 감지 (스페이스 사용 권장)'
                })

        # 파일 끝 빈 줄 확인
        if content and not content.endswith('\n'):
            self.info.append({
                'file': file_path,
                'line': 'EOF',
                'type': 'whitespace',
                'message': '파일 끝에 줄바꿈 없음'
            })

    def generate_report(self) -> str:
        """검증 리포트 생성"""
        report = []

        # 헤더
        report.append("=" * 80)
        report.append("한국어 문서 마크다운 및 Mermaid 린트 검수 리포트")
        report.append("=" * 80)
        report.append("")

        # 통계
        report.append("## 검수 통계")
        report.append(f"- 검사 파일: {self.file_count}개")
        report.append(f"- Mermaid 블록: {self.mermaid_blocks}개")
        report.append(f"- Errors (Critical): {len(self.errors)}개")
        report.append(f"- Warnings (High): {len(self.warnings)}개")
        report.append(f"- Info (Low): {len(self.info)}개")
        report.append("")

        # 오류별 분류
        error_by_type = defaultdict(list)
        warning_by_type = defaultdict(list)
        info_by_type = defaultdict(list)

        for err in self.errors:
            error_by_type[err['type']].append(err)

        for warn in self.warnings:
            warning_by_type[warn['type']].append(warn)

        for inf in self.info:
            info_by_type[inf['type']].append(inf)

        # ERROR 상세
        if self.errors:
            report.append("## 🔴 Errors (Critical - 즉시 수정 필요)")
            report.append("")

            for error_type in sorted(error_by_type.keys()):
                errors = error_by_type[error_type]
                report.append(f"### {error_type.upper()} ({len(errors)}개)")
                for err in sorted(errors, key=lambda x: str(x['file'])):
                    line_info = f":{err['line']}" if err['line'] != 'N/A' else ""
                    report.append(f"  - {err['file']}{line_info}")
                    report.append(f"    {err['message']}")
                report.append("")

        # WARNING 상세
        if self.warnings:
            report.append("## 🟡 Warnings (High Priority)")
            report.append("")

            for warning_type in sorted(warning_by_type.keys()):
                warnings = warning_by_type[warning_type]
                report.append(f"### {warning_type.upper()} ({len(warnings)}개)")

                # 파일별로 그룹화
                by_file = defaultdict(list)
                for warn in warnings:
                    by_file[warn['file']].append(warn)

                for file_path in sorted(by_file.keys()):
                    report.append(f"  {file_path}:")
                    for warn in by_file[file_path]:
                        line_info = f":{warn['line']}" if warn['line'] != 'N/A' else ""
                        report.append(f"    [{line_info}] {warn['message']}")
                report.append("")

        # INFO 상세
        if self.info:
            report.append("## ℹ️ Info (Low Priority - 선택사항)")
            report.append("")

            for info_type in sorted(info_by_type.keys()):
                infos = info_by_type[info_type]
                report.append(f"### {info_type.upper()} ({len(infos)}개)")

                # 파일별로 그룹화
                by_file = defaultdict(list)
                for inf in infos:
                    by_file[inf['file']].append(inf)

                for file_path in sorted(by_file.keys()):
                    count = len(by_file[file_path])
                    report.append(f"  {file_path} ({count}개 발견)")
                report.append("")

        # 요약
        report.append("=" * 80)
        report.append("## 우선순위별 권장사항")
        report.append("")

        if self.errors:
            report.append(f"**Priority 1 (Critical)**: {len(self.errors)}개 오류 즉시 수정 필요")
            report.append("")

        if self.warnings:
            report.append(f"**Priority 2 (High)**: {len(self.warnings)}개 경고 해결 권장")
            report.append("")

        if self.info:
            report.append(f"**Priority 3 (Low)**: {len(self.info)}개 정보 항목 검토")
            report.append("")

        report.append("=" * 80)

        return "\n".join(report)

# 실행
if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description='한국어 문서 마크다운 린트 검증')
    parser.add_argument('--path', type=str, default=str(DEFAULT_DOCS_PATH),
                       help=f'검사할 문서 경로 (기본값: {DEFAULT_DOCS_PATH})')
    parser.add_argument('--output', type=str, default=str(DEFAULT_REPORT_PATH),
                       help=f'리포트 저장 경로 (기본값: {DEFAULT_REPORT_PATH})')

    args = parser.parse_args()

    linter = KoreanDocsLinter(args.path)
    report = linter.lint_all()
    print(report)

    # 파일에도 저장
    report_path = Path(args.output)
    report_path.parent.mkdir(parents=True, exist_ok=True)
    report_path.write_text(report, encoding='utf-8')
    print(f"\n리포트 저장됨: {report_path}")
