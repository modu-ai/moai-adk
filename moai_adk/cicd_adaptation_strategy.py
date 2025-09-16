#!/usr/bin/env python3
"""
CI/CD Adaptation Strategy for MoAI-ADK Package Restructuring
Automated generation of updated CI/CD configurations
"""

import sys
from pathlib import Path
from typing import Dict, List
from dataclasses import dataclass

@dataclass
class TestCommand:
    """Represents a test command in the CI/CD pipeline."""
    name: str
    description: str
    command: str
    dependencies: List[str]
    parallel_safe: bool
    estimated_time: int  # seconds

class CICDAdaptationGenerator:
    """Generates updated CI/CD configurations for package restructuring."""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.makefile_path = project_root / "Makefile"
        self.test_script_path = project_root / "scripts" / "run-tests.sh"

    def generate_enhanced_makefile(self) -> str:
        """Generate enhanced Makefile with package-specific test targets."""
        return '''# MoAI-ADK Enhanced Makefile
# MoAI Agentic Development Kit 빌드 자동화 with Package-Specific Testing

.PHONY: build status clean dev help install test test-all test-packages test-parallel

# 기본 타겟
all: build

# 기존 빌드 명령어들 (유지)
build:
	@echo "🔨 Building MoAI-ADK..."
	@echo "🔄 Auto-syncing versions..."
	@python3 -m moai_adk.core.version_sync --verify
	@python3 -m build

build-force:
	@echo "🔨 Force building MoAI-ADK..."
	@echo "🔄 Force syncing all versions..."
	@python3 -m moai_adk.core.version_sync
	@python3 -m build

build-clean:
	@echo "🧹 Clean building MoAI-ADK..."
	@rm -rf dist/ build/ *.egg-info/
	@echo "🔄 Clean sync all versions..."
	@python3 -m moai_adk.core.version_sync
	@python3 -m build

# ============================================================================
# ENHANCED TESTING SYSTEM - Package-Specific Testing
# ============================================================================

# 전체 테스트 (기존 방식 + 새로운 패키지 테스트)
test: test-legacy test-packages
	@echo "✅ All tests completed!"

# 레거시 테스트 (기존 테스트 유지)
test-legacy:
	@echo "🧪 Running legacy test suite..."
	@./scripts/run-tests.sh

# 새로운 패키지별 테스트 시스템
test-packages: test-config test-core test-utils test-install test-integration
	@echo "📦 All package tests completed!"

# Config 패키지 테스트
test-config:
	@echo "⚙️  Testing Config package..."
	@python3 -m pytest tests/unit/config/ -v --tb=short -x
	@echo "✅ Config tests passed!"

# Core 패키지 테스트 (세분화)
test-core: test-core-managers test-core-security test-core-validation
	@echo "🏗️  All Core package tests completed!"

test-core-managers:
	@echo "👥 Testing Core Managers..."
	@python3 -m pytest tests/unit/core/managers/ -v --tb=short -x

test-core-security:
	@echo "🔐 Testing Core Security..."
	@python3 -m pytest tests/unit/core/security/ -v --tb=short -x

test-core-validation:
	@echo "✅ Testing Core Validation..."
	@python3 -m pytest tests/unit/core/validation/ -v --tb=short -x

# Utils 패키지 테스트
test-utils:
	@echo "🛠️  Testing Utils package..."
	@python3 -m pytest tests/unit/utils/ -v --tb=short -x
	@echo "✅ Utils tests passed!"

# Install 패키지 테스트
test-install:
	@echo "📦 Testing Install package..."
	@python3 -m pytest tests/unit/install/ -v --tb=short -x
	@echo "✅ Install tests passed!"

# Integration 테스트 (기존 + 새로운 패키지 통합)
test-integration:
	@echo "🔗 Testing package integration..."
	@python3 -m pytest tests/integration/ -v --tb=short -x
	@echo "✅ Integration tests passed!"

# ============================================================================
# PERFORMANCE AND ADVANCED TESTING
# ============================================================================

# 성능 테스트
test-performance:
	@echo "⚡ Running performance tests..."
	@python3 -m pytest tests/performance/ -v --benchmark-only
	@echo "✅ Performance tests completed!"

# Import 구조 검증
test-imports:
	@echo "🔍 Validating import structure..."
	@python3 tests/performance/test_import_performance.py
	@python3 tests/performance/test_dependency_cycles.py
	@echo "✅ Import validation completed!"

# 메모리 사용량 테스트
test-memory:
	@echo "🧠 Testing memory usage..."
	@python3 -m pytest tests/performance/test_memory_usage.py -v
	@echo "✅ Memory tests completed!"

# ============================================================================
# PARALLEL TESTING SYSTEM
# ============================================================================

# 병렬 테스트 실행 (pytest-xdist 필요)
test-parallel:
	@echo "🚀 Running tests in parallel..."
	@python3 -m pytest -n auto --dist=loadfile tests/ -v
	@echo "✅ Parallel tests completed!"

# 빠른 병렬 테스트 (실패 시 즉시 중단)
test-parallel-fast:
	@echo "⚡ Running fast parallel tests..."
	@python3 -m pytest -n auto --dist=loadfile -x --ff tests/ -v
	@echo "✅ Fast parallel tests completed!"

# ============================================================================
# CI/CD SPECIFIC COMMANDS
# ============================================================================

# CI 환경용 테스트 (JUnit + Coverage + Parallel)
test-ci:
	@echo "🤖 Running CI test suite..."
	@mkdir -p test-reports coverage_html
	@python3 -m pytest -n auto --dist=loadfile \\
		--cov=moai_adk \\
		--cov-report=term-missing \\
		--cov-report=html:coverage_html \\
		--cov-report=xml:coverage.xml \\
		--junit-xml=test-reports/junit.xml \\
		tests/
	@echo "✅ CI tests completed!"

# 품질 게이트 검사 (코드 커버리지 + 성능)
test-quality-gate:
	@echo "🎯 Running quality gate checks..."
	@$(MAKE) test-ci
	@$(MAKE) test-imports
	@$(MAKE) test-performance
	@python3 scripts/quality_gate_check.py
	@echo "✅ Quality gate passed!"

# ============================================================================
# DEVELOPMENT AND DEBUGGING
# ============================================================================

# 개발자용 빠른 테스트 (변경된 파일만)
test-dev:
	@echo "👨‍💻 Running developer tests..."
	@python3 -m pytest --lf --ff -v tests/
	@echo "✅ Developer tests completed!"

# 상세 테스트 (디버깅용)
test-verbose:
	@echo "🔍 Running verbose tests..."
	@python3 -m pytest -v -s --tb=long tests/
	@echo "✅ Verbose tests completed!"

# 특정 패키지 테스트 (매개변수 사용)
test-package:
	@echo "📦 Testing specific package: $(PACKAGE)"
	@python3 -m pytest tests/unit/$(PACKAGE)/ -v --tb=short
	@echo "✅ Package $(PACKAGE) tests completed!"

# ============================================================================
# MONITORING AND REPORTING
# ============================================================================

# 테스트 결과 요약 생성
test-report:
	@echo "📊 Generating test reports..."
	@python3 scripts/generate_test_report.py
	@echo "✅ Test reports generated!"

# 커버리지 보고서 생성 및 열기
coverage-report:
	@echo "📈 Generating coverage report..."
	@python3 -m pytest --cov=moai_adk --cov-report=html:coverage_html tests/
	@echo "🌐 Opening coverage report..."
	@open coverage_html/index.html || xdg-open coverage_html/index.html || echo "Open coverage_html/index.html manually"

# 테스트 성능 벤치마크
test-benchmark:
	@echo "🏁 Running test performance benchmarks..."
	@time $(MAKE) test-packages
	@echo "✅ Benchmark completed!"

# ============================================================================
# MAINTENANCE AND CLEANUP
# ============================================================================

# 테스트 아티팩트 정리
test-clean:
	@echo "🧹 Cleaning test artifacts..."
	@rm -rf test-reports/ coverage_html/ .coverage .pytest_cache/
	@find . -name "*.pyc" -delete
	@find . -name "__pycache__" -type d -exec rm -rf {} + 2>/dev/null || true
	@echo "✅ Test artifacts cleaned!"

# 테스트 환경 설정
test-setup:
	@echo "⚙️  Setting up test environment..."
	@pip install pytest pytest-cov pytest-xdist pytest-benchmark
	@python3 setup_test_structure.py .
	@echo "✅ Test environment ready!"

# ============================================================================
# HELP AND DOCUMENTATION
# ============================================================================

help-testing:
	@echo "🧪 MoAI-ADK Testing Commands:"
	@echo ""
	@echo "Basic Testing:"
	@echo "  test                 - Run all tests (legacy + packages)"
	@echo "  test-packages        - Run new package-specific tests"
	@echo "  test-config          - Test config package only"
	@echo "  test-core            - Test all core packages"
	@echo "  test-utils           - Test utils package only"
	@echo "  test-install         - Test install package only"
	@echo "  test-integration     - Test package integration"
	@echo ""
	@echo "Advanced Testing:"
	@echo "  test-performance     - Run performance benchmarks"
	@echo "  test-imports         - Validate import structure"
	@echo "  test-memory          - Test memory usage"
	@echo "  test-parallel        - Run tests in parallel"
	@echo "  test-ci              - CI/CD test suite with reports"
	@echo "  test-quality-gate    - Full quality gate validation"
	@echo ""
	@echo "Development:"
	@echo "  test-dev             - Quick developer tests"
	@echo "  test-verbose         - Detailed test output"
	@echo "  test-package PACKAGE=name - Test specific package"
	@echo ""
	@echo "Maintenance:"
	@echo "  test-clean           - Clean test artifacts"
	@echo "  test-setup           - Setup test environment"
	@echo "  coverage-report      - Generate and open coverage report"
	@echo ""

# 기존 명령어들 (유지)
status:
	@echo "📊 Checking MoAI-ADK build status..."
	@ls -la dist/ build/ *.egg-info/ 2>/dev/null || echo "No build artifacts found"
	@python3 -c "import sys; sys.path.insert(0, 'src'); from moai_adk._version import get_version_format; print(f'Current version: {get_version_format(\"short\")}')"

clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf dist/ build/ *.egg-info/ __pycache__ src/**/__pycache__ 2>/dev/null || true
	@find . -name "*.pyc" -delete 2>/dev/null || true
	@echo "✅ Cleanup completed"

# 통합 도움말 (기존 + 새로운 테스트 명령어)
help: help-testing
	@echo ""
	@echo "🗿 MoAI-ADK Build System Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  build         - Build (sync changed files only)"
	@echo "  build-force   - Force build (sync all files)"
	@echo "  build-clean   - Clean build (remove dist first)"
	@echo "  status        - Check build status"
	@echo "  clean         - Clean dist directory"
	@echo ""
	@echo "Installation:"
	@echo "  install       - Interactive installation"
	@echo "  install-auto  - Automatic installation"
	@echo ""
	@echo "Version Management:"
	@echo "  version          - Show current version info"
	@echo "  version-sync     - Sync version across all project files"
	@echo "  version-verify   - Verify version consistency"
	@echo ""

# 설정 검증 (기존 유지)
validate:
	@echo "🔍 Validating configurations..."
	@cd src/templates && python3 -c "import json; json.load(open('.claude/settings.json')); print('✅ .claude/settings.json is valid')"
	@cd src/templates && python3 -c "import json; json.load(open('.moai/config.json')); print('✅ .moai/config.json is valid')"
	@echo "✅ All configurations are valid"

# 프로덕션 배포 준비 (개선)
release: setup test-quality-gate
	@echo "🚀 Preparing for release..."
	@$(MAKE) build-clean
	@$(MAKE) test-ci
	@$(MAKE) validate
	@echo "✅ Ready for release"

# 개발 환경 설정 (기존 + 테스트 설정)
setup: permissions deps validate test-setup
	@echo "⚙️ Setting up development environment..."
	@$(MAKE) build
	@echo "✅ Development environment ready"

# 기존 명령어들 (유지)
permissions:
	@echo "🔐 Setting up permissions..."
	@chmod +x build.py
	@chmod +x src/installer.py
	@chmod +x src/templates/.claude/hooks/moai/*.py
	@echo "✅ Permissions set"

deps:
	@echo "📋 Checking dependencies..."
	@python3 -c "import sys; print(f'Python: {sys.version}')"
	@python3 -c "import json; print('✅ json module available')"
	@python3 -c "import pathlib; print('✅ pathlib module available')"
	@python3 -c "import hashlib; print('✅ hashlib module available')"
	@python3 -c "import shutil; print('✅ shutil module available')"
	@echo "✅ Core dependencies satisfied"

version:
	@python3 -c "import sys; sys.path.insert(0, 'src'); from _version import get_version_format; print(get_version_format('banner'))"
	@python3 --version

version-sync:
	@echo "🔄 Synchronizing version across all files..."
	@python3 -m moai_adk.core.version_sync

version-verify:
	@echo "🔍 Verifying version consistency..."
	@python3 -m moai_adk.core.version_sync --verify
'''

    def generate_enhanced_test_runner(self) -> str:
        """Generate enhanced test runner script with package-specific testing."""
        return '''#!/bin/bash
# Enhanced MoAI-ADK Test Runner
# 패키지별 테스트와 성능 모니터링 지원

set -e

# 색상 정의 (기존)
RED='\\033[0;31m'
GREEN='\\033[0;32m'
YELLOW='\\033[1;33m'
BLUE='\\033[0;34m'
PURPLE='\\033[0;35m'
CYAN='\\033[0;36m'
NC='\\033[0m' # No Color

# 향상된 함수 정의
print_header() {
    echo -e "${BLUE}============================================${NC}"
    echo -e "${BLUE}🧪 Enhanced MoAI-ADK Test Suite Runner${NC}"
    echo -e "${BLUE}============================================${NC}"
}

print_section() {
    echo -e "\\n${YELLOW}📋 $1${NC}"
    echo "----------------------------------------"
}

print_package_section() {
    echo -e "\\n${PURPLE}📦 $1${NC}"
    echo "----------------------------------------"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️ $1${NC}"
}

print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

# 성능 측정 함수
measure_time() {
    local start_time=$(date +%s.%N)
    "$@"
    local exit_code=$?
    local end_time=$(date +%s.%N)
    local duration=$(echo "$end_time - $start_time" | bc -l)
    echo -e "${CYAN}⏱️  Duration: ${duration}s${NC}"
    return $exit_code
}

# 프로젝트 루트 디렉토리 확인
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# 향상된 설정
PYTHON_CMD=${PYTHON_CMD:-python3}
VERBOSE=${VERBOSE:-0}
COVERAGE=${COVERAGE:-0}
JUNIT=${JUNIT:-0}
PARALLEL=${PARALLEL:-0}
PACKAGE_ONLY=${PACKAGE_ONLY:-}
PERFORMANCE=${PERFORMANCE:-0}
MEMORY_CHECK=${MEMORY_CHECK:-0}

# 명령행 인수 처리 (확장)
while [[ $# -gt 0 ]]; do
    case $1 in
        --verbose|-v)
            VERBOSE=1
            shift
            ;;
        --coverage|-c)
            COVERAGE=1
            shift
            ;;
        --junit|-j)
            JUNIT=1
            shift
            ;;
        --parallel|-p)
            PARALLEL=1
            shift
            ;;
        --package)
            PACKAGE_ONLY="$2"
            shift 2
            ;;
        --performance)
            PERFORMANCE=1
            shift
            ;;
        --memory-check)
            MEMORY_CHECK=1
            shift
            ;;
        --python)
            PYTHON_CMD="$2"
            shift 2
            ;;
        --help|-h)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  --verbose, -v       Verbose output"
            echo "  --coverage, -c      Run with coverage report"
            echo "  --junit, -j         Generate JUnit XML report"
            echo "  --parallel, -p      Run tests in parallel"
            echo "  --package PKG       Test specific package only"
            echo "  --performance       Include performance tests"
            echo "  --memory-check      Monitor memory usage"
            echo "  --python CMD        Python command to use (default: python3)"
            echo "  --help, -h          Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

print_header

# 환경 확인 (기존 + 추가 모듈)
print_section "Environment Check"
echo "Python: $($PYTHON_CMD --version)"
echo "Project Root: $PROJECT_ROOT"

# 필요한 Python 모듈 확인 (확장)
echo "Checking Python modules..."
REQUIRED_MODULES=("unittest" "json" "pathlib" "hashlib" "tempfile" "ast" "sys")
OPTIONAL_MODULES=("coverage" "pytest" "pytest_cov" "pytest_xdist" "pytest_benchmark" "psutil" "bc")

for module in "${REQUIRED_MODULES[@]}"; do
    if $PYTHON_CMD -c "import $module" 2>/dev/null; then
        print_success "$module module available"
    else
        print_error "$module module not available"
        exit 1
    fi
done

for module in "${OPTIONAL_MODULES[@]}"; do
    if $PYTHON_CMD -c "import $module" 2>/dev/null; then
        print_success "$module module available"
    else
        print_warning "$module module not available (optional)"
    fi
done

# 테스트 결과 변수 (확장)
TOTAL_TESTS=0
TOTAL_FAILURES=0
TOTAL_ERRORS=0
TOTAL_SKIPPED=0
ALL_SUCCESS=true
PERFORMANCE_RESULTS=()

# Coverage 설정 (기존 유지)
if [[ $COVERAGE -eq 1 ]]; then
    if $PYTHON_CMD -c "import coverage" 2>/dev/null; then
        COVERAGE_CMD="$PYTHON_CMD -m coverage run --source=moai_adk --omit=tests/*,*/__pycache__/*"
        COVERAGE_AVAILABLE=1
        print_success "Coverage enabled"
    else
        print_warning "Coverage requested but not available"
        COVERAGE_AVAILABLE=0
        COVERAGE_CMD="$PYTHON_CMD"
    fi
else
    COVERAGE_CMD="$PYTHON_CMD"
    COVERAGE_AVAILABLE=0
fi

# Parallel 설정
if [[ $PARALLEL -eq 1 ]]; then
    if $PYTHON_CMD -c "import pytest_xdist" 2>/dev/null; then
        PYTEST_CMD="$PYTHON_CMD -m pytest -n auto --dist=loadfile"
        PARALLEL_AVAILABLE=1
        print_success "Parallel testing enabled"
    else
        print_warning "Parallel testing requested but pytest-xdist not available"
        PYTEST_CMD="$PYTHON_CMD -m pytest"
        PARALLEL_AVAILABLE=0
    fi
else
    PYTEST_CMD="$PYTHON_CMD -m pytest"
    PARALLEL_AVAILABLE=0
fi

# ============================================================================
# 새로운 패키지별 테스트 시스템
# ============================================================================

run_package_tests() {
    local package_name=$1
    local package_path="tests/unit/$package_name"

    if [[ ! -d "$package_path" ]]; then
        print_warning "Package tests not found: $package_path"
        return 0
    fi

    print_package_section "Testing $package_name Package"

    if [[ $VERBOSE -eq 1 ]]; then
        TEST_OUTPUT=$(measure_time $PYTEST_CMD "$package_path" -v --tb=short 2>&1)
    else
        TEST_OUTPUT=$(measure_time $PYTEST_CMD "$package_path" --tb=line 2>&1)
    fi

    echo "$TEST_OUTPUT"

    # 결과 파싱
    if echo "$TEST_OUTPUT" | grep -q "failed\\|error"; then
        print_error "$package_name tests failed"
        ALL_SUCCESS=false
    else
        print_success "$package_name tests passed"
    fi

    # 통계 추출 및 업데이트
    extract_test_stats "$TEST_OUTPUT"
}

extract_test_stats() {
    local output="$1"

    # pytest 출력에서 통계 추출
    if echo "$output" | grep -q "collected"; then
        local collected=$(echo "$output" | grep "collected" | sed 's/.* \\([0-9]*\\) items.*/\\1/')
        local failed=$(echo "$output" | grep -o "[0-9]* failed" | grep -o "[0-9]*" || echo "0")
        local passed=$(echo "$output" | grep -o "[0-9]* passed" | grep -o "[0-9]*" || echo "0")
        local skipped=$(echo "$output" | grep -o "[0-9]* skipped" | grep -o "[0-9]*" || echo "0")

        TOTAL_TESTS=$((TOTAL_TESTS + collected))
        TOTAL_FAILURES=$((TOTAL_FAILURES + failed))
        TOTAL_SKIPPED=$((TOTAL_SKIPPED + skipped))
    fi
}

# 패키지별 테스트 실행
if [[ -n "$PACKAGE_ONLY" ]]; then
    print_info "Running tests for package: $PACKAGE_ONLY"
    run_package_tests "$PACKAGE_ONLY"
else
    # 모든 패키지 테스트
    print_section "Package-Specific Tests"

    # 순서대로 패키지 테스트 실행
    for package in "config" "utils" "core/managers" "core/security" "core/validation" "install"; do
        run_package_tests "$package"
    done
fi

# ============================================================================
# 성능 테스트 (새로운 기능)
# ============================================================================

if [[ $PERFORMANCE -eq 1 ]]; then
    print_section "Performance Tests"

    if [[ -d "tests/performance" ]]; then
        print_info "Running import performance tests..."
        PERF_OUTPUT=$(measure_time $PYTHON_CMD tests/performance/test_import_performance.py 2>&1)
        echo "$PERF_OUTPUT"

        if echo "$PERF_OUTPUT" | grep -q "PERFORMANCE REGRESSION"; then
            print_error "Performance regression detected"
            ALL_SUCCESS=false
        else
            print_success "Import performance tests passed"
        fi

        print_info "Checking for circular dependencies..."
        CYCLE_OUTPUT=$(measure_time $PYTHON_CMD tests/performance/test_dependency_cycles.py 2>&1)
        echo "$CYCLE_OUTPUT"

        if echo "$CYCLE_OUTPUT" | grep -q "CIRCULAR DEPENDENCY"; then
            print_error "Circular dependencies detected"
            ALL_SUCCESS=false
        else
            print_success "No circular dependencies found"
        fi
    else
        print_warning "Performance tests directory not found"
    fi
fi

# ============================================================================
# 메모리 사용량 모니터링 (새로운 기능)
# ============================================================================

if [[ $MEMORY_CHECK -eq 1 ]]; then
    print_section "Memory Usage Analysis"

    if command -v ps >/dev/null 2>&1; then
        # 테스트 시작 전 메모리 사용량
        MEMORY_BEFORE=$(ps -o pid,rss,vsz,comm -p $$ | tail -1 | awk '{print $2}')

        print_info "Memory usage monitoring enabled"
        print_info "Initial memory usage: ${MEMORY_BEFORE}KB"

        # 메모리 집약적 테스트 실행
        if [[ -f "tests/performance/test_memory_usage.py" ]]; then
            MEMORY_OUTPUT=$(measure_time $PYTHON_CMD tests/performance/test_memory_usage.py 2>&1)
            echo "$MEMORY_OUTPUT"

            if echo "$MEMORY_OUTPUT" | grep -q "MEMORY EXCEEDED"; then
                print_error "Memory usage exceeded limits"
                ALL_SUCCESS=false
            else
                print_success "Memory usage within acceptable limits"
            fi
        fi

        # 테스트 후 메모리 사용량
        MEMORY_AFTER=$(ps -o pid,rss,vsz,comm -p $$ | tail -1 | awk '{print $2}')
        MEMORY_DIFF=$((MEMORY_AFTER - MEMORY_BEFORE))

        print_info "Final memory usage: ${MEMORY_AFTER}KB"
        print_info "Memory increase: ${MEMORY_DIFF}KB"

        if [[ $MEMORY_DIFF -gt 102400 ]]; then  # 100MB threshold
            print_warning "High memory usage increase: ${MEMORY_DIFF}KB"
        fi
    else
        print_warning "Memory monitoring not available (ps command not found)"
    fi
fi

# ============================================================================
# 통합 테스트 (기존 + 새로운 패키지 통합)
# ============================================================================

print_section "Integration Tests"
if [[ -d "tests/integration" ]]; then
    print_info "Running package integration tests..."

    if [[ $VERBOSE -eq 1 ]]; then
        INTEGRATION_OUTPUT=$(measure_time $PYTEST_CMD tests/integration/ -v --tb=short 2>&1)
    else
        INTEGRATION_OUTPUT=$(measure_time $PYTEST_CMD tests/integration/ --tb=line 2>&1)
    fi

    echo "$INTEGRATION_OUTPUT"

    if echo "$INTEGRATION_OUTPUT" | grep -q "failed\\|error"; then
        print_error "Integration tests failed"
        ALL_SUCCESS=false
    else
        print_success "Integration tests passed"
    fi

    extract_test_stats "$INTEGRATION_OUTPUT"
else
    print_warning "Integration tests directory not found"
fi

# ============================================================================
# 레거시 테스트 (기존 시스템 호환성)
# ============================================================================

if [[ -z "$PACKAGE_ONLY" ]]; then
    print_section "Legacy Test Compatibility"

    # Hook 시스템 테스트 (기존 유지)
    if [[ -f "tests/test_hooks.py" ]]; then
        print_info "Running Hook system tests..."

        if [[ $VERBOSE -eq 1 ]]; then
            HOOK_OUTPUT=$(measure_time $COVERAGE_CMD tests/test_hooks.py 2>&1)
        else
            HOOK_OUTPUT=$(measure_time $COVERAGE_CMD tests/test_hooks.py 2>&1 | grep -E "(Tests run|Failures|Errors|Skipped|✅|❌)")
        fi

        echo "$HOOK_OUTPUT"

        if echo "$HOOK_OUTPUT" | grep -q "All tests passed!"; then
            print_success "Hook tests passed"
        else
            print_error "Hook tests failed"
            ALL_SUCCESS=false
        fi
    fi

    # 빌드 시스템 테스트 (기존 유지)
    if [[ -f "tests/test_build.py" ]]; then
        print_info "Running Build system tests..."

        if [[ $VERBOSE -eq 1 ]]; then
            BUILD_OUTPUT=$(measure_time $COVERAGE_CMD tests/test_build.py 2>&1)
        else
            BUILD_OUTPUT=$(measure_time $COVERAGE_CMD tests/test_build.py 2>&1 | grep -E "(Tests run|Failures|Errors|Skipped|✅|❌)")
        fi

        echo "$BUILD_OUTPUT"

        if echo "$BUILD_OUTPUT" | grep -q "All build tests passed!"; then
            print_success "Build tests passed"
        else
            print_error "Build tests failed"
            ALL_SUCCESS=false
        fi
    fi
fi

# ============================================================================
# 보고서 생성 및 아티팩트
# ============================================================================

# Coverage 리포트 생성 (개선)
if [[ $COVERAGE_AVAILABLE -eq 1 && $COVERAGE -eq 1 ]]; then
    print_section "Coverage Analysis"

    echo "Generating comprehensive coverage report..."
    $PYTHON_CMD -m coverage combine 2>/dev/null || true

    # 터미널 출력
    $PYTHON_CMD -m coverage report --show-missing --skip-covered

    # HTML 리포트 생성
    $PYTHON_CMD -m coverage html -d coverage_html 2>/dev/null && \\
        print_success "HTML coverage report generated: coverage_html/index.html"

    # XML 리포트 생성 (CI용)
    $PYTHON_CMD -m coverage xml -o coverage.xml 2>/dev/null && \\
        print_success "XML coverage report generated: coverage.xml"

    # Coverage 임계값 검사
    COVERAGE_PERCENTAGE=$($PYTHON_CMD -m coverage report | tail -1 | grep -o '[0-9]*%' | grep -o '[0-9]*')
    if [[ $COVERAGE_PERCENTAGE -lt 85 ]]; then
        print_warning "Coverage ${COVERAGE_PERCENTAGE}% below recommended 85%"
    else
        print_success "Coverage ${COVERAGE_PERCENTAGE}% meets quality standards"
    fi
fi

# JUnit XML 리포트 생성 (개선)
if [[ $JUNIT -eq 1 ]]; then
    print_section "JUnit XML Reports"

    mkdir -p test-reports

    # 통합 JUnit 보고서
    if $PYTHON_CMD -c "import pytest" 2>/dev/null; then
        print_info "Generating comprehensive JUnit XML report..."
        $PYTEST_CMD --junit-xml=test-reports/junit-all.xml tests/ --tb=short
        print_success "JUnit XML report generated: test-reports/junit-all.xml"
    else
        print_warning "pytest not available for JUnit reports"
    fi
fi

# 성능 보고서 생성
if [[ $PERFORMANCE -eq 1 ]]; then
    print_section "Performance Report"

    mkdir -p performance-reports

    # 성능 데이터 수집 및 JSON 저장
    cat > performance-reports/performance-summary.json << EOF
{
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "test_execution_summary": {
        "total_tests": $TOTAL_TESTS,
        "total_failures": $TOTAL_FAILURES,
        "total_errors": $TOTAL_ERRORS,
        "total_skipped": $TOTAL_SKIPPED
    },
    "performance_metrics": {
        "import_performance": "$(echo "${PERFORMANCE_RESULTS[@]}")",
        "memory_usage_before": "${MEMORY_BEFORE:-0}",
        "memory_usage_after": "${MEMORY_AFTER:-0}",
        "memory_increase": "${MEMORY_DIFF:-0}"
    }
}
EOF

    print_success "Performance report generated: performance-reports/performance-summary.json"
fi

# ============================================================================
# 최종 결과 요약 (향상된)
# ============================================================================

print_section "Enhanced Test Summary"

echo "📊 Test Execution Statistics:"
echo "   Total Tests: $TOTAL_TESTS"
echo "   Failures: $TOTAL_FAILURES"
echo "   Errors: $TOTAL_ERRORS"
echo "   Skipped: $TOTAL_SKIPPED"

if [[ $TOTAL_TESTS -gt 0 ]]; then
    SUCCESS_RATE=$(( (TOTAL_TESTS - TOTAL_FAILURES - TOTAL_ERRORS) * 100 / TOTAL_TESTS ))
    echo "   Success Rate: ${SUCCESS_RATE}%"
fi

# 품질 게이트 검사
QUALITY_ISSUES=0

if [[ $TOTAL_FAILURES -gt 0 ]]; then
    print_error "$TOTAL_FAILURES test failures detected"
    QUALITY_ISSUES=$((QUALITY_ISSUES + 1))
fi

if [[ $TOTAL_ERRORS -gt 0 ]]; then
    print_error "$TOTAL_ERRORS test errors detected"
    QUALITY_ISSUES=$((QUALITY_ISSUES + 1))
fi

if [[ $COVERAGE_AVAILABLE -eq 1 && $COVERAGE -eq 1 && $COVERAGE_PERCENTAGE -lt 85 ]]; then
    print_warning "Code coverage below 85%"
    QUALITY_ISSUES=$((QUALITY_ISSUES + 1))
fi

# 최종 판단
if [[ $ALL_SUCCESS == true && $QUALITY_ISSUES -eq 0 ]]; then
    print_success "All tests passed! Quality gate: PASSED 🎉"
    echo "🚀 Ready for deployment!"
    exit 0
else
    print_error "Quality gate: FAILED 💥"
    echo "🔧 Issues found: $QUALITY_ISSUES"
    echo "📋 Review test results and fix issues before deployment"
    exit 1
fi
'''

    def generate_quality_gate_script(self) -> str:
        """Generate quality gate validation script."""
        return '''#!/usr/bin/env python3
"""
Quality Gate Check Script for MoAI-ADK
Validates code quality metrics and test results
"""

import json
import sys
import xml.etree.ElementTree as ET
from pathlib import Path
from typing import Dict, Any, Tuple

class QualityGateChecker:
    """Checks various quality metrics against defined thresholds."""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.thresholds = {
            'coverage_minimum': 85.0,
            'test_success_rate_minimum': 95.0,
            'max_test_duration': 300.0,  # seconds
            'max_memory_increase': 102400,  # KB (100MB)
            'max_import_time': 0.5,  # seconds
        }
        self.quality_score = 100.0

    def check_coverage(self) -> Tuple[bool, str]:
        """Check code coverage meets minimum threshold."""
        coverage_file = self.project_root / "coverage.xml"

        if not coverage_file.exists():
            return False, "Coverage report not found"

        try:
            tree = ET.parse(coverage_file)
            root = tree.getroot()

            coverage_attr = root.get('line-rate')
            if coverage_attr:
                coverage_pct = float(coverage_attr) * 100

                if coverage_pct >= self.thresholds['coverage_minimum']:
                    return True, f"Coverage: {coverage_pct:.1f}% ✅"
                else:
                    self.quality_score -= 20
                    return False, f"Coverage: {coverage_pct:.1f}% (below {self.thresholds['coverage_minimum']}%)"
            else:
                return False, "Could not parse coverage percentage"

        except Exception as e:
            return False, f"Error reading coverage report: {e}"

    def check_test_results(self) -> Tuple[bool, str]:
        """Check test execution results."""
        junit_file = self.project_root / "test-reports" / "junit-all.xml"

        if not junit_file.exists():
            return False, "JUnit report not found"

        try:
            tree = ET.parse(junit_file)
            root = tree.getroot()

            total_tests = int(root.get('tests', 0))
            failures = int(root.get('failures', 0))
            errors = int(root.get('errors', 0))

            if total_tests == 0:
                return False, "No tests executed"

            success_rate = ((total_tests - failures - errors) / total_tests) * 100

            if success_rate >= self.thresholds['test_success_rate_minimum']:
                return True, f"Test Success Rate: {success_rate:.1f}% ✅"
            else:
                self.quality_score -= 30
                return False, f"Test Success Rate: {success_rate:.1f}% (below {self.thresholds['test_success_rate_minimum']}%)"

        except Exception as e:
            return False, f"Error reading test report: {e}"

    def check_performance_metrics(self) -> Tuple[bool, str]:
        """Check performance benchmarks."""
        perf_file = self.project_root / "performance-reports" / "performance-summary.json"

        if not perf_file.exists():
            return True, "Performance report not available (skipping)"

        try:
            with open(perf_file, 'r') as f:
                perf_data = json.load(f)

            issues = []

            # Check memory usage
            memory_increase = int(perf_data.get('performance_metrics', {}).get('memory_increase', 0))
            if memory_increase > self.thresholds['max_memory_increase']:
                issues.append(f"Memory increase: {memory_increase}KB (above {self.thresholds['max_memory_increase']}KB)")
                self.quality_score -= 10

            # Add more performance checks as needed

            if issues:
                return False, "Performance issues: " + ", ".join(issues)
            else:
                return True, "Performance metrics within acceptable ranges ✅"

        except Exception as e:
            return False, f"Error reading performance report: {e}"

    def check_import_structure(self) -> Tuple[bool, str]:
        """Check for import structure issues."""
        # This would typically run the dependency cycle checker
        try:
            import subprocess
            result = subprocess.run([
                sys.executable,
                str(self.project_root / "tests" / "performance" / "test_dependency_cycles.py")
            ], capture_output=True, text=True, timeout=30)

            if result.returncode == 0:
                return True, "Import structure validation passed ✅"
            else:
                self.quality_score -= 15
                return False, f"Import structure issues detected: {result.stderr}"

        except subprocess.TimeoutExpired:
            return False, "Import structure check timed out"
        except Exception as e:
            return False, f"Error checking import structure: {e}"

    def run_quality_gate(self) -> Dict[str, Any]:
        """Run all quality gate checks."""
        results = {
            'overall_status': 'UNKNOWN',
            'quality_score': self.quality_score,
            'checks': {},
            'issues': [],
            'passed_checks': 0,
            'total_checks': 0
        }

        # Define all checks
        checks = [
            ('coverage', self.check_coverage),
            ('test_results', self.check_test_results),
            ('performance', self.check_performance_metrics),
            ('import_structure', self.check_import_structure),
        ]

        # Run each check
        for check_name, check_func in checks:
            results['total_checks'] += 1
            try:
                passed, message = check_func()
                results['checks'][check_name] = {
                    'status': 'PASSED' if passed else 'FAILED',
                    'message': message
                }

                if passed:
                    results['passed_checks'] += 1
                else:
                    results['issues'].append(f"{check_name}: {message}")

            except Exception as e:
                results['checks'][check_name] = {
                    'status': 'ERROR',
                    'message': f"Check failed with error: {e}"
                }
                results['issues'].append(f"{check_name}: Check error")

        # Determine overall status
        results['quality_score'] = self.quality_score

        if results['passed_checks'] == results['total_checks'] and self.quality_score >= 80:
            results['overall_status'] = 'PASSED'
        elif self.quality_score >= 60:
            results['overall_status'] = 'WARNING'
        else:
            results['overall_status'] = 'FAILED'

        return results

def main():
    """Main execution function."""
    project_root = Path(".")
    if len(sys.argv) > 1:
        project_root = Path(sys.argv[1])

    checker = QualityGateChecker(project_root)
    results = checker.run_quality_gate()

    # Print results
    print("🎯 Quality Gate Analysis Results")
    print("=" * 50)

    print(f"Overall Status: {results['overall_status']}")
    print(f"Quality Score: {results['quality_score']:.1f}/100")
    print(f"Checks Passed: {results['passed_checks']}/{results['total_checks']}")

    print("\\nDetailed Results:")
    print("-" * 30)

    for check_name, check_data in results['checks'].items():
        status_emoji = "✅" if check_data['status'] == 'PASSED' else "❌" if check_data['status'] == 'FAILED' else "⚠️"
        print(f"{status_emoji} {check_name}: {check_data['message']}")

    if results['issues']:
        print("\\n🚨 Issues Found:")
        for issue in results['issues']:
            print(f"   • {issue}")

    # Save results
    results_file = project_root / "quality-gate-results.json"
    with open(results_file, 'w') as f:
        json.dump(results, f, indent=2)

    print(f"\\n📄 Results saved to: {results_file}")

    # Exit with appropriate code
    if results['overall_status'] == 'PASSED':
        print("\\n🎉 Quality Gate: PASSED")
        sys.exit(0)
    elif results['overall_status'] == 'WARNING':
        print("\\n⚠️ Quality Gate: WARNING")
        sys.exit(0)  # Allow deployment with warnings
    else:
        print("\\n💥 Quality Gate: FAILED")
        sys.exit(1)

if __name__ == "__main__":
    main()
'''

def main():
    """Main execution function."""
    if len(sys.argv) != 2:
        print("Usage: python cicd_adaptation_strategy.py <project_root>")
        sys.exit(1)

    project_root = Path(sys.argv[1])
    generator = CICDAdaptationGenerator(project_root)

    print("🔧 Generating CI/CD Adaptation Scripts...")

    # Generate enhanced Makefile
    enhanced_makefile = generator.generate_enhanced_makefile()
    makefile_new_path = project_root / "Makefile.enhanced"

    with open(makefile_new_path, 'w') as f:
        f.write(enhanced_makefile)

    print(f"✅ Enhanced Makefile generated: {makefile_new_path}")

    # Generate enhanced test runner
    enhanced_test_runner = generator.generate_enhanced_test_runner()
    test_runner_new_path = project_root / "scripts" / "run-tests-enhanced.sh"
    test_runner_new_path.parent.mkdir(exist_ok=True)

    with open(test_runner_new_path, 'w') as f:
        f.write(enhanced_test_runner)

    # Make executable
    import stat
    test_runner_new_path.chmod(test_runner_new_path.stat().st_mode | stat.S_IEXEC)

    print(f"✅ Enhanced test runner generated: {test_runner_new_path}")

    # Generate quality gate script
    quality_gate_script = generator.generate_quality_gate_script()
    quality_gate_path = project_root / "scripts" / "quality_gate_check.py"

    with open(quality_gate_path, 'w') as f:
        f.write(quality_gate_script)

    quality_gate_path.chmod(quality_gate_path.stat().st_mode | stat.S_IEXEC)

    print(f"✅ Quality gate script generated: {quality_gate_path}")

    print("\\n🚀 CI/CD Adaptation Complete!")
    print("\\nGenerated Files:")
    print(f"   📄 {makefile_new_path}")
    print(f"   📄 {test_runner_new_path}")
    print(f"   📄 {quality_gate_path}")

    print("\\n📋 Next Steps:")
    print("1. Review generated files and customize as needed")
    print("2. Replace existing Makefile and test runner after validation")
    print("3. Update CI/CD pipeline configurations")
    print("4. Test the enhanced system with: make test-packages")

if __name__ == "__main__":
    main()