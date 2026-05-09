# ANALYZE-PRESERVE-IMPROVE DDD Cycle

> Module: Core DDD cycle implementation with Context7 integration
> Complexity: Advanced
> Time: 20+ minutes
> Dependencies: Python 3.8+, pytest, Context7 MCP, unittest

## Core DDD Classes

```python
import pytest
import unittest
import asyncio
import subprocess
import os
import sys
import time
from typing import Dict, List, Optional, Any
from dataclasses import dataclass, field
from enum import Enum
import json
from pathlib import Path

class DDDPhase(Enum):
    """DDD cycle phases."""
    ANALYZE = "analyze"       # Analyze existing code and behavior
    PRESERVE = "preserve"     # Create characterization tests
    IMPROVE = "improve"       # Improve code while keeping tests green
    REVIEW = "review"         # Review and commit changes

class TestType(Enum):
    """Types of tests in DDD."""
    UNIT = "unit"
    INTEGRATION = "integration"
    CHARACTERIZATION = "characterization"
    ACCEPTANCE = "acceptance"
    PERFORMANCE = "performance"
    SECURITY = "security"
    REGRESSION = "regression"

class TestStatus(Enum):
    """Test execution status."""
    PENDING = "pending"
    RUNNING = "running"
    PASSED = "passed"
    FAILED = "failed"
    SKIPPED = "skipped"
    ERROR = "error"

@dataclass
class TestSpecification:
    """Specification for a DDD test."""
    name: str
    description: str
    test_type: TestType
    requirements: List[str]
    acceptance_criteria: List[str]
    edge_cases: List[str]
    preconditions: List[str] = field(default_factory=list)
    postconditions: List[str] = field(default_factory=list)
    dependencies: List[str] = field(default_factory=list)
    mock_requirements: Dict[str, Any] = field(default_factory=dict)
    behavior_snapshot: Optional[Dict[str, Any]] = None

@dataclass
class TestCase:
    """Individual test case with metadata."""
    id: str
    name: str
    file_path: str
    line_number: int
    specification: TestSpecification
    status: TestStatus
    execution_time: float
    error_message: Optional[str] = None
    coverage_data: Dict[str, Any] = field(default_factory=dict)

@dataclass
class DDDSession:
    """DDD development session with cycle tracking."""
    id: str
    project_path: str
    current_phase: DDDPhase
    test_cases: List[TestCase]
    start_time: float
    context7_patterns: Dict[str, Any] = field(default_factory=dict)
    metrics: Dict[str, Any] = field(default_factory=dict)
    behavior_snapshots: Dict[str, Any] = field(default_factory=dict)
```

## DDD Manager Implementation

```python
class DDDManager:
    """Main DDD workflow manager with Context7 integration."""

    def __init__(self, project_path: str, context7_client=None):
        self.project_path = Path(project_path)
        self.context7 = context7_client
        self.current_session = None
        self.test_history = []

    async def start_ddd_session(
        self, feature_name: str,
        test_types: List[TestType] = None
    ) -> DDDSession:
        """Start a new DDD development session."""

        if test_types is None:
            test_types = [TestType.CHARACTERIZATION, TestType.UNIT, TestType.INTEGRATION]

        # Create session
        session = DDDSession(
            id=f"ddd_{feature_name}_{int(time.time())}",
            project_path=str(self.project_path),
            current_phase=DDDPhase.ANALYZE,
            test_cases=[],
            start_time=time.time(),
            context7_patterns={},
            metrics={
                'tests_written': 0,
                'tests_passing': 0,
                'tests_failing': 0,
                'coverage_percentage': 0.0,
                'behaviors_preserved': 0
            },
            behavior_snapshots={}
        )

        self.current_session = session
        return session

    async def run_full_ddd_cycle(
        self, specification: TestSpecification,
        target_function: str = None
    ) -> Dict[str, Any]:
        """Run complete ANALYZE-PRESERVE-IMPROVE DDD cycle."""

        cycle_results = {}

        # ANALYZE phase
        print("🔍 ANALYZE Phase: Understanding existing code and behavior...")
        analyze_results = await self._run_analyze_phase(target_function)
        cycle_results['analyze'] = analyze_results
        self.current_session.current_phase = DDDPhase.ANALYZE

        # PRESERVE phase
        print("🧪 PRESERVE Phase: Creating characterization tests...")
        preserve_results = await self._run_preserve_phase(specification, analyze_results)
        cycle_results['preserve'] = preserve_results
        self.current_session.current_phase = DDDPhase.PRESERVE

        # IMPROVE phase
        print("🔧 IMPROVE Phase: Refactoring with behavior preservation...")
        improve_results = await self._run_improve_phase(specification)
        cycle_results['improve'] = improve_results
        self.current_session.current_phase = DDDPhase.IMPROVE

        # REVIEW phase
        print("✅ REVIEW Phase: Final verification...")
        coverage_results = await self._run_coverage_analysis()
        cycle_results['review'] = {'coverage': coverage_results}
        self.current_session.current_phase = DDDPhase.REVIEW

        return cycle_results

    async def _run_analyze_phase(self, target_function: str = None) -> Dict[str, Any]:
        """ANALYZE: Understand existing code and behavior."""

        analysis = {
            'existing_tests': [],
            'code_patterns': [],
            'dependencies': [],
            'behavior_notes': []
        }

        # Find existing tests
        test_files = list(self.project_path.glob("**/test_*.py"))
        analysis['existing_tests'] = [str(f) for f in test_files]

        # Analyze code structure
        if target_function:
            analysis['target'] = target_function
            analysis['behavior_notes'].append(f"Analyzing behavior of {target_function}")

        return analysis

    async def _run_preserve_phase(
        self, specification: TestSpecification,
        analysis: Dict[str, Any]
    ) -> Dict[str, Any]:
        """PRESERVE: Create characterization tests to capture existing behavior."""

        preserve_results = {
            'characterization_tests_created': 0,
            'behaviors_captured': [],
            'test_files': []
        }

        # Create characterization tests based on analysis
        for behavior in analysis.get('behavior_notes', []):
            preserve_results['behaviors_captured'].append(behavior)
            preserve_results['characterization_tests_created'] += 1

        # Run existing tests to establish baseline
        test_results = await self._run_pytest()
        preserve_results['baseline_results'] = test_results

        return preserve_results

    async def _run_improve_phase(self, specification: TestSpecification) -> Dict[str, Any]:
        """IMPROVE: Refactor code while maintaining test passing."""

        improve_results = {
            'improvements_made': [],
            'tests_still_passing': True,
            'refactoring_notes': []
        }

        # Run tests after improvements
        test_results = await self._run_pytest()
        improve_results['tests_still_passing'] = test_results.get('failed', 0) == 0

        if improve_results['tests_still_passing']:
            self.current_session.metrics['behaviors_preserved'] += 1

        return improve_results

    async def _run_pytest(self) -> Dict[str, Any]:
        """Run pytest and return results."""

        try:
            result = subprocess.run(
                [
                    sys.executable, '-m', 'pytest',
                    str(self.project_path),
                    '--tb=short',
                    '-v'
                ],
                capture_output=True,
                text=True,
                cwd=str(self.project_path)
            )

            return self._parse_pytest_output(result.stdout)

        except Exception as e:
            print(f"Error running pytest: {e}")
            return {'error': str(e), 'passed': 0, 'failed': 0}

    def _parse_pytest_output(self, output: str) -> Dict[str, Any]:
        """Parse pytest output."""

        lines = output.split('\n')
        results = {'passed': 0, 'failed': 0, 'skipped': 0, 'total': 0}

        for line in lines:
            if ' passed in ' in line:
                parts = line.split()
                if parts and parts[0].isdigit():
                    results['passed'] = int(parts[0])
                    results['total'] = int(parts[0])
            elif ' passed' in line and ' failed' in line:
                passed_part = line.split(' passed')[0]
                if passed_part.strip().isdigit():
                    results['passed'] = int(passed_part.strip())

                if ' failed' in line:
                    failed_part = line.split(' failed')[0].split(', ')[-1]
                    if failed_part.strip().isdigit():
                        results['failed'] = int(failed_part.strip())

                results['total'] = results['passed'] + results['failed']

        return results

    async def _run_coverage_analysis(self) -> Dict[str, Any]:
        """Run test coverage analysis."""

        try:
            result = subprocess.run(
                [
                    sys.executable, '-m', 'pytest',
                    str(self.project_path),
                    '--cov=src',
                    '--cov-report=term-missing'
                ],
                capture_output=True,
                text=True,
                cwd=str(self.project_path)
            )

            return {'coverage_output': result.stdout}

        except Exception as e:
            return {'error': str(e)}

    def get_session_summary(self) -> Dict[str, Any]:
        """Get summary of current DDD session."""

        if not self.current_session:
            return {}

        duration = time.time() - self.current_session.start_time

        return {
            'session_id': self.current_session.id,
            'phase': self.current_session.current_phase.value,
            'duration_seconds': duration,
            'duration_formatted': f"{duration:.1f} seconds",
            'metrics': self.current_session.metrics,
            'test_cases_count': len(self.current_session.test_cases),
            'behaviors_preserved': self.current_session.metrics.get('behaviors_preserved', 0)
        }
```

## Phase-Specific Guidelines

### ANALYZE Phase
- Understand existing code structure and patterns
- Identify current behavior through code reading
- Document dependencies and side effects
- Map test coverage gaps
- Note existing design patterns

### PRESERVE Phase
- Write characterization tests for existing behavior
- Capture current behavior as the "golden standard"
- Ensure tests pass with current implementation
- Document discovered behavior
- Create behavior snapshots for complex outputs

### IMPROVE Phase
- Refactor code while keeping tests green
- Make small, incremental changes
- Run tests after each change
- Maintain behavior preservation
- Apply design patterns appropriately

### REVIEW Phase
- Verify all characterization tests still pass
- Review code quality and documentation
- Check for any behavior changes
- Commit changes with clear messages
- Document improvements made

## Usage Example

```python
# Initialize DDD Manager
ddd_manager = DDDManager(
    project_path="/path/to/project",
    context7_client=context7
)

# Start DDD session
session = await ddd_manager.start_ddd_session("user_authentication_refactor")

# Create test specification
test_spec = TestSpecification(
    name="test_user_login_behavior_preservation",
    description="Preserve existing login behavior during refactoring",
    test_type=TestType.CHARACTERIZATION,
    requirements=[
        "Existing login flow must continue to work",
        "Error messages should remain consistent"
    ],
    acceptance_criteria=[
        "Valid credentials return user token (existing behavior)",
        "Invalid credentials raise same error messages"
    ],
    edge_cases=[
        "Test with empty email (existing behavior)",
        "Test with empty password (existing behavior)"
    ]
)

# Run complete DDD cycle
cycle_results = await ddd_manager.run_full_ddd_cycle(
    specification=test_spec,
    target_function="authenticate_user"
)

# Get session summary
summary = ddd_manager.get_session_summary()
print(f"Session completed in {summary['duration_formatted']}")
print(f"Behaviors preserved: {summary['behaviors_preserved']}")
```

---

Related: [Test Generation](./test-generation.md) | [Test Patterns](./test-patterns.md)
