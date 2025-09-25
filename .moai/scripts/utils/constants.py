#!/usr/bin/env python3
"""
MoAI-ADK 스크립트 공통 상수 정의

@REQ:SCRIPT-CONSTANTS-001
@FEATURE:CONSTANTS-MANAGEMENT-001
@API:GET-CONSTANTS
@DESIGN:CONFIGURATION-CENTRALIZATION-001
"""

from pathlib import Path

# 디렉터리 경로 상수
MOAI_DIR_NAME = ".moai"
CLAUDE_DIR_NAME = ".claude"
SCRIPTS_DIR_NAME = "scripts"
CHECKPOINTS_DIR_NAME = "checkpoints"
MEMORY_DIR_NAME = "memory"
HOOKS_DIR_NAME = "hooks"
INDEXES_DIR_NAME = "indexes"
SPECS_DIR_NAME = "specs"

# 파일 이름 상수
CONFIG_FILE_NAME = "config.json"
METADATA_FILE_NAME = "metadata.json"
DEVELOPMENT_GUIDE_FILE_NAME = "development-guide.md"
CLAUDE_MEMORY_FILE_NAME = "CLAUDE.md"
TAGS_INDEX_FILE_NAME = "tags.db"

# Git 관련 상수
DEFAULT_BRANCH_NAME = "main"
CHECKPOINT_TAG_PREFIX = "moai_cp/"
FEATURE_BRANCH_PREFIX = "feature/"
BUGFIX_BRANCH_PREFIX = "bugfix/"
HOTFIX_BRANCH_PREFIX = "hotfix/"

# 체크포인트 관련 상수
MAX_CHECKPOINTS = 10
CHECKPOINT_MESSAGE_MAX_LENGTH = 100
AUTO_CHECKPOINT_INTERVAL_MINUTES = 5
BACKUP_RETENTION_DAYS = 7

# Git 명령어 타임아웃 (초)
GIT_COMMAND_TIMEOUT = 30
GIT_PUSH_TIMEOUT = 60
GIT_PULL_TIMEOUT = 60

# 모드 상수
PERSONAL_MODE = "personal"
TEAM_MODE = "team"
VALID_MODES = [PERSONAL_MODE, TEAM_MODE]

# 프로젝트 유형 상수
PROJECT_TYPES = {
    "web_api": {
        "required_files": ["requirements.txt", "app.py", "api/"],
        "optional_files": ["Dockerfile", ".env.example"],
        "docs": ["API.md", "DEPLOYMENT.md"]
    },
    "cli_tool": {
        "required_files": ["setup.py", "src/", "tests/"],
        "optional_files": ["requirements.txt", "pyproject.toml"],
        "docs": ["CLI_COMMANDS.md", "INSTALLATION.md"]
    },
    "library": {
        "required_files": ["setup.py", "src/", "tests/"],
        "optional_files": ["pyproject.toml", "tox.ini"],
        "docs": ["API_REFERENCE.md", "EXAMPLES.md"]
    },
    "frontend": {
        "required_files": ["package.json", "src/", "public/"],
        "optional_files": ["tsconfig.json", "webpack.config.js"],
        "docs": ["COMPONENTS.md", "STYLING.md"]
    }
}

# TRUST 원칙 관련 상수
TRUST_PRINCIPLES = {
    "test_first": {
        "name": "Test First",
        "description": "테스트 우선",
        "weight": 0.25
    },
    "readable": {
        "name": "Readable",
        "description": "읽기 쉽게",
        "weight": 0.20
    },
    "unified": {
        "name": "Unified",
        "description": "통합 설계",
        "weight": 0.20
    },
    "secured": {
        "name": "Secured",
        "description": "안전하게",
        "weight": 0.20
    },
    "trackable": {
        "name": "Trackable",
        "description": "추적 가능",
        "weight": 0.15
    }
}

# 코드 품질 기준
QUALITY_THRESHOLDS = {
    "test_coverage_min": 85,
    "max_function_lines": 50,
    "max_file_lines": 300,
    "max_parameters": 5,
    "max_complexity": 10,
    "max_modules": 5
}

# 로깅 설정
LOG_LEVELS = {
    "DEBUG": 10,
    "INFO": 20,
    "WARNING": 30,
    "ERROR": 40,
    "CRITICAL": 50
}

# 색상 코드 (터미널 출력용)
COLORS = {
    "RESET": "\033[0m",
    "RED": "\033[31m",
    "GREEN": "\033[32m",
    "YELLOW": "\033[33m",
    "BLUE": "\033[34m",
    "MAGENTA": "\033[35m",
    "CYAN": "\033[36m",
    "WHITE": "\033[37m",
    "BOLD": "\033[1m"
}

# 이모지 상수
EMOJIS = {
    "SUCCESS": "✅",
    "ERROR": "❌",
    "WARNING": "⚠️",
    "INFO": "ℹ️",
    "CHECKPOINT": "📍",
    "BRANCH": "🌿",
    "COMMIT": "💾",
    "ROLLBACK": "↩️",
    "SYNC": "🔄",
    "BUILD": "🔨",
    "TEST": "🧪",
    "DOC": "📚"
}

# 파일 확장자 매핑
FILE_EXTENSIONS = {
    "python": [".py", ".pyi"],
    "javascript": [".js", ".jsx", ".mjs"],
    "typescript": [".ts", ".tsx"],
    "markdown": [".md", ".markdown"],
    "json": [".json"],
    "yaml": [".yaml", ".yml"],
    "text": [".txt"],
    "config": [".conf", ".config", ".ini"]
}

# 기본 에러 메시지
ERROR_MESSAGES = {
    "git_not_found": "Git이 설치되지 않았거나 PATH에 없습니다.",
    "not_git_repo": "Git 저장소가 아닙니다.",
    "no_moai_config": "MoAI 설정 파일을 찾을 수 없습니다.",
    "permission_denied": "권한이 거부되었습니다.",
    "file_not_found": "파일을 찾을 수 없습니다.",
    "invalid_mode": f"유효하지 않은 모드입니다. 가능한 값: {', '.join(VALID_MODES)}",
    "uncommitted_changes": "커밋되지 않은 변경사항이 있습니다."
}

# 성공 메시지
SUCCESS_MESSAGES = {
    "checkpoint_created": "체크포인트가 성공적으로 생성되었습니다.",
    "branch_created": "브랜치가 성공적으로 생성되었습니다.",
    "sync_completed": "동기화가 완료되었습니다.",
    "rollback_completed": "롤백이 완료되었습니다.",
    "commit_completed": "커밋이 완료되었습니다."
}

# 정규표현식 패턴
REGEX_PATTERNS = {
    "branch_name": r"^[a-zA-Z0-9._/-]+$",
    "tag_name": r"^[a-zA-Z0-9._-]+$",
    "spec_id": r"^SPEC-\d{3}$",
    "version_number": r"^\d+\.\d+\.\d+$",
    "git_commit_hash": r"^[a-f0-9]{7,40}$"
}

# 환경 변수 키
ENV_VARS = {
    "MOAI_MODE": "MOAI_MODE",
    "MOAI_DEBUG": "MOAI_DEBUG",
    "MOAI_PROJECT_ROOT": "MOAI_PROJECT_ROOT",
    "GIT_EDITOR": "GIT_EDITOR",
    "TMPDIR": "TMPDIR"
}


def get_moai_dir(project_root: Path) -> Path:
    """MoAI 디렉터리 경로 반환"""
    return project_root / MOAI_DIR_NAME


def get_claude_dir(project_root: Path) -> Path:
    """Claude 디렉터리 경로 반환"""
    return project_root / CLAUDE_DIR_NAME


def get_scripts_dir(project_root: Path) -> Path:
    """스크립트 디렉터리 경로 반환"""
    return get_moai_dir(project_root) / SCRIPTS_DIR_NAME


def get_checkpoints_dir(project_root: Path) -> Path:
    """체크포인트 디렉터리 경로 반환"""
    return get_moai_dir(project_root) / CHECKPOINTS_DIR_NAME
