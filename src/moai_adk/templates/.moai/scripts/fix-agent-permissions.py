#!/usr/bin/env python3
"""
Agent permissionMode 자동 수정 스크립트
Invalid 값을 올바른 값으로 일괄 변경
"""

import os
import re
from pathlib import Path
from typing import List, Tuple

# 매핑 규칙
PERMISSION_MODE_MAP = {
    'auto': 'dontAsk',
    'ask': 'default'
}

def fix_agent_file(file_path: Path) -> Tuple[bool, str]:
    """에이전트 파일의 permissionMode 수정"""
    try:
        content = file_path.read_text(encoding='utf-8')

        # YAML frontmatter에서 permissionMode 찾기
        pattern = r'^permissionMode:\s*(auto|ask)\s*$'

        def replace_permission_mode(match):
            old_value = match.group(1)
            new_value = PERMISSION_MODE_MAP.get(old_value, old_value)
            return f'permissionMode: {new_value}'

        new_content, count = re.subn(
            pattern,
            replace_permission_mode,
            content,
            flags=re.MULTILINE
        )

        if count > 0:
            file_path.write_text(new_content, encoding='utf-8')
            return True, f"Updated {count} occurrence(s)"

        return False, "No changes needed"

    except Exception as e:
        return False, f"Error: {str(e)}"

def main():
    """모든 에이전트 파일 처리"""
    project_root = Path(__file__).parent.parent.parent

    # 로컬 및 템플릿 에이전트 디렉토리
    agent_dirs = [
        project_root / '.claude' / 'agents' / 'moai',
        project_root / 'src' / 'moai_adk' / 'templates' / '.claude' / 'agents' / 'moai'
    ]

    total_fixed = 0
    results = []

    for agent_dir in agent_dirs:
        if not agent_dir.exists():
            continue

        for agent_file in agent_dir.glob('*.md'):
            fixed, message = fix_agent_file(agent_file)

            if fixed:
                total_fixed += 1
                results.append(f"✅ {agent_file.name}: {message}")
            else:
                results.append(f"⏭️  {agent_file.name}: {message}")

    # 결과 출력
    print("\n" + "="*60)
    print("Agent permissionMode Fix Results")
    print("="*60)

    for result in results:
        print(result)

    print(f"\n📊 Total files fixed: {total_fixed}")
    print(f"📋 Mapping: auto→dontAsk, ask→default")

if __name__ == '__main__':
    main()
