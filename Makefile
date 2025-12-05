# MoAI-ADK Development Makefile
# Simplified development workflow commands

.PHONY: help lint format test test-quick test-full clean release-patch release-minor release-major

help:
	@echo "🗿 MoAI-ADK Development Commands"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint          - Run ruff linter (with auto-fix)"
	@echo "  make format        - Format code with ruff"
	@echo ""
	@echo "Testing:"
	@echo "  make test          - Run quick tests (fail-fast)"
	@echo "  make test-full     - Run full test suite with coverage"
	@echo ""
	@echo "Release:"
	@echo "  make release-patch - Patch version bump (0.32.7 → 0.32.8)"
	@echo "  make release-minor - Minor version bump (0.32.7 → 0.33.0)"
	@echo "  make release-major - Major version bump (0.32.7 → 1.0.0)"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean         - Remove build artifacts and cache"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Code Quality
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

lint:
	@echo "🔧 Running ruff linter..."
	uv run ruff check src/ tests/ .claude/ --fix
	@echo "✅ Linting complete"

format:
	@echo "🎨 Formatting code with ruff..."
	uv run ruff format src/ tests/ .claude/
	@echo "✅ Formatting complete"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Testing
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

test:
	@echo "🧪 Running quick tests (fail-fast)..."
	uv run pytest tests/ -x --tb=short -q --maxfail=5
	@echo "✅ Quick tests passed"

test-quick: test

test-full:
	@echo "🧪 Running full test suite with coverage..."
	uv run pytest tests/ -v --tb=short --cov=src/moai_adk --cov-report=term --cov-report=html
	@echo "✅ Full tests complete"
	@echo "📊 Coverage report: htmlcov/index.html"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Release (requires /moai:99-release or manual version bump)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

release-patch:
	@echo "🚀 Starting patch release..."
	@echo "⚠️  This will trigger /moai:99-release workflow"
	@echo "Use: /moai:99-release --fast"
	@echo ""
	@echo "Or manually:"
	@echo "1. Bump version in pyproject.toml (patch)"
	@echo "2. Update CHANGELOG.md"
	@echo "3. git commit -m 'chore: Bump version to X.Y.Z'"
	@echo "4. git push origin release/vX.Y.Z"

release-minor:
	@echo "🚀 Starting minor release..."
	@echo "⚠️  This will trigger /moai:99-release workflow"
	@echo "Use: /moai:99-release --fast"

release-major:
	@echo "🚀 Starting major release..."
	@echo "⚠️  This will trigger /moai:99-release workflow"
	@echo "Use: /moai:99-release --fast"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Cleanup
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

clean:
	@echo "🧹 Cleaning build artifacts and cache..."
	rm -rf dist/
	rm -rf build/
	rm -rf *.egg-info
	rm -rf .pytest_cache
	rm -rf .ruff_cache
	rm -rf htmlcov/
	rm -rf .coverage
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	@echo "✅ Cleanup complete"
