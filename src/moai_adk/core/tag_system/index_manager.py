"""
@FEATURE:INDEX-MANAGER-001 - 실시간 TAG 인덱스 관리 최소 구현

watchdog 기반 파일 감지 및 인덱스 실시간 갱신
"""

import json
import threading
import logging
from datetime import datetime
from pathlib import Path
from typing import Dict, Any, Optional, Callable
from dataclasses import dataclass
from enum import Enum

import jsonschema
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler, FileSystemEvent

from .parser import TagParser


class WatcherStatus(Enum):
    """파일 감시 상태"""
    STOPPED = "stopped"
    RUNNING = "running"
    ERROR = "error"


@dataclass
class IndexUpdateEvent:
    """인덱스 업데이트 이벤트"""
    event_type: str
    file_path: Path
    timestamp: datetime


class TagIndexManager:
    """
    실시간 TAG 인덱스 관리 최소 구현

    TRUST 원칙 적용:
    - Test First: 테스트 요구사항에 맞춘 최소 구현
    - Readable: 명확한 인덱스 관리 로직
    - Unified: 인덱스 관리 책임만 담당
    """

    def __init__(self, watch_directory: Path, index_file: Path):
        """
        인덱스 관리자 초기화

        Args:
            watch_directory: 감시할 디렉토리
            index_file: 인덱스 파일 경로

        Raises:
            ValueError: 잘못된 입력값
        """
        # 입력 검증
        if not watch_directory:
            raise ValueError("Watch directory cannot be None or empty")
        if not index_file:
            raise ValueError("Index file path cannot be None or empty")

        self.watch_directory = Path(watch_directory)
        self.index_file = Path(index_file)
        self.parser = TagParser()

        self._observer: Optional[Observer] = None
        self._lock = threading.Lock()
        self._status = WatcherStatus.STOPPED

        # 콜백 함수
        self.on_file_changed: Optional[Callable[[IndexUpdateEvent], None]] = None

        # 구조화된 로거 설정
        self._logger = logging.getLogger(__name__)
        if not self._logger.handlers:
            handler = logging.StreamHandler()
            formatter = logging.Formatter(
                '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
            )
            handler.setFormatter(formatter)
            self._logger.addHandler(handler)
            self._logger.setLevel(logging.INFO)

        # JSON 스키마 정의
        self._index_schema = {
            "type": "object",
            "required": ["metadata", "categories", "chains", "files"],
            "properties": {
                "metadata": {
                    "type": "object",
                    "required": ["created_at", "updated_at", "version", "total_tags"],
                    "properties": {
                        "created_at": {"type": "string"},
                        "updated_at": {"type": "string"},
                        "version": {"type": "string"},
                        "total_tags": {"type": "number"}
                    }
                },
                "categories": {"type": "object"},
                "chains": {"type": "array"},
                "files": {"type": "object"}
            }
        }

    @property
    def is_watching(self) -> bool:
        """파일 감시 상태 반환"""
        return self._status == WatcherStatus.RUNNING

    def initialize_index(self) -> None:
        """빈 인덱스 구조 생성"""
        with self._lock:
            now = datetime.now().isoformat()
            empty_index = {
                "metadata": {
                    "created_at": now,
                    "updated_at": now,
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

            with open(self.index_file, 'w', encoding='utf-8') as f:
                json.dump(empty_index, f, indent=2, ensure_ascii=False)

    def start_watching(self) -> None:
        """파일 감시 시작"""
        if self._observer is not None:
            return

        event_handler = _TagFileEventHandler(self)
        self._observer = Observer()
        self._observer.schedule(event_handler, str(self.watch_directory), recursive=True)
        self._observer.start()
        self._status = WatcherStatus.RUNNING

    def stop_watching(self) -> None:
        """파일 감시 중지"""
        if self._observer is not None:
            self._observer.stop()
            self._observer.join()
            self._observer = None

        self._status = WatcherStatus.STOPPED

    def process_file_change(self, file_path: Path, event_type: str) -> None:
        """
        파일 변경 처리

        Args:
            file_path: 변경된 파일 경로
            event_type: 이벤트 타입 (created, modified, deleted)
        """
        # 입력 검증
        if not file_path or not event_type:
            self._logger.warning(f"Invalid file change parameters: {file_path}, {event_type}")
            return

        self._logger.info(f"Processing file change: {event_type} - {file_path}")

        try:
            with self._lock:
                if event_type == "deleted":
                    self._remove_file_from_index(file_path)
                else:
                    self._update_file_in_index(file_path)

                # 이벤트 콜백 호출
                if self.on_file_changed:
                    event = IndexUpdateEvent(
                        event_type=event_type,
                        file_path=file_path,
                        timestamp=datetime.now()
                    )
                    self.on_file_changed(event)

                self._logger.debug(f"Successfully processed {event_type} for {file_path}")

        except Exception as e:
            self._logger.error(f"Error processing file change {file_path}: {e}", exc_info=True)

    def load_index(self) -> Dict[str, Any]:
        """인덱스 로드"""
        if not self.index_file.exists():
            self.initialize_index()

        with open(self.index_file, 'r', encoding='utf-8') as f:
            return json.load(f)

    def save_index(self, index_data: Dict[str, Any]) -> None:
        """인덱스 저장"""
        # 메타데이터 업데이트
        index_data["metadata"]["updated_at"] = datetime.now().isoformat()

        with open(self.index_file, 'w', encoding='utf-8') as f:
            json.dump(index_data, f, indent=2, ensure_ascii=False)

    def validate_index_schema(self, index_data: Dict[str, Any]) -> bool:
        """인덱스 스키마 검증"""
        try:
            jsonschema.validate(index_data, self._index_schema)
            return True
        except jsonschema.ValidationError:
            return False

    def _update_file_in_index(self, file_path: Path) -> None:
        """파일을 인덱스에 업데이트"""
        if not file_path.exists() or not file_path.is_file():
            return

        # 텍스트 파일만 처리
        if file_path.suffix not in ['.md', '.txt', '.py', '.js', '.ts']:
            return

        try:
            content = file_path.read_text(encoding='utf-8')
        except (UnicodeDecodeError, PermissionError):
            return

        # TAG 추출
        tags = self.parser.extract_tags(content)

        # 인덱스 로드
        index_data = self.load_index()

        # 기존 파일 TAG 제거
        self._remove_file_tags_from_categories(str(file_path), index_data)

        # 새 TAG 추가
        if tags:
            index_data["files"][str(file_path)] = [
                {"category": tag.category, "identifier": tag.identifier, "description": tag.description}
                for tag in tags
            ]

            for tag in tags:
                # 카테고리 결정
                category_group = self._get_category_group(tag.category)
                if category_group not in index_data["categories"]:
                    index_data["categories"][category_group] = {}

                if tag.category not in index_data["categories"][category_group]:
                    index_data["categories"][category_group][tag.category] = {}

                index_data["categories"][category_group][tag.category][tag.identifier] = {
                    "description": tag.description or "",
                    "file": str(file_path)
                }
        else:
            # TAG가 없으면 파일 제거
            index_data["files"].pop(str(file_path), None)

        # 총 TAG 수 계산
        total_tags = sum(
            len(category_data)
            for group in index_data["categories"].values()
            for category_data in group.values()
        )
        index_data["metadata"]["total_tags"] = total_tags

        # 저장
        self.save_index(index_data)

    def _remove_file_from_index(self, file_path: Path) -> None:
        """인덱스에서 파일 제거"""
        index_data = self.load_index()
        self._remove_file_tags_from_categories(str(file_path), index_data)
        index_data["files"].pop(str(file_path), None)

        # 총 TAG 수 재계산
        total_tags = sum(
            len(category_data)
            for group in index_data["categories"].values()
            for category_data in group.values()
        )
        index_data["metadata"]["total_tags"] = total_tags

        self.save_index(index_data)

    def _remove_file_tags_from_categories(self, file_path: str, index_data: Dict[str, Any]) -> None:
        """카테고리에서 특정 파일의 TAG 제거"""
        for group in index_data["categories"].values():
            for category_data in group.values():
                tags_to_remove = [
                    tag_id for tag_id, tag_info in category_data.items()
                    if tag_info.get("file") == file_path
                ]
                for tag_id in tags_to_remove:
                    del category_data[tag_id]

    def _get_category_group(self, category: str) -> str:
        """카테고리 그룹 결정"""
        if category in ["REQ", "DESIGN", "TASK", "TEST"]:
            return "PRIMARY"
        elif category in ["VISION", "STRUCT", "TECH", "ADR"]:
            return "STEERING"
        elif category in ["FEATURE", "API", "UI", "DATA"]:
            return "IMPLEMENTATION"
        elif category in ["PERF", "SEC", "DOCS", "TAG"]:
            return "QUALITY"
        else:
            return "UNKNOWN"


class _TagFileEventHandler(FileSystemEventHandler):
    """내부 파일 이벤트 핸들러"""

    def __init__(self, manager: TagIndexManager):
        self.manager = manager

    def on_created(self, event: FileSystemEvent) -> None:
        if not event.is_directory:
            self.manager.process_file_change(Path(event.src_path), "created")

    def on_modified(self, event: FileSystemEvent) -> None:
        if not event.is_directory:
            self.manager.process_file_change(Path(event.src_path), "modified")

    def on_deleted(self, event: FileSystemEvent) -> None:
        if not event.is_directory:
            self.manager.process_file_change(Path(event.src_path), "deleted")