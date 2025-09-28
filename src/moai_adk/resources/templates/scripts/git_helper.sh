#!/bin/bash
# Git 헬퍼 함수들 (MoAI-ADK v0.1.9+)
# @FEATURE:GIT-001 Modularized Git operations support
#
# 새로운 모듈 구조:
# - GitInstallationManager: Git 설치 및 확인
# - GitStatusManager: 상태 확인 및 원격 정보
# - GitRepositoryManager: 저장소 초기화
# - GitLockManager: 잠금 시스템 관리

check_git_availability() {
    """
    Git 가용성 확인 (GitInstallationManager 호환)
    """
    if ! command -v git &> /dev/null; then
        echo "❌ Git이 설치되지 않았습니다."
        echo "💡 Git 설치 안내:"
        case "$(uname -s)" in
            Darwin*) echo "   brew install git" ;;
            Linux*)  echo "   sudo apt install git (Ubuntu/Debian)" ;;
            *)       echo "   https://git-scm.com/download" ;;
        esac
        return 1
    fi
    return 0
}

check_git_lock() {
    """
    Git index.lock 파일 검사 및 처리 (GitLockManager 호환)
    """
    if [ -f .git/index.lock ]; then
        echo "🔒 git index.lock 감지됨"

        # 활성 Git 프로세스 확인
        if pgrep -fl "git (commit|rebase|merge)" >/dev/null 2>&1; then
            echo "❌ 다른 git 작업이 진행 중입니다. 해당 작업을 종료한 후 다시 실행하세요."
            echo "💡 현재 실행 중인 git 프로세스:"
            pgrep -fl "git (commit|rebase|merge)"
            exit 1
        else
            echo "🔓 잔여 lock 파일 제거 중..."
            rm -f .git/index.lock
            echo "✅ lock 파일 제거 완료"
        fi
    fi
}

get_git_status() {
    """
    Git 상태 확인 (GitStatusManager 호환)
    """
    if ! check_git_availability; then
        return 1
    fi

    if [ ! -d .git ]; then
        echo "❌ Git 저장소가 아닙니다."
        return 1
    fi

    git status --porcelain
}

init_git_repository() {
    """
    Git 저장소 초기화 (GitRepositoryManager 호환)
    """
    if ! check_git_availability; then
        return 1
    fi

    if [ -d .git ]; then
        echo "ℹ️ 이미 Git 저장소입니다."
        return 0
    fi

    echo "🚀 Git 저장소 초기화 중..."
    git init
    local exit_code=$?

    if [ $exit_code -eq 0 ]; then
        echo "✅ Git 저장소 초기화 완료"
    else
        echo "❌ Git 저장소 초기화 실패"
        return $exit_code
    fi
}

safe_git_commit() {
    """
    안전한 Git 커밋 (GitLockManager 통합)
    """
    local commit_message="$1"

    check_git_availability || return 1
    check_git_lock

    if git diff --cached --quiet; then
        echo "ℹ️ 커밋할 변경사항이 없습니다."
        return 0
    fi

    git commit -m "$commit_message"
    local exit_code=$?

    if [ $exit_code -eq 0 ]; then
        echo "✅ 커밋 완료: $commit_message"
    else
        echo "❌ 커밋 실패 (exit code: $exit_code)"
        return $exit_code
    fi
}

safe_git_branch() {
    """
    안전한 브랜치 생성/전환 (GitLockManager 통합)
    """
    local branch_name="$1"
    local create_if_missing="${2:-false}"

    check_git_availability || return 1
    check_git_lock

    if [ "$create_if_missing" = "true" ]; then
        git checkout -b "$branch_name" 2>/dev/null || git checkout "$branch_name"
    else
        git checkout "$branch_name"
    fi
}

# 함수 export (모듈화된 Git 관리 지원)
export -f check_git_availability
export -f check_git_lock
export -f get_git_status
export -f init_git_repository
export -f safe_git_commit
export -f safe_git_branch