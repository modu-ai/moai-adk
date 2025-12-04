"""
MoAI Menu Project - Documentation Manager Module

Enhanced documentation management system with template-based generation,
multi-format output, and intelligent content organization.

Integrates patterns from moai-project-documentation skill with advanced
template processing and automation capabilities.
"""

import shutil
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, Optional


class DocumentationManager:
    """Comprehensive documentation management system."""

    def __init__(self, project_root: str, config: Dict[str, Any]):
        self.project_root = Path(project_root)
        self.config = config
        self.docs_dir = self.project_root / "docs"
        self.templates_dir = self.project_root / ".claude/skills/moai-menu-project/templates/doc-templates"
        self._ensure_directories()

    def _ensure_directories(self):
        """Ensure all necessary directories exist."""
        self.docs_dir.mkdir(parents=True, exist_ok=True)
        self.templates_dir.mkdir(parents=True, exist_ok=True)
        (self.docs_dir / "api").mkdir(exist_ok=True)
        (self.docs_dir / "architecture").mkdir(exist_ok=True)
        (self.docs_dir / "guides").mkdir(exist_ok=True)

    def initialize_documentation_structure(self) -> Dict[str, Any]:
        """
        Initialize project documentation structure based on project type.

        Returns:
            Dict with initialization results and created files.
        """
        project_type = self._detect_project_type()
        language = self.config.get("language", {}).get("conversation_language", "en")

        # Initialize core documentation files
        core_docs = {
            "product.md": self._generate_product_doc(project_type, language),
            "structure.md": self._generate_structure_doc(project_type, language),
            "tech.md": self._generate_tech_doc(project_type, language),
        }

        created_files = []

        for doc_name, content in core_docs.items():
            doc_path = self.docs_dir / doc_name
            if not doc_path.exists():
                doc_path.write_text(content, encoding="utf-8")
                created_files.append(str(doc_path))

        # Create API documentation structure
        api_structure = self._create_api_structure(project_type)

        # Create guides structure
        guides_structure = self._create_guides_structure(project_type, language)

        return {
            "project_type": project_type,
            "language": language,
            "created_files": created_files,
            "api_structure": api_structure,
            "guides_structure": guides_structure,
        }

    def _detect_project_type(self) -> str:
        """Detect project type based on project structure and configuration."""

        # Check for explicit project type in config
        project_type = self.config.get("project", {}).get("type")
        if project_type:
            return project_type

        # Analyze project structure
        src_path = self.project_root / "src"
        if not src_path.exists():
            return "unknown"

        # Web Application detection
        if (src_path / "routes" or src_path / "controllers" or src_path / "api" or src_path / "web").exists():
            if (self.project_root / "package.json").exists():
                return "web_application"

        # Mobile Application detection
        if (
            src_path / "android" or src_path / "ios" or src_path / "flutter" or self.project_root / "pubspec.yaml"
        ).exists():
            return "mobile_application"

        # CLI Tool detection
        main_files = list(src_path.glob("main.*")) + list(src_path.glob("cli.*")) + list(src_path.glob("index.*"))
        if main_files and not (src_path / "web" or src_path / "api").exists():
            return "cli_tool"

        # Library/SDK detection
        if (self.project_root / "setup.py").exists() or (self.project_root / "pyproject.toml").exists():
            setup_content = ""
            if (self.project_root / "setup.py").exists():
                setup_content = (self.project_root / "setup.py").read_text()
            if "library" in setup_content.lower() or "sdk" in setup_content.lower():
                return "library_sdk"

        # Data Science/ML detection
        if (src_path / "models" or src_path / "data" or src_path / "ml" or src_path / "pipeline").exists():
            return "data_science_ml"

        return "web_application"  # Default fallback

    def _generate_product_doc(self, project_type: str, language: str) -> str:
        """Generate product.md based on project type and language."""

        templates = {
            "en": {
                "web_application": """# Mission & Strategy

## What problem do we solve?
[Describe the core problem your web application solves]

## Who are our users?
- Primary users: [Describe target user group]
- Secondary users: [Describe secondary user groups]
- User personas: [Link to user persona documents]

## Value proposition
[What unique value does your application provide?]

# Success Metrics

## Key Performance Indicators
- User adoption: [Target within timeframe]
- User retention: [Percentage and timeframe]
- Feature utilization: [Most important features to track]
- Performance metrics: [Response times, uptime targets]

## Measurement frequency
- Daily: [Metrics tracked daily]
- Weekly: [Metrics tracked weekly]
- Monthly: [Metrics tracked monthly]

## Success examples
- "80% user adoption within 2 weeks of launch"
- "95% uptime maintained for 3 consecutive months"
- "User satisfaction score > 4.5/5"

# Next Features (SPEC Backlog)

## High Priority (Next 1-2 sprints)
- [SPEC-XXX] [Feature description]
- [SPEC-XXX] [Feature description]

## Medium Priority (Next 3-4 sprints)
- [SPEC-XXX] [Feature description]
- [SPEC-XXX] [Feature description]

## Future Considerations
- [SPEC-XXX] [Feature description]
- [SPEC-XXX] [Feature description]

---

*Last updated: {timestamp}*
*Version: 1.0.0*
""",
                "mobile_application": """# Mission & Strategy

## What problem do we solve?
[Describe the core problem your mobile app solves]

## Who are our users?
- Primary demographics: [Age, location, tech proficiency]
- Device preferences: [iOS, Android, or both]
- Usage patterns: [Daily, weekly, situational use]

## Value proposition
[What makes your mobile app unique and valuable?]

# Success Metrics

## App Store Performance
- Downloads: [Target number within timeframe]
- Ratings: [Target average rating]
- Reviews: [Target review count and sentiment]
- Store ranking: [Category ranking goals]

## User Engagement
- Daily Active Users (DAU): [Target within timeframe]
- Monthly Active Users (MAU): [Target within timeframe]
- Session duration: [Average session length target]
- Retention rate: [Day 1, 7, 30 retention targets]

## Performance Metrics
- App launch time: [Target launch speed]
- Crash rate: [Target crash percentage]
- Battery usage: [Battery consumption targets]

# Next Features (SPEC Backlog)

## Critical Features (Next release)
- [SPEC-XXX] [Feature description]
- [SPEC-XXX] [Feature description]

## User Experience Enhancements
- [SPEC-XXX] [Feature description]
- [SPEC-XXX] [Feature description]

## Platform-Specific Features
- iOS: [SPEC-XXX] [iOS-specific feature]
- Android: [SPEC-XXX] [Android-specific feature]

---

*Last updated: {timestamp}*
*Version: 1.0.0*
""",
            },
            "ko": {
                "web_application": """# 미션 및 전략

## 어떤 문제를 해결하나요?
[웹 애플리케이션이 해결하는 핵심 문제 설명]

## 사용자는 누구인가요?
- 주요 사용자: [목표 사용자 그룹 설명]
- 보조 사용자: [보조 사용자 그룹 설명]
- 사용자 페르소나: [사용자 페르소나 문서 링크]

## 가치 제안
[애플리케이션이 제공하는 고유한 가치는 무엇인가요?]

# 성공 지표

## 핵심 성과 지표 (KPI)
- 사용자 도입률: [목표 및 기간]
- 사용자 유지율: [백분율 및 기간]
- 기능 활용도: [추적할 가장 중요한 기능들]
- 성능 지표: [응답 시간, 가동 시간 목표]

## 측정 빈도
- 일간: [매일 추적되는 지표]
- 주간: [매주 추적되는 지표]
- 월간: [매월 추적되는 지표]

## 성공 예시
- "출시 2주 내 80% 사용자 도입률"
- "3개월 연속 95% 가동 시간 유지"
- "사용자 만족도 점수 4.5/5 이상"

# 다음 기능들 (SPEC 백로그)

## 높은 우선순위 (다음 1-2 스프린트)
- [SPEC-XXX] [기능 설명]
- [SPEC-XXX] [기능 설명]

## 중간 우선순위 (다음 3-4 스프린트)
- [SPEC-XXX] [기능 설명]
- [SPEC-XXX] [기능 설명]

## 미래 고려사항
- [SPEC-XXX] [기능 설명]
- [SPEC-XXX] [기능 설명]

---

*최종 업데이트: {timestamp}*
*버전: 1.0.0*
"""
            },
        }

        lang_templates = templates.get(language, templates["en"])
        template = lang_templates.get(project_type, lang_templates["web_application"])

        return template.format(
            timestamp=datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
            project_name=self.config.get("project", {}).get("name", "My Project"),
        )

    def _generate_structure_doc(self, project_type: str, language: str) -> str:
        """Generate structure.md based on project type."""

        templates = {
            "en": {
                "web_application": """# System Architecture

## Overall Design Pattern
[Describe the architectural pattern - e.g., MVC, Microservices, Serverless]

## Layers and Interactions
```mermaid
graph TB
    A[Frontend] --> B[API Gateway]
    B --> C[Backend Services]
    C --> D[Database Layer]
    C --> E[External APIs]
    F[CDN] --> A
```

## Core Modules

### Frontend Layer
- **Location**: `src/frontend/`
- **Responsibilities**: User interface, client-side logic
- **Technologies**: [React/Vue/Angular, etc.]

### Backend API Layer
- **Location**: `src/api/`
- **Responsibilities**: REST/GraphQL endpoints, business logic
- **Technologies**: [FastAPI/Express/NestJS, etc.]

### Database Layer
- **Location**: Database instances and migrations
- **Responsibilities**: Data persistence, relationships
- **Technologies**: [PostgreSQL/MongoDB/Redis, etc.]

# External Integrations

## Third-party Services
- **Payment Provider**: [Stripe/PayPal/etc.]
  - Authentication: OAuth 2.0 with API keys
  - Failure modes: Timeout handling, retry logic
  - Fallback: Manual payment processing

- **Email Service**: [SendGrid/SES/etc.]
  - Authentication: API key based
  - Failure modes: Queue for retry
  - Rate limits: [Specific limits]

- **Analytics**: [Google Analytics/Mixpanel/etc.]
  - Data flow: Client-side tracking
  - Privacy: GDPR compliance
  - Batch processing: Daily aggregation

# Traceability

## SPEC to Code Mapping
- SPEC-001 → `src/api/auth/` (Authentication system)
- SPEC-002 → `src/frontend/components/` (User dashboard)
- SPEC-003 → `src/services/payment/` (Payment processing)

## Change Tracking
- All changes reference SPEC IDs in commit messages
- Feature flags controlled via environment variables
- Database migrations numbered and tracked

## TAG System
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation updates
- `refactor:` - Code refactoring
- `perf:` - Performance improvements

---

*Last updated: {timestamp}*
*Version: 1.0.0*
"""
            },
            "ko": {
                "web_application": """# 시스템 아키텍처

## 전체 설계 패턴
[아키텍처 패턴 설명 - 예: MVC, 마이크로서비스, 서버리스]

## 레이어 및 상호작용
```mermaid
graph TB
    A[프론트엔드] --> B[API 게이트웨이]
    B --> C[백엔드 서비스]
    C --> D[데이터베이스 레이어]
    C --> E[외부 API]
    F[CDN] --> A
```

## 핵심 모듈

### 프론트엔드 레이어
- **위치**: `src/frontend/`
- **책임**: 사용자 인터페이스, 클라이언트 측 로직
- **기술**: [React/Vue/Angular 등]

### 백엔드 API 레이어
- **위치**: `src/api/`
- **책임**: REST/GraphQL 엔드포인트, 비즈니스 로직
- **기술**: [FastAPI/Express/NestJS 등]

### 데이터베이스 레이어
- **위치**: 데이터베이스 인스턴스 및 마이그레이션
- **책임**: 데이터 지속성, 관계
- **기술**: [PostgreSQL/MongoDB/Redis 등]

# 외부 통합

## 서드파티 서비스
- **결제 제공업체**: [Stripe/PayPal 등]
  - 인증: API 키와 OAuth 2.0
  - 실패 모드: 타임아웃 처리, 재시도 로직
  - 대체수단: 수동 결제 처리

- **이메일 서비스**: [SendGrid/SES 등]
  - 인증: API 키 기반
  - 실패 모드: 재시를 위한 큐
  - 속도 제한: [구체적인 제한]

- **분석**: [Google Analytics/Mixpanel 등]
  - 데이터 흐름: 클라이언트 측 추적
  - 개인정보보호: GDPR 준수
  - 일괄 처리: 일일 집계

# 추적성

## SPEC-코드 매핑
- SPEC-001 → `src/api/auth/` (인증 시스템)
- SPEC-002 → `src/frontend/components/` (사용자 대시보드)
- SPEC-003 → `src/services/payment/` (결제 처리)

## 변경 추적
- 모든 변경은 커밋 메시지에 SPEC ID 참조
- 기능 플래그는 환경 변수로 제어
- 데이터베이스 마이그레이션 번호화 및 추적

## TAG 시스템
- `feat:` - 새로운 기능
- `fix:` - 버그 수정
- `docs:` - 문서 업데이트
- `refactor:` - 코드 리팩토링
- `perf:` - 성능 개선

---

*최종 업데이트: {timestamp}*
*버전: 1.0.0*
"""
            },
        }

        lang_templates = templates.get(language, templates["en"])
        template = lang_templates.get(project_type, lang_templates["web_application"])

        return template.format(
            timestamp=datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
            project_name=self.config.get("project", {}).get("name", "My Project"),
        )

    def _generate_tech_doc(self, project_type: str, language: str) -> str:
        """Generate tech.md based on project type."""

        templates = {
            "en": {
                "web_application": """# Technology Stack

## Programming Languages
- **Frontend**: TypeScript 5.0+ (React 18+)
- **Backend**: Python 3.11+ (FastAPI)
- **Database**: PostgreSQL 15+
- **Infrastructure**: Docker, Kubernetes

## Framework Choices

### Frontend Framework: React
- **Reason**: Component-based architecture, large ecosystem
- **Version**: 18.2.0+
- **Key Libraries**: React Router, Zustand, React Query

### Backend Framework: FastAPI
- **Reason**: Modern Python, automatic API docs, async support
- **Version**: 0.104.0+
- **Key Features**: Pydantic validation, OpenAPI generation

### Database ORM: SQLAlchemy
- **Reason**: Mature, powerful, database-agnostic
- **Version**: 2.0+
- **Features**: Migration support, relationship management

# Quality Gates

## Required for Merge
- **Test Coverage**: Minimum 85% (pytest --cov)
- **Code Quality**: ruff linting (no warnings)
- **Security**: No high severity vulnerabilities (safety check)
- **Documentation**: All public APIs documented

## Enforcement Tools
```bash
# Pre-commit hooks
ruff check src/
pytest tests/ --cov=src/
safety check
mypy src/
```

## Failure Criteria
- Coverage < 85%: PR blocked
- Security vulnerabilities: PR blocked
- Linting errors: Auto-fix or PR blocked
- Missing documentation: Warning only

# Security Policy

## Secret Management
- **Environment Variables**: All secrets in environment
- **Development**: .env file (gitignored)
- **Production**: Kubernetes secrets or AWS Secrets Manager
- **Rotation**: Quarterly secret rotation required

## Vulnerability Handling
- **Monitoring**: Dependabot + security scanning
- **Response**: Critical patches within 24 hours
- **Testing**: Security tests in CI/CD pipeline

## Incident Response
1. **Detection**: Automated monitoring and alerts
2. **Assessment**: Security team evaluation within 1 hour
3. **Containment**: Isolate affected systems
4. **Resolution**: Patch and deploy within 24 hours
5. **Post-mortem**: Document and improve processes

# Deployment Strategy

## Target Environments
- **Development**: Local Docker containers
- **Staging**: Kubernetes cluster (AWS EKS)
- **Production**: Kubernetes cluster (AWS EKS)

## Release Process
```yaml
# CI/CD Pipeline stages
stages:
  - test           # Run all quality gates
  - build          # Build Docker images
  - security       # Security scanning
  - deploy_staging # Deploy to staging
  - smoke_tests    # Staging validation
  - deploy_prod    # Blue-green deployment
```

## Rollback Procedure
1. **Immediate Rollback**: Kubernetes rollback to previous version
2. **Database Rollback**: Migration rollback scripts
3. **Monitoring**: Verify system health after rollback
4. **Communication**: Notify team and stakeholders

## Environment Profiles
- **Development**: Debug mode, local database
- **Staging**: Production-like, test data
- **Production**: Optimized, monitoring enabled

---

*Last updated: {timestamp}*
*Version: 1.0.0*
"""
            },
            "ko": {
                "web_application": """# 기술 스택

## 프로그래밍 언어
- **프론트엔드**: TypeScript 5.0+ (React 18+)
- **백엔드**: Python 3.11+ (FastAPI)
- **데이터베이스**: PostgreSQL 15+
- **인프라**: Docker, Kubernetes

## 프레임워크 선택

### 프론트엔드 프레임워크: React
- **이유**: 컴포넌트 기반 아키텍처, 큰 생태계
- **버전**: 18.2.0+
- **주요 라이브러리**: React Router, Zustand, React Query

### 백엔드 프레임워크: FastAPI
- **이유**: 현대 Python, 자동 API 문서, 비동기 지원
- **버전**: 0.104.0+
- **주요 기능**: Pydantic 검증, OpenAPI 생성

### 데이터베이스 ORM: SQLAlchemy
- **이유**: 성숙하고 강력하며 데이터베이스 독립적
- **버전**: 2.0+
- **기능**: 마이그레이션 지원, 관계 관리

# 품질 게이트

## 병합 요구사항
- **테스트 커버리지**: 최소 85% (pytest --cov)
- **코드 품질**: ruff 린팅 (경고 없음)
- **보안**: 높은 심각도 취약점 없음 (safety check)
- **문서**: 모든 공개 API 문서화

## 적용 도구
```bash
# Pre-commit hooks
ruff check src/
pytest tests/ --cov=src/
safety check
mypy src/
```

## 실패 기준
- 커버리지 < 85%: PR 차단
- 보안 취약점: PR 차단
- 린팅 오류: 자동 수정 또는 PR 차단
- 문서 누락: 경고만

# 보안 정책

## 시크릿 관리
- **환경 변수**: 모든 시크릿은 환경 변수
- **개발**: .env 파일 (gitignore)
- **프로덕션**: Kubernetes 시크릿 또는 AWS Secrets Manager
- **교체**: 분기별 시크릿 교체 필요

## 취약점 처리
- **모니터링**: Dependabot + 보안 스캐닝
- **응답**: 중요 패치 24시간 내
- **테스트**: CI/CD 파이프라인에서 보안 테스트

## 인시던트 응답
1. **탐지**: 자동 모니터링 및 알림
2. **평가**: 보안 팀 1시간 내 평가
3. **격리**: 영향 받는 시스템 격리
4. **해결**: 24시간 내 패치 및 배포
5. **사후 분석**: 문서화 및 프로세스 개선

# 배포 전략

## 대상 환경
- **개발**: 로컬 Docker 컨테이너
- **스테이징**: Kubernetes 클러스터 (AWS EKS)
- **프로덕션**: Kubernetes 클러스터 (AWS EKS)

## 릴리스 프로세스
```yaml
# CI/CD 파이프라인 단계
stages:
  - test           # 모든 품질 게이트 실행
  - build          # Docker 이미지 빌드
  - security       # 보안 스캐닝
  - deploy_staging # 스테이징 배포
  - smoke_tests    # 스테이징 검증
  - deploy_prod    # 블루-그린 배포
```

## 롤백 절차
1. **즉시 롤백**: Kubernetes 이전 버전 롤백
2. **데이터베이스 롤백**: 마이그레이션 롤백 스크립트
3. **모니터링**: 롤백 후 시스템 상태 확인
4. **소통**: 팀 및 이해관계자 통보

## 환경 프로필
- **개발**: 디버그 모드, 로컬 데이터베이스
- **스테이징**: 프로덕션 유사, 테스트 데이터
- **프로덕션**: 최적화, 모니터링 활성화

---

*최종 업데이트: {timestamp}*
*버전: 1.0.0*
"""
            },
        }

        lang_templates = templates.get(language, templates["en"])
        template = lang_templates.get(project_type, lang_templates["web_application"])

        return template.format(
            timestamp=datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
            project_name=self.config.get("project", {}).get("name", "My Project"),
        )

    def _create_api_structure(self, project_type: str) -> Dict[str, Any]:
        """Create API documentation structure."""

        api_structure = {
            "openapi_spec": self._generate_openapi_spec(project_type),
            "endpoint_docs": {},
            "schema_docs": {},
            "authentication_guide": self._generate_auth_guide(project_type),
        }

        return api_structure

    def _create_guides_structure(self, project_type: str, language: str) -> Dict[str, Any]:
        """Create user guides structure."""

        guides = {
            "getting_started": self._generate_getting_started_guide(project_type, language),
            "user_guide": self._generate_user_guide(project_type, language),
            "developer_guide": self._generate_developer_guide(project_type, language),
            "deployment_guide": self._generate_deployment_guide(project_type, language),
        }

        return guides

    def _generate_openapi_spec(self, project_type: str) -> Dict[str, Any]:
        """Generate OpenAPI specification template."""

        base_spec = {
            "openapi": "3.0.0",
            "info": {
                "title": self.config.get("project", {}).get("name", "API Documentation"),
                "version": "1.0.0",
                "description": f"API documentation for {project_type}",
            },
            "servers": [
                {"url": "http://localhost:8000", "description": "Development server"},
                {"url": "https://api.example.com", "description": "Production server"},
            ],
            "paths": {
                "/health": {
                    "get": {
                        "summary": "Health check endpoint",
                        "responses": {
                            "200": {
                                "description": "Service is healthy",
                                "content": {
                                    "application/json": {
                                        "schema": {
                                            "type": "object",
                                            "properties": {
                                                "status": {"type": "string"},
                                                "timestamp": {"type": "string"},
                                            },
                                        }
                                    }
                                },
                            }
                        },
                    }
                }
            },
            "components": {
                "securitySchemes": {
                    "BearerAuth": {
                        "type": "http",
                        "scheme": "bearer",
                        "bearerFormat": "JWT",
                    }
                }
            },
        }

        return base_spec

    def _generate_auth_guide(self, project_type: str) -> str:
        """Generate authentication guide."""

        return """# Authentication Guide

## Overview
This API uses JWT (JSON Web Tokens) for authentication.

## Getting Started
1. Register for an API key
2. Obtain a JWT token
3. Include token in requests

## Token Usage
```
Authorization: Bearer <your-jwt-token>
```

## Token Expiration
- Tokens expire after 24 hours
- Refresh tokens available for extended sessions
"""

    def _generate_getting_started_guide(self, project_type: str, language: str) -> str:
        """Generate getting started guide."""

        if language == "ko":
            return f"""# 시작 가이드

## {project_type} 프로젝트 설정

### 1. 필수 요구사항
- Python 3.11+
- Node.js 18+
- Git

### 2. 프로젝트 클론
```bash
git clone <repository-url>
cd {self.config.get("project", {}).get("name", "my-project")}
```

### 3. 환경 설정
```bash
# 가상환경 생성
python -m venv venv
source venv/bin/activate  # Windows: venv\\Scripts\\activate

# 의존성 설치
pip install -r requirements.txt
npm install
```

### 4. 설정 파일
```bash
cp .env.example .env
# .env 파일에 필요한 설정값 입력
```

### 5. 데이터베이스 설정
```bash
# 데이터베이스 마이그레이션
alembic upgrade head
```

### 6. 개발 서버 시작
```bash
# 백엔드 서버
uvicorn main:app --reload --port 8000

# 프론트엔드 개발 서버
npm run dev
```

## 다음 단계
- [API 문서](./api/) 살펴보기
- [사용자 가이드](./user-guide.md) 확인
- [개발자 가이드](./developer-guide.md) 참조
"""
        else:
            return f"""# Getting Started

## {project_type} Project Setup

### 1. Prerequisites
- Python 3.11+
- Node.js 18+
- Git

### 2. Clone Project
```bash
git clone <repository-url>
cd {self.config.get("project", {}).get("name", "my-project")}
```

### 3. Environment Setup
```bash
# Create virtual environment
python -m venv venv
source venv/bin/activate  # Windows: venv\\Scripts\\activate

# Install dependencies
pip install -r requirements.txt
npm install
```

### 4. Configuration
```bash
cp .env.example .env
# Add required configuration values to .env
```

### 5. Database Setup
```bash
# Run database migrations
alembic upgrade head
```

### 6. Start Development Server
```bash
# Backend server
uvicorn main:app --reload --port 8000

# Frontend dev server
npm run dev
```

## Next Steps
- Explore [API Documentation](./api/)
- Check [User Guide](./user-guide.md)
- Reference [Developer Guide](./developer-guide.md)
"""

    def _generate_user_guide(self, project_type: str, language: str) -> str:
        """Generate user guide."""

        if language == "ko":
            return """# 사용자 가이드

## 기본 기능
- 주요 기능 설명
- 사용 방법
- 팁과 요령

## 고급 기능
- 전문가 기능
- 설정 옵션
- 자동화

## 문제 해결
- 일반적인 문제
- FAQ
- 지원 연락처
"""
        else:
            return """# User Guide

## Basic Features
- Feature descriptions
- How to use
- Tips and tricks

## Advanced Features
- Expert functionality
- Configuration options
- Automation

## Troubleshooting
- Common issues
- FAQ
- Support contacts
"""

    def _generate_developer_guide(self, project_type: str, language: str) -> str:
        """Generate developer guide."""

        if language == "ko":
            return """# 개발자 가이드

## 개발 환경
- IDE 설정
- 디버깅
- 테스트

## 코드 기여
- 코드 스타일
- PR 프로세스
- 코드 리뷰

## 아키텍처
- 시스템 설계
- 모듈 구조
- 데이터 흐름
"""
        else:
            return """# Developer Guide

## Development Environment
- IDE setup
- Debugging
- Testing

## Contributing Code
- Code style
- PR process
- Code review

## Architecture
- System design
- Module structure
- Data flow
"""

    def _generate_deployment_guide(self, project_type: str, language: str) -> str:
        """Generate deployment guide."""

        if language == "ko":
            return """# 배포 가이드

## 환경 설정
- 개발 환경
- 스테이징 환경
- 프로덕션 환경

## 배포 프로세스
- CI/CD 파이프라인
- 릴리스 절차
- 롤백

## 모니터링
- 로깅
- 메트릭
- 알림
"""
        else:
            return """# Deployment Guide

## Environment Setup
- Development environment
- Staging environment
- Production environment

## Deployment Process
- CI/CD pipeline
- Release procedure
- Rollback

## Monitoring
- Logging
- Metrics
- Alerts
"""

    def generate_documentation_from_spec(self, spec_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Generate documentation based on SPEC data.

        Args:
            spec_data: SPEC specification data

        Returns:
            Dict with generated documentation results
        """

        spec_id = spec_data.get("id", "SPEC-001")
        spec_data.get("title", "Untitled Feature")
        spec_data.get("description", "")

        # Generate feature documentation
        feature_doc = self._generate_feature_documentation(spec_data)

        # Generate API documentation if applicable
        api_doc = self._generate_api_documentation(spec_data)

        # Update product.md with new feature
        self._update_product_documentation(spec_data)

        # Update structure.md if architecture changes
        self._update_structure_documentation(spec_data)

        # Update tech.md if new technologies introduced
        self._update_tech_documentation(spec_data)

        return {
            "spec_id": spec_id,
            "feature_documentation": feature_doc,
            "api_documentation": api_doc,
            "updated_files": [
                "docs/product.md",
                "docs/structure.md",
                "docs/tech.md",
                f"docs/features/{spec_id.lower()}.md",
            ],
        }

    def _generate_feature_documentation(self, spec_data: Dict[str, Any]) -> str:
        """Generate feature documentation from SPEC."""

        spec_id = spec_data.get("id", "SPEC-001")
        title = spec_data.get("title", "Untitled Feature")
        description = spec_data.get("description", "")
        requirements = spec_data.get("requirements", [])

        doc_content = f"""# {title}

## Overview
{description}

## Requirements
"""

        for i, req in enumerate(requirements, 1):
            doc_content += f"{i}. {req}\n"

        doc_content += f"""
## Implementation Details
- **SPEC ID**: {spec_id}
- **Status**: {spec_data.get("status", "Planned")}
- **Priority**: {spec_data.get("priority", "Medium")}

## Usage Examples
[Add usage examples here]

## Testing
- Test cases: [Link to test files]
- Coverage requirements: [Percentage requirements]
- Performance tests: [Performance criteria]

## Related Documentation
- [API Documentation](../api/)
- [Architecture Guide](../structure.md)
- [Technical Stack](../tech.md)

---

*Generated from: {spec_id}*
*Last updated: {datetime.now().strftime("%Y-%m-%d %H:%M:%S")}*
"""

        return doc_content

    def _generate_api_documentation(self, spec_data: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Generate API documentation if SPEC includes API changes."""

        if "api_endpoints" not in spec_data:
            return None

        api_doc = {"endpoints": [], "schemas": [], "examples": []}

        for endpoint in spec_data.get("api_endpoints", []):
            api_doc["endpoints"].append(
                {
                    "path": endpoint.get("path", ""),
                    "method": endpoint.get("method", "GET"),
                    "description": endpoint.get("description", ""),
                    "parameters": endpoint.get("parameters", []),
                    "responses": endpoint.get("responses", {}),
                }
            )

        return api_doc

    def _update_product_documentation(self, spec_data: Dict[str, Any]):
        """Update product.md with new feature information."""

        product_path = self.docs_dir / "product.md"
        if not product_path.exists():
            return

        content = product_path.read_text(encoding="utf-8")

        # Find SPEC Backlog section
        spec_id = spec_data.get("id", "SPEC-001")
        title = spec_data.get("title", "Untitled Feature")

        # Add to next features section
        spec_entry = f"- [{spec_id}] {title}\n"

        if "# Next Features (SPEC Backlog)" in content:
            # Find insertion point after High Priority
            insertion_point = content.find("## High Priority")
            if insertion_point != -1:
                # Find end of High Priority section
                end_point = content.find("##", insertion_point + 1)
                if end_point == -1:
                    end_point = len(content)

                new_content = content[:end_point] + spec_entry + content[end_point:]
                product_path.write_text(new_content, encoding="utf-8")

    def _update_structure_documentation(self, spec_data: Dict[str, Any]):
        """Update structure.md with architecture changes."""

        if "architecture_changes" not in spec_data:
            return

        structure_path = self.docs_dir / "structure.md"
        if not structure_path.exists():
            return

        # Implementation for updating architecture documentation
        # This would be more sophisticated in practice

    def _update_tech_documentation(self, spec_data: Dict[str, Any]):
        """Update tech.md with new technologies."""

        if "technologies" not in spec_data:
            return

        tech_path = self.docs_dir / "tech.md"
        if not tech_path.exists():
            return

        # Implementation for updating technology stack documentation
        # This would be more sophisticated in practice

    def export_documentation(self, format_type: str = "markdown") -> Dict[str, Any]:
        """
        Export documentation in specified format.

        Args:
            format_type: Export format (markdown, html, pdf)

        Returns:
            Dict with export results
        """

        export_results = {
            "format": format_type,
            "files": [],
            "output_directory": "",
            "success": False,
        }

        if format_type == "markdown":
            export_results = self._export_markdown()
        elif format_type == "html":
            export_results = self._export_html()
        elif format_type == "pdf":
            export_results = self._export_pdf()
        else:
            export_results["error"] = f"Unsupported format: {format_type}"

        return export_results

    def _export_markdown(self) -> Dict[str, Any]:
        """Export documentation as markdown bundle."""

        output_dir = self.project_root / "docs-export"
        output_dir.mkdir(exist_ok=True)

        # Copy all markdown files
        markdown_files = list(self.docs_dir.glob("**/*.md"))

        exported_files = []
        for md_file in markdown_files:
            relative_path = md_file.relative_to(self.docs_dir)
            output_path = output_dir / relative_path
            output_path.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(md_file, output_path)
            exported_files.append(str(output_path))

        return {
            "format": "markdown",
            "files": exported_files,
            "output_directory": str(output_dir),
            "success": True,
        }

    def _export_html(self) -> Dict[str, Any]:
        """Export documentation as HTML."""

        # Check for markdown to HTML converter
        try:
            import jinja2  # noqa: F401 - availability check
            import markdown
        except ImportError:
            return {
                "format": "html",
                "error": "Required packages not found: markdown, jinja2",
                "success": False,
            }

        output_dir = self.project_root / "docs-html"
        output_dir.mkdir(exist_ok=True)

        # Generate HTML from markdown files
        exported_files = []
        markdown_files = list(self.docs_dir.glob("**/*.md"))

        for md_file in markdown_files:
            html_content = markdown.markdown(
                md_file.read_text(encoding="utf-8"),
                extensions=["toc", "codehilite", "tables"],
            )

            relative_path = md_file.relative_to(self.docs_dir)
            html_path = output_dir / relative_path.with_suffix(".html")
            html_path.parent.mkdir(parents=True, exist_ok=True)

            html_path.write_text(html_content, encoding="utf-8")
            exported_files.append(str(html_path))

        return {
            "format": "html",
            "files": exported_files,
            "output_directory": str(output_dir),
            "success": True,
        }

    def _export_pdf(self) -> Dict[str, Any]:
        """Export documentation as PDF."""

        # Check for PDF converter
        try:
            import markdown
            import weasyprint
        except ImportError:
            return {
                "format": "pdf",
                "error": "Required packages not found: markdown, weasyprint",
                "success": False,
            }

        output_dir = self.project_root / "docs-pdf"
        output_dir.mkdir(exist_ok=True)

        # Generate PDF from main documentation files
        main_files = ["product.md", "structure.md", "tech.md"]
        exported_files = []

        for filename in main_files:
            md_path = self.docs_dir / filename
            if md_path.exists():
                html_content = markdown.markdown(
                    md_path.read_text(encoding="utf-8"),
                    extensions=["toc", "codehilite", "tables"],
                )

                pdf_path = output_dir / filename.replace(".md", ".pdf")

                # Convert HTML to PDF
                weasyprint.HTML(string=html_content).write_pdf(pdf_path)
                exported_files.append(str(pdf_path))

        return {
            "format": "pdf",
            "files": exported_files,
            "output_directory": str(output_dir),
            "success": True,
        }

    def get_documentation_status(self) -> Dict[str, Any]:
        """Get current documentation status and metrics."""

        status = {
            "total_files": 0,
            "file_types": {},
            "last_updated": None,
            "missing_sections": [],
            "quality_metrics": {},
        }

        if not self.docs_dir.exists():
            return status

        # Count files by type
        for file_path in self.docs_dir.rglob("*"):
            if file_path.is_file():
                status["total_files"] += 1
                suffix = file_path.suffix.lower()
                status["file_types"][suffix] = status["file_types"].get(suffix, 0) + 1

        # Check for required documentation files
        required_files = ["product.md", "structure.md", "tech.md"]
        for req_file in required_files:
            if not (self.docs_dir / req_file).exists():
                status["missing_sections"].append(req_file)

        # Get last update time
        try:
            latest_file = max(self.docs_dir.rglob("*"), key=lambda f: f.stat().st_mtime)
            status["last_updated"] = datetime.fromtimestamp(latest_file.stat().st_mtime).isoformat()
        except ValueError:
            pass

        return status
