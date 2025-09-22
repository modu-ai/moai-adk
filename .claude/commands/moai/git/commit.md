---
name: moai:git:commit
description: Constitution 5원칙 기반 스마트 커밋
argument-hint: [메시지] - 커밋 메시지 또는 --auto 옵션
allowed-tools: Bash(git:*), Bash(python3:*), Read, Write, Glob, Grep
model: sonnet
---

# MoAI-ADK 단순화된 커밋 시스템

**커밋 메시지**: $ARGUMENTS

Constitution 5원칙을 준수하는 단순하고 안정적인 커밋 시스템입니다.

## 현재 상태 확인

Git 상태를 확인합니다:

!`git branch --show-current`
!`git diff --cached --name-only | wc -l`
!`git diff --name-only | wc -l`
!`python3 -c "import json; config=json.load(open('.moai/config.json')); print(config['project']['mode'])" 2>/dev/null || echo "unknown"`
!`git log --oneline -1 2>/dev/null || echo "초기 커밋 없음"`

## 커밋 실행

### 1단계: 변경사항 스테이징

모든 변경사항을 스테이징 영역에 추가합니다:

```bash
git add -A
echo "✅ 모든 변경사항 스테이징 완료"
```

### 2단계: 커밋 메시지 생성 및 커밋

인수에 따라 적절한 커밋 메시지를 생성하고 커밋을 실행합니다:

```bash
# 커밋 메시지 결정
if [[ "$ARGUMENTS" == "--auto" ]]; then
    # 자동 메시지 생성
    CHANGED_FILES=$(git diff --cached --name-only)
    FILE_COUNT=$(echo "$CHANGED_FILES" | wc -l)

    if echo "$CHANGED_FILES" | grep -q "\.md\|spec\|SPEC"; then
        COMMIT_MSG="📝 명세 및 문서 업데이트 ($FILE_COUNT개 파일)"
        CATEGORY="문서화"
    elif echo "$CHANGED_FILES" | grep -q "test\|Test"; then
        COMMIT_MSG="🧪 테스트 추가 및 개선 ($FILE_COUNT개 파일)"
        CATEGORY="테스트"
    elif echo "$CHANGED_FILES" | grep -q "\.py\|\.js\|\.ts\|\.java"; then
        COMMIT_MSG="✨ 기능 구현 및 개선 ($FILE_COUNT개 파일)"
        CATEGORY="구현"
    elif echo "$CHANGED_FILES" | grep -q "config\|\.json\|\.yml\|settings"; then
        COMMIT_MSG="⚙️ 설정 및 구성 업데이트 ($FILE_COUNT개 파일)"
        CATEGORY="설정"
    else
        COMMIT_MSG="🔧 프로젝트 업데이트 ($FILE_COUNT개 파일)"
        CATEGORY="일반"
    fi

    DETAIL="자동 생성된 커밋 - $CATEGORY 관련 변경사항"
else
    # 사용자 제공 메시지 사용
    COMMIT_MSG="$ARGUMENTS"
    DETAIL="사용자 지정 커밋 메시지"
fi

# Constitution 준수 footer 추가
FULL_MESSAGE=$(cat <<EOF
$COMMIT_MSG

$DETAIL

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)

# 커밋 실행
git commit -m "$FULL_MESSAGE"
echo "✅ 커밋 완료: $COMMIT_MSG"
```

### 3단계: 커밋 결과 확인

```bash
echo "=== 커밋 결과 ==="
echo "📝 최신 커밋: $(git log --oneline -1)"
echo "📊 총 커밋 수: $(git rev-list --count HEAD)"
echo "🎯 브랜치 상태: $(git status --porcelain | wc -l)개 미커밋 파일"
```

## 🎯 핵심 특징

- **단순성**: Constitution 5원칙 준수하는 간단한 구조
- **자동화**: 파일 변경사항 기반 지능적 메시지 생성
- **안정성**: 검증된 bash 명령어만 사용
- **표준 준수**: Claude Code와 MoAI-ADK 통합

## 사용법

### 자동 커밋 메시지 생성

```bash
/moai:git:commit --auto
```

### 사용자 지정 메시지

```bash
/moai:git:commit "JWT 인증 구현 완료"
```

## Constitution 5원칙 준수

1. **Simplicity**: 복잡한 bash 스크립트 제거, 단순한 구조
2. **Architecture**: 명확한 3단계 프로세스
3. **Testing**: 안정적인 Git 명령어 사용
4. **Observability**: 모든 커밋 과정 투명하게 출력
5. **Versioning**: 표준 Git 워크플로우 준수
