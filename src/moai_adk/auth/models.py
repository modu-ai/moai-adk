"""User model for authentication module."""

import uuid
import time
from dataclasses import dataclass
from datetime import datetime, timezone


@dataclass
class User:
    """User model for authentication.

    Represents a user in the authentication system with unique identifier,
    email, hashed password, and creation timestamp.
    """

    id: str
    email: str
    hashed_password: str
    created_at: datetime

    @classmethod
    def create(cls, email: str, hashed_password: str) -> "User":
        """Create a new user instance.

        Args:
            email: User's email address
            hashed_password: Already hashed password

        Returns:
            New User instance with generated UUID and current timestamp
        """
        # Use timezone-aware datetime for better compatibility
        created_at = datetime.now(timezone.utc)
        return cls(
            id=str(uuid.uuid4()),
            email=email,
            hashed_password=hashed_password,
            created_at=created_at
        )

    def to_dict(self) -> dict:
        """Convert user to dictionary representation.

        Returns:
            Dictionary with user's public information (excluding password)
        """
        return {
            "id": self.id,
            "email": self.email,
            "created_at": self.created_at.isoformat()
        }

    def __repr__(self) -> str:
        """String representation of the user."""
        return f"User(id={self.id}, email={self.email}, created_at={self.created_at})"
