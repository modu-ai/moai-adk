---
name: moai:git:branch
description: 스마트 브랜치 관리 시스템 (모드별 최적화)
argument-hint: [ACTION] - create, switch, list, clean, --status, --personal, --team 중 하나
allowed-tools: Bash(git:*), Bash(python3:*), Read, Write, Glob, Grep
model: haiku
---

# MoAI-ADK 브랜치 관리 시스템

**브랜치 작업**: $ARGUMENTS

모드별 최적화 전략으로 스마트한 브랜치 관리를 수행합니다.

## 현재 상태 확인

브랜치 상태를 확인합니다:

!`git branch --show-current`
!`git branch -l | wc -l`
!`git branch -r | wc -l`
!`python3 -c "import json; config=json.load(open('.moai/config.json')); print(config['project']['mode'])" 2>/dev/null || echo "unknown"`
!`git log --oneline -3`

## 브랜치 작업 실행

요청된 브랜치 작업을 수행합니다: "$ARGUMENTS"

### 브랜치 작업 종류:

**"create" 제공 시**:

- 현재 모드에 따른 새 브랜치 생성
- 개인 모드: feature/[설명] 형식
- 팀 모드: feature/SPEC-XXX-[설명] 형식

**"switch" 제공 시**:

- 지정된 브랜치로 안전하게 전환
- 필요시 변경사항 stash 처리
- 작업 디렉토리 업데이트

**"list" 제공 시**:

- 모든 브랜치와 상태 표시
- 현재 브랜치 강조
- 각 브랜치의 마지막 커밋 표시

**"clean" 제공 시**:

- 병합된 브랜치 정리
- 오래된 원격 추적 브랜치 제거
- 중요한 브랜치 보존

**"--status" 제공 시**:

- 상세한 브랜치 상태 표시
- 표시: 현재 브랜치, 앞서거나 뒤램어진 커밋 수, 작업 트리 상태

**"--personal" 제공 시**:

- 개인 모드용 브랜치 전략 설정
- 단순화된 브랜치 명명 설정

**"--team" 제공 시**:

- 팀 모드용 브랜치 전략 설정
- GitFlow 호환 브랜치 명명 설정

## 🎯 핵심 기능

### 모드별 브랜치 전략

- **개인 모드**: 간소화된 브랜치, 실험 지향
- **팀 모드**: 구조화된 GitFlow, 협업 최적화
- **자동 명명**: 작업 내용 기반 브랜치명 생성
- **스마트 정리**: 불필요한 브랜치 자동 정리

### 사용법

```bash
# 자동 브랜치 생성 (MoAI 워크플로우 연동)
/git:branch --auto "JWT 인증 기능"

# 수동 브랜치 생성
/git:branch create feature/user-auth

# 브랜치 전환
/git:branch switch feature/user-auth

# 브랜치 목록 보기
/git:branch list

# 브랜치 정리
/git:branch clean
```

## 📋 모드별 브랜치 전략

### 개인 모드 (Personal Mode)

#### 브랜치 구조

```
main
├── experiment/jwt-auth-2025-01-20    # 실험용 브랜치
├── feature/user-management           # 기능 브랜치
└── checkpoint/*                     # 체크포인트 브랜치 (자동 생성)
```

#### 명명 규칙

```bash
# 자동 생성 패턴
- feature/{description}               # 일반 기능
- experiment/{description}-{date}     # 실험적 기능
- fix/{issue-description}            # 버그 수정
- refactor/{target-component}        # 리팩토링

# 예시
- feature/jwt-authentication
- experiment/new-algorithm-20250120
- fix/login-validation-error
- refactor/user-service
```

#### 특징

- **자유로운 실험**: experiment/ 브랜치로 안전한 실험
- **간단한 병합**: main 브랜치로 직접 병합
- **자동 정리**: 오래된 실험 브랜치 자동 삭제
- **체크포인트 연동**: 체크포인트와 브랜치 연계

### 팀 모드 (Team Mode)

#### 브랜치 구조 (GitFlow)

```
main                                 # 프로덕션 코드
├── develop                          # 개발 통합 브랜치
│   ├── feature/SPEC-001-user-auth  # 기능 브랜치
│   ├── feature/SPEC-002-dashboard  # 기능 브랜치
│   └── hotfix/critical-bug-fix     # 핫픽스 브랜치
├── release/v1.2.0                  # 릴리즈 브랜치
└── hotfix/security-patch           # 긴급 수정
```

#### 명명 규칙 (MoAI 표준)

```bash
# SPEC 기반 브랜치
- feature/SPEC-{XXX}-{description}   # 명세 기반 기능
- hotfix/ISSUE-{XXX}-{description}   # 이슈 기반 수정
- release/v{MAJOR}.{MINOR}.{PATCH}   # 릴리즈 버전

# 예시
- feature/SPEC-001-jwt-authentication
- hotfix/ISSUE-042-session-timeout
- release/v1.2.0
```

## 🔧 자동 브랜치 생성

### MoAI 워크플로우 연동

```bash
# /moai:1-spec과 연동된 자동 브랜치 생성
auto_create_branch() {
    local description="$1"
    local mode=$(get_project_mode)

    if [[ "$mode" == "personal" ]]; then
        # 개인 모드: 간단한 브랜치명
        local branch_name="feature/$(echo "$description" | tr ' ' '-' | tr '[:upper:]' '[:lower:]')"
    else
        # 팀 모드: SPEC 기반 브랜치명
        local spec_id=$(get_next_spec_id)
        local branch_name="feature/SPEC-${spec_id}-$(echo "$description" | tr ' ' '-' | tr '[:upper:]' '[:lower:]')"
    fi

    # 브랜치 생성 및 전환
    git checkout -b "$branch_name"
    echo "🌿 브랜치 생성: $branch_name"

    # 메타데이터 저장
    save_branch_metadata "$branch_name" "$description" "$mode"
}
```

### 스마트 브랜치명 생성

```bash
generate_smart_branch_name() {
    local description="$1"
    local type="$2"  # feature, fix, experiment, etc.

    # 키워드 추출 및 정제
    local keywords=$(echo "$description" | \
        grep -oE '\b[a-zA-Z]{3,}\b' | \
        head -3 | \
        tr '\n' '-' | \
        sed 's/-$//')

    # 중복 방지
    local counter=1
    local base_name="${type}/${keywords}"
    local branch_name="$base_name"

    while git show-ref --verify --quiet "refs/heads/$branch_name"; do
        branch_name="${base_name}-${counter}"
        ((counter++))
    done

    echo "$branch_name"
}
```

## 📊 브랜치 상태 관리

### 브랜치 목록 표시

```bash
show_branch_list() {
    echo "🌳 브랜치 목록"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    # 현재 브랜치 강조
    local current_branch=$(git branch --show-current)
    echo "📍 현재: $current_branch"
    echo ""

    # 로컬 브랜치 목록
    echo "📂 로컬 브랜치:"
    git branch --format="%(if:equals=refs/heads/$current_branch)%(refname)%(then)* %(else)  %(end)%(refname:short)%09%(committerdate:relative)%09%(subject)" | \
        head -10

    echo ""

    # 원격 브랜치 상태
    echo "🌐 원격 브랜치 상태:"
    git for-each-ref --format="%(refname:short)%09%(upstream:track)" refs/heads | \
        grep -v "^$current_branch" | head -5

    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}
```

### 브랜치 메타데이터 관리

```json
// .moai/branches/metadata.json
{
  "branches": [
    {
      "name": "feature/jwt-authentication",
      "created": "2025-01-20T15:30:00Z",
      "description": "JWT 기반 사용자 인증 시스템",
      "mode": "personal",
      "spec_id": "SPEC-001",
      "status": "active",
      "checkpoints": [
        "checkpoint_20250120_153000",
        "checkpoint_20250120_160000"
      ]
    }
  ]
}
```

## 🧹 자동 브랜치 정리

### 정리 기준

```bash
# 개인 모드 정리 정책
personal_cleanup_policy() {
    # 1. 30일 이상 된 experiment 브랜치
    # 2. main에 병합된 feature 브랜치
    # 3. 7일 이상 된 checkpoint 브랜치
    # 4. 빈 브랜치 (변경사항 없음)
}

# 팀 모드 정리 정책
team_cleanup_policy() {
    # 1. develop에 병합된 feature 브랜치
    # 2. 릴리즈 완료된 release 브랜치
    # 3. 적용 완료된 hotfix 브랜치
    # 4. 60일 이상 된 stale 브랜치
}
```

### 안전한 브랜치 삭제

```bash
safe_branch_cleanup() {
    echo "🧹 브랜치 정리 시작"

    # 정리 대상 브랜치 식별
    local cleanup_candidates=($(identify_cleanup_candidates))

    for branch in "${cleanup_candidates[@]}"; do
        echo "🔍 브랜치 분석: $branch"

        # 병합 상태 확인
        if is_branch_merged "$branch"; then
            echo "✅ 병합됨 - 삭제 대상"
            delete_branch_safely "$branch"
        else
            echo "⚠️ 미병합 - 백업 후 삭제"
            backup_and_delete_branch "$branch"
        fi
    done
}
```

## 🔄 브랜치 전환 최적화

### 스마트 전환

```bash
smart_branch_switch() {
    local target_branch="$1"

    # 1. 현재 작업 상태 확인
    if git status --porcelain | grep -q .; then
        echo "💾 변경사항 있음 - 체크포인트 생성"
        /git:checkpoint "브랜치 전환 전 백업"
    fi

    # 2. 브랜치 존재 확인
    if ! git show-ref --verify --quiet "refs/heads/$target_branch"; then
        echo "❓ 브랜치가 없습니다. 생성하시겠습니까? (y/n)"
        read -r create_branch
        if [[ "$create_branch" == "y" ]]; then
            git checkout -b "$target_branch"
        else
            return 1
        fi
    else
        git checkout "$target_branch"
    fi

    # 3. 브랜치 전환 후 동기화
    if [[ $(get_project_mode) == "team" ]]; then
        echo "🔄 팀 모드 - 자동 동기화 수행"
        /git:sync --pull
    fi

    echo "🌿 브랜치 전환 완료: $target_branch"
}
```

## 📈 브랜치 통계 및 분석

### 브랜치 활동 분석

```json
{
  "branch_statistics": {
    "total_branches": 12,
    "active_branches": 5,
    "merged_branches": 7,
    "avg_branch_lifetime": "5.2 days",
    "most_active_type": "feature",
    "cleanup_savings": "67% storage reduction"
  },
  "branch_patterns": {
    "common_prefixes": ["feature/", "experiment/", "fix/"],
    "naming_consistency": "92%",
    "merge_success_rate": "98%"
  }
}
```

## 🎯 Constitution 5원칙 준수

1. **Simplicity**: 복잡한 Git 브랜치 작업을 단순화
2. **Architecture**: 모드별 브랜치 전략 체계화
3. **Testing**: 실험 브랜치로 안전한 테스트 환경
4. **Observability**: 모든 브랜치 활동 추적
5. **Versioning**: 체계적인 브랜치 버전 관리

## 💡 사용 시나리오

### 개인 개발 패턴

```bash
# 새 기능 시작
/git:branch --auto "사용자 대시보드"
# → feature/user-dashboard 생성 및 전환

# 실험적 기능 시도
/git:branch create experiment/new-ui-framework
# → 안전한 실험 환경

# 작업 완료 후 정리
/git:branch clean
# → 불필요한 브랜치 자동 정리
```

### 팀 개발 패턴

```bash
# SPEC 기반 브랜치 생성 (자동)
/moai:1-spec "JWT 인증 시스템"
# → feature/SPEC-001-jwt-auth 자동 생성

# 브랜치 상태 확인
/git:branch list
# → 팀 브랜치 현황 파악

# 정기 브랜치 정리
/git:branch clean
# → 병합된 브랜치 안전하게 정리
```

모든 브랜치 관리는 git-manager 에이전트와 연동되어 자동화됩니다.
