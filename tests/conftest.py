# @TEST:CLI-001 | SPEC: SPEC-CLI-001.md
"""Pytest configuration and fixtures for CLI testing"""

import pytest
from click.testing import CliRunner


@pytest.fixture
def cli_runner() -> CliRunner:
    """Click CLI 테스트를 위한 CliRunner fixture

    Returns:
        CliRunner: Click 명령어 테스트 러너
    """
    return CliRunner()
