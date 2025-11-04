#!/usr/bin/env python3
"""
Claude Code 파일 리팩토링 상태 분석 스크립트
- 파일 크기, Python 코드 블록 개수, 리팩토링 필요도 분석
"""

import re
from pathlib import Path
from dataclasses import dataclass
from typing import List, Tuple

@dataclass
class FileAnalysis:
    path: Path
    line_count: int
    python_blocks: int
    imperative_ratio: float
    priority: str  # HIGH, MEDIUM, LOW

    def __str__(self):
        return f"{self.path.name:40} | {self.line_count:4} lines | {self.python_blocks:2} py-blocks | {self.imperative_ratio:5.1%} | {self.priority}"


def count_python_blocks(content: str) -> int:
    """Python 코드 블록 개수를 센다"""
    return len(re.findall(r'```python', content, re.IGNORECASE))


def estimate_imperative_ratio(content: str) -> float:
    """명령형 비율을 추정한다 (간단한 휴리스틱)"""
    # 명령형 키워드
    imperative_keywords = [
        'Follow these steps',
        'Execute the following',
        'Perform these actions',
        'Run the following',
        'Do the following',
        'Take these actions',
        'Complete these tasks',
        '다음 단계를 따르세요',
        '다음을 실행하세요',
        '다음 작업을 수행하세요',
    ]

    # 선언형 키워드
    declarative_keywords = [
        'You are',
        'Your role is',
        'Your responsibility',
        'You should',
        'You must',
        'You will',
        'Your job is',
        '당신은',
        '당신의 역할은',
        '당신의 책임은',
    ]

    imperative_count = sum(1 for kw in imperative_keywords if kw.lower() in content.lower())
    declarative_count = sum(1 for kw in declarative_keywords if kw.lower() in content.lower())

    total = imperative_count + declarative_count
    if total == 0:
        return 0.5  # 중립

    return imperative_count / total


def determine_priority(line_count: int, python_blocks: int, imperative_ratio: float) -> str:
    """리팩토링 우선순위 결정"""
    # HIGH: 큰 파일 + Python 코드 많음 + 선언형 구조
    if line_count > 200 and python_blocks >= 3 and imperative_ratio < 0.3:
        return "🔴 HIGH"

    # HIGH: Python 코드가 매우 많음
    if python_blocks >= 5 and imperative_ratio < 0.4:
        return "🔴 HIGH"

    # MEDIUM: 중간 크기 + Python 코드 있음
    if (line_count > 100 and python_blocks >= 2) or (imperative_ratio < 0.5):
        return "🟡 MEDIUM"

    # LOW: 이미 명령형에 가깝거나 작은 파일
    return "🟢 LOW"


def analyze_file(file_path: Path) -> FileAnalysis:
    """파일 분석"""
    content = file_path.read_text(encoding='utf-8')
    lines = content.split('\n')

    # YAML frontmatter 제외
    if content.startswith('---'):
        yaml_end = content.find('---', 3)
        if yaml_end != -1:
            content = content[yaml_end+3:]

    line_count = len(lines)
    python_blocks = count_python_blocks(content)
    imperative_ratio = estimate_imperative_ratio(content)
    priority = determine_priority(line_count, python_blocks, imperative_ratio)

    return FileAnalysis(
        path=file_path,
        line_count=line_count,
        python_blocks=python_blocks,
        imperative_ratio=imperative_ratio,
        priority=priority
    )


def main():
    base_dir = Path('/Users/goos/MoAI/MoAI-ADK/.claude')

    # 분석할 파일 목록
    agent_files = sorted((base_dir / 'agents/alfred').glob('*.md'))
    command_files = sorted((base_dir / 'commands/alfred').glob('*.md'))

    all_files = agent_files + command_files

    print("=" * 100)
    print("📊 Claude Code 파일 리팩토링 상태 분석")
    print("=" * 100)
    print()

    # 분석 실행
    analyses: List[FileAnalysis] = []
    for file_path in all_files:
        analysis = analyze_file(file_path)
        analyses.append(analysis)

    # 우선순위별 분류
    high_priority = [a for a in analyses if '🔴 HIGH' in a.priority]
    medium_priority = [a for a in analyses if '🟡 MEDIUM' in a.priority]
    low_priority = [a for a in analyses if '🟢 LOW' in a.priority]

    # 출력
    print(f"📁 전체 파일: {len(analyses)}개")
    print(f"   🔴 HIGH: {len(high_priority)}개")
    print(f"   🟡 MEDIUM: {len(medium_priority)}개")
    print(f"   🟢 LOW: {len(low_priority)}개")
    print()

    if high_priority:
        print("=" * 100)
        print("🔴 HIGH 우선순위 (긴급 리팩토링 필요)")
        print("=" * 100)
        for analysis in sorted(high_priority, key=lambda a: (-a.python_blocks, -a.line_count)):
            print(f"  {analysis}")
        print()

    if medium_priority:
        print("=" * 100)
        print("🟡 MEDIUM 우선순위 (일반 리팩토링 필요)")
        print("=" * 100)
        for analysis in sorted(medium_priority, key=lambda a: (-a.python_blocks, -a.line_count)):
            print(f"  {analysis}")
        print()

    if low_priority:
        print("=" * 100)
        print("🟢 LOW 우선순위 (경량 개선)")
        print("=" * 100)
        for analysis in sorted(low_priority, key=lambda a: (-a.python_blocks, -a.line_count)):
            print(f"  {analysis}")
        print()

    # 요약 통계
    print("=" * 100)
    print("📈 요약 통계")
    print("=" * 100)
    total_lines = sum(a.line_count for a in analyses)
    total_py_blocks = sum(a.python_blocks for a in analyses)
    avg_imperative = sum(a.imperative_ratio for a in analyses) / len(analyses) if analyses else 0

    print(f"  전체 라인 수: {total_lines:,}")
    print(f"  전체 Python 블록: {total_py_blocks}")
    print(f"  평균 명령형 비율: {avg_imperative:.1%}")
    print()

    # 예상 작업 시간
    print("=" * 100)
    print("⏱️  예상 리팩토링 시간")
    print("=" * 100)
    high_hours = len(high_priority) * 0.5  # 파일당 30분
    medium_hours = len(medium_priority) * 0.3  # 파일당 20분
    low_hours = len(low_priority) * 0.15  # 파일당 10분
    total_hours = high_hours + medium_hours + low_hours

    print(f"  🔴 HIGH: {high_hours:.1f}시간 ({len(high_priority)}개 × 30분)")
    print(f"  🟡 MEDIUM: {medium_hours:.1f}시간 ({len(medium_priority)}개 × 20분)")
    print(f"  🟢 LOW: {low_hours:.1f}시간 ({len(low_priority)}개 × 10분)")
    print(f"  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print(f"  📊 전체 예상: {total_hours:.1f}시간")
    print()


if __name__ == '__main__':
    main()
