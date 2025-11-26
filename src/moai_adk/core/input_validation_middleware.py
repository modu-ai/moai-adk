"""
Enhanced Input Validation Middleware for MoAI-ADK

Production-ready input validation and normalization system that addresses tool input
validation failures identified in Claude Code debug logs. Provides intelligent parameter
mapping, version compatibility, and real-time input correction.

Author: MoAI-ADK Core Team
Version: 1.0.0
"""

import json
import logging
from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Callable, Dict, List, Optional, Set, Tuple

# Configure logging
logger = logging.getLogger(__name__)


class ValidationSeverity(Enum):
    """Input validation severity levels"""

    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class ToolCategory(Enum):
    """Categories of tools with specific validation requirements"""

    SEARCH = "search"
    FILE_OPERATIONS = "file_operations"
    TEXT_PROCESSING = "text_processing"
    DATA_ANALYSIS = "data_analysis"
    SYSTEM = "system"
    GENERAL = "general"


@dataclass
class ValidationError:
    """Individual validation error details"""

    code: str
    message: str
    path: List[str]
    severity: ValidationSeverity
    auto_corrected: bool = False
    original_value: Any = None
    corrected_value: Any = None
    suggestion: Optional[str] = None


@dataclass
class ValidationResult:
    """Result of input validation and normalization"""

    valid: bool
    normalized_input: Dict[str, Any]
    errors: List[ValidationError] = field(default_factory=list)
    warnings: List[str] = field(default_factory=list)
    transformations: List[str] = field(default_factory=list)
    processing_time_ms: float = 0.0


@dataclass
class ToolParameter:
    """Tool parameter definition for validation"""

    name: str
    param_type: str
    required: bool = False
    default_value: Any = None
    aliases: List[str] = field(default_factory=list)
    validation_function: Optional[Callable] = None
    description: str = ""
    deprecated_aliases: List[str] = field(default_factory=list)


class EnhancedInputValidationMiddleware:
    """
    Production-ready input validation middleware that addresses tool input validation
    failures from Claude Code debug logs with intelligent parameter mapping and normalization.

    Key Features:
    - Smart parameter mapping for unrecognized keys
    - Version compatibility support
    - Real-time input normalization
    - Comprehensive error tracking and correction
    - Tool-specific validation rules
    """

    def __init__(self, enable_logging: bool = True, enable_caching: bool = True):
        self.enable_logging = enable_logging
        self.enable_caching = enable_caching

        # Tool parameter definitions
        self.tool_parameters = self._load_tool_parameter_definitions()

        # Parameter mapping for compatibility
        self.parameter_mappings = self._load_parameter_mappings()

        # Validation cache
        self.validation_cache: Optional[Dict[str, Any]] = {} if enable_caching else None

        # Statistics
        self.stats = {
            "validations_performed": 0,
            "auto_corrections": 0,
            "errors_resolved": 0,
            "transformations_applied": 0,
        }

    def _load_tool_parameter_definitions(self) -> Dict[str, List[ToolParameter]]:
        """Load tool parameter definitions with compatibility information"""
        return {
            "Grep": [
                ToolParameter(
                    name="pattern",
                    param_type="string",
                    required=True,
                    description="Regex pattern to search for",
                    aliases=["regex", "search_pattern"],
                ),
                ToolParameter(
                    name="output_mode",
                    param_type="string",
                    default_value="content",
                    description="Output format mode",
                    aliases=["mode", "format"],
                    validation_function=self._validate_grep_mode,
                ),
                ToolParameter(
                    name="head_limit",
                    param_type="integer",
                    description="Limit number of results",
                    default_value=None,
                    aliases=["limit", "max_results", "count", "head"],
                    deprecated_aliases=["max"],
                ),
                ToolParameter(
                    name="path",
                    param_type="string",
                    description="Path to search in",
                    aliases=["directory", "folder", "search_path", "root"],
                ),
                ToolParameter(
                    name="file_pattern",
                    param_type="string",
                    description="File pattern to match",
                    aliases=["glob", "pattern", "files", "file_glob"],
                ),
                ToolParameter(
                    name="case_sensitive",
                    param_type="boolean",
                    default_value=False,
                    description="Case sensitive search",
                    aliases=["case", "ignore_case", "sensitive"],
                ),
                ToolParameter(
                    name="context_lines",
                    param_type="integer",
                    default_value=0,
                    description="Number of context lines",
                    aliases=["context", "before_context", "after_context", "C"],
                ),
            ],
            "Glob": [
                ToolParameter(
                    name="pattern",
                    param_type="string",
                    required=True,
                    description="Glob pattern to match",
                    aliases=["glob", "file_pattern", "search_pattern"],
                ),
                ToolParameter(
                    name="path",
                    param_type="string",
                    description="Base path for glob",
                    aliases=["directory", "folder", "root", "base_path"],
                ),
                ToolParameter(
                    name="recursive",
                    param_type="boolean",
                    default_value=True,
                    description="Recursive directory search",
                    aliases=["recurse", "recursive_search"],
                ),
            ],
            "Read": [
                ToolParameter(
                    name="file_path",
                    param_type="string",
                    required=True,
                    description="Path to file to read",
                    aliases=["path", "filename", "file", "source"],
                ),
                ToolParameter(
                    name="offset",
                    param_type="integer",
                    default_value=0,
                    description="Starting line number",
                    aliases=["start", "start_line", "begin", "from"],
                ),
                ToolParameter(
                    name="limit",
                    param_type="integer",
                    description="Number of lines to read",
                    aliases=["count", "max_lines", "lines", "size"],
                ),
            ],
            "Bash": [
                ToolParameter(
                    name="command",
                    param_type="string",
                    required=True,
                    description="Command to execute",
                    aliases=["cmd", "execute", "run", "script"],
                ),
                ToolParameter(
                    name="timeout",
                    param_type="integer",
                    default_value=10000,
                    description="Timeout in milliseconds",
                    aliases=["timeout_ms", "max_time", "time_limit"],
                ),
                ToolParameter(
                    name="working_directory",
                    param_type="string",
                    description="Working directory for command",
                    aliases=["cwd", "work_dir", "directory", "folder"],
                ),
                ToolParameter(
                    name="environment",
                    param_type="dict",
                    description="Environment variables",
                    aliases=["env", "env_vars", "variables"],
                ),
            ],
            "Task": [
                ToolParameter(
                    name="subagent_type",
                    param_type="string",
                    required=True,
                    description="Type of subagent to use",
                    aliases=["agent_type", "type", "agent", "model"],
                ),
                ToolParameter(
                    name="prompt",
                    param_type="string",
                    required=True,
                    description="Prompt for the subagent",
                    aliases=["message", "input", "query", "instruction"],
                ),
                ToolParameter(
                    name="context",
                    param_type="dict",
                    description="Additional context",
                    aliases=["data", "variables", "params"],
                ),
                ToolParameter(
                    name="debug",
                    param_type="boolean",
                    default_value=False,
                    description="Enable debug mode",
                    aliases=["verbose", "debug_mode"],
                ),
            ],
            "Write": [
                ToolParameter(
                    name="file_path",
                    param_type="string",
                    required=True,
                    description="Path to file to write",
                    aliases=["path", "filename", "file", "destination"],
                ),
                ToolParameter(
                    name="content",
                    param_type="string",
                    required=True,
                    description="Content to write",
                    aliases=["data", "text", "body", "contents"],
                ),
                ToolParameter(
                    name="create_directories",
                    param_type="boolean",
                    default_value=False,
                    description="Create parent directories if needed",
                    aliases=["mkdir", "create_dirs", "make_dirs"],
                ),
                ToolParameter(
                    name="backup",
                    param_type="boolean",
                    default_value=False,
                    description="Create backup of existing file",
                    aliases=["backup_existing", "make_backup"],
                ),
            ],
            "Edit": [
                ToolParameter(
                    name="file_path",
                    param_type="string",
                    required=True,
                    description="Path to file to edit",
                    aliases=["path", "filename", "file"],
                ),
                ToolParameter(
                    name="old_string",
                    param_type="string",
                    required=True,
                    description="String to replace",
                    aliases=["search", "find", "from", "original"],
                ),
                ToolParameter(
                    name="new_string",
                    param_type="string",
                    required=True,
                    description="Replacement string",
                    aliases=["replace", "to", "replacement"],
                ),
                ToolParameter(
                    name="replace_all",
                    param_type="boolean",
                    default_value=False,
                    description="Replace all occurrences",
                    aliases=["global", "all", "replace_all_occurrences"],
                ),
            ],
        }

    def _load_parameter_mappings(self) -> Dict[str, str]:
        """Load parameter mapping for compatibility with different versions"""
        return {
            # Grep tool mappings
            "grep_head_limit": "head_limit",
            "grep_limit": "head_limit",
            "grep_max": "head_limit",
            "grep_count": "head_limit",
            "grep_head": "head_limit",
            "max_results": "head_limit",
            "result_limit": "head_limit",
            "num_results": "head_limit",
            # Output mode mappings
            "grep_mode": "output_mode",
            "grep_format": "output_mode",
            "output_format": "output_mode",
            "display_mode": "output_mode",
            "show_mode": "output_mode",
            # Path mappings
            "search_path": "path",
            "base_path": "path",
            "root_dir": "path",
            "target_dir": "path",
            # Glob tool mappings
            "glob_pattern": "pattern",
            "file_glob": "pattern",
            "search_pattern": "pattern",
            "match_pattern": "pattern",
            # Read tool mappings
            "start_line": "offset",
            "begin_line": "offset",
            "from_line": "offset",
            "max_lines": "limit",
            "line_count": "limit",
            # Bash tool mappings
            "cmd": "command",
            "execute": "command",
            "run_command": "command",
            "timeout_ms": "timeout",
            "max_time": "timeout",
            "time_limit": "timeout",
            "work_dir": "working_directory",
            "cwd": "working_directory",
            # Task tool mappings
            "agent_type": "subagent_type",
            "message": "prompt",
            "instruction": "prompt",
            "query": "prompt",
            # Write tool mappings
            "filename": "file_path",
            "destination": "file_path",
            "data": "content",
            "text": "content",
            "body": "content",
            "make_backup": "backup",
            # Edit tool mappings
            "search": "old_string",
            "find": "old_string",
            "from": "old_string",
            "replace": "new_string",
            "to": "new_string",
            "global": "replace_all",
            "all": "replace_all",
        }

    def _validate_grep_mode(self, mode: str) -> bool:
        """Validate grep output mode"""
        valid_modes = ["content", "files_with_matches", "count"]
        return mode in valid_modes

    def validate_and_normalize_input(self, tool_name: str, input_data: Dict[str, Any]) -> ValidationResult:
        """
        Validate and normalize tool input data.

        This is the main method that addresses the tool input validation failures
        from the debug logs (Lines 476-495).
        """
        import time

        start_time = time.time()

        self.stats["validations_performed"] += 1

        result = ValidationResult(valid=True, normalized_input=input_data.copy())

        try:
            # Get tool parameters
            tool_params = self.tool_parameters.get(tool_name, [])

            if not tool_params:
                # Unknown tool - perform basic validation only
                result.warnings.append(f"Unknown tool: {tool_name}")
                return result

            # Step 1: Map unrecognized parameters
            mapped_input, mapping_errors = self._map_parameters(tool_name, result.normalized_input)
            result.normalized_input = mapped_input
            result.errors.extend(mapping_errors)

            # Step 2: Validate required parameters
            required_errors = self._validate_required_parameters(tool_params, result.normalized_input)
            result.errors.extend(required_errors)

            # Step 3: Apply default values
            self._apply_default_values(tool_params, result.normalized_input)

            # Step 4: Validate parameter values and apply type conversions
            value_errors = self._validate_parameter_values(tool_params, result.normalized_input)
            result.errors.extend(value_errors)

            # Apply type conversions from errors
            for error in value_errors:
                if error.code == "type_conversion" and error.auto_corrected and error.corrected_value is not None:
                    result.normalized_input[error.path[0]] = error.corrected_value

            # Step 5: Normalize parameter formats
            transformations = self._normalize_parameter_formats(tool_params, result.normalized_input)
            result.transformations.extend(transformations)

            # Step 6: Check for deprecated parameters
            deprecated_warnings = self._check_deprecated_parameters(tool_params, result.normalized_input)
            result.warnings.extend(deprecated_warnings)

            # Update valid status
            critical_errors = [e for e in result.errors if e.severity == ValidationSeverity.CRITICAL]
            if critical_errors:
                result.valid = False

            # Update statistics
            auto_corrected = len([e for e in result.errors if e.auto_corrected])
            self.stats["auto_corrections"] += auto_corrected
            if auto_corrected > 0:
                self.stats["errors_resolved"] += auto_corrected

            self.stats["transformations_applied"] += len(result.transformations)

            if self.enable_logging and (result.errors or result.warnings or result.transformations):
                logger.info(
                    f"Input validation for {tool_name}: "
                    f"valid={result.valid}, errors={len(result.errors)}, "
                    f"warnings={len(result.warnings)}, transformations={len(result.transformations)}"
                )

        except Exception as e:
            result.valid = False
            result.errors.append(
                ValidationError(
                    code="validation_exception",
                    message=f"Validation error: {str(e)}",
                    path=[],
                    severity=ValidationSeverity.CRITICAL,
                )
            )

            if self.enable_logging:
                logger.error(f"Exception during input validation for {tool_name}: {e}")

        result.processing_time_ms = (time.time() - start_time) * 1000

        return result

    def _map_parameters(
        self,
        tool_name: str,
        input_data: Dict[str, Any],
    ) -> Tuple[Dict[str, Any], List[ValidationError]]:
        """Map and rename unrecognized parameters to their canonical forms"""
        mapped_input = input_data.copy()
        errors = []

        # Get tool-specific parameters
        tool_params = self.tool_parameters.get(tool_name, [])

        # Create mapping of all valid parameter names and aliases
        valid_names = set()
        alias_mapping = {}

        for param in tool_params:
            valid_names.add(param.name)
            for alias in param.aliases + param.deprecated_aliases:
                alias_mapping[alias] = param.name

        # Add global parameter mappings
        global_mapping_prefix = f"{tool_name.lower()}_"
        for global_key, canonical_key in self.parameter_mappings.items():
            if global_key.startswith(global_mapping_prefix):
                short_key = global_key[len(global_mapping_prefix) :]
                alias_mapping[short_key] = canonical_key

        # Check each input parameter
        for param_name in list(mapped_input.keys()):
            if param_name in valid_names:
                # Parameter is already in canonical form
                continue

            if param_name in alias_mapping:
                # Map to canonical name
                canonical_name = alias_mapping[param_name]
                original_value = mapped_input[param_name]

                mapped_input[canonical_name] = original_value
                del mapped_input[param_name]

                errors.append(
                    ValidationError(
                        code="parameter_mapped",
                        message=f"Mapped parameter '{param_name}' to '{canonical_name}'",
                        path=[param_name],
                        severity=ValidationSeverity.LOW,
                        auto_corrected=True,
                        original_value=original_value,
                        corrected_value=original_value,
                        suggestion=f"Use '{canonical_name}' instead of '{param_name}'",
                    )
                )

            else:
                # Unknown parameter - create error
                original_value = mapped_input[param_name]

                # Suggest closest match
                suggestion = self._find_closest_parameter_match(param_name, valid_names)

                errors.append(
                    ValidationError(
                        code="unrecognized_parameter",
                        message=f"Unrecognized parameter: '{param_name}'",
                        path=[param_name],
                        severity=ValidationSeverity.HIGH,
                        original_value=original_value,
                        suggestion=suggestion,
                    )
                )

        return mapped_input, errors

    def _find_closest_parameter_match(self, param_name: str, valid_names: Set[str]) -> Optional[str]:
        """Find the closest matching valid parameter name"""
        # Convert to lowercase for comparison
        param_lower = param_name.lower()
        valid_lower = {name.lower(): name for name in valid_names}

        # Exact match (case-insensitive)
        if param_lower in valid_lower:
            return valid_lower[param_lower]

        # Find best match using Levenshtein distance
        best_match = None
        best_score = float("inf")

        for valid_lower_name, canonical_name in valid_lower.items():
            # Simple similarity score
            score = self._calculate_string_similarity(param_lower, valid_lower_name)

            if score < best_score and score < 0.7:  # Similarity threshold
                best_score = score
                best_match = canonical_name

        return best_match

    def _calculate_string_similarity(self, s1: str, s2: str) -> float:
        """Calculate similarity between two strings (0-1, lower is more similar)"""
        # Simple Levenshtein distance approximation
        if not s1:
            return len(s2)
        if not s2:
            return len(s1)

        if len(s1) < len(s2):
            s1, s2 = s2, s1

        # Calculate distance
        previous_row = list(range(len(s2) + 1))
        for i, c1 in enumerate(s1):
            current_row = [i + 1]
            for j, c2 in enumerate(s2):
                insertions = previous_row[j + 1] + 1
                deletions = current_row[j] + 1
                substitutions = previous_row[j] + (c1 != c2)
                current_row.append(min(insertions, deletions, substitutions))
            previous_row = current_row

        distance = previous_row[-1]
        max_len = max(len(s1), len(s2))
        return distance / max_len if max_len > 0 else 0

    def _validate_required_parameters(
        self,
        tool_params: List[ToolParameter],
        input_data: Dict[str, Any],
    ) -> List[ValidationError]:
        """Validate that all required parameters are present"""
        errors = []

        for param in tool_params:
            if param.required and param.name not in input_data:
                errors.append(
                    ValidationError(
                        code="missing_required_parameter",
                        message=f"Missing required parameter: '{param.name}'",
                        path=[],
                        severity=ValidationSeverity.CRITICAL,
                        suggestion=f"Add '{param.name}' parameter",
                    )
                )

        return errors

    def _apply_default_values(self, tool_params: List[ToolParameter], input_data: Dict[str, Any]) -> None:
        """Apply default values for missing optional parameters"""
        for param in tool_params:
            if not param.required and param.name not in input_data and param.default_value is not None:
                input_data[param.name] = param.default_value

    def _validate_parameter_values(
        self,
        tool_params: List[ToolParameter],
        input_data: Dict[str, Any],
    ) -> List[ValidationError]:
        """Validate parameter values against their types and constraints"""
        errors = []

        for param in tool_params:
            if param.name not in input_data:
                continue

            value = input_data[param.name]

            # Type validation
            type_errors = self._validate_parameter_type(param, value, input_data)
            errors.extend(type_errors)

            # Custom validation function
            if param.validation_function and not type_errors:
                try:
                    if not param.validation_function(value):
                        errors.append(
                            ValidationError(
                                code="validation_function_failed",
                                message=f"Parameter '{param.name}' failed custom validation",
                                path=[param.name],
                                severity=ValidationSeverity.HIGH,
                                original_value=value,
                            )
                        )
                except Exception as e:
                    errors.append(
                        ValidationError(
                            code="validation_function_error",
                            message=f"Error validating parameter '{param.name}': {str(e)}",
                            path=[param.name],
                            severity=ValidationSeverity.HIGH,
                            original_value=value,
                        )
                    )

        return errors

    def _validate_parameter_type(
        self,
        param: ToolParameter,
        value: Any,
        input_data: Dict[str, Any],
    ) -> List[ValidationError]:
        """Validate parameter value against expected type"""
        errors: List[ValidationError] = []

        # Type mapping
        type_validators = {
            "string": lambda v: isinstance(v, str),
            "integer": lambda v: isinstance(v, int) or (isinstance(v, str) and v.isdigit()),
            "boolean": lambda v: isinstance(v, bool) or str(v).lower() in ["true", "false", "1", "0"],
            "dict": lambda v: isinstance(v, dict),
            "list": lambda v: isinstance(v, list),
            "float": lambda v: isinstance(v, float) or (isinstance(v, (int, str)) and self._is_float(v)),
        }

        validator = type_validators.get(param.param_type)
        if not validator:
            return errors

        if not validator(value):
            # Try to auto-convert
            converted_value = self._convert_parameter_type(value, param.param_type)
            if converted_value is not None:
                # Apply the converted value to input_data
                input_data[param.name] = converted_value
                errors.append(
                    ValidationError(
                        code="type_conversion",
                        message=f"Converted parameter '{param.name}' from {type(value).__name__} to {param.param_type}",
                        path=[param.name],
                        severity=ValidationSeverity.LOW,
                        auto_corrected=True,
                        original_value=value,
                        corrected_value=converted_value,
                    )
                )
            else:
                errors.append(
                    ValidationError(
                        code="type_mismatch",
                        message=f"Parameter '{param.name}' expects {param.param_type}, got {type(value).__name__}",
                        path=[param.name],
                        severity=ValidationSeverity.HIGH,
                        original_value=value,
                        suggestion=f"Provide a {param.param_type} value for '{param.name}'",
                    )
                )

        return errors

    def _convert_parameter_type(self, value: Any, target_type: str) -> Any:
        """Attempt to convert value to target type"""
        try:
            if target_type == "string":
                return str(value)
            elif target_type == "integer":
                if isinstance(value, str):
                    # Handle negative numbers and leading/trailing whitespace
                    value = value.strip()
                    try:
                        return int(value)
                    except ValueError:
                        # Try to convert from float string
                        try:
                            return int(float(value))
                        except ValueError:
                            return None
                elif isinstance(value, float):
                    return int(value)
                elif isinstance(value, bool):
                    return int(value)
            elif target_type == "boolean":
                if isinstance(value, str):
                    return value.lower() in ["true", "1", "yes", "on"]
                elif isinstance(value, (int, float)):
                    return bool(value)
            elif target_type == "float":
                if isinstance(value, str):
                    return float(value)
                elif isinstance(value, int):
                    return float(value)
        except (ValueError, TypeError):
            pass

        return None

    def _is_float(self, value) -> bool:
        """Check if value can be converted to float"""
        try:
            float(value)
            return True
        except (ValueError, TypeError):
            return False

    def _normalize_parameter_formats(self, tool_params: List[ToolParameter], input_data: Dict[str, Any]) -> List[str]:
        """Normalize parameter formats for consistency"""
        transformations = []

        for param in tool_params:
            if param.name not in input_data:
                continue

            value = input_data[param.name]

            # Normalize boolean values
            if param.param_type == "boolean":
                if isinstance(value, str):
                    normalized_bool = value.lower() in ["true", "1", "yes", "on"]
                    if value != str(normalized_bool):
                        input_data[param.name] = normalized_bool
                        transformations.append(
                            f"Normalized '{param.name}' boolean from '{value}' to '{normalized_bool}'"
                        )

            # Normalize file paths
            elif param.name in ["file_path", "path", "directory"] and isinstance(value, str):
                # Convert to forward slashes and remove trailing slash
                normalized_path = value.replace("\\", "/").rstrip("/")
                if value != normalized_path:
                    input_data[param.name] = normalized_path
                    transformations.append(f"Normalized '{param.name}' path from '{value}' to '{normalized_path}'")

            # Normalize numeric formats - Always attempt conversion for numeric types
            elif param.param_type in ["integer", "float"] and isinstance(value, str):
                try:
                    if param.param_type == "integer":
                        normalized_num: int | float = int(float(value.strip()))  # Handle "123.0" -> 123
                    else:  # float
                        normalized_num = float(value.strip())

                    input_data[param.name] = normalized_num
                    transformations.append(
                        f"Normalized '{param.name}' {param.param_type} from '{value}' to '{normalized_num}'"
                    )
                except ValueError:
                    # Keep original value if conversion fails
                    pass

        return transformations

    def _check_deprecated_parameters(self, tool_params: List[ToolParameter], input_data: Dict[str, Any]) -> List[str]:
        """Check for deprecated parameter usage"""
        warnings = []

        for param in tool_params:
            if param.name not in input_data:
                continue

            # Check if parameter name is deprecated
            for deprecated_alias in param.deprecated_aliases:
                if deprecated_alias in input_data and param.name != deprecated_alias:
                    value = input_data[deprecated_alias]
                    input_data[param.name] = value
                    del input_data[deprecated_alias]

                    warnings.append(
                        f"Deprecated parameter '{deprecated_alias}' replaced with '{param.name}'. "
                        f"This alias will be removed in future versions."
                    )

        return warnings

    def get_validation_stats(self) -> Dict[str, Any]:
        """Get input validation statistics"""
        return {
            **self.stats,
            "tools_configured": len(self.tool_parameters),
            "parameter_mappings": len(self.parameter_mappings),
            "cache_size": len(self.validation_cache) if self.validation_cache else 0,
        }

    def register_tool_parameters(self, tool_name: str, parameters: List[ToolParameter]) -> None:
        """Register custom tool parameters"""
        self.tool_parameters[tool_name] = parameters

    def add_parameter_mapping(self, from_key: str, to_key: str) -> None:
        """Add custom parameter mapping"""
        self.parameter_mappings[from_key] = to_key

    def export_validation_report(self, output_path: str) -> None:
        """Export validation report to file"""
        report = {
            "generated_at": __import__("time").time(),
            "stats": self.get_validation_stats(),
            "configured_tools": list(self.tool_parameters.keys()),
            "parameter_mappings": self.parameter_mappings,
        }

        with open(output_path, "w", encoding="utf-8") as f:
            json.dump(report, f, indent=2, ensure_ascii=False)


# Global instance for easy import
validation_middleware = EnhancedInputValidationMiddleware()


def validate_tool_input(tool_name: str, input_data: Dict[str, Any]) -> ValidationResult:
    """Convenience function for tool input validation"""
    return validation_middleware.validate_and_normalize_input(tool_name, input_data)


def get_validation_stats() -> Dict[str, Any]:
    """Convenience function to get validation statistics"""
    return validation_middleware.get_validation_stats()


if __name__ == "__main__":
    # Demo script for testing the input validation middleware
    print("🔧 MoAI-ADK Enhanced Input Validation Middleware Demo")
    print("=" * 60)

    # Test cases that reproduce the debug log errors
    test_cases = [
        {
            "name": "Grep with head_limit (debug log error)",
            "tool": "Grep",
            "input": {
                "pattern": "test",
                "head_limit": 10,  # This was causing the error
                "output_mode": "content",
            },
        },
        {
            "name": "Grep with alternative parameter names",
            "tool": "Grep",
            "input": {
                "pattern": "test",
                "max_results": 20,  # Should be mapped to head_limit
                "search_path": "/src",  # Should be mapped to path
            },
        },
        {
            "name": "Grep with deprecated parameters",
            "tool": "Grep",
            "input": {
                "pattern": "test",
                "count": 15,  # Deprecated alias
                "folder": "/src",  # Should be mapped to path
            },
        },
        {
            "name": "Read with parameter aliases",
            "tool": "Read",
            "input": {
                "filename": "/path/to/file.txt",  # Should be mapped to file_path
                "start_line": 10,  # Should be mapped to offset
                "lines": 50,  # Should be mapped to limit
            },
        },
        {
            "name": "Task with mixed parameter names",
            "tool": "Task",
            "input": {
                "agent_type": "debug-helper",  # Should be mapped to subagent_type
                "message": "test message",  # Should be mapped to prompt
                "verbose": True,  # Should be mapped to debug
            },
        },
        {
            "name": "Bash with alternative parameters",
            "tool": "Bash",
            "input": {
                "cmd": "ls -la",  # Should be mapped to command
                "cwd": "/home/user",  # Should be mapped to working_directory
                "timeout_ms": 5000,  # Should be mapped to timeout
            },
        },
    ]

    for i, test_case in enumerate(test_cases, 1):
        print(f"\n{i}. {test_case['name']}")
        tool_name: str = test_case["tool"]  # type: ignore[assignment]
        tool_input: Dict[str, Any] = test_case["input"]  # type: ignore[assignment]
        print(f"   Tool: {tool_name}")
        print(f"   Original input: {tool_input}")

        result = validate_tool_input(tool_name, tool_input)

        print(f"   Valid: {result.valid}")
        print(f"   Errors: {len(result.errors)}")
        print(f"   Warnings: {len(result.warnings)}")
        print(f"   Transformations: {len(result.transformations)}")
        print(f"   Processing time: {result.processing_time_ms:.2f}ms")

        if result.errors:
            print("   Error details:")
            for error in result.errors:
                status = "✅ AUTO-CORRECTED" if error.auto_corrected else "❌ NOT FIXED"
                print(f"     • {error.message} [{status}]")
                if error.suggestion:
                    print(f"       Suggestion: {error.suggestion}")

        if result.warnings:
            print("   Warnings:")
            for warning in result.warnings:
                print(f"     • {warning}")

        if result.transformations:
            print("   Transformations:")
            for transform in result.transformations:
                print(f"     • {transform}")

        print(f"   Normalized input: {result.normalized_input}")

    print("\n📊 Validation Statistics:")
    stats = get_validation_stats()
    for key, value in stats.items():
        print(f"   {key}: {value}")

    print("\n✨ Demo completed! The Enhanced Input Validation Middleware addresses")
    print("   the tool input validation failures from the debug logs with automatic")
    print("   parameter mapping and intelligent correction.")
