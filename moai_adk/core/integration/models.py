"""
Integration testing data structures and utilities.

This module contains common data structures and utility functions
for integration testing across the MoAI-ADK system.
"""

from dataclasses import dataclass
from enum import Enum
from typing import List, Optional


class TestStatus(Enum):
    """Test status enumeration"""

    PENDING = "pending"
    RUNNING = "running"
    PASSED = "passed"
    FAILED = "failed"
    SKIPPED = "skipped"


@dataclass
class IntegrationTestResult:
    """Test result data structure"""

    test_name: str
    passed: bool
    error_message: Optional[str] = None
    execution_time: float = 0.0
    components_tested: Optional[List[str]] = None
    status: Optional[TestStatus] = None

    def __post_init__(self):
        if self.components_tested is None:
            self.components_tested = []
        if self.passed:
            self.status = TestStatus.PASSED
        elif self.error_message:
            self.status = TestStatus.FAILED
        else:
            self.status = TestStatus.PENDING


@dataclass
class TestComponent:
    """Test component definition"""

    name: str
    component_type: str
    version: str
    dependencies: Optional[List[str]] = None

    def __post_init__(self):
        if self.dependencies is None:
            self.dependencies = []


@dataclass
class TestSuite:
    """Test suite definition"""

    name: str
    description: str
    components: List[TestComponent]
    test_cases: Optional[List[str]] = None

    def __post_init__(self):
        if self.test_cases is None:
            self.test_cases = []


class IntegrationTestError(Exception):
    """Base exception for integration testing"""

    pass


class TestTimeoutError(IntegrationTestError):
    """Test timeout exception"""

    pass


class ComponentNotFoundError(IntegrationTestError):
    """Component not found exception"""

    pass
