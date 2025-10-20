#!/usr/bin/env python3
"""패키지용 Skills Frontmatter 표준화 스크립트"""

import re
from pathlib import Path

def parse_frontmatter(content: str):
    """YAML frontmatter와 본문 분리"""
    pattern = r'^---\n(.*?)\n---\n(.*)$'
    match = re.match(pattern, content, re.DOTALL)

    if not match:
        return {}, content

    frontmatter_text = match.group(1)
    body = match.group(2)

    frontmatter = {}
    current_key = None
    current_value = []

    for line in frontmatter_text.split('\n'):
        if ':' in line and not line.startswith(' ') and not line.startswith('-'):
            if current_key:
                frontmatter[current_key] = '\n'.join(current_value).strip()

            key, value = line.split(':', 1)
            current_key = key.strip()
            current_value = [value.strip()] if value.strip() else []
        elif current_key:
            current_value.append(line.strip())

    if current_key:
        frontmatter[current_key] = '\n'.join(current_value).strip()

    return frontmatter, body

def normalize_frontmatter(frontmatter):
    """공식 스펙에 맞게 정리"""
    return {
        'name': frontmatter.get('name', ''),
        'description': frontmatter.get('description', '')
    }

def format_frontmatter(frontmatter):
    """YAML 형식으로 포맷팅"""
    return f"---\nname: {frontmatter['name']}\ndescription: {frontmatter['description']}\n---"

# 패키지 Skills 디렉토리
pkg_skills_dir = Path(__file__).parent
skill_files = list(pkg_skills_dir.glob('moai-*/SKILL.md'))

print(f"\n📋 총 {len(skill_files)}개 Skills 파일 발견\n")

success_count = 0
for file_path in sorted(skill_files):
    try:
        content = file_path.read_text(encoding='utf-8')
        frontmatter, body = parse_frontmatter(content)

        if not frontmatter:
            print(f"⚠️  {file_path.parent.name}: frontmatter 없음")
            continue

        # 비공식 필드 확인
        removed_fields = []
        for field in ['tier', 'depends_on', 'allowed-tools', 'model', 'tools']:
            if field in frontmatter:
                removed_fields.append(field)

        if not removed_fields:
            print(f"✅ {file_path.parent.name}: 이미 표준화됨")
            success_count += 1
            continue

        # 정규화
        normalized = normalize_frontmatter(frontmatter)
        new_frontmatter = format_frontmatter(normalized)
        new_content = f"{new_frontmatter}\n{body}"

        # 파일 쓰기
        file_path.write_text(new_content, encoding='utf-8')
        print(f"🔧 {file_path.parent.name}: {', '.join(removed_fields)} 제거")
        print(f"   → 저장 완료")
        success_count += 1

    except Exception as e:
        print(f"❌ {file_path.name}: 오류 - {e}")

print(f"\n{'='*60}")
print(f"✅ 성공: {success_count}/{len(skill_files)}개")
print(f"{'='*60}\n")
