"""API client for MoAI Rank service with HMAC-SHA256 request signing.

This client handles all communication with the rank.mo.ai.kr service,
including secure request signing to prevent tampering.
"""

import hashlib
import hmac
import json
import time
from dataclasses import dataclass
from typing import Any, Optional

import requests

from moai_adk.rank.config import RankConfig


@dataclass
class RankInfo:
    """Ranking information for a specific period."""

    position: int
    composite_score: float
    total_participants: int


@dataclass
class UserRank:
    """Complete user ranking data."""

    username: str
    daily: Optional[RankInfo]
    weekly: Optional[RankInfo]
    monthly: Optional[RankInfo]
    all_time: Optional[RankInfo]
    total_tokens: int
    total_sessions: int
    input_tokens: int
    output_tokens: int
    last_updated: str


@dataclass
class LeaderboardEntry:
    """Single entry in the leaderboard."""

    rank: int
    username: str
    total_tokens: int
    composite_score: float
    session_count: int
    is_private: bool


@dataclass
class ApiStatus:
    """API health status."""

    status: str
    version: str
    timestamp: str


@dataclass
class SessionSubmission:
    """Session data for submission.

    Attributes:
        session_hash: Unique hash for deduplication
        ended_at: Session end timestamp (UTC ISO format)
        input_tokens: Total input tokens consumed
        output_tokens: Total output tokens generated
        cache_creation_tokens: Tokens used for cache creation
        cache_read_tokens: Tokens read from cache
        model_name: Primary model used in the session

        # Dashboard fields (for activity visualization)
        started_at: Session start timestamp (UTC ISO format)
        duration_seconds: Total session duration in seconds
        turn_count: Number of user turns (messages)
        tool_usage: Tool usage counts (e.g., {"Read": 5, "Write": 3})
        model_usage: Per-model token usage (e.g., {"claude-opus-4-5": {"input": 5000, "output": 2000}})
        code_metrics: Code change metrics (e.g., {"linesAdded": 100, "linesDeleted": 20, ...})
    """

    session_hash: str
    ended_at: str
    input_tokens: int
    output_tokens: int
    cache_creation_tokens: int = 0
    cache_read_tokens: int = 0
    model_name: Optional[str] = None

    # Dashboard fields for activity visualization
    started_at: Optional[str] = None
    duration_seconds: int = 0
    turn_count: int = 0
    tool_usage: Optional[dict[str, int]] = None
    model_usage: Optional[dict[str, dict[str, int]]] = None
    code_metrics: Optional[dict[str, int]] = None


class RankClientError(Exception):
    """Base exception for rank client errors."""

    pass


class AuthenticationError(RankClientError):
    """Raised when authentication fails."""

    pass


class ApiError(RankClientError):
    """Raised when API returns an error."""

    def __init__(self, message: str, status_code: int, details: Optional[dict] = None):
        super().__init__(message)
        self.status_code = status_code
        self.details = details or {}


class RankClient:
    """HTTP client for MoAI Rank API with HMAC authentication."""

    def __init__(
        self,
        api_key: Optional[str] = None,
        config: Optional[RankConfig] = None,
        timeout: int = 30,
    ):
        """Initialize the rank client.

        Args:
            api_key: API key for authentication (loaded from config if not provided)
            config: Configuration instance (uses defaults if not provided)
            timeout: Request timeout in seconds
        """
        self.config = config or RankConfig()
        self._api_key = api_key or RankConfig.get_api_key()
        self.timeout = timeout
        self._session = requests.Session()
        self._session.headers.update(
            {
                "User-Agent": "moai-adk/1.0",
                "Content-Type": "application/json",
                "Accept": "application/json",
            }
        )

    @property
    def api_key(self) -> Optional[str]:
        """Get the current API key."""
        return self._api_key

    @api_key.setter
    def api_key(self, value: str) -> None:
        """Set the API key."""
        self._api_key = value

    def _compute_signature(self, timestamp: str, body: str) -> str:
        """Compute HMAC-SHA256 signature for request authentication.

        The signature is computed as:
        HMAC-SHA256(api_key, timestamp + ":" + body)

        Args:
            timestamp: Unix timestamp as string
            body: Request body as JSON string

        Returns:
            Hexadecimal signature string
        """
        if not self._api_key:
            raise AuthenticationError("API key not configured")

        message = f"{timestamp}:{body}"
        signature = hmac.new(
            self._api_key.encode("utf-8"),
            message.encode("utf-8"),
            hashlib.sha256,
        ).hexdigest()

        return signature

    def _get_auth_headers(self, body: str = "") -> dict[str, str]:
        """Get authentication headers for a request.

        Args:
            body: Request body for signature computation

        Returns:
            Dictionary of authentication headers
        """
        if not self._api_key:
            raise AuthenticationError("API key not configured")

        timestamp = str(int(time.time()))
        signature = self._compute_signature(timestamp, body)

        return {
            "X-API-Key": self._api_key,
            "X-Timestamp": timestamp,
            "X-Signature": signature,
        }

    def _make_request(
        self,
        method: str,
        endpoint: str,
        data: Optional[dict] = None,
        auth: bool = False,
        hmac_auth: bool = False,
    ) -> dict[str, Any]:
        """Make an HTTP request to the API.

        Args:
            method: HTTP method (GET, POST, etc.)
            endpoint: API endpoint path
            data: Request body data (for POST)
            auth: Whether to include API key header
            hmac_auth: Whether to include HMAC signature headers

        Returns:
            Response data as dictionary

        Raises:
            AuthenticationError: If authentication fails
            ApiError: If API returns an error
            RankClientError: For other request failures
        """
        url = f"{self.config.api_base_url}{endpoint}"
        body_str = json.dumps(data) if data else ""

        headers = {}
        if hmac_auth:
            headers.update(self._get_auth_headers(body_str))
        elif auth:
            if not self._api_key:
                raise AuthenticationError("API key not configured")
            headers["X-API-Key"] = self._api_key

        try:
            if method.upper() == "GET":
                response = self._session.get(url, headers=headers, timeout=self.timeout)
            elif method.upper() == "POST":
                response = self._session.post(
                    url,
                    headers=headers,
                    data=body_str,
                    timeout=self.timeout,
                )
            else:
                raise ValueError(f"Unsupported HTTP method: {method}")

            # Parse response
            try:
                response_data = response.json()
            except json.JSONDecodeError:
                response_data = {"raw": response.text}

            # Handle errors
            if response.status_code == 401:
                raise AuthenticationError(response_data.get("error", "Authentication failed"))

            if response.status_code >= 400:
                details = response_data.get("details")
                raise ApiError(
                    response_data.get("error", f"API error: {response.status_code}"),
                    response.status_code,
                    details if isinstance(details, dict) else None,
                )

            return response_data

        except requests.RequestException as e:
            raise RankClientError(f"Request failed: {e}") from e

    def check_status(self) -> ApiStatus:
        """Check API status and availability.

        Returns:
            ApiStatus with service health information
        """
        data = self._make_request("GET", "/status")
        return ApiStatus(
            status=data.get("status", "unknown"),
            version=data.get("version", "unknown"),
            timestamp=data.get("timestamp", ""),
        )

    def get_user_rank(self) -> UserRank:
        """Get current user's rank and statistics.

        Returns:
            UserRank with complete ranking data

        Raises:
            AuthenticationError: If API key is invalid
        """
        response = self._make_request("GET", "/rank", auth=True)

        # API returns {"success": true, "data": {...}}
        data = response.get("data", response)

        def parse_rank_info(rank_data: Optional[dict]) -> Optional[RankInfo]:
            if not rank_data:
                return None
            return RankInfo(
                position=rank_data.get("position", 0),
                composite_score=rank_data.get("compositeScore", 0.0),
                total_participants=rank_data.get("totalParticipants", 0),
            )

        rankings = data.get("rankings", {})
        stats = data.get("stats", {})

        return UserRank(
            username=data.get("username", "unknown"),
            daily=parse_rank_info(rankings.get("daily")),
            weekly=parse_rank_info(rankings.get("weekly")),
            monthly=parse_rank_info(rankings.get("monthly")),
            all_time=parse_rank_info(rankings.get("allTime")),
            total_tokens=stats.get("totalTokens", 0),
            total_sessions=stats.get("totalSessions", 0),
            input_tokens=stats.get("inputTokens", 0),
            output_tokens=stats.get("outputTokens", 0),
            last_updated=data.get("lastUpdated", ""),
        )

    def get_leaderboard(
        self,
        period: str = "weekly",
        limit: int = 10,
        offset: int = 0,
    ) -> list[LeaderboardEntry]:
        """Get the leaderboard for a specific period.

        Args:
            period: Period type (daily, weekly, monthly, all_time)
            limit: Maximum number of entries (1-100)
            offset: Pagination offset

        Returns:
            List of LeaderboardEntry objects
        """
        # Use the public leaderboard endpoint (no auth needed)
        url = f"{self.config.base_url}/api/leaderboard"
        params = {
            "period": period,
            "limit": str(min(limit, 100)),
            "offset": str(offset),
        }

        try:
            response = self._session.get(url, params=params, timeout=self.timeout)
            data = response.json()

            # Validate response is a dictionary before calling .get()
            if not isinstance(data, dict):
                raise ApiError(
                    f"Unexpected response format: {type(data).__name__}",
                    response.status_code,
                )

            if response.status_code >= 400:
                raise ApiError(
                    data.get("error", "Failed to fetch leaderboard"),
                    response.status_code,
                )

            entries = []

            # API returns: {"data": {"items": [...], "pagination": {...}}}
            data_obj = data.get("data", {})
            if isinstance(data_obj, dict):
                leaderboard_data = data_obj.get("items", [])
            elif isinstance(data_obj, list):
                # Fallback for old API format
                leaderboard_data = data_obj
            else:
                leaderboard_data = []

            for item in leaderboard_data:
                # Skip non-dict items
                if not isinstance(item, dict):
                    continue
                entries.append(
                    LeaderboardEntry(
                        rank=item.get("rank", 0),
                        username=item.get("username", "Unknown"),
                        total_tokens=item.get("totalTokens", 0),
                        composite_score=item.get("compositeScore", 0.0),
                        session_count=item.get("sessionCount", 0),
                        is_private=item.get("isPrivate", False),
                    )
                )

            return entries

        except requests.RequestException as e:
            raise RankClientError(f"Failed to fetch leaderboard: {e}") from e

    def submit_session(self, session: SessionSubmission) -> dict[str, Any]:
        """Submit a Claude Code session with token usage.

        This method uses HMAC authentication to prevent tampering.

        Args:
            session: Session data to submit

        Returns:
            API response with session ID and confirmation

        Raises:
            AuthenticationError: If API key or signature is invalid
            ApiError: If submission fails
        """
        data = {
            "sessionHash": session.session_hash,
            "endedAt": session.ended_at,
            "inputTokens": session.input_tokens,
            "outputTokens": session.output_tokens,
            "cacheCreationTokens": session.cache_creation_tokens,
            "cacheReadTokens": session.cache_read_tokens,
        }

        if session.model_name:
            data["modelName"] = session.model_name

        # Dashboard fields (optional)
        if session.started_at:
            data["startedAt"] = session.started_at

        if session.duration_seconds > 0:
            data["durationSeconds"] = session.duration_seconds

        if session.turn_count > 0:
            data["turnCount"] = session.turn_count

        if session.tool_usage:
            data["toolUsage"] = session.tool_usage

        if session.model_usage:
            data["modelUsage"] = session.model_usage

        if session.code_metrics:
            data["codeMetrics"] = session.code_metrics

        return self._make_request("POST", "/sessions", data=data, hmac_auth=True)

    def submit_sessions_batch(self, sessions: list[SessionSubmission]) -> dict[str, Any]:
        """Submit multiple Claude Code sessions in a single batch request.

        This method uses HMAC authentication and is more efficient than
        submitting sessions individually when syncing many sessions.

        Args:
            sessions: List of session data to submit (max 100 per batch)

        Returns:
            API response with batch results:
            - success: Overall success status
            - processed: Number of sessions processed
            - succeeded: Number of sessions successfully submitted
            - failed: Number of sessions that failed
            - results: List of individual results per session

        Raises:
            AuthenticationError: If API key or signature is invalid
            ApiError: If batch submission fails
            ValueError: If more than 100 sessions provided
        """
        if len(sessions) > 100:
            raise ValueError("Maximum 100 sessions per batch request")

        sessions_data = []
        for session in sessions:
            session_data = {
                "sessionHash": session.session_hash,
                "endedAt": session.ended_at,
                "inputTokens": session.input_tokens,
                "outputTokens": session.output_tokens,
                "cacheCreationTokens": session.cache_creation_tokens,
                "cacheReadTokens": session.cache_read_tokens,
            }

            if session.model_name:
                session_data["modelName"] = session.model_name

            if session.started_at:
                session_data["startedAt"] = session.started_at

            if session.duration_seconds > 0:
                session_data["durationSeconds"] = session.duration_seconds

            if session.turn_count > 0:
                session_data["turnCount"] = session.turn_count

            if session.tool_usage:
                session_data["toolUsage"] = session.tool_usage

            if session.model_usage:
                session_data["modelUsage"] = session.model_usage

            if session.code_metrics:
                session_data["codeMetrics"] = session.code_metrics

            sessions_data.append(session_data)

        data = {"sessions": sessions_data}

        return self._make_request("POST", "/sessions/batch", data=data, hmac_auth=True)

    def compute_session_hash(
        self,
        input_tokens: int,
        output_tokens: int,
        cache_creation_tokens: int,
        cache_read_tokens: int,
        model_name: Optional[str],
        ended_at: str,
    ) -> str:
        """Compute a client-side session hash for deduplication.

        Args:
            input_tokens: Number of input tokens
            output_tokens: Number of output tokens
            cache_creation_tokens: Cache creation tokens
            cache_read_tokens: Cache read tokens
            model_name: Model name
            ended_at: Session end timestamp

        Returns:
            SHA-256 hash string
        """
        # Create a deterministic string from session data
        # SECURITY: Use cryptographically secure random token instead of time.time_ns()
        # to prevent session hash prediction even if timestamp is known
        import secrets

        data = ":".join(
            [
                str(input_tokens),
                str(output_tokens),
                str(cache_creation_tokens),
                str(cache_read_tokens),
                model_name or "",
                ended_at,
                # Cryptographically secure random component for uniqueness
                secrets.token_hex(16),
            ]
        )

        return hashlib.sha256(data.encode()).hexdigest()
