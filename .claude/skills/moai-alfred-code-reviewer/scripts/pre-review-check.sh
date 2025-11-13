#!/bin/bash
# Pre-review automation script for code quality checks

set -e

echo "🔍 Running automated code review checks..."

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Not in a git repository"
    exit 1
fi

# Test Coverage Check (if pytest available)
if command -v pytest &> /dev/null; then
    echo "📊 Checking test coverage..."
    if pytest --cov=src --cov-fail-under=85 --cov-report=term-missing 2>/dev/null; then
        echo "✅ Test coverage meets requirements"
    else
        echo "⚠️ Test coverage below 85% or pytest configuration missing"
    fi
else
    echo "⚠️ pytest not available, skipping coverage check"
fi

# Code Quality Checks (if ruff available)
if command -v ruff &> /dev/null; then
    echo "🧹 Running ruff linter..."
    if ruff check src/ --show-source; then
        echo "✅ Ruff checks passed"
    else
        echo "⚠️ Ruff found issues"
    fi
else
    echo "⚠️ ruff not available, skipping linting"
fi

# Type checking (if mypy available)
if command -v mypy &> /dev/null; then
    echo "🔍 Running type checks..."
    if mypy src/ --strict 2>/dev/null; then
        echo "✅ Type checks passed"
    else
        echo "⚠️ Type checking found issues"
    fi
else
    echo "⚠️ mypy not available, skipping type checking"
fi

# Security Scanning (if bandit available)
if command -v bandit &> /dev/null; then
    echo "🔒 Scanning for security issues..."
    if bandit -r src/ -f json -o bandit-report.json 2>/dev/null; then
        echo "✅ Security scan completed"
    else
        echo "⚠️ Security scanner found issues or failed to run"
    fi
else
    echo "⚠️ bandit not available, skipping security scan"
fi

echo "🎉 Automated review checks completed!"
