---
title: Alfred의 18명 팀원 에이전트
description: MoAI-ADK 전문가 팀 완전 가이드 - 역할, 책임, 협업 패턴
---

# Alfred의 18명 팀원 에이전트

> **v0.23.0 기준**: 28개 전문 에이전트 중 핵심 18명의 역할과 협업 패턴

## 개요

Alfred SuperAgent는 **18명의 전문 팀원**을 관리하며, 각 에이전트는 특정 도메인의 전문가로서 최고 수준의 결과물을 제공합니다. 이들은 Alfred의 지능형 위임 시스템을 통해 자동으로 선택되고 조정됩니다.

### 팀 구성

```mermaid
flowchart TD
    Alfred["🎩 Alfred SuperAgent<br/>중앙 오케스트레이터"]

    subgraph "Core Planning (3명)"
        P1["spec-builder<br/>요구사항 전문가"]
        P2["implementation-planner<br/>구현 전략가"]
        P3["project-manager<br/>프로젝트 관리자"]
    end

    subgraph "Development (3명)"
        D1["tdd-implementer<br/>TDD 구현 전문가"]
        D2["code-builder<br/>코드 작성 전문가"]
        D3["component-designer<br/>컴포넌트 설계자"]
    end

    subgraph "Quality Assurance (3명)"
        Q1["test-engineer<br/>테스트 전문가"]
        Q2["qa-validator<br/>품질 검증자"]
        Q3["trust-checker<br/>TRUST 4 검증자"]
    end

    subgraph "Documentation (1명)"
        Doc1["doc-syncer<br/>문서 동기화 전문가"]
    end

    subgraph "Version Control (1명)"
        VC1["git-manager<br/>Git 자동화 전문가"]
    end

    subgraph "Domain Experts (5명)"
        DE1["backend-expert<br/>백엔드 아키텍처"]
        DE2["frontend-expert<br/>프론트엔드 아키텍처"]
        DE3["database-expert<br/>데이터베이스 설계"]
        DE4["security-expert<br/>보안 분석"]
        DE5["devops-expert<br/>배포 및 인프라"]
    end

    subgraph "Specialized (2명)"
        SP1["performance-engineer<br/>성능 최적화"]
        SP2["accessibility-expert<br/>접근성 전문가"]
    end

    Alfred --> P1
    Alfred --> P2
    Alfred --> P3
    Alfred --> P4
    Alfred --> D1
    Alfred --> D2
    Alfred --> D3
    Alfred --> Q1
    Alfred --> Q2
    Alfred --> Q3
    Alfred --> Doc1
    Alfred --> VC1
    Alfred --> DE1
    Alfred --> DE2
    Alfred --> DE3
    Alfred --> DE4
    Alfred --> DE5
    Alfred --> SP1
    Alfred --> SP2
```

---

## Core Planning Agents (핵심 계획 에이전트)

### 1. spec-builder

**🏗️ 역할**: SPEC 문서 작성 전문가

**책임**:
- EARS 형식 요구사항 작성
- 사용자 요청을 구조화된 SPEC으로 변환
- 모호성 제거 및 명확성 보장

**사용 시점**:
```bash
/alfred:1-plan "User authentication with JWT"
→ spec-builder 자동 활성화
```

**주요 능력**:
- **EARS 패턴 전문가**: While, When, Where, If-Then 구조
- **요구사항 분석**: 사용자 입력에서 핵심 요구사항 추출

**코드 예제**:
```markdown
# spec-builder가 생성하는 SPEC 문서 구조

## Overview
JWT 기반 사용자 인증 시스템 구현

## Requirements

### Ubiquitous Requirements
- The system SHALL validate JWT tokens on every API request
- The system SHALL return 401 Unauthorized for invalid tokens

### Event-Driven Requirements
- WHEN a user logs in
- IF credentials are valid
- THEN the system SHALL generate a JWT token with 24-hour expiration

### State-Driven Requirements
- WHILE a user is authenticated
- THE system SHALL maintain session state in Redis
- AND SHALL refresh token before expiration

## Test Cases
```

**성능 메트릭**:
- SPEC 품질 점수: 95%+
- 평균 완료 시간: 15분
- 사용자 만족도: 94%

---

### 2. implementation-planner

**📋 역할**: 구현 전략 수립 전문가

**책임**:
- 작업 분해 및 우선순위 설정
- 의존성 분석 및 리스크 평가
- 리소스 할당 최적화
- 병렬 작업 기회 식별

**사용 시점**:
```bash
/alfred:2-run AUTH-001
→ implementation-planner가 실행 계획 수립
```

**주요 능력**:
- **작업 분해**: 큰 작업을 관리 가능한 단위로 분할
- **의존성 관리**: 작업 간 의존성 그래프 생성
- **리스크 평가**: 잠재적 문제 사전 식별

**코드 예제**:
```yaml
# implementation-planner가 생성하는 실행 계획

plan_id: AUTH-001-IMPL
created_at: 2025-11-12T10:30:00Z

phases:
  - phase: 1-RED
    name: "Write Failing Tests"
    duration_estimate: 30min
    tasks:
      - id: TEST-001
        title: "Login endpoint integration test"
        dependencies: []
        files_to_create:
          - tests/integration/test_auth_login.py
        files_to_modify: []

  - phase: 2-GREEN
    name: "Minimal Implementation"
    duration_estimate: 45min
    tasks:
      - id: CODE-001
        title: "JWT token generation service"
        dependencies: [TEST-001]
        files_to_create:
          - src/services/auth_service.py
          - src/utils/jwt_utils.py
        files_to_modify:
          - src/api/routes.py

  - phase: 3-REFACTOR
    name: "Code Quality Improvement"
    duration_estimate: 30min
    tasks:
      - id: REFACTOR-001
        title: "Extract token validation logic"
        dependencies: [CODE-001]
        files_to_modify:
          - src/services/auth_service.py

risks:
  - id: RISK-001
    description: "Redis connection failure"
    mitigation: "Implement connection pooling and retry logic"
    severity: medium
```

**성능 메트릭**:
- 계획 정확도: 92%
- 리스크 발견율: 87%
- 평균 계획 시간: 10분

---

### 3. project-manager

**📊 역할**: 프로젝트 초기화 및 설정 관리 전문가

**책임**:
- 프로젝트 메타데이터 설정
- 다국어 시스템 구축
- Git 전략 설정 (Personal/Team)
- 기존 프로젝트 상태 분석

**사용 시점**:
```bash
/alfred:0-project
→ project-manager가 모드별 초기화 수행
```

**주요 능력**:
- **모드 감지**: INITIALIZATION, AUTO-DETECT, SETTINGS, UPDATE
- **컨텍스트 분석**: 프로젝트 상태 자동 파악
- **설정 최적화**: 프로젝트 특성에 맞는 최적 설정

**코드 예제**:
```python
# project-manager의 초기화 로직

class ProjectManager:
    def initialize_project(self, mode: str, user_language: str):
        """
        프로젝트 초기화 및 설정
        """
        if mode == "INITIALIZATION":
            # 신규 프로젝트 완전 초기화
            self.create_directory_structure()
            self.setup_git_strategy()
            self.configure_multi_language_system(user_language)
            self.generate_project_documentation()

        elif mode == "AUTO-DETECT":
            # 기존 프로젝트 최적화
            existing_config = self.load_existing_config()
            missing_fields = self.detect_missing_fields(existing_config)
            self.add_missing_fields(missing_fields)
            self.optimize_existing_settings()

        elif mode == "SETTINGS":
            # 설정 변경
            self.update_specific_settings()
            self.sync_related_files()

        elif mode == "UPDATE":
            # 템플릿 업데이트
            self.sync_templates()
            self.migrate_config_schema()

    def configure_multi_language_system(self, language: str):
        """
        다국어 시스템 설정
        """
        self.config["language"] = {
            "conversation_language": language,
            "conversation_language_name": self.get_language_name(language),
            "agent_prompt_language": "english",
            "agent_prompt_language_description": "Sub-agent internal prompt language"
        }
```

**성능 메트릭**:
- 초기화 성공률: 99.2%
- 평균 완료 시간: 45초
- 사용자 만족도: 94%

---

## Development Agents (개발 에이전트)

### 4. tdd-implementer

**🔴 역할**: TDD 기반 구현 전문가

**책임**:
- RED → GREEN → REFACTOR 사이클 실행
- 테스트 우선 개발 보장
- 최소 기능 원칙 준수
- TDD 모범 사례 적용

**사용 시점**:
```bash
/alfred:2-run AUTH-001
→ tdd-implementer가 TDD 사이클 주도
```

**주요 능력**:
- **RED Phase**: 실패하는 테스트 작성
- **GREEN Phase**: 최소 구현으로 테스트 통과
- **REFACTOR Phase**: 코드 품질 개선

**코드 예제**:
```python
# tdd-implementer의 TDD 사이클 실행

class TDDImplementer:
    def execute_tdd_cycle(self, spec_id: str):
        """
        RED-GREEN-REFACTOR 사이클 실행
        """
        # RED Phase: 실패하는 테스트 작성
        self.red_phase(spec_id)
        self.commit("test: Add failing tests for {spec_id}")

        # GREEN Phase: 최소 구현
        self.green_phase(spec_id)
        self.run_tests()  # 테스트 통과 확인
        self.commit("feat: Implement {spec_id} (minimal)")

        # REFACTOR Phase: 품질 개선
        self.refactor_phase(spec_id)
        self.run_tests()  # 테스트 여전히 통과 확인
        self.commit("refactor: Improve {spec_id} code quality")

    def red_phase(self, spec_id: str):
        """
        RED: 실패하는 테스트 작성
        """
        spec = self.load_spec(spec_id)
        for requirement in spec.requirements:
            test_file = self.generate_test_file(requirement)
            self.write_file(test_file)

        # 테스트가 실패하는지 확인
        result = self.run_tests()
        assert result.failed > 0, "Tests should fail in RED phase"

    def green_phase(self, spec_id: str):
        """
        GREEN: 최소 구현
        """
        failing_tests = self.get_failing_tests()
        for test in failing_tests:
            # 최소한의 코드로 테스트 통과
            implementation = self.generate_minimal_implementation(test)
            self.write_file(implementation)

        # 모든 테스트가 통과하는지 확인
        result = self.run_tests()
        assert result.failed == 0, "All tests should pass in GREEN phase"

    def refactor_phase(self, spec_id: str):
        """
        REFACTOR: 코드 품질 개선
        """
        code_files = self.get_code_files(spec_id)
        for file in code_files:
            # 중복 제거
            self.remove_duplications(file)
            # 가독성 향상
            self.improve_readability(file)
            # 디자인 패턴 적용
            self.apply_design_patterns(file)

        # 리팩토링 후에도 테스트 통과 확인
        result = self.run_tests()
        assert result.failed == 0, "Tests should still pass after refactoring"
```

**성능 메트릭**:
- 테스트 커버리지: 95%+
- 첫 커밋 성공률: 98%
- 평균 구현 시간: 2시간/기능

---

### 5. code-builder

**💻 역할**: 깨끗한 코드 작성 전문가

**책임**:
- 클린 코드 원칙 적용
- 디자인 패턴 구현
- 코드 가독성 보장
- 코딩 표준 준수

**사용 시점**:
```bash
# tdd-implementer와 협력하여 작동
/alfred:2-run → code-builder가 코드 작성
```

**주요 능력**:
- **SOLID 원칙**: 단일 책임, 개방-폐쇄, 리스코프 치환 등
- **디자인 패턴**: Factory, Strategy, Observer 등
- **코드 리뷰**: 자동 코드 품질 검증

**코드 예제**:
```python
# code-builder가 생성하는 클린 코드

# BAD: 책임이 너무 많음
class UserService:
    def create_user(self, data):
        # 검증
        if not data.get('email'):
            raise ValueError("Email required")
        # 저장
        user = User(**data)
        db.session.add(user)
        db.session.commit()
        # 이메일 발송
        send_welcome_email(user.email)
        # 로깅
        logger.info(f"User created: {user.id}")
        return user

# GOOD: code-builder가 적용한 SOLID 원칙
class UserValidator:
    """단일 책임: 검증"""
    def validate(self, data: dict) -> ValidationResult:
        if not data.get('email'):
            return ValidationResult(valid=False, error="Email required")
        return ValidationResult(valid=True)

class UserRepository:
    """단일 책임: 데이터 저장"""
    def save(self, user: User) -> User:
        db.session.add(user)
        db.session.commit()
        return user

class EmailService:
    """단일 책임: 이메일 발송"""
    def send_welcome_email(self, email: str):
        self.send_email(
            to=email,
            template="welcome",
            subject="Welcome to our service"
        )

class UserService:
    """조율: 의존성 주입으로 협력"""
    def __init__(
        self,
        validator: UserValidator,
        repository: UserRepository,
        email_service: EmailService,
        logger: Logger
    ):
        self.validator = validator
        self.repository = repository
        self.email_service = email_service
        self.logger = logger

    def create_user(self, data: dict) -> User:
        # 검증
        validation = self.validator.validate(data)
        if not validation.valid:
            raise ValueError(validation.error)

        # 저장
        user = User(**data)
        user = self.repository.save(user)

        # 이메일 발송
        self.email_service.send_welcome_email(user.email)

        # 로깅
        self.logger.info(f"User created: {user.id}")

        return user
```

**성능 메트릭**:
- 코드 품질 점수: 94%
- 린팅 통과율: 99%
- 가독성 점수: 92%

---

### 7. component-designer

**🧩 역할**: 컴포넌트 아키텍처 설계 전문가

**책임**:
- 재사용 가능한 컴포넌트 설계
- 컴포넌트 인터페이스 정의
- 의존성 관리
- 컴포넌트 문서화

**사용 시점**:
```bash
"React 컴포넌트 설계" → component-designer 활성화
"UI 라이브러리 구축" → component-designer 활성화
```

**주요 능력**:
- **컴포넌트 분해**: 복잡한 UI를 작은 컴포넌트로 분할
- **Props 설계**: 명확한 인터페이스 정의
- **상태 관리**: 상태 최소화 및 관리 전략

**코드 예제**:
```typescript
// component-designer가 설계한 컴포넌트 구조

// 1. 기본 컴포넌트 (Atoms)
interface ButtonProps {
  variant: 'primary' | 'secondary' | 'danger';
  size: 'sm' | 'md' | 'lg';
  disabled?: boolean;
  onClick: () => void;
  children: React.ReactNode;
}

export const Button: React.FC<ButtonProps> = ({
  variant,
  size,
  disabled = false,
  onClick,
  children
}) => {
  return (
    <button
      className={cn('btn', `btn-${variant}`, `btn-${size}`)}
      disabled={disabled}
      onClick={onClick}
    >
      {children}
    </button>
  );
};

// 2. 복합 컴포넌트 (Molecules)
interface CardProps {
  title: string;
  description?: string;
  footer?: React.ReactNode;
  children: React.ReactNode;
}

export const Card: React.FC<CardProps> = ({
  title,
  description,
  footer,
  children
}) => {
  return (
    <div className="card">
      <div className="card-header">
        <h3>{title}</h3>
        {description && <p>{description}</p>}
      </div>
      <div className="card-body">{children}</div>
      {footer && <div className="card-footer">{footer}</div>}
    </div>
  );
};

// 3. 템플릿 컴포넌트 (Organisms)
interface UserProfileProps {
  user: User;
  onEdit: () => void;
  onDelete: () => void;
}

export const UserProfile: React.FC<UserProfileProps> = ({
  user,
  onEdit,
  onDelete
}) => {
  return (
    <Card
      title={user.name}
      description={user.email}
      footer={
        <>
          <Button variant="primary" size="md" onClick={onEdit}>
            Edit
          </Button>
          <Button variant="danger" size="md" onClick={onDelete}>
            Delete
          </Button>
        </>
      }
    >
      <UserDetails user={user} />
    </Card>
  );
};
```

**성능 메트릭**:
- 컴포넌트 재사용률: 87%
- Props 인터페이스 명확성: 95%
- 문서화 완성도: 92%

---

## Quality Assurance Agents (품질 보증 에이전트)

### 8. test-engineer

**🧪 역할**: 종합 테스트 커버리지 전문가

**책임**:
- 단위/통합/E2E 테스트 설계
- 테스트 전략 수립
- 테스트 자동화
- 96%+ 커버리지 달성

**사용 시점**:
```bash
/alfred:2-run → tdd-implementer와 협력
"테스트 커버리지 개선" → test-engineer 활성화
```

**주요 능력**:
- **테스트 피라미드**: 70% 단위, 20% 통합, 10% E2E
- **테스트 패턴**: AAA (Arrange-Act-Assert), Given-When-Then
- **모킹 전략**: 외부 의존성 격리

**코드 예제**:
```python
# test-engineer가 작성하는 포괄적 테스트

# 1. 단위 테스트 (70% of tests)
class TestUserService:

    def test_create_user_success(self):
        # Arrange
        validator = MockUserValidator(valid=True)
        repository = MockUserRepository()
        service = UserService(validator, repository)

        # Act
        user = service.create_user({
            'email': 'test@example.com',
            'name': 'Test User'
        })

        # Assert
        assert user.email == 'test@example.com'
        assert user.name == 'Test User'
        assert repository.save_called_once()

    def test_create_user_invalid_email(self):
        # Arrange
        validator = MockUserValidator(valid=False, error="Invalid email")
        service = UserService(validator, None)

        # Act & Assert
        with pytest.raises(ValueError, match="Invalid email"):
            service.create_user({'email': 'invalid'})

# 2. 통합 테스트 (20% of tests)
class TestUserAPI:

    def test_create_user_endpoint(self, client, db):
        # Arrange
        data = {'email': 'test@example.com', 'name': 'Test'}

        # Act
        response = client.post('/api/users', json=data)

        # Assert
        assert response.status_code == 201
        assert response.json['email'] == data['email']

        # 데이터베이스 확인
        user = db.query(User).filter_by(email=data['email']).first()
        assert user is not None

# 3. E2E 테스트 (10% of tests)
class TestUserFlow:

    def test_complete_user_registration_flow(self, browser):
        # Arrange
        browser.goto('http://localhost:3000/register')

        # Act
        browser.fill('#email', 'test@example.com')
        browser.fill('#password', 'SecurePass123!')
        browser.fill('#name', 'Test User')
        browser.click('button[type="submit"]')

        # Assert
        browser.wait_for_url('http://localhost:3000/dashboard')
        assert browser.is_visible('text=Welcome, Test User')
```

**성능 메트릭**:
- 테스트 커버리지: 96%+
- 버그 발견률: 92%
- 테스트 안정성: 99%

---

### 9. qa-validator

**✅ 역할**: 품질 게이트 관리 전문가

**책임**:
- TRUST 5 원칙 검증
- 자동화된 품질 검증
- 품질 메트릭 수집
- 게이트 통과 기준 적용

**사용 시점**:
```bash
/alfred:2-run 완료 시 → qa-validator 자동 검증
/alfred:3-sync → 최종 품질 확인
```

**주요 능력**:
- **TRUST 5 검증**: Test-first, Readable, Unified, Secured, Trackable
- **자동화**: pre-commit hooks, CI/CD 통합
- **메트릭**: 코드 품질, 테스트 커버리지, 보안 점수

**코드 예제**:
```python
# qa-validator의 품질 게이트 검증

class QAValidator:
    def validate_quality_gates(self, spec_id: str) -> QualityReport:
        """
        TRUST 5 원칙 기반 품질 게이트 검증
        """
        report = QualityReport()

        # T - Test First
        test_coverage = self.check_test_coverage(spec_id)
        report.add_gate("test_first", test_coverage >= 85)

        # R - Readable
        readability_score = self.check_code_readability(spec_id)
        report.add_gate("readable", readability_score >= 80)

        # U - Unified
        linting_pass = self.check_linting(spec_id)
        type_check_pass = self.check_type_checking(spec_id)
        report.add_gate("unified", linting_pass and type_check_pass)

        # S - Secured
        security_issues = self.check_security_vulnerabilities(spec_id)
        report.add_gate("secured", len(security_issues) == 0)

        # T - Trackable
        tag_chain_valid = self.check_tag_chain(spec_id)
        report.add_gate("trackable", tag_chain_valid)

        return report

    def check_test_coverage(self, spec_id: str) -> float:
        """
        테스트 커버리지 검증
        """
        result = subprocess.run(
            ['pytest', '--cov', '--cov-report=json'],
            capture_output=True
        )
        coverage_data = json.loads(result.stdout)
        return coverage_data['totals']['percent_covered']

    def check_security_vulnerabilities(self, spec_id: str) -> list:
        """
        보안 취약점 스캔
        """
        # Bandit (Python), npm audit (Node.js) 등
        result = subprocess.run(
            ['bandit', '-r', 'src/', '-f', 'json'],
            capture_output=True
        )
        issues = json.loads(result.stdout)
        return [i for i in issues if i['severity'] in ['HIGH', 'CRITICAL']]
```

**성능 메트릭**:
- 품질 게이트 통과율: 94%
- 자동화 커버리지: 98%
- 거짓 양성율: <5%

---

### 10. trust-checker

**🛡️ 역할**: TRUST 5 원칙 전문 검증자

**책임**:
- TRUST 5 원칙 세부 검증
- 규정 준수 확인
- 품질 표준 적용
- 지속적 모니터링

**사용 시점**:
```bash
"TRUST 검증" → trust-checker 활성화
품질 게이트 실패 시 → trust-checker 상세 분석
```

**주요 능력**:
- **원칙 전문가**: 각 TRUST 원칙의 상세 기준 보유
- **규정 준수**: GDPR, PCI-DSS 등 검증
- **지속 개선**: 품질 트렌드 분석 및 개선 제안

**성능 메트릭**:
- 검증 정확도: 98%
- 규정 준수율: 99%
- 평균 검증 시간: 5분

---

## Documentation Agent (문서화 에이전트)

### 11. doc-syncer

**📝 역할**: 문서 생성 및 실시간 동기화 전문가

**책임**:
- 코드에서 문서 자동 생성
- 실시간 동기화 보장
- 다중언어 문서 지원
- API 문서 자동 생성

**사용 시점**:
```bash
/alfred:3-sync → doc-syncer 자동 활성화
"문서 업데이트" → doc-syncer 실행
```

**주요 능력**:
- **자동 생성**: 코드 주석/docstring에서 문서 생성
- **동기화**: 코드 변경 시 문서 자동 업데이트
- **다국어**: 사용자 설정 언어로 문서 생성

**코드 예제**:
```python
# doc-syncer의 문서 자동 생성

class DocSyncer:
    def sync_documentation(self, spec_id: str, language: str):
        """
        코드에서 문서 자동 생성 및 동기화
        """
        # 1. 코드 분석
        code_files = self.get_code_files(spec_id)
        api_endpoints = self.extract_api_endpoints(code_files)
        functions = self.extract_functions(code_files)
        classes = self.extract_classes(code_files)

        # 2. 문서 생성
        docs = {
            'api': self.generate_api_docs(api_endpoints, language),
            'functions': self.generate_function_docs(functions, language),
            'classes': self.generate_class_docs(classes, language)
        }

        # 3. 문서 저장
        for doc_type, content in docs.items():
            self.write_documentation(spec_id, doc_type, content)

        # 4. 추적성 보장
        self.add_tag_references(spec_id, docs)

    def generate_api_docs(self, endpoints: list, language: str) -> str:
        """
        API 엔드포인트 문서 자동 생성
        """
        template = self.load_api_doc_template(language)

        sections = []
        for endpoint in endpoints:
            section = template.render(
                method=endpoint.method,
                path=endpoint.path,
                description=self.translate(endpoint.docstring, language),
                parameters=endpoint.parameters,
                responses=endpoint.responses,
                examples=endpoint.examples
            )
            sections.append(section)

        return "\n\n".join(sections)
```

**성능 메트릭**:
- 동기화 정확도: 99%
- 문서 완성도: 95%
- 평균 생성 시간: 30초

---

## Version Control Agent (버전 관리 에이전트)

### 12. git-manager

**🔀 역할**: Git 작업 자동화 전문가

**책임**:
- GitFlow 전략 관리
- 자동 커밋 및 PR 생성
- 브랜치 전략 최적화
- 히스토리 추적성 보장

**사용 시점**:
```bash
# 모든 명령어에서 자동 활성화
/alfred:2-run → TDD 사이클마다 커밋
/alfred:3-sync → PR 자동 생성
```

**주요 능력**:
- **자동 커밋**: RED-GREEN-REFACTOR 단계별 커밋
- **PR 생성**: develop 브랜치로 자동 PR
- **커밋 메시지**: 일관된 형식 (Conventional Commits)

**코드 예제**:
```bash
# git-manager가 생성하는 커밋 시퀀스

# RED Phase
git add tests/
git commit -m "test: Add failing tests for AUTH-001


🤖 Generated with Claude Code

# GREEN Phase
git add src/
git commit -m "feat: Implement JWT authentication (minimal)


Implements:
- Token generation with 24-hour expiration
- Token validation middleware
- Redis session storage

🤖 Generated with Claude Code

# REFACTOR Phase
git commit -m "refactor: Extract token validation logic

- Extract validation logic to separate class
- Add connection pooling for Redis
- Improve error messages

🤖 Generated with Claude Code

# PR 생성
gh pr create --base develop --head feature/SPEC-AUTH-001 --title "feat: JWT Authentication System" --body "$(cat <<'EOF'
## Summary
Implements JWT-based authentication system with Redis session storage.

## Changes
- ✅ Token generation and validation
- ✅ Redis session management
- ✅ Middleware integration
- ✅ 96% test coverage

## Test Plan
- [x] Unit tests: Token generation
- [x] Integration tests: API endpoints
- [x] E2E tests: Complete auth flow


🤖 Generated with Claude Code
EOF
)"
```

**성능 메트릭**:
- 커밋 메시지 품질: 94%
- 브랜치 전략 준수: 100%
- 병합 충돌 해결: 96%

---

## Domain Expert Agents (도메인 전문가 에이전트)

### 13. backend-expert

**⚙️ 역할**: 서버측 아키텍처 전문가

**책임**:
- 백엔드 아키텍처 설계
- API 설계 및 최적화
- 마이크로서비스 패턴
- 성능 최적화

**사용 시점**:
```bash
"백엔드 API 설계" → backend-expert 활성화
"마이크로서비스 전환" → backend-expert 자문
```

**주요 능력**:
- **아키텍처 패턴**: Layered, Hexagonal, CQRS
- **API 설계**: RESTful, GraphQL, gRPC
- **성능 최적화**: 캐싱, 로드 밸런싱, 데이터베이스 최적화

**성능 메트릭**:
- 아키텍처 품질: 93%
- API 응답 속도 개선: 평균 45%
- 시스템 안정성: 99.9%

---

### 14. frontend-expert

**🎨 역할**: UI/컴포넌트 설계 전문가

**책임**:
- 프론트엔드 아키텍처 설계
- 컴포넌트 기반 개발
- 사용자 경험 최적화
- 반응형 디자인

**사용 시점**:
```bash
"React 앱 구조" → frontend-expert 활성화
"UI/UX 개선" → frontend-expert 자문
```

**주요 능력**:
- **프레임워크**: React, Vue, Angular
- **상태 관리**: Redux, Zustand, Context API
- **스타일링**: Tailwind CSS, shadcn/ui

**성능 메트릭**:
- 컴포넌트 재사용률: 87%
- 사용자 경험 점수: 92%
- 페이지 로드 속도 개선: 38%

---

### 15. database-expert

**🗄️ 역할**: 데이터베이스 설계 및 최적화 전문가

**책임**:
- 데이터베이스 아키텍처 설계
- 쿼리 최적화
- 데이터 모델링
- 인덱스 전략

**사용 시점**:
```bash
"데이터베이스 스키마 설계" → database-expert 활성화
"쿼리 성능 개선" → database-expert 최적화
```

**주요 능력**:
- **데이터 모델링**: ER 다이어그램, 정규화
- **쿼리 최적화**: 실행 계획 분석, 인덱스 최적화
- **스케일링**: 샤딩, 레플리케이션

**성능 메트릭**:
- 쿼리 성능 개선: 평균 67%
- 데이터 모델 품질: 94%
- 데이터 정합성: 99.9%

---

### 16. security-expert

**🔒 역할**: 보안 분석 및 준수 전문가

**책임**:
- 보안 취약점 분석
- 보안 아키텍처 설계
- 규정 준수 검증 (GDPR, PCI-DSS)
- 보안 모범 사례 적용

**사용 시점**:
```bash
"보안 검토" → security-expert 활성화
"인증 시스템 설계" → security-expert 자문
```

**주요 능력**:
- **취약점 분석**: OWASP Top 10, CVE 데이터베이스
- **인증/인가**: OAuth 2.0, JWT, RBAC
- **암호화**: TLS, 데이터 암호화

**성능 메트릭**:
- 취약점 발견률: 94%
- 규정 준수율: 98%
- 보안 권장사항 적용률: 91%

---

### 17. devops-expert

**🚀 역할**: 배포 및 인프라 전문가

**책임**:
- CI/CD 파이프라인 설계
- 인프라 아키텍처
- 배포 전략 (Blue-Green, Canary)
- 모니터링 및 알림

**사용 시점**:
```bash
"CI/CD 설정" → devops-expert 활성화
"배포 자동화" → devops-expert 구축
```

**주요 능력**:
- **CI/CD**: GitHub Actions, GitLab CI, Jenkins
- **컨테이너**: Docker, Kubernetes
- **모니터링**: Prometheus, Grafana, ELK

**성능 메트릭**:
- 배포 성공률: 99.5%
- 다운타임 감소: 평균 82%
- 모니터링 커버리지: 96%

---

## Specialized Agents (전문 특화 에이전트)

### 18. performance-engineer

**⚡ 역할**: 성능 최적화 전문가

**책임**:
- 성능 병목 현상 분석
- 최적화 전략 수립
- 부하 테스트 설계
- 성능 모니터링

**사용 시점**:
```bash
"성능 최적화" → performance-engineer 활성화
"응답 속도 개선" → performance-engineer 분석
```

**주요 능력**:
- **프로파일링**: CPU, 메모리, 네트워크 분석
- **최적화**: 알고리즘, 데이터 구조, 캐싱
- **부하 테스트**: JMeter, Locust, k6

**성능 메트릭**:
- 성능 개선: 평균 45%
- 병목 발견율: 91%
- 최적화 성공률: 87%

---

### 19. accessibility-expert

**♿ 역할**: 접근성 준수 전문가

**책임**:
- WCAG 2.1 준수 보장
- 접근성 테스트 및 개선
- Inclusive Design 구현
- 스크린 리더 최적화

**사용 시점**:
```bash
"접근성 검토" → accessibility-expert 활성화
"WCAG 준수" → accessibility-expert 검증
```

**주요 능력**:
- **WCAG 검증**: Level A, AA, AAA 기준
- **스크린 리더**: NVDA, JAWS, VoiceOver 테스트
- **키보드 탐색**: Tab 순서, 포커스 관리

**성능 메트릭**:
- 접근성 점수: 97%
- WCAG 준수율: 99%
- 사용성 개선: 35%

---

## 에이전트 협업 패턴

### Pattern 1: Sequential Delegation (순차 위임)

**사용 사례**: 의존성이 있는 작업

```mermaid
flowchart LR
    A[spec-builder] -->|SPEC 완료| B[implementation-planner]
    B -->|계획 완료| C[tdd-implementer]
    C -->|구현 완료| D[test-engineer]
    D -->|테스트 완료| E[doc-syncer]
    E -->|문서 완료| F[git-manager]
```

**예제**:
```bash
/alfred:1-plan "User authentication"
→ spec-builder: SPEC 작성
→ implementation-planner: 실행 계획 수립
→ tdd-implementer: TDD 구현
→ test-engineer: 테스트 검증
→ doc-syncer: 문서 생성
→ git-manager: PR 생성
```

---

### Pattern 2: Parallel Execution (병렬 실행)

**사용 사례**: 독립적인 작업

```mermaid
flowchart TD
    Input["복잡한 작업 입력"]

    subgraph "병렬 에이전트 실행"
        A1["backend-expert<br/>API 설계"]
        A2["database-expert<br/>스키마 설계"]
        A3["security-expert<br/>보안 검토"]
        A4["performance-engineer<br/>성능 분석"]
    end

    Coordination["Alfred 조율<br/>결과 통합"]

    Input --> A1
    Input --> A2
    Input --> A3
    Input --> A4

    A1 --> Coordination
    A2 --> Coordination
    A3 --> Coordination
    A4 --> Coordination
```

**예제**:
```bash
"대규모 트래픽 처리 아키텍처 설계"
→ backend-expert: 서버 아키텍처 설계 (병렬)
→ database-expert: 데이터베이스 샤딩 전략 (병렬)
→ security-expert: 보안 아키텍처 검토 (병렬)
→ performance-engineer: 성능 요구사항 분석 (병렬)
→ Alfred: 결과 통합 및 최종 권장안 제시
```

---

### Pattern 3: Hierarchical Coordination (계층적 조정)

**사용 사례**: 복잡한 다단계 작업

```mermaid
flowchart TD
    Alfred["🎩 Alfred<br/>전체 조율"]

    subgraph "Phase 1: Planning"
        P1[spec-builder]
        P2[implementation-planner]
    end

    subgraph "Phase 2: Implementation"
        I1[backend-expert]
        I2[frontend-expert]
        I3[database-expert]
    end

    subgraph "Phase 3: Quality"
        Q1[test-engineer]
        Q2[qa-validator]
        Q3[security-expert]
    end

    subgraph "Phase 4: Deployment"
        D1[doc-syncer]
        D2[git-manager]
        D3[devops-expert]
    end

    Alfred --> P1
    Alfred --> P2
    P1 --> I1
    P2 --> I2
    P2 --> I3
    I1 --> Q1
    I2 --> Q2
    I3 --> Q3
    Q1 --> D1
    Q2 --> D2
    Q3 --> D3
```

**예제**:
```bash
"전체 인증 시스템 구축"
→ Phase 1 (Planning):
  - spec-builder: 인증 SPEC 작성
  - implementation-planner: 구현 계획 수립

→ Phase 2 (Implementation):
  - backend-expert: API 구현
  - frontend-expert: 로그인 UI 구현
  - database-expert: 사용자 스키마 설계

→ Phase 3 (Quality):
  - test-engineer: 통합 테스트
  - qa-validator: 품질 게이트 검증
  - security-expert: 보안 취약점 검토

→ Phase 4 (Deployment):
  - doc-syncer: API 문서 생성
  - git-manager: PR 생성
  - devops-expert: CI/CD 파이프라인 구축
```

---

## 자동 에이전트 선택 로직

### Alfred의 지능형 위임 알고리즘

```python
class AlfredOrchestrator:
    def select_agents(self, user_request: str, context: dict) -> list[Agent]:
        """
        사용자 요청과 컨텍스트 기반 최적 에이전트 선택
        """
        # 1. 도메인 분석
        domains = self.analyze_domains(user_request)

        # 2. 복잡도 평가
        complexity = self.assess_complexity(user_request, context)

        # 3. 작업 유형 판단
        task_type = self.classify_task_type(user_request)

        # 4. 에이전트 선택
        agents = []

        if task_type == "planning":
            agents.append(self.get_agent("spec-builder"))
            if complexity == "high":
                agents.append(self.get_agent("implementation-planner"))

        elif task_type == "implementation":
            agents.append(self.get_agent("tdd-implementer"))
            for domain in domains:
                agents.append(self.get_domain_expert(domain))

        elif task_type == "quality":
            agents.append(self.get_agent("test-engineer"))
            agents.append(self.get_agent("qa-validator"))

        # 5. 실행 전략 결정
        strategy = self.determine_execution_strategy(agents, complexity)

        return agents, strategy

    def analyze_domains(self, user_request: str) -> list[str]:
        """
        도메인 키워드 분석
        """
        domains = []

        keywords = {
            "backend": ["API", "server", "backend", "service"],
            "frontend": ["UI", "component", "frontend", "React"],
            "database": ["database", "query", "schema", "SQL"],
            "security": ["security", "auth", "encryption", "vulnerability"],
            "devops": ["deploy", "CI/CD", "docker", "kubernetes"]
        }

        for domain, keywords_list in keywords.items():
            if any(kw.lower() in user_request.lower() for kw in keywords_list):
                domains.append(domain)

        return domains
```

---

## Performance Metrics

### 전체 에이전트 성능

| 메트릭 | 평균 | 최고 | 최저 |
|--------|------|------|------|
| **작업 완료율** | 96.3% | 100% | 89% |
| **평균 응답 시간** | 12초 | 5초 | 45초 |
| **사용자 만족도** | 92.7% | 98% | 87% |
| **품질 점수** | 94.1% | 99% | 88% |

### 카테고리별 성능

| 카테고리 | 완료율 | 응답 시간 | 만족도 |
|----------|--------|----------|--------|
| **Core Planning** | 98.2% | 8초 | 95.3% |
| **Development** | 96.1% | 15초 | 93.2% |
| **Quality Assurance** | 97.5% | 10초 | 94.8% |
| **Documentation** | 99.0% | 6초 | 92.1% |
| **Domain Experts** | 94.3% | 18초 | 91.5% |

---

## Best Practices

### 1. 에이전트 활용 권장사항

**DO**:
✅ Alfred에게 명확한 작업 설명
✅ 복잡한 작업은 단계별 분해
✅ 에이전트 피드백 적극 활용
✅ 병렬 작업 기회 활용

**DON'T**:
❌ 에이전트 직접 호출 (Alfred 통해서만)
❌ 너무 모호한 요청
❌ 에이전트 결과 무시
❌ 품질 게이트 우회

### 2. 협업 최적화

**순차 작업이 필요한 경우**:
- 요구사항 정의 → 구현 → 테스트 → 문서화

**병렬 작업이 가능한 경우**:
- 백엔드 API + 프론트엔드 UI
- 데이터베이스 스키마 + 보안 검토

### 3. 품질 보장

**자동 검증 활용**:
```bash
/alfred:2-run AUTH-001
→ test-engineer: 테스트 커버리지 96%+ 보장
→ qa-validator: TRUST 5 원칙 자동 검증
→ trust-checker: 규정 준수 확인
```

---

## 다음 단계

### 관련 문서

- **[Expert Delegation System](/ko/alfred/expert-delegation-system)**: v0.23.0 전문가 위임 시스템
- **[Alfred Workflow](/ko/alfred/workflow)**: 4단계 워크플로우 가이드
- **[Commands Guide](/ko/alfred/commands)**: Alfred 명령어 완전 가이드
- **[Skills Reference](/ko/skills)**: 105개 스킬 레퍼런스

### 실습 가이드

1. **에이전트 활용 시작**: [Getting Started](/ko/getting-started)
2. **협업 패턴 적용**: [Advanced Patterns](/ko/guides/agent-patterns)
3. **품질 게이트 설정**: [Quality Assurance](/ko/guides/quality-gates)

---

## 요약

Alfred의 19명 팀원 에이전트는:

1. **전문성**: 각 도메인의 깊은 전문 지식
2. **협업**: 3가지 협업 패턴 (순차, 병렬, 계층)
3. **자동화**: 지능형 에이전트 선택 및 조율
4. **품질**: 96.3% 작업 완료율, 92.7% 사용자 만족도
5. **학습**: 지속적인 패턴 학습 및 적응

**결과**: 개발자는 비즈니스 로직에 집중, Alfred 팀이 기술 복잡성 해결

---

**Made with ❤️ by MoAI Team**
