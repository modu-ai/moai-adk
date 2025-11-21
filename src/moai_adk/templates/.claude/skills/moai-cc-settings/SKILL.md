---

name: moai-cc-settings
description: Claude Code settings management, preference customization, and user experience
  optimization. Use when customizing Claude Code behavior, managing user preferences,
  or optimizing development experience.

---

## Quick Reference (30 seconds)

Claude Code 설정 관리는 개발 환경을 커스터마이징하고 사용자 선호도를 최적화하며 개발 생산성을
향상시키는 포괄적인 설정 시스템입니다. 글로벌 설정, 프로젝트별 설정, 사용자 프로필을
유연하게 관리할 수 있으며, 접근성, 성능, 통합 옵션을 세밀하게 조정할 수 있습니다.

**핵심 기능**:
- 글로벌 및 프로젝트별 설정 관리
- 사용자 선호도 커스터마이징
- 다중 프로필 지원
- 자동 설정 검증
- 설정 동기화 및 백업


## Implementation Guide

### What It Does

Claude Code 설정은 다음을 제공합니다:

**설정 관리**:
- JSON 기반 설정 저장소
- 설정 파일 자동 감지
- 버전 관리 및 마이그레이션
- 설정 병합 및 오버라이드
- 기본값 자동 적용

**사용자 선호도**:
- 언어 및 지역 설정
- UI 테마 및 폰트
- 에디터 설정 (들여쓰기, 줄 길이)
- 단축키 커스터마이징
- 키 바인딩

**프로젝트 설정**:
- 프로젝트별 린팅 규칙
- 빌드 및 테스트 설정
- 문서화 옵션
- 배포 구성
- 팀 규칙 및 표준

### When to Use

- ✅ Claude Code 동작 커스터마이징
- ✅ 사용자 선호도 설정 및 관리
- ✅ 개발 경험 최적화
- ✅ 프로젝트 표준화
- ✅ 팀 일관성 보장
- ✅ 설정 문제 해결

### Core Settings Patterns

#### 1. 설정 파일 구조
```json
{
  "user": {
    "name": "GOOS",
    "email": "user@example.com"
  },
  "language": {
    "conversation_language": "ko",
    "agent_prompt_language": "en"
  },
  "ui": {
    "theme": "dark",
    "fontFamily": "Fira Code",
    "fontSize": 12
  },
  "editor": {
    "indentSize": 2,
    "indentType": "spaces",
    "lineLength": 88
  }
}
```

#### 2. 프로젝트별 설정
```json
{
  "project": {
    "name": "MoAI-ADK",
    "owner": "GOOS"
  },
  "constitution": {
    "test_coverage_target": 90,
    "enforce_tdd": true
  },
  "git_strategy": {
    "mode": "personal"
  }
}
```

#### 3. 설정 계층 (Settings Hierarchy)
```
시스템 기본값
  ↓
글로벌 사용자 설정 (~/.claude/settings.json)
  ↓
프로젝트 설정 (.moai/config/config.json)
  ↓
런타임 설정 (환경 변수, 커맨드 라인 옵션)
```

#### 4. 설정 검증 패턴
- 필수 필드 확인
- 타입 검증 (string, number, boolean, object, array)
- 열거값 검증 (정의된 옵션만 허용)
- 범위 검증 (최소/최대값)
- 상호 참조 검증

### Dependencies

- Claude Code settings system
- JSON 설정 파일
- 사용자 선호도 저장소
- 설정 검증 프레임워크
- 파일 감시 시스템


## Works Well With

- `moai-cc-configuration` (고급 설정 관리)
- `moai-cc-hooks` (설정 변경 이벤트)
- `moai-project-config-manager` (프로젝트 설정 CRUD)
- `moai-cc-agents` (에이전트별 설정)


## Advanced Patterns

### 1. 다중 프로필 관리

**프로필 구조**:
```json
{
  "profiles": {
    "development": {
      "theme": "light",
      "debug": true,
      "logLevel": "verbose"
    },
    "production": {
      "theme": "dark",
      "debug": false,
      "logLevel": "error"
    },
    "testing": {
      "theme": "light",
      "debug": true,
      "logLevel": "debug"
    }
  },
  "activeProfile": "development"
}
```

**프로필 전환**:
```bash
# 설정 파일에서 activeProfile 변경
# 또는 런타임 환경 변수로 오버라이드
export CLAUDE_PROFILE=production
```

### 2. 설정 동기화 및 마이그레이션

**설정 버전 관리**:
```json
{
  "version": "2.1.0",
  "migratedFrom": "2.0.0",
  "migrationHistory": [
    {
      "from": "2.0.0",
      "to": "2.1.0",
      "date": "2025-11-21",
      "changes": "Added new accessibility options"
    }
  ]
}
```

**자동 마이그레이션 프로세스**:
1. 버전 확인
2. 마이그레이션 필요 판단
3. 백업 생성
4. 설정 변환 (이전 형식 → 신규 형식)
5. 검증 및 적용

### 3. 설정 충돌 해결

**오버라이드 규칙**:
```
우선순위 (높음 → 낮음):
1. 커맨드 라인 옵션
2. 환경 변수
3. 프로젝트 설정
4. 글로벌 사용자 설정
5. 시스템 기본값
```

**충돌 처리**:
- 상위 우선순위 설정이 하위를 오버라이드
- 중요 설정은 경고 메시지 표시
- 롤백 옵션 제공

### 4. 접근성 설정 패턴

**포함 설정**:
```json
{
  "accessibility": {
    "highContrast": true,
    "fontSize": 16,
    "screenReaderSupport": true,
    "keyboardNavigation": true,
    "colorBlindMode": "deuteranopia"
  }
}
```

### 5. 성능 최적화 설정

**성능 튜닝**:
```json
{
  "performance": {
    "maxContextWindow": 128000,
    "tokenBudget": 200000,
    "cacheEnabled": true,
    "cacheTTL": 3600,
    "parallelTasks": 4
  }
}
```

### 6. 설정 검증 및 모니터링

**검증 규칙**:
```markdown
- 설정 파일 문법 검증 (JSON 유효성)
- 값 범위 검증 (min, max, enum)
- 상호 참조 검증 (의존성 확인)
- 권장사항 제시 (최적 설정)
```


## Changelog

- **v2.0.0** (2025-11-11): Added complete metadata, settings patterns
- **v1.0.0** (2025-10-22): Initial settings management


**End of Skill** | Updated 2025-11-21 | Lines: 185



