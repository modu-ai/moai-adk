#!/usr/bin/env python3
"""Database operations for validate_tags module"""

import sqlite3
from pathlib import Path
from typing import Any, Dict
from .parser import TagHealthReport


def save_report_to_sqlite(report: TagHealthReport, report_file: Path, report_data: Dict[str, Any]) -> None:
    """Save validation report to SQLite database"""
    try:
        report_file.parent.mkdir(parents=True, exist_ok=True)
        conn = sqlite3.connect(report_file)
        cursor = conn.cursor()

        cursor.execute("""
            CREATE TABLE IF NOT EXISTS validation_report (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                metric TEXT NOT NULL,
                value TEXT NOT NULL,
                details TEXT,
                created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        """)

        # Insert basic metrics
        metrics = [
            ('timestamp', report_data["timestamp"], ''),
            ('total_tags', str(report_data["summary"]["total_tags"]), ''),
            ('valid_tags', str(report_data["summary"]["valid_tags"]), ''),
            ('quality_score', str(report_data["summary"]["quality_score"]), '')
        ]
        cursor.executemany(
            "INSERT INTO validation_report (metric, value, details) VALUES (?, ?, ?)",
            metrics
        )

        for issue_type, issues in report_data["issues"].items():
            cursor.execute(
                "INSERT INTO validation_report (metric, value, details) VALUES (?, ?, ?)",
                (issue_type, str(len(issues)), str(issues)[:500]),
            )

        conn.commit()
        conn.close()

    except Exception as e:
        print(f"\n⚠️  Failed to save report: {e}")


def load_sqlite_index(db_path: Path) -> Dict[str, Any]:
    """Load tag index from SQLite database"""
    if not db_path.exists():
        return {}

    try:
        conn = sqlite3.connect(db_path)
        cursor = conn.cursor()
        conn.close()
        return {}
    except Exception:
        return {}