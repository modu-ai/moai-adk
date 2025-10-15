# @CODE:CLI-001 | SPEC: SPEC-CLI-001.md | TEST: tests/unit/test_doctor.py
"""moai doctor 명령어

시스템 진단 명령어:
- Python 버전 확인
- Git 설치 확인
- 프로젝트 구조 검증
- 언어별 도구 체인 검증
"""

import json
from pathlib import Path

import click
import questionary
from rich.console import Console
from rich.table import Table

from moai_adk.core.project.checker import SystemChecker, check_environment
from moai_adk.core.project.detector import detect_project_language

console = Console()


@click.command()
@click.option("--verbose", "-v", is_flag=True, help="Show detailed tool versions and language detection")
@click.option("--fix", is_flag=True, help="Suggest fixes for missing tools")
@click.option("--export", type=click.Path(), help="Export diagnostics to JSON file")
@click.option("--check", type=str, help="Check specific tool only")
def doctor(verbose: bool, fix: bool, export: str | None, check: str | None) -> None:
    """Check system requirements and project health

    Verifies:
    - Python version (>= 3.13)
    - Git installation
    - Project structure (.moai directory)
    - Language-specific tool chains (20+ languages)
    """
    try:
        console.print("[cyan]Running system diagnostics...[/cyan]\n")

        # 기본 환경 체크
        results = check_environment()
        diagnostics_data: dict = {"basic_checks": results}

        # Verbose 모드: 언어별 도구 검증
        if verbose or fix:
            language = detect_project_language()
            diagnostics_data["detected_language"] = language

            if verbose:
                console.print(f"[dim]Detected language: {language or 'Unknown'}[/dim]\n")

            if language:
                checker = SystemChecker()
                language_tools = checker.check_language_tools(language)
                diagnostics_data["language_tools"] = language_tools

                if verbose:
                    _display_language_tools(language, language_tools, checker)

        # Specific tool check
        if check:
            _check_specific_tool(check)
            return

        # 기본 결과 테이블 생성
        table = Table(show_header=True, header_style="bold magenta")
        table.add_column("Check", style="dim", width=40)
        table.add_column("Status", justify="center")

        for check_name, status in results.items():
            icon = "✓" if status else "✗"
            color = "green" if status else "red"
            table.add_row(check_name, f"[{color}]{icon}[/{color}]")

        console.print(table)

        # Fix 모드: 누락된 도구 설치 제안
        if fix and "language_tools" in diagnostics_data:
            _suggest_fixes(diagnostics_data["language_tools"], diagnostics_data.get("detected_language"))

        # Export 모드: JSON 파일로 저장
        if export:
            _export_diagnostics(export, diagnostics_data)

        # 전체 결과 요약
        all_passed = all(results.values())
        if all_passed:
            console.print("\n[green]✓ All checks passed[/green]")
        else:
            console.print("\n[yellow]⚠ Some checks failed[/yellow]")
            console.print("[dim]Run [cyan]moai doctor --verbose[/cyan] for detailed diagnostics[/dim]")

    except Exception as e:
        console.print(f"[red]✗ Diagnostic failed: {e}[/red]")
        raise


def _display_language_tools(language: str, tools: dict[str, bool], checker: SystemChecker) -> None:
    """언어별 도구 테이블 표시 (헬퍼 함수)"""
    table = Table(show_header=True, header_style="bold cyan", title=f"{language.title()} Tools")
    table.add_column("Tool", style="dim")
    table.add_column("Status", justify="center")
    table.add_column("Version", style="blue")

    for tool, available in tools.items():
        icon = "✓" if available else "✗"
        color = "green" if available else "red"
        version = checker.get_tool_version(tool) if available else "not installed"

        table.add_row(tool, f"[{color}]{icon}[/{color}]", version or "")

    console.print(table)
    console.print()


def _check_specific_tool(tool: str) -> None:
    """특정 도구만 검증 (헬퍼 함수)"""
    checker = SystemChecker()
    available = checker._is_tool_available(tool)
    version = checker.get_tool_version(tool) if available else None

    if available:
        console.print(f"[green]✓ {tool} is installed[/green]")
        if version:
            console.print(f"  Version: {version}")
    else:
        console.print(f"[red]✗ {tool} is not installed[/red]")


def _suggest_fixes(tools: dict[str, bool], language: str | None) -> None:
    """누락된 도구 설치 제안 (헬퍼 함수)"""
    missing_tools = [tool for tool, available in tools.items() if not available]

    if not missing_tools:
        console.print("\n[green]✓ All tools are installed[/green]")
        return

    console.print(f"\n[yellow]⚠ Missing {len(missing_tools)} tool(s)[/yellow]")

    for tool in missing_tools:
        install_cmd = _get_install_command(tool, language)
        console.print(f"  [red]✗[/red] {tool}")
        if install_cmd:
            console.print(f"    Install: [cyan]{install_cmd}[/cyan]")


def _get_install_command(tool: str, language: str | None) -> str:
    """도구별 설치 명령어 반환 (헬퍼 함수)"""
    # Common tools
    install_commands = {
        "pytest": "pip install pytest",
        "mypy": "pip install mypy",
        "ruff": "pip install ruff",
        "vitest": "npm install -D vitest",
        "biome": "npm install -D @biomejs/biome",
        "eslint": "npm install -D eslint",
        "jest": "npm install -D jest",
    }

    return install_commands.get(tool, f"# Install {tool} for {language}")


def _export_diagnostics(export_path: str, data: dict) -> None:
    """진단 결과를 JSON 파일로 저장 (헬퍼 함수)"""
    try:
        output = Path(export_path)
        output.write_text(json.dumps(data, indent=2))
        console.print(f"\n[green]✓ Diagnostics exported to {export_path}[/green]")
    except Exception as e:
        console.print(f"\n[red]✗ Failed to export diagnostics: {e}[/red]")
