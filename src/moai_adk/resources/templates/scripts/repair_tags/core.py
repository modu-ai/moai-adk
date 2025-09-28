#!/usr/bin/env python3
# @TASK:TAG-REPAIR-CORE-001
"""
TAG Repairer Core Module

Main orchestration class for TAG repair system.
Handles configuration and coordinates other modules.
"""

from pathlib import Path
from .scanner import TagScanner
from .analyzer import TagAnalyzer
from .generator import RepairGenerator
from .updater import IndexUpdater


class TagRepairer:
    """Main TAG repair orchestrator."""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.moai_path = project_root / ".moai"
        self.indexes_path = self.moai_path / "indexes"

        # Initialize components
        self.scanner = TagScanner(project_root)
        self.analyzer = TagAnalyzer()
        self.generator = RepairGenerator()
        self.updater = IndexUpdater(self.indexes_path)

        # 16-Core TAG categories
        self.tag_categories = {
            "SPEC": ["REQ", "SPEC", "DESIGN", "TASK"],
            "STEERING": ["VISION", "STRUCT", "TECH", "ADR"],
            "IMPLEMENTATION": ["FEATURE", "API", "TEST", "DATA"],
            "QUALITY": ["PERF", "SEC", "DEBT", "TODO"],
        }

        # Traceability chains
        self.traceability_chains = {
            "primary": ["REQ", "DESIGN", "TASK", "TEST"],
            "steering": ["VISION", "STRUCT", "TECH", "ADR"],
            "implementation": ["FEATURE", "API", "DATA"],
            "quality": ["PERF", "SEC", "DEBT", "TODO"],
        }

    def auto_repair_tags(self, dry_run: bool = True) -> bool:
        """Execute automatic TAG repair process."""
        try:
            # Scan all tags
            all_tags = self.scanner.scan_project_tags()

            # Analyze integrity
            analysis = self.analyzer.analyze_tag_integrity(all_tags)

            if not analysis["has_issues"]:
                return True

            # Generate repairs
            self.generator.generate_repair_preview(
                analysis, all_tags, self.tag_categories
            )

            # Apply repairs if not dry run
            if not dry_run:
                self.updater.update_traceability_index()

            return True
        except Exception:
            return False