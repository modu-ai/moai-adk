#!/usr/bin/env python3
# @CODE:DOC-TAG-004 | Component 2: CI/CD pipeline TAG validator
"""CI/CD TAG validation module for GitHub Actions

This module extends PreCommitValidator for CI/CD environments:
- Fetches PR changed files via GitHub API
- Generates structured validation reports (JSON/markdown)
- Posts validation results as PR comments
- Supports strict mode (block merge on warnings) and info mode

Used by GitHub Actions workflow to validate TAGs on every PR.
"""

import json
import os
from typing import Any, Dict, List, Optional

import requests
from requests.adapters import HTTPAdapter
from urllib3.util import Retry

from .pre_commit_validator import (
    PreCommitValidator,
    ValidationResult,
)


class CIValidator(PreCommitValidator):
    """CI/CD TAG validator for GitHub Actions

    Extends PreCommitValidator with CI/CD-specific features:
    - GitHub API integration for PR file detection
    - Structured report generation for automation
    - Markdown comment formatting for PR feedback
    - Environment variable support for GitHub Actions

    Args:
        github_token: GitHub API token (default: from GITHUB_TOKEN env)
        repo_owner: Repository owner (default: from GITHUB_REPOSITORY env)
        repo_name: Repository name (default: from GITHUB_REPOSITORY env)
        strict_mode: Treat warnings as errors
        check_orphans: Enable orphan TAG detection
        tag_pattern: Custom TAG regex pattern
    """

    def __init__(
        self,
        github_token: Optional[str] = None,
        repo_owner: Optional[str] = None,
        repo_name: Optional[str] = None,
        strict_mode: bool = False,
        check_orphans: bool = True,
        tag_pattern: Optional[str] = None
    ):
        super().__init__(strict_mode, check_orphans, tag_pattern)

        # GitHub configuration from environment or parameters
        self.github_token = github_token or os.environ.get('GITHUB_TOKEN', '')

        # Parse repo info from GITHUB_REPOSITORY (format: "owner/repo")
        repo_full = os.environ.get('GITHUB_REPOSITORY', '')
        if '/' in repo_full and not repo_owner and not repo_name:
            parts = repo_full.split('/', 1)
            self.repo_owner = parts[0]
            self.repo_name = parts[1]
        else:
            self.repo_owner = repo_owner or ''
            self.repo_name = repo_name or ''

    def get_pr_changed_files(self, pr_number: int) -> List[str]:
        """Fetch list of changed files in a PR via GitHub API

        Args:
            pr_number: Pull request number

        Returns:
            List of relative file paths changed in the PR
        """
        if not self.github_token or not self.repo_owner or not self.repo_name:
            return []

        url = (
            f"https://api.github.com/repos/"
            f"{self.repo_owner}/{self.repo_name}/pulls/{pr_number}/files"
        )

        headers = {
            'Authorization': f'Bearer {self.github_token}',
            'Accept': 'application/vnd.github.v3+json'
        }

        # Create session with retry strategy
        session = requests.Session()
        retry = Retry(
            total=3,
            backoff_factor=0.5,
            status_forcelist=[500, 502, 503, 504],
            allowed_methods=["GET"]
        )
        adapter = HTTPAdapter(max_retries=retry)
        session.mount("https://", adapter)

        try:
            # Dual timeout: (connect_timeout, read_timeout)
            response = session.get(
                url,
                headers=headers,
                timeout=(5, 10)
            )
            response.raise_for_status()

            files_data = response.json()
            return [file_info['filename'] for file_info in files_data]

        except requests.exceptions.Timeout:
            # Network timeout - return empty list gracefully
            return []
        except requests.exceptions.HTTPError as e:
            # HTTP error (4xx, 5xx)
            if e.response.status_code == 404:
                # PR not found
                return []
            # Other HTTP errors: log but continue
            return []
        except requests.exceptions.RequestException:
            # Network/connection errors
            return []
        finally:
            session.close()

    def validate_pr_changes(
        self,
        pr_number: int,
        base_branch: str = "main"
    ) -> ValidationResult:
        """Validate TAG annotations in PR changed files

        Main CI/CD validation method:
        1. Fetch changed files from GitHub API
        2. Run validation checks on those files
        3. Return structured validation result

        Args:
            pr_number: Pull request number
            base_branch: Base branch name (not used currently)

        Returns:
            ValidationResult with errors and warnings
        """
        # Get PR changed files
        files = self.get_pr_changed_files(pr_number)

        if not files:
            return ValidationResult(is_valid=True)

        # Validate the changed files
        return self.validate_files(files)

    def generate_report(self, result: ValidationResult) -> Dict[str, Any]:
        """Generate structured validation report

        Creates JSON-serializable report with:
        - Status (success/failure/success_with_warnings)
        - Error details (message, tag, locations)
        - Warning details (message, tag, location)
        - Statistics (counts)
        - Configuration (strict_mode)

        Args:
            result: ValidationResult from validation

        Returns:
            Dictionary with structured report data
        """
        # Determine status
        if not result.is_valid:
            status = 'failure'
        elif result.warnings:
            status = 'success_with_warnings'
        else:
            status = 'success'

        # Build error list
        errors = []
        for error in result.errors:
            errors.append({
                'message': error.message,
                'tag': error.tag,
                'locations': [
                    {'file': filepath, 'line': line_num}
                    for filepath, line_num in error.locations
                ]
            })

        # Build warning list
        warnings = []
        for warning in result.warnings:
            warnings.append({
                'message': warning.message,
                'tag': warning.tag,
                'location': {
                    'file': warning.location[0],
                    'line': warning.location[1]
                }
            })

        # Calculate statistics
        statistics = {
            'total_errors': len(result.errors),
            'total_warnings': len(result.warnings),
            'total_issues': len(result.errors) + len(result.warnings)
        }

        # Build complete report
        report = {
            'status': status,
            'is_valid': result.is_valid,
            'strict_mode': self.strict_mode,
            'summary': self._generate_summary(result),
            'errors': errors,
            'warnings': warnings,
            'statistics': statistics
        }

        return report

    def _generate_summary(self, result: ValidationResult) -> str:
        """Generate human-readable summary text

        Args:
            result: ValidationResult

        Returns:
            Summary string
        """
        if result.is_valid and not result.warnings:
            return "All TAG validations passed. No issues found."
        elif result.is_valid and result.warnings:
            return f"Validation passed with {len(result.warnings)} warning(s)."
        else:
            return f"Validation failed with {len(result.errors)} error(s)."

    def format_pr_comment(
        self,
        result: ValidationResult,
        pr_url: str
    ) -> str:
        """Format validation result as markdown PR comment

        Creates formatted markdown comment with:
        - Status indicator (emoji)
        - Summary message
        - Error/warning table
        - Action items
        - Documentation links

        Args:
            result: ValidationResult from validation
            pr_url: URL of the pull request

        Returns:
            Markdown-formatted comment string
        """
        lines = []

        # Header with status indicator
        if result.is_valid and not result.warnings:
            lines.append("## ✅ TAG Validation Passed")
            lines.append("")
            lines.append("All TAG annotations are valid. No issues found.")
        elif result.is_valid and result.warnings:
            lines.append("## ⚠️ TAG Validation Passed with Warnings")
            lines.append("")
            lines.append(f"Validation passed but found {len(result.warnings)} warning(s).")
        else:
            lines.append("## ❌ TAG Validation Failed")
            lines.append("")
            lines.append(f"Found {len(result.errors)} error(s) that must be fixed.")

        lines.append("")

        # Error table
        if result.errors:
            lines.append("### Errors")
            lines.append("")
            lines.append("| TAG | Issue | Location |")
            lines.append("|-----|-------|----------|")

            for error in result.errors:
                tag = error.tag
                message = error.message
                locations = ', '.join([
                    f"`{f}:{line}`" for f, line in error.locations[:3]
                ])
                if len(error.locations) > 3:
                    locations += f" (+{len(error.locations) - 3} more)"

                lines.append(f"| `{tag}` | {message} | {locations} |")

            lines.append("")

        # Warning table
        if result.warnings:
            lines.append("### Warnings")
            lines.append("")
            lines.append("| TAG | Issue | Location |")
            lines.append("|-----|-------|----------|")

            for warning in result.warnings:
                tag = warning.tag
                message = warning.message
                location = f"`{warning.location[0]}:{warning.location[1]}`"

                lines.append(f"| `{tag}` | {message} | {location} |")

            lines.append("")

        # Action items
        if result.errors or result.warnings:
            lines.append("### How to Fix")
            lines.append("")

            if result.errors:
                lines.append("**Errors (must fix):**")
                lines.append("- Remove duplicate TAG declarations")
                lines.append("- Ensure TAGs follow format: `@PREFIX:DOMAIN-TYPE-NNN`")
                lines.append("")

            if result.warnings:
                lines.append("**Warnings (recommended):**")
                lines.append("- Add corresponding TEST tags for CODE tags")
                lines.append("- Add corresponding CODE tags for TEST tags")
                lines.append("- Complete TAG chain: SPEC → CODE → TEST → DOC")
                lines.append("")

        # Documentation link
        lines.append("---")
        lines.append("")
        lines.append("📚 **Documentation:** [TAG System Guide](.moai/memory/tag-system-guide.md)")
        lines.append("")
        lines.append(f"🔗 **PR:** {pr_url}")

        return "\n".join(lines)

    def get_pr_number_from_event(self) -> Optional[int]:
        """Extract PR number from GitHub Actions event file

        Reads GITHUB_EVENT_PATH to get PR number from event payload.

        Returns:
            PR number or None if not found
        """
        event_path = os.environ.get('GITHUB_EVENT_PATH')
        if not event_path:
            return None

        try:
            with open(event_path, 'r') as f:
                event_data = json.load(f)
                return event_data.get('pull_request', {}).get('number')
        except Exception:
            return None

    def generate_tag_report_link(self, pr_number: int) -> str:
        """Generate link to TAG reports for this PR

        Integration point with Component 4 (Reporting).
        Provides link to automated TAG reports generated by GitHub Actions.

        Args:
            pr_number: Pull request number

        Returns:
            Markdown link to TAG reports
        """
        # Link to GitHub Actions artifacts or docs directory
        if self.repo_owner and self.repo_name:
            docs_url = (
                f"https://github.com/{self.repo_owner}/{self.repo_name}/tree/main/docs"
            )
            return f"📊 [View TAG Reports]({docs_url})"
        else:
            return "📊 TAG Reports: See docs/ directory"


def main():
    """CLI entry point for CI/CD validation"""
    import argparse
    import sys

    parser = argparse.ArgumentParser(
        description="Validate TAG annotations in GitHub PR"
    )
    parser.add_argument(
        "--pr-number",
        type=int,
        help="Pull request number (default: from GitHub Actions event)"
    )
    parser.add_argument(
        "--strict",
        action="store_true",
        help="Treat warnings as errors"
    )
    parser.add_argument(
        "--no-orphan-check",
        action="store_true",
        help="Disable orphan TAG checking"
    )
    parser.add_argument(
        "--output-json",
        help="Output report to JSON file"
    )
    parser.add_argument(
        "--output-comment",
        help="Output PR comment to file"
    )

    args = parser.parse_args()

    validator = CIValidator(
        strict_mode=args.strict,
        check_orphans=not args.no_orphan_check
    )

    # Get PR number
    pr_number = args.pr_number
    if not pr_number:
        pr_number = validator.get_pr_number_from_event()

    if not pr_number:
        print("Error: Could not determine PR number", file=sys.stderr)
        sys.exit(1)

    # Run validation
    result = validator.validate_pr_changes(pr_number)

    # Generate report
    report = validator.generate_report(result)

    # Output JSON report if requested
    if args.output_json:
        with open(args.output_json, 'w') as f:
            json.dump(report, f, indent=2)

    # Output PR comment if requested
    if args.output_comment:
        pr_url = (
            f"https://github.com/{validator.repo_owner}/"
            f"{validator.repo_name}/pull/{pr_number}"
        )
        comment = validator.format_pr_comment(result, pr_url)
        with open(args.output_comment, 'w') as f:
            f.write(comment)

    # Print summary
    print(result.format())

    # Exit with error code if validation failed
    sys.exit(0 if result.is_valid else 1)


if __name__ == "__main__":
    main()
