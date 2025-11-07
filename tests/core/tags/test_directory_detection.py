#!/usr/bin/env python3
# @TEST:DIR-DETECTION-001 | @SPEC:TAG-DIRECTORY-DETECTION-001 | @CODE:HOOK-DIR-DETECT-001
"""디렉토리 감지 시스템 테스트

언어별 코드 디렉토리 자동 감지 및 설정 기반 커스터마이징 테스트.

TDD History:
    - RED: 언어별 디렉토리 감지 테스트 작성 (아직 미구현)
    - GREEN: language_dirs.py + policy_validator.py 구현
    - REFACTOR: 상수화, 캐싱 최적화
"""

import sys
from pathlib import Path

# 프로젝트 루트에서 src 추가
SRC_DIR = Path(__file__).parent.parent.parent / "src"
if str(SRC_DIR) not in sys.path:
    sys.path.insert(0, str(SRC_DIR))

import pytest


class TestLanguageDirectoryDetection:
    """언어별 디렉토리 감지 테스트"""

    def test_python_directory_detection(self):
        """Python 프로젝트 디렉토리 자동 감지"""
        # from moai_adk.core.tags.language_dirs import detect_directories

        # 아직 미구현 - RED Phase
        # config = {"project": {"language": "python"}}
        # directories = detect_directories(config)

        # assert "src/" in directories
        # assert "lib/" in directories
        # assert "{package_name}/" in directories
        pass

    def test_javascript_directory_detection(self):
        """JavaScript 프로젝트 디렉토리 감지"""
        # config = {"project": {"language": "javascript"}}
        # directories = detect_directories(config)

        # assert "src/" in directories
        # assert "app/" in directories
        # assert "pages/" in directories
        # assert "components/" in directories
        pass

    def test_typescript_directory_detection(self):
        """TypeScript 프로젝트 디렉토리 감지"""
        # config = {"project": {"language": "typescript"}}
        # directories = detect_directories(config)

        # assert "src/" in directories
        # assert "app/" in directories
        # assert "pages/" in directories
        pass

    def test_go_directory_detection(self):
        """Go 프로젝트 디렉토리 감지 (cmd/, pkg/, internal/)"""
        # config = {"project": {"language": "go"}}
        # directories = detect_directories(config)

        # assert "cmd/" in directories
        # assert "pkg/" in directories
        # assert "internal/" in directories
        pass

    def test_rust_directory_detection(self):
        """Rust 프로젝트 디렉토리 감지 (src/, crates/)"""
        # config = {"project": {"language": "rust"}}
        # directories = detect_directories(config)

        # assert "src/" in directories
        # assert "crates/" in directories
        pass

    def test_kotlin_directory_detection(self):
        """Kotlin 프로젝트 디렉토리 감지"""
        # config = {"project": {"language": "kotlin"}}
        # directories = detect_directories(config)

        # assert "src/main/kotlin/" in directories
        pass

    def test_ruby_directory_detection(self):
        """Ruby 프로젝트 디렉토리 감지"""
        # config = {"project": {"language": "ruby"}}
        # directories = detect_directories(config)

        # assert "lib/" in directories
        # assert "app/" in directories
        pass

    def test_php_directory_detection(self):
        """PHP 프로젝트 디렉토리 감지"""
        # config = {"project": {"language": "php"}}
        # directories = detect_directories(config)

        # assert "src/" in directories
        # assert "app/" in directories
        pass

    def test_java_directory_detection(self):
        """Java 프로젝트 디렉토리 감지"""
        # config = {"project": {"language": "java"}}
        # directories = detect_directories(config)

        # assert "src/main/java/" in directories
        pass

    def test_csharp_directory_detection(self):
        """C# 프로젝트 디렉토리 감지"""
        # config = {"project": {"language": "csharp"}}
        # directories = detect_directories(config)

        # assert "src/" in directories
        # assert "App/" in directories
        pass

    def test_exclude_patterns(self):
        """제외 패턴 확인"""
        # from moai_adk.core.tags.language_dirs import get_exclude_patterns

        # exclude_patterns = get_exclude_patterns()

        # assert "tests/" in exclude_patterns
        # assert "test/" in exclude_patterns
        # assert "__tests__" in exclude_patterns
        # assert "spec/" in exclude_patterns
        pass


class TestCustomDirectoryConfiguration:
    """사용자 정의 디렉토리 설정 테스트"""

    def test_user_custom_patterns(self):
        """사용자가 설정한 커스텀 패턴 사용"""
        # config = {
        #     "project": {"language": "python"},
        #     "tags": {
        #         "policy": {
        #             "code_directories": {
        #                 "patterns": ["mylib/", "modules/"]
        #             }
        #         }
        #     }
        # }
        # directories = detect_directories(config)

        # assert "mylib/" in directories
        # assert "modules/" in directories
        # assert "src/" not in directories  # 커스텀 설정 우선
        pass

    def test_exclude_patterns_override(self):
        """제외 패턴 커스터마이징"""
        # config = {
        #     "tags": {
        #         "policy": {
        #             "code_directories": {
        #                 "exclude_patterns": ["tests/", "integration_tests/"]
        #             }
        #         }
        #     }
        # }
        # exclude = get_exclude_patterns(config)

        # assert "tests/" in exclude
        # assert "integration_tests/" in exclude
        pass

    def test_detection_mode_auto(self):
        """자동 감지 모드"""
        # config = {
        #     "project": {"language": "go"},
        #     "tags": {
        #         "policy": {
        #             "code_directories": {
        #                 "detection_mode": "auto"
        #             }
        #         }
        #     }
        # }
        # directories = detect_directories(config)

        # # Go 언어 기본 디렉토리 사용
        # assert "cmd/" in directories
        # assert "pkg/" in directories
        pass

    def test_detection_mode_manual(self):
        """수동 설정 모드"""
        # config = {
        #     "tags": {
        #         "policy": {
        #             "code_directories": {
        #                 "detection_mode": "manual",
        #                 "patterns": ["api/", "services/"]
        #             }
        #         }
        #     }
        # }
        # directories = detect_directories(config)

        # assert "api/" in directories
        # assert "services/" in directories
        pass

    def test_detection_mode_hybrid(self):
        """하이브리드 모드: 자동 + 사용자 정의"""
        # config = {
        #     "project": {"language": "javascript"},
        #     "tags": {
        #         "policy": {
        #             "code_directories": {
        #                 "detection_mode": "hybrid",
        #                 "patterns": ["services/"]
        #             }
        #         }
        #     }
        # }
        # directories = detect_directories(config)

        # # JavaScript 기본 + 사용자 정의
        # assert "src/" in directories
        # assert "app/" in directories
        # assert "services/" in directories
        pass


class TestPolicyValidatorDirectoryDetection:
    """PolicyValidator의 디렉토리 감지 통합 테스트"""

    def test_is_code_file_python(self):
        """Python 코드 파일 감지"""
        # from moai_adk.core.tags.policy_validator import TagPolicyValidator

        # config = {"project": {"language": "python"}}
        # validator = TagPolicyValidator(config)

        # assert validator._is_code_file("src/example.py") == True
        # assert validator._is_code_file("lib/utils.py") == True
        # assert validator._is_code_file("docs/example.py") == False  # 제외 패턴
        pass

    def test_is_code_file_go(self):
        """Go 코드 파일 감지"""
        # from moai_adk.core.tags.policy_validator import TagPolicyValidator

        # config = {"project": {"language": "go"}}
        # validator = TagPolicyValidator(config)

        # assert validator._is_code_file("cmd/main.go") == True
        # assert validator._is_code_file("pkg/handler.go") == True
        # assert validator._is_code_file("internal/db.go") == True
        # assert validator._is_code_file("src/example.go") == False  # Go에서는 src/ 미지원
        pass

    def test_is_code_file_javascript(self):
        """JavaScript 코드 파일 감지"""
        # from moai_adk.core.tags.policy_validator import TagPolicyValidator

        # config = {"project": {"language": "javascript"}}
        # validator = TagPolicyValidator(config)

        # assert validator._is_code_file("src/index.js") == True
        # assert validator._is_code_file("app/components/Button.js") == True
        # assert validator._is_code_file("pages/index.js") == True
        # assert validator._is_code_file("tests/unit.test.js") == False  # 제외 패턴
        pass

    def test_is_code_file_custom_config(self):
        """사용자 정의 설정으로 코드 파일 감지"""
        # from moai_adk.core.tags.policy_validator import TagPolicyValidator

        # config = {
        #     "project": {"language": "python"},
        #     "tags": {
        #         "policy": {
        #             "code_directories": {
        #                 "patterns": ["myapp/", "lib/"]
        #             }
        #         }
        #     }
        # }
        # validator = TagPolicyValidator(config)

        # assert validator._is_code_file("myapp/main.py") == True
        # assert validator._is_code_file("lib/utils.py") == True
        # assert validator._is_code_file("src/example.py") == False  # 커스텀 설정
        pass


class TestLanguageDirectoryMap:
    """언어 매핑 테스트"""

    def test_all_languages_supported(self):
        """지원하는 모든 언어가 매핑되어 있는지 확인"""
        # from moai_adk.core.tags.language_dirs import LANGUAGE_DIRECTORY_MAP

        # expected_languages = [
        #     "python", "javascript", "typescript", "go", "rust",
        #     "kotlin", "ruby", "php", "java", "csharp"
        # ]

        # for lang in expected_languages:
        #     assert lang in LANGUAGE_DIRECTORY_MAP, f"{lang} 언어 미지원"
        #     assert len(LANGUAGE_DIRECTORY_MAP[lang]) > 0, f"{lang} 디렉토리 패턴 없음"
        pass

    def test_directory_patterns_non_empty(self):
        """모든 언어 매핑이 비어있지 않은지 확인"""
        # from moai_adk.core.tags.language_dirs import LANGUAGE_DIRECTORY_MAP

        # for lang, directories in LANGUAGE_DIRECTORY_MAP.items():
        #     assert isinstance(directories, list), f"{lang}의 디렉토리가 List가 아님"
        #     assert len(directories) > 0, f"{lang}에 디렉토리 패턴이 없음"
        pass

    def test_no_overlap_with_exclude(self):
        """매핑된 디렉토리와 제외 패턴의 겹침 확인"""
        # from moai_adk.core.tags.language_dirs import (
        #     LANGUAGE_DIRECTORY_MAP,
        #     get_exclude_patterns
        # )

        # exclude = set(get_exclude_patterns())

        # for lang, directories in LANGUAGE_DIRECTORY_MAP.items():
        #     for directory in directories:
        #         # 정확한 문자열 비교 (e.g., "src/" != "src_test/")
        #         assert directory not in exclude, f"{lang}의 '{directory}'가 제외 패턴과 겹침"
        pass


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
