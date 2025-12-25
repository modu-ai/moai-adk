# # REMOVED_ORPHAN_CODE:TEMPLATE-001 | SPEC: SPEC-INIT-003/spec.md | Chain: TEMPLATE-001
"""Enhanced Template copy and backup processor with improved version handling and validation.

SPEC-INIT-003 v0.3.0: preserve user content
Enhanced with:
- Comprehensive version field management
- Template substitution validation
- Performance optimization
- Error handling improvements
- Configuration-driven behavior
"""

from __future__ import annotations

import json
import logging
import re
import shutil
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Dict, List, Optional

from rich.console import Console

from moai_adk.core.template.backup import TemplateBackup
from moai_adk.core.template.merger import TemplateMerger
from moai_adk.statusline.version_reader import VersionConfig, VersionReader

console = Console()


@dataclass
class TemplateProcessorConfig:
    """Configuration for TemplateProcessor behavior."""

    # Version handling configuration
    version_cache_ttl_seconds: int = 120
    version_fallback: str = "unknown"
    version_format_regex: str = r"^v?(\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?)$"
    enable_version_validation: bool = True
    preserve_user_version: bool = True

    # Template substitution configuration
    validate_template_variables: bool = True
    max_variable_length: int = 50
    allowed_variable_pattern: str = r"^[A-Z_]+$"
    enable_substitution_warnings: bool = True

    # Performance configuration
    enable_caching: bool = True
    cache_size: int = 100
    async_operations: bool = False

    # Error handling configuration
    graceful_degradation: bool = True
    verbose_logging: bool = False

    @classmethod
    def from_dict(cls, config_dict: Dict[str, Any]) -> "TemplateProcessorConfig":
        """Create config from dictionary."""
        config_dict = config_dict or {}
        return cls(
            version_cache_ttl_seconds=config_dict.get("version_cache_ttl_seconds", 120),
            version_fallback=config_dict.get("version_fallback", "unknown"),
            version_format_regex=config_dict.get("version_format_regex", r"^v?(\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?)$"),
            enable_version_validation=config_dict.get("enable_version_validation", True),
            preserve_user_version=config_dict.get("preserve_user_version", True),
            validate_template_variables=config_dict.get("validate_template_variables", True),
            max_variable_length=config_dict.get("max_variable_length", 50),
            allowed_variable_pattern=config_dict.get("allowed_variable_pattern", r"^[A-Z_]+$"),
            enable_substitution_warnings=config_dict.get("enable_substitution_warnings", True),
            enable_caching=config_dict.get("enable_caching", True),
            cache_size=config_dict.get("cache_size", 100),
            async_operations=config_dict.get("async_operations", False),
            graceful_degradation=config_dict.get("graceful_degradation", True),
            verbose_logging=config_dict.get("verbose_logging", False),
        )


class TemplateProcessor:
    """Orchestrate template copying and backups with enhanced version handling and validation."""

    # User data protection paths (never touch) - SPEC-INIT-003 v0.3.0
    PROTECTED_PATHS = [
        ".moai/specs/",  # User SPEC documents
        ".moai/reports/",  # User reports
        ".moai/project/",  # User project documents (product/structure/tech.md)
        # config.json is now FORCE OVERWRITTEN (backup in .moai-backups/)
        # Merge via /moai:0-project when optimized=false
    ]

    # Paths excluded from backups
    BACKUP_EXCLUDE = PROTECTED_PATHS

    # Common template variables with validation hints
    COMMON_TEMPLATE_VARIABLES = {
        "PROJECT_DIR": "Cross-platform project path (run /moai:0-project to set)",
        "PROJECT_NAME": "Project name (run /moai:0-project to set)",
        "AUTHOR": "Project author (run /moai:0-project to set)",
        "CONVERSATION_LANGUAGE": "Interface language (run /moai:0-project to set)",
        "MOAI_VERSION": "MoAI-ADK version (should be set automatically)",
        "MOAI_VERSION_SHORT": "Short MoAI-ADK version (without 'v' prefix)",
        "MOAI_VERSION_DISPLAY": "Display version with proper formatting",
        "MOAI_VERSION_TRIMMED": "Trimmed version for UI displays",
        "MOAI_VERSION_SEMVER": "Semantic version format (major.minor.patch)",
        "MOAI_VERSION_VALID": "Version validation status",
        "MOAI_VERSION_SOURCE": "Version source information",
        "MOAI_VERSION_CACHE_AGE": "Cache age for debugging",
        "CREATION_TIMESTAMP": "Project creation timestamp",
        "STATUSLINE_COMMAND": "Cross-platform statusline command (OS-specific)",
    }

    def __init__(self, target_path: Path, config: Optional[TemplateProcessorConfig] = None) -> None:
        """Initialize the processor with enhanced configuration.

        Args:
            target_path: Project path.
            config: Optional configuration for processor behavior.
        """
        self.target_path = target_path.resolve()
        self.template_root = self._get_template_root()
        self.backup = TemplateBackup(self.target_path)
        self.merger = TemplateMerger(self.target_path)
        self.context: dict[str, str] = {}  # Template variable substitution context
        self._version_reader: VersionReader | None = None
        self.config = config or TemplateProcessorConfig()
        self._substitution_cache: Dict[
            int, tuple[str, list[str]]
        ] = {}  # Cache for substitution results (key: hash, value: (content, warnings))
        self._variable_validation_cache: Dict[str, bool] = {}  # Cache for variable validation
        self.logger = logging.getLogger(__name__)

        if self.config.verbose_logging:
            self.logger.info(f"TemplateProcessor initialized with config: {self.config}")

    def set_context(self, context: dict[str, str]) -> None:
        """Set variable substitution context with enhanced validation.

        Args:
            context: Dictionary of template variables.
        """
        self.context = context
        self._substitution_cache.clear()  # Clear cache when context changes
        self._variable_validation_cache.clear()

        if self.config.verbose_logging:
            self.logger.debug(f"Context set with {len(context)} variables")

        # Validate template variables if enabled
        if self.config.validate_template_variables:
            self._validate_template_variables(context)

        # Add deprecation mapping for HOOK_PROJECT_DIR
        if "PROJECT_DIR" in self.context and "HOOK_PROJECT_DIR" not in self.context:
            self.context["HOOK_PROJECT_DIR"] = self.context["PROJECT_DIR"]

    def _get_version_reader(self) -> VersionReader:
        """
        Get or create version reader instance with enhanced configuration.

        Returns:
            VersionReader instance
        """
        if self._version_reader is None:
            version_config = VersionConfig(
                cache_ttl_seconds=self.config.version_cache_ttl_seconds,
                fallback_version=self.config.version_fallback,
                version_format_regex=self.config.version_format_regex,
                debug_mode=self.config.verbose_logging,
            )
            self._version_reader = VersionReader(version_config)

            if self.config.verbose_logging:
                self.logger.info("VersionReader created with enhanced configuration")
        return self._version_reader

    def _validate_template_variables(self, context: Dict[str, str]) -> None:
        """
        Validate template variables with comprehensive checking.

        Args:
            context: Dictionary of template variables to validate
        """
        import re

        if not self.config.validate_template_variables:
            return

        validation_errors: List[str] = []
        warning_messages: List[str] = []

        # Check variable names against pattern
        variable_pattern = re.compile(self.config.allowed_variable_pattern)

        for var_name, var_value in context.items():
            # Check variable name format
            if not variable_pattern.match(var_name):
                validation_errors.append(f"Invalid variable name format: '{var_name}'")
                continue

            # Check variable length
            if len(var_name) > self.config.max_variable_length:
                warning_messages.append(f"Variable name '{var_name}' exceeds maximum length")

            # Check variable value length
            if len(var_value) > self.config.max_variable_length * 2:
                warning_messages.append(f"Variable value '{var_value[:20]}...' is very long")

            # Check for potentially dangerous values
            if "{{" in var_value or "}}" in var_value:
                warning_messages.append(f"Variable '{var_name}' contains placeholder patterns")

        # Check for common variables that should be present
        missing_common_vars = []
        for common_var in self.COMMON_TEMPLATE_VARIABLES:
            if common_var not in context:
                missing_common_vars.append(common_var)

        if missing_common_vars and self.config.enable_substitution_warnings:
            warning_messages.append(f"Common variables missing: {', '.join(missing_common_vars[:3])}")

        # Report validation results
        if validation_errors and not self.config.graceful_degradation:
            raise ValueError(f"Template variable validation failed: {validation_errors}")

        if validation_errors and self.config.graceful_degradation:
            self.logger.warning(f"Template variable validation warnings: {validation_errors}")

        if warning_messages and self.config.enable_substitution_warnings:
            self.logger.warning(f"Template variable warnings: {warning_messages}")

        if self.config.verbose_logging:
            self.logger.debug(f"Template variables validated: {len(context)} variables checked")

    def get_enhanced_version_context(self) -> dict[str, str]:
        """
        Get enhanced version context with proper error handling and caching.

        Returns comprehensive version information including multiple format options
        and debugging information.

        Returns:
            Dictionary containing enhanced version-related template variables
        """
        version_context = {}
        logger = logging.getLogger(__name__)

        try:
            version_reader = self._get_version_reader()
            moai_version = version_reader.get_version()

            # Basic version information
            version_context["MOAI_VERSION"] = moai_version
            version_context["MOAI_VERSION_SHORT"] = self._format_short_version(moai_version)
            version_context["MOAI_VERSION_DISPLAY"] = self._format_display_version(moai_version)

            # Enhanced formatting options
            version_context["MOAI_VERSION_TRIMMED"] = self._format_trimmed_version(moai_version, max_length=10)
            version_context["MOAI_VERSION_SEMVER"] = self._format_semver_version(moai_version)

            # Validation and source information
            version_context["MOAI_VERSION_VALID"] = "true" if moai_version != "unknown" else "false"
            version_context["MOAI_VERSION_SOURCE"] = self._get_version_source(version_reader)

            # Performance metrics
            cache_age = version_reader.get_cache_age_seconds()
            if cache_age is not None:
                version_context["MOAI_VERSION_CACHE_AGE"] = f"{cache_age:.2f}s"
            else:
                version_context["MOAI_VERSION_CACHE_AGE"] = "uncached"

            # Additional metadata
            if self.config.enable_version_validation:
                is_valid = self._is_valid_version_format(moai_version)
                version_context["MOAI_VERSION_FORMAT_VALID"] = "true" if is_valid else "false"

            if self.config.verbose_logging:
                logger.debug(f"Enhanced version context generated: {version_context}")

        except Exception as e:
            logger.warning(f"Failed to read version for template context: {e}")
            # Use fallback version with comprehensive formatting
            fallback_version = self.config.version_fallback
            version_context["MOAI_VERSION"] = fallback_version
            version_context["MOAI_VERSION_SHORT"] = self._format_short_version(fallback_version)
            version_context["MOAI_VERSION_DISPLAY"] = self._format_display_version(fallback_version)
            version_context["MOAI_VERSION_TRIMMED"] = self._format_trimmed_version(fallback_version, max_length=10)
            version_context["MOAI_VERSION_SEMVER"] = self._format_semver_version(fallback_version)
            version_context["MOAI_VERSION_VALID"] = "false" if fallback_version == "unknown" else "true"
            version_context["MOAI_VERSION_SOURCE"] = "fallback_config"
            version_context["MOAI_VERSION_CACHE_AGE"] = "unavailable"
            version_context["MOAI_VERSION_FORMAT_VALID"] = "false"

        return version_context

    def _is_valid_version_format(self, version: str) -> bool:
        """
        Validate version format using configured regex pattern.

        Args:
            version: Version string to validate

        Returns:
            True if version format is valid
        """
        import re

        try:
            pattern = re.compile(self.config.version_format_regex)
            return bool(pattern.match(version))
        except re.error:
            # Fallback to default pattern if custom one is invalid
            default_pattern = re.compile(r"^v?(\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?)$")
            return bool(default_pattern.match(version))

    def _format_short_version(self, version: str) -> str:
        """
        Format short version by removing 'v' prefix if present.

        Args:
            version: Version string

        Returns:
            Short version string
        """
        return version[1:] if version.startswith("v") else version

    def _format_display_version(self, version: str) -> str:
        """
        Format display version with proper formatting.

        Args:
            version: Version string

        Returns:
            Display version string
        """
        if version == "unknown":
            return "MoAI-ADK unknown version"
        elif version.startswith("v"):
            return f"MoAI-ADK {version}"
        else:
            return f"MoAI-ADK v{version}"

    def _format_trimmed_version(self, version: str, max_length: int = 10) -> str:
        """
        Format version with maximum length, suitable for UI displays.

        Args:
            version: Version string
            max_length: Maximum allowed length for the version string

        Returns:
            Trimmed version string
        """
        if version == "unknown":
            return "unknown"

        # Remove 'v' prefix for trimming
        clean_version = version[1:] if version.startswith("v") else version

        # Trim if necessary
        if len(clean_version) > max_length:
            return clean_version[:max_length]
        return clean_version

    def _format_semver_version(self, version: str) -> str:
        """
        Format version as semantic version with major.minor.patch structure.

        Args:
            version: Version string

        Returns:
            Semantic version string
        """
        if version == "unknown":
            return "0.0.0"

        # Remove 'v' prefix and extract semantic version
        clean_version = version[1:] if version.startswith("v") else version

        # Extract core semantic version (remove pre-release and build metadata)
        import re

        semver_match = re.match(r"^(\d+\.\d+\.\d+)", clean_version)
        if semver_match:
            return semver_match.group(1)
        return "0.0.0"

    def _get_version_source(self, version_reader: VersionReader) -> str:
        """
        Determine the source of the version information.

        Args:
            version_reader: VersionReader instance

        Returns:
            String indicating version source
        """
        config = version_reader.get_config()
        cache_age = version_reader.get_cache_age_seconds()

        if cache_age is not None and cache_age < config.cache_ttl_seconds:
            return "config_cached"
        elif cache_age is not None:
            return "config_stale"
        else:
            return config.fallback_source.value

    def _get_template_root(self) -> Path:
        """Return the template root path."""
        # src/moai_adk/core/template/processor.py → src/moai_adk/templates/
        current_file = Path(__file__).resolve()
        package_root = current_file.parent.parent.parent
        return package_root / "templates"

    def _substitute_variables(self, content: str) -> tuple[str, list[str]]:
        """
        Substitute template variables in content with enhanced validation and caching.

        Args:
            content: Content to substitute variables in

        Returns:
            Tuple of (substituted_content, warnings_list)
        """
        warnings = []
        logger = logging.getLogger(__name__)

        # Check cache first if enabled
        cache_key = hash((frozenset(self.context.items()), content[:1000]))
        if self.config.enable_caching and cache_key in self._substitution_cache:
            cached_result = self._substitution_cache[cache_key]
            if self.config.verbose_logging:
                logger.debug("Using cached substitution result")
            return cached_result

        # Enhanced variable substitution with validation
        substitution_count = 0
        for key, value in self.context.items():
            placeholder = f"{{{{{key}}}}}"  # {{KEY}}
            if placeholder in content:
                if self.config.validate_template_variables:
                    # Validate variable before substitution
                    if not self._is_valid_template_variable(key, value):
                        warnings.append(f"Invalid variable {key} - skipped substitution")
                        continue

                safe_value = self._sanitize_value(value)
                content = content.replace(placeholder, safe_value)
                substitution_count += 1

                if self.config.verbose_logging:
                    logger.debug(f"Substituted {key}: {safe_value[:50]}...")

        # Detect unsubstituted variables with enhanced error messages
        remaining = re.findall(r"\{\{([A-Z_]+)\}\}", content)
        if remaining:
            unique_remaining = sorted(set(remaining))

            # Build detailed warning message with enhanced suggestions
            warning_parts = []
            for var in unique_remaining:
                if var in self.COMMON_TEMPLATE_VARIABLES:
                    suggestion = self.COMMON_TEMPLATE_VARIABLES[var]
                    warning_parts.append(f"{{{{{var}}}}} → {suggestion}")
                else:
                    warning_parts.append(f"{{{{{var}}}}} → Unknown variable (check template)")

            warnings.append("Template variables not substituted:")
            warnings.extend(f"  • {part}" for part in warning_parts)

            if self.config.enable_substitution_warnings:
                warnings.append("💡 Run 'uv run moai-adk update' to fix template variables")

        # Add performance information if verbose logging is enabled
        if self.config.verbose_logging:
            warnings.append(f"  📊 Substituted {substitution_count} variables")

        # Cache the result if enabled
        if self.config.enable_caching:
            result = (content, warnings)
            self._substitution_cache[cache_key] = result

            # Manage cache size
            if len(self._substitution_cache) > self.config.cache_size:
                # Remove oldest entry (simple FIFO)
                oldest_key = next(iter(self._substitution_cache))
                del self._substitution_cache[oldest_key]
                if self.config.verbose_logging:
                    logger.debug("Cache size limit reached, removed oldest entry")

        return content, warnings

    def _is_valid_template_variable(self, key: str, value: str) -> bool:
        """
        Validate a template variable before substitution.

        Args:
            key: Variable name
            value: Variable value

        Returns:
            True if variable is valid
        """
        import re

        # Check variable name format
        if not re.match(self.config.allowed_variable_pattern, key):
            return False

        # Check variable length
        if len(key) > self.config.max_variable_length:
            return False

        # Check value length
        if len(value) > self.config.max_variable_length * 2:
            return False

        # Note: {{ }} patterns are handled by sanitization, not validation

        # Check for empty values
        if not value.strip():
            return False

        return True

    def clear_substitution_cache(self) -> None:
        """Clear the substitution cache."""
        self._substitution_cache.clear()
        if self.config.verbose_logging:
            self.logger.debug("Substitution cache cleared")

    def get_cache_stats(self) -> Dict[str, Any]:
        """
        Get cache statistics.

        Returns:
            Dictionary containing cache statistics
        """
        return {
            "cache_size": len(self._substitution_cache),
            "max_cache_size": self.config.cache_size,
            "cache_enabled": self.config.enable_caching,
            "cache_hit_ratio": 0.0,  # Would need to track hits to implement this
        }

    def _sanitize_value(self, value: str) -> str:
        """Sanitize value to prevent recursive substitution and control characters.

        Args:
            value: Value to sanitize.

        Returns:
            Sanitized value.
        """
        # Remove control characters (keep printable and whitespace)
        value = "".join(c for c in value if c.isprintable() or c in "\n\r\t")
        # Prevent recursive substitution by removing placeholder patterns
        value = value.replace("{{", "").replace("}}", "")
        return value

    def _is_text_file(self, file_path: Path) -> bool:
        """Check if file is text-based (not binary).

        Args:
            file_path: File path to check.

        Returns:
            True if file is text-based.
        """
        text_extensions = {
            ".md",
            ".json",
            ".txt",
            ".py",
            ".ts",
            ".js",
            ".yaml",
            ".yml",
            ".toml",
            ".xml",
            ".sh",
            ".bash",
        }
        return file_path.suffix.lower() in text_extensions

    def _localize_yaml_description(self, content: str, language: str = "en") -> str:
        """Localize multilingual YAML description field.

        Converts multilingual description maps to single-language strings:
        description:
          en: "English text"
          ko: "Korean text"
        →
        description: "Korean text"  (if language="ko")

        Args:
            content: File content.
            language: Target language code (en, ko, ja, zh).

        Returns:
            Content with localized descriptions.
        """
        import yaml  # type: ignore[import-untyped]

        # Pattern to match YAML frontmatter
        frontmatter_pattern = r"^---\n(.*?)\n---"
        match = re.match(frontmatter_pattern, content, re.DOTALL)

        if not match:
            return content

        try:
            yaml_content = match.group(1)
            yaml_data = yaml.safe_load(yaml_content)

            # Check if description is a dict (multilingual)
            if isinstance(yaml_data.get("description"), dict):
                # Select language (fallback to English)
                descriptions = yaml_data["description"]
                selected_desc = descriptions.get(language, descriptions.get("en", ""))

                # Replace description with selected language
                yaml_data["description"] = selected_desc

                # Reconstruct frontmatter
                new_yaml = yaml.dump(yaml_data, allow_unicode=True, sort_keys=False)
                # Preserve the rest of the content
                rest_content = content[match.end() :]
                return f"---\n{new_yaml}---{rest_content}"

        except Exception:
            # If YAML parsing fails, return original content
            pass

        return content

    def _copy_file_with_substitution(self, src: Path, dst: Path) -> list[str]:
        """Copy file with variable substitution and description localization for text files.

        Args:
            src: Source file path.
            dst: Destination file path.

        Returns:
            List of warnings.
        """
        import stat

        warnings = []

        # Text files: read, substitute, write
        if self._is_text_file(src) and self.context:
            try:
                content = src.read_text(encoding="utf-8")
                content, file_warnings = self._substitute_variables(content)

                # Apply description localization for command/output-style files
                if src.suffix == ".md" and ("commands/alfred" in str(src) or "output-styles/alfred" in str(src)):
                    lang = self.context.get("CONVERSATION_LANGUAGE", "en")
                    content = self._localize_yaml_description(content, lang)

                dst.write_text(content, encoding="utf-8")
                warnings.extend(file_warnings)
            except UnicodeDecodeError:
                # Binary file fallback
                shutil.copy2(src, dst)
        else:
            # Binary file or no context: simple copy
            shutil.copy2(src, dst)

        # Ensure executable permission for shell scripts
        if src.suffix == ".sh":
            # Always make shell scripts executable regardless of source permissions
            dst_mode = dst.stat().st_mode | stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH
            dst.chmod(dst_mode)

        return warnings

    def _copy_dir_with_substitution(self, src: Path, dst: Path) -> None:
        """Recursively copy directory with variable substitution for text files.

        Args:
            src: Source directory path.
            dst: Destination directory path.
        """
        dst.mkdir(parents=True, exist_ok=True)

        for item in src.rglob("*"):
            rel_path = item.relative_to(src)
            dst_item = dst / rel_path

            if item.is_file():
                # Create parent directory if needed
                dst_item.parent.mkdir(parents=True, exist_ok=True)
                # Copy with variable substitution
                self._copy_file_with_substitution(item, dst_item)
            elif item.is_dir():
                dst_item.mkdir(parents=True, exist_ok=True)

    def copy_templates(self, backup: bool = True, silent: bool = False) -> None:
        """Copy template files into the project.

        Args:
            backup: Whether to create a backup.
            silent: Reduce log output when True.
        """
        # 1. Create a backup when existing files are present
        if backup and self._has_existing_files():
            backup_path = self.create_backup()
            if not silent:
                console.print(f"💾 Backup created: {backup_path.name}")

        # 2. Copy templates
        if not silent:
            console.print("📄 Copying templates...")

        self._copy_claude(silent)
        self._copy_moai(silent)
        self._copy_github(silent)
        self._copy_claude_md(silent)
        self._copy_gitignore(silent)
        self._copy_mcp_json(silent)

        if not silent:
            console.print("✅ Templates copied successfully")

    def _has_existing_files(self) -> bool:
        """Determine whether project files exist (backup decision helper)."""
        return self.backup.has_existing_files()

    def create_backup(self) -> Path:
        """Create a timestamped backup (delegated)."""
        return self.backup.create_backup()

    def _copy_exclude_protected(self, src: Path, dst: Path) -> None:
        """Copy content while excluding protected paths.

        Args:
            src: Source directory.
            dst: Destination directory.
        """
        dst.mkdir(parents=True, exist_ok=True)

        # PROTECTED_PATHS: only specs/ and reports/ are excluded during copying
        # project/ and config.json are preserved only when they already exist
        template_protected_paths = [
            "specs",
            "reports",
        ]

        for item in src.rglob("*"):
            rel_path = item.relative_to(src)
            rel_path_str = str(rel_path)

            # Skip template copy for specs/ and reports/
            if any(rel_path_str.startswith(p) for p in template_protected_paths):
                continue

            dst_item = dst / rel_path
            if item.is_file():
                # Preserve user content by skipping existing files (v0.3.0)
                # This automatically protects project/ and config.json
                if dst_item.exists():
                    continue
                dst_item.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(item, dst_item)
            elif item.is_dir():
                dst_item.mkdir(parents=True, exist_ok=True)

    def _copy_claude(self, silent: bool = False) -> None:
        """.claude/ directory copy with variable substitution (selective with alfred folder overwrite).


        Strategy:
        - Alfred folders (commands/agents/hooks/output-styles/alfred) → copy wholesale (delete & overwrite)
          * Creates individual backup before deletion for safety
          * Commands: 0-project.md, 1-plan.md, 2-run.md, 3-sync.md
        - Other files/folders → copy individually (preserve existing)
        """
        src = self.template_root / ".claude"
        dst = self.target_path / ".claude"

        if not src.exists():
            if not silent:
                console.print("⚠️ .claude/ template not found")
            return

        # Create .claude directory if not exists
        dst.mkdir(parents=True, exist_ok=True)

        # Alfred and Moai folders to copy wholesale (overwrite)
        # Including both legacy alfred/ and new moai/ structure
        alfred_moai_folders = [
            "hooks/alfred",
            "hooks/moai",
            "commands/alfred",  # Contains 0-project.md, 1-plan.md, 2-run.md, 3-sync.md
            "commands/moai",
            "output-styles/moai",
            "agents/alfred",
            "agents/moai",
            "skills",  # NEW: Complete replacement for skills folder
        ]

        # 1. Copy Alfred and Moai folders wholesale (backup before delete & overwrite)
        for folder in alfred_moai_folders:
            src_folder = src / folder
            dst_folder = dst / folder

            if src_folder.exists():
                # Remove existing folder (backup is already handled by create_backup() in update.py)
                if dst_folder.exists():
                    shutil.rmtree(dst_folder)

                # Create parent directory if needed
                dst_folder.parent.mkdir(parents=True, exist_ok=True)
                shutil.copytree(src_folder, dst_folder)
                if not silent:
                    console.print(f"   ✅ .claude/{folder}/ overwritten")

        # 1.5 Copy other subdirectories in parent folders (e.g., output-styles/moai, hooks/shared)
        # This ensures non-alfred subdirectories are also copied
        parent_folders_with_subdirs = ["output-styles", "hooks", "commands", "agents"]
        for parent_name in parent_folders_with_subdirs:
            src_parent = src / parent_name
            if not src_parent.exists():
                continue

            for subdir in src_parent.iterdir():
                if not subdir.is_dir():
                    continue

                # Skip alfred subdirectories (already handled above)
                if subdir.name == "alfred":
                    continue

                rel_subdir = f"{parent_name}/{subdir.name}"
                dst_subdir = dst / parent_name / subdir.name

                if dst_subdir.exists():
                    # For non-alfred directories, overwrite with merge if necessary
                    shutil.rmtree(dst_subdir)

                # Copy the subdirectory
                shutil.copytree(subdir, dst_subdir)
                if not silent:
                    console.print(f"   ✅ .claude/{rel_subdir}/ copied")

        # 2. Copy other files/folders individually (smart merge for settings.json and config.json)
        all_warnings = []
        for item in src.iterdir():
            rel_path = item.relative_to(src)
            dst_item = dst / rel_path

            # Skip Alfred parent folders (already handled above)
            if item.is_dir() and item.name in [
                "hooks",
                "commands",
                "output-styles",
                "agents",
            ]:
                continue

            if item.is_file():
                # Smart merge for settings.json
                if item.name == "settings.json":
                    self._merge_settings_json(item, dst_item)
                    # Apply variable substitution to merged settings.json (for cross-platform Hook paths)
                    if self.context:
                        content = dst_item.read_text(encoding="utf-8")
                        content, file_warnings = self._substitute_variables(content)
                        dst_item.write_text(content, encoding="utf-8")
                        all_warnings.extend(file_warnings)
                    if not silent:
                        console.print("   🔄 settings.json merged (Hook paths configured for your OS)")
                # Smart merge for config.json
                elif item.name == "config.json":
                    self._merge_config_json(item, dst_item)
                    if not silent:
                        console.print("   🔄 config.json merged (user preferences preserved)")
                else:
                    # FORCE OVERWRITE: Always copy other files (no skip)
                    warnings = self._copy_file_with_substitution(item, dst_item)
                    all_warnings.extend(warnings)
            elif item.is_dir():
                # FORCE OVERWRITE: Always copy directories (no skip)
                self._copy_dir_with_substitution(item, dst_item)

        # Print warnings if any
        if all_warnings and not silent:
            console.print("[yellow]⚠️ Template warnings:[/yellow]")
            for warning in set(all_warnings):  # Deduplicate
                console.print(f"   {warning}")

        if not silent:
            console.print("   ✅ .claude/ copy complete (variables substituted)")

    def _copy_moai(self, silent: bool = False) -> None:
        """.moai/ directory copy with variable substitution (excludes protected paths)."""
        src = self.template_root / ".moai"
        dst = self.target_path / ".moai"

        if not src.exists():
            if not silent:
                console.print("⚠️ .moai/ template not found")
            return

        # Paths excluded from template copying (specs/, reports/, .moai/config/config.json)
        template_protected_paths = [
            "specs",
            "reports",
            ".moai/config/config.json",
        ]

        all_warnings = []

        # Copy while skipping protected paths
        for item in src.rglob("*"):
            rel_path = item.relative_to(src)
            rel_path_str = str(rel_path)

            # Skip specs/ and reports/
            if any(rel_path_str.startswith(p) for p in template_protected_paths):
                continue

            dst_item = dst / rel_path
            if item.is_file():
                # FORCE OVERWRITE: Always copy files (no skip)
                dst_item.parent.mkdir(parents=True, exist_ok=True)
                # Copy with variable substitution
                warnings = self._copy_file_with_substitution(item, dst_item)
                all_warnings.extend(warnings)
            elif item.is_dir():
                dst_item.mkdir(parents=True, exist_ok=True)

        # Print warnings if any
        if all_warnings and not silent:
            console.print("[yellow]⚠️ Template warnings:[/yellow]")
            for warning in set(all_warnings):  # Deduplicate
                console.print(f"   {warning}")

        if not silent:
            console.print("   ✅ .moai/ copy complete (variables substituted)")

    def _copy_github(self, silent: bool = False) -> None:
        """.github/ directory copy with smart merge (preserves user workflows)."""
        src = self.template_root / ".github"
        dst = self.target_path / ".github"

        if not src.exists():
            if not silent:
                console.print("⚠️ .github/ template not found")
            return

        # Smart merge: preserve existing user workflows
        if dst.exists():
            self._merge_github_workflows(src, dst)
        else:
            # First time: just copy
            self._copy_dir_with_substitution(src, dst)

        if not silent:
            console.print("   🔄 .github/ merged (user workflows preserved, variables substituted)")

    def _copy_claude_md(self, silent: bool = False) -> None:
        """Copy CLAUDE.md with complete replacement (no merge)."""
        src = self.template_root / "CLAUDE.md"
        dst = self.target_path / "CLAUDE.md"

        if not src.exists():
            if not silent:
                console.print("⚠️ CLAUDE.md template not found")
            return

        # Simple copy with substitution (no merge)
        self._copy_file_with_substitution(src, dst)

        if not silent:
            console.print("   ✅ CLAUDE.md replaced (use CLAUDE.local.md for personal instructions)")

    def _merge_claude_md(self, src: Path, dst: Path) -> None:
        """Delegate the smart merge for CLAUDE.md.

        Args:
            src: Template CLAUDE.md.
            dst: Project CLAUDE.md.
        """
        self.merger.merge_claude_md(src, dst)

    def _merge_github_workflows(self, src: Path, dst: Path) -> None:
        """Delegate the smart merge for .github/workflows/.

        Args:
            src: Template .github directory.
            dst: Project .github directory.
        """
        self.merger.merge_github_workflows(src, dst)

    def _merge_settings_json(self, src: Path, dst: Path) -> None:
        """Delegate the smart merge for settings.json.

        Args:
            src: Template settings.json.
            dst: Project settings.json.
        """
        # Find the latest backup for user settings extraction
        backup_path = None
        latest_backup = self.backup.get_latest_backup()
        if latest_backup:
            backup_settings = latest_backup / ".claude" / "settings.json"
            if backup_settings.exists():
                backup_path = backup_settings

        self.merger.merge_settings_json(src, dst, backup_path)

    def _merge_config_json(self, src: Path, dst: Path) -> None:
        """Smart merge for config.json using LanguageConfigResolver priority system.

        Args:
            src: Template config.json.
            dst: Project config.json.
        """
        import json

        # Load template config
        try:
            template_config = json.loads(src.read_text(encoding="utf-8"))
        except (json.JSONDecodeError, FileNotFoundError) as e:
            console.print(f"⚠️ Warning: Could not read template config.json: {e}")
            return

        # Find latest backup config.json
        latest_backup = self.backup.get_latest_backup()
        if latest_backup:
            backup_config_path = latest_backup / ".moai" / "config" / "config.json"
            if backup_config_path.exists():
                try:
                    json.loads(backup_config_path.read_text(encoding="utf-8"))
                except json.JSONDecodeError as e:
                    console.print(f"⚠️ Warning: Could not read backup config.json: {e}")

        # Load existing project config.json
        existing_config = {}
        if dst.exists():
            try:
                existing_config = json.loads(dst.read_text(encoding="utf-8"))
            except json.JSONDecodeError as e:
                console.print(f"⚠️ Warning: Could not read existing config.json: {e}")

        # Merge with priority system: Environment > Existing User > Template
        # We'll use LanguageConfigResolver to handle this properly
        try:
            # Import LanguageConfigResolver for priority-based merging
            from moai_adk.core.language_config_resolver import LanguageConfigResolver

            # Create temporary resolver to handle merging
            temp_project_path = self.target_path / ".moai" / "config"
            temp_project_path.mkdir(parents=True, exist_ok=True)

            # Start with template config as base
            merged_config = template_config.copy()

            # Apply existing user config (higher priority than template)
            for key, value in existing_config.items():
                if key not in ["config_source"]:  # Skip metadata
                    if key in merged_config and isinstance(merged_config[key], dict) and isinstance(value, dict):
                        # Deep merge for nested objects
                        merged_config[key].update(value)
                    else:
                        merged_config[key] = value

            # Apply environment variables (highest priority)
            import os

            env_mappings = {
                "MOAI_USER_NAME": ("user", "name"),
                "MOAI_CONVERSATION_LANG": ("language", "conversation_language"),
                "MOAI_AGENT_PROMPT_LANG": ("language", "agent_prompt_language"),
                "MOAI_CONVERSATION_LANG_NAME": (
                    "language",
                    "conversation_language_name",
                ),
                "MOAI_GIT_COMMIT_MESSAGES_LANG": ("language", "git_commit_messages"),
                "MOAI_CODE_COMMENTS_LANG": ("language", "code_comments"),
                "MOAI_DOCUMENTATION_LANG": ("language", "documentation"),
                "MOAI_ERROR_MESSAGES_LANG": ("language", "error_messages"),
            }

            for env_var, (section, key) in env_mappings.items():
                env_value = os.getenv(env_var)
                if env_value:
                    if section not in merged_config:
                        merged_config[section] = {}
                    merged_config[section][key] = env_value

            # Ensure consistency
            resolver = LanguageConfigResolver(str(self.target_path))
            merged_config = resolver._ensure_consistency(merged_config)

            # Write merged config
            dst.write_text(
                json.dumps(merged_config, indent=2, ensure_ascii=False) + "\n",
                encoding="utf-8",
            )

        except ImportError:
            # Fallback: simple merge without LanguageConfigResolver
            merged_config = template_config.copy()

            # Apply existing config
            for key, value in existing_config.items():
                if key not in ["config_source"]:
                    merged_config[key] = value

            dst.write_text(
                json.dumps(merged_config, indent=2, ensure_ascii=False) + "\n",
                encoding="utf-8",
            )
            console.print("   ⚠️ Warning: Using simple merge (LanguageConfigResolver not available)")

    def _copy_gitignore(self, silent: bool = False) -> None:
        """.gitignore copy (optional)."""
        src = self.template_root / ".gitignore"
        dst = self.target_path / ".gitignore"

        if not src.exists():
            return

        # Merge with the existing .gitignore when present
        if dst.exists():
            self._merge_gitignore(src, dst)
            if not silent:
                console.print("   🔄 .gitignore merged")
        else:
            shutil.copy2(src, dst)
            if not silent:
                console.print("   ✅ .gitignore copy complete")

    def _merge_gitignore(self, src: Path, dst: Path) -> None:
        """Delegate the .gitignore merge.

        Args:
            src: Template .gitignore.
            dst: Project .gitignore.
        """
        self.merger.merge_gitignore(src, dst)

    def _copy_mcp_json(self, silent: bool = False) -> None:
        """.mcp.json copy (smart merge with existing MCP server configuration)."""
        src = self.template_root / ".mcp.json"
        dst = self.target_path / ".mcp.json"

        if not src.exists():
            return

        # Merge with existing .mcp.json when present (preserve user-added MCP servers)
        if dst.exists():
            self._merge_mcp_json(src, dst)
            if not silent:
                console.print("   🔄 .mcp.json merged (user MCP servers preserved)")
        else:
            shutil.copy2(src, dst)
            if not silent:
                console.print("   ✅ .mcp.json copy complete")

    def _merge_mcp_json(self, src: Path, dst: Path) -> None:
        """Smart merge for .mcp.json (preserve user-added MCP servers).

        Args:
            src: Template .mcp.json.
            dst: Project .mcp.json.
        """
        try:
            src_data = json.loads(src.read_text(encoding="utf-8"))
            dst_data = json.loads(dst.read_text(encoding="utf-8"))

            # Merge mcpServers: preserve user servers, update template servers
            if "mcpServers" in src_data:
                if "mcpServers" not in dst_data:
                    dst_data["mcpServers"] = {}
                # Update with template servers (preserves existing user servers)
                dst_data["mcpServers"].update(src_data["mcpServers"])

            # Write merged result back
            dst.write_text(json.dumps(dst_data, indent=2, ensure_ascii=False), encoding="utf-8")
        except json.JSONDecodeError as e:
            console.print(f"[yellow]⚠️ Failed to merge .mcp.json: {e}[/yellow]")

    def merge_config(self, detected_language: str | None = None) -> dict[str, str]:
        """Delegate the smart merge for config.json.

        Args:
            detected_language: Detected language.

        Returns:
            Merged configuration dictionary.
        """
        return self.merger.merge_config(detected_language)
