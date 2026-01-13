#!/usr/bin/env python3
"""Configuration Validator for Hooks System

Validates and normalizes configuration settings for timeout management,
resource limits, and hook behavior patterns.

Features:
- Schema validation for timeout configurations
- Default value normalization
- Configuration migration support
- Validation error reporting
- Performance impact assessment
"""

import json
import logging
from dataclasses import dataclass
from enum import Enum
from pathlib import Path
from typing import Any, Dict, List, Tuple


class ValidationLevel(Enum):
    """Severity levels for validation issues"""

    INFO = "info"
    WARNING = "warning"
    ERROR = "error"
    CRITICAL = "critical"


@dataclass
class ValidationIssue:
    """Configuration validation issue with context"""

    level: ValidationLevel
    path: str
    message: str
    suggestion: str | None = None
    current_value: str | None = None


@dataclass
class TimeoutConfigSchema:
    """Schema for timeout configuration validation"""

    min_timeout_ms: int = 500
    max_timeout_ms: int = 30000
    default_timeout_ms: int = 5000
    recommended_retry_count: int = 2
    max_retry_count: int = 5
    default_retry_delay_ms: int = 200
    max_retry_delay_ms: int = 5000


@dataclass
class ResourceConfigSchema:
    """Schema for resource configuration validation"""

    default_memory_limit_mb: int = 100
    max_memory_limit_mb: int = 1000
    default_max_workers: int = 4
    max_max_workers: int = 16
    cache_size_default: int = 100
    cache_size_max: int = 1000


class ConfigurationValidator:
    """Validates and normalizes hooks configuration

    Features:
    - Comprehensive schema validation
    - Automatic value normalization
    - Performance impact analysis
    - Migration support for older configs
    - Detailed validation reporting
    """

    def __init__(self):
        self._logger = logging.getLogger(__name__)
        self._timeout_schema = TimeoutConfigSchema()
        self._resource_schema = ResourceConfigSchema()

        # Hook-specific configuration templates
        self._hook_templates = {
            "session_start": {
                "policy": "normal",
                "timeout_ms": 5000,
                "retry_count": 1,
                "retry_delay_ms": 200,
                "graceful_degradation": True,
                "memory_limit_mb": 100,
            },
            "session_end": {
                "policy": "normal",
                "timeout_ms": 5000,
                "retry_count": 1,
                "retry_delay_ms": 500,
                "graceful_degradation": True,
                "memory_limit_mb": 150,
            },
            "pre_tool": {
                "policy": "fast",
                "timeout_ms": 2000,
                "retry_count": 1,
                "retry_delay_ms": 100,
                "graceful_degradation": True,
                "memory_limit_mb": 50,
            },
        }

    def validate_config(self, config: dict[str, Any]) -> Tuple[bool, List[ValidationIssue]]:
        """Validate complete configuration and return issues"""
        issues = []

        # Validate top-level structure
        issues.extend(self._validate_structure(config))

        # Validate timeout manager configuration
        if "timeout_manager" in config:
            issues.extend(self._validate_timeout_manager(config["timeout_manager"]))

        # Validate hook-specific configurations
        if "hook_configs" in config:
            issues.extend(self._validate_hook_configs(config["hook_configs"]))

        # Validate resource configurations
        if "resources" in config:
            issues.extend(self._validate_resources(config["resources"]))

        # Validate performance settings
        if "performance" in config:
            issues.extend(self._validate_performance(config["performance"]))

        is_valid = not any(issue.level in [ValidationLevel.ERROR, ValidationLevel.CRITICAL] for issue in issues)
        return is_valid, issues

    def _validate_structure(self, config: dict[str, Any]) -> List[ValidationIssue]:
        """Validate top-level configuration structure"""
        issues = []

        # Check for required sections (optional for backward compatibility)
        expected_sections = [
            "timeout_manager",
            "hook_configs",
            "resources",
            "performance",
        ]
        for section in expected_sections:
            if section not in config:
                issues.append(
                    ValidationIssue(
                        level=ValidationLevel.INFO,
                        path=f".{section}",
                        message=f"Missing optional '{section}' configuration section",
                        suggestion=f"Add '{section}' section to enable advanced features",
                    )
                )

        return issues

    def _validate_timeout_manager(self, timeout_config: dict[str, Any]) -> List[ValidationIssue]:
        """Validate timeout manager configuration"""
        issues = []
        base_path = ".timeout_manager"

        # Validate global timeout settings
        if "global_timeout_ms" in timeout_config:
            timeout_ms = timeout_config["global_timeout_ms"]
            if not isinstance(timeout_ms, int) or timeout_ms < 0:
                issues.append(
                    ValidationIssue(
                        level=ValidationLevel.ERROR,
                        path=f"{base_path}.global_timeout_ms",
                        message="Global timeout must be a non-negative integer",
                        current_value=str(timeout_ms),
                        suggestion="Use a positive integer like 5000 (5 seconds)",
                    )
                )
            elif timeout_ms < self._timeout_schema.min_timeout_ms:
                issues.append(
                    ValidationIssue(
                        level=ValidationLevel.WARNING,
                        path=f"{base_path}.global_timeout_ms",
                        message=f"Global timeout is very low ({timeout_ms}ms)",
                        current_value=str(timeout_ms),
                        suggestion=f"Consider using at least {self._timeout_schema.min_timeout_ms}ms",
                    )
                )
            elif timeout_ms > self._timeout_schema.max_timeout_ms:
                issues.append(
                    ValidationIssue(
                        level=ValidationLevel.WARNING,
                        path=f"{base_path}.global_timeout_ms",
                        message=f"Global timeout is very high ({timeout_ms}ms)",
                        current_value=str(timeout_ms),
                        suggestion=f"Consider using at most {self._timeout_schema.max_timeout_ms}ms",
                    )
                )

        # Validate default retry count
        if "default_retry_count" in timeout_config:
            retry_count = timeout_config["default_retry_count"]
            if not isinstance(retry_count, int) or retry_count < 0:
                issues.append(
                    ValidationIssue(
                        level=ValidationLevel.ERROR,
                        path=f"{base_path}.default_retry_count",
                        message="Default retry count must be a non-negative integer",
                        current_value=str(retry_count),
                        suggestion="Use an integer between 0 and 3",
                    )
                )
            elif retry_count > self._timeout_schema.max_retry_count:
                issues.append(
                    ValidationIssue(
                        level=ValidationLevel.WARNING,
                        path=f"{base_path}.default_retry_count",
                        message=f"High retry count may cause performance issues ({retry_count})",
                        current_value=str(retry_count),
                        suggestion=f"Consider using at most {self._timeout_schema.max_retry_count} retries",
                    )
                )

        # Validate graceful degradation setting
        if "graceful_degradation" in timeout_config:
            graceful = timeout_config["graceful_degradation"]
            if not isinstance(graceful, bool):
                issues.append(
                    ValidationIssue(
                        level=ValidationLevel.ERROR,
                        path=f"{base_path}.graceful_degradation",
                        message="Graceful degradation must be a boolean",
                        current_value=str(graceful),
                        suggestion="Use true or false",
                    )
                )

        return issues

    def _validate_hook_configs(self, hook_configs: dict[str, Any]) -> List[ValidationIssue]:
        """Validate hook-specific configurations"""
        issues = []
        base_path = ".hook_configs"

        for hook_name, config in hook_configs.items():
            hook_path = f"{base_path}.{hook_name}"
            issues.extend(self._validate_single_hook_config(hook_name, config, hook_path))

        return issues

    def _validate_single_hook_config(self, hook_name: str, config: dict[str, Any], path: str) -> List[ValidationIssue]:
        """Validate configuration for a single hook"""
        issues = []

        # Get template for this hook type
        self._get_hook_template(hook_name)

        # Validate policy
        if "policy" in config:
            policy = config["policy"]
            valid_policies = ["fast", "normal", "slow", "custom"]
            if policy not in valid_policies:
                issues.append(
                    ValidationIssue(
                        level=ValidationLevel.ERROR,
                        path=f"{path}.policy",
                        message=f"Invalid policy '{policy}'",
                        current_value=str(policy),
                        suggestion=f"Use one of: {', '.join(valid_policies)}",
                    )
                )

        # Validate timeout_ms
        if "timeout_ms" in config:
            timeout_ms = config["timeout_ms"]
            if not isinstance(timeout_ms, int) or timeout_ms <= 0:
                issues.append(
                    ValidationIssue(
                        level=ValidationLevel.ERROR,
                        path=f"{path}.timeout_ms",
                        message="Timeout must be a positive integer",
                        current_value=str(timeout_ms),
                        suggestion="Use a positive integer in milliseconds",
                    )
                )
            else:
                # Policy-based timeout validation
                if "policy" in config:
                    policy = config["policy"]
                    recommended = self._get_recommended_timeout(policy)
                    if timeout_ms < recommended * 0.5:
                        issues.append(
                            ValidationIssue(
                                level=ValidationLevel.WARNING,
                                path=f"{path}.timeout_ms",
                                message=f"Timeout is very low for '{policy}' policy",
                                current_value=str(timeout_ms),
                                suggestion=f"Consider using at least {recommended}ms",
                            )
                        )
                    elif timeout_ms > recommended * 3:
                        issues.append(
                            ValidationIssue(
                                level=ValidationLevel.WARNING,
                                path=f"{path}.timeout_ms",
                                message=f"Timeout is very high for '{policy}' policy",
                                current_value=str(timeout_ms),
                                suggestion=f"Consider using at most {recommended * 2}ms",
                            )
                        )

        # Validate retry_count
        if "retry_count" in config:
            retry_count = config["retry_count"]
            if not isinstance(retry_count, int) or retry_count < 0:
                issues.append(
                    ValidationIssue(
                        level=ValidationLevel.ERROR,
                        path=f"{path}.retry_count",
                        message="Retry count must be a non-negative integer",
                        current_value=str(retry_count),
                        suggestion="Use an integer between 0 and 3",
                    )
                )
            elif retry_count > self._timeout_schema.max_retry_count:
                issues.append(
                    ValidationIssue(
                        level=ValidationLevel.WARNING,
                        path=f"{path}.retry_count",
                        message=f"High retry count may cause delays ({retry_count})",
                        current_value=str(retry_count),
                        suggestion=f"Consider using at most {self._timeout_schema.max_retry_count} retries",
                    )
                )

        # Validate retry_delay_ms
        if "retry_delay_ms" in config:
            delay_ms = config["retry_delay_ms"]
            if not isinstance(delay_ms, int) or delay_ms < 0:
                issues.append(
                    ValidationIssue(
                        level=ValidationLevel.ERROR,
                        path=f"{path}.retry_delay_ms",
                        message="Retry delay must be a non-negative integer",
                        current_value=str(delay_ms),
                        suggestion="Use a positive integer in milliseconds",
                    )
                )
            elif delay_ms > self._timeout_schema.max_retry_delay_ms:
                issues.append(
                    ValidationIssue(
                        level=ValidationLevel.WARNING,
                        path=f"{path}.retry_delay_ms",
                        message=f"High retry delay may cause poor performance ({delay_ms}ms)",
                        current_value=str(delay_ms),
                        suggestion=f"Consider using at most {self._timeout_schema.max_retry_delay_ms}ms",
                    )
                )

        return issues

    def _validate_resources(self, resource_config: dict[str, Any]) -> List[ValidationIssue]:
        """Validate resource configuration"""
        issues = []
        base_path = ".resources"

        # Validate memory limits
        if "memory_limits" in resource_config:
            memory_config = resource_config["memory_limits"]
            if "default_mb" in memory_config:
                default_mb = memory_config["default_mb"]
                if not isinstance(default_mb, int) or default_mb <= 0:
                    issues.append(
                        ValidationIssue(
                            level=ValidationLevel.ERROR,
                            path=f"{base_path}.memory_limits.default_mb",
                            message="Default memory limit must be a positive integer",
                            current_value=str(default_mb),
                            suggestion="Use a positive integer in MB",
                        )
                    )

            if "max_mb" in memory_config:
                max_mb = memory_config["max_mb"]
                if not isinstance(max_mb, int) or max_mb <= 0:
                    issues.append(
                        ValidationIssue(
                            level=ValidationLevel.ERROR,
                            path=f"{base_path}.memory_limits.max_mb",
                            message="Maximum memory limit must be a positive integer",
                            current_value=str(max_mb),
                            suggestion="Use a positive integer in MB",
                        )
                    )

        # Validate worker configurations
        if "workers" in resource_config:
            workers_config = resource_config["workers"]
            if "default_max_workers" in workers_config:
                max_workers = workers_config["default_max_workers"]
                if not isinstance(max_workers, int) or max_workers <= 0:
                    issues.append(
                        ValidationIssue(
                            level=ValidationLevel.ERROR,
                            path=f"{base_path}.workers.default_max_workers",
                            message="Default max workers must be a positive integer",
                            current_value=str(max_workers),
                            suggestion="Use a positive integer between 1 and 8",
                        )
                    )
                elif max_workers > self._resource_schema.max_max_workers:
                    issues.append(
                        ValidationIssue(
                            level=ValidationLevel.WARNING,
                            path=f"{base_path}.workers.default_max_workers",
                            message=f"High worker count may impact system performance ({max_workers})",
                            current_value=str(max_workers),
                            suggestion=f"Consider using at most {self._resource_schema.max_max_workers} workers",
                        )
                    )

        return issues

    def _validate_performance(self, performance_config: dict[str, Any]) -> List[ValidationIssue]:
        """Validate performance-related configuration"""
        issues = []
        base_path = ".performance"

        # Validate cache configuration
        if "cache" in performance_config:
            cache_config = performance_config["cache"]
            if "size_limit" in cache_config:
                size_limit = cache_config["size_limit"]
                if not isinstance(size_limit, int) or size_limit <= 0:
                    issues.append(
                        ValidationIssue(
                            level=ValidationLevel.ERROR,
                            path=f"{base_path}.cache.size_limit",
                            message="Cache size limit must be a positive integer",
                            current_value=str(size_limit),
                            suggestion="Use a positive integer for cache size",
                        )
                    )

        # Validate monitoring settings
        if "monitoring" in performance_config:
            monitoring_config = performance_config["monitoring"]
            if "enabled" in monitoring_config:
                enabled = monitoring_config["enabled"]
                if not isinstance(enabled, bool):
                    issues.append(
                        ValidationIssue(
                            level=ValidationLevel.ERROR,
                            path=f"{base_path}.monitoring.enabled",
                            message="Monitoring enabled flag must be boolean",
                            current_value=str(enabled),
                            suggestion="Use true or false",
                        )
                    )

        return issues

    def _get_hook_template(self, hook_name: str) -> Dict[str, Any]:
        """Get configuration template for a hook type"""
        # Match hook patterns to templates
        for template_name, template in self._hook_templates.items():
            if template_name in hook_name.lower():
                return template
        return self._hook_templates["session_start"]  # Default template

    def _get_recommended_timeout(self, policy: str) -> int:
        """Get recommended timeout for a policy"""
        policy_timeouts = {"fast": 2000, "normal": 5000, "slow": 15000, "custom": 5000}
        return policy_timeouts.get(policy, 5000)

    def normalize_config(self, config: dict[str, Any]) -> Dict[str, Any]:
        """Normalize and apply defaults to configuration"""
        normalized = config.copy()

        # Ensure timeout_manager section exists
        if "timeout_manager" not in normalized:
            normalized["timeout_manager"] = {}

        timeout_manager = normalized["timeout_manager"]

        # Apply defaults to timeout_manager
        if "graceful_degradation" not in timeout_manager:
            timeout_manager["graceful_degradation"] = True

        if "default_retry_count" not in timeout_manager:
            timeout_manager["default_retry_count"] = self._timeout_schema.recommended_retry_count

        if "default_retry_delay_ms" not in timeout_manager:
            timeout_manager["default_retry_delay_ms"] = self._timeout_schema.default_retry_delay_ms

        # Ensure hook_configs section exists
        if "hook_configs" not in normalized:
            normalized["hook_configs"] = {}

        # Apply defaults to known hook types
        for hook_type, template in self._hook_templates.items():
            if hook_type not in normalized["hook_configs"]:
                # Look for existing hooks with this type
                matching_hooks = [k for k in normalized["hook_configs"].keys() if hook_type in k.lower()]
                if not matching_hooks:
                    # Add template as default
                    normalized["hook_configs"][hook_type] = template.copy()

        # Ensure resources section exists
        if "resources" not in normalized:
            normalized["resources"] = {}

        resources = normalized["resources"]
        if "memory_limits" not in resources:
            resources["memory_limits"] = {
                "default_mb": self._resource_schema.default_memory_limit_mb,
                "max_mb": self._resource_schema.max_memory_limit_mb,
            }

        if "workers" not in resources:
            resources["workers"] = {"default_max_workers": self._resource_schema.default_max_workers}

        # Ensure performance section exists
        if "performance" not in normalized:
            normalized["performance"] = {}

        performance = normalized["performance"]
        if "cache" not in performance:
            performance["cache"] = {"size_limit": self._resource_schema.cache_size_default}

        if "monitoring" not in performance:
            performance["monitoring"] = {"enabled": True}

        return normalized

    def generate_report(self, issues: list[ValidationIssue]) -> str:
        """Generate human-readable validation report"""
        if not issues:
            return "✅ Configuration validation passed with no issues found."

        lines = ["📋 Configuration Validation Report", ""]

        # Group issues by severity
        by_level: dict[ValidationLevel, List[ValidationIssue]] = {}
        for issue in issues:
            if issue.level not in by_level:
                by_level[issue.level] = []
            by_level[issue.level].append(issue)

        # Display issues by severity (most critical first)
        level_order = [
            ValidationLevel.CRITICAL,
            ValidationLevel.ERROR,
            ValidationLevel.WARNING,
            ValidationLevel.INFO,
        ]
        level_icons = {
            ValidationLevel.CRITICAL: "🚨",
            ValidationLevel.ERROR: "❌",
            ValidationLevel.WARNING: "⚠️",
            ValidationLevel.INFO: "ℹ️",
        }

        for level in level_order:
            if level in by_level:
                lines.append(f"{level_icons[level]} {level.value.upper()} ISSUES ({len(by_level[level])})")
                for issue in by_level[level]:
                    lines.append(f"   • {issue.path}: {issue.message}")
                    if issue.suggestion:
                        lines.append(f"     💡 Suggestion: {issue.suggestion}")
                lines.append("")

        return "\n".join(lines)

    def validate_and_fix_config_file(self, config_path: Path) -> Tuple[bool, Dict[str, Any], List[ValidationIssue]]:
        """Validate and optionally fix a configuration file"""
        try:
            with open(config_path, "r", encoding="utf-8") as f:
                config = json.load(f)

            # Validate configuration
            is_valid, issues = self.validate_config(config)

            if not is_valid:
                self._logger.warning(f"Configuration validation failed with {len(issues)} issues")

            # Normalize configuration (apply defaults, fix minor issues)
            normalized_config = self.normalize_config(config)

            return is_valid, normalized_config, issues

        except Exception as e:
            error_issues = [
                ValidationIssue(
                    level=ValidationLevel.CRITICAL,
                    path=".file",
                    message=f"Failed to load configuration file: {e}",
                    suggestion="Check file format and permissions",
                )
            ]
            return False, {}, error_issues


# Global validator instance
_validator = None


def get_config_validator() -> ConfigurationValidator:
    """Get the global configuration validator instance"""
    global _validator
    if _validator is None:
        _validator = ConfigurationValidator()
    return _validator


# Convenience functions
def validate_hook_config(
    config_path: Path | None = None,
) -> Tuple[bool, List[ValidationIssue]]:
    """Validate hooks configuration"""
    validator = get_config_validator()

    if config_path is None:
        config_path = Path(".moai/config/config.yaml")

    if not config_path.exists():
        # Return success with info about missing file
        return True, [
            ValidationIssue(
                level=ValidationLevel.INFO,
                path=".file",
                message="Configuration file not found, using defaults",
                suggestion="Create a configuration file to customize hook behavior",
            )
        ]

    is_valid, normalized_config, issues = validator.validate_and_fix_config_file(config_path)
    return is_valid, issues
