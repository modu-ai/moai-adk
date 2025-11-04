#!/usr/bin/env python3
"""
Report Registry Metrics & Monitoring System

Collects and analyzes metrics from report registry:
- Report generation frequency
- Storage usage trends
- Report aging analysis
- Type distribution
- Retention policy effectiveness

Output: .moai/metrics/report_metrics.json and analysis reports

Usage:
    python3 .moai/scripts/report_metrics.py            # Collect metrics
    python3 .moai/scripts/report_metrics.py --analyze  # Generate analysis
    python3 .moai/scripts/report_metrics.py --trend    # Show trends
"""

import json
import sys
from pathlib import Path
from datetime import datetime, timedelta
from collections import defaultdict
from statistics import mean, median
import argparse
from typing import Dict, List


class ReportMetrics:
    def __init__(self, reports_dir: Path = None, metrics_dir: Path = None):
        """Initialize metrics collector."""
        if reports_dir is None:
            reports_dir = Path(".moai/reports")
        if metrics_dir is None:
            metrics_dir = Path(".moai/metrics")

        self.reports_dir = reports_dir
        self.metrics_dir = metrics_dir
        self.manifest_path = reports_dir / "manifest.json"
        self.metrics_file = metrics_dir / "report_metrics.json"

        self.manifest = self._load_manifest()

    def _load_manifest(self) -> Dict:
        """Load manifest.json."""
        if self.manifest_path.exists():
            with open(self.manifest_path, 'r', encoding='utf-8') as f:
                return json.load(f)
        return None

    def _ensure_metrics_dir(self):
        """Ensure metrics directory exists."""
        self.metrics_dir.mkdir(parents=True, exist_ok=True)

    def _parse_datetime(self, iso_string: str) -> datetime:
        """Parse ISO 8601 datetime string."""
        return datetime.fromisoformat(iso_string.replace("Z", "+00:00"))

    def collect_metrics(self) -> Dict:
        """Collect current metrics from manifest."""
        if not self.manifest:
            return {}

        reports = self.manifest["reports"]
        now = datetime.now(datetime.now().astimezone().tzinfo)

        # Basic statistics
        metrics = {
            "timestamp": now.isoformat(),
            "total_reports": len(reports),
            "archived_reports": sum(1 for r in reports if r.get("archived", False)),
            "active_reports": sum(1 for r in reports if not r.get("archived", False)),
            "total_storage_bytes": sum(r.get("size_bytes", 0) for r in reports),
            "archived_storage_bytes": sum(
                r.get("size_bytes", 0) for r in reports if r.get("archived", False)
            )
        }

        # By type distribution
        by_type = defaultdict(int)
        by_type_storage = defaultdict(int)
        for report in reports:
            rtype = report["type"]
            by_type[rtype] += 1
            by_type_storage[rtype] += report.get("size_bytes", 0)

        metrics["reports_by_type"] = dict(by_type)
        metrics["storage_by_type"] = dict(by_type_storage)

        # Age analysis
        ages = []
        for report in reports:
            if not report.get("archived", False):
                generated = self._parse_datetime(report["generated_at"])
                age_days = (now - generated).days
                ages.append(age_days)

        if ages:
            metrics["age_statistics"] = {
                "min_days": min(ages),
                "max_days": max(ages),
                "mean_days": round(mean(ages), 1),
                "median_days": median(ages),
                "total_days": sum(ages)
            }

        # Retention analysis
        retention_distribution = defaultdict(int)
        for report in reports:
            if not report.get("archived", False):
                retention = report.get("retention_days", 30)
                retention_distribution[retention] += 1

        metrics["retention_distribution"] = dict(retention_distribution)

        # Historical data
        self._ensure_metrics_dir()
        if self.metrics_file.exists():
            with open(self.metrics_file, 'r') as f:
                historical = json.load(f)
                if "history" not in historical:
                    historical["history"] = []
                historical["history"].append(metrics)
                # Keep last 52 weeks (1 year)
                if len(historical["history"]) > 52:
                    historical["history"] = historical["history"][-52:]
                historical["latest"] = metrics
        else:
            historical = {
                "latest": metrics,
                "history": [metrics]
            }

        return historical

    def save_metrics(self, metrics: Dict) -> None:
        """Save metrics to file."""
        self._ensure_metrics_dir()
        with open(self.metrics_file, 'w') as f:
            json.dump(metrics, f, ensure_ascii=False, indent=2)

        print(f"✅ Metrics saved to {self.metrics_file}")

    def generate_analysis_report(self, metrics: Dict) -> str:
        """Generate analysis report from metrics."""
        latest = metrics.get("latest", {})

        report = f"""# Report Metrics Analysis

**Generated**: {latest.get('timestamp', 'Unknown')}

## Summary

- **Total Reports**: {latest.get('total_reports', 0)}
- **Active Reports**: {latest.get('active_reports', 0)}
- **Archived Reports**: {latest.get('archived_reports', 0)}
- **Total Storage**: {self._format_bytes(latest.get('total_storage_bytes', 0))}
- **Archived Storage**: {self._format_bytes(latest.get('archived_storage_bytes', 0))}

## Distribution by Type

"""

        for rtype, count in sorted(latest.get("reports_by_type", {}).items()):
            storage = latest.get("storage_by_type", {}).get(rtype, 0)
            storage_str = self._format_bytes(storage)
            report += f"- **{rtype}**: {count} reports ({storage_str})\n"

        # Age statistics
        age_stats = latest.get("age_statistics", {})
        if age_stats:
            report += f"""
## Age Statistics (Active Reports)

- **Oldest**: {age_stats.get('max_days', 0)} days
- **Newest**: {age_stats.get('min_days', 0)} days
- **Average**: {age_stats.get('mean_days', 0)} days
- **Median**: {age_stats.get('median_days', 0)} days

"""

        # Retention analysis
        retention = latest.get("retention_distribution", {})
        if retention:
            report += "## Retention Distribution\n\n"
            for days, count in sorted(retention.items()):
                report += f"- {days} days: {count} reports\n"

        # Recommendations
        report += "\n## Recommendations\n\n"

        total = latest.get("total_reports", 0)
        archived = latest.get("archived_reports", 0)

        if total > 100:
            report += "- ⚠️  Consider archival: More than 100 reports\n"
        if archived == 0 and total > 50:
            report += "- 💡 Consider setting up automatic cleanup\n"
        if latest.get("total_storage_bytes", 0) > 500 * 1024:  # 500 KB
            report += "- 💡 Storage exceeds 500 KB, consider cleanup\n"

        report += "\n---\n\nGenerated with Claude Code 🤖\n"

        return report

    def show_trends(self, metrics: Dict) -> None:
        """Display trends over time."""
        history = metrics.get("history", [])

        if len(history) < 2:
            print("Not enough historical data for trend analysis")
            return

        latest = history[-1]
        previous = history[-2]

        print("\n📈 Trends")
        print("=" * 60)

        # Calculate changes
        total_change = latest.get("total_reports", 0) - previous.get("total_reports", 0)
        storage_change = latest.get("total_storage_bytes", 0) - previous.get("total_storage_bytes", 0)

        print(f"Total Reports: {latest.get('total_reports', 0)} "
              f"({total_change:+d})")
        print(f"Storage: {self._format_bytes(latest.get('total_storage_bytes', 0))} "
              f"({self._format_bytes(storage_change):+s})")

        # Type trends
        latest_types = latest.get("reports_by_type", {})
        previous_types = previous.get("reports_by_type", {})

        print("\nBy Type:")
        for rtype in set(list(latest_types.keys()) + list(previous_types.keys())):
            latest_count = latest_types.get(rtype, 0)
            previous_count = previous_types.get(rtype, 0)
            change = latest_count - previous_count
            print(f"  - {rtype}: {latest_count} ({change:+d})")

    @staticmethod
    def _format_bytes(bytes_val: int) -> str:
        """Format bytes to human-readable string."""
        for unit in ["B", "KB", "MB", "GB"]:
            if bytes_val < 1024:
                return f"{bytes_val:.1f}{unit}"
            bytes_val /= 1024
        return f"{bytes_val:.1f}TB"


def main():
    """Command-line interface."""
    parser = argparse.ArgumentParser(
        description="Report metrics collection and analysis"
    )

    parser.add_argument("--analyze", action="store_true",
                       help="Generate analysis report")
    parser.add_argument("--trend", action="store_true",
                       help="Show trends over time")
    parser.add_argument("--save", action="store_true", default=True,
                       help="Save metrics to file (default: true)")

    args = parser.parse_args()

    # Check directories
    reports_dir = Path(".moai/reports")
    if not reports_dir.exists():
        print(f"❌ Directory not found: {reports_dir}")
        sys.exit(1)

    metrics_mgr = ReportMetrics()

    # Collect metrics
    print("📊 Collecting metrics...")
    metrics = metrics_mgr.collect_metrics()

    if not metrics:
        print("❌ Failed to collect metrics")
        sys.exit(1)

    latest = metrics.get("latest", {})

    # Save metrics
    if args.save:
        metrics_mgr.save_metrics(metrics)

    # Print summary
    print(f"\n✅ Metrics collected at {latest.get('timestamp', 'Unknown')}")
    print(f"   Total reports: {latest.get('total_reports', 0)}")
    print(f"   Storage: {metrics_mgr._format_bytes(latest.get('total_storage_bytes', 0))}")

    # Generate analysis if requested
    if args.analyze:
        analysis = metrics_mgr.generate_analysis_report(metrics)
        print(f"\n{analysis}")

    # Show trends if requested
    if args.trend:
        metrics_mgr.show_trends(metrics)


if __name__ == "__main__":
    main()
