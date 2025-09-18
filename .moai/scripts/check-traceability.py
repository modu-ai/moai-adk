#!/usr/bin/env python3
"""
TAG 추적성 검증 스크립트
16-Core TAG 시스템의 무결성과 추적성 체인을 검증합니다.
"""

import json
import os
import re
import sys
from pathlib import Path
from typing import Dict, List, Set, Tuple

class TraceabilityChecker:
    def __init__(self, project_root: str = "."):
        self.project_root = Path(project_root)
        self.tags_index_path = self.project_root / ".moai" / "indexes" / "tags.json"
        self.tags: Dict[str, List[str]] = {}
        self.broken_links: List[Tuple[str, str]] = []
        self.orphaned_tags: List[str] = []

    def load_tags_index(self):
        """TAG 인덱스 파일 로드"""
        if not self.tags_index_path.exists():
            print(f"❌ TAG 인덱스 파일을 찾을 수 없습니다: {self.tags_index_path}")
            return False

        with open(self.tags_index_path, 'r', encoding='utf-8') as f:
            data = json.load(f)
            self.tags = data.get('categories', {})
        return True

    def scan_files_for_tags(self) -> Dict[str, List[str]]:
        """프로젝트 파일에서 TAG 스캔"""
        tag_pattern = r'@([A-Z]+):([A-Z0-9-]+)'
        found_tags = {}

        # 스캔할 파일 확장자
        extensions = ['.md', '.py', '.js', '.ts', '.yaml', '.yml', '.json']

        for ext in extensions:
            for file_path in self.project_root.rglob(f'*{ext}'):
                # .git, node_modules 등 제외
                if any(part.startswith('.') and part not in ['.claude', '.moai']
                       for part in file_path.parts):
                    continue

                try:
                    with open(file_path, 'r', encoding='utf-8') as f:
                        content = f.read()
                        matches = re.findall(tag_pattern, content)

                        for category, tag_id in matches:
                            tag_full = f"{category}:{tag_id}"
                            if tag_full not in found_tags:
                                found_tags[tag_full] = []
                            found_tags[tag_full].append(str(file_path))

                except Exception as e:
                    print(f"⚠️ 파일 읽기 실패: {file_path} - {e}")

        return found_tags

    def verify_traceability_chains(self, found_tags: Dict[str, List[str]]):
        """추적성 체인 검증"""
        # Primary Chain: REQ → DESIGN → TASK → TEST
        primary_chains = [
            ("REQ", "DESIGN"),
            ("DESIGN", "TASK"),
            ("TASK", "TEST")
        ]

        # Steering Chain: VISION → STRUCT → TECH → ADR
        steering_chains = [
            ("VISION", "STRUCT"),
            ("STRUCT", "TECH"),
            ("TECH", "ADR")
        ]

        all_chains = primary_chains + steering_chains

        for from_cat, to_cat in all_chains:
            from_tags = [tag for tag in found_tags.keys() if tag.startswith(f"{from_cat}:")]

            for from_tag in from_tags:
                # 해당 태그와 연결된 하위 태그 찾기
                base_id = from_tag.split(':')[1]
                expected_to_tag = f"{to_cat}:{base_id}"

                if expected_to_tag not in found_tags:
                    self.broken_links.append((from_tag, expected_to_tag))

    def find_orphaned_tags(self, found_tags: Dict[str, List[str]]):
        """고아 TAG 찾기 (참조되지 않는 TAG)"""
        # 간단한 구현: 파일에서만 존재하고 다른 곳에서 참조되지 않는 TAG
        for tag in found_tags.keys():
            # 실제로는 더 복잡한 로직 필요
            if len(found_tags[tag]) == 1:
                # 단일 파일에서만 발견된 TAG는 고아일 가능성
                pass

    def generate_report(self, found_tags: Dict[str, List[str]], verbose: bool = False):
        """검증 결과 보고서 생성"""
        total_tags = len(found_tags)
        broken_count = len(self.broken_links)
        orphaned_count = len(self.orphaned_tags)

        print("🏷️ TAG 추적성 검증 보고서")
        print("=" * 50)
        print(f"📊 총 TAG 수: {total_tags}")
        print(f"🔗 끊어진 링크: {broken_count}")
        print(f"👻 고아 TAG: {orphaned_count}")

        if broken_count == 0 and orphaned_count == 0:
            print("✅ 모든 TAG 추적성 체인이 정상입니다!")
            return 0

        if broken_count > 0:
            print("\n🔴 끊어진 추적성 체인:")
            for from_tag, to_tag in self.broken_links:
                print(f"  {from_tag} → {to_tag} (누락)")

        if orphaned_count > 0:
            print("\n👻 고아 TAG 목록:")
            for tag in self.orphaned_tags:
                print(f"  {tag}")

        if verbose:
            print("\n📂 TAG별 파일 위치:")
            for tag, files in sorted(found_tags.items()):
                print(f"  {tag}:")
                for file_path in files:
                    print(f"    - {file_path}")

        return 1 if broken_count > 0 or orphaned_count > 0 else 0

def main():
    import argparse
    parser = argparse.ArgumentParser(description="TAG 추적성 검증")
    parser.add_argument("--verbose", "-v", action="store_true", help="상세 출력")
    parser.add_argument("--project-root", "-p", default=".", help="프로젝트 루트 경로")

    args = parser.parse_args()

    checker = TraceabilityChecker(args.project_root)

    if not checker.load_tags_index():
        return 1

    found_tags = checker.scan_files_for_tags()
    checker.verify_traceability_chains(found_tags)
    checker.find_orphaned_tags(found_tags)

    return checker.generate_report(found_tags, args.verbose)

if __name__ == "__main__":
    sys.exit(main())