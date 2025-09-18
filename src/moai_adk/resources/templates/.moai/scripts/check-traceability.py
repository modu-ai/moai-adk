#!/usr/bin/env python3
"""
TAG 추적성 검증 스크립트 (향상판)
16-Core TAG 시스템의 무결성과 추적성 체인을 검증/갱신합니다.

기능:
- 프로젝트에서 TAG 스캔(@CAT:ID)
- SPEC 디렉터리 단위로 추적성 체인(REQ→DESIGN→TASK→TEST, VISION→STRUCT→TECH→ADR) 자동 구성(--update)
- 인덱스 기반 검증(없으면 휴리스틱 체인으로 검증)
"""

import json
import re
import sys
from pathlib import Path
from typing import Dict, List, Tuple, Set


PRIMARY_CHAIN = [("REQ", "DESIGN"), ("DESIGN", "TASK"), ("TASK", "TEST")]
STEERING_CHAIN = [("VISION", "STRUCT"), ("STRUCT", "TECH"), ("TECH", "ADR")]
ALL_CHAINS = PRIMARY_CHAIN + STEERING_CHAIN


class TraceabilityChecker:
    def __init__(self, project_root: str = "."):
        self.project_root = Path(project_root)
        self.tags_index_path = self.project_root / ".moai" / "indexes" / "tags.json"
        self.index: Dict = {}
        self.broken_links: List[Tuple[str, str]] = []
        self.orphaned_tags: List[str] = []

    def load_or_init_index(self) -> None:
        if self.tags_index_path.exists():
            with open(self.tags_index_path, "r", encoding="utf-8") as f:
                try:
                    self.index = json.load(f)
                except Exception:
                    self.index = {}
        if not self.index:
            self.index = {
                "version": "16-core",
                "categories": {
                    "SPEC": {"REQ": [], "DESIGN": [], "TASK": []},
                    "STEERING": {"VISION": [], "STRUCT": [], "TECH": [], "ADR": []},
                    "IMPLEMENTATION": {"FEATURE": [], "API": [], "TEST": [], "DATA": []},
                    "QUALITY": {"PERF": [], "SEC": [], "DEBT": [], "TODO": []},
                },
                "traceability_chains": [],
                "orphaned_tags": [],
                "statistics": {
                    "total_tags": 0,
                    "complete_chains": 0,
                    "broken_links": 0,
                    "coverage_percentage": 0,
                },
            }

    def save_index(self) -> None:
        self.tags_index_path.parent.mkdir(parents=True, exist_ok=True)
        with open(self.tags_index_path, "w", encoding="utf-8") as f:
            json.dump(self.index, f, indent=2, ensure_ascii=False)

    def scan_files_for_tags(self) -> Dict[str, List[str]]:
        tag_pattern = r"@([A-Z]+):([A-Z0-9-]+)"
        found: Dict[str, List[str]] = {}
        exts = [".md", ".py", ".js", ".ts", ".yaml", ".yml", ".json"]

        for ext in exts:
            for file_path in self.project_root.rglob(f"*{ext}"):
                # 숨김 디렉토리 제외(.git 등) 단, .claude, .moai는 허용
                if any(part.startswith(".") and part not in [".claude", ".moai"] for part in file_path.parts):
                    continue
                try:
                    content = file_path.read_text(encoding="utf-8")
                except Exception:
                    continue
                for cat, tid in re.findall(tag_pattern, content):
                    tag = f"{cat}:{tid}"
                    found.setdefault(tag, []).append(str(file_path))
        return found

    def scan_files_for_links(self) -> List[Dict[str, str]]:
        """@LINK:CAT:ID->CAT:ID 형식의 명시적 링크 스캔"""
        link_pattern = r"@LINK:([A-Z]+:[A-Z0-9-]+)->([A-Z]+:[A-Z0-9-]+)"
        links: List[Dict[str, str]] = []
        exts = [".md", ".py", ".js", ".ts", ".yaml", ".yml", ".json"]
        for ext in exts:
            for file_path in self.project_root.rglob(f"*{ext}"):
                if any(part.startswith(".") and part not in [".claude", ".moai"] for part in file_path.parts):
                    continue
                try:
                    content = file_path.read_text(encoding="utf-8")
                except Exception:
                    continue
                for frm, to in re.findall(link_pattern, content):
                    links.append({"from": frm, "to": to})
        return links

    def group_by_spec(self, found: Dict[str, List[str]]) -> Dict[str, Set[str]]:
        """SPEC 디렉터리별 태그 묶기: key=SPEC-xxx, value=tags(set)."""
        groups: Dict[str, Set[str]] = {}
        for tag, paths in found.items():
            for p in paths:
                path = Path(p)
                parts = list(path.parts)
                if ".moai" in parts and "specs" in parts:
                    try:
                        spec_idx = parts.index("specs")
                        spec_name = parts[spec_idx + 1]
                        if spec_name.startswith("SPEC-"):
                            groups.setdefault(spec_name, set()).add(tag)
                    except Exception:
                        continue
        return groups

    def build_chains_from_groups(self, groups: Dict[str, Set[str]]) -> List[Dict[str, str]]:
        chains: List[Dict[str, str]] = []
        for _spec, tags in groups.items():
            by_cat: Dict[str, List[str]] = {}
            for t in tags:
                cat = t.split(":", 1)[0]
                by_cat.setdefault(cat, []).append(t)
            # Primary
            for a, b in PRIMARY_CHAIN:
                for t_from in by_cat.get(a, []):
                    for t_to in by_cat.get(b, []):
                        chains.append({"from": t_from, "to": t_to})
            # Steering
            for a, b in STEERING_CHAIN:
                for t_from in by_cat.get(a, []):
                    for t_to in by_cat.get(b, []):
                        chains.append({"from": t_from, "to": t_to})
        return chains

    def verify(self, found: Dict[str, List[str]], chains: List[Dict[str, str]]):
        # 빠른 조회 셋
        found_set = set(found.keys())
        sources = {c["from"] for c in chains}
        targets = {c["to"] for c in chains}

        # 체인이 없는 경우 휴리스틱: 같은 SPEC 내 카테고리 존재 여부만 검증
        if not chains:
            groups = self.group_by_spec(found)
            for _spec, tags in groups.items():
                tagset = set(tags)
                for a, b in ALL_CHAINS:
                    a_tags = {t for t in tagset if t.startswith(f"{a}:")}
                    b_tags = {t for t in tagset if t.startswith(f"{b}:")}
                    if a_tags and not b_tags:
                        for t in a_tags:
                            self.broken_links.append((t, f"{b}:<missing>"))
            # orphan 판단(체인이 없으면 판단 불가) 생략
            return

        # 체인 기반 검증
        linked_from = set()
        linked_to = set()
        for c in chains:
            f = c.get("from")
            t = c.get("to")
            if f in found_set and t in found_set:
                linked_from.add(f)
                linked_to.add(t)
            else:
                # 양끝이 모두 있어야 유효 체인, 없으면 끊어진 링크로 보고
                self.broken_links.append((f if f in found_set else f+"(?)", t if t in found_set else t+"(?)"))

        # 고아 태그: 발견됐지만 어떤 체인에도 포함되지 않은 태그
        for tag in found_set:
            if tag not in linked_from and tag not in linked_to:
                self.orphaned_tags.append(tag)

    def update_index(self, found: Dict[str, List[str]], chains: List[Dict[str, str]]):
        # 카테고리 별 목록 업데이트(중복 제거)
        cats = {
            "REQ": "SPEC", "DESIGN": "SPEC", "TASK": "SPEC",
            "VISION": "STEERING", "STRUCT": "STEERING", "TECH": "STEERING", "ADR": "STEERING",
            "FEATURE": "IMPLEMENTATION", "API": "IMPLEMENTATION", "TEST": "IMPLEMENTATION", "DATA": "IMPLEMENTATION",
            "PERF": "QUALITY", "SEC": "QUALITY", "DEBT": "QUALITY", "TODO": "QUALITY",
        }
        for tag in found.keys():
            cat = tag.split(":", 1)[0]
            grp = cats.get(cat)
            if grp and grp in self.index.get("categories", {}):
                arr = self.index["categories"][grp].setdefault(cat, [])
                if tag not in arr:
                    arr.append(tag)
        self.index["traceability_chains"] = chains
        self.index["orphaned_tags"] = self.orphaned_tags
        self.index["statistics"] = {
            "total_tags": len(found),
            "complete_chains": len(chains) - len(self.broken_links),
            "broken_links": len(self.broken_links),
            "coverage_percentage": 0,
        }

    def report(self, found: Dict[str, List[str]], verbose: bool) -> int:
        print("🏷️ TAG 추적성 검증 보고서")
        print("=" * 50)
        print(f"📊 총 TAG 수: {len(found)}")
        print(f"🔗 끊어진 링크: {len(self.broken_links)}")
        print(f"👻 고아 TAG: {len(self.orphaned_tags)}")
        if len(self.broken_links) == 0 and len(self.orphaned_tags) == 0:
            print("✅ 모든 TAG 추적성 체인이 정상입니다!")
        else:
            if self.broken_links:
                print("\n🔴 끊어진 추적성 체인:")
                for f, t in self.broken_links:
                    print(f"  {f} → {t} (누락)")
            if self.orphaned_tags:
                print("\n👻 고아 TAG 목록:")
                for tag in self.orphaned_tags:
                    print(f"  {tag}")
        if verbose:
            print("\n📂 TAG별 파일 위치:")
            for tag, files in sorted(found.items()):
                print(f"  {tag}:")
                for fp in files:
                    print(f"    - {fp}")
        return 1 if (self.broken_links or self.orphaned_tags) else 0


def main():
    import argparse
    parser = argparse.ArgumentParser(description="TAG 추적성 검증")
    parser.add_argument("--verbose", "-v", action="store_true", help="상세 출력")
    parser.add_argument("--project-root", "-p", default=".", help="프로젝트 루트 경로")
    parser.add_argument("--update", action="store_true", help="인덱스를 추적성 체인으로 갱신")

    args = parser.parse_args()

    checker = TraceabilityChecker(args.project_root)
    checker.load_or_init_index()

    found = checker.scan_files_for_tags()

    # 인덱스에 체인이 있으면 우선 사용, 없으면 SPEC 그룹 휴리스틱으로 생성
    chains = checker.index.get("traceability_chains") or []
    if not chains:
        groups = checker.group_by_spec(found)
        chains = checker.build_chains_from_groups(groups)
    # 명시적 @LINK 체인 추가(중복 제거)
    explicit = checker.scan_files_for_links()
    all_chains = {(c["from"], c["to"]) for c in chains}
    for link in explicit:
        key = (link["from"], link["to"])
        if key not in all_chains:
            chains.append(link)
            all_chains.add(key)

    # 검증
    checker.verify(found, chains)

    # 요청 시 인덱스 갱신
    if args.update:
        checker.update_index(found, chains)
        checker.save_index()

    return checker.report(found, args.verbose)


if __name__ == "__main__":
    sys.exit(main())
