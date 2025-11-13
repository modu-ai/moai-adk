---
title: "코드 예제"
description: "MoAI-ADK를 활용한 실전 코드 예제 모음"
---

# 코드 예제

MoAI-ADK를 사용한 실전 코드 예제를 카테고리별로 제공합니다. 모든 예제는 실행 가능하며, 테스트 코드와 Best Practices가 포함되어 있습니다.

## 📚 카테고리별 예제

### 🌐 REST API 예제

FastAPI를 활용한 RESTful API 구현 예제입니다.

- [기본 CRUD 작업](/ko/examples/rest-api/basic-crud) - 생성, 읽기, 수정, 삭제 구현
- [페이지네이션 & 정렬](/ko/examples/rest-api/pagination) - 대용량 데이터 처리
- [필터링 & 검색](/ko/examples/rest-api/filtering) - 동적 쿼리 구현
- [에러 처리 & 검증](/ko/examples/rest-api/error-handling) - 안전한 API 설계

### 🗄️ 데이터베이스 예제

SQLAlchemy와 Alembic을 사용한 데이터베이스 관리 예제입니다.

- [Alembic 마이그레이션](/ko/examples/database/migrations) - 스키마 버전 관리
- [SQLAlchemy 관계](/ko/examples/database/relationships) - 테이블 관계 설정
- [트랜잭션 처리](/ko/examples/database/transactions) - 데이터 일관성 보장
- [쿼리 최적화](/ko/examples/database/query-optimization) - 성능 향상 기법

### 🧪 테스팅 예제

Pytest를 활용한 체계적인 테스트 작성 예제입니다.

- [Pytest 단위 테스트](/ko/examples/testing/unit-tests) - TDD 기반 개발
- [통합 테스트](/ko/examples/testing/integration-tests) - 시스템 전체 테스트
- [테스트 픽스처](/ko/examples/testing/fixtures) - 재사용 가능한 테스트 데이터
- [외부 API 모킹](/ko/examples/testing/mocking) - 의존성 격리

### 🔐 인증 예제

보안 인증 시스템 구축 예제입니다.

- [JWT 토큰 생성](/ko/examples/authentication/jwt-basic) - 기본 JWT 인증
- [리프레시 토큰](/ko/examples/authentication/refresh-tokens) - 장기 세션 관리
- [OAuth2 인증](/ko/examples/authentication/oauth2) - 소셜 로그인
- [역할 기반 접근 제어](/ko/examples/authentication/rbac) - 권한 관리

### ⚡ 성능 예제

시스템 성능 최적화 기법 예제입니다.

- [Redis 캐싱](/ko/examples/performance/caching) - 응답 속도 향상
- [커넥션 풀링](/ko/examples/performance/connection-pooling) - 리소스 효율화
- [지연 로딩](/ko/examples/performance/lazy-loading) - 메모리 최적화
- [배치 처리](/ko/examples/performance/batch-processing) - 대량 데이터 처리

### 🛡️ 보안 예제

애플리케이션 보안 강화 예제입니다.

- [입력 검증](/ko/examples/security/input-validation) - Pydantic 검증
- [SQL 인젝션 방지](/ko/examples/security/sql-injection-prevention) - 안전한 쿼리
- [XSS 방어](/ko/examples/security/xss-protection) - 크로스 사이트 스크립팅 방지
- [속도 제한](/ko/examples/security/rate-limiting) - API 남용 방지

## 🎯 난이도별 예제

### 초급 (Beginner)
- REST API 기본 CRUD
- JWT 기본 인증
- 단위 테스트
- 입력 검증

### 중급 (Intermediate)
- 페이지네이션 & 필터링
- SQLAlchemy 관계
- 통합 테스트
- Redis 캐싱
- OAuth2 인증

### 고급 (Advanced)
- 쿼리 최적화
- 배치 처리
- 역할 기반 접근 제어
- 트랜잭션 처리

## 💡 예제 사용 가이드

### 1. 예제 코드 실행 방법

```bash
# 가상환경 생성
uv venv

# 의존성 설치
uv pip install -r requirements.txt

# 예제 실행
python examples/rest-api/basic-crud.py
```

### 2. 테스트 실행 방법

```bash
# 전체 테스트 실행
pytest tests/

# 특정 테스트 파일 실행
pytest tests/test_crud.py

# 커버리지 포함 실행
pytest --cov=app tests/
```

### 3. 프로젝트에 적용하기

1. 예제 코드를 프로젝트에 복사
2. 환경에 맞게 설정 수정
3. 테스트 코드 작성
4. SPEC 문서 작성
5. TDD 사이클로 개발

## 🏷️ 태그별 예제

- `#fastapi` - FastAPI 프레임워크 관련
- `#sqlalchemy` - ORM 및 데이터베이스 관련
- `#pytest` - 테스트 관련
- `#security` - 보안 관련
- `#performance` - 성능 최적화 관련
- `#authentication` - 인증/인가 관련

## 📖 관련 문서

### 튜토리얼
- [Tutorial 01: FastAPI + SQLAlchemy 프로젝트](/ko/tutorials/tutorial-01-fastapi)
- [Tutorial 02: JWT 인증 시스템](/ko/tutorials/tutorial-02-jwt-auth)
- [Tutorial 03: TDD로 API 개발하기](/ko/tutorials/tutorial-03-tdd-api)

### 가이드
- [SPEC 작성 가이드](/ko/guides/spec-writing)
- [TDD 개발 가이드](/ko/guides/tdd-development)
- [Git 워크플로우](/ko/guides/git-workflow)

### 레퍼런스
- [Skill 레퍼런스](/ko/skills/)
- [Agent 레퍼런스](/ko/agents/)
- [Command 레퍼런스](/ko/reference/commands)

## 🤝 기여 가이드

예제 추가를 원하시면 다음 형식을 따라주세요:

1. **파일 구조**
   - 제목, 카테고리, 난이도, 태그 메타데이터
   - 개요 및 사용 사례
   - 완전한 실행 가능한 코드
   - 단계별 설명
   - 테스트 코드
   - Best Practices & 주의사항
   - 관련 예제 링크

2. **코드 품질 기준**
   - Python 3.11+ 문법
   - Type hints 사용
   - 한국어 docstring
   - PEP 8 준수
   - 프로덕션 수준 코드

3. **문서 작성 기준**
   - 명확한 한국어 설명
   - 기술 용어: 한글 (English) 형식
   - 코드 주석은 한국어
   - 변수명은 영어 (관례)

## 📬 피드백

예제에 대한 피드백이나 개선 제안은 언제든 환영합니다:

- GitHub Issues: 버그 리포트 또는 개선 제안
- GitHub Discussions: 질문 또는 아이디어 공유
- Pull Request: 새로운 예제 기여

---

**다음 단계**: 관심 있는 카테고리를 선택하고 예제를 살펴보세요!
