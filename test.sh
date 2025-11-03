#!/bin/bash
# MoAI-ADK 통합 테스트 러너 (Bash + PowerShell 멀티셸 지원)
# @TAG:INTEGRATION-TEST-001 | Multi-shell test orchestration framework
# 사용법:
#   ./test.sh              # 기본 테스트 (사용 가능한 모든 셸)
#   ./test.sh bash         # Bash만 테스트
#   ./test.sh powershell   # PowerShell만 테스트
#   ./test.sh all -v       # 모든 셸 + 상세 로그
#   ./test.sh package      # 패키지 테스트만

set -e

# ==============================================================================
# 색상 정의
# ==============================================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

# ==============================================================================
# 함수 정의
# ==============================================================================

print_header() {
    echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# ==============================================================================
# 셸 가용성 확인
# ==============================================================================

check_shell_availability() {
    local shell=$1

    case "$shell" in
        bash)
            if command -v bash &>/dev/null; then
                print_success "Bash 사용 가능 ($(bash --version | head -1 | cut -d' ' -f3-4))"
                return 0
            else
                print_warning "Bash 미설치"
                return 1
            fi
            ;;
        powershell)
            if command -v pwsh &>/dev/null; then
                print_success "PowerShell 사용 가능 ($(pwsh --version 2>/dev/null | head -1 | cut -d' ' -f3))"
                return 0
            else
                print_warning "PowerShell 미설치"
                return 1
            fi
            ;;
        *)
            print_error "알 수 없는 셸: $shell"
            return 1
            ;;
    esac
}

# ==============================================================================
# Bash 테스트 실행
# ==============================================================================

run_bash_tests() {
    local test_type=${1:-all}
    local verbose=${2:-false}

    print_header "Bash 테스트 실행 [$test_type]"

    local script_path="tests/shell/bash/test-runner.sh"

    if [ ! -f "$script_path" ]; then
        print_error "테스트 스크립트 없음: $script_path"
        return 1
    fi

    if [ "$verbose" = "true" ]; then
        bash "$script_path" "tests" "$test_type" "true"
    else
        bash "$script_path" "tests" "$test_type" "false"
    fi

    return $?
}

# ==============================================================================
# PowerShell 테스트 실행
# ==============================================================================

run_powershell_tests() {
    local test_type=${1:-all}
    local verbose=${2:-false}

    print_header "PowerShell 테스트 실행 [$test_type]"

    local script_path="tests/shell/powershell/helpers/runner.ps1"

    if [ ! -f "$script_path" ]; then
        print_error "PowerShell 스크립트 없음: $script_path"
        return 1
    fi

    if [ "$verbose" = "true" ]; then
        pwsh -NoProfile -File "$script_path" -TestType "$test_type" -ShowDetails
    else
        pwsh -NoProfile -File "$script_path" -TestType "$test_type"
    fi

    return $?
}

# ==============================================================================
# 메인 로직
# ==============================================================================

main() {
    local shell_choice="${1:-all}"
    local test_type="all"
    local verbose="false"

    # 인자 파싱
    for arg in "$@"; do
        case "$arg" in
            -v | --verbose)
                verbose="true"
                ;;
            all | bash | powershell)
                shell_choice="$arg"
                ;;
            unit | package | hooks | cli | integration)
                test_type="$arg"
                ;;
        esac
    done

    print_header "MoAI-ADK 멀티셸 테스트 시작"
    echo "타임스탬프: $(date '+%Y-%m-%d %H:%M:%S')"
    echo "선택된 셸: $shell_choice"
    echo "테스트 타입: $test_type"
    if [ "$verbose" = "true" ]; then
        echo "상세 로그: 활성화"
    fi
    echo ""

    local bash_result=0
    local powershell_result=0
    local total_passed=0
    local total_failed=0

    # ==============================================================================
    # Bash 테스트
    # ==============================================================================

    if [ "$shell_choice" = "all" ] || [ "$shell_choice" = "bash" ]; then
        if check_shell_availability bash; then
            echo ""
            if run_bash_tests "$test_type" "$verbose"; then
                print_success "Bash 테스트 완료"
                bash_result=0
                ((total_passed++))
            else
                print_error "Bash 테스트 실패"
                bash_result=1
                ((total_failed++))
            fi
        else
            print_info "Bash 테스트 건너뜀 (미설치)"
        fi
    fi

    # ==============================================================================
    # PowerShell 테스트
    # ==============================================================================

    if [ "$shell_choice" = "all" ] || [ "$shell_choice" = "powershell" ]; then
        if check_shell_availability powershell; then
            echo ""
            if run_powershell_tests "$test_type" "$verbose"; then
                print_success "PowerShell 테스트 완료"
                powershell_result=0
                ((total_passed++))
            else
                print_error "PowerShell 테스트 실패"
                powershell_result=1
                ((total_failed++))
            fi
        else
            print_info "PowerShell 테스트 건너뜀 (미설치)"
        fi
    fi

    # ==============================================================================
    # 최종 결과
    # ==============================================================================

    echo ""
    print_header "최종 테스트 결과"

    if [ "$bash_result" -eq 0 ] && [ "$powershell_result" -eq 0 ]; then
        echo "모든 선택된 셸 테스트가 성공했습니다! ✓"
        echo ""
        exit 0
    else
        if [ "$bash_result" -ne 0 ]; then
            print_error "Bash 테스트 실패"
        fi
        if [ "$powershell_result" -ne 0 ]; then
            print_error "PowerShell 테스트 실패"
        fi
        echo ""
        exit 1
    fi
}

# 메인 실행
main "$@"
