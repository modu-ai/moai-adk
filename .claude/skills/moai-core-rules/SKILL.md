---
name: moai-core-rules
description: Alfred SuperAgent의 필수 규칙을 정의합니다. November 2025 enterprise standard 기반.
---

## Quick Reference

Alfred SuperAgent의 의사결정과 실행을 제어하는 핵심 프레임워크입니다.

**핵심 책임**:
1. **3-Layer Architecture 정의**: Commands → Agents → Skills 계층 분리
2. **4-Step Workflow 규칙**: ADAP (Analyze, Design, Assure, Produce) + Intent
3. **Agent-First Paradigm**: 모든 실행 작업을 agents에 위임
4. **Skill 호출 규칙**: 10+ mandatory patterns, invocation syntax
5. **AskUserQuestion 패턴**: 5가지 필수 사용 시나리오
6. **TRUST 5 Quality Gates**: T/R/U/S/T 각각의 검증 기준

**필수 사용 시나리오**:
- Skill() 호출 규칙 검증
- Command vs Agent 역할 분리
- AskUserQuestion 사용 판단
- TRUST 5 준수 확인
- TAG 체인 무결성 검증

---

## Implementation Guide

### Rule 1: 3-Layer Architecture

**계층 구조**:
```
Commands (Orchestration Only)
  ↓ Task(subagent_type="...")
Agents (Domain Expertise)
  ↓ Skill("skill-name")
Skills (Knowledge Capsules)
```

**Commands - Orchestration ONLY**:
```bash
# ✅ CORRECT: Agent에 위임
Task(
  subagent_type="tdd-implementer",
  description="Build and test application",
  prompt="Implement feature with RED-GREEN-REFACTOR cycle"
)

# ❌ WRONG: 직접 작업 실행
python setup.py build
git commit -m "Build"
```

**Agents - Domain Expertise Ownership**:
- ✅ 복잡한 분석 & 추론 (deep reasoning)
- ✅ 계획 수립 (planning)
- ✅ 의사결정 (decision-making)
- ✅ Skill 호출 및 조율
- ✅ 작업 실행 (execution)

**Skills - Knowledge Capsules (Stateless)**:
- ✅ 상태가 없음 (stateless)
- ✅ 재사용 가능 (reusable)
- ✅ Agent에 의해 호출됨
- ✅ 1000줄 이하
- ✅ 단일 주제

### Rule 2: 4-Step Agent-Based Workflow

**Phase Overview**:
```
Phase 0: INTENT
  ├─ User Request
  ├─ Ambiguous? → clarify with AskUserQuestion
  └─ NO → continue

Phase 1: ANALYZE
  ├─ WebSearch, WebFetch
  ├─ Research best practices
  └─ Identify version-specific guidance

Phase 2: DESIGN
  ├─ Latest info based architecture
  ├─ Current version specification
  └─ Official documentation links

Phase 3: ASSURE
  ├─ TRUST 5 Quality Gates
  ├─ TAG integrity validation
  └─ Compliance verification

Phase 4: PRODUCE
  ├─ Skill invocation
  ├─ File generation
  └─ Commit
```

### Rule 3: Agent-First Paradigm

**의무 위임**:
| 작업 | 위임 대상 | 패턴 |
|------|---------|------|
| 계획 수립 | plan-agent | `Task(subagent_type="plan-agent", ...)` |
| 코드 개발 | tdd-implementer | `Task(subagent_type="tdd-implementer", ...)` |
| 테스트 작성 | test-engineer | `Task(subagent_type="test-engineer", ...)` |
| 문서화 | doc-syncer | `Task(subagent_type="doc-syncer", ...)` |
| Git 작업 | git-manager | `Task(subagent_type="git-manager", ...)` |

**금지 사항**:
- ❌ Commands가 직접 bash 실행
- ❌ Commands가 파일 읽기/쓰기
- ❌ Commands가 Git 조작
- ❌ Commands가 코드 분석
- ❌ Commands가 테스트 실행

---

## Advanced Patterns

### Rule 4: 10 Mandatory Skill Invocations

| # | Skill | 용도 | Invocation |
|---|-------|------|-----------|
| 1 | moai-foundation-trust | TRUST 5 검증 | `Skill("moai-foundation-trust")` |
| 2 | moai-foundation-tags | TAG 검증 & 추적 | `Skill("moai-foundation-tags")` |
| 3 | moai-foundation-specs | SPEC 작성 & 검증 | `Skill("moai-foundation-specs")` |
| 4 | moai-foundation-ears | EARS 요구사항 형식 | `Skill("moai-foundation-ears")` |
| 5 | moai-foundation-git | Git 워크플로우 | `Skill("moai-foundation-git")` |

### Rule 5: AskUserQuestion Patterns

**Scenario 1: 기술 스택 모호**:
```
상황: "Python 웹 프레임워크 추천해줄래?"

AskUserQuestion({
  question: "어떤 유형의 애플리케이션?",
  header: "Application Type",
  options: [
    { label: "REST API", description: "High performance APIs" },
    { label: "Web Application", description: "Traditional MVC" },
    { label: "Microservice", description: "Event-driven" }
  ]
})
```

**Scenario 2: 아키텍처 결정**:
```
상황: "데이터베이스 모델을 어떻게 설계?"

AskUserQuestion({
  question: "어떤 데이터 특성?",
  header: "Data Model",
  options: [
    { label: "Relational", description: "Structured, ACID" },
    { label: "Document", description: "Flexible schema" },
    { label: "Graph", description: "Relationships" }
  ]
})
```

### Rule 6: TRUST 5 Quality Gates

**검증 기준**:

**T: Test First (85%+ Coverage)**:
```yaml
requirements:
  coverage: "≥ 85%"
  coverage_tools: ["pytest-cov", "coverage.py"]
  test_types:
    - unit_tests: "각 함수/메서드"
    - integration_tests: "모듈 간 상호작용"
    - edge_cases: "경계값, 에러 처리"
```

**R: Readable (Clean Code)**:
```yaml
requirements:
  code_standards:
    - SOLID 원칙 준수
    - DRY (Don't Repeat Yourself)
    - KISS (Keep It Simple, Stupid)
  
  documentation:
    - Function docstrings
    - Complex logic comments
    - Type hints (Python 3.10+)
```

**U: Unified (Consistent Patterns)**:
```yaml
requirements:
  consistency:
    - Same patterns across codebase
    - No duplicate logic
    - Shared utilities for common tasks
```

**S: Secured (OWASP Top 10)**:
```yaml
requirements:
  security_checks:
    - No hardcoded credentials
    - No SQL injection vectors
    - No XXE vulnerabilities
    - Input validation
    - Output encoding
```
