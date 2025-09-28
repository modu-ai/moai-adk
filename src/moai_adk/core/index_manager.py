"""
Index and reporting system manager for MoAI projects.

@TASK:INDEX-MANAGER-001 Extracted from config_project.py for TRUST compliance
@DESIGN:SEPARATED-INDEX-001 Single responsibility: tags database and reports
"""

import sqlite3
from datetime import datetime
from pathlib import Path
from typing import Any

from .._version import get_version
from ..config import Config
from ..utils.logger import get_logger

logger = get_logger(__name__)


class IndexManager:
    """Manages TAG indexing system and project reports"""

    def create_initial_indexes(self, project_path: Path, config: Config) -> list[Path]:
        """
        Create initial index files for TAG system.

        Args:
            project_path: Project root directory
            config: Project configuration

        Returns:
            List of created file paths
        """
        created_files = []

        try:
            indexes_dir = project_path / ".moai" / "indexes"
            indexes_dir.mkdir(parents=True, exist_ok=True)

            # Create SQLite tags database
            tags_db_path = self._create_tags_database(indexes_dir)
            if tags_db_path:
                created_files.append(tags_db_path)

            # Create initial sync report
            sync_report_path = self._create_sync_report(project_path, config)
            if sync_report_path:
                created_files.append(sync_report_path)

        except Exception as e:
            logger.error(f"Failed to create initial indexes: {e}")

        return created_files

    def setup_steering_config(self, project_path: Path) -> Path | None:
        """
        Setup steering configuration for project governance.

        Args:
            project_path: Project root directory

        Returns:
            Path to created governance config or None on failure
        """
        steering_dir = project_path / ".moai" / "steering"
        steering_dir.mkdir(parents=True, exist_ok=True)

        config_path = steering_dir / "governance.json"

        if config_path.exists():
            logger.info(f"Steering config already exists at {config_path}")
            return config_path

        try:
            steering_config = self._build_steering_config()

            with open(config_path, "w", encoding="utf-8") as f:
                import json
                json.dump(steering_config, f, indent=2)

            logger.info(f"Created steering config at {config_path}")
            return config_path

        except Exception as e:
            logger.error(f"Failed to create steering config: {e}")
            return None

    def _create_tags_database(self, indexes_dir: Path) -> Path | None:
        """Create SQLite database for TAG indexing"""
        tags_db_path = indexes_dir / "tags.db"

        if tags_db_path.exists():
            logger.debug(f"Tags database already exists at {tags_db_path}")
            return tags_db_path

        try:
            conn = sqlite3.connect(tags_db_path)
            cursor = conn.cursor()

            # Create tables
            self._create_database_schema(cursor)
            self._insert_initial_statistics(cursor)
            self._create_database_indexes(cursor)

            conn.commit()
            conn.close()

            logger.info(f"Created SQLite tags database at {tags_db_path}")
            return tags_db_path

        except Exception as e:
            logger.error(f"Failed to create SQLite database: {e}")
            # Create empty file to maintain structure
            tags_db_path.touch()
            return tags_db_path

    def _create_database_schema(self, cursor: sqlite3.Cursor) -> None:
        """Create database schema for tags system"""
        cursor.execute("""
            CREATE TABLE tags (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                category TEXT NOT NULL,
                identifier TEXT NOT NULL,
                file_path TEXT NOT NULL,
                line_number INTEGER NOT NULL,
                context TEXT,
                created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        """)

        cursor.execute("""
            CREATE TABLE statistics (
                key TEXT PRIMARY KEY,
                value TEXT NOT NULL,
                updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        """)

        cursor.execute("""
            CREATE TABLE tag_references (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                source_tag_id INTEGER NOT NULL,
                target_tag_id INTEGER NOT NULL,
                reference_type TEXT DEFAULT 'chain',
                created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (source_tag_id) REFERENCES tags (id),
                FOREIGN KEY (target_tag_id) REFERENCES tags (id)
            )
        """)

    def _insert_initial_statistics(self, cursor: sqlite3.Cursor) -> None:
        """Insert initial statistics data"""
        cursor.execute(
            "INSERT INTO statistics (key, value) VALUES ('version', '1.0.0')"
        )
        cursor.execute(
            "INSERT INTO statistics (key, value) VALUES ('total_tags', '0')"
        )
        cursor.execute(
            "INSERT INTO statistics (key, value) VALUES ('updated', ?)",
            (datetime.now().isoformat(),),
        )

    def _create_database_indexes(self, cursor: sqlite3.Cursor) -> None:
        """Create database indexes for performance"""
        cursor.execute("CREATE INDEX idx_tags_category ON tags(category)")
        cursor.execute("CREATE INDEX idx_tags_identifier ON tags(identifier)")
        cursor.execute("CREATE INDEX idx_tags_file ON tags(file_path)")

    def _create_sync_report(self, project_path: Path, config: Config) -> Path | None:
        """Create initial sync report"""
        reports_dir = project_path / ".moai" / "reports"
        reports_dir.mkdir(parents=True, exist_ok=True)

        sync_report_path = reports_dir / "sync-report.md"

        if sync_report_path.exists():
            logger.debug(f"Sync report already exists at {sync_report_path}")
            return sync_report_path

        try:
            report_content = self._build_initial_report(config)

            with open(sync_report_path, "w", encoding="utf-8") as f:
                f.write(report_content)

            logger.info(f"Created sync report at {sync_report_path}")
            return sync_report_path

        except Exception as e:
            logger.error(f"Failed to create sync report: {e}")
            return None

    def _build_initial_report(self, config: Config) -> str:
        """Build initial sync report content"""
        return f"""# MoAI-ADK Sync Report

## Project Information
- **Project Name**: {getattr(config, "name", "project")}
- **Created**: {datetime.now().strftime("%Y-%m-%d %H:%M:%S")}
- **MoAI Version**: {get_version()}

## Sync Status
- **Last Sync**: {datetime.now().isoformat()}
- **TAG Count**: 0
- **Document Status**: Ready

## Next Steps
1. Run `/moai:1-spec` to create your first specification
2. Use `/moai:2-build` for TDD implementation
3. Execute `/moai:3-sync` to update this report
"""

    def _build_steering_config(self) -> dict[str, Any]:
        """Build steering configuration structure"""
        return {
            "version": "1.0",
            "governance": {
                "decision_making": "consensus",
                "review_required": ["SPEC", "ADR"],
                "approval_threshold": 1,
            },
            "policies": {
                "trust_principles": {
                    "test_first": True,
                    "readable_code": True,
                    "unified_architecture": True,
                    "secured_development": True,
                    "trackable_changes": True,
                },
                "code_standards": {
                    "max_function_lines": 50,
                    "max_file_lines": 300,
                    "max_parameters": 5,
                    "max_complexity": 10,
                },
            },
            "workflows": {
                "spec_first": True,
                "tdd_required": True,
                "documentation_sync": True,
            },
            "created": datetime.now().isoformat(),
            "moai_version": get_version(),
        }