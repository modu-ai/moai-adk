"""
Link Validation Utilities
Online documentation link validation utilities
"""

import asyncio
import logging
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import List, Optional

from moai_adk.utils.common import (
    HTTPClient,
    create_report_path,
    extract_links_from_text,
    is_valid_url,
)
from moai_adk.utils.safe_file_reader import SafeFileReader

logger = logging.getLogger(__name__)


@dataclass
class LinkResult:
    """Link validation result"""

    url: str
    status_code: int
    is_valid: bool
    response_time: float
    error_message: Optional[str] = None
    checked_at: datetime = field(default_factory=datetime.now)

    def __post_init__(self):
        if self.checked_at is None:
            self.checked_at = datetime.now()


@dataclass
class ValidationResult:
    """Overall validation result"""

    total_links: int
    valid_links: int
    invalid_links: int
    results: List[LinkResult]
    completed_at: datetime = field(default_factory=datetime.now)

    def __post_init__(self):
        if self.completed_at is None:
            self.completed_at = datetime.now()

    @property
    def success_rate(self) -> float:
        """Calculate success rate"""
        if self.total_links == 0:
            return 0.0
        return (self.valid_links / self.total_links) * 100


class LinkValidator(HTTPClient):
    """Online documentation link validator"""

    def __init__(self, max_concurrent: int = 5, timeout: int = 10):
        super().__init__(max_concurrent, timeout)

    def extract_links_from_file(self, file_path: Path) -> List[str]:
        """Extract all links from file (using safe file reading)"""
        if not file_path.exists():
            logger.warning(f"File does not exist: {file_path}")
            return []

        try:
            reader = SafeFileReader()
            content = reader.read_text(file_path)
            if content is None:
                logger.error(f"Unable to read file: {file_path}")
                return []

            base_url = "https://adk.mo.ai.kr"
            links = extract_links_from_text(content, base_url)
            logger.info(f"Found {len(links)} links in file: {file_path}")
            return links
        except Exception as e:
            logger.error(f"Error during link extraction: {e}")
            return []

    async def validate_link(self, url: str) -> LinkResult:
        """Validate single link"""
        try:
            # URL validity check
            if not is_valid_url(url):
                return LinkResult(
                    url=url,
                    status_code=0,
                    is_valid=False,
                    response_time=0.0,
                    error_message="Invalid URL format",
                )

            # HTTP request
            response = await self.fetch_url(url)

            return LinkResult(
                url=url,
                status_code=response.status_code,
                is_valid=response.success,
                response_time=response.load_time,
                error_message=response.error_message,
            )

        except Exception as e:
            return LinkResult(
                url=url,
                status_code=0,
                is_valid=False,
                response_time=0.0,
                error_message=f"Unexpected error: {str(e)}",
            )

    async def validate_all_links(self, links: List[str]) -> ValidationResult:
        """Validate all links"""
        results = []

        # Split into link groups (concurrency control)
        semaphore = asyncio.Semaphore(self.max_concurrent)

        async def validate_with_semaphore(link: str):
            async with semaphore:
                result = await self.validate_link(link)
                results.append(result)
                # Log progress
                logger.info(f"Validation complete: {link} -> {result.status_code} ({result.is_valid})")
                return result

        # Validate all links asynchronously
        tasks = [validate_with_semaphore(link) for link in links]
        await asyncio.gather(*tasks)

        # Analyze results
        valid_links = sum(1 for r in results if r.is_valid)
        invalid_links = len(results) - valid_links

        return ValidationResult(
            total_links=len(results),
            valid_links=valid_links,
            invalid_links=invalid_links,
            results=results,
            completed_at=datetime.now(),
        )

    def generate_report(self, validation_result: ValidationResult) -> str:
        """Generate validation report"""
        from moai_adk.utils.common import get_summary_stats

        report = []
        report.append("# Online Documentation Link Validation Report")
        report.append(f"**Validation Time**: {validation_result.completed_at.strftime('%Y-%m-%d %H:%M:%S')}")
        report.append(f"**Total Links**: {validation_result.total_links}")
        report.append(f"**Valid Links**: {validation_result.valid_links}")
        report.append(f"**Invalid Links**: {validation_result.invalid_links}")
        report.append(f"**Success Rate**: {validation_result.success_rate:.1f}%")
        report.append("")

        # Statistics
        if validation_result.results:
            response_times = [r.response_time for r in validation_result.results]
            stats = get_summary_stats(response_times)
            report.append("## 📊 Statistics")
            report.append("")
            report.append(f"- Average Response Time: {stats['mean']:.2f}s")
            report.append(f"- Minimum Response Time: {stats['min']:.2f}s")
            report.append(f"- Maximum Response Time: {stats['max']:.2f}s")
            report.append(f"- Standard Deviation: {stats['std']:.2f}s")
            report.append("")

        # Failed links detailed report
        if validation_result.invalid_links > 0:
            report.append("## ❌ Failed Links")
            report.append("")

            for result in validation_result.results:
                if not result.is_valid:
                    report.append(f"- **{result.url}**")
                    report.append(f"  - Status Code: {result.status_code}")
                    report.append(f"  - Response Time: {result.response_time:.2f}s")
                    if result.error_message:
                        report.append(f"  - Error: {result.error_message}")
                    report.append("")

        # Successful links summary
        if validation_result.valid_links > 0:
            report.append("## ✅ Successful Links")
            report.append("")
            report.append(f"Total of {validation_result.valid_links} links validated successfully.")

        return "\n".join(report)


def validate_readme_links(readme_path: Optional[Path] = None) -> ValidationResult:
    """Validate all links in README file"""
    if readme_path is None:
        readme_path = Path("README.ko.md")

    validator = LinkValidator(max_concurrent=3, timeout=8)

    # Extract links from README file
    links = validator.extract_links_from_file(readme_path)

    if not links:
        logger.warning("No links to validate")
        return ValidationResult(total_links=0, valid_links=0, invalid_links=0, results=[])

    logger.info(f"Validating total of {len(links)} links...")

    # Perform asynchronous validation
    result = asyncio.run(validator.validate_all_links(links))

    # Generate and save report
    report = validator.generate_report(result)
    report_path = create_report_path(Path("."), "link_validation")
    report_path.write_text(report, encoding="utf-8")
    logger.info(f"Report saved to: {report_path}")

    return result


if __name__ == "__main__":
    # Execute README file link validation
    result = validate_readme_links()

    # Print results
    validator = LinkValidator()
    report = validator.generate_report(result)
    print(report)

    # Save to file
    report_path = Path("link_validation_report.md")
    report_path.write_text(report, encoding="utf-8")
    print(f"\nReport saved to: {report_path}")
