#!/usr/bin/env python3
"""
Skills 표준화 통합 스크립트
- YAML 필드 정리 (version, author, license, tags, model 제거)
- allowed-tools 추가 (누락된 스킬에만)
"""

import sys
import re
from pathlib import Path

# Alfred 에이전트 도구
ALFRED_TOOLS = ["Read", "Write", "Edit", "Bash", "TodoWrite"]
# Lang 스킬 도구
LANG_TOOLS = ["Read", "Bash"]
# Domain 스킬 도구
DOMAIN_TOOLS = ["Read", "Bash"]

def parse_yaml_frontmatter(content):
    """YAML frontmatter 파싱 (간단한 파서)"""
    if not content.startswith('---'):
        return None, content
    
    parts = content.split('---', 2)
    if len(parts) < 3:
        return None, content
    
    yaml_str = parts[1]
    body = parts[2]
    
    # YAML 파싱 (딕셔너리로)
    data = {}
    current_key = None
    in_list = False
    
    for line in yaml_str.strip().split('\n'):
        if not line.strip():
            continue
        
        # 리스트 아이템
        if line.strip().startswith('-'):
            if in_list and current_key:
                if isinstance(data[current_key], list):
                    data[current_key].append(line.strip()[1:].strip())
            continue
        
        # 키-값 쌍
        if ':' in line:
            key, value = line.split(':', 1)
            key = key.strip()
            value = value.strip()
            
            if not value:  # 리스트 시작
                in_list = True
                current_key = key
                data[key] = []
            else:
                in_list = False
                current_key = None
                data[key] = value
    
    return data, body

def build_yaml_frontmatter(data):
    """딕셔너리를 YAML frontmatter로 변환"""
    lines = []
    for key, value in data.items():
        if isinstance(value, list):
            lines.append(f"{key}:")
            for item in value:
                lines.append(f"  - {item}")
        else:
            lines.append(f"{key}: {value}")
    
    return '\n'.join(lines)

def standardize_skill(skill_file):
    """스킬 파일 표준화"""
    content = skill_file.read_text()
    
    data, body = parse_yaml_frontmatter(content)
    
    if data is None:
        print(f"⚠️  No YAML frontmatter: {skill_file}")
        return False
    
    # 보존할 필드만 추출
    preserved = {}
    
    if 'name' in data:
        preserved['name'] = data['name']
    if 'description' in data:
        preserved['description'] = data['description']
    
    # allowed-tools 처리
    if 'allowed-tools' in data:
        # 이미 있으면 유지
        preserved['allowed-tools'] = data['allowed-tools']
    else:
        # 없으면 스킬 유형별로 추가
        name = data.get('name', '')
        
        if 'alfred' in name:
            tools = ALFRED_TOOLS
        elif 'lang' in name:
            tools = LANG_TOOLS
        elif 'domain' in name:
            tools = DOMAIN_TOOLS
        elif 'claude-code' in name:
            # moai-claude-code는 이미 allowed-tools 있음 (건너뛰기)
            tools = None
        else:
            # 기본값
            tools = ["Read"]
        
        if tools:
            preserved['allowed-tools'] = tools
    
    # 파일 재작성
    new_yaml = build_yaml_frontmatter(preserved)
    new_content = f"---\n{new_yaml}\n---{body}"
    
    skill_file.write_text(new_content)
    print(f"✅ Standardized: {skill_file.name}")
    return True

def main():
    """메인 함수"""
    base_dir = Path("/Users/goos/MoAI/MoAI-ADK")
    
    # .claude/skills/
    skills_dir = base_dir / ".claude/skills"
    success_count = 0
    fail_count = 0
    
    for skill_dir in sorted(skills_dir.glob("moai-*")):
        skill_file = skill_dir / "SKILL.md"
        if skill_file.exists():
            try:
                if standardize_skill(skill_file):
                    success_count += 1
                else:
                    fail_count += 1
            except Exception as e:
                print(f"❌ Error: {skill_file.name} - {e}")
                fail_count += 1
    
    # src/moai_adk/templates/.claude/skills/
    templates_dir = base_dir / "src/moai_adk/templates/.claude/skills"
    for skill_dir in sorted(templates_dir.glob("moai-*")):
        skill_file = skill_dir / "SKILL.md"
        if skill_file.exists():
            try:
                if standardize_skill(skill_file):
                    success_count += 1
                else:
                    fail_count += 1
            except Exception as e:
                print(f"❌ Error: {skill_file.name} - {e}")
                fail_count += 1
    
    print(f"\n📊 Summary: {success_count} succeeded, {fail_count} failed")
    return 0 if fail_count == 0 else 1

if __name__ == "__main__":
    sys.exit(main())
