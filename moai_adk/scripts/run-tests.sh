#!/bin/bash
# MoAI-ADK Test Runner
# 모든 테스트를 실행하고 결과를 리포트합니다.

set -e

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 함수 정의
print_header() {
    echo -e "${BLUE}============================================${NC}"
    echo -e "${BLUE}🧪 MoAI-ADK Test Suite Runner${NC}"
    echo -e "${BLUE}============================================${NC}"
}

print_section() {
    echo -e "\n${YELLOW}📋 $1${NC}"
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

# 프로젝트 루트 디렉토리 확인
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# 기본 설정
PYTHON_CMD=${PYTHON_CMD:-python3}
VERBOSE=${VERBOSE:-0}
COVERAGE=${COVERAGE:-0}
JUNIT=${JUNIT:-0}

# 명령행 인수 처리
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
        --python)
            PYTHON_CMD="$2"
            shift 2
            ;;
        --help|-h)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  --verbose, -v    Verbose output"
            echo "  --coverage, -c   Run with coverage report"
            echo "  --junit, -j      Generate JUnit XML report"
            echo "  --python CMD     Python command to use (default: python3)"
            echo "  --help, -h       Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

print_header

# Python 버전 확인
print_section "Environment Check"
echo "Python: $($PYTHON_CMD --version)"
echo "Project Root: $PROJECT_ROOT"

# 필요한 Python 모듈 확인
echo "Checking Python modules..."
REQUIRED_MODULES=("unittest" "json" "pathlib" "hashlib" "tempfile")
for module in "${REQUIRED_MODULES[@]}"; do
    if $PYTHON_CMD -c "import $module" 2>/dev/null; then
        print_success "$module module available"
    else
        print_error "$module module not available"
        exit 1
    fi
done

# 선택적 모듈 확인
OPTIONAL_MODULES=("coverage")
for module in "${OPTIONAL_MODULES[@]}"; do
    if $PYTHON_CMD -c "import $module" 2>/dev/null; then
        print_success "$module module available"
    else
        print_warning "$module module not available (optional)"
    fi
done

# 테스트 결과 변수
TOTAL_TESTS=0
TOTAL_FAILURES=0
TOTAL_ERRORS=0
TOTAL_SKIPPED=0
ALL_SUCCESS=true

# Coverage 설정
if [[ $COVERAGE -eq 1 ]]; then
    if $PYTHON_CMD -c "import coverage" 2>/dev/null; then
        COVERAGE_CMD="$PYTHON_CMD -m coverage run --source=. --omit=tests/*,*/__pycache__/*"
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

# Hook 시스템 테스트
print_section "Hook System Tests"
if [[ -f "tests/test_hooks.py" ]]; then
    echo "Running Hook system tests..."
    
    if [[ $VERBOSE -eq 1 ]]; then
        TEST_OUTPUT=$($COVERAGE_CMD tests/test_hooks.py 2>&1)
    else
        TEST_OUTPUT=$($COVERAGE_CMD tests/test_hooks.py 2>&1 | grep -E "(Tests run|Failures|Errors|Skipped|✅|❌)")
    fi
    
    echo "$TEST_OUTPUT"
    
    # 결과 파싱
    if echo "$TEST_OUTPUT" | grep -q "All tests passed!"; then
        print_success "Hook tests passed"
    else
        print_error "Hook tests failed"
        ALL_SUCCESS=false
    fi
    
    # 통계 추출
    if echo "$TEST_OUTPUT" | grep -q "Tests run:"; then
        HOOK_TESTS=$(echo "$TEST_OUTPUT" | grep "Tests run:" | sed 's/.*Tests run: \([0-9]*\).*/\1/')
        HOOK_FAILURES=$(echo "$TEST_OUTPUT" | grep "Failures:" | sed 's/.*Failures: \([0-9]*\).*/\1/')
        HOOK_ERRORS=$(echo "$TEST_OUTPUT" | grep "Errors:" | sed 's/.*Errors: \([0-9]*\).*/\1/')
        HOOK_SKIPPED=$(echo "$TEST_OUTPUT" | grep "Skipped:" | sed 's/.*Skipped: \([0-9]*\).*/\1/')
        
        TOTAL_TESTS=$((TOTAL_TESTS + HOOK_TESTS))
        TOTAL_FAILURES=$((TOTAL_FAILURES + HOOK_FAILURES))
        TOTAL_ERRORS=$((TOTAL_ERRORS + HOOK_ERRORS))
        TOTAL_SKIPPED=$((TOTAL_SKIPPED + HOOK_SKIPPED))
    fi
else
    print_warning "Hook tests not found (tests/test_hooks.py)"
fi

# 빌드 시스템 테스트
print_section "Build System Tests"
if [[ -f "tests/test_build.py" ]]; then
    echo "Running Build system tests..."
    
    if [[ $VERBOSE -eq 1 ]]; then
        BUILD_OUTPUT=$($COVERAGE_CMD tests/test_build.py 2>&1)
    else
        BUILD_OUTPUT=$($COVERAGE_CMD tests/test_build.py 2>&1 | grep -E "(Tests run|Failures|Errors|Skipped|✅|❌)")
    fi
    
    echo "$BUILD_OUTPUT"
    
    # 결과 파싱
    if echo "$BUILD_OUTPUT" | grep -q "All build tests passed!"; then
        print_success "Build tests passed"
    else
        print_error "Build tests failed"
        ALL_SUCCESS=false
    fi
    
    # 통계 추출
    if echo "$BUILD_OUTPUT" | grep -q "Tests run:"; then
        BUILD_TESTS=$(echo "$BUILD_OUTPUT" | grep "Tests run:" | sed 's/.*Tests run: \([0-9]*\).*/\1/')
        BUILD_FAILURES=$(echo "$BUILD_OUTPUT" | grep "Failures:" | sed 's/.*Failures: \([0-9]*\).*/\1/')
        BUILD_ERRORS=$(echo "$BUILD_OUTPUT" | grep "Errors:" | sed 's/.*Errors: \([0-9]*\).*/\1/')
        BUILD_SKIPPED=$(echo "$BUILD_OUTPUT" | grep "Skipped:" | sed 's/.*Skipped: \([0-9]*\).*/\1/')
        
        TOTAL_TESTS=$((TOTAL_TESTS + BUILD_TESTS))
        TOTAL_FAILURES=$((TOTAL_FAILURES + BUILD_FAILURES))
        TOTAL_ERRORS=$((TOTAL_ERRORS + BUILD_ERRORS))
        TOTAL_SKIPPED=$((TOTAL_SKIPPED + BUILD_SKIPPED))
    fi
else
    print_warning "Build tests not found (tests/test_build.py)"
fi

# 설정 검증 테스트
print_section "Configuration Validation"
echo "Validating JSON configurations..."

if [[ -f "src/templates/.claude/settings.json" ]]; then
    if $PYTHON_CMD -c "import json; json.load(open('src/templates/.claude/settings.json'))" 2>/dev/null; then
        print_success "Claude settings.json is valid"
    else
        print_error "Claude settings.json is invalid"
        ALL_SUCCESS=false
    fi
else
    print_warning "Claude settings.json not found"
fi

if [[ -f "src/templates/.moai/config.json" ]]; then
    if $PYTHON_CMD -c "import json; json.load(open('src/templates/.moai/config.json'))" 2>/dev/null; then
        print_success "MoAI config.json is valid"
    else
        print_error "MoAI config.json is invalid"
        ALL_SUCCESS=false
    fi
else
    print_warning "MoAI config.json not found"
fi

# 빌드 시스템 기능 테스트
print_section "Build System Integration"
if [[ -f "build.py" ]]; then
    echo "Testing build system status..."
    
    if $PYTHON_CMD build.py status >/dev/null 2>&1; then
        print_success "Build system status check passed"
    else
        print_error "Build system status check failed"
        ALL_SUCCESS=false
    fi
else
    print_error "build.py not found"
    ALL_SUCCESS=false
fi

# Hook 스크립트 실행 권한 확인
print_section "Hook Scripts Permissions"
HOOK_SCRIPTS_DIR="src/templates/.claude/hooks/moai"
if [[ -d "$HOOK_SCRIPTS_DIR" ]]; then
    echo "Checking hook script permissions..."
    
    HOOK_SCRIPTS=($(find "$HOOK_SCRIPTS_DIR" -name "*.py" -type f))
    for script in "${HOOK_SCRIPTS[@]}"; do
        if [[ -x "$script" ]]; then
            print_success "$(basename "$script") has execute permission"
        else
            print_warning "$(basename "$script") missing execute permission"
        fi
    done
else
    print_warning "Hook scripts directory not found"
fi

# Coverage 리포트 생성
if [[ $COVERAGE_AVAILABLE -eq 1 && $COVERAGE -eq 1 ]]; then
    print_section "Coverage Report"
    
    echo "Generating coverage report..."
    $PYTHON_CMD -m coverage combine 2>/dev/null || true
    $PYTHON_CMD -m coverage report --show-missing
    
    # HTML 리포트 생성
    $PYTHON_CMD -m coverage html -d coverage_html 2>/dev/null && \
        print_success "HTML coverage report generated: coverage_html/index.html"
fi

# JUnit XML 리포트 생성
if [[ $JUNIT -eq 1 ]]; then
    print_section "JUnit XML Report"
    
    # unittest-xml-reporting이 있는지 확인
    if $PYTHON_CMD -c "import xmlrunner" 2>/dev/null; then
        echo "Generating JUnit XML report..."
        mkdir -p test-reports
        
        # Hook 테스트
        if [[ -f "tests/test_hooks.py" ]]; then
            $PYTHON_CMD -m xmlrunner discover -s tests -p "test_hooks.py" -o test-reports/hooks
        fi
        
        # Build 테스트
        if [[ -f "tests/test_build.py" ]]; then
            $PYTHON_CMD -m xmlrunner discover -s tests -p "test_build.py" -o test-reports/build
        fi
        
        print_success "JUnit XML reports generated in test-reports/"
    else
        print_warning "unittest-xml-reporting not available for JUnit reports"
        echo "Install with: pip install unittest-xml-reporting"
    fi
fi

# 최종 결과 요약
print_section "Test Summary"
echo "Total Tests: $TOTAL_TESTS"
echo "Failures: $TOTAL_FAILURES"
echo "Errors: $TOTAL_ERRORS" 
echo "Skipped: $TOTAL_SKIPPED"

if [[ $ALL_SUCCESS == true ]]; then
    print_success "All tests passed! 🎉"
    exit 0
else
    print_error "Some tests failed! 💥"
    exit 1
fi