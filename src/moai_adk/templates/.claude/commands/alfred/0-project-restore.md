---
name: alfred:0-project restore
description: Checkpoint 복구 (Event-Driven 자동 백업 시스템)
allowed-tools:
  - Read
  - Bash(git:*)
  - Bash(cat:*)
  - Bash(ls:*)
  - TodoWrite
---

# 🛡️ MoAI-ADK Checkpoint 복구 시스템

## 🎯 커맨드 목적

Event-Driven Checkpoint 시스템으로 생성된 자동 백업을 조회하고 복구합니다.

---

## 📊 Checkpoint 시스템 개요

MoAI-ADK는 위험한 작업 전 자동으로 checkpoint를 생성합니다:

### 자동 Checkpoint 생성 트리거

| 작업 유형 | 감지 조건 | Checkpoint 이름 |
|---------|--------|---------------|
| **대규모 삭제** | `rm -rf`, `git rm` | `before-delete-{timestamp}` |
| **Git 병합** | `git merge`, `git reset --hard` | `before-merge-{timestamp}` |
| **스크립트 실행** | `python`, `node`, `bash` | `before-script-{timestamp}` |
| **중요 파일 수정** | `CLAUDE.md`, `config.json` | `before-critical-file-{timestamp}` |
| **대규모 리팩토링** | ≥10개 파일 동시 수정 | `before-refactor-{timestamp}` |

### Checkpoint 특징

- **Local branch**: 원격 저장소 오염 방지 (로컬 전용)
- **최대 10개 유지**: FIFO + 7일 제한 (자동 정리)
- **투명한 동작**: 백그라운드에서 자동 생성, 사용자에게 알림

---

## 🔍 STEP 1: Checkpoint 목록 조회

### 방법 1: SessionStart 메시지 확인 (권장)

Claude Code 세션 시작 시 자동으로 최근 3개 checkpoint를 표시합니다:

```
🚀 MoAI-ADK Session Started
   Language: python
   Branch: develop (c3c48ac)
   Changes: 5
   SPEC Progress: 17/17 (100%)
   Checkpoints: 5 available
      - delete-20251015-143000
      - merge-20251015-142500
      - critical-file-20251015-140000
   Restore: /alfred:0-project restore
```

### 방법 2: 로그 파일 직접 확인

```bash
cat .moai/checkpoints.log
```

**로그 형식 (JSON Lines)**:
```json
{"timestamp": "2025-10-15T14:30:00", "branch": "before-delete-20251015-143000", "operation": "delete"}
{"timestamp": "2025-10-15T14:25:00", "branch": "before-merge-20251015-142500", "operation": "merge"}
```

### 방법 3: Git 브랜치 목록 확인

```bash
git branch | grep "^  before-"
```

---

## 🔄 STEP 2: Checkpoint 복구 방법

### 옵션 1: 브랜치 전환 (안전, 권장)

현재 작업을 보존하면서 checkpoint로 이동합니다.

```bash
# Checkpoint 브랜치로 전환
git checkout before-delete-20251015-143000

# 현재 상태 확인
git log --oneline -5

# 원래 브랜치로 돌아가기
git checkout develop
```

**장점**:
- 현재 작업 보존
- 안전하게 과거 상태 확인
- 언제든 돌아올 수 있음

**단점**:
- 브랜치 전환 overhead

### 옵션 2: 하드 리셋 (강력, 주의)

현재 브랜치를 checkpoint 시점으로 되돌립니다.

⚠️ **경고**: 현재 커밋되지 않은 변경사항이 모두 삭제됩니다!

```bash
# 현재 변경사항 백업 (선택)
git stash push -m "Before reset to checkpoint"

# Checkpoint로 하드 리셋
git reset --hard before-delete-20251015-143000

# 원래 HEAD 복구 (필요 시)
git reset --hard HEAD@{1}
```

**장점**:
- 빠른 복구
- 브랜치 전환 없음

**단점**:
- 현재 변경사항 손실 위험

### 옵션 3: Cherry-pick (선택적 복구)

특정 파일만 checkpoint에서 복구합니다.

```bash
# 특정 파일만 복구
git checkout before-delete-20251015-143000 -- path/to/file.py

# 여러 파일 복구
git checkout before-delete-20251015-143000 -- src/ tests/
```

**장점**:
- 선택적 복구
- 다른 변경사항 보존

---

## 🧹 STEP 3: Checkpoint 정리

### 자동 정리 (권장)

Checkpoint는 다음 조건에서 자동 정리됩니다:
- 10개 초과 시 가장 오래된 것부터 삭제 (FIFO)
- 7일 경과 시 자동 삭제

### 수동 삭제

```bash
# 특정 checkpoint 삭제
git branch -d before-delete-20251015-143000

# 모든 checkpoint 삭제 (주의!)
git branch | grep "^  before-" | xargs git branch -D
```

---

## 📝 사용 예시

### 예시 1: 실수로 파일 삭제 후 복구

```bash
# 1. Alfred가 자동으로 checkpoint 생성
# 🛡️ Checkpoint created: before-delete-20251015-143000

# 2. 실수로 파일 삭제 (예: rm -rf src/)
rm -rf src/

# 3. Checkpoint로 복구
git checkout before-delete-20251015-143000

# 4. 삭제된 파일 확인
ls src/

# 5. 현재 브랜치에 복구
git checkout develop
git checkout before-delete-20251015-143000 -- src/

# 6. 커밋
git add src/
git commit -m "🔧 FIX: 실수로 삭제된 src/ 복구"
```

### 예시 2: Git 병합 실패 후 롤백

```bash
# 1. Alfred가 자동으로 checkpoint 생성
# 🛡️ Checkpoint created: before-merge-20251015-142500

# 2. 병합 시도
git merge feature/new-feature

# 3. 충돌 발생 또는 병합 실패

# 4. Checkpoint로 롤백
git reset --hard before-merge-20251015-142500

# 5. 상태 확인
git log --oneline -5
```

### 예시 3: 중요 파일 수정 후 되돌리기

```bash
# 1. Alfred가 자동으로 checkpoint 생성
# 🛡️ Checkpoint created: before-critical-file-20251015-140000

# 2. CLAUDE.md 수정
# (잘못된 수정...)

# 3. 원래 버전 복구
git checkout before-critical-file-20251015-140000 -- CLAUDE.md

# 4. 변경사항 확인
git diff CLAUDE.md
```

---

## ⚠️ 주의사항

### Checkpoint 사용 시 유의점

1. **Local 전용**: Checkpoint는 원격에 push되지 않습니다 (로컬만 유지)
2. **임시 백업**: 영구 백업이 아닌 임시 안전망입니다 (7일 후 자동 삭제)
3. **Dirty Working Directory**: 커밋되지 않은 변경사항이 있어도 checkpoint 생성 가능

### 복구 전 확인사항

- [ ] 현재 브랜치 확인 (`git branch`)
- [ ] 커밋되지 않은 변경사항 확인 (`git status`)
- [ ] 필요 시 현재 작업 백업 (`git stash`)
- [ ] Checkpoint 시점 확인 (`.moai/checkpoints.log`)

---

## 🔗 관련 시스템

### PreToolUse Hook 통합

Checkpoint는 Claude Code의 PreToolUse hook에서 자동으로 생성됩니다:

```python
# .claude/hooks/alfred/moai_hooks.py

def handle_pre_tool_use(payload):
    """위험한 작업 전 자동 checkpoint 생성"""
    is_risky, operation_type = detect_risky_operation(tool, args)

    if is_risky:
        checkpoint_branch = create_checkpoint(cwd, operation_type)
        # 🛡️ Checkpoint created: ...
```

### Checkpoint 로그

모든 checkpoint는 `.moai/checkpoints.log`에 기록됩니다:

```bash
# 로그 파일 위치
.moai/checkpoints.log

# 최근 10개 checkpoint 조회
tail -10 .moai/checkpoints.log | jq
```

---

## 📊 Checkpoint 통계

### Checkpoint 개수 확인

```bash
# 현재 checkpoint 개수
git branch | grep "^  before-" | wc -l

# 로그 파일 라인 수
wc -l .moai/checkpoints.log
```

### Checkpoint별 디스크 사용량

```bash
# 각 checkpoint 브랜치의 커밋 개수
for branch in $(git branch | grep "before-"); do
    echo "$branch: $(git rev-list --count $branch)"
done
```

---

## 🛠️ 문제 해결

### 문제 1: Checkpoint가 생성되지 않음

**증상**: 위험한 작업을 수행해도 checkpoint가 생성되지 않음

**원인**: Git 리포지토리가 아니거나 PreToolUse hook 비활성화

**해결**:
```bash
# Git 리포지토리 확인
git rev-parse --git-dir

# Hook 설정 확인
cat .claude/settings.json | jq '.hooks.PreToolUse'

# Hook 스크립트 권한 확인
ls -la .claude/hooks/alfred/moai_hooks.py
```

### 문제 2: Checkpoint 복구 실패

**증상**: `git checkout before-...` 실행 시 에러

**원인**: 커밋되지 않은 변경사항 충돌

**해결**:
```bash
# 현재 변경사항 백업
git stash push -m "Backup before checkpoint restore"

# Checkpoint 복구
git checkout before-delete-20251015-143000

# 백업 복원 (필요 시)
git stash pop
```

### 문제 3: 로그 파일 손상

**증상**: `.moai/checkpoints.log` 파싱 에러

**해결**:
```bash
# 로그 파일 백업
cp .moai/checkpoints.log .moai/checkpoints.log.bak

# Git 브랜치 목록으로 로그 재생성
git branch | grep "^  before-" | while read branch; do
    timestamp=$(git log -1 --format=%cI $branch)
    operation=$(echo $branch | sed 's/before-//' | sed 's/-[0-9].*$//')
    echo "{\"timestamp\": \"$timestamp\", \"branch\": \"$branch\", \"operation\": \"$operation\"}"
done > .moai/checkpoints.log
```

---

## 다음 단계

Checkpoint 복구 후:

1. **변경사항 검토**: `git diff` 또는 `git log`로 확인
2. **테스트 실행**: 복구 후 정상 동작 확인
3. **커밋 및 Push**: 필요 시 원격에 반영
4. **다음 작업**: `/alfred:1-spec`, `/alfred:2-build` 등 계속 진행

---

## 관련 문서

- **SPEC**: `.moai/specs/SPEC-CHECKPOINT-EVENT-001/spec.md`
- **구현**: `src/moai_adk/core/git/checkpoint.py`
- **Hook**: `.claude/hooks/alfred/moai_hooks.py`
- **설정**: `.moai/config.json` (git_strategy.checkpoint_*)
