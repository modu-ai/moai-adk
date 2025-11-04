#!/usr/bin/env python3
"""
Automatic Report Cleanup Script

Manages .moai/reports/ directory:
- Archives old reports based on retention policy
- Creates backup of archived reports
- Updates manifest.json
- Generates cleanup summary

Usage:
    python3 .moai/scripts/cleanup_old_reports.py          # Dry run
    python3 .moai/scripts/cleanup_old_reports.py --execute # Execute cleanup
    python3 .moai/scripts/cleanup_old_reports.py --archive # Create backup

Retention Policy:
    - Default: 30 days
    - Important (spec-related): 90 days
    - Permanent tags: release, audit, critical
"""

import json
import sys
import shutil
from pathlib import Path
from datetime import datetime, timedelta
import argparse
from typing import Dict, List, Tuple


class ReportCleanup:
    def __init__(self, reports_dir: Path = None):
        """Initialize cleanup manager."""
        if reports_dir is None:
            reports_dir = Path(".moai/reports")

        self.reports_dir = reports_dir
        self.manifest_path = reports_dir / "manifest.json"
        self.archive_dir = reports_dir / "archive"
        self.manifest = self._load_manifest()

    def _load_manifest(self) -> Dict:
        """Load manifest.json."""
        if self.manifest_path.exists():
            with open(self.manifest_path, 'r', encoding='utf-8') as f:
                return json.load(f)
        return None

    def _save_manifest(self):
        """Save manifest.json."""
        if self.manifest:
            with open(self.manifest_path, 'w', encoding='utf-8') as f:
                json.dump(self.manifest, f, ensure_ascii=False, indent=2)

    def _parse_datetime(self, iso_string: str) -> datetime:
        """Parse ISO 8601 datetime string."""
        return datetime.fromisoformat(iso_string.replace("Z", "+00:00"))

    def _is_permanent(self, report: Dict) -> bool:
        """Check if report has permanent retention."""
        permanent_tags = self.manifest["retention_policy"]["permanent_tags"]
        report_tags = report.get("tags", [])
        return any(tag in report_tags for tag in permanent_tags)

    def _should_archive(self, report: Dict) -> bool:
        """Check if report should be archived based on retention policy."""
        if self._is_permanent(report):
            return False

        generated_at = self._parse_datetime(report["generated_at"])
        retention_days = report.get("retention_days", 30)
        cutoff_date = datetime.now(generated_at.tzinfo) - timedelta(days=retention_days)

        return generated_at < cutoff_date and not report.get("archived", False)

    def get_archivable_reports(self) -> List[Dict]:
        """Get list of reports that should be archived."""
        if not self.manifest:
            return []

        return [r for r in self.manifest["reports"] if self._should_archive(r)]

    def archive_reports(self, execute: bool = False) -> Tuple[int, int]:
        """Archive old reports.

        Args:
            execute: If True, actually move files. If False, dry run.

        Returns:
            Tuple of (archived_count, error_count)
        """
        archivable = self.get_archivable_reports()

        if not archivable:
            print("✅ No reports to archive")
            return 0, 0

        # Create archive directory if needed
        if execute and not self.archive_dir.exists():
            self.archive_dir.mkdir(parents=True, exist_ok=True)
            print(f"📁 Created archive directory: {self.archive_dir}")

        archived_count = 0
        error_count = 0

        for report in archivable:
            filename = report["filename"]
            source = self.reports_dir / filename
            dest = self.archive_dir / filename

            if not source.exists():
                print(f"⚠️  Source file not found: {filename}")
                error_count += 1
                continue

            try:
                if execute:
                    # Move file to archive
                    shutil.move(str(source), str(dest))
                    report["archived"] = True
                    print(f"✅ Archived: {filename}")
                else:
                    print(f"[DRY RUN] Would archive: {filename}")

                archived_count += 1

            except Exception as e:
                print(f"❌ Error archiving {filename}: {e}")
                error_count += 1

        if execute:
            self._save_manifest()

        return archived_count, error_count

    def generate_summary(self) -> str:
        """Generate cleanup summary report."""
        if not self.manifest:
            return "No manifest found"

        archivable = self.get_archivable_reports()

        summary = f"""
## Report Cleanup Summary

**Date**: {datetime.now().isoformat()}

### Statistics

- Total reports: {len(self.manifest['reports'])}
- Archived reports: {sum(1 for r in self.manifest['reports'] if r.get('archived', False))}
- Reports ready to archive: {len(archivable)}

### Reports by Type

"""

        by_type = {}
        for report in self.manifest["reports"]:
            rtype = report["type"]
            by_type[rtype] = by_type.get(rtype, 0) + 1

        for rtype, count in sorted(by_type.items()):
            summary += f"- {rtype}: {count}\n"

        if archivable:
            summary += "\n### Reports Ready for Archival\n\n"
            for report in archivable:
                generated = self._parse_datetime(report["generated_at"])
                days_old = (datetime.now(generated.tzinfo) - generated).days
                summary += f"- {report['filename']} ({days_old} days old)\n"

        return summary

    def save_cleanup_report(self, execute: bool = False) -> None:
        """Save cleanup report to file."""
        report_content = self.generate_summary()
        timestamp = datetime.now().strftime("%Y-%m-%d-%H%M%S")
        report_file = self.reports_dir / f"cleanup-report-{timestamp}.md"

        with open(report_file, 'w', encoding='utf-8') as f:
            f.write(report_content)

        print(f"\n📄 Cleanup report saved: {report_file}")

    def validate_archive(self) -> bool:
        """Validate archive directory integrity."""
        if not self.archive_dir.exists():
            print("✅ No archive directory (no archived reports)")
            return True

        # Count files
        archived_files = list(self.archive_dir.glob("*.md"))
        archived_in_manifest = sum(1 for r in self.manifest["reports"]
                                   if r.get("archived", False))

        if len(archived_files) != archived_in_manifest:
            print(f"⚠️  Mismatch: {len(archived_files)} files vs {archived_in_manifest} in manifest")
            return False

        print(f"✅ Archive valid: {len(archived_files)} files")
        return True


def main():
    """Command-line interface."""
    parser = argparse.ArgumentParser(
        description="Automatic report cleanup and archival",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Dry run (show what would be archived)
  python3 .moai/scripts/cleanup_old_reports.py

  # Execute cleanup
  python3 .moai/scripts/cleanup_old_reports.py --execute

  # Create backup and generate report
  python3 .moai/scripts/cleanup_old_reports.py --execute --report

  # Validate archive
  python3 .moai/scripts/cleanup_old_reports.py --validate
        """
    )

    parser.add_argument("--execute", action="store_true",
                       help="Execute cleanup (default is dry-run)")
    parser.add_argument("--report", action="store_true",
                       help="Generate cleanup report")
    parser.add_argument("--validate", action="store_true",
                       help="Validate archive integrity")

    args = parser.parse_args()

    # Check if .moai/reports exists
    reports_dir = Path(".moai/reports")
    if not reports_dir.exists():
        print(f"❌ Directory not found: {reports_dir}")
        sys.exit(1)

    cleanup = ReportCleanup(reports_dir)

    if args.validate:
        success = cleanup.validate_archive()
        sys.exit(0 if success else 1)

    # Perform cleanup
    print(f"\n🧹 Report Cleanup ({'EXECUTING' if args.execute else 'DRY RUN'})")
    print("=" * 60)

    archived, errors = cleanup.archive_reports(execute=args.execute)

    print(f"\n{'✅ Archived' if not errors else '⚠️  Processed'}: "
          f"{archived} reports{' (dry run)' if not args.execute else ''}")

    if errors:
        print(f"❌ Errors: {errors}")

    # Generate report if requested
    if args.report:
        cleanup.save_cleanup_report(execute=args.execute)

    # Print summary
    print(f"\n{cleanup.generate_summary()}")

    if not args.execute:
        print("\n💡 Run with --execute to apply cleanup")


if __name__ == "__main__":
    main()
