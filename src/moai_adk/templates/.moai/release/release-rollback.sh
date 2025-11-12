#!/bin/bash
# release-rollback.sh - 릴리스 자동 롤백 스크립트
#
# 역할:
#   - PyPI에서 문제 버전 제거
#   - git 히스토리 롤백
#   - GitHub 이슈 자동 생성
#   - 사용자 공지 생성
#
# 사용:
#   ./release-rollback.sh v0.22.5              # 최신 안정 버전으로 롤백
#   ./release-rollback.sh v0.22.5 v0.22.4     # 특정 버전으로 롤백
#   ./release-rollback.sh --list                # 롤백 가능 버전 확인

set -e

# ─────────────────────────────────────────────────────────
# 설정
# ─────────────────────────────────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Helper 함수 로드
source "$SCRIPT_DIR/release-helper.sh"

# ─────────────────────────────────────────────────────────
# 함수
# ─────────────────────────────────────────────────────────

# 배포된 버전 목록 가져오기
get_deployed_versions() {
  curl -s "https://pypi.org/pypi/moai-adk/json" | \
    python3 -c "import sys, json; versions = list(json.load(sys.stdin)['releases'].keys()); print('\n'.join(sorted(versions, reverse=True)[:20]))" 2>/dev/null || \
    echo "unknown"
}

# 특정 버전이 배포되었는지 확인
is_version_deployed() {
  local version=$1
  get_deployed_versions | grep -q "^$version\$" && return 0 || return 1
}

# PyPI에서 버전 제거
remove_from_pypi() {
  local version=$1

  if [[ -z "$UV_PUBLISH_TOKEN" ]]; then
    log_error "PyPI 토큰 미설정 (UV_PUBLISH_TOKEN)"
    return 1
  fi

  log_info "PyPI에서 $version 제거 중..."

  if command -v twine &> /dev/null; then
    twine delete "moai-adk==$version" --repository pypi || {
      log_error "twine 삭제 실패"
      return 1
    }
  else
    log_warning "twine 미설치 - PyPI 웹사이트에서 수동 삭제 필요"
    echo "수동 단계:"
    echo "1. https://pypi.org/project/moai-adk/ 접속"
    echo "2. Release History → v$version → Delete release"
    return 1
  fi

  log_success "PyPI에서 $version 제거 완료"
  return 0
}

# 이전 안정 버전 찾기
get_previous_stable() {
  # 로컬 git 태그에서 찾기
  git tag -l "v*" | \
    sort -V | \
    tail -2 | \
    head -1 | \
    sed 's/^v//'
}

# Git에 롤백 태그 생성
create_rollback_tag() {
  local from_version=$1
  local to_version=$2
  local tag_name="rollback-${from_version}-to-${to_version}-$(date +%Y%m%d_%H%M%S)"

  log_info "롤백 태그 생성: $tag_name"

  git tag -a "$tag_name" -m "Rollback from v$from_version to v$to_version"
  git push origin "$tag_name"

  log_success "롤백 태그 생성 완료"
}

# GitHub 이슈 생성
create_github_issue() {
  local from_version=$1
  local to_version=$2
  local reason=${3:-"Critical bug found"}

  if ! check_command "gh"; then
    log_warning "GitHub CLI 미설치 - 이슈 자동 생성 스킵"
    return 0
  fi

  log_info "GitHub 이슈 생성: 롤백 v$from_version → v$to_version"

  gh issue create \
    --title "ROLLBACK: v$from_version" \
    --body "$(cat <<EOF
## 롤백 정보

- **대상 버전**: v$from_version
- **롤백 버전**: v$to_version
- **사유**: $reason
- **시간**: $(date)

## 조치 내용

- PyPI에서 v$from_version 제거 완료
- git 롤백 태그 생성 완료

## 다음 단계

1. 문제 원인 파악
2. 버그 수정
3. v${from_version}-fixed 버전으로 재배포

## 영향 범위

사용자는 다음 명령으로 롤백:
\`\`\`bash
pip install moai-adk==$to_version
\`\`\`

---

자동 생성됨: Release Rollback Script
EOF
)" \
    --label "release" \
    --label "critical" || \
    log_warning "GitHub 이슈 생성 실패"
}

# 사용자 공지 생성
create_user_notice() {
  local from_version=$1
  local to_version=$2
  local notice_file="$PROJECT_ROOT/.moai/release/ROLLBACK_NOTICE_${from_version}.md"

  cat > "$notice_file" << EOF
# 🚨 긴급 공지: v$from_version 회수됨

**발행일**: $(date)

## 개요

v$from_version이 치명적인 버그로 인해 PyPI에서 제거되었습니다.

## 영향 받은 사용자

- v$from_version을 설치한 모든 사용자

## 권장 조치

즉시 다음 버전으로 다운그레이드하세요:

\`\`\`bash
pip install moai-adk==$to_version
\`\`\`

## 재배포 일정

- 현재: 버그 수정 중
- 예정: v${from_version}-fixed 재배포 (ETA: 24시간 이내)

## 문제 보고

발생한 문제는 [GitHub Issues](https://github.com/modu-ai/moai-adk/issues)에 보고해주세요.

---

🤖 자동 생성됨: Release Rollback System
EOF

  log_success "사용자 공지 생성: $notice_file"
}

# 롤백 요약 보고서 생성
create_rollback_report() {
  local from_version=$1
  local to_version=$2
  local report_file="$PROJECT_ROOT/.moai/reports/rollback_${from_version}_$(date +%Y%m%d_%H%M%S).md"

  # 디렉토리 생성
  mkdir -p "$(dirname "$report_file")"

  cat > "$report_file" << EOF
# 릴리스 롤백 보고서

## 기본 정보

- **롤백 시간**: $(date)
- **롤백 버전**: v$from_version → v$to_version
- **실행자**: $(whoami)
- **호스트**: $(hostname)

## 변경사항 요약

### PyPI
- v$from_version 제거 완료

### Git
- 롤백 태그 생성
- 원격 저장소에 푸시

### GitHub
- 이슈 자동 생성

## 실행 로그

\`\`\`
롤백 스크립트 실행 로그
\`\`\`

## 다음 단계 체크리스트

- [ ] 문제 원인 파악
- [ ] 버그 수정 완료
- [ ] 테스트 실행 (커버리지 85% 이상)
- [ ] v${from_version}-fixed 재배포
- [ ] 사용자 공지 발행

## 모니터링

- PyPI 다운로드 통계: [Link]
- GitHub Issues: [Link]
- 커뮤니티 반응: [Link]

---

문제가 재발생하면 이 보고서를 참고하세요.
EOF

  log_success "롤백 보고서 생성: $report_file"
}

# ─────────────────────────────────────────────────────────
# 메인 롤백 프로세스
# ─────────────────────────────────────────────────────────

main() {
  local from_version=${1:-}
  local to_version=${2:-}

  # 인자 확인
  if [[ -z "$from_version" ]]; then
    case "$from_version" in
      --list)
        log_header "배포된 버전 목록"
        get_deployed_versions
        exit 0
        ;;
      --help)
        print_help
        exit 0
        ;;
      *)
        print_help
        exit 1
        ;;
    esac
  fi

  # 버전 형식 확인
  if ! [[ "$from_version" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    log_error "잘못된 버전 형식: $from_version (기대: vX.Y.Z)"
    exit 1
  fi

  # v 접두어 제거
  from_version="${from_version#v}"

  # 이전 안정 버전 자동 검색
  if [[ -z "$to_version" ]]; then
    to_version=$(get_previous_stable)
    log_info "이전 안정 버전 자동 선택: $to_version"
  else
    to_version="${to_version#v}"
  fi

  # 최종 확인
  log_header "롤백 확인"
  log_warning "다음 조치를 수행할 예정입니다:"
  echo "  1. PyPI에서 v$from_version 제거"
  echo "  2. Git에 롤백 태그 생성"
  echo "  3. GitHub 이슈 자동 생성"
  echo "  4. 사용자 공지 생성"
  echo ""

  if ! ask_yes_no "계속하시겠습니까?"; then
    log_info "롤백 취소됨"
    exit 0
  fi

  # 실제 롤백 실행
  timer_start

  log_header "릴리스 롤백 실행 중"

  # 1. PyPI에서 제거
  log_section "1️⃣  PyPI에서 v$from_version 제거"
  if remove_from_pypi "$from_version"; then
    log_success "PyPI 제거 완료"
  else
    log_warning "PyPI 제거 실패 (수동 처리 필요)"
  fi

  # 2. Git 롤백 태그 생성
  log_section "2️⃣  Git 롤백 태그 생성"
  if check_git_repo; then
    create_rollback_tag "$from_version" "$to_version"
  fi

  # 3. GitHub 이슈 생성
  log_section "3️⃣  GitHub 이슈 생성"
  create_github_issue "$from_version" "$to_version" "Automatic rollback"

  # 4. 사용자 공지 생성
  log_section "4️⃣  사용자 공지 생성"
  create_user_notice "$from_version" "$to_version"

  # 5. 롤백 보고서 생성
  log_section "5️⃣  롤백 보고서 생성"
  create_rollback_report "$from_version" "$to_version"

  # 완료
  timer_end

  log_header "롤백 완료"
  echo "✅ v$from_version → v$to_version 롤백 완료"
  echo ""
  echo "다음 단계:"
  echo "1. 문제 원인 파악 및 수정"
  echo "2. 테스트 실행"
  echo "3. v${from_version}-fixed 재배포"
  echo ""
  echo "생성된 파일:"
  echo "- 사용자 공지: .moai/release/ROLLBACK_NOTICE_${from_version}.md"
  echo "- 롤백 보고서: .moai/reports/"
}

print_help() {
  cat << EOF
사용: release-rollback.sh <VERSION> [ROLLBACK_TO]

옵션:
  <VERSION>       제거할 버전 (예: v0.22.5 또는 0.22.5)
  [ROLLBACK_TO]   롤백 대상 버전 (기본: 이전 안정 버전)
  --list          배포된 버전 목록 표시
  --help          도움말 표시

예제:
  ./release-rollback.sh v0.22.5                # 최신 안정 버전으로 롤백
  ./release-rollback.sh v0.22.5 v0.22.4       # 특정 버전으로 롤백
  ./release-rollback.sh --list                  # 배포된 버전 목록

역할:
  - PyPI에서 문제 버전 제거
  - Git 히스토리에 롤백 태그 기록
  - GitHub 이슈 자동 생성
  - 사용자 공지 및 롤백 보고서 생성

주의:
  - PyPI 토큰 필수 (UV_PUBLISH_TOKEN)
  - GitHub CLI 설치 권장 (자동 이슈 생성용)
EOF
}

# ─────────────────────────────────────────────────────────
# 진입점
# ─────────────────────────────────────────────────────────

# 에러 트랩 설정
set_error_trap

# 도움말 확인
if [[ "$1" == "--help" ]]; then
  print_help
  exit 0
fi

# 메인 실행
main "$@"
