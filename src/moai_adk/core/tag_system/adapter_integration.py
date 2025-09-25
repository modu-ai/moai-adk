"""
Integration features: file watching, migration, and configuration.

@FEATURE:ADAPTER-INTEGRATION-001 File watching, migration, and system integration
@DESIGN:SEPARATED-INTEGRATION-001 Extracted from oversized adapter.py (631 LOC)
"""

import json
import threading
from pathlib import Path
from typing import Any

from .database import TagDatabaseManager
from .index_manager import IndexUpdateEvent, TagIndexManager
from .parser import TagParser


class AdapterIntegration:
    """통합 기능 (파일 감시, 마이그레이션 등)"""

    def __init__(self, db_manager: TagDatabaseManager, json_fallback_path: Path | None = None):
        """통합 모듈 초기화"""
        self.db_manager = db_manager
        self.json_fallback_path = json_fallback_path

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

    def migrate_from_json(self, json_path: Path) -> None:
        """JSON에서 SQLite로 마이그레이션"""
        try:
            with open(json_path, encoding='utf-8') as f:
                json_data = json.load(f)

            # 기존 데이터 클리어
            self.db_manager.clear_all_data()

            # TAG 데이터 마이그레이션
            index_data = json_data.get('index', {})
            tag_id_mapping = {}

            for tag_key, entries in index_data.items():
                if ':' not in tag_key or not entries:
                    continue

                category, identifier = tag_key.split(':', 1)
                first_entry = entries[0]

                # TAG 삽입
                tag_id = self.db_manager.insert_tag(
                    category=category,
                    identifier=identifier,
                    description=first_entry.get('context', '').replace(f'@{tag_key}', '').strip(),
                    file_path=first_entry.get('file', ''),
                    line_number=first_entry.get('line', 1)
                )
                tag_id_mapping[tag_key] = tag_id

            # 참조 관계 마이그레이션
            references_data = json_data.get('references', {})
            for source_key, target_keys in references_data.items():
                if source_key in tag_id_mapping:
                    source_id = tag_id_mapping[source_key]
                    for target_key in target_keys:
                        if target_key in tag_id_mapping:
                            target_id = tag_id_mapping[target_key]
                            self.db_manager.create_reference(source_id, target_id, 'chain')

        except Exception as e:
            raise Exception(f"JSON 마이그레이션 실패: {e}")

    def export_to_json(self, json_path: Path) -> None:
        """SQLite에서 JSON으로 내보내기"""
        try:
            # SQLite에서 모든 데이터 읽기
            all_tags = self.db_manager.get_all_tags()
            all_references = self.db_manager.get_all_references()

            # JSON 형식으로 변환
            json_data = {
                "version": "2.0.0",
                "backend": "sqlite",
                "updated": "auto-generated",
                "statistics": {
                    "total_tags": len(all_tags),
                    "categories": {}
                },
                "index": {},
                "references": {}
            }

            # 카테고리 통계
            category_counts = {}
            for tag in all_tags:
                category = tag['category']
                category_counts[category] = category_counts.get(category, 0) + 1
            json_data["statistics"]["categories"] = category_counts

            # 인덱스 데이터
            for tag in all_tags:
                tag_key = f"{tag['category']}:{tag['identifier']}"
                json_data["index"][tag_key] = [{
                    "file": tag['file_path'],
                    "line": tag.get('line_number', 1),
                    "context": f"@{tag_key} {tag.get('description', '')}"
                }]

            # 참조 데이터
            reference_map = {}
            for ref in all_references:
                source_tag = self.db_manager.get_tag_by_id(ref['source_tag_id'])
                target_tag = self.db_manager.get_tag_by_id(ref['target_tag_id'])

                if source_tag and target_tag:
                    source_key = f"{source_tag['category']}:{source_tag['identifier']}"
                    target_key = f"{target_tag['category']}:{target_tag['identifier']}"

                    if source_key not in reference_map:
                        reference_map[source_key] = []
                    reference_map[source_key].append(target_key)

            json_data["references"] = reference_map

            # 파일로 저장
            with open(json_path, 'w', encoding='utf-8') as f:
                json.dump(json_data, f, indent=2, ensure_ascii=False)

        except Exception as e:
            raise Exception(f"JSON 내보내기 실패: {e}")

    def setup_file_watching(self, project_root: Path, patterns: list[str] | None = None) -> None:
        """파일 감시 설정"""
        try:
            # 기본 패턴 설정
            if not patterns:
                patterns = ['*.py', '*.md', '*.txt', '*.js', '*.ts', '*.java', '*.cpp', '*.h']

            # IndexManager 초기화
            self.index_manager = TagIndexManager(
                project_root=project_root,
                index_file=project_root / ".moai" / "indexes" / "tags.json",
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
