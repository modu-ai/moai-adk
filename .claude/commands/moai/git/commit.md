---
name: moai:git:commit
description: 📝 Constitution 기반 스마트 커밋
argument-hint: [메시지] - 커밋 메시지 또는 --auto 옵션
allowed-tools: Bash(git:*), Bash(python3:*), Read, Write, Glob, Grep
model: sonnet
---

# MoAI 스마트 커밋 시스템

@REQ:GIT-COMMIT-001 @FEATURE:SMART-COMMIT-001 @API:COMMIT-INTERFACE-001

**요청사항**: $ARGUMENTS

Constitution 5원칙을 준수하는 단순하고 안정적인 커밋 시스템입니다.

@DESIGN:COMMIT-WORKFLOW-001 @TECH:CLAUDE-CODE-STD-001

## 현재 상태 확인

!`echo "=== Git 상태 ==="`
!`echo "📍 현재 브랜치: $(git branch --show-current)"`
!`echo "📋 Staged: $(git diff --cached --name-only | wc -l)개"`
!`echo "📝 Unstaged: $(git diff --name-only | wc -l)개"`
!`echo "🎯 모드: $(python3 -c "import json; config=json.load(open('.moai/config.json')); print(config['project']['mode'])" 2>/dev/null || echo "unknown")"`

## 커밋 처리

### 변경사항 스테이징

!`git add -A`

### 커밋 메시지 생성 및 실행

!`python3 .moai/scripts/commit_helper.py "$ARGUMENTS"`

## 사용법

```bash
# 자동 메시지 생성 커밋
/moai:git:commit --auto

# 사용자 지정 메시지 커밋
/moai:git:commit "JWT 인증 구현 완료"

# 체크포인트 커밋
/moai:git:commit --checkpoint "실험 중간 백업"
```

## 특징

- **자동 메시지 생성**: 변경 파일 분석으로 적절한 메시지 자동 생성
- **Constitution 준수**: 모든 커밋에 표준 footer 자동 추가
- **파일 유형 감지**: .md, .py, test 등 파일 유형별 적절한 이모지 적용
- **간단한 인터페이스**: 복잡한 로직 없이 스크립트로 위임
