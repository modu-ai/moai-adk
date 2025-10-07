---
id: PRODUCT-001
version: 0.1.0
status: active
created: 2025-10-01
updated: 2025-10-01
authors: ["@project-owner"]
---

# {{PROJECT_NAME}} Product Definition

## HISTORY

### v0.1.0 (2025-10-01)
- **INITIAL**: 프로젝트 제품 정의 문서 작성
- **AUTHOR**: @project-owner
- **SECTIONS**: Mission, User, Problem, Strategy, Success, Legacy

---

## @DOC:MISSION-001 핵심 미션

**[프로젝트의 핵심 미션과 목표를 정의하세요]**

### 핵심 가치 제안

[이 프로젝트가 제공하는 핵심 가치를 설명하세요]

## @SPEC:USER-001 주요 사용자층

### 1차 사용자
- **대상**: [주요 사용자층을 정의하세요]
- **핵심 니즈**: [사용자가 해결하고자 하는 문제]
- **핵심 시나리오**: [주요 사용 시나리오를 설명하세요]

### 2차 사용자 (선택사항)
- **대상**: [추가 사용자층이 있다면 정의하세요]
- **핵심 니즈**: [추가 사용자의 요구사항]

## @SPEC:PROBLEM-001 해결하는 핵심 문제

### 우선순위 높음
1. [해결하려는 주요 문제 1]
2. [해결하려는 주요 문제 2]
3. [해결하려는 주요 문제 3]

### 우선순위 중간
- [중요도가 중간인 문제들]

### 현재 실패 사례들
- [기존 솔루션의 한계나 실패 사례들]

## @DOC:STRATEGY-001 차별점 및 강점

### 경쟁 솔루션 대비 강점
1. [주요 차별점 1]
   - **발휘 시나리오**: [어떤 상황에서 이 강점이 드러나는지]

2. [주요 차별점 2]
   - **발휘 시나리오**: [구체적인 활용 시나리오]

## @SPEC:SUCCESS-001 성공 지표

### 즉시 측정 가능한 KPI
1. [측정 지표 1]
   - **베이스라인**: [목표값과 측정 방법]

2. [측정 지표 2]
   - **베이스라인**: [목표값과 측정 방법]

### 측정 주기
- **일간**: [일단위로 측정할 지표]
- **주간**: [주단위로 측정할 지표]
- **월간**: [월단위로 측정할 지표]

## Legacy Context

### 기존 자산 요약
- [활용할 기존 자산이나 리소스]
- [참고할 기존 프로젝트나 경험]

## TODO:SPEC-BACKLOG-001 다음 단계 SPEC 후보

1. **SPEC-001**: [첫 번째 구현할 기능]
2. **SPEC-002**: [두 번째 구현할 기능]
3. **SPEC-003**: [세 번째 구현할 기능]

## EARS 요구사항 작성 가이드

### EARS (Easy Approach to Requirements Syntax)

SPEC 작성 시 다음 EARS 구문을 활용하여 체계적인 요구사항을 작성하세요:

#### EARS 구문 형식
1. **Ubiquitous Requirements**: 시스템은 [기능]을 제공해야 한다
2. **Event-driven Requirements**: WHEN [조건]이면, 시스템은 [동작]해야 한다
3. **State-driven Requirements**: WHILE [상태]일 때, 시스템은 [동작]해야 한다
4. **Optional Features**: WHERE [조건]이면, 시스템은 [동작]할 수 있다
5. **Constraints**: IF [조건]이면, 시스템은 [제약]해야 한다

#### 적용 예시
```markdown
### Ubiquitous Requirements (기본 기능)
- 시스템은 사용자 관리 기능을 제공해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN 사용자가 가입하면, 시스템은 환영 이메일을 발송해야 한다

### State-driven Requirements (상태 기반)
- WHILE 사용자가 로그인된 상태일 때, 시스템은 개인화된 대시보드를 표시해야 한다

### Optional Features (선택적 기능)
- WHERE 프리미엄 계정이면, 시스템은 고급 기능을 제공할 수 있다

### Constraints (제약사항)
- IF 계정이 잠긴 상태이면, 시스템은 로그인을 거부해야 한다
```

---

_이 문서는 `/alfred:1-spec` 실행 시 SPEC 생성의 기준이 됩니다._