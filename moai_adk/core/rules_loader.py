"""
.claude/rules/ Directory Loader

Loads path-specific rules from .claude/rules/*.md files.

Based on Claude Code official documentation:
- All .md files in .claude/rules/ are automatically loaded
- YAML frontmatter with `paths` field for path-specific rules
- Glob patterns supported: **/*.ts, src/**/*.{ts,tsx}
- Subdirectories supported
- Symlinks supported
"""

from __future__ import annotations

import logging
import re
from dataclasses import dataclass
from pathlib import Path
from typing import List, Optional

try:
    import yaml  # noqa: F401

    YAML_AVAILABLE = True
except ImportError:
    YAML_AVAILABLE = False

logger = logging.getLogger(__name__)


@dataclass
class Rule:
    """A rule loaded from .claude/rules/."""

    name: str
    content: str
    paths: Optional[List[str]]  # None = applies to all files
    source_file: Path


class RulesLoader:
    """Load rules from .claude/rules/ directory."""

    def __init__(self, project_path: Path):
        """
        Initialize the rules loader.

        Args:
            project_path: Path to the project root
        """
        self.project_path = project_path.resolve()
        self.rules_dir = self.project_path / ".claude" / "rules"
        self.user_rules_dir = Path.home() / ".claude" / "rules"

    def load_rules(self, file_path: Optional[Path] = None) -> List[Rule]:
        """
        Load rules applicable to a specific file.

        Args:
            file_path: File to check for path-specific rules

        Returns:
            List of applicable rules (user rules first, then project rules)
        """
        rules = []

        # Load user rules (~/.claude/rules/)
        rules.extend(self._load_rules_from_dir(self.user_rules_dir, file_path))

        # Load project rules (.claude/rules/)
        if self.rules_dir.exists():
            rules.extend(self._load_rules_from_dir(self.rules_dir, file_path))

        return rules

    def load_all_rules(self) -> List[Rule]:
        """
        Load all rules without path filtering.

        Returns:
            List of all rules (user rules first, then project rules)
        """
        return self.load_rules(file_path=None)

    def _load_rules_from_dir(self, rules_dir: Path, file_path: Optional[Path]) -> List[Rule]:
        """
        Load all rules from a directory.

        Args:
            rules_dir: Directory containing rule files
            file_path: Optional file to filter rules by path

        Returns:
            List of applicable rules
        """
        rules: List[Rule] = []

        if not rules_dir.exists():
            return rules

        # Find all .md files (including subdirectories)
        for md_file in rules_dir.rglob("*.md"):
            # Skip symlinks that create circular references
            if md_file.is_symlink():
                try:
                    resolved = md_file.resolve()
                    # Check if this is within the rules directory (prevent loops)
                    resolved.relative_to(rules_dir)
                except (ValueError, RuntimeError):
                    logger.warning(f"Skipping circular symlink: {md_file}")
                    continue

            rule = self._parse_rule_file(md_file)
            if rule:
                # Check if rule applies to this file
                if file_path is None or self._rule_applies(rule, file_path):
                    rules.append(rule)

        return rules

    def _parse_rule_file(self, file_path: Path) -> Optional[Rule]:
        """
        Parse a rule file with YAML frontmatter.

        Args:
            file_path: Path to the rule file

        Returns:
            Rule object or None if parsing fails
        """
        try:
            content = file_path.read_text(encoding="utf-8", errors="replace")

            # Extract YAML frontmatter
            paths = None
            if YAML_AVAILABLE:
                frontmatter_match = re.match(r"^---\n(.*?)\n---", content, re.DOTALL)
                if frontmatter_match:
                    yaml_content = frontmatter_match.group(1)
                    try:
                        frontmatter_data = yaml.safe_load(yaml_content)
                        if frontmatter_data and isinstance(frontmatter_data, dict):
                            paths = frontmatter_data.get("paths")
                            if paths and not isinstance(paths, list):
                                paths = None
                    except yaml.YAMLError:
                        pass

                # Remove frontmatter from content
                if frontmatter_match:
                    content = content[frontmatter_match.end() :].lstrip()

            return Rule(
                name=file_path.stem,
                content=content,
                paths=paths,
                source_file=file_path,
            )

        except Exception as e:
            logger.warning(f"Failed to parse rule file {file_path}: {e}")
            return None

    def _rule_applies(self, rule: Rule, file_path: Path) -> bool:
        """
        Check if a rule applies to a specific file.

        Args:
            rule: Rule to check
            file_path: File to check against

        Returns:
            True if the rule applies to the file
        """
        if rule.paths is None:
            return True  # Applies to all files

        # Convert file_path to relative path from project root
        try:
            relative_path = file_path.relative_to(self.project_path)
            path_str = str(relative_path).replace("\\", "/")
        except ValueError:
            # File is outside project, use absolute path
            path_str = str(file_path).replace("\\", "/")

        # Check each pattern
        for pattern in rule.paths:
            if self._matches_pattern(pattern, path_str):
                return True

        return False

    def _matches_pattern(self, pattern: str, path: str) -> bool:
        """
        Check if a path matches a glob pattern.

        Args:
            pattern: Glob pattern (supports **, *, ?, { })
            path: Path string to check

        Returns:
            True if the path matches the pattern
        """
        # Handle brace expansion: {src,lib}/**/*.ts
        if "{" in pattern and "}" in pattern:
            # Simple brace expansion
            base, rest = pattern.split("{", 1)
            options_part, suffix = rest.split("}", 1)
            options = options_part.split(",")

            for option in options:
                expanded_pattern = f"{base}{option}{suffix}"
                if self._matches_pattern(expanded_pattern, path):
                    return True
            return False

        # Convert glob pattern to regex
        # ** matches any number of directories
        regex_pattern = pattern.replace("**", ".*")
        # * matches any characters within a directory
        regex_pattern = regex_pattern.replace("*", "[^/]*")
        # ? matches a single character
        regex_pattern = regex_pattern.replace("?", ".")
        regex_pattern = f"^{regex_pattern}$"

        return re.match(regex_pattern, path) is not None


def load_rules_for_file(project_path: Path, file_path: Path) -> List[Rule]:
    """
    Load rules applicable to a specific file.

    Convenience function that creates a RulesLoader and filters rules.

    Args:
        project_path: Path to the project root
        file_path: Path to the file

    Returns:
        List of applicable rules
    """
    loader = RulesLoader(project_path)
    return loader.load_rules(file_path)


def load_all_rules(project_path: Path) -> List[Rule]:
    """
    Load all rules without path filtering.

    Convenience function that creates a RulesLoader and loads all rules.

    Args:
        project_path: Path to the project root

    Returns:
        List of all rules
    """
    loader = RulesLoader(project_path)
    return loader.load_all_rules()
