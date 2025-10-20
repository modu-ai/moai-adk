#!/usr/bin/env python3
"""
Skills Frontmatter 표준화 스크립트

공식 Claude Code 스펙에 맞게 SKILL.md frontmatter를 정리합니다.
- 필수 필드만 유지: name, description
- 비공식 필드 제거: tier, depends_on, allowed-tools, model
"""

import re
import sys
from pathlib import Path
from typing import Dict, List


def parse_frontmatter(content: str) -> tuple[Dict[str, str], str]:
    """YAML frontmatter와 본문을 분리하여 파싱"""
    pattern = r'^---\n(.*?)\n---\n(.*)$'
    match = re.match(pattern, content, re.DOTALL)

    if not match:
        return {}, content

    frontmatter_text = match.group(1)
    body = match.group(2)

    # YAML 파싱 (간단한 key: value 형식만)
    frontmatter = {}
    current_key = None
    current_value = []

    for line in frontmatter_text.split('\n'):
        if ':' in line and not line.startswith(' ') and not line.startswith('-'):
            # 이전 키 저장
            if current_key:
                frontmatter[current_key] = '\n'.join(current_value).strip()

            # 새로운 키-값 파싱
            key, value = line.split(':', 1)
            current_key = key.strip()
            current_value = [value.strip()] if value.strip() else []
        elif current_key:
            # 멀티라인 값 (description 등)
            current_value.append(line.strip())

    # 마지막 키 저장
    if current_key:
        frontmatter[current_key] = '\n'.join(current_value).strip()

    return frontmatter, body


def normalize_frontmatter(frontmatter: Dict[str, str]) -> Dict[str, str]:
    """공식 스펙에 맞게 frontmatter 정리"""
    # 필수 필드만 유지
    official_fields = ['name', 'description']

    normalized = {}
    for field in official_fields:
        if field in frontmatter:
            normalized[field] = frontmatter[field]

    return normalized


def format_frontmatter(frontmatter: Dict[str, str]) -> str:
    """frontmatter를 YAML 형식으로 포맷팅"""
    lines = ['---']

    for key, value in frontmatter.items():
        if '\n' in value or len(value) > 80:
            # 멀티라인 값
            lines.append(f'{key}: {value}')
        else:
            # 단일 라인 값
            lines.append(f'{key}: {value}')

    lines.append('---')
    return '\n'.join(lines)


def process_skill_file(file_path: Path, dry_run: bool = False) -> bool:
    """SKILL.md 파일 처리"""
    try:
        # 파일 읽기
        content = file_path.read_text(encoding='utf-8')

        # frontmatter 파싱
        frontmatter, body = parse_frontmatter(content)

        if not frontmatter:
            print(f"⚠️  {file_path.name}: frontmatter 없음")
            return False

        # 비공식 필드 확인
        removed_fields = []
        for field in ['tier', 'depends_on', 'allowed-tools', 'model', 'tools']:
            if field in frontmatter:
                removed_fields.append(field)

        if not removed_fields:
            print(f"✅ {file_path.parent.name}: 이미 표준화됨")
            return True

        # frontmatter 정규화
        normalized = normalize_frontmatter(frontmatter)

        # 새 내용 생성
        new_frontmatter = format_frontmatter(normalized)
        new_content = f"{new_frontmatter}\n{body}"

        # 결과 출력
        print(f"🔧 {file_path.parent.name}: {', '.join(removed_fields)} 제거")

        if not dry_run:
            # 파일 쓰기
            file_path.write_text(new_content, encoding='utf-8')
            print(f"   → 저장 완료")
        else:
            print(f"   → DRY RUN (저장 안 함)")

        return True

    except Exception as e:
        print(f"❌ {file_path.name}: 오류 - {e}")
        return False


def main():
    """메인 함수"""
    # Skills 디렉토리
    skills_dir = Path(__file__).parent.parent

    # 모든 SKILL.md 파일 찾기
    skill_files = list(skills_dir.glob('moai-*/SKILL.md'))

    print(f"\n📋 총 {len(skill_files)}개 Skills 파일 발견\n")

    # dry-run 옵션
    dry_run = '--dry-run' in sys.argv

    if dry_run:
        print("🔍 DRY RUN 모드 (실제 파일 수정 안 함)\n")

    # 각 파일 처리
    success_count = 0
    for file_path in sorted(skill_files):
        if process_skill_file(file_path, dry_run):
            success_count += 1

    # 결과 요약
    print(f"\n{'='*60}")
    print(f"✅ 성공: {success_count}/{len(skill_files)}개")
    print(f"{'='*60}\n")


if __name__ == '__main__':
    main()
