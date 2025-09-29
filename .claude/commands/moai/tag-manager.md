---
name: moai:tag-manager
description: TAG 시스템 전문 관리 - 생성, 검증, 체인 무결성, 인덱스 최적화, 온디맨드 TAG 작업
argument-hint: "create DOMAIN | search KEYWORD | validate | repair | index | stats"
tools: Read, Write, Edit, MultiEdit, Glob, Bash
model: sonnet
---

# MoAI-ADK TAG Manager - 16-Core TAG 시스템 전문 관리

**TAG 명령어**: $ARGUMENTS

## 🎯 핵심 기능

MoAI-ADK의 분산 16-Core TAG 시스템을 완전히 관리하는 전용 명령어입니다.

### 지원하는 TAG 작업

1. **TAG 생성**: `create DOMAIN` - 새로운 TAG 체인 생성 및 검증
2. **TAG 검색**: `search KEYWORD` - 기존 TAG 검색 및 재사용 제안
3. **무결성 검증**: `validate` - 전체 TAG 체인 무결성 검사
4. **체인 수리**: `repair` - 끊어진 TAG 체인 자동 복구
5. **인덱스 관리**: `index` - JSONL 인덱스 최적화 및 업데이트
6. **통계 분석**: `stats` - TAG 시스템 성능 및 품질 리포트

## 🚀 실행 방법

### 1. TAG 생성 및 관리

새로운 기능에 대한 완전한 TAG 체인을 생성합니다:

```bash
# 새 기능의 TAG 체인 생성
/moai:tag-manager create AUTH

# 특정 도메인으로 TAG 생성
/moai:tag-manager create PERFORMANCE-OPTIMIZATION

# API 관련 TAG 생성
/moai:tag-manager create USER-API
```

tag-agent가 다음을 자동으로 처리합니다:
- @CATEGORY:DOMAIN-ID 형식 검증
- 기존 유사 TAG 검색 및 재사용 제안
- Primary Chain (@REQ → @DESIGN → @TASK → @TEST) 구성
- JSONL 인덱스 자동 업데이트

### 2. TAG 검색 및 재사용

기존 TAG를 검색하고 재사용 가능성을 평가합니다:

```bash
# 키워드로 TAG 검색
/moai:tag-manager search LOGIN

# 도메인별 TAG 조회
/moai:tag-manager search AUTH

# 파일별 TAG 매핑 검색
/moai:tag-manager search "src/auth/"
```

### 3. TAG 시스템 검증 및 수리

전체 TAG 시스템의 무결성을 검증하고 문제를 수리합니다:

```bash
# 전체 TAG 시스템 검증
/moai:tag-manager validate

# 끊어진 TAG 체인 자동 수리
/moai:tag-manager repair

# 고아 TAG 및 순환 참조 해결
/moai:tag-manager repair --deep
```

### 4. 성능 최적화 및 통계

TAG 시스템 성능을 최적화하고 상태를 모니터링합니다:

```bash
# JSONL 인덱스 최적화
/moai:tag-manager index

# TAG 시스템 성능 통계
/moai:tag-manager stats

# 상세 품질 리포트
/moai:tag-manager stats --detailed
```

## 📋 실행 순서

명령어 파라미터에 따라 tag-agent를 적절히 호출합니다:

### CREATE 작업
```
@agent-tag-agent "$ARGUMENTS 도메인의 새로운 TAG 체인을 생성하고 기존 TAG와의 중복을 검사해주세요"
```

### SEARCH 작업
```
@agent-tag-agent "$ARGUMENTS 키워드로 기존 TAG를 검색하고 재사용 가능한 TAG를 제안해주세요"
```

### VALIDATE 작업
```
@agent-tag-agent "전체 TAG 시스템의 무결성을 검증하고 문제점을 리포트해주세요"
```

### REPAIR 작업
```
@agent-tag-agent "끊어진 TAG 체인과 고아 TAG를 감지하고 자동으로 수리해주세요"
```

### INDEX 작업
```
@agent-tag-agent "JSONL 인덱스를 최적화하고 성능 통계를 업데이트해주세요"
```

### STATS 작업
```
@agent-tag-agent "TAG 시스템의 성능 지표와 품질 상태를 종합 분석해주세요"
```

## 🔧 고급 기능

### 배치 TAG 작업

여러 TAG를 한번에 처리할 수 있습니다:

```bash
# 여러 도메인 동시 생성
/moai:tag-manager create "AUTH,USER,API"

# 다중 키워드 검색
/moai:tag-manager search "LOGIN,AUTH,SECURITY"

# 특정 카테고리만 검증
/moai:tag-manager validate --category=PRIMARY
```

### TAG 마이그레이션

구 형식에서 새 형식으로 TAG를 마이그레이션합니다:

```bash
# 전체 TAG 형식 업그레이드
/moai:tag-manager migrate --format=new

# 특정 파일의 TAG 업그레이드
/moai:tag-manager migrate --file="src/auth.ts"
```

### 성능 튜닝

TAG 시스템 성능을 세밀하게 조정합니다:

```bash
# 인덱스 크기 최적화
/moai:tag-manager optimize --target=size

# 검색 속도 최적화
/moai:tag-manager optimize --target=speed

# 메모리 사용량 최적화
/moai:tag-manager optimize --target=memory
```

## 📊 품질 게이트

tag-agent가 모든 TAG 작업에서 다음 품질을 보장합니다:

- **형식 준수**: @CATEGORY:DOMAIN-ID 형식 100% 검증
- **중복 방지**: 95% 이상 중복 TAG 방지
- **체인 무결성**: Primary Chain 연결 100% 보장
- **인덱스 일관성**: JSONL 실시간 동기화 100%
- **성능 기준**:
  - 검색 속도 < 45ms
  - 인덱스 크기 < 500KB
  - 메모리 사용량 < 50MB

## 🔄 다른 명령어와의 연동

- **`/moai:1-spec`**: SPEC 생성 시 자동 TAG 체인 구성
- **`/moai:2-build`**: TDD 구현 시 Implementation TAG 자동 연결
- **`/moai:3-sync`**: 문서 동기화 시 TAG 참조 업데이트

## 💡 사용 시나리오

### 새 프로젝트 시작
```bash
/moai:tag-manager create PROJECT-INIT
/moai:tag-manager stats --baseline
```

### 기존 기능 확장
```bash
/moai:tag-manager search EXISTING-FEATURE
/moai:tag-manager create FEATURE-EXTENSION
```

### 시스템 점검
```bash
/moai:tag-manager validate
/moai:tag-manager repair
/moai:tag-manager stats --health
```

### 성능 최적화
```bash
/moai:tag-manager index
/moai:tag-manager optimize --target=all
/moai:tag-manager stats --performance
```

이 명령어를 통해 개발자는 TAG 관리에 대해 전혀 신경 쓰지 않고도 완전한 추적성과 품질을 확보할 수 있습니다. tag-agent가 모든 복잡한 TAG 관리 작업을 자동화하여 개발 생산성을 극대화합니다.