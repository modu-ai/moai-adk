"""@FEATURE:RELEASE-NOTES-001 Release notes converter

Converts sync-report files into structured release notes.
Supports version-based organization and changelog generation.

@REQ:RELEASE-NOTES-001 → @TASK:RELEASE-NOTES-001
"""

import logging
import re
from datetime import datetime
from pathlib import Path
from typing import Any

logger = logging.getLogger(__name__)


class ReleaseNotesConverter:
    """@FEATURE:RELEASE-NOTES-002 Convert sync reports to release notes"""

    def __init__(self, project_root: str):
        """Initialize release notes converter

        Args:
            project_root: Root directory of the project
        """
        self.project_root = Path(project_root)
        self.reports_dir = self.project_root / ".moai" / "reports"
        self.sync_report_path = self.reports_dir / "sync-report.md"

    def parse_sync_report(self) -> dict[str, Any] | None:
        """@TASK:RELEASE-NOTES-002 Parse sync-report.md file"""
        if not self.sync_report_path.exists():
            return None

        try:
            content = self.sync_report_path.read_text()
            report_data = {"date": None, "version": None, "changes": []}

            # Extract date from title
            date_match = re.search(r"# Sync Report - (\d{4}-\d{2}-\d{2})", content)
            if date_match:
                report_data["date"] = datetime.strptime(date_match.group(1), "%Y-%m-%d")

            # Extract version info
            version_match = re.search(r"- Current: (v?[\d.]+)", content)
            if version_match:
                report_data["version"] = version_match.group(1)

            # Extract changes - handle order independence
            if "## Changes Summary" in content:
                # Find the Changes Summary section
                start_idx = content.find("## Changes Summary")
                if start_idx != -1:
                    # Skip to end of the line after "## Changes Summary"
                    start_content = start_idx + len("## Changes Summary")
                    newline_after_title = content.find("\n", start_content)
                    if newline_after_title != -1:
                        start_content = newline_after_title + 1

                    # Find the next ## section or end of content
                    # Look for the next section beyond the current position
                    next_section = content.find("\n##", start_content + 1)
                    if next_section != -1:
                        changes_text = content[start_content:next_section]
                    else:
                        changes_text = content[start_content:]

                    changes_text = changes_text.strip()
                else:
                    changes_text = ""
            else:
                changes_text = ""

            if changes_text:
                # Find all @TAG entries using simpler approach
                lines = changes_text.split("\n")
                current_tag = None
                current_description = []

                for line in lines:
                    line = line.strip()
                    if line.startswith("### @"):
                        # Save previous tag if exists
                        if current_tag and current_description:
                            report_data["changes"].append(
                                {
                                    "tag": current_tag,
                                    "description": "\n".join(current_description),
                                }
                            )

                        # Start new tag
                        current_tag = line.replace("### ", "")
                        current_description = []
                    elif line and current_tag:
                        current_description.append(line)

                # Don't forget the last tag
                if current_tag and current_description:
                    report_data["changes"].append(
                        {
                            "tag": current_tag,
                            "description": "\n".join(current_description),
                        }
                    )

            return report_data

        except Exception as e:
            logger.error(f"Failed to parse sync report: {e}")
            return None

    def extract_version_info(self) -> dict[str, str]:
        """@TASK:RELEASE-NOTES-003 Extract version information"""
        version_info = {"current": "", "previous": ""}

        if not self.sync_report_path.exists():
            return version_info

        try:
            content = self.sync_report_path.read_text()

            # Extract current version
            current_match = re.search(r"- Current: (v?[\d.]+)", content)
            if current_match:
                version_info["current"] = current_match.group(1)

            # Extract previous version
            previous_match = re.search(r"- Previous: (v?[\d.]+)", content)
            if previous_match:
                version_info["previous"] = previous_match.group(1)

        except Exception as e:
            logger.error(f"Failed to extract version info: {e}")

        return version_info

    def categorize_changes(
        self, report_data: dict[str, Any] | None = None
    ) -> dict[str, list[dict[str, str]]]:
        """@TASK:RELEASE-NOTES-004 Categorize changes by TAG type"""
        if report_data is None:
            report_data = self.parse_sync_report()

        categories = {
            "FEATURE": [],
            "TEST": [],
            "TASK": [],
            "SEC": [],
            "API": [],
            "DOCS": [],
            "PERF": [],
            "OTHER": [],
        }

        if not report_data:
            return categories

        for change in report_data["changes"]:
            tag = change["tag"]

            # Extract category from tag (e.g., @FEATURE:DOCS-001 -> FEATURE)
            category_match = re.match(r"@(\w+):?", tag)
            if category_match:
                category = category_match.group(1)
                if category in categories:
                    categories[category].append(change)
                else:
                    categories["OTHER"].append(change)
            else:
                categories["OTHER"].append(change)

        return categories

    def generate_changelog(self) -> str:
        """@TASK:RELEASE-NOTES-005 Generate changelog format"""
        lines = ["# Release Notes", ""]

        report_data = self.parse_sync_report()
        if not report_data:
            lines.append("No release information available.")
            return "\n".join(lines)

        version = report_data.get("version", "Unknown")
        date = report_data.get("date")
        date_str = date.strftime("%Y-%m-%d") if date else "Unknown"

        lines.append(f"## [{version}] - {date_str}")
        lines.append("")

        # Categorize and format changes
        categorized = self.categorize_changes()

        category_labels = {
            "FEATURE": "✨ Features",
            "TEST": "🧪 Tests",
            "TASK": "📋 Tasks",
            "SEC": "🔒 Security",
            "API": "🔌 API",
            "DOCS": "📚 Documentation",
            "PERF": "⚡ Performance",
            "OTHER": "🔧 Other",
        }

        for category, label in category_labels.items():
            changes = categorized.get(category, [])
            if changes:
                lines.append(f"### {label}")
                lines.append("")

                for change in changes:
                    # Extract bullet points from description
                    description_lines = change["description"].split("\n")
                    for line in description_lines:
                        line = line.strip()
                        if line.startswith("- "):
                            lines.append(line)
                        elif line and not line.startswith("#"):
                            lines.append(f"- {line}")

                lines.append("")

        return "\n".join(lines)

    def get_version_timeline(self) -> list[dict[str, Any]]:
        """@TASK:RELEASE-NOTES-006 Get sorted version timeline"""
        versions = []

        # Scan all sync report files
        if self.reports_dir.exists():
            for report_file in self.reports_dir.glob("sync-report*.md"):
                try:
                    content = report_file.read_text()

                    # Extract version
                    version_match = re.search(r"- Current: (v?[\d.]+)", content)
                    if version_match:
                        version = version_match.group(1)

                        # Extract date
                        date_match = re.search(
                            r"# Sync Report - (\d{4}-\d{2}-\d{2})", content
                        )
                        date = None
                        if date_match:
                            date = datetime.strptime(date_match.group(1), "%Y-%m-%d")

                        versions.append(
                            {"version": version, "date": date, "file": report_file}
                        )

                except Exception as e:
                    logger.warning(f"Failed to parse {report_file}: {e}")

        # Sort by date (newest first), handle None dates
        def sort_key(version):
            if version["date"] is None:
                return datetime.min
            return version["date"]

        versions.sort(key=sort_key, reverse=True)
        return versions

    def generate_release_notes(self, docs_dir: str) -> None:
        """@TASK:RELEASE-NOTES-007 Generate release notes markdown file"""
        docs_path = Path(docs_dir)
        release_notes_file = docs_path / "release-notes.md"

        # Generate content
        content = self.generate_changelog()

        # Write to file
        release_notes_file.write_text(content)
