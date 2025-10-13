# @TEST:HOOKS-002 | SPEC: SPEC-HOOKS-002.md
"""moai_hooks.py 테스트 스위트

Self-contained Python hook script 기능을 검증합니다.
"""

import json
import sys
from pathlib import Path
from typing import Any
from unittest.mock import MagicMock, patch

import pytest


# ============================================================================
# Test Class: Language Detection (20 tests)
# ============================================================================


class TestLanguageDetection:
    """언어 감지 테스트 (20개 언어)"""

    def test_detect_python_by_pyproject(self, tmp_path: Path) -> None:
        """pyproject.toml로 Python 프로젝트 감지"""
        (tmp_path / "pyproject.toml").write_text("[tool.poetry]")

        from moai_adk.templates.claude.hooks.moai_hooks import detect_language

        result = detect_language(str(tmp_path))

        assert result == "python"

    def test_detect_typescript_by_tsconfig(self, tmp_path: Path) -> None:
        """tsconfig.json으로 TypeScript 프로젝트 감지"""
        (tmp_path / "tsconfig.json").write_text("{}")

        from moai_adk.templates.claude.hooks.moai_hooks import detect_language

        result = detect_language(str(tmp_path))

        assert result == "typescript"

    def test_detect_javascript_by_package_json(self, tmp_path: Path) -> None:
        """package.json으로 JavaScript 프로젝트 감지"""
        (tmp_path / "package.json").write_text("{}")

        from moai_adk.templates.claude.hooks.moai_hooks import detect_language

        result = detect_language(str(tmp_path))

        assert result in ["javascript", "typescript"]

    def test_detect_java_by_pom_xml(self, tmp_path: Path) -> None:
        """pom.xml로 Java 프로젝트 감지"""
        (tmp_path / "pom.xml").write_text("<project></project>")

        from moai_adk.templates.claude.hooks.moai_hooks import detect_language

        result = detect_language(str(tmp_path))

        assert result == "java"

    def test_detect_go_by_go_mod(self, tmp_path: Path) -> None:
        """go.mod로 Go 프로젝트 감지"""
        (tmp_path / "go.mod").write_text("module example")

        from moai_adk.templates.claude.hooks.moai_hooks import detect_language

        result = detect_language(str(tmp_path))

        assert result == "go"

    def test_detect_rust_by_cargo_toml(self, tmp_path: Path) -> None:
        """Cargo.toml로 Rust 프로젝트 감지"""
        (tmp_path / "Cargo.toml").write_text("[package]")

        from moai_adk.templates.claude.hooks.moai_hooks import detect_language

        result = detect_language(str(tmp_path))

        assert result == "rust"

    def test_detect_dart_by_pubspec_yaml(self, tmp_path: Path) -> None:
        """pubspec.yaml로 Dart 프로젝트 감지"""
        (tmp_path / "pubspec.yaml").write_text("name: test")

        from moai_adk.templates.claude.hooks.moai_hooks import detect_language

        result = detect_language(str(tmp_path))

        assert result == "dart"

    def test_detect_swift_by_package_swift(self, tmp_path: Path) -> None:
        """Package.swift로 Swift 프로젝트 감지"""
        (tmp_path / "Package.swift").write_text("// swift-tools-version")

        from moai_adk.templates.claude.hooks.moai_hooks import detect_language

        result = detect_language(str(tmp_path))

        assert result == "swift"

    def test_detect_kotlin_by_gradle_kts(self, tmp_path: Path) -> None:
        """build.gradle.kts로 Kotlin 프로젝트 감지"""
        (tmp_path / "build.gradle.kts").write_text("plugins {}")

        from moai_adk.templates.claude.hooks.moai_hooks import detect_language

        result = detect_language(str(tmp_path))

        assert result == "kotlin"

    def test_detect_csharp_by_csproj(self, tmp_path: Path) -> None:
        """*.csproj로 C# 프로젝트 감지"""
        (tmp_path / "test.csproj").write_text("<Project></Project>")

        from moai_adk.templates.claude.hooks.moai_hooks import detect_language

        result = detect_language(str(tmp_path))

        assert result == "csharp"

    def test_detect_php_by_composer_json(self, tmp_path: Path) -> None:
        """composer.json으로 PHP 프로젝트 감지"""
        (tmp_path / "composer.json").write_text("{}")

        from moai_adk.templates.claude.hooks.moai_hooks import detect_language

        result = detect_language(str(tmp_path))

        assert result == "php"

    def test_detect_ruby_by_gemfile(self, tmp_path: Path) -> None:
        """Gemfile로 Ruby 프로젝트 감지"""
        (tmp_path / "Gemfile").write_text("source 'https://rubygems.org'")

        from moai_adk.templates.claude.hooks.moai_hooks import detect_language

        result = detect_language(str(tmp_path))

        assert result == "ruby"

    def test_detect_elixir_by_mix_exs(self, tmp_path: Path) -> None:
        """mix.exs로 Elixir 프로젝트 감지"""
        (tmp_path / "mix.exs").write_text("defmodule Mix do end")

        from moai_adk.templates.claude.hooks.moai_hooks import detect_language

        result = detect_language(str(tmp_path))

        assert result == "elixir"

    def test_detect_scala_by_build_sbt(self, tmp_path: Path) -> None:
        """build.sbt로 Scala 프로젝트 감지"""
        (tmp_path / "build.sbt").write_text('name := "test"')

        from moai_adk.templates.claude.hooks.moai_hooks import detect_language

        result = detect_language(str(tmp_path))

        assert result == "scala"

    def test_detect_clojure_by_project_clj(self, tmp_path: Path) -> None:
        """project.clj로 Clojure 프로젝트 감지"""
        (tmp_path / "project.clj").write_text('(defproject test "0.1.0")')

        from moai_adk.templates.claude.hooks.moai_hooks import detect_language

        result = detect_language(str(tmp_path))

        assert result == "clojure"

    def test_detect_haskell_by_cabal(self, tmp_path: Path) -> None:
        """*.cabal로 Haskell 프로젝트 감지"""
        (tmp_path / "test.cabal").write_text("name: test")

        from moai_adk.templates.claude.hooks.moai_hooks import detect_language

        result = detect_language(str(tmp_path))

        assert result == "haskell"

    def test_detect_cpp_by_cmake(self, tmp_path: Path) -> None:
        """CMakeLists.txt로 C++ 프로젝트 감지"""
        (tmp_path / "CMakeLists.txt").write_text("cmake_minimum_required(VERSION 3.0)")

        from moai_adk.templates.claude.hooks.moai_hooks import detect_language

        result = detect_language(str(tmp_path))

        assert result == "cpp"

    def test_detect_c_by_makefile(self, tmp_path: Path) -> None:
        """Makefile로 C 프로젝트 감지"""
        (tmp_path / "Makefile").write_text("all:\n\tgcc main.c")

        from moai_adk.templates.claude.hooks.moai_hooks import detect_language

        result = detect_language(str(tmp_path))

        assert result == "c"

    def test_detect_shell_by_sh_files(self, tmp_path: Path) -> None:
        """*.sh 파일로 Shell 프로젝트 감지"""
        (tmp_path / "script.sh").write_text("#!/bin/bash")

        from moai_adk.templates.claude.hooks.moai_hooks import detect_language

        result = detect_language(str(tmp_path))

        assert result == "shell"

    def test_detect_lua_by_lua_files(self, tmp_path: Path) -> None:
        """*.lua 파일로 Lua 프로젝트 감지"""
        (tmp_path / "init.lua").write_text("print('hello')")

        from moai_adk.templates.claude.hooks.moai_hooks import detect_language

        result = detect_language(str(tmp_path))

        assert result == "lua"

    def test_detect_unknown_language(self, tmp_path: Path) -> None:
        """알 수 없는 언어에 대해 Unknown Language 반환"""
        from moai_adk.templates.claude.hooks.moai_hooks import detect_language

        result = detect_language(str(tmp_path))

        assert result == "Unknown Language"


# ============================================================================
# Test Class: Git Information Collection (5 tests)
# ============================================================================


class TestGitInfo:
    """Git 정보 수집 테스트"""

    @patch("subprocess.run")
    def test_get_git_info_success(self, mock_run: MagicMock, tmp_path: Path) -> None:
        """Git 리포지토리 정보 수집 성공"""
        mock_run.side_effect = [
            MagicMock(returncode=0, stdout=""),  # rev-parse --git-dir
            MagicMock(returncode=0, stdout="main\n"),  # branch --show-current
            MagicMock(returncode=0, stdout="abc1234\n"),  # rev-parse HEAD
            MagicMock(returncode=0, stdout="M file.py\n"),  # status --short
        ]

        from moai_adk.templates.claude.hooks.moai_hooks import get_git_info

        result = get_git_info(str(tmp_path))

        assert result["branch"] == "main"
        assert result["commit"] == "abc1234"
        assert result["changes"] == 1

    @patch("subprocess.run")
    def test_get_git_info_no_repo(self, mock_run: MagicMock, tmp_path: Path) -> None:
        """Git 리포지토리가 아닌 경우 빈 딕셔너리 반환"""
        mock_run.return_value = MagicMock(returncode=1, stdout="")

        from moai_adk.templates.claude.hooks.moai_hooks import get_git_info

        result = get_git_info(str(tmp_path))

        assert result == {}

    @patch("subprocess.run")
    def test_get_git_info_timeout(self, mock_run: MagicMock, tmp_path: Path) -> None:
        """Git 명령 타임아웃 시 빈 딕셔너리 반환"""
        import subprocess
        mock_run.side_effect = subprocess.TimeoutExpired("git", 2)

        from moai_adk.templates.claude.hooks.moai_hooks import get_git_info

        result = get_git_info(str(tmp_path))

        assert result == {}

    @patch("subprocess.run")
    def test_get_git_info_with_changes(self, mock_run: MagicMock, tmp_path: Path) -> None:
        """변경된 파일이 있는 경우 changes 카운트"""
        mock_run.side_effect = [
            MagicMock(returncode=0, stdout=""),
            MagicMock(returncode=0, stdout="feature\n"),
            MagicMock(returncode=0, stdout="def5678\n"),
            MagicMock(returncode=0, stdout="M file1.py\nA file2.py\nD file3.py\n"),
        ]

        from moai_adk.templates.claude.hooks.moai_hooks import get_git_info

        result = get_git_info(str(tmp_path))

        assert result["changes"] == 3

    @patch("subprocess.run")
    def test_get_git_info_no_git_command(self, mock_run: MagicMock, tmp_path: Path) -> None:
        """git 명령어가 없는 경우 빈 딕셔너리 반환"""
        mock_run.side_effect = FileNotFoundError()

        from moai_adk.templates.claude.hooks.moai_hooks import get_git_info

        result = get_git_info(str(tmp_path))

        assert result == {}


# ============================================================================
# Test Class: SPEC Count Calculation (4 tests)
# ============================================================================


class TestSpecCount:
    """SPEC 파일 카운트 테스트"""

    def test_count_specs_completed(self, tmp_path: Path) -> None:
        """완료된 SPEC 카운트"""
        specs_dir = tmp_path / ".moai" / "specs"
        specs_dir.mkdir(parents=True)

        # SPEC-001: completed
        spec1_dir = specs_dir / "SPEC-TEST-001"
        spec1_dir.mkdir()
        (spec1_dir / "spec.md").write_text("---\nstatus: completed\n---")

        # SPEC-002: draft
        spec2_dir = specs_dir / "SPEC-TEST-002"
        spec2_dir.mkdir()
        (spec2_dir / "spec.md").write_text("---\nstatus: draft\n---")

        from moai_adk.templates.claude.hooks.moai_hooks import count_specs

        result = count_specs(str(tmp_path))

        assert result["completed"] == 1
        assert result["total"] == 2
        assert result["percentage"] == 50

    def test_count_specs_empty_dir(self, tmp_path: Path) -> None:
        """.moai/specs/ 디렉토리가 비어있는 경우"""
        specs_dir = tmp_path / ".moai" / "specs"
        specs_dir.mkdir(parents=True)

        from moai_adk.templates.claude.hooks.moai_hooks import count_specs

        result = count_specs(str(tmp_path))

        assert result["completed"] == 0
        assert result["total"] == 0
        assert result["percentage"] == 0

    def test_count_specs_no_moai_dir(self, tmp_path: Path) -> None:
        """.moai/ 디렉토리가 없는 경우"""
        from moai_adk.templates.claude.hooks.moai_hooks import count_specs

        result = count_specs(str(tmp_path))

        assert result["completed"] == 0
        assert result["total"] == 0
        assert result["percentage"] == 0

    def test_count_specs_percentage_calculation(self, tmp_path: Path) -> None:
        """SPEC 완료율 계산"""
        specs_dir = tmp_path / ".moai" / "specs"
        specs_dir.mkdir(parents=True)

        # 3개 SPEC, 2개 completed
        for i in range(1, 4):
            spec_dir = specs_dir / f"SPEC-TEST-{i:03d}"
            spec_dir.mkdir()
            status = "completed" if i <= 2 else "draft"
            (spec_dir / "spec.md").write_text(f"---\nstatus: {status}\n---")

        from moai_adk.templates.claude.hooks.moai_hooks import count_specs

        result = count_specs(str(tmp_path))

        assert result["completed"] == 2
        assert result["total"] == 3
        assert result["percentage"] == 66  # int(2/3 * 100)


# ============================================================================
# Test Class: JIT Context Retrieval (6 tests)
# ============================================================================


class TestJITContext:
    """JIT Context Retrieval 테스트"""

    def test_jit_context_alfred_1_spec(self, tmp_path: Path) -> None:
        """/alfred:1-spec 명령어 감지 시 spec-metadata.md 참조"""
        memory_dir = tmp_path / ".moai" / "memory"
        memory_dir.mkdir(parents=True)
        (memory_dir / "spec-metadata.md").write_text("# SPEC Metadata")

        from moai_adk.templates.claude.hooks.moai_hooks import get_jit_context

        result = get_jit_context("/alfred:1-spec 새 기능", str(tmp_path))

        assert ".moai/memory/spec-metadata.md" in result

    def test_jit_context_alfred_2_build(self, tmp_path: Path) -> None:
        """/alfred:2-build 명령어 감지 시 development-guide.md 참조"""
        memory_dir = tmp_path / ".moai" / "memory"
        memory_dir.mkdir(parents=True)
        (memory_dir / "development-guide.md").write_text("# Dev Guide")

        from moai_adk.templates.claude.hooks.moai_hooks import get_jit_context

        result = get_jit_context("/alfred:2-build SPEC-001", str(tmp_path))

        assert ".moai/memory/development-guide.md" in result

    def test_jit_context_test_keyword(self, tmp_path: Path) -> None:
        """test 키워드 감지 시 tests/ 디렉토리 참조"""
        tests_dir = tmp_path / "tests"
        tests_dir.mkdir()

        from moai_adk.templates.claude.hooks.moai_hooks import get_jit_context

        result = get_jit_context("run pytest tests", str(tmp_path))

        assert "tests/" in result

    def test_jit_context_multiple_patterns(self, tmp_path: Path) -> None:
        """여러 패턴 매칭 시 모든 컨텍스트 파일 반환"""
        memory_dir = tmp_path / ".moai" / "memory"
        memory_dir.mkdir(parents=True)
        (memory_dir / "spec-metadata.md").write_text("# SPEC")
        (memory_dir / "development-guide.md").write_text("# Guide")
        tests_dir = tmp_path / "tests"
        tests_dir.mkdir()

        from moai_adk.templates.claude.hooks.moai_hooks import get_jit_context

        result = get_jit_context("/alfred:1-spec and run tests", str(tmp_path))

        assert ".moai/memory/spec-metadata.md" in result
        assert "tests/" in result

    def test_jit_context_file_not_found(self, tmp_path: Path) -> None:
        """파일이 존재하지 않는 경우 빈 리스트 반환"""
        from moai_adk.templates.claude.hooks.moai_hooks import get_jit_context

        result = get_jit_context("/alfred:1-spec", str(tmp_path))

        assert result == []

    def test_jit_context_empty_prompt(self, tmp_path: Path) -> None:
        """빈 프롬프트에 대해 빈 리스트 반환"""
        from moai_adk.templates.claude.hooks.moai_hooks import get_jit_context

        result = get_jit_context("", str(tmp_path))

        assert result == []


# ============================================================================
# Test Class: Hook Handlers (9 tests)
# ============================================================================


class TestHookHandlers:
    """Hook Handler 테스트 (9개 이벤트)"""

    @patch("moai_adk.templates.claude.hooks.moai_hooks.detect_language")
    @patch("moai_adk.templates.claude.hooks.moai_hooks.get_git_info")
    @patch("moai_adk.templates.claude.hooks.moai_hooks.count_specs")
    def test_handle_session_start(
        self,
        mock_count_specs: MagicMock,
        mock_git_info: MagicMock,
        mock_detect_language: MagicMock,
        tmp_path: Path,
    ) -> None:
        """SessionStart 이벤트 핸들러 테스트"""
        mock_detect_language.return_value = "python"
        mock_git_info.return_value = {
            "branch": "main",
            "commit": "abc1234",
            "changes": 2,
        }
        mock_count_specs.return_value = {"completed": 3, "total": 5, "percentage": 60}

        from moai_adk.templates.claude.hooks.moai_hooks import handle_session_start

        result = handle_session_start({"cwd": str(tmp_path)})

        assert result.message is not None
        assert "MoAI-ADK Session Started" in result.message
        assert "python" in result.message
        assert "main" in result.message
        assert "abc1234" in result.message
        assert "3/5" in result.message

    @patch("moai_adk.templates.claude.hooks.moai_hooks.get_jit_context")
    def test_handle_user_prompt_submit(
        self, mock_jit_context: MagicMock, tmp_path: Path
    ) -> None:
        """UserPromptSubmit 이벤트 핸들러 테스트"""
        mock_jit_context.return_value = [".moai/memory/spec-metadata.md"]

        from moai_adk.templates.claude.hooks.moai_hooks import (
            handle_user_prompt_submit,
        )

        result = handle_user_prompt_submit(
            {"cwd": str(tmp_path), "userPrompt": "/alfred:1-spec"}
        )

        assert result.contextFiles == [".moai/memory/spec-metadata.md"]
        assert "Loaded 1 context file(s)" in result.message

    def test_handle_pre_compact(self, tmp_path: Path) -> None:
        """PreCompact 이벤트 핸들러 테스트"""
        from moai_adk.templates.claude.hooks.moai_hooks import handle_pre_compact

        result = handle_pre_compact({"cwd": str(tmp_path)})

        assert result.message is not None
        assert "Tip:" in result.message
        assert len(result.suggestions) > 0

    def test_handle_session_end(self, tmp_path: Path) -> None:
        """SessionEnd 이벤트 핸들러 테스트 (기본 구현)"""
        from moai_adk.templates.claude.hooks.moai_hooks import handle_session_end

        result = handle_session_end({"cwd": str(tmp_path)})

        assert result.message is None or isinstance(result.message, str)

    def test_handle_pre_tool_use(self, tmp_path: Path) -> None:
        """PreToolUse 이벤트 핸들러 테스트 (기본 구현)"""
        from moai_adk.templates.claude.hooks.moai_hooks import handle_pre_tool_use

        result = handle_pre_tool_use(
            {"cwd": str(tmp_path), "toolName": "Read", "toolArgs": {}}
        )

        assert result.blocked is False

    def test_handle_post_tool_use(self, tmp_path: Path) -> None:
        """PostToolUse 이벤트 핸들러 테스트 (기본 구현)"""
        from moai_adk.templates.claude.hooks.moai_hooks import handle_post_tool_use

        result = handle_post_tool_use(
            {"cwd": str(tmp_path), "toolName": "Read", "toolArgs": {}}
        )

        assert result.message is None or isinstance(result.message, str)

    def test_handle_notification(self, tmp_path: Path) -> None:
        """Notification 이벤트 핸들러 테스트 (기본 구현)"""
        from moai_adk.templates.claude.hooks.moai_hooks import handle_notification

        result = handle_notification(
            {"cwd": str(tmp_path), "notificationMessage": "Test"}
        )

        assert result.message is None or isinstance(result.message, str)

    def test_handle_stop(self, tmp_path: Path) -> None:
        """Stop 이벤트 핸들러 테스트 (기본 구현)"""
        from moai_adk.templates.claude.hooks.moai_hooks import handle_stop

        result = handle_stop({"cwd": str(tmp_path)})

        assert result.message is None or isinstance(result.message, str)

    def test_handle_subagent_stop(self, tmp_path: Path) -> None:
        """SubagentStop 이벤트 핸들러 테스트 (기본 구현)"""
        from moai_adk.templates.claude.hooks.moai_hooks import handle_subagent_stop

        result = handle_subagent_stop({"cwd": str(tmp_path)})

        assert result.message is None or isinstance(result.message, str)


# ============================================================================
# Test Class: Main Entry Point & Integration (4 tests)
# ============================================================================


class TestMainIntegration:
    """Main entry point 및 통합 테스트"""

    def test_main_routes_session_start_event(
        self, tmp_path: Path, monkeypatch: Any
    ) -> None:
        """SessionStart 이벤트 라우팅 테스트"""
        import io

        payload = json.dumps({"event": "SessionStart", "cwd": str(tmp_path)})
        stdin = io.StringIO(payload)
        stdout = io.StringIO()

        monkeypatch.setattr("sys.stdin", stdin)
        monkeypatch.setattr("sys.stdout", stdout)
        monkeypatch.setattr("sys.argv", ["moai_hooks.py", "SessionStart"])

        from moai_adk.templates.claude.hooks.moai_hooks import main

        try:
            main()
        except SystemExit as e:
            assert e.code == 0

        output = stdout.getvalue()
        result = json.loads(output)

        assert "message" in result
        assert "MoAI-ADK Session Started" in result["message"]

    def test_main_handles_unknown_event(
        self, tmp_path: Path, monkeypatch: Any
    ) -> None:
        """알 수 없는 이벤트 처리 (no-op)"""
        import io

        payload = json.dumps({"event": "UnknownEvent", "cwd": str(tmp_path)})
        stdin = io.StringIO(payload)
        stdout = io.StringIO()

        monkeypatch.setattr("sys.stdin", stdin)
        monkeypatch.setattr("sys.stdout", stdout)
        monkeypatch.setattr("sys.argv", ["moai_hooks.py", "UnknownEvent"])

        from moai_adk.templates.claude.hooks.moai_hooks import main

        try:
            main()
        except SystemExit as e:
            assert e.code == 0

        output = stdout.getvalue()
        result = json.loads(output)

        # Unknown event returns empty result
        assert result["message"] is None
        assert result["blocked"] is False

    def test_main_handles_json_parse_error(self, monkeypatch: Any) -> None:
        """JSON parsing 오류 처리"""
        import io

        stdin = io.StringIO("{invalid json")
        stderr = io.StringIO()

        monkeypatch.setattr("sys.stdin", stdin)
        monkeypatch.setattr("sys.stderr", stderr)
        monkeypatch.setattr("sys.argv", ["moai_hooks.py", "SessionStart"])

        from moai_adk.templates.claude.hooks.moai_hooks import main

        try:
            main()
        except SystemExit as e:
            assert e.code == 1

        error_output = stderr.getvalue()
        assert "JSON parse error" in error_output

    def test_main_handles_missing_event_arg(self, monkeypatch: Any) -> None:
        """이벤트 인자 누락 처리"""
        import io

        stderr = io.StringIO()

        monkeypatch.setattr("sys.stderr", stderr)
        monkeypatch.setattr("sys.argv", ["moai_hooks.py"])  # No event arg

        from moai_adk.templates.claude.hooks.moai_hooks import main

        try:
            main()
        except SystemExit as e:
            assert e.code == 1

        error_output = stderr.getvalue()
        assert "Usage:" in error_output
