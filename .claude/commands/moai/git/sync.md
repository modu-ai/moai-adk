---
name: moai:git:sync
description: 🔄 모드별 최적화 원격 동기화 시스템
argument-hint: [ACTION] - push, pull, --auto, --status, --safe 등
allowed-tools: Bash(git:*), Bash(python3:*), Read, Write, Glob, Grep
model: sonnet
---

# MoAI 동기화 시스템

**요청사항**: $ARGUMENTS

모드별 최적화된 안전한 원격 저장소 동기화를 제공합니다.

## 현재 상태 확인

!`echo "=== 동기화 시스템 상태 ==="`
!`echo "📍 현재 브랜치: $(git branch --show-current)"`
!`echo "📤 Push 필요: $(git log origin/$(git branch --show-current)..HEAD --oneline | wc -l)개 커밋"`
!`echo "📥 Pull 필요: $(git log HEAD..origin/$(git branch --show-current) --oneline | wc -l)개 커밋"`
!`echo "🎯 모드: $(python3 -c "import json; config=json.load(open('.moai/config.json')); print(config['project']['mode'])" 2>/dev/null || echo "unknown")"`
!`echo "📝 변경사항: $(git status --porcelain | wc -l)개"`

## 동기화 처리

### 모드별 최적화 동기화

!`python3 .moai/scripts/sync_manager.py $ARGUMENTS`

## 사용법

```bash
# 자동 동기화 (모드별 최적화)
/moai:git:sync --auto

# 원격으로 Push
/moai:git:sync push

# 원격에서 Pull
/moai:git:sync pull

# 안전한 양방향 동기화
/moai:git:sync --safe

# 동기화 상태 확인
/moai:git:sync --status

# 충돌 해결 도움
/moai:git:sync --resolve
```

## 모드별 특징

### 개인 모드

- **전략**: 로컬 우선, 선택적 원격 백업
- **충돌 처리**: 자동 stash 및 rebase
- **백업**: 중요한 작업만 원격 동기화

### 팀 모드

- **전략**: 원격 우선, GitFlow 준수
- **충돌 처리**: 안전한 merge 전략
- **협업**: Pull Request 기반 동기화

## 특징

- **충돌 감지**: 동기화 전 잠재적 충돌 사전 감지
- **자동 백업**: 동기화 전 현재 상태 자동 백업
- **안전 모드**: 충돌 시 사용자 확인 후 진행
- **진행 상황**: 동기화 과정 실시간 표시
- **롤백 지원**: 동기화 실패 시 이전 상태로 복원

## 안전 장치

- 동기화 전 변경사항 자동 커밋 또는 stash
- Force push 방지 및 안전한 push 전략
- 원격 변경사항 검토 후 pull 실행
- 충돌 발생 시 단계적 해결 가이드 제공
