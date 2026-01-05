# RED-GREEN-REFACTOR TDD Cycle

> Module: Core TDD cycle implementation with Context7 integration
> Complexity: Advanced
> Time: 20+ minutes
> Dependencies: Python 3.8+, pytest, Context7 MCP, unittest

## Core TDD Classes

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

class TDDPhase(Enum):
    """TDD cycle phases."""
    RED = "red" # Write failing test
    GREEN = "green" # Make test pass
    REFACTOR = "refactor" # Improve code while keeping tests green
    REVIEW = "review" # Review and commit changes

class TestType(Enum):
    """Types of tests in TDD."""
    UNIT = "unit"
    INTEGRATION = "integration"
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
    """Specification for a TDD test."""
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
class TDDSession:
    """TDD development session with cycle tracking."""
    id: str
    project_path: str
    current_phase: TDDPhase
    test_cases: List[TestCase]
    start_time: float
    context7_patterns: Dict[str, Any] = field(default_factory=dict)
    metrics: Dict[str, Any] = field(default_factory=dict)
```

## TDD Manager Implementation

```python
class TDDManager:
    """Main TDD workflow manager with Context7 integration."""

    def __init__(self, project_path: str, context7_client=None):
        self.project_path = Path(project_path)
        self.context7 = context7_client
        self.current_session = None
        self.test_history = []

    async def start_tdd_session(
        self, feature_name: str,
        test_types: List[TestType] = None
    ) -> TDDSession:
        """Start a new TDD development session."""

        if test_types is None:
            test_types = [TestType.UNIT, TestType.INTEGRATION]

        # Create session
        session = TDDSession(
            id=f"tdd_{feature_name}_{int(time.time())}",
            project_path=str(self.project_path),
            current_phase=TDDPhase.RED,
            test_cases=[],
            start_time=time.time(),
            context7_patterns={},
            metrics={
                'tests_written': 0,
                'tests_passing': 0,
                'tests_failing': 0,
                'coverage_percentage': 0.0
            }
        )

        self.current_session = session
        return session

    async def run_full_tdd_cycle(
        self, specification: TestSpecification,
        target_function: str = None
    ) -> Dict[str, Any]:
        """Run complete RED-GREEN-REFACTOR TDD cycle."""

        cycle_results = {}

        # RED phase
        print(" RED Phase: Writing failing test...")
        # RED phase implementation here
        cycle_results['red'] = {'phase': 'RED'}

        # GREEN phase
        print(" GREEN Phase: Implementing minimum code...")
        # GREEN phase implementation here
        cycle_results['green'] = {'phase': 'GREEN'}

        # REFACTOR phase
        print(" REFACTOR Phase: Improving code quality...")
        # REFACTOR phase implementation here
        cycle_results['refactor'] = {'phase': 'REFACTOR'}

        # REVIEW phase
        print(" REVIEW Phase: Final verification...")
        coverage_results = await self._run_coverage_analysis()
        cycle_results['review'] = {'coverage': coverage_results}

        self.current_session.current_phase = TDDPhase.REVIEW
        return cycle_results

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
        """Get summary of current TDD session."""

        if not self.current_session:
            return {}

        duration = time.time() - self.current_session.start_time

        return {
            'session_id': self.current_session.id,
            'phase': self.current_session.current_phase.value,
            'duration_seconds': duration,
            'duration_formatted': f"{duration:.1f} seconds",
            'metrics': self.current_session.metrics,
            'test_cases_count': len(self.current_session.test_cases)
        }
```

## Phase-Specific Guidelines

### RED Phase
- Write the simplest possible failing test
- Test one specific behavior or requirement
- Ensure test clearly expresses intent
- Make test fail for the right reason

### GREEN Phase
- Write the simplest code to make test pass
- Don't worry about code quality yet
- Focus on making the test green quickly
- Avoid premature optimization

### REFACTOR Phase
- Improve code design while keeping tests green
- Remove duplication and improve readability
- Apply design patterns appropriately
- Ensure all tests still pass

### REVIEW Phase
- Verify test coverage meets requirements
- Review code quality and documentation
- Check for any remaining technical debt
- Commit changes with clear messages

## Usage Example

```python
# Initialize TDD Manager
tdd_manager = TDDManager(
    project_path="/path/to/project",
    context7_client=context7
)

# Start TDD session
session = await tdd_manager.start_tdd_session("user_authentication")

# Create test specification
test_spec = TestSpecification(
    name="test_user_login_valid_credentials",
    description="Test that user can login with valid credentials",
    test_type=TestType.UNIT,
    requirements=[
        "User must provide valid email and password",
        "System should authenticate user credentials"
    ],
    acceptance_criteria=[
        "Valid credentials return user token",
        "Invalid credentials raise AuthenticationError"
    ],
    edge_cases=[
        "Test with empty email",
        "Test with empty password"
    ]
)

# Run complete TDD cycle
cycle_results = await tdd_manager.run_full_tdd_cycle(
    specification=test_spec,
    target_function="authenticate_user"
)

# Get session summary
summary = tdd_manager.get_session_summary()
print(f"Session completed in {summary['duration_formatted']}")
```

---

Related: [Test Generation](./test-generation.md) | [Test Patterns](./test-patterns.md)
