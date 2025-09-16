# MoAI-ADK Makefile
# MoAI Agentic Development Kit 빌드 자동화

.PHONY: build status clean dev help install test

# 기본 타겟
all: build

# 빌드 (변경된 파일만 동기화)
build:
	@echo "🔨 Building MoAI-ADK..."
	@echo "🔄 Auto-syncing versions..."
	@python3 -m moai_adk.core.version_sync --verify
	@python3 -m build

# 강제 빌드 (모든 파일 동기화)
build-force:
	@echo "🔨 Force building MoAI-ADK..."
	@echo "🔄 Force syncing all versions..."
	@python3 -m moai_adk.core.version_sync
	@python3 -m build

# 클린 빌드
build-clean:
	@echo "🧹 Clean building MoAI-ADK..."
	@rm -rf dist/ build/ *.egg-info/
	@echo "🔄 Clean sync all versions..."
	@python3 -m moai_adk.core.version_sync
	@python3 -m build

# 빌드 상태 확인
status:
	@echo "📊 Checking MoAI-ADK build status..."
	@ls -la dist/ build/ *.egg-info/ 2>/dev/null || echo "No build artifacts found"
	@python3 -c "import sys; sys.path.insert(0, 'src'); from moai_adk._version import get_version_format; print(f'Current version: {get_version_format(\"short\")}')"

# 정리
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf dist/ build/ *.egg-info/ __pycache__ src/**/__pycache__ 2>/dev/null || true
	@find . -name "*.pyc" -delete 2>/dev/null || true
	@echo "✅ Cleanup completed"

# 개발 모드 (파일 감시)
dev:
	@echo "👀 Starting development mode..."
	@echo "Note: Install 'watchdog' for file watching: pip install watchdog"
	@python3 -c "import time; print('Development mode active - use Ctrl+C to stop'); [print('.', end='', flush=True) or time.sleep(1) for _ in iter(int, 1)]" || echo "Stopped"

# MoAI-ADK 설치 (대화형)
install:
	@echo "📦 Installing MoAI-ADK..."
	@python3 src/installer.py

# 설치 (자동)
install-auto:
	@echo "📦 Installing MoAI-ADK (auto mode)..."
	@python3 src/installer.py --auto

# 테스트 (전체 시스템)
test:
	@echo "🧪 Running comprehensive test suite..."
	@./scripts/run-tests.sh

# Hook 시스템 테스트
test-hooks:
	@echo "🧪 Testing Hook system..."
	@python3 tests/test_hooks.py

# 빌드 시스템 테스트
test-build:
	@echo "🔨 Testing Build system..."
	@python3 tests/test_build.py

# 빠른 테스트 (설정만)
test-quick:
	@echo "⚡ Quick configuration tests..."
	@cd src/templates && python3 .claude/hooks/moai/config_loader.py
	@make validate

# 상세 테스트 (verbose)
test-verbose:
	@echo "🔍 Running verbose tests..."
	@./scripts/run-tests.sh --verbose

# Coverage 테스트
test-coverage:
	@echo "📊 Running tests with coverage..."
	@./scripts/run-tests.sh --coverage

# CI 테스트 (JUnit 포함)
test-ci:
	@echo "🤖 Running CI tests..."
	@./scripts/run-tests.sh --junit --coverage

# 버전 정보
version:
	@python3 -c "import sys; sys.path.insert(0, 'src'); from _version import get_version_format; print(get_version_format('banner'))"
	@python3 --version

# 버전 관리 명령어들
version-sync:
	@echo "🔄 Synchronizing version across all files..."
	@python3 -m moai_adk.core.version_sync

version-sync-dry:
	@echo "🔍 Dry run: checking version synchronization..."
	@python3 -m moai_adk.core.version_sync --dry-run

version-verify:
	@echo "🔍 Verifying version consistency..."
	@python3 -m moai_adk.core.version_sync --verify

version-update:
	@echo "📝 Usage: make version-set VERSION=0.2.0"
	@echo "This will update _version.py and sync all files"

# version-set target temporarily disabled due to complex Python script
# Use scripts/update_version.py directly for version updates
version-set:
	@echo "⚠️  version-set target needs fixing"
	@echo "Use: python scripts/update_version.py --version VERSION"

# 도움말
help:
	@echo "🗿 MoAI-ADK Build System Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  build         - Build (sync changed files only)"
	@echo "  build-force   - Force build (sync all files)"  
	@echo "  build-clean   - Clean build (remove dist first)"
	@echo "  status        - Check build status"
	@echo "  clean         - Clean dist directory"
	@echo ""
	@echo "Development:"
	@echo "  dev           - Development mode (watch for changes)"
	@echo "  test          - Test Hook system"
	@echo ""
	@echo "Installation:"
	@echo "  install       - Interactive installation"
	@echo "  install-auto  - Automatic installation"
	@echo ""
	@echo "Version Management:"
	@echo "  version          - Show current version info"
	@echo "  version-sync     - Sync version across all project files"
	@echo "  version-sync-dry - Dry run version sync (preview only)"
	@echo "  version-verify   - Verify version consistency"
	@echo "  version-set      - Update version (Usage: make version-set VERSION=0.2.0)"
	@echo ""
	@echo "Utility:"
	@echo "  help          - Show this help"

# 설정 검증
validate:
	@echo "🔍 Validating configurations..."
	@cd src/templates && python3 -c "import json; json.load(open('.claude/settings.json')); print('✅ .claude/settings.json is valid')"
	@cd src/templates && python3 -c "import json; json.load(open('.moai/config.json')); print('✅ .moai/config.json is valid')"
	@echo "✅ All configurations are valid"

# 권한 설정
permissions:
	@echo "🔐 Setting up permissions..."
	@chmod +x build.py
	@chmod +x src/installer.py
	@chmod +x src/templates/.claude/hooks/moai/*.py
	@echo "✅ Permissions set"

# 종속성 확인
deps:
	@echo "📋 Checking dependencies..."
	@python3 -c "import sys; print(f'Python: {sys.version}')"
	@python3 -c "import json; print('✅ json module available')"
	@python3 -c "import pathlib; print('✅ pathlib module available')"
	@python3 -c "import hashlib; print('✅ hashlib module available')"
	@python3 -c "import shutil; print('✅ shutil module available')"
	@echo "✅ Core dependencies satisfied"

# 개발 환경 설정
setup: permissions deps validate
	@echo "⚙️ Setting up development environment..."
	@$(MAKE) build
	@echo "✅ Development environment ready"

# 프로덕션 배포 준비
release: setup
	@echo "🚀 Preparing for release..."
	@$(MAKE) build-clean
	@$(MAKE) test
	@$(MAKE) validate
	@echo "✅ Ready for release"