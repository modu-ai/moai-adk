"""Tests for rank.auth module."""

import threading
from unittest.mock import Mock, patch

import requests

from moai_adk.rank.auth import OAuthCallbackHandler, OAuthHandler, OAuthState, verify_api_key
from moai_adk.rank.config import RankConfig


class TestOAuthState:
    """Test OAuthState dataclass."""

    def test_oauth_state_creation(self):
        """Test creating OAuth state with all fields."""
        state = OAuthState(
            state="test_state_123",
            redirect_port=8080,
            callback_received=True,
            auth_code="auth_code_456",
            api_key="api_key_789",
            username="testuser",
            user_id="user_001",
            created_at="2024-01-01T00:00:00Z",
            error=None,
        )
        assert state.state == "test_state_123"
        assert state.redirect_port == 8080
        assert state.callback_received is True
        assert state.api_key == "api_key_789"

    def test_oauth_state_defaults(self):
        """Test OAuth state with default values."""
        state = OAuthState(state="test", redirect_port=9000)
        assert state.callback_received is False
        assert state.auth_code is None
        assert state.api_key is None


class TestOAuthCallbackHandler:
    """Test OAuthCallbackHandler class variables and state management."""

    def test_oauth_callback_handler_class_variables(self):
        """Test that OAuthCallbackHandler has required class variables."""
        assert hasattr(OAuthCallbackHandler, "oauth_state")
        assert hasattr(OAuthCallbackHandler, "_on_complete")

    def test_oauth_state_modification(self):
        """Test OAuth state can be modified through class variable."""
        initial_state = OAuthState(state="initial", redirect_port=9000)
        OAuthCallbackHandler.oauth_state = initial_state

        assert OAuthCallbackHandler.oauth_state is initial_state
        assert OAuthCallbackHandler.oauth_state.state == "initial"

        # Update state
        OAuthCallbackHandler.oauth_state.callback_received = True
        assert OAuthCallbackHandler.oauth_state.callback_received is True

        # Clean up
        OAuthCallbackHandler.oauth_state = None

    def test_oauth_callback_state_persistence(self):
        """Test OAuth state persists across handler instances."""
        state = OAuthState(state="persistent_state", redirect_port=8080)
        OAuthCallbackHandler.oauth_state = state

        # Simulate callback updating state
        if OAuthCallbackHandler.oauth_state:
            OAuthCallbackHandler.oauth_state.api_key = "test_key_123"
            OAuthCallbackHandler.oauth_state.callback_received = True

        assert OAuthCallbackHandler.oauth_state.api_key == "test_key_123"
        assert OAuthCallbackHandler.oauth_state.callback_received is True

        # Clean up
        OAuthCallbackHandler.oauth_state = None


class TestOAuthHandler:
    """Test OAuthHandler class."""

    def test_oauth_flow_initialization(self):
        """Test OAuthHandler initialization."""
        flow = OAuthHandler()
        assert flow.config is not None
        assert flow._oauth_state is None
        assert flow._server is None

    def test_oauth_flow_with_config(self):
        """Test OAuthHandler with custom config."""
        config = RankConfig(base_url="https://test.api")
        flow = OAuthHandler(config=config)
        assert flow.config.base_url == "https://test.api"

    def test_find_free_port(self):
        """Test finding a free port."""
        flow = OAuthHandler()
        port = flow._find_free_port()
        assert 1024 <= port <= 65535

    def test_build_auth_url(self):
        """Test building authorization URL."""
        flow = OAuthHandler()
        url = flow._build_auth_url("test_state", "http://localhost:8080/callback")
        assert "test_state" in url
        # URL is encoded, so check for encoded version or components
        assert "redirect_uri=" in url
        assert "api/auth/cli" in url

    def test_oauth_flow_has_required_methods(self):
        """Test that OAuthHandler has required methods."""
        flow = OAuthHandler()
        # Verify the handler has the required methods
        assert hasattr(flow, "start_oauth_flow")
        assert hasattr(flow, "_find_free_port")
        assert hasattr(flow, "_build_auth_url")
        assert hasattr(flow, "_cleanup")

    def test_exchange_code_for_key_failure(self):
        """Test code exchange failure handling."""
        flow = OAuthHandler()
        result = flow._exchange_code_for_key("invalid_code", "http://localhost:8080/callback")
        assert result is None

    @patch("moai_adk.rank.auth.requests.get")
    def test_verify_api_key_success(self, mock_get):
        """Test successful API key verification."""
        mock_response = Mock()
        mock_response.status_code = 200
        mock_get.return_value = mock_response

        result = verify_api_key("test_key")
        assert result is True

    @patch("moai_adk.rank.auth.requests.get")
    def test_verify_api_key_failure(self, mock_get):
        """Test failed API key verification."""
        mock_response = Mock()
        mock_response.status_code = 401
        mock_get.return_value = mock_response

        result = verify_api_key("invalid_key")
        assert result is False

    @patch("moai_adk.rank.auth.requests.get")
    def test_verify_api_key_request_error(self, mock_get):
        """Test API key verification with request error."""
        mock_get.side_effect = requests.RequestException("Network error")

        result = verify_api_key("test_key")
        assert result is False

    def test_cleanup(self):
        """Test OAuthHandler cleanup."""
        flow = OAuthHandler()
        flow._server = Mock()
        flow._cleanup()
        # Should not raise exception
        assert flow._server is None


class TestOAuthHandlerRunServer:
    """Test OAuthHandler._run_server method."""

    @patch("moai_adk.rank.auth.socketserver.TCPServer")
    def test_run_server_no_server(self, mock_server):
        """Test _run_server with no server."""
        flow = OAuthHandler()
        flow._server = None

        event = threading.Event()
        flow._run_server(event, timeout=1)

        # Should return without error
        assert True

    @patch("moai_adk.rank.auth.socketserver.TCPServer")
    def test_run_server_until_event(self, mock_server):
        """Test _run_server stops when event is set."""
        mock_server_instance = Mock()
        mock_server_instance.handle_request = Mock()
        mock_server.return_value = mock_server_instance

        flow = OAuthHandler()
        flow._server = mock_server_instance

        event = threading.Event()

        # Stop server after a short delay
        def stop_server():
            import time

            time.sleep(0.1)
            event.set()

        stop_thread = threading.Thread(target=stop_server, daemon=True)
        stop_thread.start()

        flow._run_server(event, timeout=5)

        stop_thread.join(timeout=2)

        # Server should have been started
        assert mock_server_instance.handle_request.called


class TestOAuthHandlerExchangeCode:
    """Test OAuthHandler._exchange_code_for_key method."""

    @patch("moai_adk.rank.auth.requests.post")
    def test_exchange_code_success(self, mock_post):
        """Test successful code exchange."""
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = {"apiKey": "test_key"}
        mock_post.return_value = mock_response

        flow = OAuthHandler()
        result = flow._exchange_code_for_key("test_code", "http://localhost:8080/callback")

        assert result is not None
        assert result.api_key == "test_key"

    @patch("moai_adk.rank.auth.requests.post")
    def test_exchange_code_failure(self, mock_post):
        """Test code exchange with failed response."""
        mock_response = Mock()
        mock_response.status_code = 400
        mock_post.return_value = mock_response

        flow = OAuthHandler()
        result = flow._exchange_code_for_key("invalid_code", "http://localhost:8080/callback")

        assert result is None

    @patch("moai_adk.rank.auth.requests.post")
    def test_exchange_code_exception(self, mock_post):
        """Test code exchange with network exception."""
        mock_post.side_effect = requests.RequestException("Network error")

        flow = OAuthHandler()
        result = flow._exchange_code_for_key("test_code", "http://localhost:8080/callback")

        assert result is None
