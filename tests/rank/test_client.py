"""Tests for rank.client module."""

from unittest.mock import Mock, patch

import pytest
import requests

from moai_adk.rank.client import (
    ApiError,
    ApiStatus,
    AuthenticationError,
    LeaderboardEntry,
    RankClient,
    RankClientError,
    RankInfo,
    SessionSubmission,
    UserRank,
)
from moai_adk.rank.config import RankConfig


class TestDataclasses:
    """Test dataclass models."""

    def test_rank_info_creation(self):
        """Test RankInfo dataclass creation."""
        info = RankInfo(
            position=1,
            composite_score=95.5,
            total_participants=1000,
        )
        assert info.position == 1
        assert info.composite_score == 95.5
        assert info.total_participants == 1000

    def test_user_rank_creation(self):
        """Test UserRank dataclass creation."""
        daily = RankInfo(position=5, composite_score=85.0, total_participants=500)
        user_rank = UserRank(
            username="testuser",
            daily=daily,
            weekly=None,
            monthly=None,
            all_time=None,
            total_tokens=100000,
            total_sessions=50,
            input_tokens=60000,
            output_tokens=40000,
            last_updated="2024-01-01T00:00:00Z",
        )
        assert user_rank.username == "testuser"
        assert user_rank.daily.position == 5
        assert user_rank.total_tokens == 100000
        assert user_rank.last_updated == "2024-01-01T00:00:00Z"

    def test_leaderboard_entry_creation(self):
        """Test LeaderboardEntry dataclass creation."""
        entry = LeaderboardEntry(
            rank=1,
            username="champion",
            total_tokens=500000,
            composite_score=99.5,
            session_count=100,
            is_private=False,
        )
        assert entry.rank == 1
        assert entry.username == "champion"
        assert entry.is_private is False

    def test_api_status_creation(self):
        """Test ApiStatus dataclass creation."""
        status = ApiStatus(
            status="healthy",
            version="1.0.0",
            timestamp="2024-01-01T00:00:00Z",
        )
        assert status.status == "healthy"
        assert status.version == "1.0.0"

    def test_session_submission_creation(self):
        """Test SessionSubmission dataclass creation."""
        session = SessionSubmission(
            session_hash="abc123",
            ended_at="2024-01-01T00:00:00Z",
            input_tokens=1000,
            output_tokens=500,
            cache_creation_tokens=100,
            cache_read_tokens=50,
            model_name="claude-opus-4-5",
            started_at="2024-01-01T00:00:00Z",
            duration_seconds=3600,
            turn_count=10,
            tool_usage={"Read": 5, "Write": 3},
            model_usage={"claude-opus-4-5": {"input": 1000, "output": 500}},
            code_metrics={"linesAdded": 100, "linesDeleted": 20},
        )
        assert session.session_hash == "abc123"
        assert session.input_tokens == 1000
        assert session.model_name == "claude-opus-4-5"
        assert session.duration_seconds == 3600
        assert session.tool_usage["Read"] == 5


class TestRankClient:
    """Test RankClient class."""

    def test_initialization_with_api_key(self):
        """Test client initialization with explicit API key."""
        client = RankClient(api_key="test_key_123")
        assert client.api_key == "test_key_123"
        assert client.config.api_base_url == "https://rank.mo.ai.kr/api/v1"

    def test_initialization_with_custom_config(self):
        """Test client initialization with custom config."""
        config = RankConfig(base_url="https://custom.api")
        client = RankClient(config=config)
        assert client.config.base_url == "https://custom.api"

    def test_initialization_with_custom_timeout(self):
        """Test client initialization with custom timeout."""
        client = RankClient(timeout=60)
        assert client.timeout == 60

    def test_api_key_property(self):
        """Test api_key property getter and setter."""
        client = RankClient(api_key="initial_key")
        assert client.api_key == "initial_key"

        client.api_key = "new_key"
        assert client.api_key == "new_key"

    def test_compute_signature(self):
        """Test HMAC signature computation."""
        client = RankClient(api_key="secret_key")
        signature = client._compute_signature("1234567890", '{"test": "data"}')
        assert isinstance(signature, str)
        assert len(signature) == 64  # SHA256 hex length

    def test_compute_signature_without_api_key(self):
        """Test signature computation raises error without API key."""
        with patch("moai_adk.rank.client.RankConfig.get_api_key", return_value=None):
            client = RankClient(api_key=None)
            with pytest.raises(AuthenticationError, match="API key not configured"):
                client._compute_signature("1234567890", "{}")

    def test_get_auth_headers(self):
        """Test authentication headers generation."""
        client = RankClient(api_key="test_key")
        headers = client._get_auth_headers('{"test": "data"}')
        assert "X-API-Key" in headers
        assert "X-Timestamp" in headers
        assert "X-Signature" in headers
        assert headers["X-API-Key"] == "test_key"

    def test_get_auth_headers_without_api_key(self):
        """Test auth headers raises error without API key."""
        with patch("moai_adk.rank.client.RankConfig.get_api_key", return_value=None):
            client = RankClient(api_key=None)
            with pytest.raises(AuthenticationError, match="API key not configured"):
                client._get_auth_headers("{}")

    @patch("moai_adk.rank.client.requests.Session")
    def test_make_request_get_success(self, mock_session):
        """Test successful GET request."""
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = {"result": "success"}
        mock_session.return_value.get.return_value = mock_response

        client = RankClient(api_key="test_key")
        result = client._make_request("GET", "/test")

        assert result == {"result": "success"}

    @patch("moai_adk.rank.client.requests.Session")
    def test_make_request_post_success(self, mock_session):
        """Test successful POST request."""
        mock_response = Mock()
        mock_response.status_code = 201
        mock_response.json.return_value = {"id": "123"}
        mock_session.return_value.post.return_value = mock_response

        client = RankClient(api_key="test_key")
        result = client._make_request("POST", "/create", data={"name": "test"})

        assert result == {"id": "123"}

    @patch("moai_adk.rank.client.requests.Session")
    def test_make_request_authentication_error(self, mock_session):
        """Test authentication error handling."""
        mock_response = Mock()
        mock_response.status_code = 401
        mock_response.json.return_value = {"error": "Invalid credentials"}
        mock_session.return_value.get.return_value = mock_response

        client = RankClient(api_key="test_key")
        with pytest.raises(AuthenticationError, match="Invalid credentials"):
            client._make_request("GET", "/protected", auth=True)

    @patch("moai_adk.rank.client.requests.Session")
    def test_make_request_api_error(self, mock_session):
        """Test API error handling."""
        mock_response = Mock()
        mock_response.status_code = 400
        mock_response.json.return_value = {"error": "Bad request"}
        mock_session.return_value.get.return_value = mock_response

        client = RankClient(api_key="test_key")
        with pytest.raises(ApiError, match="Bad request"):
            client._make_request("GET", "/error")

    @patch("moai_adk.rank.client.requests.Session")
    def test_make_request_network_error(self, mock_session):
        """Test network error handling."""
        mock_session.return_value.get.side_effect = requests.ConnectionError("Network error")

        client = RankClient(api_key="test_key")
        with pytest.raises(RankClientError, match="Request failed"):
            client._make_request("GET", "/test")

    @patch("moai_adk.rank.client.requests.Session")
    def test_make_request_unsupported_method(self, mock_session):
        """Test unsupported HTTP method raises error."""
        client = RankClient(api_key="test_key")
        with pytest.raises(ValueError, match="Unsupported HTTP method"):
            client._make_request("DELETE", "/test")

    @patch("moai_adk.rank.client.requests.Session")
    def test_check_status(self, mock_session):
        """Test checking API status."""
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = {
            "status": "healthy",
            "version": "1.0.0",
            "timestamp": "2024-01-01T00:00:00Z",
        }
        mock_session.return_value.get.return_value = mock_response

        client = RankClient(api_key="test_key")
        status = client.check_status()

        assert status.status == "healthy"
        assert status.version == "1.0.0"

    @patch("moai_adk.rank.client.requests.Session")
    def test_get_user_rank(self, mock_session):
        """Test getting user rank information."""
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = {
            "success": True,
            "data": {
                "username": "testuser",
                "rankings": {
                    "daily": {
                        "position": 5,
                        "compositeScore": 85.0,
                        "totalParticipants": 500,
                    },
                    "weekly": None,
                },
                "stats": {
                    "totalTokens": 100000,
                    "totalSessions": 50,
                    "inputTokens": 60000,
                    "outputTokens": 40000,
                },
                "lastUpdated": "2024-01-01T00:00:00Z",
            },
        }
        mock_session.return_value.get.return_value = mock_response

        client = RankClient(api_key="test_key")
        user_rank = client.get_user_rank()

        assert user_rank.username == "testuser"
        assert user_rank.daily.position == 5
        assert user_rank.daily.composite_score == 85.0
        assert user_rank.total_tokens == 100000

    @patch("moai_adk.rank.client.requests.Session")
    def test_get_leaderboard(self, mock_session):
        """Test getting leaderboard."""
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = {
            "data": {
                "items": [
                    {
                        "rank": 1,
                        "username": "champion",
                        "totalTokens": 500000,
                        "compositeScore": 99.5,
                        "sessionCount": 100,
                        "isPrivate": False,
                    },
                    {
                        "rank": 2,
                        "username": "runner_up",
                        "totalTokens": 450000,
                        "compositeScore": 95.0,
                        "sessionCount": 90,
                        "isPrivate": True,
                    },
                ],
            },
        }
        mock_session.return_value.get.return_value = mock_response

        client = RankClient()
        leaderboard = client.get_leaderboard(period="weekly", limit=10)

        assert len(leaderboard) == 2
        assert leaderboard[0].rank == 1
        assert leaderboard[0].username == "champion"
        assert leaderboard[1].username == "runner_up"
        assert leaderboard[1].is_private is True

    @patch("moai_adk.rank.client.requests.Session")
    def test_submit_session(self, mock_session):
        """Test submitting a session."""
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = {
            "success": True,
            "sessionId": "sess_123",
        }
        mock_session.return_value.post.return_value = mock_response

        client = RankClient(api_key="test_key")
        session = SessionSubmission(
            session_hash="hash123",
            ended_at="2024-01-01T00:00:00Z",
            input_tokens=1000,
            output_tokens=500,
            model_name="claude-opus-4-5",
        )

        result = client.submit_session(session)
        assert result["success"] is True
        assert result["sessionId"] == "sess_123"

    @patch("moai_adk.rank.client.requests.Session")
    def test_submit_sessions_batch(self, mock_session):
        """Test batch session submission."""
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = {
            "success": True,
            "processed": 2,
            "succeeded": 2,
            "failed": 0,
        }
        mock_session.return_value.post.return_value = mock_response

        client = RankClient(api_key="test_key")
        sessions = [
            SessionSubmission(
                session_hash="hash1",
                ended_at="2024-01-01T00:00:00Z",
                input_tokens=1000,
                output_tokens=500,
            ),
            SessionSubmission(
                session_hash="hash2",
                ended_at="2024-01-01T01:00:00Z",
                input_tokens=800,
                output_tokens=400,
            ),
        ]

        result = client.submit_sessions_batch(sessions)
        assert result["processed"] == 2
        assert result["succeeded"] == 2

    def test_submit_sessions_batch_limit(self):
        """Test batch submission limit of 100 sessions."""
        client = RankClient(api_key="test_key")
        sessions = [
            SessionSubmission(
                session_hash=f"hash{i}",
                ended_at="2024-01-01T00:00:00Z",
                input_tokens=100,
                output_tokens=50,
            )
            for i in range(101)
        ]

        with pytest.raises(ValueError, match="Maximum 100 sessions"):
            client.submit_sessions_batch(sessions)

    def test_compute_session_hash(self):
        """Test session hash computation."""
        client = RankClient(api_key="test_key")
        hash_value = client.compute_session_hash(
            input_tokens=1000,
            output_tokens=500,
            cache_creation_tokens=100,
            cache_read_tokens=50,
            model_name="claude-opus-4-5",
            ended_at="2024-01-01T00:00:00Z",
        )

        assert isinstance(hash_value, str)
        assert len(hash_value) == 64  # SHA256 hex length

    def test_compute_session_hash_without_model(self):
        """Test session hash computation without model name."""
        client = RankClient(api_key="test_key")
        hash_value = client.compute_session_hash(
            input_tokens=500,
            output_tokens=250,
            cache_creation_tokens=0,
            cache_read_tokens=0,
            model_name=None,
            ended_at="2024-01-01T00:00:00Z",
        )

        assert isinstance(hash_value, str)
        assert len(hash_value) == 64


class TestRankClientErrors:
    """Test RankClient exception classes."""

    def test_rank_client_error(self):
        """Test base RankClientError."""
        error = RankClientError("Test error")
        assert str(error) == "Test error"

    def test_authentication_error(self):
        """Test AuthenticationError."""
        error = AuthenticationError("Auth failed")
        assert isinstance(error, RankClientError)
        assert str(error) == "Auth failed"

    def test_api_error(self):
        """Test ApiError with details."""
        details = {"field": "value"}
        error = ApiError("API failed", 400, details)
        assert isinstance(error, RankClientError)
        assert str(error) == "API failed"
        assert error.status_code == 400
        assert error.details == details
