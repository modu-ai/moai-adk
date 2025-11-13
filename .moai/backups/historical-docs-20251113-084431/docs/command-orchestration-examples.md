# Command Orchestration Examples
**MoAI-ADK Agent Orchestration in Commands**
**Version**: 1.0.0
**Date**: 2025-11-12

---

## 개요

MoAI-ADK의 Commands가 어떻게 여러 agents를 조율하고 Session Manager를 활용하는지 실전 예제를 제공합니다.

**핵심 패턴**:

- Commands는 orchestration만 담당
- 모든 실제 작업은 agents에게 위임
- SessionManager로 agentId 추적 및 resume 관리
- Alfred context에 결과 저장 및 전달

---

## Command 구조 템플릿

### 기본 구조

```python
# /alfred:X-command.py (예시 - 실제는 .md 파일)

from moai_adk.core.session_manager import SessionManager, get_session_manager

def execute_command(args: dict) -> dict:
    """
    Command 실행 진입점

    Args:
        args: 사용자 입력 인자

    Returns:
        실행 결과 딕셔너리
    """
    # Session manager 초기화
    session_mgr = get_session_manager()

    # Alfred context 초기화
    alfred_context = {
        "command": "/alfred:X-command",
        "args": args,
        "agent_results": {},
        "workflow_state": {}
    }

    # STEP 1: 첫 agent 호출
    result1 = invoke_agent_1(session_mgr, alfred_context)

    # STEP 2: 결과 저장 및 다음 agent 호출
    alfred_context["agent_results"]["agent-1"] = result1
    result2 = invoke_agent_2(session_mgr, alfred_context, result1)

    # STEP 3: 최종 결과 반환
    return prepare_final_result(alfred_context)
```

---

## 예제 1: `/alfred:1-plan` - SPEC 생성 및 계획 수립

### Workflow

```
User → /alfred:1-plan "feature" → spec-builder → implementation-planner → User Approval
```

### 완전한 구현

```python
# File: .claude/commands/alfred-1-plan.py (개념적 예시)

from moai_adk.core.session_manager import SessionManager, register_agent, get_resume_id
from typing import Dict, Any, Optional

def execute_alfred_plan(feature_description: str, expert_consultation: bool = False) -> Dict[str, Any]:
    """
    /alfred:1-plan 명령 실행

    Workflow:
    1. spec-builder: SPEC 문서 생성
    2. implementation-planner: 구현 계획 수립
    3. (Optional) Expert consultation
    4. 사용자 승인 대기

    Args:
        feature_description: 기능 설명
        expert_consultation: 전문가 자문 필요 여부

    Returns:
        실행 결과 (SPEC ID, 계획 요약, 다음 단계)
    """
    # ===== 초기화 =====
    session_mgr = SessionManager()

    alfred_context = {
        "command": "/alfred:1-plan",
        "feature": feature_description,
        "agent_results": {},
        "expert_reviews": {},
        "workflow_state": {
            "current_step": "initialize",
            "completed_steps": []
        }
    }

    print(f"🎩 Alfred: Starting /alfred:1-plan for '{feature_description}'")

    # ===== STEP 1: SPEC 생성 (spec-builder) =====
    print("\n📋 STEP 1: Creating SPEC document...")
    alfred_context["workflow_state"]["current_step"] = "spec_creation"

    # Check if we should resume (unlikely for SPEC creation, but possible)
    resume_id = get_resume_id("spec-builder")

    # Invoke spec-builder
    spec_result = Task(
        subagent_type="spec-builder",
        prompt=f"""You are the spec-builder agent.

        User request: "{feature_description}"

        Your tasks:
        1. Analyze requirements and create SPEC in Korean
        2. Generate SPEC-XXX directory with proper naming
        3. Create spec.md, plan.md, acceptance.md using MultiEdit
        4. Follow EARS format and MoAI-ADK standards

        Use Skill("moai-foundation-specs") and Skill("moai-foundation-ears") for guidance.
        """,
        resume=resume_id if resume_id else None
    )

    # Register result
    register_agent(
        agent_name="spec-builder",
        agent_id=spec_result["agent_id"],
        result=spec_result,
        chain_id=f"{spec_result['spec_id']}-planning"
    )

    # Store in Alfred context
    alfred_context["agent_results"]["spec-builder"] = spec_result
    alfred_context["spec_id"] = spec_result["spec_id"]
    alfred_context["workflow_state"]["completed_steps"].append("spec_creation")

    print(f"✅ SPEC created: {spec_result['spec_id']}")
    print(f"   Files: {', '.join(spec_result['files_created'])}")

    # ===== STEP 2: 구현 계획 수립 (implementation-planner) =====
    print(f"\n🛠️ STEP 2: Creating implementation plan...")
    alfred_context["workflow_state"]["current_step"] = "planning"

    # Invoke implementation-planner
    plan_result = Task(
        subagent_type="implementation-planner",
        prompt=f"""You are the implementation-planner agent.

        SPEC has been created: {alfred_context['spec_id']}
        Location: .moai/specs/{alfred_context['spec_id']}/spec.md

        Your tasks:
        1. Read SPEC thoroughly
        2. Break down into TAG chain
        3. Identify library dependencies (use WebFetch for latest versions)
        4. Define implementation sequence with priorities
        5. Assess risks and mitigation strategies

        SPEC summary from spec-builder:
        {json.dumps(spec_result['summary'], indent=2)}

        Generate detailed plan in Korean.
        """
    )

    # Register result
    register_agent(
        agent_name="implementation-planner",
        agent_id=plan_result["agent_id"],
        result=plan_result,
        chain_id=f"{alfred_context['spec_id']}-planning"
    )

    # Store in Alfred context
    alfred_context["agent_results"]["implementation-planner"] = plan_result
    alfred_context["tag_chain"] = plan_result["tag_chain"]
    alfred_context["dependencies"] = plan_result["dependencies"]
    alfred_context["workflow_state"]["completed_steps"].append("planning")

    print(f"✅ Implementation plan created")
    print(f"   TAGs: {len(plan_result['tag_chain'])}")
    print(f"   Dependencies: {list(plan_result['dependencies'].keys())}")

    # ===== STEP 3: Expert Consultation (선택적) =====
    if expert_consultation or plan_result.get("requires_expert_review"):
        print(f"\n🧑‍💼 STEP 3: Expert consultation...")
        alfred_context["workflow_state"]["current_step"] = "expert_consultation"

        # Determine which experts to consult
        required_experts = identify_required_experts(plan_result)

        for expert_type in required_experts:
            print(f"   Consulting {expert_type}...")

            expert_result = Task(
                subagent_type=f"{expert_type}-expert",
                prompt=f"""You are the {expert_type}-expert agent.

                Review SPEC: {alfred_context['spec_id']}
                Focus: {get_expert_focus(expert_type)}

                Implementation plan:
                {json.dumps(plan_result['summary'], indent=2)}

                Provide:
                1. Architecture recommendations
                2. Risk identification
                3. Best practice suggestions
                4. Technology choices review
                """
            )

            # Register (independent sessions, no resume)
            register_agent(
                agent_name=f"{expert_type}-expert",
                agent_id=expert_result["agent_id"],
                result=expert_result,
                chain_id=f"{alfred_context['spec_id']}-review"
            )

            alfred_context["expert_reviews"][expert_type] = expert_result

        alfred_context["workflow_state"]["completed_steps"].append("expert_consultation")
        print(f"✅ {len(required_experts)} expert reviews completed")

    # ===== STEP 4: 사용자 승인 =====
    print(f"\n✅ Planning complete! Awaiting user approval...")

    # Prepare summary for user
    summary = {
        "spec_id": alfred_context["spec_id"],
        "tag_count": len(alfred_context["tag_chain"]),
        "tags": alfred_context["tag_chain"],
        "dependencies": alfred_context["dependencies"],
        "expert_reviews": list(alfred_context["expert_reviews"].keys()),
        "next_steps": [
            "/alfred:2-run " + alfred_context["spec_id"],
            "Review and modify SPEC if needed",
            "Proceed with implementation when ready"
        ]
    }

    # Ask user for next action
    user_decision = AskUserQuestion(
        questions=[{
            "question": f"Planning complete for {alfred_context['spec_id']}. What would you like to do?",
            "header": "Next Step",
            "multiSelect": False,
            "options": [
                {
                    "label": "Proceed to Implementation",
                    "description": f"Run /alfred:2-run {alfred_context['spec_id']}"
                },
                {
                    "label": "Revise SPEC",
                    "description": "Resume spec-builder to modify SPEC"
                },
                {
                    "label": "Review Later",
                    "description": "Save state and continue later"
                }
            ]
        }]
    )

    # Handle user decision
    if user_decision == "Proceed to Implementation":
        # Automatically trigger /alfred:2-run
        return execute_alfred_run(alfred_context["spec_id"], alfred_context)

    elif user_decision == "Revise SPEC":
        # Resume spec-builder with expert feedback
        revised_spec = Task(
            subagent_type="spec-builder",
            prompt=f"""Continue SPEC creation for {alfred_context['spec_id']}.

            Expert feedback received:
            {json.dumps(alfred_context['expert_reviews'], indent=2)}

            User requested revisions. Update SPEC to address concerns.
            """,
            resume=spec_result["agent_id"]  # 🔑 Resume with full context
        )

        register_agent(
            agent_name="spec-builder",
            agent_id=revised_spec["agent_id"],
            result=revised_spec,
            chain_id=f"{alfred_context['spec_id']}-planning"
        )

        return {
            "status": "revised",
            "spec_id": alfred_context["spec_id"],
            "message": "SPEC updated based on feedback"
        }

    else:  # Review Later
        return {
            "status": "pending",
            "spec_id": alfred_context["spec_id"],
            "summary": summary,
            "message": f"Planning saved. Run /alfred:2-run {alfred_context['spec_id']} when ready."
        }


def identify_required_experts(plan_result: Dict[str, Any]) -> List[str]:
    """
    Determine which expert agents to consult based on plan.

    Args:
        plan_result: Implementation plan from implementation-planner

    Returns:
        List of expert types (e.g., ["backend", "security", "frontend"])
    """
    experts = []

    # Backend expert if API/database involved
    if any(keyword in str(plan_result).lower() for keyword in ["api", "database", "server", "backend"]):
        experts.append("backend")

    # Frontend expert if UI involved
    if any(keyword in str(plan_result).lower() for keyword in ["ui", "component", "frontend", "client"]):
        experts.append("frontend")

    # Security expert if auth/security involved
    if any(keyword in str(plan_result).lower() for keyword in ["auth", "security", "password", "token"]):
        experts.append("security")

    # DevOps expert if deployment involved
    if any(keyword in str(plan_result).lower() for keyword in ["deploy", "docker", "kubernetes", "ci/cd"]):
        experts.append("devops")

    return experts


def get_expert_focus(expert_type: str) -> str:
    """Get focus area for expert type."""
    focus_map = {
        "backend": "API design, database schema, authentication strategy, security",
        "frontend": "UI/UX, component design, state management, accessibility",
        "security": "OWASP compliance, authentication, authorization, data protection",
        "devops": "Deployment strategy, CI/CD, infrastructure, monitoring"
    }
    return focus_map.get(expert_type, "General architecture and best practices")
```

---

## 예제 2: `/alfred:2-run` - TDD 구현 실행

### Workflow

```
User → /alfred:2-run SPEC-XXX → tdd-implementer (resume) → quality-gate → git-manager → doc-syncer
```

### 완전한 구현

```python
# File: .claude/commands/alfred-2-run.py (개념적 예시)

def execute_alfred_run(
    spec_id: str,
    context: Optional[Dict[str, Any]] = None
) -> Dict[str, Any]:
    """
    /alfred:2-run 명령 실행

    Workflow:
    1. Load or create implementation plan
    2. tdd-implementer: Execute TDD cycle for all TAGs (with resume)
    3. quality-gate: Verify implementation
    4. (If failed) debug-helper: Analyze and suggest fixes
    5. git-manager: Create TDD commits
    6. doc-syncer: Update documentation

    Args:
        spec_id: SPEC identifier (e.g., "SPEC-AUTH-001")
        context: Optional context from /alfred:1-plan

    Returns:
        실행 결과
    """
    # ===== 초기화 =====
    session_mgr = SessionManager()

    alfred_context = context or {
        "command": "/alfred:2-run",
        "spec_id": spec_id,
        "agent_results": {},
        "workflow_state": {}
    }

    print(f"🎩 Alfred: Starting /alfred:2-run for {spec_id}")

    # ===== STEP 1: 구현 계획 확인 =====
    if "tag_chain" not in alfred_context:
        print("\n📋 STEP 1: Loading implementation plan...")

        plan_result = Task(
            subagent_type="implementation-planner",
            prompt=f"""You are the implementation-planner agent.

            Load implementation plan for {spec_id}.
            Read SPEC from .moai/specs/{spec_id}/spec.md

            Provide:
            - TAG chain breakdown
            - Library dependencies
            - Implementation sequence
            """
        )

        register_agent(
            agent_name="implementation-planner",
            agent_id=plan_result["agent_id"],
            result=plan_result,
            chain_id=f"{spec_id}-implementation"
        )

        alfred_context["tag_chain"] = plan_result["tag_chain"]
        alfred_context["dependencies"] = plan_result["dependencies"]

    print(f"✅ Plan loaded: {len(alfred_context['tag_chain'])} TAGs to implement")

    # ===== STEP 2: TDD 구현 (tdd-implementer with resume) =====
    print(f"\n🔬 STEP 2: TDD Implementation...")

    # Create workflow chain
    session_mgr.create_chain(
        chain_id=f"{spec_id}-implementation",
        agent_sequence=["tdd-implementer", "quality-gate", "git-manager", "doc-syncer"],
        metadata={
            "spec_id": spec_id,
            "tag_chain": alfred_context["tag_chain"]
        }
    )

    # First TAG implementation
    tdd_result = Task(
        subagent_type="tdd-implementer",
        prompt=f"""You are the tdd-implementer agent.

        SPEC: {spec_id}
        TAG chain: {alfred_context['tag_chain']}

        Execute TDD cycle:
        1. RED: Write failing tests for {alfred_context['tag_chain'][0]}
        2. GREEN: Write minimal passing code
        3. REFACTOR: Improve code quality

        Report progress after each phase.
        """
    )

    # Register first execution
    tdd_agent_id = tdd_result["agent_id"]
    register_agent(
        agent_name="tdd-implementer",
        agent_id=tdd_agent_id,
        result=tdd_result,
        chain_id=f"{spec_id}-implementation"
    )

    alfred_context["agent_results"]["tdd-implementer"] = tdd_result
    completed_tags = [tdd_result["current_tag"]]

    print(f"✅ TAG {tdd_result['current_tag']} complete (1/{len(alfred_context['tag_chain'])})")

    # Continue with remaining TAGs (resume pattern)
    for i, tag in enumerate(alfred_context['tag_chain'][1:], start=2):
        print(f"\n🔬 Implementing TAG {tag} ({i}/{len(alfred_context['tag_chain'])})...")

        # Resume tdd-implementer for next TAG
        tdd_result = Task(
            subagent_type="tdd-implementer",
            prompt=f"""Continue TDD implementation for TAG {tag}.

            Previous TAG {completed_tags[-1]} is complete.
            Maintain code quality and test coverage.

            Execute RED-GREEN-REFACTOR cycle for TAG {tag}.
            """,
            resume=tdd_agent_id  # 🔑 Resume with full context
        )

        # Update resume count
        session_mgr.increment_resume_count(tdd_agent_id)

        completed_tags.append(tag)
        print(f"✅ TAG {tag} complete ({i}/{len(alfred_context['tag_chain'])})")

    # Final implementation result
    alfred_context["agent_results"]["tdd-implementer"] = tdd_result
    alfred_context["workflow_state"]["tdd_complete"] = True

    print(f"\n✅ All TAGs implemented: {len(completed_tags)}/{len(alfred_context['tag_chain'])}")

    # ===== STEP 3: 품질 검증 (quality-gate) =====
    print(f"\n🔍 STEP 3: Quality validation...")

    qa_result = Task(
        subagent_type="quality-gate",
        prompt=f"""You are the quality-gate agent.

        Verify implementation of {spec_id}:
        1. Run test suite and check coverage (target: 85%)
        2. Run linting (ruff for Python)
        3. Run type checking (mypy for Python)
        4. Verify TRUST principles compliance
        5. Validate TAG chain integrity

        Implementation summary:
        - TAGs: {completed_tags}
        - Files created: {tdd_result.get('files_created', [])}

        Provide detailed quality report.
        """
    )

    register_agent(
        agent_name="quality-gate",
        agent_id=qa_result["agent_id"],
        result=qa_result,
        chain_id=f"{spec_id}-implementation"
    )

    alfred_context["agent_results"]["quality-gate"] = qa_result

    # ===== STEP 3.1: 품질 검증 실패 처리 =====
    if qa_result["status"] != "success":
        print(f"❌ Quality validation failed")
        print(f"   Issues: {len(qa_result['issues'])}")

        # Invoke debug-helper
        print(f"\n🔧 Invoking debug-helper...")

        debug_result = Task(
            subagent_type="debug-helper",
            prompt=f"""You are the debug-helper agent.

            Quality gate failed for {spec_id}.

            Issues:
            {json.dumps(qa_result['issues'], indent=2)}

            Provide:
            1. Root cause analysis for each issue
            2. Specific fix recommendations
            3. Code snippets for fixes
            """
        )

        register_agent(
            agent_name="debug-helper",
            agent_id=debug_result["agent_id"],
            result=debug_result,
            chain_id=f"{spec_id}-debugging"
        )

        # Ask user how to proceed
        user_action = AskUserQuestion(
            questions=[{
                "question": "Quality validation failed. How would you like to proceed?",
                "header": "Fix Strategy",
                "multiSelect": False,
                "options": [
                    {
                        "label": "Auto-fix with debug-helper recommendations",
                        "description": "Resume tdd-implementer to apply fixes"
                    },
                    {
                        "label": "Manual fix",
                        "description": "Fix code manually and re-run validation"
                    },
                    {
                        "label": "Abort",
                        "description": "Stop execution and review issues"
                    }
                ]
            }]
        )

        if user_action == "Auto-fix with debug-helper recommendations":
            # Resume tdd-implementer to apply fixes
            fix_result = Task(
                subagent_type="tdd-implementer",
                prompt=f"""Apply fixes based on debug-helper analysis.

                Issues:
                {json.dumps(qa_result['issues'], indent=2)}

                Recommendations:
                {json.dumps(debug_result['recommendations'], indent=2)}

                Fix code and re-run tests.
                """,
                resume=tdd_agent_id  # 🔑 Resume to maintain context
            )

            session_mgr.increment_resume_count(tdd_agent_id)

            # Re-run quality gate
            qa_result = Task(
                subagent_type="quality-gate",
                prompt=f"Re-verify {spec_id} after fixes"
            )

            register_agent(
                agent_name="quality-gate",
                agent_id=qa_result["agent_id"],
                result=qa_result,
                chain_id=f"{spec_id}-implementation"
            )

        elif user_action == "Abort":
            return {
                "status": "failed",
                "spec_id": spec_id,
                "issues": qa_result["issues"],
                "debug_recommendations": debug_result.get("recommendations", [])
            }

    print(f"✅ Quality validation passed")
    print(f"   Coverage: {qa_result['coverage']}%")
    print(f"   Tests: {qa_result['tests_passed']}/{qa_result['tests_total']}")

    # ===== STEP 4: Git Commit (git-manager) =====
    print(f"\n📝 STEP 4: Creating Git commit...")

    commit_result = Task(
        subagent_type="git-manager",
        prompt=f"""You are the git-manager agent.

        Create TDD commit for {spec_id}:
        1. Stage implementation files
        2. Generate commit message following conventional commits
        3. Include TAG references
        4. Add quality metrics to commit message

        Implementation summary:
        - TAGs: {completed_tags}
        - Coverage: {qa_result['coverage']}%
        - Tests: {qa_result['tests_passed']}/{qa_result['tests_total']}

        Files to commit:
        {json.dumps(tdd_result.get('files_created', []), indent=2)}
        """
    )

    register_agent(
        agent_name="git-manager",
        agent_id=commit_result["agent_id"],
        result=commit_result,
        chain_id=f"{spec_id}-implementation"
    )

    alfred_context["agent_results"]["git-manager"] = commit_result

    print(f"✅ Commit created: {commit_result['commit_sha'][:7]}")
    print(f"   Message: {commit_result['commit_message'].split(chr(10))[0]}")

    # ===== STEP 5: 문서 동기화 (doc-syncer) =====
    print(f"\n📚 STEP 5: Synchronizing documentation...")

    doc_result = Task(
        subagent_type="doc-syncer",
        prompt=f"""You are the doc-syncer agent.

        Synchronize documentation for {spec_id}:
        1. Update .moai/project/product.md (new feature added)
        2. Update .moai/project/structure.md (if architecture changed)
        3. Update .moai/project/tech.md (if new libraries added)
        4. Ensure TAG chain consistency

        Implementation details:
        - TAGs: {completed_tags}
        - Features: {tdd_result.get('features_implemented', [])}
        - Dependencies: {alfred_context.get('dependencies', {})}
        """
    )

    register_agent(
        agent_name="doc-syncer",
        agent_id=doc_result["agent_id"],
        result=doc_result,
        chain_id=f"{spec_id}-implementation"
    )

    alfred_context["agent_results"]["doc-syncer"] = doc_result

    print(f"✅ Documentation synchronized")
    print(f"   Updated: {', '.join(doc_result['files_updated'])}")

    # ===== STEP 6: 최종 요약 =====
    print(f"\n🎉 Implementation complete for {spec_id}!")

    summary = {
        "status": "success",
        "spec_id": spec_id,
        "tags_implemented": completed_tags,
        "coverage": qa_result["coverage"],
        "commit_sha": commit_result["commit_sha"],
        "files_created": tdd_result.get("files_created", []),
        "files_updated": doc_result["files_updated"],
        "next_command": "/alfred:3-sync"
    }

    # Ask for next step
    user_next = AskUserQuestion(
        questions=[{
            "question": f"Implementation complete for {spec_id}. What's next?",
            "header": "Next Step",
            "multiSelect": False,
            "options": [
                {
                    "label": "Run /alfred:3-sync",
                    "description": "Synchronize all documentation and create PR"
                },
                {
                    "label": "Implement another SPEC",
                    "description": "Start /alfred:1-plan for new feature"
                },
                {
                    "label": "Review and test",
                    "description": "Manual review before proceeding"
                }
            ]
        }]
    )

    if user_next == "Run /alfred:3-sync":
        return execute_alfred_sync(spec_id, alfred_context)
    else:
        return summary
```

---

## 예제 3: Resume Pattern - 여러 문서 연속 업데이트

### Workflow

```
doc-syncer: product.md → structure.md → tech.md (모두 resume로 연결)
```

### 구현

```python
def sync_all_documents(spec_id: str, implementation_summary: Dict[str, Any]) -> Dict[str, Any]:
    """
    모든 프로젝트 문서를 연속적으로 업데이트 (resume 활용)

    Args:
        spec_id: SPEC ID
        implementation_summary: 구현 요약 정보

    Returns:
        동기화 결과
    """
    session_mgr = SessionManager()

    documents = ["product.md", "structure.md", "tech.md"]
    doc_syncer_id = None

    for i, doc in enumerate(documents):
        print(f"\n📄 Updating {doc} ({i+1}/{len(documents)})...")

        # First execution or resume
        is_first = (i == 0)

        doc_result = Task(
            subagent_type="doc-syncer",
            prompt=f"""You are the doc-syncer agent.

            {"This is the first document in a series." if is_first else f"Continue updating documents. {documents[i-1]} is complete."}

            Update {doc} with implementation of {spec_id}:
            {json.dumps(implementation_summary, indent=2)}

            Maintain consistent style and cross-references.
            """,
            resume=doc_syncer_id if not is_first else None  # 🔑 Resume from 2nd onwards
        )

        # Save agent ID on first execution
        if is_first:
            doc_syncer_id = doc_result["agent_id"]

        register_agent(
            agent_name="doc-syncer",
            agent_id=doc_result["agent_id"],
            result=doc_result,
            chain_id=f"{spec_id}-documentation"
        )

        if not is_first:
            session_mgr.increment_resume_count(doc_syncer_id)

        print(f"✅ {doc} updated")

    return {
        "status": "success",
        "documents_updated": documents,
        "agent_id": doc_syncer_id
    }
```

**Resume의 이점**:

- ✅ 일관된 스타일 유지 (같은 용어, 같은 구조)
- ✅ 문서 간 상호 참조 정확성
- ✅ 중복 설명 불필요 (첫 문서에서 이미 설명)

---

## 예제 4: 병렬 실행 - 여러 전문가 동시 자문

### Workflow

```
                  ┌─ backend-expert
Alfred (parallel) ┼─ frontend-expert → Alfred (merge results) → spec-builder (update)
                  └─ security-expert
```

### 구현

```python
import asyncio
from typing import List

async def parallel_expert_consultation(
    spec_id: str,
    experts: List[str]
) -> Dict[str, Any]:
    """
    여러 전문가에게 동시에 자문 (병렬 실행)

    Args:
        spec_id: SPEC ID
        experts: 전문가 타입 리스트 (예: ["backend", "security", "frontend"])

    Returns:
        통합 자문 결과
    """
    print(f"🧑‍💼 Consulting {len(experts)} experts in parallel...")

    # Prepare tasks (conceptual - Task() is synchronous in reality)
    expert_tasks = []

    for expert_type in experts:
        # Each expert runs independently (no resume needed)
        expert_result = Task(
            subagent_type=f"{expert_type}-expert",
            prompt=f"""You are the {expert_type}-expert agent.

            Review SPEC {spec_id} for {expert_type} concerns.

            Provide:
            - Architecture recommendations
            - Risk identification
            - Best practices
            - Technology choices
            """
        )

        # Register each independently
        register_agent(
            agent_name=f"{expert_type}-expert",
            agent_id=expert_result["agent_id"],
            result=expert_result,
            chain_id=f"{spec_id}-expert-review"
        )

        expert_tasks.append(expert_result)

    # Merge results
    merged_feedback = {
        "spec_id": spec_id,
        "expert_count": len(experts),
        "recommendations": [],
        "risks": [],
        "action_items": []
    }

    for expert_result in expert_tasks:
        merged_feedback["recommendations"].extend(expert_result.get("recommendations", []))
        merged_feedback["risks"].extend(expert_result.get("risks", []))
        merged_feedback["action_items"].extend(expert_result.get("action_items", []))

    print(f"✅ {len(experts)} expert reviews merged")

    return merged_feedback
```

---

## SessionManager 통합 패턴

### Pattern 1: 기본 등록

```python
# Agent 실행 후 즉시 등록
result = Task(subagent_type="agent-name", prompt="...")

register_agent(
    agent_name="agent-name",
    agent_id=result["agent_id"],
    result=result,
    chain_id="workflow-chain-id"
)
```

---

### Pattern 2: Resume 결정

```python
# Should resume?
should_resume_decision = session_mgr.should_resume(
    agent_name="tdd-implementer",
    current_task="Implement TAG-002",
    previous_task="Implement TAG-001"
)

if should_resume_decision:
    resume_id = get_resume_id("tdd-implementer", chain_id="SPEC-XXX-implementation")
    result = Task(subagent_type="tdd-implementer", prompt="...", resume=resume_id)
    session_mgr.increment_resume_count(resume_id)
else:
    result = Task(subagent_type="tdd-implementer", prompt="...")
```

---

### Pattern 3: Chain 결과 조회

```python
# Get all results in a chain
chain_results = session_mgr.get_chain_results("SPEC-AUTH-001-implementation")

for result in chain_results:
    print(f"{result['agent_name']}: {result['timestamp']}")
```

---

## Error Handling 패턴

### Pattern 1: Agent 실패 시 재시도

```python
max_retries = 2
for attempt in range(max_retries + 1):
    try:
        result = Task(subagent_type="agent-name", prompt="...")

        if result["status"] == "success":
            register_agent("agent-name", result["agent_id"], result)
            break
    except Exception as e:
        if attempt < max_retries:
            print(f"Retry {attempt + 1}/{max_retries}...")
            # Start new session on retry (don't resume failed session)
            continue
        else:
            # Escalate to debug-helper
            debug_result = Task(
                subagent_type="debug-helper",
                prompt=f"Analyze failure: {str(e)}"
            )
            raise
```

---

### Pattern 2: Quality Gate 실패 루프

```python
max_iterations = 3

for iteration in range(max_iterations):
    # Implement
    impl_result = Task(
        subagent_type="tdd-implementer",
        prompt=f"Implement (iteration {iteration + 1})",
        resume=tdd_id if iteration > 0 else None
    )

    # Validate
    qa_result = Task(subagent_type="quality-gate", prompt="Validate")

    if qa_result["status"] == "success":
        break  # Success!

    if iteration < max_iterations - 1:
        # Debug and retry
        debug_result = Task(subagent_type="debug-helper", prompt=f"Fix: {qa_result['issues']}")
        tdd_id = impl_result["agent_id"]
    else:
        # Max iterations reached
        raise QualityGateError("Failed after max iterations")
```

---

## Best Practices

### ✅ DO

1. **모든 agent 결과를 즉시 등록**
   ```python
   result = Task(...)
   register_agent(agent_name, result["agent_id"], result, chain_id)
   ```

2. **Chain ID 일관성 유지**
   ```python
   chain_id = f"{spec_id}-{workflow_type}"  # 예: "SPEC-AUTH-001-implementation"
   ```

3. **Resume 사용 시 increment**
   ```python
   result = Task(..., resume=resume_id)
   session_mgr.increment_resume_count(resume_id)
   ```

4. **Alfred context에 결과 저장**
   ```python
   alfred_context["agent_results"][agent_name] = result
   ```

5. **Workflow 상태 추적**
   ```python
   alfred_context["workflow_state"]["completed_steps"].append("planning")
   ```

---

### ❌ DON'T

1. **Agent가 다른 agent 호출**
   ```python
   # ❌ In agent file
   result = Task(subagent_type="other-agent", ...)
   ```

2. **파일로 agent 간 통신**
   ```python
   # ❌ In agent file
   Write(".moai/temp/plan.json", plan_data)
   # Next agent reads this file
   ```

3. **Resume 없이 연속 작업**
   ```python
   # ❌ TAG-001, TAG-002를 별개 session으로
   Task(subagent_type="tdd-implementer", prompt="TAG-001")
   Task(subagent_type="tdd-implementer", prompt="TAG-002")  # Context 손실!
   ```

4. **Resume count 미증가**
   ```python
   # ❌ Resume 사용했는데 count 안 올림
   Task(..., resume=resume_id)
   # session_mgr.increment_resume_count(resume_id) 누락
   ```

---

## 참고 자료

- **Alfred Orchestration**: `.moai/config/alfred-orchestration.yaml`
- **Agent Invocation**: `.moai/guidelines/agent-invocation.md`
- **SessionManager**: `src/moai_adk/core/session_manager.py`
- **Official Docs**: https://code.claude.com/docs/en/sub-agents

---

**Last Updated**: 2025-11-12
**Version**: 1.0.0
