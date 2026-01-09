"""Web Server Configuration

Configuration settings for the MoAI Web Backend including
server settings, database paths, and CORS configuration.
"""

from dataclasses import dataclass, field
from pathlib import Path


@dataclass
class WebConfig:
    """Configuration for the MoAI Web Backend

    Attributes:
        host: Server host address
        port: Server port number
        database_path: Path to SQLite database file
        cors_origins: List of allowed CORS origins
        debug: Enable debug mode
    """

    host: str = "127.0.0.1"
    port: int = 8080
    database_path: Path = field(default_factory=lambda: Path(".moai/web/moai.db"))
    cors_origins: list[str] = field(default_factory=lambda: ["http://localhost:3000", "http://127.0.0.1:3000"])
    debug: bool = False

    def __post_init__(self):
        """Ensure database directory exists"""
        if isinstance(self.database_path, str):
            self.database_path = Path(self.database_path)
