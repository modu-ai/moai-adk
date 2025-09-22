---
name: git:commit
description: 스마트 커밋 시스템 - Constitution 5원칙 기반 자동 커밋 메시지 생성
argument-hint: [message|--auto|--spec|--build|--sync]
allowed-tools: Bash(git:*), Read, Write, Glob, Grep
---

# Git 스마트 커밋 시스템

Constitution 5원칙을 준수하고 16-Core @TAG 시스템과 연동된 자동 커밋 메시지 생성을 제공합니다.

## 🎯 핵심 기능

### 자동 커밋 메시지 생성
- **컨텍스트 기반**: 변경된 파일 분석으로 의미 있는 메시지
- **@TAG 연동**: 16-Core TAG 시스템 자동 적용
- **MoAI 워크플로우**: spec/build/sync 단계별 커밋
- **Constitution 준수**: 5원칙 기반 메시지 구조

### 사용법

```bash
# 자동 커밋 메시지 생성
/git:commit --auto

# MoAI 워크플로우 연동 커밋
/git:commit --spec    # SPEC 단계 커밋
/git:commit --build   # BUILD 단계 커밋
/git:commit --sync    # SYNC 단계 커밋

# 수동 메시지 커밋
/git:commit "JWT 인증 로직 구현 완료"

# 체크포인트 커밋
/git:commit --checkpoint "실험 중간 백업"
```

## 📋 커밋 메시지 구조

### MoAI 표준 형식
```
{이모지} {SPEC-ID}: {작업 내용}

{상세 설명}
- 주요 변경사항 1
- 주요 변경사항 2

@TAG:{CATEGORY}-{DESCRIPTION}-{NUMBER}

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

### 예시 커밋 메시지
```
📝 SPEC-001: JWT 기반 사용자 인증 명세 작성

EARS 형식으로 사용자 인증 시스템 요구사항 정의
- 로그인/로그아웃 시나리오 추가
- 토큰 갱신 프로세스 명세
- 권한 기반 접근 제어 정의

@REQ:USER-AUTH-001
@DESIGN:JWT-TOKEN-001
@TASK:AUTH-IMPL-001

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## 🔧 자동 메시지 생성 로직

### 변경사항 분석
```bash
analyze_changes() {
    local staged_files=($(git diff --cached --name-only))
    local change_summary=""
    local emoji=""
    local tag_category=""

    # 파일 유형별 분석
    for file in "${staged_files[@]}"; do
        case "$file" in
            *.md|*SPEC*|*spec*)
                change_summary="명세 작성"
                emoji="📝"
                tag_category="REQ"
                ;;
            *test*|*Test*)
                change_summary="테스트 추가"
                emoji="🧪"
                tag_category="TEST"
                ;;
            *.py|*.js|*.ts|*.java)
                change_summary="구현 완료"
                emoji="✨"
                tag_category="FEATURE"
                ;;
            *config*|*.json|*.yml)
                change_summary="설정 업데이트"
                emoji="⚙️"
                tag_category="TECH"
                ;;
        esac
    done

    echo "$emoji:$change_summary:$tag_category"
}
```

### 컨텍스트별 메시지 생성
```bash
generate_auto_message() {
    local analysis_result="$1"
    local emoji=$(echo "$analysis_result" | cut -d: -f1)
    local summary=$(echo "$analysis_result" | cut -d: -f2)
    local tag_category=$(echo "$analysis_result" | cut -d: -f3)

    # SPEC ID 추출 (현재 브랜치 또는 최근 작업에서)
    local spec_id=$(extract_spec_id)

    # 변경된 파일 목록
    local changed_files=($(git diff --cached --name-only))
    local file_count=${#changed_files[@]}

    # 메시지 구성
    local message="${emoji} ${spec_id}: ${summary}"

    # 상세 내용 추가
    if [[ $file_count -gt 1 ]]; then
        message+="\n\n변경된 파일 ${file_count}개:"
        for file in "${changed_files[@]}"; do
            message+="\n- $(basename "$file")"
        done
    fi

    # @TAG 추가
    local tag=$(generate_context_tag "$tag_category" "$spec_id")
    message+="\n\n$tag"

    # MoAI 서명 추가
    message+="\n\n🤖 Generated with [Claude Code](https://claude.ai/code)\n\nCo-Authored-By: Claude <noreply@anthropic.com>"

    echo "$message"
}
```

## 🎭 MoAI 워크플로우 연동

### SPEC 단계 커밋
```bash
commit_spec_stage() {
    local stage="$1"  # 1, 2, 3, 4
    local spec_id=$(get_current_spec_id)

    case "$stage" in
        1)
            local message="📝 ${spec_id}: 초기 명세 작성 완료"
            local detail="EARS 형식으로 기본 요구사항 정의"
            ;;
        2)
            local message="📖 ${spec_id}: User Stories 및 시나리오 추가"
            local detail="Given-When-Then 시나리오 작성 완료"
            ;;
        3)
            local message="✅ ${spec_id}: 수락 기준 정의 완료"
            local detail="측정 가능한 수락 기준 및 테스트 조건 추가"
            ;;
        4)
            local message="🎯 ${spec_id}: 명세 완성 및 구조 생성"
            local detail="프로젝트 구조 및 초기 파일 생성"
            ;;
    esac

    # @TAG 생성
    local tag="@REQ:$(echo ${spec_id} | tr '-' '_')-$(printf "%03d" $stage)"

    commit_with_message "$message" "$detail" "$tag"
}
```

### BUILD 단계 커밋 (TDD)
```bash
commit_build_stage() {
    local phase="$1"  # RED, GREEN, REFACTOR
    local spec_id=$(get_current_spec_id)

    case "$phase" in
        "RED")
            local message="🔴 ${spec_id}: 실패 테스트 작성"
            local detail="TDD RED 단계 - 테스트 먼저 작성"
            local tag="@TEST:UNIT-$(echo ${spec_id} | tr '-' '_')-RED"
            ;;
        "GREEN")
            local message="🟢 ${spec_id}: 최소 구현 완료"
            local detail="TDD GREEN 단계 - 테스트 통과하는 최소 코드"
            local tag="@FEATURE:IMPL-$(echo ${spec_id} | tr '-' '_')-GREEN"
            ;;
        "REFACTOR")
            local message="♻️ ${spec_id}: 코드 품질 개선"
            local detail="TDD REFACTOR 단계 - 품질 향상 및 최적화"
            local tag="@DEBT:REFACTOR-$(echo ${spec_id} | tr '-' '_')-CLEAN"
            ;;
    esac

    commit_with_message "$message" "$detail" "$tag"
}
```

### SYNC 단계 커밋
```bash
commit_sync_stage() {
    local spec_id=$(get_current_spec_id)
    local mode=$(get_project_mode)

    local message="📚 ${spec_id}: 문서 동기화 완료"
    local detail="Living Document 동기화 및 PR 준비"

    if [[ "$mode" == "team" ]]; then
        detail+="\n- Draft PR → Ready for Review 전환"
        detail+="\n- 팀 리뷰 준비 완료"
    else
        detail+="\n- 개인 모드 문서 정리"
        detail+="\n- 로컬 작업 완료"
    fi

    local tag="@DOC:SYNC-$(echo ${spec_id} | tr '-' '_')-COMPLETE"

    commit_with_message "$message" "$detail" "$tag"
}
```

## 📊 16-Core @TAG 시스템 연동

### TAG 자동 생성
```bash
generate_context_tag() {
    local category="$1"
    local spec_id="$2"
    local context="$3"

    # SPEC ID에서 번호 추출
    local spec_num=$(echo "$spec_id" | grep -o '[0-9]\+' | head -1)
    local padded_num=$(printf "%03d" "$spec_num")

    case "$category" in
        "REQ")
            echo "@REQ:${context}-${padded_num}"
            ;;
        "DESIGN")
            echo "@DESIGN:${context}-${padded_num}"
            ;;
        "TASK")
            echo "@TASK:${context}-${padded_num}"
            ;;
        "TEST")
            echo "@TEST:${context}-${padded_num}"
            ;;
        "FEATURE")
            echo "@FEATURE:${context}-${padded_num}"
            ;;
        "DOC")
            echo "@DOC:${context}-${padded_num}"
            ;;
    esac
}
```

### TAG 체인 추적
```bash
update_tag_chain() {
    local new_tag="$1"
    local spec_id="$2"

    # .moai/indexes/tags.json 업데이트
    local tags_file=".moai/indexes/tags.json"
    if [[ -f "$tags_file" ]]; then
        jq --arg spec "$spec_id" --arg tag "$new_tag" \
           '.specs[$spec].tags += [$tag]' \
           "$tags_file" > "$tags_file.tmp" && \
           mv "$tags_file.tmp" "$tags_file"
    fi
}
```

## 🔍 스마트 커밋 분석

### 커밋 내용 추론
```bash
infer_commit_intent() {
    local staged_files=($(git diff --cached --name-only))
    local added_lines=$(git diff --cached --numstat | awk '{sum+=$1} END {print sum}')
    local deleted_lines=$(git diff --cached --numstat | awk '{sum+=$2} END {print sum}')

    # 변경 규모 분석
    if [[ $added_lines -gt 100 ]]; then
        echo "major_feature"
    elif [[ $deleted_lines -gt $added_lines ]]; then
        echo "refactoring"
    elif [[ ${#staged_files[@]} -eq 1 ]]; then
        echo "focused_change"
    else
        echo "multi_file_update"
    fi
}
```

### 커밋 품질 검증
```bash
validate_commit_quality() {
    local message="$1"

    # Constitution 5원칙 준수 확인
    local checks=()

    # 1. Simplicity: 메시지가 명확한가?
    if [[ ${#message} -lt 20 ]]; then
        checks+=("❌ 메시지가 너무 짧습니다 (최소 20자)")
    fi

    # 2. Architecture: @TAG가 포함되었는가?
    if ! echo "$message" | grep -q "@"; then
        checks+=("⚠️ @TAG가 누락되었습니다")
    fi

    # 3. Testing: 테스트 관련 변경사항이 있는가?
    local test_files=($(git diff --cached --name-only | grep -i test))
    if [[ ${#test_files[@]} -eq 0 ]] && git diff --cached --name-only | grep -qE '\.(py|js|ts|java)$'; then
        checks+=("⚠️ 코드 변경에 테스트가 포함되지 않았습니다")
    fi

    # 검증 결과 반환
    if [[ ${#checks[@]} -eq 0 ]]; then
        echo "✅ 커밋 품질 검증 통과"
        return 0
    else
        echo "📋 커밋 품질 검증 결과:"
        printf '%s\n' "${checks[@]}"
        return 1
    fi
}
```

## 📈 커밋 통계 및 분석

### 커밋 히스토리 분석
```json
{
  "commit_statistics": {
    "total_commits": 127,
    "auto_generated": 89,
    "manual_commits": 38,
    "avg_message_length": 78,
    "tag_coverage": "94%",
    "constitution_compliance": "96%"
  },
  "commit_patterns": {
    "most_common_emoji": "✨",
    "avg_files_per_commit": 3.2,
    "peak_commit_hour": 14,
    "spec_completion_rate": "89%"
  }
}
```

## 🎯 Constitution 5원칙 준수

1. **Simplicity**: 복잡한 커밋 작업을 자동화
2. **Architecture**: @TAG 시스템으로 구조화
3. **Testing**: TDD 단계별 커밋 지원
4. **Observability**: 모든 커밋 추적 및 분석
5. **Versioning**: 체계적인 커밋 히스토리 관리

## 💡 사용 시나리오

### 개인 개발 패턴
```bash
# 작업 중 자동 커밋
/git:commit --auto
# → 변경사항 분석 후 적절한 메시지 생성

# 체크포인트 커밋
/git:commit --checkpoint "알고리즘 실험 중"
# → 실험 중간 백업용 커밋
```

### 팀 개발 패턴 (MoAI 워크플로우)
```bash
# SPEC 단계 자동 커밋
/moai:1-spec "사용자 인증"
# → 4단계 자동 커밋 수행

# BUILD 단계 TDD 커밋
/git:commit --build RED
/git:commit --build GREEN
/git:commit --build REFACTOR
# → TDD 3단계 체계적 커밋

# SYNC 단계 완료 커밋
/git:commit --sync
# → 문서 동기화 및 PR 준비 커밋
```

모든 커밋은 git-manager 에이전트와 연동되어 Constitution 5원칙을 자동으로 준수합니다.