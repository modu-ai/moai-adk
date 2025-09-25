#!/usr/bin/env python3
# @TASK:TRACEABILITY-CHECK-011
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
from datetime import datetime
from pathlib import Path

import click

PRIMARY_CHAIN = [("REQ", "DESIGN"), ("DESIGN", "TASK"), ("TASK", "TEST")]
STEERING_CHAIN = [("VISION", "STRUCT"), ("STRUCT", "TECH"), ("TECH", "ADR")]
ALL_CHAINS = PRIMARY_CHAIN + STEERING_CHAIN


class TraceabilityChecker:
    def __init__(self, project_root: str = "."):
        self.project_root = Path(project_root)
        self.tags_index_path = self.project_root / ".moai" / "indexes" / "tags.json"
        self.index: dict = {}
        self.broken_links: list[tuple[str, str]] = []
        self.orphaned_tags: list[str] = []

    def load_or_init_index(self) -> None:
        if self.tags_index_path.exists():
            with open(self.tags_index_path, encoding="utf-8") as f:
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

    def scan_files_for_tags(self) -> dict[str, list[str]]:
        tag_pattern = r"@([A-Z]+):([A-Z0-9-]+)"
        found: dict[str, list[str]] = {}
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

    def scan_files_for_links(self) -> list[dict[str, str]]:
        """@LINK:CAT:ID->CAT:ID 형식의 명시적 링크 스캔"""
        link_pattern = r"@LINK:([A-Z]+:[A-Z0-9-]+)->([A-Z]+:[A-Z0-9-]+)"
        links: list[dict[str, str]] = []
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

    def group_by_spec(self, found: dict[str, list[str]]) -> dict[str, set[str]]:
        """SPEC 디렉터리별 태그 묶기: key=SPEC-xxx, value=tags(set)."""
        groups: dict[str, set[str]] = {}
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

    def build_inferred_chains(
        self,
        found: dict[str, list[str]],
        explicit_links: list[dict[str, str]],
    ) -> list[dict[str, str]]:
        chain_set: set[tuple[str, str]] = set()

        # 1) SPEC 디렉토리별 체인
        for tags in self.group_by_spec(found).values():
            by_cat: dict[str, list[str]] = {}
            for tag in tags:
                category = tag.split(":", 1)[0]
                by_cat.setdefault(category, []).append(tag)
            for a, b in PRIMARY_CHAIN + STEERING_CHAIN:
                for frm in by_cat.get(a, []):
                    for to in by_cat.get(b, []):
                        chain_set.add((frm, to))

        # 2) 태그 ID 접미사(-NNN) 기반 체인
        suffix_groups: dict[str, set[str]] = {}
        for tag in found:
            match = re.search(r"-(\d{3})$", tag)
            if match:
                suffix_groups.setdefault(match.group(1), set()).add(tag)
        for tags in suffix_groups.values():
            by_cat: dict[str, list[str]] = {}
            for tag in tags:
                category = tag.split(":", 1)[0]
                by_cat.setdefault(category, []).append(tag)
            for a, b in PRIMARY_CHAIN + STEERING_CHAIN:
                for frm in by_cat.get(a, []):
                    for to in by_cat.get(b, []):
                        chain_set.add((frm, to))

        # 3) 태그 루트(첫 번째 토큰) 기반 체인
        root_groups: dict[str, set[str]] = {}
        for tag in found:
            name = tag.split(":", 1)[1]
            name = re.sub(r"-(\d{3})$", "", name)
            root = name.split("-", 1)[0]
            if root:
                root_groups.setdefault(root, set()).add(tag)
        for tags in root_groups.values():
            by_cat: dict[str, list[str]] = {}
            for tag in tags:
                category = tag.split(":", 1)[0]
                by_cat.setdefault(category, []).append(tag)
            for a, b in PRIMARY_CHAIN + STEERING_CHAIN:
                for frm in by_cat.get(a, []):
                    for to in by_cat.get(b, []):
                        chain_set.add((frm, to))

        # 4) 명시적 @LINK 체인
        for link in explicit_links:
            chain_set.add((link["from"], link["to"]))

        return [
            {"from": frm, "to": to}
            for frm, to in sorted(chain_set)
        ]

    def verify(self, found: dict[str, list[str]], chains: list[dict[str, str]]):
        found_set = set(found.keys())
        linked_from: set[str] = set()
        linked_to: set[str] = set()

        for chain in chains:
            source = chain.get("from")
            target = chain.get("to")
            if source in found_set and target in found_set:
                linked_from.add(source)
                linked_to.add(target)
            else:
                missing_from = source if source in found_set else f"{source or 'unknown'}(?)"
                missing_to = target if target in found_set else f"{target or 'unknown'}(?)"
                self.broken_links.append((missing_from, missing_to))

        self.orphaned_tags = sorted(tag for tag in found_set if tag not in linked_from and tag not in linked_to)

    def update_index(self, found: dict[str, list[str]], chains: list[dict[str, str]]):
        # 카테고리 별 목록 업데이트(중복 제거)
        base_categories = {
            "SPEC": {"REQ": [], "DESIGN": [], "TASK": []},
            "STEERING": {"VISION": [], "STRUCT": [], "TECH": [], "ADR": []},
            "IMPLEMENTATION": {"FEATURE": [], "API": [], "TEST": [], "DATA": []},
            "QUALITY": {"PERF": [], "SEC": [], "DEBT": [], "TODO": []},
        }
        cat_to_group = {
            "REQ": "SPEC", "DESIGN": "SPEC", "TASK": "SPEC",
            "VISION": "STEERING", "STRUCT": "STEERING", "TECH": "STEERING", "ADR": "STEERING",
            "FEATURE": "IMPLEMENTATION", "API": "IMPLEMENTATION", "TEST": "IMPLEMENTATION", "DATA": "IMPLEMENTATION",
            "PERF": "QUALITY", "SEC": "QUALITY", "DEBT": "QUALITY", "TODO": "QUALITY",
        }

        locations = {}
        for tag, files in found.items():
            category = tag.split(":", 1)[0]
            group = cat_to_group.get(category)
            if group:
                base_categories[group].setdefault(category, []).append(tag)
            locations[tag] = sorted(set(files))

        self.index["categories"] = base_categories
        self.index["locations"] = locations
        self.index["traceability_chains"] = chains
        self.index["orphaned_tags"] = self.orphaned_tags
        self.index["statistics"] = {
            "total_tags": len(found),
            "complete_chains": max(0, len(chains) - len(self.broken_links)),
            "broken_links": len(self.broken_links),
            "coverage_percentage": 0,
        }
        total = max(1, len(found))
        coverage = round(100 * (total - len(self.orphaned_tags)) / total, 1)
        self.index["statistics"]["coverage_percentage"] = coverage
        self.index["last_updated"] = datetime.now().date().isoformat()

    def report(self, found: dict[str, list[str]], verbose: bool, strict: bool) -> int:
        click.echo("🏷️ TAG 추적성 검증 보고서")
        click.echo("=" * 50)
        click.echo(f"📊 총 TAG 수: {len(found)}")
        click.echo(f"🔗 끊어진 링크: {len(self.broken_links)}")
        click.echo(f"👻 고아 TAG: {len(self.orphaned_tags)}")
        if found:
            coverage = 100 - round(len(self.orphaned_tags) * 100 / len(found), 1)
            click.echo(f"✅ 추적성 커버리지: {coverage}%")
        if len(self.broken_links) == 0 and len(self.orphaned_tags) == 0:
            click.echo("✅ 모든 TAG 추적성 체인이 정상입니다!")
        else:
            if self.broken_links:
                click.echo("\n🔴 끊어진 추적성 체인:")
                for f, t in self.broken_links:
                    click.echo(f"  {f} → {t} (누락)")
            if self.orphaned_tags:
                click.echo("\n👻 고아 TAG 목록:")
                for tag in self.orphaned_tags:
                    click.echo(f"  {tag}")
        if verbose:
            click.echo("\n📂 TAG별 파일 위치:")
            for tag, files in sorted(found.items()):
                click.echo(f"  {tag}:")
                for fp in files:
                    click.echo(f"    - {fp}")
        if strict and (self.broken_links or self.orphaned_tags):
            return 1
        return 0


def main():
    import argparse
    parser = argparse.ArgumentParser(description="TAG 추적성 검증")
    parser.add_argument("--verbose", "-v", action="store_true", help="상세 출력")
    parser.add_argument("--project-root", "-p", default=".", help="프로젝트 루트 경로")
    parser.add_argument("--update", action="store_true", help="인덱스를 강제로 갱신 (기본: 자동 갱신)")
    parser.add_argument("--no-update", action="store_true", help="인덱스 갱신을 건너뜁니다")
    parser.add_argument("--strict", action="store_true", help="고아 TAG 또는 끊어진 링크가 있으면 종료 코드 1 반환")

    args = parser.parse_args()

    checker = TraceabilityChecker(args.project_root)
    checker.load_or_init_index()

    found = checker.scan_files_for_tags()

    explicit_links = checker.scan_files_for_links()
    stored_links = checker.index.get("traceability_chains", [])
    chains = checker.build_inferred_chains(found, explicit_links + stored_links)

    # 검증
    checker.verify(found, chains)

    do_update = True
    if args.no_update:
        do_update = False
    elif args.update:
        do_update = True

    if do_update:
        checker.update_index(found, chains)
        checker.save_index()

    return checker.report(found, args.verbose, args.strict)


if __name__ == "__main__":
    sys.exit(main())
