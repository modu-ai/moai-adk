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

Constitution 5원칙을 준수하는 Claude Code 기반 지능형 커밋 시스템입니다.

@DESIGN:COMMIT-WORKFLOW-001 @TECH:CLAUDE-CODE-STD-001

## 현재 상태 확인

!`echo "=== Git 상태 ==="`
!`echo "📍 현재 브랜치: $(git branch --show-current)"`
!`echo "📋 Staged: $(git diff --cached --name-only | wc -l)개"`
!`echo "📝 Unstaged: $(git diff --name-only | wc -l)개"`
!`echo "🎯 모드: $(python3 -c "import json; config=json.load(open('.moai/config.json')); print(config['project']['mode'])" 2>/dev/null || echo "unknown")"`

## 변경사항 스테이징

!`git add -A`

## Claude Code 기반 커밋 메시지 생성

사용자가 메시지를 제공했다면 그대로 사용하고, 그렇지 않다면 Claude가 변경사항을 분석해서 의미있는 커밋 메시지를 생성합니다.

### 1. 변경사항 분석

먼저 스테이징된 변경사항을 확인합니다:

!`git diff --cached --stat`
!`git diff --cached --name-only`

### 2. 커밋 메시지 결정

**요청된 메시지**: "$ARGUMENTS"

만약 사용자가 메시지를 제공하지 않았다면, 위의 변경사항을 분석해서 다음 가이드라인에 따라 의미있는 커밋 메시지를 생성해주세요:

- 📝 문서/명세 변경: .md, README, 명세 파일 등
- 🧪 테스트 관련: test 디렉토리, *test.py, *spec.js 등
- ✨ 기능 구현: 새로운 함수, 클래스, API 추가
- 🐛 버그 수정: 기존 코드의 오류 수정
- ♻️ 리팩토링: 코드 구조 개선, 성능 향상
- ⚙️ 설정 변경: config 파일, .json, .yml 등
- 🔧 도구/스크립트: .claude, 빌드 스크립트 등

**실제 변경 내용을 보고 구체적이고 의미있는 메시지를 생성해주세요.**

### 3. 커밋 실행

위에서 결정된 커밋 메시지로 Constitution 5원칙을 준수하는 커밋을 실행합니다.

커밋 메시지는 다음 형식을 따릅니다:

```
[생성된 메시지]

[변경사항에 대한 간단한 설명]

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

실제 커밋을 실행해주세요.

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
