#!/usr/bin/env python3
"""
Test Runner Hook for Claude Code
Automatically runs relevant tests after code changes
"""

import os
import sys
import subprocess
from pathlib import Path

def detect_project_type(file_path):
    """Detect project type based on files in the directory"""
    path = Path(file_path)
    project_root = path.parent

    # Walk up to find project root
    while project_root != project_root.parent:
        if any((project_root / marker).exists() for marker in
               ['package.json', 'requirements.txt', 'Gemfile', 'go.mod', 'Cargo.toml']):
            break
        project_root = project_root.parent

    return project_root

def run_command(cmd, cwd=None, timeout=30):
    """Run command with timeout and error handling"""
    try:
        result = subprocess.run(
            cmd,
            shell=True,
            capture_output=True,
            text=True,
            timeout=timeout,
            cwd=cwd
        )
        return result.returncode == 0, result.stdout, result.stderr
    except subprocess.TimeoutExpired:
        print("⏰ Test timeout - tests taking too long", file=sys.stderr)
        return False, "", "Timeout"
    except Exception as e:
        print(f"⚠️ Test execution error: {e}", file=sys.stderr)
        return False, "", str(e)

def run_tests_for_file(file_path):
    """Run appropriate tests based on file type and project structure"""
    path = Path(file_path)
    project_root = detect_project_type(file_path)
    ext = path.suffix.lower()

    print(f"🧪 Running tests for {path.name}...")

    # JavaScript/TypeScript projects
    if ext in ['.js', '.ts', '.jsx', '.tsx']:
        if (project_root / 'package.json').exists():
            # Try different test commands
            test_commands = [
                'npm test',
                'yarn test',
                'npm run test:unit',
                'yarn test:unit',
                f'npm test -- --testPathPattern={path.name}',
                f'jest {path.name}'
            ]

            for cmd in test_commands:
                print(f"🔍 Trying: {cmd}")
                success, stdout, stderr = run_command(cmd, cwd=project_root, timeout=60)
                if success:
                    print(f"✅ Tests passed with: {cmd}")
                    return True
                elif 'command not found' not in stderr.lower():
                    print(f"❌ Tests failed: {stderr[:200]}...")
                    return False

    # Python projects
    elif ext == '.py':
        test_indicators = ['pytest.ini', 'setup.cfg', 'pyproject.toml', 'tox.ini']
        if any((project_root / indicator).exists() for indicator in test_indicators):
            test_commands = [
                f'pytest {file_path} -v',
                f'python -m pytest {file_path} -v',
                f'pytest {path.parent} -k {path.stem}',
                'pytest',
                'python -m pytest'
            ]

            for cmd in test_commands:
                print(f"🔍 Trying: {cmd}")
                success, stdout, stderr = run_command(cmd, cwd=project_root, timeout=60)
                if success:
                    print(f"✅ Tests passed with: {cmd}")
                    return True
                elif 'command not found' not in stderr.lower():
                    print(f"❌ Tests failed: {stderr[:200]}...")
                    return False

    # Ruby projects
    elif ext == '.rb':
        if (project_root / 'Gemfile').exists():
            test_commands = [
                f'bundle exec rspec {file_path}',
                f'rspec {file_path}',
                'bundle exec rspec',
                'rspec'
            ]

            for cmd in test_commands:
                print(f"🔍 Trying: {cmd}")
                success, stdout, stderr = run_command(cmd, cwd=project_root, timeout=60)
                if success:
                    print(f"✅ Tests passed with: {cmd}")
                    return True
                elif 'command not found' not in stderr.lower():
                    print(f"❌ Tests failed: {stderr[:200]}...")
                    return False

    # Go projects
    elif ext == '.go':
        if (project_root / 'go.mod').exists():
            test_commands = [
                f'go test {path.parent}',
                'go test ./...',
                'go test .'
            ]

            for cmd in test_commands:
                print(f"🔍 Trying: {cmd}")
                success, stdout, stderr = run_command(cmd, cwd=project_root, timeout=60)
                if success:
                    print(f"✅ Tests passed with: {cmd}")
                    return True
                else:
                    print(f"❌ Tests failed: {stderr[:200]}...")
                    return False

    print(f"ℹ️ No test framework detected for {ext} files")
    return True  # Don't fail if no tests found

def main():
    file_path = os.environ.get('CLAUDE_TOOL_FILE_PATH')

    if not file_path:
        return 0

    if not os.path.exists(file_path):
        return 0

    # Skip test files themselves to avoid infinite loops
    if any(pattern in file_path.lower() for pattern in ['.test.', '.spec.', '_test.', 'test_']):
        print("ℹ️ Skipping test execution for test files")
        return 0

    try:
        success = run_tests_for_file(file_path)
        return 0 if success else 1
    except Exception as e:
        print(f"⚠️ Test runner error: {e}", file=sys.stderr)
        return 0  # Don't fail the workflow

if __name__ == "__main__":
    sys.exit(main())