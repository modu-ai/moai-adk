---
name: moai:git:checkpoint
description: 개인 모드 체크포인트 시스템 (안전한 실험 지원)
argument-hint: [메시지] - 체크포인트 메시지 또는 --list, --status, --cleanup 옵션
allowed-tools: Bash(git:*), Bash(python3:*), Read, Write, Glob, Grep
model: sonnet
---

# MoAI-ADK 단순화된 체크포인트 시스템

**체크포인트**: $ARGUMENTS

개인 모드에서 안전한 실험을 위한 Constitution 5원칙 준수 체크포인트 시스템입니다.

## 현재 상태 확인

체크포인트 시스템 상태를 확인합니다:

!`git branch --show-current`
!`git status --porcelain | wc -l`
!`python3 -c "import json; config=json.load(open('.moai/config.json')); print(config['project']['mode'])" 2>/dev/null || echo "unknown"`
!`git branch | grep -c checkpoint_ || echo "0"`

## 체크포인트 실행

### 1단계: 모드 확인

개인 모드에서만 체크포인트를 생성합니다:

```bash
# 프로젝트 모드 확인
PROJECT_MODE=$(python3 -c "import json; config=json.load(open('.moai/config.json')); print(config['project']['mode'])" 2>/dev/null || echo "unknown")

if [[ "$PROJECT_MODE" != "personal" ]]; then
    echo "⚠️ 체크포인트는 개인 모드에서만 지원됩니다."
    echo "현재 모드: $PROJECT_MODE"
    echo "개인 모드로 전환하려면: sed -i 's/\"mode\": \"team\"/\"mode\": \"personal\"/' .moai/config.json"
    exit 1
fi

echo "✅ 개인 모드 확인 완료"
```

### 2단계: 체크포인트 생성

인수에 따라 적절한 동작을 수행합니다:

```bash
# 타임스탬프 생성
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
CHECKPOINT_ID="checkpoint_$TIMESTAMP"

if [[ "$ARGUMENTS" == "--list" ]]; then
    echo "=== 체크포인트 목록 ==="
    git branch | grep "checkpoint_" | sort -r | head -10
    exit 0
elif [[ "$ARGUMENTS" == "--status" ]]; then
    echo "=== 체크포인트 시스템 상태 ==="
    echo "모드: $PROJECT_MODE"
    echo "자동 체크포인트: $(python3 -c "import json; config=json.load(open('.moai/config.json')); print(config['git_strategy']['personal']['auto_checkpoint'])" 2>/dev/null || echo "unknown")"
    echo "총 체크포인트: $(git branch | grep -c checkpoint_ || echo "0")개"
    echo "최근 체크포인트: $(git branch | grep checkpoint_ | tail -1 | xargs)"
    exit 0
elif [[ "$ARGUMENTS" == "--cleanup" ]]; then
    echo "=== 오래된 체크포인트 정리 ==="
    # 7일 이상 된 체크포인트 브랜치 삭제
    git for-each-ref --format='%(refname:short)' refs/heads/checkpoint_* | while read branch; do
        # 브랜치 생성 시간 확인 (간소화된 로직)
        BRANCH_DATE=$(echo "$branch" | grep -o '[0-9]\{8\}' | head -1)
        if [[ -n "$BRANCH_DATE" ]]; then
            DAYS_OLD=$(( ( $(date +%s) - $(date -d "${BRANCH_DATE:0:4}-${BRANCH_DATE:4:2}-${BRANCH_DATE:6:2}" +%s) ) / 86400 ))
            if [[ $DAYS_OLD -gt 7 ]]; then
                echo "🗑️ 삭제: $branch (${DAYS_OLD}일 경과)"
                git branch -D "$branch" 2>/dev/null || true
            fi
        fi
    done
    exit 0
fi

# 체크포인트 메시지 설정
if [[ -n "$ARGUMENTS" ]]; then
    CHECKPOINT_MSG="$ARGUMENTS"
else
    CHECKPOINT_MSG="Auto checkpoint $(date '+%Y-%m-%d %H:%M:%S')"
fi

echo "💾 체크포인트 생성: $CHECKPOINT_MSG"
```

### 3단계: 체크포인트 생성 실행

```bash
# 변경사항 스테이징
git add -A
echo "✅ 변경사항 스테이징 완료"

# 체크포인트 커밋 생성
CHECKPOINT_COMMIT_MSG=$(cat <<EOF
🔄 Checkpoint: $CHECKPOINT_MSG

타임스탬프: $(date '+%Y-%m-%d %H:%M:%S')
체크포인트 ID: $CHECKPOINT_ID

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)

git commit -m "$CHECKPOINT_COMMIT_MSG"
echo "✅ 체크포인트 커밋 생성 완료"

# 체크포인트 브랜치 생성
git branch "$CHECKPOINT_ID" HEAD
echo "✅ 체크포인트 브랜치 생성: $CHECKPOINT_ID"
```

### 4단계: 체크포인트 확인

```bash
echo "=== 체크포인트 생성 결과 ==="
echo "🆔 체크포인트 ID: $CHECKPOINT_ID"
echo "📝 커밋 해시: $(git rev-parse HEAD)"
echo "📅 생성 시간: $(date '+%Y-%m-%d %H:%M:%S')"
echo "📊 총 체크포인트: $(git branch | grep -c checkpoint_)개"
echo "📋 메시지: $CHECKPOINT_MSG"
```

## 🎯 핵심 특징

- **단순성**: Constitution 5원칙 준수하는 간단한 구조
- **개인 모드 전용**: 팀 모드에서는 사용 제한
- **안전성**: 실제 timestamp 값 사용, 패턴 매칭 오류 해결
- **자동 정리**: 오래된 체크포인트 자동 관리

## 사용법

### 기본 체크포인트 생성

```bash
/moai:git:checkpoint "실험 시작"
```

### 자동 메시지 체크포인트

```bash
/moai:git:checkpoint
```

### 체크포인트 목록 확인

```bash
/moai:git:checkpoint --list
```

### 시스템 상태 확인

```bash
/moai:git:checkpoint --status
```

### 오래된 체크포인트 정리

```bash
/moai:git:checkpoint --cleanup
```

## Constitution 5원칙 준수

1. **Simplicity**: 복잡한 패턴 매칭 제거, 단순한 timestamp 처리
2. **Architecture**: 명확한 4단계 프로세스
3. **Testing**: 안전한 실험 환경 제공
4. **Observability**: 모든 체크포인트 과정 투명하게 출력
5. **Versioning**: 체계적인 브랜치 기반 백업
