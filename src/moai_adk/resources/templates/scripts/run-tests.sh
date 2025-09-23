#!/bin/bash
# MoAI-ADK 통합 테스트 실행 스크립트 v0.1.12
#
# 모든 테스트와 검증을 자동화하여 실행합니다:
# - Python 단위 테스트 (pytest)
# - 통합 테스트 및 E2E 테스트
# - 개발 가이드 5원칙 검증
# - 16-Core TAG 시스템 검증
# - 라이선스 및 보안 검사
# - 코드 커버리지 측정
#
# 사용법:
#     ./scripts/run-tests.sh [옵션]
#     
# 옵션:
#     --unit          단위 테스트만 실행
#     --integration   통합 테스트만 실행
#     --coverage      커버리지 리포트 생성
#     --full          전체 검증 실행 (기본값)
#     --fast          빠른 검증 (중요한 것만)
#     --fix           자동 수정 가능한 이슈 해결
#     --verbose       상세 출력
#     --help          도움말 표시

set -e  # 오류 시 즉시 종료
set -u  # 정의되지 않은 변수 사용 시 오류

# 색상 코드 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 기본 설정
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TEST_MODE="full"
VERBOSE=false
AUTO_FIX=false
COVERAGE_ENABLED=false

# 로그 함수들
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_header() {
    echo -e "${PURPLE}🎯 $1${NC}"
    echo "$(echo "$1" | sed 's/./=/g')"
}

# 진행률 표시
show_progress() {
    local current=$1
    local total=$2
    local message=$3
    local percent=$((current * 100 / total))
    local filled=$((percent / 5))
    local empty=$((20 - filled))
    
    printf "\r${CYAN}["
    printf "%*s" $filled | tr ' ' '='
    printf "%*s" $empty | tr ' ' '-'
    printf "] %d%% %s${NC}" $percent "$message"
    
    if [ $current -eq $total ]; then
        echo
    fi
}

# 명령어 파싱
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --unit)
                TEST_MODE="unit"
                shift
                ;;
            --integration)
                TEST_MODE="integration"
                shift
                ;;
            --coverage)
                COVERAGE_ENABLED=true
                shift
                ;;
            --full)
                TEST_MODE="full"
                shift
                ;;
            --fast)
                TEST_MODE="fast"
                shift
                ;;
            --fix)
                AUTO_FIX=true
                shift
                ;;
            --verbose|-v)
                VERBOSE=true
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                log_error "알 수 없는 옵션: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 도움말 표시
show_help() {
    echo "MoAI-ADK 통합 테스트 실행 스크립트 v0.1.12"
    echo
    echo "사용법: $0 [옵션]"
    echo
    echo "옵션:"
    echo "  --unit          단위 테스트만 실행"
    echo "  --integration   통합 테스트만 실행" 
    echo "  --coverage      커버리지 리포트 생성"
    echo "  --full          전체 검증 실행 (기본값)"
    echo "  --fast          빠른 검증 (중요한 것만)"
    echo "  --fix           자동 수정 가능한 이슈 해결"
    echo "  --verbose, -v   상세 출력"
    echo "  --help, -h      이 도움말 표시"
    echo
    echo "예시:"
    echo "  $0                    # 전체 검증"
    echo "  $0 --unit --coverage  # 단위 테스트 + 커버리지"
    echo "  $0 --fast --fix       # 빠른 검증 + 자동 수정"
}

# 환경 검증
verify_environment() {
    log_header "환경 검증"
    
    # Python 설치 확인
    if ! command -v python3 &> /dev/null; then
        log_error "Python 3가 설치되지 않았습니다"
        exit 1
    fi
    
    local python_version=$(python3 --version | cut -d' ' -f2)
    log_info "Python 버전: $python_version"
    
    # 필수 패키지 확인
    local required_packages=("pytest" "pytest-cov" "colorama" "click")
    for package in "${required_packages[@]}"; do
        if ! python3 -c "import $package" &> /dev/null; then
            log_warning "$package 패키지가 설치되지 않았습니다"
            log_info "패키지 설치 중: $package"
            pip3 install "$package" || {
                log_error "$package 설치 실패"
                exit 1
            }
        fi
    done
    
    log_success "환경 검증 완료"
    echo
}

# 프로젝트 구조 확인
verify_project_structure() {
    log_header "프로젝트 구조 검증"
    
    local required_dirs=(
        "src"
        "tests"
        "scripts"
        ".claude"
        ".moai"
    )
    
    local missing_dirs=()
    for dir in "${required_dirs[@]}"; do
        if [[ ! -d "$PROJECT_ROOT/$dir" ]]; then
            missing_dirs+=("$dir")
        fi
    done
    
    if [[ ${#missing_dirs[@]} -gt 0 ]]; then
        log_warning "누락된 디렉토리: ${missing_dirs[*]}"
    else
        log_success "프로젝트 구조 검증 완료"
    fi
    
    echo
}

# 단위 테스트 실행
run_unit_tests() {
    log_header "단위 테스트 실행"
    
    if [[ ! -d "$PROJECT_ROOT/tests" ]]; then
        log_warning "tests 디렉토리가 없습니다"
        return 0
    fi
    
    cd "$PROJECT_ROOT"
    
    local pytest_args=()
    if [[ "$VERBOSE" == true ]]; then
        pytest_args+=("-v")
    else
        pytest_args+=("-q")
    fi
    
    if [[ "$COVERAGE_ENABLED" == true ]]; then
        pytest_args+=("--cov=src" "--cov-report=term-missing" "--cov-report=html")
        log_info "커버리지 측정 활성화"
    fi
    
    # 단위 테스트 실행
    if python3 -m pytest tests/unit "${pytest_args[@]}" 2>/dev/null; then
        log_success "단위 테스트 통과"
    else
        # tests/ 전체 실행 (unit 폴더가 없을 경우)
        if python3 -m pytest tests/ "${pytest_args[@]}"; then
            log_success "테스트 통과"
        else
            log_error "단위 테스트 실패"
            return 1
        fi
    fi
    
    echo
}

# 통합 테스트 실행
run_integration_tests() {
    log_header "통합 테스트 실행"
    
    if [[ -d "$PROJECT_ROOT/tests/integration" ]]; then
        cd "$PROJECT_ROOT"
        if python3 -m pytest tests/integration ${VERBOSE:+-v}; then
            log_success "통합 테스트 통과"
        else
            log_error "통합 테스트 실패"
            return 1
        fi
    else
        log_info "통합 테스트 디렉토리 없음 - 스킵"
    fi
    
    echo
}

# 개발 가이드 검증
run_constitution_check() {
    log_header "개발 가이드 5원칙 검증"
    
    local constitution_script="$SCRIPT_DIR/check_constitution.py"
    if [[ -f "$constitution_script" ]]; then
        local args=()
        if [[ "$VERBOSE" == true ]]; then
            args+=("--verbose")
        fi
        if [[ "$AUTO_FIX" == true ]]; then
            args+=("--fix")
        fi
        
        if python3 "$constitution_script" "${args[@]}"; then
            log_success "개발 가이드 5원칙 준수"
        else
            log_error "개발 가이드 위반 사항 발견"
            return 1
        fi
    else
        log_warning "개발 가이드 검증 스크립트 없음"
    fi
    
    echo
}

# TAG 시스템 검증
run_tag_validation() {
    log_header "16-Core TAG 시스템 검증"
    
    local tag_script="$SCRIPT_DIR/validate_tags.py" 
    if [[ -f "$tag_script" ]]; then
        local args=()
        if [[ "$VERBOSE" == true ]]; then
            args+=("--verbose")
        fi
        if [[ "$AUTO_FIX" == true ]]; then
            args+=("--fix")
        fi
        
        if python3 "$tag_script" "${args[@]}"; then
            log_success "TAG 시스템 무결성 확인"
        else
            log_error "TAG 시스템 이슈 발견"
            return 1
        fi
    else
        log_warning "TAG 검증 스크립트 없음"
    fi
    
    echo
}

# 라이선스 검사
run_license_check() {
    log_header "라이선스 및 보안 검사"
    
    local license_script="$SCRIPT_DIR/check-licenses.py"
    if [[ -f "$license_script" ]]; then
        if python3 "$license_script" ${VERBOSE:+--verbose}; then
            log_success "라이선스 검사 통과"
        else
            log_error "라이선스 이슈 발견"
            return 1
        fi
    else
        log_warning "라이선스 검사 스크립트 없음"
    fi
    
    # 보안 검사
    local secrets_script="$SCRIPT_DIR/check-secrets.py"
    if [[ -f "$secrets_script" ]]; then
        if python3 "$secrets_script" ${VERBOSE:+--verbose}; then
            log_success "보안 검사 통과"
        else
            log_error "보안 이슈 발견"
            return 1
        fi
    else
        log_warning "보안 검사 스크립트 없음"
    fi
    
    echo
}

# 커버리지 검사
run_coverage_check() {
    log_header "테스트 커버리지 검사"
    
    local coverage_script="$SCRIPT_DIR/check_coverage.py"
    if [[ -f "$coverage_script" ]]; then
        if python3 "$coverage_script" ${VERBOSE:+--verbose}; then
            log_success "커버리지 목표 달성"
        else
            log_error "커버리지 목표 미달"
            return 1
        fi
    else
        log_warning "커버리지 검사 스크립트 없음"
    fi
    
    echo
}

# 추적성 검증
run_traceability_check() {
    log_header "추적성 매트릭스 검증"
    
    local trace_script="$SCRIPT_DIR/check-traceability.py"
    if [[ -f "$trace_script" ]]; then
        local args=()
        if [[ "$VERBOSE" == true ]]; then
            args+=("--verbose")
        fi
        
        if python3 "$trace_script" "${args[@]}"; then
            log_success "추적성 매트릭스 검증 완료"
        else
            log_error "추적성 이슈 발견"
            return 1
        fi
    else
        log_warning "추적성 검증 스크립트 없음"
    fi
    
    echo
}

# 빠른 검증 모드
run_fast_tests() {
    log_header "빠른 검증 모드"
    
    local total_tests=4
    local current=0
    
    # 환경 검증
    ((current++))
    show_progress $current $total_tests "환경 검증"
    verify_environment >/dev/null 2>&1
    
    # 개발 가이드 검사
    ((current++))
    show_progress $current $total_tests "개발 가이드 검증"
    if ! run_constitution_check >/dev/null 2>&1; then
        log_error "개발 가이드 검증 실패"
        return 1
    fi
    
    # 기본 테스트
    ((current++))
    show_progress $current $total_tests "기본 테스트 실행"
    if ! run_unit_tests >/dev/null 2>&1; then
        log_error "기본 테스트 실패"
        return 1
    fi
    
    # TAG 검증
    ((current++))
    show_progress $current $total_tests "TAG 시스템 검증"
    if ! run_tag_validation >/dev/null 2>&1; then
        log_error "TAG 검증 실패"
        return 1
    fi
    
    log_success "빠른 검증 완료"
    echo
}

# 전체 검증 모드
run_full_tests() {
    log_header "전체 검증 모드"
    
    local tests=(
        "verify_environment"
        "verify_project_structure"
        "run_unit_tests"
        "run_integration_tests"
        "run_constitution_check"
        "run_tag_validation"
        "run_license_check"
        "run_coverage_check"
        "run_traceability_check"
    )
    
    local total_tests=${#tests[@]}
    local current=0
    local failed_tests=()
    
    for test_func in "${tests[@]}"; do
        ((current++))
        show_progress $current $total_tests "실행 중: $test_func"
        
        if ! $test_func; then
            failed_tests+=("$test_func")
        fi
    done
    
    if [[ ${#failed_tests[@]} -gt 0 ]]; then
        log_error "실패한 테스트: ${failed_tests[*]}"
        return 1
    else
        log_success "전체 검증 완료"
    fi
    
    echo
}

# 결과 요약
show_summary() {
    local exit_code=$1
    
    log_header "검증 결과 요약"
    
    if [[ $exit_code -eq 0 ]]; then
        log_success "🎉 모든 검증이 성공적으로 완료되었습니다!"
        
        if [[ "$COVERAGE_ENABLED" == true ]]; then
            log_info "📊 커버리지 리포트: htmlcov/index.html"
        fi
        
        log_info "🚀 배포 준비가 완료되었습니다"
        echo
        echo "다음 단계:"
        echo "  1. git add . && git commit -m \"feat: 모든 테스트 통과\""
        echo "  2. python -m build"
        echo "  3. python -m twine upload --repository testpypi dist/*"
    else
        log_error "💥 일부 검증이 실패했습니다"
        echo
        echo "문제 해결 방법:"
        echo "  1. 상세 로그 확인: $0 --verbose"
        echo "  2. 자동 수정 시도: $0 --fix"
        echo "  3. 개별 테스트 실행: $0 --unit"
    fi
    
    echo
}

# 메인 실행 함수
main() {
    # 시작 시간 기록
    local start_time=$(date +%s)
    
    echo "🗿 MoAI-ADK 통합 테스트 시스템 v0.1.12"
    echo "============================================"
    echo
    
    # 명령어 파싱
    parse_arguments "$@"
    
    # 프로젝트 루트로 이동
    cd "$PROJECT_ROOT"
    
    # 테스트 모드에 따른 실행
    local exit_code=0
    
    case $TEST_MODE in
        "unit")
            verify_environment
            run_unit_tests || exit_code=1
            ;;
        "integration")
            verify_environment
            run_integration_tests || exit_code=1
            ;;
        "fast")
            run_fast_tests || exit_code=1
            ;;
        "full")
            run_full_tests || exit_code=1
            ;;
        *)
            log_error "알 수 없는 테스트 모드: $TEST_MODE"
            exit 1
            ;;
    esac
    
    # 실행 시간 계산
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    log_info "⏱️  실행 시간: ${duration}초"
    
    # 결과 요약
    show_summary $exit_code
    
    exit $exit_code
}

# 스크립트가 직접 실행될 때만 main 함수 호출
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
