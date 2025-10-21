#!/usr/bin/env python3
"""
allowed-tools 필드 동기화 스크립트
.claude/skills/ 에서 src/moai_adk/templates/.claude/skills/ 로 동기화
"""

import sys
import re
from pathlib import Path

def extract_yaml_frontmatter(content: str) -> tuple:
    """
    YAML frontmatter를 추출한다
    Returns: (yaml_str, body)
    """
    if not content.startswith('---'):
        return None, None

    parts = content.split('---', 2)
    if len(parts) < 3:
        return None, None

    return parts[1], parts[2]

def parse_yaml_allowed_tools(yaml_str: str) -> list:
    """
    YAML 문자열에서 allowed-tools 값을 추출한다
    """
    lines = yaml_str.strip().split('\n')
    allowed_tools = []
    in_tools = False

    for i, line in enumerate(lines):
        if line.startswith('allowed-tools:'):
            in_tools = True
            # 한 줄 형식: allowed-tools: [Read, Write, Edit]
            if '[' in line:
                content = line.split('[')[1].split(']')[0]
                allowed_tools = [t.strip() for t in content.split(',')]
                break
            continue

        if in_tools:
            # 리스트 형식
            if line.strip().startswith('- '):
                tool = line.strip()[2:].strip()
                allowed_tools.append(tool)
            elif not line.startswith(' '):
                # 다음 필드 시작
                break

    return allowed_tools

def get_allowed_tools_from_source(skill_name: str) -> list:
    """
    .claude/skills/ 에서 allowed-tools 값을 읽어온다
    """
    source_file = Path(f"/Users/goos/MoAI/MoAI-ADK/.claude/skills/{skill_name}/SKILL.md")

    if not source_file.exists():
        return None

    content = source_file.read_text()
    yaml_str, body = extract_yaml_frontmatter(content)

    if yaml_str is None:
        return None

    return parse_yaml_allowed_tools(yaml_str)

def sync_allowed_tools(template_file: Path, allowed_tools: list):
    """
    템플릿 파일에 allowed-tools 필드를 동기화한다
    """
    content = template_file.read_text()
    yaml_str, body = extract_yaml_frontmatter(content)

    if yaml_str is None:
        return False

    # 기존 allowed-tools 제거
    new_yaml_lines = []
    skip_tools = False

    for line in yaml_str.split('\n'):
        if line.startswith('allowed-tools:'):
            skip_tools = True
            # 한 줄 형식인지 확인
            if '[' in line or ']' in line:
                skip_tools = False
            continue

        if skip_tools:
            # 리스트 형식인 경우
            if line.strip().startswith('- '):
                continue
            elif line and not line.startswith(' '):
                # 다음 필드 시작
                skip_tools = False

        new_yaml_lines.append(line)

    # allowed-tools 추가 (마지막에 추가)
    new_yaml = '\n'.join(new_yaml_lines).rstrip()
    if allowed_tools:
        tools_str = ', '.join(allowed_tools)
        new_yaml += f"\nallowed-tools:\n"
        for tool in allowed_tools:
            new_yaml += f"  - {tool}\n"

    new_content = f"---\n{new_yaml}---{body}"
    template_file.write_text(new_content)

    return True

def main():
    templates_dir = Path("/Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.claude/skills")

    synced_count = 0
    failed_count = 0

    for skill_dir in sorted(templates_dir.iterdir()):
        if not skill_dir.is_dir():
            continue

        skill_name = skill_dir.name
        template_file = skill_dir / "SKILL.md"

        if not template_file.exists():
            print(f"Skip (no SKILL.md): {skill_name}")
            continue

        # .claude/skills/ 에서 allowed-tools 읽기
        allowed_tools = get_allowed_tools_from_source(skill_name)

        if allowed_tools is None:
            print(f"Skip (no allowed-tools in source): {skill_name}")
            failed_count += 1
            continue

        # 템플릿 파일 동기화
        if sync_allowed_tools(template_file, allowed_tools):
            print(f"Synced: {skill_name} ({len(allowed_tools)} tools)")
            synced_count += 1
        else:
            print(f"Failed: {skill_name}")
            failed_count += 1

    print(f"\n✅ Synced: {synced_count}")
    print(f"❌ Failed: {failed_count}")

if __name__ == "__main__":
    main()
