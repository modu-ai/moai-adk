---
title: 설정 파일
description: MoAI-ADK 설정 옵션 전체 가이드
---

# 설정 파일

MoAI-ADK는 `.moai/config.json`과 `.claude/settings.json` 두 개의 주요 설정 파일을 사용합니다. 각 설정 파일의 구조와 옵션을 설명합니다.

## 설정 파일 개요

### 설정 파일 위치

```
프로젝트/
├── .moai/
│   └── config.json          # MoAI-ADK 프로젝트 설정
└── .claude/
    └── settings.json        # Claude Code 통합 설정
```

### 설정 우선순위

1. **명령줄 인자** (최우선)
2. **환경 변수**
3. **`.moai/config.json`**
4. **`.claude/settings.json`**
5. **기본값** (최하위)

## .moai/config.json

### 전체 구조

```json
{
  "version": "0.0.1",
  "name": "my-project",
  "description": "A my-project project built with MoAI-ADK",
  "mode": "personal",
  "features": {
    "specFirst": true,
    "tddEnforced": true,
    "tagTracking": true,
    "autoSync": false
  },
  "backup": {
    "enabled": true,
    "path": ".moai/backups",
    "retention": 7,
    "autoBackup": true,
    "beforeCommit": true,
    "compress": true
  },
  "git": {
    "defaultBranch": "main",
    "developBranch": "develop",
    "featurePrefix": "feature/",
    "requireApproval": true,
    "autoCommit": false
  },
  "quality": {
    "testCoverage": 85,
    "trustScore": 82,
    "enforceChecks": true,
    "blockCommitOnFailure": true
  },
  "hooks": {
    "enabled": true,
    "fileMonitor": true,
    "languageDetector": true,
    "policyBlock": true,
    "preWriteGuard": true,
    "runTestsAndReport": {
      "enabled": true,
      "autoTest": false,
      "onSave": false,
      "onCommit": true
    }
  },
  "logging": {
    "level": "info",
    "path": ".moai/logs",
    "maxFiles": 7,
    "maxSize": "10m"
  }
}
```

### 섹션별 상세 설명

#### 1. 기본 정보

```json
{
  "version": "0.0.1",           // MoAI-ADK 버전
  "name": "my-project",         // 프로젝트 이름
  "description": "...",         // 프로젝트 설명
  "mode": "personal"            // "personal" | "team"
}
```

**mode 옵션**:

- **`personal`**: 로컬 개발, `.moai/specs/`에 SPEC 저장
- **`team`**: GitHub Issues 통합, PR 자동화

#### 2. 기능 설정

```json
{
  "features": {
    "specFirst": true,          // SPEC-First 강제
    "tddEnforced": true,        // TDD 사이클 강제
    "tagTracking": true,        // @TAG 시스템 활성화
    "autoSync": false           // 자동 문서 동기화
  }
}
```

**옵션 설명**:

- **`specFirst`**: `true`면 SPEC 없이 `/moai:2-build` 실행 불가
- **`tddEnforced`**: `true`면 테스트 실패 시 커밋 차단
- **`tagTracking`**: `true`면 TAG BLOCK 없는 파일 경고
- **`autoSync`**: `true`면 코드 변경 시 자동으로 `/moai:3-sync` 실행

#### 3. 백업 설정

```json
{
  "backup": {
    "enabled": true,            // 백업 시스템 활성화
    "path": ".moai/backups",    // 백업 저장 경로
    "retention": 7,             // 보관 기간 (일)
    "autoBackup": true,         // 파일 변경 시 자동 백업
    "beforeCommit": true,       // 커밋 전 백업
    "compress": true            // 백업 압축
  }
}
```

**백업 전략**:

```bash
# 백업 생성 시점
- 파일 수정 전 (autoBackup: true)
- Git 커밋 전 (beforeCommit: true)
- 수동: moai restore --create-backup

# 백업 파일 형식
{filename}.{timestamp}.bak

# 예시
service.ts → service.ts.20240115-143045.bak
```

#### 4. Git 설정

```json
{
  "git": {
    "defaultBranch": "main",     // 메인 브랜치
    "developBranch": "develop",  // 개발 브랜치
    "featurePrefix": "feature/", // 기능 브랜치 접두사
    "requireApproval": true,     // 브랜치 생성/머지 승인 필요
    "autoCommit": false          // 자동 커밋 비활성화
  }
}
```

**브랜치 전략**:

```bash
# Feature 브랜치
feature/spec-001-user-auth
feature/task-implement-login

# SPEC 브랜치
feature/spec-{SPEC-ID}-{description}

# 머지 흐름
feature/spec-001 → develop → main
```

#### 5. 품질 게이트

```json
{
  "quality": {
    "testCoverage": 85,          // 최소 테스트 커버리지 (%)
    "trustScore": 82,            // 최소 TRUST 준수율 (%)
    "enforceChecks": true,       // 품질 검사 강제
    "blockCommitOnFailure": true // 실패 시 커밋 차단
  }
}
```

**품질 검증 시점**:

```bash
# Git 커밋 전
- 테스트 커버리지 확인
- TRUST 준수율 계산
- TAG 체인 검증

# 실패 시
✗ Commit blocked: Test coverage 78% < 85%
✗ Commit blocked: TRUST score 75% < 82%

# 강제 커밋 (권장하지 않음)
git commit --no-verify
```

#### 6. 훅 설정

```json
{
  "hooks": {
    "enabled": true,             // 전체 훅 활성화
    "fileMonitor": true,         // 파일 모니터링
    "languageDetector": true,    // 언어 감지
    "policyBlock": true,         // 정책 차단
    "preWriteGuard": true,       // 쓰기 전 검증
    "runTestsAndReport": {
      "enabled": true,
      "autoTest": false,         // 자동 테스트 비활성화
      "onSave": false,
      "onCommit": true           // 커밋 전 테스트
    }
  }
}
```

#### 7. 로깅 설정

```json
{
  "logging": {
    "level": "info",             // "debug" | "info" | "warn" | "error"
    "path": ".moai/logs",        // 로그 저장 경로
    "maxFiles": 7,               // 최대 파일 수
    "maxSize": "10m"             // 최대 파일 크기
  }
}
```

**로그 레벨**:

- **`debug`**: 모든 로그 (개발용)
- **`info`**: 일반 정보 (기본값)
- **`warn`**: 경고 이상
- **`error`**: 오류만

## .claude/settings.json

### 전체 구조

```json
{
  "version": "1.0",
  "project": "my-project",
  "agents": {
    "enabled": true,
    "path": "agents/moai",
    "individual": {
      "spec-builder": true,
      "code-builder": true,
      "doc-syncer": true,
      "cc-manager": true,
      "debug-helper": true,
      "git-manager": true,
      "trust-checker": true
    }
  },
  "commands": {
    "enabled": true,
    "path": "commands/moai",
    "individual": {
      "0-project": true,
      "1-spec": true,
      "2-build": true,
      "3-sync": true,
      "help": true
    }
  },
  "hooks": {
    "enabled": true,
    "path": "hooks/moai",
    "order": [
      "session-notice",
      "language-detector",
      "policy-block",
      "pre-write-guard",
      "file-monitor",
      "steering-guard",
      "run-tests-and-report",
      "claude-code-monitor"
    ]
  },
  "outputStyle": "default",
  "permissions": {
    "fileWrite": "confirm",
    "gitCommit": "confirm",
    "branchCreate": "confirm",
    "branchMerge": "confirm"
  }
}
```

### 섹션별 상세 설명

#### 1. 에이전트 설정

```json
{
  "agents": {
    "enabled": true,              // 전체 에이전트 활성화
    "path": "agents/moai",        // 에이전트 경로
    "individual": {
      "spec-builder": true,       // 개별 에이전트 제어
      "code-builder": true,
      "doc-syncer": true,
      "cc-manager": true,
      "debug-helper": true,
      "git-manager": true,
      "trust-checker": true
    }
  }
}
```

**에이전트 비활성화 예시**:

```json
{
  "individual": {
    "spec-builder": true,
    "code-builder": true,
    "doc-syncer": false,         // 비활성화
    "cc-manager": false,         // 비활성화
    "debug-helper": true,
    "git-manager": true,
    "trust-checker": true
  }
}
```

#### 2. 명령어 설정

```json
{
  "commands": {
    "enabled": true,
    "path": "commands/moai",
    "individual": {
      "0-project": true,          // /moai:0-project
      "1-spec": true,             // /moai:1-spec
      "2-build": true,            // /moai:2-build
      "3-sync": true,             // /moai:3-sync
      "help": true                // /moai:help
    }
  }
}
```

#### 3. 훅 설정

```json
{
  "hooks": {
    "enabled": true,
    "path": "hooks/moai",
    "order": [
      "session-notice",           // 실행 순서 정의
      "language-detector",
      "policy-block",
      "pre-write-guard",
      "file-monitor",
      "steering-guard",
      "run-tests-and-report",
      "claude-code-monitor"
    ]
  }
}
```

**훅 순서 중요성**:

- `session-notice`: 가장 먼저 (사용자에게 안내)
- `policy-block`: 초기 단계 (위험한 작업 차단)
- `pre-write-guard`: 파일 쓰기 전 (검증 및 백업)

#### 4. 출력 스타일

```json
{
  "outputStyle": "default"       // "beginner" | "study" | "pair" | "expert" | "default"
}
```

**스타일 옵션**:

| 스타일 | 대상 | 설명 수준 | 예시 코드 |
|--------|------|-----------|-----------|
| `beginner` | 초보자 | 상세 | 다양한 언어 |
| `study` | 학습자 | 교육적 인사이트 | 주석 포함 |
| `pair` | 페어 프로그래밍 | 협업 중심 | 설명 + 코드 |
| `expert` | 전문가 | 간결 | 코드 중심 |
| `default` | 기본 | 균형 | 적절한 설명 |

#### 5. 권한 설정

```json
{
  "permissions": {
    "fileWrite": "confirm",      // "allow" | "confirm" | "deny"
    "gitCommit": "confirm",
    "branchCreate": "confirm",
    "branchMerge": "confirm"
  }
}
```

**권한 수준**:

- **`allow`**: 자동 실행 (주의!)
- **`confirm`**: 사용자 확인 필요 (권장)
- **`deny`**: 완전 차단

## 환경 변수

### 지원하는 환경 변수

```bash
# MoAI-ADK 설정
export MOAI_MODE=development          # development | production | test
export MOAI_LOG_LEVEL=debug           # debug | info | warn | error

# Claude Code 통합
export CLAUDE_OUTPUT_STYLE=study      # beginner | study | pair | expert | default

# Git 설정
export MOAI_GIT_BRANCH=develop        # 기본 브랜치
export MOAI_REQUIRE_APPROVAL=true     # 승인 필요 여부

# 품질 게이트
export MOAI_MIN_COVERAGE=85           # 최소 커버리지 (%)
export MOAI_MIN_TRUST_SCORE=82        # 최소 TRUST 준수율 (%)
```

### 환경별 설정 예시

#### 개발 환경

```bash
# .env.development
MOAI_MODE=development
MOAI_LOG_LEVEL=debug
MOAI_GIT_BRANCH=develop
MOAI_REQUIRE_APPROVAL=true
```

#### 프로덕션 환경

```bash
# .env.production
MOAI_MODE=production
MOAI_LOG_LEVEL=warning
MOAI_GIT_BRANCH=main
MOAI_REQUIRE_APPROVAL=true
MOAI_MIN_COVERAGE=90
MOAI_MIN_TRUST_SCORE=85
```

#### 테스트 환경

```bash
# .env.test
MOAI_MODE=test
MOAI_LOG_LEVEL=error
MOAI_GIT_BRANCH=test
MOAI_REQUIRE_APPROVAL=false
MOAI_MIN_COVERAGE=100
```

## 설정 우선순위 예시

### 테스트 커버리지 결정

```bash
# 1. 명령줄 인자 (최우선)
moai status --min-coverage 90

# 2. 환경 변수
export MOAI_MIN_COVERAGE=85

# 3. .moai/config.json
{
  "quality": {
    "testCoverage": 82
  }
}

# 4. 기본값
80

# 최종 결정: 90 (명령줄 인자)
```

## 설정 검증

### 설정 파일 검증

```bash
# 설정 파일 검증
moai doctor --check-config

# 출력:
✓ Configuration Validation
  ✓ .moai/config.json: Valid
  ✓ .claude/settings.json: Valid
  ✓ All required fields present
  ✓ No conflicting settings

# 설정 출력
moai status --config

# 출력:
Current Configuration:
  Version: 0.0.1
  Mode: personal
  Test Coverage: 85%
  TRUST Score: 82%
```

### 일반적인 설정 오류

#### 1. mode 불일치

```json
// ❌ Bad: mode와 GitHub 설정 불일치
{
  "mode": "personal",
  "github": {
    "enabled": true         // personal 모드에서는 사용 불가
  }
}

// ✅ Good
{
  "mode": "team",
  "github": {
    "owner": "username",
    "repo": "project"
  }
}
```

#### 2. 권한 설정 과도하게 허용

```json
// ❌ Bad: 위험한 자동 승인
{
  "permissions": {
    "fileWrite": "allow",
    "gitCommit": "allow",
    "branchMerge": "allow"    // 위험!
  }
}

// ✅ Good: 적절한 확인
{
  "permissions": {
    "fileWrite": "confirm",
    "gitCommit": "confirm",
    "branchCreate": "confirm",
    "branchMerge": "confirm"
  }
}
```

## 고급 설정

### Personal → Team 모드 전환

```json
// Step 1: mode 변경
{
  "mode": "team"
}

// Step 2: GitHub 설정 추가
{
  "github": {
    "owner": "your-username",
    "repo": "your-project",
    "enableIssues": true,
    "enablePR": true,
    "defaultLabels": ["moai-adk", "spec-first"]
  }
}

// Step 3: 재초기화
// moai update --mode team
```

### 커스텀 훅 경로

```json
{
  "hooks": {
    "enabled": true,
    "path": "hooks/moai",
    "customPath": ".custom-hooks"  // 추가 훅 경로
  }
}
```

### 언어별 설정

```json
{
  "languages": {
    "typescript": {
      "testRunner": "vitest",
      "linter": "biome",
      "formatter": "biome"
    },
    "python": {
      "testRunner": "pytest",
      "linter": "ruff",
      "formatter": "black"
    }
  }
}
```

## 다음 단계

### CLI 명령어

- **[moai init](/cli/init)**: 프로젝트 초기화
- **[moai doctor](/cli/doctor)**: 시스템 진단
- **[moai status](/cli/status)**: 프로젝트 상태

### 고급 가이드

- **[커스텀 에이전트](/advanced/custom-agents)**: 에이전트 생성
- **[CI/CD 통합](/advanced/ci-cd)**: 자동화 파이프라인

## 참고 자료

- **설정 파일**: `.moai/config.json`, `.claude/settings.json`
- **환경 변수**: `.env`, `.env.development`, `.env.production`
- **스키마**: `config.schema.json` (JSON Schema)