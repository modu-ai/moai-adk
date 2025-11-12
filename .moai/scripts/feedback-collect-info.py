#!/usr/bin/env python3
"""
자동 피드백 정보 수집 스크립트

MoAI-ADK 버전, Python 버전, 환경 정보, Git 상태, 최근 에러 로그 등을 자동으로 수집합니다.
JSON 형식으로 출력되어 /alfred:9-feedback에서 사용됩니다.

Usage:
    python3 .moai/scripts/feedback-collect-info.py

Output:
    JSON format with collected information
"""

import json
import sys
import os
import subprocess
from pathlib import Path
from datetime import datetime

def get_moai_version():
    """MoAI-ADK 버전 조회"""
    try:
        config_path = Path(".moai/config/config.json")
        if config_path.exists():
            import json as json_module
            with open(config_path, 'r', encoding='utf-8') as f:
                config = json_module.load(f)
                return config.get("moai", {}).get("version", "unknown")
    except Exception:
        pass
    return "unknown"

def get_python_version():
    """Python 버전 조회"""
    return f"{sys.version_info.major}.{sys.version_info.minor}.{sys.version_info.micro}"

def get_os_info():
    """OS 정보 조회"""
    try:
        import platform
        return f"{platform.system()} {platform.release()}"
    except Exception:
        return "unknown"

def get_project_mode():
    """프로젝트 모드 (team/personal) 조회"""
    try:
        config_path = Path(".moai/config/config.json")
        if config_path.exists():
            import json as json_module
            with open(config_path, 'r', encoding='utf-8') as f:
                config = json_module.load(f)
                return config.get("project", {}).get("mode", "unknown")
    except Exception:
        pass
    return "unknown"

def get_current_branch():
    """현재 Git 브랜치 조회"""
    try:
        result = subprocess.run(
            ["git", "branch", "--show-current"],
            capture_output=True,
            text=True,
            timeout=2
        )
        return result.stdout.strip() if result.returncode == 0 else "unknown"
    except Exception:
        return "unknown"

def get_uncommitted_changes():
    """커밋되지 않은 변경사항 확인"""
    try:
        result = subprocess.run(
            ["git", "status", "--porcelain"],
            capture_output=True,
            text=True,
            timeout=2
        )
        if result.returncode == 0:
            lines = result.stdout.strip().split('\n')
            return len([l for l in lines if l])
        return 0
    except Exception:
        return 0

def get_recent_errors():
    """최근 에러 로그 조회"""
    try:
        logs_dir = Path(".moai/logs/sessions")
        if logs_dir.exists():
            log_files = sorted(logs_dir.glob("*.log"), reverse=True)[:1]
            if log_files:
                with open(log_files[0], 'r', encoding='utf-8') as f:
                    content = f.read()
                    # 마지막 500자만 추출
                    return content[-500:] if len(content) > 500 else content
        return ""
    except Exception:
        return ""

def get_current_spec():
    """현재 작업 중인 SPEC 감지"""
    try:
        branch = get_current_branch()
        if branch.startswith("feature/SPEC-"):
            return branch.replace("feature/", "")
        return ""
    except Exception:
        return ""

def get_git_log_summary():
    """최근 Git 커밋 요약"""
    try:
        result = subprocess.run(
            ["git", "log", "-5", "--oneline"],
            capture_output=True,
            text=True,
            timeout=2
        )
        return result.stdout.strip() if result.returncode == 0 else ""
    except Exception:
        return ""

def collect_all_info():
    """모든 정보 수집"""
    info = {
        "timestamp": datetime.now().isoformat(),
        "moai_version": get_moai_version(),
        "python_version": get_python_version(),
        "os_info": get_os_info(),
        "project_mode": get_project_mode(),
        "current_branch": get_current_branch(),
        "uncommitted_changes": get_uncommitted_changes(),
        "current_spec": get_current_spec(),
        "recent_git_commits": get_git_log_summary(),
    }

    return info

def format_info_korean(info):
    """정보를 한국어로 포맷팅"""
    formatted = []

    formatted.append("## 🔍 자동 수집된 환경 정보")
    formatted.append("")
    formatted.append(f"**MoAI-ADK 버전**: {info['moai_version']}")
    formatted.append(f"**Python 버전**: {info['python_version']}")
    formatted.append(f"**OS**: {info['os_info']}")
    formatted.append(f"**프로젝트 모드**: {info['project_mode']}")
    formatted.append("")

    formatted.append("## 📋 Git 상태")
    formatted.append("")
    formatted.append(f"**현재 브랜치**: `{info['current_branch']}`")

    if info['current_spec']:
        formatted.append(f"**작업 중인 SPEC**: {info['current_spec']}")

    formatted.append(f"**커밋되지 않은 변경사항**: {info['uncommitted_changes']}개")
    formatted.append("")

    if info['recent_git_commits']:
        formatted.append("**최근 커밋**:")
        for line in info['recent_git_commits'].split('\n'):
            if line:
                formatted.append(f"  - {line}")
        formatted.append("")

    return "\n".join(formatted)

if __name__ == "__main__":
    try:
        info = collect_all_info()

        # JSON 형식으로 출력 (기본)
        if len(sys.argv) > 1 and sys.argv[1] == "--json":
            print(json.dumps(info, ensure_ascii=False, indent=2))
        # 한국어 포맷 출력
        elif len(sys.argv) > 1 and sys.argv[1] == "--korean":
            print(format_info_korean(info))
        # 기본: JSON
        else:
            print(json.dumps(info, ensure_ascii=False, indent=2))

    except Exception as e:
        error_info = {
            "error": str(e),
            "timestamp": datetime.now().isoformat()
        }
        print(json.dumps(error_info, ensure_ascii=False, indent=2), file=sys.stderr)
        sys.exit(1)
