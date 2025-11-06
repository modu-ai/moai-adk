# @CODE:COMMON-UTILITIES-001

"""
Common Utilities
공통 유틸리티 함수들
"""

import asyncio
import json
import logging
import re
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional
from urllib.parse import urlparse

import aiohttp

logger = logging.getLogger(__name__)


@dataclass
class HTTPResponse:
    """HTTP 응답 데이터"""
    status_code: int
    url: str
    load_time: float
    success: bool
    error_message: Optional[str] = None
    timestamp: datetime = None

    def __post_init__(self):
        if self.timestamp is None:
            self.timestamp = datetime.now()


class HTTPClient:
    """HTTP 클라이언트 유틸리티"""

    def __init__(self, max_concurrent: int = 5, timeout: int = 10):
        self.max_concurrent = max_concurrent
        self.timeout = timeout
        self.session: Optional[aiohttp.ClientSession] = None

    async def __aenter__(self):
        """비동기 컨텍스트 매니저 진입"""
        connector = aiohttp.TCPConnector(limit=self.max_concurrent)
        timeout = aiohttp.ClientTimeout(total=self.timeout)
        self.session = aiohttp.ClientSession(
            connector=connector,
            timeout=timeout,
            headers={
                "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
            }
        )
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """비동기 컨텍스트 매니저 종료"""
        if self.session:
            await self.session.close()

    async def fetch_url(self, url: str) -> HTTPResponse:
        """단일 URL 가져오기"""
        try:
            start_time = asyncio.get_event_loop().time()
            async with self.session.get(url, allow_redirects=True) as response:
                load_time = asyncio.get_event_loop().time() - start_time
                success = 200 <= response.status < 300
                return HTTPResponse(
                    status_code=response.status,
                    url=str(response.url),
                    load_time=load_time,
                    success=success
                )
        except asyncio.TimeoutError:
            return HTTPResponse(
                status_code=0,
                url=url,
                load_time=self.timeout,
                success=False,
                error_message="Request timeout"
            )
        except aiohttp.ClientError as e:
            return HTTPResponse(
                status_code=0,
                url=url,
                load_time=0.0,
                success=False,
                error_message=f"HTTP client error: {str(e)}"
            )
        except Exception as e:
            return HTTPResponse(
                status_code=0,
                url=url,
                load_time=0.0,
                success=False,
                error_message=f"Unexpected error: {str(e)}"
            )

    async def fetch_urls(self, urls: List[str]) -> List[HTTPResponse]:
        """여러 URL 동시 가져오기"""
        async with self:
            tasks = [self.fetch_url(url) for url in urls]
            return await asyncio.gather(*tasks)


def extract_links_from_text(text: str, base_url: str = None) -> List[str]:
    """텍스트에서 링크 추출"""
    links = []

    # Markdown 링크 패턴: [text](url)
    markdown_pattern = r'\[([^\]]+)\]\(([^)]+)\)'
    markdown_matches = re.findall(markdown_pattern, text)

    for match in markdown_matches:
        url = match[1]
        # 상대 URL을 절대 URL로 변환
        if url.startswith(('http://', 'https://')):
            links.append(url)
        elif base_url and url.startswith('/'):
            links.append(f"{base_url}{url}")
        elif base_url and not url.startswith(('http://', 'https://', '#')):
            links.append(f"{base_url}/{url.rstrip('/')}")

    # 일반 URL 패턴
    url_pattern = r'https?://[^\s<>"\'()]+'
    url_matches = re.findall(url_pattern, text)
    links.extend(url_matches)

    logger.info(f"텍스트에서 {len(links)}개의 링크를 발견했습니다")
    return list(set(links))  # 중복 제거


def is_valid_url(url: str) -> bool:
    """URL 유효성 검사"""
    try:
        result = urlparse(url)
        return all([result.scheme, result.netloc])
    except Exception:
        return False


def create_report_path(base_path: Path, suffix: str = "report") -> Path:
    """보고서 파일 경로 생성"""
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    filename = f"{suffix}_{timestamp}.md"
    return base_path / filename


def format_duration(seconds: float) -> str:
    """시간(초)을 읽기 쉬운 형식으로 변환"""
    if seconds < 1:
        return f"{seconds*1000:.0f}ms"
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
    """가중 평균 점수 계산"""
    if not values:
        return 0.0

    if weights is None:
        weights = [1.0] * len(values)

    if len(values) != len(weights):
        raise ValueError("값과 가중치의 길이가 일치해야 합니다")

    weighted_sum = sum(v * w for v, w in zip(values, weights))
    total_weight = sum(weights)

    return weighted_sum / total_weight if total_weight > 0 else 0.0


def get_summary_stats(numbers: List[float]) -> Dict[str, float]:
    """기통계량 계산"""
    if not numbers:
        return {"mean": 0.0, "min": 0.0, "max": 0.0, "std": 0.0}

    mean = sum(numbers) / len(numbers)
    min_val = min(numbers)
    max_val = max(numbers)

    # 표준편차 계산
    if len(numbers) > 1:
        variance = sum((x - mean) ** 2 for x in numbers) / (len(numbers) - 1)
        std_dev = variance ** 0.5
    else:
        std_dev = 0.0

    return {
        "mean": mean,
        "min": min_val,
        "max": max_val,
        "std": std_dev
    }


class RateLimiter:
    """요청 속도 제한기"""

    def __init__(self, max_requests: int = 10, time_window: int = 60):
        self.max_requests = max_requests
        self.time_window = time_window
        self.requests = []

    def can_make_request(self) -> bool:
        """요청 가능 여부 확인"""
        now = datetime.now()

        # 오래된 요청 제거
        self.requests = [req_time for req_time in self.requests
                        if (now - req_time).total_seconds() < self.time_window]

        return len(self.requests) < self.max_requests

    def add_request(self):
        """요청 기록 추가"""
        if self.can_make_request():
            self.requests.append(datetime.now())
        else:
            raise RateLimitError(f"Rate limit exceeded: {self.max_requests} requests per {self.time_window}s")

    async def wait_if_needed(self):
        """요청이 가능할 때까지 대기"""
        if not self.can_make_request():
            oldest_request = min(self.requests)
            wait_time = self.time_window - (datetime.now() - oldest_request).total_seconds()
            if wait_time > 0:
                logger.info(f"Rate limiting: waiting {wait_time:.1f}s")
                await asyncio.sleep(wait_time)


class RateLimitError(Exception):
    """속도 제한 오류"""
    pass


def load_hook_timeout() -> int:
    """
    .moai/config.json에서 Hook timeout 설정 로드

    Returns:
        int: timeout 값 (밀리초), 설정이 없으면 기본값 5000 반환
    """
    try:
        config_path = Path(".moai/config.json")
        if not config_path.exists():
            return 5000  # 기본값

        with open(config_path, 'r', encoding='utf-8') as f:
            config = json.load(f)

        # hooks 섹션에서 timeout_ms 값 가져오기
        hooks_config = config.get("hooks", {})
        timeout_ms = hooks_config.get("timeout_ms", 5000)

        return int(timeout_ms)
    except (json.JSONDecodeError, FileNotFoundError, KeyError, ValueError):
        logger.warning("Failed to load hook timeout from config, using default 5000ms")
        return 5000


def get_graceful_degradation() -> bool:
    """
    .moai/config.json에서 graceful_degradation 설정 로드

    Returns:
        bool: graceful_degradation 설정값, 설정이 없으면 기본값 True 반환
    """
    try:
        config_path = Path(".moai/config.json")
        if not config_path.exists():
            return True  # 기본값

        with open(config_path, 'r', encoding='utf-8') as f:
            config = json.load(f)

        # hooks 섹션에서 graceful_degradation 값 가져오기
        hooks_config = config.get("hooks", {})
        return hooks_config.get("graceful_degradation", True)
    except (json.JSONDecodeError, FileNotFoundError, KeyError):
        logger.warning("Failed to load graceful_degradation from config, using default True")
        return True
