---
name: moai:git:rollback
description: 🔄 안전한 체크포인트 기반 롤백 시스템
argument-hint: [TARGET] - --last, checkpoint_ID, commit_hash, 또는 --list
allowed-tools: Bash(git:*), Bash(python3:*), Read, Write, Glob, Grep
model: sonnet
---

# MoAI 롤백 시스템

**요청사항**: $ARGUMENTS

안전한 체크포인트 기반 롤백으로 언제든 이전 상태로 복원할 수 있습니다.

## 현재 상태 확인

!`echo "=== 롤백 시스템 상태 ==="`
!`echo "📍 현재 브랜치: $(git branch --show-current)"`
!`echo "📝 현재 커밋: $(git log --oneline -1)"`
!`echo "💾 체크포인트: $(git branch | grep -c checkpoint_)개"`
!`echo "🎯 모드: $(python3 -c "import json; config=json.load(open('.moai/config.json')); print(config['project']['mode'])" 2>/dev/null || echo "unknown")"`

## 롤백 처리

### 안전한 체크포인트 롤백

!`python3 .moai/scripts/rollback.py $ARGUMENTS`

## 사용법

```bash
# 마지막 체크포인트로 롤백
/moai:git:rollback --last

# 특정 체크포인트로 롤백
/moai:git:rollback checkpoint_20240122_143000

# 특정 커밋으로 롤백
/moai:git:rollback a1b2c3d

# 롤백 가능한 지점 목록
/moai:git:rollback --list

# 시간 기반 롤백 (1시간 전)
/moai:git:rollback --time "1 hour ago"
```

## 특징

- **안전한 롤백**: 체크포인트 기반으로 안전하게 이전 상태 복원
- **변경사항 보호**: 현재 변경사항을 자동으로 stash에 백업
- **충돌 방지**: 롤백 전 현재 상태 검증 및 안전성 확인
- **복원 정보**: 롤백 후 이전 상태로 다시 돌아갈 수 있는 정보 제공
- **시간 기반 롤백**: 상대적 시간 표현으로 쉬운 롤백 지점 지정

## 주의사항

- 롤백은 되돌릴 수 없는 작업이므로 신중하게 사용하세요
- 팀 모드에서는 원격 브랜치 영향을 고려해야 합니다
- 체크포인트가 없는 경우 Git 기본 reset 기능을 사용합니다
