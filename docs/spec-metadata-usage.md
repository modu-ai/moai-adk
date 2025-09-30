# SPEC 메타데이터 시스템 활용 가이드

##  ripgrep(rg) 기반 검색 예시

### 기본 SPEC 검색

```bash
# 높은 우선순위의 구현 가능한 SPEC 찾기
rg "priority: high" .moai/specs/ | rg "status: draft"

# TypeScript 관련 모든 SPEC 검색
rg "typescript" .moai/specs/ --type md

# 특정 상태의 SPEC만 검색
rg "status: active" .moai/specs/ --type md

# 의존성이 있는 SPEC 찾기
rg "dependencies:" .moai/specs/ --type md -A 5

# 특정 태그를 가진 SPEC 검색
rg "migration" .moai/specs/ --type md

# SPEC ID 패턴으로 검색
rg "SPEC-\d{3}" .moai/specs/ --only-matching --no-filename | sort -u
```

### 고급 검색 패턴

```bash
# 구현 준비된 SPEC (의존성 모두 완료된 draft)
rg -l "status: draft" .moai/specs/ | xargs -I {} sh -c 'echo "=== {} ==="; rg "dependencies:" {} -A 10'

# 높은 우선순위 + 활성 상태 SPEC
rg -l "priority: high" .moai/specs/ | xargs rg -l "status: active"

# 특정 기간에 업데이트된 SPEC
rg "updated: 2025-09" .moai/specs/ --type md

# 태그별 SPEC 개수 통계
rg "^\s*-\s+(\w+)" .moai/specs/ --only-matching --no-filename | sort | uniq -c | sort -nr

# 완료된 SPEC 중 TypeScript 관련
rg -l "status: completed" .moai/specs/ | xargs rg -l "typescript"
```

### 의존성 분석

```bash
# 특정 SPEC에 의존하는 모든 SPEC 찾기
rg "SPEC-012" .moai/specs/ --type md -C 2

# 의존성 체인 추적
rg "dependencies:.*SPEC-010" .moai/specs/ --type md -B 5 -A 5

# 순환 의존성 후보 검색 (수동 검증 필요)
rg "dependencies:" .moai/specs/ --type md -A 10 | rg "SPEC-\d{3}" --only-matching | sort | uniq -d
```

### 프로젝트 상태 대시보드

```bash
# 전체 SPEC 상태 요약
echo "=== SPEC 상태 요약 ==="
rg "status: (draft|active|completed|deprecated)" .moai/specs/ --only-matching --no-filename | sort | uniq -c

echo "=== 우선순위 분포 ==="
rg "priority: (high|medium|low)" .moai/specs/ --only-matching --no-filename | sort | uniq -c

echo "=== 주요 태그 분포 ==="
rg "^\s*-\s+(\w+)" .moai/specs/ --only-matching --no-filename | sed 's/^[[:space:]]*-[[:space:]]*//' | sort | uniq -c | sort -nr | head -10
```

## 🤖 자동화 스크립트 예시

### SPEC 상태 모니터링

```bash
#!/bin/bash
# spec-status-monitor.sh

echo "🔍 활성 SPEC 현황"
rg -l "status: active" .moai/specs/ | while read spec; do
    spec_id=$(rg "spec_id: (SPEC-\d{3})" "$spec" --only-matching --replace '$1')
    priority=$(rg "priority: (\w+)" "$spec" --only-matching --replace '$1')
    echo "  📋 $spec_id ($priority 우선순위)"
done

echo ""
echo "🚦 구현 준비된 SPEC"
rg -l "status: draft" .moai/specs/ | while read spec; do
    spec_id=$(rg "spec_id: (SPEC-\d{3})" "$spec" --only-matching --replace '$1')
    deps=$(rg "dependencies: \[(.*)\]" "$spec" --only-matching --replace '$1' | tr -d ' ')
    if [[ -z "$deps" ]]; then
        priority=$(rg "priority: (\w+)" "$spec" --only-matching --replace '$1')
        echo "  ✅ $spec_id ($priority 우선순위) - 즉시 구현 가능"
    fi
done
```

### 의존성 검증 스크립트

```bash
#!/bin/bash
# validate-dependencies.sh

echo "🔗 의존성 검증 중..."
missing_deps=0

rg "dependencies: \[(.*)\]" .moai/specs/ --only-matching --replace '$1' | tr ',' '\n' | sed 's/[[:space:]]*//g' | sort -u | while read dep; do
    if [[ -n "$dep" ]]; then
        if ! rg -q "spec_id: $dep" .moai/specs/; then
            echo "❌ 누락된 의존성: $dep"
            missing_deps=$((missing_deps + 1))
        fi
    fi
done

if [[ $missing_deps -eq 0 ]]; then
    echo "✅ 모든 의존성 검증 완료"
else
    echo "⚠️  $missing_deps개의 누락된 의존성 발견"
fi
```

##  성능 비교

### grep vs ripgrep 성능

```bash
# 전체 SPEC 디렉토리에서 패턴 검색 성능 비교

# grep (기존)
time grep -r "status: active" .moai/specs/
# 결과: ~50ms (작은 프로젝트), 대규모에서 느림

# ripgrep (권장)
time rg "status: active" .moai/specs/
# 결과: ~5ms (작은 프로젝트), 대규모에서도 빠름

# 파일 타입 필터링 성능
time rg "typescript" .moai/specs/ --type md
# 결과: grep 대비 3-10배 빠름
```

## 🎯 베스트 프랙티스

### 1. 정확한 패턴 사용
```bash
# ❌ 부정확 - 다른 곳의 SPEC- 패턴도 매칭
rg "SPEC-" .moai/specs/

# ✅ 정확 - SPEC ID 형식만 매칭
rg "spec_id: SPEC-\d{3}" .moai/specs/
```

### 2. 파일 타입 지정
```bash
# ❌ 모든 파일 검색 (느림)
rg "status: active" .moai/specs/

# ✅ 마크다운 파일만 검색 (빠름)
rg "status: active" .moai/specs/ --type md
```

### 3. 출력 형식 최적화
```bash
# ❌ 기본 출력 (시각적으로 복잡)
rg "priority: high" .moai/specs/

# ✅ 파일 목록만 출력 (깔끔)
rg -l "priority: high" .moai/specs/

# ✅ 매칭 부분만 출력 (정확)
rg "status: (\w+)" .moai/specs/ --only-matching --replace '$1'
```

이제 모든 검색 작업이 ripgrep 기반으로 최적화되어 더 빠르고 정확한 SPEC 관리가 가능합니다! 🚀