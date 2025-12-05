"""
Rollback Manager for Research Integration Changes

Provides comprehensive rollback system for:
- Configuration backup and restore
- Version management
- Safe rollback procedures
- Integration with existing MoAI-ADK backup systems
- Research-specific rollback operations

Supports:
- Full system rollback
- Component-specific rollback
- Incremental rollback
- Emergency rollback
- Rollback validation and verification
"""

import hashlib
import json
import logging
import os
import shutil
import sys
from dataclasses import asdict, dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Dict, List, Tuple

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@dataclass
class RollbackPoint:
    """Represents a rollback point with metadata"""

    id: str
    timestamp: datetime
    description: str
    changes: List[str]
    backup_path: str
    checksum: str
    metadata: Dict[str, Any]


@dataclass
class RollbackResult:
    """Result of a rollback operation"""

    success: bool
    rollback_point_id: str
    message: str
    restored_files: List[str]
    failed_files: List[str] = None
    validation_results: Dict[str, Any] = None


class RollbackManager:
    """Comprehensive rollback management system"""

    def __init__(self, project_root: Path = None):
        self.project_root = project_root or Path.cwd()
        self.backup_root = self.project_root / ".moai" / "rollbacks"
        self.config_backup_dir = self.backup_root / "config"
        self.code_backup_dir = self.backup_root / "code"
        self.docs_backup_dir = self.backup_root / "docs"
        self.registry_file = self.backup_root / "rollback_registry.json"

        # Create backup directories
        self.backup_root.mkdir(parents=True, exist_ok=True)
        self.config_backup_dir.mkdir(parents=True, exist_ok=True)
        self.code_backup_dir.mkdir(parents=True, exist_ok=True)
        self.docs_backup_dir.mkdir(parents=True, exist_ok=True)

        # Load existing registry
        self.registry = self._load_registry()

        # Research-specific paths
        self.research_dirs = [
            self.project_root / ".claude" / "skills",
            self.project_root / ".claude" / "agents",
            self.project_root / ".claude" / "commands",
            self.project_root / ".claude" / "hooks",
        ]

    def create_rollback_point(self, description: str, changes: List[str] = None) -> str:
        """
        Create a rollback point before making changes

        Args:
            description: Description of the changes being made
            changes: List of specific changes (files modified, components updated)

        Returns:
            Rollback point ID
        """
        rollback_id = self._generate_rollback_id()
        timestamp = datetime.now(timezone.utc)

        logger.info(f"Creating rollback point {rollback_id}: {description}")

        try:
            # Create backup directory for this rollback point
            rollback_dir = self.backup_root / rollback_id
            rollback_dir.mkdir(parents=True, exist_ok=True)

            # Backup configuration files
            config_backup_path = self._backup_configuration(rollback_dir)

            # Backup research components
            research_backup_path = self._backup_research_components(rollback_dir)

            # Backup project files
            code_backup_path = self._backup_code_files(rollback_dir)

            # Create checksum for integrity verification
            checksum = self._calculate_backup_checksum(rollback_dir)

            # Create rollback point record
            rollback_point = RollbackPoint(
                id=rollback_id,
                timestamp=timestamp,
                description=description,
                changes=changes or [],
                backup_path=str(rollback_dir),
                checksum=checksum,
                metadata={
                    "config_backup": config_backup_path,
                    "research_backup": research_backup_path,
                    "code_backup": code_backup_path,
                    "project_root": str(self.project_root),
                    "created_by": "rollback_manager",
                    "version": "1.0.0",
                },
            )

            # Register rollback point
            self.registry[rollback_id] = asdict(rollback_point)
            self._save_registry()

            logger.info(f"Rollback point {rollback_id} created successfully")
            return rollback_id

        except Exception as e:
            logger.error(f"Failed to create rollback point: {str(e)}")
            # Cleanup partial backup
            self._cleanup_partial_backup(rollback_id)
            raise

    def rollback_to_point(
        self,
        rollback_id: str,
        validate_before: bool = True,
        validate_after: bool = True,
    ) -> RollbackResult:
        """
        Rollback to a specific rollback point

        Args:
            rollback_id: ID of rollback point to restore
            validate_before: Validate rollback point before restoration
            validate_after: Validate system after restoration

        Returns:
            RollbackResult with operation details
        """
        if rollback_id not in self.registry:
            return RollbackResult(
                success=False,
                rollback_point_id=rollback_id,
                message=f"Rollback point {rollback_id} not found",
                restored_files=[],
            )

        logger.info(f"Rolling back to point {rollback_id}")

        try:
            rollback_point = RollbackPoint(**self.registry[rollback_id])

            # Pre-rollback validation
            if validate_before:
                validation_result = self._validate_rollback_point(rollback_point)
                if not validation_result["valid"]:
                    return RollbackResult(
                        success=False,
                        rollback_point_id=rollback_id,
                        message=f"Rollback point validation failed: {validation_result['message']}",
                        restored_files=[],
                    )

            # Perform rollback
            restored_files, failed_files = self._perform_rollback(rollback_point)

            # Post-rollback validation
            validation_results = {}
            if validate_after:
                validation_results = self._validate_system_after_rollback()

            # Update registry with rollback info
            self._mark_rollback_as_used(rollback_id)

            success = len(failed_files) == 0

            result = RollbackResult(
                success=success,
                rollback_point_id=rollback_id,
                message=f"Rollback {'completed successfully' if success else 'completed with errors'}",
                restored_files=restored_files,
                failed_files=failed_files or [],
                validation_results=validation_results,
            )

            logger.info(f"Rollback {rollback_id} completed. Success: {success}")
            return result

        except Exception as e:
            logger.error(f"Rollback failed: {str(e)}")
            return RollbackResult(
                success=False,
                rollback_point_id=rollback_id,
                message=f"Rollback failed with error: {str(e)}",
                restored_files=[],
            )

    def rollback_research_integration(self, component_type: str = None, component_name: str = None) -> RollbackResult:
        """
        Specialized rollback for research integration changes

        Args:
            component_type: Type of component (skills, agents, commands, hooks)
            component_name: Specific component name to rollback

        Returns:
            RollbackResult with operation details
        """
        logger.info(f"Rolling back research integration: {component_type}:{component_name}")

        try:
            # Find relevant rollback points for research integration
            research_rollback_points = self._find_research_rollback_points(component_type, component_name)

            if not research_rollback_points:
                return RollbackResult(
                    success=False,
                    rollback_point_id="",
                    message="No suitable rollback points found for research integration",
                    restored_files=[],
                )

            # Use the most recent suitable rollback point
            latest_rollback = max(research_rollback_points, key=lambda x: x["timestamp"])

            # Perform targeted rollback
            restored_files, failed_files = self._perform_research_rollback(
                latest_rollback, component_type, component_name
            )

            # Validate research components
            validation_results = self._validate_research_components()

            success = len(failed_files) == 0

            return RollbackResult(
                success=success,
                rollback_point_id=latest_rollback["id"],
                message=f"Research integration rollback "
                f"{'completed successfully' if success else 'completed with errors'}",
                restored_files=restored_files,
                failed_files=failed_files or [],
                validation_results=validation_results,
            )

        except Exception as e:
            logger.error(f"Research integration rollback failed: {str(e)}")
            return RollbackResult(
                success=False,
                rollback_point_id="",
                message=f"Research integration rollback failed: {str(e)}",
                restored_files=[],
            )

    def list_rollback_points(self, limit: int = 10) -> List[Dict[str, Any]]:
        """
        List available rollback points

        Args:
            limit: Maximum number of rollback points to return

        Returns:
            List of rollback point information
        """
        rollback_points = []

        for rollback_id, rollback_data in self.registry.items():
            rollback_points.append(
                {
                    "id": rollback_id,
                    "timestamp": rollback_data["timestamp"],
                    "description": rollback_data["description"],
                    "changes_count": len(rollback_data.get("changes", [])),
                    "used": rollback_data.get("used", False),
                }
            )

        # Sort by timestamp (newest first) and limit
        rollback_points.sort(key=lambda x: x["timestamp"], reverse=True)
        return rollback_points[:limit]

    def validate_rollback_system(self) -> Dict[str, Any]:
        """
        Validate the rollback system integrity

        Returns:
            Validation results with system health information
        """
        issues: List[str] = []
        recommendations: List[str] = []
        validation_results = {
            "system_healthy": True,
            "issues": issues,
            "recommendations": recommendations,
            "rollback_points_count": len(self.registry),
            "backup_size": self._calculate_backup_size(),
            "last_rollback": None,
        }

        try:
            # Check backup directories exist
            required_dirs = [
                self.backup_root,
                self.config_backup_dir,
                self.code_backup_dir,
                self.docs_backup_dir,
            ]

            for dir_path in required_dirs:
                if not dir_path.exists():
                    issues.append(f"Missing backup directory: {dir_path}")
                    validation_results["system_healthy"] = False

            # Validate rollback points
            invalid_rollback_points: List[str] = []
            for rollback_id, rollback_data in self.registry.items():
                backup_path = Path(rollback_data["backup_path"])
                if not backup_path.exists():
                    invalid_rollback_points.append(rollback_id)

            if invalid_rollback_points:
                issues.append(f"Invalid rollback points: {invalid_rollback_points}")
                validation_results["system_healthy"] = False

            # Check available disk space
            backup_size: int = validation_results["backup_size"]  # type: ignore[assignment]
            free_space = shutil.disk_usage(self.backup_root).free
            if backup_size > free_space * 0.8:  # Using more than 80% of free space
                recommendations.append("Consider cleaning up old rollback points")

            # Check last rollback
            if self.registry:
                last_rollback = max(self.registry.values(), key=lambda x: x["timestamp"])
                validation_results["last_rollback"] = last_rollback["timestamp"]

        except Exception as e:
            validation_results["system_healthy"] = False
            issues.append(f"Validation error: {str(e)}")

        return validation_results

    def cleanup_old_rollbacks(self, keep_count: int = 10, dry_run: bool = True) -> Dict[str, Any]:
        """
        Clean up old rollback points

        Args:
            keep_count: Number of recent rollback points to keep
            dry_run: If True, only show what would be deleted

        Returns:
            Cleanup operation results
        """
        rollback_points = list(self.registry.values())
        rollback_points.sort(key=lambda x: x["timestamp"], reverse=True)

        # Keep the most recent rollback points
        to_keep = rollback_points[:keep_count]
        to_delete = rollback_points[keep_count:]

        if dry_run:
            return {
                "dry_run": True,
                "would_delete_count": len(to_delete),
                "would_keep_count": len(to_keep),
                "would_free_space": sum(self._get_directory_size(Path(rp["backup_path"])) for rp in to_delete),
            }

        # Perform actual cleanup
        deleted_count = 0
        freed_space = 0

        for rollback_point in to_delete:
            try:
                backup_path = Path(rollback_point["backup_path"])
                if backup_path.exists():
                    size = self._get_directory_size(backup_path)
                    shutil.rmtree(backup_path)
                    freed_space += size

                # Remove from registry
                del self.registry[rollback_point["id"]]
                deleted_count += 1

            except Exception as e:
                logger.warning(f"Failed to delete rollback point {rollback_point['id']}: {str(e)}")

        # Save updated registry
        self._save_registry()

        return {
            "dry_run": False,
            "deleted_count": deleted_count,
            "kept_count": len(to_keep),
            "freed_space": freed_space,
        }

    def _generate_rollback_id(self) -> str:
        """Generate unique rollback point ID"""
        timestamp = datetime.now(timezone.utc).strftime("%Y%m%d_%H%M%S")
        random_suffix = hashlib.md5(os.urandom(4), usedforsecurity=False).hexdigest()[:8]
        return f"rollback_{timestamp}_{random_suffix}"

    def _load_registry(self) -> Dict[str, Any]:
        """Load rollback registry from file"""
        if self.registry_file.exists():
            try:
                with open(self.registry_file, "r", encoding="utf-8") as f:
                    return json.load(f)
            except Exception as e:
                logger.warning(f"Failed to load rollback registry: {str(e)}")

        return {}

    def _save_registry(self):
        """Save rollback registry to file"""
        try:
            with open(self.registry_file, "w", encoding="utf-8") as f:
                json.dump(self.registry, f, indent=2, default=str, ensure_ascii=False)
        except Exception as e:
            logger.error(f"Failed to save rollback registry: {str(e)}")
            raise

    def _backup_configuration(self, rollback_dir: Path) -> str:
        """Backup configuration files"""
        config_backup_path = rollback_dir / "config"
        config_backup_path.mkdir(parents=True, exist_ok=True)

        # Backup .moai/config/config.json
        config_file = self.project_root / ".moai" / "config" / "config.json"
        if config_file.exists():
            shutil.copy2(config_file, config_backup_path / "config.json")

        # Backup .claude/settings.json
        settings_file = self.project_root / ".claude" / "settings.json"
        if settings_file.exists():
            shutil.copy2(settings_file, config_backup_path / "settings.json")

        # Backup .claude/settings.local.json
        local_settings_file = self.project_root / ".claude" / "settings.local.json"
        if local_settings_file.exists():
            shutil.copy2(local_settings_file, config_backup_path / "settings.local.json")

        return str(config_backup_path)

    def _backup_research_components(self, rollback_dir: Path) -> str:
        """Backup research-specific components"""
        research_backup_path = rollback_dir / "research"
        research_backup_path.mkdir(parents=True, exist_ok=True)

        for research_dir in self.research_dirs:
            if research_dir.exists():
                dir_name = research_dir.name
                target_dir = research_backup_path / dir_name
                shutil.copytree(research_dir, target_dir, dirs_exist_ok=True)

        return str(research_backup_path)

    def _backup_code_files(self, rollback_dir: Path) -> str:
        """Backup important code files"""
        code_backup_path = rollback_dir / "code"
        code_backup_path.mkdir(parents=True, exist_ok=True)

        # Backup source code
        src_dir = self.project_root / "src"
        if src_dir.exists():
            shutil.copytree(src_dir, code_backup_path / "src", dirs_exist_ok=True)

        # Backup tests
        tests_dir = self.project_root / "tests"
        if tests_dir.exists():
            shutil.copytree(tests_dir, code_backup_path / "tests", dirs_exist_ok=True)

        # Backup documentation
        docs_dir = self.project_root / "docs"
        if docs_dir.exists():
            shutil.copytree(docs_dir, code_backup_path / "docs", dirs_exist_ok=True)

        return str(code_backup_path)

    def _calculate_backup_checksum(self, backup_dir: Path) -> str:
        """Calculate checksum for backup integrity verification"""
        checksum_hash = hashlib.sha256()

        for file_path in backup_dir.rglob("*"):
            if file_path.is_file():
                with open(file_path, "rb") as f:
                    # Update hash with file content and path
                    checksum_hash.update(f.read())
                    checksum_hash.update(str(file_path.relative_to(backup_dir)).encode())

        return checksum_hash.hexdigest()

    def _validate_rollback_point(self, rollback_point: RollbackPoint) -> Dict[str, Any]:
        """Validate rollback point before restoration"""
        warnings: List[str] = []
        validation_result = {
            "valid": True,
            "message": "Rollback point is valid",
            "warnings": warnings,
        }

        try:
            # Check backup directory exists
            backup_path = Path(rollback_point.backup_path)
            if not backup_path.exists():
                validation_result["valid"] = False
                validation_result["message"] = "Backup directory not found"
                return validation_result

            # Verify checksum
            current_checksum = self._calculate_backup_checksum(backup_path)
            if current_checksum != rollback_point.checksum:
                warnings.append("Backup checksum mismatch - possible corruption")

            # Check essential files exist
            required_files = [
                backup_path / "config" / "config.json",
                backup_path / "research",
            ]

            missing_files = [f for f in required_files if not f.exists()]
            if missing_files:
                warnings.append(f"Missing backup files: {missing_files}")

        except Exception as e:
            validation_result["valid"] = False
            validation_result["message"] = f"Validation error: {str(e)}"

        return validation_result

    def _perform_rollback(self, rollback_point: RollbackPoint) -> Tuple[List[str], List[str]]:
        """Perform the actual rollback operation"""
        backup_path = Path(rollback_point.backup_path)
        restored_files: List[str] = []
        failed_files: List[str] = []

        try:
            # Restore configuration
            config_backup = backup_path / "config"
            if config_backup.exists():
                for config_file in config_backup.rglob("*"):
                    if config_file.is_file():
                        target_path = self.project_root / ".moai" / config_file.relative_to(config_backup)
                        target_path.parent.mkdir(parents=True, exist_ok=True)
                        try:
                            shutil.copy2(config_file, target_path)
                            restored_files.append(str(target_path))
                        except Exception as e:
                            failed_files.append(f"{target_path}: {str(e)}")

            # Restore research components
            research_backup = backup_path / "research"
            if research_backup.exists():
                for research_file in research_backup.rglob("*"):
                    if research_file.is_file():
                        target_path = self.project_root / research_file.relative_to(research_backup)
                        target_path.parent.mkdir(parents=True, exist_ok=True)
                        try:
                            shutil.copy2(research_file, target_path)
                            restored_files.append(str(target_path))
                        except Exception as e:
                            failed_files.append(f"{target_path}: {str(e)}")

            # Restore code files
            code_backup = backup_path / "code"
            if code_backup.exists():
                for code_file in code_backup.rglob("*"):
                    if code_file.is_file():
                        target_path = self.project_root / code_file.relative_to(code_backup)
                        target_path.parent.mkdir(parents=True, exist_ok=True)
                        try:
                            shutil.copy2(code_file, target_path)
                            restored_files.append(str(target_path))
                        except Exception as e:
                            failed_files.append(f"{target_path}: {str(e)}")

        except Exception as e:
            logger.error(f"Rollback operation failed: {str(e)}")
            failed_files.append(f"rollback_operation: {str(e)}")

        return restored_files, failed_files

    def _perform_research_rollback(
        self,
        rollback_point: Dict[str, Any],
        component_type: str = None,
        component_name: str = None,
    ) -> Tuple[List[str], List[str]]:
        """Perform targeted research component rollback"""
        backup_path = Path(rollback_point["backup_path"])
        research_backup = backup_path / "research"

        restored_files: List[str] = []
        failed_files: List[str] = []

        if not research_backup.exists():
            failed_files.append("research_backup: Research backup not found")
            return restored_files, failed_files

        try:
            # Restore specific component or all research components
            if component_type:
                component_backup_dir = research_backup / component_type
                if component_backup_dir.exists():
                    target_dir = self.project_root / ".claude" / component_type

                    if component_name:
                        # Restore specific component
                        component_file = component_backup_dir / f"{component_name}.md"
                        if component_file.exists():
                            target_file = target_dir / f"{component_name}.md"
                            target_file.parent.mkdir(parents=True, exist_ok=True)
                            shutil.copy2(component_file, target_file)
                            restored_files.append(str(target_file))
                        else:
                            failed_files.append(f"{component_name}: Component file not found in backup")
                    else:
                        # Restore entire component type
                        if target_dir.exists():
                            shutil.rmtree(target_dir)
                        shutil.copytree(component_backup_dir, target_dir)
                        restored_files.append(str(target_dir))
                else:
                    failed_files.append(f"{component_type}: Component type not found in backup")
            else:
                # Restore all research components
                for research_dir in research_backup.iterdir():
                    if research_dir.is_dir():
                        target_dir = self.project_root / ".claude" / research_dir.name
                        if target_dir.exists():
                            shutil.rmtree(target_dir)
                        shutil.copytree(research_dir, target_dir)
                        restored_files.append(str(target_dir))

        except Exception as e:
            logger.error(f"Research rollback failed: {str(e)}")
            failed_files.append(f"research_rollback: {str(e)}")

        return restored_files, failed_files

    def _validate_system_after_rollback(self) -> Dict[str, Any]:
        """Validate system state after rollback"""
        issues: List[str] = []
        validation_results = {
            "config_valid": True,
            "research_valid": True,
            "issues": issues,
        }

        try:
            # Validate configuration
            config_file = self.project_root / ".moai" / "config" / "config.json"
            if config_file.exists():
                try:
                    with open(config_file, "r", encoding="utf-8") as f:
                        json.load(f)  # Validate JSON syntax
                except json.JSONDecodeError:
                    validation_results["config_valid"] = False
                    issues.append("Invalid JSON in config.json")
            else:
                validation_results["config_valid"] = False
                issues.append("config.json not found")

            # Validate research components
            for research_dir in self.research_dirs:
                if research_dir.exists():
                    # Check for readable files
                    for file_path in research_dir.rglob("*.md"):
                        try:
                            with open(file_path, "r", encoding="utf-8") as f:
                                f.read()  # Validate file can be read
                        except Exception as e:
                            validation_results["research_valid"] = False
                            issues.append(f"Cannot read {file_path}: {str(e)}")

        except Exception as e:
            issues.append(f"Validation error: {str(e)}")

        return validation_results

    def _validate_research_components(self) -> Dict[str, Any]:
        """Validate research components after rollback"""
        issues: List[str] = []
        validation_results = {
            "skills_valid": True,
            "agents_valid": True,
            "commands_valid": True,
            "hooks_valid": True,
            "issues": issues,
        }

        component_checks = [
            ("skills", "Skills", self.project_root / ".claude" / "skills"),
            ("agents", "Agents", self.project_root / ".claude" / "agents"),
            ("commands", "Commands", self.project_root / ".claude" / "commands"),
            ("hooks", "Hooks", self.project_root / ".claude" / "hooks"),
        ]

        for component_key, component_name, component_path in component_checks:
            if component_path.exists():
                # Check component structure
                files = list(component_path.rglob("*.md"))
                if not files:
                    validation_results[f"{component_key}_valid"] = False
                    issues.append(f"{component_name} directory is empty")

                # Validate file content
                for file_path in files[:5]:  # Check first 5 files
                    try:
                        with open(file_path, "r", encoding="utf-8") as f:
                            content = f.read()
                            if not content.strip():
                                validation_results[f"{component_key}_valid"] = False
                                issues.append(f"Empty file: {file_path}")
                    except Exception as e:
                        validation_results[f"{component_key}_valid"] = False
                        issues.append(f"Cannot read {file_path}: {str(e)}")
            else:
                validation_results[f"{component_key}_valid"] = False
                issues.append(f"{component_name} directory not found")

        return validation_results

    def _find_research_rollback_points(
        self, component_type: str = None, component_name: str = None
    ) -> List[Dict[str, Any]]:
        """Find rollback points related to research integration"""
        research_rollback_points = []

        for rollback_id, rollback_data in self.registry.items():
            # Check if rollback point has research backup
            backup_path = Path(rollback_data["backup_path"])
            research_backup = backup_path / "research"

            if not research_backup.exists():
                continue

            # Check for specific component match
            if component_type:
                component_backup = research_backup / component_type
                if component_backup.exists():
                    if component_name:
                        component_file = component_backup / f"{component_name}.md"
                        if component_file.exists():
                            research_rollback_points.append(rollback_data)
                    else:
                        research_rollback_points.append(rollback_data)
            else:
                # Include any rollback with research components
                research_rollback_points.append(rollback_data)

        return research_rollback_points

    def _mark_rollback_as_used(self, rollback_id: str):
        """Mark rollback point as used in registry"""
        if rollback_id in self.registry:
            self.registry[rollback_id]["used"] = True
            self.registry[rollback_id]["used_timestamp"] = datetime.now(timezone.utc).isoformat()
            self._save_registry()

    def _cleanup_partial_backup(self, rollback_id: str):
        """Clean up partial backup if creation failed"""
        try:
            backup_dir = self.backup_root / rollback_id
            if backup_dir.exists():
                shutil.rmtree(backup_dir)
        except Exception as e:
            logger.warning(f"Failed to cleanup partial backup {rollback_id}: {str(e)}")

    def _calculate_backup_size(self) -> int:
        """Calculate total size of all backups"""
        total_size = 0
        for rollback_id, rollback_data in self.registry.items():
            backup_path = Path(rollback_data["backup_path"])
            if backup_path.exists():
                total_size += self._get_directory_size(backup_path)
        return total_size

    def _get_directory_size(self, directory: Path) -> int:
        """Get total size of directory in bytes"""
        total_size = 0
        try:
            for file_path in directory.rglob("*"):
                if file_path.is_file():
                    total_size += file_path.stat().st_size
        except Exception:
            pass  # Ignore errors in size calculation
        return total_size


# Command-line interface for rollback manager
def main():
    """Command-line interface for rollback operations"""
    import argparse

    parser = argparse.ArgumentParser(description="MoAI-ADK Rollback Manager")
    subparsers = parser.add_subparsers(dest="command", help="Available commands")

    # Create rollback point
    create_parser = subparsers.add_parser("create", help="Create rollback point")
    create_parser.add_argument("description", help="Description of changes")
    create_parser.add_argument("--changes", nargs="*", help="List of changes")

    # List rollback points
    list_parser = subparsers.add_parser("list", help="List rollback points")
    list_parser.add_argument("--limit", type=int, default=10, help="Maximum number to show")

    # Perform rollback
    rollback_parser = subparsers.add_parser("rollback", help="Rollback to point")
    rollback_parser.add_argument("rollback_id", help="Rollback point ID")
    rollback_parser.add_argument("--no-validate", action="store_true", help="Skip validation")

    # Research rollback
    research_parser = subparsers.add_parser("research-rollback", help="Rollback research components")
    research_parser.add_argument("--type", help="Component type (skills, agents, commands, hooks)")
    research_parser.add_argument("--name", help="Component name")

    # Validate system
    subparsers.add_parser("validate", help="Validate rollback system")

    # Cleanup
    cleanup_parser = subparsers.add_parser("cleanup", help="Cleanup old rollback points")
    cleanup_parser.add_argument("--keep", type=int, default=10, help="Number to keep")
    cleanup_parser.add_argument("--execute", action="store_true", help="Execute cleanup (default: dry run)")

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        return

    # Initialize rollback manager
    rollback_manager = RollbackManager()

    try:
        if args.command == "create":
            rollback_id = rollback_manager.create_rollback_point(args.description, args.changes)
            print(f"Rollback point created: {rollback_id}")

        elif args.command == "list":
            rollback_points = rollback_manager.list_rollback_points(args.limit)
            print(f"Available rollback points (showing {len(rollback_points)}):")
            for rp in rollback_points:
                status = "USED" if rp["used"] else "AVAILABLE"
                print(f"  {rp['id']} - {rp['description']} ({status})")

        elif args.command == "rollback":
            result = rollback_manager.rollback_to_point(args.rollback_id, validate_before=not args.no_validate)
            if result.success:
                print("Rollback completed successfully")
                print(f"Restored {len(result.restored_files)} files")
            else:
                print(f"Rollback failed: {result.message}")

        elif args.command == "research-rollback":
            result = rollback_manager.rollback_research_integration(args.type, args.name)
            if result.success:
                print("Research rollback completed successfully")
            else:
                print(f"Research rollback failed: {result.message}")

        elif args.command == "validate":
            validation = rollback_manager.validate_rollback_system()
            print(f"Rollback system health: {'HEALTHY' if validation['system_healthy'] else 'UNHEALTHY'}")
            if validation["issues"]:
                print("Issues found:")
                for issue in validation["issues"]:
                    print(f"  - {issue}")
            if validation["recommendations"]:
                print("Recommendations:")
                for rec in validation["recommendations"]:
                    print(f"  - {rec}")

        elif args.command == "cleanup":
            result = rollback_manager.cleanup_old_rollbacks(args.keep, dry_run=not args.execute)
            if result["dry_run"]:
                print(f"Dry run: Would delete {result['would_delete_count']} rollback points")
                print(f"Would free {result['would_free_space'] / 1024 / 1024:.1f} MB")
            else:
                print(f"Deleted {result['deleted_count']} rollback points")
                print(f"Freed {result['freed_space'] / 1024 / 1024:.1f} MB")

    except Exception as e:
        print(f"Error: {str(e)}")
        sys.exit(1)


if __name__ == "__main__":
    main()
