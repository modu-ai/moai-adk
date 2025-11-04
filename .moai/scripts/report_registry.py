#!/usr/bin/env python3
"""
Report Registry Manager

Central management system for .moai/reports/ directory.
Maintains manifest.json and provides utilities for report management.

Usage:
    python3 .moai/scripts/report_registry.py register <filename> <type> <purpose>
    python3 .moai/scripts/report_registry.py list
    python3 .moai/scripts/report_registry.py cleanup --days 30
    python3 .moai/scripts/report_registry.py validate
    python3 .moai/scripts/report_registry.py stats
"""

import json
import sys
from pathlib import Path
from datetime import datetime, timedelta
import argparse
from typing import Dict, List, Optional


class ReportRegistry:
    def __init__(self, reports_dir: Path = None):
        """Initialize report registry."""
        if reports_dir is None:
            reports_dir = Path(".moai/reports")

        self.reports_dir = reports_dir
        self.manifest_path = reports_dir / "manifest.json"
        self.manifest = self._load_manifest()

    def _load_manifest(self) -> Dict:
        """Load manifest.json or create empty template."""
        if self.manifest_path.exists():
            with open(self.manifest_path, 'r', encoding='utf-8') as f:
                return json.load(f)

        return {
            "version": "1.0",
            "generated_at": datetime.now().isoformat() + "Z",
            "description": "Central registry of all reports",
            "retention_policy": {
                "default_days": 30,
                "archive_days": 90,
                "permanent_tags": ["release", "audit", "critical"]
            },
            "naming_convention": "{type}-{purpose}-{YYYY-MM-DD-HHmm}.md",
            "report_types": {
                "sync": "동기화 결과 보고서",
                "analysis": "분석 보고서",
                "validation": "검증/검사 보고서",
                "audit": "감사 보고서",
                "implementation": "구현 결과 보고서",
                "test": "테스트 결과 보고서",
                "plan": "계획 및 전략 문서",
                "completion": "완료/최종 보고서",
                "regression": "회귀 분석 보고서"
            },
            "reports": [],
            "metadata": {
                "total_reports": 0,
                "reports_by_type": {},
                "storage_usage_bytes": 0,
                "last_cleanup": None,
                "next_cleanup": None
            }
        }

    def _save_manifest(self):
        """Save manifest.json."""
        with open(self.manifest_path, 'w', encoding='utf-8') as f:
            json.dump(self.manifest, f, ensure_ascii=False, indent=2)

    def _get_filename_id(self, filename: str) -> str:
        """Generate ID from filename (type-purpose-timestamp)."""
        return Path(filename).stem

    def register_report(self, filename: str, report_type: str, purpose: str,
                       spec_id: Optional[str] = None, tags: List[str] = None) -> bool:
        """Register a new report in the manifest."""
        report_path = self.reports_dir / filename

        if not report_path.exists():
            print(f"❌ File not found: {filename}")
            return False

        if report_type not in self.manifest["report_types"]:
            print(f"❌ Invalid report type: {report_type}")
            return False

        # Check if already registered
        if any(r["filename"] == filename for r in self.manifest["reports"]):
            print(f"⚠️  Report already registered: {filename}")
            return False

        # Get file stats
        stat = report_path.stat()
        mod_time = datetime.fromtimestamp(stat.st_mtime).isoformat() + "Z"

        # Create report entry
        report_entry = {
            "id": self._get_filename_id(filename),
            "filename": filename,
            "type": report_type,
            "purpose": purpose,
            "generated_at": mod_time,
            "generated_by": "alfred",
            "spec_id": spec_id,
            "status": "complete",
            "retention_days": 90 if spec_id else 30,
            "archived": False,
            "tags": tags or [],
            "size_bytes": stat.st_size
        }

        self.manifest["reports"].append(report_entry)
        self._update_metadata()
        self._save_manifest()

        print(f"✅ Registered: {filename} (type: {report_type})")
        return True

    def _update_metadata(self):
        """Update metadata section of manifest."""
        reports = self.manifest["reports"]

        # Count by type
        by_type = {}
        total_size = 0

        for report in reports:
            report_type = report["type"]
            by_type[report_type] = by_type.get(report_type, 0) + 1
            total_size += report.get("size_bytes", 0)

        self.manifest["metadata"] = {
            "total_reports": len(reports),
            "reports_by_type": by_type,
            "storage_usage_bytes": total_size,
            "last_cleanup": datetime.now().isoformat() + "Z",
            "next_cleanup": (datetime.now() + timedelta(days=7)).isoformat() + "Z"
        }

    def list_reports(self, report_type: Optional[str] = None, archived: bool = False) -> None:
        """List all reports, optionally filtered by type."""
        reports = self.manifest["reports"]

        if report_type:
            reports = [r for r in reports if r["type"] == report_type]

        reports = [r for r in reports if r["archived"] == archived]

        if not reports:
            print("No reports found.")
            return

        print(f"\n📋 Reports ({len(reports)} total)")
        print("=" * 100)
        print(f"{'Type':<12} | {'Purpose':<40} | {'Date':<19} | {'Size':<10}")
        print("-" * 100)

        for report in sorted(reports, key=lambda x: x["generated_at"], reverse=True):
            size = report.get("size_bytes", 0)
            size_kb = f"{size/1024:.1f}KB" if size else "N/A"

            print(f"{report['type']:<12} | {report['purpose']:<40} | "
                  f"{report['generated_at']:<19} | {size_kb:<10}")

        print("=" * 100)

        # Print statistics
        stats = self.manifest["metadata"]
        print(f"\n📊 Statistics:")
        print(f"  Total reports: {stats['total_reports']}")
        print(f"  Storage: {stats['storage_usage_bytes']/1024:.1f}KB")
        print(f"  By type: {stats['reports_by_type']}")

    def cleanup_old_reports(self, days: int = 30, dry_run: bool = True) -> None:
        """Mark old reports as archived based on retention policy."""
        cutoff_date = datetime.now() - timedelta(days=days)

        archived_count = 0
        for report in self.manifest["reports"]:
            generated_at = datetime.fromisoformat(report["generated_at"].replace("Z", "+00:00"))
            retention = report.get("retention_days", 30)
            cutoff = datetime.now() - timedelta(days=retention)

            # Don't archive permanent reports
            if any(tag in report.get("tags", []) for tag in
                   self.manifest["retention_policy"]["permanent_tags"]):
                continue

            if generated_at < cutoff and not report["archived"]:
                if not dry_run:
                    report["archived"] = True
                archived_count += 1
                print(f"{'[DRY RUN] ' if dry_run else ''}Archive: {report['filename']}")

        if not dry_run:
            self._update_metadata()
            self._save_manifest()

        print(f"\n✅ Would archive {archived_count} reports" if dry_run
              else f"✅ Archived {archived_count} reports")

    def validate_manifest(self) -> bool:
        """Validate manifest integrity."""
        errors = []

        # Check required fields
        for report in self.manifest["reports"]:
            required = ["id", "filename", "type", "purpose", "generated_at"]
            for field in required:
                if field not in report:
                    errors.append(f"Missing field '{field}' in report: {report.get('filename', 'unknown')}")

        # Check file existence
        for report in self.manifest["reports"]:
            if not (self.reports_dir / report["filename"]).exists():
                errors.append(f"File not found: {report['filename']}")

        if errors:
            print("❌ Validation failed:")
            for error in errors:
                print(f"  - {error}")
            return False

        print("✅ Manifest validation passed")
        return True

    def print_stats(self) -> None:
        """Print detailed statistics."""
        stats = self.manifest["metadata"]

        print("\n📊 Report Registry Statistics")
        print("=" * 60)
        print(f"Total reports: {stats['total_reports']}")
        print(f"Total storage: {stats['storage_usage_bytes']/1024:.1f}KB")
        print(f"Last cleanup: {self.manifest['metadata'].get('last_cleanup', 'Never')}")
        print(f"Next cleanup: {self.manifest['metadata'].get('next_cleanup', 'N/A')}")

        print(f"\nReports by type:")
        for rtype, count in stats['reports_by_type'].items():
            print(f"  - {rtype}: {count}")

        # Oldest and newest
        if self.manifest["reports"]:
            sorted_reports = sorted(self.manifest["reports"],
                                   key=lambda x: x["generated_at"])
            print(f"\nOldest report: {sorted_reports[0]['filename']} "
                  f"({sorted_reports[0]['generated_at']})")
            print(f"Newest report: {sorted_reports[-1]['filename']} "
                  f"({sorted_reports[-1]['generated_at']})")


def main():
    """Command-line interface."""
    parser = argparse.ArgumentParser(
        description="Manage .moai/reports/ directory with central registry"
    )

    subparsers = parser.add_subparsers(dest="command", help="Command to execute")

    # Register command
    register_parser = subparsers.add_parser("register", help="Register a new report")
    register_parser.add_argument("filename", help="Report filename")
    register_parser.add_argument("type", help="Report type (sync, analysis, validation, etc.)")
    register_parser.add_argument("purpose", help="Report purpose/description")
    register_parser.add_argument("--spec-id", help="Associated SPEC ID")
    register_parser.add_argument("--tags", nargs="+", help="Tags for the report")

    # List command
    list_parser = subparsers.add_parser("list", help="List all reports")
    list_parser.add_argument("--type", help="Filter by report type")
    list_parser.add_argument("--archived", action="store_true", help="Show archived reports")

    # Cleanup command
    cleanup_parser = subparsers.add_parser("cleanup", help="Mark old reports as archived")
    cleanup_parser.add_argument("--days", type=int, default=30,
                               help="Days threshold for cleanup")
    cleanup_parser.add_argument("--execute", action="store_true",
                               help="Execute cleanup (default is dry-run)")

    # Validate command
    subparsers.add_parser("validate", help="Validate manifest integrity")

    # Stats command
    subparsers.add_parser("stats", help="Print detailed statistics")

    args = parser.parse_args()

    # Check if .moai/reports exists
    reports_dir = Path(".moai/reports")
    if not reports_dir.exists():
        print(f"❌ Directory not found: {reports_dir}")
        sys.exit(1)

    registry = ReportRegistry(reports_dir)

    if args.command == "register":
        success = registry.register_report(
            args.filename, args.type, args.purpose,
            spec_id=args.spec_id, tags=args.tags
        )
        sys.exit(0 if success else 1)

    elif args.command == "list":
        registry.list_reports(report_type=args.type, archived=args.archived)

    elif args.command == "cleanup":
        registry.cleanup_old_reports(days=args.days, dry_run=not args.execute)

    elif args.command == "validate":
        success = registry.validate_manifest()
        sys.exit(0 if success else 1)

    elif args.command == "stats":
        registry.print_stats()

    else:
        parser.print_help()


if __name__ == "__main__":
    main()
