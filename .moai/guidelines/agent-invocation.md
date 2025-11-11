# Agent Invocation Standards
**MoAI-ADK Agent Orchestration Guidelines**
**Version**: 1.0.0
**Based on**: [Official Claude Code Sub-Agents Documentation](https://code.claude.com/docs/en/sub-agents)
**Date**: 2025-11-12

---

## 📋 목차

1. [핵심 원칙](#핵심-원칙)
2. [올바른 호출 패턴](#올바른-호출-패턴)
3. [잘못된 호출 패턴](#잘못된-호출-패턴)
4. [Resume 메커니즘](#resume-메커니즘)
5. [Context 전달 방법](#context-전달-방법)
6. [Agent 간 협업 패턴](#agent-간-협업-패턴)
7. [실전 예제](#실전-예제)

---

## 핵심 원칙

### 공식 문서에서 명시한 규칙

> "Each subagent operates in its own context, preventing pollution of the main conversation"

**이것이 의미하는 것**:

1. **독립된 context window**: 각 sub-agent는 격리된 환경에서 실행됩니다.
2. **직접 통신 불가**: Agent A는 Agent B와 직접 대화할 수 없습니다.
3. **Main thread 중개**: 모든 정보는 Alfred (main conversation)를 통해 전달됩니다.
4. **단일 계층 구조**: Sub-agent는 다른 sub-agent를 호출할 수 없습니다.

### MoAI-ADK 적용

```
✅ ALLOWED:
Commands → Task(subagent_type="agent-name")  // Alfred가 agent 호출
Agents → Skill("skill-name")                  // Agent가 skill 참조
Agents → Return results to Alfred             // Alfred로 결과 반환

❌ FORBIDDEN:
Agents → Task(subagent_type="other-agent")   // Agent가 다른 agent 호출
Agents → File-based communication             // .moai/plan.json 같은 파일 공유
Agents → Direct agent-to-agent calls          // 직접 통신
```

---

## 올바른 호출 패턴

### Pattern 1: Alfred가 Sequential Chain 조율

**시나리오**: `/alfred:1-plan` 명령이 SPEC 생성 및 구현 계획을 수행

```python
# Alfred (Main Orchestrator)의 실행 흐름

# STEP 1: spec-builder 호출
spec_result = Task(
    subagent_type="spec-builder",
    prompt="""You are the spec-builder agent.

    User request: "사용자 인증 기능 SPEC 생성"

    Create SPEC documents in Korean:
    - .moai/specs/SPEC-AUTH-001/spec.md
    - .moai/specs/SPEC-AUTH-001/plan.md
    - .moai/specs/SPEC-AUTH-001/acceptance.md

    Use MultiEdit for simultaneous creation.
    """
)

# STEP 2: Alfred가 결과를 main context에 저장
alfred_context = {
    "spec_builder_result": spec_result,
    "spec_id": "SPEC-AUTH-001",
    "spec_files": [
        ".moai/specs/SPEC-AUTH-001/spec.md",
        ".moai/specs/SPEC-AUTH-001/plan.md",
        ".moai/specs/SPEC-AUTH-001/acceptance.md"
    ]
}

# STEP 3: implementation-planner 호출 (spec-builder 결과 전달)
plan_result = Task(
    subagent_type="implementation-planner",
    prompt=f"""You are the implementation-planner agent.

    SPEC-builder has created SPEC-AUTH-001.
    SPEC location: .moai/specs/SPEC-AUTH-001/spec.md

    Read the SPEC and create detailed implementation plan:
    - TAG chain breakdown
    - Library dependencies
    - Implementation sequence
    - Risk assessment

    SPEC summary from previous agent:
    {spec_result['summary']}
    """
)

# STEP 4: Alfred가 다시 결과 저장
alfred_context["implementation_plan"] = plan_result

# STEP 5: 사용자에게 보고
return {
    "status": "success",
    "spec_created": alfred_context["spec_id"],
    "plan_ready": True,
    "next_command": "/alfred:2-run SPEC-AUTH-001"
}
```

**핵심 포인트**:

✅ Alfred가 모든 agent 호출을 조율합니다
✅ 각 agent 결과는 Alfred context에 저장됩니다
✅ 다음 agent는 Alfred를 통해 이전 결과를 받습니다
❌ spec-builder가 implementation-planner를 직접 호출하지 않습니다

---

### Pattern 2: Resume를 사용한 연속 작업

**시나리오**: tdd-implementer가 여러 TAG를 순차적으로 구현

```python
# Alfred의 실행 흐름

# STEP 1: TAG-001 구현 시작
result_tag_001 = Task(
    subagent_type="tdd-implementer",
    prompt="""You are the tdd-implementer agent.

    Implement TAG-001: User registration endpoint
    Follow RED-GREEN-REFACTOR cycle.
    """
)

# agentId 저장 (예: "tdd-001-abc123")
agent_id_tdd = result_tag_001["agent_id"]

# STEP 2: 같은 agent로 TAG-002 계속 구현 (resume 사용)
result_tag_002 = Task(
    subagent_type="tdd-implementer",
    prompt="""Continue implementing TAG-002: User login endpoint.

    Previous TAG-001 implementation is complete.
    Build on existing authentication infrastructure.
    Follow same TDD principles.
    """,
    resume=agent_id_tdd  # 🔑 이전 conversation history 상속
)

# STEP 3: REFACTOR 단계도 resume로 연속
result_refactor = Task(
    subagent_type="tdd-implementer",
    prompt="""REFACTOR phase: Extract common auth utilities.

    Review TAG-001 and TAG-002 implementations.
    Eliminate code duplication.
    Maintain 100% test coverage.
    """,
    resume=agent_id_tdd  # 🔑 전체 구현 history 보유
)
```

**Resume의 이점**:

✅ Full conversation history 유지 (TAG-001, TAG-002, REFACTOR 모두 기억)
✅ Context 연속성 보장 (코드 스타일, 구조 일관성)
✅ 중복 설명 불필요 (이전 결정 사항 재설명 안 함)

**Resume 사용 조건**:

- 같은 agent를 연속 호출
- 이전 작업을 이어서 진행
- Context 연속성이 필요한 경우

---

### Pattern 3: 병렬 분석 후 결과 통합

**시나리오**: SPEC 검토를 위해 여러 전문가 동시 투입

```python
# Alfred의 병렬 실행 및 통합

import asyncio

# STEP 1: 전문가 agent들 병렬 호출 (각자 독립 agentId)
backend_analysis = Task(
    subagent_type="backend-expert",
    prompt="""Review SPEC-AUTH-001 for backend architecture.

    Focus on:
    - API design patterns
    - Database schema
    - Authentication strategy
    - Security concerns
    """
)

frontend_analysis = Task(
    subagent_type="frontend-expert",
    prompt="""Review SPEC-AUTH-001 for frontend requirements.

    Focus on:
    - UI/UX considerations
    - State management
    - Form validation
    - Client-side security
    """
)

security_analysis = Task(
    subagent_type="security-expert",
    prompt="""Review SPEC-AUTH-001 for security vulnerabilities.

    Focus on:
    - OWASP top 10 compliance
    - Password policy
    - Token management
    - Rate limiting
    """
)

# STEP 2: Alfred가 모든 결과 수집
analysis_results = {
    "backend": backend_analysis,
    "frontend": frontend_analysis,
    "security": security_analysis
}

# STEP 3: Alfred가 통합 보고서 생성
integrated_report = {
    "spec_id": "SPEC-AUTH-001",
    "expert_reviews": analysis_results,
    "action_items": extract_action_items(analysis_results),
    "risks": consolidate_risks(analysis_results),
    "recommendations": merge_recommendations(analysis_results)
}

# STEP 4: 사용자에게 통합 보고
return integrated_report
```

**병렬 실행 규칙**:

✅ 각 전문가는 독립 agentId로 실행
✅ Alfred가 모든 결과를 수집하고 통합
❌ backend-expert가 frontend-expert 결과를 직접 읽지 않음

---

## 잘못된 호출 패턴

### ❌ Anti-Pattern 1: Agent가 다른 Agent 직접 호출

```python
# ❌ WRONG: tdd-implementer가 quality-gate를 직접 호출
# File: .claude/agents/alfred/tdd-implementer.md (잘못된 구현 예시)

# STEP 5: 구현 완료 후 검증 (❌ 금지된 패턴)
quality_result = Task(
    subagent_type="quality-gate",
    prompt="Verify my implementation"
)
# ❌ ERROR: Sub-agents cannot spawn other sub-agents
```

**왜 안 되는가?**

- 공식 문서: "Sub-agents CANNOT spawn other sub-agents"
- 단일 계층 구조만 허용 (Commands → Agents, NOT Agents → Agents)

**올바른 방법**:

```python
# ✅ CORRECT: Alfred가 순차 조율

# Alfred의 /alfred:2-run 명령
implementation_result = Task(subagent_type="tdd-implementer", ...)
quality_result = Task(subagent_type="quality-gate", ...)
commit_result = Task(subagent_type="git-manager", ...)
```

---

### ❌ Anti-Pattern 2: 파일 기반 Agent 간 통신

```python
# ❌ WRONG: implementation-planner가 파일에 결과 저장
# File: .claude/agents/alfred/implementation-planner.md (잘못된 구현)

# STEP 5: 계획 저장
Write(".moai/temp/plan.json", json.dumps(plan_data))
print("tdd-implementer는 .moai/temp/plan.json을 읽으세요")
# ❌ ERROR: File-based inter-agent communication is forbidden
```

**왜 안 되는가?**

- Agent는 독립된 context에서 실행
- 파일 공유는 암묵적 의존성 생성 (추적 불가)
- Alfred의 조율 역할 우회

**올바른 방법**:

```python
# ✅ CORRECT: Alfred가 명시적으로 전달

# implementation-planner는 결과를 반환만
plan_result = {
    "tag_chain": [...],
    "dependencies": [...],
    "implementation_sequence": [...]
}
return plan_result

# Alfred가 tdd-implementer에게 전달
implementation_result = Task(
    subagent_type="tdd-implementer",
    prompt=f"""
    Implementation plan from planner:
    {json.dumps(plan_result, indent=2)}

    Execute TAG-001 first...
    """
)
```

---

### ❌ Anti-Pattern 3: Agent 간 직접 메시지 전달 시도

```python
# ❌ WRONG: spec-builder가 doc-syncer에게 직접 메시지
# (이런 코드를 작성하려 시도하면 오류 발생)

# SPEC 생성 완료 후
send_message_to_agent("doc-syncer", "SPEC-AUTH-001 created, please update docs")
# ❌ ERROR: No such function exists
```

**올바른 방법**:

```python
# ✅ CORRECT: Alfred가 workflow 조율

# /alfred:1-plan 명령에서 Alfred가:
spec_result = Task(subagent_type="spec-builder", ...)

# Alfred가 판단: "문서 업데이트 필요?"
if spec_result["requires_doc_update"]:
    doc_result = Task(
        subagent_type="doc-syncer",
        prompt=f"Update docs for {spec_result['spec_id']}"
    )
```

---

## Resume 메커니즘

### Resume란?

공식 문서:

> "Resume preserves full conversation history"

**기술적 구현**:

```typescript
// 첫 실행
{
  "subagent_type": "tdd-implementer",
  "prompt": "Implement TAG-001"
}
// Returns: { "agent_id": "tdd-abc123", ... }

// Resume (conversation history 상속)
{
  "subagent_type": "tdd-implementer",
  "prompt": "Continue with TAG-002",
  "resume": "tdd-abc123"  // 🔑 이전 모든 대화 내용 로드
}
```

**Session 파일**:

- 위치: `.moai/logs/agent-transcripts/agent-tdd-abc123.jsonl`
- 내용: 전체 conversation history (user prompts + agent responses)

---

### Resume 사용 결정 트리

```
┌─ 같은 agent를 호출하는가?
│   ├─ YES ─┬─ 이전 작업을 계속하는가?
│   │       ├─ YES → resume 사용 (agentId 전달)
│   │       └─ NO → 새 session 시작
│   └─ NO → 새 session 시작 (다른 agent)
│
└─ Context 연속성이 필요한가?
    ├─ YES → resume 사용 (같은 SPEC, 같은 도메인)
    └─ NO → 새 session 시작 (독립 작업)
```

**Resume 사용 예**:

- ✅ tdd-implementer: TAG-001 → TAG-002 → TAG-003 (연속 구현)
- ✅ doc-syncer: product.md → structure.md → tech.md (연속 업데이트)
- ✅ debug-helper: Error 1 분석 → Fix 제안 → Error 2 분석 (디버깅 세션)

**새 session 사용 예**:

- ✅ spec-builder → implementation-planner (다른 agent)
- ✅ tdd-implementer (SPEC-001) → tdd-implementer (SPEC-002) (독립 SPEC)
- ✅ quality-gate 검증 (매번 독립 실행)

---

### Resume ID 추적

**SessionManager 사용** (Python 구현):

```python
from moai_adk.core.session_manager import SessionManager

session_mgr = SessionManager()

# Agent 실행 및 ID 저장
result = Task(subagent_type="tdd-implementer", prompt="Implement TAG-001")
session_mgr.register_agent_result(
    agent_name="tdd-implementer",
    agent_id=result["agent_id"],
    result=result
)

# Resume ID 조회
resume_id = session_mgr.get_resume_id(
    agent_name="tdd-implementer",
    chain_id="SPEC-AUTH-001-implementation"
)

# Resume 실행
result2 = Task(
    subagent_type="tdd-implementer",
    prompt="Continue with TAG-002",
    resume=resume_id
)
```

---

## Context 전달 방법

### Alfred의 Main Context 구조

```python
# Alfred가 유지하는 context (예시)
alfred_context = {
    # Agent 실행 결과
    "agent_results": {
        "spec-builder": {
            "agent_id": "spec-abc123",
            "spec_id": "SPEC-AUTH-001",
            "files_created": [...],
            "status": "success"
        },
        "implementation-planner": {
            "agent_id": "plan-def456",
            "tag_chain": ["TAG-001", "TAG-002", "TAG-003"],
            "dependencies": {...},
            "status": "success"
        }
    },

    # 현재 workflow 상태
    "workflow_state": {
        "current_command": "/alfred:2-run",
        "current_spec": "SPEC-AUTH-001",
        "completed_steps": ["plan", "approve"],
        "current_step": "implement"
    },

    # Session ID 매핑
    "agent_sessions": {
        "tdd-implementer": "tdd-ghi789",
        "quality-gate": "qa-jkl012"
    }
}
```

---

### Context 전달 패턴

#### Pattern A: 요약 전달 (효율성 우선)

```python
# Alfred가 이전 agent 결과를 요약하여 전달
summary = {
    "spec_id": alfred_context["agent_results"]["spec-builder"]["spec_id"],
    "key_requirements": extract_key_points(spec_content),
    "tech_stack": ["FastAPI", "SQLAlchemy", "PostgreSQL"]
}

next_result = Task(
    subagent_type="implementation-planner",
    prompt=f"""
    SPEC summary:
    {json.dumps(summary, indent=2)}

    Create detailed implementation plan...
    """
)
```

**장점**: Token 효율적, 핵심만 전달
**단점**: 세부 정보 손실 가능

---

#### Pattern B: 파일 경로 전달 (완전성 우선)

```python
# Alfred가 파일 위치를 알려주고 agent가 직접 읽음
next_result = Task(
    subagent_type="implementation-planner",
    prompt=f"""
    SPEC file location: .moai/specs/SPEC-AUTH-001/spec.md

    Read the SPEC file and create implementation plan.
    Analyze all requirements thoroughly.
    """
)
```

**장점**: 완전한 정보 접근, 정확성 보장
**단점**: Agent가 파일 읽기 필요 (추가 작업)

---

#### Pattern C: Hybrid (추천)

```python
# Alfred가 요약 + 파일 경로 둘 다 제공
next_result = Task(
    subagent_type="implementation-planner",
    prompt=f"""
    SPEC created: SPEC-AUTH-001
    Location: .moai/specs/SPEC-AUTH-001/spec.md

    Quick summary:
    - Feature: User authentication with JWT
    - Main requirements: Registration, Login, Token refresh
    - Tech stack: FastAPI + PostgreSQL

    Read the full SPEC for detailed requirements.
    Create implementation plan with TAG chain breakdown.
    """
)
```

**장점**: 빠른 이해 (요약) + 정확성 (원본 참조)
**추천 사용처**: 대부분의 agent 간 전달

---

## Agent 간 협업 패턴

### Pattern 1: Linear Chain (선형 연쇄)

```
spec-builder → implementation-planner → tdd-implementer → quality-gate → git-manager
     |               |                       |                 |              |
   SPEC 생성      구현 계획            TDD 구현          품질 검증      Git commit
```

**Alfred 조율 코드**:

```python
# /alfred:2-run SPEC-XXX 명령 실행 흐름

# Step 1: 구현 계획 (이미 /alfred:1-plan에서 완료된 경우 건너뜀)
plan = Task(subagent_type="implementation-planner", ...)
alfred_context["plan"] = plan

# Step 2: TDD 구현
implementation = Task(
    subagent_type="tdd-implementer",
    prompt=f"Implement {plan['tag_chain']} following TDD cycle",
    resume=None  # 새 작업
)
alfred_context["implementation"] = implementation

# Step 3: 품질 검증
quality = Task(
    subagent_type="quality-gate",
    prompt=f"Verify implementation of {plan['spec_id']}",
    resume=None  # 독립 검증
)

if quality["status"] != "success":
    # 실패 시 사용자에게 알리고 중단
    return {"error": "Quality gate failed", "issues": quality["issues"]}

# Step 4: Git commit
commit = Task(
    subagent_type="git-manager",
    prompt=f"Create TDD commit for {plan['spec_id']}"
)

return {"status": "success", "commit_sha": commit["sha"]}
```

---

### Pattern 2: Consultation (전문가 자문)

```
        ┌─ backend-expert
        │
Alfred ─┼─ frontend-expert  ─→ Alfred (통합) ─→ spec-builder (수정)
        │
        └─ security-expert
```

**Alfred 조율 코드**:

```python
# SPEC 검토 workflow

# Step 1: 전문가 의견 수집 (병렬 가능)
backend_review = Task(
    subagent_type="backend-expert",
    prompt="Review SPEC-AUTH-001 backend architecture"
)

security_review = Task(
    subagent_type="security-expert",
    prompt="Review SPEC-AUTH-001 security concerns"
)

# Step 2: Alfred가 피드백 통합
combined_feedback = {
    "backend": backend_review["recommendations"],
    "security": security_review["vulnerabilities"],
    "action_required": True
}

# Step 3: spec-builder에게 수정 요청
updated_spec = Task(
    subagent_type="spec-builder",
    prompt=f"""
    Expert feedback received:
    {json.dumps(combined_feedback, indent=2)}

    Update SPEC-AUTH-001 to address concerns.
    """,
    resume=original_spec_agent_id  # 원래 SPEC session 계속
)
```

---

### Pattern 3: Iterative Refinement (반복 개선)

```
tdd-implementer → quality-gate → [FAIL] → debug-helper → tdd-implementer (resume)
                       ↓
                    [PASS]
                       ↓
                  git-manager
```

**Alfred 조율 코드**:

```python
# TDD cycle with quality feedback loop

max_iterations = 3
for iteration in range(max_iterations):
    # 구현 시도
    impl_result = Task(
        subagent_type="tdd-implementer",
        prompt=f"Implement TAG-001 (iteration {iteration+1})",
        resume=tdd_agent_id if iteration > 0 else None
    )

    # 품질 검증
    qa_result = Task(
        subagent_type="quality-gate",
        prompt="Verify implementation"
    )

    if qa_result["status"] == "success":
        # 성공 → commit
        commit_result = Task(
            subagent_type="git-manager",
            prompt="Create commit"
        )
        break
    else:
        # 실패 → 디버깅 도움
        debug_result = Task(
            subagent_type="debug-helper",
            prompt=f"Analyze failures: {qa_result['issues']}"
        )

        # 다음 iteration에서 debug 결과 활용
        tdd_agent_id = impl_result["agent_id"]
```

---

## 실전 예제

### 예제 1: `/alfred:1-plan` 명령 전체 흐름

**사용자 요청**: `/alfred:1-plan "사용자 인증"`

```python
# Alfred의 /alfred:1-plan 명령 handler

def execute_alfred_plan(feature_name: str):
    """
    /alfred:1-plan 명령 실행

    Workflow:
    1. spec-builder: SPEC 문서 생성
    2. implementation-planner: 구현 계획 수립
    3. (Optional) 전문가 자문
    4. 사용자 승인 대기
    """

    # ===== STEP 1: SPEC 생성 =====
    print("📋 SPEC 생성 중...")

    spec_result = Task(
        subagent_type="spec-builder",
        prompt=f"""You are the spec-builder agent.

        User request: "{feature_name}"

        Create SPEC documents in Korean:
        1. Analyze requirements
        2. Create SPEC-XXX directory with proper naming
        3. Generate spec.md, plan.md, acceptance.md using MultiEdit

        Follow EARS format and MoAI-ADK standards.
        """
    )

    # Alfred context에 저장
    context = {
        "spec_id": spec_result["spec_id"],
        "spec_agent_id": spec_result["agent_id"],
        "spec_files": spec_result["files_created"]
    }

    print(f"✅ SPEC 생성 완료: {context['spec_id']}")

    # ===== STEP 2: 구현 계획 =====
    print("🛠️ 구현 계획 수립 중...")

    plan_result = Task(
        subagent_type="implementation-planner",
        prompt=f"""You are the implementation-planner agent.

        SPEC location: .moai/specs/{context['spec_id']}/spec.md

        Create detailed implementation plan:
        1. Read SPEC thoroughly
        2. Break down into TAG chain
        3. Identify library dependencies
        4. Define implementation sequence
        5. Assess risks

        Generate structured plan in Korean.
        """
    )

    context["plan"] = plan_result["plan"]
    context["plan_agent_id"] = plan_result["agent_id"]
    context["tag_chain"] = plan_result["tag_chain"]

    print(f"✅ 구현 계획 완료: {len(context['tag_chain'])} TAGs")

    # ===== STEP 3: 전문가 자문 (선택) =====
    if requires_expert_consultation(plan_result):
        print("🧑‍💼 전문가 자문 수집 중...")

        expert_reviews = {}

        if "backend" in plan_result["domains"]:
            expert_reviews["backend"] = Task(
                subagent_type="backend-expert",
                prompt=f"Review {context['spec_id']} for backend architecture"
            )

        if "security" in plan_result["domains"]:
            expert_reviews["security"] = Task(
                subagent_type="security-expert",
                prompt=f"Review {context['spec_id']} for security vulnerabilities"
            )

        context["expert_reviews"] = expert_reviews
        print(f"✅ {len(expert_reviews)} 전문가 의견 수집 완료")

    # ===== STEP 4: 사용자 승인 =====
    user_approval = AskUserQuestion(
        questions=[{
            "question": "구현 계획을 검토하셨나요? 진행하시겠습니까?",
            "header": "승인 필요",
            "multiSelect": False,
            "options": [
                {"label": "승인 및 구현 시작", "description": "/alfred:2-run 자동 실행"},
                {"label": "계획 수정 필요", "description": "spec-builder resume하여 수정"},
                {"label": "나중에 진행", "description": "현재 세션 종료"}
            ]
        }]
    )

    if user_approval == "승인 및 구현 시작":
        # /alfred:2-run 자동 호출
        return execute_alfred_run(context["spec_id"], context)
    elif user_approval == "계획 수정 필요":
        # spec-builder resume
        revised_spec = Task(
            subagent_type="spec-builder",
            prompt=f"""
            User requested revisions for {context['spec_id']}.

            Expert feedback:
            {json.dumps(context.get('expert_reviews', {}), indent=2)}

            Update SPEC to address concerns.
            """,
            resume=context["spec_agent_id"]  # 🔑 Resume
        )
        return {"status": "revised", "spec_id": revised_spec["spec_id"]}
    else:
        return {"status": "pending", "spec_id": context["spec_id"]}
```

---

### 예제 2: `/alfred:2-run` 명령 TDD 구현

**사용자 요청**: `/alfred:2-run SPEC-AUTH-001`

```python
def execute_alfred_run(spec_id: str, context: dict = None):
    """
    /alfred:2-run 명령 실행

    Workflow:
    1. implementation-planner: 계획 확인 (없으면 생성)
    2. tdd-implementer: TDD cycle 실행
    3. quality-gate: 검증
    4. git-manager: Commit
    5. doc-syncer: 문서 동기화
    """

    # ===== STEP 1: 계획 확인 =====
    if context is None or "plan" not in context:
        print("📋 구현 계획 로딩 중...")
        plan_result = Task(
            subagent_type="implementation-planner",
            prompt=f"Read and analyze SPEC {spec_id}, create implementation plan"
        )
        context = {"plan": plan_result["plan"], "tag_chain": plan_result["tag_chain"]}

    # ===== STEP 2: TDD 구현 =====
    print(f"🔬 TDD 구현 시작: {len(context['tag_chain'])} TAGs")

    tdd_result = Task(
        subagent_type="tdd-implementer",
        prompt=f"""You are the tdd-implementer agent.

        SPEC: {spec_id}
        TAG chain: {context['tag_chain']}

        Execute TDD cycle for all TAGs:
        1. RED: Write failing tests
        2. GREEN: Write minimal passing code
        3. REFACTOR: Improve code quality

        Report progress for each TAG.
        """
    )

    tdd_agent_id = tdd_result["agent_id"]
    context["implementation"] = tdd_result

    print(f"✅ TDD 구현 완료: {tdd_result['tags_completed']}/{len(context['tag_chain'])}")

    # ===== STEP 3: 품질 검증 =====
    print("🔍 품질 검증 중...")

    qa_result = Task(
        subagent_type="quality-gate",
        prompt=f"""You are the quality-gate agent.

        Verify implementation of {spec_id}:
        1. Test coverage (target: 85%)
        2. Code quality (linting, type checking)
        3. TRUST principles compliance
        4. TAG chain integrity

        Provide detailed report.
        """
    )

    if qa_result["status"] != "success":
        print("❌ 품질 검증 실패")

        # 디버깅 지원
        debug_result = Task(
            subagent_type="debug-helper",
            prompt=f"""Analyze quality gate failures:
            {json.dumps(qa_result['issues'], indent=2)}

            Provide root cause analysis and fix recommendations.
            """
        )

        # 사용자 선택
        user_action = AskUserQuestion(
            questions=[{
                "question": "품질 검증 실패. 어떻게 진행하시겠습니까?",
                "header": "실패 처리",
                "multiSelect": False,
                "options": [
                    {"label": "자동 수정 시도", "description": "debug-helper 제안 적용"},
                    {"label": "수동 수정", "description": "직접 코드 수정 후 재검증"},
                    {"label": "중단", "description": "현재 세션 종료"}
                ]
            }]
        )

        if user_action == "자동 수정 시도":
            # tdd-implementer resume하여 수정
            fix_result = Task(
                subagent_type="tdd-implementer",
                prompt=f"""
                Quality gate failed. Apply fixes:
                {debug_result['recommendations']}

                Re-run tests after fixes.
                """,
                resume=tdd_agent_id  # 🔑 Resume로 context 유지
            )

            # 재검증
            qa_result = Task(
                subagent_type="quality-gate",
                prompt=f"Re-verify {spec_id} after fixes"
            )

    if qa_result["status"] == "success":
        print("✅ 품질 검증 통과")

        # ===== STEP 4: Git Commit =====
        print("📝 Git commit 생성 중...")

        commit_result = Task(
            subagent_type="git-manager",
            prompt=f"""You are the git-manager agent.

            Create TDD commit for {spec_id}:
            1. Stage implementation files
            2. Generate commit message (RED/GREEN/REFACTOR)
            3. Include TAG references
            4. Follow conventional commits

            Quality report:
            - Coverage: {qa_result['coverage']}%
            - Tests passed: {qa_result['tests_passed']}/{qa_result['tests_total']}
            """
        )

        print(f"✅ Commit 생성: {commit_result['commit_sha']}")

        # ===== STEP 5: 문서 동기화 =====
        print("📚 문서 동기화 중...")

        doc_result = Task(
            subagent_type="doc-syncer",
            prompt=f"""You are the doc-syncer agent.

            Synchronize documentation for {spec_id}:
            1. Update product.md (new feature added)
            2. Update structure.md (architecture changes)
            3. Update tech.md (new libraries)
            4. Ensure TAG consistency

            Implementation summary:
            {json.dumps(tdd_result['summary'], indent=2)}
            """
        )

        print("✅ 문서 동기화 완료")

        return {
            "status": "success",
            "spec_id": spec_id,
            "commit_sha": commit_result["commit_sha"],
            "coverage": qa_result["coverage"],
            "next_command": "/alfred:3-sync"
        }
    else:
        return {
            "status": "failed",
            "spec_id": spec_id,
            "issues": qa_result["issues"]
        }
```

---

### 예제 3: Resume로 여러 SPEC 연속 처리

**시나리오**: 여러 SPEC을 한 세션에서 순차 생성

```python
def create_multiple_specs(feature_names: list[str]):
    """
    여러 SPEC을 연속으로 생성 (resume 활용)

    Resume 이점:
    - 이전 SPEC 스타일 기억
    - 일관된 구조 유지
    - 중복 질문 없음
    """

    spec_builder_id = None
    created_specs = []

    for i, feature in enumerate(feature_names):
        print(f"📋 SPEC {i+1}/{len(feature_names)}: {feature}")

        spec_result = Task(
            subagent_type="spec-builder",
            prompt=f"""You are the spec-builder agent.

            {"This is SPEC #" + str(i+1) + " in a series." if i > 0 else ""}
            {"Maintain consistent style with previous SPECs." if i > 0 else ""}

            Feature: {feature}

            Create SPEC documents in Korean.
            """,
            resume=spec_builder_id if i > 0 else None  # 🔑 2번째부터 resume
        )

        # 첫 실행 시 agent ID 저장
        if spec_builder_id is None:
            spec_builder_id = spec_result["agent_id"]

        created_specs.append(spec_result["spec_id"])
        print(f"✅ {spec_result['spec_id']} 생성 완료")

    return created_specs
```

**Resume의 효과**:

- ✅ 스타일 일관성: 같은 템플릿, 같은 용어
- ✅ 효율성: "SPEC 작성 방식 알고 있음", 재질문 없음
- ✅ Context 누적: 이전 SPEC과의 관계 파악 가능

---

## 요약 체크리스트

### Alfred (Main Orchestrator)

- [ ] 모든 sub-agent 호출을 Alfred가 조율
- [ ] Agent 결과를 alfred_context에 저장
- [ ] 다음 agent에게 context 명시적 전달
- [ ] agentId 추적 및 resume 관리
- [ ] Workflow 단계별 순차 실행

### Sub-Agents

- [ ] Task() 호출 금지 (다른 agent 스폰 불가)
- [ ] 결과를 Alfred에게 반환
- [ ] 파일 기반 agent 간 통신 금지
- [ ] Skill() 호출로 지식 참조
- [ ] 자신의 도메인 전문성에만 집중

### Resume 사용

- [ ] 같은 agent 연속 호출 시 resume 고려
- [ ] agentId 저장 및 추적
- [ ] Context 연속성이 필요한 경우 resume
- [ ] 독립 작업은 새 session

### Context 전달

- [ ] Alfred가 요약 또는 파일 경로 전달
- [ ] Hybrid 방식 권장 (요약 + 원본 참조)
- [ ] Token 효율성 고려
- [ ] 완전성과 정확성 균형

---

## 참고 자료

- **공식 문서**: [Claude Code Sub-Agents](https://code.claude.com/docs/en/sub-agents)
- **Alfred 설정**: `.moai/config/alfred-orchestration.yaml`
- **SessionManager**: `src/moai_adk/core/session_manager.py`
- **Command 예제**: `.moai/guidelines/command-orchestration-examples.md`
- **MCP 통합**: `.moai/guidelines/mcp-integration-guide.md`

---

**Last Updated**: 2025-11-12
**Version**: 1.0.0
**Maintained by**: MoAI-ADK Team
