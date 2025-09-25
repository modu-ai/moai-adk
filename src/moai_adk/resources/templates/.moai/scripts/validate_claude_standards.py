#!/usr/bin/env python3
"""
Claude Code 표준 검증 도구
MoAI-ADK cc-manager를 위한 표준 준수 검증 스크립트
"""

import sys
from pathlib import Path
from typing import Any

import click
import yaml


def validate_yaml_frontmatter(content: str) -> tuple[bool, dict[str, Any], str]:
    """
    YAML frontmatter 유효성 검사

    Returns:
        (is_valid, yaml_data, error_message)
    """
    if not content.startswith('---\n'):
        return False, {}, "YAML frontmatter missing"

    # Extract YAML frontmatter
    parts = content.split('---\n')
    if len(parts) < 3:
        return False, {}, "Invalid YAML frontmatter structure"

    yaml_content = parts[1]

    try:
        yaml_data = yaml.safe_load(yaml_content)
        if not isinstance(yaml_data, dict):
            return False, {}, "YAML frontmatter must be a dictionary"
        return True, yaml_data, ""
    except yaml.YAMLError as e:
        return False, {}, f"YAML parsing error: {e!s}"


def check_required_fields(yaml_data: dict[str, Any], required_fields: list[str]) -> list[str]:
    """
    필수 필드 존재 확인

    Returns:
        List of missing field names
    """
    missing_fields = []
    for field in required_fields:
        if field not in yaml_data:
            missing_fields.append(field)
    return missing_fields


def validate_command_structure(file_path: Path) -> tuple[bool, list[str]]:
    """
    커맨드 파일 구조 검증

    Returns:
        (is_valid, error_messages)
    """
    errors = []

    try:
        content = file_path.read_text(encoding='utf-8')
    except Exception as e:
        return False, [f"Failed to read file: {e!s}"]

    # Validate YAML frontmatter
    is_valid, yaml_data, error_msg = validate_yaml_frontmatter(content)
    if not is_valid:
        errors.append(error_msg)
        return False, errors

    # Check required fields for commands
    required_fields = ['name', 'description', 'argument-hint', 'allowed-tools', 'model']
    missing_fields = check_required_fields(yaml_data, required_fields)

    for field in missing_fields:
        errors.append(f"Missing required field: {field}")

    # Validate specific field formats
    if 'name' in yaml_data and not isinstance(yaml_data['name'], str):
        errors.append("'name' field must be a string")

    if 'description' in yaml_data and not isinstance(yaml_data['description'], str):
        errors.append("'description' field must be a string")

    if 'argument-hint' in yaml_data:
        if not isinstance(yaml_data['argument-hint'], (str, list)):
            errors.append("'argument-hint' field must be a string or list")

    if 'allowed-tools' in yaml_data:
        if not isinstance(yaml_data['allowed-tools'], (str, list)):
            errors.append("'allowed-tools' field must be a string or list")

    if 'model' in yaml_data and not isinstance(yaml_data['model'], str):
        errors.append("'model' field must be a string")

    return len(errors) == 0, errors


def validate_agent_structure(file_path: Path) -> tuple[bool, list[str]]:
    """
    에이전트 파일 구조 검증

    Returns:
        (is_valid, error_messages)
    """
    errors = []

    try:
        content = file_path.read_text(encoding='utf-8')
    except Exception as e:
        return False, [f"Failed to read file: {e!s}"]

    # Validate YAML frontmatter
    is_valid, yaml_data, error_msg = validate_yaml_frontmatter(content)
    if not is_valid:
        errors.append(error_msg)
        return False, errors

    # Check required fields for agents
    required_fields = ['name', 'description', 'tools', 'model']
    missing_fields = check_required_fields(yaml_data, required_fields)

    for field in missing_fields:
        errors.append(f"Missing required field: {field}")

    # Validate specific field formats
    if 'name' in yaml_data and not isinstance(yaml_data['name'], str):
        errors.append("'name' field must be a string")

    if 'description' in yaml_data:
        if not isinstance(yaml_data['description'], str):
            errors.append("'description' field must be a string")
        elif 'Use PROACTIVELY for' not in yaml_data['description']:
            errors.append("'description' field must contain 'Use PROACTIVELY for' pattern")

    if 'tools' in yaml_data:
        if not isinstance(yaml_data['tools'], (str, list)):
            errors.append("'tools' field must be a string or list")

    if 'model' in yaml_data and not isinstance(yaml_data['model'], str):
        errors.append("'model' field must be a string")

    return len(errors) == 0, errors


def validate_proactive_pattern(description: str) -> bool:
    """
    Check if description contains 'Use PROACTIVELY for' pattern

    Returns:
        True if pattern exists, False otherwise
    """
    return 'Use PROACTIVELY for' in description


def generate_violation_report(errors_found: list[str]) -> str:
    """
    표준 위반 사항에 대한 종합 보고서 생성

    Args:
        errors_found: 발견된 오류 목록

    Returns:
        포맷된 보고서 문자열
    """
    if not errors_found:
        return "🎉 모든 파일이 Claude Code 표준을 준수합니다!"

    report = ["🚨 Claude Code 표준 위반 사항 보고서", "=" * 50]

    # 파일별로 오류 그룹화
    file_errors = {}
    for error in errors_found:
        if ": " in error:
            file_name, error_msg = error.split(": ", 1)
            if file_name not in file_errors:
                file_errors[file_name] = []
            file_errors[file_name].append(error_msg)
        else:
            if "기타" not in file_errors:
                file_errors["기타"] = []
            file_errors["기타"].append(error)

    for file_name, file_error_list in file_errors.items():
        report.append(f"\n📁 파일: {file_name}")
        report.append("-" * 30)
        for i, error in enumerate(file_error_list, 1):
            report.append(f"  {i}. {error}")

    report.append("\n📊 요약:")
    report.append(f"  - 위반 파일 수: {len(file_errors)}")
    report.append(f"  - 총 위반 사항: {len(errors_found)}")

    return "\n".join(report)


def suggest_fixes(errors_found: list[str]) -> list[str]:
    """
    발견된 오류에 대한 수정 제안 생성

    Args:
        errors_found: 발견된 오류 목록

    Returns:
        수정 제안 목록
    """
    suggestions = []

    for error in errors_found:
        if "YAML frontmatter missing" in error:
            suggestions.append(
                "✅ YAML frontmatter 추가:\n"
                "   파일 시작에 다음 구조를 추가하세요:\n"
                "   ---\n"
                "   name: your-file-name\n"
                "   description: Clear description\n"
                "   ---"
            )
        elif "Missing required field" in error:
            field_match = error.split("'")
            if len(field_match) >= 2:
                field_name = field_match[1]
                suggestions.append(
                    f"✅ 필수 필드 '{field_name}' 추가:\n"
                    f"   YAML frontmatter에 '{field_name}: <값>' 추가"
                )
        elif "Use PROACTIVELY for" in error:
            suggestions.append(
                "✅ 프로액티브 패턴 수정:\n"
                "   description을 다음과 같이 시작하도록 수정:\n"
                "   'Use PROACTIVELY for [구체적인 트리거 조건]'"
            )
        elif "argument-hint" in error:
            suggestions.append(
                "✅ argument-hint 형식 수정:\n"
                "   문자열 또는 배열 형태로 수정:\n"
                "   argument-hint: '[param1] [param2]' 또는\n"
                "   argument-hint: ['param1', 'param2']"
            )
        elif "tools" in error or "allowed-tools" in error:
            suggestions.append(
                "✅ 도구 권한 수정:\n"
                "   최소 권한 원칙에 따라 필요한 도구만 나열:\n"
                "   tools: 'Read, Write, Edit' 또는\n"
                "   tools: ['Read', 'Write', 'Edit']"
            )

    # 중복 제거
    unique_suggestions = list(set(suggestions))

    if not unique_suggestions:
        unique_suggestions.append("❓ 구체적인 수정 제안을 생성할 수 없습니다. cc-manager 문서를 참조하세요.")

    return unique_suggestions


def main():
    """메인 실행 함수"""
    if len(sys.argv) < 2:
        click.echo("Usage: python validate_claude_standards.py <path>")
        click.echo("  <path> can be a file or directory")
        sys.exit(1)

    path = Path(sys.argv[1])

    if not path.exists():
        click.echo(f"Error: Path {path} does not exist")
        sys.exit(1)

    total_files = 0
    valid_files = 0
    errors_found = []

    if path.is_file():
        files_to_check = [path]
    else:
        # Check all .md files in commands and agents directories
        files_to_check = []
        commands_dir = path / '.claude' / 'commands'
        agents_dir = path / '.claude' / 'agents'

        if commands_dir.exists():
            files_to_check.extend(commands_dir.rglob('*.md'))

        if agents_dir.exists():
            files_to_check.extend(agents_dir.rglob('*.md'))

    for file_path in files_to_check:
        total_files += 1
        relative_path = file_path.relative_to(path) if path.is_dir() else file_path.name

        # Determine if it's a command or agent file
        if '.claude/commands' in str(file_path) or '/commands/' in str(file_path):
            is_valid, errors = validate_command_structure(file_path)
            file_type = "Command"
        elif '.claude/agents' in str(file_path) or '/agents/' in str(file_path):
            is_valid, errors = validate_agent_structure(file_path)
            file_type = "Agent"
        else:
            click.echo(f"Skipping {relative_path} (not in commands or agents directory)")
            total_files -= 1
            continue

        if is_valid:
            valid_files += 1
            click.echo(f"✅ {file_type}: {relative_path}")
        else:
            click.echo(f"❌ {file_type}: {relative_path}")
            for error in errors:
                click.echo(f"   - {error}")
            errors_found.extend([f"{relative_path}: {error}" for error in errors])

    click.echo("\n📊 Validation Summary:")
    click.echo(f"   Total files checked: {total_files}")
    click.echo(f"   Valid files: {valid_files}")
    click.echo(f"   Files with errors: {total_files - valid_files}")

    if errors_found:
        click.echo("\n🚨 Errors found:")
        for error in errors_found:
            click.echo(f"   - {error}")
        sys.exit(1)
    else:
        click.echo("\n🎉 All files pass validation!")
        sys.exit(0)


if __name__ == '__main__':
    main()
