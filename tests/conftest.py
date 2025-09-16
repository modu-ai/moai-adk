"""
Pytest configuration and shared fixtures for MoAI-ADK tests.

This file contains common test configuration, fixtures, and utilities
used across all test modules in the project.
"""

import os
import sys
import tempfile
from pathlib import Path
from typing import Generator

import pytest

# Add src to Python path for testing
test_dir = Path(__file__).parent
project_root = test_dir.parent
src_dir = project_root / "src"
sys.path.insert(0, str(src_dir))


@pytest.fixture
def temp_dir() -> Generator[Path, None, None]:
    """Create a temporary directory for tests."""
    with tempfile.TemporaryDirectory() as tmp_dir:
        yield Path(tmp_dir)


@pytest.fixture
def sample_project_dir(temp_dir: Path) -> Path:
    """Create a sample project directory structure for testing."""
    project_dir = temp_dir / "sample_project"
    project_dir.mkdir()

    # Create basic project structure
    (project_dir / "src").mkdir()
    (project_dir / "tests").mkdir()
    (project_dir / "README.md").write_text("# Sample Project\n")
    (project_dir / ".gitignore").write_text("*.pyc\n__pycache__/\n")

    return project_dir


@pytest.fixture
def mock_config():
    """Mock configuration for testing."""
    class MockConfig:
        def __init__(self):
            self.project_name = "test_project"
            self.version = "0.1.0"
            self.description = "Test project description"

        def to_dict(self):
            return {
                "project_name": self.project_name,
                "version": self.version,
                "description": self.description
            }

    return MockConfig()


# Test markers
pytest.mark.unit = pytest.mark.unit
pytest.mark.integration = pytest.mark.integration
pytest.mark.slow = pytest.mark.slow