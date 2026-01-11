"""Secure configuration and credential storage for MoAI Rank.

Credentials are stored in ~/.moai/rank/credentials.json with
restricted permissions (chmod 600) to prevent unauthorized access.
"""

import json
import os
import stat
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Optional


@dataclass
class RankCredentials:
    """User credentials for MoAI Rank API."""

    api_key: str
    username: str
    user_id: str
    created_at: str


class RankConfig:
    """Manages secure storage and retrieval of MoAI Rank credentials."""

    # Service configuration
    DEFAULT_BASE_URL = "https://rank.mo.ai.kr"
    API_VERSION = "v1"

    # Credential storage paths
    CONFIG_DIR = Path.home() / ".moai" / "rank"
    CREDENTIALS_FILE = CONFIG_DIR / "credentials.json"
    LAST_SYNC_FILE = CONFIG_DIR / "last_sync.json"

    def __init__(self, base_url: Optional[str] = None):
        """Initialize configuration with optional custom base URL.

        Args:
            base_url: Override the default API base URL (for testing)
        """
        self.base_url = base_url or os.environ.get("MOAI_RANK_URL", self.DEFAULT_BASE_URL)

    @property
    def api_base_url(self) -> str:
        """Get the full API base URL including version prefix."""
        return f"{self.base_url}/api/{self.API_VERSION}"

    @classmethod
    def ensure_config_dir(cls) -> Path:
        """Create the configuration directory with secure permissions.

        Returns:
            Path to the configuration directory
        """
        cls.CONFIG_DIR.mkdir(parents=True, exist_ok=True)
        # Set directory permissions to owner-only (700)
        os.chmod(cls.CONFIG_DIR, stat.S_IRWXU)
        return cls.CONFIG_DIR

    @classmethod
    def save_credentials(cls, credentials: RankCredentials) -> None:
        """Save credentials securely to disk.

        The credentials file is created with mode 600 (owner read/write only)
        to prevent other users from accessing the API key.

        Args:
            credentials: The credentials to save
        """
        cls.ensure_config_dir()

        # Write credentials with secure permissions
        creds_data = asdict(credentials)

        # Write to a temporary file first, then rename (atomic operation)
        temp_file = cls.CREDENTIALS_FILE.with_suffix(".tmp")

        try:
            with open(temp_file, "w") as f:
                json.dump(creds_data, f, indent=2)

            # Set file permissions to owner-only (600) before renaming
            os.chmod(temp_file, stat.S_IRUSR | stat.S_IWUSR)

            # Atomic rename
            temp_file.rename(cls.CREDENTIALS_FILE)
        except Exception:
            # Clean up temp file on error
            if temp_file.exists():
                temp_file.unlink()
            raise

    @classmethod
    def load_credentials(cls) -> Optional[RankCredentials]:
        """Load credentials from disk.

        Returns:
            RankCredentials if found and valid, None otherwise
        """
        if not cls.CREDENTIALS_FILE.exists():
            return None

        try:
            with open(cls.CREDENTIALS_FILE) as f:
                data = json.load(f)

            return RankCredentials(
                api_key=data["api_key"],
                username=data["username"],
                user_id=data["user_id"],
                created_at=data["created_at"],
            )
        except (json.JSONDecodeError, KeyError, OSError):
            return None

    @classmethod
    def delete_credentials(cls) -> bool:
        """Delete stored credentials.

        Returns:
            True if credentials were deleted, False if they didn't exist
        """
        if cls.CREDENTIALS_FILE.exists():
            cls.CREDENTIALS_FILE.unlink()
            return True
        return False

    @classmethod
    def has_credentials(cls) -> bool:
        """Check if credentials are stored.

        Returns:
            True if credentials file exists
        """
        return cls.CREDENTIALS_FILE.exists()

    @classmethod
    def get_api_key(cls) -> Optional[str]:
        """Get the stored API key.

        Returns:
            API key string if available, None otherwise
        """
        creds = cls.load_credentials()
        return creds.api_key if creds else None
