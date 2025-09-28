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
import time
import logging
import concurrent.futures
import resource
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional

import click

# Import our performance cache module
try:
    from performance_cache import PerformanceCache
except ImportError:
    # Fallback if performance_cache is not available
    PerformanceCache = None

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

        # Performance enhancements
        self.thread_count = 4  # Default thread count
        self.performance_cache = PerformanceCache(self.project_root / ".moai" / "cache") if PerformanceCache else None
        self.performance_metrics = {
            'scan_duration': 0.0,
            'files_processed': 0,
            'cache_hit_rate': 0.0,
            'thread_count_used': self.thread_count,
            'memory_usage_mb': 0.0
        }

        # Setup logging
        logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

    def _get_memory_usage_mb(self) -> float:
        """Get current memory usage in MB (cross-platform)"""
        try:
            # Using resource module (Unix-like systems)
            usage = resource.getrusage(resource.RUSAGE_SELF)
            # maxrss is in KB on Linux, bytes on macOS
            max_rss = usage.ru_maxrss
            if sys.platform == 'darwin':
                # macOS reports in bytes
                return max_rss / (1024 * 1024)
            else:
                # Linux reports in KB
                return max_rss / 1024
        except Exception:
            return 0.0

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
                    "IMPLEMENTATION": {
                        "FEATURE": [],
                        "API": [],
                        "TEST": [],
                        "DATA": [],
                    },
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
        start_time = time.time()
        tag_pattern = r"@([A-Z]+):([A-Z0-9-]+)"
        found: dict[str, list[str]] = {}
        exts = [".md", ".py", ".js", ".ts", ".yaml", ".yml", ".json"]
        files_processed = 0

        for ext in exts:
            for file_path in self.project_root.rglob(f"*{ext}"):
                # 숨김 디렉토리 제외(.git 등) 단, .claude, .moai는 허용
                if any(
                    part.startswith(".") and part not in [".claude", ".moai"]
                    for part in file_path.parts
                ):
                    continue
                try:
                    # Try to get content from cache first
                    content = None
                    if self.performance_cache:
                        content = self.performance_cache.get_cached_content(file_path)

                    if content is None:
                        # Cache miss - read from file
                        content = file_path.read_text(encoding="utf-8")
                        # Cache the content
                        if self.performance_cache:
                            self.performance_cache.cache_file_content(file_path, content)

                    files_processed += 1

                except Exception:
                    continue
                for cat, tid in re.findall(tag_pattern, content):
                    tag = f"{cat}:{tid}"
                    found.setdefault(tag, []).append(str(file_path))

        # Update performance metrics
        scan_duration = time.time() - start_time
        self.performance_metrics.update({
            'scan_duration': scan_duration,
            'files_processed': files_processed,
            'memory_usage_mb': self._get_memory_usage_mb()
        })

        if self.performance_cache:
            cache_stats = self.performance_cache.get_cache_stats()
            self.performance_metrics['cache_hit_rate'] = cache_stats['hit_rate_percent']

        # Log performance data
        logging.info(f"Performance scan metrics (SEQUENTIAL) - Duration: {scan_duration:.3f}s, Files: {files_processed}, "
                    f"Cache hit rate: {self.performance_metrics['cache_hit_rate']:.1f}%, "
                    f"Memory: {self.performance_metrics['memory_usage_mb']:.1f}MB")

        return found

    def scan_files_for_links(self) -> list[dict[str, str]]:
        """@LINK:CAT:ID->CAT:ID 형식의 명시적 링크 스캔"""
        link_pattern = r"@LINK:([A-Z]+:[A-Z0-9-]+)->([A-Z]+:[A-Z0-9-]+)"
        links: list[dict[str, str]] = []
        exts = [".md", ".py", ".js", ".ts", ".yaml", ".yml", ".json"]
        for ext in exts:
            for file_path in self.project_root.rglob(f"*{ext}"):
                if any(
                    part.startswith(".") and part not in [".claude", ".moai"]
                    for part in file_path.parts
                ):
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

        return [{"from": frm, "to": to} for frm, to in sorted(chain_set)]

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
                missing_from = (
                    source if source in found_set else f"{source or 'unknown'}(?)"
                )
                missing_to = (
                    target if target in found_set else f"{target or 'unknown'}(?)"
                )
                self.broken_links.append((missing_from, missing_to))

        self.orphaned_tags = sorted(
            tag for tag in found_set if tag not in linked_from and tag not in linked_to
        )

    def update_index(self, found: dict[str, list[str]], chains: list[dict[str, str]]):
        # 카테고리 별 목록 업데이트(중복 제거)
        base_categories = {
            "SPEC": {"REQ": [], "DESIGN": [], "TASK": []},
            "STEERING": {"VISION": [], "STRUCT": [], "TECH": [], "ADR": []},
            "IMPLEMENTATION": {"FEATURE": [], "API": [], "TEST": [], "DATA": []},
            "QUALITY": {"PERF": [], "SEC": [], "DEBT": [], "TODO": []},
        }
        cat_to_group = {
            "REQ": "SPEC",
            "DESIGN": "SPEC",
            "TASK": "SPEC",
            "VISION": "STEERING",
            "STRUCT": "STEERING",
            "TECH": "STEERING",
            "ADR": "STEERING",
            "FEATURE": "IMPLEMENTATION",
            "API": "IMPLEMENTATION",
            "TEST": "IMPLEMENTATION",
            "DATA": "IMPLEMENTATION",
            "PERF": "QUALITY",
            "SEC": "QUALITY",
            "DEBT": "QUALITY",
            "TODO": "QUALITY",
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

    def set_thread_count(self, count: int) -> None:
        """Set the number of threads for parallel processing"""
        self.thread_count = max(1, min(count, 16))  # Limit between 1 and 16
        self.performance_metrics['thread_count_used'] = self.thread_count

    def _scan_file_for_tags(self, file_path: Path) -> Dict[str, List[str]]:
        """Scan a single file for tags. Helper method for parallel processing."""
        tag_pattern = r"@([A-Z]+):([A-Z0-9-]+)"
        found: Dict[str, List[str]] = {}

        try:
            # Try to get content from cache first
            content = None
            if self.performance_cache:
                content = self.performance_cache.get_cached_content(file_path)

            if content is None:
                # Cache miss - read from file
                content = file_path.read_text(encoding="utf-8")
                # Cache the content
                if self.performance_cache:
                    self.performance_cache.cache_file_content(file_path, content)

            for cat, tid in re.findall(tag_pattern, content):
                tag = f"{cat}:{tid}"
                found.setdefault(tag, []).append(str(file_path))

        except Exception:
            pass  # Skip files that can't be read

        return found

    def parallel_scan_files_for_tags(self) -> Dict[str, List[str]]:
        """
        Parallel version of scan_files_for_tags using thread pool.
        Should be faster than sequential scanning on multi-core systems.
        Automatically adjusts thread count based on file count.
        """
        start_time = time.time()

        # Collect all files to scan
        exts = [".md", ".py", ".js", ".ts", ".yaml", ".yml", ".json"]
        files_to_scan = []

        for ext in exts:
            for file_path in self.project_root.rglob(f"*{ext}"):
                # Skip hidden directories except .claude and .moai
                if any(
                    part.startswith(".") and part not in [".claude", ".moai"]
                    for part in file_path.parts
                ):
                    continue
                files_to_scan.append(file_path)

        # Adjust thread count based on file count
        # For small file counts, use fewer threads to avoid overhead
        optimal_threads = min(self.thread_count, max(1, len(files_to_scan) // 10))

        # Parallel processing
        found: Dict[str, List[str]] = {}

        with concurrent.futures.ThreadPoolExecutor(max_workers=optimal_threads) as executor:
            # Submit all file scanning tasks
            future_to_file = {
                executor.submit(self._scan_file_for_tags, file_path): file_path
                for file_path in files_to_scan
            }

            # Collect results
            for future in concurrent.futures.as_completed(future_to_file):
                file_result = future.result()
                # Merge results
                for tag, paths in file_result.items():
                    found.setdefault(tag, []).extend(paths)

        # Update performance metrics
        scan_duration = time.time() - start_time
        self.performance_metrics.update({
            'scan_duration': scan_duration,
            'files_processed': len(files_to_scan),
            'memory_usage_mb': self._get_memory_usage_mb()
        })

        if self.performance_cache:
            cache_stats = self.performance_cache.get_cache_stats()
            self.performance_metrics['cache_hit_rate'] = cache_stats['hit_rate_percent']

        # Log performance data
        logging.info(f"Performance metrics (PARALLEL) - Duration: {scan_duration:.3f}s, Files: {len(files_to_scan)}, "
                    f"Threads: {optimal_threads}/{self.thread_count}, Cache hit rate: {self.performance_metrics['cache_hit_rate']:.1f}%, "
                    f"Memory: {self.performance_metrics['memory_usage_mb']:.1f}MB")

        return found

    def get_changed_files_since_last_scan(self) -> List[str]:
        """Get list of files that have changed since last scan"""
        if not self.performance_cache:
            return []  # Return empty list if no cache available

        # Get all potential files
        exts = [".md", ".py", ".js", ".ts", ".yaml", ".yml", ".json"]
        all_files = []

        for ext in exts:
            for file_path in self.project_root.rglob(f"*{ext}"):
                if any(
                    part.startswith(".") and part not in [".claude", ".moai"]
                    for part in file_path.parts
                ):
                    continue
                all_files.append(file_path)

        changed_files = self.performance_cache.get_changed_files_since_last_scan(all_files)
        return list(changed_files)

    def incremental_scan_files_for_tags(self, changed_files: List[str]) -> Dict[str, List[str]]:
        """Scan only the changed files for tags"""
        found: Dict[str, List[str]] = {}

        for file_path_str in changed_files:
            file_path = Path(file_path_str)
            if file_path.exists():
                file_result = self._scan_file_for_tags(file_path)
                for tag, paths in file_result.items():
                    found.setdefault(tag, []).extend(paths)

        return found

    def incremental_scan(self) -> Dict[str, List[str]]:
        """Perform incremental scan of changed files only"""
        changed_files = self.get_changed_files_since_last_scan()
        if not changed_files:
            return {}
        return self.incremental_scan_files_for_tags(changed_files)

    def save_scan_timestamp(self) -> None:
        """Save timestamp of current scan for incremental scanning"""
        if self.performance_cache:
            # Get all scannable files
            exts = [".md", ".py", ".js", ".ts", ".yaml", ".yml", ".json"]
            all_files = []

            for ext in exts:
                for file_path in self.project_root.rglob(f"*{ext}"):
                    if any(
                        part.startswith(".") and part not in [".claude", ".moai"]
                        for part in file_path.parts
                    ):
                        continue
                    all_files.append(file_path)

            self.performance_cache.save_scan_timestamp(all_files)

    def get_performance_metrics(self) -> Dict[str, float]:
        """Get current performance metrics"""
        return self.performance_metrics.copy()

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
    parser.add_argument(
        "--update", action="store_true", help="인덱스를 강제로 갱신 (기본: 자동 갱신)"
    )
    parser.add_argument(
        "--no-update", action="store_true", help="인덱스 갱신을 건너뜁니다"
    )
    parser.add_argument(
        "--strict",
        action="store_true",
        help="고아 TAG 또는 끊어진 링크가 있으면 종료 코드 1 반환",
    )
    parser.add_argument(
        "--parallel",
        action="store_true",
        help="병렬 스캔 사용 (대용량 프로젝트에 권장)",
    )
    parser.add_argument(
        "--threads",
        type=int,
        default=4,
        help="병렬 스캔 시 사용할 스레드 수 (기본: 4)",
    )

    args = parser.parse_args()

    checker = TraceabilityChecker(args.project_root)
    checker.load_or_init_index()

    # Set thread count if specified
    if args.threads:
        checker.set_thread_count(args.threads)

    # Use parallel or sequential scanning
    if args.parallel:
        found = checker.parallel_scan_files_for_tags()
    else:
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
