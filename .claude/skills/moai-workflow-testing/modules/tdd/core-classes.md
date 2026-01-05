# TDD with Context7 - Core Classes

> Sub-module: Core class implementations for TDD workflow management
> Parent: [TDD with Context7](../tdd-context7.md)

## Enumerations

### TDDPhase Enum

```python
class TDDPhase(Enum):
    """Phases of the TDD cycle."""
    RED = "red"        # Writing failing test
    GREEN = "green"    # Making test pass
    REFACTOR = "refactor"  # Improving code
    REVIEW = "review"  # Validation and documentation
```

### TestType Enum

```python
class TestType(Enum):
    """Types of tests in TDD workflow."""
    UNIT = "unit"
    INTEGRATION = "integration"
    ACCEPTANCE = "acceptance"
    PERFORMANCE = "performance"
    SECURITY = "security"
    REGRESSION = "regression"
```

### TestStatus Enum

```python
class TestStatus(Enum):
    """Status of a test case."""
    PENDING = "pending"
    PASSED = "passed"
    FAILED = "failed"
    SKIPPED = "skipped"
    ERROR = "error"
```

## Data Classes

### TestSpecification

```python
@dataclass
class TestSpecification:
    """Specification for generating test cases."""
    name: str
    description: str
    test_type: TestType
    requirements: List[str]
    acceptance_criteria: List[str]
    edge_cases: List[str]
    mock_requirements: List[str] = field(default_factory=list)
    fixture_requirements: List[str] = field(default_factory=list)
    timeout: Optional[int] = None
    tags: List[str] = field(default_factory=list)
```

### TestCase

```python
@dataclass
class TestCase:
    """Individual test case with metadata."""
    id: str
    name: str
    file_path: str
    line_number: int
    test_type: TestType
    specification: TestSpecification
    status: TestStatus
    execution_time: Optional[float] = None
    error_message: Optional[str] = None
    code: str = ""
    coverage_impact: float = 0.0
```

### TDDSession

```python
@dataclass
class TDDSession:
    """TDD session tracking all cycle activities."""
    id: str
    project_path: str
    current_phase: TDDPhase
    test_cases: List[TestCase]
    implementation_files: List[str]
    metrics: Dict[str, Any]
    context7_patterns: Dict[str, Any]
    started_at: float
    last_activity: float
```

### TDDCycleResult

```python
@dataclass
class TDDCycleResult:
    """Results of a complete TDD cycle."""
    session_id: str
    test_specification: TestSpecification
    test_file_path: str
    implementation_file_path: str
    red_phase_result: Dict[str, Any]
    green_phase_result: Dict[str, Any]
    refactor_phase_result: Dict[str, Any]
    final_coverage: float
    total_time: float
    context7_patterns_applied: List[str]
```

## TDDManager Class

```python
class TDDManager:
    """Manages TDD workflow with Context7 integration."""

    def __init__(self, project_path: str, context7_client=None):
        self.project_path = project_path
        self.context7 = context7_client
        self.context7_integration = Context7TDDIntegration(context7_client)
        self.test_generator = TestGenerator(context7_client)
        self.current_session: Optional[TDDSession] = None

    async def start_tdd_session(self, feature_name: str) -> TDDSession:
        """Start a new TDD session."""
        session_id = f"tdd_{feature_name}_{int(time.time())}"

        # Load Context7 patterns
        patterns = await self.context7_integration.load_tdd_patterns()

        session = TDDSession(
            id=session_id,
            project_path=self.project_path,
            current_phase=TDDPhase.RED,
            test_cases=[],
            implementation_files=[],
            metrics={'red_phases': 0, 'green_phases': 0, 'refactor_phases': 0},
            context7_patterns=patterns,
            started_at=time.time(),
            last_activity=time.time()
        )

        self.current_session = session
        return session

    async def run_full_tdd_cycle(
        self,
        specification: TestSpecification,
        target_function: str
    ) -> TDDCycleResult:
        """Run complete RED-GREEN-REFACTOR cycle."""
        if not self.current_session:
            self.current_session = await self.start_tdd_session("default")

        cycle_start = time.time()
        context7_patterns_applied = []

        # RED Phase - Write failing test
        red_result = await self._execute_red_phase(specification)
        self.current_session.metrics['red_phases'] += 1

        # GREEN Phase - Make test pass
        green_result = await self._execute_green_phase(
            specification, target_function, red_result
        )
        self.current_session.metrics['green_phases'] += 1

        # REFACTOR Phase - Improve code
        refactor_result = await self._execute_refactor_phase(
            specification, green_result
        )
        self.current_session.metrics['refactor_phases'] += 1
        context7_patterns_applied.extend(refactor_result.get('patterns_applied', []))

        # Run final coverage
        coverage = await self._run_coverage_analysis()

        return TDDCycleResult(
            session_id=self.current_session.id,
            test_specification=specification,
            test_file_path=red_result.get('test_file_path', ''),
            implementation_file_path=green_result.get('implementation_file_path', ''),
            red_phase_result=red_result,
            green_phase_result=green_result,
            refactor_phase_result=refactor_result,
            final_coverage=coverage.get('total_coverage', 0.0),
            total_time=time.time() - cycle_start,
            context7_patterns_applied=context7_patterns_applied
        )
```

## Phase Execution Methods

### RED Phase

```python
async def _execute_red_phase(
    self, specification: TestSpecification
) -> Dict[str, Any]:
    """Execute RED phase - write failing test."""
    self.current_session.current_phase = TDDPhase.RED

    # Generate test from specification
    test_code = await self.test_generator.generate_test(specification)

    # Determine test file path
    test_file_path = self._get_test_file_path(specification)

    # Write test to file
    self._write_test_file(test_file_path, test_code)

    # Run test - should fail
    test_result = await self._run_tests(test_file_path)

    # Create test case record
    test_case = TestCase(
        id=f"tc_{specification.name}",
        name=specification.name,
        file_path=test_file_path,
        line_number=1,
        test_type=specification.test_type,
        specification=specification,
        status=TestStatus.FAILED if test_result['failed'] > 0 else TestStatus.PASSED,
        execution_time=test_result.get('execution_time', 0),
        code=test_code
    )

    self.current_session.test_cases.append(test_case)

    return {
        'test_code': test_code,
        'test_file_path': test_file_path,
        'test_result': test_result,
        'test_case': test_case,
        'phase_success': test_result['failed'] > 0  # Should fail in RED phase
    }
```

### GREEN Phase

```python
async def _execute_green_phase(
    self, specification: TestSpecification,
    target_function: str,
    red_result: Dict[str, Any]
) -> Dict[str, Any]:
    """Execute GREEN phase - make test pass."""
    self.current_session.current_phase = TDDPhase.GREEN

    # Generate minimum implementation
    implementation = await self._generate_minimum_implementation(
        specification, target_function
    )

    # Determine implementation file path
    impl_file_path = self._get_implementation_file_path(target_function)

    # Write implementation
    self._write_implementation_file(impl_file_path, implementation)
    self.current_session.implementation_files.append(impl_file_path)

    # Run tests - should pass
    test_result = await self._run_tests(red_result['test_file_path'])

    # Update test case status
    for tc in self.current_session.test_cases:
        if tc.name == specification.name:
            tc.status = TestStatus.PASSED if test_result['passed'] > 0 else TestStatus.FAILED

    return {
        'implementation': implementation,
        'implementation_file_path': impl_file_path,
        'test_result': test_result,
        'phase_success': test_result['failed'] == 0  # All tests should pass
    }
```

### REFACTOR Phase

```python
async def _execute_refactor_phase(
    self, specification: TestSpecification,
    green_result: Dict[str, Any]
) -> Dict[str, Any]:
    """Execute REFACTOR phase - improve code quality."""
    self.current_session.current_phase = TDDPhase.REFACTOR

    # Get refactoring patterns from Context7
    refactor_patterns = await self.context7_integration.get_refactoring_patterns()

    # Generate improvements
    improvements = await self._generate_improvements(
        green_result['implementation'],
        refactor_patterns
    )

    patterns_applied = []
    successful_refactorings = []

    for improvement in improvements:
        # Apply improvement
        refactored = await self._apply_refactoring(
            green_result['implementation_file_path'],
            improvement
        )

        if refactored['success']:
            # Run tests after refactoring
            test_result = await self._run_tests(specification.name)

            if test_result['failed'] == 0:
                successful_refactorings.append(improvement)
                patterns_applied.append(improvement.get('pattern', 'custom'))
            else:
                # Rollback failed refactoring
                await self._rollback_refactoring(
                    green_result['implementation_file_path']
                )

    return {
        'improvements_suggested': len(improvements),
        'improvements_applied': len(successful_refactorings),
        'patterns_applied': patterns_applied,
        'phase_success': True
    }
```

## Helper Methods

```python
def _get_test_file_path(self, specification: TestSpecification) -> str:
    """Determine test file path based on specification."""
    test_dir = os.path.join(self.project_path, 'tests')
    os.makedirs(test_dir, exist_ok=True)

    test_type_dir = specification.test_type.value
    full_test_dir = os.path.join(test_dir, test_type_dir)
    os.makedirs(full_test_dir, exist_ok=True)

    return os.path.join(full_test_dir, f"test_{specification.name}.py")

def _get_implementation_file_path(self, target_function: str) -> str:
    """Determine implementation file path."""
    src_dir = os.path.join(self.project_path, 'src')
    os.makedirs(src_dir, exist_ok=True)
    return os.path.join(src_dir, f"{target_function}.py")

async def _run_tests(self, test_path: str) -> Dict[str, Any]:
    """Run pytest on specified path."""
    result = subprocess.run(
        ['pytest', test_path, '-v', '--tb=short', '--json-report'],
        capture_output=True,
        text=True,
        cwd=self.project_path
    )

    return {
        'passed': result.stdout.count('PASSED'),
        'failed': result.stdout.count('FAILED'),
        'errors': result.stdout.count('ERROR'),
        'execution_time': 0.0,  # Parse from output
        'output': result.stdout
    }

async def _run_coverage_analysis(self) -> Dict[str, Any]:
    """Run coverage analysis."""
    result = subprocess.run(
        ['pytest', '--cov=src', '--cov-report=json'],
        capture_output=True,
        text=True,
        cwd=self.project_path
    )

    try:
        coverage_file = os.path.join(self.project_path, 'coverage.json')
        with open(coverage_file) as f:
            coverage_data = json.load(f)
            return {'total_coverage': coverage_data.get('totals', {}).get('percent_covered', 0)}
    except Exception:
        return {'total_coverage': 0.0}
```

## Related Sub-modules

- [Test Generation](./test-generation.md) - AI-powered test creation
- [Context7 Patterns](./context7-patterns.md) - Pattern integration

---

Sub-module: `modules/tdd/core-classes.md`
