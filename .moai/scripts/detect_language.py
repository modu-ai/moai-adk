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
        # React Native 감지: package.json + (ios/ 또는 android/ 디렉토리)
        if (root / "ios").exists() or (root / "android").exists():
            try:
                package_json = root / "package.json"
                if package_json.exists():
                    with open(package_json) as f:
                        data = json.load(f)
                        if "react-native" in data.get("dependencies", {}) or "react-native" in data.get("devDependencies", {}):
                            langs.append("react-native")
            except:
                # Fallback: ios/android 디렉토리 존재하면 React Native로 추정
                langs.append("react-native")
    if (root / "go.mod").exists() or list(root.rglob("*.go")):
        langs.append("go")
    if (root / "Cargo.toml").exists() or list(root.rglob("*.rs")):
        langs.append("rust")
    if (root / "pom.xml").exists() or (root / "build.gradle").exists() or (root / "build.gradle.kts").exists() or list(root.rglob("*.java")):
        langs.append("java")
    # Kotlin 감지 (Android 포함)
    if list(root.rglob("*.kt")) or list(root.rglob("*.kts")):
        langs.append("kotlin")
    if list(root.rglob("*.sln")) or list(root.rglob("*.csproj")) or list(root.rglob("*.cs")):
        langs.append("csharp")
    if list(root.rglob("*.c")) or list(root.rglob("*.cpp")) or (root / "CMakeLists.txt").exists():
        langs.append("cpp")
    # Swift 감지 (iOS/macOS)
    if (root / "Package.swift").exists() or list(root.rglob("*.swift")) or list(root.rglob("*.xcodeproj")):
        langs.append("swift")
    # Dart/Flutter 감지
    if (root / "pubspec.yaml").exists() or list(root.rglob("*.dart")):
        langs.append("dart")
        if (root / "pubspec.yaml").exists():
            langs.append("flutter")
    return list(dict.fromkeys(langs))


def main() -> None:
    root = Path.cwd()
    langs = detect_project_languages(root)
    print(json.dumps(langs))


if __name__ == "__main__":
    main()

