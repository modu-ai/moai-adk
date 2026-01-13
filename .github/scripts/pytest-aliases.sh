#!/bin/bash
# MoAI-ADK Pytest Aliases
# Add these to your ~/.zshrc or ~/.bashrc

# Fast testing (no coverage) - for quick development iteration
alias pytest-fast='cd /Users/goos/MoAI/MoAI-ADK && PYTHONPATH=/Users/goos/MoAI/MoAI-ADK/src uv run pytest tests/ -v --tb=short'

# Coverage testing - for PR verification and quality checks
alias pytest-cov='cd /Users/goos/MoAI/MoAI-ADK && PYTHONPATH=/Users/goos/MoAI/MoAI-ADK/src uv run pytest tests/ --cov=src/moai_adk --cov-report=html --cov-report=term-missing'

# Smoke testing - for deployment verification
alias pytest-smoke='cd /Users/goos/MoAI/MoAI-ADK && PYTHONPATH=/Users/goos/MoAI/MoAI-ADK/src uv run pytest tests/ -m "smoke or critical" -v --tb=short'

# Run specific test file
alias pytest-file='cd /Users/goos/MoAI/MoAI-ADK && PYTHONPATH=/Users/goos/MoAI/MoAI-ADK/src uv run pytest'

# Show available markers
alias pytest-markers='cd /Users/goos/MoAI/MoAI-ADK && uv run pytest --markers'

# Usage:
# pytest-fast          # Quick test without coverage
# pytest-cov           # Full test with coverage
# pytest-smoke         # Smoke tests only (blocking for deployment)
# pytest-file tests/test_smoke.py  # Run specific file
# pytest-markers       # Show all available test markers
