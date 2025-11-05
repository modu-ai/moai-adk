#!/usr/bin/env python3
"""
Documentation Manager Module for MoAI-ADK

This module provides comprehensive documentation management capabilities
including multilingual support, validation, and TAG system integration.

@CODE:DOC-ONLINE-001:MAIN
"""

import json
import os
import re
from pathlib import Path
from typing import Dict, List, Optional, Tuple
from datetime import datetime
import logging

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class DocumentationManager:
    """
    Manages multilingual documentation system with TAG integration.

    Features:
    - Multilingual document management
    - TAG system validation and synchronization
    - Document quality assurance
    - Version tracking and updates
    """

    def __init__(self, config_path: str = "multilingual_config.json"):
        """
        Initialize the documentation manager.

        Args:
            config_path: Path to the multilingual configuration file
        """
        self.config_path = Path(config_path)
        self.config = self._load_config()
        self.base_path = Path(__file__).parent
        self.supported_languages = self.config.get("languages", {})
        self.default_language = self.config.get("default_language", "ko")

    def _load_config(self) -> Dict:
        """Load multilingual configuration from JSON file."""
        try:
            if self.config_path.exists():
                with open(self.config_path, 'r', encoding='utf-8') as f:
                    return json.load(f)
            else:
                logger.warning(f"Config file {self.config_path} not found, using defaults")
                return self._get_default_config()
        except Exception as e:
            logger.error(f"Error loading config: {e}")
            return self._get_default_config()

    def _get_default_config(self) -> Dict:
        """Get default configuration for multilingual documentation."""
        return {
            "default_language": "ko",
            "languages": {
                "ko": {
                    "name": "한국어",
                    "file_suffix": "-ko.md",
                    "encoding": "utf-8",
                    "direction": "ltr"
                },
                "en": {
                    "name": "English",
                    "file_suffix": "-en.md",
                    "encoding": "utf-8",
                    "direction": "ltr"
                },
                "ja": {
                    "name": "日本語",
                    "file_suffix": "-ja.md",
                    "encoding": "utf-8",
                    "direction": "ltr"
                },
                "zh": {
                    "name": "中文",
                    "file_suffix": "-zh.md",
                    "encoding": "utf-8",
                    "direction": "ltr"
                }
            },
            "version": "0.9.0",
            "last_updated": "2025-11-05",
            "tag_system": {
                "enabled": True,
                "prefixes": ["@SPEC", "@CODE", "@TEST", "@DOC"]
            }
        }

    def validate_document_structure(self) -> Tuple[bool, List[str]]:
        """
        Validate the structure of all documentation files.

        Returns:
            Tuple of (is_valid, list_of_issues)
        """
        issues = []
        is_valid = True

        # Check for required files
        required_files = [
            "README.md",
            "README-ko.md",
            "README-en.md",
            "README-ja.md",
            "README-zh.md"
        ]

        for file_name in required_files:
            file_path = self.base_path / file_name
            if not file_path.exists():
                issues.append(f"Missing required file: {file_name}")
                is_valid = False

        # Check TAG consistency
        tag_issues = self._validate_tag_consistency()
        issues.extend(tag_issues)
        if tag_issues:
            is_valid = False

        return is_valid, issues

    def _validate_tag_consistency(self) -> List[str]:
        """Validate TAG consistency across all documentation files."""
        issues = []

        # Define expected TAGs for this SPEC
        expected_tags = [
            "@CODE:DOC-ONLINE-001:MAIN",
            "@CODE:DOC-ONLINE-001:KO",
            "@CODE:DOC-ONLINE-001:EN",
            "@CODE:DOC-ONLINE-001:JA",
            "@CODE:DOC-ONLINE-001:ZH"
        ]

        # Check each language file for appropriate TAGs
        language_tag_mapping = {
            "README-ko.md": "@CODE:DOC-ONLINE-001:KO",
            "README-en.md": "@CODE:DOC-ONLINE-001:EN",
            "README-ja.md": "@CODE:DOC-ONLINE-001:JA",
            "README-zh.md": "@CODE:DOC-ONLINE-001:ZH",
            "README.md": "@CODE:DOC-ONLINE-001:MAIN"
        }

        for file_name, expected_tag in language_tag_mapping.items():
            file_path = self.base_path / file_name
            if file_path.exists():
                with open(file_path, 'r', encoding='utf-8') as f:
                    content = f.read()

                if expected_tag not in content:
                    issues.append(f"Missing {expected_tag} in {file_name}")

        return issues

    def update_document_version(self, new_version: str) -> bool:
        """
        Update version information across all documentation files.

        Args:
            new_version: New version string (e.g., "0.9.0")

        Returns:
            True if successful, False otherwise
        """
        try:
            # Update config file
            self.config["version"] = new_version
            self.config["last_updated"] = datetime.now().strftime("%Y-%m-%d")
            self._save_config()

            # Update README files
            readme_files = list(self.base_path.glob("README*.md"))
            for readme_file in readme_files:
                self._update_file_version(readme_file, new_version)

            logger.info(f"Updated all documentation to version {new_version}")
            return True

        except Exception as e:
            logger.error(f"Error updating version: {e}")
            return False

    def _update_file_version(self, file_path: Path, new_version: str):
        """Update version in a specific file."""
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # Update version line (format: **Version**: vX.X.X)
        version_pattern = r'\*\*Version\*\*: v\d+\.\d+\.\d+'
        new_version_line = f"**Version**: v{new_version}"
        content = re.sub(version_pattern, new_version_line, content)

        # Update last updated line
        date_pattern = r'\*\*Last Updated\*\*: \d{4}-\d{2}-\d{2}'
        new_date_line = f"**Last Updated**: {datetime.now().strftime('%Y-%m-%d')}"
        content = re.sub(date_pattern, new_date_line, content)

        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(content)

    def _save_config(self):
        """Save current configuration to file."""
        with open(self.config_path, 'w', encoding='utf-8') as f:
            json.dump(self.config, f, indent=2, ensure_ascii=False)

    def generate_sync_report(self) -> Dict:
        """
        Generate a comprehensive synchronization report.

        Returns:
            Dictionary containing sync status and statistics
        """
        is_valid, issues = self.validate_document_structure()

        report = {
            "timestamp": datetime.now().isoformat(),
            "version": self.config["version"],
            "languages_supported": list(self.supported_languages.keys()),
            "default_language": self.default_language,
            "validation_status": "PASS" if is_valid else "FAIL",
            "issues": issues,
            "file_count": len(list(self.base_path.glob("*.md"))),
            "tag_system_enabled": self.config.get("tag_system", {}).get("enabled", False)
        }

        return report

    def get_documentation_stats(self) -> Dict:
        """Get statistics about the documentation system."""
        files = list(self.base_path.glob("*.md"))

        stats = {
            "total_files": len(files),
            "total_size_bytes": sum(f.stat().st_size for f in files),
            "languages": {},
            "tag_count": 0
        }

        # Count files per language
        for lang_code in self.supported_languages:
            lang_files = [f for f in files if f.name.endswith(f"-{lang_code}.md")]
            stats["languages"][lang_code] = len(lang_files)

        # Count TAGs
        for file_path in files:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
                tag_count = len(re.findall(r'@(SPEC|CODE|TEST|DOC):', content))
                stats["tag_count"] += tag_count

        return stats


def main():
    """Main function for testing the documentation manager."""
    manager = DocumentationManager()

    print("=== MoAI-ADK Documentation Manager ===")
    print(f"Version: {manager.config['version']}")
    print(f"Languages: {list(manager.supported_languages.keys())}")
    print(f"Default: {manager.default_language}")

    # Validate structure
    is_valid, issues = manager.validate_document_structure()
    print(f"\nValidation Status: {'PASS' if is_valid else 'FAIL'}")

    if issues:
        print("Issues found:")
        for issue in issues:
            print(f"  - {issue}")

    # Generate stats
    stats = manager.get_documentation_stats()
    print(f"\nStatistics:")
    print(f"  Files: {stats['total_files']}")
    print(f"  Size: {stats['total_size_bytes']} bytes")
    print(f"  TAGs: {stats['tag_count']}")

    # Generate sync report
    report = manager.generate_sync_report()
    print(f"\nSync Report Generated:")
    print(f"  Status: {report['validation_status']}")
    print(f"  Issues: {len(report['issues'])}")


if __name__ == "__main__":
    main()