#!/bin/bash
# Git 헬퍼 함수들
# MoAI-ADK 명령어에서 공통으로 사용하는 Git 작업

check_git_lock() {
    """
    Git index.lock 파일 검사 및 처리
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

safe_git_commit() {
    """
    안전한 Git 커밋 (lock 체크 포함)
    """
    local commit_message="$1"

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
    안전한 브랜치 생성/전환
    """
    local branch_name="$1"
    local create_if_missing="${2:-false}"

    check_git_lock

    if [ "$create_if_missing" = "true" ]; then
        git checkout -b "$branch_name" 2>/dev/null || git checkout "$branch_name"
    else
        git checkout "$branch_name"
    fi
}

# 함수 export (다른 스크립트에서 사용 가능)
export -f check_git_lock
export -f safe_git_commit
export -f safe_git_branch