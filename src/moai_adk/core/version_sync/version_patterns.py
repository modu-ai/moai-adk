#!/usr/bin/env python3
"""
@FEATURE:VERSION-PATTERNS-001 Version Pattern Provider
Dedicated module for version pattern definitions and rules
"""

from typing import Dict, List


class VersionPatternsProvider:
    """@TASK:VERSION-PATTERNS-PROVIDER-001 Provides file-specific version patterns"""

    def __init__(self, current_version: str):
        """
        Initialize version patterns provider

        Args:
            current_version: Current version to use in replacements
        """
        self.current_version = current_version

    def get_patterns(self) -> Dict[str, List[Dict]]:
        """
        Get version patterns for all supported file types

        Returns:
            Dict mapping file patterns to replacement rules
        """
        return {
            # Python package configuration
            "pyproject.toml": [
                {
                    "pattern": r'version\s*=\s*"[^"]*"',
                    "replacement": f'version = "{self.current_version}"',
                    "description": "Python package version",
                }
            ],
            # Python source files (explicit __version__ only)
            "**/*.py": [
                {
                    "pattern": r'__version__\s*=\s*"[^"]*"',
                    "replacement": f'__version__ = "0.1.17"',
                    "description": "Python module version",
                }
            ],
            # Markdown documents
            "**/*.md": [
                {
                    "pattern": r"MoAI-ADK \(MoAI Agentic Development Kit\) v[0-9]+\.[0-9]+\.[0-9]+",
                    "replacement": f"MoAI-ADK (MoAI Agentic Development Kit) v{self.current_version}",
                    "description": "MoAI-ADK full title version",
                },
                {
                    "pattern": r"MoAI-ADK v[0-9]+\.[0-9]+\.[0-9]+",
                    "replacement": f"MoAI-ADK v{self.current_version}",
                    "description": "MoAI-ADK version in documentation",
                },
                {
                    "pattern": r"version-[0-9]+\.[0-9]+\.[0-9]+-blue",
                    "replacement": f"version-{self.current_version}-blue",
                    "description": "Version badge",
                },
                {
                    "pattern": r"moai-adk-v[0-9]+\.[0-9]+\.[0-9]+",
                    "replacement": f"moai-adk-v{self.current_version}",
                    "description": "Release archive naming",
                },
                {
                    "pattern": r"\*\*MoAI-ADK 버전\*\*: v[0-9]+\.[0-9]+\.[0-9]+",
                    "replacement": f"**MoAI-ADK 버전**: v{self.current_version}",
                    "description": "Korean version footer",
                },
            ],
            # JSON configuration files
            "**/*.json": [
                {
                    "pattern": r'"version":\s*"[^"]*"',
                    "replacement": f'"version": "{self.current_version}"',
                    "description": "JSON version field",
                },
                {
                    "pattern": r'"moai_version":\s*"[^"]*"',
                    "replacement": f'"moai_version": "{self.current_version}"',
                    "description": "MoAI specific version field",
                },
                {
                    "pattern": r'"moai_adk_version":\s*"[^"]*"',
                    "replacement": f'"moai_adk_version": "{self.current_version}"',
                    "description": "MoAI ADK version field",
                },
            ],
            # GitHub Actions workflows
            "**/*.yml": [
                {
                    "pattern": r"MoAI-ADK v[0-9]+\.[0-9]+\.[0-9]+",
                    "replacement": f"MoAI-ADK v{self.current_version}",
                    "description": "MoAI-ADK version in YAML",
                }
            ],
            # Makefile
            "Makefile": [
                {
                    "pattern": r"MoAI-ADK v[0-9]+\.[0-9]+\.[0-9]+",
                    "replacement": f"MoAI-ADK v{self.current_version}",
                    "description": "Makefile version display",
                }
            ],
            # CHANGELOG
            "CHANGELOG.md": [
                {
                    "pattern": r"MoAI-ADK v[0-9]+\.[0-9]+\.[0-9]+",
                    "replacement": f"MoAI-ADK v{self.current_version}",
                    "description": "Changelog version references",
                }
            ],
        }