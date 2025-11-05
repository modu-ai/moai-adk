# @TEST:VAL-001
"""Test suite for TAG chain repair tool.

Tests TAG chain repair functionality including plan creation,
template generation, and repair execution.

@SPEC:DOCS-004: TAG 체인 복구 및 자동 생성 도구
"""

import tempfile
import textwrap
from pathlib import Path

from src.moai_adk.core.analysis.tag_chain_repair import (
    RepairTask,
    RepairPlan,
    TagChainRepairer,
    repair_tag_chains,
)


def test_repair_task_creation():
    """Test RepairTask creation and properties."""
    task = RepairTask(
        domain="AUTH",
        number=1,
        action="create_spec",
        priority="high",
        estimated_effort="medium"
    )

    assert task.domain == "AUTH"
    assert task.number == 1
    assert task.action == "create_spec"
    assert task.priority == "high"
    assert task.estimated_effort == "medium"
    assert task.dependencies == []


def test_repair_plan_organization():
    """Test RepairPlan task organization by priority."""
    high_task = RepairTask("AUTH", 1, "create_spec", "high", "low")
    medium_task = RepairTask("CLI", 1, "create_code", "medium", "medium")
    low_task = RepairTask("UTILS", 1, "create_test", "low", "high")

    plan = RepairPlan(
        high_priority_tasks=[high_task],
        medium_priority_tasks=[medium_task],
        low_priority_tasks=[low_task],
        summary={"total": 3, "high": 1, "medium": 1, "low": 1}
    )

    priorities = plan.get_tasks_by_priority()
    assert priorities[0][0] == "high"
    assert priorities[1][0] == "medium"
    assert priorities[2][0] == "low"
    assert priorities[0][1] == [high_task]
    assert priorities[1][1] == [medium_task]
    assert priorities[2][1] == [low_task]


def test_extract_number_from_domain():
    """Test number extraction from domain strings."""
    repairer = TagChainRepairer()

    assert repairer._extract_number_from_domain("AUTH-001") == 1
    assert repairer._extract_number_from_domain("LDE-PRIORITY-005") == 5
    assert repairer._extract_number_from_domain("CLI-002") == 2
    assert repairer._extract_number_from_domain("CORE-PROJECT") == 1


def test_extract_domain_from_tag():
    """Test domain extraction from TAG strings."""
    repairer = TagChainRepairer()

    assert repairer._extract_domain_from_tag("@SPEC:AUTH-001") == "AUTH"
    assert repairer._extract_domain_from_tag("@CODE:LDE-PRIORITY-005") == "LDE-PRIORITY"
    assert repairer._extract_domain_from_tag("@TEST:CLI-002") == "CLI"


def test_create_spec_template():
    """Test SPEC template generation."""
    repairer = TagChainRepairer()

    # Test LDE template
    template = repairer._create_spec_template("LDE-PRIORITY", 1)
    assert "@SPEC:LDE-PRIORITY-001" in template
    assert "LDE 기능 개선" in template
    assert "labels:" in template
    assert "lde" in template

    # Test CORE template
    template = repairer._create_spec_template("CORE-PROJECT", 1)
    assert "@SPEC:CORE-PROJECT-001" in template
    assert "핵심 기능 개선" in template
    assert "core" in template

    # Test INSTALLER template
    template = repairer._create_spec_template("INSTALLER-QUALITY", 1)
    assert "@SPEC:INSTALLER-QUALITY-001" in template
    assert "설치기 개선" in template
    assert "installer" in template


def test_create_code_template():
    """Test CODE template generation."""
    repairer = TagChainRepairer()

    # Test LDE code template
    template = repairer._create_code_template("LDE-PRIORITY", 1)
    assert "@CODE:LDE-PRIORITY-001" in template
    assert "def lde_priority_function(" in template
    assert "from typing import" in template
    assert "if __name__ == \"__main__\":" in template

    # Test CORE code template
    template = repairer._create_code_template("CORE-PROJECT", 1)
    assert "@CODE:CORE-PROJECT-001" in template
    assert "def core_project_function(" in template


def test_create_test_template():
    """Test TEST template generation."""
    repairer = TagChainRepairer()

    template = repairer._create_test_template("AUTH", 1)
    assert "@TEST:AUTH-001" in template
    assert "class TestAuth:" in template
    assert "def test_auth_function_basic(self):" in template
    assert "import pytest" in template
    assert "if __name__ == \"__main__\":" in template


def test_repair_plan_creation():
    """Test repair plan creation from analysis result."""
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_path = Path(temp_dir)

        # Create mock analysis result
        mock_result = type('MockResult', (), {
            'broken_chain_details': [
                {'domain': 'AUTH', 'missing': ['SPEC'], 'score': 0.33},
                {'domain': 'CLI', 'missing': ['TEST'], 'score': 0.33},
                {'domain': 'UTILS', 'missing': ['CODE'], 'score': 0.33}
            ],
            'orphans_by_type': {
                'code_without_spec': ['@CODE:TEST-001'],
                'code_without_test': ['@CODE:TEST-002'],
                'test_without_code': ['@TEST:TEST-003'],
                'spec_without_code': ['@SPEC:TEST-004']
            }
        })()

        repairer = TagChainRepairer(temp_path)
        plan = repairer._create_repair_plan(mock_result)

        # Check plan structure
        assert isinstance(plan, RepairPlan)
        assert len(plan.high_priority_tasks) >= 2  # AUTH and CLI should be high priority
        assert len(plan.medium_priority_tasks) >= 2  # UTILS should be medium priority
        assert plan.summary['total_tasks'] >= 7  # 3 from broken chains + 4 from orphans

        # Check specific tasks
        auth_spec_task = next((t for t in plan.high_priority_tasks if t.domain == 'AUTH' and t.action == 'create_spec'), None)
        assert auth_spec_task is not None
        assert auth_spec_task.priority == 'high'


def test_execute_repair_task():
    """Test individual repair task execution."""
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_path = Path(temp_dir)

        repairer = TagChainRepairer(temp_path)

        # Test creating SPEC
        spec_task = RepairTask("TEST-DOMAIN", 1, "create_spec", "high", "low")
        spec_files = repairer._execute_repair_task(spec_task)

        assert len(spec_files) == 1
        spec_path = Path(spec_files[0])
        assert spec_path.exists()
        assert spec_path.name == "spec-test-domain-001.md"
        assert "@SPEC:TEST-DOMAIN-001" in spec_path.read_text()

        # Test creating CODE
        code_task = RepairTask("TEST-DOMAIN", 2, "create_code", "high", "low")
        code_files = repairer._execute_repair_task(code_task)

        assert len(code_files) == 1
        code_path = Path(code_files[0])
        assert code_path.exists()
        assert code_path.name == "test_domain.py"
        assert "@CODE:TEST-DOMAIN-002" in code_path.read_text()

        # Test creating TEST
        test_task = RepairTask("TEST-DOMAIN", 3, "create_test", "high", "low")
        test_files = repairer._execute_repair_task(test_task)

        assert len(test_files) == 1
        test_path = Path(test_files[0])
        assert test_path.exists()
        assert test_path.name == "test_test_domain.py"
        assert "@TEST:TEST-DOMAIN-003" in test_path.read_text()


def test_execute_repair_plan_dry_run():
    """Test repair plan execution with dry run."""
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_path = Path(temp_dir)

        # Create mock analysis result
        mock_result = type('MockResult', (), {
            'broken_chain_details': [
                {'domain': 'AUTH', 'missing': ['SPEC'], 'score': 0.33}
            ],
            'orphans_by_type': {
                'code_without_test': ['@CODE:TEST-001']
            }
        })()

        repairer = TagChainRepairer(temp_path)
        plan = repairer._create_repair_plan(mock_result)
        results = repairer.execute_repair_plan(plan, dry_run=True)

        # Check dry run results
        assert len(results['skipped']) == 2  # 1 from broken chain + 1 from orphans
        assert len(results['created']) == 0  # No files should be created in dry run
        assert "DRY RUN" in results['skipped'][0]


def test_execute_repair_plan_actual():
    """Test repair plan execution without dry run."""
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_path = Path(temp_dir)

        # Create mock analysis result
        mock_result = type('MockResult', (), {
            'broken_chain_details': [
                {'domain': 'TEST', 'missing': ['SPEC'], 'score': 0.33}
            ],
            'orphans_by_type': {}
        })()

        repairer = TagChainRepairer(temp_path)
        plan = repairer._create_repair_plan(mock_result)
        results = repairer.execute_repair_plan(plan, dry_run=False)

        # Check actual execution results
        assert len(results['created']) == 1  # Should create SPEC file
        assert len(results['skipped']) == 0
        assert len(results['errors']) == 0

        # Verify file was actually created
        created_file = Path(results['created'][0])
        assert created_file.exists()
        assert "TEST" in created_file.read_text()


def test_convenience_function():
    """Test convenience repair function."""
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_path = Path(temp_dir)

        result, plan, execution_results = repair_tag_chains(temp_path)

        assert isinstance(result, object)  # Should be analysis result
        assert isinstance(plan, RepairPlan)
        assert isinstance(execution_results, dict)

        # Should have tasks in the plan
        assert len(plan.high_priority_tasks) + len(plan.medium_priority_tasks) + len(plan.low_priority_tasks) > 0


def test_integration_with_real_structure():
    """Test integration with actual MoAI-ADK structure."""
    repairer = TagChainRepairer(Path("."))
    result, plan = repairer.analyze_and_create_plan()

    # Should find tasks in the real codebase
    assert result.total_chains > 0
    assert len(plan.high_priority_tasks) > 0 or len(plan.medium_priority_tasks) > 0

    # Show sample tasks
    print("\n" + "="*50)
    print("INTEGRATION TEST - TAG CHAIN REPAIR")
    print("="*50)
    print(f"Total chains: {result.total_chains}")
    print(f"Broken chains: {result.broken_chains}")
    print(f"High priority tasks: {len(plan.high_priority_tasks)}")
    print(f"Medium priority tasks: {len(plan.medium_priority_tasks)}")
    print(f"Low priority tasks: {len(plan.low_priority_tasks)}")

    if plan.high_priority_tasks:
        print("\nSample high priority tasks:")
        for task in plan.high_priority_tasks[:3]:
            print(f"- {task.action.upper()} for {task.domain}-{task.number:03d}")

    print("="*50)