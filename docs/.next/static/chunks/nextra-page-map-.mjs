import ko_meta from "../../../pages/ko/_meta.tsx";
import ko_agents_meta from "../../../pages/ko/agents/_meta.tsx";
import ko_alfred_meta from "../../../pages/ko/alfred/_meta.tsx";
import ko_case_studies_meta from "../../../pages/ko/case-studies/_meta.tsx";
import ko_examples_meta from "../../../pages/ko/examples/_meta.tsx";
import ko_examples_authentication_meta from "../../../pages/ko/examples/authentication/_meta.tsx";
import ko_examples_database_meta from "../../../pages/ko/examples/database/_meta.tsx";
import ko_examples_performance_meta from "../../../pages/ko/examples/performance/_meta.tsx";
import ko_examples_rest_api_meta from "../../../pages/ko/examples/rest-api/_meta.tsx";
import ko_examples_security_meta from "../../../pages/ko/examples/security/_meta.tsx";
import ko_examples_testing_meta from "../../../pages/ko/examples/testing/_meta.tsx";
import ko_skills_meta from "../../../pages/ko/skills/_meta.tsx";
import ko_tutorials_meta from "../../../pages/ko/tutorials/_meta.tsx";
export const pageMap = [{
  name: "features",
  route: "/features",
  children: [{
    name: "index",
    route: "/features",
    frontMatter: {
      "title": "Features Overview",
      "description": "MoAI-ADK의 핵심 기능들과 혁신적인 개발 경험"
    }
  }, {
    name: "statusline",
    route: "/features/statusline",
    frontMatter: {
      "title": "Claude Code Statusline 통합",
      "description": "Claude Code 터미널 상태 표시줄에 실시간 개발 상황 표시"
    }
  }]
}, {
  name: "index",
  route: "/",
  frontMatter: {
    "title": "MoAI-ADK Documentation",
    "description": "SPEC-First TDD Framework Complete Documentation System"
  }
}, {
  name: "ko",
  route: "/ko",
  children: [{
    data: ko_meta
  }, {
    name: "agents",
    route: "/ko/agents",
    children: [{
      data: ko_agents_meta
    }, {
      name: "19-team-members",
      route: "/ko/agents/19-team-members",
      frontMatter: {
        "title": "Alfred의 18명 팀원 에이전트",
        "description": "MoAI-ADK 전문가 팀 완전 가이드 - 역할, 책임, 협업 패턴"
      }
    }, {
      name: "index",
      route: "/ko/agents",
      frontMatter: {
        "title": "에이전트 목록",
        "description": "MoAI-ADK의 29개 전문 에이전트와 역할 소개"
      }
    }]
  }, {
    name: "alfred",
    route: "/ko/alfred",
    children: [{
      data: ko_alfred_meta
    }, {
      name: "agents",
      route: "/ko/alfred/agents",
      frontMatter: {
        "title": "Alfred Agents",
        "description": "Alfred 슈퍼에이전트의 32개 전문 에이전트 상세 가이드"
      }
    }, {
      name: "commands",
      route: "/ko/alfred/commands",
      frontMatter: {
        "title": "Alfred Commands",
        "description": "Alfred 슈퍼에이전트의 핵심 명령어 가이드"
      }
    }, {
      name: "expert-delegation-system",
      route: "/ko/alfred/expert-delegation-system",
      frontMatter: {
        "title": "Alfred Expert Delegation System v2.0",
        "description": "v0.23.0 전문가 위임 시스템 - 4단계 자동 전문가 할당 아키텍처"
      }
    }, {
      name: "workflow",
      route: "/ko/alfred/workflow",
      frontMatter: {
        "title": "Alfred Workflow",
        "description": "Alfred의 4단계 개발 워크플로우 상세 가이드"
      }
    }]
  }, {
    name: "architecture",
    route: "/ko/architecture",
    children: [{
      name: "baas",
      route: "/ko/architecture/baas",
      frontMatter: {
        "title": "BaaS 통합",
        "description": "MoAI-ADK의 Backend as a Service 통합 아키텍처"
      }
    }, {
      name: "index",
      route: "/ko/architecture",
      frontMatter: {
        "sidebarTitle": "Index"
      }
    }, {
      name: "patterns",
      route: "/ko/architecture/patterns",
      frontMatter: {
        "title": "아키텍처 패턴",
        "description": "MoAI-ADK의 핵심 아키텍처 패턴과 설계 원칙"
      }
    }]
  }, {
    name: "case-studies",
    route: "/ko/case-studies",
    children: [{
      data: ko_case_studies_meta
    }, {
      name: "ecommerce-platform",
      route: "/ko/case-studies/ecommerce-platform",
      frontMatter: {
        "title": "사례 연구: E-commerce 플랫폼 개발",
        "description": "6주 만에 MVP 완성, 87.5% 테스트 커버리지, 제로 프로덕션 버그"
      }
    }, {
      name: "enterprise-saas-security",
      route: "/ko/case-studies/enterprise-saas-security",
      frontMatter: {
        "title": "사례 연구: Enterprise SaaS 보안 구현",
        "description": "SOC 2 Type 2 준수, Multi-tenant 아키텍처, Zero-trust 보안 모델"
      }
    }, {
      name: "index",
      route: "/ko/case-studies",
      frontMatter: {
        "title": "실전 사례 연구",
        "description": "MoAI-ADK를 활용한 실제 프로젝트 성공 사례"
      }
    }, {
      name: "microservices-migration",
      route: "/ko/case-studies/microservices-migration",
      frontMatter: {
        "title": "사례 연구: Microservices 아키텍처 전환",
        "description": "Zero-downtime 마이그레이션, 95% 성능 향상, 배포 주기 10배 개선"
      }
    }]
  }, {
    name: "contributing",
    route: "/ko/contributing",
    children: [{
      name: "index",
      route: "/ko/contributing",
      frontMatter: {
        "sidebarTitle": "Index"
      }
    }, {
      name: "index",
      route: "/ko/contributing",
      frontMatter: {
        "sidebarTitle": "Index"
      }
    }]
  }, {
    name: "examples",
    route: "/ko/examples",
    children: [{
      data: ko_examples_meta
    }, {
      name: "authentication",
      route: "/ko/examples/authentication",
      children: [{
        data: ko_examples_authentication_meta
      }, {
        name: "index",
        route: "/ko/examples/authentication",
        frontMatter: {
          "title": "인증 예제",
          "description": "보안 인증 시스템 구축 예제"
        }
      }, {
        name: "jwt-basic",
        route: "/ko/examples/authentication/jwt-basic",
        frontMatter: {
          "title": "JWT 기본 인증",
          "category": "authentication",
          "difficulty": "초급",
          "tags": ["jwt", "auth", "fastapi", "security", "python"]
        }
      }, {
        name: "oauth2",
        route: "/ko/examples/authentication/oauth2",
        frontMatter: {
          "title": "OAuth2 인증",
          "category": "authentication",
          "difficulty": "중급",
          "tags": ["oauth2", "social-login", "google", "github"]
        }
      }, {
        name: "rbac",
        route: "/ko/examples/authentication/rbac",
        frontMatter: {
          "title": "역할 기반 접근 제어 (RBAC)",
          "category": "authentication",
          "difficulty": "고급",
          "tags": ["rbac", "permissions", "authorization", "security"]
        }
      }, {
        name: "refresh-tokens",
        route: "/ko/examples/authentication/refresh-tokens",
        frontMatter: {
          "title": "리프레시 토큰",
          "category": "authentication",
          "difficulty": "중급",
          "tags": ["jwt", "refresh-token", "security", "authentication"]
        }
      }]
    }, {
      name: "database",
      route: "/ko/examples/database",
      children: [{
        data: ko_examples_database_meta
      }, {
        name: "index",
        route: "/ko/examples/database",
        frontMatter: {
          "title": "데이터베이스 예제",
          "description": "SQLAlchemy와 Alembic을 사용한 데이터베이스 관리"
        }
      }, {
        name: "migrations",
        route: "/ko/examples/database/migrations",
        frontMatter: {
          "title": "Alembic 마이그레이션",
          "category": "database",
          "difficulty": "초급",
          "tags": ["alembic", "migration", "database", "sqlalchemy"]
        }
      }, {
        name: "query-optimization",
        route: "/ko/examples/database/query-optimization",
        frontMatter: {
          "title": "쿼리 최적화",
          "category": "database",
          "difficulty": "고급",
          "tags": ["performance", "optimization", "sqlalchemy", "indexing"]
        }
      }, {
        name: "relationships",
        route: "/ko/examples/database/relationships",
        frontMatter: {
          "title": "SQLAlchemy 관계",
          "category": "database",
          "difficulty": "중급",
          "tags": ["sqlalchemy", "relationships", "orm", "foreign-key"]
        }
      }, {
        name: "transactions",
        route: "/ko/examples/database/transactions",
        frontMatter: {
          "title": "트랜잭션 처리",
          "category": "database",
          "difficulty": "중급",
          "tags": ["transaction", "acid", "rollback", "sqlalchemy"]
        }
      }]
    }, {
      name: "index",
      route: "/ko/examples",
      frontMatter: {
        "title": "코드 예제",
        "description": "MoAI-ADK를 활용한 실전 코드 예제 모음"
      }
    }, {
      name: "performance",
      route: "/ko/examples/performance",
      children: [{
        data: ko_examples_performance_meta
      }, {
        name: "batch-processing",
        route: "/ko/examples/performance/batch-processing",
        frontMatter: {
          "title": "배치 처리",
          "category": "performance",
          "difficulty": "고급",
          "tags": ["batch", "async", "celery", "performance"]
        }
      }, {
        name: "caching",
        route: "/ko/examples/performance/caching",
        frontMatter: {
          "title": "Redis 캐싱",
          "category": "performance",
          "difficulty": "중급",
          "tags": ["redis", "caching", "performance", "optimization"]
        }
      }, {
        name: "connection-pooling",
        route: "/ko/examples/performance/connection-pooling",
        frontMatter: {
          "title": "커넥션 풀링",
          "category": "performance",
          "difficulty": "초급",
          "tags": ["database", "pooling", "optimization", "sqlalchemy"]
        }
      }, {
        name: "index",
        route: "/ko/examples/performance",
        frontMatter: {
          "title": "성능 예제",
          "description": "시스템 성능 최적화 기법"
        }
      }, {
        name: "lazy-loading",
        route: "/ko/examples/performance/lazy-loading",
        frontMatter: {
          "title": "지연 로딩",
          "category": "performance",
          "difficulty": "중급",
          "tags": ["sqlalchemy", "lazy-loading", "performance", "n+1"]
        }
      }]
    }, {
      name: "rest-api",
      route: "/ko/examples/rest-api",
      children: [{
        data: ko_examples_rest_api_meta
      }, {
        name: "basic-crud",
        route: "/ko/examples/rest-api/basic-crud",
        frontMatter: {
          "title": "기본 CRUD 작업",
          "category": "rest-api",
          "difficulty": "초급",
          "tags": ["fastapi", "crud", "sqlalchemy", "rest-api", "python"]
        }
      }, {
        name: "error-handling",
        route: "/ko/examples/rest-api/error-handling",
        frontMatter: {
          "title": "에러 처리 & 검증",
          "category": "rest-api",
          "difficulty": "초급",
          "tags": ["fastapi", "error-handling", "validation", "pydantic", "exception"]
        }
      }, {
        name: "filtering",
        route: "/ko/examples/rest-api/filtering",
        frontMatter: {
          "title": "필터링 & 검색",
          "category": "rest-api",
          "difficulty": "중급",
          "tags": ["fastapi", "filtering", "search", "sqlalchemy", "query"]
        }
      }, {
        name: "index",
        route: "/ko/examples/rest-api",
        frontMatter: {
          "title": "REST API 예제",
          "description": "FastAPI를 활용한 RESTful API 구현 예제"
        }
      }, {
        name: "pagination",
        route: "/ko/examples/rest-api/pagination",
        frontMatter: {
          "title": "페이지네이션 & 정렬",
          "category": "rest-api",
          "difficulty": "중급",
          "tags": ["fastapi", "pagination", "sorting", "performance", "sqlalchemy"]
        }
      }]
    }, {
      name: "security",
      route: "/ko/examples/security",
      children: [{
        data: ko_examples_security_meta
      }, {
        name: "index",
        route: "/ko/examples/security",
        frontMatter: {
          "title": "보안 예제",
          "description": "애플리케이션 보안 강화 기법"
        }
      }, {
        name: "input-validation",
        route: "/ko/examples/security/input-validation",
        frontMatter: {
          "title": "입력 검증",
          "category": "security",
          "difficulty": "초급",
          "tags": ["pydantic", "validation", "security", "fastapi"]
        }
      }, {
        name: "rate-limiting",
        route: "/ko/examples/security/rate-limiting",
        frontMatter: {
          "title": "속도 제한 (Rate Limiting)",
          "category": "security",
          "difficulty": "중급",
          "tags": ["rate-limit", "security", "redis", "fastapi"]
        }
      }, {
        name: "sql-injection-prevention",
        route: "/ko/examples/security/sql-injection-prevention",
        frontMatter: {
          "title": "SQL 인젝션 방지",
          "category": "security",
          "difficulty": "초급",
          "tags": ["sql", "security", "sqlalchemy", "injection"]
        }
      }, {
        name: "xss-protection",
        route: "/ko/examples/security/xss-protection",
        frontMatter: {
          "title": "XSS 방어",
          "category": "security",
          "difficulty": "중급",
          "tags": ["xss", "security", "sanitization", "html"]
        }
      }]
    }, {
      name: "testing",
      route: "/ko/examples/testing",
      children: [{
        data: ko_examples_testing_meta
      }, {
        name: "fixtures",
        route: "/ko/examples/testing/fixtures",
        frontMatter: {
          "title": "테스트 픽스처",
          "category": "testing",
          "difficulty": "초급",
          "tags": ["pytest", "fixtures", "setup", "testing"]
        }
      }, {
        name: "index",
        route: "/ko/examples/testing",
        frontMatter: {
          "title": "테스팅 예제",
          "description": "Pytest를 활용한 체계적인 테스트 작성"
        }
      }, {
        name: "integration-tests",
        route: "/ko/examples/testing/integration-tests",
        frontMatter: {
          "title": "통합 테스트",
          "category": "testing",
          "difficulty": "중급",
          "tags": ["pytest", "integration-test", "api", "fastapi"]
        }
      }, {
        name: "mocking",
        route: "/ko/examples/testing/mocking",
        frontMatter: {
          "title": "외부 API 모킹",
          "category": "testing",
          "difficulty": "중급",
          "tags": ["pytest", "mock", "unittest", "testing"]
        }
      }, {
        name: "unit-tests",
        route: "/ko/examples/testing/unit-tests",
        frontMatter: {
          "title": "Pytest 단위 테스트",
          "category": "testing",
          "difficulty": "초급",
          "tags": ["pytest", "unit-test", "tdd", "testing", "python"]
        }
      }]
    }]
  }, {
    name: "features",
    route: "/ko/features",
    children: [{
      name: "index",
      route: "/ko/features",
      frontMatter: {
        "sidebarTitle": "Index"
      }
    }, {
      name: "overview",
      route: "/ko/features/overview",
      frontMatter: {
        "title": "기능 개요",
        "description": "MoAI-ADK의 핵심 기능들과 SPEC-First TDD 개발 방법론"
      }
    }, {
      name: "spec-first",
      route: "/ko/features/spec-first",
      frontMatter: {
        "title": "SPEC-First 개발",
        "description": "요구사항 우선 개발 방식과 EARS 명세서 자동화"
      }
    }, {
      name: "tdd",
      route: "/ko/features/tdd",
      frontMatter: {
        "title": "테스트 주도 개발 (TDD)",
        "description": "MoAI-ADK의 자동화된 TDD 워크플로우와 87.84%+ 테스트 커버리지 달성"
      }
    }]
  }, {
    name: "getting-started",
    route: "/ko/getting-started",
    children: [{
      name: "first-project",
      route: "/ko/getting-started/first-project",
      frontMatter: {
        "title": "첫 프로젝트 상세 가이드",
        "description": "MoAI-ADK로 첫 프로젝트를 완전히 설정하는 상세 가이드 - SPEC 작성부터 배포까지 전체 과정"
      }
    }, {
      name: "index",
      route: "/ko/getting-started",
      frontMatter: {
        "title": "시작하기",
        "description": "MoAI-ADK 빠른 시작 가이드 - 설치부터 첫 프로젝트까지"
      }
    }, {
      name: "index",
      route: "/ko/getting-started",
      frontMatter: {
        "sidebarTitle": "Index"
      }
    }, {
      name: "installation",
      route: "/ko/getting-started/installation",
      frontMatter: {
        "title": "설치 가이드",
        "description": "MoAI-ADK 설치 및 초기 설정 가이드 - 시스템 요구사항, 설치 방법, 환경 설정, 문제 해결"
      }
    }, {
      name: "quick-start",
      route: "/ko/getting-started/quick-start",
      frontMatter: {
        "title": "5분 빠른 시작",
        "description": "5분 만에 MoAI-ADK로 첫 프로젝트를 시작하는 방법 - 설치부터 첫 기능 개발까지"
      }
    }]
  }, {
    name: "guides",
    route: "/ko/guides",
    children: [{
      name: "advanced",
      route: "/ko/guides/advanced",
      frontMatter: {
        "title": "고급 가이드",
        "description": "엔터프라이즈급 기능과 마스터리를 다루는 고급자 가이드"
      }
    }, {
      name: "alfred",
      route: "/ko/guides/alfred",
      children: [{
        name: "index",
        route: "/ko/guides/alfred",
        frontMatter: {
          "title": "Alfred 워크플로우",
          "description": "Alfred 워크플로우 시스템 배우기"
        }
      }]
    }, {
      name: "beginner",
      route: "/ko/guides/beginner",
      frontMatter: {
        "title": "시작하기",
        "description": "MoAI-ADK를 처음 사용하는 개발자를 위한 완전 가이드"
      }
    }, {
      name: "index",
      route: "/ko/guides",
      frontMatter: {
        "title": "가이드",
        "description": "MoAI-ADK 사용을 위한 종합 가이드"
      }
    }, {
      name: "index",
      route: "/ko/guides",
      frontMatter: {
        "sidebarTitle": "Index"
      }
    }, {
      name: "intermediate",
      route: "/ko/guides/intermediate",
      frontMatter: {
        "title": "중급 가이드",
        "description": "고급 패턴과 실전 활용법을 다루는 중급자 가이드"
      }
    }, {
      name: "tdd",
      route: "/ko/guides/tdd",
      children: [{
        name: "index",
        route: "/ko/guides/tdd",
        frontMatter: {
          "title": "TDD 구현",
          "description": "MoAI-ADK의 체계적인 접근 방식으로 TDD 마스터하기"
        }
      }]
    }]
  }, {
    name: "index",
    route: "/ko",
    frontMatter: {
      "title": "MoAI-ADK: AI 기반 SPEC-First TDD 개발 프레임워크",
      "description": "AI 지원으로 신뢰할 수 있고 유지 관리 가능한 소프트웨어를 구축하세요. 요구사항부터 문서까지 완벽한 동기화를 통한 완전 자동화."
    }
  }, {
    name: "navigation-guide",
    route: "/ko/navigation-guide",
    frontMatter: {
      "sidebarTitle": "Navigation Guide"
    }
  }, {
    name: "output-style",
    route: "/ko/output-style",
    children: [{
      name: "best-practices",
      route: "/ko/output-style/best-practices",
      frontMatter: {
        "title": "모범 사례",
        "description": "MoAI-ADK 출력 스타일과 문서화 품질을 위한 모범 사례"
      }
    }, {
      name: "index",
      route: "/ko/output-style",
      frontMatter: {
        "sidebarTitle": "Index"
      }
    }, {
      name: "personas",
      route: "/ko/output-style/personas",
      frontMatter: {
        "title": "Alfred 적응형 페르소나",
        "description": "사용자 레벨별 개인화된 개발 경험과 동적 커뮤니케이션 패턴"
      }
    }, {
      name: "r2d2-agentic",
      route: "/ko/output-style/r2d2-agentic",
      frontMatter: {
        "title": "R2-D2 에이전트 코딩 스타일",
        "description": "🤖 AI 기반 실시간 문제 해결 및 코드 생성 방식"
      }
    }]
  }, {
    name: "reference",
    route: "/ko/reference",
    children: [{
      name: "agents",
      route: "/ko/reference/agents",
      children: [{
        name: "index",
        route: "/ko/reference/agents",
        frontMatter: {
          "title": "에이전트",
          "description": "다양한 에이전트와 그들의 역할 배우기"
        }
      }]
    }, {
      name: "cli-commands",
      route: "/ko/reference/cli-commands",
      frontMatter: {
        "title": "CLI 명령어",
        "description": "MoAI-ADK의 5개 핵심 CLI 명령어와 사용법"
      }
    }, {
      name: "core-modules",
      route: "/ko/reference/core-modules",
      frontMatter: {
        "title": "코어 모듈",
        "description": "MoAI-ADK의 핵심 모듈 아키텍처와 158개 Python 파일 구조"
      }
    }, {
      name: "index",
      route: "/ko/reference",
      frontMatter: {
        "title": "참고 자료",
        "description": "API 참조 및 기술 문서"
      }
    }, {
      name: "index",
      route: "/ko/reference",
      frontMatter: {
        "sidebarTitle": "Index"
      }
    }, {
      name: "skills",
      route: "/ko/reference/skills",
      children: [{
        name: "index",
        route: "/ko/reference/skills",
        frontMatter: {
          "title": "스킬",
          "description": "사용 가능한 스킬과 그것들을 사용하는 방법 발견하기"
        }
      }]
    }]
  }, {
    name: "skills",
    route: "/ko/skills",
    children: [{
      data: ko_skills_meta
    }, {
      name: "advanced-skills",
      route: "/ko/skills/advanced-skills",
      frontMatter: {
        "sidebarTitle": "Advanced Skills"
      }
    }, {
      name: "baas",
      route: "/ko/skills/baas",
      children: [{
        name: "auth0",
        route: "/ko/skills/baas/auth0",
        frontMatter: {
          "sidebarTitle": "Auth0"
        }
      }, {
        name: "clerk",
        route: "/ko/skills/baas/clerk",
        frontMatter: {
          "sidebarTitle": "Clerk"
        }
      }, {
        name: "cloudflare",
        route: "/ko/skills/baas/cloudflare",
        frontMatter: {
          "sidebarTitle": "Cloudflare"
        }
      }, {
        name: "convex",
        route: "/ko/skills/baas/convex",
        frontMatter: {
          "sidebarTitle": "Convex"
        }
      }, {
        name: "firebase",
        route: "/ko/skills/baas/firebase",
        frontMatter: {
          "title": "Firebase 완벽 가이드",
          "description": "Google Cloud Native BaaS - Firestore NoSQL, Cloud Functions, Firebase Auth로 모바일 최적화 백엔드 구축"
        }
      }, {
        name: "neon",
        route: "/ko/skills/baas/neon",
        frontMatter: {
          "sidebarTitle": "Neon"
        }
      }, {
        name: "railway",
        route: "/ko/skills/baas/railway",
        frontMatter: {
          "sidebarTitle": "Railway"
        }
      }, {
        name: "supabase",
        route: "/ko/skills/baas/supabase",
        frontMatter: {
          "title": "Supabase 완벽 가이드",
          "description": "Enterprise PostgreSQL BaaS - RLS, Realtime, Edge Functions으로 30분 내 Production 배포"
        }
      }, {
        name: "vercel",
        route: "/ko/skills/baas/vercel",
        frontMatter: {
          "sidebarTitle": "Vercel"
        }
      }]
    }, {
      name: "baas-ecosystem",
      route: "/ko/skills/baas-ecosystem",
      frontMatter: {
        "title": "BaaS Ecosystem - 12개 Production-Ready Skills",
        "description": "9개 플랫폼, 8개 아키텍처 패턴, AI 기반 의사결정 프레임워크로 30분 내 엔터프라이즈급 백엔드 구축"
      }
    }, {
      name: "ecosystem-upgrade-v4",
      route: "/ko/skills/ecosystem-upgrade-v4",
      frontMatter: {
        "title": "Skills Ecosystem v4.0 대규모 업그레이드",
        "description": "MoAI-ADK v0.23.1 역사적 성취 - 45개 문제 Skills 자동 복구 및 281+ Skills v4.0.0 Enterprise 업그레이드 완성"
      }
    }, {
      name: "foundation",
      route: "/ko/skills/foundation",
      frontMatter: {
        "title": "Foundation Skills",
        "description": "MoAI-ADK의 핵심 기반을 구성하는 Foundation Skills 상세 가이드"
      }
    }, {
      name: "overview",
      route: "/ko/skills/overview",
      frontMatter: {
        "title": "Skills 개요",
        "description": "MoAI-ADK의 292개 Claude Skills 시스템 소개 - v4.0.0 Enterprise 업그레이드 완료"
      }
    }, {
      name: "skill-development",
      route: "/ko/skills/skill-development",
      frontMatter: {
        "sidebarTitle": "Skill Development"
      }
    }, {
      name: "validation-system",
      route: "/ko/skills/validation-system",
      frontMatter: {
        "sidebarTitle": "Validation System"
      }
    }]
  }, {
    name: "troubleshooting",
    route: "/ko/troubleshooting",
    children: [{
      name: "index",
      route: "/ko/troubleshooting",
      frontMatter: {
        "sidebarTitle": "Index"
      }
    }]
  }, {
    name: "tutorials",
    route: "/ko/tutorials",
    children: [{
      data: ko_tutorials_meta
    }, {
      name: "index",
      route: "/ko/tutorials",
      frontMatter: {
        "title": "튜토리얼",
        "description": "MoAI-ADK 실전 튜토리얼 - 초급부터 고급까지"
      }
    }, {
      name: "tutorial-01-rest-api",
      route: "/ko/tutorials/tutorial-01-rest-api",
      frontMatter: {
        "title": "Tutorial 1: 첫 REST API 개발",
        "description": "FastAPI로 REST API를 30분만에 만들어봅니다",
        "duration": "30분",
        "difficulty": "초급",
        "tags": ["tutorial", "rest-api", "fastapi", "beginner"]
      }
    }, {
      name: "tutorial-02-jwt-auth",
      route: "/ko/tutorials/tutorial-02-jwt-auth",
      frontMatter: {
        "title": "Tutorial 2: JWT 인증 시스템 구현",
        "description": "실무 수준의 JWT 기반 인증 시스템을 1시간 만에 구축합니다",
        "duration": "1시간",
        "difficulty": "중급",
        "tags": ["tutorial", "jwt", "authentication", "oauth2", "security"]
      }
    }, {
      name: "tutorial-03-database-optimization",
      route: "/ko/tutorials/tutorial-03-database-optimization",
      frontMatter: {
        "title": "Tutorial 3: 데이터베이스 성능 최적화",
        "description": "N+1 문제, 인덱스, 캐싱 전략으로 데이터베이스 성능을 극대화합니다",
        "duration": "1시간",
        "difficulty": "고급",
        "tags": ["tutorial", "database", "optimization", "postgresql", "redis", "caching"]
      }
    }, {
      name: "tutorial-04-baas-supabase",
      route: "/ko/tutorials/tutorial-04-baas-supabase",
      frontMatter: {
        "title": "Tutorial 4: BaaS 플랫폼 통합 (Supabase)",
        "description": "Supabase로 백엔드 개발 속도를 극대화합니다",
        "duration": "45분",
        "difficulty": "중급",
        "tags": ["tutorial", "baas", "supabase", "authentication", "realtime", "storage"]
      }
    }, {
      name: "tutorial-05-mcp-server",
      route: "/ko/tutorials/tutorial-05-mcp-server",
      frontMatter: {
        "title": "Tutorial 5: MCP 서버 개발",
        "description": "Model Context Protocol로 Claude AI 도구를 확장합니다",
        "duration": "1시간",
        "difficulty": "고급",
        "tags": ["tutorial", "mcp", "ai", "claude", "protocol"]
      }
    }]
  }]
}];