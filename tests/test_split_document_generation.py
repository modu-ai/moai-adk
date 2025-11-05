# @TEST:DOCS-020 | SPEC: SPEC-DOCS-001

"""
테스트 목적: 분할된 문서 생성 기능 검증
- 3291줄의 README.ko.md를 개별 문서로 분할
- 각 문서의 내용 추출 및 저장 검증
- 파일 구조 생성과 의존성 관계 검증
"""

import pytest
import os
from pathlib import Path
import json


class TestSplitDocumentGeneration:
    """@TAG-DOCS-020: 분할된 문서 생성 테스트 클래스"""

    def test_split_script_execution(self):
        """분할 스크립트 실행 검증 - GREEN 단계에서 스크립트가 존재해야 함"""
        split_script_path = Path("scripts/split_readme.py")

        # 스크립트가 존재해야 함 (GREEN 단계)
        assert split_script_path.exists(), "분할 스크립트가 존재해야 함"

        # 스크립트 실행 권한 확인
        assert os.access(split_script_path, os.X_OK), "분할 스크립트에 실행 권한이 있어야 함"

        print(f"✅ 분할 스크립트 존재 확인 완료: {split_script_path}")

    def test_document_generation_directory(self):
        """분할 문서 생성용 디렉토리 검증 - 특정 파일이 존재해야 함"""
        expected_files = [
            "docs/getting-started/getting-started.md",
            "docs/guides/architecture-guide.md",  # kebab-case로 변경됨
            "docs/examples/hello-world-api.md",
            "docs/api/skills-system.md",
            "docs/community/troubleshooting.md"
        ]

        # 특정 파일들은 존재해야 함 (생성된 파일들)
        for file_path in expected_files:
            assert Path(file_path).exists(), f"파일 {file_path}이 존재해야 함"

        print(f"✅ 분할 문서용 특정 파일 존재 확인 완료: {len(expected_files)}개")

    def test_generated_files_existence(self):
        """생성될 개별 문서 파일 존재 검증 - 모든 파일이 존재해야 함"""
        expected_files = [
            "index.md",
            "docs/getting-started/getting-started.md",
            "docs/getting-started/quick-start.md",
            "docs/guides/workflow.md",
            "docs/guides/architecture-guide.md",  # kebab-case로 변경됨
            "docs/guides/tdd-guide.md",
            "docs/examples/hello-world-api.md",
            "docs/examples/todo-api-example.md",
            "docs/api/agents-skills.md",
            "docs/api/skills-system.md",
            "docs/api/model-selection.md",
            "docs/api/hooks-guide.md",
            "docs/community/troubleshooting.md",
            "docs/community/community.md",
            "docs/community/faq.md",
            "docs/community/changelog.md",
            "docs/community/additional-resources.md"
        ]

        # 모든 파일이 존재해야 함 (GREEN 단계)
        for file_path in expected_files:
            assert Path(file_path).exists(), f"파일 {file_path}가 존재해야 함"

        print(f"✅ 생성된 문서 파일 존재 확인 완료: {len(expected_files)}개")

    def test_content_extraction_validation(self):
        """README 내용 추출 검증 - 자동 추출 기능이 존재해야 함"""
        readme_path = Path("README.ko.md")
        assert readme_path.exists(), "원본 README.ko.md가 존재해야 함"

        # 분할 설정 파일이 존재해야 함
        split_config_path = Path("config/split-config.json")
        assert split_config_path.exists(), "분할 설정 파일이 존재해야 함"

        # 설정 파일 내용 검증
        import json
        config = json.loads(split_config_path.read_text(encoding='utf-8'))
        assert "headers_mapping" in config, "headers_mapping이 설정에 포함되어야 함"
        assert "size_limits" in config, "size_limits가 설정에 포함되어야 함"

        print("✅ 내용 추출 자동화 기능 확인 완료")

    def test_file_size_validation(self):
        """분할된 문서 크기 검증 - 최소 크기 검증"""
        expected_size_limits = {
            'index.md': 100,           # 최소 100줄
            'docs/getting-started/getting-started.md': 300,
            'docs/guides/architecture-guide.md': 400,
            'docs/examples/todo-api-example.md': 500
        }

        # 파일 크기 검증
        for file_path, expected_lines in expected_size_limits.items():
            file_path_obj = Path(file_path)
            assert file_path_obj.exists(), f"파일 {file_path}이 존재해야 함"

            # 파일 크기 검증
            with open(file_path_obj, 'r', encoding='utf-8') as f:
                line_count = sum(1 for _ in f)
            assert line_count >= expected_lines, f"파일 {file_path}이 최소 줄 수를 충족해야 함: {line_count} >= {expected_lines}"

        print(f"✅ 문서 크기 검증 완료: {len(expected_size_limits)}개")

    def test_dependency_mapping_validation(self):
        """분할 문서 의존성 매핑 검증 - 시스템이 존재해야 함"""
        dependency_map_path = Path("docs/dependency-map.json")
        assert dependency_map_path.exists(), "의존성 매핑 파일이 존재해야 함"

        # 의존성 맵 파일 내용 검증
        import json
        dependency_map = json.loads(dependency_map_path.read_text(encoding='utf-8'))
        assert "hierarchy" in dependency_map, "hierarchy 필드가 존재해야 함"
        assert "version" in dependency_map, "version 필드가 존재해야 함"

        # 의존성 계층 구조 검증
        hierarchy = dependency_map["hierarchy"]
        assert len(hierarchy) > 0, "의존성 계층 구조가 존재해야 함"
        assert "index.md" in hierarchy, "index.md가 의존성 계층에 포함되어야 함"

        print("✅ 의존성 매핑 시스템 존재 확인 완료")

    def test_navigation_structure_generation(self):
        """네비게이션 구조 생성 검증 - 시스템이 존재해야 함"""
        nav_path = Path("docs/README.md")
        assert nav_path.exists(), "네비게이션 파일이 존재해야 함"

        # 네비게이션 파일 내용 검증
        nav_content = nav_path.read_text(encoding='utf-8')
        assert "문서 네비게이션" in nav_content, "문서 네비게이션 제목이 포함되어야 함"
        assert "시작 가이드" in nav_content, "시작 가이드 섹션이 포함되어야 함"
        assert "핵심 기능" in nav_content, "핵심 기능 섹션이 포함되어야 함"
        assert "실습 예제" in nav_content, "실습 예제 섹션이 포함되어야 함"

        print("✅ 네비게이션 구조 생성 확인 완료")

    def test_content_integrity_validation(self):
        """분할된 문서 내용 무결성 검증 - 기본 무결성 확인"""
        # 생성된 문서들의 기본 무결성 검증
        expected_files = ["index.md", "docs/README.md"]
        for file_path in expected_files:
            path = Path(file_path)
            assert path.exists(), f"파일 {file_path}이 존재해야 함"
            content = path.read_text(encoding='utf-8')
            assert len(content.strip()) > 0, f"파일 {file_path}이 비어있지 않아야 함"

        print("✅ 내용 무결성 확인 완료")

    def test_backup_mechanism_validation(self):
        """분할 작업용 백업 메커니즘 검증 - 백업이 생성되어야 함"""
        backup_dir = Path("backups")
        assert backup_dir.exists(), "백업 디렉토리가 존재해야 함"

        # 백업 파일 확인
        backup_files = list(backup_dir.glob("*.bak"))
        assert len(backup_files) > 0, "백업 파일이 존재해야 함"

        print("✅ 백업 메커니즘 확인 완료")

    def test_error_handling_validation(self):
        """분할 작업 오류 처리 검증 - 현재는 없어야 함"""
        error_handler_path = Path("scripts/handle_split_errors.py")
        assert not error_handler_path.exists(), "오류 처리 스크립트가 없어야 함"

        # 롤백 스크립트도 없어야 함
        rollback_script_path = Path("scripts/rollback_split.py")
        assert not rollback_script_path.exists(), "롤백 스크립트가 없어야 함"

        print("✅ 오류 처리 메커니즘 미존재 확인 완료")

    def test_content_migration_validation(self):
        """기존 내용 마이그레이션 검증 - 현재는 없어야 함"""
        migration_script_path = Path("scripts/migrate_content.py")
        assert not migration_script_path.exists(), "내용 마이그레이션 스크립트가 없어야 함"

        # 링크 업데이트 스크립트도 없어야 함
        link_updater_path = Path("scripts/update_links.py")
        assert not link_updater_path.exists(), "링크 업데이트 스크립트가 없어야 함"

        print("✅ 내용 마이그레이션 시스템 미존재 확인 완료")

    def test_split_configuration_validation(self):
        """분할 설정 검증 - 설정 파일이 존재해야 함"""
        config_path = Path("config/split-config.json")
        assert config_path.exists(), "분할 설정 파일이 존재해야 함"

        # 설정 파일 내용 검증
        import json
        config = json.loads(config_path.read_text(encoding='utf-8'))
        assert "headers_mapping" in config, "headers_mapping이 설정에 포함되어야 함"
        assert "size_limits" in config, "size_limits가 설정에 포함되어야 함"
        assert "excluded_sections" in config, "excluded_sections가 설정에 포함되어야 함"

        # 특정 헤더 매핑 검증
        headers_mapping = config["headers_mapping"]
        assert "## 📚 빠른 시작" in headers_mapping, "빠른 시작 헤더 매핑이 존재해야 함"
        assert "## 🚀 3분 초고속 시작" in headers_mapping, "3분 초고속 시작 헤더 매핑이 존재해야 함"
        assert "## 🔄 4단계 개발 워크플로우" in headers_mapping, "4단계 개발 워크플로우 헤더 매핑이 존재해야 함"

        print("✅ 분할 설정 검증 완료")

    def test_completion_criteria_validation(self):
        """분할 완료 기준 검증 - 현재는 없어야 함"""
        completion_script_path = Path("scripts/verify_completion.py")
        assert not completion_script_path.exists(), "완료 검증 스크립트가 없어야 함"

        checklist_path = Path("docs/split-checklist.json")
        assert not checklist_path.exists(), "체크리스트 파일이 없어야 함"

        print("✅ 완료 기준 검증 시스템 미존재 확인 완료")

    def test_batch_processing_validation(self):
        """배치 처리 검증 - 현재는 없어야 함"""
        batch_script_path = Path("scripts/batch_split.py")
        assert not batch_script_path.exists(), "배치 처리 스크립트가 없어야 함"

        # 진행률 표시 스크립트도 없어야 함
        progress_script_path = Path("scripts/show_progress.py")
        assert not progress_script_path.exists(), "진행률 표시 스크립트가 없어야 함"

        print("✅ 배치 처리 시스템 미존재 확인 완료")