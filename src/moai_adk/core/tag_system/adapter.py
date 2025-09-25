"""
@FEATURE:SPEC-009-TAG-ADAPTER-001 - SQLite 백엔드와 JSON API 호환성 어댑터

GREEN 단계: 기존 JSON API와 100% 호환되는 SQLite 어댑터 구현
@DESIGN:REFACTORED-ADAPTER-001 Rebuilt from 631 LOC to modular architecture (< 200 LOC)
"""

from dataclasses import dataclass
from pathlib import Path
from typing import Any

from .adapter_core import AdapterCore, ApiCompatibilityError
from .adapter_integration import AdapterIntegration
from .adapter_search import AdapterSearch
from .database import TagDatabaseManager


@dataclass
class AdapterConfiguration:
    """어댑터 설정"""
    backend_type: str
    database_path: Path | None
    json_fallback_path: Path | None
    performance_monitoring: bool = True


class TagIndexAdapter:
    """
    SQLite 백엔드와 기존 JSON API 호환성 어댑터 오케스트레이터

    TRUST 원칙 적용:
    - Test First: 기존 API 호환성 테스트 통과
    - Readable: 명확한 어댑터 패턴 구현
    - Unified: API 변환 오케스트레이션 책임만 담당
    @DESIGN:REFACTORED-ORCHESTRATOR-001 Now coordinates specialized modules
    """

    def __init__(self, database_path: Path, json_fallback_path: Path | None = None,
                 performance_monitor: Any | None = None):
        """어댑터 초기화"""
        self.database_path = Path(database_path)
        self.json_fallback_path = Path(json_fallback_path) if json_fallback_path else None
        self.performance_monitor = performance_monitor

        # SQLite 백엔드 초기화
        try:
            self.db_manager = TagDatabaseManager(self.database_path)
            self.db_manager.initialize()
            self.backend_available = True
        except Exception:
            self.backend_available = False

        # 전용 모듈들 초기화
        self.core = AdapterCore(database_path, json_fallback_path)
        if self.backend_available:
            self.search = AdapterSearch(self.db_manager)
            self.integration = AdapterIntegration(self.db_manager, json_fallback_path)

    # Core 기능 델리게이트
    def is_watching(self) -> bool:
        """파일 감시 상태 반환"""
        if self.backend_available:
            return self.integration.is_watching()
        return False

    def initialize(self) -> None:
        """어댑터 초기화"""
        return self.core.initialize()

    def initialize_index(self) -> None:
        """인덱스 초기화 (호환성용)"""
        return self.core.initialize_index()

    def load_index(self) -> dict[str, Any]:
        """인덱스 로드 (기존 JSON API 호환)"""
        return self.core.load_index()

    def save_index(self, index_data: dict[str, Any]) -> None:
        """인덱스 저장 (기존 JSON API 호환)"""
        return self.core.save_index(index_data)

    def validate_index_schema(self, index_data: dict[str, Any]) -> bool:
        """인덱스 스키마 검증"""
        return self.core.validate_index_schema(index_data)

    def get_configuration_info(self) -> dict[str, Any]:
        """설정 정보 반환"""
        return self.core.get_configuration_info()

    # Search 기능 델리게이트
    def search_by_category(self, category: str, **filters) -> list[dict[str, Any]]:
        """카테고리별 검색 (고급 필터링 포함)"""
        if not self.backend_available:
            raise ApiCompatibilityError("SQLite 백엔드를 사용할 수 없습니다")
        return self.search.search_by_category(category, **filters)

    def get_traceability_chain(self, tag_identifier: str,
                              direction: str = 'both',
                              max_depth: int = 5,
                              include_details: bool = True,
                              category_filter: list[str] | None = None) -> dict[str, Any]:
        """추적성 체인 분석 (고도화된 버전)"""
        if not self.backend_available:
            return {
                'start_tag': tag_identifier,
                'found': False,
                'error': 'SQLite 백엔드를 사용할 수 없습니다'
            }
        return self.search.get_traceability_chain(tag_identifier, direction, max_depth, include_details, category_filter)

    # Integration 기능 델리게이트
    def start_watching(self) -> None:
        """파일 감시 시작"""
        if self.backend_available:
            self.integration.start_watching()

    def stop_watching(self) -> None:
        """파일 감시 중지"""
        if self.backend_available:
            self.integration.stop_watching()

    def process_file_change(self, file_path: Path, event_type: str) -> None:
        """파일 변경 처리 (호환성용 인터페이스)"""
        if self.backend_available:
            self.integration.process_file_change(file_path, event_type)

    def migrate_from_json(self, json_path: Path) -> None:
        """JSON에서 마이그레이션"""
        if not self.backend_available:
            raise ApiCompatibilityError("SQLite 백엔드를 사용할 수 없습니다")
        self.integration.migrate_from_json(json_path)

    def export_to_json(self, json_path: Path) -> None:
        """JSON으로 내보내기"""
        if not self.backend_available:
            raise ApiCompatibilityError("SQLite 백엔드를 사용할 수 없습니다")
        self.integration.export_to_json(json_path)

    def close(self):
        """리소스 정리"""
        if self.backend_available:
            self.integration.cleanup()
        self.core.close()
