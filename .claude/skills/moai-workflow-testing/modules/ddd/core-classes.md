# DDD with Context7 - Core Classes

> Sub-module: Core class implementations for DDD workflow management
> Parent: [DDD with Context7](../ddd-context7.md)

## Enumerations

### DDDPhase Enum

```python
class DDDPhase(Enum):
    """Phases of the DDD cycle."""
    ANALYZE = "analyze"        # Analyzing existing code and behavior
    PRESERVE = "preserve"      # Creating characterization tests
    IMPROVE = "improve"        # Refactoring with behavior preservation
    REVIEW = "review"          # Validation and documentation
```

### TestType Enum

```python
class TestType(Enum):
    """Types of tests in DDD workflow."""
    UNIT = "unit"
    INTEGRATION = "integration"
    ACCEPTANCE = "acceptance"
    PERFORMANCE = "performance"
    SECURITY = "security"
    REGRESSION = "regression"
    CHARACTERIZATION = "characterization"
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

### DDDSession

```python
@dataclass
class DDDSession:
    """DDD session tracking all cycle activities."""
    id: str
    project_path: str
    current_phase: DDDPhase
    test_cases: List[TestCase]
    implementation_files: List[str]
    metrics: Dict[str, Any]
    context7_patterns: Dict[str, Any]
    started_at: float
    last_activity: float
```

### DDDCycleResult

```python
@dataclass
class DDDCycleResult:
    """Results of a complete DDD cycle."""
    session_id: str
    test_specification: TestSpecification
    test_file_path: str
    implementation_file_path: str
    analyze_phase_result: Dict[str, Any]
    preserve_phase_result: Dict[str, Any]
    improve_phase_result: Dict[str, Any]
    final_coverage: float
    total_time: float
    context7_patterns_applied: List[str]
    behavior_preserved: bool
```

## DDDManager Class

```python
class DDDManager:
    """Manages DDD workflow with Context7 integration."""

    def __init__(self, project_path: str, context7_client=None):
        self.project_path = project_path
        self.context7 = context7_client
        self.context7_integration = Context7DDDIntegration(context7_client)
        self.test_generator = TestGenerator(context7_client)
        self.current_session: Optional[DDDSession] = None

    async def start_ddd_session(self, feature_name: str) -> DDDSession:
        """Start a new DDD session."""
        session_id = f"ddd_{feature_name}_{int(time.time())}"

        # Load Context7 patterns
        patterns = await self.context7_integration.load_ddd_patterns()

        session = DDDSession(
            id=session_id,
            project_path=self.project_path,
            current_phase=DDDPhase.ANALYZE,
            test_cases=[],
            implementation_files=[],
            metrics={'analyze_phases': 0, 'preserve_phases': 0, 'improve_phases': 0},
            context7_patterns=patterns,
            started_at=time.time(),
            last_activity=time.time()
        )

        self.current_session = session
        return session

    async def run_full_ddd_cycle(
        self,
        specification: TestSpecification,
        target_function: str
    ) -> DDDCycleResult:
        """Run complete ANALYZE-PRESERVE-IMPROVE cycle."""
        if not self.current_session:
            self.current_session = await self.start_ddd_session("default")

        cycle_start = time.time()
        context7_patterns_applied = []

        # ANALYZE Phase - Understand existing code and behavior
        analyze_result = await self._execute_analyze_phase(specification)
        self.current_session.metrics['analyze_phases'] += 1

        # PRESERVE Phase - Create characterization tests
        preserve_result = await self._execute_preserve_phase(
            specification, target_function, analyze_result
        )
        self.current_session.metrics['preserve_phases'] += 1

        # IMPROVE Phase - Refactor with behavior preservation
        improve_result = await self._execute_improve_phase(
            specification, preserve_result
        )
        self.current_session.metrics['improve_phases'] += 1
        context7_patterns_applied.extend(improve_result.get('patterns_applied', []))

        # Run final coverage
        coverage = await self._run_coverage_analysis()

        return DDDCycleResult(
            session_id=self.current_session.id,
            test_specification=specification,
            test_file_path=preserve_result.get('test_file_path', ''),
            implementation_file_path=improve_result.get('implementation_file_path', ''),
            analyze_phase_result=analyze_result,
            preserve_phase_result=preserve_result,
            improve_phase_result=improve_result,
            final_coverage=coverage.get('total_coverage', 0.0),
            total_time=time.time() - cycle_start,
            context7_patterns_applied=context7_patterns_applied,
            behavior_preserved=improve_result.get('behavior_preserved', True)
        )
```

## Phase Execution Methods

### ANALYZE Phase

```python
async def _execute_analyze_phase(
    self, specification: TestSpecification
) -> Dict[str, Any]:
    """Execute ANALYZE phase - understand existing code and behavior."""
    self.current_session.current_phase = DDDPhase.ANALYZE

    # Analyze existing code structure
    code_analysis = await self._analyze_existing_code(specification)

    # Identify behavior patterns
    behavior_patterns = await self._identify_behavior_patterns(code_analysis)

    # Determine refactoring targets
    refactoring_targets = await self._identify_refactoring_targets(code_analysis)

    return {
        'code_analysis': code_analysis,
        'behavior_patterns': behavior_patterns,
        'refactoring_targets': refactoring_targets,
        'phase_success': True
    }
```

### PRESERVE Phase

```python
async def _execute_preserve_phase(
    self, specification: TestSpecification,
    target_function: str,
    analyze_result: Dict[str, Any]
) -> Dict[str, Any]:
    """Execute PRESERVE phase - create characterization tests."""
    self.current_session.current_phase = DDDPhase.PRESERVE

    # Generate characterization tests for existing behavior
    test_code = await self.test_generator.generate_characterization_test(
        specification, analyze_result['behavior_patterns']
    )

    # Determine test file path
    test_file_path = self._get_test_file_path(specification)

    # Write test to file
    self._write_test_file(test_file_path, test_code)

    # Run test - should pass (testing existing behavior)
    test_result = await self._run_tests(test_file_path)

    # Create test case record
    test_case = TestCase(
        id=f"tc_{specification.name}",
        name=specification.name,
        file_path=test_file_path,
        line_number=1,
        test_type=TestType.CHARACTERIZATION,
        specification=specification,
        status=TestStatus.PASSED if test_result['failed'] == 0 else TestStatus.FAILED,
        execution_time=test_result.get('execution_time', 0),
        code=test_code
    )

    self.current_session.test_cases.append(test_case)

    return {
        'test_code': test_code,
        'test_file_path': test_file_path,
        'test_result': test_result,
        'test_case': test_case,
        'phase_success': test_result['failed'] == 0  # Should pass in PRESERVE phase
    }
```

### IMPROVE Phase

```python
async def _execute_improve_phase(
    self, specification: TestSpecification,
    preserve_result: Dict[str, Any]
) -> Dict[str, Any]:
    """Execute IMPROVE phase - refactor with behavior preservation."""
    self.current_session.current_phase = DDDPhase.IMPROVE

    # Get improvement patterns from Context7
    improve_patterns = await self.context7_integration.get_improvement_patterns()

    # Generate improvements
    improvements = await self._generate_improvements(
        preserve_result.get('implementation', ''),
        improve_patterns
    )

    patterns_applied = []
    successful_improvements = []
    behavior_preserved = True

    for improvement in improvements:
        # Apply improvement
        improved = await self._apply_improvement(
            preserve_result.get('implementation_file_path', ''),
            improvement
        )

        if improved['success']:
            # Run characterization tests to verify behavior preservation
            test_result = await self._run_tests(preserve_result['test_file_path'])

            if test_result['failed'] == 0:
                successful_improvements.append(improvement)
                patterns_applied.append(improvement.get('pattern', 'custom'))
            else:
                # Rollback failed improvement - behavior not preserved
                await self._rollback_improvement(
                    preserve_result.get('implementation_file_path', '')
                )
                behavior_preserved = False

    return {
        'improvements_suggested': len(improvements),
        'improvements_applied': len(successful_improvements),
        'patterns_applied': patterns_applied,
        'behavior_preserved': behavior_preserved,
        'phase_success': behavior_preserved
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

Sub-module: `modules/ddd/core-classes.md`
