#!/usr/bin/env python3
"""
Detect project languages for MoAI-ADK.
Prints a JSON list to stdout when executed directly.
"""

from __future__ import annotations
import json
from pathlib import Path
from typing import List


def detect_project_languages(root: Path) -> List[str]:
    langs: List[str] = []
    if (root / "pyproject.toml").exists() or list(root.rglob("*.py")):
        langs.append("python")
    if (root / "package.json").exists() or list(root.rglob("*.{js,jsx,ts,tsx}")):
        langs.append("javascript")
        if list(root.rglob("*.ts")) or (root / "tsconfig.json").exists():
            langs.append("typescript")
    if (root / "go.mod").exists() or list(root.rglob("*.go")):
        langs.append("go")
    if (root / "Cargo.toml").exists() or list(root.rglob("*.rs")):
        langs.append("rust")
    if (root / "pom.xml").exists() or (root / "build.gradle").exists() or (root / "build.gradle.kts").exists() or list(root.rglob("*.java")):
        langs.append("java")
    if list(root.rglob("*.sln")) or list(root.rglob("*.csproj")) or list(root.rglob("*.cs")):
        langs.append("csharp")
    if list(root.rglob("*.c")) or list(root.rglob("*.cpp")) or (root / "CMakeLists.txt").exists():
        langs.append("cpp")
    return list(dict.fromkeys(langs))


def main() -> None:
    root = Path.cwd()
    langs = detect_project_languages(root)
    print(json.dumps(langs))


if __name__ == "__main__":
    main()

