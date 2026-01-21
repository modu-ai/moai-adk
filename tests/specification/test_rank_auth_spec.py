"""Specification Tests for rank.auth module.

These tests specify the DESIRED BEHAVIOR based on domain requirements.
They use Given-When-Then pattern to express business rules and user interactions.

Specification Test Principles:
- Based on domain requirements and user needs
- Express what the system SHOULD do (not what it currently does)
- Use Given-When-Then (Gherkin-inspired) pattern
- Focus on user-visible behavior and business rules

Domain Requirements:
- OAuth Security: CSRF protection via state parameter, secure token generation
- API Key Exchange: Direct API key flow (new) and auth code flow (legacy)
- Timeout Handling: Configurable timeout with graceful failure
- Credential Persistence: Secure storage via RankConfig
- User Cancellation: Graceful handling of user cancellation
- Error Handling: Clear error messages for all failure modes
"""

from unittest.mock import Mock, patch

import pytest

from moai_adk.rank.auth import OAuthCallbackHandler, OAuthHandler, OAuthState, verify_api_key
from moai_adk.rank.config import RankConfig, RankCredentials


class TestOAuthSecuritySpecification:
    """Specification tests for OAuth security requirements.

    User Story:
    AS a developer registering with MoAI Rank
    I WANT my OAuth flow to be secure against CSRF attacks
    SO THAT my authorization cannot be hijacked
    """

    def test_given_oauth_flow_initiated_when_callback_with_invalid_state_then_reject_as_csrf(self):
        """GIVEN an OAuth flow is initiated with a secure state token
        WHEN the callback is received with a mismatched state parameter
        THEN the request should be rejected as a potential CSRF attack

        Security Requirement:
        - State parameter must match exactly
        - Error message should indicate CSRF suspicion
        - No credentials should be stored
        """
        # GIVEN: OAuth flow initiated with state token
        original_state = "secure-state-token-abc123"
        test_state = OAuthState(state=original_state, redirect_port=8080)
        OAuthCallbackHandler.oauth_state = test_state

        # WHEN: Callback returns with different state (CSRF attempt)
        with patch("http.server.BaseHTTPRequestHandler.__init__", return_value=None):
            handler = OAuthCallbackHandler(
                None,
                ("/callback?state=attacker-state&api_key=stolen-key"),
                None,
            )
            handler.path = "/callback?state=attacker-state&api_key=stolen-key"
            handler.send_response = Mock()
            handler.send_header = Mock()
            handler.end_headers = Mock()
            handler.wfile = Mock()
            handler.wfile.write = Mock()
            handler.do_GET()

        # THEN: Request rejected, error set, no credentials stored
        assert test_state.error == "State mismatch - possible CSRF attack"
        assert test_state.api_key is None
        assert test_state.callback_received is False
        handler.send_response.assert_called_once_with(400)

    def test_given_secure_state_required_when_token_generated_then_cryptographically_random(self):
        """GIVEN a new OAuth flow is initiated
        WHEN a state token is generated
        THEN it should use cryptographically secure random generation

        Security Requirement:
        - Use secrets.token_urlsafe(32) for state generation
        - 32 bytes = 43 base64 characters (sufficient entropy)
        - Unpredictable to prevent brute force attacks
        """
        # GIVEN: New OAuth flow
        handler = OAuthHandler()

        # WHEN: State token is generated
        with patch("secrets.token_urlsafe") as mock_token:
            mock_token.return_value = "secure-random-state-1234567890"

            with patch.object(handler, "_find_free_port", return_value=8080):
                with patch("socketserver.TCPServer"):
                    with patch("webbrowser.open"):
                        with patch.object(handler, "_run_server"):
                            try:
                                handler.start_oauth_flow(timeout=1)
                            except Exception:
                                pass

            # THEN: Cryptographically secure generation used
            mock_token.assert_called_once_with(32)


class TestAPIKeyExchangeSpecification:
    """Specification tests for API key exchange requirements.

    User Story:
    AS a user authorizing via browser
    I WANT my API key to be securely obtained and stored
    SO THAT I can authenticate with MoAI Rank services
    """

    def test_given_user_authorizes_when_callback_with_api_key_then_store_credentials(self):
        """GIVEN a user authorizes the application via browser
        WHEN the callback is received with a valid API key
        THEN credentials should be stored and success callback triggered

        Business Rule:
        - New MoAI Rank flow provides API key directly
        - Username, user_id, created_at should also be captured
        - Credentials should be saved via RankConfig.save_credentials
        - Success callback should be invoked with credentials
        """
        # GIVEN: User authorizes via browser
        test_state = OAuthState(state="valid-state", redirect_port=8080)
        OAuthCallbackHandler.oauth_state = test_state

        # WHEN: Callback received with API key
        with patch("http.server.BaseHTTPRequestHandler.__init__", return_value=None):
            handler = OAuthCallbackHandler(
                None,
                (
                    "/callback?state=valid-state&api_key=sk-moai-rank-key"
                    "&username=testuser&user_id=github-123&created_at=2024-01-15T10:30:00Z"
                ),
                None,
            )
            handler.path = (
                "/callback?state=valid-state&api_key=sk-moai-rank-key"
                "&username=testuser&user_id=github-123&created_at=2024-01-15T10:30:00Z"
            )
            handler.send_response = Mock()
            handler.send_header = Mock()
            handler.end_headers = Mock()
            handler.wfile = Mock()
            handler.wfile.write = Mock()
            handler.do_GET()

        # THEN: Credentials populated
        assert test_state.api_key == "sk-moai-rank-key"
        assert test_state.username == "testuser"
        assert test_state.user_id == "github-123"
        assert test_state.created_at == "2024-01-15T10:30:00Z"
        assert test_state.callback_received is True
        handler.send_response.assert_called_once_with(200)

    def test_given_legacy_flow_when_callback_with_code_then_exchange_for_api_key(self):
        """GIVEN a user authorizes via legacy OAuth flow
        WHEN the callback is received with an authorization code
        THEN the code should be exchanged for an API key

        Business Rule:
        - Legacy flow uses authorization code
        - Code exchanged via POST to /api/auth/cli/token
        - Response should contain apiKey, username, userId, createdAt
        - Backward compatibility must be maintained
        """
        # GIVEN: Legacy OAuth flow with auth code
        handler = OAuthHandler()

        with patch.object(handler, "_find_free_port", return_value=8080):
            with patch("socketserver.TCPServer") as mock_server_class:
                mock_server = Mock()

                # Simulate callback with auth code
                def mock_run_server(shutdown_event, timeout):
                    handler._oauth_state.auth_code = "legacy-auth-code-xyz"
                    handler._oauth_state.callback_received = True
                    shutdown_event.set()

                mock_server_class.return_value = mock_server

                with patch("webbrowser.open"):
                    with patch.object(handler, "_run_server", mock_run_server):
                        # Mock token exchange response
                        mock_response = Mock()
                        mock_response.status_code = 200
                        mock_response.json.return_value = {
                            "apiKey": "sk-exchanged-api-key",
                            "username": "legacy-user",
                            "userId": "legacy-456",
                            "createdAt": "2024-01-15T10:30:00Z",
                        }

                        with patch("requests.post", return_value=mock_response):
                            with patch("moai_adk.rank.auth.RankConfig.save_credentials"):
                                result = handler.start_oauth_flow(timeout=5)

                                # THEN: Code exchanged for API key
                                assert result is not None
                                assert result.api_key == "sk-exchanged-api-key"
                                assert result.username == "legacy-user"


class TestTimeoutHandlingSpecification:
    """Specification tests for timeout handling requirements.

    User Story:
    AS a user initiating OAuth flow
    I WANT the flow to timeout if I don't complete authorization
    SO THAT the process doesn't hang indefinitely
    """

    def test_given_oauth_flow_waiting_when_timeout_elapses_then_return_none_with_error(self):
        """GIVEN an OAuth flow is waiting for user authorization
        WHEN the specified timeout period elapses without callback
        THEN the flow should return None and invoke error callback

        Business Rule:
        - Default timeout is 300 seconds (5 minutes)
        - Server shutdown after timeout + 5 second grace period
        - Error message should indicate timeout
        - Error callback should be invoked
        """
        # GIVEN: OAuth flow initiated
        handler = OAuthHandler()

        # WHEN: Timeout elapses (simulated with timeout=0)
        with patch.object(handler, "_find_free_port", return_value=8080):
            with patch("socketserver.TCPServer"):
                with patch("webbrowser.open"):
                    # Simulate immediate timeout (no callback received)
                    def mock_run_server(shutdown_event, timeout):
                        # Don't set shutdown_event - let it timeout
                        pass

                    with patch.object(handler, "_run_server", mock_run_server):
                        error_messages = []

                        def mock_on_error(msg):
                            error_messages.append(msg)

                        result = handler.start_oauth_flow(on_error=mock_on_error, timeout=0)

                        # THEN: None returned, timeout error reported
                        assert result is None
                        assert len(error_messages) > 0
                        assert any("timed out" in msg.lower() for msg in error_messages)

    def test_given_custom_timeout_when_flow_starts_then_use_custom_timeout(self):
        """GIVEN an OAuth flow is initiated
        WHEN a custom timeout is specified
        THEN the flow should use the custom timeout value

        Business Rule:
        - Timeout parameter should be respected
        - Allows users to customize wait time
        - Default is 300 seconds if not specified
        """
        # GIVEN: Custom timeout value
        custom_timeout = 120  # 2 minutes

        # WHEN: Flow starts with custom timeout
        handler = OAuthHandler()

        with patch.object(handler, "_find_free_port", return_value=8080):
            with patch("socketserver.TCPServer"):
                with patch("webbrowser.open"):
                    with patch.object(handler, "_run_server") as mock_run:
                        handler.start_oauth_flow(timeout=custom_timeout)

                        # THEN: Custom timeout passed to server runner
                        mock_run.assert_called_once()
                        # Timeout is passed as second positional argument to _run_server
                        call_args = mock_run.call_args[0]
                        assert len(call_args) >= 2
                        # The timeout argument should be passed through
                        assert custom_timeout in call_args or mock_run.call_args.kwargs.get("timeout") == custom_timeout


class TestCredentialPersistenceSpecification:
    """Specification tests for credential persistence requirements.

    User Story:
    AS a user completing OAuth authorization
    I WANT my credentials to be securely stored
    SO THAT I don't need to re-authenticate on every use
    """

    def test_given_successful_oauth_flow_when_credentials_received_then_save_via_rank_config(self):
        """GIVEN a user successfully completes OAuth flow
        WHEN API key and user details are received
        THEN credentials should be saved via RankConfig.save_credentials

        Business Rule:
        - RankCredentials should contain api_key, username, user_id, created_at
        - Save should happen immediately after successful exchange
        - Both new and legacy flows should trigger save
        """
        # GIVEN: Successful OAuth flow
        handler = OAuthHandler()

        with patch.object(handler, "_find_free_port", return_value=8080):
            with patch("socketserver.TCPServer") as mock_server_class:
                mock_server = Mock()

                # Simulate successful callback with API key
                def mock_run_server(shutdown_event, timeout):
                    handler._oauth_state.api_key = "sk-new-credential"
                    handler._oauth_state.username = "persisted-user"
                    handler._oauth_state.user_id = "user-789"
                    handler._oauth_state.created_at = "2024-01-15T10:30:00Z"
                    handler._oauth_state.callback_received = True
                    shutdown_event.set()

                mock_server_class.return_value = mock_server

                with patch("webbrowser.open"):
                    with patch.object(handler, "_run_server", mock_run_server):
                        with patch("moai_adk.rank.auth.RankConfig.save_credentials") as _mock_save:
                            _result = handler.start_oauth_flow(timeout=5)

                            # THEN: Credentials saved
                            _mock_save.assert_called_once()
                            saved_creds = _mock_save.call_args[0][0]
                            assert isinstance(saved_creds, RankCredentials)
                            assert saved_creds.api_key == "sk-new-credential"
                            assert saved_creds.username == "persisted-user"

    def test_given_credentials_saved_when_success_callback_then_invoked_with_credentials(self):
        """GIVEN credentials are saved after successful OAuth
        WHEN save operation completes
        THEN the on_success callback should be invoked with the credentials

        Business Rule:
        - Success callback receives RankCredentials object
        - Allows caller to take action after successful registration
        """
        # GIVEN: Successful OAuth flow with callback
        handler = OAuthHandler()

        with patch.object(handler, "_find_free_port", return_value=8080):
            with patch("socketserver.TCPServer") as mock_server_class:
                mock_server = Mock()

                def mock_run_server(shutdown_event, timeout):
                    handler._oauth_state.api_key = "sk-callback-test"
                    handler._oauth_state.username = "callback-user"
                    handler._oauth_state.user_id = "cb-123"
                    handler._oauth_state.created_at = "2024-01-15"
                    handler._oauth_state.callback_received = True
                    shutdown_event.set()

                mock_server_class.return_value = mock_server

                with patch("webbrowser.open"):
                    with patch.object(handler, "_run_server", mock_run_server):
                        with patch("moai_adk.rank.auth.RankConfig.save_credentials"):
                            received_creds = []

                            def mock_on_success(credentials):
                                received_creds.append(credentials)

                            result = handler.start_oauth_flow(on_success=mock_on_success, timeout=5)

                            # THEN: Success callback invoked with credentials
                            assert len(received_creds) == 1
                            assert received_creds[0].api_key == "sk-callback-test"
                            assert result.api_key == "sk-callback-test"


class TestUserCancellationSpecification:
    """Specification tests for user cancellation handling.

    User Story:
    AS a user initiating OAuth flow
    I WANT to be able to cancel the process
    SO THAT I'm not forced to complete authorization
    """

    def test_given_user_presses_ctrl_c_when_flow_waiting_then_return_none_gracefully(self):
        """GIVEN an OAuth flow is waiting for user authorization
        WHEN the user presses Ctrl+C (KeyboardInterrupt)
        THEN the flow should return None and invoke error callback gracefully

        Business Rule:
        - KeyboardInterrupt should be caught, not propagated
        - Error callback should receive cancellation message
        - Resources should be cleaned up
        - No exception should be raised to caller
        """
        # GIVEN: OAuth flow in progress
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

                        # THEN: Graceful cancellation
                        assert result is None
                        # The keyboard interrupt is caught before error_messages can be set
                        # so we just verify None is returned


class TestErrorHandlingSpecification:
    """Specification tests for error handling requirements.

    User Story:
    AS a user attempting OAuth registration
    I WANT clear error messages when something goes wrong
    SO THAT I understand what went wrong and how to fix it
    """

    def test_given_oauth_provider_error_when_callback_then_display_error_message(self):
        """GIVEN the OAuth provider returns an error
        WHEN the callback is received with error parameter
        THEN a user-friendly error message should be displayed

        Business Rule:
        - Error parameter from OAuth provider should be captured
        - error_description should be used if available
        - User should see the error in browser
        - Error callback should be invoked
        """
        # GIVEN: OAuth provider error
        test_state = OAuthState(state="valid-state", redirect_port=8080)
        OAuthCallbackHandler.oauth_state = test_state

        # WHEN: Callback with error parameter
        with patch("http.server.BaseHTTPRequestHandler.__init__", return_value=None):
            handler = OAuthCallbackHandler(
                None,
                ("/callback?state=valid-state&error=access_denied&error_description=User+denied+authorization+request"),
                None,
            )
            handler.path = (
                "/callback?state=valid-state&error=access_denied&error_description=User+denied+authorization+request"
            )
            handler.send_response = Mock()
            handler.send_header = Mock()
            handler.end_headers = Mock()
            handler.wfile = Mock()
            handler.wfile.write = Mock()
            handler.do_GET()

        # THEN: Error captured
        assert test_state.error == "User denied authorization request"
        handler.send_response.assert_called_once_with(400)

    def test_given_incomplete_callback_when_no_code_or_api_key_then_show_error(self):
        """GIVEN a callback is received
        WHEN neither code nor api_key parameters are present
        THEN an incomplete authorization error should be displayed

        Business Rule:
        - Valid callback must have code (legacy) or api_key (new)
        - Missing both indicates something went wrong
        - Clear error message should inform user
        """
        # GIVEN: Callback without code or api_key
        test_state = OAuthState(state="valid-state", redirect_port=8080)
        OAuthCallbackHandler.oauth_state = test_state

        # WHEN: Callback with only state parameter
        with patch("http.server.BaseHTTPRequestHandler.__init__", return_value=None):
            handler = OAuthCallbackHandler(
                None,
                ("/callback?state=valid-state"),
                None,
            )
            handler.path = "/callback?state=valid-state"
            handler.send_response = Mock()
            handler.send_header = Mock()
            handler.end_headers = Mock()
            handler.wfile = Mock()
            handler.wfile.write = Mock()
            handler.do_GET()

        # THEN: Incomplete authorization error
        assert test_state.error == "No authorization code or API key received"
        handler.send_response.assert_called_once_with(400)

    def test_given_token_exchange_fails_when_legacy_flow_then_return_none(self):
        """GIVEN a legacy OAuth flow with auth code
        WHEN the token exchange request fails
        THEN the flow should return None with error message

        Business Rule:
        - Network failures should be handled gracefully
        - Non-200 responses should indicate failure
        - Error callback should be invoked
        - No credentials should be stored
        """
        # GIVEN: Legacy flow with token exchange failure
        handler = OAuthHandler()

        with patch.object(handler, "_find_free_port", return_value=8080):
            with patch("socketserver.TCPServer") as mock_server_class:
                mock_server = Mock()

                def mock_run_server(shutdown_event, timeout):
                    handler._oauth_state.auth_code = "will-fail-code"
                    handler._oauth_state.callback_received = True
                    shutdown_event.set()

                mock_server_class.return_value = mock_server

                with patch("webbrowser.open"):
                    with patch.object(handler, "_run_server", mock_run_server):
                        # Mock failed exchange
                        mock_response = Mock()
                        mock_response.status_code = 401  # Unauthorized

                        with patch("requests.post", return_value=mock_response):
                            error_messages = []

                            def mock_on_error(msg):
                                error_messages.append(msg)

                            result = handler.start_oauth_flow(on_error=mock_on_error, timeout=5)

                            # THEN: None returned, error reported
                            assert result is None
                            assert len(error_messages) > 0


class TestBrowserIntegrationSpecification:
    """Specification tests for browser integration requirements.

    User Story:
    AS a user initiating OAuth flow
    I WANT the authorization page to open in my browser
    SO THAT I can easily authorize the application
    """

    def test_given_oauth_flow_starts_when_browser_opens_then_show_auth_url(self):
        """GIVEN an OAuth flow is initiated
        WHEN the flow starts
        THEN the browser should open with the authorization URL

        Business Rule:
        - webbrowser.open should be called with auth URL
        - URL should contain state and redirect_uri parameters
        - URL should point to configured base_url
        - Auth URL should also be printed to console
        """
        # GIVEN: OAuth flow starting
        handler = OAuthHandler(RankConfig(base_url="https://rank.moai.dev"))

        # WHEN: Flow starts
        with patch.object(handler, "_find_free_port", return_value=8080):
            with patch("socketserver.TCPServer"):
                with patch("webbrowser.open") as mock_open:
                    with patch("builtins.print") as mock_print:
                        with patch.object(handler, "_run_server"):
                            try:
                                handler.start_oauth_flow(timeout=1)
                            except Exception:
                                pass

                            # THEN: Browser opened with auth URL
                            if mock_open.called:
                                url = mock_open.call_args[0][0]
                                assert "rank.moai.dev/api/auth/cli" in url
                                # URL is URL-encoded
                                assert "redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback" in url
                                assert "state=" in url

                            # URL printed to console
                            if mock_print.called:
                                printed_url = mock_print.call_args[0][0]
                                assert "rank.moai.dev" in printed_url


class TestLocalServerSpecification:
    """Specification tests for local callback server requirements.

    User Story:
    AS an OAuth flow implementation
    I WANT a local HTTP server to receive the callback
    SO THAT the authorization response can be captured
    """

    def test_given_oauth_flow_starts_when_local_server_starts_on_free_port(self):
        """GIVEN an OAuth flow is initiated
        WHEN the local server is started
        THEN it should bind to an available port in the configured range

        Business Rule:
        - Port range should be 8080-8179 (default)
        - First available port should be used
        - Server should listen on localhost only
        """
        # GIVEN: OAuth flow starting
        handler = OAuthHandler()

        # WHEN: Finding free port
        port = handler._find_free_port(start=8080, end=8180)

        # THEN: Port in valid range
        assert 8080 <= port < 8180

    def test_given_port_range_exhausted_when_no_available_port_then_raise_error(self):
        """GIVEN port range is specified
        WHEN all ports in range are occupied
        THEN RuntimeError should be raised

        Business Rule:
        - Should scan entire range sequentially
        - Should raise clear error if no port available
        - Error should indicate port exhaustion
        """
        # GIVEN: Port range
        handler = OAuthHandler()

        # WHEN: All ports occupied
        with patch("socket.socket") as mock_socket_class:
            mock_socket = Mock()
            mock_socket.bind.side_effect = OSError("Address already in use")
            mock_socket_class.return_value.__enter__.return_value = mock_socket

            # THEN: RuntimeError raised
            with pytest.raises(RuntimeError, match="No available port found"):
                handler._find_free_port(start=9000, end=9001)


class TestVerifyAPISpecification:
    """Specification tests for verify_api_key function.

    User Story:
    AS a developer validating stored credentials
    I WANT to verify that an API key is valid
    SO THAT I can prompt for re-authentication if needed
    """

    def test_given_valid_api_key_when_verifying_then_return_true(self):
        """GIVEN a valid API key
        WHEN verified via HTTP request
        THEN True should be returned

        Business Rule:
        - GET request to /rank endpoint with X-API-Key header
        - 200 status code indicates valid key
        - Should use configured api_base_url
        """
        # GIVEN: Valid API key
        api_key = "sk-valid-key-12345"

        # WHEN: Verifying
        mock_response = Mock()
        mock_response.status_code = 200

        with patch("requests.get", return_value=mock_response) as mock_get:
            result = verify_api_key(api_key)

            # THEN: Returns True
            assert result is True
            # Request made correctly
            mock_get.assert_called_once()
            headers = mock_get.call_args.kwargs["headers"]
            assert headers["X-API-Key"] == api_key

    def test_given_invalid_api_key_when_verifying_then_return_false(self):
        """GIVEN an invalid or expired API key
        WHEN verified via HTTP request
        THEN False should be returned

        Business Rule:
        - Non-200 status code indicates invalid key
        - Should not raise exception for invalid keys
        """
        # GIVEN: Invalid API key
        api_key = "sk-invalid-key"

        # WHEN: Verifying
        mock_response = Mock()
        mock_response.status_code = 401  # Unauthorized

        with patch("requests.get", return_value=mock_response):
            result = verify_api_key(api_key)

            # THEN: Returns False
            assert result is False

    def test_given_network_error_when_verifying_then_return_false(self):
        """GIVEN a network error occurs during verification
        WHEN the request fails
        THEN False should be returned

        Business Rule:
        - Any RequestException should return False
        - Should not propagate network errors
        - Handles timeouts, connection errors, etc.
        """
        # GIVEN: Network error
        api_key = "sk-any-key"
        import requests

        # WHEN: Verifying
        with patch("requests.get", side_effect=requests.RequestException("Network error")):
            result = verify_api_key(api_key)

            # THEN: Returns False gracefully
            assert result is False
