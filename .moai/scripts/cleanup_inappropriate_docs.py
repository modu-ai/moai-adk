#!/usr/bin/env python3
"""
부적절한 문서 정리 스크립트

프로젝트 유형에 맞지 않는 문서를 감지하고 정리합니다.

@DESIGN:DOC-CLEANUP-001 - 프로젝트 유형별 문서 정리 전략
@TASK:DOC-CONDITIONAL-002 - 부적절한 문서 자동 정리
"""

import sys
from pathlib import Path
from detect_project_type import ProjectTypeDetector


class DocumentCleaner:
    """프로젝트 유형에 맞지 않는 문서를 정리하는 클래스"""

    def __init__(self, project_path: str = "."):
        self.project_path = Path(project_path).resolve()
        self.detector = ProjectTypeDetector(project_path)

    def analyze_inappropriate_docs(self) -> dict:
        """부적절한 문서 분석"""
        project_info = self.detector.detect_project_type()
        project_type = project_info["project_type"]
        required_docs = set(project_info["required_docs"])

        # 현재 존재하는 문서들
        docs_dir = self.project_path / "docs"
        existing_docs = []
        if docs_dir.exists():
            for doc_file in docs_dir.rglob("*.md"):
                existing_docs.append(doc_file.name)

        # 부적절한 문서 규칙
        inappropriate_rules = {
            "API.md": ["cli_tool", "frontend", "application"],
            "CLI_COMMANDS.md": ["web_api", "frontend", "library"],
            "components.md": ["web_api", "cli_tool", "library"],
            "endpoints.md": ["cli_tool", "frontend", "application"]
        }

        # 부적절한 문서 찾기
        inappropriate_docs = []
        for doc_name, inappropriate_types in inappropriate_rules.items():
            if doc_name in existing_docs and project_type in inappropriate_types:
                inappropriate_docs.append({
                    "file": doc_name,
                    "reason": f"{project_type} 프로젝트에는 부적절",
                    "action": "remove" if self._should_remove(doc_name) else "rename"
                })

        # 누락된 필수 문서
        missing_docs = []
        for required_doc in required_docs:
            if required_doc not in existing_docs:
                missing_docs.append(required_doc)

        return {
            "project_type": project_type,
            "inappropriate_docs": inappropriate_docs,
            "missing_docs": missing_docs,
            "suggestions": self._generate_suggestions(project_type, inappropriate_docs, missing_docs)
        }

    def _should_remove(self, doc_name: str) -> bool:
        """문서를 제거할지 이름 변경할지 결정"""
        # MoAI-ADK 패키지 자체는 API.md가 필요하므로 이름 변경
        if doc_name == "API.md" and self._is_moai_package():
            return False
        return True

    def _is_moai_package(self) -> bool:
        """현재 프로젝트가 MoAI-ADK 패키지인지 확인"""
        setup_py = self.project_path / "setup.py"
        pyproject_toml = self.project_path / "pyproject.toml"

        if setup_py.exists():
            content = setup_py.read_text()
            if "moai-adk" in content.lower() or "moai_adk" in content.lower():
                return True

        if pyproject_toml.exists():
            content = pyproject_toml.read_text()
            if "moai-adk" in content.lower() or "moai_adk" in content.lower():
                return True

        return False

    def _generate_suggestions(self, project_type: str, inappropriate_docs: list, missing_docs: list) -> list:
        """개선 제안 생성"""
        suggestions = []

        if inappropriate_docs:
            suggestions.append(f"✨ {len(inappropriate_docs)}개의 부적절한 문서가 발견되었습니다:")
            for doc in inappropriate_docs:
                if doc["action"] == "remove":
                    suggestions.append(f"  - {doc['file']}: 삭제 권장 ({doc['reason']})")
                else:
                    suggestions.append(f"  - {doc['file']}: 이름 변경 권장 ({doc['reason']})")

        if missing_docs:
            suggestions.append(f"📝 {len(missing_docs)}개의 필수 문서가 누락되었습니다:")
            for doc in missing_docs:
                suggestions.append(f"  - {doc}: {project_type} 프로젝트에 필요")

        if not inappropriate_docs and not missing_docs:
            suggestions.append("✅ 모든 문서가 프로젝트 유형에 적합합니다!")

        return suggestions


def main():
    """CLI 실행 함수"""
    project_path = sys.argv[1] if len(sys.argv) > 1 else "."

    cleaner = DocumentCleaner(project_path)
    analysis = cleaner.analyze_inappropriate_docs()

    print(f"🔍 프로젝트 유형: {analysis['project_type']}")
    print()

    for suggestion in analysis['suggestions']:
        print(suggestion)

    if analysis['inappropriate_docs'] or analysis['missing_docs']:
        print("\n🛠️ 자동 정리를 실행하시겠습니까? (y/N): ", end="")
        if input().lower() == 'y':
            print("자동 정리 기능은 아직 구현되지 않았습니다.")
            print("위의 제안을 참고하여 수동으로 정리해 주세요.")


if __name__ == "__main__":
    main()
