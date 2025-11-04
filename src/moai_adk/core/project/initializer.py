# @CODE:CORE-PROJECT-001 | @CODE:INIT-003:INIT
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

import json
import time
from datetime import datetime
from pathlib import Path

from moai_adk.core.project.phase_executor import PhaseExecutor, ProgressCallback
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

        @CODE:INIT-MEMORY-001 | Auto-generate session memory files
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
                "git_operations": "daily commits, feature branches"
            },
            "next_priorities": []
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
            "current_branch": "main"
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
                "estimated_level": "beginner"
            },
            "skip_questions": [],
            "last_updated": datetime.now().isoformat() + "Z"
        }
        user_patterns_file = memory_dir / "user-patterns.json"
        user_patterns_file.write_text(json.dumps(user_patterns, indent=2))
        created_files.append(str(user_patterns_file))

        return created_files

    def initialize(
        self,
        mode: str = "personal",
        locale: str = "en",  # Changed from "ko" to "en" (will be configurable in /alfred:0-project)
        language: str | None = None,
        backup_enabled: bool = True,
        progress_callback: ProgressCallback | None = None,
        reinit: bool = False,
    ) -> InstallationResult:
        """Execute project initialization (5-phase process)

        Args:
            mode: Project mode (personal/team) - Default: personal (configurable in /alfred:0-project)
            locale: Locale (ko/en/ja/zh) - Default: en (configurable in /alfred:0-project)
            language: Force language specification (auto-detect if None) - Will be detected in /alfred:0-project
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
            # Language detection now happens in /alfred:0-project via project-manager
            detected_language = language or "generic"

            # Phase 1: Preparation (backup and validation)
            self.executor.execute_preparation_phase(
                self.path, backup_enabled, progress_callback
            )

            # Phase 2: Directory (create directories)
            self.executor.execute_directory_phase(self.path, progress_callback)

            # Prepare config for template variable substitution (Phase 3)
            config = {
                "name": self.path.name,
                "mode": mode,
                "locale": locale,
                "language": detected_language,
                "description": "",
                "version": "0.1.0",
                "author": "@user",
            }

            # Phase 3: Resource (copy templates with variable substitution)
            resource_files = self.executor.execute_resource_phase(
                self.path, config, progress_callback
            )

            # Phase 4: Configuration (generate config.json)
            config_data: dict[str, str | bool | dict] = {
                "project": {
                    "name": self.path.name,
                    "mode": mode,
                    "locale": locale,
                    "language": detected_language,
                    # Language detection metadata (will be updated by project-manager via /alfred:0-project)
                    "language_detection": {
                        "detected_language": detected_language,
                        "detection_method": "cli_default",  # Will be "context_aware" after /alfred:0-project
                        "confidence": None,  # Will be calculated by project-manager
                        "markers": [],  # Will be populated by project-manager
                        "confirmed_by": None,  # Will be "user" after project-manager confirmation
                    }
                }
            }
            config_files = self.executor.execute_configuration_phase(
                self.path, config_data, progress_callback
            )

            # Phase 5: Validation (verify and finalize)
            self.executor.execute_validation_phase(
                self.path, mode, progress_callback
            )

            # Phase 6: Create runtime memory files (auto-generated per user/session)
            memory_files = self._create_memory_files()

            # Generate result
            duration = int((time.time() - start_time) * 1000)  # ms
            return InstallationResult(
                success=True,
                project_path=str(self.path),
                language=detected_language,
                mode=mode,
                locale=locale,
                duration=duration,
                created_files=resource_files + config_files + memory_files,
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


def initialize_project(
    project_path: Path, progress_callback: ProgressCallback | None = None
) -> InstallationResult:
    """Initialize project (for CLI command)

    Args:
        project_path: Project directory path
        progress_callback: Progress callback

    Returns:
        InstallationResult object
    """
    initializer = ProjectInitializer(project_path)
    return initializer.initialize(progress_callback=progress_callback)
