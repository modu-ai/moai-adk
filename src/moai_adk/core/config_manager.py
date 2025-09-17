"""
Configuration management utilities for MoAI-ADK.

Handles creation and management of various configuration files including
Claude Code settings, MoAI config, package.json, and other project configs.
"""

import json
from datetime import datetime
from pathlib import Path
from typing import Dict, Any, List

from ..utils.logger import get_logger
from ..config import Config
from .security import SecurityManager, SecurityError
from .._version import __version__, get_version

logger = get_logger(__name__)


class ConfigManager:
    """Manages configuration files for MoAI-ADK installation."""

    def __init__(self, security_manager: SecurityManager = None):
        """
        Initialize configuration manager.

        Args:
            security_manager: Security manager instance for validation
        """
        self.security_manager = security_manager or SecurityManager()

    def create_claude_settings(self, settings_path: Path, config: Config) -> bool:
        """
        Create Claude Code settings.json file.

        Args:
            settings_path: Full path to settings.json file
            config: Project configuration

        Returns:
            bool: True if settings file was created successfully
        """
        if settings_path.exists() and not getattr(config, 'force_overwrite', False):
            # keep existing settings (from templates)
            return True

        settings = {
            "hooks": {
                "PreToolUse": [
                    {
                        "matcher": "Edit|MultiEdit|Write|Bash",
                        "hooks": [
                            {
                                "type": "command",
                                "command": "python3 $CLAUDE_PROJECT_DIR/.claude/hooks/moai/pre_write_guard.py",
                                "timeout": 60,
                                "description": "Sensitive path protection & risk guard"
                            }
                        ]
                    },
                    {
                        "matcher": "Bash|WebFetch",
                        "hooks": [
                            {
                                "type": "command",
                                "command": "python3 $CLAUDE_PROJECT_DIR/.claude/hooks/moai/policy_block.py",
                                "description": "Dangerous commands and policy blocking"
                            }
                        ]
                    },
                    {
                        "matcher": "Edit|MultiEdit|Write",
                        "hooks": [
                            {
                                "type": "command",
                                "command": "python3 $CLAUDE_PROJECT_DIR/.claude/hooks/moai/constitution_guard.py",
                                "description": "Constitution 5 principles validation"
                            },
                            {
                                "type": "command",
                                "command": "python3 $CLAUDE_PROJECT_DIR/.claude/hooks/moai/tag_validator.py",
                                "description": "16-Core @TAG validation"
                            }
                        ]
                    }
                ],
                "PostToolUse": [
                    {
                        "matcher": "Edit|MultiEdit|Write",
                        "hooks": [
                            {
                                "type": "command",
                                "command": "python3 $CLAUDE_PROJECT_DIR/.claude/hooks/moai/post_stage_guard.py",
                                "description": "Stage validation & doc sync"
                            }
                        ]
                    }
                ],
                "SessionStart": [
                    {
                        "matcher": "*",
                        "hooks": [
                            {
                                "type": "command",
                                "command": "python3 $CLAUDE_PROJECT_DIR/.claude/hooks/moai/session_start_notice.py",
                                "description": "MoAI-ADK session initialization and project status display"
                            }
                        ]
                    }
                ]
            },
            "permissions": {
                "defaultMode": "default",
                "allow": [
                    "Read(**)",
                    "Grep",
                    "Glob",
                    "Task",
                    "Bash",
                ]
            },
        }

        try:
            self._write_json_file(settings_path, settings)
            return True
        except SecurityError:
            logger.error("Security validation failed for config file: %s", settings_path)
            raise
        except Exception as e:
            logger.error("Failed to write Claude settings file: %s", e)
            return False

    def create_claude_settings_file(self, project_path: Path, config: Config) -> Path:
        """Backward-compatible helper to create settings file under project root."""
        settings_path = project_path / ".claude" / "settings.json"
        self.create_claude_settings(settings_path, config)
        return settings_path

    def create_moai_config(self, config_path: Path, config: Config) -> bool:
        """
        Create .moai/config.json configuration file.

        Args:
            config_path: Full path to config.json file
            config: Project configuration

        Returns:
            bool: True if config file was created successfully
        """
        if config_path.exists() and not getattr(config, 'force_overwrite', False):
            return True

        moai_config = {
            "version": get_version("moai_adk"),
            "created": datetime.now().isoformat(),
            "project": {
                "name": config.name,
                "type": config.project_type,
                "template": config.template,
                "runtime": config.runtime.name,
                "tech_stack": config.tech_stack,
            },
            "templates": {
                "mode": getattr(config, 'templates_mode', 'copy')
            },
            "constitution": {
                "simplicity": {"max_projects": 3, "enforce": True},
                "architecture": {"library_first": True, "enforce": True},
                "testing": {
                    "tdd_required": True,
                    "coverage_target": 0.8,
                    "enforce": True,
                },
                "observability": {"structured_logging": True, "enforce": True},
                "versioning": {"format": "MAJOR.MINOR.BUILD", "enforce": True},
            },
            "tags": {
                "version": "16-core",
                "categories": {
                    "spec": ["REQ", "SPEC", "DESIGN", "TASK"],
                    "steering": ["VISION", "STRUCT", "TECH", "ADR"],
                    "implementation": ["FEATURE", "API", "TEST", "DATA"],
                    "quality": ["PERF", "SEC", "DEBT", "TODO"],
                },
                "validation": {"strict_mode": True, "auto_index": True},
            },
            "pipeline": {
                "stages": ["SPECIFY", "PLAN", "TASKS", "IMPLEMENT"],
                "gates": {
                    "SPECIFY": ["ears_complete", "clarification_resolved"],
                    "PLAN": ["constitution_check", "adr_generated"],
                    "TASKS": ["tdd_order", "dependency_graph"],
                    "IMPLEMENT": ["red_green_refactor", "coverage_target"],
                },
            },
            "agents": {
                "count": 11,
                "enabled": [
                    "claude-code-manager", "steering-architect", "spec-manager",
                    "plan-architect", "task-decomposer", "code-generator",
                    "test-automator", "tag-indexer", "doc-syncer",
                    "quality-auditor", "integration-manager", "deployment-specialist",
                ],
            },
        }

        try:
            self._write_json_file(config_path, moai_config)
            return True
        except SecurityError:
            logger.error("Security validation failed for config file: %s", config_path)
            raise
        except Exception as e:
            logger.error("Failed to write MoAI config file: %s", e)
            return False

    def create_moai_config_file(self, project_path: Path, config: Config) -> Path:
        """Backward-compatible helper to create config file under project root."""
        config_path = project_path / ".moai" / "config.json"
        self.create_moai_config(config_path, config)
        return config_path

    def create_package_json(self, project_path: Path, config: Config) -> Path:
        """
        Create package.json for Node.js projects.

        Args:
            project_path: Project root path
            config: Project configuration

        Returns:
            Path: Path to created package.json
        """
        package_data = {
            "name": config.name,
            "version": get_version("moai_adk"),
            "type": "module",
            "private": True,
            "scripts": {
                "dev": (
                    "next dev"
                    if "nextjs" in config.tech_stack
                    else "npm run start"
                ),
                "build": (
                    "next build"
                    if "nextjs" in config.tech_stack
                    else "echo 'No build script'"
                ),
                "start": (
                    "next start"
                    if "nextjs" in config.tech_stack
                    else "node index.js"
                ),
                "lint": (
                    "next lint" if "nextjs" in config.tech_stack else "eslint ."
                ),
            },
        }

        package_path = project_path / "package.json"
        return self._write_json_file(package_path, package_data)

    def create_initial_indexes(self, project_path: Path, config: Config) -> List[Path]:
        """
        Create initial 16-Core TAG index files.

        Args:
            project_path: Project root path
            config: Project configuration

        Returns:
            List[Path]: List of created index files
        """
        index_files = []
        indexes_dir = project_path / ".moai" / "indexes"
        indexes_dir.mkdir(parents=True, exist_ok=True)

        # 1. tags.json - 16-Core TAG system initialization
        tags_data = {
            "metadata": {
                "version": "16-core",
                "generated_at": datetime.now().isoformat(),
                "total_tags": 0,
                "categories": {
                    "SPEC": 0, "Steering": 0, "Implementation": 0,
                    "Quality": 0, "Legacy": 0,
                },
            },
            "tags": {},
            "categories": {
                "SPEC": {"REQ": [], "SPEC": [], "DESIGN": [], "TASK": []},
                "Steering": {"VISION": [], "STRUCT": [], "TECH": [], "ADR": []},
                "Implementation": {"FEATURE": [], "API": [], "TEST": [], "DATA": []},
                "Quality": {"PERF": [], "SEC": [], "DEBT": [], "TODO": []},
                "Legacy": {
                    "SPEC": [], "ADR": [], "US": [], "FR": [], "NFR": [],
                    "BUG": [], "REVIEW": [],
                },
            },
        }

        tags_file = indexes_dir / "tags.json"
        index_files.append(self._write_json_file(tags_file, tags_data))

        # 2. traceability.json - 4 chains definition
        traceability_data = {
            "metadata": {
                "version": "16-core",
                "generated_at": datetime.now().isoformat(),
                "total_links": 0,
            },
            "chains": {
                "primary": ["REQ", "DESIGN", "TASK", "TEST"],
                "steering": ["VISION", "STRUCT", "TECH", "ADR"],
                "implementation": ["FEATURE", "API", "DATA"],
                "quality": ["PERF", "SEC", "DEBT", "TODO"],
            },
            "links": [],
        }

        traceability_file = indexes_dir / "traceability.json"
        index_files.append(self._write_json_file(traceability_file, traceability_data))

        # 3. state.json - Project state management
        state_data = {
            "project_name": config.name,
            "current_stage": "INIT",
            "pipeline_status": {
                "SPECIFY": "pending", "PLAN": "pending",
                "TASKS": "pending", "IMPLEMENT": "pending",
            },
            "constitution_score": {
                "simplicity": 1.0, "architecture": 1.0, "testing": 0.0,
                "observability": 0.0, "versioning": 1.0,
            },
            "created_at": datetime.now().isoformat(),
            "updated_at": datetime.now().isoformat(),
        }

        state_file = indexes_dir / "state.json"
        index_files.append(self._write_json_file(state_file, state_data))

        return index_files

    def setup_steering_config(self, project_path: Path) -> Path:
        """
        Setup MoAI steering configuration with Constitution 5 principles.

        Args:
            project_path: Project root path

        Returns:
            Path: Path to created steering config
        """
        steering_config = {
            "autoCommit": "ask",
            "autoSync": "ask",
            "coverageTarget": 0.8,
            "testRunner": "pytest",
            "lintCommand": "flake8",
            "typecheckCommand": "mypy",
            "strictRegression": False,
            "enforceSessionBoundary": False,
            "denyExternalTools": False,
            "constitution": {
                "maxProjects": 3,
                "requireLibraries": True,
                "enforceTDD": True,
                "requireObservability": True,
                "versioningScheme": "MAJOR.MINOR.BUILD",
            },
            "pipeline": {
                "gates": {
                    "SPECIFY": {
                        "requireEARS": True,
                        "resolveClarifications": True,
                        "validateUserStories": True,
                    },
                    "PLAN": {
                        "constitutionCheck": True,
                        "requireADR": True,
                        "techResearch": True,
                    },
                    "TASKS": {
                        "enforceTDD": True,
                        "validateParallel": True,
                        "dependencyCheck": True,
                    },
                    "IMPLEMENT": {
                        "redGreenRefactor": True,
                        "coverageThreshold": 0.8,
                        "qualityGates": True,
                    },
                }
            },
        }

        config_file = project_path / ".moai" / "config.json"
        return self._write_json_file(config_file, steering_config)

    def _write_json_file(self, file_path: Path, data: Dict[str, Any]) -> Path:
        """
        Write data to JSON file with security validation.

        Args:
            file_path: Path to write JSON file
            data: Data to write

        Returns:
            Path: Path to written file

        Raises:
            SecurityError: If security validation fails
        """
        try:
            # Security validation
            if not self.security_manager.validate_file_creation(
                file_path, file_path.parent.parent
            ):
                logger.error("Security validation failed for config file: %s", file_path)
                raise SecurityError(f"Security validation failed: {file_path}")

            # Ensure parent directory exists
            file_path.parent.mkdir(parents=True, exist_ok=True)

            # Write JSON file
            with open(file_path, "w", encoding="utf-8") as f:
                json.dump(data, f, indent=2, ensure_ascii=False)

            logger.debug("Created config file: %s", file_path)
            return file_path

        except SecurityError:
            # Re-raise security errors
            raise
        except Exception as e:
            logger.error("Failed to write config file %s: %s", file_path, e)
            raise

    def validate_config_file(self, file_path: Path) -> bool:
        """
        Validate that a configuration file is valid JSON.

        Args:
            file_path: Path to config file to validate

        Returns:
            bool: True if valid
        """
        try:
            if not file_path.exists():
                logger.error("Config file does not exist: %s", file_path)
                return False

            with open(file_path, "r", encoding="utf-8") as f:
                json.load(f)

            logger.debug("Config file validation passed: %s", file_path)
            return True

        except json.JSONDecodeError as e:
            logger.error("Config file JSON validation failed %s: %s", file_path, e)
            return False
        except Exception as e:
            logger.error("Config file validation error %s: %s", file_path, e)
            return False

    def backup_config_file(self, file_path: Path) -> Path:
        """
        Create a backup of a configuration file.

        Args:
            file_path: Path to config file to backup

        Returns:
            Path: Path to backup file
        """
        import shutil
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_path = file_path.with_suffix(f".backup_{timestamp}{file_path.suffix}")

        try:
            shutil.copy2(file_path, backup_path)
            logger.info("Created config backup: %s -> %s", file_path, backup_path)
            return backup_path

        except Exception as e:
            logger.error("Failed to backup config file %s: %s", file_path, e)
            raise
