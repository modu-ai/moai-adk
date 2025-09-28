#!/usr/bin/env python3
# @FEATURE:LANGUAGE-DETECT-011
"""
MoAI-ADK Language Detector Hook

세션 시작 시 프로젝트의 주요 언어를 자동 감지하고, 테스트/빌드 도구 힌트를 출력합니다.
설정을 자동 변경하지 않고, 사용자에게 안전한 가이드를 제공합니다.
"""

from __future__ import annotations

import json
from pathlib import Path


def detect_project_languages(root: Path) -> list[str]:
    langs: list[str] = []
    if (root / "pyproject.toml").exists() or list(root.rglob("*.py")):
        langs.append("python")
    if (root / "package.json").exists() or list(root.rglob("*.{js,jsx,ts,tsx}")):
        # 간단히 js/ts를 하나로 처리
        langs.append("javascript")
        if list(root.rglob("*.ts")) or (root / "tsconfig.json").exists():
            if "typescript" not in langs:
                langs.append("typescript")
    if (root / "go.mod").exists() or list(root.rglob("*.go")):
        langs.append("go")
    if (root / "Cargo.toml").exists() or list(root.rglob("*.rs")):
        langs.append("rust")
    if (
        (root / "pom.xml").exists()
        or (root / "build.gradle").exists()
        or (root / "build.gradle.kts").exists()
        or list(root.rglob("*.java"))
    ):
        langs.append("java")
    if (
        list(root.rglob("*.sln"))
        or list(root.rglob("*.csproj"))
        or list(root.rglob("*.cs"))
    ):
        langs.append("csharp")
    if (
        list(root.rglob("*.c"))
        or list(root.rglob("*.cpp"))
        or (root / "CMakeLists.txt").exists()
    ):
        langs.append("cpp")
    return list(dict.fromkeys(langs))  # de-duplicate preserving order


def load_mappings(root: Path) -> dict:
    default = {
        "test_runners": {
            "python": "pytest",
            "javascript": "npm test",
            "typescript": "npm test",
            "go": "go test ./...",
            "rust": "cargo test",
            "java": "gradle test | mvn test",
            "csharp": "dotnet test",
            "cpp": "ctest | make test",
        },
        "linters": {
            "python": "ruff",
            "javascript": "eslint",
            "typescript": "eslint",
            "go": "golangci-lint",
            "rust": "cargo clippy",
            "java": "checkstyle",
            "csharp": "dotnet format",
            "cpp": "clang-tidy",
        },
        "formatters": {
            "python": "black",
            "javascript": "prettier",
            "typescript": "prettier",
            "go": "gofmt",
            "rust": "rustfmt",
            "java": "google-java-format",
            "csharp": "dotnet format",
            "cpp": "clang-format",
        },
    }
    mapping_path = root / ".moai" / "config" / "language_mappings.json"
    if mapping_path.exists():
        try:
            return json.loads(mapping_path.read_text(encoding="utf-8"))
        except Exception:
            return default
    return default


def main() -> None:
    try:
        root = Path.cwd()
        langs = detect_project_languages(root)
        if not langs:
            return
        m = load_mappings(root)
        print("🌐 감지된 언어:", ", ".join(langs))
        hints = []
        for lang in langs:
            tr = m.get("test_runners", {}).get(lang)
            ln = m.get("linters", {}).get(lang)
            fm = m.get("formatters", {}).get(lang)
            hint = f"- {lang}: test={tr or '-'}, lint={ln or '-'}, format={fm or '-'}"
            hints.append(hint)
        if hints:
            print("🔧 권장 도구:")
            for h in hints:
                print(h)
        print("💡 필요 시 /moai:2-build 단계에서 해당 도구를 사용해 TDD를 실행하세요.")
    except Exception:
        # 훅 실패는 세션을 방해하지 않음
        pass


if __name__ == "__main__":
    main()
