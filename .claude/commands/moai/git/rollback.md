---
name: moai:git:rollback
description: 체크포인트 기반 안전한 롤백 - 이전 상태로 되돌리기
argument-hint: [CHECKPOINT-ID] - 체크포인트 ID 또는 --list, --last, --time="30분전" 옵션
allowed-tools: Bash(git:*), Bash(python3:*), Read, Write, Glob, Grep
model: haiku
---

# MoAI-ADK 롤백 시스템

Safely rollback to previous checkpoints in personal mode.

## Current Environment Check

- Current branch: !`git branch --show-current`
- Working directory status: !`git status --porcelain`
- Project mode: !`python3 -c "import json; config=json.load(open('.moai/config.json')); print(config['project']['mode'])" 2>/dev/null || echo "unknown"`
- Available checkpoints: !`python3 -c "
import json, os
if os.path.exists('.moai/checkpoints/metadata.json'):
    with open('.moai/checkpoints/metadata.json') as f:
        data = json.load(f)
        print(f'{len(data.get(\"checkpoints\", []))} checkpoints available')
else:
    print('0 checkpoints available')
" 2>/dev/null || echo "No metadata file"`

## Task

Rollback to checkpoint: "$ARGUMENTS"

### If --list provided:
- List all available checkpoints with details
- Show: checkpoint ID, timestamp, branch, message, files changed

### If --last provided:
- Rollback to the most recent checkpoint
- Confirm before executing rollback

### If --time provided (e.g., --time="30분전"):
- Find checkpoint closest to specified time
- Show confirmation before rollback

### If checkpoint ID provided:
- Rollback to specific checkpoint
- Validate checkpoint exists before rollback

## Rollback Process:

1. **Validate personal mode**: Only allow rollback in personal mode
2. **Create safety checkpoint**: Backup current state before rollback
3. **Verify checkpoint exists**: Check .moai/checkpoints/metadata.json
4. **Restore from checkpoint**: !`git reset --hard [checkpoint-commit]`
5. **Update working directory**: Ensure clean state after rollback
6. **Log rollback action**: Record rollback in metadata

## 🎯 핵심 기능

### 다양한 롤백 방식
- **체크포인트 ID**: 특정 체크포인트로 정확한 복구
- **시간 기반**: "10분 전", "1시간 전" 등 자연어 지원
- **상대적 위치**: 마지막, 이전, N번째 체크포인트
- **태그 기반**: 중요한 지점 (spec 완료, 테스트 통과 등)

### 사용법

```bash
# 체크포인트 목록 보기
/git:rollback --list

# 마지막 체크포인트로 롤백
/git:rollback --last

# 특정 체크포인트로 롤백
/git:rollback checkpoint_20250120_153000

# 시간 기반 롤백
/git:rollback --time "10분 전"
/git:rollback --time "1시간 전"
/git:rollback --time "오늘 오전"

# 태그 기반 롤백
/git:rollback --tag spec-001-complete
/git:rollback --tag test-passing
```

## 📋 실행 과정

### 1. 체크포인트 목록 조회 (--list)
```bash
# 메타데이터 읽기
METADATA_FILE=".moai/checkpoints/metadata.json"

# 체크포인트 목록 표시
echo "📋 사용 가능한 체크포인트:"
echo "ID                           시간              메시지                   파일수"
echo "checkpoint_20250120_153000   15:30 (30분 전)   JWT 인증 로직 작업 중      5"
echo "checkpoint_20250120_150000   15:00 (1시간 전)  SPEC-001 명세 완료        3"
echo "checkpoint_20250120_140000   14:00 (2시간 전)  초기 프로젝트 설정        12"
```

### 2. 안전성 확인
```bash
# 현재 작업 상태 확인
if git status --porcelain | grep -q .; then
    echo "⚠️ 커밋되지 않은 변경사항이 있습니다."
    echo "다음 중 선택하세요:"
    echo "1. 현재 상태를 체크포인트로 저장 후 롤백"
    echo "2. 변경사항을 버리고 롤백"
    echo "3. 롤백 취소"
fi
```

### 3. 롤백 실행
```bash
# 체크포인트 정보 확인
CHECKPOINT_COMMIT=$(git rev-parse "refs/heads/${CHECKPOINT_ID}")
CHECKPOINT_BRANCH=$(get_checkpoint_branch "${CHECKPOINT_ID}")

# 현재 상태 백업 (선택사항)
if [[ "$BACKUP_CURRENT" == "yes" ]]; then
    /git:checkpoint "롤백 전 백업: $(date)"
fi

# 롤백 수행
git reset --hard "${CHECKPOINT_COMMIT}"

# 작업 디렉토리 정리
git clean -fd
```

## 🔧 고급 롤백 기능

### 시간 기반 롤백
```bash
# 시간 파싱 함수
parse_time_expression() {
    local time_expr="$1"
    case "$time_expr" in
        *"분 전")
            minutes=$(echo "$time_expr" | grep -o '[0-9]\+')
            target_time=$(date -d "-${minutes} minutes" +%s)
            ;;
        *"시간 전")
            hours=$(echo "$time_expr" | grep -o '[0-9]\+')
            target_time=$(date -d "-${hours} hours" +%s)
            ;;
        "오늘 오전")
            target_time=$(date -d "today 09:00" +%s)
            ;;
    esac
}
```

### 스마트 매칭
```bash
# 가장 가까운 체크포인트 찾기
find_closest_checkpoint() {
    local target_time="$1"
    local closest_id=""
    local min_diff=999999999

    while IFS= read -r checkpoint; do
        local cp_time=$(echo "$checkpoint" | jq -r '.timestamp')
        local cp_timestamp=$(date -d "$cp_time" +%s)
        local diff=$((target_time - cp_timestamp))

        if [[ $diff -ge 0 && $diff -lt $min_diff ]]; then
            min_diff=$diff
            closest_id=$(echo "$checkpoint" | jq -r '.id')
        fi
    done < <(jq -c '.checkpoints[]' "$METADATA_FILE")

    echo "$closest_id"
}
```

## 📊 롤백 종류별 특징

### 1. 소프트 롤백 (기본값)
- **변경사항 보존**: 작업 디렉토리 파일 유지
- **스테이징 초기화**: git add된 내용 해제
- **안전한 복구**: 실수 시 쉽게 되돌리기 가능

### 2. 하드 롤백 (--hard)
- **완전 초기화**: 모든 변경사항 삭제
- **정확한 복구**: 체크포인트 시점과 동일한 상태
- **위험성 경고**: 삭제된 내용 복구 불가

### 3. 혼합 롤백 (--mixed)
- **선택적 복구**: 특정 파일만 롤백
- **부분 적용**: 일부 변경사항만 되돌리기
- **세밀한 제어**: 고급 사용자용

## 🎯 모드별 롤백 전략

### 개인 모드 (Personal Mode)
```bash
# 자유로운 실험 지원
- 빈번한 체크포인트 활용
- 실험 실패 시 즉시 롤백
- 손실 없는 안전한 개발

# 롤백 후 자동 정리
- 불필요한 체크포인트 정리
- 브랜치 히스토리 최적화
```

### 팀 모드 (Team Mode)
```bash
# 신중한 롤백 처리
- 팀원에게 롤백 사실 알림
- PR 상태 확인 후 롤백
- 원격 브랜치 동기화 고려

# 롤백 이력 관리
- 롤백 사유 문서화
- 팀 공유용 롤백 보고서
```

## 🚨 안전장치 및 검증

### 롤백 전 검증
```bash
# 브랜치 상태 확인
check_branch_safety() {
    # 원격 브랜치와 동기화 상태 확인
    # 진행 중인 merge/rebase 감지
    # 보호된 브랜치 여부 확인
}

# 체크포인트 유효성 검증
validate_checkpoint() {
    # 체크포인트 존재 여부
    # 커밋 해시 유효성
    # 브랜치 접근 가능성
}
```

### 롤백 후 검증
```bash
# 롤백 성공 확인
verify_rollback() {
    # 타겟 커밋으로 이동 확인
    # 작업 디렉토리 상태 검증
    # 손실된 데이터 없음 확인
}
```

## 📈 통계 및 모니터링

### 롤백 히스토리
```json
{
  "rollback_history": [
    {
      "timestamp": "2025-01-20T16:00:00Z",
      "from": "a1b2c3d",
      "to": "checkpoint_20250120_153000",
      "reason": "실험 코드 롤백",
      "mode": "personal",
      "files_affected": 5
    }
  ]
}
```

### 롤백 패턴 분석
- 자주 롤백되는 파일/영역 식별
- 실험 성공률 통계
- 개발 패턴 개선 제안

## 💡 사용 패턴 및 팁

### 일반적인 롤백 시나리오
```bash
# 실험 실패 후 롤백
/git:rollback --last
# → 마지막 안전한 상태로 즉시 복구

# 잘못된 리팩토링 롤백
/git:rollback --tag "refactor-start"
# → 리팩토링 시작 전으로 되돌리기

# 특정 시점으로 롤백
/git:rollback --time "점심 먹기 전"
# → 자연어로 시점 지정

# 신중한 롤백 (검토 후)
/git:rollback --list
/git:rollback checkpoint_20250120_120000
# → 목록 확인 후 정확한 체크포인트 선택
```

### Constitution 5원칙 준수
1. **Simplicity**: 한 명령어로 모든 롤백 처리
2. **Architecture**: git-manager와 체계적 연동
3. **Testing**: 안전한 실험 환경 제공
4. **Observability**: 모든 롤백 추적 및 로깅
5. **Versioning**: 체크포인트 기반 버전 관리

모든 롤백 작업은 git-manager 에이전트가 안전하게 처리합니다.