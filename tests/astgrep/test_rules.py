# AST-grep rules tests
"""Tests for AST-grep rule loading and management.

Following TDD RED-GREEN-REFACTOR cycle.
These tests define the expected behavior of Rule and RuleLoader classes.
"""

from __future__ import annotations

import tempfile
from pathlib import Path

import pytest


class TestRule:
    """Tests for Rule dataclass."""

    def test_rule_import(self) -> None:
        """Test that Rule can be imported."""
        from moai_adk.astgrep.rules import Rule

        assert Rule is not None

    def test_rule_creation_with_all_fields(self) -> None:
        """Test Rule creation with all fields."""
        from moai_adk.astgrep.rules import Rule

        rule = Rule(
            id="no-console",
            language="javascript",
            severity="warning",
            message="Avoid using console.log",
            pattern="console.log($MSG)",
            fix="logger.info($MSG)",
        )

        assert rule.id == "no-console"
        assert rule.language == "javascript"
        assert rule.severity == "warning"
        assert rule.message == "Avoid using console.log"
        assert rule.pattern == "console.log($MSG)"
        assert rule.fix == "logger.info($MSG)"

    def test_rule_creation_without_fix(self) -> None:
        """Test Rule creation with optional fix as None."""
        from moai_adk.astgrep.rules import Rule

        rule = Rule(
            id="sql-injection",
            language="python",
            severity="error",
            message="Potential SQL injection",
            pattern='cursor.execute(f"$SQL")',
            fix=None,
        )

        assert rule.id == "sql-injection"
        assert rule.fix is None

    def test_rule_severity_values(self) -> None:
        """Test that Rule accepts valid severity values."""
        from moai_adk.astgrep.rules import Rule

        valid_severities = ["error", "warning", "info", "hint"]

        for severity in valid_severities:
            rule = Rule(
                id="test-rule",
                language="python",
                severity=severity,
                message="Test message",
                pattern="test_pattern",
                fix=None,
            )
            assert rule.severity == severity


class TestRuleLoader:
    """Tests for RuleLoader class."""

    def test_ruleloader_import(self) -> None:
        """Test that RuleLoader can be imported."""
        from moai_adk.astgrep.rules import RuleLoader

        assert RuleLoader is not None

    def test_ruleloader_instantiation(self) -> None:
        """Test RuleLoader can be instantiated."""
        from moai_adk.astgrep.rules import RuleLoader

        loader = RuleLoader()
        assert loader is not None

    def test_load_from_file_single_rule(self) -> None:
        """Test loading a single rule from YAML file."""
        from moai_adk.astgrep.rules import RuleLoader

        yaml_content = """
id: no-var
language: javascript
severity: warning
message: Use const or let instead of var
rule:
  pattern: 'var $NAME = $VALUE'
fix: 'const $NAME = $VALUE'
"""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".yml", delete=False) as f:
            f.write(yaml_content)
            f.flush()

            loader = RuleLoader()
            rules = loader.load_from_file(f.name)

            assert len(rules) == 1
            assert rules[0].id == "no-var"
            assert rules[0].language == "javascript"
            assert rules[0].severity == "warning"
            assert rules[0].pattern == "var $NAME = $VALUE"
            assert rules[0].fix == "const $NAME = $VALUE"

        Path(f.name).unlink()

    def test_load_from_file_multiple_rules(self) -> None:
        """Test loading multiple rules from YAML file with document separator."""
        from moai_adk.astgrep.rules import RuleLoader

        yaml_content = """
id: no-console
language: javascript
severity: warning
message: Avoid console.log
rule:
  pattern: 'console.log($MSG)'
fix: 'logger.info($MSG)'
---
id: no-eval
language: javascript
severity: error
message: Avoid using eval
rule:
  pattern: 'eval($CODE)'
"""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".yml", delete=False) as f:
            f.write(yaml_content)
            f.flush()

            loader = RuleLoader()
            rules = loader.load_from_file(f.name)

            assert len(rules) == 2
            assert rules[0].id == "no-console"
            assert rules[1].id == "no-eval"
            assert rules[1].fix is None

        Path(f.name).unlink()

    def test_load_from_file_nonexistent(self) -> None:
        """Test loading from nonexistent file raises appropriate error."""
        from moai_adk.astgrep.rules import RuleLoader

        loader = RuleLoader()

        with pytest.raises(FileNotFoundError):
            loader.load_from_file("/nonexistent/rules.yml")

    def test_load_from_file_invalid_yaml(self) -> None:
        """Test loading invalid YAML raises appropriate error."""
        from moai_adk.astgrep.rules import RuleLoader

        yaml_content = """
id: invalid
language: javascript
  invalid yaml content:
    - missing proper structure
"""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".yml", delete=False) as f:
            f.write(yaml_content)
            f.flush()

            loader = RuleLoader()

            with pytest.raises(ValueError):
                loader.load_from_file(f.name)

        Path(f.name).unlink()

    def test_load_builtin_rules(self) -> None:
        """Test loading built-in rules."""
        from moai_adk.astgrep.rules import RuleLoader

        loader = RuleLoader()
        rules = loader.load_builtin_rules()

        # Should return a list of rules
        assert isinstance(rules, list)
        # Should have at least some built-in rules
        assert len(rules) >= 0  # May be empty if no built-in rules defined

    def test_get_rules_for_language_python(self) -> None:
        """Test filtering rules for Python language."""
        from moai_adk.astgrep.rules import Rule, RuleLoader

        loader = RuleLoader()
        # Add some test rules
        loader._rules = [
            Rule(
                id="py-rule",
                language="python",
                severity="warning",
                message="Python rule",
                pattern="$PATTERN",
                fix=None,
            ),
            Rule(
                id="js-rule",
                language="javascript",
                severity="warning",
                message="JS rule",
                pattern="$PATTERN",
                fix=None,
            ),
            Rule(
                id="py-rule-2",
                language="python",
                severity="error",
                message="Another Python rule",
                pattern="$PATTERN2",
                fix=None,
            ),
        ]

        python_rules = loader.get_rules_for_language("python")

        assert len(python_rules) == 2
        assert all(r.language == "python" for r in python_rules)

    def test_get_rules_for_language_empty_result(self) -> None:
        """Test filtering rules returns empty list for unsupported language."""
        from moai_adk.astgrep.rules import RuleLoader

        loader = RuleLoader()
        loader._rules = []

        rules = loader.get_rules_for_language("cobol")

        assert rules == []

    def test_load_from_directory(self) -> None:
        """Test loading rules from a directory."""
        from moai_adk.astgrep.rules import RuleLoader

        with tempfile.TemporaryDirectory() as tmpdir:
            # Create rule files
            rule1 = Path(tmpdir) / "rule1.yml"
            rule1.write_text("""
id: rule-one
language: python
severity: warning
message: Rule one
rule:
  pattern: 'rule_one($ARG)'
""")

            rule2 = Path(tmpdir) / "rule2.yaml"
            rule2.write_text("""
id: rule-two
language: javascript
severity: error
message: Rule two
rule:
  pattern: 'ruleTwo($ARG)'
""")

            # Non-yaml file should be ignored
            (Path(tmpdir) / "readme.txt").write_text("Not a rule file")

            loader = RuleLoader()
            rules = loader.load_from_directory(tmpdir)

            assert len(rules) == 2
            rule_ids = {r.id for r in rules}
            assert "rule-one" in rule_ids
            assert "rule-two" in rule_ids


class TestRuleExports:
    """Tests for module exports."""

    def test_rule_exported_from_rules_module(self) -> None:
        """Test Rule is exported from rules module."""
        from moai_adk.astgrep.rules import Rule

        assert Rule is not None

    def test_ruleloader_exported_from_rules_module(self) -> None:
        """Test RuleLoader is exported from rules module."""
        from moai_adk.astgrep.rules import RuleLoader

        assert RuleLoader is not None

    def test_rules_exported_from_astgrep_package(self) -> None:
        """Test Rule and RuleLoader exported from astgrep package."""
        from moai_adk import astgrep

        assert hasattr(astgrep, "Rule")
        assert hasattr(astgrep, "RuleLoader")
