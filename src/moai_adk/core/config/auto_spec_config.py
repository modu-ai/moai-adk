"""Auto-Spec Completion Configuration Reader."""

import json
import logging
from pathlib import Path
from typing import Any, Dict, List

# Configure logging
logger = logging.getLogger(__name__)


class AutoSpecConfig:
    """
    Configuration reader for Auto-Spec Completion system.

    This class reads and validates the auto-spec completion configuration
    from the main MoAI configuration file.
    """

    def __init__(self, config_path: str = None):
        """Initialize the configuration reader."""
        self.config_path = config_path or self._get_default_config_path()
        self.config: dict[str, Any] = {}
        self.load_config()

    def _get_default_config_path(self) -> str:
        """Get the default configuration file path."""
        # Try to find config in multiple locations
        possible_paths = [
            Path.cwd() / ".moai" / "config.json",
            Path.cwd() / "config.json",
            Path.home() / ".moai" / "config.json",
        ]

        for path in possible_paths:
            if path.exists():
                return str(path)

        # Default to current directory
        return str(Path.cwd() / ".moai" / "config.json")

    def load_config(self) -> None:
        """Load configuration from file."""
        try:
            with open(self.config_path, "r", encoding="utf-8") as f:
                self.config = json.load(f)

            # Extract auto-spec completion config
            self.config = self.config.get("auto_spec_completion", {})

            logger.info(f"Loaded auto-spec completion configuration from {self.config_path}")

        except FileNotFoundError:
            logger.warning(f"Configuration file not found: {self.config_path}")
            self._load_default_config()
        except json.JSONDecodeError as e:
            logger.error(f"Invalid JSON in configuration file: {e}")
            self._load_default_config()
        except Exception as e:
            logger.error(f"Error loading configuration: {e}")
            self._load_default_config()

    def _load_default_config(self) -> None:
        """Load default configuration."""
        logger.info("Loading default auto-spec completion configuration")
        self.config = self._get_default_config()

    def _get_default_config(self) -> Dict[str, Any]:
        """Get default configuration."""
        return {
            "enabled": True,
            "trigger_tools": ["Write", "Edit", "MultiEdit"],
            "confidence_threshold": 0.7,
            "execution_timeout_ms": 1500,
            "quality_threshold": {
                "ears_compliance": 0.85,
                "min_content_length": 500,
                "max_review_suggestions": 10,
            },
            "excluded_patterns": [
                "test_*.py",
                "*_test.py",
                "*/tests/*",
                "*/__pycache__/*",
                "*/node_modules/*",
                "*/dist/*",
                "*/build/*",
            ],
            "domain_templates": {
                "enabled": True,
                "auto_detect": True,
                "supported_domains": ["auth", "api", "data", "ui", "business"],
                "fallback_domain": "general",
            },
            "spec_structure": {
                "include_meta": True,
                "include_traceability": True,
                "include_edit_guide": True,
                "required_sections": [
                    "Overview",
                    "Environment",
                    "Assumptions",
                    "Requirements",
                    "Specifications",
                    "Traceability",
                ],
            },
            "validation": {
                "enabled": True,
                "quality_grades": ["A", "B", "C", "D", "F"],
                "passing_grades": ["A", "B", "C"],
                "auto_improve": True,
                "max_iterations": 3,
            },
            "output": {
                "auto_create_files": True,
                "open_in_editor": True,
                "file_format": "markdown",
                "encoding": "utf-8",
            },
        }

    def is_enabled(self) -> bool:
        """Check if auto-spec completion is enabled."""
        return self.config.get("enabled", False)

    def get_trigger_tools(self) -> List[str]:
        """Get list of tools that trigger auto-spec completion."""
        return self.config.get("trigger_tools", [])

    def get_confidence_threshold(self) -> float:
        """Get confidence threshold for triggering auto-spec completion."""
        return self.config.get("confidence_threshold", 0.7)

    def get_execution_timeout_ms(self) -> int:
        """Get execution timeout in milliseconds."""
        return self.config.get("execution_timeout_ms", 1500)

    def get_quality_threshold(self) -> Dict[str, Any]:
        """Get quality threshold configuration."""
        return self.config.get("quality_threshold", {})

    def get_excluded_patterns(self) -> List[str]:
        """Get list of file patterns to exclude from auto-spec completion."""
        return self.config.get("excluded_patterns", [])

    def get_domain_templates_config(self) -> Dict[str, Any]:
        """Get domain templates configuration."""
        return self.config.get("domain_templates", {})

    def get_spec_structure_config(self) -> Dict[str, Any]:
        """Get spec structure configuration."""
        return self.config.get("spec_structure", {})

    def get_validation_config(self) -> Dict[str, Any]:
        """Get validation configuration."""
        return self.config.get("validation", {})

    def get_output_config(self) -> Dict[str, Any]:
        """Get output configuration."""
        return self.config.get("output", {})

    def should_exclude_file(self, file_path: str) -> bool:
        """Check if a file should be excluded from auto-spec completion."""
        excluded_patterns = self.get_excluded_patterns()

        if not excluded_patterns:
            return False

        # Convert file path to normalized string
        normalized_path = str(Path(file_path)).lower()

        for pattern in excluded_patterns:
            # Convert pattern to lowercase for case-insensitive matching
            pattern.lower()

            # Handle directory patterns
            if pattern.startswith("*/") and pattern.endswith("/*"):
                dir_pattern = pattern[2:-2]  # Remove */ and /*
                if dir_pattern in normalized_path:
                    return True
            # Handle file patterns
            elif "*" in pattern:
                # Simple wildcard matching
                regex_pattern = pattern.replace("*", ".*")
                import re

                if re.search(regex_pattern, normalized_path):
                    return True
            # Exact match
            elif pattern in normalized_path:
                return True

        return False

    def get_required_sections(self) -> List[str]:
        """Get list of required SPEC sections."""
        return self.config.get("spec_structure", {}).get("required_sections", [])

    def get_supported_domains(self) -> List[str]:
        """Get list of supported domains."""
        return self.config.get("domain_templates", {}).get("supported_domains", [])

    def get_fallback_domain(self) -> str:
        """Get fallback domain for unsupported domains."""
        return self.config.get("domain_templates", {}).get("fallback_domain", "general")

    def should_include_meta(self) -> bool:
        """Check if meta information should be included."""
        return self.config.get("spec_structure", {}).get("include_meta", True)

    def should_include_traceability(self) -> bool:
        """Check if traceability information should be included."""
        return self.config.get("spec_structure", {}).get("include_traceability", True)

    def should_include_edit_guide(self) -> bool:
        """Check if edit guide should be included."""
        return self.config.get("spec_structure", {}).get("include_edit_guide", True)

    def get_passing_quality_grades(self) -> List[str]:
        """Get list of passing quality grades."""
        return self.config.get("validation", {}).get("passing_grades", ["A", "B", "C"])

    def should_auto_improve(self) -> bool:
        """Check if auto-improvement is enabled."""
        return self.config.get("validation", {}).get("auto_improve", True)

    def get_max_improvement_iterations(self) -> int:
        """Get maximum improvement iterations."""
        return self.config.get("validation", {}).get("max_iterations", 3)

    def should_auto_create_files(self) -> bool:
        """Check if auto-creation of files is enabled."""
        return self.config.get("output", {}).get("auto_create_files", True)

    def should_open_in_editor(self) -> bool:
        """Check if files should be opened in editor."""
        return self.config.get("output", {}).get("open_in_editor", True)

    def get_file_format(self) -> str:
        """Get output file format."""
        return self.config.get("output", {}).get("file_format", "markdown")

    def get_encoding(self) -> str:
        """Get output encoding."""
        return self.config.get("output", {}).get("encoding", "utf-8")

    def is_validation_enabled(self) -> bool:
        """Check if validation is enabled."""
        return self.config.get("validation", {}).get("enabled", True)

    def is_domain_detection_enabled(self) -> bool:
        """Check if domain detection is enabled."""
        return self.config.get("domain_templates", {}).get("auto_detect", True)

    def validate_config(self) -> List[str]:
        """Validate configuration and return list of errors."""
        errors = []

        # Check required fields
        if not isinstance(self.config.get("enabled"), bool):
            errors.append("enabled must be a boolean")

        if not isinstance(self.config.get("confidence_threshold"), (int, float)):
            errors.append("confidence_threshold must be a number")
        elif not 0 <= self.config.get("confidence_threshold") <= 1:
            errors.append("confidence_threshold must be between 0 and 1")

        if not isinstance(self.config.get("execution_timeout_ms"), int):
            errors.append("execution_timeout_ms must be an integer")
        elif self.config.get("execution_timeout_ms") <= 0:
            errors.append("execution_timeout_ms must be positive")

        # Check trigger tools
        trigger_tools = self.config.get("trigger_tools", [])
        if not isinstance(trigger_tools, list):
            errors.append("trigger_tools must be a list")
        else:
            for tool in trigger_tools:
                if not isinstance(tool, str):
                    errors.append("All trigger_tools must be strings")

        # Check excluded patterns
        excluded_patterns = self.config.get("excluded_patterns", [])
        if not isinstance(excluded_patterns, list):
            errors.append("excluded_patterns must be a list")
        else:
            for pattern in excluded_patterns:
                if not isinstance(pattern, str):
                    errors.append("All excluded_patterns must be strings")

        # Check quality threshold
        quality_threshold = self.config.get("quality_threshold", {})
        if not isinstance(quality_threshold, dict):
            errors.append("quality_threshold must be a dictionary")
        else:
            if "ears_compliance" in quality_threshold:
                if not isinstance(quality_threshold["ears_compliance"], (int, float)):
                    errors.append("quality_threshold.ears_compliance must be a number")
                elif not 0 <= quality_threshold["ears_compliance"] <= 1:
                    errors.append("quality_threshold.ears_compliance must be between 0 and 1")

        return errors

    def to_dict(self) -> Dict[str, Any]:
        """Convert configuration to dictionary."""
        return self.config.copy()

    def update_config(self, updates: Dict[str, Any]) -> None:
        """Update configuration with new values."""
        self.config.update(updates)
        logger.info(f"Updated auto-spec completion configuration: {updates}")

    def save_config(self) -> None:
        """Save configuration to file."""
        try:
            # Load the full config file
            with open(self.config_path, "r", encoding="utf-8") as f:
                full_config = json.load(f)

            # Update the auto-spec completion config
            full_config["auto_spec_completion"] = self.config

            # Save back to file
            with open(self.config_path, "w", encoding="utf-8") as f:
                json.dump(full_config, f, indent=2, ensure_ascii=False)

            logger.info(f"Saved auto-spec completion configuration to {self.config_path}")

        except Exception as e:
            logger.error(f"Error saving configuration: {e}")
            raise

    def __str__(self) -> str:
        """String representation of configuration."""
        return json.dumps(self.config, indent=2, ensure_ascii=False)

    def __repr__(self) -> str:
        """String representation for debugging."""
        return f"AutoSpecConfig(enabled={self.is_enabled()}, config_path='{self.config_path}')"
