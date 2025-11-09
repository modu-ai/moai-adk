#!/usr/bin/env python3
"""
한국어 문서 헤더에서 이모지 제거 스크립트

규칙:
- 헤더(#으로 시작)에서만 이모지 제거
- GitHub 스타일 이모지 단축코드 제거 (:emoji_name:)
- 유니코드 이모지 제거
- Material Icons는 유지 (본문 사용 허용)
- 본문 이모지는 유지
"""

import re
from pathlib import Path

# 한국어 문서 경로
KO_DOCS_PATH = Path("src/ko")

# GitHub 스타일 이모지 단축코드 패턴
GITHUB_EMOJI_PATTERN = r'^(#{1,6})\s+:[a-z_]+:\s+'

# 유니코드 이모지 범위 (일반적인 이모지 유니코드 블록)
UNICODE_EMOJI_PATTERN = r'^(#{1,6})\s+[\U0001F300-\U0001F9FF\U00002600-\U000027BF\U0001F000-\U0001F02F\U0001F0A0-\U0001F0FF\U00002700-\U000027BF\U00002B50-\U00002B55\U0001F600-\U0001F64F\U0001F680-\U0001F6FF\U0001F910-\U0001F96B\U0001F980-\U0001F9E0\U00002300-\U000023FF\U00002460-\U000024FF\U00002500-\U0000257F\U00002580-\U0000259F\U000025A0-\U000025FF\U00002190-\U000021FF\U00000030-\U00000039]\uFE0F?\s+'


def remove_header_emojis(content: str) -> tuple[str, int]:
    """
    마크다운 콘텐츠의 헤더에서 이모지 제거

    Returns:
        (수정된 내용, 제거된 이모지 수)
    """
    lines = content.split('\n')
    modified_lines = []
    removed_count = 0

    for line in lines:
        original_line = line

        # GitHub 스타일 이모지 단축코드 제거
        line = re.sub(GITHUB_EMOJI_PATTERN, r'\1 ', line)

        # 유니코드 이모지 제거
        line = re.sub(UNICODE_EMOJI_PATTERN, r'\1 ', line)

        if line != original_line:
            removed_count += 1
            print(f"  수정: {original_line.strip()}")
            print(f"    → {line.strip()}")

        modified_lines.append(line)

    return '\n'.join(modified_lines), removed_count


def process_markdown_files():
    """한국어 마크다운 파일 모두 처리"""

    if not KO_DOCS_PATH.exists():
        print(f"❌ 경로를 찾을 수 없습니다: {KO_DOCS_PATH}")
        return

    # 모든 마크다운 파일 찾기
    md_files = list(KO_DOCS_PATH.rglob("*.md"))

    if not md_files:
        print("❌ 마크다운 파일을 찾을 수 없습니다.")
        return

    print(f"📁 {len(md_files)}개의 마크다운 파일 발견\n")

    total_removed = 0
    modified_files = 0

    for md_file in md_files:
        try:
            # 파일 읽기
            content = md_file.read_text(encoding='utf-8')

            # 헤더 이모지 제거
            modified_content, removed_count = remove_header_emojis(content)

            # 변경사항이 있으면 파일 업데이트
            if removed_count > 0:
                md_file.write_text(modified_content, encoding='utf-8')
                modified_files += 1
                total_removed += removed_count
                print(f"✅ {md_file.relative_to('src/ko')}: {removed_count}개 이모지 제거\n")

        except Exception as e:
            print(f"❌ 오류 ({md_file.name}): {e}\n")

    # 결과 요약
    print("=" * 60)
    print(f"✅ 완료!")
    print(f"  - 수정된 파일: {modified_files}개")
    print(f"  - 제거된 이모지: {total_removed}개")
    print("=" * 60)


if __name__ == "__main__":
    process_markdown_files()
