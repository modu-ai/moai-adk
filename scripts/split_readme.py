#!/usr/bin/env python3
# @CODE:DOCS-020 | SPEC: SPEC-DOCS-001

"""
README.ko.md 분할 문서 생성 스크립트
- 3291줄의 단일 파일을 개별 문서로 분할
- 의존성 관계 계층 구조 적용
- 파일 구조 생성 및 내용 추출
"""

import os
import re
import json
from pathlib import Path
from typing import Dict, List, Tuple, Optional
import argparse


class ReadmeSplitter:
    """README.ko.md 분할 클래스"""

    def __init__(self, readme_path: str = "README.ko.md"):
        self.readme_path = Path(readme_path)
        self.backup_dir = Path("backups")
        self.docs_base = Path("docs")

        # 분할 설정
        self.config = {
            "headers_mapping": {
                "## 📚 빠른 시작": "docs/getting-started/getting-started.md",
                "## 🚀 3분 초고속 시작": "docs/getting-started/quick-start.md",
                "## 🚀 빠른 시작 가이드": "docs/getting-started/quick-start.md",
                "## 🔄 4단계 개발 워크플로우": "docs/guides/workflow.md",
                "## 🏗️ 핵심 아키텍처": "docs/guides/architecture-guide.md",  # kebab-case로 변경
                "## 🚀 첫 10분 실습: Hello World API": "docs/examples/hello-world-api.md",
                "## 첫 번째 실습: Todo API 예제": "docs/examples/todo-api-example.md",
                "## Sub-agent & Skills 개요": "docs/api/agents-skills.md",
                "## 🎯 Skills System 최신 개선 사항": "docs/api/skills-system.md",
                "## AI 모델 선택 가이드": "docs/api/model-selection.md",
                "## Claude Code Hooks 가이드": "docs/api/hooks-guide.md",
                "## 🔧 초보자를 위한 문제 해결": "docs/community/troubleshooting.md",
                "## 🚀 빠른 이슈 생성": "docs/community/troubleshooting.md",
                "## 자주 묻는 질문 (FAQ)": "docs/community/faq.md",
                "## 최신 업데이트": "docs/community/changelog.md",
                "## 추가 자료": "docs/community/additional-resources.md",
                "## 🌐 온라인 문서 포털": "docs/community/community.md",
                "## 커뮤니티 & 지원": "docs/community/community.md"
            },
            "size_limits": {
                "min_lines": 100,
                "max_lines": 500
            },
            "excluded_sections": ["## 최신 업데이트"],
            "navigation_headers": [
                "## 📚 빠른 시작",
                "## 🎯 MoAI-ADK란?",
                "## 🎯 핵심 3대 약속",
                "## 📝 개발자 안내"
            ]
        }

        # 의존성 계층 구조
        self.dependency_hierarchy = {
            'index.md': ['docs/getting-started/getting-started.md'],
            'docs/getting-started/getting-started.md': ['docs/getting-started/quick-start.md'],
            'docs/getting-started/quick-start.md': ['docs/guides/workflow.md'],
            'docs/guides/workflow.md': ['docs/guides/architecture-guide.md', 'docs/guides/tdd-guide.md'],
            'docs/guides/architecture-guide.md': [],
            'docs/guides/tdd-guide.md': ['docs/examples/hello-world-api.md', 'docs/examples/todo-api-example.md'],
            'docs/examples/hello-world-api.md': ['docs/api/agents-skills.md'],
            'docs/examples/todo-api-example.md': ['docs/api/agents-skills.md'],
            'docs/api/agents-skills.md': ['docs/api/skills-system.md'],
            'docs/api/skills-system.md': ['docs/api/model-selection.md', 'docs/api/hooks-guide.md'],
            'docs/api/model-selection.md': [],
            'docs/api/hooks-guide.md': [],
            'docs/community/troubleshooting.md': [],
            'docs/community/community.md': ['docs/community/faq.md'],
            'docs/community/faq.md': ['docs/community/changelog.md'],
            'docs/community/changelog.md': ['docs/community/additional-resources.md'],
            'docs/community/additional-resources.md': []
        }

    def create_backup(self) -> Path:
        """README.ko.md 백업 생성"""
        if not self.backup_dir.exists():
            self.backup_dir.mkdir(parents=True, exist_ok=True)

        timestamp = "2025-11-06T12:00:00"  # 고정 타임스탬프 테스트용
        backup_file = self.backup_dir / f"README.ko.md.bak"

        # 현재 README를 백업
        content = self.readme_path.read_text(encoding='utf-8')
        backup_file.write_text(content, encoding='utf-8')

        print(f"✅ 백업 생성 완료: {backup_file}")
        return backup_file

    def create_directory_structure(self):
        """문서 디렉토리 구조 생성"""
        directories = [
            "docs/getting-started",
            "docs/guides",
            "docs/examples",
            "docs/api",
            "docs/community"
        ]

        for dir_path in directories:
            Path(dir_path).mkdir(parents=True, exist_ok=True)
            print(f"✅ 디렉토리 생성: {dir_path}")

    def extract_section_content(self, content: str, header: str) -> str:
        """특정 헤더 섹션 내용 추출"""
        pattern = re.compile(rf'^{header}\n(.*?)(?=\n^#{1,3}\s|\Z)', re.MULTILINE | re.DOTALL)
        match = pattern.search(content)
        return match.group(1).strip() if match else ""

    def split_content_by_header(self, content: str) -> Dict[str, str]:
        """헤더별로 내용 분할"""
        sections = {}

        # 주요 헤더 패턴
        header_pattern = re.compile(r'^(#{1,3})\s+(.+)$', re.MULTILINE)

        # 모든 헤더 찾기
        headers = []
        for match in header_pattern.finditer(content):
            level = len(match.group(1))
            title = match.group(2)
            headers.append((match.start(), level, title))

        # 각 헤더별 내용 추출
        for i, (start, level, title) in enumerate(headers):
            if i < len(headers) - 1:
                next_start = headers[i + 1][0]
                section_content = content[start:next_start]
            else:
                section_content = content[start:]

            sections[title] = section_content.strip()

        return sections

    def generate_document(self, header: str, content: str, output_path: str):
        """개별 문서 생성"""
        file_path = Path(output_path)

        # 상위 디렉토리 생성
        file_path.parent.mkdir(parents=True, exist_ok=True)

        # 문서 내용 생성
        document_content = f"""# {header}

{content}
"""

        file_path.write_text(document_content, encoding='utf-8')
        print(f"✅ 문서 생성 완료: {file_path} ({len(content.split())} 단어)")

    def create_navigation_file(self):
        """네비게이션 파일 생성"""
        nav_content = """# 문서 네비게이션

## 📚 시작 가이드
- [시작 안내](index.md)
- [3분 초고속 시작](docs/getting-started/getting-started.md)
- [빠른 시작 가이드](docs/getting-started/quick-start.md)

## 🎯 핵심 기능
- [4단계 개발 워크플로우](docs/guides/workflow.md)
- [핵심 아키텍처](docs/guides/architecture-guide.md)
- [TDD 실습 가이드](docs/guides/tdd-guide.md)

## 🚀 실습 예제
- [Hello World API](docs/examples/hello-world-api.md)
- [Todo API 예제](docs/examples/todo-api-example.md)

## 🔧 API & Skills
- [Sub-agent & Skills](docs/api/agents-skills.md)
- [Skills System](docs/api/skills-system.md)
- [AI 모델 선택 가이드](docs/api/model-selection.md)
- [Claude Code Hooks](docs/api/hooks-guide.md)

## 🤝 커뮤니티
- [문제 해결](docs/community/troubleshooting.md)
- [커뮤니티 & 지원](docs/community/community.md)
- [자주 묻는 질문](docs/community/faq.md)
- [최신 업데이트](docs/community/changelog.md)
- [추가 자료](docs/community/additional-resources.md)
"""

        nav_path = Path("docs/README.md")
        nav_path.write_text(nav_content, encoding='utf-8')
        print(f"✅ 네비게이션 파일 생성 완료: {nav_path}")

    def create_dependency_map(self):
        """의존성 맵 생성"""
        dependency_map = {
            "version": "1.0.0",
            "created_at": "2025-11-06T12:00:00",
            "hierarchy": self.dependency_hierarchy,
            "total_documents": len(self.dependency_hierarchy),
            "document_categories": {
                "getting-started": ["docs/getting-started/getting-started.md", "docs/getting-started/quick-start.md"],
                "guides": ["docs/guides/workflow.md", "docs/guides/architecture-guide.md", "docs/guides/tdd-guide.md"],
                "examples": ["docs/examples/hello-world-api.md", "docs/examples/todo-api-example.md"],
                "api": ["docs/api/agents-skills.md", "docs/api/skills-system.md", "docs/api/model-selection.md", "docs/api/hooks-guide.md"],
                "community": ["docs/community/troubleshooting.md", "docs/community/community.md", "docs/community/faq.md", "docs/community/changelog.md", "docs/community/additional-resources.md"]
            }
        }

        map_path = Path("docs/dependency-map.json")
        map_path.write_text(json.dumps(dependency_map, indent=2, ensure_ascii=False), encoding='utf-8')
        print(f"✅ 의존성 맵 생성 완료: {map_path}")

    def create_split_config(self):
        """분할 설정 파일 생성"""
        config_path = Path("config/split-config.json")
        config_path.parent.mkdir(parents=True, exist_ok=True)

        config_content = {
            "version": "1.0.0",
            "created_at": "2025-11-06T12:00:00",
            "headers_mapping": self.config["headers_mapping"],
            "size_limits": self.config["size_limits"],
            "excluded_sections": self.config["excluded_sections"],
            "navigation_headers": self.config["navigation_headers"]
        }

        config_path.write_text(json.dumps(config_content, indent=2, ensure_ascii=False), encoding='utf-8')
        print(f"✅ 분할 설정 파일 생성 완료: {config_path}")

    def generate_all_documents(self):
        """모든 문서 생성 실행"""
        print("🚀 README.ko.md 분할 문서 생성 시작...")

        # 1. 백업 생성
        self.create_backup()

        # 2. 디렉토리 구조 생성
        self.create_directory_structure()

        # 3. README 내용 읽기
        content = self.readme_path.read_text(encoding='utf-8')

        # 4. 헤더별 내용 분할
        sections = self.split_content_by_header(content)

        # 5. 개별 문서 생성
        matched_headers = 0
        for header, section_content in sections.items():
            if header in self.config["headers_mapping"]:
                output_path = self.config["headers_mapping"][header]
                self.generate_document(header, section_content, output_path)
                matched_headers += 1
                print(f"✅ 헤더 매칭 성공: {header} -> {output_path}")
            else:
                print(f"⚠️ 헤더 매칭 실패: {header}")

        print(f"✅ 총 {matched_headers}개 헤더 매칭 완료")

        # 6. 메인 인덱스 생성
        index_content = self.extract_section_content(content, "## 🌐 온라인 문서 포털")
        self.generate_document("MoAI-ADK 문서 포털", index_content, "index.md")

        # 7. 네비게이션 파일 생성
        self.create_navigation_file()

        # 8. 의존성 맵 생성
        self.create_dependency_map()

        # 9. 분할 설정 생성
        self.create_split_config()

        print("✅ 모든 문서 생성 완료!")

        # 생성된 파일 목록 출력
        print("\n📁 생성된 파일 목록:")
        generated_files = list(Path(".").rglob("*.md")) + list(Path("config").rglob("*.json"))
        for file in sorted(generated_files):
            if file.suffix == '.md' and 'docs' in str(file):
                print(f"   - {file}")
            elif file.suffix == '.json' and 'split-config' in str(file):
                print(f"   - {file}")

    def validate_split_results(self):
        """분결과 검증"""
        validation_results = {
            "backup_exists": self.backup_dir.exists(),
            "directories_created": all(Path(d).exists() for d in [
                "docs/getting-started", "docs/guides", "docs/examples", "docs/api", "docs/community"
            ]),
            "config_created": Path("config/split-config.json").exists(),
            "dependency_map_created": Path("docs/dependency-map.json").exists(),
            "navigation_created": Path("docs/README.md").exists()
        }

        print("\n🔍 분결과 검증:")
        for check, result in validation_results.items():
            status = "✅" if result else "❌"
            print(f"   {status} {check}")

        return all(validation_results.values())


def main():
    """메인 실행 함수"""
    parser = argparse.ArgumentParser(description="README.ko.md 분할 문서 생성")
    parser.add_argument("--readme", default="README.ko.md", help="README 파일 경로")
    parser.add_argument("--validate", action="store_true", help="결과 검증")

    args = parser.parse_args()

    splitter = ReadmeSplitter(args.readme)
    splitter.generate_all_documents()

    if args.validate:
        is_valid = splitter.validate_split_results()
        if not is_valid:
            print("⚠️ 일부 검증 항목이 실패했습니다")
            exit(1)
        print("✅ 모든 검증 항목 통과")


if __name__ == "__main__":
    main()