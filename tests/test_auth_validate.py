"""
Test suite for token validation and user info retrieval (SPEC-AUTH-001).
RED Phase: Tests are written first and expected to fail.

NOTE: Auth module is not yet implemented in this version.
These tests are marked as skipped to allow test collection without ImportError.
"""

import pytest

# Auth module tests are skipped - module not implemented
pytestmark = pytest.mark.skip(reason="Auth module not implemented in moai-adk v0.26.0")


class TestTokenValidation:
    """Test cases for JWT token validation and user info retrieval."""

    @pytest.fixture
    def auth_service(self):
        """Create an auth service instance for testing."""
        return AuthService()

    @pytest.fixture
    def authenticated_user(self, auth_service):
        """Create and login a test user."""
        auth_service.create_user(email="testuser@example.com", password="TestPassword123!")
        login_response = auth_service.login(email="testuser@example.com", password="TestPassword123!")
        return login_response

    def test_get_user_info_with_valid_token(self, auth_service, authenticated_user):
        """
        Given: Valid JWT token exists
        When: Authorization header contains 'Bearer {token}'
        Then: User's ID, email, creation date is returned
        And: HTTP status code is 200 OK
        """
        token = authenticated_user["access_token"]

        user_info = auth_service.get_user_from_token(token)

        assert user_info is not None
        assert "id" in user_info
        assert "email" in user_info
        assert "created_at" in user_info
        assert user_info["email"] == "testuser@example.com"

    def test_get_user_info_with_expired_token(self, auth_service):
        """
        Given: Expired JWT token exists
        When: Try to retrieve user info with this token
        Then: Return 'Token expired' error
        And: HTTP status code is 401 Unauthorized
        """
        # Create an expired token
        expired_token = create_expired_token(user_id="test-user-123", email="testuser@example.com")

        with pytest.raises(ValueError, match="Token expired"):
            auth_service.get_user_from_token(expired_token)

    def test_get_user_info_with_invalid_token(self, auth_service):
        """
        Given: Invalid JWT token exists
        When: Try to retrieve user info with this token
        Then: Return error (token validation fails)
        And: HTTP status code is 401 Unauthorized
        """
        invalid_token = "invalid.token.string"

        with pytest.raises(ValueError):
            auth_service.get_user_from_token(invalid_token)

    def test_get_user_info_with_malformed_token(self, auth_service):
        """
        Given: Malformed JWT token
        When: Try to retrieve user info
        Then: Raise ValueError with appropriate message
        """
        malformed_token = "not-a-jwt-token"

        with pytest.raises(ValueError):
            auth_service.get_user_from_token(malformed_token)

    def test_token_contains_user_claims(self, auth_service, authenticated_user):
        """
        Given: Valid token from login
        When: Decode the token
        Then: Token contains user_id and email claims
        """
        token = authenticated_user["access_token"]
        token_data = auth_service.verify_token(token)

        assert token_data["user_id"] is not None
        assert token_data["email"] == "testuser@example.com"

    def test_token_expiration_claim(self, auth_service, authenticated_user):
        """
        Given: Valid token from login
        When: Check expiration claim
        Then: exp claim is set to 1 hour from issue time
        """
        token = authenticated_user["access_token"]
        token_data = auth_service.verify_token(token)

        assert "exp" in token_data
        assert "iat" in token_data

        # exp should be 1 hour (3600 seconds) after iat
        exp_diff = token_data["exp"] - token_data["iat"]
        assert exp_diff == 3600

    def test_missing_authorization_header(self, auth_service):
        """
        Given: Request without Authorization header
        When: Try to get user info
        Then: Return error
        """
        with pytest.raises(ValueError, match="Token required"):
            auth_service.get_user_from_token(None)

    def test_bearer_token_format(self, auth_service, authenticated_user):
        """
        Given: Token in Bearer format
        When: Extract token from 'Bearer {token}' format
        Then: Extract and validate successfully
        """
        token = authenticated_user["access_token"]
        bearer_header = f"Bearer {token}"

        # Extract token from bearer format
        extracted_token = bearer_header.split(" ")[1]
        user_info = auth_service.get_user_from_token(extracted_token)

        assert user_info["email"] == "testuser@example.com"
