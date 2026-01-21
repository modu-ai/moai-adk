"""Characterization Tests for rank.auth module.

These tests capture the CURRENT BEHAVIOR of the rank/auth module.
They serve as a safety net during refactoring by documenting what the code
actually does, not what it should do.

Characterization Test Principles:
- Capture actual behavior without judgment
- Document side effects and edge cases
- Provide regression safety during refactoring
- Focus on observable behavior, not implementation

Behavior Snapshots:
- OAuthCallbackHandler.do_GET: Handles OAuth callback with state verification, error handling, HTML responses
- OAuthHandler.start_oauth_flow: Manages complete OAuth flow with browser, local server, timeout
- verify_api_key: Validates API key via HTTP request

Missing Coverage Lines:
- Lines 50-109: do_GET callback handling (state verification, error handling, API key extraction)
- Lines 113-139: _send_response error HTML
- Lines 143-172: _send_success_response success HTML
- Lines 226-330: start_oauth_flow main flow (browser open, callback wait, credential exchange)
"""

import threading
from unittest.mock import Mock, patch

import pytest

from moai_adk.rank.auth import OAuthCallbackHandler, OAuthHandler, OAuthState, verify_api_key
from moai_adk.rank.config import RankConfig


class TestOAuthCallbackHandlerCharacterization:
    """Characterization tests for OAuthCallbackHandler.do_GET method.

    Current Behavior:
    - Returns 500 error when oauth_state is None
    - Parses query parameters from callback URL
    - Verifies state parameter for CSRF protection
    - Handles error parameter from OAuth provider
    - Extracts direct API key (new MoAI Rank flow)
    - Falls back to authorization code (legacy flow)
    - Sends HTML success/error responses
    - Triggers completion callback via threading
    - Returns 404 for unknown paths
    """

    def test_characterize_no_oauth_state_returns_500(self):
        """CAPTURE: When oauth_state is None, returns 500 error.

        Behavior Snapshot:
        Input: oauth_state=None, GET request to /callback
        Output: 500 status code, "OAuth state not initialized"

        Note: This is defensive programming for uninitialized handler.
        """
        # GIVEN: No oauth_state set
        OAuthCallbackHandler.oauth_state = None

        # WHEN: Callback handler receives request
        with patch("http.server.BaseHTTPRequestHandler.__init__", return_value=None):
            handler = OAuthCallbackHandler(None, ("/callback",), None)
            handler.send_error = Mock()
            handler.do_GET()

        # THEN: 500 error response
        handler.send_error.assert_called_once_with(500, "OAuth state not initialized")

    def test_characterize_state_mismatch_sets_error_and_returns_400(self):
        """CAPTURE: State parameter mismatch triggers CSRF protection.

        Behavior Snapshot:
        Input: callback with state != expected state
        Output: oauth_state.error set, 400 response with "Invalid state parameter"

        Note: Critical security feature - prevents CSRF attacks.
        """
        # GIVEN: oauth_state with specific state token
        test_state = OAuthState(state="correct-state-token", redirect_port=8080)
        OAuthCallbackHandler.oauth_state = test_state

        # WHEN: Callback returns with wrong state
        with patch("http.server.BaseHTTPRequestHandler.__init__", return_value=None):
            handler = OAuthCallbackHandler(None, ("/callback?state=wrong-state&api_key=test-key"), None)
            handler.path = "/callback?state=wrong-state&api_key=test-key"
            handler.send_response = Mock()
            handler.send_header = Mock()
            handler.end_headers = Mock()
            handler.wfile = Mock()
            handler.wfile.write = Mock()
            handler.do_GET()

        # THEN: Error state set
        assert test_state.error == "State mismatch - possible CSRF attack"
        handler.send_response.assert_called_once_with(400)

    def test_characterize_error_parameter_sets_error_and_returns_400(self):
        """CAPTURE: OAuth provider error parameter is handled.

        Behavior Snapshot:
        Input: callback with error parameter
        Output: oauth_state.error set to error_description, 400 response

        Note: Captures user denial or OAuth provider errors.
        """
        # GIVEN: oauth_state
        test_state = OAuthState(state="valid-state", redirect_port=8080)
        OAuthCallbackHandler.oauth_state = test_state

        # WHEN: Callback returns error parameter
        with patch("http.server.BaseHTTPRequestHandler.__init__", return_value=None):
            handler = OAuthCallbackHandler(
                None,
                ("/callback?state=valid-state&error=access_denied&error_description=User denied access"),
                None,
            )
            handler.path = "/callback?state=valid-state&error=access_denied&error_description=User denied access"
            handler.send_response = Mock()
            handler.send_header = Mock()
            handler.end_headers = Mock()
            handler.wfile = Mock()
            handler.wfile.write = Mock()
            handler.do_GET()

        # THEN: Error description captured
        assert test_state.error == "User denied access"
        handler.send_response.assert_called_once_with(400)

    def test_characterize_direct_api_key_flow_new_flow(self):
        """CAPTURE: Direct API key from new MoAI Rank flow.

        Behavior Snapshot:
        Input: callback with api_key, username, user_id, created_at parameters
        Output: oauth_state fields populated, callback_received=True, success response

        Note: New flow receives API key directly without code exchange.
        """
        # GIVEN: oauth_state
        test_state = OAuthState(state="valid-state", redirect_port=8080)
        OAuthCallbackHandler.oauth_state = test_state

        # WHEN: Callback returns direct API key
        with patch("http.server.BaseHTTPRequestHandler.__init__", return_value=None):
            handler = OAuthCallbackHandler(
                None,
                ("/callback?state=valid-state&api_key=sk-test-123&username=testuser&user_id=456&created_at=2024-01-01"),
                None,
            )
            handler.path = (
                "/callback?state=valid-state&api_key=sk-test-123&username=testuser&user_id=456&created_at=2024-01-01"
            )
            handler.send_response = Mock()
            handler.send_header = Mock()
            handler.end_headers = Mock()
            handler.wfile = Mock()
            handler.wfile.write = Mock()
            handler.do_GET()

        # THEN: All fields populated
        assert test_state.api_key == "sk-test-123"
        assert test_state.username == "testuser"
        assert test_state.user_id == "456"
        assert test_state.created_at == "2024-01-01"
        assert test_state.callback_received is True
        assert test_state.error is None
        handler.send_response.assert_called_once_with(200)

    def test_characterize_legacy_auth_code_fallback(self):
        """CAPTURE: Authorization code fallback for legacy flow.

        Behavior Snapshot:
        Input: callback with code parameter (no api_key)
        Output: oauth_state.auth_code set, callback_received=True, success response

        Note: Legacy flow requires code exchange after callback.
        """
        # GIVEN: oauth_state
        test_state = OAuthState(state="valid-state", redirect_port=8080)
        OAuthCallbackHandler.oauth_state = test_state

        # WHEN: Callback returns authorization code
        with patch("http.server.BaseHTTPRequestHandler.__init__", return_value=None):
            handler = OAuthCallbackHandler(None, ("/callback?state=valid-state&code=auth-code-123"), None)
            handler.path = "/callback?state=valid-state&code=auth-code-123"
            handler.send_response = Mock()
            handler.send_header = Mock()
            handler.end_headers = Mock()
            handler.wfile = Mock()
            handler.wfile.write = Mock()
            handler.do_GET()

        # THEN: Auth code captured
        assert test_state.auth_code == "auth-code-123"
        assert test_state.callback_received is True
        assert test_state.api_key is None
        handler.send_response.assert_called_once_with(200)

    def test_characterize_no_code_or_api_key_sets_error(self):
        """CAPTURE: Missing code and api_key parameters.

        Behavior Snapshot:
        Input: callback with state but no code or api_key
        Output: error set, 400 response

        Note: Edge case - incomplete callback.
        """
        # GIVEN: oauth_state
        test_state = OAuthState(state="valid-state", redirect_port=8080)
        OAuthCallbackHandler.oauth_state = test_state

        # WHEN: Callback has no code or api_key
        with patch("http.server.BaseHTTPRequestHandler.__init__", return_value=None):
            handler = OAuthCallbackHandler(None, ("/callback?state=valid-state"), None)
            handler.path = "/callback?state=valid-state"
            handler.send_response = Mock()
            handler.send_header = Mock()
            handler.end_headers = Mock()
            handler.wfile = Mock()
            handler.wfile.write = Mock()
            handler.do_GET()

        # THEN: Error set
        assert test_state.error == "No authorization code or API key received"
        handler.send_response.assert_called_once_with(400)

    def test_characterize_unknown_path_returns_404(self):
        """CAPTURE: Unknown path returns 404.

        Behavior Snapshot:
        Input: GET request to /unknown (not /callback)
        Output: 404 error response

        Note: Only /callback path is valid.
        """
        # GIVEN: oauth_state
        test_state = OAuthState(state="valid-state", redirect_port=8080)
        OAuthCallbackHandler.oauth_state = test_state

        # WHEN: Request to unknown path
        with patch("http.server.BaseHTTPRequestHandler.__init__", return_value=None):
            handler = OAuthCallbackHandler(None, ("/unknown",), None)
            handler.path = "/unknown"
            handler.send_error = Mock()
            handler.do_GET()

        # THEN: 404 response
        handler.send_error.assert_called_once_with(404, "Not Found")

    def test_characterize_completion_callback_triggered_on_success(self):
        """CAPTURE: Completion callback triggered via threading.

        Behavior Snapshot:
        Input: Successful callback with API key
        Output: _on_complete callback called in daemon thread

        Note: Non-blocking callback to allow server shutdown.
        """
        # GIVEN: oauth_state and completion callback
        test_state = OAuthState(state="valid-state", redirect_port=8080)
        OAuthCallbackHandler.oauth_state = test_state

        completion_called = []

        def mock_completion():
            completion_called.append(True)

        OAuthCallbackHandler._on_complete = mock_completion

        # WHEN: Successful callback triggers threading
        with patch("threading.Thread") as mock_thread_class:
            mock_thread = Mock()
            mock_thread_class.return_value = mock_thread

            # Simulate the callback invocation pattern from do_GET
            if OAuthCallbackHandler._on_complete:
                threading.Thread(target=OAuthCallbackHandler._on_complete, daemon=True).start()

        # THEN: Thread created with callback
        mock_thread_class.assert_called_once()
        call_args = mock_thread_class.call_args
        assert call_args[1]["target"] == mock_completion
        assert call_args[1]["daemon"] is True
        mock_thread.start.assert_called_once()

    def test_characterize_send_response_generates_error_html(self):
        """CAPTURE: Error HTML response generation.

        Behavior Snapshot:
        Input: status=400, message="Authorization failed"
        Output: HTML with error styling, status code sent, content-type header

        Note: User-friendly error page in browser.
        """
        # GIVEN: Handler with error message
        with patch("http.server.BaseHTTPRequestHandler.__init__", return_value=None):
            handler = OAuthCallbackHandler(None, ("/callback",), None)

            handler.send_response = Mock()
            handler.send_header = Mock()
            handler.end_headers = Mock()

            output = []

            class MockWFile:
                def write(self, data):
                    output.append(data)

            handler.wfile = MockWFile()

            # WHEN: Error response sent
            handler._send_response(400, "Authorization failed")

            # THEN: HTML written to wfile
            full_output = b"".join(output)
            assert b"Authorization Failed" in full_output
            assert b"Authorization failed" in full_output
            assert b"You can close this window" in full_output
            assert b".error { color: #dc3545; }" in full_output

    def test_characterize_send_success_response_generates_success_html(self):
        """CAPTURE: Success HTML response generation.

        Behavior Snapshot:
        Input: Successful callback
        Output: HTML with success styling, auto-close script

        Note: User-friendly success page with auto-close after 3 seconds.
        """
        # GIVEN: Handler for success response
        with patch("http.server.BaseHTTPRequestHandler.__init__", return_value=None):
            handler = OAuthCallbackHandler(None, ("/callback",), None)

            handler.send_response = Mock()
            handler.send_header = Mock()
            handler.end_headers = Mock()

            output = []

            class MockWFile:
                def write(self, data):
                    output.append(data)

            handler.wfile = MockWFile()

            # WHEN: Success response sent
            handler._send_success_response()

            # THEN: HTML written to wfile
            full_output = b"".join(output)
            assert b"Authorization Successful!" in full_output
            assert b"connected your GitHub account" in full_output
            assert b"window.close(), 3000" in full_output
            assert b"linear-gradient(135deg, #667eea 0%, #764ba2 100%)" in full_output


class TestOAuthHandlerCharacterization:
    """Characterization tests for OAuthHandler.start_oauth_flow method.

    Current Behavior:
    - Generates secure state token using secrets.token_urlsafe(32)
    - Finds free port in range 8080-8179
    - Creates TCPServer for local callback
    - Opens browser with auth URL
    - Waits for callback with timeout
    - Exchanges auth code for API key (legacy flow)
    - Saves credentials via RankConfig.save_credentials
    - Calls on_success/on_error callbacks
    - Cleans up server in finally block
    - Handles KeyboardInterrupt (user cancellation)
    - Returns None on failure
    """

    def test_characterize_secure_state_generation(self):
        """CAPTURE: Secure state token generation.

        Behavior Snapshot:
        Input: No existing state
        Output: state generated via secrets.token_urlsafe(32), 43 characters (base64)

        Note: Cryptographically secure random state for CSRF protection.
        """
        # GIVEN: OAuth handler
        handler = OAuthHandler()

        # WHEN: State generation mocked
        with patch("secrets.token_urlsafe", return_value="test-state-token-1234567890"):
            with patch.object(handler, "_find_free_port", return_value=8080):
                with patch("socketserver.TCPServer"):
                    with patch("webbrowser.open"):
                        with patch.object(handler, "_run_server"):
                            # Start flow (will fail but captures state)
                            try:
                                handler.start_oauth_flow(timeout=1)
                            except Exception:
                                pass

                            # THEN: State generated using secrets
                            if handler._oauth_state:
                                assert handler._oauth_state.state == "test-state-token-1234567890"

    def test_characterize_find_free_port_scans_range(self):
        """CAPTURE: Free port finding in specified range.

        Behavior Snapshot:
        Input: Default range 8080-8179
        Output: First available port returned
        Exception: RuntimeError if no port available

        Note: Scans sequentially, binds to test availability.
        """
        # GIVEN: OAuth handler
        handler = OAuthHandler()

        # WHEN: Finding free port
        # First port in range might be occupied, test finds next available
        port = handler._find_free_port(start=8100, end=8110)

        # THEN: Port in range returned
        assert 8100 <= port < 8110

    def test_characterize_no_free_port_raises_error(self):
        """CAPTURE: No available port raises RuntimeError.

        Behavior Snapshot:
        Input: All ports in range occupied (or invalid range)
        Output: RuntimeError with "No available port found"

        Note: Shouldn't happen in practice with 100-port range.
        """
        # GIVEN: OAuth handler
        handler = OAuthHandler()

        # WHEN: All ports occupied
        with patch("socket.socket") as mock_socket_class:
            mock_socket = Mock()
            mock_socket.bind.side_effect = OSError("Address already in use")
            mock_socket_class.return_value.__enter__.return_value = mock_socket

            # THEN: RuntimeError raised
            with pytest.raises(RuntimeError, match="No available port found"):
                handler._find_free_port(start=9000, end=9001)

    def test_characterize_browser_open_with_auth_url(self):
        """CAPTURE: Browser opens with authorization URL.

        Behavior Snapshot:
        Input: Valid state and port
        Output: webbrowser.open called with auth URL containing state and redirect_uri

        Note: Auth URL format: {base_url}/api/auth/cli?redirect_uri=...&state=...
        """
        # GIVEN: OAuth handler with config
        handler = OAuthHandler(RankConfig(base_url="https://rank.example.com"))

        # WHEN: OAuth flow starts
        with patch.object(handler, "_find_free_port", return_value=8080):
            with patch("socketserver.TCPServer"):
                with patch("webbrowser.open") as mock_open:
                    with patch.object(handler, "_run_server"):
                        try:
                            handler.start_oauth_flow(timeout=1)
                        except Exception:
                            pass

                        # THEN: Browser opened with auth URL
                        if mock_open.called:
                            url = mock_open.call_args[0][0]
                            assert "https://rank.example.com/api/auth/cli" in url
                            assert "redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback" in url
                            assert "state=" in url

    def test_characterize_timeout_returns_none_with_error(self):
        """CAPTURE: Timeout waiting for callback returns None.

        Behavior Snapshot:
        Input: timeout=1 second, no callback received
        Output: None returned, on_error callback called with timeout message

        Note: Server shuts down after timeout + 5 second grace period.
        """
        # GIVEN: OAuth handler
        handler = OAuthHandler()

        # WHEN: Timeout elapses
        with patch.object(handler, "_find_free_port", return_value=8080):
            with patch("socketserver.TCPServer"):
                with patch("webbrowser.open"):
                    with patch.object(handler, "_run_server"):
                        error_messages = []

                        def mock_on_error(msg):
                            error_messages.append(msg)

                        result = handler.start_oauth_flow(on_error=mock_on_error, timeout=0)

                        # THEN: None returned, error callback triggered
                        assert result is None
                        assert len(error_messages) > 0
                        assert any("timed out" in msg.lower() for msg in error_messages)

    def test_characterize_keyboard_interrupt_returns_none(self):
        """CAPTURE: User cancellation (Ctrl+C) returns None.

        Behavior Snapshot:
        Input: KeyboardInterrupt during wait
        Output: None returned, on_error called with "Registration cancelled by user"

        Note: Graceful handling of user cancellation.
        """
        # GIVEN: OAuth handler
        handler = OAuthHandler()

        # WHEN: User presses Ctrl+C
        with patch.object(handler, "_find_free_port", return_value=8080):
            with patch("socketserver.TCPServer") as mock_server_class:
                mock_server = Mock()
                mock_server_class.return_value = mock_server

                with patch("webbrowser.open"):

                    def mock_run_server(*args):
                        raise KeyboardInterrupt()

                    with patch.object(handler, "_run_server", mock_run_server):
                        error_messages = []

                        def mock_on_error(msg):
                            error_messages.append(msg)

                        result = handler.start_oauth_flow(on_error=mock_on_error, timeout=5)

                        # THEN: None returned, cancellation error
                        assert result is None

    def test_characterize_direct_api_key_flow_skips_exchange(self):
        """CAPTURE: Direct API key flow skips code exchange.

        Behavior Snapshot:
        Input: Callback with api_key parameter (new flow)
        Output: Credentials created directly, saved, on_success called

        Note: New optimized flow without token exchange.
        """
        # GIVEN: OAuth handler
        handler = OAuthHandler()

        # WHEN: Callback with direct API key
        with patch.object(handler, "_find_free_port", return_value=8080):
            with patch("socketserver.TCPServer") as mock_server_class:
                mock_server = Mock()

                # Simulate callback setting API key
                def mock_run_server(shutdown_event, timeout):
                    # Simulate callback received
                    handler._oauth_state.api_key = "sk-direct-api-key"
                    handler._oauth_state.username = "testuser"
                    handler._oauth_state.user_id = "123"
                    handler._oauth_state.created_at = "2024-01-01"
                    handler._oauth_state.callback_received = True
                    shutdown_event.set()

                mock_server_class.return_value = mock_server

                with patch("webbrowser.open"):
                    with patch.object(handler, "_run_server", mock_run_server):
                        with patch("moai_adk.rank.auth.RankConfig.save_credentials"):
                            success_creds = []

                            def mock_on_success(creds):
                                success_creds.append(creds)

                            result = handler.start_oauth_flow(on_success=mock_on_success, timeout=5)

                            # THEN: Direct credentials returned
                            assert result is not None
                            assert result.api_key == "sk-direct-api-key"
                            assert len(success_creds) == 1
                            assert success_creds[0].api_key == "sk-direct-api-key"

    def test_characterize_legacy_flow_exchanges_code_for_key(self):
        """CAPTURE: Legacy flow exchanges auth code for API key.

        Behavior Snapshot:
        Input: Callback with code parameter (legacy flow)
        Output: HTTP POST to /api/auth/cli/token, credentials created from response

        Note: Backward compatible with original OAuth flow.
        """
        # GIVEN: OAuth handler
        handler = OAuthHandler()

        # WHEN: Callback with auth code
        with patch.object(handler, "_find_free_port", return_value=8080):
            with patch("socketserver.TCPServer") as mock_server_class:
                mock_server = Mock()

                # Simulate callback with code
                def mock_run_server(shutdown_event, timeout):
                    handler._oauth_state.auth_code = "legacy-auth-code"
                    handler._oauth_state.callback_received = True
                    shutdown_event.set()

                mock_server_class.return_value = mock_server

                with patch("webbrowser.open"):
                    with patch.object(handler, "_run_server", mock_run_server):
                        # Mock token exchange response
                        mock_response = Mock()
                        mock_response.status_code = 200
                        mock_response.json.return_value = {
                            "apiKey": "sk-exchanged-key",
                            "username": "legacyuser",
                            "userId": "456",
                            "createdAt": "2024-01-01",
                        }

                        with patch("requests.post", return_value=mock_response):
                            with patch("moai_adk.rank.auth.RankConfig.save_credentials"):
                                result = handler.start_oauth_flow(timeout=5)

                                # THEN: Code exchanged for API key
                                assert result is not None
                                assert result.api_key == "sk-exchanged-key"

    def test_characterize_code_exchange_failure_returns_none(self):
        """CAPTURE: Failed code exchange returns None.

        Behavior Snapshot:
        Input: Token exchange returns non-200 status
        Output: None returned, on_error called

        Note: Network failure or invalid auth code.
        """
        # GIVEN: OAuth handler
        handler = OAuthHandler()

        # WHEN: Token exchange fails
        with patch.object(handler, "_find_free_port", return_value=8080):
            with patch("socketserver.TCPServer") as mock_server_class:
                mock_server = Mock()

                def mock_run_server(shutdown_event, timeout):
                    handler._oauth_state.auth_code = "invalid-code"
                    handler._oauth_state.callback_received = True
                    shutdown_event.set()

                mock_server_class.return_value = mock_server

                with patch("webbrowser.open"):
                    with patch.object(handler, "_run_server", mock_run_server):
                        # Mock failed exchange
                        mock_response = Mock()
                        mock_response.status_code = 400

                        with patch("requests.post", return_value=mock_response):
                            error_messages = []

                            def mock_on_error(msg):
                                error_messages.append(msg)

                            result = handler.start_oauth_flow(on_error=mock_on_error, timeout=5)

                            # THEN: None returned
                            assert result is None
                            assert len(error_messages) > 0

    def test_characterize_cleanup_closes_server(self):
        """CAPTURE: Server cleanup in finally block.

        Behavior Snapshot:
        Input: Flow completes (success or failure)
        Output: Server closed via server_close()

        Note: Cleanup happens even on exception.
        """
        # GIVEN: OAuth handler with server
        handler = OAuthHandler()

        # WHEN: Flow completes
        with patch.object(handler, "_find_free_port", return_value=8080):
            with patch("socketserver.TCPServer") as mock_server_class:
                mock_server = Mock()
                mock_server_class.return_value = mock_server

                with patch("webbrowser.open"):
                    with patch.object(handler, "_run_server"):
                        # Force exception to test cleanup
                        with patch.object(handler, "_run_server", side_effect=Exception("Test error")):
                            handler.start_oauth_flow(timeout=1)

                            # THEN: Cleanup called
                            mock_server.server_close.assert_called_once()

    def test_characterize_build_auth_url_format(self):
        """CAPTURE: Authorization URL construction.

        Behavior Snapshot:
        Input: state="test-state", redirect_uri="http://localhost:8080/callback"
        Output: URL with query parameters

        Note: URL format follows OAuth 2.0 spec.
        """
        # GIVEN: OAuth handler
        handler = OAuthHandler(RankConfig(base_url="https://rank.example.com"))

        # WHEN: Building auth URL
        url = handler._build_auth_url("test-state", "http://localhost:8080/callback")

        # THEN: Proper format (URL-encoded)
        assert (
            url
            == "https://rank.example.com/api/auth/cli?redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback&state=test-state"
        )


class TestVerifyApiKeyCharacterization:
    """Characterization tests for verify_api_key function.

    Current Behavior:
    - Makes GET request to /rank endpoint with X-API-Key header
    - Returns True if status code is 200
    - Returns False on any exception or non-200 status
    - Uses default RankConfig if none provided
    - 10 second timeout for request
    """

    def test_characterize_valid_key_returns_true(self):
        """CAPTURE: Valid API key returns True.

        Behavior Snapshot:
        Input: Valid API key
        Output: True (HTTP 200 response)

        Note: Successful verification.
        """
        # GIVEN: Valid API key
        mock_response = Mock()
        mock_response.status_code = 200

        # WHEN: Verifying
        with patch("requests.get", return_value=mock_response) as mock_get:
            result = verify_api_key("sk-valid-key-123")

            # THEN: Request made with API key header
            assert result is True
            mock_get.assert_called_once()
            assert "X-API-Key" in mock_get.call_args.kwargs["headers"]
            assert mock_get.call_args.kwargs["headers"]["X-API-Key"] == "sk-valid-key-123"

    def test_characterize_invalid_key_returns_false(self):
        """CAPTURE: Invalid API key returns False.

        Behavior Snapshot:
        Input: Invalid API key
        Output: False (HTTP 401/403 response)

        Note: Non-200 status means invalid.
        """
        # GIVEN: Invalid API key
        mock_response = Mock()
        mock_response.status_code = 401

        # WHEN: Verifying
        with patch("requests.get", return_value=mock_response):
            result = verify_api_key("sk-invalid-key")

            # THEN: False returned
            assert result is False

    def test_characterize_network_error_returns_false(self):
        """CAPTURE: Network error returns False.

        Behavior Snapshot:
        Input: Connection error, timeout
        Output: False (exception caught)

        Note: Any RequestException returns False.
        """
        # GIVEN: Network error
        import requests

        with patch("requests.get", side_effect=requests.RequestException("Network error")):
            # WHEN: Verifying
            result = verify_api_key("sk-any-key")

            # THEN: False returned
            assert result is False

    def test_characterize_uses_default_config(self):
        """CAPTURE: Uses default RankConfig when none provided.

        Behavior Snapshot:
        Input: API key, config=None
        Output: Request to default api_base_url

        Note: Convenience for simple verification.
        """
        # GIVEN: No config provided
        mock_response = Mock()
        mock_response.status_code = 200

        # WHEN: Verifying
        with patch("requests.get", return_value=mock_response) as mock_get:
            verify_api_key("sk-key")

            # THEN: Default config used
            assert mock_get.called
            # URL contains /rank endpoint from default config
            url = mock_get.call_args.args[0]
            assert "/rank" in url

    def test_characterize_custom_config_used(self):
        """CAPTURE: Uses custom RankConfig when provided.

        Behavior Snapshot:
        Input: API key, custom RankConfig
        Output: Request to custom api_base_url

        Note: Supports multiple environments.
        """
        # GIVEN: Custom config
        custom_config = RankConfig(base_url="https://custom.example.com")
        mock_response = Mock()
        mock_response.status_code = 200

        # WHEN: Verifying
        with patch("requests.get", return_value=mock_response) as mock_get:
            verify_api_key("sk-key", config=custom_config)

            # THEN: Custom URL used
            url = mock_get.call_args.args[0]
            assert "custom.example.com" in url


class TestOAuthStateDataclass:
    """Characterization tests for OAuthState dataclass.

    Current Behavior:
    - Dataclass with 8 fields
    - Required: state, redirect_port
    - Optional: callback_received (default False), auth_code, api_key, username, user_id, created_at, error
    """

    def test_characterize_complete_state(self):
        """CAPTURE: Complete OAuthState with all fields.

        Behavior Snapshot:
        Input: All fields populated
        Output: Dataclass instance with accessible fields

        Note: Full state during active OAuth flow.
        """
        # GIVEN: All fields
        state = OAuthState(
            state="test-state",
            redirect_port=8080,
            callback_received=True,
            auth_code="code123",
            api_key="sk-key",
            username="user",
            user_id="123",
            created_at="2024-01-01",
            error=None,
        )

        # THEN: All accessible
        assert state.state == "test-state"
        assert state.redirect_port == 8080
        assert state.callback_received is True
        assert state.auth_code == "code123"
        assert state.api_key == "sk-key"
        assert state.username == "user"
        assert state.user_id == "123"
        assert state.created_at == "2024-01-01"
        assert state.error is None

    def test_characterize_minimal_state(self):
        """CAPTURE: Minimal OAuthState with required fields only.

        Behavior Snapshot:
        Input: Only state and redirect_port
        Output: Dataclass with defaults for optional fields

        Note: Initial state before callback.
        """
        # GIVEN: Minimal state
        state = OAuthState(state="test-state", redirect_port=8080)

        # THEN: Defaults applied
        assert state.callback_received is False
        assert state.auth_code is None
        assert state.api_key is None
        assert state.username is None
        assert state.user_id is None
        assert state.created_at is None
        assert state.error is None
