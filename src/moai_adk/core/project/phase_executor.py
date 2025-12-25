# type: ignore
"""Phase-based installation executor (SPEC-INIT-003 v0.4.2)

Runs the project initialization across five phases:
- Phase 1: Preparation (create single backup at .moai-backups/backup/)
- Phase 2: Directory (build directory structure)
- Phase 3: Resource (copy templates while preserving user content)
- Phase 4: Configuration (generate configuration files)
- Phase 5: Validation (verify and finalize)

Test coverage includes 5-phase integration tests with backup, configuration, and validation
"""

import json
import logging
import platform
import shutil
import subprocess
from collections.abc import Callable
from datetime import datetime
from pathlib import Path
from typing import Any

from rich.console import Console

from moai_adk import __version__
from moai_adk.core.project.backup_utils import (
    get_backup_targets,
    has_any_moai_files,
    is_protected_path,
)
from moai_adk.core.project.validator import ProjectValidator
from moai_adk.core.template.processor import TemplateProcessor
from moai_adk.statusline.version_reader import VersionConfig, VersionReader

console = Console()

# Progress callback type alias
ProgressCallback = Callable[[str, int, int], None]


class PhaseExecutor:
    """Execute the installation across the five phases.

    Phases:
    1. Preparation: Back up and verify the system.
    2. Directory: Create the directory structure.
    3. Resource: Copy template resources.
    4. Configuration: Generate configuration files.
    5. Validation: Perform final checks.

    Enhanced with improved version reading and context management.
    """

    # Required directory structure
    REQUIRED_DIRECTORIES = [
        ".moai/",
        ".moai/project/",
        ".moai/specs/",
        ".moai/reports/",
        ".moai/memory/",
        ".claude/",
        ".claude/logs/",
        ".github/",
    ]

    def __init__(self, validator: ProjectValidator) -> None:
        """Initialize the executor.

        Args:
            validator: Project validation helper.
        """
        self.validator = validator
        self.total_phases = 5
        self.current_phase = 0
        self._version_reader: VersionReader | None = None

    def _get_version_reader(self) -> VersionReader:
        """
        Get or create version reader instance.

        Returns:
            VersionReader instance with enhanced configuration
        """
        if self._version_reader is None:
            config = VersionConfig(
                cache_ttl_seconds=120,  # Longer cache for phase execution
                fallback_version=__version__,
                version_format_regex=r"^v?(\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?)$",
                cache_enabled=True,
                debug_mode=False,
            )
            self._version_reader = VersionReader(config)
        return self._version_reader

    def _get_enhanced_version_context(self) -> dict[str, str]:
        """
        Get enhanced version context with fallback strategies and comprehensive configuration.

        Returns:
            Dictionary containing version-related template variables with enhanced formatting
        """
        version_context = {}
        logger = logging.getLogger(__name__)

        try:
            version_reader = self._get_version_reader()
            moai_version = version_reader.get_version()

            # Enhanced version context with multiple format options
            version_context["MOAI_VERSION"] = moai_version
            version_context["MOAI_VERSION_SHORT"] = self._format_short_version(moai_version)
            version_context["MOAI_VERSION_DISPLAY"] = self._format_display_version(moai_version)
            version_context["MOAI_VERSION_TRIMMED"] = self._format_trimmed_version(moai_version, max_length=10)
            version_context["MOAI_VERSION_SEMVER"] = self._format_semver_version(moai_version)
            version_context["MOAI_VERSION_VALID"] = "true" if moai_version != "unknown" else "false"
            version_context["MOAI_VERSION_SOURCE"] = self._get_version_source(version_reader)

            # Add performance metrics for debugging
            cache_age = version_reader.get_cache_age_seconds()
            if cache_age is not None:
                version_context["MOAI_VERSION_CACHE_AGE"] = f"{cache_age:.2f}s"
            else:
                version_context["MOAI_VERSION_CACHE_AGE"] = "uncached"

        except Exception as e:
            logger.warning(f"Failed to read version for context: {e}")
            # Use fallback version with comprehensive fallback formatting
            fallback_version = __version__
            version_context["MOAI_VERSION"] = fallback_version
            version_context["MOAI_VERSION_SHORT"] = self._format_short_version(fallback_version)
            version_context["MOAI_VERSION_DISPLAY"] = self._format_display_version(fallback_version)
            version_context["MOAI_VERSION_TRIMMED"] = self._format_trimmed_version(fallback_version, max_length=10)
            version_context["MOAI_VERSION_SEMVER"] = self._format_semver_version(fallback_version)
            version_context["MOAI_VERSION_VALID"] = "true"
            version_context["MOAI_VERSION_SOURCE"] = "fallback_package"
            version_context["MOAI_VERSION_CACHE_AGE"] = "unavailable"

        return version_context

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

        # Check if we have a cached version (most likely from config)
        cache_age = version_reader.get_cache_age_seconds()
        if cache_age is not None and cache_age < config.cache_ttl_seconds:
            return "config_cached"
        elif cache_age is not None:
            return "config_stale"
        else:
            return config.fallback_version

    def execute_preparation_phase(
        self,
        project_path: Path,
        backup_enabled: bool = True,
        progress_callback: ProgressCallback | None = None,
    ) -> None:
        """Phase 1: preparation and backup.

        Args:
            project_path: Project path.
            backup_enabled: Whether backups are enabled.
            progress_callback: Optional progress callback.
        """
        self.current_phase = 1
        self._report_progress("Phase 1: Preparation and backup...", progress_callback)

        # Validate system requirements
        self.validator.validate_system_requirements()

        # Verify the project path
        self.validator.validate_project_path(project_path)

        # Create a backup when needed
        if backup_enabled and has_any_moai_files(project_path):
            self._create_backup(project_path)

    def execute_directory_phase(
        self,
        project_path: Path,
        progress_callback: ProgressCallback | None = None,
    ) -> None:
        """Phase 2: create directories.

        Args:
            project_path: Project path.
            progress_callback: Optional progress callback.
        """
        self.current_phase = 2
        self._report_progress("Phase 2: Creating directory structure...", progress_callback)

        for directory in self.REQUIRED_DIRECTORIES:
            dir_path = project_path / directory
            dir_path.mkdir(parents=True, exist_ok=True)

    def execute_resource_phase(
        self,
        project_path: Path,
        config: dict[str, str] | None = None,
        progress_callback: ProgressCallback | None = None,
    ) -> list[str]:
        """Phase 3: install resources with variable substitution.

        Args:
            project_path: Project path.
            config: Configuration dictionary for template variable substitution.
            progress_callback: Optional progress callback.

        Returns:
            List of created files or directories.
        """
        import stat

        self.current_phase = 3
        self._report_progress("Phase 3: Installing resources...", progress_callback)

        # Copy resources via TemplateProcessor in silent mode
        processor = TemplateProcessor(project_path)

        # Set template variable context (if provided)
        if config:
            # Get language settings from 'language_settings' key (dict)
            # Falls back to 'language' key for backwards compatibility
            language_config: dict[str, Any] = config.get("language_settings") or {}
            if not language_config or not isinstance(language_config, dict):
                # Backwards compatibility: try 'language' key as dict
                legacy_lang = config.get("language", {})
                language_config = legacy_lang if isinstance(legacy_lang, dict) else {}

            # Detect OS for cross-platform Hook path configuration
            hook_project_dir = "%CLAUDE_PROJECT_DIR%" if platform.system() == "Windows" else "$CLAUDE_PROJECT_DIR"

            # Detect OS for cross-platform statusline command
            # Windows: Use python -m for better PATH compatibility
            # Unix: Use moai-adk directly (assumes installed via uv tool)
            if platform.system() == "Windows":
                statusline_command = "python -m moai_adk statusline"
            else:
                statusline_command = "moai-adk statusline"

            # Get enhanced version context with fallback strategies
            version_context = self._get_enhanced_version_context()

            context = {
                **version_context,
                "CREATION_TIMESTAMP": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
                "PROJECT_NAME": config.get("name", "unknown"),
                "PROJECT_DESCRIPTION": config.get("description", ""),
                "PROJECT_MODE": config.get("mode", "personal"),
                "PROJECT_VERSION": config.get("version", "0.1.0"),
                "PROJECT_OWNER": config.get("author", "@user"),
                "AUTHOR": config.get("author", "@user"),
                "CONVERSATION_LANGUAGE": language_config.get("conversation_language", "en"),
                "CONVERSATION_LANGUAGE_NAME": language_config.get("conversation_language_name", "English"),
                "CODEBASE_LANGUAGE": config.get("language", "generic"),
                "PROJECT_DIR": hook_project_dir,
                "STATUSLINE_COMMAND": statusline_command,
            }
            processor.set_context(context)

        processor.copy_templates(backup=False, silent=True)  # Avoid progress bar conflicts

        # Post-process: Set executable permission on shell scripts
        # This is necessary because git may not preserve file permissions during clone/checkout
        scripts_dir = project_path / ".moai" / "scripts"
        logger = logging.getLogger(__name__)
        if scripts_dir.exists():
            logger.debug(f"Processing shell scripts in {scripts_dir}")
            for script_file in scripts_dir.glob("*.sh"):
                try:
                    # Add execute permission for user, group, and others
                    current_mode = script_file.stat().st_mode
                    new_mode = current_mode | stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH
                    script_file.chmod(new_mode)
                    logger.debug(f"Set executable permission on {script_file}: {oct(current_mode)} -> {oct(new_mode)}")
                except Exception as e:
                    logger.warning(f"Failed to set executable permission on {script_file}: {e}")
        else:
            logger.debug(f"Scripts directory not found: {scripts_dir}")

        # Return a simplified list of generated assets
        return [
            ".claude/",
            ".moai/",
            ".github/",
            "CLAUDE.md",
            ".gitignore",
        ]

    def execute_configuration_phase(
        self,
        project_path: Path,
        config: dict[str, str | bool | dict[Any, Any]],
        progress_callback: ProgressCallback | None = None,
    ) -> list[str]:
        """Phase 4: generate configuration.

        Args:
            project_path: Project path.
            config: Configuration dictionary.
            progress_callback: Optional progress callback.

        Returns:
            List of created files.
        """
        self.current_phase = 4
        self._report_progress("Phase 4: Generating configurations...", progress_callback)

        logger = logging.getLogger(__name__)

        # Read existing config to preserve user settings (Issue #165)
        config_path = project_path / ".moai" / "config" / "config.json"
        existing_config: dict[str, Any] = {}
        if config_path.exists():
            try:
                with open(config_path, "r", encoding="utf-8") as f:
                    existing_config = json.load(f)
                logger.debug(f"Successfully read existing config from {config_path}")
            except (json.JSONDecodeError, OSError) as e:
                logger.warning(f"Failed to read existing config: {e}. Starting fresh.")
                existing_config = {}

        # Enhanced config merging with comprehensive version preservation
        merged_config = self._merge_configuration_preserving_versions(config, existing_config)

        # Enhanced version handling using VersionReader for consistency
        try:
            version_reader = self._get_version_reader()
            current_config_version = version_reader.get_version()

            # Ensure version consistency across the merged config
            self._ensure_version_consistency(merged_config, current_config_version, existing_config)

            logger.debug(f"Version consistency check completed. Current version: {current_config_version}")
        except Exception as e:
            logger.warning(f"Version consistency check failed: {e}. Using fallback version.")
            merged_config["moai"]["version"] = __version__

        # Write final config with enhanced formatting
        self._write_configuration_file(config_path, merged_config)
        logger.info(f"Configuration file written to {config_path}")

        return [str(config_path)]

    def _merge_configuration_preserving_versions(
        self, new_config: dict[str, Any], existing_config: dict[str, Any]
    ) -> dict[str, Any]:
        """
        Merge configurations while preserving user settings and version information.

        Args:
            new_config: New configuration from initialization
            existing_config: Existing configuration from project

        Returns:
            Merged configuration dictionary
        """
        logger = logging.getLogger(__name__)
        merged_config = new_config.copy()

        # Define configuration sections with their merge strategies
        config_sections = {
            "moai": {"preserve_all": True, "priority": "user"},
            "user": {"preserve_keys": ["nickname"], "priority": "user"},
            "language": {
                "preserve_keys": [],
                "priority": "new",
            },  # Use new language config during init
            "project": {"preserve_keys": [], "priority": "new"},
            "git": {"preserve_keys": [], "priority": "new"},
        }

        for section_name, strategy in config_sections.items():
            if section_name in existing_config:
                logger.debug(f"Merging section: {section_name}")
                self._merge_config_section(merged_config, existing_config, section_name, strategy)

        return merged_config

    def _merge_config_section(
        self,
        merged_config: dict[str, Any],
        existing_config: dict[str, Any],
        section_name: str,
        strategy: dict[str, Any],
    ) -> None:
        """
        Merge a specific configuration section.

        Args:
            merged_config: Target configuration to merge into
            existing_config: Source configuration to merge from
            section_name: Name of the section to merge
            strategy: Merge strategy for this section
        """
        logger = logging.getLogger(__name__)
        if section_name not in merged_config:
            merged_config[section_name] = {}

        section_config = merged_config[section_name]
        existing_section = existing_config[section_name]

        if strategy["priority"] == "user":
            # User priority: preserve existing values
            preserve_keys = strategy.get("preserve_keys", [])
            # Convert frozenset to list if needed
            if isinstance(preserve_keys, frozenset):
                preserve_keys = list(preserve_keys)
            elif not isinstance(preserve_keys, list):
                preserve_keys = list(preserve_keys) if preserve_keys else []

            for key, value in existing_section.items():
                if strategy.get("preserve_all", False) or key in preserve_keys:
                    section_config[key] = value
                    logger.debug(f"Preserved {section_name}.{key} = {value}")
        else:
            # New priority: keep new config, but don't overwrite if exists
            for key, value in existing_section.items():
                if key not in section_config:
                    section_config[key] = value
                    logger.debug(f"Inherited {section_name}.{key} = {value}")

    def _ensure_version_consistency(
        self,
        config: dict[str, Any],
        current_version: str,
        existing_config: dict[str, Any],
    ) -> None:
        """
        Ensure version consistency across the configuration.

        Args:
            config: Configuration to update
            current_version: Current version from VersionReader
            existing_config: Existing configuration for reference
        """
        logger = logging.getLogger(__name__)

        # Ensure moai section exists
        if "moai" not in config:
            config["moai"] = {}

        # Version field priority strategy:
        # 1. User explicitly set in existing config -> preserve
        # 2. Version from config file -> use
        # 3. Current version from VersionReader -> use
        # 4. Package version -> fallback

        existing_moai = existing_config.get("moai", {})
        config_moai = config["moai"]

        # Check if user explicitly set a version in existing config
        if "version" in existing_moai:
            user_version = existing_moai["version"]
            logger.debug(f"User explicitly set version: {user_version}")
            config_moai["version"] = user_version
        elif "version" in config_moai:
            # Version already in new config, validate it
            config_version = config_moai["version"]
            if config_version == "unknown" or not self._is_valid_version_format(config_version):
                logger.debug(f"Invalid config version {config_version}, updating to current: {current_version}")
                config_moai["version"] = current_version
        else:
            # No version found, use current version
            logger.debug(f"No version found, setting to current: {current_version}")
            config_moai["version"] = current_version

    def _is_valid_version_format(self, version: str) -> bool:
        """
        Check if version format is valid.

        Args:
            version: Version string to validate

        Returns:
            True if version format is valid
        """
        import re

        pattern = r"^v?(\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?)$"
        return bool(re.match(pattern, version))

    def _write_configuration_file(self, config_path: Path, config: dict[str, Any]) -> None:
        """
        Write configuration file with enhanced formatting and error handling.

        Args:
            config_path: Path to write configuration file
            config: Configuration dictionary to write
        """
        logger = logging.getLogger(__name__)

        try:
            # Ensure parent directory exists
            config_path.parent.mkdir(parents=True, exist_ok=True)

            # Write with enhanced formatting
            with open(config_path, "w", encoding="utf-8") as f:
                json.dump(config, f, indent=2, ensure_ascii=False)

            logger.info(f"Configuration successfully written to {config_path}")

        except Exception as e:
            logger.error(f"Failed to write configuration file: {e}")
            raise

    def execute_validation_phase(
        self,
        project_path: Path,
        mode: str = "personal",
        progress_callback: ProgressCallback | None = None,
    ) -> None:
        """Phase 5: validation and wrap-up.


        Args:
            project_path: Project path.
            mode: Project mode (personal/team).
            progress_callback: Optional progress callback.
        """
        self.current_phase = 5
        self._report_progress("Phase 5: Validation and finalization...", progress_callback)

        # Validate installation results
        # Comprehensive installation validation
        # Verifies all required files including 4 Alfred command files:
        # - 0-project.md, 1-plan.md, 2-run.md, 3-sync.md
        self.validator.validate_installation(project_path)

        # Initialize Git for all modes (team, personal, solo)
        self._initialize_git(project_path)

    def _create_backup(self, project_path: Path) -> None:
        """Create a single backup (v0.4.2).

        Maintains only one backup at .moai-backups/backup/.

        Args:
            project_path: Project path.
        """
        # Define backup directory
        backups_dir = project_path / ".moai-backups"
        backup_path = backups_dir / "backup"

        # Remove existing backup if present
        if backup_path.exists():
            shutil.rmtree(backup_path)

        # Create backup directories
        backups_dir.mkdir(parents=True, exist_ok=True)
        backup_path.mkdir(parents=True, exist_ok=True)

        # Collect backup targets
        targets = get_backup_targets(project_path)
        backed_up_files: list[str] = []

        # Execute the backup
        for target in targets:
            src_path = project_path / target
            dst_path = backup_path / target

            if src_path.is_dir():
                self._copy_directory_selective(src_path, dst_path)
                backed_up_files.append(f"{target}/")
            else:
                dst_path.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(src_path, dst_path)
                backed_up_files.append(target)

        # Avoid additional console messages to prevent progress bar conflicts

    def _copy_directory_selective(self, src: Path, dst: Path) -> None:
        """Copy a directory while skipping protected paths.

        Args:
            src: Source directory.
            dst: Destination directory.
        """
        dst.mkdir(parents=True, exist_ok=True)

        for item in src.rglob("*"):
            rel_path = item.relative_to(src)

            # Skip protected paths
            if is_protected_path(rel_path):
                continue

            dst_item = dst / rel_path
            if item.is_file():
                dst_item.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(item, dst_item)
            elif item.is_dir():
                dst_item.mkdir(parents=True, exist_ok=True)

    def _initialize_git(self, project_path: Path) -> None:
        """Initialize a Git repository if not already initialized.

        Args:
            project_path: Project path.
        """
        # Check if .git directory already exists
        git_dir = project_path / ".git"
        if git_dir.exists() and git_dir.is_dir():
            # Git already initialized, skip
            return

        try:
            subprocess.run(
                ["git", "init"],
                cwd=project_path,
                check=True,
                capture_output=True,
                timeout=30,  # Default timeout for git operations
            )
            # Intentionally avoid printing to keep progress output clean
        except subprocess.TimeoutExpired:
            # Timeout is non-fatal
            pass
        except subprocess.CalledProcessError:
            # Only log on error; failures are non-fatal
            pass

    def _report_progress(self, message: str, callback: ProgressCallback | None) -> None:
        """Report progress.

        Args:
            message: Progress message.
            callback: Callback function.
        """
        if callback:
            callback(message, self.current_phase, self.total_phases)
