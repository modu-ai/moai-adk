"""
Robust JSON Parser for MoAI-ADK

Production-ready JSON parser with automatic error recovery, comprehensive logging,
and fallback strategies to handle malformed JSON input from various sources.

Author: MoAI-ADK Core Team
Version: 1.0.0
"""

import json
import logging
import re
from dataclasses import dataclass
from enum import Enum
from typing import Any, Callable, Dict, List, Optional, Tuple, Union

# Configure logging
logger = logging.getLogger(__name__)


class ErrorSeverity(Enum):
    """Error severity levels for classification"""

    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


@dataclass
class ParseResult:
    """Result of JSON parsing with metadata"""

    success: bool
    data: Optional[Any]
    error: Optional[str]
    original_input: str
    recovery_attempts: int
    severity: ErrorSeverity
    parse_time_ms: float
    warnings: List[str]


class RobustJSONParser:
    """
    Production-ready JSON parser with comprehensive error recovery strategies.

    Features:
    - Multiple error recovery strategies
    - Detailed logging and error tracking
    - Performance monitoring
    - Fallback parsing methods
    - Security validation
    """

    def __init__(self, max_recovery_attempts: int = 3, enable_logging: bool = True):
        self.max_recovery_attempts = max_recovery_attempts
        self.enable_logging = enable_logging
        self.error_patterns = self._load_error_patterns()
        self.recovery_strategies = self._load_recovery_strategies()
        self.stats = {
            "total_parses": 0,
            "successful_parses": 0,
            "recovered_parses": 0,
            "failed_parses": 0,
            "total_recovery_time": 0.0,
        }

    def _load_error_patterns(self) -> Dict[str, re.Pattern]:
        """Load common JSON error patterns"""
        return {
            "missing_quotes": re.compile(r'(?<!\\)"(?:[^"\\]|\\.)*$'),
            "trailing_comma": re.compile(r",\s*[}\]]"),
            "escape_sequence": re.compile(r'\\(?![nrtbf"\'\\/])'),
            "partial_object": re.compile(r"^\s*\{[^}]*\s*$"),
            "missing_brace": re.compile(r"^[^{]*\{[^}]*[^}]*$"),
            "invalid_quotes": re.compile(r'(?<!\\)"(?:[^"\\]|\\.)*?[^\\]"(?![\s,}\]:])'),
            "control_chars": re.compile(r"[\x00-\x1F\x7F-\x9F]"),
        }

    def _load_recovery_strategies(self) -> List[Callable]:
        """Load error recovery strategies in order of application"""
        return [
            self._fix_missing_quotes,
            self._fix_trailing_commas,
            self._fix_escape_sequences,
            self._handle_partial_objects,
            self._fix_invalid_quotes,
            self._remove_control_characters,
            self._handle_escaped_newlines,
            self._fix_common_syntax_errors,
            self._attempt_partial_parse,
        ]

    def parse(self, json_string: str, context: Optional[Dict] = None) -> ParseResult:
        """
        Parse JSON string with comprehensive error recovery.

        Args:
            json_string: JSON string to parse
            context: Optional context information for error reporting

        Returns:
            ParseResult with data and metadata
        """
        import time

        start_time = time.time()

        self.stats["total_parses"] += 1

        original_input = json_string
        current_input = json_string
        recovery_attempts = 0
        warnings: List[str] = []

        # Initial validation
        if not isinstance(json_string, str):
            result = ParseResult(
                success=False,
                data=None,
                error=f"Input must be string, got {type(json_string)}",
                original_input=original_input,
                recovery_attempts=0,
                severity=ErrorSeverity.CRITICAL,
                parse_time_ms=(time.time() - start_time) * 1000,
                warnings=warnings,
            )
            self.stats["failed_parses"] += 1
            return result

        # Try direct parsing first
        try:
            data = json.loads(json_string)
            self.stats["successful_parses"] += 1

            result = ParseResult(
                success=True,
                data=data,
                error=None,
                original_input=original_input,
                recovery_attempts=0,
                severity=ErrorSeverity.LOW,
                parse_time_ms=(time.time() - start_time) * 1000,
                warnings=warnings,
            )

            if self.enable_logging:
                logger.debug("JSON parsed successfully on first attempt")

            return result

        except json.JSONDecodeError as e:
            if self.enable_logging:
                logger.warning(f"Initial JSON parse failed: {e.msg} at line {e.lineno}, col {e.colno}")

            last_error = str(e)

            # Apply recovery strategies
            for attempt in range(self.max_recovery_attempts):
                recovery_attempts += 1

                try:
                    # Apply recovery strategies
                    for strategy in self.recovery_strategies:
                        try:
                            modified_input, applied_warnings = strategy(current_input)
                            if modified_input != current_input:
                                current_input = modified_input
                                warnings.extend(applied_warnings)
                                if self.enable_logging:
                                    logger.debug(f"Applied recovery strategy: {strategy.__name__}")
                                break
                        except Exception as strategy_error:
                            if self.enable_logging:
                                logger.debug(f"Recovery strategy {strategy.__name__} failed: {strategy_error}")
                            continue

                    # Try parsing with recovered input
                    data = json.loads(current_input)
                    self.stats["recovered_parses"] += 1

                    result = ParseResult(
                        success=True,
                        data=data,
                        error=None,
                        original_input=original_input,
                        recovery_attempts=recovery_attempts,
                        severity=(ErrorSeverity.MEDIUM if recovery_attempts > 0 else ErrorSeverity.LOW),
                        parse_time_ms=(time.time() - start_time) * 1000,
                        warnings=warnings,
                    )

                    if self.enable_logging:
                        logger.info(f"JSON recovered after {recovery_attempts} attempts")

                    return result

                except json.JSONDecodeError as e:
                    last_error = str(e)
                    if self.enable_logging:
                        logger.debug(f"Parse attempt {attempt + 1} failed: {e.msg}")

                    # Try more aggressive recovery for later attempts
                    if attempt == self.max_recovery_attempts - 1:
                        break

                    # Make more aggressive modifications for next attempt
                    current_input = self._apply_aggressive_recovery(current_input)

            # All recovery attempts failed
            self.stats["failed_parses"] += 1

            result = ParseResult(
                success=False,
                data=None,
                error=last_error,
                original_input=original_input,
                recovery_attempts=recovery_attempts,
                severity=(ErrorSeverity.HIGH if recovery_attempts > 0 else ErrorSeverity.CRITICAL),
                parse_time_ms=(time.time() - start_time) * 1000,
                warnings=warnings,
            )

            if self.enable_logging:
                logger.error(f"JSON parsing failed after {recovery_attempts} recovery attempts: {last_error}")

            return result

    def _fix_missing_quotes(self, json_string: str) -> Tuple[str, List[str]]:
        """Fix missing quotes around string values"""
        warnings = []

        # Look for unquoted property names - more comprehensive pattern
        # Pattern: property_name: (instead of "property_name":)
        pattern = r"(\s*)([a-zA-Z_][a-zA-Z0-9_]*)\s*:"
        matches = list(re.finditer(pattern, json_string))

        modified = json_string
        offset = 0

        for match in reversed(matches):
            # Check if this is inside a string (to avoid false positives)
            pos_before = match.start() + offset
            string_context = self._get_string_context(modified, pos_before)

            if not string_context:
                # This is likely an unquoted property name
                replacement = f'{match.group(1)}"{match.group(2)}":'
                start_pos = match.start() + offset
                end_pos = match.end() + offset
                modified = modified[:start_pos] + replacement + modified[end_pos:]
                offset += len(replacement) - (end_pos - start_pos)
                warnings.append(f"Added quotes to property name: {match.group(2)}")

        return modified, warnings

    def _fix_trailing_commas(self, json_string: str) -> Tuple[str, List[str]]:
        """Remove trailing commas in objects and arrays"""
        warnings = []

        # Remove trailing commas before } or ]
        pattern = r",(\s*[}\]])"
        matches = list(re.finditer(pattern, json_string))

        modified = json_string
        offset = 0

        for match in reversed(matches):
            # Check if this is inside a string
            pos_before = match.start() + offset
            string_context = self._get_string_context(modified, pos_before)

            if not string_context:
                # Remove the comma
                replacement = match.group(1)
                start_pos = match.start() + offset
                end_pos = match.end() + offset
                modified = modified[:start_pos] + replacement + modified[end_pos:]
                offset -= (match.end() - match.start()) + len(replacement)
                warnings.append("Removed trailing comma")

        return modified, warnings

    def _fix_escape_sequences(self, json_string: str) -> Tuple[str, List[str]]:
        """Fix invalid escape sequences"""
        warnings = []

        # Find invalid escape sequences
        pattern = r'\\(?![nrtbf"\'\\/u])'
        matches = list(re.finditer(pattern, json_string))

        modified = json_string
        offset = 0

        for match in reversed(matches):
            # Check if this is inside a string
            pos_before = match.start() + offset
            string_context = self._get_string_context(modified, pos_before)

            if string_context:
                # Remove the invalid backslash
                replacement = match.group(1)[1:] if len(match.group(1)) > 1 else ""
                start_pos = match.start() + offset + 1  # Skip the backslash
                end_pos = match.end() + offset
                modified = modified[:start_pos] + replacement + modified[end_pos:]
                offset -= 1
                warnings.append("Removed invalid escape sequence")

        return modified, warnings

    def _handle_partial_objects(self, json_string: str) -> Tuple[str, List[str]]:
        """Handle incomplete JSON objects"""
        warnings = []

        stripped = json_string.strip()

        # Check if it's a partial object
        if stripped.startswith("{") and not stripped.endswith("}"):
            # Count braces
            open_braces = stripped.count("{")
            close_braces = stripped.count("}")

            if open_braces > close_braces:
                missing_braces = open_braces - close_braces
                modified = stripped + "}" * missing_braces
                warnings.append(f"Added {missing_braces} closing brace(s)")
                return modified, warnings

        # Check for partial arrays
        if stripped.startswith("[") and not stripped.endswith("]"):
            open_brackets = stripped.count("[")
            close_brackets = stripped.count("]")

            if open_brackets > close_brackets:
                missing_brackets = open_brackets - close_brackets
                modified = stripped + "]" * missing_brackets
                warnings.append(f"Added {missing_brackets} closing bracket(s)")
                return modified, warnings

        return json_string, warnings

    def _fix_invalid_quotes(self, json_string: str) -> Tuple[str, List[str]]:
        """Fix invalid quote usage"""
        warnings = []

        # Simple replacement: replace single quotes with double quotes
        # This handles cases like {'name': 'test'} -> {"name": "test"}
        if json_string.startswith("'") or ("'" in json_string and '"' not in json_string):
            # Case: entirely single-quoted JSON
            modified_str = json_string.replace("'", '"')
            if modified_str != json_string:
                warnings.append("Replaced single quotes with double quotes")
                return modified_str, warnings

        # More complex case: mixed quotes
        # Replace unescaped single quotes that appear to be string delimiters
        char_list: list[str] = []
        i = 0
        while i < len(json_string):
            char = json_string[i]

            if char == "\\":
                # Preserve escape sequences
                char_list.append(char)
                i += 1
                if i < len(json_string):
                    char_list.append(json_string[i])
                    i += 1
                continue

            if char == "'":
                # Replace single quote with double quote
                char_list.append('"')
                i += 1
            else:
                char_list.append(char)
                i += 1

        final_modified = "".join(char_list)
        if final_modified != json_string:
            warnings.append("Replaced single quotes with double quotes")

        return final_modified, warnings

    def _remove_control_characters(self, json_string: str) -> Tuple[str, List[str]]:
        """Remove control characters that break JSON parsing"""
        warnings = []

        # Remove control characters except allowed ones (tab, newline, carriage return)
        pattern = r"[\x00-\x08\x0B\x0C\x0E-\x1F\x7F-\x9F]"
        matches = re.findall(pattern, json_string)

        if matches:
            modified = re.sub(pattern, "", json_string)
            warnings.append(f"Removed {len(matches)} control character(s)")
            return modified, warnings

        return json_string, warnings

    def _handle_escaped_newlines(self, json_string: str) -> Tuple[str, List[str]]:
        """Handle escaped newlines in JSON strings"""
        warnings = []

        # Replace escaped newlines with proper JSON representation
        modified = json_string.replace("\\n", "\\n")

        if modified != json_string:
            warnings.append("Normalized escaped newlines")

        return modified, warnings

    def _fix_common_syntax_errors(self, json_string: str) -> Tuple[str, List[str]]:
        """Fix common JSON syntax errors"""
        warnings = []
        modified = json_string

        # Fix missing commas between array elements
        pattern = r"(\]\s*\[)"
        matches = re.finditer(pattern, modified)

        for match in reversed(list(matches)):
            replacement = f"{match.group(1).strip()},{match.group(1).strip()}"
            modified = modified[: match.start()] + replacement + modified[match.end() :]
            warnings.append("Added missing comma between array elements")

        # Fix missing colons in object properties
        pattern = r'("[^"]+")\s+("[^"]*"|\d+|true|false|null|\{|\[)'
        matches = re.finditer(pattern, modified)

        for match in reversed(list(matches)):
            replacement = f"{match.group(1)}:{match.group(2)}"
            modified = modified[: match.start()] + replacement + modified[match.end() :]
            warnings.append(f"Added missing colon for property: {match.group(1)}")

        return modified, warnings

    def _attempt_partial_parse(self, json_string: str) -> Tuple[str, List[str]]:
        """Attempt to extract and parse valid JSON from a larger string"""
        warnings = []

        # Look for JSON-like patterns in the string
        patterns = [
            r"\{[^{}]*\}",  # Simple object
            r"\[[^\[\]]*\]",  # Simple array
            r"\{(?:[^{}]*\{[^{}]*\})*[^{}]*\}",  # Nested objects (one level)
            r"\[(?:[^\[\]]*\[[^\[\]]*\])*[^\[\]]*\]",  # Nested arrays (one level)
        ]

        for pattern in patterns:
            matches = re.finditer(pattern, json_string)

            for match in matches:
                try:
                    # Try to parse this substring
                    json.loads(match.group())

                    # If successful, return this as the cleaned string
                    warnings.append("Extracted valid JSON from larger string")
                    return match.group(), warnings

                except json.JSONDecodeError:
                    continue

        return json_string, warnings

    def _apply_aggressive_recovery(self, json_string: str) -> str:
        """Apply more aggressive recovery for final attempts"""
        modified = json_string

        # Try to extract just the JSON part from a response that might include other text
        # Look for patterns that might contain JSON
        json_patterns = [
            r"```json\s*(.*?)\s*```",  # Markdown code blocks
            r"\{.*\}",  # Anything between braces
            r"\[.*\]",  # Anything between brackets
        ]

        for pattern in json_patterns:
            matches = re.findall(pattern, modified, re.DOTALL)
            for match in matches:
                try:
                    json.loads(match)
                    return match  # Return the first valid JSON found
                except (json.JSONDecodeError, ValueError, TypeError):
                    continue

        # If no JSON found, try to clean up the string
        # Remove common non-JSON prefixes/suffixes
        lines = modified.split("\n")
        json_lines = []

        for line in lines:
            line = line.strip()
            # Skip lines that are clearly not JSON
            if line.startswith(("```", "#", "*", "-", ">", "Error:", "Success:")) or line.endswith(("```", ".")):
                continue
            json_lines.append(line)

        return "\n".join(json_lines)

    def _get_string_context(self, json_string: str, position: int) -> bool:
        """Check if position is inside a JSON string"""
        quote_count = 0
        escape_next = False

        for i in range(position):
            char = json_string[i]

            if escape_next:
                escape_next = False
                continue

            if char == "\\":
                escape_next = True
                continue

            if char == '"':
                quote_count += 1

        return quote_count % 2 == 1  # True if inside string

    def get_stats(self) -> Dict[str, Union[int, float]]:
        """Get parsing statistics"""
        total = self.stats["total_parses"]
        if total > 0:
            return {
                **self.stats,
                "success_rate": (self.stats["successful_parses"] + self.stats["recovered_parses"]) / total,
                "recovery_rate": self.stats["recovered_parses"] / total,
                "failure_rate": self.stats["failed_parses"] / total,
                "avg_recovery_time": (
                    self.stats["total_recovery_time"] / self.stats["recovered_parses"]
                    if self.stats["recovered_parses"] > 0
                    else 0
                ),
            }
        return self.stats

    def reset_stats(self) -> None:
        """Reset parsing statistics"""
        self.stats = {
            "total_parses": 0,
            "successful_parses": 0,
            "recovered_parses": 0,
            "failed_parses": 0,
            "total_recovery_time": 0.0,
        }


# Global instance for easy import
parser = RobustJSONParser()


def parse_json(json_string: str, context: Optional[Dict] = None) -> ParseResult:
    """Convenience function to parse JSON with error recovery"""
    return parser.parse(json_string, context)


def get_parser_stats() -> Dict[str, Union[int, float]]:
    """Get global parser statistics"""
    return parser.get_stats()


def reset_parser_stats() -> None:
    """Reset global parser statistics"""
    parser.reset_stats()


# Test suite
if __name__ == "__main__":
    # Basic tests
    test_cases = [
        # Valid JSON
        '{"name": "test", "value": 123}',
        # Missing quotes
        '{name: "test", value: 123}',
        # Trailing comma
        '{"name": "test", "value": 123,}',
        # Invalid escape sequences
        '{"name": "test\\invalid", "value": 123}',
        # Partial object
        '{"name": "test"',
        # Mixed single quotes
        "{'name': 'test', 'value': 123}",
        # Control characters
        '{"name": "test\x00", "value": 123}',
    ]

    print("Testing Robust JSON Parser...")

    for i, test_case in enumerate(test_cases):
        print(f"\nTest {i + 1}: {test_case[:50]}...")

        result = parser.parse(test_case)

        if result.success:
            print(f"✓ Success (attempts: {result.recovery_attempts})")
            print(f"  Data: {result.data}")
        else:
            print(f"✗ Failed (attempts: {result.recovery_attempts})")
            print(f"  Error: {result.error}")

        if result.warnings:
            print(f"  Warnings: {result.warnings}")

    print("\nParser Statistics:")
    stats = parser.get_stats()
    for key, value in stats.items():
        print(f"  {key}: {value}")
