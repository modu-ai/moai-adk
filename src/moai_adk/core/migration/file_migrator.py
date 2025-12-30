"""
File migration module for MoAI-ADK version upgrades

Handles the actual file movement and directory creation
during migration processes.

Supports migration between:
- Legacy config.json (.moai/config.json) - pre-v0.24.0
- Intermediate config.json (.moai/config/config.json) - v0.24.0
- Section YAML files (.moai/config/sections/*.yaml) - v0.36.0+
"""

import json
import logging
import shutil
from pathlib import Path
from typing import Any, Dict, List

import yaml

logger = logging.getLogger(__name__)


class FileMigrator:
    """Handles file operations during migrations"""

    def __init__(self, project_root: Path):
        """
        Initialize file migrator

        Args:
            project_root: Root directory of the project
        """
        self.project_root = Path(project_root)
        self.moved_files: List[Dict[str, str]] = []
        self.created_dirs: List[str] = []

    def create_directory(self, directory: Path) -> bool:
        """
        Create a directory if it doesn't exist

        Args:
            directory: Directory path to create

        Returns:
            True if directory was created or already exists
        """
        try:
            directory = Path(directory)
            directory.mkdir(parents=True, exist_ok=True)
            self.created_dirs.append(str(directory))
            logger.debug(f"Created directory: {directory}")
            return True
        except Exception as e:
            logger.error(f"Failed to create directory {directory}: {e}")
            return False

    def move_file(self, source: Path, destination: Path, copy_instead: bool = True) -> bool:
        """
        Move a file from source to destination

        Args:
            source: Source file path
            destination: Destination file path
            copy_instead: If True, copy instead of move (safer)

        Returns:
            True if operation was successful
        """
        source = Path(source)
        destination = Path(destination)

        if not source.exists():
            logger.warning(f"Source file not found: {source}")
            return False

        if destination.exists():
            logger.info(f"Destination already exists, skipping: {destination}")
            return True

        try:
            # Ensure destination directory exists
            destination.parent.mkdir(parents=True, exist_ok=True)

            # Copy or move file
            if copy_instead:
                shutil.copy2(source, destination)
                logger.debug(f"Copied: {source} → {destination}")
            else:
                shutil.move(str(source), str(destination))
                logger.debug(f"Moved: {source} → {destination}")

            # Record operation
            self.moved_files.append({"from": str(source), "to": str(destination)})

            return True

        except Exception as e:
            logger.error(f"Failed to move file {source} to {destination}: {e}")
            return False

    def delete_file(self, file_path: Path, safe: bool = True) -> bool:
        """
        Delete a file

        Args:
            file_path: Path to the file to delete
            safe: If True, only delete if it's a known safe file

        Returns:
            True if deletion was successful
        """
        file_path = Path(file_path)

        if not file_path.exists():
            logger.debug(f"File already deleted: {file_path}")
            return True

        # Safety check for safe mode
        if safe:
            safe_patterns = [
                # Legacy and intermediate config.json
                ".moai/config.json",
                ".moai/config/config.json",
                # Statusline configs
                ".claude/statusline-config.yaml",
                ".moai/config/statusline-config.yaml",
                # Section YAML files (v0.36.0+)
                ".moai/config/sections/user.yaml",
                ".moai/config/sections/language.yaml",
                ".moai/config/sections/project.yaml",
                ".moai/config/sections/git-strategy.yaml",
                ".moai/config/sections/quality.yaml",
                ".moai/config/sections/system.yaml",
            ]
            try:
                rel_path = str(file_path.relative_to(self.project_root))
                is_safe = any(rel_path.endswith(pattern) for pattern in safe_patterns)
            except ValueError:
                is_safe = False

            if not is_safe:
                logger.warning(f"Refusing to delete non-safe file: {file_path}")
                return False

        try:
            file_path.unlink()
            logger.debug(f"Deleted: {file_path}")
            return True
        except Exception as e:
            logger.error(f"Failed to delete file {file_path}: {e}")
            return False

    def execute_migration_plan(self, plan: Dict[str, Any]) -> Dict[str, Any]:
        """
        Execute a migration plan

        Args:
            plan: Migration plan dictionary with 'create', 'move', 'convert', 'cleanup' keys

        Returns:
            Dictionary with execution results
        """
        results: Dict[str, Any] = {
            "success": True,
            "created_dirs": 0,
            "moved_files": 0,
            "converted_files": 0,
            "cleaned_files": 0,
            "errors": [],
        }

        # Create directories
        for directory in plan.get("create", []):
            dir_path = self.project_root / directory
            if self.create_directory(dir_path):
                results["created_dirs"] += 1
            else:
                results["errors"].append(f"Failed to create directory: {directory}")
                results["success"] = False

        # Move files
        for move_op in plan.get("move", []):
            source = self.project_root / move_op["from"]
            dest = self.project_root / move_op["to"]

            if self.move_file(source, dest, copy_instead=True):
                results["moved_files"] += 1
                logger.info(f"✅ {move_op['description']}: {move_op['from']} → {move_op['to']}")
            else:
                results["errors"].append(f"Failed to move: {move_op['from']} → {move_op['to']}")
                results["success"] = False

        # Convert config.json to section YAML files (v0.36.0+ migration)
        for convert_op in plan.get("convert", []):
            source = self.project_root / convert_op["from"]
            target_dir = self.project_root / convert_op["to_dir"]

            if self.convert_json_to_section_yaml(source, target_dir):
                results["converted_files"] += 1
                logger.info(f"✅ {convert_op['description']}: {convert_op['from']} → {convert_op['to_dir']}/")
            else:
                results["errors"].append(f"Failed to convert: {convert_op['from']} → {convert_op['to_dir']}/")
                results["success"] = False

        return results

    def convert_json_to_section_yaml(self, source_json: Path, target_dir: Path) -> bool:
        """
        Convert a config.json file to section YAML files

        Args:
            source_json: Path to the source config.json file
            target_dir: Path to the target sections directory

        Returns:
            True if conversion was successful
        """
        if not source_json.exists():
            logger.warning(f"Source JSON file not found: {source_json}")
            return False

        try:
            # Read the source JSON
            with open(source_json, "r", encoding="utf-8") as f:
                config_data = json.load(f)

            # Ensure target directory exists
            target_dir.mkdir(parents=True, exist_ok=True)

            # Map JSON keys to section files
            section_mapping = self._get_section_mapping(config_data)

            # Write each section to its YAML file
            for section_name, section_data in section_mapping.items():
                if section_data:
                    yaml_file = target_dir / f"{section_name}.yaml"
                    # Skip if file already exists (don't overwrite)
                    if yaml_file.exists():
                        logger.debug(f"Section file already exists, skipping: {yaml_file}")
                        continue

                    with open(yaml_file, "w", encoding="utf-8") as f:
                        # Add header comment
                        f.write(f"# {section_name.replace('-', ' ').title()} Settings\n")
                        f.write("# Migrated from config.json\n\n")
                        yaml.dump(section_data, f, default_flow_style=False, allow_unicode=True)
                    logger.debug(f"Created section file: {yaml_file}")

            logger.info(f"Successfully converted {source_json} to section YAML files")
            return True

        except Exception as e:
            logger.error(f"Failed to convert {source_json} to section YAML: {e}")
            return False

    def _get_section_mapping(self, config_data: Dict[str, Any]) -> Dict[str, Dict[str, Any]]:
        """
        Map config.json data to section YAML structure

        Args:
            config_data: The config.json data as a dictionary

        Returns:
            Dictionary mapping section names to their data
        """
        sections: Dict[str, Dict[str, Any]] = {
            "user": {},
            "language": {},
            "project": {},
            "git-strategy": {},
            "quality": {},
            "system": {},
        }

        # Map user settings
        if "user_name" in config_data:
            sections["user"]["user"] = {"name": config_data["user_name"]}
        if "user" in config_data:
            sections["user"]["user"] = config_data["user"]

        # Map language settings
        language_keys = ["conversation_language", "agent_prompt_language", "documentation_language", "code_comments"]
        language_data = {}
        for key in language_keys:
            if key in config_data:
                language_data[key] = config_data[key]
        if "language" in config_data:
            language_data.update(config_data["language"])
        if language_data:
            sections["language"]["language"] = language_data

        # Map project settings
        project_keys = ["project_name", "project_type", "project_description", "locale", "created_at", "initialized"]
        project_data = {}
        for key in project_keys:
            if key in config_data:
                # Normalize key names
                normalized_key = key.replace("project_", "")
                project_data[normalized_key] = config_data[key]
        if "project" in config_data:
            project_data.update(config_data["project"])
        if project_data:
            sections["project"]["project"] = project_data

        # Map git strategy settings
        git_keys = ["git_strategy", "auto_branch", "branch_prefix", "commit_style"]
        git_data = {}
        for key in git_keys:
            if key in config_data:
                git_data[key] = config_data[key]
        if "git" in config_data:
            git_data.update(config_data["git"])
        if git_data:
            sections["git-strategy"]["git"] = git_data

        # Map quality settings
        quality_keys = ["tdd_enabled", "test_framework", "coverage_threshold"]
        quality_data = {}
        for key in quality_keys:
            if key in config_data:
                quality_data[key] = config_data[key]
        if "quality" in config_data:
            quality_data.update(config_data["quality"])
        if "tdd" in config_data:
            quality_data["tdd"] = config_data["tdd"]
        if quality_data:
            sections["quality"]["quality"] = quality_data

        # Map system settings
        system_data: Dict[str, Any] = {}
        if "moai_version" in config_data:
            system_data["moai"] = {"version": config_data["moai_version"]}
        if "moai" in config_data:
            system_data["moai"] = config_data["moai"]
        if "github" in config_data:
            system_data["github"] = config_data["github"]
        if system_data:
            sections["system"] = system_data

        return sections

    def cleanup_old_files(self, cleanup_list: List[str], dry_run: bool = False) -> int:
        """
        Clean up old files after successful migration

        Args:
            cleanup_list: List of file paths to clean up
            dry_run: If True, only show what would be deleted

        Returns:
            Number of files cleaned up
        """
        cleaned = 0

        for file_path in cleanup_list:
            full_path = self.project_root / file_path

            if dry_run:
                if full_path.exists():
                    logger.info(f"Would delete: {file_path}")
                    cleaned += 1
            else:
                if self.delete_file(full_path, safe=True):
                    logger.info(f"🗑️  Cleaned up: {file_path}")
                    cleaned += 1

        return cleaned

    def get_migration_summary(self) -> Dict[str, Any]:
        """
        Get summary of migration operations performed

        Returns:
            Dictionary with migration summary
        """
        return {
            "moved_files": len(self.moved_files),
            "created_directories": len(self.created_dirs),
            "operations": self.moved_files,
        }
