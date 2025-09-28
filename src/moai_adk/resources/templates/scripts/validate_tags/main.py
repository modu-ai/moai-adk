#!/usr/bin/env python3
"""
Main entry point for tag validation
"""

import sys
from pathlib import Path

from .scanner import scan_project_files
from .validator import TagValidator
from .checker import find_orphan_tags, find_broken_links
from .traceability import validate_traceability_chains
from .reporter import generate_health_report
from .exporter import create_report_data
from .database import save_report_to_sqlite


def main():
    """Main entry point for tag validation"""
    try:
        # Get project root
        if len(sys.argv) > 1:
            project_root = Path(sys.argv[1]).resolve()
        else:
            project_root = Path.cwd()

        if not project_root.exists():
            print(f"❌ Project directory not found: {project_root}")
            sys.exit(1)

        print(f"🔍 Validating @TAG system in: {project_root}")

        # Initialize validator
        validator = TagValidator(project_root)

        # Scan for tags
        all_tags = scan_project_files(project_root)
        if not all_tags:
            print("ℹ️  No tags found in project")
            sys.exit(0)

        print(f"📊 Found {len(all_tags)} tags")

        # Build tag index
        tag_index = validator.build_tag_index(all_tags)

        # Validate formats
        format_violations = []
        for tag in all_tags:
            is_valid, error = validator.validate_tag_format(tag)
            if not is_valid:
                format_violations.append(f"{tag.file_path}:{tag.line_number} - {error}")

        # Find issues
        orphan_tags = find_orphan_tags(tag_index)
        broken_links = find_broken_links(tag_index)
        chain_violations = validate_traceability_chains(tag_index, validator.traceability_chains)

        # Generate report
        report = generate_health_report(
            all_tags, tag_index, format_violations,
            orphan_tags, broken_links, chain_violations
        )

        # Print summary
        print(f"\n📋 Summary: {report.total_tags} tags, {report.valid_tags} valid, {report.quality_score:.1%} quality")
        if report.issues:
            print(f"⚠️  Issues: {len(report.issues)}")

        # Save to database
        report_data = create_report_data(report)
        report_file = project_root / ".moai/reports/tag_validation.db"
        save_report_to_sqlite(report, report_file, report_data)

        # Exit with appropriate code
        sys.exit(0 if report.quality_score >= 0.6 else 1)

    except Exception as error:
        print(f"❌ Tag validation failed: {error}")
        sys.exit(1)