"""
Integration features: file watching, migration, and configuration.

@FEATURE:ADAPTER-INTEGRATION-001 File watching, migration, and system integration
@DESIGN:SEPARATED-INTEGRATION-001 Extracted from oversized adapter.py (631 LOC)
"""

import threading
from pathlib import Path
from typing import Any

from .database import TagDatabaseManager
from .index_manager import IndexUpdateEvent, TagIndexManager
from .parser import TagParser


class AdapterIntegration:
    """통합 기능 (파일 감시 등) - SQLite 전용"""

    def __init__(self, db_manager: TagDatabaseManager):
        """통합 모듈 초기화"""
        self.db_manager = db_manager

        # 파일 감시 관련
        self.index_manager = None
        self.tag_parser = TagParser()
        self.watching_lock = threading.Lock()
        self._watching = False

    def is_watching(self) -> bool:
        """파일 감시 상태 반환"""
        return self._watching

    def start_watching(self) -> None:
        """파일 감시 시작"""
        with self.watching_lock:
            if not self._watching and self.index_manager:
                self.index_manager.start_watching()
                self._watching = True

    def stop_watching(self) -> None:
        """파일 감시 중지"""
        with self.watching_lock:
            if self._watching and self.index_manager:
                self.index_manager.stop_watching()
                self._watching = False

    def process_file_change(self, file_path: Path, event_type: str) -> None:
        """파일 변경 처리 (호환성용 인터페이스)"""
        try:
            self._process_file_change_sqlite(file_path, event_type)
        except Exception:
            # 파일 감시 오류는 로깅만 하고 계속 진행
            pass

    def _process_file_change_sqlite(self, file_path: Path, event_type: str) -> None:
        """SQLite 백엔드를 통한 파일 변경 처리"""
        try:
            if event_type in ['modified', 'created']:
                # 파일에서 TAG 파싱
                if file_path.exists():
                    parsed_tags = self.tag_parser.parse_file(file_path)

                    # 기존 TAG 삭제 (해당 파일)
                    self.db_manager.delete_tags_by_file(str(file_path))

                    # 새 TAG 추가
                    for tag_info in parsed_tags:
                        self.db_manager.insert_tag(
                            category=tag_info['category'],
                            identifier=tag_info['identifier'],
                            description=tag_info.get('description', ''),
                            file_path=str(file_path),
                            line_number=tag_info.get('line_number', 1)
                        )

            elif event_type == 'deleted':
                # 파일 삭제 시 관련 TAG 모두 삭제
                self.db_manager.delete_tags_by_file(str(file_path))

        except Exception as e:
            raise Exception(f"파일 변경 처리 실패 ({file_path}): {e}")

    # JSON 마이그레이션 기능 제거됨 - SQLite 전용 시스템

    # JSON 내보내기 기능 제거됨 - SQLite 전용 시스템

    def setup_file_watching(self, project_root: Path, patterns: list[str] | None = None) -> None:
        """파일 감시 설정"""
        try:
            # 기본 패턴 설정
            if not patterns:
                patterns = ['*.py', '*.md', '*.txt', '*.js', '*.ts', '*.java', '*.cpp', '*.h']

            # IndexManager 초기화
            self.index_manager = TagIndexManager(
                project_root=project_root,
                index_file=project_root / ".moai" / "indexes" / "tags.db",
                file_patterns=patterns
            )

            # 이벤트 핸들러 등록
            self.index_manager.add_update_listener(self._handle_index_update)

        except Exception as e:
            raise Exception(f"파일 감시 설정 실패: {e}")

    def _handle_index_update(self, event: IndexUpdateEvent) -> None:
        """인덱스 업데이트 이벤트 핸들러"""
        try:
            if event.file_path:
                self.process_file_change(Path(event.file_path), event.event_type)
        except Exception:
            # 개별 파일 처리 실패는 무시하고 계속 진행
            pass

    def get_sync_status(self) -> dict[str, Any]:
        """동기화 상태 정보"""
        try:
            all_tags = self.db_manager.get_all_tags()

            # 파일별 TAG 분포
            file_distribution = {}
            for tag in all_tags:
                file_path = tag.get('file_path', 'unknown')
                if file_path not in file_distribution:
                    file_distribution[file_path] = 0
                file_distribution[file_path] += 1

            # 카테고리별 분포
            category_distribution = {}
            for tag in all_tags:
                category = tag.get('category', 'unknown')
                if category not in category_distribution:
                    category_distribution[category] = 0
                category_distribution[category] += 1

            return {
                'total_tags': len(all_tags),
                'file_distribution': file_distribution,
                'category_distribution': category_distribution,
                'watching': self._watching,
                'backend': 'sqlite'
            }

        except Exception as e:
            return {
                'error': str(e),
                'backend': 'sqlite'
            }

    def cleanup(self):
        """리소스 정리"""
        self.stop_watching()
        if self.index_manager:
            try:
                self.index_manager.cleanup()
            except:
                pass
