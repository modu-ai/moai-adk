---
name: moai-workflow-testing
description: Comprehensive development workflow specialist combining TDD, debugging, performance optimization, code review, and quality assurance into unified development workflows
version: 1.0.0
category: workflow
tags:
  - workflow
  - testing
  - debugging
  - performance
  - quality
  - tdd
  - review
updated: 2025-11-30
status: active
author: MoAI-ADK Team
---

# Development Workflow Specialist

## Quick Reference (30 seconds)

**Unified Development Workflow** - Comprehensive development lifecycle management combining TDD, AI-powered debugging, performance optimization, automated code review, and quality assurance into integrated workflows.

**Core Capabilities**:
- 🧪 **Test-Driven Development**: RED-GREEN-REFACTOR cycle with Context7 patterns
- 🔍 **AI-Powered Debugging**: Intelligent error analysis and Context7 best practices
- ⚡ **Performance Optimization**: Real-time profiling and bottleneck detection
- 🔬 **Automated Code Review**: TRUST 5 validation with AI quality analysis
- 📊 **Quality Assurance**: Comprehensive testing and CI/CD integration
- 🎯 **Workflow Orchestration**: End-to-end development process automation

**Unified Development Workflow**:
```
Debug → Refactor → Optimize → Review → Test → Profile
   ↓        ↓         ↓        ↓      ↓       ↓
AI-     AI-       AI-      AI-    AI-     AI-
Powered Powered  Powered  Powered Powered Powered
```

**When to Use**:
- Complete development lifecycle management
- Enterprise-grade quality assurance
- Multi-language development projects
- Performance-critical applications
- Technical debt reduction initiatives
- Automated testing and CI/CD integration

---

## Implementation Guide

### AI-Powered Debugging Integration

**Intelligent Debugging with Context7**:
```python
class AIDebugger:
    """AI-powered debugging with Context7 integration."""

    def __init__(self, context7_client=None):
        self.context7 = context7_client
        self.error_patterns = self.load_error_patterns()

    async def debug_with_context7_patterns(
        self, error: Exception, context: Dict, codebase_path: str
    ) -> DebugAnalysis:
        """Debug using AI pattern recognition and Context7 best practices."""

        # Get latest debugging patterns from Context7
        debugpy_patterns = await self.context7.get_library_docs(
            context7_library_id="/microsoft/debugpy",
            topic="AI debugging patterns error analysis 2025",
            tokens=5000
        )

        # AI pattern classification
        error_analysis = self.classify_error(error, context)
        pattern_match = self.match_context7_patterns(error, debugpy_patterns)

        # Generate solutions using AI + Context7
        solutions = await self.generate_solutions(
            error_analysis, pattern_match, debugpy_patterns, context
        )

        return DebugAnalysis(
            error_type=error_analysis.type,
            confidence=error_analysis.confidence,
            context7_patterns=pattern_match,
            solutions=solutions,
            prevention_strategies=self.suggest_prevention(error_analysis)
        )

    def classify_error(self, error: Exception, context: Dict) -> ErrorAnalysis:
        """Classify error using AI pattern recognition."""

        error_info = {
            'type': type(error).__name__,
            'message': str(error),
            'traceback': traceback.format_exc(),
            'context': context,
            'frequency': self.get_error_frequency(error)
        }

        # AI classification logic
        patterns = {
            'ImportError': self.handle_import_errors,
            'AttributeError': self.handle_attribute_errors,
            'TypeError': self.handle_type_errors,
            'ValueError': self.handle_value_errors,
            'KeyError': self.handle_key_errors,
        }

        handler = patterns.get(type(error).__name__, self.handle_generic_errors)
        return handler(error_info)

    async def generate_solutions(
        self, error_analysis: ErrorAnalysis,
        pattern_match: Dict, context7_patterns: Dict,
        context: Dict
    ) -> List[Solution]:
        """Generate solutions using AI and Context7 patterns."""

        solutions = []

        # Context7-based solutions
        for pattern in pattern_match.get('matched_patterns', []):
            solution = Solution(
                type='context7_pattern',
                description=pattern['description'],
                code_example=pattern['example'],
                confidence=pattern['confidence']
            )
            solutions.append(solution)

        # AI-generated solutions
        if self.context7:
            ai_solution = await self.context7.get_library_docs(
                context7_library_id="/openai/chatgpt",
                topic=f"solve {error_analysis.type} in {context.get('language', 'python')}",
                tokens=3000
            )

            solution = Solution(
                type='ai_generated',
                description=ai_solution['description'],
                code_example=ai_solution['code_example'],
                confidence=0.8
            )
            solutions.append(solution)

        return solutions

# Usage example
debugger = AIDebugger(context7_client=context7)
try:
    # Code that might fail
    result = some_function()
except Exception as e:
    analysis = await debugger.debug_with_context7_patterns(
        e, {'file': __file__, 'function': 'some_function'}, '/project/src'
    )
    for solution in analysis.solutions:
        print(f"Solution: {solution.description}")
```

### Smart Refactoring with Technical Debt Management

**AI-Driven Code Transformation**:
```python
class AIRefactorer:
    """AI-powered refactoring with technical debt management."""

    def __init__(self, context7_client=None):
        self.context7 = context7_client
        self.technical_debt_analyzer = TechnicalDebtAnalyzer()

    async def refactor_with_intelligence(
        self, codebase_path: str, refactor_options: Dict = None
    ) -> RefactorPlan:
        """AI-driven code transformation with technical debt quantification."""

        # Get Context7 refactoring patterns
        rope_patterns = await self.context7.get_library_docs(
            context7_library_id="/python-rope/rope",
            topic="safe refactoring patterns technical debt 2025",
            tokens=4000
        )

        # Analyze technical debt
        debt_analysis = await self.technical_debt_analyzer.analyze(codebase_path)

        # AI analysis of refactoring opportunities
        refactor_opportunities = await self.identify_refactor_opportunities(
            codebase_path, debt_analysis
        )

        # Generate safe refactor plan using Rope + AI
        refactor_plan = self.create_safe_refactor_plan(
            refactor_opportunities, rope_patterns
        )

        return RefactorPlan(
            opportunities=refactor_opportunities,
            transformations=refactor_plan.transformations,
            risk_assessment=self.assess_refactor_risks(refactor_plan),
            estimated_impact=self.calculate_impact(refactor_plan),
            context7_validated=True
        )

    async def identify_refactor_opportunities(
        self, codebase_path: str, debt_analysis: TechnicalDebtAnalysis
    ) -> List[RefactorOpportunity]:
        """Identify refactoring opportunities using AI analysis."""

        opportunities = []

        # Code smell detection
        code_smells = await self.detect_code_smells(codebase_path)
        for smell in code_smells:
            if smell.severity >= 7:  # High severity
                opportunity = RefactorOpportunity(
                    type='code_smell',
                    description=smell.description,
                    location=smell.location,
                    impact=smell.impact,
                    difficulty=smell.difficulty
                )
                opportunities.append(opportunity)

        # Duplication detection
        duplications = await self.detect_code_duplication(codebase_path)
        for dup in duplications:
            if dup.similarity >= 0.8:  # 80% similarity
                opportunity = RefactorOpportunity(
                    type='duplication',
                    description=f"Duplicate code in {dup.files}",
                    location=dup.locations,
                    impact='high',
                    difficulty='medium'
                )
                opportunities.append(opportunity)

        return opportunities

    def create_safe_refactor_plan(
        self, opportunities: List[RefactorOpportunity],
        rope_patterns: Dict
    ) -> RefactorPlan:
        """Create safe refactor plan using Rope patterns."""

        transformations = []

        for opportunity in opportunities:
            if opportunity.type == 'code_smell':
                transformation = self.create_code_smell_transformation(
                    opportunity, rope_patterns
                )
            elif opportunity.type == 'duplication':
                transformation = self.create_duplication_transformation(
                    opportunity, rope_patterns
                )

            if transformation:
                transformations.append(transformation)

        # Order transformations by risk and dependencies
        transformations = self.order_transformations(transformations)

        return RefactorPlan(
            transformations=transformations,
            estimated_time=self.calculate_total_time(transformations),
            risk_score=self.calculate_overall_risk(transformations)
        )
```

### Performance Optimization with Real-time Profiling

**AI-Enhanced Performance Analysis**:
```python
class AIPerformanceOptimizer:
    """AI-powered performance optimization with real-time profiling."""

    def __init__(self, context7_client=None):
        self.context7 = context7_client
        self.profiler = None

    async def optimize_performance(
        self, codebase_path: str, performance_targets: Dict
    ) -> OptimizationPlan:
        """Optimize performance using AI and Context7 patterns."""

        # Get Context7 optimization patterns
        perf_patterns = await self.context7.get_library_docs(
            context7_library_id="/emeryberger/scalene",
            topic="performance profiling optimization GPU 2025",
            tokens=5000
        )

        # Scalene profiling with AI analysis
        scalene_profile = await self.profile_with_ai(codebase_path)

        # AI bottleneck detection
        bottlenecks = await self.detect_bottlenecks(scalene_profile, perf_patterns)

        # Generate optimization plan
        optimization_plan = self.create_optimization_plan(
            bottlenecks, scalene_profile, perf_patterns
        )

        return OptimizationPlan(
            bottlenecks=bottlenecks,
            optimizations=optimization_plan.optimizations,
            expected_improvement=self.calculate_improvement(optimization_plan),
            implementation_priority=self.prioritize_optimizations(bottlenecks)
        )

    async def profile_with_ai(self, codebase_path: str) -> ScaleneProfile:
        """Profile with Scalene and AI analysis."""

        # Initialize Scalene profiler
        import scalene
        self.profiler = scalene.Scalene()

        # Profile the application
        profile_data = self.profiler.profile(
            f"python -m pytest {codebase_path}/tests/",
            memory=True,
            cpu=True,
            gpu=True
        )

        # AI analysis of profile data
        ai_analysis = await self.analyze_profile_data(profile_data)

        return ScaleneProfile(
            raw_data=profile_data,
            ai_analysis=ai_analysis,
            hotspots=self.identify_hotspots(profile_data),
            memory_usage=self.analyze_memory_usage(profile_data),
            cpu_usage=self.analyze_cpu_usage(profile_data)
        )

    async def detect_bottlenecks(
        self, profile: ScaleneProfile, perf_patterns: Dict
    ) -> List[Bottleneck]:
        """AI-powered bottleneck detection."""

        bottlenecks = []

        # CPU bottlenecks
        for hotspot in profile.hotspots:
            if hotspot.cpu_percentage > 80:  # High CPU usage
                bottleneck = Bottleneck(
                    type='cpu',
                    location=hotspot.function,
                    description=f"High CPU usage in {hotspot.function}",
                    severity=hotspot.cpu_percentage / 100,
                    suggestions=self.get_cpu_optimization_suggestions(hotspot, perf_patterns)
                )
                bottlenecks.append(bottleneck)

        # Memory bottlenecks
        for memory_leak in profile.memory_leaks:
            bottleneck = Bottleneck(
                type='memory',
                location=memory_leak.function,
                description=f"Memory leak in {memory_leak.function}",
                severity=memory_leak.leak_size / 1000000,  # Convert to MB
                suggestions=self.get_memory_optimization_suggestions(memory_leak, perf_patterns)
            )
            bottlenecks.append(bottleneck)

        return bottlenecks
```

### Test-Driven Development with Context7 Integration

**Enhanced TDD Implementation**:
```python
class TDDOrchestrator:
    """TDD orchestrator with Context7 best practices and AI assistance."""

    def __init__(self, context7_client=None):
        self.context7 = context7_client
        self.test_generator = AITestGenerator()

    async def execute_tdd_cycle(
        self, requirement: str, codebase_path: str
    ) -> TDDResult:
        """Execute complete RED-GREEN-REFACTOR TDD cycle."""

        # RED Phase: Write failing test
        red_result = await self.red_phase(requirement, codebase_path)

        # GREEN Phase: Make test pass
        green_result = await self.green_phase(red_result.test, codebase_path)

        # REFACTOR Phase: Improve code quality
        refactor_result = await self.refactor_phase(
            green_result.code, red_result.test, codebase_path
        )

        return TDDResult(
            requirement=requirement,
            test=red_result.test,
            implementation=refactor_result.code,
            coverage=refactor_result.coverage,
            quality_score=refactor_result.quality_score
        )

    async def red_phase(self, requirement: str, codebase_path: str) -> RedResult:
        """RED phase: Generate failing test based on requirement."""

        # Get Context7 testing patterns
        test_patterns = await self.context7.get_library_docs(
            context7_library_id="/pytest-dev/pytest",
            topic="testing strategies TDD automation 2025",
            tokens=4000
        )

        # Generate test using AI
        test_code = await self.test_generator.generate_test(
            requirement, test_patterns
        )

        # Write test file
        test_file_path = self.write_test_file(test_code, requirement)

        # Run test to ensure it fails
        test_result = self.run_test(test_file_path)

        return RedResult(
            test=test_code,
            test_file_path=test_file_path,
            test_result=test_result,
            expected_to_fail=True
        )

    async def green_phase(self, test_code: str, codebase_path: str) -> GreenResult:
        """GREEN phase: Implement code to make test pass."""

        # Analyze test requirements
        test_analysis = self.analyze_test_requirements(test_code)

        # Generate minimal implementation
        implementation = await self.generate_minimal_implementation(test_analysis)

        # Write implementation
        impl_file_path = self.write_implementation_file(implementation, test_analysis)

        # Run tests to verify passing
        test_result = self.run_test_suite(codebase_path)

        return GreenResult(
            code=implementation,
            impl_file_path=impl_file_path,
            test_result=test_result,
            all_tests_passing=test_result.passed
        )
```

### Automated Code Review with TRUST 5 Validation

**AI-Powered Quality Assurance**:
```python
class AICodeReviewer:
    """AI-powered code review with TRUST 5 validation."""

    def __init__(self, context7_client=None):
        self.context7 = context7_client
        self.trust5_validator = TRUST5Validator()

    async def comprehensive_review(
        self, codebase_path: str, changes: List[FileChange]
    ) -> ReviewResult:
        """Comprehensive code review with AI and TRUST 5 validation."""

        # Get Context7 security and quality patterns
        security_patterns = await self.context7.get_library_docs(
            context7_library_id="/owasp/top-ten",
            topic="security vulnerability patterns 2025",
            tokens=3000
        )

        quality_patterns = await self.context7.get_library_docs(
            context7_library_id="/pylint-dev/pylint",
            topic="code quality best practices 2025",
            tokens=3000
        )

        # TRUST 5 validation
        trust5_analysis = await self.trust5_validator.validate(codebase_path, changes)

        # AI quality analysis
        quality_analysis = await self.analyze_code_quality(codebase_path, changes)

        # Security vulnerability detection
        security_analysis = await self.detect_security_vulnerabilities(
            changes, security_patterns
        )

        return ReviewResult(
            trust5_validation=trust5_analysis,
            quality_analysis=quality_analysis,
            security_analysis=security_analysis,
            recommendations=self.generate_recommendations(
                trust5_analysis, quality_analysis, security_analysis
            ),
            approval_status=self.determine_approval_status(trust5_analysis)
        )

    async def analyze_code_quality(
        self, codebase_path: str, changes: List[FileChange]
    ) -> QualityAnalysis:
        """AI-powered code quality analysis."""

        quality_metrics = {}

        for change in changes:
            # Analyze each changed file
            file_metrics = await self.analyze_file_quality(change)
            quality_metrics[change.file_path] = file_metrics

        # Calculate overall quality score
        overall_score = self.calculate_overall_quality_score(quality_metrics)

        return QualityAnalysis(
            file_metrics=quality_metrics,
            overall_score=overall_score,
            issues=self.identify_quality_issues(quality_metrics),
            suggestions=self.generate_quality_suggestions(quality_metrics)
        )

class TRUST5Validator:
    """TRUST 5 framework validation."""

    def __init__(self):
        self.trust_principles = {
            'Transparency': self.validate_transparency,
            'Reliability': self.validate_reliability,
            'Usability': self.validate_usability,
            'Security': self.validate_security,
            'Testability': self.validate_testability
        }

    async def validate(
        self, codebase_path: str, changes: List[FileChange]
    ) -> TRUST5Analysis:
        """Validate changes against TRUST 5 principles."""

        validation_results = {}

        for principle, validator in self.trust_principles.items():
            result = await validator(changes)
            validation_results[principle] = result

        return TRUST5Analysis(
            principles_validation=validation_results,
            overall_compliance=self.calculate_compliance(validation_results),
            critical_issues=self.identify_critical_issues(validation_results)
        )
```

---

## Advanced Patterns

### Integrated Workflow Orchestration

**Complete Development Pipeline**:
```python
class DevelopmentWorkflowOrchestrator:
    """Orchestrates complete development workflow with all components."""

    def __init__(self, project_config: Dict):
        self.config = project_config
        self.debugger = AIDebugger()
        self.refactorer = AIRefactorer()
        self.optimizer = AIPerformanceOptimizer()
        self.reviewer = AICodeReviewer()
        self.tdd = TDDOrchestrator()

    async def orchestrate_complete_workflow(
        self, task: DevelopmentTask
    ) -> WorkflowResult:
        """Orchestrate complete development workflow."""

        workflow_steps = []

        # Step 1: Requirements analysis and test generation
        if task.type == 'feature':
            tdd_result = await self.tdd.execute_tdd_cycle(
                task.requirement, task.codebase_path
            )
            workflow_steps.append(('tdd', tdd_result))

        # Step 2: Debug any issues found
        if task.has_errors:
            debug_result = await self.debugger.debug_with_context7_patterns(
                task.errors, task.context, task.codebase_path
            )
            workflow_steps.append(('debug', debug_result))

        # Step 3: Performance optimization
        if task.requires_optimization:
            perf_result = await self.optimizer.optimize_performance(
                task.codebase_path, task.performance_targets
            )
            workflow_steps.append(('optimize', perf_result))

        # Step 4: Code review and quality assurance
        review_result = await self.reviewer.comprehensive_review(
            task.codebase_path, task.changes
        )
        workflow_steps.append(('review', review_result))

        # Step 5: Refactoring if needed
        if review_result.quality_score < 0.8:
            refactor_result = await self.refactorer.refactor_with_intelligence(
                task.codebase_path
            )
            workflow_steps.append(('refactor', refactor_result))

        return WorkflowResult(
            steps=workflow_steps,
            final_quality_score=self.calculate_final_quality(workflow_steps),
            recommendations=self.generate_workflow_recommendations(workflow_steps),
            deployment_ready=self.is_deployment_ready(workflow_steps)
        )
```

---

## Works Well With

- **moai-domain-backend** - Backend development and testing
- **moai-domain-frontend** - Frontend component testing and optimization
- **moai-domain-database** - Database testing and performance optimization
- **moai-docs-generation** - Test documentation and API docs
- **moai-integration-mcp** - MCP service testing and validation

---

## Usage Examples

### CLI Integration
```bash
# Execute TDD cycle
moai-workflow tdd --requirement "User authentication" --src ./src

# Debug with AI assistance
moai-workflow debug --error-file error.log --context development

# Optimize performance
moai-workflow optimize --target-dir ./src --performance-targets config/perf.json

# Review code changes
moai-workflow review --changes changes.json --codebase ./src

# Run complete workflow
moai-workflow execute --task config/task.json
```

### Python API
```python
from moai_workflow_testing import DevelopmentWorkflowOrchestrator

# Initialize orchestrator
orchestrator = DevelopmentWorkflowOrchestrator(project_config)

# Execute complete workflow
task = DevelopmentTask(
    type='feature',
    requirement='Implement user authentication',
    codebase_path='./src',
    performance_targets={'response_time': 200}
)

result = await orchestrator.orchestrate_complete_workflow(task)
```

---

## Technology Stack

**Core Libraries**:
- **pytest**: Testing framework with plugins
- **scalene**: High-performance CPU/GPU/memory profiler
- **rope**: Python refactoring library
- **pylint**: Code quality analysis
- **coverage.py**: Test coverage measurement

**AI Integration**:
- **Context7**: Latest documentation and best practices
- **OpenAI API**: Code generation and analysis
- **Custom ML models**: Pattern recognition and optimization

**Development Tools**:
- **debugpy**: Python debugging
- **black**: Code formatting
- **mypy**: Type checking
- **bandit**: Security linting

---

**Status**: Production Ready
**Last Updated**: 2025-11-30
**Maintained by**: MoAI-ADK Workflow Team