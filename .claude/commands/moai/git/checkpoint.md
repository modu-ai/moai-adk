---
name: moai:git:checkpoint
description: 💾 개인 모드 체크포인트 시스템
argument-hint: [메시지] - 체크포인트 메시지 또는 --list, --status, --cleanup 옵션
allowed-tools: Bash(git:*), Bash(python3:*), Read, Write, Glob, Grep
model: sonnet
---

# MoAI 체크포인트 시스템

**요청사항**: $ARGUMENTS

개인 모드에서 안전한 실험을 위한 Constitution 5원칙 준수 체크포인트 시스템입니다.

## 현재 상태 확인

!`echo "=== 체크포인트 상태 ==="`
!`echo "📍 현재 브랜치: $(git branch --show-current)"`
!`echo "📝 변경사항: $(git status --porcelain | wc -l)개"`
!`echo "🎯 모드: $(python3 -c "import json; config=json.load(open('.moai/config.json')); print(config['project']['mode'])" 2>/dev/null || echo "unknown")"`
!`echo "💾 체크포인트: $(git branch | grep -c checkpoint_)개"`

## 체크포인트 처리

### 개인 모드 확인 및 실행

!`python3 .moai/scripts/checkpoint_manager.py "$ARGUMENTS"`

## 사용법

```bash
# 기본 체크포인트 생성
/moai:git:checkpoint "실험 시작"

# 자동 메시지 체크포인트
/moai:git:checkpoint

# 체크포인트 목록 확인
/moai:git:checkpoint --list

# 시스템 상태 확인
/moai:git:checkpoint --status

# 오래된 체크포인트 정리
/moai:git:checkpoint --cleanup
```

## 특징

- **개인 모드 전용**: 팀 모드에서는 사용 제한으로 충돌 방지
- **자동 타임스탬프**: checkpoint_YYYYMMDD_HHMMSS 형식으로 생성
- **안전한 실험**: 언제든 롤백 가능한 체크포인트 브랜치 생성
- **자동 정리**: 7일 이상 된 체크포인트 자동 삭제 기능
- **브랜치 기반**: Git 브랜치로 관리되어 완전한 상태 복원 가능

체크포인트는 실험적 코드 작성, 리팩토링, 위험한 변경 작업 전에 생성하여 안전한 개발 환경을 제공합니다.
