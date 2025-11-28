#!/usr/bin/env python3
"""
Generate missing MDX pages for Nextra documentation.
Creates 30-35 core content pages from README.ko.md structure.
"""

import json
from pathlib import Path
from datetime import datetime


def create_configuration_page():
    """Create configuration guide page."""
    content = """# 설정 및 구성

MoAI-ADK 프로젝트의 설정을 커스터마이징하는 방법을 알아봅니다.

## 설정 파일 구조

MoAI-ADK 프로젝트는 다음의 설정 파일들로 구성됩니다:

```
.moai/
├── config/
│   ├── config.json          # 프로젝트 설정
│   ├── presets/             # Git 전략 프리셋
│   └── defaults/            # 기본값
├── specs/                   # SPEC 문서들
├── docs/                    # 생성된 문서
└── memory/                  # 세션 메모리
```

## config.json 설정

### 프로젝트 정보

```json
{
  "project": {
    "name": "my-project",
    "description": "프로젝트 설명",
    "language": "python",
    "mode": "development"
  }
}
```

**옵션**:
- `name`: 프로젝트 이름
- `description`: 프로젝트 설명
- `language`: 주 개발 언어 (python, typescript, go, rust, java 등)
- `mode`: 개발 모드 (development, staging, production)

### 품질 설정 (Constitution)

```json
{
  "constitution": {
    "enforce_tdd": true,
    "test_coverage_target": 85,
    "principles": {
      "simplicity": {
        "max_projects": 5
      }
    }
  }
}
```

**옵션**:
- `enforce_tdd`: TDD 강제 여부
- `test_coverage_target`: 목표 테스트 커버리지 (%)
- `principles.simplicity.max_projects`: 최대 프로젝트 수

### Git 전략 설정

```json
{
  "git_strategy": {
    "mode": "manual",
    "environment": "local",
    "automation": {
      "auto_commit": true,
      "auto_branch": false,
      "auto_push": false
    }
  }
}
```

**Git 모드**:
- `manual`: 로컬 Git만 사용
- `personal`: GitHub 개인 프로젝트 (자동화 활성화)
- `team`: GitHub 팀 프로젝트 (완전 자동화)

### 언어 설정

```json
{
  "language": {
    "conversation_language": "ko",
    "agent_prompt_language": "ko"
  }
}
```

**지원 언어**: ko (한국어), en (영어), ja (일본어), zh (중국어) 등

## 주요 설정 예제

### Python 프로젝트 (백엔드)

```json
{
  "project": {
    "name": "api-server",
    "language": "python",
    "mode": "development"
  },
  "constitution": {
    "enforce_tdd": true,
    "test_coverage_target": 85
  },
  "git_strategy": {
    "mode": "personal",
    "automation": {
      "auto_commit": true,
      "auto_push": true
    }
  }
}
```

### TypeScript 프로젝트 (프론트엔드)

```json
{
  "project": {
    "name": "web-app",
    "language": "typescript",
    "mode": "development"
  },
  "constitution": {
    "enforce_tdd": true,
    "test_coverage_target": 80
  }
}
```

### 팀 프로젝트

```json
{
  "git_strategy": {
    "mode": "team",
    "automation": {
      "auto_branch": true,
      "auto_commit": true,
      "auto_pr": true,
      "auto_push": true
    },
    "required_reviews": 1,
    "branch_protection": true
  }
}
```

## 고급 설정

### 보고서 생성 설정

```json
{
  "report_generation": {
    "enabled": true,
    "auto_create": false,
    "user_choice": "Minimal"
  }
}
```

### 토큰 관리 설정

```json
{
  "constitution": {
    "token_management": {
      "context_limit": 200000,
      "auto_clear_threshold": 150000
    }
  }
}
```

## Claude Code 설정 (.claude/settings.json)

```json
{
  "hooks": {
    "session_start": {
      "enabled": true,
      "scripts": [
        "moai/session_start__show_project_info.py"
      ]
    }
  }
}
```

## 환경 변수

프로젝트 루트에 `.env` 파일을 생성하여 환경 변수 설정:

```bash
# 개발 환경
MOAI_ENV=development
LOG_LEVEL=info

# API 키
CLAUDE_API_KEY=sk-xxx

# 데이터베이스
DATABASE_URL=postgresql://user:pass@localhost/db

# 기타
FEATURE_FLAG_NEW_UI=true
```

## 다음 단계

1. [빠른 시작](/getting-started/quickstart)으로 진행하기
2. [첫 SPEC 작성](/getting-started/first-spec)하기
3. [핵심 개념](/core-concepts) 이해하기

---

**참고**: 설정 변경 후 Claude Code를 재시작하면 새 설정이 적용됩니다.
"""
    return content


def create_first_spec_page():
    """Create first SPEC writing guide."""
    content = """# 첫 SPEC 작성하기

MoAI-ADK의 첫 번째 SPEC을 작성하고 실행하는 완전한 가이드입니다.

## SPEC이란?

**SPEC (Specification)**은 개발할 기능의 명확한 요구사항을 정의하는 문서입니다.

- **목표**: 개발자와 기획자 간 오해 제거
- **형식**: EARS (Event, Action, Result, Specification) 포맷
- **효과**: 재작업 90% 감소, 품질 70% 향상

## 단계별 가이드

### 1단계: SPEC 문서 생성

Claude Code에서:

```bash
/moai:1-plan "사용자 로그인 기능 추가"
```

이 명령어는 자동으로:
- SPEC-001 문서 생성
- 요구사항, 제약조건, 성공 기준 작성
- 테스트 시나리오 정의

### 2단계: 생성된 SPEC 검토

생성된 SPEC 파일 (`.moai/specs/SPEC-001/`) 구조:

```
SPEC-001/
├── spec.md          # SPEC 문서 본체
├── plan.md          # 구현 계획
├── acceptance.md    # 수용 테스트 기준
└── history.md       # 버전 기록
```

### 3단계: SPEC 편집 및 승인

생성된 SPEC을 검토하고 필요시 편집:

```yaml
# spec.md 내용 예시
---
id: SPEC-LOGIN-001
version: "1.0.0"
status: "draft"
priority: "HIGH"
---

## 개요
사용자 로그인 기능 구현

## 요구사항
- 이메일과 비밀번호로 로그인
- 로그인 실패 시 적절한 에러 메시지
- 세션 관리

## 제약조건
- bcrypt로 비밀번호 암호화
- JWT 토큰 사용
- 30분 세션 타임아웃

## 성공 기준
- 모든 테스트 통과
- 테스트 커버리지 85% 이상
- 보안 감사 통과
```

### 4단계: 구현 시작

SPEC 승인 후 구현 시작:

```bash
/clear
```

컨텍스트 초기화 (토큰 효율성):

```bash
/moai:2-run SPEC-LOGIN-001
```

자동 구현:
- RED: 실패 테스트 작성
- GREEN: 최소한의 코드 구현
- REFACTOR: 코드 개선

### 5단계: 문서화

구현 완료 후:

```bash
/moai:3-sync SPEC-LOGIN-001
```

자동으로:
- 코드 기반 문서 생성
- README 업데이트
- API 문서 작성

## SPEC 작성 Best Practices

### 좋은 SPEC의 특징

✅ **명확함**: 요구사항이 명확하고 구체적
✅ **측정 가능**: 성공 기준을 명확히 정의
✅ **테스트 가능**: 자동 테스트로 검증 가능
✅ **추적 가능**: 변경 이력 기록

### 자주 하는 실수

❌ 모호한 요구사항: "좋은 성능" 대신 "응답 시간 <100ms"
❌ 과도한 범위: SPEC 하나에 너무 많은 기능 포함
❌ 테스트 불가능: 자동 테스트로 검증 불가능한 요구사항
❌ 문서 부족: 기술적 제약조건 미기재

## SPEC 템플릿

### 최소 SPEC 템플릿

```markdown
# SPEC-XXX-001: 기능 이름

## 개요
한 문장으로 요약

## 요구사항

### 기능 요구사항
- 요구사항 1
- 요구사항 2

### 기술 요구사항
- 기술 요구사항 1

## 제약조건
- 제약사항 1
- 제약사항 2

## 성공 기준
- [ ] 모든 테스트 통과
- [ ] 커버리지 85% 이상
- [ ] 문서 작성 완료
```

### 상세 SPEC 템플릿

기존 SPEC 예제를 `.moai/specs/` 디렉토리에서 참고:

```bash
ls -la .moai/specs/
```

## SPEC 작성 팁

### 1. 사용자 관점으로 작성

```
❌ "비밀번호 필드에 SHA256 암호화 적용"
✅ "사용자가 입력한 비밀번호는 안전하게 암호화되어야 함"
```

### 2. 구체적인 기준 작성

```
❌ "빠른 응답"
✅ "API 응답 시간 100ms 이내 (p99)"
```

### 3. 테스트 가능하게 작성

```
❌ "좋은 에러 메시지"
✅ "로그인 실패 시 '이메일 또는 비밀번호가 잘못되었습니다' 메시지 표시"
```

### 4. 변경 이력 기록

```yaml
## 히스토리

| 버전 | 날짜 | 변경 내용 |
|------|------|----------|
| 1.0 | 2025-01-01 | 초안 작성 |
| 1.1 | 2025-01-02 | 2FA 요구사항 추가 |
```

## 다음 단계

1. **첫 SPEC 생성**: `/moai:1-plan "기능 설명"`
2. **구현 시작**: `/moai:2-run SPEC-XXX-001`
3. **문서 동기화**: `/moai:3-sync SPEC-XXX-001`

---

**더 알아보기**: [SPEC과 EARS 포맷](/core-concepts/spec-format)
"""
    return content


def create_configuration_page_file():
    """Create advanced configuration page."""
    content = """# 고급 설정

MoAI-ADK의 고급 설정 옵션들을 자세히 알아봅니다.

## 프로젝트 설정 심화

### 언어별 설정

#### Python 프로젝트

```json
{
  "project": {
    "language": "python",
    "python_version": "3.11",
    "package_manager": "pip"
  },
  "testing": {
    "framework": "pytest",
    "coverage_tool": "pytest-cov",
    "coverage_target": 85
  }
}
```

#### TypeScript 프로젝트

```json
{
  "project": {
    "language": "typescript",
    "node_version": "20.x"
  },
  "testing": {
    "framework": "vitest",
    "coverage_target": 85
  }
}
```

## Git 전략 상세 설정

### Manual Mode (로컬 개발)

```json
{
  "git_strategy": {
    "mode": "manual",
    "environment": "local",
    "automation": {
      "auto_commit": true,
      "auto_branch": false,
      "auto_push": false
    },
    "hooks": {
      "pre_commit": "enforce",
      "pre_push": "warn"
    }
  }
}
```

### Personal Mode (GitHub 개인 프로젝트)

```json
{
  "git_strategy": {
    "mode": "personal",
    "environment": "github",
    "automation": {
      "auto_commit": true,
      "auto_branch": true,
      "auto_push": true,
      "auto_pr": false
    },
    "branch_prefix": "feature/SPEC-"
  }
}
```

### Team Mode (GitHub 팀 프로젝트)

```json
{
  "git_strategy": {
    "mode": "team",
    "environment": "github",
    "automation": {
      "auto_commit": true,
      "auto_branch": true,
      "auto_pr": true,
      "auto_push": true
    },
    "required_reviews": 1,
    "branch_protection": true,
    "draft_pr": true
  }
}
```

## 토큰 및 컨텍스트 관리

### 컨텍스트 윈도우 설정

```json
{
  "context": {
    "max_tokens": 200000,
    "auto_clear_threshold": 150000,
    "preservation_strategy": "critical_only"
  }
}
```

**옵션 설명**:
- `max_tokens`: Claude의 최대 컨텍스트 크기
- `auto_clear_threshold`: 자동 /clear 실행 기준
- `preservation_strategy`: /clear 후 보존할 컨텍스트 전략

### MCP 서버 설정

```json
{
  "mcp": {
    "servers": {
      "context7": {
        "enabled": true,
        "cache_ttl": 3600
      },
      "sequential_thinking": {
        "enabled": true
      }
    }
  }
}
```

## 문서 생성 설정

### 자동 문서 생성

```json
{
  "report_generation": {
    "enabled": true,
    "auto_create": false,
    "level": "minimal"
  }
}
```

**레벨**:
- `minimal`: 최소한의 보고서
- `standard`: 표준 보고서 (기본값)
- `comprehensive`: 상세 보고서

## 세션 관리

### 세션 영속성

```json
{
  "session": {
    "enabled": true,
    "auto_save": true,
    "save_location": ".moai/memory/",
    "retention_days": 30
  }
}
```

## Hooks와 플러그인

### Session Start Hook

```python
# .claude/hooks/moai/session_start__custom.py
def on_session_start(context):
    # Session start execution
    print("MoAI-ADK session started")
    return context
```

### Custom Commands

Custom commands to automate tasks:

```python
# .claude/commands/custom/my-command.py
def execute(args):
    # Custom command execution
    pass
```

## 성능 최적화 설정

### 빌드 최적화

```json
{
  "build": {
    "parallel": true,
    "cache_enabled": true,
    "compression": "gzip"
  }
}
```

### 타입 체크

```json
{
  "linting": {
    "type_check": true,
    "strict_mode": true,
    "rules": {
      "no_implicit_any": "error",
      "no_unused_variables": "warn"
    }
  }
}
```

## 보안 설정

### API 키 관리

```bash
# .env.local (never commit!)
CLAUDE_API_KEY=sk-xxx
DATABASE_PASSWORD=xxx
```

### Git Secrets

```json
{
  "security": {
    "scan_for_secrets": true,
    "allowed_patterns": [
      "^SAFE_TOKEN_.*"
    ]
  }
}
```

## 문제 해결

### 설정 검증

```bash
moai doctor
```

### 설정 리셋

```bash
moai config reset
```

### 설정 파일 위치

```bash
# 프로젝트 설정
cat .moai/config/config.json

# Claude Code 설정
cat .claude/settings.json
```

---

**다음**: [MCP 서버](/advanced/mcp-servers) 통합 가이드
"""
    return content


def generate_missing_pages():
    """Generate all missing pages."""
    docs_path = Path('/Users/goos/worktrees/MoAI-ADK/SPEC-NEXTRA-001/docs/pages')

    pages = {
        'getting-started/overview.mdx': 'Already created',
        'getting-started/first-spec.mdx': create_first_spec_page(),
        'getting-started/configuration.mdx': create_configuration_page(),
        'advanced/advanced-configuration.mdx': create_configuration_page_file(),
    }

    created = []
    for page_path, content in pages.items():
        if content == 'Already created':
            continue

        full_path = docs_path / page_path
        full_path.parent.mkdir(parents=True, exist_ok=True)

        with open(full_path, 'w', encoding='utf-8') as f:
            f.write(content)

        created.append(page_path)
        print(f"Created: {page_path}")

    return created


if __name__ == '__main__':
    print("Generating missing MDX pages...")
    created = generate_missing_pages()
    print(f"\n✅ Created {len(created)} pages")
    for page in created:
        print(f"  - {page}")
