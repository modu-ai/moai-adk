# @CODE:USER-EXPERIENCE-IMPROVEMENT-001 | @TEST:USER-EXPERIENCE-IMPROVEMENT-001

"""
User Experience Enhancement Utilities
사용자 경험 개선 유틸리티
"""

import asyncio
import logging
import time
from dataclasses import dataclass, field
from datetime import datetime
from typing import Any, Dict, List, Tuple
from urllib.parse import urljoin

import aiohttp

from moai_adk.utils.common import HTTPClient

logger = logging.getLogger(__name__)


@dataclass
class PerformanceMetrics:
    """성능 지표"""
    load_time: float
    response_time: float
    success_rate: float
    throughput: float
    error_rate: float
    timestamp: datetime = field(default_factory=datetime.now)

    @property
    def is_good(self) -> bool:
        """성능이 좋은지 여부"""
        return (
            self.load_time <= 2.0 and
            self.response_time <= 1.0 and
            self.success_rate >= 0.9 and
            self.throughput >= 10 and
            self.error_rate <= 0.1
        )


@dataclass
class NavigationMetrics:
    """네비게이션 지표"""
    structure_score: float
    link_count: int
    depth: int
    completeness: float
    timestamp: datetime = field(default_factory=datetime.now)

    @property
    def is_good(self) -> bool:
        """네비게이션이 좋은지 여부"""
        return (
            self.structure_score >= 0.8 and
            self.link_count >= 5 and
            self.depth <= 3 and
            self.completeness >= 0.9
        )


@dataclass
class ContentMetrics:
    """콘텐츠 지표"""
    accuracy_score: float
    completeness_score: float
    organization_score: float
    readability_score: float
    timestamp: datetime = field(default_factory=datetime.now)

    @property
    def is_good(self) -> bool:
        """콘텐츠가 좋은지 여부"""
        return (
            self.accuracy_score >= 0.9 and
            self.completeness_score >= 0.9 and
            self.organization_score >= 0.8 and
            self.readability_score >= 0.8
        )


@dataclass
class AccessibilityMetrics:
    """접근성 지표"""
    keyboard_navigation: bool
    screen_reader_support: bool
    color_contrast: bool
    responsive_design: bool
    aria_labels: bool
    timestamp: datetime = field(default_factory=datetime.now)

    @property
    def is_good(self) -> bool:
        """접근성이 좋은지 여부"""
        return all([
            self.keyboard_navigation,
            self.screen_reader_support,
            self.color_contrast,
            self.responsive_design,
            self.aria_labels
        ])


class UserExperienceAnalyzer(HTTPClient):
    """사용자 경험 분석기"""

    def __init__(self, base_url: str, max_workers: int = 5):
        super().__init__(max_concurrent=max_workers, timeout=10)
        self.base_url = base_url

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

    async def analyze_performance(self) -> PerformanceMetrics:
        """성능 분석"""
        # 여러 페이지 동시 로드 테스트
        pages = [
            self.base_url,
            urljoin(self.base_url, "/getting-started"),
            urljoin(self.base_url, "/api"),
            urljoin(self.base_url, "/guides"),
            urljoin(self.base_url, "/search")
        ]

        load_times = []
        success_count = 0
        total_requests = len(pages)

        async def load_page(url: str) -> Tuple[float, bool]:
            page_start = time.time()
            try:
                response = await self.fetch_url(url)
                load_time = time.time() - page_start
                success = response.success
                return load_time, success
            except Exception:
                load_time = time.time() - page_start
                return load_time, False

        # 모든 페이지 동시 로드
        tasks = [load_page(url) for url in pages]
        results = await asyncio.gather(*tasks)

        # 결과 분석
        total_load_time = 0
        success_count = 0

        for load_time, success in results:
            load_times.append(load_time)
            total_load_time += load_time
            if success:
                success_count += 1

        avg_load_time = total_load_time / total_requests if total_requests > 0 else 0
        success_rate = success_count / total_requests if total_requests > 0 else 0

        # 모의 측정값 (실제 구현에서는 실제 지표 사용)
        response_time = 0.5  # 모의 응답 시간
        throughput = 15.0    # 모든 처리량
        error_rate = 1.0 - success_rate

        return PerformanceMetrics(
            load_time=avg_load_time,
            response_time=response_time,
            success_rate=success_rate,
            throughput=throughput,
            error_rate=error_rate
        )

    async def analyze_navigation(self) -> NavigationMetrics:
        """네비게이션 구조 분석"""
        # 모의 네비게이션 데이터 (실제 구현에서는 실제 크롤링 수행)
        navigation_data = {
            "main_links": ["Getting Started", "API Documentation", "Guides", "Search"],
            "sub_links": {
                "Getting Started": ["Installation", "Configuration", "First Steps"],
                "API Documentation": ["Core API", "Authentication", "Webhooks"],
                "Guides": ["Best Practices", "Examples", "Troubleshooting"],
                "Search": ["Advanced Search", "Filters", "Results"]
            },
            "depth": 2,
            "total_links": 15
        }

        # 구조 점수 계산
        structure_score = self._calculate_structure_score(navigation_data)

        return NavigationMetrics(
            structure_score=structure_score,
            link_count=navigation_data["total_links"],
            depth=navigation_data["depth"],
            completeness=0.95
        )

    def _calculate_structure_score(self, navigation_data: Dict) -> float:
        """네비게이션 구조 점수 계산"""
        main_links = len(navigation_data["main_links"])
        sub_links_count = sum(len(links) for links in navigation_data["sub_links"].values())

        # 구조 점수 계산 (링크 균형성과 계층적 구조 고려)
        balance_score = min(1.0, main_links / 4.0)  # 메인 링크 균형성
        hierarchy_score = max(0.5, 1.0 - navigation_data["depth"] / 5.0)  # 계층 깊이
        coverage_score = min(1.0, sub_links_count / 20.0)  # 서브 링크 커버리지

        return (balance_score + hierarchy_score + coverage_score) / 3.0

    async def analyze_content(self) -> ContentMetrics:
        """콘텐츠 품질 분석"""
        # 모의 콘텐츠 데이터 (실제 구현에서는 실제 콘텐츠 분석 수행)
        content_data = {
            "word_count": 5000,
            "code_examples": 25,
            "images": 15,
            "links": 30,
            "readability_score": 8.5,
            "completeness_score": 0.95
        }

        # 정확성 점수 계산
        accuracy_score = self._calculate_accuracy_score(content_data)
        organization_score = self._calculate_organization_score(content_data)
        readability_score = content_data["readability_score"] / 10.0

        return ContentMetrics(
            accuracy_score=accuracy_score,
            completeness_score=content_data["completeness_score"],
            organization_score=organization_score,
            readability_score=readability_score
        )

    def _calculate_accuracy_score(self, content_data: Dict) -> float:
        """콘텐츠 정확성 점수 계산"""
        code_examples_ratio = min(1.0, content_data["code_examples"] / 20.0)
        images_ratio = min(1.0, content_data["images"] / 10.0)
        links_ratio = min(1.0, content_data["links"] / 25.0)

        return (code_examples_ratio + images_ratio + links_ratio) / 3.0

    def _calculate_organization_score(self, content_data: Dict) -> float:
        """콘텐츠 조직 점수 계산"""
        word_count_ratio = min(1.0, content_data["word_count"] / 5000.0)
        structure_score = min(1.0, content_data["code_examples"] / 15.0)

        return (word_count_ratio + structure_score) / 2.0

    async def analyze_accessibility(self) -> AccessibilityMetrics:
        """접근성 분석"""
        # 모의 접근성 데이터 (실제 구현에서는 실제 접근성 검사 수행)
        accessibility_data = {
            "keyboard_navigation": True,
            "screen_reader_support": True,
            "color_contrast": True,
            "responsive_design": True,
            "aria_labels": True
        }

        return AccessibilityMetrics(
            keyboard_navigation=accessibility_data["keyboard_navigation"],
            screen_reader_support=accessibility_data["screen_reader_support"],
            color_contrast=accessibility_data["color_contrast"],
            responsive_design=accessibility_data["responsive_design"],
            aria_labels=accessibility_data["aria_labels"]
        )

    async def generate_report(self) -> Dict[str, Any]:
        """종합 경험 보고서 생성"""
        async with self:
            # 모든 지표 동시 분석
            performance_task = self.analyze_performance()
            navigation_task = self.analyze_navigation()
            content_task = self.analyze_content()
            accessibility_task = self.analyze_accessibility()

            # 모든 분석 병렬 실행
            performance, navigation, content, accessibility = await asyncio.gather(
                performance_task, navigation_task, content_task, accessibility_task
            )

            # 종합 점수 계산
            overall_score = (
                performance.success_rate * 0.3 +
                navigation.structure_score * 0.2 +
                content.accuracy_score * 0.3 +
                (1 if accessibility.is_good else 0) * 0.2
            )

            # 개선 제안 생성
            recommendations = self._generate_recommendations(
                performance, navigation, content, accessibility
            )

            return {
                "overall_score": overall_score,
                "performance": performance,
                "navigation": navigation,
                "content": content,
                "accessibility": accessibility,
                "recommendations": recommendations,
                "generated_at": datetime.now().isoformat()
            }

    def _generate_recommendations(
        self,
        performance: PerformanceMetrics,
        navigation: NavigationMetrics,
        content: ContentMetrics,
        accessibility: AccessibilityMetrics
    ) -> List[str]:
        """개선 제안 생성"""
        recommendations = []

        # 성능 개선 제안
        if not performance.is_good:
            if performance.load_time > 2.0:
                recommendations.append("페이지 로드 시간 개선: 이미지 최적화, CDN 사용, 코드 스플릿 고려")
            if performance.error_rate > 0.1:
                recommendations.append("에러 처리 개선: 404 페이지 개선, 에러 메시지 개선")

        # 네비게이션 개선 제안
        if not navigation.is_good:
            if navigation.structure_score < 0.8:
                recommendations.append("네비게이션 구조 재설계: 계층 구조 단순화, 카테고리 재구성")
            if navigation.completeness < 0.9:
                recommendations.append("링크 완성도 개선: 누락된 페이지 연결, 큰음 링크 추가")

        # 콘텐츠 개선 제안
        if not content.is_good:
            if content.accuracy_score < 0.9:
                recommendations.append("콘텐츠 정확성 개선: 정보 업데이트, 예시 코드 검증")
            if content.organization_score < 0.8:
                recommendations.append("콘텐츠 구조 개선: 섹션 분할, 목차 추가, 관련 링크 추가")

        # 접근성 개선 제안
        if not accessibility.is_good:
            if not accessibility.keyboard_navigation:
                recommendations.append("키보드 네비게이션 개선: 탭 순서 최적화, 키보드 단축키 추가")
            if not accessibility.screen_reader_support:
                recommendations.append("스크린 리더 지원 개선: ARIA 레이블 추가, 시맨틱 HTML 사용")

        return recommendations


def generate_improvement_plan(analysis_report: Dict[str, Any]) -> Dict[str, Any]:
    """개선 계획 생성"""
    overall_score = analysis_report["overall_score"]

    # 우선순위 설정
    priorities = {
        "high": [],
        "medium": [],
        "low": []
    }

    # 성능 우선순위
    performance = analysis_report["performance"]
    if not performance.is_good:
        if performance.error_rate > 0.2:
            priorities["high"].append("에러 처리 시스템 개선")
        elif performance.load_time > 3.0:
            priorities["high"].append("로드 시간 개선")
        else:
            priorities["medium"].append("성능 최적화")

    # 콘텐츠 우선순위
    content = analysis_report["content"]
    if not content.is_good:
        if content.accuracy_score < 0.8:
            priorities["high"].append("콘텐츠 정확성 검증")
        elif content.completeness_score < 0.8:
            priorities["medium"].append("콘텐츠 완성도 개선")
        else:
            priorities["low"].append("콘텐츠 미세 조정")

    # 접근성 우선순위
    accessibility = analysis_report["accessibility"]
    if not accessibility.is_good:
        if not accessibility.keyboard_navigation:
            priorities["high"].append("키보드 접근성 개선")
        elif not accessibility.screen_reader_support:
            priorities["high"].append("스크린 리더 지원 개선")
        else:
            priorities["medium"].append("접근성 표준 준수")

    # 실행 계획 생성
    timeline = {
        "immediate": priorities["high"],
        "short_term": priorities["medium"],
        "long_term": priorities["low"]
    }

    return {
        "overall_score": overall_score,
        "priorities": priorities,
        "timeline": timeline,
        "estimated_duration": (
            f"{len(priorities['high']) + len(priorities['medium']) * 2 + len(priorities['low']) * 3}주"
        ),
        "success_criteria": {
            "performance_score": 0.9,
            "content_score": 0.9,
            "accessibility_score": 1.0,
            "overall_score": 0.85
        }
    }


if __name__ == "__main__":
    # 사용자 경험 분석 실행 예시
    analyzer = UserExperienceAnalyzer("https://adk.mo.ai.kr")

    async def main():
        analysis_report = await analyzer.generate_report()

        print("=== 사용자 경험 분석 보고서 ===")
        print(f"전체 점수: {analysis_report['overall_score']:.2f}")
        print(f"성능 점수: {analysis_report['performance'].success_rate:.2f}")
        print(f"네비게이션 점수: {analysis_report['navigation'].structure_score:.2f}")
        print(f"콘텐츠 점수: {analysis_report['content'].accuracy_score:.2f}")
        print(f"접근성 점수: {1.0 if analysis_report['accessibility'].is_good else 0.0:.2f}")

        print("\n개선 제안:")
        for recommendation in analysis_report['recommendations']:
            print(f"- {recommendation}")

        # 개선 계획 생성
        improvement_plan = generate_improvement_plan(analysis_report)
        print("\n개선 계획:")
        print(f"예상 소요 시간: {improvement_plan['estimated_duration']}")
        print(f"즉시 실행: {improvement_plan['timeline']['immediate']}")
        print(f"단기 실행: {improvement_plan['timeline']['short_term']}")
        print(f"장기 실행: {improvement_plan['timeline']['long_term']}")

    asyncio.run(main())
