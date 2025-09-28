#!/usr/bin/env python3
# @TASK:CONSTITUTION-VERSIONING-001
"""
Versioning Checker Module

Checks TRUST Trackable principle - version control and traceability.
Validates Git practices, tagging, and version management.
"""

import subprocess
from pathlib import Path
from typing import Dict, Any, List
import re


class VersioningChecker:
    """Checks version control and traceability practices."""

    def __init__(self, project_root: Path):
        self.project_root = project_root

    def check_versioning_principle(self) -> Dict[str, Any]:
        """Check Trackable principle: version control and traceability."""
        issues = []
        score = 100

        # Check Git repository
        git_issues = self.check_git_repository()
        if git_issues:
            issues.extend(git_issues)
            score -= min(len(git_issues) * 15, 40)

        # Check version tagging
        tagging_issues = self.check_version_tagging()
        if tagging_issues:
            issues.extend(tagging_issues)
            score -= min(len(tagging_issues) * 10, 30)

        # Check traceability
        traceability_issues = self.check_traceability()
        if traceability_issues:
            issues.extend(traceability_issues)
            score -= min(len(traceability_issues) * 5, 20)

        return {
            "passed": score >= 80,
            "score": max(score, 0),
            "issues": issues,
        }

    def check_git_repository(self) -> List[str]:
        """Check Git repository configuration."""
        issues = []

        # Check if Git repository exists
        git_dir = self.project_root / ".git"
        if not git_dir.exists():
            issues.append("No Git repository found")
            return issues

        # Check for gitignore
        gitignore = self.project_root / ".gitignore"
        if not gitignore.exists():
            issues.append("No .gitignore file found")

        return issues

    def check_version_tagging(self) -> List[str]:
        """Check version tagging practices."""
        issues = []

        try:
            # Check for version tags
            result = subprocess.run(
                ["git", "tag", "--list"],
                cwd=self.project_root,
                capture_output=True,
                text=True,
                timeout=10
            )

            if result.returncode == 0:
                tags = result.stdout.strip().split('\n') if result.stdout.strip() else []
                version_tags = [tag for tag in tags if re.match(r'v?\d+\.\d+\.\d+', tag)]

                if not version_tags:
                    issues.append("No semantic version tags found")
            else:
                issues.append("Could not check Git tags")

        except (subprocess.TimeoutExpired, FileNotFoundError):
            issues.append("Git command not available or timed out")

        return issues

    def check_traceability(self) -> List[str]:
        """Check traceability implementation."""
        issues = []

        # Check for TAG references in code
        tag_count = 0
        for file_path in self.project_root.rglob("*.py"):
            if "__pycache__" in str(file_path):
                continue
            try:
                content = file_path.read_text(encoding='utf-8', errors='ignore')
                tags = re.findall(r'@[A-Z]+:[A-Z0-9_-]+', content)
                tag_count += len(tags)
            except (UnicodeDecodeError, PermissionError):
                continue

        if tag_count == 0:
            issues.append("No traceability tags found in codebase")

        return issues