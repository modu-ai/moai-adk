# @CODE:PORTAL-LINK-001: 온라인 문서 링크 검증 유틸리티 | @TEST:PORTAL-LINK-001: 링크 검증 테스트

"""
Link Validation Utilities
온라인 문서 링크 검증 유틸리티
"""

import asyncio
import logging
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import List, Optional

from moai_adk.utils.common import HTTPClient, create_report_path, extract_links_from_text, is_valid_url

logger = logging.getLogger(__name__)


@dataclass
class LinkResult:
    """링크 검증 결과"""
    url: str
    status_code: int
    is_valid: bool
    response_time: float
    error_message: Optional[str] = None
    checked_at: datetime = None

    def __post_init__(self):
        if self.checked_at is None:
            self.checked_at = datetime.now()


@dataclass
class ValidationResult:
    """전체 검증 결과"""
    total_links: int
    valid_links: int
    invalid_links: int
    results: List[LinkResult]
    completed_at: datetime = None

    def __post_init__(self):
        if self.completed_at is None:
            self.completed_at = datetime.now()

    @property
    def success_rate(self) -> float:
        """성공률 계산"""
        if self.total_links == 0:
            return 0.0
        return (self.valid_links / self.total_links) * 100


class LinkValidator(HTTPClient):
    """온라인 문서 링크 검증기"""

    def __init__(self, max_concurrent: int = 5, timeout: int = 10):
        super().__init__(max_concurrent, timeout)

    def extract_links_from_file(self, file_path: Path) -> List[str]:
        """파일에서 모든 링크 추출"""
        if not file_path.exists():
            logger.warning(f"파일이 존재하지 않습니다: {file_path}")
            return []

        try:
            content = file_path.read_text(encoding='utf-8')
            base_url = "https://adk.mo.ai.kr"
            links = extract_links_from_text(content, base_url)
            logger.info(f"파일에서 {len(links)}개의 링크를 발견했습니다: {file_path}")
            return links
        except Exception as e:
            logger.error(f"링크 추출 중 오류 발생: {e}")
            return []

    async def validate_link(self, url: str) -> LinkResult:
        """단일 링크 검증"""
        try:
            # URL 유효성 검사
            if not is_valid_url(url):
                return LinkResult(
                    url=url,
                    status_code=0,
                    is_valid=False,
                    response_time=0.0,
                    error_message="Invalid URL format"
                )

            # HTTP 요청
            response = await self.fetch_url(url)

            return LinkResult(
                url=url,
                status_code=response.status_code,
                is_valid=response.success,
                response_time=response.load_time,
                error_message=response.error_message
            )

        except Exception as e:
            return LinkResult(
                url=url,
                status_code=0,
                is_valid=False,
                response_time=0.0,
                error_message=f"Unexpected error: {str(e)}"
            )

    async def validate_all_links(self, links: List[str]) -> ValidationResult:
        """모든 링크 검증"""
        results = []

        # 링크 그룹으로 분할 (동시성 제어)
        semaphore = asyncio.Semaphore(self.max_concurrent)

        async def validate_with_semaphore(link: str):
            async with semaphore:
                result = await self.validate_link(link)
                results.append(result)
                # 진행 상황 로깅
                logger.info(f"검증 완료: {link} -> {result.status_code} ({result.is_valid})")
                return result

        # 모든 링크를 비동기로 검증
        tasks = [validate_with_semaphore(link) for link in links]
        await asyncio.gather(*tasks)

        # 결과 분석
        valid_links = sum(1 for r in results if r.is_valid)
        invalid_links = len(results) - valid_links

        return ValidationResult(
            total_links=len(results),
            valid_links=valid_links,
            invalid_links=invalid_links,
            results=results,
            completed_at=datetime.now()
        )

    def generate_report(self, validation_result: ValidationResult) -> str:
        """검증 보고서 생성"""
        from moai_adk.utils.common import get_summary_stats

        report = []
        report.append("# 온라인 문서 링크 검증 보고서")
        report.append(f"**검증 시간**: {validation_result.completed_at.strftime('%Y-%m-%d %H:%M:%S')}")
        report.append(f"**총 링크 수**: {validation_result.total_links}")
        report.append(f"**유효한 링크**: {validation_result.valid_links}")
        report.append(f"**유효하지 않은 링크**: {validation_result.invalid_links}")
        report.append(f"**성공률**: {validation_result.success_rate:.1f}%")
        report.append("")

        # 통계 정보
        if validation_result.results:
            response_times = [r.response_time for r in validation_result.results]
            stats = get_summary_stats(response_times)
            report.append("## 📊 통계 정보")
            report.append("")
            report.append(f"- 평균 응답 시간: {stats['mean']:.2f}초")
            report.append(f"- 최소 응답 시간: {stats['min']:.2f}초")
            report.append(f"- 최대 응답 시간: {stats['max']:.2f}초")
            report.append(f"- 표준편차: {stats['std']:.2f}초")
            report.append("")

        # 실패한 링크 상세 보고
        if validation_result.invalid_links > 0:
            report.append("## ❌ 실패한 링크")
            report.append("")

            for result in validation_result.results:
                if not result.is_valid:
                    report.append(f"- **{result.url}**")
                    report.append(f"  - 상태 코드: {result.status_code}")
                    report.append(f"  - 응답 시간: {result.response_time:.2f}초")
                    if result.error_message:
                        report.append(f"  - 오류: {result.error_message}")
                    report.append("")

        # 성공한 링크 요약
        if validation_result.valid_links > 0:
            report.append("## ✅ 성공한 링크")
            report.append("")
            report.append(f"총 {validation_result.valid_links}개의 링크가 정상적으로 검증되었습니다.")

        return "\n".join(report)


def validate_readme_links(readme_path: Path = None) -> ValidationResult:
    """README 파일의 모든 링크 검증"""
    if readme_path is None:
        readme_path = Path("README.ko.md")

    validator = LinkValidator(max_concurrent=3, timeout=8)

    # README 파일에서 링크 추출
    links = validator.extract_links_from_file(readme_path)

    if not links:
        logger.warning("검증할 링크가 없습니다")
        return ValidationResult(
            total_links=0,
            valid_links=0,
            invalid_links=0,
            results=[]
        )

    logger.info(f"총 {len(links)}개의 링크를 검증합니다...")

    # 비동기 검증 수행
    result = asyncio.run(validator.validate_all_links(links))

    # 보고서 생성 및 저장
    report = validator.generate_report(result)
    report_path = create_report_path(Path("."), "link_validation")
    report_path.write_text(report, encoding='utf-8')
    logger.info(f"보고서가 저장되었습니다: {report_path}")

    return result


if __name__ == "__main__":
    # README 파일 기준 링크 검증 실행
    result = validate_readme_links()

    # 결과 출력
    validator = LinkValidator()
    report = validator.generate_report(result)
    print(report)

    # 파일로 저장
    report_path = Path("link_validation_report.md")
    report_path.write_text(report, encoding='utf-8')
    print(f"\n보고서가 저장되었습니다: {report_path}")
