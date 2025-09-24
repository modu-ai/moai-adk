#!/usr/bin/env python3
"""
프로젝트 유형 자동 감지 스크립트

@DESIGN:PROJECT-TYPE-001 - 프로젝트 유형별 문서 생성 전략
@TASK:DOC-CONDITIONAL-001 - 조건부 문서 생성 로직
"""

import json
import sys
from pathlib import Path
from typing import Dict, List, Optional


class ProjectTypeDetector:
    """프로젝트 유형을 자동 감지하는 클래스"""

    def __init__(self, project_path: str = "."):
        self.project_path = Path(project_path).resolve()

    def detect_project_type(self) -> Dict[str, any]:
        """
        프로젝트 유형을 감지하고 필요한 문서 목록을 반환

        Returns:
            프로젝트 유형, 감지된 특징, 필요한 문서 목록
        """
        features = self._analyze_project_features()
        project_type = self._classify_project_type(features)
        required_docs = self._get_required_docs(project_type, features)

        return {
            "project_type": project_type,
            "features": features,
            "required_docs": required_docs,
            "confidence": self._calculate_confidence(features, project_type)
        }

    def _analyze_project_features(self) -> Dict[str, any]:
        """프로젝트의 특징을 분석"""
        features = {
            "has_web_api": self._has_web_api(),
            "has_cli": self._has_cli(),
            "has_library": self._has_library(),
            "has_frontend": self._has_frontend(),
            "language": self._detect_language(),
            "framework": self._detect_framework(),
            "api_files": self._find_api_files(),
            "cli_files": self._find_cli_files(),
            "lib_files": self._find_lib_files()
        }
        return features

    def _has_web_api(self) -> bool:
        """Web API 프로젝트인지 확인"""
        api_indicators = [
            "src/app.py", "src/main.py", "app.py", "main.py",
            "src/api", "api", "src/controllers", "controllers",
            "src/routes", "routes", "src/endpoints", "endpoints",
            "server.js", "server.ts", "src/server.js", "src/server.ts"
        ]

        for indicator in api_indicators:
            if (self.project_path / indicator).exists():
                return True

        # FastAPI, Flask, Express 등의 패턴 확인
        return self._check_api_patterns()

    def _has_cli(self) -> bool:
        """CLI 도구인지 확인"""
        cli_indicators = [
            "src/cli", "cli", "src/commands", "commands",
            "bin/", "scripts/cli"
        ]

        for indicator in cli_indicators:
            if (self.project_path / indicator).exists():
                return True

        # CLI 패턴 확인 (argparse, click 등)
        return self._check_cli_patterns()

    def _has_library(self) -> bool:
        """라이브러리/패키지인지 확인"""
        lib_indicators = [
            "setup.py", "pyproject.toml", "package.json",
            "Cargo.toml", "go.mod", "pom.xml"
        ]

        for indicator in lib_indicators:
            if (self.project_path / indicator).exists():
                return True
        return False

    def _has_frontend(self) -> bool:
        """프론트엔드 프로젝트인지 확인"""
        frontend_indicators = [
            "src/components", "components", "src/pages", "pages",
            "public/index.html", "index.html", "src/App.js", "src/App.tsx"
        ]

        for indicator in frontend_indicators:
            if (self.project_path / indicator).exists():
                return True
        return False

    def _detect_language(self) -> str:
        """주요 언어 감지"""
        language_files = {
            "python": ["*.py", "requirements.txt", "pyproject.toml"],
            "javascript": ["*.js", "package.json", "yarn.lock"],
            "typescript": ["*.ts", "*.tsx", "tsconfig.json"],
            "go": ["*.go", "go.mod", "go.sum"],
            "rust": ["*.rs", "Cargo.toml", "Cargo.lock"],
            "java": ["*.java", "pom.xml", "build.gradle"]
        }

        for lang, patterns in language_files.items():
            for pattern in patterns:
                if list(self.project_path.rglob(pattern)):
                    return lang
        return "unknown"

    def _detect_framework(self) -> Optional[str]:
        """프레임워크 감지"""
        # Python 프레임워크
        if (self.project_path / "requirements.txt").exists():
            content = (self.project_path / "requirements.txt").read_text()
            if "fastapi" in content.lower():
                return "fastapi"
            elif "flask" in content.lower():
                return "flask"
            elif "django" in content.lower():
                return "django"
            elif "click" in content.lower():
                return "click"

        # Node.js 프레임워크
        package_json = self.project_path / "package.json"
        if package_json.exists():
            try:
                data = json.loads(package_json.read_text())
                deps = {**data.get("dependencies", {}), **data.get("devDependencies", {})}

                if "express" in deps:
                    return "express"
                elif "next" in deps or "@next/core" in deps:
                    return "nextjs"
                elif "react" in deps:
                    return "react"
                elif "vue" in deps:
                    return "vue"
            except Exception:
                pass

        return None

    def _find_api_files(self) -> List[str]:
        """API 관련 파일 찾기"""
        api_files = []
        patterns = [
            "**/api/**/*.py",
            "**/routes/**/*.py",
            "**/controllers/**/*.py",
            "**/api/**/*.js",
            "**/routes/**/*.js",
            "**/controllers/**/*.js",
            "**/api/**/*.ts",
            "**/routes/**/*.ts",
            "**/controllers/**/*.ts",
        ]

        for pattern in patterns:
            api_files.extend([
                str(f.relative_to(self.project_path))
                for f in self.project_path.rglob(pattern)
            ])
        return api_files

    def _find_cli_files(self) -> List[str]:
        """CLI 관련 파일 찾기"""
        cli_files = []
        patterns = [
            "**/cli/**/*.py",
            "**/commands/**/*.py",
            "**/cli.py",
            "**/cli/**/*.js",
            "**/commands/**/*.js",
            "**/cli.js",
        ]

        for pattern in patterns:
            cli_files.extend([
                str(f.relative_to(self.project_path))
                for f in self.project_path.rglob(pattern)
            ])
        return cli_files

    def _find_lib_files(self) -> List[str]:
        """라이브러리 관련 파일 찾기"""
        lib_files = []
        patterns = ["src/**/*.py", "lib/**/*.py", "src/**/*.js", "lib/**/*.js"]

        for pattern in patterns:
            lib_files.extend([
                str(f.relative_to(self.project_path))
                for f in self.project_path.rglob(pattern)
            ])
        return lib_files[:10]  # 최대 10개만

    def _check_api_patterns(self) -> bool:
        """코드 내 API 패턴 확인"""
        # Python API 패턴
        for py_file in self.project_path.rglob("*.py"):
            try:
                content = py_file.read_text(encoding='utf-8', errors='ignore')
                api_patterns = [
                    "@app.route", "@router.", "FastAPI(", "Flask(__name__)",
                    "from fastapi", "from flask", "def get_", "def post_"
                ]
                if any(pattern in content for pattern in api_patterns):
                    return True
            except Exception:
                continue

        # JavaScript/TypeScript API 패턴
        for js_file in list(self.project_path.rglob("*.js")) + list(self.project_path.rglob("*.ts")):
            try:
                content = js_file.read_text(encoding='utf-8', errors='ignore')
                api_patterns = [
                    "app.get(", "app.post(", "router.get(", "router.post(",
                    "express()", "fastify()", "app.listen("
                ]
                if any(pattern in content for pattern in api_patterns):
                    return True
            except Exception:
                continue

        return False

    def _check_cli_patterns(self) -> bool:
        """코드 내 CLI 패턴 확인"""
        for py_file in self.project_path.rglob("*.py"):
            try:
                content = py_file.read_text(encoding='utf-8', errors='ignore')
                cli_patterns = [
                    "@click.command", "@click.group", "argparse.ArgumentParser",
                    "if __name__ == '__main__':", "sys.argv", "parser.add_argument"
                ]
                if any(pattern in content for pattern in cli_patterns):
                    return True
            except Exception:
                continue
        return False

    def _classify_project_type(self, features: Dict) -> str:
        """특징을 바탕으로 프로젝트 유형 분류"""
        # 우선순위 기반 분류
        if features["has_web_api"] and features["has_frontend"]:
            return "fullstack"
        elif features["has_web_api"]:
            return "web_api"
        elif features["has_cli"]:
            return "cli_tool"
        elif features["has_frontend"]:
            return "frontend"
        elif features["has_library"]:
            return "library"
        else:
            return "application"

    def _get_required_docs(self, project_type: str, features: Dict) -> List[str]:
        """프로젝트 유형별 필요한 문서 목록"""
        doc_mapping = {
            "web_api": ["API.md", "endpoints.md", "authentication.md"],
            "fullstack": ["API.md", "endpoints.md", "frontend-guide.md", "deployment.md"],
            "cli_tool": ["CLI_COMMANDS.md", "usage.md", "examples.md"],
            "frontend": ["components.md", "user-guide.md", "styling.md"],
            "library": ["API_REFERENCE.md", "modules.md", "examples.md"],
            "application": ["features.md", "user-guide.md", "configuration.md"]
        }

        base_docs = doc_mapping.get(project_type, ["README.md"])

        # 조건부 문서 추가
        if features["api_files"]:
            base_docs.append("api-reference.md")
        if features["framework"]:
            base_docs.append(f"{features['framework']}-guide.md")

        return list(set(base_docs))  # 중복 제거

    def _calculate_confidence(self, features: Dict, project_type: str) -> float:
        """분류 결과의 신뢰도 계산"""
        confidence_factors = {
            "web_api": features["has_web_api"] and len(features["api_files"]) > 0,
            "cli_tool": features["has_cli"] and len(features["cli_files"]) > 0,
            "library": features["has_library"],
            "frontend": features["has_frontend"]
        }

        base_confidence = 0.7 if confidence_factors.get(project_type, False) else 0.3

        # 추가 증거가 있으면 신뢰도 증가
        if features["framework"]:
            base_confidence += 0.2
        if features["language"] != "unknown":
            base_confidence += 0.1

        return min(1.0, base_confidence)


def main():
    """CLI 실행 함수"""
    project_path = sys.argv[1] if len(sys.argv) > 1 else "."

    detector = ProjectTypeDetector(project_path)
    result = detector.detect_project_type()

    if "--json" in sys.argv:
        print(json.dumps(result, indent=2))
    else:
        print(f"프로젝트 유형: {result['project_type']}")
        print(f"신뢰도: {result['confidence']:.1%}")
        print(f"언어: {result['features']['language']}")
        if result['features']['framework']:
            print(f"프레임워크: {result['features']['framework']}")
        print(f"필요한 문서: {', '.join(result['required_docs'])}")


if __name__ == "__main__":
    main()
