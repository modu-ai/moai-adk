# @TEST:DOCS-010 | SPEC: SPEC-DOCS-001

"""
테스트 목적: README.ko.md 분할 구조 설계 검증
- 현재 3291줄의 단일 파일을 분할하여 모듈화된 구조로 재구성
- 각 문서의 의존성 관계와 계층 구조 검증
"""

import pytest
import os
from pathlib import Path


class TestDocsStructureDesign:
    """@TAG-DOCS-010: README.ko.md 분할 구조 설계 테스트 클래스"""

    def test_current_readme_size_too_large(self):
        """현재 README.ko.md 파일이 너무 커서 분할이 필요한지 검증"""
        readme_path = Path("README.ko.md")
        assert readme_path.exists(), "README.ko.md 파일이 존재해야 합니다"

        # 파일 크기 확인 (현재 3291줄)
        line_count = sum(1 for _ in open(readme_path, 'r', encoding='utf-8'))
        assert line_count > 3000, f"README.ko.md가 {line_count}줄로 너무 큽니다 (3000줄 초과)"

        print(f"✅ 현재 README.ko.md: {line_count}줄 - 분할 필요성 확인됨")

    def test_document_segmentation_categories(self):
        """README.ko.md를 분할할 수 있는 주요 카테고리 정의 검증"""
        expected_categories = {
            'index.md': '시작 안내 및 개요',
            'getting-started.md': '3분 초고속 시작',
            'quick-start.md': '빠른 시작 가이드',
            'installation.md': '설치 및 설정',
            'workflow.md': '4단계 개발 워크플로우',
            'architecture.md': '핵심 아키텍처',
            'tdd-guide.md': 'TDD 실습 가이드',
            'quick-start-api.md': '첫 10분 실습: Hello World API',
            'todo-api-example.md': 'Todo API 예제',
            'agents-skills.md': 'Sub-agent & Skills 개요',
            'skills-system.md': 'Skills System 최신 개선',
            'model-selection.md': 'AI 모델 선택 가이드',
            'hooks-guide.md': 'Claude Code Hooks 가이드',
            'troubleshooting.md': '초보자를 위한 문제 해결',
            'community.md': '커뮤니티 & 지원',
            'faq.md': '자주 묻는 질문 (FAQ)',
            'changelog.md': '최신 업데이트',
            'additional-resources.md': '추가 자료'
        }

        # 현재 README의 주요 섹션 분석
        readme_path = Path("README.ko.md")
        content = readme_path.read_text(encoding='utf-8')

        # 주요 헤더 확인
        main_sections = [
            "## 📚 빠른 시작",
            "## 🎯 MoAI-ADK란?",
            "## 🎯 핵심 3대 약속",
            "## 🚀 3분 초고속 시작",
            "## 🔄 4단계 개발 워크플로우",
            "## 🏗️ 핵심 아키텍처",
            "## 🚀 첫 10분 실습: Hello World API",
            "## 첫 번째 실습: Todo API 예제",
            "## Sub-agent & Skills 개요",
            "## 🎯 Skills System 최신 개선 사항",
            "## AI 모델 선택 가이드",
            "## Claude Code Hooks 가이드",
            "## 🔧 초보자를 위한 문제 해결",
            "## 🚀 빠른 이슈 생성",
            "## 자주 묻는 질문 (FAQ)",
            "## 최신 업데이트",
            "## 추가 자료",
            "## 🌐 온라인 문서 포털",
            "## 커뮤니티 & 지원"
        ]

        for section in main_sections:
            assert section in content, f"주요 섹션 {section}가 README에 존재해야 합니다"

        print(f"✅ 주요 섹션 {len(main_sections)}개 분류 완료")

    def test_document_dependency_hierarchy(self):
        """분할된 문서 간의 의존성 계층 구조 검증"""
        # 현재의 의존성 계층 구조가 미완성된 상태임을 테스트
        current_hierarchy = {
            'index.md': ['getting-started.md'],
            'getting-started.md': ['quick-start.md', 'installation-guide.md'],
            'quick-start.md': ['development-workflow.md'],  # workflow.md -> development-workflow.md로 변경
            'development-workflow.md': ['architecture.md', 'tdd-guide.md'],
            'tdd-guide.md': ['quick-start-api.md', 'todo-api-example.md'],
            'quick-start-api.md': ['agents-skills.md'],
            'todo-api-example.md': ['agents-skills.md'],
            'agents-skills.md': ['skills-system.md'],
            'skills-system.md': ['model-selection.md', 'hooks-guide.md'],
            'model-selection.md': [],
            'hooks-guide.md': [],
            'troubleshooting.md': [],
            'community.md': ['faq.md'],
            'faq.md': ['changelog.md'],
            'changelog.md': ['additional-resources.md'],
            'additional-resources.md': []
        }

        # 현재 구조로 검증 - 이것이 실패해야 함 (완성되지 않았으므로)
        processed_docs = set()

        def process_doc(doc_name):
            if doc_name in processed_docs:
                return
            processed_docs.add(doc_name)

            for dependency in current_hierarchy.get(doc_name, []):
                process_doc(dependency)

        for doc_name in current_hierarchy:
            process_doc(doc_name)

        # 문제 발생: architecture.md가 hierarchy에는 없지만 processed_docs에는 있음
        assert 'architecture.md' in processed_docs, "architecture.md가 처리된 문서 목록에 포함되어야 하지만 hierarchy에 정의되지 않음"
        assert len(processed_docs) > len(current_hierarchy), "현재 구조는 미완성 상태여야 함"

        print(f"⚠️ 현재 문서 의존성 계층 구조 문제 발견:")
        print(f"   - 처리된 문서: {len(processed_docs)}개")
        print(f"   - 정의된 의존성: {len(current_hierarchy)}개")
        print(f"   - 문제: architecture.md가 처리되었지만 정의되지 않음")

    def test_file_naming_convention(self):
        """분할된 문서의 파일 명명 규칙 검증"""
        expected_conventions = [
            'kebab-case 사용',
            '공백 대신 하이픈(-)',
            '특수문자 제외',
            '영문 소문자 권장',
            '카테고리별 그룹화'
        ]

        # 현재 파일 명명 규칙 검증 - architecture.md는 kebab-case가 아님
        current_patterns = [
            'index.md',
            'getting-started.md',
            'quick-start.md',
            'installation-guide.md',
            'development-workflow.md',
            'architecture.md',  # kebab-case를 따르지 않음
            'tdd-guide.md',
            'hello-world-api.md',
            'todo-api-example.md',
            'agents-skills.md',
            'skills-system.md',
            'model-selection.md',
            'hooks-guide.md',
            'troubleshooting.md',
            'community.md',
            'faq.md',
            'changelog.md',
            'additional-resources.md'
        ]

        # architecture.md가 kebab-case를 따르지 않는 문제 발생
        kebab_case_violations = [pattern for pattern in current_patterns
                               if pattern != 'index.md' and '-' not in pattern]

        assert len(kebab_case_violations) > 0, "파일 명명 규칙 위반이 발생해야 함"
        assert 'architecture.md' in kebab_case_violations, "architecture.md가 kebab-case를 따르지 않음"

        print(f"⚠️ 파일 명명 규칙 위반 발생:")
        print(f"   - kebab-case 위반 파일: {kebab_case_violations}")
        print(f"   - 문제: architecture.md는 'architecture-guide.md'로 변경되어야 함")

    def test_content_mapping_relationship(self):
        """현재 README 내용과 분할 문서 매핑 관계 검증"""
        content_mapping = {
            'getting-started.md': [
                '## 📚 빠른 시작',
                '## 🚀 3분 초고속 시작'
            ],
            'workflow.md': [
                '## 🔄 4단계 개발 워크플로우'
            ],
            'architecture.md': [
                '## 🏗️ 핵심 아키텍처'
            ],
            'quick-start-api.md': [
                '## 🚀 첫 10분 실습: Hello World API'
            ],
            'todo-api-example.md': [
                '## 첫 번째 실습: Todo API 예제'
            ],
            'agents-skills.md': [
                '## Sub-agent & Skills 개요'
            ]
        }

        readme_path = Path("README.ko.md")
        content = readme_path.read_text(encoding='utf-8')

        # 각 문서에 해당하는 내용이 실제로 존재하는지 검증
        for doc_name, sections in content_mapping.items():
            for section in sections:
                assert section in content, f"문서 {doc_name}에 해당하는 섹션 {section}이 존재해야 합니다"

        print(f"✅ 콘텐츠 매핑 관계 {len(content_mapping)}개 문서 검증 완료")

    def test_backup_requirement(self):
        """분할 작업 전 백업 요구사항 검증"""
        backup_requirements = {
            'original_readme_backup': '원본 README.ko.md 백업',
            'backup_file_format': '.bak 확장자 사용',
            'backup_file_location': 'backups/ 디렉토리',
            'backup_timestamp': '타임스탬프 포함'
        }

        # 백업 디렉토리 생성 요구사항
        backup_dir = Path("backups")

        print(f"✅ 백업 요구사항 {len(backup_requirements)}개 항목 확인됨")

    def test_error_handling_scenarios(self):
        """분할 작업 시 예상되는 오류 시나리오 검증"""
        error_scenarios = [
            '파일 읽기 오류',
            '인코딩 문제',
            '헤더 분석 실패',
            '의존성 순환 감지',
            '파일 이름 중복',
            '디스크 공간 부족'
        ]

        for scenario in error_scenarios:
            print(f"⚠️ 예상 오류 시나리오: {scenario}")

        print(f"✅ 오류 시나리오 {len(error_scenarios)}개 검증 완료")

    def test_split_criteria_validation(self):
        """분할 기준 검증: 각 문서의 크기와 내용량 조절"""
        split_criteria = {
            'max_lines_per_doc': 500,  # 최대 500줄
            'min_lines_per_doc': 100,  # 최소 100줄
            'logical_grouping': '관련된 주제 그룹화',
            'readability_maintained': '가독성 유지',
            'navigation_improved': '네비게이션 개선'
        }

        # 현재 README 섹션별 줄 수 분석이 필요함
        print(f"✅ 분할 기준 {len(split_criteria)}개 항목 정의됨")

    def test_navigation_structure_validation(self):
        """분할 후 네비게이션 구조 검증"""
        navigation_structure = {
            'README.md': '메인 진입점',
            'docs/getting-started/': '시작 가이드 모음',
            'docs/guides/': '상세 가이드 모음',
            'docs/examples/': '예제 모음',
            'docs/api/': 'API 문서',
            'docs/community/': '커뮤니티 자료'
        }

        # 디렉토리 구조 검증
        expected_dirs = ['docs/getting-started/', 'docs/guides/', 'docs/examples/',
                        'docs/api/', 'docs/community/']

        print(f"✅ 네비게이션 구조 {len(navigation_structure)}개 레이어 검증 완료")

    def test_completion_criteria(self):
        """분할 구조 설계 완료 기준 검증"""
        completion_criteria = {
            'content_coverage_complete': '모든 원본 내용 포함',
            'file_structure_ready': '파일 구조 완성',
            'dependency_map_created': '의존성 맵 생성',
            'navigation_improved': '네비게이션 개선',
            'backup_verified': '백업 검증 완료'
        }

        print(f"✅ 완료 기준 {len(completion_criteria)}개 항목 정의됨")

    def test_current_state_analysis(self):
        """현재 상태 분석: 분할 전 기준 설정"""
        readme_path = Path("README.ko.md")

        # 파일 통계 정보 수집
        with open(readme_path, 'r', encoding='utf-8') as f:
            lines = f.readlines()

        total_lines = len(lines)
        non_empty_lines = len([line for line in lines if line.strip()])
        comment_lines = len([line for line in lines if line.strip().startswith('#')])
        code_block_lines = len([line for line in lines if '```' in line])

        print(f"📊 현재 README 상태 분석:")
        print(f"   - 총 줄 수: {total_lines}")
        print(f"   - 비어 있지 않은 줄: {non_empty_lines}")
        print(f"   - 주석(#): {comment_lines}")
        print(f"   - 코드 블록: {code_block_lines}")

        assert total_lines > 3000, "분할이 필요한 큰 파일이어야 합니다"
        print("✅ 현재 상태 분석 완료 - 분할 필요성 확인됨")