"""
Package Optimizer - 패키지 크기 최적화 핵심 모듈

@DESIGN:PKG-ARCH-001 - 클린 아키텍처 기반 패키지 최적화
@REQ:OPT-CORE-001 - 패키지 크기 80% 감소 요구사항 구현
"""

import os
import time
import logging
from typing import Dict, List, Any


class PackageOptimizer:
    """패키지 크기 최적화를 담당하는 핵심 클래스"""

    def __init__(self, target_directory: str):
        """
        PackageOptimizer 초기화

        Args:
            target_directory: 최적화할 디렉터리 경로

        Raises:
            ValueError: 디렉터리가 존재하지 않을 때
        """
        if not os.path.exists(target_directory):
            raise ValueError("Directory does not exist")

        self.target_directory = target_directory
        self.is_initialized = True

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

    def calculate_directory_size(self) -> int:
        """
        디렉터리의 총 크기를 바이트 단위로 계산

        Returns:
            총 디렉터리 크기 (bytes)
        """
        total_size = 0
        try:
            for root, _, files in os.walk(self.target_directory):
                for file in files:
                    file_path = os.path.join(root, file)
                    try:
                        total_size += os.path.getsize(file_path)
                    except (OSError, FileNotFoundError):
                        # 파일에 접근할 수 없는 경우 무시
                        continue
        except OSError:
            # 디렉터리에 접근할 수 없는 경우
            pass

        return total_size

    def identify_optimization_targets(self) -> Dict[str, List[str]]:
        """
        최적화 대상 파일들을 식별

        Returns:
            중복 파일과 큰 파일 목록이 포함된 딕셔너리
        """
        targets: Dict[str, List[str]] = {
            "duplicates": [],
            "large_files": []
        }

        file_content_map: Dict[str, str] = {}
        large_file_threshold = 1000  # 1KB 이상을 큰 파일로 간주

        try:
            for root, _, files in os.walk(self.target_directory):
                for file in files:
                    file_path = os.path.join(root, file)
                    try:
                        # 파일 크기 확인
                        file_size = os.path.getsize(file_path)
                        if file_size > large_file_threshold:
                            targets["large_files"].append(file_path)

                        # 내용 기반 중복 검사
                        try:
                            with open(file_path, 'r', encoding='utf-8',
                                    errors='ignore') as f:
                                content = f.read()[:100]  # 첫 100자만 비교

                            if content in file_content_map:
                                if (file_content_map[content] not in
                                        targets["duplicates"]):
                                    targets["duplicates"].append(
                                        file_content_map[content])
                                if file_path not in targets["duplicates"]:
                                    targets["duplicates"].append(file_path)
                            else:
                                file_content_map[content] = file_path
                        except (UnicodeDecodeError, OSError):
                            # 바이너리 파일이나 읽기 불가능한 파일은 크기로만 비교
                            size_key = str(file_size)
                            if size_key in file_content_map:
                                if (file_content_map[size_key] not in
                                        targets["duplicates"]):
                                    targets["duplicates"].append(
                                        file_content_map[size_key])
                                if file_path not in targets["duplicates"]:
                                    targets["duplicates"].append(file_path)
                            else:
                                file_content_map[size_key] = file_path

                    except (OSError, FileNotFoundError):
                        continue

        except OSError:
            pass

        return targets

    def optimize(self) -> Dict[str, Any]:
        """
        패키지 최적화 실행

        Returns:
            최적화 결과 딕셔너리
        """
        start_time = time.time()

        try:
            # 초기 크기 측정
            initial_size = self.calculate_directory_size()

            # 최적화 대상 식별
            targets = self.identify_optimization_targets()

            # 최적화 실행 (최소 구현)
            files_processed = 0
            duplicates_removed = 0

            # 중복 파일 제거 (핵심 파일 보존 로직 추가)
            errors = []
            if targets["duplicates"]:
                # 핵심 파일 보존하면서 중복 제거
                files_to_remove = []
                for file_path in targets["duplicates"][1:]:
                    file_name = os.path.basename(file_path)
                    # 핵심 파일은 제거하지 않음
                    if file_name not in self.core_files:
                        files_to_remove.append(file_path)

                for file_path in files_to_remove:
                    try:
                        os.remove(file_path)
                        duplicates_removed += 1
                    except (OSError, PermissionError) as e:
                        # 에러 기록하지만 계속 진행
                        errors.append(
                            f"Failed to remove {file_path}: {str(e)}")
                        continue

            files_processed = (len(targets["duplicates"]) +
                             len(targets["large_files"]))

            # 최종 크기 측정
            final_size = self.calculate_directory_size()

            # 감소율 계산
            if initial_size > 0:
                reduction_percentage = (
                    ((initial_size - final_size) / initial_size) * 100)
            else:
                reduction_percentage = 0.0

            optimization_time = time.time() - start_time

            result = {
                "success": True,
                "initial_size": initial_size,
                "final_size": final_size,
                "reduction_percentage": reduction_percentage,
                "metrics": {
                    "files_processed": files_processed,
                    "duplicates_removed": duplicates_removed,
                    "optimization_time": optimization_time
                }
            }

            # 에러가 있으면 errors 필드 추가
            if errors:
                result["errors"] = errors

            return result

        except PermissionError as e:
            return {
                "success": False,
                "error": f"Permission denied: {str(e)}",
                "errors": [str(e)]
            }
        except Exception as e:
            return {
                "success": False,
                "error": f"Optimization failed: {str(e)}",
                "errors": [str(e)]
            }
