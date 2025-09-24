"""
@TEST:UNIT-REPORT-GENERATOR - TAG 추적성 리포트 생성 테스트

RED 단계: Jinja2 템플릿 기반 Markdown/HTML 리포트 생성 실패 테스트
"""

import pytest
import tempfile
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Any

from moai_adk.core.tag_system.report_generator import TagReportGenerator, ReportFormat, TraceabilityReport
from moai_adk.core.tag_system.parser import TagMatch
from moai_adk.core.tag_system.validator import ChainValidationResult


class TestTagReportGenerator:
    """TAG 추적성 리포트 생성 테스트 스위트"""

    def setup_method(self):
        """각 테스트 전 초기화"""
        self.temp_dir = Path(tempfile.mkdtemp())
        self.generator = TagReportGenerator(output_dir=self.temp_dir)

    def teardown_method(self):
        """각 테스트 후 정리"""
        import shutil
        shutil.rmtree(self.temp_dir, ignore_errors=True)

    def test_should_generate_tag_chain_matrix_report(self):
        """
        Given: TAG 체인 데이터
        When: 체인 매트릭스 리포트를 생성할 때
        Then: 구현 상태 매트릭스가 올바르게 생성되어야 함
        """
        # GIVEN: Primary Chain 데이터
        chain_data = [
            TagMatch(category="REQ", identifier="USER-AUTH-001", description="사용자 인증"),
            TagMatch(category="DESIGN", identifier="JWT-DESIGN-001", description="JWT 토큰 설계"),
            TagMatch(category="TASK", identifier="API-LOGIN-001", description="로그인 API 구현"),
            TagMatch(category="TEST", identifier="UNIT-AUTH-001", description="인증 테스트")
        ]

        # WHEN: 체인 매트릭스 리포트 생성
        matrix_report = self.generator.generate_chain_matrix(chain_data)

        # THEN: 올바른 매트릭스 구조 생성
        assert "PRIMARY" in matrix_report
        primary_chain = matrix_report["PRIMARY"]

        assert len(primary_chain["REQ"]) == 1
        assert "USER-AUTH-001" in primary_chain["REQ"]

        assert len(primary_chain["DESIGN"]) == 1
        assert "JWT-DESIGN-001" in primary_chain["DESIGN"]

        assert len(primary_chain["TASK"]) == 1
        assert "API-LOGIN-001" in primary_chain["TASK"]

        assert len(primary_chain["TEST"]) == 1
        assert "UNIT-AUTH-001" in primary_chain["TEST"]

    def test_should_identify_missing_tag_connections(self):
        """
        Given: 불완전한 TAG 체인
        When: 누락된 연결을 분석할 때
        Then: 누락된 TAG 연결이 식별되어야 함
        """
        # GIVEN: DESIGN이 누락된 불완전한 체인
        incomplete_chain = [
            TagMatch(category="REQ", identifier="USER-PAYMENT-001", description="결제 기능 요구사항"),
            # DESIGN 누락
            TagMatch(category="TASK", identifier="API-PAYMENT-001", description="결제 API 구현"),
            TagMatch(category="TEST", identifier="UNIT-PAYMENT-001", description="결제 테스트")
        ]

        # WHEN: 누락 분석 실행
        missing_report = self.generator.analyze_missing_connections(incomplete_chain)

        # THEN: DESIGN 누락 식별
        assert len(missing_report["missing_links"]) == 1
        missing_link = missing_report["missing_links"][0]
        assert missing_link["category"] == "DESIGN"
        assert missing_link["expected_between"] == ["USER-PAYMENT-001", "API-PAYMENT-001"]

    def test_should_generate_markdown_traceability_report(self):
        """
        Given: TAG 데이터와 분석 결과
        When: Markdown 형식으로 추적성 리포트를 생성할 때
        Then: 올바른 Markdown 구조의 리포트가 생성되어야 함
        """
        # GIVEN: 완전한 TAG 체인 데이터
        tag_data = [
            TagMatch(category="REQ", identifier="USER-SEARCH-001", description="검색 기능"),
            TagMatch(category="DESIGN", identifier="ELASTIC-DESIGN-001", description="Elasticsearch 설계"),
            TagMatch(category="TASK", identifier="API-SEARCH-001", description="검색 API 구현"),
            TagMatch(category="TEST", identifier="E2E-SEARCH-001", description="검색 E2E 테스트")
        ]

        # WHEN: Markdown 리포트 생성
        markdown_content = self.generator.generate_report(
            tags=tag_data,
            format=ReportFormat.MARKDOWN,
            title="TAG 추적성 리포트"
        )

        # THEN: 올바른 Markdown 구조
        assert "# TAG 추적성 리포트" in markdown_content
        assert "## Primary Chain 분석" in markdown_content
        assert "### REQ → DESIGN → TASK → TEST" in markdown_content
        assert "USER-SEARCH-001" in markdown_content
        assert "ELASTIC-DESIGN-001" in markdown_content
        assert "API-SEARCH-001" in markdown_content
        assert "E2E-SEARCH-001" in markdown_content

    def test_should_generate_html_traceability_report(self):
        """
        Given: TAG 데이터와 분석 결과
        When: HTML 형식으로 추적성 리포트를 생성할 때
        Then: 올바른 HTML 구조의 리포트가 생성되어야 함
        """
        # GIVEN: TAG 데이터
        tag_data = [
            TagMatch(category="REQ", identifier="USER-NOTIFICATION-001", description="알림 기능"),
            TagMatch(category="DESIGN", identifier="PUSH-DESIGN-001", description="푸시 알림 설계"),
        ]

        # WHEN: HTML 리포트 생성
        html_content = self.generator.generate_report(
            tags=tag_data,
            format=ReportFormat.HTML,
            title="HTML 추적성 리포트"
        )

        # THEN: 올바른 HTML 구조
        assert "<html>" in html_content
        assert "<head>" in html_content
        assert "<title>HTML 추적성 리포트</title>" in html_content
        assert "<body>" in html_content
        assert "USER-NOTIFICATION-001" in html_content
        assert "PUSH-DESIGN-001" in html_content
        assert "</html>" in html_content

    def test_should_calculate_implementation_coverage_percentage(self):
        """
        Given: 다양한 구현 상태의 TAG들
        When: 구현 완료율을 계산할 때
        Then: 정확한 커버리지 비율이 계산되어야 함
        """
        # GIVEN: 부분적으로 구현된 TAG들
        mixed_implementation = [
            # 완전 구현된 체인 (100%)
            TagMatch(category="REQ", identifier="USER-LOGIN-001"),
            TagMatch(category="DESIGN", identifier="AUTH-DESIGN-001"),
            TagMatch(category="TASK", identifier="API-AUTH-001"),
            TagMatch(category="TEST", identifier="UNIT-AUTH-001"),

            # 부분 구현된 체인 (50% - TASK, TEST 누락)
            TagMatch(category="REQ", identifier="USER-PROFILE-001"),
            TagMatch(category="DESIGN", identifier="PROFILE-DESIGN-001"),
        ]

        # WHEN: 커버리지 계산
        coverage_report = self.generator.calculate_implementation_coverage(mixed_implementation)

        # THEN: 정확한 커버리지 계산
        assert coverage_report["total_chains"] == 2
        assert coverage_report["complete_chains"] == 1
        assert coverage_report["coverage_percentage"] == 50.0  # 1/2 = 50%

        # 카테고리별 상세 분석
        assert coverage_report["category_coverage"]["REQ"] == 100.0  # 2/2
        assert coverage_report["category_coverage"]["DESIGN"] == 100.0  # 2/2
        assert coverage_report["category_coverage"]["TASK"] == 50.0  # 1/2
        assert coverage_report["category_coverage"]["TEST"] == 50.0  # 1/2

    def test_should_export_report_to_file(self):
        """
        Given: 생성된 리포트 콘텐츠
        When: 파일로 내보내기를 실행할 때
        Then: 지정된 경로에 올바른 파일이 생성되어야 함
        """
        # GIVEN: 리포트 콘텐츠
        tag_data = [
            TagMatch(category="REQ", identifier="USER-EXPORT-001", description="내보내기 기능")
        ]

        # WHEN: Markdown 파일로 내보내기
        output_file = self.temp_dir / "traceability_report.md"
        self.generator.export_to_file(
            tags=tag_data,
            output_path=output_file,
            format=ReportFormat.MARKDOWN
        )

        # THEN: 파일이 올바르게 생성됨
        assert output_file.exists()

        content = output_file.read_text(encoding="utf-8")
        assert "USER-EXPORT-001" in content
        assert "내보내기 기능" in content

    def test_should_handle_empty_tag_data_gracefully(self):
        """
        Given: 빈 TAG 데이터
        When: 리포트 생성을 시도할 때
        Then: 오류 없이 빈 리포트가 생성되어야 함
        """
        # GIVEN: 빈 TAG 데이터
        empty_tags = []

        # WHEN: 빈 데이터로 리포트 생성
        empty_report = self.generator.generate_report(
            tags=empty_tags,
            format=ReportFormat.MARKDOWN,
            title="빈 데이터 리포트"
        )

        # THEN: 오류 없이 기본 구조 생성
        assert "# 빈 데이터 리포트" in empty_report
        assert "TAG가 발견되지 않았습니다" in empty_report or "No tags found" in empty_report

    def test_should_use_custom_jinja2_templates(self):
        """
        Given: 사용자 정의 Jinja2 템플릿
        When: 커스텀 템플릿으로 리포트를 생성할 때
        Then: 사용자 정의 템플릿이 적용되어야 함
        """
        # GIVEN: 커스텀 템플릿 설정
        custom_template_dir = self.temp_dir / "templates"
        custom_template_dir.mkdir()

        custom_markdown_template = custom_template_dir / "custom_report.md.j2"
        custom_markdown_template.write_text("""
# 커스텀 {{ title }}

{% for tag in tags %}
- {{ tag.category }}:{{ tag.identifier }} - {{ tag.description }}
{% endfor %}

Generated on: {{ timestamp }}
        """.strip())

        # WHEN: 커스텀 템플릿으로 리포트 생성
        custom_generator = TagReportGenerator(
            output_dir=self.temp_dir,
            template_dir=custom_template_dir
        )

        tag_data = [
            TagMatch(category="REQ", identifier="USER-CUSTOM-001", description="커스텀 기능")
        ]

        custom_report = custom_generator.generate_report(
            tags=tag_data,
            format=ReportFormat.MARKDOWN,
            template_name="custom_report.md.j2",
            title="TAG 리포트"
        )

        # THEN: 커스텀 템플릿 적용됨
        assert "# 커스텀 TAG 리포트" in custom_report
        assert "REQ:USER-CUSTOM-001 - 커스텀 기능" in custom_report
        assert "Generated on:" in custom_report

    def test_should_generate_summary_statistics(self):
        """
        Given: 다양한 TAG 데이터
        When: 요약 통계를 생성할 때
        Then: 정확한 통계 정보가 제공되어야 함
        """
        # GIVEN: 다양한 카테고리의 TAG들
        diverse_tags = [
            TagMatch(category="REQ", identifier="USER-STATS-001"),
            TagMatch(category="REQ", identifier="USER-STATS-002"),
            TagMatch(category="DESIGN", identifier="ARCH-STATS-001"),
            TagMatch(category="TASK", identifier="IMPL-STATS-001"),
            TagMatch(category="FEATURE", identifier="SYSTEM-STATS-001"),
            TagMatch(category="PERF", identifier="API-500MS")
        ]

        # WHEN: 요약 통계 생성
        stats = self.generator.generate_summary_statistics(diverse_tags)

        # THEN: 정확한 통계
        assert stats["total_tags"] == 6
        assert stats["category_breakdown"]["PRIMARY"]["REQ"] == 2
        assert stats["category_breakdown"]["PRIMARY"]["DESIGN"] == 1
        assert stats["category_breakdown"]["PRIMARY"]["TASK"] == 1
        assert stats["category_breakdown"]["IMPLEMENTATION"]["FEATURE"] == 1
        assert stats["category_breakdown"]["QUALITY"]["PERF"] == 1

        # 완료율 통계 (평균/최대값 기준)
        assert stats["completion_rates"]["PRIMARY"] == 50.0  # (2+1+1+0)/4 / 2 = 50%
        assert stats["completion_rates"]["IMPLEMENTATION"] == 25.0  # 1/4 = 25%