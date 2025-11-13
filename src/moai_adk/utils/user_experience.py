"""
User Experience Enhancement Utilities

This module provides utilities for analyzing and improving user experience
across multiple dimensions including performance, navigation, content quality,
and accessibility.
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
    """Performance metrics for user experience analysis"""

    load_time: float
    response_time: float
    success_rate: float
    throughput: float
    error_rate: float
    timestamp: datetime = field(default_factory=datetime.now)

    @property
    def is_good(self) -> bool:
        """Check if performance meets quality thresholds"""
        return (
            self.load_time <= 2.0
            and self.response_time <= 1.0
            and self.success_rate >= 0.9
            and self.throughput >= 10
            and self.error_rate <= 0.1
        )


@dataclass
class NavigationMetrics:
    """Navigation structure metrics for UX analysis"""

    structure_score: float
    link_count: int
    depth: int
    completeness: float
    timestamp: datetime = field(default_factory=datetime.now)

    @property
    def is_good(self) -> bool:
        """Check if navigation structure meets quality thresholds"""
        return (
            self.structure_score >= 0.8
            and self.link_count >= 5
            and self.depth <= 3
            and self.completeness >= 0.9
        )


@dataclass
class ContentMetrics:
    """Content quality metrics for UX analysis"""

    accuracy_score: float
    completeness_score: float
    organization_score: float
    readability_score: float
    timestamp: datetime = field(default_factory=datetime.now)

    @property
    def is_good(self) -> bool:
        """Check if content quality meets quality thresholds"""
        return (
            self.accuracy_score >= 0.9
            and self.completeness_score >= 0.9
            and self.organization_score >= 0.8
            and self.readability_score >= 0.8
        )


@dataclass
class AccessibilityMetrics:
    """Accessibility metrics for UX analysis"""

    keyboard_navigation: bool
    screen_reader_support: bool
    color_contrast: bool
    responsive_design: bool
    aria_labels: bool
    timestamp: datetime = field(default_factory=datetime.now)

    @property
    def is_good(self) -> bool:
        """Check if accessibility meets quality thresholds"""
        return all(
            [
                self.keyboard_navigation,
                self.screen_reader_support,
                self.color_contrast,
                self.responsive_design,
                self.aria_labels,
            ]
        )


class UserExperienceAnalyzer(HTTPClient):
    """
    User experience analyzer for comprehensive UX assessment

    Analyzes multiple dimensions of user experience including performance,
    navigation, content quality, and accessibility. Provides actionable
    recommendations for improvement.
    """

    def __init__(self, base_url: str, max_workers: int = 5):
        super().__init__(max_concurrent=max_workers, timeout=10)
        self.base_url = base_url

    async def __aenter__(self):
        """Enter async context manager"""
        connector = aiohttp.TCPConnector(limit=self.max_concurrent)
        timeout = aiohttp.ClientTimeout(total=self.timeout)
        self.session = aiohttp.ClientSession(
            connector=connector,
            timeout=timeout,
            headers={
                "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
            },
        )
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Exit async context manager"""
        if self.session:
            await self.session.close()

    async def analyze_performance(self) -> PerformanceMetrics:
        """
        Analyze website performance metrics

        Tests multiple pages concurrently to measure load times, success rates,
        throughput, and error rates.
        """
        # Test multiple pages with concurrent loading
        pages = [
            self.base_url,
            urljoin(self.base_url, "/getting-started"),
            urljoin(self.base_url, "/api"),
            urljoin(self.base_url, "/guides"),
            urljoin(self.base_url, "/search"),
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

        # Load all pages concurrently
        tasks = [load_page(url) for url in pages]
        results = await asyncio.gather(*tasks)

        # Analyze results
        total_load_time = 0.0
        success_count = 0

        for load_time, success in results:
            load_times.append(load_time)
            total_load_time += load_time
            if success:
                success_count += 1

        avg_load_time = total_load_time / total_requests if total_requests > 0 else 0
        success_rate = success_count / total_requests if total_requests > 0 else 0

        # Mock measurements (actual implementation would use real metrics)
        response_time = 0.5  # Mock response time
        throughput = 15.0  # Mock throughput
        error_rate = 1.0 - success_rate

        return PerformanceMetrics(
            load_time=avg_load_time,
            response_time=response_time,
            success_rate=success_rate,
            throughput=throughput,
            error_rate=error_rate,
        )

    async def analyze_navigation(self) -> NavigationMetrics:
        """
        Analyze navigation structure

        Evaluates navigation hierarchy, link distribution, and completeness
        to assess ease of navigation.
        """
        # Mock navigation data (actual implementation would perform real crawling)
        navigation_data: Dict[str, Any] = {
            "main_links": ["Getting Started", "API Documentation", "Guides", "Search"],
            "sub_links": {
                "Getting Started": ["Installation", "Configuration", "First Steps"],
                "API Documentation": ["Core API", "Authentication", "Webhooks"],
                "Guides": ["Best Practices", "Examples", "Troubleshooting"],
                "Search": ["Advanced Search", "Filters", "Results"],
            },
            "depth": 2,
            "total_links": 15,
        }

        # Calculate structure score
        structure_score = self._calculate_structure_score(navigation_data)

        return NavigationMetrics(
            structure_score=structure_score,
            link_count=int(navigation_data["total_links"]),
            depth=int(navigation_data["depth"]),
            completeness=0.95,
        )

    def _calculate_structure_score(self, navigation_data: Dict) -> float:
        """
        Calculate navigation structure score

        Considers link balance and hierarchical structure to determine
        overall navigation quality.
        """
        main_links = len(navigation_data["main_links"])
        sub_links_count = sum(
            len(links) for links in navigation_data["sub_links"].values()
        )

        # Calculate structure score (considering link balance and hierarchy)
        balance_score = min(1.0, main_links / 4.0)  # Main link balance
        hierarchy_score = max(
            0.5, 1.0 - navigation_data["depth"] / 5.0
        )  # Hierarchy depth
        coverage_score = min(1.0, sub_links_count / 20.0)  # Sub-link coverage

        return (balance_score + hierarchy_score + coverage_score) / 3.0

    async def analyze_content(self) -> ContentMetrics:
        """
        Analyze content quality

        Evaluates content accuracy, completeness, organization, and readability
        to assess overall content quality.
        """
        # Mock content data (actual implementation would perform real content analysis)
        content_data = {
            "word_count": 5000,
            "code_examples": 25,
            "images": 15,
            "links": 30,
            "readability_score": 8.5,
            "completeness_score": 0.95,
        }

        # Calculate accuracy score
        accuracy_score = self._calculate_accuracy_score(content_data)
        organization_score = self._calculate_organization_score(content_data)
        readability_score = content_data["readability_score"] / 10.0

        return ContentMetrics(
            accuracy_score=accuracy_score,
            completeness_score=content_data["completeness_score"],
            organization_score=organization_score,
            readability_score=readability_score,
        )

    def _calculate_accuracy_score(self, content_data: Dict) -> float:
        """
        Calculate content accuracy score

        Evaluates the presence and quality of code examples, images, and links
        to determine content accuracy.
        """
        code_examples_ratio = min(1.0, content_data["code_examples"] / 20.0)
        images_ratio = min(1.0, content_data["images"] / 10.0)
        links_ratio = min(1.0, content_data["links"] / 25.0)

        return (code_examples_ratio + images_ratio + links_ratio) / 3.0

    def _calculate_organization_score(self, content_data: Dict) -> float:
        """
        Calculate content organization score

        Evaluates content structure and word count to determine organization quality.
        """
        word_count_ratio = min(1.0, content_data["word_count"] / 5000.0)
        structure_score = min(1.0, content_data["code_examples"] / 15.0)

        return (word_count_ratio + structure_score) / 2.0

    async def analyze_accessibility(self) -> AccessibilityMetrics:
        """
        Analyze accessibility compliance

        Evaluates keyboard navigation, screen reader support, color contrast,
        responsive design, and ARIA labels to assess accessibility.
        """
        # Mock accessibility data (actual implementation would perform real accessibility checks)
        accessibility_data = {
            "keyboard_navigation": True,
            "screen_reader_support": True,
            "color_contrast": True,
            "responsive_design": True,
            "aria_labels": True,
        }

        return AccessibilityMetrics(
            keyboard_navigation=accessibility_data["keyboard_navigation"],
            screen_reader_support=accessibility_data["screen_reader_support"],
            color_contrast=accessibility_data["color_contrast"],
            responsive_design=accessibility_data["responsive_design"],
            aria_labels=accessibility_data["aria_labels"],
        )

    async def generate_report(self) -> Dict[str, Any]:
        """
        Generate comprehensive user experience report

        Analyzes all UX dimensions concurrently and provides an overall score
        with actionable recommendations for improvement.
        """
        async with self:
            # Analyze all metrics concurrently
            performance_task = self.analyze_performance()
            navigation_task = self.analyze_navigation()
            content_task = self.analyze_content()
            accessibility_task = self.analyze_accessibility()

            # Execute all analyses in parallel
            performance, navigation, content, accessibility = await asyncio.gather(
                performance_task, navigation_task, content_task, accessibility_task
            )

            # Calculate overall score
            overall_score = (
                performance.success_rate * 0.3
                + navigation.structure_score * 0.2
                + content.accuracy_score * 0.3
                + (1 if accessibility.is_good else 0) * 0.2
            )

            # Generate improvement recommendations
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
                "generated_at": datetime.now().isoformat(),
            }

    def _generate_recommendations(
        self,
        performance: PerformanceMetrics,
        navigation: NavigationMetrics,
        content: ContentMetrics,
        accessibility: AccessibilityMetrics,
    ) -> List[str]:
        """
        Generate improvement recommendations

        Analyzes all metrics to provide specific, actionable recommendations
        for improving user experience.
        """
        recommendations = []

        # Performance improvement recommendations
        if not performance.is_good:
            if performance.load_time > 2.0:
                recommendations.append(
                    "Improve page load time: Optimize images, use CDN, consider code splitting"
                )
            if performance.error_rate > 0.1:
                recommendations.append(
                    "Improve error handling: Enhance 404 pages, improve error messages"
                )

        # Navigation improvement recommendations
        if not navigation.is_good:
            if navigation.structure_score < 0.8:
                recommendations.append(
                    "Redesign navigation structure: Simplify hierarchy, reorganize categories"
                )
            if navigation.completeness < 0.9:
                recommendations.append(
                    "Improve link completeness: Connect missing pages, add breadcrumb links"
                )

        # Content improvement recommendations
        if not content.is_good:
            if content.accuracy_score < 0.9:
                recommendations.append(
                    "Improve content accuracy: Update information, validate code examples"
                )
            if content.organization_score < 0.8:
                recommendations.append(
                    "Improve content structure: Add section divisions, table of contents, related links"
                )

        # Accessibility improvement recommendations
        if not accessibility.is_good:
            if not accessibility.keyboard_navigation:
                recommendations.append(
                    "Improve keyboard navigation: Optimize tab order, add keyboard shortcuts"
                )
            if not accessibility.screen_reader_support:
                recommendations.append(
                    "Improve screen reader support: Add ARIA labels, use semantic HTML"
                )

        return recommendations


def generate_improvement_plan(analysis_report: Dict[str, Any]) -> Dict[str, Any]:
    """
    Generate improvement plan from analysis report

    Creates a prioritized action plan based on UX analysis results, categorizing
    improvements by priority and timeline.
    """
    overall_score = analysis_report["overall_score"]

    # Set priorities
    priorities: Dict[str, List[str]] = {"high": [], "medium": [], "low": []}

    # Performance priorities
    performance = analysis_report["performance"]
    if not performance.is_good:
        if performance.error_rate > 0.2:
            priorities["high"].append("Improve error handling system")
        elif performance.load_time > 3.0:
            priorities["high"].append("Improve load time")
        else:
            priorities["medium"].append("Optimize performance")

    # Content priorities
    content = analysis_report["content"]
    if not content.is_good:
        if content.accuracy_score < 0.8:
            priorities["high"].append("Validate content accuracy")
        elif content.completeness_score < 0.8:
            priorities["medium"].append("Improve content completeness")
        else:
            priorities["low"].append("Fine-tune content")

    # Accessibility priorities
    accessibility = analysis_report["accessibility"]
    if not accessibility.is_good:
        if not accessibility.keyboard_navigation:
            priorities["high"].append("Improve keyboard accessibility")
        elif not accessibility.screen_reader_support:
            priorities["high"].append("Improve screen reader support")
        else:
            priorities["medium"].append("Ensure accessibility standards compliance")

    # Generate execution plan
    timeline = {
        "immediate": priorities["high"],
        "short_term": priorities["medium"],
        "long_term": priorities["low"],
    }

    return {
        "overall_score": overall_score,
        "priorities": priorities,
        "timeline": timeline,
        "estimated_duration": (
            f"{len(priorities['high']) + len(priorities['medium']) * 2 + len(priorities['low']) * 3} weeks"
        ),
        "success_criteria": {
            "performance_score": 0.9,
            "content_score": 0.9,
            "accessibility_score": 1.0,
            "overall_score": 0.85,
        },
    }


if __name__ == "__main__":
    # User experience analysis execution example
    analyzer = UserExperienceAnalyzer("https://adk.mo.ai.kr")

    async def main():
        analysis_report = await analyzer.generate_report()

        print("=== User Experience Analysis Report ===")
        print(f"Overall Score: {analysis_report['overall_score']:.2f}")
        print(f"Performance Score: {analysis_report['performance'].success_rate:.2f}")
        print(f"Navigation Score: {analysis_report['navigation'].structure_score:.2f}")
        print(f"Content Score: {analysis_report['content'].accuracy_score:.2f}")
        print(
            f"Accessibility Score: {1.0 if analysis_report['accessibility'].is_good else 0.0:.2f}"
        )

        print("\nImprovement Recommendations:")
        for recommendation in analysis_report["recommendations"]:
            print(f"- {recommendation}")

        # Generate improvement plan
        improvement_plan = generate_improvement_plan(analysis_report)
        print("\nImprovement Plan:")
        print(f"Estimated Duration: {improvement_plan['estimated_duration']}")
        print(f"Immediate Actions: {improvement_plan['timeline']['immediate']}")
        print(f"Short-term Actions: {improvement_plan['timeline']['short_term']}")
        print(f"Long-term Actions: {improvement_plan['timeline']['long_term']}")

    asyncio.run(main())
