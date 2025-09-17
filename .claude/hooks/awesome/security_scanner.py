#!/usr/bin/env python3
"""
Security Scanner Hook for Claude Code
Scans code for security vulnerabilities and secrets after modifications
"""

import os
import sys
import subprocess
import re
from pathlib import Path

def run_command(cmd, timeout=10):
    """Run command with timeout and return result"""
    try:
        result = subprocess.run(
            cmd,
            shell=True,
            capture_output=True,
            text=True,
            timeout=timeout
        )
        return result.returncode == 0, result.stdout, result.stderr
    except subprocess.TimeoutExpired:
        return False, "", "Timeout"
    except Exception:
        return False, "", "Error"

def check_command_available(cmd):
    """Check if a command is available in the system"""
    try:
        subprocess.run(f"command -v {cmd}", shell=True, capture_output=True, check=True)
        return True
    except:
        return False

def scan_with_semgrep(file_path):
    """Scan file using Semgrep"""
    if not check_command_available('semgrep'):
        return True, "Semgrep not available"

    print("🔍 Running Semgrep security scan...")
    success, stdout, stderr = run_command(f'semgrep --config=auto "{file_path}"', timeout=30)

    if not success and "found" in stdout.lower():
        print(f"⚠️ Semgrep found potential issues:")
        print(stdout[:500])
        return False, stdout
    elif success:
        print("✅ Semgrep: No issues found")

    return True, ""

def scan_with_bandit(file_path):
    """Scan Python files using Bandit"""
    if not file_path.endswith('.py') or not check_command_available('bandit'):
        return True, "Bandit not available or not Python file"

    print("🐍 Running Bandit security scan...")
    success, stdout, stderr = run_command(f'bandit "{file_path}"', timeout=15)

    if not success and ("high" in stdout.lower() or "medium" in stdout.lower()):
        print(f"⚠️ Bandit found security issues:")
        print(stdout[:500])
        return False, stdout

    if success:
        print("✅ Bandit: No issues found")

    return True, ""

def scan_with_gitleaks(file_path):
    """Scan for secrets using GitLeaks"""
    if not check_command_available('gitleaks'):
        return True, "GitLeaks not available"

    print("🔐 Running GitLeaks secret scan...")
    success, stdout, stderr = run_command(f'gitleaks detect --source="{file_path}" --no-git', timeout=15)

    if not success and "leak" in stdout.lower():
        print(f"⚠️ GitLeaks found potential secrets:")
        print(stdout[:500])
        return False, stdout

    if success:
        print("✅ GitLeaks: No secrets found")

    return True, ""

def scan_hardcoded_secrets(file_path):
    """Scan for hardcoded secrets using regex patterns"""
    try:
        with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
            content = f.read()

        # Common secret patterns
        secret_patterns = [
            r'(password|pwd|secret|key|token|api_key|access_key)\s*[=:]\s*["\'][^"\'\\n]{8,}["\']',
            r'(aws_access_key_id|aws_secret_access_key)\s*[=:]\s*["\'][^"\'\\n]{16,}["\']',
            r'(github_token|gh_token)\s*[=:]\s*["\'][^"\'\\n]{20,}["\']',
            r'(database_url|db_url)\s*[=:]\s*["\'][^"\'\\n]{10,}["\']',
            r'["\'][A-Za-z0-9]{32,}["\']',  # Generic long strings
        ]

        issues_found = []
        lines = content.split('\n')

        for i, line in enumerate(lines, 1):
            for pattern in secret_patterns:
                if re.search(pattern, line, re.IGNORECASE):
                    # Skip common false positives
                    if any(fp in line.lower() for fp in [
                        'example', 'placeholder', 'your_', 'xxx', '***',
                        'dummy', 'test', 'mock', 'fake', 'sample'
                    ]):
                        continue

                    issues_found.append(f"Line {i}: {line.strip()[:80]}...")

        if issues_found:
            print("🚨 Potential hardcoded secrets detected:")
            for issue in issues_found[:3]:  # Show first 3
                print(f"  {issue}")
            if len(issues_found) > 3:
                print(f"  ... and {len(issues_found) - 3} more")
            return False, "Hardcoded secrets found"

        print("✅ Secret scan: No hardcoded secrets found")
        return True, ""

    except Exception as e:
        print(f"⚠️ Error scanning for secrets: {e}", file=sys.stderr)
        return True, ""

def scan_dangerous_functions(file_path):
    """Scan for dangerous functions and practices"""
    dangerous_patterns = {
        '.py': [
            r'eval\s*\(',
            r'exec\s*\(',
            r'__import__\s*\(',
            r'subprocess\.call\(',
            r'os\.system\(',
        ],
        '.js': [
            r'eval\s*\(',
            r'innerHTML\s*=',
            r'document\.write\s*\(',
            r'setTimeout\s*\(\s*["\']',
        ],
        '.sh': [
            r'\$\([^)]*\)',
            r'`[^`]*`',
        ]
    }

    try:
        ext = Path(file_path).suffix
        if ext not in dangerous_patterns:
            return True, ""

        with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
            content = f.read()

        issues_found = []
        lines = content.split('\n')

        for pattern in dangerous_patterns[ext]:
            for i, line in enumerate(lines, 1):
                if re.search(pattern, line, re.IGNORECASE):
                    issues_found.append(f"Line {i}: Dangerous function - {line.strip()[:60]}...")

        if issues_found:
            print("⚡ Potentially dangerous functions detected:")
            for issue in issues_found[:3]:
                print(f"  {issue}")
            return False, "Dangerous functions found"

        return True, ""

    except Exception:
        return True, ""

def main():
    file_path = os.environ.get('CLAUDE_TOOL_FILE_PATH')

    if not file_path or not os.path.exists(file_path):
        return 0

    print(f"🛡️ Security scanning: {Path(file_path).name}")

    all_passed = True
    issues = []

    # Run different scanners
    scanners = [
        ("Hardcoded Secrets", scan_hardcoded_secrets),
        ("Dangerous Functions", scan_dangerous_functions),
        ("Semgrep", scan_with_semgrep),
        ("Bandit", scan_with_bandit),
        ("GitLeaks", scan_with_gitleaks),
    ]

    for scanner_name, scanner_func in scanners:
        try:
            passed, message = scanner_func(file_path)
            if not passed:
                all_passed = False
                issues.append(f"{scanner_name}: {message}")
        except Exception as e:
            print(f"⚠️ Error in {scanner_name}: {e}", file=sys.stderr)

    if not all_passed:
        print(f"\n🚨 Security issues found in {Path(file_path).name}")
        print("Please review and fix the identified issues.")
        # Don't fail the workflow, just warn
        return 0
    else:
        print(f"✅ Security scan completed: No issues found")

    return 0

if __name__ == "__main__":
    sys.exit(main())