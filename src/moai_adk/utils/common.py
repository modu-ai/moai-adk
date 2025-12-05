"""
Common Utilities
Common utility functions
"""

import asyncio
import logging
import re
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional
from urllib.parse import urlparse

import aiohttp
import yaml

logger = logging.getLogger(__name__)


@dataclass
class HTTPResponse:
    """HTTP response data"""

    status_code: int
    url: str
    load_time: float
    success: bool
    error_message: Optional[str] = None
    timestamp: datetime = field(default_factory=datetime.now)

    def __post_init__(self):
        if self.timestamp is None:
            self.timestamp = datetime.now()


class HTTPClient:
    """HTTP client utility"""

    def __init__(self, max_concurrent: int = 5, timeout: int = 10):
        self.max_concurrent = max_concurrent
        self.timeout = timeout
        self.session: Optional[aiohttp.ClientSession] = None

    async def __aenter__(self):
        """Async context manager entry"""
        connector = aiohttp.TCPConnector(limit=self.max_concurrent)
        timeout = aiohttp.ClientTimeout(total=self.timeout)
        self.session = aiohttp.ClientSession(
            connector=connector,
            timeout=timeout,
            headers={"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"},
        )
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit"""
        if self.session:
            await self.session.close()

    async def fetch_url(self, url: str) -> HTTPResponse:
        """Fetch single URL"""
        try:
            if self.session is None:
                return HTTPResponse(
                    status_code=0,
                    url=url,
                    load_time=0,
                    success=False,
                    error_message="Session not initialized",
                )
            start_time = asyncio.get_event_loop().time()
            async with self.session.get(url, allow_redirects=True) as response:
                load_time = asyncio.get_event_loop().time() - start_time
                success = 200 <= response.status < 300
                return HTTPResponse(
                    status_code=response.status,
                    url=str(response.url),
                    load_time=load_time,
                    success=success,
                )
        except asyncio.TimeoutError:
            return HTTPResponse(
                status_code=0,
                url=url,
                load_time=self.timeout,
                success=False,
                error_message="Request timeout",
            )
        except aiohttp.ClientError as e:
            return HTTPResponse(
                status_code=0,
                url=url,
                load_time=0.0,
                success=False,
                error_message=f"HTTP client error: {str(e)}",
            )
        except Exception as e:
            return HTTPResponse(
                status_code=0,
                url=url,
                load_time=0.0,
                success=False,
                error_message=f"Unexpected error: {str(e)}",
            )

    async def fetch_urls(self, urls: List[str]) -> List[HTTPResponse]:
        """Fetch multiple URLs concurrently"""
        async with self:
            tasks = [self.fetch_url(url) for url in urls]
            return await asyncio.gather(*tasks)


def extract_links_from_text(text: str, base_url: Optional[str] = None) -> List[str]:
    """Extract links from text"""
    links = []

    # Markdown link pattern: [text](url)
    markdown_pattern = r"\[([^\]]+)\]\(([^)]+)\)"
    markdown_matches = re.findall(markdown_pattern, text)

    for match in markdown_matches:
        url = match[1]
        # Convert relative URLs to absolute URLs
        if url.startswith(("http://", "https://")):
            links.append(url)
        elif base_url and url.startswith("/"):
            links.append(f"{base_url}{url}")
        elif base_url and not url.startswith(("http://", "https://", "#")):
            links.append(f"{base_url}/{url.rstrip('/')}")

    # General URL pattern
    url_pattern = r'https?://[^\s<>"\'()]+'
    url_matches = re.findall(url_pattern, text)
    links.extend(url_matches)

    logger.info(f"Found {len(links)} links in text")
    return list(set(links))  # Remove duplicates


def is_valid_url(url: str) -> bool:
    """Validate URL"""
    try:
        result = urlparse(url)
        return all([result.scheme, result.netloc])
    except Exception:
        return False


def create_report_path(base_path: Path, suffix: str = "report") -> Path:
    """Create report file path"""
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    filename = f"{suffix}_{timestamp}.md"
    return base_path / filename


def format_duration(seconds: float) -> str:
    """Convert time (seconds) to readable format"""
    if seconds < 1:
        return f"{seconds * 1000:.0f}ms"
    elif seconds < 60:
        return f"{seconds:.1f}s"
    elif seconds < 3600:
        minutes = int(seconds // 60)
        remaining_seconds = seconds % 60
        return f"{minutes}m {remaining_seconds:.0f}s"
    else:
        hours = int(seconds // 3600)
        remaining_minutes = int((seconds % 3600) // 60)
        return f"{hours}h {remaining_minutes}m"


def calculate_score(values: List[float], weights: Optional[List[float]] = None) -> float:
    """Calculate weighted average score"""
    if not values:
        return 0.0

    if weights is None:
        weights = [1.0] * len(values)

    if len(values) != len(weights):
        raise ValueError("Values and weights must have the same length")

    weighted_sum = sum(v * w for v, w in zip(values, weights))
    total_weight = sum(weights)

    return weighted_sum / total_weight if total_weight > 0 else 0.0


def get_summary_stats(numbers: List[float]) -> Dict[str, float]:
    """Calculate basic statistics"""
    if not numbers:
        return {"mean": 0.0, "min": 0.0, "max": 0.0, "std": 0.0}

    mean = sum(numbers) / len(numbers)
    min_val = min(numbers)
    max_val = max(numbers)

    # Calculate standard deviation
    if len(numbers) > 1:
        variance = sum((x - mean) ** 2 for x in numbers) / (len(numbers) - 1)
        std_dev = variance**0.5
    else:
        std_dev = 0.0

    return {"mean": mean, "min": min_val, "max": max_val, "std": std_dev}


class RateLimiter:
    """Request rate limiter"""

    def __init__(self, max_requests: int = 10, time_window: int = 60):
        self.max_requests = max_requests
        self.time_window = time_window
        self.requests: List[datetime] = []

    def can_make_request(self) -> bool:
        """Check if request can be made"""
        now = datetime.now()

        # Remove old requests
        self.requests = [req_time for req_time in self.requests if (now - req_time).total_seconds() < self.time_window]

        return len(self.requests) < self.max_requests

    def add_request(self):
        """Add request record"""
        if self.can_make_request():
            self.requests.append(datetime.now())
        else:
            raise RateLimitError(f"Rate limit exceeded: {self.max_requests} requests per {self.time_window}s")

    async def wait_if_needed(self):
        """Wait until request can be made"""
        if not self.can_make_request():
            oldest_request = min(self.requests)
            wait_time = self.time_window - (datetime.now() - oldest_request).total_seconds()
            if wait_time > 0:
                logger.info(f"Rate limiting: waiting {wait_time:.1f}s")
                await asyncio.sleep(wait_time)


class RateLimitError(Exception):
    """Rate limit error"""

    pass


def load_hook_timeout() -> int:
    """
    Load Hook timeout setting from .moai/config/config.yaml

    Returns:
        int: timeout value (milliseconds), returns default 5000 if not configured
    """
    try:
        config_path = Path(".moai/config/config.yaml")
        if not config_path.exists():
            return 5000  # Default value

        with open(config_path, "r", encoding="utf-8") as f:
            config = yaml.safe_load(f) or {}

        # Get timeout_ms value from hooks section
        hooks_config = config.get("hooks", {})
        timeout_ms = hooks_config.get("timeout_ms", 5000)

        return int(timeout_ms)
    except (yaml.YAMLError, FileNotFoundError, KeyError, ValueError):
        logger.warning("Failed to load hook timeout from config, using default 5000ms")
        return 5000


def get_graceful_degradation() -> bool:
    """
    Load graceful_degradation setting from .moai/config/config.yaml

    Returns:
        bool: graceful_degradation setting value, returns default True if not configured
    """
    try:
        config_path = Path(".moai/config/config.yaml")
        if not config_path.exists():
            return True  # Default value

        with open(config_path, "r", encoding="utf-8") as f:
            config = yaml.safe_load(f) or {}

        # Get graceful_degradation value from hooks section
        hooks_config = config.get("hooks", {})
        return hooks_config.get("graceful_degradation", True)
    except (yaml.YAMLError, FileNotFoundError, KeyError):
        logger.warning("Failed to load graceful_degradation from config, using default True")
        return True
