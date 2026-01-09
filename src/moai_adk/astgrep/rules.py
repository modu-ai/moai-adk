# AST-grep rule management
"""Rule loading and management for AST-grep analyzer.

This module provides classes for defining AST-grep rules and loading
them from YAML configuration files.
"""

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Any

import yaml


@dataclass
class Rule:
    """AST-grep rule definition.

    Represents a single AST-grep rule that can be used for pattern
    matching and code analysis.

    Attributes:
        id: Unique identifier for the rule.
        language: Programming language this rule applies to.
        severity: Severity level (error, warning, info, hint).
        message: Human-readable description of the rule.
        pattern: AST-grep pattern to match.
        fix: Optional fix pattern for automatic correction.
    """

    id: str
    language: str
    severity: str  # error, warning, info, hint
    message: str
    pattern: str
    fix: str | None = None


class RuleLoader:
    """Loader for AST-grep rules from various sources.

    Supports loading rules from individual YAML files, directories
    containing multiple rule files, and built-in rule sets.
    """

    def __init__(self) -> None:
        """Initialize the RuleLoader."""
        self._rules: list[Rule] = []

    def load_from_file(self, config_path: str) -> list[Rule]:
        """Load rules from a YAML configuration file.

        Supports both single-document and multi-document YAML files
        (documents separated by ---).

        Args:
            config_path: Path to the YAML configuration file.

        Returns:
            List of Rule objects loaded from the file.

        Raises:
            FileNotFoundError: If the file does not exist.
            ValueError: If the YAML content is invalid.
        """
        path = Path(config_path)
        if not path.exists():
            raise FileNotFoundError(f"Rule file not found: {config_path}")

        try:
            content = path.read_text()
            documents = list(yaml.safe_load_all(content))
        except yaml.YAMLError as e:
            raise ValueError(f"Invalid YAML in {config_path}: {e}") from e

        rules: list[Rule] = []
        for doc in documents:
            if doc is None:
                continue
            rule = self._parse_rule_document(doc)
            if rule:
                rules.append(rule)

        self._rules.extend(rules)
        return rules

    def _parse_rule_document(self, doc: dict[str, Any]) -> Rule | None:
        """Parse a single rule document from YAML.

        Args:
            doc: Dictionary representing a rule document.

        Returns:
            Rule object or None if parsing fails.
        """
        if not isinstance(doc, dict):
            return None

        rule_id = doc.get("id")
        language = doc.get("language")
        severity = doc.get("severity", "warning")
        message = doc.get("message", "")

        # Extract pattern from rule.pattern or direct pattern field
        pattern = None
        if "rule" in doc and isinstance(doc["rule"], dict):
            pattern = doc["rule"].get("pattern")
        elif "pattern" in doc:
            pattern = doc["pattern"]

        # Extract fix if present
        fix = doc.get("fix")

        if not rule_id or not language or not pattern:
            return None

        return Rule(
            id=rule_id,
            language=language,
            severity=severity,
            message=message,
            pattern=pattern,
            fix=fix,
        )

    def load_builtin_rules(self) -> list[Rule]:
        """Load built-in rules.

        Returns the default set of rules bundled with the analyzer.
        Currently returns an empty list; built-in rules can be added
        in future versions.

        Returns:
            List of built-in Rule objects.
        """
        # Built-in rules can be defined here or loaded from bundled files
        builtin_rules: list[Rule] = []
        self._rules.extend(builtin_rules)
        return builtin_rules

    def get_rules_for_language(self, language: str) -> list[Rule]:
        """Get all rules for a specific programming language.

        Args:
            language: The programming language to filter by.

        Returns:
            List of Rule objects for the specified language.
        """
        return [rule for rule in self._rules if rule.language == language]

    def load_from_directory(self, directory_path: str) -> list[Rule]:
        """Load rules from all YAML files in a directory.

        Scans the directory for .yml and .yaml files and loads rules
        from each one.

        Args:
            directory_path: Path to the directory containing rule files.

        Returns:
            List of all Rule objects loaded from the directory.

        Raises:
            FileNotFoundError: If the directory does not exist.
        """
        path = Path(directory_path)
        if not path.exists():
            raise FileNotFoundError(f"Directory not found: {directory_path}")

        rules: list[Rule] = []
        for yaml_file in path.glob("*.yml"):
            rules.extend(self.load_from_file(str(yaml_file)))
        for yaml_file in path.glob("*.yaml"):
            rules.extend(self.load_from_file(str(yaml_file)))

        return rules
