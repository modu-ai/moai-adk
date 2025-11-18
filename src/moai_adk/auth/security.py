"""Security utilities for password hashing and JWT token management.

This module provides cryptographic functions for:
- Password hashing and verification using bcrypt
- JWT token creation and validation
- Token expiration management
"""

import jwt
import bcrypt
import time
import os
from datetime import datetime, timedelta
from typing import Dict, Any

# Security constants
# TODO: Load from environment variable in production
SECRET_KEY = os.getenv("JWT_SECRET_KEY", "your-secret-key-change-in-production")
ALGORITHM = "HS256"
TOKEN_EXPIRE_MINUTES = 60
BCRYPT_ROUNDS = 12  # Cost factor for bcrypt hashing


def hash_password(password: str) -> str:
    """Hash a password using bcrypt with configurable cost factor.

    Uses bcrypt with salt to securely hash passwords. Each call produces
    a different hash due to random salt generation.

    Args:
        password: Plain text password to hash

    Returns:
        Hashed password string (includes salt)

    Raises:
        ValueError: If password is empty or invalid
    """
    if not password or not isinstance(password, str):
        raise ValueError("Password must be a non-empty string")

    salt = bcrypt.gensalt(rounds=BCRYPT_ROUNDS)
    return bcrypt.hashpw(password.encode("utf-8"), salt).decode("utf-8")


def verify_password(password: str, hashed_password: str) -> bool:
    """Verify a password against its bcrypt hash.

    Uses constant-time comparison to prevent timing attacks.

    Args:
        password: Plain text password to verify
        hashed_password: Hashed password to compare against

    Returns:
        True if password matches, False otherwise

    Raises:
        ValueError: If inputs are invalid
    """
    if not password or not isinstance(password, str):
        raise ValueError("Password must be a non-empty string")
    if not hashed_password or not isinstance(hashed_password, str):
        raise ValueError("Hashed password must be a non-empty string")

    try:
        return bcrypt.checkpw(password.encode("utf-8"), hashed_password.encode("utf-8"))
    except (ValueError, TypeError):
        # Invalid hash format
        return False


def create_token(
    user_id: str,
    email: str,
    expires_delta: timedelta = None
) -> str:
    """Create a signed JWT token with user information.

    Args:
        user_id: User ID to include in token
        email: User email to include in token
        expires_delta: Token expiration time delta (default: 1 hour)

    Returns:
        Signed JWT token string

    Raises:
        ValueError: If inputs are invalid
    """
    if not user_id or not isinstance(user_id, str):
        raise ValueError("user_id must be a non-empty string")
    if not email or not isinstance(email, str):
        raise ValueError("email must be a non-empty string")

    if expires_delta is None:
        expires_delta = timedelta(minutes=TOKEN_EXPIRE_MINUTES)

    # Use time.time() for consistency with JWT library
    iat_timestamp = int(time.time())
    exp_timestamp = iat_timestamp + int(expires_delta.total_seconds())

    payload = {
        "user_id": user_id,
        "email": email,
        "iat": iat_timestamp,
        "exp": exp_timestamp
    }

    token = jwt.encode(payload, SECRET_KEY, algorithm=ALGORITHM)
    return token


def create_expired_token(user_id: str, email: str) -> str:
    """Create an expired JWT token (mainly for testing).

    Args:
        user_id: User ID to include in token
        email: User email to include in token

    Returns:
        Expired JWT token string
    """
    # Create a token that expired 1 hour ago
    expires_delta = timedelta(minutes=-60)
    return create_token(user_id, email, expires_delta)


def verify_token(token: str) -> Dict[str, Any]:
    """Verify and decode a JWT token with signature validation.

    Validates:
    - Token signature (not tampered with)
    - Token expiration (not expired)
    - Token format (valid JWT structure)

    Args:
        token: JWT token to verify

    Returns:
        Decoded token payload as dictionary

    Raises:
        ValueError: If token is invalid or expired
        - "Token expired" - if token's exp claim is in the past
        - "Invalid token" - for other JWT validation failures
    """
    if not token or not isinstance(token, str):
        raise ValueError("Token must be a non-empty string")

    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        return payload
    except jwt.ExpiredSignatureError:
        raise ValueError("Token expired")
    except jwt.InvalidTokenError as e:
        raise ValueError(f"Invalid token: {str(e)}")
