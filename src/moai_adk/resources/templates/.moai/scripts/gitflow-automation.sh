#!/bin/bash

# =============================================================================
# MoAI-ADK GitFlow 자동화 스크립트 v0.2.1
# =============================================================================
# 사용법:
#   ./gitflow-automation.sh spec SPEC-001 "user authentication"
#   ./gitflow-automation.sh build SPEC-001 "implement auth API"
#   ./gitflow-automation.sh sync SPEC-001 "update documentation"
# =============================================================================

set -e  # 에러 발생 시 스크립트 종료

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 로깅 함수
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 파라미터 검증
if [ $# -lt 3 ]; then
    log_error "Usage: $0 <command> <spec-id> <description> [feature-name]"
    log_error "Commands: spec, build, sync"
    log_error "Example: $0 spec SPEC-001 'JWT user authentication' user-auth"
    exit 1
fi

COMMAND=$1
SPEC_ID=$2
DESCRIPTION=$3
FEATURE_NAME=${4:-$(echo "$DESCRIPTION" | tr ' ' '-' | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9-]//g')}

# Git 상태 확인
check_git_status() {
    log_info "Git 상태 확인 중..."

    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log_error "Git 저장소가 아닙니다. 먼저 git init을 실행하세요."
        exit 1
    fi

    if [[ -n $(git status --porcelain) ]]; then
        log_warning "작업 중인 변경사항이 있습니다. 자동으로 스태시합니다."
        git stash push -m "MoAI GitFlow: Auto-stash before $COMMAND"
    fi
}

# 브랜치 전략 함수
create_feature_branch() {
    local spec_id=$1
    local feature_name=$2
    local branch_name="feature/${spec_id}-${feature_name}"

    log_info "Feature 브랜치 생성: $branch_name"

    # develop 브랜치로 전환
    if git show-ref --verify --quiet refs/heads/develop; then
        git checkout develop
        git pull origin develop 2>/dev/null || log_warning "리모트에서 develop 브랜치를 가져올 수 없습니다."
    else
        log_warning "develop 브랜치가 없습니다. main에서 생성합니다."
        git checkout -b develop
    fi

    # Feature 브랜치 생성 및 전환
    if git show-ref --verify --quiet refs/heads/$branch_name; then
        log_info "기존 브랜치로 전환: $branch_name"
        git checkout $branch_name
    else
        log_info "새 브랜치 생성: $branch_name"
        git checkout -b $branch_name
        git push -u origin $branch_name 2>/dev/null || log_warning "리모트 브랜치 생성 실패 (로컬에서 계속)"
    fi

    echo $branch_name
}

# SPEC 단계별 커밋
commit_spec_stage() {
    local spec_id=$1
    local stage=$2
    local description=$3

    case $stage in
        "init")
            git add .moai/specs/$spec_id/spec.md
            git commit -m "feat($spec_id): Add initial EARS requirements draft

$description

- EARS 키워드 구조화 완료
- 초기 요구사항 정의
- [NEEDS CLARIFICATION] 마커 추가"
            ;;
        "stories")
            git add .moai/specs/$spec_id/user-stories.md
            git commit -m "feat($spec_id): Add user stories US-001~005

$description

- User Stories 생성 완료
- 수락 기준 초안 작성
- 우선순위 및 복잡도 평가"
            ;;
        "acceptance")
            git add .moai/specs/$spec_id/acceptance.md
            git commit -m "feat($spec_id): Add acceptance criteria with GWT scenarios

$description

- Given-When-Then 시나리오 완료
- 테스트 가능한 수락 기준 정의
- 품질 검증 체크리스트 완료"
            ;;
        "complete")
            git add .moai/specs/$spec_id/
            git commit -m "feat($spec_id): Complete $spec_id specification

$description

- SPEC 문서 최종 검토 완료
- TAG 추적성 매핑 완료
- 품질 지표 충족 확인"
            ;;
    esac

    log_success "커밋 완료: $stage stage"
}

# BUILD 단계별 커밋 (TDD)
commit_build_stage() {
    local spec_id=$1
    local stage=$2
    local description=$3

    case $stage in
        "constitution")
            git add .moai/plans/
            git commit -m "feat($spec_id): Constitution 5원칙 검증 완료

$description

- Simplicity: 복잡도 제한 확인
- Architecture: 모듈형 구조 설계
- Testing: TDD 계획 수립
- Observability: 로깅 전략 정의
- Versioning: 버전 관리 체계 확립"
            ;;
        "red")
            git add tests/
            git commit -m "test($spec_id): Add failing tests (RED phase)

$description

- 실패하는 테스트 케이스 작성
- TDD Red 단계 완료
- 테스트 커버리지 목표 설정"
            ;;
        "green")
            git add src/ tests/
            git commit -m "feat($spec_id): Implement core functionality (GREEN phase)

$description

- 테스트를 통과하는 최소 구현
- TDD Green 단계 완료
- 기능 동작 검증 완료"
            ;;
        "refactor")
            git add src/ tests/
            git commit -m "refactor($spec_id): Code optimization and cleanup (REFACTOR phase)

$description

- 코드 품질 개선
- 중복 코드 제거
- 성능 최적화
- TDD Refactor 단계 완료"
            ;;
    esac

    log_success "커밋 완료: $stage stage"
}

# SYNC 단계별 커밋
commit_sync_stage() {
    local spec_id=$1
    local stage=$2
    local description=$3

    case $stage in
        "docs")
            git add docs/ README.md
            git commit -m "docs($spec_id): Update documentation and README

$description

- API 문서 자동 생성
- README 업데이트
- 사용 가이드 동기화"
            ;;
        "tags")
            git add .moai/indexes/tags.json
            git commit -m "chore($spec_id): Update TAG system and traceability

$description

- 16-Core TAG 인덱스 업데이트
- 추적성 매트릭스 갱신
- 의존성 관계 검증"
            ;;
        "final")
            git add .
            git commit -m "chore($spec_id): Final synchronization and cleanup

$description

- 모든 문서 동기화 완료
- 품질 게이트 통과 확인
- 배포 준비 완료"
            ;;
    esac

    log_success "커밋 완료: $stage stage"
}

# Draft PR 생성
create_draft_pr() {
    local spec_id=$1
    local title=$2
    local description=$3

    log_info "Draft PR 생성 중..."

    # gh CLI 설치 확인
    if ! command -v gh &> /dev/null; then
        log_warning "GitHub CLI (gh) 가 설치되어 있지 않습니다."
        log_info "수동으로 PR을 생성해주세요: https://github.com/your-repo/compare"
        return 1
    fi

    # PR 템플릿 생성
    local pr_body=".moai/tmp/pr-body-${spec_id}.md"
    mkdir -p .moai/tmp

    cat > $pr_body << EOF
# $spec_id: $title 🚀

## 📋 변경사항 요약
$description

## 📊 생성된 파일
- [x] .moai/specs/$spec_id/spec.md - EARS 형식 요구사항
- [x] .moai/specs/$spec_id/user-stories.md - User Stories
- [x] .moai/specs/$spec_id/acceptance.md - 수락 기준

## 🏷️ TAG 매핑
- REQ:${spec_id/SPEC-/} → DESIGN:${spec_id/SPEC-/} → TASK:${spec_id/SPEC-/}

## 🔄 다음 단계
- [ ] Constitution 5원칙 검증
- [ ] TDD 구현 진행
- [ ] 문서 동기화

## 📝 체크리스트
- [x] SPEC 문서 작성 완료
- [x] User Stories 정의
- [x] 수락 기준 작성
- [x] 품질 검증 통과
- [ ] Constitution 검증 대기
- [ ] TDD 구현 대기

---
🤖 MoAI-ADK v0.2.1에서 자동 생성됨
EOF

    # Draft PR 생성
    local pr_url
    if pr_url=$(gh pr create --draft \
        --title "$spec_id: $title" \
        --body-file "$pr_body" 2>/dev/null); then
        log_success "Draft PR 생성 완료: $pr_url"
    else
        log_warning "PR 생성 실패. 수동으로 생성해주세요."
    fi

    # 임시 파일 정리
    rm -f $pr_body
}

# PR 업데이트
update_pr() {
    local spec_id=$1
    local stage=$2
    local description=$3

    log_info "PR 업데이트 중..."

    if ! command -v gh &> /dev/null; then
        log_warning "GitHub CLI (gh)가 없어 PR 업데이트를 건너뜁니다."
        return 1
    fi

    local comment="## 🔄 $stage 단계 완료

$description

진행률: $(get_progress_percentage $spec_id)% 완료

---
🤖 MoAI-ADK v0.2.1 자동 업데이트"

    gh pr comment --body "$comment" 2>/dev/null || log_warning "PR 댓글 추가 실패"
}

# 진행률 계산
get_progress_percentage() {
    local spec_id=$1
    local current_branch=$(git branch --show-current)

    # 커밋 개수로 진행률 추정
    local total_commits=$(git rev-list --count $current_branch 2>/dev/null || echo "0")

    case $COMMAND in
        "spec") echo $((total_commits * 25)) ;;  # 4단계 * 25%
        "build") echo $((25 + total_commits * 15)) ;; # SPEC 25% + 5단계 * 15%
        "sync") echo $((85 + total_commits * 5)) ;;   # 이전 85% + 3단계 * 5%
        *) echo "0" ;;
    esac
}

# PR을 Ready로 변경
make_pr_ready() {
    local spec_id=$1

    log_info "PR을 Ready 상태로 변경 중..."

    if ! command -v gh &> /dev/null; then
        log_warning "GitHub CLI (gh)가 없어 PR 상태 변경을 건너뜁니다."
        return 1
    fi

    if gh pr ready 2>/dev/null; then
        log_success "PR이 Ready 상태로 변경되었습니다."

        # 리뷰 요청
        local reviewers=$(git config moai.default-reviewers 2>/dev/null || echo "")
        if [[ -n "$reviewers" ]]; then
            gh pr edit --add-reviewer "$reviewers" 2>/dev/null || log_warning "리뷰어 추가 실패"
        fi
    else
        log_warning "PR 상태 변경 실패"
    fi
}

# 메인 실행 로직
main() {
    log_info "MoAI-ADK GitFlow 자동화 시작: $COMMAND for $SPEC_ID"

    check_git_status

    case $COMMAND in
        "spec")
            branch_name=$(create_feature_branch $SPEC_ID $FEATURE_NAME)

            log_info "SPEC 단계별 커밋 시작..."
            commit_spec_stage $SPEC_ID "init" "$DESCRIPTION"
            commit_spec_stage $SPEC_ID "stories" "$DESCRIPTION"
            commit_spec_stage $SPEC_ID "acceptance" "$DESCRIPTION"
            commit_spec_stage $SPEC_ID "complete" "$DESCRIPTION"

            create_draft_pr $SPEC_ID "$DESCRIPTION" "SPEC 문서 작성 완료"
            ;;

        "build")
            log_info "BUILD 단계별 커밋 시작..."
            commit_build_stage $SPEC_ID "constitution" "$DESCRIPTION"
            update_pr $SPEC_ID "Constitution 검증" "$DESCRIPTION"

            commit_build_stage $SPEC_ID "red" "$DESCRIPTION"
            update_pr $SPEC_ID "TDD RED" "$DESCRIPTION"

            commit_build_stage $SPEC_ID "green" "$DESCRIPTION"
            update_pr $SPEC_ID "TDD GREEN" "$DESCRIPTION"

            commit_build_stage $SPEC_ID "refactor" "$DESCRIPTION"
            update_pr $SPEC_ID "TDD REFACTOR" "$DESCRIPTION"
            ;;

        "sync")
            log_info "SYNC 단계별 커밋 시작..."
            commit_sync_stage $SPEC_ID "docs" "$DESCRIPTION"
            update_pr $SPEC_ID "문서 동기화" "$DESCRIPTION"

            commit_sync_stage $SPEC_ID "tags" "$DESCRIPTION"
            update_pr $SPEC_ID "TAG 시스템 업데이트" "$DESCRIPTION"

            commit_sync_stage $SPEC_ID "final" "$DESCRIPTION"
            make_pr_ready $SPEC_ID
            ;;

        *)
            log_error "알 수 없는 명령어: $COMMAND"
            log_error "사용 가능한 명령어: spec, build, sync"
            exit 1
            ;;
    esac

    log_success "✅ MoAI-ADK GitFlow 자동화 완료!"
    log_info "현재 브랜치: $(git branch --show-current)"
    log_info "GitHub에서 진행 상황을 확인하세요: https://github.com/$(git config remote.origin.url | sed 's/.*github.com[:/]\([^/]*\/[^/]*\).git/\1/')/pulls"
}

# 스크립트 실행
main