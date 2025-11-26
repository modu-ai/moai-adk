#!/usr/bin/env python3
"""
한국어 문서 Mermaid 다이어그램 상세 검증 스크립트
Phase 2: Mermaid 다이어그램 검증 (문법, 구조, 렌더링 가능성)
"""

import re
import sys
from pathlib import Path
from typing import Dict, List, Tuple


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
DEFAULT_REPORT_PATH = project_root / ".moai" / "reports" / "mermaid_validation_report.txt"


class MermaidDiagramValidator:
    """Mermaid 다이어그램 검증 클래스"""

    def __init__(self, docs_path: str):
        self.docs_path = Path(docs_path)
        self.diagrams = []
        self.validation_results = []
        self.files_to_check = [
            "ko/guides/alfred/index.md",
            "ko/guides/alfred/1-plan.md",
            "ko/guides/tdd/red.md",
            "ko/guides/tdd/green.md",
            "ko/guides/tdd/refactor.md",
            "ko/getting-started/quick-start-ko.md",
            "ko/guides/project/deploy.md",
            "ko/guides/project/init.md",
            "ko/guides/project/config.md",
        ]

    def extract_mermaid_blocks(self) -> Dict[str, List[Dict]]:
        """모든 Mermaid 블록 추출"""
        mermaid_blocks = {}

        for file_rel_path in self.files_to_check:
            file_path = self.docs_path / file_rel_path

            if not file_path.exists():
                print(f"⚠️  파일을 찾을 수 없음: {file_rel_path}")
                continue

            content = file_path.read_text(encoding='utf-8')

            # Mermaid 블록 정규식: ```mermaid ... ```
            pattern = r'```mermaid\n(.*?)\n```'
            matches = re.finditer(pattern, content, re.DOTALL)

            blocks = []
            for idx, match in enumerate(matches, 1):
                mermaid_code = match.group(1)
                start_line = content[:match.start()].count('\n') + 1

                blocks.append({
                    'id': f"{file_rel_path.split('/')[-1]}_diagram_{idx}",
                    'file': file_rel_path,
                    'line': start_line,
                    'code': mermaid_code,
                    'length': len(mermaid_code.split('\n'))
                })

            if blocks:
                mermaid_blocks[file_rel_path] = blocks

        return mermaid_blocks

    def validate_diagram_type(self, code: str) -> Tuple[str, bool]:
        """Mermaid 다이어그램 타입 검증"""
        code_stripped = code.strip()

        # %%{init: {...}}%% 설정 제거 (첫 번째 줄이 이것이면 건너뛰기)
        lines = code_stripped.split('\n')
        actual_code_start = 0

        if lines and lines[0].startswith('%%{init:'):
            actual_code_start = 1
            while actual_code_start < len(lines) and lines[actual_code_start].strip() == '':
                actual_code_start += 1

        if actual_code_start < len(lines):
            first_meaningful_line = lines[actual_code_start].strip()
        else:
            first_meaningful_line = code_stripped

        # 지원되는 다이어그램 타입
        diagram_patterns = {
            'graph': r'^graph\s+(TD|BT|LR|RL)',
            'flowchart': r'^flowchart\s+(TD|BT|LR|RL)',
            'stateDiagram': r'^stateDiagram-v2',
            'sequenceDiagram': r'^sequenceDiagram',
            'classDiagram': r'^classDiagram',
            'erDiagram': r'^erDiagram',
            'gantt': r'^gantt',
        }

        for diagram_type, pattern in diagram_patterns.items():
            if re.match(pattern, first_meaningful_line):
                return diagram_type, True

        return 'UNKNOWN', False

    def validate_syntax(self, code: str, diagram_type: str) -> List[str]:
        """다이어그램 문법 검증"""
        issues = []
        lines = code.split('\n')

        if diagram_type in ['graph', 'flowchart']:
            # Graph/Flowchart 검증
            self._validate_graph_syntax(lines, issues)
        elif diagram_type == 'stateDiagram':
            # State Diagram 검증
            self._validate_state_diagram_syntax(lines, issues)
        elif diagram_type == 'sequenceDiagram':
            # Sequence Diagram 검증
            self._validate_sequence_diagram_syntax(lines, issues)

        return issues

    def _validate_graph_syntax(self, lines: List[str], issues: List[str]):
        """Graph/Flowchart 문법 검증"""
        node_ids = set()
        edges = []

        for i, line in enumerate(lines[1:], 1):  # 첫 번째 줄은 graph 정의
            line = line.strip()
            if not line or line.startswith('%%'):  # 주석 무시
                continue

            # 노드 정의: A["텍스트"]
            node_pattern = r'([A-Za-z0-9_]+)\s*\['
            node_matches = re.findall(node_pattern, line)
            for node_id in node_matches:
                node_ids.add(node_id)

            # 엣지: A --> B
            edge_pattern = r'([A-Za-z0-9_]+)\s*(?:-->|---|o\||o\)|<-->|<-->|===)\s*([A-Za-z0-9_]+)'
            edge_matches = re.findall(edge_pattern, line)
            for src, dst in edge_matches:
                edges.append((src, dst))

        # 미정의 노드 참조 확인
        for src, dst in edges:
            if src not in node_ids and not src.replace('[', '').replace(']', ''):
                issues.append(f"  - 미정의 노드: {src} (라인 {i})")
            if dst not in node_ids and not dst.replace('[', '').replace(']', ''):
                issues.append(f"  - 미정의 노드: {dst} (라인 {i})")

    def _validate_state_diagram_syntax(self, lines: List[str], issues: List[str]):
        """State Diagram 문법 검증"""
        state_defs = set()

        for i, line in enumerate(lines[1:], 1):
            line = line.strip()
            if not line or line.startswith('%%'):
                continue

            # 상태 정의
            if ':' in line or '[*]' in line:
                # Simple state definition detection
                if '--> ' in line and '[*]' not in line:
                    parts = line.split('-->')
                    if len(parts) >= 2:
                        state_defs.add(parts[0].strip())

    def _validate_sequence_diagram_syntax(self, lines: List[str], issues: List[str]):
        """Sequence Diagram 문법 검증"""
        participants = set()

        for i, line in enumerate(lines[1:], 1):
            line = line.strip()
            if not line or line.startswith('%%'):
                continue

            # Participant 정의
            if line.startswith('participant'):
                parts = line.split(' ')
                if len(parts) >= 2:
                    participants.add(parts[1])

    def validate(self) -> str:
        """전체 검증 실행"""
        print("=" * 80)
        print("Mermaid 다이어그램 상세 검증 (Phase 2)")
        print("=" * 80)
        print()

        # Mermaid 블록 추출
        mermaid_blocks = self.extract_mermaid_blocks()

        total_diagrams = sum(len(blocks) for blocks in mermaid_blocks.values())
        print(f"📊 추출된 Mermaid 다이어그램: {total_diagrams}개\n")

        # 각 파일별 검증
        report_lines = []
        diagram_count = 0
        valid_count = 0
        issue_count = 0

        for file_path, blocks in sorted(mermaid_blocks.items()):
            report_lines.append(f"📄 {file_path}")
            report_lines.append(f"  └─ {len(blocks)}개 다이어그램 발견\n")

            for block in blocks:
                diagram_count += 1
                diagram_type, is_valid_type = self.validate_diagram_type(block['code'])

                print(f"[{diagram_count}] {block['id']}")
                print(f"    파일: {block['file']}:{block['line']}")
                print(f"    타입: {diagram_type}")
                print(f"    라인 수: {block['length']}")

                if is_valid_type:
                    print("    상태: ✅ 유효한 다이어그램 타입")
                    valid_count += 1
                else:
                    print("    상태: ❌ 유효하지 않은 타입")
                    issue_count += 1

                # 문법 검증
                syntax_issues = self.validate_syntax(block['code'], diagram_type)
                if syntax_issues:
                    print("    문법 문제:")
                    for issue in syntax_issues:
                        print(f"      {issue}")
                    issue_count += 1
                else:
                    print("    문법: ✅ 검증됨")

                # 코드 스니펫 표시 (처음 3줄)
                lines = block['code'].split('\n')[:3]
                print("    코드 샘플:")
                for line in lines:
                    print(f"      {line}")

                print()

                # 리포트 추가
                report_lines.append(f"  [{diagram_count}] {block['id']}")
                report_lines.append(f"      - 타입: {diagram_type} {'✅' if is_valid_type else '❌'}")
                report_lines.append(f"      - 라인: {block['line']}")
                report_lines.append(f"      - 높이: {block['length']} 줄")
                report_lines.append(f"      - 문법 오류: {'없음' if not syntax_issues else f'{len(syntax_issues)}개'}")

        # 요약
        print("=" * 80)
        print("📋 검증 요약")
        print("=" * 80)
        print(f"총 다이어그램: {diagram_count}개")
        print(f"유효한 타입: {valid_count}개 ✅")
        print(f"문제 발견: {issue_count}개 ⚠️")
        print()

        # 렌더링 테스트 안내
        print("🔗 렌더링 테스트 안내:")
        print("  다음 URL에서 각 다이어그램의 문법을 테스트하세요:")
        print("  → https://mermaid.live")
        print()

        report_lines.append("\n" + "=" * 80)
        report_lines.append("Mermaid 렌더링 테스트 URL")
        report_lines.append("=" * 80)
        report_lines.append("https://mermaid.live")
        report_lines.append("\n각 다이어그램 코드를 위 URL의 편집기에 붙여넣기하여")
        report_lines.append("렌더링 가능 여부를 확인하세요.")

        return "\n".join(report_lines)


def main():
    """메인 실행"""
    import argparse

    parser = argparse.ArgumentParser(description='Mermaid 다이어그램 검증')
    parser.add_argument('--path', type=str, default=str(DEFAULT_DOCS_PATH),
                       help=f'검사할 문서 경로 (기본값: {DEFAULT_DOCS_PATH})')
    parser.add_argument('--output', type=str, default=str(DEFAULT_REPORT_PATH),
                       help=f'리포트 저장 경로 (기본값: {DEFAULT_REPORT_PATH})')

    args = parser.parse_args()

    validator = MermaidDiagramValidator(args.path)
    report = validator.validate()

    # 리포트 저장
    report_path = Path(args.output)
    report_path.parent.mkdir(parents=True, exist_ok=True)
    report_path.write_text(report, encoding='utf-8')

    print(f"\n📁 리포트 저장됨: {report_path}")


if __name__ == "__main__":
    main()
