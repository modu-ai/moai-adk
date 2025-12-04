"""
Shell 테스트 공유 픽스처 및 설정
"""

import subprocess
import sys
from pathlib import Path
from typing import Generator

import pytest

# ==============================================================================
# Shell 감지 및 검증
# ==============================================================================


@pytest.fixture(scope="session")
def powershell_available() -> bool:
    """PowerShell이 설치되어 있는지 확인"""
    try:
        result = subprocess.run(
            ["pwsh", "--version"],
            capture_output=True,
            text=True,
            timeout=5,
        )
        return result.returncode == 0
    except (FileNotFoundError, subprocess.TimeoutExpired):
        return False


@pytest.fixture(scope="session")
def bash_available() -> bool:
    """Bash가 설치되어 있는지 확인"""
    try:
        result = subprocess.run(
            ["bash", "--version"],
            capture_output=True,
            text=True,
            timeout=5,
        )
        return result.returncode == 0
    except (FileNotFoundError, subprocess.TimeoutExpired):
        return False


# ==============================================================================
# PowerShell 명령 실행 헬퍼
# ==============================================================================


@pytest.fixture
def run_powershell():
    """PowerShell 스크립트를 실행하고 결과를 반환하는 헬퍼"""

    def _run(script: str, *args, check: bool = True, **kwargs) -> subprocess.CompletedProcess:
        """
        PowerShell 스크립트 실행

        Args:
            script: PowerShell 코드 (문자열 또는 파일 경로)
            check: True이면 실패 시 예외 발생
            **kwargs: subprocess.run에 전달될 추가 인자

        Returns:
            CompletedProcess 객체
        """
        cmd = ["pwsh", "-NoProfile", "-Command", script]
        return subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            check=check,
            timeout=30,
            **kwargs,
        )

    return _run


@pytest.fixture
def run_bash():
    """Bash 스크립트를 실행하고 결과를 반환하는 헬퍼"""

    def _run(script: str, *args, check: bool = True, **kwargs) -> subprocess.CompletedProcess:
        """
        Bash 스크립트 실행

        Args:
            script: Bash 코드 (문자열 또는 파일 경로)
            check: True이면 실패 시 예외 발생
            **kwargs: subprocess.run에 전달될 추가 인자

        Returns:
            CompletedProcess 객체
        """
        cmd = ["bash", "-c", script]
        return subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            check=check,
            timeout=30,
            **kwargs,
        )

    return _run


# ==============================================================================
# 크로스플랫폼 환경 관리
# ==============================================================================


@pytest.fixture
def moai_project_root() -> Path:
    """MoAI-ADK 프로젝트 루트 경로"""
    current = Path(__file__).parent.parent.parent
    assert (current / "pyproject.toml").exists(), "MoAI-ADK 프로젝트를 찾을 수 없음"
    return current


@pytest.fixture
def temp_test_env(tmp_path: Path) -> Generator[dict, None, None]:
    """임시 테스트 환경 설정"""
    env = {
        "project_root": tmp_path,
        "venv_path": tmp_path / ".venv",
        "scripts_path": tmp_path / "scripts",
    }

    # 필요한 디렉토리 생성
    (tmp_path / "src").mkdir()
    env["scripts_path"].mkdir()

    yield env


# ==============================================================================
# 설치 및 패키지 관리 헬퍼
# ==============================================================================


@pytest.fixture
def install_moai_package(run_powershell, run_bash):
    """MoAI-ADK 패키지를 설치하는 헬퍼"""

    def _install(project_root: Path, shell: str = "auto") -> bool:
        """
        패키지 설치

        Args:
            project_root: 프로젝트 루트 경로
            shell: "powershell", "bash", 또는 "auto" (OS 자동 감지)

        Returns:
            성공 여부
        """
        if shell == "auto":
            shell = "powershell" if sys.platform == "win32" else "bash"

        install_cmd = f"cd {project_root} && pip install -e .[dev]"

        if shell == "powershell":
            result = run_powershell(install_cmd)
        else:
            result = run_bash(install_cmd)

        return result.returncode == 0

    return _install


# ==============================================================================
# 테스트 마커
# ==============================================================================


def pytest_configure(config):
    """PowerShell 관련 마커 등록"""
    config.addinivalue_line(
        "markers",
        "powershell: PowerShell에서만 실행되는 테스트",
    )
    config.addinivalue_line(
        "markers",
        "bash: Bash에서만 실행되는 테스트",
    )
    config.addinivalue_line(
        "markers",
        "cross_platform: 여러 셸에서 실행되는 테스트",
    )
    config.addinivalue_line(
        "markers",
        "windows_only: Windows에서만 실행되는 테스트",
    )
    config.addinivalue_line(
        "markers",
        "unix_only: Unix/Linux/macOS에서만 실행되는 테스트",
    )
