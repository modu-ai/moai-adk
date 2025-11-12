#!/bin/bash
# quality-check.sh - 릴리스 전 통합 품질 검증 스크립트
# 사용: ./quality-check.sh [--fix] [--verbose]
#
# 역할:
#   - pytest: 테스트 실행 및 커버리지 확인 (85% 이상 필수)
#   - mypy: 타입 검사 (strict 모드)
#   - ruff: 코드 스타일 및 lint 검사
#   - black: 코드 포맷팅 검증
#   - bandit: 보안 취약점 스캔

set -e

# ─────────────────────────────────────────────────────────
# 설정
# ─────────────────────────────────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
FIX_MODE=false
VERBOSE=false
EXIT_CODE=0

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ─────────────────────────────────────────────────────────
# 함수
# ─────────────────────────────────────────────────────────

print_header() {
  echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
  echo -e "${BLUE}$1${NC}"
  echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
}

print_success() {
  echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
  echo -e "${RED}❌ $1${NC}"
  EXIT_CODE=1
}

print_warning() {
  echo -e "${YELLOW}⚠️  $1${NC}"
}

print_section() {
  echo -e "\n${BLUE}▶ $1${NC}"
}

# ─────────────────────────────────────────────────────────
# 인자 파싱
# ─────────────────────────────────────────────────────────

while [[ $# -gt 0 ]]; do
  case $1 in
    --fix)
      FIX_MODE=true
      shift
      ;;
    --verbose)
      VERBOSE=true
      shift
      ;;
    --help)
      echo "사용: $0 [옵션]"
      echo ""
      echo "옵션:"
      echo "  --fix      자동 수정 가능한 문제 해결 (black, ruff)"
      echo "  --verbose  상세 출력"
      echo "  --help     도움말 표시"
      exit 0
      ;;
    *)
      echo "알 수 없는 옵션: $1"
      exit 1
      ;;
  esac
done

# ─────────────────────────────────────────────────────────
# 메인 검증
# ─────────────────────────────────────────────────────────

print_header "MoAI-ADK 릴리스 품질 검증"

cd "$PROJECT_ROOT"

# 1️⃣ Git 상태 확인
print_section "1️⃣  Git 상태 확인"

if ! git status > /dev/null 2>&1; then
  print_error "Git 저장소 아님"
  exit 1
fi

if git status --porcelain | grep -q '^'; then
  print_warning "커밋되지 않은 변경사항 존재:"
  git status --short
else
  print_success "Git 상태 정상 (변경사항 없음)"
fi

# 2️⃣ pytest 테스트 실행
print_section "2️⃣  테스트 실행 (pytest)"

if ! command -v pytest &> /dev/null; then
  print_error "pytest 미설치"
  exit 1
fi

if [[ $VERBOSE == true ]]; then
  pytest tests/ -v || EXIT_CODE=$?
else
  pytest tests/ -q || EXIT_CODE=$?
fi

# 테스트 커버리지 확인
print_section "테스트 커버리지 확인"

COVERAGE_OUTPUT=$(pytest --cov=src/moai_adk tests/ --cov-report=term-missing 2>&1 | tail -1)

if echo "$COVERAGE_OUTPUT" | grep -q "TOTAL"; then
  COVERAGE=$(echo "$COVERAGE_OUTPUT" | awk '{print $NF}' | sed 's/%//')

  if (( $(echo "$COVERAGE >= 85" | bc -l) )); then
    print_success "테스트 커버리지: $COVERAGE% (요구: 85% 이상) ✓"
  else
    print_error "테스트 커버리지: $COVERAGE% (요구: 85% 이상) ✗"
    EXIT_CODE=1
  fi
else
  print_warning "커버리지 측정 불가"
fi

# 3️⃣ mypy 타입 검사
print_section "3️⃣  타입 검사 (mypy)"

if ! command -v mypy &> /dev/null; then
  print_error "mypy 미설치"
  exit 1
fi

MYPY_OUTPUT=$(mypy src/moai_adk --strict 2>&1 || true)
MYPY_ERRORS=$(echo "$MYPY_OUTPUT" | grep -c "error:" || true)

if [[ $MYPY_ERRORS -eq 0 ]]; then
  print_success "타입 검사 통과 (오류 0개)"
else
  print_error "타입 검사 실패 (오류 $MYPY_ERRORS개)"
  if [[ $VERBOSE == true ]]; then
    echo "$MYPY_OUTPUT"
  fi
  EXIT_CODE=1
fi

# 4️⃣ ruff 코드 검사
print_section "4️⃣  코드 검사 (ruff)"

if ! command -v ruff &> /dev/null; then
  print_error "ruff 미설치"
  exit 1
fi

if [[ $FIX_MODE == true ]]; then
  print_warning "ruff 자동 수정 모드"
  ruff check src/ tests/ --fix --quiet
fi

RUFF_OUTPUT=$(ruff check src/ tests/ --select=E,W,F 2>&1 || true)
RUFF_ERRORS=$(echo "$RUFF_OUTPUT" | grep -c "error\|warning" || true)

if [[ -z "$RUFF_OUTPUT" ]]; then
  print_success "코드 검사 통과 (오류 0개)"
else
  print_error "코드 검사 실패"
  if [[ $VERBOSE == true ]]; then
    ruff check src/ tests/ --select=E,W,F
  fi
  EXIT_CODE=1
fi

# 5️⃣ black 포맷팅 검사
print_section "5️⃣  코드 포맷팅 (black)"

if ! command -v black &> /dev/null; then
  print_error "black 미설치"
  exit 1
fi

if [[ $FIX_MODE == true ]]; then
  print_warning "black 자동 포맷팅 모드"
  black src/ tests/ --quiet
fi

if black --check src/ tests/ > /dev/null 2>&1; then
  print_success "포맷팅 검사 통과"
else
  print_error "포맷팅 검사 실패"
  if [[ $VERBOSE == true ]]; then
    black --diff src/ tests/ | head -30
  fi
  EXIT_CODE=1
fi

# 6️⃣ bandit 보안 검사
print_section "6️⃣  보안 검사 (bandit)"

if ! command -v bandit &> /dev/null; then
  print_warning "bandit 미설치 (보안 검사 스킵)"
else
  BANDIT_OUTPUT=$(bandit -r src/moai_adk -f json 2>&1 || true)

  CRITICAL_COUNT=$(echo "$BANDIT_OUTPUT" | grep -o '"severity": "CRITICAL"' | wc -l || true)
  HIGH_COUNT=$(echo "$BANDIT_OUTPUT" | grep -o '"severity": "HIGH"' | wc -l || true)

  if [[ $CRITICAL_COUNT -eq 0 ]]; then
    if [[ $HIGH_COUNT -eq 0 ]]; then
      print_success "보안 검사 통과 (문제 0개)"
    else
      print_warning "보안 검사: HIGH 심각도 $HIGH_COUNT개 발견"
    fi
  else
    print_error "보안 검사 실패: CRITICAL 심각도 $CRITICAL_COUNT개 발견"
    if [[ $VERBOSE == true ]]; then
      bandit -r src/moai_adk -v
    fi
    EXIT_CODE=1
  fi
fi

# 7️⃣ pip-audit 의존성 보안 검사
print_section "7️⃣  의존성 보안 검사 (pip-audit)"

if ! command -v pip-audit &> /dev/null; then
  print_warning "pip-audit 미설치 (의존성 검사 스킵)"
else
  if pip-audit -q 2>&1 | grep -q "found 0 vulnerabilities"; then
    print_success "의존성 보안 검사 통과"
  else
    print_warning "의존성 취약점 발견 (자세한 내용: pip-audit)"
  fi
fi

# ─────────────────────────────────────────────────────────
# 결과 요약
# ─────────────────────────────────────────────────────────

print_header "검증 결과 요약"

if [[ $EXIT_CODE -eq 0 ]]; then
  print_success "모든 검증 통과! 배포 준비 완료"
  echo ""
  echo "다음 단계:"
  echo "  1. /moai:release patch --dry-run  (건조 실행으로 확인)"
  echo "  2. /moai:release patch             (실제 배포)"
else
  print_error "검증 실패. 위의 오류를 수정하고 다시 실행하세요"
  echo ""
  echo "옵션:"
  echo "  ./quality-check.sh --fix       (자동 수정)"
  echo "  ./quality-check.sh --verbose   (상세 출력)"
fi

exit $EXIT_CODE
