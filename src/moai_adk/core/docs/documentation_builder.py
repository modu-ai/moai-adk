"""@FEATURE:DOCS-001 MkDocs documentation builder

Handles MkDocs Material site initialization, building, and validation.
Follows TRUST 5 principles for maintainable documentation automation.

@REQ:DOCS-SITE-001 → @TASK:DOC-BUILDER-001
"""
import logging
from pathlib import Path
from typing import Any

import yaml

logger = logging.getLogger(__name__)


class DocumentationBuilder:
    """@FEATURE:DOCS-002 Main documentation builder class"""

    def __init__(self, project_root: str):
        """Initialize documentation builder

        Args:
            project_root: Root directory of the project
        """
        self.project_root = Path(project_root)
        self.docs_dir = self.project_root / "docs"
        self.mkdocs_config = self.project_root / "mkdocs.yml"
        self._build_status = {"success": True, "error": None}

    def initialize_site(self) -> None:
        """@TASK:DOC-BUILDER-002 Initialize MkDocs site structure"""
        # Create docs directory
        self.docs_dir.mkdir(exist_ok=True)

        # Create subdirectories
        subdirs = ["getting-started", "guide", "development", "examples"]
        for subdir in subdirs:
            (self.docs_dir / subdir).mkdir(exist_ok=True)

        # Create mkdocs.yml
        self._create_mkdocs_config()

        # Create essential files
        self._create_essential_files()

    def _create_mkdocs_config(self) -> None:
        """Create basic mkdocs.yml configuration"""
        config = {
            "site_name": "MoAI-ADK Documentation",
            "site_description": "MoAI Agentic Development Kit Documentation",
            "site_url": "https://docs.moai-adk.dev",
            "repo_url": "https://github.com/modu-ai/moai-adk",
            "theme": {
                "name": "material",
                "features": [
                    "navigation.tabs",
                    "navigation.sections",
                    "navigation.expand",
                    "navigation.top",
                    "search.highlight",
                    "search.share"
                ],
                "palette": [
                    {
                        "scheme": "default",
                        "primary": "blue",
                        "accent": "blue",
                        "toggle": {
                            "icon": "material/brightness-7",
                            "name": "Switch to dark mode"
                        }
                    }
                ]
            },
            "nav": [
                {"Home": "index.md"},
                {"Getting Started": [
                    {"Installation": "getting-started/installation.md"}
                ]},
                {"Guide": [
                    {"Workflow": "guide/workflow.md"}
                ]},
                {"Development": [
                    {"Contributing": "development/contributing.md"}
                ]},
                {"Examples": [
                    {"Basic Usage": "examples/basic.md"}
                ]}
            ],
            "markdown_extensions": [
                "admonition",
                "codehilite",
                "pymdownx.superfences",
                "pymdownx.tabbed",
                "toc"
            ],
            "plugins": [
                "search"
            ]
        }

        with open(self.mkdocs_config, 'w') as f:
            yaml.dump(config, f, default_flow_style=False)

    def _create_essential_files(self) -> None:
        """Create essential documentation files"""
        # Create index.md
        index_content = """# MoAI-ADK Documentation

Welcome to MoAI-ADK (MoAI Agentic Development Kit) documentation.

## Quick Start

Get started with MoAI-ADK in minutes.

## Features

- Spec-First TDD Development
- Claude Code Integration
- Git Workflow Automation
- Living Documentation
"""
        (self.docs_dir / "index.md").write_text(index_content)

        # Create installation guide
        install_content = """# Installation

Install MoAI-ADK using pip:

```bash
pip install moai-adk
```

## Quick Setup

Initialize your project:

```bash
moai init
```
"""
        (self.docs_dir / "getting-started" / "installation.md").write_text(install_content)

        # Create workflow guide
        workflow_content = """# Workflow

MoAI-ADK follows a 4-stage development pipeline:

1. `/moai:0-project` - Project initialization
2. `/moai:1-spec` - Specification creation
3. `/moai:2-build` - TDD implementation
4. `/moai:3-sync` - Documentation sync
"""
        (self.docs_dir / "guide" / "workflow.md").write_text(workflow_content)

    def validate_config(self) -> bool:
        """@TASK:DOC-BUILDER-003 Validate MkDocs configuration"""
        try:
            if not self.mkdocs_config.exists():
                return False

            with open(self.mkdocs_config) as f:
                config = yaml.safe_load(f)

            # Basic validation
            required_keys = ["site_name", "theme", "nav"]
            for key in required_keys:
                if key not in config:
                    return False

            return True
        except Exception as e:
            logger.error(f"Config validation failed: {e}")
            return False

    def build_docs(self, incremental: bool = False) -> bool:
        """@TASK:DOC-BUILDER-004 Build documentation site"""
        try:
            # Check if docs directory exists
            if not self.docs_dir.exists():
                self._build_status = {"success": False, "error": "docs directory not found"}
                logger.error("Documentation build failed: docs directory not found")
                return False

            # Mock MkDocs build for now
            # In real implementation, this would call mkdocs.commands.build.build
            self._build_status = {"success": True, "error": None}
            return True

        except Exception as e:
            self._build_status = {"success": False, "error": str(e)}
            logger.error(f"Documentation build failed: {e}")
            return False

    def get_build_status(self) -> dict[str, Any]:
        """Get current build status information"""
        return self._build_status

    def validate_links(self) -> dict[str, list[str]]:
        """@TASK:DOC-BUILDER-005 Validate internal links"""
        missing_links = []
        valid_links = []

        if not self.docs_dir.exists():
            return {"missing": missing_links, "valid": valid_links}

        # Scan markdown files for links
        for md_file in self.docs_dir.glob("**/*.md"):
            content = md_file.read_text()
            # Simple link extraction (would be more sophisticated in real implementation)
            import re
            links = re.findall(r'\[.*?\]\(([^)]+)\)', content)

            for link in links:
                if not link.startswith(('http', '#')):  # Internal links only
                    link_path = self.docs_dir / link
                    if link_path.exists():
                        valid_links.append(link)
                    else:
                        missing_links.append(link)

        return {"missing": missing_links, "valid": valid_links}

    def check_completeness(self) -> dict[str, Any]:
        """@TASK:DOC-BUILDER-006 Check documentation completeness"""
        required_sections = [
            "getting-started/installation.md",
            "guide/workflow.md",
            "api/index.md",
            "release-notes.md"
        ]

        existing_sections = []
        missing_sections = []

        for section in required_sections:
            section_path = self.docs_dir / section
            if section_path.exists():
                existing_sections.append(section)
            else:
                missing_sections.append(section)

        coverage_percent = len(existing_sections) / len(required_sections) * 100

        return {
            "coverage_percent": coverage_percent,
            "existing_sections": existing_sections,
            "missing_sections": missing_sections
        }
