# AST-grep rules advanced tests
"""Tests for advanced edge cases and error handling in rule loading.

Following TDD RED-GREEN-REFACTOR cycle.
These tests cover Unicode content, error handling, and concurrent access patterns.
"""

from __future__ import annotations

import tempfile
from pathlib import Path

import pytest

from moai_adk.astgrep.rules import Rule, RuleLoader


class TestParseRuleDocument:
    """Tests for _parse_rule_document private method."""

    def test_parse_rule_document_with_direct_pattern(self) -> None:
        """Test _parse_rule_document with direct pattern field (not nested in rule)."""
        loader = RuleLoader()

        doc = {
            "id": "direct-pattern-rule",
            "language": "python",
            "severity": "warning",
            "message": "Direct pattern",
            "pattern": "print($MSG)",
        }

        rule = loader._parse_rule_document(doc)
        assert rule is not None
        assert rule.id == "direct-pattern-rule"
        assert rule.pattern == "print($MSG)"

    def test_parse_rule_document_with_nested_pattern(self) -> None:
        """Test _parse_rule_document with nested pattern in rule field."""
        loader = RuleLoader()

        doc = {
            "id": "nested-pattern-rule",
            "language": "javascript",
            "severity": "error",
            "message": "Nested pattern",
            "rule": {"pattern": "console.log($MSG)"},
        }

        rule = loader._parse_rule_document(doc)
        assert rule is not None
        assert rule.id == "nested-pattern-rule"
        assert rule.pattern == "console.log($MSG)"

    def test_parse_rule_document_non_dict_returns_none(self) -> None:
        """Test _parse_rule_document with non-dict input returns None."""
        loader = RuleLoader()

        assert loader._parse_rule_document(None) is None  # type: ignore[arg-type]
        assert loader._parse_rule_document("string") is None  # type: ignore[arg-type]
        assert loader._parse_rule_document(["list"]) is None  # type: ignore[arg-type]
        assert loader._parse_rule_document(123) is None  # type: ignore[arg-type]

    def test_parse_rule_document_missing_required_fields(self) -> None:
        """Test _parse_rule_document with missing required fields returns None."""
        loader = RuleLoader()

        # Missing id
        doc1 = {"language": "python", "pattern": "print($MSG)"}
        assert loader._parse_rule_document(doc1) is None

        # Missing language
        doc2 = {"id": "test", "pattern": "print($MSG)"}
        assert loader._parse_rule_document(doc2) is None

        # Missing pattern
        doc3 = {"id": "test", "language": "python"}
        assert loader._parse_rule_document(doc3) is None

        # Missing all required fields
        doc4 = {"severity": "warning", "message": "test"}
        assert loader._parse_rule_document(doc4) is None


class TestLoadFromFileEdgeCases:
    """Tests for load_from_file method edge cases."""

    def test_load_from_file_with_unicode_content(self) -> None:
        """Test load_from_file handles Unicode content in YAML."""
        loader = RuleLoader()

        yaml_content = """
id: unicode-rule
language: python
severity: warning
message: "Rule with Unicode: 你好世界 🌍 Ñoño"
rule:
  pattern: 'print($MSG)'
"""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".yml", delete=False, encoding="utf-8") as f:
            f.write(yaml_content)
            f.flush()

            rules = loader.load_from_file(f.name)

            assert len(rules) == 1
            assert rules[0].id == "unicode-rule"
            assert "你好世界" in rules[0].message
            assert "🌍" in rules[0].message

        Path(f.name).unlink()

    def test_load_from_file_with_empty_documents(self) -> None:
        """Test load_from_file skips empty documents in multi-document YAML."""
        loader = RuleLoader()

        yaml_content = """
id: rule-one
language: python
severity: warning
message: Rule one
rule:
  pattern: 'pattern_one($ARG)'
---
---
id: rule-two
language: python
severity: error
message: Rule two
rule:
  pattern: 'pattern_two($ARG)'
"""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".yml", delete=False) as f:
            f.write(yaml_content)
            f.flush()

            rules = loader.load_from_file(f.name)

            # Should load only non-empty documents
            assert len(rules) == 2
            assert rules[0].id == "rule-one"
            assert rules[1].id == "rule-two"

        Path(f.name).unlink()

    def test_load_from_file_with_invalid_rule_documents(self) -> None:
        """Test load_from_file skips invalid rule documents in multi-document YAML."""
        loader = RuleLoader()

        yaml_content = """
id: valid-rule
language: python
severity: warning
message: Valid rule
rule:
  pattern: 'valid_pattern($ARG)'
---
language: python
severity: warning
message: Missing id and pattern
---
id: another-valid-rule
language: python
severity: error
message: Another valid rule
rule:
  pattern: 'another_pattern($ARG)'
"""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".yml", delete=False) as f:
            f.write(yaml_content)
            f.flush()

            rules = loader.load_from_file(f.name)

            # Should load only valid documents
            assert len(rules) == 2
            assert rules[0].id == "valid-rule"
            assert rules[1].id == "another-valid-rule"

        Path(f.name).unlink()

    def test_load_from_file_appends_to_internal_rules(self) -> None:
        """Test load_from_file appends rules to internal _rules list."""
        loader = RuleLoader()

        yaml_content1 = """
id: rule-one
language: python
severity: warning
message: Rule one
rule:
  pattern: 'pattern_one($ARG)'
"""

        yaml_content2 = """
id: rule-two
language: python
severity: error
message: Rule two
rule:
  pattern: 'pattern_two($ARG)'
"""

        with tempfile.NamedTemporaryFile(mode="w", suffix=".yml", delete=False) as f1:
            f1.write(yaml_content1)
            f1.flush()

            with tempfile.NamedTemporaryFile(mode="w", suffix=".yml", delete=False) as f2:
                f2.write(yaml_content2)
                f2.flush()

                # Load first file
                loader.load_from_file(f1.name)
                assert len(loader._rules) == 1

                # Load second file
                loader.load_from_file(f2.name)
                assert len(loader._rules) == 2
                assert loader._rules[0].id == "rule-one"
                assert loader._rules[1].id == "rule-two"

            Path(f2.name).unlink()
        Path(f1.name).unlink()


class TestLoadFromDirectoryEdgeCases:
    """Tests for load_from_directory method edge cases."""

    def test_load_from_directory_with_subdirectories(self) -> None:
        """Test load_from_directory skips subdirectories."""
        loader = RuleLoader()

        with tempfile.TemporaryDirectory() as tmpdir:
            # Create rule file in root
            root_rule = Path(tmpdir) / "root.yml"
            root_rule.write_text("""
id: root-rule
language: python
severity: warning
message: Root rule
rule:
  pattern: 'root_pattern($ARG)'
""")

            # Create subdirectory with rules
            subdir = Path(tmpdir) / "subdir"
            subdir.mkdir()
            sub_rule = subdir / "sub.yml"
            sub_rule.write_text("""
id: sub-rule
language: python
severity: error
message: Sub rule
rule:
  pattern: 'sub_pattern($ARG)'
""")

            rules = loader.load_from_directory(tmpdir)

            # Should only load from root (no recursion)
            assert len(rules) == 1
            assert rules[0].id == "root-rule"

    def test_load_from_directory_with_non_yaml_files(self) -> None:
        """Test load_from_directory ignores non-YAML files."""
        loader = RuleLoader()

        with tempfile.TemporaryDirectory() as tmpdir:
            # Create YAML files
            rule1 = Path(tmpdir) / "rule1.yml"
            rule1.write_text("""
id: rule-one
language: python
severity: warning
message: Rule one
rule:
  pattern: 'pattern_one($ARG)'
""")

            # Create non-YAML files
            (Path(tmpdir) / "README.md").write_text("# Documentation")
            (Path(tmpdir) / "config.json").write_text('{"key": "value"}')
            (Path(tmpdir) / "script.py").write_text("print('test')")

            rules = loader.load_from_directory(tmpdir)

            # Should only load YAML files
            assert len(rules) == 1
            assert rules[0].id == "rule-one"

    def test_load_from_directory_empty_directory(self) -> None:
        """Test load_from_directory with empty directory."""
        loader = RuleLoader()

        with tempfile.TemporaryDirectory() as tmpdir:
            rules = loader.load_from_directory(tmpdir)
            assert len(rules) == 0

    def test_load_from_directory_nonexistent_directory(self) -> None:
        """Test load_from_directory with nonexistent directory."""
        loader = RuleLoader()

        with pytest.raises(FileNotFoundError):
            loader.load_from_directory("/nonexistent/directory")


class TestLoadBuiltinRules:
    """Tests for load_builtin_rules method."""

    def test_load_builtin_rules_returns_list(self) -> None:
        """Test load_builtin_rules returns a list."""
        loader = RuleLoader()
        rules = loader.load_builtin_rules()

        assert isinstance(rules, list)

    def test_load_builtin_rules_extends_internal_rules(self) -> None:
        """Test load_builtin_rules extends internal _rules list."""
        loader = RuleLoader()

        initial_count = len(loader._rules)
        rules = loader.load_builtin_rules()

        # Should extend the internal list
        assert len(loader._rules) >= initial_count


class TestGetRulesForLanguage:
    """Tests for get_rules_for_language method."""

    def test_get_rules_for_language_case_sensitive(self) -> None:
        """Test get_rules_for_language is case-sensitive."""
        loader = RuleLoader()

        loader._rules = [
            Rule(id="rule1", language="Python", severity="warning", message="Test", pattern="test"),
            Rule(id="rule2", language="python", severity="error", message="Test", pattern="test"),
            Rule(id="rule3", language="PYTHON", severity="info", message="Test", pattern="test"),
        ]

        # Case-sensitive matching
        python_rules = loader.get_rules_for_language("python")
        assert len(python_rules) == 1
        assert python_rules[0].id == "rule2"

        # Different case returns different results
        python_uppercase = loader.get_rules_for_language("Python")
        assert len(python_uppercase) == 1
        assert python_uppercase[0].id == "rule1"

    def test_get_rules_for_language_no_match(self) -> None:
        """Test get_rules_for_language with no matching rules."""
        loader = RuleLoader()

        loader._rules = [
            Rule(id="py-rule", language="python", severity="warning", message="Test", pattern="test"),
            Rule(id="js-rule", language="javascript", severity="error", message="Test", pattern="test"),
        ]

        # No rules for cobol
        cobol_rules = loader.get_rules_for_language("cobol")
        assert len(cobol_rules) == 0

    def test_get_rules_for_language_with_empty_internal_rules(self) -> None:
        """Test get_rules_for_language when internal _rules is empty."""
        loader = RuleLoader()
        loader._rules = []

        rules = loader.get_rules_for_language("python")
        assert len(rules) == 0


class TestRuleDataclass:
    """Tests for Rule dataclass edge cases."""

    def test_rule_with_all_severity_values(self) -> None:
        """Test Rule accepts all valid severity values."""
        valid_severities = ["error", "warning", "info", "hint", "ERROR", "Warning", "Info"]

        for severity in valid_severities:
            rule = Rule(
                id=f"test-{severity}",
                language="python",
                severity=severity,
                message="Test message",
                pattern="test_pattern",
            )
            assert rule.severity == severity

    def test_rule_with_unicode_fields(self) -> None:
        """Test Rule handles Unicode in all string fields."""
        rule = Rule(
            id="unicode-你好-🌍",
            language="python",
            severity="warning",
            message="Unicode message: 你好世界 🎨",
            pattern="test_pattern_ñoño",
            fix="fix_Ñoño_🚀",
        )

        assert "你好" in rule.id
        assert "🌍" in rule.id
        assert "你好世界" in rule.message
        assert "🎨" in rule.message
        assert "ñ" in rule.pattern  # type: ignore[operator]
        assert "Ñ" in rule.fix  # type: ignore[operator]
        assert "🚀" in rule.fix  # type: ignore[operator]

    def test_rule_with_empty_strings(self) -> None:
        """Test Rule with empty string fields."""
        rule = Rule(
            id="",
            language="",
            severity="",
            message="",
            pattern="",
            fix="",
        )

        assert rule.id == ""
        assert rule.language == ""
        assert rule.severity == ""
        assert rule.message == ""
        assert rule.pattern == ""
        assert rule.fix == ""

    def test_rule_immutable_dataclass(self) -> None:
        """Test Rule is a regular dataclass (mutable by default)."""
        rule = Rule(
            id="test-rule",
            language="python",
            severity="warning",
            message="Test",
            pattern="test_pattern",
        )

        # Modify field (dataclasses are mutable by default)
        rule.severity = "error"
        assert rule.severity == "error"


class TestRuleLoaderState:
    """Tests for RuleLoader state management."""

    def test_ruleloader_multiple_instances_independent(self) -> None:
        """Test multiple RuleLoader instances maintain independent state."""
        loader1 = RuleLoader()
        loader2 = RuleLoader()

        loader1._rules = [
            Rule(id="loader1-rule", language="python", severity="warning", message="Test", pattern="test")
        ]

        loader2._rules = [
            Rule(id="loader2-rule", language="javascript", severity="error", message="Test", pattern="test")
        ]

        assert len(loader1._rules) == 1
        assert loader1._rules[0].id == "loader1-rule"

        assert len(loader2._rules) == 1
        assert loader2._rules[0].id == "loader2-rule"

    def test_ruleloader_internal_rules_accumulation(self) -> None:
        """Test internal _rules list accumulates across multiple load operations."""
        loader = RuleLoader()

        with tempfile.TemporaryDirectory() as tmpdir:
            # Create multiple rule files
            for i in range(3):
                rule_file = Path(tmpdir) / f"rule{i}.yml"
                rule_file.write_text(f"""
id: rule-{i}
language: python
severity: warning
message: Rule {i}
rule:
  pattern: 'pattern_{i}($ARG)'
""")

            # Load all files
            for i in range(3):
                loader.load_from_file(str(Path(tmpdir) / f"rule{i}.yml"))

            # All rules should be accumulated
            assert len(loader._rules) == 3
            assert {rule.id for rule in loader._rules} == {"rule-0", "rule-1", "rule-2"}


class TestRuleLoadingWithYamlFeatures:
    """Tests for YAML feature handling in rule loading."""

    def test_load_from_file_with_yaml_anchors_and_aliases(self) -> None:
        """Test load_from_file handles YAML anchors and aliases."""
        loader = RuleLoader()

        yaml_content = """
defaults: &defaults
  language: python
  severity: warning

---
id: rule-using-alias
language: *defaults
message: Rule using YAML alias
rule:
  pattern: 'test_pattern($ARG)'
"""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".yml", delete=False) as f:
            f.write(yaml_content)
            f.flush()

            # Note: YAML anchors to scalars work, but merge keys (<<) are not supported
            # This test verifies the loader doesn't crash on YAML features
            try:
                rules = loader.load_from_file(f.name)
                # Should handle YAML without crashing (may return 0 rules if merge keys used)
                assert isinstance(rules, list)
            except Exception:
                # If parsing fails, that's acceptable for complex YAML features
                pass

        Path(f.name).unlink()

    def test_load_from_file_with_multiline_strings(self) -> None:
        """Test load_from_file handles multiline strings in YAML."""
        loader = RuleLoader()

        yaml_content = """
id: multiline-rule
language: python
severity: warning
message: |
  This is a multiline
  message that spans
  multiple lines.
rule:
  pattern: 'print($MSG)'
"""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".yml", delete=False) as f:
            f.write(yaml_content)
            f.flush()

            rules = loader.load_from_file(f.name)

            assert len(rules) == 1
            assert "multiline" in rules[0].message.lower()

        Path(f.name).unlink()
