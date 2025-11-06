# @CODE:VAL-001
"""TAG chain repair tool for creating missing links.

Automatically creates missing SPEC, CODE, and TEST elements to
restore broken TAG chains based on priority and domain.

@SPEC:DOCS-005: TAG 체인 복구 및 자동 생성 도구
"""

import re
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, List, Tuple

from moai_adk.core.analysis.tag_chain_analyzer import (
    ChainAnalysisResult,
    TagChainAnalyzer,
)


@dataclass
class RepairTask:
    """Represents a TAG chain repair task."""

    domain: str
    number: int
    action: str  # "create_spec", "create_code", "create_test"
    priority: str  # "high", "medium", "low"
    estimated_effort: str  # "low", "medium", "high"
    dependencies: List[str] = None

    def __post_init__(self):
        if self.dependencies is None:
            self.dependencies = []


@dataclass
class RepairPlan:
    """Represents a complete TAG chain repair plan."""

    high_priority_tasks: List[RepairTask]
    medium_priority_tasks: List[RepairTask]
    low_priority_tasks: List[RepairTask]
    summary: Dict[str, int]

    def get_tasks_by_priority(self) -> List[Tuple[str, List[RepairTask]]]:
        """Get all tasks organized by priority."""
        return [
            ("high", self.high_priority_tasks),
            ("medium", self.medium_priority_tasks),
            ("low", self.low_priority_tasks)
        ]


class TagChainRepairer:
    """Repairs broken TAG chains by creating missing elements."""

    def __init__(self, root_path: Path = Path(".")):
        self.root_path = root_path
        self.analyzer = TagChainAnalyzer(root_path)
        self.repair_templates = {
            "create_spec": self._create_spec_template,
            "create_code": self._create_code_template,
            "create_test": self._create_test_template,
        }

    def analyze_and_create_plan(self) -> Tuple[ChainAnalysisResult, RepairPlan]:
        """Analyze TAG chains and create repair plan."""
        result = self.analyzer.analyze_all_chains()
        plan = self._create_repair_plan(result)
        return result, plan

    def _create_repair_plan(self, result: ChainAnalysisResult) -> RepairPlan:
        """Create repair plan based on analysis result."""
        tasks = []
        summary = {"total_tasks": 0, "high_priority": 0, "medium_priority": 0, "low_priority": 0}

        # Define high-priority domains
        high_priority_domains = {
            "LDE", "CORE", "INSTALLER", "SEC", "AUTH", "TEST", "VALIDATOR",
            "GIT", "PROJECT", "QUALITY", "PHASE", "INIT", "UPDATE",
            "TEMPLATE", "CONFIG", "LANGUAGE", "DETECT", "NETWORK",
            "VERSION", "CACHE", "HOOK", "CLI", "BACKUP"
        }

        # Process broken chains
        for detail in result.broken_chain_details:
            domain = detail["domain"]
            missing = detail["missing"]
            number = self._extract_number_from_domain(domain)

            for element in missing:
                if element == "SPEC":
                    task = RepairTask(
                        domain=domain,
                        number=number,
                        action="create_spec",
                        priority="high" if domain in high_priority_domains else "medium",
                        estimated_effort="medium"
                    )
                elif element == "CODE":
                    task = RepairTask(
                        domain=domain,
                        number=number,
                        action="create_code",
                        priority="high" if domain in high_priority_domains else "medium",
                        estimated_effort="medium"
                    )
                elif element == "TEST":
                    task = RepairTask(
                        domain=domain,
                        number=number,
                        action="create_test",
                        priority="high" if domain in high_priority_domains else "medium",
                        estimated_effort="low"
                    )
                else:
                    continue

                tasks.append(task)
                summary["total_tasks"] += 1
                summary[f"{task.priority}_priority"] += 1

        # Process orphans
        for orphan_type, orphans in result.orphans_by_type.items():
            for tag in orphans:
                domain = self._extract_domain_from_tag(tag)
                number = self._extract_number_from_domain(domain)

                if orphan_type == "code_without_spec":
                    task = RepairTask(
                        domain=domain,
                        number=number,
                        action="create_spec",
                        priority="medium",
                        estimated_effort="low"
                    )
                elif orphan_type == "code_without_test":
                    task = RepairTask(
                        domain=domain,
                        number=number,
                        action="create_test",
                        priority="medium",
                        estimated_effort="low"
                    )
                elif orphan_type == "test_without_code":
                    task = RepairTask(
                        domain=domain,
                        number=number,
                        action="create_code",
                        priority="high",
                        estimated_effort="medium"
                    )
                elif orphan_type == "spec_without_code":
                    task = RepairTask(
                        domain=domain,
                        number=number,
                        action="create_code",
                        priority="high",
                        estimated_effort="medium"
                    )
                else:
                    continue

                tasks.append(task)
                summary["total_tasks"] += 1
                summary[f"{task.priority}_priority"] += 1

        # Sort tasks by priority and domain
        tasks.sort(key=lambda x: (
            {"high": 0, "medium": 1, "low": 2}[x.priority],
            x.domain
        ))

        # Organize by priority
        high_priority = [t for t in tasks if t.priority == "high"]
        medium_priority = [t for t in tasks if t.priority == "medium"]
        low_priority = [t for t in tasks if t.priority == "low"]

        return RepairPlan(
            high_priority_tasks=high_priority,
            medium_priority_tasks=medium_priority,
            low_priority_tasks=low_priority,
            summary=summary
        )

    def _extract_number_from_domain(self, domain: str) -> int:
        """Extract number from domain string."""
        # Look for numbers in domain (e.g., "LDE-PRIORITY-001" -> 1)
        match = re.search(r'(\d+)$', domain)
        if match:
            return int(match.group(1))
        return 1

    def _extract_domain_from_tag(self, tag: str) -> str:
        """Extract domain from TAG."""
        match = re.match(r'@[A-Z]+:([A-Z0-9-]+)\d{3}', tag)
        if match:
            return match.group(1)
        return tag.split(":")[1].rsplit("-", 1)[0]

    def _create_spec_template(self, domain: str, number: int) -> str:
        """Create SPEC template for given domain and number."""
        spec_id = f"@SPEC:{domain}-{number:03d}"

        # Domain-specific templates
        if "LDE" in domain:
            return f"""---
id: {domain.lower().replace('-', '-')}-{number}
version: 0.1.0
status: draft
created: 2025-11-05
author: @Goos
priority: high
category: feature
labels:
  - lde
  - development
  - enhancement
---

# {spec_id}

## HISTORY

### v0.1.0 (2025-11-05)
- **INITIAL**: LDE 기능 개선
- **AUTHOR**: @Goos
- **SCOPE**: 개발 경험 및 효율성 향상

---

## Environment (환경)

### 현재 상태

### 가정

### 요구사항

### Ubiquitous Requirements (기본 요구사항)

### Event-driven Requirements (이벤트 기반)

### State-driven Requirements (상태 기반)

### Constraints (제약사항)

---

## Traceability (@TAG)

### 핵심 TAG 체인

- **SPEC**: {spec_id}
- **CODE**: @CODE:{domain}-{number:03d}
- **TEST**: @TEST:{domain}-{number:03d}

---

## Specifications (상세 명세)

### 1. 기능 명세

### 2. 구현 요구사항

### 3. 테스트 요구사항

---

## 성공 지표

---

## 리스크 및 �난화 방안

---

## 다음 단계

**작성일**: 2025-11-05
**버전**: 0.1.0
**상태**: Draft
"""
        elif "CORE" in domain:
            return f"""---
id: core-{domain.lower().replace('-', '-')}-{number}
version: 0.1.0
status: draft
created: 2025-11-05
author: @Goos
priority: high
category: core
labels:
  - core
  - project
  - infrastructure
---

# {spec_id}

## HISTORY

### v0.1.0 (2025-11-05)
- **INITIAL**: 핵심 기능 개선
- **AUTHOR**: @Goos
- **SCOPE**: 프로젝트 핵심 인프라 개선

---

## Environment (환경)

### 현재 상태

### 가정

### 요구사항

### Ubiquitous Requirements (기본 요구사항)

### Event-driven Requirements (이벤트 기반)

### State-driven Requirements (상태 기반)

### Constraints (제약사항)

---

## Traceability (@TAG)

### 핵심 TAG 체인

- **SPEC**: {spec_id}
- **CODE**: @CODE:{domain}-{number:03d}
- **TEST**: @TEST:{domain}-{number:03d}

---

## Specifications (상세 명세)

### 1. 기능 명세

### 2. 구현 요구사항

### 3. 테스트 요구사항

---

## 성공 지표

---

## 리스크 및 완화 방안

---

## 다음 단계

**작성일**: 2025-11-05
**버전**: 0.1.0
**상태**: Draft
"""
        elif "INSTALLER" in domain:
            return f"""---
id: installer-{domain.lower().replace('-', '-')}-{number}
version: 0.1.0
status: draft
created: 2025-11-05
author: @Goos
priority: high
category: installer
labels:
  - installer
  - setup
  - deployment
---

# {spec_id}

## HISTORY

### v0.1.0 (2025-11-05)
- **INITIAL**: 설치기 개선
- **AUTHOR**: @Goos
- **SCOPE**: 설치 및 배포 프로세스 개선

---

## Environment (환경)

### 현재 상태

### 가정

### 요구사항

### Ubiquitous Requirements (기본 요구사항)

### Event-driven Requirements (이벤트 기반)

### State-driven Requirements (상태 기반)

### Constraints (제약사항)

---

## Traceability (@TAG)

### 핵심 TAG 체인

- **SPEC**: {spec_id}
- **CODE**: @CODE:{domain}-{number:03d}
- **TEST**: @TEST:{domain}-{number:03d}

---

## Specifications (상세 명세)

### 1. 설치 명세

### 2. 배포 요구사항

### 3. 테스트 요구사항

---

## 성공 지표

---

## 리스크 및 완화 방안

---

## 다음 단계

**작성일**: 2025-11-05
**버전**: 0.1.0
**상태**: Draft
"""
        else:
            # Generic template
            return f"""---
id: {domain.lower().replace('-', '-')}-{number}
version: 0.1.0
status: draft
created: 2025-11-05
author: @Goos
priority: medium
category: feature
labels:
  - {domain.lower()}
---

# {spec_id}

## HISTORY

### v0.1.0 (2025-11-05)
- **INITIAL**: {domain} 도메인 기능 개선
- **AUTHOR**: @Goos
- **SCOPE**: {domain} 기능 개선 및 유지보수성 향상

---

## Environment (환경)

### 현재 상태

### 가정

### 요구사항

### Ubiquitous Requirements (기본 요구사항)

### Event-driven Requirements (이벤트 기반)

### State-driven Requirements (상태 기반)

### Constraints (제약사항)

---

## Traceability (@TAG)

### 핵심 TAG 체인

- **SPEC**: {spec_id}
- **CODE**: @CODE:{domain}-{number:03d}
- **TEST**: @TEST:{domain}-{number:03d}

---

## Specifications (상세 명세)

### 1. 기능 명세

### 2. 구현 요구사항

### 3. 테스트 요구사항

---

## 성공 지표

---

## 리스크 및 완화 방안

---

## 다음 단계

**작성일**: 2025-11-05
**버전**: 0.1.0
**상태**: Draft
"""

    def _create_code_template(self, domain: str, number: int) -> str:
        """Create CODE template for given domain and number."""
        code_id = f"@CODE:{domain}-{number:03d}"

        # Determine module based on domain
        if "LDE" in domain:
            module = "lde"
        elif "CORE" in domain:
            module = "core"
        elif "INSTALLER" in domain:
            module = "installer"
        elif "CLI" in domain:
            module = "cli"
        elif "UTILS" in domain:
            module = "utils"
        else:
            module = "core"

        return f'''# {code_id}
"""{domain} 기능 구현.

{domain} 도메인의 핵심 기능을 구현합니다.

@SPEC:{domain}-{number:03d}: {domain} 기능 명세
"""

from typing import Any, Dict, List, Optional
from pathlib import Path


def {domain.lower().replace('-', '_')}_function(
    param1: Optional[str] = None,
    param2: Optional[Dict[str, Any]] = None
) -> Any:
    """{domain} 핵심 기능.

    Args:
        param1: 첫 번째 매개변수
        param2: 두 번째 매개변수 (선택적)

    Returns:
        처리 결과 객체

    Examples:
        >>> result = {domain.lower().replace('-', '_')}_function("test", {{"key": "value"}})
        >>> print(result)
        "processed_result"
    """
    if param1 is None:
        param1 = "default_value"

    if param2 is None:
        param2 = {{}}

    # 핵심 로직 구현
    result = _process_{domain.lower().replace('-', '_')}_logic(param1, param2)

    return result


def _process_{domain.lower().replace('-', '_')}_logic(
    input_data: str,
    config: Dict[str, Any]
) -> Any:
    """내부 로직 처리.

    Args:
        input_data: 입력 데이터
        config: 설정 정보

    Returns:
        처리된 결과
    """
    # 여기에 {domain} 특화 로직 구현
    processed_data = input_data.upper() if config.get("uppercase", False) else input_data

    return {{
        "status": "success",
        "input": input_data,
        "output": processed_data,
        "config": config
    }}


def validate_{domain.lower().replace('-', '_')}_input(data: Any) -> bool:
    """입력 데이터 검증.

    Args:
        data: 검증할 데이터

    Returns:
        유효성 검증 결과
    """
    if data is None:
        return False

    # 여기에 {domain} 특화 검증 로직 추가
    return isinstance(data, (str, dict, list))


# @TEST:{domain}-{number:03d}: 단위 테스트 포함
if __name__ == "__main__":
    # 간단한 실행 예시
    test_result = {domain.lower().replace('-', '_')}_function("test")
    print(f"Test result: {{test_result}}")
'''

    def _create_test_template(self, domain: str, number: int) -> str:
        """Create TEST template for given domain and number."""
        test_id = f"@TEST:{domain}-{number:03d}"

        # Determine test type based on domain
        if "LDE" in domain:
            test_type = "feature"
        elif "CORE" in domain:
            test_type = "unit"
        elif "INSTALLER" in domain:
            test_type = "integration"
        elif "CLI" in domain:
            test_type = "e2e"
        else:
            test_type = "unit"

        return f'''# {test_id}
"""{domain} 기능 테스트.

{domain} 도메인의 핵심 기능을 검증하기 위한 테스트 코드입니다.

@SPEC:{domain}-{number:03d}: {domain} 기능 명세
@CODE:{domain}-{number:03d}: {domain} 기능 구현
"""

import pytest
from pathlib import Path
from unittest.mock import Mock, patch

from src.moai_adk.{domain.lower().replace('-', '_')}.{domain.lower().replace('-', '_')} import (
    {domain.lower().replace('-', '_')}_function,
    validate_{domain.lower().replace('-', '_')}_input,
)


class Test{domain.replace('-', '_')}:
    """{domain} 기능 테스트 클래스."""

    def test_{domain.lower().replace('-', '_')}_function_basic(self):
        """기본 {domain} 기능 테스트."""
        result = {domain.lower().replace('-', '_')}_function("test_input")

        assert result["status"] == "success"
        assert result["input"] == "test_input"
        assert result["config"] == {{}}

    def test_{domain.lower().replace('-', '_')}_function_with_params(self):
        """매개변수가 포함된 {domain} 기능 테스트."""
        param2 = {{"key": "value", "uppercase": True}}
        result = {domain.lower().replace('-', '_')}_function("test", param2)

        assert result["status"] == "success"
        assert result["input"] == "test"
        assert result["output"] == "TEST"  # uppercase=True 적용
        assert result["config"] == param2

    def test_{domain.lower().replace('-', '_')}_function_defaults(self):
        """기본값 사용 테스트."""
        result = {domain.lower().replace('-', '_')}_function()

        assert result["status"] == "success"
        assert result["input"] == "default_value"

    def test_validate_{domain.lower().replace('-', '_')}_input_valid(self):
        """유효한 입력 데이터 검증 테스트."""
        assert validate_{domain.lower().replace('-', '_')}_input("valid_string") is True
        assert validate_{domain.lower().replace('-', '_')}_input({{"key": "value"}}) is True
        assert validate_{domain.lower().replace('-', '_')}_input([1, 2, 3]) is True

    def test_validate_{domain.lower().replace('-', '_')}_input_invalid(self):
        """무효한 입력 데이터 검증 테스트."""
        assert validate_{domain.lower().replace('-', '_')}_input(None) is False
        assert validate_{domain.lower().replace('-', '_')}_input(123) is False
        assert validate_{domain.lower().replace('-', '_')}_input(12.34) is False

    def test_{domain.lower().replace('-', '_')}_function_error_handling(self):
        """에러 처리 테스트."""
        with pytest.raises(Exception):
            # 예상되는 예외 상황 테스트
            {domain.lower().replace('-', '_')}_function(None, None)

    @patch('src.moai_adk.{domain.lower().replace('-', '_')}.{domain.lower().replace('-', '_')}._process_{domain.lower().replace('-', '_')}_logic')
    def test_{domain.lower().replace('-', '_')}_function_integration(self, mock_process):
        """통합 테스트."""
        mock_process.return_value = {{"mock": "result"}}

        result = {domain.lower().replace('-', '_')}_function("test_input")

        mock_process.assert_called_once_with("test_input", {{}})
        assert result == {{"mock": "result"}}


def test_{domain.lower().replace('-', '_')}_function_edge_cases():
    """경계 조건 테스트."""
    # 빈 문자열 테스트
    result = {domain.lower().replace('-', '_')}_function("")
    assert result["status"] == "success"
    assert result["input"] == ""

    # 빈 딕셔너리 테스트
    result = {domain.lower().replace('-', '_')}_function("test", {{}})
    assert result["status"] == "success"
    assert result["input"] == "test"

    # 대규모 데이터 테스트
        large_data = "x" * 10000
    result = {domain.lower().replace('-', '_')}_function(large_data)
    assert result["status"] == "success"
    assert len(result["output"]) == 10000


# @TEST:{domain}-{number:03d}: 테스트 실행
if __name__ == "__main__":
    pytest.main([__file__, "-v"])
'''

    def execute_repair_plan(self, plan: RepairPlan, dry_run: bool = True) -> Dict[str, List[str]]:
        """Execute repair plan (dry run by default)."""
        results = {"created": [], "skipped": [], "errors": []}

        for priority, tasks in plan.get_tasks_by_priority():
            for task in tasks:
                try:
                    if dry_run:
                        result = f"[DRY RUN] Would create: {task.action} for {task.domain}-{task.number:03d}"
                        results["skipped"].append(result)
                    else:
                        created_files = self._execute_repair_task(task)
                        results["created"].extend(created_files)
                except Exception as e:
                    error_msg = f"Error repairing {task.domain}-{task.number:03d}: {str(e)}"
                    results["errors"].append(error_msg)

        return results

    def _execute_repair_task(self, task: RepairTask) -> List[str]:
        """Execute a single repair task."""
        created_files = []

        if task.action == "create_spec":
            spec_content = self._create_spec_template(task.domain, task.number)
            spec_path = self.root_path / ".moai" / "specs" / f"spec-{task.domain.lower()}-{task.number:03d}.md"
            spec_path.parent.mkdir(parents=True, exist_ok=True)
            spec_path.write_text(spec_content, encoding='utf-8')
            created_files.append(str(spec_path))

        elif task.action == "create_code":
            code_content = self._create_code_template(task.domain, task.number)
            # Determine appropriate module path
            if "LDE" in task.domain:
                code_path = self.root_path / "src" / "moai_adk" / "lde" / f"{task.domain.lower()}.py"
            elif "CORE" in task.domain:
                code_path = self.root_path / "src" / "moai_adk" / "core" / f"{task.domain.lower()}.py"
            elif "UTILS" in task.domain:
                code_path = self.root_path / "src" / "moai_adk" / "utils" / f"{task.domain.lower()}.py"
            elif "CLI" in task.domain:
                code_path = self.root_path / "src" / "moai_adk" / "cli" / "commands" / f"{task.domain.lower()}.py"
            else:
                code_path = self.root_path / "src" / "moai_adk" / "core" / f"{task.domain.lower()}.py"

            code_path.parent.mkdir(parents=True, exist_ok=True)
            code_path.write_text(code_content, encoding='utf-8')
            created_files.append(str(code_path))

        elif task.action == "create_test":
            test_content = self._create_test_template(task.domain, task.number)
            # Determine appropriate test path
            if "LDE" in task.domain:
                test_path = self.root_path / "tests" / "unit" / "lde" / f"test_{task.domain.lower()}.py"
            elif "CORE" in task.domain:
                test_path = self.root_path / "tests" / "unit" / "core" / f"test_{task.domain.lower()}.py"
            elif "UTILS" in task.domain:
                test_path = self.root_path / "tests" / "unit" / "utils" / f"test_{task.domain.lower()}.py"
            elif "CLI" in task.domain:
                test_path = self.root_path / "tests" / "unit" / "cli" / f"test_{task.domain.lower()}.py"
            else:
                test_path = self.root_path / "tests" / "unit" / f"test_{task.domain.lower()}.py"

            test_path.parent.mkdir(parents=True, exist_ok=True)
            test_path.write_text(test_content, encoding='utf-8')
            created_files.append(str(test_path))

        return created_files


def repair_tag_chains(root_path: Path = Path("."), dry_run: bool = True) -> Tuple[ChainAnalysisResult, RepairPlan, Dict[str, List[str]]]:
    """Convenience function to repair TAG chains."""
    repairer = TagChainRepairer(root_path)
    result, plan = repairer.analyze_and_create_plan()
    execution_results = repairer.execute_repair_plan(plan, dry_run)
    return result, plan, execution_results


def main():
    """Main function for CLI usage."""
    import argparse

    parser = argparse.ArgumentParser(description="Repair TAG chains in MoAI-ADK")
    parser.add_argument("--path", default=".", help="Path to analyze (default: current directory)")
    parser.add_argument("--execute", action="store_true", help="Execute repairs (default: dry run)")
    parser.add_argument("--high-priority-only", action="store_true", help="Only repair high priority items")
    parser.add_argument("--output", help="Output file for JSON report")

    args = parser.parse_args()

    result, plan, execution_results = repair_tag_chains(Path(args.path), not args.execute)

    if args.output:
        import json
        with open(args.output, 'w') as f:
            json.dump({
                "analysis": {
                    "total_chains": result.total_chains,
                    "complete_chains": result.complete_chains,
                    "partial_chains": result.partial_chains,
                    "broken_chains": result.broken_chains
                },
                "plan": {
                    "summary": plan.summary,
                    "high_priority_count": len(plan.high_priority_tasks),
                    "medium_priority_count": len(plan.medium_priority_tasks),
                    "low_priority_count": len(plan.low_priority_tasks)
                },
                "execution": execution_results
            }, f, indent=2)
    else:
        print("=== TAG Chain Repair Results ===")
        print(f"Analysis: {result.total_chains} total chains, {result.broken_chains} broken ({result.complete_chains} complete)")
        print(f"Plan: {plan.summary['total_tasks']} tasks ({plan.summary['high_priority_priority']} high, {plan.summary['medium_priority_priority']} medium, {plan.summary['low_priority_priority']} low priority)")

        if args.high_priority_only:
            print("=== High Priority Tasks ===")
            for task in plan.high_priority_tasks:
                print(f"- {task.action.upper()} for {task.domain}-{task.number:03d} ({task.priority} priority)")
        else:
            print("=== All Tasks ===")
            for priority, tasks in plan.get_tasks_by_priority():
                print(f"\n{priority.upper()} Priority ({len(tasks)} tasks):")
                for task in tasks:
                    print(f"- {task.action.upper()} for {task.domain}-{task.number:03d}")

        print("\n=== Execution Results ===")
        print(f"Created: {len(execution_results['created'])} files")
        print(f"Skipped (dry run): {len(execution_results['skipped'])}")
        print(f"Errors: {len(execution_results['errors'])}")

        if execution_results['errors']:
            print("\nErrors:")
            for error in execution_results['errors']:
                print(f"- {error}")


if __name__ == "__main__":
    main()
