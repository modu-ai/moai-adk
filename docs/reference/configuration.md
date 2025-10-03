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
  "_meta": {
    "@CODE:CONFIG-001:DATA": "@DOC:JSON-CONFIG-001",
    "@SPEC:PROJECT-CONFIG-001": "@SPEC:MOAI-CONFIG-001"
  },
  "constitution": {
    "enforce_tdd": true,
    "principles": {
      "simplicity": {
        "max_projects": 5,
        "notes": "기본 권장값. 프로젝트 규모에 따라 .moai/config.json 또는 SPEC/ADR로 근거와 함께 조정하세요."
      }
    },
    "require_tags": true,
    "simplicity_threshold": 5,
    "test_coverage_target": 85
  },
  "git_strategy": {
    "personal": {
      "auto_checkpoint": true,
      "auto_commit": true,
      "branch_prefix": "feature/",
      "checkpoint_interval": 300,
      "cleanup_days": 7,
      "max_checkpoints": 50
    },
    "team": {
      "auto_pr": true,
      "develop_branch": "develop",
      "draft_pr": true,
      "feature_prefix": "feature/SPEC-",
      "main_branch": "main",
      "use_gitflow": true
    }
  },
  "pipeline": {
    "available_commands": [
      "/alfred:1-spec",
      "/alfred:2-build",
      "/alfred:3-sync",
      "/alfred:4-debug"
    ],
    "current_stage": "initialized"
  },
  "project": {
    "created_at": "{{CREATION_TIMESTAMP}}",
    "description": "{{PROJECT_DESCRIPTION}}",
    "initialized": true,
    "mode": "{{PROJECT_MODE}}",
    "name": "{{PROJECT_NAME}}",
    "version": "{{PROJECT_VERSION}}"
  },
  "tags": {
    "auto_sync": true,
    "storage_type": "code_scan",
    "categories": [
      "REQ",
      "DESIGN",
      "TASK",
      "TEST",
      "FEATURE",
      "API",
      "UI",
      "DATA"
    ],
    "code_scan_policy": {
      "no_intermediate_cache": true,
      "realtime_validation": true,
      "scan_tools": ["rg", "grep"],
      "scan_command": "rg '@TAG' -n",
      "philosophy": "TAG의 진실은 코드 자체에만 존재"
    }
  }
}
```

### 섹션별 상세 설명

#### 1. 메타데이터 (_meta)

```json
{
  "_meta": {
    "@CODE:CONFIG-STRUCTURE-001": "@DOC:JSON-CONFIG-001",
    "@SPEC:PROJECT-CONFIG-001": "@SPEC:MOAI-CONFIG-001"
  }
}
```

**설명**: 설정 파일 자체의 @TAG 추적성을 위한 메타데이터

#### 2. 헌법 (constitution)

```json
{
  "constitution": {
    "enforce_tdd": true,                     // TDD 사이클 강제
    "principles": {
      "simplicity": {
        "max_projects": 5,
        "notes": "기본 권장값. 프로젝트 규모에 따라 조정 가능"
      }
    },
    "require_tags": true,                    // @TAG 시스템 필수
    "simplicity_threshold": 5,               // 복잡도 임계값
    "test_coverage_target": 85               // 최소 커버리지 (%)
  }
}
```

**옵션 설명**:

- **`enforce_tdd`**: `true`면 테스트 실패 시 커밋 차단
- **`require_tags`**: `true`면 TAG BLOCK 없는 파일 경고
- **`simplicity_threshold`**: 모듈/함수 복잡도 기준값
- **`test_coverage_target`**: 최소 테스트 커버리지 목표

#### 3. Git 전략 (git_strategy)

##### Personal 모드

```json
{
  "personal": {
    "auto_checkpoint": true,                 // 자동 체크포인트
    "auto_commit": true,                     // 자동 커밋
    "branch_prefix": "feature/",             // 브랜치 접두사
    "checkpoint_interval": 300,              // 체크포인트 간격 (초)
    "cleanup_days": 7,                       // 정리 주기 (일)
    "max_checkpoints": 50                    // 최대 체크포인트 수
  }
}
```

**Personal 모드 특징**:
- 로컬 개발 중심
- `.moai/specs/`에 SPEC 저장
- 자동 체크포인트/커밋
- GitHub 통합 없음

##### Team 모드

```json
{
  "team": {
    "auto_pr": true,                         // 자동 PR 생성
    "develop_branch": "develop",             // 개발 브랜치
    "draft_pr": true,                        // Draft PR로 생성
    "feature_prefix": "feature/SPEC-",       // 기능 브랜치 접두사
    "main_branch": "main",                   // 메인 브랜치
    "use_gitflow": true                      // GitFlow 사용
  }
}
```

**Team 모드 특징**:
- GitHub Issues 통합
- PR 자동화
- GitFlow 브랜치 전략
- Draft PR → Ready for Review 전환

**브랜치 전략**:

```bash
# Personal 모드
feature/task-implement-login
feature/fix-auth-bug

# Team 모드
feature/SPEC-001-user-auth
feature/SPEC-002-api-gateway

# 머지 흐름 (Team)
feature/SPEC-001 → develop → main
```

#### 4. 파이프라인 (pipeline)

```json
{
  "pipeline": {
    "available_commands": [
      "/alfred:1-spec",
      "/alfred:2-build",
      "/alfred:3-sync",
      "/alfred:4-debug"
    ],
    "current_stage": "initialized"
  }
}
```

**Stage 설명**:
- `/alfred:1-spec`: SPEC 작성 (EARS 방식)
- `/alfred:2-build`: TDD 구현 (RED→GREEN→REFACTOR)
- `/alfred:3-sync`: 문서 동기화 (PR 상태 전환)
- `/alfred:4-debug`: 디버깅 및 검증 (온디맨드)

#### 5. 프로젝트 정보 (project)

```json
{
  "project": {
    "created_at": "2025-09-30T12:00:00Z",    // 생성 시간
    "description": "A TypeScript project",    // 프로젝트 설명
    "initialized": true,                      // 초기화 여부
    "mode": "personal",                       // personal | team
    "name": "my-project",                     // 프로젝트 이름
    "version": "0.1.0"                        // 버전
  }
}
```

#### 6. TAG 시스템 (tags)

```json
{
  "tags": {
    "auto_sync": true,                       // 자동 TAG 동기화
    "storage_type": "code_scan",             // 저장소 타입
    "categories": [
      "REQ",                                 // 요구사항
      "DESIGN",                              // 설계
      "TASK",                                // 작업
      "TEST",                                // 테스트
      "FEATURE",                             // 기능 구현
      "API",                                 // API 엔드포인트
      "UI",                                  // UI 컴포넌트
      "DATA"                                 // 데이터 타입
    ],
    "code_scan_policy": {
      "no_intermediate_cache": true,         // 중간 캐시 없음
      "realtime_validation": true,           // 실시간 검증
      "scan_tools": ["rg", "grep"],          // 스캔 도구
      "scan_command": "rg '@TAG' -n",        // 스캔 명령어
      "philosophy": "TAG의 진실은 코드 자체에만 존재"
    }
  }
}
```

**필수 TAG @TAG 체계**:

| 카테고리 | Core | 설명 | 필수 여부 |
|----------|------|------|-----------|
| TAG 흐름 | @SPEC → @TEST → @CODE → @DOC | SPEC → 테스트 → 코드 → 문서 | 필수 |
| @CODE 서브카테고리 | 선택적 | API/UI/DATA/DOMAIN 등 구현 세부사항 | 필수 |

**CODE-FIRST 철학**:
- TAG INDEX 파일 없음 (완전 제거)
- 코드 직접 스캔 방식 (`rg '@TAG' -n`)
- 실시간 검증
- 중간 캐시 없음

## .claude/settings.json

### 전체 구조

```json
{
  "version": "1.0",
  "project": "my-project",
  "agents": {
    "enabled": true,
    "path": "agents/alfred",
    "individual": {
      "spec-builder": true,
      "code-builder": true,
      "doc-syncer": true,
      "cc-manager": true,
      "debug-helper": true,
      "git-manager": true,
      "trust-checker": true,
      "tag-agent": true
    }
  },
  "commands": {
    "enabled": true,
    "path": "commands/alfred",
    "individual": {
      "8-project": true,
      "1-spec": true,
      "2-build": true,
      "3-sync": true,
      "help": true
    }
  },
  "hooks": {
    "enabled": true,
    "path": "hooks/alfred",
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

#### 1. 에이전트 설정 (8개)

```json
{
  "agents": {
    "enabled": true,
    "path": "agents/alfred",
    "individual": {
      "spec-builder": true,       // SPEC 작성 전담
      "code-builder": true,       // TDD 구현 전담
      "doc-syncer": true,         // 문서 동기화 전담
      "cc-manager": true,         // Claude Code 설정 전담
      "debug-helper": true,       // 오류 분석 전담
      "git-manager": true,        // Git 작업 전담
      "trust-checker": true,      // 품질 검증 전담
      "tag-agent": true           // TAG 시스템 독점 관리
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
    "trust-checker": true,
    "tag-agent": true
  }
}
```

#### 2. 명령어 설정

```json
{
  "commands": {
    "enabled": true,
    "path": "commands/alfred",
    "individual": {
      "8-project": true,          // /alfred:8-project
      "1-spec": true,             // /alfred:1-spec
      "2-build": true,            // /alfred:2-build
      "3-sync": true,             // /alfred:3-sync
      "help": true                // /alfred:help
    }
  }
}
```

#### 3. 훅 설정

```json
{
  "hooks": {
    "enabled": true,
    "path": "hooks/alfred",
    "order": [
      "session-notice",           // 세션 안내
      "language-detector",        // 언어 감지
      "policy-block",             // 정책 차단
      "pre-write-guard",          // 쓰기 전 검증
      "file-monitor",             // 파일 모니터링
      "steering-guard",           // 방향 가드
      "run-tests-and-report",     // 테스트 실행 및 보고
      "claude-code-monitor"       // Claude Code 모니터링
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
  "constitution": {
    "test_coverage_target": 82
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
moai doctor

# 출력:
✓ Configuration Validation
  ✓ .moai/config.json: Valid
  ✓ .claude/settings.json: Valid
  ✓ All required fields present
  ✓ No conflicting settings

# 설정 출력
moai status -v

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
  "project": {
    "mode": "personal"
  },
  "git_strategy": {
    "team": {
      "auto_pr": true         // personal 모드에서는 사용 불가
    }
  }
}

// ✅ Good
{
  "project": {
    "mode": "team"
  },
  "git_strategy": {
    "team": {
      "auto_pr": true,
      "develop_branch": "develop"
    }
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

```bash
# 1. moai update 명령어로 전환
moai update --mode team

# 2. .moai/config.json 자동 업데이트
{
  "project": {
    "mode": "team"
  }
}

# 3. GitHub 통합 활성화
# - GitHub Issues 연동
# - PR 자동화
# - Draft PR → Ready for Review
```

### 언어별 도구 자동 감지

MoAI-ADK는 프로젝트 언어를 자동으로 감지하여 최적의 도구를 선택합니다:

| 언어 | 테스트 | 린터 | 포매터 |
|------|--------|------|--------|
| TypeScript | Vitest | Biome | Biome |
| Python | pytest | ruff | black |
| Java | JUnit | checkstyle | google-java-format |
| Go | go test | golint | gofmt |
| Rust | cargo test | clippy | rustfmt |

**언어 감지 우선순위**:
1. `package.json` (TypeScript/JavaScript)
2. `pyproject.toml` / `requirements.txt` (Python)
3. `pom.xml` / `build.gradle` (Java)
4. `go.mod` (Go)
5. `Cargo.toml` (Rust)

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