# @TEST:CORE-PROJECT-001 | SPEC: SPEC-CORE-PROJECT-001.md
"""LanguageDetector 테스트 스위트

20개 프로그래밍 언어 자동 감지 기능을 검증합니다.
"""

from pathlib import Path

import pytest

from moai_adk.core.project.detector import LanguageDetector


class TestLanguageDetector:
    """LanguageDetector 클래스 테스트"""

    def test_detect_python_by_pyproject_toml(self, tmp_path: Path) -> None:
        """pyproject.toml로 Python 프로젝트 감지"""
        (tmp_path / "pyproject.toml").write_text("[tool.poetry]")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result == "python"

    def test_detect_python_by_py_files(self, tmp_path: Path) -> None:
        """*.py 파일로 Python 프로젝트 감지"""
        (tmp_path / "main.py").write_text("print('hello')")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result == "python"

    def test_detect_typescript_by_tsconfig(self, tmp_path: Path) -> None:
        """tsconfig.json으로 TypeScript 프로젝트 감지"""
        (tmp_path / "tsconfig.json").write_text("{}")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result == "typescript"

    def test_detect_javascript_by_package_json(self, tmp_path: Path) -> None:
        """package.json으로 JavaScript 프로젝트 감지"""
        (tmp_path / "package.json").write_text("{}")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        # javascript와 typescript 모두 package.json을 사용하므로
        # 어느 것이든 OK (우선순위에 따라 결정됨)
        assert result in ["javascript", "typescript"]

    def test_detect_java_by_pom_xml(self, tmp_path: Path) -> None:
        """pom.xml로 Java 프로젝트 감지"""
        (tmp_path / "pom.xml").write_text("<project></project>")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result == "java"

    def test_detect_go_by_go_mod(self, tmp_path: Path) -> None:
        """go.mod로 Go 프로젝트 감지"""
        (tmp_path / "go.mod").write_text("module example")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result == "go"

    def test_detect_rust_by_cargo_toml(self, tmp_path: Path) -> None:
        """Cargo.toml로 Rust 프로젝트 감지"""
        (tmp_path / "Cargo.toml").write_text("[package]")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result == "rust"

    def test_detect_dart_by_pubspec_yaml(self, tmp_path: Path) -> None:
        """pubspec.yaml로 Dart 프로젝트 감지"""
        (tmp_path / "pubspec.yaml").write_text("name: test")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result == "dart"

    def test_detect_swift_by_package_swift(self, tmp_path: Path) -> None:
        """Package.swift로 Swift 프로젝트 감지"""
        (tmp_path / "Package.swift").write_text("// swift-tools-version")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result == "swift"

    def test_detect_kotlin_by_gradle_kts(self, tmp_path: Path) -> None:
        """build.gradle.kts로 Kotlin 프로젝트 감지"""
        (tmp_path / "build.gradle.kts").write_text("plugins {}")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result == "kotlin"

    def test_detect_returns_none_for_empty_directory(self, tmp_path: Path) -> None:
        """빈 디렉토리에서 None 반환"""
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result is None

    def test_detect_in_subdirectory(self, tmp_path: Path) -> None:
        """하위 디렉토리의 파일도 감지"""
        subdir = tmp_path / "src"
        subdir.mkdir()
        (subdir / "main.rs").write_text("fn main() {}")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result == "rust"

    def test_detect_multiple_returns_all_languages(self, tmp_path: Path) -> None:
        """detect_multiple이 모든 언어 반환"""
        (tmp_path / "main.py").write_text("print('hello')")
        (tmp_path / "package.json").write_text("{}")
        detector = LanguageDetector()

        result = detector.detect_multiple(tmp_path)

        assert "python" in result
        # package.json은 typescript 또는 javascript로 감지됨
        assert any(lang in result for lang in ["typescript", "javascript"])

    def test_detect_multiple_empty_directory(self, tmp_path: Path) -> None:
        """빈 디렉토리에서 빈 리스트 반환"""
        detector = LanguageDetector()

        result = detector.detect_multiple(tmp_path)

        assert result == []

    def test_language_patterns_has_20_languages(self) -> None:
        """LANGUAGE_PATTERNS에 20개 언어 정의됨"""
        detector = LanguageDetector()

        assert len(detector.LANGUAGE_PATTERNS) == 20

    def test_all_language_patterns_have_patterns_list(self) -> None:
        """모든 언어가 패턴 리스트를 가짐"""
        detector = LanguageDetector()

        for language, patterns in detector.LANGUAGE_PATTERNS.items():
            assert isinstance(patterns, list)
            assert len(patterns) > 0

    def test_detect_ruby_by_gemfile(self, tmp_path: Path) -> None:
        """Gemfile로 Ruby 프로젝트 감지"""
        (tmp_path / "Gemfile").write_text("source 'https://rubygems.org'")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result == "ruby"

    def test_detect_php_by_composer_json(self, tmp_path: Path) -> None:
        """composer.json으로 PHP 프로젝트 감지"""
        (tmp_path / "composer.json").write_text("{}")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result == "php"

    def test_detect_elixir_by_mix_exs(self, tmp_path: Path) -> None:
        """mix.exs로 Elixir 프로젝트 감지"""
        (tmp_path / "mix.exs").write_text("defmodule Mix do end")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result == "elixir"

    def test_detect_scala_by_build_sbt(self, tmp_path: Path) -> None:
        """build.sbt로 Scala 프로젝트 감지"""
        (tmp_path / "build.sbt").write_text("name := \"test\"")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result == "scala"

    def test_detect_clojure_by_project_clj(self, tmp_path: Path) -> None:
        """project.clj로 Clojure 프로젝트 감지"""
        (tmp_path / "project.clj").write_text("(defproject test \"0.1.0\")")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result == "clojure"

    def test_detect_haskell_by_cabal(self, tmp_path: Path) -> None:
        """*.cabal로 Haskell 프로젝트 감지"""
        (tmp_path / "test.cabal").write_text("name: test")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result == "haskell"

    def test_detect_cpp_by_cmake(self, tmp_path: Path) -> None:
        """CMakeLists.txt로 C++ 프로젝트 감지"""
        (tmp_path / "CMakeLists.txt").write_text("cmake_minimum_required(VERSION 3.0)")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result == "cpp"

    def test_detect_c_by_makefile(self, tmp_path: Path) -> None:
        """Makefile로 C 프로젝트 감지"""
        (tmp_path / "Makefile").write_text("all:\n\tgcc main.c")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result == "c"

    def test_detect_csharp_by_csproj(self, tmp_path: Path) -> None:
        """*.csproj로 C# 프로젝트 감지"""
        (tmp_path / "test.csproj").write_text("<Project></Project>")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result == "csharp"

    def test_detect_shell_by_sh_files(self, tmp_path: Path) -> None:
        """*.sh 파일로 Shell 프로젝트 감지"""
        (tmp_path / "script.sh").write_text("#!/bin/bash")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result == "shell"

    def test_detect_lua_by_lua_files(self, tmp_path: Path) -> None:
        """*.lua 파일로 Lua 프로젝트 감지"""
        (tmp_path / "init.lua").write_text("print('hello')")
        detector = LanguageDetector()

        result = detector.detect(tmp_path)

        assert result == "lua"
