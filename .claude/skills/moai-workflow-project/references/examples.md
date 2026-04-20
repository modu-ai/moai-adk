# MoAI Workflow Project Examples

실용적인 예시를 통해 프로젝트 관리 시스템의 활용 패턴을 학습합니다.

---

## Example 1: 완전한 프로젝트 초기화

**Scenario**: 새로운 웹 애플리케이션 프로젝트를 다국어 지원으로 초기화하는 상황

**Input**:
```python
from moai_menu_project import MoaiMenuProject

# 프로젝트 초기화
project = MoaiMenuProject("./my-ecommerce-app")

# 완전한 프로젝트 설정
result = project.initialize_complete_project(
    language="ko",                          # 한국어 UI
    user_name="김개발",
    domains=["backend", "frontend", "mobile"],
    project_type="web_application",
    optimization_enabled=True
)
```

**Output**:
```json
{
  "initialization_status": "success",
  "project_metadata": {
    "name": "my-ecommerce-app",
    "type": "web_application",
    "initialized_at": "2025-12-06T10:30:00Z"
  },
  "language_config": {
    "conversation_language": "ko",
    "agent_prompt_language": "english",
    "documentation_language": "ko",
    "token_cost_analysis": {
      "base_cost": 1000,
      "localization_overhead": 200,
      "recommendation": "Use English for agent prompts to save 20% tokens"
    }
  },
  "documentation_structure": {
    "created_directories": [
      "docs/ko/",
      "docs/en/",
      ".moai/docs/",
      ".moai/specs/"
    ],
    "generated_files": [
      "README.md",
      "docs/ko/index.md",
      "docs/en/index.md",
      ".moai/docs/architecture.md"
    ]
  },
  "template_analysis": {
    "templates_analyzed": 12,
    "optimization_opportunities": 3,
    "size_reduction_potential": "15%"
  },
  "domains_initialized": ["backend", "frontend", "mobile"]
}
```

**Explanation**: MoaiMenuProject는 DocumentationManager, LanguageInitializer, TemplateOptimizer 세 모듈을 통합합니다. 한국어 UI 설정, 다국어 문서 구조 생성, 템플릿 분석을 한 번에 수행합니다.

---

## Example 2: SPEC 기반 문서 자동 생성

**Scenario**: 사용자 인증 SPEC에서 API 문서와 기능 문서를 자동 생성

**Input**:
```python
# SPEC 데이터 정의
spec_data = {
    "id": "SPEC-001",
    "title": "User Authentication System",
    "description": "Implement secure authentication with JWT",
    "requirements": [
        "User registration with email verification",
        "JWT token generation and validation",
        "Password reset functionality",
        "OAuth2 social login support"
    ],
    "status": "Planned",
    "priority": "High",
    "api_endpoints": [
        {
            "path": "/api/auth/register",
            "method": "POST",
            "description": "User registration endpoint",
            "request_body": {
                "email": "string",
                "password": "string",
                "name": "string"
            },
            "response": {
                "user_id": "string",
                "email": "string",
                "verification_sent": "boolean"
            }
        },
        {
            "path": "/api/auth/login",
            "method": "POST",
            "description": "User login endpoint",
            "request_body": {
                "email": "string",
                "password": "string"
            },
            "response": {
                "access_token": "string",
                "refresh_token": "string",
                "expires_in": "number"
            }
        },
        {
            "path": "/api/auth/refresh",
            "method": "POST",
            "description": "Token refresh endpoint"
        }
    ]
}

# 문서 생성
docs_result = project.generate_documentation_from_spec(spec_data)
```

**Output**:
```markdown
# 생성된 문서: docs/features/SPEC-001-authentication.md

# User Authentication System

## Overview
Implement secure authentication with JWT

**Status**: Planned | **Priority**: High

## Requirements
- [ ] User registration with email verification
- [ ] JWT token generation and validation
- [ ] Password reset functionality
- [ ] OAuth2 social login support

## API Reference

### POST /api/auth/register
User registration endpoint

**Request Body**:
| Field | Type | Required |
|-------|------|----------|
| email | string | Yes |
| password | string | Yes |
| name | string | Yes |

**Response**:
| Field | Type |
|-------|------|
| user_id | string |
| email | string |
| verification_sent | boolean |

### POST /api/auth/login
User login endpoint

**Request Body**:
| Field | Type | Required |
|-------|------|----------|
| email | string | Yes |
| password | string | Yes |

**Response**:
| Field | Type |
|-------|------|
| access_token | string |
| refresh_token | string |
| expires_in | number |

### POST /api/auth/refresh
Token refresh endpoint

---
Generated from: SPEC-001
Last Updated: 2025-12-06
```

**Explanation**: SPEC 데이터의 요구사항, API 엔드포인트, 요청/응답 스키마를 분석하여 구조화된 문서를 자동 생성합니다. 테이블 형식의 API 레퍼런스가 포함됩니다.

---

## Example 3: 템플릿 성능 최적화

**Scenario**: 기존 프로젝트의 템플릿을 분석하고 최적화하는 상황

**Input**:
```python
# 템플릿 분석
analysis = project.template_optimizer.analyze_project_templates()

print(f"분석된 템플릿: {analysis['templates_count']}")
print(f"최적화 기회: {analysis['optimization_opportunities']}")

# 최적화 옵션 설정
optimization_options = {
    "backup_first": True,                    # 백업 생성
    "apply_size_optimizations": True,        # 크기 최적화
    "apply_performance_optimizations": True, # 성능 최적화
    "apply_complexity_optimizations": True,  # 복잡도 감소
    "preserve_functionality": True           # 기능 보존
}

# 최적화 실행
optimization_result = project.optimize_project_templates(optimization_options)
```

**Output**:
```json
{
  "analysis_report": {
    "templates_analyzed": 15,
    "total_size_before": "245KB",
    "complexity_metrics": {
      "high_complexity": 3,
      "medium_complexity": 7,
      "low_complexity": 5
    },
    "bottlenecks_identified": [
      {
        "file": "templates/api-docs.md",
        "issue": "Excessive whitespace",
        "impact": "12KB reduction possible"
      },
      {
        "file": "templates/feature-spec.md",
        "issue": "Duplicate sections",
        "impact": "8KB reduction possible"
      }
    ]
  },
  "optimization_result": {
    "status": "success",
    "backup_created": ".moai/backups/templates-2025-12-06-103000/",
    "files_optimized": 12,
    "size_reduction": {
      "before": "245KB",
      "after": "198KB",
      "saved": "47KB",
      "percentage": "19.2%"
    },
    "performance_improvements": {
      "template_load_time": "-23%",
      "memory_usage": "-15%"
    },
    "complexity_reduction": {
      "high_to_medium": 2,
      "medium_to_low": 4
    }
  },
  "recommendations": [
    "Consider splitting large templates into modules",
    "Use template inheritance for common sections",
    "Enable caching for frequently used templates"
  ]
}
```

**Explanation**: TemplateOptimizer는 템플릿 파일을 분석하여 크기, 복잡도, 성능 병목을 식별합니다. 백업 생성 후 자동 최적화를 수행하며, 19.2% 크기 감소와 23% 로딩 시간 개선을 달성했습니다.

---

## Common Patterns

### Pattern 1: 언어 자동 감지

프로젝트 콘텐츠에서 언어를 자동으로 감지합니다.

```python
# 언어 자동 감지
language = project.language_initializer.detect_project_language()

print(f"감지된 언어: {language}")

# 감지 방법:
# 1. 파일 내용 분석 (주석, 문자열)
# 2. 설정 파일 검사 (package.json locale, etc.)
# 3. 시스템 로케일 확인
# 4. 디렉토리 구조 패턴

# 결과 예시
detection_result = {
    "detected_language": "ko",
    "confidence": 0.85,
    "indicators": [
        {"source": "README.md", "language": "ko", "weight": 0.4},
        {"source": "comments", "language": "ko", "weight": 0.3},
        {"source": "config.yaml", "language": "en", "weight": 0.3}
    ],
    "recommendation": "Use 'ko' for documentation, 'en' for code comments"
}
```

### Pattern 2: 다국어 문서 구조 생성

여러 언어를 지원하는 문서 구조를 생성합니다.

```python
# 다국어 문서 구조 생성
multilingual = project.language_initializer.create_multilingual_documentation_structure("ko")

print(f"생성된 구조: {multilingual}")

# 생성되는 구조:
# docs/
# ├── ko/                    # 한국어 문서 (기본)
# │   ├── index.md
# │   ├── getting-started.md
# │   └── api-reference.md
# ├── en/                    # 영어 문서 (폴백)
# │   ├── index.md
# │   ├── getting-started.md
# │   └── api-reference.md
# └── _meta.json             # 언어 협상 설정

# 언어 협상 설정
language_config = {
    "defaultLocale": "ko",
    "locales": ["ko", "en"],
    "fallback": {
        "ko": ["en"],  # 한국어 없으면 영어로
        "en": []
    }
}
```

### Pattern 3: 에이전트 프롬프트 로컬라이제이션

에이전트 프롬프트를 현지화하면서 토큰 비용을 최적화합니다.

```python
# 프롬프트 로컬라이제이션
localized = project.language_initializer.localize_agent_prompts(
    base_prompt="Generate user authentication system with JWT",
    language="ko"
)

print(f"로컬라이즈된 프롬프트: {localized}")

# 결과
localization_result = {
    "original_prompt": "Generate user authentication system with JWT",
    "localized_prompt": "JWT를 사용한 사용자 인증 시스템을 생성하세요",
    "token_analysis": {
        "original_tokens": 8,
        "localized_tokens": 12,
        "overhead": "+50%"
    },
    "recommendation": {
        "strategy": "hybrid",
        "explanation": "Use English for technical terms, Korean for instructions",
        "optimized_prompt": "JWT 기반 user authentication system 생성"
    }
}

# 비용 최적화 전략
cost_strategies = {
    "full_english": {"cost_impact": 0, "user_experience": "low"},
    "full_localized": {"cost_impact": 20, "user_experience": "high"},
    "hybrid": {"cost_impact": 10, "user_experience": "medium"}
}
```

---

## Anti-Patterns (피해야 할 패턴)

### Anti-Pattern 1: 수동 문서 구조 생성

**Problem**: 각 언어별 문서 폴더를 수동으로 생성하고 관리

```bash
# 잘못된 예시 - 수동 관리
mkdir -p docs/ko docs/en docs/ja
touch docs/ko/index.md docs/en/index.md docs/ja/index.md
# → 언어 추가마다 수동 작업 필요
# → 구조 불일치 위험
```

**Solution**: MoaiMenuProject로 자동 생성

```python
# 올바른 예시 - 자동화
project.language_initializer.create_multilingual_documentation_structure("ko")
# → 일관된 구조 보장
# → 언어 협상 설정 자동 생성
```

### Anti-Pattern 2: 최적화 없는 템플릿 사용

**Problem**: 기본 템플릿을 그대로 사용하여 성능 저하

```python
# 잘못된 예시 - 분석 없이 사용
def generate_docs():
    with open("templates/large-template.md") as f:
        template = f.read()  # 245KB 템플릿 그대로 로드
    return template.format(**data)
```

**Solution**: 최적화 후 사용

```python
# 올바른 예시 - 최적화 적용
def generate_docs():
    # 템플릿 분석 및 최적화
    if not is_optimized("templates/"):
        project.optimize_project_templates({
            "backup_first": True,
            "apply_size_optimizations": True
        })

    # 최적화된 템플릿 사용
    with open("templates/large-template.md") as f:
        template = f.read()  # 198KB로 최적화됨
    return template.format(**data)
```

### Anti-Pattern 3: SPEC 데이터 불완전

**Problem**: 필수 필드 없이 SPEC에서 문서 생성 시도

```python
# 잘못된 예시 - 불완전한 SPEC
spec_data = {
    "title": "Feature X"
    # id, requirements, api_endpoints 누락
}

docs = project.generate_documentation_from_spec(spec_data)
# → 불완전한 문서 생성
# → 빈 섹션 발생
```

**Solution**: 완전한 SPEC 데이터 제공

```python
# 올바른 예시 - 완전한 SPEC
spec_data = {
    "id": "SPEC-001",
    "title": "Feature X",
    "description": "Detailed description",
    "requirements": ["Req 1", "Req 2"],
    "status": "Planned",
    "priority": "High",
    "api_endpoints": [
        {
            "path": "/api/feature",
            "method": "POST",
            "description": "Feature endpoint"
        }
    ]
}

docs = project.generate_documentation_from_spec(spec_data)
# → 완전한 문서 생성
```

### Anti-Pattern 4: 백업 없는 최적화

**Problem**: 백업 없이 템플릿 최적화 실행

```python
# 잘못된 예시 - 백업 없음
project.optimize_project_templates({
    "backup_first": False,  # 위험!
    "apply_size_optimizations": True
})
# → 최적화 실패 시 복구 불가
```

**Solution**: 항상 백업 먼저 생성

```python
# 올바른 예시 - 백업 우선
project.optimize_project_templates({
    "backup_first": True,  # 필수!
    "apply_size_optimizations": True,
    "preserve_functionality": True
})

# 문제 발생 시 복원
if optimization_failed:
    project.restore_from_backup("templates-2025-12-06")
```

---

## CLI Quick Reference

```bash
# 프로젝트 초기화
python -m moai_menu_project.cli init \
    --language ko \
    --domains backend,frontend

# SPEC에서 문서 생성
python -m moai_menu_project.cli generate-docs \
    --spec-file .moai/specs/SPEC-001.json

# 템플릿 최적화
python -m moai_menu_project.cli optimize-templates \
    --backup \
    --dry-run  # 미리보기

# 문서 내보내기
python -m moai_menu_project.cli export-docs \
    --format html \
    --language ko \
    --output ./dist/docs

# 프로젝트 상태 확인
python -m moai_menu_project.cli status
```

---

## Performance Benchmarks

| 작업 | 소요 시간 | 메모리 사용 |
|------|----------|------------|
| 프로젝트 초기화 | 2-3초 | ~50MB |
| 언어 감지 | 500ms | ~10MB |
| SPEC → 문서 생성 | 2-5초 | ~30MB |
| 템플릿 최적화 | 10-30초 | ~100MB |
| 다국어 구조 생성 | 1-2초 | ~20MB |

---

Version: 1.0.0
Last Updated: 2025-12-06
