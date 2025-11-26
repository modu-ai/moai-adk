"""
RED Phase Tests for moai-foundation-ears skill.
These tests will initially fail and drive implementation (RED-GREEN-REFACTOR cycle).

Tests cover:
1. EARS pattern parsing (Event, Agent, Result, Scenario)
2. Pattern validation
3. Condition handling
4. Test case generation
5. Requirement analysis
6. Template generation
7. Complete validation workflow
"""

import pytest


# ===== FIXTURES =====
@pytest.fixture
def ears_parser():
    """Create an EARS parser instance."""
    # Will be implemented in GREEN phase
    from moai_adk.foundation.ears import EARSParser

    return EARSParser()


@pytest.fixture
def ears_validator():
    """Create an EARS validator instance."""
    from moai_adk.foundation.ears import EARSValidator

    return EARSValidator()


@pytest.fixture
def ears_analyzer():
    """Create an EARS analyzer instance."""
    from moai_adk.foundation.ears import EARSAnalyzer

    return EARSAnalyzer()


# ===== TEST CLASS 1: Event Pattern Parsing =====
class TestEARSEventPatternParsing:
    """Test Event pattern recognition and parsing."""

    def test_parse_simple_event_pattern(self, ears_parser):
        """Parse basic 'When...Then' event pattern.

        Given: Text with simple event pattern
        When: EARSParser.parse() is called
        Then: Pattern type should be 'event' with trigger and result extracted
        """
        requirement = """
        When a user clicks the 'Submit Order' button,
        then the system shall validate all required fields
        and display a confirmation message.
        """

        result = ears_parser.parse(requirement)

        assert result is not None
        assert result["pattern_type"] == "event"
        assert "trigger" in result
        assert "result" in result
        assert "Submit Order" in result["trigger"].lower() or "submit" in result["trigger"].lower()
        assert "validate" in result["result"].lower() or "confirmation" in result["result"].lower()

    def test_parse_event_with_agent(self, ears_parser):
        """Parse event pattern that includes agent role.

        Given: Event pattern with agent specification
        When: Parser processes the requirement
        Then: Agent role should be identified
        """
        requirement = """
        As a Customer,
        When I login with valid credentials,
        Then the system shall display my dashboard.
        """

        result = ears_parser.parse(requirement)

        assert result["pattern_type"] == "event"
        assert "agent" in result
        assert "customer" in result["agent"].lower()
        assert "trigger" in result
        assert "result" in result

    def test_parse_event_multiple_triggers(self, ears_parser):
        """Parse event pattern with multiple possible triggers.

        Given: Requirement with OR logic in triggers
        When: Parser analyzes multiple conditions
        Then: All triggers should be identified as list
        """
        requirement = """
        When a user clicks 'Submit' OR presses Enter key,
        then the system shall submit the form.
        """

        result = ears_parser.parse(requirement)

        assert result["pattern_type"] == "event"
        assert "triggers" in result or isinstance(result.get("trigger"), list)
        triggers = result.get("triggers") or [result.get("trigger")]
        assert len(triggers) >= 2 or ("Submit" in result.get("trigger", "") and "Enter" in result.get("trigger", ""))


# ===== TEST CLASS 2: Agent Pattern Parsing =====
class TestEARSAgentPatternParsing:
    """Test Agent pattern recognition."""

    def test_parse_agent_capability_pattern(self, ears_parser):
        """Parse 'As a...I shall' agent capability pattern.

        Given: Requirement defining user capability
        When: Parser analyzes agent pattern
        Then: Agent role and capability should be extracted
        """
        requirement = """
        As an Administrator,
        I shall be able to manage user accounts
        and assign role permissions.
        """

        result = ears_parser.parse(requirement)

        assert result["pattern_type"] == "agent"
        assert "agent" in result
        assert "administrator" in result["agent"].lower()
        assert "capability" in result or "action" in result
        capability = result.get("capability") or result.get("action")
        assert "manage" in capability.lower() or "permission" in capability.lower()

    def test_parse_multiple_agent_roles(self, ears_parser):
        """Parse requirements with multiple agent roles.

        Given: Requirement mentioning multiple user roles
        When: Parser extracts agent information
        Then: All agent roles should be identified
        """
        requirement = """
        As a Customer or Guest,
        I shall be able to browse the product catalog.
        """

        result = ears_parser.parse(requirement)

        assert result["pattern_type"] == "agent"
        agents = result.get("agents", [])
        if agents:
            assert len(agents) >= 2
        else:
            assert "customer" in result.get("agent", "").lower()
            assert "guest" in result.get("agent", "").lower()


# ===== TEST CLASS 3: Condition and Result Parsing =====
class TestEARSConditionResultParsing:
    """Test parsing of conditions and expected results."""

    def test_parse_condition_result_pattern(self, ears_parser):
        """Parse 'Where...Then' condition-result pattern.

        Given: Requirement with explicit condition
        When: Parser processes the condition clause
        Then: Condition and result should be separated
        """
        requirement = """
        Where password length is less than 8 characters,
        Then system shall reject the password
        and require a minimum of 8 characters.
        """

        result = ears_parser.parse(requirement)

        assert "condition" in result
        assert "result" in result
        assert "8" in result["condition"] or "password" in result["condition"].lower()
        assert "reject" in result["result"].lower()

    def test_parse_success_and_failure_paths(self, ears_parser):
        """Parse requirement with both success and failure conditions.

        Given: Complex requirement with multiple condition paths
        When: Parser analyzes all branches
        Then: Both success and failure scenarios should be identified
        """
        requirement = """
        When user submits a payment request:

        Success Path:
        Where payment is authorized,
        Then system shall create order and send confirmation.

        Failure Path:
        Where payment is declined,
        Then system shall display error message and retry prompt.
        """

        result = ears_parser.parse(requirement)

        # Should identify at least condition-result pairs
        assert "condition" in result or "conditions" in result
        assert "result" in result or "results" in result


# ===== TEST CLASS 4: Scenario Pattern Parsing =====
class TestEARSScenarioPatternParsing:
    """Test scenario-based requirement parsing."""

    def test_parse_scenario_pattern(self, ears_parser):
        """Parse business scenario requirement.

        Given: Requirement in scenario format
        When: Parser analyzes scenario structure
        Then: Scenario context, trigger, and result extracted
        """
        requirement = """
        Scenario: Monthly invoice generation

        When the first day of the month is reached,
        Where there are pending invoices,
        Then system shall:
        1. Generate invoice PDF
        2. Send via email to customer
        3. Update invoice status to "sent"
        """

        result = ears_parser.parse(requirement)

        assert result["pattern_type"] == "scenario"
        assert "scenario" in result
        assert "monthly" in result["scenario"].lower() or "invoice" in result["scenario"].lower()
        assert "trigger" in result or "when" in result
        assert "result" in result


# ===== TEST CLASS 5: Validation Rules =====
class TestEARSValidationRules:
    """Test requirement validation against EARS rules."""

    def test_validate_complete_requirement(self, ears_validator):
        """Validate a complete, well-formed EARS requirement.

        Given: Complete requirement with all EARS elements
        When: Validator checks completeness
        Then: Validation should pass with no errors
        """
        requirement = """
        As a Customer,
        When I click the checkout button,
        Where my cart contains items,
        Then the system shall display the payment form.
        """

        result = ears_validator.validate(requirement)

        assert result is not None
        assert result.get("is_valid")
        assert len(result.get("errors", [])) == 0

    def test_validate_incomplete_requirement(self, ears_validator):
        """Validate incomplete requirement missing key elements.

        Given: Incomplete requirement missing trigger
        When: Validator checks completeness
        Then: Validation should fail with clear error
        """
        requirement = """
        As a Customer, I shall be able to view my order history.
        """

        result = ears_validator.validate(requirement)

        assert not result.get("is_valid")
        errors = result.get("errors", [])
        missing = result.get("missing_elements", [])
        # Should identify missing 'When' clause
        assert len(errors) > 0 or len(missing) > 0
        assert any("when" in str(e).lower() for e in errors + missing)


# ===== TEST CLASS 6: Test Case Generation =====
class TestEARSTestCaseGeneration:
    """Test automatic test case generation from EARS requirements."""

    def test_generate_test_cases_from_scenario(self, ears_analyzer):
        """Generate test cases from scenario requirement.

        Given: EARS requirement with conditions
        When: Analyzer generates test cases
        Then: Happy path and alternative paths should be identified
        """
        requirement = """
        Scenario: User login

        When user enters email and password,
        Where credentials are valid,
        Then system shall authenticate user and show dashboard.

        Alternative:
        Where credentials are invalid,
        Then system shall show error message.
        """

        test_cases = ears_analyzer.generate_test_cases(requirement)

        assert test_cases is not None
        assert len(test_cases) >= 2  # Happy path + error path
        assert any("valid" in str(tc).lower() for tc in test_cases)
        assert any("invalid" in str(tc).lower() for tc in test_cases)


# ===== TEST CLASS 7: Complete Workflow =====
class TestEARSCompleteWorkflow:
    """Test complete EARS analysis workflow."""

    def test_end_to_end_requirement_processing(self, ears_parser, ears_validator, ears_analyzer):
        """Test complete workflow: parse → validate → generate tests.

        Given: Raw requirement text
        When: Full EARS processing pipeline executes
        Then: Parsed requirement, validation result, and test cases generated
        """
        requirement = """
        As a User,
        When I submit a form with required fields,
        Where all fields are valid,
        Then the system shall save the data
        and display a success message.
        """

        # Step 1: Parse
        parsed = ears_parser.parse(requirement)
        assert parsed is not None
        assert "pattern_type" in parsed

        # Step 2: Validate
        validated = ears_validator.validate(requirement)
        assert validated["is_valid"]

        # Step 3: Generate tests
        tests = ears_analyzer.generate_test_cases(requirement)
        assert len(tests) > 0

        # All steps should succeed for valid requirement
        assert parsed and validated.get("is_valid") and len(tests) > 0

    def test_requirement_analysis_priority_assignment(self, ears_analyzer):
        """Analyze requirement and assign priority based on type.

        Given: Different types of requirements
        When: Analyzer determines priority
        Then: Security requirements > functionality > usability
        """
        security_req = """
        When user logs in,
        Where authentication fails 5 times,
        Then system shall lock account for 30 minutes.
        """

        usability_req = """
        When user hovers over button,
        Then system shall display helpful tooltip.
        """

        security_analysis = ears_analyzer.analyze(security_req)
        usability_analysis = ears_analyzer.analyze(usability_req)

        security_priority = security_analysis.get("priority", 0)
        usability_priority = usability_analysis.get("priority", 0)

        # Security should have higher priority than usability
        assert security_priority >= usability_priority


# ===== TEST CLASS 8: Error Handling =====
class TestEARSErrorHandling:
    """Test error handling for malformed requirements."""

    def test_handle_empty_requirement(self, ears_parser):
        """Handle empty or None requirement gracefully.

        Given: Empty requirement text
        When: Parser processes empty input
        Then: Should return empty dict or raise clear error
        """
        result = ears_parser.parse("")

        # Should either return empty dict or be None
        assert result == {} or result is None or result.get("pattern_type") == "unknown"

    def test_handle_ambiguous_requirement(self, ears_validator):
        """Handle ambiguous requirement with unclear patterns.

        Given: Ambiguous requirement text
        When: Validator processes it
        Then: Should identify ambiguities and provide hints
        """
        ambiguous = """
        The system should do something when something happens.
        """

        result = ears_validator.validate(ambiguous)

        assert not result.get("is_valid")
        # Should provide suggestions
        suggestions = result.get("suggestions", [])
        errors = result.get("errors", [])
        assert len(suggestions) > 0 or len(errors) > 0


# ===== INTEGRATION TESTS =====
class TestEARSIntegration:
    """Integration tests for EARS functionality."""

    def test_real_world_api_requirement(self, ears_parser, ears_validator):
        """Test real-world API requirement."""
        requirement = """
        API Endpoint: POST /api/v1/orders

        As a Client Application,
        When posting a valid order JSON,
        Where user is authenticated AND cart has items,
        Then system shall:
        - Return HTTP 201 (Created)
        - Response includes: {id, status, total, created_at}
        - Send confirmation email to customer
        - Update inventory
        """

        parsed = ears_parser.parse(requirement)
        validated = ears_validator.validate(requirement)

        assert parsed is not None
        assert validated.get("is_valid")

    def test_real_world_business_requirement(self, ears_parser, ears_validator):
        """Test real-world business requirement."""
        requirement = """
        Scenario: Monthly report generation

        When the last day of the month is reached,
        Where user has enabled automatic reporting,
        Then system shall:
        1. Aggregate monthly metrics
        2. Generate PDF report
        3. Send report to user email
        4. Archive report in document management
        """

        parsed = ears_parser.parse(requirement)
        validated = ears_validator.validate(requirement)

        assert parsed is not None
        assert parsed["pattern_type"] == "scenario"
        assert validated.get("is_valid")
