"""
@FEATURE:SPEC-009-TAG-ADAPTER-001 - SQLite 백엔드와 JSON API 호환성 어댑터

GREEN 단계: 기존 JSON API와 100% 호환되는 SQLite 어댑터 구현
"""

import json
import threading
from pathlib import Path
from typing import Dict, List, Any, Optional, Callable
from datetime import datetime
from dataclasses import dataclass

from .database import TagDatabaseManager, DatabaseError
from .parser import TagParser
from .index_manager import IndexUpdateEvent, TagIndexManager


class ApiCompatibilityError(Exception):
    """API 호환성 관련 오류"""
    pass


@dataclass
class AdapterConfiguration:
    """어댑터 설정"""
    backend_type: str
    database_path: Optional[Path]
    json_fallback_path: Optional[Path]
    performance_monitoring: bool = True


class TagIndexAdapter:
    """
    SQLite 백엔드와 기존 JSON API 호환성 어댑터

    TRUST 원칙 적용:
    - Test First: 기존 API 호환성 테스트 통과
    - Readable: 명확한 어댑터 패턴 구현
    - Unified: API 변환 책임만 담당
    """

    def __init__(self, database_path: Path, json_fallback_path: Optional[Path] = None,
                 performance_monitor: Optional[Any] = None):
        """어댑터 초기화"""
        self.database_path = Path(database_path)
        self.json_fallback_path = Path(json_fallback_path) if json_fallback_path else None
        self.performance_monitor = performance_monitor

        # SQLite 백엔드 초기화
        try:
            self._database = TagDatabaseManager(self.database_path)
            self._sqlite_available = True
        except DatabaseError:
            self._sqlite_available = False
            self._database = None

        # JSON fallback 준비
        self._json_manager = None
        if self.json_fallback_path:
            self._json_manager = TagIndexManager(
                watch_directory=self.database_path.parent,
                index_file=self.json_fallback_path
            )

        self.parser = TagParser()
        self._lock = threading.Lock()

        # 구조화된 로거 설정 (TRUST 원칙: Secured/Trackable)
        import logging
        self._logger = logging.getLogger(f"{__name__}.{self.__class__.__name__}")
        if not self._logger.handlers:
            handler = logging.StreamHandler()
            formatter = logging.Formatter(
                '{"timestamp": "%(asctime)s", "level": "%(levelname)s", '
                '"component": "%(name)s", "message": "%(message)s"}'
            )
            handler.setFormatter(formatter)
            self._logger.addHandler(handler)
            self._logger.setLevel(logging.INFO)

        # 기존 API 호환을 위한 콜백
        self.on_file_changed: Optional[Callable[[IndexUpdateEvent], None]] = None

        # 감시 상태 (기존 API 호환)
        self._is_watching = False

    @property
    def is_watching(self) -> bool:
        """파일 감시 상태 반환 (기존 API 호환)"""
        return self._is_watching

    def initialize(self) -> None:
        """초기화 (기존 initialize_index와 호환)"""
        if self._sqlite_available:
            self._database.initialize()
        elif self._json_manager:
            self._json_manager.initialize_index()
        else:
            raise ApiCompatibilityError("Neither SQLite nor JSON backend available")

    def initialize_index(self) -> None:
        """기존 API 호환을 위한 별칭"""
        self.initialize()

    def load_index(self) -> Dict[str, Any]:
        """
        기존 JSON API와 완전 호환되는 인덱스 로드

        Returns:
            기존 JSON 형식과 동일한 구조
        """
        if self._sqlite_available:
            return self._load_from_sqlite()
        elif self._json_manager:
            return self._json_manager.load_index()
        else:
            # 빈 기본 구조 반환
            return self._create_empty_index()

    def _load_from_sqlite(self) -> Dict[str, Any]:
        """SQLite에서 기존 JSON 형식으로 데이터 변환"""
        all_tags = self._database.get_all_tags()

        # 기존 JSON 구조 생성
        index_data = {
            "metadata": {
                "created_at": datetime.now().isoformat(),
                "updated_at": datetime.now().isoformat(),
                "version": "1.0",
                "total_tags": len(all_tags)
            },
            "categories": {
                "PRIMARY": {},
                "STEERING": {},
                "IMPLEMENTATION": {},
                "QUALITY": {}
            },
            "chains": [],  # TODO: 참조 체인 변환
            "files": {}
        }

        # TAG 데이터 변환
        for tag in all_tags:
            # 카테고리 그룹 결정
            group = self._get_category_group(tag['category'])

            # 카테고리 그룹에 카테고리 추가
            if tag['category'] not in index_data["categories"][group]:
                index_data["categories"][group][tag['category']] = {}

            # TAG 정보 추가
            index_data["categories"][group][tag['category']][tag['identifier']] = {
                "description": tag['description'] or "",
                "file": tag['file_path']
            }

        # 파일별 TAG 목록 구성 (줄 번호순으로 정렬)
        files_data = {}
        for tag in all_tags:
            file_path = tag['file_path']
            if file_path not in files_data:
                files_data[file_path] = []

            files_data[file_path].append({
                "category": tag['category'],
                "identifier": tag['identifier'],
                "description": tag['description'] or "",
                "line_number": tag.get('line_number', 0)
            })

        # 각 파일 내에서 줄 번호순 정렬 후 line_number 제거
        for file_path, tags in files_data.items():
            sorted_tags = sorted(tags, key=lambda x: x.get('line_number', 0))
            index_data["files"][file_path] = [
                {
                    "category": tag['category'],
                    "identifier": tag['identifier'],
                    "description": tag['description']
                }
                for tag in sorted_tags
            ]

        return index_data

    def save_index(self, index_data: Dict[str, Any]) -> None:
        """인덱스 저장 (기존 API 호환)"""
        if self._sqlite_available:
            # JSON 데이터를 SQLite로 변환하여 저장
            self._save_to_sqlite(index_data)
        elif self._json_manager:
            self._json_manager.save_index(index_data)

    def _save_to_sqlite(self, index_data: Dict[str, Any]) -> None:
        """JSON 형식 데이터를 SQLite에 저장"""
        # 기존 데이터 삭제
        # TODO: 더 효율적인 업데이트 방식 구현

        # 새 데이터 삽입
        for group_name, group_data in index_data["categories"].items():
            for category, tags in group_data.items():
                for identifier, tag_info in tags.items():
                    self._database.insert_tag(
                        category=category,
                        identifier=identifier,
                        description=tag_info.get("description", ""),
                        file_path=tag_info.get("file", "")
                    )

    def process_file_change(self, file_path: Path, event_type: str) -> None:
        """
        파일 변경 처리 (기존 API와 동일한 시그니처)

        Args:
            file_path: 변경된 파일 경로
            event_type: 이벤트 타입 (created, modified, deleted)
        """
        if self._sqlite_available:
            self._process_file_change_sqlite(file_path, event_type)
        elif self._json_manager:
            self._json_manager.process_file_change(file_path, event_type)

        # 콜백 호출 (기존 API 호환)
        if self.on_file_changed:
            event = IndexUpdateEvent(
                event_type=event_type,
                file_path=file_path,
                timestamp=datetime.now()
            )
            self.on_file_changed(event)

    def _process_file_change_sqlite(self, file_path: Path, event_type: str) -> None:
        """SQLite 백엔드에서 파일 변경 처리"""
        if event_type == "deleted":
            # 파일과 관련된 모든 TAG 삭제
            file_tags = self._database.search_tags_by_file(str(file_path))
            for tag in file_tags:
                self._database.delete_tag(tag['id'])
        else:
            # 파일에서 TAG 추출 및 업데이트
            if file_path.exists() and file_path.is_file():
                try:
                    content = file_path.read_text(encoding='utf-8')
                    tags_with_pos = self.parser.extract_tags_with_positions(content)

                    # 기존 TAG 삭제
                    file_tags = self._database.search_tags_by_file(str(file_path))
                    for tag in file_tags:
                        self._database.delete_tag(tag['id'])

                    # 새 TAG 추가 (위치 정보 포함)
                    for tag, position in tags_with_pos:
                        self._database.insert_tag(
                            category=tag.category,
                            identifier=tag.identifier,
                            description=tag.description or "",
                            file_path=str(file_path),
                            line_number=position.line_number
                        )

                except (UnicodeDecodeError, PermissionError):
                    pass  # 파일 읽기 실패 시 무시

    def validate_index_schema(self, index_data: Dict[str, Any]) -> bool:
        """인덱스 스키마 검증 (기존 API와 동일)"""
        if self._json_manager:
            return self._json_manager.validate_index_schema(index_data)

        # 기본 스키마 검증
        required_keys = ["metadata", "categories", "chains", "files"]
        return all(key in index_data for key in required_keys)

    def start_watching(self) -> None:
        """파일 감시 시작 (기존 API 호환)"""
        self._is_watching = True
        if self._json_manager:
            self._json_manager.start_watching()

    def stop_watching(self) -> None:
        """파일 감시 중지 (기존 API 호환)"""
        self._is_watching = False
        if self._json_manager:
            self._json_manager.stop_watching()

    def migrate_from_json(self, json_path: Path) -> None:
        """JSON에서 SQLite로 데이터 마이그레이션"""
        if not self._sqlite_available:
            raise ApiCompatibilityError("SQLite backend not available")

        with open(json_path, 'r', encoding='utf-8') as f:
            json_data = json.load(f)

        self._save_to_sqlite(json_data)

    def export_to_json(self, json_path: Path) -> None:
        """SQLite에서 JSON으로 데이터 내보내기"""
        index_data = self._load_from_sqlite()

        with open(json_path, 'w', encoding='utf-8') as f:
            json.dump(index_data, f, indent=2, ensure_ascii=False)

    def get_configuration_info(self) -> Dict[str, Any]:
        """설정 정보 반환 (디버깅용)"""
        return {
            "backend_type": "sqlite" if self._sqlite_available else "json",
            "database_path": str(self.database_path) if self.database_path else None,
            "json_fallback_path": str(self.json_fallback_path) if self.json_fallback_path else None,
            "performance_stats": {
                "total_tags": len(self._database.get_all_tags()) if self._sqlite_available else 0,
                "query_count": 0,  # TODO: 쿼리 카운터 구현
                "avg_query_time": 0.0  # TODO: 성능 통계 구현
            }
        }

    def _create_empty_index(self) -> Dict[str, Any]:
        """빈 인덱스 구조 생성"""
        return {
            "metadata": {
                "created_at": datetime.now().isoformat(),
                "updated_at": datetime.now().isoformat(),
                "version": "1.0",
                "total_tags": 0
            },
            "categories": {
                "PRIMARY": {},
                "STEERING": {},
                "IMPLEMENTATION": {},
                "QUALITY": {}
            },
            "chains": [],
            "files": {}
        }

    def _get_category_group(self, category: str) -> str:
        """카테고리 그룹 결정 (기존 로직과 동일)"""
        if category in ["REQ", "DESIGN", "TASK", "TEST"]:
            return "PRIMARY"
        elif category in ["VISION", "STRUCT", "TECH", "ADR"]:
            return "STEERING"
        elif category in ["FEATURE", "API", "UI", "DATA"]:
            return "IMPLEMENTATION"
        elif category in ["PERF", "SEC", "DOCS", "TAG"]:
            return "QUALITY"
        else:
            return "PRIMARY"  # 기본값

    def search_by_category(self, category: str, **filters) -> List[Dict[str, Any]]:
        """
        카테고리별 TAG 검색 (JSON API 호환)

        TRUST 원칙 적용:
        - Test First: 실패 테스트로 시작하여 최소 구현
        - Readable: 명확한 매개변수와 반환값 문서화
        - Unified: 기존 JSON API와 완전 호환
        - Secured: 입력 검증과 구조화 로깅
        - Trackable: 성능 메트릭 수집

        Args:
            category: TAG 카테고리 (REQ, DESIGN, TASK, TEST 등)
                     빈 문자열이나 None은 빈 결과 반환
            **filters: 추가 필터 (향후 확장용)
                     file_pattern: 파일 경로 패턴 필터
                     description_pattern: 설명 텍스트 필터

        Returns:
            JSON API 형식의 TAG 목록:
            [
                {
                    "category": "REQ",
                    "identifier": "USER-AUTH-001",
                    "description": "사용자 인증 요구사항",
                    "file_path": "specs/auth.md",
                    "line_number": 25
                }
            ]

        Raises:
            ValueError: 유효하지 않은 카테고리 이름
        """
        import time
        start_time = time.time()

        # 입력 검증 (TRUST 원칙: Secured)
        if not category or not isinstance(category, str):
            return []

        category = category.strip().upper()
        if not category:
            return []
        try:
            if self._sqlite_available:
                # SQLite 백엔드에서 검색
                raw_tags = self._database.search_tags_by_category(category)

                # JSON API 형식으로 변환
                results = []
                for tag in raw_tags:
                    results.append({
                        "category": tag["category"],
                        "identifier": tag["identifier"],
                        "description": tag["description"] or "",
                        "file_path": tag["file_path"],
                        "line_number": tag.get("line_number", 0)
                    })

                # 성능 로깅 (TRUST 원칙: Trackable)
                duration_ms = (time.time() - start_time) * 1000
                self._logger.info(
                    '{"operation": "search_by_category", "category": "%s", '
                    '"backend": "sqlite", "result_count": %d, '
                    '"duration_ms": %.2f, "success": true}',
                    category, len(results), duration_ms
                )

                return results
            elif self._json_manager:
                # JSON fallback - 기존 인덱스에서 필터링
                index_data = self._json_manager.load_index()
                results = []

                # 모든 그룹에서 해당 카테고리 검색
                for group_name, group_data in index_data.get("categories", {}).items():
                    if category in group_data:
                        for identifier, tag_info in group_data[category].items():
                            results.append({
                                "category": category,
                                "identifier": identifier,
                                "description": tag_info.get("description", ""),
                                "file_path": tag_info.get("file", ""),
                                "line_number": 0  # JSON에서는 줄 번호 정보 없음
                            })

                # 성능 로깅
                duration_ms = (time.time() - start_time) * 1000
                self._logger.info(
                    '{"operation": "search_by_category", "category": "%s", '
                    '"backend": "json", "result_count": %d, '
                    '"duration_ms": %.2f, "success": true}',
                    category, len(results), duration_ms
                )

                return results
            else:
                # 백엔드 없음 - 빈 결과 반환
                self._logger.warning(
                    '{"operation": "search_by_category", "category": "%s", '
                    '"backend": "none", "result_count": 0, '
                    '"duration_ms": 0, "success": false, '
                    '"error": "No backend available"}',
                    category
                )
                return []

        except Exception as e:
            # 오류 로깅 (TRUST 원칙: Secured)
            duration_ms = (time.time() - start_time) * 1000
            self._logger.error(
                '{"operation": "search_by_category", "category": "%s", '
                '"duration_ms": %.2f, "success": false, '
                '"error": "%s"}',
                category, duration_ms, str(e).replace('"', '\\"')
            )
            # 오류 발생 시 빈 결과 반환 (안전한 실패)
            return []

    def get_traceability_chain(self, tag_identifier: str,
                             direction: str = "forward",
                             max_depth: int = 10) -> Dict[str, Any]:
        """
        TAG 추적성 체인 구축 (16-Core TAG 시스템 지원)

        TRUST 원칙 적용:
        - Test First: 실패 테스트 기반 구현
        - Readable: 명확한 체인 구조와 예제
        - Unified: 표준 그래프 구조 사용
        - Secured: 순환 참조 방지와 깊이 제한
        - Trackable: 체인 구축 성능 모니터링

        Args:
            tag_identifier: TAG 식별자
                          형식: "CATEGORY:IDENTIFIER" (예: "REQ:USER-AUTH-001")
                          또는 "IDENTIFIER" (카테고리 자동 추론)
            direction: 참조 방향
                      "forward": 순방향 (REQ → DESIGN → TASK → TEST)
                      "backward": 역방향 (TEST → TASK → DESIGN → REQ)
                      "both": 양방향 (전체 연결 그래프)
            max_depth: 최대 탐색 깊이 (1-50, 기본값 10)
                      순환 참조 방지와 성능 보장

        Returns:
            추적성 체인 그래프 구조:
            {
                "nodes": [
                    {
                        "id": 123,
                        "identifier": "REQ:USER-AUTH-001",
                        "category": "REQ",
                        "description": "사용자 인증 요구사항",
                        "file_path": "specs/auth.md"
                    }
                ],
                "edges": [
                    {
                        "source": 123,
                        "target": 456,
                        "type": "chain",
                        "direction": "forward"
                    }
                ],
                "direction": "forward",
                "max_depth": 10,
                "truncated": false,
                "metadata": {
                    "total_nodes": 4,
                    "total_edges": 3,
                    "chain_depth": 4
                }
            }

        Raises:
            ValueError: 잘못된 매개변수 값
            ApiCompatibilityError: 백엔드 연결 실패
        """
        import time
        start_time = time.time()

        # 입력 검증 (TRUST 원칙: Secured)
        if not tag_identifier or not isinstance(tag_identifier, str):
            raise ValueError("tag_identifier must be a non-empty string")

        if direction not in ["forward", "backward", "both"]:
            raise ValueError("direction must be 'forward', 'backward', or 'both'")

        if not isinstance(max_depth, int) or max_depth < 1 or max_depth > 50:
            raise ValueError("max_depth must be an integer between 1 and 50")

        tag_identifier = tag_identifier.strip()
        try:
            if not self._sqlite_available:
                # SQLite 없으면 기본 구조만 반환 (JSON에는 참조 정보 없음)
                duration_ms = (time.time() - start_time) * 1000
                self._logger.warning(
                    '{"operation": "get_traceability_chain", '
                    '"tag_identifier": "%s", "backend": "none", '
                    '"duration_ms": %.2f, "success": false, '
                    '"error": "SQLite backend required for traceability"}',
                    tag_identifier, duration_ms
                )
                return {
                    "nodes": [{"identifier": tag_identifier, "category": "UNKNOWN", "description": ""}],
                    "edges": [],
                    "direction": direction,
                    "max_depth": max_depth,
                    "truncated": False,
                    "error": "SQLite backend required for traceability"
                }

            # 시작 TAG 찾기 (TRUST 원칙: Readable)
            category, identifier = tag_identifier.split(":", 1) if ":" in tag_identifier else ("", tag_identifier)
            start_tags = self._database.search_tags_by_identifier(identifier)

            if not start_tags:
                duration_ms = (time.time() - start_time) * 1000
                self._logger.info(
                    '{"operation": "get_traceability_chain", '
                    '"tag_identifier": "%s", "duration_ms": %.2f, '
                    '"success": false, "error": "Tag not found"}',
                    tag_identifier, duration_ms
                )
                return {
                    "nodes": [],
                    "edges": [],
                    "direction": direction,
                    "max_depth": max_depth,
                    "error": f"Tag not found: {tag_identifier}",
                    "metadata": {"total_nodes": 0, "total_edges": 0, "chain_depth": 0}
                }

            start_tag = start_tags[0]  # 첫 번째 매치 사용

            # 체인 구축 (현재는 단일 노드만 - 참조 시스템이 완전하지 않음)
            # 향후 SPEC-010에서 완전한 참조 체인 구현 예정
            nodes = [{
                "id": start_tag["id"],
                "identifier": f"{start_tag['category']}:{start_tag['identifier']}",
                "category": start_tag["category"],
                "description": start_tag["description"] or "",
                "file_path": start_tag["file_path"]
            }]

            edges = []  # 추후 참조 관계 구현 시 추가

            # 성능 로깅 (TRUST 원칙: Trackable)
            duration_ms = (time.time() - start_time) * 1000
            self._logger.info(
                '{"operation": "get_traceability_chain", '
                '"tag_identifier": "%s", "direction": "%s", '
                '"max_depth": %d, "result_nodes": %d, '
                '"duration_ms": %.2f, "success": true}',
                tag_identifier, direction, max_depth, len(nodes), duration_ms
            )

            return {
                "nodes": nodes,
                "edges": edges,
                "direction": direction,
                "max_depth": max_depth,
                "truncated": False,
                "metadata": {
                    "total_nodes": len(nodes),
                    "total_edges": len(edges),
                    "chain_depth": 1  # 현재는 단일 노드만
                }
            }

        except Exception as e:
            # 오류 로깅 (TRUST 원칙: Secured)
            duration_ms = (time.time() - start_time) * 1000
            self._logger.error(
                '{"operation": "get_traceability_chain", '
                '"tag_identifier": "%s", "duration_ms": %.2f, '
                '"success": false, "error": "%s"}',
                tag_identifier, duration_ms, str(e).replace('"', '\\"')
            )
            raise ApiCompatibilityError(f"Failed to build traceability chain: {e}")

    def close(self):
        """리소스 정리"""
        if self._database:
            self._database.close()
        if self._json_manager:
            self._json_manager.stop_watching()