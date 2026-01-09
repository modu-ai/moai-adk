"""SQLite Async Database Setup

Provides async SQLite database connection management using aiosqlite.
Includes database initialization, connection pooling, and schema setup.
"""

from pathlib import Path
from typing import Optional

import aiosqlite

from moai_adk.web.config import WebConfig

# Global database connection
_db_connection: Optional[aiosqlite.Connection] = None
_config: Optional[WebConfig] = None


async def init_database(config: WebConfig | None = None) -> aiosqlite.Connection:
    """Initialize the SQLite database connection and create tables

    Args:
        config: Optional WebConfig. Uses default if not provided.

    Returns:
        The database connection instance
    """
    global _db_connection, _config

    if config is None:
        config = WebConfig()

    _config = config

    # Ensure database directory exists
    db_path = config.database_path
    db_path.parent.mkdir(parents=True, exist_ok=True)

    # Create connection
    _db_connection = await aiosqlite.connect(db_path)
    _db_connection.row_factory = aiosqlite.Row

    # Create tables
    await _create_tables(_db_connection)

    return _db_connection


async def _create_tables(db: aiosqlite.Connection) -> None:
    """Create database tables if they don't exist"""
    # Sessions table
    await db.execute("""
        CREATE TABLE IF NOT EXISTS sessions (
            id TEXT PRIMARY KEY,
            title TEXT NOT NULL DEFAULT 'New Session',
            provider TEXT NOT NULL DEFAULT 'claude',
            model TEXT NOT NULL DEFAULT 'claude-sonnet-4-20250514',
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    """)

    # Messages table
    await db.execute("""
        CREATE TABLE IF NOT EXISTS messages (
            id TEXT PRIMARY KEY,
            session_id TEXT NOT NULL,
            role TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'system')),
            content TEXT NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
        )
    """)

    await db.commit()


async def get_database() -> aiosqlite.Connection:
    """Get the current database connection

    Returns:
        The active database connection

    Raises:
        RuntimeError: If database is not initialized
    """
    global _db_connection

    if _db_connection is None:
        raise RuntimeError("Database not initialized. Call init_database() first.")

    return _db_connection


async def close_database() -> None:
    """Close the database connection"""
    global _db_connection

    if _db_connection is not None:
        await _db_connection.close()
        _db_connection = None


def get_db_path() -> Path:
    """Get the current database path

    Returns:
        Path to the database file
    """
    global _config

    if _config is None:
        return WebConfig().database_path

    return _config.database_path
