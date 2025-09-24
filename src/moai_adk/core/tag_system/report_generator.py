"""
@FEATURE:REPORT-GENERATOR-001 - TAG 추적성 리포트 생성 최소 구현

Jinja2 템플릿 기반 Markdown/HTML 리포트 생성
"""

from enum import Enum
from pathlib import Path
from typing import List, Dict, Any, Optional
from datetime import datetime
from dataclasses import dataclass

from jinja2 import Environment, FileSystemLoader, DictLoader

from .parser import TagMatch


class ReportFormat(Enum):
    """리포트 출력 형식"""
    MARKDOWN = "markdown"
    HTML = "html"
    JSON = "json"


@dataclass
class TraceabilityReport:
    """추적성 리포트 데이터"""
    title: str
    generated_at: datetime
    tags: List[TagMatch]
    matrix: Dict[str, Any]
    coverage: Dict[str, float]


class TagReportGenerator:
    """
    TAG 추적성 리포트 생성기 최소 구현

    TRUST 원칙 적용:
    - Test First: 테스트 요구사항만 구현
    - Readable: 명확한 리포트 생성 로직
    - Unified: 리포트 생성 책임만 담당
    """

    # 상수: 카테고리 그룹 매핑
    CATEGORY_GROUPS = {
        "PRIMARY": ["REQ", "DESIGN", "TASK", "TEST"],
        "STEERING": ["VISION", "STRUCT", "TECH", "ADR"],
        "IMPLEMENTATION": ["FEATURE", "API", "UI", "DATA"],
        "QUALITY": ["PERF", "SEC", "DOCS", "TAG"]
    }

    def __init__(self, output_dir: Path, template_dir: Optional[Path] = None):
        """
        리포트 생성기 초기화

        Args:
            output_dir: 출력 디렉토리
            template_dir: 템플릿 디렉토리 (None이면 내장 템플릿 사용)
        """
        self.output_dir = Path(output_dir)
        self.template_dir = template_dir

        # Jinja2 환경 설정
        if template_dir and template_dir.exists():
            self.jinja_env = Environment(loader=FileSystemLoader(template_dir))
        else:
            self._setup_default_templates()

    def generate_chain_matrix(self, tags: List[TagMatch]) -> Dict[str, Any]:
        """
        TAG 체인 매트릭스 생성

        Args:
            tags: TAG 목록

        Returns:
            체인 매트릭스 딕셔너리
        """
        matrix = {
            "PRIMARY": {"REQ": {}, "DESIGN": {}, "TASK": {}, "TEST": {}},
            "STEERING": {"VISION": {}, "STRUCT": {}, "TECH": {}, "ADR": {}},
            "IMPLEMENTATION": {"FEATURE": {}, "API": {}, "UI": {}, "DATA": {}},
            "QUALITY": {"PERF": {}, "SEC": {}, "DOCS": {}, "TAG": {}}
        }

        for tag in tags:
            group = self._get_category_group(tag.category)
            if group in matrix and tag.category in matrix[group]:
                matrix[group][tag.category][tag.identifier] = {
                    "description": tag.description or "",
                    "references": tag.references or []
                }

        return matrix

    def analyze_missing_connections(self, tags: List[TagMatch]) -> Dict[str, Any]:
        """
        누락된 연결 분석

        Args:
            tags: 분석할 TAG 목록

        Returns:
            누락 분석 결과
        """
        # Primary Chain 검사
        primary_categories = ["REQ", "DESIGN", "TASK", "TEST"]
        found_categories = set(tag.category for tag in tags if tag.category in primary_categories)
        missing_categories = [cat for cat in primary_categories if cat not in found_categories]

        missing_links = []
        if missing_categories:
            # 누락된 카테고리와 인접 카테고리 식별
            for missing_cat in missing_categories:
                idx = primary_categories.index(missing_cat)
                before = primary_categories[idx - 1] if idx > 0 else None
                after = primary_categories[idx + 1] if idx < len(primary_categories) - 1 else None

                # 전후 TAG 식별
                before_tags = [tag.identifier for tag in tags if tag.category == before] if before else []
                after_tags = [tag.identifier for tag in tags if tag.category == after] if after else []

                missing_links.append({
                    "category": missing_cat,
                    "expected_between": before_tags + after_tags
                })

        return {
            "missing_links": missing_links,
            "completeness": 1.0 - (len(missing_categories) / len(primary_categories))
        }

    def generate_report(self,
                       tags: List[TagMatch],
                       format: ReportFormat,
                       title: str = "TAG 추적성 리포트",
                       template_name: Optional[str] = None) -> str:
        """
        리포트 생성

        Args:
            tags: TAG 목록
            format: 출력 형식
            title: 리포트 제목
            template_name: 사용할 템플릿 이름

        Returns:
            생성된 리포트 내용
        """
        # 데이터 준비
        matrix = self.generate_chain_matrix(tags)
        coverage = self.calculate_implementation_coverage(tags)

        context = {
            "title": title,
            "timestamp": datetime.now().isoformat(),
            "tags": tags,
            "matrix": matrix,
            "coverage": coverage,
            "tag_count": len(tags)
        }

        # 템플릿 선택
        if template_name:
            template = self.jinja_env.get_template(template_name)
        else:
            template_name = f"default_report.{format.value}.j2"
            template = self.jinja_env.get_template(template_name)

        return template.render(**context)

    def calculate_implementation_coverage(self, tags: List[TagMatch]) -> Dict[str, Any]:
        """
        구현 완료율 계산

        Args:
            tags: TAG 목록

        Returns:
            커버리지 정보
        """
        # Primary Chain 분석
        primary_categories = ["REQ", "DESIGN", "TASK", "TEST"]
        category_counts = {cat: 0 for cat in primary_categories}

        for tag in tags:
            if tag.category in category_counts:
                category_counts[tag.category] += 1

        # 체인 완성도 계산 (가장 적은 카테고리 기준)
        min_count = min(category_counts.values()) if category_counts.values() else 0
        max_count = max(category_counts.values()) if category_counts.values() else 0

        # 완전 체인 수 (모든 카테고리가 균등한 경우)
        complete_chains = min_count
        total_potential_chains = max_count if max_count > 0 else 1

        coverage_percentage = (complete_chains / total_potential_chains) * 100

        # 카테고리별 커버리지
        category_coverage = {}
        for cat, count in category_counts.items():
            category_coverage[cat] = (count / max_count * 100) if max_count > 0 else 0

        return {
            "total_chains": max_count,
            "complete_chains": complete_chains,
            "coverage_percentage": coverage_percentage,
            "category_coverage": category_coverage
        }

    def export_to_file(self,
                      tags: List[TagMatch],
                      output_path: Path,
                      format: ReportFormat) -> None:
        """
        리포트를 파일로 내보내기

        Args:
            tags: TAG 목록
            output_path: 출력 파일 경로
            format: 출력 형식
        """
        content = self.generate_report(tags, format)

        output_path.parent.mkdir(parents=True, exist_ok=True)
        output_path.write_text(content, encoding='utf-8')

    def generate_summary_statistics(self, tags: List[TagMatch]) -> Dict[str, Any]:
        """
        요약 통계 생성

        Args:
            tags: TAG 목록

        Returns:
            요약 통계
        """
        total_tags = len(tags)

        # 카테고리별 분류
        category_breakdown = {
            "PRIMARY": {"REQ": 0, "DESIGN": 0, "TASK": 0, "TEST": 0},
            "STEERING": {"VISION": 0, "STRUCT": 0, "TECH": 0, "ADR": 0},
            "IMPLEMENTATION": {"FEATURE": 0, "API": 0, "UI": 0, "DATA": 0},
            "QUALITY": {"PERF": 0, "SEC": 0, "DOCS": 0, "TAG": 0}
        }

        for tag in tags:
            group = self._get_category_group(tag.category)
            if group in category_breakdown and tag.category in category_breakdown[group]:
                category_breakdown[group][tag.category] += 1

        # 완료율 계산 (가장 많은 카테고리 기준으로 다른 카테고리의 비율)
        completion_rates = {}
        for group_name, categories in category_breakdown.items():
            counts = list(categories.values())
            if not counts or all(c == 0 for c in counts):
                completion_rates[group_name] = 0.0
            else:
                max_count = max(counts)
                avg_coverage = sum(counts) / len(counts)
                completion_rates[group_name] = (avg_coverage / max_count * 100) if max_count > 0 else 0

        return {
            "total_tags": total_tags,
            "category_breakdown": category_breakdown,
            "completion_rates": completion_rates
        }

    def _setup_default_templates(self) -> None:
        """기본 템플릿 설정"""
        default_templates = {
            "default_report.markdown.j2": """# {{ title }}

생성 시간: {{ timestamp }}
총 TAG 수: {{ tag_count }}

{% if tag_count == 0 %}
TAG가 발견되지 않았습니다.
{% else %}

## Primary Chain 분석

### REQ → DESIGN → TASK → TEST

{% for tag in tags %}
{% if tag.category in ['REQ', 'DESIGN', 'TASK', 'TEST'] %}
- {{ tag.category }}:{{ tag.identifier }}{% if tag.description %} - {{ tag.description }}{% endif %}
{% endif %}
{% endfor %}

## 구현 커버리지

{% for group_name, group_data in matrix.items() %}
### {{ group_name }}
{% for category, category_data in group_data.items() %}
{% if category_data %}
- **{{ category }}**: {{ category_data.keys() | list | length }}개
{% for tag_id, tag_info in category_data.items() %}
  - {{ tag_id }}{% if tag_info.description %}: {{ tag_info.description }}{% endif %}
{% endfor %}
{% endif %}
{% endfor %}
{% endfor %}
{% endif %}
""",

            "default_report.html.j2": """<!DOCTYPE html>
<html>
<head>
    <title>{{ title }}</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1, h2, h3 { color: #333; }
        .tag-item { margin: 5px 0; }
        .category { font-weight: bold; }
    </style>
</head>
<body>
    <h1>{{ title }}</h1>
    <p><strong>생성 시간:</strong> {{ timestamp }}</p>
    <p><strong>총 TAG 수:</strong> {{ tag_count }}</p>

    <h2>TAG 목록</h2>
    {% for tag in tags %}
    <div class="tag-item">
        <span class="category">{{ tag.category }}:{{ tag.identifier }}</span>
        {% if tag.description %}
        - {{ tag.description }}
        {% endif %}
    </div>
    {% endfor %}

    {% if not tags %}
    <p>TAG가 발견되지 않았습니다.</p>
    {% endif %}
</body>
</html>
""",
        }

        self.jinja_env = Environment(loader=DictLoader(default_templates))

    def _get_category_group(self, category: str) -> str:
        """카테고리 그룹 결정"""
        for group_name, categories in self.CATEGORY_GROUPS.items():
            if category in categories:
                return group_name
        return "UNKNOWN"