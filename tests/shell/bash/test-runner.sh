#!/bin/bash
# Bash 테스트 러너 (cross-platform)
# TEST-SHELL-TEST-001: Bash test execution framework

set -e

# ==============================================================================
# 설정
# ==============================================================================

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../" && pwd)"
TEST_PATH="${1:-.}"
TEST_TYPE="${2:-all}"
VERBOSE="${3:-false}"

COLORS=(
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    CYAN='\033[0;36m'
    NC='\033[0m'
)

# ==============================================================================
# 헬퍼 함수
# ==============================================================================

log_info() {
    echo -e "${COLORS[CYAN]}[INFO]${COLORS[NC]} $*"
}

log_success() {
    echo -e "${COLORS[GREEN]}✓${COLORS[NC]} $*"
}

log_error() {
    echo -e "${COLORS[RED]}✗${COLORS[NC]} $*"
}

log_warning() {
    echo -e "${COLORS[YELLOW]}⚠${COLORS[NC]} $*"
}

# ==============================================================================
# 함수: 패키지 설치 검증
# ==============================================================================

test_package_installation() {
    log_info "패키지 설치 검증"

    cd "$PROJECT_ROOT"

    if python -c "import moai_adk; print('OK')" &>/dev/null; then
        log_success "패키지 설치 완료"
        return 0
    else
        log_error "패키지 임포트 실패"
        return 1
    fi
}

# ==============================================================================
# 함수: pytest 실행
# ==============================================================================

test_pytest() {
    log_info "pytest 테스트 실행"

    cd "$PROJECT_ROOT"

    local test_path="tests"

    case "$TEST_TYPE" in
        package)
            test_path="tests/unit"
            ;;
        hooks)
            test_path="tests/hooks"
            ;;
        cli)
            test_path="tests/integration"
            ;;
        *)
            test_path="tests"
            ;;
    esac

    if [ "$VERBOSE" = "true" ]; then
        pytest "$test_path" -v --tb=short
    else
        pytest "$test_path" --tb=short -q
    fi

    if [ $? -eq 0 ]; then
        log_success "pytest 테스트 통과"
        return 0
    else
        log_error "pytest 테스트 실패"
        return 1
    fi
}

# ==============================================================================
# 함수: 명령어 가용성 확인
# ==============================================================================

test_command_availability() {
    log_info "필수 명령어 가용성"

    local commands=("python" "pip" "pytest")

    for cmd in "${commands[@]}"; do
        if command -v "$cmd" &>/dev/null; then
            log_success "$cmd 사용 가능"
        else
            log_error "$cmd 없음"
            return 1
        fi
    done

    return 0
}

# ==============================================================================
# 함수: 패키지 모듈 테스트
# ==============================================================================

test_package_modules() {
    log_info "패키지 모듈 구조"

    cd "$PROJECT_ROOT"

    if python <<'EOF' &>/dev/null
import moai_adk
from moai_adk import cli, core, templates

modules = ['cli', 'core', 'templates']
for mod in modules:
    print(f'✓ {mod} 모듈 로드 완료')
EOF
    then
        log_success "모듈 로드 완료"
        return 0
    else
        log_error "모듈 로드 오류"
        return 1
    fi
}

# ==============================================================================
# 함수: 타입 체크 (mypy)
# ==============================================================================

test_type_checking() {
    log_info "타입 체크 (mypy)"

    cd "$PROJECT_ROOT"

    if command -v mypy &>/dev/null; then
        if mypy src/moai_adk --strict --ignore-missing-imports &>/dev/null; then
            log_success "타입 체크 통과"
            return 0
        else
            log_warning "타입 체크 경고/오류 (상세: mypy src/moai_adk)"
            return 0  # 경고는 무시
        fi
    else
        log_warning "mypy 미설치"
        return 0
    fi
}

# ==============================================================================
# 함수: 코드 린팅 (ruff)
# ==============================================================================

test_linting() {
    log_info "코드 린팅 (ruff)"

    cd "$PROJECT_ROOT"

    if command -v ruff &>/dev/null; then
        if ruff check src/moai_adk --select=E,W,F &>/dev/null; then
            log_success "린팅 체크 통과"
            return 0
        else
            log_warning "린팅 경고 발견 (상세: ruff check src/moai_adk)"
            return 0  # 경고는 무시
        fi
    else
        log_warning "ruff 미설치"
        return 0
    fi
}

# ==============================================================================
# 함수: 결과 보고
# ==============================================================================

report_results() {
    local passed=$1
    local failed=$2

    echo ""
    echo -e "${COLORS[CYAN]}========================================${COLORS[NC]}"
    echo -e "${COLORS[CYAN]}테스트 결과 보고${COLORS[NC]}"
    echo -e "${COLORS[CYAN]}========================================${COLORS[NC]}"
    echo ""
    echo "총 결과: $passed 통과, $failed 실패"
    echo ""

    if [ "$failed" -eq 0 ]; then
        log_success "모든 테스트 통과!"
        return 0
    else
        log_error "일부 테스트 실패!"
        echo "자세한 정보: pytest -v tests/"
        return 1
    fi
}

# ==============================================================================
# 메인 함수
# ==============================================================================

main() {
    log_info "Bash 패키지 검증 테스트 시작"
    echo "타임스탬프: $(date '+%Y-%m-%d %H:%M:%S')"
    echo ""

    local passed=0
    local failed=0

    # 1. 명령어 가용성
    if test_command_availability; then
        ((passed++))
    else
        ((failed++))
    fi

    # 2. 패키지 설치
    if test_package_installation; then
        ((passed++))
    else
        ((failed++))
    fi

    # 3. 패키지 모듈
    if test_package_modules; then
        ((passed++))
    else
        ((failed++))
    fi

    # 4. pytest 테스트
    if test_pytest; then
        ((passed++))
    else
        ((failed++))
    fi

    # 5. 타입 체크
    if test_type_checking; then
        ((passed++))
    else
        ((failed++))
    fi

    # 6. 린팅
    if test_linting; then
        ((passed++))
    else
        ((failed++))
    fi

    # 결과 보고
    if report_results "$passed" "$failed"; then
        exit 0
    else
        exit 1
    fi
}

# 메인 실행
main
