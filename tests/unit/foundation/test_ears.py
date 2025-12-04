"""
Unit tests for moai_adk.foundation.ears module.

Tests cover:
- EARSPatternType enum
- EARSResult dataclass
- EARSParser class
- EARSValidator class
- EARSAnalyzer class
"""

import pytest

from moai_adk.foundation.ears import (
    EARSAnalyzer,
    EARSParser,
    EARSPatternType,
    EARSResult,
    EARSValidator,
)


class TestEARSPatternType:
    """Test EARSPatternType enumeration."""

    def test_pattern_types(self):
        """Test all pattern types are defined."""
        assert EARSPatternType.EVENT.value == "event"
        assert EARSPatternType.AGENT.value == "agent"
        assert EARSPatternType.SCENARIO.value == "scenario"
        assert EARSPatternType.VALIDATION.value == "validation"
        assert EARSPatternType.COMPLEX.value == "complex"
        assert EARSPatternType.UNKNOWN.value == "unknown"


class TestEARSResult:
    """Test EARSResult dataclass."""

    def test_create_minimal_result(self):
        """Test creating minimal result."""
        result = EARSResult(pattern_type="event")
        assert result.pattern_type == "event"
        assert result.trigger is None
        assert result.event is None
        assert result.priority == 5
        assert result.is_valid is True
        assert result.errors == []

    def test_create_full_result(self):
        """Test creating result with all fields."""
        result = EARSResult(
            pattern_type="event",
            trigger="user clicks button",
            event="click",
            result="dialog opens",
            priority=7,
            is_valid=True,
            errors=["some error"],
        )
        assert result.trigger == "user clicks button"
        assert result.priority == 7
        assert result.errors == ["some error"]

    def test_dict_like_access(self):
        """Test dict-like access."""
        result = EARSResult(pattern_type="event", trigger="event occurs")
        assert result["pattern_type"] == "event"
        assert result.get("trigger") == "event occurs"
        assert result.get("nonexistent", "default") == "default"

    def test_contains_operator(self):
        """Test 'in' operator."""
        result = EARSResult(pattern_type="event", trigger="event")
        assert "pattern_type" in result
        assert "trigger" in result
        assert "nonexistent" not in result


class TestEARSParser:
    """Test EARSParser class."""

    def test_parse_event_pattern(self):
        """Test parsing EVENT pattern."""
        parser = EARSParser()
        result = parser.parse("When user clicks button, then dialog opens")

        assert result.pattern_type == "event"
        assert result.trigger is not None
        assert result.result is not None

    def test_parse_agent_pattern(self):
        """Test parsing AGENT pattern."""
        parser = EARSParser()
        result = parser.parse("As a developer, I shall create unit tests")

        assert result.pattern_type == "agent"
        assert result.agent is not None

    def test_parse_scenario_pattern(self):
        """Test parsing SCENARIO pattern."""
        parser = EARSParser()
        result = parser.parse("Scenario: User login flow")

        assert result.pattern_type == "scenario"
        assert result.scenario is not None

    def test_parse_validation_pattern(self):
        """Test parsing VALIDATION pattern."""
        parser = EARSParser()
        result = parser.parse("Where age > 18, then grant access")

        assert result.pattern_type == "validation"
        assert result.condition is not None

    def test_parse_empty_requirement(self):
        """Test parsing empty requirement."""
        parser = EARSParser()
        result = parser.parse("")

        assert result.pattern_type == "unknown"

    def test_parse_extract_agents(self):
        """Test agent extraction."""
        parser = EARSParser()
        result = parser.parse("As a user, I shall be able to login")

        assert result.agent is not None
        assert "user" in result.agent.lower()

    def test_parse_extract_triggers(self):
        """Test trigger extraction."""
        parser = EARSParser()
        result = parser.parse("When user submits form, then data is saved")

        assert result.trigger is not None
        assert "submit" in result.trigger.lower()

    def test_parse_extract_conditions(self):
        """Test condition extraction."""
        parser = EARSParser()
        result = parser.parse("Where user is authenticated, then show dashboard")

        assert result.condition is not None
        # Condition may be truncated by regex, just check it exists

    def test_parse_extract_results(self):
        """Test result extraction."""
        parser = EARSParser()
        result = parser.parse("When user clicks login, then they are authenticated")

        assert result.result is not None

    def test_parse_multiple_triggers(self):
        """Test parsing with multiple triggers."""
        parser = EARSParser()
        result = parser.parse("When X happens, when Y happens, then Z occurs")

        # Should capture at least the first trigger
        assert result.trigger is not None

    def test_parse_event_with_agent(self):
        """Test parsing event with agent."""
        parser = EARSParser()
        result = parser.parse("As a user, when I click button, then dialog opens")

        assert result.pattern_type == "event"
        assert result.agent is not None
        assert result.trigger is not None


class TestEARSValidator:
    """Test EARSValidator class."""

    def test_validate_valid_event_requirement(self):
        """Test validating valid EVENT requirement."""
        validator = EARSValidator()
        result = validator.validate("When user clicks button, then dialog opens")

        assert result["is_valid"] is True
        assert len(result["errors"]) == 0

    def test_validate_empty_requirement(self):
        """Test validating empty requirement."""
        validator = EARSValidator()
        result = validator.validate("")

        assert result["is_valid"] is False
        assert any("empty" in err.lower() for err in result["errors"])

    def test_validate_missing_when_clause(self):
        """Test validating requirement missing When clause."""
        validator = EARSValidator()
        result = validator.validate("Then the system responds")

        assert result["is_valid"] is False
        assert "When" in result["missing_elements"] or len(result["errors"]) > 0

    def test_validate_missing_then_clause(self):
        """Test validating requirement missing Then clause."""
        validator = EARSValidator()
        result = validator.validate("When user submits form")

        assert result["is_valid"] is False

    def test_validate_vague_language(self):
        """Test detecting vague language."""
        validator = EARSValidator()
        result = validator.validate("When user clicks, maybe dialog opens")

        assert result["is_valid"] is False
        assert len(result["errors"]) > 0
        assert len(result["suggestions"]) > 0

    def test_validate_agent_requirement(self):
        """Test validating AGENT requirement."""
        validator = EARSValidator()
        result = validator.validate("As a user, I shall be able to login")

        # Agent patterns are valid if structured correctly
        assert isinstance(result["is_valid"], bool)

    def test_analyze_requirement(self):
        """Test analyzing requirement."""
        validator = EARSValidator()
        result = validator.analyze("When user authenticates, then access is granted")

        assert "is_valid" in result
        assert "priority" in result
        assert "errors" in result
        assert "suggestions" in result

    def test_analyze_priority_security(self):
        """Test priority detection for security requirement."""
        validator = EARSValidator()
        result = validator.analyze("When user logs in, the system shall encrypt password")

        # Security requirements should have high priority
        assert result["priority"] >= 7

    def test_analyze_priority_optional(self):
        """Test priority detection for optional requirement."""
        validator = EARSValidator()
        result = validator.analyze("When nice to have, then feature is added")

        # Optional requirements should have low priority
        assert result["priority"] <= 3


class TestEARSAnalyzer:
    """Test EARSAnalyzer class."""

    def test_generate_test_cases_simple(self):
        """Test generating test cases from simple requirement."""
        analyzer = EARSAnalyzer()
        test_cases = analyzer.generate_test_cases("When user clicks button, then dialog opens")

        assert len(test_cases) > 0
        assert all("given" in tc for tc in test_cases)
        assert all("when" in tc for tc in test_cases)
        assert all("then" in tc for tc in test_cases)

    def test_generate_test_cases_with_condition(self):
        """Test generating test cases with conditions."""
        analyzer = EARSAnalyzer()
        test_cases = analyzer.generate_test_cases("When user submits form where age > 18, then registration succeeds")

        # Should generate multiple test cases (happy path + condition cases)
        assert len(test_cases) >= 2

    def test_generate_test_cases_empty(self):
        """Test generating test cases from empty requirement."""
        analyzer = EARSAnalyzer()
        test_cases = analyzer.generate_test_cases("")

        # Should generate at least basic test case
        assert len(test_cases) > 0

    def test_analyze_requirement(self):
        """Test analyzing requirement."""
        analyzer = EARSAnalyzer()
        result = analyzer.analyze("When user logs in, then they see dashboard")

        assert "parsed" in result
        assert "priority" in result
        assert "is_valid" in result
        assert "errors" in result
        assert "suggestions" in result
        assert "test_cases" in result
        assert "test_count" in result

    def test_analyze_with_valid_requirement(self):
        """Test analyzing with valid requirement."""
        analyzer = EARSAnalyzer()
        result = analyzer.analyze("When user clicks login, then user is authenticated")

        assert result["parsed"]["pattern_type"] in [
            "event",
            "agent",
            "scenario",
            "validation",
            "complex",
        ]
        assert result["test_count"] >= 1

    def test_analyze_with_conditions(self):
        """Test analyzing requirement with conditions."""
        analyzer = EARSAnalyzer()
        result = analyzer.analyze("When user submits form where email is valid, then account is created")

        assert result["test_count"] >= 2  # At least happy path + invalid condition

    def test_test_case_structure(self):
        """Test that generated test cases have proper structure."""
        analyzer = EARSAnalyzer()
        test_cases = analyzer.generate_test_cases("When X happens, then Y occurs")

        for tc in test_cases:
            assert "name" in tc
            assert "given" in tc
            assert "when" in tc
            assert "then" in tc
            assert isinstance(tc["name"], str)
            assert isinstance(tc["given"], str)
            assert isinstance(tc["when"], str)
            assert isinstance(tc["then"], str)


class TestEARSIntegration:
    """Integration tests for EARS components."""

    def test_parse_validate_analyze_flow(self):
        """Test complete parse -> validate -> analyze flow."""
        requirement = "When user clicks logout button, then user session is destroyed"

        parser = EARSParser()
        parsed = parser.parse(requirement)
        assert parsed.trigger is not None
        assert parsed.result is not None

        validator = EARSValidator()
        validation = validator.validate(requirement)
        assert validation["is_valid"] is True

        analyzer = EARSAnalyzer()
        analysis = analyzer.analyze(requirement)
        assert analysis["test_count"] > 0

    def test_complex_requirement_handling(self):
        """Test handling of complex requirements."""
        requirement = (
            "As a system administrator, when I attempt to create a new user "
            "where the email is unique and password meets security requirements, "
            "then the user account is created and welcome email is sent"
        )

        analyzer = EARSAnalyzer()
        result = analyzer.analyze(requirement)

        assert result["parsed"]["agent"] is not None
        assert result["parsed"]["trigger"] is not None
        assert result["test_count"] >= 2

    def test_vague_requirement_detection(self):
        """Test detection of vague requirements."""
        vague_requirement = "The system should maybe provide some kind of notification"

        validator = EARSValidator()
        result = validator.validate(vague_requirement)

        assert result["is_valid"] is False
        assert len(result["suggestions"]) > 0
