#!/bin/bash
# release-helper.sh - 릴리스 관련 유틸리티 함수 라이브러리
#
# 역할:
#   - 공통 함수 모음 (로깅, 파일 작업, git 명령어)
#   - 다른 스크립트에서 source로 사용
#
# 사용법:
#   source .moai/release/release-helper.sh
#   log_info "메시지"
#   check_command python3

# ─────────────────────────────────────────────────────────
# 색상 정의
# ─────────────────────────────────────────────────────────

readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly CYAN='\033[0;36m'
readonly NC='\033[0m' # No Color

# ─────────────────────────────────────────────────────────
# 로깅 함수
# ─────────────────────────────────────────────────────────

log_info() {
  echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
  echo -e "${GREEN}✅ $1${NC}"
}

log_error() {
  echo -e "${RED}❌ $1${NC}" >&2
}

log_warning() {
  echo -e "${YELLOW}⚠️  $1${NC}"
}

log_header() {
  echo ""
  echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
  echo -e "${BLUE}$1${NC}"
  echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
  echo ""
}

log_section() {
  echo -e "\n${CYAN}▶ $1${NC}"
}

# ─────────────────────────────────────────────────────────
# 검증 함수
# ─────────────────────────────────────────────────────────

# 명령어 존재 확인
check_command() {
  if ! command -v "$1" &> /dev/null; then
    log_error "명령어 찾을 수 없음: $1"
    return 1
  fi
  log_success "명령어 확인: $1"
  return 0
}

# 파일 존재 확인
check_file() {
  if [[ ! -f "$1" ]]; then
    log_error "파일 찾을 수 없음: $1"
    return 1
  fi
  log_success "파일 확인: $1"
  return 0
}

# 디렉토리 존재 확인
check_directory() {
  if [[ ! -d "$1" ]]; then
    log_error "디렉토리 찾을 수 없음: $1"
    return 1
  fi
  log_success "디렉토리 확인: $1"
  return 0
}

# Git 저장소 확인
check_git_repo() {
  if ! git rev-parse --git-dir > /dev/null 2>&1; then
    log_error "Git 저장소 아님"
    return 1
  fi
  log_success "Git 저장소 확인"
  return 0
}

# ─────────────────────────────────────────────────────────
# Git 유틸리티 함수
# ─────────────────────────────────────────────────────────

# 현재 브랜치 이름
git_current_branch() {
  git rev-parse --abbrev-ref HEAD
}

# 브랜치 정리 (원격 삭제된 브랜치 제거)
git_cleanup_branches() {
  log_section "원격 브랜치 정리"
  git fetch --prune
  log_success "원격 브랜치 정리 완료"
}

# 미커밋 변경사항 확인
git_has_changes() {
  if git status --porcelain | grep -q '^'; then
    return 0 # 변경사항 있음
  fi
  return 1 # 변경사항 없음
}

# 최신 버전 태그 가져오기
git_latest_version() {
  git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"
}

# 태그 생성
git_create_tag() {
  local tag=$1
  local message=${2:-"Release $tag"}

  if git rev-parse "$tag" > /dev/null 2>&1; then
    log_warning "태그 이미 존재: $tag"
    return 1
  fi

  git tag -a "$tag" -m "$message"
  log_success "태그 생성: $tag"
}

# 태그 푸시
git_push_tags() {
  log_section "태그 푸시"
  git push --tags
  log_success "태그 푸시 완료"
}

# ─────────────────────────────────────────────────────────
# 파일 유틸리티
# ─────────────────────────────────────────────────────────

# 파일 백업
backup_file() {
  local file=$1
  local backup="${file}.backup.$(date +%Y%m%d_%H%M%S)"

  if [[ ! -f "$file" ]]; then
    log_error "파일 찾을 수 없음: $file"
    return 1
  fi

  cp "$file" "$backup"
  log_success "파일 백업: $backup"
  echo "$backup" # 백업 파일 경로 반환
}

# 정규식으로 파일 내용 수정
update_file() {
  local file=$1
  local pattern=$2
  local replacement=$3

  if [[ ! -f "$file" ]]; then
    log_error "파일 찾을 수 없음: $file"
    return 1
  fi

  # macOS와 Linux 호환성
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s/$pattern/$replacement/" "$file"
  else
    sed -i "s/$pattern/$replacement/" "$file"
  fi

  log_success "파일 수정: $file"
}

# ─────────────────────────────────────────────────────────
# 버전 유틸리티
# ─────────────────────────────────────────────────────────

# 현재 버전 읽기
get_current_version() {
  if [[ ! -f "pyproject.toml" ]]; then
    log_error "pyproject.toml 찾을 수 없음"
    return 1
  fi

  grep 'version =' pyproject.toml | head -1 | cut -d'"' -f2
}

# 버전 비교 (1: $1 > $2, 0: 같음, -1: $1 < $2)
compare_versions() {
  local v1=$1
  local v2=$2

  if [[ "$v1" > "$v2" ]]; then
    echo 1
  elif [[ "$v1" == "$v2" ]]; then
    echo 0
  else
    echo -1
  fi
}

# ─────────────────────────────────────────────────────────
# PyPI 유틸리티
# ─────────────────────────────────────────────────────────

# PyPI에서 현재 버전 조회
pypi_get_version() {
  local package=${1:-moai-adk}
  curl -s "https://pypi.org/pypi/$package/json" | \
    python3 -c "import sys, json; print(json.load(sys.stdin)['info']['version'])" 2>/dev/null || \
    echo "unknown"
}

# PyPI 토큰 검증
check_pypi_token() {
  if [[ -z "$UV_PUBLISH_TOKEN" ]]; then
    log_error "PyPI 토큰 설정 안됨 (UV_PUBLISH_TOKEN)"
    return 1
  fi

  log_success "PyPI 토큰 확인됨"
  return 0
}

# ─────────────────────────────────────────────────────────
# 빌드 유틸리티
# ─────────────────────────────────────────────────────────

# 빌드 산출물 정리
clean_build() {
  log_section "빌드 산출물 정리"

  rm -rf build/ dist/ *.egg-info .eggs/ 2>/dev/null || true

  log_success "빌드 산출물 정리 완료"
}

# 패키지 빌드
build_package() {
  log_section "패키지 빌드"

  clean_build

  if ! check_command "python"; then
    return 1
  fi

  python3 -m pip install -q build 2>/dev/null || \
    python3 -m pip install -q --user build

  python3 -m build

  if [[ -d "dist" ]]; then
    log_success "패키지 빌드 완료"
    ls -lh dist/
    return 0
  else
    log_error "패키지 빌드 실패"
    return 1
  fi
}

# ─────────────────────────────────────────────────────────
# 확인 함수
# ─────────────────────────────────────────────────────────

# 사용자 확인 (yes/no)
confirm() {
  local prompt=$1
  local response

  read -p "$(echo -e ${CYAN}${prompt} ${NC})" -n 1 -r response
  echo

  [[ "$response" =~ ^[Yy]$ ]]
}

# 예 무조건 답변
ask_yes_no() {
  local question=$1
  local default=${2:-y}

  if [[ "$default" == "y" ]]; then
    read -p "$(echo -e ${CYAN}${question}${NC}) [Y/n]: " -r response
    [[ -z "$response" ]] || [[ "$response" =~ ^[Yy]$ ]]
  else
    read -p "$(echo -e ${CYAN}${question}${NC}) [y/N]: " -r response
    [[ "$response" =~ ^[Yy]$ ]]
  fi
}

# ─────────────────────────────────────────────────────────
# 타이머 함수
# ─────────────────────────────────────────────────────────

# 시간 측정 시작
timer_start() {
  export _TIMER_START=$(date +%s)
}

# 시간 측정 종료 및 출력
timer_end() {
  if [[ -z "$_TIMER_START" ]]; then
    log_warning "타이머 미시작"
    return
  fi

  local end=$(date +%s)
  local elapsed=$((end - _TIMER_START))
  local mins=$((elapsed / 60))
  local secs=$((elapsed % 60))

  if [[ $mins -gt 0 ]]; then
    log_info "소요 시간: ${mins}분 ${secs}초"
  else
    log_info "소요 시간: ${secs}초"
  fi

  unset _TIMER_START
}

# ─────────────────────────────────────────────────────────
# 오류 처리
# ─────────────────────────────────────────────────────────

# 에러 발생 시 정리 함수
on_error() {
  local lineno=$1
  log_error "오류 발생 (줄: $lineno)"
  exit 1
}

# trap 설정
set_error_trap() {
  trap 'on_error ${LINENO}' ERR
}

# ─────────────────────────────────────────────────────────
# 디버그 함수
# ─────────────────────────────────────────────────────────

# 환경 정보 출력
print_env_info() {
  log_header "환경 정보"

  log_info "프로젝트 디렉토리: $(pwd)"
  log_info "Python: $(python3 --version 2>&1 || echo 'Not installed')"
  log_info "Git: $(git --version 2>&1 || echo 'Not installed')"
  log_info "OS: $OSTYPE"
  log_info "현재 버전: $(get_current_version || echo 'unknown')"
}

# 변수 디버그 출력
debug_vars() {
  local -n vars_ref=$1
  log_header "디버그 변수"

  for key in "${!vars_ref[@]}"; do
    log_info "$key: ${vars_ref[$key]}"
  done
}

# ─────────────────────────────────────────────────────────
# 외보내기 (이 파일이 sourced되었을 때)
# ─────────────────────────────────────────────────────────

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  # 직접 실행된 경우 (source가 아닌 경우)
  log_header "Release Helper Functions"
  echo "이 파일은 다른 스크립트에서 source로 사용하도록 설계됨"
  echo ""
  echo "사용법:"
  echo "  source .moai/release/release-helper.sh"
  echo "  log_info \"메시지\""
  echo "  check_command python3"
  echo ""
  echo "사용 가능한 함수:"
  echo "  로깅: log_info, log_success, log_error, log_warning"
  echo "  검증: check_command, check_file, check_git_repo"
  echo "  Git: git_current_branch, git_latest_version, git_cleanup_branches"
  echo "  파일: backup_file, update_file"
  echo "  빌드: clean_build, build_package"
  echo "  버전: get_current_version, compare_versions"
  echo "  확인: confirm, ask_yes_no"
  echo "  타이머: timer_start, timer_end"
  echo ""
fi
