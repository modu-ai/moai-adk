"""
EARS (Event-Agent-Result-Scenario) Implementation.

Provides:
- EARSParser: Parse natural language requirements into EARS patterns
- EARSValidator: Validate EARS requirement completeness and clarity
- EARSAnalyzer: Analyze requirements and generate test cases

Reference: EARS methodology for structured requirement specification
"""

import re
from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Dict, List, Optional


class EARSPatternType(Enum):
    """Enumeration of EARS pattern types."""

    EVENT = "event"
    AGENT = "agent"
    SCENARIO = "scenario"
    VALIDATION = "validation"
    COMPLEX = "complex"
    UNKNOWN = "unknown"


@dataclass
class EARSResult:
    """Result of EARS parsing or analysis."""

    pattern_type: str
    trigger: Optional[str] = None
    triggers: List[str] = field(default_factory=list)
    event: Optional[str] = None
    agent: Optional[str] = None
    agents: List[str] = field(default_factory=list)
    condition: Optional[str] = None
    conditions: List[str] = field(default_factory=list)
    result: Optional[str] = None
    results: List[str] = field(default_factory=list)
    scenario: Optional[str] = None
    capability: Optional[str] = None  # For agent patterns
    action: Optional[str] = None  # Alias for capability
    priority: int = 5
    is_valid: bool = True
    errors: List[str] = field(default_factory=list)
    missing_elements: List[str] = field(default_factory=list)

    def __getitem__(self, key):
        """Allow dict-like access for backward compatibility."""
        if isinstance(key, str):
            return getattr(self, key, None)
        # Handle integer indexing - convert to attribute access
        return None

    def __contains__(self, key):
        """Support 'in' operator for dict-like behavior."""
        return hasattr(self, key) and getattr(self, key) is not None

    def get(self, key, default=None):
        """Allow dict.get() compatibility."""
        return getattr(self, key, default)


class EARSParser:
    """Parse natural language requirements into EARS patterns."""

    # Regex patterns for EARS components
    WHEN_PATTERN = re.compile(r"when\s+([^,\.]*?)(?:,|where|then|\.|$)", re.IGNORECASE)
    AGENT_PATTERN = re.compile(r"as\s+(?:an?\s+)?([^,\n]+?)(?:,|\n|when|shall|i\s+shall|\.|$)", re.IGNORECASE)
    WHERE_PATTERN = re.compile(r"where\s+([^,\.]*?)(?:,|then|$)", re.IGNORECASE)
    THEN_PATTERN = re.compile(r"then\s+([^\.]*?)(?:\.|$)", re.IGNORECASE)
    SCENARIO_PATTERN = re.compile(r"scenario\s*:\s*([^\n]+)", re.IGNORECASE)

    def parse(self, requirement: str) -> EARSResult:
        """
        Parse requirement text into EARS components.

        Args:
            requirement: Natural language requirement text

        Returns:
            EARSResult: Parsed requirement with identified pattern and components
        """
        if not requirement or not requirement.strip():
            return EARSResult(pattern_type="unknown")

        requirement = requirement.strip()

        # Try to identify pattern type
        pattern_type = self._identify_pattern_type(requirement)

        # Extract components based on pattern
        result = EARSResult(pattern_type=pattern_type)

        # Extract EARS components
        self._extract_agents(requirement, result)
        self._extract_triggers(requirement, result)
        self._extract_conditions(requirement, result)
        self._extract_results(requirement, result)
        self._extract_scenario(requirement, result)

        return result

    def _identify_pattern_type(self, text: str) -> str:
        """Identify which EARS pattern type the requirement follows."""
        text_lower = text.lower()

        # Check for scenario first (most specific)
        if re.search(r"scenario\s*:", text_lower):
            return EARSPatternType.SCENARIO.value

        # Check for when...then (EVENT) - higher priority
        has_when = re.search(r"when\s+[a-z]", text_lower)
        has_then = re.search(r"then\s+[a-z]", text_lower)
        if has_when and has_then:
            return EARSPatternType.EVENT.value

        # Check for where...then (VALIDATION)
        has_where = re.search(r"where\s+[a-z]", text_lower)
        if has_where and has_then:
            return EARSPatternType.VALIDATION.value

        # Check for as a...shall/can (AGENT)
        has_agent = re.search(r"as\s+(?:an?\s+)?[a-z]", text_lower)
        has_capability = re.search(r"shall|can|able|be able", text_lower)
        if has_agent and has_capability and not has_when:
            return EARSPatternType.AGENT.value

        # Special case: as a + when + then = EVENT with agent
        if has_agent and has_when and has_then:
            return EARSPatternType.EVENT.value

        return EARSPatternType.UNKNOWN.value

    def _extract_agents(self, text: str, result: EARSResult) -> None:
        """Extract agent/role information."""
        matches = self.AGENT_PATTERN.findall(text)
        if matches:
            if len(matches) == 1:
                result.agent = matches[0].strip()
            else:
                result.agents = [m.strip() for m in matches]
                if matches:
                    result.agent = matches[0].strip()

        # Extract capability/action for agent patterns
        # Look for: shall <action>, can <action>, able to <action>
        capability_pattern = re.compile(r"(?:shall|can|able to)\s+([^\.]*?)(?:\.|$)", re.IGNORECASE)
        cap_matches = capability_pattern.findall(text)
        if cap_matches:
            # For agent patterns, treat first capability as the result and capability
            capability = cap_matches[0].strip() if cap_matches else None
            if capability:
                result.result = capability
                result.capability = capability
                result.action = capability

    def _extract_triggers(self, text: str, result: EARSResult) -> None:
        """Extract trigger/event information."""
        matches = self.WHEN_PATTERN.findall(text)
        if matches:
            triggers = [m.strip() for m in matches if m.strip()]
            if triggers:
                if len(triggers) == 1:
                    result.trigger = triggers[0]
                else:
                    result.triggers = triggers
                    result.trigger = triggers[0]

    def _extract_conditions(self, text: str, result: EARSResult) -> None:
        """Extract condition information."""
        matches = self.WHERE_PATTERN.findall(text)
        if matches:
            conditions = [m.strip() for m in matches if m.strip()]
            if conditions:
                if len(conditions) == 1:
                    result.condition = conditions[0]
                else:
                    result.conditions = conditions
                    result.condition = conditions[0]

    def _extract_results(self, text: str, result: EARSResult) -> None:
        """Extract result/outcome information."""
        matches = self.THEN_PATTERN.findall(text)
        if matches:
            results = [m.strip() for m in matches if m.strip()]
            if results:
                if len(results) == 1:
                    result.result = results[0]
                else:
                    result.results = results
                    result.result = results[0]

    def _extract_scenario(self, text: str, result: EARSResult) -> None:
        """Extract scenario context if present."""
        match = self.SCENARIO_PATTERN.search(text)
        if match:
            result.scenario = match.group(1).strip()


class EARSValidator:
    """Validate EARS requirements for completeness and clarity."""

    # Required elements for each pattern type
    PATTERN_REQUIREMENTS = {
        "event": ["trigger", "result"],
        "agent": [
            "agent",
            "trigger",
            "result",
        ],  # Agent patterns should have trigger (When does agent use capability?)
        "scenario": ["scenario", "result"],
        "validation": ["condition", "result"],
        "complex": ["trigger", "condition", "result"],
    }

    # Keywords that indicate missing elements
    WEAK_KEYWORDS = ["should", "maybe", "might", "could", "might"]

    def validate(self, requirement: str) -> Dict[str, Any]:
        """
        Validate EARS requirement completeness.

        Args:
            requirement: Requirement text to validate

        Returns:
            Dict with validation results: {is_valid, errors, missing_elements, suggestions}
        """
        parser = EARSParser()
        parsed = parser.parse(requirement)

        errors = []
        missing_elements = []
        suggestions = []
        is_valid = True

        # Check if requirement is empty
        if not requirement or not requirement.strip():
            return {
                "is_valid": False,
                "errors": ["Requirement text is empty"],
                "missing_elements": ["When", "Then"],
                "suggestions": ["Provide a complete EARS requirement with When and Then clauses"],
            }

        # Check for required elements based on pattern type
        pattern_type = parsed.pattern_type
        req_lower = requirement.lower()

        # Check for basic EARS keywords if pattern was identified
        if pattern_type != "unknown":
            required = self.PATTERN_REQUIREMENTS.get(pattern_type, [])
            for element in required:
                if not getattr(parsed, element, None):
                    # Use 'When' for 'trigger', 'Where' for 'condition', etc.
                    element_name = {
                        "trigger": "When",
                        "condition": "Where",
                        "result": "Then",
                        "agent": "Agent",
                        "scenario": "Scenario",
                    }.get(element, element.capitalize())
                    missing_elements.append(element_name)
                    is_valid = False
        else:
            # For unknown patterns, suggest missing keywords
            if "when" not in req_lower and "as a" not in req_lower:
                missing_elements.append("When")
                is_valid = False
            if "then" not in req_lower:
                missing_elements.append("Then")
                is_valid = False

        # Check for vague language
        for weak in self.WEAK_KEYWORDS:
            if weak in req_lower:
                errors.append(f"Weak keyword '{weak}' found - use 'shall' instead")
                suggestions.append(f"Replace '{weak}' with 'shall'")
                is_valid = False

        # Check if pattern was identified
        if pattern_type == "unknown":
            errors.append("Could not identify EARS pattern (Event/Agent/Scenario/Validation)")
            suggestions.append("Use clear EARS keywords: When, As a, Where, Then")
            is_valid = False

        # Suggest improvements
        if not parsed.trigger and not parsed.agent:
            suggestions.append("Add 'When...' (event) or 'As a...' (agent) clause")
            if "When" not in missing_elements:
                missing_elements.append("When")
        if not parsed.result:
            suggestions.append("Add 'Then...' clause describing expected result")

        return {
            "is_valid": is_valid,
            "errors": errors,
            "missing_elements": missing_elements,
            "suggestions": suggestions,
        }

    def analyze(self, requirement: str) -> Dict[str, Any]:
        """
        Analyze requirement and extract metadata.

        Args:
            requirement: Requirement text

        Returns:
            Dict with analysis results including priority
        """
        validation = self.validate(requirement)
        req_lower = requirement.lower()

        # Assign priority based on requirement type
        priority = 5  # Default

        if any(word in req_lower for word in ["security", "authentication", "authorization", "encrypt"]):
            priority = 8
        elif any(word in req_lower for word in ["performance", "optimize", "fast"]):
            priority = 6
        elif any(word in req_lower for word in ["user experience", "ui", "tooltip", "hover"]):
            priority = 4
        elif any(word in req_lower for word in ["core", "critical", "essential"]):
            priority = 9
        elif any(word in req_lower for word in ["nice to have", "optional"]):
            priority = 2

        return {
            "is_valid": validation["is_valid"],
            "priority": priority,
            "errors": validation["errors"],
            "suggestions": validation["suggestions"],
        }


class EARSAnalyzer:
    """Analyze EARS requirements and generate test cases."""

    def generate_test_cases(self, requirement: str) -> List[Dict[str, str]]:
        """
        Generate test cases from EARS requirement.

        Args:
            requirement: EARS requirement text

        Returns:
            List of test case dictionaries with given/when/then structure
        """
        parser = EARSParser()
        parsed = parser.parse(requirement)

        test_cases = []

        # Happy path test
        if parsed.trigger and parsed.result:
            happy_test = {
                "name": "Happy Path",
                "given": f"A requirement with trigger: {parsed.trigger[:50]}",
                "when": parsed.trigger[:100] if parsed.trigger else "trigger occurs",
                "then": parsed.result[:100] if parsed.result else "result occurs",
            }
            test_cases.append(happy_test)

        # Condition-based test cases
        if parsed.condition:
            valid_condition = {
                "name": "Valid Condition Test",
                "given": f"Condition is met: {parsed.condition[:50]}",
                "when": parsed.trigger[:100] if parsed.trigger else "trigger occurs",
                "then": f"System {parsed.result[:80] if parsed.result else 'responds appropriately'}",
            }
            test_cases.append(valid_condition)

            # Also create inverse condition test
            invalid_condition = {
                "name": "Invalid Condition Test",
                "given": f"Condition is NOT met: {parsed.condition[:50]}",
                "when": parsed.trigger[:100] if parsed.trigger else "trigger occurs",
                "then": "System handles error appropriately",
            }
            test_cases.append(invalid_condition)

        # If no test cases were generated, create a basic one
        if not test_cases:
            basic_test = {
                "name": "Basic Test",
                "given": "Requirement is triggered",
                "when": parsed.trigger or "event occurs",
                "then": parsed.result or "expected behavior occurs",
            }
            test_cases.append(basic_test)

        return test_cases

    def analyze(self, requirement: str) -> Dict[str, Any]:
        """
        Complete analysis of EARS requirement.

        Args:
            requirement: Requirement text

        Returns:
            Dict with comprehensive analysis
        """
        parser = EARSParser()
        validator = EARSValidator()

        parsed = parser.parse(requirement)
        validation = validator.analyze(requirement)
        test_cases = self.generate_test_cases(requirement)

        return {
            "parsed": {
                "pattern_type": parsed.pattern_type,
                "agent": parsed.agent,
                "trigger": parsed.trigger,
                "condition": parsed.condition,
                "result": parsed.result,
            },
            "priority": validation["priority"],
            "is_valid": validation["is_valid"],
            "errors": validation["errors"],
            "suggestions": validation["suggestions"],
            "test_cases": test_cases,
            "test_count": len(test_cases),
        }
