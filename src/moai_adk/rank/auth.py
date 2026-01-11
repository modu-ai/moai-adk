"""OAuth flow handler for MoAI Rank registration.

This module handles the GitHub OAuth flow for registering with the
MoAI Rank service. It opens a browser for authorization and polls
for the API key once authorization is complete.
"""

import http.server
import secrets
import socketserver
import threading
import time
import urllib.parse
import webbrowser
from dataclasses import dataclass
from typing import Callable, Optional

import requests

from moai_adk.rank.config import RankConfig, RankCredentials


@dataclass
class OAuthState:
    """OAuth state during authorization flow."""

    state: str
    redirect_port: int
    callback_received: bool = False
    auth_code: Optional[str] = None
    api_key: Optional[str] = None  # Direct API key from new flow
    username: Optional[str] = None  # Username from callback
    user_id: Optional[str] = None  # GitHub user ID from callback
    created_at: Optional[str] = None  # Timestamp from callback
    error: Optional[str] = None


class OAuthCallbackHandler(http.server.BaseHTTPRequestHandler):
    """HTTP handler for OAuth callback."""

    oauth_state: Optional[OAuthState] = None
    _on_complete: Optional[Callable[[], None]] = None

    def log_message(self, format: str, *args: object) -> None:  # noqa: A002
        """Suppress default logging."""
        _ = format, args  # Unused but required by base class

    def do_GET(self) -> None:
        """Handle OAuth callback GET request."""
        if not self.oauth_state:
            self.send_error(500, "OAuth state not initialized")
            return

        # Parse query parameters
        parsed = urllib.parse.urlparse(self.path)
        params = urllib.parse.parse_qs(parsed.query)

        # Handle callback path
        if parsed.path == "/callback":
            # Verify state parameter
            state = params.get("state", [""])[0]
            if state != self.oauth_state.state:
                self.oauth_state.error = "State mismatch - possible CSRF attack"
                self._send_response(400, "Invalid state parameter")
                return

            # Check for error
            if "error" in params:
                self.oauth_state.error = params.get("error_description", ["Unknown error"])[0]
                self._send_response(400, f"Authorization failed: {self.oauth_state.error}")
                return

            # Check for direct API key (new MoAI Rank flow)
            api_key = params.get("api_key", [""])[0]
            username = params.get("username", [""])[0]
            user_id = params.get("user_id", [""])[0]
            created_at = params.get("created_at", [""])[0]

            if api_key:
                # New flow: API key received directly from server
                self.oauth_state.api_key = api_key
                self.oauth_state.username = username
                self.oauth_state.user_id = user_id
                self.oauth_state.created_at = created_at
                self.oauth_state.callback_received = True
                self._send_success_response()

                # Signal completion
                on_complete_cb = OAuthCallbackHandler._on_complete  # type: ignore[misc]
                if on_complete_cb is not None:
                    threading.Thread(target=on_complete_cb, daemon=True).start()
                return

            # Fallback: Extract authorization code (legacy flow)
            code = params.get("code", [""])[0]
            if code:
                self.oauth_state.auth_code = code
                self.oauth_state.callback_received = True
                self._send_success_response()

                # Signal completion
                on_complete_cb = OAuthCallbackHandler._on_complete  # type: ignore[misc]
                if on_complete_cb is not None:
                    threading.Thread(target=on_complete_cb, daemon=True).start()
            else:
                self.oauth_state.error = "No authorization code or API key received"
                self._send_response(400, "Authorization incomplete")
        else:
            self.send_error(404, "Not Found")

    def _send_response(self, status: int, message: str) -> None:
        """Send a simple text response."""
        self.send_response(status)
        self.send_header("Content-Type", "text/html; charset=utf-8")
        self.end_headers()
        html = f"""
        <!DOCTYPE html>
        <html>
        <head>
            <title>MoAI Rank - Authorization</title>
            <style>
                body {{ font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
                       display: flex; justify-content: center; align-items: center; height: 100vh;
                       margin: 0; background: #f5f5f5; }}
                .container {{ text-align: center; padding: 40px; background: white;
                             border-radius: 12px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }}
                .error {{ color: #dc3545; }}
            </style>
        </head>
        <body>
            <div class="container">
                <h1 class="error">Authorization Failed</h1>
                <p>{message}</p>
                <p>You can close this window.</p>
            </div>
        </body>
        </html>
        """
        self.wfile.write(html.encode())

    def _send_success_response(self) -> None:
        """Send success response HTML."""
        self.send_response(200)
        self.send_header("Content-Type", "text/html; charset=utf-8")
        self.end_headers()
        html = """
        <!DOCTYPE html>
        <html>
        <head>
            <title>MoAI Rank - Authorization Complete</title>
            <style>
                body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
                       display: flex; justify-content: center; align-items: center; height: 100vh;
                       margin: 0; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
                .container { text-align: center; padding: 40px; background: white;
                             border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.2); }
                .success { color: #28a745; }
                h1 { margin-bottom: 10px; }
                p { color: #666; margin: 10px 0; }
            </style>
        </head>
        <body>
            <div class="container">
                <h1 class="success">Authorization Successful!</h1>
                <p>You have successfully connected your GitHub account to MoAI Rank.</p>
                <p>You can close this window and return to the terminal.</p>
                <script>setTimeout(() => window.close(), 3000);</script>
            </div>
        </body>
        </html>
        """
        self.wfile.write(html.encode())


class OAuthHandler:
    """Handles the OAuth registration flow with MoAI Rank."""

    def __init__(self, config: Optional[RankConfig] = None):
        """Initialize OAuth handler.

        Args:
            config: Configuration instance (uses defaults if not provided)
        """
        self.config = config or RankConfig()
        self._server: Optional[socketserver.TCPServer] = None
        self._oauth_state: Optional[OAuthState] = None

    def _find_free_port(self, start: int = 8080, end: int = 8180) -> int:
        """Find an available port in the given range."""
        import socket

        for port in range(start, end):
            try:
                with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
                    s.bind(("localhost", port))
                    return port
            except OSError:
                continue
        raise RuntimeError("No available port found")

    def start_oauth_flow(
        self,
        on_success: Optional[Callable[[RankCredentials], None]] = None,
        on_error: Optional[Callable[[str], None]] = None,
        timeout: int = 300,
    ) -> Optional[RankCredentials]:
        """Start the OAuth authorization flow.

        This method:
        1. Generates a secure state token
        2. Starts a local HTTP server for the callback
        3. Opens the browser to the authorization URL
        4. Waits for the callback
        5. Exchanges the auth code for an API key
        6. Stores the credentials securely

        Args:
            on_success: Callback on successful registration
            on_error: Callback on error
            timeout: Maximum time to wait for authorization (seconds)

        Returns:
            RankCredentials on success, None on failure
        """
        # Generate secure state
        state = secrets.token_urlsafe(32)
        port = self._find_free_port()

        self._oauth_state = OAuthState(state=state, redirect_port=port)

        # Configure callback handler
        OAuthCallbackHandler.oauth_state = self._oauth_state

        # Server shutdown event
        shutdown_event = threading.Event()

        def trigger_shutdown():
            shutdown_event.set()

        OAuthCallbackHandler._on_complete = trigger_shutdown  # type: ignore[method-assign]

        # Start callback server
        try:
            self._server = socketserver.TCPServer(("localhost", port), OAuthCallbackHandler)
            self._server.timeout = 1  # Allow periodic timeout checks

            # Run server in background thread
            server_thread = threading.Thread(
                target=self._run_server,
                args=(shutdown_event, timeout),
                daemon=True,
            )
            server_thread.start()

            # Build authorization URL
            redirect_uri = f"http://localhost:{port}/callback"
            auth_url = self._build_auth_url(state, redirect_uri)

            # Open browser
            webbrowser.open(auth_url)

            # Print the auth URL for manual access
            print(f"\n{auth_url}")

            # Wait for callback with interruptible intervals
            # This allows Ctrl+C to be properly handled
            try:
                start_wait = time.time()
                while server_thread.is_alive():
                    # Use small intervals to allow signal processing
                    server_thread.join(timeout=0.5)
                    if time.time() - start_wait > timeout + 5:
                        break
            except KeyboardInterrupt:
                error_msg = "Registration cancelled by user"
                if on_error:
                    on_error(error_msg)
                return None

            # Check result
            if self._oauth_state.error:
                if on_error:
                    on_error(self._oauth_state.error)
                return None

            if not self._oauth_state.callback_received:
                error_msg = "Authorization timed out"
                if on_error:
                    on_error(error_msg)
                return None

            # Check if we received a direct API key (new flow)
            if self._oauth_state.api_key:
                credentials = RankCredentials(
                    api_key=self._oauth_state.api_key,
                    username=self._oauth_state.username or "unknown",
                    user_id=self._oauth_state.user_id or "",
                    created_at=self._oauth_state.created_at or "",
                )
                # Save credentials
                RankConfig.save_credentials(credentials)
                if on_success:
                    on_success(credentials)
                return credentials

            # Fallback: Exchange code for API key (legacy flow)
            credentials = self._exchange_code_for_key(
                self._oauth_state.auth_code,
                redirect_uri,
            )

            if credentials:
                # Save credentials
                RankConfig.save_credentials(credentials)
                if on_success:
                    on_success(credentials)
                return credentials
            else:
                error_msg = "Failed to obtain API key"
                if on_error:
                    on_error(error_msg)
                return None

        except Exception as e:
            error_msg = f"OAuth flow error: {e}"
            if on_error:
                on_error(error_msg)
            return None
        finally:
            self._cleanup()

    def _run_server(self, shutdown_event: threading.Event, timeout: int) -> None:
        """Run the callback server until shutdown or timeout."""
        start_time = time.time()
        server = self._server

        if server is None:
            return

        while not shutdown_event.is_set():
            if time.time() - start_time > timeout:
                break
            server.handle_request()

    def _build_auth_url(self, state: str, redirect_uri: str) -> str:
        """Build the authorization URL."""
        params = {
            "redirect_uri": redirect_uri,
            "state": state,
        }
        query = urllib.parse.urlencode(params)
        return f"{self.config.base_url}/api/auth/cli?{query}"

    def _exchange_code_for_key(
        self,
        code: Optional[str],
        redirect_uri: str,
    ) -> Optional[RankCredentials]:
        """Exchange authorization code for API key.

        Args:
            code: Authorization code from OAuth callback
            redirect_uri: The redirect URI used in the authorization

        Returns:
            RankCredentials on success, None on failure
        """
        if not code:
            return None

        try:
            response = requests.post(
                f"{self.config.base_url}/api/auth/cli/token",
                json={"code": code, "redirect_uri": redirect_uri},
                timeout=30,
            )

            if response.status_code != 200:
                return None

            data = response.json()

            return RankCredentials(
                api_key=data.get("apiKey", ""),
                username=data.get("username", "unknown"),
                user_id=data.get("userId", ""),
                created_at=data.get("createdAt", ""),
            )

        except requests.RequestException:
            return None

    def _cleanup(self) -> None:
        """Clean up resources."""
        if self._server:
            try:
                # Don't call shutdown() - it blocks waiting for serve_forever()
                # which we don't use (we use handle_request() directly)
                self._server.server_close()
            except Exception:
                pass
            self._server = None


def verify_api_key(api_key: str, config: Optional[RankConfig] = None) -> bool:
    """Verify an API key by making a test request.

    Args:
        api_key: The API key to verify
        config: Configuration instance

    Returns:
        True if the key is valid, False otherwise
    """
    cfg = config if config is not None else RankConfig()

    try:
        response = requests.get(
            f"{cfg.api_base_url}/rank",
            headers={"X-API-Key": api_key},
            timeout=10,
        )
        return response.status_code == 200
    except requests.RequestException:
        return False
