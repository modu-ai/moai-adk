"""
Duplicate Remover - 중복 파일 제거 모듈

@DESIGN:TEMPLATE-MERGE-002 - 템플릿 통합 설계 구현
@REQ:OPT-DEDUPE-002 - 중복 제거 자동화 요구사항 구현
"""

import os
import time
import hashlib
import logging
from pathlib import Path
from typing import Dict, List, Any, Optional


class DuplicateRemover:
    """중복 파일 제거를 담당하는 클래스"""

    def __init__(self, target_directory: str):
        """
        DuplicateRemover 초기화

        Args:
            target_directory: 중복 파일을 찾을 디렉터리 경로
        """
        self.target_directory = target_directory
        self.hash_algorithm = "sha256"
        self.exclude_extensions = []  # 제외할 파일 확장자 목록

        # 핵심 파일 보존 목록 (SPEC-003 요구사항)
        self.core_files = {
            "spec-builder.md",
            "code-builder.md",
            "doc-syncer.md",
            "claude-code-manager.md",
            "1-spec.md",
            "2-build.md",
            "3-sync.md"
        }

        # 로깅 설정
        self.logger = logging.getLogger(__name__)

    def calculate_file_hash(self, file_path: str) -> str:
        """
        파일의 해시값을 계산

        Args:
            file_path: 해시를 계산할 파일 경로

        Returns:
            파일의 SHA256 해시값
        """
        hash_obj = hashlib.sha256()
        try:
            with open(file_path, 'rb') as f:
                # 메모리 효율을 위해 청크 단위로 읽기
                for chunk in iter(lambda: f.read(4096), b""):
                    hash_obj.update(chunk)
        except (OSError, PermissionError):
            # 파일을 읽을 수 없는 경우 빈 해시 반환
            return ""

        return hash_obj.hexdigest()

    def find_duplicates(self) -> List[Dict[str, Any]]:
        """
        중복 파일 그룹을 찾기

        Returns:
            중복 파일 그룹 목록
        """
        hash_map = {}
        duplicates = []

        try:
            for root, dirs, files in os.walk(self.target_directory):
                for file in files:
                    file_path = os.path.join(root, file)

                    # 제외할 확장자 체크
                    file_ext = os.path.splitext(file)[1].lower()
                    if file_ext in self.exclude_extensions:
                        continue

                    try:
                        # 파일 해시 계산
                        file_hash = self.calculate_file_hash(file_path)
                        if not file_hash:  # 해시 계산 실패
                            continue

                        if file_hash in hash_map:
                            hash_map[file_hash].append(file_path)
                        else:
                            hash_map[file_hash] = [file_path]

                    except (OSError, PermissionError):
                        continue

        except OSError:
            pass

        # 중복 그룹 생성 (2개 이상의 파일이 있는 해시만)
        for file_hash, file_list in hash_map.items():
            if len(file_list) > 1:
                duplicates.append({
                    "hash": file_hash,
                    "files": file_list
                })

        return duplicates

    def _choose_file_to_preserve(self, file_list: List[str]) -> str:
        """
        중복 파일 그룹에서 보존할 파일을 선택

        Args:
            file_list: 중복 파일 목록

        Returns:
            보존할 파일 경로
        """
        # 핵심 파일이 있으면 그것을 우선 보존
        for file_path in file_list:
            file_name = os.path.basename(file_path)
            if file_name in self.core_files:
                return file_path

        # 핵심 파일이 없으면 가장 짧은 경로의 파일을 선택
        return min(file_list, key=lambda x: len(x))

    def remove_duplicates(self) -> Dict[str, Any]:
        """
        중복 파일 제거 실행

        Returns:
            제거 결과 딕셔너리
        """
        start_time = time.time()
        removed_count = 0
        saved_bytes = 0
        errors = []

        try:
            # 중복 파일 찾기
            duplicate_groups = self.find_duplicates()

            for group in duplicate_groups:
                files = group["files"]
                if len(files) <= 1:
                    continue

                # 보존할 파일 선택
                file_to_preserve = self._choose_file_to_preserve(files)
                files_to_remove = [f for f in files if f != file_to_preserve]

                # 중복 파일 제거
                for file_path in files_to_remove:
                    try:
                        # 파일 크기 기록 (삭제 전)
                        file_size = os.path.getsize(file_path)

                        # 파일 삭제
                        os.remove(file_path)

                        removed_count += 1
                        saved_bytes += file_size

                    except (OSError, PermissionError) as e:
                        errors.append(f"Failed to remove {file_path}: {str(e)}")
                        continue

            processing_time = time.time() - start_time

            return {
                "removed_count": removed_count,
                "saved_bytes": saved_bytes,
                "duplicate_groups": len(duplicate_groups),
                "processing_time": processing_time,
                "errors": errors
            }

        except Exception as e:
            return {
                "removed_count": 0,
                "saved_bytes": 0,
                "duplicate_groups": 0,
                "processing_time": time.time() - start_time,
                "errors": [f"General error: {str(e)}"]
            }