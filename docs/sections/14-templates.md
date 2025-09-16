# MoAI-ADK 템플릿 시스템

## 📚 템플릿 예시

### SPEC Template (spec.template.md)
```markdown
# SPEC-$SPEC_ID: $SPEC_TITLE

## User Stories

### US-001: [Primary User Story]
**As a** [user type]
**I want** [functionality]
**So that** [benefit]

**Acceptance Criteria:**
- [ ] [Specific testable criterion]
- [ ] [Another criterion]

## Functional Requirements

### FR-001: [Requirement Name]
**Given** [context]
**When** [action]
**Then** [expected result]

## Non-Functional Requirements

### Performance
- Response time < 200ms
- Concurrent users: 1000+

### Security
- Authentication required
- Data encryption at rest

## Key Entities

### Entity: User
```typescript
interface User {
  id: string;
  email: string;
  role: UserRole;
  createdAt: Date;
}
```

## Review Checklist
- [ ] All requirements use EARS format
- [ ] Acceptance criteria are testable
- [ ] Non-functional requirements specified
- [ ] @TAG references complete
```

### Plan Template (plan-template.md)
```markdown
# Implementation Plan for SPEC-$SPEC_ID

## Technical Context

### Current State
- Existing systems: [description]
- Dependencies: [list]
- Constraints: [limitations]

### Target State
- Proposed solution: [architecture]
- Benefits: [advantages]
- Risks: [potential issues]

## Constitution Check

### Simplicity ✅
- Components: [count] ≤ 10
- Dependencies: [depth] ≤ 3
- Configuration files: [count] ≤ 5

### Architecture ✅
- Follows established patterns
- Proper separation of concerns
- Scalable design

### Testing (NON-NEGOTIABLE) ✅
- TDD approach confirmed
- Test coverage target: 80%+
- Test strategy defined

### Observability ✅
- Logging strategy: [approach]
- Monitoring plan: [metrics]
- Error tracking: [system]

### Versioning ✅
- Git workflow: [strategy]
- Release process: [plan]
- Documentation updates: [approach]

## Implementation Phases

### Phase 0: Research
- [ ] Technology evaluation
- [ ] Proof of concept
- [ ] Risk assessment

### Phase 1: Design
- [ ] API design
- [ ] Database schema
- [ ] Architecture diagrams

### Phase 2: Implementation
- [ ] Core functionality
- [ ] Test suite
- [ ] Documentation

## Dependencies
- [ ] External APIs
- [ ] Third-party libraries
- [ ] Infrastructure requirements
```

### Task Template (tasks.template.md)
```markdown
# TDD Tasks for SPEC-$SPEC_ID

## Task Decomposition Strategy

### Red-Green-Refactor Approach
All tasks follow the TDD cycle:
1. **RED**: Write failing test
2. **GREEN**: Implement minimal code to pass
3. **REFACTOR**: Improve code quality

## Sprint Planning

### Sprint 1: Foundation (Day 1)
- [T001] Setup project structure
- [T002] Configure testing framework
- [T003] Create base models

### Sprint 2: Core Features (Day 2-3)
- [T004] [P] Implement authentication
- [T005] [P] User management API
- [T006] Data validation layer

### Sprint 3: Integration (Day 4)
- [T007] API integration tests
- [T008] End-to-end scenarios
- [T009] Performance testing

### Sprint 4: Polish (Day 5)
- [T010] Error handling
- [T011] Documentation
- [T012] Deployment preparation

## Task Details

### [T001] Setup Project Structure
**Type**: Foundation
**Effort**: 2 hours
**Dependencies**: None

**Test Cases:**
- [ ] Project builds successfully
- [ ] All dependencies installed
- [ ] Basic CI/CD pipeline works

**Implementation:**
- Create directory structure
- Setup package.json/pyproject.toml
- Configure build tools

### [T002] [P] User Authentication
**Type**: Core Feature
**Effort**: 8 hours
**Dependencies**: T001
**Parallel**: Can run with T003

**Test Cases:**
- [ ] User can register with valid email
- [ ] User can login with correct credentials
- [ ] Invalid credentials are rejected
- [ ] JWT tokens are properly generated

**Implementation:**
- Auth service interface
- Password hashing
- JWT token generation
- Login/register endpoints
```

## 동적 템플릿 생성

### TemplateEngine 사용법
```python
# 템플릿 엔진 사용 예시
engine = TemplateEngine()

# SPEC 템플릿 생성
spec_content = engine.generate_spec_template(
    spec_id="001",
    spec_title="User Authentication",
    variables={
        "PROJECT_NAME": "MyApp",
        "AUTHOR": "Development Team"
    }
)

# 동적 변수 주입
variables = {
    '$MOAI_VERSION': '0.1.16',
    '$PROJECT_NAME': 'My Project',
    '$SPEC_ID': '001',
    '$SPEC_TITLE': 'User Authentication'
}
```

## 커스텀 템플릿 생성

### 새 템플릿 추가
```bash
# 새 템플릿 파일 생성
.moai/_templates/custom/my-template.template.md

# 템플릿 등록
moai template register my-template

# 템플릿 사용
/moai:2-spec my-feature --template my-template
```

### 팀별 템플릿 관리
```json
// .moai/config.json
{
  "templates": {
    "spec_template": "custom-spec",
    "task_template": "agile-tasks",
    "custom_variables": {
      "COMPANY_NAME": "TechCorp",
      "PROJECT_PREFIX": "TC"
    }
  }
}
```

템플릿 시스템은 **일관된 문서 구조**와 **동적 콘텐츠 생성**을 통해 효율적인 개발 문서화를 지원합니다.

## 템플릿 탐색 순서 및 폴백 (vNext)

- TemplateEngine 탐색 순서:
  1) 프로젝트 로컬: `.moai/_templates/`
  2) 패키지 내장: `moai_adk.resources/templates/.moai/_templates`

- 설치 모드와 동작:
  - `templates.mode = copy`(기본): 템플릿이 프로젝트로 복사됩니다.
  - `templates.mode = package`: 템플릿 복사를 생략하고, 없으면 패키지 템플릿으로 자동 폴백합니다.

- 오버라이드 원칙:
- 동일 경로/파일명일 때 로컬 템플릿이 항상 우선합니다.
- 팀 공용 템플릿은 최소화하고 필요한 파일만 `.moai/_templates/`에 추가하세요.

예시
```bash
# 패키지 폴백 사용(로컬 템플릿 없음)
# → 패키지 템플릿이 적용됨

# 프로젝트 오버라이드 추가
mkdir -p .moai/_templates/specs
printf "# LOCAL: $SPEC_NAME ($SPEC_ID)\n" > .moai/_templates/specs/spec.template.md
# → 동일 명령 실행 시 로컬 템플릿이 우선 적용
```
