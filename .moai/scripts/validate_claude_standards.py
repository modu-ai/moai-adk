#!/usr/bin/env python3
"""
Claude Code 표준 검증 도구
MoAI-ADK cc-manager를 위한 표준 준수 검증 스크립트
"""

import sys
import os
import json
import yaml
from pathlib import Path
from typing import Dict, List, Any, Tuple


def validate_yaml_frontmatter(content: str) -> Tuple[bool, Dict[str, Any], str]:
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
        return False, {}, f"YAML parsing error: {str(e)}"


def check_required_fields(yaml_data: Dict[str, Any], required_fields: List[str]) -> List[str]:
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


def validate_command_structure(file_path: Path) -> Tuple[bool, List[str]]:
    """
    커맨드 파일 구조 검증

    Returns:
        (is_valid, error_messages)
    """
    errors = []

    try:
        content = file_path.read_text(encoding='utf-8')
    except Exception as e:
        return False, [f"Failed to read file: {str(e)}"]

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


def validate_agent_structure(file_path: Path) -> Tuple[bool, List[str]]:
    """
    에이전트 파일 구조 검증

    Returns:
        (is_valid, error_messages)
    """
    errors = []

    try:
        content = file_path.read_text(encoding='utf-8')
    except Exception as e:
        return False, [f"Failed to read file: {str(e)}"]

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


def main():
    """메인 실행 함수"""
    if len(sys.argv) < 2:
        print("Usage: python validate_claude_standards.py <path>")
        print("  <path> can be a file or directory")
        sys.exit(1)

    path = Path(sys.argv[1])

    if not path.exists():
        print(f"Error: Path {path} does not exist")
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
            print(f"Skipping {relative_path} (not in commands or agents directory)")
            total_files -= 1
            continue

        if is_valid:
            valid_files += 1
            print(f"✅ {file_type}: {relative_path}")
        else:
            print(f"❌ {file_type}: {relative_path}")
            for error in errors:
                print(f"   - {error}")
            errors_found.extend([f"{relative_path}: {error}" for error in errors])

    print(f"\n📊 Validation Summary:")
    print(f"   Total files checked: {total_files}")
    print(f"   Valid files: {valid_files}")
    print(f"   Files with errors: {total_files - valid_files}")

    if errors_found:
        print(f"\n🚨 Errors found:")
        for error in errors_found:
            print(f"   - {error}")
        sys.exit(1)
    else:
        print(f"\n🎉 All files pass validation!")
        sys.exit(0)


if __name__ == '__main__':
    main()