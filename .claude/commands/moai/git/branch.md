---
name: moai:git:branch
description: 🌿 스마트 브랜치 관리 시스템
argument-hint: [ACTION] [NAME] - create, switch, list, delete, --status 등
allowed-tools: Bash(git:*), Bash(python3:*), Read, Write, Glob, Grep
model: sonnet
---

# MoAI 브랜치 관리 시스템

**요청사항**: $ARGUMENTS

모드별 최적화된 스마트 브랜치 관리를 제공합니다.

## 현재 상태 확인

!`echo "=== 브랜치 상태 ==="`
!`echo "📍 현재 브랜치: $(git branch --show-current)"`
!`echo "📋 로컬 브랜치: $(git branch | wc -l)개"`
!`echo "🌐 원격 브랜치: $(git branch -r | wc -l)개"`
!`echo "🎯 모드: $(python3 -c "import json; config=json.load(open('.moai/config.json')); print(config['project']['mode'])" 2>/dev/null || echo "unknown")"`
!`echo "📚 최근 커밋: $(git log --oneline -1)"`

## 브랜치 작업 처리

### 모드별 최적화 브랜치 관리

!`python3 .moai/scripts/branch_manager.py $ARGUMENTS`

## 사용법

```bash
# 브랜치 목록 확인
/moai:git:branch list

# 새 브랜치 생성 (모드별 자동 네이밍)
/moai:git:branch create "새로운 기능"

# 브랜치 전환
/moai:git:branch switch main

# 브랜치 삭제
/moai:git:branch delete feature/old-branch

# 시스템 상태 확인
/moai:git:branch --status

# 정리 작업
/moai:git:branch clean
```

## 모드별 특징

### 개인 모드

- **브랜치명**: `feature/[설명]`, `experiment/[날짜]`
- **전략**: 자유로운 실험 브랜치 생성
- **정리**: 자동 체크포인트 브랜치 관리

### 팀 모드

- **브랜치명**: `feature/SPEC-XXX-[설명]`
- **전략**: GitFlow 표준 준수
- **추적**: SPEC ID 기반 브랜치 연결

## 특징

- **자동 네이밍**: 모드와 컨텍스트에 따른 지능적 브랜치명 생성
- **안전한 전환**: 변경사항 자동 stash 및 복원
- **충돌 감지**: 브랜치 전환 시 충돌 사전 감지
- **자동 정리**: 병합된 브랜치 자동 정리 기능
- **SPEC 연동**: 팀 모드에서 SPEC ID와 브랜치 자동 연결
