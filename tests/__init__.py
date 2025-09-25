"""
MoAI-ADK Test Suite

This package contains all tests for the MoAI-ADK project.

Test structure:
- test_security.py: Security manager tests
- test_config.py: Configuration system tests
- test_installer.py: Project installer tests
- test_progress_tracker.py: Progress tracking tests
- test_file_manager.py: File management tests
- test_directory_manager.py: Directory management tests
- test_config_manager.py: Configuration management tests
- test_git_manager.py: Git operations tests
- test_system_manager.py: System utilities tests

Usage:
    pytest                     # Run all tests
    pytest tests/unit/         # Run unit tests only
    pytest tests/integration/  # Run integration tests only
    pytest -v --tb=short       # Verbose output with short traceback
    pytest --cov              # Run with coverage
"""

import sys
from pathlib import Path

# Add src to path for testing
src_path = Path(__file__).parent.parent / "src"
if str(src_path) not in sys.path:
    sys.path.insert(0, str(src_path))
