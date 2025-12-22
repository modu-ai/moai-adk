# SPEC: SPEC-CORE-PROJECT-001.md, SPEC-INIT-003.md
# TEST: tests/unit/test_project_initializer.py, tests/unit/test_init_reinit.py
"""Project Initialization Module

Phase-based 5-step initialization process:
1. Preparation: Backup and validation
2. Directory: Create directory structure
3. Resource: Copy template resources
4. Configuration: Generate configuration files
5. Validation: Verification and finalization
"""
# type: ignore

import json
import stat
import time
from datetime import datetime
from pathlib import Path

from moai_adk.core.project.phase_executor import PhaseExecutor, ProgressCallback  # type: ignore
from moai_adk.core.project.validator import ProjectValidator


class InstallationResult:
    """Installation result"""

    def __init__(
        self,
        success: bool,
        project_path: str,
        language: str,
        mode: str,
        locale: str,
        duration: int,
        created_files: list[str],
        errors: list[str] | None = None,
    ) -> None:
        self.success = success
        self.project_path = project_path
        self.language = language
        self.mode = mode
        self.locale = locale
        self.duration = duration
        self.created_files = created_files
        self.errors = errors or []


class ProjectInitializer:
    """Project initializer (Phase-based)"""

    def __init__(self, path: str | Path = ".") -> None:
        """Initialize

        Args:
            path: Project root directory
        """
        self.path = Path(path).resolve()
        self.validator = ProjectValidator()
        self.executor = PhaseExecutor(self.validator)

    def _create_memory_files(self) -> list[str]:
        """Create runtime session and memory files (auto-generated per user/session)

        Returns:
            List of created memory files

        """
        memory_dir = self.path / ".moai" / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)
        created_files = []

        # 1. project-notes.json - Project tracking notes (empty on init)
        project_notes = {
            "tech_debt": [],
            "performance_bottlenecks": [],
            "recent_patterns": {
                "frequent_file_edits": [],
                "test_failures": [],
                "git_operations": "daily commits, feature branches",
            },
            "next_priorities": [],
        }
        project_notes_file = memory_dir / "project-notes.json"
        project_notes_file.write_text(json.dumps(project_notes, indent=2))
        created_files.append(str(project_notes_file))

        # 2. session-hint.json - Last session state
        session_hint = {
            "last_command": None,
            "command_timestamp": None,
            "hours_ago": None,
            "active_spec": None,
            "current_branch": "main",
        }
        session_hint_file = memory_dir / "session-hint.json"
        session_hint_file.write_text(json.dumps(session_hint, indent=2))
        created_files.append(str(session_hint_file))

        # 3. user-patterns.json - User preferences and expertise
        user_patterns = {
            "tech_preferences": {},
            "expertise_signals": {
                "ask_question_skip_rate": 0.0,
                "custom_workflows": 0,
                "estimated_level": "beginner",
            },
            "skip_questions": [],
            "last_updated": datetime.now().isoformat() + "Z",
        }
        user_patterns_file = memory_dir / "user-patterns.json"
        user_patterns_file.write_text(json.dumps(user_patterns, indent=2))
        created_files.append(str(user_patterns_file))

        return created_files

    def _create_user_settings(self) -> list[str]:
        """Create user-specific settings files (.claude/settings.local.json)

        Returns:
            List of created settings files
        """
        created_files = []
        claude_dir = self.path / ".claude"
        claude_dir.mkdir(parents=True, exist_ok=True)

        # Create settings.local.json
        settings_local = {
            "_meta": {
                "description": "User-specific Claude Code settings (gitignored - never commit)",
                "created_at": datetime.now().isoformat() + "Z",
                "note": "Edit this file to customize your local development environment",
            },
            "enabledMcpjsonServers": ["context7"],  # context7 is mandatory
        }

        # Add companyAnnouncements in user's selected language
        try:
            import sys

            utils_dir = (
                Path(__file__).parent.parent.parent / "templates" / ".claude" / "hooks" / "moai" / "shared" / "utils"
            )

            if utils_dir.exists():
                sys.path.insert(0, str(utils_dir))
                try:
                    from announcement_translator import (
                        get_language_from_config,
                        translate_announcements,
                    )

                    language = get_language_from_config(self.path)
                    announcements = translate_announcements(language, self.path)
                    settings_local["companyAnnouncements"] = announcements
                except Exception as e:
                    print(f"[ProjectInitializer] Warning: Failed to add announcements: {e}")
                finally:
                    sys.path.remove(str(utils_dir))

        except Exception as e:
            print(f"[ProjectInitializer] Warning: Announcement module not available: {e}")

        settings_local_file = claude_dir / "settings.local.json"
        settings_local_file.write_text(json.dumps(settings_local, indent=2, ensure_ascii=False))
        created_files.append(str(settings_local_file))

        return created_files

    def initialize(
        self,
        mode: str = "personal",
        locale: str = "en",  # Changed from "ko" to "en" (will be configurable in /moai:0-project)
        language: str | None = None,
        custom_language: str | None = None,
        backup_enabled: bool = True,
        progress_callback: ProgressCallback | None = None,
        reinit: bool = False,
    ) -> InstallationResult:
        """Execute project initialization (5-phase process)

        Args:
            mode: Project mode (personal/team) - Default: personal (configurable in /moai:0-project)
            locale: Locale (ko/en/ja/zh/other) - Default: en (configurable in /moai:0-project)
            language: Force language specification (auto-detect if None) - Will be detected in /moai:0-project
            custom_language: Custom language name when locale="other" (user input)
            backup_enabled: Whether to enable backup
            progress_callback: Progress callback
            reinit: Reinitialization mode (v0.3.0, SPEC-INIT-003)

        Returns:
            InstallationResult object

        Raises:
            FileExistsError: If project is already initialized (when reinit=False)
        """
        start_time = time.time()

        try:
            # Prevent duplicate initialization (only when not in reinit mode)
            if self.is_initialized() and not reinit:
                raise FileExistsError(
                    f"Project already initialized at {self.path}/.moai/\n"
                    f"Use 'python -m moai_adk status' to check the current configuration."
                )

            # Use provided language or default to generic
            # Language detection now happens in /moai:0-project via project-manager
            detected_language = language or "generic"

            # Phase 1: Preparation (backup and validation)
            self.executor.execute_preparation_phase(self.path, backup_enabled, progress_callback)

            # Phase 2: Directory (create directories)
            self.executor.execute_directory_phase(self.path, progress_callback)

            # Build language_config BEFORE Phase 3 for template variable substitution
            # This fixes the bug where language selection was not applied to section files
            language_names = {
                "ko": "Korean",
                "en": "English",
                "ja": "Japanese",
                "zh": "Chinese",
            }
            if locale == "other" and custom_language:
                language_config = {
                    "conversation_language": "other",
                    "conversation_language_name": custom_language,
                }
            elif locale in language_names:
                language_config = {
                    "conversation_language": locale,
                    "conversation_language_name": language_names[locale],
                }
            else:
                # Default fallback
                language_config = {
                    "conversation_language": locale or "en",
                    "conversation_language_name": language_names.get(locale, "English"),
                }

            # Prepare config for template variable substitution (Phase 3)
            config = {
                "name": self.path.name,
                "mode": mode,
                "locale": locale,
                "language": detected_language,  # Programming language (string)
                "language_settings": language_config,  # Language settings (dict)
                "description": "",
                "version": "0.1.0",
                "author": "@user",
            }

            # Phase 3: Resource (copy templates with variable substitution)
            resource_files = self.executor.execute_resource_phase(self.path, config, progress_callback)

            # Post-Phase 3: Fix shell script permissions
            # git may not preserve file permissions, so explicitly set them
            scripts_dir = self.path / ".moai" / "scripts"
            if scripts_dir.exists():
                for script_file in scripts_dir.glob("*.sh"):
                    try:
                        current_mode = script_file.stat().st_mode
                        new_mode = current_mode | stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH
                        # On Windows, chmod has limited effect, but we try anyway
                        # or check os.name != 'nt' if strict behavior is needed.
                        # For now, we just apply it and ignore errors if it fails.
                        script_file.chmod(new_mode)
                    except Exception:
                        pass  # Silently ignore permission errors

            # Phase 4: Configuration (generate config.json)
            # Note: language_config is already created before Phase 3
            config_data: dict[str, str | bool | dict] = {
                "project": {
                    "name": self.path.name,
                    "mode": mode,
                    "locale": locale,
                    "language": detected_language,
                    # Language detection metadata (will be updated by project-manager via /moai:0-project)
                    "language_detection": {
                        "detected_language": detected_language,
                        "detection_method": "cli_default",  # Will be "context_aware" after /moai:0-project
                        "confidence": None,  # Will be calculated by project-manager
                        "markers": [],  # Will be populated by project-manager
                        "confirmed_by": None,  # Will be "user" after project-manager confirmation
                    },
                },
                "language": language_config,
                "constitution": {
                    "enforce_tdd": True,
                    "test_coverage_target": 85,
                },
                "git_strategy": {
                    "personal": {
                        "auto_checkpoint": "disabled",
                        "checkpoint_events": [
                            "delete",
                            "refactor",
                            "merge",
                            "script",
                            "critical-file",
                        ],
                        "checkpoint_type": "local-branch",
                        "max_checkpoints": 10,
                        "cleanup_days": 7,
                        "push_to_remote": False,
                        "auto_commit": True,
                        "branch_prefix": "feature/SPEC-",
                        "develop_branch": "develop",
                        "main_branch": "main",
                        "prevent_branch_creation": False,
                        "work_on_main": False,
                    },
                    "team": {
                        "auto_pr": False,
                        "develop_branch": "develop",
                        "draft_pr": False,
                        "branch_prefix": "feature/SPEC-",
                        "main_branch": "main",
                        "use_gitflow": True,
                        "default_pr_base": "develop",
                        "prevent_main_direct_merge": False,
                    },
                },
                "session": {
                    "suppress_setup_messages": False,
                    "notes": "suppress_setup_messages: false enables SessionStart output. "
                    "Set to true to suppress messages (show again after 7 days)",
                },
            }
            config_files = self.executor.execute_configuration_phase(self.path, config_data, progress_callback)

            # Phase 5: Validation (verify and finalize)
            self.executor.execute_validation_phase(self.path, mode, progress_callback)

            # Phase 6: Create runtime memory files (auto-generated per user/session)
            memory_files = self._create_memory_files()

            # Phase 7: Create user settings (gitignored, user-specific)
            user_settings_files = self._create_user_settings()

            # Generate result
            duration = int((time.time() - start_time) * 1000)  # ms
            return InstallationResult(
                success=True,
                project_path=str(self.path),
                language=detected_language,
                mode=mode,
                locale=locale,
                duration=duration,
                created_files=resource_files + config_files + memory_files + user_settings_files,
            )

        except Exception as e:
            duration = int((time.time() - start_time) * 1000)
            return InstallationResult(
                success=False,
                project_path=str(self.path),
                language=language or "unknown",
                mode=mode,
                locale=locale,
                duration=duration,
                created_files=[],
                errors=[str(e)],
            )

    def is_initialized(self) -> bool:
        """Check if .moai/ directory exists

        Returns:
            Whether initialized
        """
        return (self.path / ".moai").exists()


def initialize_project(project_path: Path, progress_callback: ProgressCallback | None = None) -> InstallationResult:
    """Initialize project (for CLI command)

    Args:
        project_path: Project directory path
        progress_callback: Progress callback

    Returns:
        InstallationResult object
    """
    initializer = ProjectInitializer(project_path)
    return initializer.initialize(progress_callback=progress_callback)
