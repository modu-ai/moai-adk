"""Authentication service for user management and login/logout operations.

This module provides an AuthService class that handles:
- User creation and storage
- Login with email and password
- JWT token issuance and validation
- Token revocation and blacklisting
"""

from typing import Dict, Any, Optional, Set
from .models import User
from .security import (
    hash_password,
    verify_password,
    create_token,
    verify_token as verify_jwt_token
)


class AuthService:
    """Service for handling user authentication and token management.

    This service manages user accounts, authentication flows, and JWT tokens.
    It stores users in memory and maintains a token blacklist for revocation.

    Note:
        For production use, replace in-memory storage with a proper database.
    """

    def __init__(self):
        """Initialize the authentication service with empty storage."""
        self._users: Dict[str, User] = {}  # Email-indexed user storage
        self._blacklist: Set[str] = set()  # Revoked JWT tokens

    def create_user(self, email: str, password: str) -> User:
        """Create a new user with email and password.

        Password is hashed using bcrypt before storage.

        Args:
            email: User's unique email address
            password: Plain text password to hash and store

        Returns:
            Created User object with ID and timestamp

        Raises:
            ValueError: If email already exists or inputs are invalid
        """
        if not email or "@" not in email:
            raise ValueError("Email must be valid")
        if not password or len(password) < 8:
            raise ValueError("Password must be at least 8 characters")

        # Check if user already exists
        if email in self._users:
            raise ValueError(f"User with email {email} already exists")

        # Hash password securely
        hashed_password = hash_password(password)

        # Create and store user
        user = User.create(email=email, hashed_password=hashed_password)
        self._users[email] = user

        return user

    def login(self, email: str, password: str) -> Dict[str, Any]:
        """Login user and issue a JWT token.

        Authenticates user by email and password, then issues a signed JWT token
        valid for 1 hour.

        Args:
            email: User email address
            password: User password (plain text)

        Returns:
            Dictionary containing:
            - access_token: Signed JWT token
            - token_type: "bearer"
            - expires_in: 3600 (seconds, 1 hour)

        Raises:
            ValueError: If email not found ("User not found") or
                       password incorrect ("Invalid password")
        """
        if not email or not password:
            raise ValueError("Email and password are required")

        # Check if user exists
        if email not in self._users:
            raise ValueError("User not found")

        user = self._users[email]

        # Verify password (timing-safe comparison)
        if not verify_password(password, user.hashed_password):
            raise ValueError("Invalid password")

        # Issue JWT token valid for 1 hour
        token = create_token(user_id=user.id, email=user.email)

        return {
            "access_token": token,
            "token_type": "bearer",
            "expires_in": 3600  # 1 hour in seconds
        }

    def get_user_from_token(self, token: Optional[str]) -> Dict[str, Any]:
        """Get user information from a valid JWT token.

        Validates token signature, expiration, and revocation status.

        Args:
            token: JWT token string (typically from "Bearer {token}" header)

        Returns:
            Dictionary with user information:
            - id: User UUID
            - email: User email address
            - created_at: ISO8601 timestamp of account creation

        Raises:
            ValueError: If token is invalid, expired, or revoked
        """
        if not token:
            raise ValueError("Token required")

        # Check if token is blacklisted (revoked)
        if self.is_token_blacklisted(token):
            raise ValueError("Token has been revoked")

        # Verify and decode token (may raise ValueError if expired/invalid)
        payload = verify_jwt_token(token)

        # Extract claims from token
        user_id = payload.get("user_id")
        email = payload.get("email")

        # Find user (validate token refers to existing user)
        if not email or email not in self._users:
            raise ValueError("User not found")

        user = self._users[email]

        return {
            "id": user.id,
            "email": user.email,
            "created_at": user.created_at
        }

    def logout(self, token: str) -> bool:
        """Logout user by revoking (blacklisting) their token.

        Once revoked, the token cannot be used for authentication.

        Args:
            token: JWT token to revoke

        Returns:
            True if logout successful

        Raises:
            ValueError: If token is invalid, expired, or already revoked
        """
        if not token:
            raise ValueError("Token required")

        # Verify token is valid before blacklisting
        payload = verify_jwt_token(token)

        # Check if already blacklisted
        if token in self._blacklist:
            raise ValueError("Token has been revoked")

        # Add to blacklist
        self._blacklist.add(token)

        return True

    def is_token_blacklisted(self, token: str) -> bool:
        """Check if a token is in the revocation blacklist.

        Args:
            token: JWT token to check

        Returns:
            True if token is revoked, False if still valid
        """
        return token in self._blacklist

    def verify_token(self, token: str) -> Dict[str, Any]:
        """Verify and decode a JWT token without checking blacklist.

        Use this only when you need raw token validation.
        For user authentication, use get_user_from_token() instead,
        which checks blacklist.

        Args:
            token: JWT token to verify

        Returns:
            Decoded token payload

        Raises:
            ValueError: If token is invalid or expired
        """
        return verify_jwt_token(token)
